# Review — S14-40 Build the input index

**Verdict:** pending round 2
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None.

## Writer's own calls

- Replay through per-place scopes in the batch place graph, not per-row place graphs: per-row graphs measured 4× slower (persons 95 → 386 ms, events 125 → 533 ms).
- A person's place graph from its birth and death events only (D-5, from M-3): rows read the same; the index stops making a person depend on a marriage's place.
- No version bump (S14-10 precedent).
- `budgetComposeAll` = 500 ms against 220 ms measured; the first part_of from the city is logged, not budgeted.
- The `Dependents` check is a superset check (D-13), not the subset the map's first wording had.
- The `graphcache/sql.go` placeholder cleanup is handed back: `core/database` imports `graphcache`.

## Round 1

### M-1 · blocker · `closure` writes by hand the keys `buildPlaceGraph` reads
**Answer:** Fixed — D-4: `buildPlaceGraph` reads each place through a per-place `Inputs(nil)` scope and stores its keys; `closure` only unions them.

### M-2 · blocker · A fourth hand-written association-ends reader, in raw SQL, against `catalog-queries.mdc`
**Answer:** Fixed — D-9: `graphcache.Graph.AssociationInputs` reads the loaded node through `connectrules`; D-8 calls it; no `typeKey` field, no SQL.

### M-3 · blocker · Person rows record the places of every subject event
**Answer:** Fixed — D-5: picks first, place graph from birth and death events, as `PersonLifeOwners`; Out row removed; AC-1.2 and T-2 pin it.

### M-4 · major · A rule line beside one it contradicts; docs describe recording on `Graph`
**Answer:** Fixed — STEP-8 amends `catalog-queries.mdc` line 16; the working rule and the design doc now name `InputScope.Node` / `Follow`; STEP-8 `rg` checks no doc says `Graph.Node(H)`.

### M-5 · minor · Package-global recorder
**Answer:** Fixed — D-12: per-graph `SetInputRecorderForTest`.

### M-6 · minor · `InputScope.Graph()` escape hatch
**Answer:** Fixed — D-2: no accessor.

### M-7 · nit · Table naming
**Answer:** Fixed — `catalog_search_inputs`.

### X-1 · blocker · Plumbing of per-row keys
**Answer:** Fixed — D-6 gives the return types, the callers that discard them, and how `RowsForSearch` fills `Summaries.inputs` without changing `NewSummaries`.

### X-2 · blocker · T-7 equality may not hold
**Answer:** Fixed — Edges "Upkeep reach" says why a T-7 failure means an invalidation miss today; a Stop condition sends it to S14-42 rather than widening reprojection here.

### X-3 · major · Person replay's event list
**Answer:** Fixed — D-3 and D-5: scopes created up front; `lifePicks` runs in the batch phase with the row's scope; `eventPlaceGraph` follows each pick through it; events repeat their follows through `oneEvent`.

### X-4 · major · Association-end edges; `reason` column
**Answer:** Fixed — `reason` exists (`000039.sql`, `000040.sql`), but D-9 reads the loaded node instead. Merged or non-bridge associations give nil; two ends on one Property give two keys; an end with no kept value gives none.

### X-5 · major · Test fixtures
**Answer:** Fixed — T-4 uses random ids and states the zero-id case; T-5 deletes through Unpromote and merges by SQL; the test-package choice is in D-11; T-9 / T-10 name their id query and the graph.

### X-6 · minor · Step details
**Answer:** Fixed — STEP-1 wraps with `since`; T-11 is the first part_of check; the design doc section name; `inputRows` takes `autoreconciler.Querier`; D-9's seam has no test callers to update (none call `reprojectHandleBatch`); D-4 says `rememberPlaceLink` keeps every link whatever its dates. Current state: `EnsureCatalog` line 638; fixtures are `evidencelayertest.Records`; `core/hops`.
