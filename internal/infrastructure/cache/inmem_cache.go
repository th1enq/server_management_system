package cache

import (
	"sync"

	"go.uber.org/zap"
)

type InMemoryCache interface {
	Set(key string, value any) error
	Get(key string) (any, error)
	Delete(key string) error
}

type inMemoryCache struct {
	cache  sync.Map
	logger *zap.Logger
}

func NewInMemoryCache(logger *zap.Logger) InMemoryCache {
	return &inMemoryCache{
		cache:  sync.Map{},
		logger: logger,
	}
}

func (c *inMemoryCache) Set(key string, value any) error {
	c.cache.Store(key, value)
	c.logger.Info("Set value in in-memory cache", zap.String("key", key))
	return nil
}

func (c *inMemoryCache) Get(key string) (any, error) {
	value, ok := c.cache.Load(key)
	if !ok {
		c.logger.Warn("Cache miss for key", zap.String("key", key))
		return nil, ErrCacheMiss
	}
	c.logger.Info("Retrieved value from in-memory cache", zap.String("key", key))
	return value, nil
}

func (c *inMemoryCache) Delete(key string) error {
	c.cache.Delete(key)
	c.logger.Info("Deleted key from in-memory cache", zap.String("key", key))
	return nil
}
