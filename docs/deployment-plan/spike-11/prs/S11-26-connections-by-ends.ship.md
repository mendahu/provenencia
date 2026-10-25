# Ship log — S11-26 Connections file by their ends

**Plan:** [`S11-26-connections-by-ends.md`](S11-26-connections-by-ends.md) · **Branch:** `s11-26-connections-by-ends` · **PR:** #527
**State:** ready

Base `main` @ `3666ac06` (plan planned against `1e8a580`).

## Round 0 — implement
Head `a6bc830f` (`74337886` STEP-1–3, `c7fb6f60` STEP-4, `e87d76dc` STEP-5 domain docs, `517287a8` + `a6bc830f` status docs). Draft PR #527. Gate PASS (orchestrator re-ran on `a6bc830f`).
Deviations reported:
- Plan Status is bare `In progress`: `check-plan-docs.py` rejects a PR link after the status word. The link lives in `completed.md`.
- Docs commit subject "Record S11-26 as landed in pull request 527." although Status is `In progress` (wording from `implement-pr`).
- Xcode tests run once after STEP-4, not after STEP-3 alone (STEPs 1–3 are one commit).
- `rg -i "switch"` on the domain docs matches the new §5.4 sentence the plan dictates.
- Manual check (§10) not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. All routes match: `macos/App/**` and `Localizable.xcstrings` (UX); `api/ffi/handlers/promote_batch.go` and `core/database/identityclaims` (SEC); `core/database/**` and `macos/App/Features/**` models (PRF). Scope `3666ac06...a6bc830f`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1, MNT-1 | PRF, MNT | minor | Panel header builds `connections` three times per render | Fix now → F-1 | verified: `PromoteView.swift:311-316` reads `connectionCount` twice and `connections` once; `connectionCount` is `connections.filter(\.files).count`. Same arguments and values as D-6, bound once; D-6's argument shape is kept |
| UX-1 | UX | minor | Leave guard reads "on 1 rows" (no plural rule) | Fix now → F-2 | verified: `promote.leave.pageMessage` is a plain `stringUnit` (`Localizable.xcstrings:18132`); `macos-l10n.mdc:23` requires `plural` variations for a count with a noun. D-8's value stays as the `other` form, so D-8 holds |
| UX-1 (rest) | UX | minor | `promote.connections.count` and `promote.toast.filedBody` have the same plural gap | Follow-up | keys this PR does not change; PR body follow-up list |
| MNT (outside lens → CON) | CON | — | Commit `517287a8` subject says "landed" while Status is `In progress` | Rejected | subject wording comes from `implement-pr`; the plan Status, `completed.md` and the map row all read In progress/In review, matching STEP-5. Pushed history is not rewritten for a subject line |

Nits (logged): `MNT-2 · macos/ProvenenciaTests/PromoteFlowTests.swift:477 · test name could say "connections"; plan keeps it, leave for S11-27` · `UX-2 · macos/App/Platform/L10n.swift:1659 · hint "files once both its ends are filed" is not true for a selfLink row; D-7 fixed the wording, S11-27 replaces the panel`

### Fix plan 1
**F-1 — Build the connection list once in the panel header** (from PRF-1, MNT-1)
Files: `macos/App/Features/Promote/PromoteView.swift`.
Failing test first: none (same values; a performance-only change in a view property).
Do: in `connectionMeta`, `let lines = model.flow.connections`, `let filed = lines.filter(\.files).count`, then `L10n.Promote.connectionsMeta(filed: filed, unfiled: lines.count - filed)`. Keep the arguments D-6 names.
Check: scripts/check-pr.sh PASS.

**F-2 — Give the leave guard's row count a plural rule** (from UX-1)
Files: `macos/App/Resources/Localizable.xcstrings`, test in `macos/ProvenenciaTests/PromoteFlowTests.swift` or beside the existing `L10n.Promote.recordCount` expectations in `PromoteEvidenceDisplayTests.swift`. Skill: `add-localized-string`.
Failing test first: `L10n.Promote.leaveDetail(rows: 1)` == "You chose a target on 1 row. Nothing has been filed; opening Promote again proposes the graph afresh." and `leaveDetail(rows: 3)` keeps D-8's text with 3.
Do: replace the `promote.leave.pageMessage` English `stringUnit` with `variations.plural`: `one` "You chose a target on %lld row. Nothing has been filed; opening Promote again proposes the graph afresh.", `other` D-8's value. Keep the key, comment, `extractionState` and `L10n.format` call; model the JSON on `promote.sheet.recordCount`. `L10nPluralTests.everyCatalogPluralPicksAFormPerCount` then covers it too.
Check: the new test passes; scripts/check-localizable-xcstrings.py OK; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5e6de547` (`95003e1f` F-1, `5e6de547` F-2). Gate PASS (orchestrator re-ran). F-1 done; F-2 done, its test failed red first ("on 1 rows") and passes now.

## Round 2 — review
Not run: round 1 had no Fix-now blocker or major. Minors read by the orchestrator in `a6bc830f...5e6de547`:
- PRF-1, MNT-1 → **Resolved**: `connectionMeta` binds `model.flow.connections` once; D-6's arguments unchanged.
- UX-1 → **Resolved**: `promote.leave.pageMessage` has `plural` `one`/`other`; `PromoteEvidenceDisplayTests.countsArePlural` covers 1 and 3; `L10nPluralTests` picks it up.

## Result
Gate PASS · CI pending · open: none · follow-ups: UX-1 (rest: `promote.connections.count`, `promote.toast.filedBody` plurals) · to user: — · CON miss: none
