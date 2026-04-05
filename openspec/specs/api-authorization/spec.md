## ADDED Requirements

### Requirement: Protect all schedule endpoints with authentication

The system SHALL require valid JWT token for all schedule CRUD operations.

#### Scenario: Authenticated request succeeds
- **WHEN** user with valid JWT calls any schedule endpoint
- **THEN** system processes request normally

#### Scenario: Unauthenticated create request denied
- **WHEN** request to POST /schedules lacks valid token
- **THEN** system returns 401 Unauthorized

#### Scenario: Unauthenticated read request denied
- **WHEN** request to GET /schedules or GET /schedules/{id} lacks valid token
- **THEN** system returns 401 Unauthorized

#### Scenario: Unauthenticated update request denied
- **WHEN** request to PUT /schedules/{id} lacks valid token
- **THEN** system returns 401 Unauthorized

#### Scenario: Unauthenticated delete request denied
- **WHEN** request to DELETE /schedules/{id} lacks valid token
- **THEN** system returns 401 Unauthorized

### Requirement: Role-based access control for schedules

The system SHALL enforce role-based permissions for schedule operations.

#### Scenario: Viewer can read schedules
- **WHEN** user with role "viewer" requests GET /schedules
- **THEN** system returns schedule list

#### Scenario: Viewer cannot create schedules
- **WHEN** user with role "viewer" requests POST /schedules
- **THEN** system returns 403 Forbidden "Insufficient permissions - deployer role required"

#### Scenario: Viewer cannot update schedules
- **WHEN** user with role "viewer" requests PUT /schedules/{id}
- **THEN** system returns 403 Forbidden "Insufficient permissions - deployer role required"

#### Scenario: Viewer cannot delete schedules
- **WHEN** user with role "viewer" requests DELETE /schedules/{id}
- **THEN** system returns 403 Forbidden "Insufficient permissions - deployer role required"

#### Scenario: Deployer can create schedules
- **WHEN** user with role "deployer" requests POST /schedules
- **THEN** system creates schedule

#### Scenario: Deployer can update own schedules
- **WHEN** user with role "deployer" updates schedule they created
- **THEN** system updates schedule

#### Scenario: Deployer cannot update others' schedules
- **WHEN** user with role "deployer" updates schedule created by different user
- **THEN** system returns 403 Forbidden "You can only modify your own schedules"

#### Scenario: Deployer can delete own schedules
- **WHEN** user with role "deployer" deletes schedule they created
- **THEN** system deletes schedule

#### Scenario: Deployer cannot delete others' schedules
- **WHEN** user with role "deployer" deletes schedule created by different user
- **THEN** system returns 403 Forbidden "You can only delete your own schedules"

#### Scenario: Admin can update any schedule
- **WHEN** user with role "admin" updates any schedule
- **THEN** system updates schedule regardless of ownership

#### Scenario: Admin can delete any schedule
- **WHEN** user with role "admin" deletes any schedule
- **THEN** system deletes schedule regardless of ownership

### Requirement: User context in schedule operations

The system SHALL track which user performed each schedule operation.

#### Scenario: Record schedule creator
- **WHEN** authenticated user creates schedule
- **THEN** system stores user ID as schedule owner in created_by field

#### Scenario: Record schedule modifier
- **WHEN** authenticated user updates schedule
- **THEN** system stores user ID in updated_by field and updated_at timestamp

#### Scenario: Owner field auto-populated on create
- **WHEN** authenticated user creates schedule
- **THEN** system automatically adds creator's email to owners array

#### Scenario: Creator included in schedule response
- **WHEN** system returns schedule details
- **THEN** response includes created_by with user email and name

#### Scenario: Last modifier included in schedule response
- **WHEN** system returns updated schedule details
- **THEN** response includes updated_by with user email and name

### Requirement: Protect user management endpoints with admin role

The system SHALL restrict user management operations to admin users only.

#### Scenario: Admin can list users
- **WHEN** user with role "admin" requests GET /users
- **THEN** system returns user list

#### Scenario: Non-admin cannot list users
- **WHEN** user with role "deployer" or "viewer" requests GET /users
- **THEN** system returns 403 Forbidden "Admin role required"

#### Scenario: Admin can assign roles
- **WHEN** user with role "admin" requests PUT /users/{id}/role
- **THEN** system updates user role

#### Scenario: Non-admin cannot assign roles
- **WHEN** user with role "deployer" requests PUT /users/{id}/role
- **THEN** system returns 403 Forbidden "Admin role required"

#### Scenario: Admin can view any user profile
- **WHEN** user with role "admin" requests GET /users/{id}
- **THEN** system returns user profile

#### Scenario: User can view own profile
- **WHEN** authenticated user requests GET /users/{own-id}
- **THEN** system returns own profile

#### Scenario: Non-admin cannot view others' profiles
- **WHEN** user with role "deployer" requests GET /users/{other-id}
- **THEN** system returns 403 Forbidden "Can only view your own profile"

### Requirement: Authentication exempt endpoints

The system SHALL allow unauthenticated access to authentication-related endpoints.

#### Scenario: OAuth login endpoint public
- **WHEN** unauthenticated user accesses GET /auth/google/login
- **THEN** system initiates OAuth flow

#### Scenario: OAuth callback endpoint public
- **WHEN** Google redirects to GET /auth/google/callback
- **THEN** system processes callback without requiring prior authentication

#### Scenario: Health check endpoint public
- **WHEN** unauthenticated request accesses GET /health
- **THEN** system returns health status

#### Scenario: OpenAPI spec endpoint public
- **WHEN** unauthenticated request accesses GET /api/openapi.yaml
- **THEN** system returns API specification
