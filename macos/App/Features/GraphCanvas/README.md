# GraphCanvas

Reusable spatial canvas **tooling** — pan/zoom (`NSScrollView`), content-space
coordinates, and a plain grid underlay.

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
| `GraphCanvasScrollView` | AppKit magnification bridge + input mapping |
| `GraphCanvasViewportController` | Click-drag pan + **live AppKit `convert` hit-test** |
| `GraphCanvasGridView` | Optional empty grid document content |

## Coordinate seams

| Seam | Use |
| --- | --- |
| AppKit `NSView.convert` via `contentPointUnderCursor()` | Live pointer hit-test (place / ghost / future tools) under magnification |
| `GraphCanvasCoordinates` | Pure layout transforms, magnification clamp, shared pan/zoom scroll clamp, unit tests |

Do not treat `GraphCanvasCoordinates.contentPoint(fromViewPoint:)` as the
runtime pointer path — `NSScrollView` magnification does not map SwiftUI
gesture locations into the hosted document correctly (design note §7.2).

## Input mapping

| Device | Action |
| --- | --- |
| Trackpad two-finger scroll | Pan |
| Trackpad pinch | Zoom |
| Mouse wheel | Zoom toward cursor (manual anchor; not `centeredAt:`) |
| Click-drag on empty document | Pan (product hosts wire via viewport controller) |

Magnification range is **zoom-out oriented** (`0.25` … `1.25`): overview of a
large graph, then back toward identity. Deliberate zoom-*in* past ~125% is
capped so layer-scaled SwiftUI cards do not go soft.

Changing `contentID` resets magnification to `1` and re-centers the document.
