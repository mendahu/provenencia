# Ship log — S13-10 Header and sections

**Plan:** [`S13-10-header-and-sections.md`](S13-10-header-and-sections.md) · **Branch:** `s13-10-header-and-sections` · **PR:** #634
**State:** ready

## Round 0 — implement
Base `main` @ `1253dfaa`. Head `3e5ae61f`. Gate PASS (orchestrator run). Report: §10 manual check not run (AC-1.5 – 1.8, 1.10, 2.6, 2.9, 3.1 – 3.6 unverified in the app). Deviations: (1) sizing statics on non-generic `PVPageColumnsMetrics`; (2) `pageSectionLeaf` uses numbered placeholders; (3) `PersonDetailLifeLines` uses `AnyLayout`; (4) loading/failed page uses header-less `PVPageColumns(primary:)`; (5) `pageSection` is `private(set)` behind `withPageSection(_:)`; (6) T-16 clicks with `NSEvent`; (7) commit 907a21c5 does not build alone; (8) a broad `git add docs` was denied, re-run with explicit paths.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC+PRF routed for `macos/App/Features/**` models, per-scroll compaction, and history decode of the new `pageSection`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1, CON-1, UX-1 | PRF, CON, UX | blocker | Pinned header oscillates between compact and full for overflow ≈ 120 – 236pt | Fix now → F-1 | verified: `PVPageColumnsMetrics.isScrolled` applies `overflow >= compactMinOverflow` on every call, and `refreshScrolled` feeds the overflow measured under the current header; compacting grows the viewport by Δ ≈ 114 – 116pt (three independent models agree). D-17's rationale (X2-3) assumed the guard sees full-header overflow; the fix serves D-17's stated intent and AC-1.10; user 2026-10-29 approved "guard on entry" |
| UX-2, MNT-1 | UX, MNT | major | `PVChipStyle` drops disabled dimming and press feedback | Fix now → F-2 | verified: `PVChipStyleBody` is only `configuration.label.pvFocusRing(…)`; no `PVHoverEffect`, no `isEnabled`. Composer locked kind chip looks enabled (AC-4.1) |
| MNT-2 | MNT | major | "section" names the sidebar section, the page section and the Place block | To user → Fix now → F-7 | D-2 and D-5 chose `section:`; user 2026-10-29 chose "Rename now" (supersedes those names) |
| MNT-3 | MNT | minor | `PersonDetailSections` re-unwraps `context.section` | Fix now → F-3 | verified in `PersonDetailView.swift:41,64` |
| MNT-4 | MNT | minor | Header band's `pinned` duplicates `section != nil` | Fix now → F-3 | `pinned:` on the band is not in a Decision (only `headerPinned` on `PVPageColumns`, D-14) |
| MNT-5 | MNT | minor | Page background/a11y chrome written twice | Fix now → F-3 | file this PR changes |
| MNT-6 | MNT | minor | Two Layouts copy the proposal → widths code | Fix now → F-4 | file this PR adds |
| UX-3 | UX | minor | ⌘1 – ⌘5 not discoverable | Fix now → F-5 | file this PR adds; a `.help` does not touch D-9's view-scoped shortcuts |
| MNT-out-1 (→ CON) | CON | minor | `PVSegmentedControlTests` waits a fixed 200ms | Fix now → F-6 | `PVSegmentedControlTests.swift:44,55`; `waitUntil` exists in `QueryHandleWaiting.swift:24` |
| MNT-7 | MNT | nit → minor | DesignSystem README PVSelect paragraph still says "segmented `PVChip` groups" | Fix now → F-6 | stale comment in a file this PR changes |
| CON-2 | CON | minor | No hosted test for page-level section glue | Rejected | T-9 assigns these ACs to the §10 manual check, which goes to the user; place, crumbs, widths and control are unit-tested |

Nits (logged): `CON-3 · ConclusionDetailPage.swift ConclusionDetailLoaded.body · .accessibilityElement(children: .contain) not named by the plan` · `CON-4 · PersonDetailView.swift:5 · rewrap doc comment` · `UX-4 · DateValueEditorForm.swift:36-55 · VoiceOver reads "Kind" twice` · `MNT-8 · PVScrollRowIntoView.swift:678 · restart id shares the row's generic ID`

