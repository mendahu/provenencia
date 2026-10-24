# GraphCanvas

Reusable spatial canvas **tooling** — pan/zoom (`NSScrollView`), grid-space
coordinates, AppKit pointer ownership, and a plain grid underlay.

Not a product destination. Workspace places (Evidence graph today; later a
family tree or other visualizations) live in their own feature folders and
compose these types. Keep this module free of Interpretation / Evidence /
catalog types so geometry stays extractable (design note §13).

| Type | Role |
| --- | --- |
| `GraphCanvasCamera` | Magnification + content offset (value type) |
| `GraphCanvasCoordinates` | Pure math: magnification clamp and thresholds (`magnificationTolerance`, `minimumMagnification`), the zoom that fills a viewport, the item contents scale, one-screen-point stroke width, grid fade, scroll-origin clamp, grid ↔ document conversion (`documentRect`, `documentPoint(_:origin:)`) and view ↔ document conversion for a camera |
| `GraphCanvasScrollInput` | Trackpad-vs-wheel classification (pure) |
| `GraphCanvasGridMapping` | Grid cell ↔ grid point (snap unit) |
| `GraphCanvasScrollView` | AppKit magnification bridge |
| `GraphCanvasDocumentView` | Scroll document; owns mouse; hosts the base SwiftUI layer and the item layer; canvas accessibility group and rotors |
| `GraphCanvasItemView` | Base for item views, built with `init(contentSize:)`: the view is the content plus `paintOutset` on every side, and `paint()` draws in content coordinates; it renders into an image its layer shows, through `updateLayer` |
| `GraphCanvasItemLayer` / `GraphCanvasItem` / `GraphCanvasPreview` | One `GraphCanvasItemView` per item, built by the host (`makeView`, a `GraphCanvasItemView`), positioned by AppKit; rebuilt only when the item's `Equatable` model changes; owns each item's layer and its scale, their stack order (host order, raised items last; `stackIndex(of:)`), and the placement `preview` |
| `GraphCanvasPointerController` | Hit-test, select, drag (one item, or a group of peers that move with the anchor), empty-canvas pan, selection rectangle (marquee), Shift tracking, Shift-click toggle, place, connect |
| `GraphCanvasHitTarget` / `GraphCanvasPointerHitTesting` | Grid-space rects + pure helpers; a click hits the item drawn on top |
| `GraphCanvasViewportController` | Pan deltas + the live magnification |
| `GraphCanvasGridView` | Optional grid underlay: one repeated cell, not a document-sized bitmap |
| `GraphCanvasEdgeGeometry` | Cubic edges + tuck-under endpoints |

## Reuse contract

Hosts supply product policy; GraphCanvas supplies plumbing.

