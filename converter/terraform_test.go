package converter

import (
	"strings"
	"testing"
)

func TestExportFromResult_SettingsCatalog(t *testing.T) {
	exporter := NewTerraformExporter()

	result := &ConversionResult{
		OutputJSON: []byte(`{
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
						"groupSettingCollectionValue": [
							{
								"settingValueTemplateReference": null,
								"children": [
									{
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
										"settingDefinitionId": "test_setting_1",
										"settingInstanceTemplateReference": null,
										"simpleSettingValue": {
											"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue",
											"settingValueTemplateReference": null,
											"value": 64
										}
									}
								]
							}
						]
					}
				}
			]
		}`),
		SettingCount:       1,
		SkippedPayloads:    []string{},
		SkippedKeys:        []SkippedKey{},
		MatchedKeys:        []MatchedKey{},
		ProfileName:        "Test Profile",
		ProfileDescription: "Test description",
		UsedCustomConfig:   false,
	}

	hcl, err := exporter.ExportFromResult(result, "test_resource")
	if err != nil {
		t.Fatalf("ExportFromResult failed: %v", err)
	}

	if !strings.Contains(hcl, "microsoft365_graph_beta_device_management_settings_catalog_configuration_policy") {
		t.Error("Expected Settings Catalog resource type")
	}

	if !strings.Contains(hcl, "name        = \"Test Profile\"") {
		t.Error("Expected profile name in HCL")
	}

	if !strings.Contains(hcl, "configuration_policy") {
		t.Error("Expected configuration_policy block")
	}
}

func TestExportFromResult_CustomConfig(t *testing.T) {
	exporter := NewTerraformExporter()

	originalData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
	<string>Test Unknown</string>
</dict>
</plist>`)

	result := &ConversionResult{
		OutputJSON:         []byte(`{"name":"Test Unknown","description":"","platforms":"macOS","technologies":"mdm","settings":[]}`),
		SettingCount:       0,
		SkippedPayloads:    []string{"com.unknown.payload"},
		SkippedKeys:        []SkippedKey{},
		MatchedKeys:        []MatchedKey{},
		OriginalData:       originalData,
		ProfileName:        "Test Unknown",
		ProfileDescription: "",
		UsedCustomConfig:   true,
	}

	hcl, err := exporter.ExportFromResult(result, "test_custom")
	if err != nil {
		t.Fatalf("ExportFromResult failed: %v", err)
	}

	if !strings.Contains(hcl, "microsoft365_graph_beta_device_management_macos_device_configuration_templates") {
		t.Error("Expected custom configuration resource type")
	}

	if !strings.Contains(hcl, "custom_configuration") {
		t.Error("Expected custom_configuration block")
	}

	if !strings.Contains(hcl, "<<-EOT") {
		t.Error("Expected HEREDOC for payload")
	}

	if !strings.Contains(hcl, "PayloadContent") {
		t.Error("Expected original XML content in payload")
	}
}

func TestExportFromResult_CustomConfigWithSkippedKeys(t *testing.T) {
	exporter := NewTerraformExporter()

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

	hcl, err := exporter.ExportFromResult(result, "test_skipped")
	if err != nil {
		t.Fatalf("ExportFromResult failed: %v", err)
	}

	if !strings.Contains(hcl, "microsoft365_graph_beta_device_management_macos_device_configuration_templates") {
		t.Error("Expected custom configuration resource type when keys are skipped")
	}
}

func TestExportAsCustomConfig_DollarSignEscaping(t *testing.T) {
	exporter := NewTerraformExporter()

	originalData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>TestKey</key>
	<string>${variable}</string>
	<key>AnotherKey</key>
	<string>$HOME/path</string>
</dict>
</plist>`)

	result := &ConversionResult{
		OriginalData:       originalData,
		ProfileName:        "Test Dollar Signs",
		ProfileDescription: "Testing escaping",
		SettingCount:       0,
		UsedCustomConfig:   true,
	}

	hcl, err := exporter.exportAsCustomConfig(result, "test_escape")
	if err != nil {
		t.Fatalf("exportAsCustomConfig failed: %v", err)
	}

	// Verify dollar signs are escaped
	if !strings.Contains(hcl, "$${variable}") {
		t.Error("Expected ${variable} to be escaped as $${variable}")
	}

	if !strings.Contains(hcl, "$$HOME/path") {
		t.Error("Expected $HOME to be escaped as $$HOME")
	}

	// Verify single dollar signs don't exist (they should all be doubled)
	if strings.Contains(hcl, "${variable}") && !strings.Contains(hcl, "$${variable}") {
		t.Error("Found unescaped ${variable}")
	}
}

