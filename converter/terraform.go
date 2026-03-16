package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TerraformExporter exports Settings Catalog JSON to Terraform HCL
type TerraformExporter struct{}

// NewTerraformExporter creates a new Terraform exporter
func NewTerraformExporter() *TerraformExporter {
	return &TerraformExporter{}
}

// ExportFromResult converts a ConversionResult to Terraform HCL
// If any settings were skipped or no settings matched, falls back to custom configuration format
func (e *TerraformExporter) ExportFromResult(result *ConversionResult, resourceName string) (string, error) {
	// Check if we need to fall back to custom configuration
	if result.SettingCount == 0 || len(result.SkippedKeys) > 0 || len(result.SkippedPayloads) > 0 {
		return e.exportAsCustomConfig(result, resourceName)
	}

	// Use Settings Catalog format
	return e.ExportToHCL(result.OutputJSON, resourceName)
}

// exportAsCustomConfig generates a custom configuration template resource
func (e *TerraformExporter) exportAsCustomConfig(result *ConversionResult, resourceName string) (string, error) {
	var hcl strings.Builder

	fmt.Fprintf(&hcl, "resource \"microsoft365_graph_beta_device_management_macos_device_configuration_templates\" \"%s\" {\n", resourceName)
	fmt.Fprintf(&hcl, "  display_name = %q\n", result.ProfileName)
	
	if result.ProfileDescription != "" {
		fmt.Fprintf(&hcl, "  description  = %q\n", result.ProfileDescription)
	}

	hcl.WriteString("\n  custom_configuration = {\n")
	hcl.WriteString("    deployment_channel = \"deviceChannel\"\n")
	fmt.Fprintf(&hcl, "    payload_file_name  = %q\n", result.ProfileName+".mobileconfig")
	fmt.Fprintf(&hcl, "    payload_name       = %q\n", result.ProfileName)
	hcl.WriteString("    payload            = <<-EOT\n")
	
	// Escape $ as $$ for Terraform HEREDOC to prevent interpolation
	escapedPayload := strings.ReplaceAll(string(result.OriginalData), "$", "$$")
	hcl.WriteString(escapedPayload)
	
	hcl.WriteString("\n    EOT\n")
	hcl.WriteString("  }\n")

	hcl.WriteString("}\n")

	return hcl.String(), nil
}

// ExportToHCL converts Settings Catalog JSON to Terraform HCL blocks
func (e *TerraformExporter) ExportToHCL(jsonData []byte, resourceName string) (string, error) {
	var policy map[string]any
	if err := json.Unmarshal(jsonData, &policy); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	var hcl strings.Builder

	fmt.Fprintf(&hcl, "resource \"microsoft365_graph_beta_device_management_settings_catalog_configuration_policy\" \"%s\" {\n", resourceName)

	// Required attributes
	if name, ok := policy["name"].(string); ok {
		fmt.Fprintf(&hcl, "  name        = %q\n", name)
	}

	if desc, ok := policy["description"].(string); ok && desc != "" {
		fmt.Fprintf(&hcl, "  description = %q\n", desc)
	}

	if platforms, ok := policy["platforms"].(string); ok {
		fmt.Fprintf(&hcl, "  platforms   = %q\n", platforms)
	}

	// Technologies as array
	if tech, ok := policy["technologies"].(string); ok {
		fmt.Fprintf(&hcl, "  technologies = [%q]\n", tech)
	}

	// Default role scope tag
	hcl.WriteString("  role_scope_tag_ids = [\"0\"]\n")

	// Template reference
	hcl.WriteString("\n  template_reference = {\n")
	hcl.WriteString("    template_id = \"\"\n")
	hcl.WriteString("  }\n")

	// Configuration policy with settings
	if settings, ok := policy["settings"].([]any); ok && len(settings) > 0 {
		hcl.WriteString("\n  configuration_policy = {\n")
		hcl.WriteString("    settings = [\n")

		for i, setting := range settings {
			if s, ok := setting.(map[string]any); ok {
				hcl.WriteString("      {\n")
				fmt.Fprintf(&hcl, "        id = \"%d\"\n", i)

				if settingInstance, ok := s["settingInstance"].(map[string]any); ok {
					e.writeSettingInstance(&hcl, settingInstance, "        ")
				}

				hcl.WriteString("      }")
				if i < len(settings)-1 {
					hcl.WriteString(",")
				}
				hcl.WriteString("\n")
			}
		}

		hcl.WriteString("    ]\n")
		hcl.WriteString("  }\n")
	}

	hcl.WriteString("}\n")

	return hcl.String(), nil
}

