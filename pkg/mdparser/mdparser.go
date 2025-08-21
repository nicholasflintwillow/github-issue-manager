package mdparser

import (
	"os"
	"path/filepath"
	"strings"
)

// ListMarkdownFiles returns a slice of markdown file paths in the specified folder.
func ListMarkdownFiles(folder string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, filepath.Join(folder, entry.Name()))
		}
	}
	return files, nil
}

// ParseFrontMatter extracts key-value pairs from the front matter block in a markdown file.
func ParseFrontMatter(path string) (map[string]string, error) {
	result := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	inBlock := false
	var body []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "---" {
			if !inBlock {
				inBlock = true
				continue
			} else {
				inBlock = false
				continue
			}
		}
		if inBlock && line != "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove surrounding quotes if present
				value = strings.Trim(value, "\"")
				result[key] = value
			}
		} else if !inBlock {
			body = append(body, line)
		}
	}
	result["body"] = strings.Join(body, "\n")
	return result, nil
}
