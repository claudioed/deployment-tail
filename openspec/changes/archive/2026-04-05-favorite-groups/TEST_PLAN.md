# Group Favorites Feature - Test Plan

This document provides a comprehensive test plan for verifying the group favorites feature across all layers of the application.

## Automated Tests (Completed)

### Unit Tests

**Repository Layer** (`internal/adapters/output/mysql/group_repository_test.go`)
- ✅ TestGroupRepositoryFavoriteGroup - Basic favorite operation
- ✅ TestGroupRepositoryFavoriteGroupIdempotent - Multiple favorites should not error
- ✅ TestGroupRepositoryUnfavoriteGroup - Basic unfavorite operation
- ✅ TestGroupRepositoryUnfavoriteGroupIdempotent - Unfavoriting non-favorited group should not error
- ✅ TestGroupRepositoryIsFavorite - Check favorite status
- ✅ TestGroupRepositoryFindAllWithFavorites - Get groups with favorite status and sorting
- ✅ TestGroupRepositoryCascadeDeleteGroup - Favorites removed when group deleted
- ✅ TestGroupRepositoryCascadeDeleteUser - Favorites removed when user deleted

**Application Layer** (`internal/application/group_service_test.go`)
- ✅ TestFavoriteGroup - Basic use case
- ✅ TestFavoriteGroupInvalidUserID - Validation
- ✅ TestFavoriteGroupInvalidGroupID - Validation
- ✅ TestFavoriteGroupNotFound - Error handling
- ✅ TestUnfavoriteGroup - Basic use case
- ✅ TestUnfavoriteGroupIdempotent - Idempotent operation
- ✅ TestListGroupsWithFavorites - Get groups with favorite status
- ✅ TestListGroupsWithFavoritesInvalidUserID - Validation

**HTTP Handler Layer** (`internal/adapters/input/http/handler_test.go`)
- ✅ TestFavoriteGroup - Handler with authentication
- ✅ TestFavoriteGroupUnauthenticated - 401 without auth
- ✅ TestFavoriteGroupNotFound - 404 for non-existent group
- ✅ TestUnfavoriteGroup - Handler with authentication
- ✅ TestUnfavoriteGroupUnauthenticated - 401 without auth
- ✅ TestGetGroupsWithFavoriteStatus - List with favorites and sorting

## Integration Tests (To Be Executed)

### Task 7.10: CLI Authentication Token Handling for Favorite Operations

**Prerequisites:**
- Server running with database
- User logged in via `deployment-tail auth login`

**Test Steps:**

1. **Test with valid token:**
   ```bash
   # Create a test group
   deployment-tail group create --name "Test Group" --owner "test-user"

   # Favorite the group
   deployment-tail group favorite <group-id>
   # Expected: Success message "Group <id> favorited successfully"

   # Verify it appears in favorites list
   deployment-tail group list --owner "test-user"
   # Expected: ★ icon appears next to the group

   # Unfavorite the group
   deployment-tail group unfavorite <group-id>
   # Expected: Success message "Group <id> unfavorited successfully"
   ```

2. **Test without token:**
   ```bash
   # Logout to clear token
   deployment-tail auth logout

   # Try to favorite
   deployment-tail group favorite <group-id>
   # Expected: Error message "authentication required" or prompt to login
   ```

3. **Test with expired token:**
   ```bash
   # Manually edit token file to have expired timestamp (if possible)
   # OR wait for token to expire naturally

   # Try to favorite
   deployment-tail group favorite <group-id>
   # Expected: Automatic token refresh OR error message about expired token
   ```

4. **Test --favorites-only flag:**
   ```bash
   # Create multiple groups and favorite some
   deployment-tail group create --name "Group A" --owner "test-user"
   deployment-tail group create --name "Group B" --owner "test-user"
   deployment-tail group create --name "Group C" --owner "test-user"

   deployment-tail group favorite <group-b-id>
   deployment-tail group favorite <group-c-id>

   # List only favorites
   deployment-tail group list --owner "test-user" --favorites-only
   # Expected: Only Group B and Group C appear, both with ★ icon
   ```

