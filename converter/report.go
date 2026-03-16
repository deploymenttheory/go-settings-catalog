package converter

import (
	"fmt"
	"strings"
	"time"
)

// ExportReport generates a detailed conversion report
type ExportReport struct {
	Timestamp       time.Time
	TotalFiles      int
	SuccessCount    int
	FailureCount    int
	CatalogCount    int
	CustomCount     int
	FileReports     []FileReport
}

// FileReport contains detailed information about a single file conversion
type FileReport struct {
	InputPath       string
	OutputPath      string
	ResourceType    string
	SettingCount    int
	MatchedKeys     []MatchedKey
	SkippedKeys     []SkippedKey
	SkippedPayloads []string
	WasSignedFile   bool
	Error           error
}

// GenerateMarkdownReport creates a markdown-formatted report
func (r *ExportReport) GenerateMarkdownReport() string {
	var md strings.Builder

	fmt.Fprintf(&md, "# Conversion Report\n\n")
	fmt.Fprintf(&md, "**Generated:** %s\n\n", r.Timestamp.Format(time.RFC3339))

	// Summary section
	fmt.Fprintf(&md, "## Summary\n\n")
	fmt.Fprintf(&md, "| Metric | Count |\n")
	fmt.Fprintf(&md, "|--------|-------|\n")
	fmt.Fprintf(&md, "| Total Files | %d |\n", r.TotalFiles)
	fmt.Fprintf(&md, "| Successful | %d |\n", r.SuccessCount)
	fmt.Fprintf(&md, "| Failed | %d |\n", r.FailureCount)
	fmt.Fprintf(&md, "| Settings Catalog | %d |\n", r.CatalogCount)
	fmt.Fprintf(&md, "| Custom Configuration | %d |\n", r.CustomCount)
	fmt.Fprintf(&md, "\n")

	// Settings Catalog conversions
	catalogFiles := r.filesByType("Settings Catalog")
	if len(catalogFiles) > 0 {
		fmt.Fprintf(&md, "## Settings Catalog Conversions (%d)\n\n", len(catalogFiles))
		fmt.Fprintf(&md, "These files were successfully converted to Settings Catalog format with all keys matched.\n\n")
		
		for _, file := range catalogFiles {
			fmt.Fprintf(&md, "### %s\n\n", file.InputPath)
			fmt.Fprintf(&md, "- **Output:** `%s`\n", file.OutputPath)
			fmt.Fprintf(&md, "- **Settings:** %d\n", file.SettingCount)
			
			if file.WasSignedFile {
				fmt.Fprintf(&md, "- **Note:** Signature was automatically removed\n")
			}
			
			// Group matched keys by type
			exactMatches := 0
			fuzzyMatches := 0
			for _, mk := range file.MatchedKeys {
				if mk.MatchType == MatchTypeExact {
					exactMatches++
				} else if mk.MatchType == MatchTypeFuzzy {
					fuzzyMatches++
				}
			}
			
			fmt.Fprintf(&md, "- **Matched Keys:** %d total", len(file.MatchedKeys))
			if fuzzyMatches > 0 {
				fmt.Fprintf(&md, " (%d exact, %d fuzzy)", exactMatches, fuzzyMatches)
			}
			fmt.Fprintf(&md, "\n")
			
			// Show fuzzy matches if any
			if fuzzyMatches > 0 {
				fmt.Fprintf(&md, "\n**Fuzzy Matches:**\n\n")
				for _, mk := range file.MatchedKeys {
					if mk.MatchType == MatchTypeFuzzy {
						fmt.Fprintf(&md, "- `%s` → `%s` (%.2f%% similarity)\n", 
							mk.Path, mk.CatalogName, mk.SimilarityScore*100)
					}
				}
			}
			
			fmt.Fprintf(&md, "\n")
		}
	}

	// Custom Configuration conversions
	customFiles := r.filesByType("Custom Configuration")
	if len(customFiles) > 0 {
		fmt.Fprintf(&md, "## Custom Configuration Conversions (%d)\n\n", len(customFiles))
		fmt.Fprintf(&md, "These files fell back to Custom Configuration format due to unmatched keys or payloads.\n\n")
		
		for _, file := range customFiles {
			fmt.Fprintf(&md, "### %s\n\n", file.InputPath)
			fmt.Fprintf(&md, "- **Output:** `%s`\n", file.OutputPath)
			fmt.Fprintf(&md, "- **Reason:** ")
			
			if len(file.SkippedPayloads) > 0 {
				fmt.Fprintf(&md, "%d payload(s) not in catalog", len(file.SkippedPayloads))
			}
			if len(file.SkippedKeys) > 0 {
				if len(file.SkippedPayloads) > 0 {
					fmt.Fprintf(&md, ", ")
				}
				fmt.Fprintf(&md, "%d key(s) not matched", len(file.SkippedKeys))
			}
			fmt.Fprintf(&md, "\n")
			
			if file.WasSignedFile {
				fmt.Fprintf(&md, "- **Note:** Signature was automatically removed\n")
			}
			
			// Show skipped payloads
			if len(file.SkippedPayloads) > 0 {
				fmt.Fprintf(&md, "\n**Skipped Payloads:**\n\n")
				for _, sp := range file.SkippedPayloads {
					fmt.Fprintf(&md, "- `%s`\n", sp)
				}
			}
			
			// Show skipped keys with nearest matches
			if len(file.SkippedKeys) > 0 {
				fmt.Fprintf(&md, "\n**Skipped Keys:**\n\n")
				for _, sk := range file.SkippedKeys {
					fmt.Fprintf(&md, "- `%s` = `%s`\n", sk.Path, sk.Value)
					
					if len(sk.NearestMatches) > 0 {
						fmt.Fprintf(&md, "  - Nearest matches:\n")
						for _, nm := range sk.NearestMatches {
							fmt.Fprintf(&md, "    - `%s` (%.2f%% similarity)\n", 
								nm.CatalogName, nm.SimilarityScore*100)
						}
					}
				}
			}
			
			// Show matched keys if any
			if len(file.MatchedKeys) > 0 {
				exactMatches := 0
				fuzzyMatches := 0
				for _, mk := range file.MatchedKeys {
					if mk.MatchType == MatchTypeExact {
						exactMatches++
					} else if mk.MatchType == MatchTypeFuzzy {
						fuzzyMatches++
					}
				}
				
				fmt.Fprintf(&md, "\n**Matched Keys:** %d total", len(file.MatchedKeys))
				if fuzzyMatches > 0 {
					fmt.Fprintf(&md, " (%d exact, %d fuzzy)", exactMatches, fuzzyMatches)
				}
				fmt.Fprintf(&md, "\n")
				
				if fuzzyMatches > 0 {
					fmt.Fprintf(&md, "\n**Fuzzy Matches:**\n\n")
					for _, mk := range file.MatchedKeys {
						if mk.MatchType == MatchTypeFuzzy {
							fmt.Fprintf(&md, "- `%s` → `%s` (%.2f%% similarity)\n", 
								mk.Path, mk.CatalogName, mk.SimilarityScore*100)
						}
					}
				}
			}
			
			fmt.Fprintf(&md, "\n")
		}
	}

	// Failed conversions
	failedFiles := r.filesByError()
	if len(failedFiles) > 0 {
		fmt.Fprintf(&md, "## Failed Conversions (%d)\n\n", len(failedFiles))
		
		for _, file := range failedFiles {
			fmt.Fprintf(&md, "### %s\n\n", file.InputPath)
			fmt.Fprintf(&md, "- **Error:** %v\n\n", file.Error)
		}
	}

	return md.String()
}

func (r *ExportReport) filesByType(resourceType string) []FileReport {
	var files []FileReport
	for _, file := range r.FileReports {
		if file.Error == nil && strings.Contains(file.ResourceType, resourceType) {
			files = append(files, file)
		}
	}
	return files
}

func (r *ExportReport) filesByError() []FileReport {
	var files []FileReport
	for _, file := range r.FileReports {
		if file.Error != nil {
			files = append(files, file)
		}
	}
	return files
}
