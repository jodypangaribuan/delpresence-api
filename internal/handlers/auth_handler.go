package handlers

import (
	"log"
	"net/http"
	"os"
	"time"

	"delpresence-api/internal/models"
	"delpresence-api/internal/repository"
	"delpresence-api/internal/utils"
	"delpresence-api/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userRepo  *repository.UserRepository
	adminRepo *repository.AdminRepository
	tokenRepo *repository.TokenRepository
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userRepo:  repository.NewUserRepository(),
		adminRepo: repository.NewAdminRepository(),
		tokenRepo: repository.NewTokenRepository(),
	}
}

// AdminLogin handles admin login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var input models.AdminLoginInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	log.Printf("Admin login attempt with email: %s", input.Email)

	// Find user by email
	user, err := h.userRepo.GetUserByEmail(input.Email)
	if err != nil {
		log.Printf("Admin login failed: user not found with email %s", input.Email)
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Verify that the user is an admin
	if user.UserType != models.AdminType {
		log.Printf("Admin login failed: user with email %s is not an admin", input.Email)
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Verify password
	if !user.ComparePassword(input.Password) {
		log.Printf("Admin login failed: incorrect password for email %s", input.Email)
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Get admin
	admin, err := h.adminRepo.GetAdminByUserID(user.ID)
	if err != nil {
		log.Printf("Admin login failed: admin profile not found for user ID %d", user.ID)
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Generate JWT token
	tokenString, expiryTime, err := jwt.GenerateAccessToken(user.ID, "", user.FirstName, user.MiddleName, user.LastName, user.Email)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate access token")
		return
	}

	// Generate refresh token
	refreshToken := generateRandomString(32)

	// Parse refresh token expiry from environment
	refreshExpiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if refreshExpiryStr == "" {
		refreshExpiryStr = "168h" // Default 7 days
	}

	refreshExpiry, err := time.ParseDuration(refreshExpiryStr)
	if err != nil {
		refreshExpiry = 168 * time.Hour // Default 7 days
	}

	// Save refresh token to database
	refreshTokenExpiry := time.Now().Add(refreshExpiry)
	err = h.tokenRepo.CreateToken(user.ID, refreshToken, models.RefreshToken, refreshTokenExpiry)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save refresh token")
		return
	}

	// Create response
	response := map[string]interface{}{
		"user": admin.ToAdminResponse(),
		"tokens": map[string]interface{}{
			"access_token":  tokenString,
			"refresh_token": refreshToken,
			"expires_in":    int(time.Until(expiryTime).Seconds()),
			"token_type":    "Bearer",
		},
		"user_type": "admin",
	}

	log.Printf("Admin login successful for email: %s", input.Email)
	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// GetCurrentUser handles getting the current user's information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authorized")
		return
	}

	// Get user from database
	user, err := h.userRepo.GetUserByID(userID.(uint))
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	var userResponse interface{}

	// Get user profile based on user type
	switch user.UserType {
	case models.AdminType:
		admin, err := h.adminRepo.GetAdminByUserID(user.ID)
		if err != nil {
			utils.NotFoundResponse(c, "Admin profile not found")
			return
		}
		userResponse = admin.ToAdminResponse()
	default:
		userResponse = map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"first_name":  user.FirstName,
			"middle_name": user.MiddleName,
			"last_name":   user.LastName,
			"user_type":   user.UserType,
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "User information retrieved successfully", userResponse)
}

// Helper function to generate a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int(time.Now().UnixNano()%int64(len(charset)))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
