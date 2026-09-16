# Deployment Plan — Spike 5

Interpretation foundation: candidate refs, Node vocabulary, Nodes, layout storage, FFI, and a workspace place. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §11.1. Authoritative schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §4.

## Status

**Planned.** No steps landed. Landings go in [`completed.md`](completed.md).

## Goal (dogfood bar)

1. Source page → **Open interpretation graph** lands on a Source-scoped destination.
2. Create a person, an event, and a place there; each gets a `{PREFIX}-C-{TOKEN}` ref.
3. Rename and delete a Node; both appear in the audit log with a `node` entity type.
4. Sidebar → **Interpretation** lists Sources; picking one opens its graph.
5. Back/Forward and relaunch restore the graph place; returning to it is a session-cache hit.
6. A Source deleted while its graph is in history falls back to the Interpretation list, not a stale page.

## PR sequence

```text
S5-01  Candidate refs (core/ref)                  no deps
  │
  ▼
S5-02  Migration 000021                           node_types, nodes, node_positions
  │
  ▼
S5-03  core/database/nodetypes + seed Install     7 seeded types, create-time only
  │
  ▼
S5-04  core/database/nodes + audit                needs S5-01 (mint) and S5-03 (prefix)
  │
  ▼
S5-05  core/database/graphlayout                  unaudited positions
  │
  ▼
S5-06  Proto + dispatch + handlers                the FFI seam; codegen Go *and* Swift
  │
  ▼
S5-07  Section, places, query keys, store         registry plumbing; stub destination views
  │
  ├──────────────────────┐
  ▼                      ▼
S5-08                  S5-09
Entry points           Source nodes list
(list + page button)   (replaces the stub)
  │                      │
  └──────────┬───────────┘
             ▼
          S5-10  Docs, dogfood, cleanup
```

---

## Checklist

- [ ] S5-01 — Candidate ref minting in `core/ref`
- [ ] S5-02 — Interpretation schema migration
- [ ] S5-03 — Node type vocabulary and create-time seed
- [ ] S5-04 — Node CRUD with audit
- [ ] S5-05 — Graph layout positions
- [ ] S5-06 — FFI methods for node types, nodes, and positions
- [ ] S5-07 — Interpretation workspace section and place registry
- [ ] S5-08 — Interpretation Sources list and Source-page entry point
- [ ] S5-09 — Source nodes list destination
- [ ] S5-10 — Docs, dogfood, cleanup

---

## Schema (S5-02)

One migration, `core/database/migrations/000021.sql`. Two tables are transcribed from the interpretation data model §4.1–4.2 and must not drift from it; the third is new in this spike and specified in the design note §5.1.

