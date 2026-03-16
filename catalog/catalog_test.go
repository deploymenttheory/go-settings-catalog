package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCatalog(t *testing.T) {
	cat := NewCatalog()
	if cat == nil {
		t.Fatal("NewCatalog returned nil")
	}
	if cat.entriesById == nil {
		t.Error("entriesById map not initialized")
	}
	if cat.categoriesById == nil {
		t.Error("categoriesById map not initialized")
	}
	if cat.groupsByOffsetURI == nil {
		t.Error("groupsByOffsetURI map not initialized")
	}
}

func TestLoadFromBytes(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "test_setting_1",
			"offsetUri": "TestSetting1",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Test Setting 1",
			"description": "A test setting",
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition",
				"minimumValue": 1,
				"maximumValue": 100
			}
		},
		{
			"id": "test_group_1",
			"offsetUri": "com.test.payload",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": ["test_setting_1"],
			"displayName": "Test Group"
		}
	]`)

	categoriesJSON := []byte(`[
		{
			"id": "cat_1",
			"displayName": "Test Category"
		}
	]`)

	versionJSON := []byte(`{
		"date": "2026-03-10"
	}`)

	cat := NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, categoriesJSON, versionJSON)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	if cat.CatalogDate() != "2026-03-10" {
		t.Errorf("Expected catalog date '2026-03-10', got '%s'", cat.CatalogDate())
	}

	entry := cat.Entry("test_setting_1")
	if entry == nil {
		t.Fatal("Entry 'test_setting_1' not found")
	}
	if entry.DisplayName != "Test Setting 1" {
		t.Errorf("Expected display name 'Test Setting 1', got '%s'", entry.DisplayName)
	}
	if entry.ValueDefinition == nil {
		t.Error("ValueDefinition not loaded")
	}
	if entry.ValueDefinition.MinimumValue == nil || *entry.ValueDefinition.MinimumValue != 1 {
		t.Error("MinimumValue not loaded correctly")
	}
}

func TestRootGroups(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "test_group_1",
			"offsetUri": "com.test.payload",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": []
		},
		{
			"id": "test_group_2",
			"offsetUri": "com.test.payload",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": []
		}
	]`)

	cat := NewCatalog()
	_ = cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	groups := cat.RootGroups("com.test.payload")
	if len(groups) != 2 {
		t.Errorf("Expected 2 root groups, got %d", len(groups))
	}

	groups = cat.RootGroups("com.nonexistent.payload")
	if len(groups) != 0 {
		t.Errorf("Expected 0 root groups for nonexistent payload, got %d", len(groups))
	}
}

func TestFindChild_ExactMatch(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "parent_group",
			"offsetUri": "parent",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": ["child_1", "child_2"]
		},
		{
			"id": "child_1",
			"offsetUri": "EnableFeature",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": []
		},
		{
			"id": "child_2",
			"offsetUri": "FeatureValue",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": []
		}
	]`)

	cat := NewCatalog()
	_ = cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	parent := cat.Entry("parent_group")
	if parent == nil {
		t.Fatal("Parent group not found")
	}

	// Test exact match (case insensitive)
	child := cat.FindChild("enablefeature", parent.ChildIDs)
	if child == nil {
		t.Fatal("Expected to find child with exact match")
	}
	if child.ID != "child_1" {
		t.Errorf("Expected child_1, got %s", child.ID)
	}
}

func TestFindChild_FuzzyMatch(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "parent_group",
			"offsetUri": "parent",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": ["child_1"]
		},
		{
			"id": "child_1",
			"offsetUri": "show-indicators-immutable",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Show Indicators Immutable"
		}
	]`)

	cat := NewCatalog()
	_ = cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	parent := cat.Entry("parent_group")
	if parent == nil {
		t.Fatal("Parent group not found")
	}

	// Test fuzzy match - should match despite extra "process" word
	child := cat.FindChild("show-process-indicators-immutable", parent.ChildIDs)
	if child == nil {
		t.Fatal("Expected fuzzy match to find child")
	}
	if child.ID != "child_1" {
		t.Errorf("Expected child_1, got %s", child.ID)
	}
}

