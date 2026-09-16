# Interpretation graph (brainstorm)

## Status

**Brainstorm.** Not scheduled, not authoritative, and not a UI spec. Captured before the Interpretation-layer spike is planned so the shape of the client can be argued about before PRs are broken out.

Authoritative schema for everything described here is [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md). Client rules are [`macos-client-patterns.md`](../macos-client-patterns.md). Nothing in this file overrides those.

---

# 1. The premise

The Interpretation layer is Citation → Observation → Node. Nodes are candidate persons, events, places, and the bridge-like associations between them (`participation`, `location`, `relationship`), plus reified `source` Nodes.

The default way to build a client for that is one destination per table: a Citations list and detail, an Observations list and detail, a Nodes list and detail. That is how the Source layer is built today (Sources, Source types, Source fields), and it works there because a researcher genuinely does think about one Source at a time.

That will **not** work here, because the unit of thought in interpretation is not one row. Reading a single census line produces a person, a household, several relationships, a residence, an occupation, and a birth-year inference — all at once, all from one locator, all held in mind together. A UI that makes the researcher navigate between three list/detail pairs to record that is charging a navigation tax on every fact.

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

A single spatial workspace, scoped to a Source, with a small tool palette:

| Tool | Gesture | Result |
| --- | --- | --- |
| Add Person / Add Event / Add Place | click an empty grid cell | a bubble appears; a Node row is written |
| Connect | click Node A, then Node B | a line with a **bubble in the middle** — the bridge Node — plus the edges that bind it |
| Add Property | select any bubble (root or bridge) | citation composer opens: pick the locator, transcribe, fill the Observation, save; the property appears on the bubble |

Bridge bubbles are visually differentiated from root bubbles, so the association is a *thing* on the canvas the way it is a row in the catalog — but it reads as "the relationship between these two" rather than as a mysterious extra entity.

Plus one supporting destination: an **Interpretation vocabulary** browser for `node_types`, `properties`, and the `node_type_properties` bindings — which Properties exist and which Node Types they attach to. That is the direct analogue of Source types / Source fields and should reuse [`Features/CatalogVocabulary/`](../../macos/App/Features/CatalogVocabulary/).

---

# 2. What is genuinely new here

The codebase has **no prior art for either half of this**. Worth stating plainly before estimating:

| Piece | Current state |
| --- | --- |
| Spatial canvas, drag, connect, pan/zoom | Nothing. No `Canvas`, no custom gesture surface. `PVReorderableList` drag is the most complex gesture shipped. |
| Artifact rendering | Nothing. No PDFKit, no QuickLook, no AVKit. Files/Artifacts have thumbnails ([`core/derivatives`](../../core/derivatives/)) and a list row; nobody has ever *opened* an Artifact in the app. |
| Locator capture | Nothing. The selector vocabulary (`page`, `region`, `text_quote`, `time_range`) is fully specified and entirely unimplemented. |
| Structured NameValue | **Nothing, in either language.** `name_values` has no migration, no Go package, and no Swift editor — see §4.4. |
| Structured DateValue | Partial: [`Features/Dates/`](../../macos/App/Features/Dates/) has `DateValueDraft` + editor form; [`core/database/datevalues`](../../core/database/datevalues/) exists. |
| Interpretation vocabulary browser | Reusable shell exists (`CatalogVocabulary`, `PVTable`, origin markers). |
| Interpretation schema / Go / FFI | Nothing. Migrations stop at `000020.sql`; no `nodes`, `citations`, `observations` packages. |

The vocabulary browser is a known quantity. The canvas and the artifact viewer are both greenfield, and the artifact viewer is probably the *larger* of the two.

---

# 3. The graph interaction model

## 3.1 A drawn line is a macro, not a row

The visual model is `[Wm Robins] ──father of──▶ [John Robins]`. The data model deliberately does not store that. It stores a birth Event, two `participation` bridge Nodes, and six Observations, and expects the app to *project* "father of" from them (interpretation layer §1.4, §6).

So the connect tool writes a subgraph. The good news is that the "bubble in the middle of the line" design maps onto the schema almost exactly:

```text
  [Person A]  ──────  ( relationship )  ──────  [Person B]
       │                     │                       │
   free Node            free Node               free Node
                             │
              participant edges = Observations
                   └── each needs a Citation
```

