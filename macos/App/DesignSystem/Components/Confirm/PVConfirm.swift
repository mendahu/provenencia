import SwiftUI

/// The macOS-native counterpart to `components/feedback/ConfirmDialog.jsx`,
/// ported from the design system's `swift/ProvenenciaConfirm.swift`.
///
/// **`ConfirmDialog` is deliberately not ported.** Its own `prompt.md` says so:
/// the web component draws a scrim, a backdrop blur, a corner radius and a
/// shadow only because a browser gives it none of them. On macOS all four
/// belong to the window:
///
/// - macOS does **not** dim the parent window behind a sheet — the parent's
///   controls just go inactive. A dark scrim is a web/iOS idiom and reads as
///   wrong here. The absent scrim is correct, not missing.
/// - The sheet window supplies its own corner radius, shadow and material.
///   Setting `.background` / `.cornerRadius` / `.shadow` on the sheet's *panel*
///   is what produces the double-rounded, double-shadowed look — so nothing
///   here does. Styling *within* the panel is still ours: the action bar keeps
///   `Dialog.jsx`'s `--surface-sunken` band and hairline top rule.
/// - Sheets are modal to their *window*, not the app, and slide from the
///   titlebar. That comes free from `.sheet`.
///
/// So there are two right answers, and both live here:
///
/// 1. ``SwiftUI/View/pvConfirm(isPresented:copy:tone:onConfirm:)`` — a system
///    alert. The default: Apple's own pattern, fully system-drawn, and it
///    inherits keyboard, VoiceOver and Reduce Motion behaviour for free.
///    Message copy is plain text only.
/// 2. ``SwiftUI/View/pvConfirmSheet(item:copy:tone:isRunning:onConfirm:detail:)``
///    — a sheet, for when the consequence needs rich content (a mono-set key,
///    a list of affected records). Chrome still belongs to the window; this
///    only lays out content and the button row.

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
    /// Actual loss. The system tints the button's title red.
    case danger
    /// Irreversible but non-destructive — merging two people, publishing a tree.
    case irreversible

    var buttonRole: ButtonRole? { self == .danger ? .destructive : nil }

    /// Confirm-button chrome in the sheet form, where the button is content
    /// and takes design-system styling (the system alert keeps native buttons).
    var buttonVariant: PVButtonVariant { self == .danger ? .danger : .primary }
}

// MARK: - System alert (preferred)

private struct PVConfirmAlert: ViewModifier {
    @Binding var isPresented: Bool
    let copy: PVConfirmCopy
    let tone: PVConfirmTone
    let onConfirm: () -> Void

    func body(content: Content) -> some View {
        content.alert(copy.title, isPresented: $isPresented) {
            // Order matters: AppKit lays alert buttons out trailing-first, and
            // cancel is made the DEFAULT so Return dismisses safely. A
            // destructive action must never be one reflexive Return away.
            Button(String(localized: copy.cancel), role: .cancel) { isPresented = false }
                .keyboardShortcut(.defaultAction)
            Button(String(localized: copy.confirm), role: tone.buttonRole) {
                isPresented = false
                onConfirm()
            }
        } message: {
            Text(copy.message)
        }
    }
}

extension View {
    /// Native confirmation alert. Escape and Return both cancel; the confirm
    /// button carries the destructive role so the system tints it.
    ///
    /// Prefer this over ``pvConfirmSheet(item:copy:tone:isRunning:onConfirm:detail:)``
    /// unless the consequence genuinely needs formatted content.
    func pvConfirm(
        isPresented: Binding<Bool>,
        copy: PVConfirmCopy,
        tone: PVConfirmTone = .danger,
        onConfirm: @escaping () -> Void
    ) -> some View {
        modifier(PVConfirmAlert(isPresented: isPresented, copy: copy, tone: tone, onConfirm: onConfirm))
    }
}

// MARK: - Sheet, for rich consequence copy

/// Sheet body for a confirmation whose consequence needs more than a string.
///
/// Draws **no** panel background, corner radius, shadow or scrim — the sheet
/// window owns all four. The action bar's own `surfaceSunken` band is content,
/// not window chrome, and is carried over from `Dialog.jsx`'s footer.
struct PVConfirmSheetContent<Detail: View>: View {
    let copy: PVConfirmCopy
    let tone: PVConfirmTone
    let isRunning: Bool
    /// When set, confirm/cancel mint `{prefix}.confirm` / `{prefix}.cancel`.
    let accessibilityIdentifierPrefix: String?
    let onConfirm: () -> Void
    let onCancel: () -> Void
    @ViewBuilder let detail: () -> Detail

    @FocusState private var cancelFocused: Bool

