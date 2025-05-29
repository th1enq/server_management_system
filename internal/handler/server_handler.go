package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
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
	var server models.Server
	if err := c.ShouldBindBodyWithJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	err := h.serverSrv.CreateServer(c, &server)
	if err != nil {
		logger.Error("Failed to create server", err, zap.Any("server", server))

		code := models.CodeInternalServerError
		status := http.StatusInternalServerError

		if err.Error() == "server_id and server_name are required" {
			code = models.CodeValidationError
			status = http.StatusBadRequest
		} else if err.Error() == "server is already exists" {
			code = models.CodeConflict
			status = http.StatusConflict
		}

		c.JSON(status, models.NewErrorResponse(code, "Failed to create server", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		models.CodeCreated,
		"Server created successfully",
		server,
	))
}

func (h *ServerHandler) GetServer(c *gin.Context) {
	// id, err :=
}
