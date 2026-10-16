# S10-32 — Depiction value and `likeness` on Person

**Status:** Blocked — on S10-51, S10-53
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-32--depiction-value)
**Brief:** — (no new chrome: an engine value type, its value table, and a seeded Property)
**Board:** —
**Depends on:** S10-51 (owned release from `catalogmodel`, `ReleaseReplaced`, `files.PurgeAfterCommit`; not landed), S10-53 (`ingest.StageImage`; not landed)
**Blocks:** S10-54, S10-52
**Branch:** `s10-32-depiction-value`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/seeded-vocabulary.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) ("Shape", "The image", "Settled"), [`docs/interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (Properties `value_type`, value population), [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.2–3.3, working rules "A depiction's value is its own row: region plus image", "The client renders the image; Go validates it and stores its own copy", "A File is released only when nothing points at it", "Deletes follow the evidence chain", "`likeness` is new catalogs only"; map review [`../deployment-plan.review.md`](../deployment-plan.review.md) K-1, K-4, K-8
**Planned:** 2026-10-16 against `main` @ `98819f86`, with [S10-51](S10-51-file-release.md) and [S10-53](S10-53-stage-image.md) as planned
**Review:** [`S10-32-depiction-value-type.review.md`](S10-32-depiction-value-type.review.md) — Fits with changes, Ready with changes (applied); 1 round on this shape

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

A photograph can be evidence whose value is the picture. This PR adds the Observation value type `depiction`. Its value is an owned `depiction_values` row, like a date or a name value: the region of the Citation's page that shows the face (a locator document) and an image File that Go stored from the app's rendering of that region. The image goes through S10-53's staging, so it gets every check a file import gets, and Go stores its own PNG, never the app's bytes. Lineage runs region → Citation → Artifact → File. Deleting the Observation releases the value and its image, and the image's object is removed after the commit when nothing else points at it. A new catalog seeds `likeness`, the Person Property of that type, with cardinality `multiple`.

**Why now.** Every likeness surface (the composer row S10-34, the reconciler module S10-52, the gallery S10-35, Promote's crops S10-36) reads a depiction value, and none exists. S10-51 made File release generic and S10-53 made client images safe to store, so this PR adds no delete or staging code of its own.

**What stays.** Every existing value type writes, reads, reconciles, and deletes as today. Existing catalogs gain no Property; a later one-off CLI updates the dogfood projects (user, 2026-10-16). Researchers cannot create a depiction Property. Nothing in the app can make a likeness yet: the FFI field is S10-54 and the composer row is S10-34. The auto-reconciler skips depiction Observations until S10-52 gives them a module.

## 2. User stories

**US-1.** As a researcher, when I cite a face on a photograph, I record a likeness whose value is the region I drew and a picture of it that Go stored safely.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new catalog; an Artifact with an image File. `StageDepictions` then `citations.CreateWithObservations` with locator `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`, transcription `""`, and one likeness Observation whose `Depiction` has region `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}` and the path of a 300 × 200 PNG | It saves. The Citation's transcription is NULL. The Observation has `value_depiction_id` set and every other value column NULL. Its `depiction_values` row holds that region and an `image_file_id` whose File is `image/png`. The change list has `observation`, `file`, and `depiction_value` create changes, in that order. |
| AC-1.2 | AC-1.1 with the crop path a symlink, or an HTML file named `.png` | Refused with `ingest.ErrSymlink` or `ingest.ErrUnsupportedImage` before the write. No row is written and no new object is on disk. |
| AC-1.3 | AC-1.1 where the write fails after staging (a second Observation in the same call has a bad Property) | Nothing is written and the staged object is gone. |
| AC-1.4 | A region whose page selector says `artifact_page: 2`; a region with no `region` selector; a region that is not a valid locator | Refused with `locator.ErrInvalid`. Nothing is written; the staged object is gone. |
| AC-1.5 | AC-1.1 on an Artifact with no File | Refused with `observations.ErrInvalid`. |
| AC-1.6 | A likeness that also sets `ValueText`; a likeness with no `Depiction`; a `Depiction` on a `name` Property | Refused with `observations.ErrInvalid`. |
| AC-1.7 | A likeness Input whose `Depiction` was never passed through `StageDepictions` | Refused with `observations.ErrInvalid`; no File row is written. |
| AC-1.8 | `ListByCitation` after AC-1.1 | The row has `ValueDepictionID` and `Depiction` with the region and the image File id. |

**US-2.** As a researcher, editing or deleting a likeness never leaves an orphan picture and never removes a picture something else still uses.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Update the AC-1.1 Observation with a new region and a new 200 × 200 PNG | The same `depiction_values` row holds the new region and a new `image_file_id`. The old File row is gone and, after commit, its object. One `depiction_value` update change. |
| AC-2.2 | Update the AC-1.1 Observation with `ValueDepictionID` set to its own value and no `Depiction` (a polarity change) | The value row and its File are unchanged. |
| AC-2.7 | Update a `name` Observation on the Person to `likeness` with a staged `Depiction`; then update it back to `name` with a name value | The first creates a `depiction_values` row and File and releases the name value; the second releases the depiction value, its File, and (after commit) its object. |
| AC-2.3 | Delete the AC-1.1 Observation | Its `depiction_values` row and image File are gone, and after commit its object. The Citation then deletes. |
| AC-2.4 | Two likenesses whose crops re-encode to the same PNG; delete one | One File row, kept, with its object; deleting the second removes both. |
| AC-2.5 | Delete the Subject that holds the AC-1.1 likeness; then delete the likeness Observation and delete the Subject again | The first delete is refused with `subjects.ErrInUse` (a likeness is evidence, not a connection facet). After the Observation delete, the Subject deletes, and no value row, File, or object is left. |
| AC-2.6 | Recompute and rebuild the auto-reconciler for a handle whose member has a likeness | No error; no `likeness` row in `auto_reconciler_values` or `auto_reconciler_outcomes`; the handle's other values are as before. |

**US-3.** As a researcher, a new catalog has a Likeness Property on Person; an older catalog is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Create a catalog | `properties.Lookup(c, "likeness", "provenencia")` is `value_type = depiction`, `cardinality = multiple`, label *Likeness*, bound to `person` at sort order 4, and `LockedBinding("person", "likeness")` is true. |
| AC-3.2 | `ListProperties` on a new catalog | 22 Properties, one of them `likeness`. |
| AC-3.3 | A catalog created before this PR is opened | The migration runs; no `likeness` row is added. |
| AC-3.4 | Researcher Create Property with value type `depiction` | Refused with `properties.ErrInvalid`, as `term` is. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration: `properties.value_type` accepts `depiction`; `depiction_values`; `observations.value_depiction_id` | STEP-1, AC-3.3 |
| `catalogmodel` registration (table, two owned FKs); the `effects` entry through `effects.ownedValue`, which `date_values` and `name_values` move onto | STEP-2, AC-2.3, AC-2.5 |
| `properties.ValueTypeDepiction`; `researcherCreatable` replaces the inline `term` refusal | STEP-3, AC-3.4 |
| Seed `likeness` | STEP-3, AC-3.1, AC-3.2 |
| `locator.ValidateDepictionRegion` | STEP-4, AC-1.4 |
| `core/database/depictionvalues` package | STEP-5 |
| `Input.Depiction`, `observations.StageDepictions`, the write and update cases, the value column in every Observation read and audit map | STEP-6, STEP-7, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-2.1, AC-2.2 |
| Release on delete and on update (no new `deleteimpact` code); one `releaseReplacedValue` for date, name, and depiction values on update | STEP-7, AC-2.1, AC-2.3, AC-2.4, AC-2.5, AC-2.7 |
| The auto-reconciler skips depiction Observations (removed by S10-52) | STEP-8, AC-2.6 |
| Docs | STEP-9 |

| Out | Home |
| --- | --- |
| The FFI field for the crop path and the saved value; Swift `PropertyValueType.depiction` and its label; Swift copy for `ingest.unsupported_image` | S10-54 |
| Reconciling depictions; deleting STEP-8's filter | S10-52 |
| Rendering the crop and the composer row | S10-34 |
| Gallery and list thumbs | S10-35 |
| Place, event, and signature depiction keys | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| `likeness` in existing catalogs | A later one-off CLI (user, 2026-10-16); not in the map |
| Keeping a depiction's page in step when its Citation's locator is edited later | Not planned; the image is Go's stored copy, so the value stays a true picture of what was cited |
| Checking that the image shows the region | Not possible in Go without a PDF renderer; the app renders it (working rule "The client renders the image") |
| A display string for a depiction in delete-impact rows and graph cards | S10-34 |
| Hydrating the depiction region on the graph's source observations (`ListObservationsBySource` returns `ValueDepictionID` but no `Depiction` until then) | S10-54 (this PR adds the id only, D-10) |
| Swift copy for `depictionvalues.invalid`; the *Depiction* VoiceOver label (until then the Properties pill shows the raw key `depiction` and VoiceOver reads *Text*) | S10-54 |
| A depiction branch in `FillListedDisplayText`, `core/database/deleteimpact/projectors.go`, `promotealign`, `matching`, `conclusiondetails`, or any other reader of Observation values | Not planned; see the Risks row. S10-52 (reconciler), S10-36 (Promote), S10-34 (display) add theirs |
| `depictionvalues.LookupManyTx` | The first PR with a batch caller (S10-52 or S10-54) |
| One declared list of Observation value columns that drives the SQL lists, scans, and audit maps | Logged in the review log (M-8); not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Schema | `core/database/migrations/000024.sql`, `000043.sql` | `properties` rebuilt with `CHECK (value_type IN ('text','integer','date','name','subject','term'))`; `cardinality` added by `ALTER`. No later migration touches the `properties` DDL. The latest file is `000051.sql`. |
| Owned values | `core/database/migrations/000026.sql` (`observations` with `value_date_id`, `value_name_id`), `000030.sql` (`observations_value_date_id_uidx`, `observations_value_name_id_uidx`, unique partial) | A value row has one owner. |
| Migrate runner | `core/database/migrate.go` | Each step runs with `db.Exec` outside a transaction, so `PRAGMA foreign_keys=OFF` works. |
| Model | `core/database/catalogmodel/model.go` `Tables` (`date_values`, `name_values` are `BucketOwnedOutbound`), `FKs` (`observations.value_date_id → date_values`, `BucketOwnedOutbound`) | After S10-51, an owned-outbound FK is the whole release registration: `deleteimpact` derives the walk, counts every FK into `files`, and `ReleaseReplaced(tx, "table.column", oldID)` releases a replaced child. |
| Effects | `core/database/effects/registry.go` `valueSource`, `valueMembers`; `date_values` and `name_values` entries (identical but for `Entity` and column); the `date_values` entry (`Entity: "date_value"`, `Source`, `CitationSource`, `SourceScope` = `valueSource("value_date_id")`, `Handles`, `HandleScope` = `valueMembers("value_date_id")`); `files` entry (scope through `artifacts.file_id`) | Every table has an entry. |
| Properties | `core/database/properties/properties.go` `ValueTypeText` … `ValueTypeTerm`, `valueTypeOK`, `Create` (inline `if valueType == ValueTypeTerm` refusal); `macos/App/Features/Properties/PropertiesModel.swift` `SubjectPropertyValueType.researcherCreatable` (`text`, `integer`, `date`, `name`, `subject`) | Go refuses by deny-list, Swift offers by allow-list. |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties`, `seedBindings` | 21 Properties. Person: `name` 0 (locked), `sex_at_birth` 1, `gender` 2, `interests` 3. |
| Locator | `core/locator/validate.go` `Validate`, `TypeArtifact`, `TypePage`, `TypeRegion`, `TypeTextQuote`, `validateRegion`, `selectorRaw.ArtifactPage` | One validator for Citation locators; an `artifact` selector may only be first; a region is normalized points; nothing limits how many `page` selectors a locator has. |
| Value packages | `core/database/datevalues/datevalues.go` `Value`, `InsertTx(q, v)`, `UpdateTx(tx, id, v)`, `LookupTx`, `LookupManyTx`, `ErrInvalid` | The shape a value package takes. |
| Observation write | `core/database/observations/observations.go` `Input` (`Date *datevalues.Value`, `ValueDateID`), `InsertManyTx(tx *sql.Tx, …)`, `insertOne`, `resolvedValue`, `resolveValue(tx *sql.Tx, valueType, in, existing)`, `hasOtherScalars`, `sqlInsert`, `sqlGetRow`, `sqlListSelect`, `scanListed`, `dateCreateChange` | One `case` per value type; `default` is `ErrInvalid`. A date value is created inside `resolveValue` with its create change. |
| Observation update | `core/database/observations/update.go` `updateOne(tx *sql.Tx, …)`, `sqlUpdateObservation`, `observationUpdateChange`, `releaseDateValue` / `releaseNameValue` (hand counts `sqlCountDateRefs` / `sqlCountNameRefs`, always 0 under the `000030.sql` unique indexes, then `sqlDeleteDateValue` / `sqlDeleteNameValue`, returning the row map for a delete change); `core/database/observations/write.go` `Update(tx *database.Tx, …)` (calls `updateOne(tx.Tx, …)`), `Delete`, `scanRawObservation`, `observationRowMap` | A date update rewrites its row in place (`UpdateTx`) and emits a `date_value` update change. |
| Observation delete | `core/database/observations/write.go` `Delete`; `core/database/deleteimpact/facets.go` `eraseObservation` (called only from `releaseConnectionFacets`); `core/database/deleteimpact/connection.go` `subjectResourceInbound`; `core/database/subjects/subjects.go` `Delete` (`Impact`, `Refuse`) | `observations.Delete` snapshots and releases owned rows; after S10-51 it calls `files.PurgeAfterCommit`. A Subject, Citation, or Source with a non-facet Observation refuses to delete (`ErrInUse`), so a likeness leaves only through `observations.Delete` or `Update`. |
| Callers | `core/database/citations/citations.go` `CreateWithObservations` (→ `InsertManyTx`); `core/database/connect/connect.go` (→ `InsertManyTx`); `api/ffi/handlers/citations.go`, `observations.go` (`writes.Run` around `CreateWithObservations`, `AddToCitation`, `Update`) | Every Observation write goes through `insertOne` or `updateOne`. |
| Source graph | `core/database/sourcefill/load.go` `loadSourceObservations`; `core/database/graphcache/source.go` `SourceObservation`; `sourcefill` verify (`reflect.DeepEqual(stored, fresh)`); `api/ffi/handlers/source_read.go` `listedFromSource` | Reads every value column; the verify compares the whole struct; `listedFromSource` copies each field into `observations.Listed`. |
| Autoreconciler | `core/database/autoreconciler/autoreconciler.go` `sqlLoadCandidates` (ends `WHERE ic.status IN ('accepted', 'provisional') AND ic.entity_id IN (`), `recomputeBatch`; `core/database/autoreconciler/truth.go` | Every Observation on a member is grouped by Property and passed to `autoreconcile.Reconcile`, which returns `ErrUnknownValueType` for a type with no module. A likeness on a member would fail every recompute of that handle. |
| Staging | [S10-53](S10-53-stage-image.md) `ingest.StageImage(c, absPath) (*StagedImage, error)`, `InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)`, `Discard()` | Not landed. Must run outside `writes.Run`. |
| Release | [S10-51](S10-51-file-release.md) D-6 `deleteimpact.ReleaseReplaced(tx *sql.Tx, column string, oldID []byte) ([]files.Object, error)`, D-1 `files.PurgeAfterCommit(tx *database.Tx, objs)` | Not landed. |
| Swift app | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesValueTypePill` (`Text(verbatim: valueType)`); `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes` (no `depiction`); `macos/App/Platform/GoStore.swift` conclusion fields (`PropertyValueType(rawValue:) ?? .text`) | The pill shows the raw key; the composer does not offer `likeness`; the Person field list comes from `auto_reconciler_outcomes`, which has no likeness row while D-11 holds. |
| Tests | `core/database/observations/observations_test.go` `TestObservations`, `mustSeed` (its `seed` struct has no `likeness` field, its `seed.artifact` has no File, `validLocator` is a page-3 locator); `core/database/properties/properties_test.go` `TestProperties` row "create refuses term value type"; `core/database/subjectvocab/subjectvocab_test.go`; `api/ffi/handlers/subject_defs_test.go` `TestListProperties` (expects 21); `core/locator/validate_test.go`; `core/database/autoreconciler/autoreconciler_test.go` `newFixture`; `core/database/subjects/subjects_test.go` | The suites this PR extends. |

**Precedent.** `62f06f09` ("Add Property terms (S7-01b)") added a value type: migration `000024.sql` rebuilt `properties` with the wider CHECK, and `Create` refused it for researchers. The owned value row copies `date_values`: its package, its `observations` column, its unique partial index (`000030.sql`), its `effects` entry, and its create and update changes. The seed copies `interests` (`3d701390`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000052.sql` rebuilds `properties` with `'depiction'` added to the CHECK (the `000024.sql` shape, keeping `cardinality` and its CHECK), then creates `depiction_values` and its index, then `ALTER TABLE observations ADD COLUMN value_depiction_id BLOB REFERENCES depiction_values(id)` and `observations_value_depiction_id_uidx` (unique, partial). §6 has the SQL. No vocabulary `INSERT`. | SQLite cannot `ALTER` a CHECK. `ADD COLUMN` with a `REFERENCES` clause and a NULL default is allowed. One owner per value row, like dates and names. | Rebuilding `observations` — not needed for one nullable column. |
| D-2 | `catalogmodel`: `{Name: "depiction_values", PK: "id", Bucket: BucketOwnedOutbound}` beside `name_values`; FKs `{From: "observations", Column: "value_depiction_id", To: "depiction_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound}` and `{From: "depiction_values", Column: "image_file_id", To: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound}`. No `deleteimpact` code. | After S10-51 these entries are the whole release: deleting an Observation releases its value row, the value row releases its File when no other FK into `files` points at it, and every releasing delete purges. | A release entry in `deleteimpact` — S10-51 deleted that list. |
| D-3 | `properties.ValueTypeDepiction = "depiction"` *(new)*; `valueTypeOK` accepts it. `func researcherCreatable(valueType string) bool` *(new, unexported)* beside `valueTypeOK` is true for `text`, `integer`, `date`, `name`, and `subject`, the same name and members as Swift's `SubjectPropertyValueType.researcherCreatable`; `Create` returns `ErrInvalid` when it is false, in place of the inline `term` check. `TestSwiftVocabularyMatchesGo`'s hand-written "value types" list is not changed; S10-54 adds the Swift case and that entry together. | A new value type is registry-only until someone adds it to the list (review M-4); the rule has one name on both sides. A researcher depiction Property would have no composer row and no gallery. | A `registryOnly` deny-list — the next value type would be creatable by default. Allowing depiction — the user chose new-catalogs-only and a later CLI (2026-10-16). |
| D-4 | Seed `{Key: "likeness", Label: "Likeness", Description: "A picture of this person: a region of a cited page and the image of it. More than one is allowed.", ValueType: properties.ValueTypeDepiction, Cardinality: properties.CardinalityMultiple}` after `interests`; binding `{TypeKey: "person", PropertyKey: "likeness", SortOrder: 4, Locked: true}`. Install only. | Working rule "`likeness` is new catalogs only". Locked because S10-35's gallery and S10-39's pointer key into it, like `name`. | Seeding in `000052.sql` — `add-catalog-migration` forbids seeds in SQL, and the user declined it. |
| D-5 | `locator.ValidateDepictionRegion(regionJSON, citationLocatorJSON string) error` *(new)* in `core/locator/depiction.go` *(new)*. Both documents must pass `Validate`. Then, for the region: exactly one `region` selector, no `text_quote`, at most one `page` selector (an `artifact` selector is allowed first, as `Validate` already requires). The Citation locator must have at most one `page` selector. Page match: both have none → ok; both have one with equal `artifact_page` → ok; one has a page and the other does not, or the numbers differ → refused. Every refusal is `ErrInvalid`. The region need not lie inside a region of the Citation's own locator. | One named rule for the depiction region (map review K-8), with no undecided case (review X-2). A depiction shows one page, so a multi-page Citation cannot hold one. The page binds the image to the Citation's page; a face can sit outside the transcribed caption. | Requiring the region inside the Citation's region — a caption and a face are often apart. Taking the first of several pages — silent. |
| D-6 | Package `core/database/depictionvalues` *(new)*, shaped like `datevalues`: `Value{ID []byte; RegionJSON string; ImageFileID []byte}`, `ErrInvalid` (`CodeDepictionValuesInvalid = "depictionvalues.invalid"` *(new)* after `CodeNameValuesInvalid` in `core/apperr/apperr.go`, named like `datevalues.invalid`), `InsertTx(tx *sql.Tx, v Value) ([]byte, error)` (mints a UUIDv7), `UpdateTx(tx *sql.Tx, id []byte, v Value) error`, `LookupTx(tx *sql.Tx, id []byte) (Value, error)`. No `LookupManyTx` until a batch caller exists (Out). Insert and update refuse an empty region or an image id that is not 16 bytes; region rules are D-5's, checked by the caller. `effects`: `func ownedValue(entity, col string) Effect` *(new)* beside `valueSource` and `valueMembers` returns the five paths the `date_values` entry has today (`Source`, `CitationSource`, `SourceScope` = `valueSource(col)`, `Handles`, `HandleScope` = `valueMembers(col)`); `"date_values"`, `"name_values"`, and the new `"depiction_values"` become `ownedValue("date_value", "value_date_id")`, `ownedValue("name_value", "value_name_id")`, `ownedValue("depiction_value", "value_depiction_id")`. | Queries live in `core/database/<table>/` (`add-catalog-query`); the effects entry gives the value change its Source and handle scope (`add-audit-scope`). | Storing the region on `observations` — the value is one row with two parts, like a date. A third copy of the effects block (review M-5). |
| D-7 | `observations.DepictionInput{RegionJSON string; ImagePath string; staged *ingest.StagedImage}` *(new)*; `Input.Depiction *DepictionInput` and `Input.ValueDepictionID []byte` *(new)*, beside `Date` / `ValueDateID`. `observations.StageDepictions(c *database.Catalog, inputs []Input) (discard func(), err error)` *(new)* in `core/database/observations/depiction.go` *(new)*: for each input with a non-nil `Depiction`, `ingest.StageImage(c, d.ImagePath)` and set `d.staged`; on the first error, discard every stage so far and return `(func() {}, err)`. `discard` calls `Discard` on every stage and is safe to `defer`. It is called before `writes.Run`: `discard, err := observations.StageDepictions(c, inputs); if err != nil {…}; defer discard()`. | The one core entry point every handler calls (S10-54); every decode and disk write happens before the write, and a failed write leaves no object (S10-53 D-11). The pointer lets the staged image travel inside the existing `[]Input` without a new parameter on `CreateWithObservations`, `AddToCitation`, `Update`, or `InsertManyTx`. | A wrapper around `writes.Run` — every handler builds its own `Op` and result type. |
| D-8 | `resolveValue`'s `case properties.ValueTypeDepiction` *(new)* on insert: `hasOtherScalars(in, "depiction")` → `ErrInvalid`; `in.Depiction == nil` or `in.Depiction.staged == nil` → `ErrInvalid` (an unstaged path is never read inside the write); one query `SELECT a.file_id, c.locator_json FROM citations c JOIN artifacts a ON a.id = c.artifact_id WHERE c.id = ?` (`file_id` NULL → `ErrInvalid`); `locator.ValidateDepictionRegion(region, locatorJSON)`; `file, fileChanges, err := in.Depiction.staged.InsertTx(tx)`; `depictionvalues.InsertTx(tx, Value{RegionJSON, ImageFileID: file.ID})`; returns `resolvedValue{DepictionID, DepictionChanges: fileChanges then depictionCreateChange(…)}`; `insertOne` appends `DepictionChanges` after the `observation` change, where it appends `DateChange` today, so the order is `observation`, `file`, `depiction_value`. `resolveValue` gains a `citationID []byte` parameter (both callers have it). `hasOtherScalars` gains `if keep != "depiction" && (in.Depiction != nil || len(in.ValueDepictionID) > 0)`. A bare `ValueDepictionID` that is not the existing value is `ErrInvalid` (a value row has one owner). This insert path also runs on update when the Observation has no depiction value yet (D-9), as the date and name cases do. | Same place and shape as the date case. The region is trimmed and stored as sent after it validates. | Accepting a depiction on a fileless Artifact — there is no File for the lineage (map Testable). |
| D-9 | Update, in `resolveValue`, with the date case's branches: `Depiction` set and `existing.ValueDepictionID` set → `prev := LookupTx`; stage-insert and validate the new image as in D-8; `depictionvalues.UpdateTx(tx, existing.ValueDepictionID, next)`; if the File id changed, `deleteimpact.ReleaseReplaced(tx, "depiction_values.image_file_id", prev.ImageFileID)` and add its objects to `resolvedValue.Released`; changes are the `file` create change (if any) then a `depiction_value` update change (`depictionUpdateChange`). `Depiction` set and no existing value → D-8's insert path. `Depiction` nil and `ValueDepictionID` equal to `existing.ValueDepictionID` → keep it, no change. Anything else → `ErrInvalid`. A new region always comes with a new image. In `updateOne`, one helper `releaseReplacedValue(tx *sql.Tx, column, entity string, prevID, newID []byte, rowMap func() (map[string]any, error)) (*rowchange.Change, []files.Object, error)` *(new, unexported)*: no-op when `prevID` is empty or equals `newID`; otherwise it builds the delete change from `rowMap()` (called before the release), then `deleteimpact.ReleaseReplaced(tx, column, prevID)`. It is called after the `UPDATE observations` for `observations.value_date_id` (row map from `datevalues.LookupTx` + `dateValueMap`), `observations.value_name_id` (`namevalues.LookupTx` + `nameValueMap`), and `observations.value_depiction_id` (`depictionvalues.LookupTx` + `depictionValueMap` *(new)*); its change goes into `extra`, as today. `releaseDateValue`, `releaseNameValue`, `sqlCountDateRefs`, `sqlCountNameRefs`, `sqlDeleteDateValue`, and `sqlDeleteNameValue` are deleted. `updateOne` returns the released objects as a fourth result, and `Update` calls `files.PurgeAfterCommit(tx, released)`. | One way to remove a replaced value row (review M-1): the caller writes the audit change, S10-51's release removes the row and any File. The unique partial indexes make the old hand counts always 0. The date and name update shape is kept (review M-6). | A depiction-only release beside the date and name copies — three mechanisms in one function. |
| D-10 | `value_depiction_id` joins every list of Observation value columns: `Observation.ValueDepictionID`; `sqlInsert`, `sqlGetRow`, `sqlUpdateObservation`, `sqlListSelect` (plus `LEFT JOIN depiction_values dpv ON dpv.id = o.value_depiction_id`, selecting `dpv.region_json, dpv.image_file_id` into `Listed.Depiction *depictionvalues.Value` *(new)*); `scanRawObservation`, `scanListed`; the create change map in `insertOne`, `observationRowMap`, `observationUpdateChange`; `deleteimpact.eraseObservation`'s select and delete change; `sourcefill.loadSourceObservations`, `graphcache.SourceObservation.DepictionID` *(new)*, and `listedFromSource` (copies it into `ValueDepictionID`). The `eraseObservation` column keeps its delete change complete; a likeness never reaches it today (§4). | An unlisted column would drop the value from a read, an audit row, or the source graph verify. | — |
| D-11 | `sqlLoadCandidates` gains `AND p.value_type <> 'depiction'` before `AND ic.entity_id IN (`, with the comment "-- S10-52 replaces this with a depiction module." `CacheVersion` does not change (no stored row changes). | Without it, the first likeness on a member breaks every recompute of its handle (§4). S10-52 deletes the line when it adds the module. | Shipping S10-52 in this PR — the reconciler needs its own plan and review (map K-6). |
| D-12 | Audit: the `observation` create change carries `value_depiction_id`; `depictionCreateChange(id, v)`, `depictionUpdateChange(prev, next)`, and `depictionValueMap(id, v)` *(new)* in `observations.go` mirror the date ones (fields `id`, `region_json`, `image_file_id`); the `file` create change comes from `InsertTx`. The `file` change resolves no Source scope (its `effects` path goes through `artifacts.file_id`); the `depiction_value` change carries the Source. | The image File is audited like an import (map review "Writer's own calls"); the value change is what places the write. | A new `files` effects path through `depiction_values` — no reader needs it. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000052.sql` *(new)* | migration | `000024.sql`, `000030.sql` | Extends those shapes. |
| `depiction_values` table, `observations.value_depiction_id` | schema | `date_values`, `value_date_id` | An owned value row. |
| `core/database/depictionvalues` *(new)* | package | `core/database/datevalues` | Same shape; a different row. |
| `locator.ValidateDepictionRegion` in `core/locator/depiction.go` *(new)* | func | `locator.Validate` | Calls it and adds the depiction rule. |
| `properties.ValueTypeDepiction`, `researcherCreatable` *(new)* | const / func | `ValueTypeTerm`, the inline refusal, Swift `researcherCreatable` | Replaces the inline check; same list as Swift. |
| `effects.ownedValue` *(new)* | func | the `date_values` and `name_values` entries | Third copy; all three move onto it. |
| `observations.releaseReplacedValue` *(new)* | func | `releaseDateValue`, `releaseNameValue` | Replaces both on S10-51's release. |
| `observations.DepictionInput`, `StageDepictions` in `core/database/observations/depiction.go` *(new)* | type / func | `Input.Date` | Staging must run before the write. |
| `resolveValue` depiction case, `depictionCreateChange`, `depictionUpdateChange` *(new)* | case / func | the date case, `dateCreateChange`, `dateUpdateChange` | Same switch. |
| `likeness` seed rows *(new)* | seed | `interests` | Registry entries. |
| Tests T-1 to T-10 | tests | `TestObservations`, `TestProperties`, `validate_test.go` | Each extends its suite. |

**Builds on.** S10-51's owned release, `ReleaseReplaced`, and `files.PurgeAfterCommit`; S10-53's `StageImage`; the date value pattern; `add-seeded-vocabulary`.
**Deletes.** The inline `term` check in `properties.Create`; `releaseDateValue`, `releaseNameValue`, and their four SQL constants; the duplicated effects blocks.
**Next adds.** S10-54 adds `ObservationDraft` and `Observation` proto fields, calls `StageDepictions` in the three handlers, and maps `Listed.Depiction.ImageFileID` to an object path. S10-52 adds the module keyed by `image_file_id` and deletes D-11's line. S10-34 renders the crop and writes it to a temp file.

### Schema

Migration `core/database/migrations/000052.sql` *(new)*:

```sql
-- S10-32: value_type = depiction. The value is a depiction_values row: a
-- region of the Citation's page and the image File Go stored from it.
-- likeness is seeded by subjectvocab.Install at create time, not here.
PRAGMA foreign_keys=OFF;

