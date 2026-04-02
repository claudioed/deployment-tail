## MODIFIED Requirements

### Requirement: Create deployment schedule

The system SHALL allow users to create a new deployment schedule record with required fields including scheduled date/time, service name, environment, owner, and optional rollback plan.

#### Scenario: Create schedule with all required fields
- **WHEN** user creates a schedule with date "2026-04-01 14:00", service "api-service", environment "production", owner "john.doe"
- **THEN** system saves the record with status "created" and returns a unique schedule ID

#### Scenario: Create schedule with optional description
- **WHEN** user creates a schedule with required fields and description "Quarterly release"
- **THEN** system saves the record including the description

#### Scenario: Create schedule with rollback plan
- **WHEN** user creates a schedule with required fields and rollback plan "Revert to v1.2.3"
- **THEN** system saves the record including the rollback plan

#### Scenario: Reject schedule without required fields
- **WHEN** user attempts to create a schedule without specifying service name
- **THEN** system returns an error indicating service name is required

#### Scenario: Reject schedule without owner
- **WHEN** user attempts to create a schedule without specifying owner
- **THEN** system returns an error indicating owner is required

### Requirement: Read deployment schedule

The system SHALL allow users to retrieve deployment schedule records by ID or list all schedules with filtering by owner and status.

#### Scenario: Get schedule by ID
- **WHEN** user requests schedule with ID "123"
- **THEN** system returns the full schedule record with all fields including owner, status, and rollback plan

#### Scenario: List all schedules
- **WHEN** user requests all schedules
- **THEN** system returns a list of all schedule records with owner and status

#### Scenario: Filter schedules by date range
- **WHEN** user requests schedules between "2026-04-01" and "2026-04-30"
- **THEN** system returns only schedules within that date range

#### Scenario: Filter schedules by environment
- **WHEN** user requests schedules for environment "production"
- **THEN** system returns only schedules where environment is "production"

#### Scenario: Filter schedules by owner
- **WHEN** user requests schedules for owner "john.doe"
- **THEN** system returns only schedules where owner is "john.doe"

#### Scenario: Filter schedules by status
- **WHEN** user requests schedules with status "approved"
- **THEN** system returns only schedules where status is "approved"

#### Scenario: Get non-existent schedule
- **WHEN** user requests schedule with ID that doesn't exist
- **THEN** system returns an error indicating schedule not found

### Requirement: Update deployment schedule

The system SHALL allow users to update existing deployment schedule records by ID including owner, status, and rollback plan fields.

#### Scenario: Update schedule date/time
- **WHEN** user updates schedule "123" with new date "2026-04-02 10:00"
- **THEN** system saves the updated date and returns the modified record

#### Scenario: Update multiple fields
- **WHEN** user updates schedule "123" with new service "web-service" and description "Updated release"
- **THEN** system saves all updated fields

#### Scenario: Update rollback plan
- **WHEN** user updates schedule "123" with new rollback plan "Restore from backup"
- **THEN** system saves the updated rollback plan

#### Scenario: Update non-existent schedule
- **WHEN** user attempts to update schedule ID that doesn't exist
- **THEN** system returns an error indicating schedule not found

## ADDED Requirements

### Requirement: Default status for new schedules

The system SHALL set status to "created" when a new schedule is created.

#### Scenario: New schedule has created status
- **WHEN** user creates a new schedule
- **THEN** system sets status to "created" automatically

### Requirement: Track schedule owner

The system SHALL record the owner for each deployment schedule.

#### Scenario: Owner is required
- **WHEN** user creates a schedule
- **THEN** system requires owner field to be provided

#### Scenario: Owner is immutable
- **WHEN** user attempts to change the owner of an existing schedule
- **THEN** system rejects the update and returns an error

## Notes

- Owner field is immutable after creation
- Valid status values: created, approved, denied
- Status transitions should be controlled (see schedule-approval spec)
- Rollback plan is optional text field
- All filtering can be combined (date range + environment + owner + status)

## Affected Components

- **Storage Layer**: Add columns for owner, status, rollback_plan to schedules table
- **Domain Model**: Update Schedule entity with Owner, Status, and RollbackPlan value objects
- **API**: Modify all endpoints to include new fields; add filtering parameters
- **CLI**: Update commands to support new fields and filters
- **Validation**: Validate status enum values and owner format

## Rollback Plan

1. **Database Migration**: Create down migration to remove new columns
2. **API Versioning**: Consider v1 endpoints without new fields if backward compatibility needed
3. **Data Preservation**: Backup schedules with owner/status data before rollback
4. **Code Revert**: Revert commits that added owner, status, rollback plan
5. **Validation**: Ensure existing functionality works without new fields
