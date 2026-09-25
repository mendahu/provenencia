# S8-D2 — PDF Find, text selection, paste transcription

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** **S8-03** (PDFKit viewport + I-beam default), **S8-04** (Find), **S8-05** (paste from selection)  
**Depends on:** **S8-D7** / **S8-10** (viewer slot on the new composer; brief [`archive/S8-D7-composer-rethink.md`](archive/S8-D7-composer-rethink.md)); **S8-D8** / **S8-11** (citation fields section with Save citation; no footer Save — brief [`S8-D8-composer-connect-simplification.md`](S8-D8-composer-connect-simplification.md)); tool strip ([`ArtifactViewerToolChrome`](../../../../macos/App/Features/ArtifactViewer/ArtifactViewer.swift)); locators S7-07; transcription field (**S8-D1** image Auto Transcribe — do not fight that row)  
**Related:** [`../pdf-text-find.md`](../pdf-text-find.md); dogfood OCR notes  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the **S8-D7** composer viewer / transcription row.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
- **Rethink** (this brief says replace): throw away the old frames. Do not keep a before/after to ship.
- **Enhancement**: add to the existing frames. Do not start a second composer / graph / page.

1. **Clear this board’s local design-system cache.** Claude Design keeps a stale pack; drawing against it invents local copies of kit controls.
2. **Delete this board’s reference** to the design-system bundle.
3. **Pull a fresh copy** of the Provenencia design system from the main project. Do not continue until the fetched kit lists current components. If the kit looks stale or empty, delete the cache and refetch. Do **not** draw a replacement kit locally.
4. **Compose from that kit.** Instance existing components. Reach for a **bespoke / local** control only when the use is truly this domain. One call site is not a new design-system primitive.

**Reach for (kit).** Instance these first. The **UI building-block inventory** later in this brief names the snowflakes and which kit piece each situation should use.

| Situation | Use |
| --- | --- |
| Labeled value, textarea, or trailing control | Field + TextArea / Input |
| Primary / secondary / ghost action | Button; icon-only → IconButton |
| Choose one from a short list | Select |
| Searchable pick | ComboBox |
| Warning, error, or inline hint | Callout |
| Page- or pane-level empty | EmptyState |
| Confirm replace or destroy | Confirm (`item:` snapshot, not a Bool) |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

---

## 1. Objective

Make a **born-digital PDF** usable in the citation composer without OCR:

1. **Find** on the PDF tool strip — keyword field + Find (next/previous). Jump to that page and **highlight** the hit. For large registers.
2. **Text selection** as the default PDF pointer (I-beam). Today click-drag **pans** the raster. Pan must remain possible (hand tool and/or modifier) but must not steal every drag.
3. **Paste transcription from selection** on the transcription field (PDF only) — sibling to image **Auto Transcribe** (**S8-D1**).

Do **not** redesign Observation list, Save/Cancel, or image-viewer pan/zoom.

```text
[ ⟨  12  of 40  ⟩ ] [ Set page ] | [ −  100%  + ] | [ region tools ] | [ Find  ________  ⌕  ⟨ ⟩ ]

Transcription     [ Uncertain ]  [ Auto Transcribe ]   ← image only
                  [ Paste transcription from selection ]  ← PDF + non-empty selection
```

Exact placement of Find vs page/zoom/region is a board finding. Paste lives with the transcription field, not in the viewer strip.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| S7-06 PDF is a **bitmap** | Find/select require a **PDFKit page**, not `NSImage`. S8-03 is the remount. |
| Click-drag currently pans | Default becomes select; pan is a mode or modifier. Cursors must match. |
| Region tools (S7-07) | Armed region tool: draw, not select. Find can stay. Overlay coords follow the new view. |
| No text layer (scan-in-PDF) | Honest disable. Do not offer Auto Transcribe / Vision here. |
| Transcription ≠ Observation | Paste fills the textarea only. Same replace-confirm pattern as S8-D1. |
| Composer-only this spike | Do not design Source-page Find. |

