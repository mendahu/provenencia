# S10-54 — Depiction values over the FFI

**Status:** Blocked — on S10-32
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-54--depiction-values-over-the-ffi)
**Brief:** — (no new chrome: proto fields, handler wiring, Swift store types, a value-type label, and two error strings)
**Board:** —
**Depends on:** S10-32 (in progress; this plan is written against [its plan](S10-32-depiction-value-type.md) as planned)
**Blocks:** S10-34, S10-36
**Branch:** `s10-54-depiction-ffi`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) ("The image"), [S10-32 plan](S10-32-depiction-value-type.md) D-7 to D-10 and its Out rows that name S10-54, working rules "A depiction's value is its own row: region plus image", "The client renders the image; Go validates it and stores its own copy"
**Planned:** 2026-10-17 against `main` @ `ae26c629`, with [S10-32](S10-32-depiction-value-type.md) as planned (not landed)
**Review:** — (not reviewed yet)

> **Before you start.** S10-32 had not landed when this plan was written. Every name this plan marks *(S10-32)* is taken from that plan's Decisions. Read S10-32's **Landed** note first. If a name or signature it lists differs from what this plan uses, stop and report (see §11).

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

S10-32 gives the engine a `depiction` value: a region of the Citation's page plus an image File Go stored from the app's rendering of it. This PR carries that value across the FFI. A save sends the region and the absolute path of the PNG the app wrote to a temporary file; Go stages it through S10-32's one entry point (`ingest.StageDepictions`) in every handler that writes Observations, and the response returns the saved region and the image's `objects/…` path. Every read of a Citation's Observations returns the same two fields. Swift gains the `depiction` value type, its *Depiction* label, the store fields, and copy for the two errors a save can now return.

**Why now.** The composer row (S10-34) and Promote's crops (S10-36) both wait on this shape. S10-32 left the FFI, the Swift value type, and hydrating the source graph's depiction to this PR (its Out rows).

**What stays.** No chrome: the composer still offers no depiction Property, and nothing in the app writes a likeness until S10-34. Bytes never cross the FFI. Go never takes a File id from the client. Every other value type saves and reads as today; the create and append responses gain the Property summary and values the update and connect responses already carry (D-5).

## 2. User stories

**US-1.** As the app, when I save a likeness, I send its region and the path of the crop I rendered, and I get back the region and the path of the image Go stored.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new project; an Artifact whose File is an 8 × 8 PNG; a Person subject. `CreateCitationWithObservations` with locator `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`, transcription `""`, and one `ObservationDraft` on the Person with the `likeness` Property and `depiction` = { `region_json`: the S10-32 AC-1.1 region, `image_path`: a 300 × 200 opaque gray PNG in a temp folder } | It saves. The response's one Observation has `property_value_type` `depiction`, a non-empty `value_depiction_id`, `depiction.region_json` equal to the sent region, and `depiction.image_rel_path` matching `^objects/[0-9a-f]{2}/[0-9a-f]{2}/[0-9a-f]{64}\.png$`. That file exists under the project folder, and its bytes are not the crop's bytes. `depiction_input` is unset. |
| AC-1.2 | AC-1.1 with `image_path` a symlink to the crop | Refused with `ingest.symlink`. No Citation is written and no file is added to the project's objects folder. |
| AC-1.3 | AC-1.1 through `AddObservationsToCitation` on an existing Citation with that locator | Same response fields as AC-1.1. |
| AC-1.4 | `CreateCitedBridge` whose one draft carries a `depiction` with a symlinked `image_path` | Refused with `ingest.symlink`, before any write. |

**US-2.** As the app, when I read a Citation's Observations, a likeness carries its region and image path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | After AC-1.1, `GetCitation` for that Citation | The likeness row has the same `value_depiction_id`, `depiction.region_json`, and `depiction.image_rel_path` as the save response. |
| AC-2.2 | After AC-1.1, `ListObservationsBySource` for the Source | Same as AC-2.1; then `Graph().Verify()` on that catalog passes. |

**US-3.** As the app, when I edit a likeness, I send a new region and a new crop, or I keep the value as it is.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | After AC-1.1, `UpdateObservation` with the response row, `depiction_input` = { a new region on page 1, a 200 × 200 PNG of a different gray } | The response has the same `value_depiction_id`, the new region, and a new `image_rel_path`. The old path's file is gone from disk; the new one exists. |
| AC-3.2 | After AC-1.1, `UpdateObservation` with the response row, `polarity` `negative`, no `depiction_input` | Saves. `value_depiction_id`, region, and `image_rel_path` are unchanged and the file is still on disk. |
| AC-3.3 | After AC-1.1, `UpdateObservation` with the response row, no `depiction_input`, and `depiction.image_rel_path` changed to another existing object's path | Saves with the image unchanged: the saved `depiction` on a request is never read. |

