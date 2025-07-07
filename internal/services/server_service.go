package services

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/th1enq/server_management_system/internal/db"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/repository"
	"github.com/th1enq/server_management_system/internal/utils"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type IServerService interface {
	CreateServer(ctx context.Context, server *models.Server) error
	GetServer(ctx context.Context, id uint) (*models.Server, error)
	ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (*dto.ServerListResponse, error)
	UpdateServer(ctx context.Context, id uint, updates dto.ServerUpdate) (*models.Server, error)
	DeleteServer(ctx context.Context, id uint) error
	ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error)
	ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error)
	UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error
	GetServerStats(ctx context.Context) (map[string]int64, error)
	CheckServerStatus(ctx context.Context) error
	CheckServer(ctx context.Context, server models.Server) error
	GetAllServers(ctx context.Context) ([]models.Server, error)
	ClearServerCaches(ctx context.Context, server *models.Server) error
}

type serverService struct {
	logger *zap.Logger
	repo   repository.IServerRepository
	cache  db.IRedisClient
}

func NewServerService(repo repository.IServerRepository, cache db.IRedisClient, logger *zap.Logger) IServerService {
	return &serverService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}
func (s *serverService) CreateServer(ctx context.Context, server *models.Server) error {
	if exists, err := s.repo.ExistsByServerIDOrServerName(ctx, server.ServerID, server.ServerName); err != nil {
		s.logger.Error("Failed to check if server exists",
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to check if server exists: %w", err)
	} else if exists {
		s.logger.Error("Server already exists",
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
		)
		return fmt.Errorf("server with ID '%s' or name '%s' already exists", server.ServerID, server.ServerName)
	}
	err := s.repo.Create(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	s.logger.Info("Server created successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return nil
}

func (s *serverService) DeleteServer(ctx context.Context, id uint) error {
	server, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("server not found")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	s.ClearServerCaches(ctx, server)

	s.logger.Info("Server deleted successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return nil
}

func (s *serverService) ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in file")
	}

	allRows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	rowCount := len(allRows)
	if rowCount < 2 {
		return nil, fmt.Errorf("file must contain at least 2 rows (header + data)")
	}

	s.logger.Info("Starting import",
		zap.String("file", filePath),
		zap.Int("total_rows", rowCount),
	)

	result := &dto.ImportResult{
		SuccessServers: make([]string, 0),
		FailureServers: make([]string, 0),
	}

	if len(allRows) > 0 {
		err := utils.Validate(allRows[0])
		if err != nil {
			return nil, fmt.Errorf("header validation failed: %w", err)
		}
	}

	servers := make([]models.Server, 0)

	for i := 1; i < len(allRows); i++ {
		row := allRows[i]

		server, err := utils.ParseToServer(row)
		if err != nil {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers,
				fmt.Sprintf("Row %d: %s", i+1, err.Error()))
			continue
		}
		servers = append(servers, server)
	}

	if len(servers) == 0 {
		s.logger.Info("No valid servers found in file",
			zap.String("file", filePath),
			zap.Int("total_rows", rowCount),
			zap.Int("valid_rows", 0),
		)
		return result, nil
	}

	// Create batches
	var batches [][]models.Server
	batchSize := 150

	for i := 0; i < len(servers); i += batchSize {
		end := i + batchSize
		if end > len(servers) {
			end = len(servers)
		}
		batches = append(batches, servers[i:end])
	}

	// Process batches with proper concurrency control
	workerPool := workerpool.New(15) // Reduce workers to avoid overwhelming DB

	var mu sync.Mutex

	for batchIndex, batch := range batches {

		// Capture variables to avoid race condition
		currentBatch := batch
		currentIndex := batchIndex

		s.logger.Info("Processing batch",
			zap.Int("batch_index", currentIndex),
			zap.Int("batch_size", len(currentBatch)),
			zap.Int("total_batches", len(batches)),
		)

		workerPool.Submit(func() {

			// Try batch insert first
			if err := s.repo.BatchCreate(ctx, currentBatch); err != nil {
				// Fallback to individual inserts
				for _, sv := range currentBatch {
					if err := s.repo.Create(ctx, &sv); err != nil {
						mu.Lock()
						result.FailureCount++
						result.FailureServers = append(result.FailureServers,
							fmt.Sprintf("Server '%s': %s", sv.ServerID, err.Error()))
						mu.Unlock()
					} else {
						mu.Lock()
						result.SuccessCount++
						result.SuccessServers = append(result.SuccessServers, sv.ServerID)
						mu.Unlock()
					}
				}
			} else {
				mu.Lock()
				// Batch insert successful
				result.SuccessCount += len(currentBatch)
				for _, sv := range currentBatch {
					result.SuccessServers = append(result.SuccessServers, sv.ServerID)
				}
				mu.Unlock()
			}
		})
	}

	// Wait for all workers to complete
	workerPool.StopWait()

	s.logger.Info("Import completed",
		zap.String("file", filePath),
		zap.Int("total_rows", rowCount),
		zap.Int("success_count", result.SuccessCount),
		zap.Int("failure_count", result.FailureCount),
	)

	return result, nil
}

// ExportServers implements ServerService.
func (s *serverService) ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error) {
	servers, _, err := s.repo.List(ctx, filter, pagination)
	if err != nil {
		return "", fmt.Errorf("failed to get servers: %w", err)
	}

	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		return "", err
	}

	streamWriter.SetRow("A1", []interface{}{
		"Server ID", "Server Name", "Status", "Description", "IPv4", "Disk", "RAM", "Location", "OS",
	})

	for rowIndex, server := range servers {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+2)
		err = streamWriter.SetRow(cell, []interface{}{
			server.ServerID,
			server.ServerName,
			server.Status,
			server.Description,
			server.IPv4,
			server.Disk,
			server.RAM,
			server.Location,
			server.OS,
		})
		if err != nil {
			return "", err
		}
	}
	if err := streamWriter.Flush(); err != nil {
		return "", err
	}
	filePath := fmt.Sprintf("./exports/servers_%d.xlsx", time.Now().Unix())
	if err := file.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	s.logger.Info("Servers exported successfully",
		zap.String("file_path", filePath),
		zap.Int("total_servers", len(servers)),
	)

	return filePath, nil
}

