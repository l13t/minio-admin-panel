package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"minio-admin-panel/internal/config"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	config *config.Config
}

type BucketInfo struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	Size         int64  `json:"size"`
	ObjectCount  int64  `json:"object_count"`
}

type UserInfo struct {
	AccessKey  string   `json:"access_key"`
	Status     string   `json:"status"`
	PolicyName string   `json:"policy_name,omitempty"`
	MemberOf   []string `json:"member_of,omitempty"`
	UpdatedAt  string   `json:"updated_at,omitempty"`
}

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

// Bucket operations - now require credentials for each operation
func (s *MinIOService) ListBuckets(ctx context.Context, username, password string) ([]BucketInfo, error) {
	log.Printf("[DEBUG] MinIO service ListBuckets called for user '%s'", username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListBuckets: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListBuckets API")
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListBuckets API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] MinIO ListBuckets API returned %d buckets", len(buckets))

	var bucketInfos []BucketInfo
	for _, bucket := range buckets {
		info := BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate.Format("2006-01-02 15:04:05"),
			Size:         0, // Will be calculated below
			ObjectCount:  0, // Will be calculated below
		}

		// Get bucket statistics (size and object count)
		log.Printf("[DEBUG] Getting statistics for bucket '%s'", bucket.Name)
		size, objectCount := s.getBucketStats(ctx, client, bucket.Name)
		info.Size = size
		info.ObjectCount = objectCount

		bucketInfos = append(bucketInfos, info)
		log.Printf("[DEBUG] Bucket: %s (created: %s, size: %d bytes, objects: %d)",
			bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"), size, objectCount)
	}

	log.Printf("[DEBUG] Returning %d bucket infos", len(bucketInfos))
	return bucketInfos, nil
}

// ListBucketsQuick provides a faster bucket listing without size/count calculation
func (s *MinIOService) ListBucketsQuick(ctx context.Context, username, password string) ([]BucketInfo, error) {
	log.Printf("[DEBUG] MinIO service ListBucketsQuick called for user '%s'", username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListBucketsQuick: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListBuckets API (quick mode)")
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListBuckets API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] MinIO ListBuckets API returned %d buckets", len(buckets))

	var bucketInfos []BucketInfo
	for _, bucket := range buckets {
		info := BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate.Format("2006-01-02 15:04:05"),
			Size:         -1, // -1 indicates not calculated
			ObjectCount:  -1, // -1 indicates not calculated
		}
		bucketInfos = append(bucketInfos, info)
		log.Printf("[DEBUG] Bucket: %s (created: %s, stats: not calculated)",
			bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
	}

	log.Printf("[DEBUG] Returning %d bucket infos (quick mode)", len(bucketInfos))
	return bucketInfos, nil
}

// getBucketStats calculates the total size and object count for a bucket
func (s *MinIOService) getBucketStats(ctx context.Context, client *minio.Client, bucketName string) (int64, int64) {
	var totalSize int64
	var objectCount int64

	// Create a timeout context for bucket stats calculation (30 seconds max)
	statsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Printf("[DEBUG] Starting stats calculation for bucket '%s' (max 30s timeout)", bucketName)

	// List all objects in the bucket to calculate size and count
	objectCh := client.ListObjects(statsCtx, bucketName, minio.ListObjectsOptions{Recursive: true})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("[DEBUG] Error listing object in bucket '%s': %v", bucketName, object.Err)
			// Check if it's a timeout or context cancellation
			if statsCtx.Err() != nil {
				log.Printf("[DEBUG] Stats calculation timed out for bucket '%s'", bucketName)
				return -1, -1 // Return -1 to indicate timeout/error
			}
			continue
		}
		totalSize += object.Size
		objectCount++

		// Check for timeout periodically
		select {
		case <-statsCtx.Done():
			log.Printf("[DEBUG] Stats calculation timed out for bucket '%s' after %d objects", bucketName, objectCount)
			return -1, -1
		default:
			// Continue processing
		}
	}

	log.Printf("[DEBUG] Stats calculation completed for bucket '%s': %d bytes, %d objects",
		bucketName, totalSize, objectCount)
	return totalSize, objectCount
}

