# S13-31 — Header data becomes summaries

**Status:** Landed in #660 (2026-11-01)
**Spike / slice:** Spike 13 · Slice 3 — Family (the Family and Timeline reads, S13-04 and S13-05, are planned on the new names)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: code names only; every view looks and reads the same)
**Board:** —
**Depends on:** S13-02 ([#659](https://github.com/mendahu/provenencia/pull/659))
**Blocks:** S13-04, S13-05
**Branch:** `s13-31-header-data-becomes-summaries`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md) (moving a file in `project.pbxproj`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rule "One person header for every tile" (renamed here); [`macos-client-patterns.md`](../../../macos-client-patterns.md); [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md)
**Planned:** 2026-11-01 against `s13-02-gender-in-the-person-header` @ `3ee06273` (#659's head; branch from `main` after #659 merges)
**Review:** [`S13-31-header-data-becomes-summaries.review.md`](S13-31-header-data-becomes-summaries.review.md) — 2 rounds; Fits with changes / Ready with changes; every finding Fixed or answered by the user (2026-11-01)

**Landed:** the code differs from the plan in six places. (1) Commits: the Go core, the wire and the Swift side are three commits, not one per step, because renamed proto types cannot build in between; the docs are a fourth. (2) `headerHits` in the Evidence graph cards names a frame value (layout), so D-3's own rule keeps it; the plan's list named it by mistake. (3) `conclusionsummaries.NewSummaries` is added, so `ByIDs`, `RowsForSearch` and the searchindex test build one `Summaries` the same way; `handleDocument` and `matchText` take the raw id beside the head. (4) `EventDetailTests`: the three D-13 tests keep their expected strings, and the span test now expects the span in the band (`secondaryDate == rows[0].lead`) instead of the base value, which is D-13's stated change. (5) The check-plan-docs path check fails on a not-yet-Landed plan whose Current state names the moved paths, so the plan is marked Landed in the same PR. (6) Round 1: the Event Date row follows the band (D-13): the summary's choice, not the fields' displayed values, picks the row: a summary `date` is the `date` field, else a summary start or end gives the span row, else a `date` with records; the unused `PersonsFromQuerier`, `EventsFromQuerier` and `PlacesFromQuerier` are deleted and `RowsForSearch` skips a kind with no ids (so `Want.Empty` is not added).


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

"Header" names two things today: the band at the top of a Conclusion page (`ConclusionDetailHeader`) and the data type that holds one person's, event's or place's core facts for every row and tile (`PersonHeader`, `EventHeader`, `PlaceHeader`). The data type becomes a **summary** in Go, proto and Swift, so "header" means only the band (user, 2026-11-01). The band's line under the title, which the page code calls `summary` today, becomes `secondary`, as the list pages already call it (user, 2026-11-01). After this PR: **header** is the band, **secondary** is the line under a title, **summary** is the data. The Event band's date-and-place line is built by the same helper as the Events list's, from the summary (user, 2026-11-01). That is the one change a researcher can see, in one edge case (D-13); everything else looks the same.

**Why now.** S13-04 (family read) and S13-05 (timeline read) return these types to the Family and Timeline tiles, where a tile is not a header at all. Renaming before they are planned means their plans, code and tests use the new names from the start (user, 2026-11-01: right after S13-02).

**What stays.** Every field, value, rule, cache and query, except D-13's band line. The band's own names (D-2). The graph cache's display slot (`Display`, `SetDisplay`, `dropDisplays`, `DepsFunc`). Landed plans, review logs, ship logs, audits and `completed.md`, which keep the names they shipped with.

## 2. User stories

**US-1.** As a contributor, when I read code, a plan or a doc, "header" means the band at the top of a Conclusion page, "secondary" the line under a title, and "summary" the facts a row or tile shows.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The **old-name sweep** below runs at the PR head | it prints nothing |
| AC-1.2 | The **word sweep** below runs at the PR head | every line it prints is on its keep list |
| AC-1.3 | `PersonSummary`, `EventSummary`, `PlaceSummary`, `ChainedPlace` in `core/database/conclusionsummaries` | hold exactly the fields `PersonHeader`, `EventHeader`, `PlaceHeader`, `HeaderPlace` hold at the planning base; every test that passed at the base passes, renamed only |
| AC-1.4 | The app reads the Persons, Events and Places lists, a Conclusion page, search hits and a Promote proposal | each shows what it showed at the base (Swift tests below, renamed only; manual check §10) |
| AC-1.5 | The **map sweep** below runs at the PR head | it prints only lines of the S13-02 and S13-31 PR-map rows and scope entries |
| AC-1.6 | An Event whose summary has date 14 May 1817; and an Event whose `date` field has records but no displayed value, with start 1849 and end 1851 kept | the band shows "14 May 1817" and the list row shows "14 May 1817"; the band shows the 1849–1851 span, as the Events list row does (at the base the band showed no date) |

**Old-name sweep** (AC-1.1, T-4):

```bash
git grep -nP '(Person|Event|Place)Headers?|HeaderPlace|ConclusionHeader|conclusionheaders|conclusion_headers|List(Person|Event|Place)Headers|LIST_(PERSON|EVENT|PLACE)_HEADERS|HeaderDeps|HeaderDependents|header_handles|headerHandles|HeaderHandles|list_(person|event|place)_headers|EventDetailSummary|PersonHeaderDisplay' -- core api macos/App macos/ProvenenciaTests .agents .cursor ':(glob)docs/*.md' docs/ideas ':!*Generated/*' ':!*.pb.go'
```

It does not match `METHOD_GET_*_HEADER` (D-4) or the band's names (D-2). Generated code follows from `./scripts/generate-proto.sh`. The sweeps use `-P`: `git grep -E` here has no `\b`. `PersonHeaderDisplay` also catches `PersonHeaderDisplay.TitleSource` (D-11).

**Word sweep** (AC-1.2, T-5) over the files this PR changes:

```bash
git diff --name-only --diff-filter=AMR 3ee06273 -- core api macos/App macos/ProvenenciaTests ':!*Generated/*' ':!*.pb.go' | xargs git grep -nP '[Hh]eaders?\b|[Hh]eader[A-Z]\w*|\w+Headers?\b' --
```

Keep list: the band and layout names in D-2; `METHOD_GET_*_HEADER` and its `reserved` line (D-4); comments that describe the band. Any other line is a miss.

**Map sweep** (AC-1.5, T-6):

```bash
grep -nE 'PersonHeader|EventHeader|PlaceHeader|HeaderPlace|conclusionheaders|person header' docs/deployment-plan/spike-13/deployment-plan.md
```

## 3. Scope

| In | Covered by |
| --- | --- |
| Go: package `conclusionheaders` → `conclusionsummaries`, its types, functions, locals and comments, and every importer in `core/` and `api/` (D-1, D-3, D-7) | STEP-1, AC-1.3 |
| Go: `effects.Set.HeaderDeps`, `HeaderDepsWalked`; `HeaderDependents` deleted (D-5) | STEP-2 |
| Go: one split by kind (`Want`) and one by-id result (`Summaries`) for `ByIDs`, `RowsForSearch` and their three callers (D-12) | STEP-3 |
| Proto: messages, fields, `Method` values and comments named for the data; the four `oneof header` blocks folded into one `ConclusionSummary` (D-4); regenerated Go and Swift; FFI handlers | STEP-4 |
| Swift: wire mapping onto the regenerated names | STEP-4 |
| Swift: wire types, store methods, display helpers (D-1, D-6) | STEP-5 |
| Swift: the band's `summary` slot → `secondary` (D-10); the Event band line from `EventSecondaryDisplay` (D-13) | STEP-6, AC-1.6 |
| Swift: properties, locals, test names, file moves (D-3, D-6, D-9, D-11) | STEP-7 |
| Docs: living docs, skills, rules, and the map lines in D-8 | STEP-8, AC-1.5 |

| Out | Home |
| --- | --- |
| The band's names and every layout name in `macos/App` (D-2) | not planned |
| The graph cache display slot (`Display`, `SetDisplay`, `displays`) | not planned ("display" names the slot) |
| Any field, behavior, query or cache change | not planned |
| Landed plans, ship and review logs, audits, `completed.md` | not planned (records of what shipped) |
| Proto `reserved "METHOD_GET_*_HEADER"` names | not planned (D-4) |

## 4. Current state

Verified at `3ee06273`. 76 hand-written files carry a data name (`git grep -l` with the old-name sweep's pattern over `core api macos`).

| Area | Where | Today |
| --- | --- | --- |
| Go package | `core/database/conclusionheaders/` (14 files: `compose.go`, `conclusionheaders.go`, `dependents.go`, `display_register.go`, `events.go`, `order.go`, `placechains.go`, `places.go`, and tests `conclusionheaders_test.go`, `display_cache_test.go`, `export_test.go`, `placechains_internal_test.go`, `placechains_test.go`, `walks_test.go`; external tests are `package conclusionheaders_test`) | Types `PersonHeader`, `EventHeader`, `PlaceHeader`, `HeaderPlace` (`conclusionheaders.go:79`), `Headers` (`compose.go:238`, the `ByIDs` result); `headerChainDepth` (`dependents.go:18`); `HeaderDependents` (`dependents.go:190-194`), documented "HeaderDependents is Dependents", body `return Dependents(q, ids)` |
| Go importers | `core/database/autoreconciler/autoreconciler.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/persondetail/claims.go`, `core/database/searchindex/handles.go` (`handleHeaders` 187, `loadHandleHeaders` 193), `core/database/subjects/subjects.go:247` (`HeaderDependents`), `core/database/graphcache/display.go`; `api/ffi/handlers/conclusion_detail.go` (`attachHeader` 202), `conclusion_headers.go` (+ `_test.go`), `promote_align.go` (`addHeaderID` 69, `headers` 119-145), `search_catalog.go` (`fillHitHeaders` 96, `h.Header`), `write_effects.go` (+ `_test.go`) | import `conclusionheaders` |
| Effects | `core/database/effects/registry.go:117-122` `HeaderDeps`, `HeaderDepsWalked`; `core/writes/run.go:88`; `core/database/graphcache/commit.go:109-110`; `api/ffi/handlers/write_effects.go:19`, `write_effects_test.go:44` | handles whose stored display rows read the written handles |
| Proto | `api/proto/engine.proto` | messages `PersonHeader`, `EventHeader`, `PlaceHeader`, `HeaderPlace` (1246), `List{Person,Event,Place}Headers{Request,Response}` (field `headers = 1`); `oneof header { PersonHeader person; EventHeader event; PlaceHeader place; }` in `SearchHit` (1018, numbers 12-14), `ConclusionDetail` (1572, 4-6), `PromoteGraphAlignmentAlternative` (1634, 4-6), `PromoteGraphAlignmentRow` (1657, 13-15); `WriteEffects.header_handles = 6`; `METHOD_LIST_{PERSON,EVENT,PLACE}_HEADERS = 87, 91, 93`; `reserved "METHOD_GET_EVENT_HEADER", …` (44) under the comment "Per-entity headers ride GetConclusionDetail." (41); comments at 1016-1017, 1303, 1355, 1407 |
| FFI dispatch | `api/ffi/dispatch.go` | method cases |
| Go kind splits | `api/ffi/handlers/search_catalog.go` `fillHitHeaders` 102-110, `promote_align.go` `addHeaderID` 69-81, `core/database/searchindex/handles.go` `loadHandleHeaders` 200-207; `conclusionheaders.ByIDs(c, persons, events, places)` (`compose.go:256`), `RowsForSearch(g, personIDs, eventIDs, placeIDs)` (`compose.go:360`); `searchindex.handleHeaders` (187-191) rebuilds by-id maps keyed by `UUIDString` | three hand-written splits of ids by kind and a second by-id type beside `Headers` |
| Go oneof copies | `api/ffi/handlers/search_catalog.go:123-132`, `promote_align.go:119-128`, `136-145`, `conclusion_detail.go:217-239` | four switches filling the oneof |
| Swift wire types | `macos/App/Platform/GenealogyStore.swift` `CatalogHeaderPlace` (406), `CatalogPersonHeader` (460), `CatalogEventHeader` (483), `CatalogPlaceHeader` (544), `CatalogConclusionHeader` (730) and its `person`/`event`/`place` accessors (736-746); `var header: CatalogConclusionHeader?` (244, 320, 340, 700); `listPersonHeaders`, `listEventHeaders`, `listPlaceHeaders` (1372-1374) | |
| Swift mapping | `macos/App/Platform/GoStore.swift` `mapPersonHeader`, `mapEventHeader`, `mapPlaceHeader`, `mapHeaderPlace`, and four bodies producing `CatalogConclusionHeader`: `mapConclusionHeader`, `mapSearchHeader`, `mapPromoteRowHeader`, `mapPromoteAlternativeHeader` (1089-1129); `headerHandles:` (1865); `CoreInvoke.swift`; `FakeStore.swift` `seededEventHeaders`, `seededPlaceHeaders` | |
| Swift display | `macos/App/Features/Conclusions/PersonHeaderDisplay.swift` `enum PersonHeaderDisplay` (title only; `typealias TitleSource = ConclusionTitleSource` at 8, used outside only at `PersonDetailContent.swift:16`); `ConclusionHeaderDisplay.swift` `enum ConclusionHeaderDisplay` (title, secondary, ref over the union); siblings `EventTitleDisplay.swift`, `PlaceTitleDisplay.swift` | |
| Band line | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage<Content, Summary: View, PageSections>` (255), `let summary` (261), init parameters (275, 338), `ConclusionDetailLoaded<Content, Summary, PageSections>` (425, 431, 455, 469), `ConclusionDetailHeaderBand<Content, Summary>` (501), doc comment 245, `ConclusionDetailHeader<Content, Summary>` (687), comment "the kind's summary" (685); `EventDetailView.swift:26` `EventDetailSummary`; `PersonDetailView.swift:19`, `PlaceDetailView.swift:19`, `EventDetailView.swift` pass `summary:`; `EventDetailContent.swift:12,14` `summaryDate`, `summaryPlace`; `EventDetailTests.swift` (4 hits) | the band's line under the title; the list pages call the same slot `Secondary` (`ConclusionListPage.swift:41, 66`) |
| Swift callers | `Features/Conclusions/*`, `Features/Promote/PromoteFlow.swift` (`header:` 380, 564), `PromoteModel.swift:201`, `Features/Catalog/Search/CatalogSearchHitPresentation.swift`, `Features/Workspace/Session/CatalogMutation.swift` (`headerHandles` 101, 116), `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift` | |
| Xcode | `macos/Provenencia.xcodeproj/project.pbxproj` | explicit file references; each file's name appears in its `PBXFileReference` `path`, and in the `/* … */` comments of its `PBXBuildFile`, group child list and Sources phase entries. The Xcode build phase runs `scripts/build-macos-core.sh` (`project.pbxproj:2204`), so `xcode-test` rebuilds the Go library |
| Query keys | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `personsList`, `eventsList`, `placesList` | no "header"; nothing to rename |
| Persistence | `git grep` for the names in `*.json`, golden and `testdata` files; `protojson` in `core api` | none: the FFI is binary proto, and display rows live only in `graphcache` memory |
| Docs | `.agents/skills/add-audit-scope/SKILL.md:80` (`conclusionheaders.ListPersons`), `.cursor/rules/swift-tests.mdc:23` (`seededEventHeaders`), `docs/conclusion-reconciliation.md:285`, `docs/macos-client-patterns.md:105`, `docs/ideas/test-subject-fixtures.md:7`; `docs/deployment-plan/spike-13/deployment-plan.md` (D-8) | |
| Tests | Go: every test in `core/database/conclusionheaders/`, `api/ffi/handlers/conclusion_headers_test.go`, `write_effects_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/searchindex/searchindex_test.go`. Swift: the 15 files under `macos/ProvenenciaTests` the old-name sweep lists, `PersonHeaderDisplayTests.swift` among them | pass at the base |

**Precedent.** S14-01's gallery rename: `8e519655` (`git mv` of the package directory, importers updated) and `067e26a0` (proto names changed, field numbers kept, `./scripts/generate-proto.sh`, Swift mapping renamed to compile). Copy the `git mv` and the keep-the-numbers proto edit. Its behavior changes do not apply here.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Names.** Go and proto: `PersonSummary`, `EventSummary`, `PlaceSummary`, `ChainedPlace` (was `HeaderPlace`), `Summaries` (was `Headers`), `ConclusionSummary` *(new, D-4)*, package `conclusionsummaries`, `summaryChainDepth` (was `headerChainDepth`). Swift: `CatalogPersonSummary`, `CatalogEventSummary`, `CatalogPlaceSummary`, `CatalogChainedPlace`, `CatalogConclusionSummary`, `PersonTitleDisplay` (was `PersonHeaderDisplay`), `ConclusionSummaryDisplay` (was `ConclusionHeaderDisplay`) | User, 2026-11-01: "summary" for the data; `ChainedPlace` uses the existing place-chain vocabulary (`placechains.go`, `PlaceChainDisplay`) and is not confused with `PlaceSummary`. `PersonTitleDisplay` matches `EventTitleDisplay` and `PlaceTitleDisplay`, since it only builds titles | `SummaryPlace` — one word-swap from `PlaceSummary`; `PersonSummaryDisplay` — it formats a title, not a summary |
| D-2 | **What renames:** every type, package, file, function, method, proto message, field, enum value, error-context string, test name and comment that names the data. **What stays:** the band and layout names — `ConclusionDetailHeader`, `ConclusionDetailHeaderBand`, `ConclusionDetailHeaderCompactKey`, `conclusionDetailHeaderCompact`, `PVSectionHeader`, `SectionHeader`, `sectionHeaders`, `headerHeight`, `headerToBodyGap`, `headerToBodySpacing`, `headerPinned`, `headerMeta`, `headerWidth`, `headerRect`, `headerButtons`, `headerAndBody`, `headerActionHitHeight`, `headerContentHeightCited`, `headerContentHeightUncited`, `headerLabel`, `headerBitmap`, `hasHeader`, `showsHeader`, `isHeader`, `panelHeader`, `paintHeader`, `partHeader`, `brandHeader`, `columnHeaders`, `pvWorkspaceHeaderRow`, `SourcePageIdentityHeader`, `EvidenceGraphHeader`, `EvidenceGraphHeaderWidthKey`, `EvidenceGraphLabeledHeaderWidthKey`, `EvidenceCardHeaderActionHits`, `VocabularyHeader`, `VocabularyPanelHeader`, `OnboardingHeader`, `unpromoteConfirmBridgesHeader`; image-format names `JPEGHeader`, `HeaderOnlyPNG`, `ImageHeaderDecode`, `declaredPNGHeaderData`, `writeDeclaredPNGHeader`, `gifHeaderOnly`; `Provenencia_Bridging_Header_h` | Header is the band (user, 2026-11-01) | Renaming every "header" — the band's names are right |
| D-3 | **Locals and properties follow their type.** A variable, constant, parameter, property, closure, proto field or Go field whose type is a summary type, an optional or collection of one, or the union is renamed `header` → `summary`, `headers` → `summaries`, and compound names likewise. Named here so none is a judgment call: `personHeader`, `jamesHeader`, `birthHeader`, `marriageHeader`, `fireHeader`, `lifeHeader`, `eventHeader`, `placeHeader`, `personHeaders` → `…Summary` / `…Summaries`; `attachHeader` → `attachSummary`; `fillHitHeaders` → `fillHitSummaries`; `addHeaderID` → `addSummaryID`; `handleHeaders` → `handleSummaries`; `loadHandleHeaders` → `loadHandleSummaries`; `headerPictureCase(s)` → `summaryPictureCase(s)`; `headerParts` → `summaryParts`; `headerHits` → `summaryHits`; `headerPlace` → `chainedPlace`. A name whose type is a view, size or layout value stays (D-2) | The type decides; the word sweep proves it | Leaving locals as `header` — the next reader sees `header: PersonSummary` |
| D-4 | **Proto.** Names change; every kept field and enum number stays: `List{Person,Event,Place}Summaries{Request,Response}` with `repeated … summaries = 1`; `WriteEffects.summary_handles = 6`; `METHOD_LIST_{PERSON,EVENT,PLACE}_SUMMARIES = 87, 91, 93`. The four `oneof header` blocks fold into one message `ConclusionSummary { oneof kind { PersonSummary person = 1; EventSummary event = 2; PlaceSummary place = 3; } }` *(new)*; each of the four messages gets `ConclusionSummary summary = N` with the lowest number its old oneof used (`SearchHit` 12, `ConclusionDetail` 4, `PromoteGraphAlignmentAlternative` 4, `PromoteGraphAlignmentRow` 13) and reserves the other two numbers and the names `person`, `event`, `place`. Go builds it in one `conclusionSummaryProto(s conclusionsummaries.Summaries, kind string, id []byte) *engine.ConclusionSummary` *(new)*, nil when that id was not read; it holds the one kind switch, and the search and Promote sites call it in one line each. `attachSummary` keeps its read switch (each kind reads differently: derived rows for a Person, relationships for a Place) and only sets `out.Summary`. Swift maps the message in one `mapConclusionSummary` *(new)*; three Go switches and three Swift bodies go. The `reserved "METHOD_GET_*_HEADER"` line stays; its comment at 41 becomes "Per-entity summaries ride GetConclusionDetail." | The FFI is binary proto, the dylib and app ship together, and nothing stores these messages, so the rename and the fold are wire-safe (S14-01 precedent). Eight hand-kept copies of one union become two (review M-6; user, 2026-11-01: each kind's message stays its own as the detail pages diverge, only the wrapper is shared) | Four renamed oneofs — keeps four copies in each language |
| D-5 | `effects.Set.HeaderDeps` → `SummaryDeps`, `HeaderDepsWalked` → `SummaryDepsWalked`; Swift `headerHandles` → `summaryHandles`. `HeaderDependents` is deleted: it forwards to `Dependents`, which `autoreconciler.go:216` and `display_register.go` already call; its callers (`subjects.go:247`, `display_cache_test.go:51,129`, `placechains_test.go:444`, `walks_test.go:303,309,319`) call `conclusionsummaries.Dependents` | One name for one walk | Renaming the alias — keeps a forwarder under a new name |
| D-6 | **Files** move with `git mv`: the package directory; `conclusionheaders.go` → `conclusionsummaries.go`; `conclusionheaders_test.go` → `conclusionsummaries_test.go`; `api/ffi/handlers/conclusion_headers.go` (+ `_test.go`) → `conclusion_summaries.go` (+ `_test.go`); `PersonHeaderDisplay.swift` → `PersonTitleDisplay.swift`; `ConclusionHeaderDisplay.swift` → `ConclusionSummaryDisplay.swift`; `PersonHeaderDisplayTests.swift` → `PersonTitleDisplayTests.swift`. `display_register.go` and `display_cache_test.go` keep their names. In `project.pbxproj`, replace each moved file's name everywhere it appears (the `path =` value and the `/* … */` comments), keeping every object ID | `git mv` keeps history; editing in place keeps target membership | New pbxproj objects — duplicate membership |
| D-7 | Comments that call the data a header say summary, including proto comments (1016-1017, 1303, 1355, 1407) and graphcache's ("A stored summary row is composed, not finished"). Comments about the band stay | Same line as D-2 | — |
| D-8 | **Docs.** Rename in living docs: `docs/*.md`, `docs/ideas/`, `.agents/skills/add-audit-scope/SKILL.md:80`, `.cursor/rules/swift-tests.mdc:23`. In `docs/deployment-plan/spike-13/deployment-plan.md`: the working rules at lines 91 ("One person header for every tile" → "One person summary for every tile", and its body), 96, 97, 98 and 111; the S13-04 and S13-05 PR-map rows and scope entries. Leave: the S13-02 row and scope entry (landed), S13-31's own row and entry (they describe the rename), the slice diagram's PR titles, landed rows in `docs/deployment-plan/spike-14/deployment-plan.md`, plans, logs, audits, `completed.md` | Records say what shipped then; living docs say what is true now | Rewriting history — a ship log's names would no longer be in `git log` |
| D-9 | **Test names.** Go: `TestPersonHeaderTerms`, `TestPersonHeaderTermLabels`, `TestPersonHeaderSkipsNegativeLikeness`, `TestPlaceHeaderParentsAndPeriod`, `TestListPersonHeaders`, `TestListEventHeaders`, `TestListPlaceHeaders`, `TestListHeadersCarryUpdated`, `TestHeadersShowKeptValuesOnly`, `TestHeaderSearchFindsTheListTitle`, `TestHeaderProtosCarryTerms`, `TestHeaderLeadsWithEarliestPicture`, `TestHeaderLeadsWithChosenPicture`, `TestGetConclusionDetailCarriesItsHeader` → Summary; `TestHeaderDependentsIncludeChildPlaces` → `TestDependentsIncludeChildPlaces`. Swift: `PersonHeaderDisplayTests` → `PersonTitleDisplayTests`; `treatmentReadsOnlyTheHeader`, `titleIsTheHeadersChosenTitle`, `withNoHeaderTheTitleIsTheLabelThenTheRef`, `placesListWarmsTheHeaderRead`, `eventsListWarmsTheHeaderRead`, `personsListLoadsHeadersOfPromotedPersons`, `periodAndRelationshipsFillFromTheHeader`, `headerTypeFillsATitleThatOmittedIt`, `headerSuppliesSubjectsAndPlace`, `aHeaderIsTitledByItsKind` → Summary. Stay (they test the band or a list's meta line): `headerThumbIsTheLeadLikeness`, `headerFallsBackNameLabelRef`, `aSinglePictureLeadsTheHeaderWithoutAMenu`, `aPictureFillsTheGalleryAndLeadsTheHeader`, `headerMetaCountsAndShowsRefreshing`, `eventsHeaderMetaNamesTheDateSort`, `anEmphasizedRunInASectionHeaderTitleDrawsHeavier`, `headerActionHitsCenterOnTrailingIcons`, `TestEncodeJPEGOversizedHeaderAllocation` | Read test by test at the planning base | — |
| D-10 | **The band's line is secondary.** `ConclusionDetailPage<Content, Secondary: View, PageSections>`, `let secondary`, `secondary:` parameters; `ConclusionDetailLoaded<Content, Secondary, PageSections>` (425; `let secondary` 431, uses at 455 and 469), `ConclusionDetailHeaderBand<Content, Secondary>`, `ConclusionDetailHeader<Content, Secondary>`; the doc comment at 245 and the comment at 685 "the kind's secondary line"; `EventDetailSummary` → `EventDetailSecondary`; `summaryDate`, `summaryPlace` → `secondaryDate`, `secondaryPlace`; the `summary:` arguments in `PersonDetailView`, `EventDetailView`, `PlaceDetailView`; `EventDetailTests` | User, 2026-11-01: matches the list pages' `Secondary` and `ConclusionHeaderDisplay.secondary`, so "summary" means only the data | Leaving the band's `summary` — two meanings in one struct (`EventDetailContent`) |
| D-11 | `PersonHeaderDisplay.TitleSource` (a typealias of `ConclusionTitleSource`) is deleted; `PersonDetailContent.swift:16` uses `ConclusionTitleSource`, as `EventDetailContent.swift:7` does | One name for the title source type | — |
| D-12 | **One split by kind, one by-id result.** `conclusionsummaries.Want` *(new)* holds person, event and place ids; `func (w *Want) Add(kind string, id []byte)` keeps a 16-byte id of kind `vocabkey.Person`, `Event` or `Place` and ignores the rest (`search.KindPerson` and siblings equal the `vocabkey` values, `core/search/search.go:21-23`). `ByIDs(c *database.Catalog, w Want) (Summaries, error)` and `RowsForSearch(g *graphcache.Graph, w Want) (Summaries, error)` replace their three-slice forms. `fillHitHeaders`, `proposalProto` (`addHeaderID` goes) and `loadHandleHeaders` call `Want.Add`; `searchindex.handleHeaders` is deleted and `matchText` reads `Summaries` by the raw id (the `heads` map key). Each kind is still read in one call, so queries are unchanged | User's standing same-pass rule: this PR renames every copy, so the dedupe rides here (review M-11). Three splits and two by-id types become one each, and S13-04 and S13-05's mixed-kind reads add none | Out row to a follow-up — a second mechanism is answered here or by the user, not deferred (review-plan Floors) |
| D-13 | **One builder for an Event's date-and-place line** (user, 2026-11-01). `EventDetailContent` fills `secondaryDate` and `secondaryPlace` from `EventSecondaryDisplay.content(summary, locale:)`; `EventSecondaryDisplay.content` and `.line` gain `locale: Locale = .autoupdatingCurrent`, passed to `DateRowDisplay.line(date:start:end:locale:)`. `EventDetailContent.summaryDate(_:locale:)` is deleted. The rule is the summary's: the rank-1 kept `date`, else the kept start–end span (`conclusionsummaries` `eventDate` and the event compose that clears the span when a date is kept). The band ignores `extraPlaces`, as it does today. The Vitals Date row keeps its own field-based rule (`dateRow`) | The summary already feeds the Events list, search and Promote; the band joins them, so S13-05's timeline rows have one helper. Edge: a `date` with records and no kept value showed no band date; it now shows the span, as the list does. The Date row still shows that field's records | The page rule in Go — `eventDate` also drives the list's date sort, so changing it moves list order |

**Rules this plan bends.** None. It renames working rule "One person header for every tile" to "One person summary for every tile" at the user's request (2026-11-01); the rule's content is unchanged.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/conclusionsummaries` *(renamed)* | package | `core/database/conclusionheaders` | It is that package, moved |
| `ConclusionSummary` *(new)* | proto message | the four `oneof header` blocks | It replaces them (D-4) |
| `conclusionSummaryProto` *(new)* | Go func | the four switches in §4 | It replaces three; `attachSummary` keeps its reads |
| `conclusionsummaries.Want` *(new)* | Go type | the three splits by kind (§4 Go kind splits) | It replaces them; `ByIDs` and `RowsForSearch` take it |
| `mapConclusionSummary` *(new)* | Swift func | `mapConclusionHeader` and three copies | It replaces them |

Every other name is a rename of an existing type (D-1).

**Builds on.** S14-01's rename shape (`8e519655`, `067e26a0`); `./scripts/generate-proto.sh`; `add-swift-test`'s pbxproj notes.
**Deletes.** `HeaderDependents` (D-5); `EventDetailContent.summaryDate(_:locale:)` (D-13); `PersonHeaderDisplay.TitleSource` (D-11); the four `oneof header` blocks, three of the four Go wrapping switches (`attachSummary` keeps its reads), three of the four Swift union mappers (D-4); `searchindex.handleHeaders` and the three hand-written splits by kind in `fillHitHeaders`, `proposalProto` (`addHeaderID`) and `loadHandleHeaders` (D-12).
**Next adds.** S13-04 (`GetPersonFamily` returns `PersonSummary`s through `conclusionsummaries.ByIDs`), S13-05 (`ListPersonEvents` returns `EventSummary`s with {role, `PersonSummary`}), the S13-12 and S13-13 tiles read `CatalogPersonSummary`. A message that carries one summary of any kind adds one `ConclusionSummary` field.

### Reach

None — no new path into existing code, no new state reachable.

### Scale

None — no write, list or loop changes.

### Edges

None — no rule, comparison or algorithm changes. The one assumption the PR leans on, that nothing persists these names or messages, is checked in §4 (Persistence) and is a stop condition.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionsummaries` | moved from `conclusionheaders`; types renamed (D-1); `HeaderDependents` deleted (D-5); docs and comments say summary (D-7) |
| `core/database/effects`, `core/writes`, `core/database/graphcache`, `core/database/subjects` | `SummaryDeps`, `SummaryDepsWalked`; `subjects` calls `Dependents` (D-5); graphcache comments (D-7) |
| `autoreconciler`, `conclusiondetails`, `persondetail`, `searchindex` | import path, type names, D-3 locals |

### FFI

```proto
// ConclusionSummary is one handle's summary, by kind.
message ConclusionSummary {
  oneof kind {
    PersonSummary person = 1;
    EventSummary event = 2;
    PlaceSummary place = 3;
  }
}

message SearchHit {
  // (other fields unchanged)
  ConclusionSummary summary = 12;
  reserved 13, 14;
  reserved "person", "event", "place";
}
```

The same shape in `ConclusionDetail` (4; reserves 5, 6), `PromoteGraphAlignmentAlternative` (4; reserves 5, 6), `PromoteGraphAlignmentRow` (13; reserves 14, 15). Handlers: `conclusion_summaries.go` with `personSummaryProto`, `eventSummaryProto`, `placeSummaryProto`, `chainedPlaceProto`, `conclusionSummaryProto`, `unmarshalErr("list_person_summaries", …)` and siblings; `conclusion_detail.go` `attachSummary`; `dispatch.go` method cases. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view changes. Renames only (D-1, D-3, D-6, D-10, D-11): `listPersonSummaries`, `listEventSummaries`, `listPlaceSummaries` on `GenealogyStore`, `GoStore`, `FakeStore`; `mapPersonSummary`, `mapEventSummary`, `mapPlaceSummary`, `mapChainedPlace`, `mapConclusionSummary`; `seededEventSummaries`, `seededPlaceSummaries`.

### Strings

None — no user-facing copy names the type or the band's slot.

## 7. Steps

**STEP-1 — Move the Go package**
Files: `core/database/conclusionheaders/` → `core/database/conclusionsummaries/` *(new)*; its importers in `core/` and `api/ffi/handlers/` (§4 Go importers).
Do: `git mv` the directory and the two Go files named in D-6; rename the package clauses, types, `headerChainDepth`, D-3 locals, comments (D-7) and Go test names (D-9) inside the package; in the importers, change only the import path and the type names so they build. `HeaderDependents` stays until STEP-2; handler files keep their names until STEP-4.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/... ./api/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` pass; `git grep -n conclusionheaders -- core api` prints nothing.

**STEP-2 — Effects names and the forwarder**
Files: `core/database/effects/registry.go`, `core/writes/run.go`, `core/database/graphcache/commit.go`, `core/database/subjects/subjects.go`, `core/database/conclusionsummaries/dependents.go` *(new)* (moved in STEP-1) and its tests, `api/ffi/handlers/write_effects.go`, `write_effects_test.go`.
Do: D-5 on the Go side (the proto field stays `header_handles` until STEP-4). Delete `HeaderDependents`; point its callers at `Dependents`; rename `TestHeaderDependentsIncludeChildPlaces`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `git grep -nE 'HeaderDeps|HeaderDependents' -- core api ':!*.pb.go'` prints nothing.

**STEP-3 — One split by kind, one by-id result**
Files: `core/database/conclusionsummaries/compose.go` *(new)* (moved in STEP-1), its tests, `core/database/searchindex/handles.go`, `api/ffi/handlers/search_catalog.go`, `api/ffi/handlers/promote_align.go`.
Do: D-12. Add `Want` and `Want.Add`; change `ByIDs` and `RowsForSearch` to take a `Want` and return `Summaries`; move `fillHitHeaders`, `proposalProto` and `loadHandleHeaders` onto `Want.Add`; delete `searchindex.handleHeaders` and have `handleDocument` and `matchText` read `Summaries` by the raw id. The handlers still fill the old `oneof header` fields; STEP-4 changes the wire.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `git grep -nE 'addHeaderID|handleHeaders|personIDs, eventIDs, placeIDs' -- core api` prints nothing.

**STEP-4 — Proto, FFI and the Swift wire mapping**
Files: `api/proto/engine.proto`, `api/ffi/dispatch.go`, `api/ffi/handlers/*`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`; in Swift, only the lines that name a generated name this step changes (`git grep -nP 'Provenencia_Engine_V1_(Person|Event|Place)Header|Provenencia_Engine_V1_HeaderPlace|Provenencia_Engine_V1_List(Person|Event|Place)Headers|list(Person|Event|Place)Headers\b|OneOf_Header|\.header\b|headerHandles' -- macos/App macos/ProvenenciaTests ':!*Generated/*'`; the `.person`, `.event`, `.place` cases on those oneofs move to `.summary.kind`). Skill: `add-ffi-handler`.
Do: D-4 and D-7 in the proto; `git mv` the handler files (D-6); add `conclusionSummaryProto` and call it at the search and Promote sites; `attachSummary` sets `out.Summary`; rename the handler functions and D-3 locals; run `./scripts/generate-proto.sh`; in `GoStore.swift`, replace the four union bodies with one `mapConclusionSummary` that still returns `CatalogConclusionHeader` (Swift's own names change in STEP-5), and fix the other generated-name lines.
Check: in `git diff -- api/proto/engine.proto`, outside the four folded messages each removed line ending `= N;` has an added line with the same `N`, and the four folded messages show exactly D-4's kept and reserved numbers; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `scripts/check-pr.sh` PASS.

**STEP-5 — Swift wire types and store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `CoreInvoke.swift`, `FakeStore.swift`, and every file that names a renamed Swift type (`git grep -lE 'Catalog(Person|Event|Place|Conclusion)Header|CatalogHeaderPlace|list(Person|Event|Place)Headers|seeded(Event|Place)Headers' -- macos/App macos/ProvenenciaTests`).
Do: D-1 Swift type names, store methods, mappers, `FakeStore` seeds, `summaryHandles` (D-5). Type names only; properties and locals are STEP-7.
Check: `scripts/check-pr.sh` PASS.

**STEP-6 — The band's line becomes secondary**
Files: `ConclusionDetailPage.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `PlaceDetailView.swift`, `EventDetailContent.swift`, `EventDetailTests.swift`.
Do: D-10. D-13: `EventDetailContent` reads `EventSecondaryDisplay.content(summary, locale:)`; delete `summaryDate(_:locale:)`; the tests at `EventDetailTests.swift` 81, 92 and 144 give their detail a summary carrying the same dates and keep their expected strings; add T-7.
Check: `git grep -nwE 'Summary|summary|summaryDate|summaryPlace|EventDetailSummary' -- macos/App/Features/Conclusions/ConclusionDetailPage.swift 'macos/App/Features/Conclusions/*DetailView.swift' macos/App/Features/Conclusions/EventDetailContent.swift` prints only lines naming `CatalogConclusionSummary`; `scripts/check-pr.sh` PASS.

**STEP-7 — Swift names, files and tests**
Files: `PersonHeaderDisplay.swift`, `ConclusionHeaderDisplay.swift`, `PersonHeaderDisplayTests.swift`, `project.pbxproj`, `PersonDetailContent.swift`, and every Swift file the word sweep lists. Skill: `add-swift-test`.
Do: D-6 moves and pbxproj edits; D-11; D-3 properties and locals (`var summary: CatalogConclusionSummary?`, `detail.summary`); D-9 Swift test names. Work one folder at a time (`Platform/`, `Features/Conclusions/`, `Features/Promote/`, `Features/Catalog/`, `Features/Workspace/`, `macos/ProvenenciaTests/`); building between folders with Xcode or `xcodebuild` is allowed.
Check: the old-name sweep and the word sweep (keep list only) over `macos/`; `scripts/check-pr.sh` PASS.

**STEP-8 — Docs**
Files: per D-8; this plan; `docs/deployment-plan/spike-13/completed.md`.
Do: D-8 renames. After `gh pr create`: this plan's Status and Landed note, the map row status, one `completed.md` paragraph.
Check: the old-name sweep prints nothing; the word sweep prints only keep-list lines; the map sweep prints only S13-02 and S13-31 lines; `scripts/check-pr.sh` PASS (it runs `check-guidance-names.py`).

Commits: one per step. Subjects are full sentences ("Rename the conclusion header package to conclusion summaries.").

## 8. Tests

One new behavior test (T-7, D-13); otherwise each existing test passing under its new name pins the rename. The fold (D-4) is covered by the tests that read each oneof today.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/*_test.go` *(new)* (all, renamed per D-9) | the base cases → the base results | AC-1.3 |
| T-2 | Go | `api/ffi/handlers/conclusion_summaries_test.go` *(new)* `TestSummaryProtosCarryTerms`, `TestListPersonSummaries`, `TestListEventSummaries`, `TestListPlaceSummaries`; the `conclusion_detail` and `search_catalog` handler tests (read `GetSummary()`); `write_effects_test.go` (`SummaryHandles`) | the base cases → the base results | AC-1.3, D-4, D-5 |
| T-3 | Swift | the 15 test files the old-name sweep lists at the base, `PersonTitleDisplayTests` among them; `CatalogSearchTests`, `PromoteFlowTests`, `PersonDetailTests` read the folded union | the base cases → the base results | AC-1.4, D-4 |
| T-4 | Grep | old-name sweep (§2) | → no output | AC-1.1 |
| T-5 | Grep | word sweep (§2) | → keep-list lines only | AC-1.2, D-3 |
| T-6 | Grep | map sweep (§2) | → S13-02 and S13-31 lines only | AC-1.5, D-8 |
| T-7 | Swift | `macos/ProvenenciaTests/EventDetailTests.swift` `bandLineReadsTheSummary` *(new)* | summary date 14 May 1817, no date fields → band "14 May 1817"; `date` field with records and no displayed value, summary with start 1849 and end 1851 → band equals `EventSecondaryDisplay.content(summary, locale: en).date` (the span); locale `en` | AC-1.6, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `xcode-test` (pbxproj paths, the folded oneof), `guidance-names` (a skill naming an old symbol).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings: N/A, no copy changes
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated: N/A, none
- [ ] UI matches the brief inventory: N/A, no UI change
- [ ] Version bump as the header says: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan (after gh pr create)
- [ ] deployment-plan.md: map row status; docs/deployment-plan/spike-13/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/: N/A, not gated
- [ ] Domain doc updated: `conclusion-reconciliation.md` and `macos-client-patterns.md` name the new package and types
- [ ] PR description: what shipped, AC evidence (the three sweeps' output), deviations, manual check
```

## 10. Manual check

1. Build and run. Open a project with Persons, Events and Places. The Persons, Events and Places lists show the same rows, titles, dates and thumbnails as on `main`.
2. Open a Person: the band at the top shows the name, life lines, gender badge and thumb as on `main`. Open an Event: the date and place under the title match its Events list row. Open a Place (its chain): as on `main`.
3. Type a person's name in the omnibar: the hit row reads as on `main`. Open Promote on a Source: the proposal rows and their alternatives read as on `main`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A proto number is dropped in the fold | STEP-4 number check; D-4 lists each kept and reserved number |
| A pbxproj name is missed and a file drops out of a target | STEP-7 runs the full Xcode tests; D-6 edits in place |
| A band or layout name is renamed by a blind replace | D-2 keep list; D-3 type rule; the word sweep |

**Stop and report if:**

- Any name in §4 is stored or read back from disk, a JSON file, a golden file, `UserDefaults` or navigation history (a hit for an old name outside Go, proto and Swift source).
- A test other than T-7 can pass only by changing an asserted value, not a name, the oneof accessor, or (for D-13's three tests) the fixture.
- A file under §4 Current state has changed at your base so a row no longer holds (for example S13-04 landed first and added header names).
- A fifth `oneof` over person, event and place exists at your base that §4 does not list.

## 12. Open questions

None.

## 13. PR

**Title:** S13-31 — Header data becomes summaries
**Body:** objective in one sentence; D-13's band-line change and its edge case; US-1 against AC-1.1 – AC-1.5 with the three sweeps' output; the D-4 fold; deviations; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-31-header-data-becomes-summaries.md`.
