# Deployment Plan — Spike 5

Interpretation foundation: candidate refs, Node vocabulary, Nodes, layout storage, FFI, and the Interpretation section root. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §11.1. Authoritative schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §4.

## Status

**Planned.** No steps landed. Landings go in [`completed.md`](completed.md).

> **This spike ships no Node UI.** The graph is the only surface for Nodes, Citations, and Observations (design note §1.3), and the graph is Spike 6 — so there is nothing in this spike a researcher can click to make a Node. The rail is proven by Go tests. Navigation is proven by a Sources list that opens a stub destination. See *Scope boundary* below.

## Goal (dogfood bar)

Split, because most of this spike is not verifiable by clicking.

**Verifiable in the app:**

1. Sidebar → **Interpretation** exists, and lands on a list of the project's Sources.
2. Activating a Source lands on a stub / "coming soon" destination for that Source — placeholders are fine; the canvas replaces it in Spike 6.
3. Back/Forward and relaunch restore the Interpretation place; returning to it is a session-cache hit.
4. Creating a project seeds all seven Node Types.

**Verifiable only by test or inspection:**

5. `nodes.Create` mints a ref off its type's `candidate_ref_prefix` (`CPR-…`, `CEV-…`, `CPL-…`), for a person, an event, and a place.
6. Rename and delete a Node; both appear in the audit log with a `node` entity type.
7. `node_positions` round-trips through the FFI and survives reopen.
8. A `FakeStore` round-trip covers every new RPC, so Spike 6's Swift work opens against a store that already answers.

## Design track (no PRs)

**All UI is designed in Claude Design first.** Design steps carry a `D` id, produce a board rather than a diff, and are *not* PRs — but they are ordered work and they gate the PRs that implement them. Briefs: [`design/`](design/).

Both briefs are writable on day one: they depend on nothing in the Go track, so the design track runs in parallel with S5-01…S5-06 and is off the critical path entirely — provided it starts at the top of the spike rather than when the implementing PR is ready to open.

One board per surface, so each can be run independently:

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S5-D1** | Interpretation nav entry | The sidebar destination — label, icon, placement, empty badge slot | S5-07 |
| **S5-D2** | Interpretation Sources list | The Source picker behind the destination: rows, purpose copy, empty state, and a light stub behind a row | S5-08 |

S5-D1 gates S5-07 because the sidebar case commits to a label and a `PVSymbol` in that PR. Stub destination views in S5-07 are exempt — they are compile scaffolding, replaced before anything ships.

Both boards are small, and the design track is no longer a meaningful schedule risk for this spike — the weight moved to Spike 6 along with the canvas.

**Two briefs were cut** when the graph became the only Node surface (design note §1.3): a *Source page entry control* and a *Source nodes destination*. Neither is re-homed here. Spike 6 will be planned fresh once this spike's results are in, and its design track written then — writing those boards now would bake in assumptions the canvas is likely to overturn.

The destination behind a Source row is a stub until Spike 6. Early development: a plain "coming soon" placeholder is fine — do not over-design it.

## PR sequence

Design steps are shown in the order they must happen. They are not PRs and they do not block the Go track — only the UI PRs they gate.

```text
   design                          build
─────────────                ──────────────────────────────────────────────

S5-D1  Nav entry             S5-01  Node Type prefixes (core/ref)
  │                            │                              no deps
  │                            ▼
  │                          S5-02  Migration 000021
  │                            │    node_types, nodes, node_positions
  │                            ▼
  │                          S5-03  core/database/nodetypes + seed
  │                            │    7 seeded types, create-time only
  │                            ▼
  │                          S5-04  core/database/nodes + audit
  │                            │    needs S5-03 (candidate prefix)
  │                            ▼
  │                          S5-05  core/database/graphlayout
  │                            │    unaudited positions
  │                            ▼
  │                          S5-06  Proto + dispatch + handlers
  │                            │    codegen Go *and* Swift
  │                            ▼
  └────── gates ────────────▶ S5-07  Section, places, query keys, store
                               │     registry plumbing, sidebar case
S5-D2  Sources list            │
  │                            ▼
  └────── gates ────────────▶ S5-08  Interpretation Sources list
                               │     rows open a stub destination
                               ▼
                             S5-09  Docs, dogfood, cleanup
```

