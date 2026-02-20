// Package cache implements a simple in memory cache
package cache

import (
	"sync"
	"time"
)

type item struct {
	value      any
	expireTime time.Time
}

type Cache struct {
	data map[string]item
	mu   sync.RWMutex
}

func New() *Cache {
	c := &Cache{
		data: make(map[string]item),
	}
	// Cleanup expired items every 1 minute
	go c.cleanupExpired()
	return c
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = item{
		value:      value,
		expireTime: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	itm, found := c.data[key]
	if !found {
		return nil, false
	}

	if time.Now().After(itm.expireTime) {
		return nil, false
	}

	return itm.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, itm := range c.data {
			if now.After(itm.expireTime) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
