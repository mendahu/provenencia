import SwiftUI

/// Summary Observation row on the composer form (edit / remove actions).
struct CitationComposerObservationRow: View {
    let row: CitationComposerModel.ObservationRow
    let propertyLabel: String
    let summary: String
    var inert: Bool
    var onEdit: () -> Void
    var onRemove: () -> Void

    var body: some View {
        PVCard(tone: .raised, padding: PVSpacing.space5) {
            HStack(alignment: .top, spacing: PVSpacing.space4) {
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    Text(verbatim: propertyLabel.isEmpty ? "—" : propertyLabel)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                    Text(verbatim: summary)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textPrimary)
                }
                Spacer(minLength: 0)
                HStack(spacing: PVSpacing.space2) {
                    if row.polarity == "negative" {
                        Text(L10n.CitationComposer.polarityNegates)
                            .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.medium))
                            .foregroundStyle(PVColor.textSecondary)
                            .padding(.horizontal, PVSpacing.space3)
                            .padding(.vertical, PVSpacing.space1)
                            .background(
                                Capsule(style: .continuous)
                                    .fill(PVColor.surfaceSunken)
                            )
                    }
                    if !inert {
                        PVIconButton(.penLine, label: L10n.CitationComposer.editObservation, size: .sm, action: onEdit)
                        PVIconButton(
                            .trash,
                            label: L10n.CitationComposer.removeObservation,
                            size: .sm,
                            tone: .danger,
                            action: onRemove
                        )
                    }
                }
            }
        }
    }
}
