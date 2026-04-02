# Deployment Schedule Tool

A simple CRUD tool for managing deployment schedules. Built with Go, MySQL, and following hexagonal architecture with domain-driven design principles.

## Features

- **Schedule Management**: Create, read, update, and delete deployment schedules
- **Ownership Tracking**: Every schedule has an owner (immutable after creation)
- **Approval Workflow**: Three-state workflow (created → approved/denied)
- **Rollback Plans**: Optional rollback plans for operational safety
- **Web UI**: Modern, responsive web interface for schedule management
- **REST API**: Full-featured API for schedule management
- **CLI Tool**: Command-line interface for all operations
- **Advanced Filtering**: Filter by date range, environment, owner, and status
- **Persistent Storage**: MySQL database with automatic migrations

## Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) with **Domain-Driven Design** principles:

```
internal/
├── domain/           # Core business logic (entities, value objects)
├── application/      # Use cases and ports (interfaces)
├── adapters/         # Infrastructure implementations
│   ├── input/        # HTTP API and CLI
│   └── output/       # MySQL repository
└── infrastructure/   # Cross-cutting concerns (config, logging, db)
```

## Prerequisites

- Go 1.21+
- MySQL 8.0+
- Docker and Docker Compose (for local development)

## Quick Start

### 1. Start MySQL with Docker Compose

```bash
make docker-up
```

### 2. Build the project

```bash
make build
```

### 3. Run the API server

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=rootpass
export DB_NAME=deployment_schedules

# Run server (migrations run automatically)
./bin/server
```

The API server will start on `http://localhost:8080`

### 4. Use the CLI

```bash
# Set API endpoint (optional, defaults to localhost:8080)
export DEPLOYMENT_TAIL_API=http://localhost:8080

# Create a schedule
./bin/deployment-tail schedule create \
  --date "2026-04-01T14:00:00Z" \
  --service "api-service" \
  --env "production" \
  --owner "john.doe" \
  --description "Quarterly release" \
  --rollback-plan "Revert to v1.2.3 using git reset"

# List all schedules
./bin/deployment-tail schedule list

# Filter schedules
./bin/deployment-tail schedule list \
  --owner "john.doe" \
  --status "created" \
  --env "production"

# Get a specific schedule
./bin/deployment-tail schedule get <schedule-id>

# Update a schedule (owner cannot be changed)
./bin/deployment-tail schedule update <schedule-id> \
  --date "2026-04-02T10:00:00Z" \
  --rollback-plan "Updated rollback procedure"

# Approve a schedule
./bin/deployment-tail schedule approve <schedule-id>

# Deny a schedule
./bin/deployment-tail schedule deny <schedule-id>

# Delete a schedule
./bin/deployment-tail schedule delete <schedule-id>
```

## API Endpoints

- `POST /api/v1/schedules` - Create a schedule
- `GET /api/v1/schedules` - List schedules (supports filters: from, to, environment, owner, status)
- `GET /api/v1/schedules/{id}` - Get a schedule by ID
- `PUT /api/v1/schedules/{id}` - Update a schedule
- `POST /api/v1/schedules/{id}/approve` - Approve a schedule
- `POST /api/v1/schedules/{id}/deny` - Deny a schedule
- `DELETE /api/v1/schedules/{id}` - Delete a schedule
- `GET /health` - Health check

See `api/openapi.yaml` for full API specification.

## Web UI

Access the web interface at `http://localhost:8080/` when the server is running.

The Web UI provides:
- **Dashboard**: View all schedules with filtering capabilities
- **Create/Edit Forms**: User-friendly forms for schedule management
- **Detail View**: Complete schedule information including rollback plans
- **Approval Actions**: Approve or deny schedules directly from the UI
- **Real-time Updates**: Instant feedback on all operations
- **Responsive Design**: Works on desktop and mobile devices

## Approval Workflow

Schedules follow a three-state approval workflow:

1. **Created**: New schedules start in the `created` state
2. **Approved**: Schedules can be approved (created → approved)
3. **Denied**: Schedules can be denied (created → denied)

**Rules**:
- Only schedules in `created` state can be approved or denied
- Once approved or denied, the status cannot be changed
- Owner field is immutable after creation (for audit trail)
- All schedules created before this feature are automatically set to `approved` status with owner `system`

## Data Model

Each schedule includes:

