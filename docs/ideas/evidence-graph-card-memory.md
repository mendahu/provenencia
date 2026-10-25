# Evidence graph card image memory

**Status:** idea only — not roadmapped. Parked from S11-21 ([plan](../deployment-plan/spike-11/prs/S11-21-large-source-performance.md) D-17, D-18; [ship log](../deployment-plan/spike-11/prs/S11-21-large-source-performance.ship.md)), 2026-10-24. Pull it in if memory becomes a problem on a large Source.

## What happens

Since S11-21, every Evidence graph card is a [`GraphCanvasItemView`](../../macos/App/Features/GraphCanvas/GraphCanvasItemView.swift). It paints once into an image its layer shows, and reuses that image until its pixel size or appearance changes. That is what made panning and zooming smooth: a 10 s pan went from 9,970 card paints to 2.

The cost is that every card keeps its image, including cards far off screen. Memory grows with the cards on a Source times the render scale squared.

Measured on the 110-Subject dogfood Source (`phys_footprint`, Apple M1, user driving):

| Display | 1× zoom | 1.25× zoom | Zoomed out |
| --- | --- | --- | --- |
| 1× external | 231 MB | 287 MB | — |
| Built-in Retina (2×) | about 585 MB | about 774 MB | 338 MB |

Peaks while zooming reached about 1.3 GB. On Retina that is about 5 MB per card at 1×. One card's image at 2.5× should be about 2.4 MB, but each card costs about 6.5 MB. That suggests each card's pixels may be held twice.

## Options

- **Check for a second copy first.** `updateLayer` stores the `CGImage` from an `NSBitmapImageRep` as `layer.contents`, and Core Animation may keep its own uploaded copy. If so, an `IOSurface` as the layer's contents, or drawing straight into a layer-owned surface, could halve memory without changing behavior.
- **Drop the images of cards far off screen.** Clear `layer.contents` and the cached image for cards outside the visible rect plus a margin, and repaint them as they come near. That bounds memory by the window, not by the Source. It costs the occasional repaint as a card scrolls into view, which is nothing like the old per-frame repaint. This keeps every card's view mounted, so it is not the "mount only the cards near the window" option the archived [drag performance decisions](archive/evidence-graph-drag-performance.md) refused.
- **Render zoomed-out cards at a lower scale.** The item scale is `backingScale × min(max(1, zoom), 1.25)`, so cards zoomed out to 0.25× still render at the full backing scale. Letting the scale follow the zoom below 1×, down to about 0.5×, would cut zoomed-out memory 4×. The cost is one repaint per zoom settle, and soft cards during a pinch in until it settles.

## Open questions

- Which number is a problem: the steady footprint on a big Source, or the zoom peak?
- Does the second-copy theory hold? Reading `footprint` before and after replacing `layer.contents` with an `IOSurface` on one card would settle it.

It does not share a budget with [S11-31](../deployment-plan/spike-11/prs/S11-31-evidence-graph-cap.md), which caps the Go engine's loaded Evidence graphs: that is engine rows, this is app card images, and each is bounded where it lives (S11-31 plan, 2026-10-25).
