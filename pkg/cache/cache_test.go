package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacheImplementsInterface(t *testing.T) {
	var _ ILRUCache = new(LRUCache)
}

func TestCacheEmpty(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(0, time.Minute)

	c.Put(ctx, "1", interface{}(2), 0)
	_, _, err := c.Get(ctx, "1")
	require.Error(t, err)
}

func TestCachePut(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(1, time.Minute)

	_, _, err := c.Get(ctx, "1")
	require.Error(t, err)

	c.Put(ctx, "1", interface{}(2), 2*time.Minute)
	v, _, err := c.Get(ctx, "1")
	require.Nil(t, err)
	require.Equal(t, 2, v.(int))
}

func TestCachePutUpdate(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(1, time.Minute)

	c.Put(ctx, "1", interface{}(2), 2*time.Minute)
	v, _, err := c.Get(ctx, "1")
	require.Nil(t, err)
	require.Equal(t, 2, v.(int))

	c.Put(ctx, "1", interface{}(3), 10*time.Hour)
	v, _, err = c.Get(ctx, "1")
	require.Nil(t, err)
	require.Equal(t, 3, v.(int))
}

func TestCacheExpired(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(1, time.Minute)

	c.Put(ctx, "1", interface{}(2), 1*time.Second)
	time.Sleep(2 * time.Second)
	_, _, err := c.Get(ctx, "1")
	require.Error(t, err)
}

func TestCacheGet(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(5, time.Minute)

	for i := 0; i < 5; i++ {
		c.Put(ctx, strconv.Itoa(i), interface{}(i), time.Minute)
	}

	c.Get(ctx, "0")
	c.Get(ctx, "1")
	c.Put(ctx, "5", interface{}(5), time.Minute)
	c.Put(ctx, "6", interface{}(6), time.Minute)

	keys, values, err := c.GetAll(ctx)
	require.Nil(t, err)
	require.Equal(t, []string{"6", "5", "1", "0", "4"}, keys)
	require.Equal(t, []interface{}{6, 5, 1, 0, 4}, values)
}

func TestCacheEvict(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(1, time.Minute)

	c.Put(ctx, "1", interface{}(2), 2*time.Minute)

	_, _, err := c.Get(ctx, "1")
	require.Nil(t, err)

	v, err := c.Evict(ctx, "1")
	require.Nil(t, err)
	require.Equal(t, 2, v.(int))

	_, _, err = c.Get(ctx, "1")
	require.Error(t, err)

	_, err = c.Evict(ctx, "1")
	require.Error(t, err)
}

func TestCacheEvictAll(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(5, time.Minute)

	for i := 0; i < 10; i++ {
		c.Put(ctx, strconv.Itoa(i), interface{}(i), time.Minute)
	}

	c.EvictAll(ctx)

	for i := 9; i >= 0; i-- {
		c.Put(ctx, strconv.Itoa(i), interface{}(i), time.Minute)
	}

	keys, values, err := c.GetAll(ctx)
	require.Nil(t, err)
	require.Equal(t, []string{"0", "1", "2", "3", "4"}, keys)
	require.Equal(t, []interface{}{0, 1, 2, 3, 4}, values)

	// intValues := make([]int, 0, 5)
	// for _, value := range values {
	// 	intValues = append(intValues, value.(int))
	// }
	// require.Equal(t, []int{0, 1, 2, 3, 4}, intValues)
}

func TestCacheEvictAllLogic(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(5, time.Minute)

	for i := 0; i < 10; i++ {
		c.Put(ctx, strconv.Itoa(i), interface{}(i), time.Minute)
	}

	c.EvictAll(ctx)

	for i := 0; i < 10; i++ {
		_, _, err := c.Get(ctx, strconv.Itoa(i))
		require.Error(t, err)
	}

	c.EvictAll(ctx)

	for i := 3; i >= 0; i-- {
		c.Put(ctx, strconv.Itoa(i), interface{}(i), time.Minute)
	}

	keys, values, err := c.GetAll(ctx)

	require.Nil(t, err)

	require.Equal(t, []string{"0", "1", "2", "3"}, keys)
	require.Equal(t, []interface{}{0, 1, 2, 3}, values)
}

func TestCacheGetAll(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(5, time.Minute)

	for i := 0; i < 10; i++ {
		c.Put(ctx, strconv.Itoa(i), interface{}(i), time.Minute)
	}

	keys, values, err := c.GetAll(ctx)
	require.Nil(t, err)

	require.Equal(t, []string{"9", "8", "7", "6", "5"}, keys)
	require.Equal(t, []interface{}{9, 8, 7, 6, 5}, values)
}

func TestCacheGetAllExpiration(t *testing.T) {
	ctx := context.Background()
	c := NewLRUCache(5, time.Minute)

	c.Put(ctx, "1", interface{}(1), time.Minute)
	c.Put(ctx, "2", interface{}(2), time.Microsecond)
	time.Sleep(2 * time.Microsecond)

	keys, values, err := c.GetAll(ctx)
	require.Nil(t, err)

	require.Equal(t, []string{"1"}, keys)
	require.Equal(t, []interface{}{1}, values)
}
