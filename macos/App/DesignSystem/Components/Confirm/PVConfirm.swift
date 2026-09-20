import SwiftUI

/// The macOS confirm composite — counterpart to `ConfirmDialog.jsx`, composed
/// on ``PVPanel``.
///
/// **`ConfirmDialog` is deliberately not ported as a web-style modal.** Its
/// `prompt.md` says so: the web component draws a scrim, blur, corner radius
/// and shadow because a browser gives it none. On macOS those belong to the
/// window:
///
/// - macOS does **not** dim the parent behind a sheet — controls go inactive.
///   A dark scrim is a web/iOS idiom and reads wrong here.
/// - The sheet window supplies corner radius, shadow and material. Painting
///   those on panel *content* double-rounds / double-shadows — so we don't.
/// - Sheets are modal to their *window* and slide from the titlebar via
///   `.sheet`.
///
/// Provenencia's confirm is this rich sheet (``pvConfirm(item:…)``): title,
/// consequence message, optional detail (e.g. ``PVConfirmKeyChip``), cancel-
/// first focus, and a sunken Keep/Delete footer. Plain SwiftUI `.alert` stays
/// available at call sites if a future flow needs text-only system chrome; the
/// kit does not wrap it.

// MARK: - Copy model

/// The copy discipline from `ConfirmDialog.prompt.md`, as a value: the title is
/// a question naming the record ("Delete Photographer?", never "Are you
/// sure?"), the message says what is *and is not* lost, and confirm repeats the
/// verb ("Delete field", never "OK"). The cancel label names the safe outcome —
/// "Keep field" reads better than "Cancel" beside a destructive twin.
///
/// `title` and `message` are `String` because they name a record and are
/// formatted through `L10n` at the call site; the two button labels are fixed
/// UI copy.
struct PVConfirmCopy {
    let title: String
    let message: String
    let confirm: LocalizedStringResource
    let cancel: LocalizedStringResource

    init(
        title: String,
        message: String,
        confirm: LocalizedStringResource,
        cancel: LocalizedStringResource
    ) {
        self.title = title
        self.message = message
        self.confirm = confirm
        self.cancel = cancel
    }
}

enum PVConfirmTone {
    /// Actual loss. Confirm uses the danger button chrome and a trash glyph.
    case danger
    /// Irreversible but non-destructive — merging two people, publishing a tree.
    case irreversible

    var buttonRole: ButtonRole? { self == .danger ? .destructive : nil }

    /// Confirm-button chrome — design-system styling inside the panel footer.
    var buttonVariant: PVButtonVariant { self == .danger ? .danger : .primary }

    /// Leading glyph on the confirm control (`PVConfirmDialog`'s `confirmIcon`).
    var confirmIcon: PVSymbol? { self == .danger ? .trash : nil }
}

/// Pure enablement for confirm footer buttons — extracted for unit tests.
enum PVConfirmControls {
    static func isActionDisabled(isRunning: Bool) -> Bool { isRunning }
}

/// Layout constants for the confirm sheet.
enum PVConfirmLayout {
    /// Kit default (`PVConfirmDialog` width) — alert-family, slightly under form sheets.
    static let panelWidth: CGFloat = 440
}

/// Accessibility ids for confirm chrome buttons.
enum PVConfirmAccessibility {
    /// `{prefix}.{suffix}` when a prefix is set; otherwise `nil` (no id).
    static func identifier(prefix: String?, suffix: String) -> String? {
        guard let prefix else { return nil }
        return "\(prefix).\(suffix)"
    }
}

// MARK: - Sheet content

/// Sheet body for a confirmation. Layout chrome comes from ``PVPanel``
/// (message as subtitle; detail in the body slot), including the warm card fill.
struct PVConfirmContent<Detail: View>: View {
    let copy: PVConfirmCopy
    let tone: PVConfirmTone
    let isRunning: Bool
    /// When set, confirm/cancel mint `{prefix}.confirm` / `{prefix}.cancel`.
    let accessibilityIdentifierPrefix: String?
    let onConfirm: () -> Void
    let onCancel: () -> Void
    @ViewBuilder let detail: () -> Detail

    @FocusState private var cancelFocused: Bool

    init(
        copy: PVConfirmCopy,
        tone: PVConfirmTone = .danger,
        isRunning: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        onCancel: @escaping () -> Void,
        @ViewBuilder detail: @escaping () -> Detail
    ) {
        self.copy = copy
        self.tone = tone
        self.isRunning = isRunning
        self.accessibilityIdentifierPrefix = accessibilityIdentifierPrefix
        self.onConfirm = onConfirm
        self.onCancel = onCancel
        self.detail = detail
    }

