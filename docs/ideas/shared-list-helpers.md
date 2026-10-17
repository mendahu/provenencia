# Shared list helpers

**Status:** idea only — not roadmapped. Small duplication found while planning S10-41 (the Sources list filters). Each item is a cleanup, not a feature. "In-list text matching" and "Primary Subject kinds" moved into Spike 11 (S11-08, S11-23).

## Source graph mutations

`WorkspaceSession.refreshSourceGraphProgress(for:)` switches over the graph mutations by hand; `CatalogQueryRegistry` lists the same ones for `.sourceGraph`. After S10-41 the lists differ only by `.deletedSource`, which would also be correct to patch (`sourcegraphprogress.Get` returns zeros for a gone Source, and `isZero` removes the entry). A `CatalogMutation.graphSourceId` read by both would keep them in step.
