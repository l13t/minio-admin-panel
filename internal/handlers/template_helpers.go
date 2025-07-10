package handlers

import (
	"fmt"
	"log"
	"strings"
	"minio-admin-panel/internal/i18n"
	"minio-admin-panel/internal/middleware"

	"github.com/gin-gonic/gin"
)

// TemplateData represents common template data
type TemplateData struct {
	Title       string
	Username    interface{}
	PolicyName  interface{}
	Permissions interface{}
	Language    string
	T           func(string) string
	TWithParams func(string, ...interface{}) string
}

// NewTemplateData creates template data with translation functions
func NewTemplateData(c *gin.Context, title string) TemplateData {
	lang := middleware.GetLanguage(c)
	username, _ := c.Get("username")
	policyName, _ := c.Get("policy_name")
	permissions := middleware.GetUserPermissions(c)

	return TemplateData{
		Title:       title,
		Username:    username,
		PolicyName:  policyName,
		Permissions: permissions,
		Language:    lang,
		T: func(key string) string {
			return i18n.T(lang, key)
		},
		TWithParams: func(key string, params ...interface{}) string {
			// Convert variadic params to a map for the go-i18n library
			if len(params) > 0 && len(params)%2 == 0 {
				paramsMap := make(map[string]interface{})
				for i := 0; i < len(params); i += 2 {
					if key, ok := params[i].(string); ok {
						paramsMap[key] = params[i+1]
					}
				}
				return i18n.TWithParams(lang, key, paramsMap)
			}
			return i18n.T(lang, key)
		},
	}
}

// RenderWithTranslations renders template with translation support
func RenderWithTranslations(c *gin.Context, templateName string, data gin.H) {
	lang := middleware.GetLanguage(c)
	log.Printf("[DEBUG i18n] Rendering template %s with language: %s", templateName, lang)

	// Set up translation context for template functions
	contextID := SetTemplateContext(lang)
	SetCurrentContext(contextID)
	defer func() {
		CleanupTemplateContext(contextID)
	}()

	// Add translation functions to data
	if data == nil {
		data = gin.H{}
	}

	// Auto-detect current page from template name
	if templateName != "" {
		switch {
		case strings.Contains(templateName, "dashboard"):
			data["currentPage"] = "dashboard"
		case strings.Contains(templateName, "users"):
			data["currentPage"] = "users"
		case strings.Contains(templateName, "groups"):
			data["currentPage"] = "groups"
		case strings.Contains(templateName, "policies"):
			data["currentPage"] = "policies"
		case strings.Contains(templateName, "buckets"):
			data["currentPage"] = "buckets"
		case strings.Contains(templateName, "settings"):
			data["currentPage"] = "settings"
		default:
			data["currentPage"] = ""
		}
	}

	// Debug: print existing data keys
	log.Printf("[DEBUG i18n] Existing data keys: ")
	for key := range data {
		fmt.Printf("%s, ", key)
	}
	fmt.Printf("\n")

	data["language"] = lang

	// Translate all string values that look like translation keys
	translatedData := gin.H{}
	for key, value := range data {
		switch v := value.(type) {
		case string:
			// If the value looks like a translation key (contains dots), translate it
			if strings.Contains(v, ".") && !strings.Contains(v, " ") && !strings.Contains(v, "/") {
				translated := i18n.T(lang, v)
				log.Printf("[DEBUG i18n] Auto-translated '%s': '%s' -> '%s'", key, v, translated)
				translatedData[key] = translated
			} else {
				translatedData[key] = v
			}
		default:
			translatedData[key] = v
		}
	}

	// Create translation functions that work with the current language
	translatedData["t"] = func(key string) string {
		log.Printf("[DEBUG i18n] Template function t() called with key: '%s' in language: '%s'", key, lang)
		translated := i18n.T(lang, key)
		if translated == key {
			log.Printf("[DEBUG i18n] WARNING: Missing translation for key '%s' in language '%s'", key, lang)
		} else {
			log.Printf("[DEBUG i18n] SUCCESS: Translated '%s' -> '%s'", key, translated)
		}
		return translated
	}

	translatedData["tWithParams"] = func(key string, params ...interface{}) string {
		log.Printf("[DEBUG i18n] Template function tWithParams() called - key: '%s', params count: %d", key, len(params))
		if len(params) > 0 && len(params)%2 == 0 {
			paramsMap := make(map[string]interface{})
			for i := 0; i < len(params); i += 2 {
				if paramKey, ok := params[i].(string); ok {
					paramsMap[paramKey] = params[i+1]
					log.Printf("[DEBUG i18n] Param: %s = %v", paramKey, params[i+1])
				}
			}
			result := i18n.TWithParams(lang, key, paramsMap)
			log.Printf("[DEBUG i18n] tWithParams result: '%s' -> '%s'", key, result)
			return result
		}
		log.Printf("[DEBUG i18n] No valid params, falling back to simple translation")
		return i18n.T(lang, key)
	}

	translatedData["tCount"] = func(key string, count int) string {
		log.Printf("[DEBUG i18n] tCount called - key: '%s', count: %d", key, count)
		result := i18n.TWithCount(lang, key, count)
		log.Printf("[DEBUG i18n] tCount result: '%s' -> '%s'", key, result)
		return result
	}

	// Add common template data if not already present
	if _, exists := translatedData["username"]; !exists {
		username, _ := c.Get("username")
		translatedData["username"] = username
	}
	if _, exists := translatedData["policy_name"]; !exists {
		policyName, _ := c.Get("policy_name")
		translatedData["policy_name"] = policyName
	}
	if _, exists := translatedData["permissions"]; !exists {
		translatedData["permissions"] = middleware.GetUserPermissions(c)
	}

	log.Printf("[DEBUG i18n] Final template data keys: ")
	for key := range translatedData {
		fmt.Printf("%s, ", key)
	}
	fmt.Printf("\n")
	log.Printf("[DEBUG i18n] Calling c.HTML with template: %s", templateName)

	c.HTML(200, templateName, translatedData)
}
