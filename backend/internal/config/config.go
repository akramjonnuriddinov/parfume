package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	AdminUsername string
	AdminPassword string
	UploadDir     string
	StaticDir     string
	CORSOrigins   []string
	OpenAIAPIKey  string
	OpenAIModel   string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:          env("PORT", "8089"),
		DatabaseURL:   env("DATABASE_URL", "postgres://parfume:parfume@localhost:5442/parfume?sslmode=disable"),
		JWTSecret:     env("JWT_SECRET", "dev-secret-change-me"),
		AdminUsername: env("ADMIN_USERNAME", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", "secret"),
		UploadDir:     env("UPLOAD_DIR", "./uploads"),
		StaticDir:     env("STATIC_DIR", ""),
		CORSOrigins:   splitList(env("CORS_ORIGINS", "http://localhost:5175")),
		OpenAIAPIKey:  env("OPENAI_API_KEY", ""),
		OpenAIModel:   env("OPENAI_MODEL", "gpt-4.1"),
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
