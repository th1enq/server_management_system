package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type ServerHandler struct {
	serverSrv services.ServerService
	logger    *zap.Logger
}

func NewServerHandler(serverSrv services.ServerService, logger *zap.Logger) *ServerHandler {
	return &ServerHandler{
		serverSrv: serverSrv,
		logger:    logger,
	}
}

// CreateServer godoc
// @Summary Create a new server
// @Description Create a new server with the provided information
// @Tags servers
// @Accept json
// @Produce json
// @Param server body models.Server true "Server information"
// @Success 201 {object} models.APIResponse{data=models.Server}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers [post]
func (h *ServerHandler) CreateServer(c *gin.Context) {
	h.logger.Info("Starting create server request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	var server models.Server
	if err := c.ShouldBindBodyWithJSON(&server); err != nil {
		h.logger.Error("Failed to bind request body",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.logger.Info("Creating server",
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
		zap.String("request_id", c.GetString("request_id")))

	err := h.serverSrv.CreateServer(c.Request.Context(), &server)
	if err != nil {
		h.logger.Error("Failed to create server",
			zap.Error(err),
			zap.String("server_id", server.ServerID),
			zap.String("server_name", server.ServerName),
			zap.String("request_id", c.GetString("request_id")))

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

	h.logger.Info("Server created successfully",
		zap.String("server_id", server.ServerID),
		zap.String("server_name", server.ServerName),
		zap.Uint("id", server.ID),
		zap.String("request_id", c.GetString("request_id")))

	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		models.CodeCreated,
		"Server created successfully",
		server,
	))
}

// ListServers godoc
// @Summary List servers
// @Description Get list of servers with optional filters and pagination
// @Tags servers
// @Accept json
// @Produce json
// @Param server_id query string false "Filter by server ID"
// @Param server_name query string false "Filter by server name"
// @Param status query string false "Filter by status"
// @Param ipv4 query string false "Filter by IPv4"
// @Param location query string false "Filter by location"
// @Param os query string false "Filter by OS"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort query string false "Sort field" default(created_time)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {object} models.APIResponse{data=dto.ServerListResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers [get]
func (h *ServerHandler) ListServer(c *gin.Context) {
	h.logger.Info("Starting list servers request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	var filter dto.ServerFilter
	var pagination dto.Pagination

	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Failed to bind filter parameters",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid filter parameters",
			err.Error(),
		))
		return
	}

	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.logger.Error("Failed to bind pagination parameters",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	h.logger.Info("Listing servers with filters",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
		zap.String("request_id", c.GetString("request_id")))

	response, err := h.serverSrv.ListServers(c.Request.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to list servers",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to list servers",
			err.Error(),
		))
		return
	}

	h.logger.Info("Servers listed successfully",
		zap.Int64("total", response.Total),
		zap.Int("returned_count", len(response.Servers)),
		zap.Int("page", response.Page),
		zap.Int("size", response.Size),
		zap.String("request_id", c.GetString("request_id")))

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Servers listed successfully",
		response,
	))
}

// UpdateServer godoc
// @Summary Update server
// @Description Update server information
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Param updateInfo body dto.ServerUpdate true "Server update information"
// @Success 200 {object} models.APIResponse{data=models.Server}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/{id} [put]
func (h *ServerHandler) UpdateServer(c *gin.Context) {
	h.logger.Info("Starting update server request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Failed to parse server ID",
			zap.Error(err),
			zap.String("id_param", c.Param("id")),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid server ID",
			err.Error(),
		))
		return
	}

	var updateInfo dto.ServerUpdate
	if err := c.ShouldBindBodyWithJSON(&updateInfo); err != nil {
		h.logger.Error("Failed to bind update request body",
			zap.Error(err),
			zap.Uint64("server_id", id),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.logger.Info("Updating server",
		zap.Uint64("server_id", id),
		zap.Any("update_info", updateInfo),
		zap.String("request_id", c.GetString("request_id")))

	server, err := h.serverSrv.UpdateServer(c.Request.Context(), uint(id), updateInfo)
	if err != nil {
		h.logger.Error("Failed to update server",
			zap.Error(err),
			zap.Uint64("server_id", id),
			zap.Any("update_info", updateInfo),
			zap.String("request_id", c.GetString("request_id")))

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

	h.logger.Info("Server updated successfully",
		zap.Uint64("server_id", id),
		zap.String("server_name", server.ServerName),
		zap.String("request_id", c.GetString("request_id")))

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Server update successfully",
		server,
	))
}

