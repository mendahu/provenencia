# Spike 9 — Conclusion performance ledger

Evidence for Conclusion read performance: where cost comes from, what the foundation absorbs, and what was measured. Plan: [`deployment-plan.md`](deployment-plan.md#architecture). At spike close this file moves with the archive and stays the reference for later tiers.

## How to add

Add a row to **Hotspots** when a requirement or PR adds a walk, a join, or a trigger. Add a row to **Measurements** whenever a fixture is timed. Keep entries short.

## Why it scales badly without a cache

Every Property on a handle is multi-valued (one list per member, possibly several per member). Derived values walk to other handles, which are multi-valued in turn.

```text
Person (m members)
  name                 → m × name Observations                  → auto-reconciler
  birth date / place   → Participation(subject) → Event (e members) → date Observations → auto-reconciler
                                                  → Location → Place (p members) → toponyms → auto-reconciler
  death date / place   → same again
  marriage, relationships, … → one more walk each
```

Computed live, a Person on ten Sources reads hundreds of Observations plus Citation → Source joins for ranking, and a list pays that per row.

## Adopted strategy

**One derived cache: auto-reconciled values** (R3). The auto-reconciler runs at write time, one hop from each write, and stores every auto-reconciled value per (handle, Property). Association ends are auto-reconciled entity-valued Properties, so the canonical graph is the same table with a reverse index. Screens compose from it in Go (R4). Search documents are composed from the same headers (R8).

**Tiers held in reserve** — add only when measurements below say so:

1. **Derived-values table** (life dates, event names, places) if composing list rows is slow. Reintroduces multi-hop dependencies; avoid unless needed.
2. **Resident in-memory copy** of auto-reconciled values for the session if SQLite lookups on R3 dominate.
3. **Narrower Swift invalidation** using R3's affected-handles set, if reloads stop being cheap.

## Hotspots

| # | Where | Cost driver | Status |
| --- | --- | --- | --- |
| H1 | Person detail | Fan-out: each Property × members; each derived value walks another handle's members | Absorbed by R3; detail composes from cached clusters. Observation drill-down stays live for one handle. |
| H2 | Persons list | H1 per row, twice for places (birth, death) | Absorbed by R3 + set-based R4 composition. List of the deep fixture is inside the 9.4 ms three-list row. |
| H3 | Events list | Event name needs subject Person's name; event place walks Locations → Places | Composed from R3 lookups. Same 9.4 ms row. |
| H4 | Auto-reconciler evidence | Candidate provenance (Citation → Source → assessment, certainty, claim confidence) — extra joins per Observation | Joined into the candidate query (S9-14): query count per batch unchanged (`TestLoaderQueryCountIsConstant`). A credibility change recomputes every handle the Source backs (see H9); outcomes add one row per Observation to each handle's rewrite. |
| H5 | Swift invalidation | Bust-all on any write | Cheap while reloads read R3. Tier 3 if not. |
| H6 | Catalog session | Cross-Source reads and full rebuilds on one serialized session | Full rebuild of the deep fixture: 161 ms. |
| H7 | Promote | Target suggestions across all handles of a type; comparison rows (incoming × every member) | Proposal 6.7–8.2 ms; batch Done 2.3–2.8 ms on the deep fixture. |
| H8 | Search reprojection | FTS documents composed across handles; header dependents must be reprojected in the write transaction | Fixed reverse walk over R3's reverse index; covered by rebuild-equals-upkeep test. |
| H9 | R3 upkeep | Claim create / remove recomputes every Property of E plus inbound ends; credibility change touches every handle under a Source | One name rewrite on the deep fixture: 1.4 ms. A credibility change on a large Source is still unmeasured. |

## Measurements

| Date | Fixture | Operation | Rows / handles | Time | Notes |
| --- | --- | --- | --- | --- | --- |
| 2026-10-05 | `BenchmarkReconcileNames` (pure Go, dev Mac) | Reconcile one handle's names through the pipeline | 200 name candidates | ~0.58 ms | 7.7k allocs. Per-type units, folding and survivor grouping are quadratic in distinct values per type, not in candidates; fine at this size. |
| 2026-10-05 | `BenchmarkReconcileNames` after one-name combining | Same | 200 name candidates | ~0.87 ms | 10k allocs. Spelling checks on majority losers and per-type combining add work; still well under a write's budget. |
| 2026-10-08 | `deepfixture` (dev Mac, `-benchtime=1x`) | Full auto-reconciler rebuild | 475 handles | 161 ms | One Person on 10 Sources; multi-member birth, death, marriage with a Location; two relationships; a six-level place chain; 150 filler people each with a birth. |
| 2026-10-08 | same | One Observation write's upkeep | 475 handles | 1.4 ms | Rewrite of the focal Person's name. |
| 2026-10-08 | same, obituary Source (6 people, 4 events, unfiled) | Promote proposal | 475 handles in catalog | 8.2 ms | `promotealign.Propose` on the small Source. |
| 2026-10-08 | same, a register already filed | Promote proposal | 475 handles | 6.7 ms | Proposal against the deep catalog. |
| 2026-10-08 | obituary Source | Batch Done | 10 new handles | 2.8 ms | `SaveBatch` of every primary Subject as New. Proposal not included. |
| 2026-10-08 | deep catalog, a fresh 10-Subject Source | Batch Done | 475 handles already filed | 2.3 ms | Same Done shape, catalog already at fixture scale. |
| 2026-10-08 | same | List composition | persons + events + places | 9.4 ms | `ListPersons`, `ListEvents`, and `ListPlaces` together. |
| 2026-10-08 | same | One detail | focal Person, 10 members | 0.41 ms | `conclusiondetails.ForEntity`. |
| 2026-10-08 | same | Place-chain composition | 6-level chain | 2.1 ms | `ParentsAtDate` from the deepest Place. |

**Fixture (required this spike):** a seeded project where a Person sits on ~10 Sources; birth, death, and marriage events each have several members; each has one or more Locations; a few relationships. Scale it to a few hundred handles. Time: full rebuild, one Observation write's upkeep, one Promote step, each list's composition, one detail's composition.
