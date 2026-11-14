# Ship log — S15-10 Adding an Artifact asks for its copy details

**Plan:** [`S15-10-artifact-sheet-add-mode.md`](S15-10-artifact-sheet-add-mode.md) · **Branch:** `s15-10-artifact-sheet-add-mode` · **PR:** #829
**State:** ready

## Round 0 — implement
Head `0ea1c9765`. Gate PASS (orchestrator run, against `36d6ec21d`). Report deviations: (1) `SourcePageModel.apply` also calls `artifacts.workspaceChanged()` when the page's shown Artifacts differ from the new workspace, so an Artifact the sheet made and then deleted elsewhere closes the sheet (AC-4.4, T-11); a file the plan did not name. (2) `IngestFileDropRow` stores values and closures from one `init(file:…)`, no `@Binding`. (3) § 10 manual check not run (no GUI session).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX: the diff touches `macos/App/**` and `Localizable.xcstrings`. SEC+PRF: `macos/App/Features/**` models (performance path), and the PR changes how a chosen or dropped file URL is staged and ingested (`IngestFileChoice`, `IngestFileDropRow`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | A refused file in the drop row is not announced to VoiceOver | Fix now → F-1 | verified: `IngestFileDropRow.swift:89-90` builds `PVCallout(tone: .danger, …)` with `announces` at its `false` default (`PVCallout.swift:70`); the sheet's save-error Callout passes `announces: true` (`ArtifactSheet.swift:75`). The row is code this PR adds; both callers (sheet, Add file dialog) go through it |
| MNT-1 | MNT | major | A third private temp-file writer where `ArtifactFileFixtures` exists | Fix now → F-2 (same-pass cleanup) | verified: `IngestFileChoiceTests.writeTempFile` (new), `IngestMediaPolicyTests.writeTemp`, `SourcePageModelTests.writeTempPNG` share the body; `ArtifactFileFixtures.uniqueProjectDir()`/`write(_:projectDir:relPath:mode:)` exist. Supersedes T-1's "its own private temp-file helper" |
| CON-1 | CON | minor | The second-open guards (D-2 re-check, D-3 `isOpeningAdd`) are untested | Fix now → F-3 | verified: no test names `isOpeningAdd`; `FakeStore.listTypeSuggestions` has no hold |
| MNT-2 | MNT | minor | Artifact close check compares two baselines where `shown` suffices | Fix now → F-4 | verified: `SourcePageModel.swift:170` ORs `previous` and `shown`; the metadata reset above uses `shown` alone; the only Artifact removal (`confirmResourceDelete`) goes through `apply`. T-11 `anAddSheetCloseRule` and `artifactGoneClosesTheSheet` prove it |
| CON-2 | CON | nit→minor | `artifactLabelRequired` comment still says "when Add Artifact is submitted" | Fix now → F-5 | stale comment in files this PR changes (`L10n.swift`, `Localizable.xcstrings`) |
| MNT-5 | MNT | nit→minor | `visibleGraphs` doc still says "(the link checkboxes show)" | Fix now → F-5 | stale comment in `SourceArtifactsSection.swift`, which this PR changes |
| UX-2 | UX | minor | After the create succeeds, a non-`ingest.*` ingest failure says "Couldn't save the artifact … Your changes are kept." | To user (finish report) | verified: `ArtifactSheetModel.save()` catch falls to `saveErrorBody` when `file.refuse` is false, and `saveErrorBody`'s doc says "Nothing was written". But D-8 sets this exact catch order and rejects "a new error copy for add mode". The old dialog showed the raw message as a label error, no better. Rare path (non-`ingest.*` errors from `IngestArtifactFile`: already attached, Artifact gone, write failure) |
| SEC/PRF | SEC+PRF | — | No findings | — | — |

Nits (logged): `UX-3 · SourcePageArtifactsView.swift:16 · no feedback while Add artifact's suggestion read runs` · `MNT-3 · ArtifactSheetModel.swift:239,266,285 · label trimmed in three places` · `MNT-4 · SourceArtifactGraphLinks.swift:12-13 · locked/isDisabled defaults only served the deleted addForm` · `MNT-6 · SourceArtifactsSection.stagePageDrop · repeats openAdd's guard` · `MNT-7 · ArtifactSheetModel.swift:146,221 · create mode excluded twice; isCreate mixed with switch mode` · `CON-3 · ArtifactSheet.swift:176 · no-graph line still shows under locked boxes after a refused file`

### Fix plan 1

**F-1 — The drop row's refused-file Callout announces** (from UX-1)
Files: `macos/App/Features/Sources/IngestFileDropRow.swift`.
Failing test first: none practical (announcements are not observable from the model); the change is one argument.
Do: pass `announces: true` to the row's `PVCallout`, as `ArtifactSheet.swift:75` does.
Check: `grep -n "announces: true" macos/App/Features/Sources/IngestFileDropRow.swift` prints one line; scripts/check-pr.sh PASS.

