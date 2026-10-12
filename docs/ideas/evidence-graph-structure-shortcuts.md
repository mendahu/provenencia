# Shortcuts for common Evidence graph structures

**Status:** idea, 2026-10-12. Not scheduled.

Captured while dogfooding the Evidence graph. Spike 10's observation-default PR (S10-10) was descoped for this: prefilling `subject` on a new person–event, or `part_of` on a new place–place, fills one field. The repeated work is the whole cluster.

## The problem

A wedding is the same motions every time. Place a generic event, set the type to wedding, connect two people, connect a place. A birth and a residence are the same kind of cluster with different types. The seeded event types are `marriage`, `birth`, and `residence` ([`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.4).

A second case is a copy, not a new shape. An obituary listed the deceased's six grandsons, all of the same city. Each one was a person, a residence event, and a connection to that place. The event type, the dates, and the place stayed the same. Only the name changed. That cluster was built six times, then cited.

## A pattern, placed in one action

Beside the generic event control, offer a few common patterns. "Create a wedding" is the example. A modal collects the values that differ — the two people, the date, the place — and one action places the event, the people, the connections, and the place.

What that writes is a batch of Observations: the event type, the participations, the location, the names, the date. The open problem is the Citation. Those Observations still have to be cited, and the citation composer is how that happens today. Dropping empty cards and then opening the composer once per fact puts the repetition back. The useful end is one citation for the batch, and a way into the composer that already has that batch. How that handoff works is not decided.

Birth and residence are the other first candidates. The list should stay short.

## Paste a structure that is already on the canvas

Select the cards and paste a copy. The copy keeps the event types, the dates, the places, and the same Citation. The researcher changes the values that differ (the next grandson's name) and leaves the rest.

A wedding shortcut invents a shape from a form. Paste repeats a shape already on this Source and points it at the citation that is already there.

## Open questions

1. **Which patterns.** A fixed product list (`marriage`, `birth`, `residence`), or can a project add one?
2. **Who the modal may name.** New people and places only, or ones already on this graph, or ones already in the catalog?
3. **The composer.** Open it with the new Observations drafted on one Citation, or place the cards and cite them after?
4. **Paste and the Citation.** Duplicate the Observations onto the Citation that was copied, or open a new Citation that starts from the same text?
5. **What is selected.** Subjects alone, or the connections between them? A residence pasted without its place is the wrong copy.
6. **Where the copy lands.** Offset from the original so the two clusters can be told apart.
