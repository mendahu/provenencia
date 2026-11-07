# Review — S14-38 Unpromote rewrites the ends of the links it takes back

**Verdict:** pending round 2
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **Bridge delete changes with this PR (M-1, D-6).** Sharing one read of a bridge's ends means `subjects.Delete` stops recomputing and walking from a location's place end and a place relationship's `to` end. That is the structural-debt row S14-33 left as FU-4 ("not measured"), now fixed instead of deferred. Nothing stored reads a link through those ends (S14-33 F-2; this plan's succession and part_of probe). Recommendation: take it here, because the alternative keeps two copies of one read.

## Writer's own calls

- D-3: the new registry arm runs only on a claim **delete**, so filing keeps S14-33's path and cost.
- D-4: an end's "last updated" time does not move when another Subject's Unpromote takes a link back.
- D-7: one shared `promotetest.Unpromote`; the location and place-relationship builders are copied into the autoreconciler test fixture rather than shared.
- D-8: `budgetUnpromoteAtCity` is 500 ms, against 45 ms measured.
- Out row 1: telling the app about an unread place end without a walk goes to `docs/ideas/structural-debt.md`, not this PR. It needs the effects Set to carry ends apart from `Handles`.

## Round 1

### M-1 · blocker · `bridgeWalkEnds` is a second copy of `subjects.linkedHandles`
The new path ran the same join as `subjects.linkedHandles` (`subjects.go:353`), plus a filter, and kept both copies. The Out row's reason concerned moving `Delete` onto `writes.Run`, not sharing the read. Better shape: one filtered reader in `effects`, used by both; delete `linkedHandles`. This also fixes FU-4.
**Answer:** Fixed — D-1 and D-6: `effects.LinkEndHandles` and one SQL constant, used by the registry path and by `subjects.Delete`. `linkedHandles` is deleted (STEP-4, §6 Deletes). The FU-4 row is deleted in STEP-7. The behavior change goes To user above.

### M-2 · major · The unread-end fact belongs on the bridge registry, not in `hops`
`connectrules.Endpoint` is where bridges are declared. A list built from the part_of hop says something narrower than the rule, which covers the whole bridge.
**Answer:** Fixed — D-2: `connectrules.Endpoint.Unread`. `linkEnds` and the SQL exclusion are both built from `connectrules.Bridges()`. No `hops` type is added.

### M-3 · major · Succession links when the `to` end is dropped
Unpromoting the `from` end of a succeeded_by link leaves `to` out of the recompute set, which no S14-33 path does.
**Answer:** Fixed with evidence. Planning probe, D-1 prototype, York succeeded_by Toronto and Toronto part_of Ontario, each end unpromoted in turn: the documents equal a rebuild and `Graph().Verify()` returns nil in all four cases. `Predecessors`, `Successors`, `PartOf` and `Contains` are filled only on the detail read (`places.go:33`), never stored. Added AC-2.3, the T-4 table rows, a T-5 succession row, and an Edges Neighbors line. The app's place page is Out row 1.

### M-4 · major · No effects-level test of the new arm
**Answer:** Fixed — T-7 is a `TestConclusionHandleSets` subtest with negative cases (a claim create, a provisional delete).

### M-5 · minor · String-built exclusion vs bound arguments
**Answer:** Declined. `sqlPath` binds only the input id. The keys are product constants from `connectrules`, not catalog data, and a variant of `sqlPath` for one caller adds a mechanism. D-5 states the rule and panics at init on a quote.

### X-1 · blocker · T-5 cannot pass before STEP-3
**Answer:** Fixed — STEP-2's Check lists T-5's `Handles` assertion and T-7's positive case as failing before STEP-3, and T-8 as failing until STEP-4.

### X-2 · major · "FU-3 row" and "existing row" do not match the file's labels
**Answer:** Fixed — Out rows and STEP-7 quote each row's opening text.

### X-3 · major · Test helpers named or relied on but missing
**Answer:** Fixed — D-7 names `promotetest.Unpromote` and the autoreconciler `locate` / `placeRel` helpers with their Property keys. STEP-2 names the value helpers and literals. T-6 says how Leeds's handle is obtained. Fresh worlds per row are stated.

### X-4 · major · Position and shape of the new arm
**Answer:** Fixed — D-3 gives the full `union(…)` expression and why `from("subject_id", …)` and not `fromChange`.

### X-5 · major · How the SQL exclusion is generated
**Answer:** Fixed — D-5: built by `linkEndHandlesSQL()` at package var init, single-quoted literals, no `path.go` change; §6 shows the result.

### X-6 · minor · Map row text and link
**Answer:** Fixed — wiring updates the map row (Plan link, Status) and the scope entry's text: bridge delete already rewrote its ends; display rows were fine in the probe.

### X-7 · minor · Small choices left open
**Answer:** Fixed — `linkEnds` init shape (STEP-1), version value 0.4.1 and `project.pbxproj` (STEP-7), T-9 at `lf.citySubject` (STEP-5), fresh world per row (STEP-2).

### X-8 · minor · Non-bridge claim delete and self-links
**Answer:** Fixed — Edges Assumptions (a self-link is recomputed once through `UniqueBlobIDs`) and Reach row 3. T-7's negative cases cover a claim create and a provisional delete.
