# Ship log — S13-13 Timeline section

**Plan:** [`S13-13-timeline-section.md`](S13-13-timeline-section.md) · **Branch:** `s13-13-timeline-section` · **PR:** #691
**State:** ready

Base: `main` @ `92f87dba`. The plan's Status was "Blocked — on S13-33, S13-35" with its post-round-2 scope change not re-reviewed, so the loop stopped and the plan went back to plan-pr (user, 2026-11-02). The refresh checked every § 4 row at `92f87db`; review rounds 3 and 4 fixed seven blockers; Status Planned. The refresh is the branch's first commit, `71554946` (user: "Ride on S13-13 branch").

Follow-ups from planning: `EvidenceBridgeEdgeSummary.productKey(of:)` (observations) repeats the product-term check that `CatalogConclusionValue.productKey` (D-4) owns for conclusion values (round 3 M-34).

## Round 0 — implement
Head `dc2e0c86`. Gate PASS (orchestrator run against `92f87dba`: all nine checks). PR #691 (draft).
Report deviations: (1) `PersonFamilyContent.Tile` replaced by `PersonCardLink` everywhere, no typealias; (2) the line is one `Canvas` plus a non-focusable `Button` per point and cap for clicks; click-to-scroll from inside the page's `ScrollView`, tethers, `whenRequested` timing and Wide resize not run in the app (D-13 stop condition unproven either way; § 10 step 4); (3) two extra previews, "Full timeline, Wide" and "Large event, unfolded"; (4) previews not rendered; a hosted-view snapshot of Johann's list drew; (5) manual-check ACs not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: `macos/App/Features/**` models and lists, `PVPageColumns` scroll geometry; the timeline's events and participants come from the catalog unbounded in Go.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | Every scroll step re-runs the whole list body and redraws the whole lifeline | Fix now → F-1 | verified: `PersonTimelineList` keeps `sectionFrame` as `@State`, measured in `scrollSpace` (moves per scroll), and reads it in `body` to build `tethers`; the one `Canvas` takes `tethers` |
| UX-1 | UX | major | VoiceOver reads the event picture backwards | Fix now → F-2 | verified: `CachedThumbnail.tileBody` labels a raster only with `imageLabel`; the row passes none, and its fallback says "Event picture" |
| UX-2 | UX | major | "Show N more" by keyboard leaves focus nowhere | Fix now → F-3 | verified: the button is removed by its own action and no focus binding spans the list; the repo moves focus after such an action (`ConclusionDetailPage` `focusedID`). Family's same gap: Follow-up (its fold reveals tiles in a spatial diagram) |
| MNT-OL1 (→ PRF/UX) | MNT | minor | One click scrolls twice (row change, then request) | Fix now → F-4 | verified: `PVScrollRowIntoView` has `onChange(of: row)` and `onChange(of: request)`; `select` changes both |
| PRF-2 | PRF | minor | `PersonTimelineLifeline.init` is items × marks | Fix now → F-5 | code this PR adds; `marks.first` per target |
| MNT-1 | MNT | minor | `viewportHeight` and `viewportSize` both hold the height | Fix now → F-6 | same-pass, file this PR changes |
| MNT-2 | MNT | minor | Two hand-written `PersonCardLink` builders | Fix now → F-6 | same-pass cleanup: `PersonCardLink` is new; both copies touched |
| MNT-3 | MNT | minor | Role names derived twice in `row(...)` | Fix now → F-6 | same function, this PR adds it |
| MNT-4 | MNT | minor | The preview fixture computes spans in Swift | Fix now → F-7 | D-13 "Swift does no date arithmetic"; `swift-tests.mdc` |
| UX-3 | UX | minor | Role Badges read twice in a row | Fix now → F-2 | the row label already says "as …" |
| UX-4 | UX | minor | A lifeline point has no hover name | Fix now → F-8 | code this PR adds; the page names targets on hover |
| MNT-OL2 (→ UX) | MNT | minor | A placeholder repeats its visible text as its label | Fix now → F-2 | `PersonTimelineSection.swift` placeholder `Text` |
| CON-1 | CON | minor | Live-UI ACs need the manual check | Logged — the user's § 10 check | ship-pr: "The user does the manual check" |

