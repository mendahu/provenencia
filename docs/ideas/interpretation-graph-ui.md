# Interpretation graph UI (brainstorm)

## Status

**Brainstorm.** Not scheduled, not authoritative, and not a UI spec. Captured before the Interpretation-layer spike is planned so the shape of the client can be argued about before PRs are broken out.

Authoritative schema for everything described here is [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md). Client rules are [`macos-client-patterns.md`](../macos-client-patterns.md). Nothing in this file overrides those.

---

# 1. The premise

The Interpretation layer is Citation → Observation → Node. Nodes are candidate persons, events, places, and the bridge-like associations between them (`participation`, `location`, `relationship`), plus reified `source` Nodes.

The default way to build a client for that is one destination per table: a Citations list and detail, an Observations list and detail, a Nodes list and detail. That is how the Source layer is built today (Sources, Source types, Source fields), and it works there because a researcher genuinely does think about one Source at a time.

The claim behind this brainstorm is that it will **not** work here, because the unit of thought in interpretation is not one row. Reading a single census line produces a person, a household, several relationships, a residence, an occupation, and a birth-year inference — all at once, all from one locator, all held in mind together. A UI that makes the researcher navigate between three list/detail pairs to record that is charging a navigation tax on every fact.

## 1.1 The workflow the UI should match

Interpretation is *mostly* source-scoped: `nodes.source_id` is the Node's home Source, and a work session is normally "sit with one Source and map what it appears to say."

```text
open a Source
  → look at its Artifacts
      → mark what a portion of an Artifact says      (Citation)
      → say what that means                          (Observations)
          → about candidate things                   (Nodes)
      → repeat until the Source is exhausted
```

So the natural container is **one Source**, and the natural artifact of a work session is **a small graph** — typically tens of Nodes, not thousands.

## 1.2 The proposed shape

A single spatial workspace, scoped to a Source, where the researcher:

- places Nodes (person, event, place, …) on a grid;
- draws connections between them to assert associations;
- clicks a placed Node to add properties to it;
- and, on every assertion, is prompted with a **citation composer** that shows the Artifact, captures the locator, captures the transcription, and records the Observation.

Plus one supporting destination: an **Interpretation vocabulary** browser for `node_types`, `properties`, and the `node_type_properties` bindings — which Properties exist and which Node Types they attach to. That is the direct analogue of Source types / Source fields and should reuse [`Features/CatalogVocabulary/`](../../macos/App/Features/CatalogVocabulary/).

---

# 2. What is genuinely new here

The codebase has **no prior art for either half of this**. Worth stating plainly before estimating:

| Piece | Current state |
| --- | --- |
| Spatial canvas, drag, connect, pan/zoom | Nothing. No `Canvas`, no custom gesture surface. `PVReorderableList` drag is the most complex gesture shipped. |
| Artifact rendering | Nothing. No PDFKit, no QuickLook, no AVKit. Files/Artifacts have thumbnails ([`core/derivatives`](../../core/derivatives/)) and a list row; nobody has ever *opened* an Artifact in the app. |
| Locator capture | Nothing. The selector vocabulary (`page`, `region`, `text_quote`, `time_range`) is fully specified in the data model and entirely unimplemented. |
| Interpretation vocabulary browser | Reusable shell exists (`CatalogVocabulary`, `PVTable`, origin markers). |
| Interpretation schema / Go / FFI | Nothing. Migrations stop at `000020.sql`; no `core/database/nodes`, `citations`, `observations`. |

The vocabulary browser is a known quantity. The canvas and the artifact viewer are both greenfield, and the artifact viewer is probably the *larger* of the two.

---

# 3. The central design problem: bridge Nodes

This is the thing most likely to sink a naive canvas.

The visual model the researcher wants is:

```text
[Person: Wm Robins] ────father of───▶ [Person: John Robins]
```

The data model deliberately does not store that. It stores:

```text
Event E1        -- event_type   --> "birth"
Participation P1 -- person      --> Node(John)
Participation P1 -- event       --> Node(E1)
Participation P1 -- role        --> "subject"
Participation P2 -- person      --> Node(Wm)
Participation P2 -- event       --> Node(E1)
Participation P2 -- role        --> "father"
```

…and expects the application to *project* "father of" from it (interpretation layer §1.4, §6). Every one of those Observations needs a Citation.

So a drawn line is not a row. **A drawn line is a macro** that mints a bridge Node plus several Observations, all bound to one Citation. Two consequences:

