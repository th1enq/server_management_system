package repository

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
)

type IServerRepository interface {
	Create(ctx context.Context, server *models.Server) error
	GetByID(ctx context.Context, id uint) (*models.Server, error)
	GetByServerID(ctx context.Context, serverID string) (*models.Server, error)
	GetByServerName(ctx context.Context, serverName string) (*models.Server, error)
	Delete(ctx context.Context, id uint) error
	Update(ctx context.Context, server *models.Server) error
	List(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]models.Server, int64, error)
	BatchCreate(ctx context.Context, servers []models.Server) error
	UpdateStatus(ctx context.Context, serverID string, status models.ServerStatus) error
	CountByStatus(ctx context.Context, status models.ServerStatus) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	GetAll(ctx context.Context) ([]models.Server, error)
}

type serverRepository struct {
	db db.IDatabaseClient
}

func NewServerRepository(db db.IDatabaseClient) IServerRepository {
	return &serverRepository{
		db: db,
	}
}

func (s *serverRepository) BatchCreate(ctx context.Context, servers []models.Server) error {
	return s.db.WithContext(ctx).CreateInBatches(servers, len(servers))
}

func (s *serverRepository) CountByStatus(ctx context.Context, status models.ServerStatus) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Server{}).Where("status = ?", status).Count(&count)
	return count, err
}

func (s *serverRepository) Create(ctx context.Context, server *models.Server) error {
	return s.db.WithContext(ctx).Create(server)
}

func (s *serverRepository) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Server{}, id)
}

func (s *serverRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Server{}).Count(&count)
	return count, err
}

func (s *serverRepository) GetByID(ctx context.Context, id uint) (*models.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, id); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) GetAll(ctx context.Context) ([]models.Server, error) {
	var servers []models.Server
	if err := s.db.WithContext(ctx).Find(&servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *serverRepository) GetByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).Where("server_id = ?", serverID).First(&server); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) GetByServerName(ctx context.Context, serverName string) (*models.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).Where("server_name = ?", serverName).First(&server); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) List(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]models.Server, int64, error) {
	var servers []models.Server
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Server{})

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

func (s *serverRepository) Update(ctx context.Context, server *models.Server) error {
	return s.db.WithContext(ctx).Save(server)
}

func (s *serverRepository) UpdateStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	return s.db.WithContext(ctx).Model(&models.Server{}).Where("server_id = ?", serverID).Update("status", status)
}
