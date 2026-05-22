# Design: Guided Document CLI

## Summary

Introduce `apd new` as a local-first Go CLI workflow that guides users through bundled document templates, persists progress after each section, and exports a clean Markdown document with a basic AI Context Pack. The first slice optimizes for reliability, reviewability, and plain files over a rich terminal UI.

## Goals and Non-Goals

### Goals

- Provide an `apd new` workflow for predefined document types.
- Load and validate bundled YAML templates.
- Guide users through ordered sections with contextual help and examples.
- Support `/help`, `/skip`, `/back`, and `/done` during section entry.
- Persist session progress incrementally as local plain files.
- Export manually editable Markdown that does not invent missing content.
- Generate a basic AI Context Pack from captured answers and explicit metadata only.
- Establish idiomatic Go package boundaries for later growth.

### Non-Goals for This Change

- AI/model calls, prompt execution, or generated inferences.
- Backlog generation beyond captured user content.
- JSON/YAML export as user-facing output, except internal session persistence.
- Custom template authoring UI.
- Full TUI, web UI, database, cloud sync, or collaboration.
- `apd resume`, `apd edit`, `apd export`, `apd backlog`, `apd prompts`, `apd validate`, and `apd template create` as complete workflows.

## Decisions

| Topic | Decision | Rationale |
| --- | --- | --- |
| Session storage path | Default to a project-local session directory at `<working-directory>/.apd/sessions` on macOS, Linux, and Windows. Use OS path separators through Go `filepath`. | The MVP is local-first and should keep recoverable work next to the project/document context. A project-local path is deterministic, easy to inspect, and cross-platform without hidden OS-specific discovery. |
| Session storage format | Store sessions as YAML files named `<timestamp>-<type>-<slug>.session.yaml` with `schema_version`, metadata, template id/version, current section index, and section states. Write atomically via temp file then rename. | YAML is plain, editable, and already required for templates. Schema versioning enables later `resume` support. Atomic writes reduce corruption risk during interruption. |
| App-level fallback path | Defer automatic fallback. If `.apd/sessions` cannot be created or written, fail with a clear error and suggest changing directories. Future app-level paths may use macOS `~/Library/Application Support/apd`, Linux `${XDG_STATE_HOME:-~/.local/state}/apd`, and Windows `%LOCALAPPDATA%\apd`. | Silent fallback can make progress hard to find. The MVP should be predictable. |
| Input mechanism | Use standard input/output with Go standard library readers for the first slice. Do not add a prompt library yet. | This keeps dependencies and behavior small, scriptable, and easy to test. Prompt libraries can be introduced when richer selection, editing, or TUI behavior is justified. |
| Multi-line entry | Read section answers line-by-line. An answer starts when the user types non-command content and is submitted by an empty line after content. Preserve line breaks between entered content lines. Slash commands are recognized only when the current answer buffer is empty. | This supports long answers without a prompt dependency while avoiding accidental command execution inside answer text. The limitation is acceptable for MVP and should be documented in prompt text. |
| Template schema boundary | Bundle YAML templates with a narrow schema: template `id`, `name`, `description`, optional `version`, optional `aliases`, and ordered `sections`. Each section supports `id`, `title`, `required`, `description`, `help`, `example`, `questions`, and optional `context_keys` for AI Context Pack mapping. | This is enough for guided documents and AI Context Pack extraction while avoiding premature workflow logic. |
| Template features deferred | Defer branching, conditional sections, computed values, custom commands, user-authored templates, remote templates, and arbitrary generator instructions. | These features create validation, security, and UX complexity that can be added after the core template model stabilizes. |
| Markdown export policy | The main Markdown body includes answered sections only, in template order. Skipped and unanswered sections are listed in an `Open / Skipped Sections` appendix with their status and title. The AI Context Pack uses stable headings and either includes captured content or marks a heading as `Pending`. | This keeps the primary document clean while making missing information explicit and preventing invented content. |
| `custom` route | Defer `apd new custom` from the first implementation slice. Do not show Custom as selectable in `apd new` until it works. If `apd new custom` is wired, it must return a clear `not implemented yet` message and no output files. | Custom sections require authoring, validation, and editing decisions that would increase MVP scope and review burden. |
| Go CLI framework | Use Cobra for command wiring if the implementation introduces a CLI framework; keep guided input itself in `internal/cli` using stdio. | Cobra gives stable command/argument behavior and help output. Keeping prompt logic independent prevents framework leakage into domain logic. |
| Verification before configured test command | During apply, add Go unit tests and use `go test ./...` once `go.mod` exists. Until OpenSpec config is updated, verification consists of explicit manual smoke scenarios plus package-level tests run locally. | No test command is configured today, but the implementation can establish one without strict TDD being enabled. |

