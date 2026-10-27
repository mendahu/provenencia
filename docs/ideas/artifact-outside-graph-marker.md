# Name the Artifacts a graph delete leaves in no graph

**Status:** idea, not scheduled. Left over when the card marker for an Artifact in no Evidence graph was pulled into Spike 12 (S12-38, brief S12-D7; user, 2026-10-27).

## Problem

From S12-22 an Artifact may stand in no Evidence graph, and deleting a graph can leave some there. An Artifact with Citations is protected (the delete is refused and names them); one with none simply drops out of every composer. S12-38 marks such an Artifact on its card, but the delete confirm does not say beforehand which Artifacts it will leave in no graph.

## Rough shape

- The graph delete confirm lists those Artifacts as a notice, through the DeleteImpact recipe's listed groups, without refusing.
- `deleteimpact` would need a listed group that informs without blocking; today a listed group either blocks (an inbound probe) or is released.

## Decide when pulled in

Whether researchers miss the warning once the card marker exists.
