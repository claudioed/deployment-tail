## ADDED Requirements

### Requirement: Create schedule group

The system SHALL allow users to create custom schedule groups with a name, optional description, and visibility setting.

#### Scenario: Create group with name
- **WHEN** user creates a group with name "Project Alpha"
- **THEN** system creates group with default visibility "private" and returns group ID

#### Scenario: Create group with name and description
- **WHEN** user creates a group with name "Q1 Releases" and description "All releases planned for Q1 2026"
- **THEN** system creates group with both name and description

#### Scenario: Create public group
- **WHEN** user creates a group with visibility set to "public"
- **THEN** system creates group with visibility "public"
- **AND** group is visible to all authenticated users

#### Scenario: Create private group explicitly
- **WHEN** user creates a group with visibility set to "private"
- **THEN** system creates group with visibility "private"
- **AND** group is visible only to owner

#### Scenario: Validate group name required
- **WHEN** user attempts to create group without a name
- **THEN** system returns validation error "Group name is required"

#### Scenario: Validate group name length
- **WHEN** user attempts to create group with name longer than 100 characters
- **THEN** system returns validation error "Group name must be 100 characters or less"

#### Scenario: Validate visibility value
- **WHEN** user attempts to create group with invalid visibility value
- **THEN** system returns validation error "Visibility must be 'public' or 'private'"

#### Scenario: Set group owner on creation
- **WHEN** user creates a group
- **THEN** system sets the creating user as the group owner

### Requirement: List schedule groups

The system SHALL allow users to retrieve accessible schedule groups (public groups + user's private groups), with favorited groups returned first.

#### Scenario: List accessible groups
- **WHEN** user requests list of groups
- **THEN** system returns all public groups plus user's private groups

#### Scenario: Public groups visible to all users
- **WHEN** user requests list of groups
- **THEN** response includes all groups with visibility "public" regardless of owner

#### Scenario: Private groups visible only to owner
- **WHEN** user requests list of groups
- **THEN** response includes only private groups where user is the owner

#### Scenario: List groups with favorites first
- **WHEN** authenticated user requests list of groups
- **THEN** system returns favorited groups first, followed by non-favorited groups

#### Scenario: List groups ordered by name within sections
- **WHEN** authenticated user requests list of groups
- **THEN** favorited groups are ordered alphabetically by name, and non-favorited groups are ordered alphabetically by name

#### Scenario: Empty groups list
- **WHEN** no accessible groups exist and user requests list
- **THEN** system returns empty array

#### Scenario: Include isFavorite and visibility fields
- **WHEN** authenticated user requests list of groups
- **THEN** system includes isFavorite boolean and visibility field for each group

### Requirement: Get schedule group by ID

The system SHALL allow users to retrieve a specific group by ID.

#### Scenario: Get existing group
- **WHEN** user requests group by valid ID
- **THEN** system returns group with all details

#### Scenario: Get non-existent group
- **WHEN** user requests group by invalid ID
- **THEN** system returns 404 Not Found error

### Requirement: Update schedule group

The system SHALL allow users to update group name, description, and visibility.

#### Scenario: Update group name
- **WHEN** user updates group name from "Project Alpha" to "Project Beta"
- **THEN** system updates group name and returns updated group

#### Scenario: Update group description
- **WHEN** user updates group description
- **THEN** system updates description and returns updated group

#### Scenario: Update group visibility
- **WHEN** user updates group visibility from "private" to "public"
- **THEN** system updates visibility and returns updated group
- **AND** group becomes visible to all users

#### Scenario: Update multiple fields
- **WHEN** user updates name, description, and visibility
- **THEN** system updates all fields and returns updated group

#### Scenario: Cannot update group owner
- **WHEN** user attempts to update group owner
- **THEN** system ignores owner field (owner is immutable)

#### Scenario: Non-owner cannot update group
- **WHEN** non-owner attempts to update a group
- **THEN** system returns 403 Forbidden

#### Scenario: Validate updated name length
- **WHEN** user updates group name to exceed 100 characters
- **THEN** system returns validation error

#### Scenario: Validate updated visibility value
- **WHEN** user updates group visibility to invalid value
- **THEN** system returns validation error "Visibility must be 'public' or 'private'"

### Requirement: Delete schedule group

The system SHALL allow users to delete schedule groups.

#### Scenario: Delete group
- **WHEN** user deletes a group
- **THEN** system removes group and all schedule-group associations

#### Scenario: Delete group removes schedule associations
- **WHEN** user deletes a group that has schedules assigned
- **THEN** system removes group and unassigns all schedules from that group

#### Scenario: Schedules remain after group deletion
- **WHEN** user deletes a group
- **THEN** schedules previously in that group remain in the system (only associations are removed)

#### Scenario: Delete non-existent group
- **WHEN** user attempts to delete group by invalid ID
- **THEN** system returns 404 Not Found error

### Requirement: Group uniqueness

The system SHALL ensure group names are unique per owner.

#### Scenario: Prevent duplicate group names for same owner
- **WHEN** user attempts to create group with name that already exists for that user
- **THEN** system returns validation error "Group name already exists"

#### Scenario: Allow same group name for different owners
- **WHEN** different users create groups with the same name
- **THEN** system allows both groups (uniqueness is scoped to owner)

### Requirement: Include schedule count

The system SHALL include the count of schedules in each group.

#### Scenario: Group with schedules shows count
- **WHEN** user retrieves group that has 5 schedules
- **THEN** system includes scheduleCount=5 in response

#### Scenario: Empty group shows zero count
- **WHEN** user retrieves group with no schedules
- **THEN** system includes scheduleCount=0 in response

### Requirement: Group visibility affects access control

The system SHALL enforce visibility-based access control for group operations.

#### Scenario: Owner can access own private group
- **WHEN** user requests their own private group by ID
- **THEN** system returns group details

#### Scenario: Non-owner cannot access private group
- **WHEN** user requests another user's private group by ID
- **THEN** system returns 404 Not Found

#### Scenario: Any user can access public group
- **WHEN** user requests a public group by ID
- **THEN** system returns group details regardless of owner

#### Scenario: Owner can delete own private group
- **WHEN** user deletes their own private group
- **THEN** system removes group and associations

#### Scenario: Non-owner cannot delete private group
- **WHEN** user attempts to delete another user's private group
- **THEN** system returns 403 Forbidden or 404 Not Found

## Notes

- Group owner is immutable after creation (for audit trail)
- Groups are soft-deleted or hard-deleted based on system configuration
- Group name uniqueness is case-insensitive per owner
- Description field is optional and can be up to 500 characters
- Default ordering is alphabetical by name
- Schedule count is computed dynamically, not stored

## Affected Components

- **Database**: New `groups` table with columns (id, name, description, owner, created_at, updated_at)
- **Domain Layer**: New Group aggregate in `internal/domain/group/`
- **Application Layer**: New group use cases (create, list, get, update, delete)
- **API Layer**: New REST endpoints for group CRUD operations
- **Repository**: New GroupRepository interface and MySQL implementation

## Rollback Plan

1. Remove group API endpoints from OpenAPI spec
2. Remove group-related HTTP handlers
3. Remove group application layer use cases
4. Remove group domain entities
5. Remove group repository implementation
6. Run down migration to drop groups table
7. No impact on existing schedules
