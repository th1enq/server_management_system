package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/domain/dto"
	"github.com/th1enq/server_management_system/internal/interfaces/http/presenters"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

// ServerController handles HTTP requests for server operations
type ServerController struct {
	serverUseCase usecases.IServerUseCase
	presenter     presenters.ServerPresenter
	logger        *zap.Logger
}

// NewServerController creates a new server controller
func NewServerController(
	serverUseCase usecases.IServerUseCase,
	presenter presenters.ServerPresenter,
	logger *zap.Logger,
) *ServerController {
	return &ServerController{
		serverUseCase: serverUseCase,
		presenter:     presenter,
		logger:        logger,
	}
}

// CreateServer godoc
// @Summary Create a new server
// @Description Create a new server with the provided information
// @Tags servers
// @Accept json
// @Produce json
// @Param server body entities.CreateServerRequest true "Server creation request"
// @Success 201 {object} presenters.ServerResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /api/v1/servers [post]
func (sc *ServerController) CreateServer(c *gin.Context) {
	var req dto.CreateServerRequest

	// 1. Parse and validate input (Controller responsibility)
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", zap.Error(err))
		sc.presenter.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 3. Call use case (Business logic)
	createdServer, err := sc.serverUseCase.CreateServer(c.Request.Context(), req)
	if err != nil {
		sc.logger.Error("Failed to create server", zap.Error(err))
		sc.presenter.Error(c, http.StatusInternalServerError, "Failed to create server", err)
		return
	}

	// 4. Present response (Controller + Presenter responsibility)
	sc.presenter.Success(c, http.StatusCreated, "Server created successfully", createdServer)
}

// ListServers godoc
// @Summary List all servers
// @Description Get a list of all servers with optional filtering
// @Tags servers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param environment query string false "Filter by environment"
// @Success 200 {object} presenters.ServerListResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /api/v1/servers [get]
func (sc *ServerController) ListServers(c *gin.Context) {
	// 1. Parse query parameters (Controller responsibility)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	sort := c.DefaultQuery("sort", "created_time")
	order := c.DefaultQuery("order", "desc")

	// Parse filter parameters
	serverID := c.Query("server_id")
	serverName := c.Query("server_name")
	status := c.Query("status")
	ipv4 := c.Query("ipv4")
	location := c.Query("location")
	os := c.Query("os")

	// 2. Create filter and pagination objects
	filter := dto.ServerFilter{
		ServerID:   serverID,
		ServerName: serverName,
		Status:     domain.ServerStatus(status),
		IPv4:       ipv4,
		Location:   location,
		OS:         os,
	}

	pagination := dto.Pagination{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}

	// 3. Call use case
	servers, total, err := sc.serverUseCase.ListServers(c.Request.Context(), filter, pagination)
	if err != nil {
		sc.logger.Error("Failed to list servers", zap.Error(err))
		sc.presenter.Error(c, http.StatusInternalServerError, "Failed to list servers", err)
		return
	}

	// 4. Present response
	sc.presenter.SuccessList(c, http.StatusOK, "Servers retrieved successfully", servers, total, page, pageSize)
}

// GetServerByID godoc
// @Summary Get server by ID
// @Description Get a specific server by its ID
// @Tags servers
// @Accept json
// @Produce json
// @Param id path string true "Server ID"
// @Success 200 {object} presenters.ServerResponse
// @Failure 404 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /api/v1/servers/{id} [get]
func (sc *ServerController) GetServerByID(c *gin.Context) {
	// 1. Extract and validate path parameter
	idStr := c.Param("id")
	if idStr == "" {
		sc.presenter.Error(c, http.StatusBadRequest, "Server ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		sc.presenter.Error(c, http.StatusBadRequest, "Invalid server ID", err)
		return
	}

	// 2. Call use case
	server, err := sc.serverUseCase.GetServerByID(c.Request.Context(), uint(id))
	if err != nil {
		sc.logger.Error("Failed to get server", zap.String("id", idStr), zap.Error(err))
		sc.presenter.Error(c, http.StatusNotFound, "Server not found", err)
		return
	}

	// 3. Present response
	sc.presenter.Success(c, http.StatusOK, "Server retrieved successfully", server)
}

// UpdateServer godoc
// @Summary Update server
// @Description Update an existing server
// @Tags servers
// @Accept json
// @Produce json
// @Param id path string true "Server ID"
// @Param server body entities.UpdateServerRequest true "Server update request"
// @Success 200 {object} presenters.ServerResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 404 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /api/v1/servers/{id} [put]
func (sc *ServerController) UpdateServer(c *gin.Context) {
	// 1. Extract and validate path parameter
	idStr := c.Param("id")
	if idStr == "" {
		sc.presenter.Error(c, http.StatusBadRequest, "Server ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		sc.presenter.Error(c, http.StatusBadRequest, "Invalid server ID", err)
		return
	}

	// 2. Parse and validate request body
	var req dto.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.logger.Error("Invalid request body", zap.Error(err))
		sc.presenter.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 3. Call use case
	updatedServer, err := sc.serverUseCase.UpdateServer(c.Request.Context(), uint(id), req)
	if err != nil {
		sc.logger.Error("Failed to update server", zap.String("id", idStr), zap.Error(err))
		sc.presenter.Error(c, http.StatusInternalServerError, "Failed to update server", err)
		return
	}

	// 4. Present response
	sc.presenter.Success(c, http.StatusOK, "Server updated successfully", updatedServer)
}

// DeleteServer godoc
// @Summary Delete server
// @Description Delete a server by ID
// @Tags servers
// @Accept json
// @Produce json
// @Param id path string true "Server ID"
// @Success 200 {object} presenters.MessageResponse
// @Failure 404 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /api/v1/servers/{id} [delete]
func (sc *ServerController) DeleteServer(c *gin.Context) {
	// 1. Extract and validate path parameter
	idStr := c.Param("id")
	if idStr == "" {
		sc.presenter.Error(c, http.StatusBadRequest, "Server ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		sc.presenter.Error(c, http.StatusBadRequest, "Invalid server ID", err)
		return
	}

	// 2. Call use case
	err = sc.serverUseCase.DeleteServer(c.Request.Context(), uint(id))
	if err != nil {
		sc.logger.Error("Failed to delete server", zap.String("id", idStr), zap.Error(err))
		sc.presenter.Error(c, http.StatusInternalServerError, "Failed to delete server", err)
		return
	}

	// 3. Present response
	sc.presenter.Message(c, http.StatusOK, "Server deleted successfully")
}
