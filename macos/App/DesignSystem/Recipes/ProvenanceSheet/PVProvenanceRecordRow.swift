import SwiftUI

/// One record as a row: Read as · Citation · Outcome · Record · a reserved
/// column. The citation line opens the Citation; "Source" and "Observation"
/// links sit under it, and a button beside it copies the line. Place a row
/// alone (with `PVProvenanceColumns` setting its cells) or as a row of a
/// `PVProvenanceRecordTable`; the two line up when they share the columns.
///
/// Every cell but Citation has a fixed width, so a row stands alone and still
/// lines up with its neighbours (S13-36 D-7). A link calls `open` with the
/// record and the link; the caller maps it to a place.
struct PVProvenanceRecordRow: View {
    let record: PVProvenanceSheetContent.Record
    /// How the Read as cell reads; used only with Read as on.
    var face: PVProvenanceFace = .text
    var columns = PVProvenanceColumns()
    /// The prefix of the row's accessibility ids: `<prefix>.<record id>.citation`, `.copy`, `.source`, `.observation`.
    let idPrefix: String
    let open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: columns.spacing) {
            if columns.showsReadAs {
                PVProvenanceReadAs(record: record, face: face)
                    .frame(width: columns.readAsWidth, alignment: .leading)
                    // The record as one element; the citation line and the links follow as controls.
                    .accessibilityElement(children: .ignore)
                    .accessibilityLabel(record.spoken)
            }
            citation
                .frame(maxWidth: .infinity, alignment: .leading)
            PVProvenanceOutcome(mark: record.mark, phrase: record.phrase)
                .frame(width: columns.outcomeWidth, alignment: .leading)
                .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: record.observationRef)
                Text(verbatim: record.subjectRef)
            }
            .font(PVFont.mono(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textMuted)
            .textSelection(.enabled)
            .frame(width: PVProvenanceColumns.recordWidth, alignment: .leading)
            .accessibilityHidden(true)
            PVProvenanceReservedCell()
        }
        // With no Read as cell the record's one spoken line sits at the row's
        // leading edge, ahead of the citation and the links, and takes no room.
        .overlay(alignment: .topLeading) {
            if !columns.showsReadAs {
                Color.clear.frame(width: 1, height: 1)
                    .accessibilityElement(children: .ignore)
                    .accessibilityLabel(record.spoken)
            }
        }
        .padding(.vertical, PVSpacing.space3)
    }

    private var citation: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(alignment: .citationFirstLineCenter, spacing: PVProvenanceColumns.copyTrackSpacing) {
                PVInkLink(ink: record.counted ? .primary : .secondary, action: { open(record, .citation) }) {
                    Text(verbatim: record.citation.line)
                }
                    .accessibilityLabel(Text(verbatim: record.citation.line))
                    .accessibilityIdentifier("\(idPrefix).\(record.id).citation")
                    // The first line's centre sits about a third of the type size above its baseline.
                    .alignmentGuide(.citationFirstLineCenter) { $0[.firstTextBaseline] - PVTypeScale.bodySmall / 3 }
                PVClipboardButton(
                    text: record.citation.line,
                    label: L10n.ProvenanceSheet.copyCitation,
                    copiedAnnouncement: L10n.string(L10n.ProvenanceSheet.citationCopied)
                )
                .frame(width: PVProvenanceColumns.copyTrackWidth)
                .alignmentGuide(.citationFirstLineCenter) { $0.height / 2 }
                .accessibilityIdentifier("\(idPrefix).\(record.id).copy")
            }
            HStack(spacing: PVSpacing.space5) {
                PVButton(L10n.ProvenanceSheet.sourceLink, variant: .link, size: .sm) {
                    open(record, .source)
                }
                .accessibilityLabel(L10n.ProvenanceSheet.sourceLinkSpoken(record.citation.title))
                .accessibilityIdentifier("\(idPrefix).\(record.id).source")
                PVButton(L10n.ProvenanceSheet.observationLink(record.observationRef), variant: .link, size: .sm) {
                    open(record, .observation)
                }
                .accessibilityIdentifier("\(idPrefix).\(record.id).observation")
            }
        }
    }
}

private extension VerticalAlignment {
    private enum CitationFirstLineCenter: AlignmentID {
        static func defaultValue(in context: ViewDimensions) -> CGFloat { context[.top] }
    }

    /// The centre of a citation line's first line, so the copy button sits level with it at any line count.
    static let citationFirstLineCenter = VerticalAlignment(CitationFirstLineCenter.self)
}
