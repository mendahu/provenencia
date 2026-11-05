# S14-31 — Tab reaches a following text area

**Status:** Landed in #718 (2026-11-05)
**Landed:** Version bumped 0.3.3 → 0.3.4, because main was already at 0.3.3. `PVFieldFocus`'s binding init and `neighbours` are `@MainActor` (the plan left them non-isolated) so the focus claim compiles without Swift 6 sendability warnings. `focus:` sits before `activateOnAppear:` at P6–P8. P6's Tab skips the inline Save and Cancel when Full Keyboard Access is on, as P4's does. Round 1 added handle helpers beyond D-2 as same-pass cleanups: `PVFieldFocus.claimTab(shift:)` and `hasNeighbour`, and `Optional<PVFieldFocus>.isActive(own:)`, `focus(own:)` and `resign(own:)`, so no field repeats "handle, else own focus" or the Tab claim. The STEP-4 probe ran in round 1 in a titled test window, directly and inside a `.sheet`, with real `NSEvent` key-downs through `NSApp.sendEvent`. Results: (a) Shift-Tab arrives with `key` `"\u{19}"` and the shift modifier, not `"\t"`, so `.onKeyPress(.tab)` does not match it; P4 (Shift-Tab from the input, with the text area first) moved focus into the text area through the AppKit key loop, and its `previous` claim never ran. This tripped the §11 stop. D-6 amended (user, 2026-11-05): the claim also takes the back-tab key. Fixed in round 1 (F-8): `KeyEquivalent.backTab` names that character, and `PVInputTabClaim`, `PVComboBox` and `PVSelect` match it, so Shift-Tab takes the `previous` claim. The probe re-run logged `CLAIM previous from second -> first` for P4. (b) `FocusClaim` from the input landed in the `TextEditor`'s text view, and Shift-Tab from the area returned to the input. (c) The same held inside a sheet (the sheet window reported not key, so the sheet run is weaker). (d) The router attached to the input's field editor before the round 1 fix, and typing in the area after Tab then raised `undoManagerForTextView:` unrecognized on the router; with the fix no field editor carries a router and typing updates the binding. The §10 manual check was not run by the implementer.
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: every field, focus ring and dialog looks the same; only where Tab sends focus changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-31-tab-reaches-text-area`
**Version bump:** patch — a researcher-visible keyboard bug fix in the Mac app ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Tab into a text area is kit-wide" (reworded by this PR, see §5); bar outcome 26; dogfood note 2026-10-15 new-subject card Tab skips description (deleted on land); [`macos/App/DesignSystem/README.md`](../../../../macos/App/DesignSystem/README.md) built-components table (Input, TextArea rows)
**Planned:** 2026-11-05 against `main` @ `5d53a09c`
**Review:** [`S14-31-tab-reaches-text-area.review.md`](S14-31-tab-reaches-text-area.review.md) — Fits with changes — Ready with changes, 2 rounds (round 1 Reshape, accepted by the user 2026-11-05); no open blocker; To user answered

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

A researcher who types a label in a form and presses Tab lands in the description text area beside it, and Shift-Tab takes them back. Today Tab from the label skips the text area, so the researcher reaches for the mouse in every create and edit form that pairs a one-line field with a multiline one. The kit gains one focus handle for all its fields, and a two-field focus owner, `PVTabPair`, whose handles carry their Tab neighbour; this PR wires all eight Input / TextArea pairs in the app with it.

**Why now.** Dogfood note 2026-10-15 (new-subject card), pulled into Batch 7.

**What stays.** Tab and Shift-Tab between any other controls; Tab out of a text area with no neighbour (`PVTextAreaTabRouter`, 5d0657dd); Option-Tab inserting a tab character; Return in a text area inserting a newline; Return in an input submitting; Escape closing a sheet; `PVComboBox`'s own Tab claim (`onCommit` with `.tab`); every field's look.

## 2. User stories

**US-1.** As a researcher filling a form, when I press Tab in a label field that has a description text area after it, focus moves into the text area, so that I can keep typing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Evidence graph, arm Person, click the canvas; click into Label. Type `Mary`, press Tab | Caret is in Description; Description shows the focus ring; Label keeps `Mary` |
| AC-1.2 | Citation composer, New subject dialog; click into Label. Type `Ship`, press Tab | Caret is in Description |
| AC-1.3 | Each other forward pair (P3, P5–P8 in §3): caret in the input, press Tab | Caret is in that pair's text area |
| AC-1.4 | Properties → New property: caret in Label, press Tab | Caret is in Description; the read-only Key between them is skipped (it is a `Text`, not focusable, today too) |

**US-2.** As a researcher, when I press Shift-Tab in that text area, focus goes back to the label field.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1, then type `born 1820`, press Shift-Tab | Caret is in Label; Description keeps `born 1820` |
| AC-2.2 | Each forward pair P2–P3, P5–P8: caret in the text area, Shift-Tab | Caret is in that pair's input |

**US-3.** As a researcher using the composer at Wide, where Transcription is left of the Citation label (P4), Tab and Shift-Tab move between them in that order.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Composer `panelForm`, caret in Transcription, Tab | Caret is in the Citation label field |
| AC-3.2 | Caret in the Citation label field, Shift-Tab | Caret is in Transcription |

**US-4.** As a researcher, everything else about the keyboard in these forms is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Caret in a forward pair's text area, Tab | Focus leaves to the next key view, as today (`selectNextKeyView`) |
| AC-4.2 | Caret in a paired text area, Option-Tab | A tab character is inserted |
| AC-4.3 | Caret in a forward pair's input, Shift-Tab | Focus goes to the previous key view, as today |
| AC-4.4 | Artifact card edit (P8): caret in Label, Return | The edit commits, as today (`onSubmit`) |
| AC-4.5 | Evidence graph: edit a bridge (the form has no Label), caret in Description, Shift-Tab | Focus goes to the previous key view, as today; no crash |
| AC-4.6 | A `PVInput`, `PVTextArea`, `PVComboBox` or `PVSelect` with no neighbour (every other call site) | Behaves exactly as today, including the composer's Subject → Property → Value Tab chain |

## 3. Scope

The eight pairs (P-ids are used through the plan):

| Pair | View | Order |
| --- | --- | --- |
| P1 | `EvidenceGraphView.swift` `EvidenceGraphContent.createForm` | Label → Description (sheet) |
| P2 | `CitationComposerView.swift` New subject `pvFormDialog` | Label → Description (sheet) |
| P3 | `CitationComposerEvidenceColumn.swift` `sideForm` | `labelField` → `transcriptionColumn` |
| P4 | `CitationComposerEvidenceColumn.swift` `panelForm` | `transcriptionColumn` → `labelField` |
| P5 | `PropertiesView.swift` `PropertiesCreateHost.createForm` | Label → (read-only Key) → Description (sheet) |
| P6 | `PropertiesView.swift` `PropertiesIdentityEditor` | inline Label → Description |
| P7 | `SourceGraphFormDialog.swift` `SourceGraphForm.fields` | Name → Description (sheet) |
| P8 | `SourcePageArtifactsView.swift` `editFields(_:)` | Label → Description |

| In | Covered by |
| --- | --- |
| One kit focus handle `PVFieldFocus` *(new)* replaces `PVInputFocus`, `PVComboFocus`, `PVSelectFocus` (D-2) | STEP-1, AC-4.6 |
| `PVFieldFocus` carries optional `next` / `previous` claims made with `FocusClaim` (D-3) | STEP-1, manual §10 |
| `PVTabPair` *(new)* owns one pair's `FocusState` and hands out two handles in layout order (D-4) | STEP-2 |
| `PVTextArea(focus:)`; its Tab router claims `next` / `previous` before the key loop (D-5) | STEP-3, AC-2.1, AC-3.1, AC-4.1, AC-4.2 |
| `PVInput` claims Tab / Shift-Tab with `.onKeyPress(.tab)` when its handle has a neighbour; `PVComboBox` and `PVSelect` honour the same handle neighbours (D-6) | STEP-4, AC-1.1, AC-3.2, AC-4.3, AC-4.4, AC-4.6 |
| Tests | STEP-2, STEP-3; T-1 – T-4 |
| Wire P1–P4 | STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-3.1, AC-4.5 |
| Wire P5–P8 | STEP-6, AC-1.3, AC-1.4, AC-2.2 |
| Reword the working rule; README Input / TextArea rows; `add-ui-component` line (D-1) | STEP-7 |
| Manual check, version, docs | STEP-8 |

| Out | Home |
| --- | --- |
| Metadata value editors in `SourcePageMetadataView.swift`: the suggestion card's `PVInput` and the row editor's `PVTextArea` are different rows, never shown together | not planned |
| `SourcePageIdentityHeader.swift` Title (`PVTextArea`) and the type combo: not an Input / TextArea pair | not planned |
| Tab order between other controls (selects, combo boxes, checkboxes, buttons) | not planned |
| A check that flags an unwired pair | not planned (user, 2026-11-05: opt-in, reworded rule) |
| Seeding the opening caret in P1, P2, P4, P5 (today's initial responder stays) | not planned |
| `PVInput.focusField()` moving to `FocusClaim` (existing `activateOnAppear` code, unchanged) | not planned |
| A hosted-window or XCUITest that presses Tab in a real window | not planned (`add-swift-test` has no XCUITest) |
| `PVFormDialog` changes: it owns only `cancelFocused` and does not touch field focus | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Input focus handle | `macos/App/DesignSystem/Components/Input/PVInput.swift` `PVInputFocus` (line 70) | `fileprivate` `isActive`, `focus`, `install`; `install` builds a plain `TextField` with `.focused(binding, equals:)` |
| Combo focus handle | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` `PVComboFocus` (line 272) | `isActive`, `resign`, and an `install` identical to `PVInputFocus.install`. Used by `fieldText` (line 461), `fieldIsFocused` (line 426), `dismissAndBlur` (line 695) |
| Select focus handle | `macos/App/DesignSystem/Components/Select/PVSelect.swift` `PVSelectFocus` (line 34) | `isActive`, `activate`, `apply: (AnyView) -> AnyView`. Used by `PVSelectFocusModifier` (line 53), `isFocused`, `focusTrigger` (line 117) |
| Handle call sites | `macos/App/Features/Promote/PromoteView.swift` (359, 512); `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` (188, 227, 326, 344, 375, 384, 396, 410, 418, 669); `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` (443); `macos/App/Features/Sources/SourcePageIdentityHeader.swift` (204); `macos/App/Features/Names/NameValueEditorForm.swift` `formFocus` (11); `macos/App/Features/Dates/DateValueEditorForm.swift` `yearFocus` (9), `equalsFocus` (284) | Construct or store one of the three handles |
| Focus claim | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)` | The app's one focus hand-back: write next turn, write again a turn later if it did not land (23debc85) |
| Kit Tab claim precedent | `PVComboBox.swift` `field` `.onKeyPress(action: handleKey)` (line 434), `handleKey` `case .tab` (line 736) | A SwiftUI `TextField` sees Tab and Shift-Tab (`press.modifiers.contains(.shift)`) through `.onKeyPress`; returns `.handled` to claim, `.ignored` to let the key loop run. Shipped; used by `CitationComposerAuthoringBox.claimTab` (line 277) |
| Input | `PVInput.swift` `PVInput` | Three focus paths: `equalsFocus` (`focus:`), `externalFocus` (`focused:`), owned. No Tab handling: Tab goes to the AppKit key loop, which skips a following `TextEditor` (dogfood note) |
| Text area | `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` `PVTextArea` (line 53) | Owns `@FocusState isFocused` only; `.focused($isFocused)` on the `TextEditor` |
| Text area Tab | same file, `PVTextAreaTabRouter` (line 318), attached by `PVTextEditorInsetView` (line 248) through `PVTextEditorInsetClearer` | Tab → `selectNextKeyView`, Shift-Tab → `selectPreviousKeyView`; forwards every other command to SwiftUI's delegate |
| Form dialog | `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` `PVFormDialogContent` | Only `cancelFocused` on the Cancel button; does not touch field focus |
| Inline edit | `macos/App/DesignSystem/Components/InlineEdit/PVInlineEdit.swift` | `editingBody` adds `.onSubmit(submitFromKeyboard)`; no blur commit. P6's input has no `onBlur` |
| P1 | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `createForm` (line 400) | `VStack` of type chip, `createError`, `if !model.isEditingBridge { PVField { PVInput } }`, `PVField { PVTextArea }` |
| P2 | `macos/App/Features/CitationComposer/CitationComposerView.swift` (lines 123–147) | `VStack { PVField { PVInput(text: newSubjectLabelBinding, size: .sm) } PVField { PVTextArea(…, lineLimit: 1...3) } }` |
| P3, P4 | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` `sideForm` (line 50), `panelForm` (line 64), `private var labelField` (line 277), `transcriptionColumn(rows:)` (line 288) | `sideForm`: `labelField` then `transcriptionColumn(rows: 8...40)`. `panelForm`: `WeightedColumns { VStack { transcriptionColumn(rows: 6...40); footer }; labelField }` |
| P5 | `macos/App/Features/Properties/PropertiesView.swift` `PropertiesCreateHost.createForm` | Label `PVInput`, read-only Key `PVInput(isReadOnly: true)`, Description `PVTextArea` |
| P6 | same file, `PropertiesIdentityEditor` (line 621) | `body`'s `VStack { PVInlineEdit(editor: { … PVInput(activateOnAppear: true) … }); if model.isEditingIdentity { PVTextArea } … }` |
| P7 | `macos/App/Features/Sources/SourceGraphFormDialog.swift` `SourceGraphForm.fields` | Name `PVInput(activateOnAppear: true)` then `PVTextArea(lineLimit: 3...3)` |
| P8 | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `editFields(_:)` (line 383) | Label `PVInput(activateOnAppear: true)` with `.onSubmit`, then Description `PVTextArea(lineLimit: 3...8)` |
| Tests | `macos/ProvenenciaTests/PVTextAreaTests.swift` | Router moves on Tab / Shift-Tab, forwards the rest, forwards `textDidChange`; `TextViewCommandSpy` |

