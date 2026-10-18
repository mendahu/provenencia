# Evidence graph focus views

**Status:** idea only — not roadmapped. Parked from Spike 12 ([source structure](../deployment-plan/spike-12/source-structure.md)), 2026-10-18.

Multiple graphs per Source let a researcher split a big Source by hand. This also holds the "layers or filters" half of the dogfood note "Dense Evidence graphs (census-scale)" (2026-09-24, descoped from Spike 10), whose splitting half Spike 12 covers. A single graph can still be large: one obituary produced over 100 Subjects and over 200 Observations.

Ideas, no schema change needed:

- **One connected component at a time.** Pick a cluster and hide the rest.
- **Focus on a Subject.** Show it and its neighbors to N hops.
- **Filter by Citation or Artifact.** Show only Subjects observed on page 12; the data is already reachable through Observations.
- **Saved regions.** Named frames on the canvas.

Related: drag and first-paint performance is archived in [`archive/evidence-graph-drag-performance.md`](archive/evidence-graph-drag-performance.md); canvas growth and box-select landed in Spike 10.
