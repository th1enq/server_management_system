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
	}
}
