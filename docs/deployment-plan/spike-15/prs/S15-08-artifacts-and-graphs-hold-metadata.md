# S15-08 — Artifacts and graphs hold metadata

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the engine, the wire and three strings in existing copy; the delete notice draws its groups as today, and S15-09 – S15-11 add every board; user, 2026-11-11)
**Board:** —
**Depends on:** S15-06 (landed [#782](https://github.com/mendahu/provenencia/pull/782)), S15-07 (landed [#787](https://github.com/mendahu/provenencia/pull/787))
**Blocks:** S15-09, S15-10, S15-11, S15-16
**Branch:** `s15-08-artifacts-and-graphs-hold-metadata`
**Version bump:** minor, 0.9.1 → **0.10.0** (if `main`'s `VERSION` is no longer 0.9.1 when you rebase, the next minor after it): additive schema and new RPCs ([`docs/versioning.md`](../../../versioning.md)); `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** working rules "Metadata lives at three levels with one vocabulary", "One level type", "Metadata saves with its owner's write", "Suggestions have a level", "Citations render on demand, from SQL" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`evidence-explained-citation-map.md`](../evidence-explained-citation-map.md) § 6.6 (the DDL and "Knock-on work"); [`source-layer-data-model.md`](../../../source-layer-data-model.md) §§ 4.2, 5, 7, 7.1; [`catalog-deletes.md`](../../../catalog-deletes.md); [S15-07](S15-07-metadata-operations-take-an-owner.md) **Landed** and **Next adds**
**Toolchain:** STEP-1 – STEP-3 and STEP-7's docs need Go only. STEP-4 and STEP-5 need `protoc`, `protoc-gen-go` and `protoc-gen-swift` on `PATH` (`./scripts/generate-proto.sh`); STEP-6 and the Swift half of STEP-4 need macOS with Xcode (`scripts/check-pr.sh` runs `xcode-test`). On a machine without them, finish STEP-1 – STEP-3 and stop (§11)
**Planned:** 2026-11-11 against `main` @ `01fa954`
**Review:** [`S15-08-artifacts-and-graphs-hold-metadata.review.md`](S15-08-artifacts-and-graphs-hold-metadata.review.md) — round 1 answered; round 2 pending

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

An Artifact (the copy used) and an Evidence graph (the item cited) hold metadata values and a layout, with the same fields, the same rules and the same audit as a Source's. The engine stores them in the four tables of map § 6.6, wired through S15-07's per-level store table; releases them, audited, when their Artifact, graph, field or Source goes; counts them in a field's **Used by** and refuses a field delete while any of them holds a value; carries a chosen Artifact's metadata through Copy Source; and serves them over FFI: one set of board RPCs that take an owner, a board read, a draft on the Artifact and graph writes, and the Artifacts' boards on the Source page's read. No board is drawn yet.

**Why now.** Working rules "Metadata lives at three levels with one vocabulary" and "Metadata saves with its owner's write". S15-09, S15-10 and S15-11 draw the boards over these RPCs; S15-16 reads Artifact and graph values through `ListByOwnersTx`.

**What stays.** The Source page, the Source board, Copy Source's form and every view look and behave as today. The Source board's RPCs change name and take an owner on the wire, but the Swift store methods keep their signatures (S15-26 changes those). Three strings change: the Source types toast's count names all three holders, and the field delete notice gains two group headings (§6 **Strings**).

## 2. User stories

**US-1.** As the engine, I keep metadata on Artifacts and graphs exactly as on Sources, so that a census Artifact holds its provider, URL and accessed date and a household graph its dwelling, family and line.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An Artifact owner and a graph owner; `Set`, `Clear`, `DismissSuggestion`, `Reorder`, `Apply`, `CreateAll`, `ReadDraftTx`, `ListByOwnersTx`, `ListWorkspace` | Each works as for a Source: rows in `artifact_metadata` / `artifact_metadata_layout` and `evidence_graph_metadata` / `evidence_graph_metadata_layout`; changes carry `artifact_id` / `graph_id` as the owner column |
| AC-1.2 | A census-like type with `provider` suggested at `artifact` and `line` at `graph`; `ListWorkspace` on an Artifact and a graph of a Source of that type | The Artifact lists `provider` (suggested), the graph lists `line` (suggested); neither lists the other's, nor the type's `source` suggestions |
| AC-1.3 | `Set` on an Artifact owner | The revision's source scope is the Artifact's Source (`audit_transaction_scopes` row `source`, the Source id); the same for a graph owner |
| AC-1.4 | Any function with an owner at `citation`, `""` or `page`, or an Artifact owner with a 15-byte id | `ErrInvalid`; nothing written |
| AC-1.5 | `Set` on an Artifact owner whose id has no `artifacts` row | `ErrInvalid` |
| AC-1.6 | Artifact A with one value; `artifacts.Delete` | Each released `artifact_metadata` and `artifact_metadata_layout` change resolves (`effects.Resolve`) to A's Source scope, through A's delete change; likewise a `citation_note` released by a Citation delete resolves to its Source |

**US-2.** As a researcher, a field that holds a value anywhere cannot be deleted, and its **Used by** says how many records hold it, so that deleting a field never drops an Artifact's or a graph's value.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Field F with a value on one graph only; `GetDeleteImpact(metadata_field, F)` | Not allowed, gate `inbound`; one group, via `evidence_graph_metadata.field_id`, kind `evidence_graph`, total 1, listing the graph |
| AC-2.2 | Field F with values on Sources S1, S2, Artifact A1 and graph G1 | Groups `source_metadata.field_id` (2), `artifact_metadata.field_id` (1), `evidence_graph_metadata.field_id` (1); `metadatafields.UsedBy` = 4; `metadatafields.List` row F `UsedBy` = 4 |
| AC-2.3 | Field F only dismissed on an Artifact (a layout row, no value), and suggested by a type | Deletable; the delete records one `artifact_metadata_layout` delete change carrying `artifact_id` and `field_id`; `UsedBy` = 0 |
| AC-2.4 | `metadatafields.Delete` on F of AC-2.1 | `metadatafields.ErrInUse` |
| AC-2.5 | Source type T on two Sources; `sourcetypes.UsedBy` and the `sourcetypes.List` row | 2, equal to `InboundTotal(KindSourceType, T)` |

**US-3.** As the engine, an Artifact's or a graph's metadata goes with it, audited, so that history shows what was released.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Artifact A with one value and two layout rows, no Citations; `artifacts.Delete` | Changes include one `artifact_metadata` delete (fields `id`, `artifact_id`, `field_id`, `value_text`) and two `artifact_metadata_layout` deletes (`artifact_id`, `field_id`, `sort_order`, `dismissed`), before the `artifact` delete; no row left |
| AC-3.2 | Graph G with one value and one layout row; `evidencegraphs.Delete` | The same with `evidence_graph_metadata` / `evidence_graph_metadata_layout` and `graph_id`, before the `evidence_graph` delete |
| AC-3.3 | A Source with no Artifacts whose default graph holds one value; `sources.Delete` | Changes include the `evidence_graph_metadata` delete before that graph's `evidence_graph` delete; no `evidence_graph_metadata` row left (today the CASCADE would drop it unaudited) |
| AC-3.4 | A Source with no Artifacts and two graphs G1 < G2 (by id), each with one value | Changes per graph in id order: G1's metadata deletes, G1's `evidence_graph` delete, then G2's, then G2's |

**US-4.** As a researcher, Copy Source carries a chosen Artifact's own metadata, so that the copy's Artifact keeps its provider and URL (user, 2026-11-11: always, whatever the Metadata tick).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Original with Artifacts A1 (provider "FamilySearch", a dismissed suggestion) and A2 (no metadata); copy choosing A1 only, Metadata unticked | The copy's Artifact has `provider` "FamilySearch" and the same layout rows as A1, written by `CreateAll` after that Artifact's create; the copy's Source has no metadata |
| AC-4.2 | The same with Metadata ticked | Source metadata copied as today, Artifact metadata as AC-4.1 |
| AC-4.3 | The original is untouched | A1's rows are unchanged after the copy |

**US-5.** As the app, I edit any owner's board and save a form's draft with its owner, so that S15-09 – S15-11 need no engine work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `SetMetadata` with owner `{artifact, A}` and a field | Returns the entry and `WriteEffects` whose `source_scopes` holds A's Source |
| AC-5.2 | `ClearMetadata`, `DismissMetadataSuggestion`, `ReorderMetadata` with an Artifact owner and a graph owner | As the Source RPCs do today for a Source: effects, and for dismiss and reorder the owner's full list |
| AC-5.3 | `SetMetadata` with owner level `citation` or `""` | `sourcemetadata.invalid`; with an id that is not a UUIDv7, `request.invalid_id` |
| AC-5.4 | The Source page's metadata through the renamed RPCs with owner `{source, S}` | The same replies and audit as today's `SetSourceMetadata` etc. (the existing handler tests pass with call sites renamed) |
| AC-5.5 | `GetMetadataBoard` with a graph owner | The graph's entries: its type's `graph` suggestions plus extras, by its layout |
| AC-5.6 | `CreateArtifact` with `metadata` holding `provider` "FamilySearch" and layout `[provider 0]` | One revision: the `artifact` create, then one `artifact_metadata` create and one `artifact_metadata_layout` create; the Artifact's board shows the value |
| AC-5.7 | `CreateArtifact` without `metadata` | Exactly today's changes |
| AC-5.8 | `CreateArtifact` with `metadata` naming an unknown field | `sourcemetadata.invalid`; no Artifact created (one transaction) |
| AC-5.9 | `CreateEvidenceGraph` with `metadata` holding `line` "14" | One revision with the graph and its value |
| AC-5.10 | `UpdateEvidenceGraph` without `metadata` | Values and layout unchanged; with `metadata` `{values: []}` on a graph holding `line`: `line` is deleted (an empty draft clears every value, as in S15-07's `TestApply`) |
| AC-5.11 | `GetSourceWorkspace` on a Source with Artifacts A1 (one value) and A2 (none) | `artifact_metadata` holds one `MetadataBoard` per Artifact, in `artifacts` order: A1's entries with the value, A2's suggestions only |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration `000065.sql`: the four tables and their `field_id` indexes | STEP-1 |
| `catalogmodel` Tables and FKs; `effects` entries scoping by the owner's Source | STEP-1, AC-1.3 |
| `sourcemetadata` stores for `LevelArtifact` and `LevelGraph` | STEP-1, AC-1.1, AC-1.2, AC-1.4, AC-1.5 |
| Field delete: two inbound probes; layout releases on the field | STEP-1, AC-2.1, AC-2.3, AC-2.4 |
| `deleteimpact.InboundTotal`; `metadatafields.UsedBy` and `List` from it | STEP-2, AC-2.2 |
| Artifact and graph deletes release their metadata | STEP-1, AC-3.1, AC-3.2 |
| One graph removal, `deleteimpact.ReleaseGraph`, used by `evidencegraphs.Delete` and a Source delete | STEP-2, AC-3.2 – AC-3.4 |
| A deleted parent's change is found on the second hop of a scope path (`effects` ghost lookup) | STEP-1, AC-1.6 |
| `sourcetypes.UsedBy` and its `List` count from `InboundTotal` | STEP-2, AC-2.5 |
| `sourcemetadata.ListWorkspaces` and `ReadDraftsTx` (many owners) | STEP-3, AC-5.11, AC-4.1 |
| Copy Source carries each chosen Artifact's metadata | STEP-3, AC-4.1 – AC-4.3 |
| Owner RPCs replace the four Source metadata RPCs; `GetMetadataBoard`; Swift `GoStore` and `CoreInvoke` call them | STEP-4, AC-5.1 – AC-5.5 |
| `metadata` drafts on `CreateArtifactRequest`, `CreateEvidenceGraphRequest`, `UpdateEvidenceGraphRequest`; Artifact boards on `GetSourceWorkspaceResponse` | STEP-5, AC-5.6 – AC-5.11 |
| Strings: two delete-notice via headings; the Source types toast count | STEP-6 |
| Docs and version | STEP-7 |

| Out | Home |
| --- | --- |
| Any board in the app: the Artifact rows, the add flow, the graph form | S15-09, S15-10, S15-11 |
| Swift store methods that take an owner; `SourceMetadataSection` on an owner | S15-26 |
| Metadata on `UpdateArtifactRequest` (an Artifact is edited on its live board) | not planned |
| Search over Artifact or graph values | not planned (map S15-08 "Not this PR") |
| The type page's three suggestion lists | S15-12 |
| Reading Artifact and graph values into citations | S15-16 |
| Moving an Artifact's or a graph's metadata to another Source | S15-24 (it follows `artifact_id` / `graph_id`, map § 6.6) |
| `properties.UsedBy` (counts `observations` only, while `KindProperty` also has the `property_terms.property_id` edge) | To the user (review M-3): a separate quick fix or a doc line |
| Renaming `source_metadata_fields` or the package `sourcemetadata` | not planned (map § 6.6 "Name") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Stores | `core/database/sourcemetadata/owner.go` `stores`, `newStore`, `storeFor`, `requireOwner` | One entry, `vocabkey.LevelSource`; every other level is `ErrInvalid`. `newStore(level, ownerTable, ownerColumn, values, layout, typeIn)` builds all SQL from constant names; `typeIn` must select `(owner id, Source type id)` up to its `IN (` |
| Package | `core/database/sourcemetadata/sourcemetadata.go` | Package doc says "a Source today; an Artifact and an Evidence graph (S15-08)…". `ListWorkspace(c, owner)` opens a read tx and calls the unexported `listWorkspaces(tx, owners) map[string][]WorkspaceEntry` (keyed by `Owner.key`). `ReadDraftTx(tx, owner)` reads one owner through `readOwnerTx`. `listByOwners`, `listLayoutByOwners` already batch many owners per level through `sqlbatch.Query` |
| Tests | `core/database/sourcemetadata/sourcemetadata_test.go` `TestStoresAreRegistered`, `TestOwnerLevels` (asserts every level but `source` is refused), `TestListWorkspaces`, `TestApply`, `TestListByOwnersTx`; helpers `testWorld`, `newSource`, `birthOfAlice`, `applyWorld`, `values`, `countRows` | Source owners only |
| Schema | `core/database/migrations/000028.sql` (`source_metadata`), `000013.sql` (`source_metadata_layout`), `000031.sql` (indexes `source_metadata_field_id_idx`, `source_metadata_layout_field_id_idx`), `000064.sql` (latest) | The DDL to mirror; no Artifact or graph metadata table |
| Owner tables | `000008.sql` `artifacts` (`source_id` NOT NULL), `000055.sql` `evidence_graphs` (`source_id` NOT NULL, `ref` nullable UNIQUE) | Both reach their Source in one column |
| Catalog model | `core/database/catalogmodel/model.go` Tables `source_metadata`, `source_metadata_layout` (`BucketFacet`, no PK, no Kind); FKs `source_metadata.source_id` (CASCADE, Facet, Audited), `source_metadata.field_id` (NO ACTION, Resource), `source_metadata_layout.source_id` and `.field_id` (CASCADE, Facet, Audited) | `core/database/catalogmodel/pragma_test.go` `TestPragmaHonesty` refuses an unregistered table, a wrong FK and an FK column with no covering index |
| Effects | `core/database/effects/registry.go` `"source_metadata"`, `"source_metadata_layout"` (`SourceScope: field("source_id")`, `Search` DocSource); `"citations"` `SourceScope: up("artifact_id", "source_id")`; `"evidence_graph_artifacts"` `SourceScope: up("graph_id", "source_id")` | `registry_completeness_test.go` `TestEveryNonSkipTableHasAnEntry`, `TestCascadeIntoNonEmptyEffectIsAudited`, `TestSourceEvidenceIsNotActivity` (a pinned table list) |
| Field delete probe | `core/database/deleteimpact/register.go` `inboundEdgeList`: `joinEdge(KindMetadataField, "source_metadata.field_id", KindSource, COUNT(DISTINCT source_id)…, …JOIN sources s…)` | Only Source values block a field delete |
| Facet releases | `core/database/deleteimpact/facets.go` `buildFacetReleases`: Artifact (`evidence_graph_artifacts.artifact_id`, `sources.primary_artifact_id`), graph (`ViaDeletesCitations`, `evidence_graph_artifacts.graph_id`), Source (`source_notes`, `evidence_graphs` as a plain `rowFacet`, `source_metadata`, `source_credibility_assessments`, `source_metadata_layout`), field (`source_metadata_layout.field_id`); `rowFacet`, `releaseEach`, `releaseCitation`, `releaseFacets` | The Source's graph release deletes graph rows directly, so a graph's own facets never run on a Source delete. `pragma_test.go` `TestFacetReleaseHonesty` requires one release per Audited CASCADE |
| Impact | `core/database/deleteimpact/impact.go` `Impact`, `collectGroups`; `register.go` `inboundEdge.Count`, `inboundFor` | Group totals come from each edge's `Count`; nothing exported sums them |
| Used by | `core/database/metadatafields/metadatafields.go` `sqlList` (subquery `COUNT(*) FROM source_metadata`), `sqlUsedBy`, `List`, `UsedBy`, `Field.UsedBy`, `ErrInUse`, `Delete` | A private copy of the in-use rule over `source_metadata` only (map review K-8). `metadatafields` already imports `deleteimpact` |
| Artifact and graph writes | `core/database/artifacts/artifacts.go` `Create`, `Delete` (Impact → Refuse → `ReleaseFacets` → `SnapshotOwned` → DELETE → `ReleaseSnapshot`); `core/database/evidencegraphs/evidencegraphs.go` `Create`, `Update` (`ErrInvalid` for a missing graph), `Delete` | Neither imports `sourcemetadata`; `sourcemetadata` imports neither (`go list -deps`) |
| Copy Source | `core/database/sourcecopy/sourcecopy.go` `Draft.Artifacts []DraftArtifact`, `DraftArtifact{FromID, In}`, `ReadDraft` (`artifacts.ListBySourceTx`), `CreateFromDraft` (discards the new Artifact: `_, more, err := artifacts.Create(…)`), `apply` (filters by `CopyInput.ArtifactIDs`) | Artifacts are copied without metadata |
| Proto | `api/proto/engine.proto` `METHOD_SET_SOURCE_METADATA = 19`, `METHOD_CLEAR_SOURCE_METADATA = 20`, `METHOD_DISMISS_SOURCE_METADATA_SUGGESTION = 39`, `METHOD_REORDER_SOURCE_METADATA = 40`; their request and response messages (`SetSourceMetadataRequest` fields 1–5, `reserved 6, "date"`; `ClearSourceMetadataRequest`, `DismissSourceMetadataSuggestionRequest` fields 1–4; `ReorderSourceMetadataRequest` fields 1–4); `MetadataWorkspaceEntry`; `MetadataField.used_by = 7` ("How many source_metadata rows…"); `GetSourceWorkspaceResponse` (next free 9); `CreateArtifactRequest` (next free 9; `GraphLinks graph_links = 8` uses message presence); `CreateEvidenceGraphRequest`, `UpdateEvidenceGraphRequest` (next free 7) | Source-keyed board RPCs; the highest method is 118 |
| Handlers | `api/ffi/handlers/sources.go` `GetSourceWorkspace`, `SetSourceMetadata` (re-reads `ListWorkspace`, picks the entry; none → `CodeInternalUnknown`), `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata`, `metadataEntryProto`; `api/ffi/handlers/artifacts.go` `CreateArtifact`, `listArtifactsProto` (one caller: `GetSourceWorkspace`); `api/ffi/handlers/evidence_graphs.go` `CreateEvidenceGraph`, `UpdateEvidenceGraph`; `api/ffi/handlers/catalog.go` `withProjectCatalog`, `parseID`, `parseIDs`, `parseUserID`; `api/ffi/dispatch.go` `route` | Writes go through `writes.Run` with actions `update_source_metadata`, `dismiss_source_metadata_suggestion`, `reorder_source_metadata` (stored as `audit_transactions.action_type`, shown nowhere in the app) |
| Handler tests | `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace`, `TestSetSourceMetadata`, `TestDismissAndReorderSourceMetadata`, `TestCopySource`; `artifacts_test.go` `TestCreateArtifactAndIngest`, `TestDeleteArtifact`; `evidence_graphs_test.go` `TestCreateEvidenceGraph`, `TestUpdateEvidenceGraph`, `TestDeleteEvidenceGraph`; `source_defs_test.go` `TestDeleteMetadataField`, `TestUpdateMetadataField`; `write_effects_test.go` `TestEveryAuditedWriteReturnsEffects` | Source owners only |
| Domain tests | `core/database/deleteimpact/facets_test.go` `TestFacetReleaseAuditsSourceSide`; `core/database/sourcecopy/sourcecopy_test.go` `TestCopy`; `core/database/audit/scopes_test.go` `TestSourceScopes`; `core/database/metadatafields/metadatafields_test.go` | — |
| Swift | `macos/App/Platform/GoStore.swift` `setSourceMetadata`, `clearSourceMetadata`, `dismissSourceMetadataSuggestion`, `reorderSourceMetadata` (build `Provenencia_Engine_V1_SetSourceMetadataRequest` etc. with `sourceID`); `macos/App/Platform/CoreInvoke.swift` `CoreMethod.setSourceMetadata` and the other three; `macos/App/Platform/Generated/engine.pb.swift` | The `GenealogyStore` protocol, `FakeStore` and `OnboardingModelTests` stub do not name proto types and stay unchanged |
| Swift copy | `macos/App/Platform/L10n.swift` `"source_metadata.field_id"` via heading ("%lld sources record a value for this field"); `L10n.SourceTypes.toastRemovedBody` and `removedBodyWithValues` ("…the %3$lld sources already carrying a value keep it."), called from `macos/App/Features/SourceTypes/SourceTypesModel.swift` with `field.usedBy`; `macos/App/Features/CatalogVocabulary/VocabularyPane.swift` `CatalogVocabularyRow.usedBy` doc ("How many sources refer to this row"); `macos/App/Resources/Localizable.xcstrings` | Name Sources only |
| Docs | `docs/source-layer-data-model.md` §§ 4.2, 5.3, 5.4, 7, 7.1, 10; `docs/catalog-deletes.md` "Facet release" table | Source-level metadata only |

**Precedent.** S15-06 ([#782](https://github.com/mendahu/provenencia/pull/782)) for a level on the wire as a plain string checked by Go, and for regenerating the proto with the Swift store edits in the same commit. `rowFacet(KindSource, "source_metadata", …)` and `rowFacet(KindSource, "source_metadata_layout", …)` for the four new releases: same columns with the owner column renamed. `joinEdge(KindMetadataField, "source_metadata.field_id", …)` for the two new probes. `releaseCitation` for a release that runs a child's own facets before deleting it. `GraphLinks` on `CreateArtifactRequest` for an optional sub-message whose presence means "given".

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000065.sql` creates the four tables exactly as map § 6.6 (owner FK `ON DELETE CASCADE`; value `field_id` NO ACTION; layout `field_id` CASCADE; `UNIQUE (owner, field_id)` on values; `PRIMARY KEY (owner, field_id)` on layout), `STRICT`, plus one index on `field_id` per table: `artifact_metadata_field_id_idx`, `artifact_metadata_layout_field_id_idx`, `evidence_graph_metadata_field_id_idx`, `evidence_graph_metadata_layout_field_id_idx`. No separate owner index: the `UNIQUE` and the PK lead with the owner | Map § 6.6 is decided; the field index serves the delete probe and the field's layout release (`TestPragmaHonesty` covering-index rule) | A redundant owner index as `source_metadata_layout` has (`000031.sql`) |
| D-2 | `catalogmodel.Tables`: the four tables, `Bucket: BucketFacet`, no PK, no Kind, beside `source_metadata`. `catalogmodel.FKs`: `artifact_metadata.artifact_id` → `artifacts` CASCADE Facet Audited; `artifact_metadata.field_id` → `source_metadata_fields` NO ACTION Resource; `artifact_metadata_layout.artifact_id` → `artifacts` and `.field_id` → `source_metadata_fields`, both CASCADE Facet Audited; the same four for the graph tables with `graph_id` → `evidence_graphs` | Mirrors the Source rows; no Kind keeps `swift_vocabulary_test` "every delete-impact kind has a noun" unchanged | A Kind per table — nothing deletes one by id |
| D-3 | `core/database/effects/registry.go`: `"artifact_metadata"` `{Entity: "artifact_metadata", SourceScope: up("artifact_id", "source_id")}`, `"artifact_metadata_layout"` `{Entity: "artifact_metadata_layout", SourceScope: up("artifact_id", "source_id")}`, `"evidence_graph_metadata"` `{Entity: "evidence_graph_metadata", SourceScope: up("graph_id", "source_id")}`, `"evidence_graph_metadata_layout"` `{Entity: "evidence_graph_metadata_layout", SourceScope: up("graph_id", "source_id")}`. No `Search`, `Reads`, `EvidenceGraph` or `Handles`. `TestSourceEvidenceIsNotActivity` gains the four rows `false, true` | Rule "audit under the Source's scope"; S15-07 D-3 uses the table names as entity types. The Source scope reaches the app as `WriteEffects.source_scopes`, which refreshes the Source workspace that now carries Artifact boards. Search is Out; no graph read carries metadata | A `Reads` on `ReadEvidenceGraphs` — the graph list carries no metadata, and S15-11 reads the board on open |
| D-4 | `owner.go` `stores` gains `vocabkey.LevelArtifact: newStore(vocabkey.LevelArtifact, "artifacts", "artifact_id", "artifact_metadata", "artifact_metadata_layout", "SELECT a.id, s.source_type_id FROM artifacts a JOIN sources s ON s.id = a.source_id WHERE a.id IN (")` and `vocabkey.LevelGraph: newStore(vocabkey.LevelGraph, "evidence_graphs", "graph_id", "evidence_graph_metadata", "evidence_graph_metadata_layout", "SELECT g.id, s.source_type_id FROM evidence_graphs g JOIN sources s ON s.id = g.source_id WHERE g.id IN (")`. Nothing else in `sourcemetadata` changes for the levels. The package doc and `Owner`'s doc say "a Source, an Artifact or an Evidence graph" | S15-07 built the store table for exactly this (its **Next adds**); suggestions come from the owner's Source's type at the owner's level (rule "Suggestions have a level") | — |
| D-5 | Field delete: `inboundEdgeList` gains, after the Source edge, `joinEdge(KindMetadataField, "artifact_metadata.field_id", KindArtifact, "SELECT COUNT(DISTINCT artifact_id) FROM artifact_metadata WHERE field_id = ?", "SELECT a.id, a.ref FROM artifact_metadata m JOIN artifacts a ON a.id = m.artifact_id WHERE m.field_id = ? ORDER BY a.ref COLLATE NOCASE LIMIT ?")` and `joinEdge(KindMetadataField, "evidence_graph_metadata.field_id", KindEvidenceGraph, "SELECT COUNT(DISTINCT graph_id) FROM evidence_graph_metadata WHERE field_id = ?", "SELECT g.id, COALESCE(g.ref, '') FROM evidence_graph_metadata m JOIN evidence_graphs g ON g.id = m.graph_id WHERE m.field_id = ? ORDER BY COALESCE(g.ref, '') COLLATE NOCASE LIMIT ?")`. `buildFacetReleases` gains `rowFacet(KindMetadataField, "artifact_metadata_layout", "field_id", "artifact_metadata_layout", "field_id", col{"artifact_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool})` and the same for `evidence_graph_metadata_layout` with `graph_id` | Map scope: "a field in use in any of the three value tables; layout rows cascade and do not keep a field". `evidence_graphs.ref` is nullable (`000055.sql`), as `fkEdge` handles with `COALESCE` | `COUNT(*)` — equal under `UNIQUE`, but the Source edge counts distinct owners; keep one form |
| D-6 | Owner deletes: `buildFacetReleases` gains, in the Artifact block, `rowFacet(KindArtifact, "artifact_metadata", "artifact_id", "artifact_metadata", "id", col{"id", colUUID}, col{"artifact_id", colUUID}, col{"field_id", colUUID}, col{"value_text", colText})` and `rowFacet(KindArtifact, "artifact_metadata_layout", "artifact_id", "artifact_metadata_layout", "field_id", col{"artifact_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool})`, and the same two in the graph block with `KindEvidenceGraph`, `graph_id` and the graph tables. `artifacts.Delete` and `evidencegraphs.Delete` do not change: they already call `ReleaseFacets` for their kind | The `rowFacet` precedent; the change rows carry the owner column so `up(...)` and the ghost map scope them (`add-audit-scope`) | — |
| D-7 | One graph removal. `deleteimpact.releaseGraph(tx *sql.Tx, graphID []byte) (released, error)` *(new, in `facets.go`, beside `releaseCitation`)*: reads `id, ref, source_id, name, description` from `evidence_graphs` (no row → `sql.ErrNoRows`), runs `releaseFacets(tx, KindEvidenceGraph, id)`, deletes the row, and appends the `evidence_graph` delete change with `rowchange.DeletedRow` over those five columns (`ref` and `description` `nil` when NULL, as `graphFields` writes them). `ReleaseGraph(tx *database.Tx, graphID []byte) ([]rowchange.Change, error)` *(new)* wraps it as `ReleaseCitation` does (purge after commit). Callers: (1) the Source's `evidence_graphs` entry becomes `{Parent: KindSource, Via: "evidence_graphs.source_id", Remaining: "SELECT COUNT(*) FROM evidence_graphs WHERE source_id = ?", Release: releaseEach("SELECT id FROM evidence_graphs WHERE source_id = ? ORDER BY id", releaseGraph)}`, so each graph's facets, then its row and change, run graph by graph in id order (AC-3.4); (2) `evidencegraphs.Delete` keeps `getTx`, `Impact` and `Refuse`, then returns `deleteimpact.ReleaseGraph(tx, graphID)` in place of its `ReleaseFacets`, `sqlDelete` and hand-built change. `graphFields` stays (the create change uses it). `rowFacet` does not change | One graph removal, as `releaseCitation` is the one Citation removal (review M-2). Without it a Source delete drops graph metadata by CASCADE, unaudited (AC-3.3). On a Source delete a graph has no links or Citations (Artifacts block the delete, `artifacts.source_id` NO ACTION), so only its metadata facets release | A `child` parameter on `rowFacet` for one caller; leaving `evidencegraphs.Delete` with its own copy of the delete change |
| D-8 | `deleteimpact.InboundTotal(tx *sql.Tx, kind catalogmodel.Kind, id []byte) (int, error)` *(new, in `impact.go`)*: the sum of `Count` over `inboundFor(kind)`, the same edges `Impact` groups; `ErrInvalid` for a nil `tx`, an unknown kind or an id that is not 16 bytes. `metadatafields.UsedBy(c, id)` opens a read transaction (`db.Begin`, deferred `Rollback`) and returns `InboundTotal(tx, KindMetadataField, id)`. `metadatafields.List` runs in one read transaction: the field rows (`sqlList` without its subquery), then `InboundTotal` per field. `sqlUsedBy` and the subquery are deleted; `List` drops the `&usedBy` argument to `ScanField` (whose variadic `extra` stays for `sourcevocab`) and sets `f.UsedBy` from `InboundTotal`. `sourcetypes` moves the same way (review M-3): `UsedBy` and `List` read `InboundTotal(tx, KindSourceType, id)` in one read transaction; `sqlUsedBy` and the `(SELECT COUNT(*) FROM sources …)` subquery in `sqlList` go, and `List`'s `Scan` drops `&t.UsedBy`. `Field.UsedBy`'s doc: "how many Sources, Artifacts and Evidence graphs hold a value for this field: the total that blocks its delete" | Map review K-8: Used by comes from the delete-impact totals, not a private copy. User, 2026-11-11: it counts every holder, and type suggestions do not count (they never block) | Moving `properties.UsedBy` too — its count differs from its delete total (`property_terms`), a product call sent to the user; a three-subquery `sqlList` — the private copy K-8 removed; `Snapshot` per field — lists rows it does not need |
| D-9 | `sourcemetadata.ListWorkspaces(c *database.Catalog, owners []Owner) ([][]WorkspaceEntry, error)` *(new)*: `storeFor` on each owner, one read transaction, `listWorkspaces`, and the result at index `i` is `owners[i]`'s list (a repeated owner gets the same list). `ListWorkspace(c, owner)` becomes its one-owner case. `ReadDraftsTx(tx *sql.Tx, owners []Owner) ([]Draft, error)` *(new)*: `storeFor` on each, nil `tx` is `ErrInvalid`, `listByOwners` and `listLayoutByOwners` once, index `i` is `owners[i]`'s draft; `ReadDraftTx` becomes its one-owner case. Both return `nil, nil` for no owners, without a query or a transaction (as `ListByOwnersTx`), so a Copy of a Source with no Artifacts reads nothing | S15-07 **Next adds**: "S15-08 exports a many-owner board read over `listWorkspaces`". Index-aligned results keep `Owner.key` unexported | A map keyed by owner — exports `key` for one caller |
| D-10 | Copy Source: `DraftArtifact` gains `Metadata sourcemetadata.Draft`. `ReadDraft` fills it for every Artifact with one `sourcemetadata.ReadDraftsTx` call over Artifact owners. `apply` keeps each chosen Artifact's `Metadata` whatever `CopyInput.Metadata` says. `CreateFromDraft` keeps the new Artifact from `artifacts.Create` and, right after its changes, appends `sourcemetadata.CreateAll(tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelArtifact, ID: art.ID}, a.Metadata)`'s. The package doc's list of what a copy carries adds "each chosen Artifact with its own metadata". No proto change | User, 2026-11-11: an Artifact's metadata is part of the Artifact; the Metadata tick covers the Source's own fields. `CreateAll` is S15-07's copy path | Tying it to the Metadata tick — refused by the user |
| D-11 | The four Source board RPCs become owner RPCs. Proto: `message MetadataOwner { string level = 1; string id = 2; }` *(new)*, where `level` is `"source"`, `"artifact"` or `"graph"` (any other value is `sourcemetadata.invalid`); enum values keep their numbers and change names: `METHOD_SET_METADATA = 19`, `METHOD_CLEAR_METADATA = 20`, `METHOD_DISMISS_METADATA_SUGGESTION = 39`, `METHOD_REORDER_METADATA = 40`; messages renamed `SetMetadataRequest` / `Response`, `ClearMetadataRequest` / `Response`, `DismissMetadataSuggestionRequest` / `Response`, `ReorderMetadataRequest` / `Response`; in each request `source_id = 3` becomes `reserved 3; reserved "source_id";` and `MetadataOwner owner` takes the next free number (7 in `SetMetadataRequest`, after its `reserved 6`; 5 in the other three). Responses keep their fields. New: `METHOD_GET_METADATA_BOARD = 119`, `GetMetadataBoardRequest { string project_dir = 1; MetadataOwner owner = 2; }`, `GetMetadataBoardResponse { MetadataBoard board = 1; }` (D-15's message, so a board has one wire shape) | Rule "One metadata board": one set of operations over an owner, so S15-09 – S15-11 call the same RPCs as the Source page. Renaming keeps one path; the Swift protocol keeps its methods until S15-26 | New owner RPCs beside the Source ones — two paths for one operation; a `level` field beside `source_id` holding an Artifact id — a misnamed field |
| D-12 | Handlers: `api/ffi/handlers/metadata.go` *(new)* holds `SetMetadata`, `ClearMetadata`, `DismissMetadataSuggestion`, `ReorderMetadata` (moved from `sources.go` and renamed, bodies unchanged but for the owner), `GetMetadataBoard` *(new)*, `parseOwner(o *engine.MetadataOwner) (sourcemetadata.Owner, error)` *(new)* (`nil` → `sourcemetadata.ErrInvalid`; id through `parseID`; level copied as `vocabkey.Level(o.GetLevel())`, refused by `storeFor`), `parseDraft(d *engine.MetadataDraft) (*sourcemetadata.Draft, error)` *(new)* (`nil` for an absent message; field ids through `parseID`, so a bad id is `request.invalid_id` before any write), `applyDraft(tx *database.Tx, userID []byte, owner sourcemetadata.Owner, d *sourcemetadata.Draft) ([]rowchange.Change, error)` *(new)* (no changes for `nil`, else `sourcemetadata.Apply`), `artifactBoardsProto` (D-15) and `metadataEntryProto` (moved). Each handler parses `user_id`, then the owner, then `field_id` / `field_ids`, so a bad owner reports first; `unmarshalErr` labels are `set_metadata`, `clear_metadata`, `dismiss_metadata_suggestion`, `reorder_metadata`, `get_metadata_board`. Write actions: `update_metadata` (set and clear), `dismiss_metadata_suggestion`, `reorder_metadata`, at every level. `dispatch.go` `route` cases and `Method…` constants renamed; one new case | One file per concern; the actions are stored only, so one name per operation is enough | Keeping the handlers in `sources.go` — they no longer belong to a Source; per-level action names — nothing reads them |
| D-13 | `GetMetadataBoard` reads `sourcemetadata.ListWorkspace(c, owner)` in `withProjectCatalog`. A missing owner is `sourcemetadata.invalid` (as `ListWorkspace` returns); no `sources.Get` first | S15-11's graph form reads the board on open; the Source page keeps `GetSourceWorkspace` | — |
| D-14 | Drafts on the owner writes: `message MetadataDraft { repeated MetadataValue values = 1; repeated MetadataLayoutRow layout = 2; }`, `message MetadataValue { string field_id = 1; string value_text = 2; }`, `message MetadataLayoutRow { string field_id = 1; int32 sort_order = 2; bool dismissed = 3; }` *(all new)*. `CreateArtifactRequest.metadata = 9`, `CreateEvidenceGraphRequest.metadata = 7`, `UpdateEvidenceGraphRequest.metadata = 7`, each `MetadataDraft`, absent = no metadata change (proto3 message presence, as `graph_links`). Each handler calls `parseDraft` before `withProjectCatalog`, and inside the same `writes.Run` closure, after `artifacts.Create` / `evidencegraphs.Create` / `evidencegraphs.Update`, appends `applyDraft(tx, userID, owner, draft)`'s changes (one line each, review M-4); the closure's error aborts the whole write | Rule "Metadata saves with its owner's write" (`Apply`, one transaction). The handler composes the write: `sources` imports `artifacts` and `evidencegraphs`, and S15-07 put composition in handlers | Calling `Apply` inside `artifacts.Create` — couples the package to `sourcemetadata` for one caller; `CreateAll` on create — the rule names `Apply` |
| D-15 | Artifact boards on the Source page: `message MetadataBoard { string owner_id = 1; repeated MetadataWorkspaceEntry entries = 2; }` *(new)*; `GetSourceWorkspaceResponse.artifact_metadata = 9` (`repeated MetadataBoard`), one per Artifact in `artifacts` order. `artifactBoardsProto(c *database.Catalog, arts []*engine.Artifact) ([]*engine.MetadataBoard, error)` *(new, in `metadata.go`)* parses each `Id` with `parseID`, calls `sourcemetadata.ListWorkspaces(c, owners)` once over their `LevelArtifact` owners, and builds the boards with `metadataEntryProto`; `GetSourceWorkspace` calls it after `listArtifactsProto`, which does not change | Map scope: "its Artifacts' boards in one query (60 Artifacts at Scale)". Board building lives beside `metadataEntryProto` (review M-6) | A metadata field on `Artifact` — every Artifact reply would read metadata; widening `listArtifactsProto` — a second job for a list function |
| D-16 | Swift: regenerate `engine.pb.swift`; `CoreInvoke.swift` renames the four statics to `setMetadata`, `clearMetadata`, `dismissMetadataSuggestion`, `reorderMetadata` (each `Int32(Provenencia_Engine_V1_Method.<name>.rawValue)`) and adds `getMetadataBoard`; `GoStore.swift`'s four methods build the renamed requests with `req.owner.level = Level.source.rawValue` and `req.owner.id = sourceID`. No protocol, `FakeStore` or view change | Keeps the app building and the Source page unchanged; S15-26 owns the Swift owner | Adding owner parameters to the protocol now — S15-26's scope |
| D-17 | Strings (user, 2026-11-11): `L10n.swift` via headings `"artifact_metadata.field_id"` = `deleteImpact.via.artifactMetadataFieldId` "%lld artifacts record a value for this field" and `"evidence_graph_metadata.field_id"` = `deleteImpact.via.evidenceGraphMetadataFieldId` "%lld Evidence graphs record a value for this field"; `sourceTypes.toast.removedBodyWithValues` becomes "%1$@ is no longer suggested for %2$@. The field stays in this project’s vocabulary, and the %3$lld Sources, Artifacts and graphs already carrying a value keep it." (key unchanged; its comment says "how many Sources, Artifacts and graphs hold a value"); `CatalogVocabularyRow.usedBy` doc: "How many records refer to this row — the delete gate"; proto comment on `MetadataField.used_by`: "How many Sources, Artifacts and Evidence graphs hold a value for this field (the delete-impact total). Deleting is only allowed at 0." | The count now names three holders; the delete notice names its new groups | — |
| D-18 | The second hop of a scope path finds a parent deleted in the same revision. `effects/registry.go` `init` also builds `tableEntity map[string]string` *(new)* (table → entity type, `tableEntity[t.Name] = eff.Entity` beside `byEntity`); `effects/path.go` `readRowsColumn` reads `tableEntity[table]` instead of `entityTable[table]`; `entityTable`'s doc comment is corrected to "maps an entity type to its catalog table" | Today `entityTable` is keyed by entity, so `entityTable["artifacts"]` is `""` and the ghost lookup never runs: a row released with its parent (D-6, and `citation_notes` today) resolves no scope, though `add-audit-scope` "Write changes so they resolve" says it does (review M-1). Fixing it makes the four new entries resolve as declared | Leaving it and asserting only the revision's scope — an Out line deferring a defect in a path this PR adds to |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `artifact_metadata`, `artifact_metadata_layout`, `evidence_graph_metadata`, `evidence_graph_metadata_layout` *(new)* | table | `source_metadata`, `source_metadata_layout` | Map § 6.6 decided per-level tables (a level column was rejected there) |
| `stores` entries for `LevelArtifact`, `LevelGraph` | registry entry | the `LevelSource` entry | Extends S15-07's table |
| Four effect entries | registry entry | `"source_metadata"`, `"citations"` | Extends the registry |
| Two `joinEdge`s, six `rowFacet`s | registry entry | `source_metadata.field_id` edge; Source metadata `rowFacet`s | Extends the registries |
| `deleteimpact.releaseGraph`, `ReleaseGraph` *(new)* | func | `releaseCitation`, `ReleaseCitation`; `evidencegraphs.Delete`'s tail; the Source's `evidence_graphs` `rowFacet` | Replaces both graph removals with one (D-7) |
| `effects.tableEntity` *(new)* | map | `entityTable` | The inverse map the ghost lookup needed (D-18) |
| `deleteimpact.InboundTotal` *(new)* | func | `Impact`, `collectGroups` | Totals without listing; reads the same edges (searched `rg -n "func .*Total" core/database/deleteimpact`) |
| `sourcemetadata.ListWorkspaces`, `ReadDraftsTx` *(new)* | func | `ListWorkspace`, `ReadDraftTx` | The one-owner forms become their cases (D-9) |
| `sourcecopy.DraftArtifact.Metadata` *(new)* | field | `Draft.Metadata` | Same type, per Artifact |
| `MetadataOwner`, `MetadataDraft`, `MetadataValue`, `MetadataLayoutRow`, `MetadataBoard`, `GetMetadataBoardRequest` / `Response` *(new)* | proto | `GraphLinks`, `MetadataWorkspaceEntry`, `TypeSuggestion` | No message carries an owner, a draft or a board today (searched `rg -n "message Metadata" api/proto/engine.proto`) |
| `api/ffi/handlers/metadata.go` *(new)*: `GetMetadataBoard`, `parseOwner`, `parseDraft`, `applyDraft`, `artifactBoardsProto` | file / handler / helper | `sources.go` metadata handlers, `parseGraphLinks` | The moved handlers plus one read; the parsers follow `parseGraphLinks`; `applyDraft` is the one draft step for three writes |
| `CoreMethod.getMetadataBoard` *(new)* | Swift static | the other `CoreMethod` statics | One per method |
| Two via headings *(new)* | L10n key | `deleteImpact.via.sourceMetadataFieldId` | One per Via (the recipe's map) |
| `TestOwnerTables`, `TestReadDraftsTx`, `TestListWorkspacesExported`, `TestInboundTotal`, `TestMetadataOwners`, `TestGetMetadataBoard`, `TestScopeThroughDeletedParent` *(new)* | test | `TestListWorkspaces`, `TestApply`, `TestSetSourceMetadata` | New behavior |

**Builds on.** S15-07's store table, `Apply`, `CreateAll`, `listWorkspaces`, `listByOwners`; the effects registry and `up(...)`; `deleteimpact`'s `rowFacet`, `joinEdge`, `releaseFacets`; `writes.Run`; `withProjectCatalog`.
**Deletes.** `metadatafields.sqlUsedBy` and the `COUNT(*) FROM source_metadata` subquery in `sqlList`; the Source-named Go handlers `SetSourceMetadata`, `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata` (renamed and moved, D-12); the proto names `METHOD_SET_SOURCE_METADATA`, `METHOD_CLEAR_SOURCE_METADATA`, `METHOD_DISMISS_SOURCE_METADATA_SUGGESTION`, `METHOD_REORDER_SOURCE_METADATA` and the eight Source-named messages (renamed, D-11); `dispatch.go`'s four Source-named constants; `CoreMethod`'s four Source-named statics; the actions `update_source_metadata`, `dismiss_source_metadata_suggestion`, `reorder_source_metadata` in handlers; `sourcetypes.sqlUsedBy` and its `sqlList` subquery; the Source's `rowFacet(KindSource, "evidence_graphs", …)` and `evidencegraphs.Delete`'s own release, `sqlDelete` call and change (both now `ReleaseGraph`); `readRowsColumn`'s `entityTable[table]` lookup; the old action strings in `core/database/sourcemetadata/sourcemetadata_test.go` and `core/database/deleteimpact/facets_test.go` (labels become the new names); `TestOwnerLevels`' expectation that `artifact` and `graph` are refused.
**Next adds.** S15-26 adds owner parameters to the Swift store methods over these RPCs; S15-09 reads `artifact_metadata` boards and writes through the owner RPCs; S15-10 sends `CreateArtifactRequest.metadata` with suggestions from `ListTypeSuggestions(level: artifact)`; S15-11 reads `GetMetadataBoard` and sends `UpdateEvidenceGraphRequest.metadata`; S15-16 calls `ListByOwnersTx` with Artifact and graph owners; S15-24's move needs nothing (the rows follow `artifact_id` / `graph_id`).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Artifact and graph owners reach every `sourcemetadata` function | `owner.go` `requireOwner`, `listWorkspaces` (`typeIn` joins to `sources`) | A missing owner: `requireOwner` → `ErrInvalid` (AC-1.5); `listWorkspaces` → `ErrInvalid` when the `typeIn` read finds no row. An owner always has a Source (`source_id` NOT NULL) |
| A Source delete now runs each graph's facets; a graph delete goes through `ReleaseGraph` | `deleteimpact.releaseGraph` | On a Source delete the graph has no links or Citations (Artifacts block it), so only metadata releases; `TestFacetReleaseAuditsSourceSide` pins the order, `TestDelete` (evidencegraphs) and `TestDeleteEvidenceGraph` the graph delete's |
| Ghost lookups now run on every second hop | `effects` `readRowsColumn` | A deleted parent's change is read from the ghost map, as the first hop already does; `citation_notes` under a Citation delete now resolve their Source (AC-1.6) |
| Copy Source to another type | the copy's Artifacts | Their values stay as extras when the new type does not suggest them, as Source metadata does today (`TestCopy` "other type") |
| A field delete blocked by an Artifact or a graph | the DeleteImpact recipe | Draws the group with its via heading (D-17); the listed rows use the existing `artifact` and `evidence_graph` projectors and nouns |
| The renamed Source board RPCs | `GoStore` | Same replies (AC-5.4) |

### Scale

At the map's Scale (a Source with up to 60 Artifacts; 63 fields; a census Source with 60 graphs):

- `GetSourceWorkspace`: Artifact boards cost one `ListWorkspaces` call: one `typeIn` query, one suggestion read per distinct (type, level) (one here), one values query, one layout query, at most one field query — 5 queries for 60 Artifacts, not 300.
- Copy Source of 60 Artifacts: `ReadDraftsTx` is 2 queries; `CreateAll` per copied Artifact writes only its rows.
- `metadatafields.List` (the Metadata page, 63 fields): one field query plus 3 indexed `COUNT`s per field, 190 queries in one read transaction; each count uses a `field_id` index (D-1, `source_metadata_field_id_idx`). `sourcetypes.List` (26 types): one count each on `sources_source_type_id_idx`, 27 queries.
- A field delete's Impact: 3 counts and at most 3 lists.
- A Source delete: one `releaseGraph` per graph (a row read, its facets' 4 releases, a delete). A Source that can be deleted has no Artifacts, so its graphs number one or two.
- Writes: an Artifact or graph metadata change scopes its Source, so the app reloads that Source's workspace (one `GetSourceWorkspace`); no other key or cache is invalidated, and no search document changes (D-3).

### Edges

- **Assumptions.** One value per (owner, field), held by each value table's `UNIQUE`; `Apply` keys drafts the same way and refuses duplicates (S15-07). `Owner.key` keeps levels apart, so an Artifact and a graph never share a board in `listWorkspaces`. `ListWorkspaces` returns index-aligned lists; a repeated owner gets the same list (T-6). The field probe counts distinct owners, equal to rows under `UNIQUE`.
- **Neighbors.** A field held only in a layout row (dismissed or ordered, no value) does not block its delete and releases audited (AC-2.3), at every level. A field held on a graph only blocks (AC-2.1). An absent `metadata` on a write changes nothing (AC-5.7, AC-5.10); a present empty draft clears every value (AC-5.10). A Copy Source with Metadata unticked still carries Artifact metadata (AC-4.1).
- **Removed inputs.** An Artifact delete drops its metadata from citations (S15-16 reads graph, else Artifact, else Source), audited (AC-3.1). A graph delete does the same for its item fields (AC-3.2). A Source delete releases its graphs' metadata (AC-3.3). A field removed from a type's suggestions leaves Artifact and graph values as extras, as on a Source.
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** A draft on a request is data the app sends, bounded by the field vocabulary: `apply` refuses a repeated field and an unknown field id after one batched field read (`loadFields` through `sqlbatch`), before any write, so a draft of any length costs one read per bind-limit batch. `ReorderMetadata`'s `field_ids` are unchanged.
- **Audit scope on release.** A released `artifact_metadata` row's scope needs `artifacts.source_id`, and on an Artifact delete that row is gone: the second hop reads the Artifact's delete change from the ghost map (D-18). AC-1.6 and T-16 pin it per row, and for `citation_notes` under a Citation delete.

### Schema

Migration `core/database/migrations/000065.sql` *(new)*:

```sql
-- S15-08: Artifacts and Evidence graphs hold metadata values and layout, with
-- the same fields as a Source (map § 6.6).

CREATE TABLE artifact_metadata (
	id          BLOB PRIMARY KEY,
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
	field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id),
	value_text  TEXT,
	UNIQUE (artifact_id, field_id)
) STRICT;
-- evidence_graph_metadata: the same with graph_id REFERENCES evidence_graphs(id).

CREATE TABLE artifact_metadata_layout (
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
	field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
	sort_order  INTEGER NOT NULL,
	dismissed   INTEGER NOT NULL DEFAULT 0 CHECK (dismissed IN (0, 1)),
	PRIMARY KEY (artifact_id, field_id)
) STRICT;
-- evidence_graph_metadata_layout: the same with graph_id.

CREATE INDEX artifact_metadata_field_id_idx ON artifact_metadata(field_id);
-- and the three other field_id indexes (D-1).
```

Registered in `catalogmodel` per D-2. Delete class: facets of their owner (audited CASCADE); `field_id` on values is a resource FK (probe, D-5); on layout an audited CASCADE (release, D-5). Audit: D-3. No backfill, no `foreign_keys off` (new tables only).

### Go core

| Package | Change |
| --- | --- |
| `core/database/migrations/000065.sql` *(new)* | Schema above |
| `core/database/catalogmodel/model.go` | D-2 |
| `core/database/effects/registry.go`, `registry_completeness_test.go` | D-3 |
| `core/database/sourcemetadata/owner.go` | D-4 |
| `core/database/sourcemetadata/sourcemetadata.go` | Package doc (D-4); `ListWorkspaces`, `ReadDraftsTx` (D-9) |
| `core/database/deleteimpact/register.go` | D-5 edges |
| `core/database/deleteimpact/facets.go` | D-5 layout releases, D-6, D-7 (`releaseGraph`, `ReleaseGraph`, the Source entry) |
| `core/database/evidencegraphs/evidencegraphs.go` | D-7 (`Delete` returns `ReleaseGraph`) |
| `core/database/effects/registry.go`, `path.go` | D-18 |
| `core/database/sourcetypes/sourcetypes.go` | D-8 (`UsedBy`, `List`) |
| `core/database/deleteimpact/impact.go` | D-8 `InboundTotal` |
| `core/database/metadatafields/metadatafields.go` | D-8 `List`, `UsedBy`, `Field.UsedBy` doc, `ErrInUse` doc names the three value tables |
| `core/database/sourcecopy/sourcecopy.go` | D-10 |

```go
func ListWorkspaces(c *database.Catalog, owners []Owner) ([][]WorkspaceEntry, error)
func ReadDraftsTx(tx *sql.Tx, owners []Owner) ([]Draft, error)
func InboundTotal(tx *sql.Tx, kind catalogmodel.Kind, id []byte) (int, error)
func ReleaseGraph(tx *database.Tx, graphID []byte) ([]rowchange.Change, error)
```

Errors: `sourcemetadata.invalid` → an unwired level, a bad owner id length, a missing owner, a bad draft. `metadatafields.in_use` → a field held at any level. No new code.

### FFI

`api/proto/engine.proto` per D-11, D-14, D-15. Handlers per D-12 – D-15 in `api/ffi/handlers/metadata.go` *(new)*, `artifacts.go` (`CreateArtifact`, `listArtifactsProto`), `evidence_graphs.go` (`CreateEvidenceGraph`, `UpdateEvidenceGraph`), `sources.go` (`GetSourceWorkspace` sets `ArtifactMetadata`; the four metadata handlers leave). `api/ffi/dispatch.go`: four renamed constants and cases, one new. Regenerate with `./scripts/generate-proto.sh` (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`) and commit `api/proto/engine/*.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`.

```proto
message MetadataOwner {
  string level = 1; // "source", "artifact" or "graph" (vocabkey.MetadataLevels); any other value is sourcemetadata.invalid
  string id = 2;
}
message MetadataDraft {
  repeated MetadataValue values = 1;     // one per field; the stored values become these
  repeated MetadataLayoutRow layout = 2; // written as given; rows not listed stay
}
message MetadataBoard {
  string owner_id = 1;
  repeated MetadataWorkspaceEntry entries = 2;
}
```

### macOS

None — no view changes. `GoStore.swift`, `CoreInvoke.swift` and `engine.pb.swift` change per D-16 so the app builds and the Source page calls the renamed RPCs.

### Strings

L10n keys *(new)*: `deleteImpact.via.artifactMetadataFieldId` = "%lld artifacts record a value for this field"; `deleteImpact.via.evidenceGraphMetadataFieldId` = "%lld Evidence graphs record a value for this field". Changed: `sourceTypes.toast.removedBodyWithValues` (D-17). Old-scope copy this PR checks: `deleteImpact.via.sourceMetadataFieldId` ("sources record a value") stays, it names its own group; `CatalogVocabularyRow.usedBy` doc and `MetadataField.used_by` comment change (D-17). Each in `L10n.swift` and `Localizable.xcstrings`.

## 7. Steps

**STEP-1 — The four tables, wired, with their deletes and scopes**
Files: `core/database/migrations/000065.sql` *(new)*, `core/database/catalogmodel/model.go`, `core/database/effects/registry.go`, `core/database/effects/path.go`, `core/database/effects/registry_completeness_test.go`, `core/database/effects/path_test.go`, `core/database/sourcemetadata/owner.go`, `core/database/sourcemetadata/sourcemetadata.go` (docs only), `core/database/sourcemetadata/sourcemetadata_test.go`, `core/database/deleteimpact/register.go`, `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/facets_test.go`, `core/database/audit/scopes_test.go`, `core/database/artifacts/artifacts_test.go`, `core/database/evidencegraphs/evidencegraphs_test.go`, `core/database/metadatafields/metadatafields_test.go`. Skill: `add-catalog-migration`, `add-catalog-model`, `add-audit-scope`, `add-catalog-delete`.
Do: D-1 – D-6, D-18. D-5's probes and layout releases land here because D-2 registers the FKs they cover (`TestResourceAndOwnedHonesty`, `TestFacetReleaseHonesty`). Update `TestOwnerLevels`: the refused owners are `citation`, `""`, `page` and the short id. Test fixtures build their Artifact, graph, type suggestion and field inline in each new case with the package's existing helpers. Add T-1, T-2, T-3, T-4, T-9, T-16.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes, including `TestPragmaHonesty`, `TestResourceAndOwnedHonesty`, `TestFacetReleaseHonesty`, `TestEveryNonSkipTableHasAnEntry`, `TestResolversCoverEveryEntityType`, `TestStoresAreRegistered`.

**STEP-2 — One graph removal; Used by from the delete totals**
Files: `core/database/deleteimpact/facets.go`, `facets_test.go`, `impact.go`, `impact_test.go`; `core/database/evidencegraphs/evidencegraphs.go`, `evidencegraphs_test.go`; `core/database/metadatafields/metadatafields.go`, `metadatafields_test.go`; `core/database/sourcetypes/sourcetypes.go`, `sourcetypes_test.go`; `api/ffi/handlers/source_defs_test.go`. Skill: `add-catalog-delete`.
Do: D-7, D-8. `TestResolversCoverEveryEntityType` keeps finding `evidence_graph` through `releaseGraph`'s `EntityType: "evidence_graph"` literal; it needs no change. Add T-5, T-10, T-17.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes, `TestDelete` (evidencegraphs) and `TestDeleteEvidenceGraph` with assertions unchanged; `rg -n "sqlUsedBy" core/database/metadatafields core/database/sourcetypes` prints nothing.

**STEP-3 — Many-owner reads, and Copy Source carries Artifact metadata**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `sourcemetadata_test.go`; `core/database/sourcecopy/sourcecopy.go`, `sourcecopy_test.go`. Skill: `add-catalog-query`.
Do: D-9, D-10. Add T-6, T-7, T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes, every existing `TestCopy` case included.

**STEP-4 — The board RPCs take an owner**
Files: `api/proto/engine.proto`, `api/proto/engine/*.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`, `api/ffi/handlers/metadata.go` *(new)*, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`, `api/ffi/handlers/source_defs_test.go`, `api/ffi/handlers/metadata_test.go` *(new)*, `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/GoStore.swift`; old names in guidance: `.cursor/rules/catalog-deletes.mdc` and `.agents/skills/add-catalog-delete/SKILL.md` (`ClearSourceMetadata` → `ClearMetadata`), `docs/audit-revision-history.md` (`update_source_metadata` → `update_metadata`), `docs/source-layer-data-model.md` § 5.4 (`dismiss_source_metadata_suggestion`, `reorder_source_metadata` → `dismiss_metadata_suggestion`, `reorder_metadata`); the action labels in `core/database/sourcemetadata/sourcemetadata_test.go` and `core/database/deleteimpact/facets_test.go` become the new names. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-11, D-12, D-13, D-16 (the `MetadataOwner` and `MetadataBoard` messages and the renames; `MetadataDraft` waits for STEP-5). `TestSetSourceMetadata` and `TestDismissAndReorderSourceMetadata` become `TestSetMetadata` and `TestDismissAndReorderMetadata`; their fixtures send a `source` owner and their assertions stay. Add T-11, T-12.
Check: `./scripts/generate-proto.sh` leaves no diff after commit; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `rg -n "SetSourceMetadata|ClearSourceMetadata|DismissSourceMetadataSuggestion|ReorderSourceMetadata|update_source_metadata|dismiss_source_metadata_suggestion|reorder_source_metadata" api core .agents .cursor docs/*.md macos/App/Platform/CoreInvoke.swift --glob '!*.pb.*'` prints nothing (`GoStore.swift`'s method names `setSourceMetadata` and the three others stay until S15-26).

**STEP-5 — Drafts on the Artifact and graph writes; Artifact boards on the Source page**
Files: `api/proto/engine.proto` and the generated files, `api/ffi/handlers/metadata.go` *(new)* (made in STEP-4), `artifacts.go`, `evidence_graphs.go`, `sources.go`, `artifacts_test.go`, `evidence_graphs_test.go`, `sources_test.go`. Skill: `add-ffi-handler`.
Do: D-14, D-15. Add T-13, T-14, T-15.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes, `TestEveryAuditedWriteReturnsEffects` included.

**STEP-6 — Copy for the new holders**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/CatalogVocabulary/VocabularyPane.swift`, `api/proto/engine.proto` (the `used_by` comment; regenerate). Skill: `add-localized-string`.
Do: D-17.
Check: `scripts/check-pr.sh` on macOS: `xcode-test` and `swift-warnings` pass.

**STEP-7 — Version and docs**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj`; `docs/source-layer-data-model.md` (§ 4.2: a copied Artifact carries its metadata; § 5: the three levels and the four tables with their DDL as § 5.3 / § 5.4 show the Source's; § 5.4: only the store table names the levels; § 7: an Artifact's metadata facets; § 7.1: a graph's metadata facets, and one graph removal (`ReleaseGraph`) for a graph delete and a Source delete; § 10 schema list); `docs/catalog-deletes.md` (Facet release table: the four tables and their FKs as Audited; the field's three value probes); `.agents/skills/add-audit-scope/SKILL.md` ("Write changes so they resolve": the second hop reads the ghost map, D-18); `docs/deployment-plan/spike-15/completed.md`; `docs/deployment-plan/spike-15/deployment-plan.md` (map row); this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-08-artifacts-and-graphs-hold-metadata.md` reports no error; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestOwnerTables` *(new)* | For an Artifact owner and a graph owner of one Source (its type with `document_number` appended at `artifact` and `graph` via `sourcevocab.AppendSuggestion`): `Set` → a row in that level's values table and a layout row, change entity type = the table, fields carry `artifact_id` / `graph_id`; `Clear`, `DismissSuggestion`, `Reorder`, `Apply`, `CreateAll`, `ReadDraftTx`, `ListByOwnersTx` each as for a Source; `ListWorkspace` lists only that level's suggestions (AC-1.2); `Set` on an Artifact id with no row → `ErrInvalid` | AC-1.1, AC-1.2, AC-1.5, D-4 |
| T-2 | Go | `sourcemetadata_test.go` `TestOwnerLevels` (updated), `TestStoresAreRegistered` (one new assertion) | `citation`, `""`, `page`, a 15-byte Artifact id → `ErrInvalid`, nothing written; every store's tables and owner FKs registered; the keys of `stores` equal `vocabkey.MetadataLevels()` (review M-5) | AC-1.4, D-2, D-4 |
| T-3 | Go | `core/database/audit/scopes_test.go` `TestSourceScopes` *(new cases, fixtures inline)* "artifact metadata set" and "graph metadata set" | The revision's `source` scope is the owner's Source | AC-1.3, D-3 |
| T-4 | Go | `core/database/metadatafields/metadatafields_test.go` (new delete cases); `core/database/deleteimpact/facets_test.go` `TestFacetReleaseAuditsSourceSide` (new case) | AC-2.1 (one group: via, kind, total 1, the listed graph); AC-2.3 (deletable, one `artifact_metadata_layout` delete change with `artifact_id` and `field_id`); AC-2.4 (`ErrInUse`); AC-2.2's three groups | AC-2.1, AC-2.3, AC-2.4, D-5 |
| T-5 | Go | `core/database/deleteimpact/facets_test.go` `TestFacetReleaseAuditsSourceSide` *(new cases)* "source delete audits its graph's metadata", "source delete releases two graphs in id order" | AC-3.3 (`evidence_graph_metadata` and `evidence_graph_metadata_layout` deletes before the `evidence_graph` delete; no row left); AC-3.4. The existing case's change list is unchanged | AC-3.3, AC-3.4, D-7 |
| T-6 | Go | `sourcemetadata_test.go` `TestListWorkspacesExported` *(new)* | Owners `[S, A1, A2, A1]` → four lists, index-aligned, A1's twice and equal, each equal to `ListWorkspace` alone; no owners → `nil`, no error; a graph owner with no row → `ErrInvalid` | AC-5.11, D-9 |
| T-7 | Go | `core/database/sourcecopy/sourcecopy_test.go` `TestCopy` *(new cases)* "artifact metadata travels with its Artifact" and "artifact metadata travels without the Metadata tick" | AC-4.1, AC-4.2, AC-4.3; the changes run the copied Artifact's create then its `artifact_metadata` and layout creates | AC-4.1 – AC-4.3, D-10 |
| T-8 | Go | `sourcemetadata_test.go` `TestReadDraftsTx` *(new)* | Owners `[A1, G1, A1]` → drafts index-aligned; no owners → `nil`; `nil` tx → `ErrInvalid`; a `citation` owner → `ErrInvalid` | D-9 |
| T-9 | Go | `core/database/artifacts/artifacts_test.go` `TestArtifactDelete` *(new case)*; `core/database/evidencegraphs/evidencegraphs_test.go` `TestDelete` *(new case)* | AC-3.1; AC-3.2 | AC-3.1, AC-3.2, D-6 |
| T-10 | Go | `api/ffi/handlers/source_defs_test.go` `TestDeleteMetadataField` *(new case)* "a field held only on a graph is in use" | `metadatafields.in_use`; `ListMetadataFields` shows `used_by` 1 for it | AC-2.1, AC-2.4, D-8 |
| T-11 | Go | `api/ffi/handlers/metadata_test.go` `TestMetadataOwners` *(new)* | `SetMetadata`, `ClearMetadata`, `DismissMetadataSuggestion`, `ReorderMetadata` for an Artifact owner and a graph owner (AC-5.1, AC-5.2); owner level `citation` / `""` / missing owner → `sourcemetadata.invalid`; a non-UUIDv7 owner id → `request.invalid_id` (AC-5.3) | AC-5.1 – AC-5.3, D-11, D-12 |
| T-12 | Go | `api/ffi/handlers/sources_test.go` `TestSetMetadata`, `TestDismissAndReorderMetadata` (renamed); `metadata_test.go` `TestGetMetadataBoard` *(new)* | Source-owner cases pass with assertions unchanged (AC-5.4); a graph's `board` has its id as `owner_id` and lists its `graph` suggestions and values (AC-5.5) | AC-5.4, AC-5.5, D-13 |
| T-13 | Go | `api/ffi/handlers/artifacts_test.go` `TestCreateArtifactMetadata` *(new)* | AC-5.6 (changes in order; the board via `GetSourceWorkspace`); AC-5.7 (no `metadata`: the changes equal `TestCreateArtifactAndIngest`'s); AC-5.8 (unknown field → `sourcemetadata.invalid`, no `artifacts` row added); a non-UUIDv7 draft field id → `request.invalid_id` | AC-5.6 – AC-5.8, D-14 |
| T-14 | Go | `api/ffi/handlers/evidence_graphs_test.go` `TestCreateEvidenceGraph`, `TestUpdateEvidenceGraph` *(new cases)* | AC-5.9; AC-5.10 (absent → unchanged; present empty → `line` deleted) | AC-5.9, AC-5.10, D-14 |
| T-15 | Go | `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace` *(new case)* "artifact boards" | AC-5.11 (`artifact_metadata` length 2, owner ids in `artifacts` order, A1's value present, A2's suggestions only) | AC-5.11, D-15 |
| T-16 | Go | `core/database/effects/path_test.go` `TestScopeThroughDeletedParent` *(new)* | AC-1.6: after `artifacts.Delete` of an Artifact with one value, `effects.Resolve` over the revision's changes gives each `artifact_metadata` and `artifact_metadata_layout` change the Source scope; a Citation delete's `citation_note` change resolves to its Source | AC-1.6, D-18 |
| T-17 | Go | `core/database/deleteimpact/impact_test.go` `TestInboundTotal` *(new)*; `metadatafields_test.go`, `core/database/sourcetypes/sourcetypes_test.go` (new cases) | AC-2.2 (`InboundTotal` = 4; `UsedBy` = 4; `List` row = 4); AC-2.5; `InboundTotal` with a nil tx, an unknown kind or a 15-byte id → `ErrInvalid` | AC-2.2, AC-2.5, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/deleteimpact` `TestFacetReleaseHonesty`, `TestResourceAndOwnedHonesty` and `TestFacetReleaseAuditsSourceSide` (change order), `catalogmodel` `TestPragmaHonesty` (indexes), `effects` completeness and resolve tests, `audit` `TestResolversCoverEveryEntityType`, `evidencegraphs` `TestDelete`, `api/ffi/handlers` `TestEveryAuditedWriteReturnsEffects`, `api/ffi` `TestSwiftVocabularyMatchesGo`, `sourcecopy` `TestCopy`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `SetSourceMetadata`, `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata`, `METHOD_SET_SOURCE_METADATA`, `update_source_metadata`, `dismiss_source_metadata_suggestion`, `reorder_source_metadata`, `metadatafields.sqlUsedBy`, `sourcetypes.sqlUsedBy` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (GoStore's Swift method names stay until S15-26); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-15/deployment-plan.md: map row status; docs/deployment-plan/spike-15/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated)
- [ ] Domain doc updated: source-layer-data-model.md §§ 4.2, 5, 7, 7.1, 10; catalog-deletes.md Facet release table; audit-revision-history.md action example; add-audit-scope "Write changes so they resolve"; add-catalog-delete and catalog-deletes.mdc name `ClearMetadata`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the dogfood project. On a birth certificate Source: fill a suggested field, add an extra field, dismiss a suggestion, drag to reorder, clear a value. Each behaves as before, and the Source's history lists each change (the renamed RPCs).
2. **Copy Source** with Metadata ticked and one Artifact chosen: the copy's metadata board matches the original's, and the copy has the Artifact.
3. On the Metadata page, a field used on a Source cannot be deleted, and its notice reads "1 sources record a value for this field" as today. Artifact and graph values have no chrome yet; T-10 and T-13 cover them.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Source delete's change order moves and history or a test reading it breaks | D-7 keeps the graph's own change and columns; T-5 pins the order; the existing case is unchanged |
| The proto renames break the Swift build where it cannot be run (Linux) | D-16 lists every Swift file; STEP-4's `rg` check; `check-pr.sh` on macOS runs `xcode-test` |
| `List` gets slower with per-field counts | 190 indexed counts in one read transaction (§6 **Scale**); the indexes are in D-1 |
| An Artifact-level write is not scoped to its Source and the Source page does not refresh | T-3 asserts the scope; T-11 asserts `source_scopes` on the reply |
| The ghost-lookup fix adds scopes to existing revisions' resolution | It only adds a scope a declared path already names (AC-1.6); the `effects` and `audit` suites run, and a changed existing case is a Stop condition |

**Stop and report if:**

- `sourcemetadata` needs a change beyond D-4 and D-9 for a level to work (S15-07's store table is then incomplete).
- `TestFacetReleaseHonesty` or `TestPragmaHonesty` cannot pass with the registrations of D-2, D-5, D-6, D-7.
- `TestFacetReleaseAuditsSourceSide`'s existing case needs an assertion changed.
- An existing handler test needs an assertion changed (not a call site or a renamed type).
- `artifacts` or `evidencegraphs` would have to import `sourcemetadata`, or `sourcemetadata` one of them.
- Fixing `readRowsColumn` (D-18) changes the resolved scopes of an existing `effects` or `audit` test case other than adding scopes the case did not have.
- `evidencegraphs.Delete` on `ReleaseGraph` changes `TestDelete`'s or `TestDeleteEvidenceGraph`'s change list.
- `./scripts/generate-proto.sh` cannot run (no `protoc` or its plugins) and generated files would have to be edited by hand: finish STEP-1 – STEP-3, commit, and report (header **Toolchain**).
- `TestEveryAuditedWriteReturnsEffects` names a request this plan did not change.

## 12. Open questions

None.

## 13. PR

**Title:** S15-08 — Artifacts and graphs hold metadata
**Body:** objective in one sentence; what shipped against US-1 – US-5; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-08-artifacts-and-graphs-hold-metadata.md`.
