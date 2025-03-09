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
	// StaffType represents a staff user
	StaffType UserType = "staff"
)

// User represents the user model in the database
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	NimNip    string         `gorm:"unique;not null" json:"nim_nip"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"` // Password is not included in JSON responses
	UserType  UserType       `gorm:"not null;type:VARCHAR(20)" json:"user_type"`
	Major     string         `gorm:"default:null" json:"major,omitempty"`
	Faculty   string         `gorm:"default:null" json:"faculty,omitempty"`
	Position  string         `gorm:"default:null" json:"position,omitempty"`
	Verified  bool           `gorm:"default:false" json:"verified"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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

// UserRegistrationInput represents input data for user registration
type UserRegistrationInput struct {
	NimNip   string   `json:"nim_nip" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	UserType UserType `json:"user_type" binding:"required,oneof=student staff"`
	Major    string   `json:"major,omitempty"`
	Faculty  string   `json:"faculty,omitempty"`
	Position string   `json:"position,omitempty"`
}

// UserLoginInput represents input data for user login
type UserLoginInput struct {
	NimNip   string `json:"nim_nip" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID       uint     `json:"id"`
	NimNip   string   `json:"nim_nip"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	UserType UserType `json:"user_type"`
	Major    string   `json:"major,omitempty"`
	Faculty  string   `json:"faculty,omitempty"`
	Position string   `json:"position,omitempty"`
	Verified bool     `json:"verified"`
}

// ToUserResponse converts a User to UserResponse
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		NimNip:   u.NimNip,
		Name:     u.Name,
		Email:    u.Email,
		UserType: u.UserType,
		Major:    u.Major,
		Faculty:  u.Faculty,
		Position: u.Position,
		Verified: u.Verified,
	}
}
