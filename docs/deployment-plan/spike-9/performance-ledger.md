# Spike 9 — Conclusion performance ledger

Evidence for Conclusion read performance: where cost comes from, what the foundation absorbs, and what was measured. Plan: [`deployment-plan.md`](deployment-plan.md#architecture). At spike close this file moves with the archive and stays the reference for later tiers.

## How to add

Add a row to **Hotspots** when a requirement or PR adds a walk, a join, or a trigger. Add a row to **Measurements** whenever a fixture is timed. Keep entries short.

## Why it scales badly without a cache

Every Property on a handle is multi-valued (one list per member, possibly several per member). Derived values walk to other handles, which are multi-valued in turn.

```text
Person (m members)
  name                 → m × name Observations                  → resolver
  birth date / place   → Participation(subject) → Event (e members) → date Observations → resolver
                                                  → Location → Place (p members) → toponyms → resolver
  death date / place   → same again
  marriage, relationships, … → one more walk each
```

Computed live, a Person on ten Sources reads hundreds of Observations plus Citation → Source joins for ranking, and a list pays that per row.

## Adopted strategy

**One derived cache: resolved values** (R3). The resolver runs at write time, one hop from each write, and stores every ranked cluster per (handle, Property). Association ends are resolved entity-valued Properties, so the canonical graph is the same table with a reverse index. Screens compose from it in Go (R4). Search documents are composed from the same headers (R8).

**Tiers held in reserve** — add only when measurements below say so:

1. **Derived-values table** (life dates, event names, places) if composing list rows is slow. Reintroduces multi-hop dependencies; avoid unless needed.
2. **Resident in-memory copy** of resolved values for the session if SQLite lookups on R3 dominate.
3. **Narrower Swift invalidation** using R3's affected-handles set, if reloads stop being cheap.

## Hotspots

| # | Where | Cost driver | Status |
| --- | --- | --- | --- |
| H1 | Person detail | Fan-out: each Property × members; each derived value walks another handle's members | Absorbed by R3; detail composes from cached clusters. Observation drill-down stays live for one handle. |
| H2 | Persons list | H1 per row, twice for places (birth, death) | Absorbed by R3 + set-based R4 composition. Measure list composition. |
| H3 | Events list | Event name needs subject Person's name; event place walks Locations → Places | Composed from R3 lookups. Measure. |
| H4 | Resolver ranking | Candidate provenance (Citation → Source → assessment, certainty, claim confidence) — extra joins per Observation | Paid at write / rebuild time by the batched loader. Measure single-write upkeep. |
| H5 | Swift invalidation | Bust-all on any write | Cheap while reloads read R3. Tier 3 if not. |
| H6 | Catalog session | Cross-Source reads and full rebuilds on one serialized session | Measure rebuild on open. |
| H7 | Promote | Target suggestions across all handles of a type; comparison rows (incoming × every member) | Suggestions via R3 `sort_key` + R8 index; comparison live for one handle. |
| H8 | Search reprojection | FTS documents composed across handles; header dependents must be reprojected in the write transaction | Fixed reverse walk over R3's reverse index; covered by rebuild-equals-upkeep test. |
| H9 | R3 upkeep | Claim create / remove recomputes every Property of E plus inbound ends; credibility change touches every handle under a Source | Measure worst case (credibility change on a large Source). |

## Measurements

| Date | Fixture | Operation | Rows / handles | Time | Notes |
| --- | --- | --- | --- | --- | --- |
| 2026-10-05 | `BenchmarkReconcileNames` (pure Go, dev Mac) | Reconcile one handle's names through the pipeline | 200 name candidates | ~0.58 ms | 7.7k allocs. Per-type units, folding and survivor grouping are quadratic in distinct values per type, not in candidates; fine at this size. |
| 2026-10-05 | `BenchmarkReconcileNames` after one-name combining | Same | 200 name candidates | ~0.87 ms | 10k allocs. Spelling checks on majority losers and per-type combining add work; still well under a write's budget. |

**Fixture (required this spike):** a seeded project where a Person sits on ~10 Sources; birth, death, and marriage events each have several members; each has one or more Locations; a few relationships. Scale it to a few hundred handles. Time: full rebuild, one Observation write's upkeep, one Promote step, each list's composition, one detail's composition.
