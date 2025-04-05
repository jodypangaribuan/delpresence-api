package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"delpresence-api/internal/models"
	"delpresence-api/internal/repository"
	"delpresence-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// LecturerHandler menangani request terkait dosen
type LecturerHandler struct {
	lecturerRepo repository.LecturerRepository
	campusClient *utils.CampusClient
}

// NewLecturerHandler membuat instance baru LecturerHandler
func NewLecturerHandler(lecturerRepo repository.LecturerRepository) *LecturerHandler {
	return &LecturerHandler{
		lecturerRepo: lecturerRepo,
		campusClient: utils.NewCampusClient(),
	}
}

// GetLecturerProfile mengembalikan detail profil dosen
func (h *LecturerHandler) GetLecturerProfile(c *gin.Context) {
	// Get user ID from JWT claim
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Find lecturer profile by user ID
	lecturer, err := h.lecturerRepo.FindByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch lecturer profile",
		})
		return
	}

	// If lecturer profile doesn't exist, try to fetch from campus API
	if lecturer == nil {
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

		// Fetch lecturer details from campus API
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

		newLecturer, err := h.fetchLecturerDetails(campusUserIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch lecturer details from campus API: %v", err),
			})
			return
		}

		// Set user ID and save to database
		newLecturer.LecturerUserID = userID.(uint)
		newLecturer.LastSyncAt = time.Now()
		if err := h.lecturerRepo.Create(newLecturer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save lecturer details",
			})
			return
		}

		lecturer = newLecturer
	}

	c.JSON(http.StatusOK, gin.H{
		"lecturer": gin.H{
			"editable_fields": lecturer.GetEditableFields(),
			"readonly_fields": lecturer.GetReadOnlyFields(),
			"id":              lecturer.ID,
			"user_id":         lecturer.CampusUserID,
			"last_sync_at":    lecturer.LastSyncAt,
		},
	})
}

// SyncLecturerProfile memperbarui data dosen dari API kampus
func (h *LecturerHandler) SyncLecturerProfile(c *gin.Context) {
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

	// Find existing lecturer profile
	existingLecturer, err := h.lecturerRepo.FindByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch lecturer profile",
		})
		return
	}

	// Fetch updated lecturer details from campus API
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

	updatedLecturer, err := h.fetchLecturerDetails(campusUserIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch lecturer details from campus API: %v", err),
		})
		return
	}

	// Update or create lecturer record
	if existingLecturer != nil {
		// Preserve user-editable fields
		updatedLecturer.ID = existingLecturer.ID
		updatedLecturer.LecturerUserID = existingLecturer.LecturerUserID
		updatedLecturer.Avatar = existingLecturer.Avatar
		updatedLecturer.Biography = existingLecturer.Biography
		updatedLecturer.Publications = existingLecturer.Publications
		updatedLecturer.PhoneNumber = existingLecturer.PhoneNumber
		updatedLecturer.Address = existingLecturer.Address
		updatedLecturer.LastSyncAt = time.Now()

		if err := h.lecturerRepo.Update(updatedLecturer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update lecturer details",
			})
			return
		}
	} else {
		// Create a new lecturer in the database
		newLecturer := &models.Lecturer{
			LecturerUserID:   uint(userID.(float64)),
			CampusUserID:     uint(campusUserIDInt),
			EmployeeID:       updatedLecturer.EmployeeID,
			LecturerID:       updatedLecturer.LecturerID,
			IdentityNumber:   updatedLecturer.IdentityNumber,
			LecturerNumber:   updatedLecturer.LecturerNumber,
			FullName:         updatedLecturer.FullName,
			Email:            updatedLecturer.Email,
			DepartmentID:     updatedLecturer.DepartmentID,
			Department:       updatedLecturer.Department,
			AcademicRank:     updatedLecturer.AcademicRank,
			AcademicRankDesc: updatedLecturer.AcademicRankDesc,
			EducationLevel:   updatedLecturer.EducationLevel,
			Status:           "Active",
			LastSyncAt:       time.Now(),
		}

		if err := h.lecturerRepo.Create(newLecturer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save lecturer details",
			})
			return
		}

		updatedLecturer = newLecturer
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lecturer profile synced successfully",
		"lecturer": gin.H{
			"editable_fields": updatedLecturer.GetEditableFields(),
			"readonly_fields": updatedLecturer.GetReadOnlyFields(),
			"id":              updatedLecturer.ID,
			"user_id":         updatedLecturer.CampusUserID,
			"last_sync_at":    updatedLecturer.LastSyncAt,
		},
	})
}

