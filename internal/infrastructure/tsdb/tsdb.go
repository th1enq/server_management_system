package tsdb

import (
	"context"
	"fmt"

	_ "github.com/influxdata/influxdb1-client"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"go.uber.org/zap"
)

type TSDBClient interface {
	Write(ctx context.Context, metrics *entity.ServerMetrics) error
	Query(ctx context.Context, query string) (client.Result, error)
}

type tsdbClient struct {
	client client.Client
	zap    *zap.Logger
}

func (t *tsdbClient) Query(ctx context.Context, query string) (client.Result, error) {
	q := client.NewQuery(query, "server_metrics", "s")
	response, err := t.client.Query(q)
	if err != nil {
		t.zap.Error("Failed to execute query", zap.Error(err))
		return client.Result{}, err
	}
	if response.Error() != nil {
		t.zap.Error("Query returned an error", zap.Error(response.Error()))
		return client.Result{}, response.Error()
	}
	return response.Results[0], nil
}

func (t *tsdbClient) Write(ctx context.Context, metrics *entity.ServerMetrics) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "server_metrics",
		Precision: "s",
	})
	if err != nil {
		t.zap.Error("Failed to create batch points", zap.Error(err))
		return fmt.Errorf("failed to create batch points: %w", err)
	}

	tags := map[string]string{"server_id": metrics.ServerID}
	fields := map[string]interface{}{
		"cpu":  metrics.CPU,
		"ram":  metrics.RAM,
		"disk": metrics.Disk,
	}

	pt, err := client.NewPoint("metrics", tags, fields)
	if err != nil {
		t.zap.Error("Failed to create new point", zap.Error(err))
		return fmt.Errorf("failed to create new point: %w", err)
	}

	bp.AddPoint(pt)

	if err := t.client.Write(bp); err != nil {
		t.zap.Error("Failed to write batch points", zap.Error(err))
		return fmt.Errorf("failed to write batch points: %w", err)
	}
	return nil
}

func NewTSDBClient(cfg configs.TSDB, logger *zap.Logger) TSDBClient {
	client, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		Username: cfg.User,
		Password: cfg.Password,
	})
	if err != nil {
		logger.Fatal("Failed to create InfluxDB client", zap.Error(err))
	}
	logger.Info("InfluxDB client created successfully", zap.String("host", cfg.Host), zap.String("user", cfg.User))

	return &tsdbClient{
		client: client,
		zap:    logger,
	}
}
