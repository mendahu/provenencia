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

---

# Review — Spike 12 slices 5–7 (audit fixes, S12-20 – S12-38)

**Verdict:** Fits with changes
**Rounds:** 2 (round 2 because round 1 had blockers and K-2 reshapes S12-25 and S12-26). **Reviewers:** maintainability (map review), fresh agent.

## To user
- **K-1** answered (user, 2026-10-27): S12-34 moves to the head of slice 6. Working rule "The cache-key descriptor comes first in slice 6".
- **K-9** answered (user, 2026-10-27): option (a), a seeded fake. Working rule "FakeStore is a seeded fake, not a second engine"; S12-35 is "FakeStore is a seeded fake" and moves to the head of slice 6.

## Also decided with the user (2026-10-27)
- **S12-39** added: oversized JPEGs get a thumbnail from a pure-Go reduced decoder of our own (DC coefficients only), not macOS ImageIO, so a Windows build reuses it. Design: [`oversized-jpeg-thumbnails.md`](oversized-jpeg-thumbnails.md), moved from `ideas/large-jpeg-thumbnails.md`.
- Four Spike 10 likeness-audit leftovers go to Spike 14 instead of Spike 12: the gallery image source (S14-01) and the composer's drawing feedback, captions and cap announcement (S14-03).

## Writer's own calls
- K-4 pulls two items of `ideas/depictions-and-likenesses.md` into S12-27 (decode the crop before the session; one stage-then-write entry point), because the staged-object type it asks for covers them. The idea file is trimmed.
- K-15: dependencies that are only merge order are now written as a preferred order in the "run in parallel" paragraph. Real dependencies kept: S12-22 → S12-38, S12-25 → S12-26 (K-2's keyed effect), S12-23 → S12-28 (K-8), S12-32 → S12-33 (the pin test's allow-list), S12-21 → S12-37 (K-19).
- K-20: the four-column Sources list text is fixed in this edit (working rule, S12-14 scope, the S12-D5 design row and `completed.md`), so S12-20 no longer carries it.
- S12-24 keeps the index migration and the pointer releases together (K-20 nit); each fits within ten steps.

## Round 1

### K-1 · blocker · Land S12-34's per-key descriptor before the slice 6 cache work, and give it a patch slot
**Answer:** Fixed (user, 2026-10-27). S12-34 is first in slice 6 and carries the `patch` and `refresh(ids)` slots; today's hand-written updates move into them, and S12-25 and S12-26 add their reads as entries.

### K-2 · blocker · One keyed "reads" effect instead of a new WriteEffects field per partial read
**Answer:** Fixed. Working rule "A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key"; S12-25 introduces `Effect.Reads` → `Set.Reads` → `WriteEffects.reads` in the shape of `Set.Search` (citation reads, graph progress); S12-26 adds the graph list by Source.

### K-3 · blocker · S12-22: extend the existing unlink preflight and link-set writer
**Answer:** Fixed. Working rule "One unlink check, one link writer": one helper behind `Unlink` and `UnlinkImpact`, `UnlinkImpact` over (graph, Artifact) pairs with graph and reason, `applyLinks` lifted from `evidencegraphs.Update`. S12-38 shows the new lock in the graph form.

### K-4 · blocker · S12-27: one mechanism for staging stored objects
**Answer:** Fixed. Working rule "File work and decoding happen before the catalog session"; S12-27 is renamed "File work before the catalog session" and covers `ingest.File`, `StageImage` / `StageDepictions` and `derivatives.Ensure`, with `use-catalog-session` updated.

### K-5 · major · S12-32 deletes `.promotedSubject`, but it drives the sidebar recount
**Answer:** Fixed. S12-32 adds a `counts` effects list on `sources` and `canonical_entities` creates and deletes; `navCountTriggers` goes, and `.promotedSubject` goes only when `conclusionTriggers` no longer needs it.

### K-6 · major · One wire and Go shape for "which graphs"
**Answer:** Fixed. `GraphLinks { repeated string graph_ids }` on both Artifact requests (absent: default; present and empty: none), one Go type; working rule "An Artifact is the parent" rewritten.

### K-7 · major · S12-22's graph-delete refusal should be an inbound probe, not a new extra gate
**Answer:** Fixed. Working rule "An Artifact with Citations keeps at least one graph" now uses an inbound probe listing the Citations and rewords `evidencegraphs.in_use`; S12-22 adds the `viaFormats` heading.

### K-8 · major · S12-23 and S12-28 each add a Swift allowlist check
**Answer:** Fixed. Working rule "…checked first, and Go decides the type": Go's `files` filter decides, Swift checks only extension, execute bit and executable types, and the viewer extends `ThumbnailDecodePolicy`. Dependency S12-23 → S12-28 added.

### K-9 · major · S12-35 reverses the FakeStore rule; settle the direction first
**Answer:** Fixed (user, 2026-10-27, option a). S12-35 removes the fake's engine rules and scripts values instead; the working rule binds every PR that adds a store method.