**US-4.** As a researcher, the app names the depiction type and explains a refused crop.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `SubjectPropertyValueType.label("depiction")` | Resolves to *Depiction*. `researcherCreatable` does not contain `depiction`. `PropertyValueType(rawValue: "depiction")` is `.depiction`. |
| AC-4.2 | `L10n.Errors.message(code: "ingest.unsupported_image")` and `(code: "depictionvalues.invalid")` | Each returns its own copy (§6 Strings), not the unknown-error copy. |
| AC-4.3 | `GoStore.mapObservation` of a proto `Observation` with `value_depiction_id` and `depiction` set; `GoStore.mapCatalogObservation` of a `CatalogObservation` with `depiction` and `depictionInput` set; `GoStore.mapObservationDraft` with `depiction` set | Swift carries `valueDepictionID` and `depiction`; the request proto carries `value_depiction_id` and `depiction_input` and leaves `depiction` unset; the draft proto carries `depiction`. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `engine.proto`: `DepictionValueInput`, `DepictionValue`, `ObservationDraft.depiction`, `Observation.value_depiction_id`, `.depiction`, `.depiction_input`; generated Go and Swift | STEP-1, D-1 |
| `depictionvalues.LookupManyTx`; the source graph's depiction region and image (`graphcache.SourceObservation`, `sourcefill`) | STEP-2, D-6, AC-2.2 |
| Handler input mapping; `ingest.StageDepictions` in `CreateCitationWithObservations`, `AddObservationsToCitation`, `UpdateObservation`, `CreateCitedBridge`; delete the unused `observationsToInputs` | STEP-3, D-2, D-3, D-10, AC-1.2, AC-1.4, AC-3.3 |
| Responses carry the saved value and its `objects/…` path; create and append return listed rows through one helper the connect handler also uses | STEP-4, D-4, D-5, AC-1.1, AC-1.3, AC-2.1, AC-2.2, AC-3.1, AC-3.2 |
| Swift `PropertyValueType.depiction`, the *Depiction* label, the vocabulary parity entry, error copy for `ingest.unsupported_image` and `depictionvalues.invalid` | STEP-5, D-8, D-9, AC-4.1, AC-4.2 |
| Swift store types and `GoStore` mapping | STEP-6, D-7, AC-4.3 |
| Docs | STEP-7 |

| Out | Home |
| --- | --- |
| Offering a depiction Property in the composer, rendering the crop, the row that shows the image, a display string for a likeness on graph cards and composer rows | S10-34 |
| Promote's exhibit images | S10-36 |
| Reconciling depictions; the gallery | S10-52, S10-35 |
| A colour for the `depiction` Properties pill (it uses the default `PVColor.textSecondary`) | S10-34's brief, S10-D15 |
| `FakeStore` producing depiction values | S10-34, when a UI test first writes one |
| Ingest copy written for a crop rather than a file import (`ingest.symlink`, `ingest.missing`, `ingest.too_large` keep their import wording) | Not planned; a crop path is written by the app, so these reach the researcher only on an app bug |
| The File id of a depiction image in any response | Not planned; no reader needs it (S10-35 reads the reconciler cache) |

## 4. Current state

Verified at `ae26c629`, except the rows marked *(S10-32)*, which are S10-32's plan.

