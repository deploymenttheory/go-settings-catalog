package converter

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-settings-catalog/catalog"
)

func setupTestCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()

	catalogJSON := []byte(`[
		{
			"id": "device_vendor_msft_policy_config_macos_dock_largesize",
			"offsetUri": "largesize",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Large Size",
			"description": "Size of magnified icons",
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition",
				"minimumValue": 16,
				"maximumValue": 128
			}
		},
		{
			"id": "device_vendor_msft_policy_config_macos_dock_tilesize",
			"offsetUri": "tilesize",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Tile Size",
			"description": "Size of dock icons",
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition",
				"minimumValue": 16,
				"maximumValue": 128
			}
		},
		{
			"id": "device_vendor_msft_policy_config_macos_dock_orientation",
			"offsetUri": "orientation",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition",
			"childIds": [],
			"displayName": "Orientation",
			"description": "Dock orientation",
			"options": [
				{
					"itemId": "device_vendor_msft_policy_config_macos_dock_orientation_bottom",
					"name": "Bottom",
					"optionValue": "bottom"
				},
				{
					"itemId": "device_vendor_msft_policy_config_macos_dock_orientation_left",
					"name": "Left",
					"optionValue": "left"
				},
				{
					"itemId": "device_vendor_msft_policy_config_macos_dock_orientation_right",
					"name": "Right",
					"optionValue": "right"
				}
			]
		},
		{
			"id": "device_vendor_msft_policy_config_macos_dock_autohide",
			"offsetUri": "autohide",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition",
			"childIds": [],
			"displayName": "Auto Hide",
			"description": "Auto hide dock",
			"options": [
				{
					"itemId": "device_vendor_msft_policy_config_macos_dock_autohide_true",
					"name": "True",
					"optionValue": "true"
				},
				{
					"itemId": "device_vendor_msft_policy_config_macos_dock_autohide_false",
					"name": "False",
					"optionValue": "false"
				}
			]
		},
		{
			"id": "device_vendor_msft_policy_config_macos_dock",
			"offsetUri": "com.apple.dock",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": [
				"device_vendor_msft_policy_config_macos_dock_largesize",
				"device_vendor_msft_policy_config_macos_dock_tilesize",
				"device_vendor_msft_policy_config_macos_dock_orientation",
				"device_vendor_msft_policy_config_macos_dock_autohide"
			],
			"displayName": "Dock Settings"
		},
		{
			"id": "device_vendor_msft_policy_config_macos_loginitems_allowedteamidentifiers",
			"offsetUri": "AllowedTeamIdentifiers",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionDefinition",
			"childIds": [],
			"displayName": "Allowed Team Identifiers",
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValueDefinition"
			}
		},
		{
			"id": "device_vendor_msft_policy_config_macos_loginitems",
			"offsetUri": "com.apple.loginitems.managed",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": [
				"device_vendor_msft_policy_config_macos_loginitems_allowedteamidentifiers"
			],
			"displayName": "Login Items"
		},
		{
			"id": "device_vendor_msft_policy_config_macos_password_maxpinageindays",
			"offsetUri": "maxPINAgeInDays",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Max PIN Age In Days",
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition",
				"minimumValue": 1,
				"maximumValue": 730
			}
		},
		{
			"id": "device_vendor_msft_policy_config_macos_password",
			"offsetUri": "com.apple.mobiledevice.passwordpolicy",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": [
				"device_vendor_msft_policy_config_macos_password_maxpinageindays"
			],
			"displayName": "Password Policy"
		}
	]`)

	cat := catalog.NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("Failed to load test catalog: %v", err)
	}

	return cat
}

func TestConvert_SimpleSettings(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>largesize</key>
			<integer>64</integer>
			<key>tilesize</key>
			<integer>48</integer>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Dock Profile</string>
	<key>PayloadDescription</key>
	<string>Test description</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.SettingCount != 1 {
		t.Errorf("Expected 1 setting, got %d", result.SettingCount)
	}

	if result.ProfileName != "Test Dock Profile" {
		t.Errorf("Expected profile name 'Test Dock Profile', got '%s'", result.ProfileName)
	}

	if result.ProfileDescription != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", result.ProfileDescription)
	}

	if len(result.MatchedKeys) != 2 {
		t.Errorf("Expected 2 matched keys, got %d", len(result.MatchedKeys))
	}

	// Verify matched keys are exact matches
	for _, mk := range result.MatchedKeys {
		if mk.MatchType != MatchTypeExact {
			t.Errorf("Expected exact match for %s, got %s", mk.Path, mk.MatchType)
		}
		if mk.SimilarityScore != 1.0 {
			t.Errorf("Expected similarity score 1.0 for exact match, got %.2f", mk.SimilarityScore)
		}
	}
}

