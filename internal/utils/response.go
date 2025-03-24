package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response format
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// LogError logs error with timestamp and additional info
func LogError(handler string, action string, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[ERROR] [%s] %s - %s: %v\n", timestamp, handler, action, err)
}

// LogInfo logs information with timestamp
func LogInfo(handler string, action string, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[INFO] [%s] %s - %s: %s\n", timestamp, handler, action, message)
}

// LogWarning logs warning with timestamp
func LogWarning(handler string, action string, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[WARNING] [%s] %s - %s: %s\n", timestamp, handler, action, message)
}

// SuccessResponse returns a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	LogInfo("Success", "Response", message)
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
func ValidationErrorResponse(c *gin.Context, message string) {
	LogError("Validation", "Input Validation", fmt.Errorf(message))
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": message,
	})
}

// InternalServerErrorResponse returns a 500 internal server error response
func InternalServerErrorResponse(c *gin.Context, message string) {
	LogError("InternalServer", "Server Error", fmt.Errorf(message))
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"message": "Internal server error",
	})
}

// UnauthorizedResponse returns a 401 unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	LogError("Unauthorized", "Authentication", fmt.Errorf(message))
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"message": message,
	})
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
func BadRequestResponse(c *gin.Context, message string, data interface{}) {
	LogError("BadRequest", "Request Processing", fmt.Errorf(message))
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": message,
		"data":    data,
	})
}
