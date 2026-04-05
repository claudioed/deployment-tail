## Context

The deployment-tail system uses groups to organize deployment schedules. Users can create groups (projects, teams, releases) and assign schedules to them. Groups appear as tabs in the Web UI for easy filtering. However, as the number of groups grows, users must scroll through all groups to find the ones they use most frequently.

The system follows hexagonal architecture with DDD principles. The domain layer contains aggregates (Schedule, Group, User), the application layer contains use cases, and adapters handle HTTP/CLI/MySQL. The Group aggregate currently tracks name, description, owner, and timestamps, with a many-to-many relationship to schedules.

Current group listing returns groups ordered by creation date. There's no concept of user-specific group preferences or favorites.

## Goals / Non-Goals

**Goals:**
- Allow authenticated users to mark groups as favorites
- Persist favorite status per-user in the database
- Return favorited groups first in API responses and UI displays
- Provide intuitive UI for toggling favorites (star icons)
- Support favorite management via CLI
- Maintain performance with minimal query overhead

**Non-Goals:**
- Sharing favorites between users
- Favorite synchronization across devices (JWT-based auth handles this naturally)
- Favorite analytics or recommendations
- Favoriting schedules (only groups in this change)
- Limit on number of favorites per user
- Favorite ordering/prioritization beyond binary favorite/not-favorite

## Decisions

### 1. Domain Model: Separate Aggregate vs Value Object in Group

**Decision**: Create a separate `GroupFavorite` entity managed by `GroupRepository` (or new `GroupFavoriteRepository`), not a value object on Group.

**Rationale**:
- Favorites are a user-specific relationship, not intrinsic to the Group aggregate
- Group aggregate root represents a group itself; favorites are a separate concern (user preferences)
- Separates write model (Group) from read model concerns (user favorites)
- Easier to query "all favorites for a user" without loading all Group aggregates
- Follows SRP: Group handles group data, favorites handle user-group relationships

**Alternatives considered**:
- Storing favorites as a collection on User aggregate: Would require loading User to check if group is favorited, less efficient for group listing queries
- Adding `favoritedBy` collection to Group: Violates SRP, makes Group responsible for user preferences, and bloats Group with per-user data

### 2. Repository Pattern: Extend GroupRepository vs New Repository

**Decision**: Extend `GroupRepository` with favorite-related methods.

**Rationale**:
- Favorites are queried in the context of groups (e.g., "list groups, mark favorites")
- Keeps related group operations together
- Simpler dependency injection (one repository for group operations)
- Repository can optimize joined queries (groups + favorite status in single query)

**Methods to add**:
```go
FavoriteGroup(ctx context.Context, userID user.UserID, groupID group.GroupID) error
UnfavoriteGroup(ctx context.Context, userID user.UserID, groupID group.GroupID) error
IsFavorite(ctx context.Context, userID user.UserID, groupID group.GroupID) (bool, error)
FindByOwnerWithFavorites(ctx context.Context, userID user.UserID, owner string) ([]*group.Group, map[group.GroupID]bool, error)
```

**Alternatives considered**:
- Creating `GroupFavoriteRepository`: More separation of concerns, but adds complexity for querying groups with favorite status (requires two repos)

### 3. Database Schema: Junction Table with Timestamps

**Decision**: Create `group_favorites` table with `(user_id, group_id)` composite primary key and `created_at` timestamp.

**Schema**:
```sql
CREATE TABLE group_favorites (
    user_id CHAR(36) NOT NULL,
    group_id CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, group_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);
CREATE INDEX idx_group_favorites_user ON group_favorites(user_id);
```

**Rationale**:
- Composite primary key prevents duplicate favorites (idempotent operations)
- Cascading deletes automatically clean up when user or group is deleted
- `created_at` timestamp useful for future features (e.g., recently favorited)
- Index on `user_id` optimizes common query: "get all favorites for user"
- No separate `id` column needed (composite key sufficient)

**Alternatives considered**:
- Adding `favorites` JSON column to `users` table: Less normalized, harder to query, no referential integrity
- Boolean `is_favorite` per user on `groups`: Doesn't scale, requires schema change per user

### 4. API Response: Inline `isFavorite` Field vs Separate Endpoint

**Decision**: Add `isFavorite` boolean field to Group response objects when user is authenticated.

