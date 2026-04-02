## ADDED Requirements

### Requirement: Schedule status states

The system SHALL support three status states for schedules: created, approved, denied.

#### Scenario: New schedule has created status
- **WHEN** schedule is created
- **THEN** status is automatically set to "created"

#### Scenario: Display schedule status
- **WHEN** user retrieves a schedule
- **THEN** system includes current status in response

### Requirement: Approve schedule

The system SHALL allow authorized users to approve schedules.

#### Scenario: Approve created schedule
- **WHEN** user approves schedule with status "created"
- **THEN** system changes status to "approved"

#### Scenario: Cannot approve already approved schedule
- **WHEN** user attempts to approve schedule with status "approved"
- **THEN** system returns error "Schedule is already approved"

#### Scenario: Cannot approve denied schedule
- **WHEN** user attempts to approve schedule with status "denied"
- **THEN** system returns error "Cannot approve denied schedule"

### Requirement: Deny schedule

The system SHALL allow authorized users to deny schedules.

#### Scenario: Deny created schedule
- **WHEN** user denies schedule with status "created"
- **THEN** system changes status to "denied"

#### Scenario: Cannot deny already denied schedule
- **WHEN** user attempts to deny schedule with status "denied"
- **THEN** system returns error "Schedule is already denied"

#### Scenario: Cannot deny approved schedule
- **WHEN** user attempts to deny schedule with status "approved"
- **THEN** system returns error "Cannot deny approved schedule"

### Requirement: Filter by status

The system SHALL allow filtering schedules by status.

#### Scenario: Filter by approved status
- **WHEN** user requests schedules with status "approved"
- **THEN** system returns only approved schedules

#### Scenario: Filter by created status
- **WHEN** user requests schedules with status "created"
- **THEN** system returns only schedules pending approval

#### Scenario: Filter by denied status
- **WHEN** user requests schedules with status "denied"
- **THEN** system returns only denied schedules

### Requirement: Status validation

The system SHALL validate status values.

#### Scenario: Reject invalid status
- **WHEN** user attempts to set status to "invalid"
- **THEN** system returns error "Invalid status: must be created, approved, or denied"

### Requirement: Track status changes

The system SHALL update the updatedAt timestamp when status changes.

#### Scenario: Timestamp updated on approval
- **WHEN** schedule status changes from "created" to "approved"
- **THEN** system updates the updatedAt field

## Notes

- Valid status transitions:
  - created → approved
  - created → denied
  - No other transitions allowed
- Status changes are logged in updatedAt field
- Future enhancement: Add approval history/audit trail
- Initial implementation: No user authentication required (basic approval flow)

## Affected Components

- **Domain Model**: Status value object with transition validation
- **API**: New POST /api/v1/schedules/{id}/approve and /api/v1/schedules/{id}/deny endpoints
- **CLI**: New commands: schedule approve <id>, schedule deny <id>
- **Database**: status column with ENUM type

## Rollback Plan

1. Remove approval/deny endpoints from API
2. Remove status column from database (preserve in backup)
3. Remove Status value object from domain
4. Remove approve/deny commands from CLI
5. Update all code to remove status checks
