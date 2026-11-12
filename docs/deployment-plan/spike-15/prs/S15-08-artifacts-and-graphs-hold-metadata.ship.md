# Ship log — S15-08 Artifacts and graphs hold metadata

**Plan:** [`S15-08-artifacts-and-graphs-hold-metadata.md`](S15-08-artifacts-and-graphs-hold-metadata.md) · **Branch:** `s15-08-artifacts-and-graphs-hold-metadata` · **PR:** #791
**State:** ready

## Round 0 — implement
Head `473789da`. Gate PASS (orchestrator run). PR #791 (draft). Report deviations: D-18's `registry[table].Entity` was an init cycle, so `entityOf(table)` loops over the existing `entityTable` (no new map); `TestListByOwnersTx`'s refused owner moved from `graph` to `citation`; T-16 isolates each row's own scope; test placement as recorded in the plan's Landed note (`TestImpactMetadataFieldHolders`, raw-SQL cases in `metadatafields`, T-9 as subtests); manual check not run (no view change). One implementer gate run failed `promotealign` `TestLargeSourceBudgets/cold` (2.72s vs 2.5s) while two `check-pr.sh` runs overlapped; clean runs pass.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `Localizable.xcstrings`, `api/ffi/**`, migrations, `core/database/**`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1, CON-3, UX (outside), PRF (outside) | MNT, CON | blocker | `metadataFieldsList` Used by has no engine signal; Artifact/graph writes and metadata-releasing deletes leave it stale | To user → **Fix now → F-1** | User (2026-11-11): "do the invalidation correctly… a really robust cache" → engine list signal. Verified: the four new `effects/registry.go:346-361` entries declare no `Lists`; `CatalogQueryRegistry.swift:207` follows only `.mutatedSourceMetadata`, sent only by `SourcePageContext.swift:142`. Fix (a `ListMetadataFieldUsage` list) contradicts §6 Scale "no other key or cache is invalidated" |
| PRF-1 | PRF | blocker | `metadatafields.List` now 1 + 3 × fields `COUNT(DISTINCT owner)` statements on every Source metadata edit | To user → **Fix now → F-2** | User (2026-11-11) chose `COUNT(*)` + batched `InboundTotals`, superseding D-5's "keep one form" and D-8's per-row loop. sqlite3 on the PR's DDL: 0.78 ms → 19.4 ms at Scale, 96.7 ms at 4×; non-covering index search. Fix contradicts D-5 (`COUNT(*)` rejected, "keep one form") and D-8 (per-row `InboundTotal`) |
| MNT-2, PRF-2 | MNT, PRF | major | Three identical `UsedBy` bodies and three "scan, then `InboundTotal` per row" `List` loops | Fix now → F-2 | Same-pass cleanup: `metadatafields`, `sourcetypes`, `properties` copies all added by this PR; `properties.UsedBy` has only a test caller |
| MNT-3 | MNT | major | `releaseGraph` hand-builds a single-row removal; graph audit columns listed twice | Fix now → F-3 | Same-pass cleanup (supersedes D-7's "`rowFacet` does not change"): `releaseGraph`, `releaseCitation`, `releaseClaimRow` share the shape; `graphFields` has one caller now |
| MNT-4 | MNT | minor | `entityOf` scans `entityTable` per hop | Fix now → F-4 | Minor in code this PR adds; build a reverse map in the same `init()` |
| MNT-5 | MNT | minor | Board-entry → proto loop written five times; unused `*Catalog` param | Fix now → F-5 | Minor in code this PR adds |
| SEC-1 | SEC | minor | `artifactBoardsProto` round-trips ids through `parseID`; a non-v7 Artifact id fails the whole Source page | Fix now → F-6 | Traced `artifactProto` → `UUIDString` → `parseID` (v7 check) → `errInvalidID`. Use the raw bytes |
| CON-1 | CON | minor | AC-1.2 test never shows cross-level suggestions are excluded | Fix now → F-7 | `TestOwnerTables` suggests only at its own level |
| CON-2 | CON | minor | T-16 uses hand-built changes, not the real `artifacts.Delete` / `evidencegraphs.Delete` → `Resolve` | Fix now | No test resolves a real delete revision's metadata changes |
| UX-1 | UX | minor | Inspector label "Used on" beside a count that now includes terms | To user → **Fix now → F-9** | User: "In use by". Copy call; D-17 changed the value string, not the label |
| UX-2 | UX | minor | `error.metadatafields.in_use` still says "sources" only | To user → **Fix now → F-9** | User approved the wording. `L10n.swift:11272`; `ErrInUse` now fires for Artifact/graph holders |
| UX-3 | UX | minor | Copy Source Metadata tick hint "Values and field layout" reads as all metadata | To user → **Fix now → F-9** | User approved the wording. D-10 keeps Artifact metadata regardless of the tick |

Nits (logged): `UX-4 · L10n.swift:8928-8942 · one casing/name for "Evidence graph" across the three strings` · `MNT-6 · sourcecopy.go:76,95-104 · one ReadDraftsTx call` · `MNT-7 · sourcecopy.go:115-117, metadatafields.go:324 · rewrap doc comments` · `CON-4 · metadatafields.go:70-72 · UsedBy doc "Only List" (Update also sets it)` · `CON-5 · CatalogQueryRegistry.swift:349 · comment says Observations only` · `CON-6 · source_defs_test.go · fail when field missing` · `CON-7 · FakeStore.swift:2872,2932 · usedBy as group total`

### Fix plan 1

Supersedes, on the user's answers (2026-11-11): §6 Scale "no other key or cache is invalidated" (F-1); D-5's "`COUNT(*)` … keep one form" Rejected column and D-8's per-row `InboundTotal` loop (F-2); D-7's "`rowFacet` does not change" (F-3, same-pass cleanup). Record each in the plan's Landed note.

**F-1 — The field list's Used by refreshes from an engine list signal** (from MNT-1, CON-3)
Files: `core/database/effects/registry.go`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, the Swift `CatalogList` enum, `.agents/skills/add-workspace-place/SKILL.md:172`, `macos/App/Features/Sources/SourcePageContext.swift` (only if `.mutatedSourceMetadata` loses its last follower on this key). Skill: `add-workspace-place`.
Failing test first: an `effects` test that `Resolve` over (a) an `artifact_metadata` insert, (b) an `evidence_graph_metadata` delete, (c) a `source_metadata` insert reports `metadata_field_usage` in `Lists`, and an update that changes only `value` on any of them does not; plus the real `artifacts.Delete` of an Artifact holding one value reports it.
Do: add `ListMetadataFieldUsage List = "metadata_field_usage"` beside `ListPropertyUsage`, add it to `Lists`, and set `Lists: []List{ListMetadataFieldUsage}, ListsOnCreateDelete: true` on `source_metadata`, `artifact_metadata`, `evidence_graph_metadata` (not the layout tables: layout rows hold no value). Add the Swift `CatalogList` case and its row in the swift vocabulary test. `metadataFieldsList` follows `.list(.metadataFieldUsage)` the way `propertiesWorkspace` follows `.list(.propertyUsage)`, keeps `.list(.sourceVocabulary)`, and drops `invalidateOn: [.mutatedSourceMetadata]`. Fix the key's comment ("counts Sources, Artifacts and Evidence graphs holding a value") and the CON-5 comment at `:349` (Observations and terms). Update the skill line.
Check: the new tests pass; `rg -n "mutatedSourceMetadata" macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` prints only the Source-workspace key's lines (`:238`, `:246`); scripts/check-pr.sh PASS.

**F-2 — One batched Used by total; metadata edges count `COUNT(*)`** (from PRF-1, PRF-2, MNT-2)
Files: `core/database/deleteimpact/register.go`, `core/database/deleteimpact/impact.go`, `core/database/metadatafields/metadatafields.go`, `core/database/sourcetypes/sourcetypes.go`, `core/database/properties/properties.go`, `api/ffi/handlers/source_defs.go`, their tests.
Failing test first: `deleteimpact` `TestInboundTotals`: for 3 fields (one held on a Source, one on an Artifact and a graph, one unheld) and for Properties (Observations plus terms), `InboundTotals(tx, kind, ids)` equals `InboundTotal` per id, including 0 for an unheld id and with more ids than one `sqlbatch` chunk.
Do: (1) The three metadata `joinEdge` counts become `SELECT COUNT(*) FROM <table> WHERE field_id = ?` (`UNIQUE (owner, field_id)` makes it equal; note it in a comment). (2) Add `InboundTotals(tx *sql.Tx, kind catalogmodel.Kind, ids [][]byte) (map[string]int, error)` in `impact.go`: per edge of `inboundFor(kind)`, one grouped statement `SELECT <col>, COUNT(*) … WHERE <col> IN (…) GROUP BY <col>` through `sqlbatch`, summed. Each `inboundEdge` needs its grouped form: derive it from the edge's table and column (the `Via` string) for `fkEdge` and the metadata `joinEdge`s; an edge that cannot be grouped (`subjectResourceInbound`, credibility) falls back to its per-id `Count` only if its kind is never passed to `InboundTotals` — otherwise give it an explicit grouped SQL. (3) Add one `UsedBy(c *database.Catalog, kind, id) (int, error)` in `deleteimpact` that opens the read tx; `metadatafields`, `sourcetypes`, `properties` `UsedBy` wrappers are deleted and `source_defs.go:85,287` (and `properties.Update`'s in-tx read, via `InboundTotal`) call it. (4) Each `List` scans its rows, closes them, then sets `UsedBy` from one `InboundTotals` call — one shared helper if the three bodies still match. Fix `Field.UsedBy`'s doc (CON-4: List and Update).
Check: `TestInboundTotals` passes; existing `UsedBy`/`List`/`TestImpactMetadataFieldHolders` tests pass; `rg -n "func UsedBy" core/` prints only `deleteimpact`; `rg -n "InboundTotal\(" core/database/{metadatafields,sourcetypes,properties}` prints nothing inside a loop; scripts/check-pr.sh PASS.

