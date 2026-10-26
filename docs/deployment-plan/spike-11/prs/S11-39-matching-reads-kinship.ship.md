# Ship log — S11-39 Matching reads kinship shapes

**Plan:** [`S11-39-matching-reads-kinship.md`](S11-39-matching-reads-kinship.md) · **Branch:** `s11-39-matching-reads-kinship` · **PR:** #567
**State:** stopped — CI cold still red after Fix plan 4 (to user)

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

## Round 1 — fix
Head `823856b0`. Implementer's gate PASS without F-1. F-2 done (`350f777f`), F-4 done (`5a29a0a6`), F-3 done (`823856b0`): the shared large-source fixture now holds a 12-generation tree (8,190 parent links, 4,095 marriages); warm 123 ms vs 73 ms without it (budget 250 ms), cold 1.11 s (2.5 s), after a commit 478 ms (1.5 s). The memo stays out per D-8. Deviation: the tree lookups run before `bulkCanon`'s transaction (one connection; a query inside it deadlocked).

**Stopped on F-1.** `core/matchregistry/guard_test.go` marks `core/kinship` `numbers: true`: "the registry owns every tunable". `const maxPivotEnd = 10` fails `TestLayerPackages/core/kinship/numbers`. Fix plan 1 was wrong to call the bound "not a registry knob". The patch (code + `TestDeriveBound`, 10 vs 11 at each pivot) is saved, unapplied. **To user:** where the bound lives.

**F-1, user's answer (2026-10-26):** "Put the cap in core/matchregistry as a registry value (10), read in promotealign and passed into kinship.Derive."

### Fix plan 1, F-1 revised

**F-1′ — Bound what one pivot can imply, through a registry value** (from SEC-1; user-approved twice)
Files: `core/matchregistry` (the registry value and its doc), `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/database/promotealign/terms.go` (`withImplied`) and its callers in `layer.go`/`canon.go`, `docs/matching-model.md` knob inventory. Skill: `add-match-rule` (follow its steps for a registry knob, including the knob inventory row).
Start from the saved patch `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/e7ccccd4-c472-4924-8362-33987dd34c10/scratchpad/f1.patch` (logic and `TestDeriveBound`), then:
Do: add the value to the registry where values of its kind live (as `PrefixMinLength` does), named for what it bounds (for example `Kinship.MaxPivotEnd`), value 10, with a comment: a pivot (an event, or the parent between in a chain) with more participants than this at either end reads as malformed data and implies nothing; it keeps derivation linear in its input. `kinship.Derive` takes the bound as an argument (no numeric literal in `core/kinship`; it must not import `matchregistry`). `promotealign` reads it from `matchregistry.Default()` and passes it through `withImplied`. `TestDeriveBound` passes the bound in and pins 10 vs 11 at each end; `BenchmarkDerive` passes the registry's value or a test-local one. Update `core/kinship`'s package doc ("no numbers" stays true: the bound comes in as an argument).
Check: `TestDeriveBound` rows pass (fail without the skip); `TestLayerPackages` passes unchanged; scripts/check-pr.sh PASS. Add an F-1 line to the plan's Landed note.
F-1′ head `16ff9ef2`. Gate PASS (orchestrator run). `kinship.Rules{MaxPivotEnd}` from `matchregistry.Default().Kinship` (10), passed through `withImplied`; knob inventory row added; `TestLayerPackages` unchanged.

## Round 2 — review
Reviewers: SEC+PRF (round-1 majors SEC-1, PRF-1). Not sent back: CON (clean), MNT (minors only). MNT-2 read by the orchestrator in `layer.go`/`canon.go`: **Resolved**. SEC-2 (F-4): resolved, `linkSet` keeps a per-link association set.

