package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter is a small fixed-window, per-IP limiter used to protect the
// public write endpoints and the login endpoint from spam/brute force.
type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	proxies int
	header  string
	hits    map[string]*windowCount
}

type windowCount struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration, proxies int, header string) *rateLimiter {
	rl := &rateLimiter{limit: limit, window: window, proxies: proxies, header: header, hits: map[string]*windowCount{}}
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
		if !rl.allow(clientIP(r, rl.header, rl.proxies)) {
			writeError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP returns the visitor's address. Behind proxies every request comes
// from the proxy, so the address is taken from a header the edge sets (e.g.
// Cloudflare's CF-Connecting-IP), or else from X-Forwarded-For: each proxy
// appends the address it saw, so the entry `proxies` places from the end was
// written by the outermost trusted proxy and can't be forged.
func clientIP(r *http.Request, header string, proxies int) string {
	if header != "" {
		if ip := strings.TrimSpace(r.Header.Get(header)); ip != "" {
			return ip
		}
	}
	if proxies > 0 {
		var hops []string
		for _, h := range r.Header.Values("X-Forwarded-For") {
			for _, part := range strings.Split(h, ",") {
				if p := strings.TrimSpace(part); p != "" {
					hops = append(hops, p)
				}
			}
		}
		if i := len(hops) - proxies; i >= 0 && i < len(hops) {
			return hops[i]
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
