# Ship log — S11-39 Matching reads kinship shapes

**Plan:** [`S11-39-matching-reads-kinship.md`](S11-39-matching-reads-kinship.md) · **Branch:** `s11-39-matching-reads-kinship-v2` · **PR:** #572
**State:** CI fix round (user decision)

The first ship loop (#567, closed in the slice 9 re-cut) has its own ship log on `s11-39-matching-reads-kinship`. This log is the re-plan's, from `main` @ `7fdd4776`.

## Round 0 — implement
Head `6be2f361`. Gate PASS (orchestrator run, against `7fdd4776`). Report deviations: the whole `TestProposeKinshipShapes` ported in STEP-4 (AC-1.2, 1.4, 1.6 and 2.3 failed before D-7); `TestRecordingMerge` added; `walkedPeople` holds person ids only; the `linkSet.add` profile covers the whole budget test (0.60 s of 24.7 s CPU, 2.4%); manual check not run.
Local budgets (orchestrator, two runs, load average 9–11): warm at open 0.67–1.03 s (2.5 s), cold 0.52–0.71 s (2.5 s), **warm 160–190 ms (250 ms)**, after a commit 0.52–0.62 s (1.5 s), Evidence 8 ms (100 ms), one-place cold 1.16–1.31 s (5 s), one-place warm 0.88–1.00 s (3 s). Risk: CI ran about 2.3× slower than this Mac on warm proposals during #567, so warm may fail CI.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (`core/database/**`, graph and cache code; asked to profile the warm proposal over the tree). Skipped: UX (no `macos/App/**`, strings or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | The warm proposal costs about 40 ms more (base 90–96 ms, head 130–135 ms on the same tree in series), likely over CI's 250 ms budget | Fix now → F-1; CI decides F-1b | reviewer's profile: `Reader.For` 13 ms, `linkSet.add` 12.3 ms vs the inline merge's 7.5 ms, `addImplied` 9 ms, GC +18 ms; a lazy association map measured −5 ms. Verified in code: `canon.go` adds each implied link from both ends; `links.go` makes an association map per link |
| MNT-1 | MNT | major | Key → id stays a private copy (`typeByKey`) beside the new `TypeKeysByID` | Fix now → F-2 | same-pass; supersedes D-8's "`typeByKey` stays" |
| MNT-2 | MNT | major | `lookupTree` repeats `bulkCanon`'s vocabulary lookups | Fix now → F-3 | same-pass, file this PR changes |
| MNT-3 | MNT | minor | Three copies build a two-subject marriage in tests | Fix now → F-4 | same-pass |
| MNT-4 | MNT | minor | `linkSet.add`'s miss path copies `Recording` its own way | Fix now → F-5 | file this PR adds |
| MNT-5 / CON-1 | MNT, CON | minor | The S11-39 map row still says "the existing merge" | Fix now → F-6 | dedupe of CON-1 (nit) |
| SEC-1 | SEC | minor | Shapes on one link are bounded by the distinct pivots joining a pair, not by `MaxPivotEnd`; `Merge` and `CountedOnce` scan them | Follow-up | needs thousands of marriages between one couple to cost a fraction of a second (reviewer: not a hang); a linear merge needs per-link state the value type cannot carry cleanly. PR body follow-up |

Nits (logged): `MNT-6 · graphalign/align_test.go:2314 · marriageFixture comment says spouse` · `MNT-7 · links.go:359 · canon associations unused` · `MNT-8 · promotealign.go:207 · checkFixed builds the whole map for one id`

### Fix plan 1

**F-1 — Less per-link work on the warm walk** (PRF-1)
Files: `core/database/promotealign/links.go`, `core/database/promotealign/canon.go`.
Do: (a) `linkSet` creates a link's association set only when the link has associations (canon links have none); (b) `loadCanon` pre-sizes the set from the walk's link count; (c) `loadCanon` adds an implied link only from the `For` call of its `A` end (`bytes.Equal(imp.A, id)`), since both ends are walked and the other end returns the same link. No behavior change.
Check: `TestProposeKinshipShapes`, `TestLinkSet` pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run 'TestLargeSourceBudgets$' -v` passes; report warm (three runs) in the report.

**F-1b — (only if CI's warm budget is red after F-1)** a per-person `For` result in a `graphcache` `NodeMemo` slot. Planned when CI reports; not in this round.

**F-2 — One type-id lookup in both directions** (MNT-1)
Files: `core/database/graphcache/vocab.go`, `core/database/graphcache/vocab_test.go`, `core/database/promotealign/profiles.go`, `core/database/promotealign/canon.go`.
Do: `Vocabulary.TypeIDsByKey() map[string][]byte` beside `TypeKeysByID` (provenencia origin), with a `TestTypeKeysByID` row the other way; `newProfiler` sets `types: vocab.TypeIDsByKey()`; delete `typeByKey`.
Check: `git grep -n 'func typeByKey\|OriginProvenencia' -- core/database/promotealign/canon.go` prints no type filter; tests pass.

**F-3 — One vocabulary lookup in the budget fixture** (MNT-2)
Files: `core/database/promotealign/bench_test.go`.
Do: move `bulkCanon`'s `idOf` and a `termOf` to file level; `bulkCanon` looks up only the ids the tree adds and passes them with the ones it holds into `bulkTree`; delete `lookupTree`'s copies.
Check: `TestLargeSourceBudgets` passes; `grep -c 'func idOf\|idOf := func' core/database/promotealign/bench_test.go` prints 1.

**F-4 — One two-subject marriage helper** (MNT-3)
Files: `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/evidence_test.go`.
Do: `func (f *fixture) marriage(src sources.Source, art artifacts.Artifact, a, b subjects.Subject) subjects.Subject`; `kinMarriage`, `TestLayerBridgesMergeKinshipShapes` and `marriageEvidence` call it.
Check: tests pass; no inline `"marriage", nil)` followed by two `"subject"` participations remains outside the helper.

**F-5 — `Recording.Clone`** (MNT-4)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/database/promotealign/links.go`.
Do: `func (r Recording) Clone() Recording` (shapes and pivots copied); `add`'s miss path uses it; `Merge` uses the same per-shape copy. A `TestRecordingMerge` row: a clone shares no pivot bytes.
Check: kinship and promotealign tests pass.

