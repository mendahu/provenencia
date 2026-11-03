# Narrower canon walk

**Status:** idea, 2026-10-26. Was scheduled in Spike 12 (S12-43 / S12-51); parked 2026-10-29 with [`promote-ranking.md`](promote-ranking.md) when Spike 12 closed ([audit 2](../deployment-plan/spike-14/app-health-audit-2.md) B-8). From the S11-39 CI fix round, where the user raised the warm budget from 250 ms to 350 ms.

## The problem

Every proposal walks the whole connected component of the canon from its seeds. On the large fixture that is about 18,000 handles and 24,000 links, plus about 6,000 implied links, so per-link work grows with the tree, not with the Source. The warm proposal on a 4,000-person connected tree costs about 115 ms locally and 297 ms on CI (S11-39).

## A shape

- Walk only as far as alignment can use: the hops a layer link can correspond to from the seeds and their placed neighbors.
- Keep the walk deterministic.
- Measure against `TestLargeSourceBudgets`.

## Memoize the canon (2026-11-03)

A profile of warm proposals alone (`TestLargeSourceBudgets/warm`, an allocs profile before and after ten proposals) shows each one allocating about 122 MB in about 700,000 allocations with no query: `promotealign.loadCanon` is 69 % of it (about 85 MB a proposal), `kinship.(*LinkSet).Add` 26 %, `graphalign.newState` 18 %. A one-to-two-collection run over a 170 MB heap is what flaked the 350 ms budget on CI (388 ms, best of three). The layer and the candidates are already memoized on the Evidence graph (`loadLayerMemo`, `loadCandidatesMemo`); the canon is the one large piece rebuilt every time. A canon memo beside them, keyed by the fixed anchors and dropped with the handle graph, would remove most of the churn; a narrower walk would shrink what is memoized. The test now collects before each timed run and takes the best of five, so CI holds while this waits.

## Open questions

- How far a grandparent or place chain needs to reach.
- Whether the 250 ms warm figure can return.
