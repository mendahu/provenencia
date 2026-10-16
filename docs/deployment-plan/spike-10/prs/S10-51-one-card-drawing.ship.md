# Ship log — S10-51 Draw Evidence cards in one place

**Plan:** [`S10-51-one-card-drawing.md`](S10-51-one-card-drawing.md) · **Branch:** `s10-51-one-card-drawing` (stacked on `s10-44-unpromote`, #418) · **PR:** #421
**State:** ready

## Round 0 — implement
Base `1c489b80` (`s10-44-unpromote`).
Local head `d7ed5571` (unpushed). Implementer gate PASS against `s10-44-unpromote`. STEP-1…6 done; STEP-7 and the PR not started.
**Stopped on:** §11 "A `L10n.EvidenceGraph` key becomes unused after STEP-5". `promoteHelp(kind:)` (tooltip on the twin's Promote button) and `citedValueUnavailable` (twin's cited row) are now unused. The live card had no Promote tooltip at `1c489b80` either (no `help` in `EvidenceAppKitCard.swift`), so researchers have not seen it since #367. Plan §6 Strings was wrong.
Deviations: `@MainActor` added to four test suites (`EvidenceBridgeEdgeSummaryTests`, `EvidenceSubjectCardFooterTests`, `GraphCanvasEdgeGeometryTests`, `SourceGraphSnapshotTests`) after D-2; app target unchanged — accepted by the orchestrator (test callers, not app callers). Plan §4 path cell reworded after the rename. T-1/T-2 written with the code (could not compile first).
**User (2026-10-16):** "Delete promoteHelp and citedValueUnavailable and their xcstrings entries; the live card never showed them. Carry on shipping S10-51." Recorded as a deviation in the plan's Landed note.
Resumed: keys deleted, STEP-7 done. Head `d910ecdc`. Gate PASS (orchestrator run, `--base s10-44-unpromote`, against `1c489b80`). PR #421 (draft, base `s10-44-unpromote`).
Report deviations: `@MainActor` on four test suites; two keys deleted (user); §4 path cell reworded; T-1/T-2 written with the code; §10 not run.

## Round 1 — review
Lenses: all five on head `d910ecdc`. SEC clean, PRF clean, UX clean, MNT clean, CON changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | AC-2.1 / AC-2.3 have no evidence: §10 not run; no test pins the ghost's 0.62 opacity | Fix now → F-1 (opacity test); manual check → To user (the user runs §10, as on every PR) | verified: `rg "0\.62" macos/ProvenenciaTests` finds nothing; `EvidenceCardView.draw` sets 0.62 for `isGhost` |
| MNT-1 | MNT | minor | Comments still say "tagged region", "before a card has reported", "instead of an `NSHostingView`" | Fix now → F-2 | verified: `EvidenceCardLayout.swift:23,30,46,54`, `EvidenceCardView.swift:6`. Comment-only, on text this PR's purpose makes false; rides the fix round. Dedupes CON-2 |
| CON-2 | CON | nit | `EvidenceCardView` header contrasts with `NSHostingView` | → MNT-1 | same line |
| MNT-3 | MNT | nit | T-2 builds its non-ghost card with `.ghost(...)` | Fix now → F-1 (same test file, one line) | `EvidenceCardLayoutTests.swift:181` |
| MNT-2 | MNT | nit | Two `makeView`s in `GraphCanvasItemLayer.swift` | Logged | rename the private method when the file is next touched |
| UX-1 | UX | minor | A cited property with no displayable value paints a blank value line; the "—" key is now deleted | Follow-up | predates this PR (live card since #367); key deletion was the user's call; idea: render term/subject values in `ObservationValueDisplay`, or a placeholder |
| CON-OUT-1 | MNT (from CON) | minor | `armedEscHint`, `defaultLabelLocation`, `defaultLabelParticipation`, `defaultLabelRelationship` unused before this PR | Follow-up | unused at base `1c489b80` |
| SEC-OUT-1 | UX (from SEC) | minor | Ghost keeps old presentation if a type's look changes while armed | Rejected | accepted in plan D-4 |

### Fix plan 1

**F-1 — Pin the ghost's opacity, and build T-2's live card from `.resting`** (from CON-1, MNT-3)
Files: `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`.
Failing test first: `ghostPaintsAtGhostOpacity` *(new)* — render `EvidenceCardView(ghostOf: .ghost(placed(rows: 0), presentation: nil))` and `EvidenceCardView(card: EvidenceCanvasCard(placed: .subject(placed(rows: 0), nil), text: EvidenceCardText(subject: placed(rows: 0)), state: .resting), actions: [])` through the same window-and-bitmap path `measure` uses (D-6); take the highest alpha down the centre column of each; expect `abs(ghostMax / liveMax - 0.62 / 0.76) < 0.02`. Show it fails when the ghost alpha line is temporarily removed, then restore it (do not commit the removal).
Do: add the test, sharing the bitmap helper with `measure` (extract a small private helper if `measure` does not already expose the bitmap). In `ghostCardIsNotAnAccessibilityElement`, build `live` as T-1 does (`EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting)`).
Check: `EvidenceCardLayoutTests` pass; scripts/check-pr.sh PASS.

**F-2 — Comments that still describe the SwiftUI path** (from MNT-1, CON-2)
Files: `macos/App/Features/EvidenceGraph/EvidenceCardLayout.swift`, `macos/App/Features/EvidenceGraph/EvidenceCardView.swift`.
Failing test first: none (comments).
Do: `EvidenceCardLayout.swift` lines 23, 30, 46: "tagged region"/"tagged control"/"tagged regions" → "region"/"control"/"regions". Line 54: "before a card has reported" → "before its view is built". `EvidenceCardView.swift` line 6: "The canvas hosts one of these instead of an `NSHostingView`." → "The canvas hosts one of these per card." No code change.
Check: `rg -n "tagged|has reported|NSHostingView" macos/App/Features/EvidenceGraph` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5b48c30f`. Gate PASS (orchestrator run). F-1 done (`ghostPaintsAtGhostOpacity` red with the alpha line removed, green restored; `f383c455`): compares the modal alpha down the centre column with tolerance 0.06, not the max with 0.02, because fill, shadow and text composite (measured ratio 0.854 vs 0.816; a 0.76 ghost gives 1.0). Accepted. F-2 done (`5b48c30f`); three stale "has reported"/"tagged" comments remain in `EvidenceSubjectCard.swift:78`, `EvidenceBridgeCard.swift:43,148`, outside F-2's files → Logged (MNT-1b). Implementer saw one flaky `WorkspaceSessionTests.invalidateAllMatching` run (untouched code); passed on rerun → Follow-up (flaky test).

## Round 2 — review
Lenses: CON, MNT (the two with Fix-now findings). Scope `d910ecdc...5b48c30f`.
CON: CON-1 resolved (test); CON-2 resolved via F-2. MNT: MNT-1, MNT-3 resolved. No new findings. No Fix plan 2.

## Result
Gate PASS (`5b48c30f`, against `1c489b80`) · CI: none — workflows run only on PRs into `main`; runs after #418 merges and #421 is retargeted · open: none · follow-ups: UX-1, CON-OUT-1, MNT-1b, flaky `WorkspaceSessionTests.invalidateAllMatching` · to user: §10 manual check