func TestConvert_ChoiceSettings(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>orientation</key>
			<string>left</string>
			<key>autohide</key>
			<true/>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Choice Profile</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.SettingCount != 1 {
		t.Errorf("Expected 1 setting, got %d", result.SettingCount)
	}

	// Parse the output JSON to verify choice values
	var output map[string]any
	err = json.Unmarshal(result.OutputJSON, &output)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	settings := output["settings"].([]any)
	if len(settings) != 1 {
		t.Fatalf("Expected 1 setting in output, got %d", len(settings))
	}

	settingMap := settings[0].(map[string]any)
	instance := settingMap["settingInstance"].(map[string]any)
	groupValues := instance["groupSettingCollectionValue"].([]any)
	firstGroup := groupValues[0].(map[string]any)
	children := firstGroup["children"].([]any)

	if len(children) != 2 {
		t.Errorf("Expected 2 children (orientation, autohide), got %d", len(children))
	}
}

func TestConvert_SimpleCollection(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.loginitems.managed</string>
			<key>AllowedTeamIdentifiers</key>
			<array>
				<string>TEAM123</string>
				<string>TEAM456</string>
			</array>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Collection Profile</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.SettingCount != 1 {
		t.Errorf("Expected 1 setting, got %d", result.SettingCount)
	}

	// Parse output to verify collection structure
	var output map[string]any
	err = json.Unmarshal(result.OutputJSON, &output)
	if err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	settings := output["settings"].([]any)
	settingMap := settings[0].(map[string]any)
	instance := settingMap["settingInstance"].(map[string]any)
	groupValues := instance["groupSettingCollectionValue"].([]any)
	firstGroup := groupValues[0].(map[string]any)
	children := firstGroup["children"].([]any)

	if len(children) != 1 {
		t.Fatalf("Expected 1 child (collection), got %d", len(children))
	}

	collectionChild := children[0].(map[string]any)
	if collectionChild["@odata.type"] != ODataTypeSimpleCollectionInstance {
		t.Errorf("Expected simple collection instance, got %s", collectionChild["@odata.type"])
	}

	collectionValues := collectionChild["simpleSettingCollectionValue"].([]any)
	if len(collectionValues) != 2 {
		t.Errorf("Expected 2 collection values, got %d", len(collectionValues))
	}
}

func TestConvert_IntegerBoundsValidation(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	tests := []struct {
		name          string
		value         int
		shouldSkip    bool
		description   string
	}{
		{"valid_value", 365, false, "Value within bounds"},
		{"below_minimum", 0, true, "Value below minimum (1)"},
		{"above_maximum", 1000, true, "Value above maximum (730)"},
		{"at_minimum", 1, false, "Value at minimum boundary"},
		{"at_maximum", 730, false, "Value at maximum boundary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.mobiledevice.passwordpolicy</string>
			<key>maxPINAgeInDays</key>
			<integer>` + string(rune(tt.value+'0')) + `</integer>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Password Policy</string>
</dict>
</plist>`)

			// Manually construct the plist with the correct value
			mobileconfigData = []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.mobiledevice.passwordpolicy</string>
			<key>maxPINAgeInDays</key>
			<integer>%d</integer>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Password Policy</string>
</dict>
</plist>`, tt.value))

			result, err := converter.Convert(mobileconfigData, "test-profile")
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			if tt.shouldSkip {
				if len(result.SkippedKeys) == 0 {
					t.Errorf("%s: Expected key to be skipped due to bounds, but it wasn't", tt.description)
				}
			} else {
				if len(result.MatchedKeys) == 0 {
					t.Errorf("%s: Expected key to be matched, but it wasn't", tt.description)
				}
			}
		})
	}
}

func TestConvert_UnsignedTypes(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	// Test that uint64 values are correctly handled
	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>largesize</key>
			<integer>64</integer>
			<key>tilesize</key>
			<integer>48</integer>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Unsigned Types</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(result.MatchedKeys) != 2 {
		t.Errorf("Expected 2 matched keys (largesize, tilesize), got %d", len(result.MatchedKeys))
	}

	if result.SettingCount != 1 {
		t.Errorf("Expected 1 setting, got %d", result.SettingCount)
	}
}

