# Review — S10-51 Draw Evidence cards in one place

**Verdict:** Fits — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- The ghost stays in the SwiftUI base layer as an `NSViewRepresentable` around the live card (D-4), not a canvas item.
- `EvidenceSubjectCard` / `EvidenceBridgeCard` keep their names as `@MainActor enum`s (D-2); only the live card is renamed, to `EvidenceCardView` (D-9).
- `GraphCanvasItem` keeps one required `makeView`; the SwiftUI item path is removed rather than kept as generic tooling (D-3).
- A presentation change while one ghost kind stays armed keeps the old colours until the kind changes (D-4).
- The plan is written against the `s10-44-unpromote` head and stays Blocked until #418 merges.

## Round 1

### M-1 · major · The ghost card is assembled inline, and the tests build their own copy
Ghost recipe and the resting `State` would be hand-written at about five sites once `EvidenceSubjectCard.ghost` is deleted.
**Answer:** Fixed — D-5 adds `EvidenceCanvasCard.State.resting` and `EvidenceCanvasCard.ghost(_:presentation:)`; the base layer, T-1/T-2 and STEP-3's tests use them. Replacing the literal in `EvidenceGraphModelTests` is an Out line (file not otherwise touched).

### M-2 · major · Leftover dead code and stale doc comments
`promoteInset`, `shellPaddingBottom`, the `regions` DocC link, type headers, `GraphCanvasItemLayer` / `GraphCanvasDocumentView` / `EvidenceCanvasCard` comments, and `GraphCanvasItemLayer`'s `import SwiftUI`.
**Answer:** Fixed — both statics are in Deletes and D-2; §6 lists the text for each comment and README paragraph; STEP-5 writes them; T-8 searches for the deleted names and for `NSHostingView` / `import SwiftUI` in the two canvas files.

### M-3 · minor · `isGhost` on the shared `State`
**Answer:** Fixed — D-1 makes the ghost a construction mode, `EvidenceCardView(ghostOf:)`, with a private flag; `State` gains no ghost field.

### M-4 · minor · D-4's rejection reason did not match the code
**Answer:** Fixed — D-4 now rejects the item path for z-order (items sit above the base layer) and hover churn in `cards`/`pushItems`.

### M-5 · minor · "AppKit" in the live card's name
**Answer:** Fixed — D-9 and STEP-1 rename `EvidenceAppKitCard` → `EvidenceCardView` and `appKitView(for:)` → `cardView(for:)` before any new call site.

### X-1 · blocker · Test `makeView` item has no size
**Answer:** Fixed — D-3 says the view sizes itself through `intrinsicContentSize` and the layer keeps `frame.size = fittingSize`; STEP-4 names a `SizedView` test class with that intrinsic size.

### X-2 · blocker · AC-1.1's search fails on matches the plan did not delete
**Answer:** Fixed — the README paragraph and the `regions` comment move into STEP-5 with exact text in §6, before T-8 runs.

### X-3 · major · T-5's bridge had no membership
**Answer:** Fixed — STEP-6 gives the full filed-bridge fixture, membership included.

### X-4 · major · Ghost lifecycle undecided
**Answer:** Fixed — D-4: `.id(ghost.kind.rawValue)`, `updateNSView` does nothing, a stale presentation within one kind is accepted. D-1: the ghost sets no label, identifier or actions.

### X-5 · major · "Remove `import SwiftUI` where unneeded" was a judgment
**Answer:** Fixed — only `GraphCanvasItemLayer.swift` drops it (D-3); other files are an Out line.

### X-6 · minor · `EvidenceCardHeaderActionHits` neither kept nor deleted
**Answer:** Fixed — §1 What stays and §4 name it as kept.

### X-7 · minor · AC-2.3 has no automated check
**Answer:** Declined — `regions` do not carry the fill or glyph size; a bitmap probe of a 22pt square is a new test technique for a token swap. T-4 guards the regions and manual check 3 the paint.

### X-8 · minor · T-1/T-2 fixture
**Answer:** Fixed — T-1/T-2 use `placed(rows: 0)`.

### X-9 · minor · Stop condition true at the base
**Answer:** Fixed — the header says the branch is cut from `main` after #418 merges, and the status stays Blocked until then.

## Round 2

Maintainability: all five answers resolved; no new blockers; verdict **Fits**. Note: README line 66 ("cannot reach into separate hosting views") sits in the paragraph §6 rewrites. **Answer:** Fixed — §6 README row now covers lines 63–66.

### X-1 · major · D-6 left the test window size undecided
**Answer:** Fixed — D-6 sizes the window from `view.intrinsicContentSize`, bitmaps `view.bounds`, and scales by `view.bounds.height`.

### X-2 · major · D-1 contradicted the designated init
**Answer:** Fixed — D-1 makes `init(card:actions:isGhost:)` the private designated init, which skips the accessibility setup for a ghost; two convenience inits call it. T-2 also checks label and identifier.

### X-3 · minor · Statics only the chrome read
**Answer:** Fixed — STEP-5 says every other static stays; the round-2 maintainability reviewer confirmed each has a live reader.

### X-4 · nit · Ghost frame includes the outset
Logged; manual check 2 covers it.

## After review
- 2026-10-16: the user chose to ship S10-51 stacked on #418. Status → Planned; the branch is cut from `s10-44-unpromote` and the PR is based on it. No Decision, Scope or AC changed.
