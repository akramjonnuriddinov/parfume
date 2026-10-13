package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// rateLimiter is a small fixed-window, per-IP limiter used to protect the
// public write endpoints and the login endpoint from spam/brute force.
type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*windowCount
}

type windowCount struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{limit: limit, window: window, hits: map[string]*windowCount{}}
	go func() {
		for range time.Tick(window) {
			rl.mu.Lock()
			for ip, wc := range rl.hits {
				if time.Since(wc.start) > rl.window {
					delete(rl.hits, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	wc, ok := rl.hits[ip]
	if !ok || time.Since(wc.start) > rl.window {
		rl.hits[ip] = &windowCount{start: time.Now(), count: 1}
		return true
	}
	wc.count++
	return wc.count <= rl.limit
}

func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !rl.allow(ip) {
			writeError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
