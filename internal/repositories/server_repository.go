package repositories

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/models"
	"gorm.io/gorm"
)

type ServerRepository interface {
	Create(ctx context.Context, server *models.Server) error
	GetByID(ctx context.Context, id uint) (*models.Server, error)
	GetByServerID(ctx context.Context, serverID string) (*models.Server, error)
	GetByServerName(ctx context.Context, serverName string) (*models.Server, error)
	List(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) ([]models.Server, int64, error)
	Update(ctx context.Context, server *models.Server) error
	Delete(ctx context.Context, id uint) error
	BatchCreate(ctx context.Context, servers []models.Server) ([]models.Server, []models.Server, error)
	UpdateStatus(ctx context.Context, serverID string, status models.ServerStatus) error
	CountByStatus(ctx context.Context, status models.ServerStatus) (int64, error)
	GetAll(ctx context.Context) ([]models.Server, error)
}

type serverRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepository{db: db}
}

// BatchCreate implements ServerRepository.
func (s *serverRepository) BatchCreate(ctx context.Context, servers []models.Server) ([]models.Server, []models.Server, error) {
	panic("unimplemented")
}

// CountByStatus implements ServerRepository.
func (s *serverRepository) CountByStatus(ctx context.Context, status models.ServerStatus) (int64, error) {
	panic("unimplemented")
}

// Create implements ServerRepository.
func (s *serverRepository) Create(ctx context.Context, server *models.Server) error {
	return s.db.WithContext(ctx).Create(server).Error
}

// Delete implements ServerRepository.
func (s *serverRepository) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Server{}, id).Error
}

// GetAll implements ServerRepository.
func (s *serverRepository) GetAll(ctx context.Context) ([]models.Server, error) {
	panic("unimplemented")
}

// GetByID implements ServerRepository.
func (s *serverRepository) GetByID(ctx context.Context, id uint) (*models.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// GetByServerID implements ServerRepository.
func (s *serverRepository) GetByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	var server *models.Server
	if err := s.db.WithContext(ctx).Where("server_id = ?", serverID).First(&server).Error; err != nil {
		return nil, err
	}
	return server, nil
}

// GetByServerName implements ServerRepository.
func (s *serverRepository) GetByServerName(ctx context.Context, serverName string) (*models.Server, error) {
	var server *models.Server
	if err := s.db.WithContext(ctx).Where("server_name = ?", serverName).First(&server).Error; err != nil {
		return nil, err
	}
	return server, nil
}

// List implements ServerRepository.
func (s *serverRepository) List(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) ([]models.Server, int64, error) {
	var servers []models.Server
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Server{})

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

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	offset := (pagination.Page - 1) * pagination.PageSize
	orderBy := fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)

	err := query.
		Order(orderBy).
		Limit(pagination.PageSize).
		Offset(offset).
		Find(&servers).Error

	return servers, total, err
}

// Update implements ServerRepository.
func (s *serverRepository) Update(ctx context.Context, server *models.Server) error {
	return s.db.WithContext(ctx).Save(server).Error
}

// UpdateStatus implements ServerRepository.
func (s *serverRepository) UpdateStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	panic("unimplemented")
}
