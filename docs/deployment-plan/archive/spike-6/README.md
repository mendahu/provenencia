# Spike 6 — Evidence graph canvas

**Done / Go.** The spatial canvas is buildable on this stack. Geometry reuse: [`macos/App/Features/GraphCanvas/README.md`](../../../../macos/App/Features/GraphCanvas/README.md). Themes: [`interpretation-graph-ui.md`](../../../ideas/archive/interpretation-graph-ui.md).

## Decisions

- **No schema work.** Spike 5 already shipped subjects, positions, and the graph place.
- Pan/zoom is **AppKit `NSScrollView`**. Raising macOS 14 is a support-matrix decision, not a canvas requirement.
- Primaries (person / event / place) may be **uncited** (label only). Bridges from Connect were **provisional** in this spike — honesty labeled; durable cite is Spike 7.
- Geometry lives in a **neutral `GraphCanvas` module** (family-tree reuse later). Position tables stay concrete per layer, not polymorphic.
- Auto-layout, minimap, multi-select, and edge-routing polish stayed out. Unplaced tray was later **descoped** (Spike 8); create always writes a position.
