## Context

Currently, there is no system to record and track planned deployment schedules. Teams need a simple, standalone tool to manage deployment schedule records without complex integrations or automated job execution.

This is a new feature being built as a CRUD application using:
- **Golang** as the implementation language
- **MySQL** for persistent storage
- **Hexagonal Architecture** (ports and adapters) for clean separation of concerns
- **Domain-Driven Design** principles for the domain model
- **API-first** approach with OpenAPI specification

The application will store deployment schedule information (date/time, service, environment, description) and expose both REST API and CLI interfaces.

## Goals / Non-Goals

**Goals:**
- Provide simple CRUD operations for deployment schedule records via REST API
- Persist schedule data in MySQL database
- Support filtering and listing schedules by date range and environment
- Implement hexagonal architecture with clear domain boundaries
- Follow DDD principles with rich domain model
- Define API contract first using OpenAPI specification
- Provide CLI client that consumes the REST API
- Maintain clean separation between domain, application, and infrastructure layers

**Non-Goals:**
- Automated job scheduling or execution (no cron, no task queue)
- Integration with external deployment systems or CI/CD platforms
- Real-time notifications or webhooks
- Multi-user authentication or authorization (initially)
- Complex querying or reporting features
- Web UI (API and CLI only)

## Decisions

### Architecture: Hexagonal (Ports & Adapters)

**Decision**: Implement hexagonal architecture with clear separation between domain, application, and infrastructure layers.

**Layer Structure**:
```
internal/
├── domain/           # Core business logic (entities, value objects, domain services)
│   └── schedule/
├── application/      # Use cases, ports (interfaces)
│   └── ports/
│       ├── input/   # Inbound ports (use case interfaces)
│       └── output/  # Outbound ports (repository interfaces)
├── adapters/        # Infrastructure implementations
│   ├── input/
│   │   ├── http/    # REST API handlers
│   │   └── cli/     # CLI commands
│   └── output/
│       └── mysql/   # MySQL repository implementation
└── infrastructure/  # Cross-cutting concerns (config, logging)
```

**Rationale**:
- Clear dependency direction (dependencies point inward toward domain)
- Domain logic isolated from infrastructure concerns
- Easy to test domain and application layers independently
- Swap implementations without affecting core logic

**Alternatives considered**:
- Layered architecture: Less explicit about dependency direction
- Clean architecture: Similar, but hexagonal is more explicit about ports/adapters

### Domain Model: DDD Principles

**Decision**: Apply DDD tactical patterns with rich domain model.

**Domain Components**:
- **Entity**: `Schedule` aggregate root with identity (UUID)
- **Value Objects**: `ScheduledTime`, `ServiceName`, `Environment`, `Description`
- **Repository Interface**: `ScheduleRepository` (port in domain layer)
- **Domain Services**: Validation and business rules

**Schedule Entity** (domain/schedule/schedule.go):
```go
type Schedule struct {
    id          ScheduleID
    scheduledAt ScheduledTime
    service     ServiceName
    environment Environment
    description Description
    createdAt   time.Time
    updatedAt   time.Time
}
```

**Rationale**: Value objects enforce invariants, entity encapsulates business rules, repository interface keeps domain independent of persistence.

### Storage: MySQL Database

**Decision**: Use MySQL for persistent storage with dedicated schema.

**Schema**:
```sql
CREATE TABLE schedules (
    id VARCHAR(36) PRIMARY KEY,
    scheduled_at DATETIME NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    environment ENUM('production', 'staging', 'development') NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_environment (environment)
);
```

**Rationale**:
- ACID compliance for data integrity
- Better concurrency support than file-based storage
- Efficient querying with indexes
- Standard production-grade solution

**Alternatives considered**:
- PostgreSQL: Similar capabilities, MySQL chosen for existing infrastructure
- SQLite: Not suitable for multi-process access

### API Design: API-First with OpenAPI

**Decision**: Define API contract first using OpenAPI 3.0 specification before implementation.

**API Specification** (openapi.yaml):
```yaml
/api/v1/schedules:
  POST: Create schedule
  GET: List schedules (with filters)
/api/v1/schedules/{id}:
  GET: Get schedule by ID
  PUT: Update schedule
  DELETE: Delete schedule
```

**Rationale**:
- Contract-first ensures API consistency
- Generate server stubs and client SDKs
- API documentation auto-generated
- Enables parallel development of API and CLI

**Tools**: Use `oapi-codegen` for generating Go server stubs from OpenAPI spec.

### Implementation Language: Golang

**Decision**: Implement in Go 1.21+

**Key Libraries**:
- HTTP framework: `chi` or `gin` (lightweight routers)
- MySQL driver: `go-sql-driver/mysql`
- OpenAPI: `oapi-codegen` for code generation
- CLI: `cobra` for CLI structure
- Validation: `go-playground/validator`
- Migrations: `golang-migrate/migrate`

**Rationale**: Go provides excellent performance, strong typing, built-in concurrency, and great MySQL support.

### CLI Design: API Client

**Decision**: CLI acts as a thin client that consumes the REST API.

**Commands**:
```
schedule create --date "2026-04-01T14:00:00Z" --service "api-service" --env "production" [--description "text"]
schedule list [--from "2026-04-01"] [--to "2026-04-30"] [--env "production"]
schedule get <id>
schedule update <id> [--date "..."] [--service "..."] [--env "..."] [--description "..."]
schedule delete <id>
```

**Rationale**: Keeping CLI as API client ensures single source of truth (the API) and simplifies testing.

## Risks / Trade-offs

**[Risk: Database connection failures]** → Implement connection pooling, retry logic, and graceful degradation. Use health check endpoints.

**[Risk: API unavailable when using CLI]** → CLI will fail if API server is down. Document API server as prerequisite. Consider adding connection check with helpful error messages.

**[Trade-off: Increased complexity]** → Hexagonal architecture and DDD add structure but increase initial setup time. Benefits: better maintainability and testability long-term.

**[Trade-off: API server dependency]** → Requires running API server for CLI to work. Alternative: Could embed API server in CLI binary, but increases complexity.

**[Trade-off: No audit trail]** → Updates overwrite previous data. MySQL triggers or event sourcing could be added later if audit requirements emerge.

**[Risk: OpenAPI spec drift]** → Spec and implementation can diverge. Mitigation: Use code generation from spec, add contract tests.

## Migration Plan

**Deployment**:
1. Create MySQL database and schema using migration scripts
2. Build and deploy API server
3. Build CLI binary with schedule commands
4. Configure CLI with API endpoint (via config file or environment variable)
5. Run smoke tests to verify API and CLI connectivity

**Database Migration**:
- Use `golang-migrate` for versioned schema migrations
- Initial migration creates `schedules` table with indexes
- Rollback migrations included for each version

**Rollback**:
1. Stop API server
2. Rollback database migrations if needed (preserve data by default)
3. Remove schedule commands from CLI or revert to previous version
4. No impact on existing functionality (self-contained feature)

## Open Questions

- **API server deployment**: Standalone service or embedded with existing services? Port number?
- **Configuration**: Where should CLI store API endpoint config? Environment variable, config file at `~/.deployment-tail/config.yaml`, or both?
- **Date input format**: Support multiple formats (ISO 8601, relative dates like "tomorrow") or strict ISO 8601 only?
- **CLI output format**: Table by default? Support JSON output flag for scripting?
- **Authentication**: Do we need API authentication initially, or defer to future iteration?
- **Database connection**: Connection string from environment variable, config file, or CLI flag?
