package services

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
)

type ServerService interface {
	CreateServer(ctx context.Context, server *models.Server) error
	GetServer(ctx context.Context, id uint) (*models.Server, error)
	GetServerByServerID(ctx context.Context, serverID string) (*models.Server, error)
	ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error)
	UpdateServer(ctx context.Context, id uint, updates map[string]interface{}) (*models.Server, error)
	DeleteServer(ctx context.Context, id uint) error
	ImportServers(ctx context.Context, filePath string) (*models.ImportResult, error)
	ExportServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (string, error)
	UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error
	GetServerStats(ctx context.Context) (map[string]int64, error)
}

type serverService struct {
	serverRepo  repositories.ServerRepository
	redisClient *redis.Client
}

func NewServerService(serverRepo repositories.ServerRepository, redisClient *redis.Client) ServerService {
	return &serverService{
		serverRepo:  serverRepo,
		redisClient: redisClient,
	}
}

// CreateServer implements ServerService.
func (s *serverService) CreateServer(ctx context.Context, server *models.Server) error {
	panic("unimplemented")
}

// DeleteServer implements ServerService.
func (s *serverService) DeleteServer(ctx context.Context, id uint) error {
	panic("unimplemented")
}

// ExportServers implements ServerService.
func (s *serverService) ExportServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (string, error) {
	panic("unimplemented")
}

// GetServer implements ServerService.
func (s *serverService) GetServer(ctx context.Context, id uint) (*models.Server, error) {
	panic("unimplemented")
}

// GetServerByServerID implements ServerService.
func (s *serverService) GetServerByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	panic("unimplemented")
}

// GetServerStats implements ServerService.
func (s *serverService) GetServerStats(ctx context.Context) (map[string]int64, error) {
	panic("unimplemented")
}

// ImportServers implements ServerService.
func (s *serverService) ImportServers(ctx context.Context, filePath string) (*models.ImportResult, error) {
	panic("unimplemented")
}

// ListServers implements ServerService.
func (s *serverService) ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error) {
	panic("unimplemented")
}

// UpdateServer implements ServerService.
func (s *serverService) UpdateServer(ctx context.Context, id uint, updates map[string]interface{}) (*models.Server, error) {
	panic("unimplemented")
}

// UpdateServerStatus implements ServerService.
func (s *serverService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	panic("unimplemented")
}
