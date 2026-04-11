Feature: Group Favorites

  Background:
    Given a clean schedule repository

  @service @spec-group-favorites @smoke
  Scenario: User favorites a group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Production Deploys"
    When I favorite the last group
    Then no error is returned

  @service @spec-group-favorites
  Scenario: Favorite is user-specific
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Shared Group" with visibility "public"
    And I favorite the last group
    When I am authenticated as a deployer named "Bob"
    And I retrieve the group "Shared Group" by name
    Then the group is not favorited

  @service @spec-group-favorites
  Scenario: Favoriting is idempotent
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    When I favorite the last group
    And I favorite the last group
    And I favorite the last group
    Then no error is returned

  @service @spec-group-favorites
  Scenario: Cannot favorite non-existent group
    Given I am authenticated as a deployer named "Alice"
    When I favorite a group with invalid ID
    Then an error is returned
    And the error message contains "not found"

  @http @spec-group-favorites @auth
  Scenario: Unauthenticated user cannot favorite
    Given an unauthenticated client
    When I POST "/groups/12345678-1234-1234-1234-123456789abc/favorite" with body:
      """
      {}
      """
    Then the HTTP status is 401

  @service @spec-group-favorites
  Scenario: User unfavorites a group
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Temp Group"
    And I favorite the last group
    When I unfavorite the last group
    Then no error is returned

  @service @spec-group-favorites
  Scenario: Unfavoriting is idempotent
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "My Group"
    When I unfavorite the last group
    And I unfavorite the last group
    Then no error is returned

  @service @spec-group-favorites
  Scenario: Cannot unfavorite non-existent group
    Given I am authenticated as a deployer named "Alice"
    When I unfavorite a group with invalid ID
    Then an error is returned
    And the error message contains "not found"

  @http @spec-group-favorites @auth
  Scenario: Unauthenticated user cannot unfavorite
    Given an unauthenticated client
    When I DELETE "/groups/12345678-1234-1234-1234-123456789abc/favorite"
    Then the HTTP status is 401

  @service @spec-group-favorites @smoke
  Scenario: Favorited group shows isFavorite true
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Favorite Group"
    And I favorite the last group
    When I retrieve the last group by ID
    Then no error is returned
    And the last group is favorited

  @service @spec-group-favorites
  Scenario: Non-favorited group shows isFavorite false
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Normal Group"
    When I retrieve the last group by ID
    Then no error is returned
    And the last group is not favorited

  @service @spec-group-favorites
  Scenario: isFavorite field in group list
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Fav1"
    And I favorite the last group
    And I create a group named "Fav2"
    And I favorite the last group
    And I create a group named "NotFav"
    When I list all groups
    Then the group "Fav1" is favorited
    And the group "Fav2" is favorited
    And the group "NotFav" is not favorited

  @service @spec-group-favorites @smoke
  Scenario: Favorites appear first in list
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Zebra Group"
    And I create a group named "Apple Group"
    And I favorite the last group
    And I create a group named "Banana Group"
    When I list all groups
    Then the first group in the list is "Apple Group"

  @service @spec-group-favorites
  Scenario: Favorites sorted alphabetically
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Zebra Fav"
    And I favorite the last group
    And I create a group named "Apple Fav"
    And I favorite the last group
    And I create a group named "Banana Fav"
    And I favorite the last group
    When I list all groups
    Then the group list order is "Apple Fav, Banana Fav, Zebra Fav"

  @service @spec-group-favorites
  Scenario: Non-favorites sorted alphabetically
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Zebra"
    And I create a group named "Apple"
    And I create a group named "Banana"
    When I list all groups
    Then the group list order is "Apple, Banana, Zebra"

  @service @spec-group-favorites
  Scenario: User with no favorites sees alphabetical list
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Zebra"
    And I create a group named "Apple"
    When I list all groups
    Then the group list order is "Apple, Zebra"

  @service @spec-group-favorites
  Scenario: Group deletion removes favorites
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "ToDelete"
    And I favorite the last group
    When I delete the last group
    Then no error is returned
    When I list all groups
    Then the group list does not include "ToDelete"

  @cli @spec-group-favorites @wip
  Scenario: Favorite group via CLI
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "CLI Group"
    When I run CLI command "group favorite" with the last group ID
    Then the CLI output contains "favorited"

  @cli @spec-group-favorites @wip
  Scenario: Unfavorite group via CLI
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "CLI Group"
    And I favorite the last group
    When I run CLI command "group unfavorite" with the last group ID
    Then the CLI output contains "unfavorited"

  @cli @spec-group-favorites @wip
  Scenario: List only favorited groups via CLI
    Given I am authenticated as a deployer named "Alice"
    And I create a group named "Fav1"
    And I favorite the last group
    And I create a group named "NotFav"
    When I run CLI command "group list --favorites-only"
    Then the CLI output contains "Fav1"
    And the CLI output does not contain "NotFav"
