# Ship log — S10-35 Person gallery and Persons list thumb

**Plan:** [`S10-35-person-gallery.md`](S10-35-person-gallery.md) · **Branch:** `s10-35-person-gallery` · **PR:** #446
**State:** ready

## Round 0 — implement
Head `dcaa051b`. Gate PASS (orchestrator run against `79f7f042`). Report: no deviations; open details chosen — `likenesses.ForEntities` orders the gallery in Go, `PVList` gained a private designated init for an optional thumbnail, `CachedThumbnail` folds `fallbackMark` into `taskID`. Not run: §10 manual check.

## Round 1 — review
Base `79f7f042`, head `dcaa051b`. Verdicts: CON clean, MNT changes needed, UX clean, SEC clean, PRF changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1, SEC-1 | PRF, SEC | major | List, header and gallery thumbs decode full-resolution likeness crops (no pixel cap) | Fix now → F-1 (loader downsample); save-time derivative → Follow-up | verified: `ThumbnailCache.swift:47-53` is `NSImage(contentsOf:)`, lazy and unbounded; this PR adds the daily-path callers. Derivative touches the depiction write path, outside this plan |
| MNT-1 | MNT | major | Third and fourth copies of the likeness-save and image-id test helpers | Fix now → F-2 | verified: same SQL in 4 test packages. T-2b/T-3 said "package-local copy"; that is a test-table note, not a Decision; deviation is recorded in the Landed note and PR body. Handler tests stay out (plan Out line) |
| MNT-2 | MNT | major | `loadOutcomes` maps `Recorded` in two branches | Fix now → F-3 | verified: `conclusiondetails.go:390-399` |
| MNT-3, CON-5 | MNT, CON | major | DesignSystem README and `CachedThumbnail` doc stale after D-14/D-13/D-17 | Fix now → F-4 | verified: README:178 |
| UX-1 | UX | minor | Header says "No likeness recorded" when the lead image is unreadable | Fix now → F-5 | verified: `PersonDetailContent.thumbnail` always passes `personNoLikeness` as fallback label; few lines, wrong VoiceOver statement in code this PR adds |
| MNT-4 | MNT | minor | `CatalogConclusionDetail`'s doc now attached to `CatalogLikeness` | Fix now → F-6 | verified: `GenealogyStore.swift:608-619`; a two-line move of a defect this PR introduced |
| CON-1 | CON | minor | Landed note says no differences | Fix now → F-7 | verified: `taskID`, T-8 `imageLabel` nil, `PVList` private init |
| CON-2 | CON | minor | AC-1.2 test does not check weak value's order | Fix now → F-2 (rides along) | file touched by F-2 |
| CON-3 | CON | minor | AC-1.7 lead half untested; redraw subtest skips `finish` | Fix now → F-2 (rides along) | files touched by F-2 |
| CON-4 | CON | minor | `TestForEntities` is subtests, not a table | Follow-up | style only; F-2 touches the helpers, not the test shape |
| MNT-5 | MNT | minor | Lead choice inline in `composePersons` loop, no named slot | Follow-up → S10-39 | S10-39 changes this choice and can name it then |
| MNT-6 | MNT | nit | `tileBody` repeats `PVThumbnail` in both branches | Logged | — |

### Fix plan 1

**F-1 — Decode thumbnails at a bounded pixel size** (from PRF-1, SEC-1)
Files: `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`. Skill: `add-swift-test`.
Failing test first: `ThumbnailCacheTests` new case — write a PNG of 2000×1000 px to a temp project's `objects/…`, load it through `ThumbnailCache().image(projectDir:relPath:)`, expect the returned image's pixel size (longest edge of its `CGImage` representation) ≤ `ThumbnailCache.maxPixelSize`.
Do: add `static let maxPixelSize = 512` (covers the 120pt header thumb at 2x with margin). In the detached task replace `NSImage(contentsOf: url)` with `CGImageSourceCreateWithURL` + `CGImageSourceCreateThumbnailAtIndex(src, 0, [kCGImageSourceCreateThumbnailFromImageAlways: true, kCGImageSourceCreateThumbnailWithTransform: true, kCGImageSourceShouldCacheImmediately: true, kCGImageSourceThumbnailMaxPixelSize: maxPixelSize])`, wrapped as `NSImage(cgImage:size:)` at the pixel size. This decodes eagerly off the main actor. `import ImageIO` as `ArtifactRegionCrop.swift` does. Update the type doc. Do not change keys, LRU, or `cachedImage`.
Check: the test passes; existing ThumbnailCache tests pass; scripts/check-pr.sh PASS.

