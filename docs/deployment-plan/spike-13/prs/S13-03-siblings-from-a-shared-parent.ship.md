# Ship log — S13-03 Siblings from a shared parent

**Plan:** [`S13-03-siblings-from-a-shared-parent.md`](S13-03-siblings-from-a-shared-parent.md) · **Branch:** `s13-03-siblings-from-a-shared-parent` · **PR:** #666
**State:** ready

## Round 0 — implement
Base `067be776` (plan written against `0c166624`; `main` is 23 commits ahead).
Head `b205dcd5`. Gate PASS (orchestrator run against `067be776`). PR #666 (draft).
Report deviations:
1. AC-1.8 order: the test asserts the order of Ann's `n.Links` (graph sorts links by ref as text, so C13, C14, C1 … C12), not the literal `C1` … `C14`. Recorded in the Landed note.
2. D-8's `d.load(ids)` written as `d.src.load`, matching D-7's `nodeSource`.
3. Private `idLen = 16` in `canonlinks` for D-8's 16-byte rule.
4. `TestSiblingsLeaveMatchingAlone` has its own deep-copy helpers, copies of closures inside `TestRelationships/AC-1.6`.
5. The sibling memo stores `n.ID` rather than the caller's `id` slice.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes). SEC+PRF routed for `core/kinship` (matching layer) and `core/database/canonlinks` (graph cache).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC+PRF | blocker | Duplicate parent links × children: `kinship.Siblings` (and `chains`, through the shared `through`) allocates D×C edges; a crafted catalog runs the app out of memory | To user | verified: reviewer's probe, 2,000 dup × 2,000 children → 3.8 GB, 1.4 s; ~4.6× alloc per doubling. `through` copies the q→person steps once per child (`siblings.go:44`). The fix needs a registry number in `indexSteps`, which contradicts D-5 ("no `Rules` argument … linear in `len(edges)`"); D-5's linear premise is false on the duplicates axis. The `chains` product predates this PR (`067be776:core/kinship/kinship.go:350`) |
| MNT-1 | MNT | major | `neighborsOf` is the third "every neighbor of a node" loop (`around`, `deps`) | Fix now → F-2 | verified `reader.go:131-138`, `:250-253`, `:539-545`; same-pass cleanup |
| MNT-2 | MNT | major | `checkSiblings` repeats `checkFor` but for the derive step | Fix now → F-3 | verified `reader.go:84-129`; D-7 asks the next slot to be a registration |
| MNT-3 | MNT | major | `idLen` is a private copy of graphcache's 16-byte rule | Fix now → F-4 | verified `reader.go:16`, `:476`; graphcache `uniqueIDs`, `Cached`, `SetMemo` already refuse wrong-length ids. D-8's behavior holds without it |
| MNT-4 | MNT | major | `TestSiblingsLeaveMatchingAlone` copies AC-1.6's clone closures | Fix now → F-5 | same-pass cleanup; Round 0 deviation 4 |
| MNT-5 | MNT | minor | "event links then `kinEdges`" written three times | Fix now → F-6 | verified `reader.go:324`, `:497`, `:521`; code this PR adds |
| CON-1 | CON | minor (re-rated from nit) | `Implied.From` carries a parent edge twice when the person's and the parent's edges both hold it | Fix now → F-1 | verified: `siblingsOf` concatenates the person's `kinEdges` with each parent's; a filed q→person link and a birth naming both sit in both lists, so `stepsOf` indexes each twice. S13-04 shows `From` as provenance; it also doubles D for SEC-1 |

Nits (logged): `MNT-6 · core/database/canonlinks/reader.go:472 · split siblingsOf into per-person and parents steps`

SEC-1 is To user; the loop pauses for it.

User on SEC-1 (2026-11-01): "Registry cap, keep first N". This supersedes D-5's "no `Rules` argument" for the duplicates axis; there is still no bound on a parent's children.

### Fix plan 1

