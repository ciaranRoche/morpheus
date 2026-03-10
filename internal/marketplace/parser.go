package marketplace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParseMarketplace reads and parses a .claude-plugin/marketplace.json file
// from a marketplace repository root.
func ParseMarketplace(repoPath string) (*MarketplaceManifest, error) {
	manifestPath := filepath.Join(repoPath, ".claude-plugin", "marketplace.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("could not read marketplace manifest at %s: %w", manifestPath, err)
	}

	var manifest MarketplaceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("could not parse marketplace manifest: %w", err)
	}

	return &manifest, nil
}

// ParsePluginMetadata reads and parses a .claude-plugin/plugin.json file
// from a plugin directory.
func ParsePluginMetadata(pluginDir string) (*PluginMetadata, error) {
	metaPath := filepath.Join(pluginDir, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("could not read plugin metadata at %s: %w", metaPath, err)
	}

	var meta PluginMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("could not parse plugin metadata: %w", err)
	}

	return &meta, nil
}

// resolvePluginSource resolves the source field from a marketplace plugin entry
// to a local directory path. Currently only supports local relative paths.
func resolvePluginSource(repoPath string, source any) (string, error) {
	switch s := source.(type) {
	case string:
		// Local relative path like "./hyperfleet-jira"
		clean := strings.TrimPrefix(s, "./")
		return filepath.Join(repoPath, clean), nil
	case map[string]any:
		// External source -- not supported yet
		return "", fmt.Errorf("external plugin sources are not yet supported")
	default:
		return "", fmt.Errorf("unsupported source type: %T", source)
	}
}

// DiscoverPlugins loads all plugins from a marketplace repo, including their components.
func DiscoverPlugins(repoPath, marketplaceName string) ([]Plugin, error) {
	manifest, err := ParseMarketplace(repoPath)
	if err != nil {
		return nil, err
	}

	var plugins []Plugin
	for _, mp := range manifest.Plugins {
		pluginDir, err := resolvePluginSource(repoPath, mp.Source)
		if err != nil {
			// Skip plugins we can't resolve (e.g., external)
			continue
		}

		// Check if the plugin directory actually exists
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			continue
		}

		plugin := Plugin{
			Name:        mp.Name,
			Description: mp.Description,
			Version:     mp.Version,
			SourcePath:  pluginDir,
			Marketplace: marketplaceName,
		}

		// Try to get more metadata from plugin.json
		meta, err := ParsePluginMetadata(pluginDir)
		if err == nil {
			if plugin.Version == "" {
				plugin.Version = meta.Version
			}
			if plugin.Description == "" {
				plugin.Description = meta.Description
			}
			plugin.Author = meta.Author
		}

		// Discover components
		plugin.Components = discoverComponents(pluginDir)
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// discoverComponents scans a plugin directory for all known component types.
func discoverComponents(pluginDir string) []PluginComponent {
	var components []PluginComponent

	// Discover commands
	commandsDir := filepath.Join(pluginDir, "commands")
	if entries, err := os.ReadDir(commandsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				name := strings.TrimSuffix(e.Name(), ".md")
				components = append(components, PluginComponent{
					Type:     ComponentCommand,
					Name:     name,
					Path:     filepath.Join(commandsDir, e.Name()),
					FileName: e.Name(),
				})
			}
		}
	}

	// Discover skills
	skillsDir := filepath.Join(pluginDir, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				skillFile := filepath.Join(skillsDir, e.Name(), "SKILL.md")
				if _, err := os.Stat(skillFile); err == nil {
					components = append(components, PluginComponent{
						Type:     ComponentSkill,
						Name:     e.Name(),
						Path:     skillFile,
						FileName: "SKILL.md",
					})
				}
			}
		}
	}

	// Discover agents
	agentsDir := filepath.Join(pluginDir, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				name := strings.TrimSuffix(e.Name(), ".md")
				components = append(components, PluginComponent{
					Type:     ComponentAgent,
					Name:     name,
					Path:     filepath.Join(agentsDir, e.Name()),
					FileName: e.Name(),
				})
			}
		}
	}

	// Discover hooks
	hooksFile := filepath.Join(pluginDir, "hooks", "hooks.json")
	if _, err := os.Stat(hooksFile); err == nil {
		components = append(components, PluginComponent{
			Type:     ComponentHook,
			Name:     "hooks",
			Path:     hooksFile,
			FileName: "hooks.json",
		})
	}

	// Discover MCP servers
	mcpFile := filepath.Join(pluginDir, ".mcp.json")
	if _, err := os.Stat(mcpFile); err == nil {
		components = append(components, PluginComponent{
			Type:     ComponentMCP,
			Name:     "mcp",
			Path:     mcpFile,
			FileName: ".mcp.json",
		})
	}

	// Discover scripts
	scriptsDir := filepath.Join(pluginDir, "scripts")
	if entries, err := os.ReadDir(scriptsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				components = append(components, PluginComponent{
					Type:     ComponentScript,
					Name:     e.Name(),
					Path:     filepath.Join(scriptsDir, e.Name()),
					FileName: e.Name(),
				})
			}
		}
	}

	return components
}
