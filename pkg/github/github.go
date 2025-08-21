package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github-issue-manager/pkg/issuemanager"
	"github-issue-manager/pkg/logger"

	"github.com/machinebox/graphql"
)

// Client holds the GitHub GraphQL client.
type Client struct {
	GraphQL *graphql.Client
}

// IssueResult represents the result of creating an issue.
type IssueResult struct {
	Number int64  // Issue number (e.g. 123)
	NodeID string // GraphQL node ID (needed for Projects v2)
	Err    error
}

// OPTIONAL: ensure your issue model has a Type field.
// type Issue struct {
//   ...
//   Type string // e.g. "Bug", "Task", "Feature" (optional)
// }

// ReadTokenFromHostsFile reads the GitHub token from the hosts.yml file.
func ReadTokenFromHostsFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	hostsFilePath := filepath.Join(homeDir, ".config", "gh", "hosts.yml")
	data, err := ioutil.ReadFile(hostsFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Simple parser for the hosts.yml file
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, "oauth_token:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				token := strings.TrimSpace(parts[1])
				token = strings.Trim(token, " \"'")
				return token, nil
			}
		}
	}

	return "", fmt.Errorf("oauth_token not found in hosts file")
}

// --- NEW: helper to get a token (env first, then gh hosts.yml)
func (c *Client) getToken() (string, error) {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t, nil
	}
	if hostsToken, err := ReadTokenFromHostsFile(); err == nil && hostsToken != "" {
		return hostsToken, nil
	}
	return "", fmt.Errorf("GITHUB_TOKEN not set and hosts file token not found")
}

// CreateIssues creates multiple GitHub issues in dependency order.
func (c *Client) CreateIssues(ctx context.Context, owner, repo string, issues []issuemanager.Issue) {
	// Sort issues so parent issues are created before children
	sortedIssues := issuemanager.SortIssuesByDependency(issues)

	// Map to store created issue numbers by title for parent-child linking
	createdIssues := make(map[string]int64)

	for _, issue := range sortedIssues {
		// if the id isn't in the file then it's not in github
		var issueResponse IssueResult
		if issue.Id == "" {
			// Use GraphQL with Issue Type when provided, otherwise use standard GraphQL creation
			if strings.TrimSpace(issue.Type) != "" {
				issueResponse = c.CreateIssueWithTypeGraphQL(ctx, owner, repo, issue)
			} else {
				issueResponse = c.CreateIssue(ctx, owner, repo, issue) // GraphQL creation
			}

			// Read the markdown file
			filePath := filepath.Join(issue.Path, issue.FileName)
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				logger.Error("Failed to read markdown file", "file", filePath, "error", err)
				continue
			} // Update the front matter with the new issue ID
			lines := strings.Split(string(data), "\n")
			idFound := false
			for i, line := range lines {
				if strings.HasPrefix(line, "id:") {
					lines[i] = "id: " + strconv.FormatInt(issueResponse.Number, 10)
					idFound = true
					break
				}
			}

			if !idFound {
				frontMatterEnd := 0
				for i, line := range lines {
					if strings.TrimSpace(line) == "---" {
						if frontMatterEnd == 0 {
							frontMatterEnd = i
						} else {
							break
						}
					}
				}
				if frontMatterEnd != 0 {
					lines = append(
						lines[:frontMatterEnd],
						append([]string{"id: " + strconv.FormatInt(issueResponse.Number, 10)}, lines[frontMatterEnd:]...)...,
					)
				} else {
					lines = append(lines, "id: "+strconv.FormatInt(issueResponse.Number, 10))
				}
			}

			// Write back to the markdown file
			err = ioutil.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
			if err != nil {
				logger.Error("Failed to update markdown file", "file", filePath, "error", err)
			}
		} else {
			fmt.Printf("Issue '%s' already exists (#%s), updating...\n", issue.Title, issue.Id)
			idInt, err := strconv.ParseInt(issue.Id, 10, 64)
			if err != nil {
				logger.Error("Failed to convert issue ID to int64", "id", issue.Id, "error", err)
				issueResponse = IssueResult{Number: 0, Err: err}
			} else {
				// Update the existing issue
				if strings.TrimSpace(issue.Type) != "" {
					issueResponse = c.UpdateIssueWithTypeGraphQL(ctx, owner, repo, issue, idInt)
				} else {
					issueResponse = c.UpdateIssue(ctx, owner, repo, issue, idInt)
				}

				if issueResponse.Err != nil {
					logger.Error("Failed to update issue", "issue", issue.Title, "error", issueResponse.Err)
				} else {
					fmt.Printf("Successfully updated issue '%s' (#%d)\n", issue.Title, issueResponse.Number)
				}
			}
		}

		// Store the created/updated issue number for parent-child linking
		if issueResponse.Err == nil {
			createdIssues[issue.Title] = issueResponse.Number
		}

		// Add issue to project if project name is provided
		if issue.Project != "" {
			// Get issue node id by issue number using GraphQL
			issueNodeID, err := c.ResolveIssueNodeID(ctx, owner, repo, issueResponse.Number)
			if err != nil {
				fmt.Println("Error resolving issue node ID:", err)
				continue
			}

			// Resolve project name to GraphQL ID
			logger.Debug("Resolving project name to GraphQL node ID", "project", issue.Project)
			projectNodeID, err := c.ResolveProjectID(ctx, owner, issue.Project)
			if err != nil {
				fmt.Println("Error resolving project ID:", err)
			}
			// Add issue to project
			if err := c.AddIssueToProject(ctx, issueNodeID, projectNodeID); err != nil {
				fmt.Printf("Failed to add issue %s to project '%s': %v\n", issue.Title, issue.Project, err)
			}
		}
	}

	fmt.Printf("Created %d issues successfully.\n", len(createdIssues))
}

