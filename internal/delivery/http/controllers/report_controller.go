package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/usecases"
	"go.uber.org/zap"
)

type ReportController struct {
	reportUseCase   usecases.ReportUseCase
	reportPresenter presenters.ReportPresenter
	logger          *zap.Logger
}

func NewReportController(
	reportUseCase usecases.ReportUseCase,
	reportPresenter presenters.ReportPresenter,
	logger *zap.Logger,
) *ReportController {
	return &ReportController{
		reportUseCase:   reportUseCase,
		reportPresenter: reportPresenter,
		logger:          logger,
	}
}

// SendReportDaily godoc
// @Summary Send daily report
// @Description Send a daily server monitoring report for today's date
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/reports/daily [post]
func (h *ReportController) SendReportDaily(c *gin.Context) {
	now := time.Now()
	err := h.reportUseCase.SendReportForDaily(c.Request.Context(), now)
	if err != nil {
		h.logger.Error("Failed to send daily report", zap.Error(err))
		h.reportPresenter.InternalServerError(c, "Failed to send daily report", err)
		return
	}
	h.reportPresenter.DailyReportSent(c)
}

// SendReportByDate godoc
// @Summary Send report by date range
// @Description Send a server monitoring report for a specified date range to an email address
// @Tags reports
// @Accept json
// @Produce json
// @Param report body dto.ReportRequest true "Report request with date range and email"
// @Success 200 {object} domain.APIResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Security BearerAuth
// @Router /api/v1/reports [post]
func (h *ReportController) SendReportByDate(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid report request", zap.Error(err))
		h.reportPresenter.InvalidRequest(c, "Invalid request data", err)
	}

	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		h.logger.Warn("Failed to load timezone", zap.Error(err))
		h.reportPresenter.InternalServerError(c, "Failed to load timezone", err)
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartDate, loc)
	if err != nil {
		h.logger.Warn("Invalid start date format", zap.Error(err))
		h.reportPresenter.ValidationError(c, "Invalid start date format, expected YYYY-MM-DD HH:MM:SS", err)
		return
	}

	endDate, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndDate, loc)
	if err != nil {
		h.logger.Warn("Invalid end date format", zap.Error(err))
		h.reportPresenter.ValidationError(c, "Invalid end date format, expected YYYY-MM-DD HH:MM:SS", err)
		return
	}

	err = h.reportUseCase.SendReportForDateRange(c.Request.Context(), startDate, endDate, req.Email)
	if err != nil {
		h.logger.Error("Failed to send report for date range", zap.Error(err), zap.String("email", req.Email))
		h.reportPresenter.InternalServerError(c, "Failed to send report for date range", err)
		return
	}

	h.logger.Info("Report sent successfully", zap.String("email", req.Email), zap.Time("start_date", startDate), zap.Time("end_date", endDate))
	h.reportPresenter.CustomReportSent(c)
}
