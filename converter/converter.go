package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-settings-catalog/catalog"
	"howett.net/plist"
)

// MobileconfigConverter converts mobileconfig files to Intune Settings Catalog JSON
type MobileconfigConverter struct {
	catalog *catalog.Catalog
}

// NewMobileconfigConverter creates a new converter with the given catalog
func NewMobileconfigConverter(cat *catalog.Catalog) *MobileconfigConverter {
	return &MobileconfigConverter{catalog: cat}
}

// Convert converts a mobileconfig file to Intune Settings Catalog JSON
func (c *MobileconfigConverter) Convert(data []byte, profileName string) (*ConversionResult, error) {
	// Try to parse as plist first
	var root map[string]any
	wasSignedFile := false

	_, err := plist.Unmarshal(data, &root)
	if err != nil {
		// Try stripping signature
		unsigned, stripErr := StripSignature(data)
		if stripErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrNotPlist, err)
		}
		_, err = plist.Unmarshal(unsigned, &root)
		if err != nil {
			return nil, ErrNotPlist
		}
		wasSignedFile = true
	}

	// Extract payload content
	payloadContent, ok := root["PayloadContent"].([]any)
	if !ok {
		return nil, ErrNoPayloadContent
	}

	name, _ := root["PayloadDisplayName"].(string)
	if name == "" {
		name = profileName
	}
	description, _ := root["PayloadDescription"].(string)

	var settings []map[string]any
	var skippedPayloads []string
	var skippedKeys []SkippedKey
	var matchedKeys []MatchedKey

	// Process each payload
	for _, p := range payloadContent {
		payload, ok := p.(map[string]any)
		if !ok {
			continue
		}

		payloadType, _ := payload["PayloadType"].(string)
		if payloadType == "" {
			continue
		}

		// Find root groups for this payload type
		rootGroups := c.catalog.RootGroups(payloadType)
		if len(rootGroups) == 0 {
			skippedPayloads = append(skippedPayloads, payloadType)
			continue
		}

		// Filter out Payload* metadata keys
		dataKeys := make(map[string]any)
		for k, v := range payload {
			if len(k) < 7 || k[:7] != "Payload" {
				dataKeys[k] = v
			}
		}

		// Match keys to catalog entries
		groupedChildren := make(map[string][]map[string]any)

		for key, value := range dataKeys {
			matched := false
			var matchResult catalog.MatchResult
			
			for _, rootGroup := range rootGroups {
				matchResult = c.catalog.FindChildWithDetails(key, rootGroup.ChildIDs)
				if matchResult.Entry != nil {
					settingInstance := c.buildSettingInstance(matchResult.Entry, value)
					if settingInstance != nil {
						groupedChildren[rootGroup.ID] = append(groupedChildren[rootGroup.ID], settingInstance)
						matched = true
						
						// Track matched key
						matchedKeys = append(matchedKeys, MatchedKey{
							Path:            fmt.Sprintf("%s.%s", payloadType, key),
							Value:           fmt.Sprintf("%v", value),
							CatalogID:       matchResult.Entry.ID,
							CatalogName:     matchResult.Entry.DisplayName,
							MatchType:       MatchType(matchResult.MatchType),
							SimilarityScore: matchResult.SimilarityScore,
						})
						break
					}
				}
			}
			
			if !matched {
				// Convert catalog.NearestMatch to converter.NearestMatch
				var nearMatches []NearestMatch
				for _, nm := range matchResult.NearestMatches {
					nearMatches = append(nearMatches, NearestMatch{
						CatalogID:       nm.Entry.ID,
						CatalogName:     nm.Entry.DisplayName,
						SimilarityScore: nm.SimilarityScore,
					})
				}
				
				skippedKeys = append(skippedKeys, SkippedKey{
					Path:           fmt.Sprintf("%s.%s", payloadType, key),
					Value:          fmt.Sprintf("%v", value),
					NearestMatches: nearMatches,
				})
			}
		}

		// Build settings entries
		for rootGroupID, children := range groupedChildren {
			rootGroup := c.catalog.Entry(rootGroupID)
			if rootGroup == nil {
				continue
			}

			setting := map[string]any{
				"settingInstance": map[string]any{
					"@odata.type":                      ODataTypeGroupInstance,
					"settingDefinitionId":              rootGroupID,
					"settingInstanceTemplateReference": nil,
					"groupSettingCollectionValue": []map[string]any{
						{
							"settingValueTemplateReference": nil,
							"children":                      children,
						},
					},
				},
			}
			settings = append(settings, setting)
		}
	}

	// Build final output (even if settings is empty, we'll return the result for fallback handling)
	output := map[string]any{
		"name":         name,
		"description":  description,
		"platforms":    "macOS",
		"technologies": "mdm",
		"settings":     settings,
	}

	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Determine if custom config will be used
	usedCustomConfig := len(settings) == 0 || len(skippedKeys) > 0 || len(skippedPayloads) > 0
	
	// Return result even if no settings matched - exporters will handle fallback
	return &ConversionResult{
		OutputJSON:         outputJSON,
		SettingCount:       len(settings),
		SkippedPayloads:    skippedPayloads,
		SkippedKeys:        skippedKeys,
		MatchedKeys:        matchedKeys,
		WasSignedFile:      wasSignedFile,
		OriginalData:       data,
		ProfileName:        name,
		ProfileDescription: description,
		UsedCustomConfig:   usedCustomConfig,
	}, nil
}

