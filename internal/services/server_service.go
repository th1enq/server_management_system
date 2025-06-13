package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/internal/utils"
	"github.com/th1enq/server_management_system/pkg/logger"
	"github.com/xuri/excelize/v2"
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
	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("server not found")
	}
	if err := s.serverRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	logger.Info("Delete server successfully",
		zap.Uint("id", id),
		zap.String("server_id", server.ServerID))

	return nil
}

// ImportServers implements ServerService.
func (s *serverService) ImportServers(ctx context.Context, filePath string) (*models.ImportResult, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in file")
	}

	// Get all rows to count them first
	allRows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	rowCount := len(allRows)
	if rowCount < 2 {
		return nil, fmt.Errorf("file must contain at least 2 rows (header + data)")
	}

	logger.Info("Starting import process",
		zap.String("file", filePath),
		zap.Int("total_rows", rowCount-1), // excluding header
	)

	result := &models.ImportResult{
		SuccessServers: make([]string, 0),
		FailureServers: make([]string, 0),
	}

	// Validate header (first row)
	if len(allRows) > 0 {
		err := utils.Validate(allRows[0])
		if err != nil {
			return nil, fmt.Errorf("header validation failed: %w", err)
		}
	}

	servers := make([]models.Server, 0)

	benchmarks := time.Now()

	// Process data rows (skip header)
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
		logger.Warn("No valid servers to import")
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

		logger.Info("Processing batch",
			zap.Int("batch", currentIndex+1),
			zap.Int("size", len(currentBatch)),
		)

		workerPool.Submit(func() {

			// Try batch insert first
			if err := s.serverRepo.BatchCreate(ctx, currentBatch); err != nil {
				// Fallback to individual inserts
				for _, sv := range currentBatch {
					if err := s.serverRepo.Create(ctx, &sv); err != nil {
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

	logger.Info("Import completed",
		zap.Int("success_count", result.SuccessCount),
		zap.Int("failure_count", result.FailureCount),
		zap.Int("total_processed", result.SuccessCount+result.FailureCount),
		zap.Duration("duration", time.Since(benchmarks)),
	)

	return result, nil
}

// ExportServers implements ServerService.
func (s *serverService) ExportServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (string, error) {
	servers, _, err := s.serverRepo.List(ctx, filter, pagination)
	if err != nil {
		return "", fmt.Errorf("failed to get servers: %w", err)
	}

	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		return "", err
	}

	for rowIndex, server := range servers {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+1)
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
	logger.Info("Servers exported",
		zap.String("file", filePath),
		zap.Int("count", len(servers)),
	)

	return filePath, nil
}

// GetServer implements ServerService.
func (s *serverService) GetServer(ctx context.Context, id uint) (*models.Server, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("server:%d", id)

	// Try to get server from Redis
	serverJSON, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var server models.Server
		if err := json.Unmarshal([]byte(serverJSON), &server); err == nil {
			return &server, nil
		}
	}

	// Cache miss, get from database
	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Store in cache for future requests
	if serverJSON, err := json.Marshal(server); err == nil {
		s.redisClient.Set(ctx, cacheKey, serverJSON, 30*time.Minute)
	}

	return server, nil
}

// GetServerByServerID implements ServerService.
func (s *serverService) GetServerByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	server, err := s.serverRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}
	return server, nil
}

// GetServerStats implements ServerService.
func (s *serverService) GetServerStats(ctx context.Context) (map[string]int64, error) {
	// Try to get stats from cache first
	cacheKey := "server:stats"
	statsJSON, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var stats map[string]int64
		if err := json.Unmarshal([]byte(statsJSON), &stats); err == nil {
			return stats, nil
		}
	}

	// Cache miss, calculate stats
	stats := make(map[string]int64)

	// Count total servers
	var total int64
	if servers, err := s.serverRepo.GetAll(ctx); err == nil {
		total = int64(len(servers))
	}
	stats["total"] = total

	// Count by status
	onlineCount, _ := s.serverRepo.CountByStatus(ctx, "ON")
	stats["online"] = onlineCount

	offlineCount, _ := s.serverRepo.CountByStatus(ctx, "OFF")
	stats["offline"] = offlineCount

	maintenanceCount, _ := s.serverRepo.CountByStatus(ctx, "MAINTENANCE")
	stats["maintenance"] = maintenanceCount

	// Store in cache for future requests
	if statsJSON, err := json.Marshal(stats); err == nil {
		s.redisClient.Set(ctx, cacheKey, statsJSON, 5*time.Minute)
	}

	return stats, nil
}

// ListServers implements ServerService.
func (s *serverService) ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error) {
	// Generate cache key based on filter and pagination
	cacheKey := fmt.Sprintf("servers:list:%v:%d:%d", filter, pagination.Page, pagination.PageSize)

	// Try to get from cache
	responseJSON, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var response models.ServerListResponse
		if err := json.Unmarshal([]byte(responseJSON), &response); err == nil {
			return &response, nil
		}
	}

	// Cache miss, get from database
	servers, total, err := s.serverRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	response := &models.ServerListResponse{
		Total:   total,
		Servers: servers,
		Page:    pagination.Page,
		Size:    pagination.PageSize,
	}

	// Store in cache for future requests
	if responseJSON, err := json.Marshal(response); err == nil {
		s.redisClient.Set(ctx, cacheKey, responseJSON, 5*time.Minute)
	}

	return response, nil
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

	// Invalidate caches
	s.invalidateServerCaches(ctx, server)

	logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
	)
	return server, nil
}

// UpdateServerStatus implements ServerService.
func (s *serverService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	// Check if server exists
	server, err := s.serverRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Update the status
	if err := s.serverRepo.UpdateStatus(ctx, serverID, status); err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	// Invalidate caches
	s.invalidateServerCaches(ctx, server)

	logger.Info("Server status updated successfully",
		zap.String("server_id", serverID),
		zap.String("old_status", string(server.Status)),
		zap.String("new_status", string(status)),
	)

	return nil
}

// Helper method to invalidate all related caches
func (s *serverService) invalidateServerCaches(ctx context.Context, server *models.Server) {
	// Delete server cache
	s.redisClient.Del(ctx, fmt.Sprintf("server:%d", server.ID))

	// Delete server by server_id cache
	s.redisClient.Del(ctx, fmt.Sprintf("server:byServerID:%s", server.ServerID))

	// Delete stats cache
	s.redisClient.Del(ctx, "server:stats")

	// Delete list caches - using pattern matching
	pattern := "servers:list:*"
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}
}
