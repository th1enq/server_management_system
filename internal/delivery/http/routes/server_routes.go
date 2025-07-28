package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

type ServerRouter interface {
	RegisterRoutes(v1 *gin.RouterGroup)
}

type serverRouter struct {
	serverController *controllers.ServerController
	authMiddleware   *middleware.AuthMiddleware
}

func NewServerRouter(
	serverController *controllers.ServerController,
	authMiddleware *middleware.AuthMiddleware,
) ServerRouter {
	return &serverRouter{
		serverController: serverController,
		authMiddleware:   authMiddleware,
	}
}

func (h *serverRouter) RegisterRoutes(v1 *gin.RouterGroup) {
	servers := v1.Group("/servers")
	{
		servers.POST("/register", h.serverController.Register)
		servers.POST("/monitoring", h.serverController.Monitoring)
	}

	servers.Use(h.authMiddleware.RequireAuth())
	{

		servers.GET("/", h.authMiddleware.RequireAnyScope("admin:all", "server:read"), h.serverController.ListServer)
		servers.GET("/export", h.authMiddleware.RequireAnyScope("admin:all", "server:export"), h.serverController.ExportServers)

		servers.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "server:write"), h.serverController.CreateServer)
		servers.PUT("/:id", h.authMiddleware.RequireAnyScope("admin:all", "server:write"), h.serverController.UpdateServer)
		servers.POST("/import", h.authMiddleware.RequireAnyScope("admin:all", "server:import"), h.serverController.ImportServers)
		servers.DELETE("/:id", h.authMiddleware.RequireAnyScope("admin:all", "server:delete"), h.serverController.DeleteServer)
	}
}
