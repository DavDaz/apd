# Legacy Document Authoring Specification

## Purpose
Preserve current document-authoring behavior.

## Requirements

### Requirement: `apd new` and Explicit-Only Continuation
The system MUST preserve `apd new [type]` validation, template selection, interactive/TUI behavior, session persistence, Markdown filenames, and output contract. It MUST NOT require a workspace, rewrite `apd-docs/` or sessions, or alter templates. Legacy continuation MUST require explicit user choice; files MUST NOT migrate or convert automatically.

#### Scenario: Create or reject a legacy document
- GIVEN a directory without a wiki workspace
- WHEN a supported or invalid `apd new` request runs
- THEN supported input has unchanged legacy outputs; invalid input creates no wiki artifact

#### Scenario: Open legacy data by default
- GIVEN legacy APD files exist without explicit continuation
- WHEN the default guided entry opens
- THEN those files remain unchanged and no conversion is reported
