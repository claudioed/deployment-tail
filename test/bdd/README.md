# BDD Test Suite

This directory contains Behavior-Driven Development (BDD) tests using [Godog](https://github.com/cucumber/godog), the official Cucumber framework for Go.

## Overview

The BDD suite turns the 19 OpenSpec specifications (`openspec/specs/*/spec.md`) into executable Gherkin tests. Each scenario validates the system against the acceptance criteria written in the specs.

**Goals:**
- Make specs executable and directly tied to production code
- Provide a shared vocabulary (Given/When/Then) for product/QA/dev
- Complement (not replace) existing unit/integration tests
- Cover all layers: service (fast, mocked), HTTP (middleware + status codes), and UI (sidebar, forms)

## Running Tests

```bash
# Run all BDD tests
make test-bdd

# Run smoke tests only (fast, critical paths)
make test-bdd-smoke

# Run by layer
make test-bdd-service    # Service layer with mocks (fastest)
make test-bdd-http       # HTTP layer with real middleware
make test-bdd-ui         # UI layer with chromedp (Phase E)

# Run specific spec
go test -v ./test/bdd/... -godog.tags=@spec-schedule-crud

# Skip WIP scenarios
go test -v ./test/bdd/... -godog.tags="~@wip"

# CI-friendly (JUnit report)
make test-bdd-ci
```

## Tags

| Tag | Purpose |
|-----|---------|
| `@service` | Runs against application services directly with mocks (fast) |
| `@http` | Runs against real chi router via httptest |
| `@ui` | Chromedp-driven, uses httptest server serving `./web/*` |
| `@cli` | Invokes Cobra root command in-process |
| `@auth` | Exercises authentication/authorization middleware |
| `@smoke` | Minimum set that must pass pre-merge |
| `@spec-<name>` | Tags scenarios by spec (e.g., `@spec-schedule-crud`) |
| `@wip` | Work in progress, skipped via `--godog.tags=~@wip` |

## Phase A: Pilot (Current)

**Status:** ✅ Implemented

**Specs covered:**
- `schedule-crud` (create, read)
- `schedule-groups` (create group)

**Files:**
- `bdd_test.go` — Godog runner entry point
- `suite.go` — World struct, service wiring
- `hooks.go` — Test suite/scenario initialization
- `common_steps.go` — Shared auth/error/HTTP-status steps
- `schedule_steps.go` — Schedule CRUD steps
- `group_steps.go` — Group management steps
- `http_steps.go` — Generic HTTP request/assert steps
- `features/schedule-crud/*.feature` — Schedule CRUD scenarios
- `features/schedule-groups/*.feature` — Group scenarios

## Coverage Matrix

| Spec | @service | @http | @ui | Total Scenarios |
|------|----------|-------|-----|-----------------|
| schedule-crud | 5 | 2 | 0 | 7 |
| schedule-groups | 2 | 2 | 0 | 4 |
| **Phase A Total** | **7** | **4** | **0** | **11** |

*(Matrix will expand as phases B-E are implemented)*

## Upcoming Phases

| Phase | Specs | New Files |
|-------|-------|-----------|
| **B. Groups & Assignment** | group-visibility, group-favorites, schedule-group-assignment | assignment_steps.go |
| **C. Auth & Users** | api-authorization, jwt-session-management, google-oauth-signin, user-management | auth_steps.go, user_steps.go |
| **D. Domain Workflows** | schedule-ownership, schedule-approval, rollback-tracking, multi-owner-support, multi-environment-support, inline-status-edit, date-grouped-schedules, cli-authentication | cli_steps.go |
| **E. UI** | web-ui, sidebar-navigation | ui_steps.go, chromedp wiring |

## Architecture

### World Struct

The `World` struct holds all state for a single scenario:
- **Repositories:** Mocked repositories from `applicationtest`
- **Services:** Real application services under test
- **HTTP wiring:** `httptest.Server` + real middleware (for `@http`/`@ui`)
- **Browser context:** chromedp context (for `@ui`)
- **Scenario state:** Current user, last entities, errors, HTTP responses

Fresh `World` is created before each scenario and torn down after.

### Step Definitions

Step files are organized by domain concept (not feature) to maximize reuse:
- `common_steps.go` — Authentication, errors, HTTP status
- `schedule_steps.go` — Schedule CRUD operations
- `group_steps.go` — Group management
- `http_steps.go` — Generic HTTP requests
- *(Future)* `assignment_steps.go`, `auth_steps.go`, `user_steps.go`, `cli_steps.go`, `ui_steps.go`

### Feature Files

Features mirror the OpenSpec structure:
```
features/
├── schedule-crud/
│   ├── create.feature
│   ├── read.feature
│   ├── update.feature
│   └── delete.feature
├── schedule-groups/
│   └── create_group.feature
└── (more specs...)
```

## Design Decisions

- **Service wiring:** Same pattern as `handler_test.go:27` — mocks from `applicationtest`
- **HTTP wiring:** Real `jwt.JWTService` + middleware stack (copied from `route_protection_integration_test.go`)
- **OAuth mocking:** Reuses `MockGoogleClient` from `auth_handler_test.go`
- **Assertion style:** Steps return `error` (Cucumber-idiomatic) for failures
- **chromedp on CI:** Tagged `@ui` and skipped by default; dedicated CI job uses headless-shell
- **No state pollution:** Each scenario gets a fresh World via `Before` hook

## Verification

After Phase A implementation:

```bash
# 1. Dependencies installed
go mod tidy && go mod download

# 2. Run smoke tests (fast, @service only)
make test-bdd-smoke
```

Expected output:
```
Feature: Schedule CRUD — Create

  @service @spec-schedule-crud @smoke
  Scenario: Create schedule with required fields
    Given I am authenticated as a deployer named "Alice"
    When I create a schedule with:
    ...
    Then no error is returned
    And the last schedule has service name "api-service"

11 scenarios (11 passed)
XX steps (XX passed)
PASS
```

**Regression check:**
```bash
make test                # Existing unit tests still green
make test-integration    # Existing integration tests still green
```

## Contributing

When adding new scenarios:
1. Tag with `@wip` during development
2. Choose appropriate layer tags (`@service`, `@http`, `@ui`)
3. Add spec tag (e.g., `@spec-schedule-crud`)
4. Mark critical paths with `@smoke`
5. Update coverage matrix in this README
6. Convert OpenSpec scenarios manually (not sed) to preserve semantics

## References

- [Godog GitHub](https://github.com/cucumber/godog)
- [Godog Package Docs](https://pkg.go.dev/github.com/cucumber/godog)
- [Cucumber Best Practices](https://cucumber.io/docs/bdd/)
- OpenSpec source: `openspec/specs/*/spec.md`
