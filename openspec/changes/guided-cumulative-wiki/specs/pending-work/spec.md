# Pending Work Specification

## Purpose
Resumable, unambiguous source handoff state.

## Requirements

### Requirement: Atomic Resumable Lifecycle
The system MUST atomically persist each receipt-linked item as `registered`, `request-ready`, `awaiting-external-semantic-integration`, or `failed`. Reopening MUST recover unfinished work and show its deterministic next action. This slice MUST NOT set an item complete or integrated. Status MUST use `awaiting-external-semantic-integration`; emitted, accepted, or downloaded requests MUST NOT be called integrated or complete.

#### Scenario: Resume pending work
- GIVEN interruption after a `registered` item was persisted
- WHEN the workspace reopens
- THEN it remains registered and requests preparation as the next action

#### Scenario: Fail safely or hand off
- GIVEN persistence fails or a request was emitted
- WHEN status is reopened
- THEN the old state remains recoverable or status says external semantic integration is pending