Nits (logged): `CON-2 · DesignSystem/README.md TitleText row · say lineLimit is optional` · `MNT-5 · docs/macos-client-patterns.md:101 · "Family does" → Family and Timeline` · `MNT-6 · PersonTimelineLifeline.swift:217 · mark(_:) comment` · `MNT-7 · PersonTimelineContent.swift:160 · fold patches by id string` · `MNT-8 · PersonTimelineSection top(_:) and test y(_:) same switch` · `MNT-9 · PersonTimelineList could be private` · `MNT-10 · ConclusionPageSectionHost hardcodes persons.detail. root` · `UX-5 · timelineEmptyMessage repeats "events"` · `UX-6 · cut title has no .help`

### Fix plan 1

**F-1 — Only the tethers follow the scroll** (from PRF-1)
Files: `macos/App/Features/Conclusions/PersonTimelineSection.swift`.
Do: move the `sectionFrame` measurement (`onGeometryChange` in `PVPageColumnsMetrics.scrollSpace`), the `visible` rect and the `lifeline.tethers(…)` call out of `PersonTimelineList.body` into a small private overlay view that owns that `@State`, so a scroll step invalidates only it. Split the drawing in two `Canvas` layers: the line, stubs, ticks and marks read only `lifeline` (and redraw when it changes); a second layer draws the tethers and the in-view accent points, and is the only one fed by the scroll. The per-point click buttons stay with the static layer. The list's `body` reads no scroll-dependent state.
Check: `Self._printChanges()` (temporarily) in `PersonTimelineEventRow.body` prints nothing while the "Long life" preview scrolls, then removed; PersonTimeline tests pass; scripts/check-pr.sh PASS.

**F-2 — VoiceOver reads each row once** (from UX-1, UX-3, MNT-OL2)
Files: `PersonTimelineSection.swift`.
Do: `.accessibilityHidden(true)` on the row's `CachedThumbnail` (the row's spoken label names the event, as the Events list does) and on the role Badges' `HStack` (the label ends "as …"); drop the placeholder `Text`'s `.accessibilityLabel` that repeats its visible text. D-6's call is otherwise unchanged; note the hidden thumbnail in the plan's Landed note.
Check: scripts/check-pr.sh PASS.

**F-3 — Focus lands on the first revealed item after "Show N more"** (from UX-2)
Files: `PersonTimelineSection.swift`.
Do: a `@FocusState private var focused: String?` in `PersonTimelineList`, bound with `.focused($focused, equals:)` on each title button (the row id) and each card (`PersonCardLink.id`). `expand(fold)` sets it, on the next update after the content has the new items, to the first item revealed: the 51st dated row's title for `.dated`, the 51st undated row's title for `.undated`, the 16th card's id for `.participants(eventId:)`. Keep Return working on the title via `pvActivatesOnReturn(when: focused == row.id)`. Compute "the first revealed id" in `PersonTimelineContent` as a pure func `firstRevealed(by fold: Fold) -> String?` (from the unfolded content) with a test case in `PersonTimelineContentTests.foldsAndOrder`.
Failing test first: `PersonTimelineContentTests/foldsAndOrder` — `lise.firstRevealed(by: .dated)` is the 51st dated row's id; `largeEvent.firstRevealed(by: .participants(eventId:))` is the 16th participant's link id.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — One click, one scroll** (from MNT-OL1)
Files: `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift`.
Do: when the caller passes `whenRequested` (non-nil), a change of `row` alone does not scroll; only the request token's change does (on the next main-actor turn, as now). Callers that pass no token behave as today. Say so in the doc comment and the README row "ScrollRowIntoView".
Check: existing `pvScrollsRowIntoView` callers' tests pass unchanged; scripts/check-pr.sh PASS.

