# Ship log — S15-09 Artifacts show their metadata on the Source page

**Plan:** [`S15-09-artifact-metadata-boards.md`](S15-09-artifact-metadata-boards.md) · **Branch:** `s15-09-artifact-metadata-boards` · **PR:** #802
**State:** ready

## Round 0 — implement
Base `main` @ `064e1f393` (plan written against `83895e4`).
Head `da65589c2`. PR #802 (draft). Gate PASS (orchestrator run against `064e1f393`). Report deviations: `MetadataBoardStyle.suggestionsHeadingTopPadding` (a measure D-10 does not list); `MetadataBoardLabel.width` optional; `PVReorderHandle.width` is `nonisolated`; `FakeStore` uses `uniquingKeysWith`; manual check (§10) not run; AC-1.2, AC-1.7, AC-1.8 built, not seen.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX: the diff touches `macos/App/**` and `Localizable.xcstrings`. SEC+PRF: `macos/App/Features/**` models (`SourcePageModel`, `MetadataBoardModel`) and the workspace query cache patch (`CatalogQueryRegistry`, `CatalogLastUpdated`); the board renders up to 60 Artifact boards from catalog data.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | A board write patches a stale cached workspace, so a page-only Artifact disappears | Fix now → F-1 | verified: `show(_:effects:)` (`SourceArtifactsSection.swift:783`) writes `context.workspace` only; `.sourceWorkspace` has no `refreshes`; the refused-file `catch` skips `refreshCoverFromStore()`. Carries a failing test; the fix round runs it red |
| UX-1 | UX | major | Hiding a board (collapse, or the card's edit) mid-edit brings the row back blank with Save disabled | Fix now → F-2 | verified: `MetadataBoardTextEditor.draft` is seeded only in `onEdit` and `.onChange(of: isEditing)` with no `initial`; nothing ends the board's edit on collapse or `openEdit` |
| UX-2 | UX | major | A refused Artifact value reads "Invalid source metadata." | Fix now → F-3 | verified: Artifact and graph writes return `sourcemetadata.invalid` (`api/ffi/handlers/metadata.go:243`); xcstrings value is "Invalid source metadata." |
| CON-2 | CON | major | The plan has no Landed note | Fix now → F-4 | verified: only line 3 says Landed; DoD line 306 and STEP-9 ask for the note |
| MNT-1 | MNT | major | Third copy of the metadata row builder; more `CatalogArtifact` test builders | Fix now → F-5 (same-pass cleanup) | verified: `rg "private func (artifact|metadataRow|entry|boardRow|boardArtifact|pageSession)\("` lists 3 entry builders, 7 artifact builders, `pageSession` |
| MNT-2 | MNT | minor | `MetadataBoardView` branches on the style for two measures | Fix now → F-6 | minor in code this PR adds; D-10 says no style branch in the view |
| MNT-3 | MNT | minor | `metadataRows(for:)` / `setMetadataRows` switch through `default` | Fix now → F-6 | verified: `GenealogyStore.swift:1278`, `:1296` |
| UX-3 | UX | minor | VoiceOver loses the disclosure count; suggestion input and row buttons don't name the field | Fix now → F-7 | minor in code this PR adds or changes (`copyDetails`, the rewritten suggestion row, `MetadataBoardView.swift`) |
| CON-3 | CON | minor | T-6 does not prove the Source board's edit survives a note save | Fix now → F-8 | verified: test ends with only `board.editingFieldID` |
| CON-4 | CON | minor | `macos-client-patterns.md:89` list of kept mutations omits `.wroteMetadataBoard` | Fix now → F-4 | verified at line 89 |
| PRF (from MNT outside lens) | PRF | minor | `syncBoards` reassigns `boards` every sync | Rejected | SEC+PRF review: every card already reads `context.workspace` through `items`, so cards re-rendered each sync before this PR; bounded by one Source's Artifacts (60) |
| UX (from CON outside lens) | UX | minor | Card keeps `.pvLiftHover(scale: 1.02)` while hosting a board's inputs and drag list | Logged → manual check | no evidence of a defect without the app; added to §10 manual check for the user |
| UX-4 | UX | nit | Saved-row reorder is drag-only on every board | Follow-up → `docs/ideas/keyboard-reorder.md` | pre-existing on the Source board; the user chose an idea file |

Nits (logged): `MNT-4 · SourcePageChrome.swift:3 · doc comment cites the moved paddings` · `MNT-5 · MetadataBoardView doc comment · say compact draws a trailing Add and no empty state` · `MNT-6 · released_metadata_scope_test.go:94,104 · indexOf → indicesOf; setValue → setProviderValue` · `MNT-7 · SourcePageModelTests T-5 · variable image holds repositoryField()` · `CON-5 · SourcePageArtifactsView.swift:457-476 · PVDivider inside the padded VStack, D-12 puts it above`. Nits ride along in F-9 since every one is in a file this PR adds or changes.

### Fix plan 1

**F-1 — Page-only Artifact writes reach the cached workspace** (from CON-1)
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Failing test first: `aBoardWriteKeepsAnArtifactShownOnlyOnThePage` — the test in CON-1 (refused file leaves "Scan" on the page; a Source board save then `sync` from the session's handle must still show "Scan"). Run it red at head first.
Do: every place in `SourceArtifactsSection` that changes `context.workspace?.artifacts` locally (`show`, `replace`, and any other — `rg -n "context\.workspace\?\.artifacts" macos/App/Features/Sources`) writes through one private helper that also writes the cache with `context.session.setQueryValue(.sourceWorkspace(...), value:)`, the call `confirmResourceDelete`'s `"artifact"` case already uses. Reuse that call shape; do not add a second mechanism.
Check: the test passes; `rg -n "context\.workspace\?\.artifacts(\[|\.append|\.remove)" macos/App/Features/Sources` prints only lines inside the helper; scripts/check-pr.sh PASS.

**F-2 — A hidden board ends its row edit, and a recreated editor shows the value** (from UX-1)
Files: `SourcePageArtifactsView.swift` (`copyDetails` disclosure binding), `SourceArtifactsSection.swift` (`openEdit`/`beginEdit`), `MetadataBoardView.swift` (`MetadataBoardTextEditor`), `SourcePageModelTests.swift`.
Failing test first: `closingCopyDetailsEndsTheBoardEdit` — begin an edit on an Artifact board row, close that card's Copy details through the model's open-state API, expect `board.editingFieldID == nil`; and `openingTheCardEditEndsItsBoardEdit` — same with the card's edit.
Do: closing the disclosure and opening the card's edit call `board.cancelEdit()` for that Artifact's board — in the model (`SourceArtifactsSection`), not the view, so the tests reach it. In `MetadataBoardTextEditor`, change to `.onChange(of: isEditing, initial: true)` so an editor created already in edit seeds `draft` from `entry.valueText`.
Check: tests pass; check-pr PASS.

**F-3 — The metadata refusal names no level** (from UX-2)
Files: `macos/App/Resources/Localizable.xcstrings`, `macos/App/Platform/L10n.swift` (default value). Skill: `add-localized-string`.
Do: keep the key `error.sourcemetadata.invalid`; change the English to "This value can't be saved for this field." in both the catalog and `L10n` `defaultValue`. Update any test asserting the old text (`rg -n "Invalid source metadata" macos`).
Check: `rg -n "Invalid source metadata" macos` prints nothing; check-pr PASS (localizable).

**F-4 — Plan Landed note and the mutation list** (from CON-2, CON-4)
Files: `docs/deployment-plan/spike-15/prs/S15-09-artifact-metadata-boards.md`, `docs/macos-client-patterns.md`.
Do: add a `## Landed` section after the header, shaped as `S15-26-metadata-board-takes-an-owner.md:32`, listing every difference from the plan: `MetadataBoardStyle.suggestionsHeadingTopPadding`; optional `MetadataBoardLabel.width`; `PVReorderHandle.width` is `nonisolated`; `FakeStore` `uniquingKeysWith`; the stacked row's control spacing; the divider placement (after F-9, say where it ended up); plus this round's changes (F-1 helper writing the cache, F-2, F-3 copy, F-5 fixtures superseding any Out line, F-6 style measures). In line 89, add `.wroteMetadataBoard` after `.mutatedSourceWorkspace`.
Check: `grep -n "^## Landed" <plan>` prints one line; check-pr PASS (plan-docs).

**F-5 — One set of test builders for metadata rows and Artifacts** (from MNT-1, same-pass cleanup)
Files: new `macos/ProvenenciaTests/MetadataBoardFixtures.swift` (or `CatalogFixtures.swift`), `MetadataBoardModelTests.swift`, `CatalogQueryRegistryTests.swift`, `SourcePageModelTests.swift`, `SourceGraphFormModelTests.swift`, `SourceCopyModelTests.swift`, `EvidenceGraphModelTests.swift`, `CitationComposerModelTests.swift`, `CatalogArtifactTests.swift`. Skill: `add-swift-test`.
Do: one `CatalogMetadataEntry` builder (field, value, suggested, order; `hasValue = !value.isEmpty`) and one `CatalogArtifact` builder with defaulted parameters covering every existing call (id, sourceID, label, description, file, fileID — read each copy's body). Move every copy onto them and delete the copies; replace the inline `CatalogMetadataEntry(` literals in the metadata tests where the change is mechanical. Drop `pageSession`: build the session once per test and pass it to `makeModel`. Find copies by body: `rg -nE "private (static )?func \w+\([^)]*\) -> (CatalogArtifact|CatalogMetadataEntry)" macos/ProvenenciaTests`.
Check: that rg prints nothing (only the shared fixtures, which are not `private`); check-pr PASS with the same test count or more.

**F-6 — Style carries every measure; owner switches name every level** (from MNT-2, MNT-3)
Files: `MetadataBoardView.swift`, `MetadataBoardModel.swift` or wherever `MetadataBoardStyle` lives, `macos/App/Platform/GenealogyStore.swift`.
Do: add `suggestionLabelWidth: CGFloat?` and `suggestionControlSpacing: CGFloat` to `MetadataBoardStyle`, fold `MetadataBoardLayout.columnSpacing` into it and delete `MetadataBoardLayout`; the view holds no `style.stacksSuggestionLabel ?` ternary for a measure. In `GenealogyStore`, `setMetadataRows(_:for:)` guards on `metadataRows(for:) != nil` and writes once; `metadataRows(for:)` lists `case .graph, .citation:` instead of `default`.
Check: `rg -n "stacksSuggestionLabel \?" macos/App` and `rg -n "MetadataBoardLayout" macos` print nothing; `.regular` measures unchanged (label 132, columns as before); check-pr PASS.

**F-7 — VoiceOver names the count and the field** (from UX-3)
Files: `SourcePageArtifactsView.swift`, `MetadataBoardView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: the disclosure's accessibility label carries the count ("Copy details of %1$@, %2$lld values" with a plural rule, or an `accessibilityValue` with the count). The suggestion input gets `.accessibilityLabel(entry.field.label)`. The dismiss, edit, delete row buttons name the field (`L10n` functions taking the field label, following `editArtifactLabel(ref:)`), on every board.
Check: check-pr PASS (localizable).

**F-8 — T-6 proves both edits survive** (from CON-3)
Files: `SourcePageModelTests.swift`.
Do: before `model.notes.add()`, `model.metadata.beginEdit(fieldID: b.id)`; after the final `sync`, `#expect(model.metadata.editingFieldID == b.id)`.
Check: test passes.

**F-9 — Nits** (MNT-4 – MNT-7, CON-5)
Do: fix the `SourcePageChrome` doc comment; extend the `MetadataBoardView` doc comment; rename `indexOf` → `indicesOf`, `setValue` → `setProviderValue`; rename T-5's `image`; move `PVDivider` above the padded `VStack` per D-12.
Check: check-pr PASS.

## Round 1 — fix
Head `7133a83d9`. Gate PASS (orchestrator run against `064e1f393`). F-1 – F-9 done. F-1 ran red at the round-0 head (`items → []`); F-2's `openingTheCardEditEndsItsBoardEdit` ran red with the fix line removed. Deviations: F-7 hosts the saved row's pencil in the row's trailing controls (`showsEditControl: false` on `PVInlineEdit`) so it can carry the field name; F-9's divider move leaves `space4` less above the rule; F-5 adds `catalogFileArtifact` and turns `CitationComposerModelTests`' builder into a plural `scans(_:)`; commits grouped by file set.

## Round 2 — review
Reviewers: CON (CON-1, CON-2 majors), UX (UX-1, UX-2 majors), MNT (MNT-1 major). SEC+PRF not sent: no Fix-now blocker or major. Scope `da65589c2...7133a83d9`.

Round-1 majors: CON-1 **Resolved**, CON-2 **Resolved** (CON-4 closed with it), UX-1 **Resolved**, UX-2 **Resolved**, MNT-1 **Resolved**. Minors read by the orchestrator in the fix commits: MNT-2, MNT-3 **Resolved** (`MetadataBoardLayout` gone, no style ternary; switches name `.graph, .citation`); UX-3 **Resolved** (count as `accessibilityValue`, field-named input and row buttons); CON-3 **Resolved** (T-6 checks `model.metadata.editingFieldID == b.id`); nits F-9 done.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-8 (also CON outside lens) | MNT | minor (below round-2 bar) | `patchArtifacts` is the third "set `context.workspace` and `setQueryValue(.sourceWorkspace…)`" copy (`SourcePageContext.refreshCoverFromStore:144`, `SourcePageModel.swift:222` delete case) | Fix now → F-10 (same-pass cleanup) | verified: the three bodies read above; this PR added the third copy. The same-pass rule applies whatever the severity, so it rides a Fix plan 2 although no round-2 blocker exists |
| UX-5 (from CON outside lens) | UX | minor | F-7's moved pencil lost `restingAlignment.editControlTopPadding` | Fix now → F-11 | verified: `PVInlineEdit.swift:89-90` pads its own pencil; `MetadataBoardView.swift:334` draws one in `restingTrailing` with no padding. A visual regression on the Source board this round introduced |
| UX-6 (from UX outside lens) | UX | minor | F-3's "This value can't be saved for this field." also shows for stale-id reorder, clear and dismiss | Fix now → F-12 | verified: `sourcemetadata.go:320-341` returns `ErrInvalid` on those paths; the string is this PR's |

### Fix plan 2
No round-2 blocker. Fix plan 2 carries a same-pass cleanup the rule makes Fix now regardless of severity, and two minors the round-1 fixes introduced in code this PR changes. No round 3: the orchestrator reads the commits.

**F-10 — One method writes the page's workspace and its cache** (from MNT-8)
Files: `macos/App/Features/Sources/SourcePageContext.swift`, `SourceArtifactsSection.swift`, `SourcePageModel.swift`.
Do: add one `SourcePageContext` method (e.g. `setWorkspace(_:)`) that sets `workspace` and calls `session.setQueryValue(.sourceWorkspace(project:sourceId:), value:)`. Move `refreshCoverFromStore`, `patchArtifacts` and the `"artifact"` delete case onto it; keep the delete case's `apply(workspace)` call where it is.
Check: `rg -n "setQueryValue\(\s*$|setQueryValue\(CatalogQueryKey\.sourceWorkspace|sourceWorkspace\(\s*$" macos/App/Features/Sources` and `rg -n -A2 "setQueryValue" macos/App/Features/Sources` show a `.sourceWorkspace` write only inside the new method; `aBoardWriteKeepsAnArtifactShownOnlyOnThePage` still passes; check-pr PASS.

**F-11 — The saved row's pencil goes back inside `PVInlineEdit`** (from UX-5)
Files: `MetadataBoardView.swift`, `L10n.swift` (and `PVInlineEdit.swift` only if needed).
Do: restore `showsEditControl` (default true) so `PVInlineEdit` draws the pencil with its own top padding, and pass the field-named label through its existing `editLabel` parameter: make `L10n.Sources.editMetadataValueLabel(field:)` return the type `editLabel` takes. Only if that type cannot carry a formatted label, widen `editLabel` to the type `PVIconButton`'s `label` takes, at every call site. Keep identifier `<prefix>.<fieldID>.edit`. `restingTrailing` returns to the trash button only.
Check: `rg -n "showsEditControl: false" macos/App/Features/Sources` prints nothing; check-pr PASS.

**F-12 — The metadata refusal fits every write that returns it** (from UX-6)
Files: `Localizable.xcstrings`, `L10n.swift`. Skill: `add-localized-string`.
Do: change the English of `error.sourcemetadata.invalid` to "This metadata change can't be saved." in both places; key unchanged. Update the plan's Landed note line for F-3.
Check: `rg -n "can't be saved for this field" macos docs/deployment-plan/spike-15` prints nothing; check-pr PASS.

## Round 2 — fix
Head `48be23fef`. Gate PASS (orchestrator run against `064e1f393`; `origin/main` has not moved). F-10, F-11, F-12 done; orchestrator read both commits. F-10: `SourcePageContext.setWorkspace(_:)`, the only `setQueryValue` under `Features/Sources`. F-11: a `LocalizedStringResource` cannot interpolate under the `localizable` lint, so `PVInlineEdit.editLabel` widened to `any PVCopy` (the fallback F-11 allowed); unused `L10n.Sources.editMetadataValue` deleted. At the default `.top` alignment `editControlTopPadding` is 0, so UX-5 was a position change, not lost padding. F-12: "This metadata change can't be saved."

## Round 3 — trim (user, 2026-11-12)

After the user looked at the Copy details on each card, they rejected it: cards grow too tall at 60 Artifacts. The work is re-cut in brief [S15-D14](../design/S15-D14-source-page-artifacts.md) and PRs S15-33 – S15-36 (map commit `a8282e8a8`). The user chose to trim #802 before merge so the rejected chrome never reaches `main`. This is a user-directed round, not a review round.

### Fix plan 3

**F-13 — Remove the per-Artifact metadata chrome; keep the board write's cache patch and its fixes** (user)
Files: `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `SourceArtifactsSection.swift`, `SourcePageModel.swift`, `MetadataBoardView.swift`, `MetadataBoardHost.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/App/DesignSystem/README.md`. Skills: `add-localized-string`, `add-swift-test`.
Do:
- **Goes.** The card's Copy details disclosure (`SourcePageArtifactsView.copyDetails` and its call); `SourceArtifactsSection.boards`, `openBoardIDs`, `board(for:)`, `setCopyDetails(open:for:)`, `syncBoards` and its call in `SourcePageModel.apply`, and the `boards[id]?.cancelEdit()` lines in `openEdit` / `beginEdit`; `MetadataBoardStyle` entirely (the board draws exactly as `.regular` did, with no style parameter), with every member only `.compact` used and `MetadataBoardLabel.width`'s optionality if only the stacked label needed it; `SourcePageMetadataHost`'s owner parameter (it serves the Source only; the doc comment says so); `L10n.Sources.artifactCopyDetails`, `artifactCopyDetailsValue`, `artifactCopyDetailsLabel` and their catalog keys; the tests that need a per-Artifact board (`artifactBoardsFollowTheWorkspace`, `artifactBoardSaveWritesItsOwnerWithoutReload`, `closingCopyDetailsEndsTheBoardEdit`, `openingTheCardEditEndsItsBoardEdit`, `addDialogOnAnArtifactBoardWritesTheArtifact`, `deletingAnArtifactDropsItsBoard`, `invalidValueOnAnArtifactRowStaysOnTheRow`, `boardsSurviveAReload`). `DesignSystem/README.md:153` drops "the Source page's artifact rows" from `pvExpandedState`'s callers.
- **Stays.** `.wroteMetadataBoard` and `CatalogQueryPatch.metadataBoard()`; `CatalogSourceWorkspace.artifactMetadata`, `metadataRows(for:)`, `setMetadataRows(_:for:)` (S15-34 patches an Artifact's rows through them) and their tests in `CatalogQueryRegistryTests`; `SourcePageContext.setWorkspace` and `patchArtifacts`; `aBoardWriteKeepsAnArtifactShownOnlyOnThePage`; the Go scope test; `CatalogFixtures.swift`; the field-named VoiceOver labels and `editLabel: any PVCopy`; the reworded refusal; `PVReorderHandle.width`; the board chrome's move into `MetadataBoardView.swift`.
- **Rewrite.** `boardWriteKeepsAnEditBegunElsewhere` keeps AC-2.2 – AC-2.4 on the Source's board alone (a Source edit begun after a save, and one open across a note save, both survive).
- **Records.** The plan's Status stays "Landed in #802" and its **Landed** note gains a "Trimmed before merge (user, 2026-11-12)" paragraph: what went, what stayed, that US-1 moved to S15-34 / S15-35 (S15-D14). `completed.md`'s S15-09 paragraph says what actually lands. The map row's status becomes "Landed #802" when the PR merges (leave "In review, trim pending" now; the orchestrator updates it at finish). Version stays 0.11.0.
Check: `rg -n "artifactCopyDetails|MetadataBoardStyle|openBoardIDs|setCopyDetails|syncBoards|\.compact\b" macos` prints nothing in `macos/App` (a `.compact` on an unrelated kit type is fine; list any hit); the Source page's Metadata section draws as on `main`; scripts/check-pr.sh PASS.

## Round 3 — fix
Head `4660065c8`. Gate PASS (orchestrator run against `064e1f393`). F-13 done: the Check `rg` prints only `.compact` on unrelated kit types; eight Artifact-board tests removed; `boardWriteKeepsAnEditBegunElsewhere` rewritten for the Source board.

## Round 3 — review
Reviewers: CON, MNT (fresh), scope `2dc869a9d...4660065c8`. UX and SEC+PRF not sent: the round only removes chrome and state, and CON checked the Source section's measures against `main`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `keyboard-reorder.md` (added on this branch) and `MetadataBoardDialogs`' doc describe many boards that the trim removed | Fix now → orchestrator (idea file), F-14 (comment) | verified: `rg "Copy details\|many boards" docs/ideas macos/App` |
| MNT-2 | MNT | major | Style leftovers: `width: 132` at two call sites, the label/controls split, a stale "stacked label" comment | Fix now → F-14 | verified: `MetadataBoardView.swift` 84, 187, 171-193 |
| CON-1 | CON | minor | AC-2.1's page-level no-reload assertion went with T-5 | Fix now → F-14 | verified: no `getSourceWorkspaceCalls` check on a board save |

Nits (logged, ride F-14): `MNT-3 · PVReorderableList.swift:28 · drop nonisolated if it builds` · `MNT-4 · plan Landed bullet "MetadataBoardLabel.width is optional" → "(reverted by the trim)"` · `CON-2 · same as MNT-2's comment`.

### Fix plan 4

**F-14 — Finish the trim's cleanup** (from MNT-1, MNT-2, CON-1, nits)
Files: `macos/App/Features/Sources/MetadataBoardView.swift`, `MetadataBoardDialogs.swift`, `macos/App/DesignSystem/Components/ReorderableList/PVReorderableList.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`, the plan's Landed note.
Failing test first: none (no behavior change), except CON-1's assertion, which passes at head and must fail if `SourcePageMetadataHost.didWrite` applies `context.applyWrite(effects)`: check that by hand once, then restore.
Do: `MetadataBoardLabel` gets `static let width: CGFloat = 132` and loses its `width` parameter; both literals go. `MetadataBoardSuggestionRow` draws one flat `HStack` again (label, input, Save, dismiss), with no `label` / `controls(fieldID:)` split and no "stacked label" comment. `MetadataBoardDialogs`' doc drops "so a page with many boards still has one of each" ("each targets the board that asked"). Try removing `nonisolated` from `PVReorderHandle.width`; keep it only if the build warns or fails, and say which. In `boardWriteKeepsAnEditBegunElsewhere`, record `store.getSourceWorkspaceCalls` before the first save, `settleWrites()`, and expect it unchanged. In the plan's Landed note, mark the `MetadataBoardLabel.width` bullet "(reverted by the trim)".
Check: `rg -n "width: 132|stacked label|many boards" macos/App` prints nothing; scripts/check-pr.sh PASS.

## Round 4 — fix
Head `7ce528c53`. F-14 done: `MetadataBoardLabel.width` a static 132; the suggestion row one flat `HStack`; `MetadataBoardDialogs`' doc fixed; `nonisolated` dropped from `PVReorderHandle.width` (no warning without it); the Landed bullet marked reverted. The `rg` Check prints only `PVSkeletonBar(width: 132, …)` in `SourcesGraphButton.swift` and `SourcesGraphCountLine.swift`, unrelated. CON-1's assertion: the test now makes the query active (`session.apply(location:)`) and checks the count right after the board save; with `context.applyWrite(effects)` added to `didWrite` it fails (`4 == 2`), restored it passes. MNT-1, MNT-2, CON-1 **Resolved** (orchestrator read `68961be53..7ce528c53`). No round-2 review: the round had no blocker.

Merged `origin/main` (`7276092ba`, docs only) as `a1b882be9`; gate PASS against `7276092ba`.

## Result
Gate PASS · CI green (`c637ca4b1`) · open: none · trimmed to the board write's cache patch and its fixes (user, 2026-11-12); the per-Artifact chrome is re-cut as S15-33 – S15-36 under brief S15-D14 · follow-ups: UX-4 → [`docs/ideas/keyboard-reorder.md`](../../../ideas/keyboard-reorder.md) · to user: — · manual check §10 steps 1 and 4 (the Source's own board) still apply; the Artifact steps do not