```sql
CREATE TABLE node_types (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	description TEXT,
	ref_prefix  TEXT NOT NULL UNIQUE,

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

- **`nodes` has no `ON DELETE` clause** on either FK — deliberately `NO ACTION`, so a Source with Nodes cannot be deleted out from under them. The cascade for `observations` → `nodes` is a *later* decision and must be made when that table ships (§4.3), not retrofitted.
- **`node_positions` cascades both ways**, which is the deliberate contrast: layout is disposable, evidence is not.
- **No `UNIQUE (source_id, grid_x, grid_y)`.** A drag that swaps two bubbles transiently collides; let the UI nudge (§5.1).
- **Coordinates are signed** — the plane is unbounded around an origin, so no `CHECK (grid_x >= 0)`.
- **`node_positions` is unaudited.** State this in the migration comment and be ready to defend it, because the nearest sibling (`source_metadata_layout`, migration `000013`) *is* audited. The distinction: dismissing a metadata suggestion is a research decision; arranging bubbles is not.

## Seeded Node Types (S5-03)

All seven from [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1, in a code registry installed at create time only — never in a migration, never healed on open:

```text
person PER    event EVT    place PLC    relationship REL
participation PTN    location LOC    source SRN
```

Seed all seven even though only three are placeable and two categories (bridge, reification) have no UI this spike. They are rows, not features; a partial seed just means editing the registry again later, and the authoritative doc already fixes the set.

---

## Scope boundary: what "ends at the canvas" means

The one judgment call in this plan is **S5-09**, the deep destination. Three options were on the table:

| Option | Verdict |
| --- | --- |
| Render a "coming soon" placeholder | Rejected. Proves only that navigation works, and throws away the chance to validate the data path before canvas code sits on top of it. |
| Render the canvas | Rejected — that is Spike 6, and the whole point of this spike is that Spike 6 starts clean. |
| **Render a plain list of the Source's Nodes** | **Taken.** Proves migration → Go → FFI → store → session cache → view end to end, with zero canvas risk. |

The list is not throwaway. Design note §7.4 already commits to a **structured, non-canvas editing path** as both the accessibility representation of the graph and the only part of this surface that XCUITest can drive. This is that path, arriving early because it is also the cheapest way to prove the rail.

---

## Key decisions and gotchas

Findings from the pattern inventory that will otherwise cost a day each.

| Area | The trap |
| --- | --- |
| **Schema hash** | `core/database/schemahash.go` computes `expectedSchemaHash` at init by migrating an in-memory DB. There is **no committed golden constant to bump** — counterintuitive if you expect a golden file. Nothing to regenerate. |
| **`Open` rejects unknown schema** | The digest covers all of `sqlite_schema`; `catalog_test.go` proves even a rogue *index* makes a catalog unopenable. Nothing may be created at runtime — every table and index ships in `000021.sql`. |
| **Test tags** | `CGO_ENABLED=1 go test -tags fts5 ./...`. Bare `go test` on `core/database` fails confusingly, because migrations 18–19 create FTS5 virtual tables. |
| **Do not widen `ref.Valid`** | Five call sites depend on today's strict `AAA-TTTTT` semantics (`users`, `sources`, `artifacts`, `identity`, `search/refpath`). Add `MintCandidate` / `ValidateCandidate` as separate functions — the name `MintCandidate` is already reserved by [`catalog-refs.md`](../../catalog-refs.md) §4 and the `add-catalog-ref` skill. |
| **Reserved prefixes** | The guard must reject `USR`, `SRC`, `ART`, `CIT`, `OBS` **and the literal `C`**. No such helper exists today. |
| **`ref_prefix` collisions are invisible to upsert** | `ON CONFLICT (key, origin)` does not catch a duplicate `ref_prefix`; it arrives as a raw constraint error and will surface as `internal.unknown` unless mapped to its own code. |
| **Node refs need a join inside the transaction** | Unlike `sources` (constant `SRC`), minting a Node ref means reading `node_types.ref_prefix` first. Extend the `requireType(tx, …)` existence check to return the prefix, then reuse the 8-attempt retry loop from `sources.go`. |
| **Audit needs no registry entry** | `EntityType` / `ActionType` are unvalidated free-form strings — which also means a typo persists silently. Convention: singular snake_case table name, `{verb}_{entity}`. |
| **Protobuf codegen is manual and dual-target** | `scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`, and writes **both** `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Commit both; forgetting the Swift half breaks the app build. Methods start at 45 (`METHOD_SEARCH_CATALOG = 44` is current highest). |
| **`?? .sourcesList` hides a missing spec** | `WorkspaceDestinationHost.presentation(for:)` falls back silently, so a new section with no `PlaceRegistry` spec renders the ordinary Sources list while the sidebar shows Interpretation selected. Land the specs and the sidebar case in the same PR (S5-07). |
| **Two hard test gates** | `PlaceRegistryTests.registryCoversAllPlaceIDs` switches exhaustively over `PlaceID`, and `WorkspaceDestinationHostTests` asserts `known.count == 4` over `WorkspacePresentationID`. Both fail the moment a case is added. |
| **History is not forward-compatible** | `WorkspaceSection.init(from:)` throws on an unknown raw value, so history written by a build with `interpretation` fails to decode on a build without it (surfacing as `NavigationHistoryIssue.loadFailed`). Acceptable for a solo dogfood, worth knowing before bisecting. |
| **The Source page cannot navigate today** | Nothing in `Features/Sources/` except `SourcesListView` holds `@Environment(WorkspaceNavigation.self)`. S5-08 must inject it. |
| **Nodes are deliberately non-searchable** | Projecting them into the omnibar would mean a `KindNode`, a registry `KindSpec`, a projector, a `WorkspaceLocation` mapper, a `projection_version` bump, and `FakeStore` parity. Declare Nodes out of search for this spike and revisit when they carry names. |
| **Onboarding tests exercise the seed** | `sourceFixture` runs full onboarding, so a bug in `nodetypes.Install` fails tests across `api/ffi/handlers`, `core/onboarding`, and several `core/database` packages at once. |

