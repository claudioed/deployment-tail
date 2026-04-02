## Context

The current deployment-tail system supports basic CRUD operations for deployment schedules but lacks ownership tracking, approval workflows, and a user-friendly web interface. This change adds these capabilities while maintaining the existing hexagonal architecture and DDD principles.

Current state:
- REST API with OpenAPI specification
- CLI tool for schedule management
- MySQL database with basic schedule fields
- No owner tracking or approval mechanism
- No web UI (CLI only)

## Goals / Non-Goals

**Goals:**
- Add owner tracking to all schedules (immutable after creation)
- Implement approval workflow with three states: created, approved, denied
- Add rollback plan field for operational safety
- Create responsive web UI for schedule management
- Maintain existing architecture patterns (hexagonal + DDD)
- Keep backward compatibility where possible
- Extend filtering capabilities (by owner, status)

**Non-Goals:**
- User authentication system (defer to future iteration)
- Complex approval workflows with multiple approvers
- Rollback plan execution automation
- Email/notification system for approvals
- Advanced UI features (dashboards, analytics, charts)
- Mobile native apps

## Decisions

### Decision 1: Database Schema Changes

**Decision**: Add three new columns to existing `schedules` table rather than creating new tables.

**Schema changes**:
```sql
ALTER TABLE schedules
  ADD COLUMN owner VARCHAR(255) NOT NULL,
  ADD COLUMN status ENUM('created', 'approved', 'denied') NOT NULL DEFAULT 'created',
  ADD COLUMN rollback_plan TEXT,
  ADD INDEX idx_owner (owner),
  ADD INDEX idx_status (status);
```

**Rationale**:
- Simple extension of existing model
- No join complexity
- Maintains single aggregate root (Schedule)
- Easy rollback (just drop columns)

**Alternatives considered**:
- Separate approval_status table: Too complex for simple status field
- Separate owners table: Over-engineering for username storage

### Decision 2: Domain Model Updates

**Decision**: Add three new value objects to Schedule aggregate.

**New value objects**:
- `Owner`: Validates format, immutable after creation
- `Status`: Enum with transition validation (created → approved/denied only)
- `RollbackPlan`: Optional text field with length validation

**Rationale**:
- Follows existing DDD patterns
- Encapsulates validation logic
- Type safety

**Schedule entity changes**:
```go
type Schedule struct {
    id          ScheduleID
    scheduledAt ScheduledTime
    service     ServiceName
    environment Environment
    description Description
    owner       Owner          // NEW
    status      Status         // NEW
    rollbackPlan RollbackPlan  // NEW
    createdAt   time.Time
    updatedAt   time.Time
}
```

### Decision 3: API Design for Approval

**Decision**: Add dedicated approval endpoints rather than using generic update.

**New endpoints**:
- `POST /api/v1/schedules/{id}/approve` - Approve schedule
- `POST /api/v1/schedules/{id}/deny` - Deny schedule

**Rationale**:
- Clear intent and semantic meaning
- Easier to add authorization later
- Prevents accidental status changes via update
- Better audit trail potential

**Alternatives considered**:
- PATCH /schedules/{id} with status field: Less explicit, harder to secure
- PUT /schedules/{id}/status: Non-standard REST pattern

### Decision 4: Web UI Technology Stack

**Decision**: Use vanilla HTML/CSS/JavaScript with modern features (ES6+, Fetch API).

**Rationale**:
- No build step required
- Easy to deploy (static files)
- Minimal dependencies
- Fast loading
- Easy for contributors to understand

**File structure**:
```
web/
├── index.html
├── styles.css
├── app.js
├── api-client.js
└── components/
    ├── schedule-list.js
    ├── schedule-form.js
    └── schedule-detail.js
```

**Alternatives considered**:
- React: Overkill for simple CRUD UI, adds build complexity
- Vue: Similar to React, unnecessary for this scope
- Server-side templates (Go html/template): Less interactive, page reloads

**Trade-off**: Vanilla JS means more manual DOM manipulation, but avoids framework lock-in and build complexity.

### Decision 5: Owner Field Behavior

**Decision**: Owner is immutable after schedule creation.

**Rationale**:
- Clear audit trail (who created the schedule)
- Prevents ownership transfer confusion
- Simplifies authorization model
- Aligns with real-world ownership semantics

**Implementation**:
- Domain layer enforces immutability
- Update method rejects owner changes
- API returns 400 if owner modification attempted

