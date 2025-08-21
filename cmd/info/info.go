package info

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github-issue-manager/pkg/git"
	ghclient "github-issue-manager/pkg/github"
	"github-issue-manager/pkg/logger"
)

var owner string
var repo string

var Cmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about the GitHub repository",
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

		if owner == "" || repo == "" {
			logger.Error("Owner and repository name must be specified either via flags or inferred from .git/config")
			fmt.Fprintln(os.Stderr, "Owner and repository name must be specified either via flags or inferred from .git/config")
			os.Exit(1)
		}

		logger.Debug("Using owner and repo", "owner", owner, "repo", repo)

		repoInfo, err := client.GetRepositoryInfo(ctx, owner, repo)
		if err != nil {
			logger.Error("Failed to get repository info", "error", err)
			fmt.Fprintf(os.Stderr, "Failed to get repository info: %v\n", err)
			os.Exit(1)
		}

		// Handle edge case: empty repository info
		if repoInfo == nil {
			logger.Warn("Received empty repository info")
			fmt.Fprintln(os.Stderr, "Received empty repository info from GitHub")
			os.Exit(1)
		}

		jsonData, err := json.MarshalIndent(repoInfo, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal repository info to JSON", "error", err)
			fmt.Fprintf(os.Stderr, "Failed to format repository info as JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
	},
}

func init() {
	Cmd.Flags().StringVarP(&owner, "owner", "o", "", "GitHub owner name")
	Cmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub repository name")
}

func authenticate(ctx context.Context) *ghclient.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// Try to read token from hosts file
		hostsToken, err := ghclient.ReadTokenFromHostsFile()
		if err != nil {
			logger.Error("Failed to read token from hosts file", "error", err)
			fmt.Fprintf(os.Stderr, "Failed to read token from hosts file: %v\n", err)
			os.Exit(1)
		}
		if hostsToken != "" {
			token = hostsToken
		} else {
			logger.Error("GITHUB_TOKEN environment variable and hosts file token are both not set")
			fmt.Fprintln(os.Stderr, "GITHUB_TOKEN environment variable and hosts file token are both not set")
			os.Exit(1)
		}
	}

	logger.Debug("Creating GitHub client with token")
	return ghclient.NewClient(ctx, token)
}
