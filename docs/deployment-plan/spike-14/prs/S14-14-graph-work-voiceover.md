# S14-14 — Copy and VoiceOver on graph work

**Status:** Landed in [#675](https://github.com/mendahu/provenencia/pull/675) (2026-11-01)
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: no layout or component changes; copy is rewritten in place, and the one moved error line uses the kit's existing `PVErrorText` under the Citation picker (user, 2026-11-01); `PVButtonStyle` draws no default-button tint, so moving Return to Cancel draws nothing different)
**Board:** —
**Depends on:** —
**Blocks:** S14-20 (reads `PVCallout`'s `announces:`)
**Branch:** `s14-14-graph-work-voiceover`
**Version bump:** patch, 0.2.5 → 0.2.6 — accessibility, copy and safety fixes to existing screens, no feature ([`docs/versioning.md`](../../../versioning.md)). If `main` is past 0.2.5 when the PR merges, take the next patch.
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc` (one parity subtest), `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** audit [A-12](../app-health-audit.md#a-12--nothing-says-which-evidence-graph-you-are-on), [A-13](../app-health-audit.md#a-13--a-kept-linked-citation-link-discards-a-dirty-graph-form), [A-14](../app-health-audit.md#a-14--copy-says-this-source-where-the-action-is-one-graph), [A-17](../app-health-audit.md#a-17--return-confirms-destructive-deletes), [A-18](../app-health-audit.md#a-18--voiceover-is-not-told-graph-write-results-or-failures), [A-20](../app-health-audit.md#a-20--counts-without-plural-rules), [A-21](../app-health-audit.md#a-21--section-headers-are-not-voiceover-headings), audit 2 [B-18](../app-health-audit-2.md#b-18--the-sheets-status-confidence-and-argument-fields-have-no-spoken-name); working rules "Kit components announce themselves", "A danger confirm's default action is Cancel", "A dirty form asks before a link inside it navigates away", "No graph name in the canvas header", "Audit items keep their IDs"; bar outcome 17; the S14-14 scope entry's Spike 12 ship-log leftovers and its Spike 11 leftover
**Landed:** [#675](https://github.com/mendahu/provenencia/pull/675), 2026-11-01. Differences from the plan: (1) the existing `artifactSwitchShowsAFailedCitationLoad` read `fields.error` and now reads `citationLoadError` (D-16); (2) `activeCitationRef` and `announceIdentityChange()` share a private `listedCitation(_:)` lookup beside `citationRef(for:)`; (3) the kind-noun catalog entries went through the D-19 revision (a `%#@noun@` substitution), after the round-0 probe found a top-level plural variation does not compile; (4) the unknown-kind preview still passes `narrative`, which now reads "item"; (5) the kit's separate join key (D-6) is replaced by one `designSystem.spokenJoin` that the Conclusions and Last-updated joins also use. The manual check (section 10) is for the user to run with VoiceOver.
**Planned:** 2026-11-01 against `main` @ `75d9494`
**Review:** [`S14-14-graph-work-voiceover.review.md`](S14-14-graph-work-voiceover.review.md) — Fits with changes — Ready with changes, 2 rounds plus a scoped round 3 on the D-19 revision; every finding answered

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

Keyboard and VoiceOver are safe on graph work. Return in a delete confirm cancels. A dirty graph form, or a dirty Artifact card, asks "Discard changes?" before a link or another card takes it away. VoiceOver hears every toast and the error callouts and error lines on graph work, can jump between page sections by heading, hears each form field's name, and hears which Evidence graph is open on the canvas and in the Back / Forward menus. Counts read "1 subject", not "1 subjects", and a script keeps new counts honest. The leftovers come with it: a failed Citation load names the Citation under the Citation picker, a file macOS fails to open says so, and the Source type editor hands focus back to the Type control when it closes. The work is audit items A-12, A-13, A-14, A-17, A-18, A-20, A-21 and B-18, which Spike 12 planned as S12-29, S12-30 and S12-45. The map said to split the work along those parts if the plan ran past ten steps; the user chose one PR (2026-11-01).

**Why now.** Every slice 5 audit fix on graph work lands on this kit behavior. S14-20 opts the Promote page's errors into the callout announcement this PR adds.

**What stays.** No page's layout, and no component's look, changes. The canvas header names no graph (working rule "No graph name in the canvas header"). Toast timing stays at 5 s. An irreversible-tone confirm keeps Return on its confirm button. Go code is untouched; one parity subtest is added in `api/ffi/swift_vocabulary_test.go`.

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
| AC-2.3 | AC-2.1, then Keep editing (or Escape, or Return) | The question closes; the form shows "Jones family"; no navigation |
| AC-2.4 | Graph Edit form not dirty, row pressed | The sheet closes and navigates, as today |
| AC-2.5 | Graph form saving (`isSaving`) | The Kept-linked rows are disabled; a press does nothing |
| AC-2.6 | `PVDiscardGate`: `hold(f)`, then `discard()` | returns `f` once; a second `discard()` returns nil; `holdsStep` is false after |
| AC-2.7 | `PVDiscardGate`: `hold(f)`, then `keepEditing()`, then `discard()` | `keepEditing()` returns true; `discard()` returns nil: Keep editing drops the step |
| AC-2.8 | `PVDiscardGate` holding nothing: `keepEditing()`; and the sheet: a dirty row press, then Escape while the question shows (manual check 3) | returns false (the composer then cancels its held navigation); the sheet keeps the form open and drops the row's navigation |
| AC-2.9 | Artifact `ART-1` in edit, label changed from "Page 1" to "Page one", pencil on `ART-2` pressed (`beginEdit("ART-2")`) | `pendingEditSwitch != nil`; `editingArtifactID == "ART-1"`; the label buffer still reads "Page one"; `acceptsArtifactPageDrop == false` |
| AC-2.10 | AC-2.9, then `confirmEditSwitch()` | `editingArtifactID == "ART-2"`; `pendingEditSwitch == nil`; `ART-1`'s buffer is back to "Page 1" |
| AC-2.11 | AC-2.9, then `cancelEditSwitch()` | `pendingEditSwitch == nil`; `editingArtifactID == "ART-1"`; buffer "Page one" |
| AC-2.12 | `ART-1` in edit and not dirty, **Link to a graph** on `ART-2` pressed | `editingArtifactID == "ART-2"` at once; no question |
| AC-2.13 | Composer: a dirty Citation, another Citation picked (`selectCitation`), then Discard; and the same with Keep editing | As today: `pendingLeave?.keepDraft == true` while asking; the switch runs after Discard; Keep editing leaves the open Citation; the existing leave tests pass with `leaveGate` in place of `stepContinuation` |

**US-3.** As a VoiceOver user, I hear graph write results and failures, and can move by heading.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `AccessibilityAnnouncement.text(title: "Evidence graph added", message: nil)` | `"Evidence graph added"` |
| AC-3.2 | `AccessibilityAnnouncement.text(title: "Couldn't save the evidence graph", message: "The name is taken.")` | `"Couldn't save the evidence graph, The name is taken."` (through `designSystem.announcement.join`) |
| AC-3.3 | `AccessibilityAnnouncement.text(title: nil, message: "")` | `nil` (nothing is posted) |
| AC-3.4 | `AccessibilityAnnouncement.priority(danger: true)` / `(danger: false)` | `.high` / `.medium` |
| AC-3.5 | Graph added on the Source page, VoiceOver on (manual) | VoiceOver says "Evidence graph added" once |
| AC-3.6 | Graph form save fails, VoiceOver on (manual) | VoiceOver says the callout's title and message once, focus still on Save |
| AC-3.7 | Source page, VoiceOver rotor → Headings (manual) | "Evidence graphs" and every other section title are listed; VO-Cmd-H moves between them |
| AC-3.8 | Set a portrait on a Person page, VoiceOver on (manual) | The toast is spoken once, not twice (the model's own post is gone) |
| AC-3.9 | `PVCopy.pvString` for `"Jones" as String`, `L10n.Sources.graphFormSaveErrorTitle`, `AttributedString("Smith")` | `"Jones"`, `L10n.string(L10n.Sources.graphFormSaveErrorTitle)`, `"Smith"` |

**US-4.** As a VoiceOver user, every form field says its name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `PVFieldLabel.resolve(own: nil, field: "Name")` | `"Name"` |
| AC-4.2 | `PVFieldLabel.resolve(own: "Filter", field: "Name")` | `"Filter"` (a control's own label wins) |
| AC-4.3 | `PVFieldLabel.resolve(own: nil, field: nil)` | `nil` (the control keeps what it says today) |
| AC-4.4 | Graph form, VoiceOver on the description text area (manual) | VoiceOver says "Description, …", not only the text |
| AC-4.5 | Composer, VoiceOver on the citation label field (manual) | VoiceOver says the field's label once (the hand-made label is gone) |

**US-5.** As a researcher moving between graphs, the canvas and the history menus name the graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `WorkspaceLocationLabels.labels` for `.evidenceGraph(sourceId: "S1", graphId: "G1", ref: "SRC-1", title: "1881 Census")`, names `G1 → "Jones household"` | leaf `"Jones household"`, section `nil` |
| AC-5.2 | The same with names knowing no graph | leaf `"Evidence graph"` |
| AC-5.3 | `.citationComposer(sourceId: "S1", graphId: "G1", ref: "CIT-K4M2P", title: "CIT-K4M2P")`, names `G1 → "Jones household"` | leaf `"Jones household"`, section `"Citation for CIT-K4M2P"`; the Back menu row reads `"Jones household · Citation for CIT-K4M2P"` |
| AC-5.4 | A Promote location for `PER-7H2Q` on `G1`; and one with `promoteAll` | section `L10n.Promote.breadcrumb(ref: "PER-7H2Q")`; and `L10n.EvidenceGraph.promoteAll` |
| AC-5.5 | `.sourcePage(sourceId: "S1", title: "1881 Census")` | leaf `"1881 Census"`, section `nil` (unchanged) |
| AC-5.6 | `WorkspaceToolbar.breadcrumbItems` for AC-5.1 and AC-5.3 | Labels unchanged from today: `["Sources", "1881 Census", "Jones household"]` and `[…, "Jones household", "Citation for CIT-K4M2P"]` |
| AC-5.7 | `EvidenceGraphModel.canvasAccessibilityLabel` with graph name "Jones household", nothing armed; with no name known; with connect armed | `"Evidence graph, Jones household"`; `"Evidence graph"`; `L10n.string(model.connectArmedHint)` |

**US-6.** As a researcher, copy says what the action covers and counts take plurals.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | `L10n.Promote.promoting(count: 1)` / `(count: 3)` | `"1 subject on this graph"` / `"3 subjects on this graph"` |
| AC-6.2 | `L10n.CitationComposer.pageCount(count: 1)`, `.artifactCitationCount(count: 1)`, `.citationStepperCount(count: 1)` | `"1 page"`, `"1 citation"`, `"1 observation"`; with 2, the plural |
| AC-6.3 | `L10n.DeleteImpact.noun("citation", count: 1)` / `(…, count: 2)`; `noun("identity_claim")`; `noun("unknown_kind")` | `"citation"` / `"citations"`; `"identity claim"`; `"item"` |
| AC-6.4 | `L10n.DeleteImpact.viaFallback(kind: "file", total: 1)` / `(…, total: 3)` | `"1 file references this"` / `"3 files reference this"` |
| AC-6.5 | A catalog copy with a planted key `"test.planted": "%lld subjects"` (no variations), script run on it | `check-localizable-xcstrings.py` exits 1 and its output names `test.planted` |
| AC-6.6 | The shipped catalog | the script passes |
| AC-6.7 | `L10n.Promote.mapRestHint`, `graphUnavailable`, `graphUnavailableRepropose`, `L10n.EvidenceGraph.addPropertyUnavailable`, `L10n.CitationComposer.pairPickBoth` | each names the Evidence graph and none says "this Source" (§ 6 Strings) |
| AC-6.8 | Go's `catalogmodel` `Kind` constants vs the Swift `deleteImpact.kind.<kind>` keys | the two sets are equal, plus `item` on the Swift side |

**US-7.** As a researcher, an error names what failed and focus lands where I was.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-7.1 | Composer open on `cit-1` (`CIT-1`); `cit-1` removed from FakeStore's `citationsByID` and the Artifact's citation list invalidated; `refreshCitationLines()` | `citationLoadError` starts with "Couldn't open CIT-1."; `fields.error == nil` |
| AC-7.2 | AC-7.1, then `cit-1` restored and `refreshCitationLines()` | `citationLoadError == nil` |
| AC-7.3 | `selectCitation("cit-missing")`, an id never seeded | `citationLoadError` starts with "Couldn't open this citation." |
| AC-7.4 | AC-7.1 with a save error already in `fields.error` | `fields.error` keeps the save error |
| AC-7.5 | `ProjectFiles.openObject` for a stored PDF whose `launch` throws a non-refusal error; and one whose `launch` throws `ObjectRefusal.missing` | `.notOpened`; `.missing` |
| AC-7.6 | `L10n.ObjectOpen.message(for: .notOpened, filename: "scan.pdf", missing: "x")`; `(.unsafe, …)`; `(.missing, …)` | `L10n.ObjectOpen.notOpened(filename: "scan.pdf")`; `L10n.ObjectOpen.refused(filename: "scan.pdf")`; `"x"` |
| AC-7.7 | Source page, Type pencil, then the × or a picked type (manual, Full Keyboard Access on) | Focus is on the Type pencil button |
| AC-7.8 | Composer, save the depiction's page fails (manual, VoiceOver on) | The error line is spoken once (the hand post is gone) |
| AC-7.9 | AC-7.1, then `beginNewCitation()` (a blank Citation) | `citationLoadError == nil` |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-17: danger confirm's default is Cancel; README line | STEP-1, AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| A-18: `PVToast` announces; `PVCallout` and `PVErrorText` take `announces:`; the graph form, Artifact card, Source page and canvas file-open callouts and the composer's save-page error line opt in; `ConclusionLeadModel`'s posts and `savePage`'s error post go | STEP-2, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.8, AC-3.9, AC-7.8 |
| A-21: `PVSectionHeader` title is a heading | STEP-3, AC-3.7 |
| B-18: `PVField` gives its label to the kit control inside it; the composer's hand-made label goes | STEP-4, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5 |
| A-13: Kept-linked rows close through the sheet's dirty rule, then navigate; disabled while saving | STEP-5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8 |
| Leftover: an Artifact card's pencil or **Link to a graph** asks before dropping another card's unsaved edit | STEP-6, AC-2.9, AC-2.10, AC-2.11, AC-2.12 |
| The composer's held step moves onto `PVDiscardGate` (user, 2026-11-01, review M-13) | STEP-11, AC-2.13 |
| A-12: one location-label home for crumbs, history rows and the canvas VoiceOver label | STEP-7, AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, AC-5.6, AC-5.7 |
| A-14: five strings name the Evidence graph | STEP-8, AC-6.7 |
| A-20: plural variations for four keys; `DeleteImpact.noun` and the unknown-via heading read plural keys from Go's kinds; the script fails on an uninflected count | STEP-9, AC-6.1, AC-6.2, AC-6.3, AC-6.4, AC-6.5, AC-6.6, AC-6.8 |
| Leftover: `ProjectFiles.openObject` reports a failed launch | STEP-10, AC-7.5, AC-7.6 |
| Leftover: a failed Citation load names the Citation under the Citation picker | STEP-11, AC-7.1, AC-7.2, AC-7.3, AC-7.4, AC-7.9 |
| Spike 11 leftover: focus returns to the Type control after the Source type editor | STEP-12, AC-7.7 |

| Out | Home |
| --- | --- |
| A graph name in the canvas header | Refused (working rule "No graph name in the canvas header") |
| Tooltips on truncated names | Not planned (scope entry Not this PR) |
| Holding toasts longer while VoiceOver runs | Not planned (scope entry Not this PR) |
| Promote page errors scrolled into view and announced; `PromoteAlignmentRow`'s hand-rolled focus hand-back moving to `FocusClaim` | S14-20 (it edits `PromoteView.swift`) |
| Notes and Metadata rows asking before another row drops their draft (`SourceNotesSection.beginEdit(id:)`, `SourceMetadataSection.beginEdit(fieldID:)`) | Dogfood log; adopts `PVDiscardGate` when picked up |
| "On this Source" (`promote.page.column.source`) and "Promoted from this Source earlier" (`promote.row.promotedEarlier`): A-14 does not list them, and both are true of the Source | Not planned |
| Speaking a field's required or optional marker | Not planned |
| A control that is not a `PVInput`, `PVTextArea` or `PVSelect` inside a `PVField` (the drop row, the checklist; `PVComboBox` already takes a required label) learning the field label | Not planned; those rows already speak their own text |
| `NameValueEditorForm.moveAndAnnounce` posting a move by hand | Not planned; it posts after an action, not on a change of shown text (the composer's cut-failed caption post moves onto `pvAnnounce` in STEP-2, D-8) |

## 4. Current state

Verified at `75d9494`.

| Area | Where | Today |
| --- | --- | --- |
| Confirm | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` `PVConfirmContent` | Confirm button `.keyboardShortcut(.defaultAction)` for every tone; `.onAppear { cancelFocused = true }` with the comment "Return must not complete a destructive action" — but `.defaultAction` fires on Return whatever has focus (A-17). `PVConfirmTone` has `buttonRole`, `buttonVariant`, `confirmIcon` |
| Discard question | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `PVSheetChrome` | Private `confirmingDiscard`; builds the discard `PVConfirmCopy(title:message:confirm:cancel:)` inline from `L10n.DesignSystem.formDialogDiscard*`; `close` clears the binding; no way for content to ask "close, then run this" |
| Dismiss rule | `macos/App/DesignSystem/Components/Panel/PVSheetDismiss.swift` `PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard:)` | `.ignore` / `.keepEditing` / `.confirmDiscard` / `.close` |
| Kept-linked rows | `macos/App/Features/Sources/SourceGraphFormDialog.swift` `SourceGraphFormDialogModifier.body`, `SourceGraphForm.keptLinked` | `dismiss: { model = nil }` skips the dirty check (A-13); rows not disabled while saving |
| Row activation | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactControls.activate(_:onNavigate:dismiss:then:)`, `PVDeleteImpactGroupList` (`dismiss` param; doc comment names `activate(_:onNavigate:dismiss:then:)`), `PVDeleteImpactNotice` (`onDismiss`; 12 call sites with previews), presented in `.pvSheet(item: noticeItem)` | `dismiss()` then navigates in a `Task`; `then:` has no caller |
| Composer held step | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `stepContinuation` (set in `selectCitation`, `requestStep`; `CitationComposerModel+Authoring.swift` `openAuthoring`), `discardLeaveChanges()` (runs it once, else `resumeHeldNavigation()`), `keepEditingAfterLeave()` (drops it; with none held, `cancelHeldNavigation()`); `pendingLeave` lives in the composer's `presented` dialog state | The composer's own copy of "hold the next step behind Discard changes?"; no test reads `stepContinuation` |
| Citation ref | `CitationComposerModel.activeCitationRef` (`listedCitations` row, else `loadedCitationRef`); `announceIdentityChange()` repeats the `listedCitations.first(where:)` lookup | Two copies of the id → ref rule |
| Caption post | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `.onChange(of: model.depictionCaptionState) { … if state == .cutFailed { AccessibilityAnnouncement.post(…) } }` | Hand-written post on change |
| Leave guard | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `WorkspaceLeaveGuard`, `heldNavigation`, `resumeHeldNavigation()`; adopted by `CitationComposerModel`, `PromoteModel` | Holds a `go(to:)` behind a place's own discard question. A sheet's gate is a different layer: it asks for the sheet's form, then `go(to:)` still meets the place's guard |
| Env precedent | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `EnvironmentValues.pvContextMenuDismiss`; `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` `PVWidthClassKey` | The kit's shape for an environment action and value |
| Toast | `macos/App/DesignSystem/Components/Toast/PVToast.swift` `PVToast` | `title` / `message` are `String`; no announcement (A-18). Callers: `VocabularyToastOverlay` (`macos/App/Features/CatalogVocabulary/VocabularyChrome.swift`, `.id(toast)`), `OnboardingView`, `CitationComposerSubjectGroups` |
| Own posts | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` (two `AccessibilityAnnouncement.post` beside each `toast =`, one joining `"\(title), \(body)"`); `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `savePage()` (`else if let error = model.fields.error { AccessibilityAnnouncement.post(error) }`) beside the `PVErrorText(error)` with id `citationComposer.depiction.savePageError` | Each posts what a kit component shows; with kit announcements they would speak twice |
| Callout | `macos/App/DesignSystem/Components/Callout/PVCallout.swift` `PVCallout` | Stores `title: Text?` from `(any PVCopy)?`; no announcement |
| Copy protocol | `macos/App/DesignSystem/Components/Copy/PVCopy.swift` `PVCopy` (`pvText`); conformers `LocalizedStringResource`, `String`, `AttributedString` | No String form |
| ErrorText | `macos/App/DesignSystem/Components/ErrorText/PVErrorText.swift` `PVErrorText` | Non-optional `String`; no announcement |
| Announce | `macos/App/Platform/AccessibilityAnnouncement.swift` `AccessibilityAnnouncement.post(_:priority:)` | The one post path |
| Spoken joins | `macos/App/Platform/L10n.swift` `L10n.Conclusions.a11yList`, `L10n.LastUpdated.a11yJoin` | Joins are catalog keys |
| Error callouts | `SourceGraphFormDialog.swift` (`PVCallout`, `sources.graphForm.error`); `macos/App/Features/Sources/SourcePageArtifactsView.swift` `editFields` (`PVCallout`, `sources.page.artifact.<id>.saveError`); `macos/App/Features/Sources/SourcePageView.swift` (`PVCallout`, `sources.page.error`, `pageError`); `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (`PVCallout`, `evidenceGraph.fileOpen.callout`) | Silent when they appear |
| Section header | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` | Title not a heading (A-21). Precedent: `ConclusionDetailPage.swift` `.accessibilityAddTraits(.isHeader)` |
| Field | `macos/App/DesignSystem/Components/Field/PVField.swift` | `label: LocalizedStringResource?` is a sibling `Text`; content gets no name (B-18). about 43 call sites under `macos/App` outside previews |
| Field controls | `macos/App/DesignSystem/Components/Input/PVInput.swift` (`TextField(text:prompt:) { EmptyView() }`, no label parameter), `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` (`TextEditor`, no label parameter), `macos/App/DesignSystem/Components/Select/PVSelect.swift` `accessibilityLabelText` (explicit resource → spoken String → trigger text), `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` (`label:` required, already spoken) | Input and text area have no spoken name; a select inside a field speaks its trigger text |
| Hand-made label | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` `labelField` | `.accessibilityLabel(Text(L10n.CitationComposer.descriptionLabel))` on the `PVInput` — the only call-site label on a control inside a `PVField` |
| Card edit | `macos/App/Features/Sources/SourceArtifactsSection.swift` `beginEdit(_:)`, `fieldsDirty(_:)`, `cancelFields(id:)`, `savingID` | `beginEdit` cancels another card's edit with no question; its only callers are the pencil and **Link to a graph** in `SourcePageArtifactsView.metaRow` / `noGraphRow`, and tests |
| Page drop | `macos/App/Features/Sources/SourcePageModel.swift` `acceptsArtifactPageDrop` | Lists every Source page sheet that refuses a drop |
| Labels | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `BreadcrumbNames` (built inline at the `breadcrumbs` property), `breadcrumbItems(for:names:goTo:)` (graph, composer and Promote crumb labels inline, with `promoteAll` and the `scope` fallback), `locationLabels(for:)` (`nonisolated`) | The crumbs read the graph name from the session; `locationLabels` does not, so history rows show the Source's title and ref for the Source page and every canvas alike (A-12). Nothing outside the file calls `WorkspaceToolbar.` statics |
| History rows | `WorkspaceToolbar.swift` `HistoryJumpMenuHost.rows(for:navigation:)` (called from `body` and `handleKeyDown`), private `HistoryJumpRow` (`init(index:location:)`, `leafLabel`) | Calls `locationLabels(for:)`; no session. `HistoryJumpMenuHost` sits in `WorkspaceContent`, under `WorkspaceView`'s `.environment(session)` |
| Canvas label | `EvidenceGraphView.swift` `accessibilityGraphLabel`; `EvidenceGraphModel.connectArmedHint`, `armedHint(for:)` (both `LocalizedStringResource`) | Nothing armed → `L10n.Workspace.evidenceGraphTitle` ("Evidence graph") |
| A-14 copy | `macos/App/Platform/L10n.swift` `L10n.Promote.mapRestHint`, `.graphUnavailable`, `.graphUnavailableRepropose`; `L10n.EvidenceGraph` `addPropertyUnavailable` (`evidenceGraph.subject.addPropertyUnavailable`); `L10n.CitationComposer.pairPickBoth` | Say "this Source" |
| A-20 copy | `L10n.swift` `promote.page.promoting`, `citationComposer.pageCount`, `.artifactCitationCount`, `.citationStepperCount` | No variations. The Promote toast and row-summary keys gained variations in S13-23 (`promote.toast.promotedBody` is the `variations.plural` shape, which must print its number, so not for a bare noun; `promote.page.rowSummary` the `substitutions` shape) |
| Noun pick | `L10n.swift` `L10n.DeleteImpact.noun(_:count:)` (`switch` over 19 kinds, `count == 1 ? pair.0 : pair.1`, `default` "item"), `viaFallback(kind:total:)` (`total == 1 ? viaUnknownOne : viaUnknownOther`) | English-only plural choice in Swift; `narrative` and `reconciliation_claim` are not Go kinds; Go's `identity_claim` has no noun |
| Go kinds | `core/database/catalogmodel/model.go` `Kind` constants, one per `catalogmodel.Tables` entry (17: `source`, `artifact`, `citation`, `observation`, `subject`, `source_type`, `metadata_field`, `source_credibility_grade`, `subject_type`, `property`, `property_term`, `file`, `user`, `project`, `canonical_entity`, `identity_claim`, `evidence_graph`) | No Swift parity test |
| Parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (`swiftL10n`, `t.Run("every seeded term has wording", …)`) | The shape for an L10n parity subtest |
| Plural test | `macos/ProvenenciaTests/L10nPluralTests.swift` `everyCatalogPluralPicksAFormPerCount` | Walks the compiled `Localizable.stringsdict`; every plural must differ for 1 and 2 and contain the digit |
| Catalog script | `scripts/check-localizable-xcstrings.py` `main`, module `CATALOG` (used in `main` and its final `relative_to(ROOT)` message) | No count check; fixed catalog path |
| Open | `macos/App/Platform/ProjectFiles.swift` `openObject(projectDir:relPath:)` (looks up the app with `openingApplication(for:)`, `.missing` when none), `ObjectRefusal` (`.missing`, `.unsafe`) | Ignores `NSWorkspace.open`'s result. Callers each `switch` and map `.unsafe` to `L10n.ObjectOpen.refused`: `EvidenceGraphModel.openSelectedFile()` (button at `EvidenceGraphView.swift`), `SourceArtifactsSection.open(_:)` (button in `SourcePageArtifactsView.activatePreview`) |
| Load failure | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `showingLoadFailure(_:)`, `loadFailureMessage`, `loadedCitationRef`, `fetchOpenCitation` (reads the cached listing first, else `store.getCitation`) | Puts the error in `fields.error`, drawn under the Transcription field (`CitationComposerEvidenceColumn.transcriptionColumn`), naming no Citation. Callers: `refreshCitationLines` (id `activeCitationID`), `selectCitationAndLoad`, `loadCitationPreservingAuthoring`, the first-Citation load after an Artifact change (`first.citation.id`). FakeStore's `getCitation` throws its private `StoreBoom.boom` for an id missing from `citationsByID`. `lineRefreshShowsAFailedCitationLoad` reads `fields.error` |
| Citation picker | `CitationComposerEvidenceColumn.swift` `citationStepper`, `sideForm`, `panelForm` (`HStack(alignment: .bottom)` of the two steppers) | The stepper both layouts share |
| Focus | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)`; `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `TypeField` (`case combo`), `metaRow` (`.onAppear { DispatchQueue.main.async { typeFieldFocused = .combo } }`); `SourceIdentitySection.editingType` | The app's focus hand-back; the header hand-rolls one on open and does nothing on close |
| Tests | `PVConfirmTests`, `PVDeleteImpactTests` (`activate` test near line 207), `WorkspaceToolbarBreadcrumbTests` (`@MainActor`; `locationLabelsNameTheSection`), `L10nPluralTests`, `ProjectFilesTests` (`openObjectReportsMissingForSymlinkedObject`, `openObjectReportsMissingForRejectedRelPath`), `EvidenceGraphModelTests` (`openSelectedFile…`), `SourcePageModelTests`, `CitationComposerModelTests` | Cover today's behavior |

**Precedent.** S14-03 ([#652](https://github.com/mendahu/provenencia/pull/652)) moved hand-styled error lines onto `PVErrorText` and announced a failed save from its Callout: copy that "kit owns the look, one announcement per message" shape. `pvContextMenuDismiss` is the shape for `pvSheetClose`. `FocusClaim` (S14-03, `CitationComposerViewerPane.claimDrawingFocus`) is the focus hand-back. The `every seeded term has wording` subtest is the shape for the kind-noun parity subtest.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVConfirmTone.returnConfirms: Bool` *(new)*: `.danger` → false, `.irreversible` → true. `PVConfirmContent` puts `.keyboardShortcut(.defaultAction)` on Cancel when `!returnConfirms`, on confirm otherwise. Cancel keeps `cancelFocused`. The comment says Return cancels | Working rule "A danger confirm's default action is Cancel"; one pure flag a test reads | Dropping the shortcut from both buttons — Return then does nothing, and the rule says it cancels |
| D-2 | `PVConfirmCopy.discardChanges` *(new, static)* = `PVConfirmCopy(title: L10n.string(formDialogDiscardTitle), message: L10n.string(formDialogDiscardMessage), confirm: formDialogDiscardConfirm, cancel: formDialogDiscardKeep)` (the resource initializer, as today's inline copy). `PVSheetChrome` and the Artifact card switch both use it | One copy for one question | A second set of strings for the card |
| D-3 | `PVDiscardGate` *(new)*, a struct in `macos/App/DesignSystem/Components/Confirm/PVDiscardGate.swift` *(new)*, holds the one step a "Discard changes?" question waits on; the owner keeps its own question on screen. `var holdsStep: Bool`; `mutating func hold(_ step: @escaping @MainActor () -> Void)`; `mutating func discard() -> (@MainActor () -> Void)?` hands the step out once; `mutating func keepEditing() -> Bool` drops it and returns whether one was held. Owners: the sheet chrome (D-4), the Artifact cards (D-12), the composer (D-22) | One held-step type for every "ask, then do" in the app; each question stays where its presentation already lives (the sheet's `confirmingDiscard`, the card's `pendingEditSwitch`, the composer's `presented`) | A gate that also holds the question — the composer's question lives in its `presented` dialog stack, so the gate would hold it a second time |
| D-4 | `PVSheetChrome` keeps `confirmingDiscard` and gains `@State var gate = PVDiscardGate()`. It builds `PVSheetClose` *(new)* each body pass: `struct PVSheetClose { let run: @MainActor (@escaping @MainActor () -> Void) -> Void; @MainActor func callAsFunction(then next: @escaping @MainActor () -> Void = {}) }`. The chrome's `run(next)` switches on `PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard: confirmingDiscard)`: `.ignore` → nothing; `.keepEditing` → `confirmingDiscard = false`, `gate.keepEditing()`; `.confirmDiscard` → `gate.hold(next)`, `confirmingDiscard = true`; `.close` → `close()`, then `Task { @MainActor in next() }`. Escape and the outside click call `run` with an empty step, so Escape while asking keeps editing and drops a row's navigation. Discard: `confirmingDiscard = false`; `let step = gate.discard()`; `close()`; `Task { @MainActor in step?() }`. Keep: `confirmingDiscard = false`, `gate.keepEditing()`. `EnvironmentValues.pvSheetClose` *(new)* carries it to content; the default runs `next` in a `Task` and closes nothing (previews). A new `item` re-presents the sheet with a new chrome and a fresh gate | The dirty rule stays `PVSheetDismiss`; content asks the sheet instead of clearing the binding, so every row in every sheet gains it. Navigation then still meets the place's `WorkspaceLeaveGuard` (§ 4) | Passing `isDirty` down to the rows — a second copy of the rule |
| D-5 | `PVDeleteImpactControls.activate(_:onNavigate:close:)` takes `close: PVSheetClose` and calls `close { onNavigate(location) }`. `PVDeleteImpactGroupList` drops `dismiss` and reads `@Environment(\.pvSheetClose)`. `PVDeleteImpactNotice` drops `onDismiss`; its Done calls `pvSheetClose()`; its call site and previews drop the argument. The `then:` parameter goes. The group list's doc comment says a row closes the sheet through `pvSheetClose`, then navigates. `SourceGraphForm` drops `dismiss`; `keptLinked` gets `.disabled(model.isSaving)` | One close path per sheet; the notice is never dirty, so it closes at once as today; A-13 | Keeping `dismiss:` beside the environment — two ways to close |
| D-6 | `AccessibilityAnnouncement` gains `static func text(title: String?, message: String?) -> String?` (non-empty parts; both → `L10n.DesignSystem.announcementJoin(_:_:)` *(new)*, key `designSystem.announcement.join` "%1$@, %2$@"; none → nil) and `static func priority(danger: Bool) -> NSAccessibilityPriorityLevel` (`.high` / `.medium`). `View.pvAnnounce(_ text: String?, priority:)` *(new, `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift`)* posts through `AccessibilityAnnouncement.post` on appear when non-nil and on change to a different non-nil text. The same text again while the view stays mounted is not spoken again (a second identical save failure is silent; accepted). A toast remounted by `.id(toast)` speaks each new toast. Two posts in one turn are not ordered | One announcement type; a translatable join; the modifier where kit modifiers live | A `PVAnnouncement` type in `Platform/` beside `AccessibilityAnnouncement` — two names for one job; a hard-coded ", " |
| D-7 | `PVCopy` gains `var pvString: String { get }` *(new requirement)*: `LocalizedStringResource` → `L10n.string(self)`, `String` → `self`, `AttributedString` → `String(self.characters)`. `PVCallout` keeps the title as `(any PVCopy)?` beside its `Text`. `PVToast` always applies `pvAnnounce(.text(title:message:), priority: .priority(danger: tone == .danger))`. `PVCallout` and `PVErrorText` take `announces: Bool = false` *(new)* and apply it when true (callout: `.text(title: title?.pvString, message:)`, priority by `tone == .danger`; error text: `.high`). `ConclusionLeadModel` drops its two posts | Working rule "Kit components announce themselves": no call site posts its own. A toast always announces (A-18 fix); a callout or error line opts in because most callouts are static explanations | Announcing every callout — a page's info callouts would talk on every open; announcing the callout's message only — the graph form's callout has a title |
| D-8 | Opted in: the `PVCallout`s `sources.graphForm.error`, `sources.page.artifact.<id>.saveError`, `sources.page.error`, `evidenceGraph.fileOpen.callout`; the `PVErrorText`s `citationComposer.depiction.savePageError` (`savePage()` drops its `else if let error … post` branch; its success branch's `FocusClaim … { post(caption) }` stays, a caption after an action is not a kit error line) and `citationComposer.citationLoadError` (D-16). The save-page line shows any `fields.error`, so it speaks when the depiction box appears over an earlier error (accepted). The cut-failed caption's `.onChange … post` in `CitationComposerAuthoringBox` becomes `.pvAnnounce(model.depictionCaptionState == .cutFailed ? L10n.string(model.depictionCaptionState.text) : nil, priority: .medium)`. Toasts now also speak on onboarding (`OnboardingView`) and for a saved bridge (`CitationComposerSubjectGroups`) (accepted) | The scope entry's list, plus every kit error line and on-change post in files this PR edits; each appears after an action with focus elsewhere | The Source page `loadError` callout — it replaces the page, so focus is already there |
| D-9 | `PVSectionHeader` puts `.accessibilityAddTraits(.isHeader)` on `title.pvText` | A-21; precedent `ConclusionDetailPage.swift` | Per call site |
| D-10 | `EnvironmentValues.pvFieldLabel: LocalizedStringResource?` *(new, `PVField.swift`)*, default nil; `PVField` always sets it on `content`, to its `label` (nil when it has none, so an outer field's label never reaches an inner field). `PVFieldLabel.resolve(own: String?, field: String?) -> String?` *(new, `PVField.swift`)* is own → field → nil. `PVInput` applies `.accessibilityLabel(Text(verbatim:))` once on its outermost view, so every render path (read-only, `PVInputFocus.install`, external and owned focus) gets it, and `PVTextArea` on its outermost view, when `resolve(own: nil, field: fieldLabel.map(L10n.string))` is non-nil. `PVSelect.accessibilityLabelText` becomes `Text(verbatim: PVFieldLabel.resolve(own: accessibilityLabelResource.map(L10n.string) ?? accessibilitySpokenLabel, field: fieldLabel.map(L10n.string)) ?? triggerText)`. Several kit controls inside one `PVField` all take its label (accepted). `PVComboBox` is unchanged. The label `Text` in `PVField` stays visible to VoiceOver | One label-order rule for the three kit controls; fixes every form (B-18's "better" fix); a select's own label still wins | `accessibilityLabeledPair` — on macOS its effect when the content has its own label is not testable here; hiding the field's label `Text` — a field whose content is not a kit control would lose its only name |
| D-11 | `CitationComposerEvidenceColumn.labelField` drops its `.accessibilityLabel` | Duplicate of D-10 | — |
| D-12 | `SourceArtifactsSection.beginEdit(_ id:)` becomes the asking entry; today's body moves to `private func openEdit(_ id:)` *(new)*. `beginEdit`: when `editingArtifactID == id` or `savingID != nil`, nothing; when another card is in edit and `fieldsDirty(current)`, `editGate.hold { [weak self] in self?.openEdit(id) }` and `pendingEditSwitch = PendingEditSwitch()` *(new; `struct PendingEditSwitch: Identifiable { let id = UUID() }`, as the composer's `PendingLeave`)*; else `openEdit(id)`. `confirmEditSwitch()` *(new)*: `pendingEditSwitch = nil`, then `editGate.discard()?()`. `cancelEditSwitch()` *(new)*: `pendingEditSwitch = nil`, `_ = editGate.keepEditing()`. The pencil and **Link to a graph** keep calling `beginEdit`. `SourcePageView` adds `.pvConfirm(item: Binding(get: { model.artifacts.pendingEditSwitch }, set: { if $0 == nil { model.artifacts.cancelEditSwitch() } }), copy: { _ in .discardChanges }, tone: .danger, accessibilityIdentifierPrefix: "sources.page.artifact.discard", onConfirm: { model.artifacts.confirmEditSwitch() }) { _ in EmptyView() }`, as its sibling confirms build their bindings. Return keeps editing (D-1). `SourcePageModel.acceptsArtifactPageDrop` adds `&& artifacts.pendingEditSwitch == nil` | One entry point that asks; the target lives once, in the held step; the question is one value, as `pendingDelete` and `PendingLeave` are | Keeping `beginEdit` as the silent path beside a new `requestEdit` — the next caller would drop an edit |
| D-13 | `macos/App/Features/Workspace/WorkspaceLocationLabels.swift` *(new)* holds `struct LocationNames` (moved from `WorkspaceToolbar.BreadcrumbNames`; closures `@MainActor (String) -> String?`; `static var none`; `@MainActor init(session: WorkspaceSession)` reading `session.source(id:)?.title` and `session.evidenceGraph(id:)?.name`) and `enum WorkspaceLocationLabels` (`@MainActor`): `labels(for:names:) -> (leaf: String?, section: String?)` (moved from `WorkspaceToolbar.locationLabels`; `names` defaults to `.none`); `graphLabel(graphID:names:) -> String` (`names.graphName` non-empty → it, else `L10n.string(L10n.Workspace.evidenceGraphTitle)`); `surfaceLabel(for:) -> String?` (Promote: `promoteAll` → `L10n.string(L10n.EvidenceGraph.promoteAll)`, else `L10n.Promote.breadcrumb(ref: ref ?? scope)`; composer: `L10n.CitationComposer.breadcrumbCitationFor(scope:)`; canvas and others: nil; `scope` = non-empty `title`, else non-empty `ref`, else the ellipsis U+2026, as today). For a `.graph`, `.citationComposer` or `.promote` location with a `graphId`, `labels` returns leaf `graphLabel`, section `surfaceLabel`; other locations as today. `WorkspaceToolbar.breadcrumbItems(for:names:goTo:)` takes `LocationNames` and builds its graph, composer and Promote crumbs from `graphLabel` and `surfaceLabel`; its entity branch calls `labels(for:)` with `.none` | One name lookup and one label home for crumbs, history and canvas (working rule "No graph name in the canvas header": "the breadcrumbs' own name lookup"); the canvas model does not depend on a toolbar view | Statics on the `WorkspaceToolbar` view — the canvas model would reach into a toolbar type |
| D-14 | `HistoryJumpMenuHost` reads `@Environment(WorkspaceSession.self)`; `rows(for:navigation:names:)` (main actor) is called from `body` and `handleKeyDown`, both with `LocationNames(session: session)`. `HistoryJumpRow` takes `init(index:location:leafLabel:)` and stores `leafLabel`: `labels.leaf` joined with `labels.section` through `L10n.Conclusions.pageSectionLeaf` ("%1$@ · %2$@") when both exist, else `labels.leaf`. The ref column stays the location's `ref` (the composer row shows its Citation ref there too; accepted) | User, 2026-11-01: graph name, then surface; the existing Back-menu leaf format | Source title · graph name (user declined: longer) |
| D-15 | `EvidenceGraphModel.canvasAccessibilityLabel: String` *(new)*: connect armed → `L10n.string(connectArmedHint)`; a kind armed → `L10n.string(armedHint(for:))`; else, with `graphLabel(graphID:names: LocationNames(session:))` from a known name, `L10n.EvidenceGraph.canvasAccessibility(graph:)` *(new)* = "Evidence graph, %@"; with no name known, `L10n.string(L10n.Workspace.evidenceGraphTitle)`. `EvidenceGraphView.accessibilityGraphLabel` becomes `Text(verbatim: model.canvasAccessibilityLabel)` | A-12's fix; testable on the model | Computing it in the view — not testable |
| D-16 | `CitationComposerModel.citationLoadError: String?` *(new)* replaces `loadFailureMessage`. `citationRef(for id: String) -> String?` *(new)*: the `listedCitations` row's ref, else `loadedCitationRef` when `id == activeCitationID` and it is non-empty; `activeCitationRef` and `announceIdentityChange()` call it. `showingLoadFailure(citationID: String, _ work:)` *(new parameter)*: `refreshCitationLines` passes its `id`; the first-Citation load passes `first.citation.id`; `selectCitationAndLoad` and `loadCitationPreservingAuthoring` pass the unwrapped `id` (their blank case returns before the call). On failure (not cancelled) it sets `citationLoadError` to `L10n.CitationComposer.citationLoadFailed(ref:reason:)` *(new)* when `citationRef(for:)` has a ref, else `citationLoadFailedUnnamed(reason:)` *(new)*; on success it sets nil. `openBlankCitation()` and the start of `startIdentitySwitch` set it to nil (a new Citation, an Artifact change, a delete). It no longer reads or writes `fields.error`. `citationLoadErrorLine` *(new private view)* = `if let error = model.citationLoadError { PVErrorText(error, announces: true).fixedSize(horizontal: false, vertical: true).accessibilityIdentifier("citationComposer.citationLoadError") }` (size `.micro`, the field-level line). `sideForm` puts it directly under `citationStepper` inside the steppers' `VStack`; `panelForm` puts it under the steppers' `HStack`, full width, so the bottom-aligned steppers do not move | User, 2026-11-01: under the Citation picker, no brief; one id → ref rule | A callout above the steppers (user declined); a third ref lookup |
| D-17 | `ObjectRefusal.notOpened` *(new case)*. `ProjectFiles.openObject(projectDir:relPath:launch:)` becomes `@MainActor async`, with `launch: @MainActor (URL) async throws -> Void = ProjectFiles.launch`. `ProjectFiles.launch(_:)` *(new, `@MainActor static`)* holds today's app lookup (`openingApplication(for:)`; none → `throw ObjectRefusal.missing`) and awaits `NSWorkspace.shared.open([url], withApplicationAt: app, configuration: NSWorkspace.OpenConfiguration())`. `openObject` returns a thrown `ObjectRefusal` as it is and any other throw as `.notOpened`. `L10n.ObjectOpen.message(for: ObjectRefusal, filename: String, missing: String) -> String` *(new)*: `.unsafe` → `refused(filename:)`, `.notOpened` → `notOpened(filename:)` *(new)*, `.missing` → `missing`. `EvidenceGraphModel.openSelectedFile()` and `SourceArtifactsSection.open(_:)` become `async`, set their error to `L10n.ObjectOpen.message(for:filename:missing:)` with their own missing sentence, and their buttons call them in `Task { await … }`. The existing `openObject` tests and the models' open tests `await` | The open failure surfaces in the callouts that already show refusals; one mapping for both callers; one test seam that also covers the app lookup, so tests do not depend on the host's apps | A `launch` parameter on each model method — two test-only parameters; a seam after the lookup — a CI runner with no PDF app returns `.missing` |
| D-18 | `SourcePageIdentityHeader.TypeField` gains `case edit`. The Type pencil (`PVIconButton`, a `Button`) gets `.focused($typeFieldFocused, equals: .edit)`. `metaRow` gets `.onChange(of: model.identity.editingType) { was, now in if was, !now { FocusClaim.claim($typeFieldFocused, .edit) } }` on the `HStack`. The combo's `.onAppear` becomes `FocusClaim.claim($typeFieldFocused, .combo)` | `FocusClaim` is the app's focus hand-back, written for a control the same commit inserts | A hand-rolled `DispatchQueue.main.async` — a second mechanism |
| D-19 | `L10n.DeleteImpact.kindNouns: [String: LocalizedStringResource]` *(new, a declared table like `viaFormats`)* holds one plural key per Go `Kind`, `deleteImpact.kind.<kind>`, whose English value is `%#@noun@` with substitution `noun` (`argNum` 1, `formatSpecifier` `lld`; `one` / `other` the bare noun, "citation" / "citations"), as `deleteImpact.via.unknown`'s substitution shape; a top-level `variations.plural` with no number is a catalog compile error ("Plural variation requires referencing the number in the string"; STEP-9 probe, round 0), and a substitution value need not print its number (`xcstringstool compile` exit 0; resolves "citation" for 1, "citations" for 0, 2, 5; 2026-11-01); `identity_claim` reads "identity claim" / "identity claims". Each resource's `defaultValue` is `%#@noun@`, matching the catalog value as `add-localized-string` step 5 says. `noun(_:count:)` keeps its signature and its one-argument call `L10n.format(kindNouns[kind] ?? itemNoun, count)` (missing → `deleteImpact.kind.item`); the count reaches `noun` through `argNum` 1 only. `narrative` and `reconciliation_claim` go (not Go kinds; the preview that used `narrative` reads "item"). `viaFallback` reads one key `deleteImpact.via.unknown` *(new)* with value "%1$lld %2$@ %#@verb@ this" and substitution `verb` (`argNum` 1, `formatSpecifier` `lld`; `one` "references", `other` "reference"), as `promote.page.rowSummary`'s substitution shape. The 38 `.one` / `.other` keys, their `static let`s, `viaUnknownOne` and `viaUnknownOther` go. `PluralShape` gains `printsNumber: Bool` *(new)*: true when some text (the format or a form) holds a number specifier other than a `#@…@` reference at a count position. `everyCatalogPluralPicksAFormPerCount` runs its digit expectations only when `printsNumber`, keeps "one ≠ several" and the leftover-placeholder check for every key, and drops the `deleteImpact.kind.` prefix test. The compiled kind key's format is `%1$#@noun@` (scratch probe), so `printsNumber` is false for it and true for `deleteImpact.via.unknown` (`%1$lld`) | A-20: the plural rule is the catalog's; a translation with more forms works. The table is held to Go's kinds by D-20 | Counted noun keys ("%lld citation") — `alsoAffects` and the confirm title need the bare noun. A top-level `variations.plural` per kind — the catalog compiler rejects it. Separate one / other keys per kind (the compiler's advice) — keeps English's two-form rule in Swift (A-20) |
| D-20 | `TestSwiftVocabularyMatchesGo` gains `t.Run("every delete-impact kind has a noun", …)` *(new)*: the distinct `Table.Kind` values over `catalogmodel.Tables`, plus `item`, equal the `deleteImpact.kind.(\w+)` keys in `swiftL10n`, as sets | Known trap: a string vocabulary that crosses the client boundary with a fallback and no parity test | No test — the next Go kind silently reads "item" |
| D-21 | `check-localizable-xcstrings.py`: `main(catalog: Path = CATALOG)` reads `sys.argv[1]` when given; its success message prints the given path as given. `uninflected_counts(strings) -> list[str]` *(new)*: an English value with `%lld` or `%N$lld` followed by a word of letters, in an entry with no `variations` or `substitutions`, is a problem unless the word is in `NON_INFLECTING = {"of", "more", "percent", "set", "up", "down", "cannot", "to", "unselected", "already", "px"}`. Every other check still runs on the given catalog | A-20's script check; the shipped catalog's 21 non-plural `%lld` values were each read (§ 6 Edges) | A key allow-list — every new "of" format would need an entry |
| D-22 | `CitationComposerModel.stepContinuation` becomes `leaveGate: PVDiscardGate` (user, 2026-11-01). `selectCitation`, `requestStep` and `openAuthoring` call `leaveGate.hold { … }` where they set `stepContinuation`, and set `pendingLeave` as today. `discardLeaveChanges()`: `let next = leaveGate.discard()` in place of reading and clearing `stepContinuation`; the rest as today (`keepDraft`, `next()` else `resumeHeldNavigation()`). `keepEditingAfterLeave()`: `let stepping = leaveGate.keepEditing()` in place of its two `stepContinuation` lines | One held-step type across sheets, cards and the composer (review M-13); the composer's `presented` stack keeps its question | Leaving the composer's copy — the user chose to move it |

**Rules this plan bends.** The map scope entry says the plan splits along its parts past about ten steps; this plan has 13. User, 2026-11-01 ("do it all in one PR"). STEP-13 records it in the scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVConfirmTone.returnConfirms` *(new)* | property | `PVConfirmTone.buttonRole` | Extends the same enum |
| `PVConfirmCopy.discardChanges` *(new)* | static | inline copy in `PVSheetChrome` | Lifts it so the card reuses it |
| `PVDiscardGate` *(new)* | struct | `CitationComposerModel.stepContinuation`; `PVSheetDismiss`; `WorkspaceLeaveGuard` | Replaces the composer's `stepContinuation` (D-22) and serves the sheet and the cards. `PVSheetDismiss` stays the rule. `WorkspaceLeaveGuard` holds a `go(to:)` for a whole place; the gate holds a step for one owner's question, and the step's `go(to:)` still meets the place's guard |
| `PVSheetClose`, `EnvironmentValues.pvSheetClose` *(new)* | struct / env | `pvContextMenuDismiss` | Same environment shape |
| `AccessibilityAnnouncement.text`, `.priority` *(new)* | static funcs | `AccessibilityAnnouncement.post` | Same type |
| `View.pvAnnounce` *(new)* | modifier | `pvOutsideClick` (kit modifier) | Kit modifier folder |
| `designSystem.announcement.join` *(new)* | key | `conclusions.a11y.list` | A kit-owned join |
| `PVCopy.pvString` *(new)* | requirement | `PVCopy.pvText` | Same protocol, three conformers |
| `PVCallout(announces:)`, `PVErrorText(announces:)` *(new param)* | param | — | Opt-in per working rule |
| `EnvironmentValues.pvFieldLabel`, `PVFieldLabel.resolve` *(new)* | env / func | `PVSelect.accessibilityLabelText` | The select's order calls the resolver |
| `SourceArtifactsSection.openEdit` (private, today's `beginEdit` body), `confirmEditSwitch`, `cancelEditSwitch`, `editGate`, `PendingEditSwitch` *(new)* | func / struct | `pendingDelete` in `SourceNotesSection` / `SourceMetadataSection`; `CitationComposerModel.PendingLeave` | Same single-value question shape, on `PVDiscardGate` |
| `WorkspaceLocationLabels.swift` *(new)*: `LocationNames` (moved), `WorkspaceLocationLabels.labels` (moved), `graphLabel`, `surfaceLabel` *(new)* | file / struct / funcs | `WorkspaceToolbar.BreadcrumbNames`, `locationLabels`, the inline labels in `breadcrumbItems` | Moves them out of the view so the history host and canvas model share them |
| `EvidenceGraphModel.canvasAccessibilityLabel` *(new)* | property | `EvidenceGraphView.accessibilityGraphLabel` | Moves the rule onto the model |
| `L10n.EvidenceGraph.canvasAccessibility(graph:)`, `L10n.ObjectOpen.notOpened`, `L10n.ObjectOpen.message(for:filename:missing:)`, `L10n.CitationComposer.citationLoadFailed`, `.citationLoadFailedUnnamed`, `L10n.DesignSystem.announcementJoin` *(new)* | L10n | `L10n.ObjectOpen.refused` | New sentences; one refusal mapping |
| `L10n.DeleteImpact.kindNouns`, `deleteImpact.kind.<kind>` (17 + `item`), `deleteImpact.via.unknown` *(new)* | table / keys | `viaFormats`; the 38 `.one` / `.other` keys and `viaUnknownOne` / `Other` | Replace them |
| `every delete-impact kind has a noun` *(new)* | Go subtest | `every seeded term has wording` | Same test function |
| `ObjectRefusal.notOpened`, `ProjectFiles.launch` *(new)* | case / func | `ObjectRefusal.missing` | Same enum |
| `CitationComposerModel.citationLoadError`, `citationLoadErrorLine`, `citationRef(for:)` *(new)* | property / view / func | `loadFailureMessage`; `activeCitationRef` | Replaces the first; lifts the second's rule so three callers share it |
| `TypeField.edit` *(new case)* | case | `TypeField.combo` | Same enum |
| `uninflected_counts`, `NON_INFLECTING` *(new)* | script func | `unnumbered_arguments` | Same script, same shape |
| `AccessibilityAnnouncementTests.swift`, `PVDiscardGateTests.swift`, `PVFieldLabelTests.swift` *(new)* | test files | `PVConfirmTests.swift` | One suite per kit type |

**Builds on.** `PVSheetDismiss` (the one dismiss rule), `AccessibilityAnnouncement.post` (the one post path), `FocusClaim` (the focus hand-back), `PVSelect`'s label order, `pageSectionLeaf` (the Back-menu leaf format), `L10n.format` plural resolution, `PVErrorText` (S14-03), `TestSwiftVocabularyMatchesGo`, `add-ui-component`'s kit-first rule. The sheet's gate asks for a sheet's own form; the place's `WorkspaceLeaveGuard` still holds the navigation that follows (a sheet over a dirty composer asks the sheet first, then the composer).
**Deletes.** `PVDeleteImpactControls.activate`'s `dismiss:` and `then:`; `PVDeleteImpactGroupList.dismiss`; `PVDeleteImpactNotice.onDismiss` and the argument at its call site and previews; `SourceGraphForm.dismiss`; `PVSheetChrome`'s inline discard copy; `CitationComposerModel.stepContinuation`; `announceIdentityChange()`'s own ref lookup; the composer's cut-failed `.onChange … post`; `WorkspaceToolbar.BreadcrumbNames` and `WorkspaceToolbar.locationLabels` (moved); `HistoryJumpRow.leafLabel`'s computed body; `EvidenceGraphView.accessibilityGraphLabel`'s rule; `CitationComposerModel.loadFailureMessage`; `labelField`'s `.accessibilityLabel`; `ConclusionLeadModel`'s two posts; `savePage()`'s error post; the hand-rolled `DispatchQueue.main.async` focus in `SourcePageIdentityHeader`; the two `switch`es over `ObjectRefusal` in the models; `L10n.DeleteImpact.noun*` `static let`s (38), the `switch` in `noun`, and their catalog keys; `viaUnknownOne`, `viaUnknownOther` and their keys; the `narrative` and `reconciliation_claim` nouns.
**Next adds.** S14-20 opts Promote's error callouts in with `announces: true` and moves `PromoteAlignmentRow` onto `FocusClaim`. A new sheet row that navigates reads `pvSheetClose`. Notes and Metadata rows: `beginEdit` asks through `PVDiscardGate` and one `Pending…` value. A new kit text control inside `PVField` calls `PVFieldLabel.resolve`. A new Go `Kind` fails D-20 until it has a noun. A new count key with a non-inflecting word adds it to `NON_INFLECTING`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A sheet row closes through the dirty rule, then navigates | `PVSheetChrome`, `WorkspaceNavigation.go(to:)` | The pending step is dropped on Keep editing and Escape (AC-2.7, AC-2.8) and handed out once (AC-2.6); navigation runs after the sheet clears, as today, and meets the destination place's leave guard as any `go(to:)` does |
| A card switch waits on a question | `SourceArtifactsSection` | If the target card is deleted while the question shows, `beginEdit` returns at its `guard` after the old card's edit is cancelled; a file drop is refused while it shows (AC-2.9) |
| History rows read live graph names | `HistoryJumpMenuHost` | A deleted graph falls back to "Evidence graph" (AC-5.2) |
| Composer load error line | `citationLoadErrorLine` | Cleared by the next successful load (AC-7.2); an unknown ref gets the unnamed copy (AC-7.3) |
| `openObject` is async | the two open buttons | A second click while the first launch runs starts a second open, as two clicks do today |

### Scale

None — no write, list or loop changes. The history menu builds at most the history depth of rows (`NavigationHistoryStore`'s cap), each one lookup in the session's graph list.

### Edges

- **Assumptions.** `PVSheetDismiss` order (running, asking, dirty) is unchanged; `PVDiscardGate` only adds the held step (T-4). `L10n.format` resolves a substitution plural whose value omits the number (a scratch-catalog probe resolved it, 2026-11-01; STEP-9 probes it again in the app, T-12); if it does not, Stop and report; do not switch to separate one / other keys. `check-localizable-xcstrings.py` needs no change for the `%#@noun@` shape: `uninflected_counts` skips entries with `substitutions`, and a lone `#@` specifier is not an unnumbered argument.
- **Neighbors.** The irreversible confirm keeps Return (AC-1.2). A clean form closes and navigates at once (AC-2.4). A non-graph location's label is unchanged (AC-5.5), and the crumbs are unchanged (AC-5.6). A select with an explicit label keeps it (AC-4.2). An unknown kind reads "item" (AC-6.3).
- **Removed inputs.** A graph deleted while it is in history reads "Evidence graph" (AC-5.2).
- **Real data.** The shipped catalog's `%lld <word>` values with no variations, after this PR: `artifactChecklist.count`, `designSystem.select.optionPosition`, `properties.inspector.bindCount`, `artifactViewer.findMatchOf`, `a11y.nameValue.part`, `a11y.nameValue.part.moved`, `conclusions.outcome.outvotedBy` ("of"); `deleteImpact.overflow`, `deleteImpact.summary.more`, `conclusions.place.containsMore`, `promote.page.singleSummary` ("more"); `artifactViewer.zoomPercent` ("percent"); `artifactViewer.pageSet` ("set"); `a11y.nameValue.part.moveUp` / `moveDown` ("up" / "down"); `nameValue.part.error.emptyValue` ("cannot"); `citationComposer.depictionSavePage` ("to"); `promote.done.unselected` ("unselected"); `promote.section.countPromoted` ("already"); `citationComposer.imagePixels` ("px"). The four A-20 keys move to variations. T-14 runs the script on the shipped catalog. Go's kinds: 17 (§ 4).
- **Untrusted size.** None — no data the app did not write is read.

### Schema

None — no catalog change.

### Go core

None — no Go code change; one subtest in `api/ffi/swift_vocabulary_test.go` (D-20).

### FFI

None — no RPC changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Confirm default | `PVConfirmContent` | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` | D-1, D-2 |
| Discard gate | `PVDiscardGate` | `macos/App/DesignSystem/Components/Confirm/PVDiscardGate.swift` *(new)* | D-3 |
| Sheet close | `PVSheetChrome`, `PVSheetClose` | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` | D-4 |
| Impact rows | `PVDeleteImpactGroupList`, `PVDeleteImpactNotice` | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` | D-5 |
| Announcements | `pvAnnounce`, `PVToast`, `PVCallout`, `PVErrorText` | `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` *(new)*, `macos/App/DesignSystem/Components/Toast/`, `macos/App/DesignSystem/Components/Callout/`, `macos/App/DesignSystem/Components/ErrorText/` | D-6, D-7, D-8 |
| Heading | `PVSectionHeader` | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` | D-9 |
| Field name | `PVField`, `PVInput`, `PVTextArea`, `PVSelect` | `macos/App/DesignSystem/Components/Field/`, `macos/App/DesignSystem/Components/Input/`, `macos/App/DesignSystem/Components/TextArea/`, `macos/App/DesignSystem/Components/Select/` | D-10 |
| Card switch | `.pvConfirm` | `macos/App/Features/Sources/SourcePageView.swift` | D-12 |
| Labels | `WorkspaceLocationLabels` | `macos/App/Features/Workspace/WorkspaceLocationLabels.swift` *(new)* | D-13, D-14 |
| Load error | `PVErrorText` | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` | D-16 |

States: no new empty or loading state. The composer at Wide draws as at Regular (no layout change). Accessibility: as US-3, US-4, US-5.

### Strings

L10n keys *(new)*:
- `designSystem.announcement.join` = "%1$@, %2$@" (`L10n.DesignSystem.announcementJoin(_:_:)`; comment: VoiceOver join of a toast's or callout's title and message)
- `evidenceGraph.canvas.accessibilityLabel` = "Evidence graph, %@" (`L10n.EvidenceGraph.canvasAccessibility(graph:)`)
- `objectOpen.notOpened.message` = "macOS couldn't open “%@” in its app." (`L10n.ObjectOpen.notOpened(filename:)`)
- `citationComposer.citationLoadFailed` = "Couldn't open %1$@. %2$@" (`L10n.CitationComposer.citationLoadFailed(ref:reason:)`)
- `citationComposer.citationLoadFailedUnnamed` = "Couldn't open this citation. %@" (`L10n.CitationComposer.citationLoadFailedUnnamed(reason:)`)
- `deleteImpact.kind.<kind>` for the 17 Go kinds and `item`: value `%#@noun@`, substitution `noun` (`argNum` 1, `lld`) `one` / `other` = today's `.one` / `.other` values; `identity_claim` = "identity claim" / "identity claims". The forms carry no `%arg` (unlike `promote.page.rowSummary`'s). Every entry has this shape, with its own comment and nouns:

```json
"deleteImpact.kind.citation": {
  "comment": "Impact kind noun, bare, by count",
  "extractionState": "manual",
  "localizations": {
    "en": {
      "stringUnit": { "state": "translated", "value": "%#@noun@" },
      "substitutions": {
        "noun": {
          "argNum": 1,
          "formatSpecifier": "lld",
          "variations": {
            "plural": {
              "one": { "stringUnit": { "state": "translated", "value": "citation" } },
              "other": { "stringUnit": { "state": "translated", "value": "citations" } }
            }
          }
        }
      }
    }
  }
}
```

- `deleteImpact.via.unknown` = "%1$lld %2$@ %#@verb@ this", substitution `verb`: `one` "references", `other` "reference"

Changed values (same keys):
- `promote.page.mapRestHint` → "Adds a row for every other person, event and place on this Evidence graph"
- `promote.error.graphUnavailable` → "Provenencia couldn't read this Evidence graph, so it can't list the rows to promote. Go back to the graph and open Promote again."
- `promote.error.graphUnavailableRepropose` → "Provenencia couldn't read this Evidence graph to propose the rows again. Your rows and decisions are kept."
- `evidenceGraph.subject.addPropertyUnavailable` → "Add property, unavailable — this evidence graph has no artifacts" (matches the canvas callout `This evidence graph has no artifacts`)
- `citationComposer.pairPickBoth` → "Pick two subjects on this Evidence graph. The pair decides what the bridge can say"
- Comments on those five say "graph", not "Source"

Gain plural variations (default values stay the `other` form): `promote.page.promoting` ("%lld subject on this graph" / "%lld subjects on this graph"), `citationComposer.pageCount` ("%lld page" / "%lld pages"), `citationComposer.artifactCitationCount` ("%lld citation" / "%lld citations"), `citationComposer.citationStepperCount` ("%lld observation" / "%lld observations").

Removed: the 38 `deleteImpact.kind.<kind>.one` / `.other` keys (including `narrative` and `reconciliation_claim`); `deleteImpact.via.unknown.one`, `deleteImpact.via.unknown.other`.

Old-scope copy (A-14 changes what the copy says the action covers): the five keys above. `promote.page.column.source` and `promote.row.promotedEarlier` keep "this Source" (§ 3 Out).

## 7. Steps

**STEP-1 — Return cancels a danger confirm**
Files: `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift`, `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVConfirmTests.swift`. Skill: `add-ui-component`.
Do: D-1, D-2. README: replace the sentence that begins "**Focus starts on cancel**" (it wraps across two lines in the Confirmations section) with "A danger confirm's default action is Cancel: Return keeps the record, and the destructive button has no Return shortcut. Focus starts on Cancel. An irreversible confirm keeps Return on its confirm button."
Check: T-1, T-2 pass.

**STEP-2 — Kit components announce themselves**
Files: `macos/App/Platform/AccessibilityAnnouncement.swift`, `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` *(new)*, `macos/App/DesignSystem/Components/Copy/PVCopy.swift`, `macos/App/DesignSystem/Components/Toast/PVToast.swift`, `macos/App/DesignSystem/Components/Callout/PVCallout.swift`, `macos/App/DesignSystem/Components/ErrorText/PVErrorText.swift`, `macos/App/Features/Conclusions/ConclusionLeadModel.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/AccessibilityAnnouncementTests.swift` *(new)*. Skill: `add-ui-component`.
Do: D-6, D-7, D-8. README rows: the Toast row gains "(announces its title and message to VoiceOver when it appears)" after its path; Callout "`announces: true` speaks the title and message when it appears or changes (error callouts)"; ErrorText "`announces: true` speaks the line when it appears or changes"; a new Announce row after the ErrorText row: "`macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` *(new)* (`pvAnnounce(_:priority:)` speaks a text when it appears or changes; toasts, callouts and error lines use it)".
Check: T-3 passes; `rg -n "AccessibilityAnnouncement.post" macos/App/Features/Conclusions/ConclusionLeadModel.swift` has no hit; `rg -n "post\(error\)|cutFailed \{ AccessibilityAnnouncement" macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` has no hit.

**STEP-3 — Section headers are headings**
Files: `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `macos/App/DesignSystem/README.md` (SectionHeader row: "the title is a VoiceOver heading").
Do: D-9.
Check: builds; manual check 4.

**STEP-4 — Fields name their content**
Files: `macos/App/DesignSystem/Components/Field/PVField.swift`, `macos/App/DesignSystem/Components/Input/PVInput.swift`, `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/App/DesignSystem/Components/Select/PVSelect.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `macos/App/DesignSystem/README.md` (Field row: "gives its label to the `PVInput`, `PVTextArea` or `PVSelect` inside it through `pvFieldLabel`; a control's own label wins (`PVFieldLabel.resolve`)"), `macos/ProvenenciaTests/PVFieldLabelTests.swift` *(new)*. Skill: `add-ui-component`.
Do: D-10, D-11.
Check: T-10 passes.

**STEP-5 — A sheet's links ask before a dirty form closes**
Files: `macos/App/DesignSystem/Components/Confirm/PVDiscardGate.swift` *(new)*, `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`, `macos/App/DesignSystem/README.md` (sheet section: "Content closes its sheet through `@Environment(\.pvSheetClose)`: a dirty sheet asks first, then runs the step it was given. A place's `WorkspaceLeaveGuard` still holds the navigation that follows."), `macos/ProvenenciaTests/PVDiscardGateTests.swift` *(new)*, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`.
Do: D-3, D-4, D-5.
Check: T-4, T-5 pass; `rg -n "dismiss: \{ model = nil \}|onDismiss:" macos/App/Features/Sources/SourceGraphFormDialog.swift macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` has no hit.

**STEP-6 — An Artifact card asks before another card drops its edit**
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Do: D-12.
Check: T-6 passes.

**STEP-7 — The canvas and history name the graph**
Files: `macos/App/Features/Workspace/WorkspaceLocationLabels.swift` *(new)*, `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-localized-string`.
Do: D-13, D-14, D-15.
Check: T-7, T-8 pass; every existing `WorkspaceToolbarBreadcrumbTests` test passes with only the `BreadcrumbNames` → `LocationNames` and `WorkspaceToolbar.locationLabels` → `WorkspaceLocationLabels.labels` renames.

**STEP-8 — Copy names the Evidence graph**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: § 6 Strings, changed values.
Check: T-9 passes.

**STEP-9 — Counts take plurals, and the script holds them**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` (the `narrative` preview), `macos/ProvenenciaTests/L10nPluralTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`, `api/ffi/swift_vocabulary_test.go`, `scripts/check-localizable-xcstrings.py`. Skill: `add-localized-string`.
Starting point (resume, 2026-11-01): the branch at `bc8f22ff` holds STEP-1 – STEP-6 committed, and STEP-7, STEP-8 and STEP-9 uncommitted in the working tree; STEP-9's 18 `deleteImpact.kind.<kind>` entries are in the rejected top-level `variations.plural` shape, and `kindNouns`, `noun`, `viaFallback`, the D-20 subtest and D-21 are written. Commit STEP-7 and STEP-8 first, after a clean `build-for-testing`. Then convert the 18 entries in place, `citation` first.
Do: convert `deleteImpact.kind.citation` to the D-19 shape and run T-12's `citation` case (the probe: a `%#@noun@` substitution whose forms carry no number resolves, and the build has no catalog error or warning). Then D-19, D-20, D-21, and § 6 Strings plurals. The script's docstring gains a paragraph for the count check.
Check: T-11, T-12, T-13, T-14, T-18 pass.

**STEP-10 — A failed open says so**
Files: `macos/App/Platform/ProjectFiles.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Do: D-17. The existing `openSelectedFile…`, `open…` and `openObjectReports…` tests `await` the async calls.
Check: T-15, T-16 pass.

**STEP-11 — A failed Citation load names the Citation; the composer holds its step in the gate**
Files: `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/CitationComposerModelTests.swift`.
Do: D-16, D-22.
Check: T-17, T-19 pass; `rg -n stepContinuation macos/App` has no hit.

**STEP-12 — Focus returns to the Type control**
Files: `macos/App/Features/Sources/SourcePageIdentityHeader.swift`.
Do: D-18.
Check: builds; manual check 9.

**STEP-13 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md` (a `## S14-14 — Copy and VoiceOver on graph work` entry at the top); `docs/deployment-plan/spike-14/deployment-plan.md` (map row Status → `Landed #N`; S14-14 scope entry: replace "If the plan runs past about ten steps, it splits along these parts." with "One PR (user, 2026-11-01)."; in the **Form fields are named** part, after "`PVField` ties its label to its content in the kit," add "(the Promote evidence sheet B-18 named went in S13-23)"); `docs/deployment-plan/spike-14/app-health-audit.md` (A-12, A-13, A-14, A-17, A-18, A-20, A-21: fixed in S14-14); `docs/deployment-plan/spike-14/app-health-audit-2.md` (B-18: fixed in S14-14, kit fix); `.agents/skills/add-localized-string/SKILL.md` (two lines in the count section: a count in copy is a plural key; `check-localizable-xcstrings.py` fails on an uninflected `%lld <word>`. A plural word that prints no number (a bare noun or verb picked by a count) is a substitution with value `%#@noun@` and forms without `%arg`, as `deleteImpact.kind.citation`; a top-level `plural` variation must reference the number, or the catalog does not compile); `.agents/skills/add-ui-component/SKILL.md` (one line: a toast announces itself; an error callout or error line passes `announces: true`; a sheet's content closes through `pvSheetClose`; a form that asks before a step uses `PVDiscardGate`); this plan (Status, Landed note); `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj` → 0.2.6.
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-14-graph-work-voiceover.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVConfirmTests.swift` `dangerReturnCancels` *(new)* | `.danger.returnConfirms == false`; `.irreversible.returnConfirms == true` | AC-1.1, AC-1.2, D-1 |
| T-2 | Swift | `PVConfirmTests` `discardCopyIsTheKitsQuestion` *(new)* | `PVConfirmCopy.discardChanges` title, message, confirm, cancel equal the four `formDialogDiscard*` strings | D-2 |
| T-3 | Swift | `macos/ProvenenciaTests/AccessibilityAnnouncementTests.swift` *(new)* `joinsTitleAndMessage`, `nothingToSayIsNil`, `dangerIsHigh`, `copyHasAString` | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.9 inputs → outputs | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.9, D-6, D-7 |
| T-4 | Swift | `macos/ProvenenciaTests/PVDiscardGateTests.swift` *(new)* `discardHandsTheStepOutOnce`, `keepEditingDropsTheStep`, `keepEditingWithNothingHeld` | AC-2.6; AC-2.7; AC-2.8 (gate half) | AC-2.6, AC-2.7, AC-2.8, D-3 |
| T-5 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` (the `activate` test near line 207, rewritten) | `activate(row, onNavigate:, close: PVSheetClose { next in closed = true; next() })` → `closed`, then the row's location navigated | D-5 |
| T-6 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `cardSwitchAsksWhenDirty`, `cardSwitchConfirmMovesEdit`, `cardSwitchCancelKeepsEdit`, `cardSwitchCleanMovesAtOnce` *(new)* | AC-2.9, AC-2.10, AC-2.11, AC-2.12 | AC-2.9, AC-2.10, AC-2.11, AC-2.12, D-12 |
| T-7 | Swift | `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift` `locationLabelsNameTheGraph`, `locationLabelsFallBackToEvidenceGraph`, `composerAndPromoteLabelsCarryTheSurface` *(new)*; existing `evidenceGraphCrumbsNameSourceAndGraph`, `citationComposerCrumbsNameSourceAndGraph`, `promoteCrumbsNameSourceAndGraph`, `locationLabelsNameTheSection` | AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, AC-5.6 | AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, AC-5.6, D-13, D-14 |
| T-8 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `canvasLabelNamesTheGraph` *(new)* | AC-5.7, three cases | AC-5.7, D-15 |
| T-9 | Swift | `macos/ProvenenciaTests/L10nPluralTests.swift` `graphCopyNamesTheGraph` *(new)* | each of the five A-14 strings resolved; none contains "this Source"; each contains "vidence graph" | AC-6.7 |
| T-10 | Swift | `macos/ProvenenciaTests/PVFieldLabelTests.swift` *(new)* `fieldLabelFillsIn`, `ownLabelWins`, `noLabelStaysNil` | AC-4.1, AC-4.2, AC-4.3 | AC-4.1, AC-4.2, AC-4.3, D-10 |
| T-11 | Swift | `L10nPluralTests` `countsTakePlurals` *(new)* | AC-6.1, AC-6.2 | AC-6.1, AC-6.2 |
| T-12 | Swift | `PVDeleteImpactTests` `nounReadsPluralKey`, `unknownViaHeadingPlural` *(new)* | AC-6.3, AC-6.4 | AC-6.3, AC-6.4, D-19 |
| T-13 | Swift | `L10nPluralTests` `everyCatalogPluralPicksAFormPerCount` (changed per D-19: digit checks gated on `PluralShape.printsNumber`, no key prefix) | passes with the new keys; a `deleteImpact.kind.` key is checked for one ≠ several and no leftover placeholder | D-19 |
| T-14 | Script | `scripts/check-localizable-xcstrings.py` | the shipped catalog → exit 0; a scratch copy with `"test.planted": {"localizations": {"en": {"stringUnit": {"state": "translated", "value": "%lld subjects"}}}}` added, path passed as `argv[1]` → exit 1, and the output names `test.planted` | AC-6.5, AC-6.6, D-21 |
| T-15 | Swift | `macos/ProvenenciaTests/ProjectFilesTests.swift` `openObjectReportsNotOpenedWhenLaunchThrows`, `openObjectPassesARefusalThrough` *(new)* | a regular `.pdf` under `objects/`: `launch` throws `CocoaError(.fileReadUnknown)` → `.notOpened`; `launch` throws `ObjectRefusal.missing` → `.missing` | AC-7.5, D-17 |
| T-16 | Swift | `ProjectFilesTests` `refusalMessages` *(new)* | AC-7.6 | AC-7.6, D-17 |
| T-17 | Swift | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `lineRefreshShowsAFailedCitationLoad` (existing; its `fields.error` expectations become `citationLoadError` per AC-7.1, AC-7.2), `loadFailureUnlistedIsUnnamed`, `loadFailureLeavesFieldError`, `newCitationClearsLoadFailure` *(new)* | AC-7.1, AC-7.2 (remove `cit-1` from `citationsByID`, invalidate `.citationsByArtifact(project:graphId:artifactId: "art-0")`, `refreshCitationLines()`; then restore); AC-7.3 (`selectCitation("cit-missing")`, awaited through `awaitIdentitySwitch()`); AC-7.4; AC-7.9. Reasons are asserted by prefix, since FakeStore's `StoreBoom` is private | AC-7.1, AC-7.2, AC-7.3, AC-7.4, AC-7.9, D-16 |
| T-18 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo/every_delete-impact_kind_has_a_noun` *(new)* | the distinct `Table.Kind` values over `catalogmodel.Tables` plus `item` equal the Swift key set | AC-6.8, D-20 |
| T-19 | Swift | `CitationComposerModelTests` (the existing leave tests that read `pendingLeave`, near lines 802 and 1136) | pass unchanged with `leaveGate` | AC-2.13, D-22 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVDeleteImpactTests`, `WorkspaceToolbarBreadcrumbTests`, `L10nPluralTests`, `CitationComposerModelTests` (`lineRefreshShowsAFailedCitationLoad`), `EvidenceGraphModelTests` and `SourcePageModelTests` (open tests), `TestSwiftVocabularyMatchesGo`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/App` and `L10n.swift` with no hit left (`BreadcrumbNames`, `WorkspaceToolbar.locationLabels`, `loadFailureMessage`, `stepContinuation`, `activate(_:onNavigate:dismiss:then:)`, `viaUnknownOne`, `viaUnknownOther`, `deleteImpact.kind.source.one`, `nounNarrative`, `nounReconciliationClaim`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (the scope entry's split line, STEP-13)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no new visible component; the kit additions are parameters, environment values, one gate type and one modifier)
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
3. Still dirty, press a Kept-linked Citation row. "Discard changes?" shows; Keep editing keeps the form (AC-2.3); press the row again and press Escape: the question closes and nothing navigates (AC-2.8); press the row again and Discard: the sheet closes, then the Citation opens (AC-2.1, AC-2.2). Reopen the form without changes and press a row: it closes and navigates at once (AC-2.4). Change the name, press Save and, while it saves, the rows are disabled (AC-2.5).
4. VoiceOver: Add an Evidence graph. VoiceOver says "Evidence graph added" once (AC-3.5). Rotor → Headings lists "Evidence graphs" and the other section titles (AC-3.7).
5. VoiceOver: save the graph form with a forced error: VoiceOver speaks the callout once (AC-3.6). On a Person page, set a portrait: the toast is spoken once (AC-3.8). In the composer, a failed Save page on a depiction speaks its error line once (AC-7.8). An onboarding error toast is now spoken too (accepted).
6. VoiceOver: open the graph form and move to the description area: "Description" is spoken (AC-4.4). In the composer, move to the citation label field: its label is spoken once (AC-4.5).
7. Open "Jones household", then a Citation in it, then go back to the Source page. Long-press Back: rows read "1881 Census", "Jones household", "Jones household · Citation for CIT-…" (AC-5.3). VoiceOver on the canvas says "Evidence graph, Jones household" (AC-5.7).
8. In the composer, edit a Citation's transcription, pick another Citation, then Keep editing; pick it again and Discard: the switch happens only after Discard (AC-2.13). Edit an Artifact card's label without saving, then press the pencil on another card: "Discard changes?" shows; Keep editing keeps the text (AC-2.11); Discard moves the edit (AC-2.10).
9. Full Keyboard Access on: Source page, press the Type pencil, then the ×. Focus is on the Type pencil (AC-7.7). Pick a type instead: focus is on the pencil again.
10. In the composer, a Promote page and a delete notice, check "1 citation", "1 page", "1 observation", "1 subject on this graph" where one exists (AC-6.1, AC-6.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A substitution plural whose value omits the number does not resolve, or the catalog compile warns on it (a top-level plural variation with no number is a compile error, so D-19 uses a substitution) | STEP-9 probes one key first; T-12 catches a bad resolve; check-pr's no-warning build catches a warning → Stop |
| `.accessibilityLabel` inside `PVInput` overrides a call site's outer label | The only call-site label inside a `PVField` goes (D-11); T-10 holds the rule; manual check 6 |
| A toast that re-renders with the same text speaks twice | `pvAnnounce` posts on appear and on change to a new value only (D-6) |
| Moving `locationLabels` breaks a caller | Its only callers are `breadcrumbItems` and `HistoryJumpMenuHost` (§ 4), both main-actor |

**Stop and report if:**

- `PVConfirmContent`, `PVSheetChrome` or `PVDeleteImpactGroupList` no longer look as § 4 says.
- STEP-9's probe fails: `L10n.DeleteImpact.noun("citation", count: 1)` returns a key or a placeholder, or the build errors or warns on the `%#@noun@` substitution.
- The Swift build warns under `SWIFT_STRICT_CONCURRENCY = complete` for `LocationNames`' closures, `PVSheetClose` or `openObject`'s `launch`, and the fix needs an isolation change outside the files in § 7.
- `NSWorkspace.shared.open(_:withApplicationAt:configuration:)` has no `async throws` form in the SDK the project builds with.
- The planted-key run of the script passes (D-21 does not catch `%lld subjects`).
- D-20 finds a Go `Kind` beyond the 17 in § 4.
- A `PVField` call site's VoiceOver name gets worse in manual check 6 (a control speaks the field label in place of a more specific label it had).

## 12. Open questions

None.

## 13. PR

**Title:** S14-14 — Copy and VoiceOver on graph work
**Body:** objective in one sentence; what shipped against US-1 – US-7; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-14-graph-work-voiceover.md`.
