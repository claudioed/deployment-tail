# Deployment Schedule Tool

A comprehensive tool for managing deployment schedules with group organization. Built with Go, MySQL, and following hexagonal architecture with domain-driven design principles.

## Features

- **Google OAuth Authentication**: Secure authentication via Google Sign-In
- **Role-Based Access Control**: Three roles (viewer, deployer, admin) with granular permissions
- **JWT Token Management**: Secure token-based authentication with automatic refresh
- **Token Revocation**: Logout functionality with server-side token blacklist
- **Audit Trail**: Track who created and modified each schedule
- **Schedule Management**: Create, read, update, and delete deployment schedules
- **Multi-Owner Support**: Schedules can have multiple owners for collaborative management
- **Multi-Environment Deployments**: Schedule deployments across multiple environments simultaneously
- **Group Organization**: Organize schedules into logical groups (projects, teams, releases)
- **Group Visibility**: Public groups (visible to all users) or private groups (visible only to owner)
- **Group Favorites**: Mark frequently used groups as favorites for quick access with star icons
- **Sidebar Navigation**: Persistent left sidebar with all accessible groups for easy navigation
- **Date-Grouped Schedules**: Schedules organized by relative date (Today, Tomorrow, This Week, Later)
- **URL-Based Group Selection**: Bookmarkable URLs for direct links to specific groups
- **Ownership Tracking**: Every schedule and group has owners
- **Approval Workflow**: Three-state workflow (created → approved/denied)
- **Inline Status Editing**: Edit schedule status directly from the list view with keyboard support
- **Tag-Based Input**: Semicolon-separated input for adding multiple owners quickly
- **Rollback Plans**: Optional rollback plans for operational safety
- **Web UI**: Modern, responsive web interface with sidebar navigation, date grouping, and mobile support (collapsible sidebar < 768px)
- **REST API**: Full-featured API with OpenAPI 3.0 specification
- **CLI Tool**: Command-line interface for all operations with multi-value flag support and quick commands
- **Advanced Filtering**: Filter by date range, multiple environments, multiple owners, status, and groups
- **Persistent Storage**: MySQL database with automatic migrations
- **Many-to-Many Relationships**: Schedules can belong to multiple groups

## ⚠️ Breaking Changes

### v3.0 - Feature Removal (Current)

**Quick Create and Templates features have been removed for codebase simplification.**

**Removed:**
- Quick Create modal and Q keyboard shortcut
- Templates (save/load schedule configurations)
- CLI `schedule quick` command
- CLI `template` commands
- `/api/v1/templates` API endpoints

**Migration:**
- Use standard schedule creation form (Web UI) or `schedule create` (CLI)
- Group assignment moved to standard form
- Templates data will be dropped - export any critical templates before upgrading
- No functional loss - standard creation provides all capabilities

### v2.0 - Multi-Owner and Multi-Environment

**Version 2.0 introduces breaking API changes. If you're upgrading from v1.x, please read the migration guide below.**

**API Changes**

**Owner and Environment fields are now arrays:**

- `owner` (string) → `owners` (array of strings, minimum 1 required)
- `environment` (enum) → `environments` (array of enums, minimum 1 required)

**Before (v1.x):**
```json
{
  "owner": "ops-team",
  "environment": "production"
}
```

**After (v2.0):**
```json
{
  "owners": ["ops-team", "dev-team"],
  "environments": ["production", "staging"]
}
```

### Migration Guide

**Database Migration:**
- Automatic migration preserves existing data
- Single `owner` values → converted to array with one item
- Single `environment` values → converted to array with one item
- Rollback migration available if needed

**API Consumers:**
1. Update POST/PUT requests to send arrays for `owners` and `environments`
2. Update response parsing to handle arrays instead of single values
3. Update query parameters to support multiple values (e.g., `?owner=team1&owner=team2`)
4. All existing schedules will have their single values converted to single-item arrays

**CLI Users:**
- Use multiple flags: `--owner alice --owner bob --env production --env staging`
- Old single-flag usage will fail with validation errors

**Web UI:**
- New tag-based input for owners (semicolon-separated: `alice;bob;charlie`)
- Click-to-add environment tags with color coding
- Inline status editing with keyboard support (Enter to open, Escape to close)

## Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) with **Domain-Driven Design** principles:

```
internal/
├── domain/              # Core business logic
│   ├── schedule/        # Schedule aggregate (Owner, Status, RollbackPlan)
│   └── group/           # Group aggregate (GroupName, Description)
├── application/         # Use cases and ports (interfaces)
│   ├── ports/
│   │   └── input/       # Service interfaces
│   ├── schedule_service.go
│   └── group_service.go
├── adapters/            # Infrastructure implementations
│   ├── input/
│   │   └── http/        # HTTP API handlers
│   └── output/
│       └── mysql/       # Repository implementations
└── infrastructure/      # Cross-cutting concerns (config, logging, db)
```

## Authentication & Authorization

This application uses **Google OAuth 2.0** for authentication and **role-based access control** for authorization.

### User Roles

| Role | Permissions |
|------|-------------|
| **viewer** | View schedules and groups (read-only access) |
| **deployer** | Create, update, and delete own schedules; View all schedules |
| **admin** | All permissions; Approve/deny schedules; Manage users; Modify any schedule |

### Authentication Flow

1. **User clicks "Login with Google"** (Web UI or CLI)
2. **Redirected to Google OAuth consent screen**
3. **Google validates credentials and returns authorization code**
4. **Application exchanges code for Google access token**
5. **Application retrieves user profile from Google**
6. **User is registered or updated in the database**
7. **Application issues JWT token with user claims**
8. **JWT token is used for subsequent API requests**

### Setting Up Google OAuth

#### 1. Create Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google+ API**
4. Go to **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**
5. Configure OAuth consent screen:
   - Application name: "Deployment Tail"
   - User support email: Your email
   - Developer contact: Your email
6. Create OAuth Client ID:
   - Application type: **Web application**
   - Authorized redirect URIs:
     - `http://localhost:8080/auth/google/callback` (for local development)
     - `https://yourdomain.com/auth/google/callback` (for production)
   - For CLI: Add `http://localhost:8081/callback` for the local callback server
7. Save your **Client ID** and **Client Secret**

#### 2. Generate JWT Secret

Generate a secure random secret (minimum 32 characters):

```bash
# Using OpenSSL
openssl rand -base64 32

# Using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"

# Using Go
go run -c "package main; import (\"crypto/rand\"; \"encoding/base64\"; \"fmt\"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.URLEncoding.EncodeToString(b)) }"
```

**⚠️ Important**: Never commit your JWT secret to version control. Keep it secure!

#### 3. Configure Environment Variables

Copy `.env.example` to `.env` and fill in your credentials:

```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id-here.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret-here
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# JWT Configuration
JWT_SECRET=your-generated-jwt-secret-here-minimum-32-characters
JWT_EXPIRY=24h
JWT_ISSUER=deployment-tail
```

### Bootstrap Admin User

After setting up authentication, you need to create an initial admin user:

#### Option 1: Using the Seed Script

```bash
# Set your email as the admin
export ADMIN_EMAIL=your-email@example.com

# Run the seed script
go run scripts/seed_admin_user.go
```

#### Option 2: Direct Database Insert

```sql
-- Replace with your Google ID and email
INSERT INTO users (id, google_id, email, name, role, created_at, updated_at)
VALUES (
  UUID(),
  'your-google-id',  -- Get this from your first login attempt
  'your-email@example.com',
  'Your Name',
  'admin',
  NOW(),
  NOW()
);
```

**How to get your Google ID:**
1. Attempt to log in via the CLI or Web UI
2. Check the server logs for: `User authenticated: google_id=<your-id>`
3. Use that ID in the SQL statement above

#### Option 3: Promote Existing User to Admin

```sql
-- Promote user by email
UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';
```

### CLI Authentication

The CLI requires authentication for all schedule operations:

```bash
# Login (opens browser for OAuth flow)
deployment-tail auth login

# Check authentication status
deployment-tail auth status

# Logout (revokes token)
deployment-tail auth logout

# Force re-authentication (bypass cached token)
deployment-tail schedule list --force-login

# Group favorite commands
deployment-tail group list --owner ops-team              # List all groups
deployment-tail group list --owner ops-team --favorites-only  # List only favorited groups
deployment-tail group favorite <group-id>                # Mark group as favorite
deployment-tail group unfavorite <group-id>              # Remove from favorites
```

