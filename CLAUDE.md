# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Deployment-tail is a CRUD tool for managing deployment schedules with **Google OAuth authentication** and **role-based access control**. Built with Go, MySQL, following **Hexagonal Architecture** (Ports & Adapters) and **Domain-Driven Design** principles. The system provides both a REST API and CLI for managing deployment schedules.

## Architecture

### Hexagonal Architecture (Ports & Adapters)

The codebase strictly follows hexagonal architecture with clear separation of concerns:

```
internal/
├── domain/           # Core business logic - NO external dependencies
│   └── schedule/     # Aggregate root, value objects, repository interface (port)
├── application/      # Use cases and orchestration
│   └── ports/
│       ├── input/    # Inbound ports (use case interfaces)
│       └── output/   # Outbound ports (repository interfaces)
├── adapters/         # Infrastructure implementations
│   ├── input/        # Inbound adapters (HTTP handlers, CLI commands)
│   │   ├── http/     # REST API implementation
│   │   └── cli/      # Cobra CLI commands
│   └── output/       # Outbound adapters (MySQL repository)
│       └── mysql/    # Repository implementation
└── infrastructure/   # Cross-cutting concerns (config, logging, database)
```

**Critical architectural rules:**
- **Domain layer** must have zero external dependencies (only Go stdlib)
- **Application layer** depends only on domain
- **Adapters** depend on application/domain, never the reverse
- Repository interface lives in **domain**, implementation in **adapters/output**
- Use case interfaces live in **application/ports/input**

### Authentication Architecture

The application uses **Google OAuth 2.0** for authentication and **JWT** for session management:

**Authentication Flow:**
1. User initiates login → Redirected to Google OAuth
2. Google validates → Returns authorization code
3. Application exchanges code → Receives Google access token
4. Application fetches user profile → Registers/updates user in database
5. Application issues JWT → Client stores token
6. Client includes JWT in Authorization header for all requests
7. Middleware validates JWT → Extracts user context

**Key Components:**

**Domain Layer** (`internal/domain/user/`):
- `User` aggregate root with `UserID`, `GoogleID`, `Email`, `UserName`, `Role` value objects
- `user.Repository` interface (port)
- Domain errors for authentication failures

**Infrastructure** (`internal/infrastructure/`):
- `oauth.GoogleClient` - OAuth 2.0 flow implementation
- `jwt.JWTService` - Token generation, validation, refresh
- `jwt.RevocationStore` - Token blacklist with in-memory cache + DB persistence

**Application Layer** (`internal/application/`):
- `UserService` - Authentication use cases (register, login, profile, role management)
- Authorization policies for role-based access control
- Context-based user extraction for audit trail

**HTTP Middleware** (`internal/adapters/input/http/middleware/`):
- `AuthenticationMiddleware` - JWT validation, user context injection
- `RequireRole()` - Role-based authorization

**CLI Authentication** (`internal/adapters/input/cli/`):
- `auth.go` - Login, logout, status commands
- `token_store.go` - Secure local token storage (0600 permissions)
- `client.go` - Automatic JWT injection and refresh

**Token Management:**
- Default expiry: 24 hours (configurable)
- Automatic refresh when < 1 hour remaining (CLI)
- Revocation on logout with server-side blacklist
- Background sync of revoked tokens (every 60 seconds)
- Cleanup of expired blacklist entries

**Role-Based Access Control:**
- **viewer**: Read-only access to schedules
- **deployer**: Create/update/delete own schedules
- **admin**: Full access including approval workflow and user management

### Domain-Driven Design Patterns

**Value Objects**:

*Schedule* (`internal/domain/schedule/`):
- `ScheduleID` - UUID wrapper with parsing
- `ScheduledTime` - Time with UTC normalization and validation
- `ServiceName` - String with validation (non-empty, max 255 chars)
- `Environment` - Enum (production, staging, development)
- `Description` - Optional text

*User* (`internal/domain/user/`):
- `UserID` - UUID wrapper
- `GoogleID` - Google OAuth ID with validation
- `Email` - Email address with format validation
- `UserName` - Display name with max length (255 chars)
- `Role` - Enum (viewer, deployer, admin)

