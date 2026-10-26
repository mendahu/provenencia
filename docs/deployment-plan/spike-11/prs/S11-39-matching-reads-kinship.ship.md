# Ship log — S11-39 Matching reads kinship shapes

**Plan:** [`S11-39-matching-reads-kinship.md`](S11-39-matching-reads-kinship.md) · **Branch:** `s11-39-matching-reads-kinship` · **PR:** #567
**State:** fixing 1

## Round 0 — implement
Head `f631a272`. Gate PASS (orchestrator run). Report deviations: STEP-7 baseline taken from a `git archive` export of `7efa02f` rather than a worktree (export deleted); extra `cloneShapes` helper in `promotealign/terms.go`; T-6 goldens use `kinTree`/`kinSide`/`kinShape` helpers; manual check not run in the app (Go goldens cover its steps). The landing commit also picked up this ship log stub.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (`core/database/promotealign`, graph code, loads stored links and trusts their count). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC/PRF | major | `kinship.Derive` pairs every participant of a pattern event (and parents × children through a middle person), unbounded; now reached from both loaders on every proposal | To user → Fix now → F-1 | verified: `eventLinks` nested `for x in group { for y in group }` (`kinship.go:189-191`); reviewer's scratch run, 2,000 spouses on one marriage → 1,999,000 links in 4.5 s. Defect in callers this PR adds, so Fix now in shape, but the fix puts a bound in a package documented "no numbers" under the working rule "No new knob" |
| PRF-1 | PRF | major | Canon re-derives implied links on every Promote decision, no memo | Rejected (memo) · Fix now → F-3 (budget case) | D-8 and the Out line decline the memo on `BenchmarkDerive` (16 ms at 8,190 links); the reviewer confirms `TestLargeSourceBudgets` has no chains or marriages in the canon, so D-8's "budgets moved within noise" was not measured. A budget case that exercises the path tests D-8's premise; the Out line's "revisit if a budget fails" then applies |
| MNT-1 | MNT | minor | `sideLink` repeats `graphalign.Bridge` field by field | Rejected | D-5 names `sideLink` embedding `kinship.Edge` so `withImplied` reads `l.Edge`; holding `Bridge` moves the conversion into `withImplied` instead of removing it |
| MNT-2 | MNT | minor | Loader doc comments omit the merge and the implied links | Fix now → F-2 | read `layer.go:155-157`, `canon.go:32-36`: unchanged by the diff |
| SEC-2 | SEC | minor | `linkSet.add` rescans associations on each merge, O(M²) for M marriages of one couple | Fix now → F-4 | `terms.go` is added by this PR; `slices.ContainsFunc` over `have.Associations` per added association |

Nits (logged): `CON-1 · plan:17 · Review line says 2 rounds` (misread: those are the plan review's rounds) · `CON-2 · plan:145 · Codebase fit names TestWithImplied, which does not exist` · `MNT-3 · core/graphalign/align.go:326 · pass st.placedOf instead of a closure` · `MNT-4 · promotealign/terms.go · linkSet would read better in links.go`

**SEC-1, user's answer (2026-10-26):** "the cheap structural guard is fine, but I want it to be more than 2, just to account for strange familial situations. If we cap it at something like 10". Bound: 10 per end.

### Fix plan 1

**F-1 — Bound what one pivot can imply in `core/kinship`** (from SEC-1; user-approved)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`. Skill: `add-match-rule` (read it; this is a structural guard, not a registry knob, so no `matchregistry` entry).
Failing test first: add `TestDerive` rows: (a) a marriage with 11 spouse participations implies nothing for that marriage; (b) a marriage with 10 spouses implies 45 spouse links; (c) a birth with 11 parent participations (father/mother roles) implies no parent links for that birth; (d) a middle person with 11 parents implies no grandparent link through them, while one with 10 does. Also keep every existing row passing.
Do: add one unexported constant beside the patterns, `maxPivotEnd = 10`, with a comment: a pivot (an event, or the parent between in a chain) with more than 10 participants at either end reads as malformed data and implies nothing; it bounds the links one pivot makes so `Derive` stays linear in its input. In `eventLinks`, per event group, count participants with `atA` and with `atB`; if either count exceeds `maxPivotEnd`, skip the group. In `chains`, skip a middle person whose distinct parents exceed `maxPivotEnd` (count distinct `g` over steps whose child is that person). Update the package doc comment, which says "no numbers": say it holds no weights or scores, only this bound on malformed data.
Check: the new rows pass; `go test ./core/kinship/ -bench Derive` still runs; scripts/check-pr.sh PASS.

**F-2 — Say what the loaders return** (from MNT-2)
Files: `core/database/promotealign/layer.go` (`loadLayerBridges` comment), `core/database/promotealign/canon.go` (`loadCanon` comment).
Do: add one sentence to each: links go through one `linkSet`, so links with the same ends and keys merge, and the links the kinship patterns imply follow the recorded ones (with no association of their own beyond those they were made from).
Check: scripts/check-pr.sh PASS.

**F-3 — Make the large-source budgets exercise implied links** (from PRF-1; the memo stays out per D-8)
Files: `core/database/promotealign` large-source budget test (`TestLargeSourceBudgets` and its fixture).
Do: add one budget case (or extend the fixture) whose canon holds a multi-generation parent tree with marriages, at a size comparable to `BenchmarkDerive/generations=12` (about 8,000 parent links), reached from the proposal's seeds, so a warm proposal runs `withImplied` over it. Reuse the existing budget limits and fixture builders; do not add a new budget number unless no existing one fits, and if so name it in the report. If the case fails its budget, stop and report the numbers (the plan's Out line says to revisit the memo then); do not add a memo.
Check: the case passes; report its time against the budget; scripts/check-pr.sh PASS.

**F-4 — Merge associations without rescanning** (from SEC-2)
Files: `core/database/promotealign/terms.go`.
Failing test first: none needed (behavior unchanged); `TestLinkSet` must stay green, including association order.
Do: in `linkSet`, keep per link a set of its associations (for example `seen []map[string]struct{}` parallel to `links`, or a `map[seenEdge]map[string]struct{}`), filled on a miss and checked on a hit instead of `slices.ContainsFunc`. Keep association order as first added.
Check: `TestLinkSet`, `TestLayerBridgesMergeKinshipShapes` pass; scripts/check-pr.sh PASS.

Also record in the plan's Landed note: F-1 adds the bound of 10 per pivot end (user decision on SEC-1, in this ship log); F-3 adds the budget case.
