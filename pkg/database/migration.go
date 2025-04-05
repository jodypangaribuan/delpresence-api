package database

import (
	"log"
	"time"

	"delpresence-api/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// RunMigrations runs all required database migrations
func RunMigrations() error {
	log.Println("Running database migrations...")

	// Auto migrate creates/updates tables based on models
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Lecturer{},
	); err != nil {
		return err
	}

	// Create default admin account if it doesn't exist
	if err := createDefaultAdmin(); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createDefaultAdmin creates a default admin account if it doesn't exist
func createDefaultAdmin() error {
	// Check if any admin user already exists
	var count int64
	if err := DB.Model(&models.User{}).Where("user_type = ?", models.AdminType).Count(&count).Error; err != nil {
		return err
	}

	// If no admin exists, create one
	if count == 0 {
		log.Println("Creating default admin account...")

		// Begin transaction
		tx := DB.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		// Defer transaction rollback (won't do anything if tx.Commit() is called)
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Hash password
		password := "delpresence"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Current time
		now := time.Now()

		// Create admin user
		adminUser := models.User{
			Username:   "admin",
			FirstName:  "System",
			MiddleName: "",
			LastName:   "Administrator",
			Email:      "admin@delpresence.ac.id",
			Password:   string(hashedPassword),
			UserType:   models.AdminType,
			Verified:   true,
			Active:     true,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		// Save user to database
		if err := tx.Create(&adminUser).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Create admin profile
		adminProfile := models.Admin{
			UserID:      adminUser.ID,
			Position:    "System Administrator",
			Department:  "IT Department",
			AccessLevel: models.SuperAdminAccess,
			IsActive:    true,
			LoginCount:  0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Save admin profile to database
		if err := tx.Create(&adminProfile).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return err
		}

		log.Println("Default admin account created successfully")
	}

	return nil
}