// ResolveIssueNodeID resolves an issue number to its GraphQL node ID using GraphQL.
func (c *Client) ResolveIssueNodeID(ctx context.Context, owner, repo string, issueNumber int64) (string, error) {
	if issueNumber <= 0 {
		return "", fmt.Errorf("invalid issue number: %d", issueNumber)
	}

	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	// Use GraphQL to get the issue by number
	req := graphql.NewRequest(`
		query($owner: String!, $name: String!, $number: Int!) {
			repository(owner: $owner, name: $name) {
				issue(number: $number) {
					id
					number
				}
			}
		}
	`)
	req.Var("owner", owner)
	req.Var("name", repo)
	req.Var("number", int(issueNumber))
	req.Header.Set("Authorization", "Bearer "+token)

	var out struct {
		Repository struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
			} `json:"issue"`
		} `json:"repository"`
	}

	if err := c.GraphQL.Run(ctx, req, &out); err != nil {
		return "", fmt.Errorf("failed to query issue via GraphQL: %w", err)
	}

	if out.Repository.Issue.ID == "" {
		return "", fmt.Errorf("issue #%d not found in %s/%s", issueNumber, owner, repo)
	}

	return out.Repository.Issue.ID, nil
}

// ResolveProjectID resolves a project name to its GraphQL node ID.
func (c *Client) ResolveProjectID(ctx context.Context, owner string, projectName string) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	type respPage struct {
		Organization struct {
			ProjectsV2 struct {
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"organization"`
	}

	var after *string
	for {
		req := graphql.NewRequest(`
			query OrgProjects($login: String!, $first: Int = 50, $after: String) {
				organization(login: $login) {
					projectsV2(
						first: $first
						after: $after
						orderBy: { field: UPDATED_AT, direction: DESC }
					) {
						pageInfo { hasNextPage endCursor }
						nodes { id title }
					}
				}
			}
		`)
		req.Var("login", owner)
		req.Var("first", 50)
		req.Var("after", after)
		req.Header.Set("Authorization", "Bearer "+token)

		var out respPage
		if err := c.GraphQL.Run(ctx, req, &out); err != nil {
			return "", fmt.Errorf("failed to query GitHub GraphQL API: %w", err)
		}

		for _, n := range out.Organization.ProjectsV2.Nodes {
			if strings.EqualFold(strings.TrimSpace(n.Title), strings.TrimSpace(projectName)) {
				return n.ID, nil
			}
		}

		if !out.Organization.ProjectsV2.PageInfo.HasNextPage || out.Organization.ProjectsV2.PageInfo.EndCursor == nil {
			break
		}
		after = out.Organization.ProjectsV2.PageInfo.EndCursor
	}

	return "", fmt.Errorf("project with name %q not found", projectName)
}

// NewClient creates a new GitHub client with GraphQL support.
func NewClient(ctx context.Context, pat string) *Client {
	return &Client{
		GraphQL: graphql.NewClient("https://api.github.com/graphql"),
	}
}

// AddIssueToProject adds an issue to a GitHub project using GraphQL.
func (c *Client) AddIssueToProject(ctx context.Context, issueNodeID, projectNodeID string) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	req := graphql.NewRequest(`
		mutation($issueID: ID!, $projectID: ID!) {
		  addProjectV2ItemById(input: { projectId: $projectID, contentId: $issueID }) {
		    item { id }
		  }
		}
	`)
	req.Var("issueID", issueNodeID)
	req.Var("projectID", projectNodeID)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	}

	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		if strings.Contains(err.Error(), "content already exists in the project") {
			return nil
		}
		return fmt.Errorf("failed to add issue to project via GraphQL: %w", err)
	}

	if resp.AddProjectV2ItemById.Item.ID == "" {
		return fmt.Errorf("GraphQL succeeded but returned empty item id")
	}
	return nil
}

// CreateIssue creates a GitHub issue using GraphQL with proper parent relationship.
func (c *Client) CreateIssue(ctx context.Context, owner, repo string, issue issuemanager.Issue) IssueResult {
	token, err := c.getToken()
	if err != nil {
		return IssueResult{Err: err}
	}

	// Get repository node ID
	repoID, err := c.ResolveRepositoryID(ctx, owner, repo)
	if err != nil {
		return IssueResult{Err: fmt.Errorf("resolve repository id: %w", err)}
	}

	// Create the issue via GraphQL
	req := graphql.NewRequest(`
		mutation($input: CreateIssueInput!) {
			createIssue(input: $input) {
				issue {
					id
					number
					title
				}
			}
		}
	`)

	// Build input with labels if provided
	input := map[string]interface{}{
		"repositoryId": repoID,
		"title":        issue.Title,
		"body":         issue.Body,
	}

	// Add parent relationship if specified
	if strings.TrimSpace(issue.Parent) != "" {
		if parentID, err := c.ResolveParentIssueID(ctx, owner, repo, issue.Parent); err == nil {
			input["parentIssueId"] = parentID
			logger.Info("Setting parent relationship", "child", issue.Title, "parent", issue.Parent)
		} else {
			logger.Warn("Could not resolve parent issue", "parent", issue.Parent, "error", err)
		}
	}

	// Add labels if they exist
	if len(issue.Labels) > 0 {
		input["labelIds"] = c.resolveLabelIDs(ctx, owner, repo, issue.Labels)
	}

	req.Var("input", input)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		CreateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int64  `json:"number"`
				Title  string `json:"title"`
			} `json:"issue"`
		} `json:"createIssue"`
	}

	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		return IssueResult{Err: fmt.Errorf("createIssue GraphQL failed: %w", err)}
	}

	if resp.CreateIssue.Issue.ID == "" {
		return IssueResult{Err: fmt.Errorf("createIssue GraphQL returned empty issue id")}
	}

	return IssueResult{
		Number: resp.CreateIssue.Issue.Number,
		NodeID: resp.CreateIssue.Issue.ID,
		Err:    nil,
	}
}

// UpdateIssue updates an existing GitHub issue using GraphQL with proper parent relationship.
func (c *Client) UpdateIssue(ctx context.Context, owner, repo string, issue issuemanager.Issue, issueNumber int64) IssueResult {
	token, err := c.getToken()
	if err != nil {
		return IssueResult{Err: err}
	}

	// Get the issue node ID first
	issueNodeID, err := c.ResolveIssueNodeID(ctx, owner, repo, issueNumber)
	if err != nil {
		return IssueResult{Err: fmt.Errorf("resolve issue node id: %w", err)}
	}

	// Create the GraphQL mutation for updating the issue
	req := graphql.NewRequest(`
		mutation($input: UpdateIssueInput!) {
			updateIssue(input: $input) {
				issue {
					id
					number
					title
					body
				}
			}
		}
	`)

	// Build input with the fields to update
	input := map[string]interface{}{
		"id":    issueNodeID,
		"title": issue.Title,
		"body":  issue.Body,
	}

	// Handle parent relationship update using addSubIssue mutation
	if strings.TrimSpace(issue.Parent) != "" {
		if err := c.UpdateParentRelationship(ctx, owner, repo, issueNodeID, issue.Parent); err != nil {
			logger.Warn("Failed to update parent relationship", "issue", issue.Title, "error", err)
		} else {
			logger.Info("Successfully updated parent relationship", "child", issue.Title, "parent", issue.Parent)
		}
	}

	// Add labels if they exist
	if len(issue.Labels) > 0 {
		input["labelIds"] = c.resolveLabelIDs(ctx, owner, repo, issue.Labels)
	}

	req.Var("input", input)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		UpdateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int64  `json:"number"`
				Title  string `json:"title"`
				Body   string `json:"body"`
			} `json:"issue"`
		} `json:"updateIssue"`
	}

	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		return IssueResult{Err: fmt.Errorf("updateIssue GraphQL failed: %w", err)}
	}

	if resp.UpdateIssue.Issue.ID == "" {
		return IssueResult{Err: fmt.Errorf("updateIssue GraphQL returned empty issue id")}
	}

	return IssueResult{
		Number: resp.UpdateIssue.Issue.Number,
		NodeID: resp.UpdateIssue.Issue.ID,
		Err:    nil,
	}
}

// --- NEW: GraphQL path to create an issue with an Issue Type and proper parent relationship
func (c *Client) CreateIssueWithTypeGraphQL(ctx context.Context, owner, repo string, issue issuemanager.Issue) IssueResult {
	token, err := c.getToken()
	if err != nil {
		return IssueResult{Err: err}
	}

	// 1) Resolve repository node ID
	repoID, err := c.ResolveRepositoryID(ctx, owner, repo)
	if err != nil {
		return IssueResult{Err: fmt.Errorf("resolve repository id: %w", err)}
	}

	// 2) Resolve Issue Type ID (by human-friendly name like "Bug")
	typeID, err := c.ResolveIssueTypeID(ctx, owner, repo, strings.TrimSpace(issue.Type))
	if err != nil {
		return IssueResult{Err: fmt.Errorf("resolve issue type id for %q: %w", issue.Type, err)}
	}

	// 3) Create the issue via GraphQL with issueTypeId
	req := graphql.NewRequest(`
		mutation($input: CreateIssueInput!) {
			createIssue(input: $input) {
				issue {
					id
					number
					node_id: id
					title
					issueType { id name }
				}
			}
		}
	`)

	// Build input with labels if provided
	input := map[string]interface{}{
		"repositoryId": repoID,
		"title":        issue.Title,
		"body":         issue.Body,
		"issueTypeId":  typeID,
	}

	// Add parent relationship if specified
	if strings.TrimSpace(issue.Parent) != "" {
		if parentID, err := c.ResolveParentIssueID(ctx, owner, repo, issue.Parent); err == nil {
			input["parentIssueId"] = parentID
			logger.Info("Setting parent relationship", "child", issue.Title, "parent", issue.Parent)
		} else {
			logger.Warn("Could not resolve parent issue", "parent", issue.Parent, "error", err)
		}
	}

	// Add labels if they exist
	if len(issue.Labels) > 0 {
		input["labelIds"] = c.resolveLabelIDs(ctx, owner, repo, issue.Labels)
	}

	req.Var("input", input)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		CreateIssue struct {
			Issue struct {
				ID        string `json:"id"`
				Number    int64  `json:"number"`
				Title     string `json:"title"`
				IssueType struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"issueType"`
			} `json:"issue"`
		} `json:"createIssue"`
	}
	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		return IssueResult{Err: fmt.Errorf("createIssue GraphQL failed: %w", err)}
	}
	if resp.CreateIssue.Issue.ID == "" {
		return IssueResult{Err: fmt.Errorf("createIssue GraphQL returned empty issue id")}
	}

	return IssueResult{
		Number: resp.CreateIssue.Issue.Number,
		NodeID: resp.CreateIssue.Issue.ID,
		Err:    nil,
	}
}

