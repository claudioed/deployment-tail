## MODIFIED Requirements

### Requirement: Read deployment schedule

The system SHALL allow authenticated users to retrieve deployment schedule records by ID or list all schedules with filtering by owners, environments, status, and group membership, including creator and modifier information.

#### Scenario: Get schedule by ID
- **WHEN** authenticated user requests schedule with ID "123"
- **THEN** system returns the full schedule record with all fields including owners array, environments array, status, rollback plan, list of groups, created_by, and updated_by

#### Scenario: List all schedules
- **WHEN** authenticated user requests all schedules
- **THEN** system returns a list of all schedule records with owners array, environments array, status, groups array, and audit fields

#### Scenario: Filter schedules by date range
- **WHEN** authenticated user requests schedules between "2026-04-01" and "2026-04-30"
- **THEN** system returns only schedules within that date range

#### Scenario: Filter schedules by environment
- **WHEN** authenticated user requests schedules for environment "production"
- **THEN** system returns all schedules where "production" is one of the assigned environments

#### Scenario: Filter schedules by multiple environments
- **WHEN** authenticated user requests schedules with environment filters ["production", "staging"]
- **THEN** system returns schedules where ANY of the specified environments is assigned (OR logic)

#### Scenario: Filter schedules by owner
- **WHEN** authenticated user requests schedules for owner "john.doe"
- **THEN** system returns all schedules where "john.doe" is one of the assigned owners

#### Scenario: Filter schedules by multiple owners
- **WHEN** authenticated user requests schedules with owner filters ["john.doe", "jane.smith"]
- **THEN** system returns schedules where ANY of the specified owners is assigned (OR logic)

#### Scenario: Filter schedules by status
- **WHEN** authenticated user requests schedules with status "approved"
- **THEN** system returns only schedules where status is "approved"

#### Scenario: Filter schedules by group
- **WHEN** authenticated user requests schedules in group "Project Alpha"
- **THEN** system returns only schedules assigned to that group

#### Scenario: Get non-existent schedule
- **WHEN** authenticated user requests schedule with ID that doesn't exist
- **THEN** system returns an error indicating schedule not found

#### Scenario: Schedule includes empty groups array
- **WHEN** user requests schedule not assigned to any group
- **THEN** system returns schedule with groups as empty array

#### Scenario: Schedule includes multiple groups
- **WHEN** user requests schedule assigned to multiple groups
- **THEN** system returns schedule with groups array containing all assigned groups

#### Scenario: Schedule includes multiple owners
- **WHEN** user requests schedule with multiple owners
- **THEN** system returns schedule with owners array containing all assigned owners

#### Scenario: Schedule includes multiple environments
- **WHEN** authenticated user requests schedule with multiple environments
- **THEN** system returns schedule with environments array containing all assigned environments

#### Scenario: Schedule includes creator information
- **WHEN** authenticated user requests schedule
- **THEN** response includes created_by with user ID, email, name, and created_at timestamp

#### Scenario: Schedule includes modifier information
- **WHEN** authenticated user requests updated schedule
- **THEN** response includes updated_by with user ID, email, name, and updated_at timestamp

#### Scenario: Unauthenticated user cannot read schedules
- **WHEN** request without valid JWT attempts to read schedules
- **THEN** system returns 401 Unauthorized

#### Scenario: All roles can read schedules
- **WHEN** authenticated user with any role (viewer, deployer, admin) requests schedules
- **THEN** system returns schedule data

### Requirement: Create deployment schedule

The system SHALL allow authenticated users to create new deployment schedules with multiple owners and multiple environments, automatically recording the creator.

#### Scenario: Create schedule with required fields
- **WHEN** authenticated user submits create request with service name "api-service", scheduled time "2026-04-15T14:00:00Z", owners ["john.doe"], and environments ["production"]
- **THEN** system creates schedule and returns the new schedule ID with all fields including owners and environments arrays

#### Scenario: Create schedule with multiple owners
- **WHEN** authenticated user creates schedule with owners ["john.doe", "jane.smith", "bob.jones"]
- **THEN** system creates schedule with all three owners assigned

#### Scenario: Create schedule with multiple environments
- **WHEN** authenticated user creates schedule with environments ["production", "staging"]
- **THEN** system creates schedule with both environments assigned

#### Scenario: Create schedule with optional description
- **WHEN** authenticated user creates schedule with description "Deploy new feature"
- **THEN** system stores description with the schedule

#### Scenario: Create schedule with optional rollback plan
- **WHEN** authenticated user creates schedule with rollback plan "Revert to version 1.2.3"
- **THEN** system stores rollback plan with the schedule

#### Scenario: Validate owners array not empty
- **WHEN** authenticated user creates schedule with empty owners array
- **THEN** system returns error "At least one owner is required"

