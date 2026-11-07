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
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (graph cache memos), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (Rules that hold everywhere; layer packages stay pure)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** working rules "Scoring and the view are two concerns" (slice 7), "Slice 7 starts with a simple count" (slice 7, "A matching link counts"); S13-26 plan [D-2](S13-26-scoring-two-evidence-graphs.md) and its review M-5; [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §8 (Caching and decisions)
**Planned:** 2026-11-07 against `main` @ `97e6f30f`
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

The Person page's record links (S13-27) need, for each Evidence graph a person is in, every primary Subject's comparable values and the graph's links, recorded and implied. Promote already builds exactly that per Evidence graph and memoizes it, inside `promotealign`. This PR moves that loader into a shared `core/database` package, `evidencelayer`, so Promote and the Person page read one memo built once per graph, and its links come out as `kinship.Link`, the type S13-26's `recordlinks.NewGraph` takes. Promote's proposals do not change.

**Why now.** S13-27 depends on it. Without it S13-27 would add a third loader of member values beside `promotealign`'s layer memo and its canon profiler (map review K-27), or import graph alignment's types into the Person page (refused, S13-26 D-2). Having `persondetail` call Promote's loader instead was refused on 2026-11-07 (map scope entry).

**What stays.** Every Promote proposal, connection and budget. The memo still lives on the Evidence graph and drops with it. The canon side of Promote (`profiler`, `loadCanon`, the shortlist index) stays in `promotealign`.

## 2. User stories

**US-1.** As a contributor writing S13-27, when I need a person's members' values and links in one Evidence graph, I call `evidencelayer.Load` and get the same memoized layer Promote reads, with links as `[]kinship.Link`, so I hand them to `recordlinks.NewGraph` with no conversion and import nothing from graph alignment.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Promote proposal ran on Source B's Evidence graph; then `evidencelayer.Load` is called for the same graph, name pattern and terms | It returns the memoized layer: `&got.Subjects[0]` is the memo's `&Subjects[0]`, and no graph query is issued |
| AC-1.2 | `go list -deps ./core/database/evidencelayer` | The list holds no `core/graphalign` and no `core/database/promotealign` |
| AC-1.3 | A Source records Ann and Bob's marriage and a spouse relationship between them; `Load` on its graph | `Layer.Links` holds one spouse link from Ann to Bob, a `kinship.Link` with 3 associations and one shape, not `OnlyImplied` (the shape `TestLayerBridgesMergeKinshipShapes` pins today) |

**US-2.** As a researcher, when I press Promote on any Source after this PR, I see the same rows, suggestions and connections as before it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Every `promotealign` and `graphalign` test, with only the type changes STEP-4 and STEP-5 name | Pass with their assertions unchanged |
| AC-2.2 | `TestLargeSourceBudgets` (Source of 100 new people) | Each subtest within its budget in `perf_test.go` (`budgetWarm` 375 ms and the rest), unchanged |
| AC-2.3 | `PROVENENCIA_GRAPH_VERIFY=1` over `./core/database/...` | Passes: `TestMemoFollowsCommit` drops the layer memo on the same commits as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| New package `core/database/evidencelayer`: `Load` (memoized), `Layer`, `Subject`, `MemoKey`, `Decode`, `LoadMetas`; the loader code from `core/database/promotealign/layer.go`, `terms.go` and the Evidence half of `profiles.go` moves there (D-1 – D-5) | STEP-2, AC-1.1, AC-1.3 |
| `lifeEvent` moves to `subjectprofile.LifeEventOf` (D-6) | STEP-1, T-1 |
| `graphalign.Bridge` is deleted; `graphalign.Layer.Bridges` becomes `[]kinship.Link` (D-7) | STEP-4, AC-2.1 |
| `promotealign` reads `evidencelayer.Load`; `primarySubject`, `layerMemo`, `layerMemoKey` and the moved functions go (D-8, D-9) | STEP-3, AC-2.1 – AC-2.3 |
| The import guard and shared-memo tests (T-2, T-3) | STEP-6 |
| Docs that name the moved symbols (`add-match-rule`, `promote-graph-alignment.md` §8, `canonlinks` package doc) | STEP-7 |

| Out | Home |
| --- | --- |
| The record-links read, `GetPersonRecordLinks`, and any `persondetail` caller of `Load` | S13-27 |
| Profile memos for canon handles; the canon profiler (`profiler`, `reconciled`) | not planned here: [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Renaming `graphalign.Layer.Bridges` to `Links` | not planned (D-7) |
| Reading catalog-defined (non-product) Properties into the layer | not planned (S13-26 Out, the loader's scope) |
| A multi-graph load for many graphs at once | S13-27 Out, [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer memo | `core/database/promotealign/layer.go` `loadLayerMemo`, `layerMemoKey` (`"promotealign.layer\x00"` + name pattern), `layerMemo{layer, primary}` | Memoizes `graphalign.Layer` and a `map[string]primarySubject` on the Evidence graph memo. Only Promote can read it; the value type is `graphalign`'s |
| Layer build | `core/database/promotealign/layer.go` `loadLayer` | Takes `graphalign.Config` but reads only `cfg.Subject.Weights` (which kinds are primary) and `cfg.Subject.Lifespans`. Builds event values first, then person and place values with their life events, through `subjectprofile.ForSubject`; sorts Subjects by Ref; adds `loadMetas` |
| Links | `core/database/promotealign/layer.go` `loadLayerBridges` | Walks every `connectrules.Bridges()` entry, keys edges with `segmentcompare.Keys`, merges through `kinship.NewLinkSet()`, `WithImplied(terms, matchregistry.Default().Kinship)`, then copies each `kinship.Link` into a `graphalign.Bridge{Edge, Recording}` (same fields). Also returns `participation`s for lifespans |
| Primary Subjects | `core/database/promotealign/layer.go` `primarySubject{id, ref, kind}` | Read by `mergeFixed`, `checkFixed` (`promotealign.go`) and `pageEnds` (`connections.go`), each for existence and kind; `pageEnds` ranges the map, and its result is a map, so order does not matter. `mergeFixed` sorts its output |
| Term reader | `core/database/promotealign/terms.go` `observedTerms` | Called only from `loadLayerBridges` |
| Value decode | `core/database/promotealign/profiles.go` `decode`, `layerValues`, `loadMetas`, `lifeEvent` | `layerValues` is Evidence-only. `decode` is also called by `reconciled` (canon side, same file); `loadMetas` also by `index.go`; `lifeEvent` also by the canon profiler (`profiles.go` line ~301) |
| Config | `core/database/promotealign/promotealign.go` `loadConfig`, `loadInputs` | `cfg := matchregistry.Default().AlignConfig()` with the catalog's value rules; `AlignConfig` copies `r.Subject` unchanged (`core/matchregistry/registry.go`). `inputs` carries `layer` and `primary` |
| Bridge type | `core/graphalign/types.go` `Bridge`, `Layer.Bridges []Bridge`, `CanonEdge` comment "Recording is as on Bridge" | `Bridge` embeds `kinship.Edge` and `kinship.Recording`: the fields of `kinship.Link` (`core/kinship/links.go`). `align.go` ranges `layer.Bridges` and reads only embedded fields. No non-test code outside `core/database/promotealign/layer.go` builds a `Bridge` |
| Test hooks | `core/database/promotealign/export_test.go` `LayerBridges`, `MemoSnapshot`, `MemoSnapshotForTest` | Call `loadLayer` and read `layerMemoKey`/`layerMemo` directly |
| Tests | `core/database/promotealign/memo_test.go` `TestMemoIsNotMutated`, `TestMemoFollowsCommit`, `TestProposeWarmIssuesNoQuery`; `promotealign_test.go` `TestLayerBridgesMergeKinshipShapes`; `perf_test.go` `TestLargeSourceBudgets`; `core/graphalign/align_test.go` helpers `bridge`, `implied`, `canonOf`; `invariants_test.go` | Build or hold `graphalign.Bridge` values; pin the memo's identity and invalidation |
| Docs | `.agents/skills/add-match-rule/SKILL.md` (kinship bullet: "the fields of `graphalign.Bridge`"); `core/graphalign/types.go` `Bridge` comment ("S13-29's shared loader") | Name the type this PR deletes |

**Precedent.** `a8dc05b4` "Move the link merge into core/kinship so every reader shares one." moved Promote's `links.go` (deleted) into `core/kinship/links.go` and pointed Promote at it, with the tests moving with the code. Copy its shape: move, re-point, delete the old file in the same commit. `core/database/canonlinks` is the precedent for the package's role: a `core/database` reader that "matching and later readers share", imported by both `promotealign` and `persondetail`. Do not copy `canonlinks`' long package doc; one paragraph is enough.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The package is `core/database/evidencelayer` *(new)*, one file `evidencelayer.go` for `Load`, `Layer`, `Subject`, `MemoKey`, the build and the links walk, and `values.go` for `Decode`, `LoadMetas`, `layerValues`, `observedTerms`. Package doc: "Package evidencelayer loads one Evidence graph's primary Subjects with their comparable values and its links, recorded and implied, memoized on the graph. Promote and the Person page share it." | "Evidence layer" is the term Promote's code and `promote-graph-alignment.md` already use for one Source's primary Subjects and bridges | `subjectvalues` — the package also owns links. Putting it in `canonlinks` — that package reads the canon, not an Evidence graph |
| D-2 | `type Subject struct { ID []byte; Ref string; Kind string; Values subjectcompare.Values }` — the same fields, names, types and order as `graphalign.Subject`, so `graphalign.Subject(s)` converts it. `type Layer struct { Subjects []Subject; Links []kinship.Link; Metas subjectcompare.Metas; byID map[string]int }`, Subjects sorted by Ref as today. `func (l Layer) Subject(id []byte) (Subject, bool)` reads `byID` | One value type both readers use; a direct struct conversion keeps Promote's per-call cost a slice copy (D-8). `byID` replaces Promote's `primary` map with no second map | Returning `graphalign.Layer` — puts graph alignment's types in the Person page's path (S13-26 D-2, working rule "Scoring and the view are two concerns"). `Kind vocabkey.SubjectType` — breaks the direct conversion; S13-27 converts once with `vocabkey.SubjectType(s.Kind)` |
| D-3 | `func Load(g *graphcache.Graph, graphID []byte, namePattern string, terms segmentcompare.Terms) (Layer, error)`. It reads the memo at `MemoKey + namePattern`, else builds and stores the `Layer` value itself (no wrapper type). The rules come from `matchregistry.Default().Subject` inside `Load`, not from a parameter. `nil` graph → `promote.ErrInvalid`, as today | One memo key for every reader is only safe if every reader builds with the same inputs. The key already carries the name pattern; terms come from the vocabulary, whose commit drops the memo; the rules are static, so reading them inside removes the one input a second caller could pass differently. Promote's `cfg.Subject` is `Default().Subject` today (`AlignConfig`), so its output does not change | A `subjectcompare.Rules` parameter — a caller with other rules would read a layer built under Promote's, under the same key. Loading `namePattern` and `terms` inside `Load` — Promote already loads both for the canon side, and `valuerules.Load` needs a `database.Querier` the graph alone does not give |
| D-4 | `const MemoKey = "evidencelayer\x00"` *(new)*, exported so `promotealign`'s `MemoSnapshotForTest` reads the memo as stored. The old `"promotealign.layer\x00"` key goes | Test hooks in another package need the key; the stored value is the exported `Layer` | An exported `Peek` function — production API used only by tests |
| D-5 | `Decode` (was `decode`) and `LoadMetas` (was `loadMetas`) are exported from `evidencelayer`; `promotealign`'s `reconciled` and `index.go` call them. `layerValues` and `observedTerms` stay unexported in `evidencelayer`. Doc comments move with the code; `Decode`'s says "the one row decoder for an Evidence Observation and a cached handle value" | One decoder for both sides stays one (today's comment). The canon side is Promote's only, so the shared package owns the decoder and Promote imports it | Leaving `decode` in `promotealign` and copying it into `evidencelayer` — two decoders for one row shape |
| D-6 | `lifeEvent` moves to `core/subjectprofile` as `func LifeEventOf(event subjectcompare.Values, role vocabkey.Role) LifeEvent` *(new name)*, body and comment unchanged. `evidencelayer` and `promotealign`'s canon profiler call it | It is pure and builds `subjectprofile.LifeEvent` from `subjectcompare.Values`: the layer that owns `LifeEvent` owns it. It has no numeric literal, so the `matchregistry` guard (`layerPackages`, `numbers: true`) still passes | Exporting it from `evidencelayer` — the canon profiler would read a pure helper through a database package |
| D-7 | `graphalign.Bridge` is deleted. `graphalign.Layer.Bridges` keeps its name and becomes `[]kinship.Link`. `CanonEdge`'s comment reads "Recording is as on kinship.Link." | The user's answer to S13-26 review M-5 (2026-11-07): one link type. Field access (`b.A`, `b.KeyA`, `b.Recording`, `b.Associations`) compiles unchanged, so only constructors move | Renaming the field to `Links` — dozens of test edits with no reader gain inside `graphalign`, which calls them bridges throughout |
| D-8 | `promotealign.loadInputs` calls `evidencelayer.Load(c.Graph(), graphID, namePattern, terms)` and builds `graphalign.Layer{Subjects: s, Bridges: el.Links, Metas: el.Metas}` with `s[i] = graphalign.Subject(el.Subjects[i])`, a fresh slice per call. `inputs` gains `evidence evidencelayer.Layer` and loses `primary` | Align, `loadCanon` and `loadCandidatesMemo` keep their `graphalign` signatures. The copy is one struct per primary Subject (100 on the large fixture) and shares every `Values` map, which no reader changes (`TestMemoIsNotMutated`) | Converting once and memoizing a `graphalign.Layer` under a second key — two memos for one graph, the thing the map row removes |
| D-9 | `mergeFixed(g, graphID, layer evidencelayer.Layer, caller)`, `checkFixed(g, layer, f)` and `pageEnds(promoted, layer, caller, rows, only)` take the Evidence layer. Existence and kind read `layer.Subject(id)`; the two `for sk := range primary` loops in `pageEnds` become `for _, s := range layer.Subjects { sk := string(s.ID) … }`; `mergeFixed`'s capacity uses `len(layer.Subjects)` | `primarySubject` was a second index of the same Subjects. `pageEnds` writes a map and `mergeFixed` sorts, so the loop order change is invisible | Keeping a `primary` map built per call — a second index of the layer |
| D-10 | Tests move with the code they test only where the code leaves Promote's package: `TestLayerBridgesMergeKinshipShapes` stays in `promotealign_test.go` with its fixture, and `LayerBridges` returns `[]kinship.Link` from `evidencelayer.Load`. New tests: T-1 in `core/subjectprofile`, T-2 in `core/database/evidencelayer`, T-3 in `core/database/promotealign/memo_test.go` (it has the warm fixture) | The fixtures that build Sources live in `promotealign`'s test package; copying them into a new package for an unchanged behavior is a second fixture | Moving the Source fixture into `databasetest` — larger than this PR and no other reader asks for it |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: "Evidence layer", "layer memo", `EvidenceGraphMemo` callers, `decode`/`UnmarshalDate` readers in `core/database`, `LifeEvent` builders, `kinship.Link` and `graphalign.Bridge` constructors.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/evidencelayer` *(new)* | package | `core/database/promotealign` (layer half); `core/database/canonlinks` | It is the same code moved; `canonlinks` reads the canon, not an Evidence graph |
| `evidencelayer.Load` *(new)* | func | `promotealign.loadLayerMemo` + `loadLayer` | Moved and merged: the memo read and the build are one entry point |
| `evidencelayer.Layer`, `Subject` *(new)*, `Layer.Subject` *(new)* | type / method | `graphalign.Layer`, `graphalign.Subject`, `promotealign.primarySubject` | `graphalign`'s types may not reach the Person page (D-2); `primarySubject` goes |
| `evidencelayer.MemoKey` *(new)* | const | `promotealign.layerMemoKey` | Moved, exported for the test hooks (D-4) |
| `evidencelayer.Decode`, `LoadMetas` *(new names)* | func | `promotealign.decode`, `loadMetas` | Moved and exported (D-5) |
| `subjectprofile.LifeEventOf` *(new name)* | func | `promotealign.lifeEvent` | Moved to the layer that owns `LifeEvent` (D-6) |
| `TestLifeEventOf`, `TestNoGraphAlignmentImport`, `TestLoadSharesPromotesMemo` *(new)* | test | `TestMemoIsNotMutated`; `core/matchregistry/guard_test.go` | The guard there covers pure layer packages only; this one covers a `core/database` package |

**Builds on.** `graphcache.Graph.EvidenceGraphMemo` / `SetEvidenceGraphMemo` (the memo drops with its graph, `use-catalog-session`); `kinship.LinkSet` (the one link merge, `a8dc05b4`); `subjectprofile.ForSubject`.
**Deletes.** `core/database/promotealign/layer.go` and `core/database/promotealign/terms.go` (whole files); from `core/database/promotealign/profiles.go`: `decode`, `layerValues`, `loadMetas`, `lifeEvent`; `primarySubject`, `layerMemo`, `layerMemoKey`; `inputs.primary`; `graphalign.Bridge`. `participation` and `neighborTypeProperty` move to `evidencelayer` (their only readers are the layer build). `hopPersonToEvent` moves from `layer.go` to `profiles.go`, beside its only readers (the canon profiler, `profiles.go` lines ~217 and ~295).
**Next adds.** S13-27 calls `evidencelayer.Load` per graph and passes `Links` to `recordlinks.NewGraph` and `Subjects` (with each member's entity) as `recordlinks.Member`s. A later profile memo for handles ([`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md)) reuses `Decode` and `LifeEventOf`.

### Reach

None — no new path into existing code, no new state reachable. `Load` has no caller outside `promotealign` until S13-27.

### Scale

At the map's Scale for Promote (a Source of 100 new people, the 41-Subject obituary): no write changes and no key is re-keyed beyond the memo key's prefix, which drops on the same events (`TestMemoFollowsCommit`). Per proposal, Promote adds one `graphalign.Subject` slice copy (100 structs, sharing `Values`) and loses the `primary` map read. `TestLargeSourceBudgets` measures it at that Scale (AC-2.2). Slice 7's Scale (50 Evidence graphs) is S13-27's budget subtest; this PR sets no budget for it.

### Edges

- **Assumptions.** (1) One memo for two readers is right only if both build with the same inputs: name pattern (in the key), terms (vocabulary, which drops the memo), rules (`Default()`, D-3). T-3 holds the sharing. (2) Readers never change a memoized `Layer`: Promote's per-call slice copy shares `Values` and `Links`; `TestMemoIsNotMutated` keeps holding it, now through `evidencelayer.MemoKey`. (3) `pageEnds` and `mergeFixed` are order-free (map output; sorted output), so ranging `Subjects` instead of a map changes nothing; `TestProposeWarmIssuesNoQuery`'s `reflect.DeepEqual` holds it.
- **Neighbors.** None — no rule changes; the same inputs reach the same code.
- **Removed inputs.** None — nothing is decided, excluded or deleted by this PR.
- **Real data.** None — no table-driven rule changes.
- **Untrusted size.** None — the walk is today's, bounded by `kinship.Rules.MaxPivotEnd` and `MaxPairSteps`, unchanged.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/evidencelayer` *(new)* | `Load`, `Layer`, `Subject`, `Layer.Subject`, `MemoKey` (D-2 – D-4); unexported `build` (was `loadLayer`), `links` (was `loadLayerBridges`, returns `[]kinship.Link` = `set.Links()` and `[]participation`), `layerValues`, `observedTerms`, `participation`, `neighborTypeProperty`; exported `Decode`, `LoadMetas` (D-5) |
| `core/subjectprofile` | `LifeEventOf` *(new name)* (D-6) |
| `core/graphalign` | `Bridge` deleted; `Layer.Bridges []kinship.Link` (D-7) |
| `core/database/promotealign` | `loadInputs` (D-8), `mergeFixed`, `checkFixed`, `pageEnds` (D-9); `reconciled`, `index.go`, the canon profiler call `evidencelayer.Decode`, `evidencelayer.LoadMetas`, `subjectprofile.LifeEventOf` |

```go
// Package evidencelayer …
package evidencelayer

const MemoKey = "evidencelayer\x00"

type Subject struct {
	ID     []byte
	Ref    string
	Kind   string // person, event, place
	Values subjectcompare.Values
}

type Layer struct {
	Subjects []Subject       // primary Subjects, by Ref
	Links    []kinship.Link  // recorded and implied, one per ends and keys
	Metas    subjectcompare.Metas
	byID     map[string]int
}

func (l Layer) Subject(id []byte) (Subject, bool)
func Load(g *graphcache.Graph, graphID []byte, namePattern string, terms segmentcompare.Terms) (Layer, error)
func Decode(scalar propertymeta.Scalar, termID, date, name, image []byte, terms map[string]graphcache.VocabTerm) (autoreconcile.Value, error)
func LoadMetas(g *graphcache.Graph) (subjectcompare.Metas, error)
```

Errors: `promote.ErrInvalid` for a nil graph (unchanged). `evidencelayer` imports `core/database/promote` for it, as `promotealign` does today.

### FFI

None — no wire change.

### macOS

None — Go only.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — `LifeEventOf` in subjectprofile**
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/lifeevent_test.go` *(new)*, `core/database/promotealign/profiles.go`, `core/database/promotealign/layer.go`. Skill: `add-match-rule` (Rules that hold everywhere).
Do: move `lifeEvent` to `subjectprofile.LifeEventOf` (D-6); point both callers at it; add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/subjectprofile/... ./core/matchregistry/... ./core/database/promotealign/...` passes.

**STEP-2 — The evidencelayer package**
Files: `core/database/evidencelayer/evidencelayer.go` *(new)*, `core/database/evidencelayer/values.go` *(new)*. Skill: `use-catalog-session`.
Do: copy `loadLayerMemo` + `loadLayer` into `Load`/`build`, `loadLayerBridges` into `links`, `observedTerms`, `layerValues`, `decode` → `Decode`, `loadMetas` → `LoadMetas` (D-1 – D-5). `build` reads `matchregistry.Default().Subject` for `Weights` and `Lifespans`; `links` returns `set.Links()` directly; `build` fills `byID`. Nothing calls it yet.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/...` and `go vet -tags fts5 ./core/database/evidencelayer/...` pass.

**STEP-3 — Promote reads the shared loader**
Files: `core/database/promotealign/promotealign.go`, `connections.go`, `profiles.go`, `index.go`, `export_test.go`; delete `layer.go` and `terms.go`.
Do: D-8 in `loadInputs`; D-9 in `mergeFixed`, `checkFixed`, `pageEnds`; `reconciled` and `index.go` call `evidencelayer.Decode` / `LoadMetas`. `export_test.go`: `LayerBridges` returns `([]kinship.Link, error)` from `evidencelayer.Load` with `loadConfig`'s name pattern and `canonlinks.LoadTerms`; `MemoSnapshot.Layer` becomes `evidencelayer.Layer`, read with `c.Graph().EvidenceGraphMemo(graphID, evidencelayer.MemoKey+namePattern)` and asserted to `evidencelayer.Layer`; `LayerKind` fills from `Layer.Subjects`. At this step `graphalign.Layer.Bridges` is still `[]graphalign.Bridge`: convert with `graphalign.Bridge{Edge: l.Edge, Recording: l.Recording}` in `loadInputs` until STEP-4 removes it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes with no assertion edited (the `TestLayerBridgesMergeKinshipShapes` helper types change in STEP-4).

**STEP-4 — One link type**
Files: `core/graphalign/types.go`, `core/graphalign/align_test.go`, `core/graphalign/invariants_test.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/promotealign_test.go`.
Do: delete `graphalign.Bridge`; `Layer.Bridges []kinship.Link`; fix the `CanonEdge` comment (D-7). Remove STEP-3's temporary conversion. Test helpers `bridge`, `implied`, `canonOf`, `side` and the `bridges` table fields return or hold `kinship.Link`; `TestLayerBridgesMergeKinshipShapes`'s `spouses`/`build` take `[]kinship.Link`. No assertion changes.
Check: `rg -n 'graphalign\.Bridge\b' core api` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/...` passes.

**STEP-5 — Budgets and verify mode**
Files: none changed.
Do: run the budget test and verify mode.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -v ./core/database/promotealign/` passes, and its `t.Logf` lines are pasted in the PR description beside the same run on `main`; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-6 — Guard and shared-memo tests**
Files: `core/database/evidencelayer/evidencelayer_test.go` *(new)*, `core/database/promotealign/memo_test.go`.
Do: add T-2 and T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestNoGraphAlignmentImport|TestLoadSharesPromotesMemo' ./core/database/...` passes.

**STEP-7 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md` (kinship bullet: "the fields of `kinship.Link`"; one clause that an Evidence graph's links are read through `core/database/evidencelayer`), `docs/promote-graph-alignment.md` §8 (one bullet: "**The Evidence layer memo** (`evidencelayer.Load`) holds one Evidence graph's primary Subjects, values and links; Promote and the Person page share it, and it drops with the Evidence graph memo"), `deployment-plan.md` (map row), `completed.md`, this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Read Promote's Evidence layer from one loader the Person page can share."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectprofile/lifeevent_test.go` `TestLifeEventOf` *(new)* | Table: an event with a positive `event_type` "birth" and a negative "death" → `EventType` birth; `date`, `start_date`, `end_date` positive → all three in `Dates` in that key order; a negative date → left out; no type → empty `EventType`; role passed through | D-6 |
| T-2 | Go | `core/database/evidencelayer/evidencelayer_test.go` `TestNoGraphAlignmentImport` *(new)* | `go/build` import of `core/database/evidencelayer` and its deps (walk `Imports` recursively within the module) → neither `core/graphalign` nor `core/database/promotealign` is reached | AC-1.2, D-2 |
| T-3 | Go | `core/database/promotealign/memo_test.go` `TestLoadSharesPromotesMemo` *(new)* | Warm fixture: Propose on Source B's graph; then, with `SetQueryCounterForTest` counting, `evidencelayer.Load(g, graph, namePattern, terms)` → 0 queries and `&got.Subjects[0] == &MemoSnapshotForTest(...).Layer.Subjects[0]` | AC-1.1, D-3, D-4 |
| T-4 | Go | `promotealign_test.go` `TestLayerBridgesMergeKinshipShapes` (types only) | Unchanged rows, now over `[]kinship.Link` | AC-1.3, D-7 |
| T-5 | Go | `memo_test.go` `TestMemoIsNotMutated`, `TestMemoFollowsCommit`, `TestProposeWarmIssuesNoQuery` (unchanged) | As today | AC-2.1, AC-2.3, D-8 |
| T-6 | Go | `perf_test.go` `TestLargeSourceBudgets` (unchanged) | As today | AC-2.2 |
| T-7 | Go | `core/graphalign` all tests (types only) | As today | AC-2.1, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (export hooks), `core/graphalign` (test helpers), `core/matchregistry` `guard_test.go` (`subjectprofile` gains a function).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/... passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`graphalign.Bridge`, `loadLayer`, `loadLayerMemo`, `loadLayerBridges`, `layerValues`, `promotealign.layer`, `lifeEvent`, `loadMetas`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-graph-alignment.md §8 (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check, TestLargeSourceBudgets timings before and after
```

## 10. Manual check

1. Open a project with a Source of several people (the 41-Subject obituary if available). Open Promote on it before and after this branch. See the same rows, the same suggested targets and the same connections.
2. Press Done on one row, reopen Promote. See the promoted row anchored as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The per-call `graphalign.Subject` copy costs the warm budget | T-6 at Scale; timings in the PR description |
| A test hook or test outside `promotealign` reads `layerMemoKey` | `rg layerMemoKey` before STEP-3; §4 lists the only reader |
| `subjectprofile` gains a database import by mistake | `guard_test.go` (`layerPackages`) fails |

**Stop and report if:**

- `loadLayer` reads any `graphalign.Config` field other than `Subject.Weights` and `Subject.Lifespans` (D-3 would change Promote's output).
- `loadConfig` sets `cfg.Subject` to anything but `matchregistry.Default().Subject`.
- A non-test file outside `core/database/promotealign/layer.go` constructs `graphalign.Bridge`.
- `TestLargeSourceBudgets` fails a budget on this branch that passes on `main` at the same run.
- Any `promotealign` or `graphalign` test needs an assertion change, not only a type change.
- `evidencelayer` cannot build without importing `core/graphalign`.

## 12. Open questions

None.

## 13. PR

**Title:** S13-29 — One loader for a Subject's comparable values
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1 – T-7, the verify run, the budget timings before and after); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-29-one-loader-for-comparable-values.md`.