### Fix plan 1

**F-1 — The compact header cannot undo itself** (from PRF-1, CON-1, UX-1)
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`.
Failing test first: `PVPageColumnsTests.compactingDoesNotUndoItself` — for overflow in stride(120..<236, by: 4) with Δ = 116: `let c = isScrolled(offset: 50, overflow: o, wasScrolled: false)`; `#expect(isScrolled(offset: 50, overflow: o - 116, wasScrolled: c) == c)`. Add T-15 rows: (50, 70, wasScrolled: true) → true; (50, 0, true) → false; (50, 70, false) → false; (0, 300, true) → false.
Do: in `PVPageColumnsMetrics.isScrolled`, apply `compactMinOverflow` only when entering: `if !wasScrolled && overflow < compactMinOverflow { return false }`; `if wasScrolled && overflow <= 0 { return false }`; then the existing offset rules. Update the doc comment. Supersedes D-17's literal "false when overflow < 120" (user, 2026-10-29); record it in the plan's Landed note.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Chips keep disabled dimming and press feedback** (from UX-2, MNT-1)
Files: `macos/App/DesignSystem/Components/Chip/PVChip.swift`, `macos/ProvenenciaTests/PVSegmentedControlTests.swift` if a test fits.
Do: build `PVChipStyleBody` like `PVButtonBody` (`PVButton.swift:265-272`): `PVHoverEffect(isPressed: configuration.isPressed) { _ in configuration.label }.pvFocusRing(isFocused, cornerRadius:)`, keeping the focus ring outside. Do not add a parameter to `PVHoverEffect` unless the existing one cannot be composed.
Check: builds; existing `PVSegmentedControlTests.disabledItemDoesNotSelect` passes; scripts/check-pr.sh PASS.

**F-3 — Page glue tidy-ups** (from MNT-3, MNT-4, MNT-5)
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`.
Do: (a) `PersonDetailSections`: bind `let section = context.pageSection ?? .vitals` once and switch on it; the `.family, .timeline, .sources` branch uses `section` directly with no `if let`. (b) `ConclusionDetailHeaderBand`: drop `pinned`; derive it from `section != nil`; caller passes `compact: isScrolled`. (c) Apply `.background(PVColor.surfacePage).accessibilityElement(children: .contain).accessibilityIdentifier(pageIdentifier)` once on the outer view in `ConclusionDetailPage.body`, removing the copy in `ConclusionDetailLoaded.body`; build the header-less loading `PVPageColumns(primary:)` in one place.
Check: `PersonDetailTests`, `ConclusionDetail*` suites pass; T-10 grep still holds; scripts/check-pr.sh PASS.

**F-4 — One proposal → widths helper** (from MNT-6)
Files: `PVPageColumns.swift`, `PVPageColumnsTests.swift`.
Do: add `PVPageColumnsMetrics.columnWidths(proposal: ProposedViewSize, widthClass:, hasSecondary:)` returning the pane and the widths (finite-width check and the `widthContentMax + 2 * gutterPage` fallback in one place); call it from `PVPageColumnsLayout` and `PVPageColumnsBandLayout`. Both layouts decide `hasSecondary` the same way (pick one, say which in the report).
Check: `PVPageColumnsTests` pass; `grep -n "isFinite" PVPageColumns.swift` prints one line; scripts/check-pr.sh PASS.

**F-5 — Name the section shortcut in a tooltip** (from UX-3)
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: when an item has a `shortcut`, add `.help` from one new key, format `%1$@ (⌘%2$@)` (label, key). No app-menu command (D-9 stands).
Check: localizable check passes; scripts/check-pr.sh PASS.

**F-6 — Test wait and stale README line** (from MNT-out-1, MNT-7)
Files: `macos/ProvenenciaTests/PVSegmentedControlTests.swift`, `macos/App/DesignSystem/README.md`.
Do: replace the two `Task.sleep(for: .milliseconds(200))` with `waitUntil { … }` on the hosted layout being ready (`QueryHandleWaiting.swift:24`). README ~line 407: point the PVSelect paragraph at `PVSegmentedControl` instead of "segmented `PVChip` groups".
Check: `grep -n "Task.sleep" macos/ProvenenciaTests/PVSegmentedControlTests.swift` prints nothing; scripts/check-pr.sh PASS.

**F-7 — Call the page section `pageSection` everywhere** (from MNT-2; user 2026-10-29, supersedes D-2/D-5 names)
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `ConclusionDetailPage.swift`, `PersonDetailView.swift`, `WorkspaceToolbar.swift` if it uses the label, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift`, `docs/macos-client-patterns.md` if it names them.
Do: `personDetail(entityId:ref:title:pageSection:)`; `ConclusionDetailPageSectionsContext.pageSection`; in `ConclusionDetailLoaded` and `ConclusionDetailHeaderBand`, `pageSection` and `offeredPageSections` for the page-section values. Leave `WorkspaceLocation.section` (WorkspaceSection) and `ConclusionDetailSection` alone. Do F-7 before F-3 so F-3 uses the new names. Add both D-2/D-5 name changes and F-1's D-17 change to the plan's Landed note.
Check: `rg -n 'context\.section|personDetail\([^)]*section:' macos` prints nothing; scripts/check-pr.sh PASS.

