# S11-39 — Matching reads kinship shapes

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Promote's rows and the evidence sheet's relationship Card are unchanged; only the scores and lines behind them move)
**Board:** —
**Depends on:** S11-38 (landed [#564](https://github.com/mendahu/provenencia/pull/564)), S11-41 (landed [#561](https://github.com/mendahu/provenencia/pull/561))
**Blocks:** S11-40
**Branch:** `s11-39-matching-reads-kinship`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (STEP-8 edits it)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape), [§ Counting once](../kinship-shapes.md#counting-once), [§ Weight](../kinship-shapes.md#weight); [layer 6 procedure](../comparison-layers.md#layer-6-procedure); S11-14 D-11 (link weights); S11-38 D-1 (`kinship.Edge`); working rules: "Implied links, not a new comparison", "One fact, one link per side", "Counted once across sides", "No new knob", "High tolerance for churn", "Sparse data first"
**Planned:** 2026-10-26 against `main` @ `7efa02f`
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

Promote now reads each side's links through the kinship patterns. A Source that records a marriage gets the same help toward the couple as one that records a spouse relationship; a birth that names the father matches a tree that says he is the parent; a parent chain matches a recorded grandparent; and each of these works the other way round. Both loaders add `core/kinship`'s implied links through their existing same-ends, same-key merge. Graph alignment learns one rule so a fact both sides record in the same shape is credited once: an implied link adds nothing when the event (or middle parent) it passes through is already paired with the other side's.

**Why now.** It is the migrate PR after S11-38 built `core/kinship` (build, then migrate). Slice 9's Check and dogfood bar 12 depend on it.

**What stays.** Every proposal on a Source with no marriage, birth father or mother, or parent chain (the obituary golden, the sparse cases, the large-source fixture). Bands, knobs, the wire, Swift, and the Promote page. A marriage recorded on both sides scores what it scores today (user, 2026-10-26). Implied links are never filed and never shown as connections.

## 2. User stories

Fixture names: Source people Ann Ames, Bob Ames, Bea Ames, Cal Ames; the tree has handles for the same names, filed from another Source. "Bob's row" is `rowFor(propose(...), bob)` with no fixed rows unless the AC says Ann is fixed on her handle. Reference scores are from the planning prototype at `7efa02f`; tests assert the relations in the Then column ("equals", "above"), and the numbers are for the stop conditions.

**US-1.** As a researcher promoting a Source that records a kinship fact in another shape than my tree, I get the same help toward the match as when both record it the same way, so a marriage, a birth's parents or a parent chain finds the people my tree has as spouses, parents or grandparents.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Tree: Ann and Bob with a filed `spouse` relationship. Source: Ann and Bob each `spouse` in one undated `marriage` | Bob's row is on Bob's handle with the score of the same Source recording a `spouse` relationship instead (3.375, medium), above a Source whose two people are `witness` in the marriage (2.5, weak) |
| AC-1.2 | Tree: Ann and Bob each `spouse` in one filed undated `marriage` (all three promoted). Source: a `spouse` relationship | Bob's row scores as AC-1.1's direct-against-direct row (3.375, medium) |
| AC-1.3 | Tree: Ann `parent` of Bob. Source: Ann `father`, Bob (sex only) `subject` of one `birth`; Ann fixed on her handle | Bob's row's first alternative is Bob's handle, reason `via`, with the score of the same Source recording `parent` instead (0.875) |
| AC-1.4 | Tree: Ann `father`, Bob `subject` of a filed `birth`. Source: Ann `parent` of Bob (sex only); Ann fixed | As AC-1.3 |
| AC-1.5 | Tree: Ann `grandparent` of Cal. Source: Ann `parent` of Bea, Bea `parent` of Cal (named); Ann fixed | Cal's row is on Cal's handle, reason `via`, `Via.NeighborSubjectID` = Ann, `Via.Key` = `relationship·grandchild·related_to`, score as a Source recording the `grandparent` relationship (3.375, medium) |
| AC-1.6 | Tree: Ann `parent` of Bea, Bea `parent` of Cal (all promoted). Source: Ann `grandparent` of Cal (named); Ann fixed | As AC-1.5 |
| AC-1.7 | A row reached through an implied link (AC-1.5) | The page words it "via Ann → PER-…" with no relationship phrase, because the Evidence graph has no bridge between Ann and Cal (`PromoteModel.reasonText` falls back to `L10n.Promote.viaNeighbor`) |

**US-2.** As a researcher, a fact counts once, so a Source or tree that records it twice, or a fact both record in the same shape, does not inflate a match.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Tree: `spouse` relationship. Source: a `spouse` relationship and a `marriage` with both as `spouse` | `LayerBridges` holds one `relationship·spouse` bridge between Ann and Bob whose `Associations` are the relationship bridge and both participation bridges, `Implied` false; Bob's row scores as AC-1.1 (3.375) |
| AC-2.2 | Tree and Source each record one undated `marriage` with both as `spouse` (named Ann, Bob) | Bob's row scores as it does today without implied links (2.875, weak): the marriage events are paired, so the implied spouse link adds nothing |
| AC-2.3 | As AC-2.2 but Bob is sex only and Ann is fixed, so the marriage event row has too little to be placed | Bob's row's first alternative is Bob's handle with score 0.875, reason `via`: the events are not paired, so the implied link credits |
| AC-2.4 | Tree and Source each record Ann `parent` Bea `parent` Cal (named, Ann fixed) | Cal's row scores as it does today (3.375, medium): Bea is paired, so the implied grandparent link adds nothing |
| AC-2.5 | Pure: layer and canon each hold a marriage event and a spouse link that is both recorded and implied (`Implied` false on both) | The spouse links credit as two recorded spouse relationships do today (3.75 undated) |

**US-3.** As a researcher opening the evidence sheet for a row whose Source recorded a marriage, I see the spouse match and can pin it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Tree: filed `spouse` relationship. Source: a `marriage` with both as `spouse`. Evidence for Bob on his handle with Ann placed on hers | One relationship, key `relationship·spouse`, neighbor Ann, `Checked` true, `Incoming` holds every Observation of both participation bridges (6), `Members` holds one bundle (the filed spouse bridge) |
| AC-3.2 | `PinOffers` of AC-3.1's evidence | The bundle's offer is `Checked` and its `Set.Observations` start with AC-3.1's `Incoming` |
| AC-3.3 | Tree: filed `marriage`. Source: `spouse` relationship | No relationship line (the canon side derives in S11-40) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `kinship.Shape` and `Implied.Shape()` | STEP-1, AC-2.2 |
| `graphalign.Bridge` and `CanonEdge` carry `Shapes` and `Implied`; `corresponding` skips a link pair counted through a paired shape | STEP-2, AC-2.2 – AC-2.5 |
| The layer loader adds implied links through one merge with direct ones | STEP-3, AC-1.1, AC-1.3, AC-2.1 |
| `loadCanon` adds implied links through its dedupe, after the walk | STEP-4, AC-1.2, AC-1.4, AC-1.6 |
| Goldens: marriage ↔ spouse, birth parents ↔ parent, chain ↔ grandparent, each both ways; counted once; via | STEP-5, US-1, US-2 |
| The evidence query reads Source-side implied links as it reads any layer link | STEP-6, US-3 |
| Derivation cost measured | STEP-7 |
| `add-match-rule` says how callers merge and what the count-once rule reads | STEP-8 |
| `matching-model.md`, `promote-graph-alignment.md` say links are read through the kinship patterns | STEP-9 |

| Out | Home |
| --- | --- |
| The evidence query's canon side (`canonLinks`) deriving; a line saying how each side recorded the fact; bundles of an implied link | S11-40 |
| The count-once rule in the evidence sheet's relationship lines | S11-40 (map scope entry names it) |
| The sheet's wording of implied lines | S11-42 |
| The count-once rule in the duplicate check `newPairSegments` (`core/graphalign/rows.go`), which compares two Source Subjects' links | not planned — two Source people sharing one marriage's shapes is entry-error data; revisit if dogfood shows it |
| Memoizing the canon's implied links | not planned — measured at 31 ms for 8,190 parent links (STEP-7); revisit if a budget fails |
| A Swift change to the via wording | not planned — AC-1.7 already holds (`reasonsAreWordedWithLabelsAndRefs`) |
| New knobs, other patterns, trees and person pages | working rule "No new knob"; [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Pure derivation | `core/kinship/kinship.go` `Edge`, `Implied`, `Derive`; `patterns.go` `eventPatterns` | Returns implied links with `Pattern`, `Event` (marriage or birth) and `Through` (the middle parent). No caller. `Edge` has `Bridge`'s fields in its order (S11-38 D-1), so `kinship.Edge(b)` converts today |
| Layer edges | `core/graphalign/types.go` `Bridge{A, B, KeyA, KeyB, Associations}` | No shape or implied mark |
| Canon edges | `core/graphalign/types.go` `CanonEdge{From, To, KeyFrom, KeyTo}` | No associations, shape or implied mark |
| Adjacency | `core/graphalign/align.go` `adjLink{neighbor, key}`; `newState` fills `layerAdj` / `canonAdj` | Built from every bridge and edge |
| Correspondence | `core/graphalign/align.go` `(st *state) corresponding` | Canon links of a handle whose key `segmentcompare.Corresponds` to a layer link's. The one choke point: `place` (nomination), `segments` (credit) and `supports` (via) all call it |
| Layer loader | `core/database/promotealign/layer.go` `loadLayerBridges` | Walks every `connectrules` bridge, keys each edge with `segmentcompare.Keys`, merges same-ends, same-key edges inline through `seen` / `edgeSeenKey`, gathering `Associations`. Memoized with the layer (`loadLayerMemo`) |
| Canon loader | `core/database/promotealign/canon.go` `loadCanon` | Walks the whole connected component from the seeds; dedupes edges through `edgeSeen map[seenEdge]bool`. Runs on every proposal |
| Merge key | `core/database/promotealign/terms.go` `seenEdge`, `edgeSeenKey` | Ends in a fixed order with their keys |
| Evidence query | `core/database/promotealign/evidence.go` `evidence` | Builds `layerLinks` from `in.layer.Bridges` (every bridge with the row Subject at one end, Observations of its `Associations`); canon side from `canonLinks` (graph nodes, not `Canon.Edges`) |
| Readers of `Layer.Bridges` | `core/graphalign/align.go` `newState`; `core/database/promotealign/evidence.go` `evidence` | Nothing else: Apply and the subrows read the Evidence graph snapshot (`PromoteModel.bridgeFacts`) |
| Via wording | `macos/App/Features/Promote/PromoteModel.swift` `reasonText`; `PromoteFlow.bridge(between:and:)` | Phrase from the Evidence graph bridge between the two rows, else `L10n.Promote.viaNeighbor`. `viaBridgeType` / `viaRole` are mapped (`GoStore.swift`) and not read |
| Tests | `core/kinship/kinship_test.go` `TestDerive`, `TestDeriveInvariants`, `BenchmarkDerive` (generations 8–10); `core/graphalign/invariants_test.go` `TestAlignInvariants`; `core/graphalign/align_test.go` helpers `bridge`, `edge`, `partLink`, `relLink`, `nameVals`, `alignWith`; `core/database/promotealign/promotealign_test.go` fixture (`person`, `sexOnly`, `eventOn`, `participation`, `relationship`, `promote`, `propose`, `rowFor`), `TestProposeChildOfReadsAsParentOf`, `TestLayerBridgesCoverEveryConnectRule`; `evidence_test.go` `f.evidence`; `pins_test.go`; `perf_test.go` `TestLargeSourceBudgets`; `export_test.go` `LayerBridges`; `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` | No test has a marriage, a birth father or mother, or a chain on both sides of a proposal |

**Precedent.** S11-41 ([#561](https://github.com/mendahu/provenencia/pull/561)) changed both loaders' link reading in one PR; copy its shape for STEP-3 and STEP-4 (one helper, both loaders, no other caller touched). `TestProposeChildOfReadsAsParentOf` is the shape of each golden (a filed tree on `f.source`, a second Source, `propose`, `rowFor`). `TestAlignInvariants`'s fixture table takes STEP-2's new row.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/kinship` gains `type Shape struct { Pattern Pattern; Pivot []byte }` *(new)*, one way a link is recorded through a pattern: `Pivot` is the marriage or birth for an event pattern, the parent between for a chain. `func (i Implied) Shape() Shape` *(new)* returns `{i.Pattern, i.Event}` when `Event` is set, else `{i.Pattern, i.Through}`. `Edge`, `Implied` and `Derive` are otherwise unchanged | The count-once rule needs the pattern and its pivot on each link; the shape is kinship's concept, and S11-40 reads it too | Fields on `Edge` that `Derive` fills — every `TestDerive` row changes for data `Implied` already has; a `graphalign` shape type — the concept would live in the layer that only reads it |
| D-2 | `graphalign.Bridge` and `CanonEdge` gain, after their existing fields, `Shapes []kinship.Shape` (the patterns that imply the link; none for a link only recorded) and `Implied bool` (true when no bridge records the link: only patterns stand behind it). `adjLink` gains `shapes` and `implied`, filled in `newState`. `graphalign` imports `core/kinship` | A merged link can be both recorded and implied (AC-2.1), so "has shapes" is not "implied". `Implied`'s zero value is right for every recorded bridge, so no existing construction site changes | `Direct bool` — every bridge built today, in loaders and tests, would need it set |
| D-3 | `func (st *state) countedByShape(l, c adjLink) bool` *(new)* in `core/graphalign/align.go`: true when `l.implied || c.implied`, and some shape of `l` and some shape of `c` have the same `Pattern` and `st.placed[string(lShape.Pivot)]` equals `cShape.Pivot`. `corresponding` keeps a canon link only when `Corresponds` holds and `countedByShape` is false. Nothing else in `graphalign` changes | User, 2026-10-26: a fact both sides record in the same shape counts once, and only when the shape is actually paired, so a sparse marriage whose event row is not placed still gains (AC-2.3). `corresponding` is the one choke point, so nomination, credit and via agree. Placement changes already mark a pivot's layer neighbors dirty (`place`, `release`), so settling re-scores them | A static rule (skip whenever both sides share a pattern) — loses AC-2.3's credit; the stronger of the two credits — needs segment results across links (user declined) |
| D-4 | Both recorded (`Implied` false on both sides) keeps today's double credit (AC-2.5) | Recording the marriage and the relationship on both sides credits both today; this PR changes only what implied links add | Skip there too — lowers today's scores |
| D-5 | `loadLayerBridges` keeps its walk; its inline merge becomes `func mergeBridge(out []graphalign.Bridge, seen map[seenEdge]int, b graphalign.Bridge) []graphalign.Bridge` *(new, `layer.go`)*, used for direct edges and implied ones alike. On a hit it appends each association not already present (as today), appends `b.Shapes`, and sets `Implied` to `existing.Implied && b.Implied`. After the walk it builds `[]kinship.Edge` from `out` (fields `A, B, KeyA, KeyB, Associations`, copied by name), calls `kinship.Derive(edges, terms)`, and merges each `imp` as `graphalign.Bridge{A: imp.A, B: imp.B, KeyA: imp.KeyA, KeyB: imp.KeyB, Associations: imp.Associations, Shapes: []kinship.Shape{imp.Shape()}, Implied: true}` | Working rule "One fact, one link per side": the loaders' existing merge, not a second one. Direct edges all come first, so a recorded link is never merged into an implied one | A conversion `kinship.Edge(b)` — `Bridge` no longer has `Edge`'s fields (D-2); derive per bridge kind — `Derive` needs every link at once |
| D-6 | `loadCanon`: `edgeSeen` becomes `map[seenEdge]int` (the edge's index in `edges`). After the walk loop and before `prof.profile`, it builds `[]kinship.Edge` from `edges` (`A: From, B: To, KeyA: KeyFrom, KeyB: KeyTo`, no associations), calls `Derive`, and for each `imp`: on a hit appends `imp.Shape()` to that edge's `Shapes`; else appends `CanonEdge{From: imp.A, To: imp.B, KeyFrom: imp.KeyA, KeyTo: imp.KeyB, Shapes: []kinship.Shape{imp.Shape()}, Implied: true}` | `loadCanon` already walks the whole component, so a grandparent needs no extra reach (map review K-6). Implied edges after recorded ones keep the walk's link order | Derive per wave — a chain can span waves |
| D-7 | No memo for the canon's implied links | `BenchmarkDerive` at 8,190 parent links runs in 31 ms; the large-source budgets moved within noise in the prototype | A memo on the shortlist index generation — more state for a cost under the budgets |
| D-8 | The evidence query is unchanged: Source-side implied links reach it through `in.layer.Bridges`, and a Source marriage against a filed spouse relationship gives today's spouse line with both participations' Observations as `Incoming`, checked and pinned at Done (user, 2026-10-26) | Today's sentence holds until S11-42 anyway; a hold-back path is what the churn rule forbids | Filter implied bridges out of `layerLinks` until S11-40 |
| D-9 | No Swift, proto or FFI change. AC-1.7 is held by the existing `reasonsAreWordedWithLabelsAndRefs` (a via with no bridge between the rows); Go pins `Row.Via` (AC-1.5) | `viaRole` is not read; the phrase comes from the Evidence graph, which has no implied bridge | A Swift test duplicating the existing one |
| D-10 | Goldens live in `core/database/promotealign/promotealign_test.go` as one table-driven `TestProposeKinshipShapes` *(new)* plus the via assertions in its grandparent rows; the count-once rule's pure cases live in `core/graphalign/align_test.go` `TestKinshipShapesCountOnce` *(new)* and one `TestAlignInvariants` row | Derivation happens in the loaders, so end-to-end goldens need the catalog fixture; the rule itself is pure | Goldens in `graphalign` only — they would hand-build implied links and never test the loaders |
| D-11 | Golden assertions compare against a sibling row of the same table (the direct shape, the witness control, a prototype-free baseline) and assert band; exact scores appear only in a stop condition | A knob change elsewhere should not break kinship goldens that are about equality of shapes | Hard-coded scores in every row |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `kinship.Shape` *(new)*, `Implied.Shape()` *(new)* | type, method | `kinship.Implied` `Pattern`, `Event`, `Through` | Extends it: the method reads those fields; a link may carry several shapes, `Implied` is one derivation |
| `Bridge.Shapes`, `Bridge.Implied`, `CanonEdge.Shapes`, `CanonEdge.Implied` *(new)* | fields | `Bridge.Associations` (loader-set, layer-read-back data) | Extends the existing edge types; no second edge list |
| `adjLink.shapes`, `adjLink.implied` *(new)* | fields | `adjLink` | Extends it |
| `(st *state) countedByShape` *(new)* | func | `corresponding`, `segmentcompare.Corresponds` | `Corresponds` is keys only and pure layer 5; the rule needs placement, which only layer 6 has |
| `mergeBridge` *(new)* | func | the inline merge in `loadLayerBridges` | It is that merge, lifted so implied links share it |
| `TestProposeKinshipShapes`, `TestKinshipShapesCountOnce` *(new)* | tests | `TestProposeChildOfReadsAsParentOf`, `TestAlignInvariants` | New behavior |

Searched: `rg -n "Associations|edgeSeen|corresponding\(" core/graphalign core/database/promotealign`, `rg -n "kinship" core`.

**Builds on.** `kinship.Derive` in `core/kinship/kinship.go` (S11-38), the loaders' `edgeSeenKey` merge, `corresponding` as layer 6's one correspondence path, the `add-match-rule` skill.
**Deletes.** The inline merge body in `loadLayerBridges` (replaced by `mergeBridge`). The "no caller yet" sentences about `core/kinship` in `add-match-rule` (STEP-8).
**Next adds.** S11-40: `canonLinks` calls `Derive` and reads `Implied.Shape()` and `Through` for the line; its relationship lines apply the same count-once test. A new event pattern (a baptism's parents) is an `eventPatterns` row; its shape is counted once by the same rule with no change here.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `type Shape struct { Pattern Pattern; Pivot []byte }` *(new)*; `func (i Implied) Shape() Shape` *(new)* (D-1) |
| `core/graphalign` | `types.go`: `Bridge`, `CanonEdge` gain `Shapes []kinship.Shape`, `Implied bool` (D-2). `align.go`: `adjLink` gains `shapes`, `implied`; `newState` copies them; `countedByShape` *(new)*; `corresponding` calls it (D-3) |
| `core/database/promotealign` | `layer.go`: `mergeBridge` *(new)*, derivation after the walk (D-5). `canon.go`: `edgeSeen` by index, derivation after the walk (D-6) |

```go
// countedByShape reports whether a layer link and a canon link it corresponds
// to stand for one fact both sides record through the same pattern, already
// counted because the walk placed the layer pivot on the canon pivot. Two
// recorded links never are.
func (st *state) countedByShape(l, c adjLink) bool
```

Errors: none new.

### FFI

None — no wire change (D-9).

### macOS

None — no Swift change (D-9).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Shapes in `core/kinship`**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`. Skill: `add-match-rule`.
Do: add `Shape` and `Implied.Shape()` (D-1) beside `Implied`, with doc comments. In `kinship_test.go` add `TestImpliedShape` *(new)* (T-1). Extend `BenchmarkDerive`'s `generations` list to `8, 9, 10, 11, 12` (STEP-7 reads it).
Check: `go test ./core/kinship/... ./core/matchregistry/...` passes.

**STEP-2 — The count-once rule in `graphalign`**
Files: `core/graphalign/types.go`, `core/graphalign/align.go`, `core/graphalign/align_test.go`, `core/graphalign/invariants_test.go`. Skill: `add-match-rule`.
Do: D-2 and D-3. In `align_test.go` add `marriageShapes(layerShape, canonShape string, dated bool) func() (graphalign.Layer, graphalign.Canon)` *(new)* building Ann, Bob and (unless the shape is `rel`) a marriage on each side: shape `rel` is a recorded spouse bridge; `marriage` is two `partLink("spouse", "marriage")` bridges plus a spouse bridge with `Shapes: {{PatternMarriageSpouses, <marriage id>}}`, `Implied: true`; `both` is the same with `Implied: false`. Canon `Candidates` name each handle for its Subject. Add `TestKinshipShapesCountOnce` *(new)* (T-2) and the `"marriage on both sides"` row to `TestAlignInvariants` (`marriageShapes("marriage", "marriage", true)`, Ann fixed) (T-3).
Check: `go test ./core/graphalign/...` passes; every existing `graphalign` test passes unchanged.

**STEP-3 — The layer loader derives**
Files: `core/database/promotealign/layer.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-5. Add `TestLayerBridgesMergeKinshipShapes` *(new)* (T-4) through the existing `promotealign.LayerBridges` export.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run 'TestLayerBridges'` passes.

**STEP-4 — `loadCanon` derives**
Files: `core/database/promotealign/canon.go`.
Do: D-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/` passes (every existing test unchanged).

**STEP-5 — Goldens**
Files: `core/database/promotealign/promotealign_test.go`.
Do: add `TestProposeKinshipShapes` *(new)* (T-5): one table, each row a tree shape filed on `f.source` (every Subject promoted, events included), a Source shape on a new Source, and an expectation from §8. Shape builders are local closures taking `(src, art, a, b)` and returning the extra Subjects to promote; events come from `f.eventOn(src, art, "marriage" | "birth", nil)`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestProposeKinshipShapes -v` passes.

**STEP-6 — Evidence and pins**
Files: `core/database/promotealign/evidence_test.go`, `core/database/promotealign/pins_test.go`.
Do: add `TestEvidenceReadsASourceMarriageAsSpouses` *(new)* (T-6) and `TestPinOffersOfAnImpliedLayerLink` *(new)* (T-7). No production change (D-8).
Check: both pass.

**STEP-7 — Measure**
Files: none beyond STEP-1's benchmark.
Do: run `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -count=1 -v` on `main` and on the branch, and `go test -run x -bench Derive -benchmem ./core/kinship/`. Put the six budget lines from each run and the five benchmark lines in the PR body.
Check: every budget passes; `generations=12` is under 100 ms/op.

**STEP-8 — Skill**
Files: `.agents/skills/add-match-rule/SKILL.md`.
Do: rewrite the "Kinship patterns (`core/kinship`) add implied links before alignment" bullet: `Edge` carries `Bridge`'s first five fields; both loaders derive after their walk and merge implied links through their same-ends, same-key merge (`mergeBridge`, `loadCanon`'s `edgeSeen`), recording `Implied.Shape()` in `Shapes` and `Implied`; layer 6's `corresponding` skips a link pair `countedByShape` (a shared pattern whose pivots the walk paired, unless both are recorded); a new pattern is counted once by that rule with no layer 6 change.
Check: `rg -n "Nothing calls|no caller" .agents/skills/add-match-rule/SKILL.md` prints nothing about `core/kinship`.

**STEP-9 — Docs**
Files: `docs/matching-model.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: `matching-model.md` "The stack": after the profile-builder paragraph, a paragraph saying each side's links are read through the kinship patterns (`core/kinship`) before layer 6, an implied link weighs as the relationship it stands for, and a fact both sides record through the same, paired pattern counts once. Glossary rows after **segment**: **implied link**, **kinship pattern**, **shape**. `promote-graph-alignment.md` §12, after "Bridge kinds in the loaders": a bullet that both loaders add implied links from `core/kinship` through their merge, that implied links are never filed or shown as connections, and that a row reached through one says "via <neighbor>" with no phrase. Then the Definition of done's docs lines.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Match a Source marriage to the spouses a tree records."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/kinship_test.go` `TestImpliedShape` *(new)* | table: an `Implied` with `Event` E1 → `{PatternMarriageSpouses, E1}`; one with `Through` P2 and no `Event` → `{PatternParentOfParent, P2}` | D-1 |
| T-2 | Go | `core/graphalign/align_test.go` `TestKinshipShapesCountOnce` *(new)* | table over `marriageShapes`, Ann fixed, Bob's row: (rel, rel) is the baseline; (marriage, rel) and (rel, marriage) equal it; (marriage, marriage) undated is below it and equal to (marriage, marriage) with the spouse bridges removed; (marriage, marriage) with the layer marriage held `TargetNew` equals the baseline; (both, both) is above the baseline | AC-2.2, AC-2.3, AC-2.5, D-3, D-4 |
| T-3 | Go | `core/graphalign/invariants_test.go` `TestAlignInvariants` row `"marriage on both sides"` | the existing invariants (reversed input, second run, fixing a proposed row) hold | D-3 |
| T-4 | Go | `promotealign_test.go` `TestLayerBridgesMergeKinshipShapes` *(new)* | Source with a spouse relationship and a marriage of Ann and Bob → exactly one `relationship·spouse` bridge, 3 `Associations`, one `Shapes` entry with the marriage as `Pivot`, `Implied` false; marriage only → that bridge has 2 `Associations`, `Implied` true | AC-2.1, D-5 |
| T-5 | Go | `promotealign_test.go` `TestProposeKinshipShapes` *(new)* | rows for AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, each asserting target, handle, band, reason and score equality with its named sibling row; AC-1.5 and AC-1.6 also assert `Via` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| T-6 | Go | `evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses` *(new)* | AC-3.1 and AC-3.3 | AC-3.1, AC-3.3, D-8 |
| T-7 | Go | `pins_test.go` `TestPinOffersOfAnImpliedLayerLink` *(new)* | AC-3.2 | AC-3.2 |
| T-8 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` (existing) | a via row with no bridge between it and its neighbor → `L10n.Promote.viaNeighbor` | AC-1.7, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestAlignInvariants`, `TestGracieObituaryGolden`, `TestSparseNameAndBirthDate`, `TestLargeSourceBudgets` (budgets and "memo matches a fresh run"), `TestDerive`, the purity and numbers guard in `core/matchregistry/guard_test.go`.

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

`completed.md`: a new section at the top, `## S11-39 — Matching reads kinship shapes`, `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).`, then two or three sentences: Promote reads a marriage, a birth's parents and a parent chain as spouse, parent and grandparent links on both sides; a fact both sides record in the same paired shape counts once; a Source marriage against a filed spouse relationship shows the spouse line on the evidence sheet.

## 10. Manual check

1. In a project, file Source 1 with Ann Ames and Bob Ames and a `spouse` relationship between them; Promote both.
2. Create Source 2 with Ann Ames, Bob Ames and a `marriage` event in which both are `spouse`. Open Promote on it. Bob's row proposes Bob's handle at medium, as a Source 2 with a spouse relationship instead does.
3. Open Bob's evidence sheet with Ann on her handle: one spouse relationship line, checked.
4. Create Source 3 with Ann Ames, Bea Ames and Cal Ames, Ann `parent` of Bea and Bea `parent` of Cal, against a tree with Ann `grandparent` of Cal. Fix Ann. Cal's row says "via Ann Ames → PER-…".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rule makes a score depend on placement order | `corresponding` is already placement-dependent (`segments`); placing or releasing a pivot marks its layer neighbors dirty, and T-3 checks the invariants |
| A whole-family canon makes every proposal slower | D-7's measurement; budgets in `TestLargeSourceBudgets` |
| A pivot paired wrongly removes a true credit | The same wrong pairing already misleads the event path; the row stays contested if the events conflict |

**Stop and report if:**

- `corresponding` is no longer the only place `place`, `segments` and `supports` read correspondence.
- A row of §2 gives a band other than the one its AC names, or AC-1.1's direct-against-direct row is not 3.375 medium at the planning base (a knob moved: report, do not retune).
- `TestAlignInvariants` fails on the new row.
- An existing `graphalign` or `promotealign` test changes outcome.
- A budget in `TestLargeSourceBudgets` fails, or `BenchmarkDerive/generations=12` is 100 ms/op or more.
- `graphalign` importing `core/kinship` fails the guard in `core/matchregistry/guard_test.go` or makes an import cycle.

## 12. Open questions

None.

## 13. PR

**Title:** S11-39 — Matching reads kinship shapes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; STEP-7's numbers; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.md`.
