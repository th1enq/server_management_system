package jobs

import (
	"context"
	"time"

	"github.com/th1enq/server_management_system/internal/services"
)

type SendDailyReport interface {
	Run(context.Context) error
}

type sendDailyReport struct {
	reportService services.ReportService
}

func NewSendDailyReport(reportService services.ReportService) SendDailyReport {
	return &sendDailyReport{
		reportService: reportService,
	}
}
func (s *sendDailyReport) Run(ctx context.Context) error {
	return s.reportService.SendReportForDaily(ctx, time.Now())
}
