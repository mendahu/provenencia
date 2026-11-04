import SwiftUI

/// Every record behind one value, and what the auto-reconciler did with each:
/// the value and its state, a summary, then the records grouped by owning
/// entity and by value, under their column headings. A record's links call
/// `open` after the sheet closes (`\.pvSheetClose`), so a caller maps a link to
/// a place without closing anything itself. Present it through `pvSheet`.
///
/// The records scroll inside the sheet past `recordsMaxHeight`; the title,
/// header, column headings and Done stay.
struct PVProvenanceSheet: View {
    let content: PVProvenanceSheetContent
    var columns = PVProvenanceColumns()
    /// The prefix of the row ids (`PVProvenanceRecordRow`) and of the Done button's id (`<prefix>.done`).
    let idPrefix: String
    let open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void

    @Environment(\.pvSheetClose) private var sheetClose

    static let recordsMaxHeight: CGFloat = 560

    var body: some View {
        PVPanel(
            title: Text(verbatim: content.title).accessibilityHeading(.h1),
            width: columns.panelWidth
        ) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                header
                // Room for "Concluded by you" (Reconciliation Claims): zero height until then.
                Color.clear
                    .frame(height: 0)
                    .accessibilityHidden(true)
                VStack(alignment: .leading, spacing: 0) {
                    PVProvenanceColumnHeadings(columns: columns)
                    PVDivider()
                    ScrollView {
                        PVProvenanceRecordTable(
                            groups: content.groups, face: content.shown.face, columns: columns,
                            idPrefix: idPrefix, open: close
                        )
                    }
                    .frame(maxHeight: Self.recordsMaxHeight)
                }
            }
        } footer: {
            HStack {
                Spacer(minLength: 0)
                PVButton(L10n.ProvenanceSheet.done, variant: .primary) { sheetClose() }
                    .keyboardShortcut(.defaultAction)
                    .accessibilityIdentifier("\(idPrefix).done")
            }
        }
        .accessibilityElement(children: .contain)
    }

    /// The shown value or values (stated empty in italic when there is none), its state
    /// and the summary of the records behind it.
    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
            PVProvenanceValueList(content.shown)
            if let badge = content.shown.badge {
                PVProvenanceStateBadge(badge)
            }
            if let summary = content.summary {
                Text(verbatim: summary)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    /// A link closes the sheet, then opens its place.
    private func close(_ record: PVProvenanceSheetContent.Record, _ link: PVProvenanceLink) {
        sheetClose { open(record, link) }
    }
}

#Preview("Provenance sheet") {
    PVProvenanceSheet(
        content: PVProvenancePreviewData.content(groups: PVProvenancePreviewData.groups, summary: "merged from 2 Sources · 2 records"),
        idPrefix: "preview",
        open: { _, _ in }
    )
}

#Preview("Thirty records") {
    let records = (1...30).map {
        PVProvenancePreviewData.record(
            "r\($0)", readAs: "14 May 1817", title: "Parish register \($0)", detail: "p. \($0)", counted: $0 % 5 != 0
        )
    }
    PVProvenanceSheet(
        content: PVProvenancePreviewData.content(
            groups: [PVProvenanceSheetContent.Group(owner: nil, records: records)], summary: "30 records"
        ),
        idPrefix: "preview",
        open: { _, _ in }
    )
}

#Preview("Dark, owner groups") {
    PVProvenanceSheet(
        content: PVProvenancePreviewData.content(groups: PVProvenancePreviewData.ownerGroups, summary: "merged from 3 records"),
        idPrefix: "preview",
        open: { _, _ in }
    )
    .preferredColorScheme(.dark)
}

#Preview("Dark, one record") {
    PVProvenanceSheet(
        content: PVProvenancePreviewData.content(
            groups: [PVProvenanceSheetContent.Group(owner: nil, records: [PVProvenancePreviewData.census])],
            summary: "1 record"
        ),
        idPrefix: "preview",
        open: { _, _ in }
    )
    .preferredColorScheme(.dark)
}