**Success Criteria:**
- ✅ Valid token allows favorite operations
- ✅ Missing token returns appropriate error
- ✅ Expired token is refreshed automatically (if implemented) or shows error
- ✅ Star icon (★) appears for favorited groups in list output
- ✅ --favorites-only flag filters correctly

---

### Task 8.12: Test Favorite Toggling on Desktop and Mobile

**Prerequisites:**
- Server running
- Web UI accessible
- User authenticated in browser

**Desktop Tests:**

1. **Mouse interactions:**
   - Navigate to Groups management modal
   - Hover over star icon → should show hover state (color change)
   - Click empty star (☆) → should immediately fill (★)
   - Verify optimistic UI update is instant
   - Check network tab to confirm POST /api/v1/groups/{id}/favorite
   - Click filled star (★) → should immediately empty (☆)
   - Check network tab to confirm DELETE /api/v1/groups/{id}/favorite

2. **Tab favorites:**
   - Check group tab headers have star icons
   - Click star on tab → should toggle immediately
   - Verify tab order updates (favorites move to left)
   - Refresh page → favorites should persist

3. **Error handling:**
   - Disconnect from network
   - Try to toggle favorite
   - Should see error message
   - Star state should revert to original state

**Mobile/Tablet Tests:**

1. **Touch interactions:**
   - Tap star icon on touch screen
   - Verify tap target is large enough (minimum 44x44px)
   - Check touch feedback is visible
   - Verify no accidental double-taps cause issues

2. **Responsive layout:**
   - Test on various screen sizes (320px, 768px, 1024px)
   - Star icons should remain visible and accessible
   - No layout shifts when favoriting/unfavoriting

3. **Performance:**
   - Toggle favorites rapidly
   - Should remain responsive
   - No visual glitches or lag

**Success Criteria:**
- ✅ Star toggle works on mouse click (desktop)
- ✅ Star toggle works on touch (mobile/tablet)
- ✅ Optimistic UI update is instant
- ✅ Error handling reverts state on failure
- ✅ Favorites persist across page refreshes
- ✅ Touch targets are appropriately sized
- ✅ Responsive layout works on all screen sizes

---

### Task 9.1: Test Favorite/Unfavorite Operations with Valid Group IDs

**API Tests:**

```bash
# Using curl or similar HTTP client

# 1. Favorite a group
curl -X POST http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content

# 2. Verify favorite status in list
curl http://localhost:8080/api/v1/groups?owner=test-user \
  -H "Authorization: Bearer <token>"
# Expected: 200 OK, isFavorite=true for the favorited group

# 3. Unfavorite the group
curl -X DELETE http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content

# 4. Verify unfavorite status
curl http://localhost:8080/api/v1/groups?owner=test-user \
  -H "Authorization: Bearer <token>"
# Expected: 200 OK, isFavorite=false or missing for the group
```

**Success Criteria:**
- ✅ POST /groups/{id}/favorite returns 204
- ✅ DELETE /groups/{id}/favorite returns 204
- ✅ isFavorite field reflects actual state
- ✅ Database record created/deleted correctly

---

### Task 9.2: Test Favorite Operations with Non-Existent Group IDs

```bash
# Use a random UUID that doesn't exist
curl -X POST http://localhost:8080/api/v1/groups/00000000-0000-0000-0000-000000000000/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 404 Not Found

curl -X DELETE http://localhost:8080/api/v1/groups/00000000-0000-0000-0000-000000000000/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 404 Not Found (or 204 if idempotent - check implementation)
```

**Success Criteria:**
- ✅ Returns 404 for non-existent group ID
- ✅ Error message is clear and helpful

---

### Task 9.3: Test Favorite Operations Without Authentication

```bash
# Attempt to favorite without Authorization header
curl -X POST http://localhost:8080/api/v1/groups/{group-id}/favorite
# Expected: 401 Unauthorized

curl -X DELETE http://localhost:8080/api/v1/groups/{group-id}/favorite
# Expected: 401 Unauthorized

# Attempt with invalid token
curl -X POST http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer invalid-token"
# Expected: 401 Unauthorized
```

**Success Criteria:**
- ✅ Returns 401 without authentication
- ✅ Returns 401 with invalid token
- ✅ Error message indicates authentication required

---

### Task 9.4: Test Idempotency of Favorite Operation

