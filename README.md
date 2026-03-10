# morpheus

> *"I can only show you the door. You're the one that has to walk through it."*

A CLI tool that migrates [Claude Code](https://docs.anthropic.com/en/docs/claude-code) plugins from marketplace Git repos into [OpenCode](https://opencode.ai) format.

One-way. No going back. Welcome to the real world.

## What it does

morpheus clones Claude Code plugin marketplace repositories, parses their plugin manifests, and converts each component into the equivalent OpenCode configuration:

| Claude Code | OpenCode | Transformation |
|---|---|---|
| `commands/*.md` | `~/.config/opencode/commands/` | Removes `allowed-tools`, appends `argument-hint` to description |
| `skills/*/SKILL.md` | `~/.config/opencode/skills/<name>/SKILL.md` | Normalizes name to lowercase kebab-case, strips unsupported fields |
| `agents/*.md` | `~/.config/opencode/agents/` | Maps model shorthand (`haiku`/`sonnet`/`opus`) to full IDs, removes `color` |
| `.mcp.json` | `opencode.json` `mcp` section | `type: http` becomes `type: remote`, `${VAR}` becomes `{env:VAR}` |
| `hooks/hooks.json` | -- | No direct equivalent (warns user) |

## Install

```sh
go install github.com/ciaranRoche/morpheus@latest
```

Or build from source:

```sh
git clone https://github.com/ciaranRoche/morpheus.git
cd morpheus
make build
# Binary at ./bin/morpheus
```

## Usage

### Interactive TUI

Run `morpheus` with no arguments to launch the Bubble Tea TUI:

```sh
morpheus
```

Browse available plugins, view details, and install/uninstall with keyboard shortcuts.

### CLI Commands

```sh
# Register a marketplace repo
morpheus add git@github.com:org/claude-plugins.git

# List available and installed plugins
morpheus status

# Install a plugin (converts and writes to OpenCode config)
morpheus install <plugin-name>

# Uninstall a plugin (removes converted files)
morpheus uninstall <plugin-name>

# Pull latest from all registered marketplaces
morpheus update

# Health check on your setup
morpheus doctor
```

### Example workflow

```sh
# Add your team's marketplace
$ morpheus add git@github.com:openshift-hyperfleet/hyperfleet-claude-plugins.git
Marketplace "hyperfleet-claude-plugins" registered successfully!
Available plugins (5):
  - hyperfleet-jira
  - hyperfleet-standards
  - hyperfleet-architecture
  - hyperfleet-operational-readiness
  - hyperfleet-devtools

# Install the JIRA plugin
$ morpheus install hyperfleet-jira
  [  OK] command  my-sprint
  [  OK] command  my-tasks
  [  OK] command  sprint-status
  [  OK] command  team-weekly-update
  [  OK] command  triage
  [  OK] skill    jira-story-point-estimator
  [  OK] skill    jira-ticket-creator
  [  OK] skill    jira-ticket-triage

Plugin "hyperfleet-jira" installed successfully!

# Verify everything is in order
$ morpheus doctor
  [ OK ] morpheus config directory: ~/.config/morpheus
  [ OK ] OpenCode config directory: ~/.config/opencode
  [ OK ] OpenCode commands directory: ~/.config/opencode/commands
  [ OK ] OpenCode skills directory: ~/.config/opencode/skills
  [ OK ] Registered marketplaces: 1
  [ OK ] Installed plugins: 1
All checks passed. You're good to go!
```

## How it works

1. **`morpheus add <repo-url>`** clones the marketplace repo to `~/.config/morpheus/marketplaces/` and parses its `.claude-plugin/marketplace.json` manifest
2. **`morpheus install <plugin>`** reads the plugin's components, transforms each one for OpenCode compatibility, and writes them to `~/.config/opencode/`
3. **`morpheus uninstall <plugin>`** removes the converted files from the OpenCode config directory
4. **`morpheus update`** runs `git pull` on all registered marketplace repos to pick up new plugins or versions

### Skill name normalization

OpenCode requires skill names to be lowercase kebab-case matching the regex `^[a-z0-9]+(-[a-z0-9]+)*$` (1-64 chars). morpheus automatically normalizes names:

- `JIRA Ticket Creator` -> `jira-ticket-creator`
- `My_Cool_Skill` -> `my-cool-skill`
- `jira-story-pointer` -> `jira-story-pointer` (already valid)

### Limitations

- **Hooks are not converted.** Claude Code hooks (`hooks/hooks.json`) have no direct OpenCode equivalent. morpheus warns you and skips them.
- **One-way only.** There is no OpenCode-to-Claude-Code migration.
- **Marketplace repos only.** morpheus reads from Git repositories, not from `~/.claude/plugins/cache/`.

## Development

```sh
make build    # Build to ./bin/morpheus
make test     # Run tests
make fmt      # go fmt
make vet      # go vet
make clean    # Remove build artifacts
make install  # Install to $GOPATH/bin
```

## License

MIT
