package report

import (
	"time"

	"github.com/th1enq/server_management_system/internal/domain/entity"
)

type ServerUpTime struct {
	Server    entity.Server
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