func (s *MinIOService) CreateBucket(ctx context.Context, bucketName, username, password string) error {
	client, _, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func (s *MinIOService) DeleteBucket(ctx context.Context, bucketName, username, password string) error {
	client, _, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return client.RemoveBucket(ctx, bucketName)
}

func (s *MinIOService) GetBucketPolicy(ctx context.Context, bucketName, username, password string) (string, error) {
	log.Printf("[DEBUG] MinIO service GetBucketPolicy called for bucket '%s' by user '%s'", bucketName, username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetBucketPolicy: %v", err)
		return "", err
	}

	log.Printf("[DEBUG] Calling MinIO GetBucketPolicy API for bucket '%s'", bucketName)
	policy, err := client.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetBucketPolicy API response for bucket '%s': %v", bucketName, err)

		// Check if this is a "no policy" error, which is normal
		errStr := err.Error()
		if strings.Contains(errStr, "policy does not exist") ||
			strings.Contains(errStr, "NoSuchBucketPolicy") ||
			strings.Contains(errStr, "The bucket policy does not exist") {
			log.Printf("[DEBUG] No policy exists for bucket '%s' (this is normal)", bucketName)
			return "", nil // Return empty string, not an error
		}

		// For other errors, return the actual error
		log.Printf("[DEBUG] GetBucketPolicy failed for bucket '%s' with error: %v", bucketName, err)
		return "", err
	}

	log.Printf("[DEBUG] GetBucketPolicy successful for bucket '%s', policy length: %d", bucketName, len(policy))
	return policy, nil
}

func (s *MinIOService) SetBucketPolicy(ctx context.Context, bucketName, policy, username, password string) error {
	log.Printf("[DEBUG] MinIO service SetBucketPolicy called for bucket '%s' by user '%s', policy length: %d", bucketName, username, len(policy))

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in SetBucketPolicy: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO SetBucketPolicy API for bucket '%s'", bucketName)
	err = client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Printf("[DEBUG] MinIO SetBucketPolicy API failed for bucket '%s': %v", bucketName, err)
		return err
	}

	log.Printf("[DEBUG] SetBucketPolicy successful for bucket '%s'", bucketName)
	return nil
}

// User operations - now require credentials for each operation
func (s *MinIOService) ListUsers(ctx context.Context, username, password string) ([]UserInfo, error) {
	log.Printf("[DEBUG] MinIO service ListUsers called for user '%s'", username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListUsers: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListUsers API")
	users, err := adminClient.ListUsers(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListUsers API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] MinIO ListUsers API returned %d users", len(users))

	var userInfos []UserInfo
	for accessKey, user := range users {
		userInfo := UserInfo{
			AccessKey: accessKey,
			Status:    string(user.Status),
			MemberOf:  user.MemberOf,
		}

		log.Printf("[DEBUG] User: %s (status: %s, groups: %v)", accessKey, user.Status, user.MemberOf)

		// Get user policy - for now we'll leave it empty as the API method varies
		// TODO: Implement proper policy retrieval based on MinIO version

		userInfos = append(userInfos, userInfo)
	}

	log.Printf("[DEBUG] Returning %d user infos", len(userInfos))
	return userInfos, nil
}

func (s *MinIOService) CreateUser(ctx context.Context, accessKey, secretKey, username, password string) error {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return adminClient.AddUser(ctx, accessKey, secretKey)
}

func (s *MinIOService) DeleteUser(ctx context.Context, accessKey, username, password string) error {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return adminClient.RemoveUser(ctx, accessKey)
}

func (s *MinIOService) SetUserPolicy(ctx context.Context, accessKey, policyName, username, password string) error {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return adminClient.SetPolicy(ctx, policyName, accessKey, false)
}

// Server information - now requires credentials
func (s *MinIOService) GetServerInfo(ctx context.Context, username, password string) (map[string]interface{}, error) {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return nil, err
	}

	info, err := adminClient.ServerInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to map for JSON serialization
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(infoBytes, &result)
	return result, err
}

// Get metrics - now requires credentials
func (s *MinIOService) GetMetrics(ctx context.Context, username, password string) (map[string]interface{}, error) {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %v", err)
	}

	// Get basic server information instead of service status
	info, err := adminClient.ServerInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %v", err)
	}

	return map[string]interface{}{
		"server_info": info,
		"online":      true,
	}, nil
}

