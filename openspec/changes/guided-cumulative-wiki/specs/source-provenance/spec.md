# Source Provenance Specification

## Purpose
Byte-identity receipts for local source registration.

## Requirements

### Requirement: Immutable, Idempotent Source Registration
The system MUST accept only readable regular files within the selected source boundary, retain bytes unchanged, and calculate SHA-256 over exact bytes. An immutable receipt MUST contain source ID, canonical source/managed paths, SHA-256, byte length, time, declared type, and notes. A digest asserts byte identity only. Identical bytes at the same canonical path MUST return the existing identity; changed bytes MUST create a distinct identity without overwriting prior bytes or receipt.

#### Scenario: Register or re-register identical bytes
- GIVEN a permitted file with unchanged bytes
- WHEN it is registered twice
- THEN required receipt fields persist once and both calls return its identity

#### Scenario: Reject or detect a change
- GIVEN an unsafe source or a registered path with different bytes
- WHEN registration is requested
- THEN unsafe input changes nothing; different bytes preserve the earlier receipt and create a new identity
