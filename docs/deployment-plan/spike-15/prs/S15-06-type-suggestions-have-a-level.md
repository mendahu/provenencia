# S15-06 — Type suggestions have a level

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: Go, the wire and a store parameter; the Source page and the Source types page draw the same rows as today)
**Board:** —
**Depends on:** —
**Blocks:** S15-07, S15-08, S15-12, S15-14, S15-18, S15-19
**Branch:** `s15-06-type-suggestions-have-a-level`
**Version bump:** minor, 0.8.0 → **0.9.0** (if `main`'s `VERSION` is no longer 0.8.0 when you rebase, the next minor after it): `VERSION`, `core/version.go`, and all four `MARKETING_VERSION` lines in `macos/Provenencia.xcodeproj/project.pbxproj` — a format step that old apps refuse by `user_version`, and a new field on three requests ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/seeded-vocabulary.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "One level type", "Suggestions have a level", "Metadata lives at three levels with one vocabulary" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`evidence-explained-citation-map.md`](../evidence-explained-citation-map.md) § 6.6; [`source-layer-data-model.md`](../../../source-layer-data-model.md) § 5.2; [`seeded-vocabulary.md`](../../../seeded-vocabulary.md) § 2.3
**Planned:** 2026-11-11 against `main` @ `43b7247`
**Review:** [`S15-06-type-suggestions-have-a-level.review.md`](S15-06-type-suggestions-have-a-level.review.md) — round 1: Fits with changes / Ready with changes; round 2 running

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

A Source type's suggested fields gain a level: `source` (the record), `artifact` (the copy used) or `graph` (the item cited). This PR adds the one `Level` type the spike shares (those three, which hold metadata, plus `citation`, which only S15-14's built-ins use), rebuilds `source_type_metadata_fields` with `level` in its key, and makes every reader and writer of the join name a level. It is the foundation of slice 2: S15-08 suggests fields on Artifacts and graphs, S15-12 edits the three lists, S15-19 seeds them, and S15-14 reuses the type for its format prefixes.

**Why now.** Working rules "One level type" and "Suggestions have a level"; S15-07, S15-08, S15-12, S15-14, S15-18 and S15-19 all depend on it.

**What stays.** Every existing suggestion becomes a `source` row. The Source page's metadata board, the Source types page and its Fields count show exactly what they show today. No Artifact or graph values exist yet (S15-08), and no screen adds an `artifact` or `graph` row (S15-12).

## 2. User stories

**US-1.** As a researcher, when I open a project made before this PR, my Source types suggest the same fields in the same order on the Source page and on the Source types page, so that the upgrade changes nothing I see.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A catalog whose `birth_certificate` type suggests `document_number`, `record_date`, `issue_date` (sort 0, 1, 2) is opened after the migration | `ListSuggestions(c, typeID, LevelSource)` returns the three in that order; `artifact` and `graph` return none |
| AC-1.2 | A Source of a type that suggests field A at `source` and field B at `graph` | The Source page's workspace (`sourcemetadata.ListWorkspace`) lists A as a suggestion and not B |
| AC-1.3 | The Source types list for a type with 2 `source` rows and 1 `graph` row | `SuggestedFields` / `suggested_field_count` is 2 |
| AC-1.4 | The Source types page selects a type whose fake store holds one `source` and one `graph` suggestion | The page lists only the `source` one; assigning a field writes at `source` and leaves the `graph` row |

