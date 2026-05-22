status: complete

executive_summary: |
  The CLI is already close to globally installable: templates are embedded in the binary, runtime output is rooted at the process working directory, and smoke testing a built binary from a new temp folder produced `.apd/sessions/*` and `apd-docs/*` in that folder. The main gaps are product/UX and distribution: bare `apd` currently prints help instead of launching the guided flow, README only documents `go run . new`, there are no `cmd` package tests for root-command behavior, and `go.mod` uses the module path `apd`, which blocks normal remote `go install github.com/...@latest` style installation unless the module path is changed before publishing.

artifacts:
  output_file: `/Users/dadiaz/Documents/git/github/apd/explore-global-install-apd.md`
  memory: Engram save was requested, but no Engram/memory tools are available in this subagent runtime, so no memory observation was saved.
  files_retrieved:
    - `main.go` (lines 1-15) - process entrypoint delegates args/stdout to `cmd.Execute` and errors to stderr.
    - `cmd/root.go` (lines 8-18) - root command behavior; bare `apd` currently prints help.
    - `cmd/new.go` (lines 12-21) - `new` command loads bundled templates, resolves `os.Getwd`, and passes it as workflow `WorkingDir`.
    - `internal/app/new_document.go` (lines 27-137) - guided flow orchestration, default store/exporter injection, menu flow, session saves, final path printing.
    - `internal/templates/registry.go` (lines 10-58) - embedded default templates via `//go:embed bundled/*.yaml`.
    - `internal/templates/registry_test.go` (lines 36-61) - tests prove bundled templates load after `chdir` to an unrelated temp directory.
    - `internal/storage/paths.go` (lines 10-16) - session paths are rooted at `WorkingDir/.apd/sessions`.
    - `internal/storage/filesystem.go` (lines 12-49) - atomic session writes and working-directory remediation error.
    - `internal/generator/markdown.go` (lines 14-52, 111-113) - Markdown output paths are rooted at `WorkingDir/apd-docs` with remediation error.
    - `README.md` (lines 7-29, 53-60, 103-117) - current docs describe `go run . new`, outputs, and dev checks but not global install.
    - `go.mod` (lines 1-5) - module path is currently `apd`, not a canonical import path.
    - `.gitignore` (lines 1-8) - generated `.apd/` and `apd-docs/` are ignored locally.

findings:
  - Bare `apd` does not start the guided flow. `cmd.Execute` treats `len(args) == 0` the same as `--help` and prints usage (`cmd/root.go:9-12`). The future feature says users should run `apd` and get a working guided flow, so this behavior needs to change or the product requirement should be narrowed to `apd new`.
  - `apd new` is already cwd-safe for global use. `cmd/runNew` calls `os.Getwd()` and passes that directory to `app.RunNewDocument` (`cmd/new.go:17-21`). The default store/exporter are then rooted at `cfg.WorkingDir` (`internal/app/new_document.go:56-60`).
  - Templates are binary-embedded, not loaded relative to the caller's cwd. `LoadDefaultRegistry` loads `bundledTemplates` from `internal/templates/bundled/*.yaml` (`internal/templates/registry.go:10-58`). This is the most important existing support for global install.
  - Existing tests explicitly cover cwd-independent template loading by changing into a temp directory before `LoadDefaultRegistry()` (`internal/templates/registry_test.go:36-61`).
  - Runtime artifacts are already local to the folder where the user runs the command: sessions go to `.apd/sessions/<session>.session.yaml` (`internal/storage/paths.go:10-16`), Markdown goes to `apd-docs/<session>.md` (`internal/generator/markdown.go:46-52`).
  - Smoke verification: after `go build -o <tmp>/apd .`, running `printf '1\n/done\n' | <tmp>/apd new` from a fresh temp directory generated both `./.apd/sessions/*.session.yaml` and `./apd-docs/*.md` in that temp directory.
  - Distribution is not yet documented. README quick start is source-tree oriented: `go run . new` and `go run . new product` (`README.md:7-29`).
  - Remote global install is likely blocked until the module path is changed from `module apd` (`go.mod:1`) to the eventual repository import path, e.g. `github.com/<owner>/apd`, because `go install github.com/<owner>/apd@latest` expects the module declaration to match the requested module path.
  - There are no tests in `cmd/` for root command behavior (`cmd` currently reports `[no test files]` under `go test ./...`), so changing bare `apd` should add command-level tests.
  - There is a duplicate top-level `templates/` directory with YAML matching `internal/templates/bundled/`. Runtime uses only the embedded internal templates. This duplication may confuse packaging/docs unless intentionally kept as human-readable source/template reference.

