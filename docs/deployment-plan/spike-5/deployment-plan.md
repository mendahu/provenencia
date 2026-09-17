# Deployment Plan — Spike 5

Interpretation foundation: candidate refs, subject vocabulary, Subjects, layout storage, FFI, and Sources-family entry into a graph stub. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §11.1 / §1.4. Authoritative schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §4.

## Status

**Planned.** Landings go in [`completed.md`](completed.md). S5-01 through S5-04, and S5-D1 / S5-D3, are recorded there.

> **This spike ships no Subject UI.** The graph is the only surface for Subjects, Citations, and Observations (design note §1.3), and the graph is Spike 6. Product nav keeps **one Sources family** — no Interpretation sidebar section (§1.4). Entry is dual action on the Sources list → graph stub. See *Scope boundary* below.

## Goal (dogfood bar)

Split, because most of this spike is not verifiable by clicking.

**Verifiable in the app:**

1. Sidebar **Sources** is primary; Source types / Source fields / Subject types / Subject fields are nested config (Subject* stubs OK); **no** Interpretation item.
2. On the Sources list, open **Evidence graph** on a Source with an Artifact → stub; Source page still opens as today.
3. No-Artifact Sources stay listed; Evidence graph action disabled; Source page remains reachable.
4. Back/Forward and relaunch restore the Evidence graph stub place; returning is a session-cache hit.
5. Creating a project seeds all seven Subject types.

**Verifiable only by test or inspection:**

6. `subjects.Create` mints a ref off its type's `candidate_ref_prefix` (`CPR-…`, `CEV-…`, `CPL-…`).
7. Rename and delete a subject; both appear in the audit log with a `subject` entity type.
8. `subject_positions` round-trips through the FFI and survives reopen.
9. A `FakeStore` round-trip covers every new RPC.

## Design track (no PRs)

**All UI is designed in Claude Design first.** Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S5-D3** | Sources section nav | Nested Sources family; Subject types / Subject fields stubs; **no** Interpretation item | S5-07 |
| **S5-D2** | Sources list → Evidence graph | Dual action, no-Artifact gate, Evidence graph stub | S5-08 |

**S5-D1 is superseded** (designed a top-level Interpretation item). Do not implement it.

Run S5-D3 before S5-D2. Subject types / Subject fields **editors** are out of scope — nav stubs only.

## PR sequence

```text
   design                          build
─────────────                ──────────────────────────────────────────────

S5-D3  Sources nav           S5-01  Subject type prefixes (core/ref)   done
  │                            │
  │                            ▼
  │                          S5-02  Migration 000021                 done
  │                            │    subject_types, subjects, subject_positions
  │                            ▼
  │                          S5-03  core/database/subjecttypes + seed done
  │                            ▼
  │                          S5-04  core/database/subjects + audit    done
  │                            ▼
  │                          S5-05  core/database/subjectpositions
  │                            ▼
  │                          S5-06  Proto + dispatch + handlers
  │                            ▼
  └────── gates ────────────▶ S5-07  Location discriminator, Subject
                               │     types/fields sections, graph place,
                               │     stub destinations
S5-D2  List dual action        │
  │                            ▼
  └────── gates ────────────▶ S5-08  Sources list Interpret action
                               │     + no-Artifact gate
                               ▼
                             S5-09  Docs, dogfood, cleanup
```

---

## Checklist

- [x] S5-D1 — Design: Interpretation nav entry → [`completed.md`](completed.md) (**superseded** — do not implement)
- [x] S5-D3 — Design: Sources section nav (nested config + Subject*) → [`completed.md`](completed.md)
- [ ] S5-D2 — Design: Sources list → Evidence graph → [`design/`](design/)
- [x] S5-01 — Subject type prefix validation in `core/ref` → [`completed.md`](completed.md)
- [x] S5-02 — Interpretation schema migration → [`completed.md`](completed.md)
- [x] S5-03 — subject type vocabulary and create-time seed → [`completed.md`](completed.md)
- [x] S5-04 — Subject CRUD with audit → [`completed.md`](completed.md)
- [ ] S5-05 — Graph layout positions
- [ ] S5-06 — FFI methods for subject types, subjects, and positions
- [ ] S5-07 — Nested Sources nav + Subject* stubs + graph place + `WorkspaceLocation` discriminator
- [ ] S5-08 — Sources list → Evidence graph stub
- [ ] S5-09 — Docs, dogfood, cleanup

---

## Schema (S5-02)

One migration, `core/database/migrations/000021.sql`. Two tables are transcribed from the interpretation data model §4.1–4.2 and must not drift from it; the third is new in this spike and specified in the design note §5.1.

