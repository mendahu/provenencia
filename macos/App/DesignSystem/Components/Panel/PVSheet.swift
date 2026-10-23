import SwiftUI

/// The one sheet presenter. Every kit sheet (`pvFormDialog`, `pvConfirm`, the
/// delete-impact notice) and every feature sheet presents through `.pvSheet`,
/// so Escape and a click outside the sheet follow one rule: ``PVSheetDismiss``.
/// A dirty sheet asks before it closes, from either route; the visible Cancel
/// button still closes at once. Menus and popovers are not this path.
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

    var body: some View {
        sheetBody
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
                copy: PVConfirmCopy(
                    title: L10n.string(L10n.DesignSystem.formDialogDiscardTitle),
                    message: L10n.string(L10n.DesignSystem.formDialogDiscardMessage),
                    confirm: L10n.DesignSystem.formDialogDiscardConfirm,
                    cancel: L10n.DesignSystem.formDialogDiscardKeep
                ),
                tone: .danger,
                accessibilityIdentifierPrefix: PVFormDialogAccessibility.identifier(
                    prefix: accessibilityIdentifierPrefix,
                    suffix: "discard"
                ),
                onConfirm: close,
                onCancel: { confirmingDiscard = false },
                detail: { EmptyView() }
            )
        } else {
            content()
        }
    }

    private func dismiss() {
        switch PVSheetDismiss.action(isRunning: isRunning, isDirty: isDirty, isConfirmingDiscard: confirmingDiscard) {
        case .ignore:
            break
        case .keepEditing:
            confirmingDiscard = false
        case .confirmDiscard:
            confirmingDiscard = true
        case .close:
            close()
        }
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