### 2.1 What this board is not

- Not image OCR (**S8-D1** / **S8-01**).
- Not PDF page-1 **thumbnails** (**S8-02**).
- Not `text_quote` locators — parked in [`text-quote-locators.md`](../../../ideas/text-quote-locators.md).
- Not Foundation Models.

---

## 3. Implementation gates

| Ships in **S8-03** | **S8-04** | **S8-05** |
| --- | --- | --- |
| PDFKit-backed page; I-beam default; pan mode/modifier; region overlay still works; cursors | Find field + next/prev + highlight + page jump; no-text empty | Paste control; replace confirm; disabled with no selection / no text layer |

S8-04 and S8-05 **must not** start on the raster viewer.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PF-1 | PDF tool strip shows **Find** (field + commit + next/previous). Image strip unchanged. |
| PF-2 | First hit / next / prev updates **page chrome** and a visible **highlight**. Wrap or stop at ends — board picks one and says so. |
| PF-3 | No matches / no text layer: short honest copy. Not a Vision offer. |
| PF-4 | Default PDF cursor is **I-beam**; drag selects text. Pan documented (hand control and/or Option-drag). |
| PF-5 | Armed **region** tool: select-drag is off; draw as today. Disarm restores I-beam. |
| PF-6 | **Paste transcription from selection** on the transcription row for PDF. Disabled without a selection. Image shows Auto Transcribe, not this button (or both visible with the other disabled — board picks one pattern, not two competing fills). |
| PF-7 | Non-empty transcription → **replace confirm** (`.pvConfirm`), same as S8-D1. |
| PF-8 | Keyboard: Find field, Return = find next; VoiceOver on Find and Paste. |
| PF-9 | Inventory: no new kit primitive unless Find field cannot be `PVField` + `PVIconButton`. |

---

## 5. Suggested frames

1. PDF idle — I-beam, no selection; Paste disabled.
2. Text selected — Paste enabled.
3. Find typed + hits — highlight + page change.
4. Find no hits / image-only PDF.
5. Hand / pan mode (if a control).
6. Region tool armed — draw, not select.
7. Transcription row: image Artifact (Auto Transcribe) vs PDF (Paste).
8. Replace confirm on paste.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Artifact viewer tool chrome | Snowflake | **Extend** | `Features/ArtifactViewer/ArtifactViewer.swift` (`ArtifactViewerToolChrome`) | Find group after page/zoom/region. |
| Artifact media viewport | Snowflake | **Extend / remount** | `Features/ArtifactViewer/ArtifactMediaViewport.swift` | PDF path leaves raster; images stay. |
| Artifact viewer model | Snowflake | **Extend** | `Features/ArtifactViewer/ArtifactViewerModel.swift` | Keep `PDFDocument`; stop flattening PDF to `displayImage` for paint. |
| Region overlay | Snowflake | **Extend** | `Features/ArtifactViewer/ArtifactRegionOverlay.swift` | Remap to PDFKit page space. |
| Citation composer form | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerFormPane.swift` | Paste control next to S8-D1 Auto Transcribe. |
| Field / Button / IconButton | Component | Ship | `DesignSystem/Components/…` | Find field + icon buttons; Paste `PVButton`. |
| Confirm | Component | Ship | `.pvConfirm(item:)` | Replace on paste. |
| Callout | Component | Ship | `PVCallout` | Optional no-text-layer note. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| `PVFindBar` kit control | One composer PDF strip unless a second host exists. |
| Live Text on images | S8-D1 Vision path. |
| Source Artifact sheet Find | Later. |

---

## 7. Out of scope

- Vision on PDF pages
- `text_quote` locator writes
- Changing Save / Observation chrome
- Thumbnail generation (**S8-02**)

---

## 8. Handoff

1. Archive this brief when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-03** first, then **S8-04** and **S8-05** (parallel after 03).
