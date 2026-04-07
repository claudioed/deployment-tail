## ADDED Requirements

### Requirement: Group selection in Web UI Quick Create

The system SHALL provide multi-select group picker in Quick Create modal.

#### Scenario: Display group selector
- **WHEN** user opens Quick Create modal
- **THEN** system displays group selector field labeled "Groups (optional)"

#### Scenario: Load groups with favorites first
- **WHEN** user opens Quick Create modal
- **THEN** system fetches user's groups ordered with favorites first, then alphabetically

#### Scenario: Multi-select groups
- **WHEN** user selects multiple groups from picker
- **THEN** system allows selection and displays selected group names

#### Scenario: Create schedule with groups
- **WHEN** user submits Quick Create with groups selected
- **THEN** system creates schedule and assigns to all selected groups

#### Scenario: Create schedule without groups
- **WHEN** user submits Quick Create with no groups selected
- **THEN** system creates schedule without group assignments (existing behavior)

#### Scenario: Rollback on group assignment failure
- **WHEN** schedule creation succeeds but group assignment fails
- **THEN** system deletes the created schedule and shows error message

#### Scenario: Show error without closing modal
- **WHEN** group assignment fails
- **THEN** system displays error message and keeps Quick Create modal open with form values preserved

### Requirement: CLI quick command with groups flag

The system SHALL accept `--groups` flag in `schedule quick` command.

#### Scenario: Quick create with group IDs
- **WHEN** user runs `deployment-tail schedule quick <service> --env <env> --in <minutes> --groups "id1,id2"`
- **THEN** system creates schedule and assigns to groups with specified IDs

#### Scenario: Quick create with group names
- **WHEN** user runs `deployment-tail schedule quick <service> --env <env> --in <minutes> --groups "Project Alpha,Team Backend"`
- **THEN** system resolves names to IDs and assigns schedule to matching groups

#### Scenario: Quick create without groups flag
- **WHEN** user runs `deployment-tail schedule quick` without `--groups` flag
- **THEN** system creates schedule without group assignments (existing behavior)

#### Scenario: Group name resolution
- **WHEN** CLI receives group names in `--groups` flag
- **THEN** system calls list groups API to resolve names to IDs before assignment

#### Scenario: Group name not found
- **WHEN** user specifies non-existent group name
- **THEN** system returns error listing available groups

#### Scenario: Ambiguous group name
- **WHEN** user specifies group name matching multiple groups
- **THEN** system returns error listing matching groups and suggests using ID

#### Scenario: Mixed IDs and names
- **WHEN** user provides comma-separated list with both UUIDs and names
- **THEN** system parses UUIDs as IDs and resolves non-UUIDs as names

#### Scenario: Rollback on CLI group assignment failure
- **WHEN** schedule creation succeeds but group assignment fails in CLI
- **THEN** CLI deletes the created schedule and exits with error message

### Requirement: Transactional semantics

The system SHALL ensure schedule creation and group assignment succeed or fail together.

#### Scenario: Full success
- **WHEN** both schedule creation and group assignment succeed
- **THEN** system returns success with schedule details including assigned groups

#### Scenario: Schedule creation failure
- **WHEN** schedule creation fails
- **THEN** system returns error without attempting group assignment

#### Scenario: Group assignment failure with rollback
- **WHEN** schedule creation succeeds but group assignment fails
- **THEN** system deletes the created schedule and returns error

#### Scenario: Rollback failure handling
- **WHEN** group assignment fails and schedule deletion also fails
- **THEN** system logs error and returns message indicating orphaned schedule

### Requirement: Group selector UI/UX

The system SHALL provide efficient group selection experience.

#### Scenario: Checkbox-based multi-select
- **WHEN** user interacts with group selector
- **THEN** system displays checkboxes for each group

#### Scenario: Show favorite indicator
- **WHEN** user views group list in selector
- **THEN** system displays star icon or label for favorited groups

#### Scenario: Empty state
- **WHEN** user has no groups
- **THEN** system displays message "No groups available. Create groups first."

#### Scenario: Loading state
- **WHEN** system fetches groups
- **THEN** system displays loading spinner in group selector

#### Scenario: Keyboard navigation
- **WHEN** user tabs through Quick Create form
- **THEN** group selector receives focus and allows keyboard selection

### Requirement: Validation consistency

The system SHALL apply same validation rules as existing group assignment API.

#### Scenario: Validate group IDs exist
- **WHEN** user assigns to non-existent group ID
- **THEN** system returns 404 error with group ID

#### Scenario: Validate user access to groups
- **WHEN** user attempts to assign to group they cannot access
- **THEN** system returns authorization error

#### Scenario: Ignore duplicate group selections
- **WHEN** user selects same group multiple times
- **THEN** system creates only one assignment per unique group

### Requirement: Authentication enforcement

The system SHALL require authentication for group assignment during Quick Create.

#### Scenario: Authenticated Quick Create with groups
- **WHEN** authenticated user creates schedule with groups
- **THEN** system creates schedule owned by user and assigns to groups

#### Scenario: Unauthenticated request rejected
- **WHEN** unauthenticated user attempts Quick Create with groups
- **THEN** system returns 401 Unauthorized error

### Requirement: Performance optimization

The system SHALL minimize API calls and latency.

#### Scenario: Fetch groups once on modal open
- **WHEN** user opens Quick Create modal
- **THEN** system fetches groups once and caches for modal lifetime

#### Scenario: Client-side group filtering
- **WHEN** user types in group search box (if implemented)
- **THEN** system filters cached groups client-side without API calls

#### Scenario: Parallel schedule creation and group fetch
- **WHEN** user opens Quick Create modal
- **THEN** system can fetch groups in parallel with rendering form

### Requirement: Error messaging

The system SHALL provide clear, actionable error messages.

#### Scenario: Schedule creation error
- **WHEN** schedule creation fails
- **THEN** system shows error from schedule API (e.g., "Service name required")

#### Scenario: Group assignment error
- **WHEN** group assignment fails
- **THEN** system shows error with group context (e.g., "Failed to assign to group 'Project Alpha'")

#### Scenario: Rollback error
- **WHEN** rollback deletion fails
- **THEN** system shows "Schedule created but group assignment failed. Schedule ID: {id}"

#### Scenario: Network error
- **WHEN** API call fails due to network issue
- **THEN** system shows generic error with retry suggestion

## Notes

- Group assignment is optional and backward compatible
- Reuses existing POST /schedules and POST /schedules/:id/groups endpoints
- Client orchestrates 2-step flow: create → assign (with rollback on failure)
- CLI supports both group IDs and names for flexibility
- Web UI uses checkbox-based multi-select for speed
- Favorites-first ordering matches existing group list behavior
- No new backend endpoints or API changes required

## Affected Components

- **Web UI**: Quick Create modal component (add group multi-select)
- **CLI**: `schedule quick` command (add `--groups` flag and resolution logic)
- **Client-side orchestration**: Create + assign flow with rollback handling
- **API**: Reuses POST /schedules and POST /schedules/:id/groups

## Rollback Plan

1. Remove group selector from Web UI Quick Create modal
2. Remove `--groups` flag from CLI `schedule quick` command
3. No API changes to revert (no backend modifications)
4. No database changes to revert (reuses existing schema)
5. Existing group assignments remain valid
