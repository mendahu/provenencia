# App-health audit 2 — 2026-10-28

**Moved to Spike 14 on 2026-10-29** when Spike 12 closed. PR IDs below are Spike 12's: open ones moved to Spike 14 (S14-10 – S14-23, the [map](deployment-plan.md#where-the-notes-went) maps them), and the rest were descoped to [`promote-ranking.md`](../../ideas/promote-ranking.md), [`share-packages.md`](../../ideas/share-packages.md) and [`structural-debt.md`](../../ideas/structural-debt.md).

The findings behind slices 5b and 6b of the [map](deployment-plan.md). An area audit of the whole app at `main` @ 3cc609c with [`review-pr`](../../../.agents/skills/review-pr/SKILL.md) ("Area audit"), all five lenses, with Spike 11's matching rebuild and Promote flow read closest. The [first audit](app-health-audit.md) (2026-10-27, `A-n`, `L-n`) was excluded; nothing here repeats it. PR order and status are in the map; the decisions taken on these findings are its working rules under "Audit 2 fixes".

**How to use this doc.** A scope entry names its items (`B-n`). `plan-pr` reads those items here for where, evidence and fix, and re-checks each against the code first: line numbers are from 3cc609c and move. Go findings marked *probe* were proven by a failing test, since deleted; Swift findings come from reading the code (no Xcode on the audit machine) and carry manual steps.

**Baseline on 3cc609c.** `go vet` clean; 91 of 92 Go packages pass. `promotealign` `TestLargeSourceBudgets` misses its warm and after-commit budgets on the audit machine, about four times slower than the M1 the budgets were set on, so that alone is not a finding (B-8 is).

**Verdict.** The layered design held: date-compare fuzzing (50,000 pairs), walk-order shuffles over random graphs with implied links, the `MaxPivotEnd` bound and kinship cycles, the S11-45 rung order, S11-46 index and lookup agreement, the numbers guard, and cache staleness under `PROVENENCIA_GRAPH_VERIFY=1` all came back clean. The defects cluster in the late name work (S11-28, S11-45, S11-46), in how graph alignment treats a neighbor that is decided or whose record another row holds, in costs measured on a fixture far smaller than the map's **Scale**, and in the Promote page's error and stale-row paths.

## Decisions on this audit

User, 2026-10-28. They are the map's working rules under "Audit 2 fixes".

- **Proposals are liberal; the researcher decides.** The audit rated a prefix pair of two full names (Joseph / Josephine, Jean / Jeanne) and a namesake inside the lifespan slack as defects. Both are accepted as intended: Promote surfaces a candidate on forgiving evidence, and what files is the researcher's choice. Dropped as findings: the prefix pair in matching and in the reconciler fold, an exact tie placed on the lowest ref, and the lifespan slack (reworded in Spike 11's bar, B-25).
- **A name that heads its own nickname set never folds into another set's head** unless `nicknameDirections` names the pair (option a of B-3). Correct folds that are not yet reviewed, such as Jack → John (`jack` heads `{jack, jackie}`), stop until they are added to that table; the page shows both names meanwhile. **Reversed 2026-11-01** (user, planning S14-18): when one person's records hold two names of a set, folding into the set's head is intended; B-3 is dropped and the map's working rule "A nickname folds into its set's formal name, as today" replaces this one.
- **B-14 is descoped** (2026-10-29). The pin policy for an opened sheet goes back to the user to rethink; no PR in slices 5b and 6b changes how pins are chosen.

## Scheduled

| Item | Title | Priority | Lens | PR |
| --- | --- | --- | --- | --- |
| B-1 | The reconciler panics on folds that do not chain | High | CON, SEC | S12-40 |
| B-2 | A decided or held neighbor erases a one-cardinality conflict | High | CON | S12-41 |
| B-3 | An alias fold keeps the set's head, even a nickname or another name | High | CON | S12-40 |
| B-4 | A name containing the key separators forges its match key | High | SEC | S12-42 |
| B-5 | A crafted alias table hangs project open | High | SEC, PRF | S12-42 |
| B-6 | Grandparent derivation multiplies duplicate links by children | High | SEC, PRF | S12-49 |
| B-7 | The search skill teaches FakeStore to rank like Go | High | MNT | Guidance fixed with this audit's docs; code in S12-50 |
| B-8 | Promote's cost grows with the catalog, not the Source | Medium | PRF | S12-43, S12-51 |
| B-9 | Evidence on a record with many member Sources thrashes the graph cache | Medium | PRF | S12-43 |
| B-10 | A join onto a place with many events reprojects every dependent | Medium | PRF | S12-47 |
| B-11 | The possible-duplicate check is quadratic per proposal | Medium | PRF | S12-51 |
| B-12 | Quick row decisions queue full proposals | Medium | PRF | S12-44 |
| B-13 | A Subject deleted or merged elsewhere dead-ends the Promote page | Medium | UX, CON | S12-44 |
| B-14 | Opening a row's sheet freezes its pins | Medium | UX, CON | Descoped (2026-10-29) |
| B-15 | A row set to New cannot open its sheet | Medium | UX | S12-45 |
| B-16 | The page says unplaced rows start as Skip | Medium | UX | S12-44 |
| B-17 | Promote errors appear away from Done and are not announced | Medium | UX | S12-44 |
| B-18 | The sheet's Status, Confidence and argument fields have no spoken name | Medium | UX | S12-45 |
| B-19 | Initials and nicknames are compared before accents are folded | Medium | CON | S12-46 |
| B-20 | Composed and decomposed text is not `Same` | Medium | CON | S12-46 |
| B-21 | Search cannot find a place by an æ, ø, ł or œ spelling | Medium | CON | S12-47 |
| B-22 | "Jr." and "Junior" are a generation conflict | Medium | CON | S12-46 |
| B-23 | Swift decides which connections file, and misses Go's cycle refusal | Medium | MNT, CON | S12-48 |
| B-24 | Promote bands and name part types are wire words with no guard row | Medium | MNT | S12-48 |

