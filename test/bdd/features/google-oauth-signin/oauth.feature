Feature: Google OAuth Sign-In — OAuth Flow

  Background:
    Given a clean schedule repository

  # OAuth Flow Initiation

  @http @spec-google-oauth-signin @auth
  Scenario: Initiating OAuth redirects to Google
    Given an unauthenticated client
    When I GET "/auth/google"
    Then the HTTP status is 302
    And the Location header contains "accounts.google.com/o/oauth2"
    And the redirect URL contains parameter "client_id"
    And the redirect URL contains parameter "redirect_uri"
    And the redirect URL contains parameter "scope"
    And the redirect URL contains parameter "state"

  @http @spec-google-oauth-signin @auth
  Scenario: OAuth redirect includes correct scopes
    Given an unauthenticated client
    When I GET "/auth/google"
    Then the redirect URL parameter "scope" contains "email"
    And the redirect URL parameter "scope" contains "profile"

  @http @spec-google-oauth-signin @auth
  Scenario: OAuth state parameter is random
    Given an unauthenticated client
    When I GET "/auth/google"
    Then the state parameter is stored in session

  # OAuth Callback Handling

  @service @spec-google-oauth-signin @auth @smoke
  Scenario: Successful OAuth callback registers new user
    Given a Google user "alice@example.com" with name "Alice Smith"
    When I complete the OAuth callback with code "valid-code" and state "valid-state"
    Then no error is returned
    And a new user "alice@example.com" is registered
    And the user has role "viewer"
    And a JWT token is issued

  @service @spec-google-oauth-signin @auth
  Scenario: OAuth callback for existing user updates profile
    Given I am authenticated as a viewer named "Alice"
    And the user "alice@example.com" exists with name "Alice Old"
    When I complete the OAuth callback with updated name "Alice New"
    Then no error is returned
    And the user "alice@example.com" has name "Alice New"
    And the user last login is updated

  @service @spec-google-oauth-signin @auth
  Scenario: OAuth callback with invalid code fails
    Given an unauthenticated client
    When I complete the OAuth callback with code "invalid-code" and state "valid-state"
    Then an error is returned
    And the error message contains "invalid authorization code"

  @service @spec-google-oauth-signin @auth
  Scenario: OAuth callback with mismatched state fails (CSRF protection)
    Given the OAuth state "expected-state" is stored
    When I complete the OAuth callback with code "valid-code" and state "wrong-state"
    Then an error is returned
    And the error message contains "state mismatch"

  @service @spec-google-oauth-signin @auth
  Scenario: OAuth callback without state parameter fails
    Given an unauthenticated client
    When I complete the OAuth callback with code "valid-code" and no state
    Then an error is returned
    And the error message contains "missing state"

  # User Registration on First Sign-In

  @service @spec-google-oauth-signin @auth
  Scenario: First-time Google user is auto-registered
    Given a Google user "newuser@example.com" with name "New User"
    When I complete the OAuth callback
    Then a new user "newuser@example.com" is created
    And the user has role "viewer"
    And the user Google ID is stored

  @service @spec-google-oauth-signin @auth
  Scenario: Auto-registered user defaults to viewer role
    Given a Google user "newuser@example.com" with name "New User"
    When I complete the OAuth callback
    Then the user "newuser@example.com" has role "viewer"

  @service @spec-google-oauth-signin @auth
  Scenario: Existing user is not duplicated on sign-in
    Given I am authenticated as a viewer named "Alice"
    And the user "alice@example.com" exists
    When I complete the OAuth callback for "alice@example.com"
    Then no new user is created
    And only one user with email "alice@example.com" exists

  # Google Profile Sync

  @service @spec-google-oauth-signin @auth
  Scenario: User profile syncs from Google on each sign-in
    Given the user "alice@example.com" exists with name "Alice Old"
    When I complete the OAuth callback with name "Alice Updated"
    Then the user "alice@example.com" has name "Alice Updated"

  @service @spec-google-oauth-signin @auth
  Scenario: User email cannot be changed via OAuth
    Given the user "alice@example.com" exists
    When I complete the OAuth callback for "alice@example.com"
    Then the user email remains "alice@example.com"

  # Error Handling

  @service @spec-google-oauth-signin @auth
  Scenario: Google API error is handled gracefully
    Given the Google OAuth API returns an error
    When I complete the OAuth callback
    Then an error is returned
    And the error message contains "Google authentication failed"

  @service @spec-google-oauth-signin @auth
  Scenario: Missing Google profile data fails
    Given the Google OAuth response is missing email
    When I complete the OAuth callback
    Then an error is returned
    And the error message contains "email required"