**Manual authentication** (for headless environments):

```bash
# Use --manual flag to get URL and paste code manually
deployment-tail auth login --manual
```

### API Authentication

All API endpoints (except `/health` and auth endpoints) require authentication.

**Include JWT token in Authorization header:**

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/schedules
```

**Authentication endpoints:**

- `GET /auth/google/login` - Redirects to Google OAuth
- `GET /auth/google/callback` - OAuth callback (exchanges code for JWT)
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/logout` - Revoke JWT token

**User management endpoints (admin only):**

- `GET /users/me` - Get current user profile
- `GET /users` - List all users (admin only)
- `GET /users/{id}` - Get user by ID (admin only)
- `PUT /users/{id}/role` - Update user role (admin only)

### Permission Matrix

| Operation | Viewer | Deployer | Admin |
|-----------|--------|----------|-------|
| View schedules | ✅ | ✅ | ✅ |
| Create schedule | ❌ | ✅ | ✅ |
| Update own schedule | ❌ | ✅ | ✅ |
| Update any schedule | ❌ | ❌ | ✅ |
| Delete own schedule | ❌ | ✅ | ✅ |
| Delete any schedule | ❌ | ❌ | ✅ |
| Approve/Deny schedule | ❌ | ❌ | ✅ |
| View users | ❌ | ❌ | ✅ |
| Assign user roles | ❌ | ❌ | ✅ |

### Token Behavior

**Token Expiry:**
- Default: 24 hours
- Configurable via `JWT_EXPIRY` environment variable

**Automatic Refresh:**
- CLI automatically refreshes tokens when they have < 1 hour remaining
- Web UI should implement similar refresh logic

**Token Revocation:**
- Logout immediately invalidates the token
- Revoked tokens are blacklisted and synced across server instances
- Blacklist entries are automatically cleaned up after token expiry

**Security Features:**
- Tokens are signed with HS256 algorithm
- Token verification on every request
- Blacklist sync every 60 seconds (configurable)
- Secure token storage in CLI (file permissions: 0600)

## Prerequisites

- Go 1.26+
- MySQL 8.0+
- Docker and Docker Compose (for local development)
- Node.js 18+ (for frontend tests, optional)

## Quick Start

### 1. Set Up Google OAuth (Required for Authentication)