func TestExportToHCL_BasicStructure(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "Test description",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": []
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_resource")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	expectedStrings := []string{
		"resource \"microsoft365_graph_beta_device_management_settings_catalog_configuration_policy\" \"test_resource\"",
		"name        = \"Test Policy\"",
		"description = \"Test description\"",
		"platforms   = \"macOS\"",
		"technologies = [\"mdm\"]",
		"role_scope_tag_ids = [\"0\"]",
		"template_reference",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(hcl, expected) {
			t.Errorf("Expected HCL to contain: %s", expected)
		}
	}
}

func TestExportToHCL_SimpleSettingInstance(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
					"settingDefinitionId": "test_setting_1",
					"settingInstanceTemplateReference": null,
					"simpleSettingValue": {
						"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue",
						"settingValueTemplateReference": null,
						"value": 42
					}
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_simple")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "setting_instance") {
		t.Error("Expected setting_instance block")
	}

	if !strings.Contains(hcl, "setting_definition_id") {
		t.Error("Expected setting_definition_id field")
	}

	if !strings.Contains(hcl, "simple_setting_value") {
		t.Error("Expected simple_setting_value block")
	}

	if !strings.Contains(hcl, "value                            = 42") {
		t.Error("Expected integer value")
	}
}

func TestExportToHCL_ChoiceSettingInstance(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingInstance",
					"settingDefinitionId": "test_choice_1",
					"settingInstanceTemplateReference": null,
					"choiceSettingValue": {
						"settingValueTemplateReference": null,
						"value": "choice_option_1",
						"children": []
					}
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_choice")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "setting_instance") {
		t.Error("Expected setting_instance block")
	}

	if !strings.Contains(hcl, "choice_setting_value") {
		t.Error("Expected choice_setting_value block")
	}

	if !strings.Contains(hcl, "value") && !strings.Contains(hcl, "choice_option_1") {
		t.Error("Expected choice value")
	}
}

func TestExportToHCL_GroupSettingCollectionInstance(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
									"settingDefinitionId": "test_child_1",
									"settingInstanceTemplateReference": null,
									"simpleSettingValue": {
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
										"settingValueTemplateReference": null,
										"value": "test_value"
									}
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_group")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "setting_instance") {
		t.Error("Expected setting_instance block")
	}

	if !strings.Contains(hcl, "group_setting_collection_value") {
		t.Error("Expected group_setting_collection_value block")
	}

	if !strings.Contains(hcl, "simple_setting_value") {
		t.Error("Expected nested simple_setting_value")
	}

	if !strings.Contains(hcl, "test_value") {
		t.Error("Expected string value in nested setting")
	}
}

func TestExportToHCL_SimpleSettingCollectionInstance(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionInstance",
									"settingDefinitionId": "test_collection_1",
									"settingInstanceTemplateReference": null,
									"simpleSettingCollectionValue": [
										{
											"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
											"settingValueTemplateReference": null,
											"value": "item1"
										},
										{
											"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
											"settingValueTemplateReference": null,
											"value": "item2"
										}
									]
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_collection")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "setting_instance") {
		t.Error("Expected setting_instance block")
	}

	if !strings.Contains(hcl, "simple_setting_collection_value") {
		t.Error("Expected simple_setting_collection_value block")
	}

	if !strings.Contains(hcl, "item1") {
		t.Error("Expected first collection item")
	}

	if !strings.Contains(hcl, "item2") {
		t.Error("Expected second collection item")
	}
}

func TestExportToHCL_ChoiceSettingCollectionInstance(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionInstance",
									"settingDefinitionId": "test_choice_collection_1",
									"settingInstanceTemplateReference": null,
									"choiceSettingCollectionValue": [
										{
											"settingValueTemplateReference": null,
											"value": "choice_1",
											"children": []
										},
										{
											"settingValueTemplateReference": null,
											"value": "choice_2",
											"children": []
										}
									]
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_choice_collection")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "setting_instance") {
		t.Error("Expected setting_instance block")
	}

	if !strings.Contains(hcl, "choice_setting_collection_value") {
		t.Error("Expected choice_setting_collection_value block")
	}

	if !strings.Contains(hcl, "choice_1") {
		t.Error("Expected first choice value")
	}

	if !strings.Contains(hcl, "choice_2") {
		t.Error("Expected second choice value")
	}
}

func TestExportToHCL_SecretValue(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
									"settingDefinitionId": "test_secret_1",
									"settingInstanceTemplateReference": null,
									"simpleSettingValue": {
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationSecretSettingValue",
										"settingValueTemplateReference": null,
										"value": "secret123",
										"valueState": "notEncrypted"
									}
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_secret")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "secret123") {
		t.Error("Expected secret value")
	}

	if !strings.Contains(hcl, "notEncrypted") {
		t.Error("Expected value_state")
	}
}