1. **Forward direction (draw → write).** The canvas needs a small, explicit library of gestures-to-subgraph macros, and it has to be honest about what it wrote. "Connect person to event" is a different macro from "connect person to person," and the second one has to ask *which* event/relationship mediates it.
2. **Reverse direction (read → draw).** Rendering an arbitrary stored subgraph back into clean lines is lossy and sometimes impossible. Real data that has no tidy visual form:
   - a `participation` with a `person` but no `event` yet (valid — Nodes are sparse and Observations accumulate);
   - two competing `birth_date` Observations on one Node from two Citations (expected, not an error);
   - a **negative-polarity** edge — "this source denies this person was at this event" — which is not the absence of a line;
   - the same association asserted by three different Citations with different transcriptions.

A canvas that can only draw the tidy cases will silently hide the interesting ones, which is the opposite of what an evidence-first tool is for. Mitigation sketch: show bridge Nodes as **collapsible** — a clean line by default, expandable into its real bridge-Node subgraph — and give every drawn element a visual state for *conflicted*, *negated*, and *incomplete*. Do not let the canvas be the only way to see a Node's Observations.

---

# 4. Where do coordinates live?

The Interpretation schema has no `x`/`y` and should not get them. Node position is not evidence.

But position is also not throwaway: spending twenty minutes laying out a census household is real intellectual work, and it must survive quit, relaunch, and ideally project sharing ([`ideas/share-packages.md`](share-packages.md)). Options:

| Option | Notes |
| --- | --- |
| Pure auto-layout, no stored positions | Cheapest. Deterministic layout from graph shape. But researchers will rearrange to mean something ("spouses side by side"), and losing that will feel like data loss. |
| Separate catalog table (e.g. `graph_layouts`) | Travels with the project, shares, and syncs. Costs a migration and a decision about whether it is audited (probably not — it is not a research assertion). Pollutes the catalog with UI state. |
| Sidecar file in the project folder | Keeps the catalog clean. Needs its own format, versioning, and conflict story; easy to lose or desync from the catalog. |

Leaning: a catalog table, explicitly excluded from audit and from the schema-hash's "research data" story, with auto-layout as the fallback for Nodes that have no stored position (imports, cross-source Nodes, anything created outside the canvas). Auto-layout is needed **regardless** of which option wins, so it is not optional work.

Either way, dragging a Node must not write per frame. Positions batch and debounce; the catalog session serializes FFI ([`use-catalog-session`](../../.cursor/skills/use-catalog-session/SKILL.md)), so a chatty canvas would queue behind badge refreshes and list loads.

---

# 5. The citation composer

Every property edit on the canvas opens this, so it is on the hot path of the whole layer.

```text
Artifact viewer            Locator                  Interpretation
─────────────────          ─────────────────        ─────────────────
PDF page / image /         page + page_label        subject Node
audio / video              region polygon           Property
                           text_quote               polarity
                           time_range               typed value
                                                    (text/int/real/bool/
transcription                                        date/name/node)
transcription_uncertain
description
```

macOS building blocks, per media type:

- **PDF** — `PDFKit.PDFView` wrapped in `NSViewRepresentable`. Gives page navigation (feeding `artifact_page`), and `PDFSelection` gives selected text plus surrounding context, which is exactly `text_quote.exact` / `prefix` / `suffix`. `PDFPage` coordinate conversion supports normalized region points.
- **Image** — `NSImage` plus a custom overlay for polygon drawing.
- **Audio/video** — AVKit `AVPlayer` with a time observer for `time_range` (`start_ms`/`end_ms`).
- **QuickLook** (`QLPreviewView`) is tempting because it is nearly free and handles every format — but it exposes no selection or coordinate API, so it cannot produce locators. Useful as a read-only preview, not as the composer.

Region polygons need a custom drag overlay producing normalized points, and the locator invariants are strict (≥3 distinct points, non-self-intersecting, non-zero area, in-bounds). **Validate locator JSON in Go**, not Swift, so a future Windows client inherits it and so `locator_json` can never be written malformed.

Placement question: a separate window, a sheet, or macOS 14's `Inspector`? The workflow is "look at the source while you type," which argues for side-by-side and against a modal sheet. A sheet that covers the graph you are annotating is the wrong default.

---

# 6. macOS feasibility

Deployment target is macOS 14.0 (CI Xcode 16), which constrains some of the newest SwiftUI affordances.

**Rendering.** The standard pattern for this kind of view is a hybrid: SwiftUI `Canvas` for the grid and the edges (immediate-mode, GPU-backed, cheap for hundreds of lines), with real SwiftUI views for the Nodes in a `ZStack`, positioned with `.position(_:)`. `Canvas` cannot host subviews, take gestures per-shape, or expose accessibility children — so Nodes must be real views to get hit-testing, text fields, focus, and VoiceOver. That hybrid comfortably handles the tens-to-low-hundreds of Nodes a single Source produces. It would not handle a whole-project graph, which is a further reason to keep this Source-scoped.

