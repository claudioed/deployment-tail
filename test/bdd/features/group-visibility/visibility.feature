Feature: Group Visibility

  Background:
    Given a clean schedule repository

  @service @spec-group-visibility @smoke
  Scenario: Owner creates private group by default
    Given I am authenticated as a deployer named "Alice"
    When I create a group named "Private Team"
    Then no error is returned
    And the last group has visibility "private"

  @service @spec-group-visibility @smoke
  Scenario: Owner creates public group
    Given I am authenticated as a deployer named "Bob"
    When I create a group named "Public Team" with visibility "public"
    Then no error is returned
    And the last group has visibility "public"

  @service @spec-group-visibility
  Scenario: Owner changes group visibility to public
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Team Alpha" with visibility "private"
    When I update the last group visibility to "public"
    Then no error is returned
    And the last group has visibility "public"

  @service @spec-group-visibility
  Scenario: Owner changes group visibility to private
    Given I am authenticated as a deployer named "Bob"
    And I create a group named "Team Beta" with visibility "public"
    When I update the last group visibility to "private"
    Then no error is returned
    And the last group has visibility "private"

  @service @spec-group-visibility
  Scenario: User sees public groups from all users
    Given I am authenticated as a deployer named "Alice"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a public group named "Bob's Public Group"
    When I list all groups
    Then no error is returned
    And the group list includes "Bob's Public Group"

  @service @spec-group-visibility
  Scenario: User sees only their own private groups
    Given I am authenticated as a deployer named "Alice"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a private group named "Bob's Private Group"
    When I list all groups
    Then no error is returned
    And the group list does not include "Bob's Private Group"

  @service @spec-group-visibility
  Scenario: Combined group list shows public and own private
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Alice Private" with visibility "private"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a public group named "Bob Public"
    When I list all groups
    Then no error is returned
    And the group list includes "Alice Private"
    And the group list includes "Bob Public"

  @service @spec-group-visibility
  Scenario: Non-owner cannot see private group
    Given I am authenticated as a deployer named "Alice"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a private group named "Bob's Secret"
    When I retrieve the group "Bob's Secret" by name
    Then an error is returned
    And the error message contains "not found"

  @http @spec-group-visibility @auth
  Scenario: Non-owner cannot change visibility
    Given I am authenticated as a deployer named "Alice"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a public group named "Bob's Group"
    When I am authenticated as a deployer named "Alice"
    And I attempt to update group "Bob's Group" visibility to "private"
    Then an error is returned

  @service @spec-group-visibility
  Scenario: Owner can always change visibility
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group" with visibility "private"
    When I update the last group visibility to "public"
    Then no error is returned
    When I update the last group visibility to "private"
    Then no error is returned

  @service @spec-group-visibility
  Scenario: User can assign schedule to public group
    Given I am authenticated as a deployer named "Alice"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a public group named "Public Deploys"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I assign the last schedule to group "Public Deploys"
    Then no error is returned

  @service @spec-group-visibility
  Scenario: User can assign schedule to own private group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Private Group" with visibility "private"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I assign the last schedule to group "My Private Group"
    Then no error is returned

  @service @spec-group-visibility
  Scenario: User cannot assign schedule to other's private group
    Given I am authenticated as a deployer named "Alice"
    And a user "Bob" with role "deployer" exists
    And user "Bob" creates a private group named "Bob's Private Group"
    And I create a schedule with:
      | service_name | api-service          |
      | scheduled_at | 2026-06-15T14:00:00Z |
      | environments | production           |
      | owners       | alice@example.com    |
    When I assign the last schedule to group "Bob's Private Group"
    Then an error is returned

  @ui @spec-group-visibility @wip
  Scenario: Sidebar shows visibility icons
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Public Team" with visibility "public"
    And I create a group named "Private Team" with visibility "private"
    When I navigate to the web interface
    Then the sidebar shows group "Public Team" with globe icon
    And the sidebar shows group "Private Team" with lock icon
