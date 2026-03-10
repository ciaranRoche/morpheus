package cmd

import (
	"fmt"
	"os"

	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/ciaranRoche/morpheus/internal/converter"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Health check on OpenCode config and installed plugins",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("morpheus doctor - checking your setup...")

	issues := 0

	// Check morpheus config directory
	morpheusDir, err := config.MorpheusConfigDir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(morpheusDir); os.IsNotExist(err) {
		fmt.Printf("  [WARN] morpheus config directory does not exist: %s\n", morpheusDir)
		issues++
	} else {
		fmt.Printf("  [ OK ] morpheus config directory: %s\n", morpheusDir)
	}

	// Check OpenCode config directory
	ocDir, err := config.OpenCodeConfigDir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(ocDir); os.IsNotExist(err) {
		fmt.Printf("  [WARN] OpenCode config directory does not exist: %s\n", ocDir)
		fmt.Printf("         Run 'mkdir -p %s' to create it\n", ocDir)
		issues++
	} else {
		fmt.Printf("  [ OK ] OpenCode config directory: %s\n", ocDir)
	}

	// Check subdirectories
	for _, dir := range []struct {
		name   string
		getter func() (string, error)
	}{
		{"commands", config.OpenCodeCommandsDir},
		{"skills", config.OpenCodeSkillsDir},
		{"agents", config.OpenCodeAgentsDir},
	} {
		path, err := dir.getter()
		if err != nil {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("  [INFO] OpenCode %s directory does not exist (will be created on install)\n", dir.name)
		} else {
			fmt.Printf("  [ OK ] OpenCode %s directory: %s\n", dir.name, path)
		}
	}

	// Check registered marketplaces
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("  [FAIL] Could not load morpheus config: %v\n", err)
		issues++
	} else {
		fmt.Printf("  [ OK ] Registered marketplaces: %d\n", len(cfg.Marketplaces))
		for _, mp := range cfg.Marketplaces {
			if _, err := os.Stat(mp.LocalPath); os.IsNotExist(err) {
				fmt.Printf("  [FAIL] Marketplace %q directory missing: %s\n", mp.Name, mp.LocalPath)
				issues++
			} else {
				fmt.Printf("  [ OK ] Marketplace %q: %s\n", mp.Name, mp.LocalPath)
			}
		}
	}

	// Check installed plugins
	installed, err := config.LoadInstalled()
	if err != nil {
		fmt.Printf("  [FAIL] Could not load installed plugins: %v\n", err)
		issues++
	} else {
		fmt.Printf("  [ OK ] Installed plugins: %d\n", len(installed.Plugins))

		// Validate each installed plugin's components exist
		for name, plugin := range installed.Plugins {
			for _, skillName := range plugin.Components.Skills {
				if err := converter.ValidateSkillName(skillName); err != nil {
					fmt.Printf("  [WARN] Plugin %q: skill %q has invalid name: %v\n", name, skillName, err)
					issues++
				}
			}
		}
	}

	fmt.Println()
	if issues == 0 {
		fmt.Println("All checks passed. You're good to go!")
	} else {
		fmt.Printf("Found %d issue(s). See above for details.\n", issues)
	}

	return nil
}