func (e *TerraformExporter) writeSettingInstance(hcl *strings.Builder, instance map[string]any, indent string) {
	fmt.Fprintf(hcl, "%ssetting_instance = {\n", indent)

	// Write odata_type
	if odataType, ok := instance["@odata.type"].(string); ok {
		fmt.Fprintf(hcl, "%s  odata_type                          = %q\n", indent, odataType)
	}

	// Write setting_definition_id
	if settingDefID, ok := instance["settingDefinitionId"].(string); ok {
		fmt.Fprintf(hcl, "%s  setting_definition_id               = %q\n", indent, settingDefID)
	}

	// Write setting_instance_template_reference
	fmt.Fprintf(hcl, "%s  setting_instance_template_reference = null\n", indent)

	// Handle different value types
	if simpleValue, ok := instance["simpleSettingValue"].(map[string]any); ok {
		e.writeSimpleSettingValue(hcl, simpleValue, indent+"  ")
	}

	if choiceValue, ok := instance["choiceSettingValue"].(map[string]any); ok {
		e.writeChoiceSettingValue(hcl, choiceValue, indent+"  ")
	}

	if simpleCollection, ok := instance["simpleSettingCollectionValue"].([]any); ok {
		e.writeSimpleSettingCollectionValue(hcl, simpleCollection, indent+"  ")
	}

	if choiceCollection, ok := instance["choiceSettingCollectionValue"].([]any); ok {
		e.writeChoiceSettingCollectionValue(hcl, choiceCollection, indent+"  ")
	}

	if groupCollection, ok := instance["groupSettingCollectionValue"].([]any); ok {
		e.writeGroupSettingCollectionValue(hcl, groupCollection, indent+"  ")
	}

	fmt.Fprintf(hcl, "%s}\n", indent)
}

func (e *TerraformExporter) writeSimpleSettingValue(hcl *strings.Builder, value map[string]any, indent string) {
	fmt.Fprintf(hcl, "%ssimple_setting_value = {\n", indent)

	if odataType, ok := value["@odata.type"].(string); ok {
		fmt.Fprintf(hcl, "%s  odata_type                       = %q\n", indent, odataType)
	}

	fmt.Fprintf(hcl, "%s  setting_value_template_reference = null\n", indent)

	// Handle value_state for secrets
	if valueState, ok := value["valueState"].(string); ok {
		fmt.Fprintf(hcl, "%s  value_state                      = %q\n", indent, valueState)
	}

	// Write value
	if val, ok := value["value"]; ok {
		switch v := val.(type) {
		case string:
			fmt.Fprintf(hcl, "%s  value                            = %q\n", indent, v)
		case float64:
			if v == float64(int64(v)) {
				fmt.Fprintf(hcl, "%s  value                            = %d\n", indent, int64(v))
			} else {
				fmt.Fprintf(hcl, "%s  value                            = %f\n", indent, v)
			}
		case bool:
			fmt.Fprintf(hcl, "%s  value                            = %t\n", indent, v)
		}
	}

	fmt.Fprintf(hcl, "%s}\n", indent)
}

func (e *TerraformExporter) writeSimpleSettingCollectionValue(hcl *strings.Builder, collection []any, indent string) {
	fmt.Fprintf(hcl, "%ssimple_setting_collection_value = [\n", indent)

	for i, item := range collection {
		if itemMap, ok := item.(map[string]any); ok {
			fmt.Fprintf(hcl, "%s  {\n", indent)

			if odataType, ok := itemMap["@odata.type"].(string); ok {
				fmt.Fprintf(hcl, "%s    odata_type                       = %q\n", indent, odataType)
			}

			fmt.Fprintf(hcl, "%s    setting_value_template_reference = null\n", indent)

			if val, ok := itemMap["value"]; ok {
				switch v := val.(type) {
				case string:
					fmt.Fprintf(hcl, "%s    value                            = %q\n", indent, v)
				case float64:
					if v == float64(int64(v)) {
						fmt.Fprintf(hcl, "%s    value                            = %d\n", indent, int64(v))
					} else {
						fmt.Fprintf(hcl, "%s    value                            = %f\n", indent, v)
					}
				case bool:
					fmt.Fprintf(hcl, "%s    value                            = %t\n", indent, v)
				default:
					fmt.Fprintf(hcl, "%s    value                            = %q\n", indent, fmt.Sprintf("%v", v))
				}
			}

			fmt.Fprintf(hcl, "%s  }", indent)
			if i < len(collection)-1 {
				hcl.WriteString(",")
			}
			hcl.WriteString("\n")
		}
	}

	fmt.Fprintf(hcl, "%s]\n", indent)
}

