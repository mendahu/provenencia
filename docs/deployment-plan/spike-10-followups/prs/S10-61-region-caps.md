# S10-61 — Cap region size

**Status:** Draft
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
**Review:** — (not reviewed yet)

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
| AC-1.4 | `validateRegion` on a selector of 200,000 points | returns `ErrInvalid` in under 100 ms (the uncapped check takes about 90 s) |
| AC-1.5 | `depictionvalues.FromJoin` with a stored region of 257 points on a valid Citation locator | the value comes back with `RegionJSON` empty and its image kept |
| AC-1.6 | `depictionvalues.FromJoin` with a valid region on a Citation locator whose region has 257 points | the value comes back with `RegionJSON` empty and its image kept |
| AC-1.7 | `locator.ValidateDepictionRegion` with two region selectors of 200 points each in one document | returns `ErrInvalid` (the cap counts points across the document, and a depiction region already allows one region selector) |

**US-2.** As a researcher drawing a freeform region, I cannot place more points than Go accepts, so a region I draw always saves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `ArtifactRegionGeometry.isValidPolygon` on 257 points | false, before the crossing test runs |
| AC-2.2 | `ArtifactRegionGeometry.isValidPolygon` on a convex ring of 256 points | true |
| AC-2.3 | Freeform tool armed, 256 points placed, the researcher clicks away from the first point | no point is added and the system beep plays; clicking the first point still closes and commits the polygon |
| AC-2.4 | The Swift cap constant and the Go cap constant | equal; a Go test reading the Swift source fails when they differ |
| AC-2.5 | A Citation whose stored locator region has 257 points is opened in the composer | the region chip is absent (the draft decodes with no region), as for any invalid stored region today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `locator.MaxRegionPoints` = 256 and `locator.MaxDocumentBytes` = 65,536 *(new)*, checked in `decode` and before the area and crossing tests | STEP-1, STEP-2, AC-1.1–AC-1.4, AC-1.7 |
| Read path: no code change; tests prove `Readable` empties an oversized stored region or a region on an oversized Citation locator | STEP-3, AC-1.5, AC-1.6 |
| Swift `ArtifactRegionGeometry.maxRegionPoints` = 256 *(new)*, checked first in `isValidPolygon` | STEP-4, AC-2.1, AC-2.2, AC-2.5 |
| Freeform tool stops adding points at the cap | STEP-5, AC-2.3 |
| Go test that the Swift constant equals the Go one | STEP-6, AC-2.4 |
| Domain doc: the cap in §3.3 and §3.8 | STEP-7 |

