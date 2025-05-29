package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
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
	if server.ServerID == "" || server.ServerName == "" {
		return fmt.Errorf("server_id and server_name are required")
	}
	existing, _ := s.serverRepo.GetByServerID(ctx, server.ServerID)
	if existing != nil {
		return fmt.Errorf("server is already exists")
	}

	existing, _ = s.serverRepo.GetByServerName(ctx, server.ServerName)
	if existing != nil {
		return fmt.Errorf("server is already exists")
	}

	err := s.serverRepo.Create(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	logger.Info("Server created successfully",
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return nil
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
	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("server not found")
	}

	delete(updates, "server_id")
	delete(updates, "id")

	for key, value := range updates {
		switch key {
		case "server_name":
			if name, ok := value.(string); ok && name != server.ServerName {
				existing, _ := s.serverRepo.GetByServerName(ctx, name)
				if existing != nil && existing.ID != server.ID {
					return nil, fmt.Errorf("server with name is already exists")
				}
				server.ServerName = name
			}
		case "status":
			if status, ok := value.(string); ok {
				server.Status = models.ServerStatus(status)
			}
		case "ipv4":
			if ipv4, ok := value.(string); ok {
				server.IPv4 = ipv4
			}
		case "location":
			if loc, ok := value.(string); ok {
				server.Location = loc
			}
		case "os":
			if os, ok := value.(string); ok {
				server.OS = os
			}
		case "cpu":
			if cpu, ok := value.(float64); ok {
				server.CPU = int(cpu)
			}
		case "ram":
			if ram, ok := value.(float64); ok {
				server.RAM = int(ram)
			}
		case "disk":
			if disk, ok := value.(float64); ok {
				server.Disk = int(disk)
			}
		}
	}
	if err := s.serverRepo.Update(ctx, server); err != nil {
		return nil, err
	}
	logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
	)
	return server, nil
}

// UpdateServerStatus implements ServerService.
func (s *serverService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	panic("unimplemented")
}