| Field | Type | Required | Immutable | Description |
|-------|------|----------|-----------|-------------|
| id | UUID | ✓ | ✓ | Unique identifier |
| scheduledAt | DateTime | ✓ | | When deployment is scheduled |
| serviceName | String | ✓ | | Service to deploy |
| environment | Enum | ✓ | | production/staging/development |
| owner | String | ✓ | ✓ | Schedule creator (immutable) |
| status | Enum | ✓ | | created/approved/denied |
| description | String | | | Optional description |
| rollbackPlan | Text | | | Optional rollback procedure |
| createdAt | DateTime | ✓ | ✓ | Creation timestamp |
| updatedAt | DateTime | ✓ | | Last update timestamp |

## Configuration

Configure via environment variables:

### Database
- `DB_HOST` - Database host (default: `localhost`)
- `DB_PORT` - Database port (default: `3306`)
- `DB_USER` - Database user (default: `root`)
- `DB_PASSWORD` - Database password (default: empty)
- `DB_NAME` - Database name (default: `deployment_schedules`)

### Server
- `SERVER_HOST` - Server host (default: `0.0.0.0`)
- `SERVER_PORT` - Server port (default: `8080`)

### CLI
- `DEPLOYMENT_TAIL_API` - API endpoint URL (default: `http://localhost:8080`)

## Development

### Run tests

```bash
make test
```

### Format code

```bash
make fmt
```

### Generate OpenAPI stubs

```bash
make generate
```

### Database migrations

Migrations run automatically when the server starts. Migration files are in the `migrations/` directory.

To create a new migration:
1. Create `migrations/NNNNNN_description.up.sql`
2. Create `migrations/NNNNNN_description.down.sql`

## Project Structure

```
.
├── api/                    # OpenAPI specification and generated code
├── cmd/
│   ├── server/            # API server entry point
│   └── cli/               # CLI tool entry point
├── internal/
│   ├── domain/            # Domain layer (entities, value objects)
│   │   └── schedule/      # Schedule aggregate with Owner, Status, RollbackPlan
│   ├── application/       # Application layer (use cases)
│   ├── adapters/          # Adapters (HTTP, CLI, MySQL)
│   │   ├── input/         # HTTP handlers and CLI commands
│   │   └── output/        # MySQL repository implementation
│   └── infrastructure/    # Infrastructure (config, logging, db)
├── migrations/            # Database migrations
├── web/                   # Web UI (HTML, CSS, JavaScript)
│   ├── index.html        # Main HTML page
│   ├── styles.css        # Responsive CSS styles
│   └── app.js            # Application JavaScript
├── docker-compose.yml     # Docker Compose for local development
└── Makefile              # Build and development tasks
```

## CLI Output Formats

### Table format (default)

```bash
./bin/deployment-tail schedule list
```

### JSON format

```bash
./bin/deployment-tail schedule list --json
```

## Examples

### Create a production deployment with rollback plan

```bash
./bin/deployment-tail schedule create \
  --date "2026-04-15T20:00:00Z" \
  --service "payment-service" \
  --env "production" \
  --owner "ops-team" \
  --description "Payment gateway update v2.1.0" \
  --rollback-plan "1. Stop payment-service
2. Restore database backup from before deployment
3. Deploy previous version v2.0.5
4. Restart payment-service
5. Verify with health checks"
```

### List pending approvals

```bash
./bin/deployment-tail schedule list --status created
```

### List production schedules by owner

```bash
./bin/deployment-tail schedule list \
  --env production \
  --owner "ops-team"
```

### Filter by date range

```bash
./bin/deployment-tail schedule list \
  --from "2026-04-01T00:00:00Z" \
  --to "2026-04-30T23:59:59Z"
```

### Approval workflow example

```bash
# Create a schedule (status: created)
SCHEDULE_ID=$(./bin/deployment-tail schedule create \
  --date "2026-04-20T18:00:00Z" \
  --service "api-gateway" \
  --env "production" \
  --owner "alice" \
  --description "API v3 rollout" \
  --json | jq -r '.id')

# Review the schedule
./bin/deployment-tail schedule get $SCHEDULE_ID

# Approve it (status: created → approved)
./bin/deployment-tail schedule approve $SCHEDULE_ID

# Or deny it (status: created → denied)
# ./bin/deployment-tail schedule deny $SCHEDULE_ID
```

## License

MIT
