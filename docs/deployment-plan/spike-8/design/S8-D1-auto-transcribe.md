# S8-D1 — Auto Transcribe (citation composer)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-01**  
**Depends on:** **S8-D7** / **S8-10** (new composer form; brief [`archive/S8-D7-composer-rethink.md`](archive/S8-D7-composer-rethink.md)); **S8-D8** / **S8-11** (citation fields section gets its own **Save citation**; the footer Save is gone — draw on those frames, brief [`archive/S8-D8-composer-connect-simplification.md`](archive/S8-D8-composer-connect-simplification.md)); locators S7-07; transcription `PVField` + `PVTextArea` in [`CitationComposerFormPane`](../../../../macos/App/Features/CitationComposer/CitationComposerFormPane.swift) (place the control on the **S8-11** citation fields section)  
**Related:** [`docs/dogfood/ux.md`](../../../dogfood/ux.md) (OCR notes); locator decisions in [Spike 7](../../archive/spike-7/)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md) — compose kit; snowflake wiring in `Features/CitationComposer/`

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the existing **S8-D7** / **S8-D8** composer board — not a rethink, and not the Spike 7 form. Keep the current frames (Layout A, citation fields section, **Save citation**, viewer, observation stack). Add **Auto Transcribe** on the transcription field. Do not restyle anything this brief does not name.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
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

Add **Auto Transcribe** next to the citation **transcription** field. On click, OCR the **image Artifact**, or the **region polygon** when one is on the locator, and dump the text into the textarea. The researcher edits as today, and **Save citation** (S8-D8) writes the Citation. Filling the field makes the citation fields dirty, so the S8-11 unsaved-work guard applies.

**Image types only.** PDF, audio, and video do not run OCR in this MVP. PDF Find / select / paste is **S8-D2** (not this button). Image-only PDF scans are not Vision’d here.

Also design:

- In-progress / disabled states
- Confirm before **replacing** non-empty transcription
- A **warning** when the job looks like a full / huge page (accuracy may be poor, it may take a while, consider cropping) that still **lets them proceed**
- Failure and unsupported-media copy

**Do not** redesign the viewer, locator tools, Observation list, or Save/Cancel footer.

```text
Citation
  Transcription          [ Uncertain ]     [ Auto Transcribe ]
  ┌─────────────────────────────────────────────────────────┐
  │  (textarea — filled by OCR, then edited)                │
  └─────────────────────────────────────────────────────────┘
```

Placement is a board finding: trailing slot on `PVField` (beside Uncertain), a control under the label row, or under the textarea. Do not hide the button inside the textarea chrome.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Transcription ≠ Observation | Button only fills the transcription string. No Observation rows appear. |
| Locator is layered | Image + region → OCR that crop. Image, no region → whole image (warn if huge). PDF → no OCR. |
| Vision has no “too many words” error | We warn from **our** checks (no region / huge pixels / dense-page heuristic). Vision usually succeeds slowly or with junk. |
| Crop is in-memory | No “saving a clip” progress. Spinner is “Reading text…”. |
| Images only for OCR | PDF / audio / video / missing file: disabled + short reason. PDF paste-from-selection is **S8-D2** / **S8-05**, not this button. |
| Researcher owns the reading | OCR is a draft in the field. Uncertain stays a manual checkbox unless a later story says otherwise. |

### 2.1 What this board is not

- Not PDF OCR, PDF Find, or text-layer select/paste ([**S8-D2**](S8-D2-pdf-text-find.md)).
- Not Foundation Models / Apple Intelligence / draft graph cards.
- Not auto Observations, NameValue, or connect macros.
- Not a new kit **Transcribe** component unless a second call site is already known (it is not).
- Not raising macOS 14.

---

## 3. Implementation gate (S8-01)

