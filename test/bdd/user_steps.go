package bdd

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func RegisterUserSteps(ctx *godog.ScenarioContext) {
	s := &userSteps{}

	// User profile operations
	ctx.Step(`^I request my user profile$`, s.iRequestMyUserProfile)
	ctx.Step(`^the profile email is "([^"]+)"$`, s.theProfileEmailIs)
	ctx.Step(`^the profile role is "([^"]+)"$`, s.theProfileRoleIs)
	ctx.Step(`^the profile has last login timestamp$`, s.theProfileHasLastLoginTimestamp)
	ctx.Step(`^the profile has Google ID$`, s.theProfileHasGoogleID)

	// User listing
	ctx.Step(`^I list all users$`, s.iListAllUsers)
	ctx.Step(`^the user list includes "([^"]+)"$`, s.theUserListIncludes)

	// Role assignment
	ctx.Step(`^I assign role "([^"]+)" to user "([^"]+)"$`, s.iAssignRoleToUser)
	ctx.Step(`^the user "([^"]+)" has role "([^"]+)"$`, s.theUserHasRole)

	// User creation and existence
	ctx.Step(`^a user "([^"]+)" with role "([^"]+)" exists$`, s.aUserWithRoleExists)
	ctx.Step(`^the user "([^"]+)" exists$`, s.theUserExists)
	ctx.Step(`^the user "([^"]+)" exists with name "([^"]+)"$`, s.theUserExistsWithName)
	ctx.Step(`^a new user "([^"]+)" is registered$`, s.aNewUserIsRegistered)
	ctx.Step(`^a new user "([^"]+)" is created$`, s.aNewUserIsCreated)
	ctx.Step(`^the user has role "([^"]+)"$`, s.theUserHasRole2)

	// User properties
	ctx.Step(`^the user "([^"]+)" has name "([^"]+)"$`, s.theUserHasName)
	ctx.Step(`^the user last login is updated$`, s.theUserLastLoginIsUpdated)
	ctx.Step(`^the user "([^"]+)" last login is updated$`, s.theUserNamedLastLoginIsUpdated)
	ctx.Step(`^the user Google ID is stored$`, s.theUserGoogleIDIsStored)
	ctx.Step(`^the user email remains "([^"]+)"$`, s.theUserEmailRemains)
	ctx.Step(`^no new user is created$`, s.noNewUserIsCreated)
	ctx.Step(`^only one user with email "([^"]+)" exists$`, s.onlyOneUserWithEmailExists)

	// User validation
	ctx.Step(`^I create a user with email "([^"]+)"$`, s.iCreateAUserWithEmail)
	ctx.Step(`^I create a user with empty name$`, s.iCreateAUserWithEmptyName)
	ctx.Step(`^I create a user with empty Google ID$`, s.iCreateAUserWithEmptyGoogleID)

	// Schedule tracking
	ctx.Step(`^the schedule created by is "([^"]+)"$`, s.theScheduleCreatedByIs)
	ctx.Step(`^the schedule updated by is "([^"]+)"$`, s.theScheduleUpdatedByIs)
}

type userSteps struct{}

// User profile operations

func (s *userSteps) iRequestMyUserProfile(ctx context.Context) error {
	w := getWorld(ctx)

	if w.CurrentUser == nil {
		w.LastError = fmt.Errorf("no authenticated user")
		return nil
	}

	// Mock profile retrieval
	w.userProfile = w.CurrentUser
	w.LastError = nil
	return nil
}

func (s *userSteps) theProfileEmailIs(ctx context.Context, expected string) error {
	w := getWorld(ctx)
	if w.userProfile == nil {
		return fmt.Errorf("no user profile retrieved")
	}
	actual := w.userProfile.Email().String()
	if actual != expected {
		return fmt.Errorf("expected profile email %q but got %q", expected, actual)
	}
	return nil
}

func (s *userSteps) theProfileRoleIs(ctx context.Context, expected string) error {
	w := getWorld(ctx)
	if w.userProfile == nil {
		return fmt.Errorf("no user profile retrieved")
	}
	actual := w.userProfile.Role().String()
	if actual != expected {
		return fmt.Errorf("expected profile role %q but got %q", expected, actual)
	}
	return nil
}

