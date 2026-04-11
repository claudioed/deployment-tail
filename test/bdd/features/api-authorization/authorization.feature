Feature: API Authorization — RBAC and Authentication

  Background:
    Given a clean schedule repository

  # Authentication requirement for schedule endpoints

  @http @spec-api-authorization @auth @smoke
  Scenario: Unauthenticated request to create schedule returns 401
    Given an unauthenticated client
    When I POST "/schedules" with body:
      """
      {"service_name":"api-service","scheduled_at":"2026-06-15T14:00:00Z","owners":["test@example.com"],"environments":["production"]}
      """
    Then the HTTP status is 401

  @http @spec-api-authorization @auth
  Scenario: Unauthenticated request to list schedules returns 401
    Given an unauthenticated client
    When I GET "/schedules"
    Then the HTTP status is 401

  @http @spec-api-authorization @auth
  Scenario: Unauthenticated request to get schedule returns 401
    Given an unauthenticated client
    When I GET "/schedules/00000000-0000-0000-0000-000000000000"
    Then the HTTP status is 401

  @http @spec-api-authorization @auth
  Scenario: Unauthenticated request to update schedule returns 401
    Given an unauthenticated client
    When I PUT "/schedules/00000000-0000-0000-0000-000000000000" with body:
      """
      {"service_name":"updated"}
      """
    Then the HTTP status is 401

  @http @spec-api-authorization @auth
  Scenario: Unauthenticated request to delete schedule returns 401
    Given an unauthenticated client
    When I DELETE "/schedules/00000000-0000-0000-0000-000000000000"
    Then the HTTP status is 401

  # Role-Based Access Control — Viewer

  @service @spec-api-authorization @auth
  Scenario: Viewer can list schedules
    Given I am authenticated as a viewer named "Vera"
    When I list all schedules
    Then no error is returned

  @service @spec-api-authorization @auth
  Scenario: Viewer can get schedule by ID
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "api-service"
    And I am authenticated as a viewer named "Vera"
    When I get the last schedule by ID
    Then no error is returned

  @service @spec-api-authorization @auth
  Scenario: Viewer cannot create schedule
    Given I am authenticated as a viewer named "Vera"
    When I create a schedule with service name "api-service"
    Then an error is returned
    And the error message contains "role"

  @service @spec-api-authorization @auth
  Scenario: Viewer cannot update schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "api-service"
    And I am authenticated as a viewer named "Vera"
    When I update the last schedule description to "Unauthorized update"
    Then an error is returned
    And the error message contains "role"

  @service @spec-api-authorization @auth
  Scenario: Viewer cannot delete schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "api-service"
    And I am authenticated as a viewer named "Vera"
    When I delete the last schedule
    Then an error is returned
    And the error message contains "role"

  # Role-Based Access Control — Deployer

  @service @spec-api-authorization @auth @smoke
  Scenario: Deployer can create schedule
    Given I am authenticated as a deployer named "Dave"
    When I create a schedule with service name "api-service"
    Then no error is returned
    And the last schedule has created by "Dave"

  @service @spec-api-authorization @auth
  Scenario: Deployer can update own schedule
    Given I am authenticated as a deployer named "Dave"
    And I create a schedule with service name "api-service"
    When I update the last schedule description to "Updated by Dave"
    Then no error is returned
    And the last schedule has updated by "Dave"

  @service @spec-api-authorization @auth
  Scenario: Deployer can delete own schedule
    Given I am authenticated as a deployer named "Dave"
    And I create a schedule with service name "api-service"
    When I delete the last schedule
    Then no error is returned

  @service @spec-api-authorization @auth
  Scenario: Deployer cannot update another deployer's schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "alice-service"
    And I am authenticated as a deployer named "Bob"
    When I update the last schedule description to "Bob trying to update Alice's schedule"
    Then an error is returned
    And the error message contains "forbidden"

  @service @spec-api-authorization @auth
  Scenario: Deployer cannot delete another deployer's schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "alice-service"
    And I am authenticated as a deployer named "Bob"
    When I delete the last schedule
    Then an error is returned
    And the error message contains "forbidden"

  # Role-Based Access Control — Admin

  @service @spec-api-authorization @auth
  Scenario: Admin can create schedule
    Given I am authenticated as an admin named "AdminUser"
    When I create a schedule with service name "admin-service"
    Then no error is returned
    And the last schedule has created by "AdminUser"

  @service @spec-api-authorization @auth
  Scenario: Admin can update any schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "alice-service"
    And I am authenticated as an admin named "AdminUser"
    When I update the last schedule description to "Updated by admin"
    Then no error is returned
    And the last schedule has updated by "AdminUser"

  @service @spec-api-authorization @auth
  Scenario: Admin can delete any schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "alice-service"
    And I am authenticated as an admin named "AdminUser"
    When I delete the last schedule
    Then no error is returned

  # User Context Tracking

  @service @spec-api-authorization @auth
  Scenario: Created schedule tracks creating user
    Given I am authenticated as a deployer named "Alice"
    When I create a schedule with service name "tracked-service"
    Then no error is returned
    And the last schedule has created by "Alice"

  @service @spec-api-authorization @auth
  Scenario: Updated schedule tracks updating user
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with service name "tracked-service"
    And I am authenticated as an admin named "AdminUser"
    When I update the last schedule description to "Admin update"
    Then no error is returned
    And the last schedule has created by "Alice"
    And the last schedule has updated by "AdminUser"

  # Admin-Only User Management

  @service @spec-api-authorization @auth
  Scenario: Admin can assign roles
    Given I am authenticated as an admin named "AdminUser"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "deployer" to user "bob@example.com"
    Then no error is returned

  @service @spec-api-authorization @auth
  Scenario: Deployer cannot assign roles
    Given I am authenticated as a deployer named "Dave"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "deployer" to user "bob@example.com"
    Then an error is returned
    And the error message contains "admin role required"

  @service @spec-api-authorization @auth
  Scenario: Viewer cannot assign roles
    Given I am authenticated as a viewer named "Vera"
    And a user "bob@example.com" with role "viewer" exists
    When I assign role "deployer" to user "bob@example.com"
    Then an error is returned
    And the error message contains "admin role required"

  @http @spec-api-authorization @auth
  Scenario: Unauthenticated user cannot assign roles
    Given an unauthenticated client
    When I PUT "/users/bob@example.com/role" with body:
      """
      {"role":"deployer"}
      """
    Then the HTTP status is 401

  # Authentication-Exempt Endpoints

  @http @spec-api-authorization
  Scenario: Health check does not require authentication
    Given an unauthenticated client
    When I GET "/health"
    Then the HTTP status is 200

  @http @spec-api-authorization
  Scenario: OAuth initiation does not require authentication
    Given an unauthenticated client
    When I GET "/auth/google"
    Then the HTTP status is 302

  @http @spec-api-authorization
  Scenario: OAuth callback does not require authentication
    Given an unauthenticated client
    When I GET "/auth/google/callback?code=test&state=test"
    Then the HTTP status is 302
