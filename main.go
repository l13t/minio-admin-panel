package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"minio-admin-panel/internal/config"
	"minio-admin-panel/internal/handlers"
	"minio-admin-panel/internal/i18n"
	"minio-admin-panel/internal/middleware"
	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

// Build-time variables set by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Print version information
	log.Printf("MinIO Admin Panel %s (commit: %s, built: %s by %s)", version, commit, date, builtBy)

	// Initialize internationalization
	i18n.Init("en") // Default language: English
	if err := i18n.LoadDir("translations/i18n"); err != nil {
		log.Printf("Warning: Failed to load translations: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize MinIO service
	minioService := services.NewMinIOService(cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(minioService)
	bucketHandler := handlers.NewBucketHandler(minioService)
	userHandler := handlers.NewUserHandler(minioService)
	policyHandler := handlers.NewPolicyHandler(minioService)
	groupHandler := handlers.NewGroupHandler(minioService)
	serviceAccountHandler := handlers.NewServiceAccountHandler(minioService)
	apiHandler := handlers.NewAPIHandler(minioService)
	settingsHandler := handlers.NewSettingsHandler(minioService, version, commit, date, builtBy)

	// Setup Gin router
	r := gin.Default()

	// Add language middleware
	r.Use(middleware.LanguageMiddleware())

	// Load HTML templates with custom functions
	r.SetFuncMap(template.FuncMap{
		"formatBytes": formatBytes,
		"t": func(key string) string {
			// This will be overridden in template execution context
			return key
		},
		"tWithParams": func(key string, params ...interface{}) string {
			// This will be overridden in template execution context
			return key
		},
	})
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Routes
	setupRoutes(r, authHandler, bucketHandler, userHandler, policyHandler, groupHandler, serviceAccountHandler, apiHandler, settingsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting MinIO Admin Panel on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, bucketHandler *handlers.BucketHandler, userHandler *handlers.UserHandler, policyHandler *handlers.PolicyHandler, groupHandler *handlers.GroupHandler, serviceAccountHandler *handlers.ServiceAccountHandler, apiHandler *handlers.APIHandler, settingsHandler *handlers.SettingsHandler) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": version})
	})

	// Auth routes
	r.GET("/", authHandler.LoginPage)
	r.POST("/login", authHandler.Login)
	r.POST("/logout", authHandler.Logout)

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		// Dashboard - accessible to all authenticated users
		protected.GET("/dashboard", func(c *gin.Context) {
			permissions := middleware.GetUserPermissions(c)
			username, _ := c.Get("username")
			policyName, _ := c.Get("policy_name")

			// Use the helper for consistent translation handling
			handlers.RenderWithTranslations(c, "dashboard.html", gin.H{
				"title":       "MinIO Admin Panel",
				"username":    username,
				"policy_name": policyName,
				"permissions": permissions,
			})
		})

		// Bucket management - require bucket list permission
		bucketRoutes := protected.Group("/buckets")
		bucketRoutes.Use(middleware.RequirePermission("canListBuckets"))
		{
			bucketRoutes.GET("", bucketHandler.ListBuckets)
			bucketRoutes.POST("", middleware.RequirePermission("canCreateBuckets"), bucketHandler.CreateBucket)
			bucketRoutes.DELETE("/:name", middleware.RequirePermission("canDeleteBuckets"), bucketHandler.DeleteBucket)
			bucketRoutes.GET("/:name/policy", bucketHandler.GetBucketPolicy)
			bucketRoutes.PUT("/:name/policy", middleware.RequirePermission("canManagePolicies"), bucketHandler.SetBucketPolicy)
		}

		// User management - require admin permissions
		userRoutes := protected.Group("/users")
		userRoutes.Use(middleware.RequirePermission("canManageUsers"))
		{
			userRoutes.GET("", userHandler.ListUsers)
			userRoutes.POST("", userHandler.CreateUser)
			userRoutes.GET("/:name", userHandler.GetUser)
			userRoutes.GET("/:name/details", userHandler.GetUserDetails)
			userRoutes.GET("/:name/credentials", userHandler.GetUserCredentials)
			userRoutes.DELETE("/:name", userHandler.DeleteUser)
			userRoutes.PUT("/:name/credentials", userHandler.UpdateUserCredentials)
			userRoutes.PUT("/:name/status", userHandler.SetUserStatus)
			userRoutes.GET("/:name/policy", userHandler.GetUserPolicy)
			userRoutes.PUT("/:name/policy", userHandler.SetUserPolicy)
			userRoutes.PUT("/:name/groups", groupHandler.SetUserGroups)
		}

		// Group management - require admin permissions
		groupRoutes := protected.Group("/groups")
		groupRoutes.Use(middleware.RequirePermission("canManageUsers"))
		{
			groupRoutes.GET("", groupHandler.ListGroups)
			groupRoutes.POST("", groupHandler.CreateGroup)
			groupRoutes.GET("/:name", groupHandler.GetGroupInfo)
			groupRoutes.DELETE("/:name", groupHandler.DeleteGroup)
			groupRoutes.PUT("/:name/members", groupHandler.UpdateGroupMembers)
			groupRoutes.PUT("/:name/policy", groupHandler.SetGroupPolicy)
		}

		// Service Account management - require admin permissions
		serviceAccountRoutes := protected.Group("/service-accounts")
		serviceAccountRoutes.Use(middleware.RequirePermission("canManageUsers"))
		{
			serviceAccountRoutes.GET("", serviceAccountHandler.ListServiceAccounts)
			serviceAccountRoutes.POST("", serviceAccountHandler.CreateServiceAccount)
			serviceAccountRoutes.GET("/:accessKey", serviceAccountHandler.GetServiceAccountInfo)
			serviceAccountRoutes.DELETE("/:accessKey", serviceAccountHandler.DeleteServiceAccount)
		}

		// Policy management - require admin permissions
		policyRoutes := protected.Group("/policies")
		policyRoutes.Use(middleware.RequirePermission("canManageUsers"))
		{
			policyRoutes.GET("", policyHandler.ListPolicies)
			policyRoutes.GET("/:name", policyHandler.GetPolicyDocument)
			policyRoutes.POST("/:name", policyHandler.CreateOrUpdatePolicy)
			policyRoutes.PUT("/:name", policyHandler.CreateOrUpdatePolicy)
			policyRoutes.DELETE("/:name", policyHandler.DeletePolicy)
		}

		// Settings - require admin permissions
		protected.GET("/settings", middleware.RequirePermission("isAdmin"), settingsHandler.ShowSettings)

		// API routes for AJAX
		api := protected.Group("/api")
		{
			api.GET("/server-info", apiHandler.GetServerInfo)
			api.GET("/metrics", apiHandler.GetMetrics)
			api.GET("/storage-usage", apiHandler.GetStorageUsage)
			api.GET("/policies", userHandler.ListPolicies)
			api.GET("/groups", func(c *gin.Context) {
				// Forward to group handler with JSON accept header
				c.Request.Header.Set("Accept", "application/json")
				groupHandler.ListGroups(c)
			})
			api.GET("/service-accounts", serviceAccountHandler.ListServiceAccounts)
			api.POST("/service-accounts", serviceAccountHandler.CreateServiceAccount)
			api.GET("/service-accounts/:accessKey", serviceAccountHandler.GetServiceAccountInfo)
			api.DELETE("/service-accounts/:accessKey", serviceAccountHandler.DeleteServiceAccount)
		}
	}
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	if bytes < 0 {
		return "Calculating..."
	}
	if bytes == 0 {
		return "Empty"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", size, units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}
