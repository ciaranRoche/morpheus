package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// ParseClaudeMCPConfig reads and parses a .mcp.json file.
func ParseClaudeMCPConfig(path string) (map[string]ClaudeMCPServer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read MCP config %s: %w", path, err)
	}

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

	// Convert env to environment
	if len(claude.Env) > 0 {
		oc.Environment = claude.Env
	}

	return oc
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
