# S15-07 — Metadata operations take an owner

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a Go refactor; the wire and every view stay as they are)
**Board:** —
**Depends on:** S15-06 (landed [#782](https://github.com/mendahu/provenencia/pull/782))
**Blocks:** S15-08
**Branch:** `s15-07-metadata-operations-take-an-owner`
**Version bump:** none — a preparatory refactor with no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** working rules "Metadata lives at three levels with one vocabulary", "One level type", "Metadata saves with its owner's write", "Citations render on demand, from SQL" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`evidence-explained-citation-map.md`](../evidence-explained-citation-map.md) § 6.6; [`source-layer-data-model.md`](../../../source-layer-data-model.md) §§ 5.3, 5.4; [`S15-32`](S15-32-copy-source-date.md) Next adds (the `sourcecopy` overlap)
**Planned:** 2026-11-11 against `main` @ `bdef66e`
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

Every exported function of `core/database/sourcemetadata` takes an **owner** (a `vocabkey.Level` and an id) instead of a Source id. Only the Source level is wired: its tables, audit shape and behavior are exactly today's. The package also gains the two pieces later PRs build on: one batch read of values by owners (S15-16's citation resolver), and `Apply`, which saves a form's draft by comparing it with what is stored (S15-08's Artifact and graph writes), with `CreateAll` as its empty-owner case. `ListWorkspace` stops reading one field per extra value. S15-08 then adds the Artifact and graph levels as two entries in one table, not a second copy of the logic.

**Why now.** Working rules "Metadata lives at three levels with one vocabulary" and "Metadata saves with its owner's write"; S15-08 depends on it, and the map review (K-9) moved every function and the batch read here so S15-08 and S15-16 add no logic of their own.

**What stays.** The FFI wire, every handler's reply, the Source page, Copy Source, the audit rows (entity types, fields, scopes) and the catalog schema. Every existing metadata test keeps its assertions; only call sites change, except the cases §6 **Deletes** names.

## 2. User stories

**US-1.** As a researcher, when I edit, dismiss, reorder or copy a Source's metadata after this PR, everything behaves and is audited as before, so that the refactor changes nothing I see.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The existing tests of `sourcemetadata`, `sourcecopy`, `deleteimpact` and `api/ffi/handlers` | Pass with call-site changes only (§6 **Deletes** lists the two cases whose state can no longer be expressed) |
| AC-1.2 | A Source with values for `document_number` (suggested) and three extra fields `a`, `b`, `c`; `ListWorkspace` | Lists the suggestions in type order, then `a`, `b`, `c` in field-id order, each with its field's key and label, as today |
| AC-1.3 | `ListWorkspace` on a Source id with no row | `sources.ErrNotFound`, as today |

**US-2.** As the engine, I address metadata by owner, so that S15-08 adds the Artifact and graph levels as two table entries.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Any exported function with an owner at `artifact`, `graph`, `citation`, `""` or `page`, or a `source` owner whose id is not 16 bytes | `ErrInvalid`; nothing is written |
| AC-2.2 | `Set` on a `source` owner | The change is `source_metadata` with a `source_id` field, as today; the audit scope is that Source |

**US-3.** As the engine, I read the values of many owners at once, so that S15-16 renders 300 Citations in a few queries.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Sources A (fields F1, F2) and B (F1) and C (none); `ListByOwnersTx(tx, [B, A, C, A])` | Three rows, A's then B's when A's id sorts first, each owner's rows by field id, each row carrying its `Owner` |
| AC-3.2 | `ListByOwnersTx(tx, nil)` | `nil`, no error, no query |
| AC-3.3 | The owners include one at `graph` | `ErrInvalid` before any query |

