package dto

import "server_management_system/internal/domain/entity"

// ServerFilter for filtering servers
type ServerFilter struct {
	ServerID   string              `form:"server_id"`
	ServerName string              `form:"server_name"`
	Status     entity.ServerStatus `form:"status"`
	IPv4       string              `form:"ipv4"`
	Location   string              `form:"location"`
	OS         string              `form:"os"`
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
