package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/ciaranRoche/morpheus/internal/config"
)

var validSkillName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// NormalizeSkillName converts a Claude Code skill name to OpenCode kebab-case format.
// e.g., "JIRA Ticket Creator" -> "jira-ticket-creator"
func NormalizeSkillName(name string) string {
	// Convert to lowercase
	lower := strings.ToLower(name)

	// Replace any non-alphanumeric character with a hyphen
	var result strings.Builder
	prevHyphen := false
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			result.WriteRune('-')
			prevHyphen = true
		}
	}

	// Trim leading/trailing hyphens
	normalized := strings.Trim(result.String(), "-")

	// Collapse consecutive hyphens
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}

	return normalized
}

// ValidateSkillName checks if a skill name matches OpenCode's requirements.
func ValidateSkillName(name string) error {
	if len(name) == 0 || len(name) > 64 {
		return fmt.Errorf("skill name must be 1-64 characters, got %d", len(name))
	}
	if !validSkillName.MatchString(name) {
		return fmt.Errorf("skill name %q must be lowercase alphanumeric with single hyphen separators", name)
	}
	return nil
}

// ConvertSkill transforms a Claude Code SKILL.md into OpenCode format
// and writes it to the OpenCode skills directory.
func ConvertSkill(sourcePath, skillDirName string) (string, error) {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("could not read skill file %s: %w", sourcePath, err)
	}

	fm, body, err := ParseMarkdownWithFrontmatter(string(content))
	if err != nil {
		return "", fmt.Errorf("could not parse skill %s: %w", skillDirName, err)
	}

	if fm == nil {
		fm = make(Frontmatter)
	}

	// Normalize the skill name to kebab-case
	originalName, _ := fm["name"].(string)
	normalizedName := NormalizeSkillName(originalName)
	if normalizedName == "" {
		// Fall back to directory name
		normalizedName = NormalizeSkillName(skillDirName)
	}

	// Validate the normalized name
	if err := ValidateSkillName(normalizedName); err != nil {
		return "", fmt.Errorf("could not normalize skill name %q: %w", originalName, err)
	}

	// Update the name field
	fm["name"] = normalizedName

	// Ensure description exists and is within length limits
	if desc, ok := fm["description"].(string); ok {
		if len(desc) > 1024 {
			fm["description"] = desc[:1024]
		}
	}

	// Remove any fields not recognized by OpenCode
	// OpenCode supports: name, description, license, compatibility, metadata
	allowedFields := map[string]bool{
		"name": true, "description": true, "license": true,
		"compatibility": true, "metadata": true,
	}
	for key := range fm {
		if !allowedFields[key] {
			delete(fm, key)
		}
	}

	// Render the transformed skill
	output, err := RenderMarkdownWithFrontmatter(fm, body)
	if err != nil {
		return "", fmt.Errorf("could not render skill %s: %w", normalizedName, err)
	}

	// Write to OpenCode skills directory using the normalized name
	destDir, err := config.OpenCodeSkillsDir()
	if err != nil {
		return "", err
	}
	skillDir := filepath.Join(destDir, normalizedName)
	if err := config.EnsureDir(skillDir); err != nil {
		return "", err
	}

	destPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(destPath, []byte(output), 0o644); err != nil {
		return "", fmt.Errorf("could not write skill to %s: %w", destPath, err)
	}

	return normalizedName, nil
}

// RemoveSkill removes a skill directory from the OpenCode skills directory.
func RemoveSkill(skillName string) error {
	destDir, err := config.OpenCodeSkillsDir()
	if err != nil {
		return err
	}

	skillDir := filepath.Join(destDir, skillName)
	if err := os.RemoveAll(skillDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove skill %s: %w", skillDir, err)
	}
	return nil
}
