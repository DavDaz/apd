# External Agent Integration Specification

## Purpose
Deterministic, non-semantic external-agent handoff.

## Requirements

### Requirement: Complete Deterministic Integration Request
For a `request-ready` item, the system MUST emit deterministic machine-readable content with request/workspace versions, source ID, complete receipt, immutable source location, expected wiki targets, output/receipt location, contradiction policy, and `awaiting-external-semantic-integration`. The policy MUST require candidate conflicting wiki paths and prohibit APD resolution. Equivalent input MUST produce equivalent content. Emission MUST NOT alter wiki content or assert integration; incomplete input MUST fail without output or lifecycle advance.

#### Scenario: Emit a valid request
- GIVEN a complete `request-ready` item
- WHEN integration material is requested
- THEN one complete request is emitted and state becomes awaiting external semantic integration

#### Scenario: Reject incomplete material
- GIVEN a missing receipt field, source access, or target instruction
- WHEN emission is requested
- THEN no request is emitted and state is unchanged
