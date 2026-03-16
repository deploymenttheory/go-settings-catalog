package converter

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// JSONExporter exports conversion results to JSON format
type JSONExporter struct{}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{}
}

// ExportFromResult converts a ConversionResult to JSON
// If any settings were skipped or no settings matched, falls back to custom configuration format
func (e *JSONExporter) ExportFromResult(result *ConversionResult) ([]byte, error) {
	// Check if we need to fall back to custom configuration
	if result.SettingCount == 0 || len(result.SkippedKeys) > 0 || len(result.SkippedPayloads) > 0 {
		return e.exportAsCustomConfig(result)
	}

	// Use Settings Catalog format (already in result.OutputJSON)
	return result.OutputJSON, nil
}

// exportAsCustomConfig generates a macOSCustomConfiguration JSON
func (e *JSONExporter) exportAsCustomConfig(result *ConversionResult) ([]byte, error) {
	// Base64 encode the original mobileconfig
	encodedPayload := base64.StdEncoding.EncodeToString(result.OriginalData)

	now := time.Now().UTC().Format(time.RFC3339)

	customConfig := map[string]any{
		"@odata.type":           "#microsoft.graph.macOSCustomConfiguration",
		"@odata.context":        "https://graph.microsoft.com/beta/$metadata#deviceManagement/deviceConfigurations/$entity",
		"displayName":           result.ProfileName,
		"description":           result.ProfileDescription,
		"createdDateTime":       now,
		"lastModifiedDateTime":  now,
		"version":               1,
		"payloadName":           result.ProfileName,
		"payloadFileName":       result.ProfileName + ".mobileconfig",
		"payload":               encodedPayload,
		"deploymentChannel":     "deviceChannel",
		"roleScopeTagIds":       []string{"0"},
		"supportsScopeTags":     true,
	}

	return json.MarshalIndent(customConfig, "", "  ")
}
