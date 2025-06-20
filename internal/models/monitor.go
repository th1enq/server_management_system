package models

import "time"

type ServerUpTime struct {
	Server    Server  `json:"server"`
	AvgUpTime float64 `json:"avg_uptime"`
}

// DailyReport for email reports
type DailyReport struct {
	Date         time.Time      `json:"date"`
	TotalServers int64          `json:"total_servers"`
	OnlineCount  int64          `json:"online_count"`
	OfflineCount int64          `json:"offline_count"`
	AvgUptime    float64        `json:"avg_uptime_percentage"`
	Detail       []ServerUpTime `json:"detail_uptime"`
}
