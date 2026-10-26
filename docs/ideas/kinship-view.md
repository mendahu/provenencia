# Kinship view beyond matching

**Status:** idea only — not roadmapped. Left over when [kinship shapes](../deployment-plan/spike-11/kinship-shapes.md) was pulled into Spike 11 (2026-10-24), which teaches matching that a shared marriage, a birth's parents and a parent chain imply spouse, parent and grandparent links.

## What stays here

- **Family trees and person pages.** "Who are this person's spouses, parents, children and grandparents" from the canonical layer has to read every shape as one fact, or it misses relationships recorded through events. The graph cache holds the canon's implied links by end (S11-49), so a tree or person page reads them as matching does.
- **Other patterns.** Siblings from shared parents; aunts and uncles (`pibling` / `nibling`); godparents at a baptism; a baptism's or christening's parents; divorce; census households; great-grandparents and deeper.
- **Link reading below the layers.** `core/kinship` reads links through `segmentcompare.Keys`, so it imports the whole compare stack (`segmentcompare` imports `propertycompare`, `subjectcompare`, `valuecompare`). Before a tree or person page imports `core/kinship`, move `core/segmentcompare/key.go`'s link reading (`Link`, `Terms`, `Key`, `LinkOf`, `Keys`, `EventEndpoint`) to its own pure package and point `segmentcompare`, `core/kinship` and the loaders at it (S11-38 review M-3).
- **Conflicts between shapes.** A direct `A sibling B` and a shared marriage as spouses. Flag it, or stay silent?
- **A project's own patterns**, beside its own terms ([`term-management.md`](term-management.md)).
- **Citing an implied fact on a tree** ("spouses: both in the 1842 marriage, EVT-…").
