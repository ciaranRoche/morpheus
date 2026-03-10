package converter

import (
	"testing"
)

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
