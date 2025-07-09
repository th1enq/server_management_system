package dto

type ReportRequest struct {
	StartDate string `json:"start_date" binding:"required" example:"2025-06-20 00:00:00"`
	EndDate   string `json:"end_date" binding:"required" example:"2025-06-21 23:59:59"`
	Email     string `json:"email" binding:"required,email" example:"admin@example.com"`
}