func TestConvert_SkippedPayloads(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.unknown.payload</string>
			<key>SomeSetting</key>
			<string>value</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Unknown Payload</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(result.SkippedPayloads) != 1 {
		t.Errorf("Expected 1 skipped payload, got %d", len(result.SkippedPayloads))
	}

	if result.SkippedPayloads[0] != "com.unknown.payload" {
		t.Errorf("Expected skipped payload 'com.unknown.payload', got '%s'", result.SkippedPayloads[0])
	}

	if !result.UsedCustomConfig {
		t.Error("Expected UsedCustomConfig to be true when payloads are skipped")
	}
}

func TestConvert_SkippedKeys(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>largesize</key>
			<integer>64</integer>
			<key>unknownkey</key>
			<string>value</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Skipped Keys</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(result.SkippedKeys) != 1 {
		t.Errorf("Expected 1 skipped key, got %d", len(result.SkippedKeys))
	}

	if result.SkippedKeys[0].Path != "com.apple.dock.unknownkey" {
		t.Errorf("Expected skipped key path 'com.apple.dock.unknownkey', got '%s'", result.SkippedKeys[0].Path)
	}

	if !result.UsedCustomConfig {
		t.Error("Expected UsedCustomConfig to be true when keys are skipped")
	}

	// Should have nearest matches populated
	if len(result.SkippedKeys[0].NearestMatches) == 0 {
		t.Error("Expected nearest matches to be populated for skipped key")
	}
}

func TestConvert_FuzzyMatching(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "device_vendor_msft_policy_config_macos_dock_showindicators",
			"offsetUri": "show-indicators-immutable",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Show Indicators Immutable",
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition"
			}
		},
		{
			"id": "device_vendor_msft_policy_config_macos_dock",
			"offsetUri": "com.apple.dock",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": ["device_vendor_msft_policy_config_macos_dock_showindicators"],
			"displayName": "Dock Settings"
		}
	]`)

	cat := catalog.NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("Failed to load test catalog: %v", err)
	}

	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>show-process-indicators-immutable</key>
			<integer>1</integer>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Fuzzy Match</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(result.MatchedKeys) != 1 {
		t.Errorf("Expected 1 matched key via fuzzy match, got %d", len(result.MatchedKeys))
	}

	if len(result.MatchedKeys) > 0 {
		if result.MatchedKeys[0].MatchType != MatchTypeFuzzy {
			t.Errorf("Expected fuzzy match type, got %s", result.MatchedKeys[0].MatchType)
		}
		if result.MatchedKeys[0].SimilarityScore < 0.75 {
			t.Errorf("Expected similarity score >= 0.75, got %.2f", result.MatchedKeys[0].SimilarityScore)
		}
	}
}

func TestConvert_NotPlist(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	invalidData := []byte(`This is not a plist file`)

	_, err := converter.Convert(invalidData, "test-profile")
	if err == nil {
		t.Fatal("Expected error for invalid plist data")
	}

	if !strings.Contains(err.Error(), "not a valid property list") {
		t.Errorf("Expected plist error, got %v", err)
	}
}

func TestConvert_NoPayloadContent(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadDisplayName</key>
	<string>Test No Content</string>
</dict>
</plist>`)

	_, err := converter.Convert(mobileconfigData, "test-profile")
	if err == nil {
		t.Fatal("Expected error for missing PayloadContent")
	}

	if err != ErrNoPayloadContent {
		t.Errorf("Expected ErrNoPayloadContent, got %v", err)
	}
}

func TestConvert_EmptyResult(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
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
	<string>Test Empty Result</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert should not error on unmatched payloads: %v", err)
	}

	if result.SettingCount != 0 {
		t.Errorf("Expected 0 settings, got %d", result.SettingCount)
	}

	if !result.UsedCustomConfig {
		t.Error("Expected UsedCustomConfig to be true for empty result")
	}

	if len(result.OriginalData) == 0 {
		t.Error("Expected OriginalData to be preserved")
	}
}

