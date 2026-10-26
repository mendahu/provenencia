# Ship log — S11-50 Matching warms while the project opens

**Plan:** [`S11-50-warm-matching-at-open.md`](S11-50-warm-matching-at-open.md) · **Branch:** `s11-50-warm-matching-at-open` · **PR:** #570
**State:** ready

Base `main` @ `d050a503` (S11-49 landed after the plan's base `13dab35f`; it changed `graphcache/graph.go`, `graphcache/commit.go`, `promotealign/promotealign.go`, `promotealign/export_test.go`).

## Round 0 — implement
Head `9a680f5f`. Gate PASS (orchestrator run, against `d050a503`). Report deviations: tests written with the code, not seen red first; plan Status reads `Landed in [#570](…) (2026-10-26)` (plan-docs rejects bare `Landed #570`); no Landed note (no deviation); PR body has local budget figures (warm at open 425 ms, cold 373 ms, one-place cold 952 ms), CI figures still owed.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (`core/catalogsession`, `core/database/**`). Skipped: UX (no `macos/App/**`, strings or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | One value that does not decode on any person, event or place now fails the project open; a re-warm failure in `Do` replaces a committed write's nil result and repeats on every later call | To user → Fix now → F-1 | verified: `session.go` `openResearcher` closes and returns any `WarmOnOpen` error; `WarmOnOpen` returns `promotealign.Warm`'s error; `index.go` `loadShortlist` returns on the first `keyValues` error (decode of date and name blobs); `Do` sets `err = WarmOnOpen(...)` after `fn` returned nil. D-4 says "return its error" |
| PRF-1 | PRF | major | Every vocabulary or structure edit rebuilds the whole shortlist index before it returns, holding the session lock | To user → Fix now → F-1 | verified: `Do` re-runs all of `WarmOnOpen` when `WarmNeeded`; a vocabulary commit sets it (`dropDerived`), a structure commit drops every node. Researcher paths: custom term while citing, metadata fields, source types, properties. Contradicts D-4 as written |
| CON-1 | CON | minor | The plan has no Landed note | Fix now (with the fix plan) | DoD line |
| MNT-1 | MNT | minor | "Kinds loaded with no SELECT" check copied in two packages; `TestDoWarmsNameFormat` does not use `newProject` | Fix now | same-pass |
| MNT-2 | MNT | minor | Two descriptions of open-time warms in `promote-graph-alignment.md`; stale `openResearcher` comment | Fix now | files this PR changes |

Nits (logged): `MNT-3 · session.go:71 · "name index" → "shortlist index"` · `MNT-4 · export_test.go:104 · MatchIndexGenForTest takes a graph, ProfilesKeptForTest a catalog` · `MNT-5 · session.go:75 · WarmOnOpen also runs after a drop`

**User decision (2026-10-26) on SEC-1 and PRF-1:** "Change D-4: a matching warm failure never fails the open or a committed call (leave the index unbuilt; the next proposal reports the error); after a mid-session drop, re-warm only the alias index and name format, and leave the shortlist index to the next proposal."

### Fix plan 1

**F-1 — The matching warm runs at open only, and never fails a call** (SEC-1, PRF-1; D-4 revised by the user)
Files: `core/catalogsession/session.go`, `core/catalogsession/warm_test.go`, `docs/deployment-plan/spike-11/prs/S11-50-warm-matching-at-open.md` (Landed note), `.agents/skills/use-catalog-session/SKILL.md`, `docs/promote-graph-alignment.md`.
Failing test first: `TestOpenSurvivesABadValue` *(new)* in `warm_test.go`: a project from `onboarding.Complete`; inside `Do`, insert a canonical person row and an `auto_reconciler_values` row whose `value_date` blob is invalid bytes (for the person's `birth_date` or another date property), then `catalogsession.Close`; the next `Do` opens without error and `g.MatchIndex()` is absent; `promotealign.Propose` on a graph returns an error (the first proposal reports it). If opening rebuilds the reconciler cache and erases the row, stop and report.
Do:
- One unexported `warmGraph(c *database.Catalog) error` in `session.go`: `namealiases.Warm`, `nameformats.Warm`, then `c.Graph().MarkWarm()`.
- `WarmOnOpen(c)`: `warmGraph(c)`, then `promotealign.Warm(c)` with its error ignored (the shortlist index stays unbuilt; the next proposal builds it and reports the error). Its error is `warmGraph`'s only. Doc comment says so, and that `Do` re-runs only `warmGraph` after a drop.
- `Do`: after `fn` returns nil and `WarmNeeded()`, run `warmGraph`; its error never replaces `fn`'s result (the flag stays set, so the next call tries again).
- `TestDoRewarms`: AC-2.1, AC-2.2 and AC-2.3 now assert, after the drop, that the alias index and name format are present, `WarmNeeded()` is false and the match index is absent; AC-2.2 also that the kinds are not loaded. AC-2.4 and AC-2.5 unchanged.
- Plan Landed note: D-4 revised by the user on 2026-10-26 (quote), and what `WarmOnOpen` and `Do` do now.
- Docs: the use-catalog-session skill and `promote-graph-alignment.md` say the shortlist index is built at open only; after a drop it waits for the next proposal.
Check: the new test passes; `TestDoWarmsMatching`, `TestDoRewarms`, `TestLargeSourceBudgets` pass; scripts/check-pr.sh PASS.

**F-2 — One "kinds loaded" test helper** (MNT-1)
Files: `core/database/graphcache/graphcachetest/graphcachetest.go`, `core/catalogsession/warm_test.go`, `core/database/promotealign/memo_test.go`.
Do: `graphcachetest.AssertKindsLoaded(t testing.TB, c *database.Catalog, keys ...string)` (look up each provenencia subject type id the way `Canon.Entity` does — share that lookup — then require `g.Kind` to issue 0 SELECTs); both tests call it; `TestDoWarmsNameFormat` uses `newProject`.
Check: `git grep -n "FROM subject_types WHERE key = ?" -- core` prints only `graphcachetest.go`; tests pass.

**F-3 — One description of open-time warms** (MNT-2)
Files: `docs/promote-graph-alignment.md`, `core/catalogsession/session.go`.
Do: merge the "Open-time warms" and "Matching warms at open" bullets into one that names all three warms, says the shortlist index is built at open only, and what `Do` re-warms after a drop; update the `openResearcher` comment to match. Use "shortlist index" for the index (MNT-3).
Check: scripts/check-pr.sh PASS.

**F-4 — Landed note** (CON-1)
Files: the plan.
Do: the Landed note also lists: tests written with the code, the Status form, and that the CI budget figures go in the PR body when CI runs.
Check: `python3 scripts/check-plan-docs.py` 0 errors.

## Round 1 — fix
Heads `c48ef368`, `26ad6843`, `b9917db2`. Gate PASS (orchestrator run on `b9917db2`). F-1: `TestOpenSurvivesABadValue` failed first ("proto: cannot parse invalid wire-format data"), passes after; `TestDoRewarms` AC-2.1 – AC-2.3 reworked to the revised D-4. The planned `Propose` assertion was dropped (an unknown graph id errors by itself); `promotealign.Warm` returning the decode error after the open is the evidence. F-2: `graphcachetest.SubjectTypeID` and `AssertKindsLoaded`; every test lookup of a subject type id uses it (`git grep "SELECT id FROM subject_types" -- '*_test.go'` empty), including two in `graph_test.go`. F-3, F-4 done; `completed.md` updated.
CON-1, MNT-1, MNT-2: **Resolved** (read `ca8c7836`, `62fff58f`, `c48ef368`, `26ad6843`, `b9917db2`).

## Round 2 — review
Reviewers: SEC+PRF (SEC-1, PRF-1). Scope `9a680f5f...b9917db2`.
SEC+PRF: SEC-1 and PRF-1 **Resolved** to the revised D-4. No new blocker. `origin/main` had not moved; gate PASS on `b9917db2`.

## Result
Gate PASS · CI green (Go tests 40387975407 incl. TestLargeSourceBudgets, macOS tests 40387975380) · open: none · follow-ups: none · to user: SEC-1, PRF-1 (answered: D-4 revised) · CON miss: none
