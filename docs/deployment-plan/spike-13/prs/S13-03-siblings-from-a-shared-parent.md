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
**Context:** [person-detail.md § 2 Family](../person-detail.md#2-family) (R-17, R-18, R-18a); [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md); S11-51 plan D-4 – D-6 and its Landed note (git history: `f22b19a7:docs/deployment-plan/spike-11/prs/S11-51-person-links.md`); working rules: "Siblings are page-only", "One person's links, direct and implied, from one function", "The family read is served from the graph cache", "The family read is the family tree's first step", "Canonical only, filtered in Go", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-11-01 against `main` @ `0c166624`
**Review:** [`S13-03-siblings-from-a-shared-parent.review.md`](S13-03-siblings-from-a-shared-parent.review.md) — round 1 answered; round 2 pending

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

The engine can answer "who are this person's siblings" from the canon, through the one function that already returns a person's relationships. `Reader.Relationships(id, canonlinks.WithSiblings)` adds, beside the recorded links it returns today (a filed `sibling` link among them), one link to each other child of any of the person's parents, whether that parent is filed as a parent or read from a birth's `father` or `mother`. A pair found both ways is one link, recorded; each implied sibling says what it rests on (the parent it is implied through and the links to that parent). The implied siblings are stored per person in the graph cache, loaded for many people in a fixed number of batches, so a second read issues no query. Matching never asks for them: a shared parent is still not sibling evidence for Promote (R-18a).

**Why now.** The Person page's Family read (S13-04) needs siblings, and slice 3's Check is bar 3 ("their siblings, including siblings known only through a shared parent") and bar 8 ("Matching is unchanged"). The working rule "The family read is served from the graph cache" names this PR as the one that registers and fills the sibling memo slot, and "One person's links, direct and implied, from one function" makes `Relationships` the one read S13-04 calls.

**What stays.** `kinship.Derive`'s output, `eventPatterns`, `Patterns()`, `Reader.For`'s answer, `Relationships(id)` with no option, every Promote proposal, band and golden, and the warm proposal budget. No RPC, no Swift, no catalog change.

## 2. User stories

**US-1.** As the engine, when the Family read asks for a canonical person's relationships with siblings, `canonlinks` returns each sibling once with what it rests on, merged with the person's other links, so that S13-04 needs no merge of its own.

Fixture names are `canonlinks` `world` names: people Ann, Bea, Cal, Dan, Eve, `C1` … `C14`; `AB` is a relationship link filed from Ann to Bea; `B1/Ann` is Ann's participation in birth `B1`. "Sibling links" are the links of `Relationships(id, WithSiblings)` whose keys are `w.keys(relationship, sibling)`; ends are compared lesser name first (`lesserFirst`).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Bea and Cal filed as siblings (`BC`, term `sibling`), no parents; sibling links of Bea | One link: {Bea, Cal}, Associations {`BC`}, no shapes, `OnlyImplied` false (as `Relationships(Bea)` returns it today). Cal's sibling links: the same link |
| AC-1.2 | Ann filed as parent of Bea (`AB`) and of Cal (`AC`); sibling links of Bea | One link: {Bea, Cal}, Associations {`AB`, `AC`}, Shapes [`shared_parent/Ann`], `OnlyImplied` true |
| AC-1.3 | Births `B1` (Ann `father`, Bea `subject`) and `B2` (Ann `father`, Cal `subject`), nothing filed; sibling links of Bea | One link: {Bea, Cal}, Associations {`B1/Ann`, `B1/Bea`, `B2/Ann`, `B2/Cal`}, Shapes [`shared_parent/Ann`], `OnlyImplied` true |
| AC-1.4 | AC-1.2's catalog plus `BC` filed from Bea to Cal (term `sibling`); sibling links of Bea | One link: A = Bea, B = Cal (the filed order), Associations {`AB`, `AC`, `BC`}, Shapes [`shared_parent/Ann`], `OnlyImplied` false (found both ways: recorded) |
| AC-1.5 | Ann and Dan each filed as parent of Bea and of Cal (`AB`, `AC`, `DB`, `DC`); sibling links of Bea | One link: Associations {`AB`, `AC`, `DB`, `DC`}, Shapes, sorted, [`shared_parent/Ann`, `shared_parent/Dan`], `OnlyImplied` true |
| AC-1.6 | Ann filed as parent of Bea and Cal; Dan filed as parent of Cal only; sibling links of Bea and of Cal | Each is one link {Bea, Cal} with Shapes [`shared_parent/Ann`] only (a half-sibling is a sibling through the one shared parent) |
| AC-1.7 | Ann parent of Bea (`AB`); Ann and Dan filed as spouses (`AD`); Dan parent of Eve (`DE`); Bea parent of Cal (`BC2`); sibling links of Bea, Cal and Eve | None for each: a parent's spouse's child, a parent and a grandchild are not siblings |
| AC-1.8 | Ann parent of Bea and of `C1` … `C14` (links `AC1` … `AC14`, filed in that order); sibling links of Bea | 14 links, one per `Cn`, in the order `C1` … `C14` (the order of Ann's `n.Links`; no `MaxPivotEnd` bound, D-5) |
| AC-1.9 | A person with no links; 16 zero bytes (an id not in the catalog); a person filed as their own sibling | No sibling links and no error for each; the unknown id stores no answer |
| AC-1.10 | Ann parent of Bea and Cal; Cal also filed as parent of Bea (`CB`); sibling links of Bea | One link {Bea, Cal} through Ann: a relative who is also a parent is returned in both roles, unflagged (D-6) |
| AC-1.11 | AC-1.2's catalog; the caller overwrites the first byte of every end and association of Bea's sibling link; a second `Relationships(Bea, WithSiblings)` | The second result equals AC-1.2's (the caller owns what `Relationships` returns) |

**US-2.** As the engine, siblings are served from the graph cache and loaded for many people at once, so that the Family read and a later tree are cheap when warm and correct after an edit.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.2's catalog; `Relationships(Bea, WithSiblings)` twice, the second counted with `g.SetQueryCounterForTest` | The second issues 0 queries and returns a result `reflect.DeepEqual` to the first |
| AC-2.2 | Fresh catalogs: (a) Ann parent of Bea and 1 other child; (b) Ann parent of Bea and 7 other children, each child also the `subject` of a birth with Ann as `father`; (c) three unrelated families shaped like (a); no read and no commit before the counter is set; `LoadSiblings` of Bea (a, b) and of the three first children (c), counted | Each count is greater than 0, and (a), (b) and (c) are equal: a fixed number of batched loads, not one per sibling or per person |
| AC-2.3 | Ann parent of Bea; Bea has no sibling links; then `AC` (Ann parent of Cal) is filed and committed | Bea's sibling links: {Bea, Cal} through Ann |
| AC-2.4 | AC-1.2's catalog; Bea's answer stored; a commit that names only Ann | The slot holds no answer for Bea; the next read derives AC-1.2's link again and stores it |
| AC-2.5 | AC-1.2's catalog; Bea's answer stored; a vocabulary commit | The slot holds no answer; the next read derives it again |
| AC-2.6 | AC-1.2's catalog; Bea's answer stored; `g.Verify()`; then a stale answer (no links) set in the slot for Bea; `g.Verify()` | The first passes; the second fails |
| AC-2.7 | AC-1.2's catalog; Ann's links and node removed and committed; Bea's sibling links read; the suite under `PROVENENCIA_GRAPH_VERIFY=1` | No sibling links, no error, Verify passes: a parent that is not cached is skipped on both the read and the check path (D-8) |

**US-3.** As a contributor, the derivation is present and matching is unchanged, so that a shared parent is not sibling evidence (R-18a, bar 8).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `kinship.Derive` over AC-1.2's and AC-1.3's edges (as `kinship.Edge`s) | No link with sibling keys and no link with Pattern `shared_parent`; `slices.Contains(kinship.Patterns(), kinship.PatternSharedParent)` is false |
| AC-3.2 | AC-1.2's catalog; `For(Bea)` and `Relationships(Bea)` deep-copied; then `Relationships(Bea, WithSiblings)`; then `For(Bea)` and `Relationships(Bea)` again | Each equals its copy; neither holds a `shared_parent` shape |
| AC-3.3 | A tree with Ann filed as parent of Bea and of Cal, all promoted; a second Source recording Bea and Cal as siblings; propose that Source; `LoadSiblings` of the canon's Bea and Cal; propose again | The two `graphalign.Proposal`s are `reflect.DeepEqual` |
| AC-3.4 | The `promotealign`, `graphalign` and `kinship` suites, `TestDerive`, `TestFor`, `TestForMatchesDerive`, and `TestLargeSourceBudgets` three times | Pass with no existing test or budget changed |
| AC-3.5 | `rg -n 'WithSiblings\|LoadSiblings\|kinship\.Siblings' --glob '!*_test.go' core/database/promotealign core/graphalign core/segmentcompare core/subjectprofile` | Prints nothing: no matching code asks for siblings |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey.TermSibling` *(new)* and its `TestVocabKeysMatchSeed` row (D-10) | STEP-1 |
| One parent-step index in `core/kinship` that `chains` and `Siblings` share (D-4) | STEP-2, AC-3.4 |
| `kinship.Parents`, `kinship.Siblings`, `kinship.PatternSharedParent` *(new)* in `core/kinship/siblings.go` *(new)*, outside `eventPatterns`, `Derive` and `Patterns()`; `PivotKind` names its pivot (D-1 – D-5) | STEP-3, AC-1.1 – AC-1.10, AC-3.1 |
| `canonlinks` groundwork with no behavior change: `deriver.kinEdges` for `chainLinks` and `forLinks`, `cachedNodes`, one `personAnswer` type and lookup, the deriver's loader (D-7, D-9) | STEP-4, AC-3.4 |
| `Reader.LoadSiblings` *(new)*, `canonlinks.WithSiblings` *(new)* on `Relationships`, the `canonlinks.siblings` memo slot *(new)* with its `Check` (D-6, D-8, D-11) | STEP-5, AC-1.1 – AC-1.11, AC-2.1 – AC-2.7, AC-3.2 |
| A `promotealign` test that proposals are unchanged with sibling answers stored (D-12) | STEP-6, AC-3.3, AC-3.4, AC-3.5 |
| `add-match-rule`, `docs/ideas/kinship-view.md`, the `kinship` and `canonlinks` package docs say where page-only derivations go (D-13) | STEP-7 |

| Out | Home |
| --- | --- |
| `GetPersonFamily`, family units, grouping a half-sibling under a parent set, filtering merged or unaccepted people, attaching `PersonSummary`s | S13-04 |
| The Family diagram | S13-12 |
| Half-sibling and step-sibling as kinds of their own (the shapes already tell one shared parent from two) | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Sibling evidence in matching | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) (R-18a) |
| Siblings in `Reader.For` | not planned (R-18a, working rule "Siblings are page-only": `For` is matching's pattern set) |
| Flagging a relative found in two roles (sibling and parent, sibling and spouse) | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Conflicts between shapes" |
| `For` loading around many people in one batch, and `forLinks` reading the deriver's loader | S13-04 (its multi-id walk; logged in the review, M-4 and M-6) |
| Moving `segmentcompare`'s link reading to its own pure package | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Link reading below the layers" (no page imports `core/kinship` in this PR) |
| Warming every person's siblings at open | not planned (working rule "The family read is served from the graph cache") |
| A `matching-model.md` change | not planned (no knob is added) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Term keys | `core/vocabkey/vocabkey.go` `TermSpouse`, `TermParent`, `TermGrandparent` | No `sibling` constant. The seed has `relationship_type` `sibling` (`core/database/subjectvocab/registry.go`, undirected, no inverse) |
| Seed guard | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | Rows `{"relationship_type", string(vocabkey.TermSpouse)}` … `TermGrandparent` pin each term constant to the seed |
| Kinship | `core/kinship/kinship.go` `Edge`, `Edge.From`, `Edge.reads`, `Pattern`, `Implied`, `Shape`, `Recording`, `Derive`, `step`, `ParentStep`, `chains` | Pure, guard-listed (`core/matchregistry/guard_test.go` `{"core/kinship", true}`). `Derive` builds `steps []step` from `ParentStep` over `edges` and again over the implied birth links. `chains` indexes them: `byPair map[pair][]int`, `order []pair`, `childrenOf map[string][]int` (distinct children of each parent, first-seen), `parents map[string]map[string]struct{}` (distinct parents of each middle), and builds each link's `from` and deduped `assoc` from `byPair[up]` plus `byPair[pair{up.m, down.m}]`. `Implied`'s doc: "for spouses, the lesser id" |
| Patterns | `core/kinship/patterns.go` `eventPatterns`, `PivotKind`, `Patterns` | `Patterns()` is the event patterns then `parent_of_parent`. `PivotKind` returns `Event` for an event pattern, `Person` for `PatternParentOfParent`, "" otherwise; called only by `TestPatterns` |
| Merge | `core/kinship/links.go` `Link`, `LinkSet`, `NewLinkSet`, `Add`, `AddImplied`, `Links` | The one same-ends, same-keys merge. `Add` clones what it keeps; a merged link stays `OnlyImplied` only when both were |
| Canon reads | `core/database/canonlinks/canonlinks.go` package doc | Lists `EdgeOf`, `LoadTerms`, `Reader.For`, `Reader.Relationships` |
| Slots and checks | `core/database/canonlinks/reader.go` `kinshipSlot`, `forSlot`, `forResult`, `init`, `checkFor`, `check`, `pivotResult` | `checkFor` and `check` each declare the same `cached` closure over `g.Cached` returning "is not cached". `forResult` is `{rules kinship.Rules; links []kinship.Implied}` |
| Deriver | `reader.go` `deriver`, `newDeriver(g, rules, node)`, `deriver.edge`, `relationshipEdges`, `eventLinks`, `eventLinksAt`, `chainLinks`, `forLinks` | `chainLinks` builds `relationshipEdges(n)` plus the `PatternBirthParents` links of `eventLinksAt(n, eventResult)`, then `Derive`. `forLinks` takes `eventLinksAt(n, event)` as `out`, filters its own `births`, and finds "the people one parent step from id" by `ParentStep` over `relationshipEdges(n)`, then over `births` |
| Reader | `reader.go` `Reader`, `NewReader`, `Reader.pivot`, `event`, `chain`, `loadAround`, `For`, `Relationships(id)` | `For` and `pivot` each read a slot, type-assert and compare rules. `For` returns before storing when the node is nil. `Relationships(id)` adds every `relationshipEdges(n)` link (a filed `sibling` link included) then `For(id)`'s links through one `LinkSet`; the caller owns the result. Its only callers are tests |
| Memo slots | `core/database/graphcache/memo.go` `NodeMemo`, `RegisterNodeMemo`, `Memo`, `SetMemo`, `Cached` | `SetMemo` ignores an id that is not 16 bytes; `DropOnHandleCommit` empties a slot on any handle commit; `Verify` runs every slot's `Check` |
| Loading | `core/database/graphcache/graph.go` `Nodes`, `Node`, `SetQueryCounterForTest` | `Nodes(ids)` loads the missing ids in one batch; a missing entity is never cached |
| Matching's reader | `core/database/promotealign/canon.go` `loadCanon` | Calls `reader.For` per walked person; never `Relationships` |
| Tests | `core/database/canonlinks/reader_test.go` `world`, `newWorld`, `person`, `event`, `join`, `relationship`, `parent`, `keys`, `marriage`, `relShape`, `relShapes`, `relationships`, `lesserFirst`, `hasMemo`, `TestRelationships`, `TestForFollowsCommits`, `TestForMatchesDerive`, `TestVerifyCatchesAStaleResult` | `hasMemo(id)` hard-codes `kinshipSlot` (its callers are in `TestForFollowsCommits`). `w.k.Commit`, `w.k.CommitVocabulary`, `w.k.Remove` |
| Tests | `core/kinship/kinship_test.go` `terms`, `rules`, `p`, `e`, `b`, `part`, `rel`, `flip`, `link`, `TestDerive`, `TestPatterns` | `package kinship_test`. `TestPatterns` asserts `PivotKind` per pattern and "" for `"nope"` |
| Tests | `core/database/promotealign/promotealign_test.go` `newFixture`, `fixture.person`, `fixture.relationship`, `fixture.promote`, `fixture.newSource`, `fixture.graph`, `fixture.propose`, `kinTree`; `perf_test.go` `TestLargeSourceBudgets` | `kinTree` files the tree on `f.source` and `f.artifact`, promotes its people (a relationship filed with `f.relationship` reaches the canon with them; it is not promoted on its own), then proposes another Source. The fixture's catalog is `f.c` |
| Docs | [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) "Add a kinship pattern" step 2; the "Kinship patterns" bullet | Step 2: "a pattern of another shape (siblings, great-grandparents) changes `Derive` and is planned as its own PR". The bullet lists `canonlinks`' reads |
| Docs | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) lines 5, 10 | Line 5: Spike 13 takes "a page-only sibling derivation from a shared parent"; line 10 lists "Siblings from shared parents" |
| Map | [`deployment-plan.md`](../deployment-plan.md) S13-03 row | The Plan column links this file (wired with this plan) |

**Precedent.** S11-51 ([#575](https://github.com/mendahu/provenencia/pull/575)) added `Reader.Relationships` and named the loop two callers shared (`relationshipEdges`): copy that shape for `kinEdges` and `cachedNodes`, and its test layout (`TestRelationships`, `relShapes`, `lesserFirst`). S11-49 ([#569](https://github.com/mendahu/provenencia/pull/569)) registered `forSlot` with `DropOnHandleCommit` and a `Check` over `g.Cached`: copy it for the sibling slot. Do not copy `For`'s load pattern (one `loadAround` per person one step away): D-8 loads in batches.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The rule is pure, in `core/kinship/siblings.go` *(new)*: `func Siblings(person []byte, edges []Edge, terms segmentcompare.Terms) []Implied` returns only the siblings a shared parent implies, and `func Parents(person []byte, edges []Edge, terms segmentcompare.Terms) [][]byte` returns `person`'s distinct parents among `edges` in first-seen order. Neither is a row of `eventPatterns`, read by `Derive`, or in `Patterns()` | Working rule "Siblings are page-only": the derivation lives in `core/kinship` outside the pattern set. Returning implied links only, as `Derive` does, leaves recorded links to the one place that reads them (`relationshipEdges`) and the merge to `LinkSet` | `Siblings` also scanning for recorded sibling links — a second read of what `Relationships` already returns (review M-1); a rule in `Derive` — matching would read it (R-18a) |
| D-2 | `PatternSharedParent Pattern = "shared_parent"` *(new)* in `siblings.go`. An implied sibling is `Implied{Edge, Pattern: PatternSharedParent, Through: parent, Event: nil, From: steps}`, so its `Shape()` is `{shared_parent, parent}` as a chain's is `{parent_of_parent, middle}`. `Edge.A` is the lesser id of `person` and the sibling (`bytes.Compare`), keys from `segmentcompare.Keys(rel, vocabkey.TermSibling, "", terms)`. `PivotKind` returns `vocabkey.Person` for `PatternSharedParent` as for `PatternParentOfParent`; `Patterns()` is unchanged and its doc gains "Page-only codes (`PatternSharedParent`) are not in it". `Implied`'s doc becomes "for spouses and siblings, the lesser id" | One shape for "implied through a person": S13-04 says what each link rests on, and labels the pivot, with the `Recording` and `PivotKind` every other link uses (working rule "Every link says why") | A sibling type of its own — two shapes of link to merge; adding the code to `Patterns()` — the pattern set is matching's |
| D-3 | `Siblings` emits, for each parent `q` of `person` in `Parents` order and each distinct child `c` of `q` other than `person` in first-seen order, one `Implied` with `Associations` and `From` from every step `q → person` and `q → c` (each association once, first-seen). The same child under two parents is two `Implied`s, which `LinkSet` merges into one link with two shapes | The merge already exists; one implied link per pivot is what `Derive` returns for every other pattern | Merging inside `Siblings` — a second merge beside `LinkSet` |
| D-4 | `chains` and `Siblings` share one unexported parent-step index in `kinship.go`: `stepsOf(edges []Edge, terms segmentcompare.Terms) []step` *(new)* (the `ParentStep` loop `Derive` runs today), and `indexSteps(steps []step) parentSteps` *(new)* with `byPair`, `order`, `childrenOf` and `parentsOf map[string][][]byte` (distinct parents of each child, first-seen; `chains` reads its `len`), and `func (s parentSteps) through(pairs ...pair) (from []Edge, assoc [][]byte)` *(new)* (the `from` and deduped `assoc` build `chains` does). `chains` keeps its output and order exactly; `Parents` is `indexSteps(stepsOf(edges, terms)).parentsOf[string(person)]` | One copy of duplicate-step handling and association order (review M-2); the next parent-based derivation (aunts and uncles, in [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Other patterns") is a third user, not a third copy | A second index in `siblings.go` |
| D-5 | No bound in `Siblings`: no `Rules` argument, no `MaxPivotEnd`. Its work is linear in `len(edges)`; the output has one `Implied` per (shared parent, other child) | `MaxPivotEnd` (10) would drop every sibling of a family of 11 children, which is real data. A birth with too many participants already implies no parent link through `Derive`'s bound, so it reaches `Siblings` as no step | Reusing `MaxPivotEnd` on a parent's children — drops real families; a knob — nothing to tune, and the numbers guard keeps literals out of `core/kinship` |
| D-6 | Nothing is filtered or flagged: a merged or unaccepted sibling is returned (S11-51 D-5); a relative who is also a parent, child or spouse of `person` is returned in each role (AC-1.10). A step or sibling link whose ends are one id gives nothing (`ParentStep` and `LinkSet` keys) | One reach for every `canonlinks` read; S13-04 applies "Canonical only" to every relative at once; conflicts between shapes stay in the idea file | Dropping a sibling who is also a parent — a silent rule over conflicting data, which the idea file leaves open |
| D-7 | `canonlinks` groundwork, no behavior change: `cachedNodes(g *graphcache.Graph) func([]byte) (*graphcache.Node, error)` *(new)* replaces the two `cached` closures; `personAnswer{rules kinship.Rules; links []kinship.Implied}` *(new)* replaces `forResult` and is the sibling slot's type too; `func (r *Reader) answer(slot graphcache.MemoSlot, id []byte) ([]kinship.Implied, bool)` *(new)* is the one stored-answer lookup for `For` and siblings (type-assert, same rules). `Reader.pivot` keeps its own lookup: its result also carries a `kind`. The deriver gains a `load func([][]byte) error` field *(new)*, set by `newDeriver(g, rules, node, load)`: `g.Nodes` in `NewReader`, a function returning nil in `check` and `checkFor` | One copy of each (review M-5, M-6); the next per-person slot is a registration, not a fourth copy | A `siblingResult` beside `forResult` — the same shape twice; a `load` parameter on each derivation — two ways to say whether a derivation may load |
| D-8 | `func (d *deriver) siblingsOf(ids [][]byte, event func([]byte) ([]kinship.Implied, error)) (map[string][]kinship.Implied, error)` *(new)* returns one answer per id that is a node (an id with no entry is no node): `d.load(ids)`; each node is `d.g.Cached(id)`, and an id not cached gets no entry; `d.load` every neighbor of every node in one call; per node, `atEnd := d.eventLinksAt(n, event)`, `edges := d.kinEdges(n, atEnd)`, `parents := kinship.Parents(id, edges, d.terms)`; `d.load` the union of all parents in one call; `d.load` every neighbor of every cached parent in one call; per parent, its `kinEdges` once (kept in a local map by id), and a parent not cached is skipped; per node, `kinship.Siblings(id, edges plus each of its cached parents' edges, d.terms)`. It does not call `For`, `chain` or `loadAround` | At most four batched loads for any number of ids and siblings (AC-2.2), as working rule "The family read is the family tree's first step" (3) asks. Reading nodes through `g.Cached` after the loads makes the read and the check skip the same missing parent (AC-2.7) | `Relationships` of each parent — it calls `For`, one batch per child; one id per call — a generation of N people costs 4N batches (review M-4) |
| D-9 | `func (d *deriver) kinEdges(n *graphcache.Node, atEnd []kinship.Implied) []kinship.Edge` *(new)*: `d.relationshipEdges(n)`, then the `Edge` of each `PatternBirthParents` link in `atEnd`, in that order. `chainLinks` calls it with `eventLinksAt(n, eventResult)`; `forLinks` calls it with its `out` and walks the result through `kinship.ParentStep` to find the people one step away (both ends), so its private `births` filter goes. Same edges, same order, same people as today | One place says what counts as a person's parent links for canon reads (review M-3): a baptism's parents later is one filter | A copy of the build in `siblingsOf` |
| D-10 | `TermSibling BridgeTerm = "sibling"` *(new)* after `TermGrandparent`, and the row `{"relationship_type", string(vocabkey.TermSibling)}` in `TestVocabKeysMatchSeed` | `add-match-rule` step 1: a term a rule names is a pinned `vocabkey` constant | A string literal in `core/kinship` |
| D-11 | The read: `type Option struct{ siblings bool }` *(new)*, `var WithSiblings = Option{siblings: true}` *(new)*, and `func (r *Reader) Relationships(id []byte, opts ...Option) ([]kinship.Link, error)`: with `WithSiblings`, after `For`'s links it calls `r.LoadSiblings([][]byte{id})` and `AddImplied`s each stored sibling. `func (r *Reader) LoadSiblings(ids [][]byte) error` *(new)* derives (D-8, with `r.event`) and stores the answer of each id with none stored under the reader's rules, in `siblingSlot` *(new)*: `graphcache.NodeMemo{Name: "canonlinks.siblings", Check: checkSiblings, DropOnVocabulary: true, DropOnHandleCommit: true}`, value `personAnswer`; an id that is no node stores nothing. `checkSiblings` *(new)* rebuilds one id through `siblingsOf` on a deriver over `cachedNodes` with `d.eventLinks`, and fails on a mismatch or a wrong type | The map's scope entry: "reachable only from an explicit option"; working rule "One person's links, direct and implied, from one function": S13-04 calls `LoadSiblings(ids)` once for a generation, then `Relationships(id, WithSiblings)` per id from the cache. A sibling is two to four hops away, too far for `Deps`, so the slot drops on any handle commit | A separate `Reader.Siblings` read — S13-04 would merge it with `Relationships` (review M-1); a boolean parameter — every existing call would change; `Deps` naming the parents — a parent's new child is not named by a commit on the person |
| D-12 | Tests: the pure cases in `core/kinship/siblings_test.go` *(new)* `TestSiblings`; `TestPatterns` gains the `PatternSharedParent` → `Person` row; catalog cases in `reader_test.go` `TestRelationshipsWithSiblings`, `TestSiblingsFollowCommits`, `TestLoadSiblingsQueryCount`, `TestSiblingsLeaveMatchingAlone`, `TestVerifyCatchesAStaleSiblingAnswer` *(all new)*; the proposal case `TestProposeIgnoresStoredSiblings` *(new)* in `promotealign_test.go`. Helpers: `world.sibling(name, a, b)` *(new)* calls `world.relationship(name, a, b, "sibling")`; `world.siblings(id)` *(new)* returns the sibling-key links of `Relationships(id, WithSiblings)`; `hasMemo(id)` becomes `hasMemo(slot, id)` and its calls pass `kinshipSlot` | Each case at the lowest layer that can see it, as S11-51 did; one memo helper (review M-9) | A handler test — no RPC until S13-04 |
| D-13 | Docs: `add-match-rule` step 2's last clause becomes "a pattern of another shape (great-grandparents) changes `Derive` and is planned as its own PR. A page-only derivation (siblings from a shared parent: `kinship.Siblings`, read through `canonlinks.Reader.Relationships` with `WithSiblings` and its own memo slot) is not a pattern: it stays out of `eventPatterns`, `Derive` and `Patterns()`, shares `kinship.go`'s parent-step index, and a test proves matching's inputs do not change"; the "Kinship patterns" bullet's `Reader.Relationships` clause adds "(with `WithSiblings`, also the siblings a shared parent implies, for the page; `LoadSiblings` loads them for many people)". [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) line 5 adds "S13-03 shipped it (`canonlinks.WithSiblings`)"; line 10 drops "Siblings from shared parents;". The `kinship` package doc adds "`Siblings` reads the siblings a shared parent implies, for the page; matching never calls it". The `canonlinks` package doc's `Reader.Relationships` clause adds "with, on request, the siblings a shared parent implies (`WithSiblings`, `LoadSiblings`)" | Working rule "Each PR updates the docs and skills it makes false" names `add-match-rule` and `kinship-view.md` for S13-03 | Leaving step 2 saying siblings change `Derive` |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `vocabkey.TermSibling` *(new)* | const | `TermSpouse`, `TermParent`, `TermGrandparent` | A fourth row of the same list (D-10) |
| `stepsOf`, `parentSteps`, `indexSteps`, `parentSteps.through` *(new)* | funcs, type | `Derive`'s step loop and `chains`' maps | Those, named once for two users (D-4) |
| `core/kinship/siblings.go` *(new)* | file | `core/kinship/kinship.go` | `kinship.go` is the pattern set matching reads; the page-only rule sits beside it and reuses its index (D-1, D-4) |
| `kinship.PatternSharedParent` *(new)* | const | `PatternParentOfParent` | Same shape (a person between), page-only (D-2) |
| `kinship.Siblings`, `kinship.Parents` *(new)* | funcs | `chains`, `ParentStep` | Built on `ParentStep` and the shared index (D-1, D-4) |
| `cachedNodes`, `personAnswer`, `Reader.answer`, `deriver.load` *(new)* | func, type, func, field | the two `cached` closures, `forResult`, the lookup in `For` | Each a third copy otherwise (D-7) |
| `deriver.kinEdges` *(new)* | func | the edge builds in `chainLinks` and `forLinks` | Those, named once (D-9) |
| `siblingSlot`, `checkSiblings`, `deriver.siblingsOf`, `Reader.LoadSiblings` *(new)* | var, funcs | `forSlot`, `checkFor`, `forLinks`, `For` | `For` is matching's and may not hold siblings; `forLinks` loads per person (D-8, D-11) |
| `canonlinks.Option`, `WithSiblings` *(new)* | type, var | `Relationships(id)` | The explicit page-only option on the existing read (D-11) |
| `core/kinship/siblings_test.go` *(new)* | test file | `kinship_test.go` `TestDerive` | Reuses `terms`, `p`, `e`, `b`, `rel`, `part`, `flip`, `link` |
| `world.sibling`, `world.siblings` *(new)*; `hasMemo(slot, id)` | test helpers | `world.parent`, `world.relationships`, `hasMemo(id)` | One-line mirrors; one memo helper (D-12) |
| `TestProposeIgnoresStoredSiblings` *(new)* | test | `kinTree` | Reuses its tree half (T-8) |

Searched for an existing sibling read by name (`rg -n -i sibling core api --type go`: only the seed row and tests using the term as a plain link) and by behavior (`ParentStep` callers: `Derive`, `forLinks`): none.

**Builds on.** S11-51's `LinkSet` and `Reader.Relationships`, S11-49's deriver and memo slots, `kinship.ParentStep`, `chains`' index, `graphcache.RegisterNodeMemo` with `DropOnHandleCommit`, the `world` fixture.
**Deletes.** `forResult` (→ `personAnswer`); the `cached` closures in `check` and `checkFor` (→ `cachedNodes`); the edge build in `chainLinks` and the `births` filter in `forLinks` (→ `kinEdges`); `chains`' private maps and `from`/`assoc` loop and `Derive`'s step loop (→ the index); `hasMemo(id)` (→ `hasMemo(slot, id)`); the clause "a pattern of another shape (siblings, great-grandparents) changes `Derive`" in `add-match-rule` and "Siblings from shared parents;" in `kinship-view.md`.
**Next adds.** S13-04's reusable family function calls `LoadSiblings(ids)` once, then `Relationships(id, WithSiblings)` per id, and groups by parent set: the shapes give each sibling's shared parents, and a half-sibling's other parents come from that sibling's own links, which S13-04 reads (the sibling's node is cached by D-8's loads when a filed link joins it to the parent; its other births may need one more batch, S13-04's to count). Its "Canonical only" filter drops merged and unaccepted relatives. A tree repeats the same two calls per generation. An aunts-and-uncles derivation goes in `siblings.go`'s shape on the same index.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A third memo slot on every graph | `graphcache.dropHandleMemos`, `dropVocabularyMemos`, `verifyMemos` | They loop every registered slot. AC-2.4, AC-2.5, AC-2.6 |
| `deriver.eventLinksAt` reached for a parent's node | `reader.go` `eventLinksAt`, `Reader.event` | Already runs for any person in `chainLinks`; the parent's neighbors are loaded first (D-8) |
| An id or parent that is no node, or not cached (deleted meanwhile) | `deriver.siblingsOf` | Skipped through `g.Cached` on both paths; nothing stored for a missing id (AC-1.9, AC-2.7) |
| `Relationships` with an option | its test callers | Variadic: existing calls are unchanged (AC-3.2) |

### Scale

A family has no Scale line in the map. At a large family (2 parents, 15 children, each with a birth) and at a generation (three people at once): `LoadSiblings` cold is at most four `Nodes` batches (D-8), proved equal for 1 and 7 siblings and for 1 and 3 people (AC-2.2); a warm `Relationships(id, WithSiblings)` is zero queries (AC-2.1). `Siblings` is linear in the edges of the person and their parents (D-5). A handle commit empties the slot; the next read rebuilds from cached nodes. `Relationships(id, WithSiblings)` also calls `For`, which loads one batch per person one step away: that cost is S13-04's to batch (Out). Matching's warm budget does not move: `loadCanon` never asks for siblings (AC-3.4, AC-3.5).

### Edges

- **Assumptions.** `LinkSet` keys a link by its ends and keys in a fixed order, so `{Bea, Cal}` and `{Cal, Bea}` with symmetric sibling keys are one link (AC-1.1, AC-1.4). `ParentStep` rejects a step whose ends are one id (AC-1.9). Sibling keys are symmetric (`sibling` is undirected in the seed and in `segmentcomparetest.Terms()`), so `Keys` returns one key twice (Stop condition). `chains` keeps its output order through the index (`TestDerive`, `TestForMatchesDerive` unchanged).
- **Neighbors.** Not a sibling: a parent's spouse's child, a grandchild, a parent (AC-1.7). A sibling: one shared parent only (AC-1.6), two (AC-1.5), a parent known only from a birth (AC-1.3), a relative also in another role (AC-1.10).
- **Removed inputs.** A parent link removed and committed drops the stored answer; the next read loses that parent's shape, and the sibling goes when no shared parent or sibling link remains (AC-2.3 in reverse). A parent no longer cached is skipped, not an error (AC-2.7).
- **Real data.** Not table-driven: the only seeded term the rule names is `sibling`, pinned by D-10. A birth with more than `MaxPivotEnd` people at an end implies no parent today and so no sibling.
- **Untrusted size.** A parent with thousands of children gives one link per child, linear in links already loaded; the number of parents is bounded by the person's own links; `LoadSiblings` over N ids is four batches. No dimension multiplies.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `TermSibling` (D-10) |
| `core/kinship` | `kinship.go`: `stepsOf`, `parentSteps`, `indexSteps`, `through`; `Derive` and `chains` on them (D-4); `Implied` doc and package doc (D-2, D-13). `siblings.go` *(new)*: `PatternSharedParent`, `Parents`, `Siblings` (D-1 – D-3, D-5). `patterns.go`: `PivotKind` case and `Patterns()` doc (D-2) |
| `core/database/canonlinks` | `reader.go`: `cachedNodes`, `personAnswer`, `Reader.answer`, `deriver.load`, `newDeriver` signature (D-7); `kinEdges`, `chainLinks`, `forLinks` (D-9); `siblingSlot`, `checkSiblings`, `siblingsOf`, `LoadSiblings`, `Option`, `WithSiblings`, `Relationships(id, opts...)` (D-8, D-11). `canonlinks.go`: package doc (D-13) |

```go
// PatternSharedParent is how a sibling is implied through a parent both have.
// It is page-only: Derive never makes it and Patterns does not list it.
const PatternSharedParent Pattern = "shared_parent"

// Parents returns person's distinct parents among edges, read through
// ParentStep, in first-seen order.
func Parents(person []byte, edges []Edge, terms segmentcompare.Terms) [][]byte

// Siblings returns the siblings person's parents imply among edges: one link
// per parent and other child of that parent, through the parent. It returns
// only implied links; LinkSet merges them with a recorded sibling link. It is
// for the page; matching never calls it.
func Siblings(person []byte, edges []Edge, terms segmentcompare.Terms) []Implied

// WithSiblings asks Relationships for the siblings a shared parent implies.
var WithSiblings = Option{siblings: true}

// LoadSiblings derives and stores the implied siblings of each id that has
// none stored, loading every id's nodes in a fixed number of batches.
func (r *Reader) LoadSiblings(ids [][]byte) error
```

Errors: none new; the reads return `Nodes`', `Node`'s and `event`'s errors.

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

**STEP-2 — One parent-step index**
Files: `core/kinship/kinship.go`. Skill: `add-match-rule`.
Do: D-4's `stepsOf`, `parentSteps`, `indexSteps`, `through`; `Derive` and `chains` read them. No behavior change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/ ./core/matchregistry/ ./core/database/canonlinks/` passes with no test changed.

**STEP-3 — The pure rule**
Files: `core/kinship/siblings.go` *(new)*, `core/kinship/siblings_test.go` *(new)*, `core/kinship/kinship_test.go` (`TestPatterns` row), `core/kinship/kinship.go` (docs), `core/kinship/patterns.go`. Skill: `add-match-rule`.
Do: T-1 and T-2 first (they fail to build), then D-1 – D-3, D-5 and the D-2 doc lines. No numeric literal other than 0 and 1 in `siblings.go`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/ ./core/matchregistry/` passes.

**STEP-4 — `canonlinks` groundwork**
Files: `core/database/canonlinks/reader.go`, `core/database/canonlinks/reader_test.go` (`hasMemo(slot, id)` only).
Do: D-7 and D-9. No behavior change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/ ./core/database/promotealign/` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -count=1 ./core/database/canonlinks/` passes.

**STEP-5 — Siblings from the graph cache**
Files: `core/database/canonlinks/reader.go`, `core/database/canonlinks/canonlinks.go` (package doc), `core/database/canonlinks/reader_test.go`.
Do: T-3 – T-7 first, then D-8, D-11 and the package doc (D-13).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -count=1 ./core/database/canonlinks/` passes.

**STEP-6 — Prove matching is unchanged**
Files: `core/database/promotealign/promotealign_test.go`.
Do: T-8. Then the budgets and the grep.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./core/graphalign/ ./core/kinship/` passes; `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -count=3 ./core/database/promotealign/` passes; the AC-3.5 grep prints nothing.

**STEP-7 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `docs/ideas/kinship-view.md`, `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (S13-03 row Status; the Plan column is already linked), this plan (status, Landed note).
Do: D-13, then the Definition of done's docs lines.
Check: `rg -n 'siblings, great-grandparents' .agents/skills/add-match-rule/SKILL.md` prints nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-03-siblings-from-a-shared-parent.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS` (on a host without Xcode, report every other check's PASS and `xcode-test` as not run: this PR has no Swift change).

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period ("Read the siblings a shared parent implies, for the page only.", "Return a person's siblings with their relationships when the page asks."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/siblings_test.go` `TestSiblings` *(new)*, subtests named by AC | Over `rel`, `part` and `link` edges, `Siblings` and `Parents`: shared filed parent (AC-1.2); shared birth parent, passing the `PatternBirthParents` `Implied.Edge`s `Derive` returns for two births (AC-1.3); two parents give two `Implied`s, one per parent (AC-1.5); half (AC-1.6); step, grandchild and parent give nothing (AC-1.7); a recorded sibling edge alone gives nothing (AC-1.1: recorded links are `Relationships`' job); 14 children in order (AC-1.8); self and no edges (AC-1.9); a child who is also a parent is returned (AC-1.10); the same step passed twice adds no association; `Parents` order and dedup. Plus AC-3.1 over the same edges | AC-1.1, AC-1.2, AC-1.3, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-3.1, D-1, D-2, D-3, D-4, D-5, D-6 |
| T-2 | Go | `core/kinship/kinship_test.go` `TestPatterns` | Gains `PatternSharedParent.PivotKind() == vocabkey.Person`; the `Patterns()` list is unchanged | AC-3.1, D-2 |
| T-3 | Go | `core/database/canonlinks/reader_test.go` `TestRelationshipsWithSiblings` *(new)*, subtests named by AC | AC-1.1 – AC-1.11 in the catalog, built with `world`, `world.sibling`, `world.parent`, `world.relationship`, `world.event`, `world.join`; read with `world.siblings` through `relShapes` and `lesserFirst` (AC-1.4 compares A and B in filed order instead). AC-1.8's expected order is `C1` … `C14`, and the test also asserts it is the order of those children among `w.g.Cached(ann).Links` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, D-6, D-8, D-11 |
| T-4 | Go | `reader_test.go` `TestSiblingsFollowCommits` *(new)* | AC-2.3 (file and `w.k.Commit` the new parent link); AC-2.4 (`w.k.Commit(ann)` with no change; `hasMemo(siblingSlot, bea)` false, then true after a read); AC-2.5 (`w.k.CommitVocabulary()`); AC-2.7 (`w.k.Remove` Ann's two links and Ann, `w.k.Commit` each, then read Bea's sibling links) | AC-2.3, AC-2.4, AC-2.5, AC-2.7, D-8, D-11 |
| T-5 | Go | `reader_test.go` `TestLoadSiblingsQueryCount` *(new)* | AC-2.1 (warm `Relationships(bea, WithSiblings)`: 0 queries, same result) and AC-2.2 (three fresh worlds, no read or commit before `w.g.SetQueryCounterForTest`; each count > 0; all equal) | AC-2.1, AC-2.2, D-8 |
| T-6 | Go | `reader_test.go` `TestSiblingsLeaveMatchingAlone` *(new)* | AC-3.2 | AC-3.2, D-11 |
| T-7 | Go | `reader_test.go` `TestVerifyCatchesAStaleSiblingAnswer` *(new)* | AC-2.6, the shape of `TestVerifyCatchesAStaleResult` with `siblingSlot` and `personAnswer{rules: shipped}` | AC-2.6, D-11 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeIgnoresStoredSiblings` *(new)* | AC-3.3, built as `kinTree`'s tree half: on `f.source` and `f.artifact`, `ann, bea, cal := f.person(f.source, f.artifact, "Ann Ames")`, `"Bea Ames"`, `"Cal Ames"`; `f.relationship(f.source, f.artifact, ann, bea, "parent")` and the same for `cal`; `f.promote` each person, keeping each `Entity.ID`. Then `src, art := f.newSource("Source 1")`, its own `f.person` Bea and Cal, `f.relationship(src, art, sbea, scal, "sibling")`. `p1 := f.propose(f.graph(src), nil)`; `canonlinks.NewReader(f.c.Graph(), matchregistry.Default().Kinship)` then `LoadSiblings` of Bea's and Cal's `Entity.ID`s; `p2 := f.propose(f.graph(src), nil)`; `reflect.DeepEqual(p1, p2)` | AC-3.3, D-1, D-11 |
| T-9 | Go | `PROVENENCIA_GRAPH_VERIFY=1` over `./core/database/canonlinks/`; `TestLargeSourceBudgets` three times; the AC-3.5 grep | Pass; pass; prints nothing | AC-2.6, AC-2.7, AC-3.4, AC-3.5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestDerive`, `TestDeriveManyDuplicateParentSteps`, `TestDeriveInvariants` (D-4), `TestFor`, `TestForFollowsCommits`, `TestForMatchesDerive`, `TestRelationships` (D-7, D-9), the `matchregistry` guard (`siblings.go` in a numbers package), `TestVocabKeysMatchSeed` (D-10), `TestProposeKinshipShapes` and the `graphalign` goldens (AC-3.4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks); on a host without Xcode, every other check PASS and xcode-test reported not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it) — `forResult` and the add-match-rule clause (STEP-7 grep)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — no brief
- [ ] Domain doc updated: add-match-rule and docs/ideas/kinship-view.md (D-13); matching-model.md unchanged (no knob)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: no view reads siblings until S13-12, and S13-04's handler tests carry the first end-to-end proof. The reviewer checks matching instead:

1. Open a project whose tree files a parent with two children. Promote a Source that records those two children as siblings. The Promote rows read as they do on `main` (T-8 carries the proof).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| D-4 or D-9 moves a chain, a golden or `For`'s order | STEP-2 and STEP-4 change no behavior and run before any new code, with no test changed |
| A cold sibling load costs one batch per child or per person | D-8 batches; T-5 compares counts at two family sizes and at three people |
| The stored answer outlives an edit two hops away | `DropOnHandleCommit` (D-11); T-4; verify mode (T-9) |
| The check and the read disagree on a missing parent | Both read nodes through `g.Cached` after the loads (D-8); AC-2.7 under verify |

**Stop and report if:**

- `chains`, `Derive`'s step loop, `deriver.chainLinks`, `deriver.forLinks`, `Reader.For`, `forSlot` or `kinship.LinkSet` no longer reads as §4 says.
- STEP-2 or STEP-4 needs an existing test changed to pass (beyond `hasMemo`'s new argument).
- `TestLargeSourceBudgets` fails at STEP-6 where it passes on `main` in the same run conditions (three runs each).
- The `matchregistry` guard rejects `core/kinship/siblings.go` *(new)*.
- `segmentcompare.Keys(rel, vocabkey.TermSibling, "", terms)` returns two different keys, so a recorded and an implied sibling do not merge in AC-1.4.
- T-5's counts differ with D-8's loads in place.
- T-8's two proposals differ.

## 12. Open questions

None.

## 13. PR

**Title:** S13-03 — Siblings from a shared parent
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-03-siblings-from-a-shared-parent.md`.
