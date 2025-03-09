package main

import (
	"log"
	"os"
	"strings"

	"delpresence-api/internal/handlers"
	"delpresence-api/internal/middleware"
	"delpresence-api/pkg/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Set Gin mode
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Connect to database
	dbConnected := true
	if err := database.ConnectDB(); err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("Running in demo mode without database connection")
		dbConnected = false
	}

	// Create router
	router := gin.Default()

	// Configure CORS
	configCors(router)

	// Create API routes
	if dbConnected {
		setupRoutes(router)
	} else {
		// Setup demo routes
		setupDemoRoutes(router)
	}

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

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)

		// Protected routes
		authRequired := auth.Group("/")
		authRequired.Use(middleware.AuthMiddleware())
		{
			authRequired.GET("/me", authHandler.GetCurrentUser)
		}
	}

	// Add more API routes here
}

// setupDemoRoutes sets up demo routes for when database is not available
func setupDemoRoutes(router *gin.Engine) {
	// API version prefix
	api := router.Group("/api/v1")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "DelPresence API is running in demo mode (no database connection)",
		})
	})

	// Demo auth routes
	auth := api.Group("/auth")
	{
		// Demo register
		auth.POST("/register", func(c *gin.Context) {
			var input map[string]interface{}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(400, gin.H{
					"success": false,
					"message": "Validation error",
					"error":   err.Error(),
				})
				return
			}

			// Return demo response
			c.JSON(201, gin.H{
				"success": true,
				"message": "User registered successfully (DEMO MODE)",
				"data": gin.H{
					"id":        1,
					"nim_nip":   input["nim_nip"],
					"name":      input["name"],
					"email":     input["email"],
					"user_type": input["user_type"],
					"verified":  false,
				},
			})
		})

		// Demo login
		auth.POST("/login", func(c *gin.Context) {
			var input map[string]interface{}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(400, gin.H{
					"success": false,
					"message": "Validation error",
					"error":   err.Error(),
				})
				return
			}

			// Return demo response with fake tokens
			c.JSON(200, gin.H{
				"success": true,
				"message": "Login successful (DEMO MODE)",
				"data": gin.H{
					"user": gin.H{
						"id":        1,
						"nim_nip":   input["nim_nip"],
						"name":      "Demo User",
						"email":     "demo@example.com",
						"user_type": "student",
						"verified":  false,
					},
					"tokens": gin.H{
						"access_token":  "demo.access.token",
						"refresh_token": "demo.refresh.token",
						"expires_in":    86400,
					},
				},
			})
		})

		// Demo refresh token
		auth.POST("/refresh", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Token refreshed successfully (DEMO MODE)",
				"data": gin.H{
					"access_token": "demo.new.access.token",
					"expires_in":   86400,
				},
			})
		})

		// Demo logout
		auth.POST("/logout", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Logged out successfully (DEMO MODE)",
			})
		})

		// Demo me endpoint
		auth.GET("/me", func(c *gin.Context) {
			// Check for Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(401, gin.H{
					"success": false,
					"message": "Unauthorized",
				})
				return
			}

			c.JSON(200, gin.H{
				"success": true,
				"message": "User details retrieved successfully (DEMO MODE)",
				"data": gin.H{
					"id":        1,
					"nim_nip":   "12345",
					"name":      "Demo User",
					"email":     "demo@example.com",
					"user_type": "student",
					"verified":  false,
				},
			})
		})
	}
}