**F-5 — Lifeline lookups by dictionary** (from PRF-2)
Files: `macos/App/Features/Conclusions/PersonTimelineLifeline.swift`.
Do: build a `[String: CGFloat]` of point y's from `marks` once in `init`; `targets` and `mark(_:)` read it.
Check: `PersonTimelineLifelineTests` pass unchanged.

**F-6 — One builder each for the card link, role names and viewport** (from MNT-1, MNT-2, MNT-3)
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Conclusions/PersonCardLink.swift`, `PersonFamilyContent.swift`, `PersonTimelineContent.swift`.
Do: delete `viewportHeight`; `refreshScrolled()` reads `viewportSize.height` (still only on the `measures` path). Add `static func link(id:summary:role:opensOn section: ConclusionPageSection, locale:orders:) -> PersonCardLink` on `PersonCardLink` (`isFocus: false`, `PersonsList.location(summary, orders:).withPageSection(section)`); Family's relative tiles and the Timeline's participants call it; Family's focus tile stays explicit. One private `roleNames(_ values: [CatalogConclusionValue]) -> [String]` in `PersonTimelineContent` (the `PropertyTermDisplay.name(_, propertyKey: SeededPropertyKey.role)` compactMap), read by the Badges (pair with `productKey` for `isSubject`) and the participant role.
Check: `rg -n "PVPersonCardContent\(summary:" macos/App/Features/Conclusions` prints only `PersonCardLink.swift` and Family's focus tile; `rg -n "PropertyTermDisplay.name\([^)]*SeededPropertyKey.role" macos/App` prints one line; `rg -n viewportHeight macos/App/DesignSystem/Components/PageColumns` prints nothing; PersonFamilyContentTests, PVPageColumnsTests, PersonTimeline tests pass; scripts/check-pr.sh PASS.

**F-7 — Fixture spans as literals** (from MNT-4)
Files: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift`.
Do: replace `When.span`'s arithmetic with each event's `CatalogTimelineSpan` written as a literal beside its date, with values from S13-33's Go rule (a day `[y + d/365, y + (d+1)/365)` from its day-of-year, a month from its first to last day, a year `[y, y+1)`, "bef." `openStart` at its known year); one comment names the Go test the values follow (`core/database/conclusionsummaries` `TestPersonEventsLifeline`). No date arithmetic left in the file.
Check: `rg -n "fraction\(|/ 365|365\.0" macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` prints nothing; PersonTimeline tests pass unchanged.

