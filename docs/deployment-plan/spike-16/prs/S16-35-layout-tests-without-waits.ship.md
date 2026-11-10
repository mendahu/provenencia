# Ship log — S16-35 Layout tests compute their frames instead of waiting on a window

**Plan:** [`S16-35-layout-tests-without-waits.md`](S16-35-layout-tests-without-waits.md) · **Branch:** `plan/s16-35-layout-tests-without-waits` · **PR:** [#775](https://github.com/mendahu/provenencia/pull/775)
**State:** fixing 1

## Before round 0

- The user asked to ship S16-35 on the planning PR (#775) rather than a new branch (2026-11-10). The plan's **Branch** line now names `plan/s16-35-layout-tests-without-waits`; no Decision, Scope or AC changed.
- Branch base: `main` @ `db5885df` (the plan's planning base); `origin/main` has not moved.

## Round 0 — implement
Head `d7bbf6cb` (code head `68b4c935`). Gate PASS (orchestrator run against `db5885df`, load 7.4 / 11.9 / 10.9). Report deviations: T-5 is one test looping over two notices; four code commits rather than one per STEP (STEP-1 to STEP-3 share a file); the before runs used the branch head (plan docs only on top of `db5885df`); manual check 3 not run (no UI).

D-8 run tables (implementer, one `xcodebuild` at a time; load = 1/5/15-min before each run):

| When | Run | Result | Duration | Load | Failing tests |
| --- | --- | --- | --- | --- | --- |
| Before | 1 | pass | 153 s (with build) | 4.10 3.87 4.92 | — |
| Before | 2 | FAIL | 39 s | 20.92 14.55 9.33 | `SourcePageOverviewLayoutTests` ×3, `PVPageColumnsTests.wideHostPlacesTheSecondaryBeside` |
| Before | 3 | pass | 28 s | 14.60 13.73 9.27 | — |
| Before | 4 | pass | 27 s | 11.03 12.97 9.15 | — |
| Before | 5 | pass | 27 s | 14.14 13.47 9.44 | — |
| After | 1 | pass | 28 s | 17.38 17.20 12.06 | — |
| After | 2 | pass | 25 s | 13.09 16.22 11.85 | — |
| After | 3 | pass | 24 s | 10.54 15.36 11.67 | — |
| After | 4 | pass | 25 s | 9.00 14.61 11.50 | — |
| After | 5 | pass | 25 s | 8.53 14.05 11.39 | — |

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`). Skipped: SEC+PRF — no catalog data, store, list, cache or model code; the diff moves layout arithmetic into static functions that do the same work per layout pass, and deletes tests.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | Two copies of the `Subviews` → `fit` closure; D-7 tells every new layout to write a third | Fix now → F-1 | verified: `SourcePageView.swift:384` (inside `placed`) and `PVPageColumns.swift:418` (`fit(_:)`) hold the same closure; same-pass cleanup — this PR adds both copies |
| MNT-2 | MNT | minor | `add-swift-test` lists `PersonTimelinePlacement` as a `static func arrange(width:…fit:)`; it has none | Fix now → F-2 | verified: `PersonTimelinePlacement.swift` has `init(content:heights:viewportHeight:pointsPerYear:)` and helpers, no `arrange`; the example came from the plan's D-7 (the orchestrator's error, from plan review M-3). The fix keeps D-7's example and labels it as the other pure-placement shape, so no Decision changes |

Nits (logged): `MNT-3 · WorkspaceDestinationHostTests.swift:2 · unused import SwiftUI` (rides F-3) · `MNT-4 · PVPageColumns.swift:424-425 · pane(proposal, w) computed twice` (rides F-3) · `MNT-5 · SourcePageView.swift:380 · placed's doc repeats arrange's` (rides F-3) · `CON-1 · SourcePageOverviewLayoutTests.swift:7 · test helper doc says "sizes" but returns frames` (rides F-3) · `CON-2 · macos/App/DesignSystem/README.md:192 · could name arrange, layoutWidths, groupX` · Outside lens (MNT→CON): `WorkspaceNavigationTests.swift:656 · T-5's attachedNavigation() temp dirs are not removed; the file's other tests do the same` — logged, the pattern predates this PR.

UX: no findings (clean). CON: no findings above nit (clean). No blocker or major in any lens, so no round 2.

### Fix plan 1

**F-1 — One `Subviews` → `fit` adapter** (from MNT-1)
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `.agents/skills/add-swift-test/SKILL.md`.
Failing test first: none (refactor; T-2 and T-3 pin the results).
Do: add `extension LayoutSubviews { func fit(_ index: Int, _ width: CGFloat) -> CGSize }` returning `self[index].sizeThatFits(ProposedViewSize(width: width, height: nil))`, in `PVPageColumns.swift` beside `PVPageColumnsLayout` with a one-line doc ("a page `Layout`'s `fit` for its static `arrange`"). `PVPageColumnsLayout` drops its private `fit(_:)` and passes `subviews.fit`; `SourcePageOverviewLayout.placed` passes `subviews.fit`. In the D-7 bullet of `add-swift-test`, say the `Layout` passes `subviews.fit` to its `arrange`.
Check: `rg -n 'sizeThatFits\(ProposedViewSize\(width: \$1' macos/App` prints nothing; `rg -n 'func fit\(' macos/App` prints only the extension; T-2, T-3 pass; `check-guidance-names.py` passes.

**F-2 — The skill's example list says what each example is** (from MNT-2)
Files: `.agents/skills/add-swift-test/SKILL.md`.
Do: the examples of `static func arrange(width:…fit:)` are `SourcePageOverviewLayout.arrange` and `PVPageColumnsLayout.arrange`; `PersonTimelinePlacement` follows as "or a pure placement value built from measured sizes (`PersonTimelinePlacement`)".
Check: `check-guidance-names.py` passes.

**F-3 — Nits in files this PR changes** (MNT-3, MNT-4, MNT-5, CON-1)
Files: `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift`, `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift`.
Do: drop `import SwiftUI` from `WorkspaceDestinationHostTests.swift` if the file builds without it; bind `pane(proposal, w)` once in `PVPageColumnsLayout.sizeThatFits`; `placed(width:subviews:)`'s doc: "The `Subviews` adapter for `arrange`: no frames unless there are four subviews."; the test helper's doc: "Arranges four subviews of these heights at `width`."
Check: `scripts/check-pr.sh` PASS. Then one line in the plan's Landed note for F-1 (the shared adapter) and F-2.
