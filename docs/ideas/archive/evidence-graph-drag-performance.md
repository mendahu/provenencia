# Evidence graph drag performance (archived)

**Done.** A drag moves the card. Opening the graph draws every card in AppKit. Canvas: [`macos/App/Features/GraphCanvas/README.md`](../../../macos/App/Features/GraphCanvas/README.md).

## Decisions

- A drag moves one card view's frame. It does not rebuild the card.
- Subject and bridge cards are drawn in AppKit. Opening a graph builds a view for every card on the Source, including cards outside the window.
- The grid is one repeated tile. A grown or zoomed-out document does not allocate a document-sized bitmap.
- The full card stays at every zoom. A simpler zoomed-out face, mounting only the cards near the window, and a loading cover over the canvas were refused. Drawing the card in AppKit made those unnecessary.
- Drawing in AppKit alone was not enough for panning (S11-21, 2026-10-24). A card that paints in `draw(_:)` is repainted whenever its visible area changes, so a pan of the 110-Subject dogfood Source repainted every card on every frame: 9,970 paints in 10 s. Each card is now its own layer and paints through `updateLayer` into an image it reuses until its pixel size or appearance changes ([`GraphCanvasItemView`](../../../macos/App/Features/GraphCanvas/GraphCanvasItemView.swift)). A pan now paints 2 cards, and a pinch at most one per card. The memory this costs is parked in [`evidence-graph-card-memory.md`](../evidence-graph-card-memory.md).

## Leftovers

| Theme | Home |
| --- | --- |
| Edges as Core Animation layers, updated on drag without SwiftUI | Descoped. The edge layer stays where it is; S11-21's pan and zoom profiles showed edges cost almost nothing. |