**US-4.** As the engine, I save a draft of values and layout in one step, so that S15-08 writes an Artifact's or a graph's metadata in the same transaction as the Artifact or graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | An owner with nothing stored; `Apply` with values `[document_number "No. 7"]` and layout `[link 0, document_number 1 dismissed]` | The same rows and the same changes, in the same order, as `CreateAll` writes today: one `source_metadata` create, then two `source_metadata_layout` creates |
| AC-4.2 | Stored `document_number "No. 7"`; draft value `"  No. 7  "` | No change, no write |
| AC-4.3 | Stored URL `https://example.org/a`; draft `HTTPS://Example.org/a` | No change |
| AC-4.4 | Stored values A `"x"` and B `"y"`; draft values `[A "z"]` | Changes: B's `source_metadata` delete (as `Clear` writes it), then A's update (as `Set` writes it); B's layout row stays |
| AC-4.5 | Stored layout A 0, B 1; draft layout `[B 0, A 1]` | Two `source_metadata_layout` updates with `sort_order` diffs only; `dismissed` untouched |
| AC-4.6 | Stored layout A 0 not dismissed; draft layout `[A 0 dismissed]` | One update with a `dismissed` diff only |
| AC-4.7 | Stored layout rows A and C; draft layout `[A …]` | C's row stays as it is |
| AC-4.8 | A draft repeating a field in its values, or in its layout; an empty value; an unknown field; an owner with no row | `ErrInvalid`; nothing is written |
| AC-4.9 | `CreateAll` on an owner that already holds a value or a layout row | `ErrInvalid`, as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `Owner` and the per-level store table, with only `vocabkey.LevelSource` wired | STEP-1, AC-2.1 |
| `Set`, `Clear`, `DismissSuggestion`, `Reorder`, `ListWorkspace`, `List` (was `ListBySource`), `CreateAll` take an owner; `Input` and `Row` lose `SourceID` | STEP-2, AC-1.1, AC-2.2 |
| `ReadDraftTx` replaces `ListBySourceTx` and `ListLayoutTx`; `sourcecopy` reads and writes metadata as one `sourcemetadata.Draft` | STEP-2, STEP-4, AC-1.1 |
| `ListByOwnersTx` (the batch read) | STEP-3, AC-3.1 – AC-3.3 |
| `Apply`; `CreateAll` refuses a non-empty owner and then calls it | STEP-3, AC-4.1 – AC-4.9 |
| `ListWorkspace` reads the extra values' fields in one query | STEP-2, AC-1.2 |
| Handlers and every Go caller pass a `source` owner | STEP-4, AC-1.1 |
| Docs: `source-layer-data-model.md` § 5.3 DDL and § 5.4, `add-catalog-query` example | STEP-5 |

