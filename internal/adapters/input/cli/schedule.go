package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/claudioed/deployment-tail/api"
	"github.com/spf13/cobra"
)

var forceLogin bool

// NewScheduleCmd creates the schedule command group
func NewScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage deployment schedules (requires authentication)",
		Long:  "Create, view, update, and delete deployment schedules. All commands require authentication via 'deployment-tail auth login'.",
	}

	cmd.PersistentFlags().BoolVar(&forceLogin, "force-login", false, "Force re-authentication even if token exists")

	cmd.AddCommand(
		newScheduleCreateCmd(),
		newScheduleGetCmd(),
		newScheduleListCmd(),
		newScheduleUpdateCmd(),
		newScheduleDeleteCmd(),
		newScheduleApproveCmd(),
		newScheduleDenyCmd(),
	)

	return cmd
}

// checkAuthentication verifies the user is authenticated before executing a command
func checkAuthentication() error {
	tokenStore := NewTokenStore()

	// If force-login flag is set, prompt user to re-authenticate
	if forceLogin {
		return &AuthenticationError{
			Message: "Force login requested. Please run 'deployment-tail auth login' to re-authenticate",
		}
	}

	// Check if user has a valid token
	if !tokenStore.IsAuthenticated() {
		return &AuthenticationError{
			Message: "Not authenticated. Please run 'deployment-tail auth login' to authenticate",
		}
	}

	return nil
}

// handleCommandError provides user-friendly error messages for common error types
func handleCommandError(err error) error {
	if err == nil {
		return nil
	}

	// Check for authentication error
	if authErr, ok := err.(*AuthenticationError); ok {
		fmt.Fprintf(os.Stderr, "\n❌ Authentication Error\n\n")
		fmt.Fprintf(os.Stderr, "%s\n\n", authErr.Message)
		fmt.Fprintf(os.Stderr, "To authenticate, run:\n")
		fmt.Fprintf(os.Stderr, "  deployment-tail auth login\n\n")
		return fmt.Errorf("authentication required")
	}

	// Check for permission error
	if permErr, ok := err.(*PermissionError); ok {
		fmt.Fprintf(os.Stderr, "\n❌ Permission Denied\n\n")
		fmt.Fprintf(os.Stderr, "%s\n\n", permErr.Message)
		fmt.Fprintf(os.Stderr, "Your current role may not allow this operation.\n")
		fmt.Fprintf(os.Stderr, "To check your role, run:\n")
		fmt.Fprintf(os.Stderr, "  deployment-tail auth status\n\n")
		return fmt.Errorf("permission denied")
	}

	// Return original error for other cases
	return err
}

