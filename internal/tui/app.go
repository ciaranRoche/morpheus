package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ciaranRoche/morpheus/internal/config"
	"github.com/ciaranRoche/morpheus/internal/converter"
	"github.com/ciaranRoche/morpheus/internal/marketplace"
)

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewInstalling
	viewMessage
)

type model struct {
	plugins      []marketplace.Plugin
	installed    *config.InstalledPlugins
	cursor       int
	view         viewState
	width        int
	height       int
	message      string
	messageIsErr bool
	filterText   string
	marketplaces []config.MarketplaceEntry
}

type installDoneMsg struct {
	pluginName string
	results    []converter.Result
	err        error
}

func NewModel() (model, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return model{}, fmt.Errorf("could not load config: %w", err)
	}

	installed, err := config.LoadInstalled()
	if err != nil {
		return model{}, fmt.Errorf("could not load installed plugins: %w", err)
	}

	// Discover all plugins across all marketplaces
	var allPlugins []marketplace.Plugin
	for _, mp := range cfg.Marketplaces {
		plugins, err := marketplace.DiscoverPlugins(mp.LocalPath, mp.Name)
		if err != nil {
			continue
		}
		// Mark installed plugins
		for i := range plugins {
			if _, ok := installed.Plugins[plugins[i].Name]; ok {
				plugins[i].Installed = true
			}
		}
		allPlugins = append(allPlugins, plugins...)
	}

	return model{
		plugins:      allPlugins,
		installed:    installed,
		marketplaces: cfg.Marketplaces,
		view:         viewList,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case installDoneMsg:
		m.view = viewMessage
		if msg.err != nil {
			m.message = fmt.Sprintf("Error installing %s: %v", msg.pluginName, msg.err)
			m.messageIsErr = true
		} else {
			var cmds, skills, agents int
			for _, r := range msg.results {
				if r.Success {
					switch r.ComponentType {
					case marketplace.ComponentCommand:
						cmds++
					case marketplace.ComponentSkill:
						skills++
					case marketplace.ComponentAgent:
						agents++
					}
				}
			}
			m.message = fmt.Sprintf("Installed %s: %d commands, %d skills, %d agents",
				msg.pluginName, cmds, skills, agents)
			m.messageIsErr = false
			// Refresh installed status
			for i := range m.plugins {
				if m.plugins[i].Name == msg.pluginName {
					m.plugins[i].Installed = true
				}
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.view {
	case viewMessage:
		// Any key returns to list
		m.view = viewList
		m.message = ""
		return m, nil

	case viewDetail:
		switch key {
		case "q", "esc", "backspace":
			m.view = viewList
		case "enter", "i":
			return m.installSelected()
		}
		return m, nil

	case viewInstalling:
		return m, nil

	case viewList:
		filtered := m.filteredPlugins()
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}
		case "enter":
			if len(filtered) > 0 {
				m.view = viewDetail
			}
		case "i":
			if len(filtered) > 0 {
				return m.installSelected()
			}
		case "backspace":
			if len(m.filterText) > 0 {
				m.filterText = m.filterText[:len(m.filterText)-1]
				m.cursor = 0
			}
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.filterText += key
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m model) installSelected() (tea.Model, tea.Cmd) {
	filtered := m.filteredPlugins()
	if m.cursor >= len(filtered) {
		return m, nil
	}

	plugin := filtered[m.cursor]
	m.view = viewInstalling
	m.message = fmt.Sprintf("Installing %s...", plugin.Name)

	return m, func() tea.Msg {
		results := converter.ConvertPlugin(&plugin)

		// Save installation record
		installed, err := config.LoadInstalled()
		if err != nil {
			return installDoneMsg{pluginName: plugin.Name, err: err}
		}

		var commands, skills, agents []string
		for _, r := range results {
			if r.Success {
				switch r.ComponentType {
				case marketplace.ComponentCommand:
					commands = append(commands, r.Name)
				case marketplace.ComponentSkill:
					skills = append(skills, r.Name)
				case marketplace.ComponentAgent:
					agents = append(agents, r.Name)
				}
			}
		}

		installed.Plugins[plugin.Name] = config.InstalledPlugin{
			Marketplace: plugin.Marketplace,
			Version:     plugin.Version,
			InstalledAt: time.Now(),
			Components: config.InstalledComponent{
				Commands: commands,
				Skills:   skills,
				Agents:   agents,
			},
		}

		if err := config.SaveInstalled(installed); err != nil {
			return installDoneMsg{pluginName: plugin.Name, err: err}
		}

		return installDoneMsg{pluginName: plugin.Name, results: results}
	}
}

func (m model) filteredPlugins() []marketplace.Plugin {
	if m.filterText == "" {
		return m.plugins
	}
	filter := strings.ToLower(m.filterText)
	var result []marketplace.Plugin
	for _, p := range m.plugins {
		if strings.Contains(strings.ToLower(p.Name), filter) ||
			strings.Contains(strings.ToLower(p.Description), filter) {
			result = append(result, p)
		}
	}
	return result
}

func (m model) View() tea.View {
	var s strings.Builder

	switch m.view {
	case viewList:
		s.WriteString(m.renderList())
	case viewDetail:
		s.WriteString(m.renderDetail())
	case viewInstalling:
		s.WriteString(m.renderInstalling())
	case viewMessage:
		s.WriteString(m.renderMessage())
	}

	return tea.NewView(s.String())
}

func (m model) renderList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("morpheus"))
	s.WriteString(subtitleStyle.Render("  \"I can only show you the door.\""))
	s.WriteString("\n\n")

	// Filter bar
	filterDisplay := m.filterText
	if filterDisplay == "" {
		filterDisplay = dimStyle.Render("type to filter...")
	}
	s.WriteString(fmt.Sprintf("  Filter: %s\n\n", filterDisplay))

	filtered := m.filteredPlugins()

	if len(m.marketplaces) == 0 {
		s.WriteString(dimStyle.Render("  No marketplaces registered.\n"))
		s.WriteString(dimStyle.Render("  Run 'morpheus add <repo-url>' to get started.\n"))
	} else if len(filtered) == 0 {
		s.WriteString(dimStyle.Render("  No plugins found.\n"))
	}

	// Group by marketplace
	groups := make(map[string][]int)
	for i, p := range filtered {
		groups[p.Marketplace] = append(groups[p.Marketplace], i)
	}

	globalIdx := 0
	for mpName, indices := range groups {
		s.WriteString(headerStyle.Render(fmt.Sprintf("  %s", mpName)))
		s.WriteString("\n")

		for _, idx := range indices {
			p := filtered[idx]
			cursor := "  "
			style := normalStyle
			if globalIdx == m.cursor {
				cursor = "> "
				style = selectedStyle
			}

			name := style.Render(p.Name)
			version := ""
			if p.Version != "" {
				version = versionStyle.Render(fmt.Sprintf(" v%s", p.Version))
			}
			badge := ""
			if p.Installed {
				badge = installedBadge.Render(" [installed]")
			}

			s.WriteString(fmt.Sprintf("  %s%s%s%s\n", cursor, name, version, badge))
			globalIdx++
		}
		s.WriteString("\n")
	}

	s.WriteString(helpStyle.Render("  [enter] details  [i] install  [q] quit"))
	s.WriteString("\n")

	return s.String()
}