**Precedent.** `PVComboBox.handleKey` `case .tab` is the shape for `PVInput`'s Tab claim: `.onKeyPress`, claim → `.handled`, else `.ignored`. 23debc85 ("Share one focus claim…") is the shape for moving focus: every move is `FocusClaim.claim`. 5d0657dd is the router and its spy tests, which STEP-3 extends. `PVSelectFocus.apply` is the shape of the unified handle's `apply`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A pair is opt-in: the kit gives the mechanism, each Input / TextArea pair is wrapped in `PVTabPair`, and this PR wires all eight. The working rule is reworded to say so (user, 2026-11-05) | Programmatic focus between the fields is proven; the AppKit key loop SwiftUI builds is not ours to repair | Key-loop repair or geometry search — rewrites Tab for every control; a lint for unwired pairs — heuristic (user declined) |
| D-2 | `PVFieldFocus` in `macos/App/DesignSystem/Components/Input/PVFieldFocus.swift` *(new)* replaces `PVInputFocus`, `PVComboFocus` and `PVSelectFocus`. Members: `isActive: () -> Bool`, `focus: () -> Void` (write now), `resign: () -> Void`, `apply: (AnyView) -> AnyView`, `next: (@MainActor () -> Void)?`, `previous: (@MainActor () -> Void)?` (`FocusClaim.claim` is `@MainActor` and the target builds with `SWIFT_STRICT_CONCURRENCY = complete`; the struct itself stays non-isolated like the three it replaces). An internal memberwise `init(isActive:focus:resign:apply:next:previous:)` is the one stored-property init; the binding init (D-3) delegates to it, and tests use it. `PVSelect.swift`'s private `PVSelectFocusModifier` moves to `PVFieldFocus.swift` as `PVFieldFocusModifier(host: PVFieldFocus?, own: FocusState<Bool>.Binding)` (unchanged body) for `PVSelect` and `PVTextArea`. `install` goes: `PVInput` and `PVComboBox` call `focus.apply(AnyView(TextField(…).textFieldStyle(.plain)))`. `PVSelect`'s `activate` becomes `focus` (user, 2026-11-05: unify in this PR) | One focus contract for every kit field; the fourth user (`PVTextArea`) gets a handle named for what it is | Extending `PVInputFocus` — a text area holding an "input" handle and three copies kept |
| D-3 | `PVFieldFocus.init<Value: Hashable>(_ binding: FocusState<Value?>.Binding, equals value: Value, next: Value? = nil, previous: Value? = nil)`. A non-nil neighbour becomes `{ FocusClaim.claim(binding, neighbour) }`; nil stays nil. The mapping is `static func neighbours<Value>(next: Value?, previous: Value?, claim: @escaping @MainActor (Value) -> Void) -> (next: (@MainActor () -> Void)?, previous: (@MainActor () -> Void)?)` *(new)*; the binding init calls it with `{ FocusClaim.claim(binding, $0) }`, and T-4 calls it with a recorder. The Tab decision lives on the handle: `func tabClaim(shift: Bool) -> (@MainActor () -> Void)?` *(new)* returns `previous` when `shift`, else `next` | `FocusClaim` is the app's one focus hand-back and already writes twice; no `DispatchQueue` in the kit fields | A bare `DispatchQueue.main.async` — a second mechanism beside `FocusClaim` (S14-14 review M-2) |
| D-4 | `PVTabPair` in `macos/App/DesignSystem/Components/Input/PVTabPair.swift` *(new)* owns `@FocusState private var field: PVTabPairField?` (`enum PVTabPairField { case first, second }` *(new)*, with `var next: PVTabPairField?` (`.first` → `.second`, `.second` → nil) and `var previous: PVTabPairField?` (`.second` → `.first`, `.first` → nil)) and calls `content(handle(.first), handle(.second))`, where `handle(f) = PVFieldFocus($field, equals: f, next: f.next, previous: f.previous)`. Explicit `init(@ViewBuilder content: @escaping (PVFieldFocus, PVFieldFocus) -> Content)`. `body` returns `content(…)` unchanged, so the parent stack lays the fields out | Four pairs live in sheets; the `FocusState` is owned inside the sheet's content. Order is the argument order, so P4 is the same type | A `@FocusState` enum per call site — eight copies, three declared on the presenting view outside the sheet |
| D-5 | `PVTextArea(focus: PVFieldFocus? = nil)`. `.focused($isFocused)` becomes `.modifier(PVFieldFocusModifier(host: focus, own: $isFocused))` at the same place in the chain, so no handle keeps today's path; the chrome reads `focus?.isActive() ?? isFocused`; `activateOnAppear` calls `focus.focus` from its existing `DispatchQueue.main.async` (that line is unchanged). `PVTextAreaTabRouter` gains `var onTab: (@MainActor () -> Void)?` and `var onBacktab: (@MainActor () -> Void)?`: when the matching one is set it is called and the router returns `true`; when nil, today's `selectNextKeyView` / `selectPreviousKeyView`. `PVTextEditorInsetClearer(onTab:onBacktab:)` passes `focus?.next` / `focus?.previous`; its `updateNSView` calls `nsView.setTabCallbacks(onTab:onBacktab:)` *(new, on `PVTextEditorInsetView`, sets the two router properties)* before `clearInsets()` | The router already owns Tab in a `TextEditor`; the neighbour is one more case in it | `.onKeyPress` on the `TextEditor` — the text view takes Tab as a character before SwiftUI (why the router exists) |
| D-6 | `PVInput`: `focus:` takes `PVFieldFocus`. When that handle has `next` or `previous`, the field gets `.onKeyPress(.tab, phases: .down)`: `focus.tabClaim(shift: press.modifiers.contains(.shift))` non-nil → call it, `.handled`; nil → `.ignored`. No handle or no neighbour → no modifier. `PVComboBox.handleKey` `case .tab`: after the existing `onCommit` claim declines, `if let claim = focus?.tabClaim(shift:) { isOpen = false; claim(); return .handled }`, then today's `isOpen = false; return .ignored`. `PVSelect.handleTriggerKey` gains `case .tab:` with the same `focus?.tabClaim` claim, else `.ignored`. Every existing combo and select call site has nil neighbours, so none changes | Same claim shape as `PVComboBox.handleKey`, shipped on the same kind of `TextField` | An AppKit delegate router on the `NSTextField` — a second Tab mechanism in the kit (review M-1) |
| D-7 | P1 in bridge edit passes `focus: model.isEditingBridge ? nil : second` to its `PVTextArea`, so Shift-Tab keeps today's behaviour | The call site already branches on `isEditingBridge`; the router keeps one rule | A "did it land" fallback in the router |
| D-8 | P3 / P4: `labelField` becomes `labelField(focus: PVFieldFocus?)` and `transcriptionColumn(rows:)` becomes `transcriptionColumn(rows:focus:)`. `sideForm` wraps both in `PVTabPair { first, second in labelField(focus: first); transcriptionColumn(rows: 8...40, focus: second) }`; `panelForm` wraps its `WeightedColumns` in `PVTabPair { first, second in … transcriptionColumn(rows: 6...40, focus: first) … labelField(focus: second) }` | Each layout passes its own order | Wiring only `sideForm` — leaves Wide broken |
| D-9 | P6: `PVTabPair` goes inside `PropertiesIdentityEditor.body`'s `VStack`, around both of its children (the `PVInlineEdit` and the `if model.isEditingIdentity … else if …` block), so the `PVInlineEdit` editor's input and the text area below share one pair. The input keeps `activateOnAppear: true` | The two fields are in different subtrees of one view | Passing a handle down from `PropertiesContent` |
| D-10 | The pairs keep today's opening caret: no `PVTabPair` seeds focus | Out of scope; the manual check clicks into the first field where nothing places the caret today | Seeding `.first` — changes which field opens focused |

