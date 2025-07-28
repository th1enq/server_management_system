package scheduler

import (
	"context"

	"github.com/th1enq/server_management_system/internal/jobs/tasks"
	"go.uber.org/zap"
)

type JobManager interface {
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop() error
	GetScheduler() JobScheduler
}

type jobManager struct {
	scheduler        JobScheduler
	dailyReportTask  tasks.DailyReportTask
	updateStatusTask tasks.UpdateStatusTask
	logger           *zap.Logger
}

func NewJobManager(
	scheduler JobScheduler,
	dailyReportTask tasks.DailyReportTask,
	updateStatusTask tasks.UpdateStatusTask,
	logger *zap.Logger,
) JobManager {
	return &jobManager{
		scheduler:        scheduler,
		dailyReportTask:  dailyReportTask,
		updateStatusTask: updateStatusTask,
		logger:           logger,
	}
}

func (jm *jobManager) Initialize(ctx context.Context) error {
	jm.logger.Info("Initializing job manager")

	taskList := []Task{
		jm.dailyReportTask,
		jm.updateStatusTask,
	}

	for _, task := range taskList {
		if err := jm.scheduler.AddTask(task); err != nil {
			jm.logger.Error("Failed to add task",
				zap.String("task", task.GetName()),
				zap.Error(err))
			return err
		}
	}

	jm.logger.Info("Job manager initialized successfully",
		zap.Int("total_tasks", len(taskList)))
	return nil
}

func (jm *jobManager) Start(ctx context.Context) error {
	jm.logger.Info("Starting job manager")

	if err := jm.Initialize(ctx); err != nil {
		return err
	}

	if err := jm.scheduler.Start(ctx); err != nil {
		jm.logger.Error("Failed to start job scheduler", zap.Error(err))
		return err
	}

	jm.logger.Info("Job manager started successfully")
	return nil
}

func (jm *jobManager) Stop() error {
	jm.logger.Info("Stopping job manager")

	if err := jm.scheduler.Stop(); err != nil {
		jm.logger.Error("Failed to stop job scheduler", zap.Error(err))
		return err
	}

	jm.logger.Info("Job manager stopped successfully")
	return nil
}

func (jm *jobManager) GetScheduler() JobScheduler {
	return jm.scheduler
}
