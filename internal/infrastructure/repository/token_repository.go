package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/cache"
)

type tokenRepository struct {
	cache cache.CacheClient
}

func NewTokenRepository(cache cache.CacheClient) repository.TokenRepository {
	return &tokenRepository{
		cache: cache,
	}
}

func (t *tokenRepository) AddTokenToWhitelist(ctx context.Context, token string, expiration time.Duration) error {
	cacheKey := fmt.Sprintf("token:whitelist:%s", token)
	err := t.cache.Set(ctx, cacheKey, "valid", expiration)
	if err != nil {
		return err
	}
	return nil
}

func (t *tokenRepository) IsTokenWhitelisted(ctx context.Context, token string) bool {
	key := fmt.Sprintf("token:whitelist:%s", token)
	var valid string
	err := t.cache.Get(ctx, key, &valid)
	if err == nil && valid == "valid" {
		return true
	}
	return false
}

func (t *tokenRepository) RemoveTokenFromWhitelist(ctx context.Context, token string) error {
	key := fmt.Sprintf("token:whitelist:%s", token)
	return t.cache.Del(ctx, key)
}
