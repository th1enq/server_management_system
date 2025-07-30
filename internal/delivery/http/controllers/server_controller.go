package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type ServerController struct {
	serverUseCase   usecases.ServerUseCase
	serverPresenter presenters.ServerPresenter
	gatewayUseCase  usecases.GatewayUseCase
	logger          *zap.Logger
}

func NewServerController(
	serverUseCase usecases.ServerUseCase,
	serverPresenter presenters.ServerPresenter,
	gatewayUseCase usecases.GatewayUseCase,
	logger *zap.Logger,
) *ServerController {
	return &ServerController{
		serverUseCase:   serverUseCase,
		serverPresenter: serverPresenter,
		gatewayUseCase:  gatewayUseCase,
		logger:          logger,
	}
}

// Monitoring godoc
// @Summary Send server monitoring data
// @Description Send server monitoring data to the system
// @Tags servers
// @Accept json
// @Produce json
// @Param monitoring body dto.MetricsRequest true "Monitoring data"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/monitoring [post]
func (h *ServerController) Monitoring(c *gin.Context) {
	var req dto.MetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid monitoring request", zap.Error(err))
		h.serverPresenter.InvalidRequest(c, "Invalid request data", err)
		return
	}

	// Use API Gateway service instead of direct usecase call
	if err := h.gatewayUseCase.ProcessServerMetrics(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to process server metrics through API Gateway",
			zap.Error(err),
			zap.String("server_id", req.ServerID),
			zap.Int("cpu", req.CPU),
			zap.Int("ram", req.RAM),
			zap.Int("disk", req.Disk),
			zap.String("request_id", c.GetString("request_id")),
		)
		h.serverPresenter.InternalServerError(c, "Failed to process server metrics", err)
		return
	}

	h.serverPresenter.MonitoringSuccess(c, "Server metrics processed successfully")
}

// Register godoc
// @Summary Register server metrics
// @Description Register server metrics with the system
// @Tags servers
// @Accept json
// @Produce json
// @Param register body dto.RegisterMetricsRequest true "Register metrics request"
// @Success 201 {object} domain.APIResponse{data=dto.ServerResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 409 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/register [post]
func (h *ServerController) Register(c *gin.Context) {
	var req dto.RegisterMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid register metrics request", zap.Error(err))
		h.serverPresenter.InvalidRequest(c, "Invalid request data", err)
		return
	}

	IPv4 := c.ClientIP()

	reqCreate := dto.CreateServerRequest{
		ServerID:    req.ServerID,
		ServerName:  req.ServerName,
		IPv4:        IPv4,
		Description: req.Description,
		Location:    req.Location,
		OS:          req.OS,
	}

	fmt.Println(reqCreate)

	response, err := h.serverUseCase.Register(c.Request.Context(), reqCreate)
	if err != nil {
		h.logger.Error("Failed to create server",
			zap.Error(err),
			zap.String("server_id", req.ServerID),
			zap.String("server_name", req.ServerName),
			zap.String("request_id", c.GetString("request_id")))

		if err.Error() == "server_id and server_name are required" {
			h.serverPresenter.ValidationError(c, "Failed to create server", err)
		} else if err.Error() == "server is already exists" {
			h.serverPresenter.ConflictError(c, "Failed to create server", err)
		} else {
			h.serverPresenter.InternalServerError(c, "Failed to create server", err)
		}
		return
	}

	h.logger.Info("Server registered successfully",
		zap.String("server_id", req.ServerID),
		zap.String("server_name", req.ServerName),
		zap.String("request_id", c.GetString("request_id")))

	h.serverPresenter.ServerRegistered(c, response)
}

