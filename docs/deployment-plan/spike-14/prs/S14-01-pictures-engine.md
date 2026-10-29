# S14-01 — Pictures engine

**Status:** Landed in #615 (2026-10-28)
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: engine, proto and Swift wire mapping only; Event and Place pages still show no Pictures section)
**Board:** —
**Depends on:** — (coordinate with Spike 12 S12-24: it edits `core/database/likenesses/likenesses_test.go` and `conclusion-layer-data-model.md` §6 while Planned; not a behavior dependency — see Risks)
**Blocks:** S14-02
**Branch:** `s14-01-pictures-engine`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-24 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/depiction-values.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/catalog-session.mdc`
**Context:** [event-place-pictures.md § What changes](../event-place-pictures.md#what-changes), [§ Thumbnail parity](../event-place-pictures.md#thumbnail-parity), [§ Domain docs](../event-place-pictures.md#domain-docs-the-prs-update); working rules: "Event and Place each get their own picture Property", "The gallery layer is named gallery", "The gallery reads one list of gallery keys", "New seeds reach new catalogs only", "An Event or Place chooses its thumbnail"; [`seeded-vocabulary.md`](../../../seeded-vocabulary.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md)
**Planned:** 2026-10-28 against `main` @ `aae1f6b7`
**Review:** [`S14-01-pictures-engine.review.md`](S14-01-pictures-engine.review.md) — Fits / Ready, 2 rounds; every blocker Fixed
**Landed:** Place lead-test helper is `placePictureSubject` because `placeSubject` already exists in `placechains_test.go`. S12-24 preferred-delete audit assertions were preserved in `gallery_test.go`.

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

New catalogs seed Picture Properties for Event and Place. The gallery layer admits those keys beside `likeness`, returns each picture's own image File, and puts the lead picture on Event and Place headers the way a Person header already leads with a likeness. The gallery package and wire names say gallery/picture; Person page copy and the portrait pointer stay as they are.

**Why now.** Spike 14 batch 1 bar outcomes 1–3, 5 and 8 need the engine before the app (S14-02). Likeness audit SEC-4 ("Gallery image from the record") lands with the gallery SQL rewrite.

**What stays.** Column `preferred_likeness_observation_id`; RPC `SetPreferredLikeness` / method 101 / action `set_preferred_likeness`; Property key `likeness` and `vocabkey.PropertyLikeness`; Person page strings and portrait wording; Swift view types `ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`, `CachedThumbnail.likenessCrop` (S14-02); rectangle-only drawing (S14-03); existing catalogs (no seed backfill).

## 2. User stories

**US-1.** As the engine, when a new catalog is created, Event and Place each have a locked `depiction` Property labelled Picture, so a composer can file pictures on those kinds.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Fresh catalog after `subjectvocab.Install` | `properties.Lookup(c, "event_picture", proveniencia)` returns ValueType `depiction`, Cardinality `multiple`, Label `Picture`; binding on `event` has SortOrder 5 and Locked true |
| AC-1.2 | Same catalog | `place_picture` matches AC-1.1 shape; binding on `place` has SortOrder 3 and Locked true |
| AC-1.3 | `vocabkey.GalleryKeys()` | Equals `likeness`, `event_picture`, `place_picture` in that order; each key is a seeded Property with ValueType `depiction` |

**US-2.** As the engine, when an Event or Place has gallery-eligible picture Observations, `gallery.ForEntities` lists them and a Person's likeness gallery is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Place Subject with one positive `place_picture` depiction (kept); `gallery.ForEntities` for its handle | One picture; `ObservationIDs[0]` is that Observation; Person-only fixtures from today's `TestForEntities` still pass after the rename |
| AC-2.2 | Event Subject with one positive `event_picture` depiction | Same as AC-2.1 for the Event handle |

**US-3.** As the engine, when a gallery tile is read, its image File is the Observation's `depiction_values.image_file_id`, not `auto_reconciler_values.value_image_file_id`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Person with one likeness; `UPDATE auto_reconciler_values SET value_image_file_id = <other file>` for that value row; `gallery.ForEntities` | Tile `ImageFileID` equals `observationstest.ImageOf` of the likeness Observation (the depiction row), not the edited cache id |

**US-4.** As the engine, when an Event or Place has pictures, its header leads with the earliest picture, then the chosen one after `SetPreferredLikeness`, and the detail wire carries that path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Event with pictures A then B (earliest A); `conclusionheaders.EventsByIDs` / `ListEvents` | `PictureImageFileID` equals A's image |
| AC-4.2 | AC-4.1; `SetPreferredLikeness` of B's record Observation; list and ByIDs again | `PictureImageFileID` equals B's image; clear pointer → A again |
| AC-4.3 | Same sequence on a Place | Same results on `PlaceHeader.PictureImageFileID` |
| AC-4.4 | `GetConclusionDetail` for that Event (handler) | `EventHeader.picture_image_rel_path` equals the leading tile's `image_rel_path` among `ConclusionDetail.pictures` |

**US-5.** As a contributor, after the gallery rename, Go, proto and the Swift wire mapping compile and the existing Person gallery/portrait tests stay green.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Package import path | `core/database/gallery` exists; no remaining import of `core/database/likenesses` |
| AC-5.2 | Proto | Message `PictureDetail` (field numbers of today's `LikenessDetail`); `ConclusionDetail.pictures = 7`; `PersonHeader.picture_image_rel_path = 8`; `EventHeader.picture_image_rel_path = 17`; `PlaceHeader.picture_image_rel_path = 14` |
| AC-5.3 | Swift wire | `CatalogPicture`, `CatalogConclusionDetail.pictures`, `CatalogPersonHeader.pictureImageRelPath` (and Event/Place); `GoStore` maps them; app and `ProvenenciaTests` build |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey.PropertyEventPicture`, `PropertyPlacePicture`, `GalleryKeys()` *(new)* | STEP-1, AC-1.3, D-1, D-13 |
| Seed `event_picture` / `place_picture` in `core/database/subjectvocab/registry.go` (Install only; no migration) | STEP-1, AC-1.1, AC-1.2, D-2 |
| `TestSeedProductValues` cases for both keys; extend `TestVocabKeysMatchSeed`; `TestGalleryKeys` order assert | STEP-1, T-1, T-2, D-13 |
| Rename package `likenesses` → `gallery`; type `Likeness` → `Picture`; update callers | STEP-2, AC-5.1, D-3 |
| Gallery SQL: filter `p.key IN` `GalleryKeys()`; image from observation → `depiction_values` (SEC-4) | STEP-2, AC-2.*, AC-3.1, D-4, D-5 |
| `observationstest.SaveDepiction(…, key, …)` *(new)*; `SaveLikeness` / `SaveNegativeLikeness` wrappers | STEP-2, D-6 |
| `PersonHeader` / `EventHeader` / `PlaceHeader` `PictureImageFileID`; `leadPictures` helper; compose three kinds; table-driven lead tests | STEP-3, AC-4.1–AC-4.3, D-7, D-12 |
| Proto rename `LikenessDetail` → `PictureDetail`, `likenesses` → `pictures`, `likeness_image_rel_path` → `picture_image_rel_path` (keep numbers); new Event/Place fields; regenerate | STEP-4, AC-5.2, D-8 |
| FFI: `attachHeader` returns lead for all three kinds; `attachPictures` and three list handlers call `files.RelPaths` (D-11); `depictionFixture` | STEP-5, AC-4.4, T-7, D-11, D-14 |
| Swift wire: `CatalogPicture`, `pictures`, `pictureImageRelPath` on three headers; `GoStore` / `GenealogyStore` / FakeStore / tests compile | STEP-6, AC-5.3, D-9 |
| Domain docs (incl. SEC-4 gallery image clause), `add-depiction-value` skill (both trees), `.cursor/rules/depiction-values.mdc` glob → `gallery`, map/completed/plan status | STEP-7 |

