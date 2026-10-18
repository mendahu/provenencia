# S10-61 — Cap region size

**Status:** Planned
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the freeform tool ignores a click past the cap with the system beep; no control, copy, or layout changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-61-region-caps`
**Version bump:** none — Spike 10 and Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [interpretation layer §3.3 region selector](../../../interpretation-layer-data-model.md#33-region-selector), [§3.8 locator invariants](../../../interpretation-layer-data-model.md#38-locator-invariants), [Spike 10 decisions](../../archive/spike-10/README.md) ("A depiction is an owned value row … Go checks"), working rules: "Spike 10 decisions hold"
**Planned:** 2026-10-18 against `main` @ `aa60b63`
**Review:** [`S10-61-region-caps.review.md`](S10-61-region-caps.review.md) — Fits with changes / Ready with changes, 1 round

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A locator document someone else wrote can no longer stall the app. Today the region check has no limit: it compares every polygon edge with every other edge, and it runs on every read of a depiction (Source load, composer list, the auto-reconciler cache rebuild at project open). A stored region of 20,000 points takes 0.93 s per check, and about 200,000 points would take roughly 90 s per value while every other call on the project waits. This PR caps a locator document at 256 region points and 64 KiB, in Go and in the Swift mirror, so the cost of one check is bounded at about 0.3 ms.

**Why now.** Audit finding SEC-1 (2026-10-18 likeness and depiction area audit, security lens): a shared catalog can hang project open.

**What stays.** Every region the app's own tools draw stays valid: a rectangle has 4 points, an L shape 6, a circle 32, and a freeform polygon one point per click. The error code for a refused locator stays `locator.invalid`. A stored depiction region that fails the check is hidden, as an off-page region already is.

## 2. User stories

**US-1.** As a researcher who opens a project someone else made, when one of its Citations or likenesses stores an oversized region, the project opens and its Sources load without a stall, and that region is treated as invalid.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `locator.Validate` on a document whose region selectors hold 257 points in total, under 64 KiB | returns `locator.ErrInvalid` |
| AC-1.2 | `locator.Validate` on a document of 256 points in one region selector (a convex ring) | returns nil |
| AC-1.3 | `locator.Validate` on a document longer than 65,536 bytes after trimming | returns `locator.ErrInvalid` without unmarshalling |
| AC-1.4 | `validateDoc` on a decoded document holding one `normalized` region selector of 200,000 points | returns `ErrInvalid` in under 100 ms (the uncapped check takes about 90 s) |
| AC-1.5 | A stored likeness region of 257 points on a valid Citation locator is read | `ValidateDepictionRegion` returns `ErrInvalid`, so `Readable` (which empties `RegionJSON` on any error from it, proved by the existing `TestFromJoin` "off-page region" row) hides the region and keeps the image |
| AC-1.6 | A valid likeness region on a Citation locator whose region has 257 points is read | `ValidateDepictionRegion` returns `ErrInvalid`, so `Readable` hides the region and keeps the image |
| AC-1.7 | A valid likeness region (page 1) on a Citation locator with a page 1 selector and two region selectors of 129 and 128 points | `ValidateDepictionRegion` returns `ErrInvalid`: the cap counts the Citation locator's points across its selectors |

**US-2.** As a researcher drawing a freeform region, I cannot place more points than Go accepts, so a region I draw always saves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `ArtifactRegionGeometry.isValidPolygon` on 257 points | false, before the crossing test runs |
| AC-2.2 | `ArtifactRegionGeometry.isValidPolygon` on a convex ring of 256 points | true |
| AC-2.3 | `ArtifactRegionGeometry.appendFreeformVertex` on 255 points, then on 256 points | 256 points come back; then nil. The overlay beeps on nil and adds nothing; clicking the first point still closes the polygon, because the close branch runs first |
| AC-2.4 | The Swift cap constant and the Go cap constant | equal; a row in the Go Swift-mirror table fails when they differ |
| AC-2.5 | A Citation whose stored locator region has 257 points is opened in the composer | the region chip is absent (the draft decodes with no region), as for any invalid stored region today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `locator.MaxRegionPoints` = 256 and `locator.MaxDocumentBytes` = 65,536 *(new)*: the byte cap in `decode`, the point sum in `validateDoc` before any selector's area or crossing test | STEP-1, STEP-2, AC-1.1–AC-1.4, AC-1.7 |
| Read path: no code change; tests prove `ValidateDepictionRegion` refuses an oversized stored region and a region on an oversized Citation locator, which `Readable` already hides | STEP-2, AC-1.5, AC-1.6 |
| Swift `ArtifactRegionGeometry.maxRegionPoints` = 256 *(new)*, checked first in `isValidPolygon` | STEP-3, AC-2.1, AC-2.2, AC-2.5 |
| `ArtifactRegionGeometry.appendFreeformVertex` *(new)*; the freeform tool appends through it and beeps on nil | STEP-4, AC-2.3 |
| A row in `TestSwiftVocabularyMatchesGo` that keeps the Swift constant equal to the Go one; the file header and the `macos-client.mdc` mirror line widened to name mirrored limits | STEP-5, AC-2.4 |
| Domain doc: the cap in §3.3 and §3.8 | STEP-6 |

| Out | Home |
| --- | --- |
| Caching `DepictionValueDraft.hasRegion` instead of re-checking the polygon on every read (audit PRF note) | not planned: with the cap, one check costs at most about 0.3 ms |
| Swift-side byte cap on `CitationLocatorDraft.decode` | not planned: Go returns only locators it stored, decoding is linear, and the quadratic part is capped by `maxRegionPoints` |
| Swift counting points across several region selectors of one Citation locator (D-5) | not planned: the app writes one region per locator; only a hand-made locator holds several |
| A drift row for `ingest.MaxBytes` / `IngestMediaPolicy.maxBytes` | not planned in this folder: it needs a regex for `512 << 20`; it is the next row the widened mirror table takes (§6 Next adds) |
| Symlinked objects and `ThumbnailCache` decode limits (audit SEC-2) | [S10-62](S10-62-object-symlinks.md) and a later item from the audit action list |
| Decode limit in bytes in `core/raster` (audit SEC-3) | not planned in this folder; audit action list (Low) |
| A dedicated error code or copy for an oversized region | not planned: the app's tools cannot produce one, so only a hand-made catalog meets it |
| Repairing stored oversized locators | not planned: they stay in the catalog and read as invalid. Opening such a Citation in the composer and saving it rewrites the locator without the region, as it already does for any invalid stored region (`CitationLocatorDraft.decode` drops it, `encodeJSON` writes what is left, `citations.Update` validates only the incoming locator). That is not a bug in this PR |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Envelope check | `core/locator/validate.go` `decode` | Trims, unmarshals with no size limit, checks version 1 and at least one selector |
| Selector loop | `core/locator/validate.go` `validateDoc` | Checks each selector by type; region selectors go to `validateRegion`, its only caller |
| Region check | `core/locator/validate.go` `validateRegion` | Unit, at least 3 points, no repeated closing point, points in 0–1, area, then `selfIntersects`. No point limit |
| Crossing test | `core/locator/validate.go` `selfIntersects` | Every edge against every non-adjacent edge: cost grows with the square of the point count. Measured at `aa60b63`: 256 points 0.28 ms, 1,024 points 2.7 ms, 2,048 points 10 ms, 20,000 points 0.93 s |
| Depiction check | `core/locator/depiction.go` `ValidateDepictionRegion`, `regionDocPages` | `regionDocPages` decodes and validates the region document and requires exactly one region selector; then the Citation locator is decoded and passed to `validateDoc`; then pages are compared. Called on write by `core/database/observations/depiction.go` `resolveDepiction` |
| Citation write | `core/database/citations/citations.go` `normalizeCreateInput`, `Update` | `locator.Validate(in.LocatorJSON)` |
| Read path | `core/database/depictionvalues/depictionvalues.go` `Readable`, `FromJoin` | `Readable` empties `RegionJSON` whenever `ValidateDepictionRegion` returns an error. Callers: `core/database/observations/observations.go` (Observation list), `core/database/sourcefill/load.go` (Source load), `core/database/autoreconciler/autoreconciler.go` (cache rebuild, run from `EnsureCatalog` at open when `cache_version` differs) |
| Swift mirror | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `isValidPolygon`, `selfIntersects` | Same rules as Go, same quadratic crossing test, no point limit. `circleRingCount` = 32, `freeformMinimumVertices` = 3 |
| Freeform vertex rules | same file `deleteFreeformVertex(points:index:)` | Owns the minimum-vertex rule: returns nil at `freeformMinimumVertices`. Tested by `ArtifactRegionGeometryTests.freeformKeepsAtLeastThreeVertices` |
| Swift callers | `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift` `ArtifactRegionDraft.isValid`; `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` `decodeRegion`; `macos/App/Features/CitationComposer/DepictionValueDraft.swift` `hasRegion` | `decodeRegion` drops a region that fails `isValidPolygon`, once per region selector (the last valid one wins); `hasRegion` re-runs it on each read |
| Freeform tool | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` `handleFreeformClick` (private), `mouseDragged` | One point per click; `mouseDragged` adds nothing for freeform. Closing on the first point is checked before `inProgressPoints.append`. No limit, no test |
| Swift mirror test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, `swiftEnumBody`, path constants | One table `{name, path, enum, re, want}` over Swift enums; `swiftEnumBody` returns an enum's source up to its column-0 `}` (`enum ArtifactRegionGeometry` closes before `enum ArtifactResizeCursorKind`). Header comment names only wire vocabulary |
| Mirror rule | `.cursor/rules/macos-client.mdc` (line "Go wire vocabulary … `api/ffi/swift_vocabulary_test.go` fails if that file drifts from Go") | Names wire strings only |
| Shared Swift fixtures | `macos/ProvenenciaTests/ArtifactFileFixtures.swift` `ArtifactFileFixtures` | Files on disk; no geometry fixture exists |
| Tests | `core/locator/validate_test.go` `TestValidate`; `core/locator/depiction_test.go` `TestValidateDepictionRegion` (its `doc(...)` builder is a closure local to that test); `core/database/depictionvalues/depictionvalues_test.go` `TestFromJoin` ("off-page region"); `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`; `macos/ProvenenciaTests/CitationLocatorDraftTests.swift` | No size case anywhere |