func (s *serverService) GetServer(ctx context.Context, id uint) (*models.Server, error) {
	cacheKey := fmt.Sprintf("server:%d", id)
	var cachedServer *models.Server
	if err := s.cache.Get(ctx, cacheKey, &cachedServer); err == nil {
		s.logger.Info("Server retrieved from cache",
			zap.Uint("id", cachedServer.ID),
			zap.String("server_id", cachedServer.ServerID),
			zap.String("server_name", cachedServer.ServerName),
		)
		return cachedServer, nil
	} else if err != db.ErrCacheMiss {
		s.logger.Warn("Cache miss for server",
			zap.Uint("id", id),
			zap.Error(err),
		)
	}

	// Cache miss, get from database
	server, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	if err := s.cache.Set(ctx, cacheKey, server, 30*time.Minute); err != nil {
		s.logger.Error("Failed to cache server data",
			zap.Uint("id", server.ID),
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
			zap.Error(err),
		)
	}

	return server, nil
}

func (s *serverService) GetAllServers(ctx context.Context) ([]models.Server, error) {
	cacheKey := "server:all"
	var cachedServers []models.Server
	if err := s.cache.Get(ctx, cacheKey, &cachedServers); err == nil {
		s.logger.Info("All servers retrieved from cache",
			zap.Int("count", len(cachedServers)),
		)
		return cachedServers, nil
	} else if err != db.ErrCacheMiss {
		s.logger.Warn("Cache miss for all servers",
			zap.Error(err),
		)
	}

	servers, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all servers: %w", err)
	}

	if err := s.cache.Set(ctx, cacheKey, servers, 30*time.Minute); err != nil {
		s.logger.Error("Failed to cache all servers data",
			zap.Int("count", len(servers)),
			zap.Error(err),
		)
	}

	return servers, nil
}