// newScheduleCreateCmd creates the "schedule create" command
func newScheduleCreateCmd() *cobra.Command {
	var (
		scheduledAt  string
		service      string
		environments []string
		description  string
		owners       []string
		rollbackPlan string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new deployment schedule (requires 'deployer' or 'admin' role)",
		Long: `Create a new deployment schedule with multiple owners and environments.

Authentication: This command requires authentication and 'deployer' or 'admin' role.

Examples:
  deployment-tail schedule create --date 2026-04-01T14:00:00Z --service api-service --owner john.doe --owner jane.smith --env production --env staging
  deployment-tail schedule create --date 2026-04-01T14:00:00Z --service web-app --owner alice --env production --description "Deploy v2.0"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			// Parse scheduled time
			scheduledTime, err := time.Parse(time.RFC3339, scheduledAt)
			if err != nil {
				return fmt.Errorf("invalid date format (use RFC3339, e.g., 2026-04-01T14:00:00Z): %w", err)
			}

			// Convert environment strings to API type
			apiEnvironments := make([]api.CreateScheduleRequestEnvironments, len(environments))
			for i, env := range environments {
				apiEnvironments[i] = api.CreateScheduleRequestEnvironments(env)
			}

			req := api.CreateScheduleRequest{
				ScheduledAt:  scheduledTime,
				ServiceName:  service,
				Environments: apiEnvironments,
				Owners:       owners,
			}

			if description != "" {
				req.Description = &description
			}

			if rollbackPlan != "" {
				req.RollbackPlan = &rollbackPlan
			}

			schedule, err := client.CreateSchedule(ctx, req)
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("Schedule created successfully!\n")
			fmt.Printf("ID: %s\n", schedule.Id)
			fmt.Printf("Scheduled At: %s\n", schedule.ScheduledAt.Format(time.RFC3339))
			fmt.Printf("Service: %s\n", schedule.ServiceName)
			fmt.Printf("Environments: %v\n", schedule.Environments)
			fmt.Printf("Owners: %v\n", schedule.Owners)
			fmt.Printf("Status: %s\n", schedule.Status)
			if schedule.Description != nil {
				fmt.Printf("Description: %s\n", *schedule.Description)
			}
			if schedule.RollbackPlan != nil {
				fmt.Printf("Rollback Plan: %s\n", *schedule.RollbackPlan)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&scheduledAt, "date", "", "Scheduled date/time (RFC3339 format, e.g., 2026-04-01T14:00:00Z)")
	cmd.Flags().StringVar(&service, "service", "", "Service name")
	cmd.Flags().StringSliceVar(&environments, "env", []string{}, "Environments (can specify multiple: --env production --env staging)")
	cmd.Flags().StringVar(&description, "description", "", "Optional description")
	cmd.Flags().StringSliceVar(&owners, "owner", []string{}, "Owners (can specify multiple: --owner john.doe --owner jane.smith)")
	cmd.Flags().StringVar(&rollbackPlan, "rollback-plan", "", "Optional rollback plan")

	cmd.MarkFlagRequired("date")
	cmd.MarkFlagRequired("service")
	cmd.MarkFlagRequired("env")
	cmd.MarkFlagRequired("owner")

	return cmd
}

// newScheduleGetCmd creates the "schedule get" command
func newScheduleGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a schedule by ID (requires authentication)",
		Long:  "Retrieve detailed information about a specific deployment schedule.\n\nAuthentication: This command requires authentication.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.GetSchedule(ctx, args[0])
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("ID: %s\n", schedule.Id)
			fmt.Printf("Scheduled At: %s\n", schedule.ScheduledAt.Format(time.RFC3339))
			fmt.Printf("Service: %s\n", schedule.ServiceName)
			fmt.Printf("Environments: %v\n", schedule.Environments)
			fmt.Printf("Owners: %v\n", schedule.Owners)
			fmt.Printf("Status: %s\n", schedule.Status)
			if schedule.Description != nil {
				fmt.Printf("Description: %s\n", *schedule.Description)
			}
			if schedule.RollbackPlan != nil {
				fmt.Printf("Rollback Plan:\n%s\n", *schedule.RollbackPlan)
			}
			fmt.Printf("Created At: %s\n", schedule.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Updated At: %s\n", schedule.UpdatedAt.Format(time.RFC3339))

			return nil
		},
	}

	return cmd
}

// newScheduleListCmd creates the "schedule list" command
func newScheduleListCmd() *cobra.Command {
	var (
		fromStr      string
		toStr        string
		environments []string
		owners       []string
		status       string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployment schedules (requires authentication)",
		Long: `List deployment schedules with optional filters.

Authentication: This command requires authentication.

Examples:
  deployment-tail schedule list
  deployment-tail schedule list --env production --env staging
  deployment-tail schedule list --owner john.doe --owner jane.smith
  deployment-tail schedule list --status approved --env production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			var from, to *time.Time
			var statusPtr *string

			if fromStr != "" {
				t, err := time.Parse(time.RFC3339, fromStr)
				if err != nil {
					return fmt.Errorf("invalid from date: %w", err)
				}
				from = &t
			}

			if toStr != "" {
				t, err := time.Parse(time.RFC3339, toStr)
				if err != nil {
					return fmt.Errorf("invalid to date: %w", err)
				}
				to = &t
			}

			if status != "" {
				statusPtr = &status
			}

			schedules, err := client.ListSchedules(ctx, from, to, environments, owners, statusPtr)
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(schedules)
			}

			if len(schedules) == 0 {
				fmt.Println("No schedules found")
				return nil
			}

			printTable(schedules)

			return nil
		},
	}

	cmd.Flags().StringVar(&fromStr, "from", "", "Start date filter (RFC3339 format)")
	cmd.Flags().StringVar(&toStr, "to", "", "End date filter (RFC3339 format)")
	cmd.Flags().StringSliceVar(&environments, "env", []string{}, "Environment filters (can specify multiple)")
	cmd.Flags().StringSliceVar(&owners, "owner", []string{}, "Owner filters (can specify multiple)")
	cmd.Flags().StringVar(&status, "status", "", "Status filter (created, approved, denied)")

	return cmd
}

