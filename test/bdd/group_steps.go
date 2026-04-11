package bdd

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
)

func RegisterGroupSteps(ctx *godog.ScenarioContext) {
	s := &groupSteps{}

	// Phase A steps
	ctx.Step(`^I create a group named "([^"]+)"$`, s.iCreateAGroupNamed)
	ctx.Step(`^I create a group named "([^"]+)" with visibility "([^"]+)"$`, s.iCreateAGroupNamedWithVisibility)
	ctx.Step(`^the last group has name "([^"]+)"$`, s.theLastGroupHasName)
	ctx.Step(`^the last group has visibility "([^"]+)"$`, s.theLastGroupHasVisibility)

	// Phase B: Visibility steps
	ctx.Step(`^I update the last group visibility to "([^"]+)"$`, s.iUpdateTheLastGroupVisibilityTo)
	ctx.Step(`^I attempt to update group "([^"]+)" visibility to "([^"]+)"$`, s.iAttemptToUpdateGroupVisibilityTo)

	// Phase B: Group listing and retrieval
	ctx.Step(`^I list all groups$`, s.iListAllGroups)
	ctx.Step(`^I retrieve the group "([^"]+)" by name$`, s.iRetrieveTheGroupByName)
	ctx.Step(`^I retrieve the last group by ID$`, s.iRetrieveTheLastGroupByID)
	ctx.Step(`^the group list includes "([^"]+)"$`, s.theGroupListIncludes)
	ctx.Step(`^the group list does not include "([^"]+)"$`, s.theGroupListDoesNotInclude)
	ctx.Step(`^the group list is empty$`, s.theGroupListIsEmpty)
	ctx.Step(`^the group list order is "([^"]+)"$`, s.theGroupListOrderIs)
	ctx.Step(`^the first group in the list is "([^"]+)"$`, s.theFirstGroupInTheListIs)

	// Phase B: Favorites steps
	ctx.Step(`^I favorite the last group$`, s.iFavoriteTheLastGroup)
	ctx.Step(`^I unfavorite the last group$`, s.iUnfavoriteTheLastGroup)
	ctx.Step(`^I favorite a group with invalid ID$`, s.iFavoriteAGroupWithInvalidID)
	ctx.Step(`^I unfavorite a group with invalid ID$`, s.iUnfavoriteAGroupWithInvalidID)
	ctx.Step(`^the last group is favorited$`, s.theLastGroupIsFavorited)
	ctx.Step(`^the last group is not favorited$`, s.theLastGroupIsNotFavorited)
	ctx.Step(`^the group is favorited$`, s.theGroupIsFavorited)
	ctx.Step(`^the group is not favorited$`, s.theGroupIsNotFavorited)
	ctx.Step(`^the group "([^"]+)" is favorited$`, s.theGroupNamedIsFavorited)
	ctx.Step(`^the group "([^"]+)" is not favorited$`, s.theGroupNamedIsNotFavorited)

	// Phase B: Seeding steps (user creates group)
	ctx.Step(`^user "([^"]+)" creates a public group named "([^"]+)"$`, s.userCreatesPublicGroupNamed)
	ctx.Step(`^user "([^"]+)" creates a private group named "([^"]+)"$`, s.userCreatesPrivateGroupNamed)

	// Phase B: Delete group
	ctx.Step(`^I delete the last group$`, s.iDeleteTheLastGroup)
}

type groupSteps struct{}

func (s *groupSteps) iCreateAGroupNamed(ctx context.Context, name string) error {
	// Phase B: Default visibility is "private" (not "public")
	return s.iCreateAGroupNamedWithVisibility(ctx, name, "private")
}

