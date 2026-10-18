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

The engine can copy a Source in one audited write: a new Source with a chosen title and type, and, each only when chosen, the original's description, metadata with its layout, credibility, and any subset of its Artifacts. This is the half of [Copy Source](../source-structure.md#copy-source) that the S12-02 form calls, so an obituary can start from a sibling's publication, issue, page and URL instead of retyping them.

**Why now.** Slice 1 is independent of the graph work and ships first. The engine lands before the form so S12-02 is Swift only.

**What stays.** Add Source, the Source page, and every existing RPC behave as before. Nothing in the app calls `CopySource` until S12-02.

## 2. User stories

**US-1.** As a researcher, when I copy an obituary with every part chosen, I get a new Source that already carries its type, description, metadata in the same layout, credibility and page scans, so I only change what differs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Original "Obituary of Ann Smith" (type obituary, description "Medicine Hat News", metadata publication = "Medicine Hat News", url = `https://example.org/a`, a dismissed suggestion, a hand-sorted layout, grade + argument "Family-placed notice", two Artifacts with Files) copied with title "Copy of Obituary of Ann Smith", the same type, every part on | The copy has a new `SRC-` ref, that title and type, the same description; `sourcemetadata.ListWorkspace` on the copy returns the same entries, values and order as on the original; the same grade and argument; two Artifacts with new `ART-` refs and the original's labels, descriptions and `file_id`s |
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
| Transaction-scoped reads in the owning packages: `sources.GetTx`, `artifacts.ListBySourceTx`, `sourcecredibility.GetBySourceTx`, `sourcemetadata.ListBySourceTx`, `sourcemetadata.ListLayoutTx` | STEP-1, AC-1.1 |
| `sourcemetadata.CreateAll`: values and layout rows written verbatim onto a Source with none | STEP-2, AC-1.1, AC-2.4 |
| `core/database/sourcecopy`: `Draft`, `ReadDraft`, `CreateFromDraft`, `Choices`, `Copy` | STEP-3, US-1, US-2, US-3 |
| FFI `METHOD_COPY_SOURCE`, `CopySourceRequest` / `CopySourceResponse`, handler `CopySource` | STEP-4, AC-1.4 |
| Swift `CoreMethod.copySource`, `GenealogyStore.copySource`, `GoStore` and `FakeStore` implementations | STEP-5 |
| `source-layer-data-model.md` § Copy Source | STEP-6 |

