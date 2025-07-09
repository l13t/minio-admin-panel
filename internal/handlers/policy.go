package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"minio-admin-panel/internal/services"

	"github.com/gin-gonic/gin"
)

type PolicyHandler struct {
	minioService *services.MinIOService
}

func NewPolicyHandler(minioService *services.MinIOService) *PolicyHandler {
	return &PolicyHandler{
		minioService: minioService,
	}
}

// Helper function to extract credentials from context
func (h *PolicyHandler) getCredentials(c *gin.Context) (string, string, error) {
	username, usernameExists := c.Get("username")
	password, passwordExists := c.Get("password")

	if !usernameExists || !passwordExists || username == nil || password == nil {
		return "", "", fmt.Errorf("missing credentials")
	}

	return username.(string), password.(string), nil
}

// ListPolicies handles GET /policies
func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	log.Printf("[DEBUG] ListPolicies request from %s %s", c.Request.Method, c.Request.URL.Path)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in ListPolicies: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] ListPolicies for admin '%s'", username)
	policies, err := h.minioService.ListPolicies(context.Background(), username, password)
	if err != nil {
		log.Printf("[DEBUG] ListPolicies failed for admin '%s': %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] ListPolicies successful for admin '%s', found %d policies", username, len(policies))

	// Check if this is an API request
	if c.GetHeader("Accept") == "application/json" {
		log.Printf("[DEBUG] Returning JSON response with %d policies", len(policies))
		// Convert policy names to simple array for the UI
		var policyNames []string
		for policyName := range policies {
			policyNames = append(policyNames, policyName)
		}
		c.JSON(http.StatusOK, gin.H{"policies": policyNames})
		return
	}

	// Render HTML page
	log.Printf("[DEBUG] Rendering HTML page with %d policies", len(policies))
	permissions, _ := c.Get("permissions")
	policyName, _ := c.Get("policy_name")

	// Convert policies map to slice for easier template iteration
	var policyList []map[string]interface{}
	for name := range policies {
		policyList = append(policyList, map[string]interface{}{
			"name": name,
		})
	}

	c.HTML(http.StatusOK, "policies.html", gin.H{
		"title":       "Policies",
		"policies":    policyList,
		"permissions": permissions,
		"username":    username,
		"policy_name": policyName,
	})
}

// GetPolicyDocument handles GET /policies/:name
func (h *PolicyHandler) GetPolicyDocument(c *gin.Context) {
	policyName := c.Param("name")
	log.Printf("[DEBUG] GetPolicyDocument request for policy '%s'", policyName)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in GetPolicyDocument: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Getting policy document for '%s' by admin '%s'", policyName, username)

	policyDocument, err := h.minioService.GetPolicyDocument(context.Background(), policyName, username, password)
	if err != nil {
		log.Printf("[DEBUG] GetPolicyDocument failed for '%s': %v", policyName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] GetPolicyDocument successful for '%s'", policyName)
	c.JSON(http.StatusOK, gin.H{"policy": policyDocument})
}

// CreateOrUpdatePolicy handles POST/PUT /policies/:name
func (h *PolicyHandler) CreateOrUpdatePolicy(c *gin.Context) {
	policyName := c.Param("name")
	log.Printf("[DEBUG] CreateOrUpdatePolicy request for policy '%s'", policyName)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in CreateOrUpdatePolicy: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	var req struct {
		Policy string `form:"policy" json:"policy" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[DEBUG] Failed to bind request in CreateOrUpdatePolicy: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy document is required"})
		return
	}

	// Validate JSON
	var policyJSON interface{}
	if err := json.Unmarshal([]byte(req.Policy), &policyJSON); err != nil {
		log.Printf("[DEBUG] Invalid JSON in CreateOrUpdatePolicy: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy must be valid JSON"})
		return
	}

	log.Printf("[DEBUG] Creating/updating policy '%s' by admin '%s', policy length: %d", policyName, username, len(req.Policy))

	if err := h.minioService.CreateOrUpdatePolicyDocument(context.Background(), policyName, req.Policy, username, password); err != nil {
		log.Printf("[DEBUG] CreateOrUpdatePolicy failed for '%s': %v", policyName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] CreateOrUpdatePolicy successful for '%s'", policyName)
	c.JSON(http.StatusOK, gin.H{"message": "Policy updated successfully"})
}

// DeletePolicy handles DELETE /policies/:name
func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	policyName := c.Param("name")
	log.Printf("[DEBUG] DeletePolicy request for policy '%s'", policyName)

	username, password, err := h.getCredentials(c)
	if err != nil {
		log.Printf("[DEBUG] Failed to get credentials in DeletePolicy: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
		return
	}

	log.Printf("[DEBUG] Deleting policy '%s' by admin '%s'", policyName, username)

	if err := h.minioService.DeletePolicyDocument(context.Background(), policyName, username, password); err != nil {
		log.Printf("[DEBUG] DeletePolicy failed for '%s': %v", policyName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] DeletePolicy successful for '%s'", policyName)
	c.JSON(http.StatusOK, gin.H{"message": "Policy deleted successfully"})
}
