# Spike 6 — Completed steps

Finished Spike 6 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S6-NN`, `S6-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S6-D1](#s6-d1--design-canvas--primary-cards) | Design | Primary floating cards, place/create, uncited shell |
| [S6-01](#s6-01--pr-panzoom-shell) | PR | `NSScrollView` Evidence graph shell + unit-tested coordinate seam |
| [S6-02](#s6-02--pr-primary-cards--accessibility) | PR | Placed primary cards + VoiceOver representation |
| [S6-03](#s6-03--pr-click-to-place--drag--persist) | PR | Palette place/create, drag snap, position patch |
| [S6-D2](#s6-d2--design-connect--bridge-cards) | Design | Connect tool, lines, subordinate bridge cards, honesty |
| [S6-04](#s6-04--pr-connect--bridge-cards) | PR | Connect A→B, bridge create, provisional endpoints, edge draw |

---

## Steps

### S6-D1 — Design: Canvas + primary cards

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | Spike 5 Evidence graph place; S6-01 shell may already exist |
| **Deliverables** | Done. Board for Evidence graph primary surface: toggle Add Person/Event/Place tools, click-to-place + label/description create modal, floating rounded Person/Event/Place cards (icon + color tinge), and **uncited** chrome (ghost/shell/badge) for primaries with no Observations. Cited-shell contrast frame. Bridges explicitly out (never uncited). Brief archived: [`design/archive/S6-D1-canvas-bubbles.md`](design/archive/S6-D1-canvas-bubbles.md). |
| **Dogfood** | Design only — implemented in S6-02 (cards/uncited) and S6-03 (palette place/create/drag). |
| **Out** | Connect, bridge cards, cited-data rows, citation composer (S6-D2 / later). |

**Landed (design only):** primary card language and uncited vs cited-shell states for S6-02 / S6-03. Connect + bridge chrome remain **S6-D2** (citation composer is later, not D2).

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
| **Deliverables** | Done. [`SourceGraphSnapshot`](../../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift) loads placed Person/Event/Place subjects via [`CatalogQueryRegistry`](../../../macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift). [`EvidenceGraphView`](../../../macos/App/Features/EvidenceGraph/EvidenceGraphView.swift) shows floating [`EvidenceSubjectCard`](../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift)s as plain `.focusable()` views (tinge wash, uncited dashed chrome, selection + keyboard focus chrome — not Buttons) over the grid using [`GraphCanvasGridMapping`](../../../macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift). [`PVSubjectIcon`](../../../macos/App/DesignSystem/Components/Research/PVSubjectIcon.swift) + kind colors on [`PVColor`](../../../macos/App/DesignSystem/Tokens/PVColor.swift). Tab reaches cards; Space/Return activates for future inner controls; Subjects rotor. Empty state when no placed primaries. Palette tools remain real Buttons with focus rings. |
| **Tests** | Done. `SourceGraphSnapshotTests` (filter bridges/unplaced; a11y labels), `GraphCanvasGridMappingTests`, `CatalogQueryRegistryTests.sourceGraphLoadsPlacedPrimaries`. |
| **Dogfood** | Open Evidence graph for a Source that already has placed subjects. VoiceOver lists subjects. Empty graph prompts picking a tool (S6-03). |
| **Out** | Connect / bridges (S6-04); Observation-driven cited transition. |

**Landed:** first subject cards on the canvas and the accessibility representation required by design note §7.4.

**Verify:**

```bash
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourceGraphSnapshotTests \
  -only-testing:ProvenenciaTests/GraphCanvasGridMappingTests \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests
```

### S6-03 — PR: Click-to-place + drag + persist

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S6-02, S6-D1 |
| **Deliverables** | Done. Floating [`EvidenceGraphPalette`](../../../macos/App/Features/EvidenceGraph/EvidenceGraphPalette.swift) over the canvas viewport (not inside the scroll document). [`EvidenceGraphModel`](../../../macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift) arms Person/Event/Place tools, placement ghost (~62% opacity), click-to-place [`PVDialog`](../../../macos/App/DesignSystem/Components/Feedback/PVDialog.swift) (label + description), create via store + `.createdSubject` invalidate. Exclusive [`EvidenceCanvasInputMode`](../../../macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift) (idle pan/drag vs place overlay). Card drag (idle only) snaps and persists with `setSubjectPosition` + `session.setQueryValue` patch of [`SourceGraphSnapshot.updatingPosition`](../../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift); persist failure reverts + toast. Arrow-key move and Esc disarm then deselect. Cancel create disarms. |
| **Tests** | Done. [`EvidenceGraphModelTests`](../../../macos/ProvenenciaTests/EvidenceGraphModelTests.swift) (create+position, cancel disarms, input mode, persist failure revert+toast, position patch, `.createdSubject` invalidation); snapshot + registry coverage extended. |
| **Dogfood** | Arm Add Person → click grid → label → confirm; drag cards; relaunch — positions stick. Cancel and Esc disarm; confirm disarms. |
| **Out** | Unplaced tray; connect / bridges (S6-04); delete UX; Observation `isCited` wiring; keyboard shortcuts to arm tools. |

**Landed:** dogfoodable create + rearrange on the Evidence graph without per-move query invalidation.

**Verify:**

```bash
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/EvidenceGraphModelTests \
  -only-testing:ProvenenciaTests/SourceGraphSnapshotTests \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests
```

### S6-D2 — Design: Connect + bridge cards

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | S6-D1 |
| **Deliverables** | Done. Board for Connect tool (hairline-separated from Add tools), two-click A→B, bridge create (label + description only), subordinate bridge cards (188px, no pigment, honesty body), cubic A—bridge—B lines meeting card edges, selected-bridge thickens both segments. No citation / Source-viewer modal. Brief archived: [`design/archive/S6-D2-connect-edges.md`](design/archive/S6-D2-connect-edges.md). |
| **Dogfood** | Design only — implemented in S6-04. |
| **Out** | Citation composer, Observations, catalog edge migration, person→person shared-event macro. |

**Landed (design only):** connect + bridge chrome language for S6-04.

### S6-04 — PR: Connect + bridge cards

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S6-03, S6-D2 |
| **Deliverables** | Done. Connect palette tool + [`EvidenceCanvasInputMode.connecting`](../../../macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift). Two-click A→B with kind inference (`participation` / `location` / `relationship`); create sheet reuses D1 label/description; bridge subject + position via existing store APIs; provisional endpoints in [`EvidenceProvisionalLinkStore`](../../../macos/App/Features/EvidenceGraph/EvidenceProvisionalLinkStore.swift) (Application Support). [`SourceGraphSnapshot`](../../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift) includes bridges; [`EvidenceBridgeCard`](../../../macos/App/Features/EvidenceGraph/EvidenceBridgeCard.swift) (188px, honesty copy); [`EvidenceGraphEdgeLayer`](../../../macos/App/Features/EvidenceGraph/EvidenceGraphEdgeLayer.swift) + [`GraphCanvasEdgeGeometry`](../../../macos/App/Features/GraphCanvas/GraphCanvasEdgeGeometry.swift). Links VoiceOver rotor; keyboard Connect pick via Space/Return. |
| **Tests** | Done. Model connect/cancel/invalid-pair; snapshot bridges + inference; edge geometry; provisional link file round-trip; registry includes placed bridges. |
| **Dogfood** | Place Person + Event → Connect → pick both → label → save; lines through mid-card; drag endpoints (lines follow); relaunch keeps bridge subject + provisional link. Esc / cancel create leaves A held. |
| **Out** | Citation modal, Observations, edge migration, shared-event macro, fancy routers. |

**Landed:** connect gesture and bridge chrome dogfood without pretending links are cited evidence.

**Verify:**

```bash
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/EvidenceGraphModelTests \
  -only-testing:ProvenenciaTests/SourceGraphSnapshotTests \
  -only-testing:ProvenenciaTests/GraphCanvasEdgeGeometryTests \
  -only-testing:ProvenenciaTests/EvidenceProvisionalLinkStoreTests \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests
```
