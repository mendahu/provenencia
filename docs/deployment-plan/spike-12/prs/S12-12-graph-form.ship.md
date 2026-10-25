# Ship log — S12-12 Graph form

**Plan:** [`S12-12-graph-form.md`](S12-12-graph-form.md) · **Branch:** `s12-12-graph-form` · **PR:** #557
**State:** ready

## Round 0 — implement
Head `c2d8928f`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report deviations: pbxproj group id `…24` (`…23` taken on main); test suite is `PVDeleteImpactCopyTests` not `PVDeleteImpactTests`; `PVArtifactChecklistTests.Case` internal for the parameterized test; `SourceGraphFormModel` also exposes `graph` and `keptLinked`; `loadKeptLinks()` sets `.loading` first; STEP-1 and STEP-2 in one commit. Not run: the app (STEP-6 run check, §10 manual check, AC-2.8); PR screenshots missing.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and `Localizable.xcstrings`; SEC+PRF for performance on `macos/App/Features/**` models (`SourceGraphFormModel`, `SourcePageModel`); no security path touched.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, UX-2, MNT (outside lens) | CON | major | A locked `PVCheckbox` is not disabled, against D-3 | Fix now → F-1 | verified: `PVCheckbox.swift:109` is `.disabled(isDisabled)` only; the recipe's locked row passes no `isDisabled` |
| UX-1 | UX | major | The Kept linked lists have no height cap, so the footer can be pushed off the sheet | Fix now → F-2 | verified: `SourceGraphFormDialog.swift:157-175` is a bare `VStack`/`ForEach`; the notice caps the same lists at `PVDeleteImpactLayout.listMaxHeight` (`PVDeleteImpact.swift:197-209`) |
| MNT-1 | MNT | major | The page Source's type icon lookup has a fourth copy | Fix now → F-3 (same-pass cleanup) | verified: copies at `SourceGraphFormModel.swift:72`, `SourcePageArtifactsView.swift:360-364`, and two in `SourcePageIdentityHeader.swift` (~80, ~241). The same body for another Source is also in `SourceCopyModel.swift:72` and `SourcesModel.swift:273` |
| MNT-3 | MNT | minor | `sourceGraphFormDialog` copies `sourceCopyDialog`'s presenter | Fix now → F-4 (same-pass cleanup) | verified: identical `isPresented` binding and save-then-clear `onConfirm` in both files. Supersedes the plan's Out line (review M-5, "S12-13") and the matching sentence in S12-13's map scope entry |
| UX-3 | UX | minor | After an `artifact_in_use` save refusal the form gives no way out | Fix now → F-5 | verified: `save()`'s catch only sets `saveError`; `blocks` are only read on open and Read again. Extends D-7 (re-read after a refusal), it does not run the preflight on Save |
| PRF-1 | PRF | minor | The checklist builds every row and starts every thumbnail load | Fix now → F-6 | verified: `PVArtifactChecklist.swift:76-77` `ScrollView { VStack(spacing: 0) …`; minor in a file this PR adds |
| MNT-2 | MNT | minor | "Locked" spoken from two keys; the kit's locked value is overridden on the only caller | Rejected | D-3 specifies both `checkboxLocked` and the row's `lockedValue(ref:)` ("the row's value replaces the checkbox's own"); the reviewer notes the fix contradicts D-3 |
| CON-2 | CON | minor | The locked-row wiring is view-only and untested | Logged | Note only: covered by the plan's §10 manual check; F-1's change is a view modifier with no unit-test form |
| PRF-2 | PRF | minor | `UnlinkImpact` runs ~2 queries per linked Artifact | Follow-up | Go code this PR does not change; grows with one graph's Artifacts, off the main thread |

Nits (logged): `CON-3 · PVArtifactChecklistTests.swift:7 · #expect per clause` · `CON-4 · SourceGraphFormModelTests T-9 · assert updateEvidenceGraph got artifactIDs: []` · `MNT-4 · SourceGraphsSection.swift:46-53 · derive canAdd and addBlockedReason from one check` · `MNT-5 · SourceGraphFormModel.swift:80-86 · drop the duplicate-block guard` · `UX-4 · L10n graphFormDescriptionHint · "source" lowercase beside "Source"`

### Fix plan 1

**F-1 — Disable a locked checkbox without dimming it** (from CON-1, UX-2)
Files: `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift`.
Failing test first: none (view modifier; CON-2). 
Do: `.disabled(isDisabled || isLocked)`; keep `.opacity(isDisabled ? 0.5 : 1)` driven by `isDisabled` only, as D-3 says. Make sure the locked row still speaks its description and the `lockedValue(ref:)` value.
Check: scripts/check-pr.sh PASS.

