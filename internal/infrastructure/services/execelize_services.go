package services

import (
	"fmt"
	"strings"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/services"
)

type excelizeService struct{}

func NewExcelizeService() services.ExcelizeService {
	return &excelizeService{}
}

func (e *excelizeService) Validate(row []string) error {
	expectedHeaders := []string{
		"server_id",
		"server_name",
		"status",
		"ipv4",
		"description",
		"location",
		"os",
	}

	if len(row) < len(expectedHeaders) {
		return fmt.Errorf("invalid header: expected at least %d columns, got %d", len(expectedHeaders), len(row))
	}

	for i, expected := range expectedHeaders {
		if i >= len(row) || strings.TrimSpace(strings.ToLower(row[i])) != expected {
			return fmt.Errorf("invalid header at column %d: expected '%s', got '%s'", i+1, expected, row[i])
		}
	}

	return nil
}

// ParseToServer parses a row to Server model
func (e *excelizeService) ParseToServer(row []string) (entity.Server, error) {
	if len(row) < 7 {
		return entity.Server{}, fmt.Errorf("invalid row: expected at least 7 columns, got %d", len(row))
	}

	server := entity.Server{
		ServerID:    strings.TrimSpace(row[0]),
		ServerName:  strings.TrimSpace(row[1]),
		Status:      entity.ServerStatus(strings.TrimSpace(row[2])),
		IPv4:        strings.TrimSpace(row[3]),
		Description: strings.TrimSpace(row[4]),
		Location:    strings.TrimSpace(row[5]),
		OS:          strings.TrimSpace(row[6]),
	}

	// Validate required fields
	if server.ServerID == "" {
		return entity.Server{}, fmt.Errorf("server_id is required")
	}
	if server.ServerName == "" {
		return entity.Server{}, fmt.Errorf("server_name is required")
	}

	// intervalTimeString := strings.TrimSpace(row[7])
	// if intervalTimeString != "" {
	// 	intervalTime, err := strconv.ParseInt(intervalTimeString, 10, 64)
	// 	if err != nil {
	// 		return entity.Server{}, fmt.Errorf("invalid interval_time: %w", err)
	// 	}
	// 	server.IntervalTime = intervalTime
	// }

	// Validate status
	if server.Status != "" {
		validStatuses := []entity.ServerStatus{"ON", "OFF"}
		isValid := false
		for _, status := range validStatuses {
			if server.Status == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return entity.Server{}, fmt.Errorf("invalid status: %s (must be ON, OFF, or MAINTENANCE)", server.Status)
		}
	} else {
		server.Status = "OFF" // default status
	}

	return server, nil
}
