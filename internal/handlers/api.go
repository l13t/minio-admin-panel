package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	minioService *services.MinIOService
}

func NewAPIHandler(minioService *services.MinIOService) *APIHandler {
	return &APIHandler{
		minioService: minioService,
	}
}

// Helper function to extract credentials from context
func (h *APIHandler) getCredentials(c *gin.Context) (string, string, error) {
	username, usernameExists := c.Get("username")
	password, passwordExists := c.Get("password")

	if !usernameExists || !passwordExists || username == nil || password == nil {
		return "", "", fmt.Errorf("missing credentials")
	}

	return username.(string), password.(string), nil
}

// GetServerInfo returns MinIO server information
func (h *APIHandler) GetServerInfo(c *gin.Context) {
	log.Printf("[DEBUG] GetServerInfo API request")

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetServerInfo: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Getting server info for user '%s'", username)
	info, err := h.minioService.GetServerInfo(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] GetServerInfo failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetServerInfo successful")
	c.JSON(http.StatusOK, gin.H{"server_info": info})
}

// GetMetrics returns MinIO metrics
func (h *APIHandler) GetMetrics(c *gin.Context) {
	log.Printf("[DEBUG] GetMetrics API request")

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetMetrics: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Getting metrics for user '%s'", username)
	metrics, err := h.minioService.GetMetrics(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] GetMetrics failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetMetrics successful")
	c.JSON(http.StatusOK, metrics)
}

// GetStorageUsage returns storage usage statistics
func (h *APIHandler) GetStorageUsage(c *gin.Context) {
	log.Printf("[DEBUG] GetStorageUsage API request")

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetStorageUsage: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Getting storage usage for user '%s'", username)

	// Get bucket statistics to calculate total storage usage
	buckets, err := h.minioService.ListBucketsQuick(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to list buckets for storage usage: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate storage usage by iterating through buckets and getting their stats
	var totalSize int64
	var totalObjects int64
	bucketsWithStats := 0

	for _, bucket := range buckets {
		log.Printf("[DEBUG] Calculating stats for bucket '%s'", bucket.Name)
		size, objectCount := h.minioService.GetBucketStatsQuick(context.Background(), username, password, bucket.Name)
		if size >= 0 && objectCount >= 0 { // Valid stats (not timeout)
			totalSize += size
			totalObjects += objectCount
			bucketsWithStats++
		}
	}

	log.Printf("[DEBUG] Storage usage calculation complete: %d bytes across %d objects in %d buckets",
		totalSize, totalObjects, bucketsWithStats)

	c.JSON(http.StatusOK, gin.H{
		"total_size":         totalSize,
		"total_objects":      totalObjects,
		"total_buckets":      len(buckets),
		"buckets_with_stats": bucketsWithStats,
		"formatted_size":     formatBytes(totalSize),
	})
}

// formatBytes converts bytes to human readable format (same as main.go)
func formatBytes(bytes int64) string {
	if bytes < 0 {
		return "N/A"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