func TestFindChildWithDetails(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "parent_group",
			"offsetUri": "parent",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSettingGroupCollectionDefinition",
			"childIds": ["child_1", "child_2", "child_3"]
		},
		{
			"id": "child_1",
			"offsetUri": "minimize-to-application",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Minimize To Application"
		},
		{
			"id": "child_2",
			"offsetUri": "minintoapp-immutable",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Minimize Into Application Immutable"
		},
		{
			"id": "child_3",
			"offsetUri": "position-immutable",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"displayName": "Position Immutable"
		}
	]`)

	cat := NewCatalog()
	_ = cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	parent := cat.Entry("parent_group")
	if parent == nil {
		t.Fatal("Parent group not found")
	}

	// Test with key that doesn't match well enough
	result := cat.FindChildWithDetails("minimize-to-application-immutable", parent.ChildIDs)
	
	if result.Entry != nil {
		t.Error("Expected no match for key below threshold")
	}
	if result.MatchType != "none" {
		t.Errorf("Expected match type 'none', got '%s'", result.MatchType)
	}
	if len(result.NearestMatches) == 0 {
		t.Error("Expected nearest matches to be populated")
	}
	if len(result.NearestMatches) > 0 {
		if result.NearestMatches[0].Entry.DisplayName != "Minimize To Application" {
			t.Errorf("Expected first nearest match to be 'Minimize To Application', got '%s'", 
				result.NearestMatches[0].Entry.DisplayName)
		}
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		minScore float64
	}{
		{"test", "test", 1.0},
		{"test-value", "testvalue", 1.0},
		{"show-indicators", "showindicators", 1.0},
		{"completely", "different", 0.0},
		{"", "", 1.0},
		{"a", "", 0.0},
		{"", "a", 0.0},
	}

	for _, tt := range tests {
		score := stringSimilarity(tt.s1, tt.s2)
		if tt.minScore == 1.0 && score != 1.0 {
			t.Errorf("stringSimilarity(%q, %q) = %.2f, expected 1.0", tt.s1, tt.s2, score)
		} else if tt.minScore > 0 && score < tt.minScore {
			t.Errorf("stringSimilarity(%q, %q) = %.2f, expected >= %.2f", tt.s1, tt.s2, score, tt.minScore)
		}
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "def", 3},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		distance := levenshteinDistance(tt.s1, tt.s2)
		if distance != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, distance, tt.expected)
		}
	}
}

func TestOffsetKey(t *testing.T) {
	tests := []struct {
		uri      string
		expected string
	}{
		{"EnableFeature", "EnableFeature"},
		{"Settings/EnableFeature", "EnableFeature"},
		{"Applications/[{0}]/BundleID", "BundleID"},
		{"Settings/[{0}]/Values/[{1}]/Name", "Name"},
		{"", ""},
		{"/", ""},
	}

	for _, tt := range tests {
		result := offsetKey(tt.uri)
		if result != tt.expected {
			t.Errorf("offsetKey(%q) = %q, expected %q", tt.uri, result, tt.expected)
		}
	}
}

func TestLoadFromBytes_InvalidJSON(t *testing.T) {
	cat := NewCatalog()
	
	tests := []struct {
		name           string
		catalogData    []byte
		categoriesData []byte
		versionData    []byte
		errorContains  string
	}{
		{
			name:           "invalid_catalog",
			catalogData:    []byte(`{invalid json`),
			categoriesData: []byte(`[]`),
			versionData:    []byte(`{"date":"2026-03-10"}`),
			errorContains:  "failed to parse catalog JSON",
		},
		{
			name:           "invalid_categories",
			catalogData:    []byte(`[]`),
			categoriesData: []byte(`{invalid json`),
			versionData:    []byte(`{"date":"2026-03-10"}`),
			errorContains:  "failed to parse categories JSON",
		},
		{
			name:           "invalid_version",
			catalogData:    []byte(`[]`),
			categoriesData: []byte(`[]`),
			versionData:    []byte(`{invalid json`),
			errorContains:  "failed to parse version JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cat.LoadFromBytes(tt.catalogData, tt.categoriesData, tt.versionData)
			if err == nil {
				t.Fatal("Expected error for invalid JSON")
			}
			if !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("Expected error containing %q, got %v", tt.errorContains, err)
			}
		})
	}
}

func TestLoadFromBytes_SkipsInvalidEntries(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "",
			"offsetUri": "test",
			"@odata.type": "#microsoft.graph.test"
		},
		{
			"id": "valid_1",
			"offsetUri": "",
			"@odata.type": "#microsoft.graph.test"
		},
		{
			"id": "valid_2",
			"offsetUri": "test2",
			"@odata.type": ""
		},
		{
			"id": "valid_3",
			"offsetUri": "test3",
			"@odata.type": "#microsoft.graph.test"
		}
	]`)

	cat := NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	// Only valid_3 should be loaded (has all required fields)
	if cat.Entry("valid_3") == nil {
		t.Error("Expected valid_3 to be loaded")
	}
	if cat.Entry("") != nil {
		t.Error("Expected empty ID entry to be skipped")
	}
	if cat.Entry("valid_1") != nil {
		t.Error("Expected entry with empty offsetUri to be skipped")
	}
	if cat.Entry("valid_2") != nil {
		t.Error("Expected entry with empty odataType to be skipped")
	}
}

func TestLoadFromBytes_HandlesOptionsWithoutItemID(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "test_choice",
			"offsetUri": "testchoice",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition",
			"childIds": [],
			"options": [
				{
					"itemId": "",
					"optionValue": {"value": "invalid"}
				},
				{
					"itemId": "valid_option",
					"optionValue": {"value": "valid"}
				}
			]
		}
	]`)

	cat := NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	entry := cat.Entry("test_choice")
	if entry == nil {
		t.Fatal("Expected entry to be loaded")
	}

	// Should have only 1 option (the one with valid itemId)
	if len(entry.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(entry.Options))
	}
	if entry.Options[0].ItemID != "valid_option" {
		t.Errorf("Expected valid_option, got %s", entry.Options[0].ItemID)
	}
}

func TestLoadFromBytes_HandlesIntegerOptionValues(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "test_choice",
			"offsetUri": "testchoice",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationChoiceSettingDefinition",
			"childIds": [],
			"options": [
				{
					"itemId": "option_1",
					"optionValue": {"value": 123}
				}
			]
		}
	]`)

	cat := NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	entry := cat.Entry("test_choice")
	if entry == nil {
		t.Fatal("Expected entry to be loaded")
	}

	if len(entry.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(entry.Options))
	}
	if entry.Options[0].OptionValue != "123" {
		t.Errorf("Expected option value '123', got %s", entry.Options[0].OptionValue)
	}
}

