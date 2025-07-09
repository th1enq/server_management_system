package tasks

import (
	"context"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type DailyReportTask interface {
	Execute(ctx context.Context) error
	GetName() string
	GetSchedule() string
}

type dailyReportTask struct {
	reportService usecases.IReportService
	cronConfig    configs.Cron
	logger        *zap.Logger
}

func NewDailyReportTask(
	reportService usecases.IReportService,
	cronConfig configs.Cron,
	logger *zap.Logger,
) DailyReportTask {
	return &dailyReportTask{
		reportService: reportService,
		cronConfig:    cronConfig,
		logger:        logger,
	}
}

func (t *dailyReportTask) Execute(ctx context.Context) error {
	t.logger.Info("Starting daily report task")

	start := time.Now()
	reportDate := time.Now().AddDate(0, 0, -1)

	err := t.reportService.SendReportForDaily(ctx, reportDate)
	duration := time.Since(start)

	if err != nil {
		t.logger.Error("Daily report task failed",
			zap.Error(err),
			zap.Time("report_date", reportDate),
			zap.Duration("duration", duration))
		return err
	}

	t.logger.Info("Daily report task completed successfully",
		zap.Time("report_date", reportDate),
		zap.Duration("duration", duration))
	return nil
}

func (t *dailyReportTask) GetName() string {
	return t.cronConfig.DailyReport.Name
}

func (t *dailyReportTask) GetSchedule() string {
	return t.cronConfig.DailyReport.Schedule
}
