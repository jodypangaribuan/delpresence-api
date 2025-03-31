package repository

import (
	"errors"

	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository handles database operations for users
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		DB: database.DB,
	}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(user *models.User) error {
	// Check if user with email already exists
	var count int64
	if err := r.DB.Model(&models.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrUserAlreadyExists
	}

	return r.DB.Create(user).Error
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's information
func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.DB.Save(user).Error
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(userID uint) error {
	return r.DB.Delete(&models.User{}, userID).Error
}
