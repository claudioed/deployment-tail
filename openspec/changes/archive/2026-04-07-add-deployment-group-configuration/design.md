## Context

The Quick Create feature currently allows users to create deployment schedules with minimal fields (service, environment, time). Groups exist as a separate organizational feature, but can only be assigned after schedule creation through a separate workflow. This design extends Quick Create to support optional group assignment at creation time.

**Current State:**
- Quick Create: 3 required fields (service, environment, time)
- Group assignment: Separate API calls after schedule creation
- Groups API: Supports listing user's groups with favorites-first ordering
- CLI: `schedule quick` command for fast schedule creation

**Constraints:**
- Must maintain hexagonal architecture (domain → application → adapters)
- Must preserve backward compatibility (group assignment is optional)
- Must work with existing group and schedule entities
- Must support both Web UI and CLI interfaces
- Must respect authentication requirements

## Goals / Non-Goals

**Goals:**
- Enable group selection during Quick Create (both Web UI and CLI)
- Leverage existing group assignment APIs (no new backend endpoints)
- Show favorited groups first in selection list
- Support multi-select for assigning to multiple groups
- Maintain transactional consistency (schedule + group assignments succeed or fail together)

**Non-Goals:**
- Creating new groups inline during Quick Create
- Modifying existing group CRUD functionality
- Adding group fields to full schedule creation form (separate concern)
- Changing group assignment validation rules

## Decisions

### Decision 1: Reuse existing group assignment API

**Choice:** Call existing POST /schedules/:id/groups endpoint after schedule creation, rather than adding groups to POST /schedules payload.

**Rationale:**
- Maintains clean separation of concerns (schedule creation vs. group assignment)
- No OpenAPI schema changes needed
- Backend already handles group assignment validation
- Easier rollback (no API changes)

**Alternatives Considered:**
- Add `groupIds` field to POST /schedules → Rejected: Violates single responsibility, requires API changes
- Add bulk assignment endpoint → Rejected: Over-engineering for this use case

### Decision 2: Client-side orchestration for atomicity

**Choice:** Web UI and CLI handle the 2-step flow (create schedule → assign groups) with rollback on failure.

**Rationale:**
- Backend endpoints already exist and work correctly
- Client has context for error handling and user feedback
- Avoids backend changes and testing overhead

**Error Handling:**
- If schedule creation fails → No cleanup needed, return error
- If group assignment fails → Delete the just-created schedule, return error

**Alternatives Considered:**
- Backend transaction spanning both operations → Rejected: Requires new endpoint, increases complexity
- Fire-and-forget group assignment → Rejected: Leaves incomplete state, poor UX

### Decision 3: Multi-select UI with favorites-first ordering

**Choice:** Use checkbox-based multi-select, ordered by favorites first (matching group list API behavior).

**Rationale:**
- Consistent with existing group list UX patterns
- Fast access to frequently-used groups
- Supports common use case (assigning to multiple related groups)

**Alternatives Considered:**
- Dropdown single-select → Rejected: Limits users to one group
- Search-based autocomplete → Rejected: Adds complexity for unclear benefit

### Decision 4: CLI accepts comma-separated group IDs or names

**Choice:** `--groups` flag accepts either group IDs or group names: `--groups "id1,id2"` or `--groups "Project Alpha,Team Backend"`.

**Rationale:**
- IDs for scripting/automation (stable, unambiguous)
- Names for human usability (readable, memorable)
- CLI resolves names → IDs before API calls

**Implementation:**
- Try parsing as UUIDs first → Use as IDs
- Otherwise treat as names → Resolve via list groups API
- Error if name ambiguous or not found

**Alternatives Considered:**
- IDs only → Rejected: Poor developer experience
- Names only → Rejected: Fragile for automation

## Risks / Trade-offs

### Risk: Race condition between schedule creation and group assignment

**Scenario:** User creates schedule, assignment call fails midway, schedule exists without groups.

**Mitigation:**
- Client detects assignment failure
- Client calls DELETE /schedules/:id to rollback
- Show clear error message to user
- User can retry (idempotent operation)

### Trade-off: Two API calls instead of one

**Impact:** Slight latency increase (~100-200ms for second call).

**Justification:**
- Acceptable for Quick Create use case (still faster than manual workflow)
- Avoids backend changes and API versioning
- Maintains clean architecture

### Risk: Schedule deleted during group assignment rollback, but deletion fails

**Scenario:** Orphaned schedule exists, but user thinks operation failed completely.

**Mitigation:**
- Low probability (deletion is simple operation)
- If occurs, user sees schedule in list and can delete manually
- Future enhancement: Add audit log for debugging

### Trade-off: Name resolution increases CLI complexity

**Impact:** CLI must fetch group list to resolve names.

**Justification:**
- Worth the tradeoff for UX improvement
- Caching can reduce API calls in scripts
- Error messages guide users on ambiguity

## Migration Plan

**Deployment Steps:**
1. Deploy Web UI changes (new group selector in Quick Create modal)
2. Deploy CLI changes (add `--groups` flag to `schedule quick` command)
3. No database migrations needed (reuses existing schema)
4. No backend changes needed (reuses existing endpoints)

**Rollback Strategy:**
- Remove `--groups` flag from CLI
- Remove group selector from Web UI
- No data cleanup needed (group assignments remain valid)

**Backward Compatibility:**
- Existing Quick Create usage unchanged
- `--groups` flag optional, defaults to no groups
- Web UI shows group selector but empty selection is valid

## Open Questions

None. Design is straightforward extension of existing functionality.
