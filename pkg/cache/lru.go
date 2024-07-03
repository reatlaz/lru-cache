// Package cache provides an LRU cache implementation.
package cache

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"
)

// CacheItem stores the data for an LRU Cache entry.
type CacheItem struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
}

// LRUCache structure implements a cache with the Least-Recently-Used eviction policy.
type LRUCache struct {
	capacity  int
	ttl       time.Duration
	mutex     sync.RWMutex
	items     map[string]*list.Element
	orderList *list.List
}

// NewLRUCache creates an LRUCache instance with specified capacity and ttl.
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		orderList: list.New(),
	}
}

// Put adds a value to the cache with the specified key and TTL.
func (c *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ttl == 0 {
		ttl = c.ttl
	}

	if el, ok := c.items[key]; ok {
		c.orderList.MoveToFront(el)
		item := el.Value.(*CacheItem)
		item.Value = value
		item.ExpiresAt = time.Now().Add(ttl)
	} else {
		if c.orderList.Len() >= c.capacity {
			oldest := c.orderList.Back()
			if oldest == nil {
				// zero-capacity cache
				return nil
			}
			c.orderList.Remove(oldest)
			delete(c.items, oldest.Value.(*CacheItem).Key)

		}
		item := &CacheItem{
			Key:       key,
			Value:     value,
			ExpiresAt: time.Now().Add(ttl),
		}
		el := c.orderList.PushFront(item)
		c.items[key] = el
	}

	return nil
}

// Get retrieves a value from the cache by key.
func (c *LRUCache) Get(ctx context.Context, key string) (interface{}, time.Time, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if el, ok := c.items[key]; ok {
		item := el.Value.(*CacheItem)
		if item.ExpiresAt.After(time.Now()) {
			c.orderList.MoveToFront(el)
			return item.Value, item.ExpiresAt, nil
		}
		c.orderList.Remove(el)
		delete(c.items, key)
	}

	return nil, time.Time{}, errors.New("key not found")
}

// GetAll retrieves all entries in the cache.
func (c *LRUCache) GetAll(ctx context.Context) ([]string, []interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	values := make([]interface{}, 0, len(c.items))
	var next *list.Element
	for el := c.orderList.Front(); el != nil; el = next {
		item := el.Value.(*CacheItem)
		if item.ExpiresAt.After(time.Now()) {
			keys = append(keys, item.Key)
			values = append(values, item.Value)
			next = el.Next()
		} else {
			next = el.Next()
			c.orderList.Remove(el)
			delete(c.items, item.Key)
		}
	}

	return keys, values, nil
}

// Evict removes an entry from the cache by key.
func (c *LRUCache) Evict(ctx context.Context, key string) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if el, ok := c.items[key]; ok {
		item := el.Value.(*CacheItem)
		c.orderList.Remove(el)
		delete(c.items, key)
		return item.Value, nil
	}

	return nil, errors.New("key not found")
}

// EvictAll removes all entries from the cache.
func (c *LRUCache) EvictAll(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*list.Element)
	c.orderList.Init()

	return nil
}
