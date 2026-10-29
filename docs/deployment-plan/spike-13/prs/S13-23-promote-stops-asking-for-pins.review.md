# Review — S13-23 Promote stops asking for pins

**Verdict:** round 1: Fits with changes — Not ready. Every finding is answered below; round 2 runs on the refreshed plan once board S13-D11 is agreed.
**Rounds:** 1 of 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **Does the removal need a brief?** Asked 2026-10-29. Answer (user): yes. Removing chrome changes how the page looks; briefs keep the boards in sync with the app and catch errors. Done in the planning commit: brief [S13-D11](../design/S13-D11-promote-without-assessment.md) written; the working rule "No brief for the removal" replaced by "Promote's removal has a brief"; the brief gate in `plan-pr` (SKILL and template) and `plan-spike` (SKILL and map template) now names removals. The plan is `Blocked — on board S13-D11`.
- **What the Updated caption says** without the row's reason. Answer (user, 2026-10-29): the move, else the via neighbor, else nothing. Now D-3.
- **What the freed column becomes** (S13-D11 PA-2). Open until the board is agreed; the plan's D-8 and §12 take the board's choice at the refresh.

## Writer's own calls

- D-4: the Swift row keeps `assessment` and `reason` for `draftTarget` until S13-24 (the map gives the draft rule to S13-24).
- D-6: Done sends `CatalogPromoteGraphAlignmentFixed`, the type proposals already send, and `CatalogPromoteBatchRow` goes (answer to M-1).
- D-9: `PromoteUpdatedCaption` moves into `PromoteView.swift` and `PromoteReadout.swift` is deleted.
- D-11: one stale key, `promote.toast.filed`, goes with the sweep.
- D-12: two `add-match-rule` sentences move from S13-24 to S13-23, recorded in the map's working rule.
- D-14: the fake keeps its stale-revision check (see M-2).
- D-15: `Wording` goes and term labels become plain constants (answer to M-3).

## Round 1

### M-1 · blocker · Done's batch row becomes a second Target → (code, id) mapping beside `fixed(for:)`
`PromoteFlow.fixed(for:)` already maps `Target` to `CatalogPromoteGraphAlignmentFixed` (`subjectID`, `handleID`, `target`); the plan's typed `CatalogPromoteBatchRow` had the same fields under other names.
**Answer:** Fixed — D-6: `batchRows()` returns `fixed(for: filedRows)`; `CatalogPromoteBatchRow` and `batchRow(_:)` go; Apply takes `[CatalogPromoteGraphAlignmentFixed]`; the two false comments are fixed (STEP-4).

### M-2 · blocker · The rewritten FakeStore Apply keeps copying Go's refusals
`FakeStore.swift` lines 1860–1887 copy Go's `applyRow` refusals, against `.cursor/rules/swift-tests.mdc`.
**Answer:** Fixed — D-14: the fake files what it is sent and refuses nothing; no test relies on those refusals (`rg 'already_member\|promote.invalid' macos/ProvenenciaTests` is empty), and a stop condition covers one that does. The stale-revision check stays: it is the fake's audit clock, the same one `proposePromoteGraphAlignment` returns (`nextAuditRevision - 1`), and `aStaleDoneReproposes` drives it with a real write rather than a copied rule.

### M-3 · major · `L10n.PropertyTerm.Wording` becomes a one-field wrapper, with two lookups for one value
**Answer:** Fixed — D-15, STEP-5: the 39 entries become `LocalizedStringResource` constants, `resource(propertyKey:termKey:)` holds the switch, `Wording` and `wording` go, the Go subtest reads `resource`, and `add-seeded-vocabulary` and `ideas/international-names.md` name `resource` (STEP-6).

### M-4 · major · Dead fields kept in the structs and mapping this PR rewrites
**Answer:** Fixed — D-2: `CatalogPromoteBatchResult` and `CatalogPromoteBatchWritten` go and Apply returns nothing (`rg '\.written\b' macos` shows only the line that builds it). D-5: `viaBridgeType` and `viaRole` go; the Out row is removed. AC-3.1 and AC-3.2 search for them.

### M-5 · minor · The two new header mappers make four copies of the same oneof switch
**Answer:** Fixed in part — D-7 names `mapSearchHeader` and places the four mappers together. One shared proto header message is logged as an Out line for S13-24's planner, which regenerates the proto.

