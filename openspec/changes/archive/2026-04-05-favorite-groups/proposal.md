## Why

Users currently have to scroll through all groups to find the ones they work with most frequently. For teams managing many groups (projects, releases, teams), this creates friction in daily workflows. Allowing users to favorite specific groups enables quick access to their most important contexts, reducing navigation time and improving productivity.

## What Changes

- Add ability for users to favorite/unfavorite groups
- Favorite status is per-user (not global)
- Favorited groups appear first in the group list and navigation tabs
- Persist favorite preferences in the database
- Add visual indicator (star icon) to show favorite status
- Add API endpoints to manage favorites
- Update Web UI to support favoriting via clickable star icons
- Update CLI to support favorite filtering and management

## Capabilities

### New Capabilities
- `group-favorites`: Allow users to mark groups as favorites, with per-user favorite status persistence and retrieval

### Modified Capabilities
- `schedule-groups`: Extend group listing and display to prioritize favorited groups in UI and API responses

## Impact

**Domain Layer:**
- New `GroupFavorite` aggregate or value object to represent user-group favorite relationships
- New repository interface: `GroupFavoriteRepository` or extend `GroupRepository`

**Application Layer:**
- New use cases: `FavoriteGroup`, `UnfavoriteGroup`, `ListFavoriteGroups`
- Modified use cases: `ListGroups` to optionally return favorites first

**Adapters:**
- HTTP: New endpoints for favoriting operations
- HTTP: Modified group list endpoint to support favorite sorting
- CLI: New commands or flags for favorite management
- MySQL: New table `group_favorites` or junction table for user-group-favorite relationships

**API:**
- New endpoints: `POST /api/v1/groups/{id}/favorite`, `DELETE /api/v1/groups/{id}/favorite`
- Modified endpoints: `GET /api/v1/groups` with optional `?favorites=true` or favorite-first sorting
- Response schema: Add `isFavorite` boolean field to group objects

**Database:**
- New table: `group_favorites` (user_id, group_id, created_at) with composite primary key

**Web UI:**
- Add star icon to group cards/tabs
- Click to toggle favorite status
- Sort favorited groups to the top
- Visual distinction for favorited groups

**CLI:**
- Add `--favorites-only` flag to group list command
- Add `group favorite <id>` and `group unfavorite <id>` commands
