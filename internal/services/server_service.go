package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
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

	// Get first sheet
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in file")
	}

	rows, err := file.Rows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}
	defer rows.Close()

	result := &models.ImportResult{
		SuccessServers: make([]string, 0),
		FailureServers: make([]string, 0),
	}

	rowIndex := 0
	columnIndexes := make(map[string]int)

	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to read row %d: %w", rowIndex+1, err)
		}

		// Skip empty rows
		if len(row) == 0 {
			rowIndex++
			continue
		}

		// First row - parse headers
		if rowIndex == 0 {
			for i, col := range row {
				switch col {
				case "server_id":
					columnIndexes["server_id"] = i
				case "server_name":
					columnIndexes["server_name"] = i
				case "status":
					columnIndexes["status"] = i
				case "ipv4":
					columnIndexes["ipv4"] = i
				case "description":
					columnIndexes["description"] = i
				case "location":
					columnIndexes["location"] = i
				case "os":
					columnIndexes["os"] = i
				case "cpu":
					columnIndexes["cpu"] = i
				case "ram":
					columnIndexes["ram"] = i
				case "disk":
					columnIndexes["disk"] = i
				}
			}

			// Validate required columns exist
			if _, exists := columnIndexes["server_id"]; !exists {
				return nil, fmt.Errorf("server_id column is required in the Excel file")
			}
			if _, exists := columnIndexes["server_name"]; !exists {
				return nil, fmt.Errorf("server_name column is required in the Excel file")
			}

			rowIndex++
			continue
		}

		// Data rows - create server objects
		server := models.Server{}
		var serverIdentifier string

		// Extract required fields
		if idx, exists := columnIndexes["server_id"]; exists && idx < len(row) {
			server.ServerID = row[idx]
			serverIdentifier = server.ServerID
		}
		if idx, exists := columnIndexes["server_name"]; exists && idx < len(row) {
			server.ServerName = row[idx]
			if serverIdentifier == "" {
				serverIdentifier = server.ServerName
			}
		}

		// Validate required fields
		if server.ServerID == "" || server.ServerName == "" {
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("Row %d: server_id and server_name are required", rowIndex+1))
			result.FailureCount++
			rowIndex++
			continue
		}
		if _, err := s.serverRepo.GetByServerID(ctx, server.ServerID); err == nil {
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("Row %d: server_id is already exists", rowIndex+1))
			result.FailureCount++
			rowIndex++
			continue
		}

		if _, err := s.serverRepo.GetByServerName(ctx, server.ServerName); err == nil {
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("Row %d: server_name is already exists", rowIndex+1))
			result.FailureCount++
			rowIndex++
			continue
		}

		// Extract optional fields
		if idx, exists := columnIndexes["status"]; exists && idx < len(row) && row[idx] != "" {
			server.Status = models.ServerStatus(row[idx])
		} else {
			server.Status = "OFF" // default status
		}

		if idx, exists := columnIndexes["ipv4"]; exists && idx < len(row) {
			server.IPv4 = row[idx]
		}

		if idx, exists := columnIndexes["description"]; exists && idx < len(row) {
			server.Description = row[idx]
		}

		if idx, exists := columnIndexes["location"]; exists && idx < len(row) {
			server.Location = row[idx]
		}

		if idx, exists := columnIndexes["os"]; exists && idx < len(row) {
			server.OS = row[idx]
		}

		// Parse integer fields with error handling
		if idx, exists := columnIndexes["cpu"]; exists && idx < len(row) && row[idx] != "" {
			cpu, parseErr := strconv.Atoi(row[idx]); parseErr == nil {
				server.CPU = cpu
			}
		}

		if idx, exists := columnIndexes["ram"]; exists && idx < len(row) && row[idx] != "" {
			if ram, parseErr := strconv.Atoi(row[idx]); parseErr == nil {
				server.RAM = ram
			}
		}

		if idx, exists := columnIndexes["disk"]; exists && idx < len(row) && row[idx] != "" {
			if disk, parseErr := strconv.Atoi(row[idx]); parseErr == nil {
				server.Disk = disk
			}
		}

		// Try to create the server
		err = s.CreateServer(ctx, &server)
		if err != nil {
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s: %s", serverIdentifier, err.Error()))
			result.FailureCount++
		} else {
			result.SuccessServers = append(result.SuccessServers, serverIdentifier)
			result.SuccessCount++
		}

		rowIndex++
	}

	logger.Info("Import completed",
		zap.String("file", filePath),
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
	logger.Info("Servers exported",
		zap.String("file", filePath),
		zap.Int("count", len(servers)),
	)

	return filePath, nil
}

// GetServer implements ServerService.
func (s *serverService) GetServer(ctx context.Context, id uint) (*models.Server, error) {
	panic("unimplemented")
}

// GetServerByServerID implements ServerService.
func (s *serverService) GetServerByServerID(ctx context.Context, serverID string) (*models.Server, error) {
	panic("unimplemented")
}

// GetServerStats implements ServerService.
func (s *serverService) GetServerStats(ctx context.Context) (map[string]int64, error) {
	panic("unimplemented")
}

// ListServers implements ServerService.
func (s *serverService) ListServers(ctx context.Context, filter models.ServerFilter, pagination models.Pagination) (*models.ServerListResponse, error) {
	servers, total, err := s.serverRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	return &models.ServerListResponse{
		Total:   total,
		Servers: servers,
		Page:    pagination.Page,
		Size:    pagination.PageSize,
	}, nil
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
	logger.Info("Server updated successfully",
		zap.Uint("id", server.ID),
		zap.String("server_id", server.ServerID),
	)
	return server, nil
}

// UpdateServerStatus implements ServerService.
func (s *serverService) UpdateServerStatus(ctx context.Context, serverID string, status models.ServerStatus) error {
	panic("unimplemented")
}
