# S10-35 — Person gallery and Persons list thumb

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-35--person-gallery-and-list-thumb)
**Brief:** [S10-D12](../design/S10-D12-person-detail.md) (PDP-7 to PDP-9 ship here), board agreed 2026-10-16; [S10-D9](../design/archive/S10-D9-persons-list.md) (PSL-7 ships here), board agreed before S10-22
**Board:** [Person Detail](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, turn 2, frames 2j (Gallery) and 2k (Crop states); [Persons list](https://claude.ai/design/p/5c00cb55-8261-4579-b0e6-bc571ab23f78?via=share), file *Persons List.dc.html*, frame 08. Checked against exports the user attached on 2026-10-17: frame 08 matches PSL-7; frames 2j and 2k match PDP-7 to PDP-9, plus layout and copy the brief did not give (a 120pt tile column, gaps, the tile and header-thumb labels), which §6 records. §6 macOS describes every frame this PR ships in words. The implementer does not need the boards.
**Depends on:** S10-23 (landed [#407](https://github.com/mendahu/provenencia/pull/407)), S10-52 (landed [#437](https://github.com/mendahu/provenencia/pull/437)), S10-54 (landed [#433](https://github.com/mendahu/provenencia/pull/433)), S10-34 (landed [#442](https://github.com/mendahu/provenencia/pull/442); no code dependency, the manual check draws faces in its composer), S10-D12, S10-D9
**Blocks:** S10-39
**Branch:** `s10-35-person-gallery`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rules "A gallery is the reconciled likeness values", "A preferred portrait is a display pointer, not a claim", "A depiction's value is its own row: region plus image"; [depictions and likenesses](../depictions-and-likenesses.md); [S10-52 plan](S10-52-depiction-reconciler.md) D-3, D-9, Next adds, and its ship log SEC-1 ([`S10-52-depiction-reconciler.ship.md`](S10-52-depiction-reconciler.ship.md)); [S10-23 plan](S10-23-person-gender.md) D-6, D-7, Next adds; [S10-54 plan](S10-54-depiction-ffi.md) D-4 (`files.RelPaths`); [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §10 (cache)
**Planned:** 2026-10-17 against `main` @ `3ef0553` (S10-34 and S11-02 landed); first planned against `500396f`. The refresh changed no Current state row: S11-02 moved `NormalizeForm` out of `core/autoreconcile/autoreconcile.go` (`Reason` and `Displayed` unchanged), and S10-34 added `ThumbnailCache.cachedImage` and a `CachedThumbnail` call in `CitationComposerSubjectGroups.swift`, which D-13's defaulted parameters leave compiling
**Review:** [`S10-35-person-gallery.review.md`](S10-35-person-gallery.review.md) — Fits, Ready with changes (applied); 2 rounds

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

**Machines.** STEP-1 to STEP-3 and STEP-4's Go tests run on Linux. STEP-4's `./scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go`, and `protoc-gen-swift`. STEP-5 to STEP-9, `scripts/check-pr.sh`'s Xcode stage, and §10 need a Mac. On a machine without them, run every check it can, and report each one it could not run as not run.

## 1. Objective

The researcher sees every face found for a person on the Person page, in a Likenesses section under Details, earliest first; each tile opens the Citation that holds it. The Person page header thumb and the Persons list thumb show the earliest likeness instead of the person mark. A person with no likeness looks as it does today. Events and Places are unchanged.

**Why now.** S10-52 caches every likeness value with its image File, and S10-34 (landed in #442) lets the researcher draw them. Nothing shows them: S10-52's D-9 keeps depiction fields off the Person page until this gallery exists.

**What stays.** The Details rows (there is still no Likeness row: D-9's skip stays), the gender Badge and tint, the Why disclosures, the Events and Places lists and pages, Promote, search hits, and the reconciler itself. Choosing a portrait is S10-39.

## 2. User stories

**US-1.** As a researcher, when I open a Person who has likenesses, I see every one of them under Details, earliest first, so that I can see every face the evidence gives that person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person with likeness A (saved first, Source *Register*) and B (saved second, Source *Census*), both `kept` | `Detail.Likenesses` is [A, B]; the page shows a “Likenesses” section with meta “2” and two 96pt tiles, captions “Register” then “Census” |
| AC-1.2 | B's Source is graded low trust (value `weak`) | B is still listed after A |
| AC-1.3 | B is on a provisional member, or a stronger negative Observation denies B's image | Only A is listed |
| AC-1.4 | The same PNG saved on two Sources (one value, two records) | One tile; its record is the earlier-saved Observation |
| AC-1.5 | A's region is redrawn after B was saved (A gets a new image File) | Order stays [A, B] (D-2) |
| AC-1.6 | Person with no likeness | No Likenesses section; no Likeness row in Details |
| AC-1.7 | A's image File row has `media_type` other than `image/png` (crafted folder) | A is not listed and is not the lead (SEC-1, D-6) |
| AC-1.8 | A tile's Source title is empty | The caption is the Observation ref |

**US-2.** As a researcher, when I click a likeness tile, the composer opens on its Citation with that Observation in focus, the way a Why record opens.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Click (or focus and press Space on) tile A | `navigation.go(to:)` with `ReconciliationRecord.composerLocation(for: A.record)`: section Sources, A's `sourceId`, `citationId`, `artifactId`, `observationId`, surface `.citationComposer` |

**US-3.** As a researcher, the earliest likeness leads the Person page header and the Persons list, so that I recognise people at a glance.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Person with A and B | `PersonHeader.LikenessImageFileID` is A's image; the FFI `PersonHeader.likeness_image_rel_path` on `ListPersonHeaders` and on `GetConclusionDetail` is A's `objects/…` path |
| AC-3.2 | Same Person, page open | Header thumb (120pt) shows A's image filled to the square, spoken “Likeness from Register”; with gender Woman, its outline is `PVColor.lineMaternal` and the fill is not tinted (PDP-9) |
| AC-3.3 | Same Person in the Persons list | The 44pt row thumb shows A's image (PSL-7) |
| AC-3.4 | Person with no likeness | Header thumb is the person mark spoken “No likeness recorded”, with the gender tint as today; the list row shows the person mark with no flash of the empty dash |
| AC-3.5 | Path given, file missing or unreadable | Header and list fall back to the person mark (kit Thumbnail fallback); a tile shows the person mark |
| AC-3.6 | Events and Places lists and pages | Unchanged: their marks, no image |

## 3. Scope

| In | Covered by |
| --- | --- |
| `autoreconcile.Reason.InGallery` (kept or weak) beside `Displayed` | STEP-1, T-1 |
| One Go read of a handle's gallery likenesses, earliest first (`core/database/likenesses`) | STEP-1, T-2 |
| `conclusiondetails.Detail.Likenesses` with each tile's leading record | STEP-2, T-3, T-4 |
| `conclusionheaders.PersonHeader.LikenessImageFileID` (cached with the header) | STEP-3, T-5 |
| FFI: `PersonHeader.likeness_image_rel_path`, `ConclusionDetail.likenesses` | STEP-4, T-6, T-7 |
| Swift store types and GoStore mapping | STEP-5 |
| `PVThumbnail` draws its line on an image; `CachedThumbnail` fallback, image label, colours | STEP-6, T-10 |
| Header thumb through `CachedThumbnail` from `ConclusionDetailThumbnail` (PDP-9) | STEP-7, T-8 |
| Likenesses section (PDP-7, PDP-8) | STEP-8, T-9 |
| Persons list likeness thumb (PSL-7); `PVList` thumbnail view slot | STEP-9, T-11 |
| Working rule: "earliest" is first recorded | STEP-10 |

| Out | Home |
| --- | --- |
| Choosing a portrait, “Use as portrait”, the Portrait Badge, toast, header-thumb menu (PDP-10, PDP-11, PSL-8) | S10-39 |
| Archiving S10-D12 | S10-39 (the brief stays while S10-39 builds on it) |
| Place and event galleries; a likeness on Events or Places lists | Not planned (map row S10-35 Out) |
| A likeness on search hits or Promote's target picker | Not planned; `likeness_image_rel_path` is filled on the two reads named in D-8 only |
| Promote's likeness comparison | S11-20 |
| An image viewer or lightbox | Not planned (S10-D12 §2.1) |
| Moving the handler tests' likeness helpers onto `observationstest` | Not planned in this PR (S10-52 Next adds) |
| Removing D-9's depiction skip from `conclusiondetails.loadProperties` | Not planned: D-1 keeps it |
| An index on `auto_reconciler_values.value_image_file_id` | Not planned: no read filters on it (D-3) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Reasons | `core/autoreconcile/autoreconcile.go` `Reason`, `Reason.Displayed` | `Displayed` is `r == ReasonKept`. No rule names "kept or weak" |
| Cache | `core/database/migrations/000053.sql`; `core/database/autoreconciler/autoreconciler.go` `sqlInsert` | `auto_reconciler_values.value_image_file_id` holds each depiction value's image File id; `auto_reconciler_outcomes` maps each Observation to its value's `value_rank` (`000040.sql`, index on `property_id` only) |
| Graph cache | `core/database/graphcache/graph.go` `Node.Values`, `Value.ImageFileID` | Node values carry rank, reason, image id; no member Observation ids |
| PNG rule | `core/database/depictionvalues/depictionvalues.go` `Readable`, `FromJoin` | Literal `"image/png"`: a non-PNG image id is emptied. Package doc: every reader that returns a stored depiction to the app goes through `Readable` |
| Property keys | `core/vocabkey/vocabkey.go` `PropertyName` … `PropertyToponym` | No likeness key; `core/database/subjectvocab/registry.go` seeds `"likeness"` (multiple, locked, Person sort 4) |
| Batched reads | `core/database/sqlbatch/sqlbatch.go` `Query`, `Querier` | `Query(q, head, tail, args, ids, scan)` binds ids past the bind limit; `Querier` is `Query` only, which `*graphcache.Graph`, `*sql.DB`, and `*sql.Tx` all satisfy |
| Detail | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity`, `loadProperties`, `loadOutcomes`, `Outcome`, `Detail` | `loadProperties` skips subject and depiction Properties (S10-52 D-9); `loadOutcomes` drops rows whose Property is not in `byProp`, so likeness records are discarded |
| Headers | `core/database/conclusionheaders/conclusionheaders.go` `PersonHeader`; `compose.go` `composePersons`, `onePerson` | Composes from graph nodes for ids with no stored row (`need`) and stores the row (`g.SetDisplay`); no likeness. `display_register.go` `checkDisplay` recomposes and compares with `reflect.DeepEqual` |
| Header cache upkeep | `core/database/graphcache/commit.go` `OnCommit` | Drops the stored rows of every handle a commit names (`dropDisplays(roots…)`), so any likeness write drops that Person's row |
| Paths | `core/database/files/files.go` `RelPaths` | Many File ids → `objects/…` paths in one query; unknown ids omitted |
| Proto | `api/proto/engine.proto` `PersonHeader` (fields 1–7), `ConclusionDetail` (1–3, oneof 4–6), `ReconcilerOutcomeDetail` | No likeness fields |
| Handlers | `api/ffi/handlers/conclusion_headers.go` `ListPersonHeaders`, `personHeaderProto`; `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`, `conclusionDetailProto`, `attachHeader` | Outcome protos are built inline in `conclusionDetailProto`. `personHeaderProto` is also called by `promote_align.go` and `search_catalog.go` |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonHeader`, `CatalogConclusionDetail`, `CatalogReconcilerOutcome`; `macos/App/Platform/GoStore.swift` `mapPersonHeader`, `getConclusionDetail` | Outcome mapping is inline in `getConclusionDetail` |
| Header thumb | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailThumbnail` (colours only), `ConclusionDetailPage.mark` / `thumbnailLabel`, `ConclusionDetailHeader` | Draws `PVThumbnail(.init(glyph: .mark(mark), label: thumbnailLabel), size: 120, colors:)`; no image |
| Page kinds | `macos/App/Features/Conclusions/PersonDetailView.swift`, `EventDetailView.swift`, `PlaceDetailView.swift`; `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift` | Each view passes `mark:` and `thumbnailLabel:`; Event and Place use the protocol default `thumbnail` |
| Thumb | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `activeColors`, `isFilled` | Colours apply to mark and symbol states only; the image state draws `borderDefault` |
| Loader | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail` | Loads `objects/…` through `ThumbnailCache`; fallback is a `type_*` mark or MIME glyph or empty; cannot fall back to a `subject_*` mark; starts on `.empty` until its task runs |
| List | `macos/App/DesignSystem/Components/List/PVList.swift` `PVList`; `macos/App/Features/Conclusions/ConclusionListPage.swift` `ConclusionListBody`, `ConclusionListKind`; `ConclusionListRow.swift` `thumbnail(mark:)`; `PersonsListView.swift` `PersonsList` | `PVList`'s thumbnail slot is `(Item) -> PVThumbnail.Content`, so a row cannot load an image. One call site outside the kit (`ConclusionListBody`) plus two previews |
| Tile shell | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `previewWell` | Plain `Button` over the thumb, `.pvFocusRing`, `@FocusState`, help, accessibility label |
| Record → composer | `macos/App/Features/Conclusions/ReconciliationRecord.swift` `composerLocation(for:)` | Builds the composer location from a `CatalogReconcilerOutcome` |
| Tests | `core/database/autoreconciler/autoreconciler_test.go` `TestDepictionsReconcileByImage`; `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenessHasNoField`, `TestForEntityQueryCountIsConstant`; `core/database/conclusionheaders/conclusionheaders_test.go` `TestListPersons`, `TestListPersonsQueryCountIsConstant`; `api/ffi/handlers/conclusion_headers_test.go` `TestListPersonHeaders`; `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailCarriesItsHeader`; `macos/ProvenenciaTests/PersonDetailTests.swift` `treatmentRidesOnContentAndRowsStay`; `macos/ProvenenciaTests/PersonsListTests.swift` | `treatmentRidesOnContentAndRowsStay` asserts `event.thumbnail == .init()`, which D-11 makes false |

**Precedent.** The lead-value-on-the-header shape is `PersonHeader.Name` (`keptName` in `compose.go`): composed with the batch, stored with the row, recomposed by `checkDisplay`. The image-path-at-the-edge shape is S10-54's `depictionPaths` in `api/ffi/handlers/observations.go`: collect File ids, call `files.RelPaths` once per response. The tile shell is S10-18's artifact card `previewWell`. Copy those; do not copy `ListPersons`'s post-compose `Updated` fill (that value moves without a commit on the handle; a likeness does not).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The gallery is its own list, `Detail.Likenesses`, beside `Fields`. D-9's skip of depiction Properties in `loadProperties` stays. | Working rule: the Person page has no Likeness field row. A field would hand Swift every value with its reason, and Swift would repeat the admit rule. | Lifting D-9 and filtering in Swift (S10-52's Next adds) — a second copy of the admit rule outside Go. |
| D-2 | **Earliest** is first recorded: a likeness value orders by the smallest Observation id among the outcomes that went into it (`auto_reconciler_outcomes.value_rank = auto_reconciler_values.rank`); ties by rank. Answered by the user 2026-10-17; recorded as a working rule (STEP-10). | Observation ids are UUIDv7, minted at save. A region redraw keeps its Observation, so the face keeps its place. | Image File id — a redraw moves the face to the end, and S10-52 D-3 called it not a display order. Source date — many Sources have none. |
| D-3 | `likenesses.ForEntities(q sqlbatch.Querier, entityIDs [][]byte) (map[string][]Likeness, error)` *(new)* in `core/database/likenesses` *(new)* is the one read. One `sqlbatch.Query` over `auto_reconciler_values` joined to `properties` (key `vocabkey.PropertyLikeness`, origin `vocabkey.OriginProvenencia`), `auto_reconciler_outcomes` on (entity, property, rank), and `files` (LEFT JOIN, for the media type). Go keeps a row when `autoreconcile.Reason(reason).InGallery()` and `depictionvalues.ReadableImage(imageFileID, mediaType)` is non-nil (D-6). The page gallery, the header lead, and S10-39's fallback all call it. | The map row: "earliest" defined once, in one Go read; the admit rule a named method, not a `reason IN (…)` literal. A nested package per `add-catalog-query`. One query per call whatever the number of handles or likenesses. | A function in `conclusionheaders` — `conclusiondetails` would import the header package for a read it does not own. Reading the graph node — it has no member Observation ids. |
| D-4 | `func (r Reason) InGallery() bool { return r == ReasonKept \|\| r == ReasonWeak }` *(new)*, beside `Displayed`, doc: "a value a gallery lists: displayed, or carried only by weak evidence; a weak photo does not contradict another photo (working rule \"A gallery is the reconciled likeness values\")". | Map row S10-35; S10-39 reuses it. | A list of reasons in SQL — the map row forbids it. |
| D-5 | `vocabkey.PropertyLikeness = "likeness"` *(new)* in the seeded Property keys block. `subjectvocab/registry.go` keeps its literal. | Readers look up seeded Properties by a `vocabkey` constant (`PropertyName`, `PropertyGender`). The registry spells every key as a literal today; changing one line of it is out of scope. | A string literal in `likenesses`. |
| D-6 | `func ReadableImage(imageFileID []byte, mediaType string) []byte` *(new)* in `depictionvalues`: returns `imageFileID` when it is non-empty and `mediaType == "image/png"`, else nil; doc: "the image half of Readable: the File id a reader may hand to the app". `Readable` sets `v.ImageFileID = ReadableImage(v.ImageFileID, imageMediaType)`. `likenesses` calls it on each scanned row. The PNG literal stays private to it. | S10-52 SEC-1's home: the gallery drops non-PNG Files. The package doc says every reader that returns a depiction goes through `Readable`; splitting its image step keeps one rule, so a later image check reaches the gallery too (review M-1). | An exported `ImageMediaType` constant compared in `likenesses` — a second copy of the rule. Filtering in the handler after `RelPaths` — the header lead is cached in Go and would still name the File. |
| D-7 | `conclusiondetails.Detail.Likenesses []LikenessTile` *(new)*; `type LikenessTile struct { ImageFileID []byte; Record Outcome }` *(new)*: `Record` is the outcome of the value's earliest Observation (`likenesses.Likeness.ObservationIDs[0]`). `ForEntity` calls `likenesses.ForEntities(db, [][]byte{entityID})` before `loadOutcomes` and builds `leads` from each likeness's `ObservationIDs[0]`. `loadOutcomes` keeps attaching field outcomes to `f.Outcomes` as today. After it returns, `ForEntity` builds `Detail.Likenesses` in `ForEntities` order: for each likeness, `LikenessTile{ImageFileID, Record: leadsMap[string(ObservationIDs[0])]}` when the map has it. `loadOutcomes` becomes `func loadOutcomes(q Querier, entityID []byte, byProp map[string]*Field, leads map[string]bool) (map[string]Outcome, error)`. It still scans every row into an `Outcome` before the `byProp` check. In the `!ok` branch, a row whose `string(o.ObservationID)` is in `leads` is stored in the returned map (keyed by that string) instead of being dropped; any other row is dropped as today. `auto_reconciler_outcomes`' primary key (entity, property, observation) gives each Observation one row per handle and Property, so a key is written once. Lead rows are not queued for the date and name lookups, so their `Recorded` has no `Date` or `Name`; `reconcilerOutcomeProto` already maps an empty value to an empty `ConclusionValue` (`conclusionValueProto`). A lead with no outcome row gives no tile. | The tile needs exactly what a Why record needs to open the composer (Source title, Citation, Artifact, Subject, Observation). `sqlOutcomes` already reads it; this reuses its row and its proto. | A second outcome query in `likenesses` — a copy of `sqlOutcomes`' joins. |
| D-8 | `conclusionheaders.PersonHeader.LikenessImageFileID []byte` *(new)*, doc by role: "the image of the likeness the Person leads with on its page and in the Persons list; nil when none. S10-35: the earliest gallery likeness." `onePerson` sets it to `likenesses.ForEntities(...)[id][0].ImageFileID`. This is the one place the lead is chosen; S10-39 changes only this choice. Order inside a write (`core/writes/run.go` `Run`): `autoreconciler.RecomputeReported` rewrites the cache, then `searchindex.Reproject` composes rows over the open transaction with `store=false`, then the commit, then `OnCommit` drops the stored rows. A stored row is only ever composed by a read of the committed catalog graph, so it always sees the recomputed cache. `composePersons` calls `ForEntities(g, need)` once, after `need` is built and when it is non-empty, and passes each id's lead to `onePerson` as a new last parameter `lead []byte`. It is stored with the row. | Same shape as `Name`. `OnCommit` drops the row of every handle a write names, and every likeness write names its handle, so the stored lead cannot go stale. `checkDisplay` recomposes it. One query per list read of uncached rows. | Filling it after compose like `Updated` — a query on every list read, for a value the cache already keeps. |
| D-9 | Proto: `PersonHeader.likeness_image_rel_path = 8` *(new)*, filled by `ListPersonHeaders` and `GetConclusionDetail` only (doc says so, like `updated_at`). `message LikenessDetail { string image_rel_path = 1; ReconcilerOutcomeDetail record = 2; }` *(new)*; `ConclusionDetail.likenesses = 7` *(new)*. | Paths are derived at the edge (S10-54 D-4). The tile reuses the outcome message. Search and Promote do not show the thumb. | A path on every `personHeaderProto` caller — two readers that draw no thumb would resolve paths. |
| D-10 | Handlers: `ListPersonHeaders` collects every header's non-empty `LikenessImageFileID`, calls `files.RelPaths` once, and sets `ph.LikenessImageRelPath = paths[string(h.LikenessImageFileID)]` after `personHeaderProto(h)`. `attachHeader` returns `([]byte, error)`: the Person header's `LikenessImageFileID` (nil for other kinds). `attachLikenesses(c *database.Catalog, d conclusiondetails.Detail, lead []byte, out *engine.ConclusionDetail) error` *(new)* calls `files.RelPaths` once over `lead` and every `d.Likenesses[i].ImageFileID`, sets `out.GetPerson().LikenessImageRelPath` when the header is a Person, and appends one `LikenessDetail` per likeness with its path and `reconcilerOutcomeProto(l.Record)`. `reconcilerOutcomeProto(o conclusiondetails.Outcome) *engine.ReconcilerOutcomeDetail` *(new)* is the body of today's inline outcome loop, moved; `conclusionDetailProto` calls it. A likeness whose path `RelPaths` omits is still listed, with an empty path. | One `RelPaths` per response. An omitted path shows the person mark (AC-3.5), the rule S10-54 D-4 follows. | Dropping a likeness with no path — it would hide evidence the cache holds. |
| D-11 | `ConclusionDetailThumbnail` becomes `struct ConclusionDetailThumbnail: Equatable { var mark: PVMarkKey; var label: LocalizedStringResource; var imageRelPath: String = ""; var imageLabel: String? = nil; var colors: PVThumbnail.Colors? = nil }`. `ConclusionDetailPage`, `ConclusionDetailLoaded`, and `ConclusionDetailHeader` drop their `mark` and `thumbnailLabel` properties (and the page its init arguments); `ConclusionDetailBody`'s extension drops the `thumbnail` default, so Event and Place each implement `thumbnail`: Event `.init(mark: .subjectEvent, label: L10n.Conclusions.eventNoImage)`, Place `.init(mark: .subjectPlace, label: L10n.Conclusions.placeNoImage)`. Person: `PersonDetailContent` stores `var likenessImageRelPath: String`, set in `init` from `detail.personHeader?.likenessImageRelPath ?? ""` (stored like `treatment`); its computed `thumbnail` is `.init(mark: .subjectPerson, label: L10n.Conclusions.personNoLikeness, imageRelPath: likenessImageRelPath, imageLabel: leadLabel, colors: treatment?.thumbnailColors)`. `leadLabel` (computed): the tile in `likenesses` whose `imageRelPath` equals a non-empty `likenessImageRelPath` gives `L10n.Conclusions.likenessLabel(tile.sourceTitle)` (board 2j: “Likeness from {title}”); no match gives `L10n.string(L10n.Conclusions.personLikenessThumb)`. The lead is always a gallery value and values are keyed by image, so at most one tile matches. | S10-23 D-7 and Next adds: one slot, each kind supplies its whole thumb, and a likeness is not spoken as "No likeness recorded". | Keeping `mark` on the page init — two places own the thumb. |
| D-12 | `ConclusionDetailHeader` draws `CachedThumbnail(projectDir:, relPath: content.thumbnail.imageRelPath, size: 120, fallbackMark: content.thumbnail.mark, fallbackLabel: content.thumbnail.label, imageLabel: content.thumbnail.imageLabel, colors: content.thumbnail.colors)`. `ConclusionDetailPage` passes `session.projectKey.projectDir` to `ConclusionDetailLoaded` (new `projectDir: String` property), which passes it to the header and the gallery. | `CachedThumbnail` is the one loader for `objects/…` (S10-34 D-13 uses it too). | `NSImage(contentsOf:)` on the render path (forbidden by `CachedThumbnail`'s doc). |
| D-13 | `CachedThumbnail` gains `var fallbackMark: PVMarkKey? = nil`, `var fallbackLabel: LocalizedStringResource? = nil`, `var imageLabel: String? = nil`, `var colors: PVThumbnail.Colors? = nil`. The fallback choice is `static func fallbackContent(fallbackMark: PVMarkKey?, fallbackLabel: LocalizedStringResource?, typeIconKey: String?, mediaType: String, originalFilename: String) -> PVThumbnail.Content` *(new)*: `fallbackMark` first (`PVThumbnail.Content(glyph: .mark(fallbackMark), label: fallbackLabel)`), then today's type icon, MIME glyph, empty; the private instance `fallbackContent()` calls it. The type doc's preference line becomes "raster → fallback mark → type icon → MIME glyph → empty". While the loaded image shows and `imageLabel` is set, the `.tile` layout adds `.accessibilityLabel(Text(verbatim: imageLabel))` after `PVThumbnail` (it replaces the thumb's own label; the label is a formatted `String`, not a resource). The `.tile` layout passes `colors` to `PVThumbnail`. The `.tile` layout draws `fallbackContent()` directly while the trimmed `relPath` is empty, instead of the `@State` content, so a row with no path never shows the empty dash first. A row with a path shows the kit's state for `.empty` content until its task sets loading, as every `CachedThumbnail` does today; that is accepted and is not changed here. `taskID` is unchanged. | The header and list need a `subject_*` mark fallback, a spoken label on an image, and the tint pass-through S10-23 named. | A second loader view for marks — a near-copy of `CachedThumbnail`. |
| D-14 | `PVThumbnail` replaces `activeColors` and `isFilled` with two computed colours. Fill: mark, symbol → `colors?.tint ?? PVColor.surfaceInset`; image → `PVColor.surfaceInset`; empty, loading → `PVColor.surfaceSunken`. Line: mark, symbol, image → `colors?.line ?? PVColor.borderDefault`; empty, loading → `PVColor.borderSubtle`. The `Colors` doc says the line also outlines an image. | PDP-9: with an image, the tint is the outline only. Every other state draws exactly as today. | An overlay in the Person page — the thumb owns its border. |
| D-15 | `ConclusionDetailBody` gains `var likenesses: [ConclusionDetailLikeness] { get }`, default `[]` in the extension. `struct ConclusionDetailLikeness: Equatable, Identifiable { var id: String; var imageRelPath: String; var sourceTitle: String; var location: WorkspaceLocation }` *(new)* in `ConclusionDetailPage.swift`. `PersonDetailContent` stores `var likenesses: [ConclusionDetailLikeness]`, set in `init` from `detail.likenesses` in order: `id` = record Observation id, `sourceTitle` = `ReconciliationRecord.sourceTitle(for: record)`, `location` = `ReconciliationRecord.composerLocation(for: record)`. `static func sourceTitle(for outcome: CatalogReconcilerOutcome) -> String` *(new)* on `ReconciliationRecord`, beside `composerLocation(for:)`, returns `outcome.sourceTitle.isEmpty ? outcome.observationRef : outcome.sourceTitle`; `ReconciliationRecord.init` calls it in place of its inline expression. | A slot below Details that the relationship sections cannot hold (S10-D12 inventory). A record presents its Source one way, in the Why and the gallery (review M-2). | A Person-only view inside `PersonDetailView` — the shared page owns section layout. |
| D-16 | The section is a private `struct ConclusionDetailLikenessSection: View` *(new)* in `ConclusionDetailPage.swift` with `let projectDir: String`, `let likenesses: [ConclusionDetailLikeness]`, `@Environment(WorkspaceNavigation.self) private var navigation`, and `@FocusState private var focusedID: String?`. `ConclusionDetailLoaded` draws it after the Details `VStack` and before `ForEach(content.sections)`, only when `content.likenesses` is non-empty; the gap above it is the page column's `PVSpacing.space9`, as for every section. Its body is `VStack(alignment: .leading, spacing: 0) { PVSectionHeader(…); grid }` with `.accessibilityElement(children: .contain)` and the label on that `VStack`, the shape of `ConclusionDetailRelationshipSection`. Layout per §6 macOS. | Snowflake on the one page that draws it (S10-D12 inventory: Tile button and caption are Snowflake / Ship). | A new kit card component — the brief forbids it. |
| D-17 | `PVList` gains a `Thumbnail: View` generic parameter: `struct PVList<Item: Identifiable, Thumbnail: View, Primary: View, Secondary: View>`, stored `thumbnail: ((Item) -> Thumbnail)?`. The main init takes `@ViewBuilder thumbnail: @escaping (Item) -> Thumbnail` in place of `thumbnail: ((Item) -> PVThumbnail.Content)?`; the row draws `thumbnail(item).frame(width: thumbnailSize, height: thumbnailSize)`. `extension PVList where Thumbnail == EmptyView` adds the same init without `thumbnail` (stores nil). Today's `where Secondary == EmptyView` extension becomes `where Thumbnail == EmptyView, Secondary == EmptyView` and drops its `thumbnail` parameter. The first preview passes `{ _ in PVThumbnail(PVThumbnail.Content(glyph: .mark(.subjectPerson))) }`. | The slot cannot load an image today. One feature call site. The list still frames the slot. | An async closure in `PVThumbnail.Content` — a second loader. Session-side image state — reactivity for one row cell. |
| D-18 | `ConclusionListKind` gains `static func thumbnailRelPath(_ row: Row) -> String` (named by role, like `ConclusionDetailThumbnail.imageRelPath`; review M-3), default `""`; `PersonsList` returns `header.likenessImageRelPath`. `ConclusionListBody` passes `thumbnail: { CachedThumbnail(projectDir: session.projectKey.projectDir, relPath: Kind.thumbnailRelPath($0), fallbackMark: Kind.mark) }`. `ConclusionListRow.thumbnail(mark:)` stays for `OmnibarHitPresentation`; the `ConclusionListRow` doc line "**thumbnail:** always reserved; the kind's `subject_*` mark, never a photo." becomes "**thumbnail:** always reserved; the kind's `subject_*` mark, or the likeness a Person leads with." | PSL-7; Events and Places keep the default and so the mark (AC-3.6). | A Persons-only list body. |
| D-19 | No `FakeStore` change. Swift tests build `CatalogConclusionDetail` and `CatalogPersonHeader` values directly (seeded details through `seededConclusionDetails` where a store is needed). | The tests check the Swift mapping from store types; Go owns the read. | Clustering likenesses in `FakeStore` — a third copy of the rule. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/likenesses/likenesses.go` *(new)* `Likeness`, `ForEntities` | package / func | `conclusiondetails.loadOutcomes`; `autoreconciler` loader | The only read that joins values to outcomes by rank; both callers need it (D-3) |
| `Reason.InGallery` *(new)* | func | `Reason.Displayed` | Different rule (kept or weak); beside it |
| `vocabkey.PropertyLikeness` *(new)* | const | `vocabkey.PropertyGender` | Extends that block |
| `depictionvalues.ReadableImage` *(new)* | func | `depictionvalues.Readable` | The image step of `Readable`, split out so both call it (D-6) |
| `conclusiondetails.LikenessTile` *(new)*, `Detail.Likenesses` *(new)* | type / field | `Outcome`; `likenesses.Likeness` | Wraps an `Outcome` with its image; named by role so it does not read as a second `Likeness` (review M-7) |
| `PersonHeader.LikenessImageFileID` *(new)* | field | `PersonHeader.Name` | Same lead-value shape |
| `LikenessDetail` *(new)*, `likeness_image_rel_path` *(new)*, `likenesses` *(new)* | proto | `ReconcilerOutcomeDetail`, `DepictionValue.image_rel_path` | Carries the outcome message; same path field shape |
| `attachLikenesses` *(new)*, `reconcilerOutcomeProto` *(new)* | func | `depictionPaths`; inline outcome loop | `reconcilerOutcomeProto` is the inline loop moved, not a copy |
| `CatalogLikeness` *(new)*, `CatalogPersonHeader.likenessImageRelPath` *(new)*, `CatalogConclusionDetail.likenesses` *(new)*, `GoStore.mapReconcilerOutcome` *(new)* | type / field / func | `CatalogReconcilerOutcome`; inline outcome map | `mapReconcilerOutcome` is the inline map moved |
| `ConclusionDetailLikeness` *(new)*, `ConclusionDetailBody.likenesses` *(new)*, `ConclusionDetailLikenessSection` *(new)* | type / view | `ConclusionDetailSection`, `ConclusionDetailRelationshipSection` | Those hold relationship rows; a tile grid is a different layout (S10-D12 inventory) |
| `ReconciliationRecord.sourceTitle(for:)` *(new)* | func | `ReconciliationRecord.init` inline expression | The inline expression moved, called by both (D-15) |
| `CachedThumbnail.fallbackMark` / `fallbackLabel` / `imageLabel` / `colors` *(new)* | param | `CachedThumbnail.typeIconKey` | Extends the one loader. `typeIconKey` stays separate: it validates a catalog key to the `.type` family, while the fallback mark is any `PVMarkKey` the caller already holds |
| `PVList` `Thumbnail` generic *(new)* | component API | `PVList.thumbnail` | Extends the slot |
| `ConclusionListKind.thumbnailRelPath` *(new)* | func | `ConclusionListKind.updatedAt` | Same per-row accessor shape |
| `L10n.Conclusions.likenessesSection`, `personLikenessThumb`, `likenessLabel(_:)`, `likenessTileLabel(_:)`, `likenessOpenTip` *(new)* | key | `L10n.Conclusions.personNoLikeness` | New copy |

Searched for: "gallery", "earliest", "likeness" in `core/` and `macos/App`; `RelPaths` callers; `value_rank` readers; `PVList(` and `CachedThumbnail(` call sites.

**Builds on.** `add-catalog-query` (nested package), `sqlbatch.Query`, the header display cache and `checkDisplay`, `files.RelPaths` (S10-54 D-4), `CachedThumbnail` + `ThumbnailCache`, `ReconciliationRecord.composerLocation`, the S10-23 thumb slot.
**Deletes.** `PVThumbnail.activeColors` and `isFilled` (D-14); the inline outcome loop in `conclusionDetailProto` and the inline outcome map in `GoStore.getConclusionDetail` (moved to helpers); `ConclusionDetailPage.mark` / `thumbnailLabel` and the three views' `mark:` / `thumbnailLabel:` arguments (D-11); the `ConclusionDetailBody.thumbnail` default; `TestForEntityLikenessHasNoField`'s "before the gallery exists" comment.
**Next adds.** S10-39 changes only the lead choice in `onePerson`: the pointer's likeness when its Observation id is in some `Likeness.ObservationIDs` of `ForEntities`, else `[0]`. The proto, the Swift store, the header thumb, and the list thumb need no change, because both surfaces read `PersonHeader.LikenessImageFileID`. S10-39 adds the “Use as portrait” button and the Portrait Badge to `ConclusionDetailLikenessSection`; after that there are two image tiles with an open button, caption, Badge, and use-as button (this one and `SourcePageArtifactsView`'s artifact card), so S10-39 runs `evaluate-ui-component` on that shell before writing a third (review M-6). A place or event gallery is a `ConclusionDetailBody.likenesses` implementation plus a key in `ForEntities`.

### Schema

None — no catalog change. The read uses `auto_reconciler_values`, `auto_reconciler_outcomes` (primary key leads with `entity_id`), `properties`, and `files` as they are.

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` | `func (r Reason) InGallery() bool` *(new)* (D-4) |
| `core/vocabkey` | `PropertyLikeness = "likeness"` *(new)* (D-5) |
| `core/database/depictionvalues` | `ReadableImage` *(new)*; `Readable` calls it (D-6) |
| `core/database/likenesses` *(new)* | See sketch |
| `core/database/conclusiondetails` | `LikenessTile` *(new)*, `Detail.Likenesses` *(new)*, `loadOutcomes` signature (D-7); package doc gains one sentence: "Likenesses are listed apart from the fields, earliest first (likenesses.ForEntities)." |
| `core/database/conclusionheaders` | `PersonHeader.LikenessImageFileID` *(new)*; `composePersons`, `onePerson` (D-8) |

```go
// Package likenesses reads the likenesses a handle's gallery shows: its
// cached likeness values whose reason a gallery lists
// (autoreconcile.Reason.InGallery), earliest recorded first.
package likenesses

// Likeness is one likeness value a gallery shows.
type Likeness struct {
	Rank           int      // the value's rank in auto_reconciler_values
	ImageFileID    []byte
	ObservationIDs [][]byte // the records that went into it, ascending; [0] is the first recorded
}

const sqlGallery = `SELECT v.entity_id, v.rank, v.reason, v.value_image_file_id,
		COALESCE(f.media_type, ''), ao.observation_id
	FROM auto_reconciler_values v
	JOIN properties p ON p.id = v.property_id
	JOIN auto_reconciler_outcomes ao ON ao.entity_id = v.entity_id
		AND ao.property_id = v.property_id AND ao.value_rank = v.rank
	LEFT JOIN files f ON f.id = v.value_image_file_id
	WHERE p.key = ? AND p.origin = ? AND v.entity_id IN (`

// ForEntities returns each handle's gallery, keyed by string(entity id),
// earliest first: by ObservationIDs[0], then Rank. A handle with none is absent.
func ForEntities(q sqlbatch.Querier, entityIDs [][]byte) (map[string][]Likeness, error)
```

Tail `)`; args `vocabkey.PropertyLikeness, vocabkey.OriginProvenencia`; ids through `database.UniqueBlobIDs`. Empty ids → `nil, nil` with no query. Errors: only the query's.

### FFI

`api/proto/engine.proto` per D-9:

```proto
// PersonHeader
  // likeness_image_rel_path is the objects/… path of the likeness the Person
  // leads with on its page and in the Persons list (S10-35: the earliest
  // gallery likeness); empty when none. Filled on ListPersonHeaders and
  // GetConclusionDetail only.
  string likeness_image_rel_path = 8;

// LikenessDetail is one likeness a Person page's gallery shows (S10-35),
// earliest recorded first. record is its earliest Observation, as the Why
// shows it; the tile opens that record's Citation. image_rel_path is empty
// when the File's path cannot be derived.
message LikenessDetail {
  string image_rel_path = 1;
  ReconcilerOutcomeDetail record = 2;
}

// ConclusionDetail
  repeated LikenessDetail likenesses = 7;
```

Handlers per D-10, in `api/ffi/handlers/conclusion_headers.go` and `conclusion_detail.go`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Header thumb | `CachedThumbnail` → `PVThumbnail` | `ConclusionDetailPage.swift` `ConclusionDetailHeader` | D-12. 120pt, `PVRadius.sm`. Image: filled to the square, line = tint's line (D-14). No image: the kind's mark with tint and line as today |
| Likenesses header | `PVSectionHeader(title: L10n.Conclusions.likenessesSection, meta: String(content.likenesses.count))` | `ConclusionDetailLikenessSection` | PDP-7. The brief's "aside" is the kit header's mono `meta` slot (the slot the Persons list's count uses) |
| Tile grid | `LazyVGrid(columns: [GridItem(.adaptive(minimum: 120, maximum: 120), spacing: PVSpacing.space6, alignment: .top)], alignment: .leading, spacing: PVSpacing.space7)` | same | Board 2j: 120pt columns, 16pt between columns, 20pt between rows; wraps. `.padding(.top, PVSpacing.space6)` under the header (board: 16pt) |
| Tile | `CachedThumbnail(projectDir:, relPath: tile.imageRelPath, size: 96, fallbackMark: .subjectPerson)` with `.accessibilityHidden(true)` | same | PDP-7, PDP-9 states: loading bar, image, mark fallback |
| Tile button | `.overlay { Button { navigation.go(to: tile.location) } label: { Color.clear.contentShape(Rectangle()) } }`, `.buttonStyle(.plain)`, `@FocusState private var focusedID: String?`, `.focused($focusedID, equals: tile.id)`, `.pvFocusRing(focusedID == tile.id, cornerRadius: PVRadius.sm)`, `.help(Text(L10n.Conclusions.likenessOpenTip))`, `.accessibilityLabel(L10n.Conclusions.likenessTileLabel(tile.sourceTitle))`, `.accessibilityIdentifier("persons.detail.likeness.\(tile.id)")` | same | PDP-8; shell of `SourcePageArtifactsView.previewWell`. `@Environment(WorkspaceNavigation.self)` like `ReconciliationReasoningView` |
| Caption | `Text(verbatim: tile.sourceTitle)`, `PVFont.body(size: PVTypeScale.caption)`, `PVColor.textMuted`, `.lineLimit(1)`, `.truncationMode(.tail)`, `.frame(maxWidth: 120, alignment: .leading)` (the 120pt column; board 2j), `.help(Text(verbatim: tile.sourceTitle))`, `.accessibilityHidden(true)` | same | PDP-7; tile and caption in a `VStack(alignment: .leading, spacing: PVSpacing.space3)` framed `.frame(width: 120, alignment: .leading)`; the crop stays 96pt at the column's leading edge |
| List thumb | `CachedThumbnail` in `PVList`'s slot | `ConclusionListPage.swift` `ConclusionListBody` | D-17, D-18; 44pt, `PVRadius.sm` |

Frames in words. **Gallery** (Woman, three likenesses): header thumb shows the first crop filling the 120pt square with a plum (`lineMaternal`) 1pt outline, no plum fill; under Details a “Likenesses” header with mono meta “3”, then three 96pt square crops in one row, left aligned, each at the leading edge of a 120pt column with 16pt between columns, each with its Source title under it in muted caption, one line, truncated within the 120pt column with the full title as help. **Crop states** (2k): a loaded crop, a crop still loading (the kit's skeleton rule), and an unreadable crop that shows the person mark. **Dark mode**: the same with the dark plum line; tiles on `surfaceInset`. **Persons list** (frame 08): a person with a likeness shows the crop in the 44pt thumb, filled, thumb radius; the others keep the person mark.

States: no likeness → no section, mark thumbs; loading → the kit's loading bar per thumb; unreadable → person mark; 50 likenesses → `LazyVGrid` lays out visible tiles, `ThumbnailCache` keeps 100 decoded entries. Accessibility: the section is `.accessibilityElement(children: .contain)` with `.accessibilityLabel(L10n.string(L10n.Conclusions.likenessesSection))`; each tile is one button spoken “Likeness from Register. Open citation”; the header thumb with an image is spoken “Likeness from Register” (its lead's Source title); list rows keep their combined label (the row ignores children).

### Strings

L10n keys *(new)* in `L10n.Conclusions`, with `Localizable.xcstrings` entries:

- `likenessesSection` = `conclusions.detail.likenesses` "Likenesses" — section heading over a Person's likeness crops
- `likenessLabel(_ source: String) -> String` = `conclusions.person.likenessFrom` "Likeness from %@" — accessibility label of the Person page header thumb when it shows a likeness; the argument is that likeness's Source title (board 2j)
- `personLikenessThumb` = `conclusions.person.likeness` "Likeness" — the header thumb's label when no tile matches the lead image
- `likenessTileLabel(_ source: String) -> String` = `conclusions.person.likenessTile` "Likeness from %@. Open citation" — accessibility label of a gallery tile; the argument is the Source title (board 2j)
- `likenessOpenTip` = `conclusions.person.likenessOpen` "Open citation" — help on a gallery tile (board 2j)

Removed: none.

## 7. Steps

**STEP-1 — Engine: the gallery read**
Files: `core/autoreconcile/autoreconcile.go`, `core/autoreconcile/autoreconcile_test.go`, `core/vocabkey/vocabkey.go`, `core/database/depictionvalues/depictionvalues.go`, `core/database/likenesses/likenesses.go` *(new)*, `core/database/likenesses/likenesses_test.go` *(new)*, `core/database/autoreconciler/autoreconciler_test.go`. Skill: `add-catalog-query`.
Do: D-3 to D-6. Add T-1, T-2, and T-2b.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/ ./core/database/likenesses/ ./core/database/autoreconciler/ ./core/database/depictionvalues/` passes.

**STEP-2 — Engine: the detail lists likenesses**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusiondetails/conclusiondetails_test.go`.
Do: D-7. Rename `TestForEntityLikenessHasNoField` to `TestForEntityLikenesses` and extend it per T-3; add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/` passes, including `TestForEntityQueryCountIsConstant`.

**STEP-3 — Engine: the Person header carries the lead**
Files: `core/database/conclusionheaders/conclusionheaders.go`, `core/database/conclusionheaders/compose.go`, `core/database/conclusionheaders/conclusionheaders_test.go`.
Do: D-8. Update the comment in `TestListPersonsQueryCountIsConstant` to name the likeness query. Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/ ./core/database/graphcache/` passes.

**STEP-4 — FFI**
Files: `api/proto/engine.proto`, generated files from `./scripts/generate-proto.sh`, `api/ffi/handlers/conclusion_headers.go`, `api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/conclusion_headers_test.go`, `api/ffi/handlers/conclusion_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-9, D-10. Add T-6, T-7.
Check: `./scripts/generate-proto.sh` leaves no further diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-5 — Swift store types**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`.
Do: `CatalogPersonHeader.likenessImageRelPath: String = ""` *(new)*; `struct CatalogLikeness: Sendable, Equatable, Identifiable { var imageRelPath: String; var record: CatalogReconcilerOutcome; var id: String { record.observationID } }` *(new)*; `CatalogConclusionDetail.likenesses: [CatalogLikeness] = []` *(new)* (after `header`). `mapPersonHeader` sets the path. `private static func mapReconcilerOutcome(_ o: Provenencia_Engine_V1_ReconcilerOutcomeDetail) -> CatalogReconcilerOutcome` *(new)* holds today's inline map; `getConclusionDetail` uses it for fields and maps `resp.likenesses`.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` (written `xcodebuild test …` below) builds and passes.

**STEP-6 — Thumb and loader**
Files: `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift`, `macos/App/Platform/CachedThumbnail.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`. Skill: `add-ui-component`.
Do: D-13, D-14, including the type doc's preference line. Add T-10.
Check: `xcodebuild test …` passes. The image-state line is checked by hand in §10 step 3 (`PVThumbnail` has no test seam for colours).

**STEP-7 — Header thumb**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `PlaceDetailView.swift`, `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-11, D-12; keys `personLikenessThumb`, `likenessLabel`. Update `treatmentRidesOnContentAndRowsStay` per T-8.
Check: `xcodebuild test …` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-8 — Likenesses section**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `PersonDetailContent.swift`, `ReconciliationRecord.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-15, D-16; keys `likenessesSection`, `likenessTileLabel`, `likenessOpenTip`. Add T-9.
Check: `xcodebuild test …` passes.

**STEP-9 — Persons list thumb**
Files: `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/Features/Conclusions/ConclusionListPage.swift`, `ConclusionListRow.swift`, `PersonsListView.swift`, `macos/ProvenenciaTests/PersonsListTests.swift`.
Do: D-17, D-18. Add T-11.
Check: `xcodebuild test …` passes; both `PVList` previews build.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (S10-35 scope entry, slice 8 line), this plan (status, Landed note), `docs/deployment-plan/spike-10/design/README.md` (S10-D12 row notes PDP-7 to PDP-9 shipped).
Do: per the Definition of done. The working rule line D-2 names is added by the planning commit; confirm it is there.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-35-person-gallery.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/autoreconcile_test.go` `TestReasonInGallery` *(new)* | table over every `Reason*`: `kept`, `weak` → true; `folded`, `outvoted`, `denied`, `provisional`, `no_evidence`, `against` → false | D-4 |
| T-2 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestDepictionsReconcileByImage` (extended) | a helper `gallery(w)` calls `likenesses.ForEntities(w.f.c.Graph(), [][]byte{w.h})`. "two images" → 2 entries, `oa`'s image first, each `ObservationIDs` its one record; "the same pixels" → 1 entry with 2 ids ascending; "a stronger negative denies" → denied image absent; "a provisional member" → its image absent; "a low-trust Source is weak" → present; "an unreadable image is no evidence" → absent; new subtest "a redrawn face keeps its place": "two images", then `observationstest.UpdateObservation` gives `oa` a new crop (shade 150) → still 2 entries, `oa` first with its new image | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, D-2, D-3 |
| T-2b | Go | `core/database/likenesses/likenesses_test.go` `TestForEntities` *(new)* | a package-local `fixture` copied from `conclusiondetails_test.go` `newFixture`, `source`, `personOn`, and `promote` (catalog, `users.Upsert`, `subjectvocab.Install`, `sourcecredibilitygrades.Install`, `claimconfidencegrades.Install`, a `book` source type), plus `likenessOn(src, s, shade)` as in T-3. Table, each case on a fresh fixture: no ids → `nil, nil`; a handle with no likeness → absent from the map; likeness A then B on one promoted Subject → [A, B]; then `observationstest.UpdateObservation` gives A a new crop → still [A, B], A with its new image. The denied, provisional, and weak cases live in T-2, where the reconciler's fixture builds those worlds | AC-1.1, AC-1.5, D-2, D-3 |
| T-3 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenesses` (renamed) | helper `func (f *fixture) likenessOn(src sources.Source, s subjects.Subject, shade uint8) observations.Observation` *(new, test-only)*: `observationstest.FileArtifact` on `src`, then `observationstest.SaveCitation` with one depiction input (`LikenessRegion`, `Crop(t, 40, 30, shade)`). Subject `a` on *Register* gets a name and likeness A (shade 100), saved first; subject `b` on *Census* gets likeness B (shade 200), saved second; `h := f.promote(a, nil)`, `f.promote(b, h)` → `Fields` is only `name`; `Likenesses` has 2, A first; A's `Record.SourceTitle` "Register", `Record.CitationID` and `ObservationID` A's; person with a name only → `Likenesses` empty | AC-1.1, AC-1.6, D-1, D-7 |
| T-4 | Go | `conclusiondetails_test.go` `TestForEntityLikenessNeedsAPNG` *(new)* | likeness A saved, then `UPDATE files SET media_type = 'image/jpeg' WHERE id = ?` on A's image (raw SQL, no recompute) → `Likenesses` empty | AC-1.7, D-6 |
| T-5 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderLeadsWithEarliestLikeness` *(new)* | `s, _, h := f.person("Ann Lee")`; then likeness A (shade 100) on `s` through `observationstest.FileArtifact(t, f.c, userID, f.source.ID)` and `observationstest.SaveCitation`, then likeness B (shade 200) the same way: saves in that order make A earliest. After A's save, `f.list()` already names A's image (the row is recomposed after the commit); after B's save → `ListPersons` row `LikenessImageFileID` = A's image; `PersonsByIDs` same; a person with none → nil; after `c.Graph().Verify()` → no error | AC-3.1, D-8 |
| T-6 | Go | `api/ffi/handlers/conclusion_headers_test.go` `TestListPersonHeaders` (extended) | a person from `likenessFixture` with one `saveLikeness`, promoted with `PromoteSubject` → `likeness_image_rel_path` starts with "objects/" and ends ".png"; a person without → "" | AC-3.1, D-10 |
| T-7 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailListsLikenesses` *(new)* | `likenessFixture(t)` (from `citations_test.go`), `saveLikeness` with crop A then crop B (two shades) on that person, then `PromoteSubject` on the person (empty `entity_id`) → `likenesses` length 2, A first, each `image_rel_path` under "objects/", `record.citation_id` and `record.source_title` set; `person.likeness_image_rel_path` equals `likenesses[0].image_rel_path` | AC-1.1, AC-3.1, D-9, D-10 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `treatmentRidesOnContentAndRowsStay` (updated) and `headerThumbIsTheLeadLikeness` *(new)* | updated: `plain.thumbnail == .init(mark: .subjectPerson, label: L10n.Conclusions.personNoLikeness, imageLabel: "Likeness")`, event `.init(mark: .subjectEvent, label: L10n.Conclusions.eventNoImage)`, place likewise. New: header with `likenessImageRelPath "objects/ab/cd.png"`, a likeness tile with that path on Source *Register*, and gender Woman → `thumbnail.imageRelPath` that path, `imageLabel == "Likeness from Register"`, `colors` non-nil, `mark == .subjectPerson` | AC-3.2, AC-3.4, D-11 |
| T-9 | Swift | `PersonDetailTests.swift` `likenessesBecomeTilesEarliestFirst` *(new)* | detail with likenesses [A (*Register*, path a), B (source title "", ref "OBS-2")] → `content.likenesses.map(\.id)` [A, B]; captions ["Register", "OBS-2"]; A's `location.citationId` and `observationId` A's, `sourceSurface == .citationComposer`; detail with none → `[]`; `EventDetailContent(...).likenesses` → `[]` | AC-1.1, AC-1.8, AC-2.1, D-15 |
| T-10 | Swift | `macos/ProvenenciaTests/ThumbnailCacheTests.swift` `cachedThumbnailFallsBackToTheGivenMark` *(new)* | `CachedThumbnail.fallbackContent(fallbackMark: .subjectPerson, fallbackLabel: nil, typeIconKey: "type_book", mediaType: "", originalFilename: "")` → glyph `.mark(.subjectPerson)`; `fallbackMark: nil` with the same type key → the type mark (today's rule); `fallbackMark: nil`, no key, no metadata → empty | AC-3.4, AC-3.5, D-13 |
| T-11 | Swift | `macos/ProvenenciaTests/PersonsListTests.swift` `listThumbIsTheLikenessPath` *(new)* | `PersonsList.thumbnailRelPath(header with path "objects/ab/cd.png")` → that path; header without → ""; `EventsList.thumbnailRelPath(any)` → "" | AC-3.3, AC-3.6, D-18 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `graphcache` `Verify` callers (the header now holds the lead), `TestListPersonsQueryCountIsConstant` and `TestForEntityQueryCountIsConstant` (one more query, still constant), `PersonDetailTests`, `PVList` call sites.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — S10-D12 stays in design/ for S10-39; README row notes PDP-7 to PDP-9 shipped
- [ ] Domain doc updated if a rule in docs/*.md changed: none (the working rule is in deployment-plan.md)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Run on a Mac, on a test project (S10-34's composer draws the faces).

1. Open a Source with a group photo Artifact. In the composer, add a Person Subject *Ann*, draw her face, save a likeness. On a second Source, draw another face for *Ann*'s other member and save. Promote both onto one Person.
2. Open the Persons list: *Ann*'s thumb is the first face you saved. Another person with no likeness keeps the person mark, with no dash flashing first. Events and Places lists show their marks.
3. Open *Ann* (give her gender Woman first): the header thumb is the first face with a plum `lineMaternal` outline and no plum fill behind it (AC-3.2). Under Details, “Likenesses 2” and two tiles, first face first, each captioned with its Source title. Details has no Likeness row.
4. Tab to the second tile; the focus ring shows. Press Space: the composer opens on that Citation with the likeness in focus. Back returns to *Ann*.
5. Redraw the first face (Edit, redraw, save). Back on *Ann*, it is still first.
6. VoiceOver: the header thumb reads “Likeness from {Source}”; a tile reads “Likeness from {Source}. Open citation”. Hovering a tile shows “Open citation”; hovering a caption shows the full Source title.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The stored header goes stale when a likeness changes without the handle in the commit | `OnCommit` drops rows of every handle a commit names; T-5 runs `Verify` |
| `PVList`'s generic change breaks type inference at the call site | One feature call site and two previews, each named in D-17 |
| A large gallery decodes many images | `LazyVGrid`; `ThumbnailCache` caps decoded entries at 100 and decodes off the main actor |

**Stop and report if:**

- `auto_reconciler_outcomes.value_rank` does not equal the `auto_reconciler_values.rank` of the value its Observation went into (T-2 "two images" finds no outcome rows for a likeness): D-2 and D-3 rely on it.
- A `kept` or `weak` likeness value has no outcome row, so `ForEntities` cannot place it, or a lead Observation has no outcome row in `sqlOutcomes` (the header would then lead with a likeness the gallery omits).
- `checkDisplay` or `Verify` fails because the stored `PersonHeader` differs from a fresh one with the likeness lead.
- `mark` or `thumbnailLabel` of `ConclusionDetailPage`, `ConclusionDetailLoaded`, or `ConclusionDetailHeader` is read anywhere besides those three structs and the three detail views (D-11 assumes not).
- `PVList` has a call site besides `ConclusionListBody` and its two previews.

## 12. Open questions

None.

## 13. PR

**Title:** S10-35 — Person gallery and Persons list thumb
**Body:** objective in one sentence; what shipped against US-1 to US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-35-person-gallery.md`.
