## 1. Project Setup

- [x] 1.1 Initialize Go module structure with hexagonal architecture directories (internal/domain, internal/application, internal/adapters, internal/infrastructure)
- [x] 1.2 Add required Go dependencies (chi/gin, mysql driver, cobra, oapi-codegen, validator, migrate)
- [x] 1.3 Create configuration structure for database connection and API server settings
- [x] 1.4 Set up logging infrastructure

## 2. OpenAPI Specification

- [x] 2.1 Create OpenAPI 3.0 specification file (openapi.yaml) with schema definitions
- [x] 2.2 Define POST /api/v1/schedules endpoint (create schedule)
- [x] 2.3 Define GET /api/v1/schedules endpoint (list schedules with filters)
- [x] 2.4 Define GET /api/v1/schedules/{id} endpoint (get schedule by ID)
- [x] 2.5 Define PUT /api/v1/schedules/{id} endpoint (update schedule)
- [x] 2.6 Define DELETE /api/v1/schedules/{id} endpoint (delete schedule)
- [x] 2.7 Generate Go server stubs from OpenAPI spec using oapi-codegen

## 3. Database Setup

- [x] 3.1 Create initial migration file for schedules table with proper indexes
- [x] 3.2 Implement migration runner using golang-migrate
- [x] 3.3 Create rollback migration for schedules table
- [x] 3.4 Add database connection pool configuration and initialization

## 4. Domain Layer

- [x] 4.1 Create Schedule aggregate root entity with business logic
- [x] 4.2 Implement ScheduleID value object (UUID)
- [x] 4.3 Implement ScheduledTime value object with validation
- [x] 4.4 Implement ServiceName value object with validation
- [x] 4.5 Implement Environment value object (enum: production, staging, development)
- [x] 4.6 Implement Description value object
- [x] 4.7 Define ScheduleRepository interface (port) in domain layer
- [x] 4.8 Add domain validation rules and invariants

## 5. Application Layer

- [x] 5.1 Define CreateSchedule use case interface (inbound port)
- [x] 5.2 Define GetSchedule use case interface (inbound port)
- [x] 5.3 Define ListSchedules use case interface with filter parameters (inbound port)
- [x] 5.4 Define UpdateSchedule use case interface (inbound port)
- [x] 5.5 Define DeleteSchedule use case interface (inbound port)
- [x] 5.6 Implement CreateSchedule use case with validation
- [x] 5.7 Implement GetSchedule use case with error handling
- [x] 5.8 Implement ListSchedules use case with filtering logic
- [x] 5.9 Implement UpdateSchedule use case with validation
- [x] 5.10 Implement DeleteSchedule use case

## 6. Infrastructure - MySQL Repository Adapter

- [x] 6.1 Implement MySQL repository adapter for ScheduleRepository interface
- [x] 6.2 Implement Create method with transaction support
- [x] 6.3 Implement FindByID method
- [x] 6.4 Implement FindAll method with date range filtering
- [x] 6.5 Implement FindAll method with environment filtering
- [x] 6.6 Implement Update method with optimistic locking
- [x] 6.7 Implement Delete method
- [x] 6.8 Add database error mapping to domain errors

## 7. HTTP API Adapter

- [x] 7.1 Create HTTP server setup with router (chi/gin)
- [x] 7.2 Implement CreateSchedule HTTP handler
- [x] 7.3 Implement GetSchedule HTTP handler
- [x] 7.4 Implement ListSchedules HTTP handler with query parameters
- [x] 7.5 Implement UpdateSchedule HTTP handler
- [x] 7.6 Implement DeleteSchedule HTTP handler
- [x] 7.7 Add HTTP error mapping and response formatting
- [x] 7.8 Add request validation middleware
- [x] 7.9 Implement health check endpoint
- [x] 7.10 Add CORS configuration if needed

## 8. CLI Adapter

- [x] 8.1 Set up Cobra CLI structure with schedule command group
- [x] 8.2 Create HTTP client for API communication
- [x] 8.3 Implement "schedule create" command with flags
- [x] 8.4 Implement "schedule get" command
- [x] 8.5 Implement "schedule list" command with filter flags
- [x] 8.6 Implement "schedule update" command with flags
- [x] 8.7 Implement "schedule delete" command
- [x] 8.8 Add table formatting for list/get output
- [x] 8.9 Add JSON output flag for all commands
- [x] 8.10 Add API endpoint configuration (env variable or config file)
- [x] 8.11 Implement helpful error messages for API connection failures

## 9. Testing

- [x] 9.1 Write unit tests for domain entities and value objects
- [x] 9.2 Write unit tests for application use cases (with mock repository)
- [x] 9.3 Write integration tests for MySQL repository
- [x] 9.4 Write integration tests for HTTP API endpoints
- [x] 9.5 Write end-to-end tests for CLI commands
- [x] 9.6 Add contract tests to verify OpenAPI spec compliance
- [x] 9.7 Verify all spec scenarios are covered by tests

## 10. Documentation and Deployment

- [x] 10.1 Create README with setup instructions
- [x] 10.2 Document API usage with examples
- [x] 10.3 Document CLI usage with examples
- [x] 10.4 Create database setup guide
- [x] 10.5 Add configuration examples (database connection, API endpoint)
- [x] 10.6 Create Docker Compose file for local development (MySQL + API)
- [x] 10.7 Add Makefile with common tasks (build, test, migrate, run)
