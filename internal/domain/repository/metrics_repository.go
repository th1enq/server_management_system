package repository

import (
	"context"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/th1enq/server_management_system/internal/domain/entity"
)

type MetricsRepository interface {
	SaveMetrics(ctx context.Context, metrics *entity.ServerMetrics) error
	QueryMetrics(ctx context.Context, query string) (client.Result, error)
}
