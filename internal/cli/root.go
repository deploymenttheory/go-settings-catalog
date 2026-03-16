package cli

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	outputDir string
	verbose   bool
	catalogFS embed.FS
)

var rootCmd = &cobra.Command{
	Use:   "mobileconfig-to-terraform",
	Short: "Convert macOS mobileconfig files to Terraform HCL for Intune Settings Catalog",
	Long: `Convert macOS .mobileconfig files to Terraform HCL for Microsoft Intune Settings Catalog.

Commands:
  convert - Convert a single configuration file
  batch   - Batch convert multiple files or directories

Supported inputs:
  - .mobileconfig files (signed or unsigned)
  - .plist files
  - .xml preference files

Supported outputs:
  - Terraform HCL (microsoft365_graph_beta_device_management_settings_catalog_configuration_policy)
  - Settings Catalog JSON (Microsoft Graph API format)

The Microsoft Intune Settings Catalog data for macOS is embedded in the binary.`,
	Version: "0.0.1",
}

func Execute(fs embed.FS) error {
	catalogFS = fs
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mobileconfig-to-terraform.yaml)")
	rootCmd.PersistentFlags().
		StringVarP(&outputDir, "output", "o", ".", "output directory for generated files")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mobileconfig-to-terraform")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
