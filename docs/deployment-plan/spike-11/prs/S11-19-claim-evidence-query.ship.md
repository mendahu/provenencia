# Ship log — S11-19 Claim evidence query

**Plan:** [`S11-19-claim-evidence-query.md`](S11-19-claim-evidence-query.md) · **Branch:** `s11-19-claim-evidence-query` · **PR:** #513
**State:** ready

## Round 0 — implement
Base `main` @ `9ca7472e`. Head `8262d21e`. Gate PASS (orchestrator rerun). Report deviations (also in the plan's Landed note):
- D-10's collapse key also includes outcome and property cause; equal whenever the plan's key parts are equal.
- A corresponding link whose bridges all name non-members keeps its Relationship with no bundles (D-13 read as "skip the bridge, not the link").
- The loader leaves out a filed bridge Subject whose endpoint Observations do not name both ends.
- STEP-5 and STEP-6 are one commit; `subjects.SourcesOf` has its own.
- The implementer committed this ship log early, in commit 2.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (matching layers, `core/database/**`). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Every accepted member of every neighbor entity is loaded, but `Build` reads only the ones a bridge names | Fix now → F-1 | verified: `evidence.go:136` `q.members(hid)` for each neighbor; `build.go` `bundles` uses only `n.Members[ni]` that a `Bridge.Neighbor` matches |
| PRF-2 (+ MNT outside-lens PRF) | PRF | major | `members` calls `layerValues` once per member (clone and sort of the whole Source each time); `bridges` scans a Source's Observations once per bridge | Fix now → F-2 | verified: `evidence.go:220-224`, `profiles.go:94-95`, `evidence.go:299`. The batched SQL read the reviewer offers is ruled out by D-8 ("no second Observation read"); grouping by Source is within D-8 |
| MNT-1 | MNT | major | "A layer bridge from one end" rule is copied from `graphalign`'s `newState` into `Evidence` | Fix now → F-3 | verified: `align.go:123-127` and `evidence.go:82-89`. Same-pass cleanup |
| MNT-2 | MNT | major | `bridges` reads endpoint Observations with no polarity check and keeps the last value when an end names two Subjects | Fix now → F-4 | verified: `evidence.go:299-311`; `bridgeNeighbors` reads `polarity = 'positive'` only, `resolveEndpoint` treats two values as unresolved |
| SEC-1 | SEC | minor | `req.Neighbors` has no length bound; `checkFixed` queries before the duplicate check | Fix now → F-5 | verified: `evidence.go:47-57` |
| MNT-3 | MNT | minor | `routeOf` re-derives which way `Rules.Lead` chose the lead | Fix now → F-6 | verified: `sections.go:36-44`, `subjectcompare.go:167-180`. Minor in code this PR adds |
| PRF-3 | PRF | minor | Each `Evidence` call rebuilds the Source's whole layer | Later PR | written into the S11-21 map scope entry (`deployment-plan.md`, S11-21 **Ships**) on this branch |
| MNT out-of-lens CON | CON | — | `relationships` compares against `canons[0].Key` only | Rejected | `segmentcompare.Corresponds` is `a == b` (`key.go:44`), so every corresponding canon key equals `key` |

Nits (logged): `CON-1 · ship log header stale (updated by orchestrator)` · `CON-2 · completed.md says "In review" while the plan says Landed` · `MNT-4 · sections.go:131 · struct key for collapse map` · `MNT-5 · build.go:44 · boolRank(!isOne) reads reversed` · `MNT-6 · build_test.go:158 · equalIDs is slices.EqualFunc` · `MNT-7 · evidence.go:111-131 · g.Node fetched twice per neighbor` · `MNT-8 · promotealign.go:39,54 · propose repeats loadInputs' checks`

### Fix plan 1

**F-1 — Load only the neighbor members a corresponding bridge names** (from PRF-1)
Files: `core/database/promotealign/evidence.go`.
Failing test first: none practical as a behavior test (output is unchanged). Add `TestEvidenceLoadsOnlyBridgedNeighborMembers` in `evidence_test.go`: a neighbor handle with two accepted members, only one reached by a filed bridge; assert the evidence is unchanged and (through an `export_test.go` hook or a counter on `evidenceReader`) that only the bridged member's Source was read. If a read-count hook is not possible without touching `graphcache`, assert instead that `Neighbor.Members` passed to `Build` holds only the bridged member, via a small exported-for-test seam in `export_test.go`.
Do: in the neighbor loop, compute `canon` (`q.canonLinks`) first; collect every `Bridge.Neighbor` id from those links; build the neighbor's member list from `acceptedMembers(H′)` filtered to that set, then read values only for those ids. `claimevidence.Build` does not change. Keep the row entity H's members whole (sections compare every member).
Check: the new test passes; all `TestEvidence*` pass unchanged; scripts/check-pr.sh PASS.

**F-2 — Read member values and bridge Observations once per Source** (from PRF-2, MNT outside-lens PRF)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/profiles.go` if needed.
Failing test first: none (performance only). Existing `TestEvidence*` and `TestBuild` must pass unchanged.
Do: in `members`, group the member ids by Source id (keeping claim order in the output), call `g.Source` once per Source and `layerValues(sg, <set of that Source's member ids>, …)` once per Source (D-8's set signature). In `bridges`, likewise group bridge Subjects by Source and walk each Source's Observations once, bucketing by `SubjectID`. Do not add a second SQL Observation read (D-8).
Check: `TestEvidence*`, `TestBuild` pass; scripts/check-pr.sh PASS.

**F-3 — One owner for "a layer bridge seen from one end"** (from MNT-1)
Files: `core/graphalign/types.go`, `core/graphalign/align.go`, `core/database/promotealign/evidence.go`.
Failing test first: `TestBridgeFrom` in `core/graphalign` (new): A end → (B, KeyA, true); B end → (A, KeyB, true); neither → false.
Do: add `func (b Bridge) From(id []byte) (other []byte, key vocabkey.BridgeKey, ok bool)`. `newState` builds `layerAdj` through it (one call per end) and the `Evidence` layer-link loop calls it instead of its `switch`.
Check: `rg -n "KeyA|KeyB" core/graphalign core/database/promotealign -g '!*_test.go'` prints only `types.go` (the struct and `From`); graphalign and promotealign tests pass; scripts/check-pr.sh PASS.

**F-4 — Read a bridge's ends from positive, unambiguous endpoint Observations** (from MNT-2)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/evidence_test.go`.
Failing test first: `TestEvidenceBridgeEndpoints` (new), subtests: a bridge whose neighbor end is recorded with a non-positive polarity gives no bundle; a bridge whose member end names two different Subjects gives no bundle; a positive single-valued bridge gives one bundle (control).
Do: in `bridges`, skip endpoint Observations whose `Polarity` is not `positive` (match `bridgeNeighbors`' SQL); an end that names two different Subjects is unresolved and the bridge is left out (as `identityclaims.resolveEndpoint`). State the rule in the doc comment.
Check: the test passes; scripts/check-pr.sh PASS.

**F-5 — Bound the neighbor list** (from SEC-1)
Files: `core/database/promotealign/evidence.go`, `evidence_test.go`.
Failing test first: a `TestEvidenceInvalidRequest` case with more neighbors than the Source has primary Subjects → `promote.ErrInvalid`.
Do: before the neighbor loop, reject `len(req.Neighbors) > len(in.primary)` with `promote.ErrInvalid`. Leave the per-entry validation and duplicate rule as they are.
Check: the case passes; scripts/check-pr.sh PASS.

**F-6 — `subjectcompare` says how a property counts under its lead** (from MNT-3)
Files: `core/subjectcompare/subjectcompare.go`, `subjectcompare_test.go`, `core/claimevidence/sections.go`.
Failing test first: `TestRulesRole` (new): a weighted lead → role lead; a pooled property → role pooled; a period end → role period end; a lifespan Span → not ok.
Do: add a role-returning method (e.g. `func (r Rules) Role(kind, key) (lead vocabkey.PropertyKey, role Role, ok bool)`, `Role` one of lead / pooled / period end) holding `Lead`'s walk; `Lead` calls it. `routeOf` reads the role instead of re-checking `Periods`.
Check: `TestRulesRole`, `TestSections*`, all `subjectcompare` tests pass; scripts/check-pr.sh PASS.

Minors riding along (optional, implementer's call, no tests): MNT-7 (keep the node from the first loop).

## Round 1 — fix
Head `2d808c8a`. Gate PASS (orchestrator rerun). F-1 to F-6 done; MNT-7 rode along. Deviations: F-1 used the planned fallback seam (`bridgedOnly`), since there is no read-count hook without touching `graphcache`. F-4's two-Subjects subtest passed before the fix; it guards the rule and the polarity subtest exercises the fix. F-3: `layer.go:199` still names `KeyA`/`KeyB`, but it constructs the `Bridge` literal and does not read it. F-6: `Role` checks pool before period end, as `Lead` does; no default-registry property is both.

## Round 2 — review
Reviewers: MNT, SEC+PRF (both had Fix-now blocker or major). CON not sent (clean in round 1). Scope `8262d21e...2d808c8a`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | (round 1) | Resolved | `bridgedOnly` trims neighbor members before values are read (`evidence.go:138-149, 211-221`) |
| PRF-2 | PRF | major | (round 1) | Resolved | `memberValues` and `bridges` read each Source once per batch |
| MNT-1 | MNT | major | (round 1) | Resolved | `graphalign.Bridge.From`, used by `newState` and `Evidence` |
| MNT-2 | MNT | major | (round 1) | Resolved | positive-only endpoints; an ambiguous end drops the bridge; `TestEvidenceBridgeEndpoints` |
| SEC-1 | SEC | minor | (round 1) | Resolved | bound at `evidence.go:50-52` (checked by orchestrator and SEC+PRF) |
| MNT-3 | MNT | minor | (round 1) | Resolved | `subjectcompare.Rules.Role`; `routeOf` reads `RolePeriodEnd` |
| MNT out-of-lens CON | CON | — | `newState` appends `(B, KeyA)` twice for a self-bridge `A == B` | Rejected | `connect` rejects `from == to` before writing (`connect.go`, `bytes.Equal(from.ID, to.ID)` → `ErrInvalid`), so no self-bridge reaches the layer |

Nits (logged): `MNT-9 · evidence.go:225-231 · members is a one-caller wrapper beside the inline neighbor path` · `MNT-10 · subjectcompare.go:191 · Role returns RoleLead with ok=false; consider RoleNone`

No round-2 blockers; no Fix plan 2. `origin/main` unchanged since the branch was cut (`9ca7472e`).

## Round 3 — user scope change (2026-10-21)
After the PR was ready, the user asked for the query and `Propose`'s shared setup to read through the graph cache in this PR, "not at the performance review at the end". The orchestrator mapped five gaps:
- the SQL walk of the layer;
- SQL in `checkFixed` and `mergeFixed`;
- SQL in `SourcesOf` and for titles;
- profiles rebuilt on every call;
- the layer rebuilt on every call.

These became plan §14 (D-17 to D-23, AC-4.x, T-17 to T-22, STEP-9 to STEP-14).
- PRF-3 moved from S11-21 back into this PR, as D-21. S11-21 now measures the memos.
- S11-25 gains the check that Apply reuses the memos.
- New S11-31 bounds the Source cache (eviction).

The user asked for this work, so review round 3 is user-authorized.

### Fix plan 3
Implement plan §14, STEP-9 to STEP-14, in order, with each step's check and the stop conditions listed there.

**F-7 — In-memory Source walk** (STEP-9, D-17). Failing test first: T-17 against the SQL.
**F-8 — Source title and existence in the cache** (STEP-10, D-19). Failing test first: T-21.
**F-9 — Handle checks and anchors through the cache** (STEP-11, D-18). Check: invalid-input tests unchanged.
**F-10 — Member Source ids in the cache** (STEP-12, D-20). Check: `rg -n "SourcesOf|sourcesOf" core` prints nothing.
**F-11 — Memos: per-Source layer and member values, profiles, value rules** (STEP-13, D-21 to D-23). Failing tests first: T-18, T-19, T-20.
**F-12 — Docs and Landed note** (STEP-14).
Check for all: scripts/check-pr.sh PASS.

## Round 3 — fix
Head `3ba99d2b`. Gate PASS (orchestrator rerun). F-7 to F-12 done. No stop condition fired.
- **Not reproduced:** T-18 and T-19 passed before the memo code, once STEP-9 to STEP-12 had made every graph read cached. They are kept as guards. T-20 drove the memo.
- **Deviations:**
  - D-23: there is no name-format drop, so `valuerules.Load` stays per call.
  - D-22: profiles are stored only through `profiler.keep()`, called by Evidence and `loadShortlist`. This keeps `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` unchanged.
  - AC-4.1 and AC-4.2 count graph queries only. These still run per call: the name-format read, `Propose`'s `MAX(revision)`, and `attachExhibits`' reads (S11-25 deletes those).
  - Deleted as dead: `sourceeventtitles.ForSource` with its SQL, and `canonicalgraph.TestWalkSource` (T-17 holds its cases).
  - `TestSourceEvidenceIsNotActivity` now expects evidence for `sources`, as D-19 requires.

## Round 3 — review
User-authorized. Reviewers: CON (Sonnet), MNT, SEC+PRF (`graphcache`, `effects`, `core/database/**`, the matching layers). Skipped: UX (no `macos/App/**`, strings, or error codes). Scope `2088073c...3ba99d2b`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (r3) | MNT | blocker | `checkFixed` scans `Vocabulary().Types` itself beside `typeByKey`/`typeKey` | Fix now → F-13 | verified: `promotealign.go:205-210` and `canon.go:230-252`; the two already differ on origin |
| MNT-2 (r3) | MNT | major | `catalog-queries.mdc` and S12-06's Ships still name the deleted `canonicalgraph.WalkSource` | Fix now → F-14 (rule); Later PR (S12-06) | verified: `.cursor/rules/catalog-queries.mdc:17`. The orchestrator wrote the in-memory walk into S12-06's Ships and Not-this-PR lines in `spike-12/deployment-plan.md` |
| MNT-3 (r3) | MNT | minor | `loadShortlist` copies `keep()`'s body; storing profiles depends on each caller | Fix now → F-15 | verified: `index.go:81`, `profiles.go:186-190` |
| MNT-4 (r3) | MNT | minor | Memo and match-index resets repeated at each drop site | Fix now → F-16 | same-pass cleanup; `commit.go:27-29`, `graph.go:173, 271-273` |
| PRF-1 (r3) | PRF | minor | Any `sources` row edit (cover, description) drops the Source graph and memos | Fix now → F-17 | verified: `registry.go:192` `Source: self()`; `onField` is the existing pattern; create and delete carry `title` in `Fields` |
| CON-1 (r3) (+ SEC+PRF outside-lens CON) | CON | minor | AC-4.1/4.2 say "no catalog query"; tests count graph queries only; name-format reads remain | Later PR (S11-21) | D-23's fallback leaves `valuerules.Load` per call, so §14's AC wording outran its own D-23. The AC is not reworded. S11-21's scope now counts the raw queries and names the fix if over budget |
| CON outside-lens PRF | PRF | — | An unknown Source id caches a `Found=false` graph | Rejected | small; dropped by a later create (`Source: self()` / `onField("title")` on create); eviction is S11-31 |

Nits (logged): `CON-2 · memo_test.go · merge subtest fakes the commit` · `MNT-5 · sourcefill/load.go:44,205 · loadTitleEdges takes subjects twice` · `MNT-6 · layer.go:429 · unchecked assertion` · `MNT-7 · export_test.go:69-90 · MemoSnapshotForTest re-derives keys` · `MNT-8 · source_walk_test.go:160 · rename TestWalkSourceMatchesSQL` · `MNT-9 · graphcache/source.go:225 · raw "positive"` · `MNT-10 · sourceeventtitles.go:88 · assemble has one caller`

### Fix plan 4

**F-13 — One type-key lookup** (from MNT-1 r3)
Files: `core/database/promotealign/promotealign.go`.
Failing test first: none (same output for every valid input; primary Subjects are provenencia types). Existing invalid-input tests stay green.
Do: `checkFixed` builds `typeByKey(g)` and compares `typeKey(types, n.SubjectTypeID)` with `ps.kind`; drop the inline loop over `v.Types`.
Check: `rg -n "range v.Types" core/database/promotealign` prints only `canon.go`; scripts/check-pr.sh PASS.

**F-14 — The catalog-queries rule names the in-memory walk** (from MNT-2 r3)
Files: `.cursor/rules/catalog-queries.mdc`.
Do: replace the `WalkSource` sentence with: "`canonicalgraph.Walk` stays for dependent discovery and for any walk that must see rows the cache has not loaded, including inside a write. One Source's Evidence graph is walked in memory with `graphcache.Graph.WalkSource` on its loaded `SourceGraph`."
Check: `rg -n "WalkSource" .cursor docs` names no `canonicalgraph.WalkSource` outside S11-19's own plan.

**F-15 — One writer of kept profiles** (from MNT-3 r3)
Files: `core/database/promotealign/index.go`, `profiles.go`.
Do: `loadShortlist` calls `p.keep()` instead of copying its body; `keep()` stores the map whenever the stored pattern differs or none is stored. `newProfiler`'s doc comment says a caller keeps profiles with `keep()`. Keep `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` unchanged.
Check: `rg -n "m.profiles, m.profilesPattern|profilesPattern =" core/database/promotealign` prints only `keep()`; `TestProfilesLiveWithTheMatchIndex` passes.

**F-16 — One drop for derived state** (from MNT-4 r3)
Files: `core/database/graphcache/graph.go`, `commit.go`.
Do: a private `dropDerived()` resets the memos and the match index; the vocabulary branch of `OnCommit`, `SetAliasTables` and `clear()` call it. Other `matchIndexOK = false` sites that drop only the match index (a handle commit) stay as they are.
Check: `TestSourceMemoDroppedWithSource` and graphcache tests pass.

**F-17 — Only a title change drops the Source graph** (from PRF-1 r3)
Files: `core/database/effects/registry.go`, `core/database/sourcefill` test.
Failing test first: add to `TestRetitleDropsSourceGraph` (or beside it) a description-only `sources.Update` and a cover change, each asserting the graph stays loaded (`LoadedSourceForTest` or the existing hook).
Do: `"sources"` → `Source: onField("title", self())`. Create and delete still carry `title` in `Fields`; keep a create-after-unknown-id case if one exists.
Check: the tests pass; `TestSourceEvidenceIsNotActivity` updated only if its expectation is per-field; scripts/check-pr.sh PASS.

Nits riding along (implementer's call): MNT-5, MNT-6, MNT-8.

## Round 4 — fix
Head `cbda6bf3`. Gate PASS (orchestrator rerun). F-13 to F-17 done, and nits MNT-5, MNT-6 and MNT-8 rode along. Deviations: F-17 tests a description-only edit and no cover change, because a real cover change needs an Artifact with a file. The plan's T-17 row still names the test by its old name, `TestWalkSourceMatchesSQL`.

## Round 4 — review
Reviewer: MNT (it had the round-3 blocker and major). CON and SEC+PRF were not sent back: CON had only a minor, deferred, and PRF-1 (r3) was a minor that the orchestrator read in `4f4acf4e` and marked **Resolved**. F-15 and F-16 were confirmed by MNT. Scope `b79f37c2...cbda6bf3`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (r3) | MNT | blocker | (round 3) | Resolved | `promotealign.go:201-205` uses `typeByKey`/`typeKey` |
| MNT-2 (r3) | MNT | major | (round 3) | Resolved | `catalog-queries.mdc:17`; S12-06 Ships |
| MNT-3, MNT-4, PRF-1 (r3) | — | minor | (round 3) | Resolved | `keep()` sole writer; `dropDerived()`; `onField("title", self())` |

Nits (logged): `MNT-11 · sourcefill/load_test.go:85-110 · shared Source catalog setup helper` · `MNT-12 · graphcache/graph.go:310 · LoadedSourceForTest has no nil guard` · `MNT-13 · load_test.go:82 · no cover-change case`

No round-4 blockers. `origin/main` unchanged since the branch was cut (`9ca7472e`).

## Result
Gate PASS · CI pending · open: none · follow-ups: CON-1 r3 (in S11-21 scope), eviction (new S11-31), Apply memo check (S11-25), Source walk per graph (S12-06) · to user: — · CON miss: none