func TestBuildChoiceInstance(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_choice_1",
		OffsetURI: "TestChoice",
		ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition",
		Options: []catalog.CatalogOption{
			{ItemID: "test_choice_1_left", DisplayName: "Left", OptionValue: "left"},
			{ItemID: "test_choice_1_bottom", DisplayName: "Bottom", OptionValue: "bottom"},
			{ItemID: "test_choice_1_right", DisplayName: "Right", OptionValue: "right"},
		},
	}

	tests := []struct {
		name     string
		value    any
		expected string
		isNil    bool
	}{
		{"string_match", "left", "test_choice_1_left", false},
		{"case_insensitive", "BOTTOM", "test_choice_1_bottom", false},
		{"no_match", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.buildChoiceInstance(entry, tt.value)
			if tt.isNil {
				if result != nil {
					t.Errorf("Expected nil result for invalid choice, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected non-nil result")
				}
				choiceValue := result["choiceSettingValue"].(map[string]any)
				if choiceValue["value"] != tt.expected {
					t.Errorf("Expected value %s, got %s", tt.expected, choiceValue["value"])
				}
			}
		})
	}
}

func TestBuildChoiceInstance_Boolean(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_autohide",
		OffsetURI: "autohide",
		ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition",
		Options: []catalog.CatalogOption{
			{ItemID: "test_autohide_true", DisplayName: "True", OptionValue: "true"},
			{ItemID: "test_autohide_false", DisplayName: "False", OptionValue: "false"},
		},
	}

	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true_value", true, "test_autohide_true"},
		{"false_value", false, "test_autohide_false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.buildChoiceInstance(entry, tt.value)
			if result == nil {
				t.Fatal("Expected non-nil result")
			}
			choiceValue := result["choiceSettingValue"].(map[string]any)
			if choiceValue["value"] != tt.expected {
				t.Errorf("Expected value %s, got %s", tt.expected, choiceValue["value"])
			}
		})
	}
}

func TestBuildSimpleInstance_IntegerTypes(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	minVal := 16
	maxVal := 128
	entry := &catalog.CatalogEntry{
		ID:        "test_integer_1",
		OffsetURI: "TestInteger",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
		ValueDefinition: &catalog.ValueDefinition{
			ODataType:    "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition",
			MinimumValue: &minVal,
			MaximumValue: &maxVal,
		},
	}

	tests := []struct {
		name  string
		value any
		isNil bool
	}{
		{"int_value", int(64), false},
		{"int32_value", int32(64), false},
		{"int64_value", int64(64), false},
		{"uint_value", uint(64), false},
		{"uint32_value", uint32(64), false},
		{"uint64_value", uint64(64), false},
		{"float64_value", float64(64), false},
		{"string_value", "64", false},
		{"invalid_string", "not_a_number", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.buildSimpleInstance(entry, tt.value)
			if tt.isNil {
				if result != nil {
					t.Errorf("Expected nil result for invalid value, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected non-nil result")
				}
				simpleValue := result["simpleSettingValue"].(map[string]any)
				if simpleValue["@odata.type"] != ODataTypeIntegerValue {
					t.Errorf("Expected integer value type, got %s", simpleValue["@odata.type"])
				}
			}
		})
	}
}

func TestBuildSimpleCollectionInstance(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_collection_1",
		OffsetURI: "TestCollection",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionDefinition",
		ValueDefinition: &catalog.ValueDefinition{
			ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValueDefinition",
		},
	}

	values := []any{"TEAM123", "TEAM456", "TEAM789"}
	result := converter.buildSimpleCollectionInstance(entry, values)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result["@odata.type"] != ODataTypeSimpleCollectionInstance {
		t.Errorf("Expected simple collection instance type, got %s", result["@odata.type"])
	}

	collectionValues := result["simpleSettingCollectionValue"].([]map[string]any)
	if len(collectionValues) != 3 {
		t.Errorf("Expected 3 collection values, got %d", len(collectionValues))
	}

	for i, cv := range collectionValues {
		if cv["@odata.type"] != ODataTypeStringValue {
			t.Errorf("Value %d: Expected string value type, got %s", i, cv["@odata.type"])
		}
	}
}

func TestConvert_ProfileNameFallback(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>largesize</key>
			<integer>64</integer>
		</dict>
	</array>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "fallback-name")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.ProfileName != "fallback-name" {
		t.Errorf("Expected profile name 'fallback-name', got '%s'", result.ProfileName)
	}
}

