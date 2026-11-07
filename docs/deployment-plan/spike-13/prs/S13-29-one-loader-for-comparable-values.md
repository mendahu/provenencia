# S13-29 — One loader for a Subject's comparable values

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 7 — How a person's records connect
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a Go-only move with no behavior change; nothing on screen reads the new package until S13-28)
**Board:** —
**Depends on:** S13-24 (landed [#639](https://github.com/mendahu/provenencia/pull/639))
**Blocks:** S13-27
**Branch:** `s13-29-one-loader-for-comparable-values`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, and this PR adds no catalog format step ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (Rules that hold everywhere; layer packages stay pure; the guard)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc` (Evidence graph walks and graph-cache memos), `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** working rules "Scoring and the view are two concerns" (slice 7), "Slice 7 starts with a simple count" (slice 7, "A matching link counts"); S13-26 plan [D-2](S13-26-scoring-two-evidence-graphs.md) and its review M-5; [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §8 (Caching and decisions)
**Planned:** 2026-11-07 against `main` @ `97e6f30f`
**Review:** [`S13-29-one-loader-for-comparable-values.review.md`](S13-29-one-loader-for-comparable-values.review.md) — round 1: Fits with changes / Not ready; findings Fixed; M-5 answered by the user 2026-11-07 (move the fixture in this PR)

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

The Person page's record links (S13-27) need, for each Evidence graph a person is in, every primary Subject's comparable values and the graph's links, recorded and implied. Promote already builds exactly that per Evidence graph and memoizes it, inside `promotealign`. This PR moves that loader into a shared `core/database` package, `evidencelayer`, so Promote and the Person page read one memo built once per graph, with links as `kinship.Link` (the type S13-26's `recordlinks.NewGraph` takes) and Subjects as one shared pure type. The Source-building test fixture moves with it into a shared test package, so the loader's tests sit beside the loader and S13-27 builds its records with the same helpers. Promote's proposals do not change.

**Why now.** S13-27 depends on it. Without it S13-27 would add a third loader of member values beside `promotealign`'s layer memo and its canon profiler (map review K-27), or import graph alignment's types into the Person page (refused, S13-26 D-2). Having `persondetail` call Promote's loader instead was refused on 2026-11-07 (map scope entry).

**What stays.** Every Promote proposal, connection and budget. The memo still lives on the Evidence graph and drops with it. The canon side of Promote (`profiler`, `reconciled`, `loadCanon`, the shortlist index) stays in `promotealign`.

## 2. User stories

**US-1.** As a contributor writing S13-27, when I need a person's members' values and links in one Evidence graph, I call `evidencelayer.Load` and get the same memoized layer Promote reads, with links as `[]kinship.Link` and Subjects as `[]subjectcompare.Subject`, so I hand them on with no conversion and import nothing from graph alignment.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Promote proposal ran on Source B's Evidence graph; then `evidencelayer.Load(g, graph, namePattern)` is called | It returns the memoized layer: `&got.Subjects[0]` is the memo's, and no graph query is issued |
| AC-1.2 | `core/database/evidencelayer`'s own production imports | Hold neither `core/graphalign` nor `core/database/promotealign`, and `core/matchregistry`'s guard test fails if one is added |
| AC-1.3 | A Source records Ann and Bob's marriage and a spouse relationship between them; the layer of its graph | `Links` holds one spouse link, a `kinship.Link` with 3 associations and one shape, not `OnlyImplied` (`TestLoadMergesKinshipShapes`, moved from `TestLayerBridgesMergeKinshipShapes` with today's rows) |

**US-2.** As a researcher, when I press Promote on any Source after this PR, I see the same rows, suggestions and connections as before it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Every `promotealign` and `graphalign` test, with only the type changes and fixture renames STEP-1, STEP-3 and STEP-5 name | Pass with their assertions unchanged |
| AC-2.2 | `TestLargeSourceBudgets` (Source of 100 new people) | Each subtest within its budget in `perf_test.go` (`budgetWarm` 375 ms and the rest), unchanged |
| AC-2.3 | `PROVENENCIA_GRAPH_VERIFY=1` over `./core/database/...` | Passes: `TestMemoFollowsCommit` drops the layer memo on the same commits as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `subjectcompare.Subject` *(new)*; `graphalign.Subject` becomes an alias of it (D-2) | STEP-3, AC-2.1 |
| `graphalign.Bridge` is deleted; `graphalign.Layer.Bridges` becomes `[]kinship.Link` (D-7) | STEP-3, AC-2.1 |
| `lifeEvent` moves to `subjectprofile.LifeEventOf` (D-6) | STEP-2, T-1 |
| `loadMetas` moves to `valuerules.Metas` (D-5) | STEP-4 |
| New package `core/database/evidencelayer`: `Load` (memoized), `Layer`, `Layer.Subject`, `Decode`; the loader code from `core/database/promotealign/layer.go`, `core/database/promotealign/terms.go` and the Evidence half of `core/database/promotealign/profiles.go` moves there (D-1 – D-5) | STEP-4, AC-1.1, AC-1.3 |
| `promotealign` reads `evidencelayer.Load`; `primarySubject`, `layerMemo`, `layerMemoKey` and the moved functions go (D-8, D-9) | STEP-5, AC-2.1 – AC-2.3 |
| `core/database/evidencelayer/evidencelayertest` *(new)*: the Source-building fixture from `core/database/promotealign/promotealign_test.go` (user, 2026-11-07; review M-5) (D-13) | STEP-1, AC-2.1 |
| The two loader tests move into `core/database/evidencelayer`; the `LayerBridges` hook goes (D-12) | STEP-6, AC-1.3 |
| The guard learns a forbidden-import rule, with rows for `evidencelayer` and `recordlinks` (D-11) | STEP-6, T-2 |
| The shared-memo test (T-3) | STEP-6 |
| Docs and comments that name the moved symbols (D-14) | STEP-8 |

| Out | Home |
| --- | --- |
| The record-links read, `GetPersonRecordLinks`, and any `persondetail` caller of `Load`; `persondetail`'s guard row | S13-27 |
| Moving the copies of the same helpers in `grapheventtitles`, `canonicalgraph`, `conclusionsummaries`, `promote` (`batch_test.go`) and `graphcache` (`evidence_graph_walk_test.go`) onto the shared fixture | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (STEP-8 adds the entry) |
| Promote-only fixture helpers (`promote`, `propose`, `done`, `join`, `namedMembers`, `filedBirth`, `proposeConnections`, `placeRelationships`) | stay in `promotealign`'s tests (D-13) |
| Folding `graphalign.Handle` into `subjectcompare.Subject`, and `graphalign.CanonEdge` into `kinship.Link` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (review M-3, M-10) |
| `recordlinks.Member` embedding `subjectcompare.Subject` | S13-26 (its plan's call; not this PR) |
| Profile memos for canon handles; the canon profiler | [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Renaming `graphalign.Layer.Bridges` to `Links` | not planned (D-7) |
| Reading catalog-defined (non-product) Properties into the layer | not planned (S13-26 Out, the loader's scope) |
| A multi-graph load for many graphs at once | S13-27 Out, [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer memo | `core/database/promotealign/layer.go` `loadLayerMemo`, `layerMemoKey` (`"promotealign.layer\x00"` + name pattern), `layerMemo{layer, primary}` | Memoizes `graphalign.Layer` and a `map[string]primarySubject` on the Evidence graph memo. Reads the memo before `loadLayer`'s nil check; `graphcache.Graph.EvidenceGraphMemo` is nil-safe (`core/database/graphcache/evidence_graph.go`) |
| Layer build | `core/database/promotealign/layer.go` `loadLayer` | Takes `graphalign.Config` but reads only `cfg.Subject.Weights` (which kinds are primary) and `cfg.Subject.Lifespans`. Builds event values first, then person and place values with their life events, through `subjectprofile.ForSubject`; sorts Subjects by Ref; adds `loadMetas` |
| Links | `core/database/promotealign/layer.go` `loadLayerBridges` | Takes `primary` and `byKind`; walks every `connectrules.Bridges()` entry, keys edges with `segmentcompare.Keys`, merges through `kinship.NewLinkSet()`, `WithImplied(terms, matchregistry.Default().Kinship)`, then copies each `kinship.Link` into a `graphalign.Bridge{Edge, Recording}` (same fields). Also returns `participation`s for lifespans |
| Primary Subjects | `core/database/promotealign/layer.go` `primarySubject{id, ref, kind}` | Read by `mergeFixed`, `checkFixed` (`core/database/promotealign/promotealign.go`) and `pageEnds` (`core/database/promotealign/connections.go`), each for existence and kind; `pageEnds` ranges the map into a map, and `mergeFixed` sorts its output, so order does not matter |
| Term reader | `core/database/promotealign/terms.go` `observedTerms` | Called only from `loadLayerBridges` |
| Value decode | `core/database/promotealign/profiles.go` `decode`, `layerValues`, `loadMetas`, `lifeEvent`; file header "both sides of a pair are built here" | `layerValues` is Evidence-only. `decode` is also called by `reconciled` (canon side); `loadMetas` also by `core/database/promotealign/index.go`; `lifeEvent` also by the canon profiler (`profiles.go` ~301). `hopPersonToEvent` (declared in `layer.go`) is read only by the canon profiler (`profiles.go` ~217, ~295) |
| Config | `core/database/promotealign/promotealign.go` `loadConfig`, `loadInputs` | `cfg := matchregistry.Default().AlignConfig()` with the catalog's value rules; `AlignConfig` copies `r.Subject` unchanged (`core/matchregistry/registry.go`). `loadInputs` returns `promote.ErrInvalid` for a nil catalog before any load. `inputs` carries `layer` and `primary` |
| Types | `core/graphalign/types.go` `Subject{ID, Ref, Kind string, Values}`, `Bridge{kinship.Edge; kinship.Recording}`, `Layer{Subjects, Bridges []Bridge, Metas}`, `CanonEdge` comment "Recording is as on Bridge" | `Bridge` has the fields of `kinship.Link` (`core/kinship/links.go`). `align.go` ranges `layer.Bridges` and reads only embedded fields. Non-test code outside `layer.go` names `graphalign.Subject` in `index.go` and `canon.go` |
| Registry imports | `core/matchregistry/registry.go`, `core/matchregistry/alignment.go` | Import `core/graphalign` (`Registry.Alignment`, `AlignConfig`). Any package that reads `matchregistry` reaches `graphalign` transitively |
| Guard | `core/matchregistry/guard_test.go` `layerPackages`, `checkPure`, `TestLayerPackages` | One walker over pure layer packages (no `core/database`, no numbers); no forbidden-import rule for a reader package |
| Test hooks | `core/database/promotealign/export_test.go` `LayerBridges`, `MemoSnapshot{Layer, LayerKind, LayerOK}`, `MemoSnapshotForTest` | `LayerBridges` calls the unmemoized `loadLayer`; `MemoSnapshotForTest` reads `layerMemoKey`/`layerMemo`. `LayerKind` is read only through `memoJSON` |
| Tests | `core/database/promotealign/memo_test.go` `TestMemoIsNotMutated`, `TestMemoFollowsCommit`, `TestProposeWarmIssuesNoQuery`; `core/database/promotealign/promotealign_test.go` `TestLayerBridgesMergeKinshipShapes` (helpers typed `[]graphalign.Bridge`), `TestLayerBridgesCoverEveryConnectRule`; `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`; `core/graphalign/align_test.go` helpers `bridge`, `implied`, `canonOf`, `side`; `core/graphalign/invariants_test.go` | Build or hold `graphalign.Bridge` values; pin the memo's identity and invalidation |
| Fixture | `core/database/promotealign/promotealign_test.go` `fixture{t, c, source, artifact, y}`, `newFixture`, `must`, `userID`, `locator`, and the record builders `newSource`, `prop`, `term`, `bare`, `cite`, `person`, `birth`, `place`, `participation`, `locate`, `placeRel`, `relationship`, `sexOnly`, `graph`, `marriage`, `eventOn`, `datedOn`, `datedBirth` | Package `promotealign_test`; used about 400 times across `promotealign_test.go`, `memo_test.go`, `perf_test.go`, `bench_test.go`, `source_store_test.go`. `core/database/evidencegraphs/evidencegraphstest` is imported by in-package tests of `connect`, `citations`, `subjects`, `sourcecopy` and `graphprogress`, so it cannot import `connect` or `citations` without an import cycle |
| Docs | `.agents/skills/add-match-rule/SKILL.md` line 15 ("`core/database/promotealign` builds each side with the profile builder") and line 33 (`graphalign.Bridge` twice); `docs/ideas/matching-test-suite.md` (names `neighborTypeProperty` in `layer.go`); `docs/ideas/promote-ranking.md` (names `loadLayerMemo`); `core/graphalign/types.go` `Bridge` comment | Name what this PR moves or deletes |

**Precedent.** `a8dc05b4` "Move the link merge into core/kinship so every reader shares one." moved Promote's `links.go` (deleted) into `core/kinship/links.go` and pointed Promote at it. Copy its shape: move, re-point, delete the old file in the same commit. `core/database/canonlinks` is the precedent for the package's role: a `core/database` reader that "matching and later readers share", imported by both `promotealign` and `persondetail`. The guard's table shape (`layerPackages`, one walker) is the precedent for D-11. `core/database/graphcache/graphcachetest` ("builds catalogs and canonical graphs for tests of the graph cache and of the packages that read it") is the precedent for D-13: a `<pkg>test` package under the package it serves, taking `testing.TB`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The package is `core/database/evidencelayer` *(new)*: `evidencelayer.go` holds `Load`, `Layer`, `Layer.Subject`, the build and the links walk; `values.go` holds `Decode`, `layerValues`, `observedTerms`. Package doc: "Package evidencelayer loads one Evidence graph's primary Subjects with their comparable values and its links, recorded and implied, memoized on the graph. Promote and the Person page share it. Decode, the one row decoder, also serves Promote's canon side, since a cached handle value carries the same value columns." | "Evidence layer" is the term Promote's code and `promote-graph-alignment.md` already use for one Source's primary Subjects and bridges | `subjectvalues` — the package also owns links. Putting it in `canonlinks` — that package reads the canon, not an Evidence graph |
| D-2 | `subjectcompare.Subject` *(new)*: `type Subject struct { ID []byte; Ref string; Kind string; Values Values }`, doc "one primary Subject or handle with its comparable values". `core/graphalign/types.go`: `type Subject = subjectcompare.Subject` (alias; every use compiles unchanged). `evidencelayer.Layer{ Subjects []subjectcompare.Subject; Links []kinship.Link; byID map[string]int }`, Subjects sorted by Ref; `func (l Layer) Subject(id []byte) (subjectcompare.Subject, bool)` reads `byID` | One Subject-with-values type for graph alignment, the Evidence layer and S13-27, owned by the pure package that owns `Values`. Promote passes the memo's slice straight through (D-8), so no per-call copy | A third identical struct in `evidencelayer` with a per-call conversion (round 1 draft; review M-3). Returning `graphalign.Layer` — puts graph alignment's types in the Person page's path (S13-26 D-2) |
| D-3 | `func Load(g *graphcache.Graph, graphID []byte, namePattern string) (Layer, error)`. A nil `g` returns `errNilGraph` *(new, unexported, `errors.New("evidencelayer: nil graph")`)* first. Then it reads the memo at `memoKey + namePattern`, else builds and stores the `Layer` value itself. On a miss the build reads `matchregistry.Default().Subject`, `matchregistry.Default().Kinship` and `canonlinks.LoadTerms(g)` itself | One memo key for every reader is only safe if every reader builds with the same inputs: the name pattern is in the key; rules are static; terms come from the vocabulary, whose commit drops the memo. No input a caller could pass differently is left. `promotealign.loadInputs` already maps a nil catalog to `promote.ErrInvalid` before calling, so the shared reader needs no Promote error | Rules or terms as parameters (round 1 draft, review M-2). `promote.ErrInvalid` — a write package's error code in the Person page's path (review M-8). Loading `namePattern` inside — `valuerules.Load` needs a `database.Querier` the graph does not give |
| D-4 | `const memoKey = "evidencelayer\x00"` *(new)*, unexported. The old `"promotealign.layer\x00"` key goes. `promotealign`'s `MemoSnapshotForTest` reads the layer through `Load` (the memo exists after a Propose; T-3 pins that a warm `Load` issues no query) | No production symbol for tests (review M-7) | Exporting the key or a `Peek` |
| D-5 | `loadMetas` moves to `core/database/valuerules` as `func Metas(g *graphcache.Graph) (subjectcompare.Metas, error)`, body unchanged. `Layer` has no `Metas`; each caller reads them once per call (`promotealign.loadInputs` into `graphalign.Layer.Metas`, `index.go`). `decode` moves to `evidencelayer.Decode` (exported; `reconciled` calls it) | Metas are catalog-wide compare inputs, the kind `valuerules`' package doc says are "loaded in one place"; `loadLayer`'s own comment says they "are not part of one graph". S13-27 then reads them once, not per graph (review M-4). One decoder for both sides stays one | Keeping `Metas` on the per-graph layer — 50 identical maps at S13-27's Scale. Copying `decode` — two decoders for one row shape |
| D-6 | `lifeEvent` moves to `core/subjectprofile` as `func LifeEventOf(event subjectcompare.Values, role vocabkey.Role) LifeEvent` *(new name)*, body and comment unchanged. `evidencelayer` and the canon profiler call it | Pure, and builds `subjectprofile.LifeEvent`: the owning layer. No numeric literal, so the guard's `numbers` rule still passes | Exporting it from `evidencelayer` — the canon profiler would read a pure helper through a database package |
| D-7 | `graphalign.Bridge` is deleted. `graphalign.Layer.Bridges` keeps its name and becomes `[]kinship.Link`. `CanonEdge`'s comment reads "Recording is as on kinship.Link." | The user's answer to S13-26 review M-5 (2026-11-07). Field access compiles unchanged, so only constructors move | Renaming the field to `Links` — dozens of test edits; `graphalign` calls them bridges throughout |
| D-8 | `promotealign.loadInputs` calls `evidencelayer.Load(c.Graph(), graphID, namePattern)` and `valuerules.Metas(c.Graph())`, and builds `graphalign.Layer{Subjects: el.Subjects, Bridges: el.Links, Metas: metas}` — the memo's slices, not copies. `inputs` gains `evidence evidencelayer.Layer` and loses `primary`. `terms` stays in `inputs` from Promote's own `canonlinks.LoadTerms` call, for the canon side | Align, `loadCanon` and `loadCandidatesMemo` keep their signatures; no reader changes a memoized slice (`TestMemoIsNotMutated`) | A per-call copy (D-2 rejected) |
| D-9 | `mergeFixed(g, graphID, layer evidencelayer.Layer, caller)`, `checkFixed(g, layer, f)` and `pageEnds(promoted, layer, caller, rows, only)` take the Evidence layer. Existence and kind read `layer.Subject(id)`; the two `for sk := range primary` loops become `for _, s := range layer.Subjects { sk := string(s.ID) … }`; `mergeFixed`'s capacity uses `len(layer.Subjects)` | `primarySubject` was a second index of the same Subjects. Output order is unchanged (map output; sorted output) | A `primary` map rebuilt per call |
| D-10 | Inside `evidencelayer`, the build keeps an unexported local `primary map[string]subjectcompare.Subject` and `byKind map[string][][]byte` as today's `loadLayer` does; `links(g, eg, props, vocabTerms, primary, byKind, terms) ([]kinship.Link, []participation, error)` returns `set.Links()`. `byID` is filled after the Ref sort, `byID[string(s.ID)] = i`. `participation` and `neighborTypeProperty` move into the package; `hopPersonToEvent` moves from `layer.go` to `profiles.go`, beside its only readers | Moves the code as it is; only the value types change | Reshaping the walk |
| D-11 | `core/matchregistry/guard_test.go` gains `readerPackages = []struct{ dir string; forbids []string }` and `TestReaderPackages`, run by a `checkForbids` beside `checkPure` that reads each package's **direct** production imports (`pkg.Imports`) through the same `go/build` context and module root. Rows: `{"core/database/evidencelayer", {"core/graphalign", "core/database/promotealign"}}`, `{"core/recordlinks", {"core/graphalign"}}` — the second only if `core/recordlinks` exists at the base (S13-26 may land first; otherwise S13-26 adds it). S13-27 adds `persondetail` | One guard table instead of a second walker (review M-1). Direct imports, because `matchregistry` itself imports `graphalign` (`Registry.Alignment`), so every registry reader reaches it transitively (review X-1); the working rule is about the reader's own use of graph alignment's types | A transitive rule — fails on `matchregistry`. A test in `evidencelayer` with its own walker |
| D-12 | `TestLayerBridgesMergeKinshipShapes` and `TestLayerBridgesCoverEveryConnectRule` move to `core/database/evidencelayer/evidencelayer_test.go` *(new)*, package `evidencelayer_test`, renamed `TestLoadMergesKinshipShapes` and `TestLoadCoversEveryConnectRule`, assertions unchanged. They build with `evidencelayertest`, read the name pattern with `valuerules.Load(c.Graph(), db)`, and call `evidencelayer.Load`; each builds a fresh catalog, so the memo does not carry between them. `promotealign`'s `LayerBridges` hook is deleted. `MemoSnapshot` loses `LayerKind`; `Layer` becomes `evidencelayer.Layer` | The tests sit with the code they test, as `a8dc05b4` did (review M-5; user, 2026-11-07) | Keeping them in Promote through a forwarding hook (round 1 draft) |
| D-13 | `core/database/evidencelayer/evidencelayertest` *(new)*, one file `evidencelayertest.go`. Package doc: "Package evidencelayertest builds catalogs whose Sources hold people, events, places and the bridges between them, for tests of the Evidence layer and the packages that read it." `type Records struct { T testing.TB; C *database.Catalog; Source sources.Source; Artifact artifacts.Artifact; y int64 }`; `func New(t testing.TB) *Records` (today's `newFixture` body); `var UserID = []byte{1, …, 16}` and `const Locator` (today's values); `func Must(t testing.TB, err error)`. Methods, exported, bodies unchanged: `NewSource`, `Prop`, `Term`, `Bare`, `Cite`, `Person`, `Birth`, `Place`, `Participation`, `Locate`, `PlaceRel`, `Relationship`, `SexOnly`, `Graph`, `Marriage`, `EventOn`, `DatedOn`, `DatedBirth`. `promotealign_test`'s `fixture` embeds `*evidencelayertest.Records` and keeps its own `c`, `source`, `artifact` fields, set from the Records in `newFixture`, so their reads do not change; `userID = evidencelayertest.UserID`, `must = evidencelayertest.Must`, `locator = evidencelayertest.Locator`. Call sites of the moved methods are renamed mechanically, one `gofmt -r` per method (`gofmt -r 'f.person(a, b, c) -> f.Person(a, b, c)' -w core/database/promotealign/`, and the same for each arity and receiver name the files use, such as `lf.fixture`, `w.f`) | A shared Evidence-side fixture beside the Evidence layer, the way `graphcachetest` serves the canon side. `evidencegraphstest` cannot hold it (import cycle, §4). Keeping `c`, `source`, `artifact` as fields bounds the diff to the moved method names | Extending `evidencegraphstest` — cycles. Lower-case wrappers in `promotealign` forwarding to the exported methods — a second copy of every name. Renaming every `f.c` to `f.C` — a large diff with no reader gain |
| D-14 | Docs: see STEP-8 | Every doc that names a moved or deleted symbol is updated in this PR (review M-6) | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: "Evidence layer", "layer memo", `EvidenceGraphMemo` callers, `decode`/`UnmarshalDate` readers in `core/database`, `LifeEvent` builders, structs of `{ID, Ref, Kind, Values}`, `kinship.Link` and `graphalign.Bridge` constructors, import guards.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/evidencelayer` *(new)* | package | `core/database/promotealign` (layer half); `core/database/canonlinks` | The same code moved; `canonlinks` reads the canon |
| `evidencelayer.Load`, `Layer`, `Layer.Subject` *(new)* | func / type | `promotealign.loadLayerMemo` + `loadLayer`; `primarySubject` | Moved and merged; `primarySubject` goes |
| `subjectcompare.Subject` *(new)* | type | `graphalign.Subject` (becomes its alias), `graphalign.Handle` | Moved down to the owning pure package; `Handle` is logged (Out) |
| `evidencelayer.Decode`, `valuerules.Metas`, `subjectprofile.LifeEventOf` *(new names)* | func | `promotealign.decode`, `loadMetas`, `lifeEvent` | Moved (D-5, D-6) |
| `readerPackages`, `checkForbids`, `TestReaderPackages` *(new)* | test helper | `guard_test.go` `layerPackages`, `checkPure` | Extends the one guard (D-11) |
| `TestLifeEventOf`, `TestLoadSharesPromotesMemo` *(new)* | test | `memo_test.go` `TestMemoIsNotMutated` | New cases for moved code |
| `evidencelayertest.Records`, `New`, its builder methods *(new)* | test helper | `promotealign_test.go` `fixture`; `graphcachetest`; `evidencegraphstest`; the copies in five other packages (Out) | Moved from `promotealign`; `evidencegraphstest` cannot hold it (D-13) |
| `TestLoadMergesKinshipShapes`, `TestLoadCoversEveryConnectRule` *(new names)* | test | `TestLayerBridges*` | Moved (D-12) |

**Builds on.** `graphcache.Graph.EvidenceGraphMemo` / `SetEvidenceGraphMemo` (the memo drops with its graph, `.cursor/rules/catalog-queries.mdc`); `kinship.LinkSet`; `subjectprofile.ForSubject`; `canonlinks.LoadTerms`; the `matchregistry` guard table.
**Deletes.** `core/database/promotealign/layer.go` and `core/database/promotealign/terms.go` (whole files); from `core/database/promotealign/profiles.go`: `decode`, `layerValues`, `loadMetas`, `lifeEvent`; `primarySubject`, `layerMemo`, `layerMemoKey`; `inputs.primary`; `MemoSnapshot.LayerKind`; the `LayerBridges` hook; the moved fixture code in `promotealign_test.go`; `graphalign.Bridge`; `graphalign.Subject`'s struct body (now an alias).
**Next adds.** S13-27 calls `evidencelayer.Load` per graph and `valuerules.Metas` once, passes `Links` to `recordlinks.NewGraph`, adds a `persondetail` row to `readerPackages`, and builds its handler test's records with `evidencelayertest`.

### Reach

None — no new path into existing code, no new state reachable. `Load` has no caller outside `promotealign` until S13-27.

### Scale

At the map's Scale for Promote (a Source of 100 new people, the 41-Subject obituary): no write changes; the memo key's prefix changes and drops on the same events (`TestMemoFollowsCommit`). Per proposal Promote now reads `valuerules.Metas` (a loop over the vocabulary's Properties, as `index.go` already does per build) instead of the memoized copy, and skips the `primary` map; no slice is copied. `TestLargeSourceBudgets` measures it (AC-2.2). Slice 7's Scale (50 Evidence graphs) is S13-27's budget subtest; D-5 means it reads metas once, not 50 times.

### Edges

- **Assumptions.** (1) One memo for two readers is right only if both build with the same inputs: D-3 leaves only the name pattern as an input, and it is in the key. T-3 holds the sharing. (2) Readers never change a memoized `Layer`: Promote now passes the memo's `Subjects` and `Links` slices to `Align` directly; `TestMemoIsNotMutated` (JSON before and after two proposals) holds it. (3) `pageEnds` and `mergeFixed` are order-free; `TestProposeWarmIssuesNoQuery`'s `reflect.DeepEqual` holds it.
- **Neighbors.** None — no rule changes; the same inputs reach the same code.
- **Removed inputs.** None — nothing is decided, excluded or deleted by this PR.
- **Real data.** None — no table-driven rule changes.
- **Untrusted size.** None — the walk is today's, bounded by `kinship.Rules.MaxPivotEnd` and `MaxPairSteps`, unchanged.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/subjectcompare` | `Subject` *(new)* (D-2) |
| `core/subjectprofile` | `LifeEventOf` *(new name)* (D-6) |
| `core/graphalign` | `Subject` becomes an alias; `Bridge` deleted; `Layer.Bridges []kinship.Link` (D-2, D-7) |
| `core/database/valuerules` | `Metas` *(new name)* (D-5) |
| `core/database/evidencelayer` *(new)* | `Load`, `Layer`, `Layer.Subject`, `Decode`; unexported `memoKey`, `errNilGraph`, `build`, `links`, `layerValues`, `observedTerms`, `participation`, `neighborTypeProperty` (D-1, D-3, D-4, D-10) |
| `core/database/promotealign` | `loadInputs` (D-8), `mergeFixed`, `checkFixed`, `pageEnds` (D-9); `reconciled`, `index.go`, the canon profiler call `evidencelayer.Decode`, `valuerules.Metas`, `subjectprofile.LifeEventOf`; `hopPersonToEvent` moves to `profiles.go` |

```go
package evidencelayer

type Layer struct {
	Subjects []subjectcompare.Subject // primary Subjects, by Ref
	Links    []kinship.Link           // recorded and implied, one per ends and keys
	byID     map[string]int
}

func (l Layer) Subject(id []byte) (subjectcompare.Subject, bool)
func Load(g *graphcache.Graph, graphID []byte, namePattern string) (Layer, error)
func Decode(scalar propertymeta.Scalar, termID, date, name, image []byte, terms map[string]graphcache.VocabTerm) (autoreconcile.Value, error)
```

Errors: `errNilGraph` for a nil graph; graph-cache errors pass through.

### FFI

None — no wire change.

### macOS

None — Go only.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The shared Source fixture**
Files: `core/database/evidencelayer/evidencelayertest/evidencelayertest.go` *(new)*, `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/memo_test.go`, `core/database/promotealign/perf_test.go`, `core/database/promotealign/bench_test.go`, `core/database/promotealign/source_store_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: D-13. Move the builders, then rename call sites with `gofmt -r`. No test body changes beyond the renamed calls.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes, and `go test -tags fts5 -run XXX -bench . -benchtime 1x ./core/database/promotealign/` compiles and runs; `rg -n 'func \(f \*fixture\) (person|birth|place|participation|relationship|marriage)\(' core/database/promotealign` prints nothing.

**STEP-2 — `LifeEventOf` in subjectprofile**
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/lifeevent_test.go` *(new)*, `core/database/promotealign/profiles.go`, `core/database/promotealign/layer.go`. Skill: `add-match-rule` (Rules that hold everywhere).
Do: D-6; point both callers at it; add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/subjectprofile/... ./core/matchregistry/... ./core/database/promotealign/...` passes.

**STEP-3 — One Subject type and one link type**
Files: `core/subjectcompare/subjectcompare.go`, `core/graphalign/types.go`, `core/graphalign/align_test.go`, `core/graphalign/invariants_test.go`, `core/database/promotealign/layer.go`, `core/database/promotealign/export_test.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-2 (the type and the alias) and D-7. In `layer.go`, `loadLayerBridges` returns `set.Links()` as `[]kinship.Link`. `LayerBridges` returns `([]kinship.Link, error)`; `TestLayerBridgesMergeKinshipShapes`'s `spouses` and `build` take `[]kinship.Link`; the `graphalign` test helpers `bridge`, `implied`, `canonOf`, `side` and the `bridges` table fields use `kinship.Link`. No assertion changes.
Check: `rg -n 'graphalign\.Bridge\b' core api` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/subjectcompare/... ./core/graphalign/... ./core/matchregistry/... ./core/database/promotealign/...` passes.

**STEP-4 — The evidencelayer package and `valuerules.Metas`**
Files: `core/database/evidencelayer/evidencelayer.go` *(new)*, `core/database/evidencelayer/values.go` *(new)*, `core/database/valuerules/valuerules.go`. Rule: `.cursor/rules/catalog-queries.mdc`.
Do: D-1, D-3, D-5 (`Metas` and `Decode`), D-10. Nothing calls `Load` yet.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/...` and `CGO_ENABLED=1 go vet -tags fts5 ./core/database/evidencelayer/... ./core/database/valuerules/...` pass.

**STEP-5 — Promote reads the shared loader**
Files: `core/database/promotealign/promotealign.go`, `core/database/promotealign/connections.go`, `core/database/promotealign/profiles.go`, `core/database/promotealign/index.go`, `core/database/promotealign/export_test.go`; delete `core/database/promotealign/layer.go` and `core/database/promotealign/terms.go`.
Do: D-8, D-9, the `MemoSnapshot` half of D-12; `reconciled` calls `evidencelayer.Decode`; `index.go` calls `valuerules.Metas`; `hopPersonToEvent` moves to `profiles.go`. Fix the `profiles.go` header comment to "The canon side's profile loader: each handle is built here, once per match index, with the profile builder; the Evidence side is built by core/database/evidencelayer."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes with no assertion edited.

**STEP-6 — Guard, shared-memo and loader tests**
Files: `core/matchregistry/guard_test.go`, `core/database/promotealign/memo_test.go`, `core/database/evidencelayer/evidencelayer_test.go` *(new)*, `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/export_test.go`.
Do: D-11 (copy `checkPure`'s module-root lookup, `filepath.Join("..", "..")`, and its `go/build` context); add T-2's rows and T-3; D-12 (move the two loader tests, delete `LayerBridges`).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestReaderPackages|TestLayerPackages' ./core/matchregistry/` and `-run TestLoadSharesPromotesMemo ./core/database/promotealign/` and `./core/database/evidencelayer/...` pass; `rg -n LayerBridges core` prints nothing; adding `_ "github.com/mendahu/provenencia/core/graphalign"` to `evidencelayer.go` makes `TestReaderPackages` fail (then remove it).

**STEP-7 — Budgets and verify mode**
Files: none changed.
Do: run the budget test on this branch and on `main` in a second worktree (`git worktree add ../provenencia-main main`), same machine, back to back.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -v ./core/database/promotealign/` passes on the branch, and both runs' `t.Logf` lines go in the PR description; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-8 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md` (line 15: "`core/database/evidencelayer` builds an Evidence graph's side and `core/database/promotealign` the canon's, each with the profile builder"; line 33: both `graphalign.Bridge` mentions become `kinship.Link`), `docs/promote-graph-alignment.md` §8 (one bullet: "**The Evidence layer memo** (`evidencelayer.Load`) holds one Evidence graph's primary Subjects, values and links; Promote and the Person page share it, and it drops with the Evidence graph memo"), `docs/ideas/matching-test-suite.md` (`neighborTypeProperty` now in `core/database/evidencelayer/evidencelayer.go` *(new)*), `docs/ideas/promote-ranking.md` (`loadLayerMemo` → `evidencelayer.Load`), `docs/ideas/structural-debt.md` (two entries: `graphalign.Handle` → `subjectcompare.Subject` and `graphalign.CanonEdge` → `kinship.Link`; the five packages that still copy the Source-building helpers move onto `evidencelayertest`), `docs/deployment-plan/spike-13/deployment-plan.md` (map row), `docs/deployment-plan/spike-13/completed.md`, this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Read Promote's Evidence layer from one loader the Person page can share."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectprofile/lifeevent_test.go` `TestLifeEventOf` *(new)* | Table, values built with `helpers_test.go` `vcand`: positive `event_type` "birth" and negative "death" → `EventType` birth; positive `date`, `start_date`, `end_date` → all three in `Dates` in that key order; a negative date → left out; no type → empty `EventType`; role passed through | D-6 |
| T-2 | Go | `core/matchregistry/guard_test.go` `TestReaderPackages` *(new)* | Each `readerPackages` row → its direct production imports hold none of its `forbids` | AC-1.2, D-11 |
| T-3 | Go | `core/database/promotealign/memo_test.go` `TestLoadSharesPromotesMemo` *(new)* | Warm fixture: Propose on Source B's graph; then, with `SetQueryCounterForTest` counting, `evidencelayer.Load(g, graph, namePattern)` → 0 queries, and a second `Load` returns the same `&Subjects[0]` | AC-1.1, D-3, D-4 |
| T-4 | Go | `core/database/evidencelayer/evidencelayer_test.go` `TestLoadMergesKinshipShapes`, `TestLoadCoversEveryConnectRule` *(new)*, moved | Today's rows, over `evidencelayer.Load`'s `Links` | AC-1.3, D-7, D-12 |
| T-8 | Go | every `core/database/promotealign` test and benchmark on `evidencelayertest` | As today | AC-2.1, D-13 |
| T-5 | Go | `core/database/promotealign/memo_test.go` `TestMemoIsNotMutated`, `TestMemoFollowsCommit`, `TestProposeWarmIssuesNoQuery` (unchanged assertions) | As today | AC-2.1, AC-2.3, D-8, D-9 |
| T-6 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` (unchanged) | As today | AC-2.2 |
| T-7 | Go | `core/graphalign` all tests (types only) | As today | AC-2.1, D-2, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (export hooks), `core/graphalign` (test helpers), `core/matchregistry` `guard_test.go`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks), run on a Mac
- [ ] PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/... passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: `rg` across `.agents/`, `.cursor/rules/`, `docs/` (recursive, `docs/ideas/` included) and `L10n.swift` finds no hit for `graphalign.Bridge`, `loadLayerMemo`, `loadLayerBridges`, `promotealign.layer`, `layerMemoKey`, `primarySubject`, `promotealign/layer.go`, `promotealign/terms.go`, `promotealign.decode`, `promotealign.loadMetas`, `LayerBridges`, and no doc that places `neighborTypeProperty`, `observedTerms`, `layerValues` or `lifeEvent` in `promotealign`; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-graph-alignment.md §8 (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check, TestLargeSourceBudgets timings on main and on the branch
```

## 10. Manual check

1. Open a project with a Source of several people (the 41-Subject obituary if available). Open Promote on it on `main` and on this branch. See the same rows, the same suggested targets and the same connections.
2. Press Done on one row, reopen Promote. See the promoted row anchored as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Promote now reads metas per proposal instead of from the memo | T-6 at Scale; timings on `main` and the branch in the PR description |
| A reader mutates the memo's slices now that Promote passes them through | `TestMemoIsNotMutated` |
| `subjectprofile` or `subjectcompare` gains a database import | `TestLayerPackages` fails |

**Stop and report if:**

- `loadLayer` reads any `graphalign.Config` field other than `Subject.Weights` and `Subject.Lifespans` (D-3 would change Promote's output).
- `loadConfig` sets `cfg.Subject` to anything but `matchregistry.Default().Subject`.
- Promote's terms (`canonlinks.LoadTerms`) are anything but what `Load` reads itself.
- A non-test file outside `core/database/promotealign/layer.go` constructs `graphalign.Bridge`.
- `TestLargeSourceBudgets` fails a budget on the branch that passes on `main` in the back-to-back run.
- Any `promotealign` or `graphalign` test needs an assertion change, not only a type change.
- `evidencelayertest` cannot build without an import cycle (a package it imports has an in-package test that imports it).
- `evidencelayer` cannot build without a direct import of `core/graphalign` or `core/database/promotealign`.

## 12. Open questions

None.

## 13. PR

**Title:** S13-29 — One loader for a Subject's comparable values
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1 – T-7, the verify run, the budget timings on `main` and the branch); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-29-one-loader-for-comparable-values.md`.
