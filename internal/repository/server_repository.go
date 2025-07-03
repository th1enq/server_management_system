package repository

import (
	"context"
	"fmt"
	"server_management_system/internal/domain/dto"
	"server_management_system/internal/domain/entity"
	interface_repository "server_management_system/internal/domain/interface"

	"gorm.io/gorm"
)

type serverRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) interface_repository.ServerRepository {
	return &serverRepository{
		db: db,
	}
}

func (s *serverRepository) CountByStatus(ctx context.Context, status entity.ServerStatus) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Server{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (s *serverRepository) Create(ctx context.Context, server *entity.Server) error {
	return s.db.WithContext(ctx).Create(server).Error
}

func (s *serverRepository) BatchCreate(ctx context.Context, servers []entity.Server) error {
	return s.db.CreateInBatches(servers, 100).Error
}

func (s *serverRepository) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&entity.Server{}, id).Error
}

func (s *serverRepository) GetAll(ctx context.Context) ([]entity.Server, error) {
	var servers []entity.Server
	err := s.db.WithContext(ctx).Find(&servers).Error
	return servers, err
}

func (s *serverRepository) GetByID(ctx context.Context, id uint) (*entity.Server, error) {
	var server entity.Server
	if err := s.db.WithContext(ctx).First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) GetByServerID(ctx context.Context, serverID string) (*entity.Server, error) {
	var server entity.Server
	if err := s.db.WithContext(ctx).Where("server_id = ?", serverID).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) GetByServerName(ctx context.Context, serverName string) (*entity.Server, error) {
	var server entity.Server
	if err := s.db.WithContext(ctx).Where("server_name = ?", serverName).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *serverRepository) List(ctx context.Context, filter *dto.ServerFilter, pagination *dto.Pagination) ([]entity.Server, int64, error) {
	var servers []entity.Server
	var total int64

	query := s.db.WithContext(ctx).Model(&entity.Server{})

	// Apply filters
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	orderBy := fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)

	err := query.
		Order(orderBy).
		Limit(pagination.PageSize).
		Offset(offset).
		Find(&servers).Error

	return servers, total, err
}

func (s *serverRepository) Update(ctx context.Context, server *entity.Server) error {
	return s.db.WithContext(ctx).Save(server).Error
}

func (s *serverRepository) UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus) error {
	return s.db.WithContext(ctx).Model(&entity.Server{}).Where("server_id = ?", serverID).Update("status", status).Error
}
