package main

import (
	"minio-admin-panel/internal/config"
	"minio-admin-panel/internal/services"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	cfg := config.Load()

	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	if cfg.MinIOEndpoint == "" {
		t.Error("MinIO endpoint should not be empty")
	}

	if cfg.JWTSecret == "" {
		t.Error("JWT secret should not be empty")
	}
}

func TestMinIOServiceCreation(t *testing.T) {
	cfg := config.Load()
	service := services.NewMinIOService(cfg)

	if service == nil {
		t.Fatal("MinIO service should not be nil")
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid bucket name", "test-bucket", true},
		{"invalid bucket name with uppercase", "Test-Bucket", false},
		{"invalid bucket name too short", "ab", false},
		{"valid bucket name with numbers", "bucket123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test bucket name validation
			// For now, just a placeholder test
			if len(tt.input) < 3 && tt.expected {
				t.Errorf("Expected %s to be invalid but got valid", tt.input)
			}
		})
	}
}