---

## Accepted risk: an API with no consumer

`node_positions` and its RPCs (S5-05, part of S5-06) ship with **no UI exercising them** — the nodes list does not place anything on a grid. They are covered by Go tests and a `FakeStore` round-trip only.

This is deliberate. The alternative is opening Spike 6 with a migration, a Go package, and an FFI method before a single line of canvas code, which is exactly the "backend spike first" shape this plan exists to avoid. The cost of being wrong is low: the table is unaudited UI state and therefore cheap to migrate (§5.1).

---

## Forward compatibility: the family tree view

The canvas this spike leads up to is expected to be reused for an eventual family tree over Conclusion-layer canonical entities. Design note §13 works through what that implies; the short version for **this** spike is that it changes almost nothing, because two of the three shared pieces are already shared by the model docs:

- `canonical_entities.node_type_id` references the same `node_types` rows S5-03 seeds, so the tree inherits that vocabulary for free.
- Canonical refs are `{prefix}-{token}` and Node refs are `{prefix}-C-{token}` off the same `ref_prefix`, so S5-01 serves both layers. The reserved-prefix guard protects both.

The one thing it did change is the name of the layout table. It is **`node_positions`**, not `graph_node_positions`: there is deliberately no `graphs` entity (design note §5.1), and a future tree will get its own `canonical_entity_positions` rather than a shared polymorphic table — a shared subject column could not carry a foreign key, which would forfeit the cascade behavior that is the main reason this table is shaped the way it is.

**Do not pre-generalize anything else here.** The reuse worth protecting lives in the canvas geometry, which is Spike 6's problem and is addressed there by putting those primitives in a neutral module rather than inside `Features/Interpretation/`.

---

## Suggested PR titles (why-focused)

| Step | Title sketch |
| --- | --- |
| S5-01 | Mint candidate refs for interpretation nodes |
| S5-02 | Add interpretation node schema and graph layout storage |
| S5-03 | Seed node type vocabulary at catalog create |
| S5-04 | Add audited node CRUD for the interpretation layer |
| S5-05 | Persist graph node positions outside the audit trail |
| S5-06 | Expose node types, nodes, and positions over FFI |
| S5-07 | Register the interpretation workspace section and its places |
| S5-08 | Open a source's interpretation graph from the list and the source page |
| S5-09 | List and edit a source's nodes before the canvas exists |
| S5-10 | Document the interpretation foundation and close spike 5 |

---

## Parallelism

| Track | Steps |
| --- | --- |
| **Go core (critical path)** | S5-01 → S5-06 |
| **Mac client** | S5-07 → S5-08 / S5-09 (parallel) → S5-10 |

S5-01 is genuinely independent and can land any time. S5-08 and S5-09 are the only true fork; everything else is a chain, because each layer is the next one's only consumer.

No design-board dependency — there is no novel visual language in this spike. `PVList`, `PVButton`, and the existing sidebar carry all of it.

---

## Definition of done

Jake can, on his MacBook:

1. Complete the dogfood bar above end to end on a real project.
2. Point at `node_types` / `nodes` / `node_positions` and the Interpretation place as the foundation Spike 6 builds on, with no schema, Go, or FFI work left in front of the canvas.
3. Confirm the layer is *inert but honest*: Nodes exist, carry candidate refs, and are audited, and nothing yet claims they are cited.

Per [`versioning.mdc`](../../../.cursor/rules/versioning.mdc), the docs in this folder do not bump `VERSION`; cutting a release that contains the spike bumps product PATCH across `VERSION`, `core/version.go`, and both Xcode `MARKETING_VERSION` configurations.

---

## What Spike 6 inherits

So the boundary is unambiguous — on the day Spike 6 opens, these exist and work:

- `nodes` rows with candidate refs, created and deleted through an audited Go package and an FFI method.
- A seeded Node Type vocabulary with `ref_prefix` values.
- A positions table and the RPCs to read and write it.
- A `.sourceGraph(project:sourceId:)` query key owning Nodes and their positions in one payload (design note §5.3), warmed by `session.apply(location:)`.
- A workspace section, two registered places, working Back/Forward, and two entry points.
- A structured node list that doubles as the accessibility representation and the XCUITest-drivable path.

Spike 6's first PR is the `NSScrollView` bridge and a bubble.
