package handlers

import (
	"sync"
	"minio-admin-panel/internal/i18n"
)

// Template context for storing request-specific data accessible to template functions
var (
	templateContexts = make(map[uint64]string) // Maps goroutine ID to language
	contextMutex     sync.RWMutex
	contextCounter   uint64
	contextMapping   = make(map[uint64]uint64) // Maps context ID to goroutine-like ID
	contextMapMutex  sync.RWMutex
)

// SetTemplateContext stores the language for the current template rendering context
func SetTemplateContext(language string) uint64 {
	contextMapMutex.Lock()
	contextCounter++
	contextID := contextCounter
	contextMapMutex.Unlock()

	contextMutex.Lock()
	templateContexts[contextID] = language
	contextMutex.Unlock()

	return contextID
}

// GetTemplateLanguage retrieves the language for the current context
func GetTemplateLanguage(contextID uint64) string {
	contextMutex.RLock()
	defer contextMutex.RUnlock()
	if lang, exists := templateContexts[contextID]; exists {
		return lang
	}
	return "en" // fallback
}

// CleanupTemplateContext removes the context after rendering
func CleanupTemplateContext(contextID uint64) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	delete(templateContexts, contextID)
}

// Current context ID for the active template rendering (using thread-local-like storage)
var currentContextID uint64
var currentContextMutex sync.RWMutex

// SetCurrentContext sets the context for the current goroutine
func SetCurrentContext(contextID uint64) {
	currentContextMutex.Lock()
	currentContextID = contextID
	currentContextMutex.Unlock()
}

// GetCurrentContext gets the context for the current goroutine
func GetCurrentContext() uint64 {
	currentContextMutex.RLock()
	defer currentContextMutex.RUnlock()
	return currentContextID
}

// TranslateInTemplate is called from templates to translate keys
func TranslateInTemplate(key string) string {
	contextID := GetCurrentContext()
	lang := GetTemplateLanguage(contextID)
	return i18n.T(lang, key)
}

// TranslateInTemplateWithParams is called from templates to translate keys with parameters
func TranslateInTemplateWithParams(key string, params ...interface{}) string {
	contextID := GetCurrentContext()
	lang := GetTemplateLanguage(contextID)

	if len(params) > 0 && len(params)%2 == 0 {
		paramsMap := make(map[string]interface{})
		for i := 0; i < len(params); i += 2 {
			if paramKey, ok := params[i].(string); ok {
				paramsMap[paramKey] = params[i+1]
			}
		}
		return i18n.TWithParams(lang, key, paramsMap)
	}
	return i18n.T(lang, key)
}

// TranslateInTemplateWithCount is called from templates to translate keys with count
func TranslateInTemplateWithCount(key string, count int) string {
	contextID := GetCurrentContext()
	lang := GetTemplateLanguage(contextID)
	return i18n.TWithCount(lang, key, count)
}