| Area | Where | Today |
| --- | --- | --- |
| Proto | `api/proto/engine.proto` `Observation` (fields 1–17), `ObservationDraft` (fields 1–12), `NameValueInput`, `UpdateObservationRequest` (`Observation observation = 3`), `CreateCitationWithObservationsResponse`, `AddObservationsToCitationResponse`, `CreateCitedBridgeResponse`, `Artifact.thumbnail_rel_path` | An update sends an `Observation`; inserts send `ObservationDraft`. Object paths are `objects/…` strings relative to the project folder (`thumbnail_rel_path`, `SourceFileRef.rel_path`). |
| Generation | `scripts/generate-proto.sh` | Writes `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`; needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`. |
| Draft → input | `api/ffi/handlers/observations.go` `observationToInput`, `observationDraftToInput`, `observationDraftToInputAllowEmptySubject`, `valueInput`, `optionalID`; `observationsToInputs` | One mapping per message; `valueInput` takes each scalar as a parameter. `observationsToInputs` has no caller (`rg observationsToInputs` finds only its definition). |
| Writers | `api/ffi/handlers/citations.go` `CreateCitationWithObservations`; `api/ffi/handlers/observations.go` `AddObservationsToCitation`, `UpdateObservation`; `api/ffi/handlers/connect.go` `CreateCitedBridge` | Each maps its inputs before `withProjectCatalog`, then calls `writes.Run` inside it. No staging. |
| Responses | `api/ffi/handlers/observations.go` `observationProto` (ids and scalars only), `listedObservationProto` (adds Property summary, date, name); `citations.go` and `AddObservationsToCitation` return `observationProto`; `UpdateObservation` returns `listedObservationProto`; `connect.go` `CreateCitedBridge` reads `observations.ListByCitation` and keeps the rows it wrote | Create and append return no Property summary and no structured value. |
| Reads | `api/ffi/handlers/source_read.go` `listedProtosFromSource`, `listedFromSource`; `api/ffi/handlers/citations.go` `GetCitation`; `observations.go` `ListObservationsBySource` | Both reads project the Source graph cache through `listedFromSource`. |
| Object paths | `api/ffi/handlers/artifacts.go` `artifactProto`; `core/database/files/files.go` `StorageRelPath`, `LookupMany` | The handler derives a File's `objects/…` path from its checksum and media type. |
| Source graph | `core/database/graphcache/source.go` `SourceObservation`; `core/database/sourcefill/load.go` `loadSourceObservations` (batch `datevalues.LookupManyTx`, `namevalues.LookupManyTx`); `core/database/graphcache/verify.go` `Verify` | The cache is in memory; the source check compares the whole graph. |
| Batch lookup | `core/database/datevalues/datevalues.go` `LookupManyTx` | The shape a value package's batch read takes. |
| Staging | `core/ingest/stage.go` `StageImage` (`ErrUnsupportedImage` on a failed decode), `InsertTx`, `Discard`; `core/ingest/file.go` `openChecked` (empty or relative path → `ErrInvalid`, symlink → `ErrSymlink`) | Landed in S10-53. |
| S10-32 *(S10-32)* | `ingest.StageDepictions(c, inputs) (discard func(), err error)` in core/ingest/depictions.go; `observations.DepictionInput{RegionJSON, ImagePath, Image}`; `observations.Input.Depiction`, `.ValueDepictionID`; `observations.Observation.ValueDepictionID`; `observations.Listed.Depiction *depictionvalues.Value`; `depictionvalues.Value{ID, RegionJSON, ImageFileID}`, `LookupTx`, `ErrInvalid`; `apperr.CodeDepictionValuesInvalid`; `graphcache.SourceObservation.DepictionID`; `listedFromSource` copies it into `ValueDepictionID` | S10-32 D-7, D-8, D-10. An update with `Depiction` nil and `ValueDepictionID` equal to the existing value keeps it (D-9). |
| Swift value types | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType`; `macos/App/Features/Properties/PropertiesModel.swift` `SubjectPropertyValueType.label`, `researcherCreatable`; `macos/App/Platform/L10n.swift` `L10n.Properties.valueTypeTerm`; `macos/App/Features/Catalog/SubjectTypeChrome.swift` `valueTypeForeground` (default `PVColor.textSecondary`) | No `depiction` case; the label falls back to *Text*. No Swift `switch` over the enum is exhaustive, so a new case breaks no build. |
| Swift errors | `macos/App/Platform/L10n.swift` `L10n.Errors.message(code:params:)`, `dateValuesInvalid`, `nameValuesInvalid`, `ingestSymlinkHelp` | No case for `ingest.unsupported_image` or `depictionvalues.invalid`; both fall to `unknown`. |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogObservation`, `CatalogObservationDraft`, `CatalogNameValuePart`; `macos/App/Platform/GoStore.swift` `mapObservation`, `mapCatalogObservation`, `mapObservationDraft` (all `private static`), `mapWorkspaceLocationFromProto` (internal `static`, used by tests); `macos/App/Platform/FakeStore.swift` | The update request is built from the whole `CatalogObservation`. |
| Parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "value types" | S10-32 left it unchanged; this PR adds the entry with the Swift case (S10-32 D-3). |
| Tests | `api/ffi/handlers/citations_test.go` `citationFixture`, `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, `TestListObservationsBySource`, `TestGetCitation`, `TestUpdateObservation`; `api/ffi/handlers/connect_test.go` `TestCreateCitedBridge`; `api/ffi/handlers/thumbnails_test.go` `writePNG`; `api/ffi/handlers/subjects_test.go` `subjectFixture` (returns the person type id); `api/ffi/handlers/harness_test.go` `runRPC`; `macos/ProvenenciaTests/PropertiesModelTests.swift`, `L10nResolutionTests.swift` | `citationFixture`'s Artifact has no File, so a depiction cannot be saved on it. |

