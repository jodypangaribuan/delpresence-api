package repository

import (
	"errors"

	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDatabaseError      = errors.New("database error")
)

// UserRepository handles database operations for User
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		DB: database.DB,
	}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *models.User) error {
	// Check if user with same NIM/NIP or email already exists
	var existingUser models.User
	result := r.DB.Where("nim_nip = ? OR email = ?", user.NimNip, user.Email).First(&existingUser)
	if result.Error == nil {
		return ErrUserAlreadyExists
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return ErrDatabaseError
	}

	// Create the user
	result = r.DB.Create(user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := r.DB.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByNimNip retrieves a user by NIM/NIP
func (r *UserRepository) GetUserByNimNip(nimNip string) (*models.User, error) {
	var user models.User
	result := r.DB.Where("nim_nip = ?", nimNip).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser updates user information
func (r *UserRepository) UpdateUser(user *models.User) error {
	result := r.DB.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// VerifyUser updates user verification status
func (r *UserRepository) VerifyUser(userID uint) error {
	result := r.DB.Model(&models.User{}).Where("id = ?", userID).Update("verified", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser marks a user as deleted
func (r *UserRepository) DeleteUser(userID uint) error {
	result := r.DB.Delete(&models.User{}, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
