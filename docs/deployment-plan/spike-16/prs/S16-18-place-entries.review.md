# Review — S16-18 Place entries

**Verdict:** Fits with changes — Ready with changes, all findings answered.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- Answered 2026-11-13. M-1 (Apply over-drops: `roots ∪ SummaryDeps` drops 20,000 unchanged part entries when a hub is renamed, bending "updates exactly"). The user asked for the robust path, not a band-aid, and chose to fold a registry-level fix into this PR: keyed entries declare `Deps` and `Bridges`, and the registry keeps the dependency index (D-5, D-16, D-17).
- Answered 2026-11-13, before review. One chain builder: `buildPlaceGraph` reads the entries in this PR (D-7); the map's S16-18 / S16-19 split changed.

## Writer's own calls

- D-2: the entry always fills by the narrow load, even when the node is cached (one fill path; the first Places read after open costs one more query round, measured).
- D-4: an entry records exactly today's per-place keys (values, `part_of` end, those associations), none for parts or succession.
- D-10: `InputScope.Add` reports to the test recorder; D-18 adds the independent guard.
- D-17: `OnCommit` hands the registry link moves from data it already reads (no new query); merges count as moves of every link of the merged handle.
- Out: `canonlinks.kinship` adopting `Deps` goes to the structural-debt idea file, not this PR.

## Round 1

### X-1 · blocker · T-4 "missing place" and AC-3.2 cannot be built as written
`linkLoadSQL` joins the neighbour entity, so no link names a missing place; only an explicit seed does.
**Answer:** Fixed — AC-3.2 and T-4 read an unknown id as a seed (`ParentsAtDate`) and check the stored entry through `PlaceEntryForTest`.

### X-2 · major · AC-2.2 assumes the new toponym shows first
**Answer:** Fixed — AC-2.2 and T-4 assert the City entry drops and the chain's first name equals City's current `Names[0]`, which contains "Big City".

### X-3 · major · `found: false` lookups, dedupe and a missing association unstated
**Answer:** Fixed — D-7 (a `found: false` entry reads as absent), D-1 (lists deduped as `rememberPlaceLink` does; self-loop as today), D-3 (missing association errors as today).

### X-4 · minor · Test package for `placeentries_test.go`
**Answer:** Fixed — STEP-1 and STEP-6 say `package conclusionsummaries_test` (an internal test importing `graphcachetest` would cycle).

### X-5 · minor · Test mechanics left open
**Answer:** Fixed — T-4 names the delete call and the `placeRel` return; the location case moved to T-11 on `Canon`; D-15 says where the hub sits in the chain.

### X-6 · minor · Implicit items
**Answer:** Fixed — D-12 takes `v` from `vocabulary(g)`; STEP-9 updates the map rule that still named `nodeKey` / `followKeys`. D-13 was already accepted.

### M-1 · blocker · Apply drops entries no commit changed; the bent rule is not listed
**Answer:** To user, then Fixed per the user's answer (2026-11-13) — D-5, D-16, D-17: no Apply; `Deps` against `Handles` and `Bridges` against committed and merged links, kept by the registry. `SummaryDeps` is no longer used.

### M-2 · blocker · `NodesOnBridge` builds a second `Node` beside `loadChunk`
**Answer:** Fixed — D-2: `loadChunk` splits into `readNodes` (the one node literal) and its install loop; STEP-2's check finds one literal.

### M-3 · major · D-10 makes the isolation oracle check entry keys against themselves
**Answer:** Fixed — D-18 and T-8 compare each entry's inputs with a real `Inputs(nil)` scope's reads; §11 no longer claims T-5 catches D-4 drift.

### M-4 · major · D-8 trims lists no reader looks at
**Answer:** Fixed — D-8: `placeStep(c, id)` holds whole entries; the hop parameter goes.

### M-5 · major · `bridge == ""` means opposite things
**Answer:** Fixed — `NodesOnBridges(ids, bridges ...string)` and `bridges []string` down the chain; `bridgeKeys()` names every bridge.

### M-6 · major · "what a commit names" written a second time
**Answer:** Fixed by M-1's answer — no entry drops by `roots ∪ SummaryDeps` any more.

