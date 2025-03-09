package repository

import (
	"errors"
	"time"

	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"gorm.io/gorm"
)

var (
	ErrTokenNotFound   = errors.New("token not found")
	ErrTokenExpired    = errors.New("token has expired")
	ErrTokenCreateFail = errors.New("failed to create token")
)

// TokenRepository handles database operations for Token
type TokenRepository struct {
	DB *gorm.DB
}

// NewTokenRepository creates a new instance of TokenRepository
func NewTokenRepository() *TokenRepository {
	return &TokenRepository{
		DB: database.DB,
	}
}

// CreateToken creates a new refresh token in the database
func (r *TokenRepository) CreateToken(userID uint, token string, expiry time.Time) error {
	refreshToken := &models.Token{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiry,
	}

	result := r.DB.Create(refreshToken)
	if result.Error != nil {
		return ErrTokenCreateFail
	}

	return nil
}

// GetTokenByValue retrieves a token by its value
func (r *TokenRepository) GetTokenByValue(tokenStr string) (*models.Token, error) {
	var token models.Token
	result := r.DB.Where("token = ?", tokenStr).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, result.Error
	}

	// Check if token has expired
	if token.ExpiresAt.Before(time.Now()) {
		// Delete expired token
		r.DB.Delete(&token)
		return nil, ErrTokenExpired
	}

	return &token, nil
}

// DeleteToken deletes a token from the database
func (r *TokenRepository) DeleteToken(tokenStr string) error {
	result := r.DB.Where("token = ?", tokenStr).Delete(&models.Token{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// DeleteAllUserTokens deletes all tokens for a specific user
func (r *TokenRepository) DeleteAllUserTokens(userID uint) error {
	result := r.DB.Where("user_id = ?", userID).Delete(&models.Token{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (r *TokenRepository) DeleteExpiredTokens() error {
	result := r.DB.Where("expires_at < ?", time.Now()).Delete(&models.Token{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
