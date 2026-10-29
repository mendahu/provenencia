# Promote matching: what proves a join (archived)

**Not a backlog.** The problem behind Promote graph alignment, written 2026-10-05 while planning the Spike 9 compare step. Answered by [`promote-graph-alignment.md`](../../promote-graph-alignment.md) (Spike 9), and its scoring is rebuilt in [Spike 11](../../deployment-plan/archive/spike-11/README.md). The full exploration is in git history.

## The stack: how to picture the problem

Every Evidence graph is a **layer**. Layers stack so that Subjects about the same thing line up in **columns**. A canonical entity is a column, and an Identity Claim is a vertical link. Promote lays a new layer onto the stack. A vertical link's strongest evidence is the horizontal structure around it lining up too: her birth in the entity's birth column, her daughter in its child's column.

## Decisions

| | Decision |
| --- | --- |
| Checklist | Rejected. Ticking agreements Property by Property is the wrong amount of work, and it can't see neighbors, which carry most identity. |
| Unit | One proposal for the whole layer (graph alignment), decided per Subject: one Identity Claim per Subject and handle. |
| Evidence | Pins on claims, with backfill, kept (S9-17). Spike 11 rebuilds the sheet: [`claim-evidence.md`](../../deployment-plan/archive/spike-11/README.md). |
| Principles kept | Membership per Subject; create-only Promote; precision over recall; explainable and local. |

## Directions we circled (none chosen)

Matches instead of pins; no stored evidence (corroboration as a read); frozen automatic scores; key properties per kind; anchor-and-propagate; learned weights from accepts and rejects. Anchor-and-propagate became graph alignment. Learned weights are [`matching-test-suite.md`](../matching-test-suite.md).