**Precedent.** `ac70707` ("Hide a stored depiction region that no longer validates when Observations are listed.") made the read path empty a region that fails the check; this PR relies on it and adds only the cap. `deleteFreeformVertex` is the shape for `appendFreeformVertex`. `TestSwiftVocabularyMatchesGo` is the table the mirror row joins.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `MaxRegionPoints = 256` *(new, exported)* in `core/locator`, counted across **all** region selectors of one document. Its doc comment says "per document" | The app's largest shape is the 32-point circle; 256 freeform clicks is far past any face or record outline. 256 points cost 0.28 ms per check, so a read that checks a region and its Citation locator stays under 1 ms. Counting per document stops many capped selectors in one document from multiplying the cost | 1,024 (2.7 ms per check, so 1,000 crafted depictions cost about 5 s at open); a per-selector cap (a document of many capped selectors costs that many times more) |
| D-2 | `MaxDocumentBytes = 65536` *(new, exported)*: `decode` refuses a trimmed document longer than this before `json.Unmarshal` | 256 points serialize to about 12 KB at 17 significant digits; 64 KiB leaves room for page and unknown selectors and keeps unmarshalling bounded | No byte cap (unmarshal cost grows with whatever is stored); 16 KiB (too close to 256 full-precision points plus a `text_quote` a future client preserves) |
| D-3 | One point check: `validateDoc` sums `len(sel.Points)` over selectors whose trimmed type is `region`, before its selector loop, and returns `ErrInvalid` over `MaxRegionPoints`. `validateRegion` gets no second check | `validateDoc` is `validateRegion`'s only caller and every public path (`Validate`, `ValidateDepictionRegion`) goes through it, so a check inside `validateRegion` could never fire (review M-2). Summing first means no quadratic pass starts on an oversized document | A second check in `validateRegion` (dead from the day it lands); the sum in `decode` (it would duplicate the selector-type switch) |
| D-4 | An oversized document returns the existing `locator.ErrInvalid` (`locator.invalid`) | The app's tools cannot produce one, so the only source is a hand-made catalog; the existing copy covers it | A new code and string (copy nobody using the app reaches) |
| D-5 | Go counts per document; Swift `isValidPolygon` counts per polygon. The Swift constant's comment says: "Matches `locator.MaxRegionPoints`, which Go counts across a whole locator; the app writes one region per locator, so a per-polygon cap equals Go's for anything the app writes." The §3.3 doc line uses the same words | Swift's gate sees one polygon at a time, and the client never writes a second region. A hand-made Citation locator with several regions under the cap each but over it in total shows a chip in the composer while Go hides its likeness regions; that case is Out | Counting across selectors in `CitationLocatorDraft.decode` (a second counting rule for a case only a hand-made catalog reaches); renaming to `MaxDocumentRegionPoints` (the comment carries the meaning; the shorter name matches `maxRegionPoints`) |
| D-6 | No change to `Readable`, `FromJoin`, or their callers | They already empty a region that fails `ValidateDepictionRegion` (existing `TestFromJoin` "off-page region"), and the cap makes that call cheap. A stored region over the cap reads as hidden; a Citation locator over the cap hides every depiction region on that Citation | A second, cheaper read-only check (two rules for one document); new `TestFromJoin` rows (they would retest the existing branch with a second Go ring builder, review M-4) |
| D-7 | Swift: `ArtifactRegionGeometry.maxRegionPoints = 256` *(new)* beside `circleRingCount`; `isValidPolygon` starts with `guard points.count >= 3, points.count <= maxRegionPoints else { return false }` | One mirror of D-1; `decodeRegion`, `ArtifactRegionDraft.isValid`, and `hasRegion` all go through `isValidPolygon`, so each gains the cap | A cap in each caller |
| D-8 | `ArtifactRegionGeometry.appendFreeformVertex(points: [CGPoint], _ point: CGPoint) -> [CGPoint]?` *(new)* beside `deleteFreeformVertex`: returns nil when `points.count >= maxRegionPoints`, else `points + [point]`. `handleFreeformClick` keeps its close-on-first-point branch first; then `guard let next = ArtifactRegionGeometry.appendFreeformVertex(points: inProgressPoints, normalized) else { NSSound.beep(); return }`, `inProgressPoints = next`, `needsDisplay = true` | The minimum and maximum vertex rules live in one type and are unit-tested the same way (review M-1); the researcher can still close the polygon; the beep is the AppKit signal for a refused action and needs no copy | The check inline in the private handler (untestable); committing automatically at the cap (closes a shape the researcher did not finish); a toast or callout (new chrome and copy for a case nobody reaches by hand) |
| D-9 | The mirror test is a row in `TestSwiftVocabularyMatchesGo`: a `swiftRegionGeometry` path constant, `enum: "ArtifactRegionGeometry"`, `re: regexp.MustCompile(`static let maxRegionPoints = (\d+)`)`, `want: []string{strconv.Itoa(locator.MaxRegionPoints)}`. The file's header comment becomes "The app reads Go's wire vocabulary, and mirrors some of Go's limits, through Swift files this test reads", and the `macos-client.mdc` line gains "and the limits Swift mirrors (`ArtifactRegionGeometry.maxRegionPoints`)". The exact final text of the `macos-client.mdc` line and the test file's header comment is fixed in [S10-64](S10-64-thumbnail-decode-guard.md) D-10; use it word for word, naming only the mirrored constants that exist on `main` after the merge. | One table for every Swift-mirrors-Go fact (`go-tests.mdc`: one test per family); the existing `matches` and `swiftEnumBody` helpers apply unchanged; no match fails the row through the existing comparison | A separate `TestSwiftRegionCapMatchesGo` (a second shape for the same job, review M-3); a comment only (drifts silently) |
| D-10 | AC-1.4's test calls `validateDoc` directly (package `locator`) on `document{Version: 1, Selectors: []selectorRaw{{Type: TypeRegion, Unit: "normalized", Points: ringPoints(200_000)}}}` and asserts `ErrInvalid` within 100 ms | The document is built in memory, so the byte cap in `decode` cannot short-circuit it (a 200,000-point JSON is about 3.6 MB); `Unit` is set so the cap, not the unit check, refuses it; the uncapped check takes about 90 s, so the bound cannot flake | A benchmark (does not fail); calling `validateRegion` (the cap is not there, D-3); counting calls to `segmentsIntersect` (needs a test hook in production code) |
| D-11 | Test helpers: Go `ringPoints(n int) []point` and `ringRegion(n int) string` *(new)* in `core/locator/validate_test.go` (package `locator`, shared by `depiction_test.go`): n points on a circle of radius 0.4 around (0.5, 0.5), counter-clockwise from angle 0; `ringRegion` returns the selector object `{"type":"region","unit":"normalized","points":[…]}` with coordinates formatted by `strconv.FormatFloat(v, 'f', 6, 64)`. Documents are assembled in each test with `fmt.Sprintf(`{"version":1,"selectors":[%s]}`, …)`. Swift `RegionFixtures.ring(count: Int) -> [CGPoint]` and `RegionFixtures.locatorJSON(ringCount: Int) -> String` *(new)* in `macos/ProvenenciaTests/RegionFixtures.swift`, same circle, the JSON an artifact selector plus one region selector, built with `JSONSerialization` | One ring builder per language (review M-4); 6 decimals keep 256 points far under 64 KiB and distinct; the Swift fixture sits beside `ArtifactFileFixtures`, the shared-fixture home | A local copy in each test file; a builder in `depictionvalues` tests (T-6 dropped, D-6) |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `locator.MaxRegionPoints`, `locator.MaxDocumentBytes` *(new)* | const | `core/raster` `MaxEdgeLimit`; `core/ingest/stage.go` `stageMaxEdge` | Those cap images, not locator documents; the cap belongs in the package that owns locator validation (searched `rg "Max[A-Z]\w* *=" core`) |
| `ArtifactRegionGeometry.maxRegionPoints` *(new)* | const | `ArtifactRegionGeometry.circleRingCount`, `freeformMinimumVertices`; `IngestMediaPolicy.maxBytes` (the existing numeric Go-to-Swift mirror, "matches `ingest.MaxBytes`") | It sits beside the other geometry constants; there is no existing point limit in Swift (searched `rg "maxRegion\|maxPoints" macos`) |
| `ArtifactRegionGeometry.appendFreeformVertex` *(new)* | func | `ArtifactRegionGeometry.deleteFreeformVertex` | Its sibling: same shape, the opposite bound |
| Mirror row "region point cap", `swiftRegionGeometry` *(new)* | test row, const | `TestSwiftVocabularyMatchesGo` rows and path constants | Joins the existing table (D-9) |
| `ringPoints`, `ringRegion` *(new)* | test helper | `TestValidateDepictionRegion`'s local `doc(...)` closure | That closure builds a page + region document from fixed points; it is local to one test and cannot make an n-point ring |
| `RegionFixtures` *(new)* | test helper file | `macos/ProvenenciaTests/ArtifactFileFixtures.swift` | That enum writes files on disk; geometry is another concern, so it gets its own small fixture enum beside it |

