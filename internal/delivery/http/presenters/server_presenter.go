package presenters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/dto"
)

type ServerPresenter interface {
	// Success responses
	ServerCreated(c *gin.Context, response dto.CreateServerRequest)
	ServerUpdated(c *gin.Context, server dto.ServerResponse)
	ServerRetrieved(c *gin.Context, server dto.ServerResponse)
	ServersRetrieved(c *gin.Context, response []*entity.Server)
	ServerDeleted(c *gin.Context)
	ServerStatusUpdated(c *gin.Context, message string)
	ExportCompleted(c *gin.Context, filePath string)

	// Error responses
	InvalidRequest(c *gin.Context, message string, err error)
	ServerNotFound(c *gin.Context, message string)
	ValidationError(c *gin.Context, message string, err error)
	ConflictError(c *gin.Context, message string, err error)
	Unauthorized(c *gin.Context, message string)
	InternalServerError(c *gin.Context, message string, err error)
}

type serverPresenter struct{}

func NewServerPresenter() ServerPresenter {
	return &serverPresenter{}
}

func (p *serverPresenter) ServerCreated(c *gin.Context, res dto.CreateServerRequest) {
	response := domain.NewSuccessResponse(
		domain.CodeCreated,
		"Server created successfully",
		res,
	)
	c.JSON(http.StatusCreated, response)
}

func (p *serverPresenter) ServerUpdated(c *gin.Context, server dto.ServerResponse) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Server updated successfully",
		server,
	)
	c.JSON(http.StatusOK, response)
}

func (p *serverPresenter) ServerRetrieved(c *gin.Context, server dto.ServerResponse) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Server retrieved successfully",
		server,
	)
	c.JSON(http.StatusOK, response)
}

func (p *serverPresenter) ServersRetrieved(c *gin.Context, response []*entity.Server) {
	response_data := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Servers retrieved successfully",
		response,
	)
	c.JSON(http.StatusOK, response_data)
}

func (p *serverPresenter) ServerDeleted(c *gin.Context) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Server deleted successfully",
		nil,
	)
	c.JSON(http.StatusOK, response)
}

func (p *serverPresenter) ServerStatusUpdated(c *gin.Context, message string) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		message,
		nil,
	)
	c.JSON(http.StatusOK, response)
}

func (p *serverPresenter) ExportCompleted(c *gin.Context, filePath string) {
	data := map[string]interface{}{
		"file_path": filePath,
	}
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Export completed successfully",
		data,
	)
	c.JSON(http.StatusOK, response)
}

func (p *serverPresenter) InvalidRequest(c *gin.Context, message string, err error) {
	var errorMsg interface{}
	if err != nil {
		errorMsg = err.Error()
	}

	response := domain.NewErrorResponse(
		domain.CodeBadRequest,
		message,
		errorMsg,
	)
	c.JSON(http.StatusBadRequest, response)
}

func (p *serverPresenter) ServerNotFound(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeNotFound,
		message,
		nil,
	)
	c.JSON(http.StatusNotFound, response)
}

func (p *serverPresenter) ValidationError(c *gin.Context, message string, err error) {
	var errorMsg interface{}
	if err != nil {
		errorMsg = err.Error()
	}

	response := domain.NewErrorResponse(
		domain.CodeValidationError,
		message,
		errorMsg,
	)
	c.JSON(http.StatusBadRequest, response)
}

func (p *serverPresenter) ConflictError(c *gin.Context, message string, err error) {
	var errorMsg interface{}
	if err != nil {
		errorMsg = err.Error()
	}

	response := domain.NewErrorResponse(
		domain.CodeConflict,
		message,
		errorMsg,
	)
	c.JSON(http.StatusConflict, response)
}

func (p *serverPresenter) Unauthorized(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		nil,
	)
	c.JSON(http.StatusUnauthorized, response)
}

func (p *serverPresenter) InternalServerError(c *gin.Context, message string, err error) {
	var errorMsg interface{}
	if err != nil {
		errorMsg = err.Error()
	}

	response := domain.NewErrorResponse(
		domain.CodeInternalServerError,
		message,
		errorMsg,
	)
	c.JSON(http.StatusInternalServerError, response)
}
