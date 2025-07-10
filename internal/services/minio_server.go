package services

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetServerInfo retrieves MinIO server information
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

// GetMetrics retrieves basic server metrics and status
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
