package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"minio-admin-panel/internal/config"
	"minio-admin-panel/internal/handlers"
	"minio-admin-panel/internal/middleware"
	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
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
	apiHandler := handlers.NewAPIHandler(minioService)

	// Setup Gin router
	r := gin.Default()

	// Load HTML templates with custom functions
	r.SetFuncMap(template.FuncMap{
		"formatBytes": formatBytes,
	})
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Routes
	setupRoutes(r, authHandler, bucketHandler, userHandler, policyHandler, groupHandler, apiHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting MinIO Admin Panel on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, bucketHandler *handlers.BucketHandler, userHandler *handlers.UserHandler, policyHandler *handlers.PolicyHandler, groupHandler *handlers.GroupHandler, apiHandler *handlers.APIHandler) {
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

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
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
