# Ship log — S12-02 Copy Source form

**Plan:** [`S12-02-copy-source-form.md`](S12-02-copy-source-form.md) · **Branch:** `s12-02-copy-source-form` · **PR:** #485
**State:** fixing 1

Worktree: `/Users/jakerobins/Repos/provenencia-s12-02`, branched from `origin/main` @ `97eea0b3`.

## Round 0 — implement
Head `56be66b5`. Gate PASS (orchestrator run: 1213 Swift tests, 118 suites). Report deviations: `sources.copy.artifactCount` uses numbered placeholders (`%1$lld of %2$lld`) because the catalog check rejects unnumbered ones with two arguments; T-9 re-queries the Sources list after the copy because `.createdSource` marks a hidden list stale rather than reloading it (recorded in the plan's Landed note); brief links fixed after the move to `design/archive/`; §10 manual check not run; AC-3.4 `.failed` body has no automated test (`FakeStore` has no error hook, D-18).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`, `Localizable.xcstrings`), SEC+PRF (performance: `macos/App/Features/**` models — `SourceCopyModel`, `SourcesModel`, `SourcePageModel`; security paths not touched, rides along in the combined reviewer). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 (+ SEC outside-lens note) | UX | major (raised from minor) | Checklist count and All Artifacts state go wrong after an unchecked Artifact is deleted elsewhere | Fix now → F-2 | verified: `SourceArtifactChecklist.swift` header and `allRow` use raw `excluded.count`; `excluded` is never intersected with the live ids. Raised: visible wrong state ("0 of 1" beside a checked row) on the AC-3.2 path in code this PR adds. The fix keeps D-16's meaning (checked when no live Artifact is excluded) |
| MNT-1 | MNT | major | `SourcesModel.typeComboOptions` and `addForm(typeOptions:)` left over | Fix now → F-1 | verified: `SourcesListView.swift:108,385,390`, `SourcesModel.swift:330`; only the load-error guard and one test read it. §6 Deletes already lists the mapping bodies |
| MNT-2 | MNT | major | `PVSkeletonBar` belongs in its own `Components/SkeletonBar/` folder plus a README row | To user → user 2026-10-18: move it → F-4 | contradicts D-23, which names `Components/List/PVSkeletonBar.swift`; `add-ui-component` says one folder per control. User overrides D-23's path |
| UX-4 | UX | minor | Add Source Create disabled with no visible reason when type never touched | To user → user 2026-10-18: show both errors after any edit → F-5 | verified: `typeError` set only from `onChange(of: sourceTypeID)` (`SourceTitleTypeFields.swift:54`) and `create()`; contradicts D-5's display rule |
| UX-2 | UX | minor | Checklist rows: thumbnail and ref are extra VoiceOver stops; unlabeled rows have an empty-label checkbox | Fix now → F-3 (rides with F-2, same file) | verified: `SourceArtifactChecklist.swift` row builds `PVCheckbox(verbatim: "")`; thumbnail not hidden |
| UX-3 | UX | minor | Mixed All Artifacts box speaks as "not selected" | Follow-up | kit `PVCheckbox` accessibility; file not in the fix plan |
| MNT-3 | MNT | minor | Source type id → type lookup copied inline 6+ times | Follow-up | log in `docs/ideas/shared-list-helpers.md` |
| MNT-4 | MNT | minor | `openableRelPath` restates `hasFile` | Follow-up | `GenealogyStore.swift:85` vs `:105`; file not in the fix plan |

Nits (logged): `CON-1 · SourceCopyModel.swift:128 · copy() guard could include canCopy` · `MNT-5 · SourcesModel.swift · SourceDraftValidation → Draft.Validation` · `MNT-6 · SourceCopyDialog.swift · three part checkboxes as one helper` · `MNT-7 · SourceCopyModelTests loadStatePrecedence duplicates other asserts` · `UX-5 · Localizable.xcstrings · straight ' in copySource descriptionHint/loading/loadErrorTitle; §6 says typographic`

SEC+PRF and CON: clean.

### Fix plan 1

**F-1 — Drop the leftover Add Source type options path** (from MNT-1)
Files: `macos/App/Features/Sources/SourcesListView.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`.
Do: remove the `typeOptions:` parameter from `addForm`; the load-error callout guard tests `model.types.isEmpty`. Delete `SourcesModel.typeComboOptions`; the `SourcesModelTests` assertion at ~L275 checks `model.types.count == 2` instead. Leave `SourceIdentitySection.typeComboOptions` (the page header combo uses it).
Check: build; `SourcesModelTests` pass; scripts/check-pr.sh PASS.

**F-2 — Count and All Artifacts state ignore excluded ids that are no longer live** (from UX-1, SEC outside-lens note)
Files: `macos/App/Features/Sources/SourceArtifactChecklist.swift`, `macos/ProvenenciaTests/SourceCopyModelTests.swift` (only if you move the count onto the model).
Failing test first: if you put the count on the model, add `SourceCopyModelTests/selectedArtifactCountIgnoresDeletedExclusions` — uncheck a2, delete a2 through the store, reload the workspace, expect a selected count of 1 of 1. Otherwise no test (view-only); say so in the report.
Do: preferred — add `selectedArtifactCount` on `SourceCopyModel` as `artifactIDsToCopy.count` (beside `artifactIDsToCopy`, which already filters the live list) and pass what the checklist needs; or in the checklist compute `let live = excluded.intersection(artifacts.map(\.id))` once and use it for the header count, the All getter (`live.isEmpty`) and `isIndeterminate` (`!live.isEmpty && live.count < artifacts.count`). Keep `excluded` as the stored binding; do not prune it.
Check: the test (if added) fails before and passes after; scripts/check-pr.sh PASS.

**F-3 — One VoiceOver stop per Artifact row** (from UX-2)
Files: `macos/App/Features/Sources/SourceArtifactChecklist.swift`.
Do: `.accessibilityHidden(true)` on the row's `CachedThumbnail` and on the trailing ref `Text`. For an unlabeled Artifact, use the kit's box-only `PVCheckbox` initializer with a spoken label (`accessibilitySpokenLabel: artifact.displayTitle`, check the exact init in `DesignSystem/Components/Checkbox/PVCheckbox.swift`) instead of `PVCheckbox(verbatim: "")`. For a labelled Artifact, add `.accessibilityValue(Text(verbatim: artifact.ref))` to the checkbox so same-label rows can be told apart. Do not change `PVCheckbox` itself.
Check: build; scripts/check-pr.sh PASS.

**F-4 — `PVSkeletonBar` gets its own component folder** (from MNT-2; user decision 2026-10-18, overrides D-23's path)
Files: `macos/App/DesignSystem/Components/List/PVSkeletonBar.swift` → `macos/App/DesignSystem/Components/SkeletonBar/PVSkeletonBar.swift` (`git mv`), `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`.
Do: move the file; give it its own `SkeletonBar` group in the pbxproj beside the other component groups; add a `SkeletonBar` row to the README "What's built" table after `List`, naming its two callers (`PVListSkeleton`, the Copy Source loading body). Record the moved path in the plan's Landed note (D-23 named `Components/List/`).
Check: build; scripts/check-pr.sh PASS.

**F-5 — Editing either Source field shows both field errors** (from UX-4; user decision 2026-10-18, changes D-5's display rule)
Files: `macos/App/Features/Sources/SourceTitleTypeFields.swift`.
Do: in both `onChange` handlers (title and type), set `titleError = draft.validation.titleError` and `typeError = draft.validation.typeError`, so after the first edit to any field both errors show. Update the type's doc comment to say so. Nothing is shown on open. This applies to Add Source and Copy Source alike. Record the change in the plan's Landed note (D-5 said each field's error follows its own edit).
Check: build; scripts/check-pr.sh PASS. Manual: Add Source, type a title → "Choose a Source type" appears.
