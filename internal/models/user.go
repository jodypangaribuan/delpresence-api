package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserType represents the type of user
type UserType string

const (
	// StudentType represents a student user
	StudentType UserType = "student"
	// LectureType represents a lecturer user
	LectureType UserType = "lecture"
	// AdminType represents an admin user
	AdminType UserType = "admin"
)

// User represents the user model in the database
type User struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FirstName  string         `gorm:"not null" json:"first_name"`
	MiddleName string         `json:"middle_name"`
	LastName   string         `json:"last_name"`
	Email      string         `gorm:"unique;not null" json:"email"`
	Password   string         `gorm:"not null" json:"-"` // Password is not included in JSON responses
	UserType   UserType       `gorm:"not null;type:VARCHAR(20)" json:"user_type"`
	Verified   bool           `gorm:"default:false" json:"verified"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeSave hashes the password before saving to database
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// ComparePassword compares a hashed password with a plaintext password
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UserLoginInput represents input data for user login
type UserLoginInput struct {
	LoginID  string `json:"login_id" binding:"required"` // NIM or NIP
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID         uint     `json:"id"`
	FirstName  string   `json:"first_name"`
	MiddleName string   `json:"middle_name"`
	LastName   string   `json:"last_name"`
	Email      string   `json:"email"`
	UserType   UserType `json:"user_type"`
	Verified   bool     `json:"verified"`
}

// ToUserResponse converts a User to UserResponse
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:         u.ID,
		FirstName:  u.FirstName,
		MiddleName: u.MiddleName,
		LastName:   u.LastName,
		Email:      u.Email,
		UserType:   u.UserType,
		Verified:   u.Verified,
	}
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	fullName := u.FirstName
	if u.MiddleName != "" {
		fullName += " " + u.MiddleName
	}
	if u.LastName != "" {
		fullName += " " + u.LastName
	}
	return fullName
}
