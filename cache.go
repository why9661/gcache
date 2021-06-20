package gocache

import (
	"github.com/why9661/gocache/lru"
	"log"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Lru
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

func (c *cache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.lru.Get(key); !ok {
		log.Printf("Key: %v dose not exist", key)
	} else {
		c.lru.Remove(key)
	}

}
