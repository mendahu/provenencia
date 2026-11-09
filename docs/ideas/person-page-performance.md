# Person page performance

**Status:** idea only — not roadmapped. Speed work on Spike 13's Person page found by the [2026-11-08 audit](../deployment-plan/spike-16/person-page-audit.md) and descoped by the user the same day: real data has not shown the cost yet. Spike 16 caps the crafted cases instead (working rule "Crafted fan-out is capped, not engineered") and gives the Timeline a render cap (S16-02). Every figure below is an estimate from reading the code; **measure before pulling any of these in.**

Related: [`structural-debt.md`](structural-debt.md) row "S13-16 PRF-2" (a person past 64 Evidence graphs cycles the graph cache on the claims and record links reads); [`spike-15/provenance-loading.md`](../deployment-plan/spike-15/provenance-loading.md) (now Spike 15, S15-27 – S15-29) (records drawn in a provenance table).

## The record links read recomputes every pair (P-9)

`persondetail.RecordLinks` (`core/database/persondetail/recordlinks.go`) runs `recordlinks.Pairs` over every graph on every load, with no memo, while holding the catalog session. At 50 graphs that is 1,225 pairs. Any write naming the person or one of their graphs evicts the key, so the next Sources visit recomputes all of it, although one changed graph can move only its 49 pairs.

The budget was met on the lightest per-pair fixture: `recordlinks_perf_test.go`'s households have four people with a name and parent links, 44 work units a pair (27.9 ms warm, 37.9 ms cold, S13-27 ship log). Estimated by `pairWork`'s rule:

| Household (5 properties a person) | Work a pair | Work a read | Share of `MaxReadWork` |
| --- | --- | --- | --- |
| The fixture: 4 people, names only | 44 | 53,900 | 5% |
| 8 people, 7 shared, parent links | about 290 | about 355,000 | 32% |
| 12 people, implied sibling links | 466–698 | 570,000–855,000 | 52–78% |

If time follows work, the richer shapes take about 185–440 ms a read, with every other pane waiting.

**Shape.** First, add a census-sized case to `TestRecordLinksBudgets` (`censusHouseholds(f, 50, size: 8 | 12)`, birth events so parent and sibling links are implied) and log `ReadWork` and the times. If it is slow, keep each pair's lines on the Evidence graph memo (`graphcache.SetEvidenceGraphMemo`, as `evidencelayer.Load` does), stored under both graphs and keyed by the other graph and the page entity, so a commit to one graph drops only its pairs and `MaxReadWork` charges only what is recomputed. A write that names no graph (a portrait choice) need not refresh the key at all.

## Timeline participants are sent once per event (P-8)

`api/ffi/handlers/person_detail.go` (281–287) copies a participant's whole `PersonSummary` into every `EventParticipant`, although Go composes each one once (`timeline.go`). A household that repeats across 200 census events is about 50 people sent 2,000 times, then formatted again per copy on the main thread. A crafted name of 256 KB on 2,000 events is 512 MB on the wire.

**Shape.** Send each distinct person once on `ListPersonEventsResponse` and give `EventParticipant` a person id and roles; Swift builds one card content per person. Pull it in with the next Timeline wire change.

## An unfolded Timeline lays out every card at once (P-7)

After an unfold, every shown row and its cards are laid out in one non-lazy pass (`PersonTimelineStack`), each card's `CachedThumbnail` starting its `.task` together. Card rows are a fixed height (`PVTileCardSize.compact.height`), so a row's card block could be sized without laying it out and its cards mounted only near the viewport.

## The family response sends each sibling two or three times (P-23)

Siblings ship as the parents units' children, the siblings unit, and `origin_siblings`, each with a full `PersonSummary` (`api/proto/engine.proto` 1549–1553, 1680–1692; `core/database/persondetail/family.go` 173–241). The app reads only the parents units' partners and `originSiblings` (`PersonFamilyContent.swift` 224–262). Units are the family tree's shape (S13-12 D-3, [`kinship-view.md`](kinship-view.md)), so the tree may read them later. **Shape:** send `origin_siblings` as indexes into one relatives table, or leave the unit children off the page response until a tree reads them.
