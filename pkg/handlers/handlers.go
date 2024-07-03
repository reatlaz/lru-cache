package handlers

import (
	"encoding/json"
	"lrucache/pkg/cache"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

var CacheInstance *cache.LRUCache

type CacheRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   int         `json:"ttl_seconds,omitempty"`
}

type CacheResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt int64       `json:"expires_at"`
}

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
	json.NewEncoder(w).Encode(response)
}

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
	json.NewEncoder(w).Encode(response)
}

func DeleteCacheHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	_, err := CacheInstance.Evict(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllCacheHandler(w http.ResponseWriter, r *http.Request) {
	if err := CacheInstance.EvictAll(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
