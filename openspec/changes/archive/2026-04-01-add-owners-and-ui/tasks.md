## 1. Database Migration

- [x] 1.1 Create migration file 000002_add_owner_status_rollback.up.sql
- [x] 1.2 Add owner VARCHAR(255) column with NOT NULL and default 'system'
- [x] 1.3 Add status ENUM('created','approved','denied') column with NOT NULL and default 'approved'
- [x] 1.4 Add rollback_plan TEXT column (nullable)
- [x] 1.5 Add index on owner column (idx_owner)
- [x] 1.6 Add index on status column (idx_status)
- [x] 1.7 Create rollback migration 000002_add_owner_status_rollback.down.sql
- [x] 1.8 Test migrations up and down on test database

## 2. Domain Layer - Value Objects

- [x] 2.1 Create Owner value object with validation (alphanumeric, dots, hyphens, underscores, @)
- [x] 2.2 Add Owner validation for max length (255 characters)
- [x] 2.3 Create Status value object as enum (Created, Approved, Denied)
- [x] 2.4 Implement status transition validation in Status value object
- [x] 2.5 Create RollbackPlan value object with max length validation (5000 characters)
- [x] 2.6 Add unit tests for Owner value object
- [x] 2.7 Add unit tests for Status value object transitions
- [x] 2.8 Add unit tests for RollbackPlan value object

## 3. Domain Layer - Schedule Entity

- [x] 3.1 Add owner field to Schedule entity
- [x] 3.2 Add status field to Schedule entity
- [x] 3.3 Add rollbackPlan field to Schedule entity
- [x] 3.4 Update NewSchedule constructor to require owner parameter
- [x] 3.5 Update NewSchedule to set status to "created" by default
- [x] 3.6 Add Approve() method to Schedule entity
- [x] 3.7 Add Deny() method to Schedule entity
- [x] 3.8 Prevent owner modification in Update() method
- [x] 3.9 Update Reconstitute() to include new fields
- [x] 3.10 Add unit tests for schedule with owner, status, rollback plan
- [x] 3.11 Add unit tests for approve/deny methods
- [x] 3.12 Add unit tests for owner immutability

## 4. Domain Layer - Repository Interface

- [x] 4.1 Update Filters struct to include Owner filter
- [x] 4.2 Update Filters struct to include Status filter
- [x] 4.3 Add domain errors for invalid status transitions

## 5. Application Layer - Use Cases

- [x] 5.1 Update CreateScheduleCommand to include owner field
- [x] 5.2 Update CreateScheduleCommand to include rollbackPlan field (optional)
- [x] 5.3 Update CreateSchedule use case to handle owner
- [x] 5.4 Update ListSchedulesQuery to include owner filter
- [x] 5.5 Update ListSchedulesQuery to include status filter
- [x] 5.6 Update ListSchedules use case to handle new filters
- [x] 5.7 Create ApproveScheduleCommand
- [x] 5.8 Create ApproveSchedule use case
- [x] 5.9 Create DenyScheduleCommand
- [x] 5.10 Create DenySchedule use case
- [x] 5.11 Update unit tests for CreateSchedule with owner
- [x] 5.12 Add unit tests for ApproveSchedule use case
- [x] 5.13 Add unit tests for DenySchedule use case

## 6. MySQL Repository Adapter

- [x] 6.1 Update Create method to store owner, status, rollback_plan
- [x] 6.2 Update FindByID to retrieve owner, status, rollback_plan
- [x] 6.3 Update FindAll to support owner filter
- [x] 6.4 Update FindAll to support status filter
- [x] 6.5 Update Update method to prevent owner changes
- [x] 6.6 Update mapToSchedule to include new fields
- [x] 6.7 Add integration tests for owner filtering
- [x] 6.8 Add integration tests for status filtering
- [x] 6.9 Add integration tests for rollback plan storage

## 7. OpenAPI Specification Updates

- [x] 7.1 Add owner field to Schedule schema (required)
- [x] 7.2 Add status field to Schedule schema (enum: created, approved, denied)
- [x] 7.3 Add rollbackPlan field to Schedule schema (optional)
- [x] 7.4 Add owner to CreateScheduleRequest (required)
- [x] 7.5 Add rollbackPlan to CreateScheduleRequest (optional)
- [x] 7.6 Add rollbackPlan to UpdateScheduleRequest (optional)
- [x] 7.7 Add owner and status query parameters to GET /schedules
- [x] 7.8 Create POST /api/v1/schedules/{id}/approve endpoint
- [x] 7.9 Create POST /api/v1/schedules/{id}/deny endpoint
- [x] 7.10 Regenerate Go server stubs from updated OpenAPI spec

## 8. HTTP API Adapter

- [x] 8.1 Update ListSchedules handler to parse owner filter
- [x] 8.2 Update ListSchedules handler to parse status filter
- [x] 8.3 Update CreateSchedule handler to extract owner from request
- [x] 8.4 Update CreateSchedule handler to extract rollbackPlan from request
- [x] 8.5 Update toAPISchedule to include owner, status, rollbackPlan
- [x] 8.6 Create ApproveSchedule HTTP handler
- [x] 8.7 Create DenySchedule HTTP handler
- [x] 8.8 Add error handling for invalid status transitions
- [x] 8.9 Update HTTP integration tests for new fields
- [x] 8.10 Add integration tests for approve/deny endpoints

