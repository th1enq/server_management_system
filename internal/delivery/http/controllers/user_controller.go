package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type UserController struct {
	userUseCase   usecases.UserUseCase
	userPresenter presenters.UserPresenter
	log           *zap.Logger
}

func NewUserController(
	userUseCase usecases.UserUseCase,
	userPresenter presenters.UserPresenter,
	log *zap.Logger,
) *UserController {
	return &UserController{
		userUseCase:   userUseCase,
		userPresenter: userPresenter,
		log:           log,
	}
}

// GetProfile returns the current user's profile
// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.APIResponse{data=domain.User}
// @Failure 401 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Router /api/v1/users/profile [get]
func (h *UserController) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.log.Warn("GetProfile: user ID not found in context")
		h.userPresenter.Unauthorized(c, "Authentication required")
		return
	}

	h.log.Info("GetProfile: fetching user profile", zap.Uint("userID", userID))

	user, err := h.userUseCase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("GetProfile: user not found", zap.Uint("userID", userID), zap.Error(err))
		h.userPresenter.UserNotFound(c, "User not found")
		return
	}

	h.log.Info("GetProfile: user profile retrieved successfully", zap.Uint("userID", userID))
	h.userPresenter.ProfileRetrieved(c, user)
}

// UpdateProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ProfileUpdate true "Profile updates"
// @Success 200 {object} domain.APIResponse{data=domain.User}
// @Failure 400 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Router /api/v1/users/profile [put]
func (h *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.log.Warn("UpdateProfile: user ID not found in context")
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
			domain.CodeAuthError,
			"Authentication required",
			nil,
		))
		return
	}

	var updates dto.ProfileUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.log.Warn("UpdateProfile: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("UpdateProfile: updating user profile", zap.Uint("userID", userID), zap.Any("updates", updates))

	user, err := h.userUseCase.UpdateProfile(c.Request.Context(), userID, updates)
	if err != nil {
		h.log.Error("UpdateProfile: failed to update user profile", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Failed to update profile",
			err.Error(),
		))
		return
	}
	h.log.Info("UpdateProfile: user profile updated successfully", zap.Uint("userID", userID))
	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeSuccess,
		"User profile updated successfully",
		user,
	))
}

// ChangePassword allows the user to change their password
// @Summary Change password
// @Description Change the authenticated user's password
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.PasswordUpdate true "Password change data"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Router /api/v1/users/change-password [post]
func (h *UserController) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.log.Warn("ChangePassword: user ID not found in context")
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(
			domain.CodeAuthError,
			"Authentication required",
			nil,
		))
		return
	}

	var req dto.PasswordUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("ChangePassword: changing password for user", zap.Uint("userID", userID))

	if err := h.userUseCase.UpdatePassword(c.Request.Context(), userID, req); err != nil {
		h.log.Error("ChangePassword: failed to change password", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.CodeInternalServerError,
			"Failed to change password",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Password changed successfully",
		nil,
	))
}

// ListUsers returns a list of users (admin only)
// @Summary List users
// @Description Get a list of all users (admin only)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} domain.APIResponse{data=[]domain.User}
// @Failure 401 {object} domain.APIResponse
// @Failure 403 {object} domain.APIResponse
// @Router /api/v1/users [get]
func (h *UserController) ListUsers(c *gin.Context) {
	// Parse query parameters
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	h.log.Info("ListUsers: fetching users", zap.Int("limit", limit), zap.Int("offset", offset))

	users, err := h.userUseCase.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("ListUsers: failed to fetch users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.CodeInternalServerError,
			"Failed to fetch users",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeSuccess,
		"Users retrieved successfully",
		users,
	))
}

// CreateUser creates a new user (admin only)
// @Summary Create user
// @Description Create a new user (admin only)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User data"
// @Success 201 {object} domain.User
// @Failure 400 {object} domain.APIResponse{data=domain.User}
// @Failure 401 {object} domain.APIResponse
// @Failure 403 {object} domain.APIResponse
// @Router /api/v1/users [post]
func (h *UserController) CreateUser(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("CreateUser: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("CreateUser: creating user",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.String("firstName", req.FirstName),
		zap.String("lastName", req.LastName),
	)

	var createdUser *domain.User
	var err error

	if createdUser, err = h.userUseCase.CreateUser(c.Request.Context(), req); err != nil {
		h.log.Error("CreateUser: failed to create user", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Failed to create user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse(
		domain.CodeSuccess,
		"User created successfully",
		createdUser,
	))
}

// UpdateUser updates a user (admin only)
// @Summary Update user
// @Description Update a user (admin only)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body dto.UserUpdate true "User updates"
// @Success 200 {object} domain.APIResponse{data=domain.User}
// @Failure 400 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Failure 403 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Router /api/v1/users/{id} [put]
func (h *UserController) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.log.Warn("UpdateUser: invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid user ID",
			err.Error(),
		))
		return
	}

	var updates dto.UserUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("UpdateUser: updating user", zap.Uint("id", uint(id)), zap.Any("updates", updates))

	user, err := h.userUseCase.UpdateUser(c.Request.Context(), uint(id), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Failed to update user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeSuccess,
		"User updated successfully",
		user,
	))
}

// DeleteUser deletes a user (admin only)
// @Summary Delete user
// @Description Delete a user (admin only)
// @Tags users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 401 {object} domain.APIResponse
// @Failure 403 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserController) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.log.Warn("DeleteUser: invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Invalid user ID",
			err.Error(),
		))
		return
	}

	h.log.Info("DeleteUser: deleting user", zap.Uint("id", uint(id)))

	if err := h.userUseCase.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.CodeBadRequest,
			"Failed to delete user",
			err.Error(),
		))
		return
	}

	h.log.Info("DeleteUser: user deleted successfully", zap.Uint("id", uint(id)))
	c.JSON(http.StatusOK, domain.NewSuccessResponse(
		domain.CodeSuccess,
		"User deleted successfully",
		nil,
	))
}