### B-1 — The reconciler panics on folds that do not chain

**Where.** `core/autoreconcile/pipeline.go` `foldValues` (about line 290) and `root`; the fold rule in `core/autoreconcile/names.go` `fold`.
**Problem.** `foldValues` gives a `total` map only to maximal values (ones that fold into nothing) and points each other value only at a maximal value it folds into directly. That assumes folding is transitive. Initial and `Contains` folds are; the alias folds (S11-28) and prefix folds (S11-45) are not. When v folds into u and u into w but v not into w, v keeps `into == nil` and `total == nil`, and `root(v).total[s] = true` writes to a nil map. In a cycle (Dick → Ricky → Rick → Dick in the product seed) no value is maximal. The same call runs in every write that changes a person's members (`writes.Run` → `RecomputeReported`), in `subjectprofile.Build` (Promote), and on open through `EnsureCatalog` → `Rebuild` whenever `CacheVersion` differs. S11-45 bumped `CacheVersion` to 20, so every existing project rebuilds on its next open, and the rebuild never commits: a project holding such a person is likely to crash on every launch (the app reopens the active project; read from `OnboardingModel.swift`, not run). No `recover` exists yet (A-28).
**Evidence.** *Probe* (`core/autoreconcile`, the test file's `aliasRules`): given names B., Bill, William on three Sources → `panic: assignment to entry in nil map`. *Probe* (product `ValueRules`): B./Bill/William, D./Dick/Richard, P./Peggy/Margaret, Mo/Molly/Mary, Caroline/Chick/Charles, Dick/Ricky/Rick and Ann/Anna/Hannah all panic; end to end, filing a third record of Ann / Anna / Hannah through promote and join panics in `writes.Run`. *Probe* (security): Joseph / José / Joseph Jose panics with the prefix rule on and not with it off; a 60 s fuzz of `Reconcile` found the panic from three one-part names. `TestReconcileNamesInvariants` runs under `matchregistry.Default().Value`, which has no alias tables, so it never generated an alias fold.
**Fix.** In `foldValues`, resolve each value through the fold graph to a maximal value it can reach (best supported, then best ranked); a value with no reachable target, or on a cycle, is its own root and gets a `total`. Run `TestReconcileNamesInvariants` with the product alias tables and a pool of nicknames, initials, prefixes and accented forms in one- to three-word units, asserting no panic and that every root has `into == nil`. The A-28 recover is not this fix: the open would still fail.

### B-2 — A decided or held neighbor erases a one-cardinality conflict

**Where.** `core/graphalign/align.go` `segments` (about lines 564 and 575: a neighbor in `st.held`, and an unplaced neighbor whose corresponding canon neighbor is in `st.owner`, are skipped with `continue`); `place` and `release` mark only the placed Subject's own layer neighbors dirty; `rescoreAll` (about line 491) re-scores without walking; `core/graphalign/rows.go` (an unplaced row takes `best.score`'s band).
**Problem.** A link with no result adds nothing, so on subject-of-birth or subject-of-death, the one-cardinality links, the −3 conflict vanishes and the event scores 2.5 on type and date alone, over the accept bar. Setting a person to New or Skip, or another row taking the canon neighbor's record, therefore turns a contested birth into an uncontested "agrees" on another person's birth. The researcher's own decision (this person is nobody in the tree) is the evidence that is dropped. Ownership changes are not dirty marks, so a score computed before a record was taken survives into the rows. This breaks Spike 11's dogfood bar 4 ("a row whose comparisons include a conflict shows contested, whatever its band") and the layer 6 rule that accepting what Go proposed moves no other row. It is not covered by "Proposals are liberal": the proposal is allowed, hiding its conflict is not.
**Evidence.** *Probe* (`core/graphalign`): Source has Ann Ames and her 1817 birth; tree has Mary Jones and her 1817 birth. Nothing decided → Ann's birth `target="" band=none score=-0.50 contested=true`; Ann decided New or Skip → `target="handle" handle=h-mary-birth band=weak score=2.50 reason="agrees" contested=false`. End to end in `promotealign`, `Propose` with Ann New, then `ResolvePins` and `SaveBatch`: Mary's birth record has two subject participation links, and the evidence sheet shows only the event type and date lines, both checked. *Probe*: tree Mary Ames —subject→ birth 1842; Source John Ames —subject→ birth 1842 and Mary Ames —subject→ death 1880; John's birth comes back unplaced at weak 2.50 (drafted by Swift) and, once Mary is decided on her proposed record, placed on Mary's birth. A random-graph probe (300 seeded graphs with kinship links, each proposed row fixed on its own record) broke "fixing the proposed record changes nothing" 42 times. `TestAlignInvariants` (c) runs on four hand fixtures, none with a held neighbor.
**Fix.** A held-New neighbor, or an unplaced neighbor whose corresponding canon neighbor another Subject holds, is someone other than that canon neighbor: score the link `Apart` on the first corresponding canon link, so a one-cardinality link subtracts `OneConflict`. Compare a held-Skip neighbor `Unplaced` on its values, as before the decision. When `place` or `release` changes `owner[h]`, re-score and re-queue the candidates whose links reach h. Add the fixtures to `TestHeldNewAndSkipDecisions` and `TestAlignInvariants`, a fixed-seed random invariant (c), and a `promotealign` test that pins "the other person's birth keeps one subject" through Done.

### B-3 — An alias fold keeps the set's head, even a nickname or another name

