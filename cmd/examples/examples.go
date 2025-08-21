package examples

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var outputDir string
var issueType string

// Flag variables for IssueData fields
var (
	// String fields
	flagTitle          string
	flagProject        string
	flagStatus         string
	flagLabels         string
	flagParent         string
	flagDescription    string
	flagExpectedResult string
	flagActualResult   string
	flagSeverity       string
	flagPriority       string
	flagAffectedUsers  string
	flagBusinessImpact string
	flagWorkaround     string

	// String slice fields
	flagImplementationDetails []string
	flagTechnicalRequirements []string
	flagTestingStrategy       []string
	flagDesignRequirements    []string
	flagReproSteps            []string
	flagEnvironment           []string
	flagErrorDetails          []string
	flagInvestigationNotes    []string
	flagRootCause             []string

	// Special field for FixDescription
	flagFixDescription string
)

var Cmd = &cobra.Command{
	Use:   "examples",
	Short: "Generate example issue files with all available fields",
	Long:  "Generate example markdown issue files for different types (Epic, Task, Bug, Feature) with all available fields and parent-child relationships",
	Run: func(cmd *cobra.Command, args []string) {
		if issueType != "" {
			generateSingleExample(issueType)
		} else {
			generateAllExamples()
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&outputDir, "output", "o", "examples", "Output directory for example files")
	Cmd.Flags().StringVarP(&issueType, "type", "t", "", "Generate example for specific type (epic, task, bug, feature)")

	// String field flags
	Cmd.Flags().StringVarP(&flagTitle, "title", "", "", "Issue title")
	Cmd.Flags().StringVarP(&flagProject, "project", "p", "", "Project name")
	Cmd.Flags().StringVarP(&flagStatus, "status", "s", "", "Issue status")
	Cmd.Flags().StringVarP(&flagLabels, "labels", "l", "", "Issue labels (comma-separated)")
	Cmd.Flags().StringVarP(&flagParent, "parent", "", "", "Parent issue title")
	Cmd.Flags().StringVarP(&flagDescription, "description", "d", "", "Issue description")
	Cmd.Flags().StringVarP(&flagExpectedResult, "expected-result", "", "", "Expected result for bugs")
	Cmd.Flags().StringVarP(&flagActualResult, "actual-result", "", "", "Actual result for bugs")
	Cmd.Flags().StringVarP(&flagSeverity, "severity", "", "", "Bug severity")
	Cmd.Flags().StringVarP(&flagPriority, "priority", "", "", "Issue priority")
	Cmd.Flags().StringVarP(&flagAffectedUsers, "affected-users", "", "", "Affected users description")
	Cmd.Flags().StringVarP(&flagBusinessImpact, "business-impact", "", "", "Business impact description")
	// String slice field flags
	Cmd.Flags().StringArrayVarP(&flagImplementationDetails, "implementation-details", "", []string{}, "Implementation details (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagTechnicalRequirements, "technical-requirements", "", []string{}, "Technical requirements (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagTestingStrategy, "testing-strategy", "", []string{}, "Testing strategy items (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagDesignRequirements, "design-requirements", "", []string{}, "Design requirements (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagReproSteps, "repro-steps", "", []string{}, "Reproduction steps (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagEnvironment, "environment", "", []string{}, "Environment details (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagErrorDetails, "error-details", "", []string{}, "Error details (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagInvestigationNotes, "investigation-notes", "", []string{}, "Investigation notes (can be used multiple times)")
	Cmd.Flags().StringArrayVarP(&flagRootCause, "root-cause", "", []string{}, "Root cause items (can be used multiple times)")

	// Special field for FixDescription
	Cmd.Flags().StringVarP(&flagFixDescription, "fix-description", "", "", "Fix description in format '0:Description1;1:Description2'")
	Cmd.Flags().StringVarP(&flagWorkaround, "workaround", "", "", "Workaround description")
}

// parseFixDescription parses the fix description flag format "0:Description1;1:Description2"
func parseFixDescription(input string) []struct {
	Index       int
	Description string
} {
	var result []struct {
		Index       int
		Description string
	}

	if input == "" {
		return result
	}

	parts := strings.Split(input, ";")
	for _, part := range parts {
		if colonIndex := strings.Index(part, ":"); colonIndex > 0 {
			indexStr := part[:colonIndex]
			description := part[colonIndex+1:]
			if index := 0; true {
				if _, err := fmt.Sscanf(indexStr, "%d", &index); err == nil {
					result = append(result, struct {
						Index       int
						Description string
					}{Index: index, Description: description})
				}
			}
		}
	}

	return result
}

// createDefaultIssueData creates an IssueData struct with default example values
func createDefaultIssueData(issueType string) IssueData {
	switch strings.ToLower(issueType) {
	case "epic":
		return IssueData{
			Title:       "Example Epic Title",
			Project:     "Example Project",
			Status:      "planning",
			Labels:      "epic, example",
			Description: "This is an example epic description with comprehensive details about the feature or initiative.",
			ImplementationDetails: []string{
				"Define architecture and technical approach",
				"Break down into smaller tasks",
				"Establish acceptance criteria",
			},
			TechnicalRequirements: []string{
				"Scalable design for future growth",
				"Performance requirements defined",
				"Security considerations addressed",
			},
			TestingStrategy: []string{
				"Unit testing for all components",
				"Integration testing strategy",
				"User acceptance testing plan",
			},
		}
	case "task":
		return IssueData{
			Title:       "Example Task Title",
			Project:     "Example Project",
			Status:      "todo",
			Labels:      "task, example",
			Description: "This is an example task description with specific implementation details.",
			ImplementationDetails: []string{
				"Implement core functionality",
				"Add error handling",
				"Write documentation",
			},
			TechnicalRequirements: []string{
				"Response time < 500ms",
				"Handle concurrent requests",
				"Follow coding standards",
			},
			TestingStrategy: []string{
				"Unit tests for business logic",
				"Integration tests for APIs",
				"Performance testing",
			},
		}
	case "bug":
		return IssueData{
			Title:          "Example Bug Title",
			Project:        "Example Project",
			Status:         "open",
			Labels:         "bug, example",
			Description:    "This is an example bug description with reproduction steps and expected behavior.",
			ExpectedResult: "The feature should work as designed",
			ActualResult:   "The feature exhibits unexpected behavior",
			Severity:       "Medium",
			Priority:       "High",
			AffectedUsers:  "Users experiencing the specific workflow",
			BusinessImpact: "Moderate impact on user experience",
			Workaround:     "Temporary workaround available",
			ReproSteps: []string{
				"Navigate to the affected page",
				"Perform the specific action",
				"Observe the unexpected behavior",
			},
			Environment: []string{
				"Browser: Chrome, Firefox",
				"OS: Windows, macOS",
				"Version: Latest",
			},
			ErrorDetails: []string{
				"Console shows no errors",
				"Network requests complete successfully",
				"UI state inconsistent",
			},
			InvestigationNotes: []string{
				"Issue appears to be timing-related",
				"Occurs more frequently under load",
				"May be related to recent changes",
			},
			RootCause: []string{
				"Race condition in state management",
				"Insufficient validation",
			},
			FixDescription: []struct {
				Index       int
				Description string
			}{
				{0, "Add proper state synchronization"},
				{1, "Implement validation checks"},
				{2, "Add error handling"},
			},
			TestingStrategy: []string{
				"Reproduce the issue consistently",
				"Test fix in multiple environments",
				"Verify no regression in related features",
			},
		}
	case "feature":
		return IssueData{
			Title:       "Example Feature Title",
			Project:     "Example Project",
			Status:      "backlog",
			Labels:      "feature, example, enhancement",
			Description: "This is an example feature description with user value and implementation approach.",
			ImplementationDetails: []string{
				"Design user interface components",
				"Implement backend API endpoints",
				"Add data persistence layer",
			},
			TechnicalRequirements: []string{
				"Mobile-responsive design",
				"Accessibility compliance",
				"Performance optimization",
			},
			DesignRequirements: []string{
				"Consistent with design system",
				"User-friendly interface",
				"Clear visual hierarchy",
			},
			TestingStrategy: []string{
				"Usability testing with target users",
				"A/B testing for feature adoption",
				"Performance testing under load",
			},
		}
	default:
		// Default to task-like structure
		return IssueData{
			Title:       "Example Issue Title",
			Project:     "Example Project",
			Status:      "todo",
			Labels:      "example",
			Description: "This is an example issue description.",
		}
	}
}

// Template data structures

type IssueData struct {
	Title       string
	Project     string
	Status      string
	Labels      string
	Parent      string
	Description string

	ImplementationDetails []string
	TechnicalRequirements []string

	TestingStrategy []string

	DesignRequirements []string

	// Bug-specific fields
	ReproSteps         []string
	ExpectedResult     string
	ActualResult       string
	Environment        []string
	ErrorDetails       []string
	InvestigationNotes []string
	Severity           string
	Priority           string
	AffectedUsers      string
	BusinessImpact     string
	Workaround         string
	RootCause          []string
	FixDescription     []struct {
		Index       int
		Description string
	}
}

func generateAllExamples() {
	fmt.Printf("Generating example issue files in directory: %s\n", outputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Generate Epic examples (parent issues)
	generateEpicExamples()

	// Generate Task examples (child issues)
	generateTaskExamples()

	// Generate Bug examples (child issues)
	generateBugExamples()

	// Generate Feature examples (child and parent issues)
	generateFeatureExamples()

	fmt.Printf("Example files generated successfully in %s/\n", outputDir)
	fmt.Println("\nGenerated files:")
	listGeneratedFiles()
}

// applyFlagOverrides applies command-line flag values to the IssueData struct
func applyFlagOverrides(data *IssueData) {
	// Apply string field overrides
	if flagTitle != "" {
		data.Title = flagTitle
	}
	if flagProject != "" {
		data.Project = flagProject
	}
	if flagStatus != "" {
		data.Status = flagStatus
	}
	if flagLabels != "" {
		data.Labels = flagLabels
	}
	if flagParent != "" {
		data.Parent = flagParent
	}
	if flagDescription != "" {
		data.Description = flagDescription
	}
	if flagExpectedResult != "" {
		data.ExpectedResult = flagExpectedResult
	}
	if flagActualResult != "" {
		data.ActualResult = flagActualResult
	}
	if flagSeverity != "" {
		data.Severity = flagSeverity
	}
	if flagPriority != "" {
		data.Priority = flagPriority
	}
	if flagAffectedUsers != "" {
		data.AffectedUsers = flagAffectedUsers
	}
	if flagBusinessImpact != "" {
		data.BusinessImpact = flagBusinessImpact
	}
	if flagWorkaround != "" {
		data.Workaround = flagWorkaround
	}

	// Apply string slice field overrides
	if len(flagImplementationDetails) > 0 {
		data.ImplementationDetails = flagImplementationDetails
	}
	if len(flagTechnicalRequirements) > 0 {
		data.TechnicalRequirements = flagTechnicalRequirements
	}
	if len(flagTestingStrategy) > 0 {
		data.TestingStrategy = flagTestingStrategy
	}
	if len(flagDesignRequirements) > 0 {
		data.DesignRequirements = flagDesignRequirements
	}
	if len(flagReproSteps) > 0 {
		data.ReproSteps = flagReproSteps
	}
	if len(flagEnvironment) > 0 {
		data.Environment = flagEnvironment
	}
	if len(flagErrorDetails) > 0 {
		data.ErrorDetails = flagErrorDetails
	}
	if len(flagInvestigationNotes) > 0 {
		data.InvestigationNotes = flagInvestigationNotes
	}
	if len(flagRootCause) > 0 {
		data.RootCause = flagRootCause
	}

	// Apply FixDescription override
	if flagFixDescription != "" {
		data.FixDescription = parseFixDescription(flagFixDescription)
	}
}

func generateSingleExample(issueType string) {
	fmt.Printf("Generating example for type: %s\n", issueType)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Create default IssueData with example values
	data := createDefaultIssueData(issueType)

	// Apply flag overrides to the default data
	applyFlagOverrides(&data)

	// Determine template and output filename based on issue type
	var templatePath, outputFilename string
	switch strings.ToLower(issueType) {
	case "epic":
		templatePath = "templates/epic-parent.md.tmpl"
		outputFilename = "epic-example.md"
	case "task":
		templatePath = "templates/task.md.tmpl"
		outputFilename = "task-example.md"
	case "bug":
		templatePath = "templates/bug.md.tmpl"
		outputFilename = "bug-example.md"
	case "feature":
		templatePath = "templates/feature.md.tmpl"
		outputFilename = "feature-example.md"
	default:
		fmt.Printf("Unknown issue type: %s. Available types: epic, task, bug, feature\n", issueType)
		return
	}

	// Generate the markdown file using the populated data
	generateFromTemplate(templatePath, outputFilename, data)

	fmt.Printf("Example file for %s generated successfully in %s/\n", issueType, outputDir)
}

func generateEpicExamples() {
	// Parent Epic data
	parentEpicData := IssueData{
		Title:       "User Authentication System Epic",
		Project:     "Auth Team",
		Status:      "in-progress",
		Labels:      "epic, high-priority, authentication",
		Description: "Implement a comprehensive user authentication system including login, registration, password reset, multi-factor authentication, and user session management.",

		ImplementationDetails: []string{
			"Use JWT for session management",
			"Implement OAuth2 for third-party authentication",
			"Password hashing with bcrypt",
			"Rate limiting for authentication endpoints",
			"Comprehensive logging and monitoring",
		},
	}

	// Child Epic data
	childEpicData := IssueData{
		Title:       "Multi-Factor Authentication Sub-Epic",
		Project:     "Auth Team",
		Status:      "backlog",
		Labels:      "epic, security, mfa",
		Parent:      "User Authentication System Epic",
		Description: "Implement multi-factor authentication (MFA) capabilities as part of the broader authentication system. This includes SMS, email, and authenticator app support.",

		ImplementationDetails: []string{
			"Integration with SMS providers (Twilio, AWS SNS)",
			"TOTP library integration",
			"Secure backup code generation",
			"User-friendly enrollment process",
		},
	}

	generateFromTemplate("templates/epic-parent.md.tmpl", "epic-parent-example.md", parentEpicData)
	generateFromTemplate("templates/epic-child.md.tmpl", "epic-child-example.md", childEpicData)
}

func generateTaskExamples() {
	// Task with parent
	taskWithParentData := IssueData{
		Title:       "Implement User Registration API",
		Project:     "Auth Team",
		Status:      "todo",
		Labels:      "backend, api, registration",
		Parent:      "User Authentication System Epic",
		Description: "Create RESTful API endpoints for user registration including email validation, password strength requirements, and duplicate email checking.",

		ImplementationDetails: []string{
			"Use Express.js framework",
			"Mongoose for MongoDB integration",
			"bcrypt for password hashing",
			"nodemailer for email verification",
			"joi for request validation",
		},
		TechnicalRequirements: []string{
			"Response time < 500ms",
			"Support for 1000+ concurrent registrations",
			"Secure password storage",
			"GDPR compliant data handling",
		},

		TestingStrategy: []string{
			"Unit tests for validation logic",
			"Integration tests for database operations",
			"Load testing for concurrent users",
			"Security testing for common vulnerabilities",
		},
	}

	// Standalone task
	standaloneTaskData := IssueData{
		Title:       "Setup Monitoring Dashboard",
		Project:     "DevOps Team",
		Status:      "in-progress",
		Labels:      "monitoring, infrastructure, grafana",
		Description: "Set up a comprehensive monitoring dashboard using Grafana to track application performance, system metrics, and user activity.",

		ImplementationDetails: []string{
			"Deploy Grafana on Kubernetes",
			"Configure Prometheus data sources",
			"Create custom dashboards",
			"Set up Slack/email notifications",
		},
	}

	generateFromTemplate("templates/task.md.tmpl", "task-with-parent-example.md", taskWithParentData)
	generateFromTemplate("templates/task.md.tmpl", "task-standalone-example.md", standaloneTaskData)
}

func generateBugExamples() {
	// Bug with parent
	bugWithParentData := IssueData{
		Title:       "Fix Password Reset Email Not Sending",
		Project:     "Auth Team",
		Status:      "open",
		Labels:      "bug, critical, email, password-reset",
		Parent:      "User Authentication System Epic",
		Description: "Users are not receiving password reset emails when requesting password reset. The reset request appears to succeed but no email is delivered.",
		ReproSteps: []string{
			"Go to login page",
			"Click \"Forgot Password\"",
			"Enter valid email address",
			"Click \"Send Reset Link\"",
			"Check email inbox and spam folder",
		},
		ExpectedResult: "Password reset email is received within 2-3 minutes",
		ActualResult:   "No email is received",
		Environment: []string{
			"Browser: Chrome 91.0.4472.124",
			"OS: Windows 10, macOS Big Sur",
			"Environment: Production and Staging",
		},
		ErrorDetails: []string{
			"No errors in browser console",
			"Server logs show 200 response",
			"Email service logs show failed delivery attempts",
		},
		InvestigationNotes: []string{
			"Email service rate limits may be exceeded",
			"SMTP configuration might be incorrect",
			"Email templates may have formatting issues",
		},
		Severity:       "Critical",
		Priority:       "High",
		AffectedUsers:  "All users requesting password reset",
		BusinessImpact: "Users cannot recover their accounts",
		Workaround:     "Customer support can manually reset passwords through admin panel.",
		RootCause: []string{
			"Email service API key expired",
			"Rate limiting configuration too restrictive",
			"Email queue processing stuck",
		},
		FixDescription: []struct {
			Index       int
			Description string
		}{
			{0, "Update email service API credentials"},
			{1, "Adjust rate limiting configuration"},
			{2, "Implement email queue monitoring"},
			{3, "Add retry mechanism for failed emails"},
		},
		TestingStrategy: []string{
			"Test password reset flow end-to-end",
			"Verify email delivery in all environments",
			"Test rate limiting edge cases",
			"Verify email template rendering",
		},
	}

	// Standalone bug
	standaloneBugData := IssueData{
		Title:       "Memory Leak in File Upload Component",
		Project:     "Frontend Team",
		Status:      "confirmed",
		Labels:      "bug, performance, memory-leak, file-upload",
		Description: "The file upload component is causing memory leaks when uploading large files or multiple files in succession. Browser memory usage increases continuously and is not freed after upload completion.",
		ReproSteps: []string{
			"Open developer tools and monitor memory usage",
			"Navigate to file upload page",
			"Upload multiple large files (>10MB) consecutively",
			"Observe memory usage in dev tools",
		},
		ExpectedResult: "Memory usage returns to baseline after uploads complete",
		ActualResult:   "Memory usage increases with each upload and is never freed",
		Environment: []string{
			"Browser: Chrome, Firefox, Safari",
			"File sizes: >5MB consistently reproduce the issue",
			"Component: FileUploadComponent.vue",
		},
		InvestigationNotes: []string{
			"Issue occurs in all modern browsers",
			"More pronounced with image files",
			"Component unmount doesn't free memory",
			"Suspect canvas elements or blob URLs not being cleaned up",
		},
		Severity:       "Medium",
		Priority:       "High",
		AffectedUsers:  "Users uploading large files",
		BusinessImpact: "Users abandon uploads due to poor performance",
		TestingStrategy: []string{
			"Memory profiling with Chrome DevTools",
			"Load testing with various file types and sizes",
			"Automated tests for cleanup verification",
		},
	}

	generateFromTemplate("templates/bug.md.tmpl", "bug-with-parent-example.md", bugWithParentData)
	generateFromTemplate("templates/bug.md.tmpl", "bug-standalone-example.md", standaloneBugData)
}

func generateFeatureExamples() {
	// Parent feature
	parentFeatureData := IssueData{
		Title:       "Advanced Search and Filtering System",
		Project:     "Product Team",
		Status:      "planning",
		Labels:      "feature, search, user-experience, enhancement",
		Description: "Implement an advanced search and filtering system that allows users to find content quickly using multiple criteria, saved searches, and intelligent suggestions.",

		ImplementationDetails: []string{
			"Elasticsearch for search engine",
			"Real-time indexing of content",
			"Machine learning for search result ranking",
			"Redis for caching frequent searches",
			"React components for search UI",
		},
		TechnicalRequirements: []string{
			"Search response time < 200ms",
			"Support for 1000+ concurrent searches",
			"99.9% search service uptime",
			"Internationalization support",
		},

		TestingStrategy: []string{
			"A/B testing with user groups",
			"Performance testing under load",
			"Usability testing with focus groups",
			"Search relevance quality testing",
		},
	}

	// Child feature
	childFeatureData := IssueData{
		Title:       "Search Autocomplete and Suggestions",
		Project:     "Product Team",
		Status:      "backlog",
		Labels:      "feature, search, autocomplete, ui",
		Parent:      "Advanced Search and Filtering System",
		Description: "Implement intelligent autocomplete and search suggestions to help users formulate better search queries and discover relevant content.",

		ImplementationDetails: []string{
			"Debounced API calls (300ms delay)",
			"Trie data structure for efficient prefix matching",
			"Levenshtein distance for typo correction",
			"LRU cache for popular suggestions",
			"Accessibility compliance (ARIA labels, screen reader support)",
		},
		TechnicalRequirements: []string{
			"Suggestion response time < 100ms",
			"Support for 500+ concurrent suggestion requests",
			"Fuzzy matching with 2-character tolerance",
			"Mobile-responsive design",
		},
	}

	// Standalone feature
	standaloneFeatureData := IssueData{
		Title:       "Dark Mode Theme Support",
		Project:     "Frontend Team",
		Status:      "ready",
		Labels:      "feature, ui, theme, accessibility",
		Description: "Add dark mode theme support across the entire application to improve user experience in low-light conditions and reduce eye strain.",

		ImplementationDetails: []string{
			"CSS custom properties for theming",
			"Context API for theme state management",
			"localStorage for theme persistence",
			"prefers-color-scheme media query support",
			"Theme-specific asset variants",
		},
		TechnicalRequirements: []string{
			"Theme switch response time < 100ms",
			"No layout shift during theme transitions",
			"Support for all browsers (IE11+)",
			"Backwards compatibility maintained",
		},
		DesignRequirements: []string{
			"Consistent color palette across components",
			"Proper contrast ratios (4.5:1 for normal text)",
			"Brand colors adapted for dark backgrounds",
			"User testing with accessibility users",
		},

		TestingStrategy: []string{
			"Visual regression testing for all components",
			"Accessibility testing with screen readers",
			"Cross-browser compatibility testing",
			"User acceptance testing with focus groups",
		},
	}

	generateFromTemplate("templates/feature.md.tmpl", "feature-parent-example.md", parentFeatureData)
	generateFromTemplate("templates/feature.md.tmpl", "feature-child-example.md", childFeatureData)
	generateFromTemplate("templates/feature.md.tmpl", "feature-standalone-example.md", standaloneFeatureData)
}

func generateFromTemplate(templatePath, outputFilename string, data IssueData) {
	// Create custom template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Read template from embedded filesystem or local filesystem
	var tmplContent []byte
	var err error

	// Try embedded filesystem first
	tmplContent, err = templatesFS.ReadFile(templatePath)
	if err != nil {
		// Fall back to local filesystem
		tmplContent, err = os.ReadFile(templatePath)
		if err != nil {
			fmt.Printf("Error reading template %s: %v\n", templatePath, err)
			return
		}
	}

	// Parse the template
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templatePath, err)
		return
	}

	// Create the output file
	fullPath := filepath.Join(outputDir, outputFilename)
	file, err := os.Create(fullPath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", fullPath, err)
		return
	}
	defer file.Close()

	// Execute the template
	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Printf("Error executing template for %s: %v\n", fullPath, err)
		return
	}

	fmt.Printf("Created: %s\n", fullPath)
}

func listGeneratedFiles() {
	files, err := os.ReadDir(outputDir)
	if err != nil {
		fmt.Printf("Error reading output directory: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			fmt.Printf("  - %s\n", file.Name())
		}
	}
}
