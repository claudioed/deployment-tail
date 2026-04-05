package group

import (
	"testing"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
)

func TestNewGroup(t *testing.T) {
	name, _ := NewGroupName("Project Alpha")
	desc, _ := NewDescription("All schedules for Project Alpha")
	owner, _ := schedule.NewOwner("john.doe")

	group, err := NewGroup(name, desc, owner)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if group == nil {
		t.Fatal("expected group to be created")
	}

	if !group.Name().Equals(name) {
		t.Errorf("expected name %v, got %v", name, group.Name())
	}

	if group.Description().String() != desc.String() {
		t.Errorf("expected description %v, got %v", desc, group.Description())
	}

	if !group.Owner().Equals(owner) {
		t.Errorf("expected owner %v, got %v", owner, group.Owner())
	}

	// Verify ID is set
	if group.ID().String() == "" {
		t.Error("expected group ID to be generated")
	}

	// Verify timestamps are set
	if group.CreatedAt().IsZero() {
		t.Error("expected createdAt to be set")
	}

	if group.UpdatedAt().IsZero() {
		t.Error("expected updatedAt to be set")
	}
}

func TestGroup_Rename(t *testing.T) {
	name, _ := NewGroupName("Project Alpha")
	desc, _ := NewDescription("All schedules for Project Alpha")
	owner, _ := schedule.NewOwner("john.doe")

	group, _ := NewGroup(name, desc, owner)
	originalUpdatedAt := group.UpdatedAt()

	// Rename the group
	newName, _ := NewGroupName("Project Beta")
	err := group.Rename(newName)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !group.Name().Equals(newName) {
		t.Errorf("expected name %v, got %v", newName, group.Name())
	}

	// Verify updatedAt changed
	if !group.UpdatedAt().After(originalUpdatedAt) {
		t.Error("expected updatedAt to be updated after rename")
	}

	// Owner should remain unchanged (immutable)
	if !group.Owner().Equals(owner) {
		t.Error("owner should not change during rename")
	}
}

func TestGroup_UpdateDescription(t *testing.T) {
	name, _ := NewGroupName("Project Alpha")
	desc, _ := NewDescription("All schedules for Project Alpha")
	owner, _ := schedule.NewOwner("john.doe")

	group, _ := NewGroup(name, desc, owner)
	originalUpdatedAt := group.UpdatedAt()

	// Update description
	newDesc, _ := NewDescription("Updated description for Project Alpha")
	err := group.UpdateDescription(newDesc)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if group.Description().String() != newDesc.String() {
		t.Errorf("expected description %v, got %v", newDesc, group.Description())
	}

	// Verify updatedAt changed
	if !group.UpdatedAt().After(originalUpdatedAt) {
		t.Error("expected updatedAt to be updated after description change")
	}

	// Name and owner should remain unchanged
	if !group.Name().Equals(name) {
		t.Error("name should not change during description update")
	}

	if !group.Owner().Equals(owner) {
		t.Error("owner should not change during description update")
	}
}

func TestGroup_Reconstitute(t *testing.T) {
	id, _ := ParseGroupID("550e8400-e29b-41d4-a716-446655440000")
	name, _ := NewGroupName("Project Alpha")
	desc, _ := NewDescription("All schedules for Project Alpha")
	owner, _ := schedule.NewOwner("john.doe")

	// Create a new group first to get timestamps
	original, _ := NewGroup(name, desc, owner)

	// Reconstitute from storage
	reconstituted := Reconstitute(
		id,
		name,
		desc,
		owner,
		original.CreatedAt(),
		original.UpdatedAt(),
	)

	if !reconstituted.ID().Equals(id) {
		t.Errorf("expected ID %v, got %v", id, reconstituted.ID())
	}

	if !reconstituted.Name().Equals(name) {
		t.Errorf("expected name %v, got %v", name, reconstituted.Name())
	}

	if !reconstituted.Owner().Equals(owner) {
		t.Errorf("expected owner %v, got %v", owner, reconstituted.Owner())
	}

	if reconstituted.CreatedAt() != original.CreatedAt() {
		t.Error("createdAt should match")
	}

	if reconstituted.UpdatedAt() != original.UpdatedAt() {
		t.Error("updatedAt should match")
	}
}

func TestGroup_EmptyDescription(t *testing.T) {
	name, _ := NewGroupName("Project Alpha")
	emptyDesc, _ := NewDescription("")
	owner, _ := schedule.NewOwner("john.doe")

	group, err := NewGroup(name, emptyDesc, owner)

	if err != nil {
		t.Fatalf("expected no error with empty description, got %v", err)
	}

	if !group.Description().IsEmpty() {
		t.Error("expected description to be empty")
	}
}
