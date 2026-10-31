import SwiftUI

/// Which row's provenance sheet is open: the row's id, not a snapshot, so the
/// sheet follows its row across a page refresh (S13-11 AC-4.1, AC-4.2).
struct ProvenanceSheetItem: Identifiable, Equatable {
    let id: String

    /// The row with this id, or nil when it is gone or has no records.
    func row(in rows: [ReconciledValueRowModel]) -> ReconciledValueRowModel? {
        rows.first { $0.id == id && !$0.records.isEmpty }
    }
}

/// Every record behind one Vitals value, and what the auto-reconciler did with
/// each (board S13-D2): the row's value and state, a summary, and the records
/// grouped by owning entity and by value. A record opens its Source, its Citation
/// or its Observation, closing the sheet first (D-5). Nothing here is
/// Person-only: the Event and Place pages adopt it with `ReconciledValueRow.Provenance.sheet`.
struct ProvenanceSheet: View {
    let row: ReconciledValueRowModel
    var isWide = false
    let close: () -> Void

    @Environment(WorkspaceNavigation.self) private var navigation

    /// The records scroll inside the sheet past this height; the title, header,
    /// column headings and Done stay.
    static let recordsMaxHeight: CGFloat = 560

    static func width(wide: Bool) -> CGFloat { wide ? 1280 : 880 }

    var body: some View {
        PVPanel(
            title: Text(verbatim: row.label),
            width: Self.width(wide: isWide)
        ) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                header
                // Room for "Concluded by you" (Reconciliation Claims): zero height until then.
                Color.clear
                    .frame(height: 0)
                    .accessibilityHidden(true)
                VStack(alignment: .leading, spacing: 0) {
                    ReconciliationRecordRows.SheetHeadings(isWide: isWide)
                    PVDivider()
                    ScrollView {
                        ReconciliationRecordRows(
                            groups: row.recordGroups, style: row.style, provenance: .sheet,
                            open: open, isWide: isWide
                        )
                    }
                    .frame(maxHeight: Self.recordsMaxHeight)
                }
            }
        } footer: {
            HStack {
                Spacer(minLength: 0)
                PVButton(L10n.Conclusions.provenanceDone, variant: .primary) { close() }
                    .accessibilityIdentifier("conclusion.detail.provenance.done")
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("conclusion.detail.provenance.\(row.id)")
    }

    /// The shown value (stated empty in italic when there is none), its state
    /// and the summary of the records behind it.
    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
            if let lead = row.lead {
                Text(lead)
                    .font(ReconciledValueRow.font(for: row.style))
                    .foregroundStyle(PVColor.textPrimary)
                    .textSelection(.enabled)
            } else {
                Text(row.emptyText)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let badge = row.badge {
                ReconciledValueRow.badge(badge)
            }
            if let summary = row.provenanceSummary {
                Text(verbatim: summary)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    private func open(_ location: WorkspaceLocation) {
        close()
        navigation.go(to: location)
    }
}
