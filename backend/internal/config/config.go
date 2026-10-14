package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"parfume/backend/internal/storage"
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
	// Production turns on the checks below; set APP_ENV=production.
	Production     bool
	S3             storage.S3Config
	TrustedProxies int
	ClientIPHeader string
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
		Production:    env("APP_ENV", "") == "production",
		S3: storage.S3Config{
			Bucket:    env("AWS_BUCKET", ""),
			Endpoint:  env("AWS_ENDPOINT_URL", env("AWS_ENDPOINT", "")),
			Region:    env("AWS_REGION", env("AWS_DEFAULT_REGION", "auto")),
			AccessKey: env("AWS_ACCESS_KEY_ID", ""),
			SecretKey: env("AWS_SECRET_ACCESS_KEY", ""),
		},
		TrustedProxies: envInt("TRUSTED_PROXIES", 0),
		ClientIPHeader: env("CLIENT_IP_HEADER", ""),
	}
}

// Validate refuses to start production with the development defaults.
func (c Config) Validate() error {
	if !c.Production {
		return nil
	}
	if os.Getenv("JWT_SECRET") == "" || len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be set to a random string of at least 32 characters")
	}
	if os.Getenv("ADMIN_PASSWORD") == "" || c.AdminPassword == "secret" {
		return errors.New("ADMIN_PASSWORD must be set")
	}
	if c.S3.Bucket == "" {
		return errors.New("attach a bucket (AWS_BUCKET): the server disk is wiped on every deploy")
	}
	return nil
}

func envInt(key string, fallback int) int {
	if n, err := strconv.Atoi(env(key, "")); err == nil {
		return n
	}
	return fallback
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
