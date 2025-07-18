package dto

import (
	"github.com/th1enq/server_management_system/internal/domain/entity"
)

type CreateServerRequest struct {
	ServerID    string `json:"server_id" binding:"required"`
	ServerName  string `json:"server_name" binding:"required"`
	IPv4        string `json:"ipv4" binding:"required,ipv4"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	OS          string `json:"os,omitempty"`
	CPU         int    `json:"cpu,omitempty" binding:"omitempty,gte=0"`
	RAM         int    `json:"ram,omitempty" binding:"omitempty,gte=0"`  // in GB
	Disk        int    `json:"disk,omitempty" binding:"omitempty,gte=0"` // in GB
}

type UpdateServerRequest struct {
	ServerName  string `json:"server_name,omitempty"`
	IPv4        string `json:"ipv4" binding:"omitempty,ipv4"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	OS          string `json:"os,omitempty"`
	CPU         int    `json:"cpu,omitempty" binding:"omitempty,gte=0"`
	RAM         int    `json:"ram,omitempty" binding:"omitempty,gte=0"`
	Disk        int    `json:"disk,omitempty" binding:"omitempty,gte=0"`
}

// ServerFilter for filtering servers via query parameters
type ServerFilter struct {
	ServerID   string              `form:"server_id"`
	ServerName string              `form:"server_name"`
	Status     entity.ServerStatus `form:"status"`
	IPv4       string              `form:"ipv4" binding:"omitempty,ipv4"`
	Location   string              `form:"location"`
	OS         string              `form:"os"`
	CPU        int                 `form:"cpu" binding:"omitempty,gte=0"`
	RAM        int                 `form:"ram" binding:"omitempty,gte=0"`
	Disk       int                 `form:"disk" binding:"omitempty,gte=0"`
}

// ServerResponse for API responses
type ServerResponse struct {
	ServerID    string              `json:"server_id"`
	ServerName  string              `json:"server_name"`
	Status      entity.ServerStatus `json:"status"`
	IPv4        string              `json:"ipv4"`
	Description string              `json:"description,omitempty"`
	Location    string              `json:"location,omitempty"`
	OS          string              `json:"os,omitempty"`
	CPU         int                 `json:"cpu,omitempty"`
	RAM         int                 `json:"ram,omitempty"`
	Disk        int                 `json:"disk,omitempty"`
}

type ServerStatusResponse struct {
	TotalCount   int64 `json:"total_count"`
	OnlineCount  int64 `json:"online_count"`
	OfflineCount int64 `json:"offline_count"`
}

// Pagination parameters (for query)
type Pagination struct {
	Page     int    `form:"page" binding:"gte=1"`
	PageSize int    `form:"page_size" binding:"gte=1,lte=100"`
	Sort     string `form:"sort"`
	Order    string `form:"order" binding:"oneof=asc desc"`
}

// ServerListResponse for paginated API responses
type ServerListResponse struct {
	Total   int64           `json:"total"`
	Servers []entity.Server `json:"servers"`
	Page    int             `json:"page"`
	Size    int             `json:"size"`
}

// ImportResult for import operations
type ImportResult struct {
	SuccessCount   int      `json:"success_count"`
	SuccessServers []string `json:"success_servers"`
	FailureCount   int      `json:"failure_count"`
	FailureServers []string `json:"failure_servers"`
}

func FromEntityToServerResponse(server *entity.Server) *ServerResponse {
	if server == nil {
		return nil
	}
	return &ServerResponse{
		ServerID:    server.ServerID,
		ServerName:  server.ServerName,
		Status:      server.Status,
		IPv4:        server.IPv4,
		Description: server.Description,
		Location:    server.Location,
		OS:          server.OS,
		CPU:         server.CPU,
		RAM:         server.RAM,
		Disk:        server.Disk,
	}
}

func FromEntityListToServerResponseList(servers []entity.Server) []*ServerResponse {
	if len(servers) == 0 {
		return nil
	}

	responseList := make([]*ServerResponse, len(servers))
	for i, server := range servers {
		responseList[i] = FromEntityToServerResponse(&server)
	}
	return responseList
}
