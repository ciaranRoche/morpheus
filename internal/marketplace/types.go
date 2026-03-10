package marketplace

// MarketplaceManifest represents the .claude-plugin/marketplace.json file.
type MarketplaceManifest struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Owner       MarketplaceOwner    `json:"owner,omitempty"`
	Plugins     []MarketplacePlugin `json:"plugins"`
}

// MarketplaceOwner is the owner of the marketplace.
type MarketplaceOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// MarketplacePlugin is a plugin entry in the marketplace manifest.
type MarketplacePlugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version,omitempty"`
	Source      any    `json:"source"` // Can be string (local path) or object (external)
	Category    string `json:"category,omitempty"`
}

// PluginMetadata represents .claude-plugin/plugin.json within a plugin directory.
type PluginMetadata struct {
	Name        string           `json:"name"`
	Version     string           `json:"version,omitempty"`
	Description string           `json:"description"`
	Author      MarketplaceOwner `json:"author,omitempty"`
}

// ComponentType represents the type of plugin component.
type ComponentType string

const (
	ComponentCommand ComponentType = "command"
	ComponentSkill   ComponentType = "skill"
	ComponentAgent   ComponentType = "agent"
	ComponentHook    ComponentType = "hook"
	ComponentMCP     ComponentType = "mcp"
	ComponentScript  ComponentType = "script"
)

// PluginComponent represents a discovered component within a plugin.
type PluginComponent struct {
	Type     ComponentType
	Name     string
	Path     string // Full path to the component file/directory
	FileName string // Just the filename
}

// Plugin represents a fully parsed plugin with all its components.
type Plugin struct {
	Name        string
	Description string
	Version     string
	Author      MarketplaceOwner
	SourcePath  string // Path to the plugin directory on disk
	Marketplace string // Name of the marketplace it belongs to
	Components  []PluginComponent
	Installed   bool
}

// Commands returns all command components.
func (p *Plugin) Commands() []PluginComponent {
	return p.filterComponents(ComponentCommand)
}

// Skills returns all skill components.
func (p *Plugin) Skills() []PluginComponent {
	return p.filterComponents(ComponentSkill)
}

// Agents returns all agent components.
func (p *Plugin) Agents() []PluginComponent {
	return p.filterComponents(ComponentAgent)
}

// Hooks returns all hook components.
func (p *Plugin) Hooks() []PluginComponent {
	return p.filterComponents(ComponentHook)
}

// MCPServers returns all MCP components.
func (p *Plugin) MCPServers() []PluginComponent {
	return p.filterComponents(ComponentMCP)
}

func (p *Plugin) filterComponents(t ComponentType) []PluginComponent {
	var result []PluginComponent
	for _, c := range p.Components {
		if c.Type == t {
			result = append(result, c)
		}
	}
	return result
}
