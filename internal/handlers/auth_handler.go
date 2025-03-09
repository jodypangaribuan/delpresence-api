package handlers

import (
	"crypto/rand"
	"encoding/hex"
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
	tokenRepo *repository.TokenRepository
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userRepo:  repository.NewUserRepository(),
		tokenRepo: repository.NewTokenRepository(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.UserRegistrationInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Create user
	user := &models.User{
		NimNip:   input.NimNip,
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
		UserType: input.UserType,
		Major:    input.Major,
		Faculty:  input.Faculty,
		Position: input.Position,
	}

	// Save user to database
	err := h.userRepo.CreateUser(user)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			utils.BadRequestResponse(c, "User with this NIM/NIP or email already exists", nil)
		} else {
			utils.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	// Return success response
	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", user.ToUserResponse())
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.UserLoginInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Find user by NIM/NIP
	user, err := h.userRepo.GetUserByNimNip(input.NimNip)
	if err != nil {
		if err == repository.ErrUserNotFound {
			utils.UnauthorizedResponse(c, "Invalid credentials")
		} else {
			utils.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	// Verify password
	if !user.ComparePassword(input.Password) {
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Generate JWT token
	tokenString, expiryTime, err := jwt.GenerateAccessToken(user.ID, user.NimNip, user.Name, user.Email)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate access token")
		return
	}

	// Generate refresh token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate refresh token")
		return
	}

	// Parse refresh token expiry from environment
	refreshExpiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if refreshExpiryStr == "" {
		refreshExpiryStr = "168h" // Default to 7 days
	}

	refreshExpiry, err := time.ParseDuration(refreshExpiryStr)
	if err != nil {
		refreshExpiry = time.Hour * 24 * 7 // 7 days
	}

	refreshExpiryTime := time.Now().Add(refreshExpiry)

	// Save refresh token to database
	err = h.tokenRepo.CreateToken(user.ID, refreshToken, refreshExpiryTime)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save refresh token")
		return
	}

	// Clean up expired tokens
	go h.tokenRepo.DeleteExpiredTokens()

	// Calculate expiry in seconds
	expiresIn := int64(time.Until(expiryTime).Seconds())

	// Return tokens
	tokens := models.TokenPair{
		AccessToken:  tokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{
		"user":   user.ToUserResponse(),
		"tokens": tokens,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input models.RefreshTokenRequest

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Find token in database
	token, err := h.tokenRepo.GetTokenByValue(input.RefreshToken)
	if err != nil {
		if err == repository.ErrTokenNotFound || err == repository.ErrTokenExpired {
			utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		} else {
			utils.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	// Get user
	user, err := h.userRepo.GetUserByID(token.UserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			utils.UnauthorizedResponse(c, "User not found")
		} else {
			utils.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	// Generate new JWT token
	tokenString, expiryTime, err := jwt.GenerateAccessToken(user.ID, user.NimNip, user.Name, user.Email)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate access token")
		return
	}

	// Calculate expiry in seconds
	expiresIn := int64(time.Until(expiryTime).Seconds())

	// Return new access token
	tokens := gin.H{
		"access_token": tokenString,
		"expires_in":   expiresIn,
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", tokens)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var input models.RefreshTokenRequest

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		// Continue even if no refresh token is provided
		// We'll still mark the user as logged out
	}

	// If refresh token is provided, delete it
	if input.RefreshToken != "" {
		err := h.tokenRepo.DeleteToken(input.RefreshToken)
		if err != nil && err != repository.ErrTokenNotFound {
			utils.InternalServerErrorResponse(c, err.Error())
			return
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Logged out successfully", nil)
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "")
		return
	}

	user, err := h.userRepo.GetUserByID(userID.(uint))
	if err != nil {
		if err == repository.ErrUserNotFound {
			utils.UnauthorizedResponse(c, "User not found")
		} else {
			utils.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User details retrieved successfully", user.ToUserResponse())
}

// Helper function to generate a random refresh token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