**The bubble is free; the line segments cost a Citation.** A Node — including a bridge Node — needs nothing but a Source and a Node Type (§4.1). The *edges* radiating from it are Observations, and every Observation requires a Citation. That single asymmetry drives most of §4.

## 3.2 Connect is a macro with a required disambiguation step

Type-pair inference works, and the pairs match the seeded `node_type_properties` bindings — but it is not silent, because every one of these macros needs a value the app cannot guess:

| From → To | Bridge Node | Edges written | What the app must ask |
| --- | --- | --- | --- |
| person → event | `participation` | `person`, `event` | `role` — subject, father, witness, … |
| person → person | `relationship` | `participant` ×2 | `relationship_type` — and *whether* it is a direct relationship at all (see below) |
| event → place | `location` | `event`, `place` | nothing; clean |
| person → place | *none seeded* | — | residence is `person → event(residence) → place`; offer a two-hop macro or refuse |
| event → event, place → place | *none seeded* | — | refuse |

The person → person case is genuinely ambiguous: two people in one record might be in a direct `relationship` ("cousins"), or might simply both participate in one Event ("both at this wedding"), which is a pair of `participation` Nodes instead. The connect tool has to offer that choice rather than pick.

This is fine, because the gesture already opens a form (§6). The form carries the role/type picker. **The rule is: no macro silently invents vocabulary.**

## 3.3 Reverse rendering is the harder direction

Draw → subgraph is a choice the app makes. Subgraph → draw cannot always be clean. Real, legal data with no tidy visual form:

- a `participation` with a `person` but no `event` yet — valid, because Nodes are sparse and Observations accumulate;
- two competing `birth_date` Observations on one Node from two Citations — expected, not an error;
- a **negative-polarity** edge — "this source denies he was there" — which is emphatically not the absence of a line;
- the same association asserted by three Citations with different transcriptions.

A canvas that can only draw the tidy cases will silently hide the interesting ones, which is the opposite of what an evidence-first tool is for. So every drawn element needs a visual state for **conflicted**, **negated**, and **incomplete**, and the canvas must never be the only way to see a Node's Observations.

---

# 4. Creation order, lifecycle, and deletion

This section answers the "do we have to create one thing before another, and are we trapping ourselves" question.

## 4.1 What the foreign keys actually require

```text
sources ──▶ artifacts ──▶ citations ──▶ observations
                                            ▲
node_types ──▶ nodes ───────────────────────┤
properties ─────────────────────────────────┘
```

- **`nodes` is free.** It needs `source_id` (the Source whose graph is open) and `node_type_id` (seeded vocabulary). `label` and `description` are nullable. Nothing else. Click the grid, get a row.
- **`citations` needs an Artifact.** `artifact_id` is `NOT NULL`.
- **`observations` needs a Citation, a Node, and a Property**, plus exactly one typed value.

So the only hard ordering is **Artifact → Citation → Observation**, and **Node → Observation**. Nodes and Citations are independent of each other; either can come first.

**The gate worth noticing:** a Source with no Artifacts can hold Nodes but cannot hold a single Observation. The graph would be a board of uncited bubbles. That needs a deliberate empty state ("add a file to this Source before you can cite it"), not a mystery.

## 4.2 Draft status needs no schema flag

A Node with zero Observations is **already valid data** — the data model explicitly allows sparse Nodes ("creating a `person` Node does not require knowing a name"). So "uncommitted" is not a column; it is a query: *Nodes with no Observations where they are the subject.*

That gives a clean lifecycle rule:

| Node kind | Stands alone? | Uncited state |
| --- | --- | --- |
| Root — `person`, `event`, `place`, `source` | **Yes.** Created freely, one at a time. | Normal and expected. Persist immediately. |
| Bridge — `relationship`, `participation`, `location` | Legal, but semantically empty | **Never persist alone.** Commit the bridge Node, its edges, and the Citation in one transaction, or write nothing. |

The reason bridges differ: if we persisted a bridge Node and the researcher then cancelled the citation dialog, the canvas would be showing line segments that do not exist in the catalog. The connect gesture must be atomic. Root placement does not have that problem, because a lone person bubble is truthful — it says "this source seems to mention somebody," which is exactly what it means.

