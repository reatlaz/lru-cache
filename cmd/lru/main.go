package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"lrucache/internal/config"
	"lrucache/internal/handlers"
	"lrucache/pkg/cache"
	"lrucache/pkg/middlewares"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.GetConfig()
	slog.SetLogLoggerLevel(config.GetSlogLevel(cfg.LogLevel))

	handlers.CacheInstance = cache.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middlewares.AccessLogMiddleware)

	slog.Info(fmt.Sprintf("Starting server on %s", cfg.ServerHostPort))

	server := &http.Server{
		Addr:    cfg.ServerHostPort,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error(fmt.Sprintf("HTTP server error: %v", err))
			os.Exit(1)
		}
		slog.Info("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	slog.Info(fmt.Sprintf("Received signal: %s", <-sigChan))
	start := time.Now()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error(fmt.Sprintf("Server shutdown error: %v", err))
	}
	elapsed := time.Since(start)
	slog.Info(fmt.Sprintf("Graceful shutdown completed in %v", elapsed))
}
