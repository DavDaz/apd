# Apply Progress: introduce-guided-document-cli

## Current Slice

PR 3 — Markdown export, AI Context Pack, final verification/docs.

Dependency diagram:

```text
PR 1: Go module + CLI skeleton + templates + validation ✅
   ↓
PR 2: document model + guided stdio workflow + session persistence ✅
   ↓
📍 PR 3: Markdown + AI Context Pack + final verification/docs
```

## Completed Tasks

### PR 1

- Initialized Go module metadata in `go.mod` and `go.sum`.
- Added CLI entrypoint and thin command routing in `main.go`, `cmd/root.go`, and `cmd/new.go`.
- Added template schema, strict YAML loading, validation, registry resolution, embedded runtime templates, and template tests.
- Added concise bundled templates under `templates/` and embedded copies under `internal/templates/bundled/`.
- Fixed bundled template loading to use Go `embed` instead of current-working-directory lookup.
- Fixed template question YAML from ambiguous flow lists to block lists.

### PR 2

- Added document domain model and tests in `internal/document/`.
- Added guided stdio parsing, menu, commands, rendering, and tests in `internal/cli/`.
- Added `apd new` orchestration and tests in `internal/app/new_document.go`.
- Added project-local YAML session persistence and tests in `internal/storage/`.
- Fixed no-arg menu selection so one newline submits a numeric/id choice.
- Added `.apd/` to `.gitignore`.

### PR 3

- Added Markdown rendering and project-local file export in `internal/generator/markdown.go`.
- Added AI Context Pack mapping in `internal/generator/ai_context.go`.
- Added generator tests covering answered-only, skipped, pending, and `/done` with no answers.
- Wired final export into `internal/app/new_document.go`.
  - Normal completion writes Markdown and prints Markdown + session paths.
  - `/done` saves the current session, writes Markdown, and prints both paths.
  - Markdown export failure returns an error after session state is already saved.
- Updated product template to include `problem` and `goal` sections so smoke verification can exercise `/back` replacement and Context/Goals mapping.
- Added `apd-docs/` to `.gitignore` for generated local Markdown output.
- Updated `openspec/config.yaml` with `go test ./...` and `go vet ./...` verification commands.
- Updated PR 3 checkboxes in `tasks.md`.

## Files Changed

- `.gitignore`
- `go.mod`
- `go.sum`
- `main.go`
- `cmd/root.go`
- `cmd/new.go`
- `internal/app/new_document.go`
- `internal/app/new_document_test.go`
- `internal/cli/input.go`
- `internal/cli/menu.go`
- `internal/cli/commands.go`
- `internal/cli/renderer.go`
- `internal/cli/input_test.go`
- `internal/document/document.go`
- `internal/document/section.go`
- `internal/document/metadata.go`
- `internal/document/document_test.go`
- `internal/generator/ai_context.go`
- `internal/generator/markdown.go`
- `internal/generator/markdown_test.go`
- `internal/storage/paths.go`
- `internal/storage/session.go`
- `internal/storage/filesystem.go`
- `internal/storage/filesystem_test.go`
- `internal/templates/schema.go`
- `internal/templates/loader.go`
- `internal/templates/validate.go`
- `internal/templates/registry.go`
- `internal/templates/validate_test.go`
- `internal/templates/registry_test.go`
- `internal/templates/bundled/product.yaml`
- `internal/templates/bundled/change-request.yaml`
- `internal/templates/bundled/feature.yaml`
- `internal/templates/bundled/bug.yaml`
- `internal/templates/bundled/task.yaml`
- `templates/product.yaml`
- `templates/change-request.yaml`
- `templates/feature.yaml`
- `templates/bug.yaml`
- `templates/task.yaml`
- `openspec/config.yaml`
- `openspec/changes/introduce-guided-document-cli/tasks.md`
- `openspec/changes/introduce-guided-document-cli/apply-progress.md`

## Test Commands Run

| Command | Result | Notes |
| --- | --- | --- |
| `go test ./...` | Passed | Covers app, cli, document, generator, storage, and templates packages. |
| `go vet ./...` | Passed | No vet issues reported. |
| `git diff --check` | Passed | No whitespace errors reported. |
| `printf 'My answer\n\nMy goal\n\n' \| go run . new product` | Passed | Generated session and Markdown; Markdown includes answered sections and Context/Goals AI Context Pack content. |
| `printf '/skip\n/skip\n' \| go run . new product` | Passed | Generated Markdown with skipped sections in `Open / Skipped Sections`. |
| `printf '/done\n' \| go run . new product` | Passed | Generated Markdown with pending sections and Pending AI Context Pack headings. |
| `printf '/help\nMy answer\n\nMy goal\n\n' \| go run . new product` | Passed | Help output did not prevent final saved answer state. |
| `printf 'first\n\n/back\nreplacement\n\ngoal\n\n' \| go run . new product` | Passed | Session persisted replacement answer after `/back`. |
| `go run . new unknown` | Passed | Returns clear unsupported type error with non-zero exit as expected. |

## TDD Cycle Evidence

Strict TDD is not active (`openspec/config.yaml` has `strict_tdd: false`). Tests were added and run for each PR slice, including PR 3 generator/app export behavior.

## Deviations From Design

- The design allowed Cobra if a CLI framework was introduced. This implementation uses a small standard-library command dispatcher instead, keeping dependencies lower while preserving package boundaries.
- Default template resolution uses embedded runtime templates under `internal/templates/bundled`; top-level `templates/` remains the source/distribution copy for review and future template work.
- Markdown output is written to project-local `apd-docs/` instead of `.apd/` so generated documents are easy to find and manually edit. Session state remains under `.apd/sessions`.
- No roadmap placeholder commands were added because no roadmap commands were introduced.

## Remaining Tasks

- Run a final fresh review before any PR handoff.
- Optional future slices: richer templates, resume/edit/export/backlog/prompts workflows, custom template authoring, prompt libraries/TUI, and package/embed strategy refinements.

## Workload / PR Boundary

- Chosen strategy: chained PRs.
- Current boundary: PR 3 only, stacked after PR 1 and PR 2.
- Out of scope for this slice: AI/model calls, web/cloud/database, advanced TUI, custom template authoring, full backlog generation, full prompt pack generation, and complete roadmap commands.
- PR 3 adds generator/export behavior, tests, verification config, and docs while preserving PR 1/PR 2 behavior.
