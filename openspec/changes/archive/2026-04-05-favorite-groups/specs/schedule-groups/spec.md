## MODIFIED Requirements

### Requirement: List schedule groups

The system SHALL allow users to retrieve all schedule groups, with favorited groups returned first for authenticated users.

#### Scenario: List all groups
- **WHEN** user requests list of groups
- **THEN** system returns all groups with id, name, description, owner, created_at, updated_at

#### Scenario: List groups with favorites first
- **WHEN** authenticated user requests list of groups
- **THEN** system returns favorited groups first, followed by non-favorited groups

#### Scenario: List groups ordered by name within sections
- **WHEN** authenticated user requests list of groups
- **THEN** favorited groups are ordered alphabetically by name, and non-favorited groups are ordered alphabetically by name

#### Scenario: Empty groups list
- **WHEN** no groups exist and user requests list
- **THEN** system returns empty array

#### Scenario: Include isFavorite field for authenticated users
- **WHEN** authenticated user requests list of groups
- **THEN** system includes isFavorite boolean field for each group
