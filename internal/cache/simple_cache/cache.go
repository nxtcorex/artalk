package simple_cache

import (
	"sync"
	"sync/atomic"
	"time"
)

type Cache struct {
	data sync.Map
	mu   sync.Mutex
	// when the number of entries reaches gcThold, GC will be triggered
	gcThold int
	len     atomic.Int32
}

func New() *Cache {
	return &Cache{
		gcThold: 1000,
		len:     atomic.Int32{},
	}
}

func NewWithGCThold(gcThold int) *Cache {
	return &Cache{
		gcThold: gcThold,
		len:     atomic.Int32{},
	}
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

func (c *Cache) Set(key string, value interface{}, expirationParam ...time.Duration) {
	var expirationTime time.Time
	if len(expirationParam) > 0 && expirationParam[0] > 0 {
		expirationTime = time.Now().Add(expirationParam[0])
	}
	entry := cacheEntry{
		value:      value,
		expiration: expirationTime,
	}

	c.mu.Lock()
	_, exists := c.data.Load(key)
	c.data.Store(key, entry)
	if !exists {
		c.len.Add(1)
	}
	c.mu.Unlock()

	c.GC(false)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	if val, found := c.data.Load(key); found {
		entry := val.(cacheEntry)
		if !isExpired(entry) {
			c.mu.Unlock()
			return entry.value, true
		}
		c.deleteLocked(key)
	}
	c.mu.Unlock()

	c.GC(false)
	return nil, false
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	c.deleteLocked(key)
	c.mu.Unlock()
}

func (c *Cache) GC(force bool) {
	if int(c.len.Load()) < c.gcThold && !force {
		return
	}

	c.mu.Lock()
	c.data.Range(func(key, value interface{}) bool {
		entry := value.(cacheEntry)
		if !isExpired(entry) {
			return true
		}
		c.deleteLocked(key.(string))
		return true
	})
	c.mu.Unlock()
}

func (c *Cache) deleteLocked(key string) {
	if _, found := c.data.Load(key); !found {
		return
	}
	c.data.Delete(key)
	c.len.Add(-1)
}

func isExpired(entry cacheEntry) bool {
	return !entry.expiration.IsZero() && entry.expiration.Before(time.Now())
}
