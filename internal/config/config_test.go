package config

import (
	"testing"
	"time"
)

func TestValidateRejectsUnsafeProductionDefaults(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:         "production",
		Address:             ":3000",
		BaseURL:             "http://forge.internal",
		Secret:              "dev-secret-change-me",
		CookieName:          "forge_session",
		SessionTTL:          24 * time.Hour,
		ReposRoot:           "/data/repos",
		ReadTimeout:         time.Second,
		WriteTimeout:        time.Second,
		IdleTimeout:         time.Second,
		ShutdownTimeout:     time.Second,
		RequestTimeout:      time.Second,
		MaxRequestBodyBytes: 1024,
		DBMaxOpenConns:      10,
		DBMaxIdleConns:      5,
		DBConnMaxLifetime:   time.Minute,
		DBConnMaxIdleTime:   time.Minute,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateAcceptsDevelopmentFallbacks(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:         "development",
		Address:             ":3000",
		BaseURL:             "http://localhost:3000",
		Secret:              "dev-secret-change-me",
		CookieName:          "forge_session",
		SessionTTL:          24 * time.Hour,
		ReposRoot:           "/tmp/forge/repos",
		ReadTimeout:         time.Second,
		WriteTimeout:        time.Second,
		IdleTimeout:         time.Second,
		ShutdownTimeout:     time.Second,
		RequestTimeout:      time.Second,
		MaxRequestBodyBytes: 1024,
		DBMaxOpenConns:      10,
		DBMaxIdleConns:      5,
		DBConnMaxLifetime:   time.Minute,
		DBConnMaxIdleTime:   time.Minute,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to validate: %v", err)
	}
}
