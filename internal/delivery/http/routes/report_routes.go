package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

type ReportRouter interface {
	RegisterRoutes(v1 *gin.RouterGroup)
}

type reportRouter struct {
	reportController *controllers.ReportController
	authMiddleware   *middleware.AuthMiddleware
}

func NewReportRouter(
	reportController *controllers.ReportController,
	authMiddleware *middleware.AuthMiddleware,
) ReportRouter {
	return &reportRouter{
		reportController: reportController,
		authMiddleware:   authMiddleware,
	}
}

func (h *reportRouter) RegisterRoutes(v1 *gin.RouterGroup) {
	reports := v1.Group("/reports")
	reports.Use(h.authMiddleware.RequireAuth())
	{
		reports.POST("/", h.authMiddleware.RequireAnyScope("admin:all", "report:write"), h.reportController.SendReportByDate)
		reports.POST("/daily", h.authMiddleware.RequireAnyScope("admin:all", "report:write"), h.reportController.SendReportDaily)
	}
}