| Out | Home |
| --- | --- |
| Caching `DepictionValueDraft.hasRegion` instead of re-checking the polygon on every read (audit PRF note) | not planned: with the cap, one check costs at most about 0.3 ms |
| Swift-side byte cap on `CitationLocatorDraft.decode` | not planned: Go returns only locators it stored, decoding is linear, and the quadratic part is capped by `maxRegionPoints` |
| Symlinked objects and `ThumbnailCache` decode limits (audit SEC-2) | [S10-62](S10-62-object-symlinks.md) and a later item from the audit action list |
| Decode limit in bytes in `core/raster` (audit SEC-3) | not planned in this folder; audit action list (Low) |
| A dedicated error code or copy for an oversized region | not planned: the app's tools cannot produce one, so only a hand-made catalog meets it |
| Repairing stored oversized locators | not planned: they stay stored as written and read as invalid |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Envelope check | `core/locator/validate.go` `decode` | Trims, unmarshals with no size limit, checks version 1 and at least one selector |
| Region check | `core/locator/validate.go` `validateRegion` | Unit, at least 3 points, no repeated closing point, points in 0–1, area, then `selfIntersects`. No point limit |
| Crossing test | `core/locator/validate.go` `selfIntersects` | Every edge against every non-adjacent edge: cost grows with the square of the point count. Measured at `aa60b63`: 256 points 0.28 ms, 1,024 points 2.7 ms, 2,048 points 10 ms, 20,000 points 0.93 s |
| Depiction check | `core/locator/depiction.go` `ValidateDepictionRegion` | Decodes and validates the region document, then the Citation locator (`validateDoc(citation)`), then compares pages. Called on write by `core/database/observations/depiction.go` `resolveDepiction` |
| Citation write | `core/database/citations/citations.go` `normalizeCreateInput`, `Update` | `locator.Validate(in.LocatorJSON)` |
| Read path | `core/database/depictionvalues/depictionvalues.go` `Readable`, `FromJoin` | Empties `RegionJSON` when `ValidateDepictionRegion` fails. Callers: `core/database/observations/observations.go` (Observation list), `core/database/sourcefill/load.go` (Source load), `core/database/autoreconciler/autoreconciler.go` (cache rebuild, run from `EnsureCatalog` at open when `cache_version` differs) |
| Swift mirror | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `isValidPolygon`, `selfIntersects` | Same rules as Go, same quadratic crossing test, no point limit. `circleRingCount` = 32, `freeformMinimumVertices` = 3 |
| Swift callers | `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift` `ArtifactRegionDraft.isValid`; `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` `decodeRegion`; `macos/App/Features/CitationComposer/DepictionValueDraft.swift` `hasRegion` | `decodeRegion` drops a region that fails `isValidPolygon`; `hasRegion` re-runs it on each read |
| Freeform tool | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` `handleFreeformClick`, `mouseDragged` | One point per click; `mouseDragged` adds nothing for freeform. Closing on the first point is checked before appending. No limit |
| Swift constant sync | `api/ffi/swift_vocabulary_test.go` | Reads Swift source files to keep wire strings equal to Go's |
| Tests | `core/locator/validate_test.go` `TestValidate`; `core/locator/depiction_test.go` `TestValidateDepictionRegion`; `core/database/depictionvalues/depictionvalues_test.go` `TestFromJoin` ("off-page region"); `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` (`circleRingHasThirtyTwoPoints`, `bowtieIsNotAValidPolygon`) | No size case anywhere |

**Precedent.** `ac70707` ("Hide a stored depiction region that no longer validates when Observations are listed.") made the read path empty a region that fails the check; this PR relies on it and adds only the cap. Copy its `TestFromJoin` row shape. `api/ffi/swift_vocabulary_test.go` is the precedent for a Go test that reads a Swift file to keep a mirrored value equal.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `MaxRegionPoints = 256` *(new, exported)* in `core/locator`, counted across **all** region selectors of one document | The app's largest shape is the 32-point circle; 256 freeform clicks is far past any face or record outline. 256 points cost 0.28 ms per check, so a read that checks a region and its Citation locator stays under 1 ms. Counting per document stops many capped selectors in one document from multiplying the cost | 1,024 (2.7 ms per check, so 1,000 crafted depictions cost about 5 s at open); a per-selector cap (a document of many capped selectors costs that many times more) |
| D-2 | `MaxDocumentBytes = 65536` *(new, exported)*: `decode` refuses a trimmed document longer than this before `json.Unmarshal` | 256 points serialize to about 12 KB at 17 significant digits; 64 KiB leaves room for page and unknown selectors and keeps unmarshalling bounded | No byte cap (unmarshal cost grows with whatever is stored); 16 KiB (too close to 256 full-precision points plus a `text_quote` a future client preserves) |
| D-3 | The point count is checked in `validateDoc` before any selector's area or crossing test, by summing `len(sel.Points)` over region selectors; `validateRegion` also refuses a selector over the cap as its first check after the unit | Summing first means no quadratic pass starts on an oversized document; the check inside `validateRegion` keeps the function safe for any caller | Checking only in `validateRegion` (a document of two 200-point regions would pass) |
| D-4 | An oversized document returns the existing `locator.ErrInvalid` (`locator.invalid`) | The app's tools cannot produce one, so the only source is a hand-made catalog; the existing copy covers it | A new code and string (copy nobody using the app reaches) |
| D-5 | No change to `Readable`, `FromJoin`, or their callers | They already empty a region that fails `ValidateDepictionRegion`, and the cap makes that call cheap. A stored region over the cap reads as hidden; a Citation locator over the cap hides every depiction region on that Citation, because the Citation locator itself fails | A second, cheaper read-only check (two rules for one document) |
| D-6 | Swift: `ArtifactRegionGeometry.maxRegionPoints = 256` *(new)*; `isValidPolygon` returns false when `points.count > maxRegionPoints`, as its first check after the count of 3 | One mirror of D-1 beside `circleRingCount`; `decodeRegion`, `ArtifactRegionDraft.isValid`, and `hasRegion` all go through `isValidPolygon`, so each gains the cap | A cap in each caller |
| D-7 | Freeform: `handleFreeformClick` keeps its close-on-first-point branch first; after it, when `inProgressPoints.count >= ArtifactRegionGeometry.maxRegionPoints`, it calls `NSSound.beep()` and returns without appending | The researcher can still close the polygon; the beep is the AppKit signal for a refused action and needs no copy | Committing automatically at the cap (closes a shape the researcher did not finish); a toast or callout (new chrome and copy for a case nobody reaches by hand) |
| D-8 | `TestSwiftRegionCapMatchesGo` *(new)* in `api/ffi/swift_vocabulary_test.go` reads `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift`, matches `static let maxRegionPoints = (\d+)`, and compares it with `locator.MaxRegionPoints` | The file already holds the "Swift mirrors Go" tests and its helpers; a drift fails in Go CI | A comment only (drifts silently); a new test file for one test |
| D-9 | AC-1.4 is a test inside package `locator` that times one `validateRegion` call on 200,000 points against a 100 ms bound | The uncapped check takes about 90 s, so the bound cannot flake on a slow runner and proves the cap comes before the quadratic pass | A benchmark (does not fail); counting calls to `segmentsIntersect` (needs a test hook in production code) |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `locator.MaxRegionPoints`, `locator.MaxDocumentBytes` *(new)* | const | `core/raster` `MaxEdgeLimit`; `core/ingest/stage.go` `stageMaxEdge` | Those cap images, not locator documents; the cap belongs in the package that owns locator validation (searched `rg "Max[A-Z]\w* *=" core`) |
| `ArtifactRegionGeometry.maxRegionPoints` *(new)* | const | `ArtifactRegionGeometry.circleRingCount`, `freeformMinimumVertices` | It sits beside them; there is no existing point limit in Swift (searched `rg "maxRegion\|maxPoints" macos`) |
| `TestSwiftRegionCapMatchesGo` *(new)* | test | `TestSwiftVocabularyMatchesGo` in `api/ffi/swift_vocabulary_test.go` | Extends the same file and pattern; vocabulary test compares strings, this one an integer |

**Builds on.** `ac70707`'s read rule (`Readable` hides a region that fails the check); the Swift-mirror test file `api/ffi/swift_vocabulary_test.go`.
**Deletes.** None — nothing becomes dead.
**Next adds.** The idea file's polygon crop and read-only region layer ([`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md)) draw many saved regions at once; the cap bounds each one.

