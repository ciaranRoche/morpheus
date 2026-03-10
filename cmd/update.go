package cmd

import (
	"fmt"
	"time"

	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/ciaranRoche/morpheus/internal/git"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull latest from all registered marketplace repositories",
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	if len(cfg.Marketplaces) == 0 {
		fmt.Println("No marketplaces registered. Run 'morpheus add <repo-url>' first.")
		return nil
	}

	for i := range cfg.Marketplaces {
		mp := &cfg.Marketplaces[i]
		fmt.Printf("Updating %s...\n", mp.Name)

		if err := git.Pull(mp.LocalPath); err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		mp.LastUpdated = time.Now()
		fmt.Println("  Done.")
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}

	fmt.Println("\nAll marketplaces updated.")
	return nil
}
