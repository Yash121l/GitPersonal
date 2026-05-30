package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	defaultAddress     = ":3000"
	defaultBaseURL     = "http://localhost:3000"
	defaultCookieName  = "forge_session"
	defaultEnvironment = "development"
	defaultReposRoot   = "/data/repos"
	defaultSessionTTL  = 7 * 24 * time.Hour
)

type Config struct {
	Environment string
	Address     string
	BaseURL     string
	Secret      string
	CookieName  string
	SessionTTL  time.Duration
	ReposRoot   string
	DatabaseURL string
	RedisURL    string
}

func Load() (Config, error) {
	cfg := Config{
		Environment: getEnv("FORGE_ENV", defaultEnvironment),
		Address:     getEnv("FORGE_ADDR", defaultAddress),
		BaseURL:     getEnv("FORGE_BASE_URL", defaultBaseURL),
		Secret:      getEnv("FORGE_SECRET", "dev-secret-change-me"),
		CookieName:  getEnv("FORGE_COOKIE_NAME", defaultCookieName),
		ReposRoot:   getEnv("FORGE_REPOS_ROOT", defaultReposRoot),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	sessionTTL := getEnv("FORGE_SESSION_TTL", defaultSessionTTL.String())
	ttl, err := time.ParseDuration(sessionTTL)
	if err != nil {
		return Config{}, fmt.Errorf("parse FORGE_SESSION_TTL: %w", err)
	}
	cfg.SessionTTL = ttl

	if cfg.Secret == "" {
		return Config{}, errors.New("FORGE_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
