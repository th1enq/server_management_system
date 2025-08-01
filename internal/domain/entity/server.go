package entity

import "time"

type ServerStatus string

const (
	ServerStatusOn        ServerStatus = "ON"
	ServerStatusOff       ServerStatus = "OFF"
	ServerStatusUndefined ServerStatus = "UNDEFINED"
)

type Server struct {
	ID           uint
	ServerID     string
	ServerName   string
	Status       ServerStatus
	IPv4         string
	Description  string
	Location     string
	OS           string
	IntervalTime int64
	CreatedAt    time.Time
}
