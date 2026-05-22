# SDD Apply Result — PR 3: introduce-guided-document-cli

## status

completed

## executive_summary

Implemented only chained PR slice 3 for `introduce-guided-document-cli`: Markdown export, basic AI Context Pack generation, final `apd new` export wiring, verification config/docs, and final smoke verification.

The slice deliberately does **not** introduce AI/model calls, web/cloud/database features, advanced TUI, custom template authoring, full backlog generation, full prompt pack generation, or roadmap command implementations.

Markdown output is now written to project-local `apd-docs/`, while session state remains in `.apd/sessions`. Normal completion and `/done` both save a session, generate Markdown, and print both paths. Markdown export failures return an error after the session has already been persisted.

Memory note: Engram memory tools were not available in this delegated subagent toolset, so no memory save was performed.

## artifacts

- `internal/generator/markdown.go`
- `internal/generator/ai_context.go`
- `internal/generator/markdown_test.go`
- `internal/app/new_document.go`
- `internal/app/new_document_test.go`
- `internal/document/document.go`
- `internal/document/document_test.go`
- `templates/product.yaml`
- `internal/templates/bundled/product.yaml`
- `.gitignore`
- `openspec/config.yaml`
- `openspec/changes/introduce-guided-document-cli/tasks.md`
- `openspec/changes/introduce-guided-document-cli/apply-progress.md`
- `sdd-apply-pr3-introduce-guided-document-cli.md`

## tests

| Command | Result |
| --- | --- |
| `go test ./...` | Passed |
| `go vet ./...` | Passed |
| `git diff --check` | Passed |
| `printf 'My answer\n\nMy goal\n\n' \| go run . new product` | Passed; generated session and Markdown with answered sections and Context/Goals AI Context Pack content |
| `printf '/skip\n/skip\n' \| go run . new product` | Passed; generated skipped-section appendix |
| `printf '/done\n' \| go run . new product` | Passed; generated pending-section appendix and Pending AI Context Pack headings |
| `printf '/help\nMy answer\n\nMy goal\n\n' \| go run . new product` | Passed; help did not prevent final saved answer state |
| `printf 'first\n\n/back\nreplacement\n\ngoal\n\n' \| go run . new product` | Passed; session persisted replacement answer after `/back` |
| `go run . new unknown` | Passed; clear unsupported type error and non-zero exit as expected |

## changed_files

- Added generator package files/tests: `internal/generator/ai_context.go`, `internal/generator/markdown.go`, `internal/generator/markdown_test.go`.
- Updated app orchestration/tests: `internal/app/new_document.go`, `internal/app/new_document_test.go`.
- Fixed session ID nanosecond uniqueness: `internal/document/document.go`, `internal/document/document_test.go`.
- Expanded product template to two sections for product goal capture and `/back` smoke verification: `templates/product.yaml`, `internal/templates/bundled/product.yaml`.
- Updated generated-output ignores: `.gitignore`.
- Updated OpenSpec verification config and SDD tracking: `openspec/config.yaml`, `openspec/changes/introduce-guided-document-cli/tasks.md`, `openspec/changes/introduce-guided-document-cli/apply-progress.md`.

## risks

- `apd-docs/` is project-local generated output and is gitignored by default; users who want to commit generated documents will need to move/copy them or adjust ignore policy later.
- Runtime templates are embedded copies under `internal/templates/bundled` plus top-level source copies under `templates/`; future template edits must keep them in sync until a single-source embed strategy is chosen.
- Product template now has two sections, but other bundled templates remain intentionally concise.
- No roadmap commands were introduced, so placeholder behavior was not added.

## next_recommended

1. Run a fresh review of PR 3 focused on Markdown correctness, AI Context Pack non-fabrication, export failure behavior, and scope boundaries.
2. If review passes, proceed to final SDD verify/archive or PR handoff preparation.
3. Consider a later follow-up to make embedded templates single-source and to decide whether generated `apd-docs/` should remain ignored by default.

## skill_resolution

paths-injected

Loaded:

- `/Users/dadiaz/.agents/skills/golang-pro/SKILL.md`
- `/Users/dadiaz/.config/opencode/skills/work-unit-commits/SKILL.md`
- `/Users/dadiaz/.pi/agent/npm/node_modules/gentle-pi/skills/chained-pr/SKILL.md`
- `/Users/dadiaz/.config/opencode/skills/cognitive-doc-design/SKILL.md`
