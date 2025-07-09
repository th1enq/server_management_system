package domain

import "time"

type ServerUpTime struct {
	Server    Server  `json:"server"`
	AvgUpTime float64 `json:"avg_uptime"`
}

type DailyReport struct {
	StartOfDay   time.Time      `json:"start_day"`
	EndOfDay     time.Time      `json:"end_day"`
	TotalServers int64          `json:"total_servers"`
	OnlineCount  int64          `json:"online_count"`
	OfflineCount int64          `json:"offline_count"`
	AvgUptime    float64        `json:"avg_uptime_percentage"`
	Detail       []ServerUpTime `json:"detail_uptime"`
}
