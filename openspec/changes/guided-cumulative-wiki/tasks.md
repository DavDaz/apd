# Tasks: Guided Cumulative Wiki

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 1,270–1,480 total; 280–390 per slice |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 → PR 4 |
| Delivery strategy | auto-chain |
| Chain strategy | feature-branch-chain |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Work Units

| Unit | Start → finish | Lines | Focused test | Runtime harness | Rollback boundary |
|---|---|---:|---|---|---|
| PR 1 | Empty → initialized | 280–340 | `go test ./internal/wiki ./internal/storage` | N/A: no entry | `internal/wiki/*`, initialization paths |
| PR 2 | Initialized → source registered | 340–390 | `go test ./internal/storage ./internal/app` | N/A: storage seam | registration/journal files and raw copies |
| PR 3 | Registered → handoff pending | 290–350 | `go test ./internal/generator ./internal/app` | N/A: no command | renderer, request files, lifecycle transition |
| PR 4 | Services → `apd wiki` dashboard | 360–400 | `go test ./cmd ./internal/cli/tui ./internal/app` | `d=$(mktemp -d); APD_TUI=off go run . wiki "$d/wiki"` | `cmd/wiki.go`, wiki TUI, README entry |

Routing: PR #1 targets the feature/tracker integration branch. Later children target their immediate predecessor (#2 → #1, #3 → #2, #4 → #3). Only tracker/integration merges to main; retarget/rebase polluted diffs.

Excluded: models, semantic edits, query/RAG, sync, migration, lint/index automation, web UI.

## Phase 1: PR 1 — Domain and Safe Initialization

- [x] 1.1 **RED** Add tests in `internal/wiki/{model,lifecycle}_test.go` and `internal/storage/wiki_store_test.go` for IDs, next action, layout, collisions, escapes, direct and ancestor symlinks. **Dep:** —; **Accept:** fails; **Rollback:** tests; **Exclude:** registration.
- [x] 1.2 **GREEN** Create `internal/wiki/{model,lifecycle}.go` and initialize `internal/storage/wiki_store.go` with exclusive paths, versioned manifest. **Dep:** 1.1; **Accept:** passes; **Rollback:** domain/layout; **Exclude:** journals/CLI.

## Phase 2: PR 2 — Registration and Atomic Storage

- [x] 2.1 **RED** Extend `internal/storage/{wiki_store,atomic}_test.go` for regular files, symlink resolution, idempotency, changed bytes, permissions, interrupted commit, restart repair. **Dep:** PR 1; **Accept:** fails; **Rollback:** tests; **Exclude:** request rendering.
- [x] 2.2 **GREEN** Add `internal/app/wiki_workspace.go` registration use case and `internal/storage/atomic.go` journals, immutable receipts/raw copies, fsync/rename recovery. **Dep:** 2.1; **Accept:** passes without overwrite; **Rollback:** registration artifacts; **Exclude:** handoff/UI.

## Phase 3: PR 3 — Deterministic Handoff

- [ ] 3.1 **RED** Add `internal/generator/integration_request_test.go` and `internal/app/wiki_workspace_test.go` for stable YAML/fields, contradictions, incomplete emission, pending-only status. **Dep:** PR 2; **Accept:** fails; **Rollback:** tests; **Exclude:** command wiring.
- [ ] 3.2 **GREEN** Create `internal/generator/integration_request.go`; wire request-ready/explicit confirmation transitions in `internal/app/wiki_workspace.go` without wiki edits or “integrated” status. **Dep:** 3.1; **Accept:** passes; **Rollback:** renderer/request transition; **Exclude:** TUI.

## Phase 4: PR 4 — Guided Entry and Compatibility

- [ ] 4.1 **RED** Add `cmd/wiki_test.go`, `cmd/root_test.go`, and `internal/cli/tui/wiki_model_test.go` for non-TTY snapshot/no mutation, resume/status, narrow layout/error, unchanged `apd new`. **Dep:** PR 3; **Accept:** fails; **Rollback:** tests; **Exclude:** legacy conversion.
- [ ] 4.2 **GREEN** Add `cmd/wiki.go`, route it in `cmd/root.go`, and create `internal/cli/tui/wiki_model.go`; update `README.md` with `apd wiki` and pending-handoff meaning. **Dep:** 4.1; **Accept:** `go test ./... && go vet ./...` pass; **Rollback:** entry/UI/docs only; **Exclude:** semantic integration.
