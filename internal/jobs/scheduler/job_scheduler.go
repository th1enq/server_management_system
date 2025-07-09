package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Task interface {
	Execute(ctx context.Context) error
	GetName() string
	GetSchedule() string
}

type JobScheduler interface {
	Start(ctx context.Context) error
	Stop() error
	AddTask(task Task) error
	RemoveTask(taskName string) error
	GetTasks() []TaskInfo
	IsRunning() bool
}

type TaskInfo struct {
	Name     string    `json:"name"`
	Schedule string    `json:"schedule"`
	NextRun  time.Time `json:"next_run"`
	LastRun  time.Time `json:"last_run"`
	Status   string    `json:"status"`
}

type jobScheduler struct {
	cron   *cron.Cron
	tasks  map[string]taskEntry
	logger *zap.Logger
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

type taskEntry struct {
	task    Task
	entryID cron.EntryID
	lastRun time.Time
	status  string
}

func NewJobScheduler(logger *zap.Logger) JobScheduler {
	return &jobScheduler{
		cron:   cron.New(cron.WithSeconds()),
		tasks:  make(map[string]taskEntry),
		logger: logger,
	}
}

func (js *jobScheduler) Start(ctx context.Context) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	js.ctx, js.cancel = context.WithCancel(ctx)
	js.cron.Start()

	js.logger.Info("Job scheduler started")
	return nil
}

func (js *jobScheduler) Stop() error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	if js.cancel != nil {
		js.cancel()
	}

	stopCtx := js.cron.Stop()
	<-stopCtx.Done()

	js.logger.Info("Job scheduler stopped")
	return nil
}

func (js *jobScheduler) AddTask(task Task) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	taskName := task.GetName()
	schedule := task.GetSchedule()

	if existing, exists := js.tasks[taskName]; exists {
		js.cron.Remove(existing.entryID)
		js.logger.Info("Removed existing task", zap.String("task", taskName))
	}

	entryID, err := js.cron.AddFunc(schedule, js.wrapTask(task))
	if err != nil {
		js.logger.Error("Failed to add task",
			zap.String("task", taskName),
			zap.String("schedule", schedule),
			zap.Error(err))
		return err
	}

	js.tasks[taskName] = taskEntry{
		task:    task,
		entryID: entryID,
		status:  "scheduled",
	}

	js.logger.Info("Task added successfully",
		zap.String("task", taskName),
		zap.String("schedule", schedule))
	return nil
}

func (js *jobScheduler) RemoveTask(taskName string) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	entry, exists := js.tasks[taskName]
	if !exists {
		js.logger.Warn("Task not found", zap.String("task", taskName))
		return nil
	}

	js.cron.Remove(entry.entryID)
	delete(js.tasks, taskName)

	js.logger.Info("Task removed", zap.String("task", taskName))
	return nil
}

func (js *jobScheduler) GetTasks() []TaskInfo {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	var tasks []TaskInfo
	entries := js.cron.Entries()

	for taskName, entry := range js.tasks {
		var nextRun time.Time
		for _, cronEntry := range entries {
			if cronEntry.ID == entry.entryID {
				nextRun = cronEntry.Next
				break
			}
		}

		tasks = append(tasks, TaskInfo{
			Name:     taskName,
			Schedule: entry.task.GetSchedule(),
			NextRun:  nextRun,
			LastRun:  entry.lastRun,
			Status:   entry.status,
		})
	}

	return tasks
}

func (js *jobScheduler) IsRunning() bool {
	js.mutex.RLock()
	defer js.mutex.RUnlock()
	return js.ctx != nil && js.ctx.Err() == nil
}

func (js *jobScheduler) wrapTask(task Task) func() {
	return func() {
		taskName := task.GetName()

		js.logger.Info("Starting task execution", zap.String("task", taskName))

		js.updateTaskStatus(taskName, "running")

		start := time.Now()
		err := task.Execute(js.ctx)
		duration := time.Since(start)

		js.updateTaskLastRun(taskName, start)

		if err != nil {
			js.logger.Error("Task execution failed",
				zap.String("task", taskName),
				zap.Duration("duration", duration),
				zap.Error(err))
			js.updateTaskStatus(taskName, "failed")
		} else {
			js.logger.Info("Task execution completed",
				zap.String("task", taskName),
				zap.Duration("duration", duration))
			js.updateTaskStatus(taskName, "completed")
		}
	}
}

func (js *jobScheduler) updateTaskStatus(taskName, status string) {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	if entry, exists := js.tasks[taskName]; exists {
		entry.status = status
		js.tasks[taskName] = entry
	}
}

func (js *jobScheduler) updateTaskLastRun(taskName string, lastRun time.Time) {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	if entry, exists := js.tasks[taskName]; exists {
		entry.lastRun = lastRun
		js.tasks[taskName] = entry
	}
}
