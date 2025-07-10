package repository

import (
	"context"
	"time"
)

type TokenRepository interface {
	AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error
	IsTokenWhitelisted(ctx context.Context, token string) bool
	RemoveTokenFromWhitelist(ctx context.Context, token string) error
	RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) error
}
