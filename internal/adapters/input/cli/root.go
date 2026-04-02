package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiEndpoint string
	outputJSON  bool
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "deployment-tail",
		Short: "Deployment schedule management tool",
		Long:  "A CLI tool for managing deployment schedules",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiEndpoint, "api", getAPIEndpoint(), "API endpoint URL")
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output in JSON format")

	// Add subcommands
	rootCmd.AddCommand(NewScheduleCmd())

	return rootCmd
}

// Execute runs the CLI
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// getAPIEndpoint returns the API endpoint from environment or default
func getAPIEndpoint() string {
	if endpoint := os.Getenv("DEPLOYMENT_TAIL_API"); endpoint != "" {
		return endpoint
	}
	return "http://localhost:8080"
}
