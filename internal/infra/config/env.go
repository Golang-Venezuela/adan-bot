// Package config provides configuration and environment variable management
// utilities for the application, handling defaults and local `.env` file loading.
package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Getenv retrieves the value of the environment variable named by the key.
// If the variable is present, its associated value is securely returned.
// Otherwise, it logs a missing variable warning and gracefully yields the provided default value.
func Getenv(key, def string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}

	slog.Warn("Missing environment variable, using default", slog.String("key", key), slog.String("default", def))

	return def
}

// init systematically attempts to load runtime environment variables from a local `.env` file
// upon package initialization. It logs a soft warning if the file cannot be found or read.
func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Cannot load envfile", slog.Any("error", err))
	}
}
