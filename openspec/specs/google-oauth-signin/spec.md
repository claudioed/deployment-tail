## ADDED Requirements

### Requirement: Initiate Google OAuth sign-in flow

The system SHALL provide an endpoint to initiate the Google OAuth 2.0 authorization flow.

#### Scenario: User requests sign-in
- **WHEN** unauthenticated user navigates to `/auth/google/login`
- **THEN** system redirects to Google's authorization page with correct client ID and scopes

#### Scenario: Redirect includes required scopes
- **WHEN** system redirects to Google OAuth
- **THEN** request includes scopes: email, profile, openid

#### Scenario: Redirect includes callback URL
- **WHEN** system redirects to Google OAuth
- **THEN** request includes the configured redirect URI matching Google Console settings

#### Scenario: State parameter for CSRF protection
- **WHEN** system redirects to Google OAuth
- **THEN** request includes a randomly generated state parameter stored in session

### Requirement: Handle Google OAuth callback

The system SHALL handle the OAuth callback from Google and exchange authorization code for user information.

#### Scenario: Successful OAuth callback
- **WHEN** Google redirects back with valid authorization code
- **THEN** system exchanges code for access token and retrieves user profile

#### Scenario: Extract user information
- **WHEN** system receives access token from Google
- **THEN** system retrieves user's Google ID, email, and display name

#### Scenario: Validate state parameter
- **WHEN** callback includes state parameter
- **THEN** system validates it matches the value stored in session

#### Scenario: Invalid state parameter
- **WHEN** callback includes mismatched state parameter
- **THEN** system returns error "Invalid state parameter - possible CSRF attack"

#### Scenario: Missing authorization code
- **WHEN** callback lacks authorization code
- **THEN** system returns error "Authorization code not provided"

#### Scenario: Google API error
- **WHEN** Google returns error during token exchange
- **THEN** system returns error with Google's error message

#### Scenario: Invalid or expired authorization code
- **WHEN** authorization code is invalid or expired
- **THEN** system returns error "Failed to exchange authorization code"

### Requirement: User registration on first sign-in

The system SHALL automatically register new users on their first successful Google OAuth sign-in.

#### Scenario: New user first sign-in
- **WHEN** user with Google ID "12345" signs in and doesn't exist in system
- **THEN** system creates new user record with Google ID, email, name, and default role "viewer"

#### Scenario: Existing user sign-in
- **WHEN** user with existing Google ID signs in
- **THEN** system retrieves existing user record without creating duplicate

#### Scenario: Email update on sign-in
- **WHEN** existing user signs in and email changed in Google account
- **THEN** system updates user's email to match Google profile

#### Scenario: Name update on sign-in
- **WHEN** existing user signs in and name changed in Google account
- **THEN** system updates user's display name to match Google profile

### Requirement: Validate Google OAuth configuration

The system SHALL validate required OAuth configuration on startup.

#### Scenario: Missing client ID
- **WHEN** server starts without GOOGLE_CLIENT_ID environment variable
- **THEN** system fails to start with error "GOOGLE_CLIENT_ID is required"

#### Scenario: Missing client secret
- **WHEN** server starts without GOOGLE_CLIENT_SECRET environment variable
- **THEN** system fails to start with error "GOOGLE_CLIENT_SECRET is required"

#### Scenario: Missing redirect URL
- **WHEN** server starts without GOOGLE_REDIRECT_URL environment variable
- **THEN** system fails to start with error "GOOGLE_REDIRECT_URL is required"

#### Scenario: Invalid redirect URL format
- **WHEN** GOOGLE_REDIRECT_URL is not a valid HTTP/HTTPS URL
- **THEN** system fails to start with error "GOOGLE_REDIRECT_URL must be a valid URL"
