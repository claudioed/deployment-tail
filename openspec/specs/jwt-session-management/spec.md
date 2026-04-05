## ADDED Requirements

### Requirement: Issue JWT access token after authentication

The system SHALL issue a signed JWT access token after successful Google OAuth authentication.

#### Scenario: Successful authentication returns JWT
- **WHEN** user completes Google OAuth flow successfully
- **THEN** system returns JWT access token containing user ID, email, and role

#### Scenario: JWT includes standard claims
- **WHEN** system issues JWT
- **THEN** token includes claims: sub (user ID), email, role, iat (issued at), exp (expiration)

#### Scenario: JWT signed with configured secret
- **WHEN** system issues JWT
- **THEN** token is signed using HS256 algorithm with JWT_SECRET

#### Scenario: JWT expiration time
- **WHEN** system issues JWT
- **THEN** token expires after configured JWT_EXPIRY duration (default: 24 hours)

#### Scenario: JWT includes issuer claim
- **WHEN** system issues JWT
- **THEN** token includes iss claim set to "deployment-tail"

### Requirement: Validate JWT access token

The system SHALL validate JWT tokens on protected endpoints.

#### Scenario: Valid token allows access
- **WHEN** request includes valid, unexpired JWT in Authorization header
- **THEN** system extracts user context and allows request

#### Scenario: Missing token
- **WHEN** request to protected endpoint lacks Authorization header
- **THEN** system returns 401 Unauthorized "Missing authorization token"

#### Scenario: Invalid token format
- **WHEN** Authorization header doesn't use Bearer scheme
- **THEN** system returns 401 Unauthorized "Invalid authorization format"

#### Scenario: Malformed JWT
- **WHEN** token is not a valid JWT structure
- **THEN** system returns 401 Unauthorized "Invalid token format"

#### Scenario: Invalid signature
- **WHEN** token signature doesn't match JWT_SECRET
- **THEN** system returns 401 Unauthorized "Invalid token signature"

#### Scenario: Expired token
- **WHEN** token exp claim is in the past
- **THEN** system returns 401 Unauthorized "Token expired"

#### Scenario: Token missing required claims
- **WHEN** token lacks sub, email, or role claims
- **THEN** system returns 401 Unauthorized "Invalid token claims"

#### Scenario: User ID in token doesn't exist
- **WHEN** token's sub claim references non-existent user
- **THEN** system returns 401 Unauthorized "User not found"

### Requirement: Refresh JWT access token

The system SHALL allow users to refresh their JWT before expiration.

#### Scenario: Refresh valid token
- **WHEN** user requests token refresh with valid unexpired token
- **THEN** system issues new JWT with extended expiration

#### Scenario: Refresh updates user role
- **WHEN** user's role changed since token issued and requests refresh
- **THEN** new token includes updated role claim

#### Scenario: Refresh updates user email
- **WHEN** user's email changed since token issued and requests refresh
- **THEN** new token includes updated email claim

#### Scenario: Cannot refresh expired token
- **WHEN** user requests refresh with expired token
- **THEN** system returns 401 Unauthorized "Token expired - please sign in again"

#### Scenario: Refresh with revoked token
- **WHEN** user requests refresh with revoked token
- **THEN** system returns 401 Unauthorized "Token has been revoked"

### Requirement: Revoke JWT access token

The system SHALL provide mechanism to revoke JWT tokens before expiration.

#### Scenario: User logout revokes token
- **WHEN** user requests logout with valid token
- **THEN** system adds token to revocation list and returns success

#### Scenario: Revoked token rejected
- **WHEN** request uses revoked token
- **THEN** system returns 401 Unauthorized "Token has been revoked"

#### Scenario: Admin revokes user's tokens
- **WHEN** admin revokes all tokens for specific user ID
- **THEN** system adds all active tokens for that user to revocation list

#### Scenario: Revocation list cleanup
- **WHEN** system performs periodic cleanup
- **THEN** system removes revocation entries for tokens past expiration time

### Requirement: Validate JWT configuration

The system SHALL validate JWT configuration on startup.

#### Scenario: Missing JWT secret
- **WHEN** server starts without JWT_SECRET environment variable
- **THEN** system fails to start with error "JWT_SECRET is required"

#### Scenario: Weak JWT secret
- **WHEN** JWT_SECRET is shorter than 32 characters
- **THEN** system fails to start with error "JWT_SECRET must be at least 32 characters"

#### Scenario: Invalid JWT expiry format
- **WHEN** JWT_EXPIRY cannot be parsed as duration (e.g., "24h", "30m")
- **THEN** system fails to start with error "JWT_EXPIRY must be a valid duration"

#### Scenario: Default JWT expiry
- **WHEN** JWT_EXPIRY not configured
- **THEN** system defaults to 24 hours
