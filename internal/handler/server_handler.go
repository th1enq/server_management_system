package handler

import (
	"net/http"
	"strconv"

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

	err := h.serverSrv.CreateServer(c.Request.Context(), &server)
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

func (h *ServerHandler) ListServer(c *gin.Context) {

}

func (h *ServerHandler) UpdateServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid server ID",
			err.Error(),
		))
		return
	}
	var updateInfo map[string]interface{}
	if err := c.ShouldBindBodyWithJSON(&updateInfo); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
	}
	server, err := h.serverSrv.UpdateServer(c.Request.Context(), uint(id), updateInfo)
	if err != nil {
		logger.Error("Failed to update server", err, zap.Uint("id", uint(id)))
		code := models.CodeInternalServerError
		status := http.StatusInternalServerError

		if err.Error() == "server not found" {
			code = models.CodeNotFound
			status = http.StatusNotFound
		} else if err.Error() == "server with name is already exists" {
			code = models.CodeConflict
			status = http.StatusConflict
		}

		c.JSON(status, models.NewErrorResponse(code, "Failed to update server", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		models.CodeUpdated,
		"Server update successfully",
		server,
	))
}

func (h *ServerHandler) DeleteServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid server ID",
			err.Error(),
		))
		return
	}
	err = h.serverSrv.DeleteServer(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to delete server", err, zap.Uint("id", uint(id)))

		status := http.StatusInternalServerError
		code := models.CodeInternalServerError

		if err.Error() == "server not found" {
			status = http.StatusNotFound
			code = models.CodeNotFound
		}

		c.JSON(status, models.NewErrorResponse(code, "Failed to delete server", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeDeleted,
		"Server deleted successfully",
		nil,
	))
}
