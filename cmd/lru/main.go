package main

import (
	"context"
	"errors"
	"lrucache/pkg/cache"
	"lrucache/pkg/config"
	"lrucache/pkg/handlers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.GetConfig()
	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.Fatalf("Invalid log level: %v", err)
	} else {
		logrus.SetLevel(level)
	}

	handlers.CacheInstance = cache.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/lru", handlers.PostCacheHandler)
	r.Get("/api/lru/{key}", handlers.GetCacheHandler)
	r.Get("/api/lru", handlers.GetAllCacheHandler)
	r.Delete("/api/lru/{key}", handlers.DeleteCacheHandler)
	r.Delete("/api/lru", handlers.DeleteAllCacheHandler)

	logrus.Infof("Starting server on %s", cfg.ServerHostPort)
	// http.ListenAndServe(cfg.ServerHostPort, r)

	server := &http.Server{
		Addr:    cfg.ServerHostPort,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("HTTP server error: %v", err)
		}
		logrus.Info("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	logrus.Infof("Received signal: %s", <-sigChan)
	start := time.Now()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Fatalf("Server shutdown error: %v", err)
	}
	elapsed := time.Since(start)
	logrus.Infof("Graceful shutdown completed in %v", elapsed)
}
