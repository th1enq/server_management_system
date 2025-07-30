package usecases

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/query"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/domain/services"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/infrastructure/cache"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

const (
	bufferTime = 5 * time.Second
)

type ServerUseCase interface {
	RefreshStatus(ctx context.Context) error
	UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus) error
	Register(ctx context.Context, req dto.CreateServerRequest) (*dto.AuthResponse, error)
	CreateServer(ctx context.Context, req dto.CreateServerRequest) (*entity.Server, error)
	GetServerByID(ctx context.Context, serverID string) (*entity.Server, error)
	GetServer(ctx context.Context, id uint) (*entity.Server, error)
	ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]*entity.Server, error)
	UpdateServer(ctx context.Context, id uint, updates dto.UpdateServerRequest) (*entity.Server, error)
	DeleteServer(ctx context.Context, id uint) error
	ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error)
	ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error)
	GetServerStats(ctx context.Context) (dto.ServerStatusResponse, error)
	GetServerIDs(ctx context.Context) ([]string, error)
}

type serverUseCase struct {
	logger               *zap.Logger
	serverRepo           repository.ServerRepository
	tokenServices        services.TokenServices
	excelizeServices     services.ExcelizeService
	statusChangeProducer producer.StatusChangeMessageProducer
	inMemoryCache        cache.InMemoryCache
	redisCache           cache.CacheClient
	healthCheckUseCase   HealthCheckUseCase
}

func NewServerUseCase(serverRepo repository.ServerRepository, tokenServices services.TokenServices, excelizeServices services.ExcelizeService, statusChangeProducer producer.StatusChangeMessageProducer, inMemoryCache cache.InMemoryCache, redisCache cache.CacheClient, healthCheckUseCase HealthCheckUseCase, logger *zap.Logger) ServerUseCase {
	return &serverUseCase{
		serverRepo:           serverRepo,
		tokenServices:        tokenServices,
		excelizeServices:     excelizeServices,
		statusChangeProducer: statusChangeProducer,
		inMemoryCache:        inMemoryCache,
		redisCache:           redisCache,
		logger:               logger,
		healthCheckUseCase:   healthCheckUseCase,
	}
}