**`nodes.label` is the canvas's free working handle** — a nullable, non-evidentiary string. So a researcher can drop a bubble and type "head of household, line 14" with no citation at all. The UI has to keep that visually distinct from an asserted `name` Observation, or the distinction the whole layer exists to preserve gets blurred at the point of entry.

## 4.3 The real traps

1. **`node_type_id` is immutable after insert.** Dropped a Person and meant an Event? There is no UPDATE — correcting a type means a new Node with a new `ref`. So either pick the type before placing (the tool palette already does this), or offer "change type" *only* while the Node has no Observations, implemented as delete-and-recreate. Refs are cheap to burn: [`core/ref`](../../core/ref/) mints random Crockford tokens from `crypto/rand`, not a sequence, so a discarded ref leaves no visible gap.
2. **Deletion is blocked by the database, and foreign keys are enforced.** [`core/database/catalog.go`](../../core/database/catalog.go) opens with `_foreign_keys=1`, and the spec'd `observations` → `nodes` and `observations` → `citations` references carry no `ON DELETE` clause — so the default `NO ACTION` makes deleting a referenced Node *fail*. Worse, a Node can be referenced as the **subject** (`subject_node_id`) or as the **object** of an edge (`value_node_id`), so "what dies with this Node" is two queries, not one. The canvas needs an app-side cascade in a transaction plus a confirmation that counts the damage ("removes 7 Observations across 3 Citations"). Whether to write explicit `RESTRICT` or `CASCADE` is a decision to make *at migration time*, not after the UI exists.
3. **Every Property value type needs an editor.** `node_type_properties` drives the Add-Property list, and the value editor depends on `properties.value_type`: `text`, `integer`, `real`, `boolean`, `date`, `name`, `node`. That is seven editors, and `node` means a Node picker scoped to the graph.

## 4.4 The hidden dependency: NameValue

`name` is the single most common Observation a genealogist will ever record, and **the structured NameValue model does not exist in any layer** — no migration, no `core/database/namevalues`, no Swift editor. [`structured-name-model.md`](../structured-name-model.md) specifies it (a required full `form` plus optional ordered parts with an open part-type vocabulary), but nothing is built.

DateValue is in much better shape, which makes it easy to assume names are too. They are not. Building NameValue end to end is its own chunk of work sitting directly on the critical path of "record a person's name," and it should be planned explicitly rather than discovered in the middle of a slice.

---

# 5. Layout state

The Interpretation schema has no `x`/`y` and should not get them — position is not evidence. But it is not throwaway either: twenty minutes arranging a census household is real intellectual work that must survive relaunch and ideally project sharing ([`ideas/share-packages.md`](share-packages.md)).

A separate catalog table, explicitly excluded from audit, is the leaning: it travels with the project, shares, and syncs, at the cost of one migration and some UI state in the catalog.

Two simplifications worth taking:

- **Snap-to-grid means integers.** If the grid is the interaction model, make it the *storage* model: `grid_x` / `grid_y` as `INTEGER` cell coordinates rather than floats. No float drift, trivial equality, cheap conflict resolution, and snapping stops being a separate feature.
- **A tray beats an auto-layout engine.** Something has to handle Nodes with no stored position — cross-source Nodes, future imports, anything created outside the canvas. Instead of building auto-layout for v1, put unplaced Nodes in a **tray along the edge of the canvas** and let the researcher drag them onto the grid. That removes an entire algorithmic dependency and is arguably better behavior: the app never guesses at an arrangement that means something.

Dragging must not write per frame. Positions batch and debounce; the catalog session serializes FFI ([`use-catalog-session`](../../.cursor/skills/use-catalog-session/SKILL.md)), so a chatty canvas would queue behind badge refreshes and list loads.

---

# 6. The citation composer

Every property edit opens this, so it is on the hot path of the whole layer.

```text
Artifact viewer            Locator                  Interpretation
─────────────────          ─────────────────        ─────────────────
PDF page / image /         page + page_label        subject Node
audio / video              region polygon           Property
                           text_quote               polarity
                           time_range               typed value
transcription                                       (text/int/real/bool/
transcription_uncertain                              date/name/node)
description
```

macOS building blocks, per media type:

