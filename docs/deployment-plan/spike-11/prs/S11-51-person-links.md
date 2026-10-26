# S11-51 — One person's links for every reader

**Status:** Landed in [#575](https://github.com/mendahu/provenencia/pull/575) (2026-10-26)
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine move and one read with no caller in the app yet)
**Board:** —
**Depends on:** S11-39 (landed [#572](https://github.com/mendahu/provenencia/pull/572))
**Blocks:** S11-40, S13-04 (Spike 13)
**Branch:** `s11-51-person-links`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-queries.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape), [§ Where implied links live](../kinship-shapes.md#where-implied-links-live); S11-39 plan D-5 – D-7 and its Landed note (F-9); working rules: "One fact, one link per side", "The canon's implied links live in the graph cache", "One pure package, `core/kinship`"; Spike 13 working rule ["One person's links, direct and implied, from one function"](../../spike-13/deployment-plan.md#engine)
**Planned:** 2026-10-26 against `main` @ `b3b6a78e`
**Review:** [`S11-51-person-links.review.md`](S11-51-person-links.review.md) — Fits with changes / Ready with changes, 2 rounds (every finding answered)

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

The merge that makes a fact recorded two ways one link (S11-39's `linkSet`) moves out of `promotealign` into the pure `core/kinship`, beside `Recording.Merge`, and takes in `loadCanon`'s separate merge of implied links, so there is one merge with one key. `canonlinks` gains one read that returns a canonical person's relationship links, recorded and implied, merged through it. Promote's proposals do not change.

**Why now.** S11-40 (the evidence query's canon side) and Spike 13's Family read (S13-04) each need one person's links, direct and implied. Without this PR each would write its own merge beside the two loaders' (Spike 13 map review K-2; the user moved it here on 2026-10-26, and chose `core/kinship` as its home in this plan's review).

**What stays.** Every proposal, band, evidence line and pin. `loadCanon`'s F-9 cost shape from S11-39 (implied links placed into the walked edges without growing the set, `[]graphalign.CanonEdge` built once). `Reader.For` and its graph cache slots. `kinship.Derive` and the patterns.

## 2. User stories

**US-1.** As the engine, when S11-40 or the Family read asks for one canonical person's relationships, `canonlinks` returns each once, however many ways the catalog records it, so that neither reader merges.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Ann and Bob filed as spouses (relationship link `AB`), and marriage M1 with Ann and Bob each `subject`; `Relationships(Ann)` | One link: ends {Ann, Bob} with spouse keys, Associations {`AB`, `M1/Ann`, `M1/Bob`}, Shapes [{`marriage_spouses`, M1}], `OnlyImplied` false. `Relationships(Bob)` returns the same one link (ends may come in the other order) |
| AC-1.2 | Ann filed as parent of Bea (`AB`), and birth B1 with Ann `father` and Bea `subject`; `Relationships(Bea)` | One link: A = Ann, B = Bea, keys `w.keys(relationship, parent)`, Associations {`AB`, `B1/Ann`, `B1/Bea`}, Shapes [{`birth_parents`, B1}], `OnlyImplied` false |
| AC-1.3 | Ann filed as parent of Bea (`AB`), Bea filed as parent of Cal (`BC`); `Relationships(Ann)` | Two links, in this order: Ann → Bea, keys `w.keys(relationship, parent)`, Associations {`AB`}, no shapes, `OnlyImplied` false; then Ann → Cal, keys `w.keys(relationship, grandparent)`, Shapes [{`parent_of_parent`, Bea}], `OnlyImplied` true |
| AC-1.4 | Only marriage M1 with Ann and Bob each `subject`; `Relationships(Ann)` | One spouse link, `OnlyImplied` true. No participation link is returned: every link's `KeyA.Type` is `relationship` |
| AC-1.5 | A person with no relationship link and no event; 16 zero bytes (an id not in the catalog) | Both return no links and no error |
| AC-1.6 | AC-1.1's catalog; deep copies of `Relationships(Ann)` and `For(Ann)` taken first; the caller overwrites the first byte of link 0's `A` and of every association of link 0 | A second `Relationships(Ann)` and `For(Ann)` equal their deep copies |

**US-2.** As a contributor, the move changes no behavior, so that matching's goldens and budgets prove it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The `promotealign` and `graphalign` suites, unchanged | Pass |
| AC-2.2 | `TestLinkSet`'s five cases, moved to `core/kinship` with the same inputs and assertions | Pass against `kinship.LinkSet` |
| AC-2.3 | `TestLargeSourceBudgets` | Passes with the budgets in `perf_test.go` unchanged |
| AC-2.4 | `PROVENENCIA_GRAPH_VERIFY=1` on the `canonlinks` package | Passes |
| AC-2.5 | The names grep in STEP-2 and STEP-4 | Prints nothing |
| AC-2.6 | `PlaceImplied` over a set holding spouse Ann–Bob, given [spouse Ann–Bob via M1, spouse Cy–Di via M2, spouse Di–Cy via M3] | Returns [0, 1, 1] for a set of one link; the set still holds one link |

## 3. Scope

| In | Covered by |
| --- | --- |
| `sideLink`, `linkSet` (`add`, `addImplied`, `withImplied`), `seenEdge`, `edgeSeenKey` move to `core/kinship` (D-1, D-2) | STEP-1, AC-2.2, AC-2.5 |
| `LinkSet.PlaceImplied` *(new)* takes in `loadCanon`'s implied-link merge; `Implied.Recording` *(new)* is the one implied recording (D-3, D-3a) | STEP-1, AC-2.1, AC-2.3, AC-2.6 |
| `loadLayerBridges` and `loadCanon` use `kinship.LinkSet` | STEP-1, AC-2.1 |
| `TestLinkSet` moves to `core/kinship`; `export_test.go`'s `SideLink` and `LinkSetForTest` go (D-7) | STEP-1, AC-2.2 |
| `deriver.relationshipEdges` *(new)*: one read of a node's relationship links as edges, for `chainLinks`, `forLinks` and `Relationships` (D-6) | STEP-3 |
| `Reader.Relationships(id)` *(new)*: a person's recorded relationship links and `For`'s implied links, merged through one `kinship.LinkSet` (D-4, D-5) | STEP-3, AC-1.1 – AC-1.6, AC-2.4 |
| Domain docs and `add-match-rule` say the merge is `kinship.LinkSet` (D-8) | STEP-4, AC-2.5 |

| Out | Home |
| --- | --- |
| The evidence query's canon side (`canonLinks` in `core/database/promotealign/evidence.go`) reading `Reader.Relationships` | S11-40 |
| Siblings, or any other pattern | S13-03 (page-only) |
| Any RPC, the Family read and its grouping | S13-04 |
| Filtering merged or unaccepted neighbors in `Relationships` | S13-04 (its "Canonical only" rule), S11-40 |
| Memoizing `Relationships` | not planned (`For` is memoized; the recorded part is one scan of the person's links) |
| `loadCanon` calling `Relationships` per walked person | not planned (D-3) |
| Rewording history: landed plans, ship logs, review logs, `completed.md`'s S11-39 paragraph, the map's S11-39 row and scope entry | not planned (they record what landed then) |
| S11-40's scope entry (it already names `canonlinks`; its plan words the `canonLinks` / `Relationships` split) | S11-40 |
| Walking a narrower part of the canon | [`docs/ideas/narrower-canon-walk.md`](../../../ideas/narrower-canon-walk.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| The merge | `core/database/promotealign/links.go` `seenEdge`, `edgeSeenKey`, `sideLink`, `linkSet`, `newLinkSet`, `add`, `addImplied`, `withImplied` | Private to `promotealign`: the same-ends, same-keys merge, recorded and implied (S11-39 D-5). 98 lines, the whole file. Imports only `bytes`, `core/kinship`, `core/segmentcompare`, `core/vocabkey` |
| Layer loader | `core/database/promotealign/layer.go` `loadLayerBridges` | `set := newLinkSet()`, `set.add(sideLink{…})` per walked edge, `set.withImplied(terms, matchregistry.Default().Kinship)`, then reads `set.links` into `[]graphalign.Bridge`. Doc comment names "one linkSet" |
| Canon loader | `core/database/promotealign/canon.go` `loadCanon` | `set := newLinkSet()`, `set.add(sideLink{…})` per walked link; reads `set.links` into `[]graphalign.CanonEdge`; then a second merge (lines 205–221): each implied link from `reader.For` is keyed with `edgeSeenKey`, found in `set.at` or its own `impliedAt map[seenEdge]int`, and `Recording.Merge`d into `edges[i]` or appended (S11-39 F-9: no second set, for the warm budget). Doc comment names "one linkSet" |
| Kinship | `core/kinship/kinship.go` `Edge`, `Implied`, `Shape`, `Recording`, `Recording.Clone`, `Recording.Merge`, `Derive`, `ParentStep`; package doc | Pure; in the guard list (`core/matchregistry/guard_test.go`, `{"core/kinship", true}`). The package doc ends "It returns only implied links; callers merge them with direct ones." |
| Canon reads | `core/database/canonlinks/canonlinks.go` `LoadTerms`, `EdgeOf`; package doc | The package doc lists `EdgeOf`, `LoadTerms`, `Reader.For` |
| Implied links | `core/database/canonlinks/reader.go` `Reader`, `NewReader`, `For`, `loadAround`, `deriver.edge`, `deriver.chainLinks`, `deriver.forLinks` | `For(id)` returns the implied links with `id` at an end, from the `canonlinks.kinship` and `canonlinks.for` memo slots; the slice is the cache's, shared and read-only. `chainLinks` and `forLinks` each loop `n.Links`, keep `BridgeTypeKey == vocabkey.BridgeRelationship`, and call `d.edge(n, l)` |
| Test shims | `core/database/promotealign/export_test.go` `SideLink`, `LinkSetForTest` | Used only by `TestLinkSet` |
| Tests | `core/database/promotealign/promotealign_test.go` `TestLinkSet` | Five cases over `promotealign.SideLink` and `LinkSetForTest`, built with `segmentcomparetest.Terms()` and `segmentcomparetest.Bridge` |
| Tests | `core/kinship/kinship_test.go` | `package kinship_test`; a package-level `terms = segmentcomparetest.Terms()` |
| Tests | `core/database/canonlinks/reader_test.go` `world`, `newWorld`, `person`, `event`, `join`, `parent`, `marriage`, `keys`, `TestFor`, `TestForFollowsCommits` (verify branch at line 321) | Catalog fixture over `graphcachetest.Canon`; ids named for messages. `parent` hard-codes `relationship_type` `parent` |
| Tests | `core/database/promotealign/promotealign_test.go` `TestProposeKinshipShapes`, `TestLayerBridgesMergeKinshipShapes`; `pins_test.go` `TestPinOffersOfAnImpliedLayerLink`; `perf_test.go` `TestLargeSourceBudgets` | Prove the loaders' behavior and the warm budget (350 ms) |
| Docs | [`docs/matching-model.md`](../../../matching-model.md) line 29, [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) line 243, [`kinship-shapes.md`](../kinship-shapes.md) line 53, [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) line 33 | Each names `promotealign`'s `linkSet` (and `addImplied`) as the one merge |
| Map | [`deployment-plan.md`](../deployment-plan.md) line 104 (working rule "One fact, one link per side"), line 168 (S11-51 row), the S11-51 scope entry; [Spike 13 map](../../spike-13/deployment-plan.md) line 64 | Updated with this plan: they name `kinship.LinkSet` and `Reader.Relationships`. STEP-4 changes only the S11-51 row's status |

**Precedent.** S11-49 ([#569](https://github.com/mendahu/provenencia/pull/569)) moved `EdgeOf` and `LoadTerms` out of `promotealign` into `canonlinks` and exported them, with `promotealign` calling the new names; copy that shape: move the code, export the names, change the callers, keep the bodies. S11-39 ([#572](https://github.com/mendahu/provenencia/pull/572)) wrote the set and F-9; keep their bodies.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `promotealign/links.go` is deleted and its code moves to `core/kinship/links.go` *(new)*: `sideLink` → `Link`, `linkSet` → `LinkSet`, `newLinkSet` → `NewLinkSet`, `add` → `Add`, `addImplied` → `AddImplied`, `withImplied` → `WithImplied`; `seenEdge` → `linkKey` and `edgeSeenKey` → `linkKeyOf` (unexported). Inside the package, `kinship.Edge` and `kinship.Recording` lose their prefix. Bodies and doc comments are unchanged except names and "both loaders" → "every reader of a side's links". The package doc's last sentence becomes "It returns only implied links; `LinkSet` merges them with direct ones." | Pure code over `kinship` types, beside `Recording.Merge`; the Source loader uses it and does not read the canon (user, 2026-10-26, this plan's review M-3) | `core/database/canonlinks` — the map's first home; a canon-reading database package for a merge that is neither; a new `core/database/links` — a third home |
| D-2 | `LinkSet.Links() []Link` *(new)* returns the set's own slice in first-added order (read-only for callers; nil for an empty set) | The loaders read `set.links` today | Exported fields — callers could break the `at` / `assocs` parallel invariant |
| D-3 | `func (s *LinkSet) PlaceImplied(implied []Implied) []int` *(new)* returns, for each implied link, the index in `Links()` of the link with the same `linkKeyOf`, or `len(Links()) + j` where `j` (0-based) counts the distinct new keys among `implied` in order, so `at[i] == len(edges)` exactly when `loadCanon` appends. It changes nothing in the set. `loadCanon` builds `edges` from `set.Links()` sized `len(Links()) + len(implied)` as today, then for each `i`: `rec := imp.Recording()`; if `at[i] < len(edges)` it `Merge`s `rec` into `edges[at[i]].Recording`, else appends a new `CanonEdge` with cloned ends and `rec.Clone()`. `impliedAt`, `edgeSeenKey` and the `set.at` read leave `canon.go` | One key and one placing rule, in the set's file (M-1); F-9's cost shape stays: no set growth, one `CanonEdge` build | `set.AddImplied` per implied link — the shape F-9 measured and removed (S11-39 PRF-1); exporting the key and `IndexOf` so `loadCanon` keeps its own loop — a second merge across a package boundary |
| D-3a | `func (i Implied) Recording() Recording` *(new)* in `kinship.go`, beside `Implied.Shape()`: `Recording{Shapes: []Shape{i.Shape()}, OnlyImplied: true}`. `AddImplied` adds `Link{Edge: imp.Edge, Recording: imp.Recording()}`; `loadCanon` uses it (D-3). The literal appears once in the repo | How an implied link records itself is one value in one place, for both sides (round 2 M-7) | Building the literal in `AddImplied` and in `loadCanon` |
| D-4 | `func (r *Reader) Relationships(id []byte) ([]kinship.Link, error)` *(new)* in `reader.go`: `n, err := r.loadAround(id)` (nil node → nil, nil); a `kinship.NewLinkSet()`; `Add(kinship.Link{Edge: e})` for each edge of `r.d.relationshipEdges(n)`, in `n.Links` order; then `For(id)` and `AddImplied` for each, in `For`'s order; return `set.Links()` | Recorded links are read as `forLinks` reads them, so `kinship.ParentStep` reads their direction the same way; `Add` clones every byte it keeps, so the caller owns the result while `For`'s slice stays the cache's. "Relationships" names what it returns, apart from `graphcache.Link` (M-4) | Including participation links — they are a person's events, not kin; S11-40 keeps reading them through its `canonLinks`. Returning `[]kinship.Implied` plus recorded edges — two lists for the caller to merge |
| D-5 | `Relationships` does not filter: a recorded link to a merged node, or to a person with no accepted member, is returned, as `For` returns implied links today. It is not memoized | One reach for `For` and `Relationships`; each reader applies its own filter (`loadCanon` skips merged neighbors; S13-04's "Canonical only" rule) | Filtering here — S11-40 and S13-04 filter differently and neither is planned yet |
| D-6 | `func (d *deriver) relationshipEdges(n *graphcache.Node) []kinship.Edge` *(new)*, beside `deriver.edge`: each of `n.Links` with `BridgeTypeKey == vocabkey.BridgeRelationship` that `d.edge` reads, in order. `chainLinks` appends its result; `forLinks` ranges over it before `kinship.ParentStep`; `Relationships` adds each | One place decides which of a person's links are kin (M-2) | A third copy of the loop |
| D-7 | `TestLinkSet` moves to `core/kinship/links_test.go` *(new)* (`package kinship_test`): same subtest names, inputs and assertions; the local `terms` and `rules` go and the test uses `kinship_test.go`'s package-level `terms` and `rules` (`MaxPivotEnd: 10`, the registry's value); its local closures `person`, `event`, `bridgeID`, `part`, `spouse`, `flip` stay and shadow the package's `p`, `e`, `b`, `part`; building a `kinship.NewLinkSet()`, `Add`ing each input, `WithImplied(terms, rules)` and reading `Links()`; one new subtest for AC-2.6. `export_test.go` loses `SideLink` and `LinkSetForTest`. Each file the move touches drops the imports it leaves unused, and only those | The test follows the code; AC-2.2 is the map's "`TestLinkSet` unchanged" | A test shim — `kinship_test` calls the exported API |
| D-8 | Docs: `matching-model.md` (line 29), `promote-graph-alignment.md` (line 243), `kinship-shapes.md` (line 53) and `add-match-rule` (line 33) name `kinship.LinkSet` as the one merge (implied links enter through `AddImplied`, or are placed by `PlaceImplied` in `loadCanon`, each recorded by `Implied.Recording`); only the clause naming the old home changes, every other sentence stays. `add-match-rule` also names `canonlinks.Reader.Relationships`. The map's working rule, the S11-51 row and scope entry, and Spike 13's rule were updated with this plan | The plan and the map agree (user's answer, M-3, M-5) | Leaving the docs naming `promotealign` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/kinship/links.go` *(new)* | file | `core/database/promotealign/links.go` | It is that file, moved (D-1) |
| `kinship.Link`, `LinkSet`, `NewLinkSet`, `Add`, `AddImplied`, `WithImplied`, `linkKey`, `linkKeyOf` *(new names)* | types, funcs | `promotealign.sideLink`, `linkSet`, `newLinkSet`, `add`, `addImplied`, `withImplied`, `seenEdge`, `edgeSeenKey` | Moved and renamed, bodies unchanged |
| `LinkSet.Links` *(new)* | func | `linkSet.links` field reads in `layer.go` and `canon.go` | The same read across a package boundary (D-2) |
| `LinkSet.PlaceImplied` *(new)* | func | `loadCanon`'s `impliedAt` loop, `canon.go` lines 205–221 | That loop, moved beside `Add` so the key has one user outside it (D-3) |
| `Implied.Recording` *(new)* | func | `Implied.Shape`; the literal in `addImplied` and `canon.go:212` | Those two literals, named once beside `Shape` (D-3a) |
| `deriver.relationshipEdges` *(new)* | func | the loops in `deriver.chainLinks` and `deriver.forLinks` | Those loops, named once (D-6) |
| `Reader.Relationships` *(new)* | func | `Reader.For`; `deriver.forLinks` | `For` returns implied links only, cached and shared; `Relationships` adds the recorded ones and merges (D-4). Searched `canonlinks`, `promotealign`, `conclusiondetails`, `conclusionheaders` for a per-person relationship read: none |
| `core/kinship/links_test.go` *(new)* | test file | `promotealign_test.go` `TestLinkSet` | Moved (D-7) |
| `TestRelationships` *(new)* | test | `reader_test.go` `TestFor` and its `world` fixture | Reuses `world`, `person`, `marriage`, `join`, `event`, `keys` |
| `world.relationship(name, a, b, term)` *(new)* | test helper | `world.parent` | `parent` becomes a one-line call to it, and `spouse` cases call it with `spouse` (M-6) |

**Builds on.** S11-49's `canonlinks` (`EdgeOf`, `Reader`, `deriver.edge`), S11-39's `linkSet`, F-9 and `kinship.Recording`, the `world` test fixture, the guard list's `core/kinship` entry (the moved file has no numeric literal).
**Deletes.** `core/database/promotealign/links.go`; `loadCanon`'s `impliedAt` loop; `export_test.go` `SideLink` and `LinkSetForTest`; `TestLinkSet` in `promotealign_test.go` (moved); the relationship loops in `chainLinks` and `forLinks` (D-6).
**Next adds.** S11-40 replaces the relationship-bridge links `canonLinks` reads between two handles with `Reader.Relationships(from)` kept to links whose other end is `to`; participation links stay in `canonLinks`. S13-04 reads `Relationships` for parents, spouses and children. S13-03's page-only sibling derivation is asked for beside it, not added to `For` or `Relationships`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `links.go` *(new)*: D-1, D-2, D-3. Package doc: D-1 |
| `core/database/canonlinks` | `reader.go`: `deriver.relationshipEdges` (D-6), `Reader.Relationships` (D-4, D-5). Package doc adds "and one person's relationship links, recorded and implied, merged through `kinship.LinkSet` (`Reader.Relationships`)" |
| `core/database/promotealign` | `links.go` deleted. `layer.go` `loadLayerBridges`: `kinship.NewLinkSet()`, `Add(kinship.Link{…})`, `WithImplied`, `Links()`. `canon.go` `loadCanon`: D-3. Doc comments on both say `kinship.LinkSet` |

```go
// PlaceImplied returns where each implied link goes among s's links: the
// index of the link it merges into, or len(s.Links())+j for the j-th distinct
// new link. It does not change s.
func (s *LinkSet) PlaceImplied(implied []Implied) []int

// Relationships returns id's relationship links, recorded and implied, one per
// pair of ends and keys: each relationship link filed on id, then For's links,
// merged through one kinship.LinkSet. The caller owns the slice and its bytes.
func (r *Reader) Relationships(id []byte) ([]kinship.Link, error)
```

Errors: none new; `Relationships` returns `loadAround`'s and `For`'s errors.

### FFI

None — no RPC (S13-04).

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Move the merge to `core/kinship`**
Files: `core/kinship/links.go` *(new)*, `core/kinship/links_test.go` *(new)*, `core/kinship/kinship.go` (package doc), `core/database/promotealign/links.go` (deleted), `layer.go`, `canon.go`, `export_test.go`, `promotealign_test.go`. Skill: `add-match-rule`.
Do: D-1, D-2, D-3, D-3a, D-7 in one commit (the packages do not build between them). `git mv core/database/promotealign/links.go core/kinship/links.go`, then rename and export; add `PlaceImplied` and its subtest. `loadLayerBridges` and `loadCanon` call the exported names; their doc comments say `kinship.LinkSet`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/ ./core/matchregistry/ ./core/database/canonlinks/ ./core/database/promotealign/ ./core/graphalign/` passes; `go vet ./core/...` clean.

**STEP-2 — Prove the move changed nothing**
Files: none changed.
Do: run the budgets three times and the grep.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -count=3 ./core/database/promotealign/` passes; this grep prints nothing:

```bash
grep -rnwE 'sideLink|linkSet|seenEdge|edgeSeenKey|newLinkSet|impliedAt|SideLink|LinkSetForTest' core/kinship core/database/canonlinks core/database/promotealign
```

**STEP-3 — A person's merged relationships**
Files: `core/database/canonlinks/reader.go`, `canonlinks.go` (package doc), `reader_test.go`.
Do: T-1 first (fails to build without `Relationships`), then D-6, D-4, D-5 and the package doc line. `world.parent` becomes a call to `world.relationship`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -count=1 ./core/database/canonlinks/` passes.

**STEP-4 — Docs**
Files: `docs/matching-model.md` (line 29), `docs/promote-graph-alignment.md` (line 243), `docs/deployment-plan/spike-11/kinship-shapes.md` (line 53), `.agents/skills/add-match-rule/SKILL.md` (line 33), `docs/deployment-plan/spike-11/deployment-plan.md` (S11-51 row status), `docs/deployment-plan/spike-11/completed.md`, this plan (status).
Do: D-8, then the Definition of done's docs lines.
Check: this grep prints nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-51-person-links.md` reports no errors:

```bash
grep -rnwE 'sideLink|linkSet|seenEdge|edgeSeenKey|newLinkSet|SideLink|LinkSetForTest' docs/matching-model.md docs/promote-graph-alignment.md docs/deployment-plan/spike-11/kinship-shapes.md .agents/skills/add-match-rule/SKILL.md
```

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period ("Move the link merge into core/kinship so every reader shares one.", "Return one person's recorded and implied relationships from canonlinks."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/canonlinks/reader_test.go` `TestRelationships` *(new)*, subtests named by AC | AC-1.1 – AC-1.6 as written in §2, built with `world` and `world.relationship` *(new)*. AC-1.1 and AC-1.4 compare ends as a set (each link turned so its lesser name is A, keys turned with it) against `w.keys(relationship, spouse)`; AC-1.2 and AC-1.3 compare A, B and keys in order. Associations compared as sorted names, shapes as (pattern, pivot name), and `OnlyImplied` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, D-4, D-5, D-6 |
| T-2 | Go | `core/kinship/links_test.go` *(new)* `TestLinkSet` (moved) | The five existing cases, unchanged inputs and assertions, plus "PlaceImplied places without growing the set" (AC-2.6) and "an added implied link records as Implied.Recording" (the link `AddImplied` adds carries `imp.Recording()`) | AC-2.2, AC-2.6, D-1, D-3, D-3a, D-7 |
| T-3 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` (unchanged) | The budgets pass, three runs | AC-2.3, D-3 |
| T-4 | Go | `PROVENENCIA_GRAPH_VERIFY=1` over `./core/database/canonlinks/` | Every test passes with verify on | AC-2.4 |
| T-5 | Shell | The greps in STEP-2 and STEP-4 | Print nothing | AC-2.5, D-1, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestProposeKinshipShapes`, `TestLayerBridgesMergeKinshipShapes`, `TestPinOffersOfAnImpliedLayerLink`, the `graphalign` goldens (AC-2.1), `TestFor` and `TestForMatchesDerive` (D-6), the `matchregistry` guard (the moved file in `core/kinship`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — no brief
- [ ] Domain doc updated: matching-model.md, promote-graph-alignment.md, kinship-shapes.md and add-match-rule name kinship.LinkSet (D-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project whose tree files a spouse relationship for a couple and also records their marriage. Promote a Source that records the same couple's marriage. The Promote rows, bands and evidence sheet read as they did on `main` (no behavior change; the engine tests carry the proof).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The warm proposal slows past its budget | D-3 keeps F-9's cost shape; `PlaceImplied` allocates one `[]int` and one map per proposal, as `impliedAt` did. T-3 |
| A caller changes `Links()`'s slice | Doc says read-only; `Relationships` returns a set no one else holds |
| `Relationships` aliases cache bytes | `Add` clones ends, associations and pivots (S11-39 F-5); AC-1.6 |

**Stop and report if:**

- `core/database/promotealign/links.go`, `loadCanon`'s implied merge, `deriver.chainLinks` / `deriver.forLinks`, or `Reader.For` no longer reads as §4 says.
- `TestLargeSourceBudgets` fails at STEP-2 where it passes on `main` in the same run conditions (three runs each).
- Moving the set needs a body change beyond names to keep a `promotealign` test green.
- The `matchregistry` guard rejects `core/kinship/links.go` *(new)*.
- A recorded spouse link and a marriage come back as two links from `Relationships` (AC-1.1): the keys `deriver.edge` writes and the keys `kinship.Derive` writes differ.
- `world.relationship` with `spouse` fails because the test catalog has no `spouse` `relationship_type` term.

## 12. Open questions

None.

## 13. PR

**Title:** S11-51 — One person's links for every reader
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-51-person-links.md`.
