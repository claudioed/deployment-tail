## Why

When users create deployment schedules via Quick Create, they currently must create the schedule first, then separately assign it to groups in a second step. This breaks workflow momentum and requires navigating away from the creation context. Adding group selection directly to Quick Create enables users to categorize deployments at creation time.

## What Changes

- Add optional group selection to Quick Create interface (Web UI)
- Add `--groups` flag to CLI `schedule quick` command
- Support multi-select group assignment during schedule creation
- Maintain backward compatibility (group assignment remains optional)
- Show user's favorited groups first in selection list for faster access

## Capabilities

### New Capabilities
- `quick-create-with-groups`: Group assignment capability integrated into Quick Create workflow

### Modified Capabilities
- `quick-create`: Extend Quick Create to support optional group assignment at creation time

## Impact

**Affected Code:**
- Web UI: Quick Create modal component (add group multi-select field)
- CLI: `schedule quick` command (add `--groups` flag)
- API: No new endpoints required (reuses existing POST /schedules and assignment endpoints)
- Application layer: Extend quick create use case to handle group assignments

**User Experience:**
- Reduces clicks from 2 actions (create + assign) to 1 action
- Maintains optional nature (groups can still be assigned later)
- Leverages existing favorites system for smart defaults

**No Breaking Changes:**
- Group assignment is optional
- Existing Quick Create behavior unchanged when groups not specified
- All existing API contracts remain compatible