func TestBuildGroupCollectionInstance(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	entry := &catalog.CatalogEntry{
		ID:        "test_group_collection",
		OffsetURI: "TestGroupCollection",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
		ChildIDs:  []string{"child_1"},
	}

	// Mock the catalog to return the child entry
	catalogJSON := []byte(`[
		{
			"id": "child_1",
			"offsetUri": "ChildSetting",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValueDefinition"
			}
		}
	]`)
	
	testCat := catalog.NewCatalog()
	_ = testCat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	converter.catalog = testCat

	values := []any{
		map[string]any{
			"ChildSetting": "value1",
		},
		map[string]any{
			"ChildSetting": "value2",
		},
	}

	result := converter.buildGroupCollectionInstance(entry, values)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result["@odata.type"] != ODataTypeGroupInstance {
		t.Errorf("Expected group instance type, got %s", result["@odata.type"])
	}

	groupValues := result["groupSettingCollectionValue"].([]map[string]any)
	if len(groupValues) != 2 {
		t.Errorf("Expected 2 group values, got %d", len(groupValues))
	}
}

func TestBuildGroupCollectionInstance_SkipsInvalidValues(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_group",
		OffsetURI: "TestGroup",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
		ChildIDs:  []string{},
	}

	// Pass non-map values
	values := []any{
		"not a map",
		123,
		true,
	}

	result := converter.buildGroupCollectionInstance(entry, values)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	groupValues := result["groupSettingCollectionValue"].([]map[string]any)
	if len(groupValues) != 0 {
		t.Errorf("Expected 0 group values (all skipped), got %d", len(groupValues))
	}
}

func TestBuildChoiceCollectionInstance(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_choice_collection",
		OffsetURI: "TestChoiceCollection",
		ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionDefinition",
		Options: []catalog.CatalogOption{
			{ItemID: "option_1", DisplayName: "Option 1", OptionValue: "value1"},
			{ItemID: "option_2", DisplayName: "Option 2", OptionValue: "value2"},
		},
	}

	values := []any{"value1", "value2"}

	result := converter.buildChoiceCollectionInstance(entry, values)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result["@odata.type"] != ODataTypeChoiceCollectionInstance {
		t.Errorf("Expected choice collection instance type, got %s", result["@odata.type"])
	}

	choiceValues := result["choiceSettingCollectionValue"].([]map[string]any)
	if len(choiceValues) != 2 {
		t.Errorf("Expected 2 choice values, got %d", len(choiceValues))
	}
}

func TestBuildChoiceCollectionInstance_SkipsUnmatchedValues(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_choice_collection",
		OffsetURI: "TestChoiceCollection",
		ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionDefinition",
		Options: []catalog.CatalogOption{
			{ItemID: "option_1", DisplayName: "Option 1", OptionValue: "value1"},
		},
	}

	values := []any{"value1", "unmatched_value", "another_unmatched"}

	result := converter.buildChoiceCollectionInstance(entry, values)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	choiceValues := result["choiceSettingCollectionValue"].([]map[string]any)
	if len(choiceValues) != 1 {
		t.Errorf("Expected 1 choice value (2 skipped), got %d", len(choiceValues))
	}
}

func TestBuildChoiceCollectionInstance_BooleanValues(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_bool_collection",
		OffsetURI: "TestBoolCollection",
		ODataType: "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionDefinition",
		Options: []catalog.CatalogOption{
			{ItemID: "option_true", DisplayName: "True", OptionValue: "true"},
			{ItemID: "option_false", DisplayName: "False", OptionValue: "false"},
		},
	}

	values := []any{true, false, true}

	result := converter.buildChoiceCollectionInstance(entry, values)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	choiceValues := result["choiceSettingCollectionValue"].([]map[string]any)
	if len(choiceValues) != 3 {
		t.Errorf("Expected 3 choice values, got %d", len(choiceValues))
	}
}

func TestBuildSimpleInstance_StringValue(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_string",
		OffsetURI: "TestString",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
		ValueDefinition: &catalog.ValueDefinition{
			ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValueDefinition",
		},
	}

	result := converter.buildSimpleInstance(entry, "test_value")

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	simpleValue := result["simpleSettingValue"].(map[string]any)
	if simpleValue["@odata.type"] != ODataTypeStringValue {
		t.Errorf("Expected string value type, got %s", simpleValue["@odata.type"])
	}
	if simpleValue["value"] != "test_value" {
		t.Errorf("Expected value 'test_value', got %v", simpleValue["value"])
	}
}

