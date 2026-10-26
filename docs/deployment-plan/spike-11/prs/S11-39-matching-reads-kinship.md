# S11-39 — Matching reads kinship shapes

**Status:** Blocked — on S11-49, S11-50 and a re-plan (slice 9 re-cut, 2026-10-26; see the map's S11-39 scope entry and #567's ship log)
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Promote's rows and the evidence sheet's relationship Card are unchanged; only the scores and lines behind them move)
**Board:** —
**Depends on:** S11-38 (landed [#564](https://github.com/mendahu/provenencia/pull/564))
**Blocks:** S11-40
**Branch:** `s11-39-matching-reads-kinship`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (STEP-8 edits it)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape), [§ Counting once](../kinship-shapes.md#counting-once), [§ Weight](../kinship-shapes.md#weight); [layer 6 procedure](../comparison-layers.md#layer-6-procedure); S11-14 D-11 (link weights); S11-38 D-1 (`kinship.Edge`); S11-41 (landed [#561](https://github.com/mendahu/provenencia/pull/561), `segmentcompare.Keys`); working rules: "Implied links, not a new comparison", "One fact, one link per side", "Counted once across sides", "No new knob", "High tolerance for churn", "Sparse data first"
**Planned:** 2026-10-26 against `main` @ `7efa02f`
**Review:** [`S11-39-matching-reads-kinship.review.md`](S11-39-matching-reads-kinship.review.md) — Fits with changes / Ready with changes, 2 rounds (every finding fixed)

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

Promote now reads each side's links through the kinship patterns. A Source that records a marriage gets the same help toward the couple as one that records a spouse relationship; a birth that names the father matches a tree that says he is the parent; a parent chain matches a recorded grandparent; and each of these works the other way round. Both loaders share one same-ends, same-key merge, and `core/kinship`'s implied links go through it. Graph alignment learns one rule, kept in `core/kinship` so the evidence query can call it in S11-40: a fact both sides record through the same pattern counts once when the walk has paired the event (or middle parent) it passes through.

**Why now.** It is the migrate PR after S11-38 built `core/kinship` (build, then migrate). Slice 9's Check and dogfood bar 12 depend on it.

**What stays.** Every proposal on a Source with no marriage, birth father or mother, or parent chain (the obituary golden, the sparse cases, the large-source fixture). Bands, knobs, the wire, Swift, and the Promote page. A marriage recorded on both sides scores what it scores today (user, 2026-10-26). Implied links are never filed and never shown as connections.

## 2. User stories

Fixture names: Source people Ann Ames, Bob Ames, Bea Ames, Cal Ames (`f.person`); "sex only" is `f.sexOnly(src, art, "male")`; the tree has handles for the same names, filed on `f.source` with every Subject promoted, events included. Events are undated (`f.eventOn(src, art, type, nil)`). "Bob's row" is `rowFor(f.propose(graph, fixed), bob.ID)`; `fixed` is empty unless the AC says Ann is fixed on her handle. Each Then names the sibling row the test compares with (D-11); the number in brackets is the planning measurement and appears in tests only through §11's stop conditions.

**US-1.** As a researcher promoting a Source that records a kinship fact in another shape than my tree, I get the same help toward the match as when both record it the same way, so a marriage, a birth's parents or a parent chain finds the people my tree has as spouses, parents or grandparents.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Tree: Ann and Bob with a filed `spouse` relationship. Source: Ann and Bob each `spouse` in one `marriage` | Bob's row is on Bob's handle, band medium, score equal to the same tree against a Source recording a `spouse` relationship (3.375), and above a Source whose two people are `witness` in the marriage (2.5, weak) |
| AC-1.2 | Tree: Ann and Bob each `spouse` in one filed `marriage`. Source: a `spouse` relationship | Bob's row is on Bob's handle, band medium, score equal to AC-1.1's direct-against-direct row (3.375) |
| AC-1.3 | Tree: Ann `parent` of Bob. Source: Ann `father`, Bob (sex only) `subject` of one `birth`; Ann fixed | Bob's row's first alternative is Bob's handle, reason `via`, score equal to the same tree against a Source recording Ann `parent` of Bob (sex only) (0.875) |
| AC-1.4 | Tree: Ann `father`, Bob `subject` of a filed `birth`. Source: Ann `parent` of Bob (sex only); Ann fixed | As AC-1.3, compared with a tree recording `parent` |
| AC-1.5 | Tree: Ann `grandparent` of Cal. Source: Ann `parent` of Bea, Bea `parent` of Cal; Ann fixed | Cal's row is on Cal's handle, reason `via`, `Via.NeighborSubjectID` = Ann's Source Subject, `Via.Key` = `vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "grandchild", End: vocabkey.PropertyRelatedTo}` (prints `relationship·grandchild·related_to`), band medium, score equal to the same tree against a Source recording Ann `grandparent` of Cal (3.375) |
| AC-1.6 | Tree: Ann `parent` of Bea, Bea `parent` of Cal. Source: Ann `grandparent` of Cal; Ann fixed | As AC-1.5, compared with a tree recording `grandparent` |
| AC-1.7 | A row reached through an implied link (AC-1.5) | The page words it "via Ann Ames → PER-…" with no relationship phrase, because the Evidence graph has no bridge between Ann and Cal (`PromoteModel.reasonText` falls back to `L10n.Promote.viaNeighbor`) |

**US-2.** As a researcher, a fact counts once, so a Source or tree that records it twice, or a fact both record in the same shape, does not inflate a match.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Tree: `spouse` relationship. Source: a `spouse` relationship and a `marriage` with both as `spouse` | `LayerBridges` holds one `relationship·spouse` bridge between Ann and Bob: 3 `Associations` (the relationship bridge and both participation bridges), one `Shapes` entry, `OnlyImplied` false. Bob's row's score equals AC-1.1's direct-against-direct row (3.375) |
| AC-2.2 | Tree and Source each record one `marriage` with Ann and Bob as `spouse` | Bob's row is on Bob's handle, band weak, score below AC-1.1's direct-against-direct row and above AC-1.1's witness control (2.5); planning measured 2.875: the marriage events are paired, so the implied spouse link adds nothing (T-2 pins the exact equality with no implied links) |
| AC-2.3 | As AC-2.2 but Bob is sex only and Ann is fixed, so the marriage event row has too little to be placed | Bob's row's first alternative is Bob's handle, reason `via`, score equal to AC-1.3's shape for spouses: the same tree against a Source recording a `spouse` relationship to a sex-only Bob (0.875). The events are not paired, so the implied link credits |
| AC-2.4 | Tree and Source each record Ann `parent` Bea `parent` Cal; Ann fixed | Cal's row is on Cal's handle, band medium, score equal to AC-1.5's comparison row (3.375): Bea is paired, so the implied grandparent link adds nothing beyond the parent path |
| AC-2.5 | A Source with both a `spouse` relationship and a `marriage`, against a tree with only the `marriage` (D-4) | Bob's row's score equals AC-2.2's (2.875): the tree records the fact once, through the paired marriage |

**US-3.** As a researcher opening the evidence sheet for a row whose Source recorded a marriage, I see the spouse match and can pin it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Tree: filed `spouse` relationship. Source: a `marriage` with both as `spouse`. `Evidence` for Bob on his handle with Ann placed on hers | One relationship, `Key.String()` = `relationship·spouse`, `Neighbor` = Ann's Source Subject, `Checked` true, `Incoming` holds every Observation of both participation bridges (6), `Members` holds one bundle (the filed spouse bridge) |
| AC-3.2 | `promotealign.PinOffers` of AC-3.1's evidence | The bundle's offer is `Checked`, and its `Set.Observations` starts with AC-3.1's `Incoming` |
| AC-3.3 | Tree: filed `marriage`. Source: `spouse` relationship. `Evidence` for Bob as in AC-3.1 | No relationship (the canon side derives in S11-40) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `kinship.Shape`, `Implied.Shape()`, `kinship.CountedOnce` | STEP-1, T-1 |
| `graphalign.Bridge` and `CanonEdge` carry `Shapes` and `OnlyImplied`; `corresponding` skips a link pair `CountedOnce` | STEP-2, AC-2.2, AC-2.4, AC-2.5 |
| One link set (`linkSet`) is both loaders' same-ends, same-key merge, and adds implied links through it; the layer loader uses it | STEP-3, AC-1.1, AC-1.3, AC-2.1 |
| `loadCanon` uses it | STEP-4, AC-1.2, AC-1.4, AC-1.6 |
| Goldens: marriage ↔ spouse, birth parents ↔ parent, chain ↔ grandparent, each both ways; counted once; via | STEP-5, US-1, US-2 |
| The evidence query reads Source-side implied links as it reads any layer link | STEP-6, US-3 |
| Derivation cost measured | STEP-7 |
| `add-match-rule` says how callers merge and what the count-once rule reads | STEP-8 |
| `matching-model.md`, `promote-graph-alignment.md` say links are read through the kinship patterns | STEP-9 |

| Out | Home |
| --- | --- |
| The evidence query's canon side (`canonLinks`) deriving through a `linkSet`; a line saying how each side recorded the fact; bundles of an implied link | S11-40 |
| The count-once rule in the evidence sheet's relationship lines (a call to `kinship.CountedOnce`) | S11-40 (map scope entry names it) |
| The sheet's wording of implied lines | S11-42 |
| The count-once rule in the duplicate check `newPairSegments` (`core/graphalign/rows.go`). Implied Source links do reach it through `layerAdj` and credit there like recorded ones; accepted | not planned — two Source people sharing one marriage's shapes is entry-error data; revisit if dogfood shows it |
| Memoizing the canon's implied links | not planned — measured at 31 ms for 8,190 parent links (STEP-7); revisit if a budget fails |
| A Swift change to the via wording | not planned — AC-1.7 already holds (`reasonsAreWordedWithLabelsAndRefs`) |
| New knobs, other patterns, trees and person pages | working rule "No new knob"; [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Pure derivation | `core/kinship/kinship.go` `Edge`, `Implied`, `Derive`; `core/kinship/patterns.go` `eventPatterns` | Returns implied links with `Pattern`, `Event` (marriage or birth) and `Through` (the middle parent). Returned slices share the input's byte slices. No caller. `Edge` has `Bridge`'s fields in its order (S11-38 D-1) |
| Layer edges | `core/graphalign/types.go` `Bridge{A, B, KeyA, KeyB, Associations}` | No shape or implied mark. `Layer`'s comment: "one Source's primary Subjects and bridges" |
| Canon edges | `core/graphalign/types.go` `CanonEdge{From, To, KeyFrom, KeyTo}` | Comment: "one filed association hop between handles" |
| Adjacency | `core/graphalign/align.go` `adjLink{neighbor, key}`; `newState` fills `layerAdj` / `canonAdj` | Built from every bridge and edge |
| Correspondence | `core/graphalign/align.go` `(st *state) corresponding` | Canon links of a handle whose key `segmentcompare.Corresponds` to a layer link's. The one choke point: `place` (nomination), `segments` (credit) and `supports` (via) call it, and nothing else reads correspondence |
| Layer loader | `core/database/promotealign/layer.go` `loadLayerBridges` | Walks every `connectrules` bridge, keys each edge with `segmentcompare.Keys`, merges same-ends, same-key edges inline through `seen map[seenEdge]int` / `edgeSeenKey`, cloning every end and association. Memoized with the layer (`loadLayerMemo`) and shared by every caller |
| Canon loader | `core/database/promotealign/canon.go` `loadCanon` | Walks the whole connected component from the seeds; dedupes edges through `edgeSeen map[seenEdge]bool`, cloning ends. Runs on every proposal |
| Merge key | `core/database/promotealign/terms.go` `seenEdge`, `edgeSeenKey` | Ends in a fixed order with their keys |
| Evidence query | `core/database/promotealign/evidence.go` `evidence` | Builds `layerLinks` from `in.layer.Bridges` (every bridge with the row Subject at one end, Observations of its `Associations`); canon side from `canonLinks` (graph nodes, not `Canon.Edges`) |
| Readers of `Layer.Bridges` | `core/graphalign/align.go` `newState`; `core/database/promotealign/evidence.go` `evidence` | Nothing else: Apply and the subrows read the Evidence graph snapshot (`PromoteModel.bridgeFacts`) |
| Via wording | `macos/App/Features/Promote/PromoteModel.swift` `reasonText`; `PromoteFlow.bridge(between:and:)` | Phrase from the Evidence graph bridge between the two rows, else `L10n.Promote.viaNeighbor`. `viaBridgeType` / `viaRole` are mapped (`GoStore.swift`) and not read |
| Skill | `.agents/skills/add-match-rule/SKILL.md` "Kinship patterns (`core/kinship`) add implied links before alignment" bullet | Says `Edge` has "the fields of `graphalign.Bridge`" and that callers merge through their own merge |
| Tests | `core/kinship/kinship_test.go` `TestDerive`, `TestDeriveInvariants`, `BenchmarkDerive` (generations 8, 9, 10); `core/graphalign/invariants_test.go` `TestAlignInvariants`; `core/graphalign/align_test.go` helpers `id`, `bridge`, `edge`, `partLink`, `relLink`, `termProp`, `vals`, `nameVals`, `nameMetas`, `eventMetas`, `mergeMetas`, `rowBySubject`, `alignWith`; `core/database/promotealign/promotealign_test.go` fixture (`person`, `sexOnly`, `eventOn`, `participation`, `relationship`, `promote`, `newSource`, `graph`, `propose`, `rowFor`), `TestProposeChildOfReadsAsParentOf`, `TestLayerBridgesCoverEveryConnectRule`; `core/database/promotealign/evidence_test.go` `f.evidence`; `core/database/promotealign/pins_test.go`; `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`; `core/database/promotealign/export_test.go` `LayerBridges`; `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` | No test has a marriage, a birth father or mother, or a chain on both sides of a proposal, so no existing test changes outcome when either side derives |

**Precedent.** S11-41 ([#561](https://github.com/mendahu/provenencia/pull/561)) gave both loaders one shared link reading (`segmentcompare.Keys`) so a third caller would not copy it; `linkSet` is the same move for the merge and derivation. `TestProposeChildOfReadsAsParentOf` is the shape of each golden (a filed tree on `f.source`, a second Source, `propose`, `rowFor`). `TestAlignInvariants`'s fixture table takes STEP-2's new row.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/kinship` gains `type Shape struct { Pattern Pattern; Pivot []byte }` *(new)*, one way a link is recorded through a pattern. Its doc comment says `Pivot` is `Implied.Event` for an event pattern and `Implied.Through` for a chain. `func (i Implied) Shape() Shape` *(new)* returns `{i.Pattern, i.Event}` when `Event` is set, else `{i.Pattern, i.Through}`. `Edge`, `Implied` and `Derive` are otherwise unchanged | The count-once rule needs the pattern and its pivot on each link; the shape is kinship's concept, and S11-40 reads it too | Fields on `Edge` that `Derive` fills — every `TestDerive` row changes for data `Implied` already has; a `graphalign` shape type — the concept would live in the layer that only reads it |
| D-2 | `func CountedOnce(a, b []Shape, aOnlyImplied, bOnlyImplied bool, paired func(pivot []byte) ([]byte, bool)) bool` *(new)* in `core/kinship`: false when neither side is only implied; else true when some shape of `a` and some shape of `b` have the same `Pattern` and `paired(aShape.Pivot)` returns `bShape.Pivot` (`bytes.Equal`). `a` is the side whose pivots `paired` maps (the layer) | One pure rule for every caller: graph alignment now, the evidence query's relationship lines in S11-40 (which holds its own `placed` map in `evidence`). User, 2026-10-26: count once only when the shape is actually paired, so a sparse marriage whose event row is not placed still gains (AC-2.3) | A private `graphalign` method — S11-40 would copy it; a static rule (skip whenever both share a pattern) — loses AC-2.3's credit; the stronger of the two credits — needs segment results across links (user declined) |
| D-3 | `graphalign.Bridge` and `CanonEdge` gain, after their existing fields, `Shapes []kinship.Shape` (the patterns that imply the link; none for a link only recorded) and `OnlyImplied bool` (no recorded bridge stands behind the link). `adjLink` gains `shapes` and `onlyImplied`, filled in `newState`. `corresponding` keeps a canon link only when `Corresponds` holds and `kinship.CountedOnce(link.shapes, ce.shapes, link.onlyImplied, ce.onlyImplied, func(p []byte) ([]byte, bool) { h, ok := st.placed[string(p)]; return h, ok })` is false. `graphalign` imports `core/kinship`. Nothing else in `graphalign` changes | A merged link can be both recorded and implied (AC-2.1), so "has shapes" is not "only implied". `OnlyImplied`'s zero value is right for every recorded bridge, so no existing construction site changes; the name keeps "implied" from meaning two things beside `kinship.Implied`. `corresponding` is the one choke point, so nomination, credit and via agree. Placing or releasing a pivot already marks its layer neighbors dirty (`place`, `release`), so settling re-scores them | `Direct bool` — every bridge built today would need it set; `Implied bool` — clashes with `kinship.Implied`, which a merged recorded link also is |
| D-4 | The rule is exactly D-2's: two links skip only when one of them is only implied. A Source with a recorded spouse relationship and a marriage, against a tree with only the marriage, skips (AC-2.5): the tree records the fact once, through the paired marriage. Two links both recorded (and implied) keep today's double credit | Today, recording the marriage and the relationship on both sides credits both; this PR changes only what implied links add | Skip only when both are only implied — AC-2.5 would credit the tree's one marriage twice |
| D-5 | One merge in `core/database/promotealign/terms.go` beside `edgeSeenKey`: `type sideLink struct { kinship.Edge; Shapes []kinship.Shape; OnlyImplied bool }` *(new)*, `type linkSet struct { links []sideLink; at map[seenEdge]int }` *(new)* with `func newLinkSet() *linkSet` and `func (s *linkSet) add(l sideLink)` *(new)*, and `func (s *linkSet) withImplied(terms segmentcompare.Terms)` *(new)*. `add` keys `l` by `edgeSeenKey`; on a hit it appends each association the link lacks (`bytes.Clone`), appends `l.Shapes`, and sets `OnlyImplied` to `existing.OnlyImplied && l.OnlyImplied`; on a miss it appends `l` with `A`, `B` and every association cloned, and registers its index. `withImplied` calls `kinship.Derive` over the `Edge`s of `s.links` as they stand, then `add`s each `imp` as `sideLink{Edge: imp.Edge, Shapes: []kinship.Shape{imp.Shape()}, OnlyImplied: true}` | Working rule "One fact, one link per side": implied links go through the loaders' existing merge, not a second one. One merge for recorded and implied links on both sides, so cloning, association order and shape rules cannot drift, and S11-40's `canonLinks` uses the same set. Registering a miss merges two implied links of one fact (a couple with two marriages). The layer is memoized and shared, so every end and association is cloned | A separate helper for implied links beside the inline merges — two merges in `promotealign` (round-2 M-7); a merge per loader — copies that already disagreed on registering a miss (round-1 M-2); converting `kinship.Edge(b)` — `Bridge` no longer has `Edge`'s fields |
| D-6 | `loadLayerBridges` keeps its walk and keys; its `seen` map and inline merge go, and each walked edge is `set.add(sideLink{Edge: kinship.Edge{A: e.From, B: e.To, KeyA: keyA, KeyB: keyB, Associations: [][]byte{e.Association}}})` (`add` clones). After the walk, `set.withImplied(terms)`, then one conversion of `set.links` to `[]graphalign.Bridge{A, B, KeyA, KeyB, Associations, Shapes, OnlyImplied}` | Recorded links are all added first, so a recorded link is never merged into an implied one, and order stays the walk's | Derive per bridge kind — `Derive` needs every link at once |
| D-7 | `loadCanon` keeps its walk; its `edgeSeen` map goes, and each keyed link is `set.add(sideLink{Edge: kinship.Edge{A: from, B: to, KeyA: keyFrom, KeyB: keyTo}})`. After the walk and before `prof.profile`, `set.withImplied(terms)`, then one conversion of `set.links` to `[]graphalign.CanonEdge{From: A, To: B, KeyFrom: KeyA, KeyTo: KeyB, Shapes, OnlyImplied}` (associations dropped: canon implied links have none, and recorded ones never had any) | `loadCanon` already walks the whole component, so a grandparent needs no extra reach (map review K-6). Implied edges after recorded ones keep the walk's link order | Derive per wave — a chain can span waves |
| D-8 | No memo for the canon's implied links | `BenchmarkDerive` at 8,190 parent links runs in 31 ms; the large-source budgets moved within noise in the planning prototype | A memo on the shortlist index generation — more state for a cost under the budgets |
| D-9 | The evidence query is unchanged: Source-side implied links reach it through `in.layer.Bridges`, and a Source marriage against a filed spouse relationship gives today's spouse line with both participations' Observations as `Incoming`, checked and pinned at Done (user, 2026-10-26) | Today's sentence holds until S11-42 anyway; a hold-back path is what the churn rule forbids | Filter implied bridges out of `layerLinks` until S11-40 |
| D-10 | No Swift, proto or FFI change. AC-1.7 is held by the existing `reasonsAreWordedWithLabelsAndRefs` (a via with no bridge between the rows); Go pins `Row.Via` (AC-1.5) | `viaRole` is not read; the phrase comes from the Evidence graph, which has no implied bridge | A Swift test duplicating the existing one |
| D-11 | Goldens live in `core/database/promotealign/promotealign_test.go` as `TestProposeKinshipShapes` *(new)*: subtests, one per AC, each building its own tree and Sources with one `newFixture`, and comparing the row under test with a sibling row from a Source built in the same subtest (exact `Score` equality, band, reason). The count-once rule's exact equalities with no implied link live in pure `graphalign` tests (T-2), where links are built by hand | Derivation happens in the loaders, so end-to-end goldens need the catalog fixture; a knob change elsewhere moves sibling rows together, so the goldens stay about shape equality | Hard-coded scores in every row |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `kinship.Shape` *(new)*, `Implied.Shape()` *(new)* | type, method | `kinship.Implied` `Pattern`, `Event`, `Through` | Extends it: the method reads those fields; a link may carry several shapes, an `Implied` is one derivation |
| `kinship.CountedOnce` *(new)* | func | `segmentcompare.Corresponds` | `Corresponds` is keys only; the rule needs shapes and a pairing, and two layers call it |
| `Bridge.Shapes`, `Bridge.OnlyImplied`, `CanonEdge.Shapes`, `CanonEdge.OnlyImplied` *(new)* | fields | `Bridge.Associations` (loader-set data the layers carry) | Extends the existing edge types; no second edge list |
| `adjLink.shapes`, `adjLink.onlyImplied` *(new)* | fields | `adjLink` | Extends it |
| `sideLink`, `linkSet`, `newLinkSet`, `linkSet.add`, `linkSet.withImplied` *(new)* | types, funcs | the inline merge in `loadLayerBridges` (`seen`), `loadCanon`'s `edgeSeen`; `edgeSeenKey` | It is those two merges made one, on the same `edgeSeenKey`; both inline merges are deleted |
| `TestImpliedShape`, `TestCountedOnce`, `TestKinshipShapesCountOnce`, `TestWithImplied`, `TestProposeKinshipShapes`, `TestEvidenceReadsASourceMarriageAsSpouses`, `TestPinOffersOfAnImpliedLayerLink` *(new)*; `implied`, `canonOf` *(new)* test helpers | tests, helpers | `TestProposeChildOfReadsAsParentOf`, `TestAlignInvariants`, `bridge`, `edge` | New behavior; the helpers compose with `bridge` and `edge` |

Searched: `rg -n "Associations|edgeSeen|corresponding\(" core/graphalign core/database/promotealign`, `rg -n "kinship" core`, `rg -n "placed :=" core/database/promotealign`.

**Builds on.** `kinship.Derive` in `core/kinship/kinship.go` (S11-38), `edgeSeenKey`, `corresponding` as layer 6's one correspondence path, the `add-match-rule` skill.
**Deletes.** The inline merge and `seen` map in `loadLayerBridges`, and `edgeSeen` in `loadCanon`, replaced by `linkSet.add`. The skill's "the fields of `graphalign.Bridge`" claim is rewritten (STEP-8).
**Next adds.** S11-40: `canonLinks` builds a `linkSet` and calls `withImplied`, reads `Implied.Shape()` and `Through` for the line, and calls `kinship.CountedOnce` with its `placed` map for each relationship line. A new event pattern (a baptism's parents) is an `eventPatterns` row; `linkSet` and `CountedOnce` take it with no change.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `Shape`, `Implied.Shape()`, `CountedOnce` *(new)* (D-1, D-2) |
| `core/graphalign` | `types.go`: `Bridge`, `CanonEdge` gain `Shapes`, `OnlyImplied`; comments of `Layer` ("bridges, recorded or implied by the kinship patterns") and `CanonEdge` ("one filed association hop, or a link the kinship patterns imply, between handles") updated. `align.go`: `adjLink` gains `shapes`, `onlyImplied` and its comment names them; `newState` copies them; `corresponding` calls `CountedOnce` (D-3) |
| `core/database/promotealign` | `terms.go`: `sideLink`, `linkSet`, `add`, `withImplied` (D-5). `layer.go`: D-6. `canon.go`: D-7 |

```go
// CountedOnce reports whether two corresponding links stand for one fact both
// sides record through the same kinship pattern, already counted because the
// shape's pivots are paired: paired maps a pivot of a to its partner on b's
// side. Two links that are both recorded never are.
func CountedOnce(a, b []Shape, aOnlyImplied, bOnlyImplied bool, paired func(pivot []byte) ([]byte, bool)) bool

// add merges l into the link with the same ends and keys, or appends it.
func (s *linkSet) add(l sideLink)

// withImplied adds the links the kinship patterns imply to the set, through add.
func (s *linkSet) withImplied(terms segmentcompare.Terms)
```

Errors: none new.

### FFI

None — no wire change (D-10).

### macOS

None — no Swift change (D-10).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Shapes and the count-once rule in `core/kinship`**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`. Skill: `add-match-rule`.
Do: add `Shape`, `Implied.Shape()` and `CountedOnce` (D-1, D-2) beside `Implied`, with doc comments. Add `TestImpliedShape` (T-1) and `TestCountedOnce` (T-2a). Extend `BenchmarkDerive`'s `generations` list to `8, 9, 10, 11, 12`.
Check: `go test ./core/kinship/... ./core/matchregistry/...` passes.

**STEP-2 — The count-once rule in `graphalign`**
Files: `core/graphalign/types.go`, `core/graphalign/align.go`, `core/graphalign/align_test.go`, `core/graphalign/invariants_test.go`. Skill: `add-match-rule`.
Do: D-3, with the comments §6 names. In `align_test.go` add `implied(b graphalign.Bridge, p kinship.Pattern, pivot []byte, onlyImplied bool) graphalign.Bridge` *(new)* (sets `Shapes` to that one shape and `OnlyImplied`) and `canonOf(bs []graphalign.Bridge) []graphalign.CanonEdge` *(new)* (field by field, `Associations` dropped). Add the fixture builder `marriageFixture(layerShape, canonShape string, dated bool) (graphalign.Layer, graphalign.Canon)` *(new)* (shapes `rel`, `event`, `marriage`, `both`, as T-2 defines them), then `TestKinshipShapesCountOnce` (T-2) over it, then the `TestAlignInvariants` row (T-3): `build: func() (graphalign.Layer, graphalign.Canon) { return marriageFixture("marriage", "marriage", true) }`, `fixed: []graphalign.Fixed{{SubjectID: id("s-ann"), HandleID: id("h-ann")}}`.
Check: `go test ./core/graphalign/...` passes; every existing `graphalign` test passes unchanged.

**STEP-3 — One link set; the layer loader derives**
Files: `core/database/promotealign/terms.go`, `core/database/promotealign/layer.go`, `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/export_test.go`.
Do: D-5, D-6. In `export_test.go` export `type SideLink = sideLink` and `func LinkSetForTest(links []SideLink, terms segmentcompare.Terms) []SideLink` (a new set, `add` each link, `withImplied`, return the links). Add `TestLinkSet` (T-4) and `TestLayerBridgesMergeKinshipShapes` (T-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/...` passes, every existing test unchanged (§4: none has a kinship shape).

**STEP-4 — `loadCanon` derives**
Files: `core/database/promotealign/canon.go`.
Do: D-7.
Check: as STEP-3.

**STEP-5 — Goldens**
Files: `core/database/promotealign/promotealign_test.go`.
Do: add `TestProposeKinshipShapes` (T-6) as D-11 says. Shape builders are local closures `func(src sources.Source, art artifacts.Artifact, a, b subjects.Subject) []subjects.Subject` that record a shape and return the extra Subjects (events, Bea) the tree setup promotes.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestProposeKinshipShapes -v` passes.

**STEP-6 — Evidence and pins**
Files: `core/database/promotealign/evidence_test.go`, `core/database/promotealign/pins_test.go`.
Do: add `TestEvidenceReadsASourceMarriageAsSpouses` (T-7) and `TestPinOffersOfAnImpliedLayerLink` (T-8). No production change (D-9).
Check: both pass.

**STEP-7 — Measure**
Files: none beyond STEP-1's benchmark.
Do: `git worktree add ../s11-39-base 7efa02f`, and in it and in the branch run `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -count=1 -v`. On the branch run `go test -run x -bench Derive -benchmem ./core/kinship/`. Put both runs' six budget lines and the five benchmark lines in the PR body, then `git worktree remove ../s11-39-base`.
Check: every budget passes on the branch; `generations=12` is under 100 ms/op.

**STEP-8 — Skill**
Files: `.agents/skills/add-match-rule/SKILL.md`.
Do: rewrite the "Kinship patterns (`core/kinship`) add implied links before alignment" bullet to say: `Edge` holds a link's ends, keys and bridge Subjects; `graphalign.Bridge` and `CanonEdge` add `Shapes` and `OnlyImplied`, so loaders copy fields by name and never convert; both loaders merge every link, recorded or implied, through `promotealign`'s `linkSet.add`, the one merge, and add implied links with `linkSet.withImplied`; layer 6's `corresponding` skips a pair `kinship.CountedOnce` reports, and a caller outside layer 6 calls `CountedOnce` with its own pairing; a new pattern needs no change to either.
Check: `rg -n "linkSet|withImplied|CountedOnce|OnlyImplied" .agents/skills/add-match-rule/SKILL.md` prints the bullet; `rg -n "the fields of .graphalign.Bridge." .agents/skills/add-match-rule/SKILL.md` prints nothing.

**STEP-9 — Docs**
Files: `docs/matching-model.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md`, this plan.
Do:
- `matching-model.md` "The stack", after the paragraph that starts "The subject profile builder sits beside layer 4": "Before layer 6, each side's links are read through the kinship patterns (`core/kinship`): a shared marriage implies spouses, a birth's father or mother and subject imply a parent, and a parent of a parent implies a grandparent. An implied link weighs as the relationship it stands for. A fact both sides record through the same pattern counts once when the walk has paired the event or middle parent it passes through."
- Glossary rows after **segment**: **kinship pattern** | "A shape that records a kinship fact without naming it: two spouses in one marriage, a birth's father or mother and subject, a parent of a parent"; **implied link** | "A relationship link a kinship pattern adds to one side, keyed as a recorded relationship; never stored or filed"; **shape** | "One way a side records a link: a kinship pattern and the event or middle parent it passes through".
- `promote-graph-alignment.md` §12, after the "Bridge kinds in the loaders" bullet: "- **Kinship shapes:** both loaders add the links `core/kinship` implies (spouse, parent, grandparent) through their one merge (`linkSet`). Implied links are never filed and never shown as connections; a row reached through one says "via <neighbor>" with no relationship phrase. A link both sides imply through a paired event or middle parent counts once."
- `completed.md`, the map row and this plan: as the Definition of done says.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Match a Source marriage to the spouses a tree records."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/kinship_test.go` `TestImpliedShape` *(new)* | table: an `Implied` with `Event` E1 → `{PatternMarriageSpouses, E1}`; one with `Through` P2 and no `Event` → `{PatternParentOfParent, P2}` | D-1 |
| T-2a | Go | `core/kinship/kinship_test.go` `TestCountedOnce` *(new)* | table, `paired` maps E1 → E9: shared marriage pattern, pivots E1 / E9, a only implied → true; same, b only implied → true; neither only implied → false; pivots E1 / E8 → false; `paired` knows nothing → false; patterns marriage / birth → false; empty shapes → false | D-2, D-4 |
| T-2 | Go | `core/graphalign/align_test.go` `TestKinshipShapesCountOnce` *(new)* | Fixture: layer Ann `s-ann`, Bob `s-bob` (`nameVals("Ann", "Ames")`, `nameVals("Bob", "Ames")`), marriage `s-marriage`; canon `h-ann`, `h-bob`, `h-marriage` with the same values; metas `mergeMetas(nameMetas(), eventMetas())`; the marriage's values `vals(termProp("event_type", "marriage"))`, plus `vocabkey.PropertyDate` = `propertycomparetest.Date(datevaluetest.YMD(1840, 5, 12))` when dated; `Canon.Candidates` names each handle for its Subject; Ann fixed on `h-ann`. A side's shape is one of: `rel` (one `bridge(ann, bob, relLink("spouse"))`, no marriage Subject or handle), `event` (two `partLink("spouse", "marriage")` bridges to the marriage), `marriage` (`event` plus `implied(bridge(ann, bob, relLink("spouse")), PatternMarriageSpouses, marriage, true)`), `both` (as `marriage` with `onlyImplied` false). Bob's row's `Score`, undated: (rel, rel) = (marriage, rel) = (rel, marriage); (marriage, marriage) = (both, marriage) = (marriage, both) = (event, event), each below (rel, rel); (marriage, marriage) with `Fixed{SubjectID: s-marriage, Target: TargetNew}` added = (rel, rel); (both, both) above (rel, rel). Dated: (marriage, marriage) = (event, event) | AC-2.2, AC-2.3, AC-2.5, D-3, D-4 |
| T-3 | Go | `core/graphalign/invariants_test.go` `TestAlignInvariants` row `"marriage on both sides"` | T-2's fixture, (marriage, marriage), dated, Ann fixed: the existing invariants hold | D-3 |
| T-4 | Go | `core/database/promotealign/promotealign_test.go` `TestLinkSet` *(new)* | table over hand-built `SideLink`s through `LinkSetForTest` and `segmentcomparetest.Terms()`: the same recorded link added from each end → one link with both associations; a recorded spouse link plus a marriage → one spouse link, 3 associations, one shape, `OnlyImplied` false; two marriages of one couple → one implied spouse link with two shapes and 4 associations; a returned implied link's `A` does not alias the input's | D-5 |
| T-5 | Go | `core/database/promotealign/promotealign_test.go` `TestLayerBridgesMergeKinshipShapes` *(new)* | AC-2.1's Source → as AC-2.1; marriage only → the spouse bridge has 2 `Associations`, `OnlyImplied` true | AC-2.1, D-6 |
| T-6 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeKinshipShapes` *(new)* | one subtest each for AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, asserting the Then column against its sibling row | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| T-7 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses` *(new)* | AC-3.1 and AC-3.3 | AC-3.1, AC-3.3, D-9 |
| T-8 | Go | `core/database/promotealign/pins_test.go` `TestPinOffersOfAnImpliedLayerLink` *(new)* | AC-3.2 | AC-3.2 |
| T-9 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` (existing, unchanged) | a via row with no bridge between it and its neighbor → `L10n.Promote.viaNeighbor` | AC-1.7, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs (its Xcode part runs on macOS only; on another machine run every Go check and say so). Most likely to break: `TestAlignInvariants`, `TestGracieObituaryGolden`, `TestSparseNameAndBirthDate`, `TestLargeSourceBudgets` (budgets and "memo matches a fresh run"), `TestLayerBridgesCoverEveryConnectRule`, `TestDerive`, the purity and numbers guard in `core/matchregistry/guard_test.go`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md and promote-graph-alignment.md (STEP-9); add-match-rule (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check, STEP-7's numbers
```

Map row S11-39: Status `Landed #N`; the Depends on column stays `S11-38`. `completed.md`: a new section at the top, `## S11-39 — Matching reads kinship shapes`, `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).`, then two or three sentences: Promote reads a marriage, a birth's parents and a parent chain as spouse, parent and grandparent links on both sides; a fact both sides record through the same paired shape counts once; a Source marriage against a filed spouse relationship shows the spouse line on the evidence sheet.

## 10. Manual check

1. In a project, file Source 1 with Ann Ames and Bob Ames and a `spouse` relationship between them; promote both.
2. Create Source 2 with Ann Ames, Bob Ames and a `marriage` event in which both are `spouse`. Open Promote on it. Bob's row proposes Bob's handle at medium, as a Source 2 with a spouse relationship instead does.
3. Open Bob's evidence sheet with Ann on her handle: one spouse relationship line, checked.
4. File Ann Ames `grandparent` of Cal Ames on Source 3 and promote both. Create Source 4 with Ann Ames, Bea Ames and Cal Ames, Ann `parent` of Bea and Bea `parent` of Cal. Set Ann's row to her handle. Cal's row says "via Ann Ames → PER-…".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rule makes a score depend on placement order | `corresponding` is already placement-dependent (`segments`); placing or releasing a pivot marks its layer neighbors dirty, and T-3 checks the invariants |
| A whole-family canon makes every proposal slower | D-8's measurement; budgets in `TestLargeSourceBudgets` |
| A pivot paired wrongly removes a true credit | The same wrong pairing already misleads the event path; the row is contested when the events conflict |

**Stop and report if:**

- `corresponding` is no longer the only place `place`, `segments` and `supports` read correspondence.
- A §2 Then fails, or the measured value differs from the bracketed planning measurement: AC-1.1 3.375 medium, witness control 2.5, AC-1.3 0.875, AC-2.2 2.875 weak, AC-1.5 3.375 medium (a knob moved or the plan is wrong: report, do not retune).
- `TestAlignInvariants` fails on the new row.
- An existing `graphalign` or `promotealign` test changes outcome.
- A budget in `TestLargeSourceBudgets` fails, or `BenchmarkDerive/generations=12` is 100 ms/op or more.
- `graphalign` importing `core/kinship` fails the guard in `core/matchregistry/guard_test.go` or makes an import cycle.

## 12. Open questions

None.

## 13. PR

**Title:** S11-39 — Matching reads kinship shapes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; STEP-7's numbers; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.md`.
