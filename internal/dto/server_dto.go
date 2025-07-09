package dto

import "github.com/th1enq/server_management_system/internal/domain"

// CreateServerRequest for creating a new server
type CreateServerRequest struct {
	ServerID    string `json:"server_id" binding:"required"`
	ServerName  string `json:"server_name" binding:"required"`
	IPv4        string `json:"ipv4" binding:"required"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	OS          string `json:"os,omitempty"`
	CPU         int    `json:"cpu,omitempty"`
	RAM         int    `json:"ram,omitempty"`  // in GB
	Disk        int    `json:"disk,omitempty"` // in GB
}

// UpdateServerRequest for updating an existing server
type UpdateServerRequest struct {
	ServerName  *string `json:"server_name,omitempty"`
	IPv4        *string `json:"ipv4,omitempty"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
	OS          *string `json:"os,omitempty"`
	CPU         *int    `json:"cpu,omitempty"`
	RAM         *int    `json:"ram,omitempty"`
	Disk        *int    `json:"disk,omitempty"`
}

// ServerFilter for filtering servers
type ServerFilter struct {
	ServerID   string              `form:"server_id"`
	ServerName string              `form:"server_name"`
	Status     domain.ServerStatus `form:"status"`
	IPv4       string              `form:"ipv4"`
	Location   string              `form:"location"`
	OS         string              `form:"os"`
	CPU        int                 `json:"cpu,omitempty"`
	RAM        int                 `json:"ram,omitempty"`
	Disk       int                 `json:"disk,omitempty"`
}

// ServerResponse for API responses
type ServerResponse struct {
	ID          uint                `json:"id"`
	ServerID    string              `json:"server_id"`
	ServerName  string              `json:"server_name"`
	Status      domain.ServerStatus `json:"status"`
	IPv4        string              `json:"ipv4"`
	Description string              `json:"description,omitempty"`
	Location    string              `json:"location,omitempty"`
	OS          string              `json:"os,omitempty"`
	CPU         int                 `json:"cpu,omitempty"`
	RAM         int                 `json:"ram,omitempty"`
	Disk        int                 `json:"disk,omitempty"`
	CreatedTime string              `json:"created_time"`
	LastUpdated string              `json:"last_updated"`
}

// Pagination parameters
type Pagination struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Sort     string `form:"sort,default=created_time"`
	Order    string `form:"order,default=desc"`
}

// ServerListResponse for API responses
type ServerListResponse struct {
	Total   int64           `json:"total"`
	Servers []domain.Server `json:"servers"`
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
