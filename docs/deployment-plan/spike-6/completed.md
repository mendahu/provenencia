# Spike 6 — Completed steps

Finished Spike 6 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S6-NN`, `S6-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S6-D1](#s6-d1--design-canvas--primary-cards) | Design | Primary floating cards, place/create, uncited shell |
| [S6-01](#s6-01--pr-panzoom-shell) | PR | `NSScrollView` Evidence graph shell + unit-tested coordinate seam |
| [S6-02](#s6-02--pr-primary-cards--accessibility) | PR | Placed primary cards + VoiceOver representation |

---

## Steps

### S6-D1 — Design: Canvas + primary cards

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | Spike 5 Evidence graph place; S6-01 shell may already exist |
| **Deliverables** | Done. Board for Evidence graph primary surface: toggle Add Person/Event/Place tools, click-to-place + label/description create modal, floating rounded Person/Event/Place cards (icon + color tinge), and **uncited** chrome (ghost/shell/badge) for primaries with no Observations. Cited-shell contrast frame. Bridges explicitly out (never uncited). Brief archived: [`design/archive/S6-D1-canvas-bubbles.md`](design/archive/S6-D1-canvas-bubbles.md). |
| **Dogfood** | Design only — implemented in S6-02 (cards/uncited); place/create tools still S6-03. |
| **Out** | Connect, bridge cards, cited-data rows, citation composer (S6-D2 / later). |

**Landed (design only):** primary card language and uncited vs cited-shell states for S6-02 / S6-03. Connect and shared citation handoff remain **S6-D2**.

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

### S6-02 — PR: Primary cards + accessibility

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S6-01, S6-D1 |
| **Deliverables** | Done. [`SourceGraphSnapshot`](../../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift) loads placed Person/Event/Place subjects via [`CatalogQueryRegistry`](../../../macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift). [`EvidenceGraphView`](../../../macos/App/Features/EvidenceGraph/EvidenceGraphView.swift) shows floating [`EvidenceSubjectCard`](../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift)s (tinge wash, uncited dashed chrome, selection) over the grid using [`GraphCanvasGridMapping`](../../../macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift). [`PVSubjectIcon`](../../../macos/App/DesignSystem/Components/Research/PVSubjectIcon.swift) + kind colors on [`PVColor`](../../../macos/App/DesignSystem/Tokens/PVColor.swift). `accessibilityRepresentation` list + Subjects rotor. Empty state when no placed primaries. |
| **Tests** | Done. `SourceGraphSnapshotTests` (filter bridges/unplaced; a11y labels), `GraphCanvasGridMappingTests`, `CatalogQueryRegistryTests.sourceGraphLoadsPlacedPrimaries`. |
| **Dogfood** | Open Evidence graph for a Source that already has placed subjects (seed via FakeStore / FFI until S6-03 create). VoiceOver lists subjects. Empty graph shows “No subjects yet.” |
| **Out** | Palette, create modal, drag/persist (S6-03); connect / bridges (S6-04); Observation-driven cited transition. |

**Landed:** first subject cards on the canvas and the accessibility representation required by design note §7.4. Create/drag remain S6-03.

**Verify:**

```bash
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourceGraphSnapshotTests \
  -only-testing:ProvenenciaTests/GraphCanvasGridMappingTests \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests
```