Follow the [Google OAuth setup instructions](#setting-up-google-oauth) to create OAuth credentials.

### 2. Configure Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your credentials
# - Add your Google OAuth Client ID and Secret
# - Generate and add a JWT secret (minimum 32 characters)
nano .env
```

### 3. Start MySQL with Docker Compose

```bash
make docker-up
```

### 4. Build the project

```bash
make build
```

### 5. Run the API server

```bash
# Load environment variables from .env
export $(cat .env | xargs)

# Or set them manually
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=rootpass
export DB_NAME=deployment_schedules
export GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
export GOOGLE_CLIENT_SECRET=your-client-secret
export GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
export JWT_SECRET=your-jwt-secret-minimum-32-characters
export JWT_EXPIRY=24h

# Run server (migrations run automatically)
./bin/server
```

The API server will start on `http://localhost:8080`

### 6. Bootstrap Admin User

Before you can use the application, create an admin user:

```bash
# Set your email
export ADMIN_EMAIL=your-email@example.com

# Run seed script
go run scripts/seed_admin_user.go
```

Alternatively, log in once and promote yourself to admin (see [Bootstrap Admin User](#bootstrap-admin-user) section).

### 7. Authenticate

**Using the CLI:**

```bash
# Login via Google OAuth (opens browser)
./bin/deployment-tail auth login

# Check your authentication status
./bin/deployment-tail auth status
```

**Using the Web UI:**

Open your browser and navigate to `http://localhost:8080/` and click "Login with Google".

### 8. Start Using the Application

The Web UI provides:
- **Authentication**: Secure Google OAuth login with role-based access
- **Dashboard**: View all schedules with tab-based filtering
- **Group Assignment**: Assign schedules to groups during creation with multi-select
- **Quick Group Assignment**: Add schedules to groups directly from the list view
- **Group Management**: Create, edit, delete, and favorite groups with star icons
- **Tab Navigation**: Switch between "All", "Ungrouped", and group-specific views (favorited groups appear first)
- **Schedule Assignment**: Drag-and-drop or bulk assign schedules to groups
- **Create/Edit Forms**: User-friendly modals for all operations
- **Detail View**: Complete schedule information including groups, rollback plans, and audit trail
- **Approval Actions**: Approve or deny schedules directly from the UI (admin only)
- **User Management**: View and manage users and roles (admin only)
- **Responsive Design**: Optimized for desktop, tablet, and mobile devices

## API Endpoints

**Note**: All endpoints require authentication except `/health` and authentication endpoints. Role requirements are noted for each endpoint.

### Authentication (Public)

- `GET /auth/google/login` - Initiate Google OAuth flow
- `GET /auth/google/callback` - OAuth callback (exchanges code for JWT)
- `POST /auth/refresh` - Refresh JWT token (requires valid token)
- `POST /auth/logout` - Revoke JWT token (requires valid token)

### Users

- `GET /users/me` - Get current user profile (any authenticated user)
- `GET /users` - List all users (admin only)
- `GET /users/{id}` - Get user by ID (admin only)
- `PUT /users/{id}/role` - Update user role (admin only)

### Schedules

- `POST /api/v1/schedules` - Create a schedule (deployer, admin)
- `GET /api/v1/schedules` - List schedules with filters (any authenticated user)
- `GET /api/v1/schedules/{id}` - Get a schedule by ID (any authenticated user)
- `PUT /api/v1/schedules/{id}` - Update a schedule (deployer: own only, admin: any)
- `DELETE /api/v1/schedules/{id}` - Delete a schedule (deployer: own only, admin: any)
- `POST /api/v1/schedules/{id}/approve` - Approve a schedule (admin only)
- `POST /api/v1/schedules/{id}/deny` - Deny a schedule (admin only)

### Services

- `GET /api/v1/services/recent` - Get recently used service names (any authenticated user)

### Groups

- `POST /api/v1/groups` - Create a group (deployer, admin)
- `GET /api/v1/groups` - List groups for an owner with favorite status (any authenticated user)
- `GET /api/v1/groups/{id}` - Get a group by ID (any authenticated user)
- `PUT /api/v1/groups/{id}` - Update a group (deployer, admin)
- `DELETE /api/v1/groups/{id}` - Delete a group (deployer, admin)
- `POST /api/v1/groups/{id}/favorite` - Mark group as favorite (any authenticated user)
- `DELETE /api/v1/groups/{id}/favorite` - Remove favorite status (any authenticated user)
- `GET /api/v1/groups/{id}/schedules` - Get all schedules in a group (any authenticated user)

### Schedule-Group Associations

- `POST /api/v1/schedules/{id}/groups` - Assign schedule to multiple groups (deployer, admin)
- `DELETE /api/v1/schedules/{id}/groups/{groupId}` - Unassign schedule from a group (deployer, admin)
- `GET /api/v1/schedules/{id}/groups` - Get all groups for a schedule (any authenticated user)
- `POST /api/v1/groups/{id}/schedules/bulk-assign` - Bulk assign schedules to a group (deployer, admin)

### Health (Public)

- `GET /health` - Health check (no authentication required)

See `api/openapi.yaml` for full API specification.

## API Examples

**Note**: All API examples (except authentication endpoints) require a valid JWT token in the Authorization header.

### Authentication

#### Login via Google OAuth

```bash
# Step 1: Get authorization URL (redirect user to this URL)
curl http://localhost:8080/auth/google/login

# Step 2: After user authorizes, Google redirects to callback with code
# The callback endpoint exchanges the code for a JWT token automatically

# Step 3: Use the JWT token in subsequent requests
export TOKEN="your-jwt-token-here"
```

#### Get Current User Profile

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/users/me
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "deployer",
  "createdAt": "2026-03-30T21:00:00Z",
  "updatedAt": "2026-03-30T21:00:00Z",
  "lastLoginAt": "2026-04-04T10:30:00Z"
}
```

### Create a Group

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Services",
    "description": "All backend microservices",
    "owner": "ops-team"
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Backend Services",
  "description": "All backend microservices",
  "owner": "ops-team",
  "createdAt": "2026-03-30T21:00:00Z",
  "updatedAt": "2026-03-30T21:00:00Z"
}
```

### Create a Schedule

**Requires**: `deployer` or `admin` role

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "scheduledAt": "2026-04-15T20:00:00Z",
    "serviceName": "payment-service",
    "environments": ["production", "staging"],
    "owners": ["ops-team", "payment-team"],
    "description": "Payment gateway update v2.1.0",
    "rollbackPlan": "1. Stop service\n2. Restore DB backup\n3. Deploy v2.0.5\n4. Restart"
  }'
```

Response:
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "scheduledAt": "2026-04-15T20:00:00Z",
  "serviceName": "payment-service",
  "environments": ["production", "staging"],
  "owners": ["ops-team", "payment-team"],
  "status": "created",
  "description": "Payment gateway update v2.1.0",
  "rollbackPlan": "1. Stop service\n2. Restore DB backup\n3. Deploy v2.0.5\n4. Restart",
  "groups": [],
  "createdBy": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "deployer"
  },
  "createdAt": "2026-03-30T21:05:00Z",
  "updatedAt": "2026-03-30T21:05:00Z"
}
```

### Assign Schedule to Groups

```bash
curl -X POST http://localhost:8080/api/v1/schedules/660e8400-e29b-41d4-a716-446655440000/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "groupIds": [
      "550e8400-e29b-41d4-a716-446655440000",
      "770e8400-e29b-41d4-a716-446655440000"
    ],
    "assignedBy": "ops-team"
  }'
```

### List Schedules with Groups

```bash
# Filter by multiple owners
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/schedules?owner=ops-team&owner=payment-team"

# Filter by multiple environments
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/schedules?environment=production&environment=staging"
```

Response:
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "scheduledAt": "2026-04-15T20:00:00Z",
    "serviceName": "payment-service",
    "environments": ["production", "staging"],
    "owners": ["ops-team", "payment-team"],
    "status": "created",
    "groups": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Backend Services",
        "description": "All backend microservices",
        "owner": "ops-team"
      }
    ],
    "createdAt": "2026-03-30T21:05:00Z",
    "updatedAt": "2026-03-30T21:05:00Z"
  }
]
```

### Bulk Assign Schedules to a Group

```bash
curl -X POST http://localhost:8080/api/groups/550e8400-e29b-41d4-a716-446655440000/schedules/bulk-assign \
  -H "Content-Type: application/json" \
  -d '{
    "scheduleIds": [
      "660e8400-e29b-41d4-a716-446655440000",
      "880e8400-e29b-41d4-a716-446655440000",
      "990e8400-e29b-41d4-a716-446655440000"
    ],
    "assignedBy": "ops-team"
  }'
