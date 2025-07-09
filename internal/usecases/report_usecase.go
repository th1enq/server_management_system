package usecases

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/domain"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type ReportUseCase interface {
	SendReportToEmail(ctx context.Context, report *domain.DailyReport, emailTo, msg string) error
	SendReportForDateRange(ctx context.Context, startDate, endDate time.Time, emailTo string) error
	SendReportForDaily(ctx context.Context, date time.Time) error
}

type reportUseCase struct {
	cfg                configs.Email
	healthCheckUseCase HealthCheckUseCase
	logger             *zap.Logger
}

func NewReportUseCase(cfg configs.Email, healthCheckUseCase HealthCheckUseCase, logger *zap.Logger) ReportUseCase {
	return &reportUseCase{
		cfg:                cfg,
		healthCheckUseCase: healthCheckUseCase,
		logger:             logger,
	}
}

func (s *reportUseCase) SendReportToEmail(ctx context.Context, report *domain.DailyReport, emailTo, msg string) error {
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

	filePath, err := s.healthCheckUseCase.ExportReportXLSX(ctx, report)
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

func (s *reportUseCase) SendReportForDateRange(ctx context.Context, startDate time.Time, endDate time.Time, emailTo string) error {
	if startDate.After(endDate) {
		return fmt.Errorf("start date must be before end date")
	}

	report, err := s.healthCheckUseCase.CalculateAverageUptime(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to generate report :%w", err)
	}

	msg := fmt.Sprintf("Server Report - %s to %s", startDate, endDate)

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}

func (s *reportUseCase) SendReportForDaily(ctx context.Context, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report, err := s.healthCheckUseCase.CalculateAverageUptime(ctx, startOfDay, endOfDay)
	if err != nil {
		return fmt.Errorf("failed to generate report for daily: %w", err)
	}

	msg := fmt.Sprintf("Daily Server Report - %s", date.Format("2006-01-02"))
	emailTo := s.cfg.AdminEmail

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}
