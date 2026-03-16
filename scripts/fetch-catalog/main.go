package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
)

func main() {
	tenantID := flag.String("tenant", "", "Azure AD Tenant ID (required)")
	clientID := flag.String("client", "", "Azure AD Client ID (required)")
	clientSecret := flag.String("secret", "", "Azure AD Client Secret (required)")
	platform := flag.String("platform", "macOS", "Platform filter: macOS, windows10, iOS, android")
	flag.Parse()

	if *tenantID == "" || *clientID == "" || *clientSecret == "" {
		log.Fatal("Error: --tenant, --client, and --secret are required")
	}

	fmt.Println("🔐 Authenticating with Microsoft Graph...")

	cred, err := azidentity.NewClientSecretCredential(*tenantID, *clientID, *clientSecret, nil)
	if err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}

	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{
		"https://graph.microsoft.com/.default",
	})
	if err != nil {
		log.Fatalf("Failed to create Graph client: %v", err)
	}

	fmt.Println("✓ Authenticated successfully")

	ctx := context.Background()

	// Fetch settings
	fmt.Println("\n📥 Fetching Settings Catalog definitions...")
	settings, err := fetchAllSettings(ctx, client)
	if err != nil {
		log.Fatalf("Failed to fetch settings: %v", err)
	}
	fmt.Printf("✓ Fetched %d settings\n", len(settings))

	// Fetch categories
	fmt.Println("\n📥 Fetching Settings Catalog categories...")
	categories, err := fetchAllCategories(ctx, client)
	if err != nil {
		log.Fatalf("Failed to fetch categories: %v", err)
	}
	fmt.Printf("✓ Fetched %d categories\n", len(categories))

	// Filter by platform
	filteredSettings := filterByPlatform(settings, *platform)
	filteredCategories := filterCategoriesByPlatform(categories, *platform)

	fmt.Printf("\n📊 Filtered to '%s': %d settings, %d categories\n", *platform, len(filteredSettings), len(filteredCategories))

	// Save to cmd/mobileconfig-to-terraform/intune-settings-catalog-data/
	catalogDir := "cmd/mobileconfig-to-terraform/intune-settings-catalog-data"
	if err := os.MkdirAll(catalogDir, 0755); err != nil {
		log.Fatalf("Failed to create catalog directory: %v", err)
	}

	settingsPath := filepath.Join(catalogDir, fmt.Sprintf("IntuneSettingsCatalog_%s.json", *platform))
	if err := saveJSON(settingsPath, filteredSettings); err != nil {
		log.Fatalf("Failed to save settings: %v", err)
	}
	fmt.Printf("✓ Saved: %s\n", settingsPath)

	categoriesPath := filepath.Join(catalogDir, fmt.Sprintf("IntuneSettingsCategories_%s.json", *platform))
	if err := saveJSON(categoriesPath, filteredCategories); err != nil {
		log.Fatalf("Failed to save categories: %v", err)
	}
	fmt.Printf("✓ Saved: %s\n", categoriesPath)

	versionPath := filepath.Join(catalogDir, "IntuneSettingsVersion.json")
	versionData := map[string]string{"date": time.Now().Format("2006-01-02")}
	if err := saveJSON(versionPath, versionData); err != nil {
		log.Fatalf("Failed to save version: %v", err)
	}
	fmt.Printf("✓ Saved: %s\n", versionPath)

	fmt.Println("\n✅ Catalog fetch complete!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("  1. Review the updated catalog files")
	fmt.Println("  2. Rebuild the binary: make build")
	fmt.Println("  3. Commit the changes: git add cmd/intune-converter/catalog-data/ && git commit")
}

func fetchAllSettings(ctx context.Context, client *msgraphsdk.GraphServiceClient) ([]any, error) {
	settings, err := client.DeviceManagement().ConfigurationSettings().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make([]any, 0)
	if settings != nil && settings.GetValue() != nil {
		for _, setting := range settings.GetValue() {
			data, _ := json.Marshal(setting)
			var raw any
			json.Unmarshal(data, &raw)
			result = append(result, raw)
		}
	}
	return result, nil
}

func fetchAllCategories(ctx context.Context, client *msgraphsdk.GraphServiceClient) ([]any, error) {
	categories, err := client.DeviceManagement().ConfigurationCategories().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make([]any, 0)
	if categories != nil && categories.GetValue() != nil {
		for _, category := range categories.GetValue() {
			data, _ := json.Marshal(category)
			var raw any
			json.Unmarshal(data, &raw)
			result = append(result, raw)
		}
	}
	return result, nil
}

func filterByPlatform(settings []any, platform string) []any {
	filtered := make([]any, 0)
	for _, setting := range settings {
		if m, ok := setting.(map[string]any); ok {
			if app, ok := m["applicability"].(map[string]any); ok {
				if p, ok := app["platform"].(string); ok && strings.Contains(p, platform) {
					filtered = append(filtered, setting)
				}
			}
		}
	}
	return filtered
}

func filterCategoriesByPlatform(categories []any, platform string) []any {
	filtered := make([]any, 0)
	for _, category := range categories {
		if m, ok := category.(map[string]any); ok {
			if platforms, ok := m["platforms"].([]any); ok {
				for _, p := range platforms {
					if ps, ok := p.(string); ok && strings.Contains(ps, platform) {
						filtered = append(filtered, category)
						break
					}
				}
			}
		}
	}
	return filtered
}

func saveJSON(path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
