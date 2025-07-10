package services

import (
	"context"
	"log"

	"github.com/minio/madmin-go/v3"
)

// GetUserServiceAccounts returns service accounts for a user
func (s *MinIOService) GetUserServiceAccounts(ctx context.Context, username, password string) ([]map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service GetUserServiceAccounts called for user '%s'", username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetUserServiceAccounts: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListServiceAccounts API for user '%s'", username)
	serviceAccounts, err := adminClient.ListServiceAccounts(ctx, username)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListServiceAccounts API failed for user '%s': %v", username, err)
		return nil, err
	}

	var result []map[string]interface{}
	for _, sa := range serviceAccounts.Accounts {
		serviceAccountInfo := map[string]interface{}{
			"access_key":     sa.AccessKey,
			"status":         sa.AccountStatus,
			"name":           sa.Name,
			"description":    sa.Description,
			"implied_policy": sa.ImpliedPolicy,
			"parent_user":    sa.ParentUser,
		}
		result = append(result, serviceAccountInfo)
	}

	log.Printf("[DEBUG] GetUserServiceAccounts successful for user '%s', found %d service accounts", username, len(result))
	return result, nil
}

// CreateServiceAccount creates a new service account for a user
func (s *MinIOService) CreateServiceAccount(ctx context.Context, targetUser, name, description, username, password string) (map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service CreateServiceAccount called for target user '%s' by admin '%s', name: '%s'", targetUser, username, name)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in CreateServiceAccount: %v", err)
		return nil, err
	}

	// Create service account options
	opts := madmin.AddServiceAccountReq{
		TargetUser:  targetUser,
		Name:        name,
		Description: description,
	}

	log.Printf("[DEBUG] Calling MinIO AddServiceAccount API for user '%s' with name '%s'", targetUser, name)
	creds, err := adminClient.AddServiceAccount(ctx, opts)
	if err != nil {
		log.Printf("[DEBUG] MinIO AddServiceAccount API failed for user '%s': %v", targetUser, err)
		return nil, err
	}

	result := map[string]interface{}{
		"access_key":  creds.AccessKey,
		"secret_key":  creds.SecretKey,
		"name":        name,
		"description": description,
		"parent_user": targetUser,
		"status":      "enabled",
	}

	log.Printf("[DEBUG] CreateServiceAccount successful for user '%s', access key: '%s'", targetUser, creds.AccessKey)
	return result, nil
}

// DeleteServiceAccount removes a service account
func (s *MinIOService) DeleteServiceAccount(ctx context.Context, serviceAccountKey, username, password string) error {
	log.Printf("[DEBUG] MinIO service DeleteServiceAccount called for service account '%s' by admin '%s'", serviceAccountKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in DeleteServiceAccount: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO DeleteServiceAccount API for service account '%s'", serviceAccountKey)
	err = adminClient.DeleteServiceAccount(ctx, serviceAccountKey)
	if err != nil {
		log.Printf("[DEBUG] MinIO DeleteServiceAccount API failed for service account '%s': %v", serviceAccountKey, err)
		return err
	}

	log.Printf("[DEBUG] DeleteServiceAccount successful for service account '%s'", serviceAccountKey)
	return nil
}

// GetServiceAccountInfo gets detailed information about a service account
func (s *MinIOService) GetServiceAccountInfo(ctx context.Context, serviceAccountKey, username, password string) (map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service GetServiceAccountInfo called for service account '%s' by admin '%s'", serviceAccountKey, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetServiceAccountInfo: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO InfoServiceAccount API for service account '%s'", serviceAccountKey)
	saInfo, err := adminClient.InfoServiceAccount(ctx, serviceAccountKey)
	if err != nil {
		log.Printf("[DEBUG] MinIO InfoServiceAccount API failed for service account '%s': %v", serviceAccountKey, err)
		return nil, err
	}

	result := map[string]interface{}{
		"status":         saInfo.AccountStatus,
		"name":           saInfo.Name,
		"description":    saInfo.Description,
		"parent_user":    saInfo.ParentUser,
		"implied_policy": saInfo.ImpliedPolicy,
	}

	log.Printf("[DEBUG] GetServiceAccountInfo successful for service account '%s', parent: '%s'", serviceAccountKey, saInfo.ParentUser)
	return result, nil
}

// ListServiceAccounts lists all service accounts for a given user
func (s *MinIOService) ListServiceAccounts(ctx context.Context, targetUser, username, password string) ([]map[string]interface{}, error) {
	log.Printf("[DEBUG] MinIO service ListServiceAccounts called for target user '%s' by admin '%s'", targetUser, username)

	_, adminClient, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListServiceAccounts: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListServiceAccounts API for user '%s'", targetUser)
	serviceAccounts, err := adminClient.ListServiceAccounts(ctx, targetUser)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListServiceAccounts API failed for user '%s': %v", targetUser, err)
		return nil, err
	}

	var result []map[string]interface{}
	for _, sa := range serviceAccounts.Accounts {
		serviceAccountInfo := map[string]interface{}{
			"access_key":     sa.AccessKey,
			"status":         sa.AccountStatus,
			"name":           sa.Name,
			"description":    sa.Description,
			"implied_policy": sa.ImpliedPolicy,
			"parent_user":    sa.ParentUser,
			"expiration":     sa.Expiration,
		}
		result = append(result, serviceAccountInfo)
	}

	log.Printf("[DEBUG] ListServiceAccounts successful for user '%s', found %d service accounts", targetUser, len(result))
	return result, nil
}
