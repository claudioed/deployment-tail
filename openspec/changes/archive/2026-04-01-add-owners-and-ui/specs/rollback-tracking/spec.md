## ADDED Requirements

### Requirement: Store rollback plan

The system SHALL allow users to specify an optional rollback plan for each schedule.

#### Scenario: Create schedule with rollback plan
- **WHEN** user creates schedule with rollback plan "Revert to v1.2.3 using git reset"
- **THEN** system stores the rollback plan text

#### Scenario: Create schedule without rollback plan
- **WHEN** user creates schedule without rollback plan
- **THEN** system stores NULL/empty for rollback plan

### Requirement: Display rollback plan

The system SHALL display rollback plan when retrieving schedules.

#### Scenario: Show rollback plan if present
- **WHEN** user retrieves schedule with rollback plan
- **THEN** system includes rollback plan in response

#### Scenario: Show empty rollback plan
- **WHEN** user retrieves schedule without rollback plan
- **THEN** system indicates no rollback plan specified

### Requirement: Update rollback plan

The system SHALL allow updating rollback plan for existing schedules.

#### Scenario: Add rollback plan to existing schedule
- **WHEN** user updates schedule to add rollback plan "Restore from backup"
- **THEN** system saves the new rollback plan

#### Scenario: Update existing rollback plan
- **WHEN** user updates schedule with different rollback plan
- **THEN** system replaces old rollback plan with new one

#### Scenario: Remove rollback plan
- **WHEN** user updates schedule with empty rollback plan
- **THEN** system clears the rollback plan field

### Requirement: Rollback plan format

The system SHALL accept rollback plans as free-form text.

#### Scenario: Accept multi-line rollback plan
- **WHEN** user provides rollback plan with multiple lines
- **THEN** system stores the complete text including line breaks

#### Scenario: Accept rollback plan with special characters
- **WHEN** user provides rollback plan with code snippets or commands
- **THEN** system stores the text without modification

#### Scenario: Limit rollback plan length
- **WHEN** user provides rollback plan exceeding 5000 characters
- **THEN** system returns error "Rollback plan too long (max 5000 characters)"

## Notes

- Rollback plan is optional free-form text
- Maximum length: 5000 characters
- Supports multi-line text and special characters
- No validation of rollback plan content (user responsibility)
- Future enhancement: Template-based rollback plans

## Affected Components

- **Domain Model**: RollbackPlan value object
- **Database**: rollback_plan TEXT column
- **API**: rollbackPlan field in requests/responses
- **CLI**: --rollback-plan flag for create/update commands
- **Web UI**: Text area for rollback plan input

## Rollback Plan

1. Remove rollback_plan column from database
2. Remove RollbackPlan value object from domain
3. Remove rollbackPlan field from API
4. Remove --rollback-plan flag from CLI
5. Remove rollback plan input from Web UI
