## Review
- Correct: `internal/cli/menu.go:21-39` now uses `scanner.Scan()` directly for menu selection, so numeric/id choices are accepted after one newline instead of requiring multiline `ReadIntent` termination.
- Correct: guided section input still uses `input.ReadIntent()` in `internal/app/new_document.go:74`, preserving multiline answer behavior.
- Correct: regression coverage exists in `internal/cli/input_test.go:54-72` for numeric and id choices without an extra blank line.
- Correct: verification passed:
  - `go test ./...`
  - `go vet ./...`
  - `git diff --check`
  - `printf '1\n/done\n' | go run . new`

- Note: I did not write `/Users/dadiaz/Documents/git/github/apd/review-pr2-after-menu-fix.md` because the task also said “Do not edit files,” and no-edit wins for review-only instructions.

Verdict: **ready-for-pr3**.