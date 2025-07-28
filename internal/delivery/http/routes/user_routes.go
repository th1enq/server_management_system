package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

type UserRouter interface {
	RegisterRoutes(v1 *gin.RouterGroup)
}

type userRouter struct {
	userController *controllers.UserController
	authMiddleware *middleware.AuthMiddleware
}

func NewUserRouter(
	userController *controllers.UserController,
	authMiddleware *middleware.AuthMiddleware,
) UserRouter {
	return &userRouter{
		userController: userController,
		authMiddleware: authMiddleware,
	}
}

func (h *userRouter) RegisterRoutes(v1 *gin.RouterGroup) {
	users := v1.Group("/users")
	users.Use(h.authMiddleware.RequireAuth())
	{
		users.GET("/", h.authMiddleware.RequireAnyScope("admin:all", "user:read"), h.userController.ListUsers)
		users.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "user:write"), h.userController.CreateUser)
		users.PUT("/:id", h.authMiddleware.RequireAnyScope("admin:all", "user:write"), h.userController.UpdateUser)
		users.DELETE("/:id", h.authMiddleware.RequireAnyScope("admin:all", "user:delete"), h.userController.DeleteUser)
		users.GET("/profile", h.authMiddleware.RequireAnyScope("admin:all", "profile:read"), h.userController.GetProfile)
		users.PUT("/profile", h.authMiddleware.RequireAnyScope("admin:all", "profile:write"), h.userController.UpdateProfile)
		users.POST("/change-password", h.authMiddleware.RequireAnyScope("admin:all", "profile:write"), h.userController.ChangePassword)
	}
}