**Aggregate Roots**:

*Schedule*:
- Encapsulates scheduling business rules
- All fields private, accessed via getters
- `NewSchedule(createdBy UserID, ...)` - Factory for new schedules
- `Reconstitute()` - Factory for loading from storage (bypasses validation)
- `Update(updatedBy UserID, ...)` - Controlled mutation with audit trail
- Ownership checks for authorization

*User*:
- Encapsulates user identity and role
- `NewUser(googleID, email, name, role)` - Factory for registration
- `UpdateLastLogin()` - Track authentication activity
- `UpdateRole()` - Admin-only role changes
- Immutable GoogleID for audit trail

**Repository Pattern**:
- Interface defined in domain: `schedule.Repository`, `user.Repository`
- Implementation in adapters: `mysql.ScheduleRepository`, `mysql.UserRepository`
- Returns domain entities, never database models

## Key Development Commands

### Building
```bash
make build              # Build both server and CLI to bin/
go build -o bin/server cmd/server/main.go
go build -o bin/deployment-tail cmd/cli/main.go
```

### Testing
```bash
make test               # Run unit tests
make test-integration   # Run integration tests (requires MySQL)
go test -v ./internal/domain/schedule/  # Test specific package
```

### Running Locally
```bash
make docker-up          # Start MySQL via Docker Compose
make run-server         # Run API server (migrations auto-run)

# Or manually with environment variables:
export DB_PASSWORD=rootpass
./bin/server
```

### OpenAPI Code Generation
```bash
make generate           # Regenerate API stubs from openapi.yaml
# Under the hood: oapi-codegen -generate types,chi-server -package api api/openapi.yaml
```

### Database Migrations

Migrations run **automatically** when the server starts via `infrastructure.RunMigrations()`.

**Creating new migrations:**
1. Create `migrations/NNNNNN_description.up.sql` (use next sequential number)
2. Create `migrations/NNNNNN_description.down.sql`
3. Server will auto-apply on next startup

**Migration naming:** `000001_create_schedules_table.{up,down}.sql`

## OpenAPI-First Development

The API is defined in `api/openapi.yaml` first, then code is generated:

1. Edit `api/openapi.yaml` to add/modify endpoints
2. Run `make generate` to regenerate `api/generated.go`
3. Implement the interface methods in `internal/adapters/input/http/handler.go`

**Important:** The generated code uses `openapi_types.UUID` instead of `string` for UUID fields. Handle conversions appropriately.

## Adding New Features

### Adding a New Field to Schedule

1. **Domain Layer:**
   - Create value object in `internal/domain/schedule/<field>.go`
   - Add field to `Schedule` struct
   - Update `NewSchedule()` and `Reconstitute()`
   - Add getter method

2. **Database:**
   - Create migration `migrations/NNNNNN_add_<field>.up.sql`
   - Create rollback migration `.down.sql`

3. **Repository:**
   - Update `mysql.ScheduleRepository` Create/Update/FindByID methods
   - Update `mapToSchedule()` helper

4. **Application Layer:**
   - Add field to command structs in `application/ports/input/schedule_service.go`
   - Update use case implementations

5. **OpenAPI:**
   - Add field to Schema in `api/openapi.yaml`
   - Run `make generate`

6. **Adapters:**
   - Update HTTP handlers in `adapters/input/http/handler.go`
   - Update CLI commands in `adapters/input/cli/schedule.go`

### Adding a New Endpoint

1. Define in `api/openapi.yaml`
2. Run `make generate`
3. Implement interface method in HTTP handler
4. Create corresponding use case if needed
5. Add CLI command if applicable

## OpenSpec Workflow

This repository uses OpenSpec for change management. Active and archived changes live in `openspec/`.

**Current specs:** `openspec/specs/` contains canonical specifications
**Active changes:** `openspec/changes/` contains work-in-progress
**Archived changes:** `openspec/changes/archive/` contains completed work