Nine steps, not ten: the Source nodes list is gone, and the Source-page button went with it to Spike 6. A deep place ships, but only as a placeholder.

---

## Checklist

- [ ] S5-D1 — Design: Interpretation nav entry (sidebar destination) → [`design/`](design/)
- [ ] S5-D2 — Design: Interpretation Sources list → [`design/`](design/)
- [x] S5-01 — Node Type prefix validation in `core/ref` → [`completed.md`](completed.md)
- [ ] S5-02 — Interpretation schema migration
- [ ] S5-03 — Node type vocabulary and create-time seed
- [ ] S5-04 — Node CRUD with audit
- [ ] S5-05 — Graph layout positions
- [ ] S5-06 — FFI methods for node types, nodes, and positions
- [ ] S5-07 — Interpretation workspace section and root place
- [ ] S5-08 — Interpretation Sources list (stub behind a Source)
- [ ] S5-09 — Docs, dogfood, cleanup

---

## Schema (S5-02)

One migration, `core/database/migrations/000021.sql`. Two tables are transcribed from the interpretation data model §4.1–4.2 and must not drift from it; the third is new in this spike and specified in the design note §5.1.

```sql
CREATE TABLE node_types (
	id                   BLOB PRIMARY KEY,
	key                  TEXT NOT NULL,
	origin               TEXT NOT NULL,
	label                TEXT NOT NULL,
	description          TEXT,
	ref_prefix           TEXT NOT NULL UNIQUE,
	candidate_ref_prefix TEXT NOT NULL UNIQUE,

	UNIQUE (key, origin)
) STRICT;

CREATE TABLE nodes (
	id           BLOB PRIMARY KEY,
	ref          TEXT UNIQUE NOT NULL,
	source_id    BLOB NOT NULL REFERENCES sources(id),
	node_type_id BLOB NOT NULL REFERENCES node_types(id),
	label        TEXT,
	description  TEXT,

	UNIQUE (id, node_type_id)
) STRICT;

CREATE TABLE node_positions (
	source_id BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	node_id   BLOB NOT NULL REFERENCES nodes(id)   ON DELETE CASCADE,
	grid_x    INTEGER NOT NULL,
	grid_y    INTEGER NOT NULL,

	PRIMARY KEY (source_id, node_id)
) STRICT;
```

Decisions baked into that DDL, each argued in the design note:

- **`node_types` carries two prefixes**, `ref_prefix` for Conclusion-layer canonical entities and `candidate_ref_prefix` for Interpretation Nodes, because `canonical_entities` will reference these same rows (interpretation model §4.1). Both are `UNIQUE`, but that is not sufficient — see the gotcha on the shared namespace.
- **`nodes` has no `ON DELETE` clause** on either FK — deliberately `NO ACTION`, so a Source with Nodes cannot be deleted out from under them. The cascade for `observations` → `nodes` is a *later* decision and must be made when that table ships (§4.3), not retrofitted.
- **`node_positions` cascades both ways**, which is the deliberate contrast: layout is disposable, evidence is not.
- **No `UNIQUE (source_id, grid_x, grid_y)`.** A drag that swaps two bubbles transiently collides; let the UI nudge (§5.1).
- **Coordinates are signed** — the plane is unbounded around an origin, so no `CHECK (grid_x >= 0)`.
- **`node_positions` is unaudited.** State this in the migration comment and be ready to defend it, because the nearest sibling (`source_metadata_layout`, migration `000013`) *is* audited. The distinction: dismissing a metadata suggestion is a research decision; arranging bubbles is not.

## Seeded Node Types (S5-03)

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

Seed all seven even though none is placeable this spike — there is no Node UI at all. They are rows, not features; a partial seed just means editing the registry again later, and the authoritative doc already fixes the set.

---

## Scope boundary: what "ends at the canvas" means

