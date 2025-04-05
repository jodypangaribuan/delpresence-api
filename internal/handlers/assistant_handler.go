package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"delpresence-api/internal/models"
	"delpresence-api/internal/repository"
	"delpresence-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// AssistantHandler menangani request terkait asisten dosen
type AssistantHandler struct {
	assistantRepo repository.AssistantRepository
	campusClient  *utils.CampusClient
}

// NewAssistantHandler membuat instance baru AssistantHandler
func NewAssistantHandler(assistantRepo repository.AssistantRepository) *AssistantHandler {
	return &AssistantHandler{
		assistantRepo: assistantRepo,
		campusClient:  utils.NewCampusClient(),
	}
}

// GetAssistantProfile mengembalikan detail profil asisten dosen
func (h *AssistantHandler) GetAssistantProfile(c *gin.Context) {
	// Get user ID from JWT claim
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Find assistant profile by user ID
	assistant, err := h.assistantRepo.FindByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch assistant profile",
		})
		return
	}

	// If assistant profile doesn't exist, try to fetch from campus API
	if assistant == nil {
		// First check if campus_user_id is in the context (from JWT)
		campusUserID, exists := c.Get("campus_user_id")

		// If not in context, try to get from query parameter
		if !exists {
			campusUserIDStr := c.Query("campus_user_id")
			if campusUserIDStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Campus user ID not found. Please provide it as a query parameter.",
				})
				return
			}

			// Parse campusUserID from string
			var parseErr error
			var campusUserIDInt int
			campusUserIDInt, parseErr = strconv.Atoi(campusUserIDStr)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid campus user ID format",
				})
				return
			}
			campusUserID = campusUserIDInt
		}

		// Log campus user ID type and value for debugging
		log.Printf("Campus User ID: %v (type: %T)", campusUserID, campusUserID)

		// Fetch assistant details from campus API
		var campusUserIDInt int
		switch v := campusUserID.(type) {
		case int:
			campusUserIDInt = v
		case float64:
			campusUserIDInt = int(v)
		default:
			log.Printf("Unexpected campus user ID type: %T", campusUserID)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid campus user ID type",
			})
			return
		}

		newAssistant, err := h.fetchAssistantDetails(campusUserIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch assistant details from campus API: %v", err),
			})
			return
		}

		// Set user ID and save to database
		newAssistant.AssistantUserID = userID.(uint)
		newAssistant.LastSyncAt = time.Now()
		if err := h.assistantRepo.Create(newAssistant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save assistant details",
			})
			return
		}

		assistant = newAssistant
	}

	c.JSON(http.StatusOK, gin.H{
		"assistant": gin.H{
			"editable_fields": assistant.GetEditableFields(),
			"readonly_fields": assistant.GetReadOnlyFields(),
			"id":              assistant.ID,
			"user_id":         assistant.CampusUserID,
			"last_sync_at":    assistant.LastSyncAt,
		},
	})
}

// SyncAssistantProfile memperbarui data asisten dosen dari API kampus
func (h *AssistantHandler) SyncAssistantProfile(c *gin.Context) {
	// Get user ID from JWT claim
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get campus user ID from JWT claim
	campusUserID, exists := c.Get("campus_user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Campus user ID not found",
		})
		return
	}

	// Find existing assistant profile
	existingAssistant, err := h.assistantRepo.FindByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch assistant profile",
		})
		return
	}

	// Fetch updated assistant details from campus API
	var campusUserIDInt int
	switch v := campusUserID.(type) {
	case int:
		campusUserIDInt = v
	case float64:
		campusUserIDInt = int(v)
	default:
		log.Printf("Unexpected campus user ID type in Sync: %T", campusUserID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid campus user ID type",
		})
		return
	}

	updatedAssistant, err := h.fetchAssistantDetails(campusUserIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch assistant details from campus API: %v", err),
		})
		return
	}

	// Update or create assistant record
	if existingAssistant != nil {
		// Preserve user-editable fields
		updatedAssistant.ID = existingAssistant.ID
		updatedAssistant.AssistantUserID = existingAssistant.AssistantUserID
		updatedAssistant.Avatar = existingAssistant.Avatar
		updatedAssistant.Biography = existingAssistant.Biography
		updatedAssistant.PhoneNumber = existingAssistant.PhoneNumber
		updatedAssistant.Address = existingAssistant.Address
		updatedAssistant.LastSyncAt = time.Now()

		if err := h.assistantRepo.Update(updatedAssistant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update assistant details",
			})
			return
		}
	} else {
		// Create new assistant record
		updatedAssistant.AssistantUserID = userID.(uint)
		updatedAssistant.LastSyncAt = time.Now()
		if err := h.assistantRepo.Create(updatedAssistant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create assistant record",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Assistant profile synchronized successfully",
		"assistant": gin.H{
			"editable_fields": updatedAssistant.GetEditableFields(),
			"readonly_fields": updatedAssistant.GetReadOnlyFields(),
			"id":              updatedAssistant.ID,
			"user_id":         updatedAssistant.CampusUserID,
			"last_sync_at":    updatedAssistant.LastSyncAt,
		},
	})
}