```

## Group Management

Groups allow you to organize schedules into logical collections:

- **Projects**: Group schedules by project (e.g., "Customer Portal", "Payment System")
- **Teams**: Organize by team ownership (e.g., "Backend Team", "Frontend Team")
- **Releases**: Track related deployments (e.g., "Q1 Release", "Hotfix 2.1.1")
- **Environments**: Separate by environment if needed (e.g., "Production Deploys")

### Group Features

- **Many-to-Many**: Schedules can belong to multiple groups
- **Cascade Delete**: Deleting a group removes associations but preserves schedules
- **Unique Names**: Group names must be unique per owner
- **Tab Navigation**: Groups appear as tabs in the Web UI
- **Persistent State**: Active tab is saved in localStorage
- **Ungrouped View**: Special tab to show schedules without any groups

## Testing

### Run All Tests

```bash
# Backend tests (unit + integration)
make test

# Frontend tests
npm test

# With coverage
npm run test:coverage
```

### Backend Tests (114 tests)

```bash
# Run all internal tests
go test ./internal/... -v

# Run only unit tests
go test ./internal/application/... ./internal/domain/... -v

# Run integration tests (requires MySQL)
go test -tags=integration ./internal/adapters/output/mysql/... -v
```

**Test Coverage:**
- **Application Layer**: 15 tests (GroupService + ScheduleService)
- **HTTP Handlers**: 18 tests (all endpoints with mocks)
- **Repository Integration**: 17 tests (MySQL with real DB)
- **Domain Layer**: 64 tests (aggregates and value objects)

### Frontend Tests (30+ tests)

```bash
# Install dependencies
npm install

