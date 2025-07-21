package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

type Controller struct {
	serverController *controllers.ServerController
	reportController *controllers.ReportController
	authController   *controllers.AuthController
	userController   *controllers.UserController
	jobsController   *controllers.JobsController
	authMiddleware   *middleware.AuthMiddleware
}

func NewController(serverController *controllers.ServerController, reportController *controllers.ReportController, authController *controllers.AuthController, userController *controllers.UserController, jobsController *controllers.JobsController, authMiddleware *middleware.AuthMiddleware) *Controller {
	return &Controller{
		serverController: serverController,
		reportController: reportController,
		authController:   authController,
		userController:   userController,
		jobsController:   jobsController,
		authMiddleware:   authMiddleware,
	}
}

func (h *Controller) RegisterRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.authController.Login)
		auth.POST("/register", h.authController.Register)
		auth.POST("/refresh", h.authController.RefreshToken)
	}

	// Protected auth routes
	authProtected := v1.Group("/auth")
	authProtected.Use(h.authMiddleware.RequireAuth())
	{
		authProtected.POST("/logout", h.authController.Logout)
	}

	// Protected server routes
	servers := v1.Group("/servers")
	{
		servers.POST("/register", h.serverController.Register)
	}

	servers.Use(h.authMiddleware.RequireAuth())
	{
		// Read operations - requires server:read scope
		servers.GET("/", h.authMiddleware.RequireAnyScope("admin:all", "server:read"), h.serverController.ListServer)
		servers.GET("/export", h.authMiddleware.RequireAnyScope("admin:all", "server:export"), h.serverController.ExportServers)

		// Write operations - requires server:write scope
		servers.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "server:write"), h.serverController.CreateServer)
		servers.PUT("/:id", h.authMiddleware.RequireAnyScope("admin:all", "server:write"), h.serverController.UpdateServer)
		servers.POST("/import", h.authMiddleware.RequireAnyScope("admin:all", "server:import"), h.serverController.ImportServers)

		// Delete operations - requires server:delete scope
		servers.DELETE("/:id", h.authMiddleware.RequireAnyScope("admin:all", "server:delete"), h.serverController.DeleteServer)
	}

	// Protected report routes
	reports := v1.Group("/reports")
	reports.Use(h.authMiddleware.RequireAuth())
	{
		reports.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "report:write"), h.reportController.SendReportByDate)
		reports.POST("/daily", h.authMiddleware.RequireAnyScope("admin:all", "report:write"), h.reportController.SendReportDaily)
	}

	// Admin-only user management routes
	users := v1.Group("/users")
	users.Use(h.authMiddleware.RequireAuth())
	{
		users.GET("/", h.authMiddleware.RequireAnyScope("admin:all", "user:read"), h.userController.ListUsers)
		users.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "user:write"), h.userController.CreateUser)
		users.PUT("/:id", h.authMiddleware.RequireAnyScope("admin:all", "user:write"), h.userController.UpdateUser)
		users.DELETE("/:id", h.authMiddleware.RequireAnyScope("admin:all", "user:delete"), h.userController.DeleteUser)
		users.GET("/profile", h.authMiddleware.RequireAnyScope("admin:all", "profile:read"), h.userController.GetProfile)
		users.PUT("/profile", h.authMiddleware.RequireAnyScope("admin:all", "profile:write"), h.userController.UpdateProfile)
		users.POST("/change-password", h.authMiddleware.RequireAnyScope("admin:all", "profile:write"), h.userController.ChangePassword)
	}

	// Admin-only job monitoring routes (read-only for monitoring background jobs)
	jobs := v1.Group("/jobs")
	jobs.Use(h.authMiddleware.RequireAuth())
	{
		// Only monitoring endpoints - jobs run automatically in background
		jobs.GET("/", h.authMiddleware.RequireAnyScope("admin:all", "jobs:read"), h.jobsController.GetJobs)
		jobs.GET("/status", h.authMiddleware.RequireAnyScope("admin:all", "jobs:read"), h.jobsController.GetJobStatus)
	}

	return router
}
