package converter

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents parsed YAML frontmatter from a markdown file.
type Frontmatter map[string]any

// ParseMarkdownWithFrontmatter splits a markdown file into frontmatter and body.
// Returns the parsed frontmatter, the raw body, and any error.
func ParseMarkdownWithFrontmatter(content string) (Frontmatter, string, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		// No frontmatter
		return nil, content, nil
	}

	// Find the closing ---
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return nil, content, fmt.Errorf("unclosed frontmatter delimiter")
	}

	fmRaw := strings.TrimSpace(rest[:idx])
	body := strings.TrimSpace(rest[idx+4:]) // skip past \n---

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		// Fallback: try lenient line-by-line parsing for frontmatter with
		// unquoted values containing colons (common in plugin descriptions).
		fm = parseFrontmatterLenient(fmRaw)
		if fm == nil {
			return nil, "", fmt.Errorf("failed to parse frontmatter YAML: %w", err)
		}
	}

	return fm, body, nil
}

// parseFrontmatterLenient parses simple key: value frontmatter line-by-line.
// This handles cases where values contain unquoted colons that break strict YAML.
// It only supports flat string key-value pairs (no nested structures).
func parseFrontmatterLenient(raw string) Frontmatter {
	fm := make(Frontmatter)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key == "" {
			return nil // not simple key-value, give up
		}
		fm[key] = value
	}
	if len(fm) == 0 {
		return nil
	}
	return fm
}

// RenderMarkdownWithFrontmatter combines frontmatter and body back into a markdown file.
func RenderMarkdownWithFrontmatter(fm Frontmatter, body string) (string, error) {
	if fm == nil || len(fm) == 0 {
		return body + "\n", nil
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(body)
	sb.WriteString("\n")

	return sb.String(), nil
}
