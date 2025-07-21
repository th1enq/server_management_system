package services

import (
	"context"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/dto"
)

type TokenServices interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	GenerateServerAccessToken(ctx context.Context, server *entity.Server) (string, error)
	GenerateServerRefreshToken(ctx context.Context, server *entity.Server) (string, error)
	ValidateToken(tokenString string) (*dto.Claims, error)
}
