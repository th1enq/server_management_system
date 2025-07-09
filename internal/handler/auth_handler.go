package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/interfaces/middleware"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService services.IAuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService services.IAuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} models.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	response, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Error("Authentication failed", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.CodeUnauthorized,
			"Authentication failed",
			err.Error(),
		))
		return
	}

	h.logger.Info("User logged in successfully", zap.String("username", req.Username))
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Login successful",
		response))
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} models.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	registerRequest := dto.CreateUserRequest{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	response, err := h.authService.Register(c.Request.Context(), registerRequest)
	if err != nil {
		h.logger.Error("Registration failed", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.CodeConflict,
			"Registration failed",
			err.Error(),
		))
		return
	}

	h.logger.Info("User registered successfully", zap.String("username", req.Username))
	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		models.CodeCreated,
		"Registration successful",
		response,
	))
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("Failed to refresh token", zap.String("refresh_token", req.RefreshToken), zap.Error(err))
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.CodeUnauthorized,
			"Invalid refresh token",
			err.Error(),
		))
		return
	}

	h.logger.Info("Token refreshed successfully", zap.String("refresh_token", "[REDACTED]"))
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Token refreshed successfully",
		response,
	))
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user (invalidate session)
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Warn("Logout attempt without user ID")
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.CodeUnauthorized,
			"Unauthorized",
			nil,
		))
		return
	}

	err := h.authService.Logout(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to logout user", zap.Uint("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to logout",
			err.Error(),
		))
		return
	}

	h.logger.Info("User logged out successfully", zap.Uint("user_id", userID))
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Logout successful",
		nil,
	))
}
