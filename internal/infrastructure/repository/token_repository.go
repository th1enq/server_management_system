package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/infrastructure/cache"
)

type TokenRepository interface {
	AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error
	IsTokenWhitelisted(ctx context.Context, token string) bool
	RemoveTokenFromWhitelist(ctx context.Context, token string) error
	RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) error
}

type tokenRepository struct {
	cache cache.CacheClient
}

func NewTokenRepository(cache cache.CacheClient) TokenRepository {
	return &tokenRepository{
		cache: cache,
	}
}

func (t *tokenRepository) AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	cacheKey := fmt.Sprintf("user_tokens:%d", userID)
	err := t.cache.SADD(ctx, cacheKey, token)
	if err != nil {
		return err
	}
	cacheKey = fmt.Sprintf("token:whitelist:%s", token)
	err = t.cache.Set(ctx, cacheKey, "valid", expiration)
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

func (t *tokenRepository) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) error {
	cacheKey := fmt.Sprintf("user_tokens:%d", userID)
	token, err := t.cache.SMEMBERS(ctx, cacheKey)
	if err != nil {
		return err
	}
	for _, to := range token {
		t.cache.Del(ctx, fmt.Sprintf("token:whitelist:%s", to))
	}
	if err := t.cache.Del(ctx, cacheKey); err != nil {
		return err
	}
	return nil
}