**F-3 — One single-row release helper** (from MNT-3)
Files: `core/database/deleteimpact/facets.go`, `core/database/evidencegraphs/evidencegraphs.go`.
Failing test first: none new; `TestDelete/a_graph_releases_its_metadata`, `TestFacetReleaseAuditsSourceSide` and the Citation release tests must stay green with byte-identical change JSON.
Do: factor `rowFacet`'s single-row snapshot (select the `col` list by id, `releaseFacets(kind)`, delete, append `DeletedRow` built with `colJSON`) into one helper; `releaseGraph` and `releaseCitation` use it with the `col` lists the deleted graph `rowFacet` had (`ref`/`description` as nullable text). `releaseClaimRow` joins only if it can without changing its callers' signatures. Fix `graphFields`' doc (create only).
Check: listed tests pass; `rg -n "rowchange.DeletedRow\(" core/database/deleteimpact/facets.go` shows only the helper (and `releaseClaimRow` if it stays — say so in the report); scripts/check-pr.sh PASS.

**F-4 — `entityOf` reads a reverse map** (from MNT-4)
Files: `core/database/effects/path.go`, `registry.go`. Do: fill a table→entity map in the same `init()` that builds `entityTable`; `readRowsColumn` reads it. Check: `TestScopeThroughDeletedParent` passes.