```sql
CREATE TABLE subject_types (
	id                   BLOB PRIMARY KEY,
	key                  TEXT NOT NULL,
	origin               TEXT NOT NULL,
	label                TEXT NOT NULL,
	description          TEXT,
	ref_prefix           TEXT NOT NULL UNIQUE,
	candidate_ref_prefix TEXT NOT NULL UNIQUE,

	UNIQUE (key, origin)
) STRICT;

CREATE TABLE subjects (
	id           BLOB PRIMARY KEY,
	ref          TEXT UNIQUE NOT NULL,
	source_id    BLOB NOT NULL REFERENCES sources(id),
	subject_type_id BLOB NOT NULL REFERENCES subject_types(id),
	label        TEXT,
	description  TEXT,

	UNIQUE (id, subject_type_id)
) STRICT;

CREATE TABLE subject_positions (
	subject_id BLOB PRIMARY KEY REFERENCES subjects(id) ON DELETE CASCADE,
	grid_x     INTEGER NOT NULL,
	grid_y     INTEGER NOT NULL
) STRICT;
```

Decisions baked into that DDL, each argued in the design note:

- **`subject_types` carries two prefixes**, `ref_prefix` for Conclusion-layer canonical entities and `candidate_ref_prefix` for Interpretation subjects, because `canonical_entities` will reference these same rows (interpretation model §4.1). Both are `UNIQUE`, but that is not sufficient — see the gotcha on the shared namespace.
- **`subjects` has no `ON DELETE` clause** on either FK — deliberately `NO ACTION`, so a Source with Subjects cannot be deleted out from under them. The cascade for `observations` → `subjects` is a *later* decision and must be made when that table ships (§4.3), not retrofitted.
- **`subject_positions` is keyed by `subject_id` alone** — one position per subject; graph scope is `subjects.source_id`. No plan for the same subject on multiple Source graphs, so `source_id` is not denormalized onto the layout row.
- **`subject_positions.subject_id` cascades on delete**, which is the deliberate contrast: layout is disposable, evidence is not. (Source delete of layout is indirect: subjects block Source delete via `NO ACTION`, and removing subjects drops positions.)
- **No `UNIQUE (grid_x, grid_y)`.** A drag that swaps two bubbles transiently collides; let the UI nudge (§5.1).
- **Coordinates are signed** — the plane is unbounded around an origin, so no `CHECK (grid_x >= 0)`.
- **`subject_positions` is unaudited.** State this in the migration comment and be ready to defend it, because the nearest sibling (`source_metadata_layout`, migration `000013`) *is* audited. The distinction: dismissing a metadata suggestion is a research decision; arranging bubbles is not.

## Seeded Subject types (S5-03)

