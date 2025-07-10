package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type BucketHandler struct {
	minioService *services.MinIOService
}

func NewBucketHandler(minioService *services.MinIOService) *BucketHandler {
	return &BucketHandler{
		minioService: minioService,
	}
}

// Helper function to extract credentials from context
func (h *BucketHandler) getCredentials(c *gin.Context) (string, string, error) {
	username, usernameExists := c.Get("username")
	password, passwordExists := c.Get("password")

	if !usernameExists || !passwordExists || username == nil || password == nil {
		return "", "", fmt.Errorf("missing credentials")
	}

	return username.(string), password.(string), nil
}

// ListBuckets handles GET /buckets
func (h *BucketHandler) ListBuckets(c *gin.Context) {
	log.Printf("[DEBUG] ListBuckets request from %s %s", c.Request.Method, c.Request.URL.Path)
	log.Printf("[DEBUG] Request headers: Accept=%s", c.GetHeader("Accept"))

	// Extract credentials from context
	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in ListBuckets: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] ListBuckets for user '%s'", username)
	buckets, err := h.minioService.ListBuckets(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] ListBuckets failed for user '%s': %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] ListBuckets successful for user '%s', found %d buckets", username, len(buckets))

	// Check if this is an API request
	if c.GetHeader("Accept") == "application/json" {
		log.Printf("[DEBUG] Returning JSON response with %d buckets", len(buckets))
		c.JSON(http.StatusOK, gin.H{"buckets": buckets})
		return
	}

	// Render HTML page
	log.Printf("[DEBUG] Rendering HTML page with %d buckets", len(buckets))
	permissions, _ := c.Get("permissions")
	policyName, _ := c.Get("policy_name")
	RenderWithTranslations(c, "buckets.html", gin.H{
		"title":       "buckets.title",
		"buckets":     buckets,
		"permissions": permissions,
		"username":    username,
		"policy_name": policyName,
	})
}

// CreateBucket handles POST /buckets
func (h *BucketHandler) CreateBucket(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	var req struct {
		Name string `form:"name" json:"name" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bucket name is required"})
		return
	}

	if err := h.minioService.CreateBucket(context.Background(), req.Name, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Bucket created successfully"})
}

// DeleteBucket handles DELETE /buckets/:name
func (h *BucketHandler) DeleteBucket(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	bucketName := c.Param("name")

	if err := h.minioService.DeleteBucket(context.Background(), bucketName, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bucket deleted successfully"})
}

// GetBucketPolicy handles GET /buckets/:name/policy
func (h *BucketHandler) GetBucketPolicy(c *gin.Context) {
	bucketName := c.Param("name")
	log.Printf("[DEBUG] GetBucketPolicy request for bucket '%s'", bucketName)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetBucketPolicy: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Getting policy for bucket '%s' by user '%s'", bucketName, username)
	policy, err := h.minioService.GetBucketPolicy(context.Background(), bucketName, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetBucketPolicy failed for bucket '%s': %v", bucketName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetBucketPolicy successful for bucket '%s', policy length: %d", bucketName, len(policy))
	c.JSON(http.StatusOK, gin.H{"policy": policy})
}

// SetBucketPolicy handles PUT /buckets/:name/policy
func (h *BucketHandler) SetBucketPolicy(c *gin.Context) {
	bucketName := c.Param("name")
	log.Printf("[DEBUG] SetBucketPolicy request for bucket '%s'", bucketName)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in SetBucketPolicy: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	var req struct {
		Policy string `form:"policy" json:"policy"`
	}

	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[DEBUG] Failed to bind request in SetBucketPolicy: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("[DEBUG] Setting policy for bucket '%s' by user '%s', policy length: %d", bucketName, username, len(req.Policy))

	if err := h.minioService.SetBucketPolicy(context.Background(), bucketName, req.Policy, username, password); err != nil {
		log.Printf("[DEBUG] SetBucketPolicy failed for bucket '%s': %v", bucketName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] SetBucketPolicy successful for bucket '%s'", bucketName)
	c.JSON(http.StatusOK, gin.H{"message": "Bucket policy updated successfully"})
}
