package converter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	if _, ok := fm["color"]; ok {
		log.Printf("agent %s: stripped unsupported 'color' field", agentName)
		delete(fm, "color")
	}

	// Convert Claude Code tools array to OpenCode permission map
	// Claude Code: tools: [Bash, Read, WebSearch]
	// OpenCode:    permission: {bash: allow, read: allow, websearch: allow}
	if tools, ok := fm["tools"]; ok {
		if toolsSlice, ok := tools.([]interface{}); ok {
			fm["permission"] = convertToolsToPermission(toolsSlice)
			log.Printf("agent %s: converted tools array to permission map", agentName)
		} else {
			log.Printf("agent %s: stripped unrecognized 'tools' format (expected array)", agentName)
		}
		delete(fm, "tools")
	}

	// Remove "name" -- OpenCode derives agent name from the filename
	if _, ok := fm["name"]; ok {
		log.Printf("agent %s: stripped unsupported 'name' field (OpenCode derives name from filename)", agentName)
		delete(fm, "name")
	}

	// Default to subagent mode if not already set
	// Claude Code agents map to OpenCode subagents
	if _, ok := fm["mode"]; !ok {
		fm["mode"] = "subagent"
	}

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

// mapToolName converts a single Claude Code tool name to its OpenCode permission key.
// Claude Code uses PascalCase names (Bash, Read, WebSearch) while OpenCode uses
// lowercase permission keys (bash, read, websearch).
// MCP tools use double underscores in Claude Code (mcp__server__tool) and single
// underscores in OpenCode (server_tool).
func mapToolName(tool string) string {
	// Known Claude Code -> OpenCode permission key mappings
	known := map[string]string{
		"Bash":      "bash",
		"Read":      "read",
		"Write":     "edit",
		"Edit":      "edit",
		"Glob":      "glob",
		"Grep":      "grep",
		"WebSearch": "websearch",
		"WebFetch":  "webfetch",
		"Task":      "task",
		"TodoRead":  "todowrite",
		"TodoWrite": "todowrite",
	}

	if key, ok := known[tool]; ok {
		return key
	}

	// MCP tools: mcp__server-name__tool-name -> server-name_tool-name
	if strings.HasPrefix(tool, "mcp__") {
		stripped := strings.TrimPrefix(tool, "mcp__")
		return strings.Replace(stripped, "__", "_", 1)
	}

	// Unknown tools: lowercase as-is
	return strings.ToLower(tool)
}

// convertToolsToPermission converts a Claude Code tools array into an OpenCode
// permission map. Each tool in the array is mapped to its OpenCode permission key
// with the value "allow".
func convertToolsToPermission(tools []interface{}) map[string]string {
	perm := make(map[string]string)
	for _, t := range tools {
		if name, ok := t.(string); ok {
			key := mapToolName(name)
			perm[key] = "allow"
		}
	}
	return perm
}
