# Spike 2 — Source layer catalog (schema, CRUD, ingest UI)

## Status

**Current.** Planning notes for the next implementation sprint: validate the Source-layer model end-to-end (tables → Go CRUD/ingest → FFI → macOS catalog UI), with design work in Claude Design interleaved where the UX is non-trivial.

Authoritative models:

- [`source-layer-data-model.md`](../../source-layer-data-model.md)
- [`artifact-file-storage.md`](../../artifact-file-storage.md) (pointer → Source layer)
- [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §2
- [`audit-revision-history.md`](../../audit-revision-history.md)
- [`structured-date-model.md`](../../structured-date-model.md)
- [`catalog-refs.md`](../../catalog-refs.md)
- [`application-stack.md`](../../application-stack.md) (FFI granularity, `objects/` ingest)
- [`macos-client-patterns.md`](../../macos-client-patterns.md)

Claude Design handoffs for **S2-01…S2-04** live in [`design/`](design/) — self-contained requirement briefs, not the task list below.

Spike 1 left an explicit gate: **the first researched mutation must write audit**, not decorative empty tables without a write path. This spike owns that gate.

---

## Goal

A signed-in researcher can, in a real `*.provenencia` project:

1. Work inside an **app workspace** (sidebar + content) instead of a one-screen onboarding home forever.
2. Create a Source (type, title, description, notes, descriptive metadata).
3. Extend the project’s Source types and metadata fields when the seeded set is not enough.
4. Add Artifacts under a Source (including fileless / physical-only placeholders).
5. Ingest a local file into content-addressed `objects/`, attach it as an Artifact’s primary File, and see enough UI to verify the model (list, detail, metadata, type extension).

That is enough to **validate** the Source-layer schema and storage rules before Interpretation work begins.

```text
Onboarding (Spike 1) → signed in
  → App workspace (sidebar + content)
       ├── Sources (catalog → Source detail)
       │     ├── type + title + notes + metadata
       │     ├── extend types / fields (project-local)
       │     └── Artifacts
       │           ├── fileless placeholder
       │           └── ingest File → objects/{hh}/{hh}/{sha256} → Artifact.file_id
       ├── (stub / disabled) later areas — Interpretation, etc.
       └── Project / session (sign out, project label, contributor)
```

---

## In scope

- **App workspace chrome**: side navigation + content host. Spike 2 destinations: **Sources**, **Source types**, **Source fields** (no “coming soon” stubs for later layers).
- Audit write path (`audit_transactions` / `audit_changes`) used by every Source-layer mutation.
- Shared `date_values` table (schema + minimal Go helpers) so date-typed Source metadata can land without a second migration later.
- Full Source-layer table set from the Source doc: `source_types`, `sources`, `source_notes`, `source_metadata_fields`, `source_type_metadata_fields`, `source_metadata`, `artifacts`, `files`, `file_derivatives`.
- Small **seed** of types/fields (not the entire [`seeded-vocabulary.md`](../../seeded-vocabulary.md) horizon list). Prefer photograph + book + one vital/census-shaped type unless design picks a different dogfood set.
- Go domain packages for CRUD + ingest; FFI use-cases (coarse verbs); SwiftUI Source catalog UI inside the workspace, consuming `GenealogyStore`.
- Claude Design boards for workspace chrome, Source catalog IA, create/edit flows, artifact + ingest, and vocabulary extension.
- Refs: mint `SRC-…` / `ART-…` via `core/ref` on insert ([`catalog-refs.md`](../../catalog-refs.md)).

## Out of scope (later spikes)

- Interpretation (Citations, Observations, Nodes) and Conclusion — do not add sidebar placeholders for them in this spike.
- Source **credibility** assessments ([`research-judgment-model.md`](../../research-judgment-model.md)).
- Full GEDCOM / import adapters.
- Destructive primary File deletion; orphan GC of historically referenced Files.
- Rich audit UI / timeline browser (writes must exist; browsing history can wait).
- Open/create project flows beyond what Spike 1 already shipped (“continue as / that’s not me” polish is a separate spike unless it blocks dogfood).
- Windows client, sync, cloud accounts, notarization.
- Shipping the entire seeded vocabulary catalog on day one.

---

## Non-negotiable constraints (from docs)

1. **Audit atomicity** — domain write + audit revision in one SQLite transaction ([`audit-revision-history.md`](../../audit-revision-history.md) §5).
2. **No `created_at` / `updated_by` on Source tables** — attribution lives in audit.
3. **File bytes never travel over protobuf** — ingest use-case; read path returns relative `objects/…` path; Swift reads bytes ([`application-stack.md`](../../application-stack.md)).
4. **Content-addressed immutability** — `objects/{hh}/{hh}/{full hex}`; replace scan = new File + Artifact pointer update.
5. **Derivatives belong to Files**, not Artifacts; generation need not audit.
6. **Swift stays thin** — Go owns schema, ingest, validation; models + `FakeStore` for tests.
7. **Seed small; grow from use** ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) §1).

