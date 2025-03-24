package models

import (
	"time"

	"gorm.io/gorm"
)

// Student represents a student user in the database
type Student struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	NIM       string         `gorm:"uniqueIndex;not null" json:"nim"`
	Major     string         `gorm:"not null" json:"major"`
	Faculty   string         `gorm:"not null" json:"faculty"`
	Batch     string         `gorm:"not null" json:"batch"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// StudentResponse represents the student data returned in API responses
type StudentResponse struct {
	ID      uint         `json:"id"`
	UserID  uint         `json:"user_id"`
	User    UserResponse `json:"user"`
	NIM     string       `json:"nim"`
	Major   string       `json:"major"`
	Faculty string       `json:"faculty"`
	Batch   string       `json:"batch"`
}

// ToStudentResponse converts a Student to StudentResponse
func (s *Student) ToStudentResponse() StudentResponse {
	return StudentResponse{
		ID:      s.ID,
		UserID:  s.UserID,
		User:    s.User.ToUserResponse(),
		NIM:     s.NIM,
		Major:   s.Major,
		Faculty: s.Faculty,
		Batch:   s.Batch,
	}
}

// StudentRegistrationInput represents input data for student registration
type StudentRegistrationInput struct {
	NIM        string `json:"nim" binding:"required"`
	FirstName  string `json:"first_name" binding:"required"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Major      string `json:"major" binding:"required"`
	Faculty    string `json:"faculty" binding:"required"`
	Batch      string `json:"batch" binding:"required"`
}
