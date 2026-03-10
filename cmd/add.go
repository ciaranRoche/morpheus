package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/ciaranRoche/morpheus/internal/git"
	"github.com/ciaranRoche/morpheus/internal/marketplace"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <repo-url>",
	Short: "Add a Claude Code plugin marketplace repository",
	Long:  `Clones a Claude Code plugin marketplace repository and registers it for use with morpheus.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	repoURL := args[0]
	repoName := git.RepoNameFromURL(repoURL)

	fmt.Printf("Adding marketplace %q from %s...\n", repoName, repoURL)

	// Determine clone destination
	marketplacesDir, err := config.MarketplacesDir()
	if err != nil {
		return fmt.Errorf("could not determine marketplaces directory: %w", err)
	}
	if err := config.EnsureDir(marketplacesDir); err != nil {
		return fmt.Errorf("could not create marketplaces directory: %w", err)
	}

	destPath := filepath.Join(marketplacesDir, repoName)

	// Clone or pull
	if git.IsGitRepo(destPath) {
		fmt.Println("  Repository already exists, pulling latest...")
		if err := git.Pull(destPath); err != nil {
			return fmt.Errorf("could not pull: %w", err)
		}
	} else {
		fmt.Println("  Cloning repository...")
		if err := git.Clone(repoURL, destPath); err != nil {
			return fmt.Errorf("could not clone: %w", err)
		}
	}

	// Validate it's a valid marketplace
	manifest, err := marketplace.ParseMarketplace(destPath)
	if err != nil {
		// Clean up on failure
		os.RemoveAll(destPath)
		return fmt.Errorf("not a valid Claude Code marketplace: %w", err)
	}

	// Register in config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	cfg.AddMarketplace(config.MarketplaceEntry{
		Name:        repoName,
		URL:         repoURL,
		LocalPath:   destPath,
		LastUpdated: time.Now(),
	})

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}

	// Show available plugins
	fmt.Printf("\nMarketplace %q registered successfully!\n", manifest.Name)
	fmt.Printf("Available plugins (%d):\n\n", len(manifest.Plugins))

	for _, p := range manifest.Plugins {
		version := p.Version
		if version == "" {
			version = "n/a"
		}
		fmt.Printf("  - %-30s %s\n", p.Name, version)
		if p.Description != "" {
			desc := p.Description
			if len(desc) > 70 {
				desc = desc[:67] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
	}

	fmt.Printf("\nRun 'morpheus install <plugin-name>' to convert and install a plugin.\n")
	fmt.Printf("Or run 'morpheus' for the interactive TUI.\n")

	return nil
}
