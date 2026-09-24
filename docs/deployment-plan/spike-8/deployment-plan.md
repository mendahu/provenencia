# Deployment Plan — Spike 8

Pause-and-refine: make **Source → Evidence graph** data entry cheaper. Stories are independent improvements to that flow, not one schema epic. Authoritative composer/locator: [Spike 7 archive](../archive/spike-7/). Dogfood seed: [`docs/dogfood/ux.md`](../../dogfood/ux.md).

## Status

**Open.** Landings go in [`completed.md`](completed.md). More stories will be added under the same spike.

> **Goal of this spike:** cut the tedium of entering real research without reopening the Interpretation model. First slice is Auto Transcribe (Vision → transcription). Later slices join this plan as they are scoped.

## Goal (dogfood bar)

Grow this list as stories land. **By spike close**, every checked story below must be true in the app.

1. **Auto Transcribe** — in the citation composer, a control fills **transcription** from Vision OCR of the current page (or the region polygon when one is set). The researcher can edit and Save as today. Full-page / oversized jobs warn and can still proceed. No Observation writes.

Further bar items: TBD (additional data-entry stories).

## Design track

**Composer chrome for Auto Transcribe is designed in Claude Design before the implementation PR.** Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S8-D1** | Auto Transcribe in the composer | Button, progress, replace confirm, large-page warning + proceed, failure copy | **S8-01** |

## PR sequence

```text
   design                         build
─────────────               ──────────────────────────────────────────

S8-D1  Auto Transcribe UI     (more stories TBD)
  │
  └────── gates ──────────▶ S8-01  Vision + crop + fill transcription
                              │
                            (more PRs as stories are added)
                              │
                            S8-99  Dogfood close / docs
```

---

## Checklist

- [ ] S8-D1 — Design: Auto Transcribe in the citation composer → [`completed.md`](completed.md)
- [ ] S8-01 — Vision OCR + Auto Transcribe button → [`completed.md`](completed.md)
- [ ] S8-99 — Dogfood close / docs (after later stories, or when we choose to close)

---

## S8-D1 — Design: Auto Transcribe in the citation composer

Claude Design board for the **transcription** field: Auto Transcribe control, in-progress state, replace confirm, large-page / slow-job warning that can still proceed, and failure/empty states. Brief: [`design/S8-D1-auto-transcribe.md`](design/S8-D1-auto-transcribe.md). Gates **S8-01**.

Does **not** design Observation auto-fill, LLM extract, or PDF Find.

---

## S8-01 — PR: Vision OCR + Auto Transcribe

On-device Vision (`VNRecognizeTextRequest`) fills the composer **transcription** textarea. Crop in memory from the current page raster and the region locator when present. No temp file, no catalog write until the researcher Saves.

| | |
| --- | --- |
| **In** | Protocol-shaped OCR seam (tests do not call Vision); `CGImage` from the already-loaded image / PDF page raster; bounding-box crop (optional mask later); Auto Transcribe control per **S8-D1**; replace confirm if transcription is non-empty; preflight warn on artifact-only / huge bitmap / dense-page heuristics, with proceed; L10n; skip audio/video / no image. |
| **Out** | PDFKit text-layer extract; Foundation Models; writing Observations; persisted crop objects; `RecognizeDocumentsRequest` (macOS 26); raising the deployment target. |
| **Testable** | Fake recognizer fills / fails / empty; crop uses region vs full page; preflight flags large page; replace does not overwrite without confirm; button disabled while running. |
| **Depends on** | **S8-D1**. Shipped composer + locators (S7-08 / S7-06 / S7-07). |

---

## S8-99 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar) once the cluster is enough (or we stop adding stories). Record in [`completed.md`](completed.md); archive the spike. SemVer only if cutting a product release.

---

## Scope boundary

| In | Out |
| --- | --- |
| Composer transcription assist | Auto Observations / subjects / connect |
| Vision on image + PDF page raster | Audio / video OCR |
| In-memory crop from locator | Object-store crop files |
| Warn + proceed on large pages | Hard reject / Apple “too many words” (does not exist) |
| More data-entry stories as added | Spike 7 leftover honesty/polish (conflicted, tray, pinning) unless pulled in |

---

## Gotchas

1. **Transcription ≠ Observation** — OCR dumps into the citation reading only ([`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4).
2. **Vision does not refuse a newspaper page** — it usually succeeds slowly or with junk. Large-page honesty is **our** preflight (pixels / no region / post-pass observation density), not a `VNError`.
3. **Locator y-down vs Vision ROI y-up** — crop in image pixels from [`ArtifactRegionGeometry`](../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift); do not pass a polygon into `regionOfInterest` (rect only).
4. **Crop the source raster**, not the zoomed viewport bitmap.
5. **No file middleman** — `ProjectFiles.objectURL` → `NSImage` / PDF page render → `CGImage` → crop → `VNImageRequestHandler`.
6. **Hide Vision behind a protocol** — `FakeStore` / unit tests inject a recognizer.
7. **More stories do not wait on S8-01** unless they share the composer transcription chrome.

---

## Definition of done

- [ ] Checklist stories complete (or explicitly descoped)
- [ ] Dogfood bar items for landed stories met
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