When implementing features from OpenSpec:
- Read `proposal.md` for the "why"
- Read `design.md` for architectural decisions
- Read `specs/*.md` for requirements and scenarios
- Follow `tasks.md` for implementation order

## Important Patterns

### Error Handling
- Domain errors defined in `internal/domain/schedule/errors.go`
- Use `fmt.Errorf("context: %w", err)` for wrapping
- HTTP layer maps domain errors to status codes
- CLI shows user-friendly error messages

### Testing Patterns
- **Unit tests:** Domain and application layers use mocks
- **Integration tests:** Use `// +build integration` tag, require MySQL
- **Test database:** Set `TEST_DB_DSN` environment variable
- Mock repository: `MockRepository` in `application/schedule_service_test.go`

### Value Object Pattern
```go
// Enforce invariants in constructor
func NewServiceName(name string) (ServiceName, error) {
    name = strings.TrimSpace(name)
    if name == "" {
        return ServiceName{}, fmt.Errorf("service name cannot be empty")
    }
    return ServiceName{value: name}, nil
}
```

### Repository Pattern
```go
// Interface in domain, implementation in adapter
type Repository interface {
    Create(ctx context.Context, schedule *Schedule) error
    FindByID(ctx context.Context, id ScheduleID) (*Schedule, error)
    // ...
}
```

## Environment Configuration

All configuration via environment variables (see `.env.example`):

**Database:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

**Server:**
- `SERVER_HOST`, `SERVER_PORT`

**Authentication (Google OAuth):**
- `GOOGLE_CLIENT_ID` - OAuth 2.0 client ID from Google Cloud Console
- `GOOGLE_CLIENT_SECRET` - OAuth 2.0 client secret
- `GOOGLE_REDIRECT_URL` - OAuth callback URL (e.g., `http://localhost:8080/auth/google/callback`)

**JWT Configuration:**
- `JWT_SECRET` - Secret key for signing JWTs (min 32 characters, generate with `openssl rand -base64 32`)
- `JWT_EXPIRY` - Token expiry duration (e.g., `24h`, `7d`, `168h`)
- `JWT_ISSUER` - Token issuer identifier (defaults to `deployment-tail`)

**CLI:**
- `DEPLOYMENT_TAIL_API` - API endpoint for CLI

**Setup Steps:**
1. Create OAuth 2.0 credentials at https://console.cloud.google.com/apis/credentials
2. Add authorized redirect URI: `http://localhost:8080/auth/google/callback`
3. Generate a strong JWT secret: `openssl rand -base64 32`
4. Copy `.env.example` to `.env` and fill in values
5. **IMPORTANT:** Never commit `.env` or expose JWT_SECRET/GOOGLE_CLIENT_SECRET

## Common Gotchas

1. **Generated API code:** After changing `openapi.yaml`, always run `make generate` before building
2. **UUID types:** API uses `openapi_types.UUID`, convert with `uuid.MustParse()` when needed
3. **Owner is immutable:** Domain enforces this; Update() method rejects owner changes
4. **Status transitions:** Only created→approved/denied allowed (enforced in domain)
5. **Migrations auto-run:** Server startup runs migrations; no separate step needed
6. **UTC timestamps:** All time.Time values stored/compared in UTC
7. **Package naming:** Import alias required for http adapter: `httphandler "github.com/.../http"` to avoid conflict with `net/http`
8. **Authentication required:** All API endpoints (except `/health` and auth endpoints) require JWT in Authorization header
9. **JWT secret length:** Must be minimum 32 characters or server will fail validation on startup
10. **CLI authentication:** CLI commands auto-refresh tokens when < 1 hour to expiry; use `--force-login` to bypass cache
11. **Audit trail:** All schedule Create/Update operations now require authenticated user context for `createdBy`/`updatedBy` fields
12. **Role enforcement:** Authorization policies check user role; deployers can only modify own schedules, admins can modify any

## Entry Points

- **API Server:** `cmd/server/main.go` - Wires up dependencies, starts HTTP server
- **CLI:** `cmd/cli/main.go` - Delegates to `adapters/input/cli/root.go`
