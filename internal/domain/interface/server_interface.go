package interface_repository

import (
	"context"
	"server_management_system/internal/domain/dto"
	"server_management_system/internal/domain/entity"
)

type ServerRepository interface {
	Create(ctx context.Context, server *entity.Server) error
	GetByID(ctx context.Context, id uint) (*entity.Server, error)
	GetByServerID(ctx context.Context, serverID string) (*entity.Server, error)
	GetByServerName(ctx context.Context, serverName string) (*entity.Server, error)
	Update(ctx context.Context, server *entity.Server) error
	Delete(ctx context.Context, id uint) error
	BatchCreate(ctx context.Context, servers []entity.Server) error

	List(ctx context.Context, filter *dto.ServerFilter, pagination *dto.Pagination) ([]entity.Server, int64, error)
	UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus) error
	CountByStatus(ctx context.Context, status entity.ServerStatus) (int64, error)
	GetAll(ctx context.Context) ([]entity.Server, error)
}
