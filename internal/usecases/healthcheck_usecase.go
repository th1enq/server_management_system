package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/report"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"go.uber.org/zap"
)

const (
	indexName = "server_uptime"
)

type HealthCheckUseCase interface {
	InsertUptime(ctx context.Context, msg producer.StatusChangeMessage) error
	CalculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (*report.DailyReport, error)
	CalculateServerUpTime(ctx context.Context, serverID *string, startTime, endTime time.Time) (entity.ServerStatus, float64, error)
	CountLogStats(ctx context.Context, serverID *string, stat string, startTime, endTime time.Time) (int64, error)
	GetLastServerStatus(ctx context.Context, serverID *string, startTime, endTime time.Time) (entity.ServerStatus, error)
	ExportReportXLSX(ctx context.Context, report *report.DailyReport) (string, error)
}

type healthCheckUseCase struct {
	esClient *elasticsearch.Client
	logger   *zap.Logger
}

func (h *healthCheckUseCase) InsertUptime(ctx context.Context, msg producer.StatusChangeMessage) error {
	if msg.OldStatus == msg.NewStatus {
		h.logger.Warn("No status change detected, skipping uptime insertion",
			zap.String("server_id", msg.ServerID),
			zap.String("old_status", string(msg.OldStatus)),
			zap.String("new_status", string(msg.NewStatus)),
			zap.Time("timestamp", msg.Timestamp),
		)
		return nil
	}
	newDocument := map[string]interface{}{
		"server_id": msg.ServerID,
		"status":    msg.NewStatus,
		"timestamp": msg.Timestamp,
	}
	newDocumentLine, err := json.Marshal(newDocument)
	if err != nil {
		h.logger.Error("Failed to marshal uptime document", zap.Error(err))
		return fmt.Errorf("failed to marshal uptime document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: msg.ServerID + "-" + msg.Timestamp.Format(time.RFC3339),
		Body:       bytes.NewReader(newDocumentLine),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, h.esClient)
	if err != nil {
		h.logger.Error("Failed to index uptime document", zap.Error(err))
		return fmt.Errorf("failed to index uptime document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		h.logger.Error("Error indexing uptime document", zap.String("status", res.Status()))
		return fmt.Errorf("error indexing uptime document: %s", res.Status())
	}

	h.logger.Info("Successfully indexed uptime document",
		zap.String("server_id", msg.ServerID),
		zap.String("status", string(msg.NewStatus)),
		zap.Time("timestamp", msg.Timestamp),
	)

	return nil
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

func NewHealthCheckUseCase(esClient *elasticsearch.Client, logger *zap.Logger) HealthCheckUseCase {
	return &healthCheckUseCase{
		esClient: esClient,
		logger:   logger,
	}
}
