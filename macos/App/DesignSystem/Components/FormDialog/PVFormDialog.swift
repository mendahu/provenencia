import SwiftUI

/// Form dialog sheet — the macOS answer to `components/feedback/Dialog.jsx`
/// for short create/edit panels (Add Source, Add Artifact, icon pickers, etc.).
///
/// Same presentation family as the confirm sheet: a window-owned `.sheet`
/// (no custom scrim, corner radius, or shadow on the panel). Composes
/// ``PVPanel`` for chrome; the sunken footer band is content, carried over
/// from `Dialog.jsx`. Prefer ``PVConfirm`` for destructive / irreversible
/// confirmations — this type is for forms with a content slot.

struct PVFormDialogCopy {
    let title: LocalizedStringResource
    let subtitle: LocalizedStringResource?
    /// Already-localized subtitle, for copy that includes a ref or a name.
    let verbatimSubtitle: String?
    let confirm: LocalizedStringResource
    let cancel: LocalizedStringResource

    init(
        title: LocalizedStringResource,
        subtitle: LocalizedStringResource? = nil,
        verbatimSubtitle: String? = nil,
        confirm: LocalizedStringResource,
        cancel: LocalizedStringResource
    ) {
        self.title = title
        self.subtitle = subtitle
        self.verbatimSubtitle = verbatimSubtitle
        self.confirm = confirm
        self.cancel = cancel
    }
}

/// Pure enablement rules for the form-dialog footer — extracted so unit tests
/// can cover them without mounting SwiftUI.
enum PVFormDialogControls {
    static func isCancelDisabled(isRunning: Bool) -> Bool { isRunning }

    static func isConfirmDisabled(isRunning: Bool, confirmDisabled: Bool) -> Bool {
        isRunning || confirmDisabled
    }
}

/// Accessibility ids for form-dialog chrome buttons.
enum PVFormDialogAccessibility {
    /// `{prefix}.{suffix}` when a prefix is set; otherwise `nil` (no id).
    static func identifier(prefix: String?, suffix: String) -> String? {
        guard let prefix else { return nil }
        return "\(prefix).\(suffix)"
    }
}

/// Sheet body for a short form. Layout chrome comes from ``PVPanel``
/// (including the warm card fill). Corner radius, shadow, and scrim stay on
/// the sheet window.
struct PVFormDialogContent<Form: View>: View {
    let copy: PVFormDialogCopy
    let width: CGFloat
    let isRunning: Bool
    let confirmDisabled: Bool
    /// When set, confirm/cancel mint `{prefix}.confirm` / `{prefix}.cancel`.
    let accessibilityIdentifierPrefix: String?
    let onConfirm: () -> Void
    let onCancel: () -> Void
    @ViewBuilder let form: () -> Form

    @FocusState private var cancelFocused: Bool

    init(
        copy: PVFormDialogCopy,
        width: CGFloat = 480,
        isRunning: Bool = false,
        confirmDisabled: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        onCancel: @escaping () -> Void,
        @ViewBuilder form: @escaping () -> Form
    ) {
        self.copy = copy
        self.width = width
        self.isRunning = isRunning
        self.confirmDisabled = confirmDisabled
        self.accessibilityIdentifierPrefix = accessibilityIdentifierPrefix
        self.onConfirm = onConfirm
        self.onCancel = onCancel
        self.form = form
    }

    var body: some View {
        PVPanel(
            title: Text(copy.title),
            subtitle: formDialogSubtitle(copy),
            width: width,
            footerChrome: .sunken
        ) {
            form()
        } footer: {
            HStack(spacing: PVSpacing.space5) {
                Spacer(minLength: PVSpacing.space8)
                Button(L10n.string(copy.cancel)) { onCancel() }
                    .buttonStyle(.pv(.secondary, size: .lg))
                    .disabled(PVFormDialogControls.isCancelDisabled(isRunning: isRunning))
                    .focused($cancelFocused)
                    .modifier(FormDialogOptionalAccessibilityIdentifier(
                        prefix: accessibilityIdentifierPrefix,
                        suffix: "cancel"
                    ))
                Button(L10n.string(copy.confirm)) { onConfirm() }
                    .buttonStyle(.pv(.primary, size: .lg))
                    .keyboardShortcut(.defaultAction)
                    .disabled(PVFormDialogControls.isConfirmDisabled(
                        isRunning: isRunning,
                        confirmDisabled: confirmDisabled
                    ))
                    .modifier(FormDialogOptionalAccessibilityIdentifier(
                        prefix: accessibilityIdentifierPrefix,
                        suffix: "confirm"
                    ))
                    .overlay(alignment: .trailing) {
                        if isRunning {
                            ProgressView()
                                .controlSize(.small)
                                .offset(x: 22)
                        }
                    }
            }
        }
    }
}

