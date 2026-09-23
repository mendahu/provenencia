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
        pvFormDialog(
            isPresented: isPresented,
            copy: PVFormDialogCopy(
                title: mode == .add ? L10n.NameValue.titleAdd : L10n.NameValue.titleEdit,
                confirm: mode == .add ? L10n.NameValue.confirmAdd : L10n.NameValue.confirmSave,
                cancel: L10n.NameValue.cancel
            ),
            width: 560,
            confirmDisabled: !draft.wrappedValue.isValid,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: onConfirm
        ) {
            NameValueEditorForm(
                draft: draft,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix
            )
        }
    }
}