CREATE TABLE properties__new (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	description TEXT,
	value_type  TEXT NOT NULL,
	cardinality TEXT NOT NULL DEFAULT 'single'
		CHECK (cardinality IN ('single', 'multiple')),

	UNIQUE (key, origin),
	CHECK (value_type IN (
		'text', 'integer', 'date', 'name', 'subject', 'term', 'depiction'
	))
) STRICT;

INSERT INTO properties__new (id, key, origin, label, description, value_type, cardinality)
SELECT id, key, origin, label, description, value_type, cardinality FROM properties;

DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;

PRAGMA foreign_keys=ON;

CREATE TABLE depiction_values (
	id            BLOB PRIMARY KEY,
	region_json   TEXT NOT NULL,
	image_file_id BLOB NOT NULL REFERENCES files(id)
) STRICT;

CREATE INDEX depiction_values_image_file_id_idx ON depiction_values(image_file_id);

ALTER TABLE observations ADD COLUMN value_depiction_id BLOB REFERENCES depiction_values(id);

CREATE UNIQUE INDEX observations_value_depiction_id_uidx
	ON observations(value_depiction_id) WHERE value_depiction_id IS NOT NULL;
```

Delete class: `depiction_values` is owned-outbound (D-2). Audit: audited (D-12). The schema hash is derived from the migrations.

### Go core

| Package | Change |
| --- | --- |
| `core/database/catalogmodel` | D-2. |
| `core/database/effects` | D-6 entry. |
| `core/database/properties` | D-3. |
| `core/database/subjectvocab` | D-4. |
| `core/locator` | D-5. |
| `core/apperr` | `CodeDepictionValuesInvalid` (D-6). |
| `core/database/depictionvalues` *(new)* | D-6. |
| `core/database/observations` | D-7 to D-10, D-12. |
| `core/database/deleteimpact` | `eraseObservation` column only (D-10). |
| `core/database/sourcefill`, `core/database/graphcache` | D-10. |
| `core/database/autoreconciler` | D-11. |

```go
// core/database/observations/depiction.go (sketch)
type DepictionInput struct {
	RegionJSON string
	ImagePath  string
	staged     *ingest.StagedImage
}
func StageDepictions(c *database.Catalog, inputs []Input) (discard func(), err error)
```

Errors: `observations.invalid` (no staged image, fileless Artifact, value fields mixed); `locator.invalid` (region); `depictionvalues.invalid` (empty region or bad id); the `ingest` errors from staging.

### FFI

None — S10-54.

### macOS

None — S10-54 adds the Swift value type and label.

### Strings

None.

## 7. Steps

**STEP-1 — Migration**
Files: `core/database/migrations/000052.sql` *(new)*. Skill: `add-catalog-migration`.
Do: D-1, the §6 SQL exactly.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` fails only in `catalogmodel` `TestPragmaHonesty`, reporting the unregistered table `depiction_values` and its two FKs (fixed in STEP-2); report any other failure.

