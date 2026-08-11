# Wiki Workspace Specification

## Purpose
First-slice local workspace initialization and guided entry.

## Requirements

### Requirement: Safe Initialization and Guided Entry
The system MUST create versioned `raw/`, `wiki/`, and `.apd/` only at an absent child of a selected existing directory. It MUST reject collisions, non-directory parents, symlink traversal, and writes outside that parent. Initialization MUST be atomic. Default interactive entry MUST show persisted next action and status without command recall. It MUST NOT implicitly register, semantically integrate, query, ingest remotely, sync, resolve contradictions, lint, automate index/logs, convert legacy data, or provide web UI.

#### Scenario: Initialize and guide
- GIVEN a writable parent and absent child target
- WHEN initialization succeeds and the default entry opens
- THEN all managed directories, a supported manifest, and the next safe action are shown

#### Scenario: Reject unsafe initialization
- GIVEN an existing, escaping, or symlink-traversing target
- WHEN initialization is requested
- THEN it fails without changing any path