**F-2 — One temp-file test helper on `ArtifactFileFixtures`** (from MNT-1; same-pass cleanup, supersedes T-1's private-helper line)
Files: `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `IngestFileChoiceTests.swift`, `IngestMediaPolicyTests.swift`, `SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: add `static func writeTemp(_ data: Data, named: String) throws -> URL` to `ArtifactFileFixtures`, built on `uniqueProjectDir()` and `write(_:projectDir:relPath:mode:)` (the file lives in a fresh unique folder, so the name is kept as given). Move `IngestFileChoiceTests.writeTempFile`/`writeTempPNG`, `IngestMediaPolicyTests.writeTemp` and `SourcePageModelTests.writeTempPNG` onto it and delete them. Then look at every other hit of `grep -rn "temporaryDirectory" macos/ProvenenciaTests`: move any that writes bytes to one new temp file and returns or uses its URL (the shape above, by body, whatever its name, e.g. the inline writers in `SourcePageModelTests` around lines 416–653) onto the helper; leave hits that make a directory, a project or an install root. List in the report which hits you moved and which you left, with a reason each.
Check: `grep -rnE "func (writeTemp|writeTempFile|writeTempPNG)\(" macos/ProvenenciaTests` prints only the `ArtifactFileFixtures.swift` line; `grep -n "temporaryDirectory" macos/ProvenenciaTests/{IngestFileChoiceTests,IngestMediaPolicyTests}.swift` prints nothing; every moved test still passes; scripts/check-pr.sh PASS.

**F-3 — Test the second-open guards** (from CON-1)
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Failing test first: `SourcePageModelTests.anEditSheetOpenedDuringTheReadIsKept` — add `FakeStore.listTypeSuggestionsHold = FakeHold()`, shaped like `updateArtifactHold`. With the hold on, start `openAdd()` in a Task and wait until held; expect `!model.acceptsArtifactPageDrop` and a second `openAdd()` to return nil; open an edit sheet on a seeded Artifact; release; expect the first `openAdd()` to return nil and `sheet?.artifactID` to be that Artifact. Confirm it fails red with either guard (`isOpeningAdd`, the post-await `sheet == nil` re-check) removed, then restore.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — Close check reads `shown` alone** (from MNT-2)
Files: `macos/App/Features/Sources/SourcePageModel.swift`.
Do: change the condition to `shown?.artifacts != workspace.artifacts`, keeping the comment (reword it to say why `shown` is the baseline, as the metadata reset does). Update the plan's Landed note 1 to match.
Check: `anAddSheetCloseRule` and `artifactGoneClosesTheSheet` pass; scripts/check-pr.sh PASS.

**F-5 — Two stale comments** (from CON-2, MNT-5)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Sources/SourceArtifactsSection.swift`.
Do: `artifactLabelRequired`'s comment (both files) says the Artifact sheet, either mode, not "when Add Artifact is submitted". `visibleGraphs`' doc says what it feeds now (`graphLines`), not the link checkboxes.
Check: `rg -n "Add Artifact is submitted|link checkboxes show" macos` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `2c18daed2`. Gate PASS (orchestrator run). F-1 to F-5 done. F-2 added `ArtifactFileFixtures.writeTemp(_:named:)` and `writeStubPNG()`; moved the inline byte writers in `SourcePageModelTests` too; the other `temporaryDirectory` hits make folders, install roots or settings paths and stay. Deviations: F-3's second-open check asserts the hold count never reaches 2 within 300 ms (otherwise a removed guard hangs instead of failing); the F-3 test body rides in `1621f4b9d` and its `FakeStore` hold in `f9bf5b311`, so `1621f4b9d` does not build alone.
CON-1 (minor, CON not re-sent): **Resolved** — read `anEditSheetOpenedDuringTheReadIsKept` (`SourcePageModelTests.swift:2303`); implementer ran it red with each guard removed.

## Round 2 — review
Reviewers: MNT (MNT-1 major), UX (UX-1 major). Not re-sent: CON (minor only, read above), SEC+PRF (no findings).
MNT-1: **Resolved** — one writer, `ArtifactFileFixtures.writeTemp(_:named:)` (`:23`) and `writeStubPNG()` (`:30`); no copy left in the touched test files.
UX-1: **Resolved** — `IngestFileDropRow.swift:90` passes `announces: true`; both callers share the row.
MNT-2, CON-2, MNT-5 (minors): **Resolved** — read `SourcePageModel.swift:170` (`shown` alone; Landed note matches) and the corrected `artifactLabelRequired` and `visibleGraphs` comments.

No new findings. Nits (logged): `MNT-8 · S15-10 plan:274 · T-1 still names a private temp-file helper (the Landed note records F-2 superseding it)` · `MNT-9 · FakeStore.swift:194 · listTypeSuggestionsHold has no doc comment, unlike updateArtifactHold` · older private `uniqueProjectDir()` + `write` copies in `ThumbnailCacheTests` and `ArtifactViewerModelTests`, untouched by this PR

## Result
Gate PASS (head `2c18daed2`, `main` unmoved at `36d6ec21d`) · CI pending · open: none · follow-ups: none · to user: UX-2 (D-8's save-error copy after a created Artifact's non-`ingest.*` file failure), § 10 manual check
