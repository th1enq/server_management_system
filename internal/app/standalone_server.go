package app

import (
	"context"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/handler/http"
	"github.com/th1enq/server_management_system/internal/handler/jobs"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type StandaloneServer struct {
	server          http.Server
	logger          *zap.Logger
	cronConfig      configs.Cron
	sendDailyReport jobs.SendDailyReport
}

func NewStandaloneServer(httpServer http.Server, logger *zap.Logger, cronConfig configs.Cron, sendDailyReport jobs.SendDailyReport) *StandaloneServer {
	return &StandaloneServer{
		server:          httpServer,
		logger:          logger,
		cronConfig:      cronConfig,
		sendDailyReport: sendDailyReport,
	}
}

func (s *StandaloneServer) scheduleCronJobs(scheduler gocron.Scheduler) error {
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.SendDailyReport.Schedule, true),
		gocron.NewTask(func() {
			if err := s.sendDailyReport.Run(context.Background()); err != nil {
				s.logger.With(zap.Error(err)).Error("Failed to send daily report")
			}
		}),
	); err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to schedule SendDailyReport job")
		return err
	}
	return nil
}

func (s *StandaloneServer) Start() error {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to create scheduler")
		return err
	}

	defer func() {
		if shutdownErr := scheduler.Shutdown(); shutdownErr != nil {
			s.logger.With(zap.Error(shutdownErr)).Error("Failed to shutdown scheduler")
		}
	}()

	err = s.scheduleCronJobs(scheduler)
	if err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to schedule cron jobs")
		return err
	}

	go func() {
		httpStartErr := s.server.Start(context.Background())
		s.logger.With(zap.Error(httpStartErr)).Error("HTTP server failed to start")
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
	return nil
}