// CreateServer godoc
// @Summary Create a new server
// @Description Create a new server with the provided information
// @Tags servers
// @Accept json
// @Produce json
// @Param server body dto.CreateServerRequest true "Server information"
// @Success 201 {object} domain.APIResponse{data=dto.ServerResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 409 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers [post]
func (h *ServerController) CreateServer(c *gin.Context) {
	h.logger.Info("Starting create server request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	var req dto.CreateServerRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		h.logger.Warn("Failed to bind request body",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		h.serverPresenter.InvalidRequest(c, "Invalid request body", err)
		return
	}

	h.logger.Info("Creating server",
		zap.String("server_id", req.ServerID),
		zap.String("server_name", req.ServerName),
		zap.String("request_id", c.GetString("request_id")))

	_, err := h.serverUseCase.CreateServer(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create server",
			zap.Error(err),
			zap.String("server_id", req.ServerID),
			zap.String("server_name", req.ServerName),
		)

		if err.Error() == "server_id and server_name are required" {
			h.serverPresenter.ValidationError(c, "Failed to create server", err)
		} else if err.Error() == "server is already exists" {
			h.serverPresenter.ConflictError(c, "Failed to create server", err)
		} else {
			h.serverPresenter.InternalServerError(c, "Failed to create server", err)
		}
		return
	}

	h.logger.Info("Server created successfully",
		zap.String("server_id", req.ServerID),
		zap.String("server_name", req.ServerName),
		zap.String("request_id", c.GetString("request_id")))

	h.serverPresenter.ServerCreated(c, req)
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
// @Success 200 {object} domain.APIResponse{data=dto.ServerListResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers [get]
func (h *ServerController) ListServer(c *gin.Context) {
	h.logger.Info("Starting list servers request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	var filter dto.ServerFilter
	var pagination dto.Pagination

	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Failed to bind filter parameters",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		h.serverPresenter.InvalidRequest(c, "Invalid filter parameters", err)
		return
	}

	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.logger.Error("Failed to bind pagination parameters",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		h.serverPresenter.InvalidRequest(c, "Invalid pagination parameters", err)
		return
	}

	h.logger.Info("Listing servers with filters",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
		zap.String("request_id", c.GetString("request_id")))

	response, err := h.serverUseCase.ListServers(c.Request.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to list servers",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
			zap.String("request_id", c.GetString("request_id")))
		h.serverPresenter.InternalServerError(c, "Failed to list servers", err)
		return
	}

	h.logger.Info("Servers listed successfully")

	h.serverPresenter.ServersRetrieved(c, response)
}

// UpdateServer godoc
// @Summary Update server
// @Description Update server information
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Param updateInfo body dto.UpdateServerRequest true "Server update information"
// @Success 200 {object} domain.APIResponse{data=dto.ServerResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 409 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/{id} [put]
func (h *ServerController) UpdateServer(c *gin.Context) {
	h.logger.Info("Starting update server request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Failed to parse server ID",
			zap.Error(err),
			zap.String("id_param", c.Param("id")),
			zap.String("request_id", c.GetString("request_id")))
		h.serverPresenter.InvalidRequest(c, "Invalid server ID", err)
		return
	}

	var updateInfo dto.UpdateServerRequest
	if err := c.ShouldBindBodyWithJSON(&updateInfo); err != nil {
		h.logger.Error("Failed to bind update request body",
			zap.Error(err),
			zap.Uint64("server_id", id),
			zap.String("request_id", c.GetString("request_id")))
		h.serverPresenter.InvalidRequest(c, "Invalid request body", err)
		return
	}

	h.logger.Info("Updating server",
		zap.Uint64("server_id", id),
		zap.Any("update_info", updateInfo),
		zap.String("request_id", c.GetString("request_id")))

	server, err := h.serverUseCase.UpdateServer(c.Request.Context(), uint(id), updateInfo)
	if err != nil {
		h.logger.Error("Failed to update server",
			zap.Error(err),
			zap.Uint64("server_id", id),
			zap.Any("update_info", updateInfo),
			zap.String("request_id", c.GetString("request_id")))

		if err.Error() == "server not found" {
			h.serverPresenter.ServerNotFound(c, "Failed to update server")
		} else if err.Error() == "server with name is already exists" {
			h.serverPresenter.ConflictError(c, "Failed to update server", err)
		} else {
			h.serverPresenter.InternalServerError(c, "Failed to update server", err)
		}
		return
	}

	response := dto.ServerResponse{
		ServerID:    server.ServerID,
		ServerName:  server.ServerName,
		Status:      server.Status,
		IPv4:        server.IPv4,
		Location:    server.Location,
		OS:          server.OS,
		Description: server.Description,
	}

	h.logger.Info("Server updated successfully",
		zap.Uint64("server_id", id),
		zap.String("server_name", response.ServerName),
		zap.String("request_id", c.GetString("request_id")))

	h.serverPresenter.ServerUpdated(c, response)
}

