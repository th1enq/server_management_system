package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/interfaces/middleware"
)

// MiddlewareSetup handles route-level middleware configuration
// Note: Global middleware is now handled by middleware.MiddlewareManager
type MiddlewareSetup struct {
	middlewareManager *middleware.MiddlewareManager
}

// NewMiddlewareSetup creates a new middleware setup handler
func NewMiddlewareSetup(middlewareManager *middleware.MiddlewareManager) *MiddlewareSetup {
	return &MiddlewareSetup{
		middlewareManager: middlewareManager,
	}
}

// SetupForRouter configures middleware for the entire router
func (ms *MiddlewareSetup) SetupForRouter(engine *gin.Engine) {
	// Delegate to middleware manager for global middleware
	ms.middlewareManager.SetupGlobalMiddleware(engine)
}

// SetupForAPIRoutes configures middleware for API route groups
func (ms *MiddlewareSetup) SetupForAPIRoutes(rg *gin.RouterGroup) {
	// API-specific middleware can be added here
	// For example, API versioning, API key auth, etc.
}

// SetupForAdminRoutes configures middleware for admin route groups
func (ms *MiddlewareSetup) SetupForAdminRoutes(rg *gin.RouterGroup) {
	// Admin-specific middleware
	// Could include additional security, audit logging, etc.
}

// SetupAPIKeyAuth sets up API key authentication for specific routes
func (ms *MiddlewareSetup) SetupAPIKeyAuth(rg *gin.RouterGroup) {
	ms.middlewareManager.SetupAPIKeyAuth(rg)
}
