package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Flattens nested JSON into go-i18n format
func flattenTranslations(prefix string, data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested objects
			flattened := flattenTranslations(newKey, v)
			for k, v := range flattened {
				result[k] = v
			}
		default:
			// Convert the value to proper go-i18n format
			result[newKey] = map[string]interface{}{
				"other": value,
			}
		}
	}

	return result
}

func main() {
	// Directory containing translation files
	translationsDir := "translations"
	outputDir := "translations/i18n"

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	// Process all JSON files in the translations directory
	files, err := filepath.Glob(filepath.Join(translationsDir, "*.json"))
	if err != nil {
		fmt.Printf("Failed to list translation files: %v\n", err)
		return
	}

	for _, file := range files {
		// Read the translation file
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read file %s: %v\n", file, err)
			continue
		}

		// Parse the JSON data
		var translations map[string]interface{}
		if err := json.Unmarshal(data, &translations); err != nil {
			fmt.Printf("Failed to parse JSON from file %s: %v\n", file, err)
			continue
		}

		// Flatten the translations
		flattened := flattenTranslations("", translations)

		// Convert to go-i18n format
		fileName := filepath.Base(file)
		outputFile := filepath.Join(outputDir, fileName)

		// Write the flattened translations to the output file
		jsonData, err := json.MarshalIndent(flattened, "", "    ")
		if err != nil {
			fmt.Printf("Failed to marshal JSON for file %s: %v\n", file, err)
			continue
		}

		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			fmt.Printf("Failed to write output file %s: %v\n", outputFile, err)
			continue
		}

		fmt.Printf("Successfully converted %s to %s\n", file, outputFile)
	}
}
