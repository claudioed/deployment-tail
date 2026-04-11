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

*Group* (`internal/domain/group/`):
- Organizes schedules into logical collections
- `GroupID` - UUID wrapper
- `GroupName` - String with validation (non-empty, max 100 chars, unique per owner)
- `Visibility` - Enum (public, private) controlling access
- `NewGroup(owner UserID, name, visibility)` - Factory for new groups
- `Reconstitute()` - Factory for loading from storage
- `SetVisibility()` - Controlled visibility mutation
- Public groups visible to all, private only to owner

**Repository Pattern**:
- Interface defined in domain: `schedule.Repository`, `user.Repository`, `group.Repository`
- Implementation in adapters: `mysql.ScheduleRepository`, `mysql.UserRepository`, `mysql.GroupRepository`
- Returns domain entities, never database models
- Group repository filters by visibility: returns public groups + user's private groups

## Key Development Commands

### Building
```bash
make build              # Build both server and CLI to bin/
go build -o bin/server cmd/server/main.go
go build -o bin/deployment-tail cmd/cli/main.go
```

### Testing
```bash
make test               # Run unit tests (excludes BDD tests)
make test-integration   # Run integration tests (requires MySQL)
make test-bdd           # Run BDD/acceptance tests (Godog)
make test-bdd-smoke     # Run critical BDD scenarios only
go test -v ./internal/domain/schedule/  # Test specific package
```

#### Mutation Testing