### Schema

None — no catalog change. Stored locators stay as written.

### Go core

| Package | Change |
| --- | --- |
| `core/locator` | `const MaxRegionPoints = 256` and `const MaxDocumentBytes = 65536` *(new)*, doc comments citing D-1 and D-2. `decode`: after `strings.TrimSpace`, `len(locatorJSON) > MaxDocumentBytes` → `ErrInvalid`. `validateDoc`: before the selector loop, sum `len(sel.Points)` over selectors whose trimmed type is `region`; over `MaxRegionPoints` → `ErrInvalid`. `validateRegion`: after the unit check, `len(pts) > MaxRegionPoints` → `ErrInvalid` |

Errors: `locator.invalid` → a document over either cap, on write (Citation create and update, depiction save) and on read (the region is hidden by `Readable`).

### FFI

None — no message or handler changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Point cap | `ArtifactRegionGeometry.maxRegionPoints` *(new)* | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` | `static let maxRegionPoints = 256`, one line beside `circleRingCount`, with a comment "Matches `locator.MaxRegionPoints` in Go." |
| Polygon check | `ArtifactRegionGeometry.isValidPolygon` | same file | `guard points.count >= 3, points.count <= maxRegionPoints else { return false }` before the closing-point check |
| Freeform stop | `ArtifactRegionOverlay.handleFreeformClick` | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` | D-7 |

