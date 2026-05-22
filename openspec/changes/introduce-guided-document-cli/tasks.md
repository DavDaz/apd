# Tasks: Introduce Guided Document CLI

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 1,500-2,500 additions/deletions |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Go module, CLI skeleton, template schema/loader/tests → PR 2: document model, guided stdio workflow, session persistence/tests → PR 3: Markdown + AI Context Pack generation, command placeholders, verification/docs |
| Delivery strategy | auto-chain |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

## Implementation Tasks

### PR 1 — Go module, command skeleton, templates, and validation

- [x] Initialize the Go module in `go.mod` and `go.sum` for the `apd` CLI.
  - Start: repository has no runnable Go CLI module for this change.
  - Finish: `go test ./...` can run after package skeletons exist.
  - Verify: attempted `go test ./...`; blocked because `go` is not installed in this execution environment.
  - Rollback: remove `go.mod`, `go.sum`, and introduced Go files.

- [x] Add CLI entrypoint and command wiring in `main.go`, `cmd/root.go`, and `cmd/new.go`.
  - Implement root help and `apd new [type]` argument handling.
  - Keep command wiring thin; do not place persistence, generation, or guided workflow logic in `cmd`.
  - Verify: `go run` could not be executed because `go` is not installed in this execution environment.

- [x] Define template schema and validation in `internal/templates/schema.go`, `internal/templates/validate.go`, `internal/templates/loader.go`, and `internal/templates/registry.go`.
  - Support `id`, `name`, `description`, optional `version`, optional `aliases`, and ordered `sections` with `id`, `title`, `required`, `description`, `help`, `example`, `questions`, and `context_keys`.
  - Fail unknown fields and duplicate template ids, aliases, or section ids.
  - Verify: table-driven tests added in `internal/templates/*_test.go`; execution blocked because `go` is not installed.

- [x] Add bundled YAML templates under `templates/product.yaml`, `templates/change-request.yaml`, `templates/feature.yaml`, `templates/bug.yaml`, and `templates/task.yaml`.
  - Keep templates concise enough for MVP review while covering required guided sections and AI Context Pack mappings.
  - Do not enable a working `custom` template route in this PR.
  - Verify: template registry tests load every bundled template and validate aliases; execution blocked because `go` is not installed.

### PR 2 — Document model, guided stdio workflow, and session persistence

- [x] Add document domain types in `internal/document/document.go`, `internal/document/section.go`, and `internal/document/metadata.go`.
  - Model metadata, template reference, ordered section state, answer text, `pending`, `answered`, and `skipped` statuses.
  - Verify: `go test ./...` covers answer, skip, replace, `/back` navigation effects, and unchanged state for help.

- [x] Add guided CLI parsing and rendering in `internal/cli/input.go`, `internal/cli/menu.go`, `internal/cli/commands.go`, and `internal/cli/renderer.go`.
  - Support `/help`, `/skip`, `/back`, `/done`, normal answer text, and line-by-line multi-line entry submitted by an empty line after content.
  - Recognize slash commands only when the current answer buffer is empty.
  - Verify: `go test ./...` covers all commands, normal text, multi-line content, and command-like text inside answer content.

- [x] Add `apd new` orchestration in `internal/app/new_document.go`.
  - Resolve selected type or present the supported type menu.
  - Load the selected template, initialize document state, loop through sections, apply user intents, and call storage after accepted answer/skip/replacement.
  - Verify: `go test ./...` uses fake registry, fake input/output, and fake storage to cover completion, `/done`, `/back`, and unsupported type paths.

- [x] Add project-local session persistence in `internal/storage/paths.go`, `internal/storage/session.go`, and `internal/storage/filesystem.go`.
  - Store YAML sessions under `<working-directory>/.apd/sessions` using `filepath`.
  - Use schema version `1`, deterministic session metadata, and atomic temp-file-then-rename writes.
  - On write failure, return a clear error with the attempted path and remediation guidance.
  - Verify: `go test ./...` uses temporary directories for path resolution, YAML round trip, atomic replacement, and write failure behavior.

### PR 3 — Markdown export, AI Context Pack, placeholders, and verification

- [x] Add Markdown generation in `internal/generator/markdown.go`.
  - Include metadata, answered sections in template order, and an `Open / Skipped Sections` appendix for skipped or pending sections.
  - Do not invent placeholder prose inside answered sections.
  - Verify: golden tests cover answered-only, skipped, pending, and `/done` with no answers.

- [x] Add basic AI Context Pack generation in `internal/generator/ai_context.go`.
  - Use `context_keys` mappings and explicit metadata only.
  - Emit stable headings and mark absent mapped content as `Pending` or omit it according to the design contract.
  - Verify: golden tests prove captured content is included and absent content is not fabricated.

- [x] Wire final export behavior through `internal/app/new_document.go` and `cmd/new.go`.
  - On normal completion or `/done`, write Markdown to a project-local output path and print both Markdown and session paths.
  - Ensure Markdown write failures leave the session file intact and return non-zero CLI exit behavior.
  - Verify: app-level tests cover successful export and generator failure.

- [x] Add roadmap placeholder behavior only if commands are introduced in `cmd/*.go`.
  - For `apd backlog`, `apd prompts`, `apd edit`, `apd validate`, or `apd template create`, either omit the command or return a clear `not implemented yet` message with no output files.
  - Verify: command tests or manual checks confirm placeholders do not create partial artifacts.

- [x] Update verification documentation and OpenSpec config after `go test ./...` is available.
  - Update `openspec/config.yaml` test command fields or add project docs if config changes are deferred by the apply owner.
  - Verify: run `go test ./...` and record manual smoke results.

## Manual Smoke Verification Checklist

- [x] Run `go run . new product` and complete two product sections.
- [x] Confirm `.apd/sessions/<session>.session.yaml` exists and contains captured answers.
- [x] Use `/help` and confirm it does not change saved answer state.
- [x] Use `/skip` and confirm skipped state persists.
- [x] Use `/back`, replace a previous answer, and confirm the session file updates.
- [x] Use `/done` and confirm Markdown is generated.
- [x] Inspect Markdown for answered content, skipped/pending appendix entries, and AI Context Pack content or `Pending` markers only.
- [x] Confirm the workflow uses only local files and no AI credentials or network calls.
- [x] Run `go test ./...` before each chained PR handoff once `go.mod` exists.

## Out of Scope for Apply

- AI/model calls, prompt execution, web/API services, cloud sync, databases, collaboration, advanced TUI, custom template authoring UI, full resume/edit/export/backlog/prompts workflows, and user-facing JSON/YAML export beyond internal template/session files.