### K-10 · major · Bar 13 has no structural pin
**Answer:** Fixed. S12-32 adds the `api/ffi` pin test with an allow-list; S12-33 empties it. Working rule "Every audited write returns its effects, and a test holds them to it".

### K-11 · major · S12-30: put the announcements in the kit
**Answer:** Fixed. `PVToast` announces itself and `PVCallout` takes an opt-in error announcement; working rule "Kit components announce themselves". S12-30 no longer edits the files S12-38 rewrites.

### K-12 · major · What S12-38 makes dead is not listed
**Answer:** Fixed. S12-38 deletes `isOnlyArtifact`, `graphListReloaded()`, `L10n.Sources.artifactGraphOnlyArtifact` and the link and unlink RPCs on both sides (proto methods reserved), since the card is their last caller.

### K-13 · major · Skills and docs that S12-25, S12-26 and S12-27 make false are not named
**Answer:** Fixed. The spike-wide rule "Each PR updates the docs and skills it makes false" gains a slices 5–7 line; S12-20 points `add-catalog-model` at `effects/README.md`; the S12-25, S12-26 and S12-27 scope entries name their docs and skills.

### K-14 · major · S12-21: declare destination identity in `PlaceRegistry`
**Answer:** Fixed. Working rule "A destination declares its identity"; S12-21 applies `.id(…)` once from the resolved place.

### K-15 · minor · Merge-order preferences are written as dependencies; one real dependency is missing
**Answer:** Fixed (see Writer's own calls).

### K-16 · minor · S12-29: history rows should use the breadcrumbs' name lookup
**Answer:** Fixed. S12-29: one location-label function for breadcrumbs, history rows and the VoiceOver label.

### K-17 · minor · S12-24: make the optional-pointer releases `facetRelease` entries
**Answer:** Fixed. S12-24 adds two `facetRelease` entries and extends `TestFacetReleaseHonesty`.

### K-18 · minor · S12-36 leaves two of A-31's four per-kind lists
**Answer:** Fixed. S12-36 adds a projector table for `Reproject` and `RebuildAll` and a completeness test.

### K-19 · minor · S12-21's missing-graph state should not become another flag
**Answer:** Fixed. S12-21 derives it from the graph list; S12-37 depends on S12-21.

### K-20 · nit · Grouping, and the map's four-column text
**Answer:** Fixed (see Writer's own calls).

## Round 2

A fresh reviewer checked the round-1 answers and looked for new blockers in the fixes, S12-39 and the Spike 14 additions. **Verdict:** Fits with changes; every finding below is fixed in map text, and none needed a product call.

Round-1 answers: 19 of 20 resolved. K-8 held only for S12-23 (the viewer half is K-23). Notes on resolved answers, each fixed: `CreateArtifactRequest.graph_ids` (7) is reserved (K-6); S12-33's kept term RPCs still return effects, so the allow-list can empty (K-10); S12-38 goes before S12-30, which still opts two callouts it rewrites into announcing (K-11); S12-32 updates `add-workspace-place` (`navCountTriggers`), S12-24 updates `catalog-deletes.md` and `add-catalog-delete` (K-13); the declared identity replaces `ResolvedPlace.deepId` (K-14); S12-27 splits the in-session thumbnail into read, decode and write (K-4).

### K-21 · blocker · S12-25 cannot move `effects` onto `database.Querier`: import cycle
`core/database/listen.go` imports `effects`, and `Querier` lives in `database`.
**Answer:** Fixed. The item leaves S12-25 and goes back to `ideas/shared-list-helpers.md`, blocked with `graphcache` until a leaf package exists.

### K-22 · blocker · S12-26's by-Source graph-list read sits beside the catalog-wide `evidence_graphs` list key
**Answer:** Fixed. S12-26's `reads` entry replaces the list key: `effects.ListEvidenceGraphs`, the wire key, `CatalogList.evidenceGraphs`, its invalidation case and its pin row go; the working rule "The graph list has its own effect" is marked superseded; S12-26 updates `source-layer-data-model.md`.

### K-23 · blocker · S12-28's "no second type list" conflicts with `ThumbnailDecodePolicy`'s PNG-and-JPEG list
**Answer:** Fixed. The viewer checks that the sniffed type matches the media type Go returned, plus a pixel budget, through one shared check; `ThumbnailDecodePolicy` keeps PNG and JPEG for thumbnails only. Working rule and S12-28 updated.

### K-24 · blocker · S12-39's design said the decoder holds a decode slot `raster` takes; callers take it
**Answer:** Fixed. The decoder runs under its caller's `raster.Slot()` and takes none; the S12-39 Check adds a header-only JPEG bomb.

### K-25 · blocker · S14-03 and S12-28 both change whether the viewer can cut a file
**Answer:** Fixed. One predicate, `ArtifactRegionCrop.snapshot`, changed in place by S12-28; S14-03 depends on S12-28 in both maps.

**Stop:** round 2 is the last round. No blocker is open and **To user** is answered.
