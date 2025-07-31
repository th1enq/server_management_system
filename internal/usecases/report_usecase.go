package usecases

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/domain/report"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type ReportUseCase interface {
	SendReportToEmail(ctx context.Context, report *report.DailyReport, emailTo, msg string) error
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

func (s *reportUseCase) SendReportToEmail(ctx context.Context, report *report.DailyReport, emailTo, msg string) error {
	emailTemplate, err := os.ReadFile("template/email.html")
	if err != nil {
		s.logger.Error("Failed to read email template", zap.Error(err))
		return fmt.Errorf("failed to read email template: %w", err)
	}

	tmpl := string(emailTemplate)

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		s.logger.Error("Failed to parse email template", zap.Error(err))
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, report); err != nil {
		s.logger.Error("Failed to execute email template", zap.Error(err))
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filePath, err := s.healthCheckUseCase.ExportReportXLSX(ctx, report)
	if err != nil {
		s.logger.Error("Failed to export report to XLSX", zap.Error(err))
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
		s.logger.Error("Failed to send email", zap.String("emailTo", emailTo), zap.String("subject", msg), zap.Error(err))
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
		s.logger.Warn("Start date must be before end date", zap.Time("startDate", startDate), zap.Time("endDate", endDate))
		return fmt.Errorf("start date must be before end date")
	}

	report, err := s.healthCheckUseCase.CalculateAverageUptime(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to generate report for date range", zap.Time("startDate", startDate), zap.Time("endDate", endDate), zap.Error(err))
		return fmt.Errorf("failed to generate report :%w", err)
	}

	s.logger.Info("Generated report for date range",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate),
		zap.String("emailTo", emailTo),
	)

	msg := fmt.Sprintf("Server Report - %s to %s", startDate, endDate)

	s.logger.Info("Sending report for date range",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate),
		zap.String("emailTo", emailTo),
		zap.String("subject", msg),
	)

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}

func (s *reportUseCase) SendReportForDaily(ctx context.Context, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report, err := s.healthCheckUseCase.CalculateAverageUptime(ctx, startOfDay, endOfDay)
	if err != nil {
		s.logger.Error("Failed to generate report for daily", zap.Time("date", date), zap.Error(err))
		return fmt.Errorf("failed to generate report for daily: %w", err)
	}

	msg := fmt.Sprintf("Daily Server Report - %s", date.Format("2006-01-02"))
	emailTo := s.cfg.AdminEmail

	s.logger.Info("Sending daily report",
		zap.Time("date", date),
		zap.String("emailTo", emailTo),
		zap.String("subject", msg),
	)

	return s.SendReportToEmail(ctx, report, emailTo, msg)
}
