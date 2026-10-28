# Review — S12-35 FakeStore is a seeded fake

**Verdict:** Fits with changes — Ready → Fixed
**Rounds:** 2. **Reviewers:** maintainability (inherit), executability (inherit)

## To user

None.

## Writer's own calls

- Kept `insertSource`'s Default-graph append and `bumpActivity` Source-row Updated mutation as What stays (fixture convenience), outside Ships' four removals; STEP-6 aligns the working rule and map Check with that Ships surface instead of expanding In to every leftover synthesizer.
- Chose sticky `nextWriteEffects` (not a per-call queue) and applied it to every inventing `CatalogWriteEffects` return, not only `bumpSource` / `graphEffects`.
- Escape decoder covers the Swift string-literal set (`\\` `\"` `\'` `\n` `\r` `\t` `\0` `\u{n}`); dropped `\xHH`.

## Round 1

### M-1 · blocker · D-8 leaves `swift-tests.mdc` teaching a false FakeStore rule
**Answer:** Fixed — D-8 / STEP-6 rewrite the FakeStore paragraph to match In + What stays (scripting shape; forbid A-15 removals; do not claim every leftover is gone); `add-swift-test` gains a FakeStore bullet.

### M-2 · blocker · Out keeps membership and refusal logic while **Rules this plan bends** is empty
**Answer:** Fixed — STEP-6 clarifies the working rule's S12-35 sentence to the Ships surface; Out leftovers stay outside Ships and outside the pattern new methods copy; **Rules this plan bends** stays None with that alignment. Not To user: map Ships already limited S12-35 (K-9 option a).

### M-3 · major · Map Check claims a clean FakeStore; AC / Out do not deliver that
**Answer:** Fixed — STEP-6 rewrites the map Check to named deletions + `return CatalogWriteEffects(` gone + scripted tests + fixture pins (AC-1.5).

### M-4 · major · D-2 drops citation list ordering; STEP-2 / DoD do not
**Answer:** Fixed — D-2 and STEP-2 delete the `listCitationsByArtifact` ref sort with `citationIsInGraph`.

### M-5 · major · File comment still teaches Default-graph-as-engine
**Answer:** Fixed — STEP-2 rewords the `evidenceGraphsByProject` comment (seed-on-create; tests may seed directly).

### M-6 · major · `bumpActivity` hard-codes which writes move Updated beside scripted effects
**Answer:** Fixed in part — D-3 binds every current `bumpSource` call site → `bumpActivity(same sourceID)` (no "when you still want" judgment); What stays names Updated-cell fixture convenience as distinct from effects composition. Declined full "activity only from `nextWriteEffects`": more test churn than Ships asks; effects stay fully scripted.

### M-7 · minor · Linking test `createArtifactLinksOnlyItsProjectsGraphs` unnamed
**Answer:** Fixed — folded into T-7.

### M-8 · minor · `add-swift-test` silent on the pattern
**Answer:** Fixed — D-8 / STEP-6 add the skill bullet.

### M-9 · nit · Escape decoder lists `\xHH`
**Answer:** Fixed — D-7 drops `\xHH`; adds `\r` / `\'`.

### X-1 · blocker · `citationIsInGraph` has a second caller STEP-2 never specifies
**Answer:** Fixed — Current state, D-2, STEP-2 cover `citationCountsByEvidenceGraph` (keep `artifactIDs.contains`, drop membership).

### X-2 · blocker · D-3 vs STEP-3 cutover of inventing returns
**Answer:** Fixed — D-3 and STEP-3 require every `CatalogWriteEffects` return → `nextWriteEffects`, naming the inventing sites; Check forbids `return CatalogWriteEffects(`.

### X-3 · major · D-3 leaves per-call-site judgment for `bumpActivity`
**Answer:** Fixed — every current `bumpSource` → `bumpActivity(same sourceID)`; `graphEffects`-only paths get none.

### X-4 · major · T-6 unknown-graph row
**Answer:** Fixed — T-6 keeps `deleteEvidenceGraph(…"nope")` → `StoreBoom`.

### X-5 · major · `createArtifactLinksOnlyItsProjectsGraphs` unnamed
**Answer:** Fixed — T-7.

### X-6 · minor · T-10 call shape
**Answer:** Fixed — T-10 pins `valueType` / empty description / cardinality and seeded `propertyID` `p1`.

### X-7 · minor · `firstGraphID` multi-project walk
**Answer:** Fixed — D-1: any order across projects; seed order within each array.

### X-8 · nit · Where `decodeSwiftStringLiteral` lives
**Answer:** Fixed — same file as `readSwiftFixturePairs`.

## Round 2

### M-10 · blocker · Standing FakeStore working rule stays false
**Answer:** Fixed — D-8 / STEP-6 rewrite the standing working-rule paragraph (not only the S12-35 clause) to Ships + scripting shape + leftovers not the pattern.

### M-11 · blocker · Effects DoD Check misses inventing tuple returns
**Answer:** Fixed — AC-1.5 / STEP-3 / STEP-6 / Manual check gate on `CatalogWriteEffects\(` in `FakeStore.swift` (covers bare and tuple inventing sites).

### Executability round 2
**Answer:** No new findings. Verdict Ready.