**US-2.** As the engine, I keep suggestions per level, so that S15-08 and S15-12 can suggest copy fields on Artifacts and item fields on graphs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `AssignTypeField` with `level: "graph"` for field F on type T, then `ListTypeSuggestions` with `level: "graph"` | Returns `[F]`; the `source` list of T is unchanged |
| AC-2.2 | F is assigned to T at `source` and at `graph` | Both rows exist; `ListTypeSuggestions` returns F at each level |
| AC-2.3 | T has `source` rows at sort 0 and 1; F is appended at `artifact` | F's sort order is 0 (order is per level) |
| AC-2.4 | `RemoveTypeField` with `level: "graph"` on F assigned at both levels | The `graph` row goes; the `source` row stays |
| AC-2.5 | Any of the three RPCs with `level` `""`, `"Source"`, `"page"` or `"citation"` | Refused with `sourcevocab.invalid`; nothing is written |
| AC-2.6 | `vocabkey.MetadataLevels()` and Swift `MetadataLevel.allCases` | The same three strings; `api/ffi/swift_vocabulary_test.go` fails if they drift |
| AC-2.7 | `sourcevocab.Install` on a new catalog | Every seeded suggestion is a `source` row, as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey.Level` with `LevelSource`, `LevelArtifact`, `LevelGraph`, `LevelCitation`, `Levels()`, `MetadataLevels()`, `Valid()`, `HoldsMetadata()` | STEP-1, AC-2.6 |
| Migration `000064.sql`: rebuild `source_type_metadata_fields` with `level` in the primary key; existing rows become `source` | STEP-2, AC-1.1 |
| `sourcevocab` suggestion functions and `seedSuggestion` take a level | STEP-3, AC-2.3, AC-2.7 |
| One suggested-field count in `sourcetypes` (`List` and the new `SuggestedFields`), counting `source` rows only; `sourcevocab.CountSuggestions` deleted | STEP-3, AC-1.3 |
| `sourcemetadata.ListWorkspace` reads `source` suggestions only | STEP-3, AC-1.2 |
| `effects` registry and `catalogmodel`: verified unchanged (D-7) | STEP-3 |
| `level` on `ListTypeSuggestionsRequest`, `AssignTypeFieldRequest`, `RemoveTypeFieldRequest`; handlers pass it through | STEP-4, AC-2.1 – AC-2.5 |
| Swift `MetadataLevel` enum and its parity row | STEP-5, AC-2.6 |
| Swift store methods take a level; the Source types page and the suggestions key pass `.source`; `FakeStore` keeps rows per level | STEP-6, AC-1.4 |
| Docs: domain DDL, seeded-vocabulary, `add-seeded-vocabulary` suggestion shape | STEP-7 |

| Out | Home |
| --- | --- |
| Artifact and graph metadata values, layout tables and their board | S15-08 |
| `sourcemetadata` functions on an owner | S15-07 |
| The Source types page editing three lists; a level in `CatalogQueryKey.typeSuggestions` and in `CatalogTypeSuggestion`; what the Fields column counts once three lists exist; renaming `FakeStore.suggestionsByType` call sites to `suggestionsByLevel` and dropping the alias | S15-12 |
| Using `LevelCitation` (prefixes, built-ins table, Insert field groups) | S15-14, S15-17 |
| Seeding `artifact` and `graph` suggestions | S15-19 |
| Copying a type's suggestions at every level | S15-18 |
| Format prefixes and the built-ins table using `Level` | S15-14 |
| Renaming `source_metadata_fields` to `metadata_fields` | not planned (map § 6.6 "Name") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Level type | `core/vocabkey/vocabkey.go` | No level type. Precedent for a small closed key set: `PrimaryKinds()` returns a new slice each call; `ValidOrigin` and `IsPrimaryKind` report membership as free functions on strings |
| Table | `core/database/migrations/000005.sql` | `source_type_metadata_fields (source_type_id, field_id, sort_order, PRIMARY KEY (source_type_id, field_id))`; both FKs `ON DELETE CASCADE` |
| Indexes | `core/database/migrations/000031.sql` | `source_type_metadata_fields_type_id_idx` (redundant with the PK's leftmost column) and `source_type_metadata_fields_field_id_idx`. A rebuild drops both |
| Latest migration | `core/database/migrations/000063.sql` | Last step; the next is `000064.sql` |
| Rebuild precedent | `core/database/migrations/000061.sql` | `-- migrate: foreign_keys off`, `CREATE TABLE …__new`, `INSERT … SELECT`, `DROP`, `ALTER TABLE … RENAME` |
| Suggestion ops | `core/database/sourcevocab/sourcevocab.go` `EnsureSuggestion`, `AppendSuggestion`, `CountSuggestions`, `ListSuggestions`, `DeleteSuggestion` and the `sql*Suggestion*` constants | Keyed by `(type, field)`; `sqlEnsureSuggestion` and `sqlAppendSuggestion` conflict on `(source_type_id, field_id)`; `sqlAppendSuggestion` takes `MAX(sort_order)` over the whole type |
| Seed registry | `core/database/sourcevocab/registry.go` `seedSuggestion`, `seedSuggestions` | `{TypeKey, FieldKey, SortOrder}` with hand-numbered orders; three `birth_certificate` rows (0, 1, 2); `Install` calls `EnsureSuggestion` for each |
| Type list count | `core/database/sourcetypes/sourcetypes.go` `sqlList`, `List`, `Type.SuggestedFields`, `UsedBy`; `core/database/sourcevocab/sourcevocab.go` `CountSuggestions`; `macos/App/Platform/FakeStore.swift` `withSuggestedFieldCount` | The same count lives twice in Go: `sqlList`'s subquery and `sourcevocab.CountSuggestions` (only caller: `UpdateSourceType`). Both count every join row of the type. The fake derives it from `suggestionsByType` |
| Source board | `core/database/sourcemetadata/sourcemetadata.go` `ListWorkspace` | Calls `sourcevocab.ListSuggestions(c, src.SourceTypeID)` |
| Effects | `core/database/effects/registry.go` | `"source_type_metadata_fields": {None: true}`: the join is never audited and no write through `writes.Run` touches it |
| Catalog model | `core/database/catalogmodel/model.go` | Table `{Name: "source_type_metadata_fields", Bucket: BucketFacet}` and two FKs to `source_types` and `source_metadata_fields`; no columns or key declared |
| Proto | `api/proto/engine.proto` `ListTypeSuggestionsRequest` (fields 1, 2), `AssignTypeFieldRequest` and `RemoveTypeFieldRequest` (fields 1, 3, 4; 2 reserved), `TypeSuggestion` | No level |
| Handlers | `api/ffi/handlers/source_defs.go` `ListTypeSuggestions`, `AssignTypeField`, `RemoveTypeField`, `parseSuggestionJoin`, `suggestionsProto`, `UpdateSourceType` | `suggestionsProto(c, typeID)` lists every row; `UpdateSourceType` fills `UsedBy` from `sourcetypes.UsedBy` and `SuggestedFields` from `sourcevocab.CountSuggestions(c, typeID)` |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `listTypeSuggestions`, `assignTypeField`, `removeTypeField`, `CatalogTypeSuggestion`; `macos/App/Platform/GoStore.swift` same three; `macos/App/Platform/FakeStore.swift` `suggestionsByType` and the same three; `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | No level |
| Swift callers | `macos/App/Features/SourceTypes/SourceTypesModel.swift` `assignPickedField`, `removeSuggestion` (patch `suggestedFieldCount = updated.count`); `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.typeSuggestions` loader | Call the store without a level |
| Parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | Row `"primary kinds"` pairs `vocabkey.PrimaryKinds()` with `EvidencePrimaryKind` through `rawCase` |
| Tests | `core/database/sourcevocab/sourcevocab_test.go` `TestInstall`, `TestAppendSuggestion`; `core/database/sourcetypes/sourcetypes_test.go` `TestDelete` (two raw `INSERT INTO source_type_metadata_fields (source_type_id, field_id, sort_order)`, one raw `SELECT COUNT(*)` on the join, and the `SuggestedFields` case; raw because `sourcevocab` imports `sourcetypes`); `core/database/metadatafields/metadatafields_test.go` `TestDelete` (one raw insert, one raw count); `core/database/sourcecopy/sourcecopy_test.go` (`EnsureSuggestion`); `api/ffi/handlers/source_defs_test.go` `TestAssignAndRemoveTypeField`, `TestListTypeSuggestions`, the `UpdateSourceType` case; `api/ffi/handlers/sources_test.go` (the metadata fixture calls `ListTypeSuggestions` without a level); `macos/ProvenenciaTests/SourceTypesModelTests.swift` | Cover today's single-level behavior. No request in `api/ffi` sets a level |

