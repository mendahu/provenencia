# Deployment Plan — Spike 8

Pause-and-refine: make **Source → Evidence graph** data entry cheaper. Stories are independent improvements to that flow, not one schema epic. Authoritative composer/locator: [Spike 7 archive](../archive/spike-7/). Dogfood seed: [`docs/dogfood/ux.md`](../../dogfood/ux.md).

## Status

**Open.** Landings go in [`completed.md`](completed.md). More stories will be added under the same spike.

> **Goal of this spike:** cut the tedium of entering real research without reopening the Interpretation model. Slices land independently. Later stories join this plan as they are scoped.

## Goal (dogfood bar)

Grow this list as stories land. **By spike close**, every checked story below must be true in the app.

1. **Auto Transcribe** — in the citation composer, on an **image** Artifact, a control fills **transcription** from Vision OCR of the image (or the region polygon when one is set). The researcher can edit and Save as today. Full-image / oversized jobs warn and can still proceed. No Observation writes. **PDF:** button disabled / honest empty — do not OCR PDFs in this slice (copy/paste or a later text-layer story).
2. **PDF Artifact thumbnails** — PDF Artifacts show a **first-page raster** on Artifact rows (glyph only if render skips). A PDF with a raster can be pinned as Source cover. Notes: [`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md).

Further bar items: TBD (additional data-entry stories).

## Design track

**Composer chrome for Auto Transcribe is designed in Claude Design before that PR.** PDF thumbs reuse shipped `PVThumbnail` — no board. Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S8-D1** | Auto Transcribe in the composer | Button, progress, replace confirm, large-page warning + proceed, failure copy | **S8-01** |

## PR sequence

```text
   design                         build
─────────────               ──────────────────────────────────────────

S8-D1  Auto Transcribe UI
  │
  └────── gates ──────────▶ S8-01  Vision + crop + fill transcription
                              │
S8-02  PDF first-page thumbs ──┘   (no design gate; parallel with S8-01)
                              │
                            (more PRs as stories are added)
                              │
                            S8-99  Dogfood close / docs
```

---

## Checklist

- [ ] S8-D1 — Design: Auto Transcribe in the citation composer → [`completed.md`](completed.md)
- [ ] S8-01 — Vision OCR + Auto Transcribe button → [`completed.md`](completed.md)
- [ ] S8-02 — PDF first-page Artifact thumbnails → [`completed.md`](completed.md)
- [ ] S8-99 — Dogfood close / docs (after later stories, or when we choose to close)

---

## S8-D1 — Design: Auto Transcribe in the citation composer

Claude Design board for the **transcription** field: Auto Transcribe control, in-progress state, replace confirm, large-page / slow-job warning that can still proceed, and failure/empty states. Brief: [`design/S8-D1-auto-transcribe.md`](design/S8-D1-auto-transcribe.md). Gates **S8-01**.

Does **not** design Observation auto-fill, LLM extract, PDF OCR, or PDF Find.

---

## S8-01 — PR: Vision OCR + Auto Transcribe

On-device Vision (`VNRecognizeTextRequest`) fills the composer **transcription** textarea for **image** Artifacts only. Crop in memory from the loaded `NSImage` and the region locator when present. No temp file, no catalog write until the researcher Saves. PDF / audio / video: do not run Vision.

| | |
| --- | --- |
| **In** | Protocol-shaped OCR seam (tests do not call Vision); `CGImage` from the already-loaded **image**; bounding-box crop (optional mask later); Auto Transcribe control per **S8-D1**; replace confirm if transcription is non-empty; preflight warn on artifact-only / huge bitmap / dense-image heuristics, with proceed; L10n; skip PDF / audio / video / no image. |
| **Out** | PDF OCR; PDF page raster → Vision; PDFKit text-layer extract / copy-paste (later story); Foundation Models; writing Observations; persisted crop objects; `RecognizeDocumentsRequest` (macOS 26); raising the deployment target. |
| **Testable** | Fake recognizer fills / fails / empty; crop uses region vs full page; preflight flags large page; replace does not overwrite without confirm; button disabled while running. |
| **Depends on** | **S8-D1**. Shipped composer + locators (S7-08 / S7-06 / S7-07). |

---

## S8-02 — PR: PDF first-page thumbnails

macOS **PDFKit** renders page 1 of a PDF Artifact File to the same JPEG thumbnail spec as images (longest edge ≤ 256). Persist as the existing unaudited `file_derivatives` thumbnail so Artifact rows and Source cover pins just work. Go `EnsureThumbnail` still **skips** PDF decode — no engine PDF renderer in this PR.

| | |
| --- | --- |
| **In** | Page 1 only; JPEG / 256-edge; PDFKit (same as S7-06 viewer raster); FFI to store bytes as the thumbnail File + link; skip encrypted / empty / corrupt / over-budget → keep file-type glyph; Source cover pin works once a raster exists (no schema change); existing `EnsureFileThumbnail` / list lookup. |
| **Out** | User-picked cover page; video / Office posters; Quick Look; Go PDF library; `cover_mode` / pin schema; OCR; new `PVThumbnail` chrome. |
| **Testable** | Image thumbs unchanged; PDF with a first page gets a `thumbnailRelPath`; bad PDF stays skipped + glyph; pin-as-cover accepts a PDF that has a raster; idempotent ensure. |
| **Depends on** | Shipped `derivatives.EnsureThumbnail` + `PVThumbnail` glyph fallback. **Not** S8-D1 / S8-01. |

Notes: [`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md).

---

## S8-99 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar) once the cluster is enough (or we stop adding stories). Record in [`completed.md`](completed.md); archive the spike. SemVer only if cutting a product release.

---

## Scope boundary

| In | Out |
| --- | --- |
| Composer transcription assist | Auto Observations / subjects / connect |
| Vision on **image** Artifacts | PDF OCR; audio / video OCR |
| In-memory crop from locator | Object-store crop files |
| Warn + proceed on large images | Hard reject / Apple “too many words” (does not exist) |
| PDF **page-1 thumbnail** via PDFKit | PDF OCR; Go PDF decoder; user-picked thumb page |
| More data-entry stories as added | Spike 7 leftover honesty/polish (conflicted, tray, pinning) unless pulled in |

---

## Gotchas

1. **Transcription ≠ Observation** — OCR dumps into the citation reading only ([`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4).
2. **Vision does not refuse a newspaper page** — it usually succeeds slowly or with junk. Large-page honesty is **our** preflight (pixels / no region / post-pass observation density), not a `VNError`.
3. **Locator y-down vs Vision ROI y-up** — crop in image pixels from [`ArtifactRegionGeometry`](../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift); do not pass a polygon into `regionOfInterest` (rect only).
4. **Crop the source raster**, not the zoomed viewport bitmap.
5. **No file middleman** — `ProjectFiles.objectURL` → image `NSImage` → `CGImage` → crop → `VNImageRequestHandler`.
6. **PDF is not an OCR input in S8-01** — disable Auto Transcribe; researchers paste. Image-only PDF scans wait for a later story (text layer or raster OCR).
7. **Hide Vision behind a protocol** — `FakeStore` / unit tests inject a recognizer.
8. **More stories do not wait on S8-01** unless they share the composer transcription chrome.
9. **S8-02 is macOS PDFKit → catalog derivative**, not `core/derivatives` learning to parse PDF. Windows keeps the glyph until a later engine renderer.

---

## Definition of done

- [ ] Checklist stories complete (or explicitly descoped)
- [ ] Dogfood bar items for landed stories met
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
