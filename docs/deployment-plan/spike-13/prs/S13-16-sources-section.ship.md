# Ship log — S13-16 Sources section

**Plan:** [`S13-16-sources-section.md`](S13-16-sources-section.md) · **Branch:** `s13-16-sources-section` · **PR:** #722
**State:** ready

## Round 0 — implement
Stopped before STEP-7. No commit, no push, no PR, no gate. STEP-1 to STEP-6 and the STEP-8 strings sit uncommitted and unbuilt in the main checkout on `s13-16-sources-section` (cut from `4c642497`).

**Stopped on:** "`applyOpeningCamera` runs before the canvas's content size is known, so a focus cannot be centered (D-11)". The plan's Current state row "Opening a graph" is untrue in the same way: it says the canvas mounts once the rows are loaded.

Verified by the orchestrator: `EvidenceGraphView.body` shows `ProgressView` only when a query handle is missing, and otherwise mounts `EvidenceGraphContent` at once. `GraphCanvasScrollView.makeNSView` applies the opening camera on the next runloop turn, and `updateNSView` applies it again only on a `contentID` change. `model.prepare()` runs in the view's `.task`, so D-11's `openingFocus`, set in `prepare()`, is nil at a cold open. AC-2.2 fails as specified.

Options: (a) hold the canvas mount on the `ProgressView` branch until the model resolves `openingFocus`, only when `openingSubjectID` is non-nil; (b) `updateNSView` applies the focus once when `openingFocus` goes from nil to non-nil under the same `contentID`. Both go beyond D-11, so the plan changes through `plan-pr` first.

Also from the report:
- `PVTitleText(.ref)` sets its own ink, so an untitled Source's title would not turn link ink on hover; `.name` is fine. D-7 names a `PVTitleText` that fixes its own ink as a Stop condition; the implementer read it as applying only in part and carried on.
- Timeline's dated fold threads markers between event rows, so it is not a plain prefix fold. Moved onto `ConclusionFold.shown`, counting event rows only. Untested.

**Resolved 2026-11-05.** User chose option (b), apply after mount, and the plan was refreshed through `plan-pr`; review log "Refresh after the implementer's stop", verdict Ready. The refresh moved the mechanism into `GraphCanvasViewportController.center(onGridPoint:)`, widened D-11 to bridges and delete-impact rows, added `subjectId` to the place's `viewIdentity`, replaced the `PVTitleText` ink option with `Style.rowLink`, and added `ConclusionFold.firstRevealed`. Implementer resumes on the uncommitted work.