The judgment call in this plan is what sits behind a Source in the Interpretation section. Four options were considered, and the answer changed once the graph became the **only** Node surface (design note §1.3):

| Option | Verdict |
| --- | --- |
| Render the canvas | Rejected — that is Spike 6, and the whole point of this spike is that Spike 6 starts clean. |
| Render a plain list of the Source's Nodes | **Previously taken, now rejected.** It was justified as the structured non-canvas editing path the old §7.4 promised. With one surface only, that path does not exist, so the list would be a screen we build, design a board for, and delete in slice 2. |
| **Render a "coming soon" placeholder** | **Taken.** Proves navigation and the deep place end to end. The view is throwaway scaffolding Spike 6 replaces — fine in early development. |
| Nothing — no deep place at all | Rejected once placeholders were accepted. Leaving rows with nowhere to go is more awkward than a one-line stub. |

**What this costs.** The data path still is not proven by creating a Node in the UI — there is no Node UI. Migration → Go → FFI → store is proven by Go tests and a `FakeStore` round-trip. Navigation → deep place → stub view *is* exercised by clicking through.

**Why the stub is fine.** The app is not in production; a placeholder destination is cheaper than carefully designed inert rows, and Spike 6 was always going to rewrite whatever sits behind a Source. Keep the place identity stable if convenient so Spike 6 swaps the view rather than the route.

**The accessibility consequence, stated here so it is not lost.** Dropping the node list means the canvas's accessibility representation is the only path for VoiceOver and keyboard-only researchers. Design note §7.4 moves it from the polish slice into slice 2 for that reason. Nothing in *this* spike implements it, but Spike 6 cannot treat it as optional.

---

## Key decisions and gotchas

Findings from the pattern inventory that will otherwise cost a day each.

