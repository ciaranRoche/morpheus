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

	// Handle MCP servers (warn about manual config merge)
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

		for name := range servers {
			results = append(results, Result{
				ComponentType: marketplace.ComponentMCP,
				Name:          name,
				OriginalName:  name,
				Success:       true,
				Warning:       fmt.Sprintf("MCP server %q parsed -- add to opencode.json manually or run 'morpheus doctor'", name),
			})
		}
	}

	return results
}

// UninstallPlugin removes all tracked components of a plugin from OpenCode.
func UninstallPlugin(commands, skills, agents []string) []Result {
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

	return results
}
