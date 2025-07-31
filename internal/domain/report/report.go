package report

import (
	"time"
)

type ServerUpTime struct {
	ServerID  string
	AvgUpTime float64
}

type DailyReport struct {
	StartOfDay   time.Time
	EndOfDay     time.Time
	TotalServers int64
	OnlineCount  int64
	OfflineCount int64
	AvgUptime    float64
	Detail       []ServerUpTime
}
