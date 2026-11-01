# Ship log — S14-14 Copy and VoiceOver on graph work

**Plan:** [`S14-14-graph-work-voiceover.md`](S14-14-graph-work-voiceover.md) · **Branch:** `s14-14-graph-work-voiceover` · **PR:** #675
**State:** ready

## Round 0 — implement
Head `bc8f22ff` (local, not pushed). Gate not run.

Committed: STEP-1 (`d1d0eb0c`), STEP-2+3 (`8b0fd908`), STEP-4 (`6faef1e9`), STEP-5 (`ff024728`), STEP-6 (`bc8f22ff`).
Uncommitted in the tree: STEP-7, STEP-8, and the working parts of STEP-9 (four counted plural keys, T-18 Go subtest, D-21 script check). Not started: STEP-10–13, preview `narrative` removal.

**Stopped on:** STEP-9's probe. A `deleteImpact.kind.<kind>` key whose `variations.plural` one/other values carry no number ("citation" / "citations") is a hard catalog compile error, not a warning:
`Plural variation requires referencing the number in the string. … use separate top-level strings for one and greater than one.` (all 18 kinds).
`deleteImpact.via.unknown` (substitution shape) is unprobed on its own.

Deviations so far: `editingArtifactIsOneCardAtATime` gains `confirmEditSwitch()` (beginEdit now asks); `unknownViaStillListsRefs` expects "2 items" per D-19; `pvFieldLabel`/`pvSheetClose` use `EnvironmentKey` structs to match `PVWidthClass`, `PVSheetClose` default via `nonisolated(unsafe)` as `pvContextMenuDismiss`.

Needs: D-19 revised through `plan-pr`.

**Plan revised** (`0e7bbd81`, user: "Probe substitution, revise"). Scratch-catalog probe: `%#@noun@` substitution compiles and resolves. D-19 now uses it; review round 3 (scoped) answered every finding. Round 0 resumes at STEP-9.

### Round 0 resumed
Head `28af621a`. Implementer gate PASS. STEP-7 – STEP-13 done; draft PR #675.
Deviations:
1. D-12's `beginEdit` guard: the plan's "return at once when same card or `savingID != nil`" broke `cardEditLocksEveryBlock` and `cardSaveErrorStaysOnTheCardThatSaved`; guard now asks only when another card is in edit, no save is running, and the card is dirty (`0eb993ec`). Reads against D-12's wording → reviewers decide.
2. `artifactSwitchShowsAFailedCitationLoad` (not named in the plan) now asserts `citationLoadError` per D-16.
3. `activeCitationRef` and `announceIdentityChange()` share a private `listedCitation(_:)` beside `citationRef(for:)`.
4. The `narrative` preview in `PVDeleteImpact.swift` is unchanged and reads "item".
5. STEP-7/8 split by hunk; `graphCopyNamesTheGraph` rode in the STEP-9 commit.
Orchestrator gate on `28af621a`: PASS (against `be51621f`; xcode-test 92s).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC routes on `macos/App/Platform/ProjectFiles.swift`; PRF on `macos/App/Features/**` models; UX on `macos/App/**` and `Localizable.xcstrings`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Read-only `PVInput` in a `PVField` speaks the name, not the value | Fix now → F-1 | verified: read-only branch is `Text(text)` under `pvFieldNamed()`; no `accessibilityValue` in `PVInput.swift`; callers `PropertiesView.swift:790` (derived key), `OnboardingIdentifyView.swift` (locked researcher) |
| MNT-1 | MNT | major | Third copy of the `"%1$@, %2$@"` spoken-join key | Fix now → F-2 (same-pass cleanup) | verified: `designSystem.announcement.join`, `conclusions.a11y.list`, `lastUpdated.a11y.join` all `"%1$@, %2$@"`; this PR adds the third. Supersedes D-6's "a kit-owned join" beside the others. Spoken output unchanged |
| UX-4, CON-1 | UX, CON | minor → major | Card switch during a save drops the saving card's buffers; a later failure is never shown. Implementer's guard reads against D-12 | Fix now → F-3 | Re-rated: the deviation sets aside a Decision (D-12: same card or `savingID != nil` → nothing). Verified `beginEdit` at `SourceArtifactsSection.swift:182` asks only when `savingID == nil`; `cardSaveErrorStaysOnTheCardThatSaved` encodes the loss (`cardSaveError == nil`). D-12 binds; its two broken tests are updated to the safer behavior |
| UX-2 | UX | minor | Every input in the composer's compound Value editors is named "Value" | Fix now → F-4 | verified: `cascadeField` is `VStack { Text(label); PVInput }`, no own `PVField`; outer `PVField(label: valueLabel)` at `CitationComposerAuthoringBox.swift:368` |
| MNT-2 | MNT | minor | Graph-name lookup and graph-surface test have copies outside `WorkspaceLocationLabels` | Fix now → F-5 | file this PR adds |
| MNT-3 | MNT | minor | Module-wide trimming `nilIfEmpty` clashes with two same-named helpers | Fix now → F-5 | file this PR adds |
| MNT-4 | MNT, UX (outside lens) | minor | Kind-noun `defaultValue`s are bare English, not `%#@noun@` as D-19 says | Fix now → F-6 | verified `kindNouns` `defaultValue: "source"` etc.; D-19 binds |
| UX-3 | UX | minor | Discard strings' comments describe only an outside click; card confirm names no card | Fix now (comments) → F-7; Follow-up (named card message) | comments in strings this PR reuses for new cases |
| CON-3 | CON | minor | `PVSheetChrome`'s ask-then-run wiring has no unit test | Fix now → F-8 | minor in a file this PR changes |
| CON-2 | CON | minor | Planted-key check is a manual run, no committed test | Follow-up | no Python test exists in `scripts/` (a new mechanism); plan's T-14 is a Script check |

