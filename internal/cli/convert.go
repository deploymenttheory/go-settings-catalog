package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/deploymenttheory/go-settings-catalog/catalog"
	"github.com/deploymenttheory/go-settings-catalog/converter"
)

var (
	resourceName string
	outputFormat string
)

var convertCmd = &cobra.Command{
	Use:   "convert <input-file>",
	Short: "Convert a single configuration file to Intune Settings Catalog",
	Long: `Convert a single macOS configuration file to Intune Settings Catalog format.

For batch processing multiple files, use the 'batch' command instead.

Supported input formats:
  - .mobileconfig (signed or unsigned)
  - .plist
  - .xml

Supported output formats:
  - json      : Settings Catalog JSON (Microsoft Graph API)
  - terraform : Terraform HCL configuration (default)

Examples:
  # Convert to Terraform HCL
  mobileconfig-to-terraform convert profile.mobileconfig

  # Convert to JSON with custom name
  mobileconfig-to-terraform convert profile.mobileconfig -n my_policy -f json

  # Specify output directory
  mobileconfig-to-terraform convert profile.mobileconfig -o ./output`,
	Args: cobra.ExactArgs(1),
	Run:  runConvert,
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().
		StringVarP(&resourceName, "name", "n", "", "resource name (default: derived from filename)")
	convertCmd.Flags().
		StringVarP(&outputFormat, "format", "f", "terraform", "output format: json, terraform")
}

func runConvert(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Derive resource name if not provided
	if resourceName == "" {
		baseName := filepath.Base(inputPath)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		resourceName = sanitizeResourceName(baseName)
	}

	// Load embedded catalog
	cat := loadEmbeddedCatalog()

	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Convert to Settings Catalog JSON
	conv := converter.NewMobileconfigConverter(cat)
	result, err := conv.Convert(data, resourceName)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	// Output in requested format
	switch outputFormat {
	case "json":
		outputJSON(result, resourceName)
	case "terraform", "tf", "hcl":
		outputTerraform(result, resourceName)
	default:
		log.Fatalf("Unknown format: %s (supported: json, terraform)", outputFormat)
	}
}

func loadEmbeddedCatalog() *catalog.Catalog {
	cat := catalog.NewCatalog()

	// Read from embedded filesystem
	catalogData, err := catalogFS.ReadFile("intune-settings-catalog-data/IntuneSettingsCatalog_macOS.json")
	if err != nil {
		log.Fatalf("Failed to load embedded Intune Settings Catalog: %v", err)
	}

	categoriesData, err := catalogFS.ReadFile("intune-settings-catalog-data/IntuneSettingsCategories_macOS.json")
	if err != nil {
		log.Fatalf("Failed to load embedded Intune Settings Catalog categories: %v", err)
	}

	versionData, err := catalogFS.ReadFile("intune-settings-catalog-data/IntuneSettingsVersion.json")
	if err != nil {
		log.Fatalf("Failed to load embedded Intune Settings Catalog version: %v", err)
	}

	if err := cat.LoadFromBytes(catalogData, categoriesData, versionData); err != nil {
		log.Fatalf("Failed to parse catalog: %v", err)
	}

	if verbose {
		fmt.Printf("✓ Loaded catalog (date: %s)\n", cat.CatalogDate())
	}

	return cat
}

func outputJSON(result *converter.ConversionResult, name string) {
	jsonExporter := converter.NewJSONExporter()
	jsonData, err := jsonExporter.ExportFromResult(result)
	if err != nil {
		log.Fatalf("JSON export failed: %v", err)
	}

	outputPath := filepath.Join(outputDir, name+".json")
	if err := os.WriteFile(outputPath, jsonData, 0o600); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	// Determine resource type for display
	resourceType := "Settings Catalog JSON"
	if result.SettingCount == 0 || len(result.SkippedKeys) > 0 || len(result.SkippedPayloads) > 0 {
		resourceType = "Custom Configuration JSON"
	}

	printSuccess(resourceType, outputPath, result)
}

func outputTerraform(result *converter.ConversionResult, name string) {
	tfExporter := converter.NewTerraformExporter()
	hcl, err := tfExporter.ExportFromResult(result, name)
	if err != nil {
		log.Fatalf("Terraform export failed: %v", err)
	}

	outputPath := filepath.Join(outputDir, name+".tf")
	if err := os.WriteFile(outputPath, []byte(hcl), 0o600); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	// Determine resource type for display
	resourceType := "Terraform HCL (Settings Catalog)"
	if result.SettingCount == 0 || len(result.SkippedKeys) > 0 || len(result.SkippedPayloads) > 0 {
		resourceType = "Terraform HCL (Custom Configuration)"
	}

	printSuccess(resourceType, outputPath, result)
}

func printSuccess(formatName, outputPath string, result *converter.ConversionResult) {
	fmt.Printf("\n✓ Conversion successful!\n")
	fmt.Printf("  Format: %s\n", formatName)
	fmt.Printf("  Output: %s\n", outputPath)
	fmt.Printf("  Settings: %d\n", result.SettingCount)

	if result.WasSignedFile {
		fmt.Printf("  Note: Signature was automatically removed\n")
	}

	if len(result.SkippedPayloads) > 0 {
		fmt.Printf(
			"\n⚠ Skipped %d payload(s) (not in Settings Catalog)\n",
			len(result.SkippedPayloads),
		)
		if verbose {
			for _, p := range result.SkippedPayloads {
				fmt.Printf("  - %s\n", p)
			}
		}
	}

	if len(result.SkippedKeys) > 0 {
		fmt.Printf("⚠ Skipped %d key(s) (no catalog match)\n", len(result.SkippedKeys))
		if verbose && len(result.SkippedKeys) <= 10 {
			for _, k := range result.SkippedKeys {
				fmt.Printf("  - %s = %s\n", k.Path, k.Value)
			}
		}
	}
}

func sanitizeResourceName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")

	var cleaned []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned = append(cleaned, r)
		}
	}
	return string(cleaned)
}
