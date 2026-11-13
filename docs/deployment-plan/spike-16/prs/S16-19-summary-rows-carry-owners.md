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
**Context:** [design § Results per key](../cache-architecture.md#results-per-key-for-the-walks), [design § A cache a composer reads hands its inputs to the row](../cache-architecture.md#a-cache-a-composer-reads-hands-its-inputs-to-the-row), [`docs/catalog-caches.md`](../../../catalog-caches.md) (`displays`), [Spike 14 summary inputs](../../archive/spike-14/summary-inputs.md), [S16-18 plan and Landed note](S16-18-place-entries.md#landed), [`person-page-performance.md`](../../../ideas/person-page-performance.md) P-8, working rules: "The point of this batch is simpler code", "Cache entities and results per key; compose answers", "Every cache names the signal that updates it, and updates exactly", "Apply states its fan-out", "A write's one-shot graph fills only what it reads", "A cache a composer reads hands its inputs to the row", "Orders are defined and results do not move", "Memory is budgeted", "Batch 1's caps stay where S16-01 puts them", "No new persisted projections, and no stored answer without a measurement", "Measure growth"
**Planned:** 2026-11-13 against `origin/main` @ `7455758`
**Review:** [`S16-19-summary-rows-carry-owners.review.md`](S16-19-summary-rows-carry-owners.review.md) — round 1 answered; round 2 pending

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

Three Person page reads re-derive what the summary composers already walked. The event and person composers still load each place's whole node to read its names, although S16-18 put every place's names in its place entry. The Person page's life rows (`PersonDerivedRows`) walk the person's birth and death events, their locations and their place chains a second time through `PersonLifeOwners`, although the person's summary row was composed from the same walk. And the Timeline follows every event's participants itself, although it already loads each event's summary row. This PR makes the rows carry what those readers need: an event's places come from the place entries; a person's row keeps the owners behind its birth and death rows (each event with the Property its date came from, each event–place link with its place); and an event's row keeps every participant by id with their roles, from one `Follow` of the participants hop. `PersonLifeOwners`, `lifeOwners`, `eventPlaceGraph`, `placeGraph.addChains`, the Timeline's per-event walk and the collapse rule's home in `persondetail` go, and the Timeline's participant type moves onto the event row as the one participation type, `conclusionsummaries.EventParticipant` (structural debt A-33; `EventSubject` stays the title's name summary, user 2026-11-13).

**Why now.** Bar 13 ("Walks are lookups") and the batch rule "The point of this batch is simpler code": after S16-18, an event's places and a person's life owners are the last readers that walk past the cache. S16-24 reports the inputs of the Person reads, so it lands on the reads this PR leaves.

**What stays.** Every value the app shows: event titles, places and chains, the life rows and their owners in today's order, the Timeline's rows, roles and participant order. The wire (`EventSummary.subjects`, `PersonEvent.participants`) and Swift. `EventSubject` and the subjects read. The person and place rows' recorded input keys (S14-40's tests hold them); event rows also record their non-subject participation associations (D-7). S16-01's `eventLocations` cap of 20 places per event, `summaryChainDepth`, the input cap and the place entries.

## 2. User stories

**US-1.** As the engine, an event's or a person's summary row reads its places' names from the place entries, so composing an event at a city does not load the city's node with its 6,000 location links.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | City (one toponym "City") with three events located at it; the graph is dropped. `EventsByIDs(c, [event1])` | The row's places are one `ChainedPlace` with `Names` `["City"]`; afterwards `g.Cached(city)` is false |
| AC-1.2 | Pat born at Town `part_of` City; the graph is dropped. `PersonsByIDs(c, [pat])` | `Birth.Places` is Town with parents `["City"]`; `g.Cached(town)` and `g.Cached(city)` are false |
| AC-1.3 | The same catalogs composed with `RowsForSearch` | Each row's `Inputs` equal the keys it records composed alone (`CheckMatchIsolated`); the person and place rows' inputs equal those at the planning base |

**US-2.** As a researcher, the Person page's birth and death rows list the same records as today, read from the person's summary row instead of a second walk.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Each case of today's `TestPersonLifeOwners` table (two births, one dated 1820 and one undated; locations folding on one chain; one event filed at a place twice) | `PersonsByIDs` returns a row whose `Birth.Owners` and `Death.Owners` equal the case's expected owners by id, Property and place, in the same order: date owners lead first then date order, one per event; place owners one per event–place link, each with its place's own toponyms and no parents |
| AC-2.2 | `TestPersonDerivedRows`' cases | The same `[]DerivedRow` as at the planning base |
| AC-2.3 | `newFixture`: a life person born at York (`f.lifePerson`, `f.lifeEvent`, `f.placeSubject`, promoted); `PersonDerivedRows` once; then `f.Cite` a second toponym "Upper York" on York, which upkeep reports (`f.Reported`); `PersonDerivedRows` again | The birth place owner's `Place.Names` holds "Upper York": the write's `SummaryDeps` dropped the stored person row |
| AC-2.4 | A birth event filed at 21 distinct places (`TestCraftedEventPlaces`) | The row's `Birth.Owners.Places` has 20 owners; a place filed 21 times gives 21 owners |
| AC-2.5 | `PersonDerivedRows` for a merged person, a non-person id, an unknown id, a nil catalog | No rows, no error |