| Out | Home |
| --- | --- |
| The Copy form, the "Copy of" title prefill, entry points, opening the copy, the `createdSource` mutation | S12-02 |
| Graphs on the copy (a Default graph from `sources.Create`) | S12-03 |
| Copying notes, Subjects, Citations, Observations, graphs | not planned ([design](../source-structure.md#copy-source)) |
| A new error code or `L10n.Errors` entry for copy | not planned (D-8) |
| Importing a Source from a registry through `CreateFromDraft` | [registry source import](../../../ideas/registry-source-import.md) |
| Swift unit tests of `FakeStore.copySource` | S12-02 (its model tests drive it) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Source create | `core/database/sources/sources.go` `Create` | Inserts a Source with `cover_mode` default `type_icon` and no cover, mints `SRC-`, returns one `source` create change. Validates title (trimmed, non-empty) and type (`requireType`); bad input → `ErrInvalid` |
| Source read in a tx | `core/database/sources/sources.go` `getTx` | Unexported. Only `Get(c)` is exported, and it reads through `c.DB()` |
| Single connection | `core/database/catalog.go` `db.SetMaxOpenConns(1)`; `core/database/tx.go` `BeginWrite` | A read through `c.DB()` while a `writes.Run` transaction is open waits on the one connection. Reads inside a write must use the `*sql.Tx` |
| Metadata | `core/database/sourcemetadata/sourcemetadata.go` `Set`, `ListBySource`, `listLayout`, `ListWorkspace` | `Set` inserts a value and appends a layout row (`sqlLayoutAppend`); `ListBySource(c)` and `listLayout(c)` read through `c.DB()`. `ListWorkspace` orders by layout only when the Source has a layout row; older values can have no layout row (migration `000013.sql` has no backfill) |
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
| Tests | `core/database/sources/sources_test.go` `TestSources`, `TestSourceDelete`; `core/database/sourcemetadata/sourcemetadata_test.go`; `api/ffi/handlers/sources_test.go`; `api/ffi/handlers/harness_test.go` `runRPC`, `sourceFixture`, `assertLatestAuditAction` | Fixture helpers to copy: `TestSourceDelete` seeds a Subject (`subjectvocab.Install`, `subjects.Create`) |

**Precedent.** `core/database/connect` (`CreateCitedBridge`, called from `api/ffi/handlers/connect.go` inside `writes.Run`) is the same shape: a composing package under `core/database` that takes the `*database.Tx`, calls each owner package's write, and returns `(Result, []rowchange.Change, error)`. Copy its signature style and its handler call. Do not copy its private scope checks; validation stays in each owner's `Create` / `Upsert`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/database/sourcecopy` *(new)* exposes `Draft`, `ReadDraft(tx *sql.Tx, sourceID []byte) (Draft, error)`, `Choices`, `CreateFromDraft(tx *database.Tx, userID []byte, d Draft) (sources.Source, []rowchange.Change, error)`, and `Copy(tx *database.Tx, userID []byte, sourceID []byte, ch Choices) (sources.Source, []rowchange.Change, error)` = read, apply choices, create | Working rule "Copy lives in its own composing package"; `CreateFromDraft` is the slot registry import reuses | Copy inside `sources` — `sources` cannot import `sourcemetadata` (cycle) |
| D-2 | The draft is read inside the write transaction through new exported Tx reads in each owner package: `sources.GetTx`, `artifacts.ListBySourceTx`, `sourcecredibility.GetBySourceTx`, `sourcemetadata.ListBySourceTx`, `sourcemetadata.ListLayoutTx`. Each Catalog read and its Tx twin share one unexported function over `interface{ Query(string, ...any) (*sql.Rows, error) }` (or `QueryRow`), so the SQL is written once | One connection: a `c.DB()` read inside `writes.Run` blocks. Reading in the tx also makes the copy consistent with the original at write time. The SQL stays in the package that owns the table | SQL for other tables written in `sourcecopy` — splits ownership; reading the draft before `writes.Run` — not atomic with the write |
| D-3 | `sources.GetTx` is the exported rename of `getTx` (same body); every in-package caller uses `GetTx` | One function, matches `subjects.GetTx` | A second exported wrapper beside `getTx` |
| D-4 | Metadata is written by `sourcemetadata.CreateAll(tx *database.Tx, userID, sourceID []byte, values []Row, layout []LayoutRow) ([]rowchange.Change, error)` *(new)*. It refuses (`sourcemetadata.ErrInvalid`) when the Source is missing or already has any `source_metadata` or `source_metadata_layout` row. It inserts each value (new UUIDv7 id, `normalizeValue` by field data type, no layout append) and each layout row verbatim (`sort_order`, `dismissed`). `Row.ID` and `Row.SourceID` in `values` are ignored | Only verbatim rows make `ListWorkspace` on the copy equal the original's: `Set` appends layout rows the original may not have, and `DismissSuggestion` skips a field that holds a value | `Set` per value then `Reorder` and `DismissSuggestion` — changes order for values with no layout row and loses a dismissed flag on a valued field |
| D-5 | `LayoutRow` *(new)* is `struct { FieldID []byte; SortOrder int; Dismissed bool }`. `ListLayoutTx` returns rows ordered by `sort_order`, then `field_id`. `listLayout` and `ListWorkspace` are unchanged | A slice in a fixed order is what a draft carries and a test compares | Exporting the private `layout` map type — keyed by string, no field id |
| D-6 | Value change shape in `CreateAll` equals `Set`'s create change; layout change is `EntityType: "source_metadata_layout"`, `EntityID: fieldID`, `Action: create`, fields `source_id`, `field_id`, `sort_order`, `dismissed` (`Old: nil`) | Matches today's layout changes, so `effects` resolves the Source scope with no registry change ([`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md)) | A new entity type |
| D-7 | `Choices` is `struct { Title string; SourceTypeID []byte; Description, Metadata, Credibility bool; ArtifactIDs [][]byte }`. Applying it: title and type always come from `Choices`; description, metadata + layout, and credibility are kept only when their flag is on; Artifacts are the draft's Artifacts whose id is in `ArtifactIDs`, in the draft's ref order. An id not among the original's Artifacts, or a repeated id, → `sources.ErrInvalid` before any write | Matches the form in working rule "The form" | Choices as a list of part names — the form has fixed parts |
| D-8 | No new error code. Missing original, empty title, unknown type, a foreign or repeated Artifact id → `sources.ErrInvalid`; metadata and credibility failures keep their owners' codes | `sources.invalid` is already mapped under `L10n.Errors`; none of these is reachable from a correct form | `source_copy.invalid` *(rejected)* — a code plus L10n for states the UI never produces |
| D-9 | `CreateFromDraft` order: `sources.Create` (type, title, description) → `sourcemetadata.CreateAll` when the draft has values or layout rows → `sourcecredibility.Upsert` when the draft has an assessment → `artifacts.Create` per Artifact (same `FileID`, `Label`, `Description`). Changes are concatenated in that order | Each owner validates its own rows; Artifacts last so a failed metadata write aborts before any Artifact | Inserting Artifacts first — no benefit |
| D-10 | The cover is never copied: `sources.Create` already leaves `cover_mode = type_icon`, `primary_artifact_id` NULL; `CreateFromDraft` calls no `SetCover` | Design: "The copy's cover is reset" | Copying the cover to the matching new Artifact |
| D-11 | `Draft` carries no notes, Subjects, Citations, Observations or graphs; `ReadDraft` does not read them | Design: never copied | Reading and dropping them |
| D-12 | One `writes.Run` with `writes.Op{Action: "copy_source", UserID: userID}` in the handler; the closure calls `sourcecopy.Copy` | Working rule: one transaction, action `copy_source`. Audit actions are named in handlers (`create_source`, `create_cited_bridge`) | Running the closure from `sourcecopy` — packages under `core/database` do not call `writes.Run` |
| D-13 | Proto: `METHOD_COPY_SOURCE = 102`; `CopySourceRequest { project_dir = 1; user_id = 2; source_id = 3; title = 4; source_type_id = 5; copy_description = 6; copy_metadata = 7; copy_credibility = 8; repeated string artifact_ids = 9; }`; `CopySourceResponse { Source source = 1; WriteEffects effects = 2; }`. Handler `CopySource` lives in `api/ffi/handlers/sources.go` beside `CreateSource`; the response source goes through `enrichSourceProto` with `Updated = res.Activity` | Effects on the response let S12-02 patch `source_scopes` with no write-through (working rule "One title and type form") | No `WriteEffects`, as `CreateSourceResponse` — S12-02 would copy `create()`'s write-through, which the rule forbids |
| D-14 | Swift: `GenealogyStore.copySource(projectDir: String, userID: String, sourceID: String, title: String, sourceTypeID: String, copyDescription: Bool, copyMetadata: Bool, copyCredibility: Bool, artifactIDs: [String]) async throws -> (CatalogSource, CatalogWriteEffects)` *(new)*, placed after `createSource` in the protocol, `GoStore`, and `FakeStore` | Flat parameters like every store write | A new choices struct — no second caller |
| D-15 | `FakeStore.copySource` appends a new `CatalogSource` (new lowercase UUID id, ref `"SRC-FAKE1"`, chosen type and title, original description when chosen, else `""`), copies `metadataBySource` when `copyMetadata`, copies `credibilityBySource` with `id` new and `sourceID` = the copy's id when `copyCredibility` and the original has one, copies each chosen `CatalogArtifact` with a new id, `sourceID` = the copy's id, the same `fileID`, `label`, `description`, `file`; throws `StoreBoom.boom` when the original is not in `sourcesByProject[projectDir]`; returns `bumpSource(copyID, evidence: false)` | S12-02's model tests need a fake that behaves like the engine | A fake that returns a bare Source |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/sourcecopy` *(new)* | package | `core/database/connect` | Different write; same composing shape (precedent) |
| `sourcecopy.Draft`, `Choices` *(new)* | type | `sources.CreateInput` | Holds every part, not just the Source row |
| `sourcecopy.ReadDraft`, `CreateFromDraft`, `Copy` *(new)* | func | `connect.CreateCitedBridge` | — |
| `sources.GetTx` *(new, renamed from `getTx`)* | func | `sources.getTx` | Rename, not a copy (D-3) |
| `artifacts.ListBySourceTx` *(new)* | func | `artifacts.ListBySource` | Shares its SQL and scan (D-2) |
| `sourcecredibility.GetBySourceTx` *(new)* | func | `sourcecredibility.GetBySource` | Shares `sqlGetBySource` (D-2) |
| `sourcemetadata.ListBySourceTx` *(new)* | func | `sourcemetadata.ListBySource` | Shares `sqlListBySource` (D-2) |
| `sourcemetadata.ListLayoutTx`, `LayoutRow` *(new)* | func, type | `sourcemetadata.listLayout` | Returns ordered rows with field ids; `listLayout` returns a map for `ListWorkspace` (D-5) |
| `sourcemetadata.CreateAll` *(new)* | func | `sourcemetadata.Set`, `Reorder`, `DismissSuggestion` | Cannot reproduce a layout verbatim (D-4) |
| `METHOD_COPY_SOURCE`, `CopySourceRequest`, `CopySourceResponse` *(new)* | FFI | `METHOD_CREATE_SOURCE` | Different input |
| `handlers.CopySource` *(new)* | func | `handlers.CreateSource` | — |
| `CoreMethod.copySource`, `GenealogyStore.copySource` *(new)* | Swift | `createSource` | — |
| `TestCopy` *(new)*, `TestCreateAll` *(new)*, `TestCopySource` *(new)* | test | `TestCreateCitedBridge`, `TestSources` | — |

