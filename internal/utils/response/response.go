package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasNext    bool  `json:"has_next,omitempty"`
	HasPrev    bool  `json:"has_prev,omitempty"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	response := Response{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC(),
		RequestID: c.GetString("request_id"),
	}
	c.JSON(statusCode, response)
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string, err interface{}) {
	response := Response{
		Success:   false,
		Message:   message,
		Error:     err,
		Timestamp: time.Now().UTC(),
		RequestID: c.GetString("request_id"),
	}
	c.JSON(statusCode, response)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusBadRequest, message, err)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string, details interface{}) {
	errorDetail := ErrorDetail{
		Code:    "AUTH_UNAUTHORIZED",
		Message: message,
		Details: details,
	}
	Error(c, http.StatusUnauthorized, "Authentication required", errorDetail)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	errorDetail := ErrorDetail{
		Code:    "AUTH_FORBIDDEN",
		Message: message,
	}
	Error(c, http.StatusForbidden, "Access denied", errorDetail)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	errorDetail := ErrorDetail{
		Code:    "RESOURCE_NOT_FOUND",
		Message: message,
	}
	Error(c, http.StatusNotFound, "Resource not found", errorDetail)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string, details interface{}) {
	errorDetail := ErrorDetail{
		Code:    "CONFLICT",
		Message: message,
		Details: details,
	}
	Error(c, http.StatusConflict, "Resource conflict", errorDetail)
}

// ValidationError sends a 422 Unprocessable Entity response
func ValidationError(c *gin.Context, errors interface{}) {
	errorDetail := ErrorDetail{
		Code:    "VALIDATION_FAILED",
		Message: "Validation failed",
		Details: errors,
	}
	Error(c, http.StatusUnprocessableEntity, "Validation failed", errorDetail)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	errorDetail := ErrorDetail{
		Code:    "SERVER_INTERNAL",
		Message: "Internal server error",
	}
	Error(c, http.StatusInternalServerError, message, errorDetail)
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	errorDetail := ErrorDetail{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: message,
	}
	Error(c, http.StatusTooManyRequests, "Too many requests", errorDetail)
}

// WithMeta adds metadata to success response
func WithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta *Meta) {
	response := Response{
		Success:   true,
		Data:      data,
		Message:   message,
		Meta:      meta,
		Timestamp: time.Now().UTC(),
		RequestID: c.GetString("request_id"),
	}
	c.JSON(statusCode, response)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, message string, data interface{}, page, perPage int, total int64) {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	meta := &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
	WithMeta(c, http.StatusOK, message, data, meta)
}