**Precedent.** `000061.sql` (S14-36) for the table rebuild: copy its shape and the `-- migrate: foreign_keys off` line. `vocabkey.PrimaryKinds()` and the `"primary kinds"` parity row for a closed Go key set mirrored by a Swift `rawCase` enum: copy both.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `type Level string` in `core/vocabkey` with `LevelSource = "source"`, `LevelArtifact = "artifact"`, `LevelGraph = "graph"`, `LevelCitation = "citation"`. `func Levels() []Level` returns all four, a new slice in that order; `func MetadataLevels() []Level` returns the first three, the levels that hold metadata. `func (l Level) Valid() bool` is membership in `Levels()`; `func (l Level) HoldsMetadata() bool` is membership in `MetadataLevels()`. Both are exact (no trimming, no case folding). Methods on the typed key, not free functions on strings: this is the form for new typed keys in `vocabkey` (`ValidOrigin` and `IsPrimaryKind` predate typed keys) | Working rule "One level type", amended by the user's answer (2026-11-11): one type serves suggestions, owners, format prefixes and the built-ins table, and the Citation level (built-ins only, map § 6.3) is in it; `vocabkey` is pure | Three values with `citation` added by S15-14 — S15-14 would rework validation, the `CHECK` and the Swift enum; a second type for prefixes — breaks the rule (user, 2026-11-11). A `Level` in `sourcevocab` — `core/citationformat` cannot import a database package |
| D-2 | The migration's primary key is `(source_type_id, field_id, level)`; `level TEXT NOT NULL CHECK (level IN ('source', 'artifact', 'graph'))` (the `MetadataLevels()`; a type never suggests at `citation`) with **no** `DEFAULT` | Map § 6.6 key. No default: every writer names its level, and a writer that forgets fails `NOT NULL` instead of silently writing `source` | `DEFAULT 'source'` from the map sketch — hides a missed level; the backfill names `'source'` explicitly instead |
| D-3 | The rebuild recreates `source_type_metadata_fields_field_id_idx` only. `source_type_id` is covered by the primary key's leftmost column | `000031`'s `type_id_idx` duplicated the PK; `TestPragmaHonesty` accepts a PK-leftmost covering index | Recreating both — keeps a redundant index |
| D-4 | `sort_order` is per (type, level): `sqlAppendSuggestion` takes `MAX(sort_order)` over `source_type_id = ? AND level = ?` | Each level is its own ordered list (S15-12 edits three) | One order across levels — the type page would show gaps |
| D-5 | Every `sourcevocab` suggestion function takes `level vocabkey.Level` after the ids: `EnsureSuggestion(c, typeID, fieldID, level, sortOrder)`, `AppendSuggestion(c, typeID, fieldID, level)`, `ListSuggestions(c, typeID, level)`, `DeleteSuggestion(c, typeID, fieldID, level)`. Each returns `ErrInvalid` when `!level.HoldsMetadata()` (so `citation` is refused like an unknown string, matching the `CHECK`), before any query | One validation, at the owner of the join; the handlers pass the wire string through as `vocabkey.Level(req.GetLevel())` | Validating in the handler — a second check S15-08's Go callers would skip |
| D-6 | `seedSuggestion` becomes `{TypeKey string; Level vocabkey.Level; FieldKeys []string}`: one entry per (type, level), its fields in order. The three existing rows become one entry `{TypeKey: "birth_certificate", Level: vocabkey.LevelSource, FieldKeys: []string{"document_number", "record_date", "issue_date"}}`. `Install` passes each field's index in `FieldKeys` as its sort order | S15-19 writes 26 types' suggestions per level in format order (map § 6.8); order by position cannot repeat or skip a number | Hand-numbered `SortOrder` per row — S15-19 would number hundreds of rows, restarting at 0 per level (D-4), untested; a separate list per level — three near-copies |
| D-7 | `effects` registry and `catalogmodel` do not change | The join is `None` in effects (never audited, per `docs/catalog-deletes.md`); `catalogmodel` registers tables and FKs, which the rebuild keeps. `catalogmodel` `TestPragmaHonesty` and `effects` `TestEveryNonSkipTableHasAnEntry` prove it | Adding a level to either — they hold no columns |
| D-8 | Each of the three RPCs takes `string level`: `ListTypeSuggestionsRequest.level = 3`, `AssignTypeFieldRequest.level = 5`, `RemoveTypeFieldRequest.level = 5`. Each reply lists that level's suggestions only. `TypeSuggestion` does not change | Map row "the suggestion RPCs take a level"; Swift's `suggestedFieldCount = updated.count` patch stays right because the reply is the `source` list | One reply with every level tagged — changes the count patch and the cached list shape, which is S15-12's call |
| D-9 | The suggested-field count lives once, in `sourcetypes`: one SQL fragment `sqlSuggestedFields` *(new)* counts the type's `source` rows (`level` bound to `vocabkey.LevelSource`); `sqlList` uses it as its subquery and `func SuggestedFields(c *database.Catalog, typeID []byte) (int, error)` *(new)*, beside `UsedBy`, runs it alone. `UpdateSourceType` fills `SuggestedFields` from `sourcetypes.SuggestedFields`. `sourcevocab.CountSuggestions` and `sqlCountSuggestions` are deleted | One place for the rule "which rows count", which S15-12 changes; `sourcetypes` already reads the join in `sqlList`, and `sourcevocab` imports `sourcetypes`, so the count cannot live in `sourcevocab`. Source level only: the Fields column keeps today's meaning | Leaving the count in two packages — the level rule can be changed in one and missed in the other; counting every level — changes a visible number in a PR that changes nothing visible |
| D-10 | Swift `enum MetadataLevel: String, CaseIterable, Sendable { case source, artifact, graph }` in `macos/App/Platform/MetadataLevel.swift` *(new)*, one `case` per line without raw values so `rawCase` reads it; its doc comment says it mirrors `vocabkey.MetadataLevels()`, the levels that hold metadata. Parity row `"metadata levels"` against `vocabkey.MetadataLevels()`. No Swift case for `citation` until a Swift caller needs one (S15-17) | Working rule "One level type" (one Swift enum, one parity row); precedent `EvidencePrimaryKind` | Raw strings at call sites — no parity check |
| D-11 | `GenealogyStore` `listTypeSuggestions`, `assignTypeField`, `removeTypeField` gain `level: MetadataLevel` as the last parameter. `CatalogQueryKey.typeSuggestions` and the `.assignedTypeSuggestion` / `.removedTypeSuggestion` mutations stay as they are and mean the `source` list: the loader in `CatalogQueryRegistry` and `SourceTypesModel` pass `.source`. The key's doc comment says so | The type page shows one list until S15-12; S15-12 adds the level to the key with the three lists it draws | Adding the level to the key now — eight test sites and the mutation invalidations change for a list nobody shows yet |
| D-12 | `FakeStore` stores `suggestionsByLevel: [MetadataLevel: [String: [CatalogTypeSuggestion]]]`; `suggestionsByType` becomes a computed get/set over `suggestionsByLevel[.source, default: [:]]`, so existing tests stay unchanged. The computed property takes no lock: like every stored property it is read and written inside `withState` (`stateLock` is not reentrant). `deleteSourceType` clears the type at every level; `withSuggestedFieldCount` keeps reading the `.source` list (D-9) | One store per level, faithful to Go; no test churn | A second dictionary for other levels — two sources of truth |
| D-13 | No backfill test for the migration; behavior is tested in `sourcevocab`, `sourcetypes`, `sourcemetadata` and the handlers | `add-catalog-migration` step 4 | A per-migration test — the skill forbids it |

