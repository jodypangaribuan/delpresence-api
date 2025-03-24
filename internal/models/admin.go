package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin represents an admin user in the system
type Admin struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AdminResponse represents the admin data returned in API responses
type AdminResponse struct {
	ID     uint         `json:"id"`
	UserID uint         `json:"user_id"`
	User   UserResponse `json:"user"`
}

// ToAdminResponse converts an Admin to AdminResponse
func (a *Admin) ToAdminResponse() AdminResponse {
	return AdminResponse{
		ID:     a.ID,
		UserID: a.UserID,
		User:   a.User.ToUserResponse(),
	}
}

// AdminLoginInput represents input data for admin login
type AdminLoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
