package presenters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
)

type UserPresenter interface {
	// Success responses
	ProfileRetrieved(c *gin.Context, user *domain.User)
	ProfileUpdated(c *gin.Context, user *domain.User)
	PasswordChanged(c *gin.Context)
	UsersRetrieved(c *gin.Context, users []domain.User)
	UserCreated(c *gin.Context, user *domain.User)
	UserUpdated(c *gin.Context, user *domain.User)
	UserDeleted(c *gin.Context)

	// Error responses
	InvalidRequest(c *gin.Context, message string, err error)
	UserNotFound(c *gin.Context, message string)
	ValidationError(c *gin.Context, message string, err error)
	Unauthorized(c *gin.Context, message string)
	Forbidden(c *gin.Context, message string)
	ConflictError(c *gin.Context, message string, err error)
	InternalServerError(c *gin.Context, message string, err error)
}

type userPresenter struct{}

func NewUserPresenter() UserPresenter {
	return &userPresenter{}
}

func (p *userPresenter) ProfileRetrieved(c *gin.Context, user *domain.User) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Profile retrieved successfully",
		user,
	)
	c.JSON(http.StatusOK, response)
}

func (p *userPresenter) ProfileUpdated(c *gin.Context, user *domain.User) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Profile updated successfully",
		user,
	)
	c.JSON(http.StatusOK, response)
}

func (p *userPresenter) PasswordChanged(c *gin.Context) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Password changed successfully",
		nil,
	)
	c.JSON(http.StatusOK, response)
}

func (p *userPresenter) UsersRetrieved(c *gin.Context, users []domain.User) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Users retrieved successfully",
		users,
	)
	c.JSON(http.StatusOK, response)
}

func (p *userPresenter) UserCreated(c *gin.Context, user *domain.User) {
	response := domain.NewSuccessResponse(
		domain.CodeCreated,
		"User created successfully",
		user,
	)
	c.JSON(http.StatusCreated, response)
}

func (p *userPresenter) UserUpdated(c *gin.Context, user *domain.User) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"User updated successfully",
		user,
	)
	c.JSON(http.StatusOK, response)
}

func (p *userPresenter) UserDeleted(c *gin.Context) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"User deleted successfully",
		nil,
	)
	c.JSON(http.StatusOK, response)
}

func (p *userPresenter) InvalidRequest(c *gin.Context, message string, err error) {
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

func (p *userPresenter) UserNotFound(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeNotFound,
		message,
		nil,
	)
	c.JSON(http.StatusNotFound, response)
}

func (p *userPresenter) ValidationError(c *gin.Context, message string, err error) {
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

func (p *userPresenter) Unauthorized(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		nil,
	)
	c.JSON(http.StatusUnauthorized, response)
}

func (p *userPresenter) Forbidden(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeForbidden,
		message,
		nil,
	)
	c.JSON(http.StatusForbidden, response)
}

func (p *userPresenter) ConflictError(c *gin.Context, message string, err error) {
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

func (p *userPresenter) InternalServerError(c *gin.Context, message string, err error) {
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
