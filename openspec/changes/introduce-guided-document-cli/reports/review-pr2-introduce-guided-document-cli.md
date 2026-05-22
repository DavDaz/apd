## Review

Verdict: ready-for-pr3

Findings:
- minor: `internal/cli/menu.go:24-25` reuses `Input.ReadIntent()` for document-type menu selection. Because `ReadIntent()` only returns an answer after EOF or an empty line (`internal/cli/input.go:30-39`), an interactive `apd new` user choosing from the menu must press an extra blank line even though the prompt says only `Choose a document type:`. This does not break the explicit `apd new product` path or tests, but it is a PR2 UX/parser edge to fix or clarify before relying on no-arg interactive selection.

Correct:
- Domain state transitions are coherent: new documents initialize sections as `pending`, `AnswerCurrent` replaces the active answer and marks `answered`, `SkipCurrent` clears prior answer and marks `skipped`, and `Back` only moves the cursor without mutating section state (`internal/document/document.go:24-95`, `internal/document/section.go:19-36`). Tests cover answer, skip, replacement after back, and first-section back no-op (`internal/document/document_test.go`).
- Guided parser behavior matches the PR2 contract for section input: commands are recognized only before an answer buffer has content (`internal/cli/input.go:21-33`), multi-line answers terminate on an empty line, and command-like text after the first answer line remains answer content (`internal/cli/input_test.go:31-50`).
- App orchestration stays inside PR2 scope: it resolves/presents document type, initializes the document, renders sections, handles `/help`, `/back`, `/skip`, `/done`, and answers, and saves only after accepted answers/skips/replacements (`internal/app/new_document.go:52-107`). Markdown/final export is not implemented; output explicitly defers it to PR3 (`internal/app/new_document.go:94-110`).
- Session persistence uses `.apd/sessions` via `filepath`, schema version `1`, YAML session fields for metadata/template/current index/section states, temp-file write plus rename, and clear write-error guidance (`internal/storage/paths.go`, `internal/storage/session.go:9-50`, `internal/storage/filesystem.go:18-48`). Tests cover path construction, YAML round trip, replacement, and write-error messaging (`internal/storage/filesystem_test.go:15-64`).
- Generated runtime state is ignored: `.gitignore:1-2` excludes `.apd/`.
- No PR3 scope creep found in source packages: no `internal/generator` package or Markdown/AI Context Pack implementation is present; PR3 mentions are limited to explicit deferral messages/tests.

Verification run:
- `go test ./...` passed.