**F-0 — A registry cap on the steps kept per parent–child pair, shared by `Siblings` and `chains`** (from SEC-1; user decision; supersedes D-5's "no `Rules` argument")
Files: `core/kinship/kinship.go`, `core/kinship/siblings.go`, `core/kinship/siblings_test.go`, `core/kinship/kinship_test.go`, `core/matchregistry/kinship.go`, `docs/matching-model.md` (knob inventory), `core/database/canonlinks/reader.go`, the plan's Landed note. Skill: `add-match-rule`.
Failing test first: `TestSiblingsManyDuplicatesTimesChildren` in `siblings_test.go`: 2,000 `parent` edges q→person (each its own association) plus 2,000 edges q→child_i; `Siblings` returns 2,000 links, each `From` holds at most `MaxPairSteps` + 1 edges, and the call finishes under a second (same shape as `TestDeriveManyDuplicateParentSteps`). A matching `TestDeriveManyDuplicatesTimesGrandchildren` for `chains` (D dup g→m × C children of m): each grandparent link's `From` is bounded, under a second.
Do: add `MaxPairSteps int` to `kinship.Rules` with a doc comment ("most steps kept for one parent–child pair; later duplicates of the same pair are not read. Keeps `Siblings` and `chains` linear"). `indexSteps(steps, maxPairSteps)` appends a step to `byPair[p]` only while `len(byPair[p]) < maxPairSteps` (keep the first N in input order; the pair, its parent and child stay indexed, so no sibling or grandparent is lost). `Derive` passes `rules.MaxPairSteps`. `Siblings(person, edges, terms, rules Rules)` and `Parents(person, edges, terms, rules Rules)` take `Rules` and pass it to `indexSteps`. In `Siblings`, compute the person side `through(q→person)` once per parent, not once per child. `matchregistry.defaultKinship` sets `MaxPairSteps: 10` with `// decided: user, 2026-11-01 (S13-03 SEC-1): keep the first 10 steps of one parent–child pair, so duplicate links times children stays linear; real data files a handful`. Add its row to the knob inventory in `docs/matching-model.md` next to `kinship.MaxPivotEnd`. `canonlinks.siblingsOf` passes `d.rules`. Any registry validation that checks `MaxPivotEnd` > 0 checks `MaxPairSteps` the same way (follow `add-match-rule`). Landed note: D-5 superseded by SEC-1 for duplicates.
Check: the two new tests pass; `TestDeriveManyDuplicateParentSteps`, all `kinship`, `matchregistry`, `canonlinks`, `promotealign`, `graphalign` suites and goldens pass unchanged; `TestLargeSourceBudgets` passes; scripts/check-pr.sh PASS.

**F-1 — An edge in both the person's and a parent's list is read once** (from CON-1)
Files: `core/database/canonlinks/reader.go`, `core/database/canonlinks/reader_test.go`.
Failing test first: a `TestRelationshipsOptionsSiblings` case: Ann filed as parent of Bea (`AB`) and of Cal; Bea's sibling link {Bea, Cal} through Ann has `From` holding `AB` exactly once (today twice). Add the birth variant if cheap (a birth with Bea as subject and Ann as father).
Do: in `siblingsOf`, when building `all` for a node, skip a parent's edge already in the node's own edges or added from another parent. Identify an edge by the link it came from (its associations / link identity the `kinEdges` edge already carries); do not compare by end ids alone, since two distinct links between the same two people are two steps.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — One "every neighbor of a node" loop** (from MNT-1)
Files: `core/database/canonlinks/reader.go`.
Do: `around` calls `d.src.load(neighborsOf([]*graphcache.Node{n}))`; `deps` appends each root and then `neighborsOf` of the cached roots, keeping today's output order (root, then its neighbors, per root).
Check: `grep -n "l.Neighbor" core/database/canonlinks/reader.go` prints only `neighborsOf` and `eventLinksAt`; suites pass with `PROVENENCIA_GRAPH_VERIFY=1`.

**F-3 — One per-person stored-answer check** (from MNT-2)
Files: `core/database/canonlinks/reader.go`.
Do: add `checkPersonAnswer(derive func(d *deriver, id []byte) ([]kinship.Implied, bool, error)) graphcache.MemoCheck` (or the repo's check func type) doing the type assert, `newDeriver(g, res.rules, cachedNodes(g))`, derive, compare, and the two error strings. The For slot passes a derive wrapping `forLinks` (always found); the sibling slot passes one wrapping `siblingsOf([][]byte{id}, d.eventLinks)` returning found=false for no entry ("is for no cached person"). Delete `checkFor` and `checkSiblings`.
Check: `TestVerifyCatchesAStaleSiblingAnswer` and the For verify tests pass; `PROVENENCIA_GRAPH_VERIFY=1 go test ./core/database/canonlinks` passes.

**F-4 — No private id-length rule** (from MNT-3)
Files: `core/database/canonlinks/reader.go`.
Do: delete `idLen` and the length check in `siblingsOf`; keep the `seen` dedup; drop "or not 16 bytes" from its doc.
Check: `rg 'idLen' core/database/canonlinks` → no hit; AC-1.9 test passes; add a wrong-length id (e.g. 3 bytes) to the AC-1.9 case if cheap: no entry, no error.

**F-5 — One pair of clone helpers in `reader_test.go`** (from MNT-4)
Files: `core/database/canonlinks/reader_test.go`.
Do: move `cloneLinks` / `cloneImplied` from `TestRelationships/AC-1.6` to file level next to `relShapes`; `TestSiblingsLeaveMatchingAlone` uses them; delete `copyImplied` / `copyLinks`.
Check: `rg 'copyImplied|copyLinks' core` → no hit; tests pass.

**F-6 — One "a person's parent edges" step** (from MNT-5)
Files: `core/database/canonlinks/reader.go`.
Do: add `func (d *deriver) parentEdges(n *graphcache.Node, event func([]byte) ([]kinship.Implied, error)) ([]kinship.Edge, error)` = `eventLinksAt` then `kinEdges`. Use it in `chainLinks` and both `siblingsOf` loops; `forLinks` keeps `kinEdges` (it needs `atEnd` too).
Check: suites pass; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `cb3aa35a`. Gate PASS (orchestrator run). F-0 – F-6 done in one commit (the F-0 signature change and the `reader.go` hunks interleave).
Report deviations:
1. `indexSteps` treats a cap below 1 as 1: `chains` reads `byPair[up][0]`.
2. `through` replaced by `edgesOf` and `associationsOf`; `Siblings` computes the person side once per parent.
3. No registry validation added for `MaxPairSteps`: `MaxPivotEnd` has none to follow.
4. `PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/promotealign` fails `TestMemoFollowsCommit/a_handle_merged` (`links: count 1, catalog 0`). Orchestrator reproduced it on `main` at `067be776`, so it predates this PR. Follow-up.

CON-1 (minor, F-1): **Resolved.** `siblingsOf` skips a parent edge whose `edgeKey` (ends + associations, joined with no separator) it has read; `TestRelationshipsOptionsSiblings/an_edge_in_both_lists_is_read_once` failed first with `[AB AB AC]`.

## Round 2 — review
Reviewers: MNT, SEC+PRF (each had a Fix-now blocker or major). CON not re-sent: its one finding was a minor, resolved above.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 – MNT-5 | MNT | — | Round-1 Fix-now findings | Resolved | reviewer re-checked each (`l.Neighbor` grep, `checkPersonAnswer`, `rg idLen`, file-level clone helpers, `parentEdges`) |
| SEC-1 | SEC+PRF | — | Duplicates × children | Resolved | reviewer probe: three shapes to 8,000 edges, ~11–13 ms and 30–35 MB at n = 8,000, linear; largest `From` 20 (2 × cap). Round 1 was 3.8 GB at 2,000 × 2,000 |
| SEC-2 | SEC+PRF | below round-2 bar | `edgeKey` joins ends and associations with no separator; ids are not length-checked at the catalog boundary (`graphcache/load.go:344-346`, no `length(id) = 16` CHECK), so a crafted catalog could make two edges share a key and hide one sibling or `From` entry | Follow-up | round 2 only adds blockers; no panic, out-of-bounds read, or cross-project leak. Fix: length-prefix the join, or skip non-16-byte ids. Corrects the round-1 fix note above, which said the join cannot collide |

## Finish
Merged `origin/main` (`80ca74ba`, docs only) → `275d09b7`. Gate PASS against `80ca74ba`.

## Result
Gate PASS · CI pending · open: none · follow-ups: SEC-2 (`edgeKey` separator), `TestMemoFollowsCommit/a_handle_merged` fails under `PROVENENCIA_GRAPH_VERIFY=1` on `main` · to user: SEC-1 (answered 2026-11-01: registry cap, keep first N; `MaxPairSteps` = 10) · CON miss: none
