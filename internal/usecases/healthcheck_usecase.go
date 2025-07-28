package usecases

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/report"
	"go.uber.org/zap"
)

type HealthCheckUseCase interface {
	CalculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (*report.DailyReport, error)
	CalculateServerUpTime(ctx context.Context, serverID *string, startTime, endTime time.Time) (entity.ServerStatus, float64, error)
	CountLogStats(ctx context.Context, serverID *string, stat string, startTime, endTime time.Time) (int64, error)
	GetLastServerStatus(ctx context.Context, serverID *string, startTime, endTime time.Time) (entity.ServerStatus, error)
	ExportReportXLSX(ctx context.Context, report *report.DailyReport) (string, error)
}

type healthCheckUseCase struct {
	esClient      *elasticsearch.Client
	serverUseCase ServerUseCase
	logger        *zap.Logger
}

// CalculateAverageUptime implements HealthCheckUseCase.
func (h *healthCheckUseCase) CalculateAverageUptime(ctx context.Context, startTime time.Time, endTime time.Time) (*report.DailyReport, error) {
	panic("unimplemented")
}

// CalculateServerUpTime implements HealthCheckUseCase.
func (h *healthCheckUseCase) CalculateServerUpTime(ctx context.Context, serverID *string, startTime time.Time, endTime time.Time) (entity.ServerStatus, float64, error) {
	panic("unimplemented")
}

// CountLogStats implements HealthCheckUseCase.
func (h *healthCheckUseCase) CountLogStats(ctx context.Context, serverID *string, stat string, startTime time.Time, endTime time.Time) (int64, error) {
	panic("unimplemented")
}

// ExportReportXLSX implements HealthCheckUseCase.
func (h *healthCheckUseCase) ExportReportXLSX(ctx context.Context, report *report.DailyReport) (string, error) {
	panic("unimplemented")
}

// GetLastServerStatus implements HealthCheckUseCase.
func (h *healthCheckUseCase) GetLastServerStatus(ctx context.Context, serverID *string, startTime time.Time, endTime time.Time) (entity.ServerStatus, error) {
	panic("unimplemented")
}

func NewHealthCheckUseCase(esClient *elasticsearch.Client, serverUseCase ServerUseCase, logger *zap.Logger) HealthCheckUseCase {
	return &healthCheckUseCase{
		esClient:      esClient,
		serverUseCase: serverUseCase,
		logger:        logger,
	}
}
