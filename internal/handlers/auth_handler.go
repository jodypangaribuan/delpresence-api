package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"delpresence-api/internal/repository"
	"delpresence-api/internal/utils"

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

// CampusLoginResponse represents the response from campus auth API
type CampusLoginResponse struct {
	Result       bool       `json:"result"`
	Error        string     `json:"error"`
	Success      string     `json:"success"`
	User         CampusUser `json:"user"`
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
}

// CampusUser represents the user data from campus auth API
type CampusUser struct {
	UserID   int             `json:"user_id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Role     string          `json:"role"`
	Status   int             `json:"status"`
	Jabatan  []CampusJabatan `json:"jabatan"`
}

// CampusJabatan represents the position of a campus user
type CampusJabatan struct {
	StrukturJabatanID int    `json:"struktur_jabatan_id"`
	Jabatan           string `json:"jabatan"`
}

// CampusLogin handles login through campus authentication system
func (h *AuthHandler) CampusLogin(c *gin.Context) {
	// Get username and password from form data
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		utils.BadRequestResponse(c, "Username and password are required")
		return
	}

	// Create form data for the campus API
	formData := url.Values{}
	formData.Add("username", username)
	formData.Add("password", password)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new request to the campus API
	req, err := http.NewRequest("POST", "https://cis.del.ac.id/api/jwt-api/do-auth",
		strings.NewReader(formData.Encode()))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create request")
		return
	}

	// Set required headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://cis.del.ac.id")
	req.Header.Add("Referer", "https://cis.del.ac.id")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		utils.InternalServerErrorResponse(c, fmt.Sprintf("Failed to reach campus API: %v", err))
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to read response from campus API")
		return
	}

	// Check if we got a valid JSON response
	var campusResponse CampusLoginResponse
	if err := json.Unmarshal(body, &campusResponse); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to parse response from campus API")
		return
	}

	// Return the response directly to the client
	if campusResponse.Result {
		// Successful login
		c.JSON(http.StatusOK, campusResponse)
	} else {
		// Failed login
		c.JSON(http.StatusUnauthorized, campusResponse)
	}
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

	userResponse := map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"first_name":  user.FirstName,
		"middle_name": user.MiddleName,
		"last_name":   user.LastName,
		"user_type":   user.UserType,
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
