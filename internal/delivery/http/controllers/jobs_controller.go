package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/jobs/scheduler"
	"go.uber.org/zap"
)

type JobsController struct {
	jobManager    scheduler.JobManager
	jobsPresenter presenters.JobsPresenter
	logger        *zap.Logger
}

func NewJobsController(
	jobManager scheduler.JobManager,
	jobsPresenter presenters.JobsPresenter,
	logger *zap.Logger,
) *JobsController {
	return &JobsController{
		jobManager:    jobManager,
		jobsPresenter: jobsPresenter,
		logger:        logger,
	}
}

// GetJobs godoc
// @Summary Get all scheduled jobs (monitoring only)
// @Description Get information about all background jobs and their schedules
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.APIResponse{data=[]scheduler.TaskInfo}
// @Failure 500 {object} domain.APIResponse
// @Router /api/v1/jobs [get]
func (jc *JobsController) GetJobs(c *gin.Context) {
	jc.logger.Info("Fetching job information for monitoring")
	tasks := jc.jobManager.GetScheduler().GetTasks()
	jc.jobsPresenter.JobsRetrieved(c, tasks)
}

// GetJobStatus godoc
// @Summary Get job scheduler status (monitoring only)
// @Description Get the current status of the background job scheduler
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.APIResponse{data=map[string]interface{}}
// @Failure 500 {object} domain.APIResponse
// @Router /api/v1/jobs/status [get]
func (jc *JobsController) GetJobStatus(c *gin.Context) {
	jc.logger.Info("Fetching job scheduler status for monitoring")
	isRunning := jc.jobManager.GetScheduler().IsRunning()
	tasks := jc.jobManager.GetScheduler().GetTasks()

	status := map[string]interface{}{
		"scheduler_running": isRunning,
		"total_tasks":       len(tasks),
		"tasks":             tasks,
	}
	jc.jobsPresenter.JobStatusRetrieved(c, status)
}
