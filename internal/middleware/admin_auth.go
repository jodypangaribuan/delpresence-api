package middleware

import (
	"fmt"
	"os"
	"strings"

	"delpresence-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AdminAuth middleware untuk memverifikasi token JWT admin
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header diperlukan")
			c.Abort()
			return
		}

		// Check if token format is valid (Bearer token)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.UnauthorizedResponse(c, "Format token tidak valid. Gunakan format: Bearer {token}")
			c.Abort()
			return
		}

		// Extract token from header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode signing tidak valid: %v", token.Header["alg"])
			}

			// Return secret key for verification
			secretKey := os.Getenv("JWT_SECRET_KEY")
			if secretKey == "" {
				secretKey = "default_secret_key"
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			utils.UnauthorizedResponse(c, fmt.Sprintf("Token tidak valid: %v", err))
			c.Abort()
			return
		}

		// Validate token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check if token is for admin
			userType, ok := claims["user_type"].(string)
			if !ok || userType != "admin" {
				utils.ForbiddenResponse(c, "Token bukan untuk akses admin")
				c.Abort()
				return
			}

			// Set claims to context
			userID, _ := claims["user_id"].(float64)
			adminID, _ := claims["admin_id"].(float64)
			accessLevel, _ := claims["access_level"].(string)

			c.Set("user_id", uint(userID))
			c.Set("admin_id", uint(adminID))
			c.Set("access_level", accessLevel)
			c.Set("user_type", userType)

			c.Next()
		} else {
			utils.UnauthorizedResponse(c, "Token tidak dapat divalidasi")
			c.Abort()
			return
		}
	}
}
