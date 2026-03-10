package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "morpheus",
	Short: "Free your Claude Code plugins from the Matrix",
	Long: `morpheus - Claude Code Plugin to OpenCode Migration Tool

"I can only show you the door. You're the one that has to walk through it."

morpheus clones Claude Code plugin marketplace repos, converts their 
commands, skills, and agents into OpenCode format, and installs them
into your global OpenCode configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// No subcommand -- launch the interactive TUI
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(doctorCmd)
}
