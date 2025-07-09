package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	minioService *services.MinIOService
}

func NewUserHandler(minioService *services.MinIOService) *UserHandler {
	return &UserHandler{
		minioService: minioService,
	}
}

// Helper function to extract credentials from context
func (h *UserHandler) getCredentials(c *gin.Context) (string, string, error) {
	username, usernameExists := c.Get("username")
	password, passwordExists := c.Get("password")

	if !usernameExists || !passwordExists || username == nil || password == nil {
		return "", "", fmt.Errorf("missing credentials")
	}

	return username.(string), password.(string), nil
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	log.Printf("[DEBUG] ListUsers request from %s %s", c.Request.Method, c.Request.URL.Path)
	log.Printf("[DEBUG] Request headers: Accept=%s", c.GetHeader("Accept"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in ListUsers: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] ListUsers for user '%s'", username)
	users, err := h.minioService.ListUsers(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] ListUsers failed for user '%s': %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] ListUsers successful for user '%s', found %d users", username, len(users))

	// Check if this is an API request
	if c.GetHeader("Accept") == "application/json" {
		log.Printf("[DEBUG] Returning JSON response with %d users", len(users))
		c.JSON(http.StatusOK, gin.H{"users": users})
		return
	}

	// Render HTML page
	log.Printf("[DEBUG] Rendering HTML page with %d users", len(users))
	permissions, _ := c.Get("permissions")
	policyName, _ := c.Get("policy_name")

	c.HTML(http.StatusOK, "users.html", gin.H{
		"title":       "Users",
		"users":       users,
		"permissions": permissions,
		"username":    username,
		"policy_name": policyName,
	})
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	var req struct {
		AccessKey string `form:"access_key" json:"access_key" binding:"required"`
		SecretKey string `form:"secret_key" json:"secret_key" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access key and secret key are required"})
		return
	}

	if err := h.minioService.CreateUser(context.Background(), req.AccessKey, req.SecretKey, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// DeleteUser handles DELETE /users/:name
func (h *UserHandler) DeleteUser(c *gin.Context) {
	username, password, err := h.getCredentials(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")

	if err := h.minioService.DeleteUser(context.Background(), accessKey, username, password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// SetUserPolicy handles PUT /users/:name/policy
func (h *UserHandler) SetUserPolicy(c *gin.Context) {
	log.Printf("[DEBUG] SetUserPolicy request for access key '%s'", c.Param("name"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in SetUserPolicy: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")

	var req struct {
		Policy string `form:"policy" json:"policy" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[DEBUG] Failed to bind request in SetUserPolicy: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy is required"})
		return
	}

	log.Printf("[DEBUG] Setting policy '%s' for user '%s' by admin '%s'", req.Policy, accessKey, username)
	if err := h.minioService.SetUserPolicy(context.Background(), accessKey, req.Policy, username, password); err != nil {
		log.Printf("[DEBUG] SetUserPolicy failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] SetUserPolicy successful for '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"message": "User policy updated successfully"})
}

// GetUser handles GET /users/:name
func (h *UserHandler) GetUser(c *gin.Context) {
	log.Printf("[DEBUG] GetUser request for access key '%s'", c.Param("name"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetUser: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")
	log.Printf("[DEBUG] Getting user details for '%s' by admin '%s'", accessKey, username)

	user, err := h.minioService.GetUser(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetUser failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Also get user policy
	policy, err := h.minioService.GetUserPolicy(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetUserPolicy failed for '%s': %v", accessKey, err)
		// Don't fail the request, just leave policy empty
		policy = ""
	}
	user.PolicyName = policy

	log.Printf("[DEBUG] GetUser successful for '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUserCredentials handles PUT /users/:name/credentials
func (h *UserHandler) UpdateUserCredentials(c *gin.Context) {
	log.Printf("[DEBUG] UpdateUserCredentials request for access key '%s'", c.Param("name"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in UpdateUserCredentials: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")

	var req struct {
		SecretKey string `form:"secret_key" json:"secret_key" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[DEBUG] Failed to bind request in UpdateUserCredentials: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret key is required"})
		return
	}

	log.Printf("[DEBUG] Updating credentials for user '%s' by admin '%s'", accessKey, username)
	if err := h.minioService.UpdateUserCredentials(context.Background(), accessKey, req.SecretKey, username, password); err != nil {
		log.Printf("[DEBUG] UpdateUserCredentials failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] UpdateUserCredentials successful for '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"message": "User credentials updated successfully"})
}

// SetUserStatus handles PUT /users/:name/status
func (h *UserHandler) SetUserStatus(c *gin.Context) {
	log.Printf("[DEBUG] SetUserStatus request for access key '%s'", c.Param("name"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in SetUserStatus: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")

	var req struct {
		Enabled bool `form:"enabled" json:"enabled"`
	}

	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[DEBUG] Failed to bind request in SetUserStatus: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("[DEBUG] Setting status for user '%s' to enabled=%t by admin '%s'", accessKey, req.Enabled, username)
	if err := h.minioService.SetUserStatus(context.Background(), accessKey, req.Enabled, username, password); err != nil {
		log.Printf("[DEBUG] SetUserStatus failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] SetUserStatus successful for '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

// GetUserPolicy handles GET /users/:name/policy
func (h *UserHandler) GetUserPolicy(c *gin.Context) {
	log.Printf("[DEBUG] GetUserPolicy request for access key '%s'", c.Param("name"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetUserPolicy: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")
	log.Printf("[DEBUG] Getting policy for user '%s' by admin '%s'", accessKey, username)

	policy, err := h.minioService.GetUserPolicy(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetUserPolicy failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetUserPolicy successful for '%s': policy=%s", accessKey, policy)
	c.JSON(http.StatusOK, gin.H{"policy": policy})
}

// ListPolicies handles GET /api/policies
func (h *UserHandler) ListPolicies(c *gin.Context) {
	log.Printf("[DEBUG] ListPolicies request")

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in ListPolicies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Listing policies by admin '%s'", username)
	policies, err := h.minioService.ListPolicies(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] ListPolicies failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert policy names to simple array for the UI
	var policyNames []string
	for policyName := range policies {
		policyNames = append(policyNames, policyName)
	}

	log.Printf("[DEBUG] ListPolicies successful: found %d policies", len(policyNames))
	c.JSON(http.StatusOK, gin.H{"policies": policyNames})
}

// GetUserDetails handles GET /users/:name/details
func (h *UserHandler) GetUserDetails(c *gin.Context) {
	log.Printf("[DEBUG] GetUserDetails request for access key '%s'", c.Param("name"))

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetUserDetails: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	accessKey := c.Param("name")
	log.Printf("[DEBUG] Getting detailed user info for '%s' by admin '%s'", accessKey, username)

	details, err := h.minioService.GetUserDetails(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetUserDetails failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetUserDetails successful for '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"details": details})
}

// GetUserCredentials handles GET /users/:name/credentials
func (h *UserHandler) GetUserCredentials(c *gin.Context) {
	accessKey := c.Param("name")
	log.Printf("[DEBUG] GetUserCredentials request for access key '%s'", accessKey)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetUserCredentials: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Getting credentials for user '%s' by admin '%s'", accessKey, username)

	credentials, err := h.minioService.GetUserCredentials(context.Background(), accessKey, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetUserCredentials failed for '%s': %v", accessKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetUserCredentials successful for '%s'", accessKey)
	c.JSON(http.StatusOK, gin.H{"credentials": credentials})
}