SEC-1: Resolved for events and distinct parents. PRF-1: Resolved (budget case: warm 140 ms / 250 ms, cold 1.31 s / 2.5 s, after a commit 609 ms / 1.5 s).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-3 | SEC/PRF | blocker | `chains` dedupes grandparent associations with `slices.ContainsFunc` over a growing slice: quadratic in a parent link's associations, which `MaxPivotEnd` does not bound; `Rules`'s comment says Derive is linear | Fix now → F-5 | verified: `kinship.go:280-286`; reviewer's probe, 16,000 associations on one g→m link × 10 grandchildren → 4.1 s |

### Fix plan 2

**F-5 — Dedupe chain associations through a set** (from SEC-3)
Files: `core/kinship/kinship.go` (`chains`), `core/kinship/kinship_test.go`.
Failing test first: add a `BenchmarkDerive` case (or a `TestDeriveBound` row with a time-independent check) for one parent link g→m carrying several thousand associations plus m→c1..c10; and a `TestDerive` row asserting the implied grandparent link's associations are deduped and in first-added order (duplicates across the two steps appear once).
Do: in `chains`, replace the `slices.ContainsFunc(assoc, …)` scan with a `map[string]struct{}` keyed by `string(a)`, keeping first-added order, the same shape F-4 used in `linkSet.add` (`core/database/promotealign/terms.go`). Optionally collect `from` and the deduped associations once per `(g,m)` pair and reuse it across that pair's grandchildren. Keep the `Rules` comment's "linear in its input" true.
Check: the new rows pass; `go test ./core/kinship/ -bench Derive` shows the many-association case roughly linear (report the timings at two sizes); scripts/check-pr.sh PASS.

