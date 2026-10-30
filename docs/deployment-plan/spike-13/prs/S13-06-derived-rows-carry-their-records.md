# S13-06 — Derived rows carry their records

**Status:** Blocked — on board S13-D12
**Spike / slice:** Spike 13 · Slice 2 — Provenance sheet
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D12](../design/S13-D12-life-rows-why.md), board not drawn yet. The Why disclosure appears on the four life rows, so the Vitals rows look different (user, 2026-10-30: "Why shows now")
**Board:** — (not agreed yet; the handoff records where the implementer sees the frames)
**Depends on:** S13-D12 (board). Shares `core/database/conclusiondetails/conclusiondetails.go` and `api/ffi/handlers/conclusion_detail.go` with S13-24 (Planned); either may land first, see §11
**Blocks:** S13-11
**Branch:** `s13-06-derived-rows-carry-their-records`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-10 and S13-23 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-12, R-13a, R-14; working rules: "A derived row carries its owning entity's records", "Each PR updates the docs and skills it makes false", "Engine PRs run ahead of their slice", "The provenance sheet replaces the Why on the Person page only"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-06; [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §9
**Planned:** 2026-10-30 against `main` @ `705cd98`
**Review:** [`S13-06-derived-rows-carry-their-records.review.md`](S13-06-derived-rows-carry-their-records.review.md) — 2 rounds: Fits with changes / Ready with changes; every finding Fixed or answered, no blocker open

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

On the Person page, the birth date, birth place, death date and death place rows show where their value comes from, like every other Vitals row: a Why disclosure listing the records behind it. Those values belong to other records (the birth event's date, the Place it happened at; R-13a), so today the rows carry no records and send the researcher to the Event or Place page instead. This PR has `GetConclusionDetail` return, for a Person, the reconciler records of the field each life row reads, keyed by the entity that owns it and tagged by the row, and shows them in the existing Why (S13-D12).

**Why now.** S13-11's provenance sheet opens from every Vitals value, including the life rows (bar 2). It needs these records on the wire first.

**What stays.** The rows' value, Mixed badge and open arrow; every other Vitals row; `LifeFacts` and `PersonHeader` on the wire (lists and tiles carry them and gain nothing); the Event and Place pages, including the Event page's place row, which keeps no Why; no new RPC.

## 2. User stories

**US-1.** As a researcher on a Person page, when I open the Why on the birth date row, I see every record of that date on the person's birth events, so that I can tell where the date comes from without leaving the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person promoted with one birth event whose date two records give (14 May 1817 from a register, 14 May 1817 from a census) | `GetConclusionDetail` returns one `derived_rows` entry with kind `DERIVED_ROW_KIND_BIRTH_DATE` whose single owner is that event, with `field.property_key` `date` and both records in `field.outcomes` (the map's Check) |
| AC-1.2 | The same person on the Person page | The Birth date row shows the Why button; opened, its title is "Why “14 May 1817”" and it lists both records, primary ink, with no group line (LW-1 – LW-3) |
| AC-1.3 | A birth event with no `date` record and one `start_date` record of 1817 | The owner's field is `start_date` with that record |
| AC-1.4 | A person with two surviving birth events, the lead dated 1817 (two records) and the other 1818 (one record) | The Birth date entry has two owners, the 1817 event first; the row's Why lists the 1817 records in primary ink under a group line naming that event, then the 1818 record in secondary ink under a group line naming the other (LW-4 (a), LW-5) |
| AC-1.5 | A person with no birth event | No Birth date or Birth place entry; the row shows no Why button and looks as today |
| AC-1.6 | A birth event whose only date record is denied, so the row is stated empty | The Birth date entry has the event's `date` field with the denied record; the row is stated empty, shows the Why button, and the Why title is "Why “Birth date”" |

**US-2.** As a researcher, when I open the Why on the birth place row, I see the records naming the Place, so that I can check the place name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A birth event located at one Place with toponym records "York" (two) and "Eboracum" (one) | The Birth place entry has one owner, that Place, field `toponym`, three outcomes; the row's Why lists them with no group line |
| AC-2.2 | A death event located at Montréal and at Paris (two Places that do not fold) | The Death place entry has two owners; the Why lists the Place the row shows first, then the other, each under its group line, the other's records in secondary ink |
| AC-2.3 | The same Place on two surviving birth events | The Birth place entry lists that Place once |
| AC-2.4 | A Place that has a parent (York part of Yorkshire) | Only York's toponym records are listed; Yorkshire's are not |
| AC-2.6 | A person born and died at York | The Birth place and Death place entries each list York once, with the same three outcomes; York is composed once |
| AC-2.5 | A death event at two Places where the first has no kept name (its only toponym record is denied) and the second is Paris | The row shows Paris, its open arrow opens Paris, and the Why lists Paris's records first in primary ink, then the first Place's under its group line in secondary ink |

**US-3.** As the engine, I keep the life records off every other read and keep the read's cost flat.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `GetConclusionDetail` on an Event or a Place | `derived_rows` is empty |
| AC-3.2 | `ListPersonHeaders` | Its `PersonHeader` and `LifeFacts` messages are unchanged on the wire |
| AC-3.3 | `PersonDerivedRows` for a person with one birth event and one Place, and for a person with two birth events and two Places each | Both take the same number of graph queries |

## 3. Scope

| In | Covered by |
| --- | --- |
| `conclusionheaders.PersonLifeOwners` and the shared `eventDate` rule: which entity and property each life row reads, every surviving event, lead first | STEP-1, AC-1.3, AC-1.4, AC-2.2, AC-2.3, AC-2.4 |
| `conclusiondetails.fieldsOf` (split out of `ForEntity`, batched over several handles) and `PersonDerivedRows` | STEP-2, AC-1.1, AC-1.6, AC-2.1, AC-3.3 |
| `DerivedRowKind`, `DerivedRow`, `ConclusionOwnerField`, `ConclusionDetail.derived_rows`; handler fills it for a Person | STEP-3, AC-1.1, AC-3.1, AC-3.2 |
| Swift store types and GoStore mapping | STEP-4 |
| One shown Place: `DerivedPlace.shown` | STEP-5, AC-2.5 |
| The four life rows carry record groups, a Why title, owner group lines and the secondary ink rule (S13-D12 LW-1 – LW-7, LW-W1) | STEP-6, STEP-7, AC-1.2, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.5 |
| The "its page owns the Why" comments; `macos-client-patterns.md` and `conclusion-reconciliation.md` §9 | STEP-8, STEP-9 |

| Out | Home |
| --- | --- |
| The provenance control and sheet | S13-11 |
| The Event page's derived place row carrying records | Not planned this spike (map: "can take the same path later": one `DerivedRowKind` case) |
| Records of the event-to-Place location link and of the parent chain's names | Not planned (user, 2026-10-30) |
| Any change to `LifeFacts`, `PersonHeader` or list reads | Not planned (working rule) |
| The Event page choosing its date property in Swift (`EventDetailContent.dateRow`) while the Person page's comes from Go | Not planned; logged in the review log (M-15) |
| A shared Go test fixture for bridges across packages | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Detail composer | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity`, `loadProperties`, `attachValues`, `loadOutcomes`, `state`, `sqlPropertyIDs`, `sqlOutcomes` | `ForEntity` composes one handle: node lookup, `loadProperties` (`sqlPropertyIDs`, `WHERE ao.entity_id = ?`, ordered by the type's binding order), `attachValues`, `gallery.ForEntities`, `loadOutcomes` (`sqlOutcomes`, `WHERE ao.entity_id = ?`, attaches to fields by property id and returns gallery lead outcomes), then a loop that sets `State` and sorts `Outcomes` (value rank, none last, then Observation id) |
| Life facts | `core/database/conclusionheaders/conclusionheaders.go` `LifeFacts`, `HeaderPlace` | `LifeFacts.Event` is the lead event only; comment "Its page owns the date's Why". `HeaderPlace`: "whose page owns the names' Why" |
| Life walk | `core/database/conclusionheaders/compose.go` `onePerson`, `lifePick`, `lifeFrom`, `leadLife`, `lifeBefore`, `eventPlaces`, `eventWhen`, `composePersons`, `buildPlaceGraph`, `vocab.prop` | `onePerson` follows `hops.EventsOfSubject`, keeps events whose kept `event_type` key is `birth` or `death` (nil nodes skipped), and builds a `lifePick` each. `lifeFrom` and `eventWhen` each repeat "kept `date`, else kept `start_date`". `leadLife` picks the lead by `lifeBefore` and folds its Places with `eventPlaces`; `composePersons` builds the `placeGraph` from every event's location links |
| Event headers | `conclusionheaders.EventsByIDs`, `EventHeader.Title` | Composes Event headers (title plan, type) for given ids |
| Handler | `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`, `conclusionDetailProto`, `attachHeader`, `attachPictures`; `conclusion_headers.go` `eventHeaderProto`; `canonicalEntityProto` | `attachHeader` reads the kind's header; the Place case also calls `conclusionheaders.AttachPlaceRelationships` |
| Proto | `api/proto/engine.proto` `ConclusionDetail` (fields 1 – 7), `ConclusionField`, `EventHeader`, `EventTitleRule` (enum naming precedent), `HeaderPlace` | `HeaderPlace.entity`: "its page owns the names' Why" |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionDetail`, `CatalogHeaderPlace`, `CatalogLifeFacts`; `GoStore.swift` `getConclusionDetail`, `mapConclusionHeader` | Fields mapped inline in `getConclusionDetail`; `CatalogHeaderPlace.entity` and `CatalogLifeFacts.event` comments say the page owns the Why |
| Life rows | `macos/App/Features/Conclusions/PersonDetailContent.swift` `rows(_:header:locale:orders:)`, `lifeLine` | Builds the four life rows with `ReconciledValueRowModel.derived(…, opens:)`; the place row opens `side?.places.first` |
| Shown Place | `macos/App/Features/Conclusions/DerivedPlace.swift` `name`, `extra` | `name` shows the first Place whose `PlaceChainDisplay.placeLine` is non-empty; nothing exposes which Place that is |
| Row model | `macos/App/Features/Conclusions/ReconciledValueRowModel.swift` `records`, `init(field:locale:orders:)`, `spanning`, `derived`, `empty`, private `init`, `Opens`, `opens` | `records: [ReconciliationRecord]`; `spanning` already maps several fields' outcomes to records; `derived` sets `records: []`, `whyTitle: ""` |
| Row view | `ReconciledValueRow.swift` `body`, `disclosures` | Passes `model.records` to `ReconciliationReasoningView`; shows the Why button when `model.records` is non-empty |
| Why table | `ReconciliationReasoningView.swift` (takes `records`); `ReconciliationRecord.swift` (`counted` is a `var`) | One flat grid; no grouping |
| Titles | `EventTitleDisplay.title(_:locale:orders:)` (takes `CatalogEventHeader`), `EventTitleDisplay.untitled(_:)`; `PlaceTitleDisplay.titleSource(_:)` (takes `PlaceTitleParts(names:label:ref:)`) | The one titling rule per kind |
| Tests | `conclusiondetails_test.go` (`TestForEntityQueryCountIsConstant`, fixture `subjectOn`, `cite`, `promote`); `core/database/conclusionheaders/placechains_test.go` `TestBirthLocationsFoldWhenPartOf` and fixture `placeSubject`, `placeRel`, `participation`, `location`, `promoteSubject`, `pointYear`; `conclusion_detail_test.go` `TestGetConclusionDetailCarriesItsHeader`; `connect_test.go` `connectFixture`; `PersonDetailTests.swift` | `PersonDetailTests` expects the life rows' `records.isEmpty` (lines 60, 270, 306) |

**Precedent.** `attachHeader`'s Place case: a page-only read attached in the handler beside the header, never on the list header. `gallery.ForEntities` and `core/database/sqlbatch` `Query`: by-id reads for several handles, batched. `ReconciledValueRowModel.spanning`: one row's Why from several fields.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The date row carries the date field of **every** surviving birth (or death) event, ordered by `lifeBefore` (lead first). The place row carries the `toponym` field of every Place of those events: for each event in that order, `eventPlaces(g, v, pg, event)` (folded at that event's date), appending each `HeaderPlace.Entity` whose id is not yet listed. "Surviving" is exactly what `onePerson` keeps today, through the shared `lifePicks` (D-4) | User, 2026-10-30. The working rule was updated in the planning commit | Lead event only — a Mixed row from a second event would show none of the disagreement |
| D-2 | One rule for an Event's date: `eventDate(v vocab, ev *graphcache.Node) (date *datevalue.Value, count int, sortKey string, ok bool, propID []byte, err error)` *(new)* in `compose.go` reads the kept `date`, else the kept `start_date`; `propID` is the property read, `v.prop(vocabkey.PropertyDate)` when neither has a value. `lifeFrom` and `eventWhen` both call it; `lifePick` gains `dateProp []byte` | The rule lives once (review M-5); the owned field is the one the value was read from | Keeping `eventWhen`'s copy |
| D-3 | Parent-chain Places and the location bridge's records are not carried | User, 2026-10-30 | — |
| D-4 | `conclusionheaders` gains `type FieldOwner struct { Entity canonicalentities.Entity; PropertyID []byte }` *(new)*, `type LifeOwners struct { Dates, Places []FieldOwner }` *(new)*, `func PersonLifeOwners(c *database.Catalog, personID []byte) (birth, death LifeOwners, err error)` *(new)*, and `lifePicks(g, v, n) (births, deaths []lifePick, err error)` *(new)* extracted from `onePerson`, which calls it. `lifePicks` first collects the person's `follow(g, n.ID, hops.EventsOfSubject)` neighbors and calls `g.Nodes` on them before its per-event loop (`Nodes` skips cached ids, so `onePerson` under `composePersons` costs nothing more), so the walk's graph queries do not grow with the number of events. `PersonLifeOwners`: nil graph or nil or merged person → zero values, nil; (1) `lifePicks`; (2) collect every pick's `follow(g, id, hops.PlacesOfEvent)` neighbors and `buildPlaceGraph(g, v, placeIDs)`, as `composePersons` does; (3) `sort.SliceStable` each side by `lifeBefore`; (4) `Dates`: `FieldOwner{entityOf(pick.event), pick.dateProp}` per pick; `Places`: D-1, with `v.prop(vocabkey.PropertyToponym)` | `conclusionheaders` owns the walk; one walk, two callers; property ids resolved once (review M-7) | Fields on `LifeFacts` (working rule); a property key re-resolved by origin in `conclusiondetails` |
| D-5 | `ForEntity` splits into its steps (review M-2): `func fieldsOf(g *graphcache.Graph, q Querier, nodes []*graphcache.Node, leads map[string]bool) (map[string][]Field, map[string]Outcome, error)` *(new)* skips nil and merged nodes and duplicate ids (first occurrence kept), then runs, for the remaining nodes at once, `loadProperties` → `attachValues` (per node) → `loadOutcomes` → state and sort (the loop body moved into `finish(f *Field)` *(new)*), and returns fields by entity id. `sqlPropertyIDs` becomes the prefix `SELECT DISTINCT ao.entity_id, ao.property_id FROM auto_reconciler_outcomes ao WHERE ao.entity_id IN (`, read with `sqlbatch.Query(q, sqlPropertyIDs, ")", nil, ids, scan)`; `sqlOutcomes` selects `ao.entity_id` as its first column (the scan reads it first), its last line becomes `WHERE ao.entity_id IN (`, and `loadOutcomes` reads it with `sqlbatch.Query(q, sqlOutcomes, ")", nil, ids, scan)`, the scan callback holding today's row body (as `gallery.ForEntities` does; rows arrive batch by batch, and the per-entity `TypeOrder` sort and `finish` order them); `loadProperties` and `loadOutcomes` key fields by entity id then property id and order each entity's fields by its own `TypeOrder`. `ForEntity` calls `fieldsOf` with its one node and the gallery leads, then adds pictures | One composition for every field on every page, so the next rule (a Reconciliation Claim overlay, package doc) reaches the life rows; constant reads as owners grow (review M-8) | A second per-owner sequence; one SQL round per owner |
| D-6 | `func PersonDerivedRows(c *database.Catalog, personID []byte) ([]DerivedRow, error)` *(new)* in `conclusiondetails`: `PersonLifeOwners`; one `g.Nodes` over `database.UniqueBlobIDs` of every owner entity id; `fieldsOf` over those nodes with `leads` nil, so a Place on both the birth and the death row is composed once and both rows share its `Field`; each owner keeps the field whose `PropertyID` equals its own (none → owner dropped); Event owners get their `conclusionheaders.EventHeader` from one `EventsByIDs` call. Types: `type DerivedRowKind string` *(new)* with `DerivedRowBirthDate`, `DerivedRowBirthPlace`, `DerivedRowDeathDate`, `DerivedRowDeathPlace`; `type OwnerField struct { Entity canonicalentities.Entity; Event *conclusionheaders.EventHeader; Field Field }` *(new)*; `type DerivedRow struct { Kind DerivedRowKind; Owners []OwnerField }` *(new)*. A row with no owners is not returned. Rows come in the order birth date, birth place, death date, death place. `conclusiondetails` imports `conclusionheaders` (nothing in `core` imports `conclusiondetails`; no cycle) | `conclusiondetails` owns `Field`; tagged rows mean the Event place row is one more kind (review M-3) | One struct field per row |
| D-7 | Wire, beside `ConclusionField`: `enum DerivedRowKind { DERIVED_ROW_KIND_UNSPECIFIED = 0; DERIVED_ROW_KIND_BIRTH_DATE = 1; DERIVED_ROW_KIND_BIRTH_PLACE = 2; DERIVED_ROW_KIND_DEATH_DATE = 3; DERIVED_ROW_KIND_DEATH_PLACE = 4; }`; `message ConclusionOwnerField { CanonicalEntity entity = 1; ConclusionField field = 2; EventHeader event = 3; }` (`event` set for an Event owner); `message DerivedRow { DerivedRowKind kind = 1; repeated ConclusionOwnerField owners = 2; }`; `ConclusionDetail.derived_rows = 8` (`repeated DerivedRow`), comment "a Person's life rows: the records of each field a row reads off another handle, keyed by that handle (S13-06)" | Map: keyed by owning entity, on `ConclusionDetail`, no new RPC. A proto enum is a typed Swift enum, so no string vocabulary crosses the boundary | — |
| D-8 | The handler fills `derived_rows` in `attachHeader`'s Person case after the header is set (an absent header returns first, leaving it empty), through `conclusiondetails.PersonDerivedRows` and `derivedRowsProto([]conclusiondetails.DerivedRow) []*engine.DerivedRow` *(new)*, which maps each owner with `canonicalEntityProto`, `conclusionFieldProto` and, when `Event` is set, `eventHeaderProto`. `conclusionDetailProto`'s per-field body moves into `conclusionFieldProto(f conclusiondetails.Field) *engine.ConclusionField` *(new)* | Precedent `AttachPlaceRelationships` | Composing in the handler (Known traps) |
| D-9 | Swift store: `enum CatalogDerivedRowKind { case birthDate, birthPlace, deathDate, deathPlace }` *(new)*; `struct CatalogOwnerField: Sendable, Equatable { var entity: CatalogCanonicalEntity; var event: CatalogEventHeader?; var field: CatalogConclusionField }` *(new)*; `struct CatalogDerivedRow: Sendable, Equatable { var kind: CatalogDerivedRowKind; var owners: [CatalogOwnerField] }` *(new)*; `CatalogConclusionDetail.derivedRows: [CatalogDerivedRow] = []`. GoStore maps them, dropping a row of kind `.unspecified` or an unknown case; the inline field mapping moves into `private static func mapConclusionField(_:)` *(new)*; the event header maps through the function `mapConclusionHeader` uses for `.event` | Default empty keeps every existing initializer call compiling | — |
| D-10 | One shown Place (review M-4): `DerivedPlace.shown(_ places: [CatalogHeaderPlace]) -> CatalogHeaderPlace?` *(new)* returns the first Place whose `placeLine` is non-empty; `name` and `extra` use it; `PersonDetailContent.rows` opens `DerivedPlace.shown(side?.places ?? [])` instead of `.first`. The shown owner of a date row is the owner whose entity id is `side?.event?.id`; of a place row, the one whose id is `DerivedPlace.shown(…)?.entity.id` | The label, the open arrow and the ink read one fact | Go marking `shown` — Go does not know `placeLine`, a Swift formatter |
| D-11 | Record groups (review M-13, M-6): `struct ReconciliationRecordGroup: Equatable, Identifiable { var owner: OwnerLine?; var records: [ReconciliationRecord] }` *(new)* with `struct OwnerLine: Equatable { var id: String; var text: String; var spoken: String }` *(new, nested)*; `id` is the owner's entity id, else `""`. `ReconciledValueRowModel.records` becomes `recordGroups: [ReconciliationRecordGroup]` plus a computed `var records: [ReconciliationRecord] { recordGroups.flatMap(\.records) }`. `private static func records(_ fields: [CatalogConclusionField], locale: Locale, orders: NameFormatOrders) -> [ReconciliationRecord]` *(new)* maps fields' outcomes to records; `init(field:)` and `spanning` use it and make one group with `owner: nil`; `empty` makes none | One mapper, one grouping structure the sheet can read | Grouping by comparing adjacent strings in the view |
| D-12 | `derived` gains `owners: [CatalogOwnerField] = []`, `shownOwnerID: String? = nil`, `locale: Locale = .autoupdatingCurrent` and `orders: NameFormatOrders = NameFormatOrders()` (`PersonDetailContent.rows` passes its own; `EventDetailContent.placeRow`'s call compiles unchanged). Groups: the shown owner first, then the rest in Go's order, one group each, records from `records([owner.field], …)`; records of every group but the shown owner's get `counted = false` (LW-5); when no owner is the shown one, all are secondary. `owner` is set on every group when there are two or more groups; with one group it is nil, whether or not that owner is shown (LW-3). `whyTitle` is `L10n.Conclusions.whyTitle(shown ?? label)` when there are records, else `""` | The ink rule follows the value on screen | — |
| D-13 | Group line (review M-1): title = `EventTitleDisplay.title(event, locale:orders:)` when `owner.event` is set, else `EventTitleDisplay.untitled(owner.entity).text` for a date row; for a place row `PlaceTitleDisplay.titleSource(PlaceTitleParts(names: <the field's displayed values as text>, label: entity.label, ref: entity.ref)).text`. `text` = the ref alone when the title equals the ref, else `L10n.Conclusions.whyOwner(title, ref:)` *(new)* "%1$@ · %2$@" ("Birth of Ada Lovelace · EVT-4KX2M"); `spoken` = `L10n.Conclusions.a11yList(title, rest: ref)` or the ref. Built by `private static func ownerLine(_:kind:locale:orders:)` *(new)* on `ReconciledValueRowModel` | The one titling rule per kind, as every list, header and Promote use; numbered format required by `scripts/check-localizable-xcstrings.py` | Label or value text — a third title form |
| D-14 | **Pending the board (LW-4).** `ReconciliationReasoningView` takes `groups: [ReconciliationRecordGroup]` instead of `records`; `ReconciledValueRow` passes `model.recordGroups`. Before a group whose `owner` is set it draws one grid row: `Text(verbatim: owner.text)`, `pvMicroCaps()`, `PVColor.textMuted`, `.gridCellColumns(4)`, wrapping (`.fixedSize(horizontal: false, vertical: true)`, no line limit), `.accessibilityAddTraits(.isHeader)`, `.accessibilityLabel(owner.spoken)`. If the board picks LW-4 (b), this Decision, AC-1.4, AC-2.2 and AC-2.5 are rewritten before the plan is Planned | Brief recommendation (a) | — |

**Rules this plan bends.** None open. D-1 changed the working rule "A derived row carries its owning entity's records" from the lead event to every surviving event (user, 2026-10-30); the rule and the scope entry were updated in the planning commit.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `conclusionheaders.FieldOwner`, `LifeOwners`, `PersonLifeOwners`, `lifePicks`, `eventDate`, `lifePick.dateProp` *(new)* | type, func | `onePerson`, `lifeFrom`, `eventWhen` | Extracted from them; `eventWhen`'s copy of the date rule goes |
| `conclusiondetails.fieldsOf`, `finish` *(new)* | func | `ForEntity` | Its own steps, split out and batched |
| `conclusiondetails.DerivedRowKind`, `OwnerField`, `DerivedRow`, `PersonDerivedRows` *(new)* | type, func | `Detail`, `Field` | Wraps `Field`; no field duplicated |
| `DerivedRowKind`, `ConclusionOwnerField`, `DerivedRow`, `ConclusionDetail.derived_rows` *(new)* | proto | `ConclusionField`, `EventHeader`, `EventTitleRule` | Wraps them |
| `conclusionFieldProto`, `derivedRowsProto` *(new)* | func | `conclusionDetailProto` loop, `canonicalEntityProto`, `eventHeaderProto` | Moved and composed |
| `CatalogDerivedRowKind`, `CatalogOwnerField`, `CatalogDerivedRow`, `mapConclusionField` *(new)* | Swift type, func | `CatalogConclusionField`, GoStore inline mapping | Wrap and move |
| `DerivedPlace.shown` *(new)* | func | `DerivedPlace.name` | `name` now calls it |
| `ReconciliationRecordGroup`, `OwnerLine`, `recordGroups`, `records(_:locale:orders:)`, `ownerLine` *(new)* | type, property, func | `spanning`'s mapping, `records` | `records` stays as a computed view; `spanning` uses the shared mapper |
| `L10n.Conclusions.whyOwner` *(new)* | key `conclusions.why.owner` | `pageSectionLeaf` ("%1$@ · %2$@") | Different meaning and translator comment |
| `core/database/conclusiondetails/export_test.go` *(new)* | test helper | `conclusionheaders/export_test.go` | Same pattern |

**Builds on.** `ForEntity`'s composition, `sqlbatch.Query`, the `attachHeader` page-only attach, `EventTitleDisplay` and `PlaceTitleDisplay`, the existing Why disclosure.
**Deletes.** `eventWhen`'s own date fallback (it calls `eventDate`); `ForEntity`'s inline steps (moved into `fieldsOf`); `spanning`'s inline record mapping; `rows`' `side?.places.first` open target; the "its page owns the Why" comment text.
**Next adds.** The Event page's place row: one `DerivedRowKind` case and an owner list from the Event's Places. S13-11 reads `recordGroups` for the sheet, and its scope now says it removes or moves the group-line branch (STEP-9).

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A derived row with records | `ReconciledValueRow` Why, `ReconciliationReasoningView` | Takes any group list; Source link uses the record's own ids |
| A record from another handle | `ReconciliationRecord.composerLocation` | Uses the outcome's Source, graph, Subject and Citation ids |
| An owner merged or gone between reads | `PersonDerivedRows` | Skipped (D-6); both reads run inside one `withProjectCatalog` call, which the catalog session serializes |
| The place row's open arrow when the first Place has no line | `PersonDetailContent.rows` | Now opens the shown Place (D-10, AC-2.5) |

### Scale

No write. Per Person page open: one `PersonLifeOwners` walk (the hops `composePersons` already runs), one `g.Nodes`, one `EventsByIDs`, and two batched SQL reads for all owners (`sqlPropertyIDs`, `sqlOutcomes` through `sqlbatch.Query`), plus the date and name lookups `loadOutcomes` already batches. Constant in the number of owners (AC-3.3, T-8). Lists are untouched (AC-3.2).

### Edges

- **Assumptions.** `lifeBefore` is a strict order (dated first, sort key, then ref): sorting puts the lead first, as `leadLife` picks it. Places dedupe by entity id across events. `fieldsOf` with one node returns what `ForEntity` returned before (existing `ForEntity` tests unchanged). T-1, T-2.
- **Neighbors.** One owner (no group line) vs two; shown owner first vs not shown at all; `date` vs `start_date` vs neither (D-2); a nameless first Place (AC-2.5). T-1, T-5, T-6, T-7.
- **Removed inputs.** An owner merged or gone: dropped. A denied or outvoted record stays, with its outcome.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** Events per side and Places per event come from catalog links; the `IN` list has one id per owner, batched by `sqlbatch.Size`, so SQLite's bind limit never applies.

### Schema

None — no catalog change.

### Go core

As D-2, D-4, D-5, D-6. The `conclusiondetails` package doc gains: "A Person's life rows read other handles' fields through PersonDerivedRows, composed by the same steps."

Errors: none new.

### FFI

`api/proto/engine.proto` per D-7. Handler per D-8. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Life row | `ReconciledValueRow` | `macos/App/Features/Conclusions/ReconciledValueRow.swift` | Passes `recordGroups` |
| Life row model | `ReconciledValueRowModel.derived(…, owners:, shownOwnerID:)` | `ReconciledValueRowModel.swift` | D-11 – D-13 |
| Group line | Grid row in `ReconciliationReasoningView` | `ReconciliationReasoningView.swift` | D-14 |
| Shown Place | `DerivedPlace.shown` | `DerivedPlace.swift` | D-10 |
| Disclosure, open arrow, badge, card | `PVDisclosureButton`, `PVIconButton`, `PVBadge`, `PVCard` | `macos/App/DesignSystem/Components/…` | Unchanged |

`PersonDetailContent.rows` gains `derivedRows: [CatalogDerivedRow] = []`; `init(detail:)` passes `detail.derivedRows`; the `life` table gains a `CatalogDerivedRowKind` per entry, and each row takes `derivedRows.first { $0.kind == kind }?.owners ?? []`. States: no owners → no Why. Frames: copied from the board at handoff.

### Strings

L10n keys *(new)*: `L10n.Conclusions.whyOwner(_ title: String, ref: String)` → `conclusions.why.owner` = "%1$@ · %2$@", comment "Group line in a Why list naming the record the following rows belong to; arguments are its title and ref". Removed: none.

## 7. Steps

**STEP-1 — Life owners in Go**
Files: `core/database/conclusionheaders/conclusionheaders.go`, `compose.go`, `placechains_test.go`.
Do: D-2, D-4; add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/...` passes; the existing header tests are unchanged.

**STEP-2 — fieldsOf and PersonDerivedRows**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `conclusiondetails_test.go`, `export_test.go` *(new)*.
Do: D-5, D-6; add T-2 and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/...` passes, existing `ForEntity` tests unchanged; `go list -deps ./core/database/conclusionheaders | rg conclusiondetails` finds nothing.

**STEP-3 — Wire and handler**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/conclusion_detail.go`, `conclusion_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-7, D-8; `./scripts/generate-proto.sh`; add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-4 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`.
Do: D-9.
Check: the app builds (`xcodebuild test` as `scripts/check-pr.sh` runs it).

**STEP-5 — One shown Place**
Files: `macos/App/Features/Conclusions/DerivedPlace.swift`, `PersonDetailContent.swift`, `macos/ProvenenciaTests/DerivedPlaceTests.swift`.
Do: D-10 (the open target only; records come in STEP-6); add T-5.
Check: Xcode tests pass.

**STEP-6 — Life rows carry record groups**
Files: `ReconciledValueRowModel.swift`, `ReconciliationRecord.swift`, `PersonDetailContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-11 – D-13. The existing `records.isEmpty` expectations (lines 60, 270, 306) stay: those details have no `derivedRows`. Add T-6, T-7.
Check: Xcode tests pass; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-7 — Group lines in the Why**
Files: `ReconciliationReasoningView.swift`, `ReconciledValueRow.swift`. Skill: `add-ui-component`.
Do: D-14, matched to the agreed board's frames 3 and 4.
Check: manual check §10 steps 3 – 5.

**STEP-8 — Comments**
Files: `core/database/conclusionheaders/conclusionheaders.go` (`LifeFacts`: "Its page owns the date's Why" → "A Person page reads its date's records through conclusiondetails.PersonDerivedRows"; `HeaderPlace`: "whose page owns the names' Why" → "whose toponym records a Person page reads through conclusiondetails.PersonDerivedRows"), `api/proto/engine.proto` (`HeaderPlace.entity`: "entity is the Place; a Person page's life rows carry its toponym records in ConclusionDetail.derived_rows"), `macos/App/Platform/GenealogyStore.swift` (`CatalogHeaderPlace.entity` the same sense; `CatalogLifeFacts.event`: "A Person page reads its date's records from `CatalogConclusionDetail.derivedRows`"), `ReconciledValueRowModel.swift` (`Opens`: "The handle a derived value is read from"; `opens`: "A derived row opens the Event or Place it is read from. Nil otherwise."; `derived`'s doc: its record groups are the owning handles' fields, the shown owner first). Regenerate proto.
Check: `rg -n "owns[^.]*Why|Why stays on the Event or Place" core api macos/App -g '!**/Generated/**' -g '!api/proto/engine/**'` finds nothing.

**STEP-9 — Docs**
Files: `docs/macos-client-patterns.md` (the `GetConclusionDetail` payload line near line 101: "for a Person, the records behind its life rows, keyed by the owning Event or Place"), `docs/conclusion-reconciliation.md` §9 (one sentence: "A derived value's Why lists each owning entity's field records, judged on that entity"), `../completed.md`, `../deployment-plan.md` (map row status; S13-11's scope entry gains "Removes the owner group-line branch S13-06 added to `ReconciliationReasoningView`, or moves it into the sheet, which keeps S13-D12's grouping"), this plan (status, Landed note), `../design/` (S13-D12 → `design/archive/`, `design/README.md`).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show where a birth date comes from on the Person page."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionheaders/placechains_test.go` `TestPersonLifeOwners` *(new)*, table-driven. Build cases as `TestBirthLocationsFoldWhenPartOf` does: `placeSubject`, `placeRel(child, parent, propertyterms.KeyPartOf, nil, nil)`, `subjectstest.On(…, "event", "")`, `cite` with `event_type` and `date` (`pointYear`) or `start_date` (`observations.Input{PropertyID: f.prop("start_date").ID, Date: pointYear(1817)}`), `participation(person, event, "subject")`, `location`, `promoteSubject` on each. Two Places that do not fold: two `location`s to Places with no `placeRel` | (a) one birth event with `date` at York → `birth.Dates` = [{event, date id}], `birth.Places` = [{York, toponym id}]; (b) `start_date` only → start_date id; (c) birth events 1818 and 1817 → 1817 first; (d) death event at Montréal and Paris → two Places in fold order; (e) York on two birth events → once; (f) York part of Yorkshire → York only; (g) no events → zero; (h) unknown id → zero, nil | AC-1.3, AC-1.4, AC-2.2, AC-2.3, AC-2.4, D-1, D-2 |
| T-2 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestFieldsOf` *(new)*, table-driven through `func FieldsOf(c *database.Catalog, ids [][]byte) (map[string][]Field, error)` in `export_test.go` *(new, package `conclusiondetails`: loads the nodes with `c.Graph().Node`, opens `c.DB()` and calls `fieldsOf` with `leads` nil)*. `newFixture` gains `f.date, err = properties.Lookup(c, "date", vocabkey.OriginProvenencia)` beside `f.name`; date Observations are `observations.Input{SubjectID: s.ID, PropertyID: f.date.ID, Date: &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}}` through the existing `cite` | two promoted events (`subjectOn(src, "event")`, `promote`), the first with two `date` records, the second with one → by entity: first has field `date` with two outcomes sorted as `ForEntity` sorts and state merged; second has one; the first id passed twice → one entry, values not doubled; a merged id → no entry; one id → the same fields `ForEntity` returns | AC-1.1, AC-1.6, AC-2.6, D-5 |
| T-3 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailCarriesDerivedRows` *(new)*. Fixture over the FFI: `connectFixture`; `CreateCitedBridge` participation as `TestCreateCitedBridge` sends it, with the role term `subject` (from `ListPropertyTerms` on the role property); the `event_type` and `date` property ids from the same `ListProperties` response `connectFixture` reads (match `GetKey()`), the `birth` term from `ListPropertyTerms` on `event_type`; two `CreateCitationWithObservations` calls on the event, the first with `event_type` = `birth` and `date` 14 May 1817, the second with the same `date`; `promoteOK` on the person and the event | → one `derived_rows` entry, kind `DERIVED_ROW_KIND_BIRTH_DATE`, owner keyed by the event with `event` set, field `date`, two outcomes | AC-1.1 (map Check) |
| T-4 | Go | `TestGetConclusionDetailCarriesItsHeader` (extend each case) | Event and Place → `len(GetDerivedRows()) == 0` | AC-3.1 |
| T-5 | Swift | `macos/ProvenenciaTests/DerivedPlaceTests.swift` `shownSkipsANamelessPlace` *(new)* | [nameless Place, Paris] → `shown` is Paris; `name` "Paris"; `extra` 1; Person life row opens Paris | AC-2.5, D-10 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeRowsCarryTheirOwnersRecords` *(new)* | Birth date row with one owner, two outcomes → one group, `owner == nil`, 2 records, `whyTitle == "Why “14 May 1817”"`; no Death date row → records empty, `whyTitle == ""`; Birth date with one denied outcome and no displayed value → `lead == nil`, 1 record, `whyTitle == "Why “Birth date”"` | AC-1.2, AC-1.5, AC-1.6, D-12 |
| T-7 | Swift | same file `competingOwnersAreGroupedAndSecondary` *(new)* | two Birth date owners, header's lead event the second in Go order → groups in shown-first order, the other group's `counted == false`, `owner.text` "Birth of Ada Lovelace · EVT-…" from a `CatalogEventHeader` with the subject rule; owner with no header → `EventTitleDisplay.untitled` text; two Death place owners → Place titles from `PlaceTitleDisplay`; `owner.spoken` "Birth of Ada Lovelace, EVT-…" | AC-1.4, AC-2.2, AC-2.5, D-12, D-13 |
| T-8 | Go | `core/database/conclusionheaders/placechains_test.go` `TestPersonDerivedRows` *(new; external test package `conclusionheaders_test`, which may import `conclusiondetails`)*, built with the T-1 helpers | (a) one birth event with one Place vs two birth events with two Places each → after `g.Drop()` and `g.Vocabulary()`, equal graph query counts (`g.SetQueryCounterForTest`, as `TestForEntityQueryCountIsConstant`); (b) born and died at York → birth place and death place rows each one owner, York, with equal fields | AC-3.3, AC-2.6, D-5, D-6 |

AC-2.1 is covered by T-1 (a) and T-3's path and by the manual check; AC-3.2 by the unchanged `TestListPersonHeaders`.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `conclusiondetails` `ForEntity` tests (STEP-2 split), `conclusionheaders` header and display-cache tests (STEP-1), `PersonDetailTests`, `ReconciledValueDisplayTests`, `DerivedPlaceTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the S13-D12 board inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and S13-11's scope line; completed.md: one short paragraph of what shipped
- [ ] Brief S13-D12 moved to design/archive/ and design/README.md updated
- [ ] Domain docs: macos-client-patterns.md and conclusion-reconciliation.md §9 (STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person promoted with a birth event dated by two records and located at York (two toponym records). On Vitals, open the Birth date Why: two records, title "Why “14 May 1817”", no group line. Click a Source: the composer opens on that Observation; Back returns.
2. Open the Birth place Why: York's records, no group line.
3. Add a second birth event (1818) from another Source and promote it. The Birth date row shows Mixed; its Why lists both events under group lines titled as the Events list titles them, the 1818 record in secondary ink.
4. Give the death event two Places that do not fold; the Death place Why lists both Places' names under group lines.
5. Deny the first death Place's only name: the row shows the other Place, its arrow opens it, and its records come first.
6. Resize the window across 1600pt with a Why open: it stays open and in place (LW-W1). Open an Event and a Place page: they look as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S13-24 rewrites `sqlOutcomes`' columns and joins, the outcome scan and `reconcilerOutcomeProto` in the same files | This PR changes `sqlOutcomes`' first column and `WHERE`, `sqlPropertyIDs`, and moves `ForEntity`'s steps; S13-24 changes the provenance columns. Whichever lands second rebases; both are column- or clause-local |
| The `ForEntity` split changes detail output | Existing `ForEntity` tests must pass unchanged in STEP-2, and T-2 compares one-node `fieldsOf` with `ForEntity` |

**Stop and report if:**

- `onePerson` no longer walks `hops.EventsOfSubject` with the `birth` / `death` term filter, or `lifeFrom` no longer reads `date` then `start_date`.
- A birth event's date records are not in `auto_reconciler_outcomes` under the event's entity id (T-3 finds no outcomes for a promoted event with two date records).
- The participation bridge does not create the canonical link in T-3's fixture with `PromoteSubject` on each end.
- T-8's counts differ because `PersonLifeOwners`' walk issues graph queries per event that `composePersons`' walk avoids.
- The agreed board's LW-4 option or frames contradict D-12 – D-14 or AC-1.4, AC-2.2, AC-2.5.

## 12. Open questions

1. **LW-4 grouping option** — decided by the S13-D12 board. Recommendation: (a), as D-14 drafts.

## 13. PR

**Title:** S13-06 — Derived rows carry their records
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.md`.