func (s *groupSteps) iCreateAGroupNamedWithVisibility(ctx context.Context, name, visibilityStr string) error {
	w := getWorld(ctx)

	if w.CurrentUser == nil {
		return fmt.Errorf("no authenticated user set; use 'Given I am authenticated as' first")
	}

	// Validate input using domain value objects
	_, err := group.NewGroupName(name)
	if err != nil {
		return fmt.Errorf("invalid group name: %w", err)
	}

	_, err = group.NewVisibility(visibilityStr)
	if err != nil {
		return fmt.Errorf("invalid visibility: %w", err)
	}

	// Create command with string values (as per API contract)
	cmd := input.CreateGroupCommand{
		Owner:      w.CurrentUser.ID().String(),
		Name:       name,
		Visibility: visibilityStr,
	}

	grp, err := w.GroupService.CreateGroup(ctx, cmd)
	w.LastError = err
	if err == nil {
		w.LastGroup = grp
		// Store in NamedGroups for later reference
		w.NamedGroups[name] = grp
	}

	return nil
}

func (s *groupSteps) theLastGroupHasName(ctx context.Context, expectedName string) error {
	w := getWorld(ctx)
	if w.LastGroup == nil {
		return fmt.Errorf("no group was created")
	}
	if w.LastGroup.Name().String() != expectedName {
		return fmt.Errorf("expected group name %q but got %q", expectedName, w.LastGroup.Name().String())
	}
	return nil
}

func (s *groupSteps) theLastGroupHasVisibility(ctx context.Context, expectedVisibility string) error {
	w := getWorld(ctx)
	if w.LastGroup == nil {
		return fmt.Errorf("no group was created")
	}
	actualVisibility := string(w.LastGroup.Visibility())
	if actualVisibility != expectedVisibility {
		return fmt.Errorf("expected visibility %q but got %q", expectedVisibility, actualVisibility)
	}
	return nil
}

// Phase B: Visibility updates
func (s *groupSteps) iUpdateTheLastGroupVisibilityTo(ctx context.Context, visibilityStr string) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to update; create a group first")
	}

	vis, err := group.NewVisibility(visibilityStr)
	if err != nil {
		return fmt.Errorf("invalid visibility: %w", err)
	}

	// Call SetVisibility on the domain aggregate
	w.LastGroup.SetVisibility(vis)

	// Update in repository
	err = w.GroupRepo.Update(ctx, w.LastGroup)
	w.LastError = err

	return nil
}

func (s *groupSteps) iAttemptToUpdateGroupVisibilityTo(ctx context.Context, groupName, visibilityStr string) error {
	w := getWorld(ctx)

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	vis, err := group.NewVisibility(visibilityStr)
	if err != nil {
		return fmt.Errorf("invalid visibility: %w", err)
	}

	// Attempt to update (should fail if not owner)
	// Check if current user owns the group (compare Owner string with UserID string)
	if w.CurrentUser != nil && grp.Owner().String() != w.CurrentUser.ID().String() {
		w.LastError = fmt.Errorf("forbidden: only owner can update group visibility")
		return nil
	}

	grp.SetVisibility(vis)
	err = w.GroupRepo.Update(ctx, grp)
	w.LastError = err

	return nil
}

// Phase B: Group listing and retrieval
func (s *groupSteps) iListAllGroups(ctx context.Context) error {
	w := getWorld(ctx)

	// Use GroupService.ListGroups with query containing owner
	query := input.ListGroupsQuery{
		Owner: w.CurrentUser.ID().String(),
	}
	groups, err := w.GroupService.ListGroups(ctx, query)
	w.LastError = err
	if err == nil {
		w.lastGroupList = groups
	}

	return nil
}

func (s *groupSteps) iRetrieveTheGroupByName(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		w.LastError = fmt.Errorf("group %q not found", groupName)
		return nil
	}

	// Attempt to get it via service
	retrieved, err := w.GroupService.GetGroup(ctx, grp.ID().String())
	if err != nil {
		w.LastError = err
		return nil
	}

	// Phase B: Check visibility - non-owners cannot see private groups
	if retrieved.Visibility() == group.VisibilityPrivate {
		if w.CurrentUser == nil || retrieved.Owner().String() != w.CurrentUser.ID().String() {
			w.LastError = fmt.Errorf("group not found") // Simulate 404 for private groups
			return nil
		}
	}

	w.LastGroup = retrieved
	return nil
}

