package usecases

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/domain/report"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"github.com/th1enq/server_management_system/internal/infrastructure/search"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

const (
	indexName = "server_uptime"
)

type HealthCheckUseCase interface {
	InsertUptime(ctx context.Context, msg producer.StatusChangeMessage) error
	CalculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (*report.DailyReport, error)
	ExportReportXLSX(ctx context.Context, report *report.DailyReport) (string, error)
}

type healthCheckUseCase struct {
	esClient      search.IESClient
	serverUseCase ServerUseCase
	logger        *zap.Logger
}

func (h *healthCheckUseCase) CalculateAverageUptime(ctx context.Context, startTime time.Time, endTime time.Time) (*report.DailyReport, error) {
	serverIDs, err := h.serverUseCase.GetServerIDs(ctx)
	if err != nil {
		h.logger.Error("Failed to get server IDs", zap.Error(err))
		return nil, fmt.Errorf("failed to get server IDs: %w", err)
	}

	uptimeRateAvg := 0.0
	totalServers := 0
	onlineServers := 0
	detailUptime := make([]report.ServerUpTime, 0)

	workerpool := workerpool.New(15)
	var mu sync.Mutex

	for _, serverID := range serverIDs {
		id := serverID
		workerpool.Submit(func() {
			status, uptime, err := h.calculateServerUptime(ctx, id, startTime, endTime)
			if err != nil {
				h.logger.Error("Failed to calculate server uptime", zap.String("server_id", id), zap.Error(err))
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if status == entity.ServerStatusOn {
				onlineServers++
			}
			uptimeRateAvg += uptime
			totalServers++
			detailUptime = append(detailUptime, report.ServerUpTime{
				ServerID:  serverID,
				AvgUpTime: uptime,
			})
		})
	}

	workerpool.StopWait()

	var avgUptime float64 = 0
	if totalServers > 0 {
		avgUptime = uptimeRateAvg / float64(totalServers)
	}

	return &report.DailyReport{
		StartOfDay:   startTime,
		EndOfDay:     endTime,
		TotalServers: int64(totalServers),
		OnlineCount:  int64(onlineServers),
		OfflineCount: int64(totalServers - onlineServers),
		AvgUptime:    avgUptime,
		Detail:       detailUptime,
	}, nil
}

func (h *healthCheckUseCase) calculateServerUptime(ctx context.Context, serverID string, startTime, endTime time.Time) (entity.ServerStatus, float64, error) {
	logs, err := h.GetEsStatus(ctx, serverID, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get ES status", zap.Error(err))
		return entity.ServerStatusUndefined, 0, fmt.Errorf("failed to get ES status: %w", err)
	}
	uptime := 0
	lastStatus, err := h.getLastStatus(ctx, serverID, startTime)

	h.logger.Info("Calculating uptime for server",
		zap.String("server_id", serverID),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.Any("logs", logs),
	)

	if len(logs) == 0 {
		if err != nil {
			h.logger.Error("Failed to get last status", zap.Error(err))
			return entity.ServerStatusUndefined, 0, fmt.Errorf("failed to get last status: %w", err)
		}
		if lastStatus == entity.ServerStatusOn {
			uptime = int(endTime.Sub(startTime).Seconds())
		} else {
			uptime = 0
		}
	} else {
		switch logs[0].Status {
		case entity.ServerStatusOff:
			if lastStatus == entity.ServerStatusOn {
				uptime = int(logs[0].Timestamp.Sub(startTime).Seconds())
			} else {
				uptime = 0
			}
		}
		for i := 0; i < len(logs)-1; i++ {
			if logs[i].Status == entity.ServerStatusOn {
				uptime += int(logs[i+1].Timestamp.Sub(logs[i].Timestamp).Seconds())
			}
		}
		if logs[len(logs)-1].Status == entity.ServerStatusOn {
			uptime += int(endTime.Sub(logs[len(logs)-1].Timestamp).Seconds())
		}
	}

	h.logger.Info("Calculated uptime for server",
		zap.String("server_id", serverID),
		zap.Int("uptime_seconds", uptime),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)

	status := entity.ServerStatusOff

	uptimeRate := float64(uptime) / float64(endTime.Sub(startTime).Seconds()) * 100
	if (uptimeRate >= 70.0) && (uptime > 0) {
		status = entity.ServerStatusOn
	}

	return status, uptimeRate, nil

}

func (h *healthCheckUseCase) ExportReportXLSX(ctx context.Context, report *report.DailyReport) (string, error) {
	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		h.logger.Error("Failed to create stream writer", zap.Error(err))
		return "", fmt.Errorf("failed to create stream writer: %w", err)
	}

	streamWriter.SetRow("A1", []interface{}{"Server ID", "Average Uptime (%)"})
	for i, serverUptime := range report.Detail {
		row := fmt.Sprintf("A%d", i+2)
		err = streamWriter.SetRow(row, []interface{}{serverUptime.ServerID, serverUptime.AvgUpTime})
		if err != nil {
			h.logger.Error("Failed to set row in stream writer", zap.Error(err))
			return "", fmt.Errorf("failed to set row in stream writer: %w", err)
		}
	}
	if err := streamWriter.Flush(); err != nil {
		h.logger.Error("Failed to flush stream writer", zap.Error(err))
		return "", fmt.Errorf("failed to flush stream writer: %w", err)
	}
	fileName := fmt.Sprintf("report_%s.xlsx", time.Now().Format("20060102_150405"))
	if err := file.SaveAs(fileName); err != nil {
		h.logger.Error("Failed to save XLSX file", zap.Error(err))
		return "", fmt.Errorf("failed to save XLSX file: %w", err)
	}
	h.logger.Info("Report exported successfully", zap.String("file_name", fileName))
	return fileName, nil
}

