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
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ServerUseCase interface {
	Monitoring(ctx context.Context, serverID string) error
	Register(ctx context.Context, req dto.CreateServerRequest) (*dto.AuthResponse, error)
	CreateServer(ctx context.Context, req dto.CreateServerRequest) (*entity.Server, error)
	GetServer(ctx context.Context, id uint) (*entity.Server, error)
	ListServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) ([]*entity.Server, error)
	UpdateServer(ctx context.Context, id uint, updates dto.UpdateServerRequest) (*entity.Server, error)
	DeleteServer(ctx context.Context, id uint) error
	ImportServers(ctx context.Context, filePath string) (*dto.ImportResult, error)
	ExportServers(ctx context.Context, filter dto.ServerFilter, pagination dto.Pagination) (string, error)
	UpdateServerStatus(ctx context.Context, serverID string, status entity.ServerStatus) error
	GetServerStats(ctx context.Context) (dto.ServerStatusResponse, error)
	GetAllServers(ctx context.Context) ([]*entity.Server, error)
}

type serverUseCase struct {
	logger           *zap.Logger
	serverRepo       repository.ServerRepository
	tokenServices    services.TokenServices
	tokenRepository  repository.TokenRepository
	excelizeServices services.ExcelizeService
	cache            cache.CacheClient
}

func NewServerUseCase(serverRepo repository.ServerRepository, tokenServices services.TokenServices, tokenRepository repository.TokenRepository, excelizeServices services.ExcelizeService, cache cache.CacheClient, logger *zap.Logger) ServerUseCase {
	return &serverUseCase{
		serverRepo:       serverRepo,
		tokenServices:    tokenServices,
		tokenRepository:  tokenRepository,
		excelizeServices: excelizeServices,
		cache:            cache,
		logger:           logger,
	}
}

func (serverUseCase *serverUseCase) Monitoring(ctx context.Context, serverID string) error {
	cacheKey := fmt.Sprintf("heartbeat:%s", serverID)
	var intervalTime int

	if err := serverUseCase.cache.Get(ctx, cacheKey, &intervalTime); err == nil {
		expireTime := time.Duration(intervalTime+5) * time.Second
		if err := serverUseCase.cache.Expire(ctx, cacheKey, expireTime); err != nil {
			serverUseCase.logger.Error("Failed to reset heartbeat expiration",
				zap.String("server_id", serverID),
				zap.Int("interval_time", intervalTime),
			)
		}
		return nil
	}

	server, err := serverUseCase.serverRepo.UpdateStatus(ctx, serverID, entity.ServerStatusOn)
	if err != nil {
		serverUseCase.logger.Error("Failed to update server status to ON",
			zap.String("server_id", serverID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update server status: %w", err)
	}

	expireTime := time.Duration(server.IntervalTime+5) * time.Second
	if err := serverUseCase.cache.Set(ctx, cacheKey, server.IntervalTime, expireTime); err != nil {
		serverUseCase.logger.Error("Failed to set heartbeat cache",
			zap.String("server_id", serverID),
			zap.Int("interval_time", server.IntervalTime),
			zap.Error(err),
		)
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

	// Add tokens to Redis whitelist
	if err := s.tokenRepository.AddTokenToWhitelist(ctx, accessToken, time.Hour*24); err != nil {
		s.logger.Error("Failed to add access token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
	}

	if err := s.tokenRepository.AddTokenToWhitelist(ctx, refreshToken, time.Hour*24*7); err != nil {
		s.logger.Error("Failed to add refresh token to whitelist", zap.Error(err))
		return nil, fmt.Errorf("failed to whitelist token")
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
		ServerID:   req.ServerID,
		ServerName: req.ServerName,
		Status:     entity.ServerStatusOff,
		IPv4:       req.IPv4,
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
	if req.IntervalTime > 0 {
		server.IntervalTime = req.IntervalTime
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

func (s *serverUseCase) GetAllServers(ctx context.Context) ([]*entity.Server, error) {

	servers, err := s.serverRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all servers: %w", err)
	}
	s.logger.Info("All servers retrieved successfully",
		zap.Int("count", len(servers)),
	)
	return servers, nil
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
		OS:         filter.OS,
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

	cacheKey := fmt.Sprintf("heartbeat:%s", server.ServerID)
	expireTime := time.Duration(server.IntervalTime+5) * time.Second
	if err := s.cache.Set(ctx, cacheKey, server.IntervalTime, expireTime); err != nil {
		s.logger.Error("Failed to cache server heartbeat",
			zap.String("server_id", server.ServerID),
			zap.Error(err),
		)
	}

	s.logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
	)

	return server, nil
}

func (s *serverUseCase) UpdateServerStatus(ctx context.Context, serverID string, status entity.ServerStatus) error {
	_, err := s.serverRepo.GetByServerID(ctx, serverID)
	if err != nil {
		s.logger.Error("Failed to get server by ID",
			zap.String("server_id", serverID),
			zap.Error(err),
		)
		return fmt.Errorf("server not found: %w", err)
	}

	if _, err := s.serverRepo.UpdateStatus(ctx, serverID, status); err != nil {
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
