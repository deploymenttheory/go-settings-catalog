package converter

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGenerateMarkdownReport_Empty(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   0,
		SuccessCount: 0,
		FailureCount: 0,
		CatalogCount: 0,
		CustomCount:  0,
		FileReports:  []FileReport{},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "# Conversion Report") {
		t.Error("Expected report title")
	}

	if !strings.Contains(md, "| Total Files | 0 |") {
		t.Error("Expected summary table with zero counts")
	}
}

func TestGenerateMarkdownReport_SettingsCatalog(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   1,
		SuccessCount: 1,
		FailureCount: 0,
		CatalogCount: 1,
		CustomCount:  0,
		FileReports: []FileReport{
			{
				InputPath:    "test.mobileconfig",
				OutputPath:   "test.tf",
				ResourceType: "Settings Catalog Terraform HCL",
				SettingCount: 2,
				MatchedKeys: []MatchedKey{
					{
						Path:            "com.apple.dock.largesize",
						Value:           "64",
						CatalogID:       "test_id_1",
						CatalogName:     "Large Size",
						MatchType:       MatchTypeExact,
						SimilarityScore: 1.0,
					},
					{
						Path:            "com.apple.dock.tilesize",
						Value:           "48",
						CatalogID:       "test_id_2",
						CatalogName:     "Tile Size",
						MatchType:       MatchTypeExact,
						SimilarityScore: 1.0,
					},
				},
				SkippedKeys:     []SkippedKey{},
				SkippedPayloads: []string{},
				WasSignedFile:   false,
				Error:           nil,
			},
		},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "## Settings Catalog Conversions (1)") {
		t.Error("Expected Settings Catalog section")
	}

	if !strings.Contains(md, "test.mobileconfig") {
		t.Error("Expected input path")
	}

	if !strings.Contains(md, "test.tf") {
		t.Error("Expected output path")
	}

	if !strings.Contains(md, "**Settings:** 2") {
		t.Error("Expected settings count")
	}

	if !strings.Contains(md, "**Matched Keys:** 2 total") {
		t.Error("Expected matched keys count")
	}
}

func TestGenerateMarkdownReport_CustomConfig(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   1,
		SuccessCount: 1,
		FailureCount: 0,
		CatalogCount: 0,
		CustomCount:  1,
		FileReports: []FileReport{
			{
				InputPath:    "unknown.mobileconfig",
				OutputPath:   "unknown.tf",
				ResourceType: "Custom Configuration Terraform HCL",
				SettingCount: 0,
				MatchedKeys:  []MatchedKey{},
				SkippedKeys: []SkippedKey{
					{
						Path:  "com.apple.dock.unknownkey",
						Value: "value",
						NearestMatches: []NearestMatch{
							{
								CatalogID:       "test_id_1",
								CatalogName:     "Unknown Key Similar",
								SimilarityScore: 0.72,
							},
						},
					},
				},
				SkippedPayloads: []string{"com.unknown.payload"},
				WasSignedFile:   false,
				Error:           nil,
			},
		},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "## Custom Configuration Conversions (1)") {
		t.Error("Expected Custom Configuration section")
	}

	if !strings.Contains(md, "**Reason:**") {
		t.Error("Expected reason for custom config")
	}

	if !strings.Contains(md, "1 payload(s) not in catalog") {
		t.Error("Expected skipped payloads count")
	}

	if !strings.Contains(md, "1 key(s) not matched") {
		t.Error("Expected skipped keys count")
	}

	if !strings.Contains(md, "**Skipped Payloads:**") {
		t.Error("Expected skipped payloads section")
	}

	if !strings.Contains(md, "com.unknown.payload") {
		t.Error("Expected skipped payload name")
	}

	if !strings.Contains(md, "**Skipped Keys:**") {
		t.Error("Expected skipped keys section")
	}

	if !strings.Contains(md, "com.apple.dock.unknownkey") {
		t.Error("Expected skipped key path")
	}

	if !strings.Contains(md, "Nearest matches:") {
		t.Error("Expected nearest matches section")
	}

	if !strings.Contains(md, "72.00% similarity") {
		t.Error("Expected similarity score")
	}
}

func TestGenerateMarkdownReport_FuzzyMatches(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   1,
		SuccessCount: 1,
		FailureCount: 0,
		CatalogCount: 1,
		CustomCount:  0,
		FileReports: []FileReport{
			{
				InputPath:    "fuzzy.mobileconfig",
				OutputPath:   "fuzzy.tf",
				ResourceType: "Settings Catalog Terraform HCL",
				SettingCount: 1,
				MatchedKeys: []MatchedKey{
					{
						Path:            "com.apple.dock.show-process-indicators",
						Value:           "1",
						CatalogID:       "test_id_1",
						CatalogName:     "Show Indicators",
						MatchType:       MatchTypeFuzzy,
						SimilarityScore: 0.85,
					},
				},
				SkippedKeys:     []SkippedKey{},
				SkippedPayloads: []string{},
				WasSignedFile:   false,
				Error:           nil,
			},
		},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "**Matched Keys:** 1 total (0 exact, 1 fuzzy)") {
		t.Error("Expected fuzzy match breakdown")
	}

	if !strings.Contains(md, "**Fuzzy Matches:**") {
		t.Error("Expected fuzzy matches section")
	}

	if !strings.Contains(md, "85.00% similarity") {
		t.Error("Expected similarity score for fuzzy match")
	}
}