**F-6 — Map row wording** (MNT-5, CON-1)
Files: `docs/deployment-plan/spike-11/deployment-plan.md`.
Do: the S11-39 row says "through one `linkSet` merge".
Check: `python3 scripts/check-plan-docs.py` 0 errors.

The plan's Landed note gains lines for F-1 – F-5 (F-2 supersedes D-8's "`typeByKey` stays").

## Round 1 — fix
Head `87e76431`. Gate PASS (orchestrator run). F-1 – F-6 done. Deviations: `fixture.marriage` takes a role (the witness control needs it); `Shape.Clone` added with `Recording.Clone`; `linkSet.reserve(n)` after the walk (the count is unknown at creation); commits grouped (F-1+F-2, F-3+F-4).
Local warm proposal (orchestrator, three runs, load average 8.5): 142, 152, 147 ms (round 0: 160–190 ms). At #567's CI ratio (about 2.3×) that is about 330 ms against 250 ms.
CON-1/MNT-5 (F-6), MNT-3 (F-4), MNT-4 (F-5): **Resolved** (read `b90ee6d8`, `e0d0ce5c`, `87e76431`).

## Round 2 — review
Reviewers: SEC+PRF (PRF-1 blocker), MNT (MNT-1, MNT-2 majors). Scope `6be2f361...87e76431`.
MNT: MNT-1 and MNT-2 **Resolved**; no new blocker.
SEC+PRF: PRF-1 **Not resolved** — F-1 did not move the warm proposal (base 88–111 ms, round-1 head 129–142 ms, round-2 head 127–155 ms, interleaved, load 6–13). The gap is entirely the implied-link path: head with the implied loop skipped matches base. `reserve` re-hashes every walked link and grows the slice by twice the walk's links (about 7–10 ms, 22 MB per proposal); `Reader.For` is 13–16 ms per proposal; a scratch per-person `For` memo closed about two thirds of the gap; the rest is `addImplied` growing the set and the `sideLink` → `CanonEdge` copy. SEC-1: Follow-up unchanged. No new blocker.

### Fix plan 2

**F-7 — Delete `linkSet.reserve`** (PRF-1)
Files: `core/database/promotealign/links.go`, `core/database/promotealign/canon.go`.
Do: remove `reserve` and its call.
Check: tests pass.

