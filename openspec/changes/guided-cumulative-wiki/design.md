# Design: Guided Cumulative Wiki — First Implementable Slice

## Technical Approach

Add a separate `wiki` domain beside guided document authoring. `apd wiki [workspace]` opens/resumes a dashboard; application services own transitions, storage owns files, and generators own deterministic handoffs. APD stops at external-agent handoff: no model, recall/query, semantic edits, or automatic migration.

## Domain Boundaries and Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Extend `document.Document` | Reuses code but couples unrelated lifecycles | Create `internal/wiki`; leave legacy document contracts unchanged. |
| Mutable provenance records | Simple, but destroys audit identity | Immutable receipt per registration; mutable pending state references receipts. |
| Timestamp IDs | Familiar but collision-prone | Workspace ID uses 128-bit cryptographic randomness; source/work IDs are content-derived. |
| Best-effort file writes | Small, but interruption can split state | Journal each operation, publish immutable files idempotently, atomically rename authoritative state last. |
| Invoke an agent | Convenient but expands trust/process boundary | Emit files only; users/external tools perform handoff. |

Ownership follows existing patterns: `internal/wiki` validates invariants and derives `NextAction`; `internal/app` orchestrates narrow interfaces; `internal/storage` handles paths/YAML/atomicity; `internal/generator` renders requests; `cmd` and Bubble Tea are adapters only.

## Data and Lifecycle

    source file → RegisterSource → raw copy + receipt → pending work
                                              ↓
    dashboard ← NextAction ← workspace state ← deterministic request

`byte_hash` is `sha256:<64 lowercase hex>` over exact bytes, not semantic truth. `source_id = src-<20 hex>` from SHA-256 of canonical origin path, NUL, and byte hash; unchanged path/bytes is idempotent, changed bytes creates a new identity and visible pending item. `work_id = work-<20 hex>` from source ID plus `integration-v1`. Times are persisted once as UTC RFC3339Nano; struct-based YAML makes rerendered requests byte-stable.

Work transitions are `registered → request_ready → awaiting_agent`. Request publication enables `request_ready`; explicit user confirmation enables `awaiting_agent`. Every state remains pending because this slice cannot assert integration. Restart loads state, validates schema/hash references, repairs an idempotent journal, and derives the same next action. Corrupt/unsupported state is read-only; APD never guesses or rewrites it.

## Contracts and Persistence

```go
type WorkspaceService interface {
    OpenOrInitialize(context.Context, string) (wiki.Snapshot, error)
    RegisterSource(context.Context, wiki.RegisterSource) (wiki.Snapshot, error)
    ConfirmHandoff(context.Context, wiki.WorkID) (wiki.Snapshot, error)
}
type WorkspaceStore interface { Load(context.Context, string) (wiki.Workspace, error); Commit(context.Context, wiki.ChangeSet) error }
type RequestRenderer interface { Render(wiki.IntegrationRequest) ([]byte, error) }
```

Workspace layout is `raw/<source_id>-<safe-base>`, `wiki/`, `.apd/workspace.yaml`, `.apd/sources/<source_id>.yaml`, `.apd/requests/<work_id>.integration.yaml`, and `.apd/transactions/`. Initialization preflights all managed paths, creates them exclusively, records an init marker, and rolls back only paths created by APD; collisions are refused. Commits stage and fsync data, use create-exclusive publication for immutable files, then temp-file/fsync/rename `workspace.yaml`; recovery accepts an existing file only when its hash matches the journal.

Requests include schema/status, workspace/work/source IDs, receipt and raw paths, byte hash/type/notes, allowed `wiki/` scope, required updates, contradiction policy (“report; never silently resolve”), and explicit pending status. They contain no shell command or executable instruction.

## File Changes

| File | Action | Responsibility |
|---|---|---|
| `internal/wiki/{model,lifecycle}.go` | Create | Aggregate, IDs, validation, next action. |
| `internal/app/wiki_workspace.go` | Create | Use cases and ports. |
| `internal/storage/{wiki_store,atomic}.go` | Create | Layout, journals, recovery, atomic YAML. |
| `internal/generator/integration_request.go` | Create | Canonical request rendering. |
| `internal/cli/tui/wiki_model.go` | Create | Dashboard states/actions; explicit truncation and terminal-safe layout. |
| `cmd/wiki.go`, `cmd/root.go` | Create/Modify | Add `wiki`; preserve `new` routing/help/output. |
| Matching `*_test.go` files | Create/Modify | Domain, adapter, recovery, compatibility tests. |

## Entry, Security, and Recovery

`apd wiki` defaults to the current directory; an optional path is canonicalized. TTY uses Bubble Tea; non-TTY prints the snapshot and next action without mutation. Paths are absolute-cleaned, contained with `filepath.Rel`, and output names never use unchecked user segments. Registration accepts regular files only, resolves symlinks to a recorded canonical origin, rejects workspace-internal sources, and detects size/metadata changes while hashing. Permissions are `0700` for `.apd` and `0600` for provenance/state/request files; raw bytes retain no executable mode.

## Testing Strategy and Review Slices

Table-driven tests cover IDs, lifecycle, deterministic YAML, next actions, traversal/symlink/change detection, and registration. Temp-directory tests inject clock/entropy/filesystem failures for collisions, interrupted commits, restart repair, hash mismatch, and permissions. TUI tests cover resume, status, narrow terminals, and errors. Legacy tests remain green.

Delivery is auto-chained under 400 authored lines per PR: (1) domain + initialization, (2) registration + transactional storage, (3) request generation + lifecycle, (4) CLI/TUI + legacy compatibility. Each slice has focused tests and can be reverted independently.

## Threat Matrix

N/A — APD emits inert YAML and adds no shell/subprocess, VCS/PR automation, executable classification, repository routing, or process execution boundary; every matrix row is therefore N/A.

## Migration / Rollout

No migration required. Existing `.apd/sessions`, `apd-docs/`, filenames, templates, and `apd new` behavior are untouched. Legacy continuation cannot initialize a wiki implicitly.

## Open Questions

None blocking this slice.
