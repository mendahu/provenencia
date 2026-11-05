# S14-31 — Tab reaches a following text area

**Status:** Draft
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
**Context:** working rule "(Batch 7) Tab into a text area is kit-wide"; bar outcome 26; dogfood note [`ux/2026-10-15-new-subject-card-tab-skips-description.md`](../ux/2026-10-15-new-subject-card-tab-skips-description.md) (deleted on land); [`macos/App/DesignSystem/README.md`](../../../../macos/App/DesignSystem/README.md#interaction-state) § Interaction state
**Planned:** 2026-11-05 against `main` @ `5d53a09c`
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

A researcher who types a label in a form and presses Tab lands in the description text area below it, and Shift-Tab takes them back to the label. Today Tab from the label skips the text area, so the researcher reaches for the mouse in every create and edit form that pairs a one-line field with a multiline one. The fix is a kit mechanism in `PVInput` and `PVTextArea` (working rule "Tab into a text area is kit-wide"), and this PR wires it into all seven Input → TextArea pairs in the app.

**Why now.** Dogfood note 2026-10-15 (new-subject card), pulled into Batch 7.

**What stays.** Tab and Shift-Tab between any other controls, Tab *leaving* a text area forward (`PVTextAreaTabRouter`, 5d0657dd), Option-Tab inserting a tab character, Return in a text area inserting a newline, Return in an input submitting, Escape closing a sheet, and every field's look.

## 2. User stories

**US-1.** As a researcher filling a form, when I press Tab in a label field that has a description text area right after it, focus moves into the text area, so that I can keep typing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Evidence graph, arm Person, click the canvas: the new-subject card opens with the caret in Label. Type `Mary`, press Tab | Caret is in Description; Description shows the focus ring; Label keeps `Mary` |
| AC-1.2 | Citation composer, New subject dialog. Type `Ship`, press Tab | Caret is in Description |
| AC-1.3 | Each other pair in §3 In (P3–P7): caret in the input, press Tab | Caret is in that pair's text area |
| AC-1.4 | Properties → New property: caret in Label, press Tab | Caret is in Description; the read-only Key between them is skipped (it is not focusable today either) |

**US-2.** As a researcher, when I press Shift-Tab in that text area, focus goes back to the label field, so that I can fix the label.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1, then type `born 1820`, press Shift-Tab | Caret is in Label; Description keeps `born 1820` |
| AC-2.2 | Each pair P2–P7: caret in the text area, Shift-Tab | Caret is in that pair's input |

**US-3.** As a researcher, everything else about the keyboard in these forms is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Caret in a paired text area, press Tab | Focus leaves to the next key view, as today (`selectNextKeyView`) |
| AC-3.2 | Caret in a paired text area, press Option-Tab | A tab character is inserted |
| AC-3.3 | Caret in a paired input, press Shift-Tab | Focus goes to the previous key view, as today |
| AC-3.4 | Artifact card edit (P7): caret in Label, press Return | The edit commits, as today (`onSubmit`) |
| AC-3.5 | Evidence graph: edit a bridge (the form has no Label) | Description works as before; no crash, Shift-Tab goes to the previous key view |
| AC-3.6 | A `PVInput` or `PVTextArea` with no `tabPair` (every other call site) | Behaves exactly as today |

## 3. Scope

The seven pairs (P-ids are used through the plan):

| Pair | View | Input → TextArea |
| --- | --- | --- |
| P1 | `EvidenceGraphView.swift` `EvidenceGraphContent.createForm` | Label → Description (sheet) |
| P2 | `CitationComposerView.swift` New subject `pvFormDialog` | Label → Description (sheet) |
| P3 | `CitationComposerEvidenceColumn.swift` `sideForm` | `labelField` → `transcriptionColumn` |
| P4 | `PropertiesView.swift` `PropertiesCreateHost.createForm` | Label → (read-only Key) → Description (sheet) |
| P5 | `PropertiesView.swift` `PropertiesIdentityEditor` | inline Label → Description |
| P6 | `SourceGraphFormDialog.swift` `SourceGraphForm.fields` | Name → Description (sheet) |
| P7 | `SourcePageArtifactsView.swift` `editFields(_:)` | Label → Description |

| In | Covered by |
| --- | --- |
| `PVInputFocus` gains `apply` and its members become module-internal so `PVTextArea` can use it (D-2) | STEP-1 |
| `PVTabPair` + `PVTabPairFocus` *(new)* own one pair's focus (D-1, D-3) | STEP-1 |
| `PVTextArea(tabPair:)`: focus from the pair; Shift-Tab focuses the pair's input (D-4) | STEP-2, AC-2.1, AC-3.1, AC-3.2 |
| `PVInput(tabPair:)`: focus from the pair; Tab focuses the pair's text area through `PVInputTabRouter` *(new)* (D-5, D-6) | STEP-3, AC-1.1, AC-3.3, AC-3.4 |
| Unit tests for both routers and the field match | STEP-4, T-1 – T-5 |
| Wire P1, P2, P3 | STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-3.5 |
| Wire P4, P5, P6, P7 | STEP-6, AC-1.3, AC-1.4, AC-2.2 |
| DesignSystem README and `add-ui-component` say how a pair is wired (D-7) | STEP-7 |
| Manual check, docs, version | STEP-8 |

| Out | Home |
| --- | --- |
| Composer `panelForm` (Wide layout): the text area is left of the label, so it is not an Input → following TextArea pair | not planned |
| Metadata value editors in `SourcePageMetadataView.swift`: the `PVInput` and `PVTextArea` are in different rows and never shown together as a pair | not planned |
| Tab order between other controls (selects, combo boxes, checkboxes, buttons) | not planned |
| Automatic pairing with no call-site wiring (key-view-loop repair or geometry search) | not planned (D-1) |
| A hosted-window or XCUITest that presses Tab in a real window | not planned (`add-swift-test` has no XCUITest; T-1 – T-5 test the routing seams) |
| Renaming `PVInputFocus` to a field-neutral name | not planned (D-2) |
| Refactoring `PVTextEditorInsetView.nearestTextView` to the new field match | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Input focus handle | `macos/App/DesignSystem/Components/Input/PVInput.swift` `PVInputFocus` (line 70) | `fileprivate` `isActive`, `focus`, `install`; `install` builds a `TextField` with `.focused(binding, equals:)`. Only `PVInput` can use it |
| Input | same file, `PVInput` | Three focus paths: `equalsFocus` (`focus:`), `externalFocus` (`focused:`), owned. No Tab handling: Tab goes to the AppKit key loop, which skips a following `TextEditor` (dogfood note) |
| Text area | `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` `PVTextArea` | Owns `@FocusState isFocused` only; `.focused($isFocused)` on the `TextEditor`. No host focus |
| Text area Tab | same file, `PVTextAreaTabRouter` (line 318), attached by `PVTextEditorInsetView` (line 248) | Wraps the text view's delegate; Tab → `selectNextKeyView`, Shift-Tab → `selectPreviousKeyView`; forwards every other command and selector to SwiftUI's delegate. Shift-Tab has the same loop problem in reverse |
| Selectors | same file, `PVTextAreaTabRouting.movesFocus` | Tab and Shift-Tab selectors; Option-Tab (`insertTabIgnoringFieldEditor`) is not one |
| Focus handle precedent | `macos/App/DesignSystem/Components/Select/PVSelect.swift` `PVSelectFocus` | `apply: (AnyView) -> AnyView` applies `.focused(binding, equals:)` to any view — the shape D-2 copies |
| P1 | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `createForm` (line 400) | `PVInput` (only when `!model.isEditingBridge`) then `PVTextArea`, no shared focus. Presented in `pvFormDialog`; `EvidenceGraphContent.focus` is the canvas palette's `EvidenceGraphFocus`, not the form's |
| P2 | `macos/App/Features/CitationComposer/CitationComposerView.swift` (lines 123–147) | `PVInput(text: newSubjectLabelBinding, size: .sm)` then `PVTextArea(text: newSubjectDescriptionBinding, lineLimit: 1...3)` |
| P3 | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` `sideForm`, `labelField`, `transcriptionColumn(rows:)` | `labelField` then `transcriptionColumn(rows: 8...40)`; `panelForm` uses both in the other order |
| P4 | `macos/App/Features/Properties/PropertiesView.swift` `PropertiesCreateHost.createForm` | Label `PVInput`, read-only Key `PVInput(isReadOnly: true)` (a `Text`), Description `PVTextArea` |
| P5 | same file, `PropertiesIdentityEditor` (line 621) | Label `PVInput(activateOnAppear: true)` in `PVInlineEdit`'s `editor:`; `PVTextArea` below when `model.isEditingIdentity` |
| P6 | `macos/App/Features/Sources/SourceGraphFormDialog.swift` `SourceGraphForm.fields` | Name `PVInput(activateOnAppear: true)` then `PVTextArea(lineLimit: 3...3)` |
| P7 | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `editFields(_:)` | Label `PVInput(activateOnAppear: true)` with `.onSubmit`, then Description `PVTextArea(lineLimit: 3...8)` |
| Tests | `macos/ProvenenciaTests/PVTextAreaTests.swift` | Router moves on Tab / Shift-Tab, forwards the rest, forwards `textDidChange`; `TextViewCommandSpy` |

**Precedent.** 5d0657dd ("Let Tab leave a form text area…") added `PVTextAreaTabRouter` and its tests. Copy its delegate wrapping (`attach`, `responds(to:)`, `forwardingTarget(for:)`), its re-attach after SwiftUI installs its own delegate (`scheduleTabRouterRecheck`), its spy-delegate tests, and its focus write one turn later (`DispatchQueue.main.async`). `PVSelectFocus.apply` is the precedent for a focus handle any view can take.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The pair is explicit: the kit gives the mechanism, and each Input → TextArea pair is wired by wrapping it in `PVTabPair` and passing `tabPair:` to both fields. Every pair in the app (P1–P7) is wired in this PR | Programmatic focus into both fields is proven (`activateOnAppear`, `PVInputFocus`); the AppKit key loop SwiftUI builds is not ours to fix and its cause cannot be pinned without a probe that may differ by macOS version | Repairing the key view loop or a geometric "next field" search — rewrites Tab for every control, including SwiftUI views with no AppKit counterpart |
| D-2 | `PVTextArea` takes the existing `PVInputFocus`, extended with `apply: (AnyView) -> AnyView` like `PVSelectFocus`; `isActive`, `focus`, `apply` drop `fileprivate`. The name stays | A fourth near-copy of the focus handle (`PVInputFocus`, `PVComboFocus`, `PVSelectFocus`) is worse than one handle used by two fields | `PVTextAreaFocus` *(new)* — a copy; renaming to `PVFieldFocus` — a sweep across call sites this PR does not otherwise touch |
| D-3 | `PVTabPair` is a view that owns `@FocusState private var field: PVTabPairField?` and hands its content a `PVTabPairFocus { input, textArea }` (two `PVInputFocus`). Its `body` returns the content unchanged, so the parent stack lays out the fields | Four pairs live in sheets; a `FocusState` owned inside the sheet's content is in the sheet's hierarchy. One kit type keeps each call site to a wrapper and two arguments | A `@FocusState` enum per call site — seven copies, and P1/P2/P4 would declare it on the presenting view, outside the sheet |
| D-4 | `PVTextArea(tabPair:)` focuses with `tabPair.textArea`; Shift-Tab in it calls `tabPair.input.focus` (one turn later) instead of `selectPreviousKeyView`. Tab forward is unchanged | The text area only needs to know its partner backward; forward still leaves the pair | Sending Tab forward to a named next control — out of scope |
| D-5 | `PVInput(tabPair:)` installs its field with `tabPair.input` (the existing `equalsFocus` path) and a background `PVInputTabHook` *(new)* that attaches `PVInputTabRouter` *(new)* to its `NSTextField`. The router answers `control(_:textView:doCommandBy:)`: `insertTab(_:)` → `tabPair.textArea.focus` one turn later, returns `true`; every other command and selector forwards to SwiftUI's delegate | Same wrapping as `PVTextAreaTabRouter`, which is proven against SwiftUI's delegate | `.onKeyPress(.tab)` on the `TextField` — not proven in this codebase to see Tab before the field editor handles it |
| D-6 | `PVInputTabHook` picks its field by geometry: among `NSTextField` descendants within 4 ancestors of the hook, the one whose window frame contains the hook's frame centre (`PVInputTabHook.matchingField(_:hookFrame:)`, a static pure function). No match → no router, Tab behaves as today | The hook is the field's background, so it shares the field's frame; the first-descendant search used by the text area can pick another field in the same form | First-descendant search — wrong field when two inputs share a container |
| D-7 | `PVInput(tabPair:)` with `focus:` or `focused:` also set is a programmer error: `assertionFailure`, and `tabPair` wins | One focus source per field | Silent precedence |
| D-8 | `activateOnAppear` with `tabPair` calls `tabPair.input.focus` / `tabPair.textArea.focus` | Keeps P5, P6, P7 opening with the caret in the input | Owned focus — would split the field's focus from the pair's |
| D-9 | P1 with `isEditingBridge`: the text area keeps `tabPair:` and no input is drawn; Shift-Tab then focuses nothing and AppKit's previous key view is used: `PVTextAreaTabRouter` falls back to `selectPreviousKeyView` when `tabPair.input.isActive` is still false one turn after the write | No crash and today's behaviour when the partner is absent | Two code paths in `createForm` |
| D-10 | P3 wires `sideForm` only; `panelForm` passes `tabPair: nil` | In `panelForm` the text area comes first | Wiring both — a backwards pair |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: `FocusState`, `PVInputFocus`, `PVSelectFocus`, `PVComboFocus`, `TabRouter`, `insertTab`, `nextKeyView`, `onKeyPress(.tab` across `macos/App`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVInputFocus.apply` *(new member)* | func | `PVSelectFocus.apply` | Extended, not copied (D-2) |
| `PVTabPair`, `PVTabPairFocus`, `PVTabPairField` *(new)* in `macos/App/DesignSystem/Components/Input/PVTabPair.swift` | component | `CitationComposerAuthoringBox`'s per-view `CitationAuthoringFocus` | That is one feature's focus enum; seven pairs need one kit owner (D-3) |
| `PVInputTabRouter` *(new)* in `macos/App/DesignSystem/Components/Input/PVInputTabRouter.swift` | type | `PVTextAreaTabRouter` | `NSTextField` delegates use `NSControlTextEditingDelegate.control(_:textView:doCommandBy:)`, not `NSTextViewDelegate.textView(_:doCommandBy:)`; same wrapping |
| `PVInputTabHook` *(new)*, same file | `NSViewRepresentable` | `PVTextEditorInsetClearer` / `PVTextEditorInsetView` | That one clears text-view insets; it is private to `PVTextArea.swift` and finds a text view, not a field |
| `PVTextAreaTabRouter.onBacktab` *(new member)* | property | itself | Extended |
| `PVInput(tabPair:)`, `PVTextArea(tabPair:)` *(new params)* | init params | `PVInput(focus:)` | Extended |
| `PVInputTabRouterTests` *(new)* in `macos/ProvenenciaTests/PVInputTabRouterTests.swift` | tests | `PVTextAreaTests` | Different delegate protocol; reuses the spy shape |

**Builds on.** `PVInputFocus` and the `equalsFocus` install path; `PVTextAreaTabRouter` wrapping and re-attach; `PVSelectFocus.apply`; the one-turn focus write from 5d0657dd.
**Deletes.** None — nothing becomes dead; no call site loses its last caller.
**Next adds.** A new form with a label and a description wraps them in `PVTabPair` (documented in STEP-7). A pair of two inputs or a longer chain is not named in the map.

### Reach

None — no new navigation path or entry point. The one new state is P1 without its input (bridge edit), handled by D-9 and AC-3.5.

### Scale

None — no write, list or loop changes. Each pair adds one view, one `FocusState`, and one delegate wrapper per mounted input.

### Edges

- **Assumptions.** SwiftUI backs a macOS `TextField` with an `NSTextField` whose delegate receives `control(_:textView:doCommandBy:)`, and may reinstall its delegate after the first layout (handled by the recheck, as in `PVTextEditorInsetView`). T-1 – T-3 hold the forwarding; the manual check holds the real field.
- **Neighbors.** Shift-Tab in the input, Option-Tab and Return in the text area, Return in the input: unchanged (AC-3.2 – AC-3.4, T-2, T-4). A field with no `tabPair`: unchanged (AC-3.6; the hook is not mounted and `onBacktab` is nil).
- **Removed inputs.** The partner field is not drawn (P1 bridge edit): D-9, AC-3.5.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — no data read.

### Schema

None — no catalog change.

### Go core

None — Mac kit only.

### FFI

None —

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Focus handle | `PVInputFocus` | `macos/App/DesignSystem/Components/Input/PVInput.swift` | Add `let apply: (AnyView) -> AnyView` set to `AnyView(view.focused(binding, equals: value))`; `isActive`, `focus`, `apply` internal; `install` stays `fileprivate` |
| Pair owner | `PVTabPair` *(new)* | `Components/Input/PVTabPair.swift` | Sketch below |
| Input | `PVInput(tabPair: PVTabPairFocus? = nil)` | `PVInput.swift` | `tabPair` set → `equalsFocus = tabPair.input`, background `PVInputTabHook(onTab: tabPair.textArea.focus)` |
| Input router | `PVInputTabRouter` *(new)*, `PVInputTabHook` *(new)* | `Components/Input/PVInputTabRouter.swift` | D-5, D-6 |
| Text area | `PVTextArea(tabPair: PVTabPairFocus? = nil)` | `PVTextArea.swift` | `tabPair` set → `tabPair.textArea.apply(...)` replaces `.focused($isFocused)`; chrome `isFocused: tabPair.textArea.isActive()`; `PVTextEditorInsetClearer(onBacktab:)` passes the closure to its view's router through `updateNSView` |

```swift
enum PVTabPairField: Hashable { case input, textArea }

struct PVTabPairFocus {
    let input: PVInputFocus
    let textArea: PVInputFocus
}

/// Owns focus for one PVInput → PVTextArea pair: Tab in the input moves to
/// the text area, Shift-Tab in the text area moves back.
struct PVTabPair<Content: View>: View {
    @FocusState private var field: PVTabPairField?
    @ViewBuilder let content: (PVTabPairFocus) -> Content
    var body: some View { content(PVTabPairFocus(input: …(.input), textArea: …(.textArea))) }
}

final class PVInputTabRouter: NSObject, NSTextFieldDelegate {
    var onTab: (() -> Void)?
    @MainActor func attach(to field: NSTextField)
    func control(_ control: NSControl, textView: NSTextView, doCommandBy selector: Selector) -> Bool
    // responds(to:) / forwardingTarget(for:) as PVTextAreaTabRouter
}

struct PVInputTabHook: NSViewRepresentable {
    let onTab: () -> Void
    static func matchingField(_ candidates: [(field: NSTextField, frame: CGRect)], hookFrame: CGRect) -> NSTextField?
}
```

Call-site shape (P2):

```swift
PVTabPair { pair in
    PVField(label: L10n.CitationComposer.newSubjectLabel, error: …) {
        PVInput(text: newSubjectLabelBinding, size: .sm, tabPair: pair)
    }
    PVField(label: L10n.CitationComposer.descriptionLabel) {
        PVTextArea(text: newSubjectDescriptionBinding, lineLimit: 1...3, tabPair: pair)
    }
}
```

States: P1 without input (D-9). Accessibility: VoiceOver order and labels are unchanged; the move is the same focus change a click makes.

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — Focus handle and pair owner**
Files: `macos/App/DesignSystem/Components/Input/PVInput.swift`, `macos/App/DesignSystem/Components/Input/PVTabPair.swift` *(new)*. Skill: `add-ui-component`.
Do: D-2 on `PVInputFocus`; add `PVTabPairField`, `PVTabPairFocus`, `PVTabPair` per §6. Doc comment on `PVTabPair` says when to use it.
Check: the app target builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`) with no new warning.

**STEP-2 — Text area joins a pair**
Files: `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`.
Do: `tabPair:` init parameter; focus and chrome from `tabPair.textArea` (D-4); `activateOnAppear` per D-8; `PVTextAreaTabRouter.onBacktab: (() -> Void)?` — set → Shift-Tab calls it on the next turn and, one turn after that, calls `selectPreviousKeyView` when `tabPair.input.isActive()` is false (D-9); nil → today's `selectPreviousKeyView`. `PVTextEditorInsetClearer` takes `onBacktab` and sets it on the router in `updateNSView`.
Check: builds; `PVTextAreaTests` pass unchanged.

**STEP-3 — Input sends Tab to its pair**
Files: `macos/App/DesignSystem/Components/Input/PVInput.swift`, `macos/App/DesignSystem/Components/Input/PVInputTabRouter.swift` *(new)*.
Do: `PVInputTabRouter` and `PVInputTabHook` per D-5, D-6, with the re-attach recheck copied from `PVTextEditorInsetView.scheduleTabRouterRecheck` (at most 2 rechecks). `PVInput(tabPair:)` per D-5, D-7, D-8; the hook is a `.background` of the field only when `tabPair` is set.
Check: builds.

**STEP-4 — Tests**
Files: `macos/ProvenenciaTests/PVInputTabRouterTests.swift` *(new)*, `macos/ProvenenciaTests/PVTextAreaTests.swift`. Skill: `add-swift-test`.
Do: T-1 – T-5.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-5 — Wire the graph card and the composer (P1–P3)**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/CitationComposer/CitationComposerView.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`.
Do: wrap each pair in `PVTabPair` and pass `tabPair: pair` to both fields. P1: wrap the `if !model.isEditingBridge { … }` field and the Description field together (D-9). P3: `labelField` and `transcriptionColumn(rows:)` take `tabPair: PVTabPairFocus?`; `sideForm` wraps both in `PVTabPair`, `panelForm` passes nil (D-10). Keep every `.accessibilityIdentifier`.
Check: builds; manual §10 items 1–3 pass.

**STEP-6 — Wire Properties and the Source page (P4–P7)**
Files: `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/Features/Sources/SourceGraphFormDialog.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`.
Do: as STEP-5. P4: the pair wraps Label, Key and Description; only Label and Description take `tabPair`. P5: wrap the body's `VStack` content so the inline editor's input and the text area below share one pair. P7: keep `.onSubmit { commitEdit(art) }` on the input.
Check: builds; `rg -n "PVTextArea\(" macos/App/Features` — every hit in P1–P7 has `tabPair:` within its argument list; manual §10 items 4–7 pass.

**STEP-7 — Say how a pair is wired**
Files: `macos/App/DesignSystem/README.md` (§ Interaction state), `.agents/skills/add-ui-component/SKILL.md`.
Do: one short paragraph in the README: a `PVInput` followed by a `PVTextArea` in one form is wrapped in `PVTabPair` and both take `tabPair:`, so Tab and Shift-Tab move between them. One line in `add-ui-component` pointing to it.
Check: `python3 scripts/check-guidance-names.py` passes.

**STEP-8 — Version and docs**
Files: `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj`; `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row), this plan, `ux/2026-10-15-new-subject-card-tab-skips-description.md` (delete) and its row in the map's "Where the notes went".
Do: patch bump together; DoD docs lines.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVInputTabRouterTests.swift` `tabCallsOnTab` *(new)* | `NSTextField` with a spy `NSTextFieldDelegate`; router attached with `onTab` counting; `control(_:textView:doCommandBy: insertTab)` → returns `true`, count 1, spy saw nothing | AC-1.1, D-5 |
| T-2 | Swift | same file, `otherCommandsReachTheOriginalDelegate` *(new)* | `insertBacktab`, `insertNewline`, `insertTabIgnoringFieldEditor` → each returns the spy's answer (`false`), spy records all three in order, `onTab` count 0 | AC-3.3, AC-3.4 |
| T-3 | Swift | same file, `editsStillReachTheOriginalDelegate` *(new)* | `controlTextDidChange(_:)` through the router → spy change count 1 | D-5 |
| T-4 | Swift | same file, `matchingFieldPicksTheFieldUnderTheHook` *(new)* | two fields at `(0,0,200,24)` and `(0,40,200,24)`, hook frame `(0,40,200,24)` → the second; hook frame `(300,300,10,10)` → nil | D-6 |
| T-5 | Swift | `macos/ProvenenciaTests/PVTextAreaTests.swift` `backtabCallsOnBacktabWhenSet` *(new)* | router with `onBacktab` set: `insertBacktab` → returns `true`, `onBacktab` called once on the next main-queue turn; `insertTab` → returns `true`, `onBacktab` not called; router with `onBacktab` nil → existing tests unchanged | AC-2.1, AC-3.1, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVTextAreaTests` (router changes), any snapshot or test that builds `PVInput` with `focus:`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches today's look; no new kit primitive beyond PVTabPair and the routers named in §6
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Dogfood note ux/2026-10-15-new-subject-card-tab-skips-description.md deleted; its "Where the notes went" row says landed
- [ ] DesignSystem README and add-ui-component carry the pair rule (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Any project with one Source, one Evidence graph and one Artifact. Default window, then 2560pt wide.

1. **P1.** Open the graph, arm Person, click the canvas. Type `Mary`, Tab → caret in Description. Type `born 1820`, Shift-Tab → caret in Label, both values kept. Tab twice → focus leaves Description as before. Edit a bridge → Description works, Shift-Tab does not crash (AC-1.1, AC-2.1, AC-3.1, AC-3.5).
2. **P2.** Open the Citation composer, open New subject. Type `Ship`, Tab → Description; Shift-Tab → Label (AC-1.2, AC-2.2).
3. **P3.** In the composer at the default window (side form), caret in Description (the Citation label), Tab → Transcription; Shift-Tab → back. Option-Tab in Transcription inserts a tab; Return inserts a newline (AC-1.3, AC-3.2).
4. **P4.** Properties → New property. Label, Tab → Description (Key skipped); Shift-Tab → Label (AC-1.4, AC-2.2).
5. **P5.** Properties, select a researcher Property, edit its identity. Label, Tab → Description; Shift-Tab → Label (AC-1.3, AC-2.2).
6. **P6.** Source page → New graph. Name, Tab → Description; Shift-Tab → Name (AC-1.3, AC-2.2).
7. **P7.** Source page, edit an Artifact card. Label, Tab → Description; Shift-Tab → Label; Return in Label commits (AC-1.3, AC-2.2, AC-3.4).
8. Any other `PVInput` (omnibar, Properties search): Tab and Shift-Tab behave as before (AC-3.6).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| SwiftUI reinstalls the `NSTextField` delegate after the two rechecks | Same recheck as the text area; manual check item 1 shows it |
| The one-turn focus write is dropped while the field editor finishes the Tab | Use the two-turn write from 5d0657dd (`CitationComposerAuthoringBox.commit…`: write, then on the next turn write again when focus is not the target). This is the only allowed variation |
| `PVTabPair`'s body changes stack spacing | Its body returns the content unchanged; manual check compares each form with `main` |

**Stop and report if:**

- A SwiftUI `TextField` in `PVInput` is not an `NSTextField`, or `control(_:textView:doCommandBy:)` is never called for Tab on the router (log once in the delegate and press Tab).
- After the two-turn write, Tab from the input still does not land in the text area in P1 or P2.
- `PVTabPair` changes the spacing or layout of any of P1–P7 compared with `main`.
- A `FocusState` owned by `PVTabPair` inside a sheet does not move focus (P1, P2, P4 or P6).
- `PVInputFocus` cannot drop `fileprivate` on `isActive` / `focus` without a compile error outside the files named here.

## 12. Open questions

None.

## 13. PR

**Title:** S14-31 — Tab reaches a following text area
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-31-tab-reaches-text-area.md`.
