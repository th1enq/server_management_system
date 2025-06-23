package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

type ReportHandler struct {
	reportSrv services.ReportService
}

type ReportRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

func NewReportHandler(reportSrv services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportSrv: reportSrv,
	}
}

func (h *ReportHandler) SendReportDaily(c *gin.Context) {
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
		"Daily report sent successfully",
		map[string]interface{}{
			"date": time.Now().Format("2006-01-02"),
		},
	))
}

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
		logger.Error("Failed to send report", err,
			zap.String("email", req.Email),
			zap.String("start_date", req.StartDate),
			zap.String("end_date", req.EndDate),
		)

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
