# Kinship view beyond matching

**Status:** idea only — not roadmapped. Left over when [kinship shapes](../deployment-plan/spike-11/kinship-shapes.md) was pulled into Spike 11 (2026-10-24), which teaches matching that a shared marriage, a birth's parents and a parent chain imply spouse, parent and grandparent links.

## What stays here

- **Family trees and person pages.** "Who are this person's spouses, parents, children and grandparents" from the canonical layer has to read every shape as one fact, or it misses relationships recorded through events. One canonical kinship view on `core/kinship` (Spike 11), which matching already uses.
- **Other patterns.** Siblings from shared parents; aunts and uncles (`pibling` / `nibling`); godparents at a baptism; a baptism's or christening's parents; divorce; census households; great-grandparents and deeper.
- **Link reading below the layers.** `core/kinship` reads links through `segmentcompare.Keys`, so it imports the whole compare stack (`segmentcompare` imports `propertycompare`, `subjectcompare`, `valuecompare`). Before a tree or person page imports `core/kinship`, move `core/segmentcompare/key.go`'s link reading (`Link`, `Terms`, `Key`, `LinkOf`, `Keys`, `EventEndpoint`) to its own pure package and point `segmentcompare`, `core/kinship` and the loaders at it (S11-38 review M-3).
- **Bounding event fan-out.** `kinship.Derive` makes one implied link per pair of participants in an event (D-5), so one marriage with k `spouse` participations yields k(k-1)/2 links (1,000 spouses → about 500,000 links, under a second). Real data has two spouses per marriage, so S11-38 accepted it. If an imported or merged catalog ever puts hundreds of people on one event, bound participants per event where catalog data enters (the loaders), not with a knob in `core/kinship` (S11-38 review SEC-1).
- **Conflicts between shapes.** A direct `A sibling B` and a shared marriage as spouses. Flag it, or stay silent?
- **A project's own patterns**, beside its own terms ([`term-management.md`](term-management.md)).
- **Citing an implied fact on a tree** ("spouses: both in the 1842 marriage, EVT-…").
