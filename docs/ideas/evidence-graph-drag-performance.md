# Evidence graph drag performance: what's left

**Status:** idea, not scheduled. Dragging is reasonable after the localization and card fixes (below); pick this up if graphs grow or drags feel choppy again. Architecture: [`macos/App/Features/GraphCanvas/README.md`](../../macos/App/Features/GraphCanvas/README.md).

## Where things stand

Profiles of dragging cards on a real graph (Time Profiler + Hitches, Debug build):

| | Main thread while dragging | Slow frames, median / worst |
| --- | --- | --- |
| Before | ~100%, 7 hangs of 250–400 ms | 173 / 413 ms |
| After localization fixes (#242–#245) | 70–88% | 27 / 93 ms |
| After card text + AppKit-owned cards (#246, #247) | 35–50% | 27 / 80 ms |

What fixed it: `String(localized: LocalizedStringResource)` re-parsed the strings table on every call (72% of drag work); the graph's document body re-ran on every pointer event because it read `GraphCanvasPointerController.offsets` (observation is per property); cards are now AppKit-positioned `NSHostingView`s that a drag moves without SwiftUI.

## What's left: edges still live in SwiftUI

The edge layer (`EvidenceGraphEdgesHost` / `EvidenceGraphEdgeLayer`) is the one SwiftUI view that follows drag offsets. Each pointer event invalidates the 4000×4000 base hosting view, which then:

- **re-measures its whole tree** in `NSHostingView.layout()` (~48% of remaining drag work);
- **redraws the full-document edge `Canvas`** (~15%), for *any* dragged card, even one with no edges;
- runs a SwiftUI transaction for it (~10%).

A standalone reproduction confirmed the mechanism: with the edge canvas watching an offset, the base hosting view lays out on every frame; with only a card view moving, it never does. Turning off the base view's `sizingOptions` did not help.

## The idea

- **Edges in Core Animation, owned by GraphCanvas.** One `CAShapeLayer` per edge in a layer above the base view and below items. On drag, update only the paths of edges attached to the dragged item, straight from `GraphCanvasDocumentView.mouseDragged` — no SwiftUI. The host pushes edge descriptions (endpoint item ids, colors / gradient stops, selected state) the way it pushes items. Gradients: a `CAGradientLayer` masked by the shape, or a solid stroke plus a short gradient cap.
- **Grid as a pattern layer.** Replace the 4000×4000 grid `Canvas` with a tiled or pattern-filled layer. Mostly a memory win (~64 MP at 2×), not a drag cost today.
- **Viewport virtualization.** With items as views, install hosting views only for items near `documentVisibleRect`. Only worth it for graphs much larger than a Source's.

## Measuring

- Attach to the running app — never `xctrace --launch` (it resolves the app by bundle id and opened two copies): `xcrun xctrace record --template 'Time Profiler' --instrument 'Hitches' --time-limit 30s --attach <pid>`.
- Leave out the SwiftUI instrument: attaching it registers every view type on the main thread and freezes the app for ~20 s.
- Profile a Release build when you can; our own view code costs less there, framework layout and rendering about the same.
