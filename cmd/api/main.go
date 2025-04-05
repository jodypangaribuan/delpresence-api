package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"delpresence-api/internal/handlers"
	"delpresence-api/internal/middleware"
	"delpresence-api/internal/repository"
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

	// Run database migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
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
	adminHandler := handlers.NewAdminHandler()

	// Get database connection
	db := database.GetDB()

	// Setup lecturer repository and handler
	lecturerRepo := repository.NewLecturerRepository(db)
	lecturerHandler := handlers.NewLecturerHandler(lecturerRepo)

	// Setup assistant repository and handler
	assistantRepo := repository.NewAssistantRepository(db)
	assistantHandler := handlers.NewAssistantHandler(assistantRepo)

	// Auth routes
	auth := api.Group("/auth")
	{
		// Campus login endpoint (not protected)
		auth.POST("/campus/login", authHandler.CampusLogin)

		// Admin login endpoint (not protected)
		auth.POST("/admin/login", adminHandler.Login)

		// Auth required endpoints
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

	// Admin routes
	admin := api.Group("/admin")
	{
		admin.POST("/login", adminHandler.Login)

		// Admin endpoints that require auth
		adminAuth := admin.Group("")
		adminAuth.Use(middleware.AdminAuth())
		{
			adminAuth.GET("/profile", adminHandler.GetAdminProfile)
		}
	}

	// Lecturer routes
	lecturer := api.Group("/lecturer")
	lecturer.Use(middleware.AuthMiddleware()) // Protect all lecturer routes
	{
		lecturer.GET("/profile", lecturerHandler.GetLecturerProfile)
		lecturer.POST("/sync", lecturerHandler.SyncLecturerProfile)
		lecturer.PATCH("/profile", lecturerHandler.UpdateLecturerProfile)
	}

	// Assistant routes
	assistant := api.Group("/assistant")
	assistant.Use(middleware.AuthMiddleware()) // Protect all assistant routes
	{
		assistant.GET("/profile", assistantHandler.GetAssistantProfile)
		assistant.POST("/sync", assistantHandler.SyncAssistantProfile)
		assistant.PATCH("/profile", assistantHandler.UpdateAssistantProfile)
	}

	// Add more API routes here
}
