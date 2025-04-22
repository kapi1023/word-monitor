package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

var ErrCacheMiss = errors.New("cache miss")

type Cache[T any] struct {
	c *cache.Cache
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		c: cache.New(24*time.Hour, 30*time.Minute),
	}
}

func (c *Cache[T]) Set(id string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: failed to marshal: %w", err)
	}
	c.c.Set(id, data, cache.DefaultExpiration)
	return nil
}

func (c *Cache[T]) Get(id string) (*T, error) {
	val, found := c.c.Get(id)
	if !found {
		return nil, ErrCacheMiss
	}

	data, ok := val.([]byte)
	if !ok {
		return nil, ErrCacheMiss
	}

	var t T
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("cache: failed to unmarshal: %w", err)
	}
	return &t, nil
}

func (c *Cache[T]) Has(id string) bool {
	_, found := c.c.Get(id)
	return found
}
