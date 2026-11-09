# Review — S16-02 The Timeline places a missing death and builds what it shows

**Verdict:** Fits with changes — Not ready (round 1), fixed; round 2 pending
**Rounds:** 1. **Reviewers:** maintainability (Opus, fresh context), executability (Sonnet, fresh context)

## To user
- Before the plan: the map's "click targets only for shown rows" contradicts S16-D2 FH-5 (a folded point opens its page). **Answered (user, 2026-11-09):** one shared click target for the folded points; folded points lose their hover name (D-11, D-12; Rules this plan bends).
- Before the plan: no brief for moving the death placeholder. **Answered (user, 2026-11-09):** no brief, a bugfix to R-24a.
- M-1: count only the person's own burial or funeral (role subject)? **Answered (user, 2026-11-09):** subject only.
- X-5: the dated rows after the death marker never fold. **Answered (user, 2026-11-09):** log it in the person page performance idea (Out row, STEP-8).

## Writer's own calls
- A "before" stub caps at its top, an "after" stub at its bottom, whatever the kind (D-8).
- Points shift past stubs in item order, ticks by year (D-9), so a same-year baptism and burial keep the list's order on the line.
- Field 2 of `LifeMarker` is reserved, not reused (D-4).
- The "builds only what it shows" test is structural, not a wall-clock budget (T-3).
- `markerInsideHiddenRunStaysShown` and `foldedMarkersStayAfterPagedRun` keep Swift placements Go does not produce (STEP-3 table).

## Round 1

### X-1 · blocker · `timeline_test.go` is missing and breaks on `Beside`
**Answer:** Fixed — §4 Read tests row; D-15 gives every subtest's expectation; STEP-1 lists the file; T-9.

### X-2 · major · The `lee` handler subtest changes too
**Answer:** Fixed — §4 Handler test row; T-2 names its expectation.

### X-3 · blocker · D-10 leaves the id-level type open
**Answer:** Fixed — D-10 names `Slot`, how each fold variable uses it, and where `Row` is built.

### X-4 · major · Several Swift choices open
**Answer:** Fixed — STEP-3 table for every fixture; §4 Fixtures row corrected; D-13 names `keptScale` and the `rebuild()` wiring; D-12 gives the region's frame, offset and coordinate conversion; D-7 gives `place` a default; T-1 names ids 6 and 7; T-8 moves event "11".

### X-5 · major · Rows after the death marker never fold
**Answer:** To user → log it (Out row, STEP-8). With M-1's subject-only rule, only a person's own dated burials and funerals land there.

### X-6 · minor · `protoc` and `xcodebuild` are not in the review sandbox
**Answer:** Declined — the implementer runs on the Mac; STEP-2 names the tools.

### X-7 · minor · Small inconsistencies
**Answer:** Fixed — AC-2.2 counts targets per shown item with a point; T-6 states the stacked-stub case; T-3 gives 52.

### M-1 · blocker · After-death ignores the person's role
**Answer:** To user → subject only. Fixed — D-2, D-3 use `subjectcompare.EventRole`; D-14 now expects the death after the witnessed burial; T-1 gains the witness cases; Rules this plan bends lists the rule change; STEP-8 rewrites it.

### M-2 · major · Content and lifeline each resolve Go's place
**Answer:** Fixed — D-6 and D-7: the content resolves once, `Marker.place` holds the resolved place, the lifeline reads it; T-6 adds the unknown-id case.

### M-3 · major · The working rule still says "beside"
**Answer:** Fixed — Rules this plan bends; STEP-8 rewrites the rule and checks for "beside".

### M-4 · minor · `pointLabels` repeats `items`
**Answer:** Fixed — D-11: `Item.spokenLabel`; D-12: targets carry their label; `pointLabels` and the `labels` parameter go.

### M-5 · nit · `MarkerSide`'s zero value
**Answer:** Fixed — D-1 starts at `iota + 1`.