States: a stored region over the cap reads as no region (composer: no region chip; likeness line: "region hidden" caption, as today for an off-page region). Accessibility: no change.

### Strings

None — no new copy.

## 7. Steps

**STEP-1 — Byte cap in `decode`**
Files: `core/locator/validate.go`, `core/locator/validate_test.go`.
Do: add `MaxDocumentBytes` and the check in `decode` (D-2). Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/locator/...` passes.

**STEP-2 — Point cap in `validateDoc` and `validateRegion`**
Files: `core/locator/validate.go`, `core/locator/validate_test.go`, `core/locator/depiction_test.go`.
Do: add `MaxRegionPoints`, the document sum in `validateDoc`, and the selector check in `validateRegion` (D-1, D-3). Add T-1, T-2, T-4, T-5. Build the test points with a small helper inside the test file (a convex ring of n points around (0.5, 0.5), radius 0.4), named `ringRegionJSON` *(new)* in `validate_test.go` and reused from `depiction_test.go` (same package).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/locator/...` passes; T-4 runs in under 1 s.

**STEP-3 — Read-path tests**
Files: `core/database/depictionvalues/depictionvalues_test.go`.
Do: add the two `TestFromJoin` rows of T-6 (D-5). Build the 257-point JSON inline in the test file with a local helper; do not import test code from `core/locator`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/depictionvalues/... ./core/database/observations/... ./core/database/sourcefill/... ./core/database/autoreconciler/...` passes.

**STEP-4 — Swift cap in `isValidPolygon`**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift`, `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`, `macos/ProvenenciaTests/CitationLocatorDraftTests.swift`. Skill: `add-swift-test`.
Do: add `maxRegionPoints` and the guard (D-6). Add T-7 and T-8.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactRegionGeometryTests -only-testing:ProvenenciaTests/CitationLocatorDraftTests` passes.

**STEP-5 — Freeform stops at the cap**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift`.
Do: D-7, in `handleFreeformClick` after the close branch.
Check: the app builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'`); manual check step 2.

**STEP-6 — Swift and Go caps stay equal**
Files: `api/ffi/swift_vocabulary_test.go`.
Do: add `TestSwiftRegionCapMatchesGo` (D-8, T-9). Read the Swift file with `os.ReadFile` at `../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift`, as the file's other constants do.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftRegionCapMatchesGo` passes; changing the Swift value to 255 locally makes it fail (revert after).

