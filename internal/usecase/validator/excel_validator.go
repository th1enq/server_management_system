package validator

import (
	"fmt"
	"server_management_system/internal/domain/entity"
	"strconv"
	"strings"
)

// Validate validates the header row of Excel file
func Validate(row []string) error {
	expectedHeaders := []string{
		"server_id",
		"server_name",
		"status",
		"ipv4",
		"description",
		"location",
		"os",
		"cpu",
		"ram",
		"disk",
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
func ParseToServer(row []string) (entity.Server, error) {
	if len(row) < 10 {
		return entity.Server{}, fmt.Errorf("invalid row: expected at least 10 columns, got %d", len(row))
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

	// Parse CPU
	if row[7] != "" {
		cpu, err := strconv.Atoi(strings.TrimSpace(row[7]))
		if err != nil {
			return entity.Server{}, fmt.Errorf("invalid CPU value: %s", row[7])
		}
		server.CPU = cpu
	}

	// Parse RAM
	if row[8] != "" {
		ram, err := strconv.Atoi(strings.TrimSpace(row[8]))
		if err != nil {
			return entity.Server{}, fmt.Errorf("invalid RAM value: %s", row[8])
		}
		server.RAM = ram
	}

	// Parse Disk
	if row[9] != "" {
		disk, err := strconv.Atoi(strings.TrimSpace(row[9]))
		if err != nil {
			return entity.Server{}, fmt.Errorf("invalid Disk value: %s", row[9])
		}
		server.Disk = disk
	}

	// Validate status
	if server.Status != "" {
		validStatuses := []entity.ServerStatus{"ON", "OFF", "MAINTENANCE"}
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