### M-7 · minor · Four per-direction fields instead of links by hop
**Answer:** Declined — `placeGraph`'s methods read by direction (`ImmediateParents` reads `partFrom`, `ImmediateParts` `partTo`), and D-8's hop mapping is gone; S16-21 can add an accessor when it reads one hop.

### M-8 · minor · S16-21 would make Promote import the summary package
**Answer:** Fixed (noted) — Next adds names the new production import for S16-21 to decide.

### M-9 · nit · No `Canon` builder for place links
**Answer:** Fixed — `Canon.PartOf` in STEP-1.

## Round 2

Maintainability: every round-1 answer holds; the D-16 / D-17 premises hold at `44ac023` (committed association rows give bridge and ends with no new query; a removed association reaches the drop through `Handles` alone; `mergedNeighbors`' `linkRaw`s carry `bridgeKey`).

### M-10 · blocker · `Deps` and `Bridges` are a second way to say what a value was built from, beside `InputKey`
`InputKey` already means "a handle's values" (`{ID}`) and "the links at this end of this bridge" (`{ID, Bridge, Endpoint}`); the persisted input index and the summary closure use it. Proposed: one `KeyedEntry.Inputs func(id, v) []InputKey`, an index on `(ID, Bridge)`, and the commit's moved inputs as `InputKey`s.
**Answer:** Fixed — D-5, D-16, D-17, US-4, T-8 – T-10, §6 and STEP-3 / 7 / 9 now use `Inputs`; `placeDeps`, `Bridges` and `linkMove` are gone. The place entry declares `placeInputs` (its row-facing `inputs` plus `FollowKeys` of the other three place hops), and T-8 guards both lists against real reads. Matching is on `(ID, Bridge)`; `Endpoint` is not matched (D-16 says why). The user's answer (exact update in the registry, not a band-aid) is unchanged; only the vocabulary changed.

### M-11 · minor · `Deps` means the opposite of `SummaryDeps`
**Answer:** Fixed by M-10 (the field is `Inputs`).

Executability (dry run of `44ac023`, before M-10's rename; each finding is answered against the `Inputs` shape): every round-1 answer holds; every Current state row is true.

### X-7 · major · Registry internals left open
**Answer:** Fixed — D-16 names the wiring: an `entry.dropMoved` closure built in `RegisterKeyed`, `keyedValues.inputs` for unindexing in `store`, `evict`, `Drop`, `Clear`, and the order in `applyEntries` (drop when held, then Apply when set and still held).

### X-8 · major · A removed link moves no bridge
**Answer:** Fixed — with `Inputs` the removal moves `{assoc}` through `Handles`; D-16 and D-17 now say an entry that reads a link declares its association's key (as `FollowKeys` does), and T-10 gains a "remove the location" case.

### X-9 · major · STEP-5 too wide; the fill untested until after it
**Answer:** Fixed — T-8 and the test hooks move into STEP-4; the chain rewrite splits into STEP-5 (chains and the Places list) and STEP-6 (the place page). Ten steps.

### X-10 · minor · Test helper shapes
**Answer:** Fixed — `InputIndexSizeForTest[T]` signature and unit, `PlaceEntryForTest`'s return tuple, `SnapshotRegistryForTest` for T-9 / T-10, and how T-9 commits.

### X-11 · minor · `Commit` ids in AC-2.4
**Answer:** Fixed — (a) `Location(Event("birth"), city)`, `Commit(loc)`; (b) `Commit(link)`; (c) `Commit(region, into)`.

### X-12 · minor · Stale STEP reference
**Answer:** Fixed — the rules-bent line names STEP-10.

### X-13 · minor · `NodesOnBridges` edges
**Answer:** Fixed — D-2: duplicates read once, an unregistered bridge key is an error, no bridge runs no link or disambiguation query.

### X-14 · minor · Sweep and rg wording
**Answer:** Fixed — the sweep matches `NodesOnBridge(`; STEP-6's rg names the removed map names; `newLargeFixture`'s file corrected.

## Verdict

Round 2: maintainability Fits with changes (M-10 fixed, M-11 fixed); executability Ready with changes (all fixed). No open blocker; every To user item answered. Two rounds done, so the M-10 reshape was not reviewed a third time; it changes vocabulary only, and the commit-side premises round 2 checked are the same.
