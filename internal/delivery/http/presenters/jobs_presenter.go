package presenters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/jobs/scheduler"
)

type JobsPresenter interface {
	// Success responses
	JobsRetrieved(c *gin.Context, tasks []scheduler.TaskInfo)
	JobStatusRetrieved(c *gin.Context, status map[string]interface{})

	// Error responses
	Unauthorized(c *gin.Context, message string)
	InternalServerError(c *gin.Context, message string, err error)
}

type jobsPresenter struct{}

func NewJobsPresenter() JobsPresenter {
	return &jobsPresenter{}
}

func (p *jobsPresenter) JobsRetrieved(c *gin.Context, tasks []scheduler.TaskInfo) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Jobs information retrieved successfully",
		tasks,
	)
	c.JSON(http.StatusOK, response)
}

func (p *jobsPresenter) JobStatusRetrieved(c *gin.Context, status map[string]interface{}) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Job scheduler status retrieved successfully",
		status,
	)
	c.JSON(http.StatusOK, response)
}

func (p *jobsPresenter) Unauthorized(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		nil,
	)
	c.JSON(http.StatusUnauthorized, response)
}

func (p *jobsPresenter) InternalServerError(c *gin.Context, message string, err error) {
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
