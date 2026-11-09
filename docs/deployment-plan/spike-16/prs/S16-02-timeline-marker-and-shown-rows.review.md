# Review — S16-02 The Timeline places a missing death and builds what it shows

**Verdict:** Fits with changes — Ready with changes (round 2 findings fixed)
**Rounds:** 2. **Reviewers:** maintainability (Opus, fresh context), executability (Sonnet, fresh context)

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

## Round 2

### M-1 · blocker · `Slot.id` is a second copy of `Item.id`
**Answer:** Fixed — D-10: `Slot` owns the id rule; `Item.id` reads its private `slot`; Codebase fit lists `Slot`.

### M-2 · blocker · `foldedHit` copies the buttons' hit geometry
**Answer:** Fixed — D-12: `Mark.hit` beside `pointY`; buttons and `foldedHit` read it; `PersonTimelineLine.top(_:)` and `hitHeight(_:)` go (Deletes); T-7 asserts it.

### M-3 · blocker · STEP-8 leaves the map's S16-02 scope text stale
**Answer:** Fixed — STEP-7 (was STEP-8) rewrites the map row's Scope, the scope entry's Ships and Check, and links the plan; the DoD sweep covers the spike folder.

### X-8 · blocker · STEP-3 leaves the content reading `besideEventId`
**Answer:** Fixed — STEP-3 takes D-6 and `PersonTimelineContent.swift`; STEP-4 keeps the lifeline (D-7 – D-9).

### X-9 · blocker · STEP-5 breaks the build; T-5 needs STEP-6's targets
**Answer:** Fixed — the two steps merged into STEP-5 (D-10, D-11, D-12); steps renumbered (7 in all).

### X-10 · major · §6 sketch says `[Mark]`
**Answer:** Fixed — the sketch declares `Target` and `[Target]`.

### X-11 · major · T-1's mirror literal does not compile
**Answer:** Fixed — T-1 gives the `EventRole` literal and the `row(n, dated, roles...)` helper.

### X-12 · major · Two `rg` checks cannot pass
**Answer:** Fixed — STEP-2 excludes `*.pb.go` and names the one expected hit; STEP-7 searches for the stale phrases only.

### X-13 · major · Stale map wording and no Plan link
**Answer:** Fixed — as M-3.

### X-14 · minor · STEP-4 contradicts itself
**Answer:** Fixed — STEP-4 names the `unknownEnds` lines that change.

### X-15 · minor · Smaller choices
**Answer:** Fixed — T-6 names `d60`; AC-3.1 says the range changes; the Out row and the idea paragraph say every dated row after the first own burial or funeral; STEP-2 names the generated Swift case names. How the lifeline stores folded marks is left to the implementer (same result either way, as the reviewer says).

### X-6 · minor (repeated) · No `protoc` or `xcodebuild` in the review sandbox
**Answer:** Declined — the implementer works on the Mac.
