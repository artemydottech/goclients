package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_PATH", "")

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
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("DB_PATH", "/tmp/test.db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Port != "9000" || cfg.DBPath != "/tmp/test.db" {
		t.Errorf("unexpected config %+v", cfg)
	}
}
