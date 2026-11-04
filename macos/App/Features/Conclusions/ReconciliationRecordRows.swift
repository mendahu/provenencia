import SwiftUI

/// The records of a field as the Why accordion's table: Read as · Source ·
/// Outcome · reserved column (S9-D5), with owner group lines where
/// `PVProvenanceSheetContent.Group.owner` is set. The Source opens the record's
/// Citation with its Observation in focus (`open` reports `.observation`). The
/// table is the Why's own; its cells and columns are the provenance recipe's
/// (`PVProvenanceOwnerLine`, `PVProvenanceOutcome`, `PVProvenanceReadAs`,
/// `PVProvenanceColumns`), until the Event and Place pages are redesigned.
///
/// The last column is reserved for a future "conclude this value" action, so
/// adding it won't relayout the table (PD-8).
struct ReconciliationRecordRows: View {
    let groups: [PVProvenanceSheetContent.Group]
    let face: PVProvenanceFace
    /// Opens a record's place.
    let open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void

    /// The Why's four columns: Read as, Source, Outcome and the reserved one.
    private let columns = 4

    var body: some View {
        Grid(alignment: .leadingFirstTextBaseline, horizontalSpacing: PVProvenanceColumns.spacing, verticalSpacing: 0) {
            GridRow {
                heading(L10n.ProvenanceSheet.columnReadAs)
                heading(L10n.Conclusions.whySource)
                heading(L10n.ProvenanceSheet.columnOutcome)
                PVProvenanceReservedCell()
            }
            .padding(.vertical, PVSpacing.space2)
            .accessibilityHidden(true)
            ForEach(groups) { group in
                if let owner = group.owner {
                    PVDivider()
                    PVProvenanceOwnerLine(owner)
                        .gridCellColumns(columns)
                }
                ForEach(Array(group.records.enumerated()), id: \.element.id) { index, record in
                    // The first record under a group line follows the line, not a rule.
                    if index > 0 || group.owner == nil {
                        PVDivider()
                    }
                    whyRow(record)
                }
            }
        }
    }

    private func whyRow(_ record: PVProvenanceSheetContent.Record) -> some View {
        GridRow {
            PVProvenanceReadAs(record: record, face: face)
                .frame(maxWidth: .infinity, alignment: .leading)
            PVButton(record.citation.title, variant: .link, size: .sm) {
                open(record, .observation)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .help(record.citation.title)
            .accessibilityIdentifier("conclusion.detail.why.\(record.id).source")
            PVProvenanceOutcome(mark: record.mark, phrase: record.phrase)
                .frame(maxWidth: .infinity, alignment: .leading)
            PVProvenanceReservedCell()
        }
        .padding(.vertical, PVSpacing.space3)
    }

    private func heading(_ text: LocalizedStringResource) -> some View {
        Text(text)
            .pvMicroCaps()
            .foregroundStyle(PVColor.textMuted)
            .frame(maxWidth: .infinity, alignment: .leading)
    }
}