func (c *MobileconfigConverter) buildSettingInstance(entry *catalog.CatalogEntry, value any) map[string]any {
	// Check if this is a collection type based on catalog OData type
	isSimpleCollection := strings.Contains(entry.ODataType, "SimpleSettingCollectionDefinition")
	isGroupCollection := strings.Contains(entry.ODataType, "SettingGroupCollectionDefinition")
	isChoiceCollection := strings.Contains(entry.ODataType, "ChoiceSettingCollectionDefinition")
	isCollection := isSimpleCollection || isGroupCollection || isChoiceCollection

	// Check if value is an array
	valueArray, valueIsArray := value.([]any)

	// If catalog says it's a collection but value isn't an array, skip
	if isCollection && !valueIsArray {
		return nil
	}

	// If value is an array but catalog doesn't say collection, this is a mismatch - skip
	if valueIsArray && !isCollection {
		return nil
	}

	// Handle collection types
	if isSimpleCollection && valueIsArray {
		return c.buildSimpleCollectionInstance(entry, valueArray)
	}

	if isGroupCollection && valueIsArray {
		return c.buildGroupCollectionInstance(entry, valueArray)
	}

	if isChoiceCollection && valueIsArray {
		return c.buildChoiceCollectionInstance(entry, valueArray)
	}

	// Determine if this is a choice or simple setting based on catalog entry
	isChoice := len(entry.Options) > 0

	if isChoice {
		// Handle choice settings (boolean or enum)
		return c.buildChoiceInstance(entry, value)
	}

	// Handle simple settings (string, integer)
	return c.buildSimpleInstance(entry, value)
}

func (c *MobileconfigConverter) buildChoiceInstance(entry *catalog.CatalogEntry, value any) map[string]any {
	// Convert value to string for matching
	valueStr := fmt.Sprintf("%v", value)
	if b, ok := value.(bool); ok {
		if b {
			valueStr = "true"
		} else {
			valueStr = "false"
		}
	}

	// Find matching option
	var matchedItemID string
	for _, opt := range entry.Options {
		if strings.EqualFold(opt.OptionValue, valueStr) {
			matchedItemID = opt.ItemID
			break
		}
	}

	if matchedItemID == "" {
		return nil
	}

	return map[string]any{
		"@odata.type":                      ODataTypeChoiceInstance,
		"settingDefinitionId":              entry.ID,
		"settingInstanceTemplateReference": nil,
		"choiceSettingValue": map[string]any{
			"settingValueTemplateReference": nil,
			"value":                         matchedItemID,
			"children":                      []any{},
		},
	}
}

func (c *MobileconfigConverter) buildSimpleCollectionInstance(entry *catalog.CatalogEntry, values []any) map[string]any {
	// Build array of simple setting values
	var settingValues []map[string]any

	for _, val := range values {
		// Determine the OData type for each value
		// For simple collections, values are typically strings
		settingValue := map[string]any{
			"@odata.type":                   ODataTypeStringValue,
			"settingValueTemplateReference": nil,
			"value":                         fmt.Sprintf("%v", val),
		}
		settingValues = append(settingValues, settingValue)
	}

	return map[string]any{
		"@odata.type":                      ODataTypeSimpleCollectionInstance,
		"settingDefinitionId":              entry.ID,
		"settingInstanceTemplateReference": nil,
		"simpleSettingCollectionValue":     settingValues,
	}
}

