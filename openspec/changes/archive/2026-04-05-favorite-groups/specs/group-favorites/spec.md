## ADDED Requirements

### Requirement: Favorite a group

The system SHALL allow authenticated users to mark a group as a favorite.

#### Scenario: User favorites a group
- **WHEN** authenticated user marks group as favorite
- **THEN** system records favorite relationship between user and group

#### Scenario: Favorite is user-specific
- **WHEN** user A favorites a group
- **THEN** the favorite status is only visible to user A, not other users

#### Scenario: Favoriting is idempotent
- **WHEN** user favorites the same group multiple times
- **THEN** system maintains single favorite record (no duplicates)

#### Scenario: Cannot favorite non-existent group
- **WHEN** user attempts to favorite group with invalid ID
- **THEN** system returns 404 Not Found error

#### Scenario: Unauthenticated user cannot favorite
- **WHEN** unauthenticated request attempts to favorite a group
- **THEN** system returns 401 Unauthorized error

### Requirement: Unfavorite a group

The system SHALL allow authenticated users to remove a group from favorites.

#### Scenario: User unfavorites a group
- **WHEN** authenticated user unfavorites a previously favorited group
- **THEN** system removes favorite relationship

#### Scenario: Unfavoriting is idempotent
- **WHEN** user unfavorites a group that is not favorited
- **THEN** system succeeds without error (no-op)

#### Scenario: Cannot unfavorite non-existent group
- **WHEN** user attempts to unfavorite group with invalid ID
- **THEN** system returns 404 Not Found error

#### Scenario: Unauthenticated user cannot unfavorite
- **WHEN** unauthenticated request attempts to unfavorite a group
- **THEN** system returns 401 Unauthorized error

### Requirement: Include favorite status in group responses

The system SHALL include favorite status in group API responses for authenticated users.

#### Scenario: Favorited group shows isFavorite true
- **WHEN** authenticated user retrieves a group they have favorited
- **THEN** system includes isFavorite=true in response

#### Scenario: Non-favorited group shows isFavorite false
- **WHEN** authenticated user retrieves a group they have not favorited
- **THEN** system includes isFavorite=false in response

#### Scenario: isFavorite field in group list
- **WHEN** authenticated user lists groups
- **THEN** system includes isFavorite field for each group in the list

#### Scenario: Unauthenticated requests exclude isFavorite
- **WHEN** unauthenticated user retrieves groups
- **THEN** system omits isFavorite field or sets to false

### Requirement: List groups with favorites first

The system SHALL return favorited groups before non-favorited groups in list responses.

#### Scenario: Favorites appear first in list
- **WHEN** authenticated user lists groups and has favorited 2 of 5 groups
- **THEN** system returns the 2 favorited groups first, followed by 3 non-favorited groups

#### Scenario: Favorites sorted alphabetically
- **WHEN** authenticated user lists groups with multiple favorites
- **THEN** favorited groups are sorted alphabetically by name within favorites section

#### Scenario: Non-favorites sorted alphabetically
- **WHEN** authenticated user lists groups
- **THEN** non-favorited groups are sorted alphabetically by name within non-favorites section

#### Scenario: User with no favorites sees alphabetical list
- **WHEN** authenticated user lists groups but has not favorited any
- **THEN** system returns all groups alphabetically by name

#### Scenario: Unauthenticated list returns alphabetical order
- **WHEN** unauthenticated user lists groups
- **THEN** system returns groups in standard alphabetical order (no favorite sorting)

### Requirement: Cascade delete favorites

The system SHALL automatically remove favorite relationships when group or user is deleted.

#### Scenario: Group deletion removes favorites
- **WHEN** a group is deleted that has been favorited by users
- **THEN** system removes all favorite relationships for that group

#### Scenario: User deletion removes favorites
- **WHEN** a user is deleted who has favorited groups
- **THEN** system removes all favorite relationships for that user

#### Scenario: Other users' favorites unaffected by user deletion
- **WHEN** user A is deleted and user B has favorited the same group
- **THEN** user B's favorite remains intact

### Requirement: CLI favorite management

The system SHALL provide CLI commands to manage group favorites.

#### Scenario: Favorite group via CLI
- **WHEN** user runs `deployment-tail group favorite <group-id>`
- **THEN** system marks the group as favorite and displays confirmation

#### Scenario: Unfavorite group via CLI
- **WHEN** user runs `deployment-tail group unfavorite <group-id>`
- **THEN** system removes favorite and displays confirmation

#### Scenario: List only favorited groups via CLI
- **WHEN** user runs `deployment-tail group list --favorites-only`
- **THEN** system displays only groups marked as favorite

#### Scenario: List all groups shows favorite indicator
- **WHEN** user runs `deployment-tail group list`
- **THEN** system displays all groups with visual indicator (e.g., ★) for favorited groups

## Notes

- Favorite status is stored in a `group_favorites` junction table with composite primary key `(user_id, group_id)`
- Foreign key constraints with `ON DELETE CASCADE` ensure automatic cleanup
- Favorite operations are authenticated via JWT token
- `isFavorite` field is computed dynamically based on authenticated user context
- Favoriting does not affect group ownership or permissions
- No limit on number of favorites per user
- Favorites are not synchronized (tied to user account via authentication)

## Affected Components

- **Database**: New `group_favorites` table with columns (user_id, group_id, created_at)
- **Domain Layer**: Extend `GroupRepository` interface with favorite methods
- **Application Layer**: New use cases (FavoriteGroup, UnfavoriteGroup, ListGroupsWithFavorites)
- **API Layer**: New endpoints `POST /api/v1/groups/{id}/favorite`, `DELETE /api/v1/groups/{id}/favorite`
- **API Layer**: Modified endpoint `GET /api/v1/groups` includes `isFavorite` field and favorites-first sorting
- **Repository**: Extend MySQL `GroupRepository` implementation with favorite methods
- **CLI**: New subcommands `group favorite <id>`, `group unfavorite <id>`, flag `--favorites-only`

## Rollback Plan

1. Remove favorite API endpoints from OpenAPI spec
2. Remove favorite HTTP handlers
3. Remove favorite use cases from application layer
4. Remove favorite methods from repository interface and implementation
5. Run down migration to drop `group_favorites` table
6. Revert group list endpoint to standard alphabetical sorting
7. Remove `isFavorite` field from API responses
8. Remove CLI favorite commands