**US-3.** As the engine, the Timeline reads each event's participants from the event rows it already loads, through the one participation type.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Pat subject of B1; Ned subject and Pat witness of burial U. `EventsByIDs(c, [U])` | `Participants` is Ned `[subject]`, Pat `[witness]` in link order (ids and role keys); `Subjects` is Ned, as today |
| AC-3.2 | Ned holds subject and witness on one event (two links) | Ned is one participant with roles `[subject, witness]` in link order |
| AC-3.3 | Every case of `TestPersonEvents`, `TestPersonEventsLifeline` and `TestParticipantOrder` | Unchanged results, role labels included |
| AC-3.4 | The event rows are stored (a first `PersonEvents`), then `PersonEvents(c, pat)` with `SetInputRecorderForTest` on the graph | No key `{ID: an event id, Bridge: "participation", Endpoint: hops.ParticipantsOfEvent.From()}` is recorded: the Timeline follows no event's participants |
| AC-3.5 | A participant whose identity claims are provisional (`graphcachetest.Canon.Provisional`) | The event row lists them; the Timeline does not (as today) |
| AC-3.6 | `TestPersonEventsQueryCount` AC-2.2 (1 event × 1 guest vs 12 events × 3 guests and one of 41) | Equal query counts |
| AC-3.7 | After `PersonEvents`, `EventsByIDs(c, [U])` | The row's participant roles have empty `Label`: the Timeline labelled copies, not the stored row |

## 3. Scope

| In | Covered by |
| --- | --- |
| `eventPlaces` reads each location's place shell from the place graph (the entries) instead of `InputScope.Node` | STEP-2, AC-1.1, AC-1.3 |
| `eventPlaceGraph` goes; its callers call `buildPlaceGraph` over `placesOfEvents`; `composeEvents` stops loading place nodes | STEP-2, AC-1.1, AC-1.2 |
| `canonicalentities.FromNode` *(new)* replaces the three node-to-entity copies | STEP-3 |
| `LifeFacts.Owners` *(new)*: the date and place owners of a birth or death, composed in the same pass as the lead's facts; `leadLife` and `lifeOwners` become `lifeFacts`; `placeGraph.addChains` goes | STEP-4, AC-2.1, AC-2.4 |
| `PersonLifeOwners` goes; `PersonDerivedRows` reads the owners from the person's summary row | STEP-4, AC-2.2, AC-2.3, AC-2.5 |
| `rolesByNeighbor` moves to `conclusionsummaries.RolesByNeighbor`, the one collapse rule; `vocab.term` *(new)* is the one known-term check | STEP-5 |
| `conclusionsummaries.EventParticipant{ID, Roles}` (moved from `persondetail`) and `EventSummary.Participants` from one `Follow` of `hops.ParticipantsOfEvent` | STEP-6, AC-3.1, AC-3.2, AC-3.5 |
| The Timeline reads participants from the event rows and labels their roles; its per-event `Follow` goes | STEP-7, AC-3.3, AC-3.4, AC-3.6, AC-3.7 |
| `BenchmarkCacheSizes`' `displays` figure fills stored rows; before and after figures for memory, search inputs and write time | STEP-1, STEP-8 |
| The map row, scope entry and design line drop the event row's date Property and say what A-33 became (user, 2026-11-13) | Planning commit |

| Out | Home |
| --- | --- |
| An event row keeping the Property its date came from: no reader once the person row keeps its date owners (user, 2026-11-13) | not planned |
| One display type for subjects and participants (names on every participant) (user, 2026-11-13: participants are ids and roles) | not planned |
| Sending each participant once on the wire (P-8) | [`person-page-performance.md`](../../../ideas/person-page-performance.md) |
| Promote's checks reading the place entries | S16-21 |
| The Person reads reporting their inputs on the wire | S16-24 |
| A cap on participants per event | not planned — the Timeline reads every participant at base; the row holds what it read (§6 Edges) |

## 4. Current state

Verified at `origin/main` @ `7455758`.