func TestBuildSimpleInstance_SecretValue(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_secret",
		OffsetURI: "TestSecret",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
		ValueDefinition: &catalog.ValueDefinition{
			ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValueDefinition",
			IsSecret:  true,
		},
	}

	result := converter.buildSimpleInstance(entry, "secret_password")

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	simpleValue := result["simpleSettingValue"].(map[string]any)
	if simpleValue["@odata.type"] != ODataTypeSecretValue {
		t.Errorf("Expected secret value type, got %s", simpleValue["@odata.type"])
	}
	if simpleValue["valueState"] != "notEncrypted" {
		t.Errorf("Expected valueState 'notEncrypted', got %v", simpleValue["valueState"])
	}
}

func TestBuildSimpleInstance_FloatValue(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	entry := &catalog.CatalogEntry{
		ID:        "test_float",
		OffsetURI: "TestFloat",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
		ValueDefinition: &catalog.ValueDefinition{
			ODataType: "#microsoft.graph.deviceManagementConfigurationStringSettingValueDefinition",
		},
	}

	// Test non-integer float
	result := converter.buildSimpleInstance(entry, 3.14)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	simpleValue := result["simpleSettingValue"].(map[string]any)
	if simpleValue["@odata.type"] != ODataTypeStringValue {
		t.Errorf("Expected string value type for float, got %s", simpleValue["@odata.type"])
	}
}

func TestBuildSettingInstance_CollectionMismatch(t *testing.T) {
	converter := NewMobileconfigConverter(nil)

	// Entry expects a collection, but value is not an array
	collectionEntry := &catalog.CatalogEntry{
		ID:        "test_collection",
		OffsetURI: "TestCollection",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionDefinition",
	}

	result := converter.buildSettingInstance(collectionEntry, "not_an_array")
	if result != nil {
		t.Error("Expected nil result when collection entry receives non-array value")
	}

	// Entry expects a simple value, but value is an array
	simpleEntry := &catalog.CatalogEntry{
		ID:        "test_simple",
		OffsetURI: "TestSimple",
		ODataType: "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
	}

	result = converter.buildSettingInstance(simpleEntry, []any{"value1", "value2"})
	if result != nil {
		t.Error("Expected nil result when simple entry receives array value")
	}
}

func TestConvert_InvalidPayloadContent(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<string>not an array</string>
	<key>PayloadDisplayName</key>
	<string>Test Invalid</string>
</dict>
</plist>`)

	_, err := converter.Convert(mobileconfigData, "test-profile")
	if err == nil {
		t.Fatal("Expected error for invalid PayloadContent type")
	}

	if err != ErrNoPayloadContent {
		t.Errorf("Expected ErrNoPayloadContent, got %v", err)
	}
}

func TestConvert_EmptyPayloadType(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>SomeSetting</key>
			<string>value</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Empty Type</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert should not error: %v", err)
	}

	if result.SettingCount != 0 {
		t.Error("Expected 0 settings for payload without PayloadType")
	}
}

func TestConvert_InvalidPayloadStructure(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<string>not a dict</string>
		<integer>123</integer>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Invalid Payloads</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert should not error: %v", err)
	}

	if result.SettingCount != 0 {
		t.Error("Expected 0 settings for invalid payload structures")
	}
}

func TestConvert_PayloadMetadataFiltering(t *testing.T) {
	cat := setupTestCatalog(t)
	converter := NewMobileconfigConverter(cat)

	mobileconfigData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.apple.dock</string>
			<key>PayloadIdentifier</key>
			<string>com.test.identifier</string>
			<key>PayloadUUID</key>
			<string>12345</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
			<key>largesize</key>
			<integer>64</integer>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Test Metadata Filtering</string>
</dict>
</plist>`)

	result, err := converter.Convert(mobileconfigData, "test-profile")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Should only match largesize, not the Payload* metadata keys
	if len(result.MatchedKeys) != 1 {
		t.Errorf("Expected 1 matched key (Payload* keys should be filtered), got %d", len(result.MatchedKeys))
	}
	if result.MatchedKeys[0].Path != "com.apple.dock.largesize" {
		t.Errorf("Expected matched key to be largesize, got %s", result.MatchedKeys[0].Path)
	}
}

func TestNewMobileconfigConverter(t *testing.T) {
	cat := catalog.NewCatalog()
	converter := NewMobileconfigConverter(cat)

	if converter == nil {
		t.Fatal("NewMobileconfigConverter returned nil")
	}
	if converter.catalog != cat {
		t.Error("Converter catalog not set correctly")
	}
}
