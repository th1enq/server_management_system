package models

import (
	"time"

	"github.com/th1enq/server_management_system/internal/domain/entity"
)

type Server struct {
	ID            uint   `gorm:"primaryKey"`
	ServerID      string `gorm:"index;uniqueIndex;not null"`
	ServerName    string `gorm:"index;uniqueIndex;not null"`
	Status        string `gorm:"not null;default:'OFF'"`
	IPv4          string
	Description   string
	Location      string
	OS            string
	IntervalTime  int       `gorm:"default:10"`
	CreatedTime   time.Time `gorm:"autoCreateTime"`
	LastHeartbeat time.Time `gorm:"index"`
}

func FromServerEntity(s *entity.Server) *Server {
	return &Server{
		ID:            s.ID,
		ServerID:      s.ServerID,
		ServerName:    s.ServerName,
		Status:        string(s.Status),
		IPv4:          s.IPv4,
		Description:   s.Description,
		Location:      s.Location,
		OS:            s.OS,
		IntervalTime:  s.IntervalTime,
		CreatedTime:   s.CreatedTime,
		LastHeartbeat: s.LastHeartbeat,
	}
}

func FromServerEntities(servers []entity.Server) []Server {
	var models []Server
	for _, s := range servers {
		models = append(models, *FromServerEntity(&s))
	}
	return models
}

func ToServerEntity(s *Server) *entity.Server {
	return &entity.Server{
		ID:            s.ID,
		ServerID:      s.ServerID,
		ServerName:    s.ServerName,
		Status:        entity.ServerStatus(s.Status),
		IPv4:          s.IPv4,
		Description:   s.Description,
		Location:      s.Location,
		OS:            s.OS,
		IntervalTime:  s.IntervalTime,
		CreatedTime:   s.CreatedTime,
		LastHeartbeat: s.LastHeartbeat,
	}
}

func ToServerEntities(servers []Server) []*entity.Server {
	var entities []*entity.Server
	for _, s := range servers {
		entities = append(entities, ToServerEntity(&s))
	}
	return entities
}
