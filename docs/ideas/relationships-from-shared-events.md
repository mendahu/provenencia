# Relationships from shared events

**Status:** idea only, not roadmapped. Raised 2026-10-19 while shipping S11-14 (segment compare). The user will explore it separately.

## Problem

A relationship between two people can be recorded in two shapes:

- **Direct:** a relationship bridge, `A spouse B` (`relationship · spouse`).
- **Through an event:** both people take part in one event, `A spouse in marriage M` and `B spouse in marriage M` (`participation · spouse · marriage`).

Both shapes say that A and B are spouses. Today nothing in the product knows that.

- **Matching (Spike 11).** Segment compare is one hop and pure ([comparison-layers.md](../deployment-plan/spike-11/comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)). `segmentcompare.Corresponds` lines up keys by type, term and direction, so a participation key never corresponds to a relationship key.
  - If a Source records only the marriage and the research graph records only the spouse relationship (or the reverse), those links add nothing.
  - When both shapes are on both sides, each link counts on its own. B gets support from the spouse link to A, and from the participation link once the marriage event is placed. A marriage with only its type rarely places, so in practice B gets only the direct link.
- **Family trees and expanded person detail pages.** These need "who are this person's spouses, parents and children" from the canonical layer. If that answer reads only relationship bridges, it misses every relationship that was recorded through an event. The canonical layer has to understand both shapes as one fact.

## Open questions

1. **Which patterns.** Two `spouse` participations in one `marriage` mean `relationship · spouse`; `parent` and `child` participations in one `birth` mean `relationship · parent`. What else (divorce, baptism godparents, census households)? A fixed product list, or a table a project can add to (see [`term-management.md`](term-management.md))?
2. **Where it lives.** A derivation only matching reads (the loader or layer 6, S11-16), or one canonical kinship view that matching, family trees and person pages all read?
3. **Stored or derived.** Implied links worked out at read time, or written as rows?
4. **Counting once.** When a Source records both the marriage and the spouse link, how does matching avoid crediting one fact twice?
5. **Weight.** Does an implied link use the segment registry's Other pair, or its own weight? See S11-14's dogfood notes on parent/child weight.
6. **Direction.** Can a birth's roles say which end is the parent, so a directed relationship can be derived?
7. **Conflicts.** A direct `A sibling B` and a shared marriage as spouses. Flag it, or stay silent?
8. **Showing where it came from.** How does a tree or person page cite an implied relationship ("spouses: both in the 1842 marriage, EVT-…")?

## Related

- S11-14 plan, the user's D-11 call: [`S11-14-segment-compare.md`](../deployment-plan/spike-11/prs/S11-14-segment-compare.md). Parent/child and burial link weights are flagged for the matching dogfood.
- [`evidence-graph-structure-shortcuts.md`](evidence-graph-structure-shortcuts.md): placing a wedding cluster in one action makes the event shape more common.
- [`term-management.md`](term-management.md): relationship direction and inverses.
