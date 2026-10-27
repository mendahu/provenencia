# Ship log — S12-15 Graphs in the omnibar

**Plan:** [`S12-15-graph-search.md`](S12-15-graph-search.md) · **Branch:** `s12-15-graph-search` · **PR:** #583
**State:** ready

## Round 0 — implement
Head `b591110a`. Gate PASS (orchestrator run against `4ed7efa6`). Report deviations: AC-1.5 and AC-1.7 asserted on the target graph by id rather than "no graph hit" for the whole query (multi-token OR retrieval still matches the other household on `hou`; noted in the plan's Landed note); `graphHit` fixture helper; T-5 as a closure-built `rpcTest` case; Step 8 docs in the final commit; §10 manual check and app Back/Forward not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — UX for `macos/App/**` and `Localizable.xcstrings`; SEC+PRF for `core/database/migrations/000056.sql`, `api/ffi/**`, `core/search`, `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | FakeStore graph loop copies the Source loop's scoring and ref gating | Fix now → F-1 | verified: `FakeStore.swift` 1433–1466 and 1475–1506 are the same block bar `source.ref`/`graph.ref` and the values dict; same-pass cleanup (this PR adds the copy) |
| MNT-2 | MNT | major | Second Go builder of the graph-canvas location; second location struct, section and surface constants | Fix now → F-2 | verified: `search.locationFor` graph case vs `deleteimpact.graphLocation` (projectors.go:303); `search.WorkspaceLocation` vs `deleteimpact.Location`; `search.SurfaceGraph` vs `deleteimpact.surfaceGraph`; two proto mappers (`search_catalog.go:70` `locationToProto`, `delete_impact.go:98` `locationFromImpact`). Same-pass cleanup: this PR adds a copy, so it supersedes the plan's Out row (review M-9 was a plan-review answer, not a user ruling). No researcher-visible change |
| MNT-3 | MNT | minor | `add-searchable-kind` §3 still points at a direct reproject call, not the effects registry | Fix now → F-3 | verified `SKILL.md:89`; skill file this PR edits |
| MNT-4 | MNT | minor | `TestEvidenceGraphSearchFollowsWrites` write closures fail through the parent `t` | Fix now → F-4 | verified closures capture outer `t`; `f.t` is the subtest's |
| MNT-5 | MNT | minor | `graphSearchFixture.search` shadows the identical `handleFixture.search` | Fix now → F-4 | verified embedded `*handleFixture` |
| CON-1 | CON | minor | §10 manual check not run in the app | Rejected | not a code defect; the manual check is the user's step after ship (ship-pr §8.4), carried to the report |

Nits (logged): `CON-N1 · S12-15 plan §2 · AC-1.5/1.7 wording ("no graph hit") cannot hold under OR retrieval; Landed note records it` · `UX-1 · manual check not run (= CON-1)` · `UX-2 · keep §10 wording in step with OR retrieval (= CON-N1)` · `MNT-6 · graphs_test.go:158 · lowerASCII → strings.ToLower` · `MNT-7 · graphs_test.go:276 · deleted Source's graph ref passed through f.smith.Ref` · `MNT-8 · fts.go:485 · locationFor locals; read from d` (MNT-6–8 ride along in F-2/F-4 since those lines are rewritten)

SEC+PRF: no findings. UX: no findings above nit.

### Fix plan 1

**F-1 — One FakeStore scorer for ref-bearing records** (from MNT-1)
Files: `macos/App/Platform/FakeStore.swift`.
Failing test first: none (refactor; existing `findsEvidenceGraphByNameRefAndDescription` and the Source search tests in `SearchCatalogStoreTests.swift` / `OmnibarResultsModelTests.swift` must stay green).
Do: replace `refBoost(for:key:)` with one private static helper, e.g. `scoreRefRecord(values:ref:tokens:raw:refKey:inSources:) -> (score: Double, reason: String)?`, that holds the `[("title", 10), ("ref", 12), ("description", 3)]` field list, the `tokens.isEmpty ? [raw.lowercased()] : tokens` fallback, the ref gating (`max(score, 12)`, reason `"ref"`, skip on no match), the `1.1` Sources-context boost and the ref boost, and returns nil when the record is skipped. Both loops call it and only build their `CatalogSearchHit`. Delete `refBoost`.
Check: `grep -n "refBoost\|(\"description\", 3)" macos/App/Platform/FakeStore.swift` prints only lines inside the new helper; scripts/check-pr.sh PASS.

**F-2 — One Go workspace location shared by search and delete impact** (from MNT-2; supersedes the plan's Out row "One Go workspace-location type … suggested follow-up task (review M-9)")
Files: new `core/workspacelocation/` (package `workspacelocation`); `core/search/search.go`, `core/search/fts.go`, `core/search/*` users; `core/database/deleteimpact/kinds.go`, `projectors.go` and other users; `api/ffi/handlers/search_catalog.go`, `api/ffi/handlers/delete_impact.go`; tests that build either struct.
Failing test first: `TestGraph` in `core/workspacelocation` — `Graph(sourceID, graphID, "SRC-1", "1881 Census")` returns Section `sources`, SourceSurface `graph`, Ref `SRC-1`, Title and SourceTitle `1881 Census`; with an empty title, Title is `SRC-1`.
Do: the new package owns `type Location` (the union of today's two structs' fields, including `EntityID` and `GraphID`), the section constants (`Sources`, `SourceTypes`, `Metadata`, `Properties`, `Files`, `Persons`, `Events`, `Places`), the surface constants (`SurfacePage`, `SurfaceGraph`, `SurfaceCitationComposer`), and `Graph(sourceID, graphID, sourceRef, sourceTitle string) Location` carrying `deleteimpact.graphLocation`'s rule (Title falls back to the ref). Delete `search.WorkspaceLocation`, `search.Section*`, `search.SurfaceGraph`, `deleteimpact.Location`, the `deleteimpact` section/surface constants and `graphLocation`; every user moves to the new package (no type aliases). `search.locationFor`'s graph case calls `workspacelocation.Graph(d.displayParentID, d.entityID, d.displayParentRef, d.displaySubtitle)`; `deleteimpact` passes `database.UUIDString` ids. Fold `locationToProto` and `locationFromImpact` into one mapper over `workspacelocation.Location` (keep it in `api/ffi/handlers`, the file that already owns location mapping); keep `locationFromProto` beside it. Read from `d` throughout `locationFor` (MNT-8). Delete T-9 `TestEvidenceGraphHitLocationMatchesDeleteImpact` (one builder cannot drift) and update the plan's Landed note: Out row superseded by the same-pass cleanup rule; T-9 removed. Update `add-searchable-kind` / `add-catalog-delete` skill text if either names the removed types or constants. Skill: `add-catalog-delete` for the deleteimpact side.
Check: `grep -rnE "type (WorkspaceLocation|Location) struct" core api` prints only `core/workspacelocation`; `grep -rnE "surfaceGraph|SurfaceGraph *=|\"graph\"" core api --include='*.go' | grep -v _test` prints only `core/workspacelocation`; `grep -rn "func locationToProto\|func locationFromImpact" api` prints at most one mapper; `TestGraph` passes; scripts/check-pr.sh PASS.

**F-3 — Searchable-kind skill names the effects registry path** (from MNT-3)
Files: `.agents/skills/add-searchable-kind/SKILL.md`.
Do: in §3 replace "Call reproject from domain mutators in the same transaction when practical." with: add a `Search` doc on the table's `effects` registry entry (with the parent fan-out, as `source_types` and `sources` do), a case in `searchindex.Reproject`, and a loop in `RebuildAll`; plus one line that a new display column goes through `searchindex.Document`/`Upsert` and `core/search` `docColumns`/`scanDoc`.
Check: `grep -n "effects" .agents/skills/add-searchable-kind/SKILL.md` prints the new line; scripts/check-pr.sh PASS.

**F-4 — Graph search test hygiene** (from MNT-4, MNT-5, nits MNT-6, MNT-7)
Files: `core/search/graphs_test.go`.
Do: every `write` closure in `TestEvidenceGraphSearchFollowsWrites` uses `f.t` (including `seedType`, `SoleGraph`, `t.Fatal`); delete `graphSearchFixture.search` (inherited from `*handleFixture`); replace `lowerASCII` with `strings.ToLower`; give the deleted-Source case its own fixture field instead of reusing `f.smith.Ref`.
Check: `grep -n "lowerASCII\|func (f \*graphSearchFixture) search" core/search/graphs_test.go` prints nothing; `go test -tags fts5 ./core/search/` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `ed9a1ffb`. Gate PASS (orchestrator run). F-1 `256977dd`, F-2 + F-4 `ed9a1ffb`, F-3 `ab9dd201`. Deviations: F-4 shares F-2's commit; `add-catalog-delete` named none of the removed types, left unchanged; `workspacelocation.Graph` trims the title before the ref fallback (as `titleOrRef` did).
F-3 (MNT-3): **Resolved** — §3.2 names the effects `Search` doc, `searchindex.Reproject`, `RebuildAll`, `docColumns`/`scanDoc`. F-4 (MNT-4, MNT-5): **Resolved** — grep clean, `f.t` in write closures.

## Round 2 — review
Reviewers: MNT (had Fix-now majors MNT-1, MNT-2). Scope `b591110a...ed9a1ffb`. Not sent: CON, UX, SEC+PRF (no Fix-now blocker or major).
MNT-1: **Resolved** (`scoreRefRecord`, `refBoost` deleted). MNT-2: **Resolved** (`core/workspacelocation` owns `Location`, constants and `Graph`; one `locationToProto`; T-9 deleted). No new findings.

Nits (logged): `MNT-9 · core/workspacelocation/workspacelocation.go:52 · Graph inlines the trim-then-ref rule beside deleteimpact.titleOrRef; could export one TitleOrRef` · `MNT-10 · S12-15 plan:152 · Codebase fit row still names search.SurfaceGraph (Landed note supersedes)` · `MNT-11 · deleteimpact connection.go:5, probes.go:6, release_test.go:6 · unrelated import-grouping blank lines`

## Finish
Merged `origin/main` @ `06446e5d` (S11-42 plan docs only). Gate PASS against `06446e5d`.

## Result
Gate PASS · CI green · open: none · follow-ups: MNT-9 (export one `TitleOrRef`), Swift `SourceSurface` pin in `swift_vocabulary_test.go` (the other half of the plan's Out row) · to user: AC-1.5/1.7 wording (CON-N1) · CON miss: none
