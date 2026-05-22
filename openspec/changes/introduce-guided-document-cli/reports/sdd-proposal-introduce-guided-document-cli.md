# SDD Proposal Phase Result: introduce-guided-document-cli

## Status

completed

## Executive Summary

Created the OpenSpec proposal for `introduce-guided-document-cli` from `PRD_AI_Product_Decomposer_CLI.md`.

The proposal defines the MVP as a local-first Go CLI guided-document workflow centered on `apd new`, YAML templates, guided prompts, slash commands, incremental local session saving, Markdown export, and a basic AI Context Pack. It explicitly excludes AI/model calls, web/cloud/database features, collaboration, advanced TUI, custom template authoring UI, full backlog generation, and full prompt pack generation.

## Artifacts

- `openspec/changes/introduce-guided-document-cli/proposal.md`
- `sdd-proposal-introduce-guided-document-cli.md`

## Next Recommended

1. Create the OpenSpec spec artifact for the same change.
2. Define acceptance criteria and data contracts for templates, documents, sessions, and Markdown output.
3. Create design and tasks artifacts before implementation.
4. Before apply, decide delivery slicing and verification plan because no test command is currently configured.

## Risks

- Interactive CLI scope can expand quickly; keep MVP limited to `/help`, `/skip`, `/back`, and `/done`.
- Template schema should be stabilized before implementation to avoid rework.
- Session persistence paths need an explicit cross-platform decision.
- Markdown generation must not invent content beyond user input.
- Verification needs to be defined once Go project setup exists because `openspec/config.yaml` has no configured test command.

## Skill Resolution

paths-injected

Loaded the parent-injected skills:

- `/Users/dadiaz/.config/opencode/skills/cognitive-doc-design/SKILL.md`
- `/Users/dadiaz/.agents/skills/create-specification/SKILL.md`
- `/Users/dadiaz/.agents/skills/golang-pro/SKILL.md`

## Engram Persistence

Engram tools were not available in this agent toolset, so the proposal was persisted to OpenSpec files only.
