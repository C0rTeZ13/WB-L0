package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type GoCache struct {
	c *cache.Cache
}

func New(defaultTTL, cleanupInterval time.Duration) *GoCache {
	return &GoCache{
		c: cache.New(defaultTTL, cleanupInterval),
	}
}

func (g *GoCache) Get(key string) (interface{}, bool) {
	return g.c.Get(key)
}

func (g *GoCache) Set(key string, value interface{}, ttl time.Duration) {
	g.c.Set(key, value, ttl)
}

func (g *GoCache) Delete(key string) {
	g.c.Delete(key)
}
