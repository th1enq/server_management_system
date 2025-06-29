package app

import (
	"context"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/th1enq/server_management_system/internal/api/http"
	"github.com/th1enq/server_management_system/internal/api/jobs"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type StandaloneServer struct {
	server              http.Server
	logger              *zap.Logger
	cronConfig          configs.Cron
	sendDailyReport     jobs.SendDailyReport
	intervalCheckStatus jobs.IntervalCheckStatus
}

func NewStandaloneServer(httpServer http.Server, logger *zap.Logger, cronConfig configs.Cron, sendDailyReport jobs.SendDailyReport, intervalStatusCheck jobs.IntervalCheckStatus) *StandaloneServer {
	return &StandaloneServer{
		server:              httpServer,
		logger:              logger,
		cronConfig:          cronConfig,
		sendDailyReport:     sendDailyReport,
		intervalCheckStatus: intervalStatusCheck,
	}
}

func (s *StandaloneServer) scheduleCronJobs(scheduler gocron.Scheduler) error {
	// Schedule SendDailyReport job
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.SendDailyReport.Schedule, true),
		gocron.NewTask(func() {
			s.logger.Info("Running SendDailyReport job")
			if err := s.sendDailyReport.Run(context.Background()); err != nil {
				s.logger.With(zap.Error(err)).Error("Failed to send daily report")
			} else {
				s.logger.Info("SendDailyReport job completed successfully")
			}
		}),
	); err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to schedule SendDailyReport job")
		return err
	}
	s.logger.With(zap.String("schedule", s.cronConfig.SendDailyReport.Schedule)).Info("SendDailyReport job scheduled")

	// Schedule IntervalCheckStatus job
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.IntervalCheckStatus.Schedule, true),
		gocron.NewTask(func() {
			s.logger.Info("Running IntervalCheckStatus job")
			if err := s.intervalCheckStatus.Run(context.Background()); err != nil {
				s.logger.With(zap.Error(err)).Error("Failed to check server status")
			} else {
				s.logger.Info("IntervalCheckStatus job completed successfully")
			}
		}),
	); err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to schedule IntervalCheckStatus job")
		return err
	}
	s.logger.With(zap.String("schedule", s.cronConfig.IntervalCheckStatus.Schedule)).Info("IntervalCheckStatus job scheduled")

	return nil
}

func (s *StandaloneServer) Start() error {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to create scheduler")
		return err
	}

	err = s.scheduleCronJobs(scheduler)
	if err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to schedule cron jobs")
		return err
	}

	// Start the scheduler
	scheduler.Start()
	s.logger.Info("Cron scheduler started successfully")

	// Start HTTP server in a goroutine
	go func() {
		httpStartErr := s.server.Start(context.Background())
		if httpStartErr != nil {
			s.logger.With(zap.Error(httpStartErr)).Error("HTTP server failed to start")
		}
	}()

	// Block until we receive a signal
	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)

	// Gracefully shutdown the scheduler
	s.logger.Info("Shutting down cron scheduler...")
	if shutdownErr := scheduler.Shutdown(); shutdownErr != nil {
		s.logger.With(zap.Error(shutdownErr)).Error("Failed to shutdown scheduler")
	} else {
		s.logger.Info("Cron scheduler shut down successfully")
	}

	return nil
}
