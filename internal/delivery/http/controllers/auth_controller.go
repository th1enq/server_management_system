package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type AuthController struct {
	authUseCase   usecases.AuthUseCase
	authPresenter presenters.AuthPresenter
	logger        *zap.Logger
}

func NewAuthController(
	authUseCase usecases.AuthUseCase,
	authPresenter presenters.AuthPresenter,
	logger *zap.Logger,
) *AuthController {
	return &AuthController{
		authUseCase:   authUseCase,
		authPresenter: authPresenter,
		logger:        logger,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} domain.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid login request", zap.Error(err))
		ac.authPresenter.InvalidRequest(c, "Invalid request data", err)
		return
	}

	response, err := ac.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		ac.logger.Error("Authentication failed", zap.String("username", req.Username), zap.Error(err))
		ac.authPresenter.AuthenticationFailed(c, "Authentication failed", err)
		return
	}

	ac.logger.Info("User logged in successfully", zap.String("username", req.Username))
	ac.authPresenter.LoginSuccess(c, response)
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} domain.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 409 {object} domain.APIResponse
// @Router /api/v1/auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid registration request", zap.Error(err))
		ac.authPresenter.InvalidRequest(c, "Invalid request data", err)
		return
	}
	response, err := ac.authUseCase.Register(c.Request.Context(), req)
	if err != nil {
		ac.logger.Error("Registration failed", zap.String("username", req.Username), zap.Error(err))
		ac.authPresenter.RegistrationFailed(c, "Registration failed", err)
		return
	}

	ac.logger.Info("User registered successfully", zap.String("username", req.Username))
	ac.authPresenter.RegisterSuccess(c, response)
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} domain.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid refresh token request", zap.Error(err))
		ac.authPresenter.InvalidRequest(c, "Invalid request data", err)
		return
	}

	response, err := ac.authUseCase.RefreshToken(c.Request.Context(), req)
	if err != nil {
		ac.logger.Error("Failed to refresh token", zap.String("refresh_token", req.RefreshToken), zap.Error(err))
		ac.authPresenter.InvalidRefreshToken(c, "Invalid refresh token", err)
		return
	}

	ac.logger.Info("Token refreshed successfully", zap.String("refresh_token", "[REDACTED]"))
	ac.authPresenter.RefreshTokenSuccess(c, response)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user (invalidate session)
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		ac.logger.Warn("Logout attempt without user ID")
		ac.authPresenter.Unauthorized(c, "Unauthorized")
		return
	}

	err := ac.authUseCase.Logout(c.Request.Context(), userID)
	if err != nil {
		ac.logger.Error("Failed to logout user", zap.Uint("user_id", userID), zap.Error(err))
		ac.authPresenter.InternalServerError(c, "Failed to logout", err)
		return
	}

	ac.logger.Info("User logged out successfully", zap.Uint("user_id", userID))
	ac.authPresenter.LogoutSuccess(c)
}
