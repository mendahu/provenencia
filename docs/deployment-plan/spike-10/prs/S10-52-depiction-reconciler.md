# S10-52 — Depiction auto-reconciler module

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-52--depiction-auto-reconciler-module)
**Brief:** — (no new chrome: an engine module, a cache column, and a skip in the Person field read)
**Board:** —
**Depends on:** S10-32 (landed [#430](https://github.com/mendahu/provenencia/pull/430)), S10-54 (landed [#433](https://github.com/mendahu/provenencia/pull/433); this PR reuses its `depictionvalues.FromJoin`)
**Blocks:** S10-35
**Branch:** `s10-52-depiction-reconciler`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5 (pipeline), §7 (modules), §10 (cache and upkeep); [depictions and likenesses](../depictions-and-likenesses.md) ("Shape"); [S10-32 plan](S10-32-depiction-value-type.md) D-11 and its Landed note (`depictionvalues.Readable`); [S10-54 plan](S10-54-depiction-ffi.md) Out rows that name S10-52; working rule "A gallery is the reconciled likeness values"; the reconciler half of the first S10-32 plan (D-7, D-12 to D-14 in `git show 5a1a8cda:docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md`), re-keyed here from the Citation to the image File
**Planned:** 2026-10-17 against `main` @ `27ef7e2` (S10-54 landed); first planned against `233e87a` with S10-54 in progress
**Review:** [`S10-52-depiction-reconciler.review.md`](S10-52-depiction-reconciler.review.md) — Fits with changes, Ready with changes (applied); 2 rounds

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

S10-32 made a likeness an Observation whose value is a region and an image File, and kept those Observations out of the auto-reconciler (its D-11) so a likeness could not break a recompute. This PR gives `depiction` its own module. Two likenesses are the same value exactly when they hold the same image File, and Go stores one File per distinct PNG, so the same pixels drawn twice are one value. Negatives, provisional members, majority, and weak evidence then apply to faces the way they apply to names, and `likeness` (cardinality `multiple`) keeps every distinct image. The cache stores each value's image File id in a new column, the graph cache mirrors it so `Verify` covers it, and the Person field list skips depiction fields so the page does not grow a Likeness row before S10-35 draws the gallery.

**Why now.** S10-35's gallery reads the cached likeness values (working rule "A gallery is the reconciled likeness values"), and there are none until this module exists.

**What stays.** Every other value type reconciles exactly as today; only the cache version moves, so every catalog rebuilds once on open. No FFI field, no Swift change, no chrome. The cache column holds no foreign key, so it never keeps an image File alive (S10-55 counts only declared pointers).

## 2. User stories

**US-1.** As a researcher, a person's likenesses reconcile into a gallery: every distinct image stays, and the evidence rules that hold for names and dates hold for faces.

In every AC, a *photo* is an Artifact whose File is a PNG, a *likeness of shade N* is a likeness Observation on a Citation of that photo (locator page 1) whose crop is a 40 × 30 PNG filled with gray `N`, saved through `observationstest.SaveCitation` (`ingest.StageDepictions`, then `citations.CreateWithObservations`), and *the handle* is the Person handle made by promoting Subject A.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Subject A on Source *Register* has a likeness of shade 100; Subject B on Source *Census* has a likeness of shade 200. Promote A, then file B onto the handle | `auto_reconciler_values` holds two `likeness` rows for the handle. Their `value_image_file_id`s are the two Observations' image Files. Both are `kept` with support 1, and every other value column is NULL. Both Observations' outcomes are `kept`. |
| AC-1.2 | AC-1.1, but B's likeness is also shade 100 | One `likeness` row, `kept`, support 2. Both outcomes are `kept` with that row's rank. |
| AC-1.3 | AC-1.1, plus Subject C on Source *Album*, graded `high_trust`, with a **negative** likeness of shade 200, filed onto the handle | The shade-200 row is `denied`. B's outcome is `denied`, with `denied_by` C's Observation. The shade-100 row stays `kept`. |
| AC-1.4 | AC-1.1, but B is filed with a provisional Identity Claim | The shade-200 row is `provisional`, and so is B's outcome. The shade-100 row stays `kept`. |
| AC-1.5 | AC-1.1, with *Census* graded `low_trust` | The shade-100 row is `kept`; the shade-200 row is `weak`. Both rows are stored (S10-35 shows `kept` and `weak`). |
| AC-1.6 | After each of AC-1.1 to AC-1.5 | Upkeep equals a full rebuild (values, outcomes, handle search documents). The loaded graph node's `likeness` values carry the same image File ids. `Graph().Verify()` passes. |
| AC-1.7 | AC-1.1, then raw SQL sets the shade-200 image File's `media_type` to `image/jpeg`, then `autoreconciler.Rebuild` | One `likeness` row (shade 100). B's outcome is `no_evidence`: a stored likeness whose image `depictionvalues.Readable` refuses is no evidence, as a neutral term is. |
| AC-1.8 | AC-1.1, then A's likeness is updated through `observations.Update` with a new region and a crop of shade 50 | The handle has rows for shade 50 and shade 200 and none for shade 100. Upkeep equals a full rebuild. |
| AC-1.9 | AC-1.1; drop the graph, set one `likeness` row's `value_image_file_id` to another 16-byte id by raw SQL, load the handle's node | `Graph().Verify()` returns an error. |

**US-2.** As a researcher, the Person page does not show a Likeness field before the gallery exists.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A promoted Person with a `name` Observation *James Robins* and a likeness saved through `observationstest.SaveCitation` | `conclusiondetails.ForEntity` lists one field, `name`. No field has key `likeness`. |

**US-3.** As the engine, a catalog built by the previous reconciler rebuilds once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A catalog whose stored cache version is 14 is opened | `EnsureCatalog` rebuilds and stores `CacheVersion` 15. |

**US-4.** As the engine, a depiction is the same value as another exactly when both hold the same image.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `Reconcile("depiction", …, multiple)` with two candidates holding image `a` and one holding `b` | Two values: `a` with both members, `b` alone. |
| AC-4.2 | `Reconcile("depiction", …)` with a candidate whose `Value` is empty (`Depiction` nil); and with a candidate whose `Depiction` is set with an empty `ImageFileID` | `ErrValueMismatch`, as a date candidate with no `Date` is; and no values with that candidate's outcome `no_evidence`. |
| AC-4.3 | `Compatible("depiction", a, a)`, `(a, b)`, and two values whose `Depiction` has no image | `true`, `false`, `false`. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `autoreconcile.Value.Depiction`; `moduleFor` returns a key module on the image File through a shared `idKey`; `carries` exported as `Carries` with a depiction case; `knownValueType` deleted, its callers asking `moduleFor`; package doc | STEP-1, D-1, D-2, D-3, AC-4.1, AC-4.2, AC-4.3 |
| Migration `000053.sql`: `auto_reconciler_values.value_image_file_id` | STEP-2, D-4 |
| Loader sets `Value.Depiction` through `depictionvalues.FromJoin`; `hasValue` deleted in favour of `autoreconcile.Carries`; S10-32's `p.value_type <> 'depiction'` filter deleted; `insertRow` writes the column; `CacheVersion` 15 | STEP-2, D-4, D-5, D-6, D-7, AC-1.1 to AC-1.5, AC-1.7, AC-1.8, AC-3.1 |
| Graph cache mirrors the column (`Value`, `TruthRow`, `loadValues`, `valueEqual`, `truthRow`) | STEP-3, D-8, AC-1.6, AC-1.9 |
| `conclusiondetails` skips depiction fields | STEP-4, D-9, AC-2.1 |
| Shared likeness test fixtures `core/database/observations/observationstest` *(new)*; `observations_test.go` moves onto them | STEP-2, D-10 |
| `conclusion-reconciliation.md` §7 and §12; Spike 11's `hasValue` items; map, `completed.md`, this plan | STEP-5 |

| Out | Home |
| --- | --- |
| Drawing the gallery and the Persons list thumb; the `Reason` admit method (`kept` or `weak`) | S10-35 |
| An `objects/…` path for a cached image (`files.RelPaths`) | S10-35, through `files.RelPaths` (S10-54) |
| A preferred portrait | S10-39 |
| Promote's two images and a depiction comparer in `core/match` | Spike 11 S11-20 (moved from S10-36) |
| A Reconciliation Claim on a depiction | Not planned (map row S10-52 Out) |
| `depictionvalues.LookupManyTx` | Not planned: the loader joins `depiction_values` (D-5), the shape `sourcefill.loadSourceObservations` uses |
| A depiction operation in `TestRebuildEqualsUpkeep_SeededSequences`'s generator | Not planned: the generator cites on fileless Artifacts; the named scenario (T-4) covers the one new write sequence |
| An index on `value_image_file_id` | S10-35, if its gallery read needs one; no reader filters on it in this PR |
| A depiction case in `SortKey` | Not planned: depiction values order by rank, like terms (D-3) |
| One shape for `graphcache.Value` and `graphcache.TruthRow` (today two field-by-field mirrors) | Not planned; logged in the review log (M-5) |

## 4. Current state

Verified at `27ef7e2`.

| Area | Where | Today |
| --- | --- | --- |
| Value | `core/autoreconcile/autoreconcile.go` `Value` | One field per value type (`Text`, `TermID`, `SubjectID`, `EntityID`, `Date`, `Name`). No image field. |
| Type lists | `core/autoreconcile/autoreconcile.go` `knownValueType` (a `switch` over six types), `carries` (a `switch` over six types), `Reconcile`, `SortKey`; `core/autoreconcile/modules.go` `moduleFor` (returns nil for an unknown type), `keyModule`, `termKey`, `Compatible` (calls `knownValueType` and `carries`); the subject case's inline `keyOf` closure on `EntityID` | `Reconcile("depiction", …)` returns `ErrUnknownValueType`. The six types are listed in `knownValueType`, `carries`, and `moduleFor`, and a fourth time (`hasValue`) in the loader. `carries` says whether a value is present; the module's key says whether it is usable: subject carries `SubjectID` but keys `EntityID`, a neutral term carries but has no key, and an unusable value becomes `no_evidence` in the pipeline. |
| Loader | `core/database/autoreconciler/autoreconciler.go` `sqlLoadCandidates` (ends with the S10-32 lines `-- S10-52 replaces this with a depiction module.` / `AND p.value_type <> 'depiction'`), `load`, `hasValue`, `insertRow`, `sqlInsert`, `CacheVersion` (14, with a comment line per version) | Likeness Observations are never loaded, so no likeness row or outcome is cached. `load` drops a candidate with no value through `hasValue`, a copy of `carries`. |
| Truth | `core/database/autoreconciler/truth.go` `truthRows`, `truthRow` | Runs `load` and `Reconcile` for `graphcache.Verify`; copies each value field into `graphcache.TruthRow`. |
| Graph cache | `core/database/graphcache/graph.go` `Value`; `core/database/graphcache/load.go` `loadValues`; `core/database/graphcache/verify.go` `TruthRow`, `valueEqual` | Mirror the cache columns field by field. |
| Depiction rows | `core/database/depictionvalues/depictionvalues.go` `Value{ID, RegionJSON, ImageFileID}`, `Readable(v, citationLocatorJSON, imageMediaType)`, `FromJoin(id, region, imageFileID, citationLocator, imageMediaType) *Value` (S10-54: nil when `imageFileID` is empty, else the copied value through `Readable`; `TestFromJoin`); `core/database/observations/observations.go` `sqlListSelect` (`LEFT JOIN depiction_values dpv`, `LEFT JOIN citations dc`, `LEFT JOIN files dpf`) and `scanListed`; `core/database/sourcefill/load.go` `loadSourceObservations` (the same joins) | Both readers call `FromJoin`. S10-32's Landed note: every later reader of a stored depiction goes through `Readable`, which empties `ImageFileID` unless the File is `image/png`. |
| Files | `core/ingest/stage.go` `StageImage`, `(*StagedImage).InsertTx` (reuses the `files` row with the same checksum); `core/ingest/depictions.go` `StageDepictions` | The same PNG bytes are one File, so the File id is a content key. |
| Upkeep | `core/database/effects/registry.go` `"observations"` (`Handles: observationMembers`), `"depiction_values": ownedValue("depiction_value", "value_depiction_id")` | A likeness save, region edit, or delete already recomputes the member's handles. |
| Cache schema | `core/database/migrations/000040.sql` (`auto_reconciler_values` renamed), `000038.sql` (one `ADD COLUMN` on the cache), `000052.sql` (last migration, S10-32); `core/database/catalogmodel/model.go` `auto_reconciler_values` (`BucketSkip`, FKs only on `entity_id` and `property_id`) | No image column. |
| Detail read | `core/database/conclusiondetails/conclusiondetails.go` `loadProperties` (skips `properties.ValueTypeSubject`), `attachValues`, `loadOutcomes` (both skip a Property `loadProperties` left out) | Once likeness outcomes exist, `ForEntity` would list a `likeness` field with empty values. |
| Other cache readers | `core/database/matching/matching.go` `sqlEntityValues` (filtered by profile Properties), `core/database/searchindex/handles.go` `sqlHandleValues` (by key), `core/database/promotealign/canon.go` `handleValues` (all Properties), `core/database/promotealign/exhibits.go` `pairExhibits` (`match.ComparerFor("depiction")` is nil) | None branches on a depiction. `pairExhibits` calls `autoreconcile.Compatible` through `exhibitCompatible` on `match.Value.Auto()`, which carries no image, so `Carries` is false and no likeness pair is pinned. |
| Tests | `core/autoreconcile/modules_test.go` `TestModules`, `TestModuleFor`, `TestCompatible`; `core/database/autoreconciler/autoreconciler_test.go` `newFixture` (already loads `likeness` into `f.props`), `row`, `snapshot`, `outcome`, `TestLikenessIsSkipped` (S10-32 T-10, inserts raw rows), `TestEnsureCatalogRebuildsStaleVersion` (`[]int{0, 2, 3, 4, 5, 6, 7}`), `TestRebuildEqualsUpkeep_Scenarios`, `assertUpkeepEqualsRebuild`; `core/database/observations/observations_test.go` `writePNG`, `crop`, `fileArtifact`, `saveCitation`, `updateObservation`, the constants `likenessLocator` and `likenessRegion` (the real likeness save path, private to that test package; `runArtifactCreate` stays there); `core/database/namevalues/namevaluestest` (a shared test-fixture package, `Western`); `core/database/conclusiondetails/conclusiondetails_test.go` `newFixture`, `TestForEntityFieldsFollowTheRecords` | `TestLikenessIsSkipped` fails once D-11's filter is gone. The fixture's Artifacts have no File, so a likeness cannot be saved on them. |

**Precedent.** The first S10-32 plan's D-7, D-12, D-13 and D-14 (`git show 5a1a8cda:docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md`) are this PR keyed by the Citation: copy their shape and replace "Citation" with "image File". The key module copies `keyModule{keyOf: termKey}` in `modules.go`. The cache column copies `000038.sql` (one `ALTER TABLE … ADD COLUMN`) and the `CacheVersion` comment list. The loader join copies `sourcefill.loadSourceObservations` (S10-54): the `dpv` / `dpf` joins, the four nullable scans, and one `depictionvalues.FromJoin` call. The detail skip copies the `properties.ValueTypeSubject` skip in `loadProperties`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `autoreconcile.Value` gains `Depiction *depictionvalues.Value` *(new)* after `Name`, commented "the depiction_values row, read through depictionvalues.Readable; its ImageFileID is empty when the image is refused". `autoreconcile` imports `core/database/depictionvalues` (which imports only `apperr`, `locator`, and `uuid`: no cycle). `idKey(id []byte) (string, bool)` *(new)*, beside `termKey`, returns `string(id), len(id) > 0`. The subject case becomes `keyModule{keyOf: func(v Value) (string, bool) { return idKey(v.EntityID) }}`, keeping its comment, and `moduleFor(properties.ValueTypeDepiction)` returns `keyModule{keyOf: func(v Value) (string, bool) { if v.Depiction == nil { return "", false }; return idKey(v.Depiction.ImageFileID) }}`, commented "the same image File, so the same pixels; no readable image is no evidence". | Working rule "A gallery is the reconciled likeness values": the value's key is its image File. `StagedImage.InsertTx` reuses the File with the same checksum, so equal ids mean equal pixels. A key module gives folding none, majority on the whole value, and the best-ranked member's value displayed (§4). | Keying by checksum — the File id already is one, and the checksum is not loaded. A separate `imageKey` — the subject closure's body on another field (review M-4). Keying by region — two drawings of one face are the same picture only when their pixels are. A Citation key (the first S10-32 plan) — re-cut 2026-10-16. |
| D-2 | `carries` becomes exported `Carries(valueType string, v Value) bool` with `case properties.ValueTypeDepiction: return v.Depiction != nil`. `knownValueType` is deleted: `Reconcile` sets `m := moduleFor(valueType)` first and returns `ErrUnknownValueType` when it is nil, then passes `m` to `reconcile`; `Compatible` returns false when `moduleFor(valueType)` is nil and reuses that `m`. `Reconcile` and `Compatible` call `Carries`. In the package doc's last paragraph, the stale sentence "Subjects stay on an exact-key module until S9-28." becomes "Subjects reconcile by their accepted handle. Depictions reconcile by image File." | Map row: the depiction case is added in two places (`moduleFor`, `Carries`), not four, and the type list is `moduleFor` (the map's "`knownValueType` becomes `moduleFor(vt) != nil`", without a forwarding name, review M-7). Present-but-unusable is the key's call, not `Carries`'s, as for subjects and neutral terms, so a refused image is `no_evidence` and its outcome still explains it (review M-1). Exporting `Carries` lets the loader delete its copy (D-6). | Keeping `knownValueType` with a one-line body — a name that forwards. `Carries` on the image id — a refused image would vanish from the outcomes. |
| D-3 | `SortKey` is unchanged: a depiction has no sort key (`ok = false`), so its rows order by rank. | No reader sorts likenesses by value; S10-35 defines "earliest" in its own read (map row S10-35). | Sorting by File id — it is a UUIDv7 of the first save, not a display order. |
| D-4 | Migration `core/database/migrations/000053.sql` *(new)*: one `ALTER TABLE auto_reconciler_values ADD COLUMN value_image_file_id BLOB;` with a header comment (§6 Schema). No foreign key, no index, no `catalogmodel` change. `sqlInsert` lists `value_image_file_id` after `value_name`; `insertRow` writes `v.Depiction.ImageFileID` when `v.Depiction` is set and the id is non-empty, else NULL. | Same shape as `value_entity_id` (filled by the loader, no FK). The cache is `BucketSkip` and rebuilt by upkeep; an FK into `files` would be a file pointer that keeps an image alive after its Observation is gone (working rule "A File is released only when nothing points at it"). | Storing the Observation id and joining back — a folded value has several Observations. Storing the `depiction_values` id — two saves of the same pixels have two value rows but one File. |
| D-5 | `sqlLoadCandidates` selects `o.value_depiction_id, dpv.region_json, dpv.image_file_id, c.locator_json, dpf.media_type` after `o.value_subject_id`, with `LEFT JOIN depiction_values dpv ON dpv.id = o.value_depiction_id` and `LEFT JOIN files dpf ON dpf.id = dpv.image_file_id` after the `property_terms` join (`c` is the Citation the query already joins). `load` declares `depictionID, imageID []byte` and `region, locatorJSON, mediaType sql.NullString` in its `var (…)` block, scans them in SELECT order right after `subjectID`, and sets `Depiction: depictionvalues.FromJoin(depictionID, region, imageID, locatorJSON, mediaType)` while building `c`. The S10-32 comment and `AND p.value_type <> 'depiction'` lines are deleted. | `FromJoin` (S10-54) is the one place that turns these joined columns into a readable depiction, already used by `observations.scanListed` and `sourcefill.loadSourceObservations` (review M-9); its join shape is theirs. `Readable` inside it is the one rule (S10-32 Landed note): an image that is not the PNG Go stores is no evidence (AC-1.7). One query per batch, as today. | Scanning and calling `Readable` in the loader — a fourth copy of the assembly `FromJoin` replaced. `depictionvalues.LookupManyTx` and a second pass — a function the join does not need (Out). |
| D-6 | `hasValue` is deleted; `load` calls `autoreconcile.Carries(p.g.valueType, p.c.Value)` in its place, with the same comment. | `hasValue` is a copy of `carries` (§4). A likeness with a value row always carries (`FromJoin` is non-nil); one whose image `Readable` refused reaches the pipeline and is `no_evidence` (D-2). | Keeping `hasValue` and adding a depiction case — the copy the map row deletes. |
| D-7 | `CacheVersion` becomes 15, with the comment line `15: depiction values reconcile by image File (S10-52).` | The `CacheVersion` rule: resolution changes what the table holds (likeness rows and outcomes appear), so every catalog rebuilds once on open. | No bump — catalogs with likenesses would have no likeness rows until each handle is next written. |
| D-8 | The graph cache mirrors the column: `graphcache.Value.ImageFileID []byte` and `graphcache.TruthRow.ImageFileID []byte` *(new)*, after `Name`; `loadValues` selects `r.value_image_file_id` after `r.value_name` and sets `ImageFileID: cloneID(imageID)`; `valueEqual` adds `bytes.Equal(v.ImageFileID, t.ImageFileID)`; `truthRow` sets `ImageFileID` to a copy of `v.Depiction.ImageFileID` when `v.Depiction` is set. `graphcache` keeps a plain `[]byte`: it imports no value package (S10-54 D-6). No consumer reads it in this PR. | `Verify` compares the cache to truth field by field; an unmirrored column could drift unseen (AC-1.9). | Leaving the graph cache alone — `Verify` would not cover the new value. |
| D-9 | `conclusiondetails.loadProperties` skips `p.ValueType == properties.ValueTypeDepiction` beside the subject skip: `if !ok \|\| p.ValueType == properties.ValueTypeSubject \|\| p.ValueType == properties.ValueTypeDepiction`. `attachValues`'s `continue` comment becomes "a subject-valued or depiction Property, or one the outcomes no longer name". | Map row: the Person field list skips depiction fields; the gallery is S10-35's own surface. The skip sits in the display read, so the cache keeps every value and its reasoning. `loadOutcomes` already skips what `loadProperties` left out. | A Swift-side filter — every client of `GetConclusionDetail` would repeat it. Filtering in the reconciler — S10-35 reads the cache. |
| D-10 | A shared test-fixture package `core/database/observations/observationstest` *(new)*, file `observationstest.go`, shaped like `namevalues/namevaluestest`: `const LikenessLocator` and `const LikenessRegion` (today's `observations_test.go` values); `Crop(t testing.TB, w, h int, shade uint8) string` (a `w` × `h` PNG of `color.Gray{Y: shade}` at `filepath.Join(t.TempDir(), "crop.png")`); `FileArtifact(t testing.TB, c *database.Catalog, userID, sourceID []byte) artifacts.Artifact` (today's `fileArtifact`: a 300 × 200 red PNG through `ingest.File`, then an Artifact labelled "Photograph" through `writes.Run` action `create_artifact`); `SaveCitation(c *database.Catalog, userID []byte, art artifacts.Artifact, inputs ...observations.Input) (citations.CreateResult, []rowchange.Change, error)` (today's `saveCitation`); `UpdateObservation(c *database.Catalog, userID []byte, in observations.Input) ([]rowchange.Change, error)` (today's `updateObservation`: stage, defer discard, `observations.Update` through `writes.Run` action `update_observation`). The PNG writer is unexported inside it. `observations_test.go` deletes `writePNG`, `crop`, `fileArtifact`, `saveCitation`, `updateObservation`, `likenessLocator`, and `likenessRegion` and calls the package instead; besides those call-site renames, the only edit is its import list (drop what the moved helpers alone used, add `observationstest`). The reconciler and detail tests save likenesses through it (T-3, T-4, T-6). `TestLikenessIsSkipped` and its raw rows are deleted. | One likeness fixture instead of three: S10-35, S10-39, and S11-20 each need a saved likeness next (review M-2). The real path proves the dedupe D-1 relies on (`ingest.File` and `StagedImage.InsertTx` reuse a File by checksum) and the upkeep a region edit triggers (AC-1.8). The package imports `ingest`, `citations`, and `writes` with no cycle: only test packages import it. | Private copies in each test package — three copies of one fixture. Raw rows — they skip `writes.Run`, so they cannot test upkeep, and they pick File ids by hand. |
| D-11 | If S11-22 has landed when this PR is implemented (`ls core/propertymeta` succeeds on the `main` you branch from), every `properties.ValueType*` and `properties.Cardinality*` this plan names is `propertymeta.ValueType*` / `propertymeta.Cardinality*` (S11-22 D-6). Nothing else changes. | S11-22 (Spike 11, Planned) moves those constants with no alias. | Waiting for S11-22 — the two PRs are independent. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `autoreconcile.Value.Depiction` *(new)* | field | `Value.Date *datevalues.Value`, `Value.Name *namevalues.Value`; `observations.Listed.Depiction *depictionvalues.Value` | The same pointer-to-value-package shape as dates, names, and every other depiction holder; future depiction keys (place, event, signature) are Properties of the same type and use this field (Known trap Q8); S11-20's `match.Value.Auto()` passes it through as it does `Date` (review M-8). |
| `idKey` *(new)* | func | the subject case's inline closure in `moduleFor` | That closure's body, named once and used by subject and depiction (D-1). |
| `autoreconcile.Carries` *(new export of `carries`)* | func | `autoreconciler.hasValue` | Replaces that copy (D-6). |
| `auto_reconciler_values.value_image_file_id` *(new)*, `000053.sql` *(new)* | column, migration | `value_entity_id`, `000038.sql` | Same shape (D-4). |
| `graphcache.Value.ImageFileID`, `graphcache.TruthRow.ImageFileID` *(new)* | field | `Value.Name`, `TruthRow.Name` | Mirror D-4 (D-8). |
| `core/database/observations/observationstest` *(new)*: `Crop`, `FileArtifact`, `SaveCitation`, `UpdateObservation`, `LikenessLocator`, `LikenessRegion` | test-fixture package | `observations_test.go` `crop`, `fileArtifact`, `saveCitation`, `updateObservation`; `core/database/namevalues/namevaluestest` | Moves those helpers where three test packages can import them, in the `namevaluestest` shape (D-10). |
| Fixture wrappers `photo`, `likenessIn`, `citeOn`, `imageOf` in `core/database/autoreconciler/autoreconciler_test.go` *(new)* | helper | this file's `f.cite`, `textIn` | Bind `observationstest` to this file's fixture (`f.c`, `f.props`, `userID`); `f.cite` writes on fileless Artifacts with no staging. |

**Builds on.** The shared pipeline (`pipeline.go`, unchanged), `keyModule`, S10-32's `depictionvalues.Readable` and its effects entry, S10-53's checksum reuse, the `CacheVersion` rule, `TestRebuildEqualsUpkeep_Scenarios`.
**Deletes.** `autoreconciler.hasValue`; `knownValueType`; the subject key's inline closure body (now `idKey`); `observations_test.go`'s private likeness helpers (now `observationstest`); S10-32's D-11 filter and its comment in `sqlLoadCandidates`; `TestLikenessIsSkipped` and its raw-row block.
**Next adds.** S10-35 reads the handle's `likeness` rows with `value_image_file_id`, admits `kept` and `weak` through a named `autoreconcile.Reason` method, and resolves paths with `files.RelPaths`. S10-39's fallback reads the same rows. S10-35 also replaces D-9's skip: depiction Properties become the Person detail's gallery list, with the image on its own value type, so the skip is temporary. S10-35, S10-39, and S11-20 save test likenesses through `observationstest`, and their reads of a joined depiction call `depictionvalues.FromJoin`. The FFI handler tests' own likeness helpers (`api/ffi/handlers/citations_test.go` `likenessFixture`, `saveLikeness`; `thumbnails_test.go` `writeCrop`, from S10-54) can move onto `observationstest` in a later change; this PR does not touch them. A place or event depiction key is a seeded Property and needs no reconciler change. A seventh value type is a `moduleFor` case and a `Carries` case; if its value needs its own cache column, it also touches `autoreconcile.Value`, a migration, `sqlInsert` / `insertRow` and the loader scan, `graphcache.Value` and `TruthRow`, `loadValues`, `valueEqual`, `truth.go` `truthRow`, `conclusiondetails` (D-9 or a value field), and the test `row` / `snapshot`.

### Schema

Migration `core/database/migrations/000053.sql` *(new)*:

```sql
-- S10-52: a depiction value reconciles by its image File. The auto-reconciler
-- cache stores that File id with each depiction value. No foreign key: the
-- cache is derived (BucketSkip), and a cached id must never keep an image File
-- alive. Upkeep rewrites the row when the Observation changes.

ALTER TABLE auto_reconciler_values ADD COLUMN value_image_file_id BLOB;
```

Not registered in `catalogmodel` (a column with no FK on a `BucketSkip` table). Delete class: none. Audit: none (derived).

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` | `Value.Depiction`, `idKey`, `moduleFor`, `Carries`, `Reconcile`, `Compatible`, `knownValueType` deleted, package doc (D-1, D-2). |
| `core/database/autoreconciler` | `sqlLoadCandidates`, `load`, `hasValue` deleted, `sqlInsert`, `insertRow`, `CacheVersion` (D-4 to D-7); `truthRow` (D-8). |
| `core/database/graphcache` | `Value`, `TruthRow`, `loadValues`, `valueEqual` (D-8). |
| `core/database/conclusiondetails` | `loadProperties` skip, `attachValues` comment (D-9). |

```go
// core/autoreconcile/modules.go (sketch)
case properties.ValueTypeDepiction:
	// The same image File, so the same pixels; no readable image is no evidence.
	return keyModule{keyOf: func(v Value) (string, bool) {
		if v.Depiction == nil {
			return "", false
		}
		return idKey(v.Depiction.ImageFileID)
	}}

// idKey: an id is its own key; no id is no evidence.
func idKey(id []byte) (string, bool)
```

Errors: none new.

### FFI

None — no proto or handler change. S10-54 carries depictions over the FFI; S10-35 returns gallery paths.

### macOS

None — no Swift change. The Person field list loses nothing it shows today: `ForEntity` has never returned a likeness field (§4).

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The depiction module**
Files: `core/autoreconcile/autoreconcile.go`, `core/autoreconcile/modules.go`, `core/autoreconcile/modules_test.go`.
Do: D-1, D-2, D-3. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/...` passes; `rg 'func carries|func knownValueType' core` finds nothing.

**STEP-2 — Cache column, loader, and version**
Files: `core/database/migrations/000053.sql` *(new)*, `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/observations/observationstest/observationstest.go` *(new)*, `core/database/observations/observations_test.go`. Skills: `add-catalog-migration`, `add-catalog-query`.
Do: D-10's package first, and move `observations_test.go` onto it. `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/...` passes before the rest. Then D-4, D-5, D-6, D-7. In `autoreconciler_test.go`:
- `row` gains `ImageFileID []byte` after `ValueEntityID`; `snapshot` selects `value_image_file_id` last and scans it, so `assertUpkeepEqualsRebuild` compares it.
- Add the fixture wrappers. `(f *fixture) photo(src sources.Source) artifacts.Artifact` returns `observationstest.FileArtifact(f.t, f.c, userID, src.ID)` (every photo shares one red PNG File; `ingest.File` reuses it by checksum, which changes nothing here). `(f *fixture) likenessIn(s subjects.Subject, shade uint8) observations.Input` returns `{SubjectID: s.ID, PropertyID: f.props["likeness"].ID, Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(f.t, 40, 30, shade)}}`. `(f *fixture) citeOn(art artifacts.Artifact, in ...observations.Input) []observations.Observation` calls `observationstest.SaveCitation(f.c, userID, art, in...)`, `must`s the error, and returns `res.Observations`. `(f *fixture) imageOf(observationID []byte) []byte` reads the Observation's current image: `SELECT dv.image_file_id FROM observations o JOIN depiction_values dv ON dv.id = o.value_depiction_id WHERE o.id = ?` (by Observation id, so it follows a region edit).
- New imports: `github.com/mendahu/provenencia/core/database/observations/observationstest`.
- Delete `TestLikenessIsSkipped`. Add T-3 except its node-values comparison and its "Verify sees a stale image" subtest (both need STEP-3's field), T-4, and T-5. T-3's `assertUpkeepEqualsRebuild` and `Verify` calls are in this step. Remove the `github.com/google/uuid` import: only `TestLikenessIsSkipped` uses it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `catalogmodel.TestPragmaHonesty`, `TestLoaderQueryCountIsConstant`, both rebuild-equals-upkeep tests); `rg "value_type <> 'depiction'" core` finds nothing.

**STEP-3 — Graph cache mirror**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/load.go`, `core/database/graphcache/verify.go`, `core/database/autoreconciler/truth.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Do: D-8. Add the rest of T-3: the check that the loaded node's `likeness` values carry the rows' `ImageFileID`s (AC-1.6), and the subtest "Verify sees a stale image" (AC-1.9).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/autoreconciler/...` passes.

**STEP-4 — No Likeness field on the Person page**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusiondetails/conclusiondetails_test.go`.
Do: D-9. Add T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/... ./api/...` passes.

**STEP-5 — Docs**
Files: `docs/conclusion-reconciliation.md`, `docs/deployment-plan/spike-11/deployment-plan.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do:
- `conclusion-reconciliation.md` §7: a last table row — Value type **depiction**; Same value "the same image File (the same pixels)"; Fold "none"; Merge "the image"; No evidence "an image `depictionvalues.Readable` refuses". After the paragraph that starts "The same *same value* test": "A depiction's value is its image File; the loader reads it through `depictionvalues.Readable`, and the cache stores it as `value_image_file_id`. `likeness` is `multiple`, so a person's images form a gallery, which shows `kept` and `weak` values. The Person page lists no depiction field." §12: a row — Piece "Depictions by image File, cached as `value_image_file_id`"; Where "`core/autoreconcile`, `core/database/autoreconciler`".
- `deployment-plan.md`: in the S10-52 entry's **Plan:** line, replace the status after the link with "Landed #N"; the SLICE 8 line for S10-52 ends "— done, #N". `completed.md`: a `## S10-52 — Depiction auto-reconciler module` section above the current first section, in the shape of the `## S10-32 — Depiction value` section.
- Spike 11, whose `hasValue` work this PR does: in `docs/deployment-plan/spike-11/deployment-plan.md`, drop "`hasValue` copy" from the S11-06 row's scope cell and the sentence "`autoreconciler.hasValue` calls `carries`." from the S11-06 entry's **Ships.**; in `docs/deployment-plan/spike-11/comparison-layers.md`, the duplicates row "`autoreconciler.hasValue` (copy of `autoreconcile.carries`)" becomes "~~`autoreconciler.hasValue`~~ — done in S10-52: `autoreconcile.Carries`".
- This plan: Status and a **Landed** note per the DoD.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-52-depiction-reconciler.md` passes.

Push after STEP-2 only with STEP-3 in the same push: between them `Verify` does not compare the new column. Open the draft PR before STEP-5, so its number is known.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/modules_test.go` `TestModuleFor`, `TestModules`, and `TestReconcileDepictionMultiple` *(new)* | `TestModuleFor`'s list gains `properties.ValueTypeDepiction` → not nil. A helper `depiction(n byte, image string) Candidate` beside `term`. `TestModules` (single cardinality) rows `{"depiction", "the same image is one value", properties.ValueTypeDepiction, []Candidate{depiction(1, "a"), depiction(2, "a")}, [][]byte{{1, 2}}, StateMerged, ""}` and `{"depiction", "two images", properties.ValueTypeDepiction, []Candidate{depiction(1, "a"), depiction(2, "b")}, [][]byte{{1}, {2}}, StateMixed, ""}`. `depiction` returns `Candidate{ObservationID: id(n), Value: Value{Depiction: &depictionvalues.Value{ImageFileID: []byte(image)}}}`; `depiction(n, "")` is a likeness with an unreadable image. A third `TestModules` row `{"depiction", "an unreadable image is no evidence", properties.ValueTypeDepiction, []Candidate{depiction(1, "")}, nil, StateEmpty, ""}`. `TestReconcileDepictionMultiple`, table-driven: `a`, `a`, `b` under `CardinalityMultiple` → two values, members `{1,2}` and `{3}`, both `ReasonKept`, `StateMultiple`; one candidate with `Value{}` → `errors.Is(err, ErrValueMismatch)`; `depiction(1, "")` alone → no values, and `Outcomes[0].Reason == ReasonNoEvidence` | AC-4.1, AC-4.2, D-1, D-2 |
| T-2 | Go | `core/autoreconcile/modules_test.go` `TestCompatible` | rows "depiction: the same image" (`depiction(1, "a")`, `depiction(2, "a")`) → true; "depiction: different images" (`a`, `b`) → false; "depiction: an unreadable image is no evidence" (`depiction(1, "")`, `depiction(2, "")`) → false. `TestCompatible`'s rows pass `Candidate`s' values as its other rows do | AC-4.3, D-2 |
| T-3 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestDepictionsReconcileByImage` *(new)*, one `t.Run` per case below, each with its own `newFixture` | **Common start** (every case): `album := f.newSource("Album")`; `pr, pc, pa := f.photo(f.source), f.photo(f.other), f.photo(album)`; `a := f.subject("person")`, `b := f.subjectOn(f.other, "person")`; `oa := f.citeOn(pr, f.likenessIn(a, 100))[0]`; `ob := f.citeOn(pc, f.likenessIn(b, 200))[0]` (shade 100 in case 2); `h := f.promote(a)`; `f.c.Graph().Node(h)` (load the node, so the last write refreshes a loaded node). Rows are `rowsFor(f.rows(), h, f.props["likeness"].ID)`, matched by `ImageFileID` against `f.imageOf(oa.ID)` / `f.imageOf(ob.ID)`, never by index or rank. **1. "two images are two values"**: `f.join(b, h)` → AC-1.1, asserting each row's `Text`, `Integer`, `TermID`, `Date`, `Name`, and `ValueEntityID` are nil and both outcomes are `kept`. **2. "the same pixels are one value"**: `ob` is shade 100; `f.join(b, h)` → AC-1.2, and `bytes.Equal(f.imageOf(oa.ID), f.imageOf(ob.ID))`. **3. "a stronger negative denies"**: after the common start, `c := f.subjectOn(album, "person")`, `oc := f.citeOn(pa, negative(f.likenessIn(c, 200)))[0]`, `must(t, f.credibility(album, "high_trust"))`; then `f.join(b, h)`, `f.join(c, h)` → AC-1.3; the outcome of `ob.ID` has `DeniedBy` `oc.ID`. **4. "a provisional member"**: no `f.join`; B is filed onto `h` by `identityclaims.Create(tx, userID, identityclaims.CreateInput{SubjectID: b.ID, EntityID: h, Status: identityclaims.StatusProvisional})` inside `writes.Call` with action `create_identity_claim`, as `TestReconciledEvidence`'s claim cases do → AC-1.4. **5. "a low-trust Source is weak"**: `must(t, f.credibility(f.other, "low_trust"))`, then `f.join(b, h)` → AC-1.5. Cases 1 to 5 each end with `f.assertUpkeepEqualsRebuild(name)`, then `n, err := f.c.Graph().Node(h)` and a check that `n.Values[string(f.props["likeness"].ID)]` holds the same `ImageFileID`s as the rows (STEP-3), then `must(t, f.c.Graph().Verify())` → AC-1.6. **6. "an unreadable image is no evidence"**: case 1's writes, then `UPDATE files SET media_type = 'image/jpeg' WHERE id = ?` with `f.imageOf(ob.ID)`, `must(t, autoreconciler.Rebuild(db))`, `f.c.Graph().Drop()` → AC-1.7: one row, and the outcome for `ob.ID` has reason `no_evidence`. No `assertUpkeepEqualsRebuild` or `Verify` here: after a raw edit, upkeep differs from a rebuild by design. **7. "Verify sees a stale image"** (STEP-3): case 1's writes, `f.c.Graph().Drop()`, `UPDATE auto_reconciler_values SET value_image_file_id = ? WHERE entity_id = ? AND property_id = ? AND value_image_file_id = ?` with `bytes.Repeat([]byte{1}, 16)`, `h`, the likeness id, and `f.imageOf(oa.ID)`; `f.c.Graph().Node(h)` → `Verify()` returns an error (AC-1.9) | AC-1.1 to AC-1.7, AC-1.9, D-1, D-2, D-4 to D-8, D-10 |
| T-4 | Go | `TestRebuildEqualsUpkeep_Scenarios` row "a likeness region edit replaces its image value" *(new)* | T-3's common start without the Node load, then `f.join(b, h)`; `old := f.imageOf(oa.ID)`; `_, err := observationstest.UpdateObservation(f.c, userID, observations.Input{ID: oa.ID, SubjectID: a.ID, PropertyID: f.props["likeness"].ID, Depiction: &observations.DepictionInput{RegionJSON: region50, ImagePath: observationstest.Crop(f.t, 40, 30, 50)}})`, `must(f.t, err)`, where `region50` is the constant `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.5,"y":0.5},{"x":0.9,"y":0.5},{"x":0.9,"y":0.9}]}]}` declared in the scenario → the scenario's own check: no row has `ImageFileID` `old`, one row has `f.imageOf(oa.ID)` (read after the update; it differs from `old`); then the table's `assertUpkeepEqualsRebuild` | AC-1.8, D-5, D-10 |
| T-5 | Go | `TestEnsureCatalogRebuildsStaleVersion` | `[]int{0, 2, 3, 4, 5, 6, 7}` gains `14` → a catalog stored at 14 rebuilds and stores 15 | AC-3.1, D-7 |
| T-6 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenessHasNoField` *(new)* | `newFixture`; `src := f.source("Register")`; `a := f.personOn(src)`; `f.cite(a, f.nameIn(a, "given=James|surname=Robins"))`; `likeness, err := properties.Lookup(f.c, "likeness", properties.OriginProvenencia)`; `art := observationstest.FileArtifact(t, f.c, userID, src.ID)`; `observationstest.SaveCitation(f.c, userID, art, observations.Input{SubjectID: a.ID, PropertyID: likeness.ID, Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 40, 30, 100)}})` with its error checked; `h := f.promote(a, nil)` → `f.detail(h).Fields` has one field, key `name` | AC-2.1, D-9, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestRebuildEqualsUpkeep_Scenarios` and `TestRebuildEqualsUpkeep_SeededSequences` (the snapshot gains a column), `TestLoaderQueryCountIsConstant` (it cites no likeness, so it does not exercise D-5's join; the join adds no query by construction), every `Verify` caller (STEP-3), `catalogmodel.TestPragmaHonesty` and the schema hash (STEP-2), the `promotealign` exhibit tests (D-2 changes what `Compatible` accepts).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-52 Plan line and SLICE 8 line; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: conclusion-reconciliation.md §7 and §12 (STEP-5)
- [ ] rg "value_type <> 'depiction'|func hasValue|func carries|func knownValueType" core finds nothing
- [ ] Spike 11 docs no longer name hasValue as open work (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestDepictionsReconcileByImage|TestRebuildEqualsUpkeep_Scenarios|TestEnsureCatalogRebuildsStaleVersion|TestForEntityLikenessHasNoField|TestModules|TestModuleFor|TestReconcileDepictionMultiple|TestCompatible' -v ./core/autoreconcile/... ./core/database/autoreconciler/... ./core/database/conclusiondetails/...` and paste the depiction subtests' output in the PR. (AC-1.1 to AC-4.3)
2. On a Mac, open an existing dogfood project with the branch build. It opens (the cache rebuilds once at version 15), and a Person page shows the same fields as on `main`. An implementer without a Mac reports this check as not run.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `Compatible` now accepts `depiction`, and a Promote exhibit pins a likeness pair | `match.Value.Auto()` carries no image, so `Carries` is false (§4). A failing `promotealign` test is a stop condition. |
| A catalog with many likenesses rebuilds slowly on open | One rebuild, at the same query count per batch (D-5); `TestLoaderQueryCountIsConstant`. |
| The same pixels are stored as two Files | `StagedImage.InsertTx` reuses by checksum; T-3 "the same pixels are one value" asserts the two image ids are equal. |

**Stop and report if:**

- `depictionvalues.Readable` no longer has the signature `Readable(v Value, citationLocatorJSON, imageMediaType string) Value`, or no longer empties `ImageFileID` for a non-PNG File.
- `sqlLoadCandidates` no longer ends with S10-32's depiction filter, or `TestLikenessIsSkipped` is not in `autoreconciler_test.go`.
- T-3 "the same pixels are one value" finds two different image File ids for two shade-100 crops of one size.
- T-3 "a stronger negative denies" does not deny: the pipeline's denial needs more than an equal key, which proves D-1 incomplete.
- A test outside §8 fails with `ErrUnknownValueType`, `ErrValueMismatch`, or a likeness-shaped value, or a `promotealign` test fails after STEP-1.
- After moving `observations_test.go` onto `observationstest` (STEP-2), any `./core/database/observations/...` test fails or needs an edit beyond the helper names and imports.
- `000053.sql` is already taken on `main` (another PR added a migration): use the next free number and report it.

## 12. Open questions

None.

## 13. PR

**Title:** S10-52 — Depiction auto-reconciler module
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-52-depiction-reconciler.md`.
