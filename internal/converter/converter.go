package converter

import (
	"fmt"

	"github.com/ciaranRoche/morpheus/internal/marketplace"
)

// Result captures the outcome of converting a single component.
type Result struct {
	ComponentType marketplace.ComponentType
	Name          string
	OriginalName  string // For skills where the name may change
	Success       bool
	Error         error
	Warning       string
}

// ConvertPlugin converts all components of a plugin into OpenCode format.
// Returns a list of results for each component.
func ConvertPlugin(plugin *marketplace.Plugin) []Result {
	var results []Result

	// Convert commands
	for _, cmd := range plugin.Commands() {
		err := ConvertCommand(cmd.Path, cmd.Name)
		r := Result{
			ComponentType: marketplace.ComponentCommand,
			Name:          cmd.Name,
			OriginalName:  cmd.Name,
			Success:       err == nil,
			Error:         err,
		}
		results = append(results, r)
	}

	// Convert skills
	for _, skill := range plugin.Skills() {
		normalizedName, err := ConvertSkill(skill.Path, skill.Name)
		r := Result{
			ComponentType: marketplace.ComponentSkill,
			Name:          normalizedName,
			OriginalName:  skill.Name,
			Success:       err == nil,
			Error:         err,
		}
		if normalizedName == "" {
			r.Name = skill.Name
		}
		results = append(results, r)
	}

	// Convert agents
	for _, agent := range plugin.Agents() {
		err := ConvertAgent(agent.Path, agent.Name)
		r := Result{
			ComponentType: marketplace.ComponentAgent,
			Name:          agent.Name,
			OriginalName:  agent.Name,
			Success:       err == nil,
			Error:         err,
		}
		results = append(results, r)
	}

	// Handle hooks (warn, don't convert)
	for _, hook := range plugin.Hooks() {
		results = append(results, Result{
			ComponentType: marketplace.ComponentHook,
			Name:          hook.Name,
			OriginalName:  hook.Name,
			Success:       false,
			Warning:       "hooks require manual conversion to OpenCode plugins",
		})
	}

	// Convert and merge MCP servers into opencode.json
	for _, mcp := range plugin.MCPServers() {
		servers, err := ParseClaudeMCPConfig(mcp.Path)
		if err != nil {
			results = append(results, Result{
				ComponentType: marketplace.ComponentMCP,
				Name:          mcp.Name,
				OriginalName:  mcp.Name,
				Success:       false,
				Error:         fmt.Errorf("could not parse MCP config: %w", err),
			})
			continue
		}

		// Convert each server to OpenCode format
		converted := make(map[string]OpenCodeMCPServer)
		for name, srv := range servers {
			converted[name] = ConvertMCPServer(name, srv)
		}

		// Merge into opencode.json
		names, err := MergeMCPServers(converted)
		if err != nil {
			for name := range converted {
				results = append(results, Result{
					ComponentType: marketplace.ComponentMCP,
					Name:          name,
					OriginalName:  name,
					Success:       false,
					Error:         fmt.Errorf("could not merge MCP server into opencode.json: %w", err),
				})
			}
			continue
		}

		for _, name := range names {
			results = append(results, Result{
				ComponentType: marketplace.ComponentMCP,
				Name:          name,
				OriginalName:  name,
				Success:       true,
			})
		}
	}

	return results
}

// UninstallPlugin removes all tracked components of a plugin from OpenCode.
func UninstallPlugin(commands, skills, agents, mcpServers []string) []Result {
	var results []Result

	for _, name := range commands {
		err := RemoveCommand(name)
		results = append(results, Result{
			ComponentType: marketplace.ComponentCommand,
			Name:          name,
			Success:       err == nil,
			Error:         err,
		})
	}

	for _, name := range skills {
		err := RemoveSkill(name)
		results = append(results, Result{
			ComponentType: marketplace.ComponentSkill,
			Name:          name,
			Success:       err == nil,
			Error:         err,
		})
	}

	for _, name := range agents {
		err := RemoveAgent(name)
		results = append(results, Result{
			ComponentType: marketplace.ComponentAgent,
			Name:          name,
			Success:       err == nil,
			Error:         err,
		})
	}

	if len(mcpServers) > 0 {
		err := RemoveMCPServers(mcpServers)
		if err != nil {
			for _, name := range mcpServers {
				results = append(results, Result{
					ComponentType: marketplace.ComponentMCP,
					Name:          name,
					Success:       false,
					Error:         err,
				})
			}
		} else {
			for _, name := range mcpServers {
				results = append(results, Result{
					ComponentType: marketplace.ComponentMCP,
					Name:          name,
					Success:       true,
				})
			}
		}
	}

	return results
}