// UpdateLecturerProfile memperbarui bagian profil dosen yang dapat diubah pengguna
func (h *LecturerHandler) UpdateLecturerProfile(c *gin.Context) {
	// Get user ID from JWT claim
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Find existing lecturer profile
	lecturer, err := h.lecturerRepo.FindByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch lecturer profile",
		})
		return
	}

	if lecturer == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Lecturer profile not found",
		})
		return
	}

	// Parse update request
	var req struct {
		Avatar       *string `json:"avatar"`
		Biography    *string `json:"biography"`
		Publications *string `json:"publications"`
		PhoneNumber  *string `json:"phone_number"`
		Address      *string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Only update fields that are user-editable
	if req.Avatar != nil && lecturer.IsUserEditable("avatar") {
		lecturer.Avatar = *req.Avatar
	}
	if req.Biography != nil && lecturer.IsUserEditable("biography") {
		lecturer.Biography = *req.Biography
	}
	if req.Publications != nil && lecturer.IsUserEditable("publications") {
		lecturer.Publications = *req.Publications
	}
	if req.PhoneNumber != nil && lecturer.IsUserEditable("phone_number") {
		lecturer.PhoneNumber = *req.PhoneNumber
	}
	if req.Address != nil && lecturer.IsUserEditable("address") {
		lecturer.Address = *req.Address
	}

	// Save changes
	if err := h.lecturerRepo.Update(lecturer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update lecturer profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lecturer profile updated successfully",
		"lecturer": gin.H{
			"editable_fields": lecturer.GetEditableFields(),
			"readonly_fields": lecturer.GetReadOnlyFields(),
		},
	})
}

// fetchLecturerDetails retrieves lecturer details from the campus API
func (h *LecturerHandler) fetchLecturerDetails(campusUserID int) (*models.Lecturer, error) {
	url := fmt.Sprintf("https://cis.del.ac.id/api/library-api/dosen?userid=%d", campusUserID)

	log.Printf("Fetching lecturer details for campus user ID: %d from URL: %s", campusUserID, url)

	// Use campus client to make authenticated request
	response, err := h.campusClient.GetWithAuth(url)
	if err != nil {
		log.Printf("Error fetching lecturer details: %v", err)
		return nil, fmt.Errorf("error fetching lecturer details: %w", err)
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

	var campusResp models.CampusLecturerResponse
	if err := json.Unmarshal(body, &campusResp); err != nil {
		return nil, err
	}

	// Check if data is valid and contains lecturer details
	if campusResp.Result != "Ok" || len(campusResp.Data.Dosen) == 0 {
		return nil, fmt.Errorf("invalid or empty response from campus API")
	}

	// Extract lecturer data
	dosenData := campusResp.Data.Dosen[0]

	// Create new lecturer model with correct field names
	lecturer := &models.Lecturer{
		CampusUserID:     uint(campusUserID),
		EmployeeID:       dosenData.PegawaiID,
		LecturerID:       dosenData.DosenID,
		IdentityNumber:   dosenData.NIP,
		LecturerNumber:   dosenData.NIDN,
		FullName:         dosenData.Nama,
		Email:            dosenData.Email,
		DepartmentID:     dosenData.ProdiID,
		Department:       dosenData.Prodi,
		AcademicRank:     dosenData.JabatanAkademik,
		AcademicRankDesc: dosenData.JabatanAkademikDesc,
		EducationLevel:   dosenData.JenjangPendidikan,
		Status:           "Active", // Default status
		LastSyncAt:       time.Now(),
	}

	return lecturer, nil
}