#### Scenario: Validate environments array not empty
- **WHEN** authenticated user creates schedule with empty environments array
- **THEN** system returns error "At least one environment is required"

#### Scenario: Validate each owner format
- **WHEN** authenticated user creates schedule with owners ["john.doe", "invalid@#$"]
- **THEN** system returns error "Invalid owner format: invalid@#$"

#### Scenario: Validate each environment value
- **WHEN** authenticated user creates schedule with environments ["production", "invalid"]
- **THEN** system returns error "Invalid environment: invalid"

#### Scenario: Default status on creation
- **WHEN** authenticated user creates a new schedule
- **THEN** system sets status to "created"

#### Scenario: Record schedule creator
- **WHEN** authenticated user creates schedule
- **THEN** system stores user ID in created_by field and sets created_at timestamp

#### Scenario: Creator email in owners array
- **WHEN** authenticated user creates schedule
- **THEN** system automatically includes creator's email in owners array if not already present

#### Scenario: Unauthenticated user cannot create
- **WHEN** request without valid JWT attempts to create schedule
- **THEN** system returns 401 Unauthorized

#### Scenario: Viewer role cannot create
- **WHEN** authenticated user with role "viewer" attempts to create schedule
- **THEN** system returns 403 Forbidden "Insufficient permissions - deployer role required"

### Requirement: Update deployment schedule

The system SHALL allow authorized users to update existing deployment schedules including modifying owners and environments arrays, recording the modifier.

#### Scenario: Update service name
- **WHEN** authorized user updates schedule service name to "new-api-service"
- **THEN** system updates service name and returns updated schedule

#### Scenario: Update scheduled time
- **WHEN** authorized user updates scheduled time to "2026-04-20T16:00:00Z"
- **THEN** system updates scheduled time and returns updated schedule

#### Scenario: Update owners array
- **WHEN** authorized user updates owners to ["alice.williams", "bob.jones"]
- **THEN** system replaces existing owners with new owners array

#### Scenario: Update environments array
- **WHEN** authorized user updates environments to ["staging", "development"]
- **THEN** system replaces existing environments with new environments array

#### Scenario: Add owner to existing list
- **WHEN** authorized user updates schedule by adding "charlie.brown" to existing owners
- **THEN** system includes new owner in owners array

#### Scenario: Remove owner from list
- **WHEN** authorized user updates schedule by removing "john.doe" from owners
- **THEN** system removes that owner from owners array

#### Scenario: Add environment to existing list
- **WHEN** authorized user updates schedule by adding "development" to existing environments
- **THEN** system includes new environment in environments array

#### Scenario: Remove environment from list
- **WHEN** authorized user updates schedule by removing "staging" from environments
- **THEN** system removes that environment from environments array

#### Scenario: Update description
- **WHEN** authorized user updates description to "Updated deployment description"
- **THEN** system updates description and returns updated schedule

#### Scenario: Update rollback plan
- **WHEN** authorized user updates rollback plan to "New rollback procedure"
- **THEN** system updates rollback plan and returns updated schedule

#### Scenario: Cannot update non-existent schedule
- **WHEN** authorized user attempts to update schedule that doesn't exist
- **THEN** system returns error "Schedule not found"

#### Scenario: Validate owners array not empty on update
- **WHEN** authorized user updates schedule with empty owners array
- **THEN** system returns error "At least one owner is required"

#### Scenario: Validate environments array not empty on update
- **WHEN** authorized user updates schedule with empty environments array
- **THEN** system returns error "At least one environment is required"

#### Scenario: Record schedule modifier
- **WHEN** authorized user updates schedule
- **THEN** system stores user ID in updated_by field and updates updated_at timestamp

#### Scenario: Unauthenticated user cannot update
- **WHEN** request without valid JWT attempts to update schedule
- **THEN** system returns 401 Unauthorized

#### Scenario: Deployer can update own schedule
- **WHEN** user with role "deployer" updates schedule they created
- **THEN** system updates schedule successfully

#### Scenario: Deployer cannot update others' schedule
- **WHEN** user with role "deployer" updates schedule created by different user
- **THEN** system returns 403 Forbidden "You can only modify your own schedules"

#### Scenario: Admin can update any schedule
- **WHEN** user with role "admin" updates schedule created by any user
- **THEN** system updates schedule successfully

## ADDED Requirements

### Requirement: Delete deployment schedule

The system SHALL allow authorized users to delete deployment schedules with ownership restrictions.

#### Scenario: Admin deletes any schedule
- **WHEN** user with role "admin" deletes schedule
- **THEN** system deletes schedule regardless of ownership

#### Scenario: Deployer deletes own schedule
- **WHEN** user with role "deployer" deletes schedule they created
- **THEN** system deletes schedule successfully

