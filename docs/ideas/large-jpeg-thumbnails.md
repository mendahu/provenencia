# Thumbnails for oversized JPEGs

**Status:** idea only — not roadmapped. A 39 MB, 11,000×18,000 JPEG artifact saves and shows the missing-image glyph. No error.

## What happens

Thumbnail generation is in Go, in [`core/derivatives`](../../core/derivatives/). [`EncodeJPEG`](../../core/derivatives/raster/raster.go) reads the header, then returns `ErrTooLarge` when `width * height` is above `maxDecodePixels` (64,000,000) and never decodes pixels. This scan is 198 megapixels. A full RGBA decode would be about 792 MB. The cap was added so decompression bombs never allocate that (`788f14ad`).

[`generate`](../../core/derivatives/ensure.go) maps `ErrTooLarge` to `ErrUnprocessable`. [`thumbnailRelPath`](../../api/ffi/handlers/thumbnails.go) treats that as success with an empty path. The file is already stored. The glyph is `file_image_missing`. “Use as thumbnail” stays hidden because it needs a non-empty `thumbnailRelPath`.

Ingest only copies bytes. This is not an upload regression. Every later artifact response calls `EnsureThumbnail` again, re-reads the object (39 MB is under the 128 MiB source cap), and rejects it until a derivative row exists. A decoder that accepts the file would fill the thumbnail the next time the Source is opened. No re-upload.

PNG, GIF, WebP, TIFF, and BMP over the same budget skip the same way. Those formats cannot be decoded at a reduced size with the stdlib, and a header can declare a huge frame in a few bytes. Leave that refusal in place.

## Suggested fix

Keep the 64-megapixel full decode for every non-JPEG, and for JPEGs that fit. For a JPEG over the budget, downsample on darwin with ImageIO: `CGImageSourceCreateThumbnailAtIndex` and `kCGImageSourceThumbnailMaxPixelSize` set to `MaxEdge` (256 for the default thumbnail). The output buffer is bounded by that edge. Then the existing scale-and-`jpeg.Encode` path stores the usual unaudited `file_derivatives.thumbnail` row.

- Do not apply EXIF orientation on this branch. The stdlib JPEG path does not, and only oversized JPEGs would take it.
- Do not raise `maxDecodePixels` and full-decode. That is the allocation the cap blocks.
- Do not add libjpeg. The macOS dylib is already cgo. ImageIO is a system framework, so the link line is `-framework ImageIO -framework CoreGraphics -framework CoreFoundation`. No new module, and [`scripts/build-macos-core.sh`](../../scripts/build-macos-core.sh) does not grow a third-party C library.
- If ImageIO cannot produce a thumbnail, return `ErrTooLarge` so the handler still skips. A decode error that is not `ErrTooLarge` fails the artifact RPC after the file is saved.
- Non-darwin builds keep returning `ErrTooLarge`. The shipping core is the macOS dylib. A Windows build needs its own downsample (WIC, or a shared libjpeg backend) or those JPEGs stay on the glyph.

A test can encode one real JPEG just over 64 megapixels (about 8,200×8,200, quality 1) and expect a long edge of 256. That allocates roughly 270 MB once. Do not fixture the 39 MB scan. Existing PNG and GIF header bombs should still return `ErrTooLarge`.

## Open questions

- Is a darwin-only ImageIO path acceptable until Windows, or should the first fix be a portable decoder (libjpeg-turbo, statically linked into the dylib)?
- Should a failed oversized JPEG be remembered so later Source opens do not re-read 39 MB and reject it again? Today nothing is stored, so every artifact response retries.
- EXIF orientation is ignored for every JPEG thumbnail, including ones under the budget. Fixing only the oversized branch would make large scans the ones that look right. Worth a separate pass, or left alone?
- Is 256 px on the long edge enough for an 11,000×18,000 scan used as a Source cover, or does the cover want a larger derivative than the row thumbnail?
