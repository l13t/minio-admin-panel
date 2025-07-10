package services

import (
	"context"
	"encoding/json"
	"log"
)

// ListPolicies lists all available policies
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