**Builds on.** `ac70707`'s read rule (`Readable` hides a region that fails the check); `deleteFreeformVertex`; the Swift-mirror table in `api/ffi/swift_vocabulary_test.go`.
**Deletes.** None — nothing becomes dead.
**Next adds.** A drift row for `IngestMediaPolicy.maxBytes` against `ingest.MaxBytes` joins the same table. The idea file's polygon crop and read-only region layer ([`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md)) draw many saved regions at once; the cap bounds each one. A later vertex-adding path (insert on an edge, paste) goes through `appendFreeformVertex`.

### Schema

None — no catalog change. Stored locators stay as written.

### Go core

| Package | Change |
| --- | --- |
| `core/locator` | `const MaxRegionPoints = 256` and `const MaxDocumentBytes = 65536` *(new)*, doc comments citing D-1 ("per document") and D-2. `decode`: after `strings.TrimSpace`, `len(locatorJSON) > MaxDocumentBytes` → `ErrInvalid`. `validateDoc`: before the selector loop, sum `len(sel.Points)` over selectors whose `strings.TrimSpace(sel.Type) == TypeRegion`; over `MaxRegionPoints` → `ErrInvalid`. `validateRegion` unchanged |

Errors: `locator.invalid` → a document over either cap, on write (Citation create and update, depiction save) and on read (the region is hidden by `Readable`).

