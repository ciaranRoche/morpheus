package config

import (
	"os"
	"path/filepath"
)

const (
	appName = "morpheus"
)

// MorpheusConfigDir returns the path to morpheus's own configuration directory.
// Defaults to ~/.config/morpheus/
func MorpheusConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, appName), nil
}

// MarketplacesDir returns the path where marketplace repos are cloned.
func MarketplacesDir() (string, error) {
	dir, err := MorpheusConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "marketplaces"), nil
}

// OpenCodeConfigDir returns the path to the OpenCode global config directory.
// Defaults to ~/.config/opencode/
func OpenCodeConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "opencode"), nil
}

// OpenCodeCommandsDir returns the path where OpenCode commands live.
func OpenCodeCommandsDir() (string, error) {
	dir, err := OpenCodeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "commands"), nil
}

// OpenCodeSkillsDir returns the path where OpenCode skills live.
func OpenCodeSkillsDir() (string, error) {
	dir, err := OpenCodeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "skills"), nil
}

// OpenCodeAgentsDir returns the path where OpenCode agents live.
func OpenCodeAgentsDir() (string, error) {
	dir, err := OpenCodeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "agents"), nil
}

// OpenCodePluginsDir returns the path where OpenCode plugins live.
func OpenCodePluginsDir() (string, error) {
	dir, err := OpenCodeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "plugins"), nil
}

// OpenCodeConfigFile returns the path to the OpenCode config file.
func OpenCodeConfigFile() (string, error) {
	dir, err := OpenCodeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "opencode.json"), nil
}

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
