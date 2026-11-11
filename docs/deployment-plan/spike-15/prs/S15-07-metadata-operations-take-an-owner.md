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
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** working rules "Metadata lives at three levels with one vocabulary", "One level type", "Metadata saves with its owner's write", "Citations render on demand, from SQL" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`evidence-explained-citation-map.md`](../evidence-explained-citation-map.md) § 6.6; [`source-layer-data-model.md`](../../../source-layer-data-model.md) §§ 5.3, 5.4; [`S15-32`](S15-32-copy-source-date.md) Next adds (the `sourcecopy` overlap)
**Planned:** 2026-11-11 against `main` @ `bdef66e`
**Review:** [`S15-07-metadata-operations-take-an-owner.review.md`](S15-07-metadata-operations-take-an-owner.review.md) — round 1 answered; round 2 pending

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

Every exported function of `core/database/sourcemetadata` takes an **owner** (a `vocabkey.Level` and an id) instead of a Source id. Only the Source level is wired: its tables, audit shape and behavior are exactly today's. The package also gains the pieces later PRs build on: one batch read of values by owners (S15-16's citation resolver); a workspace read written over many owners, so S15-08 can return 60 Artifact boards in one read; and `Apply`, which saves a form's draft by comparing it with what is stored (S15-08's Artifact and graph writes), with `CreateAll` as its empty-owner case. `ListWorkspace` reads in one read transaction and stops reading one field per extra value. S15-08 then adds the Artifact and graph levels as two entries in one table, not a second copy of the logic.

**Why now.** Working rules "Metadata lives at three levels with one vocabulary" and "Metadata saves with its owner's write"; S15-08 depends on it, and the map review (K-9) moved every function and the batch read here so S15-08 and S15-16 add no logic of their own.

**What stays.** The FFI wire, every handler's reply and error, the Source page, Copy Source, the audit rows (entity types, fields, scopes) and the catalog schema. Every existing metadata test keeps its assertions; only call sites change, except the cases §6 **Deletes** names.

## 2. User stories