Nits (logged): `CON-4 · CitationComposerModel.swift:126 · leaveGate private(set)` · `MNT-5 · PVCallout.swift:54 · title and titleCopy hold the same input` · `MNT-6 · PVSelect.swift:98 · @Environment among private lets` · `MNT-7 · WorkspaceToolbar.swift:244 · one switch for label and crumb id` · `MNT-8 · WorkspaceToolbarBreadcrumbTests · move labels tests to WorkspaceLocationLabelsTests` · `UX-5 · addPropertyUnavailable lowercase "evidence graph" (matches plan and canvas callout)` · `UX-6 · citationLoadFailedUnnamed lowercase "citation" (matches plan § 6 Strings)`

### Fix plan 1

**F-1 — A read-only input speaks its value after the field's name** (from UX-1)
Files: `macos/App/DesignSystem/Components/Input/PVInput.swift`, `macos/App/DesignSystem/README.md` (Field row). Skill: `add-ui-component`.
Failing test first: none at unit level (SwiftUI accessibility tree); covered by manual check.
Do: in `PVInput`'s `isReadOnly` branch, add `.accessibilityValue(Text(text))` to the read-only `Text` so VoiceOver reads the field name (from `pvFieldNamed()`) then the value. README Field row: "a read-only input speaks the field name, then its value".
Check: build clean; scripts/check-pr.sh PASS.

