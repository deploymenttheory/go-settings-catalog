package converter

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONExportFromResult_SettingsCatalog(t *testing.T) {
	exporter := NewJSONExporter()

	outputJSON := []byte(`{
		"name": "Test Profile",
		"description": "Test description",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": []
				}
			}
		]
	}`)

	result := &ConversionResult{
		OutputJSON:         outputJSON,
		SettingCount:       1,
		SkippedPayloads:    []string{},
		SkippedKeys:        []SkippedKey{},
		ProfileName:        "Test Profile",
		ProfileDescription: "Test description",
		UsedCustomConfig:   false,
	}

	jsonOutput, err := exporter.ExportFromResult(result)
	if err != nil {
		t.Fatalf("ExportFromResult failed: %v", err)
	}

	// Should return the original OutputJSON
	if string(jsonOutput) != string(outputJSON) {
		t.Error("Expected output to match original OutputJSON for Settings Catalog")
	}

	var parsed map[string]any
	err = json.Unmarshal(jsonOutput, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if parsed["name"] != "Test Profile" {
		t.Errorf("Expected name 'Test Profile', got '%v'", parsed["name"])
	}
}

func TestJSONExportFromResult_CustomConfig(t *testing.T) {
	exporter := NewJSONExporter()

	originalData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.unknown.payload</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Unknown</string>
</dict>
</plist>`)

	result := &ConversionResult{
		OutputJSON:         []byte(`{"name":"Test Unknown","description":"","platforms":"macOS","technologies":"mdm","settings":[]}`),
		SettingCount:       0,
		SkippedPayloads:    []string{"com.unknown.payload"},
		OriginalData:       originalData,
		ProfileName:        "Test Unknown",
		ProfileDescription: "Test custom config",
		UsedCustomConfig:   true,
	}

	jsonOutput, err := exporter.ExportFromResult(result)
	if err != nil {
		t.Fatalf("ExportFromResult failed: %v", err)
	}

	var parsed map[string]any
	err = json.Unmarshal(jsonOutput, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	// Verify custom configuration structure
	if parsed["@odata.type"] != "#microsoft.graph.macOSCustomConfiguration" {
		t.Errorf("Expected custom configuration type, got %v", parsed["@odata.type"])
	}

	if parsed["displayName"] != "Test Unknown" {
		t.Errorf("Expected displayName 'Test Unknown', got '%v'", parsed["displayName"])
	}

	if parsed["description"] != "Test custom config" {
		t.Errorf("Expected description 'Test custom config', got '%v'", parsed["description"])
	}

	if parsed["deploymentChannel"] != "deviceChannel" {
		t.Errorf("Expected deploymentChannel 'deviceChannel', got '%v'", parsed["deploymentChannel"])
	}

	// Verify payload is base64 encoded
	payloadStr, ok := parsed["payload"].(string)
	if !ok {
		t.Fatal("Expected payload to be a string")
	}

	decoded, err := base64.StdEncoding.DecodeString(payloadStr)
	if err != nil {
		t.Fatalf("Failed to decode base64 payload: %v", err)
	}

	if !strings.Contains(string(decoded), "PayloadContent") {
		t.Error("Expected decoded payload to contain original XML")
	}
}

func TestJSONExportFromResult_CustomConfigWithSkippedKeys(t *testing.T) {
	exporter := NewJSONExporter()

	originalData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>unknownkey</key>
			<string>value</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Skipped</string>
</dict>
</plist>`)

	result := &ConversionResult{
		OutputJSON:  []byte(`{"name":"Test Skipped","description":"","platforms":"macOS","technologies":"mdm","settings":[]}`),
		SettingCount: 0,
		SkippedKeys: []SkippedKey{
			{
				Path:  "com.apple.dock.unknownkey",
				Value: "value",
			},
		},
		OriginalData:       originalData,
		ProfileName:        "Test Skipped",
		ProfileDescription: "",
		UsedCustomConfig:   true,
	}

	jsonOutput, err := exporter.ExportFromResult(result)
	if err != nil {
		t.Fatalf("ExportFromResult failed: %v", err)
	}

	var parsed map[string]any
	err = json.Unmarshal(jsonOutput, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if parsed["@odata.type"] != "#microsoft.graph.macOSCustomConfiguration" {
		t.Error("Expected custom configuration type when keys are skipped")
	}
}

func TestExportAsCustomConfig_AllFields(t *testing.T) {
	exporter := NewJSONExporter()

	originalData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>Test</key>
	<string>Data</string>
</dict>
</plist>`)

	result := &ConversionResult{
		OriginalData:       originalData,
		ProfileName:        "Test Profile Name",
		ProfileDescription: "Test Profile Description",
	}

	jsonOutput, err := exporter.exportAsCustomConfig(result)
	if err != nil {
		t.Fatalf("exportAsCustomConfig failed: %v", err)
	}

	var parsed map[string]any
	err = json.Unmarshal(jsonOutput, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	expectedFields := map[string]string{
		"@odata.type":      "#microsoft.graph.macOSCustomConfiguration",
		"displayName":      "Test Profile Name",
		"description":      "Test Profile Description",
		"payloadName":      "Test Profile Name",
		"payloadFileName":  "Test Profile Name.mobileconfig",
		"deploymentChannel": "deviceChannel",
	}

	for field, expectedValue := range expectedFields {
		actualValue, ok := parsed[field]
		if !ok {
			t.Errorf("Expected field %s to be present", field)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("Field %s: expected %q, got %q", field, expectedValue, actualValue)
		}
	}

	// Verify version is integer
	version, ok := parsed["version"].(float64)
	if !ok || version != 1 {
		t.Errorf("Expected version to be 1, got %v", parsed["version"])
	}

	// Verify roleScopeTagIds is array
	roleScopeTagIds, ok := parsed["roleScopeTagIds"].([]any)
	if !ok || len(roleScopeTagIds) != 1 || roleScopeTagIds[0] != "0" {
		t.Errorf("Expected roleScopeTagIds to be [\"0\"], got %v", parsed["roleScopeTagIds"])
	}

	// Verify supportsScopeTags is boolean
	supportsScopeTags, ok := parsed["supportsScopeTags"].(bool)
	if !ok || !supportsScopeTags {
		t.Errorf("Expected supportsScopeTags to be true, got %v", parsed["supportsScopeTags"])
	}
}