**F-5 — `metadataEntriesProto`** (from MNT-5)
Files: `api/ffi/handlers/metadata.go`, `sources.go`. Do: add `metadataEntriesProto(entries)`; the five loops call it; drop `metadataEntryProto`'s unused `*database.Catalog`. Check: handler tests pass.

**F-6 — Artifact boards key on raw ids** (from SEC-1)
Files: `api/ffi/handlers/metadata.go`, `sources.go`, `metadata_test.go`.
Failing test first: `TestGetSourceWorkspace/artifact_with_non_v7_id` — insert an Artifact whose id is a 16-byte non-v7 UUID; the Source workspace still loads with its board.
Do: `artifactBoardsProto` takes the raw `[][]byte` ids from the `artifacts.ListBySource` rows `GetSourceWorkspace` already reads; no `parseID` round trip.
Check: test passes.

**F-7 — Cross-level suggestion exclusion is tested** (from CON-1)
Files: `core/database/sourcemetadata/sourcemetadata_test.go`, `api/ffi/handlers/metadata_test.go`.
Do: in each `TestOwnerTables` subtest also suggest a second field at the other level and assert it is absent; in `newMetadataOwners` suggest one field at `artifact` only and one at `graph` only, and assert each board lists only its own. Check: tests pass (and fail if `listWorkspaces` keyed suggestions without level).

**F-8 — Scope through a real delete** (from CON-2)
Files: `core/database/artifacts/artifacts_test.go` or `effects` tests.
Do: keep the stub test; add one case that runs the real `artifacts.Delete` (one value, one layout row) and the real `evidencegraphs.Delete`, then resolves each released metadata change on its own, with the full revision as ghosts, to the Source scope. Check: passes.

