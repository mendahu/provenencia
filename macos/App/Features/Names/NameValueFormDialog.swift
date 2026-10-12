import SwiftUI

enum NameValueFormDialogMode: Equatable, Sendable {
    case add
    case edit
}

extension View {
    /// Shared NameValue modal. Hosts supply presentation + the working draft;
    /// titles, confirm verbs, and enablement come from the Names module.
    func nameValueFormDialog(
        isPresented: Binding<Bool>,
        draft: Binding<NameValueDraft>,
        mode: NameValueFormDialogMode,
        accessibilityIdentifierPrefix: String = "nameValue",
        onConfirm: @escaping () -> Void
    ) -> some View {
        modifier(NameValueFormDialogModifier(
            isPresented: isPresented,
            draft: draft,
            mode: mode,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: onConfirm
        ))
    }
}

private struct NameValueFormDialogModifier: ViewModifier {
    @Binding var isPresented: Bool
    @Binding var draft: NameValueDraft
    let mode: NameValueFormDialogMode
    let accessibilityIdentifierPrefix: String
    let onConfirm: () -> Void

    @State private var baseline: NameValueDraft?

    func body(content: Content) -> some View {
        content
            .pvFormDialog(
                isPresented: $isPresented,
                copy: PVFormDialogCopy(
                    title: mode == .add ? L10n.NameValue.titleAdd : L10n.NameValue.titleEdit,
                    confirm: mode == .add ? L10n.NameValue.confirmAdd : L10n.NameValue.confirmSave,
                    cancel: L10n.NameValue.cancel
                ),
                width: 560,
                isDirty: baseline.map { draft != $0 } ?? false,
                confirmDisabled: !draft.isValid,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                onConfirm: onConfirm
            ) {
                NameValueEditorForm(
                    draft: $draft,
                    accessibilityIdentifierPrefix: accessibilityIdentifierPrefix
                )
            }
            .onChange(of: isPresented, initial: true) { _, presented in
                if presented { baseline = draft }
            }
    }
}
