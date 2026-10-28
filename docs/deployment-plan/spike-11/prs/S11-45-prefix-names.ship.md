# Ship log — S11-45 Prefix names

**Plan:** [`S11-45-prefix-names.md`](S11-45-prefix-names.md) · **Branch:** `s11-45-prefix-names` · **PR:** — (not opened: push denied, see Round 0)
**State:** ready — awaiting user approval to push and open the PR

## Round 0 — implement
Head `40545614`. Base `7adb3bcb`. Report: implementer's gate PASS. Stopped before push: `git push` and `gh pr create --draft` were denied by the session's permission classifier, so STEP-7's push/draft PR and STEP-8's "Landed #N" docs commit are not done. Deviations: `CausePrefix` comment is trailing (gofmt alignment), wording per D-2; base moved since `784c7ae` but only `L10n.swift` changed among Current state files (unrelated), no row false. Manual check not run.
Orchestrator gate at `40545614`: PASS (diff-check, gofmt, go-vet, go-test, localizable, guidance-names, plan-docs, xcode-test, swift-warnings).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`. SEC: matching layers (`core/valuecompare`, `core/matchregistry`, `core/autoreconcile`). PRF: `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | `name-candidates.md` Roles bullet names `AliasRoles`, Words bullet gives the pre-prefix order | Fix now → F-1 | verified: lines 46–47 read "today: exact, initial, alias, then fold and spelling" and "(`valuecompare.AliasRoles`)"; file is in the diff, D-4 makes `PrefixRuleRole` the one role test |
| CON-1 / MNT-2 | CON, MNT | nit→minor | `core/autoreconcile/names.go:22` Format-agnostic line is 93 columns; D-12 says rewrap | Fix now → F-2 | verified: the line runs past the comment's ~78-column wrap; a plan instruction (D-12) not met, stale-layout in a changed file |

Nits (logged): `UX-1 · macos/App/Platform/L10n.swift:1588 · "Matches by a shortened name" has an article, "Matches by initial" does not` · `MNT-3 · core/valuecompare/name_test.go:36 · prefixCredit sits under the "Edit-distance ratios" header` · `MNT-4 · docs/ideas/daitch-mokotoff.md:13 · say which rung runs first, phonetic or prefix`
Verdicts: CON clean, UX clean, SEC+PRF clean (no findings), MNT clean with one minor.

### Fix plan 1
**F-1 — Feature doc names `PrefixRuleRole` and the built word-table order** (from MNT-1)
Files: `docs/deployment-plan/spike-11/name-candidates.md`.
Failing test first: none (docs).
Do: in § Prefix rule, in detail — Words bullet: replace "today: exact, initial, alias, then fold and spelling" with "exact, initial, alias, fold, spelling, then prefix". Roles bullet: replace "The alias tables' role set (`valuecompare.AliasRoles`): given, nickname and untyped parts." with "The prefix roles (`valuecompare.PrefixRuleRole`, the alias tables' roles): given, nickname and untyped parts."
Check: `rg -n "AliasRoles|today: exact" docs/deployment-plan/spike-11/name-candidates.md` prints nothing; scripts/check-pr.sh PASS.

**F-2 — Rewrap the Format-agnostic bullet** (from CON-1, MNT-2)
Files: `core/autoreconcile/names.go`.
Failing test first: none (comment).
Do: rewrap the "Format-agnostic" bullet's last sentence (line 22) to the comment's width (~78 columns), as the Fold bullet below it is wrapped. Wording unchanged.
Check: `awk 'length > 80' core/autoreconcile/names.go` prints no comment line from the file header; gofmt clean; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0dd38630`. Gate PASS (orchestrator run). F-1 done (`31744119`): **Resolved** — Roles bullet names `PrefixRuleRole`, Words bullet gives the built order. F-2 done (`0dd38630`): **Resolved** — bullet rewrapped, wording unchanged.

## Round 2 — review
Not run: round 1 had no Fix-now blocker or major.

## Result
Gate PASS · CI — (not pushed) · open: none · follow-ups: none · to user: push / draft PR (permission denied to implementer) · CON miss: none
