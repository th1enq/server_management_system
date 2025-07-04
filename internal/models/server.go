package models

import (
	"time"

	"gorm.io/gorm"
)

type ServerStatus string

const (
	ServerStatusOn  ServerStatus = "ON"
	ServerStatusOff ServerStatus = "OFF"
)

type Server struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ServerID    string         `gorm:"uniqueIndex;not null" json:"server_id" binding:"required"`
	ServerName  string         `gorm:"uniqueIndex;not null" json:"server_name" binding:"required"`
	Status      ServerStatus   `gorm:"default:OFF" json:"status"`
	IPv4        string         `json:"ipv4" binding:"required"`
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
	ServerID     string       `json:"server_id"`
	Status       ServerStatus `json:"status"`
	CheckedAt    time.Time    `json:"@timestamp"`
	ResponseTime int          `json:"response_time"` // in milliseconds
}