- **PDF** — `PDFKit.PDFView` wrapped in `NSViewRepresentable`. Gives page navigation (feeding `artifact_page`), and `PDFSelection` gives selected text plus surrounding context, which is exactly `text_quote.exact` / `prefix` / `suffix`. `PDFPage` coordinate conversion supports normalized region points.
- **Image** — `NSImage` plus a custom overlay for polygon drawing.
- **Audio/video** — AVKit `AVPlayer` with a time observer for `time_range`.
- **QuickLook** (`QLPreviewView`) is tempting because it handles every format nearly free — but it exposes no selection or coordinate API, so it cannot produce locators. Useful as a read-only preview, not as the composer.

Region polygons need a custom drag overlay producing normalized points, and the invariants are strict (≥3 distinct points, non-self-intersecting, non-zero area, in-bounds). **Validate locator JSON in Go**, so a future Windows client inherits it and `locator_json` can never be written malformed.

## 6.1 A pinned Citation collapses the map-first / citation-first argument

One Citation may support many Observations — the data model says so explicitly. So the composer should be able to **stay pinned** to the current Citation while the researcher keeps working: transcribe the census line once, then attach eight Observations, place three bubbles, and draw two connections against that same locator.

That matters because it dissolves a workflow question this note originally treated as either/or. Map-first ("place things, then cite them") and citation-first ("read one line, then record what it contains") become the same UI operated in a different order. Nothing has to be bet on which one researchers actually prefer.

It also argues strongly for **side-by-side, not modal**. A sheet that covers the graph you are annotating is the wrong default; the composer wants to be an inspector or a companion window that persists across successive edits.

---

# 7. macOS feasibility, and the deployment-target question

## 7.1 Rendering

The standard pattern is a hybrid: SwiftUI `Canvas` for the grid and edges (immediate-mode, GPU-backed, cheap for hundreds of lines), with real SwiftUI views for the Nodes in a `ZStack`, positioned with `.position(_:)`. `Canvas` cannot host subviews, take per-shape gestures, or expose accessibility children, so Nodes must be real views to get hit-testing, text fields, focus, and VoiceOver. That comfortably handles the tens-to-low-hundreds of Nodes one Source produces — and would not handle a whole-project graph, which is a further reason to keep this Source-scoped.

## 7.2 Pan and zoom needs AppKit regardless of target

Checked, because it drives the target question: **SwiftUI's `ScrollView` has no zoom on any current version.** Apple's own documentation states it "does not provide zooming functionality," and macOS 15's `ScrollPosition` adds programmatic scrolling to a view id, offset, or edge — *not* magnification. So the answer is `NSScrollView` via `NSViewRepresentable`, using `allowsMagnification`, `magnification`, `minMagnification` / `maxMagnification`, `setMagnification(_:centeredAt:)`, and `magnify(toFit:)`. That is also the more Mac-native result (elastic bounds, native scroll momentum) and is squarely the "escape hatch, not the default" case in [`macos-client-patterns.md`](../macos-client-patterns.md) §4.

**Concrete gotcha to plan for:** a magnified `NSScrollView` does not correctly translate points into the coordinate space of its SwiftUI children. So hit-testing and drag-to-connect cannot naively trust SwiftUI gesture locations — coordinates need resolving in the scroll view's content space. Build **one** coordinate-conversion seam, keep it a pure function over value types, and unit-test it. Getting this wrong produces bugs that feel like "the canvas is haunted."

## 7.3 Should we drop macOS 14?

Current state: local Xcode 26.6 on macOS 26.6; deployment target 14.0 in both configurations; CI runs the `macos-15` runner. Raising to 15 is cheap and CI-compatible today.

But **do not justify it with the canvas.** Pan/zoom needs AppKit either way (§7.2), which was the main hoped-for win. What raising to 15 actually buys is modest and general: `ScrollPosition` for "scroll to this Node," `onScrollGeometryChange` for viewport tracking or a future minimap, the `@Entry` macro, and possibly `LocalizedStringResource` inits that would retire the `String(localized:)` boilerplate documented in [`macos-client-patterns.md`](../macos-client-patterns.md) §6.

macOS 26 adds rich-text `TextEditor` bound to `AttributedString`. That is deliberately **not** wanted for `transcription`, which must stay faithful plain text — formatting in evidence text is a liability, not a feature. It might be interesting for notes later, but `citation_notes.body` / `observation_notes.body` are `TEXT`, so it would need a serialization decision first.

