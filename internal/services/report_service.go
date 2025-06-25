package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type ReportService interface {
	GenerateReport(ctx context.Context, startOfDay, endOfDay time.Time) (*models.DailyReport, error)
	SendReportToEmail(ctx context.Context, report *models.DailyReport, emailTo, msg string) error
	SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error
	SendReportForDaily(ctx context.Context, date time.Time) error
}

type reportService struct {
	cfg       *configs.Email
	esClient  *elasticsearch.Client
	serverSrv ServerService
	logger    *zap.Logger
}

func NewReportService(cfg *configs.Email, esClient *elasticsearch.Client, serverSrv ServerService, logger *zap.Logger) ReportService {
	return &reportService{
		cfg:       cfg,
		logger:    logger,
		esClient:  esClient,
		serverSrv: serverSrv,
	}
}

// SendDailyReport implements ReportService.
func (s *reportService) SendReportToEmail(ctx context.Context, report *models.DailyReport, emailTo, msg string) error {
	emailTemplate, err := os.ReadFile("template/email.html")
	if err != nil {
		return fmt.Errorf("failed to read email template: %w", err)
	}

	tmpl := string(emailTemplate)

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, report); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.From)
	m.SetHeader("To", emailTo)
	m.SetHeader("Subject", msg)
	m.SetBody("text/html", buf.String())

	d := gomail.NewDialer(
		s.cfg.SMTPHost,
		s.cfg.SMTPPort,
		s.cfg.Username,
		s.cfg.Password,
	)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Report sent successfully",
		zap.String("emailTo", emailTo),
		zap.String("subject", msg),
	)

	return nil
}

// SendReportForDateRange implements ReportService.
func (s *reportService) SendReportForDateRange(ctx context.Context, startDate time.Time, endDate time.Time, emailTo string) error {
	if startDate.After(endDate) {
		return fmt.Errorf("start date must be before end date")
	}

	report, err := s.GenerateReport(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to generate report :%w", err)
	}

	msg := fmt.Sprintf("Server Report - %s to %s", startDate, endDate)

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}

func (s *reportService) SendReportForDaily(ctx context.Context, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report, err := s.GenerateReport(ctx, startOfDay, endOfDay)
	if err != nil {
		return fmt.Errorf("failed to generate report for daily: %w", err)
	}

	msg := fmt.Sprintf("Daily Server Report - %s", date.Format("2006-01-02"))
	emailTo := s.cfg.AdminEmail

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}

func (s *reportService) GenerateReport(ctx context.Context, startOfDay, endOfDay time.Time) (*models.DailyReport, error) {
	serverStats, err := s.serverSrv.GetServerStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server stats: %w", err)
	}

	onlineServer := serverStats["ON"]
	offlineServer := serverStats["OFF"]

	avgUpTime, detail, err := s.calculateAverageUptime(ctx, startOfDay, endOfDay)
	if err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to calculate average uptime")
		avgUpTime = 0
	}

	report := &models.DailyReport{
		StartOfDay:   startOfDay,
		EndOfDay:     endOfDay,
		TotalServers: onlineServer + offlineServer,
		OnlineCount:  onlineServer,
		OfflineCount: offlineServer,
		AvgUptime:    avgUpTime,
		Detail:       detail,
	}
	return report, nil
}

func (s *reportService) calculateAverageUptime(ctx context.Context, startTime, endTime time.Time) (float64, []models.ServerUpTime, error) {
	servers, err := s.serverSrv.GetAllServers(ctx)
	if err != nil {
		return 0, nil, err
	}
	if len(servers) == 0 {
		return 0, nil, nil
	}

	var singleUpTime []models.ServerUpTime

	totalUpTime := 0.0
	for _, server := range servers {
		uptime, err := s.calculateServerUpTime(ctx, &server.ServerID, startTime, endTime)
		if err != nil {
			s.logger.With(zap.Error(err)).Error("Failed to calculate uptime for server", zap.String("serverID", server.ServerID))
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

func (s *reportService) calculateServerUpTime(ctx context.Context, serverID *string, startTime, endTime time.Time) (float64, error) {
	// Validate input parameters
	if serverID == nil || *serverID == "" {
		return 0, fmt.Errorf("serverID cannot be nil or empty")
	}

	if startTime.After(endTime) {
		return 0, fmt.Errorf("startTime cannot be after endTime")
	}

	query := fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": [
					{ "term": { "server_id": "%s" }},
					{ "range": { "timestamp": { "gte": "%s", "lt": "%s" }}}
				]
			}
		},
		"sort": [
			{ "@timestamp": { "order": "asc" }}
		],
		"size": 10000
	}`, *serverID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	res, err := s.esClient.Search(
		s.esClient.Search.WithContext(ctx),
		s.esClient.Search.WithIndex("vcs-sms-server-checks-*"),
		s.esClient.Search.WithBody(strings.NewReader(query)),
		s.esClient.Search.WithTrackTotalHits(true),
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
			Hits []struct {
				Source models.ServerStatusLog `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle case when no data found
	if len(body.Hits.Hits) == 0 {
		// Return 0% uptime if no monitoring data exists
		return 0, nil
	}

	// Calculate uptime percentage
	var totalDuration time.Duration
	var uptimeDuration time.Duration

	// Get first log entry
	firstLog := body.Hits.Hits[0].Source
	lastTime := firstLog.CheckedAt
	lastStatus := firstLog.Status

	// Handle the period from startTime to first check
	if firstLog.CheckedAt.After(startTime) {
		initialDuration := firstLog.CheckedAt.Sub(startTime)
		totalDuration += initialDuration
		// Assume server was in same state as first recorded status
		if firstLog.Status == models.ServerStatusOn {
			uptimeDuration += initialDuration
		}
	} else {
		// First check is before or at startTime, use startTime as reference
		lastTime = startTime
	}

	// Process all status changes
	for i := 1; i < len(body.Hits.Hits); i++ {
		currentLog := body.Hits.Hits[i].Source

		// Calculate duration for previous status
		duration := currentLog.CheckedAt.Sub(lastTime)
		totalDuration += duration

		if lastStatus == models.ServerStatusOn {
			uptimeDuration += duration
		}

		lastTime = currentLog.CheckedAt
		lastStatus = currentLog.Status
	}

	// Handle the period from last check to endTime
	if lastTime.Before(endTime) {
		finalDuration := endTime.Sub(lastTime)
		totalDuration += finalDuration

		if lastStatus == models.ServerStatusOn {
			uptimeDuration += finalDuration
		}
	}

	// Avoid division by zero
	if totalDuration == 0 {
		return 0, nil
	}

	// Calculate percentage and ensure it's between 0-100
	percentage := float64(uptimeDuration) / float64(totalDuration) * 100
	if percentage < 0 {
		percentage = 0
	} else if percentage > 100 {
		percentage = 100
	}

	return percentage, nil
}
