package repository

import (
	"context"
	"time"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/query"
)

type ServerRepository interface {
	Create(ctx context.Context, server *entity.Server) error
	GetByID(ctx context.Context, id uint) (*entity.Server, error)
	GetByServerID(ctx context.Context, serverID string) (*entity.Server, error)
	GetByServerName(ctx context.Context, serverName string) (*entity.Server, error)
	Delete(ctx context.Context, id uint) error
	Update(ctx context.Context, server *entity.Server) error
	List(ctx context.Context, filter query.ServerFilter, pagination query.Pagination) ([]*entity.Server, int64, error)
	BatchCreate(ctx context.Context, servers []entity.Server) ([]*entity.Server, error)
	UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus, timestamp time.Time) error
	CountByStatus(ctx context.Context, status entity.ServerStatus) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	GetAll(ctx context.Context) ([]*entity.Server, error)
	ExistsByServerIDOrServerName(ctx context.Context, serverID string, serverName string) (bool, error)
	GetByIPv4(ctx context.Context, ipv4 string) (*entity.Server, error)
	ExecuteRawQuery(ctx context.Context, query string, args ...interface{}) error
	GetServerIDs(ctx context.Context) ([]string, error)
	GetIntervalTime(ctx context.Context, serverID string) (int64, error)
}