# Run tests
npm test

# Watch mode
npm run test:watch

# Coverage report
npm run test:coverage
```

**Test Coverage:**
- API Helper Functions (GET, POST, PUT, DELETE)
- LocalStorage Operations (tab persistence)
- Schedule Filtering (all, ungrouped, by group)
- Form Validation (groups and schedules)
- UI State Management (modals, loading)
- Error Handling and Notifications
- Tab Navigation and Switching
- Bulk Operations
- Integration Workflows

## Data Models

### Schedule

| Field | Type | Required | Immutable | Description |
|-------|------|----------|-----------|-------------|
| id | UUID | ✓ | ✓ | Unique identifier |
| scheduledAt | DateTime | ✓ | | When deployment is scheduled |
| serviceName | String | ✓ | | Service to deploy |
| environments | Array[Enum] | ✓ (min 1) | | production/staging/development |
| owners | Array[String] | ✓ (min 1) | | Schedule owners (collaborative) |
| status | Enum | ✓ | | created/approved/denied |
| description | String | | | Optional description |
| rollbackPlan | Text | | | Optional rollback procedure |
| groups | Array | | | Groups this schedule belongs to |
| createdBy | User | ✓ | ✓ | User who created the schedule |
| updatedBy | User | | | User who last updated the schedule |
| createdAt | DateTime | ✓ | ✓ | Creation timestamp |
| updatedAt | DateTime | ✓ | | Last update timestamp |

### User

| Field | Type | Required | Immutable | Description |
|-------|------|----------|-----------|-------------|
| id | UUID | ✓ | ✓ | Unique identifier |
| googleId | String | ✓ | ✓ | Google OAuth ID |
| email | Email | ✓ | | User email from Google |
| name | String | ✓ | | User display name |
| role | Enum | ✓ | | viewer/deployer/admin |
| lastLoginAt | DateTime | | | Last login timestamp |
| createdAt | DateTime | ✓ | ✓ | Registration timestamp |
| updatedAt | DateTime | ✓ | | Last update timestamp |

### Group

| Field | Type | Required | Immutable | Description |
|-------|------|----------|-----------|-------------|
| id | UUID | ✓ | ✓ | Unique identifier |
| name | String(100) | ✓ | | Group name (unique per owner) |
| description | String(500) | | | Optional description |
| owner | String | ✓ | ✓ | Group creator (immutable) |
| createdAt | DateTime | ✓ | ✓ | Creation timestamp |
| updatedAt | DateTime | ✓ | | Last update timestamp |

### Group Favorites

Junction table for tracking user-favorited groups.

| Field | Type | Required | Immutable | Description |
|-------|------|----------|-----------|-------------|
| user_id | UUID | ✓ | ✓ | User who favorited the group (FK to users) |
| group_id | UUID | ✓ | ✓ | Group that was favorited (FK to groups) |
| created_at | DateTime | ✓ | ✓ | When the favorite was added |

**Constraints:**
- Composite primary key: (user_id, group_id) - prevents duplicate favorites
- Foreign key: user_id → users(id) with CASCADE delete
- Foreign key: group_id → groups(id) with CASCADE delete
- Index on user_id for efficient favorite lookups

## Configuration

Configure via environment variables (see `.env.example` for full list):

### Database
- `DB_HOST` - Database host (default: `localhost`)
- `DB_PORT` - Database port (default: `3306`)
- `DB_USER` - Database user (default: `root`)
- `DB_PASSWORD` - Database password (required)
- `DB_NAME` - Database name (default: `deployment_schedules`)

### Server
- `SERVER_HOST` - Server host (default: `0.0.0.0`)
- `SERVER_PORT` - Server port (default: `8080`)

### Google OAuth (Required)
- `GOOGLE_CLIENT_ID` - OAuth 2.0 Client ID from Google Cloud Console (required)
- `GOOGLE_CLIENT_SECRET` - OAuth 2.0 Client Secret (required)
- `GOOGLE_REDIRECT_URL` - OAuth callback URL (e.g., `http://localhost:8080/auth/google/callback`)

