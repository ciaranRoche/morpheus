package converter

import (
	"reflect"
	"testing"
)

func TestMapToolName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "Bash", input: "Bash", want: "bash"},
		{name: "Read", input: "Read", want: "read"},
		{name: "Write maps to edit", input: "Write", want: "edit"},
		{name: "Edit maps to edit", input: "Edit", want: "edit"},
		{name: "Glob", input: "Glob", want: "glob"},
		{name: "Grep", input: "Grep", want: "grep"},
		{name: "WebSearch", input: "WebSearch", want: "websearch"},
		{name: "WebFetch", input: "WebFetch", want: "webfetch"},
		{name: "Task", input: "Task", want: "task"},
		{name: "TodoRead maps to todowrite", input: "TodoRead", want: "todowrite"},
		{name: "TodoWrite", input: "TodoWrite", want: "todowrite"},
		{
			name:  "MCP tool with double underscores",
			input: "mcp__writing-samples__qdrant-find",
			want:  "writing-samples_qdrant-find",
		},
		{
			name:  "MCP tool with server and tool",
			input: "mcp__atlassian__jira_search",
			want:  "atlassian_jira_search",
		},
		{
			name:  "unknown tool lowercased",
			input: "SomeCustomTool",
			want:  "somecustomtool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapToolName(tt.input)
			if got != tt.want {
				t.Errorf("mapToolName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertToolsToPermission(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
		want  map[string]string
	}{
		{
			name:  "standard tools",
			input: []interface{}{"Bash", "Read", "WebSearch"},
			want: map[string]string{
				"bash":      "allow",
				"read":      "allow",
				"websearch": "allow",
			},
		},
		{
			name:  "MCP tool",
			input: []interface{}{"mcp__writing-samples__qdrant-find"},
			want: map[string]string{
				"writing-samples_qdrant-find": "allow",
			},
		},
		{
			name:  "mixed standard and MCP",
			input: []interface{}{"Bash", "WebFetch", "Glob", "Grep", "Read", "mcp__writing-samples__qdrant-find"},
			want: map[string]string{
				"bash":                        "allow",
				"webfetch":                    "allow",
				"glob":                        "allow",
				"grep":                        "allow",
				"read":                        "allow",
				"writing-samples_qdrant-find": "allow",
			},
		},
		{
			name:  "empty slice",
			input: []interface{}{},
			want:  map[string]string{},
		},
		{
			name:  "Write and Edit dedup to single edit key",
			input: []interface{}{"Write", "Edit"},
			want: map[string]string{
				"edit": "allow",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToolsToPermission(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToolsToPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapModelName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "haiku shorthand",
			input: "haiku",
			want:  "anthropic/claude-haiku-4-5",
		},
		{
			name:  "sonnet shorthand",
			input: "sonnet",
			want:  "anthropic/claude-sonnet-4-5",
		},
		{
			name:  "opus shorthand",
			input: "opus",
			want:  "anthropic/claude-opus-4-6",
		},
		{
			name:  "inherit returns empty",
			input: "inherit",
			want:  "",
		},
		{
			name:  "empty returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "full model ID passed through",
			input: "anthropic/claude-sonnet-4-5",
			want:  "anthropic/claude-sonnet-4-5",
		},
		{
			name:  "unknown model passed through",
			input: "gpt-4o",
			want:  "gpt-4o",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapModelName(tt.input)
			if got != tt.want {
				t.Errorf("mapModelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
