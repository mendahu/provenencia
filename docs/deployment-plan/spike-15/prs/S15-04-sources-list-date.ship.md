# Ship log — S15-04 The Sources list shows and sorts by the date

**Plan:** [`S15-04-sources-list-date.md`](S15-04-sources-list-date.md) · **Branch:** `s15-04-sources-list-date` · **PR:** [#774](https://github.com/mendahu/provenencia/pull/774)
**State:** ready

## Before round 0

- Branch cut locally from `main` at `918c8ad3`; `git fetch` failed (DNS cannot resolve github.com), so `main` is as of the last fetch.
- `ca89faa1` adds STEP-5a to the plan at the user's request (2026-11-10): the vertical hairline between the Source/Type cells and the Evidence graph column stops at the graph stack's bottom when a wrapped title makes the row taller. Scope row, Step and Manual check 8 added; no Decision or AC changed.

## Round 0 — implement
Head `688af0af`. Gate PASS (orchestrator run, against `918c8ad3`). Push and draft PR skipped: github.com still unreachable; "#N" placeholders in the step-8 docs commit. Report deviations: `columns` is a computed property over `@State listWidth` + `\.pvWidthClass` rather than a `let` in `body` (Landed note); a `pvListBand` sentence in `macos/App/DesignSystem/README.md`; D-14 string comments reworded; `688af0af` copies the current paste block into `S15-D1-source-page.md` because STEP-8 edits that brief and `check-plan-docs.py` failed on its stale block (kept: a docs consequence of a named Step, no product change); `pvListBand` landed in the STEP-6 commit. One earlier implementer gate run had 3 `SourcePageOverviewLayoutTests` nil-frame failures that passed on rerun with no change; the orchestrator gate passed clean. App not launched: board frames and §10 not compared.

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`, strings), SEC+PRF (`macos/App/Features/**` models, a sort over the catalog's Source list, list code). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Source page header decides dated vs undated on its own; the list and page disagree on an empty line | Fix now → F-1 | verified: `SourcePageIdentityHeader.swift:206` branches on `if let date` and draws `DateRowDisplay.line` as is; `SourcesRowDate.line` maps an empty line to nil (T-6 asserts a start-only range is nil). D-5 says one place decides; this PR edits that header's copy (D-14) |
| UX-1 | UX | minor | The two date sort labels truncate in the 220pt Sort menu when a non-date sort is current | Fix now → F-2 | verified `menuWidth: 220` at `SourcesListView.swift:266`; reviewer's measurement 165/171pt vs 160pt label room. The longest label needs ≥ 231 |
| CON-1 | CON | minor | Review file's brief link target still the pre-archive path | Fix now → F-3 | verified `S15-04-sources-list-date.review.md:3` targets `../design/S15-D2-sources-list.md`, which no longer exists |
| CON-2 | CON | minor | `688af0af` rewrites S15-D1's paste block rules 2 and 3 with no Step or Landed line | Fix now → F-3 | kept (orchestrator, round 0); the Landed note gets a line |
| CON-4 | CON | nit → minor | `SourcesListRow` type doc comment says "three columns" and "a second line of ref and last update" | Fix now → F-3 | stale comment in a file this PR changes |
| MNT-2 | MNT | minor | `.az`/`.za` compare titles inline beside the new `titleOrder` | To user | contradicts D-3 and its Out line ("`.az` and `.za` reading `titleOrder`": not planned); moving them changes the order of equal-title rows. Minor; the loop goes on and it is asked at the finish |

Nits (logged): `MNT-3/UX-2 · SourcesListRow.swift:310,314,372 · pass the line in hand to publishedHelp(date:) instead of recomputing with ?? ""` · `MNT-4 · SourcesModel.swift:344 · dateOrder beside titleOrder, not in Codebase fit` · `CON-3 · S15-D1-source-page.md:7 · missing period before "After the S15-D2 board"` (rides F-3) · `CON-5 · SourcesModelTests.swift:686 · T-6 could assert a dated spokenBase`

SEC+PRF: no findings (clean).

### Fix plan 1

**F-1 — The Source page header uses `SourcesRowDate.line` to decide dated vs undated** (from MNT-1)
Files: `macos/App/Features/Sources/SourcesRowDate.swift` *(new)*, `macos/App/Features/Sources/SourcesListRow.swift`, `macos/App/Features/Sources/SourcePageIdentityHeader.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift` or `SourcesModelTests.swift`. Skill: `add-swift-test`.
Failing test first: none practical at the view level (the header's branch is in a `@ViewBuilder`); T-6 already pins `SourcesRowDate.line(range, startYear only) == nil`.
Do: move `enum SourcesRowDate` unchanged out of `SourcesListRow.swift` into its own file `SourcesRowDate.swift` in `Features/Sources/` (the list row, the Source page and S15-05 all read it). In `SourcePageIdentityHeader.sourceDate`, replace `if let date = model.source?.date { let line = DateRowDisplay.line(date: date) …` with `if let line = SourcesRowDate.line(model.source?.date) { …` — the dated branch keeps everything it draws today, reading `line`; the `else` branch (Add published date) is unchanged. No other caller of `DateRowDisplay.line` changes.
Check: the app builds; existing `SourcePageModelTests` and T-6 pass; `rg "DateRowDisplay.line" macos/App/Features/Sources` prints only `SourcesRowDate.swift`.

**F-2 — The Sort menu fits its longest option** (from UX-1)
Files: `macos/App/Features/Sources/SourcesListView.swift`.
Do: `sortMenu`'s `menuWidth: 220` (line ~266) → `240`. The filter menus' 220 stay.
Check: the app builds.

**F-3 — Docs and comments** (from CON-1, CON-2, CON-4, CON-3)
Files: `docs/deployment-plan/spike-15/prs/S15-04-sources-list-date.review.md`, `docs/deployment-plan/spike-15/prs/S15-04-sources-list-date.md` (Landed note), `docs/deployment-plan/spike-15/design/S15-D1-source-page.md`, `macos/App/Features/Sources/SourcesListRow.swift`.
Do: the review file's Brief link target → `../design/archive/S15-D2-sources-list.md`. Landed note: one line each — the S15-D1 Paste block was brought up to date with the current `add-design-brief` text (rules 2 and 3); `SourcesRowDate` lives in its own file and the Source page header reads it (F-1); the Sort menu is 240 wide (F-2). S15-D1 line 7: add the missing period before "After the S15-D2 board". `SourcesListRow`'s type doc comment: describe the row as it is now — Source link (cover, title, the record line when narrow, the activity line of ref and last update), Type with the published date under it at Regular, a Published column at Wide, and the Evidence graph stack.
Check: `python3 scripts/check-plan-docs.py` on the plan passes; `test -e` on the review file's brief target; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `66b69414`. Gate PASS (orchestrator run). F-1 `97d57433`, F-2 `89963166`, F-3 `66b69414`. F-1's rg Check also prints `SourceDateDialog.swift:72` (the dialog's preview text, not a dated/undated decision): accepted. The implementer's gate needed three runs: layout tests in files this PR does not touch (`SourcePageOverviewLayoutTests` ×3, `PVPageColumnsTests.wideHostPlacesTheSecondaryBeside`, `PVWidthClassTests` ×2) failed with nil hosted-window frames, then passed with no change — the same flake as round 0. Not this PR's to fix; a follow-up.
UX-1 (F-2): **Resolved** — `menuWidth` 240 on `sortMenu` only, read by the orchestrator. CON-1, CON-2, CON-4 (F-3): **Resolved** — review link target exists, three Landed lines, row doc comment rewritten, read by the orchestrator.

## Round 2 — review
Reviewers: MNT (MNT-1 was a Fix-now major). Not sent back: CON, UX (minors only, resolved above); SEC+PRF (no findings). Scope `688af0af...66b69414`.
MNT-1: **Resolved** (`SourcePageIdentityHeader.swift:207` reads `SourcesRowDate.line`; the dialog's `DateRowDisplay.line` preview is not a dated/undated decision). New findings: none. Verdict clean.

## Result
Gate PASS at `66b69414` · CI not run (github.com unreachable: no push, no PR yet) · open: none · to user: MNT-2 · follow-ups: the nil-frame layout-test flake (`SourcePageOverviewLayoutTests`, `PVPageColumnsTests`, `PVWidthClassTests`)
Waiting on the network for finish steps 0–3: fetch and merge `origin/main`, gate, push, open the PR, fill every "#N", PR body, `gh pr ready`, CI.

## Finish
github.com reachable again; `origin/main` had not moved since `918c8ad3`, so no merge. Pushed and opened #774 ready for review (the user chose Create PR). "#N" placeholders filled with #774.

## To user — MNT-2 answered
User, 2026-11-10: "Yes, unify them." `.az` and `.za` move onto the one title comparison; this supersedes D-3's last sentence and its Out line ("`.az` and `.za` reading `titleOrder`": not planned). Equal titles, whose order `sorted` left unspecified until now, sort by id descending (newest added first) in both directions, the tie rule of `.added`, `.updated` and the date sorts.

### Fix plan 2
**F-4 — A–Z and Z–A share the date sorts' title comparison** (from MNT-2, user decision)
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`, the plan (Landed note, D-3's Out line row), `../completed.md` only if its S15-04 paragraph describes the A–Z sorts. Skill: `add-swift-test`.
Failing test first: `SourcesModelTests/titleSortsBreakTiesByNewestAdded` *(new)* — sources "b" (id 1), "a" (id 2), "A" (id 3), "a" (id 4) where ids are added order. `.az` → ids [4, 3, 2, 1] ("a"/"A"/"a" compare same case-insensitively, so newest first, then "b"); `.za` → [1, 4, 3, 2]. Feed the rows so that the current unstable `sorted` would not happen to give this (e.g. insert ascending id); if it passes before the fix, report it as Not reproduced and stop.
Do: `titleOrder` takes `descending: Bool = false` — the title comparison flips when descending, the `orderedSame` id-descending tie does not. `.az` sorts by `titleOrder($0, $1)`, `.za` by `titleOrder($0, $1, descending: true)`; the date sorts call it unchanged. Update `titleOrder`'s doc comment to say all four title-ordered sorts use it. In the plan's Out table, mark the "`.az` and `.za` reading `titleOrder`" row as reversed by the user (2026-11-10, F-4), and add one Landed line.
Check: the new test passes; `scripts/check-pr.sh` PASS.