// UpdateAssistantProfile memperbarui informasi profil asisten dosen yang dapat diubah
func (h *AssistantHandler) UpdateAssistantProfile(c *gin.Context) {
	// Get user ID from JWT claim
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Find assistant by user ID
	assistant, err := h.assistantRepo.FindByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch assistant profile",
		})
		return
	}

	if assistant == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Assistant profile not found",
		})
		return
	}

	// Parse request body
	var req struct {
		Avatar      *string `json:"avatar"`
		Biography   *string `json:"biography"`
		PhoneNumber *string `json:"phone_number"`
		Address     *string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update fields if provided
	if req.Avatar != nil {
		assistant.Avatar = *req.Avatar
	}
	if req.Biography != nil {
		assistant.Biography = *req.Biography
	}
	if req.PhoneNumber != nil {
		assistant.PhoneNumber = *req.PhoneNumber
	}
	if req.Address != nil {
		assistant.Address = *req.Address
	}

	// Save changes
	if err := h.assistantRepo.Update(assistant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update assistant profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Assistant profile updated successfully",
		"assistant": gin.H{
			"editable_fields": assistant.GetEditableFields(),
			"readonly_fields": assistant.GetReadOnlyFields(),
		},
	})
}

// fetchAssistantDetails retrieves assistant details from the campus API
func (h *AssistantHandler) fetchAssistantDetails(campusUserID int) (*models.Assistant, error) {
	url := fmt.Sprintf("https://cis.del.ac.id/api/library-api/pegawai?userid=%d", campusUserID)

	log.Printf("Fetching assistant details for campus user ID: %d from URL: %s", campusUserID, url)

	// Use campus client to make authenticated request
	response, err := h.campusClient.GetWithAuth(url)
	if err != nil {
		log.Printf("Error fetching assistant details: %v", err)
		return nil, fmt.Errorf("error fetching assistant details: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("campus API returned status: %d", response.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var campusResp models.CampusAssistantResponse
	if err := json.Unmarshal(body, &campusResp); err != nil {
		return nil, err
	}

	// Check if data is valid and contains assistant details
	if campusResp.Result != "Ok" || len(campusResp.Data.Pegawai) == 0 {
		return nil, fmt.Errorf("invalid or empty response from campus API")
	}

	// Extract assistant data
	pegawaiData := campusResp.Data.Pegawai[0]

	// Mendapatkan informasi jabatan (opsional)
	jobTitle := "Asisten Dosen" // Default job title
	var strukturID uint = 0

	// Mencoba mendapatkan informasi jabatan dari respons login yang disimpan di memory/session
	// Dalam implementasi nyata, ini bisa dilakukan dengan menyimpan data jabatan saat login
	// dan mengambilnya dari database atau session

	// Create new assistant model
	assistant := &models.Assistant{
		CampusUserID:   uint(campusUserID),
		EmployeeID:     pegawaiData.PegawaiID,
		IdentityNumber: pegawaiData.NIP,
		FullName:       pegawaiData.Nama,
		Email:          pegawaiData.Email,
		Department:     "Institut Teknologi Del", // Default untuk Institusi
		JobTitle:       jobTitle,                 // Dari respons login
		StrukturID:     strukturID,               // Dari respons login
		Username:       pegawaiData.UserName,
		Alias:          strings.TrimSpace(pegawaiData.Alias),  // Menghapus spasi dari respons API
		Position:       strings.TrimSpace(pegawaiData.Posisi), // Menghapus spasi dari respons API
		EmployeeStatus: pegawaiData.StatusPegawai,
		LastSyncAt:     time.Now(),
	}

	return assistant, nil
}
