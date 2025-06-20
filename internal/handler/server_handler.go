package handler

import (
	"fmt"
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

// @Summary Create a new server
// @Description Create a new server with the provided details
// @Tags servers
// @Accept json
// @Produce json
// @Param server body models.Server true "Server details"
// @Success 200 {object} models.ImportResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/servers/import [post]
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

func (h *ServerHandler) Monitors(c *gin.Context) {
	servers, err := h.serverSrv.GetServersIP(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to get ipv4 address",
			err.Error(),
		))
		return
	}

	ipInfo := map[string]interface{}{
		"type":     "tcp",
		"schedule": "@every 10s",
		"hosts":    servers,
		"name":     "vcs_sms_servers",
	}

	c.JSON(http.StatusOK, []interface{}{ipInfo})
}

func (h *ServerHandler) ListServer(c *gin.Context) {
	var filter models.ServerFilter
	var pagination models.Pagination

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid filter parameters",
			err.Error(),
		))
		return
	}

	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid pagnination parameters",
			err.Error(),
		))
		return
	}

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 10
	}

	response, err := h.serverSrv.ListServers(c.Request.Context(), filter, pagination)
	if err != nil {
		logger.Error("Failed to list servers", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to list servers",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Servers listed successfully",
		response,
	))
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

func (h *ServerHandler) ImportServers(c *gin.Context) {
	file, err := c.FormFile("file")
	fmt.Println(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"No file uploaded",
			err.Error(),
		))
		return
	}
	filePath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to save file",
			err.Error(),
		))
		return
	}
	result, err := h.serverSrv.ImportServers(c.Request.Context(), filePath)
	if err != nil {
		logger.Error("Failed to import servers", err, zap.String("file", file.Filename))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to import servers",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Servers imported successfully",
		result,
	))
}

func (h *ServerHandler) ExportServers(c *gin.Context) {
	var filter models.ServerFilter
	var pagination models.Pagination

	// Bind query parameters
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid filter parameters",
			err.Error(),
		))
		return
	}

	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	// Set default page size for export
	if pagination.PageSize == 0 {
		pagination.PageSize = 10000
	}

	filePath, err := h.serverSrv.ExportServers(c.Request.Context(), filter, pagination)
	if err != nil {
		logger.Error("Failed to export servers", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to export servers",
			err.Error(),
		))
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=servers.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.File(filePath)
}
