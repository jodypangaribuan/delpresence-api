package models

import (
	"time"

	"gorm.io/gorm"
)

// Lecture represents a lecturer/staff user in the database
type Lecture struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	NIP       string         `gorm:"column:nip;uniqueIndex;not null" json:"nip"`
	Position  string         `gorm:"not null" json:"position"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LectureResponse represents the lecturer data returned in API responses
type LectureResponse struct {
	ID       uint         `json:"id"`
	UserID   uint         `json:"user_id"`
	User     UserResponse `json:"user"`
	NIP      string       `json:"nip"`
	Position string       `json:"position"`
}

// ToLectureResponse converts a Lecture to LectureResponse
func (l *Lecture) ToLectureResponse() LectureResponse {
	return LectureResponse{
		ID:       l.ID,
		UserID:   l.UserID,
		User:     l.User.ToUserResponse(),
		NIP:      l.NIP,
		Position: l.Position,
	}
}

// LectureRegistrationInput represents input data for lecturer registration
type LectureRegistrationInput struct {
	NIP        string `json:"nip" binding:"required"`
	FirstName  string `json:"first_name" binding:"required"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Position   string `json:"position" binding:"required"`
}