---

## Definition of done

Jake can, on his MacBook, without a server:

1. Finish Spike 1 onboarding and land in an **app workspace** with a sidebar (not the old single “you’re signed in” home as the permanent shell).
2. Use the sidebar to open Sources, Source types, and Source fields.
3. Create a Source of a seeded type, edit title/description, add a note, set text (and at least one date) metadata fields suggested for that type.
4. Add a custom Source type and/or metadata field (from their nav destinations) and use them on a Source.
5. Add a fileless Artifact and an Artifact with an ingested image/PDF; confirm `files` row + bytes under `objects/`.
6. Inspect `provenencia.sqlite`: Source-layer rows present; audit revisions recorded for creates/updates; `SRC-…` / `ART-…` refs present.
7. Relaunch: workspace returns; catalog still lists the Sources; File still opens from the relative path.

---

## Step numbering

Steps are **`S2-NN`** (Spike 2, two-digit sequence). Each step has a **kind**:

| Kind | Meaning |
| --- | --- |
| **Design** | Claude Design (and/or local UX writing). No product code required to “complete,” but outputs feed later PRs. |
| **PR** | Mergeable change on `main`; leave the app buildable. |
| **Doc** | Docs/rules/skills only when a decision must land before code. |

IDs stay stable even if order of *starting* work shifts; **Depends on** is the merge/start gate.

---

## Dependency sketch

```text
S2-01 Design — App workspace chrome (sidebar + content host)
S2-02 Design — Source catalog IA + create/edit
S2-03 Design — Artifact + ingest + fileless
S2-04 Design — Types / metadata extension
        │
        ▼
S2-05 PR — Audit tables + write helper (done)   ◄── gate for research mutations
S2-06 PR — date_values schema + helpers
S2-07 PR — Source vocabulary tables + small seed
S2-08 PR — sources + source_notes CRUD (+ audit)
S2-09 PR — files store + ingest (+ audit on File create)
S2-10 PR — artifacts CRUD + attach/replace File (+ audit)
S2-11 PR — source_metadata (+ type field suggestions)
S2-12 PR — file_derivatives + thumbnail pipeline (minimal)
S2-13 PR — FFI Source use-cases
S2-14 PR — Swift app workspace layout (sidebar shell)   ◄── discrete chrome step
S2-15 PR — Swift Source catalog (list + navigate into detail)
S2-16 PR — Swift create/edit Source + notes + metadata
S2-17 PR — Swift Artifacts + file picker ingest
S2-18 PR — Swift extend types / fields
S2-19 PR — Dogfood polish (copy, empty states, errors, tests)
```

Design **S2-01…04** can start immediately; **S2-01** should land a reviewable board before **S2-14**. Core PRs **S2-05…07** can run in parallel with design. Do not start feature UI PRs (**S2-15+**) until the relevant Design step has a reviewable board (or an explicit “design enough to code” note). **S2-14** may ship with placeholder destination panes before FFI is ready.