**Recommendation: decide this on support-matrix grounds, not capability grounds.** It is a defensible call — macOS 14 is two releases behind — but it should be its own decision, not a rider on this spike. If we do raise it, bump the CI runner and both Xcode configurations together.

## 7.4 Accessibility and testing

A free-form spatial canvas is genuinely hostile to VoiceOver and keyboard-only use. It needs full keyboard parity for every canvas action, plus a structured outline of the same graph serving as the accessibility representation.

That outline pays for itself twice, because XCUITest cannot meaningfully drive a canvas (it finds controls by accessibility, not pixels). Coverage has to come from Go tests for schema, macros, and locator validation; Swift unit tests for pure geometry and coordinate conversion; and a structured editing path that *is* testable. Budget for the canvas itself being verified by hand.

---

# 8. Pros

- **It matches the unit of thought.** One Source, one graph, one session. No bouncing between three list/detail pairs to record one line of a census.
- **It makes bridge Nodes tolerable.** `participation`, `location`, and `relationship` are the price of a generic property graph and are miserable to fill in through forms. As a connect gesture with a labelled bubble, they become almost pleasant — and the bubble keeps them honest, visible as the entity they really are.
- **It is naturally bounded.** Source-scoped graphs stay small. This is why it can work where "show me my whole family tree as a graph" does not.
- **Completeness becomes visible.** A sparse Node or a dangling bridge is obvious on a canvas and invisible in a list. The graph doubles as a QA view for the interpretation of that Source.
- **Citation stays on the hot path.** If every property edit routes through the composer, "cited by construction" becomes the path of least resistance rather than a discipline to maintain.
- **It is the product differentiator.** Evidence-first genealogy tools are mostly form-driven. This is the part of Provenencia that would not be a reimplementation of something that exists.

---

# 9. Cons and risks

- **Highest-risk UI in the app, on zero prior art**, landing alongside the layer's schema, Go, and FFI. Mitigated by slicing (§11), not eliminated.
- **The canvas cannot express everything.** Polarity, competing Observations, transcription uncertainty, and cross-source Observations all resist spatial representation. Some list/detail surface is still required, so this is "graph *plus* fewer pages," not "graph instead of pages."
- **Round-trip lossiness** (§3.3), in a tool whose entire purpose is not misrepresenting evidence.
- **Position becomes state we did not want** (§5), with migration and sharing consequences.
- **Accessibility and keyboard parity are non-optional work**, not a polish pass.
- **It fights the current session cache.** `WorkspaceSession` patch/invalidate is tuned for lists and detail payloads. A graph is one large payload mutated constantly from inside the view; naive invalidation reloads everything on every write, naive patching drifts from the catalog. Needs per-Node and per-Observation patching against a graph query key.
- **Deletion is a genuine UX problem, not a button** (§4.3).
- **Scope creep is the default outcome.** "Snap to grid, then auto-layout, then edge routing, then minimap, then multi-select, then alignment guides" is an infinite backlog that produces no genealogical capability. The tray (§5) is one deliberate refusal; there will need to be more.

---

# 10. Hiccups checklist

1. **Edge macros** — the matrix in §3.2, including the refusals and the person→person disambiguation.
2. **Reverse rendering rules** — visual states for negated, conflicted, and incomplete; collapse/expand of bridge bubbles.
3. **Deletion semantics** — `ON DELETE` choice at migration time; subject *and* object references; the confirmation that counts the damage.
4. **Type correction** — delete-and-recreate, gated on having no Observations.
5. **Layout table** — integer grid cells, unaudited, plus the unplaced tray.
6. **Artifact gate** — the empty state when a Source has no Artifact.
7. **NameValue end to end** (§4.4) — schema, Go, and editor, on the critical path.
8. **Locator validation in Go**, with the full invariant set, before any UI writes `locator_json`.
9. **Coordinate conversion** under magnification (§7.2) — one seam, unit-tested.
10. **Cross-source Nodes** — a Node homed to another Source can legitimately appear here; ghost styling, and whether it is editable in this graph.
11. **Cache strategy** for a large, constantly mutated graph payload.
12. **Navigation history** — the graph is a place ([`add-workspace-location`](../../.cursor/skills/add-workspace-location/SKILL.md)); camera and selection probably are not. `WorkspaceLocation` is a flat struct of optional ids and needs new fields plus an identity decision.
13. **Undo** — one connect gesture writes several rows and people will hit ⌘Z. The audit model already says undo is a forward revision, not a deletion ([`audit-revision-history.md`](../audit-revision-history.md) §9), so undo is a Go concern, not an `UndoManager` concern.
14. **Density and filtering** — a census page yields dozens of Nodes and hundreds of Observations; needs layers/filters before it is usable on a real source.
15. **Design system** — canvas chrome, bubbles, edges, and selection from existing tokens, not a parallel visual language.
16. **Localization** — fewer strings than a form-heavy screen, but bubbles, macro menus, and every warning state still go through `L10n`.

