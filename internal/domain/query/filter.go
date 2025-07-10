package query

import "github.com/th1enq/server_management_system/internal/domain/entity"

type ServerFilter struct {
	ServerID    string
	ServerName  string
	Status      entity.ServerStatus
	Description string
	IPv4        string
	Location    string
	OS          string
	CPU         int
	RAM         int
	Disk        int
}
