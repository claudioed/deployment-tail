## ADDED Requirements

### Requirement: CLI login command

The system SHALL provide a CLI command to authenticate users via Google OAuth.

#### Scenario: User initiates login
- **WHEN** user executes `deployment-tail login`
- **THEN** CLI opens browser to Google OAuth login page

#### Scenario: Browser login flow
- **WHEN** CLI opens browser for login
- **THEN** user completes Google sign-in in browser

#### Scenario: OAuth callback to local server
- **WHEN** user completes Google OAuth in browser
- **THEN** Google redirects to local callback server started by CLI

#### Scenario: CLI receives token
- **WHEN** local callback server receives OAuth callback
- **THEN** CLI exchanges code for JWT token from API server

#### Scenario: Token stored locally
- **WHEN** CLI receives JWT token
- **THEN** CLI stores token in `~/.deployment-tail/auth.json`

#### Scenario: Login success message
- **WHEN** login completes successfully
- **THEN** CLI displays "Successfully authenticated as {email}"

#### Scenario: Already logged in
- **WHEN** user executes login and valid token exists
- **THEN** CLI displays "Already logged in as {email}" and asks to re-authenticate

#### Scenario: Browser not available
- **WHEN** CLI cannot open browser automatically
- **THEN** CLI displays URL and instructs user to open manually

#### Scenario: OAuth timeout
- **WHEN** user doesn't complete OAuth within 5 minutes
- **THEN** CLI displays "Login timeout - please try again"

### Requirement: CLI logout command

The system SHALL provide a CLI command to sign out and revoke tokens.

#### Scenario: User logs out
- **WHEN** user executes `deployment-tail logout`
- **THEN** CLI revokes token with API server and deletes local token file

#### Scenario: Logout when not logged in
- **WHEN** user executes logout without active session
- **THEN** CLI displays "Not currently logged in"

#### Scenario: Logout success message
- **WHEN** logout completes successfully
- **THEN** CLI displays "Successfully logged out"

#### Scenario: Logout with API error
- **WHEN** token revocation API call fails
- **THEN** CLI still deletes local token and warns "Logged out locally - server revocation failed"

### Requirement: Automatic token inclusion in API requests

The system SHALL automatically include stored JWT token in all CLI API requests.

#### Scenario: Authenticated request includes token
- **WHEN** CLI makes API request and valid token exists
- **THEN** CLI includes token in Authorization: Bearer header

#### Scenario: Command without login
- **WHEN** CLI command requires authentication and no token exists
- **THEN** CLI returns error "Not authenticated - run 'deployment-tail login'"

#### Scenario: Expired token detected
- **WHEN** API returns 401 with "Token expired"
- **THEN** CLI prompts "Token expired - run 'deployment-tail login' to re-authenticate"

#### Scenario: Invalid token detected
- **WHEN** API returns 401 with token validation error
- **THEN** CLI deletes local token and prompts for re-authentication

### Requirement: Automatic token refresh

The system SHALL automatically refresh expiring JWT tokens.

#### Scenario: Token near expiration
- **WHEN** CLI detects stored token expires in less than 1 hour
- **THEN** CLI automatically refreshes token before making API request

#### Scenario: Refresh updates stored token
- **WHEN** CLI refreshes token successfully
- **THEN** CLI updates `~/.deployment-tail/auth.json` with new token

#### Scenario: Refresh failure
- **WHEN** token refresh fails
- **THEN** CLI deletes local token and prompts "Session expired - run 'deployment-tail login'"

#### Scenario: Silent refresh
- **WHEN** CLI refreshes token automatically
- **THEN** CLI does not display refresh messages to user

### Requirement: Display current authentication status

The system SHALL allow users to check authentication status.

#### Scenario: Check auth status
- **WHEN** user executes `deployment-tail auth status`
- **THEN** CLI displays current user email, role, and token expiration time

#### Scenario: Status when not authenticated
- **WHEN** user checks status without valid token
- **THEN** CLI displays "Not authenticated"

#### Scenario: Status shows role
- **WHEN** authenticated user checks status
- **THEN** CLI displays "Logged in as {email} (role: {role})"

#### Scenario: Status shows token expiry
- **WHEN** authenticated user checks status
- **THEN** CLI displays "Token expires: {timestamp}"

### Requirement: Secure token storage

The system SHALL store tokens securely in user's home directory.

#### Scenario: Token file permissions
- **WHEN** CLI stores token in auth.json
- **THEN** file has permissions 0600 (owner read/write only)

#### Scenario: Token file location
- **WHEN** CLI stores token
- **THEN** file is stored at `~/.deployment-tail/auth.json`

#### Scenario: Token file format
- **WHEN** CLI stores token
- **THEN** file contains JSON with fields: token, email, role, expires_at

#### Scenario: Invalid token file
- **WHEN** auth.json exists but contains invalid JSON
- **THEN** CLI deletes file and treats user as not authenticated

#### Scenario: Create directory if needed
- **WHEN** `~/.deployment-tail/` doesn't exist
- **THEN** CLI creates directory with permissions 0700

### Requirement: Handle authentication errors gracefully

The system SHALL provide clear error messages for authentication failures.

#### Scenario: Network error during login
- **WHEN** CLI cannot connect to API server during login
- **THEN** CLI displays "Cannot connect to API server at {url}"

#### Scenario: API server returns error
- **WHEN** API returns error during OAuth exchange
- **THEN** CLI displays error message from API response

#### Scenario: Permission denied error
- **WHEN** API returns 403 Forbidden
- **THEN** CLI displays "Permission denied - insufficient role for this operation"

#### Scenario: User cancelled OAuth
- **WHEN** user closes browser without completing OAuth
- **THEN** CLI displays "Login cancelled by user"
