## Exploration: guided-cumulative-wiki

### Current State

APD is currently a local-first Go document authoring tool, not a cumulative wiki. `apd new [type]` loads one embedded YAML template, guides the user through ordered sections, saves the mutable session to `.apd/sessions/<id>.session.yaml`, and emits one editable Markdown file under `apd-docs/<id>.md`. The generated Markdown contains answered sections, an `AI Context Pack`, and an open/skipped appendix. It never reads prior documents, source files, or an existing knowledge base.

The application boundary is already reasonably small: `cmd` exposes only `new`; `internal/app` owns the shared guided workflow; `internal/document` owns section state; `internal/templates` loads strict YAML templates; `internal/storage` performs atomic YAML saves; `internal/generator` renders Markdown; and `internal/cli/tui` provides the current Bubble Tea adapter. The TUI has selection, sequential authoring, help, revisit, review, confirmation, and partial-finish behavior. The CLI remains a pipe-friendly fallback. `go test ./...` and `go vet ./...` pass.

Important limitations for the new direction are structural rather than cosmetic:

- There is no session load/resume path despite persisted YAML being described as recoverable.
- A session is a work-in-progress form, not a durable source receipt or wiki record.
- Markdown is a one-off document with no stable page identity, provenance model, links, index, log, claim status, or contradiction representation.
- Templates describe prompts and Context Pack mappings only; they do not describe ingest phases, source metadata, wiki page schemas, or maintenance rules.
- The only command is `new`; the UI cannot discover or lead the user through ingest, integration review, maintenance, or external-agent handoff.
- `BuildContextPack` is a fixed heading mapper and cannot express cumulative synthesis or citations.
- The TUI currently allows explicit partial export. That is compatible with one-off drafts but dangerous if interpreted as a trusted wiki update.
- No `openspec/specs/` main specifications or active `guided-cumulative-wiki` artifacts exist yet. The existing `introduce-guided-document-cli` change documents the one-off design and explicitly deferred resume, edit, export, validation, AI calls, and richer TUI workflows.

### Affected Areas

- `cmd/root.go`, `cmd/new.go` — command discovery must evolve from one authoring command into a small, guided workspace entry point without forcing users to memorize subcommands.
- `internal/app/new_document.go`, `internal/app/guided_workflow.go` — retain the existing authoring use case, then add a workspace/wiki workflow that can stage ingest work and route completion to review or handoff.
- `internal/cli/tui/model.go`, `internal/cli/tui/navigation.go` — the TUI is the main product surface; it needs a workflow dashboard and explicit next-action guidance, not merely better section navigation.
- `internal/templates/schema.go`, `internal/templates/validate.go`, `internal/templates/bundled/*.yaml` — extend templates only where they define deterministic intake prompts and workflow metadata; do not turn YAML into an LLM reasoning language.
- `internal/storage/paths.go`, `internal/storage/filesystem.go`, `internal/storage/session.go` — separate ephemeral authoring sessions from durable wiki workspace artifacts and add safe load/upgrade behavior for existing YAML.
- `internal/generator/markdown.go`, `internal/generator/ai_context.go` — preserve one-off rendering while adding deterministic source receipts, wiki page frontmatter/links, `index.md`, `log.md`, and external-agent work-request output.
- `internal/document/*` — introduce durable identities and workflow states only if they are shared by the adapters; avoid coupling the domain model to Markdown or an LLM provider.
- `templates/*.yaml`, `README.md`, `PRD_AI_Product_Decomposer_CLI.md` — document the new product identity and preserve the legacy document routes as compatibility behavior.
- `openspec/changes/introduce-guided-document-cli/*` — historical source of the current contract; it should remain an audit trail, while the new change explicitly supersedes the product framing rather than rewriting history.

### Karpathy Pattern Fit

Karpathy's primary gist defines three layers: immutable `raw` sources, an agent-maintained Markdown `wiki`, and a schema/instructions document. Its core operations are ingest, query, and lint. `index.md` is a content-oriented catalog and `log.md` is an append-only chronological record. New sources are integrated into multiple existing pages, with cross-references and contradictions recorded instead of silently erased. Query/recall is intentionally an agent responsibility, and the gist treats the filesystem/Git plus readable Markdown as the durable substrate.