**Rationale**:
- Client gets all needed data in one request (group + favorite status)
- No need for separate "get favorites" endpoint followed by "get groups"
- Simpler client-side logic (no need to merge two responses)
- Consistent with REST principles (resource representation includes user-specific state)

**Example response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Backend Services",
  "description": "All backend microservices",
  "owner": "ops-team",
  "isFavorite": true,
  "createdAt": "2026-03-30T21:00:00Z",
  "updatedAt": "2026-03-30T21:00:00Z"
}
```

**Alternatives considered**:
- Separate `/api/v1/users/me/favorites` endpoint: Requires two API calls, more complex client logic
- Using HTTP headers to indicate favorites: Non-standard, harder to parse

### 5. Sorting Strategy: Always Favorites-First vs Optional

**Decision**: Always return favorited groups first in `GET /api/v1/groups`, sorted alphabetically within each section (favorites, then non-favorites).

**Rationale**:
- Predictable behavior: users always see their favorites at the top
- No query parameter needed (simpler API)
- Matches user mental model: favorites are "pinned" to the top
- Web UI tabs will show favorites first automatically

**Sorting order**:
1. Favorited groups (alphabetically by name)
2. Non-favorited groups (alphabetically by name)

**Alternatives considered**:
- Optional `?favorites=first` query parameter: Adds complexity, most users want this behavior by default
- Separate `?favorites=true` filter: Requires two separate calls for "show all with favorites first"

### 6. UI Interaction: Click Star Icon vs Dedicated Button

**Decision**: Use clickable star icon (★ filled for favorited, ☆ outline for not favorited) on group cards and tabs.

**Rationale**:
- Universal pattern: star icons widely recognized for favorites
- Inline action: no need to open modal or navigate to settings
- Immediate visual feedback: icon changes on click
- Space-efficient: fits naturally on group cards and tabs

**Implementation**:
- Star icon appears on group cards in list view
- Star icon appears on each group tab in Web UI
- Click toggles favorite status (POST or DELETE to API)
- Optimistic UI update (update immediately, revert on error)

**Alternatives considered**:
- Dedicated "Favorite" button: Takes more space, less intuitive
- Right-click context menu: Not mobile-friendly, less discoverable

### 7. CLI Support: Subcommands vs Flags

**Decision**: Add `deployment-tail group favorite <group-id>` and `deployment-tail group unfavorite <group-id>` subcommands, plus `--favorites-only` flag for listing.

**Rationale**:
- Explicit commands are clearer for mutation operations
- Flag for filtering is consistent with existing CLI patterns (e.g., `--owner`)
- Follows existing CLI structure (`deployment-tail group ...`)

**Commands**:
```bash
deployment-tail group list --favorites-only      # List only favorited groups
deployment-tail group favorite <group-id>        # Mark group as favorite
deployment-tail group unfavorite <group-id>      # Remove favorite
```

**Alternatives considered**:
- `--favorite` flag on list command: Ambiguous (filter or action?)
- Separate `favorites` top-level command: Inconsistent with group operations

## Risks / Trade-offs

**[Risk]** Group deletion while favorited → **Mitigation**: Use `ON DELETE CASCADE` foreign key constraint to auto-cleanup favorites when group is deleted.

**[Risk]** User deletion with existing favorites → **Mitigation**: Use `ON DELETE CASCADE` foreign key constraint to auto-cleanup favorites when user is deleted.

**[Risk]** Performance degradation with many groups/favorites → **Mitigation**: Add index on `group_favorites(user_id)`, use LEFT JOIN for efficient single-query retrieval, limit query to current owner's groups.

**[Risk]** Optimistic UI update fails (network error, auth expired) → **Mitigation**: Revert UI state on error, show error notification, prompt re-authentication if JWT expired.

**[Trade-off]** Always sorting favorites-first vs optional sorting: Chose always-first for simplicity and predictability. Users who don't want this can ignore favorites feature entirely.

**[Trade-off]** Separate repository vs extending GroupRepository: Chose extending for query efficiency, but creates slight coupling between group and favorite concerns. Acceptable because they're tightly related in the UI.

**[Trade-off]** Inline `isFavorite` field increases response size: Minimal cost (1 boolean per group), significant UX benefit (no extra API call).
