# Spike 9 — Canonical entities

**Done.** Promote from the Evidence graph into Persons, Events, and Places, with list and detail pages and omnibar search, all reading one auto-reconciler cache. Live models: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md). Timings and the reserve tiers: [`performance-ledger.md`](performance-ledger.md).

## Decisions

- **Identity Claims are the only Conclusion truth writes.** A handle's members are accepted claims. Promote writes `accepted` only. No Reconciliation Claims, no merge, no notes.
- **Promote is one page and one Done.** Align the Evidence graph with the canonical graph; the researcher fixes the rows that need judgment; one transaction files claims, one-hop pins, and bridges. Bridges have no identity of their own: participation and location by their ends, relationships and place relationships by ends plus type. A self-link or a place cycle stays unfiled and does not fail the claim.
- **One auto-reconciler cache.** Every value type runs one pipeline. Order and outcome labels are derived, and a full rebuild matches incremental upkeep. Association ends are rows in that cache. Screens compose from it. No second derived tier.
- **Places link the hard way.** Separate Places, locked `part_of` and `succeeded_by`, a period on the Place and a span on the link. A birth place reads the chain at the birth date. Locations that share a containment graph fold into one chain; unrelated Places stay competing values. No place-nature vocabulary.
- **Search text comes from the list headers.** Person, Event, and Place are omnibar hits. The text match is 90% of the score; the other 10% is kind priority (persons, events, places, sources, then everything else). The section you are in only nudges. Subjects are not hits.

## Refused

Reconciliation Claims and `name_format`. Editing or removing an Identity Claim, and the weak-claim review, after Promote. Canonical merge. Stub handles with no Subject. Association-kind pages. Likeness thumbnails (the slot is a placeholder). Tree, timeline, and map. Gazetteer lookup. A second derived cache. Deep-fixture timings.

## Leftovers

| Item | Home |
| --- | --- |
| Remove a member, edit a claim, weak-claim review | [`identity-claim-review.md`](../../../ideas/identity-claim-review.md) |
| Gazetteer writes the same place evidence | [`place-gazetteer-service.md`](../../../ideas/place-gazetteer-service.md) |
| Name display styles | [`structured-name-model.md`](../../../structured-name-model.md) §4.5 |
| Deep-fixture timings | [`performance-ledger.md`](performance-ledger.md) — not taken; reserve tiers unused |
