package cli

import (
	"context"
	"fmt"

	"github.com/claudioed/deployment-tail/api"
	"github.com/spf13/cobra"
)

// NewGroupCmd creates the group command group
func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage deployment schedule groups (requires authentication)",
		Long:  "Create, view, update, delete, and favorite deployment schedule groups. All commands require authentication via 'deployment-tail auth login'.",
	}

	cmd.PersistentFlags().BoolVar(&forceLogin, "force-login", false, "Force re-authentication even if token exists")

	cmd.AddCommand(
		newGroupListCmd(),
		newGroupFavoriteCmd(),
		newGroupUnfavoriteCmd(),
	)

	return cmd
}

// newGroupListCmd creates the "group list" command
func newGroupListCmd() *cobra.Command {
	var (
		owner         string
		favoritesOnly bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployment schedule groups (requires authentication)",
		Long: `List deployment schedule groups with optional filters.

Authentication: This command requires authentication.

Examples:
  deployment-tail group list --owner ops-team
  deployment-tail group list --owner ops-team --favorites-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			if owner == "" {
				return fmt.Errorf("--owner flag is required")
			}

			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			groups, err := client.ListGroups(ctx, owner)
			if err != nil {
				return handleCommandError(err)
			}

			if outputJSON {
				return printJSON(groups)
			}

			if len(groups) == 0 {
				fmt.Println("No groups found")
				return nil
			}

			// Filter favorites if requested
			filteredGroups := groups
			if favoritesOnly {
				filteredGroups = make([]api.Group, 0)
				for _, g := range groups {
					if g.IsFavorite != nil && *g.IsFavorite {
						filteredGroups = append(filteredGroups, g)
					}
				}

				if len(filteredGroups) == 0 {
					fmt.Println("No favorite groups found")
					return nil
				}
			}

			// Print groups with favorite indicator
			fmt.Printf("Found %d group(s):\n\n", len(filteredGroups))
			for _, g := range filteredGroups {
				// Determine if favorite
				isFavorite := false
				if g.IsFavorite != nil && *g.IsFavorite {
					isFavorite = true
				}

				// Print with star indicator
				starIcon := " "
				if isFavorite {
					starIcon = "★"
				}

				fmt.Printf("%s ID: %v\n", starIcon, g.Id)
				fmt.Printf("  Name: %v\n", g.Name)
				fmt.Printf("  Owner: %v\n", g.Owner)
				if g.Description != nil {
					fmt.Printf("  Description: %v\n", *g.Description)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Filter groups by owner (required)")
	cmd.Flags().BoolVar(&favoritesOnly, "favorites-only", false, "Show only favorited groups")
	cmd.MarkFlagRequired("owner")

	return cmd
}

// newGroupFavoriteCmd creates the "group favorite" command
func newGroupFavoriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "favorite <group-id>",
		Short: "Mark a group as favorite (requires authentication)",
		Long: `Mark a deployment schedule group as favorite.

Authentication: This command requires authentication.

Examples:
  deployment-tail group favorite 550e8400-e29b-41d4-a716-446655440000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			groupID := args[0]
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			err := client.FavoriteGroup(ctx, groupID)
			if err != nil {
				return handleCommandError(err)
			}

			fmt.Printf("✓ Group %s marked as favorite\n", groupID)
			return nil
		},
	}

	return cmd
}

// newGroupUnfavoriteCmd creates the "group unfavorite" command
func newGroupUnfavoriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfavorite <group-id>",
		Short: "Remove favorite status from a group (requires authentication)",
		Long: `Remove favorite status from a deployment schedule group.

Authentication: This command requires authentication.

Examples:
  deployment-tail group unfavorite 550e8400-e29b-41d4-a716-446655440000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check authentication before proceeding
			if err := checkAuthentication(); err != nil {
				return handleCommandError(err)
			}

			groupID := args[0]
			ctx := context.Background()
			client := NewAPIClient(apiEndpoint)

			err := client.UnfavoriteGroup(ctx, groupID)
			if err != nil {
				return handleCommandError(err)
			}

			fmt.Printf("✓ Group %s removed from favorites\n", groupID)
			return nil
		},
	}

	return cmd
}
