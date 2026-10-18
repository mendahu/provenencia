# Review — S10-61 Cap region size

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user
- None.

## Writer's own calls
- D-1: the cap is 256 region points, counted across a whole locator document (measured 0.28 ms per check at 256; 1,024 rejected at 2.7 ms).
- D-2: a 64 KiB byte cap in `decode`, before unmarshalling.
- D-4: an oversized document reuses `locator.invalid`; no new code or copy.
- D-5: Go counts per document and Swift per polygon, stated on both constants and in §3.3; a hand-made Citation locator with several regions is Out.
- D-6: no read-path code change; `Readable` already hides a region that fails the check.
- D-8: past the cap a freeform click beeps and adds nothing; clicking the first point still closes the polygon.
- Version bump none, matching the Spike 10 and 11 convention.
- Risk row: the user checks dogfood projects for locators over about 12 KB before merge.

## Round 1

### M-1 · major · Put the freeform cap in `ArtifactRegionGeometry`, next to the existing minimum-vertex rule
The plan put the check inline in the private `handleFreeformClick`, proved only by a manual check that edits a production constant. `deleteFreeformVertex` already owns the minimum and is unit-tested.
**Answer:** Fixed — D-8 adds `ArtifactRegionGeometry.appendFreeformVertex(points:_:) -> [CGPoint]?` beside `deleteFreeformVertex`; the overlay appends through it and beeps on nil; T-10 covers 255 → 256 → nil; manual check 2 no longer edits the constant; STEP-4 and Codebase fit updated.

### M-2 · major · One point-cap check, not two: the `validateRegion` check can never run
`validateDoc` is `validateRegion`'s only caller, and every public path goes through it, so the second check is dead. T-4's 200,000-point input could not reach `validateRegion` in production.
**Answer:** Fixed — D-3 keeps only the document sum in `validateDoc`; `validateRegion` is unchanged; T-4 (now `TestValidateDocCapBeforeCrossing`, D-10) calls `validateDoc` on an in-memory document; a Stop condition guards the "only caller" assumption.

### M-3 · major · Make the Swift mirror test a row in `TestSwiftVocabularyMatchesGo`
A one-off `TestSwiftRegionCapMatchesGo` would be a second shape beside the existing table; `go-tests.mdc` asks for one test per family; the file header and `macos-client.mdc` name only wire strings; `IngestMediaPolicy.maxBytes` was missed as the existing numeric mirror.
**Answer:** Fixed — D-9: a `swiftRegionGeometry` constant and one table row; STEP-5 widens the header comment and the `macos-client.mdc` line; Codebase fit names `IngestMediaPolicy.maxBytes`; its drift row is an Out line and a Next add.

### M-4 · minor · Ring-builder helpers missing from Codebase fit, and copied across packages
**Answer:** Fixed — T-6 and old STEP-3 are dropped (verified: `Readable` empties `RegionJSON` on any `ValidateDepictionRegion` error, `core/database/depictionvalues/depictionvalues.go` `Readable`; the existing `TestFromJoin` "off-page region" row proves that branch; T-5 proves both oversized cases fail the call). D-11 names one Go helper pair (`ringPoints`, `ringRegion`) and one Swift fixture (`RegionFixtures`, a new file beside `ArtifactFileFixtures`), each in Codebase fit.

### M-5 · minor · One name, two rules: per document in Go, per polygon in Swift
**Answer:** Fixed — D-5 states the semantics on both constants and in the §3.3 doc line; the divergence for a hand-made multi-region Citation locator is an Out line. A rename to `MaxDocumentRegionPoints` is declined in D-5's Rejected column: the comment carries the meaning.

### X-1 · major · AC-1.7 and the first T-5 row cannot test the cap
`regionDocPages` already refuses a region document with two region selectors.
**Answer:** Fixed — AC-1.7 now puts two regions (129 + 128 points) in the Citation locator, which may hold several; the old T-5 row is gone; T-2 still covers the sum through `Validate`.

### X-2 · major · T-4 does not say to set `Unit`
**Answer:** Fixed — D-10 builds the selector with `Unit: "normalized"`; STEP-2 runs T-4 once before the cap and expects it to fail.

### X-3 · major · `ringRegionJSON` has no signature
**Answer:** Fixed — D-11 gives `ringPoints(n int) []point` and `ringRegion(n int) string`, the circle, the coordinate format, and that documents are assembled with `fmt.Sprintf` in each test.

### X-4 · minor · T-6 needs a per-row locator the table does not have
**Answer:** Fixed — T-6 is dropped (M-4), so `TestFromJoin` is unchanged.

### X-5 · minor · The Swift test ring helper is not specified
**Answer:** Fixed — D-11: `RegionFixtures.ring(count:)` and `RegionFixtures.locatorJSON(ringCount:)` in `macos/ProvenenciaTests/RegionFixtures.swift`, wired per `add-swift-test`.

### X-6 · minor · The Swift-vocabulary test leaves small choices open
**Answer:** Fixed — D-9 names the constant, the regex, `swiftEnumBody` over `ArtifactRegionGeometry`, `strconv.Itoa` on the Go side (the table compares strings), and a missing match fails through the existing comparison.

### X-7 · major · STEP-7 and the DoD need the PR number
**Answer:** Fixed — the docs step is limited to the domain doc; the Status, Landed note, and map row are set in `implement-pr` step 8's commit ("Record S10-61 as landed in pull request N."), with the exact strings: plan `Landed in [#N](…) (date)` (the template and the Spike 11 plans, `docs/deployment-plan/spike-11/prs/S11-06-migrate-reconciler.md` line 3) and map row `Landed #N` (the Spike 11 map).

### X-8 · minor · Checks that cannot run off a Mac; STEP-5 build flags
**Answer:** Fixed in part — STEP-4's check is now an `xcodebuild test` with `CODE_SIGNING_ALLOWED=NO`, as `scripts/check-pr.sh` `xcode_test` uses; §7 says the Xcode steps run on the user's Mac. Declined for the host requirement itself: `ship-pr` already states the implementer is on the user's Mac and `check-pr.sh` needs `xcodebuild` (`.agents/skills/ship-pr/SKILL.md` line 50). The dogfood-project risk is now the user's check before merge, with the command; manual check 3 gives the exact commands.

### X-9 · minor · Could not confirm main matches the Planned commit
**Answer:** Declined — `implement-pr` Round 0 step 2 runs `git fetch origin main` and branches from `origin/main` before the comparison; the sandbox's local `main` is a stale ref. `origin/main` on GitHub is `aa60b63`.

### X-10 · minor · AC-2.3 has no automated test
**Answer:** Fixed — by M-1 (T-10).

### X-11 · minor · Fixture search is too narrow
**Answer:** Fixed — STEP-2's search covers `api`, `core`, and `macos/ProvenenciaTests`, non-`_test` helpers included; the Stop condition names it.

### X-12 · minor · "Stored as written" is not always true
**Answer:** Fixed — the Out line "Repairing stored oversized locators" now says that re-saving such a Citation in the composer drops the region, as today for any invalid region, and that this is not a bug in the PR.

### X-13 · minor · Manual check 3 leaves its method open
**Answer:** Fixed — manual check 3 gives the `python3` ring generator and the `sqlite3 readfile` update, on a closed copy of the project; "without a stall" is dropped (a 300-point check is about 0.4 ms).

### X-14 · nit · Plan status and commit grouping
**Answer:** Fixed for grouping — §7 maps steps to commits. Declined for status: a plan stays `Draft` until review-plan passes (`plan-pr` step 7–8); the writer flips it after this log.
