package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/handler"
	"github.com/th1enq/server_management_system/internal/middleware"
)

// UserRoutes handles user management routing
type UserRoutes struct {
	userHandler    *handler.UserHandler // Using existing handler for now
	authMiddleware *middleware.AuthMiddleware
}

// NewUserRoutes creates a new user routes handler
func NewUserRoutes(
	userHandler *handler.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
) *UserRoutes {
	return &UserRoutes{
		userHandler:    userHandler,
		authMiddleware: authMiddleware,
	}
}

// Setup configures user management routes
func (ur *UserRoutes) Setup(rg *gin.RouterGroup) {
	// All user routes require authentication
	users := rg.Group("/users")
	users.Use(ur.authMiddleware.RequireAuth())

	// User management routes (admin only)
	ur.setupUserManagementRoutes(users)

	// Profile management routes (user can manage own profile)
	ur.setupProfileRoutes(users)
}

// setupUserManagementRoutes configures admin user management endpoints
func (ur *UserRoutes) setupUserManagementRoutes(rg *gin.RouterGroup) {
	// Admin-only user CRUD operations
	readScope := ur.authMiddleware.RequireAnyScope("admin:all", "user:read")
	writeScope := ur.authMiddleware.RequireAnyScope("admin:all", "user:write")
	deleteScope := ur.authMiddleware.RequireAnyScope("admin:all", "user:delete")

	// User CRUD
	rg.GET("/", readScope, ur.userHandler.ListUsers)
	rg.POST("/", writeScope, ur.userHandler.CreateUser)
	rg.PUT("/:id", writeScope, ur.userHandler.UpdateUser)
	rg.DELETE("/:id", deleteScope, ur.userHandler.DeleteUser)
}

// setupProfileRoutes configures user profile management endpoints
func (ur *UserRoutes) setupProfileRoutes(rg *gin.RouterGroup) {
	// Profile operations - users can manage their own profile
	profileReadScope := ur.authMiddleware.RequireAnyScope("admin:all", "profile:read")
	profileWriteScope := ur.authMiddleware.RequireAnyScope("admin:all", "profile:write")

	// Profile management
	rg.GET("/profile", profileReadScope, ur.userHandler.GetProfile)
	rg.PUT("/profile", profileWriteScope, ur.userHandler.UpdateProfile)
	rg.POST("/change-password", profileWriteScope, ur.userHandler.ChangePassword)
}
