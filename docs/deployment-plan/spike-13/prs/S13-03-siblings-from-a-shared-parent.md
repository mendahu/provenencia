# S13-03 — Siblings from a shared parent

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 3 — Family
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine read with no caller in the app until S13-04)
**Board:** —
**Depends on:** S11-39 (landed [#572](https://github.com/mendahu/provenencia/pull/572)), S11-51 (landed [#575](https://github.com/mendahu/provenencia/pull/575))
**Blocks:** S13-04
**Branch:** `s13-03-siblings-from-a-shared-parent`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01 and S13-02 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-queries.mdc`
**Context:** [person-detail.md § 2 Family](../person-detail.md#2-family) (R-17, R-18, R-18a); [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md); S11-51 plan D-4 – D-6 and its Landed note (git history: `f22b19a7:docs/deployment-plan/spike-11/prs/S11-51-person-links.md`); working rules: "Siblings are page-only", "The family read is served from the graph cache", "The family read is the family tree's first step", "Canonical only, filtered in Go", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-11-01 against `main` @ `0c166624`
**Review:** — (not reviewed yet)

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

The engine can answer "who are this person's siblings" from the canon: each person filed as a sibling (a `sibling` relationship link, parents unknown), and each other child of any of the person's parents, whether that parent is filed as a parent or read from a birth's `father` or `mother`. A pair found both ways is one sibling, marked recorded, and every sibling says what it rests on: the sibling link's association, or the parent and the links to that parent. The answer is stored per person in the graph cache, so a second read issues no query. Matching never sees it: a shared parent is still not sibling evidence for Promote (R-18a).

**Why now.** The Person page's Family read (S13-04) needs siblings beside S11-51's `Reader.Relationships`, and slice 3's Check is bar 3 ("their siblings, including siblings known only through a shared parent") and bar 8 ("Matching is unchanged"). The working rule "The family read is served from the graph cache" names this PR as the one that registers and fills the sibling memo slot.

**What stays.** `kinship.Derive`, `eventPatterns`, `Patterns()`, `Reader.For`, `Reader.Relationships` and their memo slots, every Promote proposal, band and golden, and the warm proposal budget. No RPC, no Swift, no catalog change.

## 2. User stories

**US-1.** As the engine, when the Family read asks for a canonical person's siblings, `canonlinks` returns each sibling once with what it rests on, so that the page shows recorded and implied siblings without a merge of its own.

Fixture names below are `canonlinks` `world` names: people Ann, Bea, Cal, Dan, Eve; `AB` is a relationship link filed from Ann to Bea; `B1/Ann` is Ann's participation in birth `B1`. "Sibling keys" is `w.keys(relationship, sibling)`; ends are compared lesser name first.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Bea and Cal filed as siblings (`BC`, term `sibling`), no parents; `Siblings(Bea)` | One link: {Bea, Cal}, sibling keys, Associations {`BC`}, no shapes, `OnlyImplied` false. `Siblings(Cal)` returns the same link |
| AC-1.2 | Ann filed as parent of Bea (`AB`) and of Cal (`AC`); `Siblings(Bea)` | One link: {Bea, Cal}, sibling keys, Associations {`AB`, `AC`}, Shapes [`shared_parent/Ann`], `OnlyImplied` true |
| AC-1.3 | Births `B1` (Ann `father`, Bea `subject`) and `B2` (Ann `father`, Cal `subject`), nothing filed; `Siblings(Bea)` | One link: {Bea, Cal}, Associations {`B1/Ann`, `B1/Bea`, `B2/Ann`, `B2/Cal`}, Shapes [`shared_parent/Ann`], `OnlyImplied` true |
| AC-1.4 | AC-1.2's catalog plus `BC` (term `sibling`); `Siblings(Bea)` | One link: Associations {`AB`, `AC`, `BC`}, Shapes [`shared_parent/Ann`], `OnlyImplied` false (found both ways: recorded) |
| AC-1.5 | Ann and Dan each filed as parent of Bea and of Cal (`AB`, `AC`, `DB`, `DC`); `Siblings(Bea)` | One link: Associations {`AB`, `AC`, `DB`, `DC`}, Shapes, sorted, [`shared_parent/Ann`, `shared_parent/Dan`], `OnlyImplied` true |
| AC-1.6 | Ann filed as parent of Bea and Cal; Dan filed as parent of Cal only; `Siblings(Bea)` and `Siblings(Cal)` | Each returns one link {Bea, Cal} with Shapes [`shared_parent/Ann`] only (a half-sibling is a sibling through the one shared parent) |
| AC-1.7 | Ann parent of Bea (`AB`); Ann and Dan filed as spouses (`AD`); Dan parent of Eve (`DE`); Bea parent of Cal (`BC2`); `Siblings` of Bea, Cal and Eve | All three return no links: a parent's spouse's child, a parent and a grandchild are not siblings |
| AC-1.8 | Ann parent of Bea and of 14 other people (`AC1` … `AC14`); `Siblings(Bea)` | 14 links, one per other child, in the order Ann's links list them (no `MaxPivotEnd` bound, D-5) |
| AC-1.9 | A person with no links; 16 zero bytes (an id not in the catalog); a person filed as their own sibling | No links and no error for each |

**US-2.** As the engine, a person's siblings are served from the graph cache, so that the Family read is cheap when warm and correct after an edit.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.2's catalog; `Siblings(Bea)` twice with `g.SetQueryCounterForTest` counting the second call | The second call issues 0 queries and returns a result `reflect.DeepEqual` to the first |
| AC-2.2 | A fresh catalog for each of: Ann parent of Bea and 1 other child; Ann parent of Bea and 7 other children (each child also the `subject` of a birth with Ann as `father`); a cold `Siblings(Bea)` counted | Both counts are equal (the cold read is a fixed number of batched loads, not one per sibling) |
| AC-2.3 | Ann parent of Bea; `Siblings(Bea)` is empty; then `AC` (Ann parent of Cal) is filed and committed | `Siblings(Bea)` returns {Bea, Cal} through Ann |
| AC-2.4 | AC-1.2's catalog; `Siblings(Bea)` stored; a commit that names only Ann | The slot holds no answer for Bea after the commit; the next `Siblings(Bea)` derives AC-1.2's link again and stores it |
| AC-2.5 | AC-1.2's catalog; `Siblings(Bea)` stored; a vocabulary commit | The slot holds no answer; the next read derives it again |
| AC-2.6 | AC-1.2's catalog; `Siblings(Bea)` stored; `g.Verify()`; then a stale answer (no links) set in the slot for Bea; `g.Verify()` | The first `Verify` passes; the second fails |
| AC-2.7 | AC-1.2's catalog; the result of `Siblings(Bea)`; the caller does nothing to it; a second call | The second call returns the same slice (shared, read-only, as `For`'s is) |

**US-3.** As a contributor, the derivation is present and matching is unchanged, so that a shared parent is not sibling evidence (R-18a, bar 8).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `kinship.Derive` over AC-1.2's and AC-1.3's edges (as `kinship.Edge`s) | Returns no link with sibling keys and no link with Pattern `shared_parent`; `kinship.Patterns()` is unchanged (`TestPatterns` passes untouched) |
| AC-3.2 | AC-1.2's catalog; `For(Bea)` and `Relationships(Bea)` deep-copied; then `Siblings(Bea)`; then `For(Bea)` and `Relationships(Bea)` again | Each equals its copy; neither holds a link with sibling keys or a `shared_parent` shape |
| AC-3.3 | A tree with Ann filed as parent of Bea and of Cal, all promoted; a Source recording Bea and Cal as siblings; propose that Source; then `canonlinks.Reader.Siblings` for the canon's Bea and Cal; propose again | The two `graphalign.Proposal`s are `reflect.DeepEqual` |
| AC-3.4 | The `promotealign`, `graphalign` and `kinship` suites, `TestForMatchesDerive`, and `TestLargeSourceBudgets` three times | Pass with no test or budget changed |
| AC-3.5 | `rg -n '\.Siblings\(' --glob '!*_test.go' core/database/promotealign core/graphalign core/segmentcompare` | Prints nothing: no production matching code calls it |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey.TermSibling` *(new)* and its `TestVocabKeysMatchSeed` row (D-10) | STEP-1 |
| `kinship.Siblings` *(new)* and `kinship.PatternSharedParent` *(new)* in `core/kinship/siblings.go` *(new)*, outside `eventPatterns`, `Derive` and `Patterns()` (D-1 – D-6) | STEP-2, AC-1.1 – AC-1.9, AC-3.1 |
| `deriver.kinEdges` *(new)*: `chainLinks`' read of a person's relationship and birth links, named once (D-9) | STEP-3, AC-3.4 |
| `canonlinks.Reader.Siblings` *(new)*, `deriver.siblingLinks` *(new)*, the `canonlinks.siblings` memo slot *(new)* with its `Check` (D-7, D-8) | STEP-4, AC-1.1 – AC-1.9, AC-2.1 – AC-2.7, AC-3.2 |
| A `promotealign` test that proposals are unchanged with sibling answers stored (D-11) | STEP-5, AC-3.3, AC-3.4 |
| `add-match-rule`, `docs/ideas/kinship-view.md`, the `kinship` and `canonlinks` package docs say where page-only derivations go (D-12) | STEP-6 |

| Out | Home |
| --- | --- |
| `GetPersonFamily`, family units, filtering merged or unaccepted people, attaching `PersonSummary`s, the multi-id walk | S13-04 |
| The Family diagram | S13-12 |
| Half-sibling and step-sibling as kinds of their own (the shapes already tell one shared parent from two) | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Sibling evidence in matching | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) (R-18a) |
| Moving `segmentcompare`'s link reading to its own pure package before a page imports `core/kinship` | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Link reading below the layers" (no page imports `core/kinship` in this PR) |
| Siblings in `Reader.For` or `Reader.Relationships` | not planned (R-18a, working rule "Siblings are page-only") |
| Flagging a pair filed both as siblings and as spouses | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Conflicts between shapes" |
| Warming every person's siblings at open | not planned (working rule "The family read is served from the graph cache") |
| A `matching-model.md` change | not planned (no knob is added) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Term keys | `core/vocabkey/vocabkey.go` `TermSpouse`, `TermParent`, `TermGrandparent` | No `sibling` constant. The seed has `relationship_type` `sibling` (`core/database/subjectvocab/registry.go`, undirected, no inverse) |
| Seed guard | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | Rows `{"relationship_type", string(vocabkey.TermSpouse)}` … `TermGrandparent` pin each term constant to the seed |
| Kinship | `core/kinship/kinship.go` `Edge`, `Edge.From`, `Edge.reads`, `Pattern`, `PatternMarriageSpouses`, `PatternBirthParents`, `PatternParentOfParent`, `Implied`, `Shape`, `Recording`, `Derive`, `ParentStep`, `chains` | Pure, guard-listed (`core/matchregistry/guard_test.go` `{"core/kinship", true}`: no numeric literal other than 0 and 1). `Implied`'s doc: "A is the relationship's person end: the parent or grandparent; for spouses, the lesser id". `ParentStep` is "the one rule for what a parent step is". No sibling rule anywhere |
| Patterns | `core/kinship/patterns.go` `eventPatterns`, `PivotKind`, `Patterns` | `Patterns()` is the event patterns then `parent_of_parent`; `PivotKind` returns "" for a code that is no pattern. Called only by `kinship_test.go` `TestPatterns` |
| Merge | `core/kinship/links.go` `Link`, `LinkSet`, `NewLinkSet`, `Add`, `AddImplied`, `Links` | The one same-ends, same-keys merge (S11-51). `Add` clones what it keeps; a merged link stays `OnlyImplied` only when both were |
| Canon reads | `core/database/canonlinks/canonlinks.go` package doc, `LoadTerms`, `EdgeOf` | The package doc lists `EdgeOf`, `LoadTerms`, `Reader.For`, `Reader.Relationships` |
| Implied links | `core/database/canonlinks/reader.go` `kinshipSlot`, `forSlot`, `init`, `checkFor`, `check`, `deriver`, `newDeriver`, `deriver.edge`, `deriver.relationshipEdges`, `deriver.eventLinks`, `deriver.eventLinksAt`, `deriver.chainLinks`, `Reader`, `NewReader`, `Reader.event`, `Reader.loadAround`, `Reader.For`, `Reader.Relationships`, `deriver.forLinks` | Two memo slots. `forSlot` (`canonlinks.for`) is `DropOnVocabulary` and `DropOnHandleCommit`, stores `forResult{rules, links}`, and its `Check` rebuilds through a deriver over `g.Cached` only. `chainLinks` builds its edges as `relationshipEdges(n)` plus each `PatternBirthParents` link of `eventLinksAt(n, eventResult)`, then calls `Derive`. `For` loads around each person one parent step away (one `Nodes` batch per such person) |
| Memo slots | `core/database/graphcache/memo.go` `NodeMemo`, `RegisterNodeMemo`, `Memo`, `SetMemo`, `Cached`, `dropHandleMemos`, `verifyMemos` | `SetMemo` ignores an id that is not 16 bytes. `Verify` runs every slot's `Check` |
| Query counter | `core/database/graphcache/graph.go` `SetQueryCounterForTest`, `Nodes`, `Node` | `Nodes(ids)` loads the missing ids in one batch |
| Matching's reader | `core/database/promotealign/canon.go` `loadCanon` | Calls `reader.For` per walked person; never `Relationships` or any sibling read |
| Tests | `core/database/canonlinks/reader_test.go` `world`, `newWorld`, `person`, `event`, `join`, `relationship`, `parent`, `keys`, `marriage`, `relShape`, `relShapes`, `lesserFirst`, `hasMemo`, `TestRelationships`, `TestForFollowsCommits`, `TestForMatchesDerive`, `TestVerifyCatchesAStaleResult` | The catalog fixture over `graphcachetest.Canon` (`w.k.Commit`, `w.k.CommitVocabulary`, `w.k.Remove`). `relShapes` reads a `[]kinship.Link` as names, associations sorted, shapes as `pattern/pivot` |
| Tests | `core/kinship/kinship_test.go` `terms`, `rules`, `p`, `e`, `b`, `part`, `rel`, `flip`, `link`, `TestDerive`, `TestPatterns`; `links_test.go` `TestLinkSet` | `package kinship_test`, ids as one-byte slices |
| Tests | `core/database/promotealign/promotealign_test.go` `newFixture`, `fixture.person`, `fixture.relationship`, `fixture.promote`, `fixture.newSource`, `fixture.graph`, `fixture.propose`, `TestProposeKinshipShapes`; `perf_test.go` `TestLargeSourceBudgets` | The Promote fixture and the warm budget |
| Docs | [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) "Add a kinship pattern" step 2 | Says "a pattern of another shape (siblings, great-grandparents) changes `Derive` and is planned as its own PR". The "Kinship patterns" bullet lists `canonlinks`' reads without a sibling read |
| Docs | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) lines 5, 10 | Line 5 says Spike 13 takes "a page-only sibling derivation from a shared parent"; line 10 lists "Siblings from shared parents" among other patterns |

