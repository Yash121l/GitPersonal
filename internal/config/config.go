package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	RequestTimeout      time.Duration
	MaxRequestBodyBytes int64
	DBMaxOpenConns      int
	DBMaxIdleConns      int
	DBConnMaxLifetime   time.Duration
	DBConnMaxIdleTime   time.Duration
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

	if cfg.ReadTimeout, err = parseDurationEnv("FORGE_READ_TIMEOUT", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = parseDurationEnv("FORGE_WRITE_TIMEOUT", 30*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.IdleTimeout, err = parseDurationEnv("FORGE_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = parseDurationEnv("FORGE_SHUTDOWN_TIMEOUT", 20*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.RequestTimeout, err = parseDurationEnv("FORGE_REQUEST_TIMEOUT", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.DBConnMaxLifetime, err = parseDurationEnv("FORGE_DB_CONN_MAX_LIFETIME", time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.DBConnMaxIdleTime, err = parseDurationEnv("FORGE_DB_CONN_MAX_IDLE_TIME", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.MaxRequestBodyBytes, err = parseInt64Env("FORGE_MAX_REQUEST_BODY_BYTES", 1<<20); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxOpenConns, err = parseIntEnv("FORGE_DB_MAX_OPEN_CONNS", 25); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxIdleConns, err = parseIntEnv("FORGE_DB_MAX_IDLE_CONNS", 25); err != nil {
		return Config{}, err
	}

	return cfg, cfg.Validate()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func (c Config) Validate() error {
	if c.Secret == "" {
		return errors.New("FORGE_SECRET must be set")
	}
	if c.CookieName == "" {
		return errors.New("FORGE_COOKIE_NAME must be set")
	}
	if c.SessionTTL <= 0 {
		return errors.New("FORGE_SESSION_TTL must be positive")
	}
	if c.ReadTimeout <= 0 || c.WriteTimeout <= 0 || c.IdleTimeout <= 0 || c.ShutdownTimeout <= 0 || c.RequestTimeout <= 0 {
		return errors.New("all timeout values must be positive")
	}
	if c.MaxRequestBodyBytes <= 0 {
		return errors.New("FORGE_MAX_REQUEST_BODY_BYTES must be positive")
	}
	if c.DBMaxOpenConns <= 0 || c.DBMaxIdleConns < 0 {
		return errors.New("database pool sizes must be non-negative and max open > 0")
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		return errors.New("FORGE_DB_MAX_IDLE_CONNS cannot exceed FORGE_DB_MAX_OPEN_CONNS")
	}
	if !filepath.IsAbs(c.ReposRoot) {
		return errors.New("FORGE_REPOS_ROOT must be an absolute path")
	}
	if _, err := url.ParseRequestURI(c.BaseURL); err != nil {
		return fmt.Errorf("parse FORGE_BASE_URL: %w", err)
	}

	switch c.Environment {
	case "development", "test", "production":
	default:
		return fmt.Errorf("unsupported FORGE_ENV %q", c.Environment)
	}

	if c.Environment == "production" {
		if c.DatabaseURL == "" {
			return errors.New("DATABASE_URL must be set in production")
		}
		if c.Secret == "dev-secret-change-me" || len(c.Secret) < 32 {
			return errors.New("FORGE_SECRET must be at least 32 characters and not use the development default in production")
		}
		if !strings.HasPrefix(strings.ToLower(c.BaseURL), "https://") {
			return errors.New("FORGE_BASE_URL must use https in production")
		}
	}

	return nil
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := getEnv(key, fallback.String())
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

func parseIntEnv(key string, fallback int) (int, error) {
	raw := getEnv(key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

func parseInt64Env(key string, fallback int64) (int64, error) {
	raw := getEnv(key, strconv.FormatInt(fallback, 10))
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}
