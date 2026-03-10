package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Clone clones a git repository to the specified destination.
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s\n%s", err, string(output))
	}
	return nil
}

// Pull runs git pull in the specified repository directory.
func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s\n%s", err, string(output))
	}
	return nil
}

// RepoNameFromURL extracts a repository name from a git URL.
// e.g., "git@github.com:org/repo-name.git" -> "repo-name"
// e.g., "https://github.com/org/repo-name.git" -> "repo-name"
func RepoNameFromURL(url string) string {
	// Handle SSH URLs: git@github.com:org/repo.git
	if strings.Contains(url, ":") && strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, "/")
		name := parts[len(parts)-1]
		return strings.TrimSuffix(name, ".git")
	}

	// Handle HTTPS URLs: https://github.com/org/repo.git
	name := filepath.Base(url)
	return strings.TrimSuffix(name, ".git")
}

// IsGitRepo checks if a directory is a git repository.
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}
