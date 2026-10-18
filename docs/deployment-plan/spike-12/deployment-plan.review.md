# Review — Spike 12 PR map

**Verdict:** Fits with changes
**Rounds:** 2. **Reviewers:** maintainability (map review), fresh agent each round

## To user
- None. Every finding was answered from the code or the skills; no product call was needed.

## Writer's own calls
- **Default graph id = its Source's id** in the migration (K-4), so the backfill needs no SQL randomness; refs come from `evidencegraphs.EnsureRefs`. A later graph gets a normal new id.
- **Last graph, last link and "Citations use this Artifact" are domain invariants** with their own `apperr` codes (K-12, option b). Only "graph has Subjects" goes through `deleteimpact`.
- **Package and symbol names:** `core/database/sourcecopy`, `core/database/evidencegraphs`, `SoleGraphTx`, `EvidenceGraph` / `evidenceGraph(graphId:)`, `graphprogress`.
- **Audit scope stays `source`**; no `graph` scope type.
- **Graphs cascade with their Source**; a Source can only be deleted with no Artifacts and no Subjects already.
- **S12-12 depends on S12-02** for the shared Artifact checklist.
- **K-14 option (a):** effects stay per Source through S12-07 and switch to graph ids in one PR (S12-08), instead of a reverse graph → Source bridge.

## Round 1

### K-1 · blocker · The graph list is folded into the Source workspace payload, but four other views also read it
`GetSourceWorkspaceResponse` reserves 6–8 for exactly this staleness; `add-workspace-place` forbids folding a shared list into a page payload.
**Answer:** Fixed — working rule "One graph list"; S12-05 adds one project-wide graph list RPC and key; S12-10, S12-11, S12-14 read it; `GetSourceWorkspace` is untouched.

### K-2 · blocker · Graph–Artifact links get written by four separate paths
`artifacts.CreateInput` and `CreateArtifactRequest` have no graph set.
**Answer:** Fixed — `evidencegraphs` `Link` / `Unlink` exist from S12-03 and every path calls them; `artifacts.CreateInput.GraphIDs` (nil = every graph) in S12-03, `CreateArtifactRequest.graph_ids` in S12-05; S12-13 is Swift only.

### K-3 · blocker · `sources.Copy` cannot live in `sources`
Verified: `core/database/sourcemetadata` imports `core/database/sources`.
**Answer:** Fixed — working rule "Copy lives in its own composing package, `core/database/sourcecopy`", with draft-read and create-from-draft steps; the second is the slot for `ideas/registry-source-import.md`.

### K-4 · major · The migration mints `GPH-` refs in SQL, against the ref skill
Verified: `add-catalog-ref` Don't column, "Mint in SQL migrations"; `users.EnsureRefs` is the precedent.
**Answer:** Fixed — option (a): working rule "Refs are minted in Go, never in SQL"; Default graph `id = source_id`, nullable `ref`, `evidencegraphs.EnsureRefs` at open. Feature doc data model updated.

### K-5 · major · The isolation check is built twice, and in parallel with the package that owns links
**Answer:** Fixed — one `requireSameGraph` in `observations` replacing `requireSubjectSource`; `connect` drops its duplicate checks; the membership reads live in `evidencegraphs` from S12-03.

### K-6 · major · S12-06's walk change is misplaced, so its S11-14 dependency is not needed
**Answer:** Fixed — S12-06 only changes the walk's join to `graph_id`; callers are re-keyed in S12-07 / S12-09; S11-14 removed from S12-06.

### K-7 · major · Missing or misplaced Spike 11 dependencies on Promote code
**Answer:** Fixed — S12-07 waits on S11-14 and S11-15 and reads `loadLayer`'s graph through `SoleGraphTx`; Promote's key and location moved from S12-08 to S12-09; S12-09 notes S11-21.

### K-8 · major · "The one bridge" has no single owner, but has at least five call sites
**Answer:** Fixed — working rule "One bridge, one function": `evidencegraphs.SoleGraphTx`, deleted in S12-08 except Promote's use, which S12-09 removes.

### K-9 · major · Graph ids are added beside the Source fields instead of replacing them, and two skills go stale
**Answer:** Fixed — churn rule now says replace, not add; working rule "Effects are graph-keyed"; the docs-and-skills rule lists each doc and skill per PR (`catalog-refs.md`, `add-catalog-ref`, effects README, `catalog-deletes.md`, `macos-client-patterns.md`, `add-workspace-place`, `add-catalog-delete`, `conclusion-layer-data-model.md`).

### K-10 · major · Source-named graph code keeps its old names after re-keying
**Answer:** Fixed — working rule "A re-keyed symbol is renamed in the PR that re-keys it", with `EvidenceGraph` as the one name.

### K-11 · major · The Copy form and the graph form each rebuild something the codebase already has, or will have twice
**Answer:** Fixed — working rule "One title and type form, one Artifact checklist"; S12-02 reuses `SourcesListView.addForm` / `SourcesModel.Draft` without the `setQueryValue` write-through; S12-12 reuses the checklist and depends on S12-02; briefs S12-D1 and S12-D3 name it.

### K-12 · major · Most of S12-04's refusals have no `deleteimpact` gate, and "the graph's Citations" has no home
**Answer:** Fixed — option (b): working rule "Refusals are of two kinds"; `evidencegraphs.CitationsUsingArtifact` is the one exported read, shared by the unlink refusal and isolation.

### K-13 · minor · S12-03's `NOT NULL` column breaks raw test inserts the row does not mention
**Answer:** Fixed — the S12-03 scope and the backfill working rule name the fixtures and move them onto the create functions.

## Round 2

Every round-1 answer holds. K-11 note: `SourcesListView.addForm` is private, so S12-02 makes it reusable — added to its scope.

### K-14 · blocker · From S12-07 to S12-08 the wire `sources` field would carry graph ids that Swift reads as Source ids
`writeEffectsProto` copies `set.Sources` unchanged; `CatalogQueryRegistry.invalidations(for:)` reads them as Source ids. Hidden on backfilled Sources by `id = source_id`, broken on new ones.
**Answer:** Fixed — option (a): working rule "Effects are graph-keyed, in one step"; S12-07 keeps `Set` per Source and drops every graph of a changed Source; S12-08 re-keys `Set`, `Effect`, the wire and Swift together.

### K-15 · blocker · No mechanism keeps the project-wide graph list fresh
**Answer:** Fixed — working rule "The graph list has its own effect": `Set.EvidenceGraphs` from S12-03, wire flag and one invalidation line in S12-05.

### K-16 · blocker · Two reads of the same graph rows
**Answer:** Fixed — S12-04 ships one `evidencegraphs.List` (full rows with linked Artifact ids, by name, optional Source filter); S12-05's RPC wraps it; no `ListBySource`.

Round 2 is the last round; no blocker is open.
