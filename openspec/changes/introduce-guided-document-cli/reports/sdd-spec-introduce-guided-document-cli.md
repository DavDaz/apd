# SDD Spec Phase Result: introduce-guided-document-cli

## Status

completed

## Executive Summary

Created the OpenSpec delta spec for `introduce-guided-document-cli`. The spec defines acceptance criteria for the MVP guided document CLI: `apd new`, document type selection, YAML template loading and validation, section guidance, MVP slash commands, incremental session persistence, Markdown export, basic AI Context Pack generation, local/offline operation, and safe handling of roadmap commands.

## Artifacts

- `openspec/changes/introduce-guided-document-cli/specs/guided-document-cli/spec.md`
- `sdd-spec-introduce-guided-document-cli.md`

## Next Recommended

1. Run SDD design for the same change.
2. Resolve design decisions for session storage path, input mechanism, template schema boundaries, Markdown export policy for skipped sections, and whether `custom` is included in the first slice.
3. Run SDD tasks after design.
4. Define verification commands before apply because `openspec/config.yaml` has no test command configured.

## Risks

- Session persistence path is still a design decision.
- Input library vs standard IO is still a design decision.
- `custom` document route is allowed only if feasible; design should decide whether to defer it.
- No test command exists yet; apply must add or document Go verification commands.

## Skill Resolution

parent-direct fallback after `sdd-spec` subagent failed with `fetch failed`. Parent created the OpenSpec spec artifact directly.
