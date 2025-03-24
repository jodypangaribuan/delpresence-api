package handlers

import (
	"crypto/rand"
	"encoding/hex"
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
	userRepo    *repository.UserRepository
	studentRepo *repository.StudentRepository
	lectureRepo *repository.LectureRepository
	adminRepo   *repository.AdminRepository
	tokenRepo   *repository.TokenRepository
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userRepo:    repository.NewUserRepository(),
		studentRepo: repository.NewStudentRepository(),
		lectureRepo: repository.NewLectureRepository(),
		adminRepo:   repository.NewAdminRepository(),
		tokenRepo:   repository.NewTokenRepository(),
	}
}

// RegisterStudent handles student registration
func (h *AuthHandler) RegisterStudent(c *gin.Context) {
	var input models.StudentRegistrationInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Start a database transaction
	tx := repository.BeginTransaction()
	if tx == nil {
		utils.InternalServerErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// 1. Create user
	user := &models.User{
		FirstName:  input.FirstName,
		MiddleName: input.MiddleName,
		LastName:   input.LastName,
		Email:      input.Email,
		Password:   input.Password,
		UserType:   models.StudentType,
	}

	if err := tx.Create(user).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create user")
		return
	}

	// 2. Create student
	student := &models.Student{
		UserID:  user.ID,
		NIM:     input.NIM,
		Major:   input.Major,
		Faculty: input.Faculty,
		Batch:   input.Batch,
	}

	if err := tx.Create(student).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create student profile")
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to complete registration")
		return
	}

	// Return success response
	student.User = *user
	utils.SuccessResponse(c, http.StatusCreated, "Student registered successfully", student.ToStudentResponse())
}

