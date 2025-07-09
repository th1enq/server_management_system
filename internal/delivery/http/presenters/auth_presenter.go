package presenters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
)

type AuthPresenter interface {
	LoginSuccess(c *gin.Context, authResponse *dto.AuthResponse)
	RegisterSuccess(c *gin.Context, authResponse *dto.AuthResponse)
	RefreshTokenSuccess(c *gin.Context, authResponse *dto.AuthResponse)
	LogoutSuccess(c *gin.Context)

	// Error responses
	InvalidRequest(c *gin.Context, message string, err error)
	AuthenticationFailed(c *gin.Context, message string, err error)
	RegistrationFailed(c *gin.Context, message string, err error)
	InvalidRefreshToken(c *gin.Context, message string, err error)
	Unauthorized(c *gin.Context, message string)
	InternalServerError(c *gin.Context, message string, err error)
}

type authPresenter struct{}

func NewAuthPresenter() AuthPresenter {
	return &authPresenter{}
}

func (p *authPresenter) LoginSuccess(c *gin.Context, authResponse *dto.AuthResponse) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Login successful",
		authResponse,
	)

	c.JSON(http.StatusOK, response)
}

func (p *authPresenter) RegisterSuccess(c *gin.Context, authResponse *dto.AuthResponse) {
	response := domain.NewSuccessResponse(
		domain.CodeCreated,
		"Registration successful",
		authResponse,
	)

	c.JSON(http.StatusCreated, response)
}

func (p *authPresenter) RefreshTokenSuccess(c *gin.Context, authResponse *dto.AuthResponse) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Token refreshed successfully",
		authResponse,
	)

	c.JSON(http.StatusOK, response)
}

func (p *authPresenter) LogoutSuccess(c *gin.Context) {
	response := domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Logout successful",
		nil,
	)

	c.JSON(http.StatusOK, response)
}

func (p *authPresenter) InvalidRequest(c *gin.Context, message string, err error) {
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

func (p *authPresenter) AuthenticationFailed(c *gin.Context, message string, err error) {
	var errorMsg interface{}
	if err != nil {
		errorMsg = err.Error()
	}

	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		errorMsg,
	)

	c.JSON(http.StatusUnauthorized, response)
}

func (p *authPresenter) RegistrationFailed(c *gin.Context, message string, err error) {
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

func (p *authPresenter) InvalidRefreshToken(c *gin.Context, message string, err error) {
	var errorMsg interface{}
	if err != nil {
		errorMsg = err.Error()
	}

	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		errorMsg,
	)

	c.JSON(http.StatusUnauthorized, response)
}

func (p *authPresenter) Unauthorized(c *gin.Context, message string) {
	response := domain.NewErrorResponse(
		domain.CodeUnauthorized,
		message,
		nil,
	)

	c.JSON(http.StatusUnauthorized, response)
}

func (p *authPresenter) InternalServerError(c *gin.Context, message string, err error) {
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
