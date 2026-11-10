# Ship log — S15-03 The Source page shows the date and the Private switch

**Plan:** [`S15-03-source-page-date-and-private-switch.md`](S15-03-source-page-date-and-private-switch.md) · **Branch:** `s15-03-source-page-date-and-private-switch` · **PR:** #766
**State:** fixing 1

Base: `main` @ `9897bf19`.

## Round 0 — implement
Head `b9c5bd75` (PR #766, draft). Gate PASS (orchestrator re-run on `b9c5bd75` against `9897bf19`).
Report deviations: version 0.5.1→0.6.0 (main at 0.5.1); Clear date is `PVButton(variant: .ghost, size: .sm, icon: .trash)` (no `tone:`); `SourcePageContext.setSourceCover` returns false when the Source is gone; the date dialog focuses the year field on open; `SourcePageModel.pageError` is a `PageError?` (D-19), four tests compare `pageError?.message`; `OnboardingModelTests.ThrowingStore` gained the two new methods; `design/README.md` unchanged. The implementer did not read the board; visuals unchecked against it. AC-1.4, 1.8, 3.6, 4.3 and SD-1–SD-8 are code only, with no automated test; manual checks 1–6 not run. `TestLargeSourceBudgets` (promotealign) missed its budget once under a full-suite load (2.8s vs 2.5s), passed on re-run.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX: `macos/App/**`, `Localizable.xcstrings`, `core/apperr`, a new error code. SEC+PRF: `core/database/**` (sources, sourcecopy, sourcemetadata), `macos/App/Features/**` models, FFI handler tests.

SEC+PRF: no findings. Its Outside-my-lens note (long phrase date overflows the header) is the same finding as UX-2.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Deleted-Source branches of the date dialog and `refreshCoverFromStore` untested; `getSourceWorkspaceError` unused | Fix now → F-2 | verified: `rg getSourceWorkspaceError macos/ProvenenciaTests` empty; `setSourceDateError` only set to `datevalues.invalid` |
| MNT-1 | MNT | major | FakeStore holds four copies of one Source write | Fix now → F-3 | verified: `FakeStore.swift` `updateSource`/`setSourceDate`/`setSourcePublic`/`setSourceCover` share steps 2–6; same-pass cleanup (this PR adds two copies) |
| UX-1 | UX | major | VoiceOver never says "off" or "switch" for the Private switch (SD-8, manual check 5) | Fix now → F-1 | verified: `PVSwitch.swift` `SwitchAccessibility` adds only `.isButton`/`.isSelected`, no value or toggle trait; this PR's header is the first Features use |
| UX-2 | UX, SEC+PRF note | minor | Long phrase date pushes the meta row past the header | Fix now → F-4 | `PVFlowLayout` measures `.unspecified`; date `Text` has no cap; file this PR adds |
| UX-3 | UX | minor | Failed Private flip shows its Callout out of sight when scrolled | Fix now → F-5 | `SourcePageView.swift` and `PVPageColumns.swift` are changed here; `scrollResetKey` already exists to reuse |
| UX-4 | UX | minor | Clear date is not announced | Fix now → F-6 | `SourceDateDialog.swift` (new file) caption is plain `Text` |
| MNT-2 | MNT | minor | Overview layout repeats the side max and gap | Fix now → F-4 | `SourcePageView.swift:90,385,397` |
| MNT-3 | MNT | minor | Guidance does not teach `leavesForSectionRoot(when:notice:)` | Fix now → F-7 | `rg leavesForSectionRoot .agents .cursor` finds only `(when:)` |
| MNT-4 | MNT | minor | `showPageError` takes an optional | Fix now → F-8 | no caller passes nil |
| MNT-5 | MNT | minor | "fail, then show" pair repeated ten times | Fix now → F-8 | same-pass cleanup of a pattern this PR adds |
| MNT-6 | MNT | minor | FakeStore workspace read returns a blank Source instead of `sources.not_found` | Fix now → F-3 (structural-debt line if tests rely on the blank read) | `FakeStore.swift:452-453` |
| CON-2 | CON | minor | Private-failure test checks raw English, misses the detail | Fix now → F-2 | `SourcePageModelTests.swift:2217` |

Nits (logged): `MNT-7 / CON-4 · PVPageColumns.swift:236-239 · re-indent band layout args` · `MNT-8 · SourcePageContext.swift:32 · gone private(set)` · `MNT-9 · ConclusionDetailPage.swift:551 · redundant notice condition` · `CON-3 · SourcePageModelTests.swift:2282 · try #require(model.workspace)` · `UX-5 · SourcePageIdentityHeader dateAndPrivate · hide "·" separators from VoiceOver` · `UX-6 · SourceDateDialog readsAs · combine label and value`

### Fix plan 1

**F-1 — The switch announces its state and role** (from UX-1)
Files: `macos/App/DesignSystem/Components/Switch/PVSwitch.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-localized-string`, `add-ui-component`.
Do: in `SwitchAccessibility`, add `.isToggle` to the traits (macOS 14 target) and an `.accessibilityValue` of new `L10n.DesignSystem.switchOn` / `switchOff` ("On" / "Off"), following how `PVCheckbox` supplies its value. Drop `.isSelected` if `.isToggle` plus the value covers it.
Check: `scripts/check-pr.sh` PASS; the plan's manual check 5 line still reads "Private, off".

**F-2 — Pin the deleted-Source branches and the Private failure copy** (from CON-1, CON-2)
Files: `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Failing test first: `deletedSourceClosesTheDateDialog` — with `setSourceDateError = FakeStore.sourceNotFound`, `form.save()` returns `true`, `saveError` is nil, `model.gone` is the deleted notice. Second case: with `getSourceWorkspaceError = FakeStore.sourceNotFound`, the path that calls `refreshCoverFromStore` leaves `gone` set and `pageError` nil (reach it through the Artifacts section if `context` is not reachable; if neither works, remove the hook and say so).
Do: change the Private-failure assertion to `#expect(model.pageError?.message == L10n.Sources.privateFailedMessage(detail: L10n.Errors.message(for: <the error set>)))`.
Check: prove each new test red by flipping the guarded line locally, then restore; check-pr PASS.

**F-3 — One FakeStore Source write** (from MNT-1, MNT-6)
Files: `macos/App/Platform/FakeStore.swift`.
Do: add a private `mutateSource(projectDir:sourceID:_ change: (inout CatalogSource) throws -> Void) throws -> (CatalogSource, CatalogWriteEffects)` doing lookup-or-`Self.sourceNotFound`, the revision bump, write-back and `(enrichCoverFields(...), nextWriteEffects)`. Move `updateSource`, `setSourceDate`, `setSourcePublic`, `setSourceCover` onto it; each keeps only its error hook and field change. Then make `getSourceWorkspace` throw `Self.sourceNotFound` for an unknown id instead of the blank `SRC-XXXXX` Source. If existing tests depend on the blank read, revert that one change and add a line to `docs/ideas/structural-debt.md` instead; report which.
Check: `grep -n "nextAuditRevision += 1" macos/App/Platform/FakeStore.swift` shows no Source-write copy outside `mutateSource` (list the remaining lines and what they belong to); check-pr PASS.

**F-4 — Header date cannot overflow; one source for the split** (from UX-2, MNT-2)
Files: `macos/App/Features/Sources/SourcePageIdentityHeader.swift`, `macos/App/Features/Sources/SourcePageView.swift`.
Do: cap the date `Text` (`.lineLimit(1)`, `.truncationMode(.tail)`, a `maxWidth` so the flow item fits) with `.help(line)`; or make date and Private separate flow items, as long as SD-1's order is kept. In `SourcePageOverviewLayout`, add `static let sideMax = PVSpacing.widthSecondaryMax`, pass it to `fillSplitSideMax`, and place the side column at `x + split.group - split.side`.
Check: `SourcePageOverviewLayoutTests` pass unchanged; check-pr PASS.

**F-5 — A new page error scrolls into view** (from UX-3)
Files: `macos/App/Features/Sources/SourcePageContext.swift`, `macos/App/Features/Sources/SourcePageView.swift`.
Do: count each time a page error is shown (a counter on `SourcePageContext` that `showPageError` increments; clearing does not change it) and fold it into the `scrollResetKey` the page already passes to `PVPageColumns`, so a new error scrolls the primary to the top where the Callout is. No new kit API.
Check: a `SourcePageModelTests` case shows the counter moves on show and not on clear; check-pr PASS.

**F-6 — Clear date is announced** (from UX-4)
Files: `macos/App/Features/Sources/SourceDateDialog.swift`.
Do: when the form starts clearing, announce the "Saving removes this Source's date." caption with the kit's announcement helper that `PVCallout(announces:)` uses (or show it as an announced info `PVCallout`).
Check: check-pr PASS.

**F-7 — Guidance teaches `notice:`** (from MNT-3)
Files: `.agents/skills/add-workspace-place/SKILL.md:162`, `.agents/skills/add-workspace-location/SKILL.md:101`, `.cursor/rules/workspace-navigation.mdc:15`.
Do: each line names `.leavesForSectionRoot(when:notice:)`, "with `notice:` for the toast that says why".
Check: `rg "leavesForSectionRoot" .agents .cursor docs` shows the `notice:` form in all three; check-pr PASS.

**F-8 — One banner writer** (from MNT-4, MNT-5)
Files: `macos/App/Features/Sources/SourcePageContext.swift`, the Sources sections listed in MNT-5.
Do: make `showPageError`'s message `String`. Add `@discardableResult func showFailure(_ error: Error, title: …) -> Bool` on `SourcePageContext` that calls `fail`, shows the banner, and returns false when the page is leaving. Move the ten banner sites onto it; keep `fail` for field-error sites and the date dialog.
Check: `rg -n "context.fail\(" macos/App/Features/Sources` prints only field-error sites, `SourceDateFormModel` and `showFailure` itself (list them); check-pr PASS.
