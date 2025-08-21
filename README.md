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
### Generate Example Issue Files

Generate sample markdown issue files with comprehensive front matter fields:

```bash
# Generate all example types in 'examples' directory
./gim examples

# Generate specific issue type
./gim examples --type bug
./gim examples --type epic
./gim examples --type task
./gim examples --type feature

# Specify custom output directory
./gim examples --output my-examples

# Generate with custom field values
./gim examples --type bug --title "Custom Bug Title" --severity "High"
```

Available issue types:
- `bug`: Bug report with reproduction steps and investigation fields
- `epic`: Epic issue for large features or initiatives
- `task`: Development task with implementation details
- `feature`: Feature request with design and technical requirements

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

#### Core Fields (All Issue Types)
- `title`: Issue title (required)
- `project`: GitHub Project name to add the issue to
- `status`: Issue status (e.g., "open", "todo", "in-progress", "done")
- `labels`: Comma-separated list of labels
- `parent`: Title of parent issue for hierarchical relationships
- `id`: GitHub issue number (auto-populated after creation)

#### Bug-Specific Fields
- `repro-steps`: Array of reproduction steps
- `expected-result`: Expected behavior description
- `actual-result`: Actual behavior observed
- `severity`: Bug severity level (e.g., "Critical", "High", "Medium", "Low")
- `priority`: Issue priority (e.g., "Urgent", "High", "Medium", "Low")
- `affected-users`: Description of affected user groups
- `business-impact`: Business impact description
- `workaround`: Available workaround description
- `environment`: Array of environment details
- `error-details`: Array of error information
- `investigation-notes`: Array of investigation findings
- `root-cause`: Array of root cause analysis items
- `fix-description`: Structured fix description with indexed items

#### Development Fields (Tasks, Features, Epics)
- `implementation-details`: Array of implementation notes
- `technical-requirements`: Array of technical specifications
- `testing-strategy`: Array of testing approach items
- `design-requirements`: Array of design specifications (features)

#### Example Front Matter
```yaml
---
title: "Fix user authentication timeout"
project: "Auth Team"
status: "open"
labels: "bug, authentication, high-priority"
---
```

### Using the Examples Command

Generate comprehensive example files to understand all available fields:

```bash
# Generate all example types
./gim examples

# Generate a specific bug example with custom values
./gim examples --type bug --title "Login Page Crash" --severity "Critical" --priority "Urgent"
```

This creates example files in the `examples/` directory showing all supported front matter fields for each issue type.
```
---
severity: "High"
priority: "Urgent"
repro-steps:
  - "Log into the application"
  - "Wait for 30 minutes without activity"
  - "Attempt to perform an action"
expected-result: "User should be prompted to re-authenticate"
actual-result: "Application throws an error and crashes"
---
```

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
