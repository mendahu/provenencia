# S16-19 — Summary composers read the place entries and keep what readers re-derive

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2c — Walks
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine refactor; every summary row, life row, Timeline row and title reads as today)
**Board:** —
**Depends on:** S16-18 (landed [#817](https://github.com/mendahu/provenencia/pull/817)), S16-01 (landed [#756](https://github.com/mendahu/provenencia/pull/756)), S16-02 (landed [#764](https://github.com/mendahu/provenencia/pull/764))
**Blocks:** S16-24 (reports the inputs of the Person reads this PR rewrites), S15-28 (composes from the settled life-row owners)
**Branch:** `s16-19-summary-rows-carry-owners`
**Version bump:** none — no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the graph cache, its displays and Verify)
**Rules that bind:** `.cursor/rules/go-tests.mdc` (table-driven tests)
**Context:** [design § Results per key](../cache-architecture.md#results-per-key-for-the-walks), [design § A cache a composer reads hands its inputs to the row](../cache-architecture.md#a-cache-a-composer-reads-hands-its-inputs-to-the-row), [`docs/catalog-caches.md`](../../../catalog-caches.md) (`displays`), [Spike 14 summary inputs](../../archive/spike-14/summary-inputs.md), [S16-18 plan and Landed note](S16-18-place-entries.md#landed), working rules: "The point of this batch is simpler code", "Cache entities and results per key; compose answers", "Every cache names the signal that updates it, and updates exactly", "Apply states its fan-out", "A write's one-shot graph fills only what it reads", "A cache a composer reads hands its inputs to the row", "Orders are defined and results do not move", "Memory is budgeted", "Batch 1's caps stay where S16-01 puts them", "No new persisted projections, and no stored answer without a measurement", "Measure growth"
**Planned:** 2026-11-13 against `main` @ `7455758`
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

Three Person page reads re-derive what the summary composers already walked. The event and person composers still load each place's whole node to read its names, although S16-18 put every place's names in its place entry. The Person page's life rows (`PersonDerivedRows`) walk the person's birth and death events, their locations and their place chains a second time through `PersonLifeOwners`, although the person's summary row was composed from the same walk. And the Timeline follows every event's participants itself, although it already loads each event's summary row. This PR makes the rows carry what those readers need: an event's places come from the place entries; a person's row keeps the owners behind its birth and death rows (each event and the Property its date came from, each event–place link and its place); and an event's row keeps every participant with their roles, from one `Follow` of the participants hop. `PersonLifeOwners`, `lifeOwners`, `eventPlaceGraph`, the Timeline's per-event walk and its collapse rule's second home go, and `EventSummary`'s subjects and the Timeline's participants become one Go type, `conclusionsummaries.EventParticipant` (structural debt A-33).

**Why now.** Bar 13 ("Walks are lookups") and the batch rule "The point of this batch is simpler code": after S16-18, an event's places and a person's life owners are the last readers that walk past the cache. S16-24 reports the inputs of the Person reads, so it lands on the reads this PR leaves.

**What stays.** Every value the app shows: event titles, places and chains, the life rows and their owners in today's order, the Timeline's rows, roles and participant order. The wire (`EventSummary.subjects`, `PersonEvent.participants`) and Swift. The rows' recorded input keys for persons and places (S14-40's tests hold them); event rows record more keys (D-7). S16-01's `eventLocations` cap of 20 places per event, `summaryChainDepth`, the input cap and the place entries.

## 2. User stories

**US-1.** As the engine, an event's or a person's summary row reads its places' names from the place entries, so composing an event at a city does not load the city's node with its 6,000 location links.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | City (one toponym "City") with three events located at it; the graph is dropped. `EventsByIDs(c, [event1])` | The row's places are one `ChainedPlace` with `Names` `["City"]`; afterwards `g.Cached(city)` is false |
| AC-1.2 | Pat born at Town `part_of` City; the graph is dropped. `PersonsByIDs(c, [pat])` | `Birth.Places` is Town with parents `["City"]`; `g.Cached(town)` and `g.Cached(city)` are false |
| AC-1.3 | The same catalogs composed with `RowsForSearch` | Each row's `Inputs` equal the keys it records composed alone (`CheckMatchIsolated`), and the person and place rows' inputs equal those at the planning base |

**US-2.** As a researcher, the Person page's birth and death rows list the same records as today, read from the person's summary row instead of a second walk.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Each case of today's `TestPersonLifeOwners` table (two births, one dated 1820 and one undated; locations folding on one chain; one event filed at a place twice) | `PersonsByIDs` returns a row whose `Birth.Owners` and `Death.Owners` equal the case's expected date owners and place owners, in the same order: date owners lead first then date order, one per event; place owners one per event–place link, each with its place's own toponyms and no parents |
| AC-2.2 | Pat with a birth at Town; `PersonDerivedRows(c, pat)` | The same `[]DerivedRow` as at the planning base (`TestPersonDerivedRows` unchanged) |
| AC-2.3 | Rename the birth's place; commit; `PersonDerivedRows` again | The birth place owner's `Place.Names` is the new toponym (the stored row was dropped through its inputs) |
| AC-2.4 | A birth event filed at 21 distinct places (`TestCraftedEventPlaces`) | The row's `Birth.Owners.Places` has 20 owners; a place filed 21 times gives 21 owners |
| AC-2.5 | `PersonDerivedRows` for a merged person, a non-person id and an unknown id | No rows, no error |

**US-3.** As the engine, the Timeline reads each event's participants from the event rows it already loads, and the event row's subjects and the Timeline's participants are one type.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Pat subject of B1; Ned subject and Pat witness of burial U. `EventsByIDs(c, [U])` | `Participants` is Ned `[subject]`, Pat `[witness]` in link order; `Subjects` is Ned only |
| AC-3.2 | Ned holds subject and witness on one event (two links) | Ned is one participant with roles `[subject, witness]` in link order |
| AC-3.3 | Every case of `TestPersonEvents`, `TestPersonEventsLifeline` and `TestParticipantOrder` | Unchanged results |
| AC-3.4 | The event rows are stored (a first `PersonEvents`), then `PersonEvents(c, pat)` with an input recorder on the graph | No key is recorded at an event's `participation` bridge end: the Timeline follows no event's participants |
| AC-3.5 | A participant whose person has no accepted member (`graphcachetest.Canon.Provisional`) | The event row lists them (as today's subjects would); the Timeline does not (as today) |
| AC-3.6 | `TestPersonEventsQueryCount` AC-2.2 (1 event × 1 guest vs 12 events × 3 guests and one of 41) | Equal query counts |

## 3. Scope

| In | Covered by |
| --- | --- |
| `eventPlaces` reads each location's place shell from the place graph (the entries) instead of `InputScope.Node` | STEP-2, AC-1.1, AC-1.3 |
| `eventPlaceGraph` goes; `composePersons` calls `buildPlaceGraph` over `placesOfEvents`; `composeEvents` stops loading place nodes | STEP-2, AC-1.1, AC-1.2 |
| `LifeFacts.Owners` *(new)*: the date and place owners of a birth or death, composed in the same pass as the lead's facts; `leadLife` and `lifeOwners` become one function | STEP-3, AC-2.1, AC-2.4 |
| `PersonLifeOwners` goes; `PersonDerivedRows` reads the owners from the person's summary row | STEP-3, AC-2.2, AC-2.3, AC-2.5 |
| `FieldOwner.Entity` becomes `FieldOwner.ID`; `PersonDerivedRows` builds each owner's entity from the node it loads | STEP-3, AC-2.2 |
| `rolesByNeighbor` moves to `conclusionsummaries.RolesByNeighbor`, the one collapse rule for the Timeline's own events and an event's participants | STEP-4, AC-3.2 |
| `conclusionsummaries.EventParticipant` *(new)* replaces `EventSubject` and `persondetail.EventParticipant`; an event row keeps `Participants` from one `Follow` of `hops.ParticipantsOfEvent`, and `Subjects` from the same links | STEP-5, AC-3.1, AC-3.2, AC-3.5 |
| The Timeline reads participants from the event rows; its per-event `Follow` goes | STEP-6, AC-3.3, AC-3.4, AC-3.6 |
| `BenchmarkCacheSizes`' `displays` figure fills stored rows, and the PR records it before and after | STEP-1, STEP-7 |

| Out | Home |
| --- | --- |
| An event row keeping the Property its date came from (the map's scope entry): no reader reads it once the person row keeps its date owners (D-6) | not planned; the map's scope entry and the design are amended in the planning commit |
| Sending each participant once on the wire (P-8) | [`person-page-performance.md`](../../../ideas/person-page-performance.md) |
| Promote's checks reading the place entries | S16-21 |
| The Person reads reporting their inputs on the wire | S16-24 |
| A wire or Swift change for the one participant type (the wire keeps `EventSubject` and `EventParticipant`) | not planned |
| A cap on participants per event | not planned — the Timeline reads every participant at base; the row holds what it read (§6 Edges) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Event places | `core/database/conclusionsummaries/compose.go` `eventPlaces(in, v, pg, ev, locations)` | Reads each location's place with `in.Node(l.Neighbor)` for `entityOf(n)` and `toponyms(n, …)`, then `pg.FoldLocations` at the event's date. The node read is what loads a city's 6,000 location links |
| Person place graph | `compose.go` `eventPlaceGraph(g, v, eventIDs)` | Checks each event has a node, follows `placesOfEvents(g.NoInputs(), …)`, calls `g.Nodes(placeIDs)` so `eventPlaces`' node reads batch (S16-18 Landed note), then `buildPlaceGraph`. Called by `composePersons` and `PersonLifeOwners` |
| Event composer | `compose.go` `composeEvents` | Follows `hops.SubjectsOfEvent` through a no-input scope to collect person ids, `eventLocations` through each row's scope, then `g.Nodes(personIDs + placeIDs)` and `buildPlaceGraph(g, v, placeIDs)`; `rec.addClosure(in, pg, neighborsOf(locations))` per row |
| Event row | `compose.go` `oneEvent` | Follows `hops.SubjectsOfEvent` through `in`; each distinct neighbor read with `in.Node`, skipped when nil, kept as `EventSubject{Entity, Name, NameValueCount}` |
| Event row type | `core/database/conclusionsummaries/events.go` `EventSummary.Subjects []EventSubject`, `EventSubject`, `eventTitle` | Subjects feed the title (`eventtitle.Parts.Subjects`) and the wire |
| Event finish | `compose.go` `finishEvent` | Fills the type term's label and the title; writes nothing through the stored row |
| Person composer | `compose.go` `composePersons`, `onePerson`, `leadLife`, `lifePicks`, `lifeFrom`, `eventDate` | `lifePicks` reads the birth and death events through `in`; `leadLife` picks the earliest by `dateKeyLess` and folds its places; `pg.addChains(in, rec, lifeEvents(…))` records every life event's locations and place closure |
| Life owners | `compose.go` `PersonLifeOwners(c, personID)`, `lifeOwners(in, v, pg, picks)` | A second walk for the Person page: `lifePicks` and `eventPlaceGraph` again, `sortByDateKey(picks)`, one date owner per event (`entityOf(event)`, `pick.dateProp`), then for each folded place every location link to it once, reading the association node for `entityOf(assoc)`; errors when a link has no association or the association node is missing |
| Owner types | `core/database/conclusionsummaries/conclusionsummaries.go` `LifeFacts`, `FieldOwner{Entity, PropertyID, Place}`, `LifeOwners{Dates, Places}` | `LifeFacts` has no owners |
| Derived rows | `core/database/conclusiondetails/conclusiondetails.go` `PersonDerivedRows` | Calls `PersonLifeOwners`, loads every owner's node, `fieldsOf` with `only`, `EventsByIDs` for the date owners, builds `OwnerField{Entity: o.Entity, Place: o.Place, Event, Field}`; drops an owner whose node is gone or merged (`fieldsOf`) |
| Entity helper | `conclusiondetails.go` `entityOf(n)` | Builds `canonicalentities.Entity` from a node, as `conclusionsummaries.entityOf` does |
| Timeline | `core/database/persondetail/timeline.go` `PersonEvents` | Follows the person's `hops.EventsOfParticipant` once (`mine`), then for each canonical event `g.Follow(event, hops.ParticipantsOfEvent)`, `g.Nodes(others)`, `rolesByNeighbor(links, label, keep)` with `keep` = not the person and `Canonical(personType)`, then `EventsByIDs`, `PersonsByIDs(people + self)`, builds `PersonEvent{Participants: []EventParticipant{Person, Roles}}` |
| Collapse rule | `timeline.go` `rolesByNeighbor`, `neighborRoles` | Links in order; skip `keep` false; dedupe neighbors first-seen; add each link's term once per neighbor when `label` knows it |
| Participant type | `timeline.go` `EventParticipant{Person []byte, Roles}`, `dropUnreadEvents` | Second event–person shape (A-33) |
| Participant order | `core/database/persondetail/participants.go` `sortParticipants(ps []EventParticipant, index)` | Rank by roles, then `index` (the person rows' list order) |
| Handler | `api/ffi/handlers/person_detail.go` `personEventProto` | Maps `r.Participants` by `p.Person` to `engine.EventParticipant{Person: personSummaryProto(…), Roles}` |
| Handler | `api/ffi/handlers/conclusion_summaries.go` `eventSummaryProto` | Maps `h.Subjects` to `engine.EventSubject` |
| Hops | `core/hops/hops.go` `SubjectsOfEvent`, `ParticipantsOfEvent` | Same bridge (`participation`) and end; `SubjectsOfEvent` keeps links whose `TermKey` is the filter's (`"subject"`). `core/database/graphfill/load.go` keeps using `SubjectsOfEvent` |
| Display check | `core/database/conclusionsummaries/display_register.go` `checkDisplay` | Verify compares a stored person or event row, finished, with a fresh one by `reflect.DeepEqual` |
| Benchmark | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` `measure("displays", …)` | Fills through `RowsForSearch`, which stores no row, so `docs/catalog-caches.md` reads 11 KB |
| Large fixture | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "inputs match isolated composition", "second part_of from the city" | Logs keys per kind and "first part_of from the city (untimed by the budget)" |
| Tests | `core/database/conclusionsummaries/placechains_test.go` `TestPersonLifeOwners`, `TestPersonDerivedRows`; `eventplaces_cap_test.go` `TestCraftedEventPlaces`, `TestEventPlacesCapInputs`, `placesWorld.owners`; `inputs_test.go` `TestInputsMatchIsolatedComposition`; `core/database/persondetail/timeline_test.go` `TestPersonEvents`, `TestPersonEventsQueryCount`, `TestPersonEventsLifeline`; `timeline_drop_test.go` `TestRolesByNeighbor`, `TestRolesByNeighborManyLinks`, `TestDropUnreadEvents`; `participants_test.go` `TestParticipantOrder`; `autoreconciler_test.go` rebuild-equals-upkeep | Hold today's results |

**Precedent.** S16-18 ([#817](https://github.com/mendahu/provenencia/pull/817)) moved the chain build onto the place entries while every row kept its inputs: copy how a composer reads an entry and lets `inputBudget.addClosure` hand the entry's keys to the row, and the shape of `TestInputsMatchIsolatedComposition` as the guard. S16-30 ([#812](https://github.com/mendahu/provenencia/pull/812)) removed `PersonDerivedRows`' second pass: copy its test of the derived rows staying equal. Do not copy S16-18's registry work: this PR adds no cache.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `eventPlaces(v vocab, pg *placeGraph, ev *graphcache.Node, locations []graphcache.Link) ([]ChainedPlace, error)` loses its `in` parameter. For each distinct `l.Neighbor` it reads `pg.shell(l.Neighbor)`; when `ok` it appends `ChainedPlace{Entity: shell.Entity, Names: shell.Names}`, else it skips the place (today's nil node). Every caller's `pg` was built with those places as seeds, so each has an entry | The entry holds exactly `entityOf(n)` and `toponyms(n, …)` from a narrow node (`fillPlaceEntries` → `placeShell`). The row still records each place's values: `addClosure` adds every seed's entry `inputs`, which start with `NodeKey(place)`, and a row recorded nothing when `rec` is nil or over the cap at base too | Keeping `in.Node(place)` beside the entry read — it loads the whole node, the read this PR exists to remove; a new entry read helper — `pg.shell` already is one |
| D-2 | `eventPlaceGraph` goes. `composePersons` builds its graph with `seeds, err := placesOfEvents(g.NoInputs(), pickIDs)` then `buildPlaceGraph(g, v, seeds)`. `composeEvents` loads only participant nodes in its batch `g.Nodes` (no place ids) | `lifePicks` never returns an event without a node, so `eventPlaceGraph`'s node check never skipped one in `composePersons`; its `g.Nodes(placeIDs)` existed only for `eventPlaces`' node reads, which D-1 removes (S16-18 Landed note) | A renamed helper for two lines with one caller left |
| D-3 | `LifeFacts` gains `Owners LifeOwners`. `FieldOwner` becomes `{ID []byte; PropertyID []byte; Place *ChainedPlace}` (`Entity` → `ID`). One function `lifeFacts(in *graphcache.InputScope, v vocab, pg *placeGraph, picks []lifePick) (LifeFacts, error)` *(new)* replaces `leadLife` and `lifeOwners`: it returns the zero `LifeFacts` for no picks; else `sortByDateKey(picks, …)` (stable), the lead is `picks[0]`; for each pick it calls `eventLocations(in, pick.event.ID)` and `eventPlaces(v, pg, pick.event, locations)`; the lead's places are `LifeFacts.Places`; each event adds one date owner `{ID: event id, PropertyID: pick.dateProp}` the first time it is seen; for each folded place, in fold order, each location link to it adds one place owner `{ID: l.Association, PropertyID: v.prop(hops.PlacesOfEvent.To()), Place: &ChainedPlace{Entity: p.Entity, Names: p.Names}}` the first time its association is seen. A link with an empty `Association` is an error, as today | One pass composes the row and its owners, so the life row reads what the summary shows (map scope: "`lifeOwners` goes"). The stable sort's first element is `leadLife`'s first strict minimum, so the lead does not move. Ids instead of entities: the row needs no association node, so a Persons list compose loads no association values; `PersonDerivedRows` loads every owner's node anyway | Keeping `Entity` on the owner — the composer would load every life event's location associations on every Persons list fill; keeping `lifeOwners` as a second function over the row's picks — two walks of the same links |
| D-4 | `PersonLifeOwners` goes. `PersonDerivedRows(c, personID)` reads `rows, err := conclusionsummaries.PersonsByIDs(c, [][]byte{personID})`; no row (unknown, merged, not a Person) returns `nil, nil`; else it takes `rows[0].Birth.Owners` and `rows[0].Death.Owners` and continues as today, with `OwnerField.Entity = entityOf(node)` from the node it loaded for the owner (its existing `conclusiondetails.entityOf`) and the owner skipped when that node is nil or merged (as `fieldsOf` drops it today). `len(personID) != 16` returns `nil, nil` before the read | The summary row is the one stored answer (working rule "Cache entities and results per key; compose answers"); Verify's display check and the rebuild-equals-upkeep tests now hold the owners too | A `PersonOwners` accessor in `conclusionsummaries` — `PersonsByIDs` already is the read |
| D-5 | `rolesByNeighbor` and `neighborRoles` move to `core/database/conclusionsummaries/participants.go` *(new)* as `RolesByNeighbor(links []graphcache.Link, label func(id []byte, key string) (Term, bool), keep func(id []byte) bool) []NeighborRoles` and `NeighborRoles{ID []byte; Roles []Term}` (its `seen` set stays unexported, built inside the function). Behavior unchanged. `persondetail` calls it for the person's own events; `oneEvent` calls it for participants | One collapse rule for both reads of a participation's roles; the composer cannot import `persondetail` | A second copy in the composer — two rules that drift |
| D-6 | The event row does not keep the Property its date came from (the map's scope entry says it does). The person row's date owners carry it (D-3); no other code reads it. The map's S16-19 row and scope entry and the design's line drop it in the planning commit | A field no one reads is dead on arrival (working rule "simpler code") | Adding `EventSummary.DateProperty` for a future reader — none is named in the map |
| D-7 | `EventParticipant{Entity canonicalentities.Entity; Name *namevalue.Value; NameValueCount int; Roles []Term}` *(new, `events.go`)* replaces `EventSubject` and `persondetail.EventParticipant`. `EventSummary.Subjects` becomes `[]EventParticipant` and `EventSummary.Participants []EventParticipant` *(new)*. `oneEvent` follows `hops.ParticipantsOfEvent` once through `in`; `RolesByNeighbor(links, known, keepAll)` with `known` = the term id is in `v.termLabel`, returning `Term{ID, Key}` unfinished; each neighbor is read with `in.Node`, skipped when nil, and becomes a participant with `entityOf`, `keptName` and its roles, in first-seen link order. `Subjects` walks the same links in order, keeps those whose `TermKey` equals `hops.SubjectsOfEvent`'s filter term, dedupes by neighbor and copies the participant with that id (absent when skipped). `composeEvents`' collect pass follows `hops.ParticipantsOfEvent` | One type for a person on an event (A-33). `Subjects` keeps today's set and order exactly (same links, same filter, same nil skip). Every participant's node is read because the type carries a name and an entity; the event row then records each participant's values and each participation association (D-7's cost, §6 Scale) | Ids-only participants plus `EventSubject` for the title — two shapes stay; names only on subjects — a half-filled type; filtering `Canonical` in the composer — changes event titles for a person with no accepted member (AC-3.5) |
| D-8 | `finishEvent` fills role labels into new slices: `h.Participants` and `h.Subjects` each become a new slice whose elements have `Roles` rebuilt through `v.finishTerm`, before `eventTitle(h)` | The stored row is shared and never written through (`finishTerm`'s contract); role labels stay off the stored row so a term rename does not drop it | Labels at compose time — a rename would leave stale labels on stored rows |
| D-9 | `PersonEvents` keeps its person `Follow` (`mine`) and the canonical-event filter. It then calls `EventsByIDs(c, events)` and, for each row, takes `h.Participants` minus the person; one `g.Nodes(ids)` over them; keeps those whose `g.Cached(id)` is `Canonical(personType)`; then `PersonsByIDs(people + self)` as today. `PersonEvent.Participants` becomes `[]conclusionsummaries.EventParticipant`; `dropUnreadEvents` and `sortParticipants` read `p.Entity.ID`; `persondetail.EventParticipant` goes; `personEventProto` reads `p.Entity.ID` and `p.Roles` | The Timeline reads the rows it loads (map scope) and keeps today's rule (not self, canonical) on the same nodes it loaded at base | Moving the canonical rule into the row — D-7's rejected filter |
| D-10 | `BenchmarkCacheSizes`' `displays` measure fills stored rows with `conclusionsummaries.ListPersons(c)` and `conclusionsummaries.ListEvents(c)` instead of `RowsForSearch`, and reports the fill's time as `displays-ms` beside its heap figure. STEP-1 records it at base with the same change applied locally (not committed), and STEP-7 records it after | The rows grow (owners, participants) and the working rule "Memory is budgeted" asks for before and after figures; today's measure stores nothing | Leaving the measure — the growth would go unmeasured |

**Rules this plan bends.** None. D-6 changes the map's scope entry, which the planning commit amends, and the user is told in the planning report.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `LifeFacts.Owners` *(new)* | field | `conclusionsummaries.LifeOwners` | Reused: the field holds the existing type |
| `lifeFacts` *(new)* | func | `leadLife`, `lifeOwners` | It is both, merged (D-3); both go |
| `EventParticipant` *(new, in `conclusionsummaries`)* | type | `EventSubject`, `persondetail.EventParticipant` | It replaces both (D-7); both go |
| `EventSummary.Participants` *(new)* | field | `EventSummary.Subjects` | Subjects is the subject-role subset the title reads; same element type |
| `RolesByNeighbor`, `NeighborRoles` *(moved)* | func, type | `persondetail.rolesByNeighbor` | Moved, not copied (D-5) |
| `core/database/conclusionsummaries/participants.go` *(new)* | file | `events.go` | Holds the moved collapse rule beside its test; `events.go` keeps the event types |
| `conclusionsummaries/participants_test.go` *(new)* | test file | `persondetail/timeline_drop_test.go` | `TestRolesByNeighbor` and `TestRolesByNeighborManyLinks` move with the function |

Searched: `rg "Owners|Participant|rolesByNeighbor|EventSubject|entityOf" core/database` — no other owner or participant type, and no other collapse rule.

**Builds on.** S16-18's place entries and `placeGraph.shell`; S14-40's `InputScope` and `inputBudget.addClosure`; the display cache and its Verify check; `PersonsByIDs` / `EventsByIDs` as the reads.
**Deletes.** `PersonLifeOwners`, `lifeOwners`, `leadLife`, `eventPlaceGraph`, `EventSubject`, `persondetail.EventParticipant`, `persondetail.rolesByNeighbor` and `neighborRoles` (moved), the Timeline's per-event `g.Follow(…, hops.ParticipantsOfEvent)` and its `walk` struct's `links`, `placesWorld.owners` in `eventplaces_cap_test.go` (replaced by reading the row), `FieldOwner.Entity`. `composeEvents`' place ids in its `g.Nodes`. After STEP-6, `placesOfEvents` keeps its callers (`composePersons`, `placeGraph.addChains`); `hops.SubjectsOfEvent` keeps `graphfill`; `conclusionsummaries.entityOf` keeps its callers.
**Next adds.** S16-24 reports a Person read's inputs: the derived rows now come from one stored row whose inputs are recorded, and the Timeline reads rows only. S15-28's records read composes from `LifeFacts.Owners`. P-8 can send each participant once by id, since the event row holds every participant.

### Reach

None — no new path into existing code. The derived rows and the Timeline read rows that exist at base; a row missing (unknown, merged, non-Person) is handled where today's walk returned nothing (D-4, D-9).

### Scale

At the map's Scale (16,000 persons on the large fixture; crafted: a 6,000-event city, 500 participants on an event, 21 places on an event):

- **Persons list fill.** Per person: no new node read (D-3's ids); per life event one `eventLocations` (already done by `addChains`) and the folded places it already computes for the lead, now for every life event (2–3 per person). Est. under 5 % of the list's fill time; measured by `displays-ms` (D-10) in STEP-1 and STEP-7.
- **Events list fill.** Per event: one `Follow` of every participation (today: subject-role links) and one node read per participant, all loaded already at open by `promotealign.Warm`; no place node read (today: one per place). On a write's one-shot graph the batch `g.Nodes` loads participants in one call.
- **Write fan-out.** An event row now records each participant's values and each participation association: a value edit on a person drops and reprojects every event they take part in, in any role (today: the events they are a subject of). On the large fixture a person takes part in about 2–4 more events than they are a subject of; STEP-7 records the event rows' key counts before and after (`TestLargeSourceBudgets` "inputs match isolated composition") and the time of a name edit on a person with the most participations.
- **Memory.** Each person row gains up to (life events + their location links) owners of 32 bytes plus a `ChainedPlace`; each event row gains its non-subject participants. STEP-1 and STEP-7 record `displays-B` at 4,000 and 16,000 persons.
- **The first `part_of` from the city** (map Check): logged by `TestLargeSourceBudgets`; STEP-1 records base, STEP-7 after.

### Edges

- **Assumptions.** D-1 assumes every place `eventPlaces` reads is a seed of `pg`: true for `composeEvents` (seeds = `neighborsOf(locations)`) and `composePersons` (seeds = `placesOfEvents` over every pick), held by AC-1.1, AC-1.2 and `TestBirthLocationsFoldWhenPartOf`. D-3 assumes `sortByDateKey` is stable (`sortByKey`, "Rows with equal keys keep their order"), so the lead is today's; held by AC-2.1's two-birth case. D-7 assumes `SubjectsOfEvent` is `ParticipantsOfEvent` filtered by `TermKey` on the same links in the same order (`Node.Follow`), held by AC-3.1 and the title cases of `core/database/conclusionsummaries/walks_test.go` ("marriage of two", "census of three").
- **Neighbors.** A participant who is both subject and witness (AC-3.2): one participant, both roles; `Subjects` order follows the first subject link, not the first link. A participant with no accepted member (AC-3.5): in the row, out of the Timeline. A place whose entry is `found: false` (no entity row): skipped, as a nil node was.
- **Removed inputs.** A participant deleted or merged: their node reads nil or merged; nil is skipped (as today's subjects); a merged node keeps today's behavior of `in.Node` (it was read and kept for subjects). The rebuild-equals-upkeep tests and Verify's display check hold every row. A person row over the input cap records no inputs (S14-40's crafted-catalog miss): its owners may then be stale until the row is dropped by its roots, as its places already may be; noted, not changed.
- **Real data.** None — no rule driven by a table or seed changes.
- **Untrusted size.** Participants per event: unbounded at base in the Timeline's read; the row now holds them (500 participants → 500 entries, one batched node load). Places per event: S16-01's cap of 20 holds (AC-2.4). Owners per life row: births × 20 places × links per place; the links per place are not capped at base (a place filed 21 times gives 21 owners) and stay so.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionsummaries` `compose.go` | `eventPlaces` (D-1); `eventPlaceGraph`, `PersonLifeOwners`, `lifeOwners`, `leadLife` go; `lifeFacts` *(new)* (D-3); `onePerson` calls `lifeFacts` for births and deaths; `composePersons` builds `pg` per D-2; `composeEvents` and `oneEvent` per D-7; `finishEvent` per D-8 |
| `conclusionsummaries.go` | `LifeFacts.Owners LifeOwners` *(new)*; `FieldOwner.ID` replaces `Entity`; comments name the summary row as the owners' home |
| `events.go` | `EventParticipant` *(new)*; `EventSubject` goes; `EventSummary.Subjects []EventParticipant`, `Participants []EventParticipant` *(new)*; `eventTitle` unchanged in behavior |
| `participants.go` *(new)* | `RolesByNeighbor`, `NeighborRoles` (D-5) |
| `core/database/conclusiondetails` `conclusiondetails.go` | `PersonDerivedRows` per D-4 |
| `core/database/persondetail` `timeline.go`, `participants.go` | D-9 |
| `api/ffi/handlers` `person_detail.go`, `conclusion_summaries.go` | `personEventProto` reads `p.Entity.ID`; `eventSummaryProto` ranges `h.Subjects` of the new type (same fields) |

Errors: none new. `lifeFacts` keeps "location link of event %x to place %x has no association"; "association %x is missing" goes with the association node read.

### FFI

None — no `engine.proto` change; two handlers read the new Go type (above).

### macOS

None — the wire is unchanged.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base**
Files: none committed. Run on `main` @ `7455758` with D-10's `displays` change applied locally.
Do: run `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign -run TestLargeSourceBudgets -v` three times and `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign -run '^$' -bench 'BenchmarkCacheSizes|BenchmarkListPlaces' -benchtime 1x`. Record in the ship log: "first part_of from the city" (three runs), the per-kind input key lines, `displays-B` and `displays-ms` at 4,000 and 16,000.
Check: the ship log holds the figures with the machine name.

**STEP-2 — Event places come from the place entries**
Files: `compose.go`; `core/database/conclusionsummaries/walks_test.go` or `placechains_test.go` for T-1, T-2.
Do: D-1, D-2. Keep `leadLife` and `lifeOwners` calling the new `eventPlaces` until STEP-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/... ./core/database/conclusiondetails/...` passes; T-1, T-2, T-3 pass.

**STEP-3 — The person row keeps its life owners**
Files: `compose.go`, `conclusionsummaries.go`, `core/database/conclusiondetails/conclusiondetails.go`, `placechains_test.go`, `eventplaces_cap_test.go`.
Do: D-3, D-4. Rewrite `TestPersonLifeOwners` as `TestPersonRowOwners` (T-4) over the same table, reading `PersonsByIDs`; `placesWorld.owners` reads the row's `Birth.Owners`.
Check: the three packages' tests pass; `rg -n "PersonLifeOwners|lifeOwners|leadLife|eventPlaceGraph" core api` prints nothing.

**STEP-4 — One collapse rule**
Files: `core/database/conclusionsummaries/participants.go` *(new)*, `core/database/conclusionsummaries/participants_test.go` *(new)*, `core/database/persondetail/timeline.go`, `core/database/persondetail/timeline_drop_test.go`.
Do: D-5. Move `TestRolesByNeighbor` and `TestRolesByNeighborManyLinks` unchanged in their cases.
Check: `go test` of both packages passes; `rg -n "rolesByNeighbor|neighborRoles" core` prints nothing.

**STEP-5 — The event row keeps its participants**
Files: `events.go`, `compose.go`, `api/ffi/handlers/conclusion_summaries.go`, `conclusionsummaries_test.go` (T-5, T-6).
Do: D-7, D-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg -n "EventSubject\b" core api --type go -g '!*.pb.go'` prints nothing.

**STEP-6 — The Timeline reads participants from the rows**
Files: `core/database/persondetail/timeline.go`, `core/database/persondetail/participants.go`, `core/database/persondetail/timeline_drop_test.go`, `core/database/persondetail/participants_test.go`, `core/database/persondetail/timeline_test.go` (T-7, T-8), `api/ffi/handlers/person_detail.go`.
Do: D-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/...` passes; `rg -n "ParticipantsOfEvent" core/database/persondetail` prints nothing.

**STEP-7 — Measure after**
Files: `core/database/promotealign/bench_test.go` (D-10), `docs/catalog-caches.md` (`displays` row).
Do: commit D-10's change; rerun STEP-1's commands on the branch; add a name edit on the large fixture's person with the most participations to the ship log (one `TestLargeSourceBudgets` log line, untimed by a budget). Update the `displays` row's size with the figures and its fill with "`ListPersons` / `ListEvents` store rows; `RowsForSearch` keeps none".
Check: the ship log holds before and after; §11's stops hold.

**STEP-8 — Docs**
Files: `completed.md`, `deployment-plan.md` (map row), this plan (status, Landed), `cache-architecture.md` (the S16-19 line if the code differs from it), the `LifeFacts` and `EventSummary` doc comments.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-19-summary-rows-carry-owners.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/walks_test.go` `TestEventRowReadsPlaceEntries` *(new)* | AC-1.1's city with three events, graph dropped → one place `["City"]`; `Cached(city)` false | AC-1.1, D-1, D-2 |
| T-2 | Go | same file `TestPersonRowReadsPlaceEntries` *(new)* | AC-1.2 → Town with parents `["City"]`; neither cached | AC-1.2, D-1, D-2 |
| T-3 | Go | `inputs_test.go` `TestInputsMatchIsolatedComposition`, `eventplaces_cap_test.go` `TestEventPlacesCapInputs` (unchanged) | Batch inputs equal isolated | AC-1.3 |
| T-4 | Go | `placechains_test.go` `TestPersonRowOwners` *(renamed from `TestPersonLifeOwners`)* | Today's cases, read from `PersonsByIDs(…)[0].Birth.Owners` / `Death.Owners` → today's expected owners by id and Property | AC-2.1, D-3 |
| T-5 | Go | `placechains_test.go` `TestPersonDerivedRows` (unchanged cases) plus a case: rename the birth place, commit → the owner's `Place.Names` is the new name; and cases merged person, non-person id, unknown id → no rows | AC-2.2, AC-2.3, AC-2.5, D-4 |
| T-6 | Go | `eventplaces_cap_test.go` `TestCraftedEventPlaces` (reads `Birth.Owners.Places` from the row) | 20 / 21 / filed-21-times → 20 / 20 / 21 owners; 4,000 places within 1 s | AC-2.4 |
| T-7 | Go | `conclusionsummaries_test.go` `TestEventParticipants` *(new)*, table | AC-3.1 → Ned `[subject]`, Pat `[witness]`, Subjects Ned; AC-3.2 → Ned `[subject, witness]`; AC-3.5 provisional participant → in `Participants` | AC-3.1, AC-3.2, AC-3.5, D-7 |
| T-8 | Go | `conclusionsummaries_test.go` `TestEventFinishLeavesStoredRow` *(new)* | `EventsByIDs` twice, then the stored display row's participants' `Roles[i].Label` is `""` | D-8 |
| T-9 | Go | `persondetail/timeline_test.go` `TestPersonEventsReadsParticipantsFromRows` *(new)* | AC-3.4: warm read with `SetInputRecorderForTest` → no key with `Bridge` `participation` and `ID` an event id; AC-3.5 provisional participant → absent from the Timeline | AC-3.4, AC-3.5, D-9 |
| T-10 | Go | `participants_test.go` `TestRolesByNeighbor`, `TestRolesByNeighborManyLinks` (moved) | Unchanged cases | D-5 |
| T-11 | Go | `TestPersonEvents`, `TestPersonEventsQueryCount`, `TestPersonEventsLifeline`, `TestParticipantOrder`, `TestDropUnreadEvents` (unchanged cases) | Unchanged results; equal query counts | AC-3.3, AC-3.6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `autoreconciler_test.go`'s rebuild-equals-upkeep (rows now carry owners and participants), `TestLargeSourceBudgets`, `TestPersonFamilyQueryCount`, the `conclusionsummaries` display check under Verify.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`PersonLifeOwners`, `lifeOwners`, `leadLife`, `eventPlaceGraph`, `EventSubject`, `rolesByNeighbor`, `persondetail.EventParticipant`, `FieldOwner.Entity`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] STEP-1 and STEP-7 figures in the ship log and the PR description; docs/catalog-caches.md `displays` row updated
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: cache-architecture.md's S16-19 line matches what landed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the large fixture project (or any project with a person born at a place `part_of` a region). Open a person's page: the header's birth and death lines and their places read as on `main`; hover or open the birth row's records: the same records as on `main`.
2. Open the Timeline section of a person who witnessed someone else's event: the event shows its other participants with their roles in the same order as on `main`.
3. Open the Events list: titles ("Birth of …", "Marriage of … and …") read as on `main`.
4. Rename the person's birth place on its Place page; return to the person: the birth row's place and its record show the new name.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A stored person row's owners go stale where `PersonLifeOwners` was fresh | The owners are read through the row's scope, so its inputs cover them (AC-1.3); Verify's display check and rebuild-equals-upkeep compare the stored row with a fresh one |
| Event rows record more keys, so person edits reproject more event rows | Measured in STEP-7 (§6 Scale); stop below |
| `Subjects` order or set moves for a person who holds two roles | D-7 builds `Subjects` from the subject-role links exactly as `SubjectsOfEvent` returns them (AC-3.2) |

**Stop and report if:**

- A row of §4 no longer holds at the planning base.
- `TestInputsMatchIsolatedComposition` or `TestLargeSourceBudgets`' "inputs match isolated composition" shows a person or place row's keys differ from the base after STEP-2 or STEP-3.
- The event rows' total input keys on the large fixture (STEP-7) are more than 2 × the STEP-1 figure, or the first `part_of` from the city is more than 1.25 × slower than STEP-1's median of three.
- `displays-B` at 16,000 persons grows by more than 20 MB.
- `displays-ms` (the Persons and Events lists' fill) at 16,000 persons is more than 1.25 × STEP-1's.
- An existing test's expected value must change for AC-2.2, AC-3.3 or the event titles to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S16-19 — Summary rows carry their life owners and participants
**Body:** objective in one sentence; what shipped against each US; AC evidence; STEP-1 / STEP-7 figures; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-19-summary-rows-carry-owners.md`.
