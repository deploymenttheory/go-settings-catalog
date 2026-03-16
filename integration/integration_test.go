package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-settings-catalog/catalog"
	"github.com/deploymenttheory/go-settings-catalog/converter"
)

func setupIntegrationCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()

	cat := catalog.NewCatalog()
	err := cat.LoadFromFiles(
		"../cmd/mobileconfig-to-terraform/intune-settings-catalog-data/IntuneSettingsCatalog_macOS.json",
		"../cmd/mobileconfig-to-terraform/intune-settings-catalog-data/IntuneSettingsCategories_macOS.json",
		"../cmd/mobileconfig-to-terraform/intune-settings-catalog-data/IntuneSettingsVersion.json",
	)
	if err != nil {
		t.Fatalf("Failed to load catalog: %v", err)
	}

	return cat
}

func TestIntegration_SimpleSettings(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/simple_settings.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "simple_settings")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result.SettingCount == 0 {
		t.Error("Expected at least one setting to be converted")
	}

	if len(result.MatchedKeys) == 0 {
		t.Error("Expected at least one matched key")
	}

	// Test Terraform export
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, "simple_test")
	if err != nil {
		t.Fatalf("Terraform export failed: %v", err)
	}

	if !strings.Contains(hcl, "resource \"microsoft365_graph_beta_device_management") {
		t.Error("Expected valid Terraform resource")
	}

	// Test JSON export
	jsonExporter := converter.NewJSONExporter()
	jsonOutput, err := jsonExporter.ExportFromResult(result)
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}

	if len(jsonOutput) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

func TestIntegration_GroupCollection(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/group_collection.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "group_collection")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Login items should have group collections
	if result.SettingCount == 0 {
		t.Error("Expected at least one setting to be converted")
	}

	// Test Terraform export
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, "group_test")
	if err != nil {
		t.Fatalf("Terraform export failed: %v", err)
	}

	if !strings.Contains(hcl, "group_setting_collection_value") {
		t.Error("Expected group_setting_collection_value block")
	}
}

func TestIntegration_NestedGroups(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/nested_groups.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "nested_groups")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Test Terraform export
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, "nested_test")
	if err != nil {
		t.Fatalf("Terraform export failed: %v", err)
	}

	if len(hcl) == 0 {
		t.Error("Expected non-empty HCL output")
	}
}

func TestIntegration_MCXPreferences(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/mcx_preferences.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "mcx_preferences")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// MCX preferences might not all match catalog
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Test both export formats
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, "mcx_test")
	if err != nil {
		t.Fatalf("Terraform export failed: %v", err)
	}

	if len(hcl) == 0 {
		t.Error("Expected non-empty HCL output")
	}

	jsonExporter := converter.NewJSONExporter()
	jsonOutput, err := jsonExporter.ExportFromResult(result)
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}

	if len(jsonOutput) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

func TestIntegration_DollarSignPayload(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/dollar_sign_payload.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "dollar_sign_payload")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// This file likely has many settings that won't match, so it should fall back to custom config
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, "dollar_test")
	if err != nil {
		t.Fatalf("Terraform export failed: %v", err)
	}

	// If it's custom config, verify dollar signs are escaped
	if strings.Contains(hcl, "custom_configuration") {
		// Count dollar signs - they should all be doubled
		originalCount := strings.Count(string(data), "$")
		if originalCount > 0 {
			// In the HCL, they should be doubled
			hclDollarCount := strings.Count(hcl, "$$")
			if hclDollarCount < originalCount {
				t.Error("Expected all dollar signs to be escaped in HEREDOC")
			}
		}
	}
}

