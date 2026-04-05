## 1. Database Migration

- [x] 1.1 Create up migration to add `group_favorites` table with user_id, group_id, created_at columns
- [x] 1.2 Add composite primary key on (user_id, group_id) to prevent duplicates
- [x] 1.3 Add foreign key constraints with ON DELETE CASCADE for user_id and group_id
- [x] 1.4 Add index on user_id for efficient favorite lookups
- [x] 1.5 Create down migration to drop `group_favorites` table

## 2. Domain Layer

- [x] 2.1 Review GroupRepository interface and determine if extending it vs new repository
- [x] 2.2 Add FavoriteGroup method to repository interface
- [x] 2.3 Add UnfavoriteGroup method to repository interface
- [x] 2.4 Add IsFavorite method to repository interface
- [x] 2.5 Add FindByOwnerWithFavorites method to repository interface that returns groups with favorite status
- [x] 2.6 Add unit tests for repository interface contract

## 3. MySQL Repository Implementation

- [x] 3.1 Implement FavoriteGroup in mysql.GroupRepository (INSERT INTO group_favorites)
- [x] 3.2 Implement UnfavoriteGroup in mysql.GroupRepository (DELETE FROM group_favorites)
- [x] 3.3 Implement IsFavorite in mysql.GroupRepository (SELECT EXISTS query)
- [x] 3.4 Implement FindByOwnerWithFavorites with LEFT JOIN to group_favorites
- [x] 3.5 Update FindByOwnerWithFavorites to sort favorites first, then alphabetically within each section
- [x] 3.6 Add integration tests for favorite operations
- [x] 3.7 Test cascade delete behavior when group is deleted
- [x] 3.8 Test cascade delete behavior when user is deleted

## 4. Application Layer Use Cases

- [x] 4.1 Create FavoriteGroup use case in application/group_service.go
- [x] 4.2 Add authorization check (must be authenticated)
- [x] 4.3 Add validation to check group exists before favoriting
- [x] 4.4 Create UnfavoriteGroup use case in application/group_service.go
- [x] 4.5 Add authorization check (must be authenticated)
- [x] 4.6 Update ListGroups use case to accept optional userID parameter
- [x] 4.7 Modify ListGroups to call FindByOwnerWithFavorites when userID is provided
- [x] 4.8 Add unit tests for FavoriteGroup use case
- [x] 4.9 Add unit tests for UnfavoriteGroup use case
- [x] 4.10 Add unit tests for ListGroups with favorites

## 5. OpenAPI Specification

- [x] 5.1 Add POST /api/v1/groups/{id}/favorite endpoint to openapi.yaml
- [x] 5.2 ADD DELETE /api/v1/groups/{id}/favorite endpoint to openapi.yaml
- [x] 5.3 Add isFavorite boolean field to Group schema in openapi.yaml
- [x] 5.4 Update GET /api/v1/groups response to include isFavorite field
- [x] 5.5 Run `make generate` to regenerate API stubs from openapi.yaml

## 6. HTTP Handlers

- [x] 6.1 Implement FavoriteGroup handler in adapters/input/http/handler.go
- [x] 6.2 Extract authenticated user from JWT context in FavoriteGroup handler
- [x] 6.3 Call FavoriteGroup use case and return 204 No Content on success
- [x] 6.4 Implement UnfavoriteGroup handler in adapters/input/http/handler.go
- [x] 6.5 Extract authenticated user from JWT context in UnfavoriteGroup handler
- [x] 6.6 Call UnfavoriteGroup use case and return 204 No Content on success
- [x] 6.7 Update GetGroups handler to extract authenticated user from JWT context
- [x] 6.8 Pass userID to ListGroups use case to enable favorite-aware listing
- [x] 6.9 Add isFavorite field to group response objects in GetGroups handler
- [x] 6.10 Add unit tests for FavoriteGroup handler
- [x] 6.11 Add unit tests for UnfavoriteGroup handler
- [x] 6.12 Add unit tests for GetGroups handler with favorite status

## 7. CLI Commands

- [x] 7.1 Add `favorite` subcommand to `deployment-tail group` in adapters/input/cli/group.go
- [x] 7.2 Implement favorite command to POST to /api/v1/groups/{id}/favorite
- [x] 7.3 Add success confirmation message for favorite command
- [x] 7.4 Add `unfavorite` subcommand to `deployment-tail group`
- [x] 7.5 Implement unfavorite command to DELETE from /api/v1/groups/{id}/favorite
- [x] 7.6 Add success confirmation message for unfavorite command
- [x] 7.7 Add `--favorites-only` flag to `group list` command
- [x] 7.8 Implement favorites-only filter to show only groups with isFavorite=true
- [x] 7.9 Add star icon (★) indicator to group list output for favorited groups
- [x] 7.10 Test CLI authentication token handling for favorite operations

## 8. Web UI Implementation

- [x] 8.1 Add star icon (☆/★) to group cards in web/index.html
- [x] 8.2 Add star icon to group tabs for favoriting
- [x] 8.3 Implement toggleFavorite JavaScript function in web/app.js
- [x] 8.4 Call POST /api/v1/groups/{id}/favorite when star is clicked (unfavorited → favorited)
- [x] 8.5 Call DELETE /api/v1/groups/{id}/favorite when star is clicked (favorited → unfavorited)
- [x] 8.6 Implement optimistic UI update (change star icon immediately)
- [x] 8.7 Add error handling to revert UI state if favorite API call fails
- [x] 8.8 Update loadGroups function to respect isFavorite field from API
- [x] 8.9 Update renderGroups to sort favorited groups first
- [x] 8.10 Add CSS styling for favorite star icon (filled vs outline)
- [x] 8.11 Add hover states and click feedback for star icon
- [x] 8.12 Test favorite toggling on desktop and mobile

## 9. Testing

- [x] 9.1 Test favorite/unfavorite operations with valid group IDs
- [x] 9.2 Test favorite operations with non-existent group IDs (should return 404)
- [x] 9.3 Test favorite operations without authentication (should return 401)
- [x] 9.4 Test idempotency of favorite operation (multiple favorites of same group)
- [x] 9.5 Test idempotency of unfavorite operation (unfavoriting non-favorited group)
- [x] 9.6 Test group list returns favorites first for authenticated user
- [x] 9.7 Test alphabetical sorting within favorites and non-favorites sections
- [x] 9.8 Test isFavorite field is accurate in group list and detail responses
- [x] 9.9 Test cascade delete when group is deleted (favorites are removed)
- [x] 9.10 Test cascade delete when user is deleted (favorites are removed)
- [x] 9.11 Test CLI favorite/unfavorite commands
- [x] 9.12 Test CLI --favorites-only flag
- [x] 9.13 Test Web UI star icon toggling and visual feedback

## 10. Documentation

- [x] 10.1 Update README.md with favorite groups feature description
- [x] 10.2 Add favorite API endpoints to README API examples section
- [x] 10.3 Add CLI favorite commands to README
- [x] 10.4 Update Web UI features section to mention favoriting
- [x] 10.5 Document `group_favorites` table schema in README Data Models section