**F-8 — Each lifeline point names its event on hover** (from UX-4)
Files: `PersonTimelineContent.swift` (`Span.spokenLabel` *(new)*: the row's spoken label), `PersonTimelineLifeline.swift` (carry it on `Mark` or a lookup), `PersonTimelineSection.swift`.
Do: `.help(spokenLabel)` on each point's and cap's click button; a placeholder cap's help is its text ("Birth unknown").
Check: a `PersonTimelineContentTests` case: every `allSpans` entry's `spokenLabel` equals its row's; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `ea6f398d`. Gate FAIL once on `PVSegmentedControlTests.selectSendsTheItemValue()` (timed out at `waitUntil { chosen.isLaidOut }`, a host-window timing test this PR does not touch; it passed at round 0), then PASS on retry. F-1 – F-8 done. Deviations: F-1's `_printChanges` scroll check not run (no scrolling preview here); F-7 keeps `y + 1` for generated year spans; F-1 draws in-view accent points over the grey ones.
Read by the orchestrator: MNT-1, MNT-2, MNT-3, PRF-2 **Resolved** (F-6's three `rg` checks print only the allowed lines; `marksByID` in `PersonTimelineLifeline.init`). MNT-4 **Not resolved** in one respect: the implementer found that Go's `datevalue.YearLine` divides by the year's own days (366 in a leap year), so the plan's "S13-33 divides every year by 365" (D-13, X-52's premise) is false, and the fixture's 1848 and 1880 day spans and two comments say otherwise. The dot threshold still holds (a leap-year day is 1/366). → F-9.

### Fix plan 1 (addendum)

**F-9 — Day spans and comments follow Go's year line** (from MNT-4, found in the fix round)
Files: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift`, `PersonTimelineLifeline.swift`, `docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` (Landed note).
Do: the fixture's day and month literals use `core/datevalue/calendar.go` `YearLine`'s rule (the fraction over the year's own day count, 366 in a leap year) for 1848 and 1880 and any other leap year in the file. `PersonTimelineLifeline`'s `dayInYears` comment says a day is at most 1/365 of a year (1/366 in a leap year), so the threshold catches both. The Landed note names the plan's wrong "divides every year by 365" premise.
Check: `rg -n "365, leap years too|divides every year by 365" macos/App` prints nothing; PersonTimeline tests pass; scripts/check-pr.sh PASS.

## Round 1 — fix (addendum)
Head `15c7abf4`. Gate PASS. F-9 done: leap-year day spans (1844-06-04, 1848-02-14, 1880-09-18, 1880-09-22) over 366 per `YearLine`; both comments corrected; Landed note (4) names the plan's wrong premise. MNT-4 **Resolved**.

## Round 2 — review
Reviewers: UX, SEC+PRF (each had a Fix-now major). Scope `dc2e0c86...15c7abf4`. CON and MNT had no major: their minors read above.
- PRF-1 **Resolved**: `PersonTimelineTethers` owns `sectionFrame` and the viewport; the list's `body` reads nothing that moves on scroll; the static `Canvas` takes only `lifeline` and labels.
- UX-1 **Resolved** (thumbnail and Badges hidden; the row speaks once). UX-2 **Resolved** in code (`@FocusState` over titles and cards; `unfold` focuses `firstRevealed`); the participant-card path carries a second focus binding inside `PVTileCard`'s Return modifier, so the manual check confirms focus lands on the 16th card.
- New blockers: none. Outside-lens note (→ MNT, minor): card ids `"\(eventId)-\(personId)"` share the focus namespace with row ids; a crafted id could collide, the worst case focus on the wrong control → Follow-up.

## Result
Gate PASS · CI green (run 41563143394 on `d87045a1`, after one CI fix round: F-10) · open: none · follow-ups: Family's "Show N more" focus gap (UX-2's Family half); `EvidenceBridgeEdgeSummary.productKey(of:)` repeats the product-term check (planning M-34); card and row ids share one focus namespace (round 2 SEC+PRF note); nits CON-2, MNT-5 – MNT-10, UX-5, UX-6 · to user: — · CON miss: none

## CI — fix
CI `test` FAIL on `233f1ebf` (run 41562001555): the CI compiler could not type-check `PersonTimelineLifelineTests.swift:44` (`let perYear = 2730 / (1880 + 265.0 / 365 + 1.0 / 365 - (1817 + 70.0 / 365))`) in reasonable time; the build failed. Local gate PASS (faster machine). The literal also still reads 1880 over 365, against F-9.

**F-10 — Test arithmetic the CI compiler can type-check** (CI gate failure)
Files: `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift`.
Do: replace the expression with typed steps read from the fixture, not re-derived: `let burialEnd: Double = 1880.7267759562842` and `let birthStart: Double = 1817.1917808219177` (the fixture's literals for events 11 and 1, as `PersonTimelinePreviewData` holds them), `let years: Double = burialEnd - birthStart`, `let perYear: Double = 2730 / years`. Scan the test files this PR adds for any other untyped multi-literal arithmetic of the same shape and type it the same way.
Check: `rg -n "\d+\.0 / 365" macos/ProvenenciaTests/PersonTimeline*` prints nothing; scripts/check-pr.sh PASS; CI green.

Head `d87045a1`. Gate PASS. F-10 done (the span named `span`: `years` is taken on the next line). CI green: `localizable`, `test` (run 41563143394).
