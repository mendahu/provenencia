# Promote: connections as subrows of the subjects they join

**Status:** idea, 2026-10-09. Waiting on design. Not scheduled.

Today's page: [`promote-graph-alignment.md`](../promote-graph-alignment.md) §3 (the page) and §9.1 (filing). Bridge filing rules: [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §5.4.

## The problem

The Promote page lists primary Subjects (persons, events, places) as rows. Bridges (participations, locations, relationships, place relationships) aren't rows. They sit in a **connections panel** at the bottom of the page, collapsed by default ("22 connections will be filed"), where each one can be switched off.

That arrangement has two problems.

1. **Out of sight.** A connection's fate depends on its two ends: it files only when both ends become handles on this Done. The rows that decide this are at the top of the page and the connection is at the bottom, collapsed. A researcher who sets John to Skip has no visible cue that "Mary, parent of John" just stopped filing, or the reverse.
2. **The switch-off isn't honest.** Bridges file themselves once both ends are handles (§5.4): on this Done, on a later Done for the same Source, or when a later claim makes the second end a member. Switching a connection off only kept it out of *this* Done, so it came back silently. Making the switch durable (a declined flag on the bridge Subject, tried in mendahu/provenencia#323 and withdrawn) adds a hidden third state: the Evidence graph says X, but don't file X. If the researcher disputes the bridge, the honest fix is on the Evidence graph (delete it or change its type or role). If they accept the reading but doubt the Source, that's credibility and reconciliation, not hiding it from the canonical graph.

## The idea

Show each connection as a **subrow** under a Subject row it joins, and drop the switch. Its state follows from its ends. The page already computes this in `PromoteFlow` (`ConnectionState`):

| State | When | Subrow reads |
| --- | --- | --- |
| files | both ends will be handles | "Mary · parent of · John" |
| waits | an end is Skip, unset, or not on this page | "… · waits for John" (dimmed) |
| self-link | both ends land on one handle | "… · both ends are one record" (warning) |

Setting John to New turns his subrows on; setting him to Skip dims them. Nothing toggles on its own, so what the page shows is what Done files. The bottom panel goes away; the footer keeps its count ("filing 12 subjects and 22 connections").

A connection the researcher disputes gets a way back to the graph (open the bridge on the Evidence graph to edit or delete it), not a switch.

## Open questions for design

1. **Which row hosts a connection?** A bridge has two ends.
   - *Mirror* it under both ends: the most direct, but every connection appears twice.
   - *One host by rule*: participations and locations under the event (the hub that ties people and places); relationships under the `person` end; place relationships under the part (child) place. Less noise, but a person row wouldn't show the events it's in.
   - A middle option: one host, plus a compact "+3 connections" on the other end that points to it.
2. **Ends that aren't rows on this page.** On a single-subject Promote, or for an end that's already filed (an anchor), the connection goes under the visible end and names the other ("Toronto, already filed").
3. **Already-filed connections:** hidden (as today), or shown dimmed for context.
4. **Busy rows:** an event with eight participants, or a person in many events. Collapse after two or three ("+5 more").
5. **Grouping by assessment:** subrows travel with their host row, so that grouping still works.
6. **The evidence sheet:** its "Through Mary" groups are the same bridges. A subrow could open the sheet at that group, and the sheet could list the row's connections.
7. **Accessibility:** a subrow reads as part of its host ("Mary, parent of John, files on Done"), so VoiceOver users hear why a connection waits.

## What it would change

- **Swift:** `PromoteFlow` gains `connections(hostedBy:)` built on the existing `resolve` and `ConnectionState`. `bridgeOff`, `toggleBridge`, and the connections panel go; `PromoteAlignmentRow` renders subrows.
- **Go:** `promote.Batch.SkipBridgeIDs` and `ApplyPromoteGraphAlignmentRequest.skip_bridge_ids` go; `FileSourceBridgesTx` and `FileBridgesTx` lose their `skip` parameter. Filing is otherwise unchanged.
- **Docs:** `promote-graph-alignment.md` §3 and §9, and the "switched off" bullet in `conclusion-layer-data-model.md` §5.4.
