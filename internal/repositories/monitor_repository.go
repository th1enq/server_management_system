package repositories

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/robfig/cron/v3"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/models"
)

type MonitorRepository interface {
	GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error)
	SendDailyReport(ctx context.Context, report *models.DailyReport, emailTo string) error
	SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error
	StartDailyReportScheduler()
	StopDailyReportScheduler()
}

type monitorRepository struct {
	cfg      *config.Config
	esClient *elasticsearch.Client
	cron     *cron.Cron
}

func NewMonitorRepository(cfg *config.Config, esClient *elasticsearch.Client, cron *cron.Cron) *monitorRepository {
	return &monitorRepository{
		cfg:      cfg,
		esClient: esClient,
		cron:     cron,
	}
}

func (s *reportService) GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error) {
	
}