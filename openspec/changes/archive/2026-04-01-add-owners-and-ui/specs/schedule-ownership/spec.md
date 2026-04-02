## ADDED Requirements

### Requirement: Track schedule owner

The system SHALL record and display the owner for each deployment schedule.

#### Scenario: Owner is set on creation
- **WHEN** user creates a schedule with owner "john.doe"
- **THEN** system stores "john.doe" as the schedule owner

#### Scenario: Owner is displayed
- **WHEN** user retrieves a schedule
- **THEN** system includes owner in the response

#### Scenario: Owner cannot be changed
- **WHEN** user attempts to update the owner field
- **THEN** system rejects the request with error "Owner cannot be modified"

### Requirement: Filter schedules by owner

The system SHALL allow filtering schedules by owner.

#### Scenario: Filter by single owner
- **WHEN** user requests schedules with owner filter "john.doe"
- **THEN** system returns only schedules owned by "john.doe"

#### Scenario: List shows multiple owners
- **WHEN** user lists all schedules
- **THEN** system displays owner for each schedule

### Requirement: Owner validation

The system SHALL validate owner format.

#### Scenario: Valid owner format
- **WHEN** user creates schedule with owner "john.doe" or "john.doe@example.com"
- **THEN** system accepts the owner value

#### Scenario: Reject empty owner
- **WHEN** user creates schedule with empty owner
- **THEN** system returns error "Owner is required"

#### Scenario: Reject invalid owner format
- **WHEN** user creates schedule with owner containing special characters like "john@#$"
- **THEN** system returns error "Invalid owner format"

## Notes

- Owner format: alphanumeric, dots, hyphens, underscores, and @ symbol allowed
- Owner field is case-sensitive
- Maximum length: 255 characters
- Owner is immutable after creation for audit purposes

## Affected Components

- **Domain Model**: Owner value object with validation
- **API**: Owner parameter in create requests; owner filter in list requests
- **CLI**: --owner flag for create and list commands
- **Database**: owner column with NOT NULL constraint

## Rollback Plan

1. Remove owner column from database (preserve data in backup)
2. Remove Owner value object from domain
3. Remove owner parameters from API
4. Remove --owner flags from CLI
5. Update tests to remove owner-related assertions
