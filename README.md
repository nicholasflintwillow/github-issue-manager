# GitHub Issue Manager

A powerful CLI tool for creating and managing GitHub issues from markdown files with support for hierarchical issue relationships, projects, and automated dependency resolution.

## Features

- **Markdown-driven Issue Creation**: Define issues in markdown files with YAML front matter
- **Hierarchical Issue Relationships**: Support for parent-child issue relationships and Epic management
- **GitHub Projects Integration**: Automatically add issues to GitHub Projects v2
- **Issue Type Support**: Create issues with specific types (Bug, Task, Epic, etc.)
- **Dependency Resolution**: Automatically sorts and creates issues in the correct order based on dependencies
- **Git Integration**: Automatically infers repository owner/name from local git configuration
- **Label Management**: Support for issue labels
- **Flexible Authentication**: Uses GitHub CLI token or environment variables
- **Structured Logging**: Configurable logging with debug, info, warn, and error levels

## Installation

```bash
go install github.com/nicholasflintwillow/github-issue-manager@latest
```

Or clone and build:

```bash
git clone https://github.com/nicholasflintwillow/github-issue-manager.git
cd github-issue-manager
go build -o gim .
```

## Usage

### Create Issues

Create GitHub issues from markdown files in a directory:

```bash
# Create issues from the default 'issues' folder
./gim create

# Create issues from a specific folder
./gim create -f path/to/issues

# Specify repository explicitly
./gim create -o owner-name -r repo-name

# Assign issues to a GitHub Project
./gim create -p "Project Name"

# Enable debug logging
./gim create --log-level debug

# Use JSON logging format
./gim create --log-json
```

### List Issues

List and preview issues before creation:

```bash
# List issues from default folder
./gim list

# List issues from specific folder
./gim list -f path/to/issues
```

### Issue Markdown Format

Create issues using markdown files with YAML front matter:

```markdown
---
title: "Issue Title"
project: "Project Name"
type: "Task"
status: "backlog"
labels: "bug, enhancement"
parent: "Parent Issue Title"
id: 42
---
## Description
Your issue description goes here.

## Implementation Details:
- Task 1
- Task 2
```

### Front Matter Fields

- `title`: Issue title (required)
- `type`: Issue type (Epic, Task, Bug, Feature, etc.)
- `project`: GitHub Project name to add the issue to
- `status`: Issue status
- `labels`: Comma-separated list of labels
- `parent`: Title of parent issue for hierarchical relationships
- `id`: GitHub issue number (auto-populated after creation)

## Architecture

The project is structured with clean separation of concerns:

- `cmd/`: Command-line interface commands (create, list)
- `pkg/git/`: Git repository integration and URL parsing
- `pkg/github/`: GitHub API client with GraphQL support
- `pkg/issuemanager/`: Issue management and dependency sorting logic
- `pkg/logger/`: Structured logging utilities
- `pkg/mdparser/`: Markdown front matter parsing

## Key Features

### Dependency Resolution
Issues are automatically sorted to ensure parent issues are created before children. Epic issues are processed last to maintain proper hierarchy.

### GraphQL Integration
Uses GitHub's GraphQL API for efficient operations including:
- Issue creation with types
- Parent-child relationship management
- Project assignment
- Label resolution

### Authentication
Supports multiple authentication methods:
- `GITHUB_TOKEN` environment variable
- GitHub CLI hosts configuration (`~/.config/gh/hosts.yml`)

## Examples

### Basic Issue Creation
```bash
./gim create -o myorg -r myrepo -p "Infrastructure Team"
```

### Epic with Child Issues
Create an epic and related tasks with automatic dependency resolution:

```markdown
# epic.md
---
title: "Phase 0: Planning and Setup"
type: "Epic"
project: "Infrastructure Team"
---

# task1.md  
---
title: "Setup Development Environment"
type: "Task"
parent: "Phase 0: Planning and Setup"
project: "Infrastructure Team"
---
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.