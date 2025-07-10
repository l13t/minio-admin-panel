package services

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/madmin-go/v3"
)

// ListUsers returns a list of all users
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

// CreateUser creates a new user
func (s *MinIOService) CreateUser(ctx context.Context, accessKey, secretKey, username, password string) error {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return adminClient.AddUser(ctx, accessKey, secretKey)
}

// DeleteUser deletes an existing user
func (s *MinIOService) DeleteUser(ctx context.Context, accessKey, username, password string) error {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return adminClient.RemoveUser(ctx, accessKey)
}

// SetUserPolicy sets the policy for a user
func (s *MinIOService) SetUserPolicy(ctx context.Context, accessKey, policyName, username, password string) error {
	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return adminClient.SetPolicy(ctx, policyName, accessKey, false)
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

// GetUserDetails gets detailed information about a user
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

// GetUserCredentials gets user credential information (secret key is masked for security)
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
