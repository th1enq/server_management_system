package usecases

import (
	"context"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"go.uber.org/zap"
)

type MetricsUseCase interface {
	Insert(ctx context.Context, metrics producer.MonitoringMessage) error
	SaveMetrics(ctx context.Context, metrics *entity.ServerMetrics) error
	QueryMetrics(ctx context.Context, query string) (client.Result, error)
}

type metricsUseCase struct {
	metricsRepo repository.MetricsRepository
	logger      *zap.Logger
}

func NewMetricsUseCase(metricsRepo repository.MetricsRepository, logger *zap.Logger) MetricsUseCase {
	return &metricsUseCase{
		metricsRepo: metricsRepo,
		logger:      logger,
	}
}

func (m *metricsUseCase) Insert(ctx context.Context, metrics producer.MonitoringMessage) error {
	m.logger.Info("Inserting metrics", zap.Any("event", metrics))
	return nil
}

func (m *metricsUseCase) SaveMetrics(ctx context.Context, metrics *entity.ServerMetrics) error {
	return m.metricsRepo.SaveMetrics(ctx, metrics)
}

func (m *metricsUseCase) QueryMetrics(ctx context.Context, query string) (client.Result, error) {
	return m.metricsRepo.QueryMetrics(ctx, query)
}
