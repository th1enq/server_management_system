package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
	"github.com/th1enq/server_management_system/internal/domain"
	"github.com/th1enq/server_management_system/internal/domain/entity"
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
// @Success 200 {object} domain.APIResponse{data=dto.UserResponse}
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
		h.log.Warn("GetProfile: user not found", zap.Uint("userID", userID), zap.Error(err))
		h.userPresenter.UserNotFound(c, "User not found")
		return
	}

	h.log.Info("GetProfile: user profile retrieved successfully", zap.Uint("userID", userID))
	h.userPresenter.ProfileRetrieved(c, dto.FromEntityToUserResponse(user))
}

// UpdateProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ProfileUpdate true "Profile updates"
// @Success 200 {object} domain.APIResponse{data=dto.UserResponse}
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
	h.userPresenter.ProfileUpdated(c, dto.FromEntityToUserResponse(user))
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

	h.log.Info("ChangePassword: password changed successfully", zap.Uint("userID", userID))
	h.userPresenter.PasswordChanged(c)
}

// ListUsers returns a list of users (admin only)
// @Summary List users
// @Description Get a list of all users (admin only)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort query string false "Sort field" default(username)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {object} domain.APIResponse{data=[]dto.UserResponse}
// @Failure 401 {object} domain.APIResponse
// @Failure 403 {object} domain.APIResponse
// @Router /api/v1/users [get]
func (h *UserController) ListUsers(c *gin.Context) {
	var pagination dto.UserPagination

	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.log.Warn("ListUsers: invalid query parameters", zap.Error(err))
		h.userPresenter.InvalidRequest(c, "Invalid query parameters", err)
	}

	users, err := h.userUseCase.ListUsers(c.Request.Context(), pagination)
	if err != nil {
		h.log.Error("ListUsers: failed to fetch users", zap.Error(err))
		h.userPresenter.InternalServerError(c, "Failed to fetch users", err)
		return
	}

	h.log.Info("ListUsers: users retrieved successfully",
		zap.Int("total", len(users)),
		zap.Any("pagination", pagination),
		zap.String("request_id", c.GetString("request_id")),
	)

	h.userPresenter.UsersRetrieved(c, dto.FromEntityListToUserResponseList(users))
}

// CreateUser creates a new user (admin only)
// @Summary Create user
// @Description Create a new user (admin only)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} domain.APIResponse{data=dto.UserResponse}
// @Failure 401 {object} domain.APIResponse
// @Failure 403 {object} domain.APIResponse
// @Router /api/v1/users [post]
func (h *UserController) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
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

	var createdUser *entity.User
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

	h.log.Info("CreateUser: user created successfully", zap.Uint("userID", createdUser.ID))
	h.userPresenter.UserCreated(c, dto.FromEntityToUserResponse(createdUser))
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
// @Success 200 {object} domain.APIResponse{data=dto.UserResponse}
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

	h.log.Info("UpdateUser: user updated successfully", zap.Uint("id", user.ID), zap.String("username", user.Username))
	h.userPresenter.UserUpdated(c, dto.FromEntityToUserResponse(user))
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
	h.userPresenter.UserDeleted(c)
}
