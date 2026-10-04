# GraphCanvas

Reusable spatial canvas **tooling** — pan/zoom (`NSScrollView`), content-space
coordinates, AppKit pointer ownership, and a plain grid underlay.

Not a product destination. Workspace places (Evidence graph today; later a
family tree or other visualizations) live in their own feature folders and
compose these types. Keep this module free of Interpretation / Evidence /
catalog types so geometry stays extractable (design note §13).

| Type | Role |
| --- | --- |
| `GraphCanvasCamera` | Magnification + content offset (value type) |
| `GraphCanvasCoordinates` | Pure camera math, mag clamp, scroll-origin clamp (layout / tests) |
| `GraphCanvasScrollInput` | Trackpad-vs-wheel classification (pure) |
| `GraphCanvasGridMapping` | Grid cell ↔ content point (snap unit) |
| `GraphCanvasScrollView` | AppKit magnification bridge |
| `GraphCanvasDocumentView` | Scroll document; owns mouse; hosts the base SwiftUI layer and the item layer; canvas accessibility group and rotors |
| `GraphCanvasItemLayer` / `GraphCanvasItem` | One `NSHostingView` per item (card), positioned by AppKit; rebuilt only when the item's `Equatable` model changes |
| `GraphCanvasPointerController` | Hit-test, select, drag, empty-canvas pan, place, connect |
| `GraphCanvasHitTarget` / `GraphCanvasPointerHitTesting` | Document rects + pure helpers |
| `GraphCanvasViewportController` | Pan deltas + live AppKit `convert` under cursor |
| `GraphCanvasGridView` | Optional empty grid document content |
| `GraphCanvasEdgeGeometry` | Cubic edges + tuck-under endpoints |

## Reuse contract

Hosts supply product policy; GraphCanvas supplies plumbing.

| Host must supply | Stays in the product folder |
| --- | --- |
| `hitTargets` (string ids + document rects) | Card chrome, palette, create dialogs |
| Pointer `mode` + callbacks (select / drag end / place / connect) | Catalog / store / FFI / session queries |
| Document `rootView` (paint-only SwiftUI base layer) | Honesty copy and product card content |
| Items pushed to `graphCanvasItems` (id, origin, model, content, rotor entry) | Rotor names and labels (localized by the host) |
| Magnification range + document size (host policy) | Auto-layout, virtualization, position tables |

**Transfers to a future tree / workflow host:** scroll shell, pointer controller + hit targets, viewport convert, optional grid and edge helpers.

**Does not transfer** (design note §13.3–§13.4): auto-layout engines, viewport virtualization, polymorphic `subject_positions`, Evidence card chrome, or a Source-scoped fixed board size. Those stay product-owned; GraphCanvas only provides the geometry primitives.

Audit bar: zero Evidence / catalog / store types under `GraphCanvas/`.

## Items: AppKit owns where cards are

Cards are not part of the host's SwiftUI document. The host pushes them to the
`graphCanvasItems` environment value (`GraphCanvasItemLayer.update(_:)`), and
each becomes its own `NSHostingView` above the base layer. A card drag moves that
one view's frame from `GraphCanvasDocumentView.mouseDragged` — no SwiftUI runs per
pointer event. An item's view is rebuilt only when its `model` changes, so push
every item whenever anything they draw changes (selection, hover, data, drag
start/end), never per drag frame.

**Never read `GraphCanvasPointerController.offsets` in a view body** that does
more than draw live geometry. Observation is per property: one read re-runs that
body on every pointer event. Read `draggingItemID` (changes at drag start and end)
to style the dragged item. The Evidence graph's edge layer is the only view that
follows `offsets`.

Item content keeps its own SwiftUI accessibility (label, actions, children). The
document view is the canvas's accessibility group and serves document-level
VoiceOver rotors (`NSAccessibilityCustomRotor`) built from items' `rotor` entries;
SwiftUI `accessibilityRotor` cannot reach into separate hosting views.

## Pointer ownership

AppKit owns **all** canvas mouse sequences. Hosted SwiftUI is paint-only
(`.allowsHitTesting(false)`). Product hosts publish `hitTargets` and wire
callbacks on `GraphCanvasPointerController`.

**Derive hit rects from measured paint, not constants.** A host's rects must
come from the laid-out views they cover, or they drift as text wraps. The
Evidence graph tags actionable pieces with `.evidenceCardHitRegion(_:)`, each
card reports an `EvidenceCardLayout`, and hit targets / edge anchors are built
from it (constants only before the first report). Verify against drawn pixels,
not `fittingSize`, which over-reports wrapped text. Do not attach SwiftUI
`DragGesture` / `.onTapGesture` / `.focusable()` to cards — that fought
first-responder and cancelled mid-drag under magnification.

## Coordinate seams

| Seam | Use |
| --- | --- |
| AppKit `NSView.convert` on the document | Live pointer hit-test under magnification |
| `GraphCanvasCoordinates` | Pure layout transforms, magnification clamp, pan/zoom clamp, unit tests |

Do not treat `GraphCanvasCoordinates.contentPoint(fromViewPoint:)` as the
runtime pointer path — `NSScrollView` magnification does not map SwiftUI
gesture locations into the hosted document correctly (design note §7.2).

## Input mapping

| Device | Action |
| --- | --- |
| Trackpad two-finger scroll | Pan |
| Trackpad pinch | Zoom |
| Mouse wheel | Zoom toward cursor (manual anchor; not `centeredAt:`) |
| Click-drag on empty document | Pan (idle mode) |
| Click / drag on hit targets | Product callbacks (select, move, place, connect) |

Magnification range defaults are **zoom-out oriented** (`0.25` … `1.25`): overview of a
large graph, then back toward identity. Hosts may choose different clamps.
Deliberate zoom-*in* past ~125% is capped in the default camera so layer-scaled
SwiftUI content does not go soft.

Changing `contentID` resets magnification to `1` and re-centers the document.
Paint `rootView` may update every representable pass — safe because AppKit owns gestures.
Item views sit above it in z-order; items flagged `isRaised` (dragged, activated) sit above other items.