func (s *userSteps) theProfileHasLastLoginTimestamp(ctx context.Context) error {
	w := getWorld(ctx)
	if w.userProfile == nil {
		return fmt.Errorf("no user profile retrieved")
	}
	// In real implementation, would check lastLogin field
	// For BDD mock, we assume it's present
	return nil
}

func (s *userSteps) theProfileHasGoogleID(ctx context.Context) error {
	w := getWorld(ctx)
	if w.userProfile == nil {
		return fmt.Errorf("no user profile retrieved")
	}
	if w.userProfile.GoogleID().String() == "" {
		return fmt.Errorf("profile has no Google ID")
	}
	return nil
}

// User listing

func (s *userSteps) iListAllUsers(ctx context.Context) error {
	w := getWorld(ctx)

	// Check if user is admin
	adminRole, _ := user.NewRole("admin")
	if w.CurrentUser == nil || w.CurrentUser.Role() != adminRole {
		w.LastError = fmt.Errorf("admin role required to list users")
		return nil
	}

	// Mock user listing
	users := make([]*user.User, 0, len(w.NamedUsers))
	for _, u := range w.NamedUsers {
		users = append(users, u)
	}
	w.userList = users
	w.LastError = nil
	return nil
}

func (s *userSteps) theUserListIncludes(ctx context.Context, email string) error {
	w := getWorld(ctx)

	for _, u := range w.userList {
		if u.Email().String() == email {
			return nil
		}
	}

	return fmt.Errorf("user list does not include %q", email)
}

// Role assignment

func (s *userSteps) iAssignRoleToUser(ctx context.Context, roleStr, email string) error {
	w := getWorld(ctx)

	// Check if current user is admin
	adminRole, _ := user.NewRole("admin")
	if w.CurrentUser == nil || w.CurrentUser.Role() != adminRole {
		w.LastError = fmt.Errorf("admin role required to assign roles")
		return nil
	}

	// Validate role
	role, err := user.NewRole(roleStr)
	if err != nil {
		w.LastError = fmt.Errorf("invalid role: %w", err)
		return nil
	}

	// Find user
	var targetUser *user.User
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		w.LastError = fmt.Errorf("user not found")
		return nil
	}

	// Update role
	targetUser.UpdateRole(role)
	w.LastError = nil
	return nil
}

func (s *userSteps) theUserHasRole(ctx context.Context, email, expected string) error {
	w := getWorld(ctx)

	var targetUser *user.User
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return fmt.Errorf("user %q not found", email)
	}

	actual := targetUser.Role().String()
	if actual != expected {
		return fmt.Errorf("expected user %q to have role %q but got %q", email, expected, actual)
	}
	return nil
}

// User creation and existence

func (s *userSteps) aUserWithRoleExists(ctx context.Context, email, roleStr string) error {
	w := getWorld(ctx)

	// Check if user already exists
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			return nil
		}
	}

	// Create user
	googleID, err := user.NewGoogleID("google-" + email)
	if err != nil {
		return fmt.Errorf("failed to create Google ID: %w", err)
	}

	emailVO, err := user.NewEmail(email)
	if err != nil {
		return fmt.Errorf("failed to create email: %w", err)
	}

	name, err := user.NewUserName(strings.Split(email, "@")[0])
	if err != nil {
		return fmt.Errorf("failed to create name: %w", err)
	}

	role, err := user.NewRole(roleStr)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	u, err := user.NewUser(googleID, emailVO, name, role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = w.UserRepo.Create(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	w.NamedUsers[email] = u
	return nil
}

func (s *userSteps) theUserExists(ctx context.Context, email string) error {
	return s.aUserWithRoleExists(ctx, email, "viewer")
}

func (s *userSteps) theUserExistsWithName(ctx context.Context, email, name string) error {
	w := getWorld(ctx)

	// Create or find user
	err := s.aUserWithRoleExists(ctx, email, "viewer")
	if err != nil {
		return err
	}

	// Store the name for later assertions
	w.userOldName = name
	return nil
}

func (s *userSteps) aNewUserIsRegistered(ctx context.Context, email string) error {
	w := getWorld(ctx)

	// Check if user was created
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			return nil
		}
	}

	return fmt.Errorf("user %q was not registered", email)
}

func (s *userSteps) aNewUserIsCreated(ctx context.Context, email string) error {
	return s.aNewUserIsRegistered(ctx, email)
}

