# Ship log — S16-06 Section switches and removals are spoken

**Plan:** [`S16-06-spoken-switches-and-removals.md`](S16-06-spoken-switches-and-removals.md) · **Branch:** `s16-06-spoken-switches-and-removals` · **PR:** [#785](https://github.com/mendahu/provenencia/pull/785)
**State:** fixing 1

## Before round 0

- Branch base: `main` @ `9a514612` (plan's Planned commit `0cd3700`; the implementer checks the Current state rows against the new base).

## Round 0 — implement
Head `b84c3d4c`. Gate PASS (orchestrator run against `9a514612`). Report deviations: version bump 0.9.0 → 0.9.1 (main had moved to 0.9.0; plan allows "the next patch on `main`"); `main` moved past `0cd3700` with no Current state row invalidated. Manual check §10 (VoiceOver, Full Keyboard Access) not run. The implementer's push of the last commit ("Record S16-06 as landed…") was denied by the permission classifier; remote is one commit behind. This log stub was committed in that commit.

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`, `Localizable.xcstrings`), SEC+PRF (`macos/App/Features/**` models: `UnpromoteFlow`, `ConclusionDetailPage`; dispatched on the "when unsure" rule — no security path matched).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-2 | UX | major | After ⌘3 from a section, Tab reaches the remaining segments before the Timeline section | To user → user (2026-11-11): one Tab stop, arrows → F-1 | verified by reading: every enabled segment is `.focusable()` (`PVSegmentedControl.swift:133`, doc comment lines 11-14), so the plan's AC-1.1 third clause and stop condition 4 contradict its own §6 States line and Out line "Tab order between segments… not planned" |
| UX-1 | UX | major | Space or a shortcut on a focused, unselected segment loses keyboard focus | Fix now → F-1 | verified by reading: `segment(_:)` puts the focus target in a different `if` branch (`focus.apply(AnyView(stop))` vs `stop`) and the target now follows the selection, so a switch rebuilds the focused segment; `switcherFocused` reads only the bound segment (`ConclusionDetailPage.swift:403`), so no claim restores it |
| CON-1 | CON | minor | T-5 passes even if the stay branch stops applying the effects | Fix now → F-2 | `confirmStaysAndSaysSo` checks nothing `session.apply` changes |
| MNT-1 | MNT | minor | Header band reaches navigation through a `select` closure and the environment | Fix now → F-3 | `select` has one caller that only calls `navigation.go`; `choose` reads `navigation` directly |
| MNT-3 | MNT | nit | `confirm` signature spelled two ways (code comment vs patterns doc) | Fix now → F-3 | stale doc text in files this PR changes |

Nits (logged): `MNT-2 · UnpromoteFlow.swift:25-40 · deletesOwnHandle and ownHandleDeletedRef search listed twice` · `MNT-4 · L10n.Conclusions.memberCount / members properties say "member" while copy says records (D-12 kept the key)` · `CON-2 · ConclusionPageSectionSwitchTests · no case with both focus flags true and different sections` · `CON-3 · ship log stale in the committed copy (this update resolves it)` · `UX-3 · L10n.swift:9192 · "provisional member" outcome still says member (outside P-17's scope; follow-up candidate)`

SEC+PRF: Clean, no findings.


**User decision on UX-2 (2026-11-11):** "One Tab stop, arrows". With a `focus:` handle, `PVSegmentedControl` is one Tab stop (the segment that carries the handle) and the arrow keys move the selection, as a native segmented control does. This supersedes the plan's Out line "A Tab order between segments … not planned" and the §6 States line "one per enabled segment"; AC-1.1 and stop condition 4 stand as written. The Conclusion switcher is the only caller that passes `focus:` (`DateValueEditorForm.swift:40`, `PropertiesView.swift:875`, `CitationComposerAuthoringBox.swift:601` pass none), so no other screen changes.

### Fix plan 1

**F-1 — A segmented control with a handle is one Tab stop; arrows move the selection** (from UX-2, UX-1)
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/DesignSystem/README.md` (row "SegmentedControl"), `macos/ProvenenciaTests/PVSegmentedControlTests.swift`, `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift`, the plan. Skill: `add-ui-component`, `add-swift-test`.
Failing test first: `PVSegmentedControlTests/arrowsMoveToTheNextEnabledSegment` — a new `static func arrowTarget(items:selection:forward:) -> Value?` returns the next (or previous) enabled item's value from the selected one, skipping disabled items, with no wrap (nil at either end, nil when nothing is enabled). Cases: forward from the middle, back from the middle, forward at the last, back at the first, skipping a disabled neighbour.
Do:
- In `segment(_:)`, only the segment `focusTarget(items:selection:)` names is `.focusable()` and carries `focus.apply`; every other enabled segment is the plain `chip(item)` (clickable, not a Tab stop). On the focusable segment keep `.onKeyPress(.space)` and add `.onKeyPress(.leftArrow)` / `.onKeyPress(.rightArrow)` that call `select(value)` with `arrowTarget(...)` and return `.handled` when it is non-nil, `.ignored` otherwise.
- The kit does not claim focus after a select: a host whose handle follows the selection re-claims it, as `ConclusionDetailHeaderBand.choose` already does when `switcherFocus != nil`. With one Tab stop, `switcherFocus != nil` is exactly "the keyboard is on the switcher", so Space, an arrow, or ⌘1–⌘5 while on the switcher keep focus on the new selected segment (UX-1). Say this in the doc comment.
- Rewrite the `PVSegmentedControl` doc comment and the `segment(_:)` comment, and the README row, to say: with a handle the control is one Tab stop on the selected segment (the first enabled one when it is disabled), Space chooses, ←/→ move the selection to the neighbouring enabled segment, and the handle's `next` / `previous` neighbours are still not read.
- `ConclusionPageSectionSwitchTests`: add a case named for the arrow path — switcherFocused true, vitals→family → "Family shown.", claims.
- Plan: amend the Out line "A Tab order between segments…" and §6 States ("one per enabled segment") with "superseded by the user on 2026-11-11 in round 1 of the ship log (UX-2): one Tab stop, arrows move the selection", and the manual check §10 step 2: replace "Tab to the Vitals segment and press Space" with "press → to move to Vitals". Do not change any AC or Decision.
Check: the new tests pass; `rg -n "every enabled segment is also a Tab stop|one per enabled segment" macos docs .agents` prints only the plan's struck-through history lines; scripts/check-pr.sh PASS.

**F-2 — T-5 proves the stay branch applies the effects** (from CON-1)
Files: `macos/ProvenenciaTests/UnpromoteFlowTests.swift`.
Failing test first: in `confirmStaysAndSaysSo`, after warming the detail key for `h-maria`, change `store.seededConclusionDetails["h-maria"]` to a different `memberCount`; after the confirm, wait until the warmed handle shows the new value. Prove it red by temporarily removing `session.apply(effects)` from the `.stay` branch, then restore it.
Do: only the test changes.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — The header band navigates itself; one spelling of `confirm`** (from MNT-1, MNT-3)
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `docs/macos-client-patterns.md`.
Do: remove `ConclusionDetailHeaderBand.select` and its argument in `ConclusionDetailLoaded.body`; `choose(_:)` calls `navigation.go(to: navigation.currentLocation.withPageSection(section))` itself. Spell the method `UnpromoteFlow.confirm(session:userID:leavingPage:announce:)` in both the `CatalogMutation.swift:119` comment and `macos-client-patterns.md:89`.
Check: `rg -n "select:" macos/App/Features/Conclusions/ConclusionDetailPage.swift` shows only the `PVSegmentedControl` argument; scripts/check-pr.sh PASS.

Also: the plan's Landed note gains one line naming the UX-2 decision and F-1.
