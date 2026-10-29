# Narrower canon walk

**Status:** idea, 2026-10-26. Was scheduled in Spike 12 (S12-43 / S12-51); parked 2026-10-29 with [`promote-ranking.md`](promote-ranking.md) when Spike 12 closed ([audit 2](../deployment-plan/spike-14/app-health-audit-2.md) B-8). From the S11-39 CI fix round, where the user raised the warm budget from 250 ms to 350 ms.

## The problem

Every proposal walks the whole connected component of the canon from its seeds. On the large fixture that is about 18,000 handles and 24,000 links, plus about 6,000 implied links, so per-link work grows with the tree, not with the Source. The warm proposal on a 4,000-person connected tree costs about 115 ms locally and 297 ms on CI (S11-39).

## A shape

- Walk only as far as alignment can use: the hops a layer link can correspond to from the seeds and their placed neighbors.
- Keep the walk deterministic.
- Measure against `TestLargeSourceBudgets`.

## Open questions

- How far a grandparent or place chain needs to reach.
- Whether the 250 ms warm figure can return.
