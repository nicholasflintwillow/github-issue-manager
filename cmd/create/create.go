package create

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github-issue-manager/pkg/git"
	ghclient "github-issue-manager/pkg/github"
	issuemanager "github-issue-manager/pkg/issuemanager"
)

// GitHubHostsConfig represents the structure of the hosts.yml file
type GitHubHostsConfig struct {
	GitHub GitHubConfig `yaml:"github.com"`
}

// GitHubConfig represents the configuration for github.com
type GitHubConfig struct {
	OAuthToken string `yaml:"oauth_token"`
}

// IssueConfig
type IssueConfig struct {
	Path     string
	FileName string
	Title    string
	Body     string
	Labels   []string
	Id       string
	// Add other issue-related configuration fields here
}

var folder string
var dryRun bool
var projectID string
var parentIssueID string
var owner string
var repo string

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create GitHub issues from markdown files",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := authenticate(ctx)

		// Infer owner and repo from .git/config if not provided via flags
		inferredOwner, inferredRepo := git.InferOwnerRepoFromGit()
		if owner == "" {
			owner = inferredOwner
		}
		if repo == "" {
			repo = inferredRepo
		}

		repoName := repo
		if repoName == "" {
			// Try to read the repository name from the current directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}
			repoName = filepath.Base(cwd)
		}

		fmt.Printf("Using owner: %s, repo: %s\n", owner, repoName)

		issues, err := issuemanager.ReadIssueFiles(folder)
		if err != nil {
			log.Fatalf("Error reading issue files: %v", err)
		}

		if projectID != "" {
			// If a project ID is provided, assign issues to the project
			fmt.Printf("Assigning issues to project ID: %s\n", projectID)
			for i := range issues {
				issues[i].Project = projectID
			}

		}

		fmt.Printf("First issue's project ID: %s\n", issues[0].Project)

		client.CreateIssues(ctx, owner, repoName, issues)
	},
}

func init() {
	// Dry run flag
	Cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode")
	Cmd.Flags().StringVarP(&folder, "folder", "f", "issues", "Folder containing issue files")
	Cmd.Flags().StringVarP(&owner, "owner", "o", "", "GitHub owner name")
	Cmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub repository name")
	Cmd.Flags().StringVarP(&projectID, "project", "p", "", "GitHub project ID to assign issues to")
	Cmd.Flags().StringVarP(&parentIssueID, "parent", "m", "", "Parent issue ID")
}

func authenticate(ctx context.Context) *ghclient.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// Try to read token from hosts file
		hostsToken, err := ghclient.ReadTokenFromHostsFile()
		if err != nil {
			log.Fatalf("Failed to read token from hosts file: %v", err)
		}
		if hostsToken != "" {
			token = hostsToken
		} else {
			log.Fatal("GITHUB_TOKEN environment variable and hosts file token are both not set")
		}
	}
	return ghclient.NewClient(ctx, token)
}
