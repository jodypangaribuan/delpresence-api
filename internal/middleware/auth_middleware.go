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
		path := c.Request.URL.Path

		// Skip authentication for certain paths if needed
		if strings.HasPrefix(path, "/api/v1/health") {
			c.Next()
			return
		}

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
		if len(tokenString) < 20 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// For student-related API endpoints, priorities student token validation
		isMahasiswaEndpoint := strings.HasPrefix(path, "/api/v1/mahasiswa")
		isAssistantEndpoint := strings.HasPrefix(path, "/api/v1/assistant")
		var userID uint

		// CASE 1: For mahasiswa or assistant endpoints, try campus token validation first
		if isMahasiswaEndpoint || isAssistantEndpoint {
			// Try to validate as campus token
			campusUserID, campusErr := jwt.ValidateCampusToken(tokenString)
			if campusErr == nil {
				// Campus token validation succeeded
				userID = uint(campusUserID)

				// Set user info in the context
				c.Set("user_id", userID)
				c.Set("campus_user_id", campusUserID)
				c.Set("campus_authenticated", true)
				c.Next()
				return
			}
		}

		// CASE 2: Try regular JWT validation for all endpoints
		claims, err := jwt.ValidateToken(tokenString)
		if err == nil {
			// Regular token validation succeeded
			userID = claims.UserID

			// Check if user exists in our database
			userRepo := repository.NewUserRepository()
			user, dbErr := userRepo.GetUserByID(userID)
			if dbErr != nil {
				if dbErr == repository.ErrUserNotFound {
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
			return
		}

		// CASE 3: If we're not on a mahasiswa endpoint and the regular token failed,
		// try campus token as last resort
		if !isMahasiswaEndpoint && !isAssistantEndpoint {
			campusUserID, campusErr := jwt.ValidateCampusToken(tokenString)
			if campusErr == nil {
				// Campus token validation succeeded for non-mahasiswa/non-assistant endpoint
				userID = uint(campusUserID)

				// Set user info in the context
				c.Set("user_id", userID)
				c.Set("campus_user_id", campusUserID)
				c.Set("campus_authenticated", true)
				c.Next()
				return
			}
		}

		// If we reach here, authentication failed
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
	}
}
