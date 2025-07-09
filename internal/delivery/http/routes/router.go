package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/th1enq/server_management_system/internal/interfaces/http/controllers"
	"github.com/th1enq/server_management_system/internal/middleware"
)

// RouterConfig holds all dependencies needed for route setup
type RouterConfig struct {
	// Controllers
	AuthController   *controllers.AuthController
	ServerController *controllers.ServerController
	UserController   *controllers.UserController
	ReportController *controllers.ReportController
	JobsController   *controllers.JobsController

	// Middleware
	AuthMiddleware    *middleware.AuthMiddleware
	MiddlewareManager *middleware.MiddlewareManager
}

// Router manages all HTTP routes
type Router struct {
	config *RouterConfig
	engine *gin.Engine
}

// NewRouter creates a new router with the given configuration
func NewRouter(config *RouterConfig) *Router {
	router := &Router{
		config: config,
		engine: gin.New(),
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// setupMiddleware configures global middleware
func (r *Router) setupMiddleware() {
	// Use middleware manager for global middleware setup
	r.config.MiddlewareManager.SetupGlobalMiddleware(r.engine)
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	// Health check endpoint
	r.engine.GET("/health", r.healthCheck)

	// Swagger documentation
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.engine.Group("/api/v1")

	// Setup route groups
	r.setupAuthRoutes(v1)
	r.setupServerRoutes(v1)
	r.setupUserRoutes(v1)
	r.setupReportRoutes(v1)
	r.setupJobsRoutes(v1)
}

// healthCheck handles health check requests
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "server_management_system",
	})
}

// setupAuthRoutes configures authentication routes
func (r *Router) setupAuthRoutes(rg *gin.RouterGroup) {
	authRoutes := NewAuthRoutes(r.config.AuthController, r.config.AuthMiddleware)
	authRoutes.Setup(rg)
}

// setupServerRoutes configures server management routes
func (r *Router) setupServerRoutes(rg *gin.RouterGroup) {
	serverRoutes := NewServerRoutes(r.config.ServerController, r.config.AuthMiddleware)
	serverRoutes.Setup(rg)
}

// setupUserRoutes configures user management routes
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {
	userRoutes := NewUserRoutes(r.config.UserController, r.config.AuthMiddleware)
	userRoutes.Setup(rg)
}

// setupReportRoutes configures report routes
func (r *Router) setupReportRoutes(rg *gin.RouterGroup) {
	reportRoutes := NewReportRoutes(r.config.ReportController, r.config.AuthMiddleware)
	reportRoutes.Setup(rg)
}

// setupJobsRoutes configures job monitoring routes
func (r *Router) setupJobsRoutes(rg *gin.RouterGroup) {
	jobsRoutes := NewJobsRoutes(r.config.JobsController, r.config.AuthMiddleware)
	jobsRoutes.Setup(rg)
}