## Round 2 — fix
Head `9c29cbaf`. Gate PASS (orchestrator run; the implementer's first run hit an unrelated Swift debounce flake, rerun passed). F-5 done: `chains` dedupes through a set; `TestDeriveManyAssociations` (20,000 associations, each twice, × 10 grandchildren) 11.8 s before → 0.03 s after. Merged `origin/main` (S12-14 plan docs only) as `d2194f2e`; gate PASS.


## Manual check — user finding (2026-10-26)
The user's tree records a marriage with both people as `subject`; a Source with a `spouse` relationship links one person (by name) but not the other. Cause: `core/kinship/patterns.go:19` reads only role `spouse` at both ends of a `marriage`, so a `subject` couple implies nothing. The `spouse` role came from `kinship-shapes.md` ("both `spouse` in one `marriage`"), and every golden used it. The evidence sheet showing nothing for a tree-side marriage is AC-3.3 (S11-40), not this bug.

**User decision (2026-10-26):** "We should align around the subject being the principal role of an event: In a birth, the subject is the person being born. In a death, the subject is the person dying. In a marriage, the subjects are the people getting married. In a migration, the subject is the person doing the migration. I think we should remove the spouse role from event participation." And: "Then you can fix the rest of these." This authorizes a third fix round past the two-round cap.

### Fix plan 3

**F-6 — A marriage's couple is its subjects; the `spouse` participation role is removed** (user decision)
Files: `core/kinship/patterns.go`, `core/kinship/kinship_test.go`, `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/registry.go` (+ its tests, `TestVocabKeysMatchSeed`), `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/evidence_test.go`, any other test or fixture that files a participation with role `spouse` (grep below). Skills: `add-match-rule`, `add-seeded-vocabulary`.
Failing test first: in `TestProposeKinshipShapes`, switch the marriage fixture (`kinMarriage`, and `participation(..., "spouse")` calls) to role `subject`; AC-1.2 (tree marriage with two subjects vs Source `spouse` relationship) and AC-1.1 must fail before the pattern change and pass after with today's scores (3.375 medium; witness control 2.5).
Do:
1. `patterns.go`: the marriage pattern reads `vocabkey.RoleSubject` at both ends (symmetric, as today). Rename nothing else; `PatternMarriageSpouses` and the implied `relationship · spouse` term stay.
2. Remove the `role` seed `spouse` from `subjectvocab/registry.go` (the `relationship_type` `spouse` stays) and the `vocabkey.RoleSpouse` constant. Seeds install at create time only; existing catalogs keep their `spouse` role term and get no migration or backfill (dev phase, "High tolerance for churn") — say so in the PR body.
3. Every test or fixture that files a participation with role `spouse` moves to `subject` (marriage) or, where it means a spouse named on another event, to a relationship or `witness` as the test's intent needs; keep each test's assertion. Find them by body: `grep -rnE 'RoleSpouse|participation\([^)]*"spouse"|kinMarriage\("spouse"\)|"spouse", "marriage"' core api macos` must print nothing at the end. A `relationship_type` `spouse` / `TermSpouse` use is correct and stays.
4. Living docs: `docs/deployment-plan/spike-11/kinship-shapes.md` (the pattern rows and §Shape text: "both `subject` of one `marriage`"), the map `deployment-plan.md` working rule "Three patterns, product terms only" (line ~98: "two `subject`s of one `marriage` imply `spouse`", noting the user decision 2026-10-26) and add a working rule: "**The subject is an event's principal** (user, 2026-10-26): a birth's subject is the person born, a death's the person dying, a marriage's subjects the couple, a migration's the person migrating. No `spouse` participation role.", `docs/matching-model.md` glossary row "kinship pattern" ("two subjects of one marriage"). Do not edit landed PR plans or ship logs other than S11-39's; add an F-6 line to the S11-39 plan's Landed note.
Check: `go test ./core/...` (with `-tags fts5` where needed) passes; the grep prints nothing; `TestDeriveBound`'s spouse rows use `subject`; scripts/check-pr.sh PASS.

## Round 3 — fix (user-authorized)
Head `c14e0d3c`. Gate PASS (orchestrator run, against `13dab35f`; `main` has not moved since). F-6 done: marriage pattern reads `subject` at both ends; `role` seed `spouse` and `vocabkey.RoleSpouse` removed; the plan's grep prints nothing. Deviations: the Swift `SeededRoleTerm.spouse`, `L10n` `roleSpouse` and three xcstrings keys removed (the Swift vocabulary parity test requires it); `segmentcomparetest` terms updated; `lifespan_test.go` and `invariants_test.go` keep a literal `"spouse"` as an arbitrary non-subject role. No UX review: the Swift change only deletes the seeded term's wording; `L10n.wording` returns nil for an unseeded term, so an existing catalog's `spouse` role shows its stored label like any catalog term. Nits (logged): `ORC-1 · lifespan_test.go, invariants_test.go · a non-subject role could be witness now that spouse is unseeded`.

## Result
Gate PASS · CI **red** (Go tests, `TestLargeSourceBudgets`) · open: CI budgets · follow-ups: none · to user: SEC-1 (answered: bound of 10, as a `matchregistry` value); marriage role (answered: subject is an event's principal; spouse role removed) · CON miss: none

## CI — budgets red (2026-10-26)
Correction: the earlier "CI green" was wrong. `gh pr checks --watch | tail` showed only the macOS workflow; **Go tests** has failed on every non-draft head since F-3 (`0b0e47bb`, `dddf84ed`, `216e0872`, `c14e0d3c`, `893ad748`): warm 278–283 ms (budget 250 ms) or cold 3.30–3.48 s (budget 2.5 s). Earlier heads were drafts, so CI skipped them. Locally the test passes.

Local measurement on `893ad748` (orchestrator, `withImplied` in `loadCanon` disabled for one run, not committed):

| | Cold | Warm | After a commit |
| --- | --- | --- | --- |
| Branch | 1.04 s | 123 ms | — |
| Canon without implied links | 0.97 s | 94 ms | 438 ms |
| Before F-3's tree (round 1 report) | 0.73 s | 73 ms | 377 ms |

CI runs about 2.3× (warm) to 3.3× (cold) slower than this Mac. On that ratio the warm failure comes from the canon's derivation (94 → 123 ms locally, about 30%), the trigger D-8's Out line named ("revisit if a budget fails"). The cold failure comes mostly from walking the larger tree itself (0.73 → 0.97 s with no derivation, est. 3.2 s on CI), a cost from before this PR that F-3's fixture measured for the first time. **To user.**

**User decisions (2026-10-26):** "Memoize the canon's implied links on this PR, keep the 12-generation tree, and fix the canon walk's cold cost on this PR too." After the design measurements: "Yes: memoize the whole canon walk and add the catalog-wide association scan for the shortlist kinds as Fix plan 4, then read CI's cold number before deciding more." D-8 is reversed.

Design (read-only Plan agent; prototypes in a `git archive` copy at `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/e7ccccd4-c472-4924-8362-33987dd34c10/scratchpad/design/tree`, env-switched; local, noisy): HEAD cold 1.03 s / warm 123 ms; implied-only memo warm ~100 ms; whole-walk memo warm 29–36 ms; plus association scan cold 0.82–0.85 s, after a commit 394 ms. A values-only kind read made cold worse (1.21 s: the walk reaches all 18,233 handles through the shared city place). Cold breakdown on HEAD: kind loads 551 ms (links 430), candidate ranking ~200 ms (predates the PR), walk 61–67 ms, implied 17–28 ms.

### Fix plan 4

**F-7 — Memoize the canon walk** (reverses D-8; user decision)
Files: `core/database/promotealign/canon.go`, `export_test.go`, `memo_test.go`. Reference prototype: `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/e7ccccd4-c472-4924-8362-33987dd34c10/scratchpad/design/tree/core/database/promotealign/canon.go` (ignore its Z* env switches).
Failing test first: `TestCanonWalkMemo` in `memo_test.go`: after two proposals with no change the memo exists and the second reuses the same backing array (`&edges[0]` equal, as `TestMemoIsNotMutated` does); subtests modelled on `TestCandidatesMemoFollowsCommits`, each asserting the warm proposal deep-equals a proposal after `Drop`: "no commit", "a handle commit", "a parent relationship filed between two filed persons" (an association commit adding an implied grandparent; build with `kinTree`/`kinSide`), "pinned, then unpinned".
Do:
1. Move the walk out of `loadCanon` unchanged (from `set := newLinkSet()` through the edges conversion) into `func walkCanon(g *graphcache.Graph, seeds [][]byte, handles map[string]*graphcache.Node, kinds map[string]string, props map[string]graphcache.VocabProperty, types map[string][]byte, terms segmentcompare.Terms) ([]graphalign.CanonEdge, error)`; it still fills `handles` and `kinds`.
2. `const canonWalkMemoKey = "promotealign.canonWalk\x00"`; `type canonWalk struct { gen uint64; seeds string; edges []graphalign.CanonEdge; handles []walkedHandle }`, `type walkedHandle struct { id []byte; kind string }` (handles in Ref-then-ID order). A stored value is never mutated.
3. `func seedsKey(seeds [][]byte) string`: sort, dedupe, concatenate.
4. `func loadCanonWalk(g, graphID, seeds, handles, kinds, props, types, terms) ([]graphalign.CanonEdge, error)`: `gen := matchIndexOf(g).gen` (read after `loadCandidatesMemo`). If `gen != 0` and the memo on the Evidence graph memo (beside `candidatesMemo`, same lifecycle: `Drop`, vocabulary commit, `SetAliasIndex`, `SetNameFormat`, a commit to that graph) has the same `gen` and `seeds`: add each stored handle missing from `handles` via `g.Node(id)`, set its kind, return `m.edges` (shared, read-only). Else `walkCanon`, and store a new `canonWalk` when `gen != 0`.
5. `loadCanon` calls `loadCanonWalk`; its doc comment says why the memo is safe, in the style of `candidatesMemo`'s.
6. `export_test.go`: `CanonWalkForTest(c *database.Catalog, graphID []byte) ([]graphalign.CanonEdge, bool)`; add `CanonEdges []graphalign.CanonEdge` to `MemoSnapshot` and fill it in `MemoSnapshotForTest`, so `TestMemoIsNotMutated` covers it via `memoJSON`.
7. Plan doc: replace D-8 with the walk memo and these measurements; remove the Out row "Memoizing the canon's implied links"; add a §11 risk row (memo memory ~6–10 MB per Evidence graph, at most 8; a pin changes the seeds and misses). Landed note: an F-7 line.
Check: `TestCanonWalkMemo` passes; `TestLargeSourceBudgets` warm, "memo matches a fresh run" and "warm queries" pass; all goldens unchanged; scripts/check-pr.sh PASS. Report warm.

**F-8 — One catalog-wide association scan for the shortlist's kinds** (cold; user decision)
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/load.go`, `core/database/graphcache/graph_test.go`, `core/database/promotealign/index.go`. Reference prototype: `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/e7ccccd4-c472-4924-8362-33987dd34c10/scratchpad/design/tree/core/database/graphcache/zscan.go` and `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/e7ccccd4-c472-4924-8362-33987dd34c10/scratchpad/design/tree/core/database/promotealign/index.go`. Skill: `add-catalog-query` (SQL conventions).
Failing test first: `TestLoadKindsMatchesNodeLoads` (fails to compile until `LoadKinds` exists). Direct inserts as in `TestCommitOfAnAssociationLoadsEndpoints`, covering: a participation with a role term; a relationship with a parent term; a location with no disambiguation; a place relationship with a term; a merged association; an association whose endpoint value is rank 2 or not kept; an association to a merged neighbor; a self-loop relationship. `graphcache.New(db).LoadKinds(person, event, place)` vs a second `graphcache.New(db)` loading each id with `Node`: links `reflect.DeepEqual` for every node, and kind lists equal `Kind`'s.
Do:
1. `graph.go`: move `Kind`'s first-read logic into `func (g *Graph) kindIDs(typeID []byte) ([][]byte, error)` (sets `kindOn`, `kinds`, `kindSet`); `Kind` calls it, otherwise unchanged. Add `func (g *Graph) LoadKinds(typeIDs [][]byte) error`: collect `kindIDs` per type, keep ids not in `g.nodes`, read them in `sqlbatch.Each` chunks with `readNodes`, then `scanLinks(present, rows)`, set each node's `Links`, then `install` (no node published before `install`). Doc: it reads every association in the catalog, cheaper when the kinds hold most handles; `Kind` then returns them without a query.
2. `load.go`: split `loadChunk` into `func (g *Graph) readNodes(ids [][]byte) ([]*Node, map[string]entityRow, error)` (entity rows, `removeNode` for absent ids, values, members; no links); `loadChunk` = `readNodes` + `loadLinks` + `install`. Add `func keepRaw(r linkRaw) bool` (`connectrules.LookupBridge` + `bridgeEndpoints`) used by both loaders. Add `func assembleLinks(raws []linkRaw, marks map[string]termMark) map[string][]Link` taken from `finishLinks`, which first stable-sorts raws by (self, assoc, selfProp, neighbor) so first-wins does not depend on row order; `finishLinks` = disambiguation + `assembleLinks`. Add `func linkScanSQL(valueTable, entityTable string) (assocQ, endsQ, termsQ string, assocArgs, endsArgs, termsArgs []any)` and export `LinkScanSQL()` beside `LinkLoadSQL`:
   - associations: every unmerged entity whose subject type is a registered bridge (key, origin) pair from `bridgePredicates`, via `e.subject_type_id IN (SELECT id FROM subject_types WHERE …)`; returns id, ref, type key;
   - endpoints: `SELECT a.entity_id, a.property_id, p.key, p.origin, a.value_entity_id FROM <values> a INDEXED BY auto_reconciler_values_edge_idx CROSS JOIN properties p ON p.id = a.property_id WHERE a.rank = 1 AND a.reason = 'kept' AND a.value_entity_id IS NOT NULL AND a.property_id IN (SELECT id FROM properties WHERE key IN (<every bridge endpoint key>))` (keys at any origin on purpose; the self side's (key, origin) is checked in Go against the registered pairs, as `bridgeEndpoints` filters today);
   - terms: the existing disambiguation SELECT with `r.property_id IN (SELECT id FROM properties WHERE key IN (<disambiguation keys>))` instead of an IN list of ids; the mark per association and property key is the least `COALESCE(t.key, '')` (today's `ORDER BY t.key`, first wins).
   Add `func (g *Graph) scanLinks(ids [][]byte, known map[string]entityRow) (map[string][]Link, error)`: run the three queries, keep only associations from the first; `entityRows(missing)` for neighbors not in `known`; build `linkRaw`s (a self end is in the wanted ids with a registered (key, origin); each other end of the same association has a different property id, an existing row and an unmerged neighbor); filter with `keepRaw`; `assembleLinks`.
3. `index.go` `loadShortlist`: collect the 16-byte type ids of `sorted` and call `g.LoadKinds(ids)` before the per-kind loop; loop and `keyValues` unchanged.
4. Leave `conclusionheaders/compose.go`'s `Kind` callers and `verify.go`'s per-id `loadLinks` (Verify stays the per-id oracle) unchanged.
5. Also add `TestLoadKindsThenKindIssuesNoQuery` (`SetQueryCounterForTest`) and `TestLinkScanUsesIndexes` (`dbtest.QueryPlan` over `LinkScanSQL()`: no `SCAN` of `auto_reconciler_values` or `canonical_entities`; endpoints use `auto_reconciler_values_edge_idx`).
6. Plan Landed note: an F-8 line (a second link loader, kept equal by `keepRaw`/`assembleLinks`, the equivalence test and Verify; self-loop duplicate rule now deterministic).
Check: the new tests pass; `go test -tags fts5 ./core/database/...` passes; `PROVENENCIA_GRAPH_VERIFY=1` run of `./core/database/promotealign` passes; `TestLargeSourceBudgets` cold and after-a-commit pass; scripts/check-pr.sh PASS. Report cold, warm and after-a-commit.

Not this round (user: read CI's cold number first): parallel candidate ranking, a ranking-only compare path, scanning values.

## Round 4 — fix (user-authorized)
Head `3620aa7f`. Gate PASS (orchestrator run). F-7 done (`dbd674c9`, `TestCanonWalkMemo`), F-8 done (`3620aa7f`, `TestLoadKindsMatchesNodeLoads`, `TestLoadKindsThenKindIssuesNoQuery`, `TestLinkScanUsesIndexes`). Deviations: associations query starts from `subject_types` (same result, plan test passes). `PROVENENCIA_GRAPH_VERIFY=1` fails only `TestMemoFollowsCommit/a_handle_merged` ("links: count 1, catalog 0"); the implementer reports it fails on `origin/main` too (raw-SQL merge then `OnCommit`) — not verified by the orchestrator yet.

Local (orchestrator, load average 8–10, unreliable): cold 1.06–1.18 s, warm 33–42 ms, after a commit 536–542 ms.

**CI, Go tests run 40353254720 on `3620aa7f`: failure.** Warm passes. Cold 3.29 s (budget 2.5 s; before: 3.30–3.48 s). After a handle commit 1.75 s (budget 1.5 s; passed before). Cause of the after-commit regression, read in code: `LoadKinds` (`graphcache/graph.go:276-315`) runs the catalog-wide `scanLinks` for any non-empty set of uncached ids, so one handle missing after a commit reads every association in the catalog. F-8 bought no measurable cold gain on CI. **To user.**
