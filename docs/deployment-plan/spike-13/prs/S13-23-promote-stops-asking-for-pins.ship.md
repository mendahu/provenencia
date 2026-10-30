# Ship log — S13-23 Promote stops asking for pins

**Plan:** [`S13-23-promote-stops-asking-for-pins.md`](S13-23-promote-stops-asking-for-pins.md) · **Branch:** `s13-23-promote-stops-asking-for-pins` · **PR:** #633
**State:** ready

## Round 0 — implement
Head `9d4d567a`. Gate PASS (orchestrator run, against `b6185fe1`: 9 checks, xcode-test included). Report deviations (all recorded in the plan's Landed note): STEP-1+2 and STEP-3+4 each landed as one commit because neither pair built apart; four tests renamed so STEP-6's `(un)?fil…` sweep stays clean; T-2 drops its two assertions naming RPCs AC-3.1 forbids; T-6 renamed; registry count 19 → 18; process-filing comments in GenealogyStore/FakeStore and an add-localized-string example reworded; the S13-D11 Paste block refreshed to the current add-design-brief wording so plan-docs passes. Manual check (§10) not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Routed: UX for `macos/App/**` and `Localizable.xcstrings`; SEC for `api/ffi/swift_vocabulary_test.go`; PRF for `macos/App/Features/**` models (PromoteModel, PromoteFlow). Skipped: none.

Verdicts: CON Clean, MNT Clean, UX Clean, SEC+PRF Clean (no findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | minor | Re-keyed toast and header counts have no plural forms | Fix now → F-1 | verified: the four keys are single `stringUnit`s at head; this PR created them |
| MNT-1 | MNT | minor | `reasonText` keeps a nested switch and an unreachable `default` | Fix now → F-2 | verified: `Reason` has four cases (`PromoteFlow.swift:50-58`); the `default` hides a new move reason from the compiler |
| UX-4 (+ MNT outside-lens) | UX | minor | Column-heading row now 28 pt tall | Fix now → F-3 | verified: base `columnHeaders` tile was `frame(width: kindTile, height: 1)` (base `PromoteView.swift:216`); head frames every tile 28 × 28 (`PromoteView.swift:683`). Base height restored; the plan does not ask to change it |
| UX-3 | UX | minor | Dead `leaveTitle(name:)` reworded instead of deleted | Fix now → F-4 | verified: `git grep leaveTitle` finds only its definition at `L10n.swift:1509`; D-11 keeps only keys Swift names |
| MNT-2 (+ CON-1) | MNT | minor | Promote domain docs still say "file" for the Promote process | Fix now → F-5 | verified: `promote-graph-alignment.md` 7, 15, 31, 35, 52-54 (PR edits this file); `conclusion-layer-data-model.md` 243, 247, 254. §9.1 bridge filing stays |
| UX-6 | UX | nit → minor | Stale "filing page" comment at `L10n.swift:1836` | Fix now → F-6 | stale comment in a file this PR changes |
| MNT-8 | MNT | nit → minor | `binding-order-editing.md:7` still points at the deleted sheet | Fix now → F-6 | stale line in a file this PR changes |
| UX-2 | UX | minor | Duplicate-warning titles say "is also promoted" before Done | To user → Fix now → F-7 | the wording was D-20's §6 Strings table (plan lines 266-267). User, 2026-10-29: "Past tense probably makes more sense if you change it to \"selected\" rather than \"promoted.\" Something like \"Gracie has also been selected to be promoted,\" etc." |
| UX outside-lens | UX→CON | — | D-19 vs `PromoteView` reading `pvWidthClass` for `listMax` | Rejected | D-19 itself says PromoteView reads the width class; "only reader" is about the target width (CON round 1 checked the same) |
| UX outside-lens | UX→CON | — | Manual check §10 not run | Logged | the user runs §10; the PR body says so |
| SEC+PRF | — | — | No findings | — | — |

Nits (logged): `CON-1 · docs/promote-graph-alignment.md:7 · "quick filing action"` (folded into F-5) · `MNT-3 · L10n.swift:1468 · duplicateBody returns String, siblings return LocalizedStringResource` · `MNT-4 · PromoteFlow.swift:447, FakeStore.swift:84 · batchRows / lastPromoteBatch names` · `MNT-5 · GoStore.swift:1123 · mapSearchHeader shape` · `MNT-6 · PromoteView.swift · Color.clear.frame(height: 1) five times` (F-3 may touch it) · `MNT-7 · interpretation-layer-data-model.md:693 · missing "and"` (rides with F-6) · `MNT-9 · add-match-rule/SKILL.md:76, 91 · repeated sentence, stub step 6` · `UX-5 · promote.warning.conflictTitle · "A stronger match" names a strength the page no longer shows`

### Fix plan 1

**F-1 — Plural forms for the Promote counts** (from UX-1)
Files: `macos/App/Resources/Localizable.xcstrings`, `macos/App/Platform/L10n.swift` if a defaultValue changes. Skill: `add-localized-string`.
Failing test first: extend `L10nResolutionTests.unpromoteCopyResolves` (or a sibling test) so it resolves `promote.toast.promotedTitle` with 1 and 0, `promote.toast.promotedBody` with 1, `promote.page.rowSummary` with 1 and 1, and `promote.section.countPromoted` with 1 and 1, and expects the singular reading ("Promoted 1 subject and 0 connections", "1 subject is still on the graph, not promoted yet", "1 row · 1 already promoted"). Note `countPromoted`'s first number has no noun, so only its wording around the second may change; if no plural form differs, leave that key alone and say so.
Do: add `variations.plural` one/other entries the way `promote.done.connections` and `promote.leave.pageMessage` do; two-argument keys use catalog substitutions, one per argument.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — One exhaustive switch in `reasonText`** (from MNT-1)
Files: `macos/App/Features/Promote/PromoteModel.swift` (around 243-269).
Failing test first: none (compile-time property); existing `movesAreWordedNamingTheNeighbor` and `aViaReasonNamesTheTermInLowercase` must still pass.
Do: one `switch reason` with four arms (`via`, `neighborPlaced`, `neighborMoved`, `neighborReleased`), each move arm looking up its neighbor itself or through a small private helper; delete the `default`. If `L10n.Promote.reasonMoveUnknown` is then still used (missing-neighbor fallback), keep it; if not, delete it and its catalog key.
Check: `rg -n 'default:' macos/App/Features/Promote/PromoteModel.swift` shows no arm in `reasonText`; tests pass; scripts/check-pr.sh PASS.

**F-3 — Heading row keeps a 1 pt tile slot** (from UX-4, MNT outside-lens)
Files: `macos/App/Features/Promote/PromoteView.swift` (`PromoteGridRow`, `columnHeaders`).
Failing test first: none (layout); this restores base behavior (base `columnHeaders` used `Color.clear.frame(width: PromoteLayout.kindTile, height: 1)`).
Do: let `PromoteGridRow` frame the tile at `kindTile` width and only force the `kindTile` height on rows that draw a tile, e.g. a `tileHeight` parameter defaulting to `kindTile` that `columnHeaders` passes as 1, or a named empty-slot that frames itself. The working, anchor and Already filed rows stay 28 × 28. If it is natural, give the grid one named empty filler for the five `Color.clear.frame(height: 1)` sites (MNT-6); do not go further.
Check: `columnHeaders`' tile slot is 1 pt tall in code; scripts/check-pr.sh PASS.

**F-4 — Delete the dead `leaveTitle(name:)`** (from UX-3)
Files: `macos/App/Platform/L10n.swift:1509-1510`, `Localizable.xcstrings` key `promote.leave.title`.
Do: delete both. Leave `static let leaveTitle` (`promote.leave.pageTitle`) alone.
Check: `git grep -n 'promote.leave.title"' macos` finds nothing; the STEP-5 stale-key check still prints `[]`; scripts/check-pr.sh PASS.

**F-5 — Promote docs say promote for the Promote process** (from MNT-2, CON-1)
Files: `docs/promote-graph-alignment.md` lines 7, 15, 31, 35, 52-54; `docs/conclusion-layer-data-model.md` lines 243, 247, 254.
Do: where "file / files / filed / filing" names the Promote process, say promote ("Promote is a quick action", "Done promotes every row", "whether it promotes on Done", "can't promote both on one handle", "about what promotes"). Leave §9.1's bridge filing (`promote-graph-alignment.md` ~179-196) and any other use that names the automatic bridge write unchanged.
Check: `rg -n -i '\b(un)?fil(e|ed|es|ing)\b' docs/promote-graph-alignment.md` prints only §9.1 bridge-filing lines; scripts/check-pr.sh PASS.

**F-6 — Stale lines in files this PR changes** (from UX-6, MNT-8, MNT-7)
Files: `macos/App/Platform/L10n.swift:1836` comment "this Source's filing page" → "this Source's page"; `docs/ideas/binding-order-editing.md:7` drop or reword "so the next evidence open follows it" so it no longer points at the deleted sheet; `docs/interpretation-layer-data-model.md:693` add the missing "and" as STEP-7's text has it.
Check: scripts/check-pr.sh PASS.

**F-7 — Duplicate warnings say "selected to be promoted"** (from UX-2; user, 2026-10-29)
Files: `macos/App/Platform/L10n.swift` and `Localizable.xcstrings` keys `promote.warning.duplicateTitle`, `promote.warning.duplicateNewTitle`; the plan's §6 Strings rows 266-267 and its Landed note.
Failing test first: if a test resolves either key, update its expectation first; otherwise add one to `L10nResolutionTests`.
Do: `promote.warning.duplicateTitle` → "%1$@ has also been selected to be promoted to %2$@"; `promote.warning.duplicateNewTitle` → "%@ has also been selected to be promoted as New and looks the same". Update the plan's §6 table rows' new-copy column to match, and add one line to the Landed note: "UX-2: the duplicate warnings say 'has also been selected to be promoted' (user, 2026-10-29), superseding D-20's table rows."
Check: the test passes; STEP-6's `(un)?fil…` sweep still prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b3e966de`. F-1 to F-7 done; nothing not reproduced. Deviations: F-4/F-7 catalog edits ride in F-1's commit (`697859de`); `promote.section.countPromoted` left alone (no noun varies); F-5 also reworded `promote-graph-alignment.md` 124 and 142 and one more phrase on `conclusion-layer-data-model.md` 254, all naming the Promote process; the remaining `fil…` hits in `promote-graph-alignment.md` name the automatic bridge write and stay; MNT-6 not taken (optional).

## Round 2 — review
Not run: round 1 had no Fix-now blocker or major. Fix commits read by the orchestrator:
- UX-1 **Resolved** — `promotedTitle` and `rowSummary` use per-argument plural substitutions, `promotedBody` a plain plural; `promoteCountsUseSingularForms` red first, then green.
- MNT-1 **Resolved** — `reasonText` is one exhaustive four-arm switch, no `default`.
- UX-4 **Resolved** — `PromoteGridRow(tileHeight:)` defaults to `kindTile`; `columnHeaders` passes 1, as at base.
- UX-3 **Resolved** — `leaveTitle(name:)` and `promote.leave.title` deleted.
- MNT-2 / CON-1 **Resolved** — Promote-process wording says promote in both domain docs; §9.1 bridge filing kept.
- UX-6, MNT-8, MNT-7 **Resolved**.
- UX-2 **Resolved** — both duplicate warnings read "has also been selected to be promoted…" per the user; plan §6 rows and Landed note updated; `promoteDuplicateWarningsSayWhatIsSelected` red first, then green.

## Result
Gate PASS (`b3e966de`; `origin/main` unchanged at `b6185fe1`) · CI pending · open: none · follow-ups: none (nits logged) · to user: UX-2, answered · CON miss: none
