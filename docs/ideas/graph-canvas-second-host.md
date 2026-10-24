# GraphCanvas for a second host (family tree)

**Status:** idea only — not roadmapped. From the S11-21 canvas health review, 2026-10-24 (user: park as a family-tree idea). Pull it in when a family tree, or any second canvas host, is planned.

The [GraphCanvas](../../macos/App/Features/GraphCanvas/README.md) module now carries pan, zoom, the pointer state machine, and items that paint once into a reused image (`GraphCanvasItemView`, `GraphCanvasItemLayer`). S11-21 also moved the paint margin, the content origin, z-order, the placement preview and the drawing toolkit into it or into the design system. What follows is what a family-tree host would still have to build or copy from the Evidence graph. Each item is design work better shaped by the tree's real needs than guessed now.

## Hit targets from measured layout

Today the Evidence graph measures each card (`EvidenceCardLayout`, `EvidenceCardLayoutStore`). It records the layout as a side effect of building the view, and rebuilds every hit target in `publishHitTargets()` from five separate `onChange` triggers and a refresh token. Before a card is measured, frames fall back to hand-written constants.

**Shape:** a `GraphCanvasItemLayout` (content size and named regions) on `GraphCanvasItemView`. `GraphCanvasItemLayer` derives `GraphCanvasHitTarget`s and edge anchors from the views it owns, in its own z-order. The host keeps only region ids, their meaning, and connect policy. This removes the layout store, the side-effect recording, the refresh token and the constants fallback. When the layer derives targets itself, S11-21's `stackIndex(of:)` seam (hit order following paint order) and the `stackIndex:` parameter of `topmostTarget` go away.

## A generic edge layer, with tree routing

`EvidenceGraphEdgeLayer` is a SwiftUI `Canvas`:
- it looks up endpoints by id, adds drag offsets and scales line width by the zoom;
- `EvidenceGraphEdgesHost` is the shim that keeps `pointer.offsets` out of the document body;
- `GraphCanvasEdgeGeometry` only has a horizontal-biased cubic.

**Shape:** a `GraphCanvasEdgeLayer` that takes edges (from id, to id, an opaque style), takes frames from the item layer, follows offsets itself, and calls a host stroke closure.

A tree also needs:
- vertical generations;
- elbow (orthogonal) routing;
- a sibling bus bar and a couple midpoint;
- side-pinned attachments (`GraphCanvasEdgeRoute`: cubic horizontal, cubic vertical, orthogonal).

## Selection, group drag and keys

Selection, toggling, marquee finish, activation, group drag and arrow-key moves live in `EvidenceGraphModel`, and the key bridge lives in `EvidenceGraphView.handleKey`.

**Shape:** a `GraphCanvasSelection` value type (`select`, `toggle`, `finishMarquee(_:additive:)`, `dragSet(anchor:)`) and a `GraphCanvasKeyCommand` mapping. What Return or Space does, snapping and persistence stay product-side.

## Growable document bounds

The board grows as cards approach its edges. This lives in `EvidenceGraphPlacement.canvas(fitting:)`, hard-wired to the subject card width as its margin, and in `EvidenceGraphDragCanvasSize`.

**Shape:** a `GraphCanvasDocumentBounds` (origin and size, `fitting(frames:reserve:margin:growthStep:)`, `union`), with the 4,000-point "downtown" and storage shift product-side.

## Host-chosen zoom range

The zoom range is `GraphCanvasCamera`'s 0.25×–1.25× statics, used by the clamp, the scroll view and `itemContentsScale`. A tree may want to zoom past 1.25× to read dates.

**Shape:** a magnification range parameter on `GraphCanvasScrollView`, passed through to the clamp and to `itemContentsScale`.

## Whole-item hover, keyboard and wiring

- **Hover:** the pointer tracks hover only on an item's nested actions. A whole-card hover needs a fake action that also fires on click. Add a hovered target id.
- **Keyboard:** `GraphCanvasDocumentView` refuses first responder, so keys come from a SwiftUI `.onKeyPress` on an outer container. Document it, or give the canvas a key path.
- **Callbacks:** callbacks are set once in `onAppear` (`wirePointer`, `syncPointerMode`). A `GraphCanvasCallbacks` parameter re-applied on update would stop stale callbacks.
- **Cursors:** the product view pushes and pops the crosshair by hand. The document view could own a cursor per tool mode.
- **Item views:** an async-resource helper (load, then `invalidatePaint()`, cancel off-window) and an accessibility helper would save each item view the boilerplate `EvidenceCardView` carries.
- **Camera storage:** `WorkspaceSession.graphCamera(for:)` is Evidence-named, and tree ids would share its key space.

## Open questions

- Is the family tree one canvas per person, or one large canvas with focus? That decides how much of the bounds and edge work is needed.
- Does the tree layout engine run in Go (like Promote) or in Swift beside the canvas?