func TestLoadFromBytes_HandlesValueDefinition(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "test_setting",
			"offsetUri": "testsetting",
			"@odata.type": "#microsoft.graph.deviceManagementConfigurationSimpleSettingDefinition",
			"childIds": [],
			"valueDefinition": {
				"@odata.type": "#microsoft.graph.deviceManagementConfigurationIntegerSettingValueDefinition",
				"minimumValue": 10,
				"maximumValue": 100,
				"isSecret": false
			}
		}
	]`)

	cat := NewCatalog()
	err := cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	entry := cat.Entry("test_setting")
	if entry == nil {
		t.Fatal("Expected entry to be loaded")
	}

	if entry.ValueDefinition == nil {
		t.Fatal("Expected ValueDefinition to be loaded")
	}
	if entry.ValueDefinition.MinimumValue == nil || *entry.ValueDefinition.MinimumValue != 10 {
		t.Error("Expected MinimumValue to be 10")
	}
	if entry.ValueDefinition.MaximumValue == nil || *entry.ValueDefinition.MaximumValue != 100 {
		t.Error("Expected MaximumValue to be 100")
	}
}

func TestCatalogDate(t *testing.T) {
	cat := NewCatalog()
	err := cat.LoadFromBytes(
		[]byte(`[]`),
		[]byte(`[]`),
		[]byte(`{"date":"2026-03-15"}`),
	)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	if cat.CatalogDate() != "2026-03-15" {
		t.Errorf("Expected catalog date '2026-03-15', got '%s'", cat.CatalogDate())
	}
}

func TestLoadFromFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	catalogPath := filepath.Join(tmpDir, "catalog.json")
	categoriesPath := filepath.Join(tmpDir, "categories.json")
	versionPath := filepath.Join(tmpDir, "version.json")
	
	catalogData := []byte(`[{"id":"test","offsetUri":"test","@odata.type":"#test","childIds":[]}]`)
	categoriesData := []byte(`[]`)
	versionData := []byte(`{"date":"2026-03-10"}`)
	
	if err := os.WriteFile(catalogPath, catalogData, 0644); err != nil {
		t.Fatalf("Failed to write catalog: %v", err)
	}
	if err := os.WriteFile(categoriesPath, categoriesData, 0644); err != nil {
		t.Fatalf("Failed to write categories: %v", err)
	}
	if err := os.WriteFile(versionPath, versionData, 0644); err != nil {
		t.Fatalf("Failed to write version: %v", err)
	}

	cat := NewCatalog()
	err := cat.LoadFromFiles(catalogPath, categoriesPath, versionPath)
	if err != nil {
		t.Fatalf("LoadFromFiles failed: %v", err)
	}

	if cat.Entry("test") == nil {
		t.Error("Expected entry to be loaded")
	}
}

func TestLoadFromFiles_MissingFiles(t *testing.T) {
	cat := NewCatalog()
	
	err := cat.LoadFromFiles("nonexistent.json", "nonexistent.json", "nonexistent.json")
	if err == nil {
		t.Fatal("Expected error for missing files")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Expected read error, got %v", err)
	}
}

func TestFindChildWithDetails_EmptyChildIds(t *testing.T) {
	cat := NewCatalog()
	_ = cat.LoadFromBytes([]byte(`[]`), []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	result := cat.FindChildWithDetails("anykey", []string{})
	
	if result.Entry != nil {
		t.Error("Expected nil entry for empty childIds")
	}
	if result.MatchType != "none" {
		t.Errorf("Expected match type 'none', got '%s'", result.MatchType)
	}
}

func TestFindChildWithDetails_NonexistentChildId(t *testing.T) {
	cat := NewCatalog()
	_ = cat.LoadFromBytes([]byte(`[]`), []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	result := cat.FindChildWithDetails("anykey", []string{"nonexistent_id"})
	
	if result.Entry != nil {
		t.Error("Expected nil entry for nonexistent child ID")
	}
}

func TestFindChild_ReturnsEntry(t *testing.T) {
	catalogJSON := []byte(`[
		{
			"id": "parent",
			"offsetUri": "parent",
			"@odata.type": "#test",
			"childIds": ["child_1"]
		},
		{
			"id": "child_1",
			"offsetUri": "testkey",
			"@odata.type": "#test",
			"childIds": []
		}
	]`)

	cat := NewCatalog()
	_ = cat.LoadFromBytes(catalogJSON, []byte(`[]`), []byte(`{"date":"2026-03-10"}`))

	parent := cat.Entry("parent")
	child := cat.FindChild("testkey", parent.ChildIDs)
	
	if child == nil {
		t.Fatal("Expected to find child")
	}
	if child.ID != "child_1" {
		t.Errorf("Expected child_1, got %s", child.ID)
	}
}

func TestLoadCategories(t *testing.T) {
	categoriesJSON := []byte(`[
		{
			"id": "cat_1",
			"displayName": "Category 1",
			"description": "Test category",
			"categoryDescription": "Category description",
			"childCategoryIds": ["cat_2"]
		}
	]`)

	cat := NewCatalog()
	err := cat.LoadFromBytes([]byte(`[]`), categoriesJSON, []byte(`{"date":"2026-03-10"}`))
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	// Categories are loaded but not exposed via public API
	// This test verifies no error occurs during loading
}

func TestMinMaxHelpers(t *testing.T) {
	if min(1, 2, 3) != 1 {
		t.Error("min(1,2,3) should be 1")
	}
	if min(3, 1, 2) != 1 {
		t.Error("min(3,1,2) should be 1")
	}
	if max(1, 2) != 2 {
		t.Error("max(1,2) should be 2")
	}
	if max(5, 3) != 5 {
		t.Error("max(5,3) should be 5")
	}
}