**Precedent.** The client-path shape is `IngestArtifactFile` in `api/ffi/handlers/artifacts.go`: an absolute path in the request, the `core/ingest` call inside `withProjectCatalog` and before `writes.Run`, and the `objects/…` path back through `files.StorageRelPath`, as `artifactProto` does. The listed-response shape is `CreateCitedBridge` in `api/ffi/handlers/connect.go` (`observations.ListByCitation`, then keep the written ids). The batch value read is `datevalues.LookupManyTx` and its use in `sourcefill.loadSourceObservations`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `engine.proto`, after `NameValueInput`: `message DepictionValueInput { string region_json = 1; string image_path = 2; }` *(new)* and `message DepictionValue { string region_json = 1; string image_rel_path = 2; }` *(new)*. `ObservationDraft` gains `DepictionValueInput depiction = 13;`. `Observation` gains `string value_depiction_id = 18; DepictionValue depiction = 19; DepictionValueInput depiction_input = 20;`. Comments in §6. | An input names a file the app wrote; a saved value names an object Go stored. Two messages keep a request from carrying a path Go must ignore and a response from carrying an empty input. `depiction_input` on `Observation` is the update's input, since an update sends an `Observation` (§4). | One message for both — every reader has to know which half is set. A `bytes` field — working rule "bytes never cross the FFI". An `image_file_id` field — Go never takes a File id from the client, and no reader needs one (Out). |
| D-2 | `func depictionInput(d *engine.DepictionValueInput) *observations.DepictionInput` *(new)* in `api/ffi/handlers/observations.go`: nil → nil; otherwise `&observations.DepictionInput{RegionJSON: strings.TrimSpace(d.GetRegionJson()), ImagePath: strings.TrimSpace(d.GetImagePath())}`. `observationDraftToInputAllowEmptySubject` sets `in.Depiction = depictionInput(d.GetDepiction())`. `observationToInput` sets `in.Depiction = depictionInput(o.GetDepictionInput())` and `in.ValueDepictionID` from `optionalID(o.GetValueDepictionId())`. `o.GetDepiction()` is never read. `valueInput` is not changed. A draft has no `value_depiction_id`. | The handler only maps; S10-32's write refuses every wrong mix (D-8, D-9 there): an empty region is `locator.invalid`, an empty or relative path is `ingest.invalid` from staging. A request's saved `depiction` is a response echo, so reading it would let the client point at any object. | Adding two parameters to `valueInput` — it already takes nine, and a depiction is not a scalar. Validating in the handler — a second copy of S10-32's rules. |
| D-3 | Each of the four writers stages inside `withProjectCatalog`, before `writes.Run`, with S10-32's call and nothing else: `discard, err := ingest.StageDepictions(c, inputs); if err != nil { return err }; defer discard()`. `UpdateObservation` passes `[]observations.Input{input}`. `CreateCitedBridge` passes `inputs` before they go into `connect.CreateInput`. | S10-32 D-7 makes this the one core entry point and requires it outside `writes.Run` and inside `catalogsession.Do` (S10-53); `Input.Depiction` is a pointer, so the staged image travels inside the same inputs and `connect`'s copy of them. A bridge draft is staged too, so no writer path can reach the write with an unstaged path (map row: every handler). | A shared wrapper around `writes.Run` — each handler has its own `Op` and result type (S10-32 D-7 rejected it). Skipping `connect` because no bridge type binds `likeness` — the binding could change; staging is the guarantee. |
| D-4 | `func depictionPaths(c *database.Catalog, rows []observations.Listed) (map[string]string, error)` *(new)* in `api/ffi/handlers/observations.go`: collects `l.Depiction.ImageFileID` for rows with a `Depiction`, calls `files.LookupMany(c, ids)` once, and maps each id to `files.StorageRelPath(f.ChecksumSHA256, f.MediaType)`. A depiction whose File is not in the result returns `sql.ErrNoRows`. `listedObservationProto(l observations.Listed, paths map[string]string) *engine.Observation` gains the `paths` parameter and sets `o.Depiction = &engine.DepictionValue{RegionJson: l.Depiction.RegionJSON, ImageRelPath: paths[string(l.Depiction.ImageFileID)]}` when `l.Depiction` is set. `observationProto` sets `ValueDepictionId: uuidString(o.ValueDepictionID)`. | One `files` query per response, not one per row; the path is derived at the edge like `artifactProto`'s, so the cache and `Listed` hold only catalog values. | Storing the path in `Listed` or the graph cache — a derived presentation value in two stores. A `files.Lookup` per row — N queries on a Source read. |
| D-5 | `func listedWritten(c *database.Catalog, citationID []byte, written []observations.Observation) ([]*engine.Observation, error)` *(new)* in `api/ffi/handlers/observations.go`: `observations.ListByCitation(c, citationID)`, keep the rows whose id is in `written`, in `ListByCitation` order, then `depictionPaths` and `listedObservationProto`. `CreateCitedBridge` replaces its inline filter with it; `CreateCitationWithObservations` and `AddObservationsToCitation` return it in place of `observationProto` rows. | Map Testable: "a save … returns its path". `observations.Observation` holds only the value id, so the path needs the listed read the connect handler already does; one helper gives every write response one shape. Swift's `mapObservation` already reads the richer fields. | A depiction-only lookup in the two handlers — a third response shape. |
| D-6 | `depictionvalues.LookupManyTx(q interface { Query(string, ...any) (*sql.Rows, error) }, ids [][]byte) (map[string]Value, error)` *(new)*, keyed by `string(id)`, missing ids omitted, the `datevalues.LookupManyTx` shape. `graphcache.SourceObservation` gains `DepictionRegion string` and `DepictionImageFileID []byte` *(new)* after S10-32's `DepictionID`. `loadSourceObservations` collects `DepictionID`s and fills both fields from one `LookupManyTx`, beside the date and name batches. `listedFromSource` sets `l.Depiction = &depictionvalues.Value{ID: o.DepictionID, RegionJSON: o.DepictionRegion, ImageFileID: o.DepictionImageFileID}` when `len(o.DepictionID) > 0`. | S10-32's Out rows give the batch read and the source graph's depiction to "the first PR with a batch caller (S10-54)". `Verify` compares the whole struct, so a stale region is caught. The cache is in memory: no version to bump, and S10-32's `depiction_value` effects entry reloads the Source on an edit. | A `valuecodec` blob like `Date` and `Name` — those are multi-field values the reconciler also stores; a region string and an id need no codec. A `*depictionvalues.Value` field — `graphcache` holds plain fields and imports no value package. |
| D-7 | Swift, in `GenealogyStore.swift` beside `CatalogNameValuePart`: `struct CatalogDepictionValueInput: Sendable, Equatable { var regionJSON: String; var imagePath: String }` and `struct CatalogDepictionValue: Sendable, Equatable { var regionJSON: String; var imageRelPath: String }` *(new)*. `CatalogObservation` gains, after `valueTermKey`, `var valueDepictionID: String = ""`, `var depiction: CatalogDepictionValue? = nil`, `var depictionInput: CatalogDepictionValueInput? = nil`. `CatalogObservationDraft` gains, after `notes`, `var depiction: CatalogDepictionValueInput? = nil`. `GoStore.mapObservation` reads `valueDepictionID`, and `depiction` from `o.depiction` when `o.hasDepiction`, else nil; `mapCatalogObservation` writes `valueDepictionID` and, when set, `depictionInput` into `depiction_input`, and never writes `depiction`; `mapObservationDraft` writes `depiction`. The three functions drop `private` (stay `static`), as `mapWorkspaceLocationFromProto` is, so a test can call them. `FakeStore` is not changed. | The Swift types mirror D-1 one to one. Defaults keep every existing initializer call compiling. The update path sends the kept id, and a new region only with a new crop (S10-32 D-9). | Reusing one Swift struct for both — the same ambiguity D-1 avoids. A `FakeStore` depiction path — no UI writes one until S10-34 (Out). |
| D-8 | `PropertyValueType` gains `case depiction` after `term`. `SubjectPropertyValueType.label` gains `case "depiction": return L10n.Properties.valueTypeDepiction` *(new key `properties.valueType.depiction`, "Depiction", comment "Property value type label: depiction (registry-only)")*; its doc comment reads "Schema also has `term` and `depiction` (registry-only) — never offered here." `researcherCreatable` is unchanged. `CatalogProperty.valueType`'s comment lists `depiction`. `TestSwiftVocabularyMatchesGo` "value types" adds `properties.ValueTypeDepiction`. `valueTypeForeground` is unchanged. | S10-32 D-3 assigns the Swift case and the parity entry to this PR, together. The Properties pill keeps the raw key and gains the right VoiceOver label. | A pill colour — chrome; S10-34's brief (Out). |
| D-9 | Error copy in `L10n.Errors`: `ingestUnsupportedImage` *(new)*, key `error.ingest.unsupported_image`, "The picture of that region couldn’t be read.", comment "FFI error ingest.unsupported_image (a depiction crop Go could not decode)"; `depictionValuesInvalid` *(new)*, key `error.depictionvalues.invalid`, "Invalid depiction value.", comment "FFI error depictionvalues.invalid". `message(code:params:)` gains `case "ingest.unsupported_image"` beside the other `ingest.` cases and `case "depictionvalues.invalid"` after `namevalues.invalid`. `ingestCallout` is not changed. | `add-ffi-handler`: every user-facing code has Swift copy. `ingest.unsupported_image` comes only from `StageImage`, which only a depiction save calls, so its copy names the picture, not a file the researcher chose. `depictionvalues.invalid` follows `dateValuesInvalid`. | An Artifact-import callout — the crop is not an import. |
| D-10 | Delete `observationsToInputs` from `api/ffi/handlers/observations.go`. | It has no caller (§4), and D-2 edits the mapping beside it. | Leaving it — the next value field would be added to a dead copy. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `DepictionValueInput`, `DepictionValue` *(new)* | proto message | `NameValueInput`, `DateValueInput` | Those are values the client composes; a depiction input is a path and a region, and its saved form is a different pair (D-1). |
| `ObservationDraft.depiction`, `Observation.value_depiction_id`, `.depiction`, `.depiction_input` *(new)* | proto field | `date`, `value_date_id` | Extends the same messages. |
| `depictionInput` *(new)* | func | `valueInput` | `valueInput` maps scalars; this maps one message to a pointer (D-2). |
| `depictionPaths` *(new)* | func | `artifactProto`'s `files.Lookup` + `StorageRelPath` | Same derivation, batched (D-4). |
| `listedWritten` *(new)* | func | the filter inline in `CreateCitedBridge` | Moves that filter out so three handlers share it (D-5). |
| `depictionvalues.LookupManyTx` *(new)* | func | `datevalues.LookupManyTx` | Same shape; S10-32 deferred it here. |
| `SourceObservation.DepictionRegion`, `.DepictionImageFileID` *(new)* | field | `Date`, `Name` | Plain fields, no codec (D-6). |
| `CatalogDepictionValueInput`, `CatalogDepictionValue` *(new)* | Swift struct | `CatalogNameValuePart`, `CatalogDateValueInput` | Mirror D-1. |
| `PropertyValueType.depiction`, `L10n.Properties.valueTypeDepiction` *(new)* | enum case / key | `.term`, `valueTypeTerm` | Same list. |
| `L10n.Errors.ingestUnsupportedImage`, `depictionValuesInvalid` *(new)* | key | `ingestSymlinkHelp`, `nameValuesInvalid` | Same switch. |
| Test helpers `likenessFixture`, `writeCrop`, `objectFiles` *(new)* in `api/ffi/handlers/citations_test.go` | helper | `citationFixture`, `writePNG` | `citationFixture`'s Artifact has no File and its subject is a place; `writePNG` writes one fixed 8 × 8 image. |
| `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` *(new)* | test | `WorkspaceNavigationTests` (calls a `GoStore` map) | No test covers Observation mapping today. |