---

## Steps

### S2-01 — Design: App workspace chrome (sidebar + content)

**Claude Design brief:** [`design/S2-01-workspace-chrome.md`](design/S2-01-workspace-chrome.md)

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | — |
| **Deliverables** | Board for the post-onboarding **app workspace**: left (or leading) sidebar / nav rail with destinations **Sources**, **Source types**, and **Source fields**; main content host; how project label + contributor identity live in the chrome; Sign Out in the app menu only; label expand/collapse toggle; selected vs idle nav states; short-window sidebar scroll. Transition from Spike 1 onboarding into this shell. No “coming soon” nav stubs. |
| **Context** | [`application-stack.md`](../../application-stack.md) (Mac owns windows/navigation); design system README lists `SidebarNav` as add-on-demand — this board decides whether to port that component or a bespoke onboarding-aligned rail. Preserve Provenencia visual language (tokens already in `macos/App/DesignSystem/`). One clear composition: sidebar + one primary content column, not a multi-panel dashboard. |
| **Out** | Source list/detail layouts (S2-02); full type/field editors (S2-04); Settings product; future-layer nav stubs. |
| **Feeds** | S2-14 (and constrains how S2-15+ mount content) |

---

### S2-02 — Design: Source catalog IA and create/edit

**Claude Design brief:** [`design/S2-02-source-catalog.md`](design/S2-02-source-catalog.md)

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | S2-01 (content area of the workspace; can parallelize once chrome frames exist) |
| **Deliverables** | Board for: Sources destination → Source list → Source detail (identity strip with `SRC-…`, type, title, description, notes list, metadata form). Create-Source flow (type picker → essentials → save). Empty states. Lives **inside** the workspace content host from S2-01 — not a separate window chrome. |
| **Context** | [`source-layer-data-model.md`](../../source-layer-data-model.md) §§3–5; existing onboarding visual language. |
| **Out** | No Citations/Observations; no credibility grade UI. |
| **Feeds** | S2-15, S2-16 |

---

### S2-03 — Design: Artifacts, ingest, fileless evidence

**Claude Design brief:** [`design/S2-03-artifacts-ingest.md`](design/S2-03-artifacts-ingest.md)

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | S2-02 (same Source detail; can parallelize early frames) |
| **Deliverables** | Artifact list under a Source; add fileless Artifact; “Add file…” ingest; replace primary File; show original filename / media type / size; preview or placeholder for images/PDFs; physical-only copy. Clarify that multiple scans = multiple Artifacts, not one Artifact with many files. |
| **Context** | Source doc §§6–8, 11 rules 7–18; stack read/write path for Files. Bytes never staged as base64 in UI models. |
| **Out** | Derivative management UI beyond “show thumbnail if present”; historical File version browser. |
| **Feeds** | S2-17 (and informs S2-12 thumbnail needs) |

---

### S2-04 — Design: Extensible types and metadata fields

