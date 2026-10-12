# Spike 10 — completed

## S10-01 — Button hit boxes

**Landed 2026-10-11** in [#361](https://github.com/mendahu/provenencia/pull/361). Toolbar place, person, event, and connect buttons take a click on their padding. The bridge citation hit sits on the painted pencil, not one button-width to the left.

## S10-02 — Zoomed-out grid and edges

**Landed 2026-10-11** in [#362](https://github.com/mendahu/provenencia/pull/362). The background grid is one repeated cell. It fades as a whole when a cell would be under about a screen point. Connector strokes stay one screen point wide at every zoom. A grown document uses the same tile.

## S10-03 — Canvas grows

**Landed 2026-10-11** in [#363](https://github.com/mendahu/provenencia/pull/363). The document stays 4000×4000 while every card has a card-width of margin. A card inside that margin extends that edge by 800 points, and `maxCell` follows. Dragging and a new composer card use the grown size. The grid tile does not change. A small graph still opens on 4000×4000.

## S10-04 — Camera survives leaving the graph

**Landed 2026-10-11** in [#364](https://github.com/mendahu/provenencia/pull/364). The workspace session stores magnification and content offset per Source. The composer, Back, and Forward restore it. Growing the document does not reset it. Quit drops it. A Source with no stored camera opens centered at 1×.

## S10-05 — Text area scrollbar

**Landed 2026-10-12** in [#368](https://github.com/mendahu/provenencia/pull/368). `PVTextArea` inserts a newline at the caret when Return is pressed. Overflow shows a scrollbar that tracks the caret. A short note does not show an empty track.

## S10-06 — Omnibar whole-word rank

**Descoped 2026-10-11** in [#368](https://github.com/mendahu/provenencia/pull/368). A title that contains `marriage` already ranks above obituaries for Marion Margaret. No change to `core/search`.

## S10-07 — Oldest user on an existing project

**Landed 2026-10-12** in [#369](https://github.com/mendahu/provenencia/pull/369). Opening a project you do not belong to selects the user with the smallest UUIDv7 id. If you are already a contributor, you stay selected. An empty user list still offers create-new. The list stays alphabetical.

## S10-08 — Enter saves a Property title

**Landed 2026-10-12** in [#370](https://github.com/mendahu/provenencia/pull/370). Enter in a single-line inline edit saves when Save is enabled. When Save is disabled, Enter leaves the edit the same way Cancel does. Escape cancels. A multiline field still takes Enter as a newline.

## S10-46 — Simplified cards when zoomed out

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Drawing the card in AppKit made a zoomed-out graph fast enough. The full card stays at every zoom. S10-D14 is not drawn.

## S10-47 — Mount cards near the window

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). The same draw builds a view for every card on the Source. Cards outside the window stay mounted.

## S10-48 — Canvas cover

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Subject and bridge cards are drawn in AppKit, and the first paint is fast enough that a loading cover is not needed. No `ProgressView` over the canvas.
