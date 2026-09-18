# Spike 6 — Completed steps

Finished Spike 6 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S6-NN`, `S6-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S6-01](#s6-01--pr-panzoom-shell) | PR | `NSScrollView` Evidence graph shell + unit-tested coordinate seam |

---

## Steps

### S6-01 — PR: Pan/zoom shell

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | Spike 5 graph place |
| **Deliverables** | Done. Neutral reusable [`macos/App/Features/GraphCanvas/`](../../../macos/App/Features/GraphCanvas/) tooling (`GraphCanvasCamera`, `GraphCanvasCoordinates`, `GraphCanvasScrollView`, `GraphCanvasGridView`) plus product destination [`EvidenceGraphView`](../../../macos/App/Features/EvidenceGraph/EvidenceGraphView.swift) mounted from [`WorkspaceDestinationHost`](../../../macos/App/Features/Workspace/WorkspaceDestinationHost.swift) for `.sourceGraph`. Stub coming-soon view removed for this destination. Camera is ephemeral (in-memory / scroll view state only). |
| **Tests** | Done. [`GraphCanvasCoordinatesTests.swift`](../../../macos/ProvenenciaTests/GraphCanvasCoordinatesTests.swift): clamp bounds, 1× / 2× / 0.5× conversion, round-trip. |
| **Dogfood** | Sources list → Evidence graph → pinch zoom and scroll pan on the empty grid; no “coming soon” stub. |
| **Out** | Bubbles, palette, subject RPCs, UserDefaults camera, a11y representation tree, SemVer bump. |

**Landed:** the AppKit pan/zoom escape hatch and the pure coordinate seam Spike 6 hit-testing must reuse. Canvas geometry lives in `GraphCanvas` as reusable tooling; `EvidenceGraph` is the product view that composes it (design note §13).

**Verify:**

```bash
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/GraphCanvasCoordinatesTests
```