**Alternative**: Allow owner transfer → Rejected due to audit complexity

### Decision 6: Default Owner Value

**Decision**: Require explicit owner parameter (no automatic detection).

**Rationale**:
- Explicit is better than implicit
- Works consistently across API and CLI
- No dependency on authentication system
- User must think about ownership

**Future**: When authentication added, can default to authenticated user.

**Alternative**: Use system username → Rejected because CLI might run on servers where username is misleading

### Decision 7: Status Transition Validation

**Decision**: Enforce valid status transitions in domain layer.

**Valid transitions**:
- created → approved ✓
- created → denied ✓
- approved → denied ✗
- denied → approved ✗
- approved → created ✗
- denied → created ✗

**Rationale**:
- Prevents confusion and invalid states
- Final states (approved/denied) are immutable
- Clear workflow: create → decide (approve or deny)

**Implementation**: Status value object validates transitions.

### Decision 8: Migration Strategy

**Decision**: Use database migration with backward-compatible rollback.

**Forward migration** (000002_add_owner_status_rollback.up.sql):
```sql
ALTER TABLE schedules
  ADD COLUMN owner VARCHAR(255) NOT NULL DEFAULT 'system',
  ADD COLUMN status ENUM('created', 'approved', 'denied') NOT NULL DEFAULT 'approved',
  ADD COLUMN rollback_plan TEXT;
```

Note: Existing schedules get default values ('system' owner, 'approved' status) to indicate they predate the approval system.

**Backward migration** (000002_add_owner_status_rollback.down.sql):
```sql
ALTER TABLE schedules
  DROP COLUMN rollback_plan,
  DROP COLUMN status,
  DROP COLUMN owner;
```

**Data preservation**: Before rollback, export schedules with new fields for potential recovery.

## Risks / Trade-offs

**[Risk: Breaking change for API clients]** → Mitigate with API versioning (v1 without owner, v2 with owner) or make owner optional initially

**[Risk: Existing schedules have fake owner]** → Acceptable trade-off; clearly documented that 'system' owner means pre-existing schedule

**[Risk: Web UI and API can drift]** → Mitigate by using OpenAPI spec as source of truth; consider contract testing

**[Risk: No authentication means anyone can approve]** → Acceptable for v1; add auth in next iteration. Document this limitation clearly.

**[Trade-off: Immutable owner]** → Can't transfer ownership, but simpler audit trail and prevents confusion

**[Trade-off: Simple status model]** → No complex workflows (multi-level approval), but easier to understand and implement

**[Risk: Rollback plan not validated]** → Users might enter invalid plans. Acceptable since it's documentation, not executable code. Future: Add validation or templates.

**[Risk: Web UI increases attack surface]** → Mitigate with input sanitization, CSP headers. No sensitive data exposed since no auth required anyway.

## Migration Plan

**Pre-deployment**:
1. Test migration on copy of production database
2. Backup production database
3. Document new fields for users

**Deployment**:
1. Apply database migration (adds columns with defaults)
2. Deploy updated API server (backward compatible)
3. Deploy web UI static files
4. Update CLI to latest version
5. Update documentation

**Post-deployment verification**:
1. Verify existing schedules still accessible
2. Create new schedule with owner field
3. Test approval workflow
4. Test web UI functionality
5. Monitor error logs

**Rollback**:
1. Export all schedule data (including new fields)
2. Run down migration (removes new columns)
3. Deploy previous API version
4. Remove web UI files
5. Restore previous CLI version
6. Verify old functionality works

**Rollback caveat**: Schedules created after deployment will lose owner/status data during rollback. Save to backup first.

## Open Questions

1. **Owner format**: Should we enforce email format, or allow any string (username, email, system id)?
   - Lean toward flexible string to support different org structures

2. **Web UI deployment**: Should UI be served by Go API server or separate static file server?
   - Recommend: Serve from Go API (add static file handler) for simpler deployment

3. **Status history**: Should we track when status changed and who changed it?
   - Defer to future iteration (needs approval_history table)

4. **Rollback plan templates**: Should we provide predefined rollback templates?
   - Defer to future iteration; start with free-form text

5. **UI framework**: If vanilla JS becomes too complex, when do we switch to framework?
   - Threshold: If DOM manipulation code exceeds 500 lines or state management becomes unwieldy

6. **API versioning**: When to introduce /api/v2 with owner as required field?
   - Recommend: Next major release (v2.0.0), for now keep v1 with owner optional
