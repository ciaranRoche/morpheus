package converter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ciaranRoche/morpheus/internal/config"
)

// ConvertCommand transforms a Claude Code command markdown file into OpenCode format
// and writes it to the OpenCode commands directory.
func ConvertCommand(sourcePath, commandName string) error {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("could not read command file %s: %w", sourcePath, err)
	}

	fm, body, err := ParseMarkdownWithFrontmatter(string(content))
	if err != nil {
		return fmt.Errorf("could not parse command %s: %w", commandName, err)
	}

	// Transform frontmatter for OpenCode compatibility
	if fm == nil {
		fm = make(Frontmatter)
	}

	// Remove Claude Code specific fields
	delete(fm, "allowed-tools")

	// Merge argument-hint into description if present
	if hint, ok := fm["argument-hint"]; ok {
		if desc, ok := fm["description"].(string); ok {
			fm["description"] = fmt.Sprintf("%s %v", desc, hint)
		}
		delete(fm, "argument-hint")
	}

	// Render the transformed command
	output, err := RenderMarkdownWithFrontmatter(fm, body)
	if err != nil {
		return fmt.Errorf("could not render command %s: %w", commandName, err)
	}

	// Write to OpenCode commands directory
	destDir, err := config.OpenCodeCommandsDir()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(destDir); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, commandName+".md")
	if err := os.WriteFile(destPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("could not write command to %s: %w", destPath, err)
	}

	return nil
}

// RemoveCommand removes a command file from the OpenCode commands directory.
func RemoveCommand(commandName string) error {
	destDir, err := config.OpenCodeCommandsDir()
	if err != nil {
		return err
	}

	destPath := filepath.Join(destDir, commandName+".md")
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove command %s: %w", destPath, err)
	}
	return nil
}