// DeleteServer godoc
// @Summary Delete server
// @Description Delete a server by ID
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/{id} [delete]
func (h *ServerController) DeleteServer(c *gin.Context) {
	h.logger.Info("Starting delete server request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Failed to parse server ID for deletion",
			zap.Error(err),
			zap.String("id_param", c.Param("id")),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid server ID",
			err.Error(),
		))
		return
	}

	h.logger.Info("Deleting server",
		zap.Uint64("server_id", id),
		zap.String("request_id", c.GetString("request_id")))

	err = h.serverUseCase.DeleteServer(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete server",
			zap.Error(err),
			zap.Uint64("server_id", id),
			zap.String("request_id", c.GetString("request_id")))

		status := http.StatusInternalServerError
		code := domain.CodeInternalServerError

		if err.Error() == "server not found" {
			status = http.StatusNotFound
			code = domain.CodeNotFound
		}

		c.JSON(status, domain.NewErrorResponse(code, "Failed to delete server", err.Error()))
		return
	}

	h.logger.Info("Server deleted successfully",
		zap.Uint64("server_id", id),
		zap.String("request_id", c.GetString("request_id")))

	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeDeleted,
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
// @Success 200 {object} domain.APIResponse{data=dto.ImportResult}
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/import [post]
func (h *ServerController) ImportServers(c *gin.Context) {
	now := time.Now()
	h.logger.Info("Starting import servers request",
		zap.String("request_id", c.GetString("request_id")),
		zap.String("user_id", c.GetString("user_id")))

	file, err := c.FormFile("file")
	fmt.Println(file)
	if err != nil {
		h.logger.Error("No file uploaded for import",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
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
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.CodeInternalServerError,
			"Failed to save file",
			err.Error(),
		))
		return
	}

	h.logger.Info("File saved, starting import process",
		zap.String("file_path", filePath),
		zap.String("request_id", c.GetString("request_id")))

	result, err := h.serverUseCase.ImportServers(c.Request.Context(), filePath)
	if err != nil {
		h.logger.Error("Failed to import servers",
			zap.Error(err),
			zap.String("file_path", filePath),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.CodeInternalServerError,
			"Failed to import servers",
			err.Error(),
		))
		return
	}

	toltalDuration := time.Since(now)

	h.logger.Info("Servers imported successfully",
		zap.Int("success_count", result.SuccessCount),
		zap.Int("failure_count", result.FailureCount),
		zap.Strings("success_servers", result.SuccessServers),
		zap.Strings("failure_servers", result.FailureServers),
		zap.String("request_id", c.GetString("request_id")),
		zap.Duration("duration", toltalDuration))

	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeSuccess,
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
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/servers/export [get]
func (h *ServerController) ExportServers(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid filter parameters",
			err.Error(),
		))
		return
	}

	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.logger.Error("Failed to bind pagination parameters for export",
			zap.Error(err),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	h.logger.Info("Exporting servers with filters",
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
		zap.String("request_id", c.GetString("request_id")))

	filePath, err := h.serverUseCase.ExportServers(c.Request.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to export servers",
			zap.Error(err),
			zap.Any("filter", filter),
			zap.Any("pagination", pagination),
			zap.String("request_id", c.GetString("request_id")))
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.CodeInternalServerError,
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