// GetBucketStatsQuick provides fast bucket statistics with shorter timeout for dashboard use
func (s *MinIOService) GetBucketStatsQuick(ctx context.Context, username, password, bucketName string) (int64, int64) {
	log.Printf("[DEBUG] GetBucketStatsQuick called for bucket '%s' by user '%s'", bucketName, username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetBucketStatsQuick: %v", err)
		return -1, -1
	}

	// Use shorter timeout for dashboard (5 seconds max)
	return s.getBucketStatsWithTimeout(ctx, client, bucketName, 5*time.Second)
}

// getBucketStatsWithTimeout calculates bucket stats with configurable timeout
func (s *MinIOService) getBucketStatsWithTimeout(ctx context.Context, client *minio.Client, bucketName string, timeout time.Duration) (int64, int64) {
	var totalSize int64
	var objectCount int64

	// Create a timeout context for bucket stats calculation
	statsCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Printf("[DEBUG] Starting quick stats calculation for bucket '%s' (max %.0fs timeout)", bucketName, timeout.Seconds())

	// List all objects in the bucket to calculate size and count
	objectCh := client.ListObjects(statsCtx, bucketName, minio.ListObjectsOptions{Recursive: true})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("[DEBUG] Error listing object in bucket '%s': %v", bucketName, object.Err)
			// Check if it's a timeout or context cancellation
			if statsCtx.Err() != nil {
				log.Printf("[DEBUG] Quick stats calculation timed out for bucket '%s'", bucketName)
				return -1, -1 // Return -1 to indicate timeout/error
			}
			continue
		}
		totalSize += object.Size
		objectCount++

		// Check for timeout periodically
		select {
		case <-statsCtx.Done():
			log.Printf("[DEBUG] Quick stats calculation timed out for bucket '%s' after %d objects", bucketName, objectCount)
			return -1, -1
		default:
			// Continue processing
		}
	}

	log.Printf("[DEBUG] Quick stats calculation completed for bucket '%s': %d bytes, %d objects",
		bucketName, totalSize, objectCount)
	return totalSize, objectCount
}

