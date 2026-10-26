# Ship log — S12-13 Artifact graph checkboxes

**Plan:** [`S12-13-artifact-graph-checkboxes.md`](S12-13-artifact-graph-checkboxes.md) · **Branch:** `s12-13-artifact-graph-checkboxes` · **PR:** #562
**State:** ready

## Round 0 — implement
Head `5b19a989`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh` against `846fa656`). PR #562 (draft). Report: deviations all recorded in the plan's Landed note — `evidenceGraphIndex` internal not private; `SourceArtifactLinkEdit` nested as `SourceArtifactsSection.LinkEdit`; `PVArtifactChecklist.maxRowsHeight` internal; added `SourceArtifactsSection.linkEditGraphs` and `FakeStore.updateArtifactCalls`; `saveFields` returns early while the link read loads; `EvidenceGraphModelTests` seeds graph links. §10 manual check, VoiceOver (LK-10) and the twelve-graph cut-off stop condition not run in the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: no security path, but performance paths `macos/App/Features/**` models and the session list index/cache.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | `cardSaveError` unkeyed; a failed save on card A shows on card B | Fix now → F-1 | verified: `beginEdit` (SourceArtifactsSection.swift:161) has no `savingID` guard; the Edit icon on other cards stays live (SourcePageArtifactsView.swift:233); the catch writes `cardSaveError` before any id check |
| PRF-1 | PRF | major | Card save reloads the whole graph list once per link/unlink | Fix now → F-2 | verified: `session.apply(effects)` inside both loops; each apply revalidates `.evidenceGraphsList`. Apply once after the loop still honours D-8 (effects through `session.apply`, one `readyValue` await); patching the cached value rejected (D-8's M-10 alternative) |
| UX-1 | UX | major | After an in-use refusal the callout says "save again" while Save is disabled with nothing left | To user → F-7 (user 2026-10-25: context copy) | verified: `checkedGraphIDs` counts locked as checked, so after the re-read `linksDirty` is false; the copy is the plan's agreed AC-2.7 / §6 string |
| UX-3 | UX | minor | Save-error copy talks about boxes and partial saves when neither applies | To user → F-7 (with UX-1) | same string; copy call |
| UX-4 | UX | minor | Unlink refusal does not name the graph | To user → F-7 (with UX-1) | same callout; copy call |
| UX-2 | UX | major | Add retry after a refused file leaves graph boxes live but ignores them | Fix now → F-3 | verified: retry branch of `create()` never reads `excludedGraphIDs`; group not disabled. Disabling keeps D-4 (boxes kept, not re-applied) |
| MNT-2 | MNT | major | `SourceArtifactGraphLinks` copies `PVArtifactChecklist`'s frame, locked row and toggle | Fix now → F-4 | same-pass cleanup; supersedes D-12's copy of the frame |
| CON-1, UX-out | CON | minor | Graph deleted during edit still gets a link call | Fix now → F-5 | verified: `toLink` iterates `edit.graphIDs` unfiltered |
| CON-3 (UX outside-lens) | CON | minor | Failed reload keeps the stale list, so `readyValue` is non-nil and locks re-read from old links | Fix now → F-5 | `QueryHandle.applyFailure(clearValue: !hadStaleValue)`; treat handle error as failure |
| MNT-3 | MNT | minor | Graph-links-artifact membership written inline three times | Fix now → F-6 | beside `hasArtifact` this PR added |
| MNT-4 | MNT | minor | Unlink-block test builder and Source-page setup copied in tests | Fix now → F-6 | same-pass cleanup into `EvidenceGraphFixtures.swift` |
| UX-5 | UX | minor | Composer callout points to card linking a one-graph Source lacks | Follow-up | edge state (one graph, no links, Artifacts exist); idea file |

Nits (logged): `CON-2 · SourceArtifactsSection.swift loadLinkLocks · reopened card may receive old task's (equivalent) locks` · `MNT-5 · SourcePageArtifactsView.swift:670 · "needs one" caption repeats SourceGraphFormDialog style` · `MNT-6 · SourcePageArtifactsView.swift:315 · outer .disabled vs PVArtifactChecklist isDisabled`

**To user, answered 2026-10-25 (UX-1, UX-3, UX-4):** "Context copy". The card stays in edit and picks the callout body by what is left to save. This supersedes the AC-2.7 / §6 `artifactCardSaveErrorMessage` wording for the two cases below. Record it in the plan's Landed note.

### Fix plan 1

**F-1 — Tie the card's save error to the card that saved** (from MNT-1)
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Failing test first: `cardSaveErrorStaysOnTheCardThatSaved`: start `saveFields(id: "P3")` with `unlinkEvidenceGraphArtifactError` set, call `beginEdit("P4")` before the save finishes, then expect no save error for P4.
Do: write the error only when `editingArtifactID == id`, or key it by card id as `fieldErrors` does. In both cases the view reads it through `cardSaveError(id:)`, which guards on `editingArtifactID == id`, matching `cardLabelError(id:)`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — One graph-list reload per card save** (from PRF-1)
Files: `SourceArtifactsSection.swift`; `CatalogWriteEffects` (wherever it is declared, `macos/App/Platform/`) if a merge is needed; `SourcePageModelTests.swift`.
Failing test first: `cardSaveReloadsTheGraphListOnce`: a save that links one graph and unlinks another calls `listEvidenceGraphs` once after the writes. Count the calls in FakeStore, adding a counter if none exists.
Do: collect the effects of the link and unlink writes that succeed. Call `context.session.apply` once after the loop, on success and in the catch (before the existing `graphListReloaded()` await), whenever any write succeeded. If `CatalogWriteEffects` has no union or merge, add one (`func merging(_:)`), with a unit test in the existing effects tests if there are any. Do not patch the cached list value: D-8 rejects that.
Check: the test passes; T-8 to T-10 still pass.

**F-3 — Disable the Add dialog's graph group on retry** (from UX-2)
Files: `SourceArtifactsSection.swift`, `SourcePageArtifactsView.swift`, `SourcePageModelTests.swift`.
Failing test first: `addArtifactRetryIgnoresGraphChoice`: after a refused file (`draft.createdID` set), unchecking every graph still leaves `canSubmitDraft` true, and `create()` makes no link call.
Do: while `draft.createdID != nil`, disable the group (`.disabled(...)` in `addForm`), and do not apply `draftNeedsGraph` in `canSubmitDraft` or `create()`. D-4 stands: no re-link.
Check: the test passes.

**F-4 — Share the checklist frame, row and toggle between the Artifact checklist and the graph links** (from MNT-2; same-pass cleanup, supersedes D-12's copied frame)
Files: `macos/App/DesignSystem/Recipes/ArtifactChecklist/` (new shared pieces), `PVArtifactChecklist.swift`, `macos/App/Features/Sources/SourceArtifactGraphLinks.swift`, `macos/App/DesignSystem/README.md`, `L10n.swift` and `Localizable.xcstrings` if the locked spoken value gets a neutral key. Skill: `add-ui-component`, `add-localized-string`.
Do: in the ArtifactChecklist recipe folder, extract:
- the bordered, capped, divided row list as one recipe view that owns the 264pt cap;
- the checklist row checkbox, which takes the label, description, id, the `excluded` binding and an optional lock reason, and owns the toggle and the locked `accessibilityValue`. The toggle arithmetic can sit on `PVChecklistState`.
`PVArtifactChecklist` composes them with its All row and thumbnails, and `SourceArtifactGraphLinks` with graph rows. Make `maxRowsHeight` private again or move it onto the shared frame. Give the locked spoken value a neutral key, so the graph rows no longer read an Artifact-worded translator comment, and keep the English identical. Pick one way to disable (MNT-6 nit): the parameter or `.disabled`. Update the ArtifactChecklist row in the DesignSystem README. No visible change.
Check: `rg -n "excluded.remove\(" macos/App` prints exactly one line, in the recipe folder. `rg -n maxRowsHeight macos/App/Features` prints nothing. `PVArtifactChecklistTests` still pass.

**F-5 — Link edit ignores deleted graphs and a failed reload** (from CON-1, CON-3)
Files: `SourceArtifactsSection.swift`, `SourcePageModelTests.swift`.
Failing tests first:
- `cardSaveSkipsAGraphDeletedDuringEdit`: check a graph that is not linked, delete it from the session, save, and expect no link call for it.
- `cardSaveReloadFailureOffersReadAgain`: a link succeeds, the unlink is refused, the next `listEvidenceGraphs` fails, and `linkEdit.loadState` becomes `.failed`.
Do: read live graph ids once (a helper on the section that filters `edit.graphIDs` by `context.session.evidenceGraph(id:) != nil`). Use it in `checkedGraphIDs` and in `saveFields`' `toLink` and `toUnlink`. After `graphListReloaded()`, treat `context.session.evidenceGraphsHandle?.error != nil` as failure, the same as nil.
Check: both tests pass.

**F-6 — Name the membership checks and share the test fixtures** (from MNT-3, MNT-4)
Files: `macos/App/Platform/GenealogyStore.swift` (beside `hasArtifact`), `SourceArtifactsSection.swift`, `macos/ProvenenciaTests/EvidenceGraphFixtures.swift`, `SourcePageModelTests.swift`, `SourceGraphFormModelTests.swift`.
Do: add `func links(_ artifactID: String) -> Bool` and `func isOnlyArtifact(_ artifactID: String) -> Bool` to `CatalogEvidenceGraph`, and read them at every `artifactIDs.contains(` and `artifactIDs == [` site in `macos/App/Features`. Add `unlinkBlock(_ artifactID:, citations:)` to `EvidenceGraphFixtures.swift`, use it from both suites, and delete `smithBlock` and `SourceGraphFormModelTests.block`.
Check:
- `rg -n "artifactIDs.contains\(|artifactIDs == \[" macos/App/Features` prints nothing.
- `rg -n "CatalogEvidenceGraphUnlinkBlock\(" macos/ProvenenciaTests` prints only the line in `EvidenceGraphFixtures.swift`.

**F-7 — Card save-error copy follows what is left to save** (from UX-1, UX-3, UX-4; user decision 2026-10-25)
Files: `SourceArtifactsSection.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`. Skill: `add-localized-string`.
Failing test first: extend T-10 (`cardSaveRefusedKeepsEditAndRereadsLocks`). Before the block is cleared, `fieldsDirty("P3")` is false, and the callout is the "keeps" body naming Jones's graph, with no "save again". Also add `cardFieldsOnlySaveErrorSaysChangesKept`: on a one-graph Source, `updateArtifact` fails and the callout is the fields-only body.
Do: the card stays in edit in every case. After the catch re-reads the locks, pick the body:
- (a) No write succeeded (fields-only, or the first write failed): new `artifactCardSaveErrorKept` = "Your changes are kept. %@".
- (b) A link or unlink was refused, and `fieldsDirty(id)` is false after the re-read: new `artifactCardSaveErrorGraphKeeps` = "%1$@ keeps this artifact. %2$@". The first argument is the refusing graph's name from `context.session.evidenceGraph(id:)`, the second is the `L10n.Errors` copy.
- (c) Otherwise, the existing `artifactCardSaveErrorMessage`.
Translator comments name each argument. Remember the refusing graph id in the loop.
Check: the tests pass; `localizable` check PASS.

## Round 1 — fix
Head `13994bdb`. Gate PASS (orchestrator re-ran). F-1 to F-7 done, each failing test run red first. Deviations:
- The F-6 grep still matches `CitationComposerModel.swift:1114`, a local `[String]`, not a graph. Accepted.
- F-7: when the first write is refused and nothing is left, the body is (b), "{graph} keeps this artifact". Accepted; it matches the user's choice.
- `ab67bbab` alone does not compile, because the L10n rename landed one commit early. The head builds.
- `PVChecklistFrame` takes `id:` as a key path.
- The locked spoken value moved to the neutral key `checklist.lockedValue`.

CON (minor only, read by the orchestrator, not re-reviewed). CON-1 **Resolved**: `liveGraphIDs` feeds `checkedGraphIDs`, `toLink` and `toUnlink`. CON-3 **Resolved**: after `graphListReloaded()`, a handle error counts as failure.

## Round 2 — review
Reviewers: MNT (MNT-1 blocker, MNT-2 major), UX (UX-1, UX-2 majors), SEC+PRF (PRF-1 major). Not re-sent: CON (only minors). Scope `5b19a989...13994bdb`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | (round 1) | Resolved | `editingArtifactID == id` guard before both writes; `cardSaveErrorStaysOnTheCardThatSaved` |
| MNT-2 | MNT | major | (round 1) | Resolved | `PVChecklistFrame` + `PVChecklistCheckbox` in the recipe folder; greps hold |
| UX-1 | UX | major | (round 1) | Resolved | `saveErrorBody` picks (a)/(b)/(c) after the re-read |
| UX-2 | UX | major | (round 1) | Resolved | `draftGraphsLocked` disables the group; `draftNeedsGraph` off on retry |
| PRF-1 | PRF | major | (round 1) | Resolved | effects merged, one `session.apply`; `cardSaveReloadsTheGraphListOnce` |

Minors fixed in round 1, confirmed by their lens in round 2: UX-3 and UX-4 (both covered by `saveErrorBody` (a)/(b)), MNT-3 (`links(_:)`/`isOnlyArtifact(_:)`), MNT-4 (`unlinkBlock` fixture). All **Resolved**. No new findings.

Nits (logged): `UX-6 · SourcePageArtifactsView.swift:85 · disabled retry group keeps its hint with no reason` · `MNT-7 · CatalogMutation.swift:101 · merging lists fields by hand; doc comment for the next field` · `MNT-8 · PVChecklistRows.swift:187 · showsTitle/speaksRef flags only the Artifact checklist uses`

## Result
Gate PASS (`13994bdb`, `main` unmoved at `846fa656`) · CI green (test, localizable) · open: none · follow-ups: UX-5 · to user: UX-1/UX-3/UX-4 answered 2026-10-25 (context copy) · CON miss: none
