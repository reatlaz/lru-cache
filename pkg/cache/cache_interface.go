package cache

import (
	"context"
	"time"
)

// ILRUCache provides methods for adding, retrieving, and deleting cache entries.
type ILRUCache interface {
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error)
	GetAll(ctx context.Context) (keys []string, values []interface{}, err error)
	Evict(ctx context.Context, key string) (value interface{}, err error)
	EvictAll(ctx context.Context) error
}
