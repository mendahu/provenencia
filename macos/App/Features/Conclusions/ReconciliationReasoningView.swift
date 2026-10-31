import SwiftUI

/// A field's Why (board S9-D5): every record the auto-reconciler considered,
/// as Read as · Source · Outcome, in a sunken card (the table is
/// `ReconciliationRecordRows`). A Source opens the record's Citation in the
/// composer. The Event and Place pages draw it; the Person page opens the
/// provenance sheet instead.
struct ReconciliationReasoningView: View {
    let title: String
    let label: String
    let style: ReconciledValueRowModel.ValueStyle
    let groups: [ReconciliationRecordGroup]

    @Environment(WorkspaceNavigation.self) private var navigation

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
                ReconciliationRecordRows(
                    groups: groups, style: style, provenance: .why,
                    open: { navigation.go(to: $0) }
                )
                .padding(.horizontal, PVSpacing.space5)
                .padding(.bottom, PVSpacing.space2)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.Conclusions.whyRecords(label))
    }
}
