# S13-33 — Timeline positions

**Status:** Planned
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
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-23, R-24, R-24a); [`structured-date-model.md`](../../../structured-date-model.md); working rules: "The timeline is a lifeline", "Timeline is direct participation, any role", "Each section reads its own query", "Engine PRs run ahead of their slice" (data, not layout), "Canonical only, filtered in Go"; the S13-05 plan [`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md) and its **Landed** note
**Planned:** 2026-11-02 against `main` @ `51dccd4`
**Review:** [`S13-33-timeline-positions.review.md`](S13-33-timeline-positions.review.md) — round 1 Fits with changes / Ready with changes; every finding Fixed but X-4 (no change needed); none open

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

The engine says where each dated event of a person's life sits on a line of years, and where the life's birth and death markers go, so the Timeline (S13-13) can draw a lifeline to scale and Swift only scales numbers (working rule "The timeline is a lifeline", user 2026-11-02). Each dated row of `ListPersonEvents` gains its span on the year line; the response gains a birth and a death marker, each the recorded event or a placeholder that names the dated event it sits beside and invents no date; each event's participants come ordered by role, then by name.

**Why now.** S13-13's lifeline needs positions, and the working rule puts them in Go; this lands before S13-13.

**What changes beyond the timeline** (user, 2026-11-02: one meaning of "dated" everywhere). `conclusionsummaries` gets one place for an event date: `datePlace`, read from one `valuecompare.DateEnds` call, gives both the order key and the span. An event whose date has an open start but a known end ("bef. 1880", a range with no start year) becomes dated and is ordered at its known end. So it moves out of the undated tail of the Events list (and `EventsByIDs`, and the timeline) to its place among 1880's events; a birth dated that way joins birth order (Family's children and siblings, `SortByBirth`) and can become the header's lead birth or death. A closed-start date keeps its place exactly. `autoreconcile.SortKey` and everything the reconciler does are unchanged; `conclusionsummaries` stops calling it.

**What stays.** `ListPersonEvents`' roles (S13-05); matching's lifespan (`subjectprofile`, `subject.Lifespans`), which is not used; `autoreconcile.SortKey`; every stored row.

## 2. User stories

**US-1.** As the Timeline section, I get each dated event's place on a line of years, from the same rule that orders events.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An event dated 12 Mar 1817 | `span` from that day's `DateEnd.From` to its `To` (1817 + 70/365 to 1817 + 71/365), no open end |
| AC-1.2 | An event dated "1835" | `span` [1835, 1836) |
| AC-1.3 | An event with the range 1844 – 1849 | `span` [1844, 1850) |
| AC-1.4 | An event dated "aft. 1870" | `span` [1870, 1871), `open_end` true |
| AC-1.5 | "abt. 1846" | The same span as "1846" (ABT is not widened) |
| AC-1.6 | An event whose only date is "bef. 1880" | `span` [1880, 1881) with `open_start` true; in the read's order among 1880's events, by ref (D-2, D-3) |
| AC-1.8 | A range with no start year, ending 1849 | `span` [1849, 1850) with `open_start` true (D-3) |
| AC-1.9 | An event with only an end date, or a date with no year | No `span`: undated (S13-05 AC-1.4 unchanged) |
| AC-1.10 | The Events list with events dated 1879, "bef. 1880", 1881 and an undated one | Order 1879, bef. 1880, 1881, then the undated one (before this PR: 1879, 1881, bef. 1880, undated) (D-2) |
| AC-1.7 | Any row | "Dated" is "has a span": one field, no second flag (D-4) |

**US-2.** As the Timeline section, I always get a birth and a death marker, placed from the records, with no invented date.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Pat: birth B (1820, subject), census C (1851), death D (1888), burial U (1890) | `birth.event` = B, `death.event` = D, the events the header shows (`PersonSummary.Birth.Event`, `.Death.Event`); both `beside` empty |
| AC-2.2 | Pat with no birth event; census C (1835), then a marriage M (1840) | `birth.event` empty, `birth.beside` = C |
| AC-2.3 | Pat with census C (1875) and funeral F (1880), no death | `death.event` empty, `death.beside` = C (F may follow death; working rule "The timeline is a lifeline") |
| AC-2.4 | Pat with only undated events | Both markers empty, both `beside` empty |
| AC-2.5 | Pat with an undated birth B (no date at all) and a census C (1835) | `birth.event` = B, `birth.beside` = C (B has no span of its own) |
| AC-2.6 | Two birth events with Pat as subject, 1820 and 1822 | `birth.event` = the 1820 one: the header's lead birth (D-6) |
| AC-2.7 | Pat is a witness at someone's birth | That birth is not Pat's birth marker (the header's pick is subject role only) |
| AC-2.8 | Death D dated 1888 before a census C dated 1890 | `death.event` = D; C keeps its own span; nothing is flagged |
| AC-2.9 | The header's lead birth is an event that is not canonical | `birth.event` empty (a placeholder): the timeline shows canonical events only (D-6) |

**US-3.** As the app, I read the new fields.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A response with spans and markers | `CatalogPersonTimeline` carries them; `CatalogPersonEvent.span` set on dated rows; `dated` reads `span != nil` |
| AC-3.2 | An empty `CatalogPersonTimeline()` | Both markers default to placeholders with no `beside` |

**US-4.** As a researcher reading an event, I meet the people who mattered most to it first.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A baptism with participants Witness Kohl, Mother Anna, Father Heinrich (that link order) | Heinrich (father), Anna (mother), Kohl (witness) (D-10) |
| AC-4.2 | A burial with pallbearers Zimmer and Abel, a participant with the researcher's role "Officiant", one with "Mourner", and one with no role | Abel, Zimmer (pallbearer, by name), then Mourner, then Officiant (unlisted roles by label), then the one with no role (D-10) |
| AC-4.3 | One participant with roles witness and informant | Ranked as a witness (the best of its roles) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One place for an event date: the order key and the span from one `DateEnds` call, an open start ordered at its known end, shared by the event order, birth order and the life lead (D-1, D-2, D-3) | STEP-1, AC-1.1 – AC-1.10 |
| The span replaces `Dated` and `HasEventDate` (D-4) | STEP-1, STEP-3, AC-1.7 |
| The after-death / before-birth table, `vocabkey.EventFuneral` and the role keys (D-7, D-10) | STEP-2, AC-2.3 |
| Spans and markers on `persondetail.PersonEvents`, markers read from the person's summary (D-5, D-6) | STEP-3, AC-2.1 – AC-2.9 |
| Each event's participants ordered by role, then by name (D-10; board S13-D4 turn 12, user 2026-11-02) | STEP-3, AC-4.1 – AC-4.3 |
| Wire and handler (D-8) | STEP-4 |
| Swift types, mapping, store, key value type (D-9) | STEP-5, AC-3.1, AC-3.2 |
| The display date rule in `docs/structured-date-model.md` (M-9) | STEP-6 |

| Out | Home |
| --- | --- |
| The lifeline view | S13-13 |
| An estimated year for an unknown birth or death; flags on events outside the life | not planned (user, 2026-11-02) |
| Reading a Julian-calendar date as Julian | not planned; `DateEnds` reads every date as Gregorian (Edges) |
| Matching's lifespan window or its knobs | not planned (working rule "The timeline is a lifeline") |
| Any change to `autoreconcile.SortKey`, `valuecompare.DateWindow` or the reconciler | not planned (user, 2026-11-02) |
| `lifePicks` reading non-canonical birth and death events (the header can name an event the timeline drops) | logged in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) by STEP-6 (working rule "Canonical only, filtered in Go") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline read | `core/database/persondetail/timeline.go` `PersonEvents(c, entityID) ([]PersonEvent, conclusionsummaries.Summaries, error)`, `PersonEvent{Event, Roles, Participants, Dated}`, `rolesByNeighbor`, `dropUnreadEvents`; one `conclusionsummaries.PersonsByIDs(c, people)` batch for the participants (line ~155) | Rows in `EventsByIDs` order; `Dated` from `conclusionsummaries.HasEventDate` after the drop (line ~173); participants in participation-ref order |
| Date keys | `core/database/conclusionsummaries/compose.go`: `dateKey{key string; ok bool; ref string}`, `dateKeyLess`, `sortEvents`, `eventKeyOf` (`dateSort(h.Date)`, else `dateSort(h.StartDate)`), `birthKeyOf`, `dateSort` (calls `autoreconcile.SortKey`), `keptDateAt` (returns the kept date and calls `SortKey` itself), `eventDate(v, ev)` (the kept `Date` whenever it is non-nil, else the kept `StartDate`), `lifeFrom`, `lifePicks`, `leadLife`; `order.go` `HasEventDate`, `SortByBirth`, `sortByDateKey`, `SortByTitle`, `titleCollator` | `SortKey`'s date key is `DateWindow(d).Lo`, nil when the first end is open, so a BEF date is undated. `eventDate` and `eventKeyOf` choose between `Date` and `StartDate` differently (a `Date` with no key). `dateKey.key` is compared only by `dateKeyLess`, inside the package; nothing compares it with the reconciler cache's `sort_key`. `keptDateAt`'s key is read only by `eventDate` → `lifeFrom`. `compose.go` imports `autoreconcile` only for `SortKey`. `HasEventDate` has one caller, `core/database/persondetail/timeline.go`. `lifePicks` compares event types with the literals `"birth"` / `"death"` |
| Date ends | `core/valuecompare/datewindow.go` `DateEnd{Year, Month, Day, Lo, Hi, From, To, Mid, Open}`, `DateEnds(d)`, `DateWindow`, `EndsWindow` | `Lo`/`Hi` proleptic Gregorian day numbers; `From`/`To` on a line of years (a whole year is `[y, y+1)`); a range whose start is after its end is swapped; "civil components are read as Gregorian"; ABT not widened; AFT an open last end; BEF an open first end |
| Lead birth / death | `compose.go` `lifePicks` (subject-role events whose `event_type` key is `birth` or `death`, not filtered to canonical), `leadLife` (least `dateKey`) | `PersonSummary.Birth.Event` / `.Death.Event`, which the header shows |
| Vocab keys | `core/vocabkey/vocabkey.go` `EventBirth`, `EventDeath`, `EventBurial`, `RoleSubject`, `RoleFather`, `RoleMother`; no `EventFuneral`, `RoleChild`, `RoleWitness`, `RoleInformant`, `RolePallbearer`. `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed`; `registry.go` seeds `event_type` `funeral` and the `role` terms subject, father, mother, child, witness, informant, pallbearer in that order | — |
| Lifespan (matching) | `core/subjectprofile/subjectprofile.go` `lifespan`; `core/matchregistry/subject.go` `Lifespans` | Pads an unknown end by 110 years and 10 more; not for display |
| Wire | `api/proto/engine.proto` `ListPersonEventsResponse{events = 1}`, `PersonEvent{event = 1, roles = 2, participants = 3, dated = 4}` | — |
| Handler | `api/ffi/handlers/person_detail.go` `ListPersonEvents` (the one non-test caller of `persondetail.PersonEvents`), `personEventProto` | Maps rows in order |
| Swift | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonEvent{event, roles, participants, dated}`, protocol `listPersonEvents(projectDir:entityID:) -> [CatalogPersonEvent]`; `GoStore.swift` `listPersonEvents` (maps `dated`); `FakeStore.swift` `seededPersonEvents`, `listPersonEvents`; `CatalogQueryRegistry.swift` `.personEvents` load | No view reads `dated` yet |
| Tests | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestHasEventDate`, `TestSortByBirth`; `core/database/persondetail/timeline_test.go` `TestPersonEvents` (`kinWorld`, `w.event(name, typ, year)`: whole years only, `w.byRef`, `w.k.Join`), `TestPersonEventsQueryCount`; `core/datevalue/datevaluetest` (`YMD`, `About`, `Before`, `After`, `Between`); `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`, `OnboardingModelTests.swift` | — |

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) added fields to an existing read, changed its Go return value, and read the person's own summary through the same `ByIDs` batch (`want.Add(vocabkey.Person, entityID)` in `core/database/persondetail/family.go`); copy that shape. S13-05 ([#680](https://github.com/mendahu/provenencia/pull/680)) is the read and its tests to extend.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Positions are on the **year line**: `From` and `To` as doubles from `valuecompare.DateEnd.From` / `.To`, not day numbers | Swift draws year ticks at whole numbers with no date arithmetic (working rule: Swift only scales) | Day numbers: a second calendar rule in Swift |
| D-2 | One place for an event date (round 1 M-2). `compose.go` gains `type TimelineSpan struct { From, To float64; OpenStart, OpenEnd bool }` *(new, exported)* and `datePlace(v *datevalue.Value) (key int, span TimelineSpan, ok bool)` *(new, unexported)*, one `valuecompare.DateEnds(*v)` call: not ok when `v` is nil, `DateEnds` is not ok, or both ends are open. **Key:** the first end's `Lo` when the first end is closed, else the last end's `Lo`. **Span** (round 1 X-3): first closed and last closed → `From` = first `From`, `To` = last `To`; last open (AFT, a range with no end year) → `From` = first `From`, `To` = first `To`, `OpenEnd`; first open (BEF, a range with no start year) → `From` = last `From`, `To` = last `To`, `OpenStart`. `dateKey.key` becomes an `int` (the day number); `dateSort` returns `datePlace`'s key; `keptDateAt` stops computing a key (it returns the date and the count); `eventDating(date, start *datevalue.Value) (*datevalue.Value, bool)` *(new, unexported)* is the one field choice: `date` when `datePlace(date)` is ok, else `start` when it is ok, else `date` if non-nil, else `start`. `eventKeyOf` (an `EventSummary`) and `eventDate` (a node: `lifeFrom` → `leadLife`, `lifeOwners`) both choose through it, and `lifeFrom` keys through `datePlace`. `EventSpan(h EventSummary) (TimelineSpan, bool)` *(new, `order.go`)* returns the chosen date's span. `compose.go` no longer imports `autoreconcile` | One value gives "dated", "where it sits" and "how long", so they cannot disagree; the keys were only ever compared inside the package (§ 4); a closed-start date keeps `DateWindow(d).Lo`, the same order as before; the event-date field choice had two copies | Keeping `SortKey` with a known-end fallback through `EndsDate` (two `DateEnds` calls and a dependency on the reconciler's key format); `EventSpan` beside `HasEventDate` held equal by a test |
| D-3 | A date with an open start and a known end is ordered at its known end's first day: "bef. 1880" at 1 Jan 1880, "bef. 12 Mar 1880" at 12 Mar 1880, among that day's events by ref. A date with both ends open, only an end date, or no year stays undated | The user's answer (2026-11-02): on the line with an open start, ordered at its known end | An estimated earlier date |
| D-4 | `PersonEvent.Dated` is replaced by `Span *conclusionsummaries.TimelineSpan` *(new field)*, set from `EventSpan` where `Dated` is set today; nil when undated. `HasEventDate` is deleted; its cases move into `TestEventSpan`. Proto field 4 (`dated`) is `reserved` (round 1 M-3) | One field for one fact; `HasEventDate` had one caller; the engine ships in the app bundle (`docs/application-stack.md`), so no older engine reads field 4 | `Span` beside `Dated` held equal by a test |
| D-5 | `PersonEvents` returns `(Timeline, conclusionsummaries.Summaries, error)`, `type Timeline struct { Events []PersonEvent; Birth, Death LifeMarker }` *(new)*, `type LifeMarker struct { Event []byte; Beside []byte }` *(new)*. `Beside` is set whenever the marker has no span of its own (a placeholder, or a recorded event with no span): for birth, the first row in order with a span whose type is not in `lifeline.beforeBirth`; for death, the last such row whose type is not in `lifeline.afterDeath`; nil when no row qualifies. Computed by `lifeMarkers(events []PersonEvent, types func(event []byte) vocabkey.EventType, birth, death []byte) (Birth, Death LifeMarker)` *(new, unexported, `lifeline.go`)*, which computes only `Beside` | The user's rule (2026-11-02): every timeline has a spot for birth and death, placed by which events may come before birth or after death; no date invented | Placing placeholders in Swift; matching's window (pads 110 years) |
| D-6 | The recorded markers are the header's (round 1 M-1): `PersonEvents` adds `entityID` to its existing `PersonsByIDs` batch, and `Birth.Event` is the person's `PersonSummary.Birth.Event` id when that event is one of the timeline's rows, else nil (a placeholder); the same for `Death`. No pick rule of its own | One owner of "the person's birth": the header and the timeline agree by construction; the batch is already made, so no query is added | A first-row-of-type-birth pick (a second pick rule kept equal by a test) |
| D-7 | The table is one declarative value in `core/database/persondetail/lifeline.go` *(new)*: `var lifeline = struct { afterDeath, beforeBirth []vocabkey.EventType }{afterDeath: []vocabkey.EventType{vocabkey.EventBurial, vocabkey.EventFuneral}}`, documented as placing placeholders and judging no event. `vocabkey.EventFuneral EventType = "funeral"` *(new)* joins the seeded-key block and `TestVocabKeysMatchSeed` gains its row. `lifePicks` compares with `vocabkey.EventBirth` / `EventDeath` instead of string literals (round 1 M-8) | Page-only, beside the read; typed keys held to the seed | `matchregistry`; a seeded flag on `event_type` terms; string literals |
| D-8 | Wire. `PersonEvent` gains `TimelineSpan span = 5;` *(new)*, `message TimelineSpan { double from = 1; double to = 2; bool open_start = 3; bool open_end = 4; }` *(new)*, absent on undated rows; `reserved 4; reserved "dated";`. `ListPersonEventsResponse` gains `LifeMarker birth = 2; LifeMarker death = 3;`, `message LifeMarker { string event_id = 1; string beside_event_id = 2; }` *(new)*, ids as `database.UUIDString`, empty for none. The handler always sends both markers (round 1 X-5). Comments state D-2's and D-5's rules | The S13-12 shape; one name, `TimelineSpan`, in Go, proto and Swift (round 1 M-7) | A new RPC |
| D-9 | Swift. `struct CatalogTimelineSpan: Sendable, Equatable { var from: Double; var to: Double; var openStart: Bool; var openEnd: Bool }` *(new)*; `CatalogPersonEvent` drops its stored `dated` for `var span: CatalogTimelineSpan? = nil` and a computed `var dated: Bool { span != nil }`; `struct CatalogLifeMarker: Sendable, Equatable { var eventId: String? = nil; var besideEventId: String? = nil }` *(new)*; `struct CatalogPersonTimeline: Sendable, Equatable { var events: [CatalogPersonEvent] = []; var birth = CatalogLifeMarker(); var death = CatalogLifeMarker() }` *(new)*. The store method becomes `listPersonEvents(projectDir:entityID:) async throws -> CatalogPersonTimeline` and the `.personEvents` key's value type is `CatalogPersonTimeline`; `GoStore` maps `span` when `hasSpan`, an empty id to nil. `FakeStore.seededPersonEvents` becomes `[String: CatalogPersonTimeline]`; the `OnboardingModelTests` stub and the tests that build `[CatalogPersonEvent]` follow | The markers belong to the timeline, not to a row; S13-13 reads `dated` by the same name | Markers as flags on rows; a stored `dated` beside `span` |
| D-10 | Participant order (user, 2026-11-02, from board S13-D4 turn 12), in `core/database/persondetail/participants.go` *(new)* beside the read (round 1 M-6). `var participantRoleOrder = []vocabkey.Role{vocabkey.RoleSubject, vocabkey.RoleFather, vocabkey.RoleMother, vocabkey.RoleChild, vocabkey.RoleWitness, vocabkey.RoleInformant, vocabkey.RolePallbearer}`, with `RoleChild`, `RoleWitness`, `RoleInformant`, `RolePallbearer` *(new)* in `vocabkey` beside `RoleSubject`. A participant's rank is the best rank of its roles: a listed role ranks by its index; an unlisted role (a researcher's own) ranks after every listed one, the unlisted role terms of the event ordered by `conclusionsummaries.SortByTitle` over their labels, keys as refs (round 1 M-5); a participant with no role ranks last (round 1 X-1). Within a rank, participants keep `PersonsByIDs`' order (title order, ties by ref, so no further tie-break is needed). `sortParticipants(ps []EventParticipant, index map[string]int)` *(new)* sorts stably by (rank, index), after `dropUnreadEvents`. `TestVocabKeysMatchSeed` gains the four role rows, and a reverse check that every seeded `role` term is in `participantRoleOrder` (round 1 M-4) | The board orders by role; order is Go's; the title order and its collator have one owner, reached through `PersonsByIDs` and `SortByTitle` | Sorting in Swift; a rank stored on seeded terms; a byte compare of labels; the board's "widow" and "officiant", which are not seeded roles |

**Rules this plan bends.** None. S13-05 D-8 made "undated" what the order puts last; D-2 widens "dated" to a known end with the user's answer (2026-11-02), keeping one rule. The working rule "Timeline is direct participation, any role" named `HasEventDate`; STEP-6 updates it (the planning commit already did, 2026-11-02).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `datePlace`, `eventDating` *(new, unexported)*; `TimelineSpan`, `EventSpan` *(new)* | type, func | `dateSort`, `eventKeyOf`, `eventDate`, `keptDateAt`, `HasEventDate`, `autoreconcile.SortKey` | Replaces the key path and the two field choices with one place |
| `persondetail.Timeline`, `LifeMarker`, `lifeMarkers`, `lifeline` *(new)* | type, func, var | `PersonEvent`, `Family`, `matchregistry.Lifespans` | The timeline's result; a page rule, not a matching knob |
| `participantRoleOrder`, `sortParticipants` *(new)* | var, func | `rolesByNeighbor`, `SortByTitle` | The order; reuses `SortByTitle` for labels and `PersonsByIDs` for names |
| `vocabkey.EventFuneral`, `RoleChild`, `RoleWitness`, `RoleInformant`, `RolePallbearer` *(new)* | const | `EventBurial`, `RoleSubject` | Seeded keys code reads by name |
| proto `TimelineSpan`, `LifeMarker` *(new)* | proto | `OriginSibling` | New data |
| `CatalogTimelineSpan`, `CatalogLifeMarker`, `CatalogPersonTimeline` *(new)* | Swift types | `CatalogPersonFamily` | The read's value as one struct |

**Builds on.** `valuecompare.DateEnds`, `persondetail.PersonEvents`, `PersonsByIDs`, `SortByTitle`, `TestVocabKeysMatchSeed`, the `.personEvents` key (S13-05).
**Deletes.** `HasEventDate` and `TestHasEventDate` (cases moved into `TestEventSpan`); `PersonEvent.Dated`; proto `PersonEvent.dated` (reserved); Swift's stored `CatalogPersonEvent.dated` (now computed); `dateSort`'s and `keptDateAt`'s `autoreconcile.SortKey` calls and `compose.go`'s `autoreconcile` import; `eventDate`'s own field choice (through `eventDating`); `lifePicks`' string literals; `[CatalogPersonEvent]` as the key's value type.
**Next adds.** S13-13 draws the line from `CatalogPersonTimeline`. Another event type allowed after death (probate) is one table entry and one `vocabkey` constant; a newly seeded role fails `TestVocabKeysMatchSeed` until it has a rank.

### Reach

None — no new path into existing code; the read already serves this key, and no view reads the new fields until S13-13.

### Scale

At the brief's sizes (a 120-event life, events with 41 participants): `datePlace` is one `DateEnds` per date (fewer calls than today's `SortKey` path); `lifeMarkers` and `sortParticipants` are one pass and one sort per event; the person's own id joins an existing `PersonsByIDs` batch, so `TestPersonEventsQueryCount` (warm: zero; cold: constant batches) stays green unchanged.

### Edges

- **Assumptions.** Rows arrive in `sortEvents` order, so the first and last rows with a span are the earliest and latest by key; spans may overlap (a long range, an open start), which only the view meets. "Dated" and the span are one value (D-2, D-4), so no equality needs a test. A closed-start date keeps `DateWindow(d).Lo` as its key, the order it had (T-7's unchanged cases).
- **Order moves** (user, 2026-11-02). Every reader of the key moves for an open-start date: `EventsByIDs` and `ListEvents` (the Events list, and Promote's and search's event rows that come through them), the timeline, `SortByBirth` (Family's children and siblings), `leadLife` (the header's birth and death when a person has several), `lifeOwners` (the order of a life row's provenance owners). Before: such a row came last with the undated. After: at its known end. No stored value changes.
- **Neighbors.** ABT → the same span as the bare date; AFT → `OpenEnd`; a range with no end year → `OpenEnd`; BEF or a range with no start year → `OpenStart`, ordered at its known end; both ends open, only an end date, or a day that does not exist → undated. A `Date` that has no place with a `StartDate` that has one → the `StartDate` everywhere now (before, the header read the `Date`).
- **Removed inputs.** A birth event unpromoted or merged: the next read takes the header's next pick or a placeholder (the key refreshes on any touched handle). A header birth that is not canonical: a placeholder on the timeline (AC-2.9; the header's own canonical gap is logged, § 3 Out).
- **Real data.** The table names two seeded types and the role order every seeded role; `TestVocabKeysMatchSeed` holds both ways.
- **Untrusted size.** None new: rows are bounded as S13-05 bounds them. A date stored with `calendar: julian` is placed as if Gregorian (up to 13 days off), the same as the order: accepted at the line's scale.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `EventFuneral`, `RoleChild`, `RoleWitness`, `RoleInformant`, `RolePallbearer` *(new)* |
| `core/database/conclusionsummaries` | `compose.go`: `TimelineSpan`, `datePlace`, `eventDating` *(new)*; `dateKey.key` an `int`; `dateSort`, `keptDateAt`, `eventDate`, `eventKeyOf`, `lifeFrom` per D-2; `lifePicks` typed keys (D-7). `order.go`: `EventSpan` *(new)*; `HasEventDate` deleted |
| `core/database/persondetail` | `timeline.go`: `Timeline` *(new)*; `PersonEvent.Span` replaces `Dated`; `PersonEvents` returns `Timeline`, adds `entityID` to its `PersonsByIDs` batch. `lifeline.go` *(new)*: `lifeline`, `LifeMarker`, `lifeMarkers`. `participants.go` *(new)*: `participantRoleOrder`, `sortParticipants` |

### FFI

`api/proto/engine.proto`: D-8. `api/ffi/handlers/person_detail.go`: `ListPersonEvents` maps `Timeline.Events` plus `span`, then always `birth` and `death` (`lifeMarkerProto` *(new)*). Regenerate (`api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`). STEP-1 – STEP-3 do not need the regenerated files.

### macOS

None — no view. Types and mapping only (D-9).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — One place for an event date**
Files: `core/database/conclusionsummaries/compose.go`, `order.go`, `conclusionsummaries_test.go`. Skill: `add-date-value`.
Do: D-1, D-2, D-3; delete `HasEventDate` and move `TestHasEventDate`'s cases into T-1; update its one caller in `core/database/persondetail/timeline.go` to `EventSpan` (the field change itself is STEP-3).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/...` passes (T-1, T-7, T-8; `TestListEvents`, `TestEventsByIDs`, `TestSortByBirth`, `TestPersonEvents`, `TestPersonFamily*` unchanged: none has an open-start date at the planning base).

**STEP-2 — Keys and the table**
Files: `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/vocabkey_test.go`, `core/database/persondetail/lifeline.go` *(new)*, `participants.go` *(new)*.
Do: D-7, and D-10's role keys and `participantRoleOrder`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/... ./core/database/conclusionsummaries/...` passes (T-2).

**STEP-3 — Spans, markers and participant order on the read**
Files: `core/database/persondetail/timeline.go`, `lifeline.go`, `participants.go`, `timeline_test.go`, `lifeline_test.go` *(new)*, `participants_test.go` *(new)* (both package `persondetail`). Skill: `add-catalog-query`.
Do: D-4, D-5, D-6, D-10. Add the test helper `func (w *kinWorld) eventOn(name string, typ vocabkey.EventType, d *datevalue.Value) []byte` *(new)* in `timeline_test.go` beside `event` (it calls `w.k.Date(id, vocabkey.PropertyDate, *d)` when `d` is non-nil), with `datevaluetest` values (round 1 X-2).
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
Files: `docs/structured-date-model.md` (a paragraph after § 3.3: the display date rule — an event's date is its kept date, else its start date, whichever has a place; it is dated when at least one end is closed; it sorts by its first day, or its known end's first day when the start is open; on a timeline it spans `From` to `To` on the year line with open ends flagged; `conclusionsummaries` owns it, separate from the reconciler's `sort_key`) (round 1 M-9); `docs/ideas/structural-debt.md` (a row, next free A-number: `lifePicks` reads non-canonical birth and death events); `docs/deployment-plan/spike-13/completed.md` (the Events list now orders a "bef." date at its known end); `deployment-plan.md` (map row; the working rule "Timeline is direct participation, any role" names the span, not `HasEventDate`, if the planning commit did not); this plan (status, Landed note).
Do: per the Definition of done. S13-13's plan names the old key value type in its § 4 on purpose; its STEP-0 refreshes it (round 1 X-6).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-33-timeline-positions.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestEventSpan` *(new)*, table-driven, replacing `TestHasEventDate` | 12 Mar 1817 → that day's [From, To); "1835" → [1835, 1836); range 1844 – 1849 → [1844, 1850); AFT 1870 → [1870, 1871) `OpenEnd`; ABT 1846 → [1846, 1847); BEF 1880 → [1880, 1881) `OpenStart`; a range with no start year ending 1849 → [1849, 1850) `OpenStart`; end date only, no year, no date, both ends open → not ok; start date only → its span; a `Date` with no year and a dated `StartDate` → the start's span | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, D-2, D-3 |
| T-7 | Go | same file `TestEventOrderAtKnownEnd` *(new)* | `EventsByIDs` and `ListEvents` over events dated 1879, BEF 1880, 1881 and none → 1879, BEF 1880, 1881, none; BEF 1880 and "1880": read both refs from the summaries and expect the lower ref first (round 1 X-2); a closed-start date's position is unchanged from the planning base (the existing cases of `TestListEvents` and `TestEventsByIDs`) | AC-1.10, D-2, D-3 |
| T-8 | Go | same file `TestSortByBirth` (case added) | Births 1815, BEF 1817, 1820, none → that order; the lead birth of a person with births BEF 1820 and 1825 is the BEF one (`PersonsByIDs`) | D-2 |
| T-2 | Go | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` (rows added) | `EventFuneral` is a seeded `event_type`; `RoleChild`, `RoleWitness`, `RoleInformant`, `RolePallbearer` are seeded `role` terms; every seeded `role` term is in `persondetail`'s `participantRoleOrder` (exported to the test through an `export_test.go` accessor) | D-7, D-10 |
| T-3 | Go | `core/database/persondetail/timeline_test.go` `TestPersonEventsLifeline` *(new)*, table-driven on `kinWorld` with `eventOn` | AC-2.1 – AC-2.9's worlds (AC-2.5's birth with no date; AC-2.6's two subject births; AC-2.9's lead birth whose claims are made provisional) → markers as listed; the recorded markers equal `PersonSummary.Birth.Event` / `.Death.Event` on canonical events | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, D-5, D-6 |
| T-4 | Go | `core/database/persondetail/lifeline_test.go` *(new)* `TestLifeMarkers` | Rows built by hand: only a burial dated → death `Beside` nil; a funeral after a census → death beside the census; no dated row → both nil; a recorded birth with no span → `Beside` the first qualifying row | D-5, D-7 |
| T-9 | Go | `core/database/persondetail/participants_test.go` *(new)* `TestParticipantOrder` | AC-4.1 – AC-4.3's cases; two unlisted roles "Officiant" and "mourner" → mourner first (case-insensitive collation); equal rank → `PersonsByIDs` order; the person's roles and the event order unchanged | AC-4.1, AC-4.2, AC-4.3, D-10 |
| T-5 | Go | `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents` (new case "carries spans and markers") | A person with a birth and an undated event → the birth row's `span` set, the undated row's absent; `birth.event_id` the birth; `death` sent with an empty `event_id` and `beside_event_id` the birth | D-8 |
| T-6 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (the `.personEvents` case) | A `FakeStore` timeline with a span and markers → the handle's value equals it, `dated` true on the spanned row; `CatalogPersonTimeline()` → placeholders with no `beside` | AC-3.1, AC-3.2, D-9 |

S13-05's and S13-12's tests whose expectations move: `TestHasEventDate` (moved into T-1); no other, because no fixture in `TestPersonEvents`, `TestListEvents`, `TestEventsByIDs`, `TestSortByBirth` or `TestPersonFamily*` has an open-start date (`rg -n "QualifierBEF" core/database/conclusionsummaries core/database/persondetail api/ffi/handlers` hits only a place-chain test). Most likely to break: `TestPersonEvents`, `TestPersonEventsQueryCount`, `TestListPersonEvents` (STEP-3, STEP-4), the Swift tests that build `[CatalogPersonEvent]` (STEP-5).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `HasEventDate`, `PersonEvent.Dated`, and `[CatalogPersonEvent]` as a store or key value type `rg`'d across `macos/`, `.agents/`, `.cursor/rules/` and `docs/*.md` with no hit left outside spike-13 plan history (S13-13's plan is refreshed by its STEP-0); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] docs/structured-date-model.md: the display date rule (STEP-6)
- [ ] docs/ideas/structural-debt.md: the lifePicks canonical row (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No view reads the fields. In the Events list, a project with an event dated only "bef. 1880" shows it among 1880's events, not at the end. In a debug build, open a Person with a recorded birth and a funeral but no death and confirm with the `.personEvents` handle in the debugger (or `TestListPersonEvents`' output) that `birth.eventId` is the birth and `death.besideEventId` is the last dated non-funeral event.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The order change reaches more than open-start dates | A closed-start key is `DateWindow(d).Lo` as before (T-7); `TestListEvents`, `TestEventsByIDs`, `TestSortByBirth` unchanged |
| The key's value type change breaks a Swift caller the plan misses | STEP-5's `rg` for `listPersonEvents` and `CatalogPersonEvent` (§ 4 lists every hit at the planning base) |
| `eventDating` moves the header's date for an event whose `Date` has no year but whose `StartDate` has one | Intended (one choice); T-1's last case; no fixture at the base has it |

**Stop and report if:**

- `PersonEvents`, `PersonEvent`, `ListPersonEventsResponse` or `CatalogPersonEvent` differ from § 4.
- Something outside `conclusionsummaries` compares a `dateKey` or `keptDateAt`'s key with the reconciler's `sort_key` (D-2's premise).
- A test outside `conclusionsummaries`, `persondetail`, `subjectvocab` and `api/ffi/handlers` changes after STEP-1 (another reader of the order this plan did not find).
- `TestPersonEventsQueryCount` needs a query after STEP-3.

## 12. Open questions

None. The BEF-only date was answered by the user on 2026-11-02: on the line with an open start, under one meaning of "dated" (D-2, D-3).

## 13. PR

**Title:** S13-33 — Timeline positions
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-33-timeline-positions.md`.
