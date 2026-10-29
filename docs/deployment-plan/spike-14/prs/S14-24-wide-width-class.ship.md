# Ship log — S14-24 Give every workspace page its width class

**Plan:** [`S14-24-wide-width-class.md`](S14-24-wide-width-class.md) · **Branch:** `s14-24-wide-width-class` · **PR:** #628
**State:** ready

## Round 0 — implement
Base `main` @ `08eb109e`. Head `1178012c`. Gate PASS (orchestrator run; xcode-test 77s). Report: no deviations; AC-2.1/AC-2.2 manual steps not run (T-2, T-3 stand in per plan §10). Base check run against `08eb109e`, not the Planned `b8ca4cfc`; `git diff b8ca4cfc origin/main -- macos` empty.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX. Skipped: SEC+PRF — no security path, no catalog-loaded data; the only new per-frame work maps pane geometry to a two-case enum whose `onGeometryChange` action fires only when the class flips, and the host is a view, not a `Features/**` model.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 (+ MNT outside-lens) | CON | major → minor | Plan has no Landed note | Fix now → F-1 | verified: plan diff changes only Status. Re-rated minor: a missing "no differences" line in a doc, with no code effect; the implementer's report says none |
| MNT-1 | MNT | minor | Test window kept alive by `_ =`, not `withExtendedLifetime` | Fix now → F-2 | verified: `PVSelectPlacementTests.swift:146,164` put `_ = window` before layout; `WorkspaceDestinationHostTests` uses `withExtendedLifetime` |
| MNT-3 | MNT | minor | `pvMeasuresWidthClass()` doc doesn't say only the host applies it | Fix now → F-3 | verified: `PVWidthClass.swift:43-45` |
| CON-2, MNT-2 | CON, MNT | minor | T-3's first mount repeats T-2's | Rejected | plan §8 rows T-2 and T-3 specify these two mounts as written (same hugging `Text` probe, same 2000 × 400 window); T-3 is the AC-1.6/D-5 evidence. CON: "No fix needed" |
| UX outside-lens | CON | — | Manual steps 2 and 4 not run | Rejected (as a finding) | manual check is the user's step (ship-pr §8); carried into the report |

UX: no findings. Nits (logged): `CON-3 · PVSelectPlacementTests.swift · _ = window idiom unclear` (covered by F-2) · `MNT-4 · HostedWindow.swift:13 · named struct instead of a positional tuple`

### Fix plan 1
**F-1 — Add the Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-14/prs/S14-24-wide-width-class.md`.
Do: under the Status line, add a Landed note as other landed spike-14 plans do, saying there are no differences from the plan (the base check ran against `08eb109e`; `git diff b8ca4cfc 08eb109e -- macos` was empty).
Check: scripts/check-pr.sh PASS (plan-docs).

**F-2 — One way to hold a test window** (from MNT-1, CON-3)
Files: `macos/ProvenenciaTests/PVSelectPlacementTests.swift`, `macos/ProvenenciaTests/PVWidthClassTests.swift`, `macos/ProvenenciaTests/HostedWindow.swift`.
Do: replace every `_ = window` / `_ = (window, bareWindow)` with `withExtendedLifetime(...)` at the end of the test, after the last layout and assertion, as `WorkspaceDestinationHostTests.swift:81` does. In `HostedWindow.swift`'s doc, name `withExtendedLifetime` as the way to hold the returned window.
Check: `rg -n "_ = window|_ = \(window" macos/ProvenenciaTests` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Say who applies the modifier** (from MNT-3)
Files: `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift`.
Do: add one sentence to `pvMeasuresWidthClass()`'s doc: applied once, by `WorkspaceDestinationHost`; pages read `\.pvWidthClass` and do not re-measure (§ 9 of macos-client-patterns.md).
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `2ba4aea4`. Gate PASS (orchestrator run). F-1 done (`84500d2f`), F-2 done (`b14ca03a`; `rg "_ = window|_ = \(window" macos/ProvenenciaTests` empty), F-3 done (`2ba4aea4`).

## Round 2 — review
Not run: round 1 left no Fix-now blocker or major (CON-1 re-rated minor). Fix commits read by the orchestrator: CON-1 **Resolved**, MNT-1 **Resolved**, MNT-3 **Resolved**.

## CI — run 40946945949 (failure)
`test` job: Swift Testing aborted with `Received unexpected event testCaseDidStart for 'PVWidthClassTests/classFollowsPaneWidth(width:expected:)' (iteration 1). Test already started.` The arguments `(PVSpacing.widthWideBreakpoint, .wide)` and `(1600, .wide)` are equal values, so the two parameterized cases share one ID. The local gate passed, so the duplicate trips only on the CI runner's Xcode.

### Fix plan CI
**F-4 — No duplicate arguments in T-1** (CI)
Files: `macos/ProvenenciaTests/PVWidthClassTests.swift`.
Do: drop the `(1600, .wide)` argument from `classFollowsPaneWidth`, and add `#expect(PVSpacing.widthWideBreakpoint == 1600)` in a separate `@Test func breakpointIsSixteenHundred()`, so the plan's T-1 inputs are all still covered and no two arguments are equal.
Check: scripts/check-pr.sh PASS; CI `test` green.

## CI fix
Head `7e3bd31e`. Gate PASS (orchestrator run). F-4 done. CI run 40948499927: `test` and `localizable` green.

## Result
Gate PASS · CI green · open: none · follow-ups: none (nit MNT-4 logged) · to user: — · CON miss: none
