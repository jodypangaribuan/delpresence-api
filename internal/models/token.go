package models

import (
	"time"

	"gorm.io/gorm"
)

// TokenType represents the type of token
type TokenType string

const (
	// RefreshToken represents a refresh token for JWT authentication
	RefreshToken TokenType = "refresh"
)

// Token represents a stored token in the database
type Token struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Token     string         `gorm:"not null;unique" json:"token"`
	Type      TokenType      `gorm:"not null;type:VARCHAR(20)" json:"type"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// TokenResponse represents the token data returned in API responses
type TokenResponse struct {
	Token     string    `json:"token"`
	Type      TokenType `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ToTokenResponse converts a Token to TokenResponse
func (t *Token) ToTokenResponse() TokenResponse {
	return TokenResponse{
		Token:     t.Token,
		Type:      t.Type,
		ExpiresAt: t.ExpiresAt,
	}
}

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Expiry in seconds
}