### FFI

None — no message or handler changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Point cap | `ArtifactRegionGeometry.maxRegionPoints` *(new)* | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` | `static let maxRegionPoints = 256`, one line beside `circleRingCount`, with the D-5 comment |
| Polygon check | `ArtifactRegionGeometry.isValidPolygon` | same file | D-7 guard before the closing-point check |
| Freeform append | `ArtifactRegionGeometry.appendFreeformVertex` *(new)* | same file, after `deleteFreeformVertex` | D-8 |
| Freeform stop | `ArtifactRegionOverlay.handleFreeformClick` | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` | D-8 |

States: a stored region over the cap reads as no region (composer: no region chip; the likeness edit dialog shows `L10n.CitationComposer.depictionCaptionHidden`, as today for an off-page region). Accessibility: no change.

### Strings

None — no new copy.

## 7. Steps

The Xcode steps and `scripts/check-pr.sh` run on the user's Mac, as `ship-pr` sets up; the Go checks run anywhere.

**STEP-1 — Byte cap in `decode`**
Files: `core/locator/validate.go`, `core/locator/validate_test.go`.
Do: add `MaxDocumentBytes` and the check in `decode` (D-2). Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/locator/...` passes.

**STEP-2 — Point cap in `validateDoc`**
Files: `core/locator/validate.go`, `core/locator/validate_test.go`, `core/locator/depiction_test.go`.
Do: add `ringPoints` and `ringRegion` (D-11). Add T-4 and run it once before the cap: it fails (no error, about 90 s). Then add `MaxRegionPoints` and the document sum in `validateDoc` (D-1, D-3). Add T-1, T-2, T-5. Before adding them, search for fixtures over the cap: `rg -n '"points"' api core macos/ProvenenciaTests` (includes non-`_test` helpers such as `core/database/observations/observationstest/observationstest.go`); none at the planning base builds more than 256 points.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/locator/... ./core/database/depictionvalues/... ./core/database/observations/... ./core/database/sourcefill/... ./core/database/autoreconciler/... ./core/database/citations/... ./api/ffi/handlers/...` passes; T-4 runs in under 1 s.

