# S15-33 — Editing an Artifact can save its metadata

**Status:** Landed in [#806](https://github.com/mendahu/provenencia/pull/806) (2026-11-12)
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: engine and wire only; the app sends no draft and reads no new field until S15-36 and S15-34)
**Board:** —
**Depends on:** S15-08 (landed [#791](https://github.com/mendahu/provenencia/pull/791))
**Blocks:** S15-36, S15-34, S15-10
**Branch:** `s15-33-artifact-update-saves-metadata`
**Version bump:** patch, 0.11.0 → **0.11.1** (or the next patch after `main`'s `VERSION` when you rebase): `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` — one request field and two response fields that old clients ignore, and a param on an existing error code; no screen or catalog change. Same call as S15-26's response fields ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md) (read only: `artifact_metadata` already scopes to its Artifact's Source since S15-08; nothing to register)
**Rules that bind:** `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "Metadata saves with its owner's write" and "An Artifact is added and edited in one sheet" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); map review K-36 ([`../deployment-plan.review.md`](../deployment-plan.review.md)); brief review M-1 ([`../design/S15-D14-source-page-artifacts.review.md`](../design/S15-D14-source-page-artifacts.review.md)) and S15-D14 SA-9 ([`../design/S15-D14-source-page-artifacts.md`](../design/S15-D14-source-page-artifacts.md)); S15-08's [plan](S15-08-artifacts-and-graphs-hold-metadata.md) (the graph update's draft); S15-26's [plan](S15-26-metadata-board-takes-an-owner.md) (board writes return the owner's rows)
**Planned:** 2026-11-13 against `main` @ `9732554`
**Review:** [`S15-33-artifact-update-saves-metadata.review.md`](S15-33-artifact-update-saves-metadata.review.md) — Fits with changes — Ready with changes, 2 rounds; every finding Fixed, Declined with evidence or Logged; nothing To user
**Runs on:** STEP-2's `./scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go` and `protoc-gen-swift`; `scripts/check-pr.sh` needs Xcode (the regenerated `engine.pb.swift` must build). Without them, stop: never hand-edit the generated `engine.pb.go` or `engine.pb.swift`

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

## Landed

- T-4 (AC-1.5, "nothing changed writes no revision") runs on `page2`, which has no layout, instead of on the Artifact the earlier subtests edit: after "empty clears" that Artifact's layout has `provider` at 1, so the AC's precondition (layout `provider` 0) would not hold.
- `errParams` is a helper in `metadata_test.go` (handlers) and in `sourcemetadata_test.go`, as T-1 and T-5 read params with it.

## 1. Objective

Editing an Artifact saves its copy details in the same write as its label, description and graph links, so the Artifact sheet (S15-34) has one Save that is one write and one audit revision. Both Artifact write replies return the Artifact's board after the write, so the page patches its cached workspace instead of reloading it (map review K-36). A draft value the engine refuses names its field, so the sheet can put the message under that row (brief review M-1, SA-9).

**Why now.** Working rule "Metadata saves with its owner's write" names `UpdateArtifactRequest` from this PR; S15-36 maps the draft and the returned rows into Swift, and S15-34 and S15-10 build the sheet on them.

**What stays.** Every view and every Swift file except the regenerated `engine.pb.swift`. `UpdateArtifact` without a draft writes what it writes today. The graph writes keep their requests and replies. `sourcemetadata.ErrInvalid`'s code and its copy ("This metadata change can't be saved.") stay; only some refusals gain a param.

## 2. User stories

**US-1.** As the engine, when an Artifact update carries a metadata draft, I save the label, description, graph links and the draft in one write, so that one Save is one revision.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Artifact A (`newMetadataOwners`' `page1`, label "Scan") given `provider` = "FamilySearch", then `url` = "https://example.org/a", each by `SetMetadata` (so A's layout holds `provider` 0, `url` 1); `UpdateArtifact` with label "Scan (front)" and a draft (built by hand, not with `draftOf`, which holds one value): values `provider` = "Ancestry", then `accessed` = "12 March 2025" (new), `url` left out; layout `accessed` 0, `provider` 1 | One revision, one more than the last; `latestRevisionTypes` is exactly `artifact update`, `artifact_metadata delete` (`url`), `artifact_metadata update` (`provider`), `artifact_metadata create` (`accessed`), `artifact_metadata_layout create` (`accessed`), `artifact_metadata_layout update` (`provider` 0 → 1); the board reads `accessed`, `provider` in that order, and no `url` entry (its layout row stays, but `ListWorkspace` lists no field that is neither suggested nor valued) |
| AC-1.2 | A holds `provider`; `UpdateArtifact` with the same label and no `metadata` | No `artifact_metadata` change; `provider` still held |
| AC-1.3 | A holds `provider`; `UpdateArtifact` with a present, empty `metadata` | `provider` deleted; the board holds no value |
| AC-1.4 | A in graph G, whose Citations use A; `UpdateArtifact` with `graph_links` that leave G, and a draft setting `provider` | `evidencegraphs.ErrArtifactInUse`; label, links and metadata unchanged; no revision |
| AC-1.5 | A with nothing changed: same label and description, no `graph_links`, a draft equal to what is stored: A holds `provider` by one `SetMetadata` (layout `provider` 0); the draft is values `provider` with the same text, layout `provider` 0 only. No layout row for an unvalued suggested entry: it has none stored, so listing it would insert one | No revision (`effects.revision` 0); the reply still carries the board |

**US-2.** As the app, both Artifact write replies carry the Artifact's board after the write, so that I patch the page's cached workspace from the reply.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `UpdateArtifact` of AC-1.1 | `UpdateArtifactResponse.metadata` equals `GetMetadataBoard` for A after the write, entry for entry (field id, value, `suggested`, `sort_order`, `open_href`) |
| AC-2.2 | `CreateArtifact` with a draft `provider` = "FamilySearch" on a Source whose type suggests `url` at the `artifact` level | `CreateArtifactResponse.metadata` equals `GetMetadataBoard` for the new Artifact: `url` suggested and empty, `provider` with its value |
| AC-2.3 | `CreateArtifact` with no draft, or `UpdateArtifact` with no draft | The reply's `metadata` is the Artifact's board (the type's `artifact` suggestions, empty, for a new Artifact) |

**US-3.** As the app, a refused draft value names its field, so that the sheet shows the refusal under that row.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `UpdateArtifact` with a draft `provider` = "Ancestry", `url` = "not a link" | `sourcemetadata.invalid` with params `[<url field id>]`; nothing written: label, links and values as before |
| AC-3.2 | A draft whose second value is "   " (blank after trim) on field `accessed` | `sourcemetadata.invalid` with params `[<accessed field id>]` |
| AC-3.3 | A draft with two refused values, `url` = "x" first, `accessed` = " " second | params `[<url field id>]` (the first in draft order) |
| AC-3.4 | `CreateArtifact` with a draft `url` = "not a link" | `sourcemetadata.invalid` with params `[<url field id>]`; no Artifact created |
| AC-3.6 | `SetMetadata` on a url field with "not a link" | `sourcemetadata.invalid` with params `[<url field id>]`, as a draft refusal |
| AC-3.5 | `sourcemetadata.Apply` with a draft that repeats a field, or names an unknown field id, or on an owner with no row | `ErrInvalid` with no params. (Through `UpdateArtifact` a missing Artifact never reaches `Apply`: `artifacts.Get` refuses it first, as today) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `UpdateArtifactRequest.metadata` (`MetadataDraft`, field 7), applied through `applyDraft` in the update's write | STEP-2, STEP-3, AC-1.1 – AC-1.5 |
| `repeated MetadataWorkspaceEntry metadata = 3` on `UpdateArtifactResponse` and `CreateArtifactResponse`, read after the write | STEP-2, STEP-3, AC-2.1 – AC-2.3 |
| One value step `normalizeFor` *(new)* that trims, normalizes and refuses with `ErrInvalid.WithParams(fieldID)` (wire form of the id); `Set` and `apply` both call it, so `Set`, `Apply` and `CreateAll` all name the field | STEP-1, AC-3.1 – AC-3.6 |
| The params contract written down: `Error.params`' proto comment, `apperr.Error`'s and `apperr.New`'s doc comments, `sourcemetadata.ErrInvalid`'s doc comment, one line in `add-ffi-handler` | STEP-1, STEP-2, D-6 |
| `ownerBoardProto` *(new)* in `metadata.go`: the one "owner's board" read; the five board handlers, `GetSourceWorkspace`'s Source board and both Artifact handlers use it | STEP-3, D-2 |
| Version bump, docs | STEP-4 |

| Out | Home |
| --- | --- |
| Any hand-written Swift: the Swift draft, `metadata:` on the store writes, mapping the returned rows and the param | S15-36 |
| The Artifact sheet, applying `.wroteMetadataBoard` from the reply | S15-34 |
| Add mode and the retry after a refused file | S15-10 |
| Graph write replies returning their rows | S15-11: its scope entry now asks its planner to decide this (review M-6) |
| A param when a value's field was deleted meanwhile (plain `ErrInvalid`; SA-9's top Callout shows it) | Not planned |

## 4. Current state

Verified at `9732554`.

| Area | Where | Today |
| --- | --- | --- |
| Update request | `api/proto/engine.proto` `UpdateArtifactRequest` (fields 1–6) | Label, description, `graph_links`; no metadata. Its comment: "patches label and/or description, and, when graph_links is present, sets the Artifact's graph links in the same write. File attach stays IngestArtifactFile (first-attach only)." The last sentence stays |
| Replies | `api/proto/engine.proto` `CreateArtifactResponse`, `UpdateArtifactResponse` (fields 1, 2) | `artifact`, `effects`; no board |
| Precedent request | `api/proto/engine.proto` `UpdateEvidenceGraphRequest.metadata = 7`, `CreateArtifactRequest.metadata = 9` | `MetadataDraft`, "absent = metadata unchanged; present = the stored values become these" |
| Precedent reply | `api/proto/engine.proto` `SetMetadataResponse.metadata = 3` | The owner's whole board in display order |
| Create handler | `api/ffi/handlers/artifacts.go` `CreateArtifact` | Parses the draft with `parseDraft`, runs `artifacts.Create` then `applyDraft` on `Owner{LevelArtifact, a.ID}` in one `writes.Run`; replies `artifact`, `effects` |
| Update handler | `api/ffi/handlers/artifacts.go` `UpdateArtifact` | `artifacts.Get`, then one `writes.Run` around `artifacts.Update(tx, userID, updated, graphs)`; re-reads with `artifacts.Get`; replies `artifact`, `effects`. No draft |
| Graph update | `api/ffi/handlers/evidence_graphs.go` `UpdateEvidenceGraph` | The shape to copy: `parseDraft` before the catalog opens; inside `writes.Run`, `evidencegraphs.Update` then `applyDraft(tx, userID, Owner{LevelGraph, graphID}, draft)`, changes appended |
| Draft helpers | `api/ffi/handlers/metadata.go` `parseDraft`, `applyDraft`, `metadataEntriesProto` | nil draft = no change; `applyDraft` calls `sourcemetadata.Apply` |
| Board read after a write | `api/ffi/handlers/metadata.go` `SetMetadata` | `sourcemetadata.ListWorkspace(c, owner)` after `writes.Run`, then `metadataEntriesProto` |
| Refusal | `core/database/sourcemetadata/sourcemetadata.go` `apply`, `Set`, `normalizeValue` | `apply` checks the whole draft before writing; its loop over `d.Values` trims at the call and returns the first `normalizeValue` refusal (blank, non-text/url type, a URL `urlshape.Canonical` refuses) as bare `ErrInvalid`, in draft order. `Set` trims `in.ValueText` at its top and calls `normalizeValue` the same way: two copies of one value step |
| Board after a write | `api/ffi/handlers/metadata.go` `SetMetadata`, `ClearMetadata`, `DismissMetadataSuggestion`, `ReorderMetadata`, `GetMetadataBoard`; `api/ffi/handlers/sources.go` `GetSourceWorkspace` (`meta` … `out.Metadata = metadataEntriesProto(meta)`) | Each repeats `sourcemetadata.ListWorkspace(c, owner)` then `metadataEntriesProto`. `artifactBoardsProto` (`metadata.go`) is the batch form over many Artifacts and stays |
| Params contract | `api/proto/engine.proto` `message Error` `params = 3` "ordered args for L10n format strings"; `.agents/skills/add-ffi-handler/SKILL.md` Delete paragraph ("do not … stuff refs into `apperr` params", for Delete refusals); `sourcemetadata.ErrInvalid` has no doc comment | Every `WithParams` caller today passes display text |
| Error params | `core/apperr/apperr.go` `(*Error).WithParams`; `api/ffi/error.go` | Params ride the wire `Error.params`; `errors.Is` ignores params |
| Swift copy | `macos/App/Platform/L10n.swift` `L10n.Errors.message(code:params:)` case `"sourcemetadata.invalid"` | Ignores params: the copy does not change |
| Tests | `core/database/sourcemetadata/sourcemetadata_test.go` `TestApply` "refused drafts write nothing" | Asserts `errors.Is(err, ErrInvalid)` and no rows; does not look at params |
| Tests | `api/ffi/handlers/artifacts_test.go` `TestCreateArtifactMetadata`, `TestUpdateArtifactGraphLinks` | Create with a draft; update links. No update-with-draft test |
| Tests | `api/ffi/handlers/evidence_graphs_test.go` `TestUpdateEvidenceGraph` "metadata: absent changes nothing, present replaces" | The update-with-draft precedent test |
| Test helpers | `api/ffi/handlers/metadata_test.go` `draftOf`, `latestRevisionTypes`, `newMetadataOwners`, `metadataByField`, `metadataFieldIDs`; `source_defs_test.go` `newUserField` (hard-codes `DataType: "text"`); `harness_test.go` `sourceFixture`, `rpc`, `mustRPC` | Reuse them. Url fields are created inline today (`sources_test.go` `urlMetadataSetFixture` and three cases) |

**Precedent.** S15-08 ([#791](https://github.com/mendahu/provenencia/pull/791)) put `MetadataDraft` on `UpdateEvidenceGraphRequest` (field 7) and `CreateArtifactRequest` (field 9): copy `UpdateEvidenceGraph`'s parse-then-`applyDraft` shape line for line. S15-26 ([#798](https://github.com/mendahu/provenencia/pull/798)) made `SetMetadataResponse` return the owner's board: copy its `ListWorkspace` after the write.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `UpdateArtifactRequest` gains `MetadataDraft metadata = 7` with the graph's comment ("absent = metadata unchanged; present = the stored values become these"). The handler parses it with `parseDraft` before the catalog opens and runs `applyDraft` after `artifacts.Update`, inside the same `writes.Run` closure, appending its changes | One write and one revision per Save (working rule); the graph update already does exactly this | A separate `ApplyMetadata` RPC — two writes, two revisions, a half-saved sheet on failure |
| D-2 | `UpdateArtifactResponse` and `CreateArtifactResponse` gain `repeated MetadataWorkspaceEntry metadata = 3`, filled on every success (draft or not) by `ownerBoardProto(c, Owner{LevelArtifact, id})` *(new)* after `writes.Run`. `ownerBoardProto` is `ListWorkspace` then `metadataEntriesProto`; `SetMetadata`, `ClearMetadata`, `DismissMetadataSuggestion`, `ReorderMetadata`, `GetMetadataBoard` and `GetSourceWorkspace`'s Source board move onto it (review M-3) | The client patches its cache from every Artifact write without asking whether a draft was sent; same read and same entry type `SetMetadataResponse` uses | `MetadataBoard` (adds an `owner_id` the client already has); filling only when a draft was sent (the client would need two code paths) |
| D-3 | A new unexported `normalizeFor(field metadatafields.Field, text string) (string, error)` trims `text`, calls `normalizeValue(field.DataType, …)`, and on refusal returns `ErrInvalid.WithParams(database.UUIDString(field.ID))`. `Set` (in place of its top-of-function trim and its `normalizeValue` call) and `apply`'s `d.Values` loop call it. Nothing else gains a param: repeated ids, unknown field ids, a missing owner, a nil tx or user stay bare `ErrInvalid` | One value step, so every value refusal names its field on a live board and a draft alike (review M-1); those are the refusals a researcher fixes on one row. The others are a client bug or a race, which SA-9's top Callout covers. The param is the id's wire form, which Swift's board rows key on | Only `apply` names it (two copies of one rule, review M-1); naming the field in `loadFields` too (no row-level UI reads it); a new error code (SA-9 shows today's copy) |
| D-4 | Reply metadata is read outside the write transaction, after `writes.Run` returns, as `SetMetadata` does; the update handler's existing `artifacts.Get` re-read stays | Same consistency as every board write today; the catalog session serializes ops | Reading inside the closure (changes `writes.Run`'s return value for one caller) |
| D-6 | A field id in `params` is a locator, never copy. The proto comment on `Error.params` becomes "ordered args for L10n format strings, or, where the error's Go doc says so, the id of the row the refusal is about"; `sourcemetadata.ErrInvalid` gains a doc comment: "A refused value carries its field's wire id as params[0], for the client to place the message on its row; it is never shown."; `add-ffi-handler` gains, after its Delete paragraph: "A value refusal may carry the id of the row it is about (a metadata field's wire id) as `params[0]`, when the error's Go doc says so; the client places the message on that row and never shows the id. Delete refusals carry no refs."  | The first id-valued param; the contract the next agent reads must say so, and copy must never interpolate it (review M-2) | A new `Error` field for locators (a wire change for one caller) |
| D-5 | Tests: one new table-driven handler test `TestUpdateArtifactMetadata` in `artifacts_test.go`; `TestCreateArtifactMetadata` gains the reply and param assertions; `TestApply` gains a "a refused value names its field" case | Handler tests drive the wire (add-ffi-handler); the param rule lives in `sourcemetadata` and is tested there | A Swift test (no Swift changes here) |

**Rules this plan bends.** None. `add-ffi-handler`'s rule against stuffing refs into `apperr` params is scoped to Delete refusals (the report is the preview); this is a value refusal, and the map's S15-33 scope entry (user-reviewed, brief review M-1) chose `ErrInvalid.WithParams(fieldID)`. D-6 writes the contract down.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `UpdateArtifactRequest.metadata` *(new)* | proto field | `UpdateEvidenceGraphRequest.metadata` | Reuses `MetadataDraft`, `parseDraft`, `applyDraft` as they are |
| `UpdateArtifactResponse.metadata`, `CreateArtifactResponse.metadata` *(new)* | proto field | `SetMetadataResponse.metadata` | Reuses `MetadataWorkspaceEntry`, `ListWorkspace`, `metadataEntriesProto` |
| `normalizeFor` *(new)* | func | `normalizeValue` (kept; `normalizeFor` wraps it) and the trims in `Set` and `apply` | Replaces the two copies of trim-then-normalize with one step that also names the field |
| `ownerBoardProto` *(new)* | func | `metadataEntriesProto`; `artifactBoardsProto` (batch, many owners); the six inline `ListWorkspace` + `metadataEntriesProto` pairs | Replaces the six; this PR would add the seventh and eighth. `artifactBoardsProto` stays: one batch read for a Source's Artifacts |
| `newField` *(new)* | test helper | `newUserField` (`source_defs_test.go`); four inline url-field creations in `sources_test.go` | `newUserField` delegates to it with `"text"`; its callers do not change; the four inline creations move onto it |
| `assertReplyBoard` *(new)* | test helper | `metadataByField`, inline `GetMetadataBoard` fetches in `evidence_graphs_test.go` | No shared "the reply is the board" assertion exists; T-2, T-4, T-7 use it |
| `TestUpdateArtifactMetadata` *(new)* | test | `TestUpdateEvidenceGraph` metadata subtest | A new test function: the Artifact handler's tests live in `artifacts_test.go` |

Searched: `parseDraft`, `applyDraft`, `metadataEntriesProto`, `ListWorkspace(` callers, `normalizeValue(` callers, `WithParams` in `core/`, `DataType: "url"` in handler tests. No new type or L10n key.

**Builds on.** S15-08's draft-on-write (`parseDraft`, `applyDraft`), S15-26's board-in-reply, `apperr.WithParams` as `metadatafields.ErrDuplicateKey.WithParams(key)` uses it.
**Deletes.** `Set`'s top-of-function `strings.TrimSpace` and both direct `normalizeValue` calls (in `Set` and `apply`; `normalizeFor` is its only caller after); the six inline board reads in `metadata.go` and `sources.go`. `UpdateArtifactRequest`'s and `Error.params`' comments are rewritten.
**Next adds.** S15-36 maps `metadata` on the store writes and the reply's rows; S15-34 applies `.wroteMetadataBoard` from `UpdateArtifactResponse.metadata` and puts the param's message under its row; S15-10 reads `CreateArtifactResponse.metadata` and retries a refused file with the draft on `UpdateArtifact`. S15-11, if it returns graph rows, is two proto fields and two `ownerBoardProto` calls.

### Reach

None — no new path into existing code: `apply` is already reached by `CreateArtifact` and the graph writes; `UpdateArtifact` is a new caller of `Apply` on an owner `CreateArtifact` already writes. No client sends the new field yet.

### Scale

At the map's Scale (a Source with 60 Artifacts, 63 seeded fields): one Save is one write and one `ListWorkspace` read for one Artifact (types, suggestions for one (type, level), values, layout, extras: five queries, bounded by the Artifact's fields). The reply carries one board, never the Source's 60. The write's effects are what S15-08 registered for `artifact` and `artifact_metadata` (the Source scope); no new key or cache.

### Edges

- **Assumptions.** `writes.Run` treats an empty change list as no revision: an update with nothing changed and an equal draft gives revision 0 and still a reply board (AC-1.5, T-4). `apply` checks the whole draft before writing, so a refusal leaves no partial write; with `artifacts.Update` first in the closure, the rolled-back transaction also drops the label change (AC-3.1, T-5).
- **Neighbors.** Absent draft vs present empty draft (AC-1.2, AC-1.3, T-3); a draft equal to the stored rows writes nothing (T-4); a refused graph unlink with a valid draft (AC-1.4, T-6).
- **Removed inputs.** A value left out of the draft is deleted; a layout row left out stays (S15-07's `Apply` contract, unchanged; AC-1.1). A field deleted between the sheet opening and Save is a bare `ErrInvalid` (D-3).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — the draft is bounded by the request as today's graph and create drafts are; `checkFieldIDs` refuses repeats.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcemetadata` | `func normalizeFor(field metadatafields.Field, text string) (string, error)` *(new)*, beside `normalizeValue`; `Set` and `apply` call it (D-3). `ErrInvalid` gains D-6's doc comment; `Apply`'s and `Set`'s doc comments say a refused value names its field |
| `api/ffi/handlers` | `func ownerBoardProto(c *database.Catalog, owner sourcemetadata.Owner) ([]*engine.MetadataWorkspaceEntry, error)` *(new)* in `metadata.go`, beside `metadataEntriesProto` (D-2). `GetMetadataBoard` wraps its result in `MetadataBoard` as today |

Errors: `sourcemetadata.invalid` with params `[fieldID]` → a value (live or draft) is blank after trim or not a valid value for its field's type; with no params → every other refusal, as today.

### FFI

`api/proto/engine.proto`:

```proto
// UpdateArtifactRequest patches label and/or description, and, when graph_links
// is present, sets the Artifact's graph links, and, when metadata is present,
// its metadata, all in the same write.
// File attach stays IngestArtifactFile (first-attach only).
message UpdateArtifactRequest {
  // fields 1–6 unchanged
  MetadataDraft metadata = 7; // absent = metadata unchanged; present = the stored values become these
}

message CreateArtifactResponse {
  Artifact artifact = 1;
  WriteEffects effects = 2;
  repeated MetadataWorkspaceEntry metadata = 3; // the Artifact's board after the write, in display order
}

message UpdateArtifactResponse {
  Artifact artifact = 1;
  WriteEffects effects = 2;
  repeated MetadataWorkspaceEntry metadata = 3; // the Artifact's board after the write, in display order
}
```

Handlers: `api/ffi/handlers/artifacts.go` `CreateArtifact`, `UpdateArtifact`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — `engine.pb.swift` is regenerated; no hand-written Swift changes (S15-36).

### Strings

None — `error.sourcemetadata.invalid` keeps its copy and ignores the param.

## 7. Steps

**STEP-1 — A refused draft value names its field**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`, `core/apperr/apperr.go` (comments only).
Do: D-3 (`normalizeFor`, `Set` and `apply` on it); D-6's `ErrInvalid` comment; `core/apperr/apperr.go`'s `Error` comment ("Params are ordered L10n format args") and `New`'s ("params are optional interpolation arguments") reworded to D-6's: ordered L10n format args or, where the code's doc says so, the id of the row the refusal is about; `Set`'s and `Apply`'s comments. Add T-1 and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcemetadata/...` passes.

**STEP-2 — The wire fields**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: the three fields and the rewritten comment of §6 FFI; D-6's `Error.params` comment and its `add-ffi-handler` line (`.agents/skills/add-ffi-handler/SKILL.md`, after the Delete paragraph); run `./scripts/generate-proto.sh`.
Check: `go build ./...` passes; `git diff --stat` shows only the proto, the two generated files and the skill.

**STEP-3 — The handlers**
Files: `api/ffi/handlers/metadata.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`, `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/artifacts_test.go`, `api/ffi/handlers/metadata_test.go`, `api/ffi/handlers/source_defs_test.go`.
Do: `ownerBoardProto`, and the five board handlers and `GetSourceWorkspace`'s Source board on it (their tests must pass unchanged); D-1 in `UpdateArtifact` (copy `UpdateEvidenceGraph`); D-2 in both Artifact handlers (after `writes.Run`, before building the reply). Test helpers: `newField(t, dir, userID, label, dataType string) string` in `source_defs_test.go`, with `newUserField` calling it with `"text"`, and the four inline url-field creations in `sources_test.go` (`urlMetadataSetFixture`; the cases "set url metadata accepts host-shaped values", "rejects file scheme", "rejects a single label") replaced by `newField(t, req.ProjectDir, req.UserId, <their label>, "url")`; `assertReplyBoard(t, dir string, owner *engine.MetadataOwner, got []*engine.MetadataWorkspaceEntry)` in `metadata_test.go` beside `draftOf` (fetches `GetMetadataBoard`, compares lengths, `proto.Equal` per entry). Add T-2 – T-7, T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-4 — Version and docs**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj`, `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan (status).
Do: the version bump of the header; the DoD's docs lines.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Save an Artifact's metadata in the same write as its edit.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestApply` "a refused value names its field" *(new case)* | From `applyWorld`, through `runApply` unless named: values `f1` = "a", `doc` = "   " → params `[database.UUIDString(doc.ID)]`; `link` = "not a link" → `[UUIDString(link.ID)]`; `link` = "x" then `doc` = " " → `[UUIDString(link.ID)]`; `runCreateAll` with `link` = "x" → `[UUIDString(link.ID)]`; each of the existing "refused drafts write nothing" cases except "blank value" → no params (extend that case's loop to assert `len(ae.Params()) == 0`, skipping "blank value", which T-1's first row covers with `[doc id]`). Params read with `errors.As(err, &ae)` on `*apperr.Error`. Each writes no row: `countRows` over `source_metadata` and `source_metadata_layout` is 0 | AC-3.2, AC-3.3, AC-3.5, D-3 |
| T-2 | Go | `api/ffi/handlers/artifacts_test.go` `TestUpdateArtifactMetadata` *(new)* "saves the edit and the draft in one revision" | AC-1.1's Artifact and draft → revision +1; `latestRevisionTypes` = `artifact update`, then the value delete, update, create, then the layout creates; reply `metadata` equals `GetMetadataBoard` | AC-1.1, AC-2.1 |
| T-3 | Go | `TestUpdateArtifactMetadata` "absent leaves metadata, empty clears" | No `metadata` → `provider` held, no `artifact_metadata` change; `&engine.MetadataDraft{}` → nothing held | AC-1.2, AC-1.3, AC-2.3 |
| T-4 | Go | `TestUpdateArtifactMetadata` "nothing changed writes no revision" | Same label, description, no links, the AC-1.5 Artifact and draft → `effects.revision` 0; reply `metadata` equals the board | AC-1.5 |
| T-5 | Go | `TestUpdateArtifactMetadata` "a refused value names its field and writes nothing" | Label change plus draft `url` = "not a link" → `apperr` code `sourcemetadata.invalid`, params `[url field id]`; label and values as before | AC-3.1 |
| T-6 | Go | `TestUpdateArtifactMetadata` "a refused unlink drops the draft" | Links leaving a graph whose Citations use A, plus a valid draft → `evidencegraphs.ErrArtifactInUse`; no metadata change | AC-1.4 |
| T-7 | Go | `TestCreateArtifactMetadata` *(extended)* | The create with `provider` → reply `metadata` equals `GetMetadataBoard` for the new Artifact; the plain create's reply `metadata` equals its board; a draft `url` = "not a link" → params `[url field id]`, no Artifact created | AC-2.2, AC-2.3, AC-3.4 |

**Fixtures.** T-2 – T-6 start from `newMetadataOwners` (`metadata_test.go`): its `artifact` owner is `f.page1`, and its `suggested` field ("Provider") is suggested at the `artifact` level, so it stands for `provider`. `accessed` is a `newUserField` text field the test creates. The url field is `newField(t, dir, userID, "URL", "url")`. T-6 copies `TestUpdateArtifactGraphLinks` "a refused unlink rolls back the label": a new page A in the default graph G (`f.createArtifactResp` with `GraphIds: []string{f.defaultGraphID}`), a Citation on it through `CreateCitationWithObservations` with `graphLinking` and `validLocatorJSON`, then `UpdateArtifact` with `&engine.GraphLinks{}` and a draft. T-7 keeps `TestCreateArtifactMetadata`'s setup (its `provider` is a text field the draft fills) and adds `url := newField(t, dir, userID, "URL", "url")`, assigned to the Source's type at the `artifact` level with `AssignTypeField` before the draft create, as `newMetadataOwners` does: AC-2.2's suggested, empty entry. The existing "board holds one entry" assertion becomes `assertReplyBoard` plus a check that `url` is suggested and empty and `provider` holds "FamilySearch". Compare a reply's `metadata` with the board through `assertReplyBoard`. Read params with `errors.As(err, &ae)` on `*apperr.Error` and `ae.Params()`, as `conclusion_detail_test.go` reads `ae.Code()`.

| T-8 | Go | `sourcemetadata_test.go` `TestSourceMetadata`, beside the blank-value `runSet` case (~line 571) | `runSet` on `link` with "not a link" → `ErrInvalid`, params `[UUIDString(link.ID)]`; `runSet` on a missing field → no params | AC-3.6, D-3 |
| T-9 | Go | `api/ffi/handlers/metadata_test.go` `TestMetadataOwners` *(extended: a new subtest "a refused url value names its field", after the owner loop, on `m.artifact` only)* | `SetMetadata` on the artifact owner with a `newField(.., "url")` field and "not a link" → params `[field id]` | AC-3.6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestApply`, `TestCreateArtifactMetadata`, `TestUpdateArtifactGraphLinks`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: no name renamed or deleted; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Rules this plan bends — none
- [ ] UI matches the brief inventory — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — n/a, no brief (S15-D14 moves with S15-34)
- [ ] Domain doc — none states this rule
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: no view sends the new field until S15-34. The handler tests drive the wire end to end (T-2 – T-7).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Swift caller decodes `metadata` as something else | Fields 3 and 7 are free on all three messages at the base; the generated Swift only adds properties |
| A test fixture lacks a url-typed field | §8 **Fixtures**: `newField(.., "url")` |
| Moving six reads onto `ownerBoardProto` changes a reply | Their existing tests must pass unchanged in STEP-3; the helper is the same two calls |

**Stop and report if:**

- `UpdateArtifactRequest` field 7, or field 3 on either reply, is taken or reserved at your base.
- `apply` no longer checks the whole draft before its first write (a refusal could then leave a partial write).
- `writes.Run` records a revision for an empty change list (AC-1.5 would not hold).
- `./scripts/generate-proto.sh` changes more than the three messages in the generated files (a protoc version drift).

## 12. Open questions

None.

## 13. PR

**Title:** S15-33 — Editing an Artifact can save its metadata
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names); deviations from the plan; manual check: none (engine only); `Plan: docs/deployment-plan/spike-15/prs/S15-33-artifact-update-saves-metadata.md`.
