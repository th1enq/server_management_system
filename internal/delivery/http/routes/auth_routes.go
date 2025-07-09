package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/interfaces/http/controllers"
	"github.com/th1enq/server_management_system/internal/middleware"
)

// AuthRoutes handles authentication-related routing
type AuthRoutes struct {
	authController *controllers.AuthController
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthRoutes creates a new auth routes handler
func NewAuthRoutes(
	authController *controllers.AuthController,
	authMiddleware *middleware.AuthMiddleware,
) *AuthRoutes {
	return &AuthRoutes{
		authController: authController,
		authMiddleware: authMiddleware,
	}
}

// Setup configures authentication routes
func (ar *AuthRoutes) Setup(rg *gin.RouterGroup) {
	// Public auth routes (no authentication required)
	authPublic := rg.Group("/auth")
	{
		// Authentication endpoints
		authPublic.POST("/login", ar.authController.Login)
		authPublic.POST("/register", ar.authController.Register)
		authPublic.POST("/refresh", ar.authController.RefreshToken)
	}

	// Protected auth routes (authentication required)
	authProtected := rg.Group("/auth")
	authProtected.Use(ar.authMiddleware.RequireAuth())
	{
		// Session management
		authProtected.POST("/logout", ar.authController.Logout)
	}
}
