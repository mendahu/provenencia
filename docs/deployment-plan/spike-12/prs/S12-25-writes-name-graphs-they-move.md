# S12-25 — Writes name only the graphs they move

**Status:** Landed #621 (2026-10-29)
**Spike / slice:** Spike 12 · Slice 6 — Audit fixes: Medium
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: effects, wire and session cache only; no view draws differently)
**Board:** —
**Depends on:** S12-34 (landed [#614](https://github.com/mendahu/provenencia/pull/614)); S12-22 (landed [#600](https://github.com/mendahu/provenencia/pull/600), preferred merge order); S12-35 (landed [#617](https://github.com/mendahu/provenencia/pull/617), FakeStore scripting shape)
**Blocks:** S12-26
**Branch:** `s12-25-writes-name-graphs-they-move`
**Version bump:** none — internal effects and cache wiring; nothing a researcher sees as a product version ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md) (step 6 points at the README this PR edits), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/catalog-audit.mdc`
**Context:** Audit [A-7](../app-health-audit.md#a-7--a-citation-artifact-or-source-title-write-reloads-every-graph-of-its-source), [A-16](../app-health-audit.md#a-16--source-keyed-fields-on-graph-keyed-paths); working rules "A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key", "FakeStore is a seeded fake, not a second engine", "The cache-key descriptor comes first in slice 6"; map review [K-2](../deployment-plan.review.md), [K-13](../deployment-plan.review.md); `effects.Set.Search` (shape to copy); S12-34 `CatalogQueryRefresh`; bar outcome 12 first half
**Planned:** 2026-10-28 against `main` @ `504f4f9b`
**Review:** [`S12-25-writes-name-graphs-they-move.review.md`](S12-25-writes-name-graphs-they-move.review.md) — Fits with changes / Ready → Fixed, 2 rounds

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

A Citation save on a many-graph Source today names every Evidence graph of that Source, so the app drops every opened graph and refetches every graph's progress. This PR makes each write name only the graphs whose loaded rows it moves, and carries partial reads (citation counts, citations by Artifact, graph progress) as one keyed `reads` effect in the shape of `Set.Search`, so Swift refreshes only those graphs through S12-34's `refresh` slot.

**Why now.** Audit A-7 and A-16; working rule "A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key" (map review K-2). S12-34's descriptor `refresh` slot is landed, so the progress and citation updates are entries, not new session functions.

**What stays.** What each screen shows; graph-list reloads by Source (S12-26); hidden-key stale behavior (S12-26); the first-visit progress scan (audit L-12); Promote still returns no `WriteEffects` (S12-32).

## 2. User stories

**US-1.** As the engine, when a Citation, Artifact or Source-title write resolves, the Evidence graphs named for graphcache are only the graphs whose loaded rows move, so a Citation with Observations on one of sixty graphs names that one graph and an Artifact names none.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source has Default and Jones graphs; a Citation write whose Observations are only on Jones | `Set.EvidenceGraphs` is Jones only |
| AC-1.2 | The same Source; an Artifact create or label update (no Observation change) | `Set.EvidenceGraphs` is empty |
| AC-1.3 | The same Source; a Source title edit | `Set.EvidenceGraphs` is Default and Jones |
| AC-1.4 | A Citation with no Observations on an Artifact linked to Jones only | `Set.EvidenceGraphs` is empty; `Set.Reads` for `citations` names Jones |

**US-2.** As the engine, when a write moves citation reads or graph progress counts, it reports those as keyed `reads` (not `citation_sources`, and not via `evidence_graphs` for progress), so progress rides only Subjects, Observations and Identity Claims.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Citation create with one Observation on Jones | `WriteEffects.reads` has `citations` → Jones; no `citation_sources`; `evidence_graphs` is Jones (graphcache) and has no progress-only fan-out beyond that |
| AC-2.2 | A Subject create on Jones | `reads` has `evidence_graph_progress` → Jones; `evidence_graphs` is Jones |
| AC-2.3 | An Artifact label update | `reads` has no `evidence_graph_progress`; `evidence_graphs` empty |
| AC-2.4 | An Identity Claim create on a Jones Subject | `reads` has `evidence_graph_progress` → Jones |

**US-3.** As a researcher, when I save a Citation on one graph of a multi-graph Source, only that graph's citation reads and (when counts change) progress refresh; other graphs' caches stay.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Sources list warmed with progress for g-1 and g-2; apply effects with `reads` progress for g-2 only | Store `getEvidenceGraphProgress` is called for g-2 only; g-1 map entry unchanged |
| AC-3.2 | Graph surface warmed with `.citationCounts` for g-1; apply `reads` citations for g-1 | That counts key reloads; a warmed counts key for g-2 does not |
| AC-3.3 | Composer shows Artifact a1 with its citations key in `visibleKeys`; apply a Citation write's effects | `applyWrite` calls `session.apply` only (no `session.query` after); the citations key revalidates because it is visible |

**US-4.** As a contributor, `EvidenceGraphProgress` no longer carries `source_id`, and `graphprogress.ErrInvalid` uses an `evidencegraphs` code.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `GetEvidenceGraphProgress` / `ListEvidenceGraphProgress` proto decode | No `source_id` field on the wire (reserved); Swift `EvidenceGraphProgress` has no `sourceId` |
| AC-4.2 | `graphprogress.Get` with a malformed graph id | Error code `evidencegraphs.invalid` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `Effect.Reads` / `Set.Reads` / resolve / `WriteEffects.reads`; reserve `citation_sources` (D-1, D-2) | STEP-1, STEP-2, AC-2.* |
| Registry: Citation and Artifact `EvidenceGraph` narrowed; `CitationSource` removed; Reads on citation, observation, link, value, subject and claim tables (D-3, D-4) | STEP-1, AC-1.*, AC-2.* |
| `prefer` path + citation observation / artifact-link helpers (D-3) | STEP-1 |
| `EvidenceGraphProgress.source_id` reserved; Go/Swift drop; `graphprogress.ErrInvalid` → `evidencegraphs.invalid` (D-5) | STEP-2, STEP-3, AC-4.* |
| Swift `CatalogWriteEffects.reads`, `CatalogQueryFollows.reads`, progress + citation `refresh` entries; drop `citationSources` invalidation (D-6, D-7) | STEP-3, AC-3.1, AC-3.2 |
| Composer Artifact key visible via PlaceRegistry; `applyWrite` is `apply` only; identityKey omits artifact (D-8) | STEP-4, AC-3.3 |
| FakeStore scripts `reads`; progress seeds drop `sourceId` (D-9) | STEP-3, STEP-5 |
| Tests for ACs; docs/skills (effects README, add-audit-scope, add-workspace-place, macos-client-patterns rule 2) (D-10) | STEP-5, STEP-6 |

| Out | Home |
| --- | --- |
| Graph list by Source via a `reads` entry; drop catalog-wide `evidence_graphs` list key | S12-26 |
| Hidden `.evidenceGraph` / `.sourceWorkspace` → `.markStale`; `QueryHandle.isStale` | S12-26 |
| First-visit progress scan | audit L-12 / not planned |
| Promote returning `WriteEffects` (progress still unions `mutation?.evidenceGraphId` on the progress follows) | S12-32 |
| Moving `invalidations(for effects:)` onto descriptor rows | S13-01 |
| Sidebar `counts` list effect | S12-32 |

## 4. Current state

Verified at `504f4f9b`.

| Area | Where | Today |
| --- | --- | --- |
| Effect / Set | `core/database/effects/registry.go` `Effect`, `Set` | `CitationSource Path` / `CitationSources [][]byte`; `Search []SearchDoc` / `Search []SearchIDs`; no `Reads`. Comments tie progress to `EvidenceGraphs` |
| Registry rows | same file `"artifacts"`, `"citations"`, `"observations"`, `"evidence_graph_artifacts"`, `ownedValue` | Artifacts/citations: `EvidenceGraph: graphsOfSource(...)`, `CitationSource` to Source; observations and links set `CitationSource`; owned values set `CitationSource: valueSource(col)` |
| Path helpers | `registry.go` `observationGraph`, `graphsOfSource`, `valueGraph`, `valueSource`; `path.go` `union`, `sqlPath` | No `prefer`; no citation-observation path; `sqlPath.check` returns `""` |
| Resolve | `core/database/effects/resolve.go` | Appends `CitationSources` and `Search`; no Reads walk |
| Completeness | `registry_completeness_test.go` `TestSourceEvidenceIsNotActivity`, `TestEvidenceGraphPathsEndAtGraphs` | Pins artifacts/citations evidence+citation true; EvidenceGraph paths must end at `evidence_graphs` |
| Resolve tests | `evidence_graphs_test.go` `TestEvidenceGraphsResolve` | Citation and artifact create expect Default+Jones; comment says Artifact/Citation name every graph of the Source |
| Wire | `api/proto/engine.proto` `WriteEffects` | `citation_sources = 11`; `evidence_graphs = 14` comment includes progress; no `reads`. `EvidenceGraphProgress.source_id = 1` |
| FFI map | `api/ffi/handlers/write_effects.go` `writeEffectsProto` | Maps `CitationSources`; no Reads |
| graphprogress | `core/database/graphprogress/graphprogress.go` `Progress`, `Get`, `ErrInvalid` | `SourceID` filled via `evidencegraphs.SourceOf`; `ErrInvalid` = `CodeSourcesInvalid` |
| Swift effects | `CatalogMutation.swift` `CatalogWriteEffects` | `citationSources: [String]`; no `reads` |
| GoStore map | `GoStore.swift` `mapWriteEffects`, `mapEvidenceGraphProgress` | Maps `citationSources` and `sourceID` |
| Follows / refresh | `CatalogLastUpdated.swift` `CatalogQueryFollows`, `CatalogQueryRefresh` | Cases `.handle`, `.source`, `.evidenceGraphs`; progress refresh follows `.evidenceGraphs` |
| Invalidation | `CatalogQueryRegistry.swift` `invalidations(for:project:)` | Non-empty `citationSources` → `.allCached(.citationCounts)` + `.allCached(.citationsByArtifact)` |
| Descriptors | same file `.citationCounts`, `.citationsByArtifact`, `.evidenceGraphProgress` | Citation kinds: no `refresh`, comments name `citationSources`; progress: `refresh` with `.evidenceGraphs` |
| Composer | `CitationComposerContext.swift` `applyWrite`; `CitationComposerEntry.identityKey`; `PlaceRegistry` `.sourceCitationComposer` | Post-`apply` `session.query` when `citationSources` non-empty (D-28); identityKey includes `artifactID`; place adds `.citationsByArtifact` only when `location.artifactId` set |
| FakeStore | `FakeStore.swift` `nextWriteEffects`, progress seeds | Scripts effects; no `reads`; progress includes `sourceId` |
| Docs / skills | `core/database/effects/README.md`; `.agents/skills/add-audit-scope/SKILL.md`; `.agents/skills/add-workspace-place/SKILL.md` (cache refresh + "citationSources"); `docs/macos-client-patterns.md` rule 2 | Say Citation/Artifact name every Source graph; citation reads via `citationSources`; progress tied to `EvidenceGraphs` |

**Precedent.** [#528](https://github.com/mendahu/provenencia/pull/528) / S12-08 reserved `WriteEffects.sources` and added `evidence_graphs = 14` in one cut — copy that reserve-and-add for `citation_sources` → `reads`. `Set.Search` / `SearchDoc` / `SearchIDs` in `registry.go` is the resolve shape to copy for Reads (Search stays server-only; Reads go on the wire). S12-34 D-6's `CatalogQueryRefresh` + `CatalogQueryFollows` is the Swift slot citation and progress entries extend. S12-35's `nextWriteEffects` scripting is how FakeStore reports the new field (no engine composition).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Add `type Read string` with `ReadCitations Read = "citations"` (citation counts and citations-by-artifact) and `ReadEvidenceGraphProgress Read = "evidence_graph_progress"`, plus `var Reads = []Read{ReadCitations, ReadEvidenceGraphProgress}` beside `List` / `Lists`. `type ReadPath struct { Read Read; Path Path }` and `type ReadIDs struct { Read Read; IDs [][]byte }` beside `SearchDoc` / `SearchIDs`. `Effect.Reads []ReadPath`; `Set.Reads []ReadIDs`. `Resolve` walks `eff.Reads` like Search (dedupe by `Read`, append ids). `validate` requires each `Read` ∈ `Reads` (same as Lists) and a non-zero Path; `paths()` includes each `ReadPath.Path`. Delete `Effect.CitationSource` and `Set.CitationSources`. Swift: `enum CatalogRead: String` *(new beside `CatalogList` in `CatalogMutation.swift`)* with the same raw values; `GoStore.mapWriteEffects` `compactMap`s unknown reads away; `api/ffi/swift_vocabulary_test.go` gains a "catalog reads" row over `effects.Reads` like "catalog-wide lists". `CatalogQueryFollows.reads(CatalogRead)` | Working rule: one keyed effect in the shape of `Set.Search`; closed wire vocabulary like `Lists` / `CatalogList` (macos-client.mdc); S12-26 adds one `Reads` entry and one enum case | Open `string` read names (mistype refreshes nothing); separate read names per citation key; keeping `CitationSource` beside `Reads` |
| D-2 | Proto `WriteEffects`: `reserved 11;` and `reserved "citation_sources";`; add `message WriteEffectsRead { string read = 1; repeated string ids = 2; }` and `repeated WriteEffectsRead reads = 15;`. Comment on `evidence_graphs = 14` becomes "Evidence graphs whose loaded rows this write changed (graphcache drops each)" — no progress. `writeEffectsProto` maps `set.Reads`; drops `CitationSources`. Regenerate with `./scripts/generate-proto.sh` | Same reserve-and-add as S12-08; field 15 is free | Reusing field 11 for a different shape (old clients would mis-decode) |
| D-3 | Path helpers beside `observationGraph`: `citationObservationGraphs = chain(inbound("observations", "citation_id"), observationGraph)`; `prefer(primary, fallback Path) Path` *(new in `path.go`)* — runs `primary`, and if its ids are empty runs `fallback`; `check` requires both arms end at the same table. `citationArtifactGraphs = chain(up("artifact_id"), across("evidence_graph_artifacts", "artifact_id", "graph_id"))` *(reuses `across` as `membersOf` does; both FKs are already in `catalogmodel`)*. Registry: **citations** `EvidenceGraph: citationObservationGraphs`, `Reads: [{ReadCitations, prefer(citationObservationGraphs, citationArtifactGraphs)}]`, no `CitationSource`; **artifacts** `EvidenceGraph` empty, no citation Reads (child citation/link/observation changes carry their own), keep `SourceScope`/`Search`; **observations** keep `EvidenceGraph: observationGraph`, `Reads: [{ReadCitations, observationGraph}, {ReadEvidenceGraphProgress, observationGraph}]`; **evidence_graph_artifacts** `Reads: [{ReadCitations, field("graph_id")}]`, no `CitationSource`; **subjects** `Reads: [{ReadEvidenceGraphProgress, field("graph_id")}]` plus existing `EvidenceGraph`; **identity_claims** / **identity_claim_evidence** add progress `Reads` on their existing graph paths; **ownedValue** replace `CitationSource` with `Reads: [{ReadCitations, valueGraph(col)}]` (graph-keyed; `SourceScope` stays `valueSource`) | A-7 / working rule; citation reads need the no-Observation fallback (map Ships); Artifact field writes do not move loaded graph rows or citation membership (links and cascaded citation rows do); FK walk stays in `catalogmodel` | `union` of observation and artifact graphs for every Citation (re-widens a Citation with Observations); `sqlPath` for the link table (`check` returns `""`, drifts from `catalogmodel`); Artifact `EvidenceGraph: graphsOfSource` kept; `CitationSource` left on owned values pointing at Sources |
| D-4 | `Effect.EvidenceGraph` / `Set.EvidenceGraphs` mean graphcache drop only. Progress leaves that path: only `ReadEvidenceGraphProgress` on Subjects, Observations and Identity Claims (and claim evidence) carries progress ids. Source title keeps `EvidenceGraph: onField("title", graphsOfSource(self()))` with no progress read | Working rule: progress only on count-changing writes; title holds the Source name in the graph's rows/search | Progress still following every `evidence_graphs` id (A-7) |
| D-5 | `EvidenceGraphProgress`: `reserved 1;` / `reserved "source_id";` keep fields 2,3,6,7. Go `Progress` drops `SourceID`; `Get` drops the `SourceOf` call; `List` SQL groups by `graph_id` only (drop `source_id` from select/group). `ErrInvalid = apperr.New(apperr.CodeEvidenceGraphsInvalid, apperr.KindUser)` (same code string as `evidencegraphs.ErrInvalid`). Swift `EvidenceGraphProgress` drops `sourceId`; `zeros` / FakeStore / tests stop passing it; `mapEvidenceGraphProgress` stops reading it | A-16; Swift UI never read `sourceId` | Keeping `source_id` on the wire "for Free"; a new `graphprogress.*` code beside `evidencegraphs.invalid` |
| D-6 | Swift: `struct CatalogWriteEffectsRead: Sendable, Equatable { var read: CatalogRead; var ids: [String] }`; `CatalogWriteEffects.reads: [CatalogWriteEffectsRead]`; delete `citationSources`. `CatalogQueryFollows.reads(CatalogRead)` *(new case)*: ids = the matching `effects.reads` entry's ids; for `.evidenceGraphProgress` only, also union `mutation?.evidenceGraphId` when absent (Promote until S12-32). Delete `CatalogQueryFollows.evidenceGraphs` (its only caller was progress refresh). `.evidenceGraphProgress.refresh.follows` becomes `.reads(.evidenceGraphProgress)`; its `run` keeps today's Get-one + map merge over `ids`. Extract the visible/hidden switch inside `WorkspaceSession.apply(invalidations:)` into `applyHiddenPolicy(to key: CatalogQueryKey)` *(new)* — same `.evict` / `.markStale` (insert `invalidatedKeys` only) / `.revalidate` (`invalidate` then `revalidate`) behavior; `apply(invalidations:)` calls it per cached key. Citation kinds: each gains `refresh: CatalogQueryRefresh(follows: .reads(.citations))` whose `run` returns unless the key's `graphId` is in `ids`, then calls `applyHiddenPolicy(to: key)` (descriptor `load` runs only on revalidate). Delete the `citationSources` branch in `invalidations(for:)` | Map: refresh slot; per-graph; Promote still patches progress via mutation; one hidden-policy function for invalidations and citation refresh; store call stays on the descriptor `load` | A second `refreshCitationRead` that re-implements markStale (S12-34 rejected session-hardcoded helpers; S12-26's `isStale` would miss it); duplicate store calls + `setQueryValue`; keeping `.evidenceGraphs` follows with no caller |
| D-7 | `invalidations(for:)` still emits one `.key(.evidenceGraph(...))` per `effects.evidenceGraphs` id. Citation kinds keep `invalidateOn: []` and `whenHidden: .markStale`. Citation `refresh` uses D-6's `applyHiddenPolicy` so a hidden warmed key is marked stale the same way as any other invalidation | One hidden-policy path; Scale stays one graph's keys | Eager refetch of every hidden citations-by-artifact key of that graph |
| D-8 | Composer visible key: remove `artifactID` from `CitationComposerEntry.identityKey` (graph/subject/citation/connect ids stay). Add `WorkspaceNavigation.replaceCurrent(with:)` *(new)* that applies `history.replaceCurrent` (or `apply` when no history) so the current entry updates without pushing. On `CitationComposerModel.selectedArtifactID`'s setter (not a rebuilt location): copy `navigation.currentLocation`, set `artifactId` to the new value (or clear it), and `replaceCurrent` that copy so every other deep id (`observationId`, `citationId`, and the rest) and display fields (`title`, `ref`, `sourceTitle`) stay. Init may still write `context.artifactID` directly (navigation is not wired yet; the arrival location already carries the Artifact). `CitationComposerContext.applyWrite` becomes `session.apply(effects)` only (delete the post-apply `query`). Do not add `composerLocation` | Audit A-16 "through PlaceRegistry"; rule 2 forbids `query` after `apply`; remount was why the location could not name the Artifact; one location value cannot drop `observationId` or wipe crumb titles | Rebuilding `.citationComposer` from a second field list; session API for "extra visible keys"; pushing `go(to:)` on every Artifact switch; leaving the D-28 query |
| D-9 | FakeStore: no composition of `reads`; tests set `nextWriteEffects` with the `reads` and `evidenceGraphs` the case needs. Progress fixtures drop `sourceId`. No new FakeStore engine rule | Working rule / S12-35 | Synthesizing citation reads from graph membership in the fake |
| D-10 | Docs in this PR: rewrite `core/database/effects/README.md` (Artifact/Citation graphs, Reads, progress split); `.agents/skills/add-audit-scope/SKILL.md` lines that name `CitationSource` / progress-on-EvidenceGraph, and the stale "Set Source too" line (rewrite onto `EvidenceGraph` / `Reads`); `.agents/skills/add-workspace-place/SKILL.md` refresh row and the "citationSources" sentence; `docs/macos-client-patterns.md` rule 2 (`reads` instead of `citationSources`; progress via reads). `add-catalog-model` step 6 stays pointed at the README (no field list). Map row → Planned then Landed; `docs/deployment-plan/spike-12/completed.md` paragraph on land | Map review K-13 | Leaving skills teaching `citation_sources` or `Effect.Source` |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Read`, `Reads`, `ReadPath`, `ReadIDs`, `Effect.Reads`, `Set.Reads` *(new)* | type / field / const | `List` / `Lists`; `SearchDoc` / `SearchIDs` | Lists are catalog-wide flags; Search is server-only; Reads ride the wire with ids |
| `prefer` *(new)* | func | `union` | `union` always merges both arms; citation EvidenceGraph must stay observation-only |
| `citationObservationGraphs`, `citationArtifactGraphs` *(new)* | var | `observationGraph`; `across` / `membersOf` | Citation-specific ends; artifact graphs reuse `across`, not `sqlPath` |
| `WriteEffectsRead`, `reads = 15` *(new)* | proto / FFI | `lists` + Search shape | Need read name + ids per entry |
| `CatalogRead`, `CatalogWriteEffectsRead`, `CatalogQueryFollows.reads` *(new)* | Swift type / case | `CatalogList`; progress refresh follows | Closed vocabulary; ids from `reads` |
| `WorkspaceSession.applyHiddenPolicy(to:)` *(new)* | helper | switch inside `apply(invalidations:)` | Extracted once; citation refresh and invalidations share it |
| Citation descriptor `refresh` entries *(new)* | slot | `.evidenceGraphProgress` refresh | Same slot; citation bodies call `applyHiddenPolicy` when the key's graph is in `ids` |
| `WorkspaceNavigation.replaceCurrent(with:)` *(new)* | func | `fallbackToSectionRoot` / `NavigationHistoryStore.replaceCurrent` | Public same-place update for visibleKeys |

**Builds on.** `Set.Search` resolve loop; `List`/`Lists` closed set; `across` FK walk; S12-08 reserve-and-add; S12-34 descriptor `refresh` / `follows`; S12-35 `nextWriteEffects`; PlaceRegistry conditional `.citationsByArtifact`.
**Deletes.** `Effect.CitationSource`, `Set.CitationSources`, `WriteEffects.citation_sources`, `CatalogWriteEffects.citationSources`; `CatalogQueryFollows.evidenceGraphs`; `Progress.SourceID` / proto `source_id` / Swift `sourceId`; `CitationComposerContext.applyWrite`'s post-`apply` `query`; `identityKey`'s artifact segment; docs/skill lines that teach Source-wide citation effects or progress-on-every-`evidence_graphs`.
**Next adds.** S12-26: one more `effects.Read` constant + `CatalogRead` case + graph-list `refresh`. S12-32: Promote returns effects so progress follows can drop the mutation union.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Composer location gains/changes `artifactId` via `replaceCurrent` without remount | `CitationComposerEntry` / `PlaceRegistry.sourceCitationComposer` | identityKey no longer includes artifact (no remount); place adds/drops `.citationsByArtifact` from `visibleKeys`; empty artifactId omits the key (same as today) |

### Scale

At the map's Scale (a census with 60 graphs and 300 Citations on one Source): a Citation save whose Observations sit on one graph resolves one Evidence graph id and one `citations` read id (plus progress when the batch includes Observations). Swift revalidates at most that graph's visible citation keys and one progress Get — not ×60; hidden citation keys of that graph go stale through the same `applyHiddenPolicy` path as other invalidations. An Artifact label edit names no graphs and no progress read. A Source title edit still names all 60 for graphcache (title in each graph's rows) and does not refresh progress.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | D-1, D-3, D-4: types, `prefer`, registry, resolve, README, completeness and evidence_graphs tests |
| `core/database/graphprogress` | D-5: drop `SourceID`; `ErrInvalid` code; SQL without `source_id` |

Errors: `evidencegraphs.invalid` → malformed graph id in `graphprogress.Get` (AC-4.2).

### FFI

`api/proto/engine.proto`: D-2 and D-5. Handler `write_effects.go` maps `Reads`. `evidence_graph_progress.go` `protoProgress` omits `SourceId`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Write effects decode | — | `GoStore.mapWriteEffects` | `reads` |
| Follows / refresh | — | `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift` | D-6 |
| Composer apply / location | — | `CitationComposerContext`, `CitationComposerEntry`, `CitationComposerModel` / View, `WorkspaceNavigation` | D-8 |
| Progress type | — | `GenealogyStore.EvidenceGraphProgress` | drop `sourceId` |

States: unchanged. Accessibility: unchanged.

### Strings

None — no user-facing copy. Sweep still `rg`s deleted symbol names (`citation_sources`, `CitationSource`, `citationSources`, `sourceId` on progress) across skills, rules, docs and `L10n.swift`.

## 7. Steps

**STEP-1 — Effects: Reads, narrowed graphs, drop CitationSource**
Files: `core/database/effects/registry.go`, `path.go`, `resolve.go`, `registry_completeness_test.go`, `evidence_graphs_test.go`, `reads_test.go` *(new)*. Skill: none beyond go-tests.
Do: D-1, D-3, D-4. Replace the `citation` column of `TestSourceEvidenceIsNotActivity` with a `TestReadsDeclare` in `reads_test.go` that pins which tables list `ReadCitations` and which list `ReadEvidenceGraphProgress` (same shape as `TestEvidenceGraphsList`). Update `TestEvidenceGraphsResolve` expectations for citation (Jones only when Observations on Jones; empty EvidenceGraphs for citation-only with no Observations) and artifact create (empty EvidenceGraphs). Pin AC-1.4 in `TestCitationReadsPreferArtifactGraphs` in `reads_test.go`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes.

**STEP-2 — Wire + graphprogress**
Files: `api/proto/engine.proto`, `./scripts/generate-proto.sh`, `api/ffi/handlers/write_effects.go`, `write_effects_test.go`, `evidence_graph_progress.go`, `core/database/graphprogress/*`. Skill: `add-ffi-handler`.
Do: D-2, D-5. Pin write_effects tests for `reads` and reserved `citation_sources`. Pin graphprogress invalid code.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ ./core/database/graphprogress/...` passes.

**STEP-3 — Swift effects, follows, refresh**
Files: `CatalogMutation.swift`, `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift` (`applyHiddenPolicy(to:)`), `GoStore.swift`, `GenealogyStore.swift`, `FakeStore.swift` (progress type + map only), `api/ffi/swift_vocabulary_test.go`. Skill: `add-workspace-place`.
Do: D-6, D-7, D-9 (types). Add `CatalogRead` and the vocabulary-test row. Delete `CatalogQueryFollows.evidenceGraphs`. Extract `applyHiddenPolicy`; wire citation refreshes. Update `CatalogQueryRegistryTests` / `WorkspaceSessionTests` that assert `citationSources` → allCached and progress-via-`evidenceGraphs`.
Check: `rg citationSources macos/App` and `rg "case evidenceGraphs" macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` have no leftover App hits for the deleted field/case (tests may still be updating in STEP-5).

**STEP-4 — Composer visible Artifact key**
Files: `CitationComposerEntry.swift`, `CitationComposerContext.swift`, `CitationComposerModel.swift`, `WorkspaceNavigation.swift`, related tests. Skill: `add-workspace-place`, `add-swift-test`.
Do: D-8. `selectedArtifactID` setter copies `currentLocation`, sets `artifactId`, `replaceCurrent`. Rewrite `composerWriteReloadsTheArtifactListItShows` so the location includes `artifactId`, the key is visible, and `applyWrite` does not call `query`.
Check: that test and PlaceRegistry/composer tests pass under `scripts/check-pr.sh`'s Xcode slice when run later.

**STEP-5 — Tests**
Files: effects tests (done in STEP-1), `WorkspaceSessionTests`, `CatalogQueryRegistryTests`, `CitationComposerModelTests`, progress-related seeds in FakeStore tests, `EvidenceGraphStoreTests` if they pass `sourceId`. Skill: `add-swift-test`, go-tests.
Do: T-1 – T-8. Script `nextWriteEffects` with `reads` where tests applied `citationSources`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/database/graphprogress/... ./api/ffi/handlers/`; Xcode test classes named in §8.

**STEP-6 — Docs**
Files: `core/database/effects/README.md`, `.agents/skills/add-audit-scope/SKILL.md`, `.agents/skills/add-workspace-place/SKILL.md` (and `.claude/skills` mirrors if the repo keeps them in sync via check), `docs/macos-client-patterns.md` rule 2, this plan status, `docs/deployment-plan/spike-12/deployment-plan.md` map row, `docs/deployment-plan/spike-12/completed.md` on land. Skill: none.
Do: D-10.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-25-writes-name-graphs-they-move.md`; `rg` sweep in DoD; `scripts/check-guidance-names.py` via check-pr.

Commits: one per behavior. Subjects like "Name only the Evidence graphs a Citation's Observations move." and "Refresh citation reads and graph progress from WriteEffects.reads."

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `evidence_graphs_test.go` `TestEvidenceGraphsResolve` (update) | Citation with Observations on Jones → EvidenceGraphs Jones only; Artifact create → empty EvidenceGraphs; title edit → both graphs | AC-1.1, AC-1.2, AC-1.3, D-3 |
| T-2 | Go | `reads_test.go` *(new)* `TestCitationReadsPreferArtifactGraphs` | Citation with no Observations, Artifact linked to Jones → Reads `citations` = Jones; EvidenceGraphs empty | AC-1.4, D-3 |
| T-3 | Go | `reads_test.go` *(new)* `TestReadsDeclare`; completeness update | artifacts: EvidenceGraph unset; tables with `ReadCitations` / `ReadEvidenceGraphProgress` match D-3 | AC-2.2, AC-2.3, AC-2.4, D-1, D-4 |
| T-4 | Go | `write_effects_test.go` (update) | Result with Reads → proto `reads`; no `citation_sources` payload | AC-2.1, D-2 |
| T-5 | Go | `graphprogress` test *(new or update)*; `api/ffi/handlers` progress proto test *(update)* | malformed id → `evidencegraphs.invalid`; Get/List and `protoProgress` omit `source_id` / `SourceID` | AC-4.1, AC-4.2, D-5 |
| T-6 | Swift | `WorkspaceSessionTests` (update progress test) | `reads` progress for g-2 only → one Get; g-1 unchanged | AC-3.1, D-6 |
| T-7 | Swift | `WorkspaceSessionTests` / `CatalogQueryRegistryTests` | `reads` citations for g-1 reloads g-1 counts only | AC-3.2, D-6 |
| T-8 | Swift | `WorkspaceSessionTests.composerWriteReloadsTheArtifactListItShows` (rewrite) | location with `artifactId` → key visible; `applyWrite` → fetching without post-`query` | AC-3.3, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `WorkspaceSessionTests`, `CatalogQueryRegistryTests`, `CitationComposerModelTests`, `SourcesModelTests` (progress seeds), `core/database/effects`, `api/ffi/handlers` write_effects.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`CitationSource`, `citation_sources`, `citationSources` as the wire/effect field, `EvidenceGraphProgress.source_id` / Swift `sourceId` on that type); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: `docs/macos-client-patterns.md` rule 2; `core/database/effects/README.md`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has at least two Evidence graphs and Citations on one of them (the dogfood census split, or any multi-graph Source).
2. Open the Sources list so progress is warm. Save a Citation edit on graph A that does not add/remove Observations. Confirm graph B's progress counts on the list do not flicker/reload; graph A's citation UI updates.
3. Add a Subject on graph A. Confirm only graph A's progress counts change on the list.
4. Edit an Artifact label. Confirm no graph progress refetch for sibling graphs.
5. In the Citation composer, switch Artifacts and save a Citation: the list for that Artifact updates with no remount flash of the whole composer chrome.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `prefer` / sqlPath fails `check` when hung off EvidenceGraph | Citation EvidenceGraph uses only `citationObservationGraphs`; sqlPath only on Reads |
| Composer `replaceCurrent` fights leave-guard / dirty form | Call it only from `CitationComposerModel` Artifact selection paths that already clear or hold navigation today; stop if leave-guard must see a push |
| Promote progress regresses | D-6 unions `mutation?.evidenceGraphId` into progress follows |
| Hidden citation keys | D-6 / D-7 mark them stale via `applyHiddenPolicy` (same path as other invalidations); visible keys revalidate |

**Stop and report if:**

- `Set.Search` resolve loop no longer matches what §4 describes (cannot copy for Reads).
- `CatalogQueryRefresh` / `CatalogQueryFollows` are gone or renamed so D-6 cannot attach (S12-34 drift).
- Making `identityKey` omit artifact remounts the composer on Artifact switch for another reason (find the other identity source).
- `across("evidence_graph_artifacts", "artifact_id", "graph_id")` fails `check` against `catalogmodel` (both FKs are present at the planning base — stop if that changed).
- `scripts/check-pr.sh` cannot pass without implementing an Out line (S12-26 graph list, and so on).

## 12. Open questions

None.

### Landed

- `TestCitationReadsPreferArtifactGraphs` is in `reads_resolve_test.go` (package `effects_test`), not `reads_test.go`, because `TestReadsDeclare` needs the unexported registry and one file cannot be both packages.
- `SourcePageModelTests` and `SourcesModelTests` progress tests apply a `reads` progress entry; they relied on progress following `evidenceGraphs`.
- `WorkspaceSession.apply(invalidations:)` dedupes keys before `applyHiddenPolicy(to:)`, so a key named twice still reloads once.
- A link-row delete (graph delete) now names the deleted graph in the `citations` read; the `citation_sources` Source-less case it replaces named none.

## 13. PR

**Title:** S12-25 — Writes name only the graphs they move
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-25-writes-name-graphs-they-move.md`.