### JWT Authentication (Required)
- `JWT_SECRET` - Secret key for signing JWT tokens (minimum 32 characters, required)
- `JWT_EXPIRY` - Token expiry duration (default: `24h`)
- `JWT_ISSUER` - Token issuer identifier (default: `deployment-tail`)

### CLI
- `DEPLOYMENT_TAIL_API` - API endpoint URL (default: `http://localhost:8080`)

**Example `.env` file:**

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=rootpass
DB_NAME=deployment_schedules

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Google OAuth
GOOGLE_CLIENT_ID=123456789-abcdefg.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-secret-here
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# JWT
JWT_SECRET=your-secure-random-secret-minimum-32-characters-long
JWT_EXPIRY=24h
JWT_ISSUER=deployment-tail
```

## Development

### Run tests

```bash
make test          # Backend tests
npm test           # Frontend tests
```

### Format code

```bash
make fmt           # Go code formatting
```

### Generate OpenAPI stubs

```bash
make generate      # Regenerate API handlers from openapi.yaml
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
│   │   ├── schedule/      # Schedule aggregate
│   │   └── group/         # Group aggregate
│   ├── application/       # Application layer (use cases)
│   │   ├── schedule_service.go
│   │   ├── group_service.go
│   │   └── testing.go     # Shared test mocks
│   ├── adapters/          # Adapters (HTTP, MySQL)
│   │   ├── input/
│   │   │   └── http/      # HTTP handlers with tests
│   │   └── output/
│   │       └── mysql/     # Repository implementations
│   └── infrastructure/    # Infrastructure (config, logging, db)
├── migrations/            # Database migrations
├── web/                   # Web UI (HTML, CSS, JavaScript)
│   ├── index.html        # Main HTML page
│   ├── styles.css        # Responsive CSS styles
│   ├── app.js            # Application JavaScript
│   └── tests/            # Frontend tests
│       ├── setup.js      # Jest configuration
│       └── app.test.js   # Test suite (30+ tests)
├── docker-compose.yml     # Docker Compose for local development
├── package.json          # Frontend dependencies and test scripts
└── Makefile              # Build and development tasks
```

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

## Examples

### Complete Group Workflow

```bash
# 1. Create a group
GROUP_ID=$(curl -s -X POST http://localhost:8080/api/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Q1 2026 Release",
    "description": "First quarter feature releases",
    "owner": "release-manager"
  }' | jq -r '.id')

# 2. Create multiple schedules
for service in api web worker; do
  curl -X POST http://localhost:8080/api/schedules \
    -H "Content-Type: application/json" \
    -d "{
      \"scheduledAt\": \"2026-04-01T10:00:00Z\",
      \"serviceName\": \"$service-service\",
      \"environments\": [\"production\"],
      \"owners\": [\"release-manager\"],
      \"description\": \"Q1 release deployment\"
    }"
done

# 3. Get schedule IDs
SCHEDULE_IDS=$(curl -s "http://localhost:8080/api/schedules?owner=release-manager" | jq -r '.[].id')

# 4. Bulk assign to group
curl -X POST "http://localhost:8080/api/groups/$GROUP_ID/schedules/bulk-assign" \
  -H "Content-Type: application/json" \
  -d "{
    \"scheduleIds\": $(echo $SCHEDULE_IDS | jq -R 'split(\"\n\") | map(select(length > 0))'),
    \"assignedBy\": \"release-manager\"
  }"

# 5. List all schedules in the group
curl "http://localhost:8080/api/groups/$GROUP_ID/schedules" | jq
```

### Filter Ungrouped Schedules

```bash
# Get all schedules
curl "http://localhost:8080/api/schedules?owner=ops-team" | \
  jq '[.[] | select(.groups == null or (.groups | length) == 0)]'