// UpdateIssueWithTypeGraphQL updates an existing GitHub issue with issue type using GraphQL and proper parent relationship.
func (c *Client) UpdateIssueWithTypeGraphQL(ctx context.Context, owner, repo string, issue issuemanager.Issue, issueNumber int64) IssueResult {
	token, err := c.getToken()
	if err != nil {
		return IssueResult{Err: err}
	}

	// Get the issue node ID first
	issueNodeID, err := c.ResolveIssueNodeID(ctx, owner, repo, issueNumber)
	if err != nil {
		return IssueResult{Err: fmt.Errorf("resolve issue node id: %w", err)}
	}

	// Resolve Issue Type ID if provided
	var typeID string
	if strings.TrimSpace(issue.Type) != "" {
		typeID, err = c.ResolveIssueTypeID(ctx, owner, repo, strings.TrimSpace(issue.Type))
		if err != nil {
			return IssueResult{Err: fmt.Errorf("resolve issue type id for %q: %w", issue.Type, err)}
		}
	}

	// Create the GraphQL mutation for updating the issue with type
	req := graphql.NewRequest(`
		mutation($input: UpdateIssueInput!) {
			updateIssue(input: $input) {
				issue {
					id
					number
					title
					body
					issueType { id name }
				}
			}
		}
	`)

	// Build input with the fields to update
	input := map[string]interface{}{
		"id":    issueNodeID,
		"title": issue.Title,
		"body":  issue.Body,
	}

	// Add issue type if provided
	if typeID != "" {
		input["issueTypeId"] = typeID
	}

	// Handle parent relationship update using addSubIssue mutation
	if strings.TrimSpace(issue.Parent) != "" {
		if err := c.UpdateParentRelationship(ctx, owner, repo, issueNodeID, issue.Parent); err != nil {
			logger.Warn("Failed to update parent relationship", "issue", issue.Title, "error", err)
		} else {
			logger.Info("Successfully updated parent relationship", "child", issue.Title, "parent", issue.Parent)
		}
	}

	// Add labels if they exist
	if len(issue.Labels) > 0 {
		input["labelIds"] = c.resolveLabelIDs(ctx, owner, repo, issue.Labels)
	}

	req.Var("input", input)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		UpdateIssue struct {
			Issue struct {
				ID        string `json:"id"`
				Number    int64  `json:"number"`
				Title     string `json:"title"`
				Body      string `json:"body"`
				IssueType struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"issueType"`
			} `json:"issue"`
		} `json:"updateIssue"`
	}

	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		return IssueResult{Err: fmt.Errorf("updateIssue GraphQL failed: %w", err)}
	}

	if resp.UpdateIssue.Issue.ID == "" {
		return IssueResult{Err: fmt.Errorf("updateIssue GraphQL returned empty issue id")}
	}

	return IssueResult{
		Number: resp.UpdateIssue.Issue.Number,
		NodeID: resp.UpdateIssue.Issue.ID,
		Err:    nil,
	}
}

// --- NEW: Resolve repository node ID (needed by createIssue)
func (c *Client) ResolveRepositoryID(ctx context.Context, owner, repo string) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	req := graphql.NewRequest(`
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) { id }
		}
	`)
	req.Var("owner", owner)
	req.Var("name", repo)
	req.Header.Set("Authorization", "Bearer "+token)

	var out struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}
	if err := c.GraphQL.Run(ctx, req, &out); err != nil {
		return "", fmt.Errorf("repository query failed: %w", err)
	}
	if out.Repository.ID == "" {
		return "", fmt.Errorf("repository id empty for %s/%s", owner, repo)
	}
	return out.Repository.ID, nil
}

// --- NEW: Resolve Issue Type ID by name (e.g., "Bug", "Task") for a repo
func (c *Client) ResolveIssueTypeID(ctx context.Context, owner, repo, typeName string) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(typeName) == "" {
		return "", fmt.Errorf("issue type name is empty")
	}

	req := graphql.NewRequest(`
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				issueTypes(first: 50) {
					nodes { id name }
				}
			}
		}
	`)
	req.Var("owner", owner)
	req.Var("name", repo)
	req.Header.Set("Authorization", "Bearer "+token)

	var out struct {
		Repository struct {
			IssueTypes struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"issueTypes"`
		} `json:"repository"`
	}
	if err := c.GraphQL.Run(ctx, req, &out); err != nil {
		return "", fmt.Errorf("issueTypes query failed: %w", err)
	}
	for _, n := range out.Repository.IssueTypes.Nodes {
		if strings.EqualFold(strings.TrimSpace(n.Name), strings.TrimSpace(typeName)) {
			return n.ID, nil
		}
	}
	return "", fmt.Errorf("issue type %q not found/enabled in %s/%s", typeName, owner, repo)
}

