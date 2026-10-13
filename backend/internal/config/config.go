package config

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string // development | production
	Port        string
	DatabaseURL string

	CORSOrigins    []string
	SessionTTL     time.Duration
	CookieSecure   bool
	CookieSameSite http.SameSite
	SeedPassword   string
}

// Load membaca .env (bila ada) lalu environment sistem.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("file .env tidak ditemukan, memakai environment sistem")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL belum diatur")
	}

	ttlJam, err := strconv.Atoi(getenv("SESSION_TTL_HOURS", "168"))
	if err != nil || ttlJam <= 0 {
		return nil, errors.New("SESSION_TTL_HOURS harus berupa angka positif")
	}

	return &Config{
		Env:         getenv("APP_ENV", "development"),
		Port:        getenv("PORT", "8080"),
		DatabaseURL: dsn,

		CORSOrigins:    splitList(getenv("CORS_ORIGINS", "http://localhost:3031,http://localhost:5173,https://1lw5n254-3031.asse.devtunnels.ms")),
		SessionTTL:     time.Duration(ttlJam) * time.Hour,
		CookieSecure:   getenv("COOKIE_SECURE", "false") == "true",
		CookieSameSite: parseSameSite(getenv("COOKIE_SAMESITE", "lax")),
		SeedPassword:   getenv("SEED_PASSWORD", "MedClaim#demo2026"),
	}, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitList(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}
