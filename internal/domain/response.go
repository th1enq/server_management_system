package domain

// APIResponse standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Common response codes
const (
	// Success codes
	CodeSuccess = "SUCCESS"
	CodeCreated = "CREATED"
	CodeUpdated = "UPDATED"
	CodeDeleted = "DELETED"

	// Error codes
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeDatabaseError       = "DATABASE_ERROR"
	CodeAuthError           = "AUTH_ERROR"
	CodeTokenExpired        = "TOKEN_EXPIRED"
	CodeInvalidToken        = "INVALID_TOKEN"
)

// NewSuccessResponse creates a new success response
func NewSuccessResponse(code, message string, data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string, details interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Code:    code,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
