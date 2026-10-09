# Spike 9 — Conclusion performance ledger

Evidence for Conclusion read performance: where cost comes from, what the foundation absorbs, and what was measured. Decisions: [`README.md`](README.md). This file stays the reference for later tiers.

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
| H2 | Persons list | H1 per row, twice for places (birth, death) | Absorbed by the cache and set-based composition. List composition was not timed. |
| H3 | Events list | Event name needs subject Person's name; event place walks Locations → Places | Composed from cache lookups. Not timed. |
| H4 | Auto-reconciler evidence | Candidate provenance (Citation → Source → assessment, certainty, claim confidence) — extra joins per Observation | Joined into the candidate query (S9-14): query count per batch unchanged (`TestLoaderQueryCountIsConstant`). A credibility change recomputes every handle the Source backs (see H9); outcomes add one row per Observation to each handle's rewrite. |
| H5 | Swift invalidation | Bust-all on any write | Cheap while reloads read R3. Tier 3 if not. |
| H6 | Catalog session | Cross-Source reads and full rebuilds on one serialized session | Rebuild on open was not timed. |
| H7 | Promote | Target suggestions across all handles of a type; comparison rows (incoming × every member) | Suggestions via R3 `sort_key` + R8 index; comparison live for one handle. |
| H8 | Search reprojection | FTS documents composed across handles; header dependents must be reprojected in the write transaction | `RecomputeTx` reprojects the handle and `HeaderDependents`. A bridge delete snapshots those dependents first. Covered by rebuild-equals-upkeep. |
| H9 | R3 upkeep | Claim create / remove recomputes every Property of E plus inbound ends; credibility change touches every handle under a Source | Worst case (a credibility change on a large Source) was not timed. |

## Measurements

| Date | Fixture | Operation | Rows / handles | Time | Notes |
| --- | --- | --- | --- | --- | --- |
| 2026-10-05 | `BenchmarkReconcileNames` (pure Go, dev Mac) | Reconcile one handle's names through the pipeline | 200 name candidates | ~0.58 ms | 7.7k allocs. Per-type units, folding and survivor grouping are quadratic in distinct values per type, not in candidates; fine at this size. |
| 2026-10-05 | `BenchmarkReconcileNames` after one-name combining | Same | 200 name candidates | ~0.87 ms | 10k allocs. Spelling checks on majority losers and per-type combining add work; still well under a write's budget. |

**Deep fixture: not taken.** A seeded project of a few hundred handles (a Person on about ten Sources, multi-member events with Locations, a place hierarchy, a Promote proposal and Done) was descoped at spike close. Those timings were not recorded. The two rows above do not justify a second derived tier. The reserve tiers stay unused.
