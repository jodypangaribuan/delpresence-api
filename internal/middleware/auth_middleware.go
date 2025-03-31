package middleware

import (
	"log"
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
		log.Printf("Processing request: %s %s", c.Request.Method, path)

		// Skip authentication for certain paths if needed
		if strings.HasPrefix(path, "/api/v1/health") {
			c.Next()
			return
		}

		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("Request denied: Authorization header is missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			log.Printf("Request denied: Invalid auth header format: %s", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := headerParts[1]
		if len(tokenString) < 20 {
			log.Printf("Request denied: Token too short (%d chars)", len(tokenString))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		log.Printf("Token received. Length: %d, First 15 chars: %s...",
			len(tokenString),
			tokenString[:min(15, len(tokenString))])

		// For student-related API endpoints, priorities student token validation
		isMahasiswaEndpoint := strings.HasPrefix(path, "/api/v1/mahasiswa")
		var userID uint

		// CASE 1: For mahasiswa endpoints, try campus token validation first
		if isMahasiswaEndpoint {
			log.Printf("Mahasiswa endpoint: trying campus token validation first")

			// Try to validate as campus token
			campusUserID, campusErr := jwt.ValidateCampusToken(tokenString)
			if campusErr == nil {
				// Campus token validation succeeded
				userID = uint(campusUserID)
				log.Printf("Success: Campus auth for user ID: %d", userID)

				// Set user info in the context
				c.Set("user_id", userID)
				c.Set("campus_authenticated", true)
				c.Next()
				return
			}

			log.Printf("Campus token validation failed: %v, trying regular token", campusErr)
		}

		// CASE 2: Try regular JWT validation for all endpoints
		claims, err := jwt.ValidateToken(tokenString)
		if err == nil {
			// Regular token validation succeeded
			userID = claims.UserID
			log.Printf("Success: Regular token validation for user ID: %d", userID)

			// Check if user exists in our database
			userRepo := repository.NewUserRepository()
			user, dbErr := userRepo.GetUserByID(userID)
			if dbErr != nil {
				if dbErr == repository.ErrUserNotFound {
					log.Printf("User with ID %d not found in database", userID)
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				} else {
					log.Printf("Database error looking up user %d: %v", userID, dbErr)
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
		if !isMahasiswaEndpoint {
			log.Printf("Regular token validation failed: %v, trying campus token", err)

			campusUserID, campusErr := jwt.ValidateCampusToken(tokenString)
			if campusErr == nil {
				// Campus token validation succeeded for non-mahasiswa endpoint
				userID = uint(campusUserID)
				log.Printf("Success: Campus auth for user ID: %d on non-mahasiswa endpoint", userID)

				// Set user info in the context
				c.Set("user_id", userID)
				c.Set("campus_authenticated", true)
				c.Next()
				return
			}

			log.Printf("Campus token validation also failed: %v", campusErr)
		}

		// If we reach here, authentication failed
		log.Printf("Authentication failed: all validation methods failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
	}
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