| Host must supply | Stays in the product folder |
| --- | --- |
| `hitTargets` (string ids + grid-space rects, see [Coordinate seams](#coordinate-seams)) | Card chrome, palette, create dialogs |
| Pointer `mode` + callbacks (select / toggle select / drag began / drag end / marquee ended (ids, additive) / Shift held changed / place / connect) | Catalog / store / FFI / session queries |
| Document `rootView` (paint-only SwiftUI base layer) | Honesty copy and product card content |
| Items pushed to `graphCanvasItems` (id, origin, model, `makeView`, rotor entry) | Rotor names and labels (localized by the host) |
| Document size (host policy). The zoom range is a GraphCanvas default and not host-configurable yet ([second-host idea](../../../../docs/ideas/graph-canvas-second-host.md)) | Auto-layout, virtualization, position tables |

The document view reports Shift through `onShiftHeldChanged`; the host decides what it engages.

**Transfers to a future tree / workflow host:** scroll shell, the item layer and `GraphCanvasItemView` (settled-scale paint, the preview), grid-space coordinates, pointer controller + hit targets, viewport, optional grid and edge helpers.

**Does not transfer** (design note §13.3–§13.4): auto-layout engines, viewport virtualization, polymorphic `subject_positions`, Evidence card chrome, or a Source-scoped fixed board size. Those stay product-owned; GraphCanvas only provides the geometry primitives.

Audit bar: zero Evidence / catalog / store types under `GraphCanvas/`.

## Items: AppKit owns where items are

Items (the Evidence graph's cards) are not part of the host's SwiftUI document. The host pushes them to the
`graphCanvasItems` environment value (`GraphCanvasItemLayer.update(_:)`), and
each becomes its own `GraphCanvasItemView` above the base layer. An item drag moves that
one view's frame from `GraphCanvasDocumentView.mouseDragged` — no SwiftUI runs per
pointer event. An item's view is rebuilt only when its `model` changes, so push
every item whenever anything they draw changes (selection, hover, data, drag
start/end), never per drag frame.

Item views subclass `GraphCanvasItemView` and paint in `paint()`. Each item is its
own layer showing an image of that paint; it repaints when built, when its pixel
size or appearance changes, or on `invalidatePaint()`. `needsDisplay` alone, a pan
or a zoom step shows the last image again.

A host can show one stand-in while it waits for the researcher, such as the item a
click would create: set `GraphCanvasItemLayer.preview` to a
`GraphCanvasPreview(origin:model:makeView:)`. It is placed like an item and drawn below
every item, so cards cover it, but is not one: no id, rotor entry or stack index, and `update` leaves it alone.
Nil removes it.

**Never read `GraphCanvasPointerController.offsets` in a view body** that does
more than draw live geometry. Observation is per property: one read re-runs that
body on every pointer event. Read `draggingItemID` (changes at drag start and end)
to style the dragged item. `draggingItemID` is the anchor, the item under the
pointer. `draggingPeerIDs` is the rest of a group drag: `onDragBegan` returns every
id that moves (anchor included), and each peer gets the anchor's offset. Read
`draggingPeerIDs` the same way. `EvidenceGraphEdgesHost` and `EvidenceGraphDragCanvasSize` are the only views
that follow `offsets`.

**The item contract, in one place.** Subclass `GraphCanvasItemView` and pass
`init(contentSize:)`. Paint in `paint()` in content coordinates (0,0 is the
content's top-left; the base adds the paint margin). Add no subviews. Push
grid-space origins. Repaint with `invalidatePaint()`. Draw with `PVPainter` and
resolve colors with `PVPainter.color(_:in:)`.

Items set their own AppKit accessibility (label, custom actions). The
document view is the canvas's accessibility group and serves document-level
VoiceOver rotors (`NSAccessibilityCustomRotor`) built from items' `rotor` entries,
because items are separate views.

## Pointer ownership

AppKit owns **all** canvas mouse sequences. Hosted SwiftUI is paint-only
(`.allowsHitTesting(false)`). Product hosts publish `hitTargets` and wire
callbacks on `GraphCanvasPointerController`.

**Derive hit rects from measured paint, not constants.** A host's rects must
come from the laid-out views they cover, or they drift as text wraps. The
Evidence graph's `EvidenceCardView` measures its pieces into an `EvidenceCardLayout`
and paints from the same measurement; hit targets and edge anchors are built from
it (constants only before a card is built). Verify against drawn pixels,
not `fittingSize`, which over-reports wrapped text. Do not attach SwiftUI
`DragGesture` / `.onTapGesture` / `.focusable()` to cards — that fought
first-responder and cancelled mid-drag under magnification.

## Coordinate seams

| Seam | Use |
| --- | --- |
| Grid space | What hosts speak: item origins, hit targets, gesture points (`onPlace`, `onHover`, `marqueeRect`), and the input to base-layer drawing |
| Document space | The AppKit document view's own coordinates. `GraphCanvasDocumentView.contentOrigin` is the grid point its top-left stands on; the document view converts each mouse event to a grid point once (`gridPoint(forDocumentPoint:)`) and places each item at `origin − contentOrigin`. SwiftUI base content reads the origin from `\.graphCanvasContentOrigin` and converts with `GraphCanvasCoordinates.documentRect(_:origin:)` / `documentPoint(_:origin:)` |
| AppKit `NSView.convert` on the document | Live pointer hit-test under magnification |
| `GraphCanvasCoordinates` | Pure layout transforms, magnification clamp, pan/zoom clamp, unit tests |

Do not treat `GraphCanvasCoordinates.documentPoint(fromViewPoint:)` as the
runtime pointer path — `NSScrollView` magnification does not map SwiftUI
gesture locations into the hosted document correctly (design note §7.2).

## Input mapping

| Device | Action |
| --- | --- |
| Trackpad two-finger scroll | Pan |
| Trackpad pinch | Zoom |
| Mouse wheel | Zoom toward cursor (manual anchor; not `centeredAt:`) |
| Click-drag on empty document | Pan (idle mode, Shift up) |
| Shift-drag on empty document, or any empty-document drag in `.selecting` mode | Marquee: a grid-space rectangle. On release the pointer calls `onMarqueeEnded` with the ids of the hit targets it overlaps (a shared edge is not a hit) and whether Shift was down at press (additive). A drag that starts on a hit target never marquees |
| Shift-click on a hit target | `onToggleSelect` (no `onSelect`, no item action) |
| Shift key | Document view reports it through `onShiftHeldChanged` (local `flagsChanged` monitor; `false` while a text field has focus or the window is not key) |
| Click / drag on hit targets | Product callbacks (select, move, place, connect) |

Magnification range defaults are **zoom-out oriented** (`0.25` … `1.25`): overview of a
large graph, then back toward identity. The zoom-out floor rises when `0.25` would
make the document smaller than the window, so the area past the canvas edge stays
out of view. The range is a GraphCanvas default and not host-configurable yet ([second-host idea](../../../../docs/ideas/graph-canvas-second-host.md)).
Deliberate zoom-*in* past ~125% is capped in the default camera so layer-scaled
SwiftUI content does not go soft. Items re-render once the zoom has held still for 0.15 s (a magnification change or the end of a live pinch restarts the timer; the scroll
view sets `GraphCanvasItemLayer.contentsScale`); the cap protects the base SwiftUI layer.

`GraphCanvasViewportController.magnification` is the live zoom. The grid fades as a whole when a cell would be under about a screen point; product strokes that should stay one screen point wide use `GraphCanvasCoordinates.screenPixelDocumentWidth`.

Changing `contentID` restores `restoredCamera`, or centers at 1× when the host has none. A content-size change leaves the camera where it is.
`contentOrigin` is the grid-space point the document's top-left stands on. When a host moves it, the scroll offset shifts by the same delta so the viewport stays on the same content, including during an item drag.
Paint `rootView` may update every representable pass — safe because AppKit owns gestures.
Item views sit above it in z-order; items flagged `isRaised` (dragged, activated) sit above other items.