// newScheduleUpdateCmd creates the "schedule update" command
func newScheduleUpdateCmd() *cobra.Command {
	var (
		scheduledAt  string
		service      string
		environments []string
		owners       []string
		description  string
		rollbackPlan string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a deployment schedule (requires 'deployer' or 'admin' role)",
		Long: `Update a deployment schedule. Owners and environments replace the entire list.

Authentication: This command requires authentication and 'deployer' or 'admin' role.
Note: Deployers can only update schedules they created. Admins can update any schedule.

Examples:
  deployment-tail schedule update <id> --env production --env staging
  deployment-tail schedule update <id> --owner alice --owner bob
  deployment-tail schedule update <id> --service new-service --description "Updated deployment"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			req := api.UpdateScheduleRequest{}

			if scheduledAt != "" {
				t, err := time.Parse(time.RFC3339, scheduledAt)
				if err != nil {
					return fmt.Errorf("invalid date format: %w", err)
				}
				req.ScheduledAt = &t
			}

			if service != "" {
				req.ServiceName = &service
			}

			if len(environments) > 0 {
				apiEnvironments := make([]api.UpdateScheduleRequestEnvironments, len(environments))
				for i, env := range environments {
					apiEnvironments[i] = api.UpdateScheduleRequestEnvironments(env)
				}
				req.Environments = &apiEnvironments
			}

			if len(owners) > 0 {
				req.Owners = &owners
			}

			if description != "" {
				req.Description = &description
			}

			if rollbackPlan != "" {
				req.RollbackPlan = &rollbackPlan
			}

			schedule, err := client.UpdateSchedule(ctx, args[0], req)
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("Schedule updated successfully!\n")
			fmt.Printf("ID: %s\n", schedule.Id)
			fmt.Printf("Scheduled At: %s\n", schedule.ScheduledAt.Format(time.RFC3339))
			fmt.Printf("Service: %s\n", schedule.ServiceName)
			fmt.Printf("Environments: %v\n", schedule.Environments)
			fmt.Printf("Owners: %v\n", schedule.Owners)
			fmt.Printf("Status: %s\n", schedule.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&scheduledAt, "date", "", "New scheduled date/time")
	cmd.Flags().StringVar(&service, "service", "", "New service name")
	cmd.Flags().StringSliceVar(&environments, "env", []string{}, "New environments (replaces entire list)")
	cmd.Flags().StringSliceVar(&owners, "owner", []string{}, "New owners (replaces entire list)")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&rollbackPlan, "rollback-plan", "", "New rollback plan")

	return cmd
}

// newScheduleDeleteCmd creates the "schedule delete" command
func newScheduleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a deployment schedule (requires 'deployer' or 'admin' role)",
		Long: `Delete a deployment schedule.

Authentication: This command requires authentication and 'deployer' or 'admin' role.
Note: Deployers can only delete schedules they created. Admins can delete any schedule.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			if err := client.DeleteSchedule(ctx, args[0]); err != nil {
				return handleCommandError(err)
			}

			fmt.Printf("Schedule %s deleted successfully\n", args[0])

			return nil
		},
	}

	return cmd
}

// newScheduleApproveCmd creates the "schedule approve" command
func newScheduleApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a deployment schedule (requires 'admin' role)",
		Long: `Approve a deployment schedule.

Authentication: This command requires authentication and 'admin' role.
Only admins can approve schedules.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.ApproveSchedule(ctx, args[0])
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("Schedule %s approved successfully\n", args[0])
			fmt.Printf("Status: %s\n", schedule.Status)

			return nil
		},
	}

	return cmd
}

// newScheduleDenyCmd creates the "schedule deny" command
func newScheduleDenyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deny <id>",
		Short: "Deny a deployment schedule (requires 'admin' role)",
		Long: `Deny a deployment schedule.

Authentication: This command requires authentication and 'admin' role.
Only admins can deny schedules.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.DenySchedule(ctx, args[0])
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("Schedule %s denied successfully\n", args[0])
			fmt.Printf("Status: %s\n", schedule.Status)

			return nil
		},
	}

	return cmd
}
