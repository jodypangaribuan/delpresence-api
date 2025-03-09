package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response format
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse returns a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// ValidationErrorResponse returns a validation error response
func ValidationErrorResponse(c *gin.Context, err interface{}) {
	ErrorResponse(c, http.StatusBadRequest, "Validation error", err)
}

// InternalServerErrorResponse returns a 500 internal server error response
func InternalServerErrorResponse(c *gin.Context, err interface{}) {
	ErrorResponse(c, http.StatusInternalServerError, "Internal server error", err)
}

// UnauthorizedResponse returns a 401 unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// ForbiddenResponse returns a 403 forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	ErrorResponse(c, http.StatusForbidden, message, nil)
}

// NotFoundResponse returns a 404 not found response
func NotFoundResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	ErrorResponse(c, http.StatusNotFound, message, nil)
}

// BadRequestResponse returns a 400 bad request response
func BadRequestResponse(c *gin.Context, message string, err interface{}) {
	if message == "" {
		message = "Bad request"
	}
	ErrorResponse(c, http.StatusBadRequest, message, err)
}
