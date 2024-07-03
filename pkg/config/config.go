// Package config provides useful functions for the LRU cache service configuration.
package config

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/caarlos0/env"
)

type config struct {
	ServerHostPort  string        `env:"SERVER_HOST_PORT" envDefault:"localhost:8080"`
	CacheSize       int           `env:"CACHE_SIZE" envDefault:"10"`
	DefaultCacheTTL time.Duration `env:"DEFAULT_CACHE_TTL" envDefault:"1m"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"warn"`
}

// GetConfig returns the LRU Cache service configurations.
// Config values are taken from CLI flags with fallbacks to enviroment variables or default values.
func GetConfig() *config {
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		slog.Error(fmt.Sprintf("Failed to parse env vars: %v", err))
	}

	flag.StringVar(&cfg.ServerHostPort, "server-host-port", cfg.ServerHostPort, "Server host and port")
	flag.IntVar(&cfg.CacheSize, "cache-size", cfg.CacheSize, "Cache size")
	flag.DurationVar(&cfg.DefaultCacheTTL, "default-cache-ttl", cfg.DefaultCacheTTL, "Default cache TTL")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level")
	flag.Parse()

	return cfg
}

// GetSlogLevel converts the log level name to the integer-based log level used by the log/slog package.
func GetSlogLevel(logLevel string) slog.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn(fmt.Sprintf("Unknown log level: %s. Using default level `warn`\n", logLevel))
		return slog.LevelWarn
	}
}
