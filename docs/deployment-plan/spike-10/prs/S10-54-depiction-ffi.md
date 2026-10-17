# S10-54 — Depiction values over the FFI

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-54--depiction-values-over-the-ffi)
**Brief:** — (no new chrome: proto fields, handler wiring, Swift store types, a value-type label, and two error strings)
**Board:** —
**Depends on:** S10-32 (landed in [#430](https://github.com/mendahu/provenencia/pull/430); read [its Landed note](S10-32-depiction-value-type.md#landed))
**Blocks:** S10-34, S11-20 (was S10-36, moved to Spike 11 on 2026-10-17)
**Branch:** `s10-54-depiction-ffi`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) ("The image"), [S10-32 plan](S10-32-depiction-value-type.md) D-7 to D-10, its Out rows that name S10-54, and its Landed note (`depictionvalues.Readable`), working rules "A depiction's value is its own row: region plus image", "The client renders the image; Go validates it and stores its own copy"
**Planned:** 2026-10-17 against `main` @ `233e87ad` (S10-32 landed); first planned 2026-10-17 against `ae26c629` with S10-32 as planned
**Review:** [`S10-54-depiction-ffi.review.md`](S10-54-depiction-ffi.review.md) — Fits with changes, Not ready → changes applied; 1 round. Refresh after S10-32 landed: 2 rounds, Fits — Ready

> **Refreshed after S10-32 landed.** Every name this plan marks *(S10-32)* was checked against `main` @ `233e87ad`. The one change S10-32's Landed note makes to this plan: every reader that returns a stored depiction goes through `depictionvalues.Readable`, which empties a region that is not valid on the Citation's page and an image File that is not `image/png`. D-4, D-6, AC-2.3, and T-12 carry that.

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

S10-32 gives the engine a `depiction` value: a region of the Citation's page plus an image File Go stored from the app's rendering of it. This PR carries that value across the FFI. A save sends the region and the absolute path of the image the app wrote to a temporary file; Go stages it through S10-32's one entry point (`ingest.StageDepictions`) in every handler that writes Observations, and the response returns the saved region and the image's `objects/…` path. Every read of a Citation's Observations returns the same two fields, through S10-32's `depictionvalues.Readable`, so a stored value Go would not show comes back with that half empty. Swift gains the `depiction` value type, its *Depiction* label, the store fields, and copy for the two errors a save can now return.

**Why now.** The composer row (S10-34) and Promote's crops (S11-20) both wait on this shape. S10-32 left the FFI, the Swift value type, and hydrating the source graph's depiction to this PR (its Out rows).

**What stays.** No chrome: the composer still offers no depiction Property, and nothing in the app writes a likeness until S10-34. Bytes never cross the FFI. Go never takes a File id from the client. Every other value type saves and reads as today; the create and append responses gain the Property summary and values the update and connect responses already carry (D-5).

## 2. User stories

**US-1.** As the app, when I save a likeness, I send its region and the path of the crop I rendered, and I get back the region and the path of the image Go stored.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new project; an Artifact whose File is an 8 × 8 PNG; a Person subject. `CreateCitationWithObservations` with locator `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`, transcription `""`, and one `ObservationDraft` on the Person with the `likeness` Property and `depiction` = { `region_json`: the S10-32 AC-1.1 region, `image_path`: a 300 × 200 opaque gray JPEG in a temp folder } | It saves. The response's one Observation has `property_value_type` `depiction`, a non-empty `value_depiction_id`, `depiction.region_json` equal to the sent region, and `depiction.image_rel_path` matching `^objects/[0-9a-f]{2}/[0-9a-f]{2}/[0-9a-f]{64}\.png$`. That file exists under the project folder and decodes as a 300 × 200 PNG, though the crop was a JPEG: Go stored its own encoding. |
| AC-1.2 | AC-1.1 with `image_path` a symlink to the crop | Refused with `ingest.symlink`. No Citation is written and no file is added to the project's objects folder. |
| AC-1.3 | AC-1.1 through `AddObservationsToCitation` on an existing Citation with that locator | Same response fields as AC-1.1. |
| AC-1.4 | `CreateCitedBridge` whose drafts include one with a `depiction` whose `image_path` is a symlink | Refused with `ingest.symlink`, before any write. |

**US-2.** As the app, when I read a Citation's Observations, a likeness carries its region and image path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | After AC-1.1, `GetCitation` for that Citation | The likeness row has the same `value_depiction_id`, `depiction.region_json`, and `depiction.image_rel_path` as the save response. |
| AC-2.2 | After AC-1.1, `ListObservationsBySource` for the Source | Same as AC-2.1; then `Graph().Verify()` on that catalog passes. |
| AC-2.3 | After AC-1.1, the stored `depiction_values.region_json` changed by raw SQL to a region on `artifact_page` 9; separately, the stored image File's `media_type` changed to `application/pdf` | For each damage, `GetCitation` and `ListObservationsBySource` return the same likeness row with its `value_depiction_id`. The off-page region reads `depiction.region_json` `""` with the image path kept; the PDF media type reads `depiction.image_rel_path` `""` with the region kept (`depictionvalues.Readable`). |

**US-3.** As the app, when I edit a likeness, I send a new region and a new crop beside the row, or I keep the value as it is.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | After AC-1.1, `UpdateObservation` with the response row as `observation` and the request's `depiction_input` = { a new region on page 1, a 200 × 200 JPEG of a different gray } | The response has a new `value_depiction_id` (S10-32 D-9: a new crop replaces the value row), the new region, and a new `image_rel_path`. The old path's file is gone from disk; the new one exists. |
| AC-3.2 | After AC-1.1, `UpdateObservation` with the response row, `polarity` `negative`, no `depiction_input` | Saves. `value_depiction_id`, region, and `image_rel_path` are unchanged and the file is still on disk. |
| AC-3.3 | After AC-1.1, `UpdateObservation` with the response row whose `depiction.image_rel_path` is changed to the Artifact's own `objects/…` path, no `depiction_input` | Saves with the image unchanged: an `Observation`'s `depiction` is response data and is never read from a request. |

**US-4.** As a researcher, the app names the depiction type and explains a refused crop.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `SubjectPropertyValueType.label("depiction")` and `PropertyValueType.depiction.label` | Each resolves to *Depiction*. `researcherCreatable` does not contain `depiction`. `PropertyValueType(rawValue: "depiction")` is `.depiction`. An unknown raw value still labels as *Text*. |
| AC-4.2 | `L10n.Errors.message(code: "ingest.unsupported_image")` and `(code: "depictionvalues.invalid")` | Each returns its own copy (§6 Strings), not the unknown-error copy. |
| AC-4.3 | `GoStore.mapObservation` of a proto `Observation` with `value_depiction_id` and `depiction` set; `GoStore.mapObservationDraft` with `depiction` set; `GoStore.updateObservationRequest` with a `CatalogObservation` that has `depiction` set and a `depictionInput` | Swift carries `valueDepictionID` and `depiction`; the draft proto carries `depiction`; the update request carries `observation.value_depiction_id` and the request's `depiction_input`, and leaves `observation.depiction` unset. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `engine.proto`: `DepictionValueInput`, `DepictionValue`, `ObservationDraft.depiction`, `Observation.value_depiction_id` and `.depiction`, `UpdateObservationRequest.depiction_input`; generated Go and Swift | STEP-1, D-1 |
| `files.RelPaths`, and `sourceCoverThumbnails` moved onto it; the source graph's depiction region and image, through `depictionvalues.Readable` (`graphcache.SourceObservation`, `sourcefill`) | STEP-2, D-4, D-6, AC-2.2, AC-2.3 |
| Handler input mapping; `ingest.StageDepictions` in `CreateCitationWithObservations`, `AddObservationsToCitation`, `UpdateObservation`, `CreateCitedBridge`; delete the unused `observationsToInputs` | STEP-3, D-2, D-3, D-10, AC-1.2, AC-1.4, AC-3.3 |
| Responses carry the saved value and its `objects/…` path; create and append return listed rows through one helper the connect handler also uses | STEP-4, D-4, D-5, AC-1.1, AC-1.3, AC-2.1, AC-2.2, AC-2.3, AC-3.1, AC-3.2 |
| Swift `PropertyValueType.depiction` and its `label`, `SubjectPropertyValueType.label` delegating to it, the vocabulary parity entry, error copy for `ingest.unsupported_image` and `depictionvalues.invalid` | STEP-5, D-8, D-9, AC-4.1, AC-4.2 |
| Swift store types, the `updateObservation` crop parameter, and `GoStore` mapping | STEP-6, D-7, AC-4.3 |
| Docs, and the `add-ffi-handler` section on client paths and object paths | STEP-7, D-11 |

| Out | Home |
| --- | --- |
| Offering a depiction Property in the composer, rendering the crop, the row that shows the image, a display string for a likeness on graph cards and composer rows (`FillListedDisplayText` gives a likeness empty text) | S10-34 |
| Promote's exhibit images (a `DepictionValue` on `PromoteGraphAlignmentComparison`, paths through `files.RelPaths`) | S11-20 |
| Reconciling depictions; the gallery | S10-52, S10-35 |
| `depictionvalues.LookupManyTx` | S10-52, if its loader needs a batch read; this PR reads the value with a join (D-6) |
| A colour for the `depiction` Properties pill (it uses the default `PVColor.textSecondary`) | S10-34's brief, S10-D15 |
| `FakeStore` producing depiction values | S10-34, when a UI test first writes one |
| `researcherCreatable` typed as `[PropertyValueType]` | Logged in the review log (M-3); not planned |
| Ingest copy written for a crop rather than a file import (`ingest.symlink`, `ingest.missing`, `ingest.too_large` keep their import wording) | Not planned; a crop path is written by the app, so these reach the researcher only on an app bug |
| The File id of a depiction image in any response | Not planned; no reader needs it (S10-35 reads the reconciler cache) |

## 4. Current state

Verified at `233e87ad`, after S10-32 landed. Rows marked *(S10-32)* are what S10-32 added.

| Area | Where | Today |
| --- | --- | --- |
| Proto | `api/proto/engine.proto` `Observation` (fields 1–17), `ObservationDraft` (fields 1–12), `NameValueInput`, `UpdateObservationRequest` (fields 1–3: `project_dir`, `user_id`, `Observation observation`), `CreateCitationWithObservationsResponse`, `AddObservationsToCitationResponse`, `CreateCitedBridgeResponse`, `Artifact.thumbnail_rel_path` | `Observation` is the row six responses return. Object paths are `objects/…` strings relative to the project folder (`thumbnail_rel_path`, `SourceFileRef.rel_path`). |
| Generation | `scripts/generate-proto.sh` | Writes `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`; needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`. |
| Draft → input | `api/ffi/handlers/observations.go` `observationToInput`, `observationDraftToInput`, `observationDraftToInputAllowEmptySubject`, `valueInput`, `optionalID`; `observationsToInputs` | One mapping per message; `valueInput` takes each scalar as a parameter. `observationsToInputs` has no caller (`rg observationsToInputs` finds only its definition). |
| Writers | `api/ffi/handlers/citations.go` `CreateCitationWithObservations`; `api/ffi/handlers/observations.go` `AddObservationsToCitation`, `UpdateObservation`; `api/ffi/handlers/connect.go` `CreateCitedBridge` | Each maps its inputs before `withProjectCatalog`, then calls `writes.Run` inside it. No staging. |
| Responses | `api/ffi/handlers/observations.go` `observationProto` (ids and scalars only), `listedObservationProto` (adds Property summary, date, name); `citations.go` and `AddObservationsToCitation` return `observationProto` rows in write order; `UpdateObservation` returns `listedObservationProto`; `connect.go` `CreateCitedBridge` reads `observations.ListByCitation` (ordered by ref) and keeps the rows it wrote | Create and append return no Property summary and no structured value. Every Swift caller sends one draft and reads `.first`. |
| Reads | `api/ffi/handlers/source_read.go` `listedProtosFromSource`, `listedFromSource`; `api/ffi/handlers/citations.go` `GetCitation`; `observations.go` `ListObservationsBySource` | Both reads project the Source graph cache through `listedFromSource`. |
| Object paths | `core/database/files/files.go` `StorageRelPath`, `LookupMany`; `api/ffi/handlers/thumbnails.go` `sourceCoverThumbnails` (batch: `files.LookupMany`, then `StorageRelPath`, skipping a File it cannot map), `thumbnailRelPath`, `existingThumbnailRelPath`; `api/ffi/handlers/artifacts.go` `artifactProto` | Each handler derives `objects/…` paths itself; there is no batch helper in `files`. |
| Source graph | `core/database/graphcache/source.go` `SourceObservation`; `core/database/sourcefill/load.go` `loadSourceObservations` (one `SELECT` over `observations` joined to `subjects`, then batch `datevalues.LookupManyTx`, `namevalues.LookupManyTx` for their codec blobs); `core/database/graphcache/verify.go` `Verify` | The cache is in memory; the source check compares the whole graph. `SourceGraph` holds no Citations, so `listedFromSource` has no Citation locator to pass to `Readable`. |
| Staging | `core/ingest/stage.go` `StageImage` (PNG or JPEG in, PNG out; `ErrUnsupportedImage` on a failed decode), `InsertTx`, `Discard`; `core/ingest/file.go` `openChecked` (empty or relative path → `ErrInvalid`, symlink → `ErrSymlink`) | Landed in S10-53. |
| S10-32 *(S10-32)* | `ingest.StageDepictions(c, inputs) (discard func(), err error)` in `core/ingest/depictions.go` (skips an input with no `Depiction`); `observations.DepictionInput{RegionJSON, ImagePath, Image StagedFile}` in `core/database/observations/depiction.go`; `observations.Input.Depiction`, `.ValueDepictionID`; `observations.Observation.ValueDepictionID`; `observations.Listed.Depiction *depictionvalues.Value`, filled in `scanListed` from `sqlListSelect`'s `LEFT JOIN depiction_values dpv`, `LEFT JOIN citations dc`, `LEFT JOIN files dpf` and passed through `depictionvalues.Readable` (set only when `dpv.image_file_id` is not NULL); `depictionvalues.Value{ID, RegionJSON, ImageFileID}`; `depictionvalues.Readable(v Value, citationLocatorJSON, imageMediaType string) Value` (empties `RegionJSON` unless `locator.ValidateDepictionRegion` passes against the Citation's locator; empties `ImageFileID` unless the media type is `image/png`); `apperr.CodeDepictionValuesInvalid`; `graphcache.SourceObservation.DepictionID`; `sourcefill.loadSourceObservations` selects `o.value_depiction_id`; `listedFromSource` copies it into `ValueDepictionID` | S10-32 D-7, D-8, D-10, and its Landed note. An update with `Depiction` nil and `ValueDepictionID` equal to the existing value keeps it (D-9). The `citations` effects entry reloads the Source, so the source graph sees a Citation locator edit. |
| Swift value types | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType`; `macos/App/Features/Properties/PropertiesModel.swift` `SubjectPropertyValueType.label` (a `switch` on raw strings, default *Text*), `researcherCreatable`; `macos/App/Platform/L10n.swift` `L10n.Properties.valueTypeTerm`; `macos/App/Features/Catalog/SubjectTypeChrome.swift` `valueTypeForeground` (default `PVColor.textSecondary`) | No `depiction` case; the label falls back to *Text* with no compile error. No Swift `switch` over the enum is exhaustive today. |
| Swift errors | `macos/App/Platform/L10n.swift` `L10n.Errors.message(code:params:)`, `dateValuesInvalid`, `nameValuesInvalid`, `ingestSymlinkHelp` | No case for `ingest.unsupported_image` or `depictionvalues.invalid`; both fall to `unknown`. |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogObservation` (cached in `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift` `SourceGraphRows`), `CatalogObservationDraft`, `CatalogNameValuePart`, protocol `updateObservation(projectDir:userID:observation:)`, the `extension GenealogyStore` with `promoteSubject` defaults; implementers `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, and the test store in `macos/ProvenenciaTests/OnboardingModelTests.swift`; `GoStore` `mapObservation`, `mapCatalogObservation`, `mapObservationDraft` (all `private static`), `mapWorkspaceLocationFromProto` (internal `static`, used by tests) | Two callers of `updateObservation`. |
| Parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "value types" | S10-32 left it unchanged; this PR adds the entry with the Swift case (S10-32 D-3). |
| Skill | `.agents/skills/add-ffi-handler/SKILL.md` | Says nothing about client file paths, staging, or returning `objects/…` paths. |
| Tests | `api/ffi/handlers/citations_test.go` `citationFixture`, `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, `TestListObservationsBySource`, `TestGetCitation`, `TestUpdateObservation`; `api/ffi/handlers/connect_test.go` `connectFixture`, `TestCreateCitedBridge` "creates participation"; `api/ffi/handlers/thumbnails_test.go` `writePNG`, `TestListSourcesAndWorkspaceThumbnails`; `api/ffi/handlers/subjects_test.go` `subjectFixture` (fourth result is the person type id); `api/ffi/handlers/harness_test.go` `runRPC` (skips `after` when `wantErr`); `core/ingest/stage_test.go` `objectFiles`, `stagePNG`; `core/database/files/files_test.go` `TestStorageRelPath`, `TestLookupInsert`; `macos/ProvenenciaTests/PropertiesModelTests.swift`, `L10nResolutionTests.swift` | `citationFixture`'s Artifact has no File, so a depiction cannot be saved on it. |

**Precedent.** The client-path shape is `IngestArtifactFile` in `api/ffi/handlers/artifacts.go`: an absolute path in the request, the `core/ingest` call inside `withProjectCatalog` and before `writes.Run`, and the `objects/…` path back through `files.StorageRelPath`. The listed-response shape is `CreateCitedBridge` in `api/ffi/handlers/connect.go` (`observations.ListByCitation`, then keep the written ids). The batch path derivation to move into `files` is the end of `sourceCoverThumbnails`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `engine.proto`, after `NameValueInput`: `message DepictionValueInput { string region_json = 1; string image_path = 2; }` *(new)* and `message DepictionValue { string region_json = 1; string image_rel_path = 2; }` *(new)*. `ObservationDraft` gains `DepictionValueInput depiction = 13;`. `Observation` gains `string value_depiction_id = 18; DepictionValue depiction = 19;`. `UpdateObservationRequest` gains `DepictionValueInput depiction_input = 4;`. Comments in §6. | An input names a file the app wrote; a saved value names an object Go stored. `Observation` is the saved row six responses return, so it carries only saved data; the update's new crop sits on the request, beside the row (review M-2). | One message for both — every reader has to know which half is set. The input on `Observation` — a temp path on the cached row type. A `bytes` field — working rule "bytes never cross the FFI". An `image_file_id` field — Go never takes a File id from the client, and no reader needs one (Out). |
| D-2 | `func depictionInput(d *engine.DepictionValueInput) *observations.DepictionInput` *(new)* in `api/ffi/handlers/observations.go`: nil → nil; otherwise `&observations.DepictionInput{RegionJSON: strings.TrimSpace(d.GetRegionJson()), ImagePath: strings.TrimSpace(d.GetImagePath())}`. `observationDraftToInputAllowEmptySubject` sets `in.Depiction = depictionInput(d.GetDepiction())`. `observationToInput` sets `in.ValueDepictionID` from `optionalID(o.GetValueDepictionId())` and does not read `o.GetDepiction()`. `UpdateObservation` sets `input.Depiction = depictionInput(req.GetDepictionInput())` after `observationToInput`. `valueInput` is not changed. A draft has no `value_depiction_id`. | The handler only maps; S10-32's write refuses every wrong mix (its D-8, D-9): an empty region is `locator.invalid`, an empty or relative path is `ingest.invalid` from staging. | Adding two parameters to `valueInput` — it already takes nine, and a depiction is not a scalar. Validating in the handler — a second copy of S10-32's rules. |
| D-3 | Each of the four writers stages inside `withProjectCatalog`, before `writes.Run`, with S10-32's call and nothing else: `discard, err := ingest.StageDepictions(c, inputs); if err != nil { return err }; defer discard()`. `UpdateObservation` passes `[]observations.Input{input}`. `CreateCitedBridge` passes `inputs` before they go into `connect.CreateInput`. | S10-32 D-7 makes this the one core entry point and requires it outside `writes.Run` and inside `catalogsession.Do` (S10-53); `Input.Depiction` is a pointer, so the staged image travels inside the same inputs and `connect`'s copy of them. A bridge draft is staged too, so no writer path reaches the write with an unstaged path (map row: every handler). | A shared wrapper around `writes.Run` — each handler has its own `Op` and result type (S10-32 D-7 rejected it). Skipping `connect` because no bridge type binds `likeness` — the binding could change; staging is the guarantee. |
| D-4 | `func RelPaths(c *database.Catalog, ids [][]byte) (map[string]string, error)` *(new)* in `core/database/files/files.go`, beside `LookupMany`: one `LookupMany`, then `StorageRelPath` per File; keyed by `string(id)`; an id with no row, or a row `StorageRelPath` refuses, is omitted (the rule `sourceCoverThumbnails` follows today). `sourceCoverThumbnails` replaces its `files.LookupMany` loop with `files.RelPaths` over the derived ids, with no change in what it returns. In handlers, `func depictionPaths(c *database.Catalog, rows []observations.Listed) (map[string]string, error)` *(new)* in `api/ffi/handlers/observations.go` collects each non-empty `l.Depiction.ImageFileID`, calls `files.RelPaths` once, and returns its map and error. An id `Readable` emptied is not collected, and an id the map omits gives an empty path: the region still shows and the image does not, the same rule `Readable` applies. `listedObservationProto(l observations.Listed, paths map[string]string) *engine.Observation` gains `paths` and sets `o.Depiction = &engine.DepictionValue{RegionJson: l.Depiction.RegionJSON, ImageRelPath: paths[string(l.Depiction.ImageFileID)]}` when `l.Depiction` is set (a nil or omitted id reads `""` from the map). `observationProto` sets `ValueDepictionId: uuidString(o.ValueDepictionID)`. Every caller returns `depictionPaths`'s error unchanged. | One home for "File ids → `objects/…` paths" in the package that owns `files`; S10-35 and S11-20 call it instead of a third and fourth loop (review M-1). One query per response, not one per row. The path is derived at the edge, so the cache and `Listed` hold only catalog values. An empty path in place of an error follows S10-32's landed `Readable`: a stored value Go would not show reads with that half empty, and one bad row does not fail a whole Source read (this replaces the first plan's `sql.ErrNoRows`, review X-5). | A handler-only helper typed on `[]observations.Listed` — S11-20's Promote rows are not `Listed`. Storing the path in `Listed` or the graph cache — a derived value in two stores. A `files.Lookup` per row — N queries on a Source read. Moving `thumbnailRelPath` and `artifactProto` (single-File) onto it — no gain for one id; not in this PR. |
| D-5 | `func listedWritten(c *database.Catalog, citationID []byte, written []observations.Observation) ([]*engine.Observation, error)` *(new)* in `api/ffi/handlers/observations.go`: `observations.ListByCitation(c, citationID)`, keep the rows whose id is in `written`, return them in `written` order (draft order, as create and append return today; `ListByCitation` sorts by ref). `CreateCitedBridge` moves from ref order to draft order; no caller reads its order (its test counts rows, and Swift reads the rows by id), with `depictionPaths` and `listedObservationProto`. `CreateCitedBridge` replaces its inline filter with it; `CreateCitationWithObservations` and `AddObservationsToCitation` return it in place of `observationProto` rows. | Map Testable: "a save … returns its path". `observations.Observation` holds only the value id, so the path needs the listed read the connect handler already does; one helper gives every write response one shape and keeps the order callers get today. | A depiction-only lookup in the two handlers — a third response shape. Ref order — changes create and append order for no reader's benefit (review X-4). |
| D-6 | `graphcache.SourceObservation` gains `DepictionRegion string` and `DepictionImageFileID []byte` *(new)* after S10-32's `DepictionID`, holding the value as `depictionvalues.Readable` returns it. `loadSourceObservations`'s `SELECT` gains `LEFT JOIN depiction_values dpv ON dpv.id = o.value_depiction_id`, `LEFT JOIN citations dc ON dc.id = o.citation_id`, and `LEFT JOIN files dpf ON dpf.id = dpv.image_file_id` (the three joins `sqlListSelect` uses) and selects `dpv.region_json, dpv.image_file_id, dc.locator_json, dpf.media_type` (each `sql.NullString` or `[]byte`). `func FromJoin(id []byte, region sql.NullString, imageFileID []byte, citationLocator, imageMediaType sql.NullString) *Value` *(new)* in `core/database/depictionvalues/depictionvalues.go`, beside `Readable`: nil when `imageFileID` is empty; otherwise a `Value` with copied `id` and `imageFileID` slices and `region.String`, passed through `Readable(v, citationLocator.String, imageMediaType.String)`. Its doc comment names the three joins a caller makes. `scanListed` in `core/database/observations/observations.go` replaces its `if len(depictionImage) > 0 { … }` block with `l.Depiction = depictionvalues.FromJoin(depictionID, depictionRegion, depictionImage, citationLocator, imageMediaType)`, with no change in what it returns. The `sourcefill` loop calls the same function and, when the result is non-nil, copies its `RegionJSON` and `ImageFileID` into the two fields; `sourcefill` imports `depictionvalues` beside `datevalues`. `listedFromSource` sets `l.Depiction = &depictionvalues.Value{ID: o.DepictionID, RegionJSON: o.DepictionRegion, ImageFileID: o.DepictionImageFileID}` when `len(o.DepictionID) > 0` (the same fact `FromJoin` gates on: `depiction_values.image_file_id` is `NOT NULL`, migration `000052.sql`). | One function turns the joined row into a readable value, so both Observation projections show a depiction one way and a later `Readable` input changes one place (refresh review M-1) (review M-4; S10-32 Landed: "later readers of a depiction go through `Readable`"). `Readable` needs the Citation's locator, and `SourceGraph` holds no Citations, so it runs at load, not in `listedFromSource`. The `citations` and `depiction_values` effects entries reload the Source, so the stored result follows a locator or value edit; `Verify` reloads and compares the whole struct. A region and an id need no codec, unlike the date and name batches. The cache is in memory: no version to bump. | `depictionvalues.LookupManyTx` and a third batch pass — a function, a test, and a loop the join does not need (Out). Storing the raw value and the Citation locator in the cache, and calling `Readable` in `listedFromSource` — a locator copied per row for one check. A `valuecodec` blob — no codec is needed. A `*depictionvalues.Value` field — `graphcache` holds plain fields and imports no value package. |
| D-7 | Swift, in `GenealogyStore.swift` beside `CatalogNameValuePart`: `struct CatalogDepictionValueInput: Sendable, Equatable { var regionJSON: String; var imagePath: String }` and `struct CatalogDepictionValue: Sendable, Equatable { var regionJSON: String; var imageRelPath: String }` *(new)*. `CatalogObservation` gains, after `valueTermKey`, `var valueDepictionID: String = ""` and `var depiction: CatalogDepictionValue? = nil`. `CatalogObservationDraft` gains, after `notes`, `var depiction: CatalogDepictionValueInput? = nil`. The protocol requirement becomes `updateObservation(projectDir:userID:observation:depictionInput: CatalogDepictionValueInput?)`; the existing `extension GenealogyStore` gains `updateObservation(projectDir:userID:observation:)` calling it with `nil`, so the two callers do not change. `GoStore`, `FakeStore`, and the `OnboardingModelTests.swift` store implement the new requirement (`FakeStore` and the test store ignore `depictionInput`). In `GoStore`: `mapObservation` reads `valueDepictionID`, and `depiction` from `o.depiction` when `o.hasDepiction`, else nil; `mapCatalogObservation` writes `valueDepictionID` and never writes `depiction`; `mapObservationDraft` writes `depiction`; `static func updateObservationRequest(projectDir: String, userID: String, observation: CatalogObservation, depictionInput: CatalogDepictionValueInput?) -> Provenencia_Engine_V1_UpdateObservationRequest` *(new)* builds the request, and `updateObservation` sends it. `mapObservation` and `mapObservationDraft` drop `private` (stay `static`), as `mapWorkspaceLocationFromProto` is, so a test can call them. | The cached row type holds only saved data; S10-34 passes the crop beside the row instead of on a copy of a cached row (review M-2). Defaults keep every existing initializer and caller compiling. | `CatalogObservation.depictionInput` — a temp path on the session cache's row type. A `FakeStore` depiction path — no UI writes one until S10-34 (Out). |
| D-8 | `PropertyValueType` gains `case depiction` after `term` and `var label: LocalizedStringResource` *(new)*, an exhaustive `switch` over every case returning the existing `L10n.Properties.valueType…` keys and `L10n.Properties.valueTypeDepiction` *(new key `properties.valueType.depiction`, "Depiction", comment "Property value type label: depiction (registry-only)")*. `SubjectPropertyValueType.label(_:)` becomes `PropertyValueType(rawValue: valueType)?.label ?? L10n.Properties.valueTypeText`. Its doc comment reads "Schema also has `term` and `depiction` (registry-only) — never offered here." `researcherCreatable` is unchanged. `CatalogProperty.valueType`'s comment lists `depiction`. `TestSwiftVocabularyMatchesGo` "value types" adds `properties.ValueTypeDepiction`. `valueTypeForeground` is unchanged. | S10-32 D-3 assigns the Swift case and the parity entry to this PR, together. The labels live on the enum the parity test checks, so the next value type does not compile without one (review M-3). | A second `case "depiction"` in the string switch — the list stays untested. A pill colour — chrome; S10-34's brief (Out). |
| D-9 | Error copy in `L10n.Errors`: `ingestUnsupportedImage` *(new)*, key `error.ingest.unsupported_image`, "The picture of that region couldn’t be read.", comment "FFI error ingest.unsupported_image (a depiction crop Go could not decode)"; `depictionValuesInvalid` *(new)*, key `error.depictionvalues.invalid`, "Invalid depiction value.", comment "FFI error depictionvalues.invalid". `message(code:params:)` gains `case "ingest.unsupported_image"` beside the other `ingest.` cases and `case "depictionvalues.invalid"` after `namevalues.invalid`. `ingestCallout` is not changed. | `add-ffi-handler`: every user-facing code has Swift copy. `ingest.unsupported_image` comes only from `StageImage`, which only a depiction save calls, so its copy names the picture, not a file the researcher chose. `depictionvalues.invalid` follows `dateValuesInvalid`. | An Artifact-import callout — the crop is not an import. |
| D-10 | Delete `observationsToInputs` from `api/ffi/handlers/observations.go`. | It has no caller (§4), and D-2 edits the mapping beside it. | Leaving it — the next value field would be added to a dead copy. |
| D-11 | `add-ffi-handler` gains a section "Files from the client, paths to the client": a request names a client file by absolute path, never bytes; stage it inside `withProjectCatalog` before `writes.Run` (`ingest.StageDepictions` for depictions, `ingest.File` for Artifacts) and `defer discard()`; never take a File id from the client; return `objects/…` paths through `files.RelPaths` (many) or `files.StorageRelPath` (one). | This PR is the second instance after `IngestArtifactFile`, and S10-34, S10-35, and S11-20 each return image paths (review M-5). | Leaving it to precedent — the next three PRs would reverse-engineer two handlers. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `DepictionValueInput`, `DepictionValue` *(new)* | proto message | `NameValueInput`, `DateValueInput` | Those are values the client composes; a depiction input is a path and a region, and its saved form is a different pair (D-1). |
| `ObservationDraft.depiction`, `Observation.value_depiction_id`, `.depiction`, `UpdateObservationRequest.depiction_input` *(new)* | proto field | `date`, `value_date_id` | Extends the same messages (D-1). |
| `files.RelPaths` *(new)* | func | `sourceCoverThumbnails`'s loop, `files.LookupMany`, `files.StorageRelPath` | Moves that loop into `files`; thumbnails and depictions both use it (D-4). |
| `depictionInput` *(new)* | func | `valueInput` | `valueInput` maps scalars; this maps one message to a pointer (D-2). |
| `depictionPaths` *(new)* | func | `files.RelPaths` | Adds only the id collection from `Listed`; a File `RelPaths` omits gives an empty path (D-4). |
| `listedWritten` *(new)* | func | the filter inline in `CreateCitedBridge` | Moves that filter out so three handlers share it (D-5). |
| `SourceObservation.DepictionRegion`, `.DepictionImageFileID` *(new)* | field | `Date`, `Name` | Plain fields, read by the joins `scanListed` uses and filled by `FromJoin` (D-6). |
| `depictionvalues.FromJoin` *(new)* | func | `scanListed`'s depiction block, `depictionvalues.Readable` | Moves that block into `depictionvalues` so `scanListed` and `sourcefill` share it (D-6). |
| `CatalogDepictionValueInput`, `CatalogDepictionValue` *(new)* | Swift struct | `CatalogNameValuePart`, `CatalogDateValueInput` | Mirror D-1. |
| `PropertyValueType.depiction`, `PropertyValueType.label`, `L10n.Properties.valueTypeDepiction` *(new)* | enum case / property / key | `SubjectPropertyValueType.label`, `valueTypeTerm` | The string switch becomes a lookup into the enum (D-8). |
| `GoStore.updateObservationRequest` *(new)* | Swift func | the request built inline in `GoStore.updateObservation` | Extracted so a test can check it (D-7). |
| `L10n.Errors.ingestUnsupportedImage`, `depictionValuesInvalid` *(new)* | key | `ingestSymlinkHelp`, `nameValuesInvalid` | Same switch. |
| Test helpers `likenessFixture`, `objectRelPaths` in `api/ffi/handlers/citations_test.go`, `writeCrop` in `api/ffi/handlers/thumbnails_test.go` *(new)* | helper | `citationFixture`, `writePNG`; `core/ingest/stage_test.go` `objectFiles` (catalog in, absolute paths out), `stagePNG` (bytes, not a file) | `citationFixture`'s Artifact has no File and its subject is a place; `writePNG` writes one fixed PNG; the ingest helpers live in another package and return other shapes, so the handler helper takes a different name (review M-6). |
| `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` *(new)* | test | `WorkspaceNavigationTests` (calls a `GoStore` map) | No test covers Observation mapping today. |

**Builds on.** S10-32's `ingest.StageDepictions`, `Input.Depiction`, `Listed.Depiction`, `depictionvalues.Readable`, its `sqlListSelect` joins, and its effects entry; S10-53's staging errors; the `IngestArtifactFile` path-in, rel-path-out shape; `add-ffi-handler`.
**Deletes.** `observationsToInputs`; `scanListed`'s inline depiction block; the inline written-id filter in `CreateCitedBridge`; the File-to-path loop in `sourceCoverThumbnails`; the string `switch` in `SubjectPropertyValueType.label`.
**Next adds.** S10-34 sets `CatalogObservationDraft.depiction` from the composer, passes `depictionInput` to `updateObservation`, and shows `CatalogObservation.depiction.imageRelPath`. S11-20 adds a `DepictionValue` to `PromoteGraphAlignmentComparison` and resolves its paths with `files.RelPaths`. S10-35 resolves gallery paths with `files.RelPaths`. A future depiction key (place, event) needs no FFI change.

### Schema

None — no catalog change. S10-32 adds `depiction_values`.

### Go core

| Package | Change |
| --- | --- |
| `core/database/files` | `RelPaths` (D-4). |
| `core/database/depictionvalues` | `FromJoin` (D-6). |
| `core/database/observations` | `scanListed` calls `FromJoin` (D-6). |
| `core/database/graphcache` | `SourceObservation` fields (D-6). |
| `core/database/sourcefill` | `loadSourceObservations` joins, calls `depictionvalues.FromJoin`, and fills them (D-6). |

```go
// core/database/files/files.go (sketch)
// RelPaths returns the objects/… path of each File, keyed by id. An id with
// no row, or whose path cannot be derived, is omitted.
func RelPaths(c *database.Catalog, ids [][]byte) (map[string]string, error)
```

Errors: none new in Go. A save can now return `ingest.invalid`, `ingest.missing`, `ingest.symlink`, `ingest.not_a_file`, `ingest.empty`, `ingest.too_large`, `ingest.unsupported_image`, `locator.invalid`, `observations.invalid`, `depictionvalues.invalid`.

### FFI

`api/proto/engine.proto`, no new `Method`:

```proto
// DepictionValueInput is a depiction to save: a region locator on the
// Citation's page and the absolute path of the PNG or JPEG the app rendered
// from that region. Go checks and re-encodes it; the file's bytes are never
// stored and never cross the FFI.
message DepictionValueInput {
  string region_json = 1;
  string image_path = 2;
}

// DepictionValue is a saved depiction: its region locator and the objects/…
// path, relative to the project folder, of the PNG Go stored. Either is empty
// when the stored half is not one Go would show (a region not on the
// Citation's page, an image that is not a stored PNG).
message DepictionValue {
  string region_json = 1;
  string image_rel_path = 2;
}

// in ObservationDraft
  DepictionValueInput depiction = 13;

// in Observation
  string value_depiction_id = 18;
  // Response data; never read from a request.
  DepictionValue depiction = 19;

// in UpdateObservationRequest
  // A new region and its crop for the depiction Observation. Unset keeps the
  // value named by observation.value_depiction_id.
  DepictionValueInput depiction_input = 4;
```

Handlers: `api/ffi/handlers/observations.go` (D-2, D-3, D-4, D-5, D-10), `api/ffi/handlers/citations.go` (D-3, D-5), `api/ffi/handlers/connect.go` (D-3, D-5), `api/ffi/handlers/source_read.go` (D-4, D-6), `api/ffi/handlers/thumbnails.go` (D-4). Regenerate with `./scripts/generate-proto.sh`.

### macOS

No view changes.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value type and label | `PropertyValueType.depiction`, `PropertyValueType.label` | `macos/App/Platform/InterpretationValueKinds.swift` | D-8 |
| Label lookup | `SubjectPropertyValueType.label` | `macos/App/Features/Properties/PropertiesModel.swift` | D-8; the pill's VoiceOver label |
| Store types and protocol | `CatalogDepictionValueInput`, `CatalogDepictionValue`, fields, `updateObservation` | `macos/App/Platform/GenealogyStore.swift` | D-7 |
| Mapping | `mapObservation`, `mapCatalogObservation`, `mapObservationDraft`, `updateObservationRequest` | `macos/App/Platform/GoStore.swift` | D-7 |
| Error copy | `L10n.Errors` | `macos/App/Platform/L10n.swift` | D-9 |

States: none; nothing renders a depiction. Accessibility: the Properties pill for `likeness` reads *Depiction*.

### Strings

L10n keys *(new)*, each in `L10n.swift` and `macos/App/Resources/Localizable.xcstrings`:

- `properties.valueType.depiction` = "Depiction"
- `error.ingest.unsupported_image` = "The picture of that region couldn’t be read."
- `error.depictionvalues.invalid` = "Invalid depiction value."

Removed: none.

## 7. Steps

**STEP-1 — Proto**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: D-1, the §6 FFI block exactly. Run `./scripts/generate-proto.sh` and commit both generated files. The script needs `protoc`, `protoc-gen-go`, and `protoc-gen-swift`; never hand-edit a generated file. Without them, stop at this step (§11).
Check: `CGO_ENABLED=1 go build ./...` succeeds; `git diff --stat` shows only these three files.

**STEP-2 — Core reads: File paths and the source graph's depiction**
Files: `core/database/files/files.go`, `core/database/files/files_test.go`, `core/database/depictionvalues/depictionvalues.go`, `core/database/depictionvalues/depictionvalues_test.go`, `core/database/observations/observations.go`, `api/ffi/handlers/thumbnails.go`, `core/database/graphcache/source.go`, `core/database/sourcefill/load.go`. Skill: `add-catalog-query`.
Do: D-4's `files.RelPaths` and the `sourceCoverThumbnails` move; D-6 except `listedFromSource` (STEP-4), including `FromJoin` and moving `scanListed` onto it. Add T-1 and T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/... ./api/ffi/handlers/...` passes, `TestListSourcesAndWorkspaceThumbnails` and the `observations` likeness rows unedited.

**STEP-3 — Map and stage depiction inputs**
Files: `api/ffi/handlers/observations.go`, `api/ffi/handlers/citations.go`, `api/ffi/handlers/connect.go`, `api/ffi/handlers/citations_test.go`, `api/ffi/handlers/thumbnails_test.go`, `api/ffi/handlers/connect_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-2, D-3, D-10. Add the test helpers:
- In `thumbnails_test.go`, beside `writePNG`: `writeCrop(t, w, h int, shade uint8) string`, a `w` × `h` JPEG (quality 90) filled with opaque gray `shade`, written to `t.TempDir()` as `crop.jpg`; returns its absolute path. A JPEG, so the stored PNG cannot be the crop's bytes.
- In `citations_test.go`: `likenessFixture(t) (dir, userID, sourceID, artifactID, personID, likenessID string)`: `subjectFixture`; `CreateArtifact`; `writePNG` to `filepath.Join(t.TempDir(), "photo.png")` and `IngestArtifactFile` with that path; `CreateSubject` with `subjectFixture`'s fourth result (the person type id), label "Ada"; `ListProperties` and the id whose key is `likeness`.
- In `citations_test.go`: `objectRelPaths(t, dir string) []string`, every regular file under `dir/objects`, sorted, as paths relative to `dir`.
- In `citations_test.go`: constants `likenessLocatorJSON` = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}` and `likenessRegionJSON` = `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`.
Add T-3 and T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-4 — Return the saved value and its path**
Files: `api/ffi/handlers/observations.go`, `api/ffi/handlers/citations.go`, `api/ffi/handlers/connect.go`, `api/ffi/handlers/source_read.go`, `api/ffi/handlers/citations_test.go`. Skill: `add-ffi-handler`.
Do: D-4's handler half, D-5, and D-6's `listedFromSource` line. `listedProtosFromSource` collects its `Listed` rows into a slice, calls `depictionPaths` once, then builds the protos. `UpdateObservation` calls `depictionPaths(c, []observations.Listed{row})`. Add T-2, T-4, T-5, T-6, T-12. The existing rows of `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, and `TestCreateCitedBridge` pass unedited.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-5 — Swift value type, label, and error copy**
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Platform/GenealogyStore.swift` (the `CatalogProperty.valueType` comment only; the rest of this file is STEP-6), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/PropertiesModelTests.swift`, `macos/ProvenenciaTests/L10nResolutionTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-8, D-9. Add T-8, T-9, T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `python3 scripts/check-localizable-xcstrings.py` passes; on macOS, the `xcodebuild test` command in `scripts/check-pr.sh` passes. Without a Mac, report the Xcode check as not run.

**STEP-6 — Swift store types and mapping**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-7. Add T-11. Register the test file in the ProvenenciaTests target only, per `add-swift-test` "Xcode project": a `PBXFileReference` with id `AA0000000000000000000A12` in group `AA00000000000000000000B5`, and a `PBXBuildFile` with id `AA0000000000000000000A13` in Sources phase `AA00000000000000000000B7`. Both ids were unused at the planning base; if either is taken, use the next unused `AA…` id.
Check: on macOS, the `xcodebuild test` command in `scripts/check-pr.sh` passes. Without a Mac, report this check as not run.

**STEP-7 — Docs**
Files: `.agents/skills/add-ffi-handler/SKILL.md`, `docs/interpretation-layer-data-model.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `add-ffi-handler`: D-11's section, after "Catalog RPCs", citing `IngestArtifactFile` and `CreateCitationWithObservations` as the two examples.
- `interpretation-layer-data-model.md`: after S10-32's `value_type = 'depiction'` paragraph add "Over the FFI a save sends `DepictionValueInput` (the region and the absolute path of the app's rendering); Go stages it before the write. Responses and reads return `DepictionValue`: the region and the image's `objects/…` path. The client never sends or receives the image's File id."
- `deployment-plan.md`: the S10-54 entry's **Plan:** line (added when this plan was written) ends "— Landed #N"; the SLICE 8 line for S10-54 gains "— done, #N". `completed.md`: a `## S10-54 — Depiction values over the FFI` section above the current first section, in the S10-50 shape.
- This plan: Status and a **Landed** note per the DoD.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-54-depiction-ffi.md` passes.

STEP-3 and STEP-4 belong together: after STEP-3 a likeness saves but the create response has no path. Push only after STEP-4 passes. Open the draft PR before STEP-7, so its number is known.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/files/files_test.go` `TestRelPaths` *(new)*, table-driven | two Files inserted with `Insert` in one transaction, as `TestLookupInsert` does (`image/png` with checksum `aa…`, `application/pdf` with `bb…`), and one unknown 16-byte id → a map of the two, each equal to `StorageRelPath` of its row; no ids → empty map | D-4 |
| T-2 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationWithObservations` row "saves a likeness and returns its image path" *(new)* | `likenessFixture`; draft with `likenessRegionJSON` and `writeCrop(t, 300, 200, 128)` → AC-1.1 fields; `objectRelPaths` gains exactly the returned path; `png.DecodeConfig` on that file gives 300 × 200 | AC-1.1, D-1, D-3, D-4, D-5 |
| T-3 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationRefusesSymlinkedCrop` *(new)*, a plain test (`runRPC` skips `after` on an error) | `likenessFixture`; record `objectRelPaths`; AC-1.1 request with `image_path` an `os.Symlink` to `writeCrop(t, 300, 200, 128)` → `errors.Is(err, ingest.ErrSymlink)`; `objectRelPaths` unchanged; `ListCitationsByArtifact` for the Artifact returns no Citation; `t.Cleanup` calls `catalogsession.CloseAll()` | AC-1.2, D-3 |
| T-4 | Go | `TestAddObservationsToCitation` row "adds a likeness" *(new)* | `likenessFixture`; `CreateCitationWithObservations` with `likenessLocatorJSON` and no drafts; `AddObservationsToCitation` with the AC-1.1 draft → the AC-1.1 fields | AC-1.3, D-5 |
| T-5 | Go | `TestGetCitation` row "returns a likeness" and `TestListObservationsBySource` row "returns a likeness" *(new)* | after the T-2 save → AC-2.1 fields; for `ListObservationsBySource`, then `catalogsession.Do(dir, func(c *database.Catalog) error { return c.Graph().Verify() })` → nil | AC-2.1, AC-2.2, D-6 |
| T-12 | Go | `api/ffi/handlers/citations_test.go` `TestLikenessReadHidesDamagedValue` *(new)*, table-driven, a plain test like T-3 | each case: `likenessFixture` and the T-2 save; damage by raw SQL inside `catalogsession.Do(dir, …)` on `c.DB()`, as `core/database/observations/observations_test.go` "likeness read drops a stored region that is not valid" does; then `catalogsession.CloseAll()` so the next read loads the Source fresh. Case "region off the page": `UPDATE depiction_values SET region_json = ?` with `likenessRegionJSON` on `artifact_page` 9 → `GetCitation` and `ListObservationsBySource` rows keep `value_depiction_id` and `image_rel_path`, `depiction.region_json` `""`. Case "image not a PNG": `UPDATE files SET media_type = 'application/pdf'` for the image File → region kept, `image_rel_path` `""`. Each case then `Graph().Verify()` → nil; `t.Cleanup` calls `catalogsession.CloseAll()` | AC-2.3, D-4, D-6 |
| T-13 | Go | `core/database/depictionvalues/depictionvalues_test.go` `TestFromJoin` *(new)*, table-driven | empty `imageFileID` → nil; a valid region on the Citation's page with `image/png` → the value with both halves; an off-page region → `RegionJSON` `""`; `image/jpeg` → `ImageFileID` nil; the returned slices do not alias the inputs (mutate the input after the call) | D-6 |
| T-6 | Go | `TestUpdateObservation` rows "replaces a likeness image", "keeps a likeness on a polarity change", "ignores a sent image path" *(new)* | after the T-2 save: request `depiction_input` with region points `(0.5,0.5),(0.9,0.5),(0.9,0.9)` on page 1 and `writeCrop(t, 200, 200, 64)` → a value id different from the saved one, new region, new path; old path absent from `objectRelPaths`. Polarity `negative`, no input → value id, region, path unchanged. No input and `observation.depiction.image_rel_path` set to the Artifact's own `objects/…` path → path unchanged | AC-3.1, AC-3.2, AC-3.3, D-1, D-2, D-4 |
| T-7 | Go | `api/ffi/handlers/connect_test.go` `TestCreateCitedBridge` row "stages a depiction draft" *(new)* | the "creates participation" request, built the same way from `connectFixture`, with a fourth draft `{PropertyId: personProp, Depiction: &engine.DepictionValueInput{RegionJson: likenessRegionJSON, ImagePath: link}}`, where `reqFn` makes `link` with `os.Symlink(writeCrop(t, 300, 200, 128), filepath.Join(t.TempDir(), "link.jpg"))` → `wantErrIs: ingest.ErrSymlink`. Staging runs before the write, so the Property does not matter | AC-1.4, D-3 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "value types" | Go list adds `properties.ValueTypeDepiction` → matches Swift | AC-4.1, D-8 |
| T-9 | Swift | `macos/ProvenenciaTests/PropertiesModelTests.swift` `depictionIsRegistryOnlyAndLabelled` *(new)* | `L10n.string(SubjectPropertyValueType.label("depiction")) == "Depiction"`; `L10n.string(PropertyValueType.depiction.label) == "Depiction"`; `L10n.string(SubjectPropertyValueType.label("nope")) == "Text"`; `!researcherCreatable.contains("depiction")`; `PropertyValueType(rawValue: "depiction") == .depiction` | AC-4.1, D-8 |
| T-10 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` `depictionSaveErrorsHaveCopy` *(new)* | `message(code: "ingest.unsupported_image")` == "The picture of that region couldn’t be read."; `message(code: "depictionvalues.invalid")` == "Invalid depiction value."; neither equals `message(code: "internal.unknown")` | AC-4.2, D-9 |
| T-11 | Swift | `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` *(new)* `mapsSavedDepiction`, `draftSendsDepiction`, `updateSendsInputBesideTheRow` | proto with `valueDepictionID` "v" and `depiction` (region "r", path "objects/a.png") → `CatalogObservation` with those; draft with `depiction` → proto `depiction` set; `updateObservationRequest` with a row that has `valueDepictionID` "v" and `depiction` set, and `depictionInput` (region "r2", path "/tmp/c.png") → request `depictionInput` set, `observation.hasDepiction` false, `observation.valueDepictionID` "v" | AC-4.3, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the existing create, append, and connect rows (D-5 changes their response rows), `TestListSourcesAndWorkspaceThumbnails` and the cover tests (STEP-2 moves `sourceCoverThumbnails`), `TestGetCitationSecondReadUsesTheLoadedSource` (STEP-4 batches `listedProtosFromSource`), and the source graph verify (STEP-2).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (the three §6 keys)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI in this PR)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-54 Plan line and SLICE 8 line; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: interpretation-layer-data-model.md; add-ffi-handler skill section (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestCreateCitationWithObservations|TestCreateCitationRefusesSymlinkedCrop|TestLikenessReadHidesDamagedValue|TestAddObservationsToCitation|TestGetCitation|TestListObservationsBySource|TestUpdateObservation|TestCreateCitedBridge|TestSwiftVocabularyMatchesGo|TestRelPaths' -v ./api/... ./core/database/files/...` and paste the likeness rows' output in the PR. (AC-1.1 to AC-3.3, AC-2.3)
2. On a Mac, build the branch, create a new project, and open **Properties → Person**. The Likeness row's pill reads `depiction`; with VoiceOver on, it reads *Depiction*. The composer's Property menu still has no Likeness. Open a Source with an Artifact: its cover and Artifact thumbnails show as before. (AC-4.1) An implementer without a Mac reports this check as not run.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A stored depiction Go would not show | `FromJoin`, and so `Readable`, at both reads (D-6); an empty path in place of an error (D-4); T-12. |
| The richer create and append responses change a Swift caller | `GoStore` maps both through `mapObservation`, which already reads every listed field; order stays draft order (D-5). |
| Moving `sourceCoverThumbnails` onto `files.RelPaths` changes a cover | `RelPaths` omits what the loop skipped (D-4); the thumbnail tests stay unedited. |
| A Source with many likenesses reads slowly | One `files.RelPaths` per response (D-4) and the depiction read rides the Source load's one `SELECT` (D-6). |
| The client sends a saved `depiction` pointing at another object | Never read (D-2); T-6 "ignores a sent image path". |

**Stop and report if:**

- A name in §4's *(S10-32)* row is not on `main` as written there.
- `observations.Update` does not return `Listed.Depiction` filled (`UpdateObservation` reads it for its response).
- `sourcefill` cannot import `depictionvalues` without an import cycle.
- `./scripts/generate-proto.sh` cannot run (a generator is missing), or changes any generated message other than the five D-1 names.
- An existing row of `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, or `TestCreateCitedBridge` fails after STEP-4 for a reason other than a field it now receives, or a thumbnail or cover test fails after STEP-2.
- `TestUpdateObservation` "keeps a likeness on a polarity change" fails with `observations.invalid` (S10-32 D-9's keep branch is not as planned).
- A Swift `switch` outside `PropertyValueType.label` fails to compile after STEP-5.

## 12. Open questions

None.

## 13. PR

**Title:** S10-54 — Depiction values over the FFI
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-54-depiction-ffi.md`.