**F-2 — Cap the Kept linked lists like the delete notice** (from UX-1)
Files: `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`.
Do: Extract the notice's `ScrollView { VStack(alignment: .leading, spacing: PVSpacing.space7) { … } }.frame(maxHeight: PVDeleteImpactLayout.listMaxHeight)` into one internal recipe view in `PVDeleteImpact.swift` (for example `PVDeleteImpactGroupStack<Content: View>` taking a `@ViewBuilder` content, or taking `[(title, group)]` plus `onNavigate`/`dismiss`). The notice and the form's `keptLinked` both draw their `PVDeleteImpactGroupList`s through it. The form's lock callout stays above the scroll, outside it.
Check: `grep -n "listMaxHeight" -r macos/App` prints only the constant and the one shared view; scripts/check-pr.sh PASS.

**F-3 — One lookup for a Source type's icon** (from MNT-1)
Files: `macos/App/Features/Sources/SourcePageContext.swift`, `SourceGraphFormModel.swift`, `SourcePageArtifactsView.swift`, `SourcePageIdentityHeader.swift`, `SourceCopyModel.swift`, `SourcesModel.swift`; the helper beside `CatalogSourceType` (wherever that type's Swift extensions live; else `SourcePageContext.swift`).
Do: Add one helper over `[CatalogSourceType]` that returns a type id's `iconKey` (for example `extension Array where Element == CatalogSourceType { func iconKey(forTypeID: String?) -> String? }`). Add `var sourceTypeIconKey: String?` on `SourcePageContext` built on it. Point `SourceGraphFormModel.sourceTypeIconKey`, `SourcePageArtifactsView`'s private `sourceTypeIconKey`, and both `SourcePageIdentityHeader` sites at the context property (or the helper where the view only has `model.identity`). Move `SourceCopyModel.sourceTypeIconKey` and `SourcesModel.swift:273` onto the helper. Leave the `.label` lookups alone.
Check: `grep -rn -A2 "types" macos/App/Features | grep -n "\.iconKey"` and `grep -rnE "first \{ \\\$0\.id == .*\}\??\.iconKey" macos/App` print nothing outside the helper; also check multi-line `.first { $0.id == … }?` followed by `.iconKey` with `grep -rn -B2 "^\s*\.iconKey$" macos/App/Features` — expect no hits. scripts/check-pr.sh PASS.

**F-4 — `pvFormDialog(item:)` for model-backed forms** (from MNT-3)
Files: `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, `SourceCopyDialog.swift`, `SourceGraphFormDialog.swift`, `macos/App/DesignSystem/README.md` (FormDialog row), `docs/deployment-plan/spike-12/deployment-plan.md` (S12-13 scope entry), the S12-12 plan's Landed note. Skill: `add-ui-component`.
Do: Add `pvFormDialog(item: Binding<Model?>, …)` to the kit that owns the `isPresented` binding (non-nil while presented; clearing is refused while the caller's `isRunning` is true) and the confirm step (an `async -> Bool` action; on true, set the item to nil). Keep `copy`, `width`, `isRunning`, `isDirty`, `confirmDisabled`, `accessibilityIdentifierPrefix` and the form builder as they are on `pvFormDialog(isPresented:)`, which builds on. Move `sourceCopyDialog` and `sourceGraphFormDialog` onto it; their copy-open navigation after a successful Copy stays in the Copy modifier. Behavior unchanged. In the S12-13 map scope entry, drop the sentence "If its dialogs are the third form … (S12-12 plan review M-5)" and say instead that `pvFormDialog(item:)` exists for its dialogs. In the plan's Landed note, record that this supersedes the Out line (review M-5).
Check: `rg 'get: \{ model != nil \}' macos/App` prints nothing; `SourceCopyModelTests` and `SourceGraphFormModelTests` pass; scripts/check-pr.sh PASS.

**F-5 — Re-read kept links after an `artifact_in_use` refusal** (from UX-3)
Files: `SourceGraphFormModel.swift`, `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`, `FakeStore.swift` if it needs a way to change the unlink blocks between calls.
Failing test first: `SourceGraphFormModelTests/editRefusedInUseRereadsKeptLinks` — edit a graph, uncheck `a2`, make the store's `updateEvidenceGraph` throw `evidencegraphs.artifact_in_use` and its next unlink impact return a block for `a2`; after `save()` returns false, `saveError` is set, `draft` is unchanged, `locked` contains `a2`, and `checkedArtifactIDs` contains `a2`.
Do: In `save()`'s catch, in edit mode, when the error's code is `evidencegraphs.artifact_in_use`, `await loadKeptLinks()` after setting `saveError`. Keep `saveError` visible after the re-read.
Check: the test passes; scripts/check-pr.sh PASS.

**F-6 — Build checklist rows lazily** (from PRF-1)
Files: `macos/App/DesignSystem/Recipes/ArtifactChecklist/PVArtifactChecklist.swift`.
Do: The rows' `VStack(spacing: 0)` inside the `ScrollView` becomes `LazyVStack(spacing: 0)`. The All row and count stay outside, as now.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5f0f58b2`. Gate PASS (orchestrator re-ran). F-1 `dd548d73`, F-2 `7ff8ec4a` (`PVDeleteImpactGroupStack`), F-3 `753ddf02` (`[CatalogSourceType].iconKey(forTypeID:)` in `GenealogyStore.swift`, `SourcePageContext.sourceTypeIconKey`), F-4 `bddf68ed` + `5f0f58b2` (`pvFormDialog(item:)`; S12-13 scope entry and Landed note updated), F-5 `2d9cd949` (`editRefusedInUseRereadsKeptLinks` red first), F-6 `67528124`. Deviations: the Copy modifier navigates inside `onConfirm` before the kit clears the item (same main-actor turn). One implementer gate run was flaky in `OmnibarResultsModelTests` / `CatalogSearchModelTests` (debounce timing, untouched code); passed on rerun.

PRF-1 (minor, no re-review): **Resolved** — `PVArtifactChecklist.swift` rows are a `LazyVStack(spacing: 0)`.

## Round 2 — review
Reviewers: CON (Sonnet; CON-1), MNT (MNT-1), UX (UX-1). Skipped: SEC+PRF (only minors, PRF-1 resolved above). Scope `c2d8928f...5f0f58b2`.

CON-1 Resolved (CON), UX-1 and UX-2 Resolved (UX), MNT-1 Resolved with one copy left (MNT). No new blockers from any reviewer.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (leftover) | MNT | minor | `PromoteModel.attachSource` still looks up a type's icon by hand | Fix now → F-7 | verified: `PromoteModel.swift:473` `types?.first(where: { $0.id == source.sourceTypeID })` then `type.iconKey`, the same body as `[CatalogSourceType].iconKey(forTypeID:)`. Not a new round-2 finding: it is the copy F-3 missed (same-pass rule: find copies by body, every package). `SourceTypesModel.swift:277,316` use the whole type, not only its icon, so they are not copies |

MNT-3 (minor, no re-review): **Resolved** — `pvFormDialog(item:)` in `PVFormDialog.swift`; `rg 'get: \{ model != nil \}' macos/App` prints nothing; S12-13 scope entry updated. UX-3 (minor): **Resolved** — `save()` re-reads on `evidencegraphs.artifact_in_use` in edit mode; `editRefusedInUseRereadsKeptLinks` passes (CON and UX round 2 both read it).

### Fix plan 2
**F-7 — Move Promote's type icon lookup onto the helper** (from MNT-1, completing F-3)
Files: `macos/App/Features/Promote/PromoteModel.swift`.
Failing test first: none (refactor; behavior unchanged, existing Promote tests cover `sourceMark`).
Do: In `attachSource()`, replace the `if let type = types?.first(where: …) { sourceMark = PVMarkKey(catalogKey: type.iconKey) }` with `types?.iconKey(forTypeID: source.sourceTypeID)` from `GenealogyStore.swift`, setting `sourceMark` only when it returns a key, as today.
Check: `grep -rn -B4 "\.iconKey" macos/App | grep first` prints only `GenealogyStore.swift` (the helper) and the `SourcesModel.swift:269` `.label` line; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `c5e1a604`. Gate PASS (orchestrator re-ran). F-7 done. The F-7 check grep's `-B4` window also catches the unrelated whole-type lookups in `PromoteModel.swift:471` and `SourceTypesModel.swift:277,316`; no hand-rolled icon lookup is left outside the helper. `origin/main` unchanged since the branch was cut (`bfb852a9`).

## Result
Gate PASS · CI green (macOS tests on `89f22a27`) · open: none · follow-ups: PRF-2 · to user: — · CON miss: none

After ready: the S12-14 scope commit (`89f22a27`, user 2026-10-25: Source page graph-row counts) rides on this PR.
