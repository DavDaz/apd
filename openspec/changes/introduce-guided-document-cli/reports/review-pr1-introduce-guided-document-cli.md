## Review

**Verdict:** needs-fix

### Findings

- **major:** Default template loading is cwd-dependent, so the templates are not actually bundled with the CLI binary. `cmd/new.go:19` calls `templates.LoadDefaultRegistry()`, and `internal/templates/registry.go:47-50` loads `templates/` from `os.DirFS(".")`. Running an installed `apd` outside the repo/template directory will fail to find bundled templates, which conflicts with the PR 1 boundary/intent of bundled templates and offline local use. Prefer embedding the templates with `//go:embed` or otherwise loading from a deterministic packaged location.

- **blocker (release verification):** Go is not available in this environment, so compilation/tests could not be executed. Evidence: `go test ./...` failed with `/bin/bash: go: command not found` / exit 127. This is not a code defect, but release should be blocked until `go test ./...` passes in an environment with Go installed.

### Correct / in scope

- CLI skeleton is thin and stays inside PR 1 scope: `cmd/root.go:8-17` only routes root/new/help, and `cmd/new.go:23-31` lists supported types or confirms a template load without implementing workflow, storage, or Markdown generation.
- Template decoding rejects unknown YAML fields via `yaml.Decoder.KnownFields(true)` in `internal/templates/loader.go:15-22`, with test coverage in `internal/templates/validate_test.go:33-40`.
- Required template fields and section fields are validated in `internal/templates/validate.go:9-36`, including non-empty sections and duplicate section ids.
- Registry validation catches duplicate template ids and alias conflicts using normalized keys in `internal/templates/registry.go:16-43`, with tests in `internal/templates/registry_test.go:9-51`.
- Section order is preserved by YAML slice order and covered in `internal/templates/validate_test.go:37-40`.
- The custom route is disabled: bundled templates omit `custom`, and `internal/templates/registry_test.go:48-50` asserts `Resolve("custom")` is false.
- Bundled template YAML files are valid against the implemented schema and concise: `templates/{bug,change-request,feature,product,task}.yaml` each define id/name/description/version/aliases and one required section.
- No PR 2/PR 3 scope creep observed: no storage/session packages, guided workflow implementation, Markdown generator, or AI context generator code is present.

### Review workload / changed-files risk

- Source/template slice is moderate: 15 implementation/test/template files, about 424 lines excluding SDD docs. Main risk is packaging/runtime behavior of templates rather than review size.
