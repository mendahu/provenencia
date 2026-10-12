# Evidence graph drag performance (archived)

**Done.** A drag moves the card. Opening the graph draws every card in AppKit. Canvas: [`macos/App/Features/GraphCanvas/README.md`](../../../macos/App/Features/GraphCanvas/README.md).

## Decisions

- A drag moves one card view's frame. It does not rebuild the card.
- Subject and bridge cards are drawn in AppKit. Opening a graph builds a view for every card on the Source, including cards outside the window.
- The grid is one repeated tile. A grown or zoomed-out document does not allocate a document-sized bitmap.
- The full card stays at every zoom. A simpler zoomed-out face, mounting only the cards near the window, and a loading cover over the canvas were refused. Drawing the card in AppKit made those unnecessary.

## Leftovers

| Theme | Home |
| --- | --- |
| Edges as Core Animation layers, updated on drag without SwiftUI | Descoped. The edge layer stays where it is. |
