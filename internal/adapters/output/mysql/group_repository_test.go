// +build integration

package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func setupGroupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpass@tcp(localhost:3306)/deployment_schedules_test?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create users table (required for foreign key constraints)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id CHAR(36) PRIMARY KEY,
			google_id VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			role ENUM('viewer', 'deployer', 'admin') NOT NULL DEFAULT 'viewer',
			last_login_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Create groups table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS groups (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description VARCHAR(500),
			owner VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_group_name_per_owner (name, owner)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create groups table: %v", err)
	}

	// Create schedule_groups table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schedule_groups (
			schedule_id VARCHAR(36) NOT NULL,
			group_id VARCHAR(36) NOT NULL,
			assigned_at DATETIME NOT NULL,
			assigned_by VARCHAR(255) NOT NULL,
			PRIMARY KEY (schedule_id, group_id)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create schedule_groups table: %v", err)
	}

	// Create group_favorites table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS group_favorites (
			user_id CHAR(36) NOT NULL,
			group_id CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, group_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
			INDEX idx_group_favorites_user (user_id)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create group_favorites table: %v", err)
	}

	return db
}

func cleanupGroupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM group_favorites")
	if err != nil {
		t.Logf("warning: failed to cleanup group_favorites: %v", err)
	}

	_, err = db.Exec("DELETE FROM schedule_groups")
	if err != nil {
		t.Logf("warning: failed to cleanup schedule_groups: %v", err)
	}

	_, err = db.Exec("DELETE FROM groups")
	if err != nil {
		t.Logf("warning: failed to cleanup groups: %v", err)
	}

	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Logf("warning: failed to cleanup users: %v", err)
	}

	db.Close()
}

func TestGroupRepositoryCreate(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test description")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)

	err := repo.Create(context.Background(), grp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was saved
	retrieved, err := repo.FindByID(context.Background(), grp.ID())
	if err != nil {
		t.Fatalf("expected to find group, got error: %v", err)
	}

	if !retrieved.ID().Equals(grp.ID()) {
		t.Errorf("expected ID %v, got %v", grp.ID(), retrieved.ID())
	}

	if retrieved.Name().String() != "Project Alpha" {
		t.Errorf("expected name 'Project Alpha', got %v", retrieved.Name().String())
	}
}

func TestGroupRepositoryCreateDuplicateName(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test description")
	owner, _ := schedule.NewOwner("test-user")

	grp1, _ := group.NewGroup(name, desc, owner)
	err := repo.Create(context.Background(), grp1)
	if err != nil {
		t.Fatalf("expected no error on first create, got %v", err)
	}

	// Try to create another group with same name and owner
	grp2, _ := group.NewGroup(name, desc, owner)
	err = repo.Create(context.Background(), grp2)
	if err != group.ErrDuplicateGroupName {
		t.Errorf("expected ErrDuplicateGroupName, got %v", err)
	}
}

func TestGroupRepositoryFindAll(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	owner, _ := schedule.NewOwner("test-user")

	// Create multiple groups for the same owner
	for i := 0; i < 3; i++ {
		name, _ := group.NewGroupName("Group " + string(rune('A'+i)))
		desc, _ := group.NewDescription("Test")
		grp, _ := group.NewGroup(name, desc, owner)
		repo.Create(context.Background(), grp)
	}

	// Create a group for a different owner
	otherOwner, _ := schedule.NewOwner("other-user")
	name, _ := group.NewGroupName("Other Group")
	desc, _ := group.NewDescription("Test")
	grp, _ := group.NewGroup(name, desc, otherOwner)
	repo.Create(context.Background(), grp)

	groups, err := repo.FindAll(context.Background(), owner)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("expected 3 groups for test-user, got %d", len(groups))
	}
}

func TestGroupRepositoryFindByID(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test description")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	retrieved, err := repo.FindByID(context.Background(), grp.ID())
	if err != nil {
		t.Fatalf("expected to find group, got error: %v", err)
	}

	if !retrieved.ID().Equals(grp.ID()) {
		t.Errorf("expected ID %v, got %v", grp.ID(), retrieved.ID())
	}
}

