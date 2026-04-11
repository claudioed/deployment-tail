Feature: User Management — CRUD and Role Assignment

  Background:
    Given a clean schedule repository

  # User Profile Retrieval

  @service @spec-user-management @auth @smoke
  Scenario: User can retrieve own profile
    Given I am authenticated as a deployer named "Alice"
    When I request my user profile
    Then no error is returned
    And the profile email is "alice@example.com"
    And the profile role is "deployer"

  @http @spec-user-management @auth
  Scenario: Unauthenticated user cannot retrieve profile
    Given an unauthenticated client
    When I GET "/users/me"
    Then the HTTP status is 401

  @service @spec-user-management @auth
  Scenario: Profile includes last login timestamp
    Given I am authenticated as a deployer named "Alice"
    When I request my user profile
    Then the profile has last login timestamp

  @service @spec-user-management @auth
  Scenario: Profile includes Google ID
    Given I am authenticated as a deployer named "Alice"
    When I request my user profile
    Then the profile has Google ID

  # User Listing

  @service @spec-user-management @auth
  Scenario: Admin can list all users
    Given I am authenticated as an admin named "AdminUser"
    And a user "alice@example.com" with role "deployer" exists
    And a user "bob@example.com" with role "viewer" exists
    When I list all users
    Then no error is returned
    And the user list includes "alice@example.com"
    And the user list includes "bob@example.com"

  @service @spec-user-management @auth
  Scenario: Deployer cannot list users
    Given I am authenticated as a deployer named "Dave"
    When I list all users
    Then an error is returned
    And the error message contains "admin role required"

  @service @spec-user-management @auth
  Scenario: Viewer cannot list users
    Given I am authenticated as a viewer named "Vera"
    When I list all users
    Then an error is returned
    And the error message contains "admin role required"

  @http @spec-user-management @auth
  Scenario: Unauthenticated user cannot list users
    Given an unauthenticated client
    When I GET "/users"
    Then the HTTP status is 401

  # Role Assignment

  @service @spec-user-management @auth @smoke
  Scenario: Admin can assign viewer role
    Given I am authenticated as an admin named "AdminUser"
    And a user "alice@example.com" with role "viewer" exists
    When I assign role "deployer" to user "alice@example.com"
    Then no error is returned
    And the user "alice@example.com" has role "deployer"

  @service @spec-user-management @auth
  Scenario: Admin can assign deployer role
    Given I am authenticated as an admin named "AdminUser"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "deployer" to user "bob@example.com"
    Then no error is returned
    And the user "bob@example.com" has role "deployer"

  @service @spec-user-management @auth
  Scenario: Admin can assign admin role
    Given I am authenticated as an admin named "AdminUser"
    And a user "carol@example.com" with role "deployer" exists
    When I assign role "admin" to user "carol@example.com"
    Then no error is returned
    And the user "carol@example.com" has role "admin"

  @service @spec-user-management @auth
  Scenario: Deployer cannot assign roles
    Given I am authenticated as a deployer named "Dave"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "deployer" to user "bob@example.com"
    Then an error is returned
    And the error message contains "admin role required"

  @service @spec-user-management @auth
  Scenario: Viewer cannot assign roles
    Given I am authenticated as a viewer named "Vera"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "deployer" to user "bob@example.com"
    Then an error is returned
    And the error message contains "admin role required"

  @http @spec-user-management @auth
  Scenario: Unauthenticated user cannot assign roles
    Given an unauthenticated client
    When I PUT "/users/bob@example.com/role" with body:
      """
      {"role":"deployer"}
      """
    Then the HTTP status is 401

  # Role Assignment Validation

  @service @spec-user-management @auth
  Scenario: Invalid role assignment fails
    Given I am authenticated as an admin named "AdminUser"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "invalid-role" to user "bob@example.com"
    Then an error is returned
    And the error message contains "invalid role"

  @service @spec-user-management @auth
  Scenario: Assigning role to non-existent user fails
    Given I am authenticated as an admin named "AdminUser"
    When I assign role "deployer" to user "nonexistent@example.com"
    Then an error is returned
    And the error message contains "user not found"

  # User Auto-Registration

  @service @spec-user-management @auth
  Scenario: New Google sign-in auto-creates user
    Given a Google user "newuser@example.com" with name "New User"
    When I complete the OAuth callback
    Then a new user "newuser@example.com" is created
    And the user has role "viewer"

  @service @spec-user-management @auth
  Scenario: Auto-registered user has default viewer role
    Given a Google user "newuser@example.com" with name "New User"
    When I complete the OAuth callback
    Then the user "newuser@example.com" has role "viewer"

  # User Activity Tracking

  @service @spec-user-management @auth
  Scenario: Last login is updated on authentication
    Given I am authenticated as a deployer named "Alice"
    And I wait 1 second
    When I authenticate again as "Alice"
    Then the user "alice@example.com" last login is updated

  @service @spec-user-management @auth
  Scenario: Created schedule tracks creator
    Given I am authenticated as a deployer named "Alice"
    When I create a schedule with service name "tracked-service"
    Then the schedule created by is "Alice"

  @service @spec-user-management @auth
  Scenario: Updated schedule tracks updater
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "tracked-service"
    And I am authenticated as an admin named "AdminUser"
    When I update the last schedule description to "Admin update"
    Then the schedule updated by is "AdminUser"

  # User Data Validation

  @service @spec-user-management @auth
  Scenario: Email must be valid format
    Given I am authenticated as an admin named "AdminUser"
    When I create a user with email "invalid-email"
    Then an error is returned
    And the error message contains "invalid email"

  @service @spec-user-management @auth
  Scenario: User name cannot be empty
    Given I am authenticated as an admin named "AdminUser"
    When I create a user with empty name
    Then an error is returned
    And the error message contains "name cannot be empty"

  @service @spec-user-management @auth
  Scenario: Google ID cannot be empty
    Given I am authenticated as an admin named "AdminUser"
    When I create a user with empty Google ID
    Then an error is returned
    And the error message contains "Google ID cannot be empty"

  # Profile Update

  @service @spec-user-management @auth
  Scenario: User profile is updated on each sign-in
    Given the user "alice@example.com" exists with name "Alice Old"
    When I complete the OAuth callback with name "Alice New"
    Then the user "alice@example.com" has name "Alice New"

  @service @spec-user-management @auth
  Scenario: User email remains immutable
    Given the user "alice@example.com" exists
    When I complete the OAuth callback for "alice@example.com"
    Then the user email remains "alice@example.com"
