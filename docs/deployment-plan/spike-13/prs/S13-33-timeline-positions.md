# S13-33 — Timeline positions

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine read's new fields, their wire and their Swift mapping; nothing draws them until S13-13)
**Board:** —
**Depends on:** S13-05 (landed [#680](https://github.com/mendahu/provenencia/pull/680))
**Blocks:** S13-13
**Branch:** `s13-33-timeline-positions`
**Runs on:** macOS: STEP-4 needs `protoc` and its plugins (`./scripts/generate-proto.sh`), and STEP-5 and `scripts/check-pr.sh` need Xcode. STEP-1 – STEP-3 run on Linux with Go
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-05 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-23, R-24, R-24a); working rules: "The timeline is a lifeline", "Timeline is direct participation, any role", "Each section reads its own query", "Engine PRs run ahead of their slice" (data, not layout); the S13-05 plan [`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md) and its **Landed** note
**Planned:** 2026-11-02 against `main` @ `51dccd4`
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

The engine says where each dated event of a person's life sits on a line of years, and where the life's birth and death markers go, so the Timeline (S13-13) can draw a lifeline to scale and Swift only scales numbers (working rule "The timeline is a lifeline", user 2026-11-02). Each dated row of `ListPersonEvents` gains its span on the year line; the response gains a birth and a death marker, each the recorded event or a placeholder that names the dated event it sits beside and invents no date.

**Why now.** S13-13's lifeline needs positions, and the working rule puts them in Go; this lands before S13-13.

**What changes beyond the timeline** (user, 2026-11-02: one meaning of "dated" everywhere). An event whose date has an open start but a known end ("bef. 1880", a range with no start year) becomes dated in `conclusionsummaries`' one date-order rule and is ordered at its known end. So it moves out of the undated tail of the Events list (and `EventsByIDs`, and the timeline) to its place among 1880's events; a birth dated that way joins birth order (Family's children and siblings, `SortByBirth`) and can become the header's lead birth or death. `autoreconcile.SortKey` and everything the reconciler does are unchanged.

**What stays.** `ListPersonEvents`' roles and participants (S13-05); matching's lifespan (`subjectprofile`, `subject.Lifespans`), which is not used; `autoreconcile.SortKey`; every stored row.

## 2. User stories

**US-1.** As the Timeline section, I get each dated event's place on a line of years, computed by the same rule that orders and dates events.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An event dated 12 Mar 1817 | `span` from 1817 + 70/365 to 1817 + 71/365 (`DateEnd.From` and `.To` of that day), `open_end` false |
| AC-1.2 | An event dated "1835" | `span` [1835, 1836) |
| AC-1.3 | An event with the range 1844 – 1849 | `span` [1844, 1850) |
| AC-1.4 | An event dated "aft. 1870" | `span` from 1870 to 1871, `open_end` true |
| AC-1.5 | "abt. 1846" | The same span as "1846" (the qualifier ABT does not widen it) |
| AC-1.6 | An event whose only date is "bef. 1880" | `dated` true; `span` from 1880 to 1881 with `open_start` true; in the read's order among 1880's events, by ref (D-2, D-3) |
| AC-1.8 | A range with no start year, ending 1849 | `dated` true; `span` [1849, 1850) with `open_start` true (D-3) |
| AC-1.9 | An event with only an end date, or a date with no year | No `span`, `dated` false: still undated (S13-05 AC-1.4 unchanged) |
| AC-1.10 | The Events list with events dated 1879, "bef. 1880", 1881 and an undated one | Order 1879, bef. 1880, 1881, then the undated one (before this PR: 1879, 1881, bef. 1880, undated) (D-2) |
| AC-1.7 | Every row of every fixture | `span` present exactly when `dated` is true |

**US-2.** As the Timeline section, I always get a birth and a death marker, placed from the records, with no invented date.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Pat: birth B (1820, subject), census C (1851), death D (1888), burial U (1890) | `birth.event` = B, `death.event` = D; both `beside` empty |
| AC-2.2 | Pat with no birth event; census C (1835) first, then a marriage | `birth.event` empty, `birth.beside` = C |
| AC-2.3 | Pat with census C (1875) and funeral F (1880), no death | `death.event` empty, `death.beside` = C (F may follow death; working rule "The timeline is a lifeline") |
| AC-2.4 | Pat with only undated events | Both markers empty, both `beside` empty |
| AC-2.5 | Pat with an undated birth B (no date at all) and a census C (1835) | `birth.event` = B, `birth.beside` = C (B has no span of its own) |
| AC-2.6 | Two birth events with Pat as subject, 1820 and 1822 | `birth.event` = the 1820 one, the event `PersonSummary.Birth.Event` names |
| AC-2.7 | Pat is a witness at someone's birth | That birth is not Pat's birth marker (subject role only) |
| AC-2.8 | Death D dated 1888 before a census C dated 1890 | `death.event` = D; C keeps its own span; nothing is flagged |

**US-3.** As the app, I read the new fields and keep working against an engine that does not send them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A response with spans and markers | `CatalogPersonTimeline` carries them; `CatalogPersonEvent.span` set on dated rows |
| AC-3.2 | A response with no `birth` or `death` (an older engine) | Both markers are placeholders with no `beside` |

**US-4.** As a researcher reading an event, I meet the people who mattered most to it first.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A baptism with participants Witness Kohl, Mother Anna, Father Heinrich (that link order) | Heinrich (father), Anna (mother), Kohl (witness) (D-10) |
| AC-4.2 | A burial with pallbearers Zimmer and Abel, and a participant with the researcher's role "Officiant" | Abel, Zimmer (pallbearer, by name), then the officiant (a role not in the list) |
| AC-4.3 | One participant with roles witness and informant | Ranked as a witness (the best of its roles) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One date-order rule, open start ordered at its known end, shared by the event order, birth order, the life lead and the span (D-2, D-3) | STEP-1, AC-1.1 – AC-1.10 |
| The after-death / before-birth table and `vocabkey.EventFuneral` (D-6) | STEP-2, AC-2.3 |
| Spans and markers on `persondetail.PersonEvents` (D-4, D-5, D-7) | STEP-3, AC-2.1 – AC-2.8 |
| Wire and handler (D-8) | STEP-4 |
| Swift types, mapping, store, key value type (D-9) | STEP-5, AC-3.1, AC-3.2 |
| Each event's participants ordered by role, then by name (D-10; board S13-D4 turn 12, user 2026-11-02) | STEP-3, AC-4.1 – AC-4.3 |

| Out | Home |
| --- | --- |
| The lifeline view | S13-13 |
| An estimated year for an unknown birth or death; flags on events outside the life | not planned (user, 2026-11-02) |
| Reading a Julian-calendar date as Julian | not planned; `DateEnds` reads every date as Gregorian (Edges) |
| Matching's lifespan window or its knobs | not planned (working rule "The timeline is a lifeline") |
| Any change to `autoreconcile.SortKey`, `valuecompare.DateWindow` or the reconciler | not planned (user, 2026-11-02) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline read | `core/database/persondetail/timeline.go` `PersonEvents(c, entityID) ([]PersonEvent, conclusionsummaries.Summaries, error)`, `PersonEvent{Event, Roles, Participants, Dated}`, `dropUnreadEvents` | Rows in `EventsByIDs` order; `Dated` from `conclusionsummaries.HasEventDate` after the drop |
| Dated rule | `core/database/conclusionsummaries/order.go` `HasEventDate`, `SortByBirth`, `sortByDateKey`; `compose.go` `eventKeyOf`, `birthKeyOf`, `dateSort`, `keptDateAt` (its key feeds `eventDate(v, ev)` → `lifeFrom` → `leadLife` and `lifeOwners`), `sortEvents` (called by `EventsByIDs` and `ListEvents`) | `eventKeyOf`: `dateSort(h.Date)`, else `dateSort(h.StartDate)`; `birthKeyOf`: `dateSort(h.Birth.Date)`; `keptDateAt` calls `autoreconcile.SortKey` itself. All are `valuecompare.DateWindow(d).Lo`: a date whose first end is open (BEF, a range with no start year) has no key and is undated. Outside the package only `persondetail` (`HasEventDate` in `timeline.go`; `SortByBirth` in `family.go`) reads them; `autoreconcile.SortKey`'s other callers are the reconciler's and are not touched |
| Date ends | `core/valuecompare/datewindow.go` `DateEnd{Year, Month, Day, Lo, Hi, From, To, Mid, Open}`, `DateEnds(d)` | `Lo`/`Hi` are proleptic Gregorian day numbers (1970-01-01 = 0); `From`/`To` place a bound on a line of years (a whole year is `[y, y+1)`); "Civil components are read as Gregorian"; ABT is not widened; AFT has an open last end |
| Lead birth / death | `core/database/conclusionsummaries/compose.go` `lifePicks` (subject-role events whose `event_type` key is `birth` or `death`), `leadLife` (the least `dateKey`: dated first, earliest, then ref) | The `PersonSummary.Birth` / `.Death` the header shows |
| Event type | `conclusionsummaries.EventSummary.Type` (`TermFacts{Term *Term}`), `Term{ID, Key, Origin, Label}` | — |
| Vocab keys | `core/vocabkey/vocabkey.go` `EventBirth`, `EventDeath`, `EventBurial`, `EventMarriage`, `RoleSubject`; no `EventFuneral`. `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` holds each to a seed term; `registry.go` seeds `event_type` `funeral` | — |
| Lifespan (matching) | `core/subjectprofile/subjectprofile.go` `lifespan`; `core/matchregistry/subject.go` `Lifespans` | Pads an unknown end by `MaxYears` (110) and `SlackYears` (10); not for display |
| Wire | `api/proto/engine.proto` `ListPersonEventsResponse{events = 1}`, `PersonEvent{event = 1, roles = 2, participants = 3, dated = 4}` | — |
| Handler | `api/ffi/handlers/person_detail.go` `ListPersonEvents`, `personEventProto` | Maps rows in order |
| Swift | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonEvent{event, roles, participants, dated}`, protocol `listPersonEvents(projectDir:entityID:) -> [CatalogPersonEvent]`; `GoStore.swift` `listPersonEvents`; `FakeStore.swift` `seededPersonEvents`, `listPersonEvents`; `CatalogQueryRegistry.swift` `.personEvents` load | — |
| Tests | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestHasEventDate`; `core/database/persondetail/timeline_test.go` `TestPersonEvents` (`kinWorld`, `w.event(name, type, year)`, `w.k.Join`), `TestPersonEventsQueryCount`; `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (builds `CatalogPersonEvent`, `QueryHandle<[CatalogPersonEvent]>`), `WorkspaceSessionTests.swift` (`[CatalogPersonEvent]()`), `OnboardingModelTests.swift` (a store stub's `listPersonEvents`) | — |

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) added two fields to an existing read (`GetPersonFamilyResponse.person`, `origin_siblings`), changed `PersonFamily`'s return value, and mapped them with Swift defaults for an older engine; copy that shape. S13-05 ([#680](https://github.com/mendahu/provenencia/pull/680)) is the read and its tests to extend.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Positions are on the **year line**: `from` and `to` as doubles from `valuecompare.DateEnd.From` and `.To` (a whole year is `[y, y+1)`), not day numbers | Swift draws year ticks at whole numbers with no date arithmetic (working rule "The timeline is a lifeline": Swift only scales) | Day numbers (`Lo`/`Hi`): Swift would convert days to years, a second calendar rule |
| D-2 | One date-order rule for the package (user, 2026-11-02). `dateSort(v *datevalue.Value) (string, bool)` in `compose.go` keeps `autoreconcile.SortKey`'s key when it has one; otherwise, when `valuecompare.DateEnds(*v)` gives an open first end and a closed last end, it returns the key of the known end: `autoreconcile.SortKey` of `valuecompare.EndsDate(last, last)`, so the key format stays the reconciler's and `SortKey` is not changed. `keptDateAt` computes its key through `dateSort` instead of calling `SortKey` itself. So `eventKeyOf` (event order, `HasEventDate`), `birthKeyOf` (`SortByBirth`) and `lifeFrom` (the lead birth and death, `lifeOwners`) share it. `eventDating(h EventSummary) (*datevalue.Value, dateKey)` *(new, unexported)* is the field choice `eventKeyOf` makes today (`Date` when its key is ok, else `StartDate`); `eventKeyOf` returns its key and `EventSpan(h EventSummary) (from, to float64, openStart, openEnd, ok bool)` *(new, `order.go`, beside `HasEventDate`)* reads its date through `valuecompare.DateEnds`. `ok` is exactly `HasEventDate(h)` | One meaning of "dated" for the Events list, the timeline, birth order and the header; the reconciler keeps its own key | A timeline-only rule (two meanings of "dated"); changing `autoreconcile.SortKey` (the reconciler reads it); a second key format in `conclusionsummaries` |
| D-3 | A date with an open start and a known end is ordered at its known end: "bef. 1880" at 1 Jan 1880 (the known bound's first day), "bef. 12 Mar 1880" at 12 Mar 1880, among that day's events by ref. Its span is `from` = the known end's `From`, `to` = its `To`, `open_start` true. A date with both ends open, only an end date, or no year stays undated | The user's answer: on the line with an open start, ordered at its known end | Ordering it at an estimated earlier date (a date no record gives) |
| D-4 | `PersonEvent` gains `Span *Span` *(new field)*, `type Span struct { From, To float64; OpenStart, OpenEnd bool }` *(new)*, set with `Dated` in the loop that sets `Dated` today, from `conclusionsummaries.EventSpan`; nil when not dated | Same place `Dated` is set, after the drop | A separate pass |
| D-5 | `PersonEvents` returns `(Timeline, conclusionsummaries.Summaries, error)`, `type Timeline struct { Events []PersonEvent; Birth, Death LifeMarker }` *(new)*, `type LifeMarker struct { Event []byte; Beside []byte }` *(new)*. `Event` is the recorded marker's event, nil for a placeholder. `Beside` is set whenever the marker has no span of its own (a placeholder, or a recorded event that is undated): for birth, the first row in order with a span whose type is not in `lifeline.beforeBirth`; for death, the last row in order with a span whose type is not in `lifeline.afterDeath`; nil when no row qualifies. Computed by `lifeMarkers(events []PersonEvent, summaries) (birth, death LifeMarker)` *(new, unexported, `lifeline.go`)* after spans are set | The user's rule (2026-11-02): every timeline has a spot for birth and death, placed by which events may come before birth or after death; no date invented | Placing placeholders in Swift (a domain rule in the client); an estimated year from matching's window (pads 110 years) |
| D-6 | The recorded birth marker is the **first** row in order whose event type key is `birth` and whose `Roles` hold the key `subject`; the death marker, the **first** such row with type `death`. Keys are compared as `lifePicks` compares them (key only). With rows in `sortEvents` order and `lifeFrom` keyed by the same `dateSort` (D-2), this is the same pick as `leadLife` (least `dateKey`), a BEF-only birth included, so it equals `PersonSummary.Birth.Event` / `.Death.Event` whenever that event is canonical; T-3 holds the equality | The marker agrees with the header the page already shows | The latest death; a second pick rule |
| D-7 | The table is one declarative value in `core/database/persondetail/lifeline.go` *(new)*: `var lifeline = struct { afterDeath, beforeBirth []vocabkey.EventType }{afterDeath: []vocabkey.EventType{vocabkey.EventBurial, vocabkey.EventFuneral}}` with a doc comment that it only places placeholders and judges no event. `vocabkey.EventFuneral EventType = "funeral"` *(new)* joins the seeded-key block, and `TestVocabKeysMatchSeed` gains its row, so every table entry is a seeded `event_type` term | Page-only, beside the read that uses it (as the sibling rule is page-only); typed keys held to the seed by the existing guard | `core/matchregistry` (matching's registry; the working rule keeps matching's lifespan out); a seeded flag on `event_type` terms (a catalog change for one page rule); string literals |
| D-8 | Wire. `PersonEvent` gains `TimelineSpan span = 5;` *(new)*, `message TimelineSpan { double from = 1; double to = 2; bool open_start = 3; bool open_end = 4; }` *(new)*, absent on undated rows. `ListPersonEventsResponse` gains `LifeMarker birth = 2; LifeMarker death = 3;`, `message LifeMarker { string event_id = 1; string beside_event_id = 2; }` *(new)*, ids as `database.UUIDString`, empty for none. Comments state D-2's and D-5's rules. The handler maps them; regenerate with `./scripts/generate-proto.sh` | The S13-12 shape (fields added to the read, older engines tolerated) | A new RPC (working rule "Each section reads its own query": one key) |
| D-9 | Swift. `struct CatalogTimelineSpan: Sendable, Equatable { var from: Double; var to: Double; var openStart: Bool; var openEnd: Bool }` *(new)*; `CatalogPersonEvent` gains `var span: CatalogTimelineSpan? = nil`; `struct CatalogLifeMarker: Sendable, Equatable { var eventId: String? = nil; var besideEventId: String? = nil }` *(new)*; `struct CatalogPersonTimeline: Sendable, Equatable { var events: [CatalogPersonEvent] = []; var birth = CatalogLifeMarker(); var death = CatalogLifeMarker() }` *(new)*. The store method becomes `listPersonEvents(projectDir:entityID:) async throws -> CatalogPersonTimeline` and the `.personEvents` key's value type is `CatalogPersonTimeline`; `GoStore` maps `span` when `hasSpan`, markers when `hasBirth` / `hasDeath`, an empty id string to nil. `FakeStore.seededPersonEvents` becomes `[String: CatalogPersonTimeline]`; the stub in `OnboardingModelTests` and the tests that build `[CatalogPersonEvent]` follow | The markers belong to the timeline, not to a row; defaults keep an older engine working (AC-3.2) | Markers as flags on rows (a placeholder has no row); keeping `[CatalogPersonEvent]` and a second key |

| D-10 | Participant order (user, 2026-11-02, from board S13-D4 turn 12). Each event's participants are ordered by **role rank**, then by **name**, then by ref. Rank comes from one declarative list in `lifeline.go`, `var participantRoleOrder = []vocabkey.Role{subject, father, mother, child, witness, informant, pallbearer}` (the seeded `role` terms, typed keys added to `vocabkey` beside `RoleSubject` and held to the seed by `TestVocabKeysMatchSeed`, as D-7); a participant's rank is the best rank among its roles; a role not in the list (a researcher's own) ranks after every listed one, and among those by the role's label. Name order is `conclusionsummaries.PersonsByIDs`' order, which the read already loads (it is title order, `SortByTitle` with the catalog's name format): `PersonEvents` keeps an index from that result and sorts each event's `Participants` stably by (rank, index). Applied after `dropUnreadEvents`. The person's own roles and the event order do not change | The board orders by role ("widow, children, officiant, mourners A–Z"); order is data, decided in Go (working rule "Engine PRs run ahead of their slice"); the title order has one owner (`SortByTitle`), reached through `PersonsByIDs` with no second name sort | Sorting in Swift; a rank stored on the seeded terms (a catalog change for one page rule); the board's "widow" and "officiant", which are not seeded roles (a researcher's own role sorts after the seeded ones) |

**Rules this plan bends.** None. S13-05 D-8 made "undated" what the order puts last with `autoreconcile.SortKey`'s key; D-2 widens "dated" to a known end with the user's answer (2026-11-02), keeping one rule, so S13-05's plan stays true in shape (one rule for the order and `Dated`).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `eventDating` *(new, unexported)*; `EventSpan` *(new)*; `dateSort`'s known-end fallback | func | `eventKeyOf`, `HasEventDate`, `dateSort`, `keptDateAt`, `valuecompare.DateEnds` / `EndsDate` | Extends them in place; `keptDateAt` drops its own `SortKey` call for `dateSort` |
| `persondetail.Span`, `Timeline`, `LifeMarker`, `lifeMarkers` *(new)* | type, func | `PersonEvent`, `Family` | The timeline's own result |
| `lifeline` table *(new)* | var | `matchregistry` `Lifespans` | Matching's; the page rule is not a matching knob |
| `vocabkey.EventFuneral` *(new)* | const | `vocabkey.EventBurial` | One more seeded key code reads by name |
| `TimelineSpan`, `LifeMarker` *(new)* | proto | `PersonEvent`, `OriginSibling` | New data |
| `CatalogTimelineSpan`, `CatalogLifeMarker`, `CatalogPersonTimeline` *(new)* | Swift types | `CatalogPersonFamily` | Same pattern: the read's value as one struct |

**Builds on.** `valuecompare.DateEnds`, `conclusionsummaries.HasEventDate` / `eventKeyOf`, `persondetail.PersonEvents`, `TestVocabKeysMatchSeed`, the `.personEvents` key (S13-05).
**Deletes.** None — nothing becomes dead. `[CatalogPersonEvent]` as the key's value type is replaced by `CatalogPersonTimeline`.
**Deletes (cont.).** `keptDateAt`'s direct `autoreconcile.SortKey` call (now through `dateSort`).
**Next adds.** S13-13 draws the line from `CatalogPersonTimeline`. Another event type allowed after death (probate) is one table entry and one `vocabkey` constant.

### Reach

None — no new path into existing code; the read already serves this key, and no view reads the new fields until S13-13.

### Scale

At the brief's sizes (a 120-event life, events with 41 participants): `EventSpan` is one `DateEnds` per row, `lifeMarkers` one pass over the rows; no query is added (spans come from the summaries already read). `TestPersonEventsQueryCount` (warm: zero queries; cold: constant batches) stays green unchanged.

### Edges

- **Assumptions.** Rows arrive in `sortEvents` order (S13-05 D-8), so "first" and "last" rows with a span are the earliest and latest by sort key (the first end, or the known end of an open start); spans may overlap (a long range, an open start), which only the view meets. `EventSpan`'s `ok` equals `HasEventDate` because both take the date from `eventDating` and `dateSort` has a key exactly when at least one end is closed and `DateEnds` is ok (T-1 asserts the equality on every case). The key format is `autoreconcile.SortKey`'s, so known-end keys compare with every other key.
- **Order moves** (user, 2026-11-02). Every reader of `dateSort` moves for an open-start date: `EventsByIDs` and `ListEvents` (the Events list, Promote's and search's event rows that come through them), `HasEventDate` (the timeline's `Dated`), `SortByBirth` (Family's children and siblings), `leadLife` (the header's birth and death when a person has several), `lifeOwners` (the order of a life row's provenance owners). Before: such a row came last with the undated. After: at its known end. No stored value changes; the graph cache's display rows are rebuilt per process.
- **Neighbors.** ABT → the same span as the bare date; AFT → `open_end`; a range with no end year → `open_end`; BEF or a range with no start year → `open_start`, ordered at its known end; both ends open, only an end date, or a day that does not exist → undated.
- **Removed inputs.** A birth event unpromoted or merged: the next read picks the next birth row or a placeholder (the key refreshes on any touched handle). A lead birth that is not canonical: the header shows it and the timeline does not; the marker falls to the next canonical birth row or a placeholder (accepted; T-3 asserts equality on canonical events only).
- **Real data.** The table names two seeded types; `TestVocabKeysMatchSeed` holds both to the seed.
- **Untrusted size.** None new: rows are bounded as S13-05 bounds them; the marker pass is linear. A date stored with `calendar: julian` is placed as if Gregorian (up to 13 days off), the same as the order: accepted at the line's scale.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `EventFuneral EventType = "funeral"` *(new)* |
| `core/database/conclusionsummaries` | `dateSort` gains the known-end fallback; `keptDateAt` keys through it; `eventDating` *(new)*; `eventKeyOf` calls it; `func EventSpan(h EventSummary) (from, to float64, openStart, openEnd, ok bool)` *(new)* in `order.go` |
| `core/database/persondetail` | `timeline.go`: `Span`, `Timeline` *(new)*; `PersonEvent.Span`; `PersonEvents` returns `Timeline`. `lifeline.go` *(new)*: `lifeline`, `LifeMarker`, `lifeMarkers` |

### FFI

`api/proto/engine.proto`: D-8. `api/ffi/handlers/person_detail.go`: `ListPersonEvents` maps `Timeline.Events` as today plus `span`, then `birth` and `death` (`lifeMarkerProto` *(new)*). Regenerate (`api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`).

### macOS

None — no view. Types and mapping only (D-9).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — One dating rule and the span**
Files: `core/database/conclusionsummaries/compose.go`, `order.go`, `conclusionsummaries_test.go`. Skill: `add-date-value`.
Do: D-1, D-2, D-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/...` passes (T-1, T-7, T-8; `TestHasEventDate` gains its cases; `TestListEvents`, `TestEventsByIDs`, `TestSortByBirth`, `TestPersonEvents`, `TestPersonFamily*` unchanged: none has an open-start date at the planning base).

**STEP-2 — The table and the funeral key**
Files: `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/vocabkey_test.go`, `core/database/persondetail/lifeline.go` *(new)*.
Do: D-7, and D-10's role keys (`vocabkey.RoleChild`, `RoleWitness`, `RoleInformant`, `RolePallbearer` where missing) with `participantRoleOrder`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/...` passes (T-2, with the role rows).

**STEP-3 — Spans and markers on the read**
Files: `core/database/persondetail/timeline.go`, `lifeline.go`, `timeline_test.go`, `lifeline_test.go` *(new, package `persondetail`)*. Skill: `add-catalog-query`.
Do: D-4, D-5, D-6, D-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/...` passes (T-3, T-4, T-9; `TestPersonEvents` and `TestPersonEventsQueryCount` unchanged but for the return value at their call sites).

**STEP-4 — Wire**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/person_detail.go`, `person_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-8.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes (T-5).

**STEP-5 — Swift types and mapping**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`, `OnboardingModelTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: D-9.
Check: the app builds; T-6 passes; `scripts/check-pr.sh` prints PASS.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md` (say the Events list now orders a "bef." date at its known end), `deployment-plan.md` (map row), this plan (status, Landed note). No domain doc names `HasEventDate` at the planning base.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-33-timeline-positions.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestEventSpan` *(new)*, table-driven | 12 Mar 1817 → [From, To) of that day; "1835" → [1835, 1836); range 1844 – 1849 → [1844, 1850); AFT 1870 → [1870, 1871) open end; ABT 1846 → [1846, 1847); BEF 1880 → [1880, 1881) open start; a range with no start year ending 1849 → [1849, 1850) open start; end date only, no year, no date → not ok; start date only → its span. Every case: `ok == HasEventDate` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, D-2, D-3 |
| T-7 | Go | same file `TestEventOrderAtKnownEnd` *(new)*; `TestHasEventDate` (cases added: "bef only" → true, "range with no start" → true, "both ends open" → false) | `EventsByIDs` and `ListEvents` over events dated 1879, BEF 1880, 1881 and none → 1879, BEF 1880, 1881, none; BEF 1880 and "1880" with refs a and b → by ref | AC-1.10, D-2, D-3 |
| T-8 | Go | same file `TestSortByBirth` (case added) | Births 1815, BEF 1817, 1820, none → that order; the lead birth of a person with births BEF 1820 and 1825 is the BEF one (`PersonsByIDs`) | D-2, D-6 |
| T-2 | Go | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` (rows added) | `vocabkey.EventFuneral` and D-10's role keys are seeded terms; `vocabkey.EventFuneral` is a seeded `event_type` term | D-7 |
| T-3 | Go | `core/database/persondetail/timeline_test.go` `TestPersonEventsLifeline` *(new)*, table-driven on `kinWorld` | AC-2.1 – AC-2.8's worlds → markers as listed; AC-2.1 and AC-2.6 also assert the marker equals `PersonSummary.Birth.Event` / `.Death.Event` read through `conclusionsummaries.PersonsByIDs`; every row's `Span != nil` equals `Dated` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-1.7, D-5, D-6 |
| T-4 | Go | `core/database/persondetail/lifeline_test.go` *(new)* `TestLifeMarkers` (package `persondetail`) | Rows built by hand: only a burial dated → death `Beside` nil (no row qualifies); a funeral before a census → death beside the census; no dated row → both nil | D-5, D-7 |
| T-5 | Go | `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents` (new case "carries spans and markers") | A person with a birth and an undated event → the birth row's `span` set, the undated row's absent; `birth.event_id` the birth; `death.event_id` empty, `death.beside_event_id` the birth | D-8 |
| T-9 | Go | `core/database/persondetail/lifeline_test.go` `TestParticipantOrder` *(new)*, table-driven on hand-built rows | AC-4.1 – AC-4.3's cases; equal rank and name → by ref; the person's roles and the event order unchanged | AC-4.1, AC-4.2, AC-4.3, D-10 |
| T-6 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (the `.personEvents` case) | A `FakeStore` timeline with a span and markers → the handle's value equals it; an empty `CatalogPersonTimeline()` → placeholders with no `beside` | AC-3.1, AC-3.2, D-9 |

S13-05's and S13-12's tests whose expectations move: none at the planning base (no `TestPersonEvents`, `TestHasEventDate`, `TestListEvents`, `TestEventsByIDs`, `TestSortByBirth` or `TestPersonFamily*` fixture has an open-start date; checked with `rg -n "QualifierBEF" core/database/conclusionsummaries core/database/persondetail api/ffi/handlers`, whose one hit is a place-chain test). Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `TestPersonEvents`, `TestPersonEventsQueryCount`, `TestListPersonEvents` (STEP-3, STEP-4), the Swift tests that build `[CatalogPersonEvent]` (STEP-5).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `[CatalogPersonEvent]` as a store or key value type `rg`'d across `macos/`, `.agents/`, `.cursor/rules/` and `docs/*.md` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No view reads the fields. In a debug build, open a Person with a recorded birth and a funeral but no death, choose Timeline (still the empty place), and confirm with the `.personEvents` handle in the debugger, or `TestListPersonEvents`' output, that `birth.eventId` is the birth and `death.besideEventId` is the last dated non-funeral event.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The order change reaches more than open-start dates | T-1's equality and T-7; `TestListEvents`, `TestEventsByIDs`, `TestSortByBirth` unchanged; a date with a closed first end keeps `SortKey`'s key byte for byte |
| The key's value type change breaks a Swift caller the plan misses | STEP-5's `rg` for `listPersonEvents` and `CatalogPersonEvent` (§ 4 lists every hit at the planning base) |

**Stop and report if:**

- `PersonEvents`, `PersonEvent`, `ListPersonEventsResponse` or `CatalogPersonEvent` differ from § 4.
- `EventSpan`'s `ok` cannot equal `HasEventDate` on some date (T-1): the two rules disagree.
- A test outside `conclusionsummaries`, `persondetail` and `api/ffi/handlers` changes after STEP-1 (another reader of the order this plan did not find).
- `valuecompare.EndsDate(last, last)` does not give the known end's key for some date.
- `TestPersonEventsQueryCount` needs a query after STEP-3.
- A marker pick in T-3 differs from `PersonSummary.Birth.Event` on a canonical event.

## 12. Open questions

None. Question 1 (a BEF-only date on the line) was answered by the user on 2026-11-02: it goes on the line with an open start, under one meaning of "dated" (D-2, D-3).

## 13. PR

**Title:** S13-33 — Timeline positions
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-33-timeline-positions.md`.