**Rules this plan bends.** Working rule "(Batch 7) Tab into a text area is kit-wide" says the fix lives in `PVInput` / `PVTextArea` / `PVFormDialog` so every pair gets Tab. D-1 puts the mechanism in the kit but needs each pair wrapped in `PVTabPair`, and `PVFormDialog` does not change. The user accepted opt-in wiring on 2026-11-05. STEP-7 rewords the rule to: "**(Batch 7) Tab into a text area is kit-wide.** The kit gives the mechanism: a `PVInput` and a `PVTextArea` in one form are wrapped in `PVTabPair`, so Tab and Shift-Tab move between them. S14-31 wires every pair in the app; a new pair is wrapped the same way (user, 2026-11-05). Binds S14-31." and updates the S14-31 scope entry's "via shared focus / key-view loop" to "via `PVTabPair`".

## 6. Design

### Codebase fit

Searched: `FocusState`, `FocusClaim`, `PVInputFocus`, `PVSelectFocus`, `PVComboFocus`, `onKeyPress`, `case .tab`, `TabRouter`, `insertTab`, `nextKeyView` across `macos/App`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVFieldFocus` *(new)* | type | `PVInputFocus`, `PVComboFocus`, `PVSelectFocus` | Replaces all three (D-2) |
| `PVTabPair`, `PVTabPairField` *(new)* | component | `CitationComposerAuthoringBox`'s `@FocusState focus: CitationAuthoringFocus?` | That is one feature's chain; eight pairs need one kit owner (D-4) |
| `PVTextAreaTabRouter.onTab`, `.onBacktab` *(new members)*; `PVTextEditorInsetView.setTabCallbacks(onTab:onBacktab:)` *(new)* | properties, func | itself | Extended |
| `PVFieldFocusModifier` | modifier | `PVSelect.swift` `PVSelectFocusModifier` | Moved and renamed, body unchanged |
| `PVInput`'s `.onKeyPress(.tab)`; `PVComboBox` / `PVSelect` `.tab` claim through the handle | modifier | `PVComboBox.handleKey` `case .tab` | Same shape; one neighbour rule for every field that takes the handle |
| `PVTextArea(focus:)` *(new param)* | init param | `PVInput(focus:)` | Same handle |
| `PVFieldFocusTests` *(new)* in `macos/ProvenenciaTests/PVFieldFocusTests.swift`; `PVFieldFocus.tabClaim(shift:)`, `PVFieldFocus.neighbours(next:previous:claim:)` *(new)* | tests, funcs | `PVTextAreaTests`; `PVComboBox.handleKey` | The Tab decision and the pair order as pure functions, testable without a window |

**Builds on.** `FocusClaim` (23debc85), `PVComboBox`'s Tab claim, `PVTextAreaTabRouter` (5d0657dd), `PVSelectFocus.apply`.
**Deletes.** `PVInputFocus`, `PVComboFocus`, `PVSelectFocus` and both `install` closures; `PVSelect`'s `activate` name; `PVSelectFocusModifier` (moved and renamed `PVFieldFocusModifier`). Nothing else becomes dead.
**Next adds.** A new form with a label and a description wraps them in `PVTabPair`. A longer chain (three fields) is a `PVFieldFocus` built with `next` / `previous` at a call site that owns its own `FocusState`; no new type.

### Reach

None — no new navigation path or entry point. P1 without its input is handled at the call site (D-7, AC-4.5).

### Scale

None — no write, list or loop changes.

### Edges

- **Assumptions.** A SwiftUI `TextField` sees forward Tab through `.onKeyPress` (held by `PVComboBox` and `claimTab` in production). That it also sees Shift-Tab is not shown in production (the combo only reads the Shift modifier to decline); STEP-4 probes it first. `FocusClaim` lands a write into a `TextEditor` bound with `.focused(binding, equals:)` (as `activateOnAppear` does today with an owned `FocusState`). The manual check holds both.
- **Neighbors.** A handle with no neighbour: no `.onKeyPress`, router callbacks nil → today's behaviour (AC-4.6, T-2, T-3). Option-Tab, Return: router forwards them (T-2, existing tests). `PVComboBox` / `PVSelect` with a `PVFieldFocus`: neighbours are nil at every existing call site, so `tabClaim` is nil and Tab is `.ignored` as today; the combo's `onCommit` claim still runs first. P4 (`panelForm`): the Uncertain checkbox, the reading action and the footer's Save sit between Transcription and the Citation label in layout order; the pair's Tab jumps over them. They are Tab stops only with Full Keyboard Access on (off by default), and remain reachable from the label's forward Tab and by click; accepted.
- **Removed inputs.** The partner field is not drawn (P1 bridge edit): D-7.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — no data read.

### Schema

None — no catalog change.

### Go core

None — Mac kit only.

### FFI

None —

### macOS

```swift
struct PVFieldFocus {
    let isActive: () -> Bool
    let focus: () -> Void
    let resign: () -> Void
    let apply: (AnyView) -> AnyView
    let next: (@MainActor () -> Void)?
    let previous: (@MainActor () -> Void)?

