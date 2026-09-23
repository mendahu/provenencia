# PDF text find / search in the Artifact viewer

**Status:** idea only — not roadmapped.

## Problem

Citation composers (and later Source-page viewers) will open multi-page PDFs. Researchers often need to **jump to a known phrase** in a long scan or typed PDF (a surname, a volume label, a date string) without paging blindly.

## Idea

Add **Find** in the PDF Artifact viewer using PDFKit’s built-in search:

- `PDFDocument.findString` / `beginFindString` → `[PDFSelection]`
- Highlight matches in the viewer; next / previous; optional case-insensitive
- Jump page chrome to the page that owns the current hit

S7-06 ships page nav + zoom/pan via a page raster in `ArtifactMediaViewport`. Find almost certainly wants a **PDFKit-backed page** (or live `PDFView`) so selections map to real text — not OCR inventing text from a bitmap.

## Why later

Locators in Spike 7 focus on **page + region polygon**. Find is complementary navigation chrome, not a Citation locator type by itself (though hits could later seed `text_quote` if cheap).

## Open questions

- Find bar placement vs existing page/zoom tool strip (composer Frame 1).
- Image-only Artifacts and image PDFs with no text layer — honest empty / “no searchable text.”
- Whether Find is composer-only or also a future Source Artifact sheet.
- Interaction with region-draw mode (S7-07): disable Find while drawing, or keep both?

## Related

- Spike 7 viewers: [`deployment-plan/spike-7/deployment-plan.md`](../deployment-plan/spike-7/deployment-plan.md) **S7-06** / **S7-07**
- PDFKit: `PDFDocument.findString(_:withOptions:)`
- Parking lot cousin: transcription paste via **text selection** is in-scope for **S7-07**; full Find UI is not
