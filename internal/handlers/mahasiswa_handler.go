package handlers

import (
	"delpresence-api/internal/models"
	"delpresence-api/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// MahasiswaHandler handles student-related requests
type MahasiswaHandler struct {
	campusClient *utils.CampusClient
}

// NewMahasiswaHandler creates a new MahasiswaHandler
func NewMahasiswaHandler() *MahasiswaHandler {
	return &MahasiswaHandler{
		campusClient: utils.NewCampusClient(),
	}
}

// GetMahasiswaByUserID fetches student information by user ID
func (h *MahasiswaHandler) GetMahasiswaByUserID(c *gin.Context) {
	// Parse user ID from query parameter
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID is required",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID format",
		})
		return
	}

	// Fetch student information from the campus API
	mahasiswaInfo, err := h.campusClient.GetMahasiswaByUserID(userID)
	if err != nil {
		// Check if this is a "no student found" error
		if strings.Contains(err.Error(), "no student found") {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("No student found with user ID: %d", userID),
			})
			return
		}
		// For other errors, return 500
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch student information: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   mahasiswaInfo,
	})
}

// GetMahasiswaDetailByNIM fetches detailed student information by NIM
func (h *MahasiswaHandler) GetMahasiswaDetailByNIM(c *gin.Context) {
	// Parse NIM from query parameter
	nim := c.Query("nim")
	if nim == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "NIM is required",
		})
		return
	}

	// Fetch detailed student information from the campus API
	mahasiswaDetail, err := h.campusClient.GetMahasiswaDetailByNIM(nim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch student details: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   mahasiswaDetail,
	})
}

// GetMahasiswaComplete fetches complete student information by user ID
// This is a convenience method that fetches both basic info and details
func (h *MahasiswaHandler) GetMahasiswaComplete(c *gin.Context) {
	// Get user ID from context if set by middleware
	userIDFromContext, exists := c.Get("user_id")

	var userID int
	if exists {
		// Use the ID from the authenticated token
		userID = int(userIDFromContext.(uint))
		log.Printf("Using user ID from token: %d", userID)
	} else {
		// Parse user ID from query parameter as fallback
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			log.Printf("Error: No user_id provided in request")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "User ID is required",
			})
			return
		}

		var err error
		userID, err = strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Error: Invalid user_id format: %s", userIDStr)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid user ID format",
			})
			return
		}
		log.Printf("Using user ID from query parameter: %d", userID)
	}

	// Check if this is a campus-authenticated request
	campusAuth, campusAuthExists := c.Get("campus_authenticated")
	isCampusAuth := campusAuthExists && campusAuth.(bool)

	log.Printf("Processing complete student data request for user ID: %d (campus auth: %v)", userID, isCampusAuth)

	// Step 1: Fetch basic student information to get the NIM
	log.Printf("Fetching basic student info for user ID: %d", userID)
	mahasiswaInfo, err := h.campusClient.GetMahasiswaByUserID(userID)
	if err != nil {
		log.Printf("Error fetching student info: %v", err)
		// Check if this is a "no student found" error
		if strings.Contains(err.Error(), "no student found") {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("No student found with user ID: %d", userID),
			})
			return
		}
		// For other errors, return 500
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch student information: " + err.Error(),
		})
		return
	}

	log.Printf("Found student with NIM: %s, Name: %s", mahasiswaInfo.Nim, mahasiswaInfo.Nama)

	// Step 2: Fetch detailed student information using the NIM
	log.Printf("Fetching detailed student info for NIM: %s", mahasiswaInfo.Nim)
	mahasiswaDetail, err := h.campusClient.GetMahasiswaDetailByNIM(mahasiswaInfo.Nim)
	if err != nil {
		log.Printf("Error fetching student details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch student details: " + err.Error(),
		})
		return
	}

	log.Printf("Successfully retrieved details for student: %s", mahasiswaDetail.Nama)

	// Step 3: Combine the information into a response
	response := models.MahasiswaComplete{
		BasicInfo: *mahasiswaInfo,
		Details:   *mahasiswaDetail,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}
