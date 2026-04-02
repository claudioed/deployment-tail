package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/claudioed/deployment-tail/api"
	"github.com/spf13/cobra"
)

// NewScheduleCmd creates the schedule command group
func NewScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage deployment schedules",
		Long:  "Create, view, update, and delete deployment schedules",
	}

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

// newScheduleCreateCmd creates the "schedule create" command
func newScheduleCreateCmd() *cobra.Command {
	var (
		scheduledAt  string
		service      string
		environment  string
		description  string
		owner        string
		rollbackPlan string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new deployment schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			// Parse scheduled time
			scheduledTime, err := time.Parse(time.RFC3339, scheduledAt)
			if err != nil {
				return fmt.Errorf("invalid date format (use RFC3339, e.g., 2026-04-01T14:00:00Z): %w", err)
			}

			req := api.CreateScheduleRequest{
				ScheduledAt: scheduledTime,
				ServiceName: service,
				Environment: api.CreateScheduleRequestEnvironment(environment),
				Owner:       owner,
			}

			if description != "" {
				req.Description = &description
			}

			if rollbackPlan != "" {
				req.RollbackPlan = &rollbackPlan
			}

			schedule, err := client.CreateSchedule(ctx, req)
			if err != nil {
				return err
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("Schedule created successfully!\n")
			fmt.Printf("ID: %s\n", schedule.Id)
			fmt.Printf("Scheduled At: %s\n", schedule.ScheduledAt.Format(time.RFC3339))
			fmt.Printf("Service: %s\n", schedule.ServiceName)
			fmt.Printf("Environment: %s\n", schedule.Environment)
			fmt.Printf("Owner: %s\n", schedule.Owner)
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
	cmd.Flags().StringVar(&environment, "env", "", "Environment (production, staging, development)")
	cmd.Flags().StringVar(&description, "description", "", "Optional description")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner of the schedule")
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
		Short: "Get a schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.GetSchedule(ctx, args[0])
			if err != nil {
				return err
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("ID: %s\n", schedule.Id)
			fmt.Printf("Scheduled At: %s\n", schedule.ScheduledAt.Format(time.RFC3339))
			fmt.Printf("Service: %s\n", schedule.ServiceName)
			fmt.Printf("Environment: %s\n", schedule.Environment)
			fmt.Printf("Owner: %s\n", schedule.Owner)
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
		fromStr string
		toStr   string
		env     string
		owner   string
		status  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployment schedules",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			var from, to *time.Time
			var environment *string
			var ownerPtr *string
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

			if env != "" {
				environment = &env
			}

			if owner != "" {
				ownerPtr = &owner
			}

			if status != "" {
				statusPtr = &status
			}

			schedules, err := client.ListSchedules(ctx, from, to, environment, ownerPtr, statusPtr)
			if err != nil {
				return err
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
	cmd.Flags().StringVar(&env, "env", "", "Environment filter")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner filter")
	cmd.Flags().StringVar(&status, "status", "", "Status filter (created, approved, denied)")

	return cmd
}

// newScheduleUpdateCmd creates the "schedule update" command
func newScheduleUpdateCmd() *cobra.Command {
	var (
		scheduledAt  string
		service      string
		environment  string
		description  string
		rollbackPlan string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a deployment schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if environment != "" {
				env := api.UpdateScheduleRequestEnvironment(environment)
				req.Environment = &env
			}

			if description != "" {
				req.Description = &description
			}

			if rollbackPlan != "" {
				req.RollbackPlan = &rollbackPlan
			}

			schedule, err := client.UpdateSchedule(ctx, args[0], req)
			if err != nil {
				return err
			}

			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("Schedule updated successfully!\n")
			fmt.Printf("ID: %s\n", schedule.Id)
			fmt.Printf("Scheduled At: %s\n", schedule.ScheduledAt.Format(time.RFC3339))
			fmt.Printf("Service: %s\n", schedule.ServiceName)
			fmt.Printf("Environment: %s\n", schedule.Environment)
			fmt.Printf("Status: %s\n", schedule.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&scheduledAt, "date", "", "New scheduled date/time")
	cmd.Flags().StringVar(&service, "service", "", "New service name")
	cmd.Flags().StringVar(&environment, "env", "", "New environment")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&rollbackPlan, "rollback-plan", "", "New rollback plan")

	return cmd
}

// newScheduleDeleteCmd creates the "schedule delete" command
func newScheduleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a deployment schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			if err := client.DeleteSchedule(ctx, args[0]); err != nil {
				return err
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
		Short: "Approve a deployment schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.ApproveSchedule(ctx, args[0])
			if err != nil {
				return err
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
		Short: "Deny a deployment schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.DenySchedule(ctx, args[0])
			if err != nil {
				return err
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
