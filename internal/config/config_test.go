package config

import (
	"log/slog"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"PORT", "DB_PATH", "LOG_LEVEL", "MAX_BODY_BYTES"} {
		t.Setenv(key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.DBPath != "./data.db" {
		t.Errorf("DBPath = %q, want ./data.db", cfg.DBPath)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.MaxBodyBytes != 1<<20 {
		t.Errorf("MaxBodyBytes = %d, want 1 MiB", cfg.MaxBodyBytes)
	}
}

func TestLoadReadsEnvironment(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9000")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("MAX_BODY_BYTES", "2048")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Port != "9000" || cfg.DBPath != "/tmp/test.db" {
		t.Errorf("unexpected config %+v", cfg)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.MaxBodyBytes != 2048 {
		t.Errorf("MaxBodyBytes = %d, want 2048", cfg.MaxBodyBytes)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"port is not a number", "PORT", "http"},
		{"port is zero", "PORT", "0"},
		{"port is too large", "PORT", "70000"},
		{"unknown log level", "LOG_LEVEL", "chatty"},
		{"body limit is not a number", "MAX_BODY_BYTES", "big"},
		{"body limit is negative", "MAX_BODY_BYTES", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(tt.key, tt.value)

			if _, err := Load(); err == nil {
				t.Fatalf("expected an error for %s=%s", tt.key, tt.value)
			}
		})
	}
}