**Rules this plan bends.** The working rule "One level type" named three levels → `citation` joins the type so S15-14's prefixes and built-ins table use it (review M-1) → user, 2026-11-11: four values, with a metadata subset → the map's rule is amended in the commit that wires this plan, and STEP-7 has nothing left to do for it. D-2 drops the map sketch's `DEFAULT 'source'`; no working rule names a default, and STEP-7 updates § 6.6's sketch to match.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `vocabkey.Level`, `LevelSource`, `LevelArtifact`, `LevelGraph`, `LevelCitation`, `Levels`, `MetadataLevels`, `Level.Valid`, `Level.HoldsMetadata` *(new)* | type / const / func | `vocabkey.PrimaryKinds`, `vocabkey.ValidOrigin` | Extends the same package with the same shape; no level type exists (searched `rg -n "Level\b" core/`: only unrelated log and fold levels) |
| `core/database/migrations/000064.sql` *(new)* | migration | `000061.sql` | A new step; shipped steps are never edited |
| `seedSuggestion.Level`, `seedSuggestion.FieldKeys` *(new)* | field | `seedSuggestion` | Reshapes the existing row (D-6) |
| `sourcetypes.SuggestedFields`, `sqlSuggestedFields` *(new)* | func / SQL | `sqlList`'s subquery, `sourcevocab.CountSuggestions`, `sourcetypes.UsedBy` | Replaces the two copies with one, beside `UsedBy` (D-9) |
| `level` on three requests *(new)* | proto field | `ListTypeSuggestionsRequest`, `AssignTypeFieldRequest`, `RemoveTypeFieldRequest` | Extends; no new method |
| `macos/App/Platform/MetadataLevel.swift` `MetadataLevel` *(new)* | Swift enum | `EvidencePrimaryKind`, `CatalogSearchKind` | No Swift level type exists (`rg -n "enum .*Level" macos/App`) |
| `FakeStore.suggestionsByLevel` *(new)* | test store | `FakeStore.suggestionsByType` | Becomes the storage under the existing name |
| Parity row `"metadata levels"` *(new)* | test row | `"primary kinds"` | A row in the existing table |

