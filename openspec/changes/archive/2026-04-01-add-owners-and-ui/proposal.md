## Why

Teams need visibility into who owns each deployment schedule, approval workflows to control when deployments can proceed, and a web-based interface for managing schedules. Currently, schedules lack ownership tracking and approval mechanisms, and the CLI-only interface limits accessibility for non-technical team members.

## What Changes

- Add owner field to track who created each deployment schedule
- Add status field with states: created, approved, denied
- Add approval workflow capability to transition schedules between states
- Create web UI for viewing, creating, updating, and approving schedules
- Add rollback plan tracking for each schedule
- Update API endpoints to support new fields and filtering

## Capabilities

### New Capabilities

- `schedule-ownership`: Track schedule ownership and filter schedules by owner
- `schedule-approval`: Manage schedule status (created, approved, denied) with approval workflow
- `web-ui`: Browser-based interface for schedule management
- `rollback-tracking`: Store and display rollback plans for each schedule

### Modified Capabilities

- `schedule-crud`: Add owner, status, and rollback plan fields to schedule records

## Impact

- **Database**: New columns for owner, status, rollback_plan; new indexes
- **Domain Model**: Updated Schedule entity with new value objects (Owner, Status, RollbackPlan)
- **API**: Modified endpoints to include new fields; new approval endpoint
- **CLI**: Updated commands to support new fields and filtering
- **New Web UI**: Frontend application (React/Vue/HTML) with API integration
- **Authentication**: May need basic auth to identify owners (or use system username initially)