**STEP-2 — Model and effects**
Files: `core/database/catalogmodel/model.go`, `core/database/effects/registry.go`. Skills: `add-catalog-model`, `add-audit-scope`.
Do: D-2 and D-6's effects change (`ownedValue`, three entries).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-3 — Value type and seed**
Files: `core/database/properties/properties.go`, `core/database/properties/properties_test.go`, `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/subjectvocab_test.go`, `api/ffi/handlers/subject_defs_test.go`. Skill: `add-seeded-vocabulary`.
Do: D-3, D-4. Add T-1 and T-2; `TestListProperties` expects 22 and finds `likeness`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/... ./core/database/subjectvocab/... ./api/ffi/handlers/...` passes.

**STEP-4 — The region rule**
Files: `core/locator/depiction.go` *(new)*, `core/locator/depiction_test.go` *(new)*.
Do: D-5. Add T-3.
Check: `CGO_ENABLED=1 go test ./core/locator/...` passes.

**STEP-5 — The value package**
Files: `core/apperr/apperr.go`, `core/database/depictionvalues/depictionvalues.go` *(new)*, `core/database/depictionvalues/depictionvalues_test.go` *(new)*. Skill: `add-catalog-query`.
Do: D-6. Add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/depictionvalues/...` passes.

**STEP-6 — Write a depiction Observation**
Files: `core/database/observations/depiction.go` *(new)*, `core/database/observations/observations.go`, `core/database/observations/write.go`, `core/database/observations/observations_test.go`, `core/database/deleteimpact/facets.go`, `core/database/sourcefill/load.go`, `core/database/graphcache/source.go`, `api/ffi/handlers/source_read.go`.
Do: D-7, D-8, D-10 (every list except the two in `update.go`), D-12. Until STEP-7 the depiction case returns `ErrInvalid` when `existing != nil`; `sqlUpdateObservation` does not list `value_depiction_id` yet, so an update leaves it as it is. Add `likeness properties.Property` to `mustSeed`'s `seed` struct, read with `properties.Lookup(c, "likeness", properties.OriginProvenencia)`. In `observations_test.go` add a helper `fileArtifact(t, c, userID, sourceID) artifacts.Artifact` that writes a 300 × 200 PNG with `image/png`, imports it with `ingest.File`, and creates an Artifact on that File, and a helper `crop(t, w, h int) string` that writes a PNG of that size to `t.TempDir()` and returns its path. Add T-5 and T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-7 — Update and delete release**
Files: `core/database/observations/observations.go`, `core/database/observations/update.go`, `core/database/observations/write.go`, `core/database/observations/observations_test.go`, `core/database/subjects/subjects_test.go`.
Do: D-9, and D-10's `sqlUpdateObservation` and `observationUpdateChange`. Add T-7, T-8, T-9. The existing date and name update rows in `TestObservations` must pass unedited.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-8 — Keep depictions out of the reconciler**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Do: D-11. Add `"likeness"` to `newFixture`'s Property key loop. Add T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/... ./core/database/graphcache/...` passes.

STEP-6 to STEP-8 belong together: after STEP-6 a likeness on a promoted member fails recompute until STEP-8. Push only after STEP-8 passes. Open the draft PR before STEP-9, so its number is known.

**STEP-9 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/seeded-vocabulary.md`, `docs/source-layer-data-model.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `.agents/skills/add-audit-scope/SKILL.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `interpretation-layer-data-model.md`: add `'depiction'` after `'term'` in the `properties` DDL sketch and the value-type sentence. Below the `term` paragraph: "`value_type = 'depiction'` means the value is a `depiction_values` row (`value_depiction_id`): a region locator on the Citation's page and `image_file_id`, an image File Go stored from the app's rendering of that region. The row is owned by its Observation and released with it. `likeness` on `person` is the seeded depiction Property." Name `term` and `depiction` as registry-only in "Who may create". In the `observations` DDL sketch add `value_depiction_id BLOB REFERENCES depiction_values(id),` after `value_term_id` and the `observations_value_depiction_id_uidx` index after the name one; change "DateValue and NameValue rows are exclusive to one Observation (`000030`)." to "DateValue, NameValue, and DepictionValue rows are exclusive to one Observation (`000030`, `000052`)."
- `source-layer-data-model.md` §6: after S10-53's sentence on app-rendered images add "A depiction value points at such a File (`depiction_values.image_file_id`); it is released when no row points at it."
- `seeded-vocabulary.md` §3.2: after `interests` add key `likeness`, value_type `depiction`, note "region plus image; cardinality multiple; locked on person; new catalogs only (S10-32)". §3.3: `person          likeness            # depiction; multiple; locked`.
- `add-seeded-vocabulary` skill: the value-type line adds **`depiction`** after **`term`**; "**`value_type = term` and `depiction` Properties are Install/registry only**".
- `add-audit-scope` skill: after the paragraph that begins "An entity whose effect is `None`", add "An owned value row (`date_value`, `name_value`, `depiction_value`) uses `effects.ownedValue`. A depiction's image `file` row is audited; thumbnails (`file_derivative`) stay `None`."
- `deployment-plan.md`: the S10-32 **Plan:** line reads "— Landed #N"; the SLICE 8 line gains "— done, #N". `completed.md`: a `## S10-32 — Depiction value` section above the current first section, in the S10-50 shape.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/properties/properties_test.go` `TestProperties` row "create refuses term value type" | becomes a loop over `ValueTypeTerm`, `ValueTypeDepiction`, and `"not_a_type"` → `ErrInvalid` for each; new row "upsert accepts depiction" → ok | AC-3.4, D-3 |
| T-2 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestInstallSeedsLikeness` *(new)*; `api/ffi/handlers/subject_defs_test.go` `TestListProperties` | `Install` → `likeness` depiction, multiple, label *Likeness*, person sort 4, locked; `ListProperties` → 22 with `likeness` | AC-3.1, AC-3.2, D-4 |
| T-3 | Go | `core/locator/depiction_test.go` `TestValidateDepictionRegion` *(new)*, table-driven | AC-1.1 region with a page-1 Citation → ok; region page 2 → `ErrInvalid`; region page, Citation none → `ErrInvalid`; Citation page, region none → `ErrInvalid`; no page on either → ok; two `page` selectors on the region, or on the Citation → `ErrInvalid`; `artifact` then `region` with no page on either → ok; no region selector → `ErrInvalid`; two region selectors → `ErrInvalid`; a `text_quote` → `ErrInvalid`; invalid JSON → `ErrInvalid` | AC-1.4, D-5 |
| T-4 | Go | `core/database/depictionvalues/depictionvalues_test.go` `TestDepictionValues` *(new)* | insert, lookup, update; empty region → `ErrInvalid`; 15-byte image id → `ErrInvalid` | D-6 |
| T-5 | Go | `core/database/observations/observations_test.go` `TestObservations` row "likeness saves its region and image" *(new)* | `fileArtifact`; `StageDepictions` with `crop(t, 300, 200)`; `CreateWithObservations` in `writes.Run` → AC-1.1 rows and changes; `ListByCitation` → AC-1.8 | AC-1.1, AC-1.8, D-7, D-8, D-10, D-12 |
| T-6 | Go | same file, row "likeness refuses" *(new)*, table-driven | symlink crop path and HTML-as-PNG → staging error, no new file under `objects/`; page-2 region → `locator.ErrInvalid`; fileless `seed.artifact` → `ErrInvalid`; `HasText` set → `ErrInvalid`; `Depiction` on a `name` Property → `ErrInvalid`; unstaged `Depiction` → `ErrInvalid`, no `files` row; a second bad input in the same call → nothing written and, after the deferred `discard`, no new object. Each row counts `files`, `depiction_values`, and objects before and after | AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, D-7, D-8 |
| T-7 | Go | same file, row "likeness update replaces the image" *(new)* | AC-2.1 → same value id, new region and File, old File row gone, old object gone after commit, one `depiction_value` update change; AC-2.2 → value and File unchanged; changing the Observation's Property to `name` with a name value → `depiction_values` row and File gone, a `depiction_value` delete change; AC-2.7 both directions; the existing date and name update rows still emit their `date_value` / `name_value` delete changes | AC-2.1, AC-2.2, AC-2.7, D-9 |
| T-8 | Go | same file, row "likeness delete releases its image" *(new)* | AC-2.3: `observations.Delete` in `writes.Run` → value row, File, and (after commit) object gone; then `citations.Delete` succeeds. AC-2.4: two likenesses staged from the same `crop` → one File; delete one → File and object stay; delete the other → gone | AC-2.3, AC-2.4, D-2 |
| T-9 | Go | `core/database/subjects/subjects_test.go` `TestDeleteSubjectWithLikeness` *(new)* | a Person with a likeness (built as in T-5); `subjects.Delete` in `writes.Run` → `subjects.ErrInUse`, value row and File present; `observations.Delete`, then `subjects.Delete` → ok; no `depiction_values` row, no File, no object | AC-2.5, D-2 |
| T-10 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestLikenessIsSkipped` *(new)* | `newFixture`; a Person with a `name` and a likeness (the test needs no image: `f.cite(nameInput)` first; then raw SQL on `f.c.DB()` inserts a `files` row with a 64-hex checksum, a `depiction_values` row, and an Observation on the same Citation and Subject with `ref` from `ref.Mint(ref.PrefixObservation)` and `value_depiction_id`; then `f.promote`, so upkeep reads the raw rows after they exist) → `name` reconciled as before, no `likeness` row in `auto_reconciler_values` or `auto_reconciler_outcomes`; `f.assertUpkeepEqualsRebuild`; `f.c.Graph().Verify()` passes | AC-2.6, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestPragmaHonesty` (STEP-1 to STEP-2), the effects registry coverage test (STEP-2), `TestListProperties` (count), `TestObservations` rows that compare full change maps (STEP-6), the source-graph verify (STEP-6).

