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
**Context:** [`person-detail.md`](../person-detail.md) R-12, R-13a, R-14; working rules: "A derived row carries its owning entity's records", "Each PR updates the docs and skills it makes false", "Engine PRs run ahead of their slice", "The provenance sheet replaces the Why on the Person page only"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-06
**Planned:** 2026-10-30 against `main` @ `705cd98`
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

On the Person page, the birth date, birth place, death date and death place rows show where their value comes from, like every other Vitals row: a Why disclosure listing the records behind it. Those values belong to other records (the birth event's date, the Place it happened at; R-13a), so today the rows carry no records and send the researcher to the Event or Place page instead. This PR has `GetConclusionDetail` return, for a Person, the reconciler records of the field each life row reads, keyed by the entity that owns it, and shows them in the existing Why (S13-D12).

**Why now.** S13-11's provenance sheet opens from every Vitals value, including the life rows (bar 2). It needs these records on the wire first; the map lets this engine PR land ahead of its slice.

**What stays.** The rows' value, Mixed badge and open arrow; every other Vitals row; `LifeFacts` and `PersonHeader` on the wire (lists and tiles carry them and gain nothing); the Event and Place pages, including the Event page's place row, which keeps no Why; no new RPC.

## 2. User stories

**US-1.** As a researcher on a Person page, when I open the Why on the birth date row, I see every record of that date on the person's birth events, so that I can tell where the date comes from without leaving the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person promoted with one birth event whose date two records give (14 May 1817 from a register, 14 May 1817 from a census) | `GetConclusionDetail` returns `person_life.birth_date` with one entry keyed by that event, whose `field.property_key` is `date` and whose `field.outcomes` hold both records (the map's Check) |
| AC-1.2 | The same person on the Person page | The Birth date row shows the Why button; opened, its title is "Why “14 May 1817”" and it lists both records, primary ink, with no group line (LW-1 – LW-3) |
| AC-1.3 | A birth event with no `date` record and one `start_date` record of 1817 | The entry's field is `start_date` with that record |
| AC-1.4 | A person with two surviving birth events, the lead dated 1817 (two records) and the other 1818 (one record) | `birth_date` has two entries, the 1817 event first; the row's Why lists the 1817 records in primary ink under a group line naming that event, then the 1818 record in secondary ink under a group line naming the other (LW-4 (a), LW-5) |
| AC-1.5 | A person with no birth event | `birth_date` and `birth_place` are empty; the row shows no Why button and looks as today |
| AC-1.6 | A birth event whose only date record is denied, so the row is stated empty | `birth_date` has the event's `date` field with the denied record; the row is stated empty, shows the Why button, and the Why title is "Why “Birth date”" |

**US-2.** As a researcher, when I open the Why on the birth place row, I see the records naming the Place, so that I can check the place name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A birth event located at one Place with toponym records "York" (two) and "Eboracum" (one) | `birth_place` has one entry keyed by that Place, field `toponym`, with three outcomes; the row's Why lists them with no group line |
| AC-2.2 | A death event located at Montréal and at Paris (two Places that do not fold) | `death_place` has two entries, the one the row shows first; the Why lists each Place's names under its group line, the second Place's records in secondary ink |
| AC-2.3 | The same Place on two surviving birth events | `birth_place` lists that Place once |
| AC-2.4 | A Place that has a parent (York part of Yorkshire) | Only York's toponym records are listed; Yorkshire's are not |

**US-3.** As the engine, I keep the life records off every other read.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `GetConclusionDetail` on an Event or a Place | `person_life` is unset |
| AC-3.2 | `ListPersonHeaders` | Its `PersonHeader` and `LifeFacts` messages are unchanged on the wire |

## 3. Scope

| In | Covered by |
| --- | --- |
| `conclusionheaders.PersonLifeOwners`: which entity and property each life row reads, every surviving event, lead first | STEP-1, AC-1.3, AC-1.4, AC-2.2 – AC-2.4 |
| `conclusiondetails.PersonLife` and its per-owner field composition | STEP-2, AC-1.1, AC-1.6, AC-2.1 |
| `ConclusionOwnedField`, `PersonLifeRecords`, `ConclusionDetail.person_life`; handler fills it for a Person | STEP-3, AC-1.1, AC-3.1, AC-3.2 |
| Swift store types and GoStore mapping | STEP-4 |
| The four life rows carry records, a Why title, owner group lines and the secondary ink rule (S13-D12 LW-1 – LW-7, LW-W1) | STEP-5, STEP-6, AC-1.2, AC-1.4 – AC-1.6, AC-2.1, AC-2.2 |
| The "its page owns the Why" comments (Go, proto, Swift) | STEP-7 |

