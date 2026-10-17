# S10-56 — Hold Shift to select; Select disarms after one rectangle

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 5 — Artifacts and multi-select
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-56--hold-shift-to-select)
**Brief:** — (no new chrome: the palette Select button, its armed paint, its help, and the rectangle already exist from S10-D6; this PR changes when they show and what a click does)
**Board:** —
**Depends on:** S10-20 (landed [#396](https://github.com/mendahu/provenencia/pull/396)), S10-51 (landed [#421](https://github.com/mendahu/provenencia/pull/421))
**Blocks:** —
**Branch:** `s10-56-quick-select`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** working rule "Box-select is a marquee of subject cards." in [`../deployment-plan.md`](../deployment-plan.md#working-rules), [S10-20 plan](S10-20-open-file-box-select.md), [`macos/App/Features/GraphCanvas/README.md`](../../../../macos/App/Features/GraphCanvas/README.md) (GraphCanvas stays free of Evidence types)
**Planned:** 2026-10-17 against `main` @ `1c99ef4`
**Review:** [`S10-56-quick-select.review.md`](S10-56-quick-select.review.md) — Fits, Ready with changes (applied); 2 rounds

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

Multi-select on the Evidence graph is slow to use: the researcher clicks Select, draws a rectangle, then clicks Select again to get click-drag pan back. After this PR, holding Shift engages Select (the palette button lights) and releasing Shift returns to pan. Shift-click adds a card to the selection or takes it out, so a group can be built one card at a time and then dragged together. Select armed from the palette disarms itself when the rectangle is released.

**Why now.** The user asked for it on 2026-10-17 after using S10-20's box-select on real graphs.

**What stays.** A click on empty canvas clears the selection. Escape and a second click of Select still disarm. Dragging a selected card still moves the whole selection. A rectangle drawn from the palette (no Shift) still replaces the selection. Place and Connect still win over Select. The help copy "Hold Shift to select" is unchanged.

## 2. User stories

**US-1.** As a researcher, when I hold Shift over the graph, Select engages, and when I let go, empty-canvas drag pans again, so that I switch between selecting and panning without the palette.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | No tool armed, graph window key, no text field focused. Press Shift. | `model.inputMode == .selecting`; the palette Select button paints armed. |
| AC-1.2 | AC-1.1, then release Shift. | `inputMode == .idle`; Select paints unarmed; an empty-canvas drag pans. |
| AC-1.3 | Place person armed (or Connect armed). Press Shift. | `inputMode` stays `.placing(.person)` (or `.connecting`); Select does not paint armed. |
| AC-1.4 | Select armed from the palette. Press and release Shift. | `armedSelect` is still `true`; `inputMode == .selecting` throughout. |
| AC-1.5 | Shift held, then the window resigns key (Cmd-Tab away) with Shift still down, then Shift released in the other app. | On resign, `inputMode == .idle`; returning to the window does not leave Select stuck on. |
| AC-1.6 | The omnibar (or any `NSText` field editor) is first responder. Type a capital letter. | Select does not paint armed; `inputMode` is unchanged. |
| AC-1.7 | The omnibar is first responder. Shift-click a card on the canvas. | The card toggles, and Select paints armed while Shift stays down. |

**US-2.** As a researcher, when I arm Select on the palette and draw one rectangle, Select disarms on release, so that the next empty-canvas drag pans.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Select armed from the palette, Shift up. Drag a rectangle over `p1` and `p2`, release. | Selection is `{p1, p2}`; `armedSelect == false`; `inputMode == .idle`. |
| AC-2.2 | Select armed. Drag a rectangle over empty canvas only, release. | Selection is empty; `armedSelect == false`. |
| AC-2.3 | Select armed. Click a card (no drag). | That card is the selection; `armedSelect` stays `true`. |
| AC-2.4 | Select armed. Click empty canvas (no drag). | Selection is empty; `armedSelect` stays `true`. |

**US-3.** As a researcher, when I Shift-click cards one by one, each one joins the selection or leaves it, so that I can drag the group afterwards.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Selection `{p1}`. Shift-click `p2`. | Selection `{p1, p2}`. |
| AC-3.2 | Selection `{p1, p2}`. Shift-click `p1`. | Selection `{p2}`. |
| AC-3.3 | Selection `{p1}`. Shift-click bridge `b1`. | Selection `{p1, b1}`. |
| AC-3.4 | Selection `{p1, p2}`, Shift released. Drag `p1` by one cell. | `p1` and `p2` both move one cell and persist (existing `beginGroupDrag` / `commitGroupDrag`). |
| AC-3.5 | Shift-click lands on a card's nested action (edit pencil, Add property). | The card is toggled; the action does not run. |
| AC-3.6 | Selection `{p1}`, Shift held. Drag a rectangle over `p2` and `p3`, release. | Selection `{p1, p2, p3}`. |
| AC-3.7 | `p1` is activated (`activatedSubjectID == "p1"`), selection `{p1}`. Shift-click `p1`. | Selection empty; `activatedSubjectID == nil`. |

**US-4.** As a researcher, clicking empty canvas still clears the selection.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Selection `{p1, p2}`. Click empty canvas, Shift up. | Selection empty. |
| AC-4.2 | Selection `{p1, p2}`. Shift held, click empty canvas (no drag). | Selection empty. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Shift held engages Select; release disengages; Place and Connect win | STEP-1, STEP-3, STEP-4, AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| Shift state clears when the window resigns key; ignored while an `NSText` is first responder, until a Shift-press on the canvas | STEP-3, AC-1.5, AC-1.6, AC-1.7 |
| Palette Select paints armed while Shift engages it | STEP-4, AC-1.1 |
| A rectangle release disarms palette-armed Select, whatever it caught | STEP-1, STEP-2, AC-2.1, AC-2.2 |
| Shift-click on a card toggles it in the selection (bridges included); skips the card action | STEP-1, STEP-2, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.7 |
| A Shift rectangle adds to the selection; a no-Shift rectangle replaces it | STEP-1, STEP-2, AC-3.6, AC-2.1 |
| Empty-canvas click clears, with or without Shift | STEP-2, AC-4.1, AC-4.2 |
| Working rule and canvas README updated | STEP-5 |

| Out | Home |
| --- | --- |
| Shift-drag of an unselected card adding it before the drag (it stays: the drag replaces the selection with that card, as today) | not planned |
| Cmd-click toggle | not planned |
| A keyboard / VoiceOver way to add one card to the selection | dogfood log |
| New copy, a new palette hint, or a banner while Shift is held | not planned |
| Shift engaging Select on any canvas other than the Evidence graph | not planned (no other host today) |
| A shared "window event watch" helper for local monitor + resign-key observer (`PVComboBox`, `PVContextMenu`, this PR) | logged in the review log for a design-system hardening pass; not planned in Spike 10 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Pointer gestures | `macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift` `GraphCanvasPointerController` | `mouseDown(documentPoint:windowPoint:shiftDown:)` stores `shiftDown` in `Gesture.pendingClick`. `mouseDragged` turns an empty-canvas `pendingClick` into `.marquee(origin:current:)` when `mode == .selecting || shiftDown`, else pans. `mouseUp` in `.pendingClick` for `.idle, .selecting` calls `onSelect(targetID)` (and `onCardAction` on an action zone) or `onDeselect()`; it ignores `shiftDown`. `.marquee` release calls `onMarqueeEnded(Set<String>)`. |
| Tool mode | `macos/App/Features/GraphCanvas/GraphCanvasHitTarget.swift` `GraphCanvasPointerToolMode` | `.idle, .placing, .connecting, .selecting`. No change needed. |
| Document view | `macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift` `GraphCanvasDocumentView` | `acceptsFirstResponder` is `false`, so it never gets `flagsChanged`. Reads `event.modifierFlags.contains(.shift)` only on `mouseDown`. `viewDidMoveToWindow` updates tracking areas. |
| Model state | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `armedSelect`, `toggleSelect()`, `disarm()`, `disarmCiteTools()`, `selectSubject(id:)`, `replacePrimarySelection(_:)`, `handleKey(.escape)` | `armedSelect` is the only Select state. `replacePrimarySelection` intersects with the snapshot's subject and bridge ids, sets `selectedSubjectIDs`, clears `activatedSubjectID` when it fell out. Nothing disarms Select after a rectangle. |
| Input mode | `macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift` `EvidenceGraphModel.inputMode` | `isSheetPresented` → `.idle`; `armedConnect` → `.connecting`; `armedKind` → `.placing`; then `armedSelect ? .selecting : .idle`. |
| Palette | `macos/App/Features/EvidenceGraph/EvidenceGraphPalette.swift` `selectTool` | `armed: model.armedSelect`. |
| View wiring | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `wirePointer()`, `syncPointerMode()` | `pointer.onMarqueeEnded = { ids in model.replacePrimarySelection(ids) }`; `onSelect` → `selectSubject(id:)`; `onDeselect` → `selectSubject(id: nil)`. `.onChange(of: inputMode)` calls `syncPointerMode()`. |
| Tests | `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift` `shiftDragOnEmptyCanvasMarquees`, `selectingModeMarqueesWithoutShift`, `marqueeReleaseClearsTheRect`, `marqueeFollowsADocumentOriginShift` | `marqueeReleaseClearsTheRect` assigns `onMarqueeEnded = { delivered = $0 }` (one argument); it changes with D-3. |
| Tests | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `groupModel()`, `rectangleSelectionKeepsBridges`, `groupDragKeepsTheSet`, `selectStaysAvailableWithoutAnArtifact`, `selectAndTheCiteToolsDisarmEachOther` | `groupModel()` builds persons `p1` (2,3), `p2` (4,3), `p3` (8,3) and bridge `b1` (3,3). New model tests use it. |

**Precedent.** S10-20 ([#396](https://github.com/mendahu/provenencia/pull/396)) added the marquee: a callback on `GraphCanvasPointerController`, wired in `EvidenceGraphView.wirePointer()` to an `EvidenceGraphModel` method, with pointer tests in `GraphCanvasPointerHitTestingTests` (`marqueeController(mode:)`) and model tests on `groupModel()`. Copy that shape. The local event monitor follows `PVFormDialog`'s `installMonitor()` / `removeMonitor()` (`macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`): `NSEvent.addLocalMonitorForEvents`, `MainActor.assumeIsolated`, return the event, remove on teardown.

## 5. Decisions

All product calls (D-4, D-5, D-6, D-7) were answered by the user on 2026-10-17.

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Shift state is a transient model flag `EvidenceGraphModel.shiftSelectHeld` *(new)*, separate from `armedSelect`. `inputMode` returns `.selecting` when `armedSelect || shiftSelectHeld`, after the existing sheet / Connect / Place checks. | Release must restore exactly what was there before; folding Shift into `armedSelect` would disarm a palette-armed Select on release (breaks AC-1.4). Ordering keeps Place and Connect winning (AC-1.3). | Writing `armedSelect` on Shift — because release cannot tell who armed it. |
| D-2 | Shift is observed in `GraphCanvasDocumentView` with a local `NSEvent` monitor for `.flagsChanged`, installed in `viewDidMoveToWindow` when `window != nil` and removed when it is `nil` and in `deinit`. It only reports events whose `event.window === window`. It reports `false` instead of the flag when `window?.firstResponder is NSText`. It also reports `false` on `NSWindow.didResignKeyNotification` for its window. It forwards to `pointer.setShiftHeld(_:)` *(new)*, which dedupes and calls `onShiftHeldChanged: ((Bool) -> Void)?` *(new)*. | The document view refuses first responder, so `flagsChanged(with:)` never reaches it; a monitor does. The plumbing is generic, so it lives in GraphCanvas, and the Evidence policy (D-1) stays in the model. The `NSText` guard keeps typing capitals in the omnibar from lighting Select (AC-1.6). Resign-key reset prevents a stuck Shift (AC-1.5). | Overriding `flagsChanged` on the document view — never called. A SwiftUI `onModifierKeysChanged` on the graph — needs focus and does not see the window-level state reliably on macOS 14. An app-wide monitor in `ProvenenciaApp` — puts canvas policy in app setup. |
| D-3 | `onMarqueeEnded` becomes `((Set<String>, Bool) -> Void)?`, the `Bool` being `additive`. `Gesture.marquee` gains `additive: Bool`, set from the `pendingClick`'s `shiftDown`. | The pointer already knows Shift at press time; carrying it in the gesture is exact even if Shift is released mid-drag. | Reading `model.shiftSelectHeld` at release — wrong when Shift is let go before mouse-up. |
| D-4 | A Shift rectangle unions with the selection; a no-Shift rectangle replaces it. | User answer. Lets Shift-click and Shift-rectangle build one group. | Always replace. |
| D-5 | Every rectangle release sets `armedSelect = false`, including one that caught nothing. Plain clicks (card or empty canvas) with Select armed do not disarm. | User answer: one rectangle per arm. A click is not a rectangle. | Disarm only when the rectangle selected something. |
| D-6 | Shift-click on a card (bridge included) toggles it: in the selection → removed, not in → added. On a nested action zone, Shift-click toggles and does not call `onCardAction`. The pointer calls a new `onToggleSelect: ((String) -> Void)?` *(new)* instead of `onSelect`. | User answer (toggle). Skipping the action keeps Shift-click one meaning. | Add-only. Running the action as well — a Shift-click on the pencil would open the edit sheet mid-selection. |
| D-7 | A click on empty canvas clears the selection whether or not Shift is down (pointer still calls `onDeselect`). | User requirement 4, stated without exception. | Finder's "Shift-click empty keeps the selection". |
| D-8 | Model methods: `toggleSubjectInSelection(id:)` *(new)* and `finishRectangleSelection(_ ids: Set<String>, additive: Bool)` *(new)*. `finishRectangleSelection` computes `additive ? selectedSubjectIDs.union(ids) : ids`, passes it to the existing `replacePrimarySelection(_:)` (which filters to cards on the graph and clears a stale activation), then sets `armedSelect = false`. `toggleSubjectInSelection` builds the toggled set and passes it to `replacePrimarySelection(_:)` too. `setShiftSelectHeld(_ held: Bool)` *(new)* assigns `shiftSelectHeld` plainly; the pointer's `setShiftHeld(_:)` is the one place that dedupes (T-5). | One filter and one activation rule for every selection path; `replacePrimarySelection` already owns both. | Duplicating the snapshot filter in each new method. |
| D-9 | The palette reads `armed: model.inputMode == .selecting`, inline like its siblings (`armed: model.armedConnect`, `armed: model.armedKind == kind`). | `inputMode` is the one precedence source; the button lights for Shift and for the palette alike, and not while Place or Connect is armed. While a sheet is up the button reads unarmed, which matches the canvas being idle then. | Reading `armedSelect || shiftSelectHeld` in the palette — duplicates `inputMode`'s precedence. A named wrapper (`inputMode == .selecting`) — one caller, no policy. |
| D-10 | Tests: pointer behavior in `GraphCanvasPointerHitTestingTests` (Swift Testing, pure controller); model behavior in `EvidenceGraphModelTests` on `groupModel()`. The monitor itself is checked manually (§10). | Matches S10-20's split; an `NSEvent` monitor cannot be driven from unit tests without synthesizing events. | XCUITest — not used in this repo. |
| D-11 | Press-time Shift (`shiftDown` in `Gesture.pendingClick`) stays the only signal for pointer behavior (marquee, additive, toggle), and it is not guarded by text focus: a Shift-press on the canvas is canvas intent. `GraphCanvasDocumentView.mouseDown` calls `pointer.setShiftHeld(event.modifierFlags.contains(.shift))` before `pointer.mouseDown(…)`, so the held state (paint) agrees with the press even when a text field kept focus; the next `flagsChanged` reports through D-2's guard again. | Without it, with the omnibar focused, a Shift-drag draws a rectangle while Select stays dark (the canvas never takes first responder). Guarding the press too would make Shift-click silently fail after typing in the omnibar. | Applying the `NSText` guard to `mouseDown` — breaks Shift-click after the omnibar. Dropping the guard — capitals typed in the omnibar flash Select. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `GraphCanvasPointerController.onToggleSelect` *(new)* | callback | `onSelect` | `onSelect` means "replace with this one"; the host must tell the two apart. Searched `onSelect`, `toggle` in `GraphCanvas/`. |
| `GraphCanvasPointerController.setShiftHeld(_:)`, `shiftHeld` *(new, private(set))*, `onShiftHeldChanged` *(new)* | func / property / callback | `setPointerInside(_:)` | Same shape (pointer-level state set by the document view); a separate flag because it is a different input. |
| `GraphCanvasDocumentView` flags monitor *(new private members `flagsMonitor`, `resignKeyObserver`; new `static func reportedShift(_:textHasFocus:)`)* | private / func | `PVComboBox.startWatching(parent:)` (local monitor + `didResignKeyNotification` observer), `PVFormDialogOutsideClickView.installMonitor()` (`nonisolated(unsafe)` token, `deinit`) | Copies that pairing; no shared helper exists (searched `addLocalMonitorForEvents`: 9 call sites in `macos/App`, each private). This is the third monitor + resign-key copy; a shared window-event watch is the Out line below. |
| `EvidenceGraphModel.shiftSelectHeld`, `setShiftSelectHeld(_:)` *(new)* | property / func | `armedSelect` | See D-1. |
| `EvidenceGraphModel.toggleSubjectInSelection(id:)` *(new)* | func | `replacePrimarySelection(_:)` | Wraps it (D-8). |
| `EvidenceGraphModel.finishRectangleSelection(_:additive:)` *(new)* | func | `replacePrimarySelection(_:)` | Wraps it and adds the disarm (D-5, D-8). |

**Builds on.** The S10-20 pointer-callback → model-method wiring in `EvidenceGraphView.wirePointer()`; `inputMode` precedence in `EvidenceCanvasInputMode.swift`; `replacePrimarySelection` as the single selection writer.
**Deletes.** The view's `onMarqueeEnded` closure that calls `replacePrimarySelection` directly is replaced by one that calls `finishRectangleSelection`. Nothing else becomes dead.
**Next adds.** A future canvas host (family tree in `macos/App/Features/GraphCanvas/README.md`) gets the Shift signal and the Shift-click toggle by wiring `onShiftHeldChanged` and `onToggleSelect`; it supplies its own engage precedence, as `EvidenceCanvasInputMode.swift` does here.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None —

### macOS

**`GraphCanvasPointerController`** (`macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift`)

```swift
private(set) var shiftHeld = false                     // (new)
var onShiftHeldChanged: ((Bool) -> Void)?              // (new)
var onToggleSelect: ((String) -> Void)?                // (new)
var onMarqueeEnded: ((Set<String>, Bool) -> Void)?     // (changed: + additive)

func setShiftHeld(_ held: Bool)                        // (new) no-op when unchanged; else set + onShiftHeldChanged?(held)

private enum Gesture {
    case marquee(origin: CGPoint, current: CGPoint, additive: Bool)   // (changed)
    …
}
```

- `mouseDragged`, `.pendingClick` empty-canvas branch: `.marquee(origin: start, current: documentPoint, additive: shiftDown)`. The `.marquee` case in `mouseDragged` and `shiftActiveDocumentOrigin(by:)` carries `additive` through unchanged.
- `mouseUp`, `.marquee`: `onMarqueeEnded?(ids, additive)`.
- `mouseUp`, `.pendingClick(_, targetID, shiftDown)`, `.idle, .selecting`: if `targetID != nil && shiftDown` → `onToggleSelect?(targetID)` and return (no `onSelect`, no `onCardAction`). Otherwise the current code, unchanged. `targetID == nil` → `onDeselect?()` regardless of `shiftDown` (D-7).
- `.placing` and `.connecting` branches are unchanged.

**`GraphCanvasDocumentView`** (`macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift`)

- `mouseDown(with:)`: before `pointer.mouseDown(…)`, call `pointer.setShiftHeld(event.modifierFlags.contains(.shift))` (D-11).
- `viewDidMoveToWindow()`: after `updateTrackingAreas()`, call `installShiftTracking()` *(new, private)* when `window != nil`, else `removeShiftTracking()` *(new, private)*.
- Stored tokens: `nonisolated(unsafe) private var flagsMonitor: Any?` and `nonisolated(unsafe) private var resignKeyObserver: Any?`, the same annotation `PVFormDialogOutsideClickView.monitor` uses, because `deinit` touches them.
- `installShiftTracking()`: call `removeShiftTracking()` first. (The seed below sets `pointer.shiftHeld`; the host syncs its own state from `pointer.shiftHeld` when it wires `onShiftHeldChanged`, STEP-3.) Seed the state: `pointer.setShiftHeld(Self.reportedShift(NSEvent.modifierFlags, textHasFocus: window?.firstResponder is NSText))`, so Shift already held when the graph appears engages Select. Then `flagsMonitor = NSEvent.addLocalMonitorForEvents(matching: [.flagsChanged]) { [weak self] event in … }`; in the handler, `MainActor.assumeIsolated`, ignore events whose `event.window !== self?.window`, then `pointer.setShiftHeld(Self.reportedShift(event.modifierFlags, textHasFocus: self.window?.firstResponder is NSText))`; return the event. Then `resignKeyObserver = NotificationCenter.default.addObserver(forName: NSWindow.didResignKeyNotification, object: window, queue: .main) { [weak self] _ in MainActor.assumeIsolated { self?.pointer.setShiftHeld(false) } }` (block form; the token is what `removeObserver` takes).
- `static func reportedShift(_ flags: NSEvent.ModifierFlags, textHasFocus: Bool) -> Bool` *(new, internal so T-9 can call it)*: `!textHasFocus && flags.contains(.shift)`.
- `removeShiftTracking()`: `NSEvent.removeMonitor(flagsMonitor)` and `NotificationCenter.default.removeObserver(resignKeyObserver)` when non-nil, nil both, then `pointer.setShiftHeld(false)`.
- `deinit`: remove the monitor and the observer the way `PVFormDialogOutsideClickView.deinit` removes its monitor. It does not call `pointer` (not isolated); `viewDidMoveToWindow(nil)` already reset Shift before teardown.

**`EvidenceGraphModel`** (`macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`), next to `armedSelect`:

```swift
/// Shift is down over the graph (D-1). Engages Select without arming it.
private(set) var shiftSelectHeld = false                           // (new)
func setShiftSelectHeld(_ held: Bool)                              // (new)
func toggleSubjectInSelection(id: String)                          // (new) D-6, D-8
func finishRectangleSelection(_ ids: Set<String>, additive: Bool)  // (new) D-4, D-5, D-8
```

Update the `armedSelect` doc comment: "Select tool armed from the palette: an empty-canvas drag draws a rectangle, and its release disarms it (exclusive with Place and Connect). Holding Shift engages the same mode without arming it."

**`EvidenceCanvasInputMode.swift`**: `return (armedSelect || shiftSelectHeld) ? .selecting : .idle`. Update the `.selecting` case comment: "Select armed or Shift held: an empty-canvas drag draws a rectangle."

**`EvidenceGraphPalette.swift` `selectTool`**: `armed: model.inputMode == .selecting`.

**`EvidenceGraphView.wirePointer()`**:

```swift
pointer.onToggleSelect = { id in model.toggleSubjectInSelection(id: id) }
pointer.onMarqueeEnded = { ids, additive in model.finishRectangleSelection(ids, additive: additive) }
pointer.onShiftHeldChanged = { held in model.setShiftSelectHeld(held) }
```

`syncPointerMode()` already runs on `inputMode` change, so the pointer's `mode` follows Shift with no further wiring.

States: no new chrome. Accessibility: the palette button's armed accessibility value already follows its `armed` argument, so VoiceOver reads Select as on while Shift engages it.

### Strings

None — no new or changed copy. `L10n.EvidenceGraph.toolSelectHelp` ("Hold Shift to select") stays and becomes literally true.

## 7. Steps

**STEP-1 — Pointer: toggle callback, additive marquee, Shift state**
Files: `macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift`, `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift`. Skill: `add-swift-test`.
Do: §6 `GraphCanvasPointerController`. Update `marqueeReleaseClearsTheRect` to the two-argument closure and assert `additive == true`. Add T-1, T-2, T-3, T-4, T-5.
Also change `EvidenceGraphView.wirePointer()`'s `onMarqueeEnded` closure to `{ ids, _ in model.replacePrimarySelection(ids) }` so the app builds; STEP-2 replaces it.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/GraphCanvasPointerHitTestingTests` passes.

**STEP-2 — Model: toggle, rectangle finish, disarm after rectangle**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Do: add `toggleSubjectInSelection(id:)` and `finishRectangleSelection(_:additive:)` (D-8). Wire `onToggleSelect` and `onMarqueeEnded` in `wirePointer()`. Add T-6, T-7.
Check: `-only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes.

**STEP-3 — Shift engages Select**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift`, `macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: `shiftSelectHeld`, `setShiftSelectHeld(_:)`, the `inputMode` change, the document view monitor and the `mouseDown(with:)` call (§6, D-11), and `onShiftHeldChanged` wiring. `wirePointer()` sets `onShiftHeldChanged` and then calls `model.setShiftSelectHeld(pointer.shiftHeld)`, because the install-time seed in `viewDidMoveToWindow` can run before `.onAppear` wires the callback. Add T-8, T-9.
Check: `-only-testing:ProvenenciaTests/EvidenceGraphModelTests` and `-only-testing:ProvenenciaTests/GraphCanvasPointerHitTestingTests` pass; in the running app, holding Shift and dragging empty canvas draws a rectangle, and after release a drag pans; Cmd-Tab away with Shift down and back, then drag: it pans (§10 step 9, pointer half).

**STEP-4 — Palette paints Select while Shift engages it**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphPalette.swift`. Skill: `add-ui-component` (no new component; confirm `paletteTool` is reused as is).
Do: `armed: model.inputMode == .selecting`.
Check: build succeeds; §10 steps 1, 10, and 11 (the Select paint) behave as written.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-10/deployment-plan.md` (working rule "Box-select is a marquee of subject cards.", S10-56 map row and scope entry status), `docs/deployment-plan/spike-10/completed.md`, `macos/App/Features/GraphCanvas/README.md`, this plan (status, Landed note).
Do: per the Definition of done. Map PR-sequence row: append " — done, #N". Scope entry `**Plan:**` line: "— Landed", and add the same "Landed YYYY-MM-DD in #N, see completed.md" line under the heading that S10-20's entry has (copy its markup). `completed.md`: a section `## S10-56 — Hold Shift to select`, then the landed line in the form the other sections use, then one short paragraph. In the working rule, replace the four sentences from "Hold Shift and drag empty canvas" through "Bridge cards are not selected and do not move with the group." with: "Holding Shift engages Select and lights its palette button; releasing Shift returns to pan. Select armed on the palette (its help is “Hold Shift to select”) disarms when its rectangle is released. Person, event, place, and bridge cards that intersect the rectangle are selected; a Shift rectangle adds to the selection, a plain one replaces it. Shift-click adds or removes one card. Empty-canvas drag pans when Select is disarmed and Shift is up; a click on empty canvas clears the selection." (The bridge sentence is already false at the base: `replacePrimarySelection` keeps bridges, `rectangleSelectionKeepsBridges`.) In the GraphCanvas README: the `GraphCanvasPointerController` row in the type table gains "Shift tracking, Shift-click toggle"; the "Shift-drag on empty document" row says the pointer calls `onMarqueeEnded` with the ids of the hit targets it overlaps and whether Shift was down at press (additive); add a row "Shift-click on a hit target | `onToggleSelect` (no `onSelect`, no card action)" and a row "Shift key | Document view reports it through `onShiftHeldChanged` (local `flagsChanged` monitor; `false` while a text field has focus or the window is not key)". The "Click / drag on hit targets" row stays. In "Reuse contract", the parenthetical in the "Pointer `mode` + callbacks" cell becomes "select / toggle select / drag began / drag end / marquee ended (ids, additive) / Shift held changed / place / connect", and add one sentence under the table: "The document view reports Shift through `onShiftHeldChanged`; the host decides what it engages."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-56-quick-select.md` reports no errors.

Host: the Swift checks and §10 need macOS with Xcode (`scripts/check-pr.sh` runs `xcode-test` on macOS only). On a host without `xcodebuild`, report each Swift check and each §10 step as not run, with the reason; never as passed.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Shift-click adds or removes a card from the graph selection.", "Select disarms after its rectangle is released.", "Holding Shift engages Select on the Evidence graph."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerPressTests.shiftClickOnACardToggles` *(new)* | `marqueeController()`; `mouseDown` at (220,220) `shiftDown: true`, `mouseUp` same point → `onToggleSelect` gets `"card"`; `onSelect` not called. Same with `mode: .selecting` → same result. | AC-3.1, D-6 |
| T-2 | Swift | `GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerPressTests.shiftClickOnAnActionSkipsTheAction` *(new)* | `GraphCanvasPointerPressTests.controller()` (card (0,0,100,100), action `promote` at y 56…100); press and release at (50,70) Shift press/release inside the action → `onToggleSelect` called once, `onCardAction` not called. Without Shift → `onSelect` and `onCardAction` both called (today's behavior). | AC-3.5, D-6 |
| T-3 | Swift | `GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerPressTests.emptyClickDeselectsWithOrWithoutShift` *(new)* | `marqueeController()`; press/release at (10,10) with `shiftDown: false` then `true` → `onDeselect` called twice; `onToggleSelect` never. | AC-4.1, AC-4.2, D-7 |
| T-4 | Swift | `GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerPressTests.marqueeReportsAdditiveFromShift` *(new)*, plus `marqueeReleaseClearsTheRect` updated | Shift drag (10,10)→(80,90) → `additive == true`; `mode: .selecting`, no Shift, same drag → `additive == false`. | AC-3.6, AC-2.1, D-3 |
| T-5 | Swift | `GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerPressTests.shiftHeldReportsOnlyChanges` *(new)* | `marqueeController()`; `setShiftHeld(true)` twice, `setShiftHeld(false)` → `onShiftHeldChanged` receives `[true, false]`; `shiftHeld == false`. | AC-1.1, AC-1.2, D-2 |
| T-6 | Swift | `EvidenceGraphModelTests.swift` `shiftClickTogglesCardsInTheSelection` *(new)* | `groupModel()`; `selectSubject(id: "p1")`; toggle `p2` → `{p1,p2}`; toggle `b1` → `{p1,p2,b1}`; toggle `p1` → `{p2,b1}`; toggle `"ghost"` → unchanged. `activateSubject(id: "p3")`, toggle `p3` → empty and `activatedSubjectID == nil`. Then `replacePrimarySelection(["p1","p2"])`, `beginGroupDrag(anchor: "p1") == ["p1","p2"]`. | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.7, D-8 |
| T-7 | Swift | `EvidenceGraphModelTests.swift` `rectangleReplacesOrAddsAndDisarmsSelect` *(new)* | `groupModel()`; `toggleSelect()`; `finishRectangleSelection(["p1","p2"], additive: false)` → `{p1,p2}`, `armedSelect == false`. `selectSubject(id: "p3")`; `finishRectangleSelection(["p1"], additive: true)` → `{p1,p3}`. `toggleSelect()`; `finishRectangleSelection([], additive: false)` → empty, `armedSelect == false`. `toggleSelect()`; `selectSubject(id: "p1")`; `selectSubject(id: nil)` → `armedSelect == true`. | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-4, D-5 |
| T-8 | Swift | `EvidenceGraphModelTests.swift` `shiftEngagesSelectWithoutArmingIt` *(new)* | Setup as `selectAndTheCiteToolsDisarmEachOther` (`makeModel(store: makeStore())`, `await prepare()`). In order: (1) `setShiftSelectHeld(true)` → `inputMode == .selecting`, `armedSelect == false`; (2) `setShiftSelectHeld(false)` → `.idle`; (3) `toggleArm(.person)`, `setShiftSelectHeld(true)` → `.placing(.person)`; (4) `disarm()`, `setShiftSelectHeld(false)`, `toggleSelect()`, `setShiftSelectHeld(true)`, `setShiftSelectHeld(false)` → `armedSelect == true`, `.selecting`; (5) `setShiftSelectHeld(true)`, `finishRectangleSelection([], additive: false)` → `armedSelect == false`, `.selecting`; (6) `setShiftSelectHeld(false)` → `.idle`. | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1, D-9 |
| T-9 | Swift | `GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerHitTestingTests.shiftReportIgnoresTextFocus` *(new)* | `GraphCanvasDocumentView.reportedShift([.shift], textHasFocus: false)` → `true`; `([.shift], true)` → `false`; `([], false)` → `false`; `([.shift, .command], false)` → `true`. | AC-1.6, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `GraphCanvasPointerHitTestingTests` (marquee closure arity), `EvidenceGraphModelTests` `selectStaysAvailableWithoutAnArtifact` and `selectAndTheCiteToolsDisarmEachOther` (they read `inputMode`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; paletteTool reused)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the working rule from STEP-5; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: macos/App/Features/GraphCanvas/README.md pointer row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a Source with an artifact and at least three person cards and one relationship bridge.

1. Hold Shift: the palette Select button lights. Drag empty canvas: a rectangle draws. Release Shift: the button goes dark; drag empty canvas: it pans. (AC-1.1, AC-1.2)
2. Arm Place person, hold Shift: Select stays dark; click empty canvas: the create dialog opens. Cancel. (AC-1.3)
3. Click Select, hold and release Shift: Select stays lit. Draw a rectangle over two cards and release: both selected, Select goes dark, the next empty-canvas drag pans. (AC-1.4, AC-2.1)
4. Click Select, draw a rectangle over nothing: Select goes dark. (AC-2.2)
5. Click one card. Shift-click a second and a bridge: all three selected. Shift-click the first: it drops out. Release Shift, drag a selected card: the rest move with it and stay after reload. (AC-3.1, AC-3.2, AC-3.3, AC-3.4)
6. Shift-click a card's edit pencil: the card toggles and no sheet opens. (AC-3.5)
7. With two selected, hold Shift and draw a rectangle over a third: three selected. (AC-3.6)
8. Click empty canvas (with and without Shift): selection clears. (AC-4.1, AC-4.2)
9. Hold Shift, Cmd-Tab to another app, release Shift there, come back: Select is dark. (AC-1.5)
10. Click in the omnibar and type a capital letter: Select stays dark. (AC-1.6)
11. With the omnibar still focused, hold Shift and click a card: it toggles and Select lights; release Shift: Select goes dark. (AC-1.7)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The monitor leaks across graph opens (two monitors, double callbacks) | `installShiftTracking()` removes any existing monitor first; `setShiftHeld` dedupes; `viewDidMoveToWindow(nil)` removes it. |
| Shift stuck on after focus leaves | Resign-key observer reports `false` (AC-1.5). |
| A sheet (create/edit dialog) opens while Shift is held | `inputMode` already returns `.idle` while `isSheetPresented`; the Shift flag is ignored there. |

**Stop and report if:**

- `GraphCanvasDocumentView.acceptsFirstResponder` is `true` at your base, or something already overrides `flagsChanged` on the graph (D-2's reason is gone).
- `replacePrimarySelection(_:)` no longer filters to the snapshot's card ids or no longer clears a stale `activatedSubjectID` (D-8 relies on both).
- Any other caller of `onMarqueeEnded` exists besides `EvidenceGraphView` and the tests (`rg onMarqueeEnded macos`).
- The local `.flagsChanged` monitor does not fire while the pointer is over the canvas in the running app (manual step 1 fails) — report rather than switching to a polling or responder-chain approach.
- Making T-8 pass needs a change to `toggleSelect()`, `disarm()`, or `handleKey(.escape)`.

## 12. Open questions

None.

## 13. PR

**Title:** S10-56 — Hold Shift to select; Select disarms after one rectangle
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3, and US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-56-quick-select.md`.