Resumed. Head `9149e924`. Gate PASS (orchestrator re-ran). Report: deviations: T-11 changes the Timeline: a marker inside the hidden run now folds away with its neighbors, where before it stayed after "Show N more" (the refreshed plan's D-10 says no behavior change and T-11 specifies the change, so the plan contradicts itself); STEP-7's `rg` phrase does not match the comment, which says what D-11 asks; `PVInkLink`'s preview uses `surfacePage`. Manual check not run (no GUI).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX routed by `macos/App/**` and strings. SEC+PRF routed by `macos/App/Features/**` models and list code built from wire records.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | `PVInkLink` owns a private `@FocusState` and a hand-written ring | Fix now → F-1 | Verified: S13-17's review log (R-4) passes a `FocusState` binding to the list, so the caller must own `.focused`. `structural-debt.md` A-40 logs this ring |
| MNT-2 | MNT | blocker | Row spoken label adds two catalog joins beside `spokenJoin` | Fix now → F-2 | Verified: `L10n.DesignSystem.spokenJoin(_:)`. Same-pass cleanup; supersedes D-12's two keys, spoken output unchanged |
| MNT-3 | MNT | blocker | Private `nonEmpty` copies `trimmedNonEmpty` | Fix now → F-3 | Verified, by body. Same-pass cleanup |
| UX-1 | UX, MNT outside lens | major | Graph link shows no focus and ignores Return | Fix now → F-4 | Verified: `PersonSourcesLineView` adds neither; brief SS-6 and frame 14d ask for focus |
| CON-3 | Orchestrator | major | T-11 changes the Timeline; D-10 and §1 say it stays | Fix now → F-5 | Verified from the diff: a marker in the hidden run used to stay after "Show N more". The plan refresh wrote T-11 wrong. §1 and D-10 win; T-11 is corrected and the Landed note says why. Neither reviewer rated it |
| SEC-1 | SEC, PRF | major | Source rows have no outer bound | Measure → F-6, then decide | Verified: plain `VStack`, and `make` runs in the host's body. A fold needs a board; lazy rows put Tab order at risk with no app to check. S13-36 SEC-1 is the precedent: measure, then To user |
| MNT-4 | MNT | major | `filterLabel` keeps a third lookup copy | Fix now → F-7 | Verified: `SourcesModel.swift:343`. Same-pass cleanup |
| MNT-5 | MNT | major | Focus after unfold is a second mechanism | Fix now → F-8 | Verified against `PersonTimelineSection.unfold` |
| MNT-6 | MNT | major | `PVInkLink` takes `identifier:` and `accessibilityLabel:` | Fix now → F-1 | Verified: `add-ui-component` step 6; supersedes STEP-5's signature |
| UX-2 | UX | minor | Back to a Subject-naming location re-centers over the researcher's pan | To user (at finish) | D-11 and Reach chose this; forward-only is a product call |
| UX-3 | UX | minor | Failed type read looks like "no type" | To user (at finish) | AC-1.3 draws an empty cell; a Callout is chrome the board does not have |
| UX-4 | UX | minor | Graph line parts can wrap | Fix now → F-4 | Verified: only `subjectName` has a line limit |
| UX-5 | UX | minor | An unfolded Source folds again after Back | To user (at finish) | Family and Timeline do the same; a fix is session state across three sections |
| PRF-1, MNT-14 | PRF, MNT | minor | The name rule sorts every Observation | Fix now → F-9 | Verified: `EvidenceGraphSnapshot.swift:69` |
| PRF-2 | PRF | minor | A Person past the 64-graph cache reloads every read | Follow-up | S13-01's code; the map's 50-graph budget is under the cap. Ask the user at finish |
| MNT-7 | MNT | minor | `sourcesTypeCell()` overloads the list's modifier | Fix now → F-10 | Verified |
| MNT-8 | MNT | minor | `pageLocation(sourceId:ref:title:)` only forwards | Rejected | D-5, from plan review M-27 |
| MNT-9 | MNT | minor | The `expanded` reset on `entityId` is dead | Fix now → F-10 | Verified: `.personDetail` `viewIdentity` is `entityId`; three copies. Same-pass cleanup |
| MNT-10, MNT-15 | MNT | minor | Timeline restates the fold rule; `firstRevealed` can trap | Fix now → F-10 | Verified. MNT-15 is a nit but a crash in added code |
| MNT-11 | MNT | minor | `contentOffset` inputs; a fifth scroll-and-reflect pair | Fix now → F-10 (the pair); Rejected (the inputs) | Pair is a same-pass cleanup. The inputs are D-11's, and T-10 proves them at 2× |
| MNT-12 | MNT | minor | `macos-client-patterns.md` is stale | Fix now → F-11 | Verified: line 104 |
| MNT-13 | MNT | minor | Sources has no preview | Fix now → F-11 | Verified; manual item 3 needs one |
| UX-6 | UX | nit | An untitled Source speaks its ref twice | Fix now → F-2 | One line in the same expression |

Nits (logged): `CON-1 · PersonTimelineContentTests.swift:118 · test name` (F-5 renames it) · `CON-2 · PersonSourcesSection.swift:130 · row label on a 1×1 overlay; check with VoiceOver` · `MNT-16 · PVTitleText.swift:119 · mono font repeated in both arms`

### Fix plan 1

**F-1 — `PVInkLink` leaves focus and ids to its caller** (from MNT-1, MNT-6)
Files: `macos/App/DesignSystem/Components/InkLink/PVInkLink.swift`, `PVProvenanceRecordRow.swift`, `PersonSourcesSection.swift`, `docs/ideas/structural-debt.md`. Skill: `add-ui-component`.
Do: delete `@FocusState`, `.focused`, `identifier`, `accessibilityLabel` and `PVInkLinkSpokenLabel`. Draw the ring in the style body from `@Environment(\.isFocused)`, reusing what `PVFocusRingButtonBody` does. Return: `.pvActivatesOnReturn(when: true, action)` on the `Button`; `onKeyPress` fires only while focus is on it, so no Bool is needed. Both callers chain `.accessibilityLabel` and `.accessibilityIdentifier`. Close A-40 in `structural-debt.md`, or re-point it if any of it stays.
Check: T-5 passes; `rg 'FocusState' macos/App/DesignSystem/Components/InkLink` finds nothing; scripts/check-pr.sh PASS.

**F-2 — One spoken join** (from MNT-2, UX-6)
Files: `PersonSourcesContent.swift`, `L10n.swift`, `Localizable.xcstrings`, `PersonSourcesContentTests.swift`. Skill: `add-localized-string`.
Failing test first: T-2 — an untitled Source's label holds its ref once.
Do: `L10n.DesignSystem.spokenJoin([title.text, typeLabel ?? "", date, untitled ? "" : group.sourceRef]) ?? ""`. Delete `sourcesRowAccessibility`, `sourcesRowAccessibilityUntyped` and both catalog entries.
Check: T-2; `check-localizable-xcstrings.py`.

**F-3 — `trimmedNonEmpty`** (from MNT-3)
Files: `PersonSourcesContent.swift`.
Do: delete the private `nonEmpty`; call `String.trimmedNonEmpty` from `WorkspaceLocationLabels.swift`.
Check: `grep -rnE "func \w+\(_ \w+: String\) -> String\?" macos/App/Features/Conclusions` prints no trim helper; T-1, T-2 pass.

**F-4 — The graph link: focus, Return, one line** (from UX-1, UX-4)
Files: `PersonSourcesSection.swift`.
Do: on the line's link, `.pvFocusRing(focused.wrappedValue == line.id, cornerRadius: PVRadius.xs)` and `.pvActivatesOnReturn(when: focused.wrappedValue == line.id) { open(line.location) }`, as `PersonTimelineSection.swift:297`. Both refs: `.lineLimit(1)`, `.fixedSize(horizontal: true, vertical: false)`. Link: `.lineLimit(1)`, `.truncationMode(.tail)`, `.help(line.graphTitle)`. Subject name stays first to cut.
Check: scripts/check-pr.sh PASS; add these to the plan's manual check item 4.

**F-5 — The Timeline keeps its marker** (from CON-3, CON-1)
Files: `PersonTimelineContent.swift`, `PersonTimelineContentTests.swift`, the plan (T-11, Edges, Landed note), `completed.md` if it repeats the change.
Failing test first: `markerInsideHiddenRunStaysShown` (renamed from `markerInsideHiddenRunKeepsItsPlace`) — 60 dated events, a birth marker past the 50th → 50 event rows, `.more` with 10, the marker after `.more`, as at `4c642497`.
Do: walk `before`; the first hidden event row becomes `.more`, other hidden event rows drop, everything else stays. Keep `ConclusionFold.shown` and the shared `counted`. Correct the plan's T-11 and Edges, and say in the Landed note that the review reversed the refresh's T-11.
Check: the test; `git show 4c642497:macos/App/Features/Conclusions/PersonTimelineContent.swift` gives the same list for that input.

**F-6 — Measure the Source rows** (from SEC-1). No commit.
Do: a throwaway test: `PersonSourcesList` in an `NSHostingView` at 880pt, first `layoutSubtreeIfNeeded`, Debug build, three runs each at 14, 300 and 1,200 Sources with six lines each. Also time `PersonSourcesContent.init` at the same sizes. Do not commit it. Report the numbers.

**F-7 — One type lookup** (from MNT-4)
Files: `SourcesModel.swift`.
Do: `filterLabel` uses `types.label(forTypeID:)`.
Check: `grep -rn 'first { \$0.id == typeID }' macos/App` prints only `label(forTypeID:)`.

**F-8 — Unfold focus as the Timeline does it** (from MNT-5)
Files: `PersonSourcesSection.swift`.
Do: `expand(row.id); Task { @MainActor in focusedLine = row.firstRevealedLineID }`. Delete `pendingFocus` and the `.onChange`.
Check: scripts/check-pr.sh PASS. Plan's manual item 3 confirms it with the keyboard.

**F-9 — The name rule without the full sort** (from PRF-1, MNT-14)
Files: `EvidenceGraphSnapshot.swift`.
Do: filter to asserted naming rows with a value, then `min(by: EvidenceGraphSnapshot.observationDisplayOrder)`. Same result.
Check: T-4 and the existing snapshot tests pass.

**F-10 — Small cleanups** (from MNT-7, MNT-9, MNT-10, MNT-11, MNT-15)
Files: `PersonSourcesSection.swift`, `PersonFamilySection.swift`, `PersonTimelineSection.swift`, `PersonTimelineContent.swift`, `ConclusionShowMoreRow.swift`, `GraphCanvasScrollView.swift`.
Do: the four cell modifiers become `private`, named `personSources…Cell()`. Delete `.onChange(of: entityId) { self.expanded = [] }` in all three sections. Timeline: drop `where event.participants.count > limit`, keep the fold tuple, use its `.hidden`. `firstRevealed` uses `list.dropFirst(limit).first` under its guard. Extract the `clip.scroll(to:)` + `reflectScrolledClipView` pair; all five sites call it.
Check: `grep -n 'reflectScrolledClipView' macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift` prints one line; `grep -rn 'onChange(of: entityId)' macos/App/Features/Conclusions` prints nothing; T-9, T-10 pass.

**F-11 — Docs and previews** (from MNT-12, MNT-13)
Files: `docs/macos-client-patterns.md`, `PersonSourcesSection.swift`, `macos/App/Features/Conclusions/PersonSourcesPreviewData.swift` *(new)*.
Do: line 104 names Sources and `supportingQueryKeys`. `#Preview`s over `PersonSourcesList` with static content, as `PersonTimelineSection.swift:614` does: regular, untitled Source, folded.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0b62ae9c`. Gate PASS (orchestrator re-ran). F-1 to F-5 and F-7 to F-11 done. Deviations: F-10's `reflectScrolledClipView` grep also prints the override and its `super` call, and five sites now share `scrollClip(to:)`; D-10's marker wording rewritten to match the corrected T-11; STEP-5's `PVInkLink` signature text left as written, superseded by F-1.

F-6, `PersonSourcesList` in an `NSHostingView` at 880pt, six lines a Source, first `layoutSubtreeIfNeeded`, Debug build, three runs. 14 Sources: 226, 215, 202 ms. 300 Sources: 4,611, 5,597, 6,330 ms. 1,200 Sources: crashed in SwiftUI, "precondition failure: exhausted data space". `PersonSourcesContent.init`: 1.5 to 1.8 ms at 14, 27 to 28 ms at 300. Layout is the cost, 15 to 21 ms a Source. No Release figure. SEC-1 → To user.

## Round 2 — review
Reviewers: MNT, UX, SEC+PRF. CON skipped: it had no Fix-now blocker or major.

Round-1 Fix-now blockers and majors: MNT-1 to MNT-6 and UX-1 Resolved by their reviewers. Minors read by the orchestrator from the reviewers' checks: UX-4, UX-6, PRF-1, MNT-14 Resolved. SEC-1 Not resolved, raised to SEC-2.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-2 | SEC, PRF | blocker | Unbounded Source rows crash SwiftUI, and restore reopens the section | To user | Verified: `pageSection` is persisted in the location. F-6's crash is one Debug run in a throwaway test; no Release figure. A fold adds a state the board lacks; lazy rows risk Tab order |
| MNT-17 | MNT | blocker | The graph link's ring is added by the feature, against `PVButton`'s `.link` rule | To user | Verified: `PVButton.swift:287`. The fix changes every `.link` button's focus look, which a researcher can see, and reverses the kit's written rule |

**To user, answered (2026-11-05):** SEC-2: fold Source rows at 50. MNT-17: the ring goes in the kit, on every `.link` button. UX-2: keep as planned, Follow-up. UX-3, UX-5: dropped, no record.

### Fix plan 2

**F-12 — Fold the Source rows at 50** (from SEC-2, SEC-1)
Files: `PersonSourcesContent.swift`, `PersonSourcesSection.swift`, `PersonSourcesContentTests.swift`, `PersonSourcesPreviewData.swift`, the plan (D-10, Edges, Scale, T-13, manual item 3, Landed note).
Failing test first: T-13 `foldPastFiftySources` *(new)* — 60 Source groups → 50 rows, `hiddenRowCount` 10, `firstRevealedRowLineID` the first line of the 51st; 51 → 50, 1; 50 → 50, 0; `rowsExpanded: true` → 60, 0.
Do: `PersonSourcesContent.rowFoldLimit = 50`. A `rowsExpanded: Bool` input folds `rows` through `ConclusionFold.shown(limit: 50, minimumHidden: 1, expanded:)`, with `hiddenRowCount` and `firstRevealedRowLineID` from `ConclusionFold.firstRevealed`. `PersonSourcesSection` holds `rowsExpanded` as `@State`, so it resets on each mount and a restored section always opens folded. `PersonSourcesList` draws `ConclusionShowMoreRow(count:expand:)` after the last row, with the unfold focus move of F-8. The existing "Show N more" string; no new string. The section's header line counts every Source and record, not the folded ones.
Check: T-13; manual item 3 gains "a Person with more than 50 Sources".

**F-13 — A link button's focus ring lives in the kit** (from MNT-17)
Files: `macos/App/DesignSystem/Components/Button/PVButton.swift`, `PersonSourcesSection.swift`, `macos/App/DesignSystem/README.md` if it states the rule.
Do: `PVButtonBody`'s `.link` branch draws `.pvFocusRing(isFocused, cornerRadius: PVRadius.xs)`, always applied and outset by `PVElevation.focusRingWidth` so it clears the glyphs and takes no layout (pad out, ring, pad back). Rewrite the comment at `PVButton.swift:288`. Delete the `.pvFocusRing` on the graph link; `.pvActivatesOnReturn(when:)` stays. The plan's Landed note says every `.link` button gains the ring, by the user's choice.
Check: `rg 'pvFocusRing' macos/App/Features/Conclusions/PersonSourcesSection.swift` finds nothing; scripts/check-pr.sh PASS. Manual item 4 gains a `.link` button on the provenance sheet.

## Round 2 — fix
Head `b8786eec`. Gate PASS (orchestrator re-ran). F-12 and F-13 done; read by the orchestrator, no round 3. Deviations: a 60-Source fixture and a "Many Sources" preview added. Orchestrator corrected the stale InkLink row in `DesignSystem/README.md`.

## CI — fix round
CI run 42154834585 on `29247ec6`: `test` failed, `centerOnGridPoint` with 3 issues, `(2753.75, 1003.75)` for `(2750, 1000)`. Local gate passed. The 7.5pt gap at 2× is a 15pt legacy scroller on the CI host, so the test depends on the machine's scroller setting; `center(onGridPoint:)` itself centers in the clip it has.

**F-14 — T-10 does not depend on the host's scrollers** (from CI)
Files: `macos/ProvenenciaTests/GraphCanvasViewportControllerTests.swift`.
Do: the test's scroll view sets `hasVerticalScroller = false` and `hasHorizontalScroller = false`, so the clip is the full 1000×800 on any host. No product change.
Check: test passes locally; scripts/check-pr.sh PASS; CI green.

Head `a5bd4ab5`. Gate PASS (orchestrator re-ran). F-14 done; no product change.

## Result
Gate PASS · CI green on `b5cd9f33` · open: none · follow-ups: UX-2, PRF-2 · to user: SEC-2, MNT-17, UX-2, UX-3, UX-5 answered · CON miss: none