    init(
        isActive: @escaping () -> Bool,
        focus: @escaping () -> Void,
        resign: @escaping () -> Void,
        apply: @escaping (AnyView) -> AnyView,
        next: (@MainActor () -> Void)? = nil,
        previous: (@MainActor () -> Void)? = nil
    )

    init<Value: Hashable>(
        _ binding: FocusState<Value?>.Binding,
        equals value: Value,
        next: Value? = nil,
        previous: Value? = nil
    )

    func tabClaim(shift: Bool) -> (@MainActor () -> Void)?
    static func neighbours<Value>(
        next: Value?, previous: Value?,
        claim: @escaping @MainActor (Value) -> Void
    ) -> (next: (@MainActor () -> Void)?, previous: (@MainActor () -> Void)?)
}

enum PVTabPairField: Hashable {
    case first, second
    var next: PVTabPairField? { self == .first ? .second : nil }
    var previous: PVTabPairField? { self == .second ? .first : nil }
}

/// Owns focus for two fields in one form, in layout order: Tab in the first
/// moves to the second, Shift-Tab in the second moves back.
struct PVTabPair<Content: View>: View {
    @FocusState private var field: PVTabPairField?
    private let content: (PVFieldFocus, PVFieldFocus) -> Content
    init(@ViewBuilder content: @escaping (PVFieldFocus, PVFieldFocus) -> Content)
    private func handle(_ f: PVTabPairField) -> PVFieldFocus {
        PVFieldFocus($field, equals: f, next: f.next, previous: f.previous)
    }
    var body: some View { content(handle(.first), handle(.second)) }
}
```

Call-site shape (P2):

```swift
PVTabPair { label, description in
    PVField(label: L10n.CitationComposer.newSubjectLabel, error: …) {
        PVInput(text: newSubjectLabelBinding, size: .sm, focus: label)
    }
    PVField(label: L10n.CitationComposer.descriptionLabel) {
        PVTextArea(text: newSubjectDescriptionBinding, lineLimit: 1...3, focus: description)
    }
}
```

`PVTabPair` goes inside each form's existing `VStack`, around the two `PVField`s (and, for P5, the Key field between them), not around the `VStack`. P6 is the same rule (D-9). The one exception: P4 wraps `WeightedColumns` (D-8).

States: P1 bridge edit (D-7). Accessibility: VoiceOver order and labels are unchanged; the move is the same focus change a click makes.

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — One focus handle**
Files: `macos/App/DesignSystem/Components/Input/PVFieldFocus.swift` *(new)*, `PVInput.swift`, `ComboBox/PVComboBox.swift`, `Select/PVSelect.swift`, and the call sites in §4 "Handle call sites"; Skill: `add-ui-component`.
Do: D-2, D-3 (including `neighbours` and `tabClaim`). Delete the three old structs. Each call site changes only the type name (`PVSelectFocus(` → `PVFieldFocus(`, and so on); `PVSelect` calls `focus.focus()` where it called `activate()`.
Check: `rg -n "PVInputFocus|PVComboFocus|PVSelectFocus" macos .agents .cursor docs/*.md` has no hit; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-2 — The pair owner**
Files: `macos/App/DesignSystem/Components/Input/PVTabPair.swift` *(new)*.
Do: D-4 per §6. Write T-3, T-4 in `macos/ProvenenciaTests/PVFieldFocusTests.swift` *(new)*. Skill: `add-swift-test`.
Check: the same `xcodebuild test` passes.

**STEP-3 — Text area takes a handle and its neighbours**
Files: `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/ProvenenciaTests/PVTextAreaTests.swift`.
Do: D-5. With no handle, the owned `@FocusState isFocused` path stays as it is. Write T-1, T-2. Skill: `add-swift-test`.
Check: `PVTextAreaTests` pass (existing and new).

**STEP-4 — Input claims Tab for its neighbour**
Files: `macos/App/DesignSystem/Components/Input/PVInput.swift`, `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift`, `macos/App/DesignSystem/Components/Select/PVSelect.swift`.
Do: first, the probe: add `.onKeyPress(.tab, phases: .down)` to the field with a temporary `print(press.modifiers)` returning `.ignored`, run the app, press Tab and Shift-Tab in P2's Label. Both must print; if Shift-Tab does not, stop (§11). Remove the print. Then D-6 in the three files above.
Check: tests pass.

**STEP-5 — Wire the graph card and the composer (P1–P4)**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/CitationComposer/CitationComposerView.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`.
Do: §6 call-site shape; D-7 for P1; D-8 for P3 / P4. Keep every `.accessibilityIdentifier`.
Check: builds; manual §10 items 1–4 pass.

**STEP-6 — Wire Properties and the Source page (P5–P8)**
Files: `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`.
Do: §6 call-site shape; P5 wraps Label, Key and Description, and only Label and Description take `focus:`; D-9 for P6; P8 keeps `.onSubmit { commitEdit(art) }`.
Check: builds; `rg -n "PVTabPair \{" macos/App/Features` lists exactly 8 hits in the 6 files named in STEP-5 and STEP-6; manual §10 items 5–8 pass.

**STEP-7 — Rule and kit docs**
Files: `docs/deployment-plan/spike-14/deployment-plan.md` (working rule and S14-31 scope entry, per §5 "Rules this plan bends"), `macos/App/DesignSystem/README.md` (built-components table: the Input and TextArea rows say a pair is wrapped in `PVTabPair`; one row for `PVFieldFocus` as the focus handle every kit field takes), `.agents/skills/add-ui-component/SKILL.md` (one line pointing at the README rows).
Do: as listed.
Check: `rg -n "PVTabPair" docs/deployment-plan/spike-14/deployment-plan.md macos/App/DesignSystem/README.md .agents/skills/add-ui-component/SKILL.md` has a hit in each file; `python3 scripts/check-guidance-names.py` passes.

**STEP-8 — Version and docs**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`; `docs/deployment-plan/spike-14/completed.md`; `deployment-plan.md` (PR map row S14-31: Status → `Landed #N`; "Where the notes went" row for the dogfood note; the S14-14 scope line naming `PVComboFocus` → `PVFieldFocus`); this plan; `docs/deployment-plan/spike-14/ux/2026-10-15-new-subject-card-tab-skips-description.md` (delete).
Do: patch bump together, 0.3.2 → 0.3.3 (or the next patch on `main` at land); DoD docs lines.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVTextAreaTests.swift` `tabAndBacktabCallTheNeighbourWhenSet` *(new)* | Router with `onTab` and `onBacktab` counting: `insertTab` → `true`, `onTab` 1; `insertBacktab` → `true`, `onBacktab` 1; spy delegate saw neither | AC-2.1, AC-3.1, D-5 |
| T-2 | Swift | same file, `noNeighbourKeepsTodaysRouting` *(new)* | Router with `onTab` nil, `onBacktab` set: `insertTab` → `true`, `onBacktab` 0 (key loop path, no window); `insertTabIgnoringFieldEditor` and `insertNewline` → forwarded to the spy, `false` | AC-4.1, AC-4.2, AC-4.6 |
| T-3 | Swift | `macos/ProvenenciaTests/PVFieldFocusTests.swift` `tabClaimPicksTheNeighbour` *(new)* | Handle (memberwise init) whose `next` and `previous` count calls: `tabClaim(shift: false)` runs `next`; `tabClaim(shift: true)` runs `previous`; handle with only `next`, `shift: true` → nil; no neighbours → nil both ways | AC-1.1, AC-3.2, AC-4.3, AC-4.6, D-3, D-6 |
| T-4 | Swift | same file, `tabPairOrder` *(new)* | `PVTabPairField.first.next == .second`, `.first.previous == nil`, `.second.previous == .first`, `.second.next == nil`; `PVFieldFocus.neighbours(next: PVTabPairField.second, previous: nil, claim: recorder)` → `next` non-nil and calling it records `.second`, `previous` nil | D-3, D-4 |

Handles for T-1, T-3 and T-4 are built with the memberwise-style init the test needs: `PVFieldFocus` gets an `init(isActive:focus:resign:apply:next:previous:)` *(new, internal)* used by tests and by the binding init.

Not unit-tested, checked by hand (§10): `FocusClaim` landing focus in each mounted field, `PVTabPair` layout, D-7, D-9. The order and the claim mapping are held by T-3 and T-4.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVTextAreaTests`; `CitationComposerModelTests` (the composer chain that uses the renamed handles); any test that builds `PVSelect` or `PVComboBox` with a focus handle.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: PVInputFocus, PVComboFocus, PVSelectFocus `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (STEP-7)
- [ ] UI looks as on main; no new kit primitive beyond PVFieldFocus and PVTabPair
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Dogfood note deleted; its "Where the notes went" row says landed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Any project with one Source, one Evidence graph and one Artifact. Default window, then 2560pt wide.

1. **P1.** Open the graph, arm Person, click the canvas, click into Label. Type `Mary`, Tab → caret in Description. Type `born 1820`, Shift-Tab → caret in Label, both values kept. Tab twice → focus leaves Description as before. Edit a bridge, caret in Description, Shift-Tab → previous key view, no crash (AC-1.1, AC-2.1, AC-4.1, AC-4.5).
2. **P2.** Composer, New subject. Click into Label, type `Ship`, Tab → Description; Shift-Tab → Label (AC-1.2, AC-2.2).
3. **P3.** Composer at the default window (side form). Caret in the Citation label, Tab → Transcription; Shift-Tab → back. Option-Tab in Transcription inserts a tab; Return inserts a newline (AC-1.3, AC-2.2, AC-4.2).
4. **P4.** Composer at 2560pt (panel form). Caret in Transcription, Tab → Citation label; Shift-Tab → Transcription (AC-3.1, AC-3.2). Open an authoring box: Subject → Property → Value Tab chain works as before (AC-4.6).
5. **P5.** Properties → New property. Label, Tab → Description (Key skipped); Shift-Tab → Label (AC-1.4, AC-2.2).
6. **P6.** Properties, a researcher Property, edit its identity. Label, Tab → Description; Shift-Tab → Label (AC-1.3, AC-2.2).
7. **P7.** Source page → New graph. Name, Tab → Description; Shift-Tab → Name (AC-1.3, AC-2.2).
8. **P8.** Source page, edit an Artifact card. Label, Tab → Description; Shift-Tab → Label; Return in Label commits (AC-1.3, AC-2.2, AC-4.4).
9. Promote's target select and record search, the Source type combo, the composer's Subject / Property combos: focus, Escape and Tab behave as before (AC-4.6, the renamed handle and the new `.tab` cases).
10. Each of P1–P8 looks the same as on `main` (spacing unchanged by `PVTabPair`).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `PVTabPair`'s body changes stack spacing | Its body returns the content unchanged; manual item 10 |
| The handle rename touches the composer's focus chain | Type-name change only; `CitationComposerModelTests` and manual item 4 |

**Stop and report if:**

- The STEP-4 probe: `.onKeyPress(.tab)` on `PVInput`'s field is not called for Tab or for Shift-Tab.
- `FocusClaim.claim` does not move focus into the `PVTextArea` in P1 or P2.
- A `FocusState` owned by `PVTabPair` inside a sheet does not move focus (P1, P2, P5 or P7).
- `PVTabPair` changes the spacing or layout of any of P1–P8 compared with `main`.
- Renaming a handle needs a change beyond the type name or `activate` → `focus` at any call site.

## 12. Open questions

None.

## 13. PR

**Title:** S14-31 — Tab reaches a following text area
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-31-tab-reaches-text-area.md`.
