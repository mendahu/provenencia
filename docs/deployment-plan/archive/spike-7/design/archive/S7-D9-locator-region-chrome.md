# S7-D9 — Locator chrome (tools, summary list, layered selectors)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-07** (locator tools — artifact + page + region)  
**Depends on:** Shipped composer + Artifact viewers (**S7-08**, **S7-06**); locator validation in Go (**S7-03** / `core/locator` — extend in S7-07); citation composer board Frames ([`S7-D4-citation-composer.md`](S7-D4-citation-composer.md))  
**Schema SoT:** [`interpretation-layer-data-model.md`](../../../../../interpretation-layer-data-model.md) §3 (incl. **`artifact`** selector)  
**Related briefs:** [`S7-D4`](S7-D4-citation-composer.md) — tool strip + locator strip; do not redesign the whole composer place here  
**Schedule:** Immediately before **S7-07**. Do not implement region drawing until this board locks the look.

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`../README.md`](../README.md) first.

---

## 1. Objective

Design the **full locator experience** in the citation composer viewer pane — not only the polygon overlay:

1. **Default entire-artifact cite** — every new Citation starts with an `artifact` selector (no checkbox).
2. **Layer refinements** — optional **Set Page** (PDF) and optional **region** polygon tools on top.
3. **Locator summary list** — readable rows for each layer (Entire artifact → page 2 → polygon), with per-row remove.
4. **On-canvas overlay** — one region, dim-outside focus, constrained edit.

```text
Citation locator (layered, outer → inner)
  artifact          ← always present (default / floor)
  page?             ← PDF: required before/with region; Set Page or auto-layered on first polygon
  region?           ← at most one polygon on image / that page
```

Small Artifacts (birth certificate, single photo): leave the default — save with `artifact` only. No forced polygon.

**Keep this board focused.** Do not redesign Observation form, NameValue, or PDF Find.

---

## 2. Schema decision: default `artifact`, then layer

**Do not leave `locator_json` empty.** Citations require a locator. **Do not** use a “Cite entire artifact” checkbox.

**Lock:**

1. Opening the composer (or picking an Artifact) **prepopulates**:

```json
{
  "version": 1,
  "selectors": [{ "type": "artifact" }]
}
```

2. **Set Page** (PDF toolbar) appends/updates a `page` selector **after** `artifact`.
3. Region tools append/replace a single `region` **after** `artifact` (and after `page` when present).
4. **PDF dependency:** a `region` on a paginated Artifact **must** follow a `page`. If the researcher draws a polygon before Set Page, the app **auto-layers** `page` using the **current viewer page**, then the region — same end state as Set Page then draw. Images (no pages) stay `artifact` → `region`.
5. Removing a layer peels inward: remove region → keep artifact (+ page); remove page → keep artifact **and drop region** (region without page is invalid on PDFs); you **cannot** remove `artifact` — it is the floor. “Clear region” only clears the polygon layer.

Example after Set Page 2 + rectangle:

```json
{
  "version": 1,
  "selectors": [
    { "type": "artifact" },
    { "type": "page", "artifact_page": 2 },
    { "type": "region", "unit": "normalized", "points": [ ... ] }
  ]
}
```

| Approach | Verdict |
| --- | --- |
| Optional / empty locator | Reject |
| Checkbox “Cite entire artifact” XOR page/region | Reject — extra mode; easy to get wrong |
| **Default `artifact`, layer page/region** | **Ship** — entire-artifact cite is zero extra clicks; refinements are additive |

Region tools still emit `region` polygons only (circle ≈ point ring).

---

## 3. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Locators are an ordered selector list | Summary is a **list of layers**, outer → inner. |
| `artifact` is the floor | Always listed; never deleted; Save always has a locator. |
| Page is optional until a region needs it | **Set Page** on PDF tool strip; images have no page tool. |
| **PDF: region ⇒ page** | Polygon without a page is invalid. First region create auto-inserts current viewer page if missing. |
| One region max | New shape replaces existing region layer only. |
| Schema polygon points | Rect / L / circle / freeform → normalized `region`. |
| Design system color | Stroke / handles / dim veil use `PVColor` tokens. |

### 3.1 What this board is not

- Not full composer IA — **S7-D4**.
- Not PDF Find (ideas parking lot).
- Not `text_quote` as a required locator.
- Not Evidence graph card chrome.

### 3.2 Implementation gate (S7-07)

| Ships in S7-07 (after this board) | Does **not** ship there |
| --- | --- |
| `artifact` validate; default chain on compose | Empty locator; entire-artifact checkbox |
| Set Page + region tools; layered summary list | Multiple regions |
| Dim-outside + edit + clear region layer | Full PDF Find UI |
| PDF text selection for paste | Required `text_quote` |

---

## 4. Locked product behavior

### 4.1 Locator summary (list — required)

A persistent **locator field** listing current layers:

