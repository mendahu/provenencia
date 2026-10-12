# Spike 10 — completed

## S10-46 — Simplified cards when zoomed out

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Drawing the card in AppKit made a zoomed-out graph fast enough. The full card stays at every zoom. S10-D14 is not drawn.

## S10-47 — Mount cards near the window

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). The same draw builds a view for every card on the Source. Cards outside the window stay mounted.

## S10-48 — Canvas cover

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Subject and bridge cards are drawn in AppKit, and the first paint is fast enough that a loading cover is not needed. No `ProgressView` over the canvas.