### M-6 · minor · The empty assessment slot is a fourth spacer spelling in three hand-kept column grids
**Answer:** Fixed — D-8: `PromoteLayout.arrowColumn` replaces the literal `14`s and the grid is expressed in `PromoteLayout`; under the board's option (a) the slot is one named, commented `PromoteFreeColumn`. One grid view is an Out line (not planned). Next adds no longer claims S14-22 meets no column.

### M-7 · nit · "Reason" ends up naming two things
**Answer:** Logged. S13-24 deletes the wire `reason` from Swift; the next plan that touches `PromoteFlow.Reason` after that can rename it.

### X-1 · blocker · No Swift or Xcode check can run in this environment
**Answer:** Fixed — D-16 and a stop condition: the PR is implemented and checked on macOS, where `scripts/check-pr.sh` runs `xcodebuild`. The DoD's check-pr line says "on macOS".

### X-2 · blocker · STEP-2's check cannot pass when STEP-2 ends
**Answer:** Fixed — STEP-2's `rg` searches `row\.contested` in the Promote feature and tests, not `GoStore`'s `r.contested`, which STEP-4 removes and AC-3.2 checks.

### X-3 · major · STEP-1 and STEP-2 share one list of tests to delete
**Answer:** Fixed — §8 lists deleted and narrowed tests by step, and D-13 states the rule (delete a test in the step that removes the last member it names).

### X-4 · major · Fixture rewrites and `Reason(code:…)` are not decided
**Answer:** Fixed — D-4 gives `Reason`'s four cases and two initializers, how `init?(move:)` treats `.via`, `.taken`, unknown codes and an empty neighbor, and which initializers go. D-13 says how both fixtures change. `reasonCodesParseThroughTheWireVocabulary` is narrowed (STEP-2); T-3 covers the initializers.

### X-5 · major · The docs edits for matching-model and promote-graph-alignment point at text that is not there
**Answer:** Fixed — §4's Docs row quotes the lines, and STEP-6 gives each replacement sentence (`matching-model.md` lines 5 and 122, `promote-graph-alignment.md` lines 7, 35 and 62, and the skill and idea lines).

### X-6 · minor · The plan is not yet in the map, and its own plan-docs check warns
**Answer:** Fixed — the planning commit links the plan in the map row (Status `Blocked — on board S13-D11`); STEP-7 sets `Landed in #N`. The header's Status and the Implementer block's "Planned" were never in conflict: the block is the template's verbatim text; a new stop condition refuses to start unless Status is `Planned`.

### X-7 · minor · Spots where I would have to pick
**Answer:** Fixed — D-3 computes `caption` on every non-anchor row, a kept decided row included; T-3 names `moveNeighborSubjectID` and adds the empty-neighbor case; T-1 and AC-1.3 state the expected `CatalogPromoteGraphAlignmentFixed` values; D-10 and STEP-1/STEP-2 say which lines of `graphSnapshot()`'s label block each step removes; §4 notes that `board(of:)`'s `labels:` closure is `propertyLabel(key:)`'s only caller, so STEP-1 can remove both.

## Brief S13-D11 review (maintainability, Opus)

Verdict: Fits with changes. Paste block verbatim, no new kit piece, no safety rule waived. Findings and answers:

- **M-1 · major · PA-3 renamed the headings.** The copy is "On this Source" and "Files on" (`L10n.Promote.columnSource`, `columnTarget`). **Fixed** — PA-3, the inventory row, AC-1.2, §6 Strings and the manual check keep today's copy; only Assessment goes.
- **M-2 · major · Three hand-kept copies of the grid.** **Fixed** — the brief's Column grid row and the plan's D-8 make one private `PromoteGridRow` the headings, working row and Already filed row fill; `PromoteFreeColumn` and the "one grid view is a later tidy" Out line are gone.
- **M-3 · minor · PA-2 (b) contradicted itself.** **Fixed** — (b) gives the freed width to the Source name; target (330 pt) and badges move right.
- **M-4 · minor · "Filed earlier" copy and its catalog comment.** **Fixed** — the brief quotes "Filed from this Source earlier"; STEP-2 rewrites `promote.row.filedEarlier`'s comment.
- **M-5 · minor · PA-8's spoken sentence.** **Fixed** — PA-8 says the menu reads as today ("Files Mary Smith on" and the chosen option); a Domain fact quotes today's reading.
- **M-6 · minor · PA-5 left out the off-page fallbacks.** **Fixed** — PA-5 and frame 2 include them; Objective item 4 says "when the proposal names a settling move or a via neighbor".
- **M-7 · nit · Name `PromoteReadout.swift`.** **Fixed** — the inventory says the file is deleted and the caption becomes private.
