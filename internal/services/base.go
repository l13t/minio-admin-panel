package services

import (
	"context"
	"fmt"
	"log"

	"minio-admin-panel/internal/config"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOService provides MinIO administration functionality
type MinIOService struct {
	config *config.Config
}

// BucketInfo represents bucket information
type BucketInfo struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	Size         int64  `json:"size"`
	ObjectCount  int64  `json:"object_count"`
}

// UserInfo represents user information
type UserInfo struct {
	AccessKey  string   `json:"access_key"`
	Status     string   `json:"status"`
	PolicyName string   `json:"policy_name,omitempty"`
	MemberOf   []string `json:"member_of,omitempty"`
	UpdatedAt  string   `json:"updated_at,omitempty"`
}

// NewMinIOService creates a new MinIO service instance
func NewMinIOService(cfg *config.Config) *MinIOService {
	return &MinIOService{
		config: cfg,
	}
}

// CreateClients creates MinIO client and admin client with provided credentials
func (s *MinIOService) CreateClients(username, password string) (*minio.Client, *madmin.AdminClient, error) {
	endpoint := s.config.GetMinIOEndpoint()
	log.Printf("[DEBUG] Creating MinIO clients for user '%s' to endpoint '%s' (SSL: %t)",
		username, endpoint, s.config.MinIOUseSSL)

	// Initialize MinIO client with provided credentials
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(username, password, ""),
		Secure: s.config.MinIOUseSSL,
	})
	if err != nil {
		log.Printf("[DEBUG] Failed to create MinIO client: %v", err)
		return nil, nil, fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	// Initialize MinIO admin client with provided credentials
	adminClient, err := madmin.New(endpoint, username, password, s.config.MinIOUseSSL)
	if err != nil {
		log.Printf("[DEBUG] Failed to create MinIO admin client: %v", err)
		return nil, nil, fmt.Errorf("failed to initialize MinIO admin client: %v", err)
	}

	log.Printf("[DEBUG] Successfully created MinIO clients for user '%s'", username)
	return minioClient, adminClient, nil
}

// ValidateCredentials validates MinIO admin credentials by testing connection
// This validates the provided username/password against MinIO directly
func (s *MinIOService) ValidateCredentials(username, password string) (*UserInfo, error) {
	log.Printf("[DEBUG] Validating credentials for user '%s'", username)

	// Create clients with provided credentials to test them
	minioClient, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients during validation: %v", err)
		return nil, fmt.Errorf("failed to create MinIO clients: %v", err)
	}

	// Test basic MinIO connection
	log.Printf("[DEBUG] Testing basic MinIO connection for user '%s'", username)
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		log.Printf("[DEBUG] ListBuckets failed for user '%s': %v", username, err)
		return nil, fmt.Errorf("invalid MinIO credentials or connection failed: %v", err)
	}
	log.Printf("[DEBUG] Basic connection successful for user '%s', found %d buckets", username, len(buckets))

	// Test admin capabilities
	log.Printf("[DEBUG] Testing admin privileges for user '%s'", username)
	users, err := adminClient.ListUsers(context.Background())
	if err != nil {
		log.Printf("[DEBUG] Admin privilege test failed for user '%s': %v", username, err)
		return nil, fmt.Errorf("user does not have admin privileges: %v", err)
	}
	log.Printf("[DEBUG] Admin privileges confirmed for user '%s', found %d users in system", username, len(users))

	// Create user info for the authenticated admin
	userInfo := &UserInfo{
		AccessKey:  username, // Using username as identifier
		Status:     "enabled",
		PolicyName: "admin",
	}

	log.Printf("[DEBUG] Credential validation successful for user '%s'", username)
	return userInfo, nil
}

// GetUserPermissions returns admin permissions for authenticated MinIO user
// This validates the credentials against MinIO and grants permissions based on admin capabilities
func (s *MinIOService) GetUserPermissions(username, password string) map[string]bool {
	log.Printf("[DEBUG] Getting user permissions for user '%s'", username)

	// Test the credentials by creating clients
	minioClient, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients for permission check: %v", err)
		return map[string]bool{
			"canListBuckets":    false,
			"canCreateBuckets":  false,
			"canDeleteBuckets":  false,
			"canManageUsers":    false,
			"canManagePolicies": false,
			"isAdmin":           false,
		}
	}

	permissions := map[string]bool{
		"canListBuckets":    false,
		"canCreateBuckets":  false,
		"canDeleteBuckets":  false,
		"canManageUsers":    false,
		"canManagePolicies": false,
		"isAdmin":           false,
	}

	// Test bucket operations
	log.Printf("[DEBUG] Testing bucket permissions for user '%s'", username)
	if _, err := minioClient.ListBuckets(context.Background()); err == nil {
		log.Printf("[DEBUG] User '%s' has bucket permissions", username)
		permissions["canListBuckets"] = true
		permissions["canCreateBuckets"] = true
		permissions["canDeleteBuckets"] = true
	} else {
		log.Printf("[DEBUG] User '%s' failed bucket permissions test: %v", username, err)
	}

	// Test admin operations
	log.Printf("[DEBUG] Testing admin permissions for user '%s'", username)
	if _, err := adminClient.ListUsers(context.Background()); err == nil {
		log.Printf("[DEBUG] User '%s' has admin permissions", username)
		permissions["canManageUsers"] = true
		permissions["canManagePolicies"] = true
		permissions["isAdmin"] = true
	} else {
		log.Printf("[DEBUG] User '%s' failed admin permissions test: %v", username, err)
	}

	log.Printf("[DEBUG] Final permissions for user '%s': %+v", username, permissions)
	return permissions
}
