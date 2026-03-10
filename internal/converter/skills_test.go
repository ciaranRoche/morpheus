package converter

import (
	"testing"
)

func TestNormalizeSkillName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already kebab-case",
			input: "jira-ticket-creator",
			want:  "jira-ticket-creator",
		},
		{
			name:  "uppercase with spaces",
			input: "JIRA Ticket Creator",
			want:  "jira-ticket-creator",
		},
		{
			name:  "mixed case with underscores",
			input: "My_Cool_Skill",
			want:  "my-cool-skill",
		},
		{
			name:  "special characters",
			input: "skill@v2.0!",
			want:  "skill-v2-0",
		},
		{
			name:  "multiple consecutive separators",
			input: "foo---bar___baz",
			want:  "foo-bar-baz",
		},
		{
			name:  "leading and trailing separators",
			input: "--foo-bar--",
			want:  "foo-bar",
		},
		{
			name:  "single word lowercase",
			input: "linter",
			want:  "linter",
		},
		{
			name:  "single word uppercase",
			input: "LINTER",
			want:  "linter",
		},
		{
			name:  "numbers mixed in",
			input: "v2-jira-3-tools",
			want:  "v2-jira-3-tools",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "jira-story-pointer to estimator-like normalization",
			input: "jira-story-pointer",
			want:  "jira-story-pointer",
		},
		{
			name:  "jira-triage",
			input: "jira-triage",
			want:  "jira-triage",
		},
		{
			name:  "camelCase",
			input: "mySkillName",
			want:  "myskillname",
		},
		{
			name:  "dots and slashes",
			input: "org.example/skill-name",
			want:  "org-example-skill-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSkillName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeSkillName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple",
			input:   "my-skill",
			wantErr: false,
		},
		{
			name:    "valid single word",
			input:   "linter",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "v2-jira-tool",
			wantErr: false,
		},
		{
			name:    "valid long name",
			input:   "this-is-a-reasonably-long-skill-name-that-fits",
			wantErr: false,
		},
		{
			name:    "invalid empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid uppercase",
			input:   "MySkill",
			wantErr: true,
		},
		{
			name:    "invalid leading hyphen",
			input:   "-my-skill",
			wantErr: true,
		},
		{
			name:    "invalid trailing hyphen",
			input:   "my-skill-",
			wantErr: true,
		},
		{
			name:    "invalid double hyphen",
			input:   "my--skill",
			wantErr: true,
		},
		{
			name:    "invalid underscore",
			input:   "my_skill",
			wantErr: true,
		},
		{
			name:    "invalid spaces",
			input:   "my skill",
			wantErr: true,
		},
		{
			name:    "too long (65 chars)",
			input:   "a-b-c-d-e-f-g-h-i-j-k-l-m-n-o-p-q-r-s-t-u-v-w-x-y-z-aa-bb-cc-dddd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkillName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSkillName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
