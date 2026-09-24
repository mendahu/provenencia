# S8-D1 — Auto Transcribe (citation composer)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-01**  
**Depends on:** Shipped citation composer (S7-08 / S7-06 / S7-07); transcription `PVField` + `PVTextArea` in [`CitationComposerFormPane`](../../../../macos/App/Features/CitationComposer/CitationComposerFormPane.swift)  
**Related:** [`docs/dogfood/ux.md`](../../../dogfood/ux.md) (OCR notes); locator chrome [S7-D9](../../archive/spike-7/design/archive/S7-D9-locator-region-chrome.md)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md) — compose kit; snowflake wiring in `Features/CitationComposer/`

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Add **Auto Transcribe** next to the citation **transcription** field. On click, OCR the **current page raster**, or the **region polygon** when one is on the locator, and dump the text into the textarea. The researcher edits as today and Save still writes the Citation.

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
| Locator is layered | Region present → OCR that crop. Else → current image or current PDF page. |
| Vision has no “too many words” error | We warn from **our** checks (no region / huge pixels / dense-page heuristic). Vision usually succeeds slowly or with junk. |
| Crop is in-memory | No “saving a clip” progress. Spinner is “Reading text…”. |
| Image + PDF only | Audio / video / missing file: disabled + short reason, not a crash. |
| Researcher owns the reading | OCR is a draft in the field. Uncertain stays a manual checkbox unless a later story says otherwise. |

### 2.1 What this board is not

- Not PDF Find / text-layer extract ([`ideas/pdf-text-find.md`](../../../ideas/pdf-text-find.md)).
- Not Foundation Models / Apple Intelligence / draft graph cards.
- Not auto Observations, NameValue, or connect macros.
- Not a new kit **Transcribe** component unless a second call site is already known (it is not).
- Not raising macOS 14.

---

## 3. Implementation gate (S8-01)

| Ships in **S8-01** | Does **not** ship there |
| --- | --- |
| Auto Transcribe control + states on the transcription field | Observation / subject extract |
| Vision on page raster; in-memory crop from region | PDFKit `string` / `findString`; persisted crop files |
| Replace confirm; large-page warn + proceed; fail copy | Hard-block full page; Live Text overlay |
| Protocol-shaped recognizer for tests | Calling real Vision from unit tests |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| AT-1 | Board shows transcription **idle** (empty + with existing text), **running**, **success** (field filled), **failed**, and **unsupported** (no raster). |
| AT-2 | Auto Transcribe is a `PVButton` (or `PVIconButton` + accessible name). Copy via L10n. Disabled while `inert`, while running, and when there is no OCR-able raster. |
| AT-3 | Region on the locator → copy/tooltip can say the crop is the drawn region. Artifact-only / page-only → implies the visible page. |
| AT-4 | Non-empty transcription → **confirm replace** (`.pvConfirm`, not a Bool sheet). Cancel leaves the field. |
| AT-5 | Large-page / no-region / oversized-bitmap warning: `PVCallout` and/or `.pvConfirm` with **Proceed** and **Cancel**. Cancel does not start OCR. Proceed runs it. |
| AT-6 | Running: button busy; field not silently editable mid-flight or clearly locked — board picks one. No fake determinate % unless Vision progress is actually shown. |
| AT-7 | Failure: short recoverable copy (could not read, nothing found). Do not empty a previous transcription on failure. |
| AT-8 | Keyboard / VoiceOver: control is in the form tree (composer is its own place). |
| AT-9 | UI inventory: no new DesignSystem component unless a finding says the kit is missing a confirm/callout pattern (it is not). |

---

## 5. Suggested frames

1. Idle — empty transcription, image Artifact, no region.
2. Idle — PDF page + rectangle region (button implies crop).
3. Replace confirm — field already has text.
4. Large-page warning — artifact-only newspaper page; Proceed / Cancel.
5. Running.
6. Success — textarea filled; researcher can edit.
7. Fail — Vision error / empty result; prior text kept.
8. Unsupported — audio Artifact or no file; button disabled + caption.

---

## 6. UI building-block inventory

Provenencia UI is layered as **components / recipes / snowflakes** ([`docs/design-system-layers.md`](../../../design-system-layers.md)). Paths from `macos/App/` unless noted.

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
- PDF text-layer extract and Find
- Persisting OCR text before Save
- Multiple regions
- Changing locator tools

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-01** against the board: Vision + in-memory crop + L10n + tests with a fake recognizer.
