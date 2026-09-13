import SwiftUI

/// The vocabulary destinations' page header: display title and description on
/// the left, the origin count line and the add button on the right. Mints
/// `<prefix>.countLine` and `<prefix>.add`.
struct VocabularyHeader: View {
    let title: LocalizedStringResource
    let description: LocalizedStringResource
    let countLine: String
    let addLabel: LocalizedStringResource
    let isAddDisabled: Bool
    let identifierPrefix: String
    let onAdd: () -> Void

    var body: some View {
        HStack(alignment: .bottom, spacing: PVSpacing.space8) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(title)
                    .font(PVFont.display(size: PVTypeScale.h1))
                    .foregroundStyle(PVColor.textDisplay)
                Text(description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: PVSpacing.measureProse, alignment: .leading)
            }
            Spacer(minLength: PVSpacing.space6)
            HStack(spacing: PVSpacing.space6) {
                Text(countLine)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("\(identifierPrefix).countLine")
                PVButton(addLabel, variant: .primary, icon: .plus) {
                    onAdd()
                }
                .disabled(isAddDisabled)
                .accessibilityIdentifier("\(identifierPrefix).add")
            }
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.top, PVSpacing.space8)
        .padding(.bottom, PVSpacing.space6)
    }
}

extension View {
    /// The vocabulary destinations' toast: top-trailing, slides in from the
    /// trailing edge, honours Reduce Motion. The tone travels with the toast
    /// (`VocabularyToast.tone`) — the model decides how an event reads.
    func vocabularyToastOverlay(_ toast: Binding<VocabularyToast?>, identifier: String) -> some View {
        modifier(VocabularyToastOverlay(toast: toast, identifier: identifier))
    }
}

private struct VocabularyToastOverlay: ViewModifier {
    @Binding var toast: VocabularyToast?
    let identifier: String
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func body(content: Content) -> some View {
        content
            .overlay(alignment: .topTrailing) {
                if let toast {
                    PVToast(
                        tone: toast.tone,
                        title: toast.title,
                        message: toast.body,
                        onDismiss: { self.toast = nil }
                    )
                    .id(toast)
                    .padding(PVSpacing.space8)
                    .transition(.move(edge: .trailing).combined(with: .opacity))
                    .accessibilityIdentifier(identifier)
                }
            }
            .animation(reduceMotion ? nil : PVMotion.easeStandard, value: toast)
    }
}

/// The detail panel's header, shared by add / locked / editable states of
/// both destinations: eyebrow with the delete affordance, display title,
/// origin badge beside the mono key (and, when the caller has one, a usage
/// line), then the key hint. Mints `<prefix>.detail.key`,
/// `<prefix>.detail.usage` and `<prefix>.delete`.
struct VocabularyPanelHeader: View {
    let eyebrow: LocalizedStringResource
    let title: String
    let origin: String
    let keyText: String
    /// Optional evidence `type_*` key — Source types show a 22pt mark beside the title.
    var iconKey: String? = nil
    /// The "used by N sources" line — only Source types carries one so far.
    var usageLine: String?
    let keyHint: LocalizedStringResource
    /// Disabled rather than hidden when the row cannot be deleted — the
    /// tooltip is where the reason lives (a plugin owns it, or it is in use).
    let showsDelete: Bool
    let canDelete: Bool
    let deleteTooltip: LocalizedStringResource
    let identifierPrefix: String
    let onDelete: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                Text(eyebrow)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
                if showsDelete {
                    PVIconButton(.trash, label: deleteTooltip, size: .sm, tone: .danger) {
                        onDelete()
                    }
                    .disabled(!canDelete)
                    .accessibilityIdentifier("\(identifierPrefix).delete")
                }
            }
            HStack(alignment: .center, spacing: PVSpacing.space5) {
                if let iconKey, !iconKey.isEmpty {
                    ZStack {
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .fill(PVColor.surfaceSunken)
                            .overlay(
                                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                                    .stroke(PVColor.borderSubtle, lineWidth: 1)
                            )
                        PVEvidenceIcon(
                            PVEvidenceIconKey(catalogKey: iconKey),
                            size: 22,
                            decorative: true
                        )
                    }
                    .frame(width: 36, height: 36)
                    .accessibilityIdentifier("\(identifierPrefix).detail.icon")
                }
                Text(title)
                    .font(PVFont.display(size: PVTypeScale.h2))
                    .foregroundStyle(PVColor.textDisplay)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            HStack(spacing: PVSpacing.space5) {
                OriginBadge(origin: origin)
                Text(keyText)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textSecondary)
                    .accessibilityIdentifier("\(identifierPrefix).detail.key")
                if let usageLine {
                    Text(usageLine)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textFaint)
                        .accessibilityIdentifier("\(identifierPrefix).detail.usage")
                }
            }
            Text(keyHint)
                .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }
}

/// A micro-caps label over its content — the locked detail's reading layout.
struct VocabularyLabeledSection<Content: View>: View {
    let label: LocalizedStringResource
    @ViewBuilder let content: Content

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(label)
                .pvMicroCaps()
                .foregroundStyle(PVColor.textMuted)
            content
        }
    }
}

/// The add/edit form's action row: primary submit beside a ghost secondary
/// (cancel while adding, revert while editing — the caller decides the words
/// and the work). Mints `<prefix>.form.submit` and `<prefix>.form.secondary`.
struct VocabularyFormActions: View {
    let primaryLabel: LocalizedStringResource
    let secondaryLabel: LocalizedStringResource
    let isSaving: Bool
    let canSubmit: Bool
    let isSecondaryDisabled: Bool
    let identifierPrefix: String
    let onPrimary: () -> Void
    let onSecondary: () -> Void

    var body: some View {
        HStack(spacing: PVSpacing.space5) {
            PVButton(primaryLabel, variant: .primary, loading: isSaving) {
                onPrimary()
            }
            .disabled(!canSubmit)
            .accessibilityIdentifier("\(identifierPrefix).form.submit")
            PVButton(secondaryLabel, variant: .ghost) {
                onSecondary()
            }
            .disabled(isSecondaryDisabled)
            .accessibilityIdentifier("\(identifierPrefix).form.secondary")
        }
    }
}