#### Scenario: Deployer cannot delete others' schedule
- **WHEN** user with role "deployer" attempts to delete schedule created by different user
- **THEN** system returns 403 Forbidden "You can only delete your own schedules"

#### Scenario: Viewer cannot delete schedules
- **WHEN** user with role "viewer" attempts to delete schedule
- **THEN** system returns 403 Forbidden "Insufficient permissions - deployer role required"

#### Scenario: Unauthenticated user cannot delete
- **WHEN** request without valid JWT attempts to delete schedule
- **THEN** system returns 401 Unauthorized

#### Scenario: Delete non-existent schedule
- **WHEN** authorized user attempts to delete schedule that doesn't exist
- **THEN** system returns 404 Not Found "Schedule not found"

#### Scenario: Soft delete preserves audit trail
- **WHEN** authorized user deletes schedule
- **THEN** system marks schedule as deleted but preserves record with deleted_by user ID and deleted_at timestamp

### Requirement: Include owners and environments arrays in schedule response

The system SHALL include owners and environments as arrays when returning schedule data.

#### Scenario: Schedule response includes owners array
- **WHEN** system returns a schedule
- **THEN** response includes owners field as an array of strings

#### Scenario: Schedule response includes environments array
- **WHEN** system returns a schedule
- **THEN** response includes environments field as an array of strings

#### Scenario: Owners ordered alphabetically
- **WHEN** system returns schedule with multiple owners
- **THEN** owners array is ordered alphabetically

#### Scenario: Environments ordered alphabetically
- **WHEN** system returns schedule with multiple environments
- **THEN** environments array is ordered alphabetically

#### Scenario: List response includes owners and environments for each schedule
- **WHEN** system returns list of schedules
- **THEN** each schedule includes its owners array and environments array

## Notes

- **Authentication & Authorization:**
  - All schedule operations now require authentication with valid JWT token
  - Role-based access control enforced: viewer (read-only), deployer (CRUD on own schedules), admin (CRUD on all schedules)
  - Unauthenticated requests return 401 Unauthorized
  - Unauthorized requests return 403 Forbidden with specific error message

- **Audit Trail:**
  - System automatically tracks created_by, updated_by, and deleted_by user IDs
  - Creator's email automatically added to owners array on creation
  - Audit timestamps: created_at, updated_at, deleted_at included in responses
  - User information (email, name) populated in responses for created_by and updated_by
  - Soft delete approach maintains audit trail for compliance

- **Owners & Environments (from previous changes):**
  - Owners field changed from single string to array of strings (BREAKING CHANGE)
  - Environments field changed from single enum to array of enums (BREAKING CHANGE)
  - Filtering by owner or environment uses OR logic (match ANY in filter list)
  - At least one owner and one environment required for all schedules
  - Deduplication of owners and environments happens at domain layer
  - Responses always include owners and environments sorted alphabetically
  - Groups array still included in responses (from previous spec changes)
  - Update operation replaces entire owners/environments arrays (no partial update semantics)

## Affected Components

- **Domain Layer**:
  - Schedule aggregate manages owners and environments as collections
  - Schedule aggregate includes CreatedBy, UpdatedBy value objects (user.UserID type)
  - NewSchedule() factory accepts createdBy parameter
  - Update() method accepts updatedBy parameter

- **Application Layer**:
  - All use cases accept AuthenticatedUser context parameter
  - Authorization policies enforce role-based access control
  - Ownership checks for deployer role on update/delete operations

- **API Request/Response**:
  - Owner and environment fields are arrays in OpenAPI spec
  - Authorization header required with Bearer JWT token
  - Responses include created_by and updated_by user information

- **HTTP Handlers**:
  - AuthenticationMiddleware validates JWT and extracts user context
  - RequireRole middleware enforces role-based restrictions
  - Handlers extract authenticated user from request context

- **Repository**:
  - JOIN queries to load owners and environments from junction tables
  - JOIN with users table to populate creator/modifier information
  - Soft delete queries filter out deleted schedules unless explicitly requested

- **Database Schema**:
  - `schedule_owners` and `schedule_environments` junction tables
  - `schedules` table includes created_by, updated_by, deleted_by, deleted_at columns
  - Foreign keys to users table for audit fields

- **OpenAPI Spec**:
  - Update Schedule schema - owner and environment fields become arrays
  - Add bearerAuth security scheme
  - Add security requirements to all schedule endpoints
  - Add User schema for creator/modifier information

## Rollback Plan

1. Run down migration to restore single `owner` and `environment` columns in `schedules` table
2. Revert API schema to single-value owner and environment
3. Revert domain layer to single Owner and Environment value objects
4. Drop `schedule_owners` and `schedule_environments` junction tables
5. First owner and first environment from arrays preserved during rollback
