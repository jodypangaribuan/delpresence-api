package middleware

import (
	"net/http"
	"strings"

	"delpresence-api/internal/repository"
	"delpresence-api/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := headerParts[1]

		// Validate the token
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			var statusCode int
			var message string

			switch err {
			case jwt.ErrExpiredToken:
				statusCode = http.StatusUnauthorized
				message = "Token has expired"
			case jwt.ErrInvalidToken:
				statusCode = http.StatusUnauthorized
				message = "Invalid token"
			default:
				statusCode = http.StatusInternalServerError
				message = "Failed to process token"
			}

			c.JSON(statusCode, gin.H{"error": message})
			c.Abort()
			return
		}

		// Check if user exists
		userRepo := repository.NewUserRepository()
		user, err := userRepo.GetUserByID(claims.UserID)
		if err != nil {
			if err == repository.ErrUserNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			c.Abort()
			return
		}

		// Set user info in the context
		c.Set("user_id", user.ID)
		c.Set("user_type", user.UserType)
		c.Set("email", user.Email)
		c.Set("first_name", user.FirstName)
		c.Set("middle_name", user.MiddleName)
		c.Set("last_name", user.LastName)

		c.Next()
	}
}
