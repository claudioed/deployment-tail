# Testing Guide

This document describes the testing strategy and how to run tests for the deployment-tail project.

## Test Structure

The project follows a layered testing approach:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test integration with external systems (MySQL)
3. **API Tests** - Test HTTP endpoints
4. **E2E Tests** - Test complete workflows via CLI

## Running Tests

### Unit Tests

Run all unit tests (no external dependencies required):

```bash
make test
# or
go test -v ./...
```

### Integration Tests

Integration tests require a running MySQL instance. They are tagged with `// +build integration`.

1. Start MySQL:
```bash
make docker-up
```

2. Run integration tests:
```bash
make test-integration
# or
go test -v -tags=integration ./...
```

### API Contract Tests

Verify that the implementation matches the OpenAPI specification:

```bash
# Start the server
make run-server

# Run contract tests (in another terminal)
go test -v ./internal/adapters/input/http/...
```

## Test Coverage

### Domain Layer Tests (`internal/domain/schedule/`)

- Value object validation (ScheduleID, ServiceName, Environment, etc.)
- Entity business logic
- Domain rules and invariants

**Example**:
```go
func TestEnvironmentValidation(t *testing.T) {
    tests := []struct {
        name    string
        env     string
        wantErr bool
    }{
        {"valid production", "production", false},
        {"invalid environment", "invalid", true},
    }
    // ... test implementation
}
```

### Application Layer Tests (`internal/application/`)

- Use case logic with mock repositories
- Error handling
- Validation

**Example**:
```go
func TestCreateSchedule(t *testing.T) {
    repo := NewMockRepository()
    service := NewScheduleService(repo)

    cmd := input.CreateScheduleCommand{
        ScheduledAt: time.Now().Add(24 * time.Hour),
        ServiceName: "test-service",
        Environment: "production",
    }

    sch, err := service.CreateSchedule(context.Background(), cmd)
    // ... assertions
}
```

### Repository Integration Tests (`internal/adapters/output/mysql/`)

- Database operations
- Query filtering
- Error handling

**Setup**:
```bash
# Set test database DSN (optional)
export TEST_DB_DSN="root:rootpass@tcp(localhost:3306)/deployment_schedules_test?parseTime=true"

# Run integration tests
go test -v -tags=integration ./internal/adapters/output/mysql/...
```

### HTTP API Tests (`internal/adapters/input/http/`)

- Endpoint responses
- Status codes
- Request validation
- Error responses

### CLI E2E Tests

Test complete workflows:

1. Start API server
2. Execute CLI commands
3. Verify results

**Example**:
```bash
# Create schedule
./bin/deployment-tail schedule create \
  --date "2026-04-01T14:00:00Z" \
  --service "test" \
  --env "production"

# Verify it exists
./bin/deployment-tail schedule list
```

## Test Database Setup

For integration and E2E tests, create a separate test database:

```sql
CREATE DATABASE deployment_schedules_test;
```

The test database schema is created automatically by integration tests.

## Mocking

We use custom mocks for external dependencies:

- **MockRepository**: Mock implementation of `schedule.Repository`
- Located in test files alongside the code they test

## Best Practices

1. **Unit tests should be fast** - No external dependencies
2. **Integration tests use build tags** - `// +build integration`
3. **Clean up test data** - Use `defer` to cleanup in integration tests
4. **Test error cases** - Not just happy paths
5. **Use table-driven tests** - For testing multiple scenarios

## Continuous Integration

In CI/CD:

```bash
# Run unit tests (fast, no dependencies)
go test -v -short ./...

# Run all tests (requires MySQL)
docker-compose up -d mysql
go test -v -tags=integration ./...
```

## Coverage Report

Generate a coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Spec Scenario Coverage

All scenarios from `specs/schedule-crud/spec.md` should have corresponding tests:

- ✓ Create schedule with all required fields
- ✓ Create schedule with optional description
- ✓ Reject schedule without required fields
- ✓ Get schedule by ID
- ✓ List all schedules
- ✓ Filter schedules by date range
- ✓ Filter schedules by environment
- ✓ Get non-existent schedule
- ✓ Update schedule date/time
- ✓ Update multiple fields
- ✓ Update non-existent schedule
- ✓ Delete existing schedule
- ✓ Delete non-existent schedule
- ✓ Data persists after restart

Use this checklist to ensure complete test coverage.