func (s *serverService) GetServerStats(ctx context.Context) (map[string]int64, error) {
	// Try to get stats from cache first
	cacheKey := "server:stats"
	var cachedStats map[string]int64
	if err := s.cache.Get(ctx, cacheKey, &cachedStats); err == nil {
		s.logger.Info("Server stats retrieved from cache",
			zap.Any("stats", cachedStats),
		)
		return cachedStats, nil
	} else if err != db.ErrCacheMiss {
		s.logger.Warn("Cache miss for server stats",
			zap.Error(err),
		)
	}

	// Cache miss, calculate stats
	stats := make(map[string]int64)

	total, err := s.repo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count servers: %w", err)
	}
	stats["total"] = total

	// Count by status
	onlineCount, _ := s.repo.CountByStatus(ctx, models.ServerStatusOn)
	stats["online"] = onlineCount

	offlineCount, _ := s.repo.CountByStatus(ctx, models.ServerStatusOff)
	stats["offline"] = offlineCount

	if err := s.cache.Set(ctx, cacheKey, stats, 30*time.Minute); err != nil {
		s.logger.Error("Failed to cache server stats",
			zap.Any("stats", stats),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to cache server stats: %w", err)
	}

	return stats, nil
}

func (s *serverService) ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (*dto.ServerListResponse, error) {
	servers, total, err := s.repo.List(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	response := &dto.ServerListResponse{
		Total:   total,
		Servers: servers,
		Page:    pagination.Page,
		Size:    pagination.PageSize,
	}

	return response, nil
}

func (s *serverService) UpdateServer(ctx context.Context, id uint, updates dto.ServerUpdate) (*models.Server, error) {
	server, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("server not found")
	}

	if existing, err := s.repo.GetByServerName(ctx, updates.ServerName); err == nil && existing.ID != id {
		return nil, fmt.Errorf("server name already exists")
	}
	server.ServerName = updates.ServerName
	server.Status = updates.Status
	server.IPv4 = updates.IPv4
	server.Description = updates.Description
	server.Location = updates.Location
	server.OS = updates.OS
	server.CPU = updates.CPU
	server.RAM = updates.RAM
	server.Disk = updates.Disk

	if err := s.repo.Update(ctx, server); err != nil {
		return nil, err
	}

	s.logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	s.ClearServerCaches(ctx, server)

	return server, nil
}

func (s *serverService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	server, err := s.repo.GetByServerID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, serverID, status); err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	s.ClearServerCaches(ctx, server)

	s.logger.Info("Server status updated successfully",
		zap.String("server_id", serverID),
		zap.String("status", string(status)),
	)

	return nil
}

func (s *serverService) CheckServerStatus(ctx context.Context) error {
	s.logger.Info("Starting server health check")

	// Get all servers
	servers, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get servers", zap.Error(err))
		return err
	}

	s.logger.Info("Checking servers", zap.Int("count", len(servers)))

	workerpool := workerpool.New(50)
	mu := sync.Mutex{}
	var errs []error

	for _, srv := range servers {
		server := srv
		workerpool.Submit(func() {
			checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			err := s.CheckServer(checkCtx, server)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		})
	}
	workerpool.StopWait()
	if len(errs) > 0 {
		return fmt.Errorf("some servers failed health check: %v", errs)
	}
	return nil
}

func (s *serverService) CheckServer(ctx context.Context, server models.Server) error {
	status := models.ServerStatusOff

	// Try to ping the server
	if server.IPv4 != "" {
		timeout := 5 * time.Second // Default timeout since config might not have Monitoring
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", server.IPv4), timeout)
		if err == nil {
			conn.Close()
			status = models.ServerStatusOn
		}
	}

	if server.Status != status {
		err := s.repo.UpdateStatus(ctx, server.ServerID, status)
		if err != nil {
			s.logger.Error("Failed to update server status",
				zap.Error(err),
				zap.String("server_id", server.ServerID),
				zap.String("status", string(status)),
			)
			return fmt.Errorf("failed to update server status: %w", err)
		}
	}

	s.logger.Info("Server checked",
		zap.String("server_id", server.ServerID),
		zap.String("status", string(status)),
	)
	return nil
}

func (s *serverService) ClearServerCaches(ctx context.Context, server *models.Server) error {
	s.cache.Del(ctx, fmt.Sprintf("server:%d", server.ID))
	s.cache.Del(ctx, "server:stats")
	s.cache.Del(ctx, "server:all")
	return nil
}
