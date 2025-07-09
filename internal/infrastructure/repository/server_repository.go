package repository

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
)

type ServerRepository interface {
	Create(ctx context.Context, server *domain.Server) error
	GetByID(ctx context.Context, id uint) (*domain.Server, error)
	GetByServerID(ctx context.Context, serverID string) (*domain.Server, error)
	GetByServerName(ctx context.Context, serverName string) (*domain.Server, error)
	Delete(ctx context.Context, id uint) error
	Update(ctx context.Context, server *domain.Server) error
	List(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]domain.Server, int64, error)
	BatchCreate(ctx context.Context, servers []domain.Server) error
	UpdateStatus(ctx context.Context, serverID string, status domain.ServerStatus) error
	CountByStatus(ctx context.Context, status domain.ServerStatus) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	GetAll(ctx context.Context) ([]domain.Server, error)
	ExistsByServerIDOrServerName(ctx context.Context, serverID string, serverName string) (bool, error)
}

type serverRepository struct {
	db database.DatabaseClient
}

func NewServerRepository(db database.DatabaseClient) ServerRepository {
	return &serverRepository{
		db: db,
	}
}

func (s *serverRepository) BatchCreate(ctx context.Context, servers []domain.Server) error {
	return s.db.WithContext(ctx).CreateInBatches(servers, len(servers))
}

func (s *serverRepository) CountByStatus(ctx context.Context, status domain.ServerStatus) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&domain.Server{}).Where("status = ?", status).Count(&count)
	return count, err
}

func (s *serverRepository) Create(ctx context.Context, server *domain.Server) error {
	return s.db.WithContext(ctx).Create(server)
}

func (s *serverRepository) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&domain.Server{}, id)
}

func (s *serverRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&domain.Server{}).Count(&count)
	return count, err
}

func (s *serverRepository) GetByID(ctx context.Context, id uint) (*domain.Server, error) {
	var server domain.Server
	if err := s.db.WithContext(ctx).First(&server, id); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) GetAll(ctx context.Context) ([]domain.Server, error) {
	var servers []domain.Server
	if err := s.db.WithContext(ctx).Find(&servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *serverRepository) GetByServerID(ctx context.Context, serverID string) (*domain.Server, error) {
	var server domain.Server
	if err := s.db.WithContext(ctx).Where("server_id = ?", serverID).First(&server); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) GetByServerName(ctx context.Context, serverName string) (*domain.Server, error) {
	var server domain.Server
	if err := s.db.WithContext(ctx).Where("server_name = ?", serverName).First(&server); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) List(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]domain.Server, int64, error) {
	var servers []domain.Server
	var total int64

	query := s.db.WithContext(ctx).Model(&domain.Server{})

	if filter.ServerID != "" {
		query = query.Where("server_id LIKE ?", "%"+filter.ServerID+"%")
	}
	if filter.ServerName != "" {
		query = query.Where("server_name LIKE ?", "%"+filter.ServerName+"%")
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.IPv4 != "" {
		query = query.Where("ipv4 LIKE ?", "%"+filter.IPv4+"%")
	}
	if filter.Location != "" {
		query = query.Where("location LIKE ?", "%"+filter.Location+"%")
	}
	if filter.OS != "" {
		query = query.Where("os LIKE ?", "%"+filter.OS+"%")
	}

	if err := query.Count(&total); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	orderBy := fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)

	err := query.
		Order(orderBy).
		Limit(pagination.PageSize).
		Offset(offset).
		Find(&servers)

	return servers, total, err
}

func (s *serverRepository) Update(ctx context.Context, server *domain.Server) error {
	return s.db.WithContext(ctx).Save(server)
}

func (s *serverRepository) UpdateStatus(ctx context.Context, serverID string, status domain.ServerStatus) error {
	return s.db.WithContext(ctx).Model(&domain.Server{}).Where("server_id = ?", serverID).Update("status", status)
}

func (s *serverRepository) ExistsByServerIDOrServerName(ctx context.Context, serverID string, serverName string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&domain.Server{}).
		Where("server_id = ? OR server_name = ?", serverID, serverName).
		Count(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
