package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Bundle holds localization bundles for all supported languages
var (
	bundle          *goi18n.Bundle
	localizers      map[string]*goi18n.Localizer
	mutex           sync.RWMutex
	defaultLanguage string
	availableLangs  []string
)

// Init initializes the i18n system with the specified default language
func Init(defLang string) {
	bundle = goi18n.NewBundle(language.Make(defLang))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	localizers = make(map[string]*goi18n.Localizer)
	defaultLanguage = defLang

	// Always ensure default language has a localizer
	localizers[defLang] = goi18n.NewLocalizer(bundle, defLang)
	availableLangs = []string{defLang}

	fmt.Printf("Initialized i18n system with default language: %s\n", defLang)
}

// LoadDir loads all translation files from the specified directory
func LoadDir(dir string) error {
	if bundle == nil {
		return fmt.Errorf("i18n bundle not initialized, call Init() first")
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list translation files: %w", err)
	}

	mutex.Lock()
	defer mutex.Unlock()

	for _, file := range files {
		// Extract language code from filename (e.g., "en.json" -> "en")
		lang := strings.TrimSuffix(filepath.Base(file), ".json")

		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading translation file %s: %v\n", file, err)
			continue
		}

		// Load the file into the bundle
		_, err = bundle.ParseMessageFileBytes(data, file)
		if err != nil {
			fmt.Printf("Error parsing translation file %s: %v\n", file, err)
			continue
		}

		// Create a localizer for this language
		localizers[lang] = goi18n.NewLocalizer(bundle, lang)

		// Track available languages
		if !contains(availableLangs, lang) {
			availableLangs = append(availableLangs, lang)
		}

		fmt.Printf("Loaded translations for language: %s from %s\n", lang, file)
	}

	return nil
}

// T translates a message for the given language
func T(lang, msgID string) string {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Printf("[DEBUG] T() called with lang='%s', msgID='%s'\n", lang, msgID)

	// Get the appropriate localizer
	localizer, exists := localizers[lang]
	if !exists {
		// Fall back to default if language not found
		fmt.Printf("[DEBUG] Language '%s' not found, using default '%s'\n", lang, defaultLanguage)
		localizer = localizers[defaultLanguage]
	}

	// Get the translation
	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID: msgID,
	})

	if err != nil {
		fmt.Printf("[DEBUG] Translation error for '%s' in language '%s': %v\n", msgID, lang, err)

		// Try with default language if different
		if lang != defaultLanguage {
			fmt.Printf("[DEBUG] Attempting fallback to default language '%s' for key '%s'\n", defaultLanguage, msgID)
			defMsg, defErr := localizers[defaultLanguage].Localize(&goi18n.LocalizeConfig{
				MessageID: msgID,
			})

			if defErr == nil {
				fmt.Printf("[DEBUG] Fallback successful for '%s': '%s'\n", msgID, defMsg)
				return defMsg
			}
			fmt.Printf("[DEBUG] Fallback also failed for '%s': %v\n", msgID, defErr)
		}

		// Return the key if all attempts fail
		fmt.Printf("[DEBUG] All translation attempts failed for '%s', returning key\n", msgID)
		return msgID
	}

	fmt.Printf("[DEBUG] Translation successful for '%s': '%s'\n", msgID, msg)
	return msg
}

// TWithParams translates a message with template parameters
func TWithParams(lang, msgID string, params map[string]interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()

	localizer, exists := localizers[lang]
	if !exists {
		localizer = localizers[defaultLanguage]
	}

	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: params,
	})

	if err != nil {
		fmt.Printf("Translation error for %s with params in %s: %v\n", msgID, lang, err)

		// Try with default language if different
		if lang != defaultLanguage {
			defMsg, defErr := localizers[defaultLanguage].Localize(&goi18n.LocalizeConfig{
				MessageID:    msgID,
				TemplateData: params,
			})

			if defErr == nil {
				return defMsg
			}
		}

		// Return the key if all attempts fail
		return msgID
	}

	return msg
}

// TWithCount translates a message with plural forms
func TWithCount(lang, msgID string, count int) string {
	mutex.RLock()
	defer mutex.RUnlock()

	localizer, exists := localizers[lang]
	if !exists {
		localizer = localizers[defaultLanguage]
	}

	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:   msgID,
		PluralCount: count,
	})

	if err != nil {
		// Fall back to default language
		if lang != defaultLanguage {
			defMsg, defErr := localizers[defaultLanguage].Localize(&goi18n.LocalizeConfig{
				MessageID:   msgID,
				PluralCount: count,
			})

			if defErr == nil {
				return defMsg
			}
		}

		// Return key if all attempts fail
		return msgID
	}

	return msg
}

// GetAvailableLanguages returns a list of available languages
func GetAvailableLanguages() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	return append([]string{}, availableLangs...)
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
