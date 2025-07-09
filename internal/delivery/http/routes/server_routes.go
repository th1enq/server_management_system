package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/interfaces/http/controllers"
	"github.com/th1enq/server_management_system/internal/middleware"
)

// ServerRoutes handles server management routing
type ServerRoutes struct {
	serverController *controllers.ServerController
	authMiddleware   *middleware.AuthMiddleware
}

// NewServerRoutes creates a new server routes handler
func NewServerRoutes(
	serverController *controllers.ServerController,
	authMiddleware *middleware.AuthMiddleware,
) *ServerRoutes {
	return &ServerRoutes{
		serverController: serverController,
		authMiddleware:   authMiddleware,
	}
}

// Setup configures server management routes
func (sr *ServerRoutes) Setup(rg *gin.RouterGroup) {
	// All server routes require authentication
	servers := rg.Group("/servers")
	servers.Use(sr.authMiddleware.RequireAuth())

	// Read operations - requires server:read scope
	sr.setupReadRoutes(servers)

	// Write operations - requires server:write scope
	sr.setupWriteRoutes(servers)

	// Delete operations - requires server:delete scope
	sr.setupDeleteRoutes(servers)

	// Import/Export operations - requires special scopes
	sr.setupImportExportRoutes(servers)
}

// setupReadRoutes configures read-only server routes
func (sr *ServerRoutes) setupReadRoutes(rg *gin.RouterGroup) {
	readScope := sr.authMiddleware.RequireAnyScope("admin:all", "server:read")

	rg.GET("/", readScope, sr.serverController.ListServers)
	rg.GET("/:id", readScope, sr.serverController.GetServerByID)
}

// setupWriteRoutes configures server creation and update routes
func (sr *ServerRoutes) setupWriteRoutes(rg *gin.RouterGroup) {
	writeScope := sr.authMiddleware.RequireAnyScope("admin:all", "server:write")

	rg.POST("/", writeScope, sr.serverController.CreateServer)
	rg.PUT("/:id", writeScope, sr.serverController.UpdateServer)
}

// setupDeleteRoutes configures server deletion routes
func (sr *ServerRoutes) setupDeleteRoutes(rg *gin.RouterGroup) {
	deleteScope := sr.authMiddleware.RequireAnyScope("admin:all", "server:delete")

	rg.DELETE("/:id", deleteScope, sr.serverController.DeleteServer)
}

// setupImportExportRoutes configures import/export routes
func (sr *ServerRoutes) setupImportExportRoutes(rg *gin.RouterGroup) {
	exportScope := sr.authMiddleware.RequireAnyScope("admin:all", "server:export")
	importScope := sr.authMiddleware.RequireAnyScope("admin:all", "server:import")

	rg.GET("/export", exportScope, sr.serverController.ExportServers)
	rg.POST("/import", importScope, sr.serverController.ImportServers)
}