**Where.** `core/valuecompare/aliases.go` `AliasIndex.Expands` (about line 156); the seed in `core/database/namealiases/nicknames_seed.go`; the reviewed table `core/database/namealiases/nickname_directions.go`.
**Problem.** `Expands` keeps `into` over `from` when a set headed by `into` holds `from`; its reverse check fires only when a set headed by `from` also holds `into`. Upstream heads include nicknames and sets that join different names, and the fold ignores support, so a person's conclusion shows the wrong name from correct input: the researcher joined the right records. Seed sets behind the examples: `{chick, caroline, charlotte, chuck}`, `{ivan, john}`, `{elbert, albert, bert}`, `{hannah, anna, nan, nanny}`.
**Evidence.** *Probe* (`autoreconciler.ValueRules`, one Source per name): Charlotte, Charlotte, Chick → Chick; Albert ×3, Elbert → Elbert; Anna, Anna, Hannah → Hannah; John, John, Ivan → Ivan; Fred, Bill → Bill. An enumeration of the product index found 804 folds whose `from` heads its own set (`ann` → `antoinette`, `jack` → `jacqueline`, `bob` → `bert` among them).
**Fix.** Decided: option a. In `Expands`, a `from` that heads any set does not fold unless a Direction names (`into`, `from`); the reverse check reads "heads a set", not "heads a set that also holds `into`". Test rows for the four examples above and for Jack / John as the known cost. Re-pin the seed, bump `CacheVersion`, and update where the head is described: `docs/conclusion-reconciliation.md`, `docs/seeded-vocabulary.md` and `add-seeded-vocabulary`.

### B-4 — A name containing the key separators forges its match key

**Where.** `core/valuecompare/name.go` `nameKey` (about line 84; joins `role=words` with `|`), the `Same`-on-equal-keys branch and `sameParts`; `core/valuecompare/text.go` `positionalParts` (about line 142).
**Problem.** `NormalizeForm` keeps Unicode symbols and `namevalue.Validate` accepts them, so a one-part surname `Smith|given=John` keys `family=smith|given=john`, the same key as given John + surname Smith. `Compare` treats equal keys as equal word lists: with the longer list on side A it indexes past the shorter one and panics; the other way it scores `Same 1.00`. The shortlist files both under the same key, so they always meet. A typed value can do this, not only a crafted catalog.
**Evidence.** *Probe* (`core/valuecompare`): `Compare(John Smith, "Smith|given=John")` → `panic: runtime error: index out of range [1] with length 1`; reversed → `relation=same sim=1.00 detail=exact`. End to end, a canonical person filed with that name makes `promotealign.Propose` panic via `graphalign.Candidates` ← `loadCanon`. `claimevidence` collapses lines by `Key(Same)`, so the forged key also merges two names into one evidence line.
**Fix.** Build keys unambiguously (escape or length-prefix roles and tokens); on equal keys, compare the role and word lists themselves; `positionalParts` refuses unequal lengths. Add native fuzz targets for `Compare`, `Key` and `WordKeys` (none exists).

### B-5 — A crafted alias table hangs project open

**Where.** `core/valuecompare/shortlist.go` `wordKeys` (one `a:` key per set holding a word); `core/valuecompare/aliases.go` `SetKeys` and `Hit`; `core/database/namealiases/namealiases.go` `Load` (no bound); `core/catalogsession` `WarmOnOpen` → `promotealign.Warm` → `loadShortlist`.
**Problem.** `name_alias_*` rows come from the catalog, and nothing bounds how many sets one token is in. Every person whose given word sits in N sets is filed under N keys while the project opens, and `Hit` loops all N sets on every word-pair compare. S11-15's round-1 fix bounded a set's size, not the number of sets per token.
**Evidence.** *Probe*: 100 people named John, one table of 100,000 sets `{a<i>, john}` (about 3 MB of rows): `WarmOnOpen` 4.8 ms without the table, 18.75 s and 6,862 MB allocated with it (452 MB live). `WordKeys("John Smith")` returns 200,006 keys in 220 ms with 200,000 sets.
**Fix.** Validate alias tables on load, as `nameformats.LoadID` validates weights: registry bounds (`matchregistry`, with knob inventory rows) on sets per token and on sets and tokens per table; over a bound, leave the table out of the built index and let the product tables stand in, so the project still opens (returning an error, as `LoadID` does, would stop the open: map review K-33). A `promotealign` budget test with a hostile table.

### B-6 — Grandparent derivation multiplies duplicate links by children

**Where.** `core/kinship/kinship.go` `chains` (about lines 327–382: each implied link copies every duplicate up-step into `From` and dedupes every association); callers `core/database/canonlinks/reader.go` `relationshipEdges` (each filed link its own edge) and `core/kinship/links.go` `WithImplied`.
**Problem.** The pivot bound counts distinct parents of the middle person, not duplicate g → m links and not the middle person's children. With D duplicates and C children, derivation makes C links each carrying D `From` edges and D associations. S11-49's SEC-1 fix made the loop linear and was tested with one child.
**Evidence.** *Probe* (`core/kinship`), canon shape: 2,000 edges → 1,001,000 `From` edges, 1.0 s, 831 MB; 4,000 → 8.9 s, 3,833 MB; 8,000 → 57 s, 15,513 MB. Source shape (`LinkSet` merges duplicates into one edge): 4,001 edges → 16,004,000 associations, 4.7 s, 3,106 MB. `TestDeriveManyDuplicateParentSteps` uses one child, `TestDeriveManyAssociations` ten.
**Fix.** Carry each distinct (grandparent, middle) pair's `From` edges and associations once and share them across the children; bound the implied output per middle person through `kinship.Rules` (a registry value with an inventory row). A test with D duplicates × C children.

### B-7 — The search skill teaches FakeStore to rank like Go

