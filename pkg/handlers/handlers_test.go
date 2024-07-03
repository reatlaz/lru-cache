package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"lrucache/pkg/cache"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func setupTests(cacheCap int) http.Handler {
	r := chi.NewRouter()
	CacheInstance = cache.NewLRUCache(cacheCap, time.Minute)

	r.Post("/api/lru", PostCacheHandler)
	r.Get("/api/lru/{key}", GetCacheHandler)
	r.Get("/api/lru", GetAllCacheHandler)
	r.Delete("/api/lru/{key}", DeleteCacheHandler)
	r.Delete("/api/lru", DeleteAllCacheHandler)

	return r
}

func TestPostHandler(t *testing.T) {
	r := setupTests(10)

	testCases := []struct {
		id      string
		reqBody []byte
		status  int
	}{
		{
			id:      "Request with TTL",
			reqBody: []byte(`{"key": "test", "value": "helloworld", "ttl_seconds": 30}`),
			status:  http.StatusCreated,
		},
		{
			id:      "Request without TTL",
			reqBody: []byte(`{"key": "test2", "value": 9000}`),
			status:  http.StatusCreated,
		},
		{
			id:      "Bad request",
			reqBody: []byte(`{"key": "bad", "value": "helloworld", "ttl_seconds": "string"}`),
			status:  http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/lru", bytes.NewBuffer(tc.reqBody))
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tc.status, recorder.Code)
		})
	}
}

func TestGetHandler(t *testing.T) {
	r := setupTests(10)

	CacheInstance.Put(context.Background(), "hi", "helloworld", time.Minute)

	testCases := []struct {
		id     string
		key    string
		value  interface{}
		status int
	}{
		{
			id:     "Normal",
			key:    "hi",
			value:  "helloworld",
			status: http.StatusOK,
		},
		{
			id:     "Bad key",
			key:    "aboba",
			status: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/lru/"+tc.key, nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tc.status, recorder.Code)

			if tc.status == http.StatusOK {
				var resp CacheResponse
				if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tc.key, resp.Key)
				assert.Equal(t, tc.value, resp.Value)
			}
		})
	}
}

func TestGetAllHandler(t *testing.T) {

	testCases := []struct {
		id       string
		keys     []string
		values   []interface{}
		status   int
		cacheCap int
	}{
		{
			id:       "One item",
			keys:     []string{"test"},
			values:   []interface{}{"helloworld"},
			status:   http.StatusOK,
			cacheCap: 2,
		},
		{
			id:       "Zero-length cache",
			status:   http.StatusNoContent,
			cacheCap: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			r := setupTests(tc.cacheCap)

			CacheInstance.Put(context.Background(), "test", "helloworld", time.Minute)

			req, err := http.NewRequest("GET", "/api/lru", nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tc.status, recorder.Code)

			if tc.status == http.StatusOK {
				resp := make(map[string]interface{})
				if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}

				keys := resp["keys"].([]interface{})
				values := resp["values"].([]interface{})

				for i, key := range tc.keys {
					assert.Equal(t, key, keys[i].(string))
				}
				for i, value := range tc.values {
					assert.Equal(t, value, values[i])
				}
			}
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	r := setupTests(5)

	CacheInstance.Put(context.Background(), "hi", "helloworld", time.Minute)

	testCases := []struct {
		id     string
		key    string
		status int
	}{
		{
			id:     "Normal",
			key:    "hi",
			status: http.StatusNoContent,
		},
		{
			id:     "Bad key",
			key:    "aboba",
			status: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", "/api/lru/"+tc.key, nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tc.status, recorder.Code)
		})
	}
}

func TestDeleteAllHandler(t *testing.T) {
	r := setupTests(10)

	CacheInstance.Put(context.Background(), "test", "helloworld", time.Minute)

	testCases := []struct {
		id     string
		status int
	}{
		{
			id:     "Delete all keys",
			status: http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", "/api/lru", nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tc.status, recorder.Code)

			keys, _, _ := CacheInstance.GetAll(context.Background())
			assert.Equal(t, 0, len(keys))
		})
	}
}
