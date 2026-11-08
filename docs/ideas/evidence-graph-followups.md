# Evidence graph follow-ups

**Status:** idea only — not roadmapped. Paint, memory, and selection leftovers from the Evidence graph / GraphCanvas. A family-tree second host is separate ([`graph-canvas-second-host.md`](graph-canvas-second-host.md)); placing common card clusters is separate ([`evidence-graph-structure-shortcuts.md`](evidence-graph-structure-shortcuts.md)).

## Card image memory

Parked from S11-21 (2026-10-24). Pull it in if memory becomes a problem on a large Source.

Since S11-21, every Evidence graph card is a [`GraphCanvasItemView`](../../macos/App/Features/GraphCanvas/GraphCanvasItemView.swift). It paints once into an image its layer shows, and reuses that image until its pixel size or appearance changes. That is what made panning and zooming smooth: a 10 s pan went from 9,970 card paints to 2.

The cost is that every card keeps its image, including cards far off screen. Memory grows with the cards on a Source times the render scale squared.

Measured on the 110-Subject dogfood Source (`phys_footprint`, Apple M1, user driving):

| Display | 1× zoom | 1.25× zoom | Zoomed out |
| --- | --- | --- | --- |
| 1× external | 231 MB | 287 MB | — |
| Built-in Retina (2×) | about 585 MB | about 774 MB | 338 MB |

Peaks while zooming reached about 1.3 GB. On Retina that is about 5 MB per card at 1×. One card's image at 2.5× should be about 2.4 MB, but each card costs about 6.5 MB. That suggests each card's pixels may be held twice.

**Options:**

- **Check for a second copy first.** `updateLayer` stores the `CGImage` from an `NSBitmapImageRep` as `layer.contents`, and Core Animation may keep its own uploaded copy. If so, an `IOSurface` as the layer's contents, or drawing straight into a layer-owned surface, could halve memory without changing behavior.
- **Drop the images of cards far off screen.** Clear `layer.contents` and the cached image for cards outside the visible rect plus a margin, and repaint them as they come near. That bounds memory by the window, not by the Source. It costs the occasional repaint as a card scrolls into view. This keeps every card's view mounted (S11-21 refused mount-only-near-window).
- **Render zoomed-out cards at a lower scale.** The item scale is `backingScale × min(max(1, zoom), 1.25)`, so cards zoomed out to 0.25× still render at the full backing scale. Letting the scale follow the zoom below 1×, down to about 0.5×, would cut zoomed-out memory 4×.

**Open questions:** Which number is a problem — steady footprint or zoom peak? Does the second-copy theory hold?

It does not share a budget with S11-31, which caps the Go engine's loaded Evidence graphs: that is engine rows, this is app card images.

## Card shadows fall upward

S11-21 ship log; dogfood 2026-10-29. `PVPainter` passes a positive shadow offset straight to `NSShadow` in a flipped view, so card shadows cast up instead of down. Shadows should cast downward in the flipped graph canvas coordinate space.

## Card colors may shift on wide-gamut screens

S11-21 ship log; dogfood 2026-10-29. Check by eye before fixing. Cached card images are drawn in untagged device RGB (`GraphCanvasItemView`), so colors may differ from the live views on a P3 display. Cached paints should use a color space that matches the live views.

## No keyboard way to add one card to a multi-selection

Deferred from S10-57; dogfood 2026-10-17. Shift-click adds or removes one card, but keyboard and VoiceOver users have no way to do the same. Wanted: a keyboard (and VoiceOver) path that toggles the focused card in the selection the way Shift-click does. When GraphCanvas grows a shared selection type ([`graph-canvas-second-host.md`](graph-canvas-second-host.md)), put the additive key there so every host gets it.

## No reloading sign on a revisited canvas

`isRereadingGraph` is drawn only in the error callout (`EvidenceGraphView.swift`), so a revisited stale canvas shows its old cards with no sign that it is rereading. Needs a brief.
