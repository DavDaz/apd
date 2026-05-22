# Proposal: Introduce Guided Document CLI

## Intent

Introduce the first MVP slice of `apd`: a local, Go-based command-line assistant that guides users through structured product and technical documentation flows, saves progress incrementally, and exports clean Markdown with a basic AI Context Pack.

The change turns `apd` into an AI-oriented product decomposition tool rather than a generic PRD generator. Its primary value is the guided thinking flow; Markdown output is the initial delivery format.

## Problem Statement

Users who start or change software projects often move from vague ideas directly into improvised code, or they rely on heavy PRD templates that are difficult to complete and not optimized for AI-assisted development. They need a lightweight terminal workflow that asks contextual questions, supports optional sections, and produces useful local documents for human review and AI coding tools without requiring an AI model, web service, database, or cloud account.

## Scope

### In Scope for MVP

- Add the `apd new` workflow.
- Let users select a document type:
  - Product Decomposition
  - Change Request
  - Feature Spec
  - Bug / Issue Analysis
  - Technical Task Pack
  - Custom / empty document route, if feasible within the first slice.
- Load predefined document templates from YAML files.
- Guide the user section by section with:
  - title
  - description
  - contextual help
  - example
  - guide questions
  - optional answer capture
- Support interactive commands during document creation:
  - `/help`
  - `/skip`
  - `/back`
  - `/done`
- Save session progress incrementally to local files so interrupted work can be recovered by later phases.
- Export a clean Markdown document from captured answers.
- Generate a basic AI Context Pack section from the collected document content.
- Keep all output files manually editable.
- Keep implementation local-first and offline-capable.

### Explicitly Out of Scope for MVP

- AI/model calls or prompt execution.
- Web server or API service.
- Cloud synchronization.
- Database-backed persistence.
- Collaboration or multi-user features.
- Advanced terminal UI/TUI.
- Custom template authoring UI.
- Full backlog generation.
- Full prompt pack generation.
- JSON/YAML export beyond what is needed for internal session/template persistence.
- `apd resume`, `apd edit`, `apd export`, `apd backlog`, `apd prompts`, `apd validate`, and `apd template create` as complete user-facing commands, except where lightweight placeholders are useful for command discovery.

## Affected Areas

| Area | Expected Impact |
| --- | --- |
| CLI entrypoint | Introduce root command behavior and `apd new` command flow. |
| Template system | Add YAML template schema, loader, validation, and bundled default templates. |
| Guided prompt engine | Add section traversal, command handling, answer capture, and navigation state. |
| Document model | Define document, section, answer, metadata, and skipped-section concepts. |
| Local storage | Add incremental session persistence under a local project/application directory. |
| Markdown generation | Render answered and skipped sections into readable Markdown with metadata and AI Context Pack. |
| Project structure | Establish idiomatic Go package boundaries for CLI, templates, document model, storage, and generators. |
| Tests and verification | Add manual and automated verification around template loading, command behavior, session saving, and Markdown output once implementation begins. |

## Proposed Direction

Implement the MVP as a modular Go CLI with small internal packages:

- `cmd`: command wiring and user-facing CLI commands.
- `internal/cli`: prompt rendering, menu selection, and command parsing.
- `internal/templates`: YAML schema, loader, registry, and validation.
- `internal/document`: document, section, answer, and metadata models.
- `internal/storage`: filesystem session persistence.
- `internal/generator`: Markdown and AI Context Pack generation.

The first implementation should prioritize clear interfaces over advanced UI. Dependencies should remain minimal. The tool must work offline and store all generated artifacts as plain files.

## Success Criteria

The change is successful when:

- A user can run `apd new` from the terminal.
- The CLI presents supported document types or accepts a supported type argument.
- The selected YAML template is loaded and validated.
- The CLI displays contextual guidance for each section.
- The user can enter multi-line or sufficiently long answers.
- `/help` shows section guidance without losing the current answer state.
- `/skip` marks a section as skipped and continues.
- `/back` returns to the previous section and allows changing the answer.
- `/done` completes the session early and exports current content.
- Progress is saved incrementally after each completed or skipped section.
- A Markdown file is generated with a readable structure and a basic AI Context Pack.
- The workflow does not require network access, an AI provider, a database, or a web server.

## Risks

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Interactive CLI behavior becomes too complex for the first slice. | Larger review burden and delayed MVP. | Keep commands limited to `/help`, `/skip`, `/back`, and `/done`; defer advanced TUI and editing. |
| Template schema changes after implementation starts. | Rework across loader, prompts, and generator. | Define the schema in the spec before apply and add validation tests. |
| Session persistence path is ambiguous across operating systems. | Lost or hard-to-find progress files. | Specify deterministic local paths and document them before implementation. |
| Markdown generation invents content beyond user input. | Reduces trust and AI-readiness. | Render only captured content and clearly mark missing information as skipped or pending. |
| Scope creep into backlog, prompt generation, or custom template authoring. | MVP exceeds review budget. | Treat backlog/prompts/custom authoring as later changes; only basic AI Context Pack is included. |
| No test command is configured for the project yet. | Verification may be inconsistent. | Define a verification plan before apply and add Go test commands once `go.mod` exists. |

## Rollback Plan

- Revert files introduced for the CLI command, template loader, prompt engine, storage, and generator.
- Remove bundled templates and generated documentation examples from the change branch.
- Delete any new OpenSpec artifacts for this change if the proposal is abandoned.
- Because the MVP is local-only and does not introduce external services, database migrations, or cloud state, rollback should be limited to source and documentation changes.

## Open Questions for Later Phases

- What exact filesystem path should be used for session files across macOS, Linux, and Windows?
- Should `apd new custom` be included in the first implementation slice or deferred after predefined templates are stable?
- Should the initial CLI use only standard input/output, or introduce a prompt library immediately?
- What exact Markdown filename and output directory conventions should be used for each document type?
- Should placeholder commands be added for roadmap discoverability, or should unsupported commands be omitted until implemented?

## References

- Source PRD: `PRD_AI_Product_Decomposer_CLI.md`
- Change name: `introduce-guided-document-cli`
- OpenSpec config: `openspec/config.yaml`
