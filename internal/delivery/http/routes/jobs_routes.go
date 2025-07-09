package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/interfaces/http/controllers"
	"github.com/th1enq/server_management_system/internal/middleware"
)

// JobsRoutes handles job monitoring routing
type JobsRoutes struct {
	jobsController *controllers.JobsController
	authMiddleware *middleware.AuthMiddleware
}

// NewJobsRoutes creates a new jobs routes handler
func NewJobsRoutes(
	jobsController *controllers.JobsController,
	authMiddleware *middleware.AuthMiddleware,
) *JobsRoutes {
	return &JobsRoutes{
		jobsController: jobsController,
		authMiddleware: authMiddleware,
	}
}

// Setup configures job monitoring routes
func (jr *JobsRoutes) Setup(rg *gin.RouterGroup) {
	// All job routes require authentication and admin permissions
	jobs := rg.Group("/jobs")
	jobs.Use(jr.authMiddleware.RequireAuth())

	// Job monitoring routes (read-only)
	jr.setupMonitoringRoutes(jobs)
}

// setupMonitoringRoutes configures job monitoring endpoints
func (jr *JobsRoutes) setupMonitoringRoutes(rg *gin.RouterGroup) {
	// Admin-only job monitoring (read-only)
	readScope := jr.authMiddleware.RequireAnyScope("admin:all", "jobs:read")

	// Job monitoring endpoints
	rg.GET("/", readScope, jr.jobsController.GetJobs)
	rg.GET("/status", readScope, jr.jobsController.GetJobStatus)

	// Note: No POST/PUT/DELETE endpoints because jobs run automatically
	// Jobs are controlled by scheduler, not HTTP API
}
