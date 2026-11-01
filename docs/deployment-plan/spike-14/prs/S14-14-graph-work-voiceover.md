# S14-14 — Copy and VoiceOver on graph work

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: no layout or component changes; copy is rewritten in place, and the one moved error line uses the kit's existing `PVErrorText` under the Citation picker (user, 2026-11-01); `PVButtonStyle` draws no default-button tint, so moving Return to Cancel draws nothing different)
**Board:** —
**Depends on:** —
**Blocks:** S14-20 (reads `PVCallout`'s `announces:`)
**Branch:** `s14-14-graph-work-voiceover`
**Version bump:** patch, 0.2.5 → 0.2.6 — accessibility, copy and safety fixes to existing screens, no feature ([`docs/versioning.md`](../../../versioning.md)). If `main` is past 0.2.5 when the PR merges, take the next patch.
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** audit [A-12](../app-health-audit.md#a-12--nothing-says-which-evidence-graph-you-are-on), [A-13](../app-health-audit.md#a-13--a-kept-linked-citation-link-discards-a-dirty-graph-form), [A-14](../app-health-audit.md#a-14--copy-says-this-source-where-the-action-is-one-graph), [A-17](../app-health-audit.md#a-17--return-confirms-destructive-deletes), [A-18](../app-health-audit.md#a-18--voiceover-is-not-told-graph-write-results-or-failures), [A-20](../app-health-audit.md#a-20--counts-without-plural-rules), [A-21](../app-health-audit.md#a-21--section-headers-are-not-voiceover-headings), audit 2 [B-18](../app-health-audit-2.md#b-18--the-sheets-status-confidence-and-argument-fields-have-no-spoken-name); working rules "Kit components announce themselves", "A danger confirm's default action is Cancel", "A dirty form asks before a link inside it navigates away", "No graph name in the canvas header", "Audit items keep their IDs"; bar outcome 17; the S14-14 scope entry's three Spike 12 ship-log leftovers
**Planned:** 2026-11-01 against `main` @ `75d9494`
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

Keyboard and VoiceOver are safe on graph work. Return in a delete confirm cancels. A dirty graph form, or a dirty Artifact card, asks "Discard changes?" before a link or another card takes it away. VoiceOver hears every toast and the error callouts on graph work, can jump between page sections by heading, hears each form field's name, and hears which Evidence graph is open on the canvas and in the Back / Forward menus. Counts read "1 subject", not "1 subjects", and a script keeps new counts honest. Three Spike 12 ship-log leftovers come with it: a failed Citation load names the Citation under the Citation picker, a file macOS fails to open says so, and the Source type editor hands focus back to the Type control when it closes. The work is audit items A-12, A-13, A-14, A-17, A-18, A-20, A-21 and B-18, which Spike 12 planned as S12-29, S12-30 and S12-45. The map said to split the work along those parts if the plan ran past ten steps. The user chose one PR anyway (2026-11-01).

**Why now.** Every slice 5 audit fix on graph work lands on this kit behavior. S14-20 opts the Promote page's errors into the callout announcement this PR adds.

**What stays.** No page's layout, and no component's look, changes. The canvas header names no graph (working rule "No graph name in the canvas header"). Toast timing stays at 5 s. An irreversible-tone confirm keeps Return on its confirm button. Go is untouched.

## 2. User stories

**US-1.** As a keyboard user, Return in a delete confirm keeps the record, so a reflex press loses nothing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `PVConfirmTone.danger` | `returnConfirms == false`: Cancel carries `.keyboardShortcut(.defaultAction)`, the confirm button has no keyboard shortcut |
| AC-1.2 | `PVConfirmTone.irreversible` | `returnConfirms == true`: the confirm button carries `.defaultAction`, as today |
| AC-1.3 | Delete Evidence graph confirm open, Return pressed (manual) | The sheet closes; the graph is still listed |
| AC-1.4 | A dirty graph form's "Discard changes?" question, Return pressed (manual) | The question closes and the form keeps its typed name |

**US-2.** As a researcher with unsaved edits, a link or another card asks before it drops them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Graph Edit form, name changed from "Jones household" to "Jones family", a Kept-linked Citation row `CIT-K4M2P` pressed | The kit's "Discard changes?" question shows; no navigation |
| AC-2.2 | AC-2.1, then Discard | The sheet closes, then the workspace goes to `CIT-K4M2P`'s location |
| AC-2.3 | AC-2.1, then Keep editing | The question closes; the form shows "Jones family"; no navigation |
| AC-2.4 | Graph Edit form not dirty, row pressed | The sheet closes and navigates, as today |
| AC-2.5 | Graph form saving (`isSaving`) | The Kept-linked rows are disabled; a press does nothing |
| AC-2.6 | `PVSheetPendingClose`: `request(isRunning: false, isDirty: true, then: f)` → `.confirmDiscard`; `discardConfirmed()` | returns `f` once; a second call returns nil |
| AC-2.7 | `PVSheetPendingClose`: `request(isRunning: false, isDirty: true, then: f)`, then `keepEditing()`, then `discardConfirmed()` | `discardConfirmed()` returns nil: Keep editing drops the pending navigation |
| AC-2.8 | Artifact `ART-1` in edit, label changed from "Page 1" to "Page one", pencil on `ART-2` pressed | `pendingEditSwitch?.targetID == "ART-2"`; `editingArtifactID == "ART-1"`; the label buffer still reads "Page one" |
| AC-2.9 | AC-2.8, then `confirmEditSwitch()` | `editingArtifactID == "ART-2"`; `pendingEditSwitch == nil`; `ART-1`'s buffer is back to "Page 1" |
| AC-2.10 | AC-2.8, then `cancelEditSwitch()` | `pendingEditSwitch == nil`; `editingArtifactID == "ART-1"`; buffer "Page one" |
| AC-2.11 | `ART-1` in edit and not dirty, **Link to a graph** on `ART-2` pressed | `editingArtifactID == "ART-2"` at once; no question |

**US-3.** As a VoiceOver user, I hear graph write results and failures, and can move by heading.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `PVAnnouncement.text(title: "Evidence graph added", message: nil)` | `"Evidence graph added"` |
| AC-3.2 | `PVAnnouncement.text(title: "Couldn't save the evidence graph", message: "The name is taken.")` | `"Couldn't save the evidence graph, The name is taken."` |
| AC-3.3 | `PVAnnouncement.text(title: nil, message: "")` | `nil` (nothing is posted) |
| AC-3.4 | `PVAnnouncement.priority(danger: true)` / `(danger: false)` | `.high` / `.medium` |
| AC-3.5 | Graph added on the Source page, VoiceOver on (manual) | VoiceOver says "Evidence graph added" once |
| AC-3.6 | Graph form save fails, VoiceOver on (manual) | VoiceOver says the callout's title and message once, focus still on Save |
| AC-3.7 | Source page, VoiceOver rotor → Headings (manual) | "Evidence graphs" and every other section title are listed; VO-Cmd-H moves between them |
| AC-3.8 | Set a portrait on a Person page, VoiceOver on (manual) | The toast is spoken once, not twice (the model's own post is gone) |

**US-4.** As a VoiceOver user, every form field says its name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `PVFieldLabel.resolve(own: nil, field: "Name")` | `"Name"` |
| AC-4.2 | `PVFieldLabel.resolve(own: "Filter", field: "Name")` | `"Filter"` (a call site's own label wins) |
| AC-4.3 | `PVFieldLabel.resolve(own: nil, field: nil)` | `nil` (the control keeps what it says today) |
| AC-4.4 | Graph form, VoiceOver on the description text area (manual) | VoiceOver says "Description, …", not only the text |
| AC-4.5 | Composer, VoiceOver on the citation label field (manual) | VoiceOver says the field's label once (the hand-made label is gone) |

**US-5.** As a researcher moving between graphs, the canvas and the history menus name the graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `locationLabels` for `.evidenceGraph(sourceId: "S1", graphId: "G1", ref: "SRC-1", title: "1881 Census")`, names `G1 → "Jones household"` | leaf `"Jones household"`, section `nil` |
| AC-5.2 | The same with names knowing no graph | leaf `"Evidence graph"` |
| AC-5.3 | `.citationComposer(sourceId: "S1", graphId: "G1", ref: "CIT-K4M2P", title: "CIT-K4M2P")`, names `G1 → "Jones household"` | leaf `"Jones household"`, section `"Citation for CIT-K4M2P"`; the Back menu row reads `"Jones household · Citation for CIT-K4M2P"` |
| AC-5.4 | A Promote location for `PER-7H2Q` on `G1` | leaf `"Jones household"`, section `"Promote PER-7H2Q"` |
| AC-5.5 | `.sourcePage(sourceId: "S1", title: "1881 Census")` | leaf `"1881 Census"`, section `nil` (unchanged) |
| AC-5.6 | `breadcrumbItems` for AC-5.1 and AC-5.3 | Labels unchanged from today: `["Sources", "1881 Census", "Jones household"]` and `[…, "Jones household", "Citation for CIT-K4M2P"]` |
| AC-5.7 | `EvidenceGraphModel.canvasAccessibilityLabel` with graph name "Jones household", nothing armed | `"Evidence graph, Jones household"`; with no name known, `"Evidence graph"` |

**US-6.** As a researcher, copy says what the action covers and counts take plurals.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | `L10n.Promote.promoting(count: 1)` / `(count: 3)` | `"1 subject on this graph"` / `"3 subjects on this graph"` |
| AC-6.2 | `L10n.CitationComposer.pageCount(count: 1)`, `.artifactCitationCount(count: 1)`, `.citationStepperCount(count: 1)` | `"1 page"`, `"1 citation"`, `"1 observation"`; with 2, the plural |
| AC-6.3 | `L10n.DeleteImpact.noun("citation", count: 1)` / `(…, count: 2)` | `"citation"` / `"citations"`, read from one plural key |
| AC-6.4 | `L10n.DeleteImpact.viaFallback(kind: "file", total: 1)` / `(…, total: 3)` | `"1 file references this"` / `"3 files reference this"` |
| AC-6.5 | A catalog copy with a planted key `"test.planted": "%lld subjects"` (no variations), script run on it | `check-localizable-xcstrings.py` exits 1 and names `test.planted` |
| AC-6.6 | The shipped catalog | the script passes |
| AC-6.7 | `L10n.Promote.mapRestHint`, `graphUnavailable`, `graphUnavailableRepropose`, `L10n.EvidenceGraph.addPropertyUnavailable`, `L10n.CitationComposer.pairPickBoth` | each names the Evidence graph and none says "this Source" (§ 6 Strings) |

**US-7.** As a researcher, an error names what failed and focus lands where I was.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-7.1 | Composer, switching to listed Citation `CIT-K4M2P` whose load throws `catalog.not_found` | `citationLoadError == "Couldn't open CIT-K4M2P. <message for catalog.not_found>"`; `fields.error == nil` |
| AC-7.2 | AC-7.1, then a successful load of another Citation | `citationLoadError == nil` |
| AC-7.3 | A load failure for a Citation id not in `listedCitations` | `citationLoadError` starts "Couldn't open this citation." |
| AC-7.4 | AC-7.1 with a save error already in `fields.error` | `fields.error` keeps the save error |
| AC-7.5 | `ProjectFiles.openObject` for a stored PDF whose `launch` throws | returns `.notOpened` |
| AC-7.6 | The canvas opens a file and macOS fails | `fileOpenError == L10n.ObjectOpen.notOpened(filename: "scan.pdf")` |
| AC-7.7 | The Source page opens an Artifact's file and macOS fails | `pageError == L10n.ObjectOpen.notOpened(filename: …)` |
| AC-7.8 | Source page, Type pencil, then the × or a picked type (manual, Full Keyboard Access on) | Focus is on the Type pencil button |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-17: danger confirm's default is Cancel; README line | STEP-1, AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| A-18: `PVToast` announces; `PVCallout` and `PVErrorText` take `announces:`; the graph form, Artifact card, Source page and canvas file-open callouts opt in; `ConclusionLeadModel`'s own posts go | STEP-2, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.8 |
| A-21: `PVSectionHeader` title is a heading | STEP-3, AC-3.7 |
| B-18: `PVField` gives its label to the kit control inside it; the composer's hand-made label goes | STEP-4, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5 |
| A-13: Kept-linked rows close through the sheet's dirty rule, then navigate; disabled while saving | STEP-5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7 |
| Leftover: an Artifact card's pencil or **Link to a graph** asks before dropping another card's unsaved edit | STEP-6, AC-2.8, AC-2.9, AC-2.10, AC-2.11 |
| A-12: one location-label function for crumbs, history rows and the canvas VoiceOver label | STEP-7, AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, AC-5.6, AC-5.7 |
| A-14: five strings name the Evidence graph | STEP-8, AC-6.7 |
| A-20: plural variations for four keys; `DeleteImpact.noun` and the unknown-via heading read plural keys; the script fails on an uninflected count | STEP-9, AC-6.1, AC-6.2, AC-6.3, AC-6.4, AC-6.5, AC-6.6 |
| Leftover: `ProjectFiles.openObject` reports a failed launch | STEP-10, AC-7.5, AC-7.6, AC-7.7 |
| Leftover: a failed Citation load names the Citation under the Citation picker | STEP-11, AC-7.1, AC-7.2, AC-7.3, AC-7.4 |
| Spike 11 leftover: focus returns to the Type control after the Source type editor | STEP-12, AC-7.8 |

| Out | Home |
| --- | --- |
| A graph name in the canvas header | Refused (working rule "No graph name in the canvas header") |
| Tooltips on truncated names | Not planned (scope entry Not this PR) |
| Holding toasts longer while VoiceOver runs | Not planned (scope entry Not this PR) |
| Promote page errors scrolled into view and announced | S14-20 |
| "On this Source" (`promote.page.column.source`) and "Promoted from this Source earlier" (`promote.row.promotedEarlier`): A-14 does not list them, and both are true of the Source | Not planned |
| Speaking a field's required or optional marker | Not planned |
| A kit control that is not a `PVInput`, `PVTextArea`, `PVSelect` or `PVComboBox` inside a `PVField` (the drop row, the checklist) learning the field label | Not planned; those rows already speak their own text |
| Announcing the composer's other error paths (`savePage`'s own post) | Not planned; not a toast or callout |

## 4. Current state

Verified at `75d9494`.

| Area | Where | Today |
| --- | --- | --- |
| Confirm | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` `PVConfirmContent` | Confirm button `.keyboardShortcut(.defaultAction)` for every tone; `.onAppear { cancelFocused = true }` with the comment "Return must not complete a destructive action" — but `.defaultAction` fires on Return whatever has focus (A-17). `PVConfirmTone` has `buttonRole`, `buttonVariant`, `confirmIcon` |
| Discard question | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `PVSheetChrome` | Private `confirmingDiscard` state; builds the discard `PVConfirmCopy` inline from `L10n.DesignSystem.formDialogDiscard*`; `close` is the binding clear; no way for content to ask "close, then run this" |
| Dismiss rule | `macos/App/DesignSystem/Components/Panel/PVSheetDismiss.swift` `PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard:)` | `.ignore` / `.keepEditing` / `.confirmDiscard` / `.close` |
| Kept-linked rows | `macos/App/Features/Sources/SourceGraphFormDialog.swift` `SourceGraphFormDialogModifier.body`, `SourceGraphForm.keptLinked` | `dismiss: { model = nil }` skips the dirty check (A-13); rows not disabled while saving |
| Row activation | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactControls.activate(_:onNavigate:dismiss:then:)`, `PVDeleteImpactGroupList` (`dismiss` param), `PVDeleteImpactNotice` (passes `onDismiss`) | `dismiss()` then navigates in a `Task`; `then:` has no caller |
| Env precedent | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `EnvironmentValues.pvContextMenuDismiss`; `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` `PVWidthClassKey` | The kit's shape for an environment action and value |
| Toast | `macos/App/DesignSystem/Components/Toast/PVToast.swift` `PVToast` | `.accessibilityElement(children: .combine)`; no announcement (A-18). Callers: `VocabularyToastOverlay` (`macos/App/Features/CatalogVocabulary/VocabularyChrome.swift`), `OnboardingView`, `CitationComposerSubjectGroups` |
| Own posts | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` (two `AccessibilityAnnouncement.post` beside each `toast =`) | Posts what its toast shows; with a kit announcement it would speak twice |
| Callout | `macos/App/DesignSystem/Components/Callout/PVCallout.swift` `PVCallout` | No announcement |
| ErrorText | `macos/App/DesignSystem/Components/ErrorText/PVErrorText.swift` `PVErrorText` | No announcement |
| Announce | `macos/App/Platform/AccessibilityAnnouncement.swift` `AccessibilityAnnouncement.post(_:priority:)` | The one post path |
| Error callouts | `SourceGraphFormDialog.swift` (`sources.graphForm.error`); `macos/App/Features/Sources/SourcePageArtifactsView.swift` `editFields` (`sources.page.artifact.<id>.saveError`); `macos/App/Features/Sources/SourcePageView.swift` (`sources.page.error`, `pageError`); `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (`evidenceGraph.fileOpen.callout`) | Silent when they appear |
| Section header | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` | Title not a heading (A-21). Precedent: `ConclusionDetailPage.swift` `.accessibilityAddTraits(.isHeader)` |
| Field | `macos/App/DesignSystem/Components/Field/PVField.swift` | Label is a sibling `Text`; content gets no name (B-18). 40 call sites under `macos/App` |
| Field controls | `macos/App/DesignSystem/Components/Input/PVInput.swift` (`TextField(text:prompt:) { EmptyView() }`), `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` (`TextEditor`), `macos/App/DesignSystem/Components/Select/PVSelect.swift` `accessibilityLabelText` (explicit → spoken → trigger text), `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` (`label:` required, already spoken) | Input and text area have no spoken name; a select inside a field speaks its trigger text |
| Hand-made label | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` `labelField` | `.accessibilityLabel(Text(L10n.CitationComposer.descriptionLabel))` on the `PVInput` |
| Card edit | `macos/App/Features/Sources/SourceArtifactsSection.swift` `beginEdit(_:)`, `fieldsDirty(_:)`, `cancelFields(id:)` | `beginEdit` cancels another card's edit with no question; called by the pencil and **Link to a graph** in `SourcePageArtifactsView.metaRow` / `noGraphRow` |
| Crumbs | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `BreadcrumbNames`, `breadcrumbItems(for:names:goTo:)`, `locationLabels(for:)` (`nonisolated`) | The crumbs read the graph name from the session; `locationLabels` does not, so history rows show the Source's title and ref for the Source page and every canvas alike (A-12) |
| History rows | `WorkspaceToolbar.swift` `HistoryJumpMenuHost.rows(for:navigation:)`, private `HistoryJumpRow.leafLabel` | Calls `locationLabels(for:)`; no session access |
| Canvas label | `EvidenceGraphView.swift` `accessibilityGraphLabel` | Nothing armed → `L10n.Workspace.evidenceGraphTitle` ("Evidence graph") |
| A-14 copy | `macos/App/Platform/L10n.swift` `L10n.Promote.mapRestHint`, `.graphUnavailable`, `.graphUnavailableRepropose`; `L10n.EvidenceGraph` `addPropertyUnavailable` (`evidenceGraph.subject.addPropertyUnavailable`); `L10n.CitationComposer.pairPickBoth` | Say "this Source" |
| A-20 copy | `L10n.swift` `promote.page.promoting`, `citationComposer.pageCount`, `.artifactCitationCount`, `.citationStepperCount` | No variations. The Promote toast and row-summary keys gained variations in S13-23 |
| Noun pick | `L10n.swift` `L10n.DeleteImpact.noun(_:count:)` (`count == 1 ? pair.0 : pair.1` over 19 `.one` / `.other` key pairs), `viaFallback(kind:total:)` (`total == 1 ? viaUnknownOne : viaUnknownOther`) | English-only plural choice in Swift |
| Plural test | `macos/ProvenenciaTests/L10nPluralTests.swift` `everyCatalogPluralPicksAFormPerCount` | Every compiled plural must differ for 1 and 2 and contain the digit |
| Catalog script | `scripts/check-localizable-xcstrings.py` `main` | No count check; reads the fixed `CATALOG` path |
| Open | `macos/App/Platform/ProjectFiles.swift` `openObject(projectDir:relPath:)`, `ObjectRefusal` (`.missing`, `.unsafe`) | Ignores `NSWorkspace.open`'s result. Callers: `EvidenceGraphModel.openSelectedFile()`, `SourceArtifactsSection.open(_:)` |
| Load failure | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `showingLoadFailure(_:)`, `loadFailureMessage` | Puts the error in `fields.error`, which draws under the Transcription field (`CitationComposerEvidenceColumn.transcriptionColumn`), and names no Citation. Callers: `refreshCitationLines`, `selectCitationAndLoad`, `loadCitationPreservingAuthoring`, the first-Citation load after an Artifact change. FakeStore's `getCitation` throws `StoreBoom.boom` for an id missing from `citationsByID`; `lineRefreshShowsAFailedCitationLoad` reads `fields.error` |
| Citation picker | `CitationComposerEvidenceColumn.swift` `citationStepper` | The stepper both layouts share |
| Type editor | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `TypeField` (`case combo`), `metaRow`; `SourceIdentitySection.editingType` | Focus goes to the combo on open; nothing takes it on close. Precedent: `PromoteView.swift` `PromoteAlignmentRow` `.onChange(of: model.recordSearch?.subjectID)` → `targetFocus = true` |
| Tests | `PVConfirmTests`, `PVDeleteImpactTests` (`activate` test near line 207), `WorkspaceToolbarBreadcrumbTests` (`@MainActor`; `locationLabelsNameTheSection`), `L10nPluralTests`, `ProjectFilesTests` (`openObjectReports…`), `EvidenceGraphModelTests` (`openSelectedFile…`), `SourcePageModelTests`, `CitationComposerModelTests` | Cover today's behavior |

**Precedent.** S14-03 ([#652](https://github.com/mendahu/provenencia/pull/652)) moved hand-styled error lines onto `PVErrorText` and announced a failed save from its Callout: copy that "kit owns the look, one announcement per message" shape. `pvContextMenuDismiss` is the shape for `pvSheetClose`. `PromoteAlignmentRow`'s `targetFocus` hand-back is the shape for STEP-12.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVConfirmTone.returnConfirms: Bool` *(new)*: `.danger` → false, `.irreversible` → true. `PVConfirmContent` puts `.keyboardShortcut(.defaultAction)` on Cancel when `!returnConfirms`, on confirm otherwise. Cancel keeps `cancelFocused`. The comment says Return cancels | Working rule "A danger confirm's default action is Cancel"; one pure flag a test reads | Dropping the shortcut from both buttons — Return then does nothing, and the rule says it cancels |
| D-2 | `PVConfirmCopy.discardChanges` *(new, static)* is the kit's discard copy (`L10n.DesignSystem.formDialogDiscard*`). `PVSheetChrome` and the Artifact card switch both use it | One copy for one question | A second set of strings for the card |
| D-3 | `PVSheetPendingClose` *(new struct, `PVSheet.swift`)*: `mutating func request(isRunning:isDirty:then:) -> PVSheetDismiss` runs `PVSheetDismiss.action` and holds `then` when the answer is `.confirmDiscard`; `mutating func discardConfirmed() -> (@MainActor () -> Void)?` hands it out once; `mutating func keepEditing()` drops it. `PVSheetChrome` keeps one in `@State` in place of `confirmingDiscard` (§ 6 macOS) and publishes `EnvironmentValues.pvSheetClose: PVSheetClose` *(new)*, a struct with `func callAsFunction(then:)`. On `.close` it calls `close()` and runs `then` in a `Task { @MainActor in }`, as `activate` does today. Escape and the outside click call it with an empty `then` | The dirty rule lives in the sheet; content asks the sheet instead of clearing the binding, so every row in every sheet gains it | Passing `isDirty` down to the rows — a second copy of the rule; calling `activate(…, then:)` — it still dismisses first |
| D-4 | `PVDeleteImpactControls.activate(_:onNavigate:close:)` takes `close: PVSheetClose`. `PVDeleteImpactGroupList` drops its `dismiss` parameter and reads `@Environment(\.pvSheetClose)`. The unused `then:` parameter goes. `PVDeleteImpactNotice` keeps `onDismiss` for its Done button | Both callers are in a `pvSheet`, so both close through the same rule; the notice is never dirty, so it closes at once as today | Keeping `dismiss:` beside the environment — two ways to close |
| D-5 | `SourceGraphForm.keptLinked` gets `.disabled(model.isSaving)` | A-13; the sheet also ignores a close while running (`PVSheetDismiss.ignore`) | — |
| D-6 | `PVAnnouncement` *(new enum, `macos/App/Platform/AccessibilityAnnouncement.swift`)*: `static func text(title: String?, message: String?) -> String?` joins non-empty parts with ", " (nil when both are empty); `static func priority(danger: Bool) -> NSAccessibilityPriorityLevel` (`.high` / `.medium`). `View.pvAnnounce(_ text: String?, priority:)` *(new modifier)* posts through `AccessibilityAnnouncement.post` on appear and on change to a new non-nil text | One rule for every kit announcement; testable without VoiceOver | Each component calling `post` in its own `onAppear` |
| D-7 | `PVToast` always applies `pvAnnounce(PVAnnouncement.text(title:message:), priority: .priority(danger: tone == .danger))`. `PVCallout` and `PVErrorText` take `announces: Bool = false` *(new)* and apply the same modifier when true (danger tone → high; `PVErrorText` → high). `ConclusionLeadModel` drops its two `post` calls | Working rule "Kit components announce themselves": no call site posts its own. A toast always announces (A-18 fix); a callout opts in because most callouts are static explanations | Announcing every callout — a page's info callouts would talk on every open |
| D-8 | Callouts that opt in: `sources.graphForm.error`, `sources.page.artifact.<id>.saveError`, `sources.page.error`, `evidenceGraph.fileOpen.callout`. The composer's new load error line (D-17) opts in on `PVErrorText` | The scope entry's list; each appears after an action with focus elsewhere | The Source page `loadError` callout — it replaces the page, so focus is already there |
| D-9 | `PVSectionHeader` puts `.accessibilityAddTraits(.isHeader)` on `title.pvText` | A-21; precedent `ConclusionDetailPage.swift` | Per call site |
| D-10 | `EnvironmentValues.pvFieldLabel: LocalizedStringResource?` *(new, `PVField.swift`)*, default nil. `PVField` sets it on `content`. `PVFieldLabel.resolve(own: Text?, field: Text?) -> Text?` *(new)* is own → field → nil (the pure rule tests read as Strings: see T-10). `PVInput` and `PVTextArea` apply `.accessibilityLabel` from the resolved field label when it is non-nil. `PVSelect.accessibilityLabelText` becomes explicit → spoken → field → trigger text. `PVComboBox` is unchanged: its `label:` is required and already spoken. The label `Text` in `PVField` stays visible to VoiceOver | Fixes every form in the kit (B-18's "better" fix); a call site's own label still wins | `accessibilityLabeledPair` — on macOS its effect when the content has its own label is not testable here, and the select's own label would still win; hiding the field's label `Text` — a field whose content is not a kit control would lose its only name |
| D-11 | `CitationComposerEvidenceColumn.labelField` drops its `.accessibilityLabel` | Duplicate of D-10 | — |
| D-12 | `SourceArtifactsSection.requestEdit(_ id:)` *(new)*: when another card is in edit and `fieldsDirty(current)`, set `pendingEditSwitch = PendingEditSwitch(targetID: id)` *(new, `Identifiable`)*; else `beginEdit(id)`. `confirmEditSwitch()` *(new)* runs `beginEdit(target)` and clears it; `cancelEditSwitch()` *(new)* clears it. The pencil and **Link to a graph** call `requestEdit`. `SourcePageView` adds `.pvConfirm(item: pendingEditSwitch, copy: { _ in .discardChanges }, onConfirm: confirmEditSwitch)` with prefix `sources.page.artifact.discard` | The scope entry extends the dirty-form rule to the card; `beginEdit` keeps its one job | Asking inside `beginEdit` — it is also the reload path |
| D-13 | `WorkspaceToolbar.BreadcrumbNames` is renamed `LocationNames` *(rename)*. `locationLabels(for:names:)` (names default `.none`) loses `nonisolated` and is main-actor. For `.graph`, `.citationComposer` and `.promote` locations with a `graphId` it returns leaf = `graphLabel(graphID:names:)` *(new)* (`names.graphName` → `L10n.Workspace.evidenceGraphTitle`), section = the composer or Promote crumb label (`nil` on the canvas). `breadcrumbItems` builds its graph, composer and Promote crumbs from the same two helpers, so the crumbs do not change | One function for both (scope entry); the crumbs' labels stay byte-for-byte | A history-only lookup — a second name rule |
| D-14 | `HistoryJumpRow` stores `leafLabel` computed in `HistoryJumpMenuHost.rows(for:navigation:names:)` (main actor) from `locationLabels`, joined with `L10n.Conclusions.pageSectionLeaf` ("%1$@ · %2$@") as a page section is today. `HistoryJumpMenuHost` reads `@Environment(WorkspaceSession.self)` to build `LocationNames`. The ref column stays the location's `ref` | User, 2026-11-01: graph name, then surface; the existing Back-menu leaf format | Source title · graph name (user declined: longer) |
| D-15 | `EvidenceGraphModel.canvasAccessibilityLabel` *(new)*: armed hints as today; else `L10n.EvidenceGraph.canvasAccessibility(graph:)` *(new)* = "Evidence graph, %@" with the graph's name from `session.evidenceGraph(id:)?.name` through `WorkspaceToolbar.graphLabel`; with no name, `L10n.Workspace.evidenceGraphTitle`. The view's `accessibilityGraphLabel` reads it | A-12's fix; testable on the model | Computing it in the view — not testable |
| D-16 | `L10n.DeleteImpact.noun(_:count:)` resolves one plural key per kind, `deleteImpact.kind.<kind>` *(new)*, via `L10n.format(_, count)`; each key's `one` / `other` variation is the bare noun ("citation" / "citations"), with no number. The 38 `.one` / `.other` keys and their 38 `static let`s go. `viaFallback` reads one plural key `deleteImpact.via.unknown` *(new)* (`%1$lld %2$@ references this` / `…reference this`); `viaUnknownOne` / `viaUnknownOther` go. `L10nPluralTests.everyCatalogPluralPicksAFormPerCount` skips the digit checks for keys starting `deleteImpact.kind.` and keeps "one ≠ other" | A-20: the plural rule is the catalog's, so a translation with more forms works; a stringsdict variation may omit its number | Counted noun keys ("%lld citation") — `alsoAffects` and the confirm title need the bare noun |
| D-17 | `CitationComposerModel.citationLoadError: String?` *(new)* replaces `loadFailureMessage`. `showingLoadFailure(citationID:_:)` *(new parameter)* sets it on failure to `L10n.CitationComposer.citationLoadFailed(ref:reason:)` *(new)* when `listedCitations` has the id, else `citationLoadFailedUnnamed(reason:)` *(new)*, and clears it on success. It no longer reads or writes `fields.error`. `citationStepper` shows `PVErrorText(citationLoadError, announces: true)` under the stepper | User, 2026-11-01: under the Citation picker, no brief | A callout above the steppers (user declined) |
| D-18 | `ObjectRefusal.notOpened` *(new case)*. `openObject(projectDir:relPath:launch:)` becomes `async` with `launch: (URL, URL) async throws -> Void = ProjectFiles.launch` *(new static, the `NSWorkspace.shared.open(_:withApplicationAt:configuration:)` async call)*; a throw returns `.notOpened`. `EvidenceGraphModel.openSelectedFile()` and `SourceArtifactsSection.open(_:)` become `async`, map `.notOpened` to `L10n.ObjectOpen.notOpened(filename:)` *(new)*, and their buttons call them in a `Task`. A model test injects a throwing launch through a `launch` parameter on each model method, defaulted | The open failure surfaces in the callouts that already show refusals | Logging only — the researcher sees nothing |
| D-19 | `SourcePageIdentityHeader.TypeField` gains `case edit`. The Type pencil gets `.focused($typeFieldFocused, equals: .edit)`; `.onChange(of: model.identity.editingType)` true → false sets `typeFieldFocused = .edit` in `DispatchQueue.main.async` | Precedent `PromoteAlignmentRow` | Focusing the pill — it is not a control |
| D-20 | `check-localizable-xcstrings.py` gains `uninflected_counts(strings) -> list[str]` *(new)*: an English value with `%lld` or `%N$lld` followed by a word of letters, in an entry with no `variations` or `substitutions`, is a problem unless the word is in `NON_INFLECTING = {"of", "more", "percent", "set", "up", "down", "cannot", "to", "unselected", "already", "px"}`. The script takes an optional catalog path (`argv[1]`, default `CATALOG`) | A-20's script check; the shipped catalog's 20 non-plural `%lld` words were each read (§ 6 Edges) | A key allow-list — every new "of" format would need an entry |

**Rules this plan bends.** The map scope entry says the plan splits along its parts past about ten steps; this plan has 13. User, 2026-11-01 ("do it all in one PR"). STEP-13 records it in the scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVConfirmTone.returnConfirms` *(new)* | property | `PVConfirmTone.buttonRole` | Extends the same enum |
| `PVConfirmCopy.discardChanges` *(new)* | static | inline copy in `PVSheetChrome` | Lifts it so the card reuses it |
| `PVSheetPendingClose`, `PVSheetClose`, `EnvironmentValues.pvSheetClose` *(new)* | struct / env | `PVSheetDismiss`, `pvContextMenuDismiss` | Wraps `PVSheetDismiss` with the pending navigation; same env shape as the context menu |
| `PVAnnouncement`, `View.pvAnnounce` *(new)* | enum / modifier | `AccessibilityAnnouncement.post` | Posts through it; adds the join and the on-change rule the kit shares |
| `PVCallout(announces:)`, `PVErrorText(announces:)` *(new param)* | param | — | Opt-in per working rule |
| `EnvironmentValues.pvFieldLabel`, `PVFieldLabel.resolve` *(new)* | env / func | `PVSelect.accessibilityLabelText` | Adds one fallback step to the select's existing order |
| `SourceArtifactsSection.requestEdit`, `confirmEditSwitch`, `cancelEditSwitch`, `PendingEditSwitch` *(new)* | func / struct | `pendingDelete` snapshots in `SourceNotesSection` / `SourceMetadataSection` | Same snapshot-for-`pvConfirm(item:)` shape |
| `WorkspaceToolbar.LocationNames` *(rename of `BreadcrumbNames`)*, `graphLabel(graphID:names:)` *(new)* | struct / func | `BreadcrumbNames`, the inline graph label in `breadcrumbItems` | The names now serve the history menu too; the helper lifts the inline label |
| `EvidenceGraphModel.canvasAccessibilityLabel` *(new)* | property | `EvidenceGraphView.accessibilityGraphLabel` | Moves the rule onto the model |
| `L10n.EvidenceGraph.canvasAccessibility(graph:)`, `L10n.ObjectOpen.notOpened`, `L10n.CitationComposer.citationLoadFailed`, `.citationLoadFailedUnnamed` *(new)* | L10n | `L10n.ObjectOpen.refused` | New sentences |
| `deleteImpact.kind.<kind>` (19), `deleteImpact.via.unknown` *(new keys)* | key | the 38 `.one` / `.other` keys and `viaUnknownOne` / `Other` | Replace them |
| `ObjectRefusal.notOpened`, `ProjectFiles.launch` *(new)* | case / func | `ObjectRefusal.missing` | Same enum |
| `CitationComposerModel.citationLoadError` *(new)* | property | `loadFailureMessage` | Replaces it |
| `TypeField.edit` *(new case)* | case | `TypeField.combo` | Same enum |
| `uninflected_counts`, `NON_INFLECTING` *(new)* | script func | `unnumbered_arguments` | Same script, same shape |

**Builds on.** `PVSheetDismiss` (the one dismiss rule), `AccessibilityAnnouncement.post` (the one post path), `PVSelect`'s label order, `pageSectionLeaf` (the Back-menu leaf format), `L10n.format` plural resolution, `PVErrorText` (S14-03), `add-ui-component`'s kit-first rule.
**Deletes.** `PVDeleteImpactControls.activate`'s `dismiss:` and `then:` parameters; `PVDeleteImpactGroupList.dismiss`; `PVSheetChrome.confirmingDiscard`; the inline discard copy in `PVSheetChrome`; `WorkspaceToolbar.BreadcrumbNames` (renamed); `HistoryJumpRow.leafLabel`'s computed body; `EvidenceGraphView.accessibilityGraphLabel`'s rule (it reads the model); `CitationComposerModel.loadFailureMessage`; `labelField`'s `.accessibilityLabel`; `ConclusionLeadModel`'s two posts; `L10n.DeleteImpact.noun*` `static let`s (38) and their catalog keys; `viaUnknownOne`, `viaUnknownOther` and their keys.
**Next adds.** S14-20 opts Promote's error callouts in with `announces: true`. A new sheet row that navigates reads `pvSheetClose`. A new kit text control inside `PVField` reads `pvFieldLabel`. A new count key with a non-inflecting word adds it to `NON_INFLECTING`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A sheet row closes through the dirty rule, then navigates | `PVSheetChrome`, `WorkspaceNavigation.go(to:)` | The pending `then` is dropped on Keep editing (AC-2.7) and handed out once (AC-2.6); navigation runs after the sheet clears, as today |
| A card switch waits on a question | `SourceArtifactsSection` | If the target card is deleted while the question shows, `beginEdit` returns at its `guard` and the old card's edit is already cancelled; `items` refresh does not clear `pendingEditSwitch`, so the confirm still closes on Discard or Keep |
| History rows read live graph names | `HistoryJumpMenuHost` | A deleted graph falls back to "Evidence graph" (AC-5.2) |
| Composer load error line | `citationStepper` | Cleared by the next successful load (AC-7.2); a Citation not listed gets the unnamed copy (AC-7.3) |
| `openObject` is async | the two open buttons | A second click while the first launch runs starts a second open, as two clicks do today |

### Scale

None — no write, list or loop changes. The history menu builds at most the history depth of rows (`NavigationHistoryStore`'s cap), each one dictionary lookup in the session's graph list.

### Edges

- **Assumptions.** `PVSheetDismiss` order (running, confirming, dirty) is unchanged; `PVSheetPendingClose` only adds the held `then` (T-4). `L10n.format` resolves a stringsdict plural whose variation omits the number (T-12); if it does not, Stop.
- **Neighbors.** The irreversible confirm keeps Return (AC-1.2). A clean form closes and navigates at once (AC-2.4). A non-graph location's label is unchanged (AC-5.5), and the crumbs are unchanged (AC-5.6). A select with an explicit label keeps it (AC-4.2).
- **Removed inputs.** A graph deleted while it is in history reads "Evidence graph" (AC-5.2).
- **Real data.** The shipped catalog's `%lld <word>` values with no variations, after this PR: `artifactChecklist.count`, `designSystem.select.optionPosition`, `properties.inspector.bindCount`, `artifactViewer.findMatchOf`, `a11y.nameValue.part.moved`, `conclusions.outcome.outvotedBy` ("of"); `deleteImpact.overflow`, `deleteImpact.summary.more`, `conclusions.place.containsMore`, `promote.page.singleSummary` ("more"); `artifactViewer.zoomPercent` ("percent"); `artifactViewer.pageSet` ("set"); `a11y.nameValue.part.moveUp` / `moveDown` ("up" / "down"); `nameValue.part.error.emptyValue` ("cannot"); `citationComposer.depictionSavePage` ("to"); `promote.done.unselected` ("unselected"); `promote.section.countPromoted` ("already"); `citationComposer.imagePixels` ("px"). The four A-20 keys move to variations. T-14 runs the script on the shipped catalog.
- **Untrusted size.** None — no data the app did not write is read.

### Schema

None — no catalog change.

### Go core

None — Swift and one script.

### FFI

None — no RPC changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Confirm default | `PVConfirmContent` | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` | D-1, D-2 |
| Sheet close request | `PVSheetChrome`, `PVSheetPendingClose`, `PVSheetClose` | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` | D-3 |
| Impact rows | `PVDeleteImpactGroupList` | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` | D-4 |
| Announcements | `PVToast`, `PVCallout`, `PVErrorText`, `PVAnnouncement` | `macos/App/DesignSystem/Components/Toast/`, `macos/App/DesignSystem/Components/Callout/`, `macos/App/DesignSystem/Components/ErrorText/`, `macos/App/Platform/AccessibilityAnnouncement.swift` | D-6, D-7 |
| Heading | `PVSectionHeader` | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` | D-9 |
| Field name | `PVField`, `PVInput`, `PVTextArea`, `PVSelect` | `macos/App/DesignSystem/Components/Field/`, `macos/App/DesignSystem/Components/Input/`, `macos/App/DesignSystem/Components/TextArea/`, `macos/App/DesignSystem/Components/Select/` | D-10 |
| Card switch | `.pvConfirm` | `macos/App/Features/Sources/SourcePageView.swift` | D-12 |
| History rows | `HistoryJumpMenuHost` | `macos/App/Features/Workspace/WorkspaceToolbar.swift` | D-13, D-14 |
| Load error | `PVErrorText` | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` | D-17; `.accessibilityIdentifier("citationComposer.citationLoadError")` |

`PVSheetChrome` state, in place of `confirmingDiscard`: `@State private var pending = PVSheetPendingClose()`; it shows the discard question while `pending.isConfirming` *(new property: true from a `.confirmDiscard` answer until `discardConfirmed()` or `keepEditing()`)*. Escape while the question shows calls `keepEditing()` (today's `.keepEditing`).

```swift
struct PVSheetClose {
    let request: (@escaping @MainActor () -> Void) -> Void
    func callAsFunction(then next: @escaping @MainActor () -> Void = {}) { request(next) }
}
```

The environment default runs `next` in a `Task` and closes nothing (previews).

States: no new empty or loading state. Accessibility: as US-3, US-4, US-5.

### Strings

L10n keys *(new)*:
- `evidenceGraph.canvas.accessibilityLabel` = "Evidence graph, %@" (`L10n.EvidenceGraph.canvasAccessibility(graph:)`)
- `objectOpen.notOpened.message` = "macOS couldn't open “%@” in its app." (`L10n.ObjectOpen.notOpened(filename:)`)
- `citationComposer.citationLoadFailed` = "Couldn't open %1$@. %2$@" (`L10n.CitationComposer.citationLoadFailed(ref:reason:)`)
- `citationComposer.citationLoadFailedUnnamed` = "Couldn't open this citation. %@" (`L10n.CitationComposer.citationLoadFailedUnnamed(reason:)`)
- `deleteImpact.kind.<kind>` for `source`, `artifact`, `evidence_graph`, `citation`, `observation`, `subject`, `source_type`, `metadata_field`, `source_credibility_grade`, `subject_type`, `property`, `property_term`, `canonical_entity`, `user`, `project`, `file`, `narrative`, `reconciliation_claim`, `item`: plural variations, `one` = today's `.one` value, `other` = today's `.other` value
- `deleteImpact.via.unknown`: `one` "%1$lld %2$@ references this", `other` "%1$lld %2$@ reference this"

Changed values (same keys):
- `promote.page.mapRestHint` → "Adds a row for every other person, event and place on this Evidence graph"
- `promote.error.graphUnavailable` → "Provenencia couldn't read this Evidence graph, so it can't list the rows to promote. Go back to the graph and open Promote again."
- `promote.error.graphUnavailableRepropose` → "Provenencia couldn't read this Evidence graph to propose the rows again. Your rows and decisions are kept."
- `evidenceGraph.subject.addPropertyUnavailable` → "Add property, unavailable — this evidence graph has no artifacts" (matches the canvas callout `This evidence graph has no artifacts`)
- `citationComposer.pairPickBoth` → "Pick two subjects on this Evidence graph. The pair decides what the bridge can say"
- Comments on those five say "graph", not "Source"

Gain plural variations (default values stay the `other` form): `promote.page.promoting` ("%lld subject on this graph" / "%lld subjects on this graph"), `citationComposer.pageCount` ("%lld page" / "%lld pages"), `citationComposer.artifactCitationCount` ("%lld citation" / "%lld citations"), `citationComposer.citationStepperCount` ("%lld observation" / "%lld observations").

Removed: the 38 `deleteImpact.kind.<kind>.one` / `.other` keys; `deleteImpact.via.unknown.one`, `deleteImpact.via.unknown.other`.

Old-scope copy (A-14 changes what the copy says the action covers): the five keys above. `promote.page.column.source` and `promote.row.promotedEarlier` keep "this Source" (§ 3 Out).

## 7. Steps

**STEP-1 — Return cancels a danger confirm**
Files: `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift`, `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVConfirmTests.swift`. Skill: `add-ui-component`.
Do: D-1, D-2. README: replace "**Focus starts on cancel**, so Return cannot complete a destructive action by reflex." with "A danger confirm's default action is Cancel: Return keeps the record, and the destructive button has no Return shortcut. Focus starts on Cancel. An irreversible confirm keeps Return on its confirm button."
Check: T-1, T-2 pass.

**STEP-2 — Kit components announce themselves**
Files: `macos/App/Platform/AccessibilityAnnouncement.swift`, `macos/App/DesignSystem/Components/Toast/PVToast.swift`, `macos/App/DesignSystem/Components/Callout/PVCallout.swift`, `macos/App/DesignSystem/Components/ErrorText/PVErrorText.swift`, `macos/App/Features/Conclusions/ConclusionLeadModel.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/DesignSystem/README.md` (Toast, Callout and ErrorText rows), `macos/ProvenenciaTests/AccessibilityAnnouncementTests.swift` *(new)*. Skill: `add-ui-component`.
Do: D-6, D-7, D-8.
Check: T-3 passes; `rg -n "AccessibilityAnnouncement.post" macos/App/Features/Conclusions/ConclusionLeadModel.swift` has no hit.

**STEP-3 — Section headers are headings**
Files: `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `macos/App/DesignSystem/README.md` (SectionHeader row: "the title is a VoiceOver heading").
Do: D-9.
Check: builds; manual check 4.

**STEP-4 — Fields name their content**
Files: `macos/App/DesignSystem/Components/Field/PVField.swift`, `macos/App/DesignSystem/Components/Input/PVInput.swift`, `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/App/DesignSystem/Components/Select/PVSelect.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `macos/App/DesignSystem/README.md` (Field row: "gives its label to the `PVInput`, `PVTextArea` or `PVSelect` inside it through `pvFieldLabel`; a control's own label wins"), `macos/ProvenenciaTests/PVFieldLabelTests.swift` *(new)*. Skill: `add-ui-component`.
Do: D-10, D-11.
Check: T-10 passes.

**STEP-5 — A sheet's links ask before a dirty form closes**
Files: `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`, `macos/App/DesignSystem/README.md` (sheet section: content closes through `pvSheetClose`), `macos/ProvenenciaTests/PVSheetPendingCloseTests.swift` *(new)*, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`.
Do: D-3, D-4, D-5. `SourceGraphForm` drops `dismiss`.
Check: T-4, T-5 pass; `rg -n "dismiss: \{ model = nil \}" macos/App` has no hit.

**STEP-6 — An Artifact card asks before another card drops its edit**
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Do: D-12.
Check: T-6 passes.

**STEP-7 — The canvas and history name the graph**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-localized-string`.
Do: D-13, D-14, D-15.
Check: T-7, T-8 pass; every existing `WorkspaceToolbarBreadcrumbTests` test passes unchanged except the `BreadcrumbNames` → `LocationNames` rename.

**STEP-8 — Copy names the Evidence graph**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: § 6 Strings, changed values.
Check: `rg -n "this Source" macos/App/Platform/L10n.swift` lists only `promote.page.column.source`, `promote.row.promotedEarlier`, the Source-scoped keys outside A-14 (`citationComposer` `None on this Source`, the Artifact checklist note) and comments; T-9 passes.

**STEP-9 — Counts take plurals, and the script holds them**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/L10nPluralTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`, `scripts/check-localizable-xcstrings.py`. Skill: `add-localized-string`.
Do: D-16, D-20, § 6 Strings plurals. The script's docstring gains a paragraph for the count check.
Check: T-11, T-12, T-13, T-14 pass.

**STEP-10 — A failed open says so**
Files: `macos/App/Platform/ProjectFiles.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Do: D-18. Existing `openSelectedFile…` and `openObjectReports…` tests `await` the new async calls.
Check: T-15, T-16 pass.

**STEP-11 — A failed Citation load names the Citation**
Files: `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/CitationComposerModelTests.swift`.
Do: D-17.
Check: T-17 passes.

**STEP-12 — Focus returns to the Type control**
Files: `macos/App/Features/Sources/SourcePageIdentityHeader.swift`.
Do: D-19.
Check: builds; manual check 9.

**STEP-13 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md` (a `## S14-14 — Copy and VoiceOver on graph work` entry at the top); `docs/deployment-plan/spike-14/deployment-plan.md` (map row Status → `Landed #N`; S14-14 scope entry: replace "If the plan runs past about ten steps, it splits along these parts." with "One PR (user, 2026-11-01)."; the Form fields part: "the Promote evidence sheet B-18 named went in S13-23; the kit fix covers every form"); `docs/deployment-plan/spike-14/app-health-audit.md` (A-12, A-13, A-14, A-17, A-18, A-20, A-21: fixed in S14-14); `docs/deployment-plan/spike-14/app-health-audit-2.md` (B-18: fixed in S14-14, kit fix); `.agents/skills/add-localized-string/SKILL.md` (one line: a count in copy is a plural key; `check-localizable-xcstrings.py` fails on an uninflected `%lld <word>`); `.agents/skills/add-ui-component/SKILL.md` (one line: a toast announces itself; an error callout or error line passes `announces: true`; a sheet's content closes through `pvSheetClose`); this plan (Status, Landed note); `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj` → 0.2.6.
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-14-graph-work-voiceover.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVConfirmTests.swift` `dangerReturnCancels` *(new)* | `.danger.returnConfirms == false`; `.irreversible.returnConfirms == true` | AC-1.1, AC-1.2, D-1 |
| T-2 | Swift | `PVConfirmTests` `discardCopyIsTheKitsQuestion` *(new)* | `PVConfirmCopy.discardChanges` title, message, confirm, cancel equal the four `formDialogDiscard*` strings | D-2 |
| T-3 | Swift | `macos/ProvenenciaTests/AccessibilityAnnouncementTests.swift` *(new)* `joinsTitleAndMessage`, `nothingToSayIsNil`, `dangerIsHigh` | AC-3.1, AC-3.2, AC-3.3, AC-3.4 inputs → outputs | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-6 |
| T-4 | Swift | `macos/ProvenenciaTests/PVSheetPendingCloseTests.swift` *(new)* `dirtyHoldsThenUntilDiscard`, `keepEditingDropsThen`, `cleanClosesAtOnce`, `runningIgnores` | AC-2.6; AC-2.7; `isDirty: false` → `.close`, nothing held; `isRunning: true` → `.ignore` | AC-2.6, AC-2.7, AC-2.4, AC-2.5, D-3 |
| T-5 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` (the `activate` test near line 207, rewritten) | `activate(row, onNavigate:, close: PVSheetClose { next in closed = true; next() })` → `closed`, then the row's location navigated | D-4 |
| T-6 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `cardSwitchAsksWhenDirty`, `cardSwitchConfirmMovesEdit`, `cardSwitchCancelKeepsEdit`, `cardSwitchCleanMovesAtOnce` *(new)* | AC-2.8, AC-2.9, AC-2.10, AC-2.11 | AC-2.8, AC-2.9, AC-2.10, AC-2.11, D-12 |
| T-7 | Swift | `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift` `locationLabelsNameTheGraph` *(new)*, `locationLabelsFallBackToEvidenceGraph` *(new)*, `composerAndPromoteLabelsCarryTheSurface` *(new)*; existing `evidenceGraphCrumbsNameSourceAndGraph`, `citationComposerCrumbsNameSourceAndGraph`, `promoteCrumbsNameSourceAndGraph`, `locationLabelsNameTheSection` | AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, AC-5.6 | AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, AC-5.6, D-13, D-14 |
| T-8 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `canvasLabelNamesTheGraph` *(new)* | AC-5.7 | AC-5.7, D-15 |
| T-9 | Swift | `macos/ProvenenciaTests/L10nPluralTests.swift` `graphCopyNamesTheGraph` *(new)* | each of the five A-14 strings resolved; none contains "this Source"; each contains "vidence graph" | AC-6.7 |
| T-10 | Swift | `macos/ProvenenciaTests/PVFieldLabelTests.swift` *(new)* `fieldLabelFillsIn`, `ownLabelWins`, `noLabelStaysNil` | AC-4.1, AC-4.2, AC-4.3 | AC-4.1, AC-4.2, AC-4.3, D-10 |
| T-11 | Swift | `L10nPluralTests` `countsTakePlurals` *(new)* | AC-6.1, AC-6.2 | AC-6.1, AC-6.2 |
| T-12 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` `nounReadsPluralKey`, `unknownViaHeadingPlural` *(new)* | AC-6.3, AC-6.4; `noun("unknown_kind", count: 1)` → "item" | AC-6.3, AC-6.4, D-16 |
| T-13 | Swift | `L10nPluralTests` `everyCatalogPluralPicksAFormPerCount` (changed per D-16) | passes with the 20 new keys | D-16 |
| T-14 | Script | `scripts/check-localizable-xcstrings.py` | the shipped catalog → exit 0; a scratch copy with `"test.planted": {"localizations": {"en": {"stringUnit": {"state": "translated", "value": "%lld subjects"}}}}` added, path passed as `argv[1]` → exit 1, output names `test.planted` | AC-6.5, AC-6.6, D-20 |
| T-15 | Swift | `macos/ProvenenciaTests/ProjectFilesTests.swift` `openObjectReportsNotOpenedWhenLaunchThrows` *(new)* | a regular `.pdf` under `objects/`, `launch` throws → `.notOpened` | AC-7.5, D-18 |
| T-16 | Swift | `EvidenceGraphModelTests` `openSelectedFileReportsALaunchFailure` *(new)*; `SourcePageModelTests` `openReportsALaunchFailure` *(new)* | throwing `launch` → `fileOpenError` / `pageError` equals `L10n.ObjectOpen.notOpened(filename:)` | AC-7.6, AC-7.7 |
| T-17 | Swift | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `loadFailureNamesTheCitation`, `loadFailureClearsOnNextLoad`, `loadFailureUnlistedIsUnnamed`, `loadFailureLeavesFieldError` *(new)* | Two seeded Citations `cit-1` (`CIT-1`) and `cit-2` (`CIT-2`); after `prepare()`, `store.citationsByID["cit-2"] = nil` (the listed row stays in the session) and `selectCitation("cit-2")` → AC-7.1 with `CIT-2`; restore and select `cit-1` → AC-7.2; `lineRefreshShowsAFailedCitationLoad` (existing, its `fields.error` expectations become `citationLoadError`) → AC-7.3; a save error set in `fields.error` first → AC-7.4 | AC-7.1, AC-7.2, AC-7.3, AC-7.4, D-17 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVDeleteImpactTests`, `WorkspaceToolbarBreadcrumbTests`, `L10nPluralTests`, `CitationComposerModelTests` (`lineRefreshShowsAFailedCitationLoad` reads `fields.error` after a load failure), `EvidenceGraphModelTests` (`openSelectedFile…`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`BreadcrumbNames`, `loadFailureMessage`, `confirmingDiscard`, `viaUnknownOne`, `viaUnknownOther`, `deleteImpact.kind.source.one`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (the scope entry's split line, STEP-13)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no new kit component; the kit additions are parameters, environment values and one modifier)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — 0.2.6
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none in docs/*.md; macos/App/DesignSystem/README.md (STEP-1 – STEP-5), the two audits, add-localized-string and add-ui-component (STEP-13)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with a project that has a Source "1881 Census" with two Evidence graphs, "Jones household" and "Smith household", each with a Citation, and an Artifact in each graph. VoiceOver on (Cmd-F5) where a line says so.

1. Source page → Evidence graphs → delete "Smith household" → the confirm shows → press Return. The sheet closes and "Smith household" is still listed (AC-1.3).
2. Edit "Jones household", type "Jones family" in the name, press Escape, then Return on "Discard changes?". The form stays with "Jones family" (AC-1.4).
3. Still dirty, press a Kept-linked Citation row. "Discard changes?" shows; Keep editing keeps the form (AC-2.3); press the row again and Discard: the sheet closes, then the Citation opens (AC-2.1, AC-2.2).
4. VoiceOver: Add an Evidence graph. VoiceOver says "Evidence graph added" once (AC-3.5). Rotor → Headings lists "Evidence graphs" and the other section titles (AC-3.7).
5. VoiceOver: save the graph form with a name another graph uses (or a forced error): VoiceOver speaks the callout once (AC-3.6). On a Person page, set a portrait: the toast is spoken once (AC-3.8).
6. VoiceOver: open the graph form and move to the description area: "Description" is spoken (AC-4.4). In the composer, move to the citation label field: its label is spoken once (AC-4.5).
7. Open "Jones household", then a Citation in it, then go back to the Source page. Long-press Back: rows read "1881 Census", "Jones household", "Jones household · Citation for CIT-…" (AC-5.3). VoiceOver on the canvas says "Evidence graph, Jones household" (AC-5.7).
8. Edit an Artifact card's label without saving, then press the pencil on another card: "Discard changes?" shows; Keep editing keeps the text (AC-2.10); Discard moves the edit (AC-2.9).
9. Full Keyboard Access on: Source page, press the Type pencil, then the ×. Focus is on the Type pencil (AC-7.8). Pick a type instead: focus is on the pencil again.
10. In the composer, a Promote page and a delete notice, check "1 citation", "1 page", "1 observation", "1 subject on this graph" where one exists (AC-6.1, AC-6.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A stringsdict plural whose variation omits the number does not resolve, or `xcstringstool` warns on it | T-12 catches the first; check-pr's no-warning build catches the second → Stop |
| `.accessibilityLabel` inside `PVInput` overrides a call site's outer label | D-10 applies the field label only when the resolver returns one, and a call-site label is `own`; T-10 holds the rule; manual check 6 |
| A toast that re-renders with the same text speaks twice | `pvAnnounce` posts on appear and on change to a new value only (D-6) |
| `locationLabels` losing `nonisolated` breaks a caller | Its only callers are `breadcrumbItems` and `HistoryJumpMenuHost` (§ 4), both main-actor |

**Stop and report if:**

- `PVConfirmContent`, `PVSheetChrome` or `PVDeleteImpactGroupList` no longer look as § 4 says.
- T-12 fails because `L10n.DeleteImpact.noun("citation", count: 1)` returns a key or a placeholder, or the build warns on a plural variation without a number.
- The Swift build warns under `SWIFT_STRICT_CONCURRENCY = complete` for `LocationNames`' closures or `PVSheetClose`, and the fix needs an isolation change outside the files in § 7.
- `NSWorkspace.shared.open(_:withApplicationAt:configuration:)` has no `async throws` form in the SDK the project builds with.
- The planted-key run of the script passes (D-20 does not catch `%lld subjects`).
- A `PVField` call site's VoiceOver name gets worse in manual check 6 (a control speaks the field label in place of a more specific label it had).

## 12. Open questions

None.

## 13. PR

**Title:** S14-14 — Copy and VoiceOver on graph work
**Body:** objective in one sentence; what shipped against US-1 – US-7; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-14-graph-work-voiceover.md`.