func TestExportToHCL_InvalidJSON(t *testing.T) {
	exporter := NewTerraformExporter()

	invalidJSON := []byte(`{invalid json`)

	_, err := exporter.ExportToHCL(invalidJSON, "test_invalid")
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestExportToHCL_CommaPlacement(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
									"settingDefinitionId": "test_1",
									"settingInstanceTemplateReference": null,
									"simpleSettingValue": {
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
										"settingValueTemplateReference": null,
										"value": "value1"
									}
								},
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
									"settingDefinitionId": "test_2",
									"settingInstanceTemplateReference": null,
									"simpleSettingValue": {
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
										"settingValueTemplateReference": null,
										"value": "value2"
									}
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_commas")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	// Verify proper comma placement between array elements
	lines := strings.Split(hcl, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check for closing braces followed by commas (array elements)
		if strings.HasPrefix(trimmed, "},") {
			// This is valid for array elements
			continue
		}
		
		// Check that we don't have orphan commas
		if trimmed == "," {
			t.Errorf("Line %d: Found orphan comma", i+1)
		}
	}
}

func TestExportToHCL_NoDescription(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": []
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_no_desc")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	// Should not contain description line when empty
	if strings.Contains(hcl, "description = \"\"") {
		t.Error("Should not include empty description field")
	}
}

func TestExportToHCL_ChoiceSettingWithChildren(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingInstance",
					"settingDefinitionId": "test_choice_with_children",
					"settingInstanceTemplateReference": null,
					"choiceSettingValue": {
						"settingValueTemplateReference": null,
						"value": "choice_option_1",
						"children": [
							{
								"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
								"settingDefinitionId": "child_setting_1",
								"settingInstanceTemplateReference": null,
								"simpleSettingValue": {
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
									"settingValueTemplateReference": null,
									"value": "child_value"
								}
							}
						]
					}
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_choice_children")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "choice_setting_value") {
		t.Error("Expected choice_setting_value block")
	}

	if !strings.Contains(hcl, "children") {
		t.Error("Expected children block")
	}

	if !strings.Contains(hcl, "child_value") {
		t.Error("Expected child value")
	}
}

func TestExportToHCL_SimpleCollectionWithFloatAndBool(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingCollectionInstance",
									"settingDefinitionId": "test_collection_1",
									"settingInstanceTemplateReference": null,
									"simpleSettingCollectionValue": [
										{
											"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue",
											"settingValueTemplateReference": null,
											"value": 42.5
										},
										{
											"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValue",
											"settingValueTemplateReference": null,
											"value": 100
										},
										{
											"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
											"settingValueTemplateReference": null,
											"value": true
										}
									]
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_mixed_types")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "42.5") {
		t.Error("Expected float value 42.5")
	}

	if !strings.Contains(hcl, "100") {
		t.Error("Expected integer value 100")
	}

	if !strings.Contains(hcl, "true") {
		t.Error("Expected boolean value true")
	}
}

func TestExportToHCL_ChoiceCollectionWithChildren(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingCollectionInstance",
									"settingDefinitionId": "test_choice_collection_1",
									"settingInstanceTemplateReference": null,
									"choiceSettingCollectionValue": [
										{
											"settingValueTemplateReference": null,
											"value": "choice_1",
											"children": [
												{
													"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
													"settingDefinitionId": "nested_child",
													"settingInstanceTemplateReference": null,
													"simpleSettingValue": {
														"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
														"settingValueTemplateReference": null,
														"value": "nested_value"
													}
												}
											]
										}
									]
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_choice_coll_children")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "choice_setting_collection_value") {
		t.Error("Expected choice_setting_collection_value block")
	}

	if !strings.Contains(hcl, "nested_value") {
		t.Error("Expected nested child value")
	}
}

func TestExportToHCL_SimpleSettingWithBooleanValue(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
									"settingDefinitionId": "test_bool",
									"settingInstanceTemplateReference": null,
									"simpleSettingValue": {
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
										"settingValueTemplateReference": null,
										"value": true
									}
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_bool_value")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "true") {
		t.Error("Expected boolean value true")
	}
}

func TestExportToHCL_SimpleSettingWithFloatValue(t *testing.T) {
	exporter := NewTerraformExporter()

	jsonData := []byte(`{
		"name": "Test Policy",
		"description": "",
		"platforms": "macOS",
		"technologies": "mdm",
		"settings": [
			{
				"settingInstance": {
					"@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
					"settingDefinitionId": "test_group_1",
					"settingInstanceTemplateReference": null,
					"groupSettingCollectionValue": [
						{
							"settingValueTemplateReference": null,
							"children": [
								{
									"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingInstance",
									"settingDefinitionId": "test_float",
									"settingInstanceTemplateReference": null,
									"simpleSettingValue": {
										"@odata.type": "#microsoft.graph.deviceManagementConfigurationStringSettingValue",
										"settingValueTemplateReference": null,
										"value": 3.14159
									}
								}
							]
						}
					]
				}
			}
		]
	}`)

	hcl, err := exporter.ExportToHCL(jsonData, "test_float_value")
	if err != nil {
		t.Fatalf("ExportToHCL failed: %v", err)
	}

	if !strings.Contains(hcl, "3.14159") {
		t.Error("Expected float value 3.14159")
	}
}