**Builds on.** `vocabkey` as the shared key package; the `000061` rebuild pattern; `TestSwiftVocabularyMatchesGo`; the `sourcevocab` registry and `Install` (`add-seeded-vocabulary`).
**Deletes.** The index `source_type_metadata_fields_type_id_idx` (D-3). `sourcevocab.CountSuggestions` and `sqlCountSuggestions` (D-9); `seedSuggestion.SortOrder` (D-6).
**Next adds.** S15-07 and S15-08 key owners by `vocabkey.Level` and call `ListSuggestions(c, typeID, LevelArtifact | LevelGraph)`; S15-12 adds `level` to `CatalogQueryKey.typeSuggestions` and passes `.artifact` / `.graph` through the store parameter added here; S15-19 adds registry rows with `Level`; S15-14 maps `{source.…}`, `{artifact.…}`, `{graph.…}`, `{citation.…}` prefixes to `vocabkey.Level` and gives each built-in a level; the list form's rule is "Source only" over the same type.

### Reach

None — no new path into existing code. No screen writes an `artifact` or `graph` row in this PR; such a row is reachable only through the RPCs, and every reader filters by level.

### Scale

At the map's Scale (26 types, 63 fields; S15-19's seed is at most a few hundred join rows): every query is per type and level on a table of at most a few hundred rows, served by the primary key. The migration copies every existing row once. No list or loop changes its per-row work.

### Edges

- **Assumptions.** `ListWorkspace` assumes each suggested field appears once (its `seen` map). With levels, a field at `source` and `graph` is two rows, but `ListWorkspace` reads one level, so each field appears once. T-5 holds it.
- **Neighbors.** A field at two levels keeps both rows (T-3, AC-2.2); removing it at one level leaves the other (T-3, AC-2.4); appending at a new level starts its order at 0 (T-2, AC-2.3); re-appending at the same level does not move it (T-2, existing case).
- **Removed inputs.** A field or type delete cascades every level's rows, as today (`metadatafields_test.go` `TestDelete` "unused field with suggestion and layout still erases", updated to insert with a level).
- **Real data.** The shipped registry has three suggestions, all `source`; `TestInstall` checks them (T-4).
- **Untrusted size.** None — the level is one string checked against three values before any query.

