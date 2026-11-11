# Review — S15-08 Artifacts and graphs hold metadata

**Verdict:** round 1 — Fits with changes / Not ready
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- **`properties.UsedBy`** (M-3): it counts `observations` only, while a Property's delete is also blocked by `property_terms.property_id`, so its "used by" can read 0 on a Property that cannot be deleted. Recommendation: a separate `quick-pr` that moves it onto `deleteimpact.InboundTotal` after this PR lands, or a doc line saying why it differs. Not in S15-08.

## Writer's own calls
- Owner RPCs replace the four Source board RPCs (renamed, same enum numbers) rather than living beside them (D-11); the Swift store keeps its method names until S15-26 (D-16).
- One action name per operation at every level (`update_metadata`, `dismiss_metadata_suggestion`, `reorder_metadata`); Source writes stop recording the old names (D-12). The action is stored only.
- `GetMetadataBoard` is added now for S15-11's graph form (D-13).
- Metadata on `UpdateArtifactRequest` is not planned: an Artifact is edited on its live board.
- The Source delete's graph release runs each graph's facets (D-7), found in research: today graph metadata would CASCADE unaudited.

## Round 1

### M-1 · blocker · The plan adds four audit-scope entries on a lookup it knows is broken
`effects/path.go` `readRowsColumn` looks up `entityTable[table]`, but `entityTable` is keyed by entity, so second-hop ghosts never match; the plan deferred the fix as Out.
**Answer:** Fixed — D-18 (`tableEntity`, `readRowsColumn`), AC-1.6, T-16 (per row, and `citation_notes`), STEP-1; the Out row is gone; STEP-7 updates `add-audit-scope`; a Stop condition guards existing scopes.

### M-2 · major · `rowFacetOf` adds a mode to `rowFacet` for one caller
**Answer:** Fixed — D-7 is now `releaseGraph` / `ReleaseGraph` beside `releaseCitation`, used by the Source facet (through `releaseEach`) and by `evidencegraphs.Delete`; `rowFacet` unchanged.

### M-3 · major · `sourcetypes` keeps its own copy of the used-by count
**Answer:** Fixed for `sourcetypes` — D-8, AC-2.5, T-17, STEP-2. `properties` → To user (its count differs from its delete total).

### M-4 · major · The draft step is written into three handlers
**Answer:** Fixed — D-12 `parseDraft` (before the session, `nil` when absent) and `applyDraft`; D-14 one line per handler.

### M-5 · minor · No check that `stores` covers every metadata level; sweep misses names
**Answer:** Fixed — T-2 asserts `stores` keys equal `vocabkey.MetadataLevels()`; DoD sweep and STEP-4 name all three old actions and their doc, skill and rule hits; test labels move to the new names.

### M-6 · minor · `listArtifactsProto` grows a second job; two wire shapes for a board
**Answer:** Fixed — D-15 `artifactBoardsProto` in `metadata.go`, `listArtifactsProto` unchanged; D-11 `GetMetadataBoardResponse { MetadataBoard board = 1; }`.

### X-1 · blocker · STEP-1's check needs D-5
**Answer:** Fixed — D-5 moved into STEP-1, which says why.

### X-2 · blocker · `protoc` and Xcode are not on this machine
**Answer:** Fixed — header **Toolchain** line names what each step needs; the Stop condition says to finish STEP-1 – STEP-3 and report. The plan cannot install the tools: running STEP-4 – STEP-6 needs a machine that has them.

### X-3 · major · Sweep names hits no step edits
**Answer:** Fixed — STEP-4 lists `.cursor/rules/catalog-deletes.mdc`, `add-catalog-delete`, `docs/audit-revision-history.md` and the § 5.4 line with their replacements; its `rg` check covers them.

### X-4 · major · STEP-4 omits `source_defs_test.go`
**Answer:** Fixed — listed in STEP-4; core test action labels move to the new names (M-5).

### X-5 · major · `rowFacetOf` hides an entity type from `TestResolversCoverEveryEntityType`
**Answer:** Fixed by M-2 — no `rowFacetOf`; `releaseGraph` carries an `EntityType: "evidence_graph"` literal the scan finds (STEP-2 says so).

### X-6 · major · Order across several graphs not decided
**Answer:** Fixed — D-7: graph by graph in id order (`releaseEach`); AC-3.4; T-5 two-graph case.

### X-7 · minor · `ScanField` argument and `List` transaction
**Answer:** Fixed — D-8 names both.

### X-8 · minor · Handler choices
**Answer:** Fixed — D-12: parse order, `unmarshalErr` labels; `parseDraft` errors are `request.invalid_id` before any write (T-13); D-9: no owners → `nil, nil`.

### X-9 · minor · Fixtures for T-3 and T-5
**Answer:** Fixed — STEP-1 says fixtures are built inline with existing helpers.

### X-10 · nit
Logged.
