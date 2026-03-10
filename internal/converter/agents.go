package converter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ciaranRoche/morpheus/internal/config"
)

// ConvertAgent transforms a Claude Code agent markdown file into OpenCode format
// and writes it to the OpenCode agents directory.
func ConvertAgent(sourcePath, agentName string) error {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("could not read agent file %s: %w", sourcePath, err)
	}

	fm, body, err := ParseMarkdownWithFrontmatter(string(content))
	if err != nil {
		return fmt.Errorf("could not parse agent %s: %w", agentName, err)
	}

	if fm == nil {
		fm = make(Frontmatter)
	}

	// Transform Claude Code agent fields to OpenCode format
	// Claude Code: model values like "inherit", "haiku", "sonnet", "opus"
	// OpenCode: full model IDs like "anthropic/claude-haiku-4-5"
	if model, ok := fm["model"].(string); ok {
		fm["model"] = mapModelName(model)
	}

	// Claude Code uses "color" for display -- OpenCode doesn't support this, remove it
	delete(fm, "color")

	// Render the transformed agent
	output, err := RenderMarkdownWithFrontmatter(fm, body)
	if err != nil {
		return fmt.Errorf("could not render agent %s: %w", agentName, err)
	}

	// Write to OpenCode agents directory
	destDir, err := config.OpenCodeAgentsDir()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(destDir); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, agentName+".md")
	if err := os.WriteFile(destPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("could not write agent to %s: %w", destPath, err)
	}

	return nil
}

// RemoveAgent removes an agent file from the OpenCode agents directory.
func RemoveAgent(agentName string) error {
	destDir, err := config.OpenCodeAgentsDir()
	if err != nil {
		return err
	}

	destPath := filepath.Join(destDir, agentName+".md")
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove agent %s: %w", destPath, err)
	}
	return nil
}

// mapModelName converts Claude Code model shorthand to OpenCode model IDs.
func mapModelName(model string) string {
	switch model {
	case "haiku":
		return "anthropic/claude-haiku-4-5"
	case "sonnet":
		return "anthropic/claude-sonnet-4-5"
	case "opus":
		return "anthropic/claude-opus-4-6"
	case "inherit", "":
		return "" // Let OpenCode use default
	default:
		return model // Pass through if already a full model ID
	}
}