func (m model) renderDetail() string {
	var s strings.Builder
	filtered := m.filteredPlugins()
	if m.cursor >= len(filtered) {
		return "No plugin selected"
	}

	p := filtered[m.cursor]

	s.WriteString(titleStyle.Render(fmt.Sprintf("  %s", p.Name)))
	s.WriteString("\n")

	if p.Version != "" {
		s.WriteString(fmt.Sprintf("  Version: %s\n", versionStyle.Render(p.Version)))
	}
	if p.Author.Name != "" {
		s.WriteString(fmt.Sprintf("  Author:  %s\n", p.Author.Name))
	}
	s.WriteString(fmt.Sprintf("  Source:  %s\n", dimStyle.Render(p.Marketplace)))

	if p.Installed {
		s.WriteString(fmt.Sprintf("  Status:  %s\n", installedBadge.Render("installed")))
	} else {
		s.WriteString(fmt.Sprintf("  Status:  %s\n", dimStyle.Render("not installed")))
	}

	s.WriteString("\n")

	if p.Description != "" {
		desc := p.Description
		if len(desc) > 200 {
			desc = desc[:197] + "..."
		}
		s.WriteString(fmt.Sprintf("  %s\n\n", desc))
	}

	// Components
	s.WriteString(componentStyle.Render("  Components:"))
	s.WriteString("\n")

	if cmds := p.Commands(); len(cmds) > 0 {
		names := make([]string, len(cmds))
		for i, c := range cmds {
			names[i] = "/" + c.Name
		}
		s.WriteString(fmt.Sprintf("    Commands: %s\n", strings.Join(names, ", ")))
	}

	if skills := p.Skills(); len(skills) > 0 {
		names := make([]string, len(skills))
		for i, sk := range skills {
			names[i] = sk.Name
		}
		s.WriteString(fmt.Sprintf("    Skills:   %s\n", strings.Join(names, ", ")))
	}

	if agents := p.Agents(); len(agents) > 0 {
		names := make([]string, len(agents))
		for i, a := range agents {
			names[i] = a.Name
		}
		s.WriteString(fmt.Sprintf("    Agents:   %s\n", strings.Join(names, ", ")))
	}

	if hooks := p.Hooks(); len(hooks) > 0 {
		s.WriteString(fmt.Sprintf("    Hooks:    %s\n", dimStyle.Render("(manual conversion required)")))
	}

	if mcp := p.MCPServers(); len(mcp) > 0 {
		s.WriteString(fmt.Sprintf("    MCP:      %s\n", dimStyle.Render("(manual config merge needed)")))
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  [i/enter] install  [esc] back  [q] quit"))
	s.WriteString("\n")

	return s.String()
}

func (m model) renderInstalling() string {
	return fmt.Sprintf("\n  %s\n", m.message)
}

func (m model) renderMessage() string {
	var s strings.Builder
	s.WriteString("\n")
	if m.messageIsErr {
		s.WriteString(errorStyle.Render(fmt.Sprintf("  %s", m.message)))
	} else {
		s.WriteString(successStyle.Render(fmt.Sprintf("  %s", m.message)))
	}
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("  Press any key to continue..."))
	s.WriteString("\n")
	return s.String()
}

// Run starts the Bubble Tea TUI.
func Run() error {
	m, err := NewModel()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