**Precedent.** S11-51 ([#575](https://github.com/mendahu/provenencia/pull/575)) added `Reader.Relationships` beside `For` and `deriver.relationshipEdges` to name a loop two callers shared: copy that shape for `Reader.Siblings` and `deriver.kinEdges`, and its test layout (`TestRelationships` with `relShapes` and `lesserFirst`). S11-49 ([#569](https://github.com/mendahu/provenencia/pull/569)) registered `forSlot` with `DropOnHandleCommit` and a `Check` over `g.Cached`: copy `checkFor` and `forResult` for the sibling slot. Do not copy `For`'s load pattern (one `loadAround` per person one step away): D-8 loads in batches.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The rule is one pure function in `core/kinship/siblings.go` *(new)*: `func Siblings(person []byte, edges []Edge, terms segmentcompare.Terms) []Link`. It is not a row of `eventPatterns`, not read by `Derive`, and not in `Patterns()` | Working rule "Siblings are page-only": the derivation lives in `core/kinship` but outside the pattern set matching reads. A pure function over edges is testable without a catalog, as `Derive` is | A row in `eventPatterns` or a rule in `Derive` — matching would read it (R-18a); a rule inside `canonlinks` — the rule would not be pure and could not be tested over plain edges |
| D-2 | `PatternSharedParent Pattern = "shared_parent"` *(new)*, declared in `siblings.go`. An implied sibling is a `kinship.Implied` with `Pattern: PatternSharedParent`, `Through:` the shared parent, `Event: nil`, and `From:` the edges it was read from, so its `Shape()` is `{shared_parent, parent}` as a chain's is `{parent_of_parent, middle}`. `Patterns()` and `PivotKind` are unchanged; `Patterns()`' doc gains "Page-only codes (`PatternSharedParent`) are not in it". `Implied`'s doc becomes "for spouses and siblings, the lesser id" | One shape for "implied through a person" lets S13-04 say what each link rests on with the same `Recording` every other link carries (working rule "Every link says why") | A sibling type of its own — S13-04 would merge two shapes of link; adding the code to `Patterns()` — the pattern set is matching's |
| D-3 | `Siblings` builds its answer through one `kinship.NewLinkSet()`: first, in `edges` order, `Add(Link{Edge: e})` for each edge whose keys read as `segmentcompare.Keys(rel, vocabkey.TermSibling, "", terms)` from either end (through `Edge.reads`), with `person` at one end and a different id at the other; then `AddImplied` for each implied sibling, parents in first-seen order, each parent's children in first-seen order. It returns `set.Links()` | A pair found both ways merges into one link that is not `OnlyImplied`: "a pair found both ways appears once, as recorded" falls out of `Recording.Merge` with no rule of its own | Two lists (recorded, implied) — S13-04 would merge them; implied first — the link would carry the implied edge's end order instead of the filed one |
| D-4 | A parent step is what `kinship.ParentStep` reads, from any edge in `edges` (a filed `parent` or `child` link, or a `PatternBirthParents` link the caller passes in). `person`'s parents are the `parent` ends of steps whose `child` is `person`. Each such parent's children are the `child` ends of that parent's steps other than `person`. The implied link for parent `q` and child `c`: `Edge{A, B}` = `person` and `c` with the lesser id (`bytes.Compare`) as A, keys from `Keys(rel, TermSibling, "", terms)`, `Associations` = every association of every step `q → person` and `q → c` in `edges`, first-seen order, each once; `From` = those steps. Repeated steps (the same link passed twice) add no second association | `ParentStep` is the one rule for a parent step; reading birth links through it makes "a shared parent (direct or implied)" one rule | Reading roles or event types in `Siblings` — a second parent rule beside `ParentStep` |
| D-5 | `Siblings` has no bound: no `Rules` argument, no `MaxPivotEnd`. Its work is linear in `len(edges)` (each edge is read once into maps; each (parent, child) pair adds one implied link), and its output has at most one link per other child of a parent | `MaxPivotEnd` (10) would drop every sibling of a family with 11 children, which is real data; the cost the bound exists for (pairs of pairs) does not arise here. A birth event with too many participants already implies no parent link through `Derive`'s bound, so it reaches `Siblings` as no step | Reusing `MaxPivotEnd` on a parent's children — drops real families; a new knob — no tunable here, and the numbers guard keeps literals out of `core/kinship` |
| D-6 | Nothing is filtered: a sibling who is a merged node or has no accepted member is returned, as `Relationships` returns such links (S11-51 D-5). A person filed as their own sibling, and a step whose ends are one id, give nothing | One reach for every `canonlinks` read; S13-04 applies "Canonical only" to every relative at once | Filtering here — S13-04 would filter parents and spouses elsewhere, two rules |
| D-7 | `func (r *Reader) Siblings(id []byte) ([]kinship.Link, error)` *(new)* in `core/database/canonlinks/reader.go`. It reads `siblingSlot` *(new)*, registered in `init` as `graphcache.NodeMemo{Name: "canonlinks.siblings", Check: checkSiblings, DropOnVocabulary: true, DropOnHandleCommit: true}`, storing `siblingResult{rules kinship.Rules; links []kinship.Link}` *(new)*. A stored answer under other rules is derived again. The returned slice and its links are the cache's: shared and read-only, as `For`'s are. `checkSiblings` *(new)* rebuilds through a deriver over `g.Cached` (the shape of `checkFor`) and fails on a mismatch or a wrong stored type | A sibling is two to four hops away, too far for `Deps` to name, so the slot drops on any handle commit (working rule "The family read is served from the graph cache"). The rules are part of the answer because the birth links it reads come from `Derive` under `MaxPivotEnd` | A new package (`persondetail`) owning the slot — the slot's `Check` needs `canonlinks`' deriver; `Deps` naming the parents — a parent's new child is not named by a commit on the person, and the child's commit does not name the person |
| D-8 | `func (d *deriver) siblingLinks(id []byte, load func([][]byte) error, event func([]byte) ([]kinship.Implied, error)) ([]kinship.Link, error)` *(new)*: `n := d.node(id)` (nil node → nil, nil); `load` the ids of `n.Links`' neighbors; `edges := d.kinEdges(n, event)`; the parents are the `ParentStep` parents of `id` among `edges`, distinct, in order; `load(parents)`; `load` every neighbor of every parent node in one call; `edges` gains `d.kinEdges(parent, event)` for each parent node in order; return `kinship.Siblings(id, edges, d.terms)`. `Reader.Siblings` passes `r.g.Nodes` and `r.event`; `checkSiblings` passes a `load` that does nothing and `d.eventLinks`. It does not call `For`, `Relationships` or `chain` | A cold read is at most four batched loads (the person, their neighbors, the parents, the parents' neighbors) plus the event results the kinship slot already caches, whatever the number of siblings (AC-2.2). `For` loads around each person one step away, one batch per child of each parent | Reading `Relationships` of the person and each parent — it calls `For`, which loads one batch per child and adds grandparent chains the rule ignores |
| D-9 | `func (d *deriver) kinEdges(n *graphcache.Node, event func([]byte) ([]kinship.Implied, error)) ([]kinship.Edge, error)` *(new)*, beside `relationshipEdges`: `d.relationshipEdges(n)`, then the `Edge` of each `PatternBirthParents` link of `d.eventLinksAt(n, event)`, in that order. `chainLinks` becomes `edges, err := d.kinEdges(n, eventResult)` then `kinship.Derive(edges, d.terms, d.rules)`, with the same edges in the same order as today | One place reads a person's kin edges; `chainLinks` and `siblingLinks` both need exactly those (S11-51 D-6 did the same for `relationshipEdges`) | A copy of the loop in `siblingLinks` |
| D-10 | `TermSibling BridgeTerm = "sibling"` *(new)* in `core/vocabkey/vocabkey.go` after `TermGrandparent`, and the row `{"relationship_type", string(vocabkey.TermSibling)}` in `TestVocabKeysMatchSeed` | `add-match-rule` step 1: a term a rule names is a `vocabkey` constant pinned to the seed | A string literal `"sibling"` in `core/kinship` |
| D-11 | Tests: the pure cases in `core/kinship/siblings_test.go` *(new)* `TestSiblings`; the catalog cases in `core/database/canonlinks/reader_test.go` `TestSiblings`, `TestSiblingsFollowCommits`, `TestSiblingsQueryCount`, `TestSiblingsLeaveMatchingAlone`, `TestVerifyCatchesAStaleSiblingAnswer` *(all new)*; the proposal case in `core/database/promotealign/promotealign_test.go` `TestProposeIgnoresStoredSiblings` *(new)*, which builds a `canonlinks.NewReader(f.c.Graph(), matchregistry.Default().Kinship)` in the test. `world.sibling(name, a, b)` *(new)* is a one-line call to `world.relationship(name, a, b, "sibling")`; `world.siblings(id)` *(new)* and `world.hasSiblingAnswer(id)` *(new)* mirror `relationships` and `hasMemo` | Each case is tested at the lowest layer that can see it, as S11-51 did | A handler test — there is no RPC until S13-04 |
| D-12 | Docs: `add-match-rule` step 2's last clause becomes "a pattern of another shape (great-grandparents) changes `Derive` and is planned as its own PR. A page-only derivation (siblings from a shared parent: `kinship.Siblings`, read through `canonlinks.Reader.Siblings` and its own memo slot) is not a pattern: it stays out of `eventPatterns`, `Derive` and `Patterns()`, and a test proves matching's inputs do not change"; the "Kinship patterns" bullet's list of `canonlinks` reads adds "and `Reader.Siblings`, page-only". `docs/ideas/kinship-view.md` line 5 adds "S13-03 shipped it (`canonlinks.Reader.Siblings`)"; line 10 drops "Siblings from shared parents;". The `kinship` package doc adds "`Siblings` reads a person's siblings for the page; matching never calls it". The `canonlinks` package doc adds "and one person's siblings, recorded and through a shared parent, for the page (`Reader.Siblings`)" | Working rule "Each PR updates the docs and skills it makes false" names `add-match-rule` and `kinship-view.md` for S13-03 | Leaving step 2 saying siblings change `Derive` |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `vocabkey.TermSibling` *(new)* | const | `TermSpouse`, `TermParent`, `TermGrandparent` | A fourth row of the same list (D-10) |
| `core/kinship/siblings.go` *(new)* | file | `core/kinship/kinship.go` (`Derive`, `chains`) | `kinship.go` is the pattern set matching reads; a file of its own keeps the page-only rule apart from it (D-1) |
| `kinship.PatternSharedParent` *(new)* | const | `PatternParentOfParent` | Same shape (a person between), different rule; not in `Patterns()` (D-2) |
| `kinship.Siblings` *(new)* | func | `kinship.Derive`, `chains` | `Derive` is matching's; `chains` bounds by `MaxPivotEnd` and pairs grandparents (D-1, D-5). It reuses `ParentStep`, `Edge.reads`, `LinkSet`, `Implied`, `Recording` |
| `canonlinks.siblingSlot`, `siblingResult`, `checkSiblings` *(new)* | var, type, func | `forSlot`, `forResult`, `checkFor` | A second per-person answer with the same drop rules; a separate slot keeps matching's `For` answer untouched (D-7, AC-3.2) |
| `Reader.Siblings` *(new)* | func | `Reader.For`, `Reader.Relationships` | Neither returns siblings, and neither may (R-18a) |
| `deriver.siblingLinks` *(new)* | func | `deriver.forLinks` | `forLinks` loads one batch per person one step away and derives chains (D-8) |
| `deriver.kinEdges` *(new)* | func | the edge build in `deriver.chainLinks` | That build, named once for two callers (D-9) |
| `core/kinship/siblings_test.go` *(new)* | test file | `kinship_test.go` `TestDerive` | Reuses `terms`, `p`, `b`, `rel`, `part`, `flip`, `link` |
| `world.sibling`, `world.siblings`, `world.hasSiblingAnswer` *(new)* | test helpers | `world.parent`, `world.relationships`, `world.hasMemo` | One-line mirrors (D-11) |
| `TestProposeIgnoresStoredSiblings` *(new)* | test | `TestProposeKinshipShapes` | Reuses `newFixture`, `person`, `relationship`, `promote`, `newSource`, `graph`, `propose` |

Searched for an existing sibling read by name (`rg -n -i sibling core api --type go`: only the seed row and `segmentcompare`/`graphalign` tests that use the term as a plain link) and by behavior (`ParentStep` callers: `Derive`, `forLinks`): none.

**Builds on.** S11-51's `kinship.LinkSet` and `Reader.Relationships`, S11-49's `canonlinks` deriver and memo slots, `kinship.ParentStep`, `graphcache.RegisterNodeMemo` with `DropOnHandleCommit`, the `world` fixture.
**Deletes.** The edge build in `deriver.chainLinks` (moved into `kinEdges`, D-9). The clause "a pattern of another shape (siblings, great-grandparents) changes `Derive`" in `add-match-rule` and "Siblings from shared parents;" in `kinship-view.md` line 10.
**Next adds.** S13-04's reusable family function calls `Reader.Siblings` per id beside `Reader.Relationships` and adds the sibling links to the person's units; its "Canonical only" filter drops merged and unaccepted siblings. A future tree reads the same slot per generation. Another page-only derivation (aunts and uncles through a parent's siblings) goes beside `Siblings` in `siblings.go`'s shape, never in `Derive`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A third memo slot on every graph | `graphcache.dropHandleMemos`, `dropVocabularyMemos`, `verifyMemos` | They loop every registered slot; a slot with nothing stored is skipped. AC-2.4, AC-2.5, AC-2.6 |
| `deriver.eventLinksAt` reached for a parent's node | `reader.go` `eventLinksAt`, `Reader.event` | It already runs for any person in `chainLinks`; the parent's event neighbors are loaded first (D-8), so `r.event` reads cached nodes |
| An id that is no node, or a parent id that is no node (deleted meanwhile) | `deriver.siblingLinks` | `d.node` returns nil: the person gives no links; a nil parent node is skipped (AC-1.9) |

### Scale

A family has no Scale line in the map. At what a large family looks like (2 parents, 15 children, each child with a birth): a cold `Siblings` is at most four `Nodes` batches (D-8), measured equal for 2 and 8 children (AC-2.2); a warm one is zero queries (AC-2.1). `Siblings` is linear in the edges of the person and their parents (D-5): 15 children give at most 2 × 15 parent steps in and 14 links out. A handle commit empties the slot; the next read rebuilds from cached nodes. Matching's warm budget does not move: `loadCanon` never reads the slot (AC-3.4, AC-3.5).

### Edges

- **Assumptions.** `LinkSet` keys a link by its ends and keys in a fixed order, so `{Bea, Cal}` and `{Cal, Bea}` with symmetric sibling keys are one link (AC-1.1, AC-1.4). `ParentStep` rejects a step whose ends are one id (AC-1.9). Sibling keys are symmetric (`relationship_type` `sibling` is undirected in the seed and in `segmentcomparetest.Terms()`), so `Keys` returns one key twice and `Edge.reads` matches from either end (AC-1.1).
- **Neighbors.** Closest on the "not a sibling" side: a parent's spouse's child (step-sibling), a grandchild, a parent (AC-1.7). Closest on the "sibling" side: one shared parent only (AC-1.6), two shared parents (AC-1.5), a parent known only from a birth (AC-1.3).
- **Removed inputs.** A parent link removed and committed drops the stored answer; the next read loses that parent's shape, and the sibling goes when no shared parent or sibling link remains (AC-2.3 and AC-2.4 in reverse). The recorded sibling link carries `OnlyImplied` false; with it gone the link reads implied again.
- **Real data.** Not table-driven: the only seeded term the rule names is `sibling`, pinned by D-10. A birth with more than `MaxPivotEnd` people at an end implies no parent today and so no sibling (D-5).
- **Untrusted size.** A parent with thousands of children gives one link per child, linear in links already loaded; the number of parents is bounded by the person's own links. No dimension multiplies (D-5).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `TermSibling` (D-10) |
| `core/kinship` | `siblings.go` *(new)*: `PatternSharedParent`, `Siblings` (D-1 – D-6). `kinship.go`: `Implied` doc (D-2), package doc (D-12). `patterns.go`: `Patterns()` doc (D-2) |
| `core/database/canonlinks` | `reader.go`: `siblingSlot`, `siblingResult`, `checkSiblings`, registration in `init` (D-7); `deriver.kinEdges` and `chainLinks` (D-9); `deriver.siblingLinks` (D-8); `Reader.Siblings` (D-7). `canonlinks.go`: package doc (D-12) |

```go
// PatternSharedParent is how a sibling is implied through a parent both have.
// It is page-only: Derive never makes it and Patterns does not list it.
const PatternSharedParent Pattern = "shared_parent"

// Siblings returns person's siblings among edges: each recorded sibling link
// with person at one end, and one link to each other child of each of
// person's parents, implied through that parent (Through), merged through one
// LinkSet so a pair found both ways is one link, recorded. Parents are read
// through ParentStep. It is for the page; matching never calls it.
func Siblings(person []byte, edges []Edge, terms segmentcompare.Terms) []Link

// Siblings returns id's siblings, recorded and through a shared parent, from
// one stored answer per person that any handle commit drops. The slice and its
// links are the cache's, shared and read-only. Nothing is filtered.
func (r *Reader) Siblings(id []byte) ([]kinship.Link, error)
```

Errors: none new; `Reader.Siblings` returns `Nodes`', `Node`'s and `event`'s errors.

### FFI

None — no RPC (S13-04).

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — The sibling term**
Files: `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/vocabkey_test.go`. Skill: `add-match-rule` (step 1 of "Add a kinship pattern").
Do: D-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/vocabkey/ ./core/database/subjectvocab/` passes.

**STEP-2 — The pure rule**
Files: `core/kinship/siblings.go` *(new)*, `core/kinship/siblings_test.go` *(new)*, `core/kinship/kinship.go` (`Implied` doc, package doc), `core/kinship/patterns.go` (`Patterns` doc). Skill: `add-match-rule`.
Do: T-1 first (it fails to build), then D-1 – D-6. No numeric literal other than 0 and 1 in `siblings.go`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/ ./core/matchregistry/` passes (the `matchregistry` guard reads `core/kinship`).

**STEP-3 — One read of a person's kin edges**
Files: `core/database/canonlinks/reader.go`.
Do: D-9. Behavior does not change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/ ./core/database/promotealign/` passes with no test changed.

**STEP-4 — Siblings from the graph cache**
Files: `core/database/canonlinks/reader.go`, `core/database/canonlinks/canonlinks.go` (package doc), `core/database/canonlinks/reader_test.go`.
Do: T-2 – T-6 first, then D-7, D-8 and the package doc line (D-12).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -count=1 ./core/database/canonlinks/` passes.

**STEP-5 — Prove matching is unchanged**
Files: `core/database/promotealign/promotealign_test.go`.
Do: T-7 (D-11). Then run the budgets and the grep.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./core/graphalign/ ./core/kinship/` passes; `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -count=3 ./core/database/promotealign/` passes; `rg -n '\.Siblings\(' --glob '!*_test.go' core/database/promotealign core/graphalign core/segmentcompare` prints nothing.

**STEP-6 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `docs/ideas/kinship-view.md`, `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (S13-03 row status), this plan (status, Landed note).
Do: D-12, then the Definition of done's docs lines.
Check: `rg -n 'siblings, great-grandparents' .agents/skills/add-match-rule/SKILL.md` prints nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-03-siblings-from-a-shared-parent.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period ("Read a person's siblings from a shared parent or a sibling link, for the page only.", "Store each person's siblings in the graph cache until the next handle commit."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/siblings_test.go` `TestSiblings` *(new)*, subtests named by AC | AC-1.1 – AC-1.7, AC-1.9 (self and no edges) over `rel`, `part` and `link` edges: recorded only; shared filed parent; shared birth parent (pass the `PatternBirthParents` `Implied.Edge`s `Derive` returns for two births); both ways (`OnlyImplied` false); two parents (two shapes, associations once each); half; step, grandchild and parent give nothing; the same parent step passed twice adds no association; a `flip`ped recorded sibling edge reads the same. Plus AC-3.1: `Derive` over the same edges holds no link with sibling keys or `PatternSharedParent`, and `slices.Contains(kinship.Patterns(), kinship.PatternSharedParent)` is false | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.9, AC-3.1, D-1, D-2, D-3, D-4, D-5, D-6 |
| T-2 | Go | `core/database/canonlinks/reader_test.go` `TestSiblings` *(new)*, subtests named by AC | AC-1.1 – AC-1.9 in the catalog, built with `world`, `world.sibling`, `world.parent`, `world.relationship`, `world.event`, `world.join`; read with `world.siblings` through `relShapes` and `lesserFirst`; AC-1.8's 14 links compared in order | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-2.7, D-6, D-7, D-8 |
| T-3 | Go | `reader_test.go` `TestSiblingsFollowCommits` *(new)* | AC-2.3 (file and `w.k.Commit` the new parent link), AC-2.4 (`w.k.Commit(ann)` with no change; `hasSiblingAnswer(bea)` false, then true after a read), AC-2.5 (`w.k.CommitVocabulary()`) | AC-2.3, AC-2.4, AC-2.5, D-7 |
| T-4 | Go | `reader_test.go` `TestSiblingsQueryCount` *(new)* | AC-2.1 (warm: 0 queries, same result) and AC-2.2 (cold: equal counts for 1 and 7 other children, each also born of a birth with Ann as `father`), counted with `w.g.SetQueryCounterForTest` | AC-2.1, AC-2.2, D-8 |
| T-5 | Go | `reader_test.go` `TestSiblingsLeaveMatchingAlone` *(new)* | AC-3.2: deep copies of `For(bea)` and `Relationships(bea)` before `Siblings(bea)` equal the reads after; no link in either has sibling keys or a `shared_parent` shape | AC-3.2, D-7 |
| T-6 | Go | `reader_test.go` `TestVerifyCatchesAStaleSiblingAnswer` *(new)* | AC-2.6, the shape of `TestVerifyCatchesAStaleResult` with `siblingSlot` and `siblingResult{rules: shipped}` | AC-2.6, D-7 |
| T-7 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeIgnoresStoredSiblings` *(new)* | AC-3.3: tree Ann parent of Bea and of Cal, all promoted; Source 1 records Bea sibling Cal; `p1 := f.propose(f.graph(src), nil)`; `Siblings` of the canon's Bea and Cal; `p2 := f.propose(f.graph(src), nil)`; `reflect.DeepEqual(p1, p2)` | AC-3.3, D-1, D-7 |
| T-8 | Go | `PROVENENCIA_GRAPH_VERIFY=1` over `./core/database/canonlinks/`; `TestLargeSourceBudgets` three times; the STEP-5 grep | Pass; pass; prints nothing | AC-2.6, AC-3.4, AC-3.5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestForMatchesDerive`, `TestFor`, `TestForFollowsCommits` (D-9's `chainLinks` change), the `matchregistry` guard (`siblings.go` in a numbers package), `TestVocabKeysMatchSeed` (D-10), `TestProposeKinshipShapes` and the `graphalign` goldens (AC-3.4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it) — this PR renames nothing; the D-12 clause grep in STEP-6 prints nothing
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — no brief
- [ ] Domain doc updated: add-match-rule and ideas/kinship-view.md (D-12); matching-model.md unchanged (no knob)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: no view reads siblings until S13-12, and S13-04's handler tests carry the first end-to-end proof. The reviewer checks matching instead:

1. Open a project whose tree files a parent with two children. Promote a Source that records those two children as siblings. The Promote rows read as they do on `main` (T-7 carries the proof).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| D-9 changes `chainLinks`' edge order and a chain or golden moves | `kinEdges` keeps today's order (relationship edges, then births); STEP-3 runs before any new behavior, with no test changed |
| A cold `Siblings` loads one batch per child | D-8 batches the parents and their neighbors; T-4 compares counts at two family sizes |
| The stored answer outlives an edit two hops away | `DropOnHandleCommit` (D-7); T-3; verify mode (T-8) |
| A caller edits the shared slice | Doc says read-only, as `For`'s; S13-04 adds the links to its own `LinkSet`, which clones |

**Stop and report if:**

- `deriver.chainLinks`, `deriver.eventLinksAt`, `Reader.For`, `forSlot` or `kinship.LinkSet` no longer reads as §4 says.
- STEP-3 needs a test changed to pass.
- `TestLargeSourceBudgets` fails at STEP-5 where it passes on `main` in the same run conditions (three runs each).
- The `matchregistry` guard rejects `core/kinship/siblings.go` *(new)*.
- `segmentcompare.Keys(rel, vocabkey.TermSibling, "", terms)` returns two different keys (the term reads directed), so a recorded and an implied sibling do not merge in AC-1.4.
- T-4's cold counts differ between the two family sizes with D-8's loads in place.
- T-7's two proposals differ, or `p1` differs from a proposal made with no `Siblings` call at all.

## 12. Open questions

None.

## 13. PR

**Title:** S13-03 — Siblings from a shared parent
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-03-siblings-from-a-shared-parent.md`.
