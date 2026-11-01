import SwiftUI

/// The one sheet presenter. Every kit sheet (`pvFormDialog`, `pvConfirm`, the
/// delete-impact notice) and every feature sheet presents through `.pvSheet`,
/// so Escape and a click outside the sheet follow one rule: ``PVSheetDismiss``.
/// A dirty sheet asks before it closes, from either route; the visible Cancel
/// button still closes at once. Menus and popovers are not this path.
///
/// Content closes its sheet through ``PVSheetClose`` (`@Environment(\.pvSheetClose)`),
/// so a row that closes the sheet and then navigates meets the same dirty rule.
///
/// The chrome wraps a sheet's content with both close routes and the discard question.
private struct PVSheetChrome<Content: View>: View {
    let isRunning: Bool
    let isDirty: Bool
    /// Prefix for the discard question's `.confirm` / `.cancel` ids.
    let accessibilityIdentifierPrefix: String?
    let close: () -> Void
    @ViewBuilder let content: () -> Content

    @State private var confirmingDiscard = false
    @State private var gate = PVDiscardGate()

    var body: some View {
        sheetBody
            .environment(\.pvSheetClose, PVSheetClose(run: { next in run(next) }))
            // A zero-size button owns Escape, so the rule lives in one place and
            // no visible Cancel needs its own key.
            .background {
                Button(action: dismiss) { EmptyView() }
                    .keyboardShortcut(.cancelAction)
                    .focusable(false)
                    .frame(width: 0, height: 0)
                    .opacity(0)
                    .accessibilityHidden(true)
            }
            .pvOutsideClick(isEnabled: !confirmingDiscard, perform: dismiss)
    }

    @ViewBuilder
    private var sheetBody: some View {
        if confirmingDiscard {
            PVConfirmContent(
                copy: .discardChanges,
                tone: .danger,
                accessibilityIdentifierPrefix: PVFormDialogAccessibility.identifier(
                    prefix: accessibilityIdentifierPrefix,
                    suffix: "discard"
                ),
                onConfirm: discard,
                onCancel: keepEditing,
                detail: { EmptyView() }
            )
        } else {
            content()
        }
    }

    /// Escape and a click outside: the dirty rule with no step to run after.
    private func dismiss() {
        run {}
    }

    /// Closes the sheet by the dirty rule, then runs `next`. A dirty sheet asks
    /// first and holds `next` behind the question; Keep editing drops it.
    private func run(_ next: @escaping @MainActor () -> Void) {
        switch PVSheetCloseRule.outcome(
            isRunning: isRunning, isDirty: isDirty, isConfirmingDiscard: confirmingDiscard,
            next: next, gate: &gate
        ) {
        case .nothing:
            break
        case .keepEditing:
            keepEditing()
        case .askToDiscard:
            confirmingDiscard = true
        case .closeThenRun:
            close()
            Task { @MainActor in next() }
        }
    }

    private func discard() {
        confirmingDiscard = false
        let step = gate.discard()
        close()
        Task { @MainActor in step?() }
    }

    private func keepEditing() {
        confirmingDiscard = false
        _ = gate.keepEditing()
    }
}

/// What a close request does to a sheet, apart from the views: the dirty rule
/// ``PVSheetDismiss`` picks the action, and a dirty sheet holds `next` in the
/// gate behind the "Discard changes?" question.
enum PVSheetCloseRule {
    enum Outcome: Equatable {
        case nothing
        case keepEditing
        case askToDiscard
        case closeThenRun
    }

    @MainActor
    static func outcome(
        isRunning: Bool,
        isDirty: Bool,
        isConfirmingDiscard: Bool,
        next: @escaping @MainActor () -> Void,
        gate: inout PVDiscardGate
    ) -> Outcome {
        switch PVSheetDismiss.action(isRunning: isRunning, isDirty: isDirty, isConfirmingDiscard: isConfirmingDiscard) {
        case .ignore:
            return .nothing
        case .keepEditing:
            return .keepEditing
        case .confirmDiscard:
            gate.hold(next)
            return .askToDiscard
        case .close:
            return .closeThenRun
        }
    }
}

/// What content calls to close its sheet: a dirty sheet asks "Discard changes?"
/// first, and `next` runs only once the sheet is closed. Keep editing, Escape on
/// the question, or a click outside it drops `next`.
struct PVSheetClose {
    let run: @MainActor (@escaping @MainActor () -> Void) -> Void

    @MainActor
    func callAsFunction(then next: @escaping @MainActor () -> Void = {}) {
        run(next)
    }
}

private struct PVSheetCloseKey: EnvironmentKey {
    /// Outside a sheet (previews) nothing closes; the step runs.
    nonisolated(unsafe) static let defaultValue = PVSheetClose(run: { next in
        Task { @MainActor in next() }
    })
}

extension EnvironmentValues {
    var pvSheetClose: PVSheetClose {
        get { self[PVSheetCloseKey.self] }
        set { self[PVSheetCloseKey.self] = newValue }
    }
}

extension View {
    /// Sheet for a snapshot item. Escape and an outside click close it through
    /// ``PVSheetDismiss``; `isDirty` makes both ask before closing. Dismissal
    /// stays on the caller's binding so an async action can keep the sheet up
    /// while `isRunning`.
    func pvSheet<Item: Identifiable, Content: View>(
        item: Binding<Item?>,
        isRunning: Bool = false,
        isDirty: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        @ViewBuilder content: @escaping (Item) -> Content
    ) -> some View {
        sheet(item: item) { value in
            PVSheetChrome(
                isRunning: isRunning,
                isDirty: isDirty,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                close: { item.wrappedValue = nil }
            ) {
                content(value)
            }
        }
    }

    /// Sheet for a flag. Same rules as ``pvSheet(item:isRunning:isDirty:accessibilityIdentifierPrefix:content:)``.
    func pvSheet<Content: View>(
        isPresented: Binding<Bool>,
        isRunning: Bool = false,
        isDirty: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) -> some View {
        sheet(isPresented: isPresented) {
            PVSheetChrome(
                isRunning: isRunning,
                isDirty: isDirty,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                close: { isPresented.wrappedValue = false },
                content: content
            )
        }
    }
}