recommended_changes:
  - Make bare `apd` start the default guided flow by delegating to `runNew(nil, out)` when `len(args) == 0`. Keep `apd --help` and `apd -h` as help-only.
  - Update root help text to show both happy paths:
    - `apd` starts guided mode.
    - `apd new [type]` starts guided mode directly or with type selection.
  - Add `cmd` package tests before changing behavior:
    - `Execute(nil, out)` starts/attempts the guided flow rather than printing help. Because `runNew` currently binds `os.Stdin`, testability may require introducing a small root config or factoring command execution to accept input/working-dir dependencies.
    - `Execute([]string{"--help"}, out)` still prints usage and does not create files.
    - `Execute([]string{"new", "product"}, out)` remains supported.
  - Consider dependency injection for `cmd.runNew` input and cwd so command tests can avoid real stdin/cwd. The app layer is already testable through `NewDocumentConfig`; the `cmd` layer is the remaining hard-to-test boundary.
  - Decide distribution target before implementation:
    - If install is from a cloned checkout only, document `go install .` from repo root.
    - If install is public/global from GitHub, change `go.mod` module path to the canonical repo path and update imports from `apd/...` to that path.
  - Keep embedded templates as the production source for global install. If top-level `templates/` remains, document it as source/reference or add a check that it stays synchronized with `internal/templates/bundled/`.
  - Keep all writes rooted at `WorkingDir`; do not introduce config/state under `$HOME` unless explicitly required later.

readme_changes:
  - Replace Quick Start step 1 with install-first flow:
    ```bash
    go install github.com/<owner>/apd@latest
    cd /path/to/any/project-or-new-folder
    apd
    ```
  - Add a fallback for local development:
    ```bash
    go run . new
    # or after building locally:
    go install .
    apd
    ```
  - State that templates are bundled into the binary; users do not need to copy template files into each project.
  - State that all generated artifacts are written to the current working directory:
    - `.apd/sessions/` for recoverable YAML session state.
    - `apd-docs/` for generated Markdown.
  - Update the flow diagram first node from `Run apd new` to `Run apd` or `apd new`.
  - Add troubleshooting notes:
    - Ensure `$GOBIN` or `$GOPATH/bin` is on `PATH` after `go install`.
    - Run `pwd` before `apd` if artifacts appear in the wrong folder.
    - If write errors occur, run from a writable project/new directory.

acceptance_criteria:
  - A user can install the CLI once with the documented global install command.
  - From any writable existing project directory, running `apd` starts the guided document type selection without requiring source checkout files.
  - From a newly-created empty directory, running `apd` starts the same flow and writes `.apd/sessions/*.session.yaml` plus `apd-docs/*.md` in that directory.
  - `apd --help` remains non-interactive and prints usage.
  - `apd new` and `apd new <type>` remain supported for explicit command usage.
  - Bundled templates load successfully regardless of process cwd.
  - README documents install, PATH expectations, cwd artifact behavior, direct type usage, and development commands.

suggested_tests:
  - Unit: add `cmd` tests for root behavior, help behavior, unknown commands, and `new` dispatch. Prefer factoring command execution to inject `io.Reader` and working directory rather than depending on global `os.Stdin`/`os.Getwd`.
  - Unit: keep/extend `internal/templates.TestLoadDefaultRegistry` because it already verifies templates load after changing cwd to a temp directory.
  - Integration smoke command:
    ```bash
    go build -o /tmp/apd-smoke ./
    tmp=$(mktemp -d)
    (cd "$tmp" && printf '1\n/done\n' | /tmp/apd-smoke)
    test -d "$tmp/.apd/sessions"
    test -d "$tmp/apd-docs"
    find "$tmp/.apd/sessions" -name '*.session.yaml' | grep .
    find "$tmp/apd-docs" -name '*.md' | grep .
    ```
  - Root-command integration after behavior change:
    ```bash
    tmp=$(mktemp -d)
    (cd "$tmp" && printf '1\n/done\n' | /tmp/apd-smoke)
    ```
    This should work without the `new` subcommand.
  - Help smoke:
    ```bash
    /tmp/apd-smoke --help
    ```
    Confirm it exits successfully, prints usage, and does not create `.apd` or `apd-docs` in cwd.
  - Standard checks:
    ```bash
    go test ./...
    go vet ./...
    ```
  - Optional install verification once module path is canonical:
    ```bash
    GOBIN=$(mktemp -d) go install github.com/<owner>/apd@<version-or-commit>
    tmp=$(mktemp -d)
    (cd "$tmp" && printf '1\n/done\n' | "$GOBIN/apd")
    ```

risks:
  - Changing bare `apd` from help to interactive mode is a UX-breaking change for anyone expecting no-arg help. Mitigate by keeping `--help` reliable and making docs explicit.
  - `cmd.runNew` currently depends on `os.Stdin` and `os.Getwd`, which makes root-command behavior harder to unit test cleanly.
  - The module path change required for public `go install github.com/...@latest` will touch imports across the codebase and should be done deliberately.
  - Duplicate template directories can drift. Since global install depends on embedded templates, unsynchronized top-level templates may cause confusing docs or tests later.
  - Generated artifacts use absolute paths in final output. This is clear, but README examples may prefer explaining that paths point inside the current directory.
  - No resume flow was inspected/implemented; sessions are saved but the current feature request only asks for initial guided flow/artifact placement.

skill_resolution: paths-injected
