import SwiftUI

/// Form dialog sheet — the macOS answer to `components/feedback/Dialog.jsx`
/// for short create/edit panels (Add Source, later Add Artifact, etc.).
///
/// Same presentation family as `PVConfirmSheetContent`: a window-owned
/// `.sheet` (no custom scrim, corner radius, or shadow on the panel). The
/// sunken footer band is content, carried over from `Dialog.jsx`. Prefer
/// `PVConfirm` for destructive / irreversible confirmations — this type is
/// for forms with a content slot.

struct PVDialogCopy {
    let title: LocalizedStringResource
    let subtitle: LocalizedStringResource?
    let confirm: LocalizedStringResource
    let cancel: LocalizedStringResource

    init(
        title: LocalizedStringResource,
        subtitle: LocalizedStringResource? = nil,
        confirm: LocalizedStringResource,
        cancel: LocalizedStringResource
    ) {
        self.title = title
        self.subtitle = subtitle
        self.confirm = confirm
        self.cancel = cancel
    }
}

/// Sheet body for a short form. Draws **no** panel background, corner radius,
/// shadow or scrim — the sheet window owns those.
struct PVDialogContent<Form: View>: View {
    let copy: PVDialogCopy
    let isRunning: Bool
    let confirmDisabled: Bool
    /// When set, confirm/cancel mint `{prefix}.confirm` / `{prefix}.cancel`.
    let accessibilityIdentifierPrefix: String?
    let onConfirm: () -> Void
    let onCancel: () -> Void
    @ViewBuilder let form: () -> Form

    @FocusState private var cancelFocused: Bool

    private let width: CGFloat = 480

    init(
        copy: PVDialogCopy,
        isRunning: Bool = false,
        confirmDisabled: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        onCancel: @escaping () -> Void,
        @ViewBuilder form: @escaping () -> Form
    ) {
        self.copy = copy
        self.isRunning = isRunning
        self.confirmDisabled = confirmDisabled
        self.accessibilityIdentifierPrefix = accessibilityIdentifierPrefix
        self.onConfirm = onConfirm
        self.onCancel = onCancel
        self.form = form
    }

    var body: some View {
        VStack(spacing: 0) {
            headerAndForm
            actionBar
        }
        .frame(width: width)
    }

    private var headerAndForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(copy.title)
                    .font(PVFont.display(size: PVTypeScale.h3))
                    .foregroundStyle(PVColor.textDisplay)
                    .fixedSize(horizontal: false, vertical: true)
                if let subtitle = copy.subtitle {
                    Text(subtitle)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineSpacing((PVLineHeight.relaxed - 1) * PVTypeScale.bodySmall)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            form()
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space9)
    }

    private var actionBar: some View {
        HStack(spacing: PVSpacing.space5) {
            Spacer(minLength: PVSpacing.space8)
            Button(String(localized: copy.cancel)) { onCancel() }
                .buttonStyle(.pv(.secondary, size: .lg))
                .keyboardShortcut(.cancelAction)
                .disabled(isRunning)
                .focused($cancelFocused)
                .modifier(OptionalAccessibilityIdentifier(prefix: accessibilityIdentifierPrefix, suffix: "cancel"))
            Button(String(localized: copy.confirm)) { onConfirm() }
                .buttonStyle(.pv(.primary, size: .lg))
                .keyboardShortcut(.defaultAction)
                .disabled(isRunning || confirmDisabled)
                .modifier(OptionalAccessibilityIdentifier(prefix: accessibilityIdentifierPrefix, suffix: "confirm"))
                .overlay(alignment: .trailing) {
                    if isRunning {
                        ProgressView()
                            .controlSize(.small)
                            .offset(x: 22)
                    }
                }
        }
        .padding(.horizontal, PVSpacing.space8)
        .padding(.vertical, PVSpacing.space8)
        .frame(maxWidth: .infinity)
        .background(PVColor.surfaceSunken)
        .overlay(alignment: .top) {
            PVDivider()
        }
    }
}

/// Applies `{prefix}.{suffix}` only when a prefix is provided.
private struct OptionalAccessibilityIdentifier: ViewModifier {
    let prefix: String?
    let suffix: String

    func body(content: Content) -> some View {
        if let prefix {
            content.accessibilityIdentifier("\(prefix).\(suffix)")
        } else {
            content
        }
    }
}

extension View {
    /// Sheet-based form dialog. Dismissal is left to the caller's binding so
    /// an async Create can stay on screen while it runs and keep field errors
    /// visible if validation or the store fails.
    func pvDialog<Form: View>(
        isPresented: Binding<Bool>,
        copy: PVDialogCopy,
        isRunning: Bool = false,
        confirmDisabled: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        @ViewBuilder form: @escaping () -> Form
    ) -> some View {
        sheet(isPresented: isPresented) {
            PVDialogContent(
                copy: copy,
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

#Preview("Dialog — Add source") {
    PVDialogContent(
        copy: PVDialogCopy(
            title: "Add source",
            subtitle: "A thin record now — artifacts, notes and metadata live on the Source page.",
            confirm: "Create source",
            cancel: "Cancel"
        ),
        onConfirm: {},
        onCancel: {}
    ) {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(label: "Type", required: true) {
                Text("Photograph")
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            PVField(label: "Title", required: true) {
                PVInput(text: .constant(""))
            }
        }
    }
    .background(PVColor.surfaceCard)
}
