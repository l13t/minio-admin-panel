package middleware

import (
	"fmt"
	"minio-admin-panel/internal/i18n"
	"strings"

	"github.com/gin-gonic/gin"
)

// LanguageMiddleware handles language detection and setting
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := getLanguageFromRequest(c)
		c.Set("language", lang)

		// Always ensure the cookie is set or updated
		if lang != "" {
			// Set/update the language cookie with each request to refresh expiry
			// Secure: false to allow HTTP testing, HttpOnly: false to allow JS access
			c.SetCookie("language", lang, 86400*30, "/", "", false, false) // 30 days
			fmt.Printf("Set language cookie: %s\n", lang)
		}

		c.Next()
	}
}

// getLanguageFromRequest determines the language from various sources
func getLanguageFromRequest(c *gin.Context) string {
	// 1. Check for explicit language parameter
	if lang := c.Query("lang"); lang != "" {
		if isValidLanguage(lang) {
			fmt.Printf("Using language from URL parameter: %s\n", lang)
			return lang
		}
	}

	// 2. Check for language cookie
	if lang, err := c.Cookie("language"); err == nil && lang != "" {
		if isValidLanguage(lang) {
			fmt.Printf("Using language from cookie: %s\n", lang)
			return lang
		}
	}

	// 3. Check Accept-Language header
	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if isValidLanguage(lang) {
			return lang
		}
	}

	// 4. Default to English
	return "en"
}

// parseAcceptLanguage parses the Accept-Language header
func parseAcceptLanguage(acceptLang string) string {
	// Simple parsing - take the first language preference
	languages := strings.Split(acceptLang, ",")
	if len(languages) > 0 {
		lang := strings.TrimSpace(languages[0])
		// Remove quality values (e.g., "en-US;q=0.9" -> "en-US")
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}
		// Convert to lowercase and take the primary language
		lang = strings.ToLower(lang)
		if idx := strings.Index(lang, "-"); idx != -1 {
			lang = lang[:idx]
		}
		return lang
	}
	return "en"
}

// isValidLanguage checks if the language is supported
func isValidLanguage(lang string) bool {
	supportedLanguages := i18n.GetAvailableLanguages()
	for _, supported := range supportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// GetLanguage returns the current language from context
func GetLanguage(c *gin.Context) string {
	if lang, exists := c.Get("language"); exists {
		return lang.(string)
	}
	return "en"
}

// T is a template helper function for translations
func T(c *gin.Context, key string) string {
	lang := GetLanguage(c)
	return i18n.T(lang, key)
}

// TWithParams is a template helper function for translations with parameters
func TWithParams(c *gin.Context, key string, params ...interface{}) string {
	lang := GetLanguage(c)

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
}

// TWithCount is a template helper function for translations with pluralization
func TWithCount(c *gin.Context, key string, count int) string {
	lang := GetLanguage(c)
	return i18n.TWithCount(lang, key, count)
}
