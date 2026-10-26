# S11-39 — Matching reads kinship shapes

**Status:** Landed in [#572](https://github.com/mendahu/provenencia/pull/572) (2026-10-26)
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Promote's rows and the evidence sheet's relationship Card are unchanged; only the scores and lines behind them move)
**Board:** —
**Depends on:** S11-49 (landed [#569](https://github.com/mendahu/provenencia/pull/569)), S11-50 (landed [#570](https://github.com/mendahu/provenencia/pull/570))
**Blocks:** S11-40
**Branch:** `s11-39-matching-reads-kinship-v2`, cut from `main` (the closed #567's branch keeps the old name)
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-queries.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape), [§ Where implied links live](../kinship-shapes.md#where-implied-links-live), [§ Counting once](../kinship-shapes.md#counting-once), [§ Weight](../kinship-shapes.md#weight); [layer 6 procedure](../comparison-layers.md#layer-6-procedure); S11-14 D-11 (link weights); working rules: "Implied links, not a new comparison", "The subject is an event's principal", "The canon's implied links live in the graph cache", "One fact, one link per side", "Counted once across sides", "No new weight", "Matching warms behind the open spinner"; closed [#567](https://github.com/mendahu/provenencia/pull/567) and its ship log on `s11-39-matching-reads-kinship` (the first build, measurements, the user's decisions)
**Planned:** 2026-10-26 against `main` @ `ca51a03c` (re-plan; the 2026-10-26 plan against `7efa02f` shipped as #567 and was closed in the slice 9 re-cut)
**Landed:** `TestProposeKinshipShapes` was ported whole with the canon-loader commit (AC-1.2, 1.4, 1.6 and 2.3 failed before it and pass after), so the goldens commit holds only the evidence and pin tests. `core/kinship` also has `TestRecordingMerge` for `Recording.Merge`. `BenchmarkDerive` runs generations 11 and 12, as in `fcc3b8b2`. Local budget figures: warm proposal 144 ms of 250 ms, cold 566 ms, after a commit 547 ms; `linkSet.add` is 2.4% of the budget test's CPU samples.
**Review:** [`S11-39-matching-reads-kinship.review.md`](S11-39-matching-reads-kinship.review.md) — re-plan: Fits with changes / Ready with changes, 2 rounds (every finding answered)

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

Promote reads each side's kinship links through the patterns: a marriage's two subjects, a birth's father or mother and subject, and a parent chain match the spouses, parents and grandparents the other side records directly, and the reverse. A fact both sides record through the same paired event counts once. The Source side derives when its layer loads; the canon side reads the graph cache's results through `canonlinks.Reader.For` (S11-49). This is the user's dogfood case: a tree marriage with two subjects against a Source spouse relationship proposes both people.

**Why now.** S11-49 put the canon's implied links in the graph cache and S11-50 moved the catalog-wide loading behind the open spinner. Matching is the first reader. #567 built the first version; this plan keeps its tested parts (shapes, the count-once rule, the merge, the goldens, the 12-generation budget fixture) and drops what the re-cut replaced (deriving over the walked component, the walk memo, the second link loader).

**What stays.** Weights (an implied link weighs as the direct link it stands for). The evidence query's canon side and the sheet's wording (S11-40, S11-42). No Swift, proto or FFI change.

## 2. User stories

Fixture names: Source people Ann Ames, Bob Ames, Bea Ames, Cal Ames (`f.person`); "sex only" is `f.sexOnly(src, art, "male")`; the tree has handles for the same names, filed on `f.source` with every Subject promoted, events included. Events are undated (`f.eventOn(src, art, type, nil)`). "Bob's row" is `rowFor(f.propose(graph, fixed), bob.ID)`; `fixed` is empty unless the AC says Ann is fixed on her handle. Each Then names the sibling row the test compares with (D-11); the number in brackets is #567's measurement and appears in tests only through §11's stop conditions.

**US-1.** As a researcher promoting a Source that records a kinship fact in another shape than my tree, I get the same help toward the match as when both record it the same way.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Tree: Ann and Bob with a filed `spouse` relationship. Source: Ann and Bob each `subject` of one `marriage` | Bob's row is on Bob's handle, band medium, score equal to the same tree against a Source recording a `spouse` relationship (3.375), and above a Source whose two people are `witness` in the marriage (2.5, weak) |
| AC-1.2 | Tree: Ann and Bob each `subject` of one filed `marriage`. Source: a `spouse` relationship | Bob's row is on Bob's handle, band medium, score equal to AC-1.1's direct-against-direct row (3.375) |
| AC-1.3 | Tree: Ann `parent` of Bob. Source: Ann `father`, Bob (sex only) `subject` of one `birth`; Ann fixed | Bob's row's first alternative is Bob's handle, reason `via`, score equal to the same tree against a Source recording Ann `parent` of Bob (sex only) (0.875) |
| AC-1.4 | Tree: Ann `father`, Bob `subject` of a filed `birth`. Source: Ann `parent` of Bob (sex only); Ann fixed | As AC-1.3, compared with a tree recording `parent` |
| AC-1.5 | Tree: Ann `grandparent` of Cal. Source: Ann `parent` of Bea, Bea `parent` of Cal; Ann fixed | Cal's row is on Cal's handle, reason `via`, `Via.NeighborSubjectID` = Ann's Source Subject, `Via.Key` = `vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "grandchild", End: vocabkey.PropertyRelatedTo}`, band medium, score equal to the same tree against a Source recording Ann `grandparent` of Cal (3.375) |
| AC-1.6 | Tree: Ann `parent` of Bea, Bea `parent` of Cal. Source: Ann `grandparent` of Cal; Ann fixed | As AC-1.5, compared with a tree recording `grandparent` |
| AC-1.7 | A row reached through an implied link (AC-1.5) | The page words it "via Ann Ames → PER-…" with no relationship phrase (`PromoteModel.reasonText` falls back to `L10n.Promote.viaNeighbor`) |

**US-2.** As a researcher, a fact counts once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Tree: `spouse` relationship. Source: a `spouse` relationship and a `marriage` with both as `subject` | `LayerBridges` holds one `relationship·spouse` bridge between Ann and Bob: 3 `Associations`, one `Shapes` entry, `OnlyImplied` false. Bob's row's score equals AC-1.1's direct-against-direct row (3.375) |
| AC-2.2 | Tree and Source each record one `marriage` with Ann and Bob as `subject` | Bob's row is on Bob's handle, band weak, score below AC-1.1's direct-against-direct row and above AC-1.1's witness control (2.875): the marriage events are paired, so the implied spouse link adds nothing |
| AC-2.3 | As AC-2.2 but Bob is sex only and Ann is fixed, so the marriage event row has too little to be placed | Bob's row's first alternative is Bob's handle, reason `via`, score equal to the same tree against a Source recording a `spouse` relationship to a sex-only Bob (0.875): the events are not paired, so the implied link credits |
| AC-2.4 | Tree and Source each record Ann `parent` Bea `parent` Cal; Ann fixed | Cal's row is on Cal's handle, band medium, score equal to AC-1.5's comparison row (3.375) |
| AC-2.5 | A Source with both a `spouse` relationship and a `marriage`, against a tree with only the `marriage` (D-4) | Bob's row's score equals AC-2.2's (2.875) |

**US-3.** As a researcher opening the evidence sheet for a row whose Source recorded a marriage, I see the spouse match and can pin it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Tree: filed `spouse` relationship. Source: a `marriage` with both as `subject`. `Evidence` for Bob on his handle with Ann placed on hers | One relationship, `Key.String()` = `relationship·spouse`, `Neighbor` = Ann's Source Subject, `Checked` true, `Incoming` holds every Observation of both participation bridges (6), `Members` holds one bundle (the filed spouse bridge) |
| AC-3.2 | `promotealign.PinOffers` of AC-3.1's evidence | The bundle's offer is `Checked`, and its `Set.Observations` starts with AC-3.1's `Incoming` |
| AC-3.3 | Tree: filed `marriage` with two `subject`s. Source: `spouse` relationship. `Evidence` for Bob as in AC-3.1 | No relationship (the canon side's evidence is S11-40) |

**US-4.** As a researcher with a large tree, Promote stays inside its budgets.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestLargeSourceBudgets` on the large-source fixture plus `823856b0`'s 12-generation parent tree with a marriage per couple (`bulkTree`: 4,095 persons, 4,094 parent links, 2,047 marriages), rooted at a name Source A files | Every budget passes locally (`-v`, figures in the PR body) and in CI's Go tests workflow (CI logs figures only on a failure) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `kinship.Shape`, `kinship.Recording`, `Implied.Shape()`, `kinship.CountedOnce` | STEP-1, T-1, T-2a |
| `graphalign.Bridge`/`CanonEdge` embed `kinship.Recording`; `corresponding` skips a counted-once pair | STEP-2, AC-2.2, AC-2.3, AC-2.5 |
| One `linkSet` merge in `promotealign`; the layer loader adds the Source side's implied links | STEP-3, AC-2.1 |
| `loadCanon` adds the canon's implied links from `canonlinks.Reader.For` for each walked person; the canon-side goldens | STEP-4, AC-1.2, AC-1.4, AC-1.6 |
| One type-id → kind lookup (`Vocabulary.TypeKeysByID`, `profiler.kind`) for every caller | STEP-5 |
| The other goldens, evidence and pin tests | STEP-6, AC-1.1, AC-1.3, AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-3.1, AC-3.2, AC-3.3 |
| The 12-generation tree in the budget fixture | STEP-7, AC-4.1 |
| Docs | STEP-8 |

| Out | Home |
| --- | --- |
| The evidence query's canon side; the shape on the relationship line | S11-40 |
| The sheet's wording | S11-42 |
| A memo of the walk; a second link loader; deriving over the walked component | not planned — replaced by S11-49's cache (slice 9 re-cut) |
| Keeping the shortlist index current on a handle commit | [`docs/ideas/name-index-on-commit.md`](../../../ideas/name-index-on-commit.md) |
| New weights, other patterns | working rule "No new weight"; [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Moving `displays` onto the memo slots | not planned in this PR (S11-49 follow-up) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Patterns | `core/kinship/kinship.go` `Implied`, `Derive(edges, terms, rules)`, `ParentStep`; `core/kinship/patterns.go` | `Implied` carries `Pattern`, `Event`, `Through`; no `Shape` or `CountedOnce` |
| Canon read | `core/database/canonlinks/reader.go` `NewReader(g, rules)`, `(*Reader).For(id)`; `canonlinks.go` `EdgeOf`, `LoadTerms` | `For` returns the implied links with `id` as an end, from cached per-pivot results; no caller outside tests |
| Registry | `core/matchregistry/registry.go` `Registry.Kinship` | The pivot bound, 10 |
| Alignment | `core/graphalign/types.go` `Bridge`, `CanonEdge`; `core/graphalign/align.go` `adjLink`, `newState`, `(*state).corresponding` (line 315) | `corresponding` keeps a canon link when `segmentcompare.Corresponds` holds; links carry no shape |
| Layer loader | `core/database/promotealign/layer.go` `loadLayerBridges` | Keys each walked bridge and merges same-ends, same-key edges inline through `seen map[seenEdge]int`, associations deduped with `slices.ContainsFunc`; no implied links |
| Canon loader | `core/database/promotealign/canon.go` `loadCanon` | Walks the component from the seeds in waves (`walked map[string]bool`, `queue`); keys each link with `canonlinks.EdgeOf`; dedupes through `edgeSeen map[seenEdge]bool`; no implied links |
| Type lookup | `core/database/promotealign/canon.go` `typeKey(types, id)` (callers `canon.go` twice, `promotealign.go` `checkFixed`, `profiles.go` `(*profiler).kind`); `typeByKey`; `core/database/graphcache/vocab.go` `TermsByID`, `PropertiesByID` | `typeKey` scans the key→id map on every call; `Vocabulary` has no types-by-id helper |
| Merge key | `core/database/promotealign/terms.go` `seenEdge`, `edgeSeenKey` | A struct of two id strings and two `BridgeKey`s (strings) |
| Evidence | `core/database/promotealign/evidence.go` | Reads `in.layer.Bridges` for the Source side and `canonLinks` for the canon side; unchanged by this PR |
| Budgets | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `newLargeFixture`; `core/database/promotealign/bench_test.go` `catalogShape`, `largeShape`, `bulkCanon` | S11-50's shape: "warm at open", cold after `WarmOnOpen`; no parent tree |
| Tests | `core/graphalign/align_test.go` (`TestGracieObituaryGolden`, `TestSparseNameAndBirthDate`), `core/graphalign/invariants_test.go` `TestAlignInvariants`, `core/database/promotealign/promotealign_test.go` `TestLayerBridgesCoverEveryConnectRule`, `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` | No kinship goldens |
| CI | `.github/workflows/go-test.yml` | Runs `go test -tags fts5 ./api/... ./core/...` without `-v` on non-draft PRs, so budget figures print only on a failure |

**Precedent.** Closed #567 (`origin/s11-39-matching-reads-kinship`) built this once against an older `main`. Port, by hand where `main` has moved, from these commits only:
- `fcc3b8b2` — `kinship.Shape`, `Implied.Shape()`, `CountedOnce` and their tests (`TestImpliedShape`, `TestCountedOnce`).
- `377ea0a3` — `graphalign` `Bridge`/`CanonEdge` fields, `adjLink`, `newState`, `corresponding`, `TestKinshipShapesCountOnce`, the `TestAlignInvariants` row; `c14e0d3c`'s `align_test.go` hunk (marriage role `subject`).
- `f143c82b` and `5a29a0a6` — `sideLink`, `linkSet` (association set per link), `export_test.go` `SideLink`/`LinkSetForTest`, `TestLinkSet`, `TestLayerBridgesMergeKinshipShapes`, the layer loader; **not** its `loadCanon` hunk (D-7 replaces it) and **not** its `withImplied` call shape (`Derive` now takes `rules`).
- `350f777f` — the loader doc comments, adapted to D-7.
- `40403706` — `TestProposeKinshipShapes` and its `kin…` helpers; `c14e0d3c`'s `promotealign_test.go` hunk (`kinMarriage("subject")`).
- `f3c2847e` — `TestEvidenceReadsASourceMarriageAsSpouses`, `TestPinOffersOfAnImpliedLayerLink`; `c14e0d3c`'s `evidence_test.go` hunk.
- `823856b0` — the 12-generation tree in `bench_test.go` (`TreeGenerations`, `TreeRoot`, `bulkTree`, `lookupTree`) and `largeShape`'s tree root; `c14e0d3c`'s `bench_test.go` hunk (marriage role `subject`). Fit it to S11-50's `perf_test.go`.
- `3ebb356e` — the `matching-model.md` paragraph, `promote-graph-alignment.md` §12 line and the `add-match-rule` bullet, adapted to the cache read; the glossary rows "kinship pattern" and "pivot" already exist.
In every ported line, a marriage participation role `"spouse"` becomes `"subject"`: `kinMarriage`, `marriageEvidence`, `marriageFixture`'s `partLink`, and `bench_test.go`'s `lookupTree` (its `termOf("role", "spouse")` and the tree ids' `spouse` field read the `subject` role term). Every `relationship_type` spouse stays.
Do **not** port `16ff9ef2`, `9c29cbaf`, `c14e0d3c`'s other hunks (on `main` through S11-49), `dbd674c9` or `3620aa7f` (replaced).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/kinship` gains `type Shape struct { Pattern Pattern; Pivot []byte }` *(new)*; `func (i Implied) Shape() Shape` *(new)* returns `{i.Pattern, i.Event}` when `Event` is set, else `{i.Pattern, i.Through}`; and `type Recording struct { Shapes []Shape; OnlyImplied bool }` *(new)*, how one link is recorded on one side (zero value: recorded only), with `func (r *Recording) Merge(o Recording)` *(new)*: appends each shape of `o` it lacks (same `Pattern` and `bytes.Equal` pivot), clones pivots, and sets `OnlyImplied` to both being only implied. `Implied`'s `Event` and `Through` stay | The count-once rule needs each link's pattern and pivot; one value carries them on every side and S11-40 adds one field, not two (round 1 M-3); shapes stay distinct when a canon link is read from both ends (M-2) | Fields on `Edge`; a `Shapes`/`OnlyImplied` pair on each carrier; folding `Event`/`Through` into an embedded `Shape` (M-7) — changes S11-49's landed API for no gain here |
| D-2 | `func CountedOnce(a, b Recording, paired func(pivot []byte) ([]byte, bool)) bool` *(new)*: false when neither side is only implied; else true when some shape of `a` and some of `b` share `Pattern` and `paired(aShape.Pivot)` returns `bShape.Pivot`. `a` is the layer side | One rule for alignment now and the evidence lines in S11-40. User, 2026-10-26: count once only when the pivot is actually paired, so a sparse marriage still gains (AC-2.3) | A static "same pattern" skip — loses AC-2.3 |
| D-3 | `graphalign.Bridge`, `CanonEdge` and `adjLink` embed `kinship.Recording` after their existing fields (filled in `newState`); `corresponding` keeps a canon link only when `Corresponds` holds and `kinship.CountedOnce(link.Recording, ce.Recording, st.placedOf)` is false (`(*state).placedOf`, `align.go:298`, already answers where a layer Subject is placed) | Zero values keep every recorded construction site unchanged; one placement lookup in the package (round 2 M-8) | A second `paired` closure on `state` |
| D-4 | Two links skip only when one of them is only implied (D-2). Recorded on both sides keeps today's credit | Only what implied links add changes | Skip only when both are only implied — AC-2.5 would double-credit |
| D-5 | One merge in `promotealign/links.go` *(new)*: `sideLink{kinship.Edge; kinship.Recording}` *(new)*, `linkSet{links []sideLink; at map[seenEdge]int; assocs []map[string]struct{}}` *(new)* with `newLinkSet()`, `add(l sideLink)`, `addImplied(imp kinship.Implied)` and `withImplied(terms segmentcompare.Terms, rules kinship.Rules)`; `seenEdge` and `edgeSeenKey` move there from `terms.go` (`observedTerms` stays). `add` keys by `edgeSeenKey`: on a hit it appends the associations the link lacks (the per-link set from `5a29a0a6`) and calls `Recording.Merge`; on a miss it clones ends, associations and every shape's pivot (as `Merge` does on a hit). `addImplied` adds `sideLink{Edge: imp.Edge, Recording: kinship.Recording{Shapes: []kinship.Shape{imp.Shape()}, OnlyImplied: true}}`; it is the one place implied links enter. `withImplied` runs `kinship.Derive(edges, terms, rules)` over the set's edges and `addImplied`s each result. `export_test.go`: `type SideLink = sideLink`, `func LinkSetForTest(links []SideLink, terms segmentcompare.Terms, rules kinship.Rules) []SideLink` | One fact, one link per side, one merge, in a file named for it (M-6) | A merge per loader |
| D-6 | `loadLayerBridges` adds each walked edge to a `linkSet` (its `seen` map and inline merge go), then `withImplied(terms, matchregistry.Default().Kinship)`, then converts to `[]graphalign.Bridge` with `Shapes` and `OnlyImplied` | The Source layer is small and memoized with the layer; it derives itself (working rule) | Reading the Source side through the graph cache — Source Subjects are not canonical nodes |
| D-7 | `loadCanon` builds one `canonlinks.Reader` per call (`canonlinks.NewReader(g, matchregistry.Default().Kinship)`), adds recorded links to a `linkSet` during the walk (its `edgeSeen` map goes), and appends each walked person's id to a `walkedPeople [][]byte` slice in wave order, where it sets `walked[id]`. After the walk, it calls `r.For(id)` for each id in that order and `addImplied`s each link whose two ends are both keys of `handles` (recorded links first, so order is deterministic), then converts to `[]graphalign.CanonEdge` with `Recording` (no associations). `loadCanon` never calls `kinship.Derive` | The cache holds one result per pivot (S11-49); a link read from both ends merges in the set. Ends outside the walk would name a handle the alignment does not hold | `withImplied` over the walked edges — the design the re-cut replaced; calling `For` on events — an event is never an implied link's end |
| D-8 | One type lookup: `graphcache.Vocabulary.TypeKeysByID() map[string]string` *(new)* (provenencia origin only, as `typeByKey` filters) beside `TermsByID`; `newProfiler` stores it on the profiler as `typeKeys`, and `(*profiler).kind(n)` reads `p.typeKeys[string(n.SubjectTypeID)]`; `loadCanon` uses `string(prof.kind(n))` for both of its lookups (its `kinds` map stays `map[string]string`); `checkFixed`, which runs before any profiler exists, reads `v.TypeKeysByID()` from `g.Vocabulary()` where it reads `typeByKey(g)` today, so no caller signature changes. `typeKey` is deleted; `typeByKey` stays for key → id (`index.go`). The merge key stays `seenEdge`; STEP-7 reports `linkSet.add`'s share of a warm proposal (`-cpuprofile`) in the PR body | #567's warm profile had `typeKey` at about 10 ms per proposal; one lookup for the package (M-1). Interning keys was not measured (M-5) | A `loadCanon`-local map — two lookups in one package; interning `BridgeKey`s |
| D-9 | The evidence query is unchanged: Source-side implied links reach it through `in.layer.Bridges` (AC-3.1); the canon side is S11-40 (AC-3.3) | User, 2026-10-26 | — |
| D-10 | No Swift, proto or FFI change; AC-1.7 is held by the existing `reasonsAreWordedWithLabelsAndRefs` | The via phrase falls back when no bridge joins the rows | A duplicate Swift test |
| D-11 | Goldens live in `promotealign_test.go` `TestProposeKinshipShapes` (one subtest per AC, each comparing with a sibling row by exact `Score`, band and reason); the exact count-once equalities live in pure `graphalign` tests (`TestKinshipShapesCountOnce`) | A knob change moves sibling rows together | Hard-coded scores |
| D-12 | The large-source fixture gains the 12-generation tree from `823856b0` (marriage role `subject`); every subtest of `TestLargeSourceBudgets` runs over it; budgets are unchanged. The PR body reports the local `-v` figures; CI shows pass or fail (it logs figures only on a failure) | AC-4.1; S11-50's warm at open is what makes them hold on CI | A separate tree budget case; changing the CI workflow to `-v` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `kinship.Shape`, `Recording`, `Recording.Merge`, `Implied.Shape`, `CountedOnce` *(new)* | types, funcs | `kinship.Implied` (`Pattern`, `Event`, `Through`) | Reads those fields; the rule belongs with the patterns |
| `Recording` embedded in `Bridge`, `CanonEdge`, `adjLink` *(new)* | field | `Bridge.Associations` | Same extension style |
| `promotealign/links.go`, `sideLink`, `linkSet`, `addImplied` *(new)* | file, types | `seenEdge`/`edgeSeenKey` and the two inline merges | Replaces both merges |
| `Vocabulary.TypeKeysByID` *(new)* | func | `TermsByID`, `PropertiesByID` | Same shape; replaces `typeKey`'s scan |
| `SideLink`, `LinkSetForTest` *(new)* | test exports | `export_test.go` `LayerBridges` | Same file |
| `kin…` test helpers *(new)* | helpers | `fixture` methods in `promotealign_test.go` | Built on them |

**Builds on.** `canonlinks.Reader.For` (S11-49); `matchregistry.Default().Kinship`; `kinship.Derive`; the fixture in `promotealign_test.go`; S11-50's budget test.
**Deletes.** `loadLayerBridges`' `seen` map and inline merge; `loadCanon`'s `edgeSeen` map; `typeKey`.
**Next adds.** S11-40 reads `Shapes` and `CountedOnce` for the relationship lines and `For` for the canon side. A new pattern changes `core/kinship` and `canonlinks` only.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | D-1, D-2 |
| `core/graphalign` | D-3; imports `core/kinship` |
| `core/database/graphcache` | `Vocabulary.TypeKeysByID` (D-8) |
| `core/database/promotealign` | D-5 – D-8 |

### FFI

None — no wire change.

### macOS

None — no chrome (D-10).

### Strings

None.

## 7. Steps

**STEP-1 — Shapes and the count-once rule.**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`. Skill: `add-match-rule`.
Failing test first: `TestImpliedShape`, `TestCountedOnce` (ported from `fcc3b8b2`).
Do: D-1, D-2.
Check: `go test ./core/kinship/ ./core/matchregistry/` passes (purity guard included).

**STEP-2 — Graph alignment counts a paired shape once.**
Files: `core/graphalign/types.go`, `core/graphalign/align.go`, `core/graphalign/align_test.go`, `core/graphalign/invariants_test.go`.
Failing test first: `TestKinshipShapesCountOnce` (ported from `377ea0a3`, roles `subject`).
Do: D-3, D-4.
Check: `go test ./core/graphalign/ ./core/kinship/ ./core/matchregistry/` passes (the purity guard sees the new import); every existing test unchanged.

**STEP-3 — One link merge; the layer loader adds implied links.**
Files: `core/database/promotealign/links.go` *(new)*, `core/database/promotealign/terms.go`, `core/database/promotealign/layer.go`, `core/database/promotealign/export_test.go`, `core/database/promotealign/promotealign_test.go`.
Failing test first: `TestLinkSet` (with a row: the same implied link added twice → one shape), `TestLayerBridgesMergeKinshipShapes`.
Do: D-5, D-6 (ports `f143c82b`, `5a29a0a6`, `350f777f`'s layer comment).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -skip TestLargeSourceBudgets` passes.

**STEP-4 — The canon loader reads the cache.**
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/promotealign_test.go`.
Failing test first: the `kin…` helpers and `TestProposeKinshipShapes` subtests AC-1.2, AC-1.4 and AC-1.6 (ported from `40403706`); they fail before D-7.
Do: D-7; `loadCanon`'s doc comment says recorded links come from the walk and implied ones from `canonlinks.Reader.For`, merged in one set.
Check: the three subtests pass; `go test ./core/graphalign/` (`TestGracieObituaryGolden`, `TestSparseNameAndBirthDate`) and `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -skip TestLargeSourceBudgets` (`TestLayerBridgesCoverEveryConnectRule`, every memo test) pass unchanged; `grep -n 'kinship.Derive' core/database/promotealign/canon.go` prints nothing.

**STEP-5 — One type lookup.**
Files: `core/database/graphcache/vocab.go`, `core/database/graphcache/vocab_test.go`, `core/database/promotealign/profiles.go`, `core/database/promotealign/canon.go`, `core/database/promotealign/promotealign.go`.
Do: D-8; add a `TestTypeKeysByID` row in `vocab_test.go` (a seeded type's id → its key; a user-origin type absent).
Check: `grep -rn 'typeKey(' core/database/promotealign` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/... -skip TestLargeSourceBudgets` passes unchanged.

**STEP-6 — Goldens, evidence and pins.**
Files: `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/evidence_test.go`, `core/database/promotealign/pins_test.go`.
Do: port the rest of `TestProposeKinshipShapes` (`40403706`), `TestEvidenceReadsASourceMarriageAsSpouses` and `TestPinOffersOfAnImpliedLayerLink` (`f3c2847e`) with `subject` marriage roles.
Check: each subtest passes.

**STEP-7 — The 12-generation tree in the budgets.**
Files: `core/database/promotealign/bench_test.go`, `core/database/promotealign/perf_test.go`.
Do: D-12, porting `823856b0` with the marriage role `subject`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` passes; the PR body lists every figure and `linkSet.add`'s share of a warm proposal (D-8).

**STEP-8 — Docs.**
Files: `docs/matching-model.md` (add after line 28 a "Before layer 6" paragraph: each side's links are read through the patterns, the Source side derived per layer and the canon side through `canonlinks.Reader.For`, merged in one `linkSet`, and the count-once rule; add glossary rows "implied link" and "shape" from `3ebb356e`), `docs/promote-graph-alignment.md` (§12: the same), `.agents/skills/add-match-rule/SKILL.md` (rewrite main's current kinship bullet, line 33: implied links enter through `linkSet.addImplied`, the one merge, and carry a `kinship.Recording`; the Layer 5 line 32: correspondence is `Corresponds`, except a pair `kinship.CountedOnce` skips at layer 6), `docs/deployment-plan/spike-11/kinship-shapes.md` (§ Counting once: name the one `linkSet` merge, not `edgeSeenKey` and `loadCanon`'s dedupe), `docs/deployment-plan/spike-11/deployment-plan.md` (the working rule "One fact, one link per side": the same; and the map row), `docs/deployment-plan/spike-11/completed.md` (newest first), this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports 0 errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/kinship_test.go` `TestImpliedShape` *(new)* | an `Implied` with `Event` E1 → `{PatternMarriageSpouses, E1}`; one with `Through` P2 → `{PatternParentOfParent, P2}` | D-1 |
| T-2a | Go | `core/kinship/kinship_test.go` `TestCountedOnce` *(new)* | `paired` maps E1 → E9: shared pattern, pivots E1/E9, a only implied → true; b only implied → true; neither → false; E1/E8 → false; nothing paired → false; different patterns → false; empty → false | D-2, D-4 |
| T-2 | Go | `core/graphalign/align_test.go` `TestKinshipShapesCountOnce` *(new)* | #567's fixture with `subject` roles: (rel, rel) = (marriage, rel) = (rel, marriage); (marriage, marriage) = (both, marriage) = (marriage, both) = (event, event), each below (rel, rel); (marriage, marriage) with the marriage fixed New = (rel, rel); (both, both) above (rel, rel); dated (marriage, marriage) = (event, event) | AC-2.2, AC-2.3, AC-2.5, D-3, D-4 |
| T-3 | Go | `core/graphalign/invariants_test.go` `TestAlignInvariants` row `"marriage on both sides"` | T-2's fixture, (marriage, marriage), dated, Ann fixed: invariants hold | D-3 |
| T-4 | Go | `core/database/promotealign/promotealign_test.go` `TestLinkSet` *(new)* | the same recorded link from each end → one link, both associations; a spouse relationship plus a marriage → one link, 3 associations, one shape, `OnlyImplied` false; two marriages of one couple → one implied link, two shapes, 4 associations; the same implied link added twice → one shape (rows with implied links are built from recorded participation and relationship edges, so `withImplied`'s `Derive` produces them); a returned link's `A` does not alias the input's | D-1, D-5 |
| T-5 | Go | `core/database/promotealign/promotealign_test.go` `TestLayerBridgesMergeKinshipShapes` *(new)* | AC-2.1's Source → as AC-2.1; marriage only → the spouse bridge has 2 associations, `OnlyImplied` true | AC-2.1, D-6 |
| T-6 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeKinshipShapes` *(new)* | one subtest each for AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 against its sibling row | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, D-7, D-11 |
| T-7 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsASourceMarriageAsSpouses` *(new)* | AC-3.1 and AC-3.3 | AC-3.1, AC-3.3, D-9 |
| T-8 | Go | `core/database/promotealign/pins_test.go` `TestPinOffersOfAnImpliedLayerLink` *(new)* | AC-3.2 | AC-3.2 |
| T-9 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` (existing) | a via row with no bridge → `L10n.Promote.viaNeighbor` | AC-1.7, D-10 |
| T-10 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` over the tree | every subtest within its budget, locally and in CI | AC-4.1, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestAlignInvariants`, `TestGracieObituaryGolden`, `TestSparseNameAndBirthDate`, `TestLargeSourceBudgets` ("memo matches a fresh run" included), `TestLayerBridgesCoverEveryConnectRule`, the purity guard.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS
- [ ] CI's Go tests workflow passes TestLargeSourceBudgets over the tree; the PR body lists the local -v figures
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Domain docs updated: matching-model.md, promote-graph-alignment.md, add-match-rule SKILL.md
- [ ] PR description: what shipped, AC evidence, the commits ported from #567, deviations, manual check
```

## 10. Manual check

1. File Source 1 with Ann Ames and Bob Ames both `subject` of a `marriage`; promote both and the event.
2. Create Source 2 with Ann Ames, Bob Ames and a `spouse` relationship. Open Promote: both rows propose their handles; Bob's at medium.
3. File Source 3 with Ann Ames and Bob Ames and a `spouse` relationship; promote. Create Source 4 with Ann and Bob both `subject` of a `marriage`; open Bob's evidence sheet with Ann on her handle: one spouse relationship line, checked.
4. File Ann Ames `grandparent` of Cal Ames on Source 5 and promote. Create Source 6 with Ann `parent` of Bea and Bea `parent` of Cal; set Ann's row to her handle. Cal's row says "via Ann Ames → PER-…".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `For` per walked person is slow on a cold first proposal | Results are cached per pivot (S11-49); AC-4.1 measures it; S11-50 warms the kinds at open |
| An implied link names a handle the walk did not reach | D-7 adds only links whose two ends were walked |
| A score moves in an existing golden | Stop condition |

**Stop and report if:**

- A golden other than the new kinship ones changes score.
- A sibling equality in T-6 differs from #567's measurement (§2's bracketed figures).
- `TestLargeSourceBudgets` exceeds a budget locally or in CI.
- `canonlinks.Reader.For` returns a link whose `Edge` keys differ from `EdgeOf`'s for the same pair read directly (the merge would keep two links).
- A ported hunk needs a file this plan does not name.

## 12. Open questions

None.

## 13. PR

**Title:** S11-39 — Matching reads kinship shapes
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; the commits ported from #567; deviations; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.md`.
