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
    /// Already-localized title, for copy that includes a ref ("Edit ART-7QK2M").
    /// When set it is drawn instead of `title`.
    let verbatimTitle: String?
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
        self.verbatimTitle = nil
        self.subtitle = subtitle
        self.verbatimSubtitle = verbatimSubtitle
        self.confirm = confirm
        self.cancel = cancel
    }

    /// Copy whose title and subtitle are already localized strings.
    init(
        verbatimTitle: String,
        verbatimSubtitle: String?,
        confirm: LocalizedStringResource,
        cancel: LocalizedStringResource
    ) {
        // `title` is never drawn when `verbatimTitle` is set.
        self.title = LocalizedStringResource(stringLiteral: verbatimTitle)
        self.verbatimTitle = verbatimTitle
        self.subtitle = nil
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

    /// Return confirms the form, unless an inner edit (a row in edit, an add
    /// row) is open: then Return belongs to that edit.
    static func confirmShortcut(innerEditOpen: Bool) -> KeyboardShortcut? {
        innerEditOpen ? nil : .defaultAction
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

    /// The body scrolls between the fixed title and footer, up to
    /// ``maxScrollingBodyHeight``.
    let scrollsBody: Bool
    /// The `.id` of the section the scrolling body opens at the top of.
    let scrollTarget: AnyHashable?

    /// Tallest the scrolling body grows before it scrolls.
    static var maxScrollingBodyHeight: CGFloat { 560 }

    @FocusState private var cancelFocused: Bool
    @State private var bodyHeight: CGFloat = 0
    @Environment(\.pvSheetInnerEditOpen) private var innerEditOpen

    init(
        copy: PVFormDialogCopy,
        width: CGFloat = 480,
        isRunning: Bool = false,
        confirmDisabled: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        scrollsBody: Bool = false,
        scrollTarget: AnyHashable? = nil,
        onConfirm: @escaping () -> Void,
        onCancel: @escaping () -> Void,
        @ViewBuilder form: @escaping () -> Form
    ) {
        self.scrollsBody = scrollsBody
        self.scrollTarget = scrollTarget
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
            title: formDialogTitle(copy),
            subtitle: formDialogSubtitle(copy),
            width: width,
            footerChrome: .sunken
        ) {
            formBody
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
                    .keyboardShortcut(PVFormDialogControls.confirmShortcut(innerEditOpen: innerEditOpen))
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

extension PVFormDialogContent {
    @ViewBuilder
    fileprivate var formBody: some View {
        if scrollsBody {
            ScrollView(.vertical) {
                form()
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .onGeometryChange(for: CGFloat.self) { $0.size.height } action: { bodyHeight = $0 }
            }
            .frame(height: min(bodyHeight, Self.maxScrollingBodyHeight))
            .pvScrollsRowIntoView(scrollTarget, anchor: .top, animated: false, scrollsOnAppear: true)
        } else {
            form()
        }
    }
}

private func formDialogTitle(_ copy: PVFormDialogCopy) -> Text {
    if let verbatim = copy.verbatimTitle, !verbatim.isEmpty {
        return Text(verbatim: verbatim)
    }
    return Text(copy.title)
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
        scrollsBody: Bool = false,
        scrollTarget: AnyHashable? = nil,
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
                scrollsBody: scrollsBody,
                scrollTarget: scrollTarget,
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
        scrollsBody: Bool = false,
        scrollTarget: AnyHashable? = nil,
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
            scrollsBody: scrollsBody,
            scrollTarget: scrollTarget,
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

#Preview("Form dialog — scrolling body") {
    PVFormDialogContent(
        copy: PVFormDialogCopy(
            verbatimTitle: "Edit ART-7QK2M",
            verbatimSubtitle: "SRC-4KD2P · 1880 U.S. census",
            confirm: "Save artifact",
            cancel: "Cancel"
        ),
        width: 560,
        scrollsBody: true,
        scrollTarget: 3,
        onConfirm: {},
        onCancel: {}
    ) {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            ForEach(0..<8, id: \.self) { index in
                PVField(label: "Section \(index)") {
                    Text(verbatim: String(repeating: "Body ", count: 40))
                }
                .id(index)
            }
        }
    }
    .background(PVColor.surfaceCard)
}