APD currently implements only a deterministic subset of intake: guided capture and Markdown export. It should become the local workflow/orchestration layer around the pattern, not a second conversational agent. The smallest coherent APD-owned workspace is:

```text
<project>/
├── raw/                         # immutable user-curated sources
├── wiki/
│   ├── index.md                 # generated catalog of pages
│   ├── log.md                   # append-only workflow/event log
│   ├── schema.md                # human/agent-maintained operating contract
│   ├── sources/                 # source receipt pages or source metadata
│   └── pages/                   # durable, externally readable Markdown pages
└── .apd/
    ├── sessions/                # resumable guided UI state
    └── work/                    # deterministic ingest/handoff manifests
```

`raw/` and `wiki/` are intentionally ordinary files. APD should atomically create/update only artifacts it owns (receipts, index, log, manifests), while an external agent owns semantic wiki integration unless the user explicitly chooses a future local provider.

### Approaches

1. **Add a cumulative workspace and guided ingest/handoff around the existing engine (recommended)** — preserve `apd new` as legacy authoring, add a discoverable dashboard route that initializes/opens `raw/` + `wiki/`, captures source metadata and integration intent, validates deterministic invariants, and emits a structured work request for an external agent.
   - Pros: smallest coherent migration; keeps APD offline and deterministic; matches Karpathy's file-first pattern; makes provenance and human review explicit; external agents can read plain Markdown/YAML; existing documents remain usable.
   - Cons: semantic page integration is not performed by APD itself; the first version needs a clear handoff contract and safe partial/error states.
   - Effort: Medium

2. **Embed an LLM provider and make APD perform integration locally** — APD reads raw sources, calls a configured model, updates pages, resolves links, and writes index/log entries.
   - Pros: one apparent application flow and less manual handoff.
   - Cons: nondeterministic writes, provider/configuration burden, difficult rollback and testing, model-specific behavior, privacy/cost concerns, and a misleading boundary between capture and knowledge authority.
   - Effort: High

3. **Replace APD with a full wiki/query platform** — add search/RAG, conversational recall, graph UI, editors, background watchers, and remote synchronization.
   - Pros: broad end-to-end feature set.
   - Cons: directly violates the approved boundary that recall is external; creates a database/search/UI product before the durable ingest contract is proven; high review and maintenance cost.
   - Effort: High

### Recommendation

Choose Approach 1. Reframe APD as a guided, local-first wiki intake and maintenance coordinator. The UI should start with “What do you want to do next?” and make the happy path visible: initialize/open workspace → add or select an immutable source → capture source receipt and emphasis → validate → generate an external-agent integration request → review the resulting changes → run deterministic health checks. Users should not need to know whether the underlying action is called `ingest`, `lint`, `resume`, or `handoff`; commands may exist for automation, but the TUI must expose them as actions and show the next safe action.

The first implementation slice should not attempt semantic synthesis. It should provide:

1. A versioned workspace manifest and deterministic directory initialization.
2. A guided source intake flow that never edits `raw/`, assigns a stable source ID, records path/hash/timestamps, and stores user emphasis/notes.
3. A structured, human-readable external-agent work request containing source receipt, current wiki/index context, required provenance, expected page/index/log updates, and contradiction policy.
4. Deterministic validation for file layout, required frontmatter/links where APD owns the contract, index coverage, log syntax, and unresolved integration status.
5. A review/maintenance dashboard that can resume pending intake work and explain exactly what the user should do next.
6. Compatibility export for existing one-off documents, with an explicit “draft/work request” status rather than treating them as integrated wiki knowledge.

The LLM is therefore **not required for APD's ingest capture**. It is required for semantic integration unless a human or external deterministic process performs that work. APD should generate a structured work request, not claim that a Context Pack is a cumulative wiki update. An external agent can read the request, raw source, schema, index, and relevant pages; it can then create a proposed change set. APD's later role is to validate receipts/invariants and record the outcome, not silently resolve disagreements.

