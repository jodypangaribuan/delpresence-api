package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"delpresence-api/internal/models"
	"delpresence-api/internal/services"
	"delpresence-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequestResponse(c, "Token is required", nil)
		return
	}

	// Get token from database
	storedToken, err := h.tokenRepo.GetTokenByValue(token, models.VerificationToken)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid or expired verification token")
		return
	}

	// Get user
	user, err := h.userRepo.GetUserByID(storedToken.UserID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	// Mark user as verified
	user.Verified = true
	if err := h.userRepo.UpdateUser(user); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to verify email")
		return
	}

	// Delete the verification token
	h.tokenRepo.DeleteToken(token)

	// Get frontend URL from environment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Redirect to frontend login page with verified=true parameter
	redirectURL := frontendURL + "/login?verified=true"
	c.Redirect(http.StatusFound, redirectURL)
}

// ForgotPassword handles password reset requests
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Find user by email
	user, err := h.userRepo.GetUserByEmail(input.Email)
	if err != nil {
		// Don't reveal if the email exists or not for security reasons
		utils.SuccessResponse(c, http.StatusOK, "If your email is registered, you will receive a password reset link", nil)
		return
	}

	// Is this an admin?
	isAdmin := user.UserType == models.AdminType

	// Generate and send password reset email
	go h.sendPasswordResetEmail(user, isAdmin)

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "If your email is registered, you will receive a password reset link", nil)
}

// AdminForgotPassword handles password reset requests for admins
func (h *AuthHandler) AdminForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Find admin by email
	user, err := h.userRepo.GetUserByEmail(input.Email)
	if err != nil || user.UserType != models.AdminType {
		// Don't reveal if the email exists or not for security reasons
		utils.SuccessResponse(c, http.StatusOK, "If your email is registered, you will receive a password reset link", nil)
		return
	}

	// Generate and send password reset email
	go h.sendPasswordResetEmail(user, true)

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "If your email is registered, you will receive a password reset link", nil)
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequestResponse(c, "Token is required", nil)
		return
	}

	var input struct {
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Get token from database
	storedToken, err := h.tokenRepo.GetTokenByValue(token, models.PasswordResetToken)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid or expired reset token")
		return
	}

	// Get user
	user, err := h.userRepo.GetUserByID(storedToken.UserID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	// Update password
	user.Password = input.Password
	if err := h.userRepo.UpdateUser(user); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to reset password")
		return
	}

	// Delete the reset token
	h.tokenRepo.DeleteToken(token)

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "Password reset successfully", nil)
}

// ResendVerification handles resending verification email
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Find user by email
	user, err := h.userRepo.GetUserByEmail(input.Email)
	if err != nil {
		// Don't reveal if the email exists or not for security reasons
		utils.SuccessResponse(c, http.StatusOK, "Jika email Anda terdaftar, Anda akan menerima email verifikasi baru", nil)
		return
	}

	// Check if user is already verified
	if user.Verified {
		utils.SuccessResponse(c, http.StatusOK, "Akun ini sudah terverifikasi. Silakan login", nil)
		return
	}

	// Delete any existing verification tokens for this user
	h.tokenRepo.DeleteUserTokensByType(user.ID, models.VerificationToken)

	// Generate and send a new verification email
	go h.sendVerificationEmail(user)

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "Email verifikasi baru telah dikirim. Silakan periksa email Anda", nil)
}

// ValidateResetToken validates a password reset token
func (h *AuthHandler) ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequestResponse(c, "Token is required", nil)
		return
	}

	// Get token from database
	_, err := h.tokenRepo.GetTokenByValue(token, models.PasswordResetToken)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid or expired reset token")
		return
	}

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "Token is valid", nil)
}

// Helper functions for email verification and password reset

// sendVerificationEmail generates a verification token and sends an email
func (h *AuthHandler) sendVerificationEmail(user *models.User) {
	// Generate a verification token
	token, err := generateSecureToken()
	if err != nil {
		log.Printf("Failed to generate verification token: %v", err)
		return
	}

	// Store the token in the database
	expiry := time.Now().Add(24 * time.Hour) // 24 hour expiry
	err = h.tokenRepo.CreateToken(user.ID, token, models.VerificationToken, expiry)
	if err != nil {
		log.Printf("Failed to store verification token: %v", err)
		return
	}

	// Send the verification email
	emailService := services.NewEmailService()
	err = emailService.SendVerificationEmail(user.Email, user.FullName(), token)
	if err != nil {
		log.Printf("Failed to send verification email: %v", err)
	}
}

// sendPasswordResetEmail generates a password reset token and sends an email
func (h *AuthHandler) sendPasswordResetEmail(user *models.User, isAdmin bool) {
	// Generate a reset token
	token, err := generateSecureToken()
	if err != nil {
		log.Printf("Failed to generate password reset token: %v", err)
		return
	}

	// Delete any existing password reset tokens for this user
	h.tokenRepo.DeleteUserTokensByType(user.ID, models.PasswordResetToken)

	// Store the token in the database
	expiry := time.Now().Add(1 * time.Hour) // 1 hour expiry
	err = h.tokenRepo.CreateToken(user.ID, token, models.PasswordResetToken, expiry)
	if err != nil {
		log.Printf("Failed to store password reset token: %v", err)
		return
	}

	// Send the password reset email
	emailService := services.NewEmailService()
	err = emailService.SendPasswordResetEmail(user.Email, user.FullName(), token, isAdmin)
	if err != nil {
		log.Printf("Failed to send password reset email: %v", err)
	}
}

// generateSecureToken generates a secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
