package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

type AuthRouter interface {
	RegisterRoutes(v1 *gin.RouterGroup)
}

type authRouter struct {
	authController *controllers.AuthController
	authMiddleware *middleware.AuthMiddleware
}

func NewAuthRouter(
	authController *controllers.AuthController,
	authMiddleware *middleware.AuthMiddleware,
) AuthRouter {
	return &authRouter{
		authController: authController,
		authMiddleware: authMiddleware,
	}
}

func (r *authRouter) RegisterRoutes(v1 *gin.RouterGroup) {
	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", r.authController.Login)
		auth.POST("/register", r.authController.Register)
		auth.POST("/refresh", r.authController.RefreshToken)
	}

	// Protected auth routes
	authProtected := v1.Group("/auth")
	authProtected.Use(r.authMiddleware.RequireAuth())
	{
		authProtected.POST("/logout", r.authController.Logout)
	}
}