### Product Boundaries

**APD should implement now:**

- Guided UI-driven ingest and maintenance navigation.
- Immutable-source registration and provenance receipts (stable ID, relative path, content hash, captured time, source type, optional human label).
- Durable workspace initialization and resumable local sessions.
- Markdown/YAML contracts readable by external agents and humans.
- `index.md` generation from known page metadata and `log.md` append-only event recording.
- Integration work-request generation and explicit statuses such as `captured`, `ready-for-agent`, `proposed`, `accepted`, `rejected`, and `needs-review`.
- Contradiction/conflict recording as unresolved references between grounded claims or source/page records; never automatic winner selection.
- Deterministic lint checks for missing provenance, orphan/index gaps, broken links, malformed metadata, duplicate IDs, stale pending work, and unrecorded integration outcomes.
- Legacy `apd new` behavior and conversion/export paths that do not corrupt old sessions or documents.

**APD should not implement yet:**

- Conversational query/recall, embeddings, RAG, vector databases, or a search engine beyond simple local validation/navigation.
- An embedded LLM provider, automatic semantic page editing, autonomous contradiction resolution, or confidence-based invention.
- A web UI, multi-user permissions, cloud sync, background daemon, graph visualization, or Obsidian replacement.
- Automatic ingestion of arbitrary URLs/PDFs/Slack/email without an explicit immutable-source boundary and user confirmation.
- Backlog/task/prompt generation as the product center; those are legacy adapters or future derived views, not the wiki authority.

### Migration and Compatibility

Do not rewrite existing `apd-docs/` or `.apd/sessions/` in place. Keep the current one-off route stable and classify its Markdown as legacy/draft documentation. Add an explicit migration path that can register an existing Markdown file as a raw source or create a source receipt referencing it, preserving its original bytes and recording the import event. A later external agent may derive wiki pages from that source; APD must not infer that the old `AI Context Pack` is already integrated knowledge.

Session YAML should be versioned independently from wiki artifacts. Add load/resume through a compatibility reader that accepts schema version 1 and writes the current version only when the user continues the session. Existing template IDs/versions and generated filenames should remain valid. The legacy `new` flow can later offer “Save as wiki intake” without changing its existing output contract.

The migration should be staged in review-sized slices:

1. Define workspace contracts, statuses, provenance, index/log formats, and deterministic lint rules.
2. Add workspace initialization, source registration, and resumable guided intake while preserving `apd new`.
3. Add external-agent work-request generation and a UI review/next-action loop.
4. Add index/log maintenance and lint reporting.
5. Add legacy registration/conversion and only then consider richer semantic integration adapters.

This keeps the planned work within the 400-line review budget per delivery slice; the full initiative will likely require chained PRs under the requested `auto-chain` strategy.

### Risks

- If APD updates semantic wiki pages itself without a stable claim/provenance contract, it will create untrustworthy summaries that look authoritative.
- A source hash alone proves byte identity, not truth; contradiction records must preserve competing grounded claims and remain open for review.
- `index.md` and `log.md` can become noisy or inconsistent unless their ownership, append/update rules, and lint semantics are explicit.
- Existing partial-export behavior is safe for drafts but unsafe if users mistake the output for integrated wiki state; statuses and UI copy must distinguish these.
- Project-local `raw/` and `wiki/` paths can collide with an existing user workspace; initialization must detect and confirm rather than overwrite.
- External-agent handoff can fail or be ignored; APD needs resumable pending work and a visible “awaiting integration/review” state.
- Template evolution may break old sessions; template version compatibility and migration behavior must be specified before implementation.
- The existing untracked `.codegraph/` index and `~/` artifact are workspace state, not product changes; they must not be included in implementation commits.

### Ready for Proposal

Yes. The product boundary is sufficiently clear for a proposal and subsequent specifications. The proposal should name the first slice as guided workspace initialization plus deterministic source intake and external-agent handoff, explicitly preserve legacy one-off document generation, and defer semantic LLM integration and recall/query. The next phase should define the file contracts and Given/When/Then acceptance scenarios before designing the TUI changes.
