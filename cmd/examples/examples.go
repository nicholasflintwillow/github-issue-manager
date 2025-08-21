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
}

// Template data structures
type UserStory struct {
	AsA    string
	IWant  string
	SoThat string
}

type IssueData struct {
	Title                 string
	Project               string
	Status                string
	Labels                string
	Parent                string
	Description           string
	Domain                string
	AcceptanceCriteria    []string
	ImplementationDetails []string
	TechnicalRequirements []string
	Dependencies          []string
	Timeline              string
	TestingStrategy       []string
	UserStory             UserStory
	BusinessValue         []string
	SuccessMetrics        []string
	DesignRequirements    []string

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

func generateSingleExample(issueType string) {
	fmt.Printf("Generating example for type: %s\n", issueType)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	switch strings.ToLower(issueType) {
	case "epic":
		generateEpicExamples()
	case "task":
		generateTaskExamples()
	case "bug":
		generateBugExamples()
	case "feature":
		generateFeatureExamples()
	default:
		fmt.Printf("Unknown issue type: %s. Available types: epic, task, bug, feature\n", issueType)
		return
	}

	fmt.Printf("Example file(s) for %s generated successfully in %s/\n", issueType, outputDir)
}

func generateEpicExamples() {
	// Parent Epic data
	parentEpicData := IssueData{
		Title:       "User Authentication System Epic",
		Project:     "Auth Team",
		Status:      "in-progress",
		Labels:      "epic, high-priority, authentication",
		Description: "Implement a comprehensive user authentication system including login, registration, password reset, multi-factor authentication, and user session management.",
		Domain:      "authentication",
		AcceptanceCriteria: []string{
			"Users can register with email and password",
			"Users can login with email and password",
			"Password reset functionality via email",
			"Multi-factor authentication support",
			"Session management and logout",
			"Security audit and penetration testing completed",
		},
		ImplementationDetails: []string{
			"Use JWT for session management",
			"Implement OAuth2 for third-party authentication",
			"Password hashing with bcrypt",
			"Rate limiting for authentication endpoints",
			"Comprehensive logging and monitoring",
		},
		Dependencies: []string{
			"Database schema design",
			"Email service integration",
			"Security review approval",
		},
		Timeline: "8-10 weeks",
	}

	// Child Epic data
	childEpicData := IssueData{
		Title:       "Multi-Factor Authentication Sub-Epic",
		Project:     "Auth Team",
		Status:      "backlog",
		Labels:      "epic, security, mfa",
		Parent:      "User Authentication System Epic",
		Description: "Implement multi-factor authentication (MFA) capabilities as part of the broader authentication system. This includes SMS, email, and authenticator app support.",
		AcceptanceCriteria: []string{
			"SMS-based MFA",
			"Email-based MFA",
			"TOTP authenticator app support",
			"Backup codes generation and validation",
			"MFA enrollment and management UI",
		},
		ImplementationDetails: []string{
			"Integration with SMS providers (Twilio, AWS SNS)",
			"TOTP library integration",
			"Secure backup code generation",
			"User-friendly enrollment process",
		},
		Dependencies: []string{
			"Parent epic: User Authentication System Epic",
			"SMS service provider setup",
			"Email service integration",
		},
		Timeline: "4-5 weeks",
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
		AcceptanceCriteria: []string{
			"POST /api/auth/register endpoint created",
			"Email format validation",
			"Password strength validation (min 8 chars, special chars, numbers)",
			"Duplicate email prevention",
			"Email verification workflow",
			"Proper error handling and response codes",
			"Unit tests with >90% coverage",
			"API documentation updated",
		},
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
		Dependencies: []string{
			"Database schema implementation",
			"Email service configuration",
			"Authentication middleware",
		},
		Timeline: "3-4 days",
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
		AcceptanceCriteria: []string{
			"Grafana instance deployed",
			"Database performance metrics",
			"Application error tracking",
			"User activity metrics",
			"Alert rules configured",
			"Dashboard shared with team",
			"Documentation created",
		},
		ImplementationDetails: []string{
			"Deploy Grafana on Kubernetes",
			"Configure Prometheus data sources",
			"Create custom dashboards",
			"Set up Slack/email notifications",
		},
		Dependencies: []string{
			"Prometheus metrics collection",
			"Kubernetes cluster access",
			"Slack webhook configuration",
		},
		Timeline: "2-3 days",
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
		UserStory: UserStory{
			AsA:    "power user of the platform",
			IWant:  "advanced search and filtering capabilities",
			SoThat: "I can quickly find the specific content I'm looking for without browsing through multiple pages",
		},
		AcceptanceCriteria: []string{
			"Full-text search across multiple content types",
			"Multiple filter categories (date, author, tags, type)",
			"Search suggestions and autocomplete",
			"Saved search functionality",
			"Search result ranking and sorting",
			"Search history tracking",
			"Export search results",
			"Search analytics for admins",
		},
		BusinessValue: []string{
			"Improved user engagement and retention",
			"Reduced support tickets for \"can't find content\"",
			"Better content discoverability",
			"Increased user productivity",
		},
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
		Dependencies: []string{
			"Elasticsearch cluster setup",
			"Content indexing pipeline",
			"UI/UX design approval",
			"Performance testing infrastructure",
		},
		SuccessMetrics: []string{
			"Search usage increase by 40%",
			"Average time to find content reduced by 60%",
			"User satisfaction score >4.5/5",
			"Search abandonment rate <10%",
		},
		Timeline: "6-8 weeks",
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
		UserStory: UserStory{
			AsA:    "user typing in the search box",
			IWant:  "to see relevant suggestions and autocomplete options",
			SoThat: "I can quickly find what I'm looking for without typing the full query",
		},
		AcceptanceCriteria: []string{
			"Real-time search suggestions as user types",
			"Autocomplete for search terms (minimum 2 characters)",
			"Popular searches suggestions",
			"Recent searches for logged-in users",
			"Typo tolerance and spell correction",
			"Keyboard navigation support (arrow keys, enter, escape)",
			"Click tracking for suggestion analytics",
			"Configurable suggestion limits",
		},
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
		Dependencies: []string{
			"Parent feature: Advanced Search and Filtering System",
			"Search analytics infrastructure",
			"A/B testing framework",
		},
		SuccessMetrics: []string{
			"70% of searches use autocomplete suggestions",
			"Search completion rate increase by 30%",
			"Average characters typed reduced by 40%",
			"Zero accessibility violations",
		},
		Timeline: "2-3 weeks",
	}

	// Standalone feature
	standaloneFeatureData := IssueData{
		Title:       "Dark Mode Theme Support",
		Project:     "Frontend Team",
		Status:      "ready",
		Labels:      "feature, ui, theme, accessibility",
		Description: "Add dark mode theme support across the entire application to improve user experience in low-light conditions and reduce eye strain.",
		UserStory: UserStory{
			AsA:    "user who works in low-light environments or has visual sensitivity",
			IWant:  "the ability to switch to a dark theme",
			SoThat: "I can use the application comfortably without eye strain",
		},
		AcceptanceCriteria: []string{
			"Dark theme available for all UI components",
			"Theme toggle switch in user preferences",
			"Theme preference persisted across sessions",
			"System theme detection and automatic switching",
			"Smooth transitions between themes",
			"High contrast ratios for accessibility (WCAG AA compliance)",
			"Dark mode support for emails and notifications",
			"Theme-aware images and graphics",
		},
		BusinessValue: []string{
			"Improved user satisfaction and retention",
			"Better accessibility compliance",
			"Modern user interface standards",
			"Reduced eye strain complaints",
		},
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
		Dependencies: []string{
			"Design system color palette update",
			"Asset creation for dark theme variants",
			"Accessibility audit completion",
		},
		SuccessMetrics: []string{
			"30% of users adopt dark mode within 1 month",
			"User satisfaction score increase by 0.3 points",
			"Accessibility complaint reduction by 50%",
			"Theme switch usage >5 times per active user",
		},
		Timeline: "3-4 weeks",
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
