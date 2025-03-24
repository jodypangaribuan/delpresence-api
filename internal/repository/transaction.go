package repository

import (
	"delpresence-api/pkg/database"

	"gorm.io/gorm"
)

// BeginTransaction starts a new database transaction
func BeginTransaction() *gorm.DB {
	return database.DB.Begin()
}
