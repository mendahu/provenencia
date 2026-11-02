# S13-05 — Person timeline read

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine read, an RPC, a store method and a cache key; no view reads them until S13-13)
**Board:** —
**Depends on:** S13-01 (landed [#655](https://github.com/mendahu/provenencia/pull/655)), S13-02 (landed [#659](https://github.com/mendahu/provenencia/pull/659)), S13-31 (landed [#660](https://github.com/mendahu/provenencia/pull/660))
**Blocks:** S13-13
**Branch:** `s13-05-person-timeline-read`
**Runs on:** macOS: STEP-5 needs `protoc` and its plugins, and STEP-7, STEP-8 and `scripts/check-pr.sh` need Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01 and S13-04 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-2, R-23 – R-28), working rules: "Each section reads its own query", "One invalidation rule for person-section keys", "One person summary for every tile", "Canonical only, filtered in Go", "Timeline is direct participation, any role", "Engine PRs run ahead of their slice"
**Planned:** 2026-11-02 against `main` @ `f4a1c94`
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

The engine can answer "what happened in this person's life": every canonical event the person took part in directly, in any role, earliest first with the undated ones last and flagged, each as an `EventSummary` with the person's role and the other canonical participants, each a `PersonSummary` with their role. The app can fetch it under its own per-person cache key. A second read issues no graph query, and a first read loads in a fixed number of batches however many events and participants there are.

**Why now.** S13-13 (Timeline section, brief S13-D4) draws this read. Every piece it composes has landed: S13-01's accepted-members helper and per-person key pattern, S13-02's gender terms on the summary, S13-31's names, and S13-04's canonical-relative rule and section-read shape.

**What stays.** `GetConclusionDetail`, `ListPersonClaims`, `GetPersonFamily` and every view are unchanged. `hops.EventsOfSubject` and `hops.SubjectsOfEvent` keep their `subject` filter: a summary's birth, death, event subjects and title still read subject-role links only. `EventSummary.Subjects` is unchanged. The Events list order and its code are unchanged.

## 2. User stories

People and events below are built with `graphcachetest.Canon` in `kinWorld` (`core/database/persondetail/family_test.go`): each is canonical with one accepted member unless the AC says otherwise. "Birth B (1850): Kid subject, Pat father" is a birth event dated 1850 with two participation links. Events are written `Title-word(roles; participants)`.

**US-1.** As the engine, when asked for a canonical person's events, I return every event they took part in directly, in date order, so that the Timeline section can draw the life from birth to death.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Pat: birth B1 (1820) as subject; marriage M (1845) with Sam, both subject; census C (1851) as subject with Kid as subject; burial U (1890) of Ned, Pat witness | Four events in order B1, M, C, U. Pat's roles: subject, subject, subject, witness. U lists Ned (subject); M lists Sam (subject); C lists Kid (subject); B1 lists none |
| AC-1.2 | Birth B2 (1850): Kid subject, Sam mother; Pat not linked. A recorded parent link Pat→Kid | B2 is not in Pat's events (R-23) |
| AC-1.3 | Birth B2 (1850): Kid subject, Pat father, Sam mother | B2 is in Pat's events with role father; participants Kid (subject) and Sam (mother), in participation-ref order |
| AC-1.4 | Event R (residence) with no date; event T (travel) with only a start date 1860; event V (migration) with only an end date 1870; birth B1 (1820) | Order B1, T, then R and V by ref. `Dated` is true for B1 and T, false for R and V |
| AC-1.5 | Two undated events and two events dated 1850 | Equal dates order by event ref; undated order by event ref (the Events list's rule) |
| AC-1.6 | Pat has two participation links to census C, one as subject and one as informant | C appears once, with roles subject then informant (link order); a second link with the same role adds no second role |
| AC-1.7 | Pat's participation link in C has no role term | C appears with no role |
| AC-1.8 | Census C: Pat, Kid, Ann subjects. Kid's only claim is provisional; Ann was merged into Bea | C lists no Kid, no Ann, no Bea (Bea has no participation link in C) |
| AC-1.9 | Event E1's only claim is provisional; event E2 was merged into E3, and Pat's link points at E2 | Neither E1 nor E2 is in Pat's events (graph loads skip merged neighbors) |
| AC-1.10 | Entity id unknown, of an Event, or merged; a malformed id | `conclusiondetails.ErrNotFound` |
| AC-1.11 | An unmerged Person with no participation link | Zero events, no error |
| AC-1.12 | A Person whose only claim is provisional, subject of birth B1 | B1 is returned (the page opens for a provisional-only Person, as the family read does) |
| AC-1.13 | A participant summary missing mid-read (a node dropped between the shown check and the summary read) | That participant is left out; the event stays (D-7) |

**US-2.** As the engine, I serve the timeline from the graph cache, so that the page reads it cheaply.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `PersonEvents(Pat)` read twice on AC-1.1's catalog | The second read issues zero graph queries (`SetQueryCounterForTest`) and returns an equal result |
| AC-2.2 | A cold `PersonEvents` on (a) Pat with one dated event with one Place and one co-participant; (b) Pat with twelve dated events, one of them a census with 41 co-participants (the map's Scale obituary), each other with three; every event located at its own Place (`Canon.Location`) and every co-participant a person with a dated birth event at a Place, in both (a) and (b), so the same summary batches run | Both issue the same number of graph queries, and it is more than zero |
| AC-2.3 | After AC-2.1's warm read, a new participation of Pat as witness in event W is filed and committed (`Canon.Commit` with Pat, W and the participation id) | The next read includes W with role witness |

**US-3.** As the app, I fetch a person's events over FFI.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ListPersonEvents` on AC-1.1 + AC-1.3's catalog | Five `PersonEvent`s in date order; each `event` is that event's `EventSummary` (entity id, ref, title); U's `roles` is one `ConclusionTerm` with key `witness` and label `Witness`; U's `participants` is Ned's `PersonSummary` with role `subject`; B2's participants are Kid then Sam; `dated` is true for all five |
| AC-3.2 | Unknown or malformed `entity_id`, or an Event's id | Error code `conclusiondetails.not_found` |

**US-4.** As the app, I keep a person's events in the session cache and refresh them when they may be stale.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Seeded events for e1 in `FakeStore`; a query of `.personEvents(project:, entityId: "e1")` | The handle's value is the seed; e9 (unseeded) throws `conclusiondetails.not_found` |
| AC-4.2 | A promote (`.promotedSubject`), a Source delete, a Property update | Each invalidates `.allCached(.personEvents)`; a mutation that is not a Conclusion trigger does not |
| AC-4.3 | A warmed `.personEvents` key for e1; write effects touching only `e7` (an event's or co-participant's handle) | The key is revalidated; with no touched handle it is not; the key's hidden policy is `.evict` and its stale policy `.sessionFresh` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `hops.EventsOfParticipant` and `hops.ParticipantsOfEvent`: participation in any role | STEP-1, AC-1.1 |
| `conclusionsummaries.HasEventDate`: the one "dated" rule, exported beside the event order it mirrors | STEP-2, AC-1.4, AC-1.5 |
| `persondetail.PersonEvents`: load, filter to canonical events and participants, label roles, attach summaries, order | STEP-3, STEP-4, AC-1.1 – AC-1.13, AC-2.1 – AC-2.3 |
| `ListPersonEvents` RPC (proto, regenerate, handler, dispatch) | STEP-5, STEP-6, AC-3.1, AC-3.2 |
| Swift: wire mapping, `GenealogyStore.listPersonEvents`, `FakeStore` seed, `.personEvents` key and descriptor | STEP-7, STEP-8, AC-4.1 – AC-4.3 |
| Docs the PR makes false | STEP-9 |

| Out | Home |
| --- | --- |
| The section: rows, titles through `EventTitleDisplay`, the Undated group, participant tiles, links to Event and Person pages | S13-13 |
| `ConclusionPageSection.timeline.queryKey` returning `.personEvents` (the place warms it, the switcher's refresh follows it) | S13-13 (S13-12 plan D-8: "S13-13 and S13-16 each add one case") |
| Drawing an event depiction in the placeholder slot (R-26); the event's lead picture path is on the wire already, through `EventSummary.picture_image_rel_path` | S13-13 keeps the placeholder; filling it is not planned this spike ([person-detail.md § Not this spike](../person-detail.md#not-this-spike)) |
| Folding `EventSummary.subjects` (`EventSubject`) into {role, `PersonSummary`} participants. Logged: `subjects` feeds `eventtitle` and the lists and is subject-role and not filtered to canonical; the timeline's participants are a page read beside it. Two shapes stay, each with one reader | Not planned |
| Showing that a role is disputed (more than one kept role value on a participation) | Not planned; the role is the participation's rank-1 kept term, as `graphcache.Link.TermKey` is everywhere |
| Indirect events (a child's birth the person is not linked to, a spouse's death) | Not planned (R-23 excludes them) |
| A per-person memo slot for the timeline | Not planned (D-6: every input is already cached) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Hops | `core/hops/hops.go` `EventsOfSubject`, `SubjectsOfEvent`, `MustHop`, `NewHop`, `Hop.Reverse` | `EventsOfSubject` is `MustHop("participation", "person", "event", &TermFilter{role, subject})`. `NewHop` accepts a nil filter. No hop crosses participation without the filter |
| Following a hop | `core/database/graphcache/follow.go` `Graph.Follow` | Returns a node's links of the hop's bridge and end, filtered by term only when the hop has a filter; loads the node on first read; link order is association ref, then neighbor ref |
| Links | `core/database/graphcache/graph.go` `Link` (`Neighbor`, `Association`, `AssociationRef`, `TermID`, `TermKey`, `BridgeTypeKey`); `core/database/graphcache/load.go` `finishLinks`, `disambiguation` | A participation link carries the association's rank-1 kept `role` term, or no term. Merged neighbors are not loaded as links |
| Node checks | `graph.go` `Node.Is`, `Node.AcceptedMembers`; `Graph.Nodes`, `Graph.Node`, `memo.go` `Graph.Cached` | `Is`: not nil, not merged, of the type |
| Vocabulary | `core/database/graphcache/vocab.go` `Graph.Vocabulary`, `Vocabulary.TermsByID`, `VocabTerm`, `TypeIDsByKey` | Cached; a vocabulary write drops it. `conclusiondetails.attachValues` labels terms through `TermsByID` |
| Canonical rule | `core/database/persondetail/family.go` `shownPerson`, `personTypeID` | `shownPerson(n, personTypeID)` = `n.Is(personTypeID) && len(n.AcceptedMembers()) > 0`; its doc names it the package's one canonical-relative rule. One caller, `Families`' `kinOf` |
| Section reads | `family.go` `PersonFamily`, `dropUnread`; `claims.go` `Claims` | `PersonFamily`: `conclusiondetails.Handle`, the read, one `conclusionsummaries.ByIDs`, drop what has no summary, sort with `SortByBirth` |
| Page-handle guard | `core/database/conclusiondetails/conclusiondetails.go` `Handle`, `ErrNotFound` | nil catalog or graph, a malformed id, unknown or merged → `ErrNotFound` |
| Summaries | `core/database/conclusionsummaries/compose.go` `ByIDs`, `Want.Add`, `Summaries.Event`, `Summaries.Person`, `Summaries.PictureImageFileIDs`, `events`, `composeEvents`; `events.go` `EventSummary`, `Term` | `EventsByIDs` returns rows sorted by `sortEvents`; `Summaries` is a map and loses that order |
| Date order | `compose.go` `sortEvents`, `eventKeyOf`, `dateKey`, `dateKeyLess`, `dateSort`; `order.go` `sortByDateKey`, `SortByBirth` | `sortEvents(rows)` = `sortByDateKey(rows, eventKeyOf)`: date, else start date, dated first, then ref. Its one caller is `events`. `SortByBirth` is the exported generic beside it (S13-12) |
| Summary wire | `api/ffi/handlers/conclusion_summaries.go` `eventSummaryProto`, `personSummaryProto`, `termProto` | Map a summary with its picture path from one `files.RelPaths` call |
| Handlers | `api/ffi/handlers/person_detail.go` `ListPersonClaims`, `GetPersonFamily` | `parseID` failure → `ErrNotFound`, `withProjectCatalog`, one read, map |
| Dispatch | `api/ffi/dispatch.go` `MethodGetPersonFamily`, `case MethodGetPersonFamily` | One const and one case per method |
| Proto | `api/proto/engine.proto` `Method` (last: `METHOD_GET_PERSON_FAMILY = 113`), `GetPersonFamilyRequest`, `EventSummary`, `PersonSummary`, `ConclusionTerm` | Numbers 114 and up are unused |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogEventSummary`, `CatalogPersonSummary`, `CatalogConclusionValue`, `CatalogPersonFamily`, protocol `getPersonFamily`; `GoStore.swift` `getPersonFamily`, `mapEventSummary`, `mapPersonSummary`, `mapConclusionTerm` (private); `FakeStore.swift` `seededPersonFamilies`, `getPersonFamily`; `CoreInvoke.swift` `CoreMethod.getPersonFamily` | S13-04's shape |
| Cache key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personFamily`, `Kind.personFamily`, the `kind` and `project` switches; `CatalogQueryRegistry.swift` descriptor `case .personFamily`, `conclusionTriggers` | `.personFamily` uses `refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)]` |
| Section keys | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | `.timeline` returns nil; stays nil in this PR |
| Tests | `core/database/persondetail/family_test.go` `kinWorld`, `newKinWorld`, `person`; `family_query_test.go` `TestPersonFamilyQueryCount`; `family_drop_test.go` (package `persondetail`); `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestSortByBirth`; `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `everyKindHasOneDescriptor` (asserts 20 kinds), `invalidationsComeFromRegistrySpecs`, `personFamilyInvalidateOnPromote`, `personFamilyLoadsThroughTheStore`; `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personFamilyFollowsEveryTouchedHandle`; `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | `invalidationsComeFromRegistrySpecs` lists the exact invalidations of `.deletedSource` and `.promotedSubject` and asserts every Conclusion trigger stales `.allCached(.personFamily)` |
| Builders | `core/database/graphcache/graphcachetest/graphcachetest.go` `Canon.Entity`, `Person`, `Event`, `Join`, `Birth`, `Marriage`, `Date`, `Merge`, `Provisional`, `Commit`, `observe`; `Link` always files a term | No builder files an event's location (a bridge with no disambiguating term); tests elsewhere use their own `f.bridge("location", …)` fixtures |

**Precedent.** S13-04 ([#671](https://github.com/mendahu/provenencia/pull/671)) added a per-person read in `persondetail` that attaches summaries at the edge, its handler, proto, `CoreMethod`, store method, `FakeStore` seed, `CatalogQueryKey` case and descriptor, and their tests: copy that file list, each file's shape, its query-count test (`TestPersonFamilyQueryCount`) and its `dropUnread` split for the mid-read drop (`family_drop_test.go`). Do not copy S13-04's `canonlinks` or `kinship` changes: the timeline needs none.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Layers. `core/hops` names the any-role participation; `conclusionsummaries` owns event date order (`EventsByIDs`' list order) and "dated" (`HasEventDate`); `core/database/persondetail/timeline.go` *(new)* loads, filters, labels roles, attaches summaries and emits rows in that order; the handler only maps | Working rules "Each section reads its own query" (`persondetail`) and "Timeline is direct participation, any role" (the hop in `core/hops`; "ordered by `conclusionsummaries.sortEvents`"); the date rule already lives in `conclusionsummaries` | Walking `n.Links` by bridge key in `persondetail` — a link walk with no named hop (map review K-5); sorting in Swift — "data, not layout" puts order in Go |
| D-2 | Two hops in the `hops` var block, after `SubjectsOfEvent`: `EventsOfParticipant = MustHop("participation", "person", "event", nil)` *(new)* and `ParticipantsOfEvent = EventsOfParticipant.Reverse()` *(new)*. The block's comment gains one sentence: "A participation in any role joins a person to an event; walks read the role off the link." `EventsOfSubject` and `SubjectsOfEvent` are unchanged | The map's scope ("an any-role participation hop in `core/hops`, beside `EventsOfSubject`"); `NewHop` already accepts a nil filter and `Follow` skips the term check without one | Widening `EventsOfSubject` — every summary's birth, death and title reads subject role only |
| D-3 | One canonical rule for anything a section names. `shownPerson(n, personTypeID)` in `family.go` becomes `canonical(n *graphcache.Node, typeID []byte) bool` (same body: `n.Is(typeID) && len(n.AcceptedMembers()) > 0`); its doc says it is the package's one rule for a canonical entity a section names (a relative, an event, a participant), the working rule "Canonical only". `Families`' one call (inside `kinOf`, beside the local slice `shown`, which keeps its name) passes `typeID` as today. The timeline calls `canonical(n, eventTypeID)` for each event and `canonical(n, personTypeID)` for each co-participant. The person themself is checked by `Handle` then `n.Is(personTypeID)`: an Event or Place handle is `ErrNotFound`, and a provisional-only person has a timeline (AC-1.12), as `PersonFamily` does | R-2; working rule "Canonical only, filtered in Go" binds S13-05; S13-04 D-7 named this function for S13-05's participants. Events are entities a section names, so the same rule applies. `canonical`, not `shown`: `kinOf` already has a local `shown` that would hide the function (review M-2) | A second `shownEvent` — two copies of one rule; checking events with `Is` only — an event whose every claim was made provisional would still show |
| D-4 | `personTypeID(g)` is deleted. `Families` and `PersonEvents` each call `g.Vocabulary()` once and read `v.TypeIDsByKey()[vocabkey.Person]` (and, in the timeline, `[vocabkey.Event]` and `v.TermsByID()`) | One vocabulary read per call; a one-line forwarder is not worth a name (review M-7) | Renaming it `typeIDs` — still a forwarder, and the timeline would read the vocabulary twice |
| D-5 | One row per event, the person's roles on it. One private function in `timeline.go`, `rolesByNeighbor(links []graphcache.Link, terms map[string]graphcache.VocabTerm, keep func(id []byte) bool) []neighborRoles` *(new)*, with `type neighborRoles struct { id []byte; roles []conclusionsummaries.Term }` *(new)*, is the one collapse rule for both walks: links in input order; a link whose neighbor fails `keep` is skipped; neighbors are de-duplicated in first-seen order; each link adds its `TermID` to that neighbor's roles unless that id is already there; a link with no `TermID`, or whose term is missing from `terms`, adds no role (AC-1.6, AC-1.7). A role is `conclusionsummaries.Term{ID, Key, Origin, Label}` from the `VocabTerm`. The person's events are `rolesByNeighbor(g.Follow(person, hops.EventsOfParticipant), terms, keepAll)`, then each kept only when `canonical` after the load (D-6) | User, 2026-11-02 (review M-1): a row carries every role, so none is dropped silently. A participation is one association per person and event (`identityclaims.FileBridgesTx`: role stays off the key), so two links to one event come from data the app did not shape (a merge). One named function so the rule lives once (review M-4). Labels from the cached vocabulary, as `conclusiondetails.attachValues` does, with no query warm | One row per link — the same event twice on a timeline; keeping the first role only — offered to the user and declined; writing the loop in both walks — two copies of one rule |
| D-6 | Loads, in this order, all through the graph cache: (1) `conclusiondetails.Handle(c, id)`, then `n.Is(personTypeID)`; (2) the person's links (D-5) and one `g.Nodes(eventIDs)`; (3) per event that is `canonical`, `g.Follow(event, hops.ParticipantsOfEvent)` (the node is cached by step 2), collecting every neighbor that is not the person; one `g.Nodes(participantIDs)`; (4) every canonical check reads `g.Cached` only, never `g.Node`; (5) `conclusionsummaries.EventsByIDs(c, eventIDs)` with every canonical event and `conclusionsummaries.PersonsByIDs(c, participantIDs)` with every canonical participant, then `conclusionsummaries.NewSummaries(persons, events, nil)` for the handler: the same two calls `ByIDs` makes. No memo slot: nodes, the vocabulary and the summaries' display rows are already cached, so a warm read recomputes linear work with no SQL (AC-2.1) | Working rule "Each section reads its own query", read from the graph cache like the family (S13-04 D-9); a fixed number of batches cold (AC-2.2); `EventsByIDs` returns its rows in list order, which `Summaries` would lose (review M-3) | `g.Node` per participant — one query per participant cold; a memo slot — a slot and drop rule for no query saved; `ByIDs` then a re-sort — a second entry point to the event order |
| D-7 | Participants. For each event, `rolesByNeighbor(g.Follow(event, hops.ParticipantsOfEvent), terms, keep)` where `keep` is "not the person and `canonical` among cached nodes": link order is association ref, then neighbor ref, the order `EventSummary.Subjects` already uses. After the summaries are read, an event with no summary is dropped and a participant with no summary is dropped (a node dropped mid-read; the next read is right, as `dropUnread` does); an event whose participants are all dropped stays. This drop is one function, `dropUnreadEvents(events []PersonEvent, summaries conclusionsummaries.Summaries) []PersonEvent` *(new)*, so `timeline_drop_test.go` (package `persondetail`) can test it as `family_drop_test.go` tests `dropUnread` | R-27: "the other canonical people who took part, each with their role"; the family read's drop rule (S13-04 D-10) | Ordering participants by role or title — a layout call S13-D4 has not made; if the board needs another order, a follow-up adds it (working rule "Engine PRs run ahead of their slice") |
| D-8 | Order and "dated". Rows are emitted by walking `EventsByIDs`' result in its order (`sortEvents`: date, else start date, dated first, then ref); `sortEvents`, `events` and `compose.go`'s sort stay unchanged. `HasEventDate(h EventSummary) bool` *(new)* in `order.go` returns `eventKeyOf(h).ok`; `PersonEvents` sets each row's `Dated` from it. "Undated" is exactly what the order puts last (a date or start date with a sort key), so Swift's Undated group and the order cannot disagree (AC-1.4: an end date alone is undated). `Dated` stays on the timeline row, not on `EventSummary` (review M-6, declined: no other `EventSummary` reader groups by it, and the summary is a stored display row on every list wire) | The map's rule names `sortEvents` as the order; reading `EventsByIDs`' order keeps one entry point (review M-3) | An exported `SortByEventDate` — a second entry point to one sort; Swift deciding "undated" from `date == nil && startDate == nil` — a second copy of the rule that misses a date with no sort key |
| D-9 | Types in `timeline.go`: `type PersonEvent struct { Event []byte; Roles []conclusionsummaries.Term; Participants []EventParticipant; Dated bool }`, `type EventParticipant struct { Person []byte; Roles []conclusionsummaries.Term }` (the proto and Swift names, review M-8). `func PersonEvents(c *database.Catalog, entityID []byte) ([]PersonEvent, conclusionsummaries.Summaries, error)`. The returned rows and summaries agree: every `Event` and `Person` has a summary. `PersonEvents` runs, in this order: the loads (D-6), `EventsByIDs` and `PersonsByIDs`, one row per event in `EventsByIDs`' order (an event id with no summary has no row), then `dropUnreadEvents` for the participants (D-7), then each `Dated` from `HasEventDate`. `dropUnreadEvents` neither orders nor sets `Dated` | "Ids inside, summaries at the edge", as `PersonFamily` returns `Family` and `Summaries` | Returning `EventSummary` values inside `PersonEvent` — the handler would map twice the shape `Summaries` already indexes |
| D-10 | Wire: `ListPersonEvents` (method 114) returns `repeated PersonEvent events`; `PersonEvent` has `EventSummary event`, `repeated ConclusionTerm roles`, `repeated EventParticipant participants`, `bool dated`; `EventParticipant` has `PersonSummary person`, `repeated ConclusionTerm roles`. The handler maps through `eventSummaryProto`, `personSummaryProto` and `termProto`, with picture paths from one `files.RelPaths(c, summaries.PictureImageFileIDs())`, which covers event lead pictures and participant portraits. A row or participant with no summary is a broken invariant (D-9): the handler returns an error, as `familyMemberProto` does | One summary type per tile (working rule "One person summary for every tile"); `ConclusionTerm` is the one term message | A `string role_key` — Swift would label it from a list it does not cache for roles |
| D-11 | Swift key `.personEvents(project:, entityId:)` *(new)*, placed after `.personFamily` in the `CatalogQueryKey` cases, in `Kind`, in the `kind` and `project` switches, and in the descriptor switch: `stalePolicy: .sessionFresh`, `invalidateOn: conclusionTriggers`, `whenHidden: .evict`, `invalidation: .allCached(.personEvents)`, `refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)]`. Its descriptor comment says why: an edit to an event or a co-participant touches their id, not this person's, and a new participation is not in the cached value | Working rule "One invalidation rule for person-section keys" ("unless the plan finds a keyed effect": none names the person for an event edit); S13-04 D-12; the read is cheap warm (AC-2.1) | Following only the event and participant ids the value names — a new event (AC-2.3's) is not in the cached value |
| D-12 | Swift types in `GenealogyStore.swift`: `CatalogPersonEvent { event: CatalogEventSummary; roles: [CatalogConclusionValue]; participants: [CatalogEventParticipant]; dated: Bool }` (`Identifiable`, `id` is `event.entity.id`) and `CatalogEventParticipant { person: CatalogPersonSummary; roles: [CatalogConclusionValue] }`; roles map through `GoStore.mapConclusionTerm` (always `.term`), as `CatalogTermFacts.term` does. Store: `listPersonEvents(projectDir:entityID:) async throws -> [CatalogPersonEvent]`. Swift does not sort, group or title: S13-13 titles with `EventTitleDisplay` and groups by `dated` | S13-04's `Catalog*` shape; the existing summary and term mappers | A Swift role type — a fourth term shape beside `CatalogConclusionValue.term` |
| D-13 | Builders in `graphcachetest.go`: `Canon.Link` skips its `Term` call when `termProp == ""` (its doc says so), and `Canon.Location(event, place []byte) []byte` *(new)* is `k.Link("location", map[string][]byte{"event": event, "place": place}, "", "")`; AC-1.7's role-less link is `k.Link("participation", map[string][]byte{"person": pat, "event": c}, "", "")` (review M-5, X-1). Fixtures use `vocabkey.EventType("census")` (and `residence`, `travel`, `migration`, all seeded) and `vocabkey.PropertyStartDate` / `PropertyEndDate` with `Canon.Date`; add no `vocabkey` constants. Expected ties (AC-1.3, AC-1.4, AC-1.5, AC-1.6) are not taken from creation order: refs are `TST-%d` from a counter and compare as strings (`tst-10` < `tst-9`), so each tie test reads the refs back (`g.Cached(id)` or the summaries) and derives the expected order from them. Tests at each layer: `HasEventDate` in `conclusionsummaries_test.go` beside `TestSortByBirth`; `rolesByNeighbor` and `dropUnreadEvents` in `timeline_drop_test.go` *(new, package `persondetail`)*; the read in `core/database/persondetail/timeline_test.go` *(new, package `persondetail_test`)* on `kinWorld` (reused, not copied); the wire in `person_detail_test.go` with the catalog built inside `catalogsession.Do`, as `TestGetPersonFamily`; Swift in the two registry test files. The hops are covered by the read's tests (a witness appears, AC-1.1); `MustHop` panics at package init on a bad hop | Each case at the lowest layer that sees it (S13-04 D-14); one link builder, not a near-copy | A `Canon.Location` with its own body — a copy of `Link`'s; a `hops` unit test — `MustHop` already fails at init |

**Rules this plan bends.** Working rule "Timeline is direct participation, any role" (`deployment-plan.md`) says each row carries "the person's role term" and participants as {role, `PersonSummary`}; D-5 carries a list of roles for both → a merge can leave two links with two roles, and one role would drop a recorded one → user, 2026-11-02: all roles → STEP-9 rewords the rule, the map row's Scope and the S13-05 scope entry to "roles".

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `hops.EventsOfParticipant`, `hops.ParticipantsOfEvent` *(new)* | vars | `EventsOfSubject`, `SubjectsOfEvent` | Those filter to subject role and every summary relies on it (D-2) |
| `conclusionsummaries.HasEventDate` *(new)* | func | `eventKeyOf`, `SortByBirth` | Exports the "dated" half of the one event date rule; the order itself is `EventsByIDs`' (D-8) |
| `persondetail` `timeline.go` *(new)*: `PersonEvent`, `EventParticipant`, `PersonEvents`, `rolesByNeighbor`, `neighborRoles`, `dropUnreadEvents` | file, types, funcs | `family.go` `PersonFamily`, `dropUnread`; `conclusiondetails.attachValues` (term labels) | The third section read in the package the working rule names; one collapse rule for both walks (D-5); the drop is per event and participant, not per unit |
| `canonical` *(renamed from `shownPerson`)*; `personTypeID` deleted | func | itself | Widened to events (D-3); the forwarder goes (D-4) |
| `ListPersonEvents`, `PersonEvent`, `EventParticipant` *(new)* | FFI, proto | `GetPersonFamily*`, `FamilyMember`, `EventSummary`, `PersonSummary`, `ConclusionTerm` | Reuses the summaries and the term message (D-10) |
| handler `ListPersonEvents`, `personEventProto` *(new)* | funcs | `GetPersonFamily`, `familyMemberProto`, `eventSummaryProto` | Maps one more shape; reuses every summary mapper |
| `CatalogPersonEvent`, `CatalogEventParticipant` *(new)* | Swift types | `CatalogPersonFamily`, `CatalogFamilyMember` | Members reuse `CatalogEventSummary`, `CatalogPersonSummary`, `CatalogConclusionValue` (D-12) |
| `listPersonEvents` (protocol, `GoStore`, `FakeStore`), `seededPersonEvents`, `CoreMethod.listPersonEvents` *(new)* | Swift funcs, var | `getPersonFamily`, `seededPersonFamilies` | One per method, as S13-04 |
| `.personEvents` key, `Kind.personEvents`, descriptor *(new)* | key | `.personFamily` | Same refresh shape (D-11) |
| `graphcachetest.Canon.Location(event, place []byte) []byte` *(new)*; `Canon.Link` with an empty `termProp` | test helper | `Canon.Link`, `Canon.Join` | `Location` is one `Link` call; `Link` skips the term when none is given (D-13); `hops.PlacesOfEvent` already checks the `location` end keys |

Searched for an existing timeline or participation read by name (`rg -n -i 'timeline|participat' core/database api/ffi/handlers --type go -l`: `graphfill`, `promotealign` and `conclusionsummaries` walk participations for their own rows; none returns a person's events with roles) and by behavior (callers of `hops.EventsOfSubject`: `conclusionsummaries` `composePersons`, `lifePicks`, `Dependents`; all subject-role, all for a summary).

**Builds on.** `core/hops`; `graphcache.Follow`, `Nodes`, `Cached`, `Vocabulary`; S13-04's `shownPerson`, `dropUnread` and handler shape; `conclusionsummaries.EventsByIDs`, `PersonsByIDs`, `NewSummaries`, `eventKeyOf`; `files.RelPaths`; `graphcachetest.Canon`; `kinWorld`; `applyHiddenPolicyOnAny`.
**Deletes.** `persondetail.personTypeID` (→ `g.Vocabulary()` inline, D-4). `shownPerson` is renamed `canonical`, not kept beside it. Nothing else becomes dead.
**Next adds.** S13-13 reads `.personEvents`, adds `ConclusionPageSection.timeline.queryKey`, titles with `EventTitleDisplay` and groups on `dated`. An event depiction in R-26's slot reads `event.pictureImageRelPath`, already on the wire. A Place page timeline ("what happened here") is `EventsAtPlace` plus the same participant walk, if a later spike asks.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| `Graph.Follow` reached with a hop that has no filter | `core/database/graphcache/follow.go` `Follow` | The `filtered` branch is skipped; every participation link of the end is returned (AC-1.1's witness) |
| An event or participant merged, deleted or provisional meanwhile | `shown` (D-3), `dropUnreadEvents` (D-7) | Not shown (AC-1.8, AC-1.9); dropped mid-read (AC-1.13) |
| A person with no events | `PersonEvents`, handler | Zero rows, no error (AC-1.11) |
| A participation with no role, or two to one event | D-5 | No role (AC-1.7); one row, both roles (AC-1.6) |
| `.personEvents` revalidated by any touched handle | `WorkspaceSession.revalidatedKeys` | Hidden keys evict; the visible one reloads once per `apply` (AC-4.3) |

### Scale

The map's **Scale** names the 41-Subject obituary and a Source of 100 people. A timeline's size is the person's events × each event's participants. Cold `PersonEvents`: `Handle` (one node load), one `Nodes` batch of events, one `Nodes` batch of participants, then `ByIDs`' fixed batches for events (their subjects and places) and for persons (their births, deaths and places). AC-2.2 proves the count equal for one event with one co-participant and for twelve events with a 41-participant census among them. Warm: zero graph queries (AC-2.1); the handler adds one `files.RelPaths` query when any summary leads with a picture, as `GetPersonFamily` does. Work is linear in links read. Writes: none. Swift: every write that touches any handle reloads a visible Timeline section once (D-11), the same cost the Family key already carries.

### Edges

- **Assumptions.** `Follow` returns links in association-ref order, so event and participant order before sorting is stable across reads (AC-2.1's equality). `sortByDateKey` is stable and ties break by ref, so the timeline order is total. Graph loads skip merged neighbors (`load.go`, `merged_into_id IS NULL`), and `shown` checks `Merged` again so the rule does not rest on the SQL. A participation's `TermID` is the rank-1 kept role (`disambiguation`), one per link.
- **Neighbors.** In: any role (subject, father, mother, witness, informant, pallbearer, a researcher's own role term), an event dated by start date only. Out: a child's birth the person is not linked to (AC-1.2), an event linked only to a merged or provisional-only entity (AC-1.9). Undated: no date and no start date, or an end date only (AC-1.4).
- **Removed inputs.** A participant not shown is left off the event; the event stays. An event not shown is left off the timeline with all its participants. A removed participation drops on commit and the next read loses it; a new one appears (AC-2.3).
- **Real data.** The seed has seven `role` terms (`subject`, `father`, `mother`, `child`, `witness`, `informant`, `pallbearer`); none is special-cased. A test row uses `witness` and one uses `informant`.
- **Untrusted size.** Events per person and participants per event are bounded only by the catalog. Each costs one map lookup and one cached-node check; summaries are read in fixed batches (AC-2.2). Duplicate links to one event or one participant collapse (D-5, D-7). No dimension multiplies past the links read.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/hops` | `hops.go`: `EventsOfParticipant`, `ParticipantsOfEvent` *(new)*; the var block's comment (D-2) |
| `core/database/conclusionsummaries` | `order.go`: `HasEventDate` *(new)* (D-8). `compose.go` unchanged |
| `core/database/persondetail` | `family.go`: `shownPerson` → `canonical`, `personTypeID` deleted (D-3, D-4). `timeline.go` *(new)*: `PersonEvent`, `EventParticipant`, `PersonEvents`, `rolesByNeighbor`, `neighborRoles`, `dropUnreadEvents` (D-5 – D-9). Package doc unchanged |

```go
// HasEventDate reports whether the event order (EventsByIDs, ListEvents)
// places h among the dated rows: a date, else a start date, with a sort key.
func HasEventDate(h EventSummary) bool

// PersonEvent is one event a person took part in directly: the person's roles
// on it, the other canonical participants, and whether it is dated.
type PersonEvent struct {
	Event        []byte
	Roles        []conclusionsummaries.Term
	Participants []EventParticipant
	Dated        bool
}

// EventParticipant is one other canonical person on the event and their roles.
type EventParticipant struct {
	Person []byte
	Roles  []conclusionsummaries.Term
}

// PersonEvents is the Timeline section's read: every canonical event the
// person took part in, in any role, in event date order, and the summary of
// every event and participant it names. An unknown or merged handle, or one
// that is no Person, is conclusiondetails.ErrNotFound.
func PersonEvents(c *database.Catalog, entityID []byte) ([]PersonEvent, conclusionsummaries.Summaries, error)
```

Errors: `conclusiondetails.ErrNotFound` (`conclusiondetails.not_found`); otherwise the loads' errors.

### FFI

`api/proto/engine.proto`: `METHOD_LIST_PERSON_EVENTS = 114` *(new)*. Handler `api/ffi/handlers/person_detail.go` `ListPersonEvents` *(new)*; dispatch `MethodListPersonEvents` *(new)*. Regenerate with `./scripts/generate-proto.sh`.

```proto
// ListPersonEventsRequest asks for the events a canonical person took part in.
// An unknown or merged handle, or one that is no Person, is
// conclusiondetails.not_found.
message ListPersonEventsRequest {
  string project_dir = 1;
  string entity_id = 2;
}

// ListPersonEventsResponse is every canonical event the person took part in
// directly, in any role: dated first, earliest first, then by ref, the order
// of ListEventSummaries. Canonical events and participants only.
message ListPersonEventsResponse {
  repeated PersonEvent events = 1;
}

// PersonEvent is one event of a person's timeline: the event, the person's
// roles on it (one per distinct role term, usually one; empty when the
// participation has none), the other canonical participants in participation
// order, and dated, false for an event the order puts last.
message PersonEvent {
  EventSummary event = 1;
  repeated ConclusionTerm roles = 2;
  repeated EventParticipant participants = 3;
  bool dated = 4;
}

// EventParticipant is one other canonical person on an event and their roles.
message EventParticipant {
  PersonSummary person = 1;
  repeated ConclusionTerm roles = 2;
}
```

### macOS

None — no view. The session and store layer only:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire types | `CatalogPersonEvent`, `CatalogEventParticipant` *(new)* | `macos/App/Platform/GenealogyStore.swift` | D-12 |
| Store | `listPersonEvents(projectDir:entityID:) async throws -> [CatalogPersonEvent]` *(new)* | `GenealogyStore.swift` protocol; `GoStore.swift`; `FakeStore.swift` (`seededPersonEvents: [String: [CatalogPersonEvent]]`, unseeded → `conclusiondetails.not_found`, no `recordedCalls`) | As `getPersonFamily` |
| Method | `CoreMethod.listPersonEvents` *(new)* | `macos/App/Platform/CoreInvoke.swift` | |
| Key | `.personEvents(project:, entityId:)`, `Kind.personEvents` *(new)* | `CatalogQueryKey.swift`; descriptor in `CatalogQueryRegistry.swift` | D-11 |

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Any-role participation hops**
Files: `core/hops/hops.go`.
Do: D-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/hops/ ./core/database/graphcache/... ./core/database/conclusionsummaries/` passes.

**STEP-2 — Export the "dated" rule**
Files: `core/database/conclusionsummaries/order.go`, `conclusionsummaries_test.go`. Skill: `add-catalog-query`; `.cursor/rules/go-tests.mdc`.
Do: `HasEventDate` (D-8). Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/` passes with no existing test changed; `git diff --stat` shows no change to `compose.go`.

**STEP-3 — One canonical rule for every section**
Files: `core/database/persondetail/family.go`.
Do: D-3's rename and D-4's deletion only; no behavior change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/` passes with no test changed.

**STEP-4 — The timeline read**
Files: `core/database/persondetail/timeline.go` *(new)*, `timeline_test.go` *(new)*, `timeline_drop_test.go` *(new)*, `core/database/graphcache/graphcachetest/graphcachetest.go` (`Canon.Link` with an empty `termProp`, `Canon.Location`). Skill: `add-catalog-query`.
Do: D-5 – D-9. Add T-2, T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/` passes; `go list -deps ./core/database/persondetail | grep api/ffi/handlers` prints nothing.

**STEP-5 — Wire messages**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: §6 FFI. Run `./scripts/generate-proto.sh`.
Check: `go build ./...` passes; the generated files hold `METHOD_LIST_PERSON_EVENTS`.

**STEP-6 — The handler**
Files: `api/ffi/handlers/person_detail.go`, `api/ffi/dispatch.go`, `api/ffi/handlers/person_detail_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-10. Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-7 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (its `ThrowingStore` conformer gains `listPersonEvents`, throwing `StoreBoom.boom` as its `getPersonFamily` does). Skill: `add-swift-test`.
Do: D-12 and §6 macOS rows 1 – 3.
Check: `xcodebuild build` for the app scheme succeeds with no new warning.

**STEP-8 — The cache key**
Files: `macos/App/Features/Workspace/Session/CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-11. Add T-6 – T-8; update `invalidationsComeFromRegistrySpecs` and `everyKindHasOneDescriptor` (T-9).
Check: `xcodebuild test` passes the two suites.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-13/deployment-plan.md` (map row: its Plan column links this plan from planning, Status → `Landed #N`; the S13-05 scope entry: "`ListPersonEvents(personId)` in `core/database/persondetail`" → "`ListPersonEvents` RPC over `persondetail.PersonEvents`"; the working rule "Timeline is direct participation, any role", the map row's Scope and the S13-05 scope entry: and "the person's role term" / "{role, `PersonSummary`}" → "the person's roles" / "{roles, `PersonSummary`}" (Rules this plan bends)), `docs/ideas/structural-debt.md` (two rows, next free A-numbers: the two event-person shapes, `EventSubject` from `hops.SubjectsOfEvent`, subject role and not filtered to canonical, and `EventParticipant` from `hops.ParticipantsOfEvent`, canonical, to fold when an Event page lists participants; and the per-person query keys `.personFamily`, `.personEvents` and S13-27's sharing one descriptor builder, with `everyKindHasOneDescriptor` deriving its count from `Kind.allCases`, touched with S13-27), `docs/deployment-plan/spike-13/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-05-person-timeline-read.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestHasEventDate` *(new)* | Table: a date 1850 → true; a start date only → true; an end date only → false; no date → false; no date and no start date but a `datevalue.Value` with no year (`Kind: datevalue.KindPoint`, `StartYear` nil) as the date → false | AC-1.4, D-8 |
| T-2 | Go | `core/database/persondetail/timeline_test.go` (package `persondetail_test`) `TestPersonEvents` *(new)* | One subtest per AC-1.1 – AC-1.12, on `kinWorld` with `Canon.Event`, `Join`, `Birth`, `Marriage`, `Date`, `Merge`, `Provisional`, `Commit`; assertions on event ids in order, role keys and labels, participant ids in order, `Dated`; every returned id has a summary | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, D-3, D-5, D-7, D-9 |
| T-3 | Go | `timeline_test.go` `TestPersonEventsQueryCount` *(new)* | AC-2.1; AC-2.2 (a) vs (b), each event located at its own Place (`Canon.Location`); AC-2.3 | AC-2.1 – AC-2.3, D-6 |
| T-4 | Go | `core/database/persondetail/timeline_drop_test.go` (package `persondetail`) `TestDropUnreadEvents`, `TestRolesByNeighbor` *(new)* | `TestRolesByNeighbor`: links to A (subject), A (informant), A (subject again), B (no term), C (a term missing from `terms`), D (`keep` false) → A with subject then informant, B with none, C with none, no D, in that order. `TestDropUnreadEvents`: two events, one with no summary, the other with two participants, one with no summary → one event, one participant; an event whose only participant has no summary stays with none | AC-1.6, AC-1.7, AC-1.13, D-5, D-7 |
| T-5 | Go | `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents` *(new)* | AC-3.1 on a project from `onboarding.Complete`, the catalog written with `graphcachetest.NewCanon` inside `catalogsession.Do` (as `TestGetPersonFamily`); AC-3.2 for an unknown id, a malformed id and an Event's id | AC-3.1, AC-3.2, D-10 |
| T-6 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personEventsInvalidateOnPromote` *(new)* | `.promotedSubject` → contains `.allCached(.personEvents)`; key kind, project, `.evict`, `.sessionFresh`; `Kind.personEvents.descriptor.refreshes.map(\.follows) == [.touchedHandles]` | AC-4.2, AC-4.3, D-11 |
| T-7 | Swift | `CatalogQueryRegistryTests.swift` `personEventsLoadThroughTheStore` *(new)* | AC-4.1 with one seeded event | AC-4.1, D-12 |
| T-8 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personEventsFollowEveryTouchedHandle` *(new)* | A warmed key for e1; `revalidatedKeys(for: CatalogWriteEffects(handles: ["e7"]))` names it; `CatalogWriteEffects.empty` names nothing | AC-4.3 |
| T-9 | Swift | `CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs`, `everyKindHasOneDescriptor` (changed) | Three places in `invalidationsComeFromRegistrySpecs`: the `.deletedSource` list and the `.promotedSubject` list each gain `.allCached(.personEvents)` after `.allCached(.personFamily)`, and the trigger loop gains `#expect(invalidations.contains(.allCached(.personEvents)) == isTrigger, "\(mutation)")`; `everyKindHasOneDescriptor` expects 21 kinds | AC-4.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestPersonFamily*` (STEP-3) and `invalidationsComeFromRegistrySpecs`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected: no copy)
- [ ] Sweep: `shownPerson` and `personTypeID` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `docs/deployment-plan/spike-13/deployment-plan.md` and `L10n.swift` with no hit left outside landed plans and ship logs; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed — none changes; the map's working rule says "roles" (STEP-9); `docs/ideas/structural-debt.md` gains its two rows
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No view reads the timeline until S13-13; the reviewer checks the read and that nothing visible moved:

1. `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/ ./core/database/persondetail/ ./api/ffi/handlers/ -run 'HasEventDate|PersonEvents|DropUnreadEvents|RolesByNeighbor|ListPersonEvents' -v` and read T-2's subtest names against §2.
2. Open the app on the product owner's project: a Person page looks and behaves as before (nothing calls `listPersonEvents` yet).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `ByIDs` or `Vocabulary` issues a query on a warm read | AC-2.1 counts the whole `PersonEvents` |
| A cold read loads participants one by one | AC-2.2 grows the participants to 41 on one event and the events to twelve |

**Stop and report if:**

- `Graph.Follow` with `EventsOfParticipant` returns a link whose `BridgeTypeKey` is not `participation`, or misses a non-subject role link (AC-1.1).
- A second `PersonEvents` read issues a graph query (AC-2.1) and the query comes from code outside `persondetail` and `conclusionsummaries`.
- The cold counts in AC-2.2 differ, and the difference comes from `conclusionsummaries`' summary composition (a per-row load there is outside this PR).
- Renaming `shownPerson` or deleting `personTypeID` changes any `persondetail` test result (STEP-3).
- `EventsByIDs` no longer returns its rows in `sortEvents` order (D-8 relies on its doc).
- `graphcachetest.NewCanon` fails inside `catalogsession.Do` on an onboarded project (T-5).
- `go list -deps ./core/database/persondetail` lists a handler package.
- `protoc`, `protoc-gen-go` or `protoc-gen-swift` is missing (STEP-5), or `xcodebuild` is not available (STEP-7, STEP-8, check-pr): the Swift and proto steps run on macOS; report and do not hand-edit generated files.

## 12. Open questions

None.

## 13. PR

**Title:** S13-05 — Person timeline read
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-05-person-timeline-read.md`.
