package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
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
	esClient   *elasticsearch.Client
	serverRepo repositories.ServerRepository
}

// SendDailyReport implements ReportService.
func (s *reportService) SendDailyReport(ctx context.Context, report *models.DailyReport, emailTo string) error {
	panic("unimplemented")
}

// SendReportForDateRange implements ReportService.
func (s *reportService) SendReportForDateRange(ctx context.Context, startDate time.Time, endDate time.Time, emailTo string) error {
	panic("unimplemented")
}

// StartDailyReportScheduler implements ReportService.
func (s *reportService) StartDailyReportScheduler() {
	panic("unimplemented")
}

// StopDailyReportScheduler implements ReportService.
func (s *reportService) StopDailyReportScheduler() {
	panic("unimplemented")
}

func NewReportService(cfg *config.Config, esClient *elasticsearch.Client, serverRepo repositories.ServerRepository) ReportService {
	return &reportService{
		cfg:        cfg,
		esClient:   esClient,
		serverRepo: serverRepo,
	}
}

func (s *reportService) GenerateDailyReport(ctx context.Context, date time.Time) (*models.DailyReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

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
		TotalServers: onlineServer + offlineServer,
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
		uptime, err := s.calculateServerUpTime(ctx, server.ServerID, startTime, endTime)
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

func (s *reportService) calculateServerUpTime(ctx context.Context, serverID string, startTime, endTime time.Time) (float64, error) {
	query := `{
		"query": {
			"bool": {
				"must": [
					{ "term": { "server_id": "` + serverID + `" }},
					{ "range": { "@timestamp": { "gte": "` + startTime.Format(time.RFC3339) + `", "lt": "` + endTime.Format(time.RFC3339) + `" }}}
				]
			}
		},
		"sort": [
			{ "@timestamp": { "order": "asc" }}
		],
		"size": 10000
	}`

	fmt.Println("Elasticsearch query:", query)

	res, err := s.esClient.Search(
		s.esClient.Search.WithContext(ctx),
		s.esClient.Search.WithIndex("vcs-sms-server-checks-*"),
		s.esClient.Search.WithBody(strings.NewReader(query)),
		s.esClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var body struct {
		Hits struct {
			Hits []struct {
				Source models.ServerStatusLog `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return 0, err
	}

	if len(body.Hits.Hits) == 0 {
		return 0, nil
	}

	// Calculate uptime percentage
	totalDuration := endTime.Sub(startTime)
	uptimeDuration := time.Duration(0)

	var lastStatus models.ServerStatus
	var lastTime time.Time

	for _, hit := range body.Hits.Hits {
		log := hit.Source
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

	if lastStatus == models.ServerStatusOn && !lastTime.IsZero() {
		uptimeDuration += endTime.Sub(lastTime)
	}

	return float64(uptimeDuration) / float64(totalDuration) * 100, nil
}
