package presenters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
)

type ReportPresenter interface {
	// Success responses
	DailyReportSent(c *gin.Context)
	CustomReportSent(c *gin.Context)
	ReportGenerated(c *gin.Context, filePath string)

	// Error responses
	InvalidRequest(c *gin.Context, message string, err error)
	ValidationError(c *gin.Context, message string, err error)
	Unauthorized(c *gin.Context, message string)
	InternalServerError(c *gin.Context, message string, err error)
}

type reportPresenter struct{}

func NewReportPresenter() ReportPresenter {
	return &reportPresenter{}
}

func (p *reportPresenter) DailyReportSent(c *gin.Context) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Daily report sent successfully",
		nil,
	)
	c.JSON(http.StatusOK, response)
}

func (p *reportPresenter) CustomReportSent(c *gin.Context) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Custom report sent successfully",
		nil,
	)
	c.JSON(http.StatusOK, response)
}

func (p *reportPresenter) ReportGenerated(c *gin.Context, filePath string) {
	data := map[string]interface{}{
		"file_path": filePath,
	}
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Report generated successfully",
		data,
	)
	c.JSON(http.StatusOK, response)
}

func (p *reportPresenter) InvalidRequest(c *gin.Context, message string, err error) {
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

func (p *reportPresenter) ValidationError(c *gin.Context, message string, err error) {
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

func (p *reportPresenter) Unauthorized(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		nil,
	)
	c.JSON(http.StatusUnauthorized, response)
}

func (p *reportPresenter) InternalServerError(c *gin.Context, message string, err error) {
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