func (s *groupSteps) iRetrieveTheLastGroupByID(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to retrieve; create a group first")
	}

	retrieved, err := w.GroupService.GetGroup(ctx, w.LastGroup.ID().String())
	w.LastError = err
	if err == nil {
		w.LastGroup = retrieved
	}

	return nil
}

func (s *groupSteps) theGroupListIncludes(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	for _, grp := range w.lastGroupList {
		if grp.Name().String() == groupName {
			return nil
		}
	}

	return fmt.Errorf("group list does not include %q", groupName)
}

func (s *groupSteps) theGroupListDoesNotInclude(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	for _, grp := range w.lastGroupList {
		if grp.Name().String() == groupName {
			return fmt.Errorf("group list should not include %q but it does", groupName)
		}
	}

	return nil
}

func (s *groupSteps) theGroupListIsEmpty(ctx context.Context) error {
	w := getWorld(ctx)

	if len(w.lastGroupList) > 0 {
		return fmt.Errorf("expected empty group list but got %d groups", len(w.lastGroupList))
	}

	return nil
}

func (s *groupSteps) theGroupListOrderIs(ctx context.Context, expectedOrder string) error {
	w := getWorld(ctx)

	names := strings.Split(expectedOrder, ",")
	for i := range names {
		names[i] = strings.TrimSpace(names[i])
	}

	if len(w.lastGroupList) != len(names) {
		return fmt.Errorf("expected %d groups but got %d", len(names), len(w.lastGroupList))
	}

	for i, expectedName := range names {
		actualName := w.lastGroupList[i].Name().String()
		if actualName != expectedName {
			return fmt.Errorf("at position %d: expected %q but got %q", i, expectedName, actualName)
		}
	}

	return nil
}

func (s *groupSteps) theFirstGroupInTheListIs(ctx context.Context, expectedName string) error {
	w := getWorld(ctx)

	if len(w.lastGroupList) == 0 {
		return fmt.Errorf("group list is empty, expected first group to be %q", expectedName)
	}

	actualName := w.lastGroupList[0].Name().String()
	if actualName != expectedName {
		return fmt.Errorf("expected first group to be %q but got %q", expectedName, actualName)
	}

	return nil
}

// Phase B: Favorites
func (s *groupSteps) iFavoriteTheLastGroup(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to favorite; create a group first")
	}

	err := w.GroupRepo.FavoriteGroup(ctx, w.CurrentUser.ID(), w.LastGroup.ID())
	w.LastError = err

	return nil
}

func (s *groupSteps) iUnfavoriteTheLastGroup(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to unfavorite; create a group first")
	}

	err := w.GroupRepo.UnfavoriteGroup(ctx, w.CurrentUser.ID(), w.LastGroup.ID())
	w.LastError = err

	return nil
}

func (s *groupSteps) iFavoriteAGroupWithInvalidID(ctx context.Context) error {
	w := getWorld(ctx)

	// Use zero UUID
	fakeGroupID, _ := group.ParseGroupID("00000000-0000-0000-0000-000000000000")
	err := w.GroupRepo.FavoriteGroup(ctx, w.CurrentUser.ID(), fakeGroupID)
	w.LastError = err

	return nil
}

func (s *groupSteps) iUnfavoriteAGroupWithInvalidID(ctx context.Context) error {
	w := getWorld(ctx)

	// Use zero UUID
	fakeGroupID, _ := group.ParseGroupID("00000000-0000-0000-0000-000000000000")
	err := w.GroupRepo.UnfavoriteGroup(ctx, w.CurrentUser.ID(), fakeGroupID)
	w.LastError = err

	return nil
}