| Area | Where | Today |
| --- | --- | --- |
| Event places | `core/database/conclusionsummaries/compose.go` `eventPlaces(in, v, pg, ev, locations)` | Reads each location's place with `in.Node(l.Neighbor)` for `entityOf(n)` and `toponyms(n, …)`, then `pg.FoldLocations` at the event's date. The node read loads a city's 6,000 location links |
| Place graph helper | `compose.go` `eventPlaceGraph(g, v, eventIDs)` | Checks each event has a node, `placesOfEvents(g.NoInputs(), …)`, `g.Nodes(placeIDs)` so `eventPlaces`' node reads batch (S16-18 Landed note), then `buildPlaceGraph`. Called by `composePersons` and `PersonLifeOwners` |
| Event composer | `compose.go` `composeEvents` | Follows `hops.SubjectsOfEvent` through a no-input scope to collect person ids, `eventLocations` through each row's scope, `g.Nodes(personIDs + placeIDs)`, `buildPlaceGraph(g, v, placeIDs)`; `rec.addClosure(in, pg, neighborsOf(locations))` per row |
| Event row | `compose.go` `oneEvent` | Follows `hops.SubjectsOfEvent` through `in`; each distinct neighbor read with `in.Node`, skipped when nil, kept as `EventSubject{Entity, Name, NameValueCount}` |
| Event row type | `core/database/conclusionsummaries/events.go` `EventSummary`, `EventSubject`, `eventTitle` | No participants |
| Person composer | `compose.go` `composePersons`, `onePerson`, `leadLife`, `lifePicks`, `lifeFrom`, `eventDate` | `leadLife` picks the earliest by `dateKeyLess` and folds its places; `pg.addChains(in, rec, lifeEvents(…))` follows every life event's locations again and records the place closure |
| Chains helper | `core/database/conclusionsummaries/placechains.go` `placeGraph.addChains` | One caller (`composePersons`): `placesOfEvents(in, eventIDs)` then `rec.addClosure` |
| Life owners | `compose.go` `PersonLifeOwners(c, personID)`, `lifeOwners(in, v, pg, picks)` | A second walk for the Person page: `lifePicks` and `eventPlaceGraph` again, `sortByDateKey(picks)`, one date owner per event (`entityOf(event)`, `pick.dateProp`), then per folded place every location link to it once, reading the association node for `entityOf(assoc)`; errors on an empty `Association` or a missing association node |
| Owner types | `core/database/conclusionsummaries/conclusionsummaries.go` `LifeFacts`, `FieldOwner{Entity, PropertyID, Place}`, `LifeOwners{Dates, Places}` | `LifeFacts` has no owners |
| Derived rows | `core/database/conclusiondetails/conclusiondetails.go` `PersonDerivedRows` | Calls `PersonLifeOwners`, loads every owner's node, `fieldsOf` with `only`, `EventsByIDs` for the date owners, builds `OwnerField{Entity: o.Entity, Place, Event, Field}`; an owner whose node is gone or merged is dropped by `fieldsOf` |
| Node to entity | `compose.go` `entityOf`, `conclusiondetails.go` `entityOf`, `api/ffi/handlers/source_read.go` (inline `canonicalentities.Entity{…}` from `n`) | Three copies of one constructor; `canonicalentities` already imports `graphcache` |
| Term check | `compose.go` `TermLabeler` | Reports false when the id is not in `v.termLabel`, else the finished term |
| Timeline | `core/database/persondetail/timeline.go` `PersonEvents` | Follows the person's `hops.EventsOfParticipant` once (`mine`), then per canonical event `g.Follow(event, hops.ParticipantsOfEvent)`, `g.Nodes(others)`, `rolesByNeighbor(links, label, keep)` with `keep` = not the person and `Canonical(personType)`, then `EventsByIDs`, `PersonsByIDs(people + self)`, `PersonEvent{Participants: []EventParticipant{Person, Roles}}` |
| Collapse rule | `timeline.go` `rolesByNeighbor`, `neighborRoles{id, roles, seen}` | Links in order; skip `keep` false; dedupe neighbors first-seen; add each link's term once per neighbor when `label` knows it |
| Participant type | `timeline.go` `EventParticipant{Person []byte, Roles}`, `dropUnreadEvents` | Built per read by the Timeline |
| Participant order | `core/database/persondetail/participants.go` `sortParticipants(ps []EventParticipant, index)` | Rank by roles, then `index` |
| Handler | `api/ffi/handlers/person_detail.go` `personEventProto` | Maps `r.Participants` by `p.Person` |
| Hops | `core/hops/hops.go` `SubjectsOfEvent`, `ParticipantsOfEvent` | Same bridge and end; `Node.Follow` (`core/database/graphcache/follow.go`) applies the subject filter. `FollowKeys` records the same bridge-end key for both |
| Display check | `core/database/conclusionsummaries/display_register.go` `checkDisplay` | Verify compares a stored person or event row, finished, with a fresh one |
| Stored rows | `compose.go` `persons` / `events` with `store=true, rec=nil` | Lists and `*ByIDs` store rows and record no inputs; `RowsForSearch` records inputs and stores nothing; search reprojection persists them in `catalog_search_inputs`, and a commit drops the stored rows of its roots and `SummaryDeps` (`core/database/graphcache/commit.go`) |
| Benchmark | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` `measure("displays", …)` | Fills through `RowsForSearch`, which stores no row (11 KB in `docs/catalog-caches.md`); `nodes-ms` is timed with `time.Since(start)` |
| Large fixture | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | Logs keys per kind ("inputs match isolated composition") and "first part_of from the city (untimed by the budget)" |
| Tests | `core/database/conclusionsummaries/placechains_test.go` `TestPersonLifeOwners`, `TestPersonDerivedRows`, `newFixture`; `eventplaces_cap_test.go` `TestCraftedEventPlaces`, `TestEventPlacesCapInputs`, `placesWorld.owners`; `inputs_test.go` `TestInputsMatchIsolatedComposition`; `walks_test.go`; `core/database/persondetail/timeline_test.go` `TestPersonEvents`, `TestPersonEventsQueryCount`, `TestPersonEventsLifeline`; `timeline_drop_test.go` `TestRolesByNeighbor`, `TestRolesByNeighborManyLinks`, `TestDropUnreadEvents`; `participants_test.go` `TestParticipantOrder`; `core/database/autoreconciler/autoreconciler_test.go` rebuild-equals-upkeep | Hold today's results |

**Precedent.** S16-18 ([#817](https://github.com/mendahu/provenencia/pull/817)) moved the chain build onto the place entries while every row kept its inputs: copy how a composer reads an entry and lets `inputBudget.addClosure` hand the entry's keys to the row, and `composeEvents`' shape (follow the locations once, pass them on, `rec.addClosure` over their places). S16-30 ([#812](https://github.com/mendahu/provenencia/pull/812)) removed `PersonDerivedRows`' second pass: copy its test of the derived rows staying equal. Do not copy S16-18's registry work: this PR adds no cache.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `eventPlaces(v vocab, pg *placeGraph, ev *graphcache.Node, locations []graphcache.Link) ([]ChainedPlace, error)` loses its `in` parameter. For each distinct `l.Neighbor` it reads `pg.shell(l.Neighbor)`; when `ok` it appends `ChainedPlace{Entity: shell.Entity, Names: shell.Names}`, else it skips the place (today's nil node) | The entry holds exactly `entityOf(n)` and `toponyms(n, …)` from a narrow node (`fillPlaceEntries` → `placeShell`). Every caller's `pg` has those places as seeds. The row still records each place's values: `addClosure` adds every seed's entry `inputs`, which start with `NodeKey(place)`; with `rec` nil or over the cap a row recorded nothing at base too | Keeping `in.Node(place)` — it loads the whole node, the read this PR removes |
| D-2 | `eventPlaceGraph` goes in STEP-2. `composePersons` and, until STEP-4, `PersonLifeOwners` build their graph with `seeds, err := placesOfEvents(g.NoInputs(), pickIDs)` then `buildPlaceGraph(g, v, seeds)`. `composeEvents` loads only person nodes in its batch `g.Nodes` (no place ids). Until STEP-4, `lifeOwners` keeps reading association nodes through `in` | `lifePicks` never returns an event without a node, so the node check never skipped one; `g.Nodes(placeIDs)` existed only for `eventPlaces`' node reads (S16-18 Landed note) | A renamed helper for two lines |
| D-3 | `canonicalentities.FromNode(n *graphcache.Node) Entity` *(new, `core/database/canonicalentities/canonicalentities.go`)* copies `ID` and `SubjectTypeID` and takes `Ref`, `Argument`, `Label`. `conclusionsummaries.entityOf`, `conclusiondetails.entityOf` and the inline copy in `source_read.go` go; every caller calls `canonicalentities.FromNode` | Three copies of one constructor; D-5 adds a caller (maintainability Q5: a third copy is extracted in the PR that touches one). No import cycle: `canonicalentities` imports `graphcache` today | A fourth caller of a private copy |
| D-4 | `LifeFacts` gains `Owners LifeOwners`. `FieldOwner` becomes `{ID []byte; PropertyID []byte; Place *ChainedPlace}`. `lifeFacts(in *graphcache.InputScope, v vocab, pg *placeGraph, picks []lifePick) (LifeFacts, [][]byte, error)` *(new)* replaces `leadLife` and `lifeOwners`. No picks → zero `LifeFacts`, no seeds. Else it sorts a copy of `picks` with `sortByDateKey(sorted, func(p lifePick) dateKey { return p.key })` (stable); the lead is `sorted[0]`. Per pick it calls `eventLocations(in, pick.event.ID)` once, appends `neighborsOf(locations)` to the returned seeds, and calls `eventPlaces(v, pg, pick.event, locations)`. The lead's places are `LifeFacts.Places`. Each event adds one date owner `{ID: pick.event.ID, PropertyID: pick.dateProp}` the first time it is seen; for each folded place, in fold order, each location link to it adds one place owner `{ID: l.Association, PropertyID: v.prop(hops.PlacesOfEvent.To()), Place: &ChainedPlace{Entity: p.Entity, Names: p.Names}}` the first time its association is seen. An empty `Association` is an error ("location link of event %x to place %x has no association"). `onePerson` returns the birth and death seeds; `composePersons` calls `rec.addClosure(in, pg, seeds)` per row; `placeGraph.addChains` goes | One pass composes the row and its owners (map: "`lifeOwners` goes"), and persons record place inputs as `composeEvents` does, without a second follow of the same locations. The stable sort's first element is `leadLife`'s first strict minimum, so the lead does not move. Ids, not entities: the Persons list fill loads no association node; `PersonDerivedRows` loads every owner's node anyway. Sorting a copy leaves `lives` untouched | `Entity` on the owner — every Persons list fill would load every life location's association; keeping `addChains` — a second follow of the locations `lifeFacts` read |
| D-5 | `PersonLifeOwners` goes. `PersonDerivedRows(c, personID)` returns `nil, nil` when `c == nil`, `c.Graph() == nil` or `len(personID) != 16`; else `rows, err := conclusionsummaries.PersonsByIDs(c, [][]byte{personID})`; no row (unknown, merged, not a Person) returns `nil, nil`; else it takes `rows[0].Birth.Owners` and `rows[0].Death.Owners` and continues as today, keyed on `o.ID`, with `OwnerField.Entity = canonicalentities.FromNode(node)` from the owner's loaded node; an owner whose node is nil or merged is skipped (as `fieldsOf` drops it today) | The summary row is the one stored answer; Verify's display check and rebuild-equals-upkeep now hold the owners too | A `PersonOwners` accessor — `PersonsByIDs` already is the read |
| D-6 | The event row does not keep the Property its date came from (user, 2026-11-13). The person row's date owners carry it | No reader | `EventSummary.DateProperty` with no reader |
| D-7 | A-33 (user, 2026-11-13): the participation type moves from `persondetail` to `events.go` as `conclusionsummaries.EventParticipant{ID []byte; Roles []Term}`, and `EventSummary` gains `Participants []EventParticipant`. `oneEvent` follows `hops.ParticipantsOfEvent` once through `in` and calls `RolesByNeighbor(links, v.term, func([]byte) bool { return true })`; each `NeighborRoles` becomes `EventParticipant{ID, Roles}` in first-seen link order. No participant node is read. Roles are stored unfinished (`Term{ID, Key}`) and `finishEvent` leaves them so. `Subjects` keep today's `follow(in, n.ID, hops.SubjectsOfEvent)` and `EventSubject` | The Timeline reads participant names from person rows (`personEventProto`), so the row needs ids and roles only (the shape P-8 names). The row records the participation bridge end (already recorded for subjects) and every participation association's values: a role edit on a witness drops the event row, which now shows that role; a person's value edit does not. The hop keeps its own subject filter (`Node.Follow`), and the second follow records no new key | Every participant with entity and name — fields nobody reads, a person's edit would reproject every event they witnessed, one persisted input row per participant (review M-1); filtering subjects from the participants' links by hand — a second home for the hop's filter (M-2) |
| D-8 | `vocab.term(id []byte, key string) (Term, bool)` *(new)* returns `Term{ID, Key}` when `id` is in `v.termLabel`. `TermLabeler` calls it and then `finishTerm`; `oneEvent` passes it to `RolesByNeighbor` | One known-term rule for both role reads (M-8) | A second `v.termLabel` check in the closure |
| D-9 | `rolesByNeighbor` and `neighborRoles` move to `events.go` as `RolesByNeighbor(links []graphcache.Link, label func(id []byte, key string) (Term, bool), keep func(id []byte) bool) []NeighborRoles` and `NeighborRoles{ID []byte; Roles []Term}` (the `seen` set becomes a local map inside the function). Behavior unchanged. `persondetail` calls it for the person's own events | One collapse rule; the composer cannot import `persondetail`; the rule sits beside the type it builds (M-7) | A second copy; a new `participants.go` named like `persondetail`'s ranking file |
| D-10 | `PersonEvents` keeps its person `Follow` (`mine`) and the canonical-event filter, then calls `EventsByIDs(c, events)`. For each row it takes `h.Participants` minus the person, one `g.Nodes(ids)` over them all, keeps those whose `g.Cached(id)` is `Canonical(personType)`, and builds a new `EventParticipant{ID, Roles}` per kept participant with a new `Roles` slice: each stored role through `label(r.ID, r.Key)`, dropped when false. Then `PersonsByIDs(people + self)` as today. `PersonEvent.Participants` becomes `[]conclusionsummaries.EventParticipant`; `persondetail.EventParticipant` goes; `dropUnreadEvents`, `sortParticipants` and `personEventProto` read `p.ID` | The Timeline reads the rows it loads and keeps today's rule (not self, canonical) and labels (`TermLabeler`). Building new slices keeps the stored row unwritten (AC-3.7) | Labels on the row (`finishEvent`) — every Events list read would relabel roles only the Timeline reads |
| D-11 | `PersonDerivedRows` reads the stored person row. A row composed in a search batch over `inputKeyCap` records no inputs, so on such a crafted catalog a write to a life event's place does not drop the person's stored row and its owners can show the old place name until the person, the event or the place's other roots are written — the miss S14-40 accepted for the row's places (user, 2026-11-13) | One stored answer (working rule); the miss reaches only a crafted catalog (`maxInputKeysPerCall` is above 50 × the large fixture's largest call) | Composing the row fresh for the Person page — a second compose path beside the stored one |
| D-12 | `BenchmarkCacheSizes`' `displays` measure fills stored rows with `conclusionsummaries.ListPersons(c)` and `conclusionsummaries.ListEvents(c)` instead of `RowsForSearch`, and reports `displays-ms` the way `nodes-ms` is taken (`start := time.Now()` before, `time.Since(start).Milliseconds()` after). The figures are compared before and after this PR only, not with the 11 KB on record (`ListPersons` also reads activity) | The rows grow and "Memory is budgeted" asks for before and after; today's measure stores nothing | Leaving it — the growth would go unmeasured |

**Rules this plan bends.** None. D-6, D-7 and D-11 are the user's answers of 2026-11-13; the map and the design are amended in the planning commit.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `canonicalentities.FromNode` *(new)* | func | `conclusionsummaries.entityOf`, `conclusiondetails.entityOf`, inline in `source_read.go` | Replaces all three (D-3) |
| `LifeFacts.Owners` *(new)* | field | `LifeOwners` | Holds the existing type |
| `lifeFacts` *(new)* | func | `leadLife`, `lifeOwners`, `placeGraph.addChains` | Merges them (D-4); all three go |
| `EventParticipant` *(moved)*, `EventSummary.Participants` *(new)* | type, field | `persondetail.EventParticipant` | Moved and renamed `Person` → `ID` (D-7) |
| `RolesByNeighbor`, `NeighborRoles` *(moved)* | func, type | `persondetail.rolesByNeighbor` | Moved, not copied (D-9) |
| `vocab.term` *(new)* | method | `TermLabeler`'s check | `TermLabeler` calls it (D-8) |

Searched: `rg "entityOf|Owners|Participant|rolesByNeighbor|termLabel\[" core api --type go`.

**Builds on.** S16-18's place entries and `placeGraph.shell`; S14-40's `InputScope` and `inputBudget.addClosure`; the display cache and its Verify check; `PersonsByIDs` / `EventsByIDs`.
**Deletes.** `PersonLifeOwners`, `lifeOwners`, `leadLife`, `eventPlaceGraph`, `placeGraph.addChains`, both private `entityOf`s and the inline copy in `source_read.go`, `persondetail.EventParticipant` (moved), `persondetail.rolesByNeighbor` and `neighborRoles` (moved), the Timeline's per-event `g.Follow(…, hops.ParticipantsOfEvent)` and its `walk.links`, `placesWorld.owners`, `FieldOwner.Entity`, `composeEvents`' place ids in `g.Nodes`, the "association %x is missing" error. Kept: `placesOfEvents` (`composePersons`), `hops.SubjectsOfEvent` (`oneEvent`, `graphfill`).
**Next adds.** S16-24 reports a Person read's inputs: the derived rows come from one stored row whose inputs are recorded, and the Timeline reads rows only. S15-28's records read composes from `LifeFacts.Owners` (owner, Property). P-8 sends each participant once by id: the event row already holds the ids and roles.

### Reach

None — no new path into existing code. A missing row (unknown, merged, non-Person) is handled where today's walk returned nothing (D-5, D-10).

### Scale

At the map's Scale (16,000 persons on the large fixture; crafted: a 6,000-event city, 500 participants on an event, 21 places on an event):

- **Persons list fill.** Per person no new node read (D-4); each life event's locations are followed once (today twice: `leadLife` and `addChains`), and the non-lead events' places are folded too (2–3 events per person). Measured by `displays-ms` (D-12).
- **Events list fill.** Per event one more `Follow` (participants) over links already loaded, no participant node read, no place node read (today one per place).
- **Write fan-out.** Event rows record their non-subject participation associations (a birth with father and mother: +2 keys). A role or date edit on such an association reprojects that event; a person's value edit reprojects what it does today. Persisted: one more `catalog_search_inputs` row per non-subject participation. Measured: the table's row count and a name edit's time on the person with the most participations, before and after.
- **Reindex at open.** `RowsForSearch` composes participants and every life event's owners for every row; measured by the time `TestLargeSourceBudgets` logs to reproject every handle (STEP-1 adds the log line if absent: see STEP-1).
- **Memory.** Each person row gains up to (life events + their location links) owners; each event row gains its participants as ids and roles. `displays-B` at 4,000 and 16,000.
- **The first `part_of` from the city** (map Check): logged by `TestLargeSourceBudgets`.

### Edges

- **Assumptions.** D-1: every place `eventPlaces` reads is a seed of `pg` (`composeEvents`: `neighborsOf(locations)`; `composePersons`: every pick's places), held by AC-1.1, AC-1.2, `TestBirthLocationsFoldWhenPartOf`. D-4: `sortByDateKey` is stable (`sortByKey`), so the lead is today's; AC-2.1's two-birth case. D-7: the participants' first-seen order is `Node.Follow`'s link order (association ref, then neighbor ref), as the Timeline's was.
- **Neighbors.** A person holding two roles on one event (AC-3.2): one participant, both roles in link order; `Subjects` unchanged. A participant with no accepted member (AC-3.5): in the row, out of the Timeline. A place whose entry is `found: false`: skipped, as a nil node was. A role term the vocabulary does not know: no role, as `TermLabeler` gave none.
- **Removed inputs.** A participant deleted: their participation association goes, the event's participation end moves, the row is dropped. A participant merged: links are rewritten to the survivor by the merge commit (roots). The rebuild-equals-upkeep tests and Verify's display check hold every row. Over the input cap: D-11.
- **Real data.** None — no rule driven by a table or seed changes.
- **Untrusted size.** Participants per event: unbounded at base in the Timeline's read; the row holds them as ids and roles (500 participants → 500 entries, no node read). Places per event: S16-01's cap of 20 holds (AC-2.4). Owners per life row: births × 20 places × links per place, uncapped at base and here.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/canonicalentities` | `FromNode` *(new)* (D-3) |
| `core/database/conclusionsummaries` `compose.go` | `eventPlaces` (D-1); `eventPlaceGraph`, `PersonLifeOwners`, `lifeOwners`, `leadLife`, `entityOf` go; `lifeFacts` *(new)* (D-4); `onePerson`, `composePersons` per D-4; `oneEvent`, `composeEvents` per D-2, D-7; `vocab.term` *(new)*, `TermLabeler` per D-8 |
| `placechains.go` | `addChains` goes |
| `conclusionsummaries.go` | `LifeFacts.Owners` *(new)*; `FieldOwner.ID` replaces `Entity`; comments name the summary row as the owners' home |
| `events.go` | `EventParticipant` *(moved)*, `EventSummary.Participants` *(new)*, `RolesByNeighbor`, `NeighborRoles` *(moved)* |
| `core/database/conclusiondetails` `conclusiondetails.go` | `PersonDerivedRows` per D-5; `entityOf` goes |
| `core/database/persondetail` `timeline.go`, `participants.go` | D-10 |
| `api/ffi/handlers` `person_detail.go`, `source_read.go` | `personEventProto` reads `p.ID`; `source_read.go` calls `canonicalentities.FromNode` |