**F-8 — Cache each person's implied links until the next commit** (PRF-1; was F-1b)
Files: `core/database/graphcache/memo.go`, `core/database/graphcache/commit.go`, `core/database/graphcache/graph_test.go`, `core/database/canonlinks/reader.go`, `core/database/canonlinks/reader_test.go`.
Failing test first: `TestNodeMemoDropsOnAnyCommit` *(new)* in `graph_test.go`: a slot with `DropOnHandleCommit` keeps entries across a first load and drops all of them on any commit that names a handle, `Drop` and a vocabulary commit. `TestForFollowsCommits` keeps passing (a person's cached answer never outlives a commit).
Do:
- `NodeMemo` gains `DropOnHandleCommit bool` *(new)*: `OnCommit` empties such a slot whenever the commit names any handle (in the same place it drops the match index), as well as on `Drop`/`clear`. A person's `For` answer reads nodes up to three hops away (a birth, the parent, the parent's events), so the slot's drop is the whole slot, not a hop count.
- `canonlinks` registers a second slot, `canonlinks.for` (`DropOnHandleCommit: true`, `DropOnVocabulary: true`, `Check` rebuilding `For` from the cached nodes and comparing). `Reader.For(id)` reads the slot first and stores its result (the full answer, the cache's shared links, with the `Rules` it was derived under, as `pivotResult` does) on a miss.
- `loadCanon` is unchanged apart from F-7 and F-9: it still filters to the `A` end.
Check: the new test passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/ ./core/database/graphcache/...` passes; `TestLargeSourceBudgets -v` warm reported (three runs).

**F-9 — Merge implied links without growing the walked set** (PRF-1)
Files: `core/database/promotealign/links.go`, `core/database/promotealign/canon.go`.
Do: `loadCanon` collects the implied links (A-end filtered, both ends walked) in their own slice; one that shares a key with a walked link merges into it (`Recording.Merge`); the rest are appended to the edges at conversion. `loadCanon` builds `[]graphalign.CanonEdge` once, sized to the total, without an intermediate `sideLink` copy of the implied links. Order stays: walked links in walk order, then implied links in `walkedPeople` order. `TestProposeKinshipShapes` and `TestLinkSet` unchanged.
Check: tests pass; `TestLargeSourceBudgets -v` warm reported (three runs, interleaved with the base `7fdd4776` if the implementer can build it in the scratchpad from `git show`; otherwise head only).

The plan's Landed note gains F-7 – F-9. After this round the PR goes ready and CI decides the warm budget; a red CI is one more fix round, then the user.

## Round 2 — fix
Head `1c9ce3f6`. Gate PASS (orchestrator run). F-7 (`reserve` deleted), F-8 (`NodeMemo.DropOnHandleCommit`; `canonlinks.for` slot caches each person's `For` until the next commit; `TestNodeMemoDropsOnAnyCommit` failed first), F-9 (implied links merged into the edge list without a second set) done.
Warm proposal, interleaved with base `7fdd4776` built in the scratchpad (implementer): base 71.6 / 76.8 / 63.1 ms, head 115.5 / 120.3 / 111.8 ms — still about 45 ms (1.7×) over base; projected CI about 260–280 ms against 250 ms. Other budgets far inside.
Note: `PROVENENCIA_GRAPH_VERIFY=1` fails `promotealign` `TestMemoFollowsCommit/a_handle_merged` on `87e76431` and on S11-50's base too (S11-50 implementer report); not this round's.

PRF-1 stays open after the two-round cap: the PR goes to the user before being marked ready.

**User decision (2026-10-26):** "mark #572 ready so CI measures it. If the warm budget is green, finish; if red, come back with the profile and options."
Merged `origin/main` (S11-51 added to the map; conflict in `deployment-plan.md` resolved: this branch's S11-39 row and working-rule wording plus main's S11-51 sentence and rows) as `ce26b56a`; gate PASS.

## Result
Gate PASS · CI **red** (Go tests 40403797939: warm proposal 297 ms against 250 ms; macOS tests green) · open: PRF-1 (CI decides) · follow-ups: SEC-1; verify-mode `a_handle_merged` on main · to user: PRF-1 (answered: run CI) · CON miss: none

**User decision (2026-10-26) on the red warm budget:** "Raise the warm budget to 350 ms (recording it as a user decision in the plan and S11-21's figures), and add an idea for walking only the part of the tree a proposal needs."

### CI fix round
**F-10 — Warm budget 350 ms**
Files: `core/database/promotealign/perf_test.go` (`budgetWarm = 350 * time.Millisecond`; the comment above the constants says the warm figure was raised from 250 ms by the user on 2026-10-26 when the fixture gained the 12-generation tree and matching read kinship links), `docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` (a dated note under D-1 and beside AC-1.2: warm raised to 350 ms by the user, 2026-10-26, S11-39; the rest of D-1 unchanged), this PR's plan (Landed note: the user's decision, quoted, with CI's 297 ms), `docs/ideas/narrower-canon-walk.md` *(new)* and `docs/ideas/README.md`.
Idea file content: every proposal walks the whole connected component of the canon from its seeds (on the large fixture about 18,000 handles and 24,000 links, plus about 6,000 implied links), so per-link work grows with the tree, not with the Source; the warm proposal on a 4,000-person connected tree costs about 115 ms locally and 297 ms on CI (S11-39). The shape to explore: walk only as far as alignment can use (the hops a layer link can correspond to from the seeds and their placed neighbors), keep the walk deterministic, and measure against `TestLargeSourceBudgets`; open questions: how far a grandparent or place chain needs to reach, and whether the 250 ms warm figure can return.
Check: `scripts/check-pr.sh` PASS; `python3 scripts/check-plan-docs.py` 0 errors.
