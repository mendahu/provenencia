# Deployment plan — Spike 16

Spike 16 is an odds-and-ends spike ([odd and even spikes](../README.md#odd-and-even-spikes)): small fixes, UX refinements and follow-ups, in batches, each a slice with its own Check. Batch 1 takes the high and medium findings of the [Person page audit](person-page-audit.md) (2026-11-08), an area audit of Spike 13's Person page with all five [`review-pr`](../../../.agents/skills/review-pr/SKILL.md) lenses. Findings that edit the same code or the same concern share a PR (user, 2026-11-08), unless the map review found a reason to keep them apart ([review](deployment-plan.review.md)); the rest are parked in idea files ([Where the notes went](#where-the-notes-went)). Batch 2 (2026-11-08) reworks the catalog's caches so the conclusion layer's code gets simpler: entities and indexes held in Go memory and updated exactly on commit, answers composed from them, and Swift following what each value was built from ([design](cache-architecture.md)).

**A brief covers one view.** A PR that adds, changes or removes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a run of small, independent improvements, each batch landing on its own.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project. A batch adds its outcomes here when it is cut.

**Batch 1 — Person page audit fixes** (user, 2026-11-08; [audit](person-page-audit.md)):

1. **A burial-only life reads in order.** A person known only from a parish burial reads Birth unknown, the undated events, Death unknown, then the burial, in the list and on the lifeline. A person with a dated baptism and burial and no death has Death unknown just before the burial, on the list and on the lifeline, so the life between them is drawn before the death.
2. **A crafted project cannot stall the Person page.** A project holding an event with 10,000 places, a person with 20,000 children, a person with 50,000 parents, and 20,000 of one person's records in one Evidence graph opens, and each Person section draws in under a second.
3. **A long list unfolds a page at a time.** A person with 2,000 events, one of them with 500 participants, opens on Timeline without a stall; each Show more, in any section, reveals at most a page of 50 and moves focus to the first item it reveals; every event still has its point on the lifeline. Hovering a graph on the record links diagram whose record line is folded lights up that Source's Show more, and a date edit from another window rescales the lifeline to the new range.
4. **Provenance shows today's words.** Rename a Source, or fix a Citation's label or page: the Vitals provenance sheet, the Gallery captions and the likeness viewer's record row show the new words, as the Sources section does.
5. **A failed reload says so.** When a section's reload fails, its rows stay and a notice with Try again sits above them; a first load that fails offers Try again too; VoiceOver hears both.
6. **The diagram says when it is partial.** A person whose records were too many to compare, or too many to draw, shows a note in the diagram column instead of drawing every link thin or dropping graphs silently.
7. **Keyboard and VoiceOver keep their place.** ⌘1–⌘5 announces the section and leaves focus on the selected segment; removing a record that keeps the person is announced; Show more in Family, Timeline and Sources moves focus to the first item it reveals (bar 3).
8. **One word for a person's records.** The Person page's header counts "7 records" (the Event and Place headers too), the Remove confirm ends "…, and the person stays", and a person deleted from another window leaves with "This person no longer exists".

**Batch 2 — Cache architecture** (user, 2026-11-08; [design](cache-architecture.md)):

9. **No cache empties itself on any write.** After an Observation of a filed person is edited, the next Promote proposal and the Family section reuse the kinship index and the handle profiles instead of rebuilding them; `rg DropOnHandleCommit core` finds nothing, and Verify (`PROVENENCIA_GRAPH_VERIFY=1`) checks every cache.
10. **Promote's cost follows the Source, not the catalog.** On the large fixture at 16,000 persons a warm proposal is at least three times faster than today's on the same machine, the first proposal after a handle commit is no slower than 1.5 times a warm one, and neither more than doubles from 4,000 to 16,000 persons. Its suggestions are the same as today's.
11. **A Person page opens from memory.** With the catalog warm, opening a Person on Vitals, its provenance sheet, its Gallery and its Sources section issues no SQL; a Source rename shows on every page without reloading any Evidence graph.
12. **Lists and progress never rescan the catalog.** The Sources list's graph progress needs no full-table count; the Sources section of a person in 80 Evidence graphs loads no whole graph; a write updates only the rows it touched in the Persons, Events and Places lists.
13. **Walks are lookups.** A city's place chain reads without loading its 6,000 location links; the Family section reads in one pass; the life rows and the Timeline take their owners and participants from summary rows.
14. **Back is instant.** Returning to a Person, Event or Place page shows its last value at once while it refreshes; a write that touches nothing a cached section was built from leaves it alone; a person's other sections are ready before they are opened.
15. **One table names every cache.** `docs/catalog-caches.md` (new in S16-10) lists every Go and Swift cache with what builds it and the signal that updates it; at 16,000 persons the Go caches together stay under 40 MB; every value the app derives from cached data is built once per change of its inputs.

**Scale.** Real data, which must stay as fast as today: a person in 50 Evidence graphs (Spike 13's budget) with one or two records per graph, 200 events with 10 participants each (12 shown per event at Wide), 20 children in four parent sets, one to three places per event. Crafted data, which only has to stay bounded: each axis in bar 2 and 3, at least 20 times past its cap.

**Scale (Batch 2).** The S11-21 large fixture at 4,000 and 16,000 persons (8,000 town events; 18,000 and 37,000 handles); one place with 6,000 events, and 24,000 for growth; a person in 50 and in 80 Evidence graphs; for the entity records, 3,000 Sources, 10,000 Evidence graphs and 50,000 Observations. Every speed claim is measured at 4,000 and 16,000 (or at the input it grows with) against today on the same machine.

## Where the notes went

| Audit item | PR | Brief |
| --- | --- | --- |
| P-1 A missing death sits after a burial | S16-02 | — |
| P-2 Folding an event's places is quadratic | S16-08 | — |
| P-3 The Family read is quadratic in a crafted wide family | S16-01 | — |
| P-4 The claims read rescans its graph per record | S16-01 | — |
| P-5 A failed refresh is silent | S16-04 | S16-D1 |
| P-6 Vitals and Gallery keep old Source words | S16-09 | — |
| P-7 The Timeline builds every event | S16-07 (pages), S16-02 (build what is shown) | — |
| P-10 The diagram draws uncounted links as weak | S16-01 (engine), S16-05 (note) | S16-D1 |
| P-11 Switching sections is silent | S16-06 | — |
| P-12 A Remove that keeps the person is silent | S16-06 | — |
| P-13 Section plumbing | S16-03 | — |
| P-14 Fold focus copied per section | S16-07 | — |
| P-16 The leave-on-delete decision is not tested (map review K-8) | S16-06 | — |
| P-15 Hovering a node whose record line is folded highlights nothing | S16-07 | — |
| P-17 "handle", "members" and "This record no longer exists" on the Person page | S16-04 (the not-found line), S16-06 (the count and the confirm) | S16-D1 (the not-found line) |
| P-20 The Timeline keeps an old scale after a date edit | S16-02 | — |
| P-21 `ListPersonClaims` answers for a non-Person handle | S16-01 | — |
| P-22 (part) The record links wire's unread fields and missing `source_ref` | S16-01 | — |
| P-24 `edgeKey` joins ids with no separator | S16-16 (moved from S16-01 with Batch 2, which rewrites that code) | — |
| P-26 (part) `ConclusionDetailPage.swift` holds 15 types | S16-03 | — |
| `CatalogQueryStalePolicy` is unused ([`structural-debt.md`](../../ideas/structural-debt.md), S14-11 plan) | S16-09 | — |

The last eight rows were parked in idea files and folded into the PRs that already touch their files (user, 2026-11-08).

**Batch 2** came from the [cache review](cache-architecture.md) and pulls in these parked ideas:

| Source | PR |
| --- | --- |
| Cache review: one registry for catalog-wide caches; [`promote-ranking.md`](../../ideas/promote-ranking.md) B-8's "one keyed slot"; `canonicalgraph` left with no production caller by S14-42 | S16-10 |
| Cache review: outcome records on the node; the gallery from the node | S16-11 |
| Entity survey: Source records; a Source rename evicting every loaded graph | S16-12 |
| Entity survey: Subject records; provenance records composed | S16-13, S16-14 |
| Entity survey: the Evidence-graph catalog; graph progress scans | S16-15 |
| Cache review: the kinship index; [`structural-debt.md`](../../ideas/structural-debt.md) row "The kinship memo slots … drop on every handle commit"; P-24 | S16-16, S16-17 |
| Cache review: the place hierarchy index | S16-18 |
| Cache review: summary rows carry life owners and participants; [`structural-debt.md`](../../ideas/structural-debt.md) A-33 (two event–person shapes) | S16-19 |
| [`promote-ranking.md`](../../ideas/promote-ranking.md) "Name index on commit" (B-8, B-9); [`structural-debt.md`](../../ideas/structural-debt.md) row "`graphalign.Handle` has the fields of `subjectcompare.Subject`" (its `Handle` half) | S16-20 |
| [`promote-ranking.md`](../../ideas/promote-ranking.md) "Narrower canon walk" (B-8's walk); the same structural-debt row's `CanonEdge` half | S16-21, S16-22 |
| Cache review (Swift): stale while refreshing; values report their inputs; lists merge touched rows; keys as plain values; one derived-value primitive with [`person-page-performance.md`](../../ideas/person-page-performance.md)'s "Section content is rebuilt on every pass"; prefetch | S16-23 – S16-29 |

**Back in the ideas folder** (user, 2026-11-08) —
- [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md): P-8 (participants sent once per event), P-9 (record links memo and a household-sized budget), P-23 (siblings sent two or three times), and lazy Timeline cards after an unfold. Memoising section content went to Batch 2's S16-27.
- Batch 2 leaves parked: stored summary rows, a canonical links table and stored shortlist keys (rule "No new persisted projections"); a batch Evidence-graph read past 64 graphs ([`structural-debt.md`](../../ideas/structural-debt.md), S13-16 PRF-2); Promote's per-graph compare and alike-pair memos (B-11, [`promote-ranking.md`](../../ideas/promote-ranking.md)).
- [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md): P-18, P-19, and the Conclusion lists' silent refresh failure (map review K-2).
- [`ideas/structural-debt.md`](../../ideas/structural-debt.md): P-22 (the rest), P-25 (A-41), P-26 (the rest), P-27.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Batch 1 — Person page audit fixes

- **(Batch 1) Crafted fan-out is capped, not engineered** (user, 2026-11-08). A cost that only a crafted or badly broken project reaches gets a fixed cap checked before the work, not a faster algorithm. Real data never reaches a cap. Past a cap a read keeps the first items in the order it already reads them and drops the rest without saying so; the record links read is the exception (rule "The record links read is bounded and says so"). The caps:
  - an event's places, 20, applied in one helper that the four reads of an event's places in `conclusionsummaries/compose.go` (`eventPlaceGraph`, `lifeOwners`, `composeEvents`, `eventPlaces`) all use, so no read walks the rest (S16-08);
  - the Family read's relatives per role (parents, children, spouses, siblings), 100, in one capped kin per person that both of `Families`' loops use (the relatives it loads, and `Around` and `FamilyUnits`) (S16-01);
  - implied siblings only through a person with at most `kinship.Rules.MaxPivotEnd` parents, the bound `parentUnits` already uses to stop pairing (S16-01);
  - a person's records per Evidence graph in the claims read, `recordlinks` `Rules.MaxEntityMembers` (10, user 2026-11-07: "malformed") (S16-01).

  A cap matching does not tune is a package constant beside the read it bounds, with a `// decided: user, 2026-11-08` comment, as `graphcache` `maxWalkEdges` and `PersonRecordLinksContent.maxArcs` are; an existing registry value is reused, not copied. Each cap has a timing test at least 20 times past it (bar 2). For crafted data only, this supersedes "nothing is hidden for good" and "Go still returns every member" in Spike 13's fold rules (`spike-13/deployment-plan.md` "Timeline lists fold"; `PersonFamilyContent.swift`). Binds S16-01, S16-08.
- **(Batch 1) The record links read is bounded and says so.** `recordlinks.Pairs` reports whether it counted (it already knows: past `MaxReadWork`, or no pairs when the lines alone pass it). `persondetail.RecordLinks` sends at most `maxRecordLinkLines` lines (75,000, the number the diagram draws today), strong lines first, as a package constant that carries `maxArcs`'s `// decided:` provenance, not a match-registry knob, and the response says when it did not count or dropped lines. Swift's `PersonRecordLinksContent.maxArcs` and its strong-first pass go in the same PR, since Go now owns that bound. The 1,000-node cap stays in Swift, where it bounds the circle, and the diagram's note covers it too: one note says the drawing is partial, whichever side cut it. The note is about the read, not a line, so Spike 13's rule that the view shows no data about a line holds. Binds S16-01, S16-05.
- **(Batch 1) A marker's position is one value.** A life marker that has no span of its own sits beside one dated row or before one, never both: one `oneof` on the wire (beside or before) and one Swift enum, in place of an empty `beside_event_id` that means both "nothing dated" and "nothing qualifies". A placeholder death sits just before the earliest dated after-death row (burial, funeral) whenever one is dated, else beside the last dated row; a placeholder birth mirrors it (just after the last dated before-birth row, else before the first dated row) (user, 2026-11-08: "just before burial"). This replaces Spike 13's "just after the latest dated event that cannot come after death", which drew the life between a baptism and a burial after the death; [R-24a](../spike-13/person-detail.md#3-timeline) and Spike 13's rule "The timeline is a lifeline" say so. Swift places what Go names, and the lifeline puts a "before" stub at that row's start (Spike 13 rule "Positions are data"). Binds S16-02.
- **(Batch 1) One paged fold that moves focus** (user, 2026-11-08: a render cap or load more, kept simple; map review K-1). `ConclusionFold`, already "the one fold rule of a Conclusion page's lists", gains a page: each Show more reveals at most 50 more, in every section, and one fold state type (which folds, how many pages each has revealed) replaces the per-section `Set`s and Sources' `rowsExpanded`. `firstRevealed` names the first item of the page just revealed, and `ConclusionShowMoreRow` claims focus for it through `FocusClaim`; the private `Task` hops go. One focus keeper in `macos/App/Platform/`, beside `FocusClaim`, holds focus across a width change for Gallery and Timeline ([`macos-client-patterns.md`](../../macos-client-patterns.md) Wide rule 5). The folds' first limits stay as they are (50 events, 5 / 12 participants; user, 2026-11-03). Binds S16-07, S16-02.
- **(Batch 1) The Timeline builds what it shows.** `PersonTimelineContent` builds rows and person cards only for what the folds show; a folded row keeps its id, span and spoken label for the lifeline and the markers. The lifeline draws every point and makes a click target only for a shown row. Binds S16-02.
- **(Batch 1) One open section, one list of reads.** `WorkspaceLocation.openPageSection` resolves the open page section once (`pageSection ?? offered.first`). `ConclusionPageSection` declares its reads as one list, each tagged own, companion or supporting, in place of `queryKey`, `companionQueryKey` and `supportingQueryKeys`. The place registry, the page's refreshing state, Try again (S16-04) and the section hosts all read that list. The section host's identifier root comes from the page identifier. [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) gains the page-section sub-place and [`macos-client-patterns.md`](../../macos-client-patterns.md) its sectioned-page lines. Binds S16-03, S16-04, S16-06.
- **(Batch 1) A refresh failure is read beside its value** (map review K-2). `QueryHandle` gains one reading that is loading, failed, or loaded with an optional refresh failure, and a failure revision that bumps on each failure so a notice can re-announce; `read` is unchanged for its other callers. `ConclusionDetailPresentation` is built from it, and the page and each section host show the kit's `PVRetryCallout` (it announces and has Try again; precedent S14-D7) above the stale content, in its compact form in a column, and on a first-load failure. The page frame is shared, so Event and Place pages gain it too. The Conclusion lists' presentation adopts it later, with its own brief ([idea](../../ideas/person-page-followups.md)); `EvidenceGraphModel`'s and `PromoteModel`'s attempt counters may follow. [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) ("Stale-while-revalidate") and the `QueryHandle` row of [`macos-client-patterns.md`](../../macos-client-patterns.md) teach it. Binds S16-04.
- **(Batch 1) A cached detail follows the Sources its records cite.** `.conclusionDetail` refreshes when a write names a Source that one of its field records, derived-row owners or pictures cites (`applyHiddenPolicy(followsValue: .source, …)`), so a Source rename and a Citation or Artifact label edit reach Vitals, the Gallery and the viewer, on Event and Place pages as well. Fan-out: `sourceScopes` is the audit "last activity" scope, which 16 effect entries set, so any write inside a cited Source revalidates a visible detail (one read, `GetConclusionDetail`); a hidden detail is evicted, as today. [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md)'s slot list gains `.source`, and [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys" says what the detail key follows. Binds S16-09.
- **(Batch 1) Every outcome on the page is spoken.** No Toast is a visual choice (S13-17 RC-6), not a silence: a section switch announces the section, and a Remove that keeps the person announces it, with spoken-only strings. The leave-or-stay decision after a Remove moves out of `PersonSourcesSection.confirm` into a function a unit test reaches, which returns leave, or stay with the announcement. Binds S16-06.
- **(Batch 1) A person's records are records** (user, 2026-11-08). The page uses one word for the Identity Claims behind a canonical entity. The header's count reads "7 records" on every Conclusion page (one string, `conclusions.memberCount`). The Unpromote confirm ends "…, and the person stays", worded per kind like `confirmMessageDeletes`, the Evidence graph's confirm included. A Person page whose person was deleted or merged elsewhere says "This person no longer exists" in its leave Toast and its section notices; the kind supplies that line, so Event and Place keep the shared one. Binds S16-04, S16-06.
- **(Batch 1) Refused: memoising and re-shaping for speed** (user, 2026-11-08). P-8 and P-9 are descoped, and P-7 gets a render cap and pages, not a memo. Their notes are in [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md); measure before pulling any in. Batch 2 later builds every section's content once per value through one derived-value primitive (S16-27), as part of the cache architecture rather than as a speed fix (user, 2026-11-08); P-8 and P-9 stay descoped.

### Batch 2 — Cache architecture

- **(Batch 2) Cache entities and indexes; compose answers** (user, 2026-11-08). Go memory holds records keyed by id (canonical nodes with their outcome records, Subject records, Source records, the Evidence-graph catalog, the vocabulary) and indexes for walks that repeat across requests (kinship, place hierarchy, Promote's profiles and name index). Readers compose their answer from them at read time. Summary rows (displays) stay the one stored answer, because search needs them and S14-42 keeps them exact. Binds S16-10 – S16-29.
- **(Batch 2) Every cache names the signal that updates it, and updates exactly.** Each cache maps to `effects.Set` fields (`Handles`, `SourceScopes`, `EvidenceGraphs`, `Reads`, `Deleted`, `SummaryDeps`) or to a vocabulary or structure drop, and updates only the entries those name. Nothing empties itself on any handle commit: `DropOnHandleCommit` and `dropHandleMemos` go with their last user (S16-16), and dropping the match index on every handle commit goes with S16-20. A new rule of that kind fails review: S16-10 adds the Known trap to [`review-pr`](../../../.agents/skills/review-pr/SKILL.md)'s maintainability lens. Binds S16-10 – S16-22.
- **(Batch 2) One registry for catalog-wide caches.** A graph index registers with Build (lazy, on any graph), Apply (on commit, with the commit's `effects.Set`, on the session graph), Inputs (the `InputKey`s an entry was built from), Check (run by Verify against a fresh Build), DropOnVocabulary and a memory Cap ([design](cache-architecture.md#one-home-for-a-catalog-wide-cache-s16-10)). Node memos stay for results per node and Evidence-graph memos for results per graph. The match index, alias index and name format move onto it. Every Go cache this batch adds registers there and has a row in `docs/catalog-caches.md` (S16-10). Binds S16-10, S16-12, S16-13, S16-15, S16-16, S16-18, S16-20.
- **(Batch 2) Built lazily on any graph, kept only on the session graph.** A write composes summary rows on a one-shot graph over its transaction (`searchindex` `newGraph = graphcache.Over`); an index a composer reads builds there on first use and is discarded with it, and is kept and updated only on the session graph. Binds S16-18, S16-19.
- **(Batch 2) An index a composer reads hands its inputs to the row.** A summary composer that reads an index instead of `Node` and `Follow` adds the entry's recorded `InputKey`s to its `InputScope`, keeping S14-40's input cap and its "stop the chain at the depth its inputs cover" rule, so S14-42's lookup still finds every dependent row. S14-40's tests (inputs equal isolated composition; stored inputs equal a rebuild after every upkeep write) hold it. Binds S16-18, S16-19.
- **(Batch 2) A Source's words live once.** Title, ref, type, cover, credibility and Artifact labels live in the Source record. `EvidenceGraph` and Subject records hold ids; a provenance record is the node's outcome row, the Subject record and the Source record composed. The `sources` effect stops naming a Source's graphs for a title or type change once no reader takes Source words from a graph, and the app's claims and record links keys follow `.source` instead. Binds S16-12, S16-13, S16-14.
- **(Batch 2) Warm at open, update on commit; no background re-warm.** The kinship index and Promote's full profiles are built in `WarmOnOpen` from the nodes `promotealign.Warm` already loads, behind the open spinner; other caches build on first read. After a write, Apply keeps each cache current inside `OnCommit`; nothing re-warms in the background, because every engine call waits in `catalogsession`'s one queue. Binds S16-16, S16-20.
- **(Batch 2) Orders are defined and results do not move.** Every index defines its iteration order; the kinship index keeps `For`'s order. Promote's suggestions are held equal to today's by an equivalence test on the large fixture (S16-22) and the matching goldens; Family, Timeline, summaries and search are held by their existing tests and the rebuild-equals-upkeep tests. Binds S16-11 – S16-22.
- **(Batch 2) Memory is budgeted.** Every cache has a cap and a size measured at 16,000 persons, recorded in `docs/catalog-caches.md` (S16-10); the Go caches together stay under 40 MB there (est. 20–40 MB). Subject records keep an LRU of their own. Binds S16-10 – S16-22.
- **(Batch 2) Batch 1's caps stay.** The kinship index keeps S16-01's relatives-per-role and sibling caps; the place index keeps S16-08's places-per-event cap; the claims read keeps S16-01's records-per-graph cap when it reads Subject records. Binds S16-14, S16-16, S16-17, S16-18.
- **(Batch 2) Swift caches one value per view, follows its inputs and shows it stale while refreshing.** Per-entity Conclusion keys go from `.evict` to `.markStale`, and only ids in `effects.deleted` are evicted, with a cap of 50 hidden entities. Each per-entity read reports the handles, graphs and Sources it was built from, and one `.followsInputs` refresh replaces the hand-written selectors. Swift derives nothing the engine knows; one `session.derived` primitive builds what it derives for display once per change of its inputs. Binds S16-23 – S16-29.
- **(Batch 2) No Go cache for a view that is one query.** The Source workspace lists, source types, metadata fields, Citations by Artifact and subject positions keep their single reads; Swift's per-view key is enough. Binds S16-10 – S16-29.
- **(Batch 2) No new persisted projections, and no stored answer without a measurement.** Stored summary rows, a canonical links table and stored shortlist keys are refused for now ([design](cache-architecture.md#not-this-batch)). A composed answer is memoised only after it measures slow at the batch's Scale. Binds S16-10 – S16-29.
- **(Batch 2) Measure growth.** A PR that claims speed adds a budget at 4,000 and 16,000 persons on the large fixture (or at the input it grows with), and records today's figure on the same machine in its ship log. Binds S16-11 – S16-29.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S16-01 | Cap the Person page reads against crafted projects | 1a | Go: the Family read's relatives per role and implied siblings, and a person's records per graph in the claims read, capped before the work; the record links read sends at most 75,000 lines and says when it did not count or dropped lines, and Swift's own arc cap goes; one Person check for every read, and `source_ref` on the record links wire in place of two unread fields | — | — | — | Scoped |
| S16-08 | Cap the places an event folds | 1a | Go: one helper for an event's places, deduped and capped at 20, used by the four reads in `conclusionsummaries` | — | S14-40 (Spike 14) | — | Scoped |
| S16-07 | One paged fold that moves focus | 1a | Preparatory, Swift: `ConclusionFold` reveals a page of 50 at a time in every section, one fold state type, Show more moves focus to the first revealed item (Family gains it), one focus keeper across width changes; hovering a graph whose record line is folded lights up its Show more | — | — | — | Scoped |
| S16-02 | The Timeline places a missing death and builds what it shows | 1a | Go and Swift: a placeholder life marker sits beside or before one row, so a missing death precedes a burial; the Timeline builds rows and cards only for what the folds show and makes click targets only for shown rows; a date edit rescales the lifeline | — | S16-07 | — | Scoped |
| S16-09 | Conclusion pages follow the Sources their records cite | 1b | Swift: `.conclusionDetail` refreshes when a write names a Source its records cite, so Vitals, the Gallery and the viewer show today's Source, Citation and Artifact words; the unused `CatalogQueryStalePolicy` goes | — | — | — | Scoped |
| S16-03 | One open section and one list of section reads | 1b | Preparatory, Swift only: `WorkspaceLocation.openPageSection` and one read list per `ConclusionPageSection`, read by the place registry, the page, the Person fill and the section hosts; `add-workspace-location` teaches page sections; no visible change; `ConclusionDetailPage.swift` split into the files its types belong to | — | — | — | Scoped |
| S16-04 | A failed reload keeps the rows and says so | 1b | A `QueryHandle` reading with a refresh failure beside the value; the page and each section show an announced Try again notice above stale rows and on a first-load failure; a deleted person's page says the person no longer exists | S16-D1 | S16-03, S16-D1 | — | Scoped |
| S16-05 | The record links diagram says when it is partial | 1b | Swift: the diagram column shows a note when the read did not count facts, Go dropped lines, or the circle dropped graphs | S16-D1 | S16-01, S16-D1 | — | Scoped |
| S16-06 | Section switches and removals are spoken | 1c | Swift: ⌘1–⌘5 announces the section and focus lands on the selected segment (a `PVSegmentedControl` fix); a Remove that keeps the person is announced, from a leave-or-stay decision a unit test reaches; the header counts records and the Remove confirm says the person stays | — | S16-03 | — | Scoped |
| S16-10 | Graph indexes: one home for catalog-wide caches | 2a | Preparatory, Go: a registered graph index (lazy Build, Apply on commit, Inputs, Check, cap); the match index, alias index and name format move onto it; `canonicalgraph` goes; `docs/catalog-caches.md` lists every cache; no behavior change | — | S14-42 (Spike 14) | — | Scoped |
| S16-11 | Outcome records live on the canonical node | 2b | Go: each node holds its outcome rows and chosen likeness; the detail read's reconciler SQL and `gallery.ForEntities` go, and the gallery composes from the node | — | S14-42 (Spike 14) | — | Scoped |
| S16-12 | Source records in the graph cache | 2b | Go and Swift: one record per Source (words, type, cover, credibility, Artifacts), updated from `SourceScopes`; graphs stop copying Source words, a rename stops evicting graphs, and the claims and record links keys follow `.source` | — | S16-10 | — | Scoped |
| S16-13 | Subject records in the graph cache | 2b | Go: one record per Subject with its claim and Observations, built lazily and dropped by graph; provenance records compose from node, Subject and Source records, and `sqlOutcomes` goes | — | S16-10, S16-11, S16-12 | — | Scoped |
| S16-14 | The Sources section reads Subject records | 2b | Go: the claims read composes from Subject and Source records instead of loading whole graphs and rescanning them, keeping S16-01's cap | — | S16-13, S16-01 | — | Scoped |
| S16-15 | An Evidence-graph catalog in the graph cache | 2b | Go: each graph's shell and counts and each Source's graphs, updated from `Reads`; graph progress reads it instead of full-table counts, and claims and record links order graphs from it | — | S16-10 | — | Scoped |
| S16-16 | A kinship index kept current on commit | 2c | Go: one index of recorded and implied relationship links by person, updated from each commit's roots; `For` and siblings read it; the `for` and `siblings` memo slots, `DropOnHandleCommit` and `dropHandleMemos` go; `edgeKey` made unambiguous | — | S16-10, S16-01 | — | Scoped |
| S16-17 | The Family read is one pass over the kinship index | 2c | Go: `Families` reads the index once, keeping S16-01's caps; the `canonlinks` loaders only it used go | — | S16-16 | — | Scoped |
| S16-18 | A place hierarchy index | 2c | Go: every place's shell and its `part_of` and `succeeded_by` links with their input keys, built lazily on any graph and updated on the session graph; the place-chain builders collapse onto it | — | S16-10, S16-08, S14-42 (Spike 14) | — | Scoped |
| S16-19 | Summary rows carry their life owners and participants | 2c | Go: person rows keep their birth and death owners, event rows their date's Property and participants by role; `PersonLifeOwners` goes and the Timeline reads participants from rows; one event–person shape | — | S16-18, S16-02 | — | Scoped |
| S16-20 | Promote's profiles and name index survive writes | 2d | Go: handle profiles as a node memo with exact `Deps`; the shortlist index on the registry, patched on commit; only Subjects whose keys meet a changed handle re-rank; `Warm` fills profiles | — | S16-10 | — | Scoped |
| S16-21 | Promote's place-link checks read the place index | 2d | Go: `Connections` takes the filed `part_of` and `succeeded_by` links from the place index instead of every canon edge | — | S16-18 | — | Scoped |
| S16-22 | Promote reads each handle's neighbourhood instead of walking the catalog | 2d | Go: `Align` reads a `CanonView` over the kinship index and profiles; `loadCanon`'s walk and its link merge go; results held equal on the large fixture | — | S16-16, S16-20, S16-21 | — | Scoped |
| S16-23 | Conclusion pages show their last value while they refresh | 2e | Swift: per-entity Conclusion keys mark stale instead of evicting; only deleted entities are evicted; hidden ones capped at 50 | — | S16-04 | — | Scoped |
| S16-28 | Cache keys as plain values | 2e | Preparatory, Swift: `CatalogQueryKey` becomes a value of kind, project and ids, with the same factories; the key enum's switches and `guard case` blocks go | — | — | — | Scoped |
| S16-24 | Person, Event and Place reads report their inputs | 2e | Go: the detail and section reads return the handles, graphs and Sources they were built from | — | S16-17, S14-42 (Spike 14) | — | Scoped |
| S16-25 | Swift keys follow the inputs their values report | 2e | Swift: one `.followsInputs` refresh replaces `personSection`, `personGraphsSection`, the per-kind id closures and S16-09's Source follow | — | S16-24, S16-28, S16-03, S16-09 | — | Scoped |
| S16-26 | Conclusion lists merge the rows a write touched | 2e | Go and Swift: the Persons, Events and Places lists re-read only the touched rows by id with Go's sort key, through the existing merged read | — | S14-42 (Spike 14) | — | Scoped |
| S16-27 | One primitive for values derived from cached data | 2e | Swift: `session.derived` builds a value once per change of its inputs; the four hand memos and the section host's per-pass build move onto it | — | S16-03, S16-07 | — | Scoped |
| S16-29 | Prefetch a person's other sections and a hovered person | 2e | Swift: after the open section settles, warm Family and Timeline; after a 250 ms hover on a person, warm their page | — | S16-23, S16-25 | — | Scoped |
| S16-99 | Dogfood close / docs | — | Bar pass; the ship logs' follow-ups triaged; the audit's [Close the loop](person-page-audit.md#close-the-loop) classes added to the lenses' Known traps; [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S16-D1 | Person page: notices for a failed reload (the header band, a full-width section, the Sources diagram column) and for a record links diagram that is partial | **S16-04**, **S16-05** | An enhancement of the Spike 13 Person page boards (S13-D1, S13-D10). Reuses `PVRetryCallout` as the Promote page does (S14-D7). Draws Regular and Wide. The Conclusion page frame is shared, so the board also says the Event and Place pages take the same notice. Copy: the notice titles, Try again, the diagram's partial note |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
BATCH 1 — Person page audit fixes

SLICE 1a — Bounded reads, folds and the Timeline
  S16-01  Cap the Person page reads against crafted projects
  S16-08  Cap the places an event folds                      (after S14-40)
  S16-07  One paged fold that moves focus                   (preparatory)
  S16-02  The Timeline places a missing death and builds what it shows
  Check: bar 1, 2 and 3.

SLICE 1b — Current data and honest notices
  S16-09  Conclusion pages follow the Sources their records cite
  S16-03  One open section and one list of section reads    (preparatory)
  ✎ S16-D1 ──▶ S16-04  A failed reload keeps the rows and says so
  ✎ S16-D1 ──▶ S16-05  The record links diagram says when it is partial
  Check: bar 4, 5 and 6.

SLICE 1c — Keyboard and VoiceOver
  S16-06  Section switches and removals are spoken
  Check: bar 7 (with S16-07's focus moves).

BATCH 2 — Cache architecture

SLICE 2a — One home for caches
  S16-10  Graph indexes: one home for catalog-wide caches        (preparatory; after S14-42)
  Check: bar 15's table; no behavior change.

SLICE 2b — Entity records
  S16-11  Outcome records live on the canonical node             (after S14-42)
  S16-12  Source records in the graph cache
  S16-13  Subject records in the graph cache
  S16-14  The Sources section reads Subject records
  S16-15  An Evidence-graph catalog in the graph cache
  Check: bar 11 and 12 (progress, the Sources section).

SLICE 2c — Indexes for the walks
  S16-16  A kinship index kept current on commit
  S16-17  The Family read is one pass over the kinship index
  S16-18  A place hierarchy index
  S16-19  Summary rows carry their life owners and participants
  Check: bar 13, and bar 9 for the Family section.

SLICE 2d — Promote
  S16-20  Promote's profiles and name index survive writes
  S16-21  Promote's place-link checks read the place index
  S16-22  Promote reads each handle's neighbourhood instead of walking the catalog
  Check: bar 9 and 10.

SLICE 2e — The app's cache
  S16-23  Conclusion pages show their last value while they refresh
  S16-28  Cache keys as plain values                              (preparatory)
  S16-24  Person, Event and Place reads report their inputs
  S16-25  Swift keys follow the inputs their values report
  S16-26  Conclusion lists merge the rows a write touched
  S16-27  One primitive for values derived from cached data
  S16-29  Prefetch a person's other sections and a hovered person
  Check: bar 12 (lists), 14 and 15.

S16-99 — Dogfood close / docs
```

S16-01, S16-07, S16-09 and S16-03 can start together; S16-08 waits for Spike 14's S14-40. Shared files, so whichever lands second rebases:
- S16-01 and S16-02 both change `api/proto/engine.proto` and its `GoStore` mapping (and regenerate the protobuf).
- S16-03, S16-04 and S16-06 all edit `ConclusionDetailPage.swift`; S16-03 and S16-07 both edit `PersonSourcesSection.swift` and `PersonTimelineSection.swift`.
- S16-04 and S16-05 both change the Sources diagram column (`PersonSourcesSection` `diagram`, `PersonRecordLinksColumn`): S16-04 lands first, and S16-05 places its note beside S16-04's notice.

Batch 2 waits for Spike 14's S14-42. After S16-10, slices 2b, 2c and 2d run as three streams (records, walks, Promote) that meet at S16-22; slice 2e's Swift PRs can start as soon as their Batch 1 prerequisites land. Shared files: S16-11, S16-13 and S16-14 all edit `core/database/conclusiondetails/conclusiondetails.go` or `persondetail/claims.go` in that order; S16-18 and S16-19 both edit `conclusionsummaries/compose.go`, and S16-19 also `persondetail/timeline.go`, after S16-02; S16-12 and S16-25 both change the claims and record links descriptors in `CatalogQueryRegistry.swift`, S16-12 first; S16-28 lands before S16-25 so S16-25 writes its selector once.

### Dependencies at a glance

```text
S14-40 ──▶ S16-08
S16-01 ──▶ S16-05
S16-07 ──▶ S16-02
S16-03 ──▶ S16-04
S16-03 ──▶ S16-06
S16-D1 ──▶ S16-04
S16-D1 ──▶ S16-05
S16-09

S14-42 ──▶ S16-10 ──▶ S16-12 ──▶ S16-13 ──▶ S16-14
S14-42 ──▶ S16-11 ──▶ S16-13
S16-01 ──▶ S16-14
S16-10 ──▶ S16-13
S16-10 ──▶ S16-15
S16-10 ──▶ S16-16 ──▶ S16-17 ──▶ S16-24 ──▶ S16-25 ──▶ S16-29
S16-01 ──▶ S16-16
S16-10 ──▶ S16-18 ──▶ S16-19
S16-08 ──▶ S16-18
S14-42 ──▶ S16-18
S16-02 ──▶ S16-19
S16-18 ──▶ S16-21 ──▶ S16-22
S16-10 ──▶ S16-20 ──▶ S16-22
S16-16 ──▶ S16-22
S16-04 ──▶ S16-23 ──▶ S16-29
S14-42 ──▶ S16-24
S16-28 ──▶ S16-25
S16-03 ──▶ S16-25
S16-09 ──▶ S16-25
S14-42 ──▶ S16-26
S16-03 ──▶ S16-27
S16-07 ──▶ S16-27
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Each names its audit items; the evidence, the failing tests and the line numbers are in [the audit](person-page-audit.md).

### Slice 1a — Bounded reads, folds and the Timeline

#### S16-01 — Cap the Person page reads against crafted projects

**Ships.** The Person-read caps of rule "Crafted fan-out is capped, not engineered", each checked before the work it bounds, with the audit's seed test as a timing test at least 20 times past the cap: one capped kin per person in `Families`, used by both of its loops, and no implied siblings through a person past `MaxPivotEnd` parents (P-3); a person's records per graph in `memberGraphs` (P-4). `recordlinks.Pairs` reports whether it counted; `persondetail.RecordLinks` sends at most `maxRecordLinkLines` lines, strong first, and the response says when it did not count or dropped lines; `GoStore` decodes it, and Swift's `maxArcs` and strong-first pass go (P-10). Two small items ride along in the same files: every `persondetail` read checks the Person through `personNode`, `Claims` and `Families` included (P-21); `RecordLinksGraph` gains `source_ref` and loses the unread `source_id` and `source_type_id`, so the Sources section no longer hands `graphSourceRefs` to `PersonRecordLinksContent`, and the "name, else ref" and "title, else ref" fallbacks are written once (P-22, part).
**Not this PR.** Faster algorithms (refused); an event's places (S16-08); the note in the diagram (S16-05); the record links memo ([idea](../../ideas/person-page-performance.md)); the claims read owning every graph label (the rest of P-22, [`structural-debt.md`](../../ideas/structural-debt.md)); `edgeKey` (P-24, S16-16, which rewrites `canonlinks`' loading).
**Check.** Bar 2 for the family, sibling and claims axes; a handler test where a read past `MaxReadWork` says it did not count; the real-data tests (`TestPersonFamilyQueryCount`, the record links budgets at 50 graphs) unchanged. The `persondetail` table tests gain a non-Person handle for every read.
**Context.** Audit P-3, P-4, P-10, P-21, P-22; rules "Crafted fan-out is capped, not engineered", "The record links read is bounded and says so"; `core/database/persondetail/family.go`, `claims.go`, `recordlinks.go`, `core/kinship/siblings.go`, `core/recordlinks/recordlinks.go`, `api/ffi/handlers/person_detail.go`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `PersonSourcesContent.swift`, `PersonSourcesSection.swift`; [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md) for the field. Lines this makes false, updated here: [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) line 33 (what keeps `Siblings` linear); the comment on `kinship.Siblings` ("no MaxPivotEnd bound"); [`matching-model.md`](../../matching-model.md) rows for `MaxPivotEnd` and `MaxEntityMembers` (their new page readers) and its record links paragraph (the read now says when it did not count); the `PersonFamilyContent.swift` comment "Go still returns every member".

#### S16-08 — Cap the places an event folds

**Ships.** One helper reads an event's places, drops duplicate ids and keeps at most 20 in link order, and the four reads in `conclusionsummaries/compose.go` (`eventPlaceGraph`, `lifeOwners`, `composeEvents`, `eventPlaces`) use it, so `FoldLocations` and the place graph never see more (P-2). A timing test folds an event with 10,000 places under a second.
**Not this PR.** A linear `FoldLocations` (refused); the Person reads (S16-01).
**Check.** Bar 2 for the places axis, including a project whose search projection is rebuilt on open; the summary tests unchanged.
**Context.** Audit P-2; rule "Crafted fan-out is capped, not engineered"; `core/database/conclusionsummaries/compose.go`, `placechains.go`. **Depends on S14-40** (Spike 14, Planned), which adds input scopes to these reads and routes them through one scope `follow`; its map names this PR.

#### S16-07 — One paged fold that moves focus

**Ships.** `ConclusionFold` gains a page of 50 and one fold state type that Family, Timeline and Sources use in place of their `Set`s and Sources' `rowsExpanded`; `firstRevealed` names the first item of the page just revealed and `ConclusionShowMoreRow` claims focus for it through `FocusClaim` (Family gains it; the `Task` hops in the Timeline and Sources go) (P-14, P-7's pages). One focus keeper in `macos/App/Platform/` holds focus across a width change for Gallery and Timeline. When the record links diagram highlights a graph whose record line is folded, that Source's Show more takes the highlight (P-15; the content already maps folded claims to graphs in `claimGraphs`).
**Not this PR.** Building only shown rows and the lifeline's click targets (S16-02); Back restoring a section's folds (P-19, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 3's paging and bar 7's fold focus: content tests name the revealed item after each page for every section, Family included; a test unfolds 2,000 hidden items in pages of 50. A content test highlights a folded line's Show more.
**Context.** Audit P-7, P-14, P-15; rule "One paged fold that moves focus"; `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `PersonTimelineSection.swift`, `PersonTimelineContent.swift`, `PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `ConclusionGallerySection.swift`, `macos/App/Platform/FocusClaim.swift`; [`macos-client-patterns.md`](../../macos-client-patterns.md) "One sectioned detail page" (teaches the fold) and Wide rule 5 (cites the keeper).

#### S16-02 — The Timeline places a missing death and builds what it shows

**Ships.** A life marker's position becomes one value, beside or before a row, on the wire (`oneof`), in `CatalogLifeMarker` and in Go's `LifeMarker`; `lifeMarkers` places a placeholder death just before the earliest dated after-death row whenever one is dated (the mirror for birth), `PersonTimelineContent` inserts the marker there and `PersonTimelineLifeline` puts its stub at that row's start (P-1; user, 2026-11-08). `PersonTimelineContent` builds rows and person cards only for what S16-07's folds show, and the lifeline makes click targets only for shown rows (P-7). The lifeline keeps a chosen scale only while its year range is unchanged as well as its ids, so a date edit from another window rescales it (P-20).
**Not this PR.** Memoising the content, participants once per read, lazy cards ([idea](../../ideas/person-page-performance.md)).
**Check.** Bar 1 and 3; P-1's failing test and Go `TestLifeMarkers` cases pass ("only a burial dated", and "a baptism and a burial": the death sits before the burial, and its stub at the burial's start, not the baptism's end); `PersonTimelineLifelineTests` asserts click targets only for shown rows; a content test builds 2,000 events × 500 participants and touches only shown rows and cards. A placement test moves one event's year under the same ids and gets a new scale.
**Context.** Audit P-1, P-7, P-20; rules "A marker's position is one value", "The Timeline builds what it shows"; Spike 13 rules "The timeline is a lifeline" and "Timeline lists fold", [R-24a](../spike-13/person-detail.md#3-timeline); `core/database/persondetail/lifeline.go`, `timeline.go`, `api/proto/engine.proto`, `api/ffi/handlers/person_detail.go`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `macos/App/Features/Conclusions/PersonTimelineContent.swift`, `PersonTimelineLifeline.swift`, `PersonTimelineSection.swift`.

### Slice 1b — Current data and honest notices

#### S16-09 — Conclusion pages follow the Sources their records cite

**Ships.** The `.conclusionDetail` descriptor follows `.source` for the Sources its field records, derived-row owners and pictures cite (P-6). The unused `CatalogQueryStalePolicy` goes while the cache is open here (parked from the S14-11 plan: no descriptor sets `.refetchOnRevisit`): the enum (`CatalogQueryStalePolicy.swift`), the `stalePolicy:` argument of every descriptor arm, `CatalogQueryRegistry.stalePolicy(for:)` and the five test assertions that read it.
**Not this PR.** Event and Place pages' other keys; a narrower Source effect for renames only (the audit scope is what the cache can follow).
**Check.** Bar 4; P-6's `WorkspaceSessionTests` case passes and `personClaimsFollowHandlesAndGraphs`' "the detail page follows only the handle" assertion is updated; manual: a Source rename, then Back to a Person whose birth cites it. `rg CatalogQueryStalePolicy macos` finds nothing.
**Context.** Audit P-6; rule "A cached detail follows the Sources its records cite" (its fan-out); `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `Session/CatalogQueryStalePolicy.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogQueryRegistryTests.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) (slot list), [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys".

#### S16-03 — One open section and one list of section reads

**Ships.** `WorkspaceLocation.openPageSection`; one read list per `ConclusionPageSection` replacing its three slot functions, read by `PlaceRegistry`, the page's refreshing state, the Person fill and both section hosts; the host's identifier root from the page identifier; [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) and [`macos-client-patterns.md`](../../macos-client-patterns.md) updated (P-13). No visible change. First, `ConclusionDetailPage.swift` (15 types in 736 lines) is split by moving types beside their kin: the gallery model (`ConclusionDetailPicture`, `ConclusionDetailGallery`) to `ConclusionGallerySection.swift`, `ConclusionDetailRowsSection` beside `ReconciledValueRow`, and the Place relationship model and section beside `PlaceRelationshipRow`, so S16-04 and S16-06 edit a smaller file (P-26, part).
**Not this PR.** Event and Place sections; a per-case spec table for label and glyph (the next section is one list entry either way).
**Check.** Unit tests: a first section that reads its own key is in the place's query keys and stays visible across a write; the page, the place and the fill agree on the open section for every location; the Person page's identifiers are unchanged.
**Context.** Audit P-13, P-26; rule "One open section, one list of reads"; `macos/App/Features/Workspace/WorkspaceLocation.swift`, `Session/PlaceRegistry.swift`, `WorkspaceSession.swift` (visible keys), `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `PersonDetailView.swift`, `PersonSourcesSection.swift`; `PlaceRegistryTests`, `WorkspaceNavigationTests`.

#### S16-04 — A failed reload keeps the rows and says so

**Ships.** `QueryHandle`'s reading with a refresh failure beside the value and a failure revision; `ConclusionDetailPresentation` built from it; the page and each section host draw `PVRetryCallout` above stale content (compact in the diagram column) and on a first-load failure, as S16-D1 draws them; Try again re-queries the open section's reads (S16-03); the skill and the doc teach the reading (P-5). A Person page whose person was deleted or merged elsewhere says "This person no longer exists" in its leave Toast and its section notices; the kind supplies the line, so Event and Place keep the shared one (P-17).
**Not this PR.** `QueryHandle.read` for its other callers; the Conclusion lists ([idea](../../ideas/person-page-followups.md)); `EvidenceGraphModel`'s and `PromoteModel`'s attempt counters; Back restoring a section's place (P-19).
**Check.** Bar 5, and bar 8's not-found line; P-5's FakeStore test passes; a test that a second failure bumps the failure revision.
**Context.** Audit P-5, P-17; rules "A refresh failure is read beside its value", "One open section, one list of reads", "A person's records are records"; brief S16-D1; S14-20 for the retry pattern; `macos/App/Features/Workspace/Session/QueryHandle.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) "Stale-while-revalidate", [`macos-client-patterns.md`](../../macos-client-patterns.md) (the `QueryHandle` row).

#### S16-05 — The record links diagram says when it is partial

**Ships.** `PersonRecordLinksContent` reads S16-01's field and its own 1,000-node cut into one partial state, and `PersonRecordLinksColumn` shows S16-D1's note, without uncounted arcs drawn as weak (P-10).
**Not this PR.** Any data about a line (Spike 13 rule "The view is a circular link diagram"); hover on a folded record line (P-15, S16-07).
**Check.** Bar 6; a content test for each of "did not count", "lines dropped" and "graphs dropped".
**Context.** Audit P-10; rule "The record links read is bounded and says so"; brief S16-D1; lands after S16-04 (same column); `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `PersonRecordLinksColumn.swift`.

### Slice 1c — Keyboard and VoiceOver

#### S16-06 — Section switches and removals are spoken

**Ships.** `PVSegmentedControl`'s `focus:` handle lands on the selected segment rather than the first enabled one (the component and its doc comment); the switcher passes a handle, a section switch announces the section and claims the selected segment when the old focus is gone (P-11). The leave-or-stay decision after a Remove moves into a function on `UnpromoteFlow` or a static helper that returns leave, or stay with the announcement, and the Sources section posts it; both branches are unit-tested (P-12, P-16). The wording of rule "A person's records are records": the header's count reads records on every Conclusion page, and the Unpromote confirm ends "…, and the person stays", per kind (P-17).
**Not this PR.** Fold focus (S16-07); Back restoring scroll, folds and focus (P-19, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 7, and bar 8's count and confirm; unit tests for both leave-or-stay branches, replacing `UnpromoteFlowTests.confirmReportsTheEffects`' inline copy; manual VoiceOver steps for ⌘3 from a Family tile.
**Context.** Audit P-11, P-12, P-16, P-17; rules "Every outcome on the page is spoken", "A person's records are records", "One open section, one list of reads"; `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `PersonSourcesSection.swift`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift`; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).

### Batch 2 — Cache architecture

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). The design, the evidence and the estimates are in [`cache-architecture.md`](cache-architecture.md); each PR that claims speed measures it first (rule "Measure growth").

### Slice 2a — One home for caches

#### S16-10 — Graph indexes: one home for catalog-wide caches

**Ships.** `graphcache` registers graph indexes beside node memos (Build lazily on any graph, Apply in `OnCommit` with the commit's `effects.Set` on the session graph, Inputs, Check in Verify, DropOnVocabulary, Cap), and the match index, alias index and name format move onto it with today's behavior (the match index still drops on a handle commit until S16-20). `canonicalgraph` goes: its last production caller left with S14-42, and the inputs test keeps its oracle query inline. `docs/catalog-caches.md` (new in S16-10) lists every Go and Swift cache with its key, what builds it, the signal that updates it, its cap and its size; the maintainability lens gains the Known trap "a cache that empties on any write".
**Not this PR.** Any new cache (S16-11 on).
**Check.** Bar 15's table; Verify passes with the moved caches checked; Promote's and the lists' budgets unchanged.
**Context.** [Design § One home](cache-architecture.md#one-home-for-a-catalog-wide-cache-s16-10); rules "One registry for catalog-wide caches", "Every cache names the signal that updates it"; `core/database/graphcache/graph.go`, `commit.go`, `memo.go`, `verify.go`, `core/database/namealiases`, `core/database/nameformats`, `core/database/promotealign/index.go`, `core/catalogsession/session.go`, `core/database/canonicalgraph`; [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §8 (also its stale "S9-28 caches it as canonical edges", §7); [`.agents/skills/review-pr/lenses/maintainability.md`](../../../.agents/skills/review-pr/lenses/maintainability.md). **Depends on S14-42** (Spike 14), which last changed `commit.go`; its map names this batch.

### Slice 2b — Entity records

#### S16-11 — Outcome records live on the canonical node

**Ships.** `Node` holds each Property's outcome rows (Observation, Subject, reason, value rank, denied by, vote) and the chosen likeness, loaded in one more batched query beside `Values` and reloaded with the node from `Set.Handles`. The detail read takes them from the node: `sqlPropertyIDs` and the reconciler half of `sqlOutcomes` go, and the record shells (Citation, Artifact, Source) stay one SQL read keyed by Observation until S16-13. `gallery.ForEntities` goes: the detail page's pictures and the summaries' lead picture compose from the node. Verify's node truth check covers the outcomes.
**Not this PR.** The record shells (S16-12, S16-13); any wire change.
**Check.** Bar 11 in part: a warm Person open runs no reconciler or gallery SQL (query counter); the detail and summary tests unchanged; nodes at 16,000 persons grow by under 6 MB.
**Context.** [Design § Entity records](cache-architecture.md#entity-records); rule "Cache entities and indexes; compose answers"; `core/database/graphcache/graph.go`, `load.go`, `verify.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/gallery`, `core/database/conclusionsummaries/compose.go` (`leadPictures`); [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §2.7, [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §10. Check first whether `value_image_file_id` is already limited to readable images. **Depends on S14-42** (same `compose.go`).

#### S16-12 — Source records in the graph cache

**Ships.** A registered index of Source records (ref, title, type, cover, credibility grade, Artifacts with label and file), built at open in two queries and updated from `Set.SourceScopes`; credibility grades join the vocabulary. `EvidenceGraph` stops copying Source words and `loadGraphHead`'s join goes; the claims grouping, the record links labels and the outcome record shells take Source words from the record. The `sources` effect stops naming a Source's graphs for a title or type change, and the app's `personClaims` and `personRecordLinks` keys follow `.source` for the Sources in their value instead.
**Not this PR.** Subject records (S16-13); the Evidence-graph catalog (S16-15).
**Check.** Bar 11's rename: a Source rename reloads no loaded Evidence graph and drops no graph memo (graph cache test), and the Sources section, record links and provenance sheet show the new title; the app's claims key reloads on the rename (Swift test).
**Context.** Rule "A Source's words live once"; `core/database/graphcache/evidence_graph.go`, `core/database/graphfill/load.go`, `core/database/effects/registry.go` (`sources`), `core/database/autoreconciler/autoreconciler.go` (`ProvenanceJoins`), `core/database/persondetail/claims.go`, `recordlinks.go`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`; `effects` completeness test; `docs/catalog-caches.md` (S16-10).

#### S16-13 — Subject records in the graph cache

**Ships.** A registered index of Subject records (ref, label, type, graph, Source, accepted claim, and the Observations with each Citation's label, page and Artifact and the uncertainty flag), built lazily from a loaded graph or one batched query, dropped for every Subject of a graph in `Set.EvidenceGraphs`, under its own LRU cap. Provenance records compose from the node's outcome rows (S16-11), the Subject record and the Source record (S16-12): `sqlOutcomes`, `toponymsOf` and the date and name lookups of the detail read go, and so does the second pass of `PersonDerivedRows`.
**Not this PR.** The Sources section (S16-14); record links, which still load whole graphs.
**Check.** Bar 11: a warm Person open on Vitals with its provenance sheet issues no SQL; the provenance and detail tests unchanged; Subject records for 50,000 Observations measured against the cap.
**Context.** [Design § Entity records](cache-architecture.md#entity-records); rules "A Source's words live once", "Memory is budgeted"; `core/database/graphfill`, `core/database/graphcache`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusionsummaries/compose.go` (`PersonLifeOwners` callers); [`ideas/provenance.md`](../../ideas/provenance.md) (records loaded on open of the sheet: this PR moves them to memory instead).

#### S16-14 — The Sources section reads Subject records

**Ships.** The claims read composes each claim from Subject and Source records: `memberGraphs` loads no whole Evidence graph and does not rescan a graph per record, the handlers build each graph's label map once, and S16-01's records-per-graph cap stays at the record read. The record links read takes its claims from the same composition.
**Not this PR.** Record links' graph loads (they compare whole graphs); a batch graph read ([`structural-debt.md`](../../ideas/structural-debt.md)).
**Check.** Bar 12: the Sources section of a person in 80 Evidence graphs loads no whole graph (graph cache counter); the claims tests and S16-01's timing test unchanged.
**Context.** Audit P-4; rule "Batch 1's caps stay"; `core/database/persondetail/claims.go`, `recordlinks.go`, `api/ffi/handlers/person_detail.go`, `source_read.go`.

#### S16-15 — An Evidence-graph catalog in the graph cache

**Ships.** A registered index of graph shells (Source, ref, name, subject, filed and Observation counts) and each Source's graphs in list order, built at open or first read and updated from `Set.Reads` (`evidence_graph_progress`, `evidence_graphs`). `graphprogress` reads it instead of its full-table `GROUP BY`s; `evidencegraphs.List` and `SourceOf` on read paths read it; the claims and record links reads group and order graphs from it.
**Not this PR.** A batch graph read; the Swift `EvidenceGraphIndex`.
**Check.** Bar 12: a warm Sources list issues no progress `GROUP BY` (query counter); the progress and list tests unchanged.
**Context.** [Design § Entity records](cache-architecture.md#entity-records); `core/database/graphprogress/graphprogress.go`, `core/database/evidencegraphs`, `core/database/effects/registry.go` (`Reads`), `core/database/persondetail`.

### Slice 2c — Indexes for the walks

#### S16-16 — A kinship index kept current on commit

**Ships.** A registered kinship index: the per-pivot results the `canonlinks.kinship` slot holds today (with its one-hop `Deps`), a reverse map from each person to the implied links naming them, and each person's recorded relationship links. Apply recomputes the pivots of the commit's roots and their neighbours; a vocabulary commit rebuilds it; `WarmOnOpen` builds it after `promotealign.Warm`. `Reader.For` and `Relationships` keep their signatures and read it; siblings are the parents' children, read from it. The `for` and `siblings` slots, `DropOnHandleCommit` and `dropHandleMemos` go. `edgeKey` is made unambiguous, or goes with the merge it keyed (P-24). S16-01's sibling cap stays.
**Not this PR.** Rewriting the Family read (S16-17) or Promote's canon (S16-22).
**Check.** Bar 9: after an Observation edit on a filed person, `For` answers from the index with no recompute beyond the commit's pivots (counter); Verify passes; the kinship, Family and Promote tests and goldens unchanged; the index measured at 16,000 persons.
**Context.** [Design § Indexes](cache-architecture.md#indexes-for-the-walks); rules "Every cache names the signal that updates it", "Orders are defined and results do not move", "Warm at open, update on commit"; `core/database/canonlinks/reader.go`, `core/kinship`, `core/database/graphcache/memo.go`, `commit.go`, `core/catalogsession/session.go`; delete the [`structural-debt.md`](../../ideas/structural-debt.md) row "The kinship memo slots … drop on every handle commit"; update [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) line 33 and [`matching-model.md`](../../matching-model.md) where they describe `canonlinks`.

#### S16-17 — The Family read is one pass over the kinship index

**Ships.** `persondetail.Families` reads each person's kin from the index once, with S16-01's relatives-per-role cap at that read, instead of loading relationships, re-reading them and filtering in two loops; the `canonlinks` loaders only it used (`LoadRelationships`, `LoadSiblings` and their helpers) go.
**Not this PR.** Family tree reads beyond the nuclear family ([`kinship-view.md`](../../ideas/kinship-view.md)).
**Check.** Bar 13's Family half: the Family tests, `TestPersonFamilyQueryCount` and S16-01's wide-family timing test unchanged or better.
**Context.** Rule "Batch 1's caps stay"; `core/database/persondetail/family.go`, `core/database/canonlinks/reader.go`.

#### S16-18 — A place hierarchy index

**Ships.** A registered index of every place's shell and its `part_of` and `succeeded_by` links with decoded dates, both directions, and each place's recorded `InputKey`s. It builds lazily on any graph (one batched association load) and, on the session graph, Apply rebuilds a place only when its shell or place-relationship links move. Composers read chains from it and add each place's keys to the row's scope, with S14-40's input cap and depth cut; date-dependent chains stay functions over it. `buildPlaceGraph`, `placeStep`, `rememberShell`, `rememberPlaceLink`, `eventPlaceGraph` and `placeNeighborhood` collapse onto it, and the Places list stops rebuilding chains on each read. S16-08's places-per-event cap stays.
**Not this PR.** Promote's place-link checks (S16-21).
**Check.** Bar 13's city: reading a city's chain loads none of its location links (counter); S14-40's inputs tests and the rebuild-equals-upkeep tests pass; the first `part_of` filed from a city with 6,000 events measured against S14-40's figure; the per-write build measured.
**Context.** [Design § The write path](cache-architecture.md#the-write-path-composes-on-a-one-shot-graph) and [§ Indexes hand their inputs](cache-architecture.md#indexes-hand-their-inputs-to-the-row); rules "Built lazily on any graph, kept only on the session graph", "An index a composer reads hands its inputs to the row"; `core/database/conclusionsummaries/compose.go`, `placechains.go`, `core/database/searchindex/handles.go`. **Depends on S14-42** (Spike 14) and S16-08 (same reads).

#### S16-19 — Summary rows carry their life owners and participants

**Ships.** A person's summary row keeps its birth and death owner events and their event–place links; an event's row keeps the Property its date came from and its participants by role, from one `Follow` of the participants hop. `PersonLifeOwners` and `lifeOwners` go (the detail's derived rows read the row), the Timeline reads participants from the event rows it already loads instead of walking each event, and `EventSummary.subjects` and `EventParticipant` become one shape (A-33).
**Not this PR.** Sending each participant once on the wire (P-8, [idea](../../ideas/person-page-performance.md)).
**Check.** Bar 13: the summary, derived-row and Timeline tests unchanged; S14-40's inputs tests pass; the Timeline read issues no per-event participant walk (counter).
**Context.** [Design § Indexes](cache-architecture.md#indexes-for-the-walks); rule "An index a composer reads hands its inputs to the row"; `core/database/conclusionsummaries/compose.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/persondetail/timeline.go`, `participants.go`; delete [`structural-debt.md`](../../ideas/structural-debt.md) A-33. Lands after S16-02 (same `timeline.go`).

### Slice 2d — Promote

#### S16-20 — Promote's profiles and name index survive writes

**Ships.** A handle's profile is a node memo whose `Deps` are the handle and its events. The shortlist index moves onto the registry and Apply patches it from the commit's handles (old keys out, new keys in, generation kept for the rest); the candidates memo re-ranks only Subjects whose lookup keys meet a changed handle, and survives its graph leaving the LRU. `indexGen`, the generation checks and the profiler's shared map go; `Warm` fills full profiles in the same pass; Verify compares the patched index with a fresh build. `graphalign.Handle` folds onto `subjectcompare.Subject`.
**Not this PR.** The canon walk (S16-22).
**Check.** Bar 10's second half: the first proposal after a handle commit is no slower than 1.5 times a warm one at 4,000 and 16,000 persons; Promote's tests and goldens unchanged; Verify passes.
**Context.** [`promote-ranking.md`](../../ideas/promote-ranking.md) "Name index on commit" (moved here; delete it there) and its open questions (`shortlist` removal, merges through `OnCommit`'s merged list); audit 2 B-8; `core/database/promotealign/profiles.go`, `index.go`, `canon.go`, `core/shortlist`, `core/graphalign`, `core/database/graphcache/commit.go`; [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §8; the [`structural-debt.md`](../../ideas/structural-debt.md) row on `graphalign.Handle` (its `Handle` half).

#### S16-21 — Promote's place-link checks read the place index

**Ships.** `graphalign.Connections` and Promote's connections read the filed `part_of` and `succeeded_by` links from the place index, so the loop check no longer needs every canon edge loaded.
**Not this PR.** The bridge filer's `filedLinks`, which must read its own transaction.
**Check.** Promote's connection and loop tests unchanged; a proposal's connections read no canon edge for place links.
**Context.** `core/graphalign/connections.go`, `core/database/promotealign/connections.go`; S14-22 D-7 (why the check needs every filed link).

#### S16-22 — Promote reads each handle's neighbourhood instead of walking the catalog

**Ships.** `graphalign` reads a `CanonView` (a handle's profile and links), served by S16-20's profiles and the kinship index; `Align`, `corresponding`, `place` and `segments` read the seeds, the handles a placed pair nominates and their neighbours. `loadCanon`'s walk, its global `LinkSet` and `PlaceImplied` loop, and `newState`'s adjacency build go; `graphalign.CanonEdge` folds onto `kinship.Link`. Tests keep `Canon{Handles, Edges}` through an adapter.
**Not this PR.** Ranking on Evidence graph scoring ([`promote-ranking.md`](../../ideas/promote-ranking.md)).
**Check.** Bar 10: a warm proposal at 16,000 persons at least three times faster than today's on the same machine, and under twofold growth from 4,000; an equivalence test runs today's walk and the view on the large fixture and gets the same proposal; the goldens and `TestAlignInvariants` unchanged.
**Context.** [`promote-ranking.md`](../../ideas/promote-ranking.md) "Narrower canon walk" (moved here; delete it there); audit 2 B-8; rule "Orders are defined and results do not move"; `core/database/promotealign/canon.go`, `core/graphalign/align.go`, `perf_test.go`; [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §7.

### Slice 2e — The app's cache

#### S16-23 — Conclusion pages show their last value while they refresh

**Ships.** The per-entity Conclusion keys (detail, claims, family, events, record links) go from `.evict` to `.markStale`; `apply` evicts the keys whose entity is in `effects.deleted`; a not-found reload over a stale value reads as gone (S16-04's reading); hidden per-entity keys are capped at 50 entities, least recently shown first.
**Not this PR.** Narrower refreshes (S16-25).
**Check.** Bar 14: Back to a visited Person shows its last value with "Refreshing"; a person deleted in another window leaves; session tests for each kind.
**Context.** Rule "Swift caches one value per view…"; `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `QueryHandle.swift`; [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys" (the S9-15 exception goes), [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md).

#### S16-28 — Cache keys as plain values

**Ships.** `CatalogQueryKey` becomes a value of kind, project and ids, with static factories so call sites keep their spelling; the `Kind`, `kind` and `project` switches, each descriptor's `guard case … mismatchedCatalogKey` and the id-closure `guard case`s go. No behavior change.
**Not this PR.** Changing any key's refreshes (S16-25).
**Check.** Registry and session tests unchanged.
**Context.** `macos/App/Features/Workspace/Session/CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) (how a key is declared).

#### S16-24 — Person, Event and Place reads report their inputs

**Ships.** The detail, claims, family, timeline and record links reads return `inputs` (handles, graphs, Sources) with their value: handles recorded through an `InputScope` or the kinship index's entries, graphs and Sources listed by the read. `GoStore` decodes them; nothing reads them yet.
**Not this PR.** Following them in Swift (S16-25).
**Check.** Handler tests: each read's inputs include every relative, co-participant, graph and cited Source its value shows; an input key kind for a graph and a Source added beside `InputKey`'s handle kind.
**Context.** [Design § Swift](cache-architecture.md#swift); `core/database/persondetail`, `core/database/conclusiondetails`, `api/ffi/handlers`, `api/proto/engine.proto`, `graphcache` `InputKey` (S14-40's `inputs.go`); [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md). **Depends on S14-42** and S16-17.

#### S16-25 — Swift keys follow the inputs their values report

**Ships.** One `.followsInputs` refresh (handles, graphs, Sources, plus the key's own id) replaces `personSection`, `personGraphsSection`, `applyHiddenPolicy(followsValue:)`'s per-kind id closures and S16-09's Source follow on the Conclusion keys.
**Not this PR.** Lists (S16-26).
**Check.** Bar 14: a write that names nothing a cached section read leaves it untouched; one session test per relation (parent, implied sibling, partner, child, co-participant, cited Source) reloads it.
**Context.** Rule "Swift caches one value per view…"; `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `CatalogLastUpdated.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md), [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys".

#### S16-26 — Conclusion lists merge the rows a write touched

**Ships.** `ListPersonSummaries`, `ListEventSummaries` and `ListPlaceSummaries` take an id filter and return each row's sort key; the three lists' refresh becomes the existing `.merged` read (replace, remove deleted or merged, insert by Go's sort key) and they revalidate while hidden.
**Not this PR.** The lists' failure notice ([idea](../../ideas/person-page-followups.md)).
**Check.** Bar 12's lists: a write re-reads only its touched rows (session test); the list order matches a full reload on the large fixture.
**Context.** `core/database/conclusionsummaries`, `api/ffi/handlers`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` (`.merged`, S14-11). **Depends on S14-42**, which makes the touched set exact.

#### S16-27 — One primitive for values derived from cached data

**Ships.** `session.derived(id, inputs:, extra:)` builds a value once per change of its inputs' revisions (and name orders, fold state, width class). `EvidenceGraphIndex`'s revision bookkeeping, `EvidenceGraphModel`'s tuple memo, `CitationComposerContext`'s `MemoToken`, the uncached `evidenceGraphSnapshot` builds (Promote, Unpromote) and `ConclusionPageSectionHost`'s per-pass `make` move onto it.
**Not this PR.** Engine-side memos.
**Check.** Bar 15's last clause: a test counts one build per input revision through an `apply`; the moved features' tests unchanged.
**Context.** [`person-page-performance.md`](../../ideas/person-page-performance.md) "Section content is rebuilt on every pass of its host" (moved here; delete it there); `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/CitationComposer/`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift`, `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift`.

#### S16-29 — Prefetch a person's other sections and a hovered person

**Ships.** `PlaceRegistry` gains prefetch keys (Family and Timeline; not record links or a cold claims read), warmed after the open section's keys settle, one at a time; a 250 ms hover on a person tile or card warms that person's detail. Each prefetch is measured against the next foreground read it could delay.
**Not this PR.** Engine-side warming (rule "Warm at open, update on commit").
**Check.** Bar 14's last clause: switching to Family or Timeline after the page settles shows no spinner; the prefetch's cost on the next click measured.
**Context.** `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Features/Conclusions/PersonCardLink.swift`; rule "Swift caches one value per view…".