func TestGroupRepositoryFindByIDNotFound(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	id := group.NewGroupID()

	_, err := repo.FindByID(context.Background(), id)
	if err != group.ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroupRepositoryUpdate(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Original description")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	// Update the group
	newName, _ := group.NewGroupName("Project Beta")
	newDesc, _ := group.NewDescription("Updated description")
	updatedGrp := group.Reconstitute(
		grp.ID(),
		newName,
		newDesc,
		grp.Owner(),
		grp.CreatedAt(),
		grp.UpdatedAt(),
	)

	err := repo.Update(context.Background(), updatedGrp)
	if err != nil {
		t.Fatalf("expected no error on update, got %v", err)
	}

	// Verify the update
	retrieved, _ := repo.FindByID(context.Background(), grp.ID())
	if retrieved.Name().String() != "Project Beta" {
		t.Errorf("expected name 'Project Beta', got %v", retrieved.Name().String())
	}
	if retrieved.Description().String() != "Updated description" {
		t.Errorf("expected description 'Updated description', got %v", retrieved.Description().String())
	}
}

func TestGroupRepositoryUpdateNotFound(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)

	err := repo.Update(context.Background(), grp)
	if err != group.ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroupRepositoryDelete(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	err := repo.Delete(context.Background(), grp.ID())
	if err != nil {
		t.Fatalf("expected no error on delete, got %v", err)
	}

	// Verify it was deleted
	_, err = repo.FindByID(context.Background(), grp.ID())
	if err != group.ErrGroupNotFound {
		t.Error("expected group to be deleted")
	}
}

func TestGroupRepositoryDeleteNotFound(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	id := group.NewGroupID()

	err := repo.Delete(context.Background(), id)
	if err != group.ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroupRepositoryAddSchedule(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	scheduleID := schedule.NewScheduleID()

	err := repo.AddSchedule(context.Background(), grp.ID(), scheduleID, "test-user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the association
	scheduleIDs, err := repo.GetSchedulesInGroup(context.Background(), grp.ID())
	if err != nil {
		t.Fatalf("expected no error getting schedules, got %v", err)
	}

	if len(scheduleIDs) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(scheduleIDs))
	}

	if scheduleIDs[0].String() != scheduleID.String() {
		t.Errorf("expected schedule ID %v, got %v", scheduleID, scheduleIDs[0])
	}
}

func TestGroupRepositoryAddScheduleDuplicate(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	scheduleID := schedule.NewScheduleID()

	// Add once
	repo.AddSchedule(context.Background(), grp.ID(), scheduleID, "test-user")

	// Add again - should not error (idempotent)
	err := repo.AddSchedule(context.Background(), grp.ID(), scheduleID, "test-user")
	if err != nil {
		t.Errorf("expected no error on duplicate add, got %v", err)
	}
}

func TestGroupRepositoryRemoveSchedule(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	scheduleID := schedule.NewScheduleID()
	repo.AddSchedule(context.Background(), grp.ID(), scheduleID, "test-user")

	// Remove the schedule
	err := repo.RemoveSchedule(context.Background(), grp.ID(), scheduleID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was removed
	scheduleIDs, _ := repo.GetSchedulesInGroup(context.Background(), grp.ID())
	if len(scheduleIDs) != 0 {
		t.Errorf("expected 0 schedules after removal, got %d", len(scheduleIDs))
	}
}

func TestGroupRepositoryRemoveScheduleNotAssigned(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	scheduleID := schedule.NewScheduleID()

	// Remove without adding - should not error (idempotent)
	err := repo.RemoveSchedule(context.Background(), grp.ID(), scheduleID)
	if err != nil {
		t.Errorf("expected no error on removing non-existent assignment, got %v", err)
	}
}

func TestGroupRepositoryGetSchedulesInGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")

	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	// Add multiple schedules
	scheduleIDs := []schedule.ScheduleID{
		schedule.NewScheduleID(),
		schedule.NewScheduleID(),
		schedule.NewScheduleID(),
	}

	for _, sid := range scheduleIDs {
		repo.AddSchedule(context.Background(), grp.ID(), sid, "test-user")
	}

	retrieved, err := repo.GetSchedulesInGroup(context.Background(), grp.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 schedules, got %d", len(retrieved))
	}
}

func TestGroupRepositoryGetGroupsForSchedule(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	owner, _ := schedule.NewOwner("test-user")

	// Create multiple groups
	groupIDs := make([]group.GroupID, 3)
	for i := 0; i < 3; i++ {
		name, _ := group.NewGroupName("Group " + string(rune('A'+i)))
		desc, _ := group.NewDescription("Test")
		grp, _ := group.NewGroup(name, desc, owner)
		repo.Create(context.Background(), grp)
		groupIDs[i] = grp.ID()
	}

	scheduleID := schedule.NewScheduleID()

	// Assign schedule to all groups
	for _, gid := range groupIDs {
		repo.AddSchedule(context.Background(), gid, scheduleID, "test-user")
	}

	groups, err := repo.GetGroupsForSchedule(context.Background(), scheduleID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}
}

// Favorite operations tests

func createTestUser(t *testing.T, db *sql.DB, id, googleID, email, name string) {
	_, err := db.Exec(`
		INSERT INTO users (id, google_id, email, name, role)
		VALUES (?, ?, ?, ?, 'viewer')
	`, id, googleID, email, name)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
}

func TestGroupRepositoryFavoriteGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	// Create test user
	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	// Create test group
	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	// Import user domain package
	userIDObj, _ := user.NewUserID(userID)

	// Favorite the group
	err := repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's favorited
	isFav, err := repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error checking favorite, got %v", err)
	}
	if !isFav {
		t.Error("expected group to be favorited")
	}
}

func TestGroupRepositoryFavoriteGroupIdempotent(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	userIDObj, _ := user.NewUserID(userID)

	// Favorite once
	err := repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error on first favorite, got %v", err)
	}

	// Favorite again - should not error (idempotent)
	err = repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Errorf("expected no error on duplicate favorite, got %v", err)
	}
}

func TestGroupRepositoryUnfavoriteGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	userIDObj, _ := user.NewUserID(userID)

	// Favorite first
	repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())

	// Now unfavorite
	err := repo.UnfavoriteGroup(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's not favorited
	isFav, err := repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error checking favorite, got %v", err)
	}
	if isFav {
		t.Error("expected group to not be favorited")
	}
}

func TestGroupRepositoryUnfavoriteGroupIdempotent(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	userIDObj, _ := user.NewUserID(userID)

	// Unfavorite without favoriting - should not error (idempotent)
	err := repo.UnfavoriteGroup(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Errorf("expected no error on unfavoriting non-favorited group, got %v", err)
	}
}

func TestGroupRepositoryIsFavorite(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	userIDObj, _ := user.NewUserID(userID)

	// Not favorited initially
	isFav, err := repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if isFav {
		t.Error("expected group to not be favorited initially")
	}

	// Favorite it
	repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())

	// Check again
	isFav, err = repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !isFav {
		t.Error("expected group to be favorited after favoriting")
	}
}

func TestGroupRepositoryFindAllWithFavorites(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	owner, _ := schedule.NewOwner("test-user")

	// Create multiple groups
	grp1Name, _ := group.NewGroupName("Alpha")
	grp1Desc, _ := group.NewDescription("Test")
	grp1, _ := group.NewGroup(grp1Name, grp1Desc, owner)
	repo.Create(context.Background(), grp1)

	grp2Name, _ := group.NewGroupName("Beta")
	grp2Desc, _ := group.NewDescription("Test")
	grp2, _ := group.NewGroup(grp2Name, grp2Desc, owner)
	repo.Create(context.Background(), grp2)

	grp3Name, _ := group.NewGroupName("Gamma")
	grp3Desc, _ := group.NewDescription("Test")
	grp3, _ := group.NewGroup(grp3Name, grp3Desc, owner)
	repo.Create(context.Background(), grp3)

	userIDObj, _ := user.NewUserID(userID)

	// Favorite Beta and Gamma
	repo.FavoriteGroup(context.Background(), userIDObj, grp2.ID())
	repo.FavoriteGroup(context.Background(), userIDObj, grp3.ID())

	// Get all groups with favorites
	groups, favorites, err := repo.FindAllWithFavorites(context.Background(), userIDObj, owner)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}

	// Check favorites map
	if !favorites[grp2.ID()] {
		t.Error("expected Beta to be favorited")
	}
	if !favorites[grp3.ID()] {
		t.Error("expected Gamma to be favorited")
	}
	if favorites[grp1.ID()] {
		t.Error("expected Alpha to not be favorited")
	}

	// Verify favorites are sorted first (Beta and Gamma should come before Alpha)
	// Beta and Gamma are favorited, Alpha is not
	// Within favorites, they should be alphabetically sorted
	if groups[0].Name().String() != "Beta" {
		t.Errorf("expected first group to be Beta (favorited), got %v", groups[0].Name().String())
	}
	if groups[1].Name().String() != "Gamma" {
		t.Errorf("expected second group to be Gamma (favorited), got %v", groups[1].Name().String())
	}
	if groups[2].Name().String() != "Alpha" {
		t.Errorf("expected third group to be Alpha (not favorited), got %v", groups[2].Name().String())
	}
}

func TestGroupRepositoryCascadeDeleteGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	userIDObj, _ := user.NewUserID(userID)

	// Favorite the group
	repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())

	// Verify it's favorited
	isFav, _ := repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if !isFav {
		t.Fatal("expected group to be favorited before deletion")
	}

	// Delete the group
	err := repo.Delete(context.Background(), grp.ID())
	if err != nil {
		t.Fatalf("expected no error on delete, got %v", err)
	}

	// Verify the favorite was cascaded deleted (isFavorite should return false for non-existent group)
	isFav, err = repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error checking favorite after group deletion, got %v", err)
	}
	if isFav {
		t.Error("expected favorite to be cascade deleted when group is deleted")
	}
}

func TestGroupRepositoryCascadeDeleteUser(t *testing.T) {
	db := setupGroupTestDB(t)
	defer cleanupGroupTestDB(t, db)

	repo := NewGroupRepository(db)

	userID := "user-1"
	createTestUser(t, db, userID, "google-123", "user@test.com", "Test User")

	name, _ := group.NewGroupName("Project Alpha")
	desc, _ := group.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	grp, _ := group.NewGroup(name, desc, owner)
	repo.Create(context.Background(), grp)

	userIDObj, _ := user.NewUserID(userID)

	// Favorite the group
	repo.FavoriteGroup(context.Background(), userIDObj, grp.ID())

	// Verify it's favorited
	isFav, _ := repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if !isFav {
		t.Fatal("expected group to be favorited before user deletion")
	}

	// Delete the user
	_, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	// Verify the favorite was cascade deleted
	isFav, err = repo.IsFavorite(context.Background(), userIDObj, grp.ID())
	if err != nil {
		t.Fatalf("expected no error checking favorite after user deletion, got %v", err)
	}
	if isFav {
		t.Error("expected favorite to be cascade deleted when user is deleted")
	}
}
