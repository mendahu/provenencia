# Spike 9 — Canonical entities (MVP)

## Status

**Open.** Requirements, design track (13 briefs, one per view), and PR sequence in vertical slices (S9-01…S9-35): [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md) — S9-01…S9-12, S9-13a, S9-13, S9-13b, S9-14, S9-15, S9-16, S9-17, S9-18 and S9-34a (and S9-D1, S9-D2, S9-D5, S9-D8, S9-D9, S9-D10) landed.

Spikes 5–8 built the Interpretation layer: Sources → Citations → Observations on Subjects, drawn on an Evidence graph. Nothing yet says *these three census lines are the same James*. Spike 9 opens the **Conclusion layer**: canonical Persons, Events, and Places assembled from Subjects through **Identity Claims**, with read-only pages that project their values.

Four strands:

1. **Promote** — the Identity Claim workflow, started from a subject card on the Evidence graph. Mint a new handle or file onto an existing one, check the supporting Observations, and save that one claim. Then walk to connected Subjects one at a time, or stop there. This is the hard part, and without it every list is empty.
2. **Pages** — sidebar lists for Persons, Events, and Places, and a detail page for each. Mostly display.
3. **Value resolution + foundation** — every Property on a handle is multi-valued. A Go auto-reconciler produces ranked clusters: single, auto-reconciled (names and dates), or top-ranked by provenance with a *mixed* indicator. Its output lives in one derived **auto-reconciler cache**, maintained in each write's transaction, that every screen composes from. A Reconciliation Claim will trump all of it once that ships.
4. **Search** — Persons, Events, and Places in the omnibar: refs, full text on auto-reconciled names and toponyms, hit rows that read like list rows.

> **Resolution is not reconciliation.** Auto-reconciling names and dates and ranking by provenance are display. The cache is derived and rebuildable; no truth row is written. Reconciliation Claims (researcher-committed values) are out of scope for this spike ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §2.2, §7).

> **Identity Claims are the only Conclusion truth writes.** Canonical entity rows, Identity Claims, and their Observation pins — plus the derived cache and search documents they feed. No merge, no Reconciliation, no notes.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Requirements, dogfood bar, design surfaces, scope, gotchas, open questions |
| [**Completed**](completed.md) | Finished steps |
| [**Performance ledger**](performance-ledger.md) | Conclusion read hotspots, adopted cache strategy, reserve tiers, timings |
| [Design briefs](design/) | Claude Design — S9-D1…D13 open |

## Relationship to earlier spikes

Spike 7 made subject types product-seeded and gave each a `ref_prefix` (`PER`, `EVT`, `PLC`, …) next to its `candidate_ref_prefix`. Canonical refs already have a namespace. Spike 8 settled the composer and graph chrome that Promote starts from, and reserved `reconciliation_claim` in delete Impact.

Authoritative model: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md). The promote comparison (§5.3) and neighborhood walk (§5.4) are written there as "future UI"; this spike makes them current.

## Out of scope (for this spike)

- Reconciliation Claims (schema, UI, and the `name_format` claim)
- Managing Identity Claims after Promote: removing a member, removing or editing a claim, adding evidence later, and the **weak-claim review alert** for claims whose evidence was deleted (Spike 10; model §5.2, [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md))
- Canonical merge (`merged_into_id` behavior)
- Stub handles created without a Subject (asserted / inferred Places, "Mother of James")
- `canonical_entity_notes`
- Pages for association kinds (Participation, Location, Relationship). Promote creates those handles; they have no list or detail page.
- Likeness / depiction thumbnails ([`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md))
- Family tree, timeline, map (Narrative projections)
