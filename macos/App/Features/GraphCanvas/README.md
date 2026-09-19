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
| `GraphCanvasDocumentView` | Scroll document; owns mouse; hosts paint-only SwiftUI |
| `GraphCanvasPointerController` | Hit-test, select, drag, empty-canvas pan, place, connect |
| `GraphCanvasHitTarget` / `GraphCanvasPointerHitTesting` | Document rects + pure helpers |
| `GraphCanvasViewportController` | Pan deltas + live AppKit `convert` under cursor |
| `GraphCanvasGridView` | Optional empty grid document content |
| `GraphCanvasEdgeGeometry` | Cubic edges + tuck-under endpoints |

## Pointer ownership

AppKit owns **all** canvas mouse sequences. Hosted SwiftUI is paint-only
(`.allowsHitTesting(false)`). Product hosts publish `hitTargets` and wire
callbacks on `GraphCanvasPointerController`. Do not attach SwiftUI
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

Magnification range is **zoom-out oriented** (`0.25` … `1.25`): overview of a
large graph, then back toward identity. Deliberate zoom-*in* past ~125% is
capped so layer-scaled SwiftUI cards do not go soft.

Changing `contentID` resets magnification to `1` and re-centers the document.
Paint `rootView` may update every representable pass — safe because AppKit owns gestures.
