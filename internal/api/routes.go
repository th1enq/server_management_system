package api

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/wire"
)

func SetupRoutes(router *gin.Engine, app *wire.App) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	v1 := router.Group("/api/v1")

	servers := v1.Group("/servers")
	{
		servers.POST("/", app.ServerHandler.CreateServer)
		servers.PUT("/:id", app.ServerHandler.UpdateServer)
		servers.DELETE("/:id", app.ServerHandler.DeleteServer)
		servers.GET("/", app.ServerHandler.ListServer)
		servers.POST("/import", app.ServerHandler.ImportServers)
		servers.GET("/export", app.ServerHandler.ExportServers)
	}

	reports := v1.Group("/reports")
	{
		reports.POST("/", app.ReportHandler.SendReportByDate)
		reports.POST("/daily", app.ReportHandler.SendReportDaily)
	}
}