| Layer present | Example row | Remove |
| --- | --- | --- |
| `artifact` | **Entire artifact** | Not removable (floor) |
| `page` | Currently selected: **page 2** | Removes page; also clears region if any |
| `region` | **Rectangle** / L / Circle / Polygon | Removes region only |

- Always visible while composing.
- Plain language, not raw JSON.
- Tool-strip **Clear** (or list action) clears **region** (and optionally page — board may offer Clear region vs Reset to entire artifact). Reset to entire artifact = selectors back to `[{type: artifact}]`.

### 4.2 No entire-artifact checkbox

Entire-artifact cite **is** the default list state (only the Entire artifact row). Researchers who need more use **Set Page** and/or region tools. Thin composer’s `markWholeImageLocator` stub is **replaced** by this default.

### 4.3 Tool strip (viewer top)

| Control | Role |
| --- | --- |
| **Set Page** | PDF only — commits current viewer page into the locator (append/update `page` after `artifact`). May live beside page chevrons or as an explicit button — board picks placement. |
| **Rectangle** / **L** / **Circle** / **Freeform** | Create region layer. On PDF, if no `page` yet, **also** layer `page` = current viewer page, then the region. |
| **Clear** | Peel region (and/or reset to artifact-only — label clearly) |

Page nav (chevrons / field) still moves the viewer; **Set Page** writes the page without requiring a polygon. Prefer explicit Set Page for page-only cites so browsing does not silently change the cite — but **drawing a region always pins a page** (current viewer page) when one was missing.

Zoom group stays. Mutually exclusive region-tool arming.

### 4.4 Focus dimming

Region committed ⇒ dim outside / clear inside. Artifact-only or artifact+page with no region ⇒ **no** dim veil.

### 4.5 Edit after create

| Shape | Edit |
| --- | --- |
| **Rectangle** | Corners (optional edges); axis-aligned |
| **L-shape** | Handles preserve that L orientation |
| **Circle** | Radius / bbox; stays circular |
| **Freeform** | Move vertices; right-click delete (floor 3) |

Freeform: close on origin; step back removes last vertex while drawing.

### 4.6 Visual language

Stroke / handles / veil → named `PVColor` tokens. In-progress freeform ≠ committed.

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| LR-1 | Image + PDF frames with same region chrome + dim veil when region set. |
| LR-2 | Tool strip: Set Page (PDF), rect, L×4, circle, freeform, Clear. |
| LR-3 | Default locator = `artifact` only; never empty. |
| LR-4 | Layer page/region on top; summary list shows each layer. |
| LR-4b | **PDF:** region requires page; auto-layer current viewer page when drawing a region without a page yet. |
| LR-5 | At most one region; replace on new commit. |
| LR-6 | Dim outside when region exists. |
| LR-7 | Constrained edit (rect / L / circle); freeform vertex edit + right-click delete + step-back. |
| LR-8 | PVColor tokens for stroke / veil / handles. |
| LR-9 | Remove page / remove region from list; artifact row not removable. |
| LR-10 | No Cite-entire-artifact checkbox. |
| LR-11 | L10n / a11y notes for tools and list rows. |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Artifact viewer canvas | Snowflake | **Extend** | `Features/ArtifactViewer/` | Overlay host. |
| Region overlay + dim veil | Snowflake | **New** | Composer or ArtifactViewer | No GraphCanvas fork. |
| Region + Set Page tools | Host | **Extend** | `CitationComposerViewerPane` | Replace Draw region stub; add Set Page. |
| Locator summary list | Snowflake / recipe | **New** | Composer viewer pane | Layer rows + remove. |
| IconButton / Button | Component | Ship | DesignSystem | Tools + list actions. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Empty / null locator | Floor is `artifact`. |
| Entire-artifact checkbox | Default chain already is the entire Artifact. |
| Multiple regions | One region layer. |
| GraphCanvas tools | Wrong magnification model. |

---

## 7. Screen / frame inventory (minimum)

1. Fresh compose — summary: Entire artifact only; tools idle; no veil.
2. PDF — Set Page → summary adds page row; viewer on that page.
3. Page + rectangle — three summary rows; dim + handles.
3b. PDF — draw rectangle **without** Set Page first → summary still gets page (current viewer page) + region.
4. Image Artifact — no Set Page; artifact → region only (no page layer).
5. Each L orientation; circle; freeform create/close/edit/delete/step-back.
6. Remove region (page + artifact remain); remove page (back to artifact; region cleared); Clear/reset to artifact-only.
7. PDF zoomed region; color token annotations.

---

## 8. Out of scope

- Observation form / NameValue.
- PDF Find (ideas parking lot).
- `text_quote` UI.
- Rotated (non-axis-aligned) rectangles by default.

---

## 9. Deliverable

Claude Design board for default artifact layer + Set Page + region tools + summary list + overlay. Archive when done; implement **S7-07** against the board (incl. Go `artifact` validate + layered chains). The board HTML still said “Entire document”; product copy is **Entire artifact**.
