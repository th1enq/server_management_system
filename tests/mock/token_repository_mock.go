package mock

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type TokenRepositoryMock struct {
	mock.Mock
}

// AddTokenToWhitelist implements repository.TokenRepository.
func (m *TokenRepositoryMock) AddTokenToWhitelist(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	args := m.Called(ctx, token, userID, expiration)
	return args.Error(0)
}

func (m *TokenRepositoryMock) IsTokenWhitelisted(ctx context.Context, token string) bool {
	args := m.Called(ctx, token)
	return args.Bool(0)
}

func (m *TokenRepositoryMock) RemoveTokenFromWhitelist(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *TokenRepositoryMock) RemoveUserTokensFromWhitelist(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
