package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/robfig/cron/v3"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

type ReportService interface {
	GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error)
	SendDailyReport(ctx context.Context, report *models.DailyReport, emailTo string) error
	SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error
	StartDailyReportScheduler()
	StopDailyReportScheduler()
}

type reportService struct {
	cfg        *config.Config
	esClient   *elastic.Client
	cron       *cron.Cron
	serverRepo repositories.ServerRepository
}

func NewreportService(cfg *config.Config, esClient *elastic.Client, cron *cron.Cron, serverRepo repositories.ServerRepository) *reportService {
	return &reportService{
		cfg:        cfg,
		esClient:   esClient,
		cron:       cron,
		serverRepo: serverRepo,
	}
}

func (s *reportService) GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	totalServer, err := s.serverRepo.CountByStatus(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to count total server: %w", err)
	}

	onlineServer, err := s.serverRepo.CountByStatus(ctx, models.ServerStatusOn)
	if err != nil {
		return nil, fmt.Errorf("failed to count online server: %w", err)
	}

	offlineServer, err := s.serverRepo.CountByStatus(ctx, models.ServerStatusOff)
	if err != nil {
		return nil, fmt.Errorf("failed to count offline server: %w", err)
	}

	avgUpTime, detail, err := s.calculateAverageUptime(ctx, startOfDay, endOfDay)
	if err != nil {
		logger.Error("Failed to calculate average uptime", err)
		avgUpTime = 0
	}

	report := &models.DailyReport{
		Date:         date,
		TotalServers: totalServer,
		OnlineCount:  onlineServer,
		OfflineCount: offlineServer,
		AvgUptime:    avgUpTime,
		Detail:       detail,
	}
	return report, nil
}

func (s *reportService) calculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (float64, []models.ServerUpTime, error) {
	servers, err := s.serverRepo.GetAll(ctx)
	if err != nil {
		return 0, nil, err
	}
	if len(servers) == 0 {
		return 0, nil, nil
	}

	var singleUpTime []models.ServerUpTime

	totalUpTime := 0.0
	for _, server := range servers {
		uptime, err := s.calculateServerUpTime(ctx, server.ID, startTime, endTime)
		if err != nil {
			logger.Error("Failed to calculated server uptime", err, zap.String("server_id", server.ServerID))
			continue
		}
		singleUpTime = append(singleUpTime, models.ServerUpTime{
			Server:    server,
			AvgUpTime: uptime,
		})
		totalUpTime += uptime
	}
	return totalUpTime / float64(len(servers)), singleUpTime, nil
}

func (s *reportService) calculateServerUpTime(ctx context.Context, serverID uint, startTime, endTime time.Time) (float64, error) {
	query := elastic.NewBoolQuery().
		Must(
			elastic.NewTermQuery("server_id", serverID),
			elastic.NewRangeQuery("checked_at").
				Gte(startTime).
				Lt(endTime),
		)

	searchResult, err := s.esClient.Search().
		Index("server_status_logs").
		Query(query).
		Sort("checked_at", true).
		Size(10000).
		Do(ctx)

	if err != nil {
		return 0, err
	}
	if searchResult.Hits.TotalHits.Value == 0 {
		return 0, nil
	}

	// Calculate uptime percentage
	totalDuration := endTime.Sub(startTime)
	uptimeDuration := time.Duration(0)

	var lastStatus models.ServerStatus
	var lastTime time.Time

	for _, hit := range searchResult.Hits.Hits {
		var log models.ServerStatusLog
		if err := json.Unmarshal(hit.Source, &log); err != nil {
			continue
		}

		if lastTime.IsZero() {
			lastTime = log.CheckedAt
			lastStatus = log.Status
			continue
		}

		if lastStatus == models.ServerStatusOn {
			uptimeDuration += log.CheckedAt.Sub(lastTime)
		}

		lastTime = log.CheckedAt
		lastStatus = log.Status
	}

	// Add remaining time if last status was ON
	if lastStatus == models.ServerStatusOn && !lastTime.IsZero() {
		uptimeDuration += endTime.Sub(lastTime)
	}

	return float64(uptimeDuration) / float64(totalDuration) * 100, nil
}
