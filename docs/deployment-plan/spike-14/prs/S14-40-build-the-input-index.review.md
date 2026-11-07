# Review — S14-40 Build the input index

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

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

## Round 2

Both reviewers confirmed every round-1 answer against the code; no Current state row was false (`26ada55` → HEAD changes no `core` file). New findings, all Fixed.

### M-8 · blocker · The test recorder derives the key rules a second time
**Answer:** Fixed — D-2: unexported `nodeKey` and `followKeys` in `inputs.go`, used by the scope and by the recorder (D-12).

### M-9 · blocker · D-9's end rule differs from every existing reader
**Answer:** Fixed — D-9 takes `graphcache.Rank1(n.Kept(prop))`, as `canonlinks.EdgeOf`, `linkLoadSQL`, `endpointIDs` and `AssociationEndpoints` do; it takes ids and reads the vocabulary once per call; T-1 adds a rank-2 kept end that gives no key.

### M-10 · blocker · D-5 attributed a row's follows a second way
**Answer:** Fixed — D-5 keeps `eventPlaceGraph(g, v, eventIDs)` and passes the pick ids; each person row records `Follow(pick, PlacesOfEvent)` for every birth and death pick through its own scope.

### Minor · rule wording "reads only the row's own handle"
**Answer:** Fixed — STEP-8 says "reads no other handle".

### X-7 · blocker · T-9 cost
**Answer:** Fixed — one `graphcache.Over(db)` graph with the recorder reset per row (the probe ran all rows this way in 2.3 s).

### X-8 · major · Memory of one graph per call
**Answer:** Fixed — D-11: at a full rebuild it holds every handle's node, what the open catalog's graph holds after its warm; dropped when the call returns.

### X-9 · major · AC-3.2 / T-11 / STEP-1 named different calls
**Answer:** Fixed — T-11 logs the first, untimed `f.PlaceRel` once; the budgeted second call stays; the Stop condition is "not under 0.8 × the STEP-1 figure".

### X-10 · major · T-5 fixture
**Answer:** Fixed — promoted with `promotetest.Save`; `ClearAll` before `ReprojectHandles` so the upkeep that already ran cannot have written the rows.

### X-11 · major · `follow` helper
**Answer:** Fixed — D-3: `follow` takes a scope (nil-safe); batch phases pass `g.NoInputs()`; every `g.Node` and `follow` in the five functions goes through the scope.

### X-12 · minor · D-4 scopes, D-9 vocabulary, T-8 ids, T-6 premise
**Answer:** Fixed — D-4 keeps one scope per place across both loops and notes `follow` records the association, so the `linked` dedupe loses nothing; D-9 names `g.Vocabulary()`; T-8 uses T-9's query; T-6 states why a person always has a key. Current state: the place-chain functions are in `compose.go`; `placeNeighborhood` is a fourth `buildPlaceGraph` caller, unchanged.

### X-13 · minor · STEP-8 `rg` could not fail
**Answer:** Fixed — the map's S14-40 entry no longer says `Graph.Node` / `Graph.Follow`; STEP-8's `rg` looks for "through `Graph.(Node|Follow)`" in the spike folder, skills and rules.
