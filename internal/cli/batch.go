package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/deploymenttheory/go-settings-catalog/catalog"
	"github.com/deploymenttheory/go-settings-catalog/converter"
)

var (
	batchRecursive bool
	batchContinue  bool
	batchWorkers   int
)

var batchCmd = &cobra.Command{
	Use:   "batch <input-path>",
	Short: "Batch convert multiple configuration files",
	Long: `Batch convert multiple macOS configuration files to Intune Settings Catalog format.

Input can be:
  - A directory containing .mobileconfig, .plist, or .xml files
  - A glob pattern (e.g., "profiles/*.mobileconfig")

Supported input formats:
  - .mobileconfig (signed or unsigned)
  - .plist
  - .xml

Supported output formats:
  - json      : Settings Catalog JSON (Microsoft Graph API)
  - terraform : Terraform HCL configuration (default)

Examples:
  # Convert all mobileconfig files in a directory
  mobileconfig-to-terraform batch ./profiles

  # Convert recursively
  mobileconfig-to-terraform batch ./profiles -r

  # Convert with glob pattern
  mobileconfig-to-terraform batch "profiles/**/*.mobileconfig"

  # Convert to JSON format
  mobileconfig-to-terraform batch ./profiles -f json

  # Continue on errors
  mobileconfig-to-terraform batch ./profiles --continue`,
	Args: cobra.ExactArgs(1),
	Run:  runBatch,
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.Flags().
		StringVarP(&outputFormat, "format", "f", "terraform", "output format: json, terraform")
	batchCmd.Flags().
		BoolVarP(&batchRecursive, "recursive", "r", false, "recursively process subdirectories")
	batchCmd.Flags().
		BoolVar(&batchContinue, "continue", false, "continue processing on errors")
	batchCmd.Flags().
		IntVarP(&batchWorkers, "workers", "w", 4, "number of parallel workers")
}

func runBatch(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	cat := loadEmbeddedCatalog()

	files, err := findInputFiles(inputPath, batchRecursive)
	if err != nil {
		log.Fatalf("Failed to find input files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No input files found in: %s", inputPath)
	}

	fmt.Printf("Found %d file(s) to convert\n", len(files))
	if verbose {
		fmt.Printf("Using %d worker(s)\n", batchWorkers)
	}

	results := processBatch(cat, files, batchWorkers)

	printBatchSummary(results)
	
	// Generate and save detailed report
	report := generateExportReport(results)
	reportPath := filepath.Join(outputDir, "conversion-report.md")
	if err := os.WriteFile(reportPath, []byte(report.GenerateMarkdownReport()), 0o644); err != nil {
		log.Printf("Warning: Failed to write report: %v", err)
	} else {
		fmt.Printf("\n📊 Detailed report saved to: %s\n", reportPath)
	}
}

type batchResult struct {
	InputPath  string
	OutputPath string
	Success    bool
	Error      error
	Result     *converter.ConversionResult
}

func processBatch(cat *catalog.Catalog, files []string, workers int) []batchResult {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []batchResult
	)

	jobs := make(chan string, len(files))
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for inputPath := range jobs {
				result := processFile(cat, inputPath)

				mu.Lock()
				results = append(results, result)
				mu.Unlock()

				if !result.Success && !batchContinue {
					break
				}
			}
		}()
	}

	wg.Wait()
	return results
}

