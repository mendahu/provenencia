# Review — S10-54 Depiction values over the FFI

**Verdict:** Fits with changes — Not ready → all findings answered and applied. No blocker is open.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

Round 2 was not run. The two blockers (X-1, X-2) were test and project-file gaps, fixed in place, and neither reshaped the plan. The plan stays `Blocked — on S10-32`; when S10-32 lands, the refresh (plan-pr "Refresh an existing plan") reruns review-plan.

## To user

- **Error copy (D-9).** "The picture of that region couldn’t be read." for `ingest.unsupported_image`, and "Invalid depiction value." for `depictionvalues.invalid`. **Answered:** keep both (user, 2026-10-17).

## Writer's own calls

- D-1: the update's new crop rides on `UpdateObservationRequest.depiction_input`, not on the `Observation` row.
- D-4: `files.RelPaths` is added, and `sourceCoverThumbnails` moves onto it in this PR.
- D-5: the create and append responses now return listed rows, in draft order.
- D-6: the source graph reads the depiction with a join; `depictionvalues.LookupManyTx` is not added.
- D-8: labels move onto `PropertyValueType` as an exhaustive switch.
- D-10: the unused `observationsToInputs` is deleted.
- `FakeStore` produces no depictions, and the pill keeps the default colour (both go to S10-34).

## Round 1

### M-1 · major · `depictionPaths` makes another copy of "File id → objects path"
**Answer:** Fixed. D-4 adds `files.RelPaths` and moves `sourceCoverThumbnails` onto it; `depictionPaths` only collects ids. Next adds names S10-35 and S10-36 as callers.

### M-2 · major · An update-only input field on the row message and the cached Swift row
**Answer:** Fixed. D-1 puts `depiction_input` on `UpdateObservationRequest`. D-7 adds a `depictionInput` parameter to `updateObservation`, with an extension overload so callers do not change. `CatalogObservation` holds only saved data. T-11 checks the request.

### M-3 · minor · A second, untested Swift list of value types
**Answer:** Fixed for labels: D-8 adds `PropertyValueType.label`, and `SubjectPropertyValueType.label` delegates to it. Typing `researcherCreatable` as `[PropertyValueType]` is logged (Out), not planned.

### M-4 · minor · The source graph could use the LEFT JOIN S10-32 already uses
**Answer:** Fixed. D-6 joins `depiction_values` in `loadSourceObservations`. `LookupManyTx` is dropped (Out → S10-52 if needed), and the old T-1 is replaced by `TestRelPaths`.

### M-5 · minor · The client-path / objects-path pattern is not in the skill
**Answer:** Fixed. D-11 and STEP-7 add the section to `add-ffi-handler`.

### M-6 · nit · The new test helpers have siblings
**Answer:** Fixed. The handler helper is renamed `objectRelPaths`; `writeCrop` sits beside `writePNG` in `thumbnails_test.go`; Codebase fit lists `core/ingest/stage_test.go` `objectFiles` and `stagePNG`.

### X-1 · blocker · "Bytes differ from the crop" is probably false
**Answer:** Fixed. `writeCrop` writes a JPEG; AC-1.1 and T-2 assert a 300 × 200 PNG under `objects/`.

### X-2 · blocker · STEP-6 omits the Xcode project registration
**Answer:** Fixed. STEP-6 lists `project.pbxproj` with ids `AA…A12` and `AA…A13`, checked as unused at the planning base, plus the group and Sources phase.

### X-3 · major · T-7 leaves the request undefined
**Answer:** Fixed. T-7 builds the "creates participation" request with a fourth draft on `personProp` and a symlink made in `reqFn`.

### X-4 · major · D-5 changes response order
**Answer:** Fixed. `listedWritten` returns rows in `written` order. Every Swift caller sends one draft and reads `.first` (`CitationObservationRows.swift`, `CitationConnections.swift`, `CitationComposerModel+Authoring.swift`).

### X-5 · major · D-4 ignores `StorageRelPath`'s error
**Answer:** Fixed. `files.RelPaths` omits underivable rows, as `sourceCoverThumbnails` does today. `depictionPaths` returns `sql.ErrNoRows` for a missing id, and every caller returns that error unchanged.

### X-6 · major · Checks that cannot run
**Answer:** Fixed. STEP-1 forbids hand-editing generated files and stops without the generators (§11). Without a Mac, STEP-5 and STEP-6 report the Xcode check as not run. The S10-32 dependency is the plan's Blocked status.

### X-7 · minor · Size and temptation
**Answer:** Fixed in part. `likenessFixture` names `subjectFixture`'s fourth result. STEP-5 owns the `CatalogProperty.valueType` comment. Out records that `FillListedDisplayText` gives a likeness empty text until S10-34. STEP-4 is declined as a split: its response and read-path work share `depictionPaths` and land together.

### X-8 · minor · T-1 names a row that may not exist
**Answer:** Fixed by M-4: T-1 is now `TestRelPaths` *(new)*.

### Disagreements with the S10-32 plan
**Answer:** The header keeps "in progress", per the user (2026-10-17). §11 now stops if `observations.Update` does not return `Listed.Depiction` filled.