**Builds on.** S10-32's `ingest.StageDepictions`, `Input.Depiction`, `Listed.Depiction`, and its effects entry; S10-53's staging errors; the `IngestArtifactFile` path-in, rel-path-out shape; `add-ffi-handler`.
**Deletes.** `observationsToInputs`; the inline written-id filter in `CreateCitedBridge`.
**Next adds.** S10-34 sets `CatalogObservationDraft.depiction` from the composer and shows `CatalogObservation.depiction.imageRelPath`. S10-36 reads the same field on Promote rows. A future depiction key (place, event) needs no FFI change.

### Schema

None — no catalog change. S10-32 adds `depiction_values`.

### Go core

| Package | Change |
| --- | --- |
| `core/database/depictionvalues` (S10-32 *(new)*) | `LookupManyTx` (D-6). |
| `core/database/graphcache` | `SourceObservation` fields (D-6). |
| `core/database/sourcefill` | `loadSourceObservations` fills them (D-6). |

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
// path, relative to the project folder, of the PNG Go stored.
message DepictionValue {
  string region_json = 1;
  string image_rel_path = 2;
}

// in ObservationDraft
  DepictionValueInput depiction = 13;

// in Observation
  string value_depiction_id = 18;
  // Set on responses; never read from a request.
  DepictionValue depiction = 19;
  // Read on UpdateObservation only: a new region and its crop. Unset keeps
  // the value named by value_depiction_id.
  DepictionValueInput depiction_input = 20;
