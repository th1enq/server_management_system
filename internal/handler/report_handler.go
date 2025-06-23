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
	err := h.reportSrv.SendReportForDaily(c.Request.Context(), time.Now())
	if err != nil {
		logger.Error("Failed to send daily report", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to send daily report",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Report sent successfully",
		nil,
	))
}
