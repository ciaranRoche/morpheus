package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	configFileName    = "config.json"
	installedFileName = "installed.json"
)

// MarketplaceEntry represents a registered marketplace in morpheus config.
type MarketplaceEntry struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	LocalPath   string    `json:"localPath"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Config is morpheus's own configuration.
type Config struct {
	Marketplaces []MarketplaceEntry `json:"marketplaces"`
}

// InstalledComponent tracks what was installed for a plugin.
type InstalledComponent struct {
	Commands []string `json:"commands,omitempty"`
	Skills   []string `json:"skills,omitempty"`
	Agents   []string `json:"agents,omitempty"`
	MCP      []string `json:"mcp,omitempty"`
}

// InstalledPlugin tracks an installed plugin.
type InstalledPlugin struct {
	Marketplace string             `json:"marketplace"`
	Version     string             `json:"version"`
	InstalledAt time.Time          `json:"installedAt"`
	Components  InstalledComponent `json:"components"`
}

// InstalledPlugins tracks all installed plugins.
type InstalledPlugins struct {
	Plugins map[string]InstalledPlugin `json:"plugins"`
}

// LoadConfig loads morpheus's configuration from disk.
func LoadConfig() (*Config, error) {
	dir, err := MorpheusConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, configFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes morpheus's configuration to disk.
func SaveConfig(cfg *Config) error {
	dir, err := MorpheusConfigDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, configFileName), data, 0o644)
}

// FindMarketplace looks up a marketplace entry by name.
func (c *Config) FindMarketplace(name string) *MarketplaceEntry {
	for i := range c.Marketplaces {
		if c.Marketplaces[i].Name == name {
			return &c.Marketplaces[i]
		}
	}
	return nil
}

// AddMarketplace adds or updates a marketplace entry.
func (c *Config) AddMarketplace(entry MarketplaceEntry) {
	for i := range c.Marketplaces {
		if c.Marketplaces[i].Name == entry.Name {
			c.Marketplaces[i] = entry
			return
		}
	}
	c.Marketplaces = append(c.Marketplaces, entry)
}

// RemoveMarketplace removes a marketplace by name.
func (c *Config) RemoveMarketplace(name string) {
	filtered := c.Marketplaces[:0]
	for _, m := range c.Marketplaces {
		if m.Name != name {
			filtered = append(filtered, m)
		}
	}
	c.Marketplaces = filtered
}

// LoadInstalled loads the installed plugins registry.
func LoadInstalled() (*InstalledPlugins, error) {
	dir, err := MorpheusConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, installedFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &InstalledPlugins{Plugins: make(map[string]InstalledPlugin)}, nil
		}
		return nil, err
	}

	var inst InstalledPlugins
	if err := json.Unmarshal(data, &inst); err != nil {
		return nil, err
	}
	if inst.Plugins == nil {
		inst.Plugins = make(map[string]InstalledPlugin)
	}
	return &inst, nil
}

// SaveInstalled writes the installed plugins registry to disk.
func SaveInstalled(inst *InstalledPlugins) error {
	dir, err := MorpheusConfigDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, installedFileName), data, 0o644)
}