### Schema

Migration `core/database/migrations/000064.sql` *(new)*:

```sql
-- S15-06: a Source type suggests fields per level (the Source, an Artifact,
-- an Evidence graph); the level is part of the key, so a field may be
-- suggested at more than one level. Existing rows become 'source'.
-- migrate: foreign_keys off

CREATE TABLE source_type_metadata_fields__new (
	source_type_id  BLOB NOT NULL REFERENCES source_types(id) ON DELETE CASCADE,
	field_id        BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
	level           TEXT NOT NULL CHECK (level IN ('source', 'artifact', 'graph')),
	sort_order      INTEGER,

	PRIMARY KEY (source_type_id, field_id, level)
) STRICT;

INSERT INTO source_type_metadata_fields__new (source_type_id, field_id, level, sort_order)
SELECT source_type_id, field_id, 'source', sort_order FROM source_type_metadata_fields;

DROP TABLE source_type_metadata_fields;
ALTER TABLE source_type_metadata_fields__new RENAME TO source_type_metadata_fields;

CREATE INDEX source_type_metadata_fields_field_id_idx ON source_type_metadata_fields(field_id);
```

Registered in `catalogmodel` as today (no change, D-7). Delete class: facet join, cascades with either parent. Audit: none (`None: true`).

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | D-1. Doc comment: the layers a citation reads: the Source, one of its Artifacts, one of its Evidence graphs, the Citation; the first three hold metadata values and type suggestions, the Citation only built-ins |
| `core/database/sourcevocab` | The five functions of D-5. SQL: `sqlEnsureSuggestion` inserts `level` and conflicts on `(source_type_id, field_id, level)`; `sqlAppendSuggestion` inserts `level`, takes `MAX(sort_order)` `WHERE source_type_id = ? AND level = ?`, conflicts on the three columns; `sqlListSuggestions` and `sqlDeleteSuggestion` add `AND j.level = ?` / `AND level = ?`. `CountSuggestions` and `sqlCountSuggestions` are deleted (D-9). `seedSuggestion` and `Install` per D-6. Package doc: "type↔field suggestions per level" |
| `core/database/sourcetypes` | `sqlSuggestedFields` and `SuggestedFields` (D-9); `sqlList` embeds the fragment and `List` binds `string(vocabkey.LevelSource)`. `Type.SuggestedFields` comment: "how many fields this type suggests at the Source level" |
| `core/database/sourcemetadata` | `ListWorkspace` calls `sourcevocab.ListSuggestions(c, src.SourceTypeID, vocabkey.LevelSource)` |

Errors: `sourcevocab.invalid` → a level outside `MetadataLevels()`, `citation` included, as for a bad id today.

### FFI

`api/proto/engine.proto`: `string level = 3;` on `ListTypeSuggestionsRequest`; `string level = 5;` on `AssignTypeFieldRequest` and `RemoveTypeFieldRequest`, each commented "`source`, `artifact` or `graph` (`vocabkey.Levels`); any other value is `sourcevocab.invalid`". The three message comments say the reply lists that level's suggestions. Regenerate with `./scripts/generate-proto.sh`.

Handlers in `api/ffi/handlers/source_defs.go`: `suggestionsProto(c, typeID, level vocabkey.Level)`; `ListTypeSuggestions`, `AssignTypeField`, `RemoveTypeField` pass `vocabkey.Level(req.GetLevel())` to `sourcevocab` and to `suggestionsProto`; `UpdateSourceType` calls `sourcetypes.SuggestedFields(c, typeID)` (D-9). `parseSuggestionJoin` does not change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Level enum | `MetadataLevel` *(new)* | `macos/App/Platform/MetadataLevel.swift` | D-10 |
| Store protocol | `GenealogyStore` `listTypeSuggestions(projectDir:typeID:level:)`, `assignTypeField(projectDir:typeID:fieldID:level:)`, `removeTypeField(projectDir:typeID:fieldID:level:)` | `macos/App/Platform/GenealogyStore.swift` | D-11 |
| Go store | `GoStore` same three set `req.level = level.rawValue` | `macos/App/Platform/GoStore.swift` | |
| Fake store | `FakeStore` same three read and write `suggestionsByLevel[level]` | `macos/App/Platform/FakeStore.swift` | D-12 |
| Test stub | `ThrowingStore` same three signatures | `macos/ProvenenciaTests/OnboardingModelTests.swift` | |
| Callers | `SourceTypesModel.assignPickedField`, `removeSuggestion` pass `level: .source`; the `.typeSuggestions` loader passes `level: .source` | `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` | D-11 |