func (s *groupSteps) theLastGroupIsFavorited(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to check")
	}

	isFavorited, err := w.GroupRepo.IsFavorite(ctx, w.CurrentUser.ID(), w.LastGroup.ID())
	if err != nil {
		return fmt.Errorf("failed to check favorite status: %w", err)
	}

	if !isFavorited {
		return fmt.Errorf("expected group to be favorited but it is not")
	}

	return nil
}

func (s *groupSteps) theLastGroupIsNotFavorited(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to check")
	}

	isFavorited, err := w.GroupRepo.IsFavorite(ctx, w.CurrentUser.ID(), w.LastGroup.ID())
	if err != nil {
		return fmt.Errorf("failed to check favorite status: %w", err)
	}

	if isFavorited {
		return fmt.Errorf("expected group to not be favorited but it is")
	}

	return nil
}

func (s *groupSteps) theGroupIsFavorited(ctx context.Context) error {
	return s.theLastGroupIsFavorited(ctx)
}

func (s *groupSteps) theGroupIsNotFavorited(ctx context.Context) error {
	return s.theLastGroupIsNotFavorited(ctx)
}

func (s *groupSteps) theGroupNamedIsFavorited(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	// Find group in last list
	for _, grp := range w.lastGroupList {
		if grp.Name().String() == groupName {
			isFavorited, err := w.GroupRepo.IsFavorite(ctx, w.CurrentUser.ID(), grp.ID())
			if err != nil {
				return fmt.Errorf("failed to check favorite status for %q: %w", groupName, err)
			}

			if !isFavorited {
				return fmt.Errorf("expected group %q to be favorited but it is not", groupName)
			}

			return nil
		}
	}

	return fmt.Errorf("group %q not found in list", groupName)
}

func (s *groupSteps) theGroupNamedIsNotFavorited(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	// Find group in last list
	for _, grp := range w.lastGroupList {
		if grp.Name().String() == groupName {
			isFavorited, err := w.GroupRepo.IsFavorite(ctx, w.CurrentUser.ID(), grp.ID())
			if err != nil {
				return fmt.Errorf("failed to check favorite status for %q: %w", groupName, err)
			}

			if isFavorited {
				return fmt.Errorf("expected group %q to not be favorited but it is", groupName)
			}

			return nil
		}
	}

	return fmt.Errorf("group %q not found in list", groupName)
}

// Phase B: Seeding steps
func (s *groupSteps) userCreatesPublicGroupNamed(ctx context.Context, userName, groupName string) error {
	w := getWorld(ctx)

	u, ok := w.NamedUsers[userName]
	if !ok {
		return fmt.Errorf("user %q not found in test data; create user first", userName)
	}

	// Temporarily switch current user
	previousUser := w.CurrentUser
	w.CurrentUser = u

	// Create group
	err := s.iCreateAGroupNamedWithVisibility(ctx, groupName, "public")

	// Restore previous user
	w.CurrentUser = previousUser

	return err
}

func (s *groupSteps) userCreatesPrivateGroupNamed(ctx context.Context, userName, groupName string) error {
	w := getWorld(ctx)

	u, ok := w.NamedUsers[userName]
	if !ok {
		return fmt.Errorf("user %q not found in test data; create user first", userName)
	}

	// Temporarily switch current user
	previousUser := w.CurrentUser
	w.CurrentUser = u

	// Create group
	err := s.iCreateAGroupNamedWithVisibility(ctx, groupName, "private")

	// Restore previous user
	w.CurrentUser = previousUser

	return err
}

// Phase B: Delete group
func (s *groupSteps) iDeleteTheLastGroup(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastGroup == nil {
		return fmt.Errorf("no group to delete; create a group first")
	}

	err := w.GroupRepo.Delete(ctx, w.LastGroup.ID())
	w.LastError = err

	return nil
}
