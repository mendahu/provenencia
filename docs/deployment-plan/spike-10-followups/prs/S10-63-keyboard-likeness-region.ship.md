# Ship log — S10-63 Keyboard region for a likeness

**Plan:** [`S10-63-keyboard-likeness-region.md`](S10-63-keyboard-likeness-region.md) · **Branch:** `s10-63-keyboard-likeness-region` · **PR:** #467
**State:** fixing 1

Worktree: `/Users/jakerobins/Repos/provenencia-s10-63`, from `origin/main` @ `90c8d083`.

## Round 0 — implement
Head `883220c5`. Gate PASS (orchestrator run: 7/7 checks; 1166 Swift tests, the 10 new ones among them). Report: main had moved from `aa60b63` to `90c8d083`, with no Current-state row broken (S10-61's 256-point polygon cap does not affect the 4-point whole-file rectangle). Deviations: the three L10n keys landed in the STEP-2 commit, because the model returns them; the manual check (§10) is left for the user. AC-1.9 and AC-3.1–3.4 (focus/VoiceOver) are covered only by the manual check.

## Round 1 — review
Lenses: CON clean (1 minor, 1 nit) · MNT changes needed (1 major, 1 minor, 2 nits) · UX clean (2 minors, 1 outside-lens note) · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Focus double hop now hand-written three times | To user → Fix now → F-1 (user, 2026-10-18: extract, overriding D-11's copy) | Verified: `CitationComposerViewerPane.swift:75-83` copies `CitationComposerAuthoringBox.swift:293-298` and `:478-483`. D-11 and the Precedent line bind the copy, so extracting it contradicts a Decision |
| CON-1 / MNT-4 | CON, MNT | minor | `editingSavedLikeness(region: String? = nil)` differs from STEP-5 and is missing from the Landed note | Fix now → F-2 | Verified at `CitationComposerDepictionTests.swift:169-180`. A default argument cannot call the instance helper. Fix is a Landed bullet; rides along if a fix round runs |
| UX-3 (UX outside-lens) | MNT | minor | `CitationDepictionValueArea` doc comment says "the button that draws", but the area now has three buttons | Fix now → F-3 | Verified at `CitationComposerAuthoringBox.swift:914-915`. Rides along if a fix round touches the file |
| MNT-2 | MNT | minor | `priority:` parameter has no caller that sets it | Rejected | D-16 binds the signature with `priority: … = .medium`, and UX-1's fallback would use it |
| UX-1 | UX | minor | Drawing-mode announcement may be cut off by the focus move | Follow-up (manual check step 7) | Low confidence. Code matches D-11 as written, and the plan's Stop line covers step 7 failing. Confirm during the manual check |
| UX-2 | UX | minor | Needs-page caption and region-hidden Callout still name only Draw | Follow-up | Copy is Out (UX-12/UX-13). Add both strings to that audit item |
| MNT-3 | MNT | nit | `@discardableResult` on `useDepictionRegion` with no discarding caller | Logged | |
| CON-2 | CON | nit | AC-2.6 proven through `isDepictionRedrawn` only | Logged | `editWithoutRedrawKeepsTheImage` covers the flag-to-save step |

### Fix plan 1

**F-1 — One helper for the focus double hop** (from MNT-1; the user overrode D-11's copy on 2026-10-18)
Files: `macos/App/Platform/FocusClaim.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`. Skill: `add-ui-component` (Platform layer).
Failing test first: none. This is view focus timing with no model seam, so the behavior is unchanged and covered by manual check steps 6 and 7.
Do: add one `@MainActor` helper that takes a `FocusState<V>.Binding` (V: Hashable), a value and an optional `then: () -> Void`. On the next main-queue turn it writes the value and runs `then`; on the turn after, it writes the value again if the binding still differs. It carries the single comment that explains why the write happens twice (the target may not be in the hierarchy yet, or a dismissing sheet may swallow the first write), and the per-site comments go. Register the file in `project.pbxproj` with four objects, as `AccessibilityAnnouncement.swift` was, using unused `AA…` IDs. Move both `claimTab` bodies (`CitationComposerAuthoringBox.swift` ~293 and ~478) and `claimDrawingFocus` (`CitationComposerViewerPane.swift` ~75) onto it. The viewer pane passes `$drawingCancelFocused`, `true`, and the `AccessibilityAnnouncement.post(…depictionDrawingHint)` as `then`, so the order D-11 sets (focus write, then announce, in the first turn) is unchanged. Record in the plan's Landed note that D-11's copy became a shared helper by the user's call.
Check: `rg -n "DispatchQueue.main.async" macos/App/Features/CitationComposer` no longer shows the nested double-write pattern; scripts/check-pr.sh PASS.

**F-2 — Landed note: `editingSavedLikeness` signature** (from CON-1, MNT-4)
Files: `docs/deployment-plan/spike-10-followups/prs/S10-63-keyboard-likeness-region.md` (Landed section).
Do: add a bullet saying `editingSavedLikeness` takes `region: String? = nil` and falls back to today's region in the body, because a default argument cannot call the instance helper `regionJSON`. No code change.
Check: `scripts/check-plan-docs.py` 0 errors.

**F-3 — Value area doc comment names all three buttons** (from UX-3)
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` (~914).
Do: reword the `CitationDepictionValueArea` doc comment so it lists the preview, the caption, and the Draw, whole-file and citation's-region buttons.
Check: scripts/check-pr.sh PASS.
