package main

import (
	"lrucache/pkg/cache"
	"lrucache/pkg/config"
	"lrucache/pkg/handlers"
	"net/http"

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
	http.ListenAndServe(cfg.ServerHostPort, r)
}
