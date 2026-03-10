package cmd

import (
	"fmt"
	"time"

	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/ciaranRoche/morpheus/internal/converter"
	"github.com/ciaranRoche/morpheus/internal/marketplace"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <plugin-name>",
	Short: "Convert and install a Claude Code plugin into OpenCode",
	Long:  `Converts a Claude Code plugin's commands, skills, and agents into OpenCode format and installs them.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	// Load config and find the plugin across all marketplaces
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	if len(cfg.Marketplaces) == 0 {
		return fmt.Errorf("no marketplaces registered -- run 'morpheus add <repo-url>' first")
	}

	// Search all marketplaces for the plugin
	var foundPlugin *marketplace.Plugin
	for _, mp := range cfg.Marketplaces {
		plugins, err := marketplace.DiscoverPlugins(mp.LocalPath, mp.Name)
		if err != nil {
			continue
		}
		for i := range plugins {
			if plugins[i].Name == pluginName {
				foundPlugin = &plugins[i]
				break
			}
		}
		if foundPlugin != nil {
			break
		}
	}

	if foundPlugin == nil {
		return fmt.Errorf("plugin %q not found in any registered marketplace", pluginName)
	}

	fmt.Printf("Installing %q from %s...\n\n", foundPlugin.Name, foundPlugin.Marketplace)

	// Convert all components
	results := converter.ConvertPlugin(foundPlugin)

	// Display results
	var commands, skills, agents []string
	hasErrors := false

	for _, r := range results {
		icon := "OK"
		if !r.Success && r.Warning == "" {
			icon = "FAIL"
			hasErrors = true
		} else if r.Warning != "" {
			icon = "WARN"
		}

		fmt.Printf("  [%4s] %-8s %s", icon, r.ComponentType, r.Name)
		if r.OriginalName != r.Name && r.OriginalName != "" {
			fmt.Printf(" (was: %s)", r.OriginalName)
		}
		fmt.Println()

		if r.Error != nil {
			fmt.Printf("         %v\n", r.Error)
		}
		if r.Warning != "" {
			fmt.Printf("         %s\n", r.Warning)
		}

		// Track installed components
		if r.Success {
			switch r.ComponentType {
			case marketplace.ComponentCommand:
				commands = append(commands, r.Name)
			case marketplace.ComponentSkill:
				skills = append(skills, r.Name)
			case marketplace.ComponentAgent:
				agents = append(agents, r.Name)
			}
		}
	}

	// Save installation record
	installed, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("could not load installed plugins: %w", err)
	}

	installed.Plugins[pluginName] = config.InstalledPlugin{
		Marketplace: foundPlugin.Marketplace,
		Version:     foundPlugin.Version,
		InstalledAt: time.Now(),
		Components: config.InstalledComponent{
			Commands: commands,
			Skills:   skills,
			Agents:   agents,
		},
	}

	if err := config.SaveInstalled(installed); err != nil {
		return fmt.Errorf("could not save installed plugins: %w", err)
	}

	fmt.Println()
	if hasErrors {
		fmt.Println("Installation completed with errors. Some components could not be converted.")
	} else {
		fmt.Printf("Plugin %q installed successfully!\n", pluginName)
	}

	fmt.Printf("\nComponents installed: %d commands, %d skills, %d agents\n",
		len(commands), len(skills), len(agents))

	return nil
}
