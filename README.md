# Deployment Schedule Tool

A simple CRUD tool for managing deployment schedules. Built with Go, MySQL, and following hexagonal architecture with domain-driven design principles.

## Features

- Create, read, update, and delete deployment schedules
- REST API for schedule management
- CLI tool for interacting with the API
- Filter schedules by date range and environment
- Persistent storage in MySQL

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
  --description "Quarterly release"

# List all schedules
./bin/deployment-tail schedule list

# Get a specific schedule
./bin/deployment-tail schedule get <schedule-id>

# Update a schedule
./bin/deployment-tail schedule update <schedule-id> \
  --date "2026-04-02T10:00:00Z"

# Delete a schedule
./bin/deployment-tail schedule delete <schedule-id>
```

## API Endpoints

- `POST /api/v1/schedules` - Create a schedule
- `GET /api/v1/schedules` - List schedules (with filters)
- `GET /api/v1/schedules/{id}` - Get a schedule by ID
- `PUT /api/v1/schedules/{id}` - Update a schedule
- `DELETE /api/v1/schedules/{id}` - Delete a schedule
- `GET /health` - Health check

See `api/openapi.yaml` for full API specification.

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
│   ├── application/       # Application layer (use cases)
│   ├── adapters/          # Adapters (HTTP, CLI, MySQL)
│   └── infrastructure/    # Infrastructure (config, logging, db)
├── migrations/            # Database migrations
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

### Create a production deployment

```bash
./bin/deployment-tail schedule create \
  --date "2026-04-15T20:00:00Z" \
  --service "payment-service" \
  --env "production" \
  --description "Payment gateway update"
```

### List production schedules

```bash
./bin/deployment-tail schedule list --env production
```

### Filter by date range

```bash
./bin/deployment-tail schedule list \
  --from "2026-04-01T00:00:00Z" \
  --to "2026-04-30T23:59:59Z"
```

## License

MIT
