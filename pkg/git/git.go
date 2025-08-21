package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github-issue-manager/pkg/logger"
)

// ParseGitRemoteURL parses a GitHub URL (SSH or HTTPS) and returns the owner and repo.
// For SSH: git@github.com:owner/repo.git
// For HTTPS: https://github.com/owner/repo.git
func ParseGitRemoteURL(url string) (owner, repo string, err error) {
	// Handle SSH format
	if strings.HasPrefix(url, "git@github.com:") {
		// Remove the prefix
		url = strings.TrimPrefix(url, "git@github.com:")
		// Remove the .git suffix if present
		url = strings.TrimSuffix(url, ".git")
		// Split by "/"
		parts := strings.Split(url, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH URL format: %s", url)
		}
		return parts[0], parts[1], nil
	}

	// Handle HTTPS format
	if strings.HasPrefix(url, "https://github.com/") {
		// Remove the prefix
		url = strings.TrimPrefix(url, "https://github.com/")
		// Remove the .git suffix if present
		url = strings.TrimSuffix(url, ".git")
		// Split by "/"
		parts := strings.Split(url, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid HTTPS URL format: %s", url)
		}
		return parts[0], parts[1], nil
	}

	// If neither format matches, return an error
	return "", "", fmt.Errorf("unsupported URL format: %s", url)
}

// InferOwnerRepoFromGit attempts to infer the GitHub owner and repository name
// from the local .git configuration.
func InferOwnerRepoFromGit() (inferredOwner, inferredRepo string) {
	// Check if .git directory exists
	_, err := os.Stat(".git")
	if os.IsNotExist(err) {
		logger.Info("No .git directory found, cannot infer owner/repo")
		return "", ""
	}
	if err != nil {
		logger.Error("Error checking for .git directory", "error", err)
		return "", ""
	}

	// Read .git/config file
	data, err := ioutil.ReadFile(".git/config")
	if err != nil {
		logger.Error("Failed to read .git/config", "error", err)
		return "", ""
	}

	// Parse the config file to find the remote origin URL
	configStr := string(data)
	lines := strings.Split(configStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "[remote \"origin\"]") {
			// Look for the URL in the next lines
			for j := i + 1; j < len(lines); j++ {
				urlLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(urlLine, "url = ") {
					url := strings.TrimPrefix(urlLine, "url = ")
					// Use ParseGitRemoteURL to extract owner and repo
					owner, repo, err := ParseGitRemoteURL(url)
					if err != nil {
						logger.Error("Failed to parse remote URL", "error", err)
						return "", ""
					}
					return owner, repo
				}
				// Stop if we reach another section
				if strings.HasPrefix(urlLine, "[") {
					break
				}
			}
			break
		}
	}

	logger.Warn("No remote origin found in .git/config")
	return "", ""
}
