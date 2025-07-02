package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/th1enq/server_management_system/internal/handler"
	"github.com/th1enq/server_management_system/internal/middleware"
)

type Handler struct {
	serverHandler  *handler.ServerHandler
	reportHandler  *handler.ReportHandler
	authHandler    *handler.AuthHandler
	userHandler    *handler.UserHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewHandler(serverHandler *handler.ServerHandler, reportHandler *handler.ReportHandler, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware) *Handler {
	return &Handler{
		serverHandler:  serverHandler,
		reportHandler:  reportHandler,
		authHandler:    authHandler,
		userHandler:    userHandler,
		authMiddleware: authMiddleware,
	}
}

func (h *Handler) RegisterRoutes() *gin.Engine {
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
		auth.POST("/login", h.authHandler.Login)
		auth.POST("/register", h.authHandler.Register)
		auth.POST("/refresh", h.authHandler.RefreshToken)
	}

	// Protected auth routes
	authProtected := v1.Group("/auth")
	authProtected.Use(h.authMiddleware.RequireAuth())
	{
		authProtected.POST("/logout", h.authHandler.Logout)
	}

	// Protected server routes
	servers := v1.Group("/servers")
	servers.Use(h.authMiddleware.RequireAuth())
	{
		// Read operations - requires server:read scope
		servers.GET("/", h.authMiddleware.RequireScope("server:read"), h.serverHandler.ListServer)
		servers.GET("/export", h.authMiddleware.RequireScope("server:export"), h.serverHandler.ExportServers)

		// Write operations - requires server:write scope
		servers.POST("/", h.authMiddleware.RequireScope("server:write"), h.serverHandler.CreateServer)
		servers.PUT("/:id", h.authMiddleware.RequireScope("server:write"), h.serverHandler.UpdateServer)
		servers.POST("/import", h.authMiddleware.RequireScope("server:import"), h.serverHandler.ImportServers)

		// Delete operations - requires server:delete scope
		servers.DELETE("/:id", h.authMiddleware.RequireScope("server:delete"), h.serverHandler.DeleteServer)
	}

	// Protected report routes
	reports := v1.Group("/reports")
	reports.Use(h.authMiddleware.RequireAuth())
	{
		reports.POST("/", h.authMiddleware.RequireScope("report:write"), h.reportHandler.SendReportByDate)
		reports.POST("/daily", h.authMiddleware.RequireScope("report:write"), h.reportHandler.SendReportDaily)
	}

	// Admin-only user management routes
	users := v1.Group("/users")
	users.Use(h.authMiddleware.RequireAuth())
	{
		users.GET("/", h.authMiddleware.RequireAnyScope("admin:all", "user:read"), h.userHandler.ListUsers)
		users.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "user:write"), h.userHandler.CreateUser)
		users.PUT("/:id", h.authMiddleware.RequireAnyScope("admin:all", "user:write"), h.userHandler.UpdateUser)
		users.DELETE("/:id", h.authMiddleware.RequireAnyScope("admin:all", "user:delete"), h.userHandler.DeleteUser)
		users.GET("/profile", h.authMiddleware.RequireAnyScope("profile:read"), h.userHandler.GetProfile)
		users.PUT("/profile", h.authMiddleware.RequireScope("profile:write"), h.userHandler.UpdateProfile)
		users.POST("/change-password", h.authMiddleware.RequireScope("profile:write"), h.userHandler.ChangePassword)
	}

	return router
}