**US-1.** As a researcher, when I edit, dismiss, reorder or copy a Source's metadata after this PR, everything behaves and is audited as before, so that the refactor changes nothing I see.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The existing tests of `sourcemetadata`, `sourcecopy`, `deleteimpact`, `audit` and `api/ffi/handlers` | Pass with call-site changes only (§6 **Deletes** lists the cases whose state can no longer be expressed) |
| AC-1.2 | A Source with values for `document_number` (suggested) and three extra fields `a`, `b`, `c`; `ListWorkspace` | Lists the suggestions in type order, then `a`, `b`, `c` in field-id order, each with its field's key and label, as today |
| AC-1.3 | `GetSourceWorkspace` with a Source id that has no row | `sources.not_found`, as today (the handler's `sources.Get` answers before `ListWorkspace`); `ListWorkspace` itself returns `ErrInvalid` for a missing owner |

**US-2.** As the engine, I address metadata by owner, so that S15-08 adds the Artifact and graph levels as two table entries.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Any exported function with an owner at a level of `vocabkey.Levels()` that has no store entry (today `artifact`, `graph`, `citation`), or at `""` or `page`, or a `source` owner whose id is not 16 bytes | `ErrInvalid`; nothing is written |
| AC-2.2 | `Set` on a `source` owner | The change is `source_metadata` with a `source_id` field, as today; the audit scope is that Source |
| AC-2.3 | Each store entry | Its value and layout tables are in `catalogmodel.Tables`, each with an FK from its owner column; so `effects` `TestEveryNonSkipTableHasAnEntry` requires their effect entries |

**US-3.** As the engine, I read the values of many owners at once, so that S15-16 renders 300 Citations in a few queries.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Sources A (fields F1, F2) and B (F1) and C (none); `ListByOwnersTx(tx, [B, A, C, A])` | Three rows, A's then B's when A's id sorts first, each owner's rows by field id, each row carrying its `Owner` |
| AC-3.2 | `ListByOwnersTx(tx, nil)` | `nil`, no error, no query |
| AC-3.3 | The owners include one at `graph` | `ErrInvalid` before any query |
| AC-3.4 | The unexported workspace body over Sources A (type T1, one layout row) and B (type T2, none) | A's entries in A's layout order and B's in T2's suggestion order: the same lists `ListWorkspace` returns for each alone |

**US-4.** As the engine, I save a draft of values and layout in one step, so that S15-08 writes an Artifact's or a graph's metadata in the same transaction as the Artifact or graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | An owner with nothing stored; `Apply` with values `[document_number "No. 7"]` and layout `[link 0, document_number 1 dismissed]` | The same rows and the same changes, in the same order, as `CreateAll` writes today: one `source_metadata` create, then two `source_metadata_layout` creates |
| AC-4.2 | Stored `document_number "No. 7"`; draft value `"  No. 7  "` | No change, no write |
| AC-4.3 | Stored URL `https://example.org/a`; draft `HTTPS://Example.org/a` | No change |
| AC-4.4 | Stored values A `"x"` and B `"y"`; draft values `[A "z"]` | Changes: B's `source_metadata` delete (as `Clear` writes it), then A's update (as `Set` writes it); B's layout row stays |
| AC-4.5 | Stored layout A 0, B 1; draft layout `[B 0, A 1]` | Two `source_metadata_layout` updates, each with fields `source_id`, `field_id` (old = new) and `sort_order` {old, new}, and no `dismissed` key |
| AC-4.6 | Stored layout A 0 not dismissed; draft layout `[A 0 dismissed]` | One update with fields `source_id`, `field_id` (old = new) and `dismissed` {false, true}, and no `sort_order` key |
| AC-4.7 | Stored layout rows A and C; draft layout `[A …]` | C's row stays as it is |
| AC-4.8 | A draft repeating a field in its values, or in its layout; an empty value; an unknown field; an owner with no row | `ErrInvalid`; nothing is written |
| AC-4.9 | `CreateAll` on an owner that already holds a value or a layout row | `ErrInvalid`, as today |
| AC-4.10 | Stored values A and B with layout rows; `Apply` with an empty `Draft` | Two `source_metadata` deletes; both layout rows stay. An empty draft clears every value: it is not a no-op |
| AC-4.11 | `ReadDraftTx` on a `source` owner whose Source does not exist | An empty `Draft` and no error (it reads rows, not the owner) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `Owner` and the per-level store table, with only `vocabkey.LevelSource` wired | STEP-1, AC-2.1, AC-2.3 |
| `Set`, `Clear`, `DismissSuggestion`, `Reorder`, `ListWorkspace`, `CreateAll` take an owner; `Input` and `Row` lose `SourceID` | STEP-1, STEP-2, AC-1.1, AC-2.2 |
| `ReadDraftTx` replaces `ListBySource`, `ListBySourceTx` and `ListLayoutTx`; `sourcecopy` reads and writes metadata as one `sourcemetadata.Draft` | STEP-2, AC-1.1, AC-4.11 |
| One layout change builder for every layout write | STEP-2, AC-4.5, AC-4.6 |
| `ListWorkspace` over an unexported many-owner body, in one read transaction, with one field read; `sourcevocab.ListSuggestionsTx` and `metadatafields.ListByIDsTx` | STEP-2, AC-1.2, AC-3.4 |
| `ListByOwnersTx` (the batch read) | STEP-3, AC-3.1 – AC-3.3 |
| `Apply`; `CreateAll` reads the owner, refuses a non-empty one and applies | STEP-3, AC-4.1 – AC-4.10 |
| Handlers and every Go caller pass a `source` owner | STEP-2, AC-1.1, AC-1.3 |
| Docs: `source-layer-data-model.md` §§ 5.3, 5.4, `add-catalog-query` example | STEP-4 |

| Out | Home |
| --- | --- |
| The Artifact and graph tables, their store entries, audit, delete, Copy Source of Artifact values, FFI | S15-08 |
| An exported many-owner workspace read | S15-08 (the Source page's Artifact boards) |
| A caller of `ListByOwnersTx` | S15-16 |
| A caller of `Apply` other than `CreateAll` | S15-08 (Artifact and graph writes) |
| Where a field added on a draft board sits in the order | S15-26 (the board on a draft sends its whole order as `Draft.Layout`; D-7) |
| Renaming the package `sourcemetadata` or the table `source_metadata_fields` | not planned (map § 6.6 "Name") |
| Search over Artifact or graph values | not planned (S15-08 "Not this PR") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Package | `core/database/sourcemetadata/sourcemetadata.go` | 700 lines; every SQL constant names `source_metadata` / `source_metadata_layout` / `source_id` directly; imports `sources` only for `sources.Get` in `ListWorkspace` |
| Types | same, `Row{ID, SourceID, FieldID, ValueText}`, `Input{SourceID, FieldID, ValueText}`, `WorkspaceEntry`, `LayoutRow{FieldID, SortOrder, Dismissed}` | Source-keyed |
| Writes | same, `Set(tx, userID, in Input)`, `Clear(tx, userID, sourceID, fieldID)`, `DismissSuggestion(tx, userID, sourceID, fieldID)`, `Reorder(tx, userID, sourceID, fieldIDs)`, `CreateAll(tx, userID, sourceID, values []Input, layout []LayoutRow)` | `Set` inserts through `insertValueTx` and appends a layout row (`sqlLayoutAppend`); updates inline with `sqlUpdate`. `Clear` deletes inline. `CreateAll` returns `nil, nil` on empty input before any check, refuses a Source with any row (`sqlHasAny`) or an `Input.SourceID` that differs from `sourceID`, then inserts values (`insertValueTx`) and layout rows verbatim (`sqlLayoutInsert`): changes are values in input order, then layout in input order. A missing Source is `ErrInvalid` (`requireSource`) |
| Layout changes | same, `CreateAll`, `DismissSuggestion`, `Reorder` | Built inline in three shapes: `CreateAll` create {`source_id`, `field_id`, `sort_order`, `dismissed`}; `DismissSuggestion` create or update {`source_id`, `field_id`, `dismissed`} with `source_id`/`field_id` old = new; `Reorder` create or update {`source_id`, `field_id`, `sort_order`} likewise. `EntityID` is the field id |
| Reads | same, `ListBySource(c, sourceID)` (only callers: `ListWorkspace` and `sourcecopy_test.go`), `ListBySourceTx(tx, sourceID)`, `ListLayoutTx(tx, sourceID)`, `ListWorkspace(c, sourceID)` | `ListWorkspace` runs separate `c.DB()` reads with no read transaction: `sources.Get` (missing → `sources.ErrNotFound`, commit `c3b78f6`), `sourcevocab.ListSuggestions(c, typeID, vocabkey.LevelSource)`, values, `listLayout`, then `getField(db, v.FieldID)` **once per extra value** |
| Field reads | `sourcemetadata.go` `sqlFieldGet`, `getField`; `core/database/metadatafields/metadatafields.go` `sqlGetByID`, `GetByID`, `getByIDTx` | The same field-row scan in two packages |
| Suggestions | `core/database/sourcevocab/sourcevocab.go` `ListSuggestions(c, typeID, level)` | Catalog form only; no Tx twin |
| Value changes | `sourcemetadata.go` | `EntityType "source_metadata"`, `EntityID` the row id, fields `id`, `source_id`, `field_id`, `value_text` |
| Schema | `core/database/migrations/000009.sql`, `000031.sql` | `source_metadata` has `UNIQUE (source_id, field_id)` (the domain doc's § 5.3 DDL omits it) and a `date_value_id` column that no code reads or writes (`000028.sql` moved its dates into `value_text`; § 5.3 says "There is no `date_value_id` on this table", which describes the code, not the column); layout PK `(source_id, field_id)` |
| Effects | `core/database/effects/registry.go` `"source_metadata"`, `"source_metadata_layout"`; `core/database/effects/path.go` `readChangeColumn` | `SourceScope: field("source_id")`, read from the change's fields; a non-delete change without `source_id` falls back to a row read by `pkOf`, which `source_metadata_layout` (no `PK` in `catalogmodel`) cannot serve. Unchanged by this PR |
| Catalog model | `core/database/catalogmodel/model.go` `Tables`, `FKs` | `source_metadata` and `source_metadata_layout` with FKs from `source_id` to `sources`; `effects` `TestEveryNonSkipTableHasAnEntry` requires an effect entry per table |
| Import graph | `go list -deps` | `sources` imports `artifacts` and `evidencegraphs`; `sourcemetadata` imports `sources`. Owner writes are composed in handlers (`api/ffi/handlers/artifacts.go`, `evidence_graphs.go`) or composing packages (`sourcecopy`) |
| Copy Source | `core/database/sourcecopy/sourcecopy.go` `Draft.Metadata []sourcemetadata.Input`, `Draft.Layout []sourcemetadata.LayoutRow`, `ReadDraft`, `CreateFromDraft`, `apply` | `ReadDraft` calls `ListBySourceTx` and `ListLayoutTx`; `CreateFromDraft` sets each `Input.SourceID` and calls `CreateAll`; `apply` nils both fields when Metadata is off |
| Handlers | `api/ffi/handlers/sources.go` `GetSourceWorkspace` (calls `sources.Get` before `ListWorkspace`), `SetSourceMetadata`, `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata` | Pass `sourceID` to eight `sourcemetadata` calls: `ListWorkspace` four times, `Set`, `Clear`, `DismissSuggestion`, `Reorder` once each |
| Batch helpers | `core/database/sqlin.go` `UniqueBlobIDs`, `SQLInPlaceholders`, `BlobArgs` | Used by `autoreconciler`, `searchindex` for `IN (…)` reads |
| Tests | `core/database/sourcemetadata/sourcemetadata_test.go` `TestSourceMetadata`, `TestCreateAll` and helpers `runSet`, `runClear`, `runDismissSuggestion`, `runReorder`, `runCreateAll`; `core/database/sourcecopy/sourcecopy_test.go` (`ListWorkspace`, `ListLayoutTx` in `layoutOf`, `ListBySource` in `valueKeys`, `Set`, `DismissSuggestion`, `Reorder`); `core/database/deleteimpact/facets_test.go` `runMetadataSet`, `runDismissMetadata`; `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace` ("unknown source is not found"), `TestSetSourceMetadata`, `TestDismissAndReorderSourceMetadata` | Source-keyed call sites |
| Skill | `.agents/skills/add-catalog-query/SKILL.md` "Reads inside a write" | Names `sourcemetadata.ListLayoutTx` as an example Tx twin |

**Precedent.** `core/database/deleteimpact/facets.go` `rowFacet` is a function that takes constant table and column names and builds its SQL once with `fmt.Sprintf`, one call per facet table: copy that construction for `newStore` (here the result is a struct, D-2). `core/database/autoreconciler/autoreconciler.go` (`sqlLoadCandidates` + `database.SQLInPlaceholders` + `database.BlobArgs`) for each `IN (…)` read. `deleteimpact.Snapshot` and `evidencegraphs.List` for a multi-query read in one read transaction.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `type Owner struct { Level vocabkey.Level; ID []byte }` in `sourcemetadata`. Callers write the literal `sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: sourceID}`; no constructor | Rule "One level type" (owners use `vocabkey.Level`). A literal names the level at every call site | `SourceOwner(id)` helpers — one more name per level for S15-08 to add |
| D-2 | One unexported descriptor per level in `core/database/sourcemetadata/owner.go` *(new)*: `type store struct` with the names `ownerColumn`, `values`, `layout` (strings, used in change rows) and the SQL fields `typeIn` (a prefix: the owner's id and its Source type id for owners `IN (`), `ownerExists`, `insert`, `update`, `delete`, `getPair`, `listIn` (prefix `SELECT id, <ownerColumn>, field_id, COALESCE(value_text, '') FROM <values> WHERE <ownerColumn> IN (`), `layoutListIn` (prefix `SELECT <ownerColumn>, field_id, sort_order, dismissed FROM <layout> WHERE <ownerColumn> IN (`), `layoutAppend`, `layoutDismiss`, `layoutOrder`, `layoutInsert`, `layoutUpdate` (*new*: `UPDATE <layout> SET sort_order = ?, dismissed = ? WHERE <ownerColumn> = ? AND field_id = ?`), `layoutGet` — each the same text as today's constant of that role with the names substituted. A prefix is completed per call with `database.SQLInPlaceholders(n)` and its suffix (`) ORDER BY <ownerColumn>, field_id` for `listIn`, `) ORDER BY <ownerColumn>, sort_order, field_id` for `layoutListIn`, `)` for `typeIn`), as `autoreconciler` `sqlLoadCandidates` is. `func newStore(ownerTable, ownerColumn, values, layout, typeIn string) *store` builds them once with `fmt.Sprintf`; `typeIn` is passed whole. `var stores = map[vocabkey.Level]*store{vocabkey.LevelSource: newStore("sources", "source_id", "source_metadata", "source_metadata_layout", "SELECT id, source_type_id FROM sources WHERE id IN (")}`. `func storeFor(o Owner) (*store, error)` returns `ErrInvalid` for a level with no entry or an id that is not 16 bytes; it is the first check of every exported function. `sqlHasAny`, `sqlLayoutList` and `sqlListBySource` have no store field: the `IN` reads replace them | S15-08 adds two entries; nothing else in the package names a table. `rowFacet` precedent | A `switch` per function — eight places for S15-08 to edit; one table with a level column — S15-08's tables are already decided (map § 6.6) |
| D-3 | Audit changes use `store.values` / `store.layout` as `EntityType` and `store.ownerColumn` as the owner field key. For the Source level this is exactly today's `source_metadata`, `source_metadata_layout`, `source_id` | The effects registry scopes by `field("source_id")`; S15-08 registers its two tables under their names | Separate entity names in the descriptor — equal to the table names at every level the map plans |
| D-4 | `Input` becomes `{FieldID []byte; ValueText string}` and `Row` becomes `{ID []byte; Owner Owner; FieldID []byte; ValueText string}` | The owner is the function's argument; a second id in the payload is a second source of truth that `CreateAll` had to cross-check | Keeping `Input.SourceID` — meaningless for Artifacts and graphs |
| D-5 | Exported after this PR: `Owner`, `Input`, `Row`, `Draft`, `LayoutRow`, `WorkspaceEntry`, `ErrInvalid`; `Set(tx, userID, owner, in Input)`, `Clear(tx, userID, owner, fieldID)`, `DismissSuggestion(tx, userID, owner, fieldID)`, `Reorder(tx, userID, owner, fieldIDs)`, `CreateAll(tx, userID, owner, d Draft)`, `Apply(tx, userID, owner, d Draft)` *(new)*, `ListWorkspace(c, owner)`, `ReadDraftTx(tx *sql.Tx, owner) (Draft, error)` *(new)*, `ListByOwnersTx(tx *sql.Tx, owners []Owner) ([]Row, error)` *(new)*. `ListBySource`, `ListBySourceTx` and `ListLayoutTx` are deleted | The map row: every exported function on an owner. `ReadDraftTx` is the read twin of `Apply` and covers the three deleted reads (their only outside callers are `sourcecopy` and its test) | Keeping `List(c, owner)` — its only caller outside the package would be a test (review M-6) |
| D-6 | `type Draft struct { Values []Input; Layout []LayoutRow }` *(new)*. `ReadDraftTx` returns the owner's values in field-id order (`listByOwners`) and its layout rows by `sort_order, field_id` (`listLayoutByOwners`); a nil `tx` is `ErrInvalid` | One shape for what a form or a copy holds | — |
| D-7 | `Apply` makes the stored values equal `d.Values` and makes each listed layout row equal its `d.Layout` row, verbatim. It reads what is stored with `ReadDraftTx`'s body and calls the unexported `apply(tx, userID, st *store, owner, stored, d Draft)`, which: (1) checks, before any write: `tx`, `RequireUserID`, owner row exists (`ownerExists`; `ErrInvalid`), every field id 16 bytes and found by `metadatafields.ListByIDsTx` (one read), each field once in `Values` and once in `Layout`, and every value normalizes (`normalizeValue` after `TrimSpace`; empty is `ErrInvalid`); (2) value changes: a delete for each stored value not in the draft, in field-id order, then for each draft value in draft order an insert (no stored row) or an update (normalized text differs); (3) layout changes, in draft order: an insert (no stored row) or, when `sort_order` or `dismissed` differs, one update (`layoutUpdate` writes both columns), each change built by `layoutChange` (D-8) with only the differing columns beyond owner and `field_id`. A layout row is handled the same whether or not its field holds a value. Layout rows not in the draft are kept. Apply never appends a layout row for a new value, unlike `Set`: a caller with a board sends its whole order as `Draft.Layout`, which the board already holds. An empty draft deletes every stored value (AC-4.10) | Rule "Metadata saves with its owner's write": compares the draft with what is stored and runs the steps `Set`, `Clear`, `DismissSuggestion` and `Reorder` use. Verbatim layout is what `CreateAll` does today (its test "values with no layout rows write no layout" pins it), so `CreateAll` becomes the empty-owner case with identical rows and changes (AC-4.1) | Deleting layout rows missing from the draft — no operation removes a layout row today (a dismissed suggestion would silently come back); appending a layout row per new value as `Set` does — breaks `CreateAll`'s verbatim copy, and appending only when the owner already has layout (review M-8) gives one draft two meanings |
| D-8 | Every row change is built in one place. Values: `insertValueTx` (exists), `updateValueTx` *(new, from `Set`'s inline update)*, `deleteValueTx` *(new, from `Clear`'s inline delete)*; `Set`, `Clear` and `apply` call them. Layout: `layoutChange(st *store, owner Owner, fieldID []byte, action rowchange.Action, diffs map[string]rowchange.FieldDiff) rowchange.Change` *(new)* always writes `<ownerColumn>` and `field_id` (`{nil, id}` on create, `{id, id}` on update) and adds `diffs`; `insertLayoutTx` and `updateLayoutTx` *(new)* and `Reorder` and `DismissSuggestion` all build their changes with it. `Reorder` and `DismissSuggestion` keep their upsert SQL (`layoutOrder`, `layoutDismiss`), which the effects path does not see; their change fields stay exactly today's | One shape per layout change at every level: the owner column is what `effects` scopes by (`readChangeColumn`), so no path can drop it (review M-1, M-2) | Rewriting `Reorder` and `DismissSuggestion` onto `layoutUpdate` — their atomic upserts (append at `MAX + 1`, keep `dismissed`) are the rule they implement |
| D-9 | `CreateAll(tx, userID, owner, d)`: `storeFor`, `tx`, `RequireUserID`; `nil, nil` when `d` is empty (as today, before any read); reads `stored` with `ReadDraftTx`'s body; `ErrInvalid` when `stored` has any value or layout row; then `return apply(tx, userID, st, owner, stored, d)` | Map row: `CreateAll` becomes Apply's empty-owner case. One read of what an owner holds serves the refusal and the compare (review M-9) | A separate `hasAny` query — a second reading of the same state |
| D-10 | `ListByOwnersTx`: `storeFor` on every owner first; a nil `tx` is `ErrInvalid`; `nil, nil` for no owners; then the unexported `listByOwners(q database.Querier, owners)`: per level in `vocabkey.MetadataLevels()` order, one `listIn` query over `database.UniqueBlobIDs` of that level's ids (safe after `storeFor`, which refused every id that is not 16 bytes). Rows carry their `Owner`. `listLayoutByOwners(q, owners) (map[string][]LayoutRow, error)` *(new)* does the same with `layoutListIn`, keyed by `string(level)+string(id)`. No chunking: S15-16's largest call (300 graphs, map Scale) binds 300 values, far under SQLite's 32,766 | One query per level; `UNIQUE (source_id, field_id)` and the layout PK serve them | A query per owner — the fan-out the rule "Citations render on demand" forbids |
| D-11 | `ListWorkspace(c, owner)`: `storeFor`; opens one read transaction (`db.Begin`, deferred `Rollback`, as `deleteimpact.Snapshot`) and calls the unexported `listWorkspaces(tx *sql.Tx, owners []Owner) (map[string][]WorkspaceEntry, error)` *(new)* with one owner, keyed as D-10. `listWorkspaces` reads: each owner's Source type with one `typeIn` query per level (an owner with no row → `ErrInvalid`); suggestions with `sourcevocab.ListSuggestionsTx(tx, typeID, level)` once per distinct (type, level); values with `listByOwners`; layout with `listLayoutByOwners`; and every extra value's field across all owners with one `metadatafields.ListByIDsTx` (skipped when there is no extra; a field not found → `ErrInvalid`). Then today's merge, per owner, unchanged. `sources.Get` and the `sources` import go | Map rows: S15-07 "fields in one query"; S15-08 "its Artifacts' boards in one query (60 Artifacts at Scale)", which then adds an export over `listWorkspaces` (review M-3). One read transaction is the `add-catalog-query` rule for a multi-query read (review M-4). Dropping the `sources` import keeps the descriptor free of owner packages, whose `artifacts` and `evidencegraphs` S15-08 would otherwise pull toward a cycle (review M-7) | A per-owner workspace now and a batch one in S15-08 — two merges; keeping `notFound` per store — imports each owner package; `GetSourceWorkspace` already answers `sources.not_found` first (AC-1.3) |
| D-12 | `sourcevocab.ListSuggestionsTx(tx *sql.Tx, typeID []byte, level vocabkey.Level) ([]Suggestion, error)` *(new)*: `ListSuggestions`' body moves to `listSuggestions(q database.Querier, …)` and both call it, with the same checks | The skill's Tx twin, added with its first caller outside the package | — |
| D-13 | `metadatafields.ListByIDsTx(tx *sql.Tx, ids [][]byte) ([]metadatafields.Field, error)` *(new)*: one `SELECT id, key, origin, label, data_type, COALESCE(description, '') FROM source_metadata_fields WHERE id IN (…)` over `database.UniqueBlobIDs(ids)`; returns the fields found, in no set order; `nil` for no ids. `sourcemetadata` maps by id and turns a missing one into its `ErrInvalid`. `getField` and `sqlFieldGet` in `sourcemetadata` are deleted; `Set`, `DismissSuggestion` and `Reorder` read their fields through it | The field row is read in the package that owns its table (`catalog-queries.mdc`); one scan instead of two copies (review M-5) | A private `getFields` in `sourcemetadata` — a third copy of the field scan |
| D-14 | The write checks keep today's errors: a missing owner in a write is `ErrInvalid` (`requireOwner`, from `requireSource`, with `ownerExists`: `SELECT 1 FROM <ownerTable> WHERE id = ?`) | Handler tests pin `sourcemetadata.ErrInvalid` on writes (`api/ffi/handlers/sources_test.go`) | — |
| D-15 | `sourcecopy.Draft` replaces `Metadata` and `Layout` with `Metadata sourcemetadata.Draft`. `ReadDraft` sets it from `sourcemetadata.ReadDraftTx`; `CreateFromDraft` calls `sourcemetadata.CreateAll(tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: src.ID}, d.Metadata)`; `apply` sets `d.Metadata = sourcemetadata.Draft{}` when Metadata is off | No loop re-stamping `SourceID`; S15-08 copies an Artifact's draft with the same two calls | Keeping two fields — the pair always travels together |
| D-16 | Tests keep their helpers' `sourceID` parameters; the helpers (`runSet`, `runClear`, `runDismissSuggestion`, `runReorder`, `runCreateAll`, `runMetadataSet`, `runDismissMetadata`) build the owner inside. `runSet` and `runMetadataSet` take `(c, userID, sourceID, in Input)`; `runCreateAll` takes `(c, userID, sourceID, d Draft)`. In-package tests that called `ListBySource` call a test helper `values(t, c, sourceID) []Row` *(new)* over `ReadDraftTx`'s body in a read transaction; `sourcecopy_test.go` reads `ReadDraftTx(tx, owner)` | Test bodies change only where they build `Input` or call the package directly | Changing every case to build owners — noise in 40 cases |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `sourcemetadata.Owner` *(new)* | type | `conclusionsummaries.FieldOwner` (a Conclusion owner and a Property), `vocabkey.Level` | Different domain; this pairs the shared `Level` with an id (searched `rg -n "type \w*Owner struct" core`) |
| `core/database/sourcemetadata/owner.go` *(new)*: `store`, `stores`, `newStore`, `storeFor`, `requireOwner` | file / type / func | `deleteimpact` `rowFacet` | Same construction (constant names, SQL built once); a different package's concern |
| `sourcemetadata.Draft` *(new)* | type | `sourcecopy.Draft.Metadata` + `.Layout` | Replaces that pair (D-15) |
| `sourcemetadata.Apply`, `apply` *(new)* | func | `CreateAll`, `Set`, `Clear` | `CreateAll` becomes its empty-owner case; the steps are shared (D-8) |
| `sourcemetadata.ReadDraftTx` *(new)* | func | `ListBySource`, `ListBySourceTx`, `ListLayoutTx` | Replaces all three (D-5) |
| `sourcemetadata.ListByOwnersTx`, `listByOwners`, `listLayoutByOwners` *(new)* | func | `listBySource`, `listLayout`, `listLayoutRows` | Replace the single-owner reads with one body each (D-10) |
| `listWorkspaces` *(new)* | func | `ListWorkspace`'s body | The body itself, over owners (D-11) |
| `updateValueTx`, `deleteValueTx`, `insertLayoutTx`, `updateLayoutTx`, `layoutChange` *(new)* | func | inline code in `Set`, `Clear`, `CreateAll`, `DismissSuggestion`, `Reorder` | Extracted, not added beside (D-8) |
| `sourcevocab.ListSuggestionsTx`, `listSuggestions` *(new)* | func | `sourcevocab.ListSuggestions` | Its Tx twin (D-12) |
| `metadatafields.ListByIDsTx` *(new)* | func | `metadatafields.getByIDTx`, `sourcemetadata.getField` | The batch read in the owning package; replaces `getField` (D-13) |
| `TestApply`, `TestListByOwnersTx`, `TestOwnerLevels`, `TestStoresAreRegistered`, `TestListWorkspaces` *(new)*; `metadatafields` `TestListByIDsTx` *(new)*; helper `values` *(new)* | test | `TestCreateAll` | New behavior |

**Builds on.** `vocabkey.Level` (S15-06); the `rowFacet` construction; `database.UniqueBlobIDs` / `SQLInPlaceholders` / `BlobArgs`; `catalogmodel.Tables` and `FKs`; the effects registry entries for `source_metadata` and `source_metadata_layout` (unchanged).
**Deletes.** `ListBySource`, `ListBySourceTx`, `ListLayoutTx`, `listBySource`, `listLayout`, `listLayoutRows`, `sqlListBySource`, `sqlLayoutList`, `sqlHasAny`, `sqlSourceExists`, `requireSource` (becomes `requireOwner`), `getField`, `sqlFieldGet`, every hand-written table name in the SQL constants (moved into `newStore`), the `sources` import, `Input.SourceID`, `Row.SourceID`, `CreateAll`'s `Input.SourceID` cross-check, the three inline layout change builders, `sourcecopy.Draft.Layout` and `CreateFromDraft`'s `SourceID` loop. Test cases: `TestCreateAll` "input for another source is refused" (the state cannot be expressed, D-4); `TestCreateAll` "ListLayoutTx returns rows by sort order" becomes "ReadDraftTx returns layout rows by sort order" with the same assertions on `.Layout`.
**Next adds.** S15-08 adds `vocabkey.LevelArtifact` and `vocabkey.LevelGraph` entries to `stores` (owner table, column, two tables, a `typeIn` joining to `sources`); `TestStoresAreRegistered` then requires their `catalogmodel` rows and, through `effects`' completeness test, their effect entries. S15-08 exports a many-owner board read over `listWorkspaces` for the Source page's Artifacts; calls `Apply` from the handlers that compose the Artifact and graph writes (not from `artifacts` or `evidencegraphs`: `sources` imports both); and calls `CreateAll` for a copied Artifact. S15-16 calls `ListByOwnersTx` once per render. S15-26's board on a draft sends its whole order as `Draft.Layout`.

### Reach

None — no new path into existing code. Every caller passes a `source` owner; the other levels are refused before any query (AC-2.1). `ListWorkspace` on a missing Source now returns `ErrInvalid` instead of `sources.ErrNotFound`; its only callers are handlers, and the one that can see a missing Source, `GetSourceWorkspace`, calls `sources.Get` first; the write handlers call it after a write that already refused a missing Source with `ErrInvalid` (AC-1.3).

### Scale

At the map's Scale: `ListWorkspace` drops from 4 + (extra values) queries to at most 5 (type, suggestions, values, layout, fields), in one read transaction; a Source with 20 extra values made 24 queries. S15-08's 60 Artifact boards through `listWorkspaces` cost 4 queries plus one suggestion read per distinct (type, level), not 60 × 5. `ListByOwnersTx` issues one query per level present: at S15-16's 30 Sources + 60 Artifacts + 300 graphs, three queries once S15-08 wires the levels, each served by the `UNIQUE (owner, field_id)` index. `Apply` reads the owner's values and layout once (two queries) and its fields once, and writes only differing rows; a 63-field draft is at most 63 value and 63 layout statements. No write's invalidation changes: the change rows are identical, so `effects` scopes, search docs and reads follow as today.

### Edges

- **Assumptions.** `listWorkspaces` and `apply` key stored values by field id: one value per (owner, field), held by `UNIQUE (source_id, field_id)` (`000009.sql`; S15-08's tables carry the same `UNIQUE`, map § 6.6). `Apply` keys the draft the same way and refuses duplicates itself (AC-4.8, T-3) rather than relying on the constraint, because a duplicate layout field would reach an `UPDATE` that the constraint never sees. `listWorkspaces` keys owners by `string(level)+string(id)`: the level strings differ and ids are 16 bytes, so two owners never share a key. The merge assumes one suggestion list per owner: owners of one type share the list read once, which is read-only.
- **Neighbors.** Equal after trim (AC-4.2) and equal after URL canonicalization (AC-4.3) write nothing; a value present in both with different text is an update, not a delete and insert (AC-4.4); a layout row differing only in `dismissed` or only in `sort_order` carries only that column beyond owner and field (AC-4.5, AC-4.6); a dismissed layout row on a valued field is written as given (AC-4.1, as `CreateAll` today), and `ListWorkspace` already shows a valued field whatever its flag. Two owners in one workspace read, one with layout and one without, each keep their own order rule (AC-3.4).
- **Removed inputs.** A value left out of the draft is deleted with `Clear`'s change shape, and its layout row stays (as `Clear` today), so the field keeps its place if refilled. A layout row left out of the draft stays (AC-4.7): a draft never undoes a dismiss by omission.
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** None — owners come from engine code, not from data the app did not write; a draft is bounded by the field vocabulary (63 at Scale) and duplicates are refused before any write.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcemetadata/owner.go` *(new)* | D-1, D-2, D-14 |
| `core/database/sourcemetadata/sourcemetadata.go` | D-3 – D-11. Package doc: "descriptive metadata values and layout for an owner: a Source today; an Artifact and an Evidence graph (S15-08) share the same functions". Function doc comments say "owner" where they said "Source" and keep each rule sentence |
| `core/database/sourcevocab/sourcevocab.go` | D-12 |
| `core/database/metadatafields/metadatafields.go` | D-13 |
| `core/database/sourcecopy/sourcecopy.go` | D-15 |
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

func listWorkspaces(tx *sql.Tx, owners []Owner) (map[string][]WorkspaceEntry, error)
func layoutChange(st *store, owner Owner, fieldID []byte, action rowchange.Action, diffs map[string]rowchange.FieldDiff) rowchange.Change
```

Errors: `sourcemetadata.invalid` → an unwired level, a bad id, a missing owner (reads and writes), a duplicate or unknown field, an empty value. `sources.not_found` for a missing Source stays on the wire through `GetSourceWorkspace`'s `sources.Get`.

### FFI

None — no proto change; handlers change call sites only.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy changes.

## 7. Steps

**STEP-1 — The owner and the Source store**
Files: `core/database/sourcemetadata/owner.go` *(new)*, `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`. Skill: `add-catalog-query`, `add-catalog-model`.
Do: D-1, D-2, D-14. Replace every SQL constant that names `source_metadata`, `source_metadata_layout`, `source_id` or `sources` with the `store` field of the same role (`sqlHasAny`, `sqlLayoutList`, `sqlListBySource` stay as constants until STEP-2 deletes them). The exported signatures do not change yet: each function uses `stores[vocabkey.LevelSource]`. Add `TestStoresAreRegistered` (T-8).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/...` passes, with the existing tests unchanged.

**STEP-2 — Every function takes an owner, and every caller passes one**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`, `core/database/sourcevocab/sourcevocab.go`, `core/database/sourcevocab/sourcevocab_test.go`, `core/database/metadatafields/metadatafields.go`, `core/database/metadatafields/metadatafields_test.go`, `core/database/sourcecopy/sourcecopy.go`, `core/database/sourcecopy/sourcecopy_test.go`, `core/database/deleteimpact/facets_test.go`, `api/ffi/handlers/sources.go`. Skill: `add-catalog-query`, `add-audit-scope`, `add-ffi-handler`.
Do: D-3 – D-6, D-8 (all but `updateLayoutTx`), D-10 (`listByOwners`, `listLayoutByOwners`), D-11, D-12, D-13, D-15, D-16, and the handler rows of §6 **Go core**. `CreateAll` takes `(tx, userID, owner, d Draft)` and keeps its body for now. Delete per §6 **Deletes**. Signatures and callers change in this one step so the tree builds at every commit.
Check: `rg -n "sourcemetadata\.(ListBySource|ListBySourceTx|ListLayoutTx)\b|sourcemetadata\.Input\{SourceID" --glob '*.go'` prints nothing; `CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...` and `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` pass, including T-1, T-6, T-7, T-9, T-10, T-11.

**STEP-3 — Apply and the batch read**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`.
Do: D-7 (with `updateLayoutTx`), D-9 (`CreateAll` now calls `apply`), `ListByOwnersTx` (D-10).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes with T-2, T-3, T-4, T-5 and every `TestCreateAll` and `sourcecopy` case.

**STEP-4 — Docs**
Files: `docs/source-layer-data-model.md` (§ 5.3: the DDL gains `UNIQUE (source_id, field_id)` and the unused `date_value_id BLOB REFERENCES date_values(id)` column as `000009.sql` has them, and the sentence "There is no `date_value_id` on this table" becomes "The `date_value_id` column is unused: no code reads or writes it, and catalog dates are text in `value_text`"; § 5.4 one sentence: `sourcemetadata` takes an owner, and `Apply` saves a draft of values and layout, writing layout rows verbatim and keeping rows the draft leaves out), `.agents/skills/add-catalog-query/SKILL.md` ("Reads inside a write": `sourcemetadata.ListLayoutTx` → `sourcemetadata.ReadDraftTx`), `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-07-metadata-operations-take-an-owner.md` reports no error; `scripts/check-pr.sh` prints `check-pr: PASS` on macOS (on another OS, report every check but `xcode-test` and `swift-warnings` passing, and say the Xcode checks were not run: this PR changes no Swift).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestSourceMetadata`, `TestCreateAll` (existing) | Every case but the deleted one passes with the helpers of D-16; assertions unchanged | AC-1.1, AC-2.2, AC-4.9 |
| T-2 | Go | `sourcemetadata_test.go` `TestApply` *(new)* | AC-4.1 (rows and change entity types and order equal `CreateAll`'s case "dismissed layout row on a valued field stays dismissed"); AC-4.2; AC-4.3; AC-4.4 (change actions and entity types in order, B's layout row still there); AC-4.5 and AC-4.6 (exact field keys); AC-4.7; AC-4.10; AC-4.11 | AC-4.1 – AC-4.7, AC-4.10, AC-4.11, D-7, D-8 |
| T-3 | Go | `sourcemetadata_test.go` `TestApply` *(new)* "refused drafts write nothing" | Repeated field in values; repeated field in layout; `"   "` value; unknown field id; owner id with no Source → each `ErrInvalid`, `source_metadata` and `source_metadata_layout` counts unchanged | AC-4.8 |
| T-4 | Go | `sourcemetadata_test.go` `TestListByOwnersTx` *(new)* | AC-3.1 (order by owner id, then field id; `Owner` on each row); AC-3.2; AC-3.3 | AC-3.1 – AC-3.3, D-10 |
| T-5 | Go | `sourcemetadata_test.go` `TestOwnerLevels` *(new)* | For each of `Set`, `Clear`, `DismissSuggestion`, `Reorder`, `Apply`, `CreateAll`, `ListWorkspace`, `ReadDraftTx`, `ListByOwnersTx`: owners at every level in `vocabkey.Levels()` with no `stores` entry, at `""` and `page`, and a `source` owner with a 15-byte id → `ErrInvalid`; for the writes, the `source_metadata` and `source_metadata_layout` counts are unchanged. `ReadDraftTx` and `ListByOwnersTx` with a nil `tx` → `ErrInvalid` | AC-2.1, D-2 |
| T-6 | Go | `sourcemetadata_test.go` `TestSourceMetadata` *(new cases)* "list workspace reads several extras" and "list workspace on a missing source is invalid" | AC-1.2 with three extra `text` fields made with `runFieldCreate`; a missing Source → `ErrInvalid` | AC-1.2, D-11 |
| T-7 | Go | `core/database/sourcecopy/sourcecopy_test.go`, `core/database/deleteimpact/facets_test.go`, `core/database/audit/audit_test.go`, `api/ffi/handlers/sources_test.go` (existing; `TestGetSourceWorkspace` "unknown source is not found" among them) | Pass with call-site changes only | AC-1.1, AC-1.3, D-15 |
| T-8 | Go | `sourcemetadata_test.go` `TestStoresAreRegistered` *(new, internal test)* | For each `stores` entry: `values` and `layout` are names in `catalogmodel.Tables`, and `catalogmodel.FKs` has an FK from each on `ownerColumn` | AC-2.3 |
| T-9 | Go | `sourcemetadata_test.go` `TestListWorkspaces` *(new, internal test)* | AC-3.4: two Sources of different types, one with a layout row → each list equals `ListWorkspace` for that Source alone | AC-3.4, D-11 |
| T-10 | Go | `core/database/metadatafields/metadatafields_test.go` `TestListByIDsTx` *(new)* | Two field ids plus a repeat and a missing id → the two fields; `nil` ids → `nil`, no error | D-13 |
| T-11 | Go | `core/database/sourcevocab/sourcevocab_test.go` `TestAppendSuggestion` (new case) | `ListSuggestionsTx` in a read transaction returns what `ListSuggestions` returns; level `citation` → `ErrInvalid` | D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/sourcecopy` (the snapshot compares layout and workspace of original and copy), `core/database/audit` `TestRecord` (counts `source_metadata` creates), `core/database/effects` scope tests, `api/ffi/handlers` `TestSetSourceMetadata` and `TestDismissAndReorderSourceMetadata`.

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
| A change shape drifts (a field renamed, an entity type built wrong) and audit scopes or search stop following | D-3 keeps the names; `layoutChange` always writes the owner column (D-8); T-1 keeps today's assertions; the `audit` and `effects` suites run in `check-pr.sh` |
| `Apply`'s changes for an empty owner differ from today's `CreateAll` in order | T-2's AC-4.1 compares entity types and order with `TestCreateAll`'s case; `sourcecopy`'s snapshot test compares the copy |
| Building SQL with `fmt.Sprintf` reaches data | `newStore` takes only the constant names in `stores`; no argument comes from a caller |
| `ListWorkspace` inside a write handler's read transaction waits on the write | Every handler calls it after `writes.Run` returns (`SetSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata`), so no transaction is open |

**Stop and report if:**

- `rg -ln "sourcemetadata\.[A-Z]" --glob '*.go'` lists a file other than `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`, `core/database/sourcecopy/sourcecopy.go`, `core/database/sourcecopy/sourcecopy_test.go` and `core/database/deleteimpact/facets_test.go`.
- A test in §4 needs an assertion changed (not a call site) other than the cases §6 **Deletes** names.
- `TestApply`'s AC-4.1 case cannot match `CreateAll`'s rows and change order without appending a layout row (D-7 is wrong).
- A handler test expects an error other than `sourcemetadata.ErrInvalid` on a metadata write, or `TestGetSourceWorkspace` "unknown source is not found" fails.
- The effects or audit suites fail on a `source_metadata` or `source_metadata_layout` change.
- `metadatafields` or `sourcevocab` cannot take the new Tx function without an import `sourcemetadata` already has as a dependent (an import cycle).

## 12. Open questions

None.

## 13. PR

**Title:** S15-07 — Metadata operations take an owner
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-07-metadata-operations-take-an-owner.md`.
