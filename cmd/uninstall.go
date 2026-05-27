package cmd

import (
	"fmt"

	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/ciaranRoche/morpheus/internal/converter"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <plugin-name>",
	Short: "Remove an installed plugin from OpenCode",
	Long:  `Removes all components of an installed plugin from the OpenCode configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	installed, err := config.LoadInstalled()
	if err != nil {
		return fmt.Errorf("could not load installed plugins: %w", err)
	}

	plugin, exists := installed.Plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %q is not installed", pluginName)
	}

	fmt.Printf("Uninstalling %q...\n\n", pluginName)

	results := converter.UninstallPlugin(
		plugin.Components.Commands,
		plugin.Components.Skills,
		plugin.Components.Agents,
		plugin.Components.MCP,
	)

	for _, r := range results {
		icon := "OK"
		if !r.Success {
			icon = "FAIL"
		}
		fmt.Printf("  [%4s] %-8s %s\n", icon, r.ComponentType, r.Name)
		if r.Error != nil {
			fmt.Printf("         %v\n", r.Error)
		}
	}

	// Remove from installed registry
	delete(installed.Plugins, pluginName)
	if err := config.SaveInstalled(installed); err != nil {
		return fmt.Errorf("could not save installed plugins: %w", err)
	}

	fmt.Printf("\nPlugin %q uninstalled.\n", pluginName)
	return nil
}
