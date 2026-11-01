# Kinship view beyond matching

**Status:** idea only — not roadmapped. Left over when [kinship shapes](../deployment-plan/archive/spike-11/README.md) was pulled into Spike 11 (2026-10-24), which teaches matching that a shared marriage, a birth's parents and a parent chain imply spouse, parent and grandparent links.

**Spike 13** ([person detail](../deployment-plan/spike-13/person-detail.md#2-family)) takes the Person page's nuclear family (parents, siblings, spouses, children) and a page-only sibling derivation from a shared parent; S13-03 shipped it (`canonlinks.Options{Siblings: true}`), and S13-04 shipped the family read as units over a list of ids (`persondetail.Families`). Sibling evidence for matching stays here.

## What stays here

- **Family trees and person pages.** "Who are this person's spouses, parents, children and grandparents" from the canonical layer has to read every shape as one fact, or it misses relationships recorded through events. The graph cache holds the canon's implied links per pivot, read through `canonlinks.Reader.For` (S11-49), so a tree or person page reads them as matching does. Spike 13's family read (S13-03, S13-04) is shaped as a tree's first step: family units over a list of ids, served from the graph cache, each link saying what it rests on ([working rule](../deployment-plan/spike-13/deployment-plan.md#engine)). A tree repeats it per generation.
- **Other patterns.** Aunts and uncles (`pibling` / `nibling`); godparents at a baptism; a baptism's or christening's parents; divorce; census households; great-grandparents and deeper.
- **Link reading below the layers.** `core/kinship` reads links through `segmentcompare.Keys`, so it imports the whole compare stack (`segmentcompare` imports `propertycompare`, `subjectcompare`, `valuecompare`). Before a tree or person page imports `core/kinship`, move `core/segmentcompare/key.go`'s link reading (`Link`, `Terms`, `Key`, `LinkOf`, `Keys`, `EventEndpoint`) to its own pure package and point `segmentcompare`, `core/kinship` and the loaders at it (S11-38 review M-3).
- **Conflicts between shapes.** A direct `A sibling B` and a shared marriage as spouses. Flag it, or stay silent?
- **A project's own patterns**, beside its own terms ([`term-management.md`](term-management.md)).
- **Citing an implied fact on a tree** ("spouses: both in the 1842 marriage, EVT-…").