func (h *healthCheckUseCase) InsertUptime(ctx context.Context, msg producer.StatusChangeMessage) error {
	if msg.OldStatus == msg.NewStatus {
		h.logger.Info("No status change detected, skipping indexing",
			zap.String("server_id", msg.ServerID),
			zap.String("old_status", string(msg.OldStatus)),
			zap.String("new_status", string(msg.NewStatus)),
		)
		return nil
	}
	newDocument := map[string]interface{}{
		"server_id": msg.ServerID,
		"status":    msg.NewStatus,
		"timestamp": msg.Timestamp.Format(time.RFC3339),
	}

	if err := h.esClient.Insert(ctx,
		indexName,
		fmt.Sprintf("%s-%s", msg.ServerID, msg.Timestamp.Format(time.RFC3339)),
		newDocument,
	); err != nil {
		return err
	}

	return nil
}

func (h *healthCheckUseCase) GetEsStatus(ctx context.Context, serverID string, startTime, endTime time.Time) ([]dto.EsStatus, error) {
	query := map[string]interface{}{
		"size": 10000,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"term": map[string]interface{}{"server_id.keyword": serverID}},
					{"range": map[string]interface{}{
						"timestamp": map[string]interface{}{
							"gte": startTime.Format(time.RFC3339),
							"lte": endTime.Format(time.RFC3339),
						},
					}},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"timestamp": map[string]string{"order": "asc"}},
		},
	}

	var parsedResult struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Status    entity.ServerStatus `json:"status"`
					Timestamp time.Time           `json:"timestamp"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := h.esClient.Exec(ctx, indexName, query, &parsedResult); err != nil {
		return nil, err
	}

	results := make([]dto.EsStatus, 0)
	for _, hit := range parsedResult.Hits.Hits {
		results = append(results, dto.EsStatus{
			Status:    hit.Source.Status,
			Timestamp: hit.Source.Timestamp,
		})
	}

	return results, nil
}

func (h *healthCheckUseCase) getLastStatus(ctx context.Context, serverID string, startTime time.Time) (entity.ServerStatus, error) {
	query := map[string]interface{}{
		"size": 1,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"term": map[string]interface{}{"server_id.keyword": serverID}},
					{"range": map[string]interface{}{
						"timestamp": map[string]interface{}{
							"lt": startTime.Format(time.RFC3339),
						},
					}},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"timestamp": map[string]string{"order": "desc"}},
		},
	}

	var parsedResult struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Status entity.ServerStatus `json:"status"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := h.esClient.Exec(ctx, indexName, query, &parsedResult); err != nil {
		return entity.ServerStatusOff, nil
	}

	if len(parsedResult.Hits.Hits) == 0 {
		h.logger.Info("No previous status found for server", zap.String("server_id", serverID))
		return entity.ServerStatusOff, nil
	}
	return parsedResult.Hits.Hits[0].Source.Status, nil
}

func NewHealthCheckUseCase(esClient search.IESClient, serverUseCase ServerUseCase, logger *zap.Logger) HealthCheckUseCase {
	return &healthCheckUseCase{
		esClient:      esClient,
		serverUseCase: serverUseCase,
		logger:        logger,
	}
}
