# Ship log — S16-06 Section switches and removals are spoken

**Plan:** [`S16-06-spoken-switches-and-removals.md`](S16-06-spoken-switches-and-removals.md) · **Branch:** `s16-06-spoken-switches-and-removals` · **PR:** [#785](https://github.com/mendahu/provenencia/pull/785)
**State:** ready

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

## Round 1 — fix
Head `d100bfc7` (pushed; carries the earlier unpushed commit). Gate PASS (orchestrator run against `9a514612`). F-1 done (`arrowsMoveToTheNextEnabledSegment` red by compile failure, then green; `arrowOnTheSwitcherSpeaksAndClaims` added). F-2 done (red with `session.apply(effects)` removed: `memberCount` stayed 2; deviation: the test seeds a first detail for `h-maria` because the fixture had none). F-3 done (`CatalogMutation.swift:119` already had the full spelling; only `macos-client-patterns.md:89` changed). The F-1 commit also carries the `ConclusionDetailPage.swift` part of F-3.

Orchestrator check of the minors: CON-1 **Resolved** (red/green above). MNT-1 **Resolved** (`select` property and argument removed; `choose` calls `navigation.go`). MNT-3 **Resolved** (patterns doc now says `confirm(session:userID:leavingPage:announce:)`).

## Round 2 — review
Reviewers: UX (had UX-1, UX-2 majors). Not re-sent: CON, MNT (minors only, checked above), SEC+PRF (clean).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-2 (r2) | UX | blocker (re-rated from major) | With Full Keyboard Access on, the unselected segments are still Tab stops, so AC-1.1's "next Tab goes into the section" fails | Fix now → F-4 | verified by reading: the `else` branch of `segment(_:)` (`PVSegmentedControl.swift:155-157`) is a plain `chip(item)`, a SwiftUI `Button` (`PVChip.swift:96-112`) with no `.focusable(false)`; under Keyboard navigation every button is a key-loop stop. Re-rated to blocker: an AC fails in the setup the plan's own manual check uses (§10 turns Full Keyboard Access on) |
| UX-1 (r2) | UX | blocker (re-rated from major) | Under Full Keyboard Access, Space on an unselected segment's button drops focus | Fix now → F-4 | same branch: the button is not the bound view, so `switcherFocus` is nil and nothing claims; the segment then changes `if` branch and is rebuilt. Taking the non-target chips out of the key loop removes the path |
| UX-r2 note | UX | minor | Manual check step 2 says "press → to move to Vitals" from Gallery; order is vitals, family, timeline, gallery, sources and arrows do not wrap | Fix now → F-4 | `ConclusionPageSection.swift:38-43` |
| UX-r2 note | UX | minor | Plan Codebase fit "Neighbors" still says Space on an unselected segment keeps focus through `.focusable()` | Fix now → F-4 | plan line 182 no longer describes the code after F-1 |

### Fix plan 2

**F-4 — With a handle, the segments that do not carry it leave the key loop** (from UX-2 r2, UX-1 r2, the two round-2 notes)
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/DesignSystem/README.md` (row "SegmentedControl"), the plan. Skill: `add-ui-component`.
Failing test first: none possible (key-loop membership under Full Keyboard Access is not reachable from a unit test); manual check §10 step 1 with Full Keyboard Access on holds it.
Do:
- In `segment(_:)`, when `focus != nil`, the `else` branch is `chip(item).focusable(false)`; with no handle it stays `chip(item)` (no change for the other three callers). Mouse clicks still choose.
- Doc comment and README row: "one Tab stop, with or without Full Keyboard Access: the segments without the handle leave the key loop".
- Plan §10 step 2: replace "press → to move to Vitals: hear "Vitals shown."; focus stays on Vitals. Tab into the Vitals rows" with "press → to move to Sources: hear "Sources shown."; focus stays on Sources. Press ← three times to reach Family, hearing each section; Tab into the Family tiles", and adjust the following "click the Family segment" to "click the Vitals segment … hear "Vitals shown."". Plan Codebase fit "Neighbors": replace the "Space on a segment other than the selected one…" sentence with "Only the selected segment is in the key loop (with or without Full Keyboard Access), so Space, ←/→ or ⌘1–⌘5 on the switcher always start from the bound segment and `switcherFocus != nil` claims the new one (ship log round 1, UX-2 decision; round 2, F-4)."
Check: `rg -n "focusable\(false\)" macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` prints the one line; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `39a023ca`. Gate PASS (orchestrator run against `9a514612`). F-4 done: `else if focus != nil { chip(item).focusable(false) }` (`PVSegmentedControl.swift:158`); doc comment, README row, plan §10 step 2 and Codebase fit "Neighbors" rewritten. No failing test possible (key-loop membership under Full Keyboard Access); manual check §10 step 1 with Full Keyboard Access on holds it. Orchestrator read the commit: UX-1 (r2) and UX-2 (r2) **Resolved in code, pending manual check**; both round-2 notes **Resolved**.

## Result
Gate PASS (also after merging `main` @ `d161dc19`) · CI green · open: none in code; the §10 manual check (VoiceOver, Full Keyboard Access) is the only evidence for AC-1.1's Tab clause, UX-1 and UX-2 · follow-ups: none (UX-3 fixed as F-5 at the user's request) · to user: UX-2 answered (one Tab stop, arrows)

## After ready — user request
The user asked (2026-11-11) to fix UX-3 on this PR ("fix the member copy here"): it moves from Follow-up to Fix now.

**F-5 — The provisional outcome says "record"** (from UX-3)
Files: `macos/App/Platform/L10n.swift` (`L10n.Conclusions.outcomeProvisional`), `macos/App/Resources/Localizable.xcstrings` (`conclusions.outcome.provisional`), `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift:243`. Skill: `add-localized-string`.
Failing test first: change the expectation at `ReconciledValueDisplayTests.swift:243` to `"provisional record"`; it fails.
Do: the default value and the catalog `en` value become "provisional record"; the comment becomes "Auto-reconciler outcome: the record's Subject is only provisionally linked to this person, so it is reasoning only". Key unchanged. Leave `docs/conclusion-reconciliation.md` (it describes the data model, where "member" is the domain term). Add one line to the plan's Landed note naming the user's request and F-5.
Check: `rg -n '"provisional member"' macos` prints nothing; scripts/check-pr.sh PASS.

F-5 done: head `bb5758e7`. Red first (`otherReasonsHaveAPhraseAndAMark` failed on "provisional member" vs "provisional record"), then green. Gate PASS (orchestrator run). UX-3 **Resolved**; follow-ups: none.