**F-2 — One spoken join for the kit and every feature** (from MNT-1; same-pass cleanup)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Platform/AccessibilityAnnouncement.swift`, `macos/App/DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift`, the Conclusions callers (`ReconciliationRecord.swift`, `PersonsListView.swift`, `PersonGenderTreatment.swift`, `ReconciledValueDisplay.swift`, `ConclusionDetailPage.swift`, `ReconciledValueRowModel.swift`, `EventDetailView.swift`, `EventsListView.swift`), `macos/ProvenenciaTests/AccessibilityAnnouncementTests.swift`, plan Landed note. Skill: `add-localized-string`.
Failing test first: none (output unchanged); existing tests that read these labels must stay green.
Do: rename `L10n.DesignSystem.announcementJoin(_:_:)` to `L10n.DesignSystem.spokenJoin(_:_:)`, key `designSystem.spokenJoin`, value "%1$@, %2$@", comment "VoiceOver join of two spoken parts (a label and its detail, a title and its message)". Move every caller of `L10n.Conclusions.a11yList(_:rest:)` and `L10n.LastUpdated.a11yJoin(_:_:)` onto it. Delete those two functions, their resources, and the keys `conclusions.a11y.list`, `lastUpdated.a11y.join`, `designSystem.announcement.join` from the catalog. Landed note: supersedes D-6's separate kit join key (ship log MNT-1).
Check: `grep -rnE "a11yList|a11yJoin|announcementJoin|announcement\.join|a11y\.list|a11y\.join" macos .agents docs/*.md` prints nothing; `grep -rn '"%1\$@, %2\$@"' macos/App/Resources/Localizable.xcstrings` prints exactly one line; scripts/check-pr.sh PASS.

**F-3 — Follow D-12: a card switch or re-open does nothing while a save runs or on the same card** (from UX-4, CON-1)
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`, plan Landed note (remove deviation 1).
Failing test first: `SourcePageModelTests/cardSaveErrorStaysOnTheCardThatSaved` changed to: during the held save `beginEdit("P4")` → after release, `editingArtifactID == "P3"` and `cardSaveError` is the kept error (`L10n.Sources.artifactCardSaveErrorKept(…)`), `pendingEditSwitch == nil`. And a new `beginEditOnTheSameCardKeepsItsEdits` *(new)*: `beginEdit("P3")`, change `labels["P3"]`, `beginEdit("P3")` → the label is still the edited one.
Do: `beginEdit` per D-12 exactly: return when `editingArtifactID == id` or `savingID != nil`; ask when another card is in edit and dirty; else `openEdit`. `cardEditLocksEveryBlock`: where it calls `beginEdit("P9")` again to reload blocks, first end P9's edit with `cancelFields(id: "P9")` (or the section's existing close-edit path) so the reload is a fresh open; the test's assertions stay.
Check: both tests pass; scripts/check-pr.sh PASS.

**F-4 — Compound Value editors name their own inputs** (from UX-2)
Files: `macos/App/Features/Dates/DateValueEditorForm.swift`, `macos/App/Features/Names/NameValueEditorForm.swift`.
Failing test first: none (accessibility tree); manual check.
Do: on each editor's root view, `.environment(\.pvFieldLabel, nil)` so the outer "Value" field name does not reach its inputs. Then each input that has a visible sibling label (`cascadeField`'s `label`, the name-part rows' part label, the full-form and phrase inputs' label) gets `.environment(\.pvFieldLabel, <that label's string>)`, so VoiceOver reads "Year", "Day", "Phrase" and so on.
Check: build clean; scripts/check-pr.sh PASS.

**F-5 — `WorkspaceLocationLabels` owns the graph-name and graph-surface rules** (from MNT-2, MNT-3)
Files: `macos/App/Features/Workspace/WorkspaceLocationLabels.swift`, `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/Workspace/WorkspaceLocation.swift`.
Failing test first: none (refactor); `WorkspaceToolbarBreadcrumbTests` and `EvidenceGraphModelTests.canvasLabelNamesTheGraph` stay green.
Do: add `WorkspaceLocationLabels.graphName(graphID:names:) -> String?`; `graphLabel` and `EvidenceGraphModel.canvasAccessibilityLabel` both call it. Add `WorkspaceLocation.graphSurfaceID: String?` *(new)* (the `graphId` when `sourceSurface` is `.graph`, `.citationComposer` or `.promote`); `breadcrumbItems` and `labels(for:names:)` use it. Rename the trimming `nilIfEmpty` extension in `WorkspaceLocationLabels.swift` to `trimmedNonEmpty` and make it `fileprivate` if `WorkspaceToolbar` no longer needs it, else internal under the new name.
Check: `grep -rn "nilIfEmpty" macos/App/Features/Workspace/WorkspaceLocationLabels.swift macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` prints nothing; scripts/check-pr.sh PASS.

**F-6 — Kind-noun `defaultValue`s match the catalog** (from MNT-4)
Files: `macos/App/Platform/L10n.swift`, plan Landed note.
Do: every `kindNouns` resource and `itemNoun` take `defaultValue: "%#@noun@"` (D-19); `viaFallback`'s resource takes `defaultValue: "%1$lld %2$@ %#@verb@ this"` (its catalog value).
Check: `PVDeleteImpactTests`, `L10nPluralTests` pass; scripts/check-pr.sh PASS.

**F-7 — Discard strings' comments cover every use** (from UX-3)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: the four `designSystem.formDialog.discard*` comments (both places) say the confirm shows when a form sheet or an Artifact card with unsaved edits is about to drop them (outside click, Escape, a row that navigates, another card's edit). Copy values unchanged.
Check: scripts/check-pr.sh PASS.

**F-8 — The sheet's ask-then-run rule is a tested pure function** (from CON-3)
Files: `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/ProvenenciaTests/PVDiscardGateTests.swift`. Skill: `add-swift-test`.
Failing test first: `PVDiscardGateTests/sheetCloseOutcomes` *(new)*: for each `PVSheetDismiss` outcome (running, asking, dirty, clean) the pure function returns the right action, and on dirty the step is held in the gate; Keep editing drops it; Discard runs it after close.
Do: move the `PVSheetChrome.run` switch on `PVSheetDismiss` into an `internal` static function beside `PVSheetChrome` (taking the dismiss state and the gate, returning what to do), and have `run`, `discard` and `keepEditing` call it. Behavior unchanged.
Check: the test passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `33b8ff62`. Implementer gate PASS; orchestrator gate PASS. F-1 – F-8 done.
Deviations: F-2's grep check was wrong (orchestrator's): 7 other keys share `"%1$@, %2$@"` (`workspace.{persons,events,places}.rowAccessibility`, `citationComposer.depictionLineLabel`, `citationComposer.bridgeEndsReading`, `sources.list.graphBlockLabel`, `sources.list.graphButtonLabelNoStatus`). Each is a purpose-named label with its own translator comment (what its two arguments are), not a generic join, and this PR does not touch them; left as is. F-1's commit also carries F-2's `PVLastUpdated.swift` line (no amend).
Resolved without re-review (read by the orchestrator): F-7 (UX-3 comments), F-8 (CON-3, `PVSheetCloseRule.outcome` + `sheetCloseOutcomes`) — both go back with their reviewers anyway.

## Round 2 — review
Reviewers: MNT (MNT-1), UX (UX-1, UX-4), CON (CON-1, deduped into UX-4 and re-rated major). Skipped: SEC+PRF (no Fix-now finding). Scope `28af621a...33b8ff62`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Spoken-join key | Resolved | one `designSystem.spokenJoin`; grep over old names empty |
| UX-1 | UX | major | Read-only input value | Resolved | `PVInput.swift:139-140` `accessibilityValue(Text(text))` |
| UX-4, CON-1 | UX, CON | major | Card switch during a save | Resolved | `beginEdit` matches D-12; `cardSaveErrorStaysOnTheCardThatSaved`, `beginEditOnTheSameCardKeepsItsEdits` |
| CON-3 | CON | minor | Sheet close rule untested | Resolved | `PVSheetCloseRule.outcome`, `PVDiscardGateTests.sheetCloseOutcomes` |
| UX-2, MNT-2, MNT-3, MNT-4, UX-3 (comments) | — | minor | F-4 – F-7 | Resolved | read in the fix commits by the round-2 reviewers and the orchestrator |

No new blocker. Nits (logged): `MNT-9 · PVSheet.swift:99 · PVSheetCloseRule.Outcome mirrors PVSheetDismiss's cases` · `UX-7 · NameValueEditorForm.swift:182 · every name-part input speaks "Part value" with no position`

## Finish
Merged `origin/main` (`54d1e20a`, docs only: the S14-18 plan) as `ed6bbd1e`; gate PASS against `54d1e20a`.

## Result
Gate PASS · CI pending · open: none · follow-ups: CON-2 (committed test for the planted-key run of `check-localizable-xcstrings.py`; needs a first Python test home), UX-3 (name the Artifact in the card's "Discard changes?" message) · to user: — · CON miss: none (CON found CON-1 but rated it minor; triage re-rated it major for setting aside D-12)