**Builds on.** `add-catalog-query` owner packages (each table's SQL stays in its package), the `connect` composing-package pattern, `writes.Run` and the existing effects registry entries, `add-ffi-handler` with `writeEffectsProto`.
**Deletes.** `sources.getTx` (renamed to `GetTx`).
**Next adds.** S12-02 calls `store.copySource`. S12-03 makes `sources.Create` mint the Default graph, which every copy then gets with no change here. Registry import builds a `Draft` and calls `CreateFromDraft`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sources` | `func GetTx(tx *sql.Tx, id []byte) (Source, error)` (rename of `getTx`) |
| `core/database/artifacts` | `func ListBySourceTx(tx *sql.Tx, sourceID []byte) ([]Artifact, error)` *(new)*; `ListBySource` and it share `listBySource(q, sourceID)` |
| `core/database/sourcecredibility` | `func GetBySourceTx(tx *sql.Tx, sourceID []byte) (Assessment, error)` *(new)*, `sql.ErrNoRows` when none |
| `core/database/sourcemetadata` | `func ListBySourceTx(tx *sql.Tx, sourceID []byte) ([]Row, error)` *(new)*; `type LayoutRow` *(new)*; `func ListLayoutTx(tx *sql.Tx, sourceID []byte) ([]LayoutRow, error)` *(new)*; `func CreateAll(...)` *(new)* per D-4, D-6 |
| `core/database/sourcecopy` *(new)* | `sourcecopy.go` *(new)* per D-1, D-7, D-9 – D-11 |

```go
// Draft is everything a copy can carry from a Source.
type Draft struct {
	SourceTypeID []byte
	Title        string
	Description  string
	Metadata     []sourcemetadata.Row
	Layout       []sourcemetadata.LayoutRow
	Credibility  *sourcecredibility.Assessment // nil when the Source has none
	Artifacts    []artifacts.Artifact          // ref order
}
```

Errors: `sources.invalid` → D-8 cases. `source_metadata.invalid` → `CreateAll` on a Source that already has rows (not reachable from `Copy`). `source_credibility.invalid` → a grade that no longer exists.

### FFI

`api/proto/engine.proto`: `METHOD_COPY_SOURCE = 102` *(new)*, `CopySourceRequest` *(new)*, `CopySourceResponse` *(new)* per D-13. Handler `CopySource` *(new)* in `api/ffi/handlers/sources.go`: parse `user_id` (`parseUserID`), `source_id` and `source_type_id` (`parseID`), each `artifact_ids` entry (`parseID`); `withProjectCatalog`; `writes.Run` per D-12; response `Source` + `Effects: writeEffectsProto(wrote)`. `api/ffi/dispatch.go`: `MethodCopySource` *(new)* and its `case`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view change. Store only: `CoreMethod.copySource` *(new)* in `macos/App/Platform/CoreInvoke.swift`; `copySource` in `macos/App/Platform/GenealogyStore.swift` (protocol), `macos/App/Platform/GoStore.swift` (builds `Provenencia_Engine_V1_CopySourceRequest`, returns `(Self.mapSource(resp.source), Self.mapWriteEffects(resp.effects))`), and `macos/App/Platform/FakeStore.swift` (D-15).

### Strings

None — no user-facing copy and no new error code (D-8).

## 7. Steps

**STEP-1 — Transaction reads in the owner packages**
Files: `core/database/sources/sources.go`, `core/database/artifacts/artifacts.go`, `core/database/sourcecredibility/assessments.go`, `core/database/sourcemetadata/sourcemetadata.go`. Skill: `add-catalog-query`.
Do: D-2, D-3, D-5. Rename `getTx` → `GetTx` and update its callers in `sources`. Add the Tx reads; route each Catalog read through the shared unexported function.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sources/... ./core/database/artifacts/... ./core/database/sourcecredibility/... ./core/database/sourcemetadata/...` passes; `rg -n 'func getTx' core/database/sources` prints nothing.

**STEP-2 — `sourcemetadata.CreateAll`**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`. Skills: `add-catalog-query`, `add-audit-scope`.
Do: D-4, D-6. Add T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/... -run TestCreateAll -v` passes.

**STEP-3 — `sourcecopy`**
Files: `core/database/sourcecopy/sourcecopy.go` *(new)*, `core/database/sourcecopy/sourcecopy_test.go` *(new)*.
Do: D-1, D-7, D-9 – D-11. Add T-1 – T-7 (one table, `TestCopy`). Seed the Subject the way `TestSourceDelete` does; run copies through `writes.Run` with `Action: "copy_source"`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecopy/... -v` passes.

**STEP-4 — FFI `CopySource`**
Files: `api/proto/engine.proto`, generated Go and Swift from `./scripts/generate-proto.sh`, `api/ffi/dispatch.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-12, D-13. Add T-9.
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-5 — Swift store method**
Files: `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`.
Do: D-14, D-15.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes.

**STEP-6 — Docs**
Files: `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: add `## 4.2 Copy Source` after `## 4.1 source_notes` in `docs/source-layer-data-model.md`: what is copied when chosen (description; metadata values and layout rows verbatim; credibility; each chosen Artifact as a new row with the same label, description and File), what is never copied (notes, Subjects, Citations, Observations), the cover reset, one audited `copy_source` write, and the two-tools rule linked to [`source-structure.md`](../source-structure.md#the-two-tools). Then the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-01-copy-source-engine.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcecopy/sourcecopy_test.go` `TestCopy` *(new)*, table-driven | "all parts": AC-1.1 fixture → new ref ≠ original's; title, type, description; `ListWorkspace` equal (field ids, values, `SortOrder`, `Suggested`); grade + argument; Artifacts' labels, descriptions, `FileID` equal and refs new; cover `type_icon` + nil | AC-1.1, AC-1.2, D-4, D-9, D-10 |
| T-2 | Go | `TestCopy` | "original untouched": snapshot the original's row, notes, `ListWorkspace`, layout, credibility, Artifacts before; equal after | AC-1.3 |
| T-3 | Go | `TestCopy` | "nothing chosen" → empty description; zero metadata, layout, credibility, Artifact rows for the copy | AC-2.1 |
| T-4 | Go | `TestCopy` | "second Artifact only" → one Artifact matching the second | AC-2.2, D-7 |
| T-5 | Go | `TestCopy` | "note and Subject not copied" → zero `source_notes`, `subjects` rows for the copy; zero `citations` on its Artifacts | AC-2.3, D-11 |
| T-6 | Go | `TestCopy` | "other type" → chosen type; metadata and layout still equal | AC-2.4 |
| T-7 | Go | `TestCopy` | "blank title", "unknown source", "foreign Artifact", "repeated Artifact", "unknown type" → `errors.Is(err, sources.ErrInvalid)`; Sources count unchanged | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-8 |
| T-8 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestCreateAll` *(new)*, table-driven | values with no layout rows → values present, zero layout rows; a dismissed layout row on a valued field → kept dismissed; a URL value → canonical form; a Source with an existing value → `ErrInvalid`; `ListLayoutTx` returns rows by `sort_order` | D-4, D-5, D-6 |
| T-9 | Go | `api/ffi/handlers/sources_test.go` `TestCopySource` *(new)*, `runRPC` | bad proto → error; bad `source_id` → error; ok copy of a Source with one Artifact → response title, type, `SRC-` ref, `effects.revision` > 0, copy id in `effects.source_scopes`, latest audit action `copy_source` (`assertLatestAuditAction`); a foreign `artifact_ids` entry → error | AC-1.4, D-12, D-13 |

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
- `METHOD_COPY_SOURCE = 102` collides with a value on `main`; report the free number you would use.
- `ListWorkspace` on a copy made with `CreateAll` differs from the original's for the same type (T-1 fails on order or `Suggested`), which would mean layout is not the whole story of field order.
- `effects.Resolve` rejects a `CreateAll` change (an error naming `source_metadata_layout`).
- `sources.ErrInvalid` is not mapped under `L10n.Errors` in `macos/App/Platform/L10n.swift` (`rg -n 'sources.invalid' macos/App`).

## 12. Open questions

None.

## 13. PR

**Title:** S12-01 — Copy Source engine
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-01-copy-source-engine.md`.
