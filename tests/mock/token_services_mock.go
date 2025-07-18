package mock

import (
	"github.com/stretchr/testify/mock"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/dto"
)

type TokenServicesMock struct {
	mock.Mock
}

func (m *TokenServicesMock) GenerateAccessToken(user *entity.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *TokenServicesMock) GenerateRefreshToken(user *entity.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *TokenServicesMock) ValidateToken(tokenString string) (*dto.Claims, error) {
	args := m.Called(tokenString)
	if claims, ok := args.Get(0).(*dto.Claims); ok {
		return claims, args.Error(1)
	}
	return nil, args.Error(1)
}