// RegisterLecture handles lecturer registration
func (h *AuthHandler) RegisterLecture(c *gin.Context) {
	var input models.LectureRegistrationInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Start a database transaction
	tx := repository.BeginTransaction()
	if tx == nil {
		utils.InternalServerErrorResponse(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// 1. Create user
	user := &models.User{
		FirstName:  input.FirstName,
		MiddleName: input.MiddleName,
		LastName:   input.LastName,
		Email:      input.Email,
		Password:   input.Password,
		UserType:   models.LectureType,
	}

	if err := tx.Create(user).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create user")
		return
	}

	// 2. Create lecture
	lecture := &models.Lecture{
		UserID:   user.ID,
		NIP:      input.NIP,
		Position: input.Position,
	}

	if err := tx.Create(lecture).Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create lecturer profile")
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		utils.InternalServerErrorResponse(c, "Failed to complete registration")
		return
	}

	// Return success response
	lecture.User = *user
	utils.SuccessResponse(c, http.StatusCreated, "Lecturer registered successfully", lecture.ToLectureResponse())
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.UserLoginInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Find user by login ID (NIM or NIP)
	var user *models.User
	var err error
	var userType string
	var userResponse interface{}

	log.Printf("Login attempt with loginID: %s", input.LoginID)

	// Check if it's a student (lookup by NIM)
	student, err := h.studentRepo.GetStudentByNIM(input.LoginID)
	if err == nil {
		user = &student.User
		userType = "student"
		userResponse = student.ToStudentResponse()
		log.Printf("Successfully found student with NIM: %s", input.LoginID)
	} else {
		log.Printf("Not a student (error: %v), checking if it's a lecturer", err)

		// Check if it's a lecture (lookup by NIP)
		lecture, err := h.lectureRepo.GetLectureByNIP(input.LoginID)
		if err == nil {
			user = &lecture.User
			userType = "lecture"
			userResponse = lecture.ToLectureResponse()
			log.Printf("Successfully found lecturer with NIP: %s", input.LoginID)
		} else {
			log.Printf("Not a lecturer either (error: %v)", err)
			utils.UnauthorizedResponse(c, "Invalid credentials")
			return
		}
	}

	// Verify password
	if !user.ComparePassword(input.Password) {
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
	refreshToken, err := generateRefreshToken()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate refresh token")
		return
	}

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
	err = h.tokenRepo.CreateToken(user.ID, refreshToken, refreshTokenExpiry)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save refresh token")
		return
	}

	// Create response
	response := map[string]interface{}{
		"user": userResponse,
		"tokens": map[string]interface{}{
			"access_token":  tokenString,
			"refresh_token": refreshToken,
			"expires_in":    int(time.Until(expiryTime).Seconds()),
			"token_type":    "Bearer",
		},
		"user_type": userType,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
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
	refreshToken, err := generateRefreshToken()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate refresh token")
		return
	}

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
	err = h.tokenRepo.CreateToken(user.ID, refreshToken, refreshTokenExpiry)
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

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input models.RefreshTokenRequest

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Get refresh token from database
	token, err := h.tokenRepo.GetTokenByValue(input.RefreshToken)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid refresh token")
		return
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		// Delete expired token
		_ = h.tokenRepo.DeleteToken(token.Token)
		utils.UnauthorizedResponse(c, "Refresh token expired")
		return
	}

	// Get user
	user, err := h.userRepo.GetUserByID(token.UserID)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not found")
		return
	}

	// Get user profile based on user type
	var userResponse interface{}
	var userType string

	if user.UserType == models.StudentType {
		student, err := h.studentRepo.GetStudentByUserID(user.ID)
		if err != nil {
			utils.UnauthorizedResponse(c, "Student profile not found")
			return
		}
		student.User = *user
		userResponse = student.ToStudentResponse()
		userType = "student"
	} else if user.UserType == models.LectureType {
		lecture, err := h.lectureRepo.GetLectureByUserID(user.ID)
		if err != nil {
			utils.UnauthorizedResponse(c, "Lecturer profile not found")
			return
		}
		lecture.User = *user
		userResponse = lecture.ToLectureResponse()
		userType = "lecture"
	} else if user.UserType == models.AdminType {
		admin, err := h.adminRepo.GetAdminByUserID(user.ID)
		if err != nil {
			utils.UnauthorizedResponse(c, "Admin profile not found")
			return
		}
		admin.User = *user
		userResponse = admin.ToAdminResponse()
		userType = "admin"
	} else {
		utils.UnauthorizedResponse(c, "Invalid user type")
		return
	}

	// Generate new access token
	tokenString, expiryTime, err := jwt.GenerateAccessToken(user.ID, "", user.FirstName, user.MiddleName, user.LastName, user.Email)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate access token")
		return
	}

	// Generate new refresh token
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate refresh token")
		return
	}

	// Delete old refresh token
	err = h.tokenRepo.DeleteToken(token.Token)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete old refresh token")
		return
	}

	// Parse refresh token expiry from environment
	refreshExpiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if refreshExpiryStr == "" {
		refreshExpiryStr = "168h" // Default 7 days
	}

	refreshExpiry, err := time.ParseDuration(refreshExpiryStr)
	if err != nil {
		refreshExpiry = 168 * time.Hour // Default 7 days
	}

	// Save new refresh token to database
	refreshTokenExpiry := time.Now().Add(refreshExpiry)
	err = h.tokenRepo.CreateToken(user.ID, newRefreshToken, refreshTokenExpiry)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save refresh token")
		return
	}

	// Create response
	response := map[string]interface{}{
		"user": userResponse,
		"tokens": map[string]interface{}{
			"access_token":  tokenString,
			"refresh_token": newRefreshToken,
			"expires_in":    int(time.Until(expiryTime).Seconds()),
			"token_type":    "Bearer",
		},
		"user_type": userType,
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var input models.RefreshTokenRequest

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Delete refresh token
	err := h.tokenRepo.DeleteToken(input.RefreshToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to logout")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

// GetCurrentUser handles getting the current user's information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get user
	user, err := h.userRepo.GetUserByID(userID.(uint))
	if err != nil {
		utils.UnauthorizedResponse(c, "User not found")
		return
	}

	// Get user profile based on user type
	var userResponse interface{}
	var userType string

	if user.UserType == models.StudentType {
		student, err := h.studentRepo.GetStudentByUserID(user.ID)
		if err != nil {
			utils.UnauthorizedResponse(c, "Student profile not found")
			return
		}
		student.User = *user
		userResponse = student.ToStudentResponse()
		userType = "student"
	} else if user.UserType == models.LectureType {
		lecture, err := h.lectureRepo.GetLectureByUserID(user.ID)
		if err != nil {
			utils.UnauthorizedResponse(c, "Lecturer profile not found")
			return
		}
		lecture.User = *user
		userResponse = lecture.ToLectureResponse()
		userType = "lecture"
	} else if user.UserType == models.AdminType {
		admin, err := h.adminRepo.GetAdminByUserID(user.ID)
		if err != nil {
			utils.UnauthorizedResponse(c, "Admin profile not found")
			return
		}
		admin.User = *user
		userResponse = admin.ToAdminResponse()
		userType = "admin"
	} else {
		utils.UnauthorizedResponse(c, "Invalid user type")
		return
	}

	// Create response
	response := map[string]interface{}{
		"user":      userResponse,
		"user_type": userType,
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", response)
}

// generateRefreshToken generates a random refresh token
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