func (s *serverUseCase) GetServerByID(ctx context.Context, serverID string) (*entity.Server, error) {
	server, err := s.serverRepo.GetByServerID(ctx, serverID)
	if err != nil {
		s.logger.Error("Failed to get server by ID",
			zap.String("server_id", serverID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get server by ID: %w", err)
	}
	return server, nil
}

func (s *serverUseCase) UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus) error {
	if err := s.serverRepo.UpdateStatus(ctx, serverID, status); err != nil {
		s.logger.Error("Failed to update server status",
			zap.String("server_id", serverID),
			zap.String("status", string(status)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update server status: %w", err)
	}
	s.logger.Info("Server status updated successfully",
		zap.String("server_id", serverID),
		zap.String("status", string(status)),
	)
	return nil
}

func (s *serverUseCase) RefreshStatus(ctx context.Context) error {
	serverIDs, err := s.GetServerIDs(ctx)
	if err != nil {
		s.logger.Error("Failed to get server IDs for refresh",
			zap.Error(err),
		)
		return fmt.Errorf("failed to get server IDs: %w", err)
	}
	if len(serverIDs) == 0 {
		s.logger.Info("No servers found for status refresh")
		return nil
	}

	for _, serverID := range serverIDs {
		var intervalCheckTime int64
		cacheKey := fmt.Sprintf("heartbeat:%s", serverID)
		if err := s.redisCache.Get(ctx, cacheKey, &intervalCheckTime); err != nil {
			s.logger.Info("Failed to get heartbeat timestamp from cache",
				zap.String("cache_key", cacheKey),
			)

			server, err := s.GetServerByID(ctx, serverID)
			if err != nil {
				s.logger.Error("Failed to get server by ID",
					zap.String("server_id", serverID),
					zap.Error(err),
				)
				return fmt.Errorf("failed to get server by ID: %w", err)
			}

			oldStatus := server.Status

			if err := s.serverRepo.UpdateStatus(ctx, serverID, entity.ServerStatusOff); err != nil {
				s.logger.Error("Failed to update server status to OFF",
					zap.String("server_id", serverID),
					zap.Error(err),
				)
				return fmt.Errorf("failed to update server status to OFF: %w", err)
			}

			intervalCheckTime := server.IntervalTime

			s.statusChangeProducer.Produce(ctx, producer.StatusChangeMessage{
				ServerID:  serverID,
				OldStatus: oldStatus,
				NewStatus: entity.ServerStatusOff,
				Timestamp: time.Now(),
				Interval:  intervalCheckTime,
			})
		}
	}
	return nil
}

func (s *serverUseCase) Register(ctx context.Context, req dto.CreateServerRequest) (*dto.AuthResponse, error) {
	server, err := s.CreateServer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to register server: %w", err)
	}

	accessToken, err := s.tokenServices.GenerateServerAccessToken(ctx, server)
	if err != nil {
		s.logger.Error("Failed to generate access token for server",
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenServices.GenerateServerRefreshToken(ctx, server)
	if err != nil {
		s.logger.Error("Failed to generate refresh token for server",
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *serverUseCase) CreateServer(ctx context.Context, req dto.CreateServerRequest) (*entity.Server, error) {
	if exists, err := s.serverRepo.ExistsByServerIDOrServerName(ctx, req.ServerID, req.ServerName); err != nil {
		s.logger.Error("Failed to check if server exists",
			zap.String("server_id", req.ServerID),
			zap.String("server_name", req.ServerName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check if server exists: %w", err)
	} else if exists {
		s.logger.Warn("Server already exists",
			zap.String("server_id", req.ServerID),
			zap.String("server_name", req.ServerName),
		)
		return nil, fmt.Errorf("server with ID '%s' or name '%s' already exists", req.ServerID, req.ServerName)
	}

	server := &entity.Server{
		ServerID:     req.ServerID,
		ServerName:   req.ServerName,
		Status:       entity.ServerStatusOff,
		IPv4:         req.IPv4,
		IntervalTime: req.IntervalTime,
	}
	if req.Description != "" {
		server.Description = req.Description
	}
	if req.Location != "" {
		server.Location = req.Location
	}
	if req.OS != "" {
		server.OS = req.OS
	}

	err := s.serverRepo.Create(ctx, server)
	if err != nil {
		s.logger.Error("Failed to create server",
			zap.String("server_id", req.ServerID),
			zap.String("server_name", req.ServerName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	s.inMemoryCache.Delete("list_server_id")

	s.healthCheckUseCase.InsertUptime(ctx, producer.StatusChangeMessage{
		ServerID:  server.ServerID,
		OldStatus: entity.ServerStatusOn,
		NewStatus: entity.ServerStatusOff,
		Timestamp: time.Now(),
		Interval:  server.IntervalTime,
	})

	s.logger.Info("Server created successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return server, nil
}

func (s *serverUseCase) DeleteServer(ctx context.Context, id uint) error {
	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get server by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return fmt.Errorf("server not found")
	}
	if err := s.serverRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete server",
			zap.Uint("id", server.ID),
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete server: %w", err)
	}

	s.inMemoryCache.Delete("list_server_id")

	s.logger.Info("Server deleted successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return nil
}

func (s *serverUseCase) ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		s.logger.Error("Failed to open import file",
			zap.String("file_path", filePath),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		s.logger.Error("No sheets found in import file",
			zap.String("file_path", filePath),
		)
		return nil, fmt.Errorf("no sheets found in file")
	}

	allRows, err := file.GetRows(sheets[0])
	if err != nil {
		s.logger.Error("Failed to get rows from import file",
			zap.String("file_path", filePath),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	rowCount := len(allRows)
	if rowCount < 2 {
		s.logger.Error("Import file must contain at least 2 rows (header + data)",
			zap.String("file_path", filePath),
		)
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
		err := s.excelizeServices.Validate(allRows[0])
		if err != nil {
			s.logger.Error("Header validation failed",
				zap.String("file_path", filePath),
				zap.Error(err),
			)
			return nil, fmt.Errorf("header validation failed: %w", err)
		}
	}

	servers := make([]entity.Server, 0)

	for i := 1; i < len(allRows); i++ {
		row := allRows[i]

		server, err := s.excelizeServices.ParseToServer(row)
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
	var batches [][]entity.Server
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

	successID := make(map[string]bool)

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
			successServer, err := s.serverRepo.BatchCreate(ctx, currentBatch)
			if err == nil {
				mu.Lock()
				result.SuccessCount += len(successServer)
				for _, server := range successServer {
					result.SuccessServers = append(result.SuccessServers, server.ServerID)
					successID[server.ServerID] = true
				}
				mu.Unlock()
			}
		})
	}

	workerPool.StopWait()

	result.FailureCount = rowCount - result.SuccessCount - 1
	for _, server := range servers {
		if _, exists := successID[server.ServerID]; !exists {
			result.FailureServers = append(result.FailureServers, server.ServerID)
		}
	}

	s.inMemoryCache.Delete("list_server_id")

	s.logger.Info("Import completed",
		zap.String("file", filePath),
		zap.Int("total_rows", rowCount),
		zap.Int("success_count", result.SuccessCount),
		zap.Int("failure_count", result.FailureCount),
	)

	return result, nil
}

func (s *serverUseCase) ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error) {
	servers, err := s.ListServers(ctx, filter, pagination)
	if err != nil {
		s.logger.Error("Failed to list servers for export",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
		)
		return "", fmt.Errorf("failed to get servers: %w", err)
	}

	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		s.logger.Error("Failed to create stream writer for export",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
		)
		return "", err
	}

	streamWriter.SetRow("A1", []interface{}{
		"Server ID", "Server Name", "Status", "Description", "IPv4", "Location", "OS",
	})

	for rowIndex, server := range servers {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+2)
		err = streamWriter.SetRow(cell, []interface{}{
			server.ServerID,
			server.ServerName,
			server.Status,
			server.Description,
			server.IPv4,
			server.Location,
			server.OS,
		})
		if err != nil {
			s.logger.Error("Failed to write server data to export file",
				zap.Error(err),
				zap.String("server_id", server.ServerID),
				zap.String("server_name", server.ServerName),
			)
			return "", err
		}
	}
	if err := streamWriter.Flush(); err != nil {
		s.logger.Error("Failed to flush stream writer for export",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
		)
		return "", err
	}
	filePath := fmt.Sprintf("./exports/servers_%d.xlsx", time.Now().Unix())
	if err := file.SaveAs(filePath); err != nil {
		s.logger.Error("Failed to save export file",
			zap.Error(err),
			zap.String("file_path", filePath),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
		)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	s.logger.Info("Servers exported successfully",
		zap.String("file_path", filePath),
		zap.Int("total_servers", len(servers)),
	)

	return filePath, nil
}

func (s *serverUseCase) GetServer(ctx context.Context, id uint) (*entity.Server, error) {
	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	s.logger.Info("Server retrieved successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return server, nil
}

func (s *serverUseCase) GetServerIDs(ctx context.Context) ([]string, error) {
	cacheKey := "list_server_id"
	servers, err := s.inMemoryCache.Get(cacheKey)
	serverIDs := make([]string, 0)
	if err != nil {
		s.logger.Error("Failed to get server IDs from cache",
			zap.String("cache_key", cacheKey),
			zap.Error(err),
		)
		serverIDs, err = s.serverRepo.GetServerIDs(ctx)
		if err != nil {
			s.logger.Error("Failed to get server IDs from repository",
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to get server IDs: %w", err)
		}
	} else {
		serverIDs, _ = servers.([]string)
	}
	if err := s.inMemoryCache.Set(cacheKey, serverIDs); err != nil {
		s.logger.Error("Failed to set server IDs in cache",
			zap.String("cache_key", cacheKey),
			zap.Error(err),
		)
	}
	s.logger.Info("Server IDs retrieved successfully",
		zap.Int("total_ids", len(serverIDs)),
	)
	return serverIDs, nil
}

func (s *serverUseCase) GetServerStats(ctx context.Context) (dto.ServerStatusResponse, error) {
	stats := dto.ServerStatusResponse{
		TotalCount:   0,
		OnlineCount:  0,
		OfflineCount: 0,
	}

	total, err := s.serverRepo.CountAll(ctx)
	if err != nil {
		s.logger.Error("Failed to count total servers",
			zap.Error(err),
		)
		return dto.ServerStatusResponse{}, fmt.Errorf("failed to count total servers: %w", err)
	}
	stats.TotalCount = total

	// Count by status
	onlineCount, _ := s.serverRepo.CountByStatus(ctx, entity.ServerStatusOn)
	stats.OnlineCount = onlineCount

	offlineCount, _ := s.serverRepo.CountByStatus(ctx, entity.ServerStatusOff)
	stats.OfflineCount = offlineCount
	return stats, nil
}

func (s *serverUseCase) ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]*entity.Server, error) {
	queryFilter := query.ServerFilter{
		ServerID:   filter.ServerID,
		ServerName: filter.ServerName,
		Status:     filter.Status,
		IPv4:       filter.IPv4,
		Location:   filter.Location,
		Disk:       filter.Disk,
	}

	queryPagination := query.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Sort:     pagination.Sort,
		Order:    pagination.Order,
	}

	if queryPagination.Page < 1 {
		queryPagination.Page = 1
	}
	if queryPagination.PageSize < 1 {
		queryPagination.PageSize = 10
	}
	if queryPagination.Sort == "" {
		queryPagination.Sort = "created_at"
	}
	if queryPagination.Order == "" {
		queryPagination.Order = "desc"
	}

	servers, total, err := s.serverRepo.List(ctx, queryFilter, queryPagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	s.logger.Info("Servers listed successfully",
		zap.Int64("total", total),
		zap.Int("page", pagination.Page),
		zap.Int("page_size", pagination.PageSize),
	)

	return servers, nil
}

func (s *serverUseCase) UpdateServer(ctx context.Context, id uint, updates dto.UpdateServerRequest) (*entity.Server, error) {
	server, err := s.serverRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get server by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("server not found")
	}
	if updates.ServerName != "" {
		if existing, err := s.serverRepo.GetByServerName(ctx, updates.ServerName); err == nil && existing.ID != id {
			s.logger.Warn("Server name already exists",
				zap.String("server_name", updates.ServerName),
				zap.Uint("existing_id", existing.ID),
			)
			return nil, fmt.Errorf("server name already exists")
		}
		server.ServerName = updates.ServerName
	}
	if updates.IPv4 != "" {
		server.IPv4 = updates.IPv4
	}
	if updates.Description != "" {
		server.Description = updates.Description
	}
	if updates.Location != "" {
		server.Location = updates.Location
	}
	if updates.OS != "" {
		server.OS = updates.OS
	}
	if updates.IntervalTime > 0 {
		server.IntervalTime = updates.IntervalTime
	}

	if err := s.serverRepo.Update(ctx, server); err != nil {
		return nil, err
	}

	s.logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return server, nil
}
