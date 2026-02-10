package jikan

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type memoryItem struct {
	data   []byte
	expiry int64
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]memoryItem
	stop  chan struct{}
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]memoryItem),
		stop:  make(chan struct{}),
	}
	go c.cleanup(5 * time.Minute)
	return c
}

func (c *MemoryCache) Get(ctx context.Context, key string, dst interface{}) error {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().UnixNano() > item.expiry {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return CacheMissError{}
	}
	return json.Unmarshal(item.data, dst)
}

func (c *MemoryCache) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.items[key] = memoryItem{data: b, expiry: time.Now().Add(ttl).UnixNano()}
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Stop() { close(c.stop) }

func (c *MemoryCache) cleanup(interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		select {
		case <-t.C:
			now := time.Now().UnixNano()
			c.mu.Lock()
			for k, v := range c.items {
				if now > v.expiry {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			t.Stop()
			return
		}
	}
}
