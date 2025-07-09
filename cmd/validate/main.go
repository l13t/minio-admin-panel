package main

import (
	"context"
	"fmt"
	"minio-admin-panel/internal/config"
	"minio-admin-panel/internal/services"
	"os"
)

func main() {
	fmt.Println("🔧 MinIO Admin Panel Configuration Validator")
	fmt.Println("====================================================")

	// Load configuration
	cfg := config.Load()

	fmt.Printf("MinIO Host: %s\n", cfg.MinIOHost)
	fmt.Printf("MinIO Port: %d\n", cfg.MinIOPort)
	fmt.Printf("MinIO Endpoint: %s\n", cfg.GetMinIOEndpoint())
	fmt.Printf("MinIO Use SSL: %t\n", cfg.MinIOUseSSL)
	fmt.Printf("JWT Secret: %s\n", maskString(cfg.JWTSecret))

	fmt.Println("\n🧪 Testing MinIO Connection...")
	fmt.Println("Note: To test the connection, please provide MinIO admin credentials:")

	// Create MinIO service
	minioService := services.NewMinIOService(cfg)

	// Get test credentials from environment or use defaults
	testUsername := getEnvOrDefault("TEST_MINIO_USER", "minioadmin")
	testPassword := getEnvOrDefault("TEST_MINIO_PASSWORD", "minioadmin")

	fmt.Printf("Testing with username: %s\n", testUsername)

	// Try to validate credentials
	userInfo, err := minioService.ValidateCredentials(testUsername, testPassword)
	if err != nil {
		fmt.Printf("❌ Failed to validate credentials: %v\n", err)
		fmt.Println("\nPossible issues:")
		fmt.Println("1. MinIO server is not running or not accessible")
		fmt.Println("2. Invalid credentials")
		fmt.Println("3. User does not have admin privileges")
		fmt.Println("4. Network connectivity issues")
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully authenticated as: %s (Policy: %s)\n", userInfo.AccessKey, userInfo.PolicyName)

	// Test permissions
	permissions := minioService.GetUserPermissions(testUsername, testPassword)
	fmt.Println("\n🔐 User Permissions:")
	for permission, enabled := range permissions {
		status := "❌"
		if enabled {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, permission)
	}

	// Test basic operations
	fmt.Println("\n🪣 Testing Basic Operations...")

	// List buckets
	buckets, err := minioService.ListBuckets(context.Background(), testUsername, testPassword)
	if err != nil {
		fmt.Printf("❌ Failed to list buckets: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully listed %d buckets\n", len(buckets))
		for _, bucket := range buckets {
			fmt.Printf("  - %s (created: %s)\n", bucket.Name, bucket.CreationDate)
		}
	}

	fmt.Println("\n🎉 Configuration validation completed successfully!")
	fmt.Println("Your MinIO Admin Panel is ready to use.")
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
