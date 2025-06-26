package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"go.uber.org/zap"
)

type ReportHandler struct {
	reportSrv services.ReportService
	logger    *zap.Logger
}

// ReportRequest represents the request payload for generating reports by date range
// @Description Request structure for generating reports within a specific date range
type ReportRequest struct {
	StartDate string `json:"start_date" binding:"required" example:"2025-06-20 00:00:00"` // Start date in format YYYY-MM-DD HH:MM:SS
	EndDate   string `json:"end_date" binding:"required" example:"2025-06-21 23:59:59"`   // End date in format YYYY-MM-DD HH:MM:SS
	Email     string `json:"email" binding:"required,email" example:"admin@example.com"`  // Email address to send the report to
}

func NewReportHandler(reportSrv services.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		reportSrv: reportSrv,
		logger:    logger,
	}
}

// SendReportDaily godoc
// @Summary Send daily report
// @Description Send a daily server monitoring report for today's date
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/reports/daily [post]
func (h *ReportHandler) SendReportDaily(c *gin.Context) {
	err := h.reportSrv.SendReportForDaily(c.Request.Context(), time.Now())
	if err != nil {
		h.logger.Error("Failed to send daily report", zap.Error(err))

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to send daily report",
			err.Error(),
		))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Daily report sent successfully",
		map[string]interface{}{
			"date": time.Now().Format("2006-01-02"),
		},
	))
}

// SendReportByDate godoc
// @Summary Send report by date range
// @Description Send a server monitoring report for a specified date range to an email address
// @Tags reports
// @Accept json
// @Produce json
// @Param report body ReportRequest true "Report request with date range and email"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /api/v1/reports [post]
func (h *ReportHandler) SendReportByDate(c *gin.Context) {
	var req ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	startDate, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartDate, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid start date format, expected YYYY-MM-DD HH:MM:SS",
			err.Error(),
		))
		return
	}

	endDate, err := time.ParseInLocation("2006-01-02 15:04:05 ", req.EndDate, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.CodeBadRequest,
			"Invalid end date format, expected YYYY-MM-DD HH:MM:SS",
			err.Error(),
		))
		return
	}

	err = h.reportSrv.SendReportForDateRange(c.Request.Context(), startDate, endDate, req.Email)
	if err != nil {

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.CodeInternalServerError,
			"Failed to send report",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.CodeSuccess,
		"Report sent successfully",
		map[string]interface{}{
			"email":      req.Email,
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
	))
}
