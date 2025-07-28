package repository

import (
	"context"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/tsdb"
)

type metricsRepository struct {
	tsdb tsdb.TSDBClient
}

func NewMetricsRepository(tsdb tsdb.TSDBClient) repository.MetricsRepository {
	return &metricsRepository{
		tsdb: tsdb,
	}
}

func (m *metricsRepository) SaveMetrics(ctx context.Context, metrics *entity.ServerMetrics) error {
	if err := m.tsdb.Write(ctx, metrics); err != nil {
		return err
	}
	return nil
}

func (m *metricsRepository) QueryMetrics(ctx context.Context, query string) (client.Result, error) {
	response, err := m.tsdb.Query(ctx, query)
	if err != nil {
		return client.Result{}, err
	}
	return response, nil
}
