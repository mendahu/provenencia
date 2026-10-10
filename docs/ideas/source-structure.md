# Source structure: multiple Evidence graphs and nested Sources

**Status:** problem statement only — no solution chosen, not roadmapped. Captured from dogfooding on real Sources (2026-10-10).

The current Source schema assumes one flat Source with one Evidence graph. Two real cases strain that. They may share a fix, so they are written down together.

## Problem 1: one Source, more than one Evidence graph

A single Evidence graph per Source is already large on a simple Source. A grandmother's obituary produced **over 100 Subjects and over 200 Observations**.

A Source like a nationwide census is worse. The families the researcher cares about are scattered across it and unconnected, apart from having been counted in the same enumeration. They could all go on one graph under a Source called "that census", but nothing would link them to each other and the graph would keep growing.

### Idea: several Evidence graphs per Source

Let a Source hold more than one Evidence graph, created from a section on the Source page much like Artifacts.

It is not only "create another graph". Graphs and Artifacts relate, and there is no single rule for how:

- **Census:** a scan of one page, with the Evidence graph tied to that page. When composing a Citation on that graph, the other pages should not be selectable.
- **Book:** a large local history that mentions many families, periods, events and places, with one PDF copy covering all of them. Several graphs share one Artifact. Here an Artifact is *not* scoped to a graph.
- **One-Artifact Sources:** often there is only a single Artifact, so the question is moot.

So the link between a graph and its Artifacts should be an association that speeds up and constrains the UI (for example, offering only the associated Artifacts in the Citation composer), not a hard rule.

## Problem 2: nested Sources

Sources often sit inside other Sources. A newspaper obituary has three levels:

1. **The obituary.** This is the thing that matters semantically: an obituary, a memorial, a notice of death.
2. **The newspaper issue** it was found in, full of unrelated articles.
3. **The publication** the issue belongs to (for example, the Medicine Hat News).

The schema doesn't fully capture this. While entering obituaries this caused real confusion: is the Source type of an obituary "obituary", or "newspaper" because that is where it was found?

### Idea: parent Sources

A parent Source could carry shared information and act as a template:

- Create a publication Source and record its details once.
- Create each issue as a child of it, inheriting the publication's Source metadata.
- The obituary is a child of the issue.

The two problems may connect: the Evidence graph might belong to the child Sources and not the parents. A publication or an issue would have no graph of its own.

## Open questions

- Is a graph's link to an Artifact a required scope, an optional hint, or a default that can be overridden?
- How does the Source page show several graphs next to the existing Artifact list?
- How deep can Source nesting go, and is the depth fixed (publication → issue → item) or open?
- Which fields does a child inherit from its parent, and can it override them?
- Do graphs attach only to leaf Sources, or can any Source have one?
- What happens to existing Sources: they have one implicit graph and no parent.

## Related

- [`source-to-source-relationships.md`](source-to-source-relationships.md): how one Source talks about another, via `mentions` / `remark`. That is commentary, not containment; this note is about containment.
- [`evidence-graph-drag-performance.md`](evidence-graph-drag-performance.md): large graphs are also a performance problem.
- Dogfood notes: *Evidence graph canvas is too small* and *Box-select and move multiple Subject cards* in [`../dogfood/ux.md`](../dogfood/ux.md).
