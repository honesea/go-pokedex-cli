package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	store map[string]cacheEntry
	mu    *sync.RWMutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(staleLimit time.Duration) Cache {
	newCache := Cache{
		store: make(map[string]cacheEntry),
		mu:    &sync.RWMutex{},
	}
	go newCache.reapLoop(staleLimit)
	return newCache
}

func (c Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[key]

	if !ok {
		return []byte{}, false
	}

	return entry.val, true
}

func (c Cache) reapLoop(staleLimit time.Duration) {
	for {
		time.Sleep(staleLimit)

		c.mu.Lock()
		for key, entry := range c.store {
			if time.Since(entry.createdAt) > staleLimit {
				delete(c.store, key)
			}
		}
		c.mu.Unlock()
	}
}
