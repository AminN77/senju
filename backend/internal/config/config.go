// Package config loads process configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime settings for the API process and readiness wiring.
type Config struct {
	APIPort int
	// LogLevel is a zerolog level name (debug, info, warn, error). Empty defaults to info.
	LogLevel string

	PostgresDSN    string
	ClickHousePing string
	MinIOHealthURL string
	NATSAddr       string
}

// Load reads configuration from the environment.
func Load() (Config, error) {
	portStr := os.Getenv("API_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil {
		return Config{}, fmt.Errorf("API_PORT: parse: %w", err)
	}
	if port < 1 || port > 65535 {
		return Config{}, errors.New("API_PORT: must be a decimal integer between 1 and 65535")
	}

	chBase := strings.TrimRight(os.Getenv("CLICKHOUSE_HTTP_URL"), "/")
	var chPing string
	if chBase != "" {
		chPing = chBase + "/ping"
	}

	return Config{
		APIPort:        port,
		LogLevel:       os.Getenv("LOG_LEVEL"),
		PostgresDSN:    postgresDSN(),
		ClickHousePing: chPing,
		MinIOHealthURL: strings.TrimSpace(os.Getenv("MINIO_HEALTH_URL")),
		NATSAddr:       strings.TrimSpace(os.Getenv("NATS_ADDR")),
	}, nil
}

func postgresDSN() string {
	if d := strings.TrimSpace(os.Getenv("POSTGRES_DSN")); d != "" {
		return d
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		return ""
	}
	port := getenvDefault("POSTGRES_PORT", "5432")
	user := getenvDefault("POSTGRES_USER", "senju")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := getenvDefault("POSTGRES_DB", "senju")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, db)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
