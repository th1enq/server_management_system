package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type HealthCheckUseCase interface {
	CalculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (*domain.DailyReport, error)
	CalculateServerUpTime(ctx context.Context, serverID *string, startTime, endTime time.Time) (float64, error)
	CountLogStats(ctx context.Context, serverID *string, stat string, startTime, endTime time.Time) (int64, error)
	ExportReportXLSX(ctx context.Context, report *domain.DailyReport) (string, error)
}

type healthCheckUseCase struct {
	esClient      *elasticsearch.Client
	serverUseCase ServerUseCase
	logger        *zap.Logger
}

func NewHealthCheckUseCase(esClient *elasticsearch.Client, serverUseCase ServerUseCase, logger *zap.Logger) HealthCheckUseCase {
	return &healthCheckUseCase{
		esClient:      esClient,
		serverUseCase: serverUseCase,
		logger:        logger,
	}
}

func (h *healthCheckUseCase) CalculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (*domain.DailyReport, error) {
	servers, err := h.serverUseCase.GetAllServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers: %w", err)
	}
	if len(servers) == 0 {
		return &domain.DailyReport{
			StartOfDay:   startTime,
			EndOfDay:     endTime,
			TotalServers: 0,
			OnlineCount:  0,
			OfflineCount: 0,
			AvgUptime:    0,
			Detail:       []domain.ServerUpTime{},
		}, nil
	}

	var singleUpTime []domain.ServerUpTime
	totalUpTime := 0.0
	onlineCount := 0
	offlineCount := 0
	for _, server := range servers {
		if server.Status == domain.ServerStatusOff {
			offlineCount++
		} else {
			onlineCount++
		}
		uptime, err := h.CalculateServerUpTime(ctx, &server.ServerID, startTime, endTime)
		if err != nil {
			h.logger.With(zap.Error(err)).Error("Failed to calculate uptime for server", zap.String("serverID", server.ServerID))
			continue
		}
		singleUpTime = append(singleUpTime, domain.ServerUpTime{
			Server:    server,
			AvgUpTime: uptime,
		})
		totalUpTime += uptime
	}
	avgUptime := totalUpTime / float64(len(servers))

	report := &domain.DailyReport{
		StartOfDay:   startTime,
		EndOfDay:     endTime,
		TotalServers: int64(len(servers)),
		OnlineCount:  int64(onlineCount),
		OfflineCount: int64(offlineCount),
		AvgUptime:    avgUptime,
		Detail:       singleUpTime,
	}

	return report, nil
}

func (h *healthCheckUseCase) CountLogStats(ctx context.Context, serverID *string, stat string, startTime, endTime time.Time) (int64, error) {
	query := fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": [
					{ "term": { "server_id": "%s" }},
					{ "range": { "timestamp": { "gte": "%s", "lt": "%s" }}}
				]
			}
		},
		"size": 10000
	}`, *serverID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	res, err := h.esClient.Search(
		h.esClient.Search.WithContext(ctx),
		h.esClient.Search.WithIndex("vcs-sms-server-checks-*"),
		h.esClient.Search.WithBody(strings.NewReader(query)),
		h.esClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return 0, fmt.Errorf("elasticsearch search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var body struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			}
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	total := body.Hits.Total.Value

	return total, nil
}

func (h *healthCheckUseCase) CalculateServerUpTime(ctx context.Context, serverID *string, startTime, endTime time.Time) (float64, error) {
	if startTime.After(endTime) {
		return 0, fmt.Errorf("startTime cannot be after endTime")
	}

	onlineCount, err := h.CountLogStats(ctx, serverID, "ON", startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("failed to count online logs: %w", err)
	}
	offlineCount, err := h.CountLogStats(ctx, serverID, "OFF", startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("failed to count offline logs: %w", err)
	}

	totalCount := onlineCount + offlineCount
	if totalCount == 0 {
		return 0, nil
	}

	return float64(onlineCount) / float64(totalCount) * 100, nil
}

func (h *healthCheckUseCase) ExportReportXLSX(ctx context.Context, report *domain.DailyReport) (string, error) {
	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		return "", err
	}

	streamWriter.SetRow("A1", []interface{}{
		"Server ID", "Server Name", "Status", "Description", "IPv4", "Disk", "RAM", "Location", "OS", "Uptime",
	})

	for rowIndex, detail := range report.Detail {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+2)
		err = streamWriter.SetRow(cell, []interface{}{
			detail.Server.ServerID,
			detail.Server.ServerName,
			detail.Server.Status,
			detail.Server.Description,
			detail.Server.IPv4,
			detail.Server.Disk,
			detail.Server.RAM,
			detail.Server.Location,
			detail.Server.OS,
			detail.AvgUpTime,
		})
		if err != nil {
			return "", err
		}
	}
	if err := streamWriter.Flush(); err != nil {
		return "", err
	}
	filePath := fmt.Sprintf("./exports/daily_report%d.xlsx", time.Now().Unix())
	if err := file.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	h.logger.Info("Daily report file exported successfully",
		zap.String("file_path", filePath),
	)

	return filePath, nil

}