---

# 11. Vertical slices

Not backend-then-frontend. Each slice cuts through migration → Go → FFI → UI and ends at something usable, so canvas risk is retired early without betting the layer on it.

**Slice 0 — Interpretation vocabulary.** `node_types`, `properties`, `node_type_properties`: migration, seed, Go, FFI, and a vocabulary destination reusing `CatalogVocabulary`. Durable regardless of what happens to the canvas, and a hard prerequisite for everything else — nothing can be asserted until Properties exist and are bound to Node Types.

**Slice 1 — bubbles on a grid (the risk-retirement slice).** `nodes` table plus the layout table; create / label / move / delete for **`person` only**; canvas with the `NSScrollView` bridge, Add Person, drag, snap, select, persist. No Citations, no Observations, no artifact viewer. This proves the AppKit bridge, coordinate conversion under magnification, layout persistence, session-cache behavior under canvas mutation, and the shape of the accessibility outline — while risking approximately zero interpretation semantics. If the canvas is a bad idea, we learn it here, cheaply.

**Slice 2 — events and places.** Same machinery, two more Node Types. Cheap, and makes the graph feel like the real thing.

**Slice 3 — artifact viewer and Citations.** PDF and image only; `page`, `region`, `text_quote` selectors; Go-side locator validation. The largest single unknown, but now there is a canvas to hang it on.

**Slice 4 — the first Observation.** "Add Property" on a root bubble → composer → one Observation, `text` value type only. This is the slice where the whole vertical path is proven end to end.

**Slice 5 — the rest of the value types**, including NameValue end to end (§4.4) and reuse of the existing DateValue editor.

**Slice 6 — the connect tool.** Bridge macros, the disambiguation form, and the pinned Citation (§6.1). Deliberately after single Observations work, because the macros are compound versions of them.

**Slice 7 — honesty and polish.** Negated / conflicted / uncited visual states, filtering, the accessibility outline, undo.

The property that matters: **the layer is dogfoodable from Slice 4 onward, and the canvas is proven at Slice 1.** Risk is retired early and there is still a fallback if the spatial UI turns out to be wrong.

---

# 12. Naming

**Decided: Interpretation graph.** It matches the layer name and the data model's own language — the Interpretation layer *is* "a cited property graph" ([`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §1.2), so the UI name and the schema vocabulary agree instead of introducing a third word for the same thing. Layer vocabulary stays intact inside it: Citations, Observations, and Nodes keep their names.

Considered and set aside:

| Candidate | Why not |
| --- | --- |
| Interpretation map | "Map" reads well but adds a term the data model does not use. |
| Source map | Emphasizes Source scope, but sounds geographic — awkward alongside `place` Nodes. |
| Evidence map | Legible to non-experts, but "evidence" sits closer to the Source layer in our vocabulary. |
| Worksheet | Plays down the visual and plays up the work session; too humble for the layer's primary surface. |

---

# 13. Open questions

- Does the canvas *create* root Nodes only, or also adopt Nodes created elsewhere (imports, other Sources)?
- Can one graph span Sources (a "case view"), or is Source scope hard?
- Are `source` Nodes and source-to-source `mentions` edges on this canvas, or a different view?
- Does the person → person disambiguation (§3.2) earn its complexity, or should person → person simply always mean `relationship` and let shared-event modelling go through the event bubble?
- How does Conclusion-layer work ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md)) surface here later — same canvas with a layer toggle, or a separate reconciliation view? "Not now" is fine; "never" would be a mistake.
