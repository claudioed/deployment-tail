## 1. Web UI - Group Selector Component

- [x] 1.1 Add group multi-select field to Quick Create modal component
- [x] 1.2 Implement API call to fetch user's groups on modal open (GET /groups with favorites-first ordering)
- [x] 1.3 Add loading state for group fetching (spinner or skeleton)
- [x] 1.4 Implement checkbox-based multi-select UI with favorite indicators
- [x] 1.5 Add empty state message when user has no groups
- [x] 1.6 Implement keyboard navigation support for group selector

## 2. Web UI - Quick Create Orchestration

- [x] 2.1 Update Quick Create submit handler to accept selected group IDs
- [x] 2.2 Implement 2-step flow: create schedule → assign to groups
- [x] 2.3 Add error handling for schedule creation failure (show error, keep modal open)
- [x] 2.4 Add error handling for group assignment failure (rollback + delete schedule, show error)
- [x] 2.5 Add error handling for rollback failure (log and show orphaned schedule message)
- [x] 2.6 Update success flow to close modal and refresh schedule list
- [x] 2.7 Preserve form values on error (don't clear inputs on failure)

## 3. CLI - Add --groups Flag

- [x] 3.1 Add `--groups` flag to `schedule quick` command definition (accepts comma-separated string)
- [x] 3.2 Implement group ID/name parser (detect UUIDs vs. names)
- [x] 3.3 Add group name resolution logic (fetch groups, match names to IDs)
- [x] 3.4 Handle group not found error (list available groups)
- [x] 3.5 Handle ambiguous group name error (list matches, suggest using ID)
- [x] 3.6 Update CLI help text to document `--groups` flag usage

## 4. CLI - Quick Create Orchestration

- [x] 4.1 Update CLI quick create to accept group IDs from flag
- [x] 4.2 Implement 2-step flow: create schedule → assign to groups
- [x] 4.3 Add error handling for schedule creation failure (exit with error)
- [x] 4.4 Add error handling for group assignment failure (rollback + delete schedule, exit with error)
- [x] 4.5 Add error handling for rollback failure (log and show orphaned schedule message)
- [x] 4.6 Update success output to show assigned groups

## 5. Testing - Web UI

- [x] 5.1 Test Quick Create with no groups selected (existing behavior)
- [x] 5.2 Test Quick Create with single group selected
- [x] 5.3 Test Quick Create with multiple groups selected
- [x] 5.4 Test group fetch loading state
- [x] 5.5 Test empty groups state (user has no groups)
- [x] 5.6 Test error handling: schedule creation fails
- [x] 5.7 Test error handling: group assignment fails (verify rollback)
- [x] 5.8 Test favorites-first ordering in group selector
- [x] 5.9 Test keyboard navigation through form

## 6. Testing - CLI

- [x] 6.1 Test `schedule quick` without `--groups` flag (existing behavior)
- [x] 6.2 Test `schedule quick --groups "id1,id2"` with valid UUIDs
- [x] 6.3 Test `schedule quick --groups "Project Alpha,Team Backend"` with valid names
- [x] 6.4 Test mixed IDs and names: `--groups "id1,Project Alpha"`
- [x] 6.5 Test group name not found error
- [x] 6.6 Test ambiguous group name error
- [x] 6.7 Test error handling: schedule creation fails
- [x] 6.8 Test error handling: group assignment fails (verify rollback)

## 7. Documentation

- [x] 7.1 Update Quick Create user documentation to mention group selection
- [x] 7.2 Update CLI help text and examples with `--groups` flag
- [x] 7.3 Add troubleshooting section for group assignment errors
- [x] 7.4 Update API documentation if needed (likely no changes)

## 8. Manual Integration Testing

- [x] 8.1 Test end-to-end Quick Create with groups in Web UI (with real backend)
- [x] 8.2 Test end-to-end CLI quick create with groups (with real backend)
- [x] 8.3 Verify group assignments appear in schedule details
- [x] 8.4 Verify orphaned schedules don't occur during rollback scenarios
- [x] 8.5 Test with user who has no groups
- [x] 8.6 Test with user who has many groups (10+) to verify performance