```

Handlers: `api/ffi/handlers/observations.go` (D-2, D-3, D-4, D-5, D-10), `api/ffi/handlers/citations.go` (D-3, D-5), `api/ffi/handlers/connect.go` (D-3, D-5), `api/ffi/handlers/source_read.go` (D-4, D-6). Regenerate with `./scripts/generate-proto.sh`.

### macOS

No view changes.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value type | `PropertyValueType.depiction` | `macos/App/Platform/InterpretationValueKinds.swift` | D-8 |
| Label | `SubjectPropertyValueType.label` | `macos/App/Features/Properties/PropertiesModel.swift` | D-8; the pill's VoiceOver label |
| Store types | `CatalogDepictionValueInput`, `CatalogDepictionValue`, fields | `macos/App/Platform/GenealogyStore.swift` | D-7 |
| Mapping | `mapObservation`, `mapCatalogObservation`, `mapObservationDraft` | `macos/App/Platform/GoStore.swift` | D-7 |
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
Do: D-1, the §6 FFI block exactly. Run `./scripts/generate-proto.sh` and commit both generated files.
Check: `CGO_ENABLED=1 go build ./...` succeeds; `git diff --stat` shows only these three files.

**STEP-2 — The source graph carries the depiction**
Files: `core/database/depictionvalues/depictionvalues.go`, `core/database/depictionvalues/depictionvalues_test.go` (both S10-32 *(new)*), `core/database/graphcache/source.go`, `core/database/sourcefill/load.go`. Skill: `add-catalog-query`.
Do: D-6, except `listedFromSource` (STEP-4). Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-3 — Map and stage depiction inputs**
Files: `api/ffi/handlers/observations.go`, `api/ffi/handlers/citations.go`, `api/ffi/handlers/connect.go`, `api/ffi/handlers/citations_test.go`, `api/ffi/handlers/connect_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-2, D-3, D-10. In `citations_test.go` add the helpers:
- `likenessFixture(t) (dir, userID, sourceID, artifactID, personID, likenessID string)`: `subjectFixture`; `CreateArtifact`; `writePNG` to `filepath.Join(t.TempDir(), "photo.png")` and `IngestArtifactFile` with that path; `CreateSubject` with the person type id `subjectFixture` returns, label "Ada"; `ListProperties` and the id whose key is `likeness`.
- `writeCrop(t, w, h int, shade uint8) string`: a `w` × `h` PNG filled with opaque gray `shade`, written to `t.TempDir()`; returns its absolute path. It never equals `writePNG`'s transparent 8 × 8 image.
- `objectFiles(t, dir string) []string`: every regular file under `dir/objects`, sorted, as paths relative to `dir`.
- constants `likenessLocatorJSON` = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}` and `likenessRegionJSON` = `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`.
Add T-3 and T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-4 — Return the saved value and its path**
Files: `api/ffi/handlers/observations.go`, `api/ffi/handlers/citations.go`, `api/ffi/handlers/connect.go`, `api/ffi/handlers/source_read.go`, `api/ffi/handlers/citations_test.go`. Skill: `add-ffi-handler`.
Do: D-4, D-5, and D-6's `listedFromSource` line. `listedProtosFromSource` collects its `Listed` rows into a slice, calls `depictionPaths` once, then builds the protos. `UpdateObservation` calls `depictionPaths(c, []observations.Listed{row})`. Add T-2, T-4, T-5, T-6. The existing rows of `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, and `TestCreateCitedBridge` pass unedited.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-5 — Swift value type, label, and error copy**
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Platform/GenealogyStore.swift` (the `CatalogProperty.valueType` comment only), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/PropertiesModelTests.swift`, `macos/ProvenenciaTests/L10nResolutionTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-8, D-9. Add T-8, T-9, T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `python3 scripts/check-localizable-xcstrings.py` passes; on macOS, the `xcodebuild test` command in `scripts/check-pr.sh` passes.

**STEP-6 — Swift store types and mapping**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-7. Add T-11.
Check: on macOS, the `xcodebuild test` command in `scripts/check-pr.sh` passes. Without a Mac, report this check as not run.

**STEP-7 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `interpretation-layer-data-model.md`: after S10-32's `value_type = 'depiction'` paragraph add "Over the FFI a save sends `DepictionValueInput` (the region and the absolute path of the app's rendering); Go stages it before the write. Responses and reads return `DepictionValue`: the region and the image's `objects/…` path. The client never sends or receives the image's File id."
- `deployment-plan.md`: the S10-54 entry's **Plan:** line (added when this plan was written) ends "— Landed #N"; the SLICE 8 line for S10-54 gains "— done, #N". `completed.md`: a `## S10-54 — Depiction values over the FFI` section above the current first section, in the S10-50 shape.
- This plan: Status and a **Landed** note per the DoD.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-54-depiction-ffi.md` passes.

STEP-3 and STEP-4 belong together: after STEP-3 a likeness saves but the create response has no path. Push only after STEP-4 passes. Open the draft PR before STEP-7, so its number is known.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/depictionvalues/depictionvalues_test.go` `TestDepictionValues` row "lookup many" *(new)* | two inserted values and one unknown 16-byte id → a map of the two, keyed by `string(id)`, with region and image id; no ids → empty map | D-6 |
| T-2 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationWithObservations` row "saves a likeness and returns its image path" *(new)* | `likenessFixture`; draft with `likenessRegionJSON` and `writeCrop(t, 300, 200, 128)` → AC-1.1 fields; `objectFiles` gains exactly the returned path; its bytes differ from the crop file's | AC-1.1, D-1, D-3, D-4, D-5 |
| T-3 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationRefusesSymlinkedCrop` *(new)*, a plain test (`runRPC` skips `after` on an error) | `likenessFixture`; record `objectFiles`; AC-1.1 request with `image_path` an `os.Symlink` to `writeCrop(t, 300, 200, 128)` → `errors.Is(err, ingest.ErrSymlink)`; `objectFiles` unchanged; `ListCitationsByArtifact` for the Artifact returns no Citation; end with `catalogsession.CloseAll()` in `t.Cleanup` | AC-1.2, D-3 |
| T-4 | Go | `TestAddObservationsToCitation` row "adds a likeness" *(new)* | `likenessFixture`; `CreateCitationWithObservations` with `likenessLocatorJSON` and no drafts; `AddObservationsToCitation` with the AC-1.1 draft → the AC-1.1 fields | AC-1.3, D-5 |
| T-5 | Go | `TestGetCitation` row "returns a likeness" and `TestListObservationsBySource` row "returns a likeness" *(new)* | after the T-2 save → AC-2.1 fields; for `ListObservationsBySource`, then `catalogsession.Do(dir, func(c *database.Catalog) error { return c.Graph().Verify() })` → nil | AC-2.1, AC-2.2, D-6 |
| T-6 | Go | `TestUpdateObservation` rows "replaces a likeness image", "keeps a likeness on a polarity change", "ignores a sent image path" *(new)* | after the T-2 save: `depiction_input` with region points `(0.5,0.5),(0.9,0.5),(0.9,0.9)` on page 1 and `writeCrop(t, 200, 200, 64)` → same value id, new region, new path; old path absent from `objectFiles`. Polarity `negative`, no input → value id, region, path unchanged. No input and `depiction.image_rel_path` set to the Artifact's own `objects/…` path → path unchanged | AC-3.1, AC-3.2, AC-3.3, D-2, D-4 |
| T-7 | Go | `api/ffi/handlers/connect_test.go` `TestCreateCitedBridge` row "stages a depiction draft" *(new)* | the test's existing bridge request with one draft carrying `depiction` = { `likenessRegionJSON`, a symlink path } → `wantErrIs: ingest.ErrSymlink` | AC-1.4, D-3 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "value types" | Go list adds `properties.ValueTypeDepiction` → matches Swift | AC-4.1, D-8 |
| T-9 | Swift | `macos/ProvenenciaTests/PropertiesModelTests.swift` `depictionIsRegistryOnlyAndLabelled` *(new)* | `L10n.string(SubjectPropertyValueType.label("depiction")) == "Depiction"`; `!researcherCreatable.contains("depiction")`; `PropertyValueType(rawValue: "depiction") == .depiction` | AC-4.1, D-8 |
| T-10 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` `depictionSaveErrorsHaveCopy` *(new)* | `message(code: "ingest.unsupported_image")` == "The picture of that region couldn’t be read."; `message(code: "depictionvalues.invalid")` == "Invalid depiction value."; neither equals `message(code: "internal.unknown")` | AC-4.2, D-9 |
| T-11 | Swift | `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` *(new)* `mapsSavedDepiction`, `updateSendsInputNotSavedValue`, `draftSendsDepiction` | proto with `valueDepictionID` "v" and `depiction` (region "r", path "objects/a.png") → `CatalogObservation` with those; `CatalogObservation` with `depiction` and `depictionInput` (region "r2", path "/tmp/c.png") → proto `depictionInput` set, `hasDepiction` false, `valueDepictionID` kept; draft with `depiction` → proto `depiction` set | AC-4.3, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the existing create, append, and connect rows (D-5 changes their response rows), `TestGetCitationSecondReadUsesTheLoadedSource` (STEP-4 batches `listedProtosFromSource`), and the source graph verify (STEP-2).

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
- [ ] Domain doc updated: interpretation-layer-data-model.md (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestCreateCitationWithObservations|TestAddObservationsToCitation|TestGetCitation|TestListObservationsBySource|TestUpdateObservation|TestCreateCitedBridge|TestSwiftVocabularyMatchesGo' -v ./api/...` and paste the likeness rows' output in the PR. (AC-1.1 to AC-3.3)
2. On a Mac, build the branch, create a new project, and open **Properties → Person**. The Likeness row's pill reads `depiction`; with VoiceOver on, it reads *Depiction*. The composer's Property menu still has no Likeness. (AC-4.1) An implementer without a Mac reports this check as not run.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-32 lands with different names or signatures | The note under the header; §4's *(S10-32)* row lists every name used. Refresh this plan before implementing if S10-32's Landed note changes any of them. |
| The richer create and append responses change a Swift caller | `GoStore` maps both through `mapObservation`, which already reads every listed field. The manual check opens the composer. |
| A Source with many likenesses reads slowly | One `files.LookupMany` per response (D-4) and one `LookupManyTx` per Source load (D-6). |
| The client sends a saved `depiction` pointing at another object | Never read (D-2); T-6 "ignores a sent image path". |

**Stop and report if:**

- S10-32's Landed note renames or re-signs any name in §4's *(S10-32)* row, or `ingest.StageDepictions` is not safe to call with inputs that have no depiction.
- `observations.Listed` after S10-32 has no `Depiction` field, or `graphcache.SourceObservation` has no `DepictionID`.
- An existing row of `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, or `TestCreateCitedBridge` fails after STEP-4 for a reason other than a field it now receives.
- `TestUpdateObservation` "keeps a likeness on a polarity change" fails with `observations.invalid` (S10-32 D-9's keep branch is not as planned).
- `./scripts/generate-proto.sh` changes any generated message other than the five D-1 names.
- A Swift `switch` over `PropertyValueType` fails to compile after STEP-5.

## 12. Open questions

None.

## 13. PR

**Title:** S10-54 — Depiction values over the FFI
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-54-depiction-ffi.md`.
