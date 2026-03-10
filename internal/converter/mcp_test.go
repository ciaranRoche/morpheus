package converter

import (
	"testing"
)

func TestTransformEnvVarSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "dollar-brace syntax",
			input: "Bearer ${MY_TOKEN}",
			want:  "Bearer {env:MY_TOKEN}",
		},
		{
			name:  "multiple vars",
			input: "${HOST}:${PORT}",
			want:  "{env:HOST}:{env:PORT}",
		},
		{
			name:  "no vars",
			input: "plain string",
			want:  "plain string",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "just a var",
			input: "${TOKEN}",
			want:  "{env:TOKEN}",
		},
		{
			name:  "unclosed brace left as-is",
			input: "${UNCLOSED",
			want:  "${UNCLOSED",
		},
		{
			name:  "var with underscores",
			input: "${MY_LONG_VAR_NAME}",
			want:  "{env:MY_LONG_VAR_NAME}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformEnvVarSyntax(tt.input)
			if got != tt.want {
				t.Errorf("transformEnvVarSyntax(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertMCPServer(t *testing.T) {
	tests := []struct {
		name     string
		srvName  string
		input    ClaudeMCPServer
		wantType string
		wantURL  string
		wantCmd  []string
		wantEnv  bool
	}{
		{
			name:    "http type with URL becomes remote",
			srvName: "my-server",
			input: ClaudeMCPServer{
				Type: "http",
				URL:  "https://example.com/mcp",
			},
			wantType: "remote",
			wantURL:  "https://example.com/mcp",
		},
		{
			name:    "empty type with URL becomes remote",
			srvName: "my-server",
			input: ClaudeMCPServer{
				URL: "https://example.com/mcp",
			},
			wantType: "remote",
			wantURL:  "https://example.com/mcp",
		},
		{
			name:    "empty type without URL becomes local",
			srvName: "my-server",
			input: ClaudeMCPServer{
				Type: "",
			},
			wantType: "local",
		},
		{
			name:    "stdio type becomes local",
			srvName: "my-server",
			input: ClaudeMCPServer{
				Type:    "stdio",
				Command: "npx",
				Args:    []string{"-y", "mcp-server"},
			},
			wantType: "local",
			wantCmd:  []string{"npx", "-y", "mcp-server"},
		},
		{
			name:    "preserves env",
			srvName: "my-server",
			input: ClaudeMCPServer{
				Type:    "stdio",
				Command: "node",
				Args:    []string{"server.js"},
				Env:     map[string]string{"PORT": "3000"},
			},
			wantType: "local",
			wantCmd:  []string{"node", "server.js"},
			wantEnv:  true,
		},
		{
			name:    "transforms header env vars",
			srvName: "my-server",
			input: ClaudeMCPServer{
				Type:    "http",
				URL:     "https://example.com",
				Headers: map[string]string{"Authorization": "Bearer ${TOKEN}"},
			},
			wantType: "remote",
			wantURL:  "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMCPServer(tt.srvName, tt.input)

			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}

			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}

			if !got.Enabled {
				t.Error("Enabled should be true")
			}

			if tt.wantCmd != nil {
				if len(got.Command) != len(tt.wantCmd) {
					t.Errorf("Command length = %d, want %d", len(got.Command), len(tt.wantCmd))
				} else {
					for i := range tt.wantCmd {
						if got.Command[i] != tt.wantCmd[i] {
							t.Errorf("Command[%d] = %q, want %q", i, got.Command[i], tt.wantCmd[i])
						}
					}
				}
			}

			if tt.wantEnv && got.Environment == nil {
				t.Error("expected Environment to be set")
			}

			// Check header env var transformation
			if tt.input.Headers != nil && tt.input.Headers["Authorization"] == "Bearer ${TOKEN}" {
				if got.Headers["Authorization"] != "Bearer {env:TOKEN}" {
					t.Errorf("header not transformed: got %q", got.Headers["Authorization"])
				}
			}
		})
	}
}
