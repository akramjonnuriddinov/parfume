package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"parfume/backend/internal/api"
	"parfume/backend/internal/assistant"
	"parfume/backend/internal/auth"
	"parfume/backend/internal/config"
	"parfume/backend/internal/database"
	"parfume/backend/internal/storage"
	"parfume/backend/internal/store"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(ctx, pool); err != nil {
		log.Fatalf("seed: %v", err)
	}

	st := store.New(pool)
	if err := st.EnsureAdmin(ctx, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("ensure admin: %v", err)
	}

	files, err := storage.New(cfg.S3, cfg.UploadDir)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: api.NewRouter(api.Options{
			Store:          st,
			Auth:           auth.NewManager(cfg.JWTSecret),
			Assistant:      assistant.New(st, cfg.OpenAIAPIKey, cfg.OpenAIModel),
			Files:          files,
			StaticDir:      cfg.StaticDir,
			CORSOrigins:    cfg.CORSOrigins,
			TrustedProxies: cfg.TrustedProxies,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("API listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("server stopped")
}