| Out | Home |
| --- | --- |
| The Artifact and graph tables, their store entries, audit, delete, Copy Source of Artifact values, FFI | S15-08 |
| A caller of `ListByOwnersTx` | S15-16 |
| A caller of `Apply` other than `CreateAll` | S15-08 (Artifact and graph writes) |
| `ListWorkspace` reading in one read transaction | not planned — every caller holds the exclusive catalog session (`withProjectCatalog`), so no write interleaves |
| Renaming the package `sourcemetadata` or the table `source_metadata_fields` | not planned (map § 6.6 "Name") |
| Search over Artifact or graph values | not planned (S15-08 "Not this PR") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Package | `core/database/sourcemetadata/sourcemetadata.go` | 700 lines; every SQL constant names `source_metadata` / `source_metadata_layout` / `source_id` directly |
| Types | same, `Row{ID, SourceID, FieldID, ValueText}`, `Input{SourceID, FieldID, ValueText}`, `WorkspaceEntry`, `LayoutRow{FieldID, SortOrder, Dismissed}` | Source-keyed |
| Writes | same, `Set(tx, userID, in Input)`, `Clear(tx, userID, sourceID, fieldID)`, `DismissSuggestion(tx, userID, sourceID, fieldID)`, `Reorder(tx, userID, sourceID, fieldIDs)`, `CreateAll(tx, userID, sourceID, values []Input, layout []LayoutRow)` | `Set` inserts through `insertValueTx` and appends a layout row (`sqlLayoutAppend`); updates inline with `sqlUpdate`. `Clear` deletes inline. `CreateAll` returns `nil, nil` on empty input before any check, refuses a Source with any row (`sqlHasAny`) or an `Input.SourceID` that differs from `sourceID`, then inserts values (`insertValueTx`) and layout rows verbatim (`sqlLayoutInsert`): changes are values in input order, then layout in input order. A missing Source is `ErrInvalid` (`requireSource`) |
| Reads | same, `ListBySource(c, sourceID)`, `ListBySourceTx(tx, sourceID)`, `ListLayoutTx(tx, sourceID)`, `ListWorkspace(c, sourceID)` | `ListWorkspace` calls `sources.Get` (a missing Source is `sources.ErrNotFound`, commit `c3b78f6`), `sourcevocab.ListSuggestions(c, typeID, vocabkey.LevelSource)`, then `getField(db, v.FieldID)` **once per extra value** |
| Change shapes | same | Values: `EntityType "source_metadata"`, `EntityID` the row id, fields `id`, `source_id`, `field_id`, `value_text`. Layout: `EntityType "source_metadata_layout"`, `EntityID` the field id, fields `source_id`, `field_id` and the changed `sort_order` / `dismissed` |
| Schema | `core/database/migrations/000009.sql`, `000031.sql` | `source_metadata` has `UNIQUE (source_id, field_id)` (the domain doc's DDL omits it); layout PK `(source_id, field_id)` |
| Effects | `core/database/effects/registry.go` `"source_metadata"`, `"source_metadata_layout"` | `SourceScope: field("source_id")`; unchanged by this PR |
| Copy Source | `core/database/sourcecopy/sourcecopy.go` `Draft.Metadata []sourcemetadata.Input`, `Draft.Layout []sourcemetadata.LayoutRow`, `ReadDraft`, `CreateFromDraft`, `apply` | `ReadDraft` calls `ListBySourceTx` and `ListLayoutTx`; `CreateFromDraft` sets each `Input.SourceID` and calls `CreateAll`; `apply` nils both fields when Metadata is off |
| Handlers | `api/ffi/handlers/sources.go` `GetSourceWorkspace`, `SetSourceMetadata`, `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata` | Pass `sourceID` to seven `sourcemetadata` calls |
| Batch helpers | `core/database/sqlin.go` `UniqueBlobIDs`, `SQLInPlaceholders`, `BlobArgs` | Used by `autoreconciler`, `searchindex` for `IN (…)` reads |
| Tests | `core/database/sourcemetadata/sourcemetadata_test.go` `TestSourceMetadata`, `TestCreateAll` and helpers `runSet`, `runClear`, `runDismissSuggestion`, `runReorder`, `runCreateAll`; `core/database/sourcecopy/sourcecopy_test.go` (`ListWorkspace`, `ListLayoutTx` in `layoutOf`, `ListBySource`, `Set`, `DismissSuggestion`, `Reorder`); `core/database/deleteimpact/facets_test.go` `runMetadataSet`, `runDismissMetadata`; `api/ffi/handlers/sources_test.go` metadata cases | Source-keyed call sites |
| Skill | `.agents/skills/add-catalog-query/SKILL.md` "Reads inside a write" | Names `sourcemetadata.ListLayoutTx` as an example Tx twin |

**Precedent.** `core/database/deleteimpact/facets.go` `rowFacet` builds its SQL once with `fmt.Sprintf` from constant table and column names, one descriptor per facet table: copy that for the per-level store. `core/database/autoreconciler/autoreconciler.go` (`sqlLoadCandidates` + `database.SQLInPlaceholders` + `database.BlobArgs`) for the `IN (…)` batch read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `type Owner struct { Level vocabkey.Level; ID []byte }` in `sourcemetadata`. Callers write the literal `sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: sourceID}`; no constructor | Rule "One level type" (owners use `vocabkey.Level`). A literal names the level at every call site | `SourceOwner(id)` helpers — one more name per level for S15-08 to add |
| D-2 | One unexported descriptor per level in `core/database/sourcemetadata/owner.go` *(new)*: `type store struct { ownerTable, ownerColumn, values, layout, typeOf string; notFound error; …built SQL }` and `var stores = map[vocabkey.Level]*store{vocabkey.LevelSource: newStore(…)}`. `newStore` builds every statement once with `fmt.Sprintf` from those constant names. The Source entry: `ownerTable "sources"`, `ownerColumn "source_id"`, `values "source_metadata"`, `layout "source_metadata_layout"`, `typeOf "SELECT source_type_id FROM sources WHERE id = ?"`, `notFound sources.ErrNotFound`. `func storeFor(o Owner) (*store, error)` returns `ErrInvalid` for a level with no entry or an id that is not 16 bytes; it is the first check of every exported function | S15-08 adds two entries; nothing else in the package names a table. `rowFacet` precedent | A `switch` per function — eight places for S15-08 to edit; one table with a level column — S15-08's tables are already decided (map § 6.6) |
| D-3 | Audit changes use `store.values` / `store.layout` as `EntityType` and `store.ownerColumn` as the owner field key. For the Source level this is exactly today's `source_metadata`, `source_metadata_layout`, `source_id` | The effects registry scopes by `field("source_id")`; S15-08 registers its two tables under their names | Separate entity names in the descriptor — equal to the table names at every level the map plans |
| D-4 | `Input` becomes `{FieldID []byte; ValueText string}` and `Row` becomes `{ID []byte; Owner Owner; FieldID []byte; ValueText string}` | The owner is the function's argument; a second id in the payload is a second source of truth that `CreateAll` had to cross-check | Keeping `Input.SourceID` — meaningless for Artifacts and graphs |
| D-5 | Signatures: `Set(tx, userID, owner, in Input)`, `Clear(tx, userID, owner, fieldID)`, `DismissSuggestion(tx, userID, owner, fieldID)`, `Reorder(tx, userID, owner, fieldIDs)`, `CreateAll(tx, userID, owner, d Draft)`, `Apply(tx, userID, owner, d Draft)` *(new)*, `List(c, owner)` (renamed from `ListBySource`), `ListWorkspace(c, owner)`, `ReadDraftTx(tx *sql.Tx, owner) (Draft, error)` *(new)*, `ListByOwnersTx(tx *sql.Tx, owners []Owner) ([]Row, error)` *(new)*. `ListBySourceTx` and `ListLayoutTx` are deleted | The map row: every exported function on an owner. `ReadDraftTx` is the read twin of `Apply`: the one thing a write path reads, and it covers both deleted Tx reads (their only outside caller is `sourcecopy`) | Keeping both Tx reads beside `ReadDraftTx` — two ways to read the same rows |
| D-6 | `type Draft struct { Values []Input; Layout []LayoutRow }` *(new)*. `ReadDraftTx` returns the owner's values in field-id order and its layout rows by `sort_order, field_id` | One shape for what a form or a copy holds | — |
| D-7 | `Apply` makes the stored values equal `d.Values` and makes each listed layout row equal its `d.Layout` row, verbatim. In order: (1) checks, before any write: `storeFor`, `tx`, `RequireUserID`, owner row exists (`ErrInvalid`), every field id 16 bytes, exists, appears once in `Values` and once in `Layout`, and every value normalizes (`normalizeValue` after `TrimSpace`; empty is `ErrInvalid`); (2) reads the stored values and layout (`ReadDraftTx`'s body); (3) value changes: a delete for each stored value not in the draft, in field-id order, then for each draft value in draft order an insert (no stored row) or an update (normalized text differs); (4) layout changes, in draft order: an insert (no stored row) or one update whose fields carry only the columns that differ. Layout rows not in the draft are kept. Apply never appends a layout row for a new value, unlike `Set` | Rule "Metadata saves with its owner's write": compares the draft with what is stored and runs the steps `Set`, `Clear`, `DismissSuggestion` and `Reorder` use. Verbatim layout is what `CreateAll` does today (its test "values with no layout rows write no layout" pins it), so `CreateAll` becomes the empty-owner case with identical rows and changes (AC-4.1) | Deleting layout rows missing from the draft — no operation removes a layout row today (a dismissed suggestion would silently come back); appending a layout row per new value as `Set` does — breaks `CreateAll`'s verbatim copy |
| D-8 | The write steps are shared, not copied: `insertValueTx` (exists), `updateValueTx` *(new, from `Set`'s inline update)*, `deleteValueTx` *(new, from `Clear`'s inline delete)*, `insertLayoutTx` *(new, from `CreateAll`'s loop)*, `updateLayoutTx` *(new)*. `Set` and `Clear` call them; their change shapes do not change. `Reorder` and `DismissSuggestion` keep their upsert SQL (`sqlLayoutOrder`, `sqlLayoutDismiss`) | One place per row change, so S15-08's levels get the same audit shape | Apply calling `Set` / `Clear` per field — `Set` appends layout rows and re-reads the field per call |
| D-9 | `CreateAll(tx, userID, owner, d)`: `storeFor`, `tx`, `RequireUserID`; `nil, nil` when `d` is empty (as today); `ErrInvalid` when the owner holds any value or layout row (`sqlHasAny` per store); then `return Apply(tx, userID, owner, d)` | Map row: `CreateAll` becomes Apply's empty-owner case | Deleting `CreateAll` — Copy Source's refusal of a non-empty target is a real rule |
| D-10 | `ListByOwnersTx`: `storeFor` on every owner first; `nil, nil` for no owners; then per level in `vocabkey.MetadataLevels()` order, one query `SELECT id, <ownerColumn>, field_id, COALESCE(value_text, '') FROM <values> WHERE <ownerColumn> IN (…) ORDER BY <ownerColumn>, field_id` over `database.UniqueBlobIDs` of that level's ids. Rows carry their `Owner`. No chunking: S15-16's largest call (300 graphs, map Scale) binds 300 values, far under SQLite's 32,766 | One query per level; `UNIQUE (source_id, field_id)` serves it | A query per owner — the fan-out the rule "Citations render on demand" forbids; a `UNION ALL` across levels — one level exists |
| D-11 | `List(c, owner)` and `ReadDraftTx` share the batch read's unexported body `listByOwners(q database.Querier, owners []Owner)` with one owner; `sqlListBySource` is deleted | One SQL path for values | — |
| D-12 | `ListWorkspace(c, owner)`: reads the owner's Source type with `store.typeOf` (no row → `store.notFound`, so `sources.ErrNotFound` for a Source, AC-1.3), suggestions with `sourcevocab.ListSuggestions(c, typeID, owner.Level)`, values with `listByOwners`, layout as today, and the extra values' fields with one `getFields(q, ids) (map[string]metadatafields.Field, error)` *(new)* query (`SELECT … FROM source_metadata_fields WHERE id IN (…)`), skipped when there is no extra. A missing field is `ErrInvalid`, as `getField` returns today. The merge rules and order do not change. `sources.Get` is no longer called | Map row: fields in one query. `typeOf` keeps the Source-type lookup per level, so the Artifact and graph levels read their Source's type (map § 6.6 "Templating follows the Source's type") | Moving `getFields` to `metadatafields` — `sourcemetadata` already reads fields privately (`getField`); a new export with one caller |
| D-13 | The write checks keep today's errors: a missing owner in a write is `ErrInvalid` (`requireOwner`, from `requireSource`, reading `SELECT 1 FROM <ownerTable> WHERE id = ?`) | Handler tests pin `sourcemetadata.ErrInvalid` on writes (`api/ffi/handlers/sources_test.go`) | `notFound` on writes — changes a wire error code |
| D-14 | `sourcecopy.Draft` replaces `Metadata` and `Layout` with `Metadata sourcemetadata.Draft`. `ReadDraft` sets it from `sourcemetadata.ReadDraftTx`; `CreateFromDraft` calls `sourcemetadata.CreateAll(tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: src.ID}, d.Metadata)`; `apply` sets `d.Metadata = sourcemetadata.Draft{}` when Metadata is off | No loop re-stamping `SourceID`; S15-08 copies an Artifact's draft with the same two calls | Keeping two fields — the pair always travels together |
| D-15 | Tests keep their helpers' `sourceID` parameters; the helpers (`runSet`, `runClear`, `runDismissSuggestion`, `runReorder`, `runCreateAll`, `runMetadataSet`, `runDismissMetadata`) build the owner inside. `runSet` and `runMetadataSet` take `(c, userID, sourceID, in Input)`; `runCreateAll` takes `(c, userID, sourceID, d Draft)` | Test bodies change only where they build `Input` or call the package directly | Changing every case to build owners — noise in 40 cases |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `sourcemetadata.Owner` *(new)* | type | `conclusionsummaries.FieldOwner` (a Conclusion owner and a Property), `vocabkey.Level` | Different domain; this pairs the shared `Level` with an id (searched `rg -n "type \w*Owner struct" core`) |
| `core/database/sourcemetadata/owner.go` *(new)*: `store`, `stores`, `newStore`, `storeFor`, `requireOwner` | file / type / func | `deleteimpact` `rowFacet` | Same pattern (constant names, SQL built once); a different package's concern |
| `sourcemetadata.Draft` *(new)* | type | `sourcecopy.Draft.Metadata` + `.Layout` | Replaces that pair (D-14) |
| `sourcemetadata.Apply` *(new)* | func | `CreateAll`, `Set`, `Clear` | `CreateAll` becomes its empty-owner case; the steps are shared (D-8) |
| `sourcemetadata.ReadDraftTx` *(new)* | func | `ListBySourceTx`, `ListLayoutTx` | Replaces both (D-5) |
| `sourcemetadata.ListByOwnersTx`, `listByOwners` *(new)* | func | `listBySource` | Replaces the single-owner query with one body (D-11) |
| `updateValueTx`, `deleteValueTx`, `insertLayoutTx`, `updateLayoutTx` *(new)* | func | inline code in `Set`, `Clear`, `CreateAll` | Extracted, not added beside (D-8) |
| `getFields` *(new)* | func | `getField`, `metadatafields.GetByID` | Batch form beside `getField` (D-12) |
| `TestApply`, `TestListByOwnersTx`, `TestOwnerLevels` *(new)* | test | `TestCreateAll` | New behavior |

**Builds on.** `vocabkey.Level` (S15-06); the `rowFacet` descriptor pattern; `database.UniqueBlobIDs` / `SQLInPlaceholders` / `BlobArgs`; the effects registry entries for `source_metadata` and `source_metadata_layout` (unchanged).
**Deletes.** `ListBySourceTx`, `ListLayoutTx`, `listBySource`, `sqlListBySource`, `requireSource` (becomes `requireOwner`), `sqlSourceExists`, every hand-written table name in the SQL constants (moved into `newStore`), `Input.SourceID`, `Row.SourceID`, `CreateAll`'s `Input.SourceID` cross-check, `sourcecopy.Draft.Layout` and `CreateFromDraft`'s `SourceID` loop. Test cases: `TestCreateAll` "input for another source is refused" (the state cannot be expressed, D-4); `TestCreateAll` "ListLayoutTx returns rows by sort order" becomes "ReadDraftTx returns layout rows by sort order" with the same assertions on `.Layout`.
**Next adds.** S15-08 adds `vocabkey.LevelArtifact` and `vocabkey.LevelGraph` entries to `stores` (owner table, column, two tables, `typeOf` joining to `sources`, `notFound`), passes `Owner{LevelArtifact, id}` from the Artifact handlers, and calls `Apply` inside the Artifact and graph create and update writes and `CreateAll` for a copied Artifact. S15-16 calls `ListByOwnersTx` once per render with the Sources, Artifacts and graphs of the batch.

### Reach

None — no new path into existing code. Every caller passes a `source` owner; the other levels are refused before any query (AC-2.1).

### Scale

At the map's Scale: `ListWorkspace` drops from 1 + (extra values) field reads to at most one (a Source with 20 extra values made 20 queries; now 1). `ListByOwnersTx` issues one query per level present; at S15-16's 30 Sources + 60 Artifacts + 300 graphs that is three queries once S15-08 wires the levels, each served by the `UNIQUE (owner, field_id)` index. `Apply` on a form draft reads the owner's values and layout once (two queries) and writes only differing rows; a 63-field draft is at most 63 value and 63 layout statements. No write's invalidation changes: the change rows are identical, so `effects` scopes, search docs and reads follow as today.

### Edges

- **Assumptions.** `ListWorkspace` and `Apply` key stored values by field id: one value per (owner, field), held by `UNIQUE (source_id, field_id)` (`000009.sql`; S15-08's tables carry the same `UNIQUE`, map § 6.6). `Apply` keys the draft the same way and refuses duplicates itself (AC-4.8, T-4) rather than relying on the constraint, because a duplicate layout field would reach an `UPDATE` that the constraint never sees.
- **Neighbors.** Equal after trim (AC-4.2) and equal after URL canonicalization (AC-4.3) write nothing; a value present in both with different text is an update, not a delete and insert (AC-4.4); a layout row differing only in `dismissed` or only in `sort_order` diffs only that column (AC-4.5, AC-4.6); a dismissed layout row on a valued field is written as given (AC-4.1, as `CreateAll` today), and `ListWorkspace` already shows a valued field whatever its flag.
- **Removed inputs.** A value left out of the draft is deleted with `Clear`'s change shape, and its layout row stays (as `Clear` today), so the field keeps its place if refilled. A layout row left out of the draft stays (AC-4.7): a draft never undoes a dismiss by omission.
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** None — owners come from engine code, not from data the app did not write; a draft is bounded by the field vocabulary (63 at Scale) and duplicates are refused before any write.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcemetadata/owner.go` *(new)* | D-1, D-2, D-13. Package doc moves to `sourcemetadata.go` unchanged except: "descriptive metadata values and layout for an owner: a Source today; an Artifact and an Evidence graph (S15-08) share the same functions" |
| `core/database/sourcemetadata/sourcemetadata.go` | D-3 – D-12. Doc comments say "owner" where they said "Source" and keep each rule sentence |
| `core/database/sourcecopy/sourcecopy.go` | D-14 |
| `api/ffi/handlers/sources.go` | Each of `GetSourceWorkspace`, `SetSourceMetadata`, `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata` builds `owner := sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: sourceID}` once and passes it; `SetSourceMetadata` builds `sourcemetadata.Input{FieldID: fieldID, ValueText: req.GetValueText()}` |

Sketch:

```go
// Owner is the record metadata hangs off: a Source today.
type Owner struct {
	Level vocabkey.Level
	ID    []byte
}

type Draft struct {
	Values []Input     // one per field
	Layout []LayoutRow // written verbatim
}

func Apply(tx *database.Tx, userID []byte, owner Owner, d Draft) ([]rowchange.Change, error)
func ListByOwnersTx(tx *sql.Tx, owners []Owner) ([]Row, error)
func ReadDraftTx(tx *sql.Tx, owner Owner) (Draft, error)
```

Errors: `sourcemetadata.invalid` → an unwired level, a bad id, a missing owner on a write, a duplicate or unknown field, an empty value; `sources.not_found` → `ListWorkspace` on a missing Source (unchanged).

### FFI

None — no proto change; handlers change call sites only.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy changes.

## 7. Steps

**STEP-1 — The owner and the Source store**
Files: `core/database/sourcemetadata/owner.go` *(new)*, `core/database/sourcemetadata/sourcemetadata.go`. Skill: `add-catalog-query`.
Do: D-1, D-2, D-13. Move every SQL constant that names `source_metadata`, `source_metadata_layout`, `source_id` or `sources` into fields of `store` built by `newStore`; `sqlFieldGet` stays a constant. Nothing calls the new store yet except `requireOwner`.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/...` succeeds.

**STEP-2 — Every function takes an owner**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`. Skill: `add-catalog-query`, `add-audit-scope`.
Do: D-3 – D-5, D-8 (extract the four steps), D-11, D-12, D-15, and `ReadDraftTx` (D-6). `CreateAll` takes `(tx, userID, owner, d Draft)` and keeps its body for now. Delete per §6 **Deletes** (not the `sourcecopy` lines yet). Update the test helpers and call sites; assertions stay.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/...` passes, including T-1, T-5, T-6.

**STEP-3 — Apply and the batch read**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`.
Do: D-7, D-9 (`CreateAll` now calls `Apply`), D-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/...` passes with T-2, T-3, T-4 and every `TestCreateAll` case.

**STEP-4 — Callers pass a Source owner**
Files: `core/database/sourcecopy/sourcecopy.go`, `core/database/sourcecopy/sourcecopy_test.go`, `core/database/deleteimpact/facets_test.go`, `api/ffi/handlers/sources.go`. Skill: `add-ffi-handler`.
Do: D-14 and the handler rows of §6 **Go core**. `sourcecopy_test.go` `layoutOf` reads `sourcemetadata.ReadDraftTx(tx, owner)` and returns `.Layout`; its `ListBySource` call becomes `List`.
Check: `rg -n "ListBySource|ListLayoutTx|Input\{SourceID|\.SourceID" core/database/sourcemetadata core/database/sourcecopy core/database/deleteimpact api/ffi/handlers --glob '*.go'` finds no `sourcemetadata` hit; `CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...` and `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` pass (T-7).

**STEP-5 — Docs**
Files: `docs/source-layer-data-model.md` (§ 5.3 DDL gains `UNIQUE (source_id, field_id)` as `000009.sql` has it; § 5.4 one sentence: `sourcemetadata` takes an owner, and `Apply` saves a draft of values and layout, writing layout rows verbatim and keeping rows the draft leaves out), `.agents/skills/add-catalog-query/SKILL.md` ("Reads inside a write": `sourcemetadata.ListLayoutTx` → `sourcemetadata.ReadDraftTx`), `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-07-metadata-operations-take-an-owner.md` and `scripts/check-pr.sh` print no error.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestSourceMetadata`, `TestCreateAll` (existing) | Every case but the deleted one passes with the helpers of D-15; assertions unchanged | AC-1.1, AC-2.2, AC-4.9 |
| T-2 | Go | `sourcemetadata_test.go` `TestApply` *(new)* | AC-4.1 (rows and change entity types and order equal `CreateAll`'s case "dismissed layout row on a valued field stays dismissed"); AC-4.2; AC-4.3; AC-4.4 (change actions and entity types in order, B's layout row still there); AC-4.5; AC-4.6 (changed field keys only); AC-4.7 | AC-4.1 – AC-4.7, D-7 |
| T-3 | Go | `sourcemetadata_test.go` `TestApply` *(new)* "refused drafts write nothing" | Repeated field in values; repeated field in layout; `"   "` value; unknown field id; owner id with no Source → each `ErrInvalid`, `source_metadata` and `source_metadata_layout` counts unchanged | AC-4.8 |
| T-4 | Go | `sourcemetadata_test.go` `TestListByOwnersTx` *(new)* | AC-3.1 (order by owner id, then field id; `Owner` on each row); AC-3.2; AC-3.3 | AC-3.1 – AC-3.3, D-10 |
| T-5 | Go | `sourcemetadata_test.go` `TestOwnerLevels` *(new)* | For each of `Set`, `Clear`, `DismissSuggestion`, `Reorder`, `Apply`, `CreateAll`, `List`, `ListWorkspace`, `ReadDraftTx`: owner level `artifact`, `graph`, `citation`, `""`, `page` and a `source` owner with a 15-byte id → `ErrInvalid`, no row written | AC-2.1, D-2 |
| T-6 | Go | `sourcemetadata_test.go` `TestSourceMetadata` *(new cases)* "list workspace reads several extras" and "list workspace on a missing source is not found" | AC-1.2 with three extra fields; AC-1.3 → `errors.Is(err, sources.ErrNotFound)` | AC-1.2, AC-1.3, D-12 |
| T-7 | Go | `core/database/sourcecopy/sourcecopy_test.go`, `core/database/deleteimpact/facets_test.go`, `api/ffi/handlers/sources_test.go` (existing) | Pass with call-site changes only | AC-1.1, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/sourcecopy` (the snapshot compares layout and workspace of original and copy), `core/database/audit` `TestRecord` (counts `source_metadata` creates), `api/ffi/handlers` `TestSetSourceMetadata` and `TestDismissAndReorderSourceMetadata`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `sourcemetadata.ListBySource`, `sourcemetadata.ListBySourceTx`, `sourcemetadata.ListLayoutTx`, `Input.SourceID` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory (no UI)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-15/deployment-plan.md: map row status; docs/deployment-plan/spike-15/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated)
- [ ] Domain doc updated: source-layer-data-model.md §§ 5.3, 5.4; add-catalog-query "Reads inside a write" example
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the dogfood project. On a birth certificate Source: fill a suggested field, add an extra field, dismiss a suggestion, drag to reorder, clear a value. Each behaves as before, and the Source's history lists each change.
2. **Copy Source** with Metadata ticked: the copy's metadata board shows the same fields, values, order and dismissed suggestions as the original.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A change shape drifts (a field renamed, an entity type built wrong) and audit scopes or search stop following | D-3 keeps the names; T-1 keeps today's assertions; the `audit` and `effects` suites run in `check-pr.sh` |
| `Apply`'s changes for an empty owner differ from today's `CreateAll` in order | T-2's AC-4.1 compares entity types and order with `TestCreateAll`'s case; `sourcecopy`'s snapshot test compares the copy |
| Building SQL with `fmt.Sprintf` reaches data | `newStore` takes only the constant names in `stores`; no argument comes from a caller |

**Stop and report if:**

- `rg -n "sourcemetadata\." --glob '*.go'` finds a caller outside `api/ffi/handlers/sources.go`, `core/database/sourcecopy`, `core/database/deleteimpact/facets_test.go` and the package itself.
- A test in §4 needs an assertion changed (not a call site) other than the two cases §6 **Deletes** names.
- `TestApply`'s AC-4.1 case cannot match `CreateAll`'s rows and change order without appending a layout row (D-7 is wrong).
- A handler test expects an error other than `sourcemetadata.ErrInvalid` on a metadata write, or `ListWorkspace` on a missing Source returns anything but `sources.ErrNotFound`.
- The effects or audit suites fail on a `source_metadata` or `source_metadata_layout` change.

## 12. Open questions

None.

## 13. PR

**Title:** S15-07 — Metadata operations take an owner
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-07-metadata-operations-take-an-owner.md`.
