package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/deploymenttheory/go-settings-catalog/internal/cli"
)

//go:embed intune-settings-catalog-data/*.json
var intuneSettingsCatalogFS embed.FS

func main() {
	if err := cli.Execute(intuneSettingsCatalogFS); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
