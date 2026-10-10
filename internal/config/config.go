package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port         string
	DBPath       string
	LogLevel     slog.Level
	MaxBodyBytes int64
}

func Load() (Config, error) {
	cfg := Config{
		Port:   getenv("PORT", "8080"),
		DBPath: getenv("DB_PATH", "./data.db"),
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535, got %q", cfg.Port)
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(strings.ToLower(getenv("LOG_LEVEL", "info")))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error: %w", err)
	}

	cfg.MaxBodyBytes = 1 << 20
	if raw := os.Getenv("MAX_BODY_BYTES"); raw != "" {
		cfg.MaxBodyBytes, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || cfg.MaxBodyBytes <= 0 {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES must be a positive number, got %q", raw)
		}
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
