# Proposal: Guided Cumulative Wiki

## Intent

APD will guide a local-first cumulative wiki without requiring workflow or command recall. APD owns deterministic state, provenance, and handoffs; external agents own semantic integration and recall/query.

## Product Outcome

Users initialize workspaces, register sources, resume work, and obtain integration requests with explicit status and next action.

## Scope

### Full Initiative
- Guided intake, review, linting, index/log stewardship, contradiction tracking, and legacy migration.

### First Implementable Slice
- Initialize a versioned, collision-safe `raw/`, `wiki/`, `.apd/` workspace.
- Register provenance without modifying source bytes.
- Persist pending work and show the next safe action.
- Emit structured agent requests covering inputs, provenance, expected updates, contradictions, and status.
- Preserve `apd new` behavior and outputs.

### Non-Goals
- Embedded models, semantic edits, autonomous conflict resolution, recall/query, RAG, web UI, cloud sync, or remote ingestion.
- Full initiative linting, index/log automation, and legacy conversion in the first slice.

## Capabilities

### New Capabilities
- `wiki-workspace`: Versioned initialization and next-action navigation.
- `source-provenance`: Immutable registration with identity, path, hash, time, type, and notes.
- `pending-work`: Resumable state with explicit lifecycle status.
- `external-agent-integration`: Structured integration requests and outcomes.
- `legacy-document-authoring`: Compatibility for `apd new` sessions and outputs.

### Modified Capabilities
None; no main OpenSpec capabilities currently exist.

## Approach

Add domain contracts, atomic persistence, and a TUI dashboard over the existing application boundary. Keep Markdown/YAML agent-readable; never label unspecified model behavior “AI-ready.”

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `cmd/`, `internal/app/`, `internal/cli/tui/` | Modified | Entry, resume, next action |
| `internal/storage/`, `internal/document/` | Modified | Workspace, receipts, pending state |
| `internal/generator/` | Modified | Integration-request output |

## Migration and Compatibility

Never rewrite `apd-docs/` or existing sessions. Upgrade legacy sessions only on explicit continuation. Preserve filenames, templates, and `apd new`.

## Risks

| Risk | Mitigation |
|---|---|
| Path collision | Refuse overwrite |
| Handoff mistaken as integrated | Show pending status |
| Hash mistaken as truth | Define byte identity only |
| Review overload | Chain reversible slices |

## Rollback Plan

Remove new entry points and user-created workspace artifacts; retain legacy files and `apd new`. Each slice remains independently revertible.

## Success Criteria

- [ ] Workspace initialization never overwrites existing paths.
- [ ] Re-registering unchanged bytes preserves source identity; changed bytes are detected.
- [ ] Interrupted intake resumes with the same pending state and next action.
- [ ] Every request includes receipt, context, required updates, contradiction policy, and status.
- [ ] Existing `apd new` compatibility tests remain green.

## Proposal Question Round

Resolved: external agents remain semantic authorities; slice one ends at handoff; compatibility outranks automatic migration.
