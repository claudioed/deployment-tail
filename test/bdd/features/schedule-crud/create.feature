Feature: Schedule CRUD — Create

  Background:
    Given a clean schedule repository

  @service @spec-schedule-crud @smoke
  Scenario: Create schedule with required fields
    Given I am authenticated as a deployer named "Alice"
    When I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    Then no error is returned
    And the last schedule has service name "api-service"
    And the last schedule status is "created"

  @http @spec-schedule-crud @auth @smoke
  Scenario: Unauthenticated user cannot create
    Given an unauthenticated client
    When I POST "/schedules" with body:
      """
      {"service_name":"api","scheduled_at":"2026-06-15T14:00:00Z","owners":["a@b.com"],"environments":["production"]}
      """
    Then the HTTP status is 401

  @http @spec-schedule-crud @auth
  Scenario: Viewer role cannot create
    Given I am authenticated as a viewer named "Vera"
    When I POST "/schedules" with body:
      """
      {"service_name":"api","scheduled_at":"2026-06-15T14:00:00Z","owners":["vera@x.com"],"environments":["production"]}
      """
    Then the HTTP status is 400
    And the response JSON field "message" contains "requires deployer or admin role"

  @service @spec-schedule-crud
  Scenario: Create schedule with description
    Given I am authenticated as a deployer named "Bob"
    When I create a schedule with:
      | service_name | auth-service          |
      | scheduled_at | 2026-06-16T10:00:00Z  |
      | environments | staging               |
      | owners       | bob@example.com       |
      | description  | Staging deployment    |
    Then no error is returned
    And the last schedule has service name "auth-service"

  @service @spec-schedule-crud
  Scenario: Create schedule with multiple environments
    Given I am authenticated as a deployer named "Charlie"
    When I create a schedule with:
      | service_name | data-pipeline             |
      | scheduled_at | 2026-06-17T08:00:00Z      |
      | environments | staging,production        |
      | owners       | charlie@example.com       |
    Then no error is returned
    And the last schedule has service name "data-pipeline"