States: unchanged. Accessibility: unchanged.

### Strings

None — no user-facing copy changes.

## 7. Steps

**STEP-1 — The level type**
Files: `core/vocabkey/vocabkey.go`, `core/vocabkey/vocabkey_test.go`.
Do: add D-1 below the origin constants.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/vocabkey/...` passes with T-1.

**STEP-2 — The migration**
Files: `core/database/migrations/000064.sql` *(new)*. Skill: `add-catalog-migration`, `add-catalog-model`.
Do: write the Schema sketch exactly.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/catalogmodel/...` passes (`TestPragmaHonesty`). Other `core/database` packages fail to compile or insert until STEP-3; that is expected.

**STEP-3 — Readers and writers take a level**
Files: `core/database/sourcevocab/sourcevocab.go`, `core/database/sourcevocab/registry.go`, `core/database/sourcevocab/sourcevocab_test.go`, `core/database/sourcetypes/sourcetypes.go`, `core/database/sourcetypes/sourcetypes_test.go`, `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`, `core/database/metadatafields/metadatafields_test.go`, `core/database/sourcecopy/sourcecopy_test.go`. Skill: `add-catalog-query`, `add-seeded-vocabulary`.
Do: D-4, D-5, D-6, D-9 and the `ListWorkspace` change. The raw inserts in `sourcetypes_test.go` (two) and `metadatafields_test.go` (one) add the `level` column with `'source'`; the raw counts there need no change. `sourcecopy_test.go` passes `vocabkey.LevelSource`. Confirm `core/database/effects/registry.go` and `core/database/catalogmodel/model.go` need no edit (D-7).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes with T-2 – T-6.

**STEP-4 — The RPCs take a level**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/source_defs.go`, `api/ffi/handlers/source_defs_test.go`, `api/ffi/handlers/sources_test.go`. Skill: `add-ffi-handler`.
Do: the FFI section. Every `ListTypeSuggestionsRequest`, `AssignTypeFieldRequest` and `RemoveTypeFieldRequest` literal under `api/ffi` (`rg -n "TypeSuggestionsRequest\{|TypeFieldRequest\{" api/ffi`) sets `Level: "source"`, except the new invalid-level cases.
Check: `./scripts/generate-proto.sh` then `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes with T-7, T-8.

**STEP-5 — The Swift enum and its parity row**
Files: `macos/App/Platform/MetadataLevel.swift` *(new)*, `api/ffi/swift_vocabulary_test.go`.
Do: D-10. Add the path constant `swiftMetadataLevel` and the row `{name: "metadata levels", path: swiftMetadataLevel, enum: "MetadataLevel", re: rawCase, want: sorted(levelStrings()...)}`, with a helper that maps `vocabkey.MetadataLevels()` to strings. The Xcode project uses synchronized folders, so the new file needs no project edit.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes (T-9).

