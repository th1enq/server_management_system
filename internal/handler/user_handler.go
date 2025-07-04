package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/middleware"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/models/dto"
	"github.com/th1enq/server_management_system/internal/services"

	"go.uber.org/zap"
)

type UserHandler struct {
	userService services.UserService
	log         *zap.Logger
}

func NewUserHandler(userService services.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

// GetProfile returns the current user's profile
// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.log.Warn("GetProfile: user ID not found in context")
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.CodeAuthError,
			"Authentication required",
			nil,
		))
		return
	}

	h.log.Info("GetProfile: fetching user profile", zap.Uint("userID", userID))

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("GetProfile: user not found", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.CodeNotFound,
			"User not found",
			nil,
		))
		return
	}

	h.log.Info("GetProfile: user profile retrieved successfully", zap.Uint("userID", userID))
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"User profile retrieved successfully",
		user,
	))
}

// UpdateProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ProfileUpdate true "Profile updates"
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.log.Warn("UpdateProfile: user ID not found in context")
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.CodeAuthError,
			"Authentication required",
			nil,
		))
		return
	}

	var updates dto.ProfileUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.log.Warn("UpdateProfile: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("UpdateProfile: updating user profile", zap.Uint("userID", userID), zap.Any("updates", updates))

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, updates)
	if err != nil {
		h.log.Error("UpdateProfile: failed to update user profile", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Failed to update profile",
			err.Error(),
		))
		return
	}
	h.log.Info("UpdateProfile: user profile updated successfully", zap.Uint("userID", userID))
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
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
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /api/v1/users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.log.Warn("ChangePassword: user ID not found in context")
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.CodeAuthError,
			"Authentication required",
			nil,
		))
		return
	}

	var req dto.PasswordUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("ChangePassword: changing password for user", zap.Uint("userID", userID))

	if err := h.userService.UpdatePassword(c.Request.Context(), userID, req); err != nil {
		h.log.Error("ChangePassword: failed to change password", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to change password",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
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
// @Success 200 {object} models.APIResponse{data=[]models.User}
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
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

	users, err := h.userService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("ListUsers: failed to fetch users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to fetch users",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
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
// @Success 201 {object} models.User
// @Failure 400 {object} models.APIResponse{data=models.User}
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("CreateUser: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
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

	var createdUser *models.User
	var err error

	if createdUser, err = h.userService.CreateUser(c.Request.Context(), req); err != nil {
		h.log.Error("CreateUser: failed to create user", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Failed to create user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		models.CodeSuccess,
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
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.log.Warn("UpdateUser: invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid user ID",
			err.Error(),
		))
		return
	}

	var updates dto.UserUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.log.Info("UpdateUser: updating user", zap.Uint("id", uint(id)), zap.Any("updates", updates))

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Failed to update user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
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
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.log.Warn("DeleteUser: invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid user ID",
			err.Error(),
		))
		return
	}

	h.log.Info("DeleteUser: deleting user", zap.Uint("id", uint(id)))

	if err := h.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Failed to delete user",
			err.Error(),
		))
		return
	}

	h.log.Info("DeleteUser: user deleted successfully", zap.Uint("id", uint(id)))
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"User deleted successfully",
		nil,
	))
}