## Architecture

```txt
apd/
├── main.go
├── cmd/
│   ├── root.go
│   └── new.go
├── internal/
│   ├── app/
│   │   └── new_document.go
│   ├── cli/
│   │   ├── input.go
│   │   ├── menu.go
│   │   ├── commands.go
│   │   └── renderer.go
│   ├── templates/
│   │   ├── schema.go
│   │   ├── loader.go
│   │   ├── registry.go
│   │   └── validate.go
│   ├── document/
│   │   ├── document.go
│   │   ├── section.go
│   │   └── metadata.go
│   ├── storage/
│   │   ├── paths.go
│   │   ├── session.go
│   │   └── filesystem.go
│   └── generator/
│       ├── markdown.go
│       └── ai_context.go
└── templates/
    ├── product.yaml
    ├── change-request.yaml
    ├── feature.yaml
    ├── bug.yaml
    └── task.yaml
```

### Package Responsibilities

| Package | Responsibility | Must Not Do |
| --- | --- | --- |
| `cmd` | Define root command, `new` command, args, flags, help, and exit behavior. | Contain template parsing, session persistence, or document generation logic. |
| `internal/app` | Orchestrate the `new` use case across templates, CLI interaction, storage, and generation. | Know YAML or Markdown formatting details. |
| `internal/cli` | Render prompts, menus, command help, and parse `/help`, `/skip`, `/back`, `/done`. | Persist sessions or mutate domain models directly beyond returning user intents. |
| `internal/templates` | Define YAML schema, load bundled templates, resolve aliases, and validate required fields/order. | Execute template-defined logic or read user session files. |
| `internal/document` | Define document, section state, answer state, skipped state, metadata, and template references. | Depend on CLI, filesystem, YAML, or Markdown packages. |
| `internal/storage` | Resolve `.apd/sessions`, serialize session YAML, and perform atomic writes. | Decide prompt flow or Markdown layout. |
| `internal/generator` | Render Markdown and AI Context Pack from the document model. | Infer facts not present in captured answers or metadata. |

## Data Flow

1. User runs `apd new` or `apd new <type>`.
2. `cmd/new.go` validates arguments and calls `internal/app`.
3. `internal/templates` loads the bundled registry and resolves the selected template or aliases.
4. `internal/app` creates a `document.Document` with metadata and section states.
5. `internal/cli` presents each section and returns one user intent: answer, help, skip, back, or done.
6. `internal/app` applies the intent to the document model.
7. `internal/storage` saves the full session after every accepted answer, skip, or replacement.
8. On completion or `/done`, `internal/generator` writes the Markdown document to the output directory.
9. The CLI prints the Markdown path and session path.

## Contracts

### Template YAML Contract

```yaml
id: product
name: Product Decomposition
description: Guide a new product idea into AI-ready project context.
version: 1
aliases: [product-decomposition]
sections:
  - id: problem
    title: Problem
    required: true
    description: Describe the current pain or need.
    help: Focus on the problem, not the solution.
    example: Users cannot verify whether a document was officially issued.
    questions:
      - What is happening today?
      - Who is affected?
      - What consequence does it create?
    context_keys: [context]
```

Validation rules:

- Template `id`, `name`, `description`, and non-empty `sections` are required.
- Section `id` and `title` are required.
- Template ids, aliases, and section ids must be unique within their registry scope.
- Section order is the YAML order.
- Unknown fields should fail validation during MVP to prevent template drift.

### Session YAML Contract

```yaml
schema_version: 1
session_id: "20260522-153000-product-example"
template_id: product
template_version: 1
document_type: product
title: "Product Decomposition"
created_at: "2026-05-22T15:30:00Z"
updated_at: "2026-05-22T15:35:00Z"
current_section_index: 2
sections:
  - id: problem
    title: Problem
    status: answered
    answer: "Users cannot verify documents without manual calls."
    updated_at: "2026-05-22T15:32:00Z"
  - id: actors
    title: Actors
    status: skipped
    answer: ""
    updated_at: "2026-05-22T15:34:00Z"
```

Section status values:

- `pending`: no answer and not skipped.
- `answered`: answer content is present.
- `skipped`: user explicitly skipped the section.

### Markdown Output Contract