**STEP-6 — Swift store calls name the level**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` (doc comment on `typeSuggestions` only), `macos/ProvenenciaTests/SourceTypesModelTests.swift`. Skill: `add-swift-test`.
Do: D-11, D-12; the protocol doc comments say the level.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes with T-10, T-11 and no Swift warning.

**STEP-7 — Docs and version**
Files: `docs/source-layer-data-model.md` § 5.2 (DDL with `level`, one sentence on levels), `docs/seeded-vocabulary.md` § 2.3 (one sentence: rows carry a level; the seed's rows are `source`), `.agents/skills/add-seeded-vocabulary/SKILL.md` (`seedSuggestions` — one entry per (type, level) with its field keys in order; the order is the position), `docs/deployment-plan/spike-15/evidence-explained-citation-map.md` § 6.6 sketch (drop `DEFAULT 'source'`, D-2), `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj`, `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-06-type-suggestions-have-a-level.md` and `scripts/check-pr.sh` print no error.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/vocabkey/vocabkey_test.go` `TestLevels` *(new)* | `Levels()` → `[source artifact graph citation]` and `MetadataLevels()` → `[source artifact graph]`, each a fresh slice per call; table over `source`, `artifact`, `graph`, `citation`, `""`, `Source`, ` source`, `page` → `Valid` true for the first four only, `HoldsMetadata` true for the first three only | AC-2.5, AC-2.6, D-1 |
| T-2 | Go | `core/database/sourcevocab/sourcevocab_test.go` `TestAppendSuggestion` (new cases) | Type with Publisher, Author at `source`; append Author at `artifact` → `artifact` list `[Author]` sort 0, `source` list unchanged; `AppendSuggestion` with level `page` or `citation` → `ErrInvalid` | AC-2.3, AC-2.5, D-4 |
| T-3 | Go | `core/database/sourcevocab/sourcevocab_test.go` `TestAppendSuggestion` (new case) | Author at `source` and `graph`; `DeleteSuggestion` at `graph` → `graph` empty, `source` `[Author]` | AC-2.2, AC-2.4 |
| T-4 | Go | `core/database/sourcevocab/sourcevocab_test.go` `TestInstall` | After Install, `ListSuggestions(cert, LevelSource)` returns `document_number`, `record_date`, `issue_date` with sort orders 0, 1, 2; `LevelArtifact` and `LevelGraph` return none | AC-1.1, AC-2.7, D-6 |
| T-5 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestSourceMetadata` (new case) | Source of a type with A at `source`, B at `graph` → `ListWorkspace` has A suggested, no B | AC-1.2 |
| T-6 | Go | `core/database/sourcetypes/sourcetypes_test.go` `TestDelete` (the `SuggestedFields` case) | `book`'s field inserted by raw SQL at `source` and the same field again at `graph` (raw: `sourcevocab` imports `sourcetypes`) → `List` gives `SuggestedFields` 1 and `SuggestedFields(c, book)` returns 1 | AC-1.3, D-9 |
| T-7 | Go | `api/ffi/handlers/source_defs_test.go` `TestAssignAndRemoveTypeField` (new case) | Assign F at `graph`, then at `source`; list `graph` → `[F]`; list `source` → `[F]`; remove at `graph` → reply `[]`, list `source` → `[F]` | AC-2.1, AC-2.2, AC-2.4 |
| T-8 | Go | `api/ffi/handlers/source_defs_test.go` `TestListTypeSuggestions`, `TestAssignAndRemoveTypeField` (new cases) | `level` `""`, `"Source"`, `"page"`, `"citation"` on each RPC → error `sourcevocab.invalid`; no row written | AC-2.5 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row `"metadata levels"` *(new)* | `MetadataLevel` cases = `vocabkey.MetadataLevels()` | AC-2.6, D-10 |
| T-10 | Swift | `macos/ProvenenciaTests/SourceTypesModelTests.swift` `selectingShowsOnlySourceLevelSuggestions` *(new)* | Fake holds A at `.source`, B at `.graph` for `t1`; select `t1` → suggestions `[A]` | AC-1.4, D-11 |
| T-11 | Swift | `macos/ProvenenciaTests/SourceTypesModelTests.swift` `assigningWritesAtSourceLevelOnly` *(new)* | Fake holds B at `.graph`; assign A → `suggestionsByLevel[.source]["t1"]` `[A]`, `[.graph]` still `[B]`, `suggestedFieldCount` 1 | AC-1.4, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/catalogmodel` `TestPragmaHonesty`, the schema hash self-check, `core/database/metadatafields` `TestDelete`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` and `MetadataModelTests.swift` (they set `suggestionsByType`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-15/deployment-plan.md: map row status; docs/deployment-plan/spike-15/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: source-layer-data-model.md § 5.2, seeded-vocabulary.md § 2.3; add-seeded-vocabulary suggestion shape; map § 6.6 sketch (D-2)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project made before this PR (the dogfood project). Open a birth certificate Source: the metadata board shows the same suggested fields in the same order as before.
2. Open **Source types**, select Birth certificate: the suggestions list and the Fields column show the same fields and count as before. Assign a field and remove it: both work as before.
3. Create a new project: the Birth certificate type suggests its three seeded fields.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rebuild loses rows or breaks a foreign key | `-- migrate: foreign_keys off` runs `foreign_key_check` before commit; T-4 and manual check 1 read the rows after upgrade |
| A test fixture inserts into the join without `level` and fails `NOT NULL` | STEP-3 lists every raw insert found by `rg -n "INTO source_type_metadata_fields"`; rerun that search before STEP-3's check |
| `FakeStore`'s computed `suggestionsByType` hides a level bug in Swift tests | T-10 and T-11 drive the model through other levels |

**Stop and report if:**

- `rg -n "source_type_metadata_fields" --glob '*.go' --glob '*.swift' --glob '*.sql'` finds a reader or writer not listed in § 4 (shipped migrations and the generated `engine.pb.*` comments excepted).
- A Go caller of `sourcevocab.CountSuggestions` other than `UpdateSourceType` exists.
- `TestPragmaHonesty` asks for a `catalogmodel` change (D-7 is wrong).
- `ON CONFLICT(source_type_id, field_id, level)` is refused by SQLite as matching no unique index.
- A Swift test other than those in § 8 needs an assertion changed (not a signature) to pass.
- `rawCase` does not read `MetadataLevel` as written in D-10.

## 12. Open questions

None.

## 13. PR

**Title:** S15-06 — Type suggestions have a level
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-06-type-suggestions-have-a-level.md`.