```

## Web UI Features

### Sidebar Navigation
- **Persistent Left Sidebar**: Always visible group list (240px on desktop)
- **All Schedules**: View all schedules across all groups
- **Ungrouped**: View schedules not assigned to any group
- **Group List**: All accessible groups (public + your private groups)
- **Visual Indicators**: 🌐 for public groups, 🔒 for private groups, ★ for favorites
- **Favorites First**: Starred groups appear at the top
- **URL-Based**: Bookmarkable URLs (`#all`, `#ungrouped`, `#group/{id}`)
- **Mobile Responsive**: Collapsible sidebar < 768px with hamburger menu

### Date-Grouped Schedules
- **Today**: Schedules for the current day
- **Tomorrow**: Schedules for the next day
- **This Week**: Schedules 2-7 days away
- **Later**: Schedules beyond this week
- **Collapsible Sections**: Click to expand/collapse, state persists
- **Time Display**: HH:MM format for each schedule
- **OVERDUE Badge**: Red badge for past schedules

### Group Management
- **Visibility Control**: Create public (visible to all) or private (owner only) groups
- **Inline Settings**: Gear icon on hover for group owners
- **Create from Sidebar**: "+ New Group" button in sidebar
- **Quick Favorites**: Click star icon to favorite/unfavorite
- **Name and Description**: Validation with character limits
- **Delete Protection**: Schedules preserved when groups are deleted

### Schedule Operations
- Assign schedules to multiple groups simultaneously
- Unassign schedules from groups
- Bulk assignment of multiple schedules to a single group
- Visual group badges on each schedule card

### Responsive Design
- **Desktop (≥ 768px)**: Fixed 240px sidebar + flexible content area
- **Mobile (< 768px)**: Hidden sidebar with hamburger menu overlay
- **Smooth Transitions**: Slide-in/out animations
- **Touch-Friendly**: Tap backdrop or select group to close mobile sidebar
- **Adaptive Layout**: CSS Grid for flexible layouts

## Troubleshooting

### Database Connection Issues

```bash
# Check if MySQL is running
docker ps | grep mysql

# View MySQL logs
docker logs deployment-tail-mysql

# Reset database
make docker-down
make docker-up
```

### API Issues

```bash
# Check server logs
./bin/server

# Test health endpoint
curl http://localhost:8080/health

# Verify API endpoints
curl http://localhost:8080/api/schedules
curl http://localhost:8080/api/groups
```

### Frontend Issues

```bash
# Clear browser cache and localStorage
# Open browser console and run:
localStorage.clear();
location.reload();

# Check browser console for JavaScript errors
# Ensure you're accessing via http://localhost:8080 (not file://)
```

### Group Assignment Errors

**"Schedule created but group assignment failed"**
- The schedule was created successfully but couldn't be assigned to groups
- Check the schedule ID in the error message and manually assign groups via the UI
- Verify the groups still exist and you have access to them

**"Group name not found" (CLI)**
- The group name you specified doesn't exist or you don't have access
- List available groups: `deployment-tail group list`
- Use group ID instead of name for automation: `--groups "<uuid>"`

**"Ambiguous group name" (CLI)**
- Multiple groups match the name you provided
- The error lists all matching groups with their IDs
- Use the specific group ID: `--groups "550e8400-e29b-41d4-a716-446655440000"`

**Orphaned schedule after rollback failure**
- Rare case where schedule creation succeeded, group assignment failed, and deletion also failed
- The error message includes the orphaned schedule ID
- Manually delete the schedule via UI or: `deployment-tail schedule delete <id>`

## Design & UX

- **UX research report**: [`docs/ux-research.md`](docs/ux-research.md) — heuristic evaluation, personas, prioritized findings, design-token inventory, and the backlog of deferred UX work.
- **Design tokens**: Defined in the `:root` block of `web/styles.css` (colors, typography, spacing, radius, shadow, z-index, motion). New web components should consume tokens; legacy components are being migrated incrementally.
- **Accessibility baseline**: The web UI provides ARIA landmark roles, a skip-to-main-content link, a global `:focus-visible` ring, focus traps on all modals, and polite/assertive ARIA live regions for notifications. See `openspec/changes/ux-ui-research-logged-user/specs/ui-accessibility-baseline/spec.md` for requirements.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `make test && npm test`
5. Format code: `make fmt`
6. Submit a pull request

## License

MIT