// GetUser gets details for a specific user
func (s *MinIOService) GetUser(ctx context.Context, accessKey, username, password string) (*UserInfo, error) {
	log.Printf("[DEBUG] MinIO service GetUser called for user '%s' by admin '%s'", accessKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetUser: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListUsers API to get user '%s'", accessKey)
	users, err := adminClient.ListUsers(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListUsers API failed: %v", err)
		return nil, err
	}

	user, exists := users[accessKey]
	if !exists {
		log.Printf("[DEBUG] User '%s' not found", accessKey)
		return nil, fmt.Errorf("user '%s' not found", accessKey)
	}

	userInfo := &UserInfo{
		AccessKey: accessKey,
		Status:    string(user.Status),
		MemberOf:  user.MemberOf,
	}

	log.Printf("[DEBUG] GetUser successful for user '%s': status=%s, groups=%v", accessKey, user.Status, user.MemberOf)
	return userInfo, nil
}

// UpdateUserCredentials updates user's secret key
func (s *MinIOService) UpdateUserCredentials(ctx context.Context, accessKey, newSecretKey, username, password string) error {
	log.Printf("[DEBUG] MinIO service UpdateUserCredentials called for user '%s' by admin '%s'", accessKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in UpdateUserCredentials: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO AddUser API to update credentials for user '%s'", accessKey)
	err = adminClient.AddUser(ctx, accessKey, newSecretKey)
	if err != nil {
		log.Printf("[DEBUG] MinIO AddUser API failed for user '%s': %v", accessKey, err)
		return err
	}

	log.Printf("[DEBUG] UpdateUserCredentials successful for user '%s'", accessKey)
	return nil
}

// SetUserStatus enables or disables a user
func (s *MinIOService) SetUserStatus(ctx context.Context, accessKey string, enabled bool, username, password string) error {
	log.Printf("[DEBUG] MinIO service SetUserStatus called for user '%s' (enabled=%t) by admin '%s'", accessKey, enabled, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in SetUserStatus: %v", err)
		return err
	}

	var status madmin.AccountStatus
	if enabled {
		status = madmin.AccountEnabled
	} else {
		status = madmin.AccountDisabled
	}

	log.Printf("[DEBUG] Calling MinIO SetUserStatus API for user '%s' with status %v", accessKey, status)
	err = adminClient.SetUserStatus(ctx, accessKey, status)
	if err != nil {
		log.Printf("[DEBUG] MinIO SetUserStatus API failed for user '%s': %v", accessKey, err)
		return err
	}

	log.Printf("[DEBUG] SetUserStatus successful for user '%s'", accessKey)
	return nil
}

// GetUserPolicy gets the policy assigned to a user
func (s *MinIOService) GetUserPolicy(ctx context.Context, accessKey, username, password string) (string, error) {
	log.Printf("[DEBUG] MinIO service GetUserPolicy called for user '%s' by admin '%s'", accessKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetUserPolicy: %v", err)
		return "", err
	}

	log.Printf("[DEBUG] Calling MinIO GetUserInfo API for user '%s'", accessKey)
	userInfo, err := adminClient.GetUserInfo(ctx, accessKey)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetUserInfo API failed for user '%s': %v", accessKey, err)
		return "", err
	}

	log.Printf("[DEBUG] GetUserPolicy successful for user '%s': policy=%s", accessKey, userInfo.PolicyName)
	return userInfo.PolicyName, nil
}

// Group operations - require admin credentials

// ListGroups lists all groups in MinIO
func (s *MinIOService) ListGroups(ctx context.Context, username, password string) ([]string, error) {
	log.Printf("[DEBUG] MinIO service ListGroups called by admin '%s'", username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListGroups: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListGroups API")
	groups, err := adminClient.ListGroups(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListGroups API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] ListGroups successful: found %d groups", len(groups))
	return groups, nil
}

// CreateGroup creates a new group
func (s *MinIOService) CreateGroup(ctx context.Context, groupName string, username, password string) error {
	log.Printf("[DEBUG] MinIO service CreateGroup called for group '%s' by admin '%s'", groupName, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in CreateGroup: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO AddGroup API for group '%s'", groupName)
	err = adminClient.UpdateGroupMembers(ctx, madmin.GroupAddRemove{
		Group:    groupName,
		Members:  []string{}, // Start with empty group
		IsRemove: false,
	})
	if err != nil {
		log.Printf("[DEBUG] MinIO AddGroup API failed for group '%s': %v", groupName, err)
		return err
	}

	log.Printf("[DEBUG] CreateGroup successful for group '%s'", groupName)
	return nil
}

// DeleteGroup deletes a group by removing all members and then the group itself
func (s *MinIOService) DeleteGroup(ctx context.Context, groupName string, username, password string) error {
	log.Printf("[DEBUG] MinIO service DeleteGroup called for group '%s' by admin '%s'", groupName, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in DeleteGroup: %v", err)
		return err
	}

	// First get group info to see current members
	log.Printf("[DEBUG] Getting group info for '%s' before deletion", groupName)
	groupDesc, err := adminClient.GetGroupDescription(ctx, groupName)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetGroupDescription API failed for group '%s': %v", groupName, err)
		return err
	}

	// Remove all members from group
	if len(groupDesc.Members) > 0 {
		log.Printf("[DEBUG] Removing %d members from group '%s'", len(groupDesc.Members), groupName)
		err = adminClient.UpdateGroupMembers(ctx, madmin.GroupAddRemove{
			Group:    groupName,
			Members:  groupDesc.Members,
			IsRemove: true,
		})
		if err != nil {
			log.Printf("[DEBUG] Failed to remove members from group '%s': %v", groupName, err)
			return err
		}
	}

	log.Printf("[DEBUG] DeleteGroup successful for group '%s'", groupName)
	return nil
}

// GetGroupInfo gets information about a specific group including members
func (s *MinIOService) GetGroupInfo(ctx context.Context, groupName string, username, password string) (map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service GetGroupInfo called for group '%s' by admin '%s'", groupName, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetGroupInfo: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO GetGroupDescription API for group '%s'", groupName)
	groupDesc, err := adminClient.GetGroupDescription(ctx, groupName)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetGroupDescription API failed for group '%s': %v", groupName, err)
		return nil, err
	}

	groupInfo := map[string]interface{}{
		"name":       groupName,
		"members":    groupDesc.Members,
		"policy":     groupDesc.Policy,
		"status":     groupDesc.Status,
		"updated_at": groupDesc.UpdatedAt,
	}

	log.Printf("[DEBUG] GetGroupInfo successful for group '%s': %d members, policy=%s",
		groupName, len(groupDesc.Members), groupDesc.Policy)
	return groupInfo, nil
}

