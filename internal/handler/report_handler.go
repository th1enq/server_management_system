package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/pkg/logger"
)

type ReportHandler struct {
	reportSrv services.ReportService
}

func NewReportHandler(reportSrv services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportSrv: reportSrv,
	}
}

func (h *ReportHandler) GetTodayReport(c *gin.Context) {
	report, err := h.reportSrv.GenerateDailyReport(c.Request.Context(), time.Now())
	if err != nil {
		logger.Error("failed to generate daily report", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to generate daily report",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Daily report generated successfully",
		report,
	))
}