| Out | Home |
| --- | --- |
| The provenance control and sheet | S13-11 |
| The Event page's derived place row carrying records | Not planned this spike (map: "can take the same path later"; Event keeps its Why until it adopts sections) |
| Records of the event-to-Place location link and of the parent chain's names | Not planned (user, 2026-10-30) |
| Any change to `LifeFacts`, `PersonHeader` or list reads | Not planned (working rule) |
| A shared Go test fixture for bridges across packages | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Detail composer | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity`, `loadProperties`, `attachValues`, `loadOutcomes`, `state` | Composes one handle's `Field`s (values and `Outcome`s) from the graph cache and `auto_reconciler_outcomes`. `ForEntity`'s closing loop sets each field's `State` and sorts `Outcomes` (by value rank, none last, then Observation id). Nothing composes another handle's field |
| Life facts | `core/database/conclusionheaders/conclusionheaders.go` `LifeFacts`, `HeaderPlace` | `LifeFacts.Event` is the lead event only; comment "Its page owns the date's Why". `HeaderPlace` comment: "whose page owns the names' Why" |
| Life walk | `core/database/conclusionheaders/compose.go` `onePerson`, `lifePick`, `lifeFrom`, `leadLife`, `lifeBefore`, `eventPlaces`, `composePersons` | `onePerson` follows `hops.EventsOfSubject`, keeps events whose `event_type` kept term is `birth` or `death`, and builds a `lifePick` each (`lifeFrom` reads the kept `date`, else `start_date`). `leadLife` picks the lead by `lifeBefore` and folds its Places with `eventPlaces` (toponyms, folded by a `placeGraph` built in `composePersons` from every event's location links). Which date property was read is not kept |
| Handler | `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`, `conclusionDetailProto`, `attachHeader`, `attachPictures` | Maps `Detail` to proto; `attachHeader` reads the kind's header (Place case also calls `conclusionheaders.AttachPlaceRelationships`) |
| Proto | `api/proto/engine.proto` `ConclusionDetail` (fields 1 – 7), `ConclusionField`, `LifeFacts`, `HeaderPlace` | `HeaderPlace.entity` comment "its page owns the names' Why" |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionDetail`, `CatalogHeaderPlace`; `macos/App/Platform/GoStore.swift` `getConclusionDetail` | Fields mapped inline in `getConclusionDetail`; `CatalogHeaderPlace.entity` comment "its page owns the names' Why" |
| Life rows | `macos/App/Features/Conclusions/PersonDetailContent.swift` `rows(_:header:locale:orders:)` | Builds the four life rows with `ReconciledValueRowModel.derived(…, opens:)`, no records |
| Row model | `macos/App/Features/Conclusions/ReconciledValueRowModel.swift` `Opens`, `opens`, `derived` | `derived` sets `whyTitle: ""`, `records: []`; comments say the Why stays on the Event or Place |
| Row view | `macos/App/Features/Conclusions/ReconciledValueRow.swift` `disclosures` | Shows the Why `PVDisclosureButton` when `model.records` is non-empty, before the open `PVIconButton` |
| Why table | `macos/App/Features/Conclusions/ReconciliationReasoningView.swift`; `ReconciliationRecord.swift` `ReconciliationRecord` (`counted` is a `var`) | One flat grid of records; no grouping |
| Tests | `core/database/conclusiondetails/conclusiondetails_test.go` (fixture: `subjectOn`, `cite`, `promote`, `field`); `core/database/conclusionheaders/walks_test.go` + `conclusionheaders_test.go` (fixture: `participation`, `location`, `event`, `place`, `promoteSubject`, `pointYear`); `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailCarriesItsHeader`; `macos/ProvenenciaTests/PersonDetailTests.swift` `rowsAreNameThenLifeThenRecordedFields` (expects the life rows' `records.isEmpty`) | |

**Precedent.** `attachHeader`'s Place case: a page-only read (`conclusionheaders.AttachPlaceRelationships`) attached in the handler beside the header, never on the list header. Copy that placement for `person_life`. For the per-owner composition, copy `ForEntity`'s own sequence (`loadProperties` → `attachValues` → `loadOutcomes` → state and sort), narrowed to one field.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The date row carries the date field of **every** surviving birth (or death) event, ordered by `lifeBefore` (the lead first). The place row carries the `toponym` field of every Place of those events: for each event in that order, `eventPlaces(g, v, pg, event)` (folded at that event's date), appending each returned `HeaderPlace.Entity` whose id is not yet listed. "Surviving" means exactly the events `onePerson` keeps today (the shared `lifePicks`, D-4): a nil node is skipped as there; no other filter is added | User, 2026-10-30 (explain the Mixed badge; each Place's toponyms). The working rule was updated in the planning commit | Lead event only — a Mixed row from a second event would show none of the disagreement |
| D-2 | An event's owned property is the one its value was read from: `date` when `lifeFrom` found a kept `date`, else `start_date` when it found a kept `start_date`, else `date` | Same rule as the lead value, kept in Go beside `lifeFrom` | Sending both fields and letting Swift choose — a client copy of an engine rule (maintainability Known traps) |
| D-3 | Parent-chain Places and the location bridge's records are not carried | User, 2026-10-30 | — |
| D-4 | The rule "which entity and property each life row reads" lives in `conclusionheaders` as `PersonLifeOwners` *(new)*, sharing the event walk with `onePerson` through an extracted `lifePicks` *(new)*. `lifePick` gains `dateProp string` *(new)*, set in `lifeFrom` | `conclusionheaders` owns `lifeFrom`, `lifeBefore` and `eventPlaces`; one walk, two callers | Fields on `LifeFacts` — the working rule keeps life records off lists; a re-walk in `conclusiondetails` — a second copy of the walk |
| D-5 | The records are composed in `conclusiondetails` by `PersonLife` *(new)*, which calls `conclusionheaders.PersonLifeOwners` and a private `ownedField` *(new)* per owner. `conclusiondetails` imports `conclusionheaders` (nothing in `core` imports `conclusiondetails`; no cycle) | `conclusiondetails` owns `Field` and `Outcome`; the handler stays thin | Composing in the handler — domain logic in a handler (Known traps) |
| D-6 | `ownedField(g, q, o)` runs, in order: `n := g.Node(o.Entity.ID)` (nil or `n.Merged` → no entry); `v := g.Vocabulary()`; `fields, _ := loadProperties(g, q, o.Entity.ID, n.SubjectTypeID)`; it keeps the one `f` with `f.PropertyKey == o.PropertyKey` and `v.PropertiesByID()[string(f.PropertyID)].Origin == vocabkey.OriginProvenencia` (none → no entry; `Field` gains no origin); `byProp := map[string]*Field{string(f.PropertyID): f}`; `attachValues(g, n, byProp)`; `loadOutcomes(q, o.Entity.ID, byProp, nil)`; `finish(f)`. `ForEntity`'s state-and-sort loop body moves into `finish(f *Field)` *(new)*, called by both. `TestForEntityQueryCountIsConstant` is left as it is (`ForEntity` is unchanged) | Same output shape as a field on its own page; a user property that shares the key is not the product's | A new SQL read — the existing `sqlOutcomes` already serves one entity |
| D-7 | Wire: `ConclusionOwnedField` *(new)* {`CanonicalEntity entity = 1; ConclusionField field = 2;`}, `PersonLifeRecords` *(new)* {`repeated ConclusionOwnedField birth_date = 1; birth_place = 2; death_date = 3; death_place = 4;`}, `ConclusionDetail.person_life = 8` set only for a Person. Each entry is keyed by its owning entity | Map: keyed by owning entity, on `ConclusionDetail`, no new RPC; typed rows so Swift does not guess which entry is which | A flat `repeated ConclusionOwnedField` — Swift would have to know the competing event ids, which only Go has |
| D-8 | The handler fills `person_life` in `attachHeader`'s Person case, after the header is set (when the header is absent the case already returns, and `person_life` stays unset), through `conclusiondetails.PersonLife` and `personLifeProto(l conclusiondetails.Life) *engine.PersonLifeRecords` *(new)*, which maps each `Owned` to `&engine.ConclusionOwnedField{Entity: canonicalEntityProto(o.Entity), Field: conclusionFieldProto(o.Field)}`. `conclusionDetailProto`'s per-field body moves into `conclusionFieldProto(f conclusiondetails.Field) *engine.ConclusionField` *(new)*, reused for owned fields | Precedent `AttachPlaceRelationships`; one field mapper | — |
| D-9 | Swift: `CatalogOwnedField` *(new)* {`entity: CatalogCanonicalEntity`, `field: CatalogConclusionField`}, `CatalogPersonLife` *(new)* {`birthDate`, `birthPlace`, `deathDate`, `deathPlace: [CatalogOwnedField] = []`}, `CatalogConclusionDetail.personLife: CatalogPersonLife = CatalogPersonLife()`. GoStore's inline field mapping moves into `private static func mapConclusionField(_:)` *(new)*, reused | Default empty keeps every existing initializer call compiling | Optional — every reader would unwrap to empty |
| D-10 | `ReconciledValueRowModel.derived` gains `owned: [CatalogOwnedField] = []`, `locale: Locale = .autoupdatingCurrent` and `orders: NameFormatOrders = NameFormatOrders()`; `PersonDetailContent.rows` passes its own. Records are `owned` in order, each field's outcomes mapped with `ReconciliationRecord(outcome:in:locale:orders:)`. Records of every entry after the first get `counted = false` (LW-5). When there are two or more entries, each record's `owner` *(new, `ReconciliationRecord.OwnerLine?`)* is its entry's group line; with one entry it stays nil (LW-3). `whyTitle` is `L10n.Conclusions.whyTitle(shown ?? label)` when records exist, else `""` | The first entry is the owner the row shows (lead event; the lead's first Place, which `DerivedPlace.name` shows) | A separate records property for derived rows — a second Why path |
| D-11 | `struct ReconciliationRecord.OwnerLine: Equatable { var id: String; var text: String; var spoken: String }` *(new)*. `id` is the owner's entity id. The owner's title is the owned field's first displayed value as text (`ReconciledValueDisplay.string(for:locale:orders:)`: "14 May 1817" for an event's date, "Montréal" for a Place's toponym), else `entity.label` trimmed, else empty. `text` is `L10n.Conclusions.whyOwner(title, ref:)` *(new)* "%1$@ · %2$@" ("14 May 1817 · EVT-4KX2M"), or the ref alone when the title is empty; `spoken` is `L10n.Conclusions.a11yList(title, rest: ref)`, or the ref alone. Built by `ReconciledValueRowModel.ownerLine(_:locale:orders:)` *(new, private static)*. This is the writer's copy call; the board may reword it | One rule for events and Places: each owner is named by the value it holds, which is what tells two birth events apart. Numbered format required for two arguments by `scripts/check-localizable-xcstrings.py` (S13-10 Landed note) | The event's label — usually empty, so the line would be the ref alone; `EventTitleDisplay` — needs an `EventHeader`, which owned fields do not carry; "Birth" from the row's side — the same word on every group, telling nothing apart |
| D-12 | **Pending the board (LW-4).** `ReconciliationReasoningView` draws a group line row before a record whose `owner` is non-nil and whose `owner.id` differs from the previous record's: `Text(verbatim: owner.text)` in `pvMicroCaps()`, `PVColor.textMuted`, spanning the grid's columns (`.gridCellColumns(4)`), `.accessibilityAddTraits(.isHeader)`, `.accessibilityLabel(owner.spoken)`. If the board picks LW-4 (b) instead, this Decision and AC-1.4 / AC-2.2 are rewritten before the plan is Planned | Brief recommendation (a) | — |