Test-suite quality is guarded by **mutation testing** using [Gremlins](https://github.com/go-gremlins/gremlins) (`go-gremlins/gremlins`) — the industry-standard Go mutation testing tool. Mutation testing flips operators, conditionals, and other small constructs in `internal/**` and re-runs the suite to confirm the tests actually catch the regressions, not just that lines were executed.

```bash
make mutation-install     # One-time: install the gremlins binary (prints instructions)
make mutation-test-dry    # Fast: list runnable mutants without executing tests
make mutation-test        # Full campaign over ./internal/... (fails below thresholds)
```

**Thresholds** are configured in `.gremlins.yaml` at the repo root:

- `efficacy: 70` — killed / (killed + lived)
- `mutant-coverage: 80` — (killed + lived) / total mutants

Moderate starting values; tighten as the suite matures. Generated code (`api/generated.go`), command wiring (`cmd/**`), and hand-written mocks (`internal/application/applicationtest/mocks.go`, `**/*_mock.go`) are excluded.

**CI:** `.github/workflows/mutation.yml` runs Gremlins on every pull request that touches `internal/**`, `go.mod`, `go.sum`, or `.gremlins.yaml`, and is also available via `workflow_dispatch`. The workflow uploads `gremlins-output.json` as an artifact for post-run inspection.

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
- **BDD/Acceptance tests:** Godog-based tests in `test/bdd/` directory
- **Test database:** Set `TEST_DB_DSN` environment variable
- Mock repository: `MockRepository` in `application/schedule_service_test.go`

#### BDD Tests (Behavior-Driven Development)

The BDD test suite (`test/bdd/`) uses [Godog v0.15](https://github.com/cucumber/godog) to turn OpenSpec specifications into executable Gherkin tests.

**Running BDD tests:**
```bash
make test-bdd          # All BDD tests (~300ms)
make test-bdd-smoke    # Critical scenarios only
go test -v ./test/bdd  # Direct invocation
```

**Test layers:**
- `@service` - Fast tests against application services with mocks
- `@http` - Tests against real HTTP middleware stack via httptest
- `@ui` - Chromedp-driven browser tests (Phase E, future)

**Architecture:**
- `World` struct holds per-scenario state (fresh for each scenario)
- Step definitions in `test/bdd/*_steps.go` (common, schedule, group, http)
- Feature files in `test/bdd/features/` mirror OpenSpec structure
- Reuses `applicationtest` mocks from existing test infrastructure

**Current coverage (Phase A):**
- `schedule-crud` - 6/7 scenarios passing (1 @wip)
- `schedule-groups` - 4/4 scenarios passing

See `test/bdd/README.md` for detailed documentation and `test/bdd/PILOT_COMPLETE.md` for Phase A status.

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

### Web UI Architecture

The web UI (`web/` directory) is built with vanilla JavaScript following modern patterns:

**Structure:**
- `index.html` - Single-page app structure with sidebar + main content layout
- `styles.css` - CSS Grid layout, responsive breakpoints, component styles
- `app.js` - Application logic, rendering, state management

**Layout:**
- **CSS Grid**: Sidebar (240px) + flexible content area on desktop
- **Responsive**: < 768px breakpoint switches to mobile overlay sidebar
- **Sidebar Navigation**: Persistent group list with favorites, visibility icons, settings
- **Main Content**: Date-grouped schedules (Today, Tomorrow, This Week, Later)

**Key Components:**

*Sidebar* (`renderSidebar()`, `renderSidebarGroups()`):
- Displays "All Schedules", "Ungrouped", and accessible groups
- Visual indicators: 🌐 public, 🔒 private, ★ favorite
- Settings gear icon on hover (owner only)
- Mobile: Hamburger menu + slide-in overlay

*Date Grouping* (`groupSchedulesByDate()`, `renderDateGroupedSchedules()`):
- Organizes schedules by relative date sections
- Collapsible sections with localStorage persistence
- Time display (HH:MM) and OVERDUE badges

*URL-Based Routing*:
- Hash-based navigation: `#all`, `#ungrouped`, `#group/{id}`
- Browser back/forward support via `hashchange` event
- Bookmarkable URLs for direct group access

*Group Management*:
- Visibility toggle (public/private) in creation/edit modal
- Inline favorite toggle from sidebar
- CRUD operations with optimistic UI updates

*User Chip* (`renderUserChip()`, `toggleUserChipMenu()`, `signOut()`):
- Header-right identity surface with avatar (initials), display name, and role badge (viewer/deployer/admin)
- Dropdown menu: email, role, sign-out
- Keyboard accessible: Enter/Space opens, Escape closes, Arrow keys navigate
- Falls back to a minimal "sign out only" chip when `/users/me` fails
- Sign-out calls `POST /auth/logout` (server-side revocation) then clears localStorage and redirects

**State Management:**
- `currentUser` module-level state in `web/app.js` — **single source of truth** for the authenticated user. Populated from `GET /users/me` on bootstrap (before `loadGroupsAndRenderSidebar()` and `loadSchedules()`).
- `getCurrentUser()` returns `currentUser?.email ?? null` — **never** reads from the `filter-owner` input (that field is a search filter, not an identity source).
- URL hash for selected group
- localStorage for collapse states, tab preferences, and `auth_token` only (never PII)
- No global state management library - functional approach

**API Integration:**
- Fetch API for HTTP requests
- JWT bearer token in Authorization header
- `fetchCurrentUser()` handles 401 (redirect to sign-in), 404 (minimal-chip fallback), and 5xx/network (notification + minimal chip)
- Error handling with user-friendly notifications routed to ARIA live regions (polite for success/info, assertive for errors)

**Design Tokens:**
- `web/styles.css` defines a token layer in `:root` covering: status/environment/role colors, surface/border/text/focus/hover/pressed colors, typography, spacing (`--space-1` through `--space-8`), radius, shadow, z-index (`--z-sidebar`, `--z-dropdown`, `--z-modal`, `--z-toast`), and motion
- New components should consume tokens; legacy components are being migrated incrementally
- The full token list and inventory is documented in `docs/ux-research.md`

**Accessibility:**
- `role="banner"` on header, `role="navigation"` + `aria-label="Groups"` on sidebar, semantic `<main>` landmark
- Skip-to-main-content link (visible on keyboard focus only)
- Global `:focus-visible` ring using `--color-focus-ring` token
- `trapFocus(modalElement, onClose)` helper used by all modals (group modal, assign-group modal, quick-assign overlay)
- Polite and assertive ARIA live regions for notifications
- Sidebar gear icon revealed on `:focus-within` (keyboard accessible)

**Responsive Behavior:**
- Desktop (≥ 768px): Fixed sidebar, visible groups, full user chip
- Mobile (< 768px): Hidden sidebar, hamburger menu in header, overlay on open
- Narrow (< 480px): User chip collapses to avatar only
- Smooth CSS transitions for slide-in/out

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
