package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/handler"
	"github.com/th1enq/server_management_system/internal/middleware"
)

// ReportRoutes handles report-related routing
type ReportRoutes struct {
	reportHandler  *handler.ReportHandler // Using existing handler for now
	authMiddleware *middleware.AuthMiddleware
}

// NewReportRoutes creates a new report routes handler
func NewReportRoutes(
	reportHandler *handler.ReportHandler,
	authMiddleware *middleware.AuthMiddleware,
) *ReportRoutes {
	return &ReportRoutes{
		reportHandler:  reportHandler,
		authMiddleware: authMiddleware,
	}
}

// Setup configures report routes
func (rr *ReportRoutes) Setup(rg *gin.RouterGroup) {
	// All report routes require authentication and report permissions
	reports := rg.Group("/reports")
	reports.Use(rr.authMiddleware.RequireAuth())

	// Report generation routes
	rr.setupReportGenerationRoutes(reports)
}

// setupReportGenerationRoutes configures report generation endpoints
func (rr *ReportRoutes) setupReportGenerationRoutes(rg *gin.RouterGroup) {
	writeScope := rr.authMiddleware.RequireAnyScope("admin:all", "report:write")

	// Manual report generation
	rg.POST("/", writeScope, rr.reportHandler.SendReportByDate)

	// Daily report generation
	rg.POST("/daily", writeScope, rr.reportHandler.SendReportDaily)
}