func (e *TerraformExporter) writeChoiceSettingCollectionValue(hcl *strings.Builder, collection []any, indent string) {
	fmt.Fprintf(hcl, "%schoice_setting_collection_value = [\n", indent)

	for i, item := range collection {
		if itemMap, ok := item.(map[string]any); ok {
			fmt.Fprintf(hcl, "%s  {\n", indent)

			fmt.Fprintf(hcl, "%s    setting_value_template_reference = null\n", indent)

			if val, ok := itemMap["value"].(string); ok {
				fmt.Fprintf(hcl, "%s    value                            = %q\n", indent, val)
			}

			// Handle children if present
			if children, ok := itemMap["children"].([]any); ok && len(children) > 0 {
				fmt.Fprintf(hcl, "%s    children = [\n", indent)
				for _, child := range children {
					if childMap, ok := child.(map[string]any); ok {
						fmt.Fprintf(hcl, "%s      {\n", indent)
						e.writeChildInstance(hcl, childMap, indent+"        ")
						fmt.Fprintf(hcl, "%s      },\n", indent)
					}
				}
				fmt.Fprintf(hcl, "%s    ]\n", indent)
			} else {
				fmt.Fprintf(hcl, "%s    children                         = []\n", indent)
			}

			fmt.Fprintf(hcl, "%s  }", indent)
			if i < len(collection)-1 {
				hcl.WriteString(",")
			}
			hcl.WriteString("\n")
		}
	}

	fmt.Fprintf(hcl, "%s]\n", indent)
}

func (e *TerraformExporter) writeChoiceSettingValue(hcl *strings.Builder, value map[string]any, indent string) {
	fmt.Fprintf(hcl, "%schoice_setting_value = {\n", indent)

	fmt.Fprintf(hcl, "%s  setting_value_template_reference = null\n", indent)

	// Write value
	if val, ok := value["value"].(string); ok {
		fmt.Fprintf(hcl, "%s  value                            = %q\n", indent, val)
	}

	// Write children
	if children, ok := value["children"].([]any); ok && len(children) > 0 {
		fmt.Fprintf(hcl, "%s  children = [\n", indent)
		for _, child := range children {
			if childMap, ok := child.(map[string]any); ok {
				fmt.Fprintf(hcl, "%s    {\n", indent)
				e.writeChildInstance(hcl, childMap, indent+"      ")
				fmt.Fprintf(hcl, "%s    },\n", indent)
			}
		}
		fmt.Fprintf(hcl, "%s  ]\n", indent)
	} else {
		fmt.Fprintf(hcl, "%s  children                         = []\n", indent)
	}

	fmt.Fprintf(hcl, "%s}\n", indent)
}

func (e *TerraformExporter) writeGroupSettingCollectionValue(hcl *strings.Builder, collection []any, indent string) {
	fmt.Fprintf(hcl, "%sgroup_setting_collection_value = [\n", indent)

	for i, item := range collection {
		if itemMap, ok := item.(map[string]any); ok {
			fmt.Fprintf(hcl, "%s  {\n", indent)
			fmt.Fprintf(hcl, "%s    setting_value_template_reference = null\n", indent)

			if children, ok := itemMap["children"].([]any); ok {
				fmt.Fprintf(hcl, "%s    children = [\n", indent)
				for _, child := range children {
					if childMap, ok := child.(map[string]any); ok {
						fmt.Fprintf(hcl, "%s      {\n", indent)
						e.writeChildInstance(hcl, childMap, indent+"        ")
						fmt.Fprintf(hcl, "%s      },\n", indent)
					}
				}
				fmt.Fprintf(hcl, "%s    ]\n", indent)
			}

			fmt.Fprintf(hcl, "%s  }", indent)
			if i < len(collection)-1 {
				hcl.WriteString(",")
			}
			hcl.WriteString("\n")
		}
	}

	fmt.Fprintf(hcl, "%s]\n", indent)
}

func (e *TerraformExporter) writeChildInstance(hcl *strings.Builder, child map[string]any, indent string) {
	if odataType, ok := child["@odata.type"].(string); ok {
		fmt.Fprintf(hcl, "%sodata_type                          = %q\n", indent, odataType)
	}

	if settingDefID, ok := child["settingDefinitionId"].(string); ok {
		fmt.Fprintf(hcl, "%ssetting_definition_id               = %q\n", indent, settingDefID)
	}

	fmt.Fprintf(hcl, "%ssetting_instance_template_reference = null\n", indent)

	// Handle ALL value types for children (children can be any instance type)
	if simpleValue, ok := child["simpleSettingValue"].(map[string]any); ok {
		e.writeSimpleSettingValue(hcl, simpleValue, indent)
	}

	if choiceValue, ok := child["choiceSettingValue"].(map[string]any); ok {
		e.writeChoiceSettingValue(hcl, choiceValue, indent)
	}

	if simpleCollection, ok := child["simpleSettingCollectionValue"].([]any); ok {
		e.writeSimpleSettingCollectionValue(hcl, simpleCollection, indent)
	}

	if choiceCollection, ok := child["choiceSettingCollectionValue"].([]any); ok {
		e.writeChoiceSettingCollectionValue(hcl, choiceCollection, indent)
	}

	if groupCollection, ok := child["groupSettingCollectionValue"].([]any); ok {
		e.writeGroupSettingCollectionValue(hcl, groupCollection, indent)
	}
}