    /// Alert-family width. A sheet should not size itself to the parent window.
    private let width: CGFloat = 420

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
        VStack(spacing: 0) {
            message
            actionBar
        }
        .frame(width: width)
        // Focus starts on cancel, matching the web component and AppKit's own
        // destructive alerts — Return must not complete a destructive action.
        .onAppear { cancelFocused = true }
    }

    private var message: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            // Alerts are centred; a sheet carrying body copy reads better
            // leading-aligned.
            Text(copy.title)
                .font(PVFont.display(size: PVTypeScale.h3))
                .foregroundStyle(PVColor.textDisplay)
                .fixedSize(horizontal: false, vertical: true)

            Text(copy.message)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
                .lineSpacing((PVLineHeight.relaxed - 1) * PVTypeScale.bodySmall)
                .fixedSize(horizontal: false, vertical: true)

            detail()
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space9)
    }

    /// `Dialog.jsx` sets its footer on `--surface-sunken` with a hairline top
    /// rule, and `ConfirmDialog` passes its buttons straight into that slot —
    /// so the band is part of the design, not browser-only chrome. The
    /// "draw none of it" rule this file follows names four things the *window*
    /// owns (scrim, blur, corner radius, shadow); an internal action bar is
    /// content, the same as the divider above it, so it is ours to paint.
    /// `swift/ProvenenciaConfirm.swift` omits it; this restores it.
    /// The buttons take design-system chrome (`.pv(…)`), not the system
    /// `.borderedProminent` pill — they sit *inside* the panel, in a band this
    /// view already paints, so they are content by the same rule as the
    /// `surfaceSunken` fill above. Nothing native is lost by the swap: the
    /// keyboard shortcuts, focus, roles and disabled state live on `Button`
    /// itself, and the destructive tint was already ours (`PVColor.danger`).
    /// Only the *window's* chrome — corner radius, shadow, slide-in — stays
    /// system-drawn, and the plain `.pvConfirm` alert stays fully native.
    private var actionBar: some View {
        HStack(spacing: PVSpacing.space5) {
            Spacer(minLength: PVSpacing.space8)
            Button(String(localized: copy.cancel)) { onCancel() }
                .buttonStyle(.pv(.secondary, size: .lg))
                .keyboardShortcut(.cancelAction)
                .disabled(isRunning)
                .focused($cancelFocused)
                .modifier(ConfirmOptionalAccessibilityIdentifier(prefix: accessibilityIdentifierPrefix, suffix: "cancel"))
            Button(String(localized: copy.confirm), role: tone.buttonRole) { onConfirm() }
                .buttonStyle(.pv(tone.buttonVariant, size: .lg))
                .disabled(isRunning)
                .modifier(ConfirmOptionalAccessibilityIdentifier(prefix: accessibilityIdentifierPrefix, suffix: "confirm"))
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

extension View {
    /// Sheet-based confirmation for rich consequence copy, keyed to the
    /// record it names. `copy` and `detail` render under the title — a
    /// mono-set key, an affected-record list — and both receive the record
    /// as a snapshot.
    ///
    /// **Why `item:` and not `isPresented: Bool`:** a confirmation's copy
    /// names a record held by a model, and dismissing clears that model
    /// state. With a `Bool` binding the content closures read the model
    /// *live*, so the name and detail blank out the moment the state clears —
    /// a visible flash while the window is still animating away.
    /// `.sheet(item:)` hands the closures the last non-nil record, so the
    /// sheet slides out still showing what it named.
    ///
    /// **Deviation from `swift/ProvenenciaConfirm.swift`:** the reference
    /// clears its presentation state before invoking `onConfirm`, which would
    /// close the sheet the instant a confirm starts — leaving its own
    /// `isRunning` spinner and any failure with nowhere to show. Dismissal is
    /// left to the caller's binding instead, so an async action can stay on
    /// screen while it runs and report an error in `detail` if it fails.
    func pvConfirmSheet<Item: Identifiable, Detail: View>(
        item: Binding<Item?>,
        copy: @escaping (Item) -> PVConfirmCopy,
        tone: PVConfirmTone = .danger,
        isRunning: Bool = false,
        accessibilityIdentifierPrefix: String? = nil,
        onConfirm: @escaping () -> Void,
        @ViewBuilder detail: @escaping (Item) -> Detail
    ) -> some View {
        sheet(item: item) { value in
            PVConfirmSheetContent(
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
        if let prefix {
            content.accessibilityIdentifier("\(prefix).\(suffix)")
        } else {
            content
        }
    }
}

/// A released/affected identifier shown under a confirmation's message — the
/// `<code>` slot in `ConfirmDialog`'s own example, and the reason a delete
/// confirmation earns a sheet rather than a plain alert.
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

#Preview("Confirm — sheet with released key") {
    PVConfirmSheetContent(
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
