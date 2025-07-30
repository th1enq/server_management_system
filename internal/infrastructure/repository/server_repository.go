package repository

import (
	"context"
	"fmt"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/query"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/models"
)

type serverRepository struct {
	db database.DatabaseClient
}

func NewServerRepository(db database.DatabaseClient) repository.ServerRepository {
	return &serverRepository{
		db: db,
	}
}

func (s *serverRepository) GetIntervalTime(ctx context.Context, serverID string) (int64, error) {
	var interval int64
	err := s.db.WithContext(ctx).Model(&models.Server{}).
		Select("interval_time").
		Where("server_id = ?", serverID).
		Scan(&interval)
	if err != nil {
		return 0, err
	}
	return interval, nil
}

func (s *serverRepository) GetServerIDs(ctx context.Context) ([]string, error) {
	var serverIDs []string
	err := s.db.WithContext(ctx).Model(&models.Server{}).Pluck("server_id", &serverIDs)
	if err != nil {
		return nil, err
	}
	return serverIDs, nil
}

func (s *serverRepository) GetByIPv4(ctx context.Context, ipv4 string) (*entity.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).Where("ipv4 = ?", ipv4).First(&server); err != nil {
		return nil, err
	}
	return models.ToServerEntity(&server), nil
}

func (s *serverRepository) BatchCreate(ctx context.Context, servers []entity.Server) ([]*entity.Server, error) {
	var modelsServers []models.Server
	model := models.FromServerEntities(servers)

	err := s.db.WithContext(ctx).BatchCreateOnConflict(model, &modelsServers)

	if err != nil {
		return nil, err
	}
	return models.ToServerEntities(modelsServers), nil
}

func (s *serverRepository) CountByStatus(ctx context.Context, status entity.ServerStatus) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Server{}).Where("status = ?", status).Count(&count)
	return count, err
}

func (s *serverRepository) Create(ctx context.Context, server *entity.Server) error {
	model := models.FromServerEntity(server)
	return s.db.WithContext(ctx).Create(model)
}

func (s *serverRepository) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Server{}, id)
}

func (s *serverRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Server{}).Count(&count)
	return count, err
}

func (s *serverRepository) GetByID(ctx context.Context, id uint) (*entity.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).First(&server, id); err != nil {
		return nil, err
	}
	return models.ToServerEntity(&server), nil
}

func (s *serverRepository) GetAll(ctx context.Context) ([]*entity.Server, error) {
	var servers []models.Server
	if err := s.db.WithContext(ctx).Find(&servers); err != nil {
		return nil, err
	}
	return models.ToServerEntities(servers), nil
}

func (s *serverRepository) GetByServerID(ctx context.Context, serverID string) (*entity.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).Where("server_id = ?", serverID).First(&server); err != nil {
		return nil, err
	}
	return models.ToServerEntity(&server), nil
}

func (s *serverRepository) GetByServerName(ctx context.Context, serverName string) (*entity.Server, error) {
	var server models.Server
	if err := s.db.WithContext(ctx).Where("server_name = ?", serverName).First(&server); err != nil {
		return nil, err
	}
	return models.ToServerEntity(&server), nil
}

func (s *serverRepository) List(ctx context.Context, filter query.ServerFilter, pagination query.Pagination) ([]*entity.Server, int64, error) {
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
	if filter.CPU > 0 {
		query = query.Where("cpu = ?", filter.CPU)
	}

	if filter.RAM > 0 {
		query = query.Where("ram = ?", filter.RAM)
	}

	if filter.Disk > 0 {
		query = query.Where("disk = ?", filter.Disk)
	}

	if err := query.Count(&total); err != nil {
		return nil, 0, err
	}

	offset := pagination.Offset()
	orderBy := fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)

	err := query.
		Order(orderBy).
		Limit(pagination.PageSize).
		Offset(offset).
		Find(&servers)

	return models.ToServerEntities(servers), total, err
}

func (s *serverRepository) Update(ctx context.Context, server *entity.Server) error {
	model := models.FromServerEntity(server)
	return s.db.WithContext(ctx).Save(model)
}

func (s *serverRepository) UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus) error {
	var server models.Server
	if err := s.db.WithContext(ctx).Model(&server).Where("server_id = ?", serverID).Update("status", status); err != nil {
		return err
	}
	return nil
}

func (s *serverRepository) ExistsByServerIDOrServerName(ctx context.Context, serverID string, serverName string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Server{}).
		Where("server_id = ? OR server_name = ?", serverID, serverName).
		Count(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *serverRepository) ExecuteRawQuery(ctx context.Context, query string, args ...interface{}) error {
	return s.db.WithContext(ctx).Exec(query, args...)
}