All seven from [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1, in a code registry installed at create time only — never in a migration, never healed on open:

```text
key              ref_prefix    candidate_ref_prefix
person           PER           CPR
event            EVT           CEV
place            PLC           CPL
relationship     REL           CRL
participation    PTN           CPA
location         LOC           CLO
source           SRN           CSR
```

Seed all seven even though none is placeable this spike — there is no Subject UI at all. They are rows, not features; a partial seed just means editing the registry again later, and the authoritative doc already fixes the set.

---

## Scope boundary: what "ends at the canvas" means

The judgment call is what **Interpret** opens. The product no longer has an Interpretation section — entry is dual action on the Sources list (design note §1.4).

| Option | Verdict |
| --- | --- |
| Render the canvas | Rejected — Spike 6. |
| Render a plain list of the Source's Subjects | Rejected — graph is the only Subject surface (§1.3). |
| **Render a "coming soon" placeholder** | **Taken.** Proves navigation and the deep place. Spike 6 replaces the view. |
| Second Interpretation Sources list | Rejected — duplicates Sources and empties the layer nav. |

**Also in this spike's UI:** Subject types / Subject fields **nav stubs** (S5-D3 / S5-07). Real editors wait for the vocabulary slice.

**Location gotcha:** Source page and graph share `.sources` + `sourceId`. S5-07 must extend `WorkspaceLocation` with a page-vs-graph discriminator (design note §11.1.2) or history collapses the two places.

**Accessibility:** canvas representation is slice-2 work in Spike 6 (§7.4).

---

## Key decisions and gotchas

Findings from the pattern inventory that will otherwise cost a day each.

| Area | The trap |
| --- | --- |
| **Schema hash** | `core/database/schemahash.go` computes `expectedSchemaHash` at init by migrating an in-memory DB. There is **no committed golden constant to bump** — counterintuitive if you expect a golden file. Nothing to regenerate. |
| **`Open` rejects unknown schema** | The digest covers all of `sqlite_schema`; `catalog_test.go` proves even a rogue *index* makes a catalog unopenable. Nothing may be created at runtime — every table and index ships in `000021.sql`. |
| **Test tags** | `CGO_ENABLED=1 go test -tags fts5 ./...`. Bare `go test` on `core/database` fails confusingly, because migrations 18–19 create FTS5 virtual tables. |
| **One ref format — do not reintroduce a candidate form** | Candidate subjects are ordinary `AAA-TTTTT` refs off `subject_types.candidate_ref_prefix`. There is no `MintCandidate`, no candidate validator, and nothing in `core/search` to teach. An earlier draft of S5-01 built an infix marker (`PER-C-7KD45`) and it was reverted; see [`catalog-refs.md`](../../catalog-refs.md) §2. |
| **Two prefixes per Subject type, one namespace** | `subject_types` carries `ref_prefix` (canonical, `PER`) and `candidate_ref_prefix` (Node, `CPR`) because `canonical_entities` shares the table. Both draw from the same three-letter space, so a new prefix must be checked against **both** columns — two per-column `UNIQUE` constraints do not express that. The leading `C` is convention; do not validate it. |
| **Reserved prefixes** | The guard must reject `USR`, `SRC`, `ART`, `CIT`, `OBS`. Shipped as `ref.ValidatePrefix` in S5-01. |
| **Prefix collisions are invisible to upsert** | `ON CONFLICT (key, origin)` does not catch a duplicate prefix; it arrives as a raw constraint error and will surface as `internal.unknown` unless mapped to its own code. |
| **Subject refs need a join inside the transaction** | Unlike `sources` (constant `SRC`), minting a Subject ref means reading `subject_types.candidate_ref_prefix` first. Extend the `requireType(tx, …)` existence check to return the prefix, then reuse the 8-attempt retry loop from `sources.go`. |
| **Audit needs no registry entry** | `EntityType` / `ActionType` are unvalidated free-form strings — which also means a typo persists silently. Convention: singular snake_case table name, `{verb}_{entity}`. |
| **Protobuf codegen is manual and dual-target** | `scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`, and writes **both** `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Commit both; forgetting the Swift half breaks the app build. Methods start at 45 (`METHOD_SEARCH_CATALOG = 44` is current highest). |
| **`?? .sourcesList` hides a missing spec** | `WorkspaceDestinationHost.presentation(for:)` falls back silently. Land Subject types/fields specs and the graph presentation in S5-07. |
| **Two hard test gates** | `PlaceRegistryTests` / `WorkspaceDestinationHostTests` — expect new cases for Subject types, Subject fields, and the graph presentation. |
| **History is not forward-compatible** | Unknown `WorkspaceSection` raw values fail decode. Adding `subject-types` / `subject-fields` is fine going forward; removing a shipped section is not. |
| **Nested sidebar is new** | Today's rail is flat `WorkspaceSection.allCases`. S5-D3 requires a parent/children presentation — not only two new section cases. Budget view work in S5-07. |
| **Page vs graph collide without a discriminator** | `WorkspaceLocation ==` is section + sourceId + fieldId + typeId. Source page and Evidence graph need a new identity field under `.sources` (design note §11.1.2). Legacy entries without it mean `.page`. |
| **Sources list needs `WorkspaceNavigation`** | Inject `@Environment(WorkspaceNavigation.self)` for Interpret → `go(to:)`. |
| **Stub destination is enough** | Graph and Subject types/fields may be one-line stubs. Do not build editors or a node list. |
| **No-Artifact gate** | Confirm list payload exposes artifact presence (or extend the query). |
| **Subjects are deliberately non-searchable** | Projecting them into the omnibar would mean a `KindNode`, a registry `KindSpec`, a projector, a `WorkspaceLocation` mapper, a `projection_version` bump, and `FakeStore` parity. Declare Subjects out of search for this spike and revisit when they carry names. |
| **Onboarding tests exercise the seed** | `sourceFixture` runs full onboarding, so a bug in `subjecttypes.Install` fails tests across `api/ffi/handlers`, `core/onboarding`, and several `core/database` packages at once. |

---

## Accepted risk: an entire layer with no consumer

This was originally scoped as "`subject_positions` ships with no UI exercising it." It is now broader: **`subjects`, `subject_types`, and `subject_positions` all ship with no UI exercising them.** Every RPC added in S5-06 is called only by Go tests and a `FakeStore` round-trip. Nothing a researcher can click reaches any of it.

This is deliberate, and it is the same argument as before, only larger. The alternative is opening Spike 6 with a migration, three Go packages, and a set of FFI methods before a single line of canvas code — exactly the "backend spike first" shape this plan exists to avoid.

Because the risk grew, the mitigation has to be stronger than it was:

- **Coverage is the deliverable, not a side effect.** Every RPC needs a Go handler test *and* `FakeStore` parity, because `FakeStore` is what Spike 6's Swift work develops against. A `FakeStore` that disagrees with the real store is the failure mode that will cost Spike 6 a day.
- **Round-trip, do not just write.** Assert reads after reopen, not only successful inserts — persistence across relaunch is precisely what no human will verify this spike.
- **The schema is the expensive mistake.** Go packages and RPCs are cheap to change in Spike 6; the migration is not, once a dogfood project exists. Spend the review effort on `000021.sql` and treat the rest as provisional. `subject_positions` remains the cheapest of the three to migrate later, being unaudited UI state (§5.1).

---

## Forward compatibility: the family tree view

The canvas this spike leads up to is expected to be reused for an eventual family tree over Conclusion-layer canonical entities. Design note §13 works through what that implies; the short version for **this** spike is that it changes almost nothing, because two of the three shared pieces are already shared by the model docs:

- `canonical_entities.subject_type_id` references the same `subject_types` rows S5-03 seeds, so the tree inherits that vocabulary for free.
- Canonical refs and Subject refs share one format, differing only in which of the type's two prefixes they are minted from (`PER-…` against `CPR-…`), so S5-01's guard serves both layers and a tree view needs no new ref work.

The one thing it did change is the name of the layout table. It is **`subject_positions`**, not `graph_subject_positions`: there is deliberately no `graphs` entity (design note §5.1), and a future tree will get its own `canonical_entity_positions` rather than a shared polymorphic table — a shared subject column could not carry a foreign key, which would forfeit the cascade behavior that is the main reason this table is shaped the way it is.

**Do not pre-generalize anything else here.** The reuse worth protecting lives in the canvas geometry, which is Spike 6's problem and is addressed there by putting those primitives in a neutral module rather than inside `Features/Interpretation/`.

---

## Suggested PR titles (why-focused)

| Step | Title sketch |
| --- | --- |
| S5-01 | Guard subject type ref prefixes in `core/ref` |
| S5-02 | Add interpretation node schema and graph layout storage |
| S5-03 | Seed subject type vocabulary at catalog create |
| S5-04 | Add audited subject CRUD for the interpretation layer |
| S5-05 | Persist graph node positions outside the audit trail |
| S5-06 | Expose subject types, subjects, and positions over FFI |
| S5-07 | Nest Sources config nav and register the evidence graph place |
| S5-08 | Open an evidence graph from the Sources list |
| S5-09 | Document the interpretation foundation and close spike 5 |

---

## Parallelism

| Track | Steps |
| --- | --- |
| **Design (no PRs)** | S5-D3 (done) → S5-D2 (S5-D1 superseded) |
| **Go core (critical path)** | S5-01 → S5-06 |
| **Mac client** | S5-07 → S5-08 → S5-09 |

S5-01 and S5-D3 are done. Go chain is still the critical path. Design: finish S5-D2 next.

---

## Definition of done

Jake can, on his MacBook:

1. Complete the in-app half of the dogfood bar — nested Sources config, Evidence graph stub from the list, no-Artifact rows behave.
2. Run `CGO_ENABLED=1 go test -tags fts5 ./...` with FakeStore round-trips for every new RPC.
3. Point at `subject_types` / `subjects` / `subject_positions`, the Evidence graph place, and the location discriminator as what Spike 6 builds on.
4. Confirm honesty: stubs do not claim Subjects exist or can be created yet.

Per [`versioning.mdc`](../../../.cursor/rules/versioning.mdc), the docs in this folder do not bump `VERSION`; cutting a release that contains the spike bumps product PATCH across `VERSION`, `core/version.go`, and both Xcode `MARKETING_VERSION` configurations.

---

## What Spike 6 inherits

So the boundary is unambiguous — on the day Spike 6 opens, these exist and work:

- `subjects` rows with candidate refs, created and deleted through an audited Go package and an FFI method.
- A seeded Subject type vocabulary with both `ref_prefix` and `candidate_ref_prefix` values.
- A positions table and the RPCs to read and write it.
- A `.sourceGraph(project:sourceId:)` query key (design note §5.3).
- `WorkspaceLocation` page-vs-graph discriminator under Sources.
- Subject types / Subject fields sidebar stubs (editors still later).
- Sources list Interpret → graph stub, working Back/Forward.
- A `FakeStore` that answers every new RPC.

Spike 6's first PR swaps the stub for the `NSScrollView` bridge and a bubble.

Not inherited (Spike 6 / later):

- Real canvas; Source-page Interpret control; canvas a11y (§7.4).
- Subject types / Subject fields **editors**.