**F-2 — Share the likeness test helpers in `observationstest`** (from MNT-1, CON-2, CON-3)
Files: `core/database/observations/observationstest/observationstest.go`, `core/database/likenesses/likenesses_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusionheaders/conclusionheaders_test.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Failing test first: none (refactor); CON-2/CON-3 assertions are new and should pass on today's code.
Do: add `SaveLikeness(t, c, userID, sourceID, subjectID []byte, shade uint8) observations.Observation` (`FileArtifact` + `SaveCitation` with one `LikenessRegion`/`Crop(t, 40, 30, shade)` depiction input — the body of today's `likenessOn`) and `ImageOf(t, c, observationID []byte) []byte` (the shared `depiction_values` join). Replace `likenessOn`/`imageOf`/inline copies in the four test files with them; delete the local copies. Do not touch `api/ffi/handlers` tests. CON-2: in "a low-trust Source is weak", assert `g[0].ObservationIDs[0] == w.oa.ID` and `g[1].ObservationIDs[0] == w.ob.ID`. CON-3: add `finish(t, w, "redraw")` to "a redrawn face keeps its place"; in `TestForEntityLikenessNeedsAPNG`, after the `UPDATE`, `f.c.Graph().Drop()` then `conclusionheaders.PersonsByIDs` for the handle and assert `LikenessImageFileID` is nil (put it in whichever test package avoids an import cycle; if none does, put this assertion in `conclusionheaders_test.go` as its own case).
Check: `go test -tags fts5` on the four packages passes; scripts/check-pr.sh PASS.

**F-3 — One `Recorded` mapping in `loadOutcomes`** (from MNT-2)
Files: `core/database/conclusiondetails/conclusiondetails.go`.
Do: move the three `o.Recorded.Text/Integer/TermID` assignments above `f, ok := byProp[...]`, beside `o.ValueRank`; the `!ok` branch keeps only the `leads` check and `continue`.
Check: `go test -tags fts5 ./core/database/conclusiondetails/` passes; scripts/check-pr.sh PASS.

**F-4 — Correct the thumbnail docs** (from MNT-3, CON-5)
Files: `macos/App/DesignSystem/README.md`, `macos/App/Platform/CachedThumbnail.swift`.
Do: README `PVThumbnail` line → "optional `colors:`: tint and line on mark and symbol states; the line also outlines an image (S10-23, S10-35)". README `PVList` line → "thumbnail slot (a view; S10-35)". `CachedThumbnail` first doc line: add "fallback mark" to the state list.
Check: scripts/check-pr.sh PASS.

**F-5 — Name an unreadable lead "Likeness", not "No likeness recorded"** (from UX-1)
Files: `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Failing test first: in `headerThumbIsTheLeadLikeness`, a Person with a non-empty `likenessImageRelPath` → `thumbnail` fallback label is `L10n.Conclusions.personLikenessThumb`; with an empty path it stays `personNoLikeness`.
Do: in `thumbnail`, pass `personLikenessThumb` as the fallback label when `likenessImageRelPath` is non-empty. No new string.
Check: the test passes; scripts/check-pr.sh PASS.

**F-6 — Re-attach `CatalogConclusionDetail`'s doc** (from MNT-4)
Files: `macos/App/Platform/GenealogyStore.swift`.
Do: move `CatalogLikeness` and its own two doc lines above the "One handle's detail (S9-15)…" doc so that doc sits directly on `CatalogConclusionDetail`.
Check: scripts/check-pr.sh PASS.

**F-7 — List the differences in the Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-10/prs/S10-35-person-gallery.md`.
Do: replace "No differences from this plan" with: `taskID` also folds in `fallbackMark` (D-13 said unchanged); T-8's plain thumb has `imageLabel` nil, not "Likeness"; `PVList` has a private `init(optionalThumbnail:)` (D-17 did not name it); the likeness test helpers are shared in `observationstest` rather than package-local (T-2b, T-3); `ThumbnailCache` decodes at ≤ 512 px (review PRF-1/SEC-1).
Check: `scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0b08d414`. Gate PASS (orchestrator run). F-1 to F-7 done. F-1 and F-5 tests ran red first. F-3 was committed inside F-2's commit (`3cf897e7`). Implementer added one line to `completed.md` (512 px decode) and committed this ship log in `0b08d414`, against instructions; content unchanged.

## Round 2 — review
Scope `dcaa051b...0b08d414`. Lenses: all five (each had a Fix-now finding).
Verdicts: CON clean, MNT changes needed (MNT-1 remnant), UX clean, SEC clean, PRF clean. No blockers.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1, SEC-1 | PRF, SEC | — | Bounded decode | Resolved | `ThumbnailCache.swift:58-67`; `decodesAtABoundedPixelSize` |
| MNT-1 | MNT | minor (remnant) | One inline `ImageOf` query left in `TestForEntityLikenessNeedsAPNG` | Follow-up | `conclusiondetails_test.go:276-278`; one-line swap; round 2 opens blockers only |
| MNT-2, MNT-3, MNT-4 | MNT | — | — | Resolved | verified by MNT round 2 |
| UX-1 | UX | — | — | Resolved | `PersonDetailContent.swift:32-33`, tested both branches |
| CON-1, CON-2, CON-3, CON-5 | CON | — | — | Resolved | verified by CON round 2 |
| PRF note | PRF→UX | — | 512 px cap may soften the artifact well | Rejected | the well loads `thumbnailRelPath`, a ≤256 px derivative (`ensure.go:60`); UX and CON round 2 confirm |
| MNT-7 | MNT | nit | `conclusionheaders_test.go:329-332` closures only forward to shared helpers | Logged | — |

No Fix plan 2.

## Result
Gate PASS at `0b08d414` · CI pending · open: none · follow-ups: depiction save-time thumbnail derivative (PRF-1), MNT-1 remnant, CON-4, MNT-5 (S10-39) · to user: —