func processFile(cat *catalog.Catalog, inputPath string) batchResult {
	result := batchResult{
		InputPath: inputPath,
		Success:   false,
	}

	baseName := filepath.Base(inputPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	resourceName := sanitizeResourceName(baseName)

	data, err := os.ReadFile(inputPath)
	if err != nil {
		result.Error = fmt.Errorf("read failed: %w", err)
		if verbose {
			fmt.Printf("✗ %s: %v\n", inputPath, err)
		}
		return result
	}

	conv := converter.NewMobileconfigConverter(cat)
	convResult, err := conv.Convert(data, resourceName)
	if err != nil {
		result.Error = fmt.Errorf("conversion failed: %w", err)
		if verbose {
			fmt.Printf("✗ %s: %v\n", inputPath, err)
		}
		return result
	}

	result.Result = convResult

	var outputPath string
	switch outputFormat {
	case "json":
		jsonExporter := converter.NewJSONExporter()
		jsonData, err := jsonExporter.ExportFromResult(convResult)
		if err != nil {
			result.Error = fmt.Errorf("json export failed: %w", err)
			if verbose {
				fmt.Printf("✗ %s: %v\n", inputPath, err)
			}
			return result
		}

		outputPath = filepath.Join(outputDir, resourceName+".json")
		if err := os.WriteFile(outputPath, jsonData, 0o600); err != nil {
			result.Error = fmt.Errorf("write failed: %w", err)
			if verbose {
				fmt.Printf("✗ %s: %v\n", inputPath, err)
			}
			return result
		}
	case "terraform", "tf", "hcl":
		tfExporter := converter.NewTerraformExporter()
		hcl, err := tfExporter.ExportFromResult(convResult, resourceName)
		if err != nil {
			result.Error = fmt.Errorf("terraform export failed: %w", err)
			if verbose {
				fmt.Printf("✗ %s: %v\n", inputPath, err)
			}
			return result
		}

		outputPath = filepath.Join(outputDir, resourceName+".tf")
		if err := os.WriteFile(outputPath, []byte(hcl), 0o600); err != nil {
			result.Error = fmt.Errorf("write failed: %w", err)
			if verbose {
				fmt.Printf("✗ %s: %v\n", inputPath, err)
			}
			return result
		}
	default:
		result.Error = fmt.Errorf("unknown format: %s", outputFormat)
		return result
	}

	result.OutputPath = outputPath
	result.Success = true

	if verbose {
		fmt.Printf("✓ %s → %s\n", inputPath, outputPath)
	}

	return result
}

func findInputFiles(inputPath string, recursive bool) ([]string, error) {
	var files []string

	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	if !info.IsDir() {
		if isValidInputFile(inputPath) {
			return []string{inputPath}, nil
		}
		return nil, fmt.Errorf("not a valid input file: %s", inputPath)
	}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != inputPath {
				return filepath.SkipDir
			}
			return nil
		}

		if isValidInputFile(path) {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(inputPath, walkFunc); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

func isValidInputFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".mobileconfig" || ext == ".plist" || ext == ".xml"
}

func printBatchSummary(results []batchResult) {
	var (
		successCount         int
		failureCount         int
		totalSettings        int
		totalSkippedPayloads int
		totalSkippedKeys     int
	)

	for _, r := range results {
		if r.Success {
			successCount++
			if r.Result != nil {
				totalSettings += r.Result.SettingCount
				totalSkippedPayloads += len(r.Result.SkippedPayloads)
				totalSkippedKeys += len(r.Result.SkippedKeys)
			}
		} else {
			failureCount++
		}
	}

	separator := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("Batch Conversion Summary\n")
	fmt.Printf("%s\n", separator)
	fmt.Printf("Total files:       %d\n", len(results))
	fmt.Printf("Successful:        %d\n", successCount)
	fmt.Printf("Failed:            %d\n", failureCount)
	fmt.Printf("Total settings:    %d\n", totalSettings)

	if totalSkippedPayloads > 0 {
		fmt.Printf("Skipped payloads:  %d\n", totalSkippedPayloads)
	}
	if totalSkippedKeys > 0 {
		fmt.Printf("Skipped keys:      %d\n", totalSkippedKeys)
	}

	fmt.Printf("Output directory:  %s\n", outputDir)
	fmt.Printf("%s\n", separator)

	if failureCount > 0 && verbose {
		fmt.Printf("\nFailed conversions:\n")
		for _, r := range results {
			if !r.Success {
				fmt.Printf("  ✗ %s: %v\n", r.InputPath, r.Error)
			}
		}
	}

	if failureCount > 0 && !batchContinue {
		os.Exit(1)
	}
}

func generateExportReport(results []batchResult) *converter.ExportReport {
	report := &converter.ExportReport{
		Timestamp:    time.Now(),
		TotalFiles:   len(results),
		FileReports:  make([]converter.FileReport, 0, len(results)),
	}

	for _, r := range results {
		var resourceType string
		if r.Success && r.Result != nil {
			if r.Result.UsedCustomConfig {
				resourceType = "Custom Configuration"
				report.CustomCount++
			} else {
				resourceType = "Settings Catalog"
				report.CatalogCount++
			}
			report.SuccessCount++
		} else {
			report.FailureCount++
		}

		fileReport := converter.FileReport{
			InputPath:    r.InputPath,
			OutputPath:   r.OutputPath,
			ResourceType: resourceType,
			Error:        r.Error,
		}

		if r.Result != nil {
			fileReport.SettingCount = r.Result.SettingCount
			fileReport.MatchedKeys = r.Result.MatchedKeys
			fileReport.SkippedKeys = r.Result.SkippedKeys
			fileReport.SkippedPayloads = r.Result.SkippedPayloads
			fileReport.WasSignedFile = r.Result.WasSignedFile
		}

		report.FileReports = append(report.FileReports, fileReport)
	}

	return report
}