// DeleteServer godoc
// @Summary Delete server
// @Description Delete a server by ID
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/{id} [delete]
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	h.logger.Info("Starting delete server request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Failed to parse server ID for deletion",
			zap.Error(err),
			zap.String("id_param", c.Param("id")),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid server ID",
			err.Error(),
		))
		return
	}

	h.logger.Info("Deleting server",
		zap.Uint64("server_id", id),
		zap.String("request_id", c.GetString("request_id")))

	err = h.serverSrv.DeleteServer(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete server",
			zap.Error(err),
			zap.Uint64("server_id", id),
			zap.String("request_id", c.GetString("request_id")))

		status := http.StatusInternalServerError
		code := models.CodeInternalServerError

		if err.Error() == "server not found" {
			status = http.StatusNotFound
			code = models.CodeNotFound
		}

		c.JSON(status, models.NewErrorResponse(code, "Failed to delete server", err.Error()))
		return
	}

	h.logger.Info("Server deleted successfully",
		zap.Uint64("server_id", id),
		zap.String("request_id", c.GetString("request_id")))

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeDeleted,
		"Server deleted successfully",
		nil,
	))
}

// ImportServers godoc
// @Summary Import servers from Excel file
// @Description Import multiple servers from an Excel file
// @Tags servers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file"
// @Success 200 {object} models.APIResponse{data=dto.ImportResult}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/import [post]
func (h *ServerHandler) ImportServers(c *gin.Context) {
	h.logger.Info("Starting import servers request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	file, err := c.FormFile("file")
	fmt.Println(file)
	if err != nil {
		h.logger.Error("No file uploaded for import",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"No file uploaded",
			err.Error(),
		))
		return
	}

	h.logger.Info("Processing import file",
		zap.String("filename", file.Filename),
		zap.Int64("file_size", file.Size),
		zap.String("request_id", c.GetString("request_id")))

	filePath := "/tmp/" + uuid.New().String() + "_" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		h.logger.Error("Failed to save uploaded file",
			zap.Error(err),
			zap.String("filename", file.Filename),
			zap.String("file_path", filePath),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to save file",
			err.Error(),
		))
		return
	}

	h.logger.Info("File saved, starting import process",
		zap.String("file_path", filePath),
		zap.String("request_id", c.GetString("request_id")))

	result, err := h.serverSrv.ImportServers(c.Request.Context(), filePath)
	if err != nil {
		h.logger.Error("Failed to import servers",
			zap.Error(err),
			zap.String("file_path", filePath),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to import servers",
			err.Error(),
		))
		return
	}

	h.logger.Info("Servers imported successfully",
		zap.Int("success_count", result.SuccessCount),
		zap.Int("failure_count", result.FailureCount),
		zap.Strings("success_servers", result.SuccessServers),
		zap.Strings("failure_servers", result.FailureServers),
		zap.String("request_id", c.GetString("request_id")))

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Servers imported successfully",
		result,
	))
}

// ExportServers godoc
// @Summary Export servers to Excel file
// @Description Export servers to an Excel file with optional filters
// @Tags servers
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param server_id query string false "Filter by server ID"
// @Param server_name query string false "Filter by server name"
// @Param status query string false "Filter by status"
// @Param ipv4 query string false "Filter by IPv4"
// @Param location query string false "Filter by location"
// @Param os query string false "Filter by OS"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10000)
// @Param sort query string false "Sort field" default(created_time)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {file} binary
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/export [get]
func (h *ServerHandler) ExportServers(c *gin.Context) {
	h.logger.Info("Starting export servers request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	var filter dto.ServerFilter
	var pagination dto.Pagination

	// Bind query parameters
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Failed to bind filter parameters for export",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid filter parameters",
			err.Error(),
		))
		return
	}

	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.logger.Error("Failed to bind pagination parameters for export",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	h.logger.Info("Exporting servers with filters",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
		zap.String("request_id", c.GetString("request_id")))

	filePath, err := h.serverSrv.ExportServers(c.Request.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to export servers",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to export servers",
			err.Error(),
		))
		return
	}

	h.logger.Info("Servers exported successfully",
		zap.String("export_file_path", filePath),
		zap.String("request_id", c.GetString("request_id")))

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=servers.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.File(filePath)
}
