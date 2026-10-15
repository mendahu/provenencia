import AppKit
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

/// Whether a mouse-down belongs to the sheet. A system open or save panel
/// counts as inside: its clicks are not a click away from the form.
enum PVFormDialogClick {
    static func isInsideSheet(
        eventWindow: NSWindow?,
        sheet: NSWindow,
        modalWindow: NSWindow?
    ) -> Bool {
        if modalWindow is NSSavePanel { return true }
        guard let eventWindow else { return false }
        if eventWindow.level.rawValue >= NSWindow.Level.popUpMenu.rawValue { return true }
        var current: NSWindow? = eventWindow
        while let window = current {
            if window === sheet || window is NSSavePanel { return true }
            current = window.parent
        }
        return false
    }
}

/// What an outside click does. Running work is left alone. A dirty sheet
/// asks before it closes. A clean or read-only sheet closes.
enum PVFormDialogDismiss: Equatable {
    case ignore
    case close
    case confirmDiscard

    static func action(isRunning: Bool, isDirty: Bool) -> PVFormDialogDismiss {
        if isRunning { return .ignore }
        if isDirty { return .confirmDiscard }
        return .close
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
                    .keyboardShortcut(.cancelAction)
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
    /// Sheet-based form dialog. Dismissal is left to the caller's binding so
    /// an async Create can stay on screen while it runs and keep field errors
    /// visible if validation or the store fails.
    ///
    /// A click outside the sheet closes it. When `isDirty` is true, that click
    /// asks to discard or keep editing. The Cancel button still closes
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
        modifier(PVFormDialogModifier(
            isPresented: isPresented,
            copy: copy,
            width: width,
            isRunning: isRunning,
            isDirty: isDirty,
            confirmDisabled: confirmDisabled,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: onConfirm,
            form: form
        ))
    }
}

private struct PVFormDialogModifier<Form: View>: ViewModifier {
    @Binding var isPresented: Bool
    let copy: PVFormDialogCopy
    let width: CGFloat
    let isRunning: Bool
    let isDirty: Bool
    let confirmDisabled: Bool
    let accessibilityIdentifierPrefix: String?
    let onConfirm: () -> Void
    @ViewBuilder let form: () -> Form

    @State private var confirmingDiscard = false

    func body(content: Content) -> some View {
        content.sheet(isPresented: $isPresented, onDismiss: { confirmingDiscard = false }) {
            sheetBody
                .background {
                    PVFormDialogOutsideClick(isEnabled: !confirmingDiscard) {
                        switch PVFormDialogDismiss.action(isRunning: isRunning, isDirty: isDirty) {
                        case .ignore:
                            break
                        case .close:
                            isPresented = false
                        case .confirmDiscard:
                            confirmingDiscard = true
                        }
                    }
                }
        }
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
                onConfirm: { isPresented = false },
                onCancel: { confirmingDiscard = false },
                detail: { EmptyView() }
            )
        } else {
            PVFormDialogContent(
                copy: copy,
                width: width,
                isRunning: isRunning,
                confirmDisabled: confirmDisabled,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                onConfirm: onConfirm,
                onCancel: { isPresented = false },
                form: form
            )
        }
    }
}

/// Sees a mouse-down that lands outside this sheet. A macOS sheet does not
/// deliver that click to the parent window, so the presenter watches the
/// event stream instead of adding a scrim.
private struct PVFormDialogOutsideClick: NSViewRepresentable {
    var isEnabled: Bool
    var onOutsideClick: () -> Void

    func makeNSView(context: Context) -> PVFormDialogOutsideClickView {
        let view = PVFormDialogOutsideClickView()
        view.isEnabled = isEnabled
        view.onOutsideClick = onOutsideClick
        return view
    }

    func updateNSView(_ nsView: PVFormDialogOutsideClickView, context: Context) {
        nsView.isEnabled = isEnabled
        nsView.onOutsideClick = onOutsideClick
    }
}

private final class PVFormDialogOutsideClickView: NSView {
    var isEnabled = true
    var onOutsideClick: () -> Void = {}
    /// Touched from `deinit`, which is not isolated to the main actor.
    nonisolated(unsafe) private var monitor: Any?

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        if window == nil {
            removeMonitor()
        } else {
            installMonitor()
        }
    }

    deinit {
        if let monitor {
            NSEvent.removeMonitor(monitor)
        }
    }

    private func installMonitor() {
        guard monitor == nil else { return }
        monitor = NSEvent.addLocalMonitorForEvents(
            matching: [.leftMouseDown, .rightMouseDown]
        ) { [weak self] event in
            MainActor.assumeIsolated {
                self?.handle(event)
            }
            return event
        }
    }

    private func removeMonitor() {
        if let monitor {
            NSEvent.removeMonitor(monitor)
        }
        monitor = nil
    }

    private func handle(_ event: NSEvent) {
        guard isEnabled, let sheet = window else { return }
        if PVFormDialogClick.isInsideSheet(
            eventWindow: event.window,
            sheet: sheet,
            modalWindow: NSApp.modalWindow
        ) { return }
        onOutsideClick()
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
