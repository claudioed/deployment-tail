Feature: JWT Session Management — Token Lifecycle

  Background:
    Given a clean schedule repository

  # JWT Issuance

  @service @spec-jwt-session-management @auth @smoke
  Scenario: Successful login issues JWT token
    Given I am authenticated as a deployer named "Alice"
    When I request my user profile
    Then no error is returned
    And a JWT token is present

  @service @spec-jwt-session-management @auth
  Scenario: JWT contains user claims
    Given I am authenticated as a deployer named "Alice"
    When I decode the current JWT
    Then the JWT claim "email" equals "alice@example.com"
    And the JWT claim "role" equals "deployer"
    And the JWT has claim "exp"
    And the JWT has claim "iat"

  # JWT Validation

  @http @spec-jwt-session-management @auth
  Scenario: Missing Authorization header returns 401
    Given an unauthenticated client
    When I GET "/schedules"
    Then the HTTP status is 401
    And the response JSON field "error" contains "missing"

  @http @spec-jwt-session-management @auth
  Scenario: Malformed Authorization header returns 401
    Given I set Authorization header to "InvalidFormat"
    When I GET "/schedules"
    Then the HTTP status is 401
    And the response JSON field "error" contains "invalid"

  @http @spec-jwt-session-management @auth
  Scenario: Invalid Bearer token format returns 401
    Given I set Authorization header to "Bearer not-a-jwt"
    When I GET "/schedules"
    Then the HTTP status is 401
    And the response JSON field "error" contains "invalid"

  @http @spec-jwt-session-management @auth
  Scenario: Expired JWT returns 401
    Given I am authenticated as a deployer named "Alice"
    When I use an expired JWT
    And I GET "/schedules"
    Then the HTTP status is 401
    And the response JSON field "error" contains "expired"

  @http @spec-jwt-session-management @auth
  Scenario: Revoked JWT returns 401
    Given I am authenticated as a deployer named "Alice"
    When I revoke the current JWT
    And I GET "/schedules"
    Then the HTTP status is 401
    And the response JSON field "error" contains "revoked"

  @http @spec-jwt-session-management @auth
  Scenario: JWT with invalid signature returns 401
    Given I am authenticated as a deployer named "Alice"
    When I tamper with the JWT signature
    And I GET "/schedules"
    Then the HTTP status is 401
    And the response JSON field "error" contains "invalid"

  # JWT Refresh

  @service @spec-jwt-session-management @auth
  Scenario: Valid JWT can be refreshed
    Given I am authenticated as a deployer named "Alice"
    When I refresh the current JWT
    Then no error is returned
    And a new JWT token is issued

  @service @spec-jwt-session-management @auth
  Scenario: Refreshed JWT has updated expiry
    Given I am authenticated as a deployer named "Alice"
    And I wait 2 seconds
    When I refresh the current JWT
    Then the new JWT expiry is later than the old JWT expiry

  @service @spec-jwt-session-management @auth
  Scenario: Expired JWT cannot be refreshed
    Given I am authenticated as a deployer named "Alice"
    When I use an expired JWT
    And I refresh the current JWT
    Then an error is returned
    And the error message contains "expired"

  @service @spec-jwt-session-management @auth
  Scenario: Revoked JWT cannot be refreshed
    Given I am authenticated as a deployer named "Alice"
    When I revoke the current JWT
    And I refresh the current JWT
    Then an error is returned
    And the error message contains "revoked"

  # JWT Revocation

  @service @spec-jwt-session-management @auth @smoke
  Scenario: User can revoke their own JWT
    Given I am authenticated as a deployer named "Alice"
    When I revoke the current JWT
    Then no error is returned

  @service @spec-jwt-session-management @auth
  Scenario: Revoked JWT is blacklisted
    Given I am authenticated as a deployer named "Alice"
    When I revoke the current JWT
    Then the JWT is in the revocation store

  @service @spec-jwt-session-management @auth
  Scenario: Using revoked JWT after logout fails
    Given I am authenticated as a deployer named "Alice"
    And I revoke the current JWT
    When I request my user profile
    Then an error is returned
    And the error message contains "revoked"

  @service @spec-jwt-session-management @auth
  Scenario: Revocation is idempotent
    Given I am authenticated as a deployer named "Alice"
    When I revoke the current JWT
    And I revoke the current JWT
    Then no error is returned

  # Token Expiry Configuration

  @service @spec-jwt-session-management
  Scenario: JWT expiry is configurable
    Given the JWT expiry is set to 1 hour
    When I am authenticated as a deployer named "Alice"
    Then the JWT expires in 1 hour

  @service @spec-jwt-session-management
  Scenario: Default JWT expiry is 24 hours
    Given the JWT expiry is set to default
    When I am authenticated as a deployer named "Alice"
    Then the JWT expires in 24 hours

  # Revocation Store Cleanup

  @service @spec-jwt-session-management
  Scenario: Expired revocation entries are cleaned up
    Given I am authenticated as a deployer named "Alice"
    And I revoke the current JWT
    When the revocation cleanup runs
    Then expired revocation entries are removed
