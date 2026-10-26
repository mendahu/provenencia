# S11-49 — Kinship links in the graph cache

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine read with no caller yet; the one Swift change removes a seeded term's wording)
**Board:** —
**Depends on:** S11-38 (landed [#564](https://github.com/mendahu/provenencia/pull/564))
**Blocks:** S11-39, S11-40
**Branch:** `s11-49-kinship-graph-cache`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape), [§ Where implied links live](../kinship-shapes.md#where-implied-links-live); working rules: "Three patterns, product terms only", "The subject is an event's principal", "The canon's implied links live in the graph cache", "Bounded per pivot", "High tolerance for churn"; [#567's ship log](https://github.com/mendahu/provenencia/blob/s11-39-matching-reads-kinship/docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.ship.md) (measurements, the user's decisions)
**Planned:** 2026-10-26 against `main` @ `13dab35f`
**Review:** [`S11-49-kinship-graph-cache.review.md`](S11-49-kinship-graph-cache.review.md) — Fits with changes / Ready with changes, 2 rounds (every finding answered)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The open catalog's graph cache answers "which kinship links do the patterns imply for this person?" from one cached result per pivot (a marriage or birth event, or the parent between in a chain), dropped when a commit touches what it read. Matching (S11-39), claim evidence (S11-40) and later trees and person pages read it instead of each deriving over the links they walked. The patterns also gain the user's three 2026-10-26 decisions that so far live only on #567's branch: a marriage's couple are its two `subject`s and the `spouse` participation role leaves the product vocabulary, a pivot with more than 10 participants at either end implies nothing, and the grandparent derivation stays linear.

**Why now.** S11-39's first ship loop ([#567](https://github.com/mendahu/provenencia/pull/567)) derived over the whole walked component on every proposal (about 20–30 ms on a 12-generation tree), then memoized the walk per Source, which missed on every pin, was thrown away by every commit and kept one catalog-sized copy per Source (up to 64). The user's call: "putting the relationship shortcuts into the cache is smart".

**What stays.** No reader changes behavior: `loadCanon`, the layer loader and `canonLinks` keep what they do today (S11-39, S11-40 move them onto the read). Weights, alignment and the evidence sheet are unchanged. Nothing is written to the catalog.

## 2. User stories

**US-1.** As the engine, when a caller asks for one person's implied links, it gets every spouse, parent and grandparent link the three patterns imply with that person as an end, from cached per-pivot results, so callers never derive.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Canon: marriage M1 with Ann and Bob each `subject`. `Reader.For(Ann)` | One link: `relationship·spouse` Ann–Bob (lesser id first), `Pattern` `marriage_spouses`, `Event` M1, `Associations` both participations |
| AC-1.2 | Canon: birth B1 with Ann `father` and Bea `subject`. `For(Bea)` | One link: `relationship·parent`, Ann parent of Bea, `Pattern` `birth_parents`, `Event` B1 |
| AC-1.3 | Canon: filed `relationship·parent` Ann→Bea and Bea→Cal. `For(Ann)` and `For(Cal)` | Each holds one `relationship·grandparent` Ann→Cal, `Pattern` `parent_of_parent`, `Through` Bea, `Associations` both relationships |
| AC-1.4 | Canon: births B1 (Ann `father`, Bea `subject`) and B2 (Bea `mother`, Cal `subject`). `For(Cal)` | Holds the parent link Bea→Cal (B2) and the grandparent Ann→Cal through Bea, whose `Associations` are the four participations |
| AC-1.5 | Canon: marriage M2 with 11 `subject`s; M3 with 10. `For` on a subject of each | M2 gives no link; M3 gives 9 spouse links for that subject |
| AC-1.6 | Canon: marriage M1 where Ann is `subject` and Bob is `witness`; a catalog-own event type `wedding` with two subjects | No link from either |
| AC-1.7 | `For` on a person with no event or relationship links | Empty, no error |

**US-2.** As the engine, after a commit the cached links are what a fresh derivation over the committed catalog gives.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1, then Bob's participation in M1 is removed and committed | `For(Ann)` is empty |
| AC-2.2 | AC-1.1, then M1's `event_type` changes to `residence` and is committed | `For(Ann)` is empty |
| AC-2.3 | AC-1.3 with only Ann→Bea filed; then Bea→Cal is filed and committed | `For(Ann)` gains the grandparent Ann→Cal |
| AC-2.4 | AC-1.4, then B1's `father` participation is removed and committed | `For(Cal)` keeps the parent link from B2 and loses the grandparent |
| AC-2.5 | Any cached result, then a vocabulary commit | Every kinship result is dropped; the next `For` derives again |
| AC-2.6 | Each AC above run with `PROVENENCIA_GRAPH_VERIFY=1` | Passes: Verify finds every cached result equal to a fresh derivation |
| AC-2.7 | AC-1.1 derived without Bob loaded; then Bob is loaded by `g.Node(Bob)` (no commit) | M1's cached result stays (a first load is not a change) |

**US-3.** As a researcher, a marriage I record with the couple as its subjects reads as a marriage, and a new project no longer offers `Spouse` as an event participation role.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `kinship.Derive` over two `subject` participations in one `marriage` | One spouse link |
| AC-3.2 | A new catalog's seeded `role` terms | No `spouse`; `subject`, `father`, `mother`, `child`, `witness`, `informant`, `pallbearer` remain; `relationship_type` `spouse` remains |

## 3. Scope

| In | Covered by |
| --- | --- |
| The bound: `kinship.Rules{MaxPivotEnd}`, `Derive(edges, terms, rules)`, `matchregistry` value 10 and its knob row | STEP-1, AC-1.5 |
| Grandparent associations deduped through a set (linear) | STEP-2, T-2 |
| Marriage pattern reads `subject`; `spouse` participation role leaves the seed, `vocabkey`, test terms, test literals and Swift | STEP-3, AC-3.1, AC-3.2 |
| Shared vocabulary lookup (`Vocabulary.PropertiesByID`) and the kept-value read (`Node.Kept`), with their copies moved onto them | STEP-4 |
| `graphcache` registered per-node memo slots, dropped on commit through the owner's deps function; the dead `details` memo removed; `graphcachetest` canon builder | STEP-5, AC-2.5, AC-2.7 |
| `core/database/canonlinks` *(new)*: `LoadTerms` and `EdgeOf` moved from `promotealign` | STEP-6 |
| `Reader.For` over cached pivots | STEP-7, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7 |
| Commit, oracle and verify tests | STEP-8, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.6 |
| Docs: kinship-shapes, kinship-view, matching-model, add-match-rule, completed, map | STEP-9 |

| Out | Home |
| --- | --- |
| Any reader of `Reader.For` (`loadCanon`, `canonLinks`); `promotealign`'s `typeKey` scan | S11-39, S11-40 |
| The Source side's derivation (per layer) | S11-39 |
| `kinship.Shape`, `CountedOnce`, graph alignment's count-once rule | S11-39 |
| Moving `displays` onto the memo slots | not planned in this PR — its drop runs at install and through the header deps walk (`HeaderDepsWalked`); the slot types already fit it (D-5) |
| Comparing `Reason` with `autoreconcile.ReasonKept` instead of the literal `"kept"` | not planned — logged in the review |
| Warming pivots at open | not planned — derived on first read (about 20–30 ms per catalog); S11-50 warms kinds and the name index |
| Existing catalogs' `spouse` role term (kept; implies nothing) | not planned — dev phase, no migration or backfill (user, 2026-10-26) |
| Patterns beyond the three; link reading in its own pure package | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Moving other tests' direct canonical inserts (`graphprogress`, `conclusionheaders`, `core/database/promotealign/bench_test.go`) onto the builder | not planned — their shapes differ (identity claims, bulk inserts) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Derivation | `core/kinship/kinship.go` `Derive(edges []Edge, terms segmentcompare.Terms) []Implied` | Event patterns, then chains over direct and implied parent steps. No bound: `eventLinks` pairs every participant (quadratic per event); `chains` dedupes associations with `slices.ContainsFunc` (quadratic per triple). Package doc says "no numbers". Only `core/kinship` tests call it |
| Patterns | `core/kinship/patterns.go` `eventPatterns` | Marriage reads `vocabkey.RoleSpouse` at both ends |
| Role constant | `core/vocabkey/vocabkey.go` `RoleSpouse` (line 97) | Read by `patterns.go` and `core/kinship/kinship_test.go` (lines 57, 135, 202) |
| Seed | `core/database/subjectvocab/registry.go` | `{PropertyKey: "role", Key: "spouse", Label: "Spouse"}` (line 240); `relationship_type` `spouse` (line 246) stays |
| Seed check | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | Lists `vocabkey.RoleSpouse` |
| Test terms and literals | `core/segmentcompare/segmentcomparetest/terms.go` (participation `"spouse"`, line 32); `core/subjectprofile/invariants_test.go:153` (`"spouse"` in a role list); `core/subjectprofile/lifespan_test.go:68` (`vocabkey.Role("spouse")` on a marriage) | Use the participation role. Every other `"spouse"` in `core` tests (`promotealign`, `graphalign`, `segmentcompare`, `claimevidence`, `promote`, `connect`, `subjectvocab_test.go:343`) is the `relationship_type` term and stays |
| Swift | `macos/App/Platform/ConclusionVocabulary.swift` `SeededRoleTerm.spouse` (line 304); `macos/App/Platform/L10n.swift` `roleSpouse` and its `wording` case (lines 2623, 2896); `macos/App/Resources/Localizable.xcstrings` `propertyTerm.role.spouse`, `evidenceGraph.share.role.spouse`, `evidenceGraph.share.role.spouse.similar` | Wording for the seeded role; `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` keeps Swift equal to the Go seed. Relationship-term spouse strings stay |
| Registry | `core/matchregistry/registry.go` `Registry`, `Default()` | No kinship section |
| Knob inventory | `docs/matching-model.md` | No kinship paragraph, glossary row or knob row; `TestKnobInventoryMatchesRegistry` checks the knob table against the registry |
| Graph cache memos | `core/database/graphcache/graph.go` `Graph.details`; `core/database/graphcache/display.go` `Detail`, `SetDetail`, `SetDetailCheck`, `detailFn`; `core/database/graphcache/verify.go` `Verify` (the `g.details` loop, line 68) | `details` has no caller outside `graphcache` (`grep -rn 'SetDetail(\|\.Detail(' core api`): dead |
| Commit drops | `core/database/graphcache/commit.go` `OnCommit` | Collects `roots` (committed ids plus `detachAssociation`'s old and new endpoints), reloads, then `dropCommittedDisplays(roots, set)`; owners say what else to drop through `SetDisplayDeps` |
| First loads | `core/database/graphcache/load.go` `loadChunk` → `install` | `install` runs on every load, first or replacing |
| Vocabulary maps | `core/database/graphcache/vocab.go` `Vocabulary.TermsByID`; `core/database/promotealign/profiles.go` `propertiesByID`; `core/database/conclusiondetails/conclusiondetails.go` `loadProperties` (inline props map, line 263) | Two hand-built props-by-id maps; no types-by-id |
| Kept value | `core/database/conclusionheaders/compose.go` `kept`, `rank1`; `core/database/promotealign/canon.go` `keptTerm` | Both read the rank-1 `kept` row of a property |
| Link → key | `core/database/promotealign/canon.go` `canonKeys` (callers `canon.go:155`, `evidence.go:396`; `selfKind` at `canon.go:142` and `fromKind` near `evidence.go:390` exist only to pass to it) | Reads a node's link as an edge with its two keys |
| Terms | `core/database/promotealign/terms.go` `loadTerms` (callers `promotealign.go:108`, `export_test.go` `LayerBridges`) | Builds `segmentcompare.Terms` from the vocabulary |
| Test catalog | `core/database/graphcache/graph_test.go` `openCatalog`, `TestCommitOfAnAssociationLoadsEndpoints` (inline `typeID`/`propID` closures and direct inserts) | Private to `graphcache_test` |

**Precedent.** #567's branch `s11-39-matching-reads-kinship` holds three commits to port, in this order, not to merge:
- `16ff9ef2` — the bound. Port its `core/kinship/kinship.go`, `core/kinship/kinship_test.go` (`TestDeriveBound`) and `core/matchregistry/registry.go` hunks, `core/matchregistry/kinship.go` *(new)*, and from `docs/matching-model.md` only the knob row `kinship.MaxPivotEnd` (it goes after the `alignment.Shortlist.Dated.event.Date` row, before `evidence.Checked`). Not its `promotealign` hunks.
- `9c29cbaf` — the set dedupe in `chains` and `TestDeriveManyAssociations` (as is).
- `c14e0d3c` — the role. Port its `patterns.go`, `vocabkey.go`, `subjectvocab`, `segmentcomparetest`, `kinship_test.go`, Swift and xcstrings hunks. Not its `docs`, `promotealign` or `graphalign` hunks (they edit #567's own tests).

The registered memo copies how `conclusionheaders` registers `SetDisplayDeps` and how `graphfill` registers `SetEvidenceGraphCheck` in `init`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `kinship.Rules{MaxPivotEnd int}` *(new)* and `Derive(edges, terms, rules)`; `matchregistry.Registry.Kinship kinship.Rules` *(new)*, default 10 in `core/matchregistry/kinship.go` *(new)* with the `decided:` comment from `16ff9ef2`. An event group with more than `MaxPivotEnd` at either end, and a middle person with more than `MaxPivotEnd` distinct parents, imply nothing | User, 2026-10-26: a bound, more than 2, a registry value passed in; the purity guard bans literals in `core/kinship` | A constant in `core/kinship` — fails `TestLayerPackages` |
| D-2 | `chains` dedupes a triple's associations through a `map[string]struct{}`, first-added order (`9c29cbaf`) | #567 round 2 SEC-3: 16,000 associations took 4.1 s | — |
| D-3 | The marriage pattern reads `vocabkey.RoleSubject` at both ends. `vocabkey.RoleSpouse`, the `role` seed `spouse`, its `segmentcomparetest` term, `SeededRoleTerm.spouse`, `L10n` `roleSpouse` and its three xcstrings keys are removed; `invariants_test.go:153` uses `"witness"` and `lifespan_test.go:68` uses `vocabkey.RoleSubject`. `PatternMarriageSpouses`, `TermSpouse` and every `relationship_type` spouse stay | User, 2026-10-26: "We should align around the subject being the principal role of an event" and "remove the spouse role from event participation" | Reading both roles — the user removed `spouse` |
| D-4 | One cached result **per pivot**, computed on first read. Each edge handed to `Derive` carries `Associations: [][]byte{l.Association}` of its link. An **event** result is `Derive` over the event node's participation links. A **chain** result (person m) is `Derive` over m's `relationship` links plus the `birth_parents` links with m as an end from m's event neighbors' event results; `Derive` reads only the parent links among them. Neighbors in `Node.Links` are already unmerged (the link query in `load.go`) | Each pattern reads one node's links, plus one hop for a chain. Computing on read is correct whatever is loaded and costs nothing for nodes never asked | Deriving at install and filing by end — install also runs on first loads, and a chain's inputs may not be loaded yet |
| D-5 | `graphcache` gains registered per-node memo slots: `type NodeMemo struct { Name string; Check MemoCheck; Deps MemoDeps; DropOnVocabulary bool }` *(new)*, `MemoCheck` (today's `DisplayCheck`, renamed; `SetDisplayCheck` and the slots both take it), `type MemoDeps func(g *Graph, roots [][]byte) ([][]byte, error)` *(new)* (the shape `conclusionheaders`' deps fits through a one-line wrapper when `displays` moves later), `type MemoSlot int` *(new)*, `func RegisterNodeMemo(m NodeMemo) MemoSlot` *(new)* (package-level, called from an owner's `init`), `(g *Graph) Memo(slot MemoSlot, id []byte) (any, bool)` *(new)*, `(g *Graph) SetMemo(slot MemoSlot, id []byte, v any)` *(new)*, `(g *Graph) Cached(id []byte) (*Node, bool)` *(new)* (no load), which replaces `LoadedForTest` (its six uses in `graph_test.go` move to it). `OnCommit`, after the reload and beside `dropCommittedDisplays`, drops each slot's entries for `roots` and for `Deps(g, roots)`; its `ListVocabulary` branch empties slots with `DropOnVocabulary`; `clear` empties all; `Verify` runs each slot's `Check` over its entries and skips a slot whose `Check` is nil. `Memo` and `SetMemo` follow `Display`'s guards (nil graph, an id that is not 16 bytes). `install` and `removeNode` do not drop memos. `details`, `Detail`, `SetDetail`, `SetDetailCheck` and `detailFn` are deleted | One mechanism decides what a write drops, and the owner says how far (round 1 M-1); the next per-node memo is one registration (M-2); a first load keeps results (AC-2.7); kinship does not read the alias index or name format, so `dropDerived` leaves it (M-7) | Dropping in `install` — first loads would throw results away and graphcache would hard-code one hop; a third hand-rolled `map[string]any` |
| D-6 | New package `core/database/canonlinks` owns canon link reading and the implied-links read: `LoadTerms`, `EdgeOf`, `Reader`, `NewReader`, `For`. It registers its slot into the unexported `kinshipSlot graphcache.MemoSlot` (`Name: "canonlinks.kinship"`, `DropOnVocabulary: true`) in `init`; its tests are internal (`package canonlinks`) so they reach `kinshipSlot`; its `Deps` returns every `Link.Neighbor` of each root's cached node, plus the root | The walk and the evidence query read filed links through `EdgeOf` too, so the name says canon links, not only implied ones (M-5); one hop is the owner's rule, so a two-hop pattern later changes only this package | `impliedlinks` — the walk would call an implied-links package for direct links; inside `promotealign` — S11-40 and later readers are not matching |
| D-7 | `promotealign.loadTerms` moves to `canonlinks.LoadTerms(g *graphcache.Graph) (segmentcompare.Terms, error)`; `canonKeys` moves to `canonlinks.EdgeOf` (signature below), dropping the `selfKind` parameter, so `selfKind` (`canon.go:142`) and `fromKind` (`evidence.go`, the line before `canonKeys`) are deleted; check `types` is still used. `promotealign.keptTerm` is deleted: `EdgeOf` takes the `event_type` property id (`eventTypeProp`) in place of the props map and reads the event type as the `TermKey` of `graphcache.Rank1(n.Kept(eventTypeProp))`; `canonlinks.EventTypeProperty(v graphcache.Vocabulary) []byte` *(new)* finds that id (key `event_type`, origin `provenencia`) for `promotealign` and the Reader, once per call site. `neighborTypeProperty` stays in `promotealign` (the layer loader reads it); `canonlinks` names `vocabkey.PropertyEventType` | Moves, not copies (round 1 X-3) | Duplicating them — two link readings drift |
| D-8 | `graphcache.Vocabulary.PropertiesByID() map[string]VocabProperty` *(new)* beside `TermsByID`; `promotealign.propertiesByID` and `conclusiondetails.loadProperties`'s inline map call `PropertiesByID`. `(n *Node) Kept(propID []byte) []Value` *(new)* returns the `kept` rows and `func Rank1(rows []Value) (Value, bool)` *(new)* picks the rank-1 row (moved from `conclusionheaders` `kept` and `rank1`, which are deleted; `keptName`, `keptText`, `keptTerm` and `keptDateAt` call them); `Vocabulary.TypeOrder`'s inline props map uses `PropertiesByID` | One place builds each vocabulary lookup and one reads the kept value (M-3, M-6) | A third props map in `NewReader` |
| D-9 | `canonlinks.NewReader(g *graphcache.Graph, rules kinship.Rules) (*Reader, error)` loads terms and the `event_type` property id (`EventTypeProperty`) once; a neighbor is an event when its `Link.NeighborTypeKey` is `event`. `(*Reader).For(id []byte) ([]kinship.Implied, error)` loads `id` and its neighbors (`g.Node`, `g.Nodes`), takes the event result of each event neighbor in `Node.Links` order, then finds `id`'s parents and children: first from its `relationship·parent` edges (in `Node.Links` order, either direction), then from the `birth_parents` links in those event results; each person is taken once, at first sight (a set by id). It loads each such person m and m's neighbors, and takes m's chain result. It returns only links with `id` as `A` or `B`: event results' links first, then chain results', each in the result's order. It does not dedupe: links from different pivots differ in `Event` or `Through` | Every choice the walk and the evidence query will rely on is fixed here (round 1 X-4) | Returning whole pivot results — callers would filter |
| D-10 | The stored value is `pivotResult{kind pivotKind; rules kinship.Rules; links []kinship.Implied}` *(new)*, `pivotKind` `pivotEvent` or `pivotChain`. A result with other `Rules` is derived again and replaced. The `Check` rebuilds the result of the stored kind from the current cached nodes and compares with `reflect.DeepEqual`; a node it needs that is not cached is an error | The check knows what to rebuild (X-5); tests can use another bound without reading the registry's result | Inferring the kind from the node's type at check time |
| D-11 | Results are shared and read-only. `For` returns a new slice; the `Implied` values' byte slices are the cache's and are never written | Same rule as the Evidence graph memos (`TestMemoIsNotMutated`) | Cloning on every read |
| D-12 | `core/database/graphcache/graphcachetest` *(new)*: `OpenCatalog(t testing.TB) *database.Catalog` (moved from `graph_test.go` `openCatalog`; every caller in `graph_test.go`, `vocab_test.go` and `evidence_graph_walk_test.go` moves to it; `graphcachetest` blank-imports `core/database/autoreconciler` so `graphcache.SetTruth` is registered for verify mode, as `graph_test.go` does today), and `Canon` built by `NewCanon(t testing.TB, c *database.Catalog) *Canon` with `Entity(kind string) []byte` (16-byte ids `{3,1,0,…,n}` from a counter, ref `TST-<n>`), `Term(entity []byte, propKey, termKey string)` (a rank-1 `kept` term value), `Link(kind string, ends map[string][]byte, termProp, termKey string) []byte` (an association whose ends are keyed by endpoint property: `person`/`event` for a participation, `person`/`related_to` for a relationship), `Remove(id []byte)` (deletes the entity's values and row), `Retype(entity []byte, propKey, termKey string)` (replaces the kept term), `Commit(ids ...[]byte)` (`OnCommit(effects.Commit{Effects: effects.Set{Handles: ids}})`) and `CommitVocabulary()` (`OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}})`). `TestCommitOfAnAssociationLoadsEndpoints` and `TestCommitReloadsASharedEndpointOnce` move onto it | One canon builder for `graphcache` and `canonlinks` tests and S11-39's later goldens (M-8) | A private builder in `canonlinks_test` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `kinship.Rules` *(new)*, `core/matchregistry/kinship.go` *(new)* | type, file | #567 `16ff9ef2` | Ported from it |
| `NodeMemo`, `MemoCheck`, `MemoDeps`, `MemoSlot`, `RegisterNodeMemo`, `Memo`, `SetMemo`, `Cached` *(new)* | types, funcs | `details` (dead), `SetDisplayDeps`, `SetEvidenceGraphCheck` | Replaces `details`; reuses the deps-and-check registration shape |
| `Vocabulary.PropertiesByID` *(new)* | func | `Vocabulary.TermsByID` | Same shape; two copies move onto it |
| `EventTypeProperty` *(new)* | func | `promotealign` `keptTerm`'s key scan | Resolves the id once instead of matching keys per value |
| `Node.Kept` *(new)* | method | `conclusionheaders` `kept`, `promotealign` `keptTerm` | Moved; both copies go |
| `core/database/canonlinks` *(new)* | package | `promotealign` `canonKeys`, `loadTerms` | They move here (D-6, D-7) |
| `Reader`, `NewReader`, `For`, `pivotResult` *(new)* | type, funcs | `loadCanon`'s walk | Reads one person's pivots, not a component |
| `graphcachetest` *(new)* | test package | `graph_test.go` `openCatalog` and inline closures; `segmentcomparetest`, `namevaluetest` | Exported fixture packages are the repo's pattern; the private copies move onto it |

**Builds on.** `kinship.Derive`; `OnCommit`'s `roots`; the deps and check registration of `conclusionheaders` and `graphfill`; `matchregistry`'s section-per-layer layout.
**Deletes.** `graphcache` `details`, `Detail`, `SetDetail`, `SetDetailCheck`, `detailFn`, the `details` Verify loop; `promotealign` `loadTerms`, `canonKeys`, `keptTerm`, `propertiesByID`'s body; `conclusionheaders` `kept`; `conclusiondetails`' inline props map; `graph_test.go` `openCatalog` and the inline closures of `TestCommitOfAnAssociationLoadsEndpoints`; `vocabkey.RoleSpouse`, the `role` seed `spouse`, its Swift wording and three xcstrings keys.
**Next adds.** S11-39: `loadCanon` calls `Reader.For` per walked node and replaces its `typeKey` scan. S11-40: `canonLinks` calls `For`. A tree or person page calls `For`. A new event pattern is a row in `eventPatterns`; a two-hop pattern changes `canonlinks`' pivots, `For` and its `Deps` only. Another per-node memo is one `RegisterNodeMemo`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `Rules` *(new)*; `Derive(edges []Edge, terms segmentcompare.Terms, rules Rules) []Implied`; the bound; set dedupe; marriage reads `RoleSubject`; package doc: "the catalog's bridge terms and the Rules come in as arguments" |
| `core/vocabkey` | `RoleSpouse` removed |
| `core/matchregistry` | `Registry.Kinship kinship.Rules` *(new)*; `defaultKinship()` *(new)* = 10 |
| `core/database/subjectvocab` | `role` `spouse` removed from the seed |
| `core/database/graphcache` | D-5 memo slots; `Cached`; `PropertiesByID`, `Node.Kept` (D-8); `details` removed |
| `core/database/graphcache/graphcachetest` *(new)* | D-12 |
| `core/database/canonlinks` *(new)* | `canonlinks.go`: `LoadTerms`, `EdgeOf` (moved); `reader.go`: `Reader`, `NewReader`, `For`, `pivotResult`, `eventResult`, `chainResult`, `init` (slot, `Deps`, `Check`) |
| `core/database/promotealign` | Calls `canonlinks.LoadTerms`, `canonlinks.EdgeOf`, `Vocabulary.PropertiesByID`; no behavior change |
| `core/database/conclusionheaders`, `core/database/conclusiondetails` | Call `Node.Kept` and `PropertiesByID`; no behavior change |

```go
// EdgeOf reads one filed link of self as an edge between its two ends,
// ordered as recorded, with the key it has from each. neighbor is read only
// when it is the event end; nil is allowed otherwise. ok is false for an
// unregistered bridge.
func EdgeOf(self []byte, n *graphcache.Node, l graphcache.Link, neighbor *graphcache.Node,
	eventTypeProp []byte, terms segmentcompare.Terms) (from, to []byte, keyFrom, keyTo vocabkey.BridgeKey, ok bool)
```

Errors: none new; `For` returns the graph's load errors.

### FFI

None — no wire change.

### macOS

None — no chrome. Removed: `SeededRoleTerm.spouse`, `L10n` `roleSpouse` and its `wording` case.

### Strings

Removed: `propertyTerm.role.spouse`, `evidenceGraph.share.role.spouse`, `evidenceGraph.share.role.spouse.similar`. None added.

## 7. Steps

**STEP-1 — Bound what one pivot implies.**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/matchregistry/kinship.go` *(new)*, `core/matchregistry/registry.go`, `docs/matching-model.md` (knob row only). Skill: `add-match-rule`.
Do: D-1, porting `16ff9ef2` as the Precedent says. Every `Derive` call in `core/kinship` tests passes a test-local `Rules{MaxPivotEnd: 10}`.
Check: `go test ./core/kinship/ ./core/matchregistry/` passes, including `TestDeriveBound`, `TestLayerPackages` and `TestKnobInventoryMatchesRegistry`. Set the bound to 1000 once, see `TestDeriveBound`'s 11-rows fail, then revert.

**STEP-2 — Linear grandparent associations.**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`.
Do: D-2, porting `9c29cbaf`.
Check: `TestDeriveManyAssociations` passes in under 1 s.

**STEP-3 — The couple are the marriage's subjects.**
Files: `core/kinship/patterns.go`, `core/kinship/kinship_test.go`, `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/vocabkey_test.go`, `core/segmentcompare/segmentcomparetest/terms.go`, `core/subjectprofile/invariants_test.go`, `core/subjectprofile/lifespan_test.go`, `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-seeded-vocabulary`, `add-localized-string`.
Failing test first: `TestDerive`'s marriage rows use `subject` (AC-3.1) and fail.
Do: D-3, porting `c14e0d3c`'s hunks for these files after STEP-1 and STEP-2.
Check: `grep -rnE 'RoleSpouse|SeededRoleTerm\.spouse|roleSpouse|role\.spouse' core api macos` prints nothing (a comment in `api/ffi/swift_vocabulary_test.go` naming the old line may stay); `scripts/check-pr.sh` passes, including `TestVocabKeysMatchSeed` and `TestSwiftVocabularyMatchesGo`.

**STEP-4 — One place for vocabulary lookups and the kept value.**
Files: `core/database/graphcache/vocab.go`, `core/database/graphcache/graph.go` (`Node.Kept`), `core/database/promotealign/profiles.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusionheaders/compose.go`.
Do: D-8.
Check: `grep -rn 'func kept(\|func propertiesByID' core/database` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes unchanged.

**STEP-5 — Registered per-node memos, dropped on commit.**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/display.go`, `core/database/graphcache/commit.go`, `core/database/graphcache/verify.go`, `core/database/graphcache/memo.go` *(new)*, `core/database/graphcache/graph_test.go`, `core/database/graphcache/vocab_test.go`, `core/database/graphcache/evidence_graph_walk_test.go`, `core/database/graphcache/graphcachetest/graphcachetest.go` *(new)*.
Failing test first: `TestNodeMemoDropsOnCommit` (T-7).
Do: D-5, D-12. Put `RegisterNodeMemo`, `Memo`, `SetMemo` and `Cached` in `memo.go`; delete `details` and its API.
Check: T-7 passes; `TestCommitOfAnAssociationLoadsEndpoints` passes on the builder; `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes.

**STEP-6 — Move link reading and terms to `canonlinks`.**
Files: `core/database/canonlinks/canonlinks.go` *(new)*, `core/database/promotealign/canon.go`, `core/database/promotealign/evidence.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/terms.go`, `core/database/promotealign/export_test.go`.
Do: D-7. Two `canonKeys` call sites and two `loadTerms` call sites (`promotealign.go`, `export_test.go` `LayerBridges`) move. No behavior change.
Check: `grep -rn 'canonKeys\|func loadTerms\|func keptTerm' core/database/promotealign` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/` passes unchanged.

**STEP-7 — `Reader.For` over cached pivots.**
Files: `core/database/canonlinks/reader.go` *(new)*, `core/database/canonlinks/reader_test.go` *(new)*.
Failing test first: `TestFor` (T-4).
Do: D-4, D-6, D-9, D-10, D-11. Register the slot in `init` with its `Deps`; leave `Check` nil until STEP-8.
Check: T-4 passes.

**STEP-8 — Commits, oracle and verify.**
Files: `core/database/canonlinks/reader.go` *(new)*, `core/database/canonlinks/reader_test.go` *(new)*, both created in STEP-7.
Failing test first: `TestForFollowsCommits` (T-5).
Do: the `Check` (D-10).
Check: T-5, T-6, T-8 pass; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/` passes.

**STEP-9 — Docs.**
Files: `docs/deployment-plan/spike-11/kinship-shapes.md` (§ Where implied links live: drops go through the commit and `canonlinks`' deps; the "One pure package" paragraph: `EdgeOf` and `LoadTerms` live in `canonlinks`), `docs/ideas/kinship-view.md` ("by end" → "per pivot, through `canonlinks.Reader.For`"), `docs/matching-model.md` (glossary rows "kinship pattern" and "pivot"), `.agents/skills/add-match-rule/SKILL.md` (line 33's bullet: `Derive(edges, terms, rules)`, the bound, the canon read through `canonlinks`; "Add a kinship pattern": a step for `canonlinks` — the pivot kind, what `For` reads, and how far `Deps` reaches), `docs/deployment-plan/spike-11/completed.md` (newest first), `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports 0 errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/kinship_test.go` `TestDeriveBound` *(ported)* | 10 vs 11 spouses → 45 links vs none; 10 vs 11 birth parents → 10 vs none; 10 vs 11 parents of a middle person → 10 vs none | AC-1.5, D-1 |
| T-2 | Go | `core/kinship/kinship_test.go` `TestDeriveManyAssociations` *(ported)* | one parent link with 20,000 associations, each twice, × 10 grandchildren → under 1 s, each association once in first-added order | D-2 |
| T-3 | Go | `core/kinship/kinship_test.go` `TestDerive` (rows updated) | two `subject`s of a `marriage` → one spouse link; `witness` → none | AC-3.1, AC-1.6 |
| T-4 | Go | `core/database/canonlinks/reader_test.go` *(new)* `TestFor` *(new)* | one subtest each for AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7 on `graphcachetest.Canon`; compares `Pattern`, `Event`/`Through`, ends, keys and `Associations` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, D-4, D-9 |
| T-5 | Go | `core/database/canonlinks/reader_test.go` *(new)* `TestForFollowsCommits` *(new)* | one subtest each for AC-2.1 (`Remove` + `Commit`), AC-2.2 (`Retype` + `Commit`), AC-2.3 (`Link` + `Commit`), AC-2.4 (`Remove` + `Commit`), AC-2.5 (`CommitVocabulary`, then `Memo` reports no entry), AC-2.7 (`g.Node(Bob)`, then `Memo` still holds M1's result) | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.7, D-5 |
| T-6 | Go | `core/database/canonlinks/reader_test.go` *(new)* `TestForMatchesDerive` *(new)* | two marriages, three births and two filed parent links over six people: for each person, `For` equals `kinship.Derive` over every `EdgeOf` edge in the canon, filtered to links with that person as an end; links compared sorted by (`Pattern`, `Event`, `Through`, `A`, `B`), `Associations` as sets | D-4, D-9 |
| T-7 | Go | `core/database/graphcache/graph_test.go` `TestNodeMemoDropsOnCommit` *(new)* | a test slot whose `Deps` returns the roots' neighbors: `SetMemo` on an event and a participant; `Commit(event)` → both gone; an untouched node's entry stays; `g.Node` on a new id keeps every entry; `Drop` → all gone; a vocabulary commit empties a `DropOnVocabulary` slot and keeps another | D-5 |
| T-8 | Go | T-4, T-5, T-6 under `PROVENENCIA_GRAPH_VERIFY=1`; `core/database/canonlinks/reader_test.go` *(new)* `TestVerifyCatchesAStaleResult` *(new)* | a result replaced through `SetMemo` with a wrong link → `Verify` returns an error | AC-2.6, D-10 |
| T-9 | Go | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed`, `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (existing) | pass without `spouse` | AC-3.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (STEP-6), `core/database/conclusionheaders` and `conclusiondetails` (STEP-4), `core/subjectprofile` (STEP-3).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/ passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] No stale strings: the three removed keys are gone from L10n.swift and Localizable.xcstrings
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Domain docs updated: kinship-shapes.md, ideas/kinship-view.md, matching-model.md, add-match-rule SKILL.md
- [ ] PR description: what shipped, AC evidence, deviations, the ported commits from #567
```

## 10. Manual check

None in the app for the read — no reader yet. The reviewer runs:

1. `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/ -run 'TestFor' -v` and sees every AC subtest pass.
2. Creates a new project, adds an event participation, and opens the role menu: `Spouse` is not offered.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A commit changes a pivot's inputs without the pivot or a neighbor in `roots` | `roots` already holds committed ids and every old and new association endpoint; `Deps` adds one hop; T-5 and verify mode over every test |
| A result reads a node a later commit removed | The removed node is a root and its neighbors are its `Deps`; `Check` errors on a missing node (T-8) |
| Moving `kept` or `propertiesByID` changes a header or detail | STEP-4's check runs those packages' tests unchanged |

**Stop and report if:**

- A path replaces or removes cached nodes outside `OnCommit` and `clear` while memos exist (other than a first load).
- A chain result needs a node more than one hop from m.
- Porting `16ff9ef2`, `9c29cbaf` or `c14e0d3c` needs a hunk outside the files this plan names.
- Moving `canonKeys`, `kept` or `propertiesByID` changes any existing test's result.
- `TestForMatchesDerive` differs from `For` on any person.

## 12. Open questions

None.

## 13. PR

**Title:** S11-49 — Kinship links in the graph cache
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; the commits ported from #567; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-49-kinship-graph-cache.md`.
