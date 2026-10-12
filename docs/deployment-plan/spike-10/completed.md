# Spike 10 — completed

## S10-01 — Button hit boxes

**Landed 2026-10-11** in [#361](https://github.com/mendahu/provenencia/pull/361). Toolbar place, person, event, and connect buttons take a click on their padding. The bridge citation hit sits on the painted pencil, not one button-width to the left.

## S10-02 — Zoomed-out grid and edges

**Landed 2026-10-11** in [#362](https://github.com/mendahu/provenencia/pull/362). The background grid is one repeated cell. It fades as a whole when a cell would be under about a screen point. Connector strokes stay one screen point wide at every zoom. A grown document uses the same tile.

## S10-03 — Canvas grows

**Landed 2026-10-11** in [#363](https://github.com/mendahu/provenencia/pull/363). The document stays 4000×4000 while every card has a card-width of margin. A card inside that margin extends that edge by 800 points, and `maxCell` follows. Dragging and a new composer card use the grown size. The grid tile does not change. A small graph still opens on 4000×4000.

## S10-04 — Camera survives leaving the graph

**Landed 2026-10-11** in [#364](https://github.com/mendahu/provenencia/pull/364). The workspace session stores magnification and content offset per Source. The composer, Back, and Forward restore it. Growing the document does not reset it. Quit drops it. A Source with no stored camera opens centered at 1×.

## S10-06 — Omnibar whole-word rank

**Descoped 2026-10-11.** A title that contains `marriage` already ranks above obituaries for Marion Margaret. No change to `core/search`.

## S10-46 — Simplified cards when zoomed out

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Drawing the card in AppKit made a zoomed-out graph fast enough. The full card stays at every zoom. S10-D14 is not drawn.

## S10-47 — Mount cards near the window

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). The same draw builds a view for every card on the Source. Cards outside the window stay mounted.

## S10-48 — Canvas cover

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Subject and bridge cards are drawn in AppKit, and the first paint is fast enough that a loading cover is not needed. No `ProgressView` over the canvas.