| Area | The trap |
| --- | --- |
| **Schema hash** | `core/database/schemahash.go` computes `expectedSchemaHash` at init by migrating an in-memory DB. There is **no committed golden constant to bump** — counterintuitive if you expect a golden file. Nothing to regenerate. |
| **`Open` rejects unknown schema** | The digest covers all of `sqlite_schema`; `catalog_test.go` proves even a rogue *index* makes a catalog unopenable. Nothing may be created at runtime — every table and index ships in `000021.sql`. |
| **Test tags** | `CGO_ENABLED=1 go test -tags fts5 ./...`. Bare `go test` on `core/database` fails confusingly, because migrations 18–19 create FTS5 virtual tables. |
| **One ref format — do not reintroduce a candidate form** | Candidate Nodes are ordinary `AAA-TTTTT` refs off `node_types.candidate_ref_prefix`. There is no `MintCandidate`, no candidate validator, and nothing in `core/search` to teach. An earlier draft of S5-01 built an infix marker (`PER-C-7KD45`) and it was reverted; see [`catalog-refs.md`](../../catalog-refs.md) §2. |
| **Two prefixes per Node Type, one namespace** | `node_types` carries `ref_prefix` (canonical, `PER`) and `candidate_ref_prefix` (Node, `CPR`) because `canonical_entities` shares the table. Both draw from the same three-letter space, so a new prefix must be checked against **both** columns — two per-column `UNIQUE` constraints do not express that. The leading `C` is convention; do not validate it. |
| **Reserved prefixes** | The guard must reject `USR`, `SRC`, `ART`, `CIT`, `OBS`. Shipped as `ref.ValidatePrefix` in S5-01. |
| **Prefix collisions are invisible to upsert** | `ON CONFLICT (key, origin)` does not catch a duplicate prefix; it arrives as a raw constraint error and will surface as `internal.unknown` unless mapped to its own code. |
| **Node refs need a join inside the transaction** | Unlike `sources` (constant `SRC`), minting a Node ref means reading `node_types.candidate_ref_prefix` first. Extend the `requireType(tx, …)` existence check to return the prefix, then reuse the 8-attempt retry loop from `sources.go`. |
| **Audit needs no registry entry** | `EntityType` / `ActionType` are unvalidated free-form strings — which also means a typo persists silently. Convention: singular snake_case table name, `{verb}_{entity}`. |
| **Protobuf codegen is manual and dual-target** | `scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`, and writes **both** `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Commit both; forgetting the Swift half breaks the app build. Methods start at 45 (`METHOD_SEARCH_CATALOG = 44` is current highest). |
| **`?? .sourcesList` hides a missing spec** | `WorkspaceDestinationHost.presentation(for:)` falls back silently, so a new section with no `PlaceRegistry` spec renders the ordinary Sources list while the sidebar shows Interpretation selected. Land the root spec, deep stub spec, and the sidebar case across S5-07 / S5-08. |
| **Two hard test gates** | `PlaceRegistryTests.registryCoversAllPlaceIDs` switches exhaustively over `PlaceID`, and `WorkspaceDestinationHostTests` asserts `known.count == 4` over `WorkspacePresentationID`. Both fail the moment a case is added — expect **two** new cases (root + deep stub). |
| **History is not forward-compatible** | `WorkspaceSection.init(from:)` throws on an unknown raw value, so history written by a build with `interpretation` fails to decode on a build without it (surfacing as `NavigationHistoryIssue.loadFailed`). Acceptable for a solo dogfood, worth knowing before bisecting. |
| **The Source page cannot navigate today** | Nothing in `Features/Sources/` except `SourcesListView` holds `@Environment(WorkspaceNavigation.self)`. Not this spike's problem any more — the Source-page button moved to Spike 6 — but it is the first thing that work will hit. |
| **Stub destination is enough** | S5-08's deep place can be a one-line "coming soon" view. Do not build a node list behind it; do not over-design the placeholder. Spike 6 replaces the view. |
| **Nodes are deliberately non-searchable** | Projecting them into the omnibar would mean a `KindNode`, a registry `KindSpec`, a projector, a `WorkspaceLocation` mapper, a `projection_version` bump, and `FakeStore` parity. Declare Nodes out of search for this spike and revisit when they carry names. |
| **Onboarding tests exercise the seed** | `sourceFixture` runs full onboarding, so a bug in `nodetypes.Install` fails tests across `api/ffi/handlers`, `core/onboarding`, and several `core/database` packages at once. |

---

## Accepted risk: an entire layer with no consumer

This was originally scoped as "`node_positions` ships with no UI exercising it." It is now broader: **`nodes`, `node_types`, and `node_positions` all ship with no UI exercising them.** Every RPC added in S5-06 is called only by Go tests and a `FakeStore` round-trip. Nothing a researcher can click reaches any of it.

This is deliberate, and it is the same argument as before, only larger. The alternative is opening Spike 6 with a migration, three Go packages, and a set of FFI methods before a single line of canvas code — exactly the "backend spike first" shape this plan exists to avoid.

Because the risk grew, the mitigation has to be stronger than it was:

- **Coverage is the deliverable, not a side effect.** Every RPC needs a Go handler test *and* `FakeStore` parity, because `FakeStore` is what Spike 6's Swift work develops against. A `FakeStore` that disagrees with the real store is the failure mode that will cost Spike 6 a day.
- **Round-trip, do not just write.** Assert reads after reopen, not only successful inserts — persistence across relaunch is precisely what no human will verify this spike.
- **The schema is the expensive mistake.** Go packages and RPCs are cheap to change in Spike 6; the migration is not, once a dogfood project exists. Spend the review effort on `000021.sql` and treat the rest as provisional. `node_positions` remains the cheapest of the three to migrate later, being unaudited UI state (§5.1).

---

## Forward compatibility: the family tree view

The canvas this spike leads up to is expected to be reused for an eventual family tree over Conclusion-layer canonical entities. Design note §13 works through what that implies; the short version for **this** spike is that it changes almost nothing, because two of the three shared pieces are already shared by the model docs:

- `canonical_entities.node_type_id` references the same `node_types` rows S5-03 seeds, so the tree inherits that vocabulary for free.
- Canonical refs and Node refs share one format, differing only in which of the type's two prefixes they are minted from (`PER-…` against `CPR-…`), so S5-01's guard serves both layers and a tree view needs no new ref work.

The one thing it did change is the name of the layout table. It is **`node_positions`**, not `graph_node_positions`: there is deliberately no `graphs` entity (design note §5.1), and a future tree will get its own `canonical_entity_positions` rather than a shared polymorphic table — a shared subject column could not carry a foreign key, which would forfeit the cascade behavior that is the main reason this table is shaped the way it is.

**Do not pre-generalize anything else here.** The reuse worth protecting lives in the canvas geometry, which is Spike 6's problem and is addressed there by putting those primitives in a neutral module rather than inside `Features/Interpretation/`.

---

## Suggested PR titles (why-focused)

| Step | Title sketch |
| --- | --- |
| S5-01 | Guard node type ref prefixes in `core/ref` |
| S5-02 | Add interpretation node schema and graph layout storage |
| S5-03 | Seed node type vocabulary at catalog create |
| S5-04 | Add audited node CRUD for the interpretation layer |
| S5-05 | Persist graph node positions outside the audit trail |
| S5-06 | Expose node types, nodes, and positions over FFI |
| S5-07 | Register the interpretation workspace section and its places |
| S5-08 | Add the interpretation sources list ahead of the canvas |
| S5-09 | Document the interpretation foundation and close spike 5 |

---

## Parallelism

| Track | Steps |
| --- | --- |
| **Design (no PRs)** | S5-D1 → S5-D2, from day one, parallel to the Go core |
| **Go core (critical path)** | S5-01 → S5-06 |
| **Mac client** | S5-07 → S5-08 → S5-09 |

S5-01 is genuinely independent and can land any time. Everything else is now a straight chain with no fork, because each layer is the next one's only consumer and the branch that used to exist — the nodes list running parallel to the entry points — is gone.

The design track is no longer a real schedule risk. Two small boards remain, both extending shipped surfaces (`PVList`, the existing sidebar) rather than inventing anything, and the Mac client work is two PRs instead of three. Start S5-D1 alongside S5-01 anyway, because it costs nothing to.

**The critical path is now almost entirely Go**, which is the main structural change from the previous plan: S5-02 through S5-06 is five sequential backend PRs with no client work able to overtake them.

---

## Definition of done

Jake can, on his MacBook:

1. Complete the in-app half of the dogfood bar above on a real project — reach the Interpretation section, open a Source, and land on the stub without surprise.
2. Run `CGO_ENABLED=1 go test -tags fts5 ./...` and see the test-only half of the bar covered, including a `FakeStore` round-trip for every new RPC.
3. Point at `node_types` / `nodes` / `node_positions` and the Interpretation places as the foundation Spike 6 builds on, with no schema, Go, or FFI work left in front of the canvas.
4. Confirm the layer is *honest*: the stub does not claim Nodes exist or can be created yet.

Per [`versioning.mdc`](../../../.cursor/rules/versioning.mdc), the docs in this folder do not bump `VERSION`; cutting a release that contains the spike bumps product PATCH across `VERSION`, `core/version.go`, and both Xcode `MARKETING_VERSION` configurations.

---

## What Spike 6 inherits

So the boundary is unambiguous — on the day Spike 6 opens, these exist and work:

- `nodes` rows with candidate refs, created and deleted through an audited Go package and an FFI method.
- A seeded Node Type vocabulary with both `ref_prefix` and `candidate_ref_prefix` values.
- A positions table and the RPCs to read and write it.
- A `.sourceGraph(project:sourceId:)` query key owning Nodes and their positions in one payload (design note §5.3), warmed by `session.apply(location:)`.
- A workspace section, root + deep places, the Sources list, a stub behind a Source, and working Back/Forward.
- A `FakeStore` that answers every new RPC, so Swift work can start without the dylib.

Spike 6's first PR is the `NSScrollView` bridge and a bubble — swapping the stub view for the canvas.

And these are explicitly **not** inherited — Spike 6 owns them, and they are unplanned on purpose so the canvas's results can inform them:

- The real canvas destination behind a Source row (replacing the stub).
- The Source-page **Open interpretation graph** button, and its design board.
- The canvas accessibility representation and keyboard parity, which design note §7.4 makes slice-2 work rather than polish.