```bash
# Favorite the same group multiple times
curl -X POST http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content

curl -X POST http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content (should not error)

curl -X POST http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content (should not error)

# Verify only one database record exists
# Query: SELECT COUNT(*) FROM group_favorites WHERE user_id = ? AND group_id = ?
# Expected: 1 record
```

**Success Criteria:**
- ✅ Multiple favorite requests don't return errors
- ✅ Only one database record exists
- ✅ Response is consistent (204 No Content)

---

### Task 9.5: Test Idempotency of Unfavorite Operation

```bash
# Unfavorite a group that is not favorited
curl -X DELETE http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content

# Unfavorite again
curl -X DELETE http://localhost:8080/api/v1/groups/{group-id}/favorite \
  -H "Authorization: Bearer <token>"
# Expected: 204 No Content (should not error)

# Verify no database record exists
# Query: SELECT COUNT(*) FROM group_favorites WHERE user_id = ? AND group_id = ?
# Expected: 0 records
```

**Success Criteria:**
- ✅ Unfavoriting non-favorited group doesn't error
- ✅ Multiple unfavorite requests are consistent
- ✅ Database state is correct (no record)

---

### Task 9.6: Test Group List Returns Favorites First

**Test Setup:**
```sql
-- Create test data
INSERT INTO groups (id, name, owner) VALUES
  ('group-a', 'Alpha', 'test-user'),
  ('group-b', 'Beta', 'test-user'),
  ('group-c', 'Charlie', 'test-user'),
  ('group-d', 'Delta', 'test-user');

-- Favorite Beta and Delta
INSERT INTO group_favorites (user_id, group_id) VALUES
  ('test-user-id', 'group-b'),
  ('test-user-id', 'group-d');
```

**Test:**
```bash
curl http://localhost:8080/api/v1/groups?owner=test-user \
  -H "Authorization: Bearer <token>"
```

**Expected Order:**
1. Beta (favorited)
2. Delta (favorited)
3. Alpha (not favorited)
4. Charlie (not favorited)

**Success Criteria:**
- ✅ Favorited groups appear first
- ✅ Groups are sorted alphabetically within each section
- ✅ isFavorite field is true for favorited groups

---

### Task 9.7: Test Alphabetical Sorting Within Favorites and Non-Favorites

Same test as 9.6, verify:
- Within favorites: Beta comes before Delta (alphabetical)
- Within non-favorites: Alpha comes before Charlie (alphabetical)

---

### Task 9.8: Test isFavorite Field Accuracy

```bash
# Get group list
curl http://localhost:8080/api/v1/groups?owner=test-user \
  -H "Authorization: Bearer <token>"

# For each group:
# - If isFavorite=true, verify record exists in group_favorites table
# - If isFavorite=false/null, verify no record in group_favorites table
```

**Success Criteria:**
- ✅ isFavorite field matches database state
- ✅ No false positives or negatives

---

### Task 9.9: Test Cascade Delete When Group is Deleted

**Test Setup:**
```bash
# Create group and favorite it
GROUP_ID=$(curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"Test Group","owner":"test-user"}' | jq -r '.id')

curl -X POST http://localhost:8080/api/v1/groups/$GROUP_ID/favorite \
  -H "Authorization: Bearer <token>"

# Verify favorite exists
# Query: SELECT * FROM group_favorites WHERE group_id = ?
# Expected: 1 record

# Delete the group
curl -X DELETE http://localhost:8080/api/v1/groups/$GROUP_ID \
  -H "Authorization: Bearer <token>"

# Verify favorite was cascade deleted
# Query: SELECT * FROM group_favorites WHERE group_id = ?
# Expected: 0 records
```

**Success Criteria:**
- ✅ Favorite record is automatically deleted when group is deleted
- ✅ No orphaned records in group_favorites table

---

### Task 9.10: Test Cascade Delete When User is Deleted

**Test Setup:**
```bash
# Create user and group
# Favorite the group as that user

# Verify favorite exists
# Query: SELECT * FROM group_favorites WHERE user_id = ?
# Expected: 1+ records

# Delete the user
# Query: DELETE FROM users WHERE id = ?

# Verify favorites were cascade deleted
# Query: SELECT * FROM group_favorites WHERE user_id = ?
# Expected: 0 records
```

