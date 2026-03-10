package converter

import (
	"testing"
)

func TestParseMarkdownWithFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFM     Frontmatter
		wantBody   string
		wantErr    bool
		checkField string
		checkValue any
	}{
		{
			name: "standard frontmatter",
			input: `---
description: A simple command
allowed-tools: Bash
---

# Hello World`,
			wantFM:     Frontmatter{"description": "A simple command", "allowed-tools": "Bash"},
			wantBody:   "# Hello World",
			checkField: "description",
			checkValue: "A simple command",
		},
		{
			name:     "no frontmatter",
			input:    "# Just a title\n\nSome body text.",
			wantFM:   nil,
			wantBody: "# Just a title\n\nSome body text.",
		},
		{
			name: "frontmatter with unquoted colon in value",
			input: `---
description: Weekly update - nested display : activity type -> Epic
allowed-tools: Bash
argument-hint: [project-key]
---

# Weekly Update`,
			wantBody:   "# Weekly Update",
			checkField: "description",
			checkValue: "Weekly update - nested display : activity type -> Epic",
		},
		{
			name:    "unclosed frontmatter",
			input:   "---\ndescription: broken\nno closing delimiter",
			wantErr: true,
		},
		{
			name: "empty frontmatter",
			input: `---
---

Body only.`,
			wantFM:   Frontmatter{},
			wantBody: "Body only.",
		},
		{
			name:     "empty input",
			input:    "",
			wantFM:   nil,
			wantBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := ParseMarkdownWithFrontmatter(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMarkdownWithFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}

			if tt.checkField != "" {
				val, ok := fm[tt.checkField]
				if !ok {
					t.Errorf("frontmatter missing field %q", tt.checkField)
				} else if val != tt.checkValue {
					t.Errorf("fm[%q] = %v, want %v", tt.checkField, val, tt.checkValue)
				}
			}
		})
	}
}

func TestParseFrontmatterLenient(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantKeys []string
	}{
		{
			name:     "simple key-value pairs",
			input:    "description: hello world\nallowed-tools: Bash",
			wantKeys: []string{"description", "allowed-tools"},
		},
		{
			name:     "value with colon",
			input:    "description: nested display : activity type\ntools: Bash",
			wantKeys: []string{"description", "tools"},
		},
		{
			name:    "empty input",
			input:   "",
			wantNil: true,
		},
		{
			name:    "only comments",
			input:   "# comment\n# another",
			wantNil: true,
		},
		{
			name:     "skips empty lines and comments",
			input:    "key1: val1\n\n# comment\nkey2: val2",
			wantKeys: []string{"key1", "key2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := parseFrontmatterLenient(tt.input)

			if tt.wantNil {
				if fm != nil {
					t.Errorf("expected nil, got %v", fm)
				}
				return
			}

			if fm == nil {
				t.Fatal("expected non-nil Frontmatter")
			}

			for _, key := range tt.wantKeys {
				if _, ok := fm[key]; !ok {
					t.Errorf("missing expected key %q in frontmatter", key)
				}
			}
		})
	}
}

func TestRenderMarkdownWithFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		fm      Frontmatter
		body    string
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "standard render",
			fm:   Frontmatter{"description": "test"},
			body: "# Title",
			check: func(s string) bool {
				return len(s) > 0 &&
					contains(s, "---") &&
					contains(s, "description: test") &&
					contains(s, "# Title")
			},
		},
		{
			name: "nil frontmatter returns body only",
			fm:   nil,
			body: "just body",
			check: func(s string) bool {
				return s == "just body\n"
			},
		},
		{
			name: "empty frontmatter returns body only",
			fm:   Frontmatter{},
			body: "just body",
			check: func(s string) bool {
				return s == "just body\n"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderMarkdownWithFrontmatter(tt.fm, tt.body)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil && !tt.check(result) {
				t.Errorf("output check failed, got:\n%s", result)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
