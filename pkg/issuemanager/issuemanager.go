package issuemanager

import (
	"github-issue-manager/pkg/logger"
	mdparser "github-issue-manager/pkg/mdparser"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Issue represents the configuration for a GitHub issue.
type Issue struct {
	Path     string
	FileName string
	Title    string
	Body     string
	Labels   []string
	Type     string
	Id       string
	Project  string
	Parent   string // Parent issue title for hierarchical relationships
}

// ReadIssueFiles reads markdown files from the specified directory and extracts issue information.
func ReadIssueFiles(dir string) ([]Issue, error) {
	var issues []Issue
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		frontMatter, err := mdparser.ParseFrontMatter(filepath.Join(dir, file.Name()))
		if err != nil {
			logger.Error("Error parsing front matter", "file", file.Name(), "error", err)
			continue
		}
		labels := []string{}
		if lbls, ok := frontMatter["labels"]; ok {
			// Assuming labels are comma-separated
			for _, label := range strings.Split(lbls, ",") {
				trimmed := strings.TrimSpace(label)
				if trimmed != "" {
					labels = append(labels, trimmed)
				}
			}
		}
		issue := Issue{
			Path:     dir,
			FileName: file.Name(),
			Title:    frontMatter["title"],
			Body:     frontMatter["body"],
			Labels:   labels,
			Type:     frontMatter["type"],
			Project:  frontMatter["project"],
			Parent:   frontMatter["parent"],
			Id:       frontMatter["id"], // ID will be set after issue creation
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

// SortIssuesByDependency sorts issues so that parent issues are created before child issues,
// with epics processed last to ensure all their child issues are created first.
func SortIssuesByDependency(issues []Issue) []Issue {
	var sortedIssues []Issue
	var remainingIssues []Issue
	var epicIssues []Issue

	// Separate epics from regular issues
	var regularIssues []Issue
	for _, issue := range issues {
		if strings.EqualFold(strings.TrimSpace(issue.Type), "epic") {
			epicIssues = append(epicIssues, issue)
		} else {
			regularIssues = append(regularIssues, issue)
		}
	}

	// First, process regular issues without parents (root issues)
	for _, issue := range regularIssues {
		if strings.TrimSpace(issue.Parent) == "" {
			sortedIssues = append(sortedIssues, issue)
		} else {
			remainingIssues = append(remainingIssues, issue)
		}
	}

	// Then, iteratively add regular issues whose parents are already in the sorted list
	for len(remainingIssues) > 0 {
		var stillRemaining []Issue
		addedInThisIteration := false

		for _, issue := range remainingIssues {
			parentExists := false

			// Check if parent is already in sorted list
			for _, sortedIssue := range sortedIssues {
				if strings.EqualFold(strings.TrimSpace(sortedIssue.Title), strings.TrimSpace(issue.Parent)) {
					parentExists = true
					break
				}
			}

			if parentExists {
				sortedIssues = append(sortedIssues, issue)
				addedInThisIteration = true
			} else {
				stillRemaining = append(stillRemaining, issue)
			}
		}

		// If we couldn't resolve any dependencies in this iteration,
		// it means we have orphaned children (parent doesn't exist)
		// Add them anyway to prevent infinite loop
		if !addedInThisIteration {
			logger.Warn("Found regular issues with missing parents, adding them anyway")
			for _, issue := range stillRemaining {
				logger.Warn("Issue references missing parent", "issue", issue.Title, "parent", issue.Parent)
				sortedIssues = append(sortedIssues, issue)
			}
			break
		}

		remainingIssues = stillRemaining
	}

	// Finally, process epics (both with and without parents)
	var remainingEpics []Issue

	// First add epics without parents
	for _, epic := range epicIssues {
		if strings.TrimSpace(epic.Parent) == "" {
			sortedIssues = append(sortedIssues, epic)
		} else {
			remainingEpics = append(remainingEpics, epic)
		}
	}

	// Then add epics with parents (in dependency order)
	for len(remainingEpics) > 0 {
		var stillRemainingEpics []Issue
		addedInThisIteration := false

		for _, epic := range remainingEpics {
			parentExists := false

			// Check if parent is already in sorted list
			for _, sortedIssue := range sortedIssues {
				if strings.EqualFold(strings.TrimSpace(sortedIssue.Title), strings.TrimSpace(epic.Parent)) {
					parentExists = true
					break
				}
			}

			if parentExists {
				sortedIssues = append(sortedIssues, epic)
				addedInThisIteration = true
			} else {
				stillRemainingEpics = append(stillRemainingEpics, epic)
			}
		}

		// If we couldn't resolve any epic dependencies in this iteration,
		// add them anyway to prevent infinite loop
		if !addedInThisIteration {
			logger.Warn("Found epic issues with missing parents, adding them anyway")
			for _, epic := range stillRemainingEpics {
				logger.Warn("Epic references missing parent", "epic", epic.Title, "parent", epic.Parent)
				sortedIssues = append(sortedIssues, epic)
			}
			break
		}

		remainingEpics = stillRemainingEpics
	}

	return sortedIssues
}
