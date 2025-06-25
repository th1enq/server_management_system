package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	handler "github.com/th1enq/server_management_system/internal/controller"
)

type Handler struct {
	serverHandler *handler.ServerHandler
	reportHandler *handler.ReportHandler
}

func NewHandler(serverHandler *handler.ServerHandler, reportHandler *handler.ReportHandler) *Handler {
	return &Handler{
		serverHandler: serverHandler,
		reportHandler: reportHandler,
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

	servers := v1.Group("/servers")
	{
		servers.POST("/", h.serverHandler.CreateServer)
		servers.PUT("/:id", h.serverHandler.UpdateServer)
		servers.DELETE("/:id", h.serverHandler.DeleteServer)
		servers.GET("/", h.serverHandler.ListServer)
		servers.POST("/import", h.serverHandler.ImportServers)
		servers.GET("/export", h.serverHandler.ExportServers)
	}

	reports := v1.Group("/reports")
	{
		reports.POST("/", h.reportHandler.SendReportByDate)
		reports.POST("/daily", h.reportHandler.SendReportDaily)
	}

	return router
}