func (c *MobileconfigConverter) buildChoiceCollectionInstance(entry *catalog.CatalogEntry, values []any) map[string]any {
	// Build array of choice setting values
	var choiceValues []map[string]any

	for _, val := range values {
		// Convert value to string for matching
		valueStr := fmt.Sprintf("%v", val)
		if b, ok := val.(bool); ok {
			if b {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		}

		// Find matching option
		var matchedItemID string
		for _, opt := range entry.Options {
			if strings.EqualFold(opt.OptionValue, valueStr) {
				matchedItemID = opt.ItemID
				break
			}
		}

		if matchedItemID == "" {
			continue
		}

		choiceValue := map[string]any{
			"settingValueTemplateReference": nil,
			"value":                         matchedItemID,
			"children":                      []any{},
		}
		choiceValues = append(choiceValues, choiceValue)
	}

	return map[string]any{
		"@odata.type":                      ODataTypeChoiceCollectionInstance,
		"settingDefinitionId":              entry.ID,
		"settingInstanceTemplateReference": nil,
		"choiceSettingCollectionValue":     choiceValues,
	}
}

func (c *MobileconfigConverter) buildGroupCollectionInstance(entry *catalog.CatalogEntry, values []any) map[string]any {
	// Build array of group setting values
	var groupValues []map[string]any

	for _, val := range values {
		// Each value should be a map representing a group with children
		valMap, ok := val.(map[string]any)
		if !ok {
			continue
		}

		// Build children for this group item
		var children []map[string]any

		// Process each key in the group as a potential child setting
		for key, childValue := range valMap {
			// Skip Payload* keys
			if len(key) >= 7 && key[:7] == "Payload" {
				continue
			}

			// Find the child entry in the catalog
			if childEntry := c.catalog.FindChild(key, entry.ChildIDs); childEntry != nil {
				if childInstance := c.buildSettingInstance(childEntry, childValue); childInstance != nil {
					children = append(children, childInstance)
				}
			}
		}

		groupValue := map[string]any{
			"settingValueTemplateReference": nil,
			"children":                      children,
		}
		groupValues = append(groupValues, groupValue)
	}

	return map[string]any{
		"@odata.type":                      ODataTypeGroupInstance,
		"settingDefinitionId":              entry.ID,
		"settingInstanceTemplateReference": nil,
		"groupSettingCollectionValue":      groupValues,
	}
}

func (c *MobileconfigConverter) buildSimpleInstance(entry *catalog.CatalogEntry, value any) map[string]any {
	var settingValue map[string]any

	// Check if this is a secret value
	isSecret := entry.ValueDefinition != nil && entry.ValueDefinition.IsSecret

	// Check if this should be an integer based on catalog definition
	isInteger := entry.ValueDefinition != nil &&
		strings.Contains(entry.ValueDefinition.ODataType, "IntegerSettingValueDefinition")

	// Determine value type based on catalog definition
	if isSecret {
		// Secret values need special handling
		settingValue = map[string]any{
			"@odata.type":                   ODataTypeSecretValue,
			"settingValueTemplateReference": nil,
			"value":                         fmt.Sprintf("%v", value),
			"valueState":                    "notEncrypted",
		}
	} else if isInteger {
		// Force integer type based on catalog
		var intValue int64
		switch v := value.(type) {
		case int:
			intValue = int64(v)
		case int32:
			intValue = int64(v)
		case int64:
			intValue = v
		case uint:
			intValue = int64(v)
		case uint32:
			intValue = int64(v)
		case uint64:
			intValue = int64(v)
		case float64:
			intValue = int64(v)
		case string:
			// Try to parse string as integer
			if parsed, err := fmt.Sscanf(v, "%d", &intValue); err == nil && parsed == 1 {
				// Successfully parsed
			} else {
				return nil
			}
		default:
			return nil
		}

		// Validate against catalog bounds
		if entry.ValueDefinition != nil {
			if entry.ValueDefinition.MinimumValue != nil && intValue < int64(*entry.ValueDefinition.MinimumValue) {
				return nil
			}
			if entry.ValueDefinition.MaximumValue != nil && intValue > int64(*entry.ValueDefinition.MaximumValue) {
				return nil
			}
		}

		settingValue = map[string]any{
			"@odata.type":                   ODataTypeIntegerValue,
			"settingValueTemplateReference": nil,
			"value":                         intValue,
		}
	} else {
		// String value
		switch v := value.(type) {
		case int, int32, int64:
			settingValue = map[string]any{
				"@odata.type":                   ODataTypeIntegerValue,
				"settingValueTemplateReference": nil,
				"value":                         v,
			}
		case float64:
			if v == float64(int64(v)) {
				settingValue = map[string]any{
					"@odata.type":                   ODataTypeIntegerValue,
					"settingValueTemplateReference": nil,
					"value":                         int64(v),
				}
			} else {
				settingValue = map[string]any{
					"@odata.type":                   ODataTypeStringValue,
					"settingValueTemplateReference": nil,
					"value":                         fmt.Sprintf("%v", v),
				}
			}
		default:
			settingValue = map[string]any{
				"@odata.type":                   ODataTypeStringValue,
				"settingValueTemplateReference": nil,
				"value":                         fmt.Sprintf("%v", v),
			}
		}
	}

	return map[string]any{
		"@odata.type":                      ODataTypeSimpleInstance,
		"settingDefinitionId":              entry.ID,
		"settingInstanceTemplateReference": nil,
		"simpleSettingValue":               settingValue,
	}
}