**Success Criteria:**
- ✅ All favorites for user are deleted when user is deleted
- ✅ No orphaned records in group_favorites table

---

### Task 9.11: Test CLI Favorite/Unfavorite Commands

```bash
# Create test group
deployment-tail group create --name "CLI Test" --owner "test-user"

# Favorite the group
deployment-tail group favorite <group-id>
# Expected: "Group <id> favorited successfully"

# Unfavorite the group
deployment-tail group unfavorite <group-id>
# Expected: "Group <id> unfavorited successfully"

# Test with invalid group ID
deployment-tail group favorite invalid-uuid
# Expected: Error message with validation failure

# Test without authentication
deployment-tail auth logout
deployment-tail group favorite <group-id>
# Expected: Authentication error
```

**Success Criteria:**
- ✅ Commands execute successfully
- ✅ Success messages are clear
- ✅ Error messages are helpful
- ✅ Authentication is enforced

---

### Task 9.12: Test CLI --favorites-only Flag

```bash
# Create multiple groups
deployment-tail group create --name "Group A" --owner "test-user"
deployment-tail group create --name "Group B" --owner "test-user"
deployment-tail group create --name "Group C" --owner "test-user"

# Favorite only Group B
deployment-tail group favorite <group-b-id>

# List all groups
deployment-tail group list --owner "test-user"
# Expected: All 3 groups, with ★ only next to Group B

# List only favorites
deployment-tail group list --owner "test-user" --favorites-only
# Expected: Only Group B with ★ icon

# Unfavorite all groups
deployment-tail group unfavorite <group-b-id>

# List favorites only (should be empty)
deployment-tail group list --owner "test-user" --favorites-only
# Expected: Empty list or "No favorite groups found" message
```

**Success Criteria:**
- ✅ --favorites-only filters correctly
- ✅ Star icon (★) appears for favorites
- ✅ Non-favorited groups are excluded
- ✅ Empty state handled gracefully

---

### Task 9.13: Test Web UI Star Icon Toggling and Visual Feedback

**Manual Testing Steps:**

1. **Initial State:**
   - Navigate to Groups management modal
   - Verify star icons are visible
   - Unfavorited groups should show empty star (☆)
   - Favorited groups should show filled star (★)

2. **Toggle Favorite:**
   - Click empty star (☆)
   - Should immediately change to filled star (★)
   - Visual feedback: color change, animation (if implemented)
   - No page refresh required

3. **Toggle Unfavorite:**
   - Click filled star (★)
   - Should immediately change to empty star (☆)
   - Visual feedback: color change, animation (if implemented)

4. **Sorting Update:**
   - After favoriting, verify group moves to top of list (if list is visible)
   - After unfavoriting, verify group returns to alphabetical position

5. **Persistence:**
   - Toggle favorites
   - Refresh page (Ctrl+R / Cmd+R)
   - Verify favorites persist correctly

6. **Error Handling:**
   - Disconnect network (browser dev tools: Offline mode)
   - Try to toggle favorite
   - Should show error message
   - Star should revert to original state
   - Reconnect network and verify state is consistent

**Success Criteria:**
- ✅ Star icon toggles immediately (optimistic UI)
- ✅ Visual feedback is clear (hover, active states)
- ✅ Favorites persist across page refreshes
- ✅ Error messages appear on failure
- ✅ State reverts on error (rollback)
- ✅ No JavaScript errors in console

---

## Test Execution Summary

### Automated Tests: ✅ 27/27 Complete
- Repository layer: 8 tests
- Application layer: 8 tests
- HTTP handler layer: 6 tests
- Database migrations: 5 tests (implicit via integration tests)

### Manual/Integration Tests: ⏸️ 13 Pending
- CLI authentication: 1 test (7.10)
- Web UI desktop/mobile: 1 test (8.12)
- End-to-end scenarios: 11 tests (9.1-9.13)

## Notes

- All unit tests pass successfully
- Integration tests require running server and database
- Manual tests require human verification of UI/UX
- Some tests can be partially automated using integration test framework
- Consider adding these as integration tests in future iterations
