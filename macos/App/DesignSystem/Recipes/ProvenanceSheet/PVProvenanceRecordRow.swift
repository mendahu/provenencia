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
    let face: PVProvenanceFace
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
            outcome
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: record.observationRef)
                Text(verbatim: record.subjectRef)
            }
            .font(PVFont.mono(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textMuted)
            .textSelection(.enabled)
            .frame(width: PVProvenanceColumns.recordWidth, alignment: .leading)
            .accessibilityHidden(true)
            Color.clear.frame(width: PVProvenanceColumns.actionWidth, height: 1)
                .accessibilityHidden(true)
        }
        .padding(.vertical, PVSpacing.space3)
    }

    /// With no Read as cell the Outcome cell carries the record's one spoken line.
    @ViewBuilder
    private var outcome: some View {
        let cell = PVProvenanceOutcome(mark: record.mark, phrase: record.phrase)
            .frame(width: columns.outcomeWidth, alignment: .leading)
        if columns.showsReadAs {
            cell.accessibilityHidden(true)
        } else {
            cell
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(record.spoken)
        }
    }

    private var citation: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(alignment: .citationFirstLineCenter, spacing: PVProvenanceColumns.copyTrackSpacing) {
                CitationLineButton(record: record, idPrefix: idPrefix) { open(record, .citation) }
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

/// A record's citation line as a link: ink by whether the record counted (so the
/// table still reads by ink), underlined at 35%, wrapping, never truncated (D-18).
private struct CitationLineButton: View {
    let record: PVProvenanceSheetContent.Record
    let idPrefix: String
    let action: () -> Void

    @FocusState private var focused: Bool

    var body: some View {
        Button(action: action) {
            Text(verbatim: record.citation.line)
                .multilineTextAlignment(.leading)
                .fixedSize(horizontal: false, vertical: true)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(CitationLineButtonStyle(counted: record.counted))
        .focused($focused)
        .pvFocusRing(focused, cornerRadius: PVRadius.xs)
        .accessibilityLabel(Text(verbatim: record.citation.line))
        .accessibilityIdentifier("\(idPrefix).\(record.id).citation")
    }
}

/// Kit hover and press on the citation line: link ink on hover, else primary when
/// the record counted, secondary when it did not.
private struct CitationLineButtonStyle: ButtonStyle {
    let counted: Bool

    func makeBody(configuration: Configuration) -> some View {
        PVHoverEffect(isPressed: configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            let base = counted ? PVColor.textPrimary : PVColor.textSecondary
            let ink = showHover || configuration.isPressed ? PVColor.textLinkHover : base
            configuration.label
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(ink)
                .underline(true, color: ink.opacity(0.35))
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