**STEP-3 — Swift cap in `isValidPolygon`**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift`, `macos/ProvenenciaTests/RegionFixtures.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`, `macos/ProvenenciaTests/CitationLocatorDraftTests.swift`. Skill: `add-swift-test` (wire `RegionFixtures.swift` into the ProvenenciaTests target as that skill says).
Do: add `maxRegionPoints` with the D-5 comment and the D-7 guard. Add `RegionFixtures` (D-11). Add T-7 and T-8.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactRegionGeometryTests -only-testing:ProvenenciaTests/CitationLocatorDraftTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-4 — Freeform stops at the cap**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift`, `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift`, `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`.
Do: add `appendFreeformVertex` and route `handleFreeformClick` through it (D-8). Add T-10.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactRegionGeometryTests CODE_SIGNING_ALLOWED=NO` passes; manual check step 2.

**STEP-5 — Swift and Go caps stay equal**
Files: `api/ffi/swift_vocabulary_test.go`, `.cursor/rules/macos-client.mdc`.
Do: add the `swiftRegionGeometry` path constant (`"../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift"`), the table row, the `strconv` and `core/locator` imports, the widened header comment, and the widened rule line (D-9, T-9).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; changing the Swift value to 255 locally makes the "region point cap" row fail (revert after).

**STEP-6 — Docs**
Files: `docs/interpretation-layer-data-model.md`.
Do: in §3.3's "For version 1" list, add "a locator document holds at most 256 region points in total, counted across its region selectors (`locator.MaxRegionPoints`), and is at most 64 KiB (`locator.MaxDocumentBytes`); the app writes one region per locator, and its freeform tool stops at the cap". In §3.8, extend invariant 9 with "and the document holds at most 256 region points". The plan Status, its Landed note, and the map row are set after the draft PR exists, in `implement-pr` step 8's commit "Record S10-61 as landed in pull request N.": the plan Status becomes `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)` and the S10-61 row in `docs/deployment-plan/spike-10-followups/deployment-plan.md` becomes `Landed #N`.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-61-region-caps.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. Grouping: STEP-1 and STEP-2 are one commit (Go refuses an oversized locator); STEP-3 and STEP-4 are one commit (the app refuses and stops drawing an oversized region); STEP-5 is one commit; STEP-6 is one commit; the landed record is `implement-pr` step 8's commit.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/locator/validate_test.go` `TestValidate` (new rows) | `ringRegion(256)` in a document → nil; `ringRegion(257)` → `ErrInvalid` | AC-1.1, AC-1.2, D-1 |
| T-2 | Go | `core/locator/validate_test.go` `TestValidate` (new row) | a page selector plus `ringRegion(129)` and `ringRegion(128)` → `ErrInvalid` | AC-1.1, D-3 |
| T-3 | Go | `core/locator/validate_test.go` `TestValidate` (new rows) | a valid artifact-only document padded with spaces inside the JSON to 65,537 bytes → `ErrInvalid`; the same at 65,536 bytes → nil | AC-1.3, D-2 |
| T-4 | Go | `core/locator/validate_test.go` `TestValidateDocCapBeforeCrossing` *(new)* | D-10 document of 200,000 `normalized` ring points → `ErrInvalid`, elapsed under 100 ms; fails before the cap | AC-1.4, D-10 |
| T-5 | Go | `core/locator/depiction_test.go` `TestValidateDepictionRegion` (new rows) | page 1 + `ringRegion(257)` region document with a page 1 Citation locator → `ErrInvalid`; a valid page 1 region with a Citation locator of page 1 + `ringRegion(257)` → `ErrInvalid`; a valid page 1 region with a Citation locator of page 1 + `ringRegion(129)` + `ringRegion(128)` → `ErrInvalid` | AC-1.5, AC-1.6, AC-1.7, D-6 |
| T-7 | Swift | `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` `polygonOverPointCapIsInvalid` *(new)* | `RegionFixtures.ring(count: 256)` → true; `ring(count: 257)` → false | AC-2.1, AC-2.2, D-7 |
| T-8 | Swift | `macos/ProvenenciaTests/CitationLocatorDraftTests.swift` `decodeDropsRegionOverPointCap` *(new)* | `RegionFixtures.locatorJSON(ringCount: 257)` → `CitationLocatorDraft.decode(...).region == nil`; `ringCount: 256` → region present | AC-2.5 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "region point cap" *(new)* | the Swift literal equals `locator.MaxRegionPoints` | AC-2.4, D-9 |
| T-10 | Swift | `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` `freeformStopsAtPointCap` *(new)* | `appendFreeformVertex(points: ring(count: 255), p)` → 256 points; `appendFreeformVertex(points: ring(count: 256), p)` → nil | AC-2.3, D-8 |