**STEP-7 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/deployment-plan/spike-10-followups/deployment-plan.md` (map row), this plan (status).
Do: in §3.3's "For version 1" list, add "a locator document holds at most 256 region points in total (`locator.MaxRegionPoints`) and is at most 64 KiB (`locator.MaxDocumentBytes`); the app's freeform tool stops at the cap". In §3.8, extend invariant 9 with "and the document holds at most 256 region points". Then the Definition of done lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-61-region-caps.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/locator/validate_test.go` `TestValidate` (new rows) | one region of 256 ring points → nil; one region of 257 → `ErrInvalid` | AC-1.1, AC-1.2, D-1 |
| T-2 | Go | `core/locator/validate_test.go` `TestValidate` (new row) | a page selector plus two region selectors of 129 and 128 points → `ErrInvalid` | AC-1.1, D-3 |
| T-3 | Go | `core/locator/validate_test.go` `TestValidate` (new rows) | a valid artifact-only document padded with spaces inside the JSON to 65,537 bytes → `ErrInvalid`; the same at 65,536 bytes → nil | AC-1.3, D-2 |
| T-4 | Go | `core/locator/validate_test.go` `TestValidateRegionCapBeforeCrossing` *(new)* | `validateRegion` on a `selectorRaw` of 200,000 ring points → `ErrInvalid`, elapsed under 100 ms | AC-1.4, D-9 |
| T-5 | Go | `core/locator/depiction_test.go` `TestValidateDepictionRegion` (new rows) | region document with two region selectors of 200 points → `ErrInvalid`; valid region with a Citation locator of 257 region points → `ErrInvalid` | AC-1.7, AC-1.6 |
| T-6 | Go | `core/database/depictionvalues/depictionvalues_test.go` `TestFromJoin` (new rows "oversized region", "oversized citation locator") | 257-point region, valid Citation locator → `RegionJSON` "" and image kept; valid region, 257-point Citation region → `RegionJSON` "" and image kept | AC-1.5, AC-1.6, D-5 |
| T-7 | Swift | `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` `polygonOverPointCapIsInvalid` *(new)* | convex ring of 256 points → true; 257 points → false | AC-2.1, AC-2.2, D-6 |
| T-8 | Swift | `macos/ProvenenciaTests/CitationLocatorDraftTests.swift` `decodeDropsRegionOverPointCap` *(new)* | locator JSON with a 257-point region → `CitationLocatorDraft.decode(...).region == nil` | AC-2.5 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftRegionCapMatchesGo` *(new)* | the Swift literal equals `locator.MaxRegionPoints` | AC-2.4, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestValidate`, `TestValidateDepictionRegion`, `TestFromJoin`, `ArtifactRegionGeometryTests`, `CitationLocatorDraftTests`, and any test fixture that builds a large region (search `rg -n "points" core --glob '*_test.go'` for fixtures over 256 points before STEP-2).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no UI change beyond D-7)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-10-followups/deployment-plan.md: map row status (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: no brief)
- [ ] Domain doc updated: docs/interpretation-layer-data-model.md §3.3 and §3.8 (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project, cite a photo, and draw a circle region on the Citation. Save. The region chip shows and the Citation saves (the 32-point circle is under the cap).
2. Arm the freeform tool and click points; close the polygon on the first point. It commits. Then arm it again and, using a test build with `maxRegionPoints` set to 5, click six points away from the first: the sixth click beeps and adds nothing; clicking the first point closes the 5-point polygon. Restore the value to 256.
3. In a copy of that project, edit the Citation's `locator_json` with `sqlite3` so its region holds 300 points (a ring), and add a likeness on that Citation first. Open the project: it opens without a stall, the composer shows the Citation with no region chip, and the likeness line says its region is hidden.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A dogfood Citation already stores a freeform region over 256 points | It reads as no region after this PR. STEP-2's fixture search and the manual check catch test data; a real dogfood project is checked with `sqlite3 provenencia.sqlite "SELECT ref, length(locator_json) FROM citations ORDER BY 2 DESC LIMIT 5"` before landing |
| T-4's time bound flakes | The bound is 100 ms against a capped cost of microseconds and an uncapped cost of about 90 s |

**Stop and report if:**

- `core/locator/validate.go` `validateRegion` or `selfIntersects` no longer matches §4 (a later PR already capped or replaced it).
- A test fixture, seed, or dogfood project at the planning base stores a region over 256 points or a locator over 64 KiB.
- `Readable` no longer empties a region that fails `ValidateDepictionRegion` (D-5 rests on it).
- `isValidPolygon` is no longer the one Swift gate that `decodeRegion`, `ArtifactRegionDraft.isValid`, and `hasRegion` share.

## 12. Open questions

None.

## 13. PR

**Title:** S10-61 — Cap region size
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-61-region-caps.md`.