private func formDialogSubtitle(_ copy: PVFormDialogCopy) -> Text? {
    if let verbatim = copy.verbatimSubtitle, !verbatim.isEmpty {
        return Text(verbatim: verbatim)
    }
    return copy.subtitle.map { Text($0) }
}

/// Applies `{prefix}.{suffix}` only when a prefix is provided.
private struct FormDialogOptionalAccessibilityIdentifier: ViewModifier {
    let prefix: String?
    let suffix: String

    func body(content: Content) -> some View {
        if let id = PVFormDialogAccessibility.identifier(prefix: prefix, suffix: suffix) {
            content.accessibilityIdentifier(id)
        } else {
            content
        }
    }
}

extension View {
    /// Sheet-based form dialog, presented through ``pvSheet(isPresented:isRunning:isDirty:accessibilityIdentifierPrefix:content:)``.
    /// Dismissal is left to the caller's binding so an async Create can stay on
    /// screen while it runs and keep field errors visible if validation or the
    /// store fails.
    ///
    /// Escape and a click outside the sheet close it; when `isDirty` is true
    /// either asks to discard or keep editing. The Cancel button still closes
    /// immediately. Menus and popovers are not this path.
    func pvFormDialog<Form: View>(
        isPresented: Binding<Bool>,
        copy: PVFormDialogCopy,
        width: CGFloat = 480,
        isRunning: Bool = false,
        isDirty: Bool = false,
        confirmDisabled: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        @ViewBuilder form: @escaping () -> Form
    ) -> some View {
        pvSheet(
            isPresented: isPresented,
            isRunning: isRunning,
            isDirty: isDirty,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix
        ) {
            PVFormDialogContent(
                copy: copy,
                width: width,
                isRunning: isRunning,
                confirmDisabled: confirmDisabled,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                onConfirm: onConfirm,
                onCancel: { isPresented.wrappedValue = false },
                form: form
            )
        }
    }
}

extension View {
    /// ``pvFormDialog(isPresented:copy:width:isRunning:isDirty:confirmDisabled:accessibilityIdentifierPrefix:onConfirm:form:)``
    /// for a form backed by a model the host sets: presented while `item` is
    /// non-nil. Escape, Cancel and a click outside clear `item`, except while
    /// `isRunning` (the caller's save in flight). `onConfirm` runs the save and
    /// answers whether it succeeded; on true the dialog clears `item`, on false
    /// it stays open so the model's errors stay visible.
    func pvFormDialog<Model, Form: View>(
        item: Binding<Model?>,
        copy: PVFormDialogCopy,
        width: CGFloat = 480,
        isRunning: Bool = false,
        isDirty: Bool = false,
        confirmDisabled: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping @MainActor (Model) async -> Bool,
        @ViewBuilder form: @escaping (Model) -> Form
    ) -> some View {
        pvFormDialog(
            isPresented: Binding(
                get: { item.wrappedValue != nil },
                set: { presented in
                    if !presented, !isRunning { item.wrappedValue = nil }
                }
            ),
            copy: copy,
            width: width,
            isRunning: isRunning,
            isDirty: isDirty,
            confirmDisabled: confirmDisabled,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: {
                guard let current = item.wrappedValue else { return }
                Task { @MainActor in
                    if await onConfirm(current) { item.wrappedValue = nil }
                }
            }
        ) {
            if let current = item.wrappedValue {
                form(current)
            }
        }
    }
}

#Preview("Form dialog — Add source") {
    PVFormDialogContent(
        copy: PVFormDialogCopy(
            title: "Add source",
            subtitle: "A thin record now — artifacts, notes and metadata live on the Source page.",
            confirm: "Create source",
            cancel: "Cancel"
        ),
        onConfirm: {},
        onCancel: {}
    ) {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(label: "Type", marker: .required) {
                Text("Photograph")
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            PVField(label: "Title", marker: .required) {
                PVInput(text: .constant(""))
            }
        }
    }
    .background(PVColor.surfaceCard)
}
