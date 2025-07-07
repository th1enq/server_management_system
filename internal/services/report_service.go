package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type IReportService interface {
	SendReportToEmail(ctx context.Context, report *models.DailyReport, emailTo, msg string) error
	SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error
	SendReportForDaily(ctx context.Context, date time.Time) error
}

type reportService struct {
	cfg                configs.Email
	healthCheckService IHealthCheckService
	serverService      IServerService
	logger             *zap.Logger
}

func NewReportService(cfg configs.Email, healthCheckService IHealthCheckService, logger *zap.Logger) IReportService {
	return &reportService{
		cfg:                cfg,
		healthCheckService: healthCheckService,
		logger:             logger,
	}
}

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

	filePath, err := s.healthCheckService.ExportReportXLSX(ctx, report)
	if err != nil {
		return fmt.Errorf("failed to export report to XLSX: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.From)
	m.SetHeader("To", emailTo)
	m.SetHeader("Subject", msg)
	m.SetBody("text/html", buf.String())
	m.Attach(filePath)

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

func (s *reportService) SendReportForDateRange(ctx context.Context, startDate time.Time, endDate time.Time, emailTo string) error {
	if startDate.After(endDate) {
		return fmt.Errorf("start date must be before end date")
	}

	report, err := s.healthCheckService.CalculateAverageUptime(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to generate report :%w", err)
	}

	msg := fmt.Sprintf("Server Report - %s to %s", startDate, endDate)

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}

func (s *reportService) SendReportForDaily(ctx context.Context, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report, err := s.healthCheckService.CalculateAverageUptime(ctx, startOfDay, endOfDay)
	if err != nil {
		return fmt.Errorf("failed to generate report for daily: %w", err)
	}

	msg := fmt.Sprintf("Daily Server Report - %s", date.Format("2006-01-02"))
	emailTo := s.cfg.AdminEmail

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}