**Rules this plan bends.** None open. D-1 changed the working rule "A derived row carries its owning entity's records" from the lead event to every surviving event (user, 2026-10-30); the rule and the S13-06 scope entry were updated in the planning commit.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `conclusionheaders.FieldOwner`, `LifeOwners`, `PersonLifeOwners` *(new)* | type, func | `onePerson`, `leadLife` | Extends them through `lifePicks`; `leadLife` keeps returning the lead only for headers |
| `lifePicks`, `lifePick.dateProp` *(new)* | func, field | `onePerson`'s loop | Extracted from it, not copied |
| `conclusiondetails.Owned` (its `Entity` is the owner's, from `conclusionheaders.FieldOwner`; no second `entityOf`), `Life`, `PersonLife`, `ownedField`, `finish` *(new)* | type, func | `ForEntity` | Reuses `loadProperties`, `attachValues`, `loadOutcomes`; `finish` is `ForEntity`'s loop body moved |
| `conclusiondetails/export_test.go` `OwnedField` *(new)* | test helper | `conclusionheaders/export_test.go` | Same pattern, its own package's internals |
| `ConclusionOwnedField`, `PersonLifeRecords`, `ConclusionDetail.person_life` *(new)* | proto | `ConclusionField`, `ConclusionDetail.header` | Wraps `ConclusionField`; no field duplicated |
| `conclusionFieldProto`, `personLifeProto` *(new)* | func | `conclusionDetailProto` loop, `canonicalEntityProto` | Moved, not copied; `personLifeProto` composes the two |
| `CatalogOwnedField`, `CatalogPersonLife`, `mapConclusionField` *(new)* | Swift type, func | `CatalogConclusionField`, GoStore inline mapping | Wrap and move |
| `ReconciliationRecord.OwnerLine`, `.owner`, `ReconciledValueRowModel.ownerLine` *(new)* | type, property, func | `Opens.event`, `Opens.place` titles | `Opens.place` needs a `CatalogHeaderPlace`; owned fields carry the toponym field instead |
| `L10n.Conclusions.whyOwner` *(new)* | key `conclusions.why.owner` | `pageSectionLeaf` ("%1$@ · %2$@") | Different meaning and translator comment |

**Builds on.** `ForEntity`'s composition, the `attachHeader` page-only attach, the existing `ReconciledValueRow` Why disclosure and `ReconciliationReasoningView`.
**Deletes.** The "its page owns the Why" comment text (STEP-7). No code becomes dead: `leadLife` and `ForEntity` keep their callers.
**Next adds.** The Event page's place row takes a `ConclusionOwnedField` list the same way (a new field on `ConclusionDetail` or an `EventPlaceRecords` message). S13-11 reads `ReconciledValueRowModel.records` and each record's `owner` for the sheet.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A derived row with records | `ReconciledValueRow` Why disclosure, `ReconciliationReasoningView` | Already handles any record list; the Source link opens the composer by the record's own ids |
| A record from an entity other than the page's handle | `ReconciliationRecord.composerLocation` | Uses the outcome's Source, graph, Subject and Citation ids, never the page entity |
| An owner deleted or merged between the header read and the field read | `ownedField` | Skipped (D-6); both reads run inside one `withProjectCatalog` call, so the catalog session serializes them |

### Scale

No write. Per Person page open: one `PersonLifeOwners` walk (graph cache, the same hops `onePerson` runs) and, per owner, `loadProperties` (one SQL), `loadOutcomes` (one SQL plus date and name lookups). Owners are 2 × (events + their Places); a typical person has 1 birth and 1 death event with 1 Place each: 4 owners, about 12 SQL reads added to a detail that already runs several. Lists are untouched (AC-3.2).

### Edges

- **Assumptions.** `lifeBefore` is a strict order (dated before undated, then sort key, then ref): sorting picks by it puts the lead first, as `leadLife` does. `eventPlaces` folds per event at that event's date; Places are deduped by entity id across events. Tested by T-1.
- **Neighbors.** One event (no group lines) vs two (group lines); `date` vs `start_date` (D-2); an event with neither: owner `date`, which has records only if a non-kept date record exists (AC-1.6). T-1, T-2, T-5.
- **Removed inputs.** An owner merged or gone: skipped. A record denied or outvoted: still listed with its outcome (it is in `auto_reconciler_outcomes`).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** Events per side and Places per event come from catalog links; each owner costs two SQL reads. No bound is added: the same counts already drive `composePersons`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionheaders` | `type FieldOwner struct { Entity canonicalentities.Entity; PropertyKey string }` *(new)*; `type LifeOwners struct { Dates, Places []FieldOwner }` *(new)*; `func PersonLifeOwners(c *database.Catalog, personID []byte) (birth, death LifeOwners, err error)` *(new)*: nil graph, nil or merged node → zero values, nil. `func lifePicks(g *graphcache.Graph, v vocab, n *graphcache.Node) (births, deaths []lifePick, err error)` *(new)*, `onePerson` calls it. `PersonLifeOwners`: (1) `lifePicks` on the person node; (2) collect `follow(g, pick.event.ID, hops.PlacesOfEvent)` neighbors of every pick of both sides and `pg, err := buildPlaceGraph(g, v, placeIDs)`, as `composePersons` does; (3) sort each side with `sort.SliceStable` by `lifeBefore`; (4) `Dates` = one `FieldOwner{Entity: entityOf(pick.event), PropertyKey: pick.dateProp}` per pick; `Places` per D-1 with `PropertyKey: vocabkey.PropertyToponym` |
| `core/database/conclusiondetails` | `type Owned struct { Entity canonicalentities.Entity; Field Field }` *(new)*; `type Life struct { BirthDate, BirthPlace, DeathDate, DeathPlace []Owned }` *(new)*; `func PersonLife(c *database.Catalog, personID []byte) (Life, error)` *(new)*; `func ownedField(g *graphcache.Graph, q Querier, o conclusionheaders.FieldOwner) (Field, bool, error)` *(new)*; `func finish(f *Field)` *(new)*. Package doc gains one sentence: a Person's life rows read other handles' fields through `PersonLife` |

Errors: none new; graph and SQL errors propagate.

### FFI

`api/proto/engine.proto`: messages `ConclusionOwnedField` and `PersonLifeRecords` *(new)* per D-7, beside `ConclusionField`; `ConclusionDetail` gains `PersonLifeRecords person_life = 8;` with the comment "set for a Person only: the records behind its four life rows, keyed by the event or Place that owns each (S13-06)". Handler `api/ffi/handlers/conclusion_detail.go` per D-8. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Life row | `ReconciledValueRow` (unchanged) | `macos/App/Features/Conclusions/ReconciledValueRow.swift` | Shows Why because the model now has records |
| Life row model | `ReconciledValueRowModel.derived(…, owned:)` | `macos/App/Features/Conclusions/ReconciledValueRowModel.swift` | D-10, D-11 |
| Group line | Row inside `ReconciliationReasoningView` | `macos/App/Features/Conclusions/ReconciliationReasoningView.swift` | D-12; no new component |
| Disclosure, open arrow, badge, card | `PVDisclosureButton`, `PVIconButton`, `PVBadge`, `PVCard` | `macos/App/DesignSystem/Components/…` | Unchanged |

`PersonDetailContent.rows` gains `life: CatalogPersonLife = CatalogPersonLife()`; `init(detail:)` passes `detail.personLife`, and each life row passes its list (`birthDate`, `birthPlace`, `deathDate`, `deathPlace`, in the `life` table's order). States: no entries → no Why (as today); entries with no outcomes cannot occur (D-6). Accessibility: LW-7 per D-12. Frames: copied from the board at handoff.

### Strings

L10n keys *(new)*: `L10n.Conclusions.whyOwner(_ title: String, ref: String)` → `conclusions.why.owner` = "%1$@ · %2$@", comment "Group line in a Why list naming the record the following rows belong to; arguments are its title and ref". Removed: none.

## 7. Steps

**STEP-1 — Life owners in Go**
Files: `core/database/conclusionheaders/conclusionheaders.go`, `compose.go`, `placechains_test.go`.
Do: D-2, D-4 and the `conclusionheaders` row of §6 Go core; add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/...` passes, existing header tests unchanged.

**STEP-2 — Owned fields in conclusiondetails**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `conclusiondetails_test.go`, `export_test.go` *(new)*.
Do: D-5, D-6; add T-2 (fixture: an event Subject via `subjectOn(src, "event")` with date Observations; look up `date` with `properties.Lookup` as `newFixture` does for `name`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/...` passes with `TestForEntityQueryCountIsConstant` unchanged.

**STEP-3 — Wire and handler**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/conclusion_detail.go`, `conclusion_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-7, D-8; `./scripts/generate-proto.sh`; add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-4 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`.
Do: D-9.
Check: the app builds (`xcodebuild test` as `scripts/check-pr.sh` runs it).

**STEP-5 — Life rows carry records**
Files: `macos/App/Features/Conclusions/ReconciledValueRowModel.swift`, `ReconciliationRecord.swift`, `PersonDetailContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-10, D-11; `rowsAreNameThenLifeThenRecordedFields` stays as it is (its detail has the default empty `personLife`, so the life rows still have no records); add T-5 – T-7.
Check: Xcode tests pass; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-6 — Group lines in the Why**
Files: `macos/App/Features/Conclusions/ReconciliationReasoningView.swift`. Skill: `add-ui-component`.
Do: D-12, matched to the agreed board's frames 3 and 4.
Check: manual check §10 steps 3 – 4.

**STEP-7 — Comments**
Files: `core/database/conclusionheaders/conclusionheaders.go` (`LifeFacts`: "Its page owns the date's Why" → "A Person page reads its date's records through conclusiondetails.PersonLife"; `HeaderPlace`: "whose page owns the names' Why" → "whose toponym records a Person page reads through conclusiondetails.PersonLife"), `api/proto/engine.proto` (`HeaderPlace.entity`: "entity is the Place; a Person page's life rows carry its toponym records in ConclusionDetail.person_life"), `macos/App/Platform/GenealogyStore.swift` (`CatalogHeaderPlace.entity`, same sense; `CatalogLifeFacts.event`: "Its page owns the date's Why" → "A Person page reads its date's records from `CatalogConclusionDetail.personLife`"), `ReconciledValueRowModel.swift` (`Opens`: "The handle a derived value is read from"; `opens`: "A derived row opens the Event or Place it is read from. Nil otherwise."; `derived`'s doc: records are the owning handles' fields, the first the one shown).
Do: as listed; regenerate proto.
Check: `rg -n "owns[^.]*Why|Why stays on the Event or Place" core api macos/App -g '!**/Generated/**' -g '!api/proto/engine/**'` finds nothing. (`EventDetailContent.swift`'s "No Why: that stays on the Place" is still true and does not match.)

**STEP-8 — Docs**
Files: `../completed.md`, `../deployment-plan.md` (map row status), this plan (status, Landed note), `../design/` (S13-D12 → `design/archive/`, `design/README.md`).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show where a birth date comes from on the Person page."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionheaders/placechains_test.go` `TestPersonLifeOwners` *(new)*, table-driven. Build each case as `TestBirthLocationsFoldWhenPartOf` does: `placeSubject`, `placeRel(child, parent, propertyterms.KeyPartOf, nil, nil)`, `subjectstest.On(…, "event", "")`, `cite` with `event_type` and `date` (`pointYear`) or `start_date` (`observations.Input{PropertyID: f.prop("start_date").ID, Date: pointYear(1817)}`), `participation(person, event, "subject")`, `location`, `promoteSubject` on each. Two Places that do not fold are two `location`s to Places with no `placeRel` between them | (a) one birth event with `date` at York → `birth.Dates` = [{event, `date`}], `birth.Places` = [{York, `toponym`}]; (b) `start_date` only → `start_date`; (c) birth events 1818 and 1817 → 1817 first; (d) death event at Montréal and Paris → two Places, fold order; (e) York on two birth events → once; (f) York part of Yorkshire → York only; (g) no events → zero; (h) unknown id → zero, nil | AC-1.3, AC-1.4, AC-2.2, AC-2.3, AC-2.4, D-1, D-2 |
| T-2 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestOwnedField` *(new)*, table-driven, through `OwnedField(c *database.Catalog, o conclusionheaders.FieldOwner) (Field, bool, error)` in `core/database/conclusiondetails/export_test.go` *(new, package `conclusiondetails`; it opens `c.DB()` and calls `ownedField(c.Graph(), db, o)`, as `conclusionheaders/export_test.go` exposes its internals)* | promoted event with two `date` records → field `date`, two outcomes sorted as `ForEntity` sorts, state merged; owner key with no outcomes → skipped; merged node → skipped | AC-1.1, AC-1.6, D-6 |
| T-3 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailCarriesPersonLife` *(new)*. Fixture over the FFI: `connectFixture` (person, event, artifact, property ids); `CreateCitedBridge` participation as `TestCreateCitedBridge` sends it but with the role term `subject` (`ListPropertyTerms` on the role property); two `CreateCitationWithObservations` calls on the event, the first with `event_type` = `birth` and a `date` 14 May 1817, the second with the same `date`; `promoteOK` on the person and on the event | Person + birth event + two date records → `person_life.birth_date[0]` keyed by the event, field `date`, two outcomes | AC-1.1 (map Check) |
| T-4 | Go | `TestGetConclusionDetailCarriesItsHeader` (extend each case) | Event and Place → `GetPersonLife() == nil` | AC-3.1 |
| T-5 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeRowsCarryTheirOwnersRecords` *(new)* | `personLife.birthDate` one entry, two outcomes → Birth date row `records.count == 2`, `whyTitle == "Why “14 May 1817”"`, every `owner == nil`; empty `deathDate` → Death date row records empty, `whyTitle == ""` | AC-1.2, AC-1.5, D-10 |
| T-6 | Swift | same file `competingOwnersAreGroupedAndSecondary` *(new)* | two `birthDate` entries → records in entry order, second entry's `counted == false`, `owner.text` "14 May 1817 · EVT-4KX2M" and "2 Jun 1818 · EVT-…" per D-11, `owner.spoken` "14 May 1817, EVT-4KX2M"; two `deathPlace` entries → Place titles from the toponym's displayed value | AC-1.4, AC-2.2, D-10, D-11 |
| T-7 | Swift | same file `anEmptyLifeRowWithRecordsTitlesByLabel` *(new)* | `birthDate` entry with one denied outcome, no displayed value → row `lead == nil`, records 1, `whyTitle == "Why “Birth date”"` | AC-1.6 |

AC-2.1 and AC-3.2 are covered by T-1 (a) plus T-3's handler path, and by the manual check and the unchanged `TestListPersonHeaders` respectively.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `conclusionheaders` header and display-cache tests (STEP-1 extraction), `PersonDetailTests`, `ReconciledValueDisplayTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (STEP-8)
- [ ] UI matches the S13-D12 board inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief S13-D12 moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person promoted with a birth event dated by two records and located at York (two toponym records). On the Person page, Vitals, open the Birth date Why: two records, title "Why “14 May 1817”", no group line. Click a Source: the composer opens on that Observation; Back returns.
2. Open the Birth place Why: York's records, no group line.
3. Add a second birth event (1818) for the same person from another Source and promote it. The Birth date row shows Mixed; its Why lists both events under group lines, the 1818 record in secondary ink.
4. Give the death event two Places that do not fold; the Death place Why lists both Places' names under group lines.
5. Resize the window across 1600pt with a Why open: it stays open and in place (LW-W1). Open an Event and a Place page: they look as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S13-24 rewrites `sqlOutcomes`, the outcome scan and `reconcilerOutcomeProto` in the same files | This PR does not edit those lines; only `ForEntity`'s closing loop (→ `finish`) and `conclusionDetailProto`'s loop (→ `conclusionFieldProto`) move. Whichever lands second rebases |
| Extracting `lifePicks` changes header output | Existing `conclusionheaders` tests must pass unchanged in STEP-1 |

**Stop and report if:**

- `onePerson` no longer walks `hops.EventsOfSubject` with the `birth` / `death` term filter, or `lifeFrom` no longer reads `date` then `start_date`.
- A birth event's date records are not in `auto_reconciler_outcomes` under the event's entity id (T-3 finds no outcomes for a promoted event with two date records).
- The participation bridge does not create the canonical link in T-3's fixture with `PromoteSubject` on each end (the event is not found from the person).
- Importing `conclusionheaders` from `conclusiondetails` creates a cycle (`go list -deps ./core/database/conclusionheaders | rg conclusiondetails` finds it).
- The agreed board's LW-4 option or frames contradict D-10 – D-12 or AC-1.4 / AC-2.2.

## 12. Open questions

1. **LW-4 grouping option** — decided by the S13-D12 board. Recommendation: (a), as D-12 drafts.

## 13. PR

**Title:** S13-06 — Derived rows carry their records
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.md`.