func TestGenerateMarkdownReport_FailedConversions(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   2,
		SuccessCount: 1,
		FailureCount: 1,
		CatalogCount: 1,
		CustomCount:  0,
		FileReports: []FileReport{
			{
				InputPath:    "success.mobileconfig",
				OutputPath:   "success.tf",
				ResourceType: "Settings Catalog Terraform HCL",
				SettingCount: 1,
				Error:        nil,
			},
			{
				InputPath:  "failed.mobileconfig",
				OutputPath: "",
				Error:      errors.New("invalid plist format"),
			},
		},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "| Failed | 1 |") {
		t.Error("Expected failed count in summary")
	}

	if !strings.Contains(md, "## Failed Conversions (1)") {
		t.Error("Expected failed conversions section")
	}

	if !strings.Contains(md, "failed.mobileconfig") {
		t.Error("Expected failed file path")
	}

	if !strings.Contains(md, "invalid plist format") {
		t.Error("Expected error message")
	}
}

func TestGenerateMarkdownReport_SignedFile(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   1,
		SuccessCount: 1,
		FailureCount: 0,
		CatalogCount: 1,
		CustomCount:  0,
		FileReports: []FileReport{
			{
				InputPath:       "signed.mobileconfig",
				OutputPath:      "signed.tf",
				ResourceType:    "Settings Catalog Terraform HCL",
				SettingCount:    1,
				MatchedKeys:     []MatchedKey{},
				SkippedKeys:     []SkippedKey{},
				SkippedPayloads: []string{},
				WasSignedFile:   true,
				Error:           nil,
			},
		},
	}

	md := report.GenerateMarkdownReport()

	if !strings.Contains(md, "**Note:** Signature was automatically removed") {
		t.Error("Expected signature removal note")
	}
}

func TestFilesByType(t *testing.T) {
	report := &ExportReport{
		FileReports: []FileReport{
			{ResourceType: "Settings Catalog Terraform HCL", Error: nil},
			{ResourceType: "Custom Configuration Terraform HCL", Error: nil},
			{ResourceType: "Settings Catalog JSON", Error: nil},
			{ResourceType: "Failed", Error: errors.New("test error")},
		},
	}

	catalogFiles := report.filesByType("Settings Catalog")
	if len(catalogFiles) != 2 {
		t.Errorf("Expected 2 Settings Catalog files, got %d", len(catalogFiles))
	}

	customFiles := report.filesByType("Custom Configuration")
	if len(customFiles) != 1 {
		t.Errorf("Expected 1 Custom Configuration file, got %d", len(customFiles))
	}
}

func TestFilesByError(t *testing.T) {
	report := &ExportReport{
		FileReports: []FileReport{
			{InputPath: "success1.mobileconfig", Error: nil},
			{InputPath: "failed1.mobileconfig", Error: errors.New("error 1")},
			{InputPath: "success2.mobileconfig", Error: nil},
			{InputPath: "failed2.mobileconfig", Error: errors.New("error 2")},
		},
	}

	failedFiles := report.filesByError()
	if len(failedFiles) != 2 {
		t.Errorf("Expected 2 failed files, got %d", len(failedFiles))
	}

	for _, file := range failedFiles {
		if file.Error == nil {
			t.Errorf("File %s should have an error", file.InputPath)
		}
	}
}

