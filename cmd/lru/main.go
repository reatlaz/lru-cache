package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"lrucache/internal/config"
	"lrucache/internal/handlers"
	"lrucache/pkg/cache"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func accessLogMiddleware(handlerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			slog.Debug(
				"Request",
				"method", r.Method,
				"path", r.URL.Path,
				"handler", handlerName,
				"duration", time.Since(start),
			)
		})
	}
}

func main() {
	cfg := config.GetConfig()
	slog.SetLogLoggerLevel(config.GetSlogLevel(cfg.LogLevel))

	handlers.CacheInstance = cache.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.With(accessLogMiddleware("PostCacheHandler")).Post("/api/lru", handlers.PostCacheHandler)
	r.With(accessLogMiddleware("GetCacheHandler")).Get("/api/lru/{key}", handlers.GetCacheHandler)
	r.With(accessLogMiddleware("GetAllCacheHandler")).Get("/api/lru", handlers.GetAllCacheHandler)
	r.With(accessLogMiddleware("DeleteCacheHandler")).Delete("/api/lru/{key}", handlers.DeleteCacheHandler)
	r.With(accessLogMiddleware("DeleteAllCacheHandler")).Delete("/api/lru", handlers.DeleteAllCacheHandler)

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