**Where.** `.agents/skills/add-searchable-kind/SKILL.md` (checklist "FakeStore mirrors kinds + location rules + ref exact/prefix boost" and the Swift step "FakeStore scores in memory with the same kind/location rules"); `.cursor/rules/searchable-kinds.mdc` ("FakeStore mirrors Go kind rules"); the code they produced in `macos/App/Platform/FakeStore.swift` (`searchPriority`, `orderedByMixedScore`, `scoreRefRecord`, `scoreSearchFields`).
**Problem.** Both contradict the binding rule "FakeStore is a seeded fake, not a second engine" and `.cursor/rules/swift-tests.mdc`. The next search kind, or the kind-scoped picker S11-36/37 runs on, grows a second ranking engine nothing pins to Go. S12-35's Ships list does not name these functions.
**Fix.** The skill and the rule are fixed with this audit's docs: FakeStore returns seeded hits in seeded order for the requested kinds. S12-35 landed without touching search, so S12-50 removes the fake's ranking (the four functions above, `classifyRefQuery`, `tokenizeSearch`, the location rules in `vocabularySearchHits`) and the `SearchCatalogStoreTests` cases that test it.

### B-8 — Promote's cost grows with the catalog, not the Source

**Where.** `core/database/graphcache/commit.go` (`dropMatchIndex` on every handle commit); `core/database/promotealign/canon.go` `loadCandidatesMemo` (keyed on the index generation) and the canon walk (no depth bound); `core/database/promotealign/index.go` `loadShortlist`; `core/graphalign/align.go` `segments` (every corresponding canon link).
**Problem.** Any write that names a handle (every Promote Done, any Observation edit on a filed Subject anywhere, through `effects` `observationMembers`) drops the match index and, through its generation, every Source's ranked candidates. The next proposal re-ranks every Subject against everything its keys reach; the first proposal after open pays the same (S11-50 warms the index only); every warm re-proposal (each row decision) walks the whole connected canon from scratch. The S11-21 budget fixture is about a ninth of the map's **Scale**, so "inside budget" never tested it; the cost was deferred four times (S11-15 review M-8, S11-21, S11-46 PRF-2, S11-50).
**Evidence.** *Probe* on the S11-21 large fixture's Source A, audit machine: 4,000 persons → 16,000 persons (`shape.Persons=16000`, `TownEvents=8000`): warm at open 1.65 → 4.16 s; first proposal after open 1.88 → 4.28 s; warm re-proposal 376 → 661–779 ms; after a handle commit 1.82 → 4.05–4.19 s; after a promote and an edit on another Source 1.89 → 4.12 s. Shortlist per Subject, all fully scored by `graphalign.Candidates` before TopK: events average 886 → 1,418 (max 7,628), persons 337 → 1,203 (max 3,140). pprof after a commit at 16k: `loadCandidatesMemo` 2.91 of 3.82 s; warm at 16k: `loadCanon` 1.67 of 2.06 s. One-place Source, warm: 2.82 s against a 3 s budget, `segments` → `CompareNeighbor` 5.04 of 8.24 s over three runs.
**Fix.** Keep the shortlist index current on commit (`Index.Remove` and `Add` of the dirty ids, the generation stable); re-rank only Subjects whose lookup keys meet a changed handle's old or new keys ([`ideas/name-index-on-commit.md`](../../ideas/name-index-on-commit.md)); hold profiles in a `graphcache.NodeMemo` whose deps are the handle and its events; bound the walk ([`ideas/narrower-canon-walk.md`](../../ideas/narrower-canon-walk.md)). Budget subtests at the map's **Scale** and "after editing a filed Subject".

### B-9 — Evidence on a record with many member Sources thrashes the graph cache

**Where.** `core/database/promotealign/evidence.go` (about line 151, `q.members(row.HandleID)` reads every accepted member's graph; `memberValues`); `core/database/graphcache/evidence_graph.go` (`maxEvidenceGraphs = 64`, `evictEvidenceGraphs`); `core/database/promotealign/perf_test.go` (its place-row budget uses a city with one member).
**Problem.** One evidence request loads one Evidence graph per member Source. Past the cap, the scan cycles the LRU, so every open of that sheet reloads every member graph, and it evicts the proposing Source's own graph with its layer and candidates memos, so the next decision is a cold proposal though nothing was committed.
**Evidence.** *Probe*: a person with 70 members on Source A: first open 420 SELECTs and Source A's graph no longer loaded; re-propose with nothing changed 1.76 s then 1.53 s against 356 ms warm; second open 216 SELECTs again. A place with 100 member Sources: a repeat open issues 566 SELECTs (2 at 40 members), and `MemoSnapshotForTest(...).LayerOK=false`.
**Fix.** Hold member values per handle as a `graphcache.NodeMemo`, dropped when the handle commits (an Observation edit on a member already names the handle); read member graphs for a request without inserting them into the LRU, or re-touch the proposing graph after; a budget subtest with a 100-member place row.

### B-10 — A join onto a place with many events reprojects every dependent

**Where.** `core/database/autoreconciler/autoreconciler.go` `RecomputeReported` (about lines 185–207) → `searchindex.ReprojectHandles` → `conclusionheaders.RowsForSearch`, which reads through a one-shot `graphcache` (`loadChunk`, `loadLinks`).
**Problem.** Joining one place Subject to a city handle reprojects the search documents of the city and every dependent (its events and their people), recomposing each from SQL, even when the city's header text did not change.
**Evidence.** *Probe*: `f.join(pl, city)` on the large fixture's 6,000-event city: 6.06 s and 6.21 s (promoting a place onto a new handle: 5–8 ms). pprof: `RecomputeReported` 5.71 s → `ReprojectHandles` 5.58 s → `RowsForSearch` 4.41 s → `loadChunk` 4.09 s → `loadLinks` 3.43 s. `OnCommit` naming the city: 84–98 ms reloading its 6,000 links.
**Fix.** Reproject dependents only when a root's header-visible row changed (compare the root's header before and after the recompute; a join that only moves `support` changes nothing shown), and compose unchanged dependents from the warm graph. To confirm the reach: time `ApplyPromoteGraphAlignment` on the one-place Source and count the dependents.

### B-11 — The possible-duplicate check is quadratic per proposal

