package tasks

import (
	"context"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type ServerHealthCheckTask interface {
	Execute(ctx context.Context) error
	GetName() string
	GetSchedule() string
}

type serverHealthCheckTask struct {
	serverService usecases.IServerService
	cronConfig    configs.Cron
	logger        *zap.Logger
}

func NewServerHealthCheckTask(
	serverService usecases.IServerService,
	cronConfig configs.Cron,
	logger *zap.Logger,
) ServerHealthCheckTask {
	return &serverHealthCheckTask{
		serverService: serverService,
		cronConfig:    cronConfig,
		logger:        logger,
	}
}

func (t *serverHealthCheckTask) Execute(ctx context.Context) error {
	t.logger.Info("Starting server health check task")

	start := time.Now()
	err := t.serverService.CheckServerStatus(ctx)
	duration := time.Since(start)

	if err != nil {
		t.logger.Error("Server health check task failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return err
	}

	t.logger.Info("Server health check task completed successfully",
		zap.Duration("duration", duration))
	return nil
}

func (t *serverHealthCheckTask) GetName() string {
	return t.cronConfig.HealthCheckServer.Name
}

func (t *serverHealthCheckTask) GetSchedule() string {
	return t.cronConfig.HealthCheckServer.Schedule
}
