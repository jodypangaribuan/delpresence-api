package repository

import (
	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"
)

// AdminRepository handles admin database operations
type AdminRepository struct{}

// NewAdminRepository creates a new instance of AdminRepository
func NewAdminRepository() *AdminRepository {
	return &AdminRepository{}
}

// GetAdminByUserID retrieves an admin by user ID
func (r *AdminRepository) GetAdminByUserID(userID uint) (*models.Admin, error) {
	var admin models.Admin
	result := database.DB.Where("user_id = ?", userID).Preload("User").First(&admin)
	if result.Error != nil {
		return nil, result.Error
	}
	return &admin, nil
}

// GetAdminByEmail retrieves an admin by email
func (r *AdminRepository) GetAdminByEmail(email string) (*models.Admin, error) {
	var user models.User
	result := database.DB.Where("email = ? AND user_type = ?", email, models.AdminType).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return r.GetAdminByUserID(user.ID)
}

// CreateAdmin creates a new admin
func (r *AdminRepository) CreateAdmin(admin *models.Admin) error {
	return database.DB.Create(admin).Error
}

// CheckAdminExists checks if an admin with the given email exists
func (r *AdminRepository) CheckAdminExists(email string) bool {
	var count int64
	database.DB.Model(&models.User{}).
		Where("email = ? AND user_type = ?", email, models.AdminType).
		Count(&count)
	return count > 0
}