T-6 is not used: D-6 drops it.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestValidate`, `TestValidateDepictionRegion`, `TestFromJoin`, `TestSwiftVocabularyMatchesGo`, `ArtifactRegionGeometryTests`, `CitationLocatorDraftTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no UI change beyond D-8)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in [#N](…) (date); a Landed note lists every difference from the plan (implement-pr step 8)
- [ ] docs/deployment-plan/spike-10-followups/deployment-plan.md: S10-61 row status Landed #N (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: no brief)
- [ ] Domain doc updated: docs/interpretation-layer-data-model.md §3.3 and §3.8 (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project, cite a photo, and draw a circle region on the Citation. Save. The region chip shows and the Citation saves (the 32-point circle is under the cap).
2. (AC-2.3) Arm the freeform tool, click four points, and close the polygon on the first point. It commits, as before. The stop at 256 points is proved by T-10; the beep is the one `NSSound.beep()` line in the diff.
3. (AC-1.5, AC-1.6, AC-2.5) In a copy of that project, with the app closed, add a likeness on that Citation, then run `python3 -c 'import math,json; print(json.dumps({"version":1,"selectors":[{"type":"artifact"},{"type":"region","unit":"normalized","points":[{"x":round(0.5+0.4*math.cos(2*math.pi*i/300),6),"y":round(0.5+0.4*math.sin(2*math.pi*i/300),6)} for i in range(300)]}]}))' > /tmp/ring.json` and `sqlite3 provenencia.sqlite "UPDATE citations SET locator_json = readfile('/tmp/ring.json') WHERE ref = '<the Citation ref>'"` (a photo Citation, so no page selector). Open the copy: the composer shows the Citation with no region chip, and opening the likeness for edit shows the hidden-region caption (`L10n.CitationComposer.depictionCaptionHidden`).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A dogfood Citation already stores a freeform region over 256 points | It reads as no region after this PR. Before landing, the user checks each dogfood project with `sqlite3 <project>/provenencia.sqlite "SELECT ref, length(locator_json) FROM citations ORDER BY 2 DESC LIMIT 5"`; a locator over about 12 KB holds more than 256 points and goes to the user before merge |
| T-4's time bound flakes | The bound is 100 ms against a capped cost of microseconds and an uncapped cost of about 90 s |

**Stop and report if:**

- `core/locator/validate.go` `validateRegion` or `selfIntersects` no longer matches §4 (a later PR already capped or replaced it).
- `validateDoc` is no longer the only caller of `validateRegion` (D-3 rests on it).
- The STEP-2 fixture search finds a test fixture or seed that builds a region over 256 points or a locator over 64 KiB.
- `Readable` no longer empties a region that fails `ValidateDepictionRegion` (D-6 rests on it).
- `isValidPolygon` is no longer the one Swift gate that `decodeRegion`, `ArtifactRegionDraft.isValid`, and `hasRegion` share.

## 12. Open questions

None.

## 13. PR

**Title:** S10-61 — Cap region size
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-61-region-caps.md`.