Errors: none new; "association %x is missing" goes with the association node read.

### FFI

None — no `engine.proto` change.

### macOS

None — the wire is unchanged.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base**
Files: none committed; `core/database/promotealign/bench_test.go` (D-12) and a timing log in `perf_test.go` applied locally on `origin/main` @ `7455758`.
Do: locally apply D-12's `displays` change; in `TestLargeSourceBudgets` log `SELECT COUNT(*) FROM catalog_search_inputs` after the fixture is built, the time of a `searchindex.ReprojectHandles` over every handle, and the time of one name Cite on the person with the most participations. Run `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign -run TestLargeSourceBudgets -v -count=1` three times and `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign -run '^$' -bench BenchmarkCacheSizes -benchtime 1x`. Record in the ship log `docs/deployment-plan/spike-16/prs/S16-19-summary-rows-carry-owners.ship.md` *(new)*, headed by the output of `hostname`: "first part_of from the city", the per-kind key lines, the `catalog_search_inputs` count, reproject-all time, the name edit time (three runs each), `displays-B` and `displays-ms` at 4,000 and 16,000.
Check: the ship log holds every figure.

**STEP-2 — Event places come from the place entries**
Files: `compose.go`; `core/database/conclusionsummaries/walks_test.go` (T-1, T-2).
Do: D-1, D-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/... ./core/database/conclusiondetails/...` passes; `rg -n "eventPlaceGraph" core` prints nothing.

**STEP-3 — One node-to-entity constructor**
Files: `core/database/canonicalentities/canonicalentities.go`, `compose.go`, `placechains.go`, `conclusiondetails.go`, `api/ffi/handlers/source_read.go`.
Do: D-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg -n "func entityOf" core api` prints nothing.

