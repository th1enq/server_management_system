package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

type JobsRouter interface {
	RegisterRoutes(v1 *gin.RouterGroup)
}

type jobsRouter struct {
	jobsController *controllers.JobsController
	authMiddleware *middleware.AuthMiddleware
}

func NewJobsRouter(
	jobsController *controllers.JobsController,
	authMiddleware *middleware.AuthMiddleware,
) JobsRouter {
	return &jobsRouter{
		jobsController: jobsController,
		authMiddleware: authMiddleware,
	}
}

func (r *jobsRouter) RegisterRoutes(v1 *gin.RouterGroup) {
	// Admin-only job monitoring routes (read-only for monitoring background jobs)
	jobs := v1.Group("/jobs")
	jobs.Use(r.authMiddleware.RequireAuth())
	{
		// Only monitoring endpoints - jobs run automatically in background
		jobs.GET("/", r.authMiddleware.RequireAnyScope("admin:all", "jobs:read"), r.jobsController.GetJobs)
		jobs.GET("/status", r.authMiddleware.RequireAnyScope("admin:all", "jobs:read"), r.jobsController.GetJobStatus)
	}
}
