package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

type config struct {
	ServerHostPort  string        `env:"SERVER_HOST_PORT" envDefault:"localhost:8080"`
	CacheSize       int           `env:"CACHE_SIZE" envDefault:"10"`
	DefaultCacheTTL time.Duration `env:"DEFAULT_CACHE_TTL" envDefault:"1m"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"warn"`
}

func GetConfig() *config {
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		logrus.Fatalf("Failed to parse env vars: %v", err)
	}

	flag.StringVar(&cfg.ServerHostPort, "server-host-port", cfg.ServerHostPort, "Server host and port")
	flag.IntVar(&cfg.CacheSize, "cache-size", cfg.CacheSize, "Cache size")
	flag.DurationVar(&cfg.DefaultCacheTTL, "default-cache-ttl", cfg.DefaultCacheTTL, "Default cache TTL")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level")
	flag.Parse()

	return cfg
}
