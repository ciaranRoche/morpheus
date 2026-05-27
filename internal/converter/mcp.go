package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ciaranRoche/morpheus/internal/config"
)

// ClaudeMCPServer represents a single MCP server from Claude Code's .mcp.json.
type ClaudeMCPServer struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// OpenCodeMCPServer represents a single MCP server in OpenCode's config format.
type OpenCodeMCPServer struct {
	Type        string            `json:"type"`
	URL         string            `json:"url,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Enabled     bool              `json:"enabled"`
	Headers     map[string]string `json:"headers,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// claudeMCPConfigEnvelope represents the top-level structure of .mcp.json
// which wraps server definitions under a "mcpServers" key.
type claudeMCPConfigEnvelope struct {
	MCPServers map[string]ClaudeMCPServer `json:"mcpServers"`
}

// ParseClaudeMCPConfig reads and parses a .mcp.json file.
// Claude Code wraps server definitions under a "mcpServers" key.
// This function handles both the wrapped format and flat format for compatibility.
func ParseClaudeMCPConfig(path string) (map[string]ClaudeMCPServer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read MCP config %s: %w", path, err)
	}

	// Try the envelope format first: {"mcpServers": {"name": {...}}}
	var envelope claudeMCPConfigEnvelope
	if err := json.Unmarshal(data, &envelope); err == nil && len(envelope.MCPServers) > 0 {
		return envelope.MCPServers, nil
	}

	// Fall back to flat format: {"name": {...}}
	var servers map[string]ClaudeMCPServer
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("could not parse MCP config: %w", err)
	}

	return servers, nil
}

// ConvertMCPServer transforms a Claude Code MCP server entry into OpenCode format.
func ConvertMCPServer(name string, claude ClaudeMCPServer) OpenCodeMCPServer {
	oc := OpenCodeMCPServer{
		Enabled: true,
	}

	switch claude.Type {
	case "http", "":
		if claude.URL != "" {
			oc.Type = "remote"
			oc.URL = claude.URL
		} else {
			oc.Type = "local"
		}
	case "stdio":
		oc.Type = "local"
	default:
		oc.Type = "local"
	}

	// Convert headers, transforming ${ENV_VAR} to {env:ENV_VAR}
	if len(claude.Headers) > 0 {
		oc.Headers = make(map[string]string)
		for k, v := range claude.Headers {
			oc.Headers[k] = transformEnvVarSyntax(v)
		}
	}

	// Convert command + args to OpenCode command array
	if claude.Command != "" {
		oc.Command = append([]string{claude.Command}, claude.Args...)
	}

	// Convert env to environment, transforming ${VAR} to {env:VAR}
	if len(claude.Env) > 0 {
		oc.Environment = make(map[string]string)
		for k, v := range claude.Env {
			oc.Environment[k] = transformEnvVarSyntax(v)
		}
	}

	return oc
}

// openCodeConfig represents the relevant parts of opencode.json for MCP merging.
// We use map[string]any to preserve all existing fields we don't touch.
type openCodeConfig map[string]any

// loadOpenCodeConfig reads and parses the opencode.json file.
func loadOpenCodeConfig() (openCodeConfig, error) {
	path, err := config.OpenCodeConfigFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return openCodeConfig{}, nil
		}
		return nil, fmt.Errorf("could not read opencode.json: %w", err)
	}

	var cfg openCodeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse opencode.json: %w", err)
	}

	return cfg, nil
}

// saveOpenCodeConfig writes the config back to opencode.json.
func saveOpenCodeConfig(cfg openCodeConfig) error {
	path, err := config.OpenCodeConfigFile()
	if err != nil {
		return err
	}

	dir, err := config.OpenCodeConfigDir()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal opencode.json: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// MergeMCPServers reads opencode.json, adds or updates the given MCP server
// entries, and writes the file back. Returns the list of server names that
// were merged.
func MergeMCPServers(servers map[string]OpenCodeMCPServer) ([]string, error) {
	cfg, err := loadOpenCodeConfig()
	if err != nil {
		return nil, err
	}

	// Get or create the mcp section
	mcpSection, ok := cfg["mcp"].(map[string]any)
	if !ok {
		mcpSection = make(map[string]any)
	}

	var names []string
	for name, server := range servers {
		mcpSection[name] = server
		names = append(names, name)
	}

	cfg["mcp"] = mcpSection

	if err := saveOpenCodeConfig(cfg); err != nil {
		return nil, err
	}

	return names, nil
}

// RemoveMCPServers removes the named MCP servers from opencode.json.
func RemoveMCPServers(names []string) error {
	if len(names) == 0 {
		return nil
	}

	cfg, err := loadOpenCodeConfig()
	if err != nil {
		return err
	}

	mcpSection, ok := cfg["mcp"].(map[string]any)
	if !ok {
		return nil // nothing to remove
	}

	for _, name := range names {
		delete(mcpSection, name)
	}

	cfg["mcp"] = mcpSection
	return saveOpenCodeConfig(cfg)
}

// transformEnvVarSyntax converts Claude Code env var syntax to OpenCode format.
// e.g., "${TOKEN}" or "$TOKEN" -> "{env:TOKEN}"
func transformEnvVarSyntax(s string) string {
	// Handle ${VAR_NAME} syntax
	result := s
	for strings.Contains(result, "${") {
		start := strings.Index(result, "${")
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		varName := result[start+2 : start+end]
		result = result[:start] + "{env:" + varName + "}" + result[start+end+1:]
	}

	return result
}
