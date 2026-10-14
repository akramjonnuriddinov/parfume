package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"parfume/backend/internal/assistant"
	"parfume/backend/internal/auth"
	"parfume/backend/internal/storage"
	"parfume/backend/internal/store"
)

type Handler struct {
	store     *store.Store
	auth      *auth.Manager
	assistant *assistant.Assistant
	files     storage.Storage
}

type Options struct {
	Store       *store.Store
	Auth        *auth.Manager
	Assistant   *assistant.Assistant
	Files       storage.Storage
	StaticDir   string
	CORSOrigins []string
	// TrustedProxies is how many proxies sit in front of the server; the
	// client IP for rate limiting is then read from X-Forwarded-For.
	TrustedProxies int
	// ClientIPHeader, when set, names a header holding the client IP that
	// the edge in front of the server overwrites (CF-Connecting-IP).
	ClientIPHeader string
}

func NewRouter(opts Options) http.Handler {
	h := &Handler{store: opts.Store, auth: opts.Auth, assistant: opts.Assistant, files: opts.Files}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: opts.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	writeLimiter := newRateLimiter(20, time.Minute, opts.TrustedProxies, opts.ClientIPHeader)
	loginLimiter := newRateLimiter(10, time.Minute, opts.TrustedProxies, opts.ClientIPHeader)
	chatLimiter := newRateLimiter(15, time.Minute, opts.TrustedProxies, opts.ClientIPHeader)

	r.Route("/api", func(r chi.Router) {
		// The assistant may call OpenAI several times per message.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(90 * time.Second))
			r.Get("/chat", h.chatStatus)
			r.With(chatLimiter.Middleware).Post("/chat", h.chat)
			r.With(writeLimiter.Middleware).Post("/chat/order", h.chatConfirmOrder)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(30 * time.Second))

			r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			})

			r.Get("/categories", h.listCategories)
			r.Get("/brands", h.listBrands)
			r.Get("/products", h.listProducts)
			r.Get("/products/{id}", h.getProduct)
			r.Get("/settings", h.getSettings)
			r.With(writeLimiter.Middleware).Post("/orders", h.createOrder)
			r.With(writeLimiter.Middleware).Post("/messages", h.createMessage)

			r.Route("/admin", func(r chi.Router) {
				r.With(loginLimiter.Middleware).Post("/login", h.login)

				r.Group(func(r chi.Router) {
					r.Use(h.auth.Middleware)

					r.Get("/me", h.me)
					r.Post("/password", h.changePassword)
					r.Get("/stats", h.stats)
					r.Post("/uploads", h.upload)

					r.Get("/categories", h.adminListCategories)
					r.Post("/categories", h.createCategory)
					r.Put("/categories/{id}", h.updateCategory)
					r.Delete("/categories/{id}", h.deleteCategory)

					r.Get("/products", h.adminListProducts)
					r.Post("/products", h.createProduct)
					r.Get("/products/{id}", h.adminGetProduct)
					r.Put("/products/{id}", h.updateProduct)
					r.Delete("/products/{id}", h.deleteProduct)

					r.Get("/orders", h.adminListOrders)
					r.Get("/orders/{id}", h.adminGetOrder)
					r.Patch("/orders/{id}", h.adminUpdateOrder)
					r.Delete("/orders/{id}", h.adminDeleteOrder)

					r.Get("/messages", h.adminListMessages)
					r.Patch("/messages/{id}", h.adminUpdateMessage)
					r.Delete("/messages/{id}", h.adminDeleteMessage)

					r.Get("/settings", h.getSettings)
					r.Put("/settings", h.adminSaveSettings)

					r.Get("/chats", h.adminListChats)
					r.Get("/chats/{id}", h.adminGetChat)
					r.Delete("/chats/{id}", h.adminDeleteChat)
					r.Get("/missing-products", h.adminMissingProducts)
					r.Post("/missing-products/delete", h.adminDeleteMissingProduct)

				})
			})
		})

		r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
			writeError(w, http.StatusNotFound, "not found")
		})
	})

	r.Get("/uploads/*", h.serveUpload)
	r.Head("/uploads/*", h.serveUpload)

	if opts.StaticDir != "" {
		r.Get("/*", spaHandler(opts.StaticDir).ServeHTTP)
	}

	return r
}
