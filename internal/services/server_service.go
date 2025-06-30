package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/internal/utils"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ServerService interface {
	CreateServer(ctx context.Context, server *models.Server) error
	GetServer(ctx context.Context, id uint) (*models.Server, error)
	ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error)
	UpdateServer(ctx context.Context, id uint, updates map[string]interface{}) (*models.Server, error)
	DeleteServer(ctx context.Context, id uint) error
	ImportServers(ctx context.Context, filePath string) (*models.ImportResult, error)
	ExportServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (string, error)
	UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error
	GetServerStats(ctx context.Context) (map[string]int64, error)
	GetAllServers(ctx context.Context) ([]models.Server, error)
	CheckServerStatus(ctx context.Context) error
	CheckServer(ctx context.Context, server models.Server)
}

type serverService struct {
	logger      *zap.Logger
	serverRepo  repositories.ServerRepository
	redisClient *redis.Client
}

func NewServerService(serverRepo repositories.ServerRepository, redisClient *redis.Client, logger *zap.Logger) ServerService {
	return &serverService{
		serverRepo:  serverRepo,
		redisClient: redisClient,
		logger:      logger,
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

	s.logger.Info("Server created successfully",
		zap.Uint("id", server.ID),
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

	s.logger.Info("Server deleted successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

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

	s.logger.Info("Starting import",
		zap.String("file", filePath),
		zap.Int("total_rows", rowCount),
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

	s.logger.Info("Import completed",
		zap.String("file", filePath),
		zap.Int("total_rows", rowCount),
		zap.Int("success_count", result.SuccessCount),
		zap.Int("failure_count", result.FailureCount),
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

	s.logger.Info("Servers exported successfully",
		zap.String("file_path", filePath),
		zap.Int("total_servers", len(servers)),
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
	onlineCount, _ := s.serverRepo.CountByStatus(ctx, models.ServerStatusOn)
	stats["online"] = onlineCount

	offlineCount, _ := s.serverRepo.CountByStatus(ctx, models.ServerStatusOff)
	stats["offline"] = offlineCount

	// Store in cache for future requests
	if statsJSON, err := json.Marshal(stats); err == nil {
		s.redisClient.Set(ctx, cacheKey, statsJSON, 5*time.Minute)
	}

	return stats, nil
}

// ListServers implements ServerService.
func (s *serverService) ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error) {
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

	s.logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
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

	s.logger.Info("Server status updated successfully",
		zap.String("server_id", serverID),
		zap.String("status", string(status)),
	)

	return nil
}

func (s *serverService) GetAllServers(ctx context.Context) ([]models.Server, error) {
	// Try to get from cache first
	cacheKey := "servers:all"

	// Try to get servers from Redis
	serversJSON, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var servers []models.Server
		if err := json.Unmarshal([]byte(serversJSON), &servers); err == nil {
			return servers, nil
		}
	}

	// Cache miss, get from database
	servers, err := s.serverRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all servers: %w", err)
	}

	// Store in cache for future requests
	if serversJSON, err := json.Marshal(servers); err == nil {
		s.redisClient.Set(ctx, cacheKey, serversJSON, 30*time.Minute)
	}

	return servers, nil
}

// Helper method to invalidate all related caches
func (s *serverService) invalidateServerCaches(ctx context.Context, server *models.Server) {
	// Delete server cache
	s.redisClient.Del(ctx, fmt.Sprintf("server:%d", server.ID))

	// Delete server by server_id cache
	s.redisClient.Del(ctx, fmt.Sprintf("server:byServerID:%s", server.ServerID))

	// Delete stats cache
	s.redisClient.Del(ctx, "server:stats")
}

func (s *serverService) CheckServerStatus(ctx context.Context) error {
	s.logger.Info("Starting server health check")

	// Get all servers
	servers, err := s.serverRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get servers", zap.Error(err))
		return err
	}

	s.logger.Info("Checking servers", zap.Int("count", len(servers)))

	for _, server := range servers {
		go s.CheckServer(ctx, server)
	}
	return nil
}

func (s *serverService) CheckServer(ctx context.Context, server models.Server) {
	startTime := time.Now()
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
		err := s.serverRepo.UpdateStatus(ctx, server.ServerID, status)
		if err != nil {
			s.logger.Error("Failed to update server status",
				zap.Error(err),
				zap.String("server_id", server.ServerID),
				zap.String("status", string(status)),
			)
		}
	}

	responseTime := time.Since(startTime).Milliseconds()
	s.logger.Info("Server checked",
		zap.String("server_id", server.ServerID),
		zap.String("status", string(status)),
		zap.Int64("response_time", responseTime),
	)
}
