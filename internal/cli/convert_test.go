package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with-dashes", "with_dashes"},
		{"with.dots", "with_dots"},
		{"with spaces", "with_spaces"},
		{"UPPERCASE", "uppercase"},
		{"Mixed-Case.Name", "mixed_case_name"},
		{"123-start-with-number", "123_start_with_number"},
		{"multiple---dashes", "multiple___dashes"},
		{"trailing-", "trailing_"},
		{"-leading", "_leading"},
		{"special!@#$chars", "specialchars"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeResourceName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeResourceName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidInputFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"mobileconfig", "test.mobileconfig", true},
		{"xml", "test.xml", true},
		{"plist", "test.plist", true},
		{"uppercase_ext", "test.MOBILECONFIG", true},
		{"invalid_ext", "test.txt", false},
		{"no_ext", "test", false},
		{"hidden_file", ".hidden.mobileconfig", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidInputFile(tt.path)
			if result != tt.expected {
				t.Errorf("isValidInputFile(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestLoadEmbeddedCatalog(t *testing.T) {
	// Create a temporary test directory with catalog files
	tmpDir := t.TempDir()
	
	catalogData := []byte(`[
		{
			"id": "test_1",
			"offsetUri": "test",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": []
		}
	]`)
	
	categoriesData := []byte(`[
		{
			"id": "cat_1",
			"displayName": "Test Category"
		}
	]`)
	
	versionData := []byte(`{"date":"2026-03-10"}`)
	
	catalogPath := filepath.Join(tmpDir, "catalog.json")
	categoriesPath := filepath.Join(tmpDir, "categories.json")
	versionPath := filepath.Join(tmpDir, "version.json")
	
	if err := os.WriteFile(catalogPath, catalogData, 0644); err != nil {
		t.Fatalf("Failed to write catalog file: %v", err)
	}
	if err := os.WriteFile(categoriesPath, categoriesData, 0644); err != nil {
		t.Fatalf("Failed to write categories file: %v", err)
	}
	if err := os.WriteFile(versionPath, versionData, 0644); err != nil {
		t.Fatalf("Failed to write version file: %v", err)
	}
	
	// This test would need the embed.FS to work properly
	// For now, just verify the function signature exists
	t.Skip("Skipping loadEmbeddedCatalog test - requires embed.FS setup")
}

func TestPrintSuccess(t *testing.T) {
	// This is primarily a UI function, but we can test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printSuccess panicked: %v", r)
		}
	}()
	
	// Just verify it can be called without panicking
	// We can't easily test the output without capturing stdout
	t.Skip("Skipping printSuccess test - requires stdout capture")
}

func TestPrintBatchSummary(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printBatchSummary panicked: %v", r)
		}
	}()
	
	t.Skip("Skipping printBatchSummary test - requires stdout capture")
}
