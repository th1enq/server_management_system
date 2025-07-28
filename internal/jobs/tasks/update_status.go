package tasks

import (
	"context"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type UpdateStatusTask interface {
	Execute(ctx context.Context) error
	GetName() string
	GetSchedule() string
}

type updateStatusTask struct {
	serverUseCase usecases.ServerUseCase
	cronConfig    configs.Cron
	logger        *zap.Logger
}

func NewUpdateStatusTask(
	serverUseCase usecases.ServerUseCase,
	cronConfig configs.Cron,
	logger *zap.Logger,
) UpdateStatusTask {
	return &updateStatusTask{
		serverUseCase: serverUseCase,
		cronConfig:    cronConfig,
		logger:        logger,
	}
}

func (t *updateStatusTask) Execute(ctx context.Context) error {
	t.logger.Info("Starting update status task")

	if err := t.serverUseCase.RefreshStatus(ctx); err != nil {
		t.logger.Error("Update status task failed", zap.Error(err))
		return err
	}

	return nil
}

func (t *updateStatusTask) GetName() string {
	return t.cronConfig.UpdateStatus.Name
}

func (t *updateStatusTask) GetSchedule() string {
	return t.cronConfig.UpdateStatus.Schedule
}