**Pan and zoom.** This is the roughest edge. SwiftUI has no built-in two-axis zoomable scroll view on macOS 14 (programmatic scroll and zoom affordances largely land in macOS 15). Realistic choices: hold `offset`/`scale` in state and drive them with `MagnifyGesture` (macOS 14+) plus scroll-wheel events, or wrap `NSScrollView` in an `NSViewRepresentable` and get native scrolling, elastic bounds, and magnification for free. The AppKit wrapper is the more Mac-native result and is exactly the "escape hatch, not the default" case in [`macos-client-patterns.md`](../macos-client-patterns.md) §4.

**Drag-to-connect.** A `DragGesture` on a connection handle, tracked in a named coordinate space, hit-testing the target Node on end. Mechanically fine; the hard part is the macro semantics in §3, not the gesture.

**Keep the geometry pure.** Grid snapping, hit-testing, auto-layout, and edge routing should be free functions over value types, unit-testable in `ProvenenciaTests` with no view and no store. That is the only way any of this gets test coverage — see below.

**Accessibility.** A free-form spatial canvas is genuinely hostile to VoiceOver and keyboard-only use, and this is the biggest honest cost of the approach. It needs: full keyboard parity for every canvas action, and a structured outline of the same graph that serves as the accessibility representation. Which leads to the next point.

**Testing.** XCUITest cannot meaningfully drive a canvas (it finds controls by accessibility, not pixels). Coverage has to come from elsewhere: Go tests for schema, macros, and locator validation; Swift unit tests for pure geometry and for the graph model; and a structured (non-canvas) editing path that *is* testable. Budget for the canvas itself being verified by hand.

---

# 7. Pros of the graph approach

- **It matches the unit of thought.** One Source, one map, one session. No bouncing between three list/detail pairs to record one line of a census.
- **It makes bridge Nodes tolerable.** `participation`, `location`, and `relationship` are the price of a generic property graph, and they are miserable to fill in through forms. As an edge macro they become nearly invisible — the researcher draws a line and the app does the reification.
- **It is naturally bounded.** Source-scoped graphs stay small. This is the reason it can work where "show me my whole family tree as a graph" does not.
- **Completeness becomes visible.** A sparse Node or a dangling bridge is obvious on a canvas and invisible in a list. The map doubles as a QA view for the interpretation of that Source.
- **Citation stays on the hot path.** If every property edit routes through the composer, "cited by construction" becomes the path of least resistance rather than a discipline the researcher has to maintain.
- **It is the product differentiator.** Evidence-first genealogy tools are mostly form-driven. This is the part of Provenencia that would not be a reimplementation of something that exists.

---

# 8. Cons and risks

- **Highest-risk UI in the app, built on zero prior art**, and it lands at the same time as the layer's schema, Go, and FFI. Two unknowns multiplied.
- **The canvas cannot express everything.** Polarity, competing Observations, transcription uncertainty, and cross-source Observations all resist spatial representation. Some list/detail surface is still required, so "graph instead of pages" is really "graph *plus* fewer pages."
- **Round-trip lossiness** (§3). Draw → subgraph is a choice the app makes on the researcher's behalf; subgraph → draw cannot always be clean. Both directions need explicit, documented rules or the canvas will quietly misrepresent stored evidence.
- **Position becomes state we did not want** (§4), with migration, sharing, and sync consequences.
- **Accessibility and keyboard parity are real, non-optional work**, not a polish pass.
- **It is nearly untestable by our normal ladder** (§6), on the most semantically dangerous writes in the product — multi-row transactions that mint Nodes and Observations.
- **It fights the current session cache.** `WorkspaceSession` patch/invalidate is tuned for lists and detail payloads ([`macos-client-patterns.md`](../macos-client-patterns.md) §1). A graph is one large payload mutated constantly from within the view; naive invalidation would reload the whole graph on every keystroke-ish write, and naive patching would drift from the catalog. This needs a deliberate answer, probably per-Node and per-Observation patching against a graph query key.
- **Scope creep is the default outcome.** "Snap to grid, then auto-layout, then edge routing, then minimap, then multi-select, then alignment guides" is an infinite backlog that produces no genealogical capability.
- **Wrong-workflow risk.** The premise is map-first (place Nodes, then cite). For record-type sources — census rows, registers, certificates — the real flow may be citation-first: transcribe one locator, then spit out the eight facts it contains. If citation-first dominates in practice, the canvas is the wrong primary surface and an "extract from this citation" form is the right one. Worth testing on real sources *before* building the canvas.