func TestGenerateMarkdownReport_MixedScenarios(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   4,
		SuccessCount: 3,
		FailureCount: 1,
		CatalogCount: 2,
		CustomCount:  1,
		FileReports: []FileReport{
			{
				InputPath:    "catalog1.mobileconfig",
				OutputPath:   "catalog1.tf",
				ResourceType: "Settings Catalog Terraform HCL",
				SettingCount: 3,
				MatchedKeys: []MatchedKey{
					{Path: "key1", MatchType: MatchTypeExact, SimilarityScore: 1.0},
					{Path: "key2", MatchType: MatchTypeExact, SimilarityScore: 1.0},
					{Path: "key3", MatchType: MatchTypeFuzzy, SimilarityScore: 0.82},
				},
				Error: nil,
			},
			{
				InputPath:    "catalog2.mobileconfig",
				OutputPath:   "catalog2.tf",
				ResourceType: "Settings Catalog Terraform HCL",
				SettingCount: 1,
				MatchedKeys: []MatchedKey{
					{Path: "key1", MatchType: MatchTypeExact, SimilarityScore: 1.0},
				},
				WasSignedFile: true,
				Error:         nil,
			},
			{
				InputPath:       "custom1.mobileconfig",
				OutputPath:      "custom1.tf",
				ResourceType:    "Custom Configuration Terraform HCL",
				SettingCount:    0,
				SkippedPayloads: []string{"com.unknown.payload"},
				SkippedKeys: []SkippedKey{
					{
						Path:  "com.apple.test.unknownkey",
						Value: "value",
						NearestMatches: []NearestMatch{
							{CatalogName: "Similar Key", SimilarityScore: 0.68},
						},
					},
				},
				Error: nil,
			},
			{
				InputPath: "failed.mobileconfig",
				Error:     errors.New("not a valid plist"),
			},
		},
	}

	md := report.GenerateMarkdownReport()

	// Verify summary
	if !strings.Contains(md, "| Total Files | 4 |") {
		t.Error("Expected total files count")
	}
	if !strings.Contains(md, "| Successful | 3 |") {
		t.Error("Expected success count")
	}
	if !strings.Contains(md, "| Failed | 1 |") {
		t.Error("Expected failure count")
	}
	if !strings.Contains(md, "| Settings Catalog | 2 |") {
		t.Error("Expected catalog count")
	}
	if !strings.Contains(md, "| Custom Configuration | 1 |") {
		t.Error("Expected custom count")
	}

	// Verify sections exist
	if !strings.Contains(md, "## Settings Catalog Conversions (2)") {
		t.Error("Expected Settings Catalog section")
	}
	if !strings.Contains(md, "## Custom Configuration Conversions (1)") {
		t.Error("Expected Custom Configuration section")
	}
	if !strings.Contains(md, "## Failed Conversions (1)") {
		t.Error("Expected Failed Conversions section")
	}

	// Verify fuzzy match details
	if !strings.Contains(md, "(2 exact, 1 fuzzy)") {
		t.Error("Expected fuzzy match breakdown for catalog1")
	}
	if !strings.Contains(md, "82.00% similarity") {
		t.Error("Expected similarity score for fuzzy match")
	}

	// Verify signed file note
	if !strings.Contains(md, "**Note:** Signature was automatically removed") {
		t.Error("Expected signature removal note")
	}

	// Verify custom config details
	if !strings.Contains(md, "**Skipped Payloads:**") {
		t.Error("Expected skipped payloads section")
	}
	if !strings.Contains(md, "com.unknown.payload") {
		t.Error("Expected skipped payload")
	}
	if !strings.Contains(md, "**Skipped Keys:**") {
		t.Error("Expected skipped keys section")
	}
	if !strings.Contains(md, "com.apple.test.unknownkey") {
		t.Error("Expected skipped key")
	}
	if !strings.Contains(md, "68.00% similarity") {
		t.Error("Expected nearest match similarity")
	}

	// Verify failed conversion
	if !strings.Contains(md, "not a valid plist") {
		t.Error("Expected error message")
	}
}

func TestGenerateMarkdownReport_CustomConfigWithMatchedKeys(t *testing.T) {
	report := &ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   1,
		SuccessCount: 1,
		FailureCount: 0,
		CatalogCount: 0,
		CustomCount:  1,
		FileReports: []FileReport{
			{
				InputPath:    "partial.mobileconfig",
				OutputPath:   "partial.tf",
				ResourceType: "Custom Configuration Terraform HCL",
				SettingCount: 0,
				MatchedKeys: []MatchedKey{
					{
						Path:            "com.apple.dock.largesize",
						Value:           "64",
						CatalogID:       "test_id_1",
						CatalogName:     "Large Size",
						MatchType:       MatchTypeExact,
						SimilarityScore: 1.0,
					},
					{
						Path:            "com.apple.dock.show-indicators",
						Value:           "1",
						CatalogID:       "test_id_2",
						CatalogName:     "Show Indicators",
						MatchType:       MatchTypeFuzzy,
						SimilarityScore: 0.88,
					},
				},
				SkippedKeys: []SkippedKey{
					{
						Path:  "com.apple.dock.unknownkey",
						Value: "value",
					},
				},
				SkippedPayloads: []string{},
				WasSignedFile:   false,
				Error:           nil,
			},
		},
	}

	md := report.GenerateMarkdownReport()

	// Should show matched keys even in custom config
	if !strings.Contains(md, "**Matched Keys:** 2 total (1 exact, 1 fuzzy)") {
		t.Error("Expected matched keys count in custom config")
	}

	if !strings.Contains(md, "**Fuzzy Matches:**") {
		t.Error("Expected fuzzy matches section in custom config")
	}

	if !strings.Contains(md, "88.00% similarity") {
		t.Error("Expected fuzzy match similarity in custom config")
	}
}