// resolveLabelIDs converts label names to their GraphQL node IDs.
func (c *Client) resolveLabelIDs(ctx context.Context, owner, repo string, labelNames []string) []string {
	if len(labelNames) == 0 {
		return nil
	}

	token, err := c.getToken()
	if err != nil {
		logger.Error("Failed to get token for label resolution", "error", err)
		return nil
	}

	req := graphql.NewRequest(`
		query($owner: String!, $name: String!, $first: Int!) {
			repository(owner: $owner, name: $name) {
				labels(first: $first) {
					nodes {
						id
						name
					}
				}
			}
		}
	`)
	req.Var("owner", owner)
	req.Var("name", repo)
	req.Var("first", 100) // Should be enough for most repos
	req.Header.Set("Authorization", "Bearer "+token)

	var out struct {
		Repository struct {
			Labels struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"labels"`
		} `json:"repository"`
	}

	if err := c.GraphQL.Run(ctx, req, &out); err != nil {
		logger.Error("Failed to resolve label IDs", "error", err)
		return nil
	}

	var labelIDs []string
	for _, labelName := range labelNames {
		for _, label := range out.Repository.Labels.Nodes {
			if strings.EqualFold(strings.TrimSpace(label.Name), strings.TrimSpace(labelName)) {
				labelIDs = append(labelIDs, label.ID)
				break
			}
		}
	}

	return labelIDs
}

// ResolveParentIssueID resolves a parent issue title to its GraphQL node ID.
func (c *Client) ResolveParentIssueID(ctx context.Context, owner, repo, parentTitle string) (string, error) {
	if strings.TrimSpace(parentTitle) == "" {
		return "", fmt.Errorf("parent title is empty")
	}

	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	// Search for issues by title using GraphQL
	req := graphql.NewRequest(`
		query($query: String!, $first: Int!) {
			search(query: $query, type: ISSUE, first: $first) {
				nodes {
					... on Issue {
						id
						title
						number
						repository {
							owner { login }
							name
						}
					}
				}
			}
		}
	`)

	// Build search query: title + repository scope
	searchQuery := fmt.Sprintf(`"%s" repo:%s/%s in:title`, strings.TrimSpace(parentTitle), owner, repo)
	req.Var("query", searchQuery)
	req.Var("first", 10) // Should be enough to find the parent issue
	req.Header.Set("Authorization", "Bearer "+token)

	var out struct {
		Search struct {
			Nodes []struct {
				ID         string `json:"id"`
				Title      string `json:"title"`
				Number     int    `json:"number"`
				Repository struct {
					Owner struct {
						Login string `json:"login"`
					} `json:"owner"`
					Name string `json:"name"`
				} `json:"repository"`
			} `json:"nodes"`
		} `json:"search"`
	}

	if err := c.GraphQL.Run(ctx, req, &out); err != nil {
		return "", fmt.Errorf("failed to search for parent issue: %w", err)
	}

	// Look for exact title match
	for _, issue := range out.Search.Nodes {
		if strings.EqualFold(strings.TrimSpace(issue.Title), strings.TrimSpace(parentTitle)) &&
			strings.EqualFold(issue.Repository.Owner.Login, owner) &&
			strings.EqualFold(issue.Repository.Name, repo) {
			return issue.ID, nil
		}
	}

	return "", fmt.Errorf("parent issue with title %q not found in %s/%s", parentTitle, owner, repo)
}

// getIssueNumberFromNodeID retrieves the issue number from a GraphQL node ID.
func (c *Client) getIssueNumberFromNodeID(ctx context.Context, nodeID string) (int64, error) {
	token, err := c.getToken()
	if err != nil {
		return 0, err
	}

	req := graphql.NewRequest(`
		query($id: ID!) {
			node(id: $id) {
				... on Issue {
					number
				}
			}
		}
	`)
	req.Var("id", nodeID)
	req.Header.Set("Authorization", "Bearer "+token)

	var out struct {
		Node struct {
			Number int64 `json:"number"`
		} `json:"node"`
	}

	if err := c.GraphQL.Run(ctx, req, &out); err != nil {
		return 0, fmt.Errorf("failed to get issue number from node ID: %w", err)
	}

	return out.Node.Number, nil
}

// UpdateParentRelationship updates the parent relationship for an existing issue using addSubIssue mutation.
func (c *Client) UpdateParentRelationship(ctx context.Context, owner, repo, childNodeID, parentTitle string) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	// Resolve parent issue ID from title
	parentNodeID, err := c.ResolveParentIssueID(ctx, owner, repo, parentTitle)
	if err != nil {
		return fmt.Errorf("failed to resolve parent issue ID: %w", err)
	}

	// Use addSubIssue mutation to establish parent-child relationship
	req := graphql.NewRequest(`
		mutation($input: AddSubIssueInput!) {
			addSubIssue(input: $input) {
				issue {
					id
					title
				}
			}
		}
	`)

	input := map[string]interface{}{
		"issueId":       parentNodeID,
		"subIssueId":    childNodeID,
		"replaceParent": true, // Replace existing parent if one exists
	}

	req.Var("input", input)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		AddSubIssue struct {
			Issue struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"issue"`
		} `json:"addSubIssue"`
	}

	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		// Check if the error is about duplicate sub-issues, which means the relationship already exists
		if strings.Contains(err.Error(), "duplicate sub-issues") {
			logger.Info("Parent relationship already exists for issue")
			return nil
		}
		return fmt.Errorf("addSubIssue GraphQL failed: %w", err)
	}

	return nil
}

