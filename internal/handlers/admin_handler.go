package handlers

import (
	"net/http"

	"delpresence-api/internal/models"
	"delpresence-api/internal/repository"
	"delpresence-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// AdminHandler menangani request terkait admin
type AdminHandler struct {
	adminRepo *repository.AdminRepository
}

// NewAdminHandler membuat instance AdminHandler baru
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		adminRepo: repository.NewAdminRepository(),
	}
}

// Login menangani proses login admin
func (h *AdminHandler) Login(c *gin.Context) {
	var request models.AdminLoginRequest

	// Binding request JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.BadRequestResponse(c, "Format request tidak valid")
		return
	}

	// Validasi input
	if request.Username == "" || request.Password == "" {
		utils.BadRequestResponse(c, "Username dan password wajib diisi")
		return
	}

	// Dapatkan IP client
	clientIP := c.ClientIP()

	// Proses login
	response, err := h.adminRepo.LoginAdmin(request.Username, request.Password, clientIP)
	if err != nil {
		utils.UnauthorizedResponse(c, err.Error())
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// GetAdminProfile mengembalikan profil lengkap admin
func (h *AdminHandler) GetAdminProfile(c *gin.Context) {
	// Ambil user_id dari token JWT (via middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User tidak terautentikasi")
		return
	}

	// Convert ke uint
	userIDUint, ok := userID.(uint)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format")
		return
	}

	// Dapatkan profil admin
	adminWithUser, err := h.adminRepo.GetAdminByUserID(userIDUint)
	if err != nil {
		utils.NotFoundResponse(c, "Profil admin tidak ditemukan")
		return
	}

	// Convert ke response format
	response := adminWithUser.Admin.ToAdminResponse(adminWithUser.User)

	utils.SuccessResponse(c, http.StatusOK, "Profil admin berhasil diambil", response)
}
