# Apply Progress: Guided Cumulative Wiki

## Scope
PR 1 and PR 2 completed cumulatively. This apply batch implemented PR 2 only — registration and transactional storage. Delivery is `auto-chain` with `feature-branch-chain`; PR 2 targets PR 1 when a PR is created. No branch, commit, push, PR, or review lifecycle was created.

## Completed Tasks
- [x] 1.1 RED tests for workspace/source/work IDs, lifecycle next actions, versioned layout, collisions, escapes, direct symlink parents, and symlinked ancestor components.
- [x] 1.2 GREEN `internal/wiki/{model,lifecycle}.go` and `internal/storage/wiki_store.go` safe initialization with every component of the selected existing parent validated by `Lstat` before writes.
- [x] 2.1 RED tests for regular files, source-boundary safety, symlink canonicalization, exact-byte SHA-256, idempotency, changed bytes, permissions, interrupted journal commit, and restart recovery.
- [x] 2.2 GREEN registration use case, immutable raw copies/receipts, fsync plus rename journal writes, and restart recovery.

## RED → GREEN Evidence
| Task | RED evidence | GREEN evidence |
|---|---|---|
| 1.1 | Initial implementation lacked wiki symbols; focused package tests failed before production code existed. | `go test ./internal/wiki ./internal/storage` — exit 0. |
| 1.2 | Ancestor-symlink regression initially failed with `Initialize() error = nil`. | Focused package tests passed after `Lstat` ancestor validation. |
| 2.1 | `go test ./internal/storage ./internal/app -count=1` — exit 1: `Receipt` undefined and `WikiStore.Register` undefined (registration tests written before implementation). | Registration tests pass: exact byte copy/hash, canonical symlink origin, boundary rejection, idempotency, changed-byte preservation, unreadable mode rejection, and journal recovery. |
| 2.2 | Same RED command failed before receipt/store/app symbols existed. | `go test ./internal/storage ./internal/app -count=1` — exit 0; 2 packages passed. |

## Work Unit Evidence
| Focused test command and exact result | Runtime harness command/scenario and exact result | Rollback boundary |
|---|---|---|
| `go test ./internal/storage ./internal/app -count=1` — exit 0; both packages passed. | N/A — PR 2 has no CLI/TUI or other runtime boundary; temp-directory registration and recovery tests exercise the storage seam. | Revert `internal/app/wiki_workspace.go`, `internal/storage/atomic.go`, and PR 2 registration additions in `internal/storage/wiki_store.go` plus their tests; existing `apd new` code remains untouched. |

## Full Verification
- `go test ./... -count=1` — exit 0; all packages passed.
- `go vet ./...` — exit 0; no findings.

## Changed Paths (PR 2)
- `internal/app/wiki_workspace.go`
- `internal/app/wiki_workspace_test.go`
- `internal/storage/atomic.go`
- `internal/storage/wiki_store.go`
- `internal/storage/wiki_store_test.go`
- `openspec/changes/guided-cumulative-wiki/tasks.md`
- `openspec/changes/guided-cumulative-wiki/apply-progress.md`

## Review Budget and Native Attempt Evidence
- Incremental PR 2 authored implementation-and-test lines: 374 additions/deletions (excludes OpenSpec artifacts and pre-existing PR 1 files); within the 400-line budget.
- Request ID: `guided-cumulative-wiki-pr2-20260811-01`
- Token: `sha256:e986555db38a3fbd5522e934978005b34e9baf047ff09305ee0de7586cdc6ea9`
- Settlement inputs: focused tests, full tests, and vet all exited 0; tasks 2.1–2.2 are visibly complete; PR 2 boundary is registration/storage only.

## Remaining Tasks
- [ ] 3.1–3.2 Deterministic handoff.
- [ ] 4.1–4.2 CLI/TUI entry and legacy compatibility.

## PR 2 Recovery Correction

This bounded recovery correction changes only `internal/storage/atomic.go` and `internal/storage/wiki_store_test.go`. Recovery now removes both immutable artifacts and the journal whenever publication is incomplete or mismatched; it retains both artifacts only when both byte hashes match the durable journal. Creation and cleanup fsync their parent directories. Task checkboxes remain unchanged because PR 2 was already marked complete and this correction's focused, full, and vet evidence passed.

```json
{"schema":"gentle-ai.remediation-result/v1","lineage_id":"sha256:8b4bf632e7abfb4015ad6cd3945c106f2464a4b8b5ca8fc6676e66a1e855bb4a","generation":2,"fix_batch":2,"failed_evidence_revision":"sha256:a66072f8f3b21b9801520c928e3161e9bb5b345a152e46f997947e3fb9013997","request_id":"guided-cumulative-wiki-pr2-recovery-20260811","work_unit":"pr2-transaction-recovery","status":"corrected","changed_lines":126}
```
```json
{"schema":"gentle-ai.remediation-evidence/v1","lineage_id":"sha256:8b4bf632e7abfb4015ad6cd3945c106f2464a4b8b5ca8fc6676e66a1e855bb4a","generation":2,"fix_batch":2,"failed_evidence_revision":"sha256:a66072f8f3b21b9801520c928e3161e9bb5b345a152e46f997947e3fb9013997","focused_test":"go test ./internal/storage ./internal/app -count=1: exit 0","full_test":"go test ./... -count=1: exit 0","vet":"go vet ./...: exit 0","runtime_harness":"N/A: PR 2 recovery has no command or UI boundary; temp-directory restart tests exercise the storage boundary.","rollback_boundary":"internal/storage/atomic.go and internal/storage/wiki_store_test.go","evidence_revision":"sha256:4d0357a953098c037b1100ed7138431d8ab61a55b5bdfafdd5d1f6c95aca63f0","digest_recipe":"SHA-256(git diff --binary -- internal/storage/atomic.go internal/storage/wiki_store_test.go + newline + --test-summary-- + newline + exact three command/result lines)"}
```