Order: F-1, F-2, F-7, F-3, F-4, F-5, F-6.

## Round 1 — fix
Head `45b2d7dd`. Gate PASS (orchestrator run). F-1 – F-7 done. F-1's `compactingDoesNotUndoItself` ran red on the old guard first. F-4 added `unboundedProposalSizesToTheGroup`; both layouts take `hasSecondary` from `PVPageColumns`. Deviations: the implementer's first gate failed (a missed `section:` call site, its own new test's premise), fixed in `693151ca`, `4b8aff1c`; `dbfd86c2` does not build alone; `4b8aff1c` lacks the Co-Authored-By line; two graph-progress `waitUntil` tests recorded issues once on a loaded machine and did not repeat in two later full runs (not changed).

## Round 2 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF — each had a Fix-now blocker or major in round 1. Scope `3e5ae61f...45b2d7dd`.

Round-1 items: PRF-1/CON-1/UX-1 Resolved for a one-line header (Δ ≈ 116) — reopened as UX-5/PRF-2 below. UX-2/MNT-1 Resolved. MNT-2 Resolved. Minors MNT-3, MNT-4, MNT-5, MNT-6, UX-3, MNT-out-1, MNT-7: Resolved (CON round 2 traced each hunk to F-3 – F-6; orchestrator read `ConclusionDetailPage.swift`, `PVPageColumns.swift`, `PVSegmentedControl.swift`, `PVSegmentedControlTests.swift` diffs).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-5, PRF-2 | UX, PRF | blocker | Compact header still undoes itself when the full header is >120pt taller than the compact one (wrapped name) | Fix now → F-8 | verifier Confirmed at `45b2d7dd`: title (`ConclusionDetailPage.swift:596-606`) and life lines wrap (no `lineLimit`); Δ from font `hhea` metrics ≈ 115 (one line) to ≈ 150 (two-line title + badge) to ≈ 171 (plus a wrapped life line); `isScrolled` line 69 expands on `overflow <= 0` regardless of offset; `compactingDoesNotUndoItself` hard-codes Δ = 116 |

Nits (logged): `CON-5, MNT-9 · PVSegmentedControl.swift:101 · .help("") on segments with no shortcut` · `MNT-10 · PVSegmentedControl.swift:33 · switch title repeated in chip(_:)` · `MNT-11 · S13-10 plan Landed note says "four places", lists six` · `MNT-12 · WorkspaceNavigationTests.swift:759, PersonDetailView.swift:41 · locals named section for a page section`

### Fix plan 2

**F-8 — Enter compact only when the measured header shrink leaves overflow** (from UX-5, PRF-2)
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`, the plan's Landed note.
Failing test first: generalize `compactingDoesNotUndoItself` over Δ ∈ {116, 150, 171}: for each Δ and each overflow in stride(120..<Δ+40, by: 4): `let c = isScrolled(offset: 50, overflow: o, wasScrolled: false, bandShrink: Δ)`; `#expect(isScrolled(offset: 50, overflow: o - Δ, wasScrolled: c, bandShrink: Δ) == c)`. Add T-15 rows: (50, 130, false, Δ 150) → false; (50, 170, false, Δ 150) → true; (50, 130, false, Δ 116) → true. Run red against the current signature/logic first (the Δ = 150 case fails).
Do:
- `PVPageColumnsMetrics.isScrolled(offset:overflow:wasScrolled:bandShrink:)` — keep it pure. Entering (`!wasScrolled`) needs `overflow >= max(compactMinOverflow, bandShrink + compactOffset)`; the existing `wasScrolled && overflow <= 0` exit and offset hysteresis stay. Doc comment says why.
- `PVPageColumns` measures the pinned header band's height in **both** forms at the band's current width: the shown form with `onGeometryChange` (as viewport and content are measured), and the other form drawn `.hidden()` and `.accessibilityHidden(true)` behind the band (e.g. in `.background`, sized `.fixedSize(horizontal: false, vertical: true)` at the band's width) and measured the same way. `bandShrink = fullHeight − compactHeight`, clamped ≥ 0; `refreshScrolled` passes it and also runs when either height changes. Unpinned pages are unchanged.
- If the hidden form cannot be measured at the band's width, fall back to remembering the last measured height of each form (bandShrink = `compactMinOverflow` until both have been seen) and say so in the report; do not invent a third approach.
- Hosted test if feasible: host a pinned `PVPageColumns` whose header is `Color.clear.frame(height: isScrolled ? 50 : 200)` and read back the measured `bandShrink` (150) through `PVPageColumnsMetrics` or a test hook; if the hosted view cannot report it, say so.
- Plan Landed note: amend item (6) to say entry uses the measured shrink, and fix the opening count (MNT-11).
Check: the tests pass; `rg -n "overflow - 116" macos/ProvenenciaTests` prints only the Δ = 116 parameter, not a hard-coded loop; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `0f054cb0`. Gate PASS (orchestrator run). F-8 done: `compactingDoesNotUndoItself(bandShrink:)` over Δ 116/150/171 ran red first (Δ 171), passes; T-15 rows added; Landed note item (6) and the count fixed (MNT-11). Deviations: both header forms are drawn hidden in the band's `.background` (`measuredBand(compact:)`), not the shown form plus one hidden form, to avoid a stale-`isScrolled` slot swap; no hosted test (`bandShrink` is private `@State`); the graph-progress tests flaked once more on the implementer's first run and passed on rerun.

Orchestrator read of `0f054cb0`:

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| ORC-1 | UX, PRF | major (unverified) | The two hidden measured copies of the header band include the section switcher, so ⌘1 – ⌘5 are registered three times, and the thumb and its context menu are drawn three times | To user | `ConclusionDetailPage.swift:303-304,365` — the `header:` builder is `ConclusionDetailHeaderBand`, which draws `switcher(selection:)` → `PVSegmentedControl` → `.keyboardShortcut` (`PVSegmentedControl.swift:143`). `.hidden()` + `.allowsHitTesting(false)` do not disable keyboard shortcuts with certainty; not tested. No round 3 without the user. User 2026-10-29: "One small fix, then finish" → F-9; thumbs drawn three times → Follow-up. F-9 Resolved (orchestrator read `0f054cb0..32c9f390`) |

### Fix plan 2b (user-approved, no re-review; orchestrator checks the diff)

**F-9 — Hidden measured copies cannot act** (from ORC-1)
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`.
Do: in `measuredBand(compact:)`, add `.disabled(true)` beside `.hidden()` / `.accessibilityHidden(true)` / `.allowsHitTesting(false)`, and say in the comment that the copies must not register the switcher's keyboard shortcuts. Nothing else.
Check: scripts/check-pr.sh PASS.

## Round 2b — fix
Head `32c9f390`. Gate PASS (orchestrator run). F-9 done: `.disabled(true)` on `measuredBand(compact:)`, three lines, only `PVPageColumns.swift`. `origin/main` still `1253dfaa`; no merge needed.

## Result
Gate PASS · CI pending · open: none (§10 manual check is the user's) · follow-ups: ORC-1 (header drawn three times by the measured copies), nits CON-3, CON-4, CON-5/MNT-9, MNT-8, MNT-10, MNT-11 (fixed), MNT-12, UX-4; graph-progress `waitUntil` test flake · to user: MNT-2 (rename — yes), D-17 guard (yes), ORC-1 (one small fix — yes) · CON miss: none
