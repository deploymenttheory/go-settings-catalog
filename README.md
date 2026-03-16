# macOS Configuration Profile to Intune Settings Catalog Converter

[![Release](https://img.shields.io/github/v/release/deploymenttheory/go-settings-catalog)](https://github.com/deploymenttheory/go-settings-catalog/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deploymenttheory/go-settings-catalog)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/deploymenttheory/go-settings-catalog)](https://goreportcard.com/report/github.com/deploymenttheory/go-settings-catalog)
[![License](https://img.shields.io/github/license/deploymenttheory/go-settings-catalog)](LICENSE)
![Status: Beta](https://img.shields.io/badge/status-beta-yellow)

Convert macOS configuration profiles (`.mobileconfig`) to Microsoft Intune Settings Catalog format for Terraform or Microsoft Graph API.

## Overview

This CLI tool converts macOS configuration profiles into Microsoft Intune Settings Catalog policies, enabling infrastructure-as-code workflows for device management. It intelligently maps mobileconfig settings to Intune's Settings Catalog using exact and fuzzy matching, with automatic fallback to Custom Configuration for unsupported settings.

## Key Features

- **Dual Output Formats**
  - Terraform HCL for `microsoft365_graph_beta_device_management_settings_catalog_configuration_policy` resource
  - Settings Catalog JSON for direct Microsoft Graph API usage

- **Intelligent Matching**
  - Exact key matching for standard settings
  - Fuzzy matching using Levenshtein distance (75% similarity threshold)
  - Automatic fallback to Custom Configuration for unmatched settings

- **Batch Processing**
  - Convert multiple files or entire directories
  - Parallel processing with configurable workers
  - Recursive directory traversal
  - Continue-on-error mode

- **Automatic Handling**
  - CMS/PKCS#7 signature removal from signed profiles
  - Terraform variable escaping (`$` → `$$` in HEREDOC)
  - Integer bounds validation against catalog definitions
  - Type coercion for unsigned integers (uint64, uint32, uint)

- **Reporting**
  - Markdown conversion reports
  - Match type tracking (exact, fuzzy, skipped)
  - Nearest match suggestions for unmatched keys
  - Similarity scores for fuzzy matches

- **Embedded SettingsCatalog Lookup**
  - No external API calls required
  - Catalog data embedded in binary
  - Automated updates via GitHub Actions pipeline

## Installation

Download the latest binary for your platform from the [Releases](https://github.com/deploymenttheory/go-settings-catalog/releases) page.

Binaries are automatically built and published via release-please when new versions are tagged.

## Usage

### Convert Command

Convert a single configuration file to Intune Settings Catalog format.

```bash
# Convert to Terraform HCL (default)
mobileconfig-to-terraform convert profile.mobileconfig

# Convert to Settings Catalog JSON
mobileconfig-to-terraform convert profile.mobileconfig -f json

# Specify output directory and resource name
mobileconfig-to-terraform convert profile.mobileconfig -o ./output -n my_policy
```

**Options:**
- `-f, --format` - Output format: `terraform` (default) or `json`
- `-n, --name` - Resource name (default: derived from filename)
- `-o, --output` - Output directory (default: current directory)

**Supported Input Formats:**
- `.mobileconfig` - macOS configuration profiles (signed or unsigned)
- `.plist` - Property list files
- `.xml` - XML preference files

### Batch Command

Convert multiple configuration files or directories with parallel processing.

```bash
# Convert all files in a directory
mobileconfig-to-terraform batch ./profiles

# Convert recursively through subdirectories
mobileconfig-to-terraform batch ./profiles -r

# Convert to JSON format
mobileconfig-to-terraform batch ./profiles -f json -o ./output

# Continue on errors (don't stop on first failure)
mobileconfig-to-terraform batch ./profiles --continue

# Use 8 parallel workers for faster processing
mobileconfig-to-terraform batch ./profiles -w 8
```

**Options:**
- `-f, --format` - Output format: `terraform` (default) or `json`
- `-r, --recursive` - Recursively process subdirectories
- `-w, --workers` - Number of parallel workers (default: 4)
- `--continue` - Continue processing on errors
- `-o, --output` - Output directory (default: current directory)

**Batch Output:**
- Converted files (`.tf` or `.json`)
- `conversion-report.md` - Detailed conversion report with match statistics

## Output Formats

### Terraform HCL

**Settings Catalog:**
```hcl
resource "microsoft365_graph_beta_device_management_settings_catalog_configuration_policy" "dock_settings" {
  name         = "Dock Configuration"
  platforms    = "macOS"
  technologies = ["mdm"]
  
  configuration_policy = {
    settings = [
      # Converted settings with proper HCL block structure
    ]
  }
}
```

**Custom Configuration:**
```hcl
resource "microsoft365_graph_beta_device_management_macos_device_configuration_templates" "custom_profile" {
  display_name = "Custom Profile"
  
  custom_configuration = {
    deployment_channel = "deviceChannel"
    payload_file_name  = "profile.mobileconfig"
    payload_name       = "Custom Profile"
    payload            = <<-EOT
<?xml version="1.0" encoding="UTF-8"?>
<!-- Original mobileconfig XML -->
    EOT
  }
}
```

### Settings Catalog JSON

**Settings Catalog:**
```json
{
  "name": "Dock Configuration",
  "description": "",
  "platforms": "macOS",
  "technologies": "mdm",
  "settings": [
    {
      "settingInstance": {
        "@odata.type": "#microsoft.graph.deviceManagementConfigurationGroupSettingCollectionInstance",
        "settingDefinitionId": "...",
        "groupSettingCollectionValue": [...]
      }
    }
  ]
}
```

**Custom Configuration:**
```json
{
  "@odata.type": "#microsoft.graph.macOSCustomConfiguration",
  "displayName": "Custom Profile",
  "payloadName": "Custom Profile",
  "payloadFileName": "profile.mobileconfig",
  "payload": "PD94bWwgdm...",
  "deploymentChannel": "deviceChannel"
}
```
## Supported Settings

The tool supports all macOS settings in Microsoft's Intune Settings Catalog, including:

- **System Preferences** - Dock, Finder, Login Items, Accessibility
- **Security & Privacy** - FileVault, Firewall, Gatekeeper, Privacy Preferences
- **Network** - Wi-Fi, VPN, Proxy, Network Extensions
- **Applications** - Managed app preferences (Chrome, Safari, etc.)
- **Device Management** - Software Update, System Extensions, Kernel Extensions

Settings not in the catalog automatically fall back to macOS custom configuration profile template format.

## Requirements

- macOS, Linux, or Windows
- No external dependencies (catalog data embedded in binary)
- For Terraform output: [microsoft365 Terraform provider](https://registry.terraform.io/providers/deploymenttheory/microsoft365)

## Catalog Updates

The embedded Intune Settings Catalog data is automatically updated via GitHub Actions pipeline. Updates are triggered when Microsoft releases new catalog versions and are published as new releases.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development

```bash
# Clone repository
git clone https://github.com/deploymenttheory/go-settings-catalog.git
cd go-settings-catalog

# Run tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# Build
go build -o bin/mobileconfig-to-terraform ./cmd/mobileconfig-to-terraform
```

See [TESTING.md](TESTING.md) for detailed test coverage information.

## License

[MIT License](LICENSE)

## Support

For questions, issues, or feature requests:
- Open an issue on [GitHub](https://github.com/deploymenttheory/go-settings-catalog/issues)
- Join our [Discord community](https://discord.gg/Uq8zG6g7WE)