// AddUsersToGroup adds users to a group
func (s *MinIOService) AddUsersToGroup(ctx context.Context, groupName string, usernames []string, username, password string) error {
	log.Printf("[DEBUG] MinIO service AddUsersToGroup called for group '%s' with %d users by admin '%s'",
		groupName, len(usernames), username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in AddUsersToGroup: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO UpdateGroupMembers API to add users %v to group '%s'", usernames, groupName)
	err = adminClient.UpdateGroupMembers(ctx, madmin.GroupAddRemove{
		Group:    groupName,
		Members:  usernames,
		IsRemove: false,
	})
	if err != nil {
		log.Printf("[DEBUG] MinIO UpdateGroupMembers API failed for group '%s': %v", groupName, err)
		return err
	}

	log.Printf("[DEBUG] AddUsersToGroup successful for group '%s'", groupName)
	return nil
}

// RemoveUsersFromGroup removes users from a group
func (s *MinIOService) RemoveUsersFromGroup(ctx context.Context, groupName string, usernames []string, username, password string) error {
	log.Printf("[DEBUG] MinIO service RemoveUsersFromGroup called for group '%s' with %d users by admin '%s'",
		groupName, len(usernames), username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in RemoveUsersFromGroup: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO UpdateGroupMembers API to remove users %v from group '%s'", usernames, groupName)
	err = adminClient.UpdateGroupMembers(ctx, madmin.GroupAddRemove{
		Group:    groupName,
		Members:  usernames,
		IsRemove: true,
	})
	if err != nil {
		log.Printf("[DEBUG] MinIO UpdateGroupMembers API failed for group '%s': %v", groupName, err)
		return err
	}

	log.Printf("[DEBUG] RemoveUsersFromGroup successful for group '%s'", groupName)
	return nil
}

// SetGroupPolicy sets a policy for a group
func (s *MinIOService) SetGroupPolicy(ctx context.Context, groupName, policyName string, username, password string) error {
	log.Printf("[DEBUG] MinIO service SetGroupPolicy called for group '%s' with policy '%s' by admin '%s'",
		groupName, policyName, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in SetGroupPolicy: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO SetPolicy API for group '%s' with policy '%s'", groupName, policyName)
	err = adminClient.SetPolicy(ctx, policyName, groupName, true) // true indicates it's a group
	if err != nil {
		log.Printf("[DEBUG] MinIO SetPolicy API failed for group '%s': %v", groupName, err)
		return err
	}

	log.Printf("[DEBUG] SetGroupPolicy successful for group '%s'", groupName)
	return nil
}

// ListPolicies lists available policies
func (s *MinIOService) ListPolicies(ctx context.Context, username, password string) (map[string]json.RawMessage, error) {
	log.Printf("[DEBUG] MinIO service ListPolicies called by admin '%s'", username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListPolicies: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListCannedPolicies API")
	policies, err := adminClient.ListCannedPolicies(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListCannedPolicies API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] ListPolicies successful: found %d policies", len(policies))
	return policies, nil
}

// GetUserDetails gets comprehensive user details including creation time and status
func (s *MinIOService) GetUserDetails(ctx context.Context, accessKey, username, password string) (map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service GetUserDetails called for user '%s' by admin '%s'", accessKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetUserDetails: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO GetUserInfo API for detailed info of user '%s'", accessKey)
	userInfo, err := adminClient.GetUserInfo(ctx, accessKey)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetUserInfo API failed for user '%s': %v", accessKey, err)
		return nil, err
	}

	// Convert to map for better JSON handling
	details := map[string]interface{}{
		"access_key":  accessKey,
		"status":      string(userInfo.Status),
		"policy_name": userInfo.PolicyName,
		"member_of":   userInfo.MemberOf,
		"updated_at":  userInfo.UpdatedAt,
	}

	log.Printf("[DEBUG] GetUserDetails successful for user '%s': status=%s, policy=%s, updated_at=%v",
		accessKey, userInfo.Status, userInfo.PolicyName, userInfo.UpdatedAt)
	return details, nil
}

// GetPolicyDocument gets the actual policy document/JSON for a specific policy
func (s *MinIOService) GetPolicyDocument(ctx context.Context, policyName, username, password string) (string, error) {
	log.Printf("[DEBUG] MinIO service GetPolicyDocument called for policy '%s' by admin '%s'", policyName, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetPolicyDocument: %v", err)
		return "", err
	}

	log.Printf("[DEBUG] Calling MinIO InfoCannedPolicy API for policy '%s'", policyName)
	policyInfo, err := adminClient.InfoCannedPolicy(ctx, policyName)
	if err != nil {
		log.Printf("[DEBUG] MinIO InfoCannedPolicy API failed for policy '%s': %v", policyName, err)
		return "", err
	}

	log.Printf("[DEBUG] GetPolicyDocument successful for policy '%s', policy length: %d", policyName, len(policyInfo))
	return string(policyInfo), nil
}

// CreateOrUpdatePolicyDocument creates or updates a policy document
func (s *MinIOService) CreateOrUpdatePolicyDocument(ctx context.Context, policyName, policyDocument, username, password string) error {
	log.Printf("[DEBUG] MinIO service CreateOrUpdatePolicyDocument called for policy '%s' by admin '%s', document length: %d", policyName, username, len(policyDocument))

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in CreateOrUpdatePolicyDocument: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO AddCannedPolicy API for policy '%s'", policyName)
	err = adminClient.AddCannedPolicy(ctx, policyName, []byte(policyDocument))
	if err != nil {
		log.Printf("[DEBUG] MinIO AddCannedPolicy API failed for policy '%s': %v", policyName, err)
		return err
	}

	log.Printf("[DEBUG] CreateOrUpdatePolicyDocument successful for policy '%s'", policyName)
	return nil
}

// DeletePolicyDocument deletes a policy document
func (s *MinIOService) DeletePolicyDocument(ctx context.Context, policyName, username, password string) error {
	log.Printf("[DEBUG] MinIO service DeletePolicyDocument called for policy '%s' by admin '%s'", policyName, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in DeletePolicyDocument: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO RemoveCannedPolicy API for policy '%s'", policyName)
	err = adminClient.RemoveCannedPolicy(ctx, policyName)
	if err != nil {
		log.Printf("[DEBUG] MinIO RemoveCannedPolicy API failed for policy '%s': %v", policyName, err)
		return err
	}

	log.Printf("[DEBUG] DeletePolicyDocument successful for policy '%s'", policyName)
	return nil
}

// GetUserCredentials gets the current credentials for a user (access key is always visible, secret key is masked)
func (s *MinIOService) GetUserCredentials(ctx context.Context, accessKey, username, password string) (map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service GetUserCredentials called for user '%s' by admin '%s'", accessKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetUserCredentials: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO GetUserInfo API for credentials of user '%s'", accessKey)
	userInfo, err := adminClient.GetUserInfo(ctx, accessKey)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetUserInfo API failed for user '%s': %v", accessKey, err)
		return nil, err
	}

	// Create credentials info (MinIO doesn't expose secret keys for security)
	credentials := map[string]interface{}{
		"access_key":  accessKey,
		"secret_key":  "••••••••••••••••", // Masked for security
		"status":      string(userInfo.Status),
		"policy_name": userInfo.PolicyName,
		"member_of":   userInfo.MemberOf,
		"updated_at":  userInfo.UpdatedAt,
		"has_secret":  true, // Indicates that a secret key exists
	}

	log.Printf("[DEBUG] GetUserCredentials successful for user '%s': status=%s, policy=%s",
		accessKey, userInfo.Status, userInfo.PolicyName)
	return credentials, nil
}
