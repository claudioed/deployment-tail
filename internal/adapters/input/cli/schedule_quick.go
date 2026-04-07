package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/claudioed/deployment-tail/api"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newScheduleQuickCmd creates the "schedule quick" command
func newScheduleQuickCmd() *cobra.Command {
	var (
		quickEnvs        []string
		quickNow         bool
		quickIn          int
		quickInHours     int
		quickAt          string
		quickDescription string
		quickRollback    string
		quickGroups      string
	)

	cmd := &cobra.Command{
		Use:     "quick <service-name>",
		Aliases: []string{"q"},
		Short:   "Quickly create a schedule with minimal fields",
		Long: `Create a schedule quickly with just service name, environments, and time.

Examples:
  # Schedule for now
  deployment-tail schedule quick api-service --env staging --now

  # Schedule in 30 minutes
  deployment-tail schedule quick api-service --env staging --in 30

  # Schedule in 2 hours
  deployment-tail schedule quick api-service --env production --in-hours 2

  # Schedule at specific time today
  deployment-tail schedule quick api-service --env production --at 14:30

  # Multiple environments
  deployment-tail schedule quick api-service --env staging --env production --in 30

  # With optional fields
  deployment-tail schedule quick api-service -e staging --in 30 \
    --description "Hotfix deployment" \
    --rollback "kubectl rollout undo"

  # With group assignment (by ID or name)
  deployment-tail schedule quick api-service -e staging --in 30 \
    --groups "550e8400-e29b-41d4-a716-446655440000,660e8400-e29b-41d4-a716-446655440001"
  deployment-tail schedule quick api-service -e staging --in 30 \
    --groups "Project Alpha,Team Backend"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			serviceName := args[0]

			// Validate environments
			if len(quickEnvs) == 0 {
				return fmt.Errorf("at least one environment is required (use --env)")
			}

			// Calculate scheduled time
			scheduledTime, err := calculateScheduledTime(quickNow, quickIn, quickInHours, quickAt)
			if err != nil {
				return fmt.Errorf("failed to calculate scheduled time: %w", err)
			}

			// Get current user from token
			tokenStore := NewTokenStore()
			token, err := tokenStore.LoadToken()
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			// Convert environment strings to API type
			apiEnvironments := make([]api.CreateScheduleRequestEnvironments, len(quickEnvs))
			for i, env := range quickEnvs {
				apiEnvironments[i] = api.CreateScheduleRequestEnvironments(env)
			}

			// Build schedule request with smart defaults
			req := api.CreateScheduleRequest{
				ScheduledAt:  scheduledTime,
				ServiceName:  serviceName,
				Environments: apiEnvironments,
				Owners:       []string{token.Email},
			}

			if quickDescription != "" {
				req.Description = &quickDescription
			}

			if quickRollback != "" {
				req.RollbackPlan = &quickRollback
			}

			// Create schedule
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			schedule, err := client.CreateSchedule(ctx, req)
			if err != nil {
				return handleCommandError(err)
			}

			// If groups specified, resolve and assign them
			var assignedGroupNames []string
			if quickGroups != "" {
				groupIDs, groupNames, err := resolveGroupIdentifiers(ctx, client, quickGroups, token.Email)
				if err != nil {
					// Rollback: delete the just-created schedule
					rollbackErr := client.DeleteSchedule(ctx, schedule.Id.String())
					if rollbackErr != nil {
						fmt.Fprintf(os.Stderr, "\n⚠️  Warning: Failed to delete schedule during rollback.\n")
						fmt.Fprintf(os.Stderr, "Orphaned schedule ID: %s\n", schedule.Id)
						fmt.Fprintf(os.Stderr, "Please delete it manually with: deployment-tail schedule delete %s\n\n", schedule.Id)
					}
					return fmt.Errorf("group resolution failed: %w", err)
				}

				// Assign schedule to groups
				err = client.AssignScheduleToGroups(ctx, schedule.Id.String(), groupIDs)
				if err != nil {
					// Rollback: delete the just-created schedule
					rollbackErr := client.DeleteSchedule(ctx, schedule.Id.String())
					if rollbackErr != nil {
						fmt.Fprintf(os.Stderr, "\n⚠️  Warning: Failed to delete schedule during rollback.\n")
						fmt.Fprintf(os.Stderr, "Orphaned schedule ID: %s\n", schedule.Id)
						fmt.Fprintf(os.Stderr, "Please delete it manually with: deployment-tail schedule delete %s\n\n", schedule.Id)
					}
					return handleCommandError(fmt.Errorf("group assignment failed: %w", err))
				}

				assignedGroupNames = groupNames
			}

			// Display success
			if outputJSON {
				return printJSON(schedule)
			}

			fmt.Printf("✓ Schedule created successfully!\n\n")
			fmt.Printf("ID:           %s\n", schedule.Id)
			fmt.Printf("Service:      %s\n", schedule.ServiceName)
			fmt.Printf("Environments: %s\n", strings.Join(envSliceToStringSlice(schedule.Environments), ", "))
			fmt.Printf("Scheduled:    %s (%s)\n",
				schedule.ScheduledAt.Local().Format("2006-01-02 15:04:05"),
				formatRelativeTime(scheduledTime))
			fmt.Printf("Owner:        %s\n", strings.Join(schedule.Owners, ", "))
			fmt.Printf("Status:       %s\n", schedule.Status)
			if len(assignedGroupNames) > 0 {
				fmt.Printf("Groups:       %s\n", strings.Join(assignedGroupNames, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&quickEnvs, "env", "e", []string{}, "Target environment(s) (can specify multiple)")
	cmd.Flags().BoolVar(&quickNow, "now", false, "Schedule for right now")
	cmd.Flags().IntVar(&quickIn, "in", 0, "Schedule in N minutes from now")
	cmd.Flags().IntVar(&quickInHours, "in-hours", 0, "Schedule in N hours from now")
	cmd.Flags().StringVar(&quickAt, "at", "", "Schedule at specific time today (HH:MM format)")
	cmd.Flags().StringVar(&quickDescription, "description", "", "Optional deployment description")
	cmd.Flags().StringVar(&quickRollback, "rollback", "", "Optional rollback plan")
	cmd.Flags().StringVar(&quickGroups, "groups", "", "Optional comma-separated group IDs or names to assign (e.g., 'id1,id2' or 'Project Alpha,Team Backend')")

	cmd.MarkFlagRequired("env")

	return cmd
}

func calculateScheduledTime(quickNow bool, quickIn int, quickInHours int, quickAt string) (time.Time, error) {
	now := time.Now()

	// Count how many time flags are set
	timeFlags := 0
	if quickNow {
		timeFlags++
	}
	if quickIn > 0 {
		timeFlags++
	}
	if quickInHours > 0 {
		timeFlags++
	}
	if quickAt != "" {
		timeFlags++
	}

	// Validate only one time flag is set
	if timeFlags == 0 {
		// Default to now if no time flag specified
		return now, nil
	}
	if timeFlags > 1 {
		return time.Time{}, fmt.Errorf("specify only one time option (--now, --in, --in-hours, or --at)")
	}

	// Calculate based on flag
	if quickNow {
		return now, nil
	}

	if quickIn > 0 {
		return now.Add(time.Duration(quickIn) * time.Minute), nil
	}

	if quickInHours > 0 {
		return now.Add(time.Duration(quickInHours) * time.Hour), nil
	}

	if quickAt != "" {
		return parseTimeToday(quickAt)
	}

	return now, nil
}

func parseTimeToday(timeStr string) (time.Time, error) {
	// Parse HH:MM format
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format, expected HH:MM")
	}

	var hour, minute int
	if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute); err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}

	if hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("hour must be between 0 and 23")
	}

	if minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("minute must be between 0 and 59")
	}

	now := time.Now()
	scheduledTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	return scheduledTime, nil
}

func formatRelativeTime(t time.Time) string {
	duration := time.Until(t)

	if duration < 0 {
		return "in the past"
	}

	if duration < time.Minute {
		return "now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("in %d minute%s", minutes, plural(minutes))
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		remainingMinutes := int(duration.Minutes()) % 60
		if remainingMinutes == 0 {
			return fmt.Sprintf("in %d hour%s", hours, plural(hours))
		}
		return fmt.Sprintf("in %dh %dm", hours, remainingMinutes)
	}

	days := int(duration.Hours() / 24)
	return fmt.Sprintf("in %d day%s", days, plural(days))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func envSliceToStringSlice(envs []api.ScheduleEnvironments) []string {
	result := make([]string, len(envs))
	for i, env := range envs {
		result[i] = string(env)
	}
	return result
}

func templateEnvSliceToStringSlice(envs []api.TemplateEnvironments) []string {
	result := make([]string, len(envs))
	for i, env := range envs {
		result[i] = string(env)
	}
	return result
}

func templateEnvSliceToScheduleEnvSlice(envs []api.TemplateEnvironments) []api.ScheduleEnvironments {
	result := make([]api.ScheduleEnvironments, len(envs))
	for i, env := range envs {
		result[i] = api.ScheduleEnvironments(env)
	}
	return result
}

// resolveGroupIdentifiers parses comma-separated group identifiers (IDs or names)
// and returns slice of group IDs and names. Detects UUIDs as IDs, resolves names via API.
func resolveGroupIdentifiers(ctx context.Context, client *APIClient, groupsFlag, owner string) ([]string, []string, error) {
	// Parse comma-separated values
	identifiers := strings.Split(groupsFlag, ",")
	var groupIDs []string
	var groupNames []string
	var namesToResolve []string

	// First pass: separate UUIDs from names
	for _, identifier := range identifiers {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			continue
		}

		// Try to parse as UUID
		if _, err := uuid.Parse(identifier); err == nil {
			// It's a valid UUID, use as ID
			groupIDs = append(groupIDs, identifier)
			groupNames = append(groupNames, identifier) // Use ID as placeholder name
		} else {
			// It's a name, needs resolution
			namesToResolve = append(namesToResolve, identifier)
		}
	}

	// If there are names to resolve, fetch groups from API
	if len(namesToResolve) > 0 {
		groups, err := client.ListGroups(ctx, owner)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch groups for name resolution: %w", err)
		}

		// Build a map of group names to IDs
		nameToGroupMap := make(map[string][]api.Group)
		for _, group := range groups {
			nameToGroupMap[group.Name] = append(nameToGroupMap[group.Name], group)
		}

		// Resolve each name
		for _, name := range namesToResolve {
			matches := nameToGroupMap[name]
			if len(matches) == 0 {
				// Group not found - list available groups
				return nil, nil, fmt.Errorf("group not found: '%s'\n\nAvailable groups:\n%s",
					name, formatAvailableGroups(groups))
			}
			if len(matches) > 1 {
				// Ambiguous name - list matches
				return nil, nil, fmt.Errorf("ambiguous group name: '%s' matches multiple groups\n\nMatching groups:\n%s\n\nPlease use group ID instead",
					name, formatMatchingGroups(matches))
			}

			// Exactly one match - use it
			groupIDs = append(groupIDs, matches[0].Id.String())
			groupNames = append(groupNames, matches[0].Name)
		}
	}

	if len(groupIDs) == 0 {
		return nil, nil, fmt.Errorf("no valid group identifiers provided")
	}

	return groupIDs, groupNames, nil
}

// formatAvailableGroups formats the list of available groups for error messages
func formatAvailableGroups(groups []api.Group) string {
	if len(groups) == 0 {
		return "  (no groups available)"
	}

	var builder strings.Builder
	for _, g := range groups {
		builder.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", g.Name, g.Id))
	}
	return builder.String()
}

// formatMatchingGroups formats the list of matching groups for error messages
func formatMatchingGroups(groups []api.Group) string {
	var builder strings.Builder
	for _, g := range groups {
		builder.WriteString(fmt.Sprintf("  - %s (ID: %s, Owner: %s)\n", g.Name, g.Id, g.Owner))
	}
	return builder.String()
}
