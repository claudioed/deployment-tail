Feature: Schedule Group Assignment

  Background:
    Given a clean schedule repository

  @service @spec-schedule-group-assignment @smoke
  Scenario: Assign schedule to single group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Project Alpha"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I assign the last schedule to group "Project Alpha"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Assign schedule to multiple groups
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Project Alpha"
    And I create a group named "Team Backend"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I assign the last schedule to groups "Project Alpha, Team Backend"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Assign already assigned schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "My Group"
    When I assign the last schedule to group "My Group"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Assign to non-existent group
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I assign the last schedule to a non-existent group
    Then an error is returned
    And the error message contains "not found"

  @service @spec-schedule-group-assignment
  Scenario: Assign non-existent schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    When I assign a non-existent schedule to group "My Group"
    Then an error is returned
    And the error message contains "not found"

  @service @spec-schedule-group-assignment
  Scenario: Unassign schedule from group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Project Alpha"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "Project Alpha"
    When I unassign the last schedule from group "Project Alpha"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Unassign from non-member group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I unassign the last schedule from group "My Group"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Unassign from non-existent group
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I unassign the last schedule from a non-existent group
    Then an error is returned
    And the error message contains "not found"

  @service @spec-schedule-group-assignment @smoke
  Scenario: Get groups for schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Group A"
    And I create a group named "Group B"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to groups "Group A, Group B"
    When I list groups for the last schedule
    Then no error is returned
    And the group list includes "Group A"
    And the group list includes "Group B"

  @service @spec-schedule-group-assignment
  Scenario: Get groups for ungrouped schedule
    Given I am authenticated as a deployer named "Alice"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I list groups for the last schedule
    Then no error is returned
    And the group list is empty

  @service @spec-schedule-group-assignment
  Scenario: Get groups ordered by name
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Zebra"
    And I create a group named "Apple"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to groups "Zebra, Apple"
    When I list groups for the last schedule
    Then the group list order is "Apple, Zebra"

  @service @spec-schedule-group-assignment @smoke
  Scenario: Get schedules in group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Project Alpha"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "Project Alpha"
    When I list schedules in group "Project Alpha"
    Then no error is returned
    And the schedule list includes the last schedule

  @service @spec-schedule-group-assignment
  Scenario: Get schedules in empty group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Empty Group"
    When I list schedules in group "Empty Group"
    Then no error is returned
    And the schedule list is empty

  @service @spec-schedule-group-assignment
  Scenario: Bulk assign schedules
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Bulk Group"
    And I create 3 schedules
    When I bulk assign the 3 schedules to group "Bulk Group"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Bulk assign empty list
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    When I bulk assign 0 schedules to group "My Group"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Bulk unassign schedules
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Bulk Group"
    And I create 3 schedules
    And I bulk assign the 3 schedules to group "Bulk Group"
    When I bulk unassign the 3 schedules from group "Bulk Group"
    Then no error is returned

  @service @spec-schedule-group-assignment
  Scenario: Bulk unassign with non-members
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    And I create 2 schedules
    When I bulk unassign the 2 schedules from group "My Group"
    Then no error is returned

  @service @spec-schedule-group-assignment @smoke
  Scenario: List ungrouped schedules
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Some Group"
    And I create a schedule with:
      | service_name | grouped-service      |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "Some Group"
    And I create a schedule with:
      | service_name | ungrouped-service    |
      | scheduled_at | 2026-06-16T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I list ungrouped schedules
    Then no error is returned
    And the schedule list includes service "ungrouped-service"
    And the schedule list does not include service "grouped-service"

  @service @spec-schedule-group-assignment
  Scenario: No ungrouped schedules
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "My Group"
    When I list ungrouped schedules
    Then no error is returned
    And the schedule list is empty

  @service @spec-schedule-group-assignment
  Scenario: Delete schedule removes associations
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "My Group"
    When I delete the last schedule
    And I list schedules in group "My Group"
    Then the schedule list is empty

  @service @spec-schedule-group-assignment
  Scenario: Delete group removes associations
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "ToDelete"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    And I assign the last schedule to group "ToDelete"
    When I delete the last group
    And I list groups for the last schedule
    Then the group list is empty
