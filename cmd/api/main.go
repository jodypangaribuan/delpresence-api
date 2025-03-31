package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"delpresence-api/internal/handlers"
	"delpresence-api/internal/middleware"
	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Get the executable path
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)

	// Try to load .env from multiple locations
	envPaths := []string{
		".env",                        // Current directory
		"../../.env",                  // Project root when running from cmd/api
		filepath.Join(exPath, ".env"), // Binary location
	}

	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			log.Printf("Loaded .env from: %s", path)
			break
		}
	}

	if !envLoaded {
		log.Println("Warning: .env file not found, using default values")
	}

	// Set Gin mode
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Connect to database
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		log.Printf("Warning: Migration error: %v", err)
	}

	// Create router
	router := gin.Default()

	// Configure CORS
	configCors(router)

	// Create API routes
	setupRoutes(router)

	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server running at http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func configCors(router *gin.Engine) {
	// Get allowed origins from environment
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"http://localhost:3000"}

	if allowedOriginsStr != "" {
		allowedOrigins = strings.Split(allowedOriginsStr, ",")
	}

	// Configure CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = allowedOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true

	router.Use(cors.New(config))
}

func setupRoutes(router *gin.Engine) {
	// API version prefix
	api := router.Group("/api/v1")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "DelPresence API is running",
		})
	})

	// Create handlers
	authHandler := handlers.NewAuthHandler()
	mahasiswaHandler := handlers.NewMahasiswaHandler()

	// Auth routes
	auth := api.Group("/auth")
	{
		// Keep only admin login
		auth.POST("/admin/login", authHandler.AdminLogin)

		// Keep the me endpoint
		authRequired := auth.Group("/")
		authRequired.Use(middleware.AuthMiddleware())
		{
			authRequired.GET("/me", authHandler.GetCurrentUser)
		}
	}

	// Mahasiswa routes
	mahasiswa := api.Group("/mahasiswa")
	mahasiswa.Use(middleware.AuthMiddleware()) // Protect all mahasiswa routes
	{
		mahasiswa.GET("", mahasiswaHandler.GetMahasiswaByUserID)
		mahasiswa.GET("/", mahasiswaHandler.GetMahasiswaByUserID)
		mahasiswa.GET("/by-user-id", mahasiswaHandler.GetMahasiswaByUserID)
		mahasiswa.GET("/by-nim", mahasiswaHandler.GetMahasiswaDetailByNIM)
		mahasiswa.GET("/complete", mahasiswaHandler.GetMahasiswaComplete)
	}

	// Add more API routes here
}

// runMigrations runs any necessary database migrations
func runMigrations() error {
	// Migrate name field to first_name, middle_name, last_name
	if err := migrateNameFields(); err != nil {
		return err
	}

	// Create admin user if it doesn't exist
	if err := createAdminUser(); err != nil {
		return err
	}

	return nil
}

// migrateNameFields splits the name field into first_name, middle_name, and last_name
func migrateNameFields() error {
	// Check if the migration has already been run
	var count int64
	if err := database.DB.Model(&models.User{}).Where("first_name != ''").Count(&count).Error; err != nil {
		return err
	}

	// If there are already users with first_name set, assume migration has been run
	if count > 0 {
		log.Println("Name field migration already completed")
		return nil
	}

	// This migration was intended to run once to convert old 'name' field data
	// Since the 'name' field has been removed from the User struct, we can't directly access it
	// We'll use a raw SQL query to check if the name column exists

	var nameColumnExists bool
	result := database.DB.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'name')").Scan(&nameColumnExists)
	if result.Error != nil {
		return result.Error
	}

	if !nameColumnExists {
		log.Println("Name column no longer exists, skipping migration")
		return nil
	}

	// Use raw SQL to get user IDs and names
	type OldUser struct {
		ID   uint
		Name string
	}
	var oldUsers []OldUser
	if err := database.DB.Raw("SELECT id, name FROM users").Scan(&oldUsers).Error; err != nil {
		return err
	}

	// Process each user
	for _, user := range oldUsers {
		// Skip if empty name
		if user.Name == "" {
			continue
		}

		// Split name into parts
		nameParts := strings.Fields(user.Name)

		// Update user with split name
		updates := map[string]interface{}{
			"first_name":  "",
			"middle_name": "",
			"last_name":   "",
		}

		if len(nameParts) > 0 {
			updates["first_name"] = nameParts[0]
		}

		if len(nameParts) > 2 {
			// Middle parts become middle name
			updates["middle_name"] = strings.Join(nameParts[1:len(nameParts)-1], " ")
			updates["last_name"] = nameParts[len(nameParts)-1]
		} else if len(nameParts) == 2 {
			// If only two parts, assume first and last name
			updates["last_name"] = nameParts[1]
		}

		// Update the user
		if err := database.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	log.Println("Successfully migrated name fields for all users")
	return nil
}

// createAdminUser creates an admin user if it doesn't exist
func createAdminUser() error {
	// Import necessary repositories
	db := database.GetDB()
	var count int64

	// Check if admin user exists
	db.Model(&models.User{}).
		Where("email = ? AND user_type = ?", "admin@del.ac.id", models.AdminType).
		Count(&count)

	// If admin doesn't exist, create one
	if count == 0 {
		log.Println("Creating admin user...")
		// Start a transaction
		tx := db.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		defer tx.Rollback()

		// Create admin user
		user := &models.User{
			FirstName: "Admin",
			LastName:  "Del",
			Email:     "admin@del.ac.id",
			Password:  "delpresence",
			UserType:  models.AdminType,
			Verified:  true,
		}

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Create admin profile
		admin := &models.Admin{
			UserID: user.ID,
		}

		if err := tx.Create(admin).Error; err != nil {
			return err
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return err
		}

		log.Println("Admin user created successfully")
	} else {
		log.Println("Admin user already exists")
	}

	return nil
}