// RemoveParentRelationship removes a parent-child relationship using removeSubIssue mutation.
func (c *Client) RemoveParentRelationship(ctx context.Context, owner, repo, childNodeID, parentTitle string) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}

	// Resolve parent issue ID from title
	parentNodeID, err := c.ResolveParentIssueID(ctx, owner, repo, parentTitle)
	if err != nil {
		return fmt.Errorf("failed to resolve parent issue ID: %w", err)
	}

	// Use removeSubIssue mutation to break parent-child relationship
	req := graphql.NewRequest(`
		mutation($input: RemoveSubIssueInput!) {
			removeSubIssue(input: $input) {
				issue {
					id
					title
				}
			}
		}
	`)

	input := map[string]interface{}{
		"issueId":    parentNodeID,
		"subIssueId": childNodeID,
	}

	req.Var("input", input)
	req.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		RemoveSubIssue struct {
			Issue struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"issue"`
		} `json:"removeSubIssue"`
	}

	if err := c.GraphQL.Run(ctx, req, &resp); err != nil {
		return fmt.Errorf("removeSubIssue GraphQL failed: %w", err)
	}

	return nil
}

// ValidateProjectID validates that a project exists for the given organization.
func (c *Client) ValidateProjectID(ctx context.Context, owner string, project string) (bool, error) {
	token, err := c.getToken()
	if err != nil {
		return false, err
	}

	request := graphql.NewRequest(`
	query OrgProjects($login: String!, $first: Int = 20, $after: String) {
		organization(login: $login) {
			projectsV2(
			first: $first
			after: $after
			orderBy: { field: UPDATED_AT, direction: DESC }
			) {
				totalCount
				pageInfo { hasNextPage endCursor }
				nodes {
					id
					number
					title
					url
					public
					shortDescription
					updatedAt
				}
			}
		}
	}
	`)
	request.Var("login", owner)
	request.Var("first", 10)
	request.Header.Set("Authorization", "Bearer "+token)

	var resp struct {
		Node struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	if err := c.GraphQL.Run(ctx, request, &resp); err != nil {
		return false, fmt.Errorf("failed to query GitHub GraphQL API: %w", err)
	}

	return resp.Node.ID != "", nil
}