func TestIntegration_AllFixtures(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)
	tfExporter := converter.NewTerraformExporter()
	jsonExporter := converter.NewJSONExporter()

	fixtures, err := filepath.Glob("../testdata/fixtures/*.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to glob fixtures: %v", err)
	}

	if len(fixtures) == 0 {
		t.Skip("No test fixtures found")
	}

	for _, fixturePath := range fixtures {
		t.Run(filepath.Base(fixturePath), func(t *testing.T) {
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("Failed to read fixture: %v", err)
			}

			result, err := conv.Convert(data, filepath.Base(fixturePath))
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			// Test Terraform export
			hcl, err := tfExporter.ExportFromResult(result, "test")
			if err != nil {
				t.Errorf("Terraform export failed: %v", err)
			}
			if len(hcl) == 0 {
				t.Error("Expected non-empty HCL output")
			}

			// Test JSON export
			jsonOutput, err := jsonExporter.ExportFromResult(result)
			if err != nil {
				t.Errorf("JSON export failed: %v", err)
			}
			if len(jsonOutput) == 0 {
				t.Error("Expected non-empty JSON output")
			}
		})
	}
}

func TestIntegration_ReportGeneration(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/simple_settings.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "simple_settings")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Create a report
	report := &converter.ExportReport{
		TotalFiles:   1,
		SuccessCount: 1,
		FailureCount: 0,
		CatalogCount: 1,
		CustomCount:  0,
		FileReports: []converter.FileReport{
			{
				InputPath:       "simple_settings.mobileconfig",
				OutputPath:      "output/simple_settings.tf",
				ResourceType:    "Settings Catalog Terraform HCL",
				SettingCount:    result.SettingCount,
				MatchedKeys:     result.MatchedKeys,
				SkippedKeys:     result.SkippedKeys,
				SkippedPayloads: result.SkippedPayloads,
				WasSignedFile:   result.WasSignedFile,
				Error:           nil,
			},
		},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "# Conversion Report") {
		t.Error("Expected report title")
	}

	if !strings.Contains(md, "| Total Files | 1 |") {
		t.Error("Expected total files in report")
	}

	if len(md) < 100 {
		t.Error("Report seems too short")
	}
}