---

# 9. Hiccups checklist

Things that will each need an answer, roughly in the order they will bite:

1. **Edge macros.** Which gestures exist, what subgraph each writes, how the researcher sees what was written, and how it is undone.
2. **Reverse rendering rules.** Collapse/expand of bridge Nodes; visual states for negated, conflicted, and incomplete.
3. **Layout persistence.** Table vs sidecar vs pure auto-layout; audited or not; auto-layout for unpositioned Nodes.
4. **Composer surface.** Window vs inspector vs sheet, and whether it stays pinned across successive edits.
5. **Artifact viewer scope.** PDF and image only for v1? Audio/video locators are specified but need AVKit and their own UI.
6. **Locator validation in Go**, with the full invariant set, before any UI can write `locator_json`.
7. **Cross-source Nodes.** A Node homed to another Source can legitimately appear in this graph (Observations are not confined by `source_id`). Ghost styling, and can you edit it here?
8. **Cache strategy** for a large, constantly mutated graph payload.
9. **Navigation history.** The graph is a place ([`add-workspace-location`](../../.cursor/skills/add-workspace-location/SKILL.md)); camera position and selection probably are *not*. `WorkspaceLocation` is a flat struct of optional ids and will need new fields plus an identity decision.
10. **Undo.** Drawing one line writes several rows; researchers will expect ⌘Z. The audit model already says undo is a forward revision, not a deletion ([`audit-revision-history.md`](../audit-revision-history.md) §9) — so undo is a Go concern, not an `UndoManager` concern.
11. **Density and filtering.** A census page can yield dozens of Nodes and hundreds of Observations. Needs layers/filters (persons only, collapse bridges) before it is usable on a real source.
12. **Design system.** Canvas chrome, Node cards, edge styling, and selection need to come from existing tokens rather than a parallel visual language.
13. **Localization.** Less text than a form-heavy screen, but Node cards, macro menus, and every warning state still go through `L10n`.
14. **Vocabulary dependency.** The graph is unusable until `node_types`, `properties`, and `node_type_properties` are seeded and browsable — so the vocabulary destination is a *prerequisite*, not a companion.

---

# 10. Sequencing instinct

Not a PR plan — just the dependency order this brainstorm implies, and where the de-risking belongs.

```text
schema + Go + FFI              ← no UI risk, pure data model work
  → Interpretation vocabulary  ← known pattern (CatalogVocabulary), unblocks everything
    → artifact viewer          ← greenfield; PDFKit; largest single unknown
      → citation composer      ← locator capture + transcription
        → structured editing   ← testable path; also the a11y representation
          → graph canvas       ← highest risk, built last, on a working layer
```

The important property of that order: **the layer is shippable and dogfoodable before the canvas exists.** If the canvas slips or proves to be the wrong idea after real use, the spike still delivered the Interpretation layer. If the canvas is built first and the layer is built around it, there is no fallback.

The counter-argument is that a structured editing path built first may be "good enough" and the canvas never happens. That is a real risk and worth naming — but it is also the honest test of whether the canvas is the differentiator this note claims it is.

---

# 11. Naming

**Decided: Interpretation graph.** It matches the layer name and the data model's own language — the Interpretation layer *is* "a cited property graph" ([`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §1.2), so the UI name and the schema vocabulary agree instead of introducing a third word for the same thing. Layer vocabulary stays intact inside it: Citations, Observations, and Nodes keep their names.

Considered and set aside:

| Candidate | Why not |
| --- | --- |
| Interpretation map | "Map" reads well but adds a term the data model does not use. |
| Source map | Emphasizes Source scope, but sounds geographic — awkward alongside `place` Nodes. |
| Evidence map | Legible to non-experts, but "evidence" sits closer to the Source layer in our vocabulary. |
| Worksheet | Plays down the visual and plays up the work session; too humble for the layer's primary surface. |

---

# 12. Open questions

- Is the real workflow map-first or citation-first? Testable now, on a real census page, before any canvas code exists.
- Does the canvas *create* Nodes, or only arrange Nodes created elsewhere?
- Can one graph span Sources (a "case view"), or is Source scope hard?
- Are `source` Nodes and source-to-source `mentions` edges on this canvas, or is that a different view?
- How does Conclusion-layer work ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md)) surface here later — same canvas with a layer toggle, or a separate reconciliation view? Deciding "not now" is fine; deciding "never" would be a mistake.
