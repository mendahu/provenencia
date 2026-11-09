# S16-01 — Cap the Person page reads and an event's places against crafted projects

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 1a — Bounded reads, folds and the Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Person page draws the same rows and lines on real data; the diagram's partial note is S16-04's, brief S16-D1)
**Board:** —
**Depends on:** S14-40 (Spike 14, landed [#746](https://github.com/mendahu/provenencia/pull/746))
**Blocks:** S16-04, S16-12, S16-16, S16-19
**Branch:** `s16-01-cap-person-page-reads`
**Version bump:** patch — a fix: crafted projects no longer stall the Person page, and two wire fields old clients ignore ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the kinship and record links lines it makes false), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (two response fields), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (handler tests), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-2, P-3, P-4, P-10, P-21; working rules "(Batch 1) Crafted fan-out is capped, not engineered", "(Batch 1) The record links read is bounded and says so", "(Batch 2) Batch 1's caps stay where S16-01 puts them"; [`matching-model.md`](../../../matching-model.md) (kinship and record links); map review K-3, K-4, K-9, K-28 ([`../deployment-plan.review.md`](../deployment-plan.review.md))
**Planned:** 2026-11-09 against `main` @ `0095ba7`
**Review:** [`S16-01-cap-person-page-reads.review.md`](S16-01-cap-person-page-reads.review.md) — round 1 answered; round 2 pending

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

A crafted or badly broken project can no longer stall the Person page or project open. Four reads get a fixed cap, checked before the work they bound: a person's relatives per role (100), implied siblings through a person with more than `kinship.Rules.MaxPivotEnd` parents (none), a person's records per Evidence graph (`recordlinks` `MaxEntityMembers`, 10), and an event's places (20). The record links read sends at most 75,000 lines, strong first, and says when it did not count facts or dropped lines; the app's own arc cap goes, since Go now owns it. Every `persondetail` read decides "is this a Person" in one place.

**Why now.** The Person page audit (P-2, P-3, P-4, P-10, P-21) found quadratic reads a crafted catalog reaches, one of them on every project open; the user ruled them capped, not engineered (2026-11-08). S16-04 (the diagram's partial note), S16-12, S16-16 and S16-19 build on the fields and caps this PR places.

**What stays.** Real data never reaches a cap, so every section, summary, search document and diagram reads as today on a real project. No catalog change, no new knob in the match registry, no change to matching. The diagram draws nothing new (S16-04 reads the new fields).

## 2. User stories

**US-1.** As a researcher, when I open a Person in a crafted project, each section draws in under a second, so a broken import cannot freeze the app.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Person with 4,000 children filed by `parent` relationships; `persondetail.PersonFamily` cold (first read after the commit) | Returns in under 1 s; the partner-or-one-parent unit holds exactly 100 children, the first 100 in link order. Planning measured 3.2 s today on the same shape |
| AC-1.2 | `kinship.Siblings` → `KinOf` → `FamilyUnits` in memory, for a person with 50,000 parents and one sibling sharing all of them | Returns in under 1 s; `Siblings` returns no implied link; the person's parents role holds 100 links |
| AC-1.3 | An event with 4,000 places (one `location` link each), the birth of a Person; `conclusionsummaries.EventsByIDs`, `PersonLifeOwners` and `RowsForSearch` over the event and the person | Each returns in under 1 s; the event summary holds at most 20 places, the first 20 in link order. Planning measured 1.2 s and 1.3 s today for the first two |
| AC-1.4 | A Person holding 20,000 accepted records in one Evidence graph; `persondetail.Claims` | Returns in under 1 s with one claim group of 10 claims, the first 10 in member order. Planning measured 3.7 s today |
| AC-1.5 | The real-data tests: `TestPersonFamilyQueryCount`, `TestRecordLinksBudgets` (50 and 65 household graphs), `TestPersonFamily`, `TestClaimsGroupsBySource`, `TestPersonLifeOwners`, `TestInputsMatchIsolatedComposition`, the `conclusionsummaries` list and by-id tests | Unchanged and passing |

**US-2.** As a researcher, the record links read tells the app when it did not count facts or sent fewer lines than it found, so S16-04 can say the diagram is partial instead of drawing every line thin.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `recordlinks.Pairs` on a read whose work is at most `MaxReadWork` | `counted` is true; pairs as today |
| AC-2.2 | `recordlinks.Pairs` past `MaxReadWork` by pair work (lines listed at 0 facts) | `counted` is false; every pair listed, every line 0 facts, as today |
| AC-2.3 | `recordlinks.Pairs` past `MaxReadWork` by the lines alone | `counted` is false; no pair |
| AC-2.4 | `persondetail` record links over four household graphs (18 strong lines) with `MaxReadWork` 18 (past it by pair work), then 1 (past it by lines alone) | `NotCounted` true and `LinesDropped` false both times: first every line weak, then no pair |
| AC-2.5 | The same four graphs with a line cap of 5 | `LinesDropped` true; 5 lines kept, in read order; a pair left with no line is absent |
| AC-2.6 | `recordlinks.CapLines` on 1,500,000 lines (20 times the cap), 10 of them strong, spread through the pairs | Returns in under 1 s; 75,000 lines kept; all 10 strong lines kept; the weak ones kept are the first 74,990 weak in read order; each kept line keeps its pair and its place in the pair |
| AC-2.7 | `GetPersonRecordLinksResponse` built from a read with `NotCounted` and `LinesDropped` set | The wire's `not_counted` and `lines_dropped` are true; from an ordinary read both are false |
| AC-2.8 | `GoStore.getPersonRecordLinks` decodes a response | `CatalogPersonRecordLinks.notCounted` and `linesDropped` carry the wire values |
| AC-2.9 | `PersonRecordLinksContent` given 80,000 lines between two graphs | Draws 80,000 arcs: Swift no longer cuts arcs. 1,005 graphs still draw 1,000 nodes |

**US-3.** As the engine, every `persondetail` read decides whether an id is a Person page's in `personNode`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `persondetail.Claims` with an Event handle | `conclusiondetails.ErrNotFound` (today it answers) |
| AC-3.2 | `persondetail.PersonFamily`, `PersonEvents`, `RecordLinks`, `Claims` with an unknown id, a merged Person, a malformed id and an Event handle | `conclusiondetails.ErrNotFound` from each |
| AC-3.3 | `persondetail.Families` with ids that are not Persons among Person ids | Those ids are skipped, as today; the query count stays fixed (`TestPersonFamilyQueryCount` AC-2.3) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `kinship.Rules.MaxKinPerRole` (100, in `matchregistry`); `kinship.KinOf` keeps at most that many links per role, in link order (P-3) | STEP-1, D-1, AC-1.1, AC-1.2 |
| `kinship.Siblings` implies nothing for a person with more than `Rules.MaxPivotEnd` parents (P-3) | STEP-1, D-2, AC-1.2 |
| `persondetail.Families` reads one capped kin per person for both loops; the person gate goes through `personNode` (P-3, P-21) | STEP-2, D-3, D-4, AC-1.1, AC-3.2, AC-3.3 |
| `persondetail.PersonFamily` and `Claims` gate through `personNode` (P-21) | STEP-2, STEP-3, D-4, AC-3.1, AC-3.2 |
| A person's records per Evidence graph capped at `MaxEntityMembers` in one function `memberGraphs` calls (P-4) | STEP-3, D-5, AC-1.4 |
| An event's places capped at `maxEventPlaces` (20) in `eventLocations`, which `placesOfEvents`, `leadLife`, `lifeOwners` and `composeEvents` read (P-2) | STEP-4, D-6, AC-1.3 |
| `recordlinks.Pairs` reports whether it counted (P-10) | STEP-5, D-7, AC-2.1 – AC-2.3 |
| `persondetail.RecordLinks` sends at most `maxRecordLinkLines` (75,000) lines, strong first, and sets `NotCounted` and `LinesDropped` (P-10) | STEP-6, D-8, D-9, D-10, AC-2.4 – AC-2.6 |
| `GetPersonRecordLinksResponse` gains `not_counted` and `lines_dropped`; the handler maps them (P-10) | STEP-7, D-11, D-12, AC-2.7 |
| `GoStore` decodes them; `PersonRecordLinksContent.maxArcs` and its strong-first pass go (P-10) | STEP-8, D-13, AC-2.8, AC-2.9 |
| Test builders: `graphcachetest.Canon` 24-bit id counter, `Canon.Bulk` and `Canon.Members` | STEP-2, STEP-3, D-15 |
| Lines this PR makes false: `add-match-rule` lines 33 and 34, `matching-model.md` record links paragraph and the `MaxPivotEnd` and `MaxEntityMembers` rows, the `kinship.Siblings` comment, the `PersonFamilyContent.swift` comment, the proto request comments | STEP-1, STEP-8, STEP-9 |
| Version, `completed.md`, map status | STEP-10 |

| Out | Home |
| --- | --- |
| The diagram's partial note, and reading `notCounted`, `linesDropped` and the 1,000-node cut in `PersonRecordLinksContent` | S16-04 |
| `RecordLinksGraph` `source_ref` and its unread `source_id` and `source_type_id` (P-22 part) | S16-12 |
| `edgeKey` joining ids with no separator (P-24) | S16-16 |
| `Families` reading each person's kin once through one pass, and the `canonlinks` loaders only it uses | S16-16 |
| The claims read rescanning a whole Evidence graph per record below the cap; the handlers' per-claim label map | S16-13 |
| Saying a line is partial when one pair passes `MaxPairWork` or one entity passes `MaxEntityMembers` (those lines keep 0 facts silently, as today) | [`docs/ideas/person-page-followups.md`](../../../ideas/person-page-followups.md), a row this PR adds (STEP-10) |
| A cap on how many Evidence graphs one person spans (linear today; 50 is the real Scale) | [`docs/ideas/person-page-followups.md`](../../../ideas/person-page-followups.md), a row this PR adds (STEP-10) |
| A faster `FoldLocations`, Family read or claims read (algorithms) | not planned (rule "Crafted fan-out is capped, not engineered") |
| The record links memo and a household-sized budget (P-9) | [`docs/ideas/person-page-performance.md`](../../../ideas/person-page-performance.md) |

## 4. Current state

Verified at `0095ba7`.

| Area | Where | Today |
| --- | --- | --- |
| Kin by role | `core/kinship/families.go` `KinOf` (line 77), `Kin`, `Roles`, `(Kin).Around` (line 60), `FamilyUnits` (line 151), `parentUnits` (line 170, `pairing := len(parents) <= rules.MaxPivotEnd`), `partnerUnits` (line 289) | `KinOf` puts every link in every role whose key it carries, no bound. `parentUnits` scans each sibling's shapes per parent (`slices.ContainsFunc`, line 235); `partnerUnits` checks children per child (line 315) |
| Implied siblings | `core/kinship/siblings.go` `Siblings` (line 29), comment lines 26–28 ("It has no MaxPivotEnd bound …") | One implied link per parent × other child of that parent; a sibling sharing P parents carries P shapes. Its only production caller is `canonlinks` `(*deriver).siblingsOf` (`core/database/canonlinks/reader.go` line 616), which the Verify check re-derives through the same function |
| Kinship rules | `core/matchregistry/kinship.go` `defaultKinship` (`MaxPivotEnd: 10`, `MaxPairSteps: 10`); `core/kinship/kinship.go` `Rules`; `core/matchregistry/guard_test.go` `layerPackages` (`core/kinship` holds no literal other than 0 and 1); `.cursor/rules/match-rules.mdc` | A new knob is a `Rules` field, a registry value with its `decided:` line, and a `matching-model.md` inventory row |
| Family read | `core/database/persondetail/family.go` `Families` (line 23), `PersonFamily` (line 134) | `Families` loads with `LoadRelationships(want, Siblings)`, skips an id whose cached node is not a Person (`n.Is(typeID)`, line 60), names every relative of the unfiltered kin, loads them all (line 78), then re-reads the person's `Relationships` and filters to canonical relatives (`kinOf`, line 84) for `Around` and `FamilyUnits`. `PersonFamily` gates with `conclusiondetails.Handle` (line 135) and returns `ErrNotFound` when `Families` returns nothing |
| Kinship reader | `core/database/canonlinks/reader.go` `(*Reader).Relationships` (line 412), `LoadRelationships` (line 446), `LoadSiblings`, `(*deriver).siblingsOf` (line 577), `forLinks` (line 639) | Unchanged by this PR. Each relative's `For` walks its parents' links (P-3); with relatives capped, at most 400 per person are loaded |
| Person gate | `core/database/persondetail/claims.go` `personNode` (line 52), "the one place that decides whether an id is a Person page's" | Used by `RecordLinks` (`recordlinks.go` line 28) and `PersonEvents` (`timeline.go` line 91). `Claims` (line 41) uses `conclusiondetails.Handle` only, so an Event handle with accepted members answers |
| Claims | `claims.go` `claims` (line 68), `memberGraphs` (line 118) | `memberGraphs` walks every `n.AcceptedMembers()`, and per member scans the graph's `Members` and `Subjects`; `claims` scans every Observation of the graph per member. M members of one graph cost M × (2S + O) |
| Claims handler | `api/ffi/handlers/person_detail.go` `ListPersonClaims` (line 19), `source_read.go` `rowsOf` | Per claim, `lc.rowsOf` rebuilds a label map of every Subject of the graph |
| Record links read | `core/database/persondetail/recordlinks.go` `PersonRecordLinks`, `RecordLinks` (line 27), `linksBetween` (line 57) | Builds one `recordlinks.Graph` per graph and returns `recordlinks.Pairs` as is. No line cap, nothing says whether it counted |
| Record links count | `core/recordlinks/recordlinks.go` `Pairs` (line 181), `readPlan` (line 212), `Rules.MaxReadWork`, `MaxPairWork`, `MaxEntityMembers`; `core/matchregistry/recordlinks.go` (`MaxReadWork: 1_100_000`) | Past `MaxReadWork` by pair work, every pair is listed with 0-fact lines (`skip`); past it by lines alone, `readPlan` returns nil and `Pairs` returns nil. Callers: `persondetail.linksBetween`, `recordlinks_test.go` (5 calls), `invariants_test.go` (2 calls) |
| Record links export | `core/recordlinks/export_test.go` `Lines`, `PairWork`, `ReadWork` | |
| Record links wire | `api/proto/engine.proto` `GetPersonRecordLinksResponse` (line 1636: `graphs = 1`, `pairs = 2`), `RecordLinksPair`, `RecordLink`; `ListPersonClaimsRequest` (line 1503) and `GetPersonFamilyRequest` (line 1538) comments say "An unknown or merged handle is conclusiondetails.not_found" | |
| Record links handler | `api/ffi/handlers/person_detail.go` `GetPersonRecordLinks` (line 197), `recordLinkStrengthProto` | Maps graphs and pairs inline in the closure |
| Handler tests | `api/ffi/handlers/person_detail_test.go` `TestGetPersonRecordLinks` (line 437), `TestRecordLinkStrengthProto` (line 426) | |
| Event places | `core/database/conclusionsummaries/compose.go` `leadLife` (line 765, `follow(in, lead.event.ID, hops.PlacesOfEvent)` line 775), `eventPlaceGraph` (line 830), `placesOfEvents` (line 859), `lifeOwners` (line 879, `follow` line 889), `composeEvents` (line 936, `follow` line 951), `eventPlaces` (line 1186, drops duplicate ids only), `buildPlaceGraph` (line 1282); `placechains.go` `(*placeGraph).addChains` (line 113, calls `placesOfEvents`), `FoldLocations` (line 360, every ordered pair of places) | Nothing bounds an event's places. Reached by every event summary, every person's birth and death rows, the Timeline, the Events list, Event pages, and `searchindex` reprojection on open (`RowsForSearch`, line 421) |
| Input recording | `core/database/graphcache/inputs.go` `(*InputScope).Follow` (line 85) | Records the hop's bridge end at the event and every returned association's values |
| Swift wire model | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonRecordLinks` (line 816: `graphs`, `pairs`) | |
| Swift decode | `macos/App/Platform/GoStore.swift` `getPersonRecordLinks` (line 1264) | Maps graphs and pairs |
| Swift fake | `macos/App/Platform/FakeStore.swift` `getPersonRecordLinks` (line 1718), `seededPersonRecordLinks` | Returns the seeded value; needs no change |
| Swift diagram content | `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` `maxArcs` (line 16), `maxNodes` (line 18), `init` (lines 28–50: counts `valid`, two strong-first passes past `maxArcs`) | Swift owns the 75,000-arc cap and the strong-first order |
| Swift tests | `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift` `arcsAndNodesStopAtTheCap` (line 109), `householdScale` (line 101) | |
| Family fold comment | `macos/App/Features/Conclusions/PersonFamilyContent.swift` line 23: "Go still returns every member, only the first paint is bounded." | False once `KinOf` caps |
| Person tests | `core/database/persondetail/family_query_test.go` `TestPersonFamilyQueryCount`; `family_test.go` `TestPersonFamily` AC-1.12 (line 201: unknown, event, merged, malformed); `timeline_test.go` `TestPersonEvents` AC-1.10 (line 268); `recordlinks_test.go` `TestRecordLinks` "AC-1.6 an id that is no Person" (line 144); `claims_test.go` `TestClaimsMissingAndEmpty` (line 203: no Event case); `recordlinks_perf_test.go` `TestRecordLinksBudgets`, `households`; `export_test.go` `ParticipantRoleOrder` | |
| Test builders | `core/database/graphcache/graphcachetest/graphcachetest.go` `Canon`, `NewCanon`, `next` (16-bit counter in bytes 14–15: 65,536 ids per Canon), `Entity` (line 136), `observe` (line 159: recomputes the entity's kept rows after every Observation), `Location`, `Parent`, `Join`, `Event`, `Commit` (line 321: recomputes the ids it is given); `persondetail` `kinWorld` (`family_test.go` line 25); `databasetest` `Fastest`, `Since`, `WithinBudget` (`budget.go`) | Planning measured: 300 children through `Canon.Parent` take 26 s to build (each `observe` recomputes, quadratic); with the per-Observation recompute skipped and one `Commit` of every id, 4,000 children build in 5.6 s and 4,000 places in 7.2 s |
| Docs | `.agents/skills/add-match-rule/SKILL.md` line 33 (kinship: "`Rules.MaxPairSteps` … linear for … `Siblings`", "`KinOf` and `FamilyUnits` … for the page"), line 34 (record links: "`Pairs` is the entry …"); `docs/matching-model.md` line 33 (record links paragraph, ends "no line is listed"), line 213 (`kinship.MaxPivotEnd` row), line 217 (`recordlinks.MaxEntityMembers` row) | |

**Planning probes** (throwaway, reverted): a 100-cap in `KinOf` and a 20-place cap in a shared event-places helper took `PersonFamily` cold at 4,000 children from 3.2 s to 0.17 s, and `EventsByIDs` at 4,000 places from 1.2 s to 3 ms. `Claims` at 20,000 records in one graph took 3.7 s.

**Precedent.** `graphcache` `maxWalkEdges` (`core/database/graphcache/evidence_graph.go` line 220, commit `9cfb3854` "Stop one Evidence graph walk from building more than 50,000 edges.") is a crafted-data cap as a package constant with its provenance, checked before the work, with one test just under and one at the cap (`TestWalkEvidenceGraphEdgeCap`). Copy its shape for each cap here. `core/database/persondetail/export_test.go` `ParticipantRoleOrder` is the shape of the test exports. `TestRecordLinkStrengthProto` is the shape of a handler mapper test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `kinship.Rules` (`core/kinship/kinship.go`) gains `MaxKinPerRole int` *(new)*, documented as "the most links `KinOf` keeps in one role; more is crafted data, and the page shows the first ones". `core/matchregistry/kinship.go` sets `MaxKinPerRole: 100` with `// decided: user, 2026-11-08 (Spike 16 rule "Crafted fan-out is capped, not engineered"); a registry knob because kinship is a layer package (user, 2026-11-09): a real family holds a few dozen at most`. `KinOf` becomes `KinOf(person []byte, links []Link, terms segmentcompare.Terms, rules Rules) Kin` and appends a link to a role only while that role holds fewer than `rules.MaxKinPerRole`, so each role keeps its first links in link order. A value below 1 keeps nothing, like every other bound in `Rules`; the registry is the one source of the value. The two `kinship.Rules` literals in `core/kinship/kinship_test.go` (`rules`, line 22, and `capped`, line 372) gain `MaxKinPerRole: 100`. `docs/matching-model.md` gains a knob inventory row (STEP-9). Tests read `matchregistry.Default().Kinship.MaxKinPerRole` | The map rule puts the cap in `KinOf`, so both of `Families`' loops and every later reader of a person's kin get it (map review K-28); S16-16 keeps calling it. `core/kinship` is a layer package: `.cursor/rules/match-rules.mdc` and `core/matchregistry/guard_test.go` (`layerPackages`, numbers true) allow no literal there. `Parents`, `Siblings` and `FamilyUnits` already take `rules` | A package constant in `core/kinship` — fails `TestLayerPackages` and bends `match-rules.mdc` (review M-1); capping in `persondetail` — a second reader of kin would miss it |
| D-2 | `kinship.Siblings` returns nil when `person` has more than `rules.MaxPivotEnd` distinct parents (`len(idx.parentsOf[string(person)])`, already distinct), checked before the loop over parents. The comment at lines 26–28 becomes: a family of eleven children is real data, so the number of other children is not bounded; more parents than `Rules.MaxPivotEnd` is malformed and implies no sibling, the bound `parentUnits` uses to stop pairing; `Rules.MaxPairSteps` keeps the work linear in the edges when a pair is filed many times | The rule names this bound; it keeps a sibling's shared-parent shapes at 10 or fewer, which ends `parentUnits`' P² placement. In `kinship`, the one function `canonlinks` and its Verify check both call | Bounding in `canonlinks.siblingsOf` — a second place that knows the rule; bounding a parent's children — real families pass 10 |
| D-3 | `Families` keeps, per person read, the capped kin of its first loop (`kinship.KinOf` over `Relationships(id, Siblings)`, unfiltered) in a map, names relatives from it, and in its second loop filters that same kin to canonical Persons through `shownKin(g, typeID, id, kin) kinship.Kin` *(new, unexported, `family.go`)* instead of re-reading the person's `Relationships`. A relative's kin is read as today: `Relationships(other, Options{})`, then `KinOf`, then `shownKin`. Past the cap, a merged or non-canonical relative among the first 100 leaves fewer than 100 shown in that role; the next one is not promoted in (crafted data only) | "One capped kin that both loops use, so no relative is read that was not loaded" (audit P-3, map review K-4). Filtering a role's links after `KinOf` equals `KinOf` over filtered links below the cap, because the filter is per link | Re-reading and filtering before `KinOf` as today — past the cap the second loop's first 100 can include a relative the first loop never loaded |
| D-4 | `personNode` is the person gate of every one-id `persondetail` read: `Claims` and `PersonFamily` call it in place of `conclusiondetails.Handle` (`RecordLinks` and `PersonEvents` already do). `Families`, the batch read under `PersonFamily`, keeps its skip as it is: `g.Cached(id)` then `n.Is(typeID)`, which is the predicate `personNode` applies (`Node.Is` is false for a nil or merged node), on a node already loaded, so it issues no query. `personNode` stays in `claims.go` | P-21: one place decides a page id. `personNode` loads through `Graph().Node`, which queries for an id not loaded (an unknown id), so calling it per id in `Families` would add queries (review X-1) | `personNode` per id in `Families` — a query per unknown id; moving `personNode` to its own file — a move the PR does not need (logged, review M-5) |
| D-5 | `claims.go` gains `boundedMembers(n *graphcache.Node) []graphcache.Member` *(new)*: `n.AcceptedMembers()` keeping at most `matchregistry.Default().RecordLinks.MaxEntityMembers` per `GraphID`, in member order; every accepted member counts, the person's first record included. `memberGraphs` iterates `boundedMembers(n)` in place of `n.AcceptedMembers()` | The rule: one function the claims read calls, which S16-13 keeps calling; the registry value is reused, not copied | A new constant 10 in `persondetail` — a copy of a registry value; capping inside `claims` — after `memberGraphs` already scanned |
| D-6 | `compose.go` gains `const maxEventPlaces = 20` with `// decided: user, 2026-11-08 (Spike 16 rule "Crafted fan-out is capped, not engineered"): a real event names one to three places` and `eventLocations(in *graphcache.InputScope, eventID []byte) ([]graphcache.Link, error)` *(new)*: it follows `hops.PlacesOfEvent` through `follow` and keeps every link whose place is among the first 20 distinct places (by `Link.Neighbor`, whether or not that place's node loads later), in link order. `placesOfEvents` stays the place-id list for the place graph and becomes a `neighborsOf` loop over `eventLocations`. `leadLife`, `lifeOwners` and `composeEvents` call `eventLocations` in place of `follow(…, hops.PlacesOfEvent)`. STEP-10 updates the map's two working rules and S16-19's scope, which name `placesOfEvents` as the cap's home, to name `eventLocations`. `eventPlaces` is unchanged: it takes its caller's links | One capped read of an event's places for all four reads (map review K-4), so the place graph, the folded places and the life owners agree. Every link to a kept place stays because `lifeOwners` lists one owner per association | A name one letter from `placesOfEvents` (`eventLocations`) for a function that returns links — review M-2; capping inside `eventPlaces` or `FoldLocations` — `eventPlaceGraph` and `composeEvents` would still load and build every place; capping links instead of places — two links to one place would count twice |
| D-7 | `recordlinks.Pairs` returns `(pairs []Pair, counted bool)`. `counted` is false exactly when the read's work passes `MaxReadWork` (both `skip` and the nil `readPlan`), true otherwise, a read of no graph included. A pair over `MaxPairWork` or an entity over `MaxEntityMembers` does not change it. The doc comment says so | The rule names `MaxReadWork` only; the other caps are per pair or per entity and stay silent (Out) | A `Read` struct — one bool does not need a type; reporting per-pair overs — not in the rule |
| D-8 | `persondetail.PersonRecordLinks` gains `NotCounted bool` and `LinesDropped bool`. `NotCounted` is `!counted`. The wire and Swift use the same names (`not_counted`, `notCounted`), so the zero value everywhere is the ordinary read | proto3 bools default to false; one name end to end | A `Counted` field — its zero value would mean "not counted" on the wire and in `FakeStore` seeds |
| D-9 | `core/database/persondetail/recordlinks.go` gains `const maxRecordLinkLines = 75_000` with `// decided: S13-28 round 1 (SEC-1), the caps-for-adversarial-only rule; moved from PersonRecordLinksContent.maxArcs by S16-01 (user, 2026-11-08: Go owns the bound): 20 times the 3,675 lines of S13-27's 50-graph household Scale`, and `core/recordlinks/recordlinks.go` gains `CapLines(pairs []Pair, most int) ([]Pair, bool)` *(new)*, beside `Pairs`; `most` is at least 1. A line is strong when `Line.Strength == StrengthStrong`, else weak. With at most `most` lines in all, it returns the pairs as they are and false. Otherwise it keeps the first `most` strong lines in read order (pair order, then line order), then, while fewer than `most` are kept, the first weak lines in read order; each kept line stays in its pair at its place; kept pairs keep their order; a pair left with no line is dropped; it returns true. It builds new `Lines` slices and does not modify its input | The rule: strong first, a package constant carrying `maxArcs`' provenance, not a registry knob. `CapLines` is pure on `recordlinks` types and holds no literal, so it sits beside the package that defines strong (review M-4); the number stays in `persondetail`, which is no layer package | `capLines` in `persondetail` — a pure operation tested through the catalog harness; sorting strong lines to the front of each pair — moves lines the app reads in order; a per-pair cap — the bound is the total the diagram draws |
| D-10 | `RecordLinks` becomes a call of `recordLinks(c *database.Catalog, entityID []byte, cfg recordlinks.Config, most int) (PersonRecordLinks, error)` *(new, unexported)* with `matchregistry.Default().RecordLinksConfig()` and `maxRecordLinkLines`. `linksBetween(c *database.Catalog, page []byte, graphs []graphcache.EvidenceGraph, cfg recordlinks.Config) ([]recordlinks.Pair, bool, error)` takes `cfg` from its caller, returns `Pairs`' `counted`, and still sets `cfg.Value` from `valuerules.Load` on its copy, so a caller's `RecordLinks` and `Subject` sections are kept. `recordLinks` calls `recordlinks.CapLines(pairs, most)` after `linksBetween`. `core/database/persondetail/export_test.go` adds `RecordLinksWith = recordLinks` | A test reaches "past `MaxReadWork`" and "past the line cap" on four graphs instead of 1,500 (planning: a `Canon` fixture past `MaxReadWork` needs about 1,500 graphs) | A package variable a test overwrites — global state; a 1,500-graph fixture — minutes to build |
| D-11 | `api/proto/engine.proto` `GetPersonRecordLinksResponse` gains `bool not_counted = 3;` and `bool lines_dropped = 4;`; its comment says: not_counted is true when the read passed its work bound and compared nothing (every line weak, or no pair when the lines alone passed it); lines_dropped is true when more lines were found than the 75,000 sent, strong lines sent first. `ListPersonClaimsRequest` and `GetPersonFamilyRequest` comments become "An unknown or merged handle, or one that is no Person, is conclusiondetails.not_found." | P-10 and P-21; additive fields old clients ignore | A repeated "dropped count" — the note S16-04 draws is a yes or no |
| D-12 | The handler's response building moves into `recordLinksProto(links persondetail.PersonRecordLinks) *engine.GetPersonRecordLinksResponse` *(new, `person_detail.go`)*, which `GetPersonRecordLinks` calls; it also sets the two fields. `TestRecordLinksProto` *(new)* table-tests it | The map's Check asks for a handler test of a read that did not count; the mapper is what the handler adds, and `persondetail` tests prove the read (T-9) | A handler test over a real catalog past `MaxReadWork` — see D-10 |
| D-13 | Swift: `CatalogPersonRecordLinks` gains `var notCounted = false` and `var linesDropped = false`; `GoStore.getPersonRecordLinks` sets them from `resp.notCounted` and `resp.linesDropped`. `PersonRecordLinksContent` loses `maxArcs`, the `valid` count and the two passes: it appends one arc per line of each kept pair, in the read's order. `maxNodes` stays. Nothing in Swift reads the two fields yet | The rule: Go owns the line bound and Swift's cap goes in this PR; the circle's node cap stays in Swift; S16-04 reads the fields | Keeping `maxArcs` as a second guard — two owners of one bound |
| D-14 | Timing tests use `databasetest.Fastest(2, …)`, `Since` and `WithinBudget` with a 1 s budget (bar 2, "under a second"). Building the fixture is outside the timed call. A catalog read is timed cold: each run calls `c.Graph().Drop()` then `catalogsession.WarmOnOpen(c)` outside `Since`, as `TestRecordLinksBudgets` "cold" does. Each test is at least 20 times past its cap: 4,000 children (40×), 50,000 parents (5,000× `MaxPivotEnd`), 4,000 places (200×), 20,000 records in one graph (2,000×), 1,500,000 lines (20×). Each also asserts the cap's count (100, 0 implied, 20, 10, 75,000) | Rule "Each cap has a timing test at least 20 times past it"; a count catches a cap that is skipped on a fast machine. Sizes are where today's code fails the budget (planning probes) with fixtures that build in seconds | Bar 2's 20,000 children — about 30 s to build per run of the suite |
| D-15 | `graphcachetest.Canon` changes three ways. (a) `next` puts the counter in bytes 13–15 (`id[13], id[14], id[15] = byte(k.n>>16), byte(k.n>>8), byte(k.n)`), so one Canon mints 16,777,216 ids. (b) `Bulk(build func())` *(new)*: while `build` runs, `observe` writes its Observation, skips its recompute and adds the entity to a set; when `build` returns, `Bulk` calls `autoreconciler.RecomputeReported(k.db, set)` once and appends what it reports to `k.pending`, exactly what the per-Observation recomputes would have added. The caller then calls `Commit` as usual; it lists nothing extra. (c) `Members(entity []byte, n int)` *(new)*: reads the entity's `subject_type_id` from `canonical_entities`, then in one `db.Begin` transaction inserts n rows each into `subjects (id, ref, source_id, graph_id, subject_type_id)` (`id` from `next(3)`, ref `SUB-M-<k.n>`, `k.source`, `k.graph`) and `identity_claims (id, subject_id, entity_id, subject_type_id, status)` (`id` from `next(9)`, status `accepted`), with no Observation, as `Entity` inserts its own Subject and claim; the caller then calls `Commit(entity)`, which reloads the node and its members | Planning: 300 children through `Canon` took 26 s (each `observe` recomputes); skipping the per-Observation recompute built 4,000 in 5.6 s; one transaction built 20,000 records in 0.4 s. `Bulk` records what it skipped, so no caller owes a list (review M-3). Existing `Canon` tests keep the per-Observation recompute outside `Bulk` | Deferring every `observe`'s recompute to `Commit` for all `Canon` users (review M-3's first shape) — changes when kept rows appear for six packages' tests, which this PR does not re-check; a raw-SQL bulk writer in each test, as `promotealign`'s `bulkCanon` — a third copy of the catalog's insert shapes |
| D-16 | The places timing test lives in `core/database/conclusionsummaries/eventplaces_cap_test.go` *(new, package `conclusionsummaries_test`)* on `graphcachetest.OpenCatalog`, `nameformats.Warm` and `Canon`; the family and claims timing tests in `core/database/persondetail/caps_test.go` *(new)* on `kinWorld`; the kinship one in `core/kinship/families_test.go`; `CapLines` in `core/recordlinks/recordlinks_test.go` | Each beside the read it bounds | One crafted-catalog test file for all — crosses packages |

**Rules this plan bends.** The map's working rule "Crafted fan-out is capped, not engineered" says "a cap matching does not tune is a package constant beside the read it bounds"; for `KinOf` that contradicts `.cursor/rules/match-rules.mdc` ("Numbers live only in `core/matchregistry`"). The user chose the registry knob (2026-11-09, review M-1): D-1. The map rule now adds "a cap inside a matching layer package is a `Rules` knob in `core/matchregistry`" (updated with this plan). Beyond that, the caps supersede Spike 13's "nothing is hidden for good" and "Go still returns every member" for crafted data, as the working rule "Crafted fan-out is capped, not engineered" already says (user, 2026-11-08); STEP-8 updates the Swift comment that states it.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `kinship.Rules.MaxKinPerRole` *(new)* | registry knob | `kinship.Rules.MaxPivotEnd`, `recordlinks.Rules.MaxEntityMembers` | Same shape: a malformed-data stop in a layer package lives in the registry with an inventory row |
| `persondetail.shownKin` *(new)* | func | the `kinOf` closure in `Families` (line 84) | Replaces the closure's filter half; the closure's re-read goes |
| `persondetail.boundedMembers` *(new)* | func | `(*graphcache.Node).AcceptedMembers` | `AcceptedMembers` serves the detail's member count and other readers that need every member; the cap is the Person page's |
| `conclusionsummaries.eventLocations` *(new)* | func | `placesOfEvents` (line 859), `follow` | `placesOfEvents` returns place ids of many events for the place graph; three callers need one event's location links. `placesOfEvents` now loops over it, so the cap is in one place |
| `conclusionsummaries.maxEventPlaces` *(new)* | const | `summaryChainDepth` (`placechains.go`) | That bounds how far a chain climbs, not how many places an event names |
| `persondetail.maxRecordLinkLines` *(new)* | const | `PersonRecordLinksContent.maxArcs` (deleted) | Moves to Go |
| `recordlinks.CapLines` *(new)* | func | the strong-first passes in `PersonRecordLinksContent.init` (deleted) | Moves to Go, beside `Pairs` and `StrengthStrong` |
| `persondetail.recordLinks` *(new)* | func | `RecordLinks` | `RecordLinks` becomes its one-line caller |
| `PersonRecordLinks.NotCounted`, `LinesDropped` *(new)* | fields | — | |
| `GetPersonRecordLinksResponse.not_counted`, `lines_dropped` *(new)* | proto fields | — | |
| `handlers.recordLinksProto` *(new)* | func | the inline loop in `GetPersonRecordLinks` | Moves out of the closure so a test reaches it |
| `CatalogPersonRecordLinks.notCounted`, `linesDropped` *(new)* | Swift fields | — | |
| `graphcachetest.(*Canon).Bulk`, `Members` *(new)*; `next` widened | test helpers | `Canon.observe`, `Canon.Entity`, `promotealign` `bulkCanon` | Extends the shared builder; `bulkCanon` is a private raw-SQL copy for one package |
| `TestKinOfCap`, `TestSiblingsParentBound`, `TestCraftedFamily`, `TestCraftedClaims`, `TestClaimsRecordsPerGraph`, `TestCraftedEventPlaces`, `TestEventPlacesCapInputs`, `TestCapLines`, `TestRecordLinksSaysWhenPartial`, `TestRecordLinksProto`, `TestPairsCounted` *(new)* | tests | §8 | |

**Builds on.** The working rule's cap shape (`maxWalkEdges`, `maxArcs` provenance); the match registry's `MaxEntityMembers` and `MaxPivotEnd`, read, not copied; `personNode`; S14-40's `placesOfEvents`; `databasetest` budgets; `graphcachetest.Canon`.

**Deletes.** `PersonRecordLinksContent.maxArcs`, its `valid` count, its `passes` and the strong-first comment; the `Families` `kinOf` closure's re-read of the person's `Relationships`; `conclusiondetails.Handle` calls in `Claims` and `PersonFamily`; the `n.Is(typeID)` gate in `Families`' first loop; the `follow(…, hops.PlacesOfEvent)` calls in `leadLife`, `lifeOwners` and `composeEvents`; the Swift test `arcsAndNodesStopAtTheCap`'s arc half; the comment "It has no MaxPivotEnd bound" on `Siblings`; "Go still returns every member" in `PersonFamilyContent.swift`.

**Next adds.** S16-04 reads `notCounted`, `linesDropped` and its own node cut into one partial state. S16-13 keeps `boundedMembers` at the record read. S16-16 rewrites `Families` around the capped `KinOf`. S16-19 keeps `placesOfEvents` and `eventLocations` for an event's places. S16-12 changes `RecordLinksGraph` beside these fields.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Family section with a role cut at 100 | `PersonFamilyContent` folds and Show more | Content is laid out from what Go sends; any count draws. No new id or referent |
| A Sources section with a graph's records cut at 10 | `PersonSourcesContent` | Same: fewer claims in a group draw as fewer rows |
| An event summary with places cut at 20 | Event page, Events list, Timeline, the person's birth and death rows, search documents | Each reads `EventSummary.Places` of any length |
| A record links read with fewer lines than found | `PersonRecordLinksContent` | Draws what it gets; S16-04 adds the note |

### Scale

At the map's real Scale (a person in 50 Evidence graphs with one or two records per graph, 20 children in four parent sets, one to three places per event, 3,675 lines at 50 household graphs) no cap is reached: each read does the work it does today plus one counter per role, graph, event or line. `TestRecordLinksBudgets` (cold 1.5 s, repeat 300 ms at 50 graphs; 1.5 s at 65) and `TestPersonFamilyQueryCount` hold it. At the crafted sizes (bar 2) each capped read was measured in planning: Family 0.17 s at 4,000 children, an event's summary 3 ms at 4,000 places. The work still linear in the crafted input, before the cap: `Relationships` merges all of a person's links into one `LinkSet`; `Follow` returns and records all of an event's location links; `AcceptedMembers` lists all of a person's members; `readPlan` and `Pairs` run before `capLines`, bounded by `MaxReadWork`. No write changes; no cache key changes.

### Edges

- **Assumptions.** `FamilyUnits` reads `kinOf[id]` for each id of `kin.Around(person)`; D-3 builds `Around` from the shown capped kin, a subset of the relatives loaded and read, so every entry exists (T-3). `KinOf` keeps link order inside each role; the cap takes a prefix, so `Around`'s order and `parentUnits`' parent indices hold (T-1). `parentUnits` places a sibling by its shared-parent shapes against the kept parents; with more than `MaxPivotEnd` parents `Siblings` implies none (D-2) and `parentUnits` does not pair, so no sibling carries a shape for a parent past the cap. `eventPlaces` drops duplicate ids; `eventLocations` keeps every link to a kept place, so `lifeOwners` still lists one owner per association (T-6). S14-40's tests (`TestInputsMatchIsolatedComposition`) require the batch and the isolated row to read the same places; both go through `eventLocations` (T-7).
- **Neighbors.** 100 links in a role keep all 100, 101 keep the first 100 (T-1). 10 parents imply siblings as today, 11 imply none (T-2). 20 distinct places keep all, 21 keep the first 20, and 21 links to one place keep all 21 (T-6). 10 records in one graph keep all, 11 keep the first 10, and 10 records in each of two graphs keep all 20 (T-5). Exactly 75,000 lines: nothing dropped, `LinesDropped` false; 75,001: dropped (T-8). Work equal to `MaxReadWork` counts; one more does not (T-10).
- **Removed inputs.** Past a cap, the dropped relatives, records, places and lines are not shown and their summaries are not read; nothing says so except the record links read (rule). A record past the cap still feeds the record links count, which reads the whole Evidence layer. A place past the cap is in no search document of the event, so searching its name does not find the event. Its location link's keys are still recorded by `Follow`, so a write to it recomposes the row (over-invalidation, never stale): when a kept link is deleted, the next place moves in on the recompose (T-7).
- **Real data.** No table or seed drives these rules. The caps against real data: 100 per role against families of a few dozen; 10 parents against two to four; 20 places against one to three; 10 records per graph against one or two; 75,000 lines against 3,675 at 50 graphs.
- **Untrusted size.** Bounded before the costly work: relatives per role (100) before `LoadRelationships(relatives)` and `FamilyUnits`; parents for implied siblings (10) before the sibling loop; records per graph (10) before the graph scans and the handlers' label maps; places per event (20) before `g.Nodes`, `buildPlaceGraph` and `FoldLocations`; lines (75,000) before the wire. Left linear and unbounded: a person's links (`Relationships`), an event's location links (`Follow`), a person's members (`AcceptedMembers`), a person's Evidence graphs (Out, idea row), and `canonlinks`' batched loads of a hub's neighbors.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/kinship` | `Rules.MaxKinPerRole` *(new)*; `KinOf(person, links, terms, rules)` capped (D-1); `Siblings` parent bound and comment (D-2) |
| `core/matchregistry` | `defaultKinship` sets `MaxKinPerRole: 100` with its `decided:` line (D-1) |
| `core/recordlinks` | `CapLines` *(new)* (D-9); `func Pairs(page []byte, graphs []Graph, metas subjectcompare.Metas, cfg Config) (pairs []Pair, counted bool)` (D-7); every test call site takes the second value |
| `core/database/persondetail` | `Families` (D-3, D-4), `shownKin` *(new)*; `PersonFamily`, `Claims` through `personNode` (D-4); `boundedMembers` *(new)* (D-5); `PersonRecordLinks.NotCounted`, `LinesDropped`; `recordLinks` *(new)*, `maxRecordLinkLines` *(new)*; `linksBetween(c, page, graphs, cfg)` (D-8 – D-10) |
| `core/database/conclusionsummaries` | `maxEventPlaces`, `eventLocations` *(new)*; `placesOfEvents`, `leadLife`, `lifeOwners`, `composeEvents` read it (D-6) |
| `core/database/graphcache/graphcachetest` | `next` widened; `Canon.Bulk`, `Canon.Members` *(new)* (D-15) |

Errors: none new. A non-Person id is `conclusiondetails.ErrNotFound` (`conclusiondetails.not_found` on the wire), as other `persondetail` reads already return.

### FFI

`api/proto/engine.proto`: `GetPersonRecordLinksResponse` fields `not_counted = 3`, `lines_dropped = 4` *(new)*, comments per D-11. Handler `api/ffi/handlers/person_detail.go` `recordLinksProto` *(new)* (D-12). Regenerate with `./scripts/generate-proto.sh` (Go and Swift), commit the generated files.

```proto
message GetPersonRecordLinksResponse {
  repeated RecordLinksGraph graphs = 1;
  repeated RecordLinksPair pairs = 2;
  bool not_counted = 3;
  bool lines_dropped = 4;
}
```

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire model | `CatalogPersonRecordLinks` | `macos/App/Platform/GenealogyStore.swift` | `notCounted`, `linesDropped` *(new)*, default false |
| Decode | `GoStore.getPersonRecordLinks` | `macos/App/Platform/GoStore.swift` | Sets both fields |
| Diagram content | `PersonRecordLinksContent` | `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` | `maxArcs` and the strong-first passes go; `maxNodes` stays (D-13) |
| Family comment | `PersonFamilyContent` | `macos/App/Features/Conclusions/PersonFamilyContent.swift` | Line 23 becomes: "Go returns at most 100 relatives per role (`kinship.KinOf`); here only the first paint is bounded." |

States: unchanged; no view changes. Accessibility: unchanged.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Cap kin per role and implied siblings in `kinship`**
Files: `core/kinship/kinship.go`, `core/kinship/families.go`, `core/kinship/siblings.go`, `core/matchregistry/kinship.go`, `core/database/persondetail/family.go` (the two `KinOf` calls take `rules`), `core/kinship/kinship_test.go` (two literals), `core/kinship/families_test.go`, `core/kinship/siblings_test.go`. Skill: `add-match-rule` ("Add a knob").
Do: write T-1 and T-2 first and see T-2's timing case fail; then D-1 and D-2, the `Siblings` comment included; every `KinOf` test call passes `rules`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/... ./core/matchregistry/... ./core/database/canonlinks/... ./core/database/persondetail/...` passes, `TestLayerPackages` included, the existing `TestKinOf`, `FuzzFamilyUnits` and `siblings_test.go` cases unchanged.

**STEP-2 — One capped kin for the Family read; the person gate**
Files: `core/database/persondetail/family.go`, `core/database/graphcache/graphcachetest/graphcachetest.go`, `core/database/persondetail/caps_test.go` *(new)*, `core/database/persondetail/family_test.go`.
Do: widen `next` and add `Canon.Bulk` (D-15); write T-3 and see it fail over budget today; then D-3 and D-4 in `Families` and `PersonFamily`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./core/database/graphcache/...` passes, with `TestPersonFamilyQueryCount`, `TestPersonFamily` and `TestPersonFamilyOriginSiblings` unchanged.

**STEP-3 — Cap a person's records per graph; `Claims` through `personNode`**
Files: `core/database/persondetail/claims.go`, `core/database/graphcache/graphcachetest/graphcachetest.go`, `core/database/persondetail/caps_test.go` *(new)*, added in STEP-2, `core/database/persondetail/claims_test.go`.
Do: add `Canon.Members` (D-15); write T-4 and see it fail today; write T-5; then D-5 and `Claims` through `personNode` (D-4); add the Event row to `TestClaimsMissingAndEmpty` (T-11).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/...` passes, with the claims tests and `TestRecordLinksBudgets` unchanged.

**STEP-4 — Cap an event's places**
Files: `core/database/conclusionsummaries/compose.go`, `core/database/conclusionsummaries/eventplaces_cap_test.go` *(new)*.
Do: write T-6 and T-7 and see T-6's timing fail today; then D-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/searchindex/... ./core/database/persondetail/...` passes, with `TestPersonLifeOwners`, `TestInputsMatchIsolatedComposition`, `TestCanonicalWalks` and `TestEventsByIDs` unchanged.

**STEP-5 — `Pairs` reports whether it counted**
Files: `core/recordlinks/recordlinks.go`, `core/recordlinks/recordlinks_test.go`, `core/recordlinks/invariants_test.go`, `core/database/persondetail/recordlinks.go` (call site only).
Do: D-7 and `CapLines` (D-9); T-10 and T-8; every existing call takes the second value (`got, _ :=` where the test does not check it).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/recordlinks/... ./core/database/persondetail/...` passes.

**STEP-6 — The record links read caps its lines and says when it is partial**
Files: `core/database/persondetail/recordlinks.go`, `core/database/persondetail/export_test.go`, `core/database/persondetail/recordlinks_test.go`.
Do: D-8, `maxRecordLinkLines` (D-9), D-10; T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/...` passes, with `TestRecordLinks` and `TestRecordLinksBudgets` unchanged.

**STEP-7 — The wire carries both flags**
Files: `api/proto/engine.proto`, generated Go under `api/proto/engine/` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/person_detail.go`, `api/ffi/handlers/person_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-11 (fields and the two request comments); `./scripts/generate-proto.sh`; D-12; T-12.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes, with `TestGetPersonRecordLinks`, `TestListPersonClaims` and `TestGetPersonFamily` unchanged.

**STEP-8 — Swift decodes the flags and drops its arc cap**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `macos/App/Features/Conclusions/PersonFamilyContent.swift` (comment), `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`. Skill: `add-swift-test`.
Do: D-13; T-13 replaces `arcsAndNodesStopAtTheCap`; the `PersonFamilyContent` comment per §6 macOS.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` (the command `scripts/check-pr.sh` runs) passes with no Swift warning; `rg -n "maxArcs" macos` finds nothing.

**STEP-9 — Docs and skill lines this PR makes false**
Files: `.agents/skills/add-match-rule/SKILL.md` (lines 33, 34), `docs/matching-model.md` (line 33 paragraph, rows at lines 213 and 217).
Do: line 33 of the skill: `Siblings` implies nothing for a person with more than `Rules.MaxPivotEnd` parents, and `KinOf` keeps at most `Rules.MaxKinPerRole` links per role, for the page. Line 34: `Pairs` reports whether it counted. `matching-model.md` record links paragraph gains: "`Pairs` reports whether it counted; the Person page read (`persondetail.RecordLinks`) sends at most 75,000 lines, strong first, and says when it did not count or dropped lines." A row for `kinship.MaxKinPerRole` *(new)* goes beside `kinship.MaxPivotEnd`: value 100, source "user, 2026-11-08 (Spike 16): more is crafted; registry because kinship is a layer package (2026-11-09)", meaning "Most links `KinOf` keeps in one role for the Person page; past it the first ones in link order". The `kinship.MaxPivotEnd` row's meaning gains "; on the Person page, a person with more parents has no implied sibling (`kinship.Siblings`)"; the `recordlinks.MaxEntityMembers` row's meaning gains "; the Person page's claims read keeps at most this many of a person's records per graph".
Check: `python3 scripts/check-guidance-names.py` passes; `python3 scripts/check-plan-docs.py --links docs/matching-model.md` passes.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status, Landed note), `docs/ideas/person-page-followups.md` (two rows), `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`).
Do: per the Definition of done. The map's two working rules and S16-19's scope already name `eventLocations` and the registry knob (changed with this plan). The two idea rows: "A record links line over `MaxPairWork` or `MaxEntityMembers` keeps 0 facts and nothing says so (S16-01 reports only `MaxReadWork`)" and "Nothing caps how many Evidence graphs one person spans; the Person page reads are linear in it (S16-01)". Version 0.4.1 → 0.4.2.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-01-cap-person-page-reads.md` passes; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/families_test.go` `TestKinOfCap` *(new)* | Person with 100 child links → 100 in `RoleChild`; 101 → the first 100 in link order; 101 parent links and 3 spouse links → 100 parents, 3 spouses; 2,000 child links → `Around` is the first 100 children | D-1, AC-1.1 |
| T-2 | Go | `core/kinship/siblings_test.go` `TestSiblingsParentBound` *(new)* | 10 parents, each with one other child → 10 implied links as today; 11 parents → nil. Edges built as `TestSiblingsManyDuplicatesTimesChildren` builds them (`segmentcompare.Keys` over `segmentcomparetest.Bridge`). 50,000 parents and one sibling sharing all: `Siblings` → `KinOf` → `FamilyUnits` under 1 s (`WithinBudget`), no implied link, parents role 100 | D-2, AC-1.2 |
| T-3 | Go | `core/database/persondetail/caps_test.go` `TestCraftedFamily` *(new)* | `kinWorld` with `Canon.Bulk`: Pat with 4,000 children by `Parent`, one `Commit` of every id; cold `PersonFamily(Pat)` under 1 s; the unit holds 100 children, the first 100 in link order. Fails today (planning: 3.2 s) | AC-1.1, D-3 |
| T-4 | Go | `caps_test.go` `TestCraftedClaims` *(new)* | `kinWorld` Pat with `Canon.Members(Pat, 20000)`; `Claims(Pat)` under 1 s; one group, 10 claims, the first 10 in member order. Fails today (planning: 3.7 s) | AC-1.4, D-5 |
| T-5 | Go | `core/database/persondetail/claims_test.go` `TestClaimsRecordsPerGraph` *(new)* | `kinWorld` Pat with `Members(Pat, 9)` (10 records with Pat's own) → 10 claims; `Members(Pat, 10)` (11) → 10, the first 10 in member order | D-5 |
| T-6 | Go | `core/database/conclusionsummaries/eventplaces_cap_test.go` `TestCraftedEventPlaces` *(new)* | `Canon.Bulk`: a birth event of Pat with 4,000 places, one `Location` each; `EventsByIDs([event])`, `PersonLifeOwners(Pat)` and `RowsForSearch(g, want{event, Pat})` each under 1 s; the event summary holds 20 places, the first 20 in link order; Pat's birth owners list 20 place owners. Table rows: 20 places → 20; 21 → the first 20; one place filed by 21 `Location` links → one place and 21 place owners. Fails today at 4,000 (planning: 1.2 s) | AC-1.3, D-6 |
| T-7 | Go | `eventplaces_cap_test.go` `TestEventPlacesCapInputs` *(new)* | `Canon`: a birth event of Pat with 25 places; `conclusionsummariestest.CheckMatchIsolated` over the event and Pat passes (batch inputs equal isolated composition); after `Canon.Remove` of the first place's location link and `Commit` of the event and the link, `EventsByIDs([event])` holds places 2 to 21 | AC-1.5, D-6 |
| T-8 | Go | `core/recordlinks/recordlinks_test.go` `TestCapLines` *(new)* | `recordlinks.CapLines`: 3 pairs of 2 lines, most 6 → unchanged, false; most 5 → the first 5 in read order, true; strong lines at the end of the read and most 2 → the 2 strong lines, the pairs with none dropped; 1,500,000 lines over 1,500 pairs with 10 strong, most 75,000 → under 1 s, 75,000 kept, 10 strong kept, input slices unchanged | AC-2.5, AC-2.6, D-9 |
| T-9 | Go | `recordlinks_test.go` `TestRecordLinksSaysWhenPartial` *(new)* | `households(f, 4)` (6 pairs × 3 strong lines; 18 lines, so the lines alone count 18 work) through `RecordLinksWith`: default config and 75,000 → both flags false, 18 strong lines; `MaxReadWork` 18 → `NotCounted` true, 6 pairs, every line weak; `MaxReadWork` 1 → `NotCounted` true, no pair; default config and most 5 → `LinesDropped` true, `NotCounted` false, 5 lines | AC-2.4, AC-2.5, D-8, D-10 |
| T-10 | Go | `core/recordlinks/recordlinks_test.go` `TestPairsCounted` *(new)* | The household graphs at 3 with `MaxReadWork` = `ReadWork` → counted; `ReadWork` − 1 by pair work → not counted, every pair listed at 0 facts; a limit below the lines alone → not counted, no pair; no graphs → counted | AC-2.1, AC-2.2, AC-2.3, D-7 |
| T-11 | Go | `claims_test.go` `TestClaimsMissingAndEmpty` (row added) | An Event handle (`f.Birth` promoted) → `ErrNotFound` | AC-3.1, AC-3.2, D-4 |
| T-12 | Go | `api/ffi/handlers/person_detail_test.go` `TestRecordLinksProto` *(new)* | A `PersonRecordLinks` of 2 graphs and 1 pair with both flags false → wire flags false, ids mapped; `NotCounted` true → `not_counted` true; `LinesDropped` true → `lines_dropped` true | AC-2.7, D-12 |
| T-13 | Swift | `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift` `allLinesDrawAndNodesStopAtTheCap` *(new, replaces `arcsAndNodesStopAtTheCap`)* | 80,000 weak lines between two graphs → 80,000 arcs; 1,005 graphs → 1,000 nodes and the one arc between kept graphs | AC-2.9, D-13 |
| T-14 | Swift | `PersonRecordLinksContentTests.swift` `catalogRecordLinksDefaultsToAnOrdinaryRead` *(new)* | `CatalogPersonRecordLinks()` → `notCounted` and `linesDropped` false | AC-2.8, D-13 |
| T-15 | Go | `core/database/persondetail/family_query_test.go` `TestPersonFamilyQueryCount` AC-2.3 (one case added); `family_test.go` `TestPersonFamily` AC-1.12 (unchanged) | `Families` over an unknown id, an Event, a merged Person, a malformed id and Pat → one Family; one and three people issue the same number of queries; *(new case in AC-2.3)* three people plus two unknown ids issue as many queries as three people | AC-3.3, AC-3.2, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestPersonFamilyQueryCount` (AC-2.3 counts queries in `Families`), `TestInputsMatchIsolatedComposition`, `TestPersonLifeOwners`, `TestRecordLinksBudgets`, the `recordlinks` invariants and fuzz tests (new return value).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/matching-model.md (STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
- [ ] Sweep for this PR: `rg -n "maxArcs|no MaxPivotEnd bound|Go still returns every member" macos core .agents docs --glob '!docs/deployment-plan/**'` finds nothing (the spike's plan, map and review logs quote the old names on purpose)
- [ ] The two rows in docs/ideas/person-page-followups.md exist (STEP-10)
```

## 10. Manual check

Bar 2's crafted project cannot be built by hand; T-2, T-3, T-4, T-6 and T-8 prove it at the sizes in D-14. The manual check proves real data is unchanged.

1. Open a real project with a person filed in several Sources who has parents, siblings, a spouse and children, and events with places. Open the Person page on Family: the same parent sets, siblings, partners and children as on `main`.
2. Open Sources: the same groups and records as on `main`; the record links diagram draws the same nodes and lines (thick and thin) as on `main`.
3. Open Timeline: every event shows the same places as on `main`. Open one event's page: its places match.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A real project reaches a cap and loses rows silently | Caps are 5 to 20 times past real data (Edges, Real data); the record links read says so |
| `Canon.Bulk` leaves kept rows stale and Verify fails | `Bulk` recomputes every entity it observed, once, and hands the reported rows to the next `Commit`; run the package once with `PROVENENCIA_GRAPH_VERIFY=1` |
| T-3 still over budget after the cap, from `canonlinks` loading a hub's neighbors per relative | Planning measured 0.17 s with the cap; see stop condition |
| `TestPersonFamilyQueryCount` AC-2.3 counts one more query per id | `Families` keeps `g.Cached` + `n.Is`; `personNode` runs once per `PersonFamily`, in place of its `Handle` call (D-4) |

**Stop and report if:**

- T-3 is over 1 s with D-1 and D-3 in place: the remaining cost is in `canonlinks` (`For` per relative), which this plan does not change.
- `TestPersonFamily`, `TestPersonFamilyOriginSiblings` or `FuzzFamilyUnits` changes output on a case below every cap.
- `TestInputsMatchIsolatedComposition`, `TestPersonLifeOwners` or a `searchindex` rebuild-equals-upkeep test fails after STEP-4.
- `TestRecordLinksBudgets` goes over its budget after STEP-6.
- `Canon.Bulk` cannot keep Verify (`PROVENENCIA_GRAPH_VERIFY=1`) passing for T-3 or T-6.
- An existing `Canon` test changes result after `next` is widened (a test that reads ids or refs by their bytes).
- `./scripts/generate-proto.sh` (protoc, `protoc-gen-go`, `protoc-gen-swift`) or `xcodebuild` is not available where STEP-7 or STEP-8 runs: those steps run on macOS with the repo's toolchain, as `scripts/check-pr.sh` `xcode-test` says. Never hand-edit `engine.pb.go` or `engine.pb.swift`.
- A timing test passes its count assertions but is over budget only on the machine running it: report the figure from `WithinBudget`'s log and the planning figure; do not raise the budget.
- A caller of `recordlinks.Pairs` outside the files in §4 appears at the base.

## 12. Open questions

None.

## 13. PR

**Title:** S16-01 — Cap the Person page reads and an event's places against crafted projects
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-01-cap-person-page-reads.md`.
