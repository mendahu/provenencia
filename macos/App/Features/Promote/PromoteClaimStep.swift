import SwiftUI

/// The last step for each subject: the Identity Claim's status, confidence
/// and argument, under a summary of what one save writes (S9-D10 frames
/// D10-01 – 05). Inputs lock while the write runs and after a refusal.
struct PromoteClaimStep: View {
    @Bindable var model: PromoteModel
    let grades: QueryHandle<[CatalogClaimConfidenceGrade]>?
    let properties: QueryHandle<PropertiesSnapshot>?

    private var isLocked: Bool { model.controls.isLocked }

    var body: some View {
        let prefix = model.refPrefix(in: properties?.value)
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(title: L10n.Promote.claimStep, meta: model.stepText)

            if let title = model.blockedTitle {
                PVCallout(tone: .danger, title: title, message: model.blockedMessage, compact: true)
                    .accessibilityIdentifier("promote.claim.blocked")
            } else if let error = model.saveError {
                PVCallout(tone: .danger, title: L10n.Promote.saveFailedTitle, message: error, compact: true)
                    .accessibilityIdentifier("promote.claim.error")
            }

            PromoteClaimSummary(model: model, prefix: prefix)

            HStack(alignment: .top, spacing: 20) {
                PVField(label: L10n.Promote.statusLabel, hint: L10n.Promote.statusHint) {
                    PVSelect(
                        selection: .constant(model.statusSelection),
                        options: model.statusOptions,
                        isDisabled: isLocked,
                        accessibilityLabel: L10n.Promote.statusLabel,
                        accessibilityIdentifier: "promote.claim.status"
                    )
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                PVField(label: L10n.Promote.confidenceLabel, hint: L10n.Promote.confidenceHint) {
                    PVSelect(
                        selection: Binding(
                            get: { model.confidenceSelection },
                            set: { model.setConfidence($0) }
                        ),
                        options: PromoteModel.confidenceOptions(grades?.value ?? []),
                        isDisabled: isLocked,
                        accessibilityLabel: L10n.Promote.confidenceLabel,
                        accessibilityIdentifier: "promote.claim.confidence"
                    )
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }

            PVField(label: L10n.Promote.argumentLabel, hint: model.argumentHint) {
                PVTextArea(
                    text: Binding(
                        get: { model.argument },
                        set: { model.setArgument($0) }
                    ),
                    lineLimit: 4...10,
                    prompt: L10n.Promote.argumentPrompt(model.kind)
                )
                .disabled(isLocked)
                .accessibilityIdentifier("promote.claim.argument")
            }
        }
    }
}

/// What one save writes: the subject → the handle (new, with its pending
/// ref, or the chosen one), the pins it carries, and a line about the handle.
private struct PromoteClaimSummary: View {
    let model: PromoteModel
    let prefix: String?

    var body: some View {
        let target = model.summaryTarget(prefix: prefix)
        PVCard(tone: .sunken) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                Text(L10n.Promote.claimSummary)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.textMuted)
                HStack(alignment: .center, spacing: 10) {
                    name(model.subject.name)
                    ref(model.subject.ref)
                    PVIcon(.arrowRight, size: 14)
                        .foregroundStyle(PVColor.textFaint)
                        .accessibilityHidden(true)
                    name(target.title)
                    if let targetRef = target.ref {
                        ref(targetRef)
                    }
                    Spacer(minLength: PVSpacing.space4)
                    PVBadge(L10n.Promote.noPins, icon: .pin)
                }
                Text(verbatim: model.summaryLine(prefix: prefix))
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(.vertical, 14)
            .padding(.horizontal, 16)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("promote.claim.summary")
    }

    private func name(_ text: String) -> some View {
        Text(verbatim: text)
            .font(PVFont.display(size: 15, weight: PVFontWeight.medium))
            .foregroundStyle(PVColor.textPrimary)
            .lineLimit(1)
    }

    private func ref(_ text: String) -> some View {
        Text(verbatim: text)
            .font(PVFont.mono(size: 12))
            .foregroundStyle(PVColor.textSecondary)
            .lineLimit(1)
    }
}
