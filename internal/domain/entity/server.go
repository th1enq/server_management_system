package entity

import "time"

type ServerStatus string

const (
	ServerStatusOn  ServerStatus = "ON"
	ServerStatusOff ServerStatus = "OFF"
)

type Server struct {
	ID            uint
	ServerID      string
	ServerName    string
	Status        ServerStatus
	IPv4          string
	Description   string
	Location      string
	OS            string
	IntervalTime  int
	CreatedTime   time.Time
	LastHeartbeat time.Time
}
