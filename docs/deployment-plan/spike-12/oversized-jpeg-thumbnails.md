# Thumbnails for oversized JPEGs

Design for S12-39. Pulled in from `ideas/large-jpeg-thumbnails.md` (user, 2026-10-27), which proposed a macOS-only ImageIO path; the user chose a portable, purpose-built Go decoder instead, so Windows reuses it. PR order and status are in the [map](deployment-plan.md).

## Problem

A 39 MB, 11,000 × 18,000 JPEG Artifact saves and shows the missing-image glyph, with no error. `core/raster` `EncodeJPEG` reads the header and returns `ErrTooLarge` when `width × height` is over `maxDecodePixels` (64,000,000), without decoding pixels. That cap keeps a decompression bomb from allocating; this scan is 198 megapixels, about 792 MB as RGBA. `derivatives` maps `ErrTooLarge` to `ErrUnprocessable`, `thumbnailRelPath` treats it as success with an empty path, and every later Artifact response tries again: `EnsureThumbnail` re-reads the 39 MB object and refuses it until a derivative row exists. "Use as thumbnail" stays hidden because it needs a non-empty path.

## Shape: a reduced decode that reads only DC coefficients

A thumbnail needs at most 256 px on its long edge, and a JPEG already stores each 8 × 8 block's average as its DC coefficient. Reading only the DC coefficients gives the image at 1/8 scale with no inverse DCT and no full-size buffer: the 198 MP scan becomes a 1,375 × 2,250 image (about 3 MP), which the existing scale-and-encode path turns into the usual thumbnail.

- **Where.** A small decoder in `core/raster`, used by `EncodeJPEG` only when the format is JPEG and the declared size is over `maxDecodePixels`. JPEGs within the budget keep the standard-library path, so nothing that works today changes. `EncodePNG` and staged images never take it.
- **What it reads.** Markers SOI, SOF0, SOF1, SOF2, DHT, DQT, DRI, SOS, EOI, skipping APPn and COM. Baseline: Huffman-decode every block, keep the DC value, discard the AC values (they must still be decoded to reach the next block). Progressive: read the DC scans (the first scan and any DC refinement), skip the AC scans' bytes without decoding them. Restart intervals. One to three components with any sampling factors; chroma is upsampled to the luma grid; YCbCr to RGB as the standard library does. A pixel is `clamp(DC × Q[0] / 8 + 128)`.
- **What it refuses.** Arithmetic coding, 12-bit samples, lossless JPEG, and CMYK or YCCK (four components) return `ErrTooLarge`, as today, so the handler still skips them. They are rare for scans.
- **Bounds.** The 1/8 image counts against `maxDecodePixels`; every count read from the file (components, tables, scan headers, restart interval) is checked before use; a truncated or malformed stream returns an error, never a partial image. The decoder holds the decode slot `raster` already takes.
- **Tests.** Go native fuzzing over the decoder (seed corpus from the standard library's test JPEGs; the seeds run in CI). A JPEG just over 64 MP built in the test from a lazy `image.Image` that computes each pixel, so the fixture never needs a full buffer, expects a 256 px long edge. A DC-decoded image of a small baseline and a small progressive JPEG matches the standard library's decode downscaled by 8, within a tolerance. PNG and GIF header bombs still return `ErrTooLarge`.

About 400–700 lines of Go and its tests, in one PR. If the plan runs past about ten steps, progressive support splits into its own PR.

## Not this design

- **EXIF orientation.** The standard-library path ignores it for every JPEG; fixing only oversized ones would make large scans the only ones that look right.
- **A larger cover derivative.** Whether a 256 px thumbnail is enough for a Source cover from such a scan.
- **Other formats over the budget.** PNG, GIF, WebP, TIFF and BMP cannot be decoded at reduced size this cheaply, and a header can declare a huge frame in a few bytes; their refusal stays.
- **The Mac viewer.** Opening the original in the Artifact viewer is S12-28's bounded decode.