**F-9 — Copy** (from UX-1, UX-2, UX-3; user, 2026-11-11). Skill: `add-localized-string`.
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: `properties.inspector.usedOn` → "In use by" (comment: "Inspector label beside the count of Observations and terms using the Property"); `error.metadatafields.in_use` → "That metadata field still holds a value on a Source, Artifact or Evidence graph."; the Copy Source Metadata hint → "This Source's values and field layout. Artifacts keep their own." Update both files together.
Check: localizable check PASS; scripts/check-pr.sh PASS.

Riding along (nits, optional, only if touched anyway): MNT-6, MNT-7, CON-6.

## Round 1 — fix
Head `e6f4fa06`. Gate PASS (orchestrator run). F-1 – F-9 done. Deviations: F-8 resolves the whole real revision (no public per-change `Resolve`), so the per-change ghost scope stays pinned by the stub test only; F-7 is a new `TestMetadataBoardsKeepTheirLevel` instead of changing `newMetadataOwners`; F-6 gives `listArtifactsProto` a second return value (raw ids); F-3 leaves `releaseClaimRow` out of `releaseRow` (its callers need the columns back); `InboundTotals` falls back to per-id `Count` for the credibility edge and `subjectResourceInbound` (no grouped form). Nits MNT-6, MNT-7, CON-6 not done; CON-4, CON-5 rode along.

## Round 2 — review
Reviewers: MNT, SEC+PRF (Fix-now blocker or major in round 1). Not sent: CON, UX (minors only; orchestrator checks their fixes below).

Orchestrator check of CON and UX minors (fix commits `473789da..e6f4fa06`):
- UX-1, UX-2, UX-3 (F-9): **Resolved**. `L10n.swift` and `Localizable.xcstrings` carry "In use by" (comment updated), "That metadata field still holds a value on a Source, Artifact or Evidence graph." and "This Source's values and field layout. Artifacts keep their own."
- CON-1 (F-7): **Resolved**. `TestOwnerTables` suggests `f2` at the other level and asserts it is absent; `TestMetadataBoardsKeepTheirLevel` suggests an artifact-only and a graph-only field and checks each board and the Source page's Artifact boards.
- CON-2 (F-8): **Not resolved**, partly. The real `artifacts.Delete` / graph `Delete` revisions now go through `effects.Resolve` and give the Source scope, but resolved as a whole batch the Artifact's own delete already scopes the Source, so the test cannot tell whether each released metadata change resolves through its ghost parent. Missing: resolve each released change alone with the rest of the revision as ghosts (an in-package `effects` test over the real changes, or an exported test hook). Per-change ghost resolution stays pinned only by the stub `TestScopeThroughDeletedParent`. Minor; carried to the result as open.
- CON-3: closed by F-1 (comment fixed; engine list signal).

MNT round 2: **Clean**, no new blocker. MNT-1, MNT-2, MNT-3 **Resolved** (`ListMetadataFieldUsage` on the three value tables, Swift mirror checked by `swift_vocabulary_test.go` from `effects.Lists`; one `deleteimpact.UsedBy` / `FillUsedBy` / `InboundTotals`, every edge of the three passed kinds has `Totals`; `releaseRow` is the single-row path for `releaseCitation` and `releaseGraph`, change JSON unchanged).

SEC+PRF round 2: **Clean**, no new blocker. PRF-1 **Resolved** (grouped `COUNT(*)` per edge on covering indexes through `sqlbatch`; sqlite3 on the table shapes: 1.8 + 2.1 ms at 40k + 40k values, 7.3 + 7.2 ms at 4×, linear, against 19.4 / 96.7 ms in round 1). PRF-2 **Resolved** (`FillUsedBy` tail in all three `List`s). SEC-1 **Resolved** (raw ids; `TestGetSourceWorkspace/artifact_with_non_v7_id`).

Nits (logged, round 2): `MNT · facets.go · releaseRow and rowFacet both build the column list and field map` · `MNT · CatalogQueryRegistry.swift · .mutatedSourceMetadata now acts like .mutatedSourceWorkspace on its only key`

No Fix plan 2 (no round-2 blockers). `origin/main` unmoved since the branch was cut (`cba6e263`); gate on `e6f4fa06` stands.

## Result
Gate PASS · CI green (test ×2, localizable) · open: CON-2 (minor, per-change ghost resolution over a real delete) · follow-ups: CON-2 → S15-09 map scope entry (user, 2026-11-11), nits MNT-6, MNT-7, CON-6, CON-7, UX-4 · to user: MNT-1, PRF-1, UX-1 – UX-3 (answered 2026-11-11, fixed)
