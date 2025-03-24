package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"delpresence-api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB establishes a connection to the database
func ConnectDB() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSL_MODE")
	timezone := os.Getenv("DB_TIMEZONE")

	// Set default values if environment variables are not set
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if dbname == "" {
		dbname = "delpresence"
	}
	if sslMode == "" {
		sslMode = "disable"
	}
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}

	// Create DSN string for PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, password, dbname, port, sslMode, timezone)

	// Configure logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Include params in SQL log
			Colorful:                  true,        // Enable color
		},
	)

	// Open connection
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}

	// Migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Student{},
		&models.Lecture{},
		&models.Token{},
		&models.Admin{},
	)
	if err != nil {
		return err
	}

	// Get underlying SQL DB to set connection pool settings
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to PostgreSQL database successfully!")
	return nil
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return DB
}
