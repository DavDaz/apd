# guided-document-cli Specification

## ADDED Requirements

### Requirement: Start guided document creation

The system SHALL provide an `apd new` command that starts a guided document creation workflow from the terminal.

#### Scenario: Start without type argument

- **GIVEN** the user runs `apd new`
- **WHEN** no document type is provided
- **THEN** the CLI presents the supported document types for selection
- **AND** the supported types include Product Decomposition, Change Request, Feature Spec, Bug / Issue Analysis, Technical Task Pack, and Custom if enabled for the MVP slice

#### Scenario: Start with supported type argument

- **GIVEN** the user runs `apd new product`
- **WHEN** `product` maps to a bundled template
- **THEN** the CLI loads the Product Decomposition template
- **AND** starts the guided section workflow without requiring the selection menu

#### Scenario: Reject unsupported type argument

- **GIVEN** the user runs `apd new unknown`
- **WHEN** `unknown` does not map to a bundled template
- **THEN** the CLI exits with a clear error message
- **AND** lists the supported document type arguments

### Requirement: Load and validate YAML templates

The system SHALL load document templates from YAML files that define template metadata and ordered sections.

#### Scenario: Load valid bundled template

- **GIVEN** a bundled YAML template with `id`, `name`, `description`, and `sections`
- **WHEN** the CLI starts a workflow for that template
- **THEN** the template loader returns a validated template model
- **AND** preserves section order from the YAML file

#### Scenario: Reject template missing required metadata

- **GIVEN** a YAML template missing `id` or `name`
- **WHEN** the template loader validates it
- **THEN** validation fails with a message naming the missing field

#### Scenario: Reject section missing required fields

- **GIVEN** a YAML template containing a section without `id` or `title`
- **WHEN** the template loader validates it
- **THEN** validation fails with a message naming the invalid section and missing field

### Requirement: Guide users through document sections

The system SHALL guide the user through template sections and display contextual information for each section.

#### Scenario: Display section guidance

- **GIVEN** a template section has `title`, `description`, `help`, `example`, and `questions`
- **WHEN** the section is presented
- **THEN** the CLI displays the title and description
- **AND** makes help, example, and guide questions available before or during answer entry

#### Scenario: Capture section answer

- **GIVEN** the user is answering a section
- **WHEN** the user submits text content
- **THEN** the answer is associated with that section
- **AND** the workflow advances to the next section

#### Scenario: Accept long answers

- **GIVEN** the user needs to enter multi-sentence content
- **WHEN** the answer contains line breaks or long text supported by the chosen input mechanism
- **THEN** the CLI preserves the submitted content in the document model and session file

### Requirement: Support guided workflow commands

The system SHALL support the MVP slash commands `/help`, `/skip`, `/back`, and `/done` during document creation.

#### Scenario: Show help without changing answer state

- **GIVEN** the user is on a section
- **WHEN** the user enters `/help`
- **THEN** the CLI displays the section help, example, and guide questions
- **AND** remains on the same section
- **AND** does not overwrite any saved answer for that section

#### Scenario: Skip optional or unwanted section

- **GIVEN** the user is on a section
- **WHEN** the user enters `/skip`
- **THEN** the section is marked as skipped
- **AND** the session is saved
- **AND** the workflow advances to the next section

#### Scenario: Navigate back to previous section

- **GIVEN** the user is not on the first section
- **WHEN** the user enters `/back`
- **THEN** the workflow returns to the previous section
- **AND** allows the user to review or replace the previous answer or skipped state

#### Scenario: Complete early

- **GIVEN** the user is in the middle of a document workflow
- **WHEN** the user enters `/done`
- **THEN** the CLI finalizes the current document with available answers
- **AND** exports Markdown using answered and skipped section state

### Requirement: Save session progress incrementally

The system SHALL persist local session progress after each completed, skipped, or modified section.

#### Scenario: Save after answer

- **GIVEN** the user submits an answer for a section
- **WHEN** the answer is accepted
- **THEN** the session file is updated with the document metadata, selected template, current section state, and answer content

#### Scenario: Save after skip

- **GIVEN** the user skips a section
- **WHEN** the skip is accepted
- **THEN** the session file records the section as skipped
- **AND** preserves any prior answers for other sections

#### Scenario: Session files are local plain files

- **GIVEN** a workflow is in progress
- **WHEN** session data is written
- **THEN** it is stored as a local plain file under a deterministic application/project session directory
- **AND** the file does not require a database, network, or cloud service

### Requirement: Export clean Markdown document

The system SHALL generate a readable Markdown document from the current document model.

#### Scenario: Export answered sections

- **GIVEN** a document contains answered sections
- **WHEN** Markdown export runs
- **THEN** the output includes document metadata and each answered section with its title and answer content
- **AND** section order matches the source template order

#### Scenario: Represent skipped or missing sections safely

- **GIVEN** a document contains skipped or unanswered sections
- **WHEN** Markdown export runs
- **THEN** the output does not invent missing content
- **AND** skipped or unanswered content is omitted or clearly marked as skipped/pending according to the selected export policy

#### Scenario: Write manually editable output

- **GIVEN** Markdown export completes
- **WHEN** the user opens the generated file in a text editor
- **THEN** the file is readable, plain Markdown, and manually editable without tool-specific binary data

### Requirement: Generate basic AI Context Pack

The system SHALL append or generate a basic AI Context Pack derived only from captured document content and explicit metadata.

#### Scenario: Generate AI Context Pack from available content

- **GIVEN** a document has captured answers for context, goals, constraints, entities, rules, flows, criteria, or tasks
- **WHEN** Markdown export runs
- **THEN** the output includes an AI Context Pack section summarizing available captured content under stable headings

#### Scenario: Do not invent absent AI context

- **GIVEN** relevant source sections are skipped or unanswered
- **WHEN** the AI Context Pack is generated
- **THEN** the generator marks missing information as pending or omits it
- **AND** does not fabricate rules, entities, tasks, or acceptance criteria

### Requirement: Remain local-first and offline-capable

The system SHALL complete the MVP workflow without mandatory network access, AI provider credentials, database services, or cloud synchronization.

#### Scenario: Run without network access

- **GIVEN** the machine has no network connectivity
- **WHEN** the user runs the guided workflow with bundled templates
- **THEN** document creation, session saving, and Markdown export still work locally

#### Scenario: Avoid mandatory AI credentials

- **GIVEN** no AI provider API key is configured
- **WHEN** the user runs `apd new`
- **THEN** the CLI does not require an AI key
- **AND** still generates the basic AI Context Pack from captured content

### Requirement: Keep roadmap commands out of MVP behavior

The system SHALL not present incomplete roadmap features as working MVP functionality.

#### Scenario: Roadmap command is not implemented

- **GIVEN** a command such as `apd backlog`, `apd prompts`, `apd edit`, `apd validate`, or `apd template create` is not implemented in this change
- **WHEN** the user invokes it, if the command exists as a placeholder
- **THEN** the CLI clearly states that the feature is not implemented yet
- **AND** does not create partial or misleading output