AC-3.3 has no test: migrations are not tested directly (`add-catalog-migration`); the schema hash self-check covers the DDL, and the manual check covers an older catalog.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none in this PR)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI in this PR)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-32 status and SLICE 8 line; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: interpretation-layer-data-model.md, source-layer-data-model.md §6, seeded-vocabulary.md §3.2–3.3, add-seeded-vocabulary and add-audit-scope skills
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Build `main` at the **Planned** commit, create a project, and quit. Build this branch and open that project. It opens; **Properties → Person** has no Likeness row. Create a new project: **Properties → Person** lists Likeness with the pill `depiction` (the pill prints the raw key; VoiceOver reads *Text* until S10-54). The composer's Property menu does not list Likeness. (AC-3.1, AC-3.3) An implementer without a Mac reports this check as not run.
2. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestObservations/likeness|TestDeleteSubjectWithLikeness|TestLikenessIsSkipped|TestValidateDepictionRegion' -v ./core/...` and paste the result in the PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-51 or S10-53 lands with a different API | This plan waits on both; refresh D-2, D-7 to D-9 against their Landed notes. |
| A reader of Observation value columns is missed | D-10 lists every reader that copies or audits the whole value set. The other readers of Observation values select by Property key, match profile, or value column, and skip a likeness on purpose: `matching`, `promotealign` (`match.ComparerFor` is nil for an unknown type, so `pairExhibits` skips it), `conclusiondetails` (reads outcomes, none while D-11 holds), `core/database/identityclaims/bridgefile.go`, `sourceeventtitles`, `searchindex`, `canonicalgraph`, `conclusionheaders`, and `core/database/deleteimpact/projectors.go` (an empty display value). Do not add depiction branches to them (Out). |
| VoiceOver reads the Likeness pill as *Text* until S10-54 | Accepted; nothing in the app can write a likeness before S10-54 and S10-34. |
| The image does not show the region the researcher drew | The app renders it (S10-34). Go checks only that it is a safe image; the same trust as typed text. |
| An Artifact's File is replaced after a likeness is saved | The likeness keeps Go's stored image; the region refers to the page as cited. Accepted (Out). |
| A large crop slows the save | S10-53 caps the file at 32 MiB and the pixels at 64 MP, and downscales to 2048. |

**Stop and report if:**

- A migration after `000043.sql` changed the `properties` DDL.
- `TestPragmaHonesty` fails after STEP-2 for a reason other than an unregistered FK.
- Deleting a likeness Observation (T-8) needs any `deleteimpact` code beyond D-10's column (S10-51's generic release did not cover it).
- An existing date or name update test changes result after `releaseReplacedValue` (STEP-7).
- `ingest` imports `observations`, so D-7's import would cycle.
- Any test fails with `ErrUnknownValueType` after STEP-8.

## 12. Open questions

None.

## 13. PR

**Title:** S10-32 — Depiction value and likeness on Person
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md`.
