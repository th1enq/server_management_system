package repositories

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/robfig/cron/v3"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/models"
)

type MonitorService interface {
	GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error)
	SendDailyReport(ctx context.Context, report *models.DailyReport, emailTo string) error
	SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error
	StartDailyReportScheduler()
	StopDailyReportScheduler()
}

type monitorService struct {
	cfg      *config.Config
	esClient *elasticsearch.Client
	cron     *cron.Cron
}

func NewMonitorService(cfg *config.Config, esClient *elasticsearch.Client, cron *cron.Cron) *monitorRepository {
	return &monitorService{
		cfg:      cfg,
		esClient: esClient,
		cron:     cron,
	}
}

func (s *monitorService) GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	
}