**Claude Design brief:** [`design/S2-04-extensible-vocabulary.md`](design/S2-04-extensible-vocabulary.md)

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | S2-02 |
| **Deliverables** | UX for the **Source types** and **Source fields** destinations: add a project-local Source type; add a metadata field (`text` / `date`); attach suggested fields to a type; pick suggested fields when editing a Source; use a field not suggested for the type. Builtin vs custom affordance (label only — not a privileged subclass). Top-level nav placement is fixed by S2-01. |
| **Context** | Source doc §§3, 5; [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §1–2 (seed small). Agree the **dogfood seed set** (proposed default below) and write it on the board. |
| **Proposed seed (amend in Design)** | Types: `photograph`, `book`, `birth_certificate`. Fields: enough for those three type suggestion lists to feel real (subset of §2.2–2.3, not the full horizon). |
| **Feeds** | S2-07 seed content, S2-18 |

---

### S2-05 — PR: Audit tables and atomic write helper

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — (merge before any Source mutation PR) |
| **Deliverables** | Done. Migration `000003` + `core/database/audit.Record` on `sql.Tx`; create/update/delete JSON and monotonic revision tested. |
| **Context** | Spike 1 explicit gate. Skills: [`add-catalog-migration`](../../../.cursor/skills/add-catalog-migration/SKILL.md), [`add-catalog-query`](../../../.cursor/skills/add-catalog-query/SKILL.md). Entity types will include `source`, `artifact`, `file`, `source_metadata`, etc. as later PRs land — helper should accept `entity_type` string. |
| **Notes** | Do not audit the audit tables. `user_id` from install/session UUID already in `users`. |

---

### S2-06 — PR: `date_values` schema and minimal helpers

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — (can parallel S2-05; must merge before date metadata writes in S2-11) |
| **Deliverables** | Migration for `date_values` ([`structured-date-model.md`](../../structured-date-model.md)). Go insert/lookup helpers for a **small** set of kinds needed by Source metadata (exact / year-only / `ABT` / simple range as required by dogfood — document chosen subset in the PR). Unit tests. No parser UI. |
| **Context** | Shared infrastructure, not Source-owned. Keep vocabulary flexible; do not freeze a full GEDCOM date language in this PR. |

---

### S2-07 — PR: Source vocabulary tables + small seed

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-04 seed decision (or interim seed with note “Design may trim”) |
| **Deliverables** | Migration: `source_types`, `source_metadata_fields`, `source_type_metadata_fields`. Seed on **Create** (and document behavior for Open of older DBs: migrate schema; seed only if empty — pin the rule in the PR). Go list/upsert for types and fields (project-local `builtin=0` rows). Tests for uniqueness of `key` and suggestion join rows. |
| **Context** | Source doc §§3, 5.1–5.2; vocabulary doc §2. Stable UUIDs for builtin rows are nice-to-have for tests; if used, document how they are minted. |
| **Out** | No `sources` rows yet. |

---

### S2-08 — PR: `sources` + `source_notes` CRUD (audited)

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-05, S2-07 |
| **Deliverables** | Migration: `sources`, `source_notes`. Mint `SRC-…` on create. Domain ops: create/update Source; add/update/delete note; list Sources (enough for catalog); get Source by id/ref. Every mutation goes through audit (`action_type` e.g. `create_source`, `update_source`, …). Go tests with Fake/real catalog. |
| **Context** | Source doc §4; [`catalog-refs.md`](../../catalog-refs.md); `ref.PrefixSource`. |
| **Out** | FFI/UI; metadata; artifacts. |

---

### S2-09 — PR: Files store + ingest (audited File create)

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-05 |
| **Deliverables** | Migration: `files`. Go ingest: read path on disk → SHA-256 → write `objects/{hh}/{hh}/{hex}` if absent → upsert `files` by checksum → return `file_id` + relative path. Dedup same bytes = same File row. Audit File **create** when a new row appears (reuse of existing checksum is not a new research event — document). Tests with temp project dirs. |
| **Context** | Source doc §6; stack object naming. No Artifact linkage yet. |
| **Out** | Derivatives; Swift file UI. |

---

### S2-10 — PR: `artifacts` CRUD + attach/replace File

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-08, S2-09 |
| **Deliverables** | Migration: `artifacts`. Mint `ART-…`. Create fileless Artifact; attach File; replace `file_id` (new File, old File retained). List by `source_id`. Audited. Tests for ON DELETE CASCADE from Source and null `file_id`. |
| **Context** | Source doc §§7, 9, 11. Primary File deletion still unsupported. |

---

### S2-11 — PR: `source_metadata` (text + date)

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-06, S2-07, S2-08 |
| **Deliverables** | Migration: `source_metadata`. Set/clear metadata values; validate `data_type` vs `date_value_id` / `value_text` in Go. Helpers to load suggested fields for a Source’s type plus any extra values present. Audited (`update_source_metadata` or per-row changes under one revision). Tests. |
| **Context** | Source doc §5.3. Date fidelity: keep `value_text` even when structured date exists. |

---

### S2-12 — PR: `file_derivatives` + minimal thumbnail

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-09 (ideally after S2-03 confirms preview needs) |
| **Deliverables** | Migration: `file_derivatives`. Generate a small image thumbnail into `derivatives/` (or objects — follow Source doc: derivative Files still live in the `files` store; path convention as documented). Idempotent by `(source_file_id, derivative_type)`. **No** research audit for generation. Tests with a tiny fixture image. |
| **Context** | Source doc §8. Skip non-image MIME types cleanly. |
| **Notes** | If Design S2-03 accepts placeholders-only for v0, this PR may shrink to schema + stub “ensure thumbnail” no-op — say so in the PR rather than inventing preview UI debt. |

---

### S2-13 — PR: FFI Source use-cases

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-08…S2-11 (S2-12 optional) |
| **Deliverables** | Proto methods + handlers for coarse ops, e.g. `ListSources`, `GetSourceWorkspace`, `CreateSource`, `UpdateSource`, `AddSourceNote`, `SetSourceMetadata`, `CreateArtifact`, `IngestArtifactFile` (path in request), `ListSourceTypes` / `CreateSourceType` / `CreateMetadataField` as needed by UI. Map `apperr` codes; add L10n cases for new codes. Handler tests via `runRPC`. **No** file bytes in protobuf. |
| **Context** | Skills: [`add-ffi-handler`](../../../.cursor/skills/add-ffi-handler/SKILL.md); stack § FFI granularity. Extend `GenealogyStore` + `FakeStore` in the same PR or with S2-15 — prefer same PR if FakeStore otherwise blocks Swift tests. |

---

### S2-14 — PR: Swift app workspace layout (sidebar shell)

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-01 (design enough); Spike 1 onboarding “signed in” as the entry point to replace |
| **Deliverables** | Replace the permanent onboarding home as the post-sign-in shell with an **app workspace**: sidebar (or nav rail) + content host. Destinations: **Sources**, **Source types**, **Source fields** (empty panes OK until later PRs); session/project affordances matching S2-01; **Sign Out** via app menu only. Selection state drives which content view is shown. Unit tests for navigation selection if modeled explicitly. L10n for nav labels. Design-system component only if S2-01 chose to port `SidebarNav`; otherwise keep chrome in `Features/Workspace/` (or similar) without inventing a second design language. |
| **Context** | [`macos-client-patterns.md`](../../macos-client-patterns.md); [`application-stack.md`](../../application-stack.md) (navigation is a Mac concern). This step is **chrome only** — no Source CRUD UI. |
| **Out** | Source list/detail (S2-15+); full type/field editors (S2-18); interpreting other layers. |
| **Notes** | May merge **before** S2-13 if placeholders do not call new RPCs. Do not bury this inside the Source catalog PR. |

---

### S2-15 — PR: Swift Source catalog (list + navigate into detail)

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-02 (design enough), S2-13, S2-14 |
| **Deliverables** | Sources destination content: list rows (title/type/`SRC-…`), empty state, navigate into a Source detail shell (sections stubbed if needed). New `Features/Sources/` model+views mounted in the workspace content host. Unit tests with `FakeStore`. L10n via skill. |
| **Context** | Mount under S2-14 — do not reintroduce a separate top-level window chrome. |
| **Out** | Full edit forms (S2-16). |

---

### S2-16 — PR: Swift create/edit Source, notes, metadata

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-02, S2-15, S2-13 |
| **Deliverables** | Create Source; edit title/description/type; notes CRUD UI; metadata editor driven by type suggestions + existing values; basic date field UX (structured subset matching S2-06). Wire errors to toast/L10n. Model tests. |
| **Context** | Match S2-02 board; prefer `Error?` / existing onboarding patterns. |

---

### S2-17 — PR: Swift Artifacts + file ingest

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-03, S2-15, S2-13 (S2-12 if thumbnails shown) |
| **Deliverables** | Artifact list; add fileless; NSOpenPanel → pass path to ingest RPC; show filename/MIME/size; display thumbnail or placeholder; replace file. Entitlements/docs folder usage already present — confirm any new file-access copy. Tests for model state transitions with FakeStore. |
| **Context** | Stack: Swift must not write `objects/` itself. |

---

### S2-18 — PR: Swift extend types and metadata fields

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-04, S2-16, S2-13 |
| **Deliverables** | UI to create custom Source type; create metadata field; attach field to type (sort order); use custom type/field in create/edit. Builtin rows not deletable (or delete refused with clear error). Mounted in the **Source types** / **Source fields** destinations from S2-01. |
| **Context** | S2-04 board; vocabulary principles. |

---

### S2-19 — PR: Dogfood polish and regression net

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S2-14, S2-16, S2-17, S2-18 |
| **Deliverables** | Empty/error copy pass; accessibility identifiers for workspace nav + Source flows; Go+Swift test gaps closed for happy paths and one failure each (duplicate type key, ingest missing file, audit present after create). Update [`deployment-plan/README.md`](../README.md) when archiving this spike. Optional: short “how to dogfood Source catalog” note in README or spike retro. |
| **Out** | Product SemVer bump only if cutting a release ([`versioning.md`](../../versioning.md)). |

---

## Suggested PR titles (why-focused)

| Step | Title sketch |
| --- | --- |
| S2-05 | Record research mutations in append-only audit revisions |
| S2-06 | Add shared genealogical date_values storage |
| S2-07 | Seed extensible Source types and metadata fields |
| S2-08 | Create and update Sources with notes under audit |
| S2-09 | Ingest immutable content-addressed Files into the project |
| S2-10 | Attach Artifacts and primary Files to Sources |
| S2-11 | Store descriptive Source metadata (text and dates) |
| S2-12 | Generate disposable File thumbnails |
| S2-13 | Expose Source catalog use-cases over FFI |
| S2-14 | Replace post-onboarding home with a sidebar workspace shell |
| S2-15 | Add a Source catalog inside the app workspace |
| S2-16 | Let researchers create and describe Sources in the UI |
| S2-17 | Let researchers attach and ingest Artifact files |
| S2-18 | Let researchers extend Source types and metadata fields |
| S2-19 | Harden the Source catalog for first dogfood |

---

## Parallelism and team split

| Track | Steps |
| --- | --- |
| **Design (Claude Design)** | S2-01 → S2-02 → S2-03 / S2-04 (overlap OK after chrome frames) |
| **Core schema / Go** | S2-05, S2-06 in parallel; then S2-07 → S2-08; S2-09 parallel to S2-08 after S2-05; S2-10 after both; S2-11 after S2-06/07/08; S2-12 after S2-09 |
| **FFI + Mac** | S2-13 after domain; **S2-14 workspace chrome** (can precede FFI if placeholders); S2-15…18 follow Design + FFI + workspace |

Prefer **many small PRs** over combining schema+UI. Do not fold the workspace shell into the first Source list PR. Do not put ingest into the first vocabulary migration “while we’re there.”

---

## Explicit non-goals checklist (keep PRs honest)

- [ ] No Interpretation tables or RPCs (sidebar stubs only)
- [ ] No Source credibility UI
- [ ] No File bytes in protobuf
- [ ] No deletion of primary Files
- [ ] No full vocabulary dump from `seeded-vocabulary.md`
- [ ] No audit history browser required for done

---

## After Spike 2

Likely next spikes (not scheduled here):

1. Fill sidebar destinations beyond Sources / types / fields (Interpretation entry, Settings) when those features exist.
2. Audit history UI / “what changed” for a Source.
3. Citation → Observation → Node (Interpretation) on top of Artifacts.
4. Project open/create UX polish beyond Spike 1 onboarding.
5. Richer date entry and more seeded types driven by real cataloging sessions.