```md
# Product Decomposition

Generated by apd on 2026-05-22.
Template: product v1

## Problem

Users cannot verify documents without manual calls.

## AI Context Pack

### Context

Users cannot verify documents without manual calls.

### Goals

Pending.

### Constraints

Pending.

## Open / Skipped Sections

| Section | Status |
| --- | --- |
| Actors | Skipped |
| Goals | Pending |
```

Rules:

- Preserve template section order.
- Do not emit placeholder prose in answered sections.
- Do not invent AI Context Pack content.
- Escape or normalize only what is necessary to keep valid Markdown.
- Keep generated Markdown free of binary or tool-specific metadata.

## Error Handling

- Unsupported type arguments exit non-zero and list supported type arguments.
- Invalid templates exit non-zero and name the invalid file and field.
- Session write failures exit non-zero with the attempted path and remediation guidance.
- Markdown write failures exit non-zero and leave the session file intact.
- `/back` on the first section displays a friendly message and stays on the current section.
- `/done` with no answers still exports a Markdown file with metadata, pending AI Context Pack headings, and an `Open / Skipped Sections` appendix.

## Tests and Verification

OpenSpec currently has no configured test command and `strict_tdd` is false. The implementation phase should still add automated tests and then update verification guidance once `go.mod` exists.

### Automated Tests to Add During Apply

- `internal/templates`: table-driven validation tests for valid templates, missing metadata, missing section fields, duplicate ids, and unknown fields.
- `internal/cli`: command parser tests for `/help`, `/skip`, `/back`, `/done`, and normal answer text.
- `internal/document`: section status transition tests for answer, skip, replace, and back navigation.
- `internal/storage`: path resolution and YAML round-trip tests using temporary directories.
- `internal/generator`: golden Markdown tests for answered, skipped, pending, and AI Context Pack cases.
- `cmd`: argument resolution tests for supported and unsupported document types where practical.

### Manual Smoke Verification Before a Test Command Is Configured

1. Run `apd new product` and complete two sections.
2. Confirm `.apd/sessions/<session>.session.yaml` exists and contains the captured answers.
3. Use `/help` and confirm it does not change saved answer state.
4. Use `/skip` and confirm skipped state persists.
5. Use `/back`, replace a previous answer, and confirm the session file updates.
6. Use `/done` and confirm Markdown is generated.
7. Inspect Markdown to ensure answered content appears, skipped/pending sections are listed, and AI Context Pack contains only captured content or `Pending`.
8. Run without network access and confirm the workflow does not require credentials or services.

Recommended command once Go files exist:

```bash
go test ./...
```

## Tradeoffs

| Choice | Benefit | Cost / Risk | Mitigation |
| --- | --- | --- | --- |
| Project-local `.apd/sessions` instead of OS app directory | Easy to find, project-scoped, cross-platform. | Sessions are tied to the current working directory and may be created in unexpected folders. | Print the session path after creation and document the working-directory behavior. |
| YAML for both templates and sessions | Plain, editable, one parser dependency. | YAML can be ambiguous if schema validation is loose. | Strict validation and schema versioning. |
| Standard IO instead of prompt library | Small dependency set, scriptable, easier tests. | Less polished UX and limited multi-line editing. | Keep prompt text explicit; revisit prompt library after MVP. |
| Defer `custom` route | Keeps first slice focused and reviewable. | Users cannot create arbitrary documents yet. | Provide clear not-implemented behavior and plan a follow-up change. |
| Answered-only main Markdown body | Clean readable output. | Reviewers might miss skipped sections. | Add `Open / Skipped Sections` appendix. |
| Cobra for commands, stdio for prompts | Good command UX without coupling prompts to UI framework. | Adds one CLI dependency. | Keep framework isolated in `cmd`. |

## Rollout Plan

1. Add Go module and package skeleton.
2. Add command wiring for `apd new` and supported type aliases.
3. Add template schema, loader, bundled YAML templates, and validation tests.
4. Add document model and guided workflow loop.
5. Add project-local session storage with atomic writes.
6. Add Markdown and AI Context Pack generation.
7. Add automated tests and manual smoke notes.
8. Update OpenSpec verification config or project docs with `go test ./...` after the command is available.

## Deferred Decisions

- App-level global session storage paths are specified as a future fallback but not implemented in this MVP.
- Rich prompt libraries, TUI behavior, and advanced multiline editing are deferred.
- User-authored custom templates and `apd new custom` are deferred.
- Resume/edit/export/backlog/prompts workflows are deferred beyond clear placeholder behavior if commands are present.