**STEP-4 — The person row keeps its life owners**
Files: `compose.go`, `placechains.go`, `conclusionsummaries.go`, `conclusiondetails.go`, `core/database/conclusionsummaries/placechains_test.go`, `eventplaces_cap_test.go`.
Do: D-4, D-5. Rewrite `TestPersonLifeOwners` as `TestPersonRowOwners` (T-4): same table; its expected owner ids (association ids for place owners, event ids for date owners) compare with `FieldOwner.ID`. `placesWorld.owners` reads `PersonsByIDs(…)[0].Birth.Owners`.
Check: the three packages' tests pass; `rg -n "PersonLifeOwners|lifeOwners|leadLife|addChains" core api` prints nothing.

**STEP-5 — One collapse rule and one term check**
Files: `events.go`, `compose.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`, `core/database/persondetail/timeline.go`, `timeline_drop_test.go`.
Do: D-8, D-9. Move `TestRolesByNeighbor` and `TestRolesByNeighborManyLinks` to `conclusionsummaries_test.go` with the same cases, reading `.ID` / `.Roles`.
Check: both packages pass; `rg -n "rolesByNeighbor|neighborRoles" core` prints nothing.

**STEP-6 — The event row keeps its participants**
Files: `events.go`, `compose.go`, `conclusionsummaries_test.go` (T-7).
Do: D-7. Until STEP-7 the Timeline keeps its own walk.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-7 — The Timeline reads participants from the rows**
Files: `core/database/persondetail/timeline.go`, `core/database/persondetail/participants.go`, `core/database/persondetail/timeline_drop_test.go`, `core/database/persondetail/participants_test.go`, `core/database/persondetail/timeline_test.go` (T-8, T-9), `api/ffi/handlers/person_detail.go`.
Do: D-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/...` passes; `rg -n "ParticipantsOfEvent" core/database/persondetail` prints nothing; `rg -n "type EventParticipant" core` prints only `events.go`.

**STEP-8 — Measure after**
Files: `core/database/promotealign/bench_test.go` (D-12, committed), `perf_test.go` (STEP-1's three log lines, committed), `docs/catalog-caches.md` (`displays` row).
Do: rerun STEP-1's commands on the branch; record the figures beside the base. Update the `displays` row's size and its fill ("`ListPersons` / `ListEvents` store rows; `RowsForSearch` keeps none").
Check: the ship log holds before and after; §11's stops hold.

**STEP-9 — Docs**
Files: `completed.md`, `deployment-plan.md` (map row status), this plan (status, Landed), the `LifeFacts`, `FieldOwner`, `EventSummary` and `EventParticipant` doc comments.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-19-summary-rows-carry-owners.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/walks_test.go` `TestEventRowReadsPlaceEntries` *(new)* | AC-1.1's city with three events, graph dropped → one place `["City"]`; `Cached(city)` false | AC-1.1, D-1, D-2 |
| T-2 | Go | same file `TestPersonRowReadsPlaceEntries` *(new)* | AC-1.2 → Town with parents `["City"]`; neither cached | AC-1.2, D-1, D-2 |
| T-3 | Go | `inputs_test.go` `TestInputsMatchIsolatedComposition`, `eventplaces_cap_test.go` `TestEventPlacesCapInputs` (unchanged) | Batch inputs equal isolated | AC-1.3, D-4 |
| T-4 | Go | `placechains_test.go` `TestPersonRowOwners` *(renamed from `TestPersonLifeOwners`)* | Today's cases read from the row → today's owners by id, Property and place | AC-2.1, D-4 |
| T-5 | Go | `placechains_test.go` `TestPersonDerivedRows` (unchanged cases) plus cases: AC-2.3's toponym Cite → "Upper York"; merged person, non-person id, unknown id, nil catalog → no rows | AC-2.2, AC-2.3, AC-2.5, D-5 |
| T-6 | Go | `eventplaces_cap_test.go` `TestCraftedEventPlaces` (owners from the row) | 20 / 21 / filed-21-times → 20 / 20 / 21 owners; 4,000 places within 1 s | AC-2.4 |
| T-7 | Go | `conclusionsummaries_test.go` `TestEventParticipants` *(new)*, table | AC-3.1 → Ned `[subject]`, Pat `[witness]`, Subjects Ned; AC-3.2 → Ned `[subject, witness]`; AC-3.5 provisional participant → in `Participants` | AC-3.1, AC-3.2, AC-3.5, D-7 |
| T-8 | Go | `core/database/persondetail/timeline_test.go` `TestPersonEventsReadsParticipantsFromRows` *(new)* | AC-3.4 with the recorder: no event-end participation key; AC-3.5 provisional participant → absent; AC-3.7 → stored roles' `Label` empty | AC-3.4, AC-3.5, AC-3.7, D-10 |
| T-9 | Go | `TestPersonEvents`, `TestPersonEventsQueryCount`, `TestPersonEventsLifeline`, `TestParticipantOrder`, `TestDropUnreadEvents` (unchanged cases) | Unchanged results; equal query counts | AC-3.3, AC-3.6 |
| T-10 | Go | `conclusionsummaries_test.go` `TestRolesByNeighbor`, `TestRolesByNeighborManyLinks` (moved) | Unchanged cases | D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: rebuild-equals-upkeep in `autoreconciler_test.go`, `TestLargeSourceBudgets`, `TestPersonFamilyQueryCount`, the `conclusionsummaries` display check under Verify, the source read tests (`source_read.go`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg -n "PersonLifeOwners|lifeOwners|leadLife|eventPlaceGraph|addChains|rolesByNeighbor|persondetail\.EventParticipant|func entityOf" .agents .cursor/rules docs/*.md core api` has no hit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory (no UI)
- [ ] Version bump as the header says — none
- [ ] STEP-1 and STEP-8 figures in the ship log and the PR description; docs/catalog-caches.md `displays` row updated
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ (not gated)
- [ ] Domain doc: cache-architecture.md's S16-19 line matches what landed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person born at a place `part_of` a region. On the person's page the birth and death lines and their places read as on `main`; the birth row's records are the same as on `main`.
2. Open the Timeline of a person who witnessed someone else's event: the event shows its other participants with their roles in the same order as on `main`.
3. Open the Events list: titles ("Birth of …", "Marriage of … and …") read as on `main`.
4. Rename the person's birth place on its Place page; return to the person: the birth row's place and its record show the new name.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A stored person row's owners go stale where `PersonLifeOwners` was fresh | The owners are read through the row's scope (AC-1.3), and a write drops the row through the persisted inputs (AC-2.3); Verify's display check and rebuild-equals-upkeep compare the stored row with a fresh one. The crafted over-cap case is D-11 |
| Event rows record more keys | Measured in STEP-8; stops below |
| A stored row's slices written through by a reader | D-10 builds new slices; AC-3.7 |

**Stop and report if:**

- A row of §4 no longer holds at `origin/main` @ `7455758`.
- `TestInputsMatchIsolatedComposition` or `TestLargeSourceBudgets`' "inputs match isolated composition" shows a person or place row's keys changed after STEP-2 or STEP-4.
- After STEP-8: `catalog_search_inputs` holds more than 1.5 × STEP-1's rows; the name edit, reproject-all or the first `part_of` from the city is more than 1.25 × slower than STEP-1's median of three; `displays-ms` at 16,000 persons is more than 1.25 × STEP-1's; `displays-B` at 16,000 grows by more than 20 MB.
- An existing test's expected value must change for AC-2.2, AC-3.3 or the event titles to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S16-19 — Summary rows carry their life owners and participants
**Body:** objective in one sentence; what shipped against each US; AC evidence; STEP-1 / STEP-8 figures; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-19-summary-rows-carry-owners.md`.
