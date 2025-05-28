package models

import (
	"time"

	"gorm.io/gorm"
)

type ServerStatus string

type Server struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ServerID    string         `gorm:"uniqueIndex;not null" json:"server_id"`
	ServerName  string         `gorm:"uniqueIndex;not null" json:"server_name"`
	Status      ServerStatus   `gorm:"default:OFF" json:"status"`
	IPv4        string         `json:"ipv4"`
	Description string         `json:"description,omitempty"`
	Location    string         `json:"location,omitempty"`
	OS          string         `json:"os,omitempty"`
	CPU         int            `json:"cpu,omitempty"`
	RAM         int            `json:"ram,omitempty"`  // in GB
	Disk        int            `json:"disk,omitempty"` // in GB
	CreatedTime time.Time      `gorm:"autoCreateTime" json:"created_time"`
	LastUpdated time.Time      `gorm:"autoUpdateTime" json:"last_updated"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ServerStatusLog for tracking status changes
type ServerStatusLog struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	ServerID     string       `gorm:"index;not null" json:"server_id"`
	Status       ServerStatus `json:"status"`
	CheckedAt    time.Time    `json:"checked_at"`
	ResponseTime int          `json:"response_time"` // in milliseconds
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
}

// ServerFilter for filtering servers
type ServerFilter struct {
	ServerID   string       `form:"server_id"`
	ServerName string       `form:"server_name"`
	Status     ServerStatus `form:"status"`
	IPv4       string       `form:"ipv4"`
	Location   string       `form:"location"`
	OS         string       `form:"os"`
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
	Total   int64    `json:"total"`
	Servers []Server `json:"servers"`
	Page    int      `json:"page"`
	Size    int      `json:"size"`
}

// ImportResult for import operations
type ImportResult struct {
	SuccessCount   int      `json:"success_count"`
	SuccessServers []string `json:"success_servers"`
	FailureCount   int      `json:"failure_count"`
	FailureServers []string `json:"failure_servers"`
}

// DailyReport for email reports
type DailyReport struct {
	Date         time.Time `json:"date"`
	TotalServers int64     `json:"total_servers"`
	OnlineCount  int64     `json:"online_count"`
	OfflineCount int64     `json:"offline_count"`
	AvgUptime    float64   `json:"avg_uptime_percentage"`
}
