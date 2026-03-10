package cmd

import (
	"fmt"

	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed plugins and registered marketplaces",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	installed, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("could not load installed plugins: %w", err)
	}

	// Show marketplaces
	fmt.Println("Marketplaces:")
	if len(cfg.Marketplaces) == 0 {
		fmt.Println("  (none registered)")
	} else {
		for _, mp := range cfg.Marketplaces {
			fmt.Printf("  %-30s %s\n", mp.Name, mp.URL)
			fmt.Printf("  %-30s Last updated: %s\n", "", mp.LastUpdated.Format("2006-01-02 15:04"))
		}
	}

	fmt.Println()

	// Show installed plugins
	fmt.Println("Installed Plugins:")
	if len(installed.Plugins) == 0 {
		fmt.Println("  (none installed)")
	} else {
		for name, plugin := range installed.Plugins {
			version := plugin.Version
			if version == "" {
				version = "n/a"
			}
			fmt.Printf("  %-25s v%-10s from %s\n", name, version, plugin.Marketplace)
			fmt.Printf("  %-25s Installed: %s\n", "", plugin.InstalledAt.Format("2006-01-02 15:04"))

			total := len(plugin.Components.Commands) + len(plugin.Components.Skills) + len(plugin.Components.Agents)
			fmt.Printf("  %-25s Components: %d commands, %d skills, %d agents\n",
				"",
				len(plugin.Components.Commands),
				len(plugin.Components.Skills),
				len(plugin.Components.Agents))

			_ = total
			fmt.Println()
		}
	}

	return nil
}
