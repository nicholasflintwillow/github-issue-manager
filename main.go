package main

import (
	"github-issue-manager/cmd/create"
	"github-issue-manager/cmd/examples"
	"github-issue-manager/cmd/list"
	"github-issue-manager/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	logLevel   string
	jsonFormat bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "github-issue-manager",
		Short: "A CLI tool to create GitHub issues",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize logger with flags
			logger.Init(logger.LogLevel(logLevel), jsonFormat)
		},
	}

	// Add persistent flags for logging
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVar(&jsonFormat, "log-json", false, "Output logs in JSON format")

	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(create.Cmd)
	rootCmd.AddCommand(examples.Cmd)
	rootCmd.Execute()
}