| Out | Home |
| --- | --- |
| Event/Place Pictures section, gallery slot, list thumbnails, Promote "picture" wording, Swift view-layer rename (`ConclusionDetailLikeness*`, `LikenessTileView`, `CachedThumbnail.likenessCrop`) | S14-02 |
| Rectangle-only depiction drawing; composer draw feedback | S14-03 |
| Renaming column `preferred_likeness_observation_id` | not planned (working rule) |
| Renaming RPC `SetPreferredLikeness` / method 101 / `set_preferred_likeness` | not planned (column-named API; working rule keeps the column) |
| Seed backfill into existing catalogs | not planned (working rule "New seeds reach new catalogs only") |
| Property key or label `depiction` | not planned (design) |
| Go-side rectangle enforcement on stored regions | S14-03 Out / not planned |
| Spike 13 Gallery section redesign | S13-14 |
| Person Timeline event picture | Spike 13 R-26 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Gallery package | `core/database/likenesses/likenesses.go` `ForEntities`, `Lead`, `ByObservation`, type `Likeness`, `sqlGallery` | Filters `p.key = ?` with `vocabkey.PropertyLikeness`; reads `v.value_image_file_id` + files media type through `depictionvalue.ReadableImage` |
| Gallery tests | `core/database/likenesses/likenesses_test.go` `TestForEntities`, `TestForEntitiesPreferred`, `TestSetPreferredLikeness` | Person + `SaveLikeness` only |
| Vocab keys | `core/vocabkey/vocabkey.go` | `PropertyLikeness = "likeness"`; no Event/Place picture keys; no `GalleryKeys` |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties` / `seedBindings` | One depiction Property: `likeness` on person sort 1 locked; Event bindings end at `end_date` sort 4; Place at `end_date` sort 2 |
| Seed tests | `core/database/subjectvocab/subjectvocab_test.go` `TestSeedProductValues` | Likeness case `onPerson: true`; no Event/Place picture cases; no gallery-key test |
| Preferred pointer | `core/database/canonicalentities/canonicalentities.go` `SetPreferredLikeness` | Membership via `likenesses.ForEntities` + `ByObservation`; no kind check |
| Detail tiles | `core/database/conclusiondetails/conclusiondetails.go` `LikenessTile`, `Detail.Likenesses`, `ForEntity` | Already calls `likenesses.ForEntities` for every kind; tiles empty for Event/Place because SQL filters likeness only |
| Person header lead | `core/database/conclusionheaders/compose.go` Persons path; `PersonHeader.LikenessImageFileID` | Loads gallery, sets lead; tests `TestPersonHeaderLeadsWithEarliestLikeness`, `TestPersonHeaderLeadsWithPortrait` |
| Event/Place headers | `core/database/conclusionheaders/events.go` `EventHeader`; `core/database/conclusionheaders/places.go` `PlaceHeader`; `core/database/conclusionheaders/compose.go` `oneEvent`, `placeShell` | No image field; compose does not call the gallery |
| Detail FFI | `api/ffi/handlers/conclusion_detail.go` `attachHeader`, `attachLikenesses` | Lead returned and path set only for Person |
| List FFI | `api/ffi/handlers/conclusion_headers.go` `ListPersonHeaders` / `ListEventHeaders` / `ListPlaceHeaders` | Only Person list fills likeness paths |
| Proto | `api/proto/engine.proto` | `LikenessDetail`; `ConclusionDetail.likenesses = 7`; `PersonHeader.likeness_image_rel_path = 8`; EventHeader last field 16; PlaceHeader last field 13 |
| Swift wire | `GenealogyStore.swift` `CatalogLikeness`, `CatalogConclusionDetail.likenesses`, `CatalogPersonHeader.likenessImageRelPath`; `GoStore.mapPersonHeader` / detail map | Event/Place headers have no image path |
| Test helper | `core/database/observations/observationstest/observationstest.go` `SaveLikeness`, `saveLikeness` | Hard-codes property key `likeness` |
| Docs | `docs/seeded-vocabulary.md` §3.2–3.3; `interpretation-layer-data-model.md` depiction paragraph; `conclusion-layer-data-model.md` §6; `conclusion-reconciliation.md` §7; `.agents/skills/add-depiction-value/SKILL.md` (+ `.claude` twin) | Person/`likeness` only; skill names `likenesses` and `SaveLikeness` |

**Precedent.** [#446](https://github.com/mendahu/provenencia/pull/446) (`f404522b`, `17223d66`) created `likenesses.ForEntities` and Person header lead — extract that load into `leadPictures` and call it from all three compose paths. [#430](https://github.com/mendahu/provenencia/pull/430) (`e5f67980`) seeded locked `likeness` — seed Event/Place pictures the same way (`seedProperties` + `seedBindings`, Install only). [#450](https://github.com/mendahu/provenencia/pull/450) preferred pointer — keep `SetPreferredLikeness` and the column name. Do not copy a kind check into `SetPreferredLikeness`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `vocabkey.GalleryKeys()` returns `[]string{"likeness", "event_picture", "place_picture"}` in that order; `PropertyEventPicture` / `PropertyPlacePicture` constants sit beside `PropertyLikeness` | Working rule "one list of gallery keys"; next depiction Property that belongs in a gallery is one append | Filter on `PropertyLikeness` alone — Event/Place never appear; hard-code three keys only in SQL with no list — next key needs a SQL edit |
| D-2 | Seed rows: `event_picture` / `place_picture`, Label `Picture`, Description `A picture of this event: a region of a cited page and the image of it. More than one is allowed.` (and the same for place), ValueType depiction, Cardinality multiple; bindings event sort 5 locked, place sort 3 locked | Working rule + design table; mirrors `likeness` description with kind noun | One shared `picture` key for Event and Place — rejected by working rule; migration INSERT — rejected by "New seeds reach new catalogs only" |
| D-3 | Package directory `core/database/gallery/` *(new)*; type `Picture`; funcs keep `ForEntities` / `ByObservation` / `Lead` | Working rule examples (`likenesses` → `gallery`; item = picture) | Keep package name `likenesses` with only SQL changes — layer still named for the Person key |
| D-4 | `sqlGallery` joins `observations o ON o.id = ao.observation_id` and `depiction_values dv ON dv.id = o.value_depiction_id`, selects `dv.image_file_id` and that file's media type; still groups by `(entity_id, rank)` from reconciler values | SEC-4; `ImageOf` already reads the depiction row | Keep reading `value_image_file_id` — edited cache can show another PNG |
| D-5 | Gallery property filter: package-level SQL head built with `database.SQLInPlaceholders(len(vocabkey.GalleryKeys()))` for `p.key IN (…)` after `p.origin = ?`; `sqlbatch.Query` prefix args are origin then each `GalleryKeys()` entry, then entity ids | `core/database/sqlin.go` `SQLInPlaceholders` is the repo helper; `sqlbatch.Query` already binds prefix args then the id batch | Hand-rolled `strings.Repeat("?,", …)`; `OR` of equality predicates; post-filter in Go |
| D-6 | `SaveDepiction(t, c, userID, sourceID, subjectID, key string, shade uint8)` looks up `key` at proveniencia origin; `SaveLikeness` / `SaveNegativeLikeness` call it with `"likeness"` (negative keeps a polarity helper that calls an internal with polarity) | Scope + skill update; existing call sites stay one-liners | Duplicate SaveEventPicture helpers — three near-copies |
| D-7 | Rename `PersonHeader.LikenessImageFileID` → `PictureImageFileID`; add the same field on `EventHeader` and `PlaceHeader`. Private `leadPictures(g, ids) (map[string][]byte, error)` *(new)* in `conclusionheaders` wraps `gallery.ForEntities` + `gallery.Lead`. `composePersons` / `composeEvents` / `composePlaces` call it for their `need` ids and set `PictureImageFileID`; never set the lead inside `onePerson`'s callers' `placeShell` / `oneEvent` body — `placeShell` stays shell-only (parent-chain callers) | Working rule "one mechanism"; one compose helper so a later lead change cannot miss a kind | Copy the Persons gallery block into Events and Places separately; set lead inside `placeShell` — would run per parent-chain node |
| D-8 | Proto: rename message `LikenessDetail` → `PictureDetail`; field `ConclusionDetail.pictures = 7`; rename `PersonHeader` field 8 to `picture_image_rel_path`; add `picture_image_rel_path = 17` on `EventHeader` and `= 14` on `PlaceHeader`; keep `SetPreferredLikeness*` message names | Working rule (rename layer names, keep field numbers and the preferred API) | New message beside `LikenessDetail` — second wire type; rename preferred RPC — column-named Out |
| D-9 | Swift wire: `CatalogPicture`; `CatalogConclusionDetail.pictures`; `pictureImageRelPath` on `CatalogPersonHeader` / `CatalogEventHeader` / `CatalogPlaceHeader`; feature views and tests update only identifiers that must rename for compile (read `detail.pictures`, `header.pictureImageRelPath`); leave `ConclusionDetailLikenessSection` type names to S14-02 | Map scope "Swift wire mapping renamed to compile" | Rename view-layer types in this PR — S14-02 owns them |
| D-10 | `TestSeedProductValues`: replace `onPerson bool` with `bindType string` (`"person"` / `"event"` / `"place"` / `""`); empty skips binding asserts; likeness uses `"person"`; new cases for both picture keys | Existing table only checked person bindings | Separate TestEventPictureSeed — second table for the same product-values check |
| D-11 | No `pictureRelPaths` wrapper (`files.RelPaths` already unique-ids). Each of `ListPersonHeaders` / `ListEventHeaders` / `ListPlaceHeaders` keeps its own short collect-`PictureImageFileID` then `files.RelPaths` then assign loop (same shape as today's Person list). `attachPictures` collects lead + tile ids, calls `files.RelPaths` once, sets `PictureImageRelPath` on the Person, Event, or Place header present and fills tile paths | `files.RelPaths` is the shared resolver; a pass-through helper adds a name with no work | New `pictureRelPaths` that only forwards to `RelPaths`; a generic `leadPaths[H]` helper — three five-line loops stay clearer than a type-param collector for three call sites |
| D-12 | Rename `TestPersonHeaderLeadsWithEarliestLikeness` → `TestHeaderLeadsWithEarliestPicture` and `TestPersonHeaderLeadsWithPortrait` → `TestHeaderLeadsWithChosenPicture`; each is one table over kind rows `{person/likeness, event/event_picture, place/place_picture}`. Row setup returns `(subjects.Subject, handleID []byte)`: person keeps today's `f.person` (use subject + handle); event/place add `eventSubject` / `placeSubject` *(new)* helpers that create, optionally cite, promote, and return subject + handle (do not change existing `f.event` / `f.place` signatures used elsewhere). Then `SaveDepiction` / `SaveLikeness` on the subject; list/ByIDs/lead extractors per kind. `TestPersonHeaderSkipsNegativeLikeness` stays Person-only | `.cursor/rules/go-tests.mdc` one `TestX` per behavior family; `f.event`/`f.place` today return only `[]byte` so they cannot feed `SaveDepiction` alone | Separate mirrored `TestEventHeader…` functions; overloading `f.event`/`f.place` return types and breaking existing callers |
| D-13 | Extend `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` with `PropertyLikeness`, `PropertyEventPicture`, `PropertyPlacePicture` in its seeded-Property loop, and assert each `GalleryKeys()` entry is seeded `depiction`; add `TestGalleryKeys` in `core/vocabkey/vocabkey_test.go` with `slices.Equal(GalleryKeys(), []string{PropertyLikeness, PropertyEventPicture, PropertyPlacePicture})` | One seed-vs-vocabkey test already exists; AC-1.3 order needs an exact assert | New `TestGalleryKeysAreSeededDepictions` beside `TestVocabKeysMatchSeed` |
| D-14 | FFI test helper `depictionFixture(t, kind, key string)` *(new)* beside `likenessFixture` in `api/ffi/handlers/citations_test.go`; `likenessFixture` becomes a wrapper for `("person", "likeness")`. T-7 uses `depictionFixture(t, "event", "event_picture")` | Same generalization as D-6; `saveLikeness` already takes property id | Second 30-line Event fixture |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `vocabkey.PropertyEventPicture` / `PropertyPlacePicture` / `GalleryKeys` *(new)* | const / func | `PropertyLikeness` | New keys; list is the working-rule extension point |
| `event_picture` / `place_picture` seed rows *(new)* | registry rows | `likeness` seed row | New Properties; same registry lists |
| Package `core/database/gallery/` *(new)* | package | `core/database/likenesses/` | Rename of that package (D-3), not a second gallery |
| `gallery.Picture` *(new name)* | type | `likenesses.Likeness` | Rename |
| `observationstest.SaveDepiction` *(new)* | helper | `SaveLikeness` | Generalizes it; wrappers remain |
| `EventHeader.PictureImageFileID` / `PlaceHeader.PictureImageFileID` *(new)* | fields | `PersonHeader` lead field | Same lead rule on the other headers |
| `conclusionheaders.leadPictures` *(new)* | func | Persons compose block at `compose.go` ~L485–494 | Shared helper (D-7), not three copies |
| `handlers.depictionFixture` *(new)* | test helper | `likenessFixture` in `citations_test.go` | Generalizes it (D-14) |
| Proto `EventHeader`/`PlaceHeader` `picture_image_rel_path` *(new)* | fields | Person field 8 | Parity |
| `CatalogEventHeader`/`CatalogPlaceHeader.pictureImageRelPath` *(new)* | fields | Person wire field | Parity |

**Builds on.** `subjectvocab` Install registry (`add-seeded-vocabulary`); `TestVocabKeysMatchSeed`; `gallery.ForEntities` + `Lead` + `SetPreferredLikeness` membership (no kind check); Persons compose gallery load; `database.SQLInPlaceholders` + `sqlbatch.Query` prefix args; `observationstest.ImageOf` join shape for SEC-4; `./scripts/generate-proto.sh` + `add-ffi-handler`; header fixtures `f.event` / `f.place`.
**Deletes.** Package path `core/database/likenesses/` (moved); type name `Likeness` / `LikenessTile` / proto `LikenessDetail` / Swift `CatalogLikeness`; identifiers `Detail.Likenesses`, `ConclusionDetail.likenesses`, `PersonHeader.LikenessImageFileID`, `likeness_image_rel_path`, `attachLikenesses` (renamed to `attachPictures`); SQL use of `v.value_image_file_id` as the tile image source (replaced by depiction join). `SaveLikeness` remains as a wrapper. `SetPreferredLikeness` and the preferred column stay. Separate per-kind header lead test functions are not added (D-12 folds them into the Person tests' tables).
**Next adds.** S14-02 reads Event/Place `pictureImageRelPath` and `pictures` for pages, lists and the gallery slot. A fourth gallery key is one `GalleryKeys` entry plus a seed row and one kind row in the lead tests. Spike 13 R-26 can read Event pictures once headers carry them. S12-24's preferred-delete audit assertions live in the moved gallery test file after this rename.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Event/Place handle with gallery pictures | `gallery.ForEntities`, header compose, `GetConclusionDetail` pictures | Empty gallery → nil/`""` lead (same as Person today); preferred Observation not in gallery → earliest leads (`Lead` / `ByObservation` already); merged handle still `not_found` on detail |
| `SetPreferredLikeness` on Event/Place after pictures exist | `canonicalentities.SetPreferredLikeness` | Already refuses non-gallery Observation (`ErrInvalid`); display cache Verify still compares fresh Event/Place headers including `PictureImageFileID` |

### Scale

At Event/Place list size: `leadPictures` runs one `gallery.ForEntities` call over the ids being composed (same batching as Persons today via `sqlbatch`). `SetPreferredLikeness` still one row update. No per-child fan-out write. Tile image join is one extra join per gallery row, not a second query per tile. Search re-index (`RowsForSearch` / `EventsFromQuerier` / `PlacesFromQuerier`) pays one batched gallery query per write that re-composes Event or Place headers — the same as Person re-index already pays for likeness leads.

### Schema

None — no catalog change. Seeds arrive at create-time Install only.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `PropertyEventPicture`, `PropertyPlacePicture`; `func GalleryKeys() []string` *(new)* |
| `core/database/subjectvocab` | Two `seedProperties` + two `seedBindings`; `TestSeedProductValues`; extend `TestVocabKeysMatchSeed` (D-13) |
| `core/database/gallery` *(renamed)* | `Picture`; `ForEntities` uses D-4/D-5 SQL; tests gain Place/Event cases + SEC-4 case |
| `core/database/observations/observationstest` | `SaveDepiction` *(new)*; wrappers |
| `core/database/canonicalentities` | Import `gallery`; comments say gallery pictures |
| `core/database/conclusiondetails` | `PictureTile` *(rename)*; `Detail.Pictures`; import `gallery` |
| `core/database/conclusionheaders` | `PictureImageFileID` on three headers; `leadPictures` *(new)*; three compose paths call it |
| `core/database/autoreconciler` | Test import path only if it imports the gallery package |

Errors: none new. `SetPreferredLikeness` still `canonicalentities.ErrInvalid` when the Observation is not in the gallery.

### FFI

`api/proto/engine.proto`: rename per D-8; regenerate with `./scripts/generate-proto.sh`.
Handlers: D-11 (no pass-through helper). `conclusion_detail.go` `attachHeader` returns lead file id for Person, Event and Place; rename `attachLikenesses` → `attachPictures` (one `files.RelPaths` for lead + tiles; set `PictureImageRelPath` on the header present). `conclusion_headers.go` three list handlers each collect leads and call `files.RelPaths`. No new Method. `SetPreferredLikeness` handler comments say handle gallery portrait/thumbnail pointer (still Person-oriented copy in the app until S14-02). Tests: `depictionFixture` (D-14).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire tile | `CatalogPicture` *(rename)* | `GenealogyStore.swift` | Map in `GoStore` |
| Detail body field | `pictures` | `CatalogConclusionDetail` | Feature views read it to compile |
| Header lead path | `pictureImageRelPath` | Three `Catalog*Header` types | Lists may ignore Event/Place until S14-02 |

States: empty path `""` when no picture — unchanged. Accessibility: no new chrome.

### Strings

None — no L10n keys. Sweep still `rg`s renamed Go/proto/Swift wire names across `.agents/`, `.cursor/rules/`, `docs/*.md` and skills. Person section title `L10n.Conclusions.likenessesSection` stays (S14-02).

## 7. Steps

**STEP-1 — Gallery keys and seeds**
Files: `core/vocabkey/vocabkey.go`, `core/vocabkey/vocabkey_test.go`, `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/subjectvocab_test.go`, `core/database/subjectvocab/vocabkey_test.go`. Skill: `add-seeded-vocabulary`.
Do: Add D-1 constants and `GalleryKeys`. Add D-2 seed properties and bindings. Extend `TestSeedProductValues` per D-10 with `event_picture` and `place_picture` cases. Apply D-13: extend `TestVocabKeysMatchSeed`; add `TestGalleryKeys` with exact order assert for AC-1.3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/vocabkey/ ./core/database/subjectvocab/` passes.

**STEP-2 — Gallery package rename, multi-key filter, SEC-4 image, SaveDepiction**
Files: move `core/database/likenesses/likenesses.go` → `core/database/gallery/gallery.go` *(new)* and `likenesses_test.go` → `gallery_test.go` *(new)*; rewrite SQL per D-4/D-5; rename type to `Picture`; update imports in `core/database/canonicalentities/canonicalentities.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusiondetails/conclusiondetails_test.go` (rename `LikenessTile` → `PictureTile`, `Likenesses` → `Pictures` here), `core/database/conclusionheaders/compose.go`, `core/database/autoreconciler/autoreconciler_test.go`; update `api/ffi/handlers/conclusion_detail.go` field reads (`d.Pictures`) so the handler package type-checks, leaving path-fill behavior for STEP-5; `core/database/observations/observationstest/observationstest.go` per D-6 (reword package doc off "saves likenesses" only); extend gallery tests for Place/Event via `subjectOn(src, kind)` *(new)* (AC-2) and SEC-4 (AC-3.1). If S12-24 has already extended the preferred-delete subtest on `main`, keep those assertions in `gallery_test.go`. Skill: `add-depiction-value`.
Do: Mechanical rename + SQL + helper. Keep Person gallery cases green under new names. Rename locals that today shadow the package after import (`gallery, err := likenesses.ForEntities` in `compose.go` and `canonicalentities.go`) to `tiles` (or `pictures`). Gallery-test subject helper: one `subjectOn(src, kind string)` *(new)* beside today's `personOn`, not separate `eventOn`/`placeOn` near-copies. This step does not claim `./api/ffi/...` green beyond type-check of the renamed detail fields.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/gallery/ ./core/database/canonicalentities/ ./core/database/observations/... ./core/database/conclusiondetails/` passes; `rg 'core/database/likenesses' --glob '*.go'` is empty among packages this step owns (AC-5.1 for the package path).

**STEP-3 — Header lead on all three kinds**
Files: `core/database/conclusionheaders/conclusionheaders.go`, `core/database/conclusionheaders/events.go`, `core/database/conclusionheaders/places.go`, `core/database/conclusionheaders/compose.go`, `core/database/conclusionheaders/conclusionheaders_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`. Skill: —.
Do: D-7 (`leadPictures`; three compose paths; `placeShell` unchanged). D-12: rename and table-drive the two lead tests; add `eventSubject` / `placeSubject` returning `(Subject, handle)`; save depictions on the subject; matching list/ByIDs/lead extractors. Extend detail tests so a Place's pictures appear in `Detail.Pictures` (reuse gallery `subjectOn` from STEP-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/ ./core/database/conclusiondetails/` passes.

**STEP-4 — Proto rename and Event/Place fields**
Files: `api/proto/engine.proto`; generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler` (regen).
Do: D-8. Run `./scripts/generate-proto.sh`. Fix Go compile errors in handlers only as needed to reach STEP-5.
Check: generated files committed; `gofmt` clean on touched Go; proto field numbers match AC-5.2.

**STEP-5 — FFI attach lead for every kind**
Files: `api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/conclusion_headers.go`, `api/ffi/handlers/citations_test.go` (`depictionFixture`), `api/ffi/handlers/conclusion_detail_test.go`, `api/ffi/handlers/conclusion_headers_test.go`. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-11 (per-handler collect + `files.RelPaths`; `attachPictures` one RelPaths for lead + tiles). `attachHeader` returns lead id for Event and Place. D-14: `depictionFixture`; T-7 builds an Event via `depictionFixture(t, "event", "event_picture")`. Existing Person preferred tests updated to new proto names stay green.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-6 — Swift wire mapping**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift` if it constructs likeness tiles; Catalog call sites that break on rename (`macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Features/Conclusions/PersonsListView.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`, `macos/ProvenenciaTests/PersonsListTests.swift`). Do not rename `ConclusionDetailLikenessSection` or edit `ConclusionDetailPage.swift` view-layer types (S14-02). Skill: —.
Do: D-9. Map Event/Place `pictureImageRelPath` in `mapEventHeader` / `mapPlaceHeader`. App and unit tests compile.
Check: On the Mac CI/`scripts/check-pr.sh` runner, Xcode tests pass with the renamed Catalog fields. A Linux cloud agent records the Go side green and leaves Xcode evidence to CI when `xcode-test` cannot run locally.

**STEP-7 — Docs**
Files: `docs/seeded-vocabulary.md`, `docs/interpretation-layer-data-model.md`, `docs/conclusion-layer-data-model.md`, `docs/conclusion-reconciliation.md`, `.agents/skills/add-depiction-value/SKILL.md`, `.claude/skills/add-depiction-value/SKILL.md`, `.cursor/rules/depiction-values.mdc` (glob `core/database/likenesses/**` → `core/database/gallery/**`), `docs/deployment-plan/spike-14/event-place-pictures.md` Domain docs bullet (preferred pointer is the handle's), `completed.md`, `deployment-plan.md` map row, this plan (status + Landed note when merged).
Do: Per Definition of done and [design § Domain docs](../event-place-pictures.md#domain-docs-the-prs-update), with the working-rule correction that the preferred pointer is the handle's (not Person-only). In `conclusion-reconciliation.md` and the skill's gallery bullet: the gallery shows each record's own `depiction_values.image_file_id`, not the reconciler cache copy (SEC-4). Skill text: gallery package, `GalleryKeys`, `SaveDepiction`; Do-not line still forbids a likeness field on the Person page. If S12-24 already rewrote the preferred-pointer audit sentence, keep that sentence and widen "Person" to handle.
Check: links resolve; `python3 scripts/check-guidance-names.py` passes; `rg` old package path and `LikenessDetail` / `CatalogLikeness` in skills/rules/docs has no stale hit the sweep requires.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Seed Picture Properties for Event and Place."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestSeedProductValues` cases `event_picture`, `place_picture` | Registry label/type/cardinality/binding order/lock match D-2 | AC-1.1, AC-1.2, D-2, D-10 |
| T-2 | Go | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` extended; `core/vocabkey/vocabkey_test.go` `TestGalleryKeys` *(new)* | Seeded depiction keys include likeness + both pictures; `GalleryKeys()` equals `[likeness, event_picture, place_picture]` | AC-1.3, D-1, D-13 |
| T-3 | Go | `core/database/gallery/gallery_test.go` *(new)* Place/Event cases | `SaveDepiction` place_picture / event_picture → one gallery picture on that handle; Person likeness fixture unchanged | AC-2.1, AC-2.2 |
| T-4 | Go | `core/database/gallery/gallery_test.go` *(new)* SEC-4 case | Edited `value_image_file_id` → tile still `ImageOf(obs)` | AC-3.1, D-4 |
| T-5 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestHeaderLeadsWithEarliestPicture` / `TestHeaderLeadsWithChosenPicture` *(renamed)* | Per kind row via `eventSubject`/`placeSubject`: earliest then preferred then clear | AC-4.1, AC-4.2, AC-4.3, D-7, D-12 |
| T-6 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` Place pictures case *(new)* | Place with picture → `Detail.Pictures` non-empty; Person cases renamed stay green | AC-2.1, AC-5.1 |
| T-7 | Go | `api/ffi/handlers/conclusion_detail_test.go` Event header path *(new)* via `depictionFixture` | Detail Event `picture_image_rel_path` == leading `pictures[i].image_rel_path` where `leads` | AC-4.4, AC-5.2, D-14 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` / `PersonsListTests.swift` updated to `CatalogPicture` / `pictureImageRelPath` / `pictures` | Same assertions as today after rename | AC-5.3, D-9 |
| T-9 | Go | STEP-2 check + STEP-4 check (commands, not a `TestX`) | `rg 'core/database/likenesses' --glob '*.go'` empty; proto field numbers match AC-5.2 | AC-5.1, AC-5.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `core/database/gallery/` *(new)*, `core/database/conclusionheaders/`, `core/database/conclusiondetails/`, `api/ffi/handlers/`, `ProvenenciaTests` Person detail/list.

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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: seeded-vocabulary.md; interpretation-layer-data-model.md depiction paragraph; conclusion-layer-data-model.md §6 gallery + preferred pointer as the handle's; conclusion-reconciliation.md depiction gallery sentence + SEC-4 image source; add-depiction-value skill (both trees); `.cursor/rules/depiction-values.mdc` gallery glob
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Engine PR: no new chrome to click. Automated tests cover AC-1.* through AC-5.*. Smoke after merge (before S14-02):

1. Create a **new** project. On an Event Subject in the composer, the Property list offers Picture (`event_picture`). On a Place Subject, Picture (`place_picture`). (AC-1.1, AC-1.2)
2. Open an existing dogfood project: Person likeness gallery and portrait still behave as before; no Event/Place Picture Properties appear (create-time seed only).

Slice Check (bar outcomes 1–9) is completed by S14-02 + S14-03 together; this PR's automated tests are the gate.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Display cache for Event/Place omits picture lead after prefer | Table-driven lead tests call `Graph().Verify()` after `SetPreferredLikeness` (T-5) |
| Proto rename breaks Swift call sites beyond wire | STEP-6 lists Catalog call sites only; compile is the check |
| SEC-4 join drops rows when depiction missing | Same as today's nil image drop via `ReadableImage`; stop if gallery empties for a known good likeness fixture |
| S12-24 (Spike 12, Planned) edits `likenesses_test.go` and `conclusion-layer-data-model.md` §6 | STEP-2 keeps any preferred-delete audit assertions already on `main`; STEP-7 keeps S12-24's audit sentence when present and widens Person→handle |

**Stop and report if:**

- `gallery.ForEntities` / `SetPreferredLikeness` membership no longer matches §4 after a `main` move.
- Event/Place header display cache cannot include `PictureImageFileID` without changing graphcache APIs (Out).
- `GalleryKeys` length 0 or a key lacks a depiction seed (T-2 fails and registry cannot be fixed inside In).
- Regenerating proto requires changing field numbers to compile (Out: numbers are fixed).
- Merging would drop S12-24's preferred-delete audit assertions that already landed on `main` in the gallery test file.

## 12. Open questions

None.

## 13. PR

**Title:** S14-01 — Pictures engine
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-01-pictures-engine.md`.
