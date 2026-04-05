## ADDED Requirements

### Requirement: User profile retrieval

The system SHALL allow authenticated users to retrieve user profiles.

#### Scenario: Get own profile
- **WHEN** authenticated user requests GET /users/me
- **THEN** system returns user's profile with ID, email, name, role, Google ID, created_at, updated_at

#### Scenario: Admin gets any user profile
- **WHEN** admin requests GET /users/{id}
- **THEN** system returns requested user's profile

#### Scenario: Non-admin restricted to own profile
- **WHEN** non-admin user requests GET /users/{other-id}
- **THEN** system returns 403 Forbidden

#### Scenario: Non-existent user profile
- **WHEN** request for user ID that doesn't exist
- **THEN** system returns 404 Not Found "User not found"

#### Scenario: Profile excludes sensitive data
- **WHEN** system returns user profile
- **THEN** response excludes internal implementation details (e.g., password hashes, OAuth tokens)

### Requirement: List all users

The system SHALL allow admins to list all registered users.

#### Scenario: Admin lists all users
- **WHEN** admin requests GET /users
- **THEN** system returns paginated list of all users with ID, email, name, role

#### Scenario: User list pagination
- **WHEN** admin requests GET /users?page=2&limit=20
- **THEN** system returns second page with up to 20 users

#### Scenario: Default pagination limits
- **WHEN** admin requests GET /users without pagination params
- **THEN** system returns first page with default limit of 50 users

#### Scenario: Filter users by role
- **WHEN** admin requests GET /users?role=admin
- **THEN** system returns only users with admin role

#### Scenario: Search users by email
- **WHEN** admin requests GET /users?email=john
- **THEN** system returns users whose email contains "john"

#### Scenario: Non-admin cannot list users
- **WHEN** user with role "deployer" requests GET /users
- **THEN** system returns 403 Forbidden "Admin role required"

### Requirement: Assign user roles

The system SHALL allow admins to assign roles to users.

#### Scenario: Admin assigns deployer role
- **WHEN** admin requests PUT /users/{id}/role with body {"role": "deployer"}
- **THEN** system updates user role to "deployer"

#### Scenario: Admin assigns admin role
- **WHEN** admin requests PUT /users/{id}/role with body {"role": "admin"}
- **THEN** system updates user role to "admin"

#### Scenario: Admin downgrades user to viewer
- **WHEN** admin requests PUT /users/{id}/role with body {"role": "viewer"}
- **THEN** system updates user role to "viewer"

#### Scenario: Invalid role value
- **WHEN** admin requests role assignment with invalid role "superuser"
- **THEN** system returns 400 Bad Request "Invalid role: must be viewer, deployer, or admin"

#### Scenario: Non-admin cannot assign roles
- **WHEN** user with role "deployer" requests PUT /users/{id}/role
- **THEN** system returns 403 Forbidden "Admin role required"

#### Scenario: Cannot assign role to non-existent user
- **WHEN** admin assigns role to user ID that doesn't exist
- **THEN** system returns 404 Not Found "User not found"

#### Scenario: Role change revokes active tokens
- **WHEN** admin changes user's role
- **THEN** system revokes all active JWT tokens for that user

#### Scenario: Self-service role change prohibited
- **WHEN** admin attempts to change their own role
- **THEN** system returns 403 Forbidden "Cannot modify your own role"

### Requirement: User registration during OAuth

The system SHALL automatically create user records during Google OAuth sign-in.

#### Scenario: First-time user auto-registered
- **WHEN** user completes Google OAuth for first time
- **THEN** system creates user with Google ID, email, name from OAuth profile and default role "viewer"

#### Scenario: Duplicate Google ID prevented
- **WHEN** system creates user and Google ID already exists
- **THEN** system retrieves existing user instead of creating duplicate

#### Scenario: Email uniqueness enforced
- **WHEN** user signs in and email already registered to different Google ID
- **THEN** system returns error "Email already registered with different account"

#### Scenario: Default role for new users
- **WHEN** new user auto-registered via OAuth
- **THEN** system assigns role "viewer"

#### Scenario: User creation timestamp
- **WHEN** new user created
- **THEN** system records created_at timestamp

### Requirement: Track user activity

The system SHALL maintain user activity timestamps.

#### Scenario: Record last sign-in time
- **WHEN** user completes OAuth sign-in
- **THEN** system updates user's last_login_at timestamp

#### Scenario: Update profile modification time
- **WHEN** admin modifies user's role
- **THEN** system updates user's updated_at timestamp

#### Scenario: Activity included in profile response
- **WHEN** system returns user profile
- **THEN** response includes last_login_at and updated_at

### Requirement: Validate user data

The system SHALL validate user profile data.

#### Scenario: Valid email format required
- **WHEN** user created with invalid email format
- **THEN** system returns error "Invalid email format"

#### Scenario: Email required
- **WHEN** user created without email
- **THEN** system returns error "Email is required"

#### Scenario: Google ID required
- **WHEN** user created without Google ID
- **THEN** system returns error "Google ID is required"

#### Scenario: Name required
- **WHEN** user created without name
- **THEN** system returns error "Name is required"

#### Scenario: Email max length
- **WHEN** email exceeds 255 characters
- **THEN** system returns error "Email too long - maximum 255 characters"

#### Scenario: Name max length
- **WHEN** name exceeds 255 characters
- **THEN** system returns error "Name too long - maximum 255 characters"
