# Review — S11-14 Segment compare

**Verdict:** pending round 2 (round 1: Fits with changes — Not ready)
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- **`segment.OneConflict` = 3.** It's a new knob, the points a differing birth or death subtracts. No source dictated it. The estimate equals the name and date conflict penalties.
- **Placement comes from the walk.** `Compare` takes `Together` or `Apart` rather than judging an unplaced neighbor by entity compare. The walk compares placed neighbors only, as today; S11-16 adds the third placement.
- **`part_of` gates are generic `Gates` entries keyed by `BridgeKey`.** They replace the inventory's named `PartOf` field.
- **`graphcache.Link.Directed` and `InverseKey` are deleted** (M-8), not only logged. Nothing outside `graphcache` reads them, and `loadTerms` is the one source.
- **`connectrules.productBridges` reads the new constants, and `propertyterms.KeyPartOf` points at `vocabkey.TermPartOf`** (M-4). `conclusionheaders.RelPartOf` is left to "moves when next edited".

User calls of 2026-10-19 (asked during planning): one-cardinality holds from both ends (D-5); bridge properties are deferred to S11-16 (D-10).

## Round 1

### X-1 · blocker · AC-1.2 contradicts D-3's canonical-term rule
`child` sorts before `parent`, so the canonical term is `child`, and AC-1.2's `parent` keys could never pass.
**Answer:** Fixed. AC-1.2 now expects `relationship·child·related_to` from Mary's end and `relationship·child·person` from John's end, and says why. AC-4.2 asserts `Via.Key.Term` is `child`; the stop condition reads the same.

### X-2 · blocker · D-7 gives location links an event type, the canon side does not
**Answer:** Fixed, together with M-1. `linkOf` is the one builder, and it keeps the event type only for a participation (today's rule on both sides). `canonKeys` calls `connectrules.LookupBridge` then `linkOf`. AC-4.5 / T-10 (`TestProposePlaceThroughALocation`) proves a layer location corresponds to a canon one.

### M-1 · blocker · The event-type rule is written twice and the copies disagree
**Answer:** Fixed — D-7: one builder, `linkOf`, used by `loadLayerBridges` and `canonKeys`; a Risks row names the drift.

### X-3 · major · T-14 / the stop condition check `via_role`, but no test asserts it
**Answer:** Fixed. AC-4.2 adds a `Via.Key` assertion (Type `relationship`, Term `child`) to `TestProposeRelationshipFromEitherEnd`. The handler maps the key's fields one-to-one (D-8), so that test now pins the value the wire sends.

### X-4 · major · Steps hide choices
**Answer:** Fixed, each in D-6 / D-7:
- `segments` works per layer link. It takes the first canon link in `canonAdj` order that lands on the placed handle, else the first that corresponds (apart). Neighbor kinds come from `st.subjects` / `st.handles`.
- `edgeSeenKey` swaps ids and keys together.
- An unobserved term gives an empty `Term`.
- `canonKeys` goes through `LookupBridge` and `linkOf`.
- `Terms` entries are missing for bridges with no terms (the zero `Term`).
- The place-relationship walk change is in §1: the vocabulary is product-locked, so no link changes today.
- US-2 now gives D-4's 8-argument call.

### X-5 · major · No rule for a flipped Target in an unlisted ported test
**Answer:** Fixed — D-11 allows a fixture reshape that keeps the test's intent (the S11-13 precedent), each listed in the Landed note; a (c) test flipping stays a stop.

### X-6 · minor · Fixture work undefined
**Answer:** Fixed. US-4 names the fixture helpers and the dated-birth pattern (`TestProposeOneHopDateExhibit`). AC-4.3 names Source C. STEP-6 spells out `LayerBridges`' calls. AC-4.3 now asserts the Target, not a pick score.

### X-7 · minor · Docs and map steps point at changed text
**Answer:** Fixed.
- D-10 says the map scope entries were edited at planning.
- STEP-7 makes the map row's status its only map edit.
- D-12 rewrites §Segments' bridge bullet and §Match registry's sketch.
- The header's Context says D-2 replaces that sketch.
- The `.To.ConflictFactor` row's source is "decided: the whole's end had no conflict factor".

### X-8 · minor · Checks that cannot run on Linux
**Answer:** Declined. `scripts/check-pr.sh` runs `xcodebuild test` by design, and the implementer ships from macOS (`ship-pr` runs it there). STEP-7 and §10 now say "on macOS". The Go checks in each step run on any platform.

### X-9 · nit · Notes
Logged. `newPairSegments` now calls subject compare once per pair (M-9).

### M-2 · major · Correspondence inlined at four or five sites, and `Result.Corresponds` duplicates the cause
**Answer:** Fixed. `segmentcompare.Corresponds` (D-3) is the one rule; every walk site and `Compare` call it (D-6). `Result.Corresponds` is dropped; `Cause == not_corresponding` says it.

### M-3 · major · `One` needs four entries for two facts
**Answer:** Fixed. `One` keys have an empty `End` and hold for both ends (D-4, D-5): `participation·subject·birth` and `·death`. T-4 checks either end gives the same apart evidence. T-11 and AC-5.2 check `End` is empty.

### M-4 · major · New constants beside the literals they name; a third `part_of`
**Answer:** Fixed. `connectrules.productBridges` uses the constants, and `propertyterms.KeyPartOf = string(vocabkey.TermPartOf)` (D-2, STEP-1). `RelPartOf` and the `hops.go` / `layer.go` literals join the "moves when next edited" Out line.

### M-5 · minor · `Kind` names a primary Subject type in vocabkey
**Answer:** Fixed. The renames are `BridgeKey.Type`, `Link.Type`, `Link.SameType`, and the constants `BridgeParticipation`, `BridgeRelationship`, `BridgeLocation` and `BridgePlaceRelationship`.

### M-6 · minor · Cause codes have no home in the skill
**Answer:** Fixed — D-12 extends "Add a relation, cause or unit code" for layer 5 and lists the causes in §Segments' "Built in S11-14" paragraph.

### M-7 · minor · Test helpers re-encode `Key`'s rule
**Answer:** Fixed.
- New package `segmentcomparetest.Terms()` (D-1), with a `layerPackages` row.
- AC-5.2 checks it against the seed.
- The `align_test.go` helpers build keys through `Key` (D-11).
- The test-only terms move to seeded ones.

### M-8 · minor · graphcache leftovers
**Answer:** Fixed. D-7 deletes `Link.Directed` and `Link.InverseKey` with their term-mark columns. It also corrects the `Link` doc, and the stats slot doc says nothing fills it until S11-15.

### M-9 · minor · `newPairSegments` scores each neighbor pair twice
**Answer:** Fixed. `neighborsMatch` and `pairNode` are deleted. One `Compare` per pair supplies both the bar (`Neighbor.Evidence ≥ AcceptScore`) and the evidence (D-6). `Compare` keeps `Neighbor` when gated, so the bar still reads.

### M-10 · nit · `bridgeKey` beside `BridgeKey`; a `Role` alias
**Answer:** The rename is Fixed: `edgeSeenKey`. The alias is Declined (Out): `Role` types participation roles for lifespans, and an alias would let a relationship term pass as a role.
