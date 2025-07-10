package services

import (
	"context"
	"log"

	"github.com/minio/madmin-go/v3"
)

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