## 9. CLI Updates

- [x] 9.1 Add --owner flag to schedule create command (required)
- [x] 9.2 Add --rollback-plan flag to schedule create command (optional)
- [x] 9.3 Add --rollback-plan flag to schedule update command (optional)
- [x] 9.4 Add --owner filter flag to schedule list command
- [x] 9.5 Add --status filter flag to schedule list command
- [x] 9.6 Update table output to display owner and status columns
- [x] 9.7 Create schedule approve <id> command
- [x] 9.8 Create schedule deny <id> command
- [x] 9.9 Update schedule get output to show rollback plan
- [x] 9.10 Update CLI client to call approve endpoint
- [x] 9.11 Update CLI client to call deny endpoint
- [x] 9.12 Add CLI tests for new commands

## 10. Web UI - Project Setup

- [x] 10.1 Create web/ directory structure
- [x] 10.2 Create index.html with basic layout
- [x] 10.3 Create styles.css with responsive design
- [x] 10.4 Create app.js main application file
- [x] 10.5 Create api-client.js for API communication
- [x] 10.6 Add static file serving to Go API server

## 11. Web UI - Schedule List View

- [x] 11.1 Create schedule-list.js component
- [x] 11.2 Implement fetch schedules from API
- [x] 11.3 Display schedules in table (id, service, environment, scheduled time, owner, status)
- [x] 11.4 Add filter inputs (date range, environment, owner, status)
- [x] 11.5 Implement filter functionality
- [x] 11.6 Add "Create Schedule" button
- [x] 11.7 Add click handlers to view schedule details
- [x] 11.8 Add responsive table design for mobile

## 12. Web UI - Create Schedule Form

- [x] 12.1 Create schedule-form.js component
- [x] 12.2 Add form fields (scheduled time, service, environment, owner, description, rollback plan)
- [x] 12.3 Implement form validation
- [x] 12.4 Add submit handler to POST to API
- [x] 12.5 Display success/error messages
- [x] 12.6 Clear form after successful creation
- [x] 12.7 Add cancel button to return to list

## 13. Web UI - Schedule Detail View

- [x] 13.1 Create schedule-detail.js component
- [x] 13.2 Display full schedule information
- [x] 13.3 Show rollback plan in text area
- [x] 13.4 Add "Edit" button (for created/approved schedules)
- [x] 13.5 Add "Delete" button with confirmation
- [x] 13.6 Add "Approve" button (only for created status)
- [x] 13.7 Add "Deny" button (only for created status)
- [x] 13.8 Implement approve action
- [x] 13.9 Implement deny action
- [x] 13.10 Update UI after status change
- [x] 13.11 Add "Back to List" button

## 14. Web UI - Edit Schedule Form

- [x] 14.1 Pre-fill form with existing schedule data
- [x] 14.2 Make owner field read-only
- [x] 14.3 Allow editing other fields (service, environment, scheduled time, description, rollback plan)
- [x] 14.4 Add submit handler to PUT to API
- [x] 14.5 Display success/error messages
- [x] 14.6 Return to detail view after successful update

## 15. Web UI - Polish and UX

- [x] 15.1 Add loading indicators during API calls
- [x] 15.2 Add error handling for network failures
- [x] 15.3 Style status badges (created=yellow, approved=green, denied=red)
- [x] 15.4 Add date/time picker for scheduled time field
- [x] 15.5 Add auto-refresh or manual refresh button for schedule list
- [x] 15.6 Add favicon and page title
- [x] 15.7 Test on multiple browsers (Chrome, Firefox, Safari)
- [x] 15.8 Test responsive design on mobile devices

## 16. Documentation

- [x] 16.1 Update README with owner, status, and rollback plan features
- [x] 16.2 Document approval workflow in README
- [x] 16.3 Add Web UI usage section to README
- [x] 16.4 Update API documentation with new endpoints
- [x] 16.5 Document new CLI commands and flags
- [x] 16.6 Add migration guide for existing users
- [x] 16.7 Update .env.example with any new configuration
- [x] 16.8 Add screenshots of Web UI to documentation

## 17. Testing and Validation

- [x] 17.1 Run all unit tests and verify they pass
- [x] 17.2 Run integration tests with MySQL
- [x] 17.3 Test complete workflow: create → approve → verify
- [x] 17.4 Test complete workflow: create → deny → verify
- [x] 17.5 Test owner immutability (attempt to change owner)
- [x] 17.6 Test invalid status transitions (approved → denied)
- [x] 17.7 Test filtering by owner
- [x] 17.8 Test filtering by status
- [x] 17.9 Test rollback plan storage and retrieval
- [x] 17.10 Verify backward compatibility with existing schedules
- [x] 17.11 Test Web UI end-to-end workflow
- [x] 17.12 Verify all spec scenarios are covered by tests
