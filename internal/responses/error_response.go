package responses

import (
	"fmt"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     ErrorDetail `json:"error"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorDetail struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"-"`
}

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Tag     string      `json:"tag,omitempty"`
	Value   interface{} `json:"value,omitempty"`
}

type ErrorCodes struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// Predefined error codes with proper HTTP status codes
var (
	// Auth errors (401)
	ErrCodeUnauthorized        = NewErrorCode("AUTH_UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrCodeInvalidCredentials  = NewErrorCode("AUTH_INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrCodeInvalidToken        = NewErrorCode("AUTH_INVALID_TOKEN", "Invalid or malformed token", http.StatusUnauthorized)
	ErrCodeTokenExpired        = NewErrorCode("AUTH_TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)
	ErrCodeEmailNotVerified    = NewErrorCode("AUTH_EMAIL_NOT_VERIFIED", "Email address not verified", http.StatusUnauthorized)
	ErrCodeAccountSuspended    = NewErrorCode("AUTH_ACCOUNT_SUSPENDED", "Account has been suspended", http.StatusUnauthorized)
	ErrCodeInvalidRefreshToken = NewErrorCode("AUTH_INVALID_REFRESH_TOKEN", "Invalid refresh token", http.StatusUnauthorized)

	// Authorization errors (403)
	ErrCodeForbidden        = NewErrorCode("AUTH_FORBIDDEN", "Access denied", http.StatusForbidden)
	ErrCodeInsufficientRole = NewErrorCode("AUTH_INSUFFICIENT_ROLE", "Insufficient permissions for this action", http.StatusForbidden)

	// Resource errors (404)
	ErrCodeUserNotFound     = NewErrorCode("RESOURCE_USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrCodeResourceNotFound = NewErrorCode("RESOURCE_NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrCodeRoleNotFound     = NewErrorCode("RESOURCE_ROLE_NOT_FOUND", "Role not found", http.StatusNotFound)

	// Conflict errors (409)
	ErrCodeEmailExists         = NewErrorCode("CONFLICT_EMAIL_EXISTS", "Email address already registered", http.StatusConflict)
	ErrCodeResourceConflict    = NewErrorCode("CONFLICT_RESOURCE", "Resource conflict", http.StatusConflict)
	ErrCodeUserAlreadyVerified = NewErrorCode("CONFLICT_ALREADY_VERIFIED", "User already verified", http.StatusConflict)

	// Validation errors (422)
	ErrCodeValidation   = NewErrorCode("VALIDATION_FAILED", "Validation failed", http.StatusUnprocessableEntity)
	ErrCodeInvalidInput = NewErrorCode("VALIDATION_INVALID_INPUT", "Invalid input provided", http.StatusUnprocessableEntity)

	// Rate limiting (429)
	ErrCodeTooManyRequests = NewErrorCode("RATE_LIMIT_EXCEEDED", "Too many requests, please try again later", http.StatusTooManyRequests)
	ErrCodeLoginBlocked    = NewErrorCode("RATE_LIMIT_LOGIN_BLOCKED", "Too many login attempts, account temporarily blocked", http.StatusTooManyRequests)

	// Server errors (500)
	ErrCodeInternalServer     = NewErrorCode("SERVER_INTERNAL", "Internal server error", http.StatusInternalServerError)
	ErrCodeDatabaseError      = NewErrorCode("SERVER_DATABASE_ERROR", "Database error occurred", http.StatusInternalServerError)
	ErrCodeServiceUnavailable = NewErrorCode("SERVER_SERVICE_UNAVAILABLE", "Service temporarily unavailable", http.StatusServiceUnavailable)
)

func NewErrorCode(code, description string, status int) ErrorCodes {
	return ErrorCodes{
		Code:        code,
		Description: description,
		Status:      status,
	}
}

func NewErrorResponse(code, message string, details interface{}, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}

// String returns a string representation of the error
func (e *ErrorResponse) String() string {
	return fmt.Sprintf("%s: %s", e.Error.Code, e.Error.Message)
}

// Helper functions for common errors
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

func NewValidationErrors(errors []ValidationError) []ValidationError {
	return errors
}
