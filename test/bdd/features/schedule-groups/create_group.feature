Feature: Schedule Groups — Create Group

  Background:
    Given a clean schedule repository

  @service @spec-schedule-groups @smoke
  Scenario: Create public group
    Given I am authenticated as a deployer named "Alice"
    When I create a group named "Production Deploys" with visibility "public"
    Then no error is returned
    And the last group has name "Production Deploys"
    And the last group has visibility "public"

  @service @spec-schedule-groups
  Scenario: Create private group
    Given I am authenticated as a deployer named "Bob"
    When I create a group named "My Private Group" with visibility "private"
    Then no error is returned
    And the last group has name "My Private Group"
    And the last group has visibility "private"

  @http @spec-schedule-groups @auth
  Scenario: Unauthenticated user cannot create group
    Given an unauthenticated client
    When I POST "/groups" with body:
      """
      {"name":"Test Group","visibility":"public"}
      """
    Then the HTTP status is 401

  @http @spec-schedule-groups @auth
  Scenario: Viewer role cannot create group
    Given I am authenticated as a viewer named "Vera"
    When I POST "/groups" with body:
      """
      {"name":"Test Group","visibility":"public"}
      """
    Then the HTTP status is 400