    var body: some View {
        PVPanel(
            title: Text(copy.title),
            subtitle: Text(copy.message),
            width: PVConfirmLayout.panelWidth,
            footerChrome: .sunken
        ) {
            detail()
        } footer: {
            HStack(spacing: PVSpacing.space5) {
                Spacer(minLength: PVSpacing.space8)
                Button(String(localized: copy.cancel)) { onCancel() }
                    .buttonStyle(.pv(.secondary, size: .lg))
                    .keyboardShortcut(.cancelAction)
                    .disabled(PVConfirmControls.isActionDisabled(isRunning: isRunning))
                    .focused($cancelFocused)
                    .modifier(ConfirmOptionalAccessibilityIdentifier(
                        prefix: accessibilityIdentifierPrefix,
                        suffix: "cancel"
                    ))
                PVButton(
                    copy.confirm,
                    variant: tone.buttonVariant,
                    size: .lg,
                    icon: tone.confirmIcon,
                    loading: isRunning,
                    action: onConfirm
                )
                .keyboardShortcut(.defaultAction)
                .disabled(PVConfirmControls.isActionDisabled(isRunning: isRunning))
                .modifier(ConfirmOptionalAccessibilityIdentifier(
                    prefix: accessibilityIdentifierPrefix,
                    suffix: "confirm"
                ))
            }
        }
        // Focus starts on cancel — Return must not complete a destructive action.
        .onAppear { cancelFocused = true }
    }
}

extension View {
    /// Confirmation sheet keyed to the record it names. `copy` and `detail`
    /// render under the title — a mono-set key, an affected-record list — and
    /// both receive the record as a snapshot.
    ///
    /// **Why `item:` and not `isPresented: Bool`:** a confirmation's copy
    /// names a record held by a model, and dismissing clears that model
    /// state. With a `Bool` binding the content closures read the model
    /// *live*, so the name and detail blank out the moment the state clears —
    /// a visible flash while the window is still animating away.
    /// `.sheet(item:)` hands the closures the last non-nil record, so the
    /// sheet slides out still showing what it named.
    ///
    /// Dismissal is left to the caller's binding so an async action can stay
    /// on screen while it runs (`isRunning`) and report an error in `detail`
    /// if it fails.
    func pvConfirm<Item: Identifiable, Detail: View>(
        item: Binding<Item?>,
        copy: @escaping (Item) -> PVConfirmCopy,
        tone: PVConfirmTone = .danger,
        isRunning: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        @ViewBuilder detail: @escaping (Item) -> Detail
    ) -> some View {
        sheet(item: item) { value in
            PVConfirmContent(
                copy: copy(value),
                tone: tone,
                isRunning: isRunning,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                onConfirm: onConfirm,
                onCancel: { item.wrappedValue = nil },
                detail: { detail(value) }
            )
        }
    }
}

/// Applies `{prefix}.{suffix}` only when a prefix is provided.
private struct ConfirmOptionalAccessibilityIdentifier: ViewModifier {
    let prefix: String?
    let suffix: String

    func body(content: Content) -> some View {
        if let id = PVConfirmAccessibility.identifier(prefix: prefix, suffix: suffix) {
            content.accessibilityIdentifier(id)
        } else {
            content
        }
    }
}

/// A released/affected identifier shown under a confirmation's message — the
/// `<code>` slot in `ConfirmDialog`'s own example, and the reason a delete
/// confirmation uses a sheet with a detail slot.
struct PVConfirmKeyChip: View {
    let label: LocalizedStringResource
    let value: String

    var body: some View {
        HStack(spacing: PVSpacing.space3) {
            Text(label)
                .pvMicroCaps()
                .foregroundStyle(PVColor.textMuted)
            Text(value)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textPrimary)
                .padding(.horizontal, PVSpacing.space3)
                .padding(.vertical, PVSpacing.space1)
                .background(
                    PVColor.surfaceSunken,
                    in: RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                )
        }
        .accessibilityElement(children: .combine)
    }
}

#Preview("Confirm — released key") {
    PVConfirmContent(
        copy: PVConfirmCopy(
            title: "Delete Photographer?",
            message: "No source in this project carries a value for this field, so nothing is lost. The key is released and can be minted again by a later field with the same label.",
            confirm: "Delete field",
            cancel: "Keep field"
        ),
        onConfirm: {},
        onCancel: {}
    ) {
        PVConfirmKeyChip(label: "Key released", value: "photographer")
    }
    .background(PVColor.surfaceCard)
}
