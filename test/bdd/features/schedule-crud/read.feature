Feature: Schedule CRUD — Read

  Background:
    Given a clean schedule repository

  @service @spec-schedule-crud @smoke
  Scenario: Get schedule by ID
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I retrieve the schedule by ID
    Then no error is returned
    And the schedule has service name "api-service"

  @service @spec-schedule-crud @wip
  Scenario: Get non-existent schedule returns error
    Given I am authenticated as a deployer named "Bob"
    When I retrieve the schedule by ID
    Then an error is returned
    And the error message contains "schedule not found"