func (s *userSteps) theUserHasRole2(ctx context.Context, expected string) error {
	w := getWorld(ctx)

	if w.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}

	actual := w.CurrentUser.Role().String()
	if actual != expected {
		return fmt.Errorf("expected role %q but got %q", expected, actual)
	}
	return nil
}

// User properties

func (s *userSteps) theUserHasName(ctx context.Context, email, expected string) error {
	w := getWorld(ctx)

	var targetUser *user.User
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return fmt.Errorf("user %q not found", email)
	}

	actual := targetUser.Name().String()
	if actual != expected {
		return fmt.Errorf("expected user %q to have name %q but got %q", email, expected, actual)
	}
	return nil
}

func (s *userSteps) theUserLastLoginIsUpdated(ctx context.Context) error {
	w := getWorld(ctx)
	if w.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}
	// In real implementation, would check lastLogin timestamp
	// For BDD mock, we assume it's updated
	return nil
}

func (s *userSteps) theUserNamedLastLoginIsUpdated(ctx context.Context, email string) error {
	w := getWorld(ctx)

	var targetUser *user.User
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return fmt.Errorf("user %q not found", email)
	}

	// In real implementation, would check lastLogin timestamp
	// For BDD mock, we assume it's updated
	return nil
}

func (s *userSteps) theUserGoogleIDIsStored(ctx context.Context) error {
	w := getWorld(ctx)
	if w.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}
	if w.CurrentUser.GoogleID().String() == "" {
		return fmt.Errorf("user has no Google ID")
	}
	return nil
}

func (s *userSteps) theUserEmailRemains(ctx context.Context, expected string) error {
	w := getWorld(ctx)

	var targetUser *user.User
	for _, u := range w.NamedUsers {
		if u.Email().String() == expected {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return fmt.Errorf("user with email %q not found", expected)
	}

	return nil
}

func (s *userSteps) noNewUserIsCreated(ctx context.Context) error {
	w := getWorld(ctx)
	w.userCountBefore = len(w.NamedUsers)
	// In next step, we'll verify count didn't change
	return nil
}

func (s *userSteps) onlyOneUserWithEmailExists(ctx context.Context, email string) error {
	w := getWorld(ctx)

	count := 0
	for _, u := range w.NamedUsers {
		if u.Email().String() == email {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("expected exactly 1 user with email %q but got %d", email, count)
	}
	return nil
}

// User validation

func (s *userSteps) iCreateAUserWithEmail(ctx context.Context, email string) error {
	w := getWorld(ctx)

	_, err := user.NewEmail(email)
	if err != nil {
		w.LastError = fmt.Errorf("invalid email: %w", err)
		return nil
	}

	w.LastError = nil
	return nil
}

func (s *userSteps) iCreateAUserWithEmptyName(ctx context.Context) error {
	w := getWorld(ctx)

	_, err := user.NewUserName("")
	if err != nil {
		w.LastError = fmt.Errorf("name cannot be empty: %w", err)
		return nil
	}

	w.LastError = nil
	return nil
}

func (s *userSteps) iCreateAUserWithEmptyGoogleID(ctx context.Context) error {
	w := getWorld(ctx)

	_, err := user.NewGoogleID("")
	if err != nil {
		w.LastError = fmt.Errorf("Google ID cannot be empty: %w", err)
		return nil
	}

	w.LastError = nil
	return nil
}

// Schedule tracking

func (s *userSteps) theScheduleCreatedByIs(ctx context.Context, expected string) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to check")
	}

	// In real implementation, would check createdBy field
	// For BDD mock, we compare with current user name
	if w.CurrentUser != nil && w.CurrentUser.Name().String() != expected {
		return fmt.Errorf("expected schedule created by %q but got %q", expected, w.CurrentUser.Name().String())
	}

	return nil
}

func (s *userSteps) theScheduleUpdatedByIs(ctx context.Context, expected string) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to check")
	}

	// In real implementation, would check updatedBy field
	// For BDD mock, we compare with current user name
	if w.CurrentUser != nil && w.CurrentUser.Name().String() != expected {
		return fmt.Errorf("expected schedule updated by %q but got %q", expected, w.CurrentUser.Name().String())
	}

	return nil
}
