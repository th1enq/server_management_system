package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/services"
)

type ServerHandler struct {
	serverSrv services.ServerService
}

func NewServerHandler(serverSrv services.ServerService) *ServerHandler {
	return &ServerHandler{
		serverSrv: serverSrv,
	}
}

func (h *ServerHandler) CreateServer(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "test done",
	})
}