**Where.** `core/graphalign/rows.go` `wouldBeNew` (about line 126), `flagAlikeNewRows` (136), `alikeNew` (160, a full `st.evaluate`, not memoized).
**Problem.** Every proposal, each row decision's included, compares every pair of would-be-new rows of a kind with a full subject compare and neighbor segments. S11-47 made nearly every unplaced row qualify: its plan's §6 Scale said "most unplaced rows had no candidate", and on the fixture 27 of 30 unplaced rows have one.
**Evidence.** *Probe*, a Source of N new people and N births, warm proposal: N=50 236 ms, N=100 683 ms, N=200 2.29 s; pprof `flagAlikeNewRows` → `alikeNew` → `subjectcompare.Compare` 6.43 of 6.77 s.
**Fix.** Shortlist the would-be-new Subjects among themselves (`shortlist.NewIndex` over them, `Add` with index keys, `Lookup` with lookup keys) and compare only pairs that meet; memoize the symmetric pair; an all-new 100-Subject budget subtest.

### B-12 — Quick row decisions queue full proposals

**Where.** `macos/App/Features/Promote/PromoteModel.swift` `proposeAfterDecision` (about line 224, one `Task` per decision) and `repropose` (about 382, the generation checked only after the call); `isProposing` gates only Done (127); `macos/App/Platform/CoreInvoke.swift` (a `Task.detached` per FFI call); `core/catalogsession` `Do`.
**Problem.** Each decision sends a full proposal; Go runs them one after another under the session and Swift keeps the last. The same holds for `loadEvidence` when the researcher switches sheets.
**Evidence.** From the code path. Five quick picks on the one-place Source hold the session for five proposals (B-8's 2.8 s each on the audit machine), and every other catalog call waits behind them.
**Fix.** At most one proposal in flight; keep the latest pending fixed set and send it when the current call returns, dropping those between. Precedent: `CatalogSearchModel.scheduleSearch`.

### B-13 — A Subject deleted or merged elsewhere dead-ends the Promote page

**Where.** `PromoteModel.swift` `repropose` (about lines 382–411, sends `flow.fixed(for: flow.rows.filter(\.decided))` unfiltered) and `file` (413–441, only `promote.stale` re-proposes); `core/database/promotealign/promotealign.go` `checkFixed` (about 175–201: a Subject missing from `primary`, or a handle `nil` or merged, → `promote.ErrInvalid`); `api/ffi/handlers/promote_batch.go` (about line 55, `ResolvePins` runs before `SaveBatch`'s revision check); `core/database/promotealign/evidence.go` `mergeFixed`.
**Problem.** After another window deletes a Subject the page holds as a row, or merges a decided row's record, every re-proposal fails (the decided pair reaches `checkFixed`), every evidence sheet fails (each row's neighbors include that pair), and every Done fails with `promote.invalid` from `ResolvePins`, or with `promote.stale`, which re-proposes and fails again. The only text is "This subject can't be promoted", naming no row; the stale callout claims "Provenencia proposed the rows again. Rows you decided are kept" although that proposal failed; Done stays enabled (`canFinish` ignores the proposal error). S11-35's review raised it (SEC-out-1) and it was Rejected on the premise that the researcher changes the row; nothing on the page says which.
**Evidence.** Manual: open Promote, Map the rest, set "Mary Smith" to New Person; ⌘N, open the same graph, delete Mary's Subject; back in the first window pick any target on another row → the callout, every sheet reads "The evidence could not be loaded", every Done fails.
**Fix.** Keep `checkFixed` the one judge (map review K-30): Propose and the evidence read release a fixed pair it finds invalid (its Subject gone, its record gone or merged, its kind changed) instead of failing, and the proposal returns each released pair with a reason, typed on the wire with a guard row; Swift words a released row with the row-note pattern and keeps the other decisions. Run the revision check before `ResolvePins`; a `promote.invalid` at Done proposes again once, as stale does, clearing the stale note when that proposal fails. A Swift-side filter was considered and dropped: it reads a cached graph that can still list the deleted Subject, so it would give two behaviors for one case.

### B-14 — Opening a row's sheet freezes its pins

**Status.** Descoped by the user, 2026-10-29: how pins are chosen once a sheet has been opened is a bigger question than this item, and the user is rethinking it. The defect below stands as found; map review K-32's proposal (Go composes every row's pins at Apply from its defaults plus the researcher's overrides) is in the [map review log](deployment-plan.review.md) as input, not a decision.
**Where.** `macos/App/Features/Promote/PromoteFlow.swift` `setEvidence` (about lines 528–538: the first read takes `checkedPinIDs`, a later read only intersects with `offeredPinIDs`) and `batchRow` (589–597: `defaultPins` only while `board == nil`); `core/database/promotealign/pins.go` `ResolvePins`.
**Problem.** Once a row's sheet has been opened, Done sends that read's pin ids. A line or relationship Go would check by default that is offered only after a neighbor row is placed is never pinned, and starts unchecked when the sheet reopens. A row whose sheet was never opened gets it from Go's defaults, so the same page files different evidence depending on whether the researcher looked first. This bends the working rule "The pin policy has one owner, in Go".
**Evidence.** Manual: Ada suggested on PER-1, her spouse John unset; open and close Ada's sheet untouched; set John to PER-1's spouse; reopen: "…they share a spouse" is unchecked, though `core/claimevidence/build.go` checks an agreeing relationship; Done without reopening sends the old ids only.
**Fix.** An opened row the researcher never toggled keeps `defaultPins: true`; on a re-read, add the newly offered default-checked ids (`checkedPinIDs` minus the previous offers).

### B-15 — A row set to New cannot open its sheet

**Where.** `macos/App/Features/Promote/PromoteView.swift` (about lines 343–366: the only opener draws when `row.selectionReadout != nil`, which needs a handle target); `PromoteEvidenceSheet.swift` (about 67–68 and 109–114: the `.newKind` title and claim-field branches nothing reaches); `PromoteFlow.swift` `batchRow` (612–618 sends `confidenceGradeID` and `argument` for New).
**Problem.** Since S11-47 every New row is the researcher's choice, and each files with no confidence grade and no argument; no control sets them. S11-20 AC-1.4 and board S11-D2 (ES-19, E13c) draw the claim fields for New; the S11-20 ship log says its manual check was not run, and the model tests call `openSheet` directly.
**Evidence.** `openSheet` has one caller, `PromoteView.swift` about line 345. Manual: set any row to New Person; the assessment column stays empty and nothing opens the sheet.
**Fix.** A New row gets an opener in the assessment column (the S11-D1 rows board decides its look: brief S12-D8).

### B-16 — The page says unplaced rows start as Skip

**Where.** `macos/App/Platform/L10n.swift` `promote.page.openedFrom` (about line 1712: "Opened from %@ · weak and unmatched rows start as Skip"); `Localizable.xcstrings`; shown by `PromoteView.swift`.
**Problem.** Unplaced rows open unset with a red border and block Done (`draftTarget` never returns Skip). S11-47's Strings sweep looked for "New or Skip" phrasings, not "start as Skip".
**Fix.** Reword ("rows Promote can't place start unset") or delete the line and its key.

### B-17 — Promote errors appear away from Done and are not announced

**Where.** `PromoteView.swift` (about lines 86–105: `notices` inside the `ScrollView`; Done in the fixed footer at about 248); `PromoteModel.swift` (about 166–168: a load error has no retry; 405–410 and 435–441).
**Problem.** On a long mapped page the researcher presses Done at the bottom; the stale callout, a save error or a re-proposal error appears at the top of the scroll area, nothing scrolls to it, and nothing posts an announcement. A failed load has no Try again. Rows marked Updated after a re-proposal are not announced. A-18 covered only the Source page's graph writes.
**Evidence.** No `AccessibilityAnnouncement.post` under `Features/Promote`. Manual: map the rest of a 30-row Source, scroll to the bottom, edit any record in a second window, press Done: nothing changes near Done.
**Fix.** Announce stale, save and proposal errors through the kit (S12-30's `PVCallout` opt-in announcement); show the error by the footer or scroll to it; a Try again on a failed load that calls `load()`.

### B-18 — The sheet's Status, Confidence and argument fields have no spoken name

**Where.** `macos/App/Features/Promote/PromoteEvidenceSheet.swift` (about lines 120–138).
**Problem.** Neither `PVSelect` passes `accessibilityLabel`, so VoiceOver reads the trigger text ("Not stated, Not stated"); the `PVTextArea` has no label, and `PVField` does not tie its label to its content. Eleven of the fifteen `PVSelect` sites under `Features` pass one.
**Fix.** Pass the labels (`L10n.Promote.status`, `.confidence`) and label the text area; better, `PVField` labels its content in the kit, so every form gains it.
**Fixed in S14-14.**

### B-19 — Initials and nicknames are compared before accents are folded

**Where.** `core/valuecompare/words.go` (about line 294: the initial rung compares the first rune of the unfolded tokens, ahead of the fold rung; about 302: `Hit(x, y)` on unfolded tokens); `core/valuecompare/aliases.go` `AliasToken` (`TextKey` at `Same`); `core/valuecompare/shortlist.go` `wordKeys` (`a:` keys from the token as written); `core/autoreconcile/names.go` `fold` (`Expands` on the written text).
**Problem.** "E. Tremblay" is `Different` from "Émile Tremblay", and "É." from "E." although both share the `Equivalent` key: a Key-contract break. The product seed is ASCII ("helene"), so "Hélène" never meets "Lena", and S11-44's Des reaches "Désirée" only by prefix.
**Evidence.** *Probe*: `g=E.|s=Tremblay` / `g=Émile|s=Tremblay` → `different cause=no_match`; `g=É.` / `g=E.` → `different`, keys at Equivalent both `family=tremblay|given=e`; `Helene` / `Lena` → `equivalent 0.992 alias`, `Hélène` / `Lena` → `different`. A 20,000-pair Key-contract fuzz over accented pools found only this class; `TestKeyContract`'s name pools have no accented initial.
**Fix.** Check the folded tokens before the initial rung and compare the initial on them; try aliases on the folded tokens too (`Hit`, `SetKeys`, the reconciler's `Expands`), keeping "an accent fold alone never folds"; bump `CacheVersion` and `ProjectionVersion`; add the rows to `TestNameCompare`, `TestWordKeysShareOnSameOrEquivalent`, `TestReconcileNamesAliases` and "É." to `namePools`.

### B-20 — Composed and decomposed text is not `Same`

**Where.** `core/valuecompare/words.go` `wordRune` and `joinedKey`, `name.go` `nameKey` (no canonical composition at the `Same` level); Swift `TextMatch` decomposes only at `.equivalent`; nothing normalizes on write.
**Problem.** Decomposed input (common in text pasted from PDFs) compares `Equivalent` to the composed form, and the reconciler never groups `Equivalent`, so identical-looking values split. Compatibility forms (ligatures, full-width letters) are not even `Equivalent`, and a zero-width space makes a `Near`.
**Evidence.** *Probe* (`core/autoreconcile`): Montréal NFC vs NFD → `equivalent`; name → `given=José|given=José|surname=Silva`; text state `mixed`, two values. *Probe* (`core/valuecompare`): "Griﬃth" / "Griffith" → `different`; full-width "Ｍｏｎｔｒｅａｌ" / "Montreal" → `different`; "Mont​real" / "Montreal" → `near`.
**Fix.** Compose to NFC in the `Same` key, `nameKey` and `AliasToken`; at `Equivalent`, also fold compatibility forms and drop format characters (zero-width space, soft hyphen); mirror both in Swift `TextMatch` with fixtures in `TestSwiftTextMatchMirrorsTextKey`; bump `CacheVersion` and `ProjectionVersion`.

### B-21 — Search cannot find a place by an æ, ø, ł or œ spelling

**Where.** `core/database/searchindex/handles.go` `distinctAlternates` (about line 312, dedupes at `TextKey(…, Equivalent)`); the FTS tokenizer `unicode61 remove_diacritics 2` (`migrations/000018.sql`).
**Problem.** An alternate equivalent to the title is dropped from the document, but SQLite's FTS does not fold æ, ø, ł or œ the way Go's `Words.Folds` does, so neither the title nor the dropped alternate matches. S11-07 D-6 assumed retrieval folds like rescoring; its checks used Montréal, which `remove_diacritics` happens to handle. A regression: before S11-07 only byte-equal alternates were dropped.
**Evidence.** *Probe* (`core/search`), a place with two toponyms, the title first: "Aeroskobing" (title Ærøskøbing) → no hits; "Soby" (Søby), "Lask" (Łask), "Oeuilly" (Œuilly) → no hits; "Strasse" (Straße) → fuzzy only, score 5.23; control: alternate "Aeroskobing Havn" (not deduped) is found, score 35.55. The Places list filter (`TextMatch`) finds "Søby" from "Soby".
**Fix.** Index the `Equivalent`-folded form so FTS retrieval folds the way rescoring does (a folded column; a migration), or dedupe at a key that folds only what `remove_diacritics 2` folds; bump `ProjectionVersion`.

### B-22 — "Jr." and "Junior" are a generation conflict

**Where.** `core/valuecompare/name.go` step 3 (`sharesAny(ga, gb)` on raw generation tokens).
**Problem.** Generation tokens get no alias or normalization, and since S11-03 a mismatch is a hard `Different` with the name-conflict penalty (before, `SuffixConflict` scaled the score by 0.3).
**Evidence.** *Probe*, James Robins with each suffix: Jr. / Junior → `different generation_conflict`; Sr. / Senior and III / 3rd likewise. Reconciler: `given=James|surname=Robins|suffix=Jr.|suffix=Junior`.
**Fix.** One generation-token table (jr / junior / jnr, sr / senior / snr, ii / 2nd / second, iii / 3rd / third, …), seeded or in the registry with an inventory row, read by `nameKey` and the conflict check; bump `CacheVersion`.

### B-23 — Swift decides which connections file, and misses Go's cycle refusal

**Where.** `macos/App/Features/Promote/PromoteFlow.swift` `connections` (about 337–352), `resolve` (413–428, "The one place that decides"), `connectionCount` (310); Go `core/database/identityclaims/bridgefile.go` (about 154–157 and 222–228, `placeRelationshipWouldCycle`).
**Problem.** The rule "both ends are handles → files; the same handle → self-link refused" lives in Swift for the subrows and the footer count. Go's filer also refuses a `part_of` that would close a cycle, so Swift counts it as filing, Go leaves it unfiled at Done, and nothing says so. `promote-graph-alignment.md` §9.1 claims alignment flags it before Done; no Go code does. Every new refusal in the filer has to be copied into Swift.
**Fix.** Go reports per connection whether it will file and why not, computed from the page's fixed pairs through the filer's own predicates (self-link, cycle); Swift only words it. Correct §9.1.

### B-24 — Promote bands and name part types are wire words with no guard row

**Where.** `PromoteFlow.swift` `Assessment` (about lines 43–47, `rawValue ?? .none`, decoded from `String` fields in `GenealogyStore.swift`); `macos/App/Features/Names/NamePartType.swift` (mirrors `namevalue.PartTypes()`, `isAllowed` re-validates); `api/ffi/swift_vocabulary_test.go` (no row for either).
**Problem.** `.cursor/rules/macos-client.mdc` says every such enum is a guard row, in `Platform/`. A renamed band silently reads `.none`, and S11-29's `former_surname` part type will be rejected in the name editor. S11-47 D-13 typed the row target but not `assessment`; `add-name-value` step 6 omits the match-registry role, the format seed roles and the Swift case.
**Fix.** Move both enums to `Platform/`, type the `assessment` fields, decode once in `GoStore`, add the two guard rows; extend `add-name-value` step 6 with the three missing places.

## Not scheduled

Low items. A plan pulls one in only with the user's OK.

- **B-25.** Lifespan slack: a person who died up to 12 years before the Subject was born is proposed (`subjectprofile.go` window ± `SlackYears` 10, plus `Dates.Tolerance` 2). Accepted by "Proposals are liberal"; Spike 11's dogfood bar 2 and `comparison-layers.md`'s example are reworded with this audit's docs.
- **B-26.** Swift drafts a weak unplaced row onto its first alternative (`PromoteFlow.swift` `draftTarget`), a branch keyed to how `AcceptBar` relates to `Bands.Weak`, which only the registry knows; dormant with the shipped values except through B-2's stale scores. Let Go send the draft, or drop the branch.
- **B-27.** The sheet words a prefix pair of two full names "Matches by a shortened name" (`promote.sheet.line.containsPrefix`); liberal proposals lean on honest wording.
- **B-28.** Promote takes four registry sections through `graphalign.Config` and reads Kinship, Evidence and `Subject.Pools` from `matchregistry.Default()` inline (`promotealign` canon, layer, evidence, values); the layer memo is not keyed on the kinship rules.
- **B-29.** `TestProductAliasTablesPinCacheVersion` pins only the alias seed; the registry values the reconciler cache reads (folds, particles, `PrefixMinLength`, …) and the `Words` rules behind `ProjectionVersion` are not pinned.
- **B-30.** `.cursor/rules/match-rules.mdc` globs miss `core/autoreconcile`, `core/database/namealiases`, `core/database/promotealign`, `core/database/canonlinks` and `core/database/valuerules`.
- **B-31.** The name word bound runs after splitting, keying and sorting, and `w:` and `a:` shortlist keys are not bounded (a 100,000-word name: `Compare` 792 ms, `WordKeys` 827 ms).
- **B-32.** A found record reads "Little evidence · Your choice" before Promote scores it, and keeps it if the re-proposal fails (`PromoteFlow.hold`).
- **B-33.** An unset row with a candidate looks like one with none, and the target menu shows no band; with liberal proposals, offered names are easy to miss.
- **B-34.** VoiceOver line labels omit the outcome (`spokenLine`), so "Matches by initial" and "Matches by a shortened name" sound like agreement.
- **B-35.** The sheet goes blank when a re-proposal moves the open row's record (`PromoteFlow.merge` drops the board without reloading).
- **B-36.** A property Card is headed "they have a similarity in date" over lines that all conflict (`promote.sheet.similarityIn`); the user's call.
- **B-37.** Unread wire fields under the retired word "score" (`Alternative.score`, `Row.score`, `via_bridge_type`, `via_role`) and `graphalign.Subject.Provenance`, a multiplier that is always 1.
- **B-38.** Docs made false by S11-14, S11-16 and S11-47: the `graphalign` package doc and the `ProposePromoteGraphAlignment` proto comment ("a handle, New, or Skip"), `promote-graph-alignment.md` §3 and §12, `promote-find-another-record.md` (a row with no candidate "opens New"), `comparison-layers.md` (the stats, the weak-band gate), the shortlist `Keys` doc, Spike 11's map row S11-30 and `completed.md` statuses, the `alignment.MinSimilarity` Start column.
- **B-39.** `namealiases.AddProductNicknames` has no caller (S11-44 ships no command).

## Gate gaps

[`review-pr`](../../../.agents/skills/review-pr/SKILL.md) "Close the loop": where each item came from and the gate that should have caught it. The guidance changes ship with this doc.

| Class | Items | Gate and why it missed | Guidance added |
| --- | --- | --- | --- |
| A new kind of input breaks an assumption old code made | B-1 | S11-28 and S11-45 reviews; the fold's transitivity was never listed, and the invariant test ran without the shipped tables | `plan-pr` research "Find the edges" and template §6 **Edges**; `review-plan` Plan checks "Assumptions"; conformance "Assumptions a new case breaks", "Production configuration in tests", known trap "A new kind of case in an old loop" |
| A rule a table drives, reviewed from its examples | B-3 | S11-28 D-1a weighed two-way pairs only; nobody ran the fold over the seed | §6 **Edges** "Real data"; Plan checks "Real data"; conformance "A rule a table drives", known trap "A table nobody read" |
| Removing an input removes the evidence it carried; a result kept past a change it depends on | B-2 | S11-16 D-6; held and owned neighbors on a one-cardinality link were not asked about, and owner changes were not dirty marks | §6 **Edges** "Removed inputs"; Plan checks "Removed inputs"; conformance "Excluded is not neutral", "Results that depend on changing state", known trap "A decision that drops evidence"; `plan-pr` §5 decided or deleted inputs |
| A claim over all inputs, tested on hand fixtures | B-2, B-19 | `TestAlignInvariants` and `TestKeyContract` pools | Conformance "Claims over all inputs"; `review-pr` evidence rules (generated inputs) |
| A widened rule whose nearest cases nobody listed | B-19, decision on the prefix rule | S11-45's tests and reviews covered its examples only; the user accepted the neighbors once they were listed | §6 **Edges** "Neighbors"; Plan checks "Neighbors"; conformance "Neighbors of a widened rule" |
| Keys built by joining values | B-4 | S11-03 X-3 read the key's content, not its separators | Plan checks "Untrusted size"; security "Keys built by joining values", "Panics from crafted data", "Fuzz the comparers too", known trap "A separator in the value" |
| A bound on one dimension of an untrusted collection | B-5, B-6, B-31 | S11-15 bounded set size; S11-49 tested one child | §6 **Edges** "Untrusted size"; security "Every dimension of an untrusted collection", "Bounds before the work", known trap "One axis bounded"; `ship-pr` routing by what code does, and the matching packages added to its paths |
| A budget met below the target scale; a cap smaller than one request | B-8, B-9 | S11-21 "inside budget" on a ninth-size fixture, then four deferrals; S11-31's cap never tested past its size | Template §6 **Scale** comment; `plan-pr` gate; Plan checks "Scale" and the Floors; performance step 6 "Measure growth", look-fors and known traps; `ship-pr` "A deferral is not a measurement" |
| A plan's scale claim never counted | B-11 | S11-47 §6 Scale | Plan checks "Scale"; performance known trap "A plan's 'most rows are cheap' never counted" |
| A finding rejected on an untested premise | B-13 | S11-35 SEC-out-1 | `ship-pr` Rejected needs evidence for its premise; `review-plan` Floors |
| A state no control reaches | B-15 | S11-20; its manual check was not run | UX "Every state has a way in", known trap "A sheet branch with no opener" |
| Copy that describes a behavior the PR changed | B-16 | S11-47's sweep grepped one phrasing | UX "Copy that describes a behavior", known trap "A header that kept the old default" |
| An error shown away from the action | B-17 | Pre-Spike 11; never asked | UX "Errors where the researcher acted" |
| An id held across a round trip | B-13 | Never asked | UX "An id held across a round trip", known trap "A deleted item that blocks the page"; executability "Edges" |
| One equivalence, two layers that disagree | B-20, B-21 | S11-07 and S11-08 checked one easy case | Conformance "One equivalence, two implementations" |
| A client rule keyed to engine values; a wire vocabulary with no guard | B-23, B-24, B-26 | S11-27 "state from its ends"; S11-47 kept the weak draft by an Out line | Maintainability known traps "A client decision keyed to engine values", "A string vocabulary that crosses the client boundary" |
| A skill teaching against a binding rule | B-7 | S12-20 fixed the same skill without searching it for the opposite instruction | Maintainability known trap "A skill or rule line that contradicts a binding rule" |