| Ships in **S8-01** | Does **not** ship there |
| --- | --- |
| Auto Transcribe control + states on the transcription field | Observation / subject extract |
| Vision on **image** raster; in-memory crop from region | PDF Vision; PDFKit `string` / `findString`; persisted crop files |
| Replace confirm; large-page warn + proceed; fail copy | Hard-block full page; Live Text overlay |
| Protocol-shaped recognizer for tests | Calling real Vision from unit tests |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| AT-1 | Board shows transcription **idle** (empty + with existing text), **running**, **success** (field filled), **failed**, **PDF** (disabled + paste hint), and **unsupported** (audio / no file). |
| AT-2 | Auto Transcribe is a `PVButton` (or `PVIconButton` + accessible name). Copy via L10n. Disabled while `inert`, while running, when the Artifact is not an image, and when there is no image raster. |
| AT-3 | Region on an image locator → copy/tooltip can say the crop is the drawn region. Image, no region → whole image. |
| AT-4 | Non-empty transcription → **confirm replace** (`.pvConfirm`, not a Bool sheet). Cancel leaves the field. |
| AT-5 | Large-page / no-region / oversized-bitmap warning: `PVCallout` and/or `.pvConfirm` with **Proceed** and **Cancel**. Cancel does not start OCR. Proceed runs it. |
| AT-6 | Running: button busy; field not silently editable mid-flight or clearly locked — board picks one. No fake determinate % unless Vision progress is actually shown. |
| AT-7 | Failure: short recoverable copy (could not read, nothing found). Do not empty a previous transcription on failure. |
| AT-8 | Keyboard / VoiceOver: control is in the form tree (composer is its own place). |
| AT-9 | UI inventory: no new DesignSystem component unless a finding says the kit is missing a confirm/callout pattern (it is not). |

---

## 5. Suggested frames

1. Idle — empty transcription, image Artifact, no region.
2. Idle — image + rectangle region (button implies crop).
2b. Idle — PDF Artifact: Auto Transcribe disabled (Find / paste are **S8-D2**).
3. Replace confirm — field already has text.
4. Large-page warning — artifact-only newspaper page; Proceed / Cancel.
5. Running.
6. Success — textarea filled; researcher can edit.
7. Fail — Vision error / empty result; prior text kept.
8. Unsupported — audio Artifact or no file; button disabled + caption. PDF is frame 2b, not this.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Provenencia UI is layered as **components / recipes / snowflakes** ([`docs/design-system-layers.md`](../../../design-system-layers.md)). Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Citation composer form pane | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerFormPane.swift` | Host the control on the transcription `PVField` / under the textarea. |
| Citation composer model | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerModel.swift` | Running flag, preflight, replace/warn presentation, apply text. |
| OCR / crop seam | Snowflake (not DS) | **New** | Prefer `Features/CitationComposer/` or a tiny `Features/ArtifactOCR/` helper — **protocol** + Vision adapter | No catalog types in DesignSystem. Tests inject a fake. |
| Field | Component | Ship | `DesignSystem/Components/Field/PVField.swift` | Existing trailing slot already holds Uncertain. Board decides if the button shares that row. |
| TextArea | Component | Ship | `DesignSystem/Components/TextArea/` | Fill `model.transcription` only. |
| Button / IconButton | Component | Ship | `Components/Button/`, `Components/IconButton/` | Auto Transcribe. |
| Confirm | Component | Ship | `.pvConfirm(item:)` | Replace; optionally large-page proceed. Snapshot via `item:`, not Bool. |
| Callout | Component | Ship | `DesignSystem/Components/Callout/PVCallout.swift` | Optional inline warning above the field (actions slot already exists). Do not invent a banner. |
| Uncertain checkbox | Snowflake | Ship | Same form pane | Do not auto-toggle in this brief. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| `PVAutoTranscribe` kit control | One call site. Compose button + confirm. |
| Viewer “Live Text” overlay | Different product; transcription field is the MVP. |
| Progress as a new kit spinner | Reuse existing busy/disabled button patterns. |
| Message center on the graph | This is composer-only. |

---

## 7. Out of scope

- Writing Observations or subjects
- Foundation Models / PCC
- PDF OCR, PDF Find, and text-layer select/paste (**S8-D2**)
- Persisting OCR text before Save
- Multiple regions
- Changing locator tools

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-01** against the board: Vision + in-memory crop + L10n + tests with a fake recognizer.
