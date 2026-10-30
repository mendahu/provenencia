import SwiftUI

/// A field's Why (board S9-D5): every record the auto-reconciler considered,
/// as Read as · Source · Outcome. The last column is reserved for a future
/// "conclude this value" action, so adding it won't relayout the table
/// (PD-8). A Source opens the record's Citation in the composer.
struct ReconciliationReasoningView: View {
    let title: String
    let label: String
    let style: ReconciledValueRowModel.ValueStyle
    let groups: [ReconciliationRecordGroup]

    @Environment(WorkspaceNavigation.self) private var navigation

    /// Reserved for the conclude action.
    static let actionColumnWidth: CGFloat = 88

    var body: some View {
        PVCard(tone: .sunken, cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: 0) {
                Text(title)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .padding(.horizontal, PVSpacing.space5)
                    .padding(.top, PVSpacing.space5)
                    .padding(.bottom, PVSpacing.space3)
                    .accessibilityAddTraits(.isHeader)
                Grid(alignment: .leadingFirstTextBaseline, horizontalSpacing: 14, verticalSpacing: 0) {
                    GridRow {
                        heading(L10n.Conclusions.whyReadAs)
                        heading(L10n.Conclusions.whySource)
                        heading(L10n.Conclusions.whyOutcome)
                        Color.clear.frame(width: Self.actionColumnWidth, height: 1)
                    }
                    .padding(.vertical, PVSpacing.space2)
                    .accessibilityHidden(true)
                    ForEach(groups) { group in
                        if let owner = group.owner {
                            Rectangle()
                                .fill(PVColor.borderSubtle)
                                .frame(height: 1)
                            ownerLine(owner)
                        }
                        ForEach(Array(group.records.enumerated()), id: \.element.id) { index, record in
                            // The first record under a group line follows the line, not a rule.
                            if group.owner == nil || index > 0 {
                                Rectangle()
                                    .fill(PVColor.borderSubtle)
                                    .frame(height: 1)
                            }
                            GridRow {
                                Text(record.readAs)
                                    .font(readAsFont(record))
                                    .foregroundStyle(record.counted ? PVColor.textPrimary : PVColor.textSecondary)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                PVButton(record.sourceTitle, variant: .link, size: .sm) {
                                    navigation.go(to: record.location)
                                }
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .help(record.sourceTitle)
                                .accessibilityIdentifier("conclusion.detail.why.\(record.id).source")
                                ReconciliationOutcome(mark: record.mark, phrase: record.phrase)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                Color.clear.frame(width: Self.actionColumnWidth, height: 1)
                                    .accessibilityHidden(true)
                            }
                            .padding(.vertical, PVSpacing.space3)
                        }
                    }
                }
                .padding(.horizontal, PVSpacing.space5)
                .padding(.bottom, PVSpacing.space2)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.Conclusions.whyRecords(label))
    }

    /// A group's heading: the owning entity as its list titles it, then its
    /// ref. Text, never a link.
    private func ownerLine(_ owner: ReconciliationRecordGroup.OwnerLine) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(verbatim: owner.title)
                .pvMicroCaps()
                .foregroundStyle(PVColor.textMuted)
            if let ref = owner.ref {
                Text(verbatim: "· \(ref)")
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .fixedSize(horizontal: false, vertical: true)
        .frame(maxWidth: .infinity, alignment: .leading)
        .gridCellColumns(4)
        .padding(.top, PVSpacing.space5)
        .padding(.bottom, PVSpacing.space2)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.isHeader)
        .accessibilityLabel(owner.spoken)
    }

    private func heading(_ text: LocalizedStringResource) -> some View {
        Text(text)
            .pvMicroCaps()
            .foregroundStyle(PVColor.textMuted)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func readAsFont(_ record: ReconciliationRecord) -> Font {
        if record.readAsIsPhrase {
            return PVFont.body(size: PVTypeScale.bodySmall, italic: true)
        }
        switch style {
        case .text: return PVFont.body(size: PVTypeScale.bodySmall)
        case .date: return PVFont.mono(size: PVTypeScale.caption)
        case .place: return PVFont.body(size: PVTypeScale.bodySmall, italic: true)
        }
    }
}
