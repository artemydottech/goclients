package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port   string
	DBPath string
}

func Load() (Config, error) {
	cfg := Config{
		Port:   getenv("PORT", "8080"),
		DBPath: getenv("DB_PATH", "./data.db"),
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func (c Config) String() string {
	return fmt.Sprintf("port=%s db=%s", c.Port, c.DBPath)
}