func TestIntegration_CustomConfigFallback(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	// Create a mobileconfig with unknown payload type
	unknownPayload := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadType</key>
			<string>com.completely.unknown.payload</string>
			<key>SomeSetting</key>
			<string>value</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Unknown Payload Test</string>
</dict>
</plist>`)

	result, err := conv.Convert(unknownPayload, "unknown_test")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !result.UsedCustomConfig {
		t.Error("Expected UsedCustomConfig to be true")
	}

	if len(result.SkippedPayloads) == 0 {
		t.Error("Expected skipped payloads")
	}

	// Test Terraform export falls back to custom config
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, "unknown_test")
	if err != nil {
		t.Fatalf("Terraform export failed: %v", err)
	}

	if !strings.Contains(hcl, "microsoft365_graph_beta_device_management_macos_device_configuration_templates") {
		t.Error("Expected custom configuration resource type")
	}

	if !strings.Contains(hcl, "custom_configuration") {
		t.Error("Expected custom_configuration block")
	}

	// Test JSON export falls back to custom config
	jsonExporter := converter.NewJSONExporter()
	jsonOutput, err := jsonExporter.ExportFromResult(result)
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}

	if !strings.Contains(string(jsonOutput), "#microsoft.graph.macOSCustomConfiguration") {
		t.Error("Expected custom configuration JSON type")
	}
}

func TestIntegration_FuzzyMatchingInRealProfile(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)

	data, err := os.ReadFile("../testdata/fixtures/simple_settings.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	result, err := conv.Convert(data, "fuzzy_test")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Check if any fuzzy matches were made
	hasFuzzyMatch := false
	for _, mk := range result.MatchedKeys {
		if mk.MatchType == converter.MatchTypeFuzzy {
			hasFuzzyMatch = true
			if mk.SimilarityScore < 0.75 {
				t.Errorf("Fuzzy match %s has score %.2f below threshold", mk.Path, mk.SimilarityScore)
			}
		}
	}

	// It's okay if there are no fuzzy matches - this is just informational
	if hasFuzzyMatch {
		t.Logf("Found %d fuzzy matches in real profile", countFuzzyMatches(result.MatchedKeys))
	}
}

func TestIntegration_EndToEndPipeline(t *testing.T) {
	cat := setupIntegrationCatalog(t)
	conv := converter.NewMobileconfigConverter(cat)
	tfExporter := converter.NewTerraformExporter()
	jsonExporter := converter.NewJSONExporter()

	fixtures, err := filepath.Glob("../testdata/fixtures/*.mobileconfig")
	if err != nil {
		t.Fatalf("Failed to glob fixtures: %v", err)
	}

	if len(fixtures) == 0 {
		t.Skip("No test fixtures found")
	}

	var fileReports []converter.FileReport

	for _, fixturePath := range fixtures {
		data, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", fixturePath, err)
			continue
		}

		result, err := conv.Convert(data, filepath.Base(fixturePath))
		if err != nil {
		fileReports = append(fileReports, converter.FileReport{
			InputPath: filepath.Base(fixturePath),
			Error:     err,
		})
			continue
		}

		// Export to both formats
		_, tfErr := tfExporter.ExportFromResult(result, "test")
		_, jsonErr := jsonExporter.ExportFromResult(result)

		if tfErr != nil || jsonErr != nil {
			t.Errorf("Export failed for %s: tf=%v, json=%v", fixturePath, tfErr, jsonErr)
		}

		resourceType := "Settings Catalog"
		if result.UsedCustomConfig {
			resourceType = "Custom Configuration"
		}

		fileReports = append(fileReports, converter.FileReport{
			InputPath:       filepath.Base(fixturePath),
			OutputPath:      "test.tf",
			ResourceType:    resourceType + " Terraform HCL",
			SettingCount:    result.SettingCount,
			MatchedKeys:     result.MatchedKeys,
			SkippedKeys:     result.SkippedKeys,
			SkippedPayloads: result.SkippedPayloads,
			WasSignedFile:   result.WasSignedFile,
			Error:           nil,
		})
	}

	// Generate report
	report := &converter.ExportReport{
		TotalFiles:   len(fixtures),
		SuccessCount: len(fixtures) - countErrors(fileReports),
		FailureCount: countErrors(fileReports),
		CatalogCount: countByType(fileReports, "Settings Catalog"),
		CustomCount:  countByType(fileReports, "Custom Configuration"),
		FileReports:  fileReports,
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "# Conversion Report") {
		t.Error("Expected report to be generated")
	}

	t.Logf("Processed %d fixtures: %d catalog, %d custom, %d failed",
		len(fixtures), report.CatalogCount, report.CustomCount, report.FailureCount)
}

func TestIntegration_CatalogLoading(t *testing.T) {
	cat := catalog.NewCatalog()
	err := cat.LoadFromFiles(
		"../cmd/mobileconfig-to-terraform/intune-settings-catalog-data/IntuneSettingsCatalog_macOS.json",
		"../cmd/mobileconfig-to-terraform/intune-settings-catalog-data/IntuneSettingsCategories_macOS.json",
		"../cmd/mobileconfig-to-terraform/intune-settings-catalog-data/IntuneSettingsVersion.json",
	)
	if err != nil {
		t.Fatalf("Failed to load catalog: %v", err)
	}

	if cat.CatalogDate() == "" {
		t.Error("Expected catalog date to be set")
	}

	// Verify some known entries exist
	knownPayloads := []string{
		"com.apple.dock",
		"com.apple.loginitems.managed",
	}

	for _, payload := range knownPayloads {
		groups := cat.RootGroups(payload)
		if len(groups) == 0 {
			t.Logf("Warning: No root groups found for %s (catalog may not include this payload)", payload)
		}
	}
}

func countFuzzyMatches(matches []converter.MatchedKey) int {
	count := 0
	for _, m := range matches {
		if m.MatchType == converter.MatchTypeFuzzy {
			count++
		}
	}
	return count
}

func countErrors(reports []converter.FileReport) int {
	count := 0
	for _, r := range reports {
		if r.Error != nil {
			count++
		}
	}
	return count
}

func countByType(reports []converter.FileReport, resourceType string) int {
	count := 0
	for _, r := range reports {
		if r.Error == nil && strings.Contains(r.ResourceType, resourceType) {
			count++
		}
	}
	return count
}
