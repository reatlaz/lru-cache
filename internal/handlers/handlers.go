// Package handlers provides HTTP handlers for the LRU cache service.
package handlers

import (
	"encoding/json"
	"lrucache/pkg/cache"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

// CacheInstance is the instance of the LRU cache implementation in use.
var CacheInstance *cache.LRUCache

// CacheRequest represents the request payload for cache operations.
type CacheRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   int         `json:"ttl_seconds,omitempty"`
}

// CacheResponse represents the response payload for cache operations.
type CacheResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt int64       `json:"expires_at"`
}

// PostCacheHandler handles the creation of a new cache entry.
// It reads the key, value, and TTL from the request body and stores the data in the cache.
func PostCacheHandler(w http.ResponseWriter, r *http.Request) {
	var req CacheRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.TTL) * time.Second
	if err := CacheInstance.Put(r.Context(), req.Key, req.Value, ttl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetCacheHandler retrieves a cache entry by key.
// It responds with the value and expiration time of the cache entry.
func GetCacheHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	value, expiresAt, err := CacheInstance.Get(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := CacheResponse{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt.Unix(),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAllCacheHandler retrieves all cache entries.
// It responds with a list of keys and corresponding values.
func GetAllCacheHandler(w http.ResponseWriter, r *http.Request) {
	keys, values, err := CacheInstance.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(keys) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := map[string]interface{}{
		"keys":   keys,
		"values": values,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteCacheHandler removes a cache entry by key.
func DeleteCacheHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	_, err := CacheInstance.Evict(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAllCacheHandler removes all cache entries.
func DeleteAllCacheHandler(w http.ResponseWriter, r *http.Request) {
	if err := CacheInstance.EvictAll(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
