# S12-01 — Copy Source engine

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: engine, FFI and store method only; the form is S12-02 behind S12-D1)
**Board:** —
**Depends on:** —
**Blocks:** S12-02
**Branch:** `s12-01-copy-source-engine`
**Version bump:** none — no screen calls the new RPC yet, and Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [Copy Source](../source-structure.md#copy-source), working rules: "The form", "Copy lives in its own composing package, `core/database/sourcecopy`", "High tolerance for churn", "Each PR updates the docs and skills it makes false" (S12-01 line)
**Planned:** 2026-10-18 against `main` @ `f39e2e1`
**Review:** [`S12-01-copy-source-engine.review.md`](S12-01-copy-source-engine.review.md) — Fits with changes / Ready with changes, 1 round

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

The engine can copy a Source in one audited write: a new Source with a chosen title and type, and, each only when chosen, the original's description, metadata with its layout, credibility, and any subset of its Artifacts. This is the half of [Copy Source](../source-structure.md#copy-source) that the S12-02 form calls, so an obituary can start from a sibling's publication, issue, page and URL instead of retyping them.

**Why now.** Slice 1 is independent of the graph work and ships first. The engine lands before the form so S12-02 is Swift only.

**What stays.** Add Source, the Source page, and every existing RPC behave as before. Nothing in the app calls `CopySource` until S12-02.

## 2. User stories

**US-1.** As a researcher, when I copy an obituary with every part chosen, I get a new Source that already carries its type, description, metadata in the same layout, credibility and page scans, so I only change what differs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Original "Obituary of Ann Smith" (the test fixture in STEP-3: type `obituary`, description "Medicine Hat News", metadata publication = "Medicine Hat News", url = `https://example.org/a`, one dismissed suggestion, a hand-sorted layout, grade + argument "Family-placed notice", two Artifacts with Files) copied with title "Copy of Obituary of Ann Smith", the same type, every part on | The copy has a new `SRC-` ref, that title and type, the same description; `sourcemetadata.ListWorkspace` on the copy returns the same entries, values and order as on the original; the same grade and argument; two Artifacts with new `ART-` refs and the original's labels, descriptions and `file_id`s |
| AC-1.2 | AC-1.1, where the original's cover is one of its Artifacts | The copy's `cover_mode` is `type_icon` and `primary_artifact_id` is NULL |
| AC-1.3 | AC-1.1 | The original's row, notes, metadata, layout, credibility and Artifacts read back exactly as before the copy |
| AC-1.4 | AC-1.1 through `CopySource` over FFI | The response carries the copy and `WriteEffects` with `revision` > 0 and the copy's id in `source_scopes`; the latest audit revision's action is `copy_source` |

**US-2.** As a researcher, when I leave a part unchecked, the copy does not carry it, and the parts that are never copied never appear.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's original copied with description, metadata and credibility off and no Artifacts | The copy has an empty description, no `source_metadata` and no `source_metadata_layout` rows, no credibility row, no Artifacts |
| AC-2.2 | The original has two Artifacts; only the second is chosen | The copy has one Artifact, matching the second's label, description and `file_id` |
| AC-2.3 | The original has a note and a Subject | The copy has no `source_notes` row and no `subjects` row, and no `citations` row points at its Artifacts |
| AC-2.4 | AC-1.1 copied with a different Source type chosen | The copy has the chosen type and still carries every metadata value and layout row |

**US-3.** As a researcher, a copy that cannot be made writes nothing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Title is `"   "` | `sources.ErrInvalid` (`sources.invalid`); the Sources count is unchanged |
| AC-3.2 | The source id names no Source | `sources.ErrInvalid` |
| AC-3.3 | The chosen Artifact ids include one from another Source, or the same id twice | `sources.ErrInvalid`; nothing is written |
| AC-3.4 | The source type id names no type | `sources.ErrInvalid` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Transaction-scoped reads in the owning packages: `sources.GetTx`, `artifacts.ListBySourceTx`, `sourcecredibility.GetBySourceTx`, `sourcemetadata.ListBySourceTx`, `sourcemetadata.ListLayoutTx`; one layout row type and one layout scan in `sourcemetadata` | STEP-1, AC-1.1 |
| `sourcemetadata.CreateAll`: values and layout rows written verbatim onto a Source with none, sharing `Set`'s value insert | STEP-2, AC-1.1, AC-2.4 |
| `core/database/sourcecopy`: `Draft`, `DraftArtifact`, `ReadDraft`, `CreateFromDraft`, `CopyInput`, `Copy` | STEP-3, US-1, US-2, US-3 |
| FFI `METHOD_COPY_SOURCE`, `CopySourceRequest` / `CopySourceResponse`, handler `CopySource` | STEP-4, AC-1.4 |
| Swift `CoreMethod.copySource`, `GenealogyStore.copySource`, `GoStore` and `FakeStore` implementations | STEP-5 |
| `source-layer-data-model.md` § Copy Source; `add-catalog-query` skill: reads inside a write, composing packages | STEP-6 |

| Out | Home |
| --- | --- |
| The Copy form, the "Copy of" title prefill, entry points, opening the copy, the `createdSource` mutation | S12-02 |
| Graphs on the copy (a Default graph from `sources.Create`) | S12-03 |
| Copying notes, Subjects, Citations, Observations, graphs | not planned ([design](../source-structure.md#copy-source)) |
| A new error code or `L10n.Errors` entry for copy | not planned (D-8) |
| Importing a Source from a registry through `CreateFromDraft` | [registry source import](../../../ideas/registry-source-import.md) |
| Swift unit tests of `FakeStore.copySource` | S12-02 (its model tests drive it) |
| `CreateSourceResponse` gains `WriteEffects` and Add Source drops its write-through | not planned in this spike; raised to the user in the review log (M-9) |
| Other duplicate Catalog/Tx pairs outside the four packages STEP-1 names | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Source create | `core/database/sources/sources.go` `Create` | Inserts a Source with `cover_mode` default `type_icon` and no cover, mints `SRC-`, returns one `source` create change. Validates title (trimmed, non-empty) and type (`requireType`); bad input → `ErrInvalid` |
| Source read in a tx | `core/database/sources/sources.go` `getTx` | Unexported. Only `Get(c)` is exported, and it reads through `c.DB()` |
| Single connection | `core/database/catalog.go` `db.SetMaxOpenConns(1)`; `core/database/tx.go` `BeginWrite` | A read through `c.DB()` while a `writes.Run` transaction is open waits on the one connection. Reads inside a write must use the `*sql.Tx` |
| Metadata | `core/database/sourcemetadata/sourcemetadata.go` `Set`, `ListBySource`, `listLayout`, `ListWorkspace` | `Set` inserts a value and appends a layout row (`sqlLayoutAppend`); `ListBySource(c)` and `listLayout(c)` read through `c.DB()`. `ListWorkspace` orders by layout only when the Source has a layout row; older values can have no layout row (migration `000013.sql` has no backfill) |
| Layout reads | same file, `listLayout`, `getLayoutTx`, private type `layout` | Two scans of `source_metadata_layout` columns into a private `layout {SortOrder, Dismissed}` with no field id. `sqlLayoutList` has no `ORDER BY` |
| Field reads | same file, `getFieldTx`, `lookupField` | The same `sqlFieldGet` and scan, once over a Tx and once over `c.DB()` |
| Layout changes | same file, `DismissSuggestion`, `Reorder` | Changes are `EntityType: "source_metadata_layout"`, `EntityID: fieldID`, fields `source_id`, `field_id`, and `dismissed` or `sort_order` |
| Credibility | `core/database/sourcecredibility/assessments.go` `GetBySource`, `Upsert` | `GetBySource(c)` reads through `c.DB()`; `Upsert` creates the one row for a Source and validates the grade |
| Artifacts | `core/database/artifacts/artifacts.go` `Create`, `ListBySource` | `Create` takes `SourceID`, `FileID` (nil = fileless), `Label`, `Description`, mints `ART-`. `ListBySource(c)` reads through `c.DB()`, ordered by ref |
| Effects | `core/database/effects/registry.go` `registry` | `sources`, `source_metadata`, `source_metadata_layout`, `source_credibility_assessments`, `artifacts` already resolve their Source scope from `source_id`. No new entity type |
| Composing write precedent | `core/database/connect/connect.go` `CreateCitedBridge` | A package that composes several table packages in one `*database.Tx`, returning the changes for the caller's `writes.Run` |
| Handler | `api/ffi/handlers/sources.go` `CreateSource` | `writes.Run` with `Action: "create_source"`, `enrichSourceProto`, no `WriteEffects` on the response |
| Dispatch | `api/ffi/dispatch.go` `MethodCreateSource`, the `switch` | One constant and one `case` per method |
| Proto | `api/proto/engine.proto` `enum Method` | Highest value `METHOD_SET_PREFERRED_LIKENESS = 101`. `WriteEffects` message exists |
| Swift store | `macos/App/Platform/GenealogyStore.swift` protocol; `macos/App/Platform/GoStore.swift` `createArtifact`; `macos/App/Platform/CoreInvoke.swift` `CoreMethod` | Flat-parameter store methods; a write returns `(value, CatalogWriteEffects)` via `Self.mapWriteEffects` |
| Fake store | `macos/App/Platform/FakeStore.swift` `createSource`, `createArtifact`, `bumpSource` | Dictionaries per Source: `sourcesByProject`, `metadataBySource`, `credibilityBySource`, `artifactsBySource` |
| Querier precedent | `core/database/sourceeventtitles/sourceeventtitles.go` `Querier`, `core/database/sqlbatch/sqlbatch.go` `Querier` | A named per-package `type Querier interface { Query(...) }` that both `*sql.DB` and `*sql.Tx` satisfy |
| Toolchain | `scripts/generate-proto.sh`, `scripts/check-pr.sh` | Need `protoc`, `protoc-gen-go`, `protoc-gen-swift` and Xcode (macOS). Generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` are committed and never hand-edited |
| Tests | `core/database/sources/sources_test.go` `TestSources`, `TestSourceDelete`; `core/database/sourcemetadata/sourcemetadata_test.go`; `api/ffi/handlers/sources_test.go`; `api/ffi/handlers/harness_test.go` `runRPC`, `sourceFixture`, `assertLatestAuditAction` | Fixture helpers to copy: `TestSourceDelete` seeds a Subject (`subjectvocab.Install`, `subjects.Create`) |

**Precedent.** `core/database/connect` (`CreateCitedBridge`, called from `api/ffi/handlers/connect.go` inside `writes.Run`) is the same shape: a composing package under `core/database` that takes the `*database.Tx`, calls each owner package's write, and returns `(Result, []rowchange.Change, error)`. Copy its signature style and its handler call. Do not copy its private scope checks; validation stays in each owner's `Create` / `Upsert`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/database/sourcecopy` *(new)* exposes `Draft`, `DraftArtifact`, `ReadDraft(tx *sql.Tx, sourceID []byte) (Draft, error)`, `CopyInput`, `CreateFromDraft(tx *database.Tx, userID []byte, d Draft) (sources.Source, []rowchange.Change, error)`, and `Copy(tx *database.Tx, userID []byte, in CopyInput) (sources.Source, []rowchange.Change, error)` = read, apply the input, create. `Copy` returns `sources.Create`'s `Source` unchanged (`Updated` and `HasArtifact` unset; the handler fills `Updated`). `Copy` checks `tx != nil` and `database.RequireUserID(userID, sources.ErrInvalid)` before reading | Working rule "Copy lives in its own composing package"; `CreateFromDraft` is the slot registry import reuses | Copy inside `sources` — `sources` cannot import `sourcemetadata` (cycle) |
| D-2 | The draft is read inside the write transaction through new exported Tx reads in each owner package: `sources.GetTx`, `artifacts.ListBySourceTx`, `sourcecredibility.GetBySourceTx`, `sourcemetadata.ListBySourceTx`, `sourcemetadata.ListLayoutTx`. In `artifacts`, `sourcecredibility` and `sourcemetadata`, each Catalog read and its Tx twin call one unexported function that takes a package-level `type querier interface { Query(string, ...any) (*sql.Rows, error); QueryRow(string, ...any) *sql.Row }` *(new, unexported, one per package)*, so the SQL and scan are written once. The Tx twins check `len(id) == 16` and return their package's `ErrInvalid`, as the Catalog reads do. `sourcemetadata.lookupField` and `getFieldTx` fold into one `getField(q querier, fieldID)` | One connection: a `c.DB()` read inside `writes.Run` blocks. Reading in the tx also makes the copy consistent with the original at write time. The SQL stays in the package that owns the table | SQL for other tables written in `sourcecopy` — splits ownership; reading the draft before `writes.Run` — not atomic with the write |
| D-3 | `sources.GetTx(tx *sql.Tx, id []byte)` *(new)* checks `len(id) == 16` (else `ErrInvalid`) and calls the private `getTx`, exactly like `subjects.GetTx`. In-package callers keep `getTx` | Every exported Get rejects a bad id the same way (`sources.Get`, `subjects.GetTx`) | Renaming `getTx` — drops the length check |
| D-4 | Metadata is written by `sourcemetadata.CreateAll(tx *database.Tx, userID, sourceID []byte, values []Input, layout []LayoutRow) ([]rowchange.Change, error)` *(new)*. Empty `values` and `layout` → nil, nil. Otherwise: `requireSource`; refuse with `ErrInvalid` when the Source already has a `source_metadata` or `source_metadata_layout` row (`sqlHasAny` *(new)*: `SELECT EXISTS (SELECT 1 FROM source_metadata WHERE source_id = ?) OR EXISTS (SELECT 1 FROM source_metadata_layout WHERE source_id = ?)`); refuse with `ErrInvalid` when an `Input.SourceID` is not `sourceID`. Each value goes through `insertValueTx` (D-6); an empty `ValueText` → `ErrInvalid` (from `normalizeValue`). Then each layout row is inserted verbatim with `sqlLayoutInsert` *(new)*: `INSERT INTO source_metadata_layout (source_id, field_id, sort_order, dismissed) VALUES (?, ?, ?, ?)`, after `getField` (unknown field → `ErrInvalid`). A layout row for a field with no value is allowed. Changes: all value changes in input order, then all layout changes in input order | Only verbatim rows make `ListWorkspace` on the copy equal the original's: `Set` appends layout rows the original may not have, and `DismissSuggestion` skips a field that holds a value | `Set` per value then `Reorder` and `DismissSuggestion` — changes order for values with no layout row and loses a dismissed flag on a valued field |
| D-5 | `LayoutRow` *(new)* is `struct { FieldID []byte; SortOrder int; Dismissed bool }` and replaces the private `layout` type. One unexported `listLayoutRows(q querier, sourceID) ([]LayoutRow, error)` runs `sqlLayoutList` with `ORDER BY sort_order, field_id` added. `ListLayoutTx` returns it; `listLayout` builds its map and max order from it (same results for `ListWorkspace`); `getLayoutTx` returns `LayoutRow` | One row type and one scan for the table, so a new layout column reaches every read | A second layout type and scan beside `layout` (M-1) |
| D-6 | `Set`'s create branch moves into `insertValueTx(tx *sql.Tx, sourceID []byte, field metadatafields.Field, valueText string) (Row, rowchange.Change, error)` *(new, unexported)*: trim, `normalizeValue`, `sqlInsert`, the `source_metadata` create change. `Set` calls it and then `sqlLayoutAppend`; `CreateAll` calls it with no layout append. The layout change in `CreateAll` is `EntityType: "source_metadata_layout"`, `EntityID: fieldID`, `Action: create`, fields `source_id`, `field_id`, `sort_order`, `dismissed` (`Old: nil`) | Matches today's layout changes, so `effects` resolves the Source scope with no registry change ([`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md)) | A new entity type |
| D-7 | `CopyInput` is `struct { SourceID []byte; Title string; SourceTypeID []byte; Description, Metadata, Credibility bool; ArtifactIDs [][]byte }`. Applying it: title and type always come from `CopyInput`, passed untrimmed to `sources.Create`, which trims and validates (nil, short or unknown type → `sources.ErrInvalid`); description, metadata + layout, and credibility are kept only when their flag is on; Artifacts are the draft's Artifacts whose id is in `ArtifactIDs`, in the draft's ref order. An id not among the original's Artifacts, or a repeated id, → `sources.ErrInvalid` before any write. `Credibility` on while the original has none → no assessment, no error | Matches the form in working rule "The form"; one input struct per composing write, like `connect.CreateInput` | A separate `Choices` type beside a `sourceID` argument (M-7) |
| D-8 | No new error code. `ReadDraft` maps `sql.ErrNoRows` from `sources.GetTx` to `sources.ErrInvalid` and treats `sql.ErrNoRows` from `sourcecredibility.GetBySourceTx` as no assessment. Missing original, empty title, unknown type, a foreign or repeated Artifact id → `sources.ErrInvalid`; metadata and credibility failures keep their owners' codes | `sources.invalid` is already mapped under `L10n.Errors`; none of these is reachable from a correct form | `source_copy.invalid` *(rejected)* — a code plus L10n for states the UI never produces |
| D-9 | `Draft` is built from the owners' input types (M-3): `Source sources.CreateInput`, `Metadata []sourcemetadata.Input`, `Layout []sourcemetadata.LayoutRow`, `Credibility *sourcecredibility.UpsertInput`, `Artifacts []DraftArtifact` with `DraftArtifact { FromID []byte; In artifacts.CreateInput }` (`FromID` is the original's id, used only to match `CopyInput.ArtifactIDs`). `ReadDraft` leaves every `SourceID` nil. `CreateFromDraft` order: `sources.Create(d.Source)` → set `SourceID` on each input → `sourcemetadata.CreateAll` → `sourcecredibility.Upsert` when `Credibility != nil` → `artifacts.Create` per Artifact. Changes are concatenated in that order | Each owner validates its own rows; Artifacts last so a failed metadata write aborts before any Artifact | Inserting Artifacts first — no benefit |
| D-10 | The cover is never copied: `sources.Create` already leaves `cover_mode = type_icon`, `primary_artifact_id` NULL; `CreateFromDraft` calls no `SetCover` | Design: "The copy's cover is reset" | Copying the cover to the matching new Artifact |
| D-11 | `Draft` carries no notes, Subjects, Citations, Observations or graphs; `ReadDraft` does not read them | Design: never copied | Reading and dropping them |
| D-12 | One `writes.Run` with `writes.Op{Action: "copy_source", UserID: userID}` in the handler; the closure calls `sourcecopy.Copy` | Working rule: one transaction, action `copy_source`. Audit actions are named in handlers (`create_source`, `create_cited_bridge`) | Running the closure from `sourcecopy` — packages under `core/database` do not call `writes.Run` |
| D-13 | Proto: `METHOD_COPY_SOURCE = 102`; `CopySourceRequest { project_dir = 1; user_id = 2; source_id = 3; title = 4; source_type_id = 5; copy_description = 6; copy_metadata = 7; copy_credibility = 8; repeated string artifact_ids = 9; }`; `CopySourceResponse { Source source = 1; WriteEffects effects = 2; }`. Handler `CopySource` lives in `api/ffi/handlers/sources.go` beside `CreateSource`; the response source goes through `enrichSourceProto` with `Updated = res.Activity` | Effects on the response let S12-02 patch `source_scopes` with no write-through (working rule "One title and type form") | No `WriteEffects`, as `CreateSourceResponse` — S12-02 would copy `create()`'s write-through, which the rule forbids |
| D-14 | Swift: `GenealogyStore.copySource(projectDir: String, userID: String, sourceID: String, title: String, sourceTypeID: String, copyDescription: Bool, copyMetadata: Bool, copyCredibility: Bool, artifactIDs: [String]) async throws -> (CatalogSource, CatalogWriteEffects)` *(new)*, placed after `createSource` in the protocol, `GoStore`, and `FakeStore` | Flat parameters like every store write | A new choices struct — no second caller |
| D-15 | `FakeStore.copySource` appends a new `CatalogSource` (new lowercase UUID id, ref `"SRC-FAKE1"`, chosen type and title, original description when chosen, else `""`), copies `metadataBySource` when `copyMetadata`, copies `credibilityBySource` with `id` new and `sourceID` = the copy's id when `copyCredibility` and the original has one, copies each chosen `CatalogArtifact` with a new lowercase UUID id, ref `"ART-FAKE1"`, `sourceID` = the copy's id, the same `fileID`, `label`, `description`, `file`, `thumbnailRelPath`; sets the copy's `hasArtifact` to whether any Artifact was copied and leaves its cover fields at their defaults; copies metadata entries as they are (stale `suggested` flags under another type are accepted in the fake); throws `StoreBoom.boom` when the original is not in `sourcesByProject[projectDir]`; returns `bumpSource(copyID, evidence: false)` | S12-02's model tests need a fake that behaves like the engine | A fake that returns a bare Source |
| D-16 | STEP-4 and STEP-5 run only where `protoc`, `protoc-gen-go`, `protoc-gen-swift` and Xcode are installed (the macOS dev machine `scripts/check-pr.sh` targets). Generated files are produced by `./scripts/generate-proto.sh` only | Generated code is committed; a hand edit drifts from the generator | Hand-editing `engine.pb.go` / `engine.pb.swift` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/sourcecopy` *(new)* | package | `core/database/connect` | Different write; same composing shape (precedent) |
| `sourcecopy.Draft`, `DraftArtifact`, `CopyInput` *(new)* | type | `sources.CreateInput`, `sourcemetadata.Input`, `sourcecredibility.UpsertInput`, `artifacts.CreateInput` | Built from those inputs (D-9); `CopyInput` mirrors `connect.CreateInput` |
| `sourcecopy.ReadDraft`, `CreateFromDraft`, `Copy` *(new)* | func | `connect.CreateCitedBridge` | — |
| `sources.GetTx` *(new)* | func | `sources.getTx`, `subjects.GetTx` | Length-checking wrapper, as `subjects.GetTx` (D-3) |
| `artifacts.ListBySourceTx` *(new)* | func | `artifacts.ListBySource` | Shares its SQL and scan (D-2) |
| `sourcecredibility.GetBySourceTx` *(new)* | func | `sourcecredibility.GetBySource` | Shares `sqlGetBySource` (D-2) |
| `sourcemetadata.ListBySourceTx` *(new)* | func | `sourcemetadata.ListBySource` | Shares `sqlListBySource` (D-2) |
| `sourcemetadata.ListLayoutTx`, `LayoutRow`, `listLayoutRows` *(new)* | func, type | `sourcemetadata.listLayout`, private `layout` | `LayoutRow` replaces `layout`; `listLayout` builds on `listLayoutRows` (D-5) |
| `sourcemetadata.insertValueTx`, `getField` *(new)* | func | `Set` create branch; `getFieldTx`, `lookupField` | Extracted and merged, not copied (D-2, D-6) |
| `querier` *(new)* in `artifacts`, `sourcecredibility`, `sourcemetadata` | type | `sourceeventtitles.Querier`, `sqlbatch.Querier` | Unexported, one per package, like the neighbors |
| `sourcemetadata.CreateAll` *(new)* | func | `sourcemetadata.Set`, `Reorder`, `DismissSuggestion` | Those cannot reproduce a layout verbatim (D-4); shares `insertValueTx` with `Set` |
| `METHOD_COPY_SOURCE`, `CopySourceRequest`, `CopySourceResponse` *(new)* | FFI | `METHOD_CREATE_SOURCE` | Different input |
| `handlers.CopySource` *(new)* | func | `handlers.CreateSource` | — |
| `CoreMethod.copySource`, `GenealogyStore.copySource` *(new)* | Swift | `createSource` | — |
| `TestCopy` *(new)*, `TestCreateAll` *(new)*, `TestCopySource` *(new)* | test | `TestCreateCitedBridge`, `TestSources` | — |

**Builds on.** `add-catalog-query` owner packages (each table's SQL stays in its package), the `connect` composing-package pattern, `writes.Run` and the existing effects registry entries, `add-ffi-handler` with `writeEffectsProto`.
**Deletes.** The private `layout` type, `sourcemetadata.getFieldTx` and `lookupField` (merged into `getField`), the duplicated scan loops in the four packages' Catalog reads.
**Next adds.** S12-02 calls `store.copySource`. S12-03 makes `sources.Create` mint the Default graph, which every copy then gets with no change here. Registry import builds a `Draft` and calls `CreateFromDraft`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sources` | `func GetTx(tx *sql.Tx, id []byte) (Source, error)` *(new)* (D-3) |
| `core/database/artifacts` | `func ListBySourceTx(tx *sql.Tx, sourceID []byte) ([]Artifact, error)` *(new)*; `ListBySource` and it share `listBySource(q, sourceID)` |
| `core/database/sourcecredibility` | `func GetBySourceTx(tx *sql.Tx, sourceID []byte) (Assessment, error)` *(new)*, `sql.ErrNoRows` when none |
| `core/database/sourcemetadata` | `func ListBySourceTx(tx *sql.Tx, sourceID []byte) ([]Row, error)` *(new)*; `type LayoutRow` *(new)*; `func ListLayoutTx(tx *sql.Tx, sourceID []byte) ([]LayoutRow, error)` *(new)*; `func CreateAll(...)` *(new)*; `insertValueTx`, `getField`, `listLayoutRows` *(new, unexported)* per D-2, D-4 – D-6 |
| `core/database/sourcecopy` *(new)* | `sourcecopy.go` *(new)* per D-1, D-7, D-9 – D-11 |

```go
// Draft is everything a copy can carry from a Source.
type Draft struct {
	Source      sources.CreateInput
	Metadata    []sourcemetadata.Input
	Layout      []sourcemetadata.LayoutRow
	Credibility *sourcecredibility.UpsertInput // nil when the Source has none
	Artifacts   []DraftArtifact                // the original's ref order
}

// DraftArtifact is one Artifact to create; FromID is the Artifact it was read from.
type DraftArtifact struct {
	FromID []byte
	In     artifacts.CreateInput
}
```

Errors: `sources.invalid` → D-8 cases. `sourcemetadata.invalid` → `CreateAll` on a Source that already has rows (not reachable from `Copy`). `sourcecredibility.invalid` → a grade that no longer exists.

### FFI

`api/proto/engine.proto`: `METHOD_COPY_SOURCE = 102` *(new)*, `CopySourceRequest` *(new)*, `CopySourceResponse` *(new)* per D-13. Handler `CopySource` *(new)* in `api/ffi/handlers/sources.go`: parse `user_id` (`parseUserID`), `source_id` and `source_type_id` (`parseID`), each `artifact_ids` entry (`parseID`), returning a parse error before opening the catalog, as `CreateSource` does; `withProjectCatalog`; `writes.Run` per D-12; response `Source` + `Effects: writeEffectsProto(wrote)`. `api/ffi/dispatch.go`: `MethodCopySource` *(new)* and its `case`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view change. Store only: `CoreMethod.copySource` *(new)* in `macos/App/Platform/CoreInvoke.swift`; `copySource` in `macos/App/Platform/GenealogyStore.swift` (protocol), `macos/App/Platform/GoStore.swift` (builds `Provenencia_Engine_V1_CopySourceRequest`, returns `(Self.mapSource(resp.source), Self.mapWriteEffects(resp.effects))`), and `macos/App/Platform/FakeStore.swift` (D-15).

### Strings

None — no user-facing copy and no new error code (D-8).

## 7. Steps

**STEP-1 — Transaction reads in the owner packages**
Files: `core/database/sources/sources.go`, `core/database/artifacts/artifacts.go`, `core/database/sourcecredibility/assessments.go`, `core/database/sourcemetadata/sourcemetadata.go`. Skill: `add-catalog-query`.
Do: D-2, D-3, D-5. Add `sources.GetTx`. Add the `querier` type and Tx reads; route each Catalog read through the shared unexported function. Replace `layout` with `LayoutRow`; merge `getFieldTx` and `lookupField` into `getField`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes; `rg -n 'func lookupField|func getFieldTx|type layout struct' core/database/sourcemetadata` prints nothing.

**STEP-2 — `sourcemetadata.CreateAll`**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`. Skills: `add-catalog-query`, `add-audit-scope`.
Do: D-6 first (`Set` now calls `insertValueTx`; existing `Set` tests pass unchanged), then D-4. Add T-8, using the package's existing `runSet`, `runFieldCreate` and `runDismissSuggestion` helpers.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/... -run TestCreateAll -v` passes.

**STEP-3 — `sourcecopy`**
Files: `core/database/sourcecopy/sourcecopy.go` *(new)*, `core/database/sourcecopy/sourcecopy_test.go` *(new)*.
Do: D-1, D-7 – D-11. Add T-1 – T-7 (one table, `TestCopy`).
Test fixture (one helper, `seedObituary(t) (c *database.Catalog, userID []byte, orig sources.Source, arts []artifacts.Artifact)` *(new)* in `sourcecopy_test.go`): `database.Create(t.TempDir(), "t.provenencia")`; `users.Upsert` with a minted `USR-` ref (as `connect_test.go`); `sourcetypes.Upsert` keys `obituary` and `newspaper` (origin `vocabkey.OriginUser`); `metadatafields.Upsert` fields `publication` (`metadatafields.DataTypeText`), `url` (`DataTypeURL`) and `page` (`DataTypeText`); `sourcevocab.EnsureSuggestion(obituary, publication, 0)`, `(obituary, url, 1)`, `(obituary, page, 2)`; `sourcecredibilitygrades.Install` then the first of `sourcecredibilitygrades.List`; the Source through `sources.Create` in `writes.Run`; metadata through `sourcemetadata.Set` (publication, url), `DismissSuggestion` (page) and `Reorder` (url, publication, page); credibility through `sourcecredibility.Upsert`; two Files through `ingest.File` on temp files with different bytes (as `mustFile` in `artifacts_test.go`) and two `artifacts.Create`; cover through `sources.NormalizeCover` + `SetCover` on the first Artifact only in the cover case (needs a real PNG: copy `writeTinyPNG` from `sources_test.go`); a note through `sources.AddNote`; a Subject the way `TestSourceDelete` seeds one. Run each copy through `writes.Run` with `Action: "copy_source"`. Count Sources with `sources.Count`, notes with `sources.ListNotes`, other rows with `c.DB()` queries outside the write.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecopy/... -v` passes.

**STEP-4 — FFI `CopySource`**
Files: `api/proto/engine.proto`, generated Go and Swift from `./scripts/generate-proto.sh`, `api/ffi/dispatch.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-12, D-13, D-16. Add T-9. Its fixture seeds through RPCs: `sourceFixture`, `CreateSource`, then `CreateArtifact` (fileless, label "Scan").
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-5 — Swift store method**
Files: `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`.
Do: D-14, D-15.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes.

**STEP-6 — Docs**
Files: `docs/source-layer-data-model.md`, `.agents/skills/add-catalog-query/SKILL.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: add `## 4.2 Copy Source` after `## 4.1 source_notes` in `docs/source-layer-data-model.md`: what is copied when chosen (description; metadata values and layout rows verbatim; credibility; each chosen Artifact as a new row with the same label, description and File), what is never copied (notes, Subjects, Citations, Observations), the cover reset, one audited `copy_source` write, and the two-tools rule linked to [`source-structure.md`](../source-structure.md#the-two-tools). Add a section **Reads inside a write** to `add-catalog-query`: the catalog has one connection, so a read inside `writes.Run` takes the `*sql.Tx` through an exported `…Tx` twin that shares one unexported function over the package `querier`; a write across several tables lives in a composing package (`core/database/connect`, `core/database/sourcecopy`) that takes the `*database.Tx` and never calls `writes.Run`. Then the Definition of done: plan Status `Landed in #N (date)` with a Landed note; map row status `Landed #N`; a `## S12-01 — Copy Source engine` paragraph at the top of `completed.md`, under the intro line, in the Spike 11 `completed.md` style.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-01-copy-source-engine.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcecopy/sourcecopy_test.go` `TestCopy` *(new)*, table-driven | "all parts": AC-1.1 fixture → new ref ≠ original's; title, type, description; `ListWorkspace` equal by `Field.ID`, `Value.ValueText`, `SortOrder`, `Suggested` in order (a local `workspaceKeys` helper maps entries to comparable structs); grade + argument; Artifacts' labels, descriptions, `FileID` equal and refs new; cover `type_icon` + nil | AC-1.1, AC-1.2, D-4, D-9, D-10 |
| T-2 | Go | `TestCopy` | "original untouched": snapshot the original's row, notes, `ListWorkspace`, layout, credibility, Artifacts before; equal after | AC-1.3 |
| T-3 | Go | `TestCopy` | "nothing chosen" → empty description; zero metadata, layout, credibility, Artifact rows for the copy | AC-2.1 |
| T-4 | Go | `TestCopy` | "second Artifact only" → one Artifact matching the second | AC-2.2, D-7 |
| T-5 | Go | `TestCopy` | "note and Subject not copied" → zero `source_notes`, `subjects` rows for the copy; zero `citations` on its Artifacts | AC-2.3, D-11 |
| T-6 | Go | `TestCopy` | "other type" → chosen type; metadata and layout still equal | AC-2.4 |
| T-7 | Go | `TestCopy` | "blank title", "unknown source", "foreign Artifact", "repeated Artifact", "unknown type" → `errors.Is(err, sources.ErrInvalid)`; Sources count unchanged | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-8 |
| T-8 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestCreateAll` *(new)*, table-driven | values with no layout rows → values present, zero layout rows; a dismissed layout row on a valued field → kept dismissed; a URL value `HTTPS://Example.org/a` → the form `Set` stores; a Source with an existing value → `ErrInvalid`; an `Input.SourceID` of another Source → `ErrInvalid`; an unknown field id → `ErrInvalid`; empty inputs → no changes; `ListLayoutTx` returns rows by `sort_order` | D-4, D-5, D-6 |
| T-9 | Go | `api/ffi/handlers/sources_test.go` `TestCopySource` *(new)*, `runRPC` | bad proto → error; bad `source_id` → error; ok copy of a Source with one Artifact (all flags on) → response title, type, `SRC-` ref, `effects.revision` > 0, copy id in `effects.source_scopes`, latest audit action `copy_source` (`assertLatestAuditAction`); a malformed `artifact_ids` entry → error; a well-formed id of another Source's Artifact → error with code `sources.invalid` | AC-1.4, D-12, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/sources` tests (the `getTx` rename), `core/database/effects` `registry_completeness_test.go` (if a change shape is wrong), Swift build (protocol conformance of `FakeStore` and `GoStore`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added (D-8)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/source-layer-data-model.md § 4.2 Copy Source
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — no screen calls `CopySource` until S12-02. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecopy/... ./api/ffi/handlers/ -run 'TestCopy|TestCopySource' -v` and reads the "all parts" and "nothing chosen" cases' output.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Catalog read slips into the copy transaction and hangs the test | D-2: every draft read takes the `*sql.Tx`; T-1 – T-7 run through `writes.Run` |
| `CreateAll` change shape does not resolve a Source scope | D-6 copies today's shapes; `registry_completeness_test.go` and T-9's `source_scopes` assertion catch it |
| S12-03 lands first and `sources.Create` starts minting a Default graph | No change needed here: the copy gets its Default graph through `sources.Create`, and S12-03 owns the graph tests |

**Stop and report if:**

- `sources.Create`, `sourcemetadata.Set`, `sourcecredibility.Upsert` or `artifacts.Create` no longer has the signature in §4.
- `protoc`, `protoc-gen-go`, `protoc-gen-swift` or Xcode is missing where you run STEP-4 or STEP-5 (D-16). Finish STEP-1 – STEP-3 and STEP-6's docs, report the Go evidence, and do not hand-edit generated files.
- `METHOD_COPY_SOURCE = 102` collides with a value on `main`; report the free number you would use.
- `ListWorkspace` on a copy made with `CreateAll` differs from the original's for the same type (T-1 fails on order or `Suggested`), which would mean layout is not the whole story of field order.
- `effects.Resolve` rejects a `CreateAll` change (an error naming `source_metadata_layout`).
- `sources.ErrInvalid` is not mapped under `L10n.Errors` in `macos/App/Platform/L10n.swift` (`rg -n 'sources.invalid' macos/App`).

## 12. Open questions

None.

## 13. PR

**Title:** S12-01 — Copy Source engine
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-01-copy-source-engine.md`.
