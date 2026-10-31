import SwiftUI

/// The records of a field as one table, shared by the Why accordion and the
/// provenance sheet (S13-11). Both draw owner group lines where
/// `ReconciliationRecordGroup.owner` is set.
///
/// - `.why`: Read as · Source · Outcome · reserved column (S9-D5). The Source
///   opens the record's Citation with its Observation in focus.
/// - `.sheet`: Read as · Citation · Outcome · Record · reserved column. The
///   citation line opens the Citation; "Source" and "Observation" links sit under
///   it. Value headings sit over runs of records that count under different values.
///
/// The last column is reserved for a future "conclude this value" action, so
/// adding it won't relayout the table (PD-8).
struct ReconciliationRecordRows: View {
    let groups: [ReconciliationRecordGroup]
    let style: ReconciledValueRowModel.ValueStyle
    let provenance: ReconciledValueRow.Provenance
    /// Opens a record's place. The sheet closes itself first (D-5).
    let open: (WorkspaceLocation) -> Void
    /// The sheet at Wide: Read as and Outcome are capped, the Citation column takes the rest.
    var isWide = false

    /// Reserved for the conclude action.
    static let actionColumnWidth: CGFloat = 88
    static let columnSpacing: CGFloat = 14

    /// The sheet's fixed columns; the Citation column takes the rest.
    static func readAsWidth(wide: Bool) -> CGFloat { wide ? 220 : 140 }
    static func outcomeWidth(wide: Bool) -> CGFloat { wide ? 280 : 190 }
    static let recordWidth: CGFloat = 112

    private var isSheet: Bool { provenance == .sheet }
    private var columns: Int { isSheet ? 5 : 4 }

    var body: some View {
        Grid(alignment: .leadingFirstTextBaseline, horizontalSpacing: Self.columnSpacing, verticalSpacing: 0) {
            if !isSheet {
                GridRow {
                    heading(L10n.Conclusions.whyReadAs)
                    heading(L10n.Conclusions.whySource)
                    heading(L10n.Conclusions.whyOutcome)
                    Color.clear.frame(width: Self.actionColumnWidth, height: 1)
                }
                .padding(.vertical, PVSpacing.space2)
                .accessibilityHidden(true)
            }
            ForEach(groups) { group in
                if let owner = group.owner {
                    Rectangle()
                        .fill(isSheet ? PVColor.borderStrong : PVColor.borderSubtle)
                        .frame(height: 1)
                    ownerLine(owner)
                }
                let segments = segments(of: group)
                ForEach(Array(segments.enumerated()), id: \.offset) { segmentIndex, segment in
                    if let heading = segment.heading {
                        Rectangle()
                            .fill(PVColor.borderSubtle)
                            .frame(height: 1)
                        valueHeading(heading)
                    }
                    ForEach(Array(segment.records.enumerated()), id: \.element.id) { index, record in
                        // The first record under a group or value line follows the line, not a rule.
                        let followsLine = segment.heading != nil || (segmentIndex == 0 && group.owner != nil)
                        if index > 0 || !followsLine {
                            Rectangle()
                                .fill(PVColor.borderSubtle)
                                .frame(height: 1)
                        }
                        if isSheet {
                            sheetRow(record)
                        } else {
                            whyRow(record)
                        }
                    }
                }
            }
        }
    }

    /// The runs a group draws: its value groups in the sheet, one run in the Why.
    private func segments(of group: ReconciliationRecordGroup) -> [ReconciliationRecordGroup.ValueGroup] {
        if isSheet { return group.values }
        return group.records.isEmpty ? [] : [.init(heading: nil, records: group.records)]
    }

    private func whyRow(_ record: ReconciliationRecord) -> some View {
        GridRow {
            readAs(record)
                .frame(maxWidth: .infinity, alignment: .leading)
            PVButton(record.sourceTitle, variant: .link, size: .sm) {
                open(record.location)
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

    private func sheetRow(_ record: ReconciliationRecord) -> some View {
        GridRow {
            readAs(record)
                .frame(width: Self.readAsWidth(wide: isWide), alignment: .leading)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                CitationLineButton(record: record) { open(record.citationLocation) }
                HStack(spacing: PVSpacing.space5) {
                    PVButton(L10n.Conclusions.provenanceSourceLink, variant: .link, size: .sm) {
                        open(record.sourceLocation)
                    }
                    .accessibilityIdentifier("conclusion.detail.provenance.\(record.id).source")
                    PVButton(L10n.Conclusions.provenanceObservationLink(record.observationRef), variant: .link, size: .sm) {
                        open(record.location)
                    }
                    .accessibilityIdentifier("conclusion.detail.provenance.\(record.id).observation")
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            ReconciliationOutcome(mark: record.mark, phrase: record.phrase)
                .frame(width: Self.outcomeWidth(wide: isWide), alignment: .leading)
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: record.observationRef)
                Text(verbatim: record.subjectRef)
            }
            .font(PVFont.mono(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textMuted)
            .textSelection(.enabled)
            .frame(width: Self.recordWidth, alignment: .leading)
            .accessibilityElement(children: .combine)
            Color.clear.frame(width: Self.actionColumnWidth, height: 1)
                .accessibilityHidden(true)
        }
        .padding(.vertical, PVSpacing.space3)
    }

    private func readAs(_ record: ReconciliationRecord) -> some View {
        Text(record.readAs)
            .font(readAsFont(record))
            .foregroundStyle(record.counted ? PVColor.textPrimary : PVColor.textSecondary)
    }

    /// A group's heading: the owning entity as its list titles it, then its
    /// ref, then "· shown" on the owner behind the shown value. Text, never a link.
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
            if owner.shown {
                Text(L10n.Conclusions.provenanceOwnerShownMark)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .fixedSize(horizontal: false, vertical: true)
        .frame(maxWidth: .infinity, alignment: .leading)
        .gridCellColumns(columns)
        .padding(.top, PVSpacing.space5)
        .padding(.bottom, PVSpacing.space2)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.isHeader)
        .accessibilityLabel(owner.spoken)
    }

    /// Over a run of records that count under one displayed value, or under none.
    @ViewBuilder
    private func valueHeading(_ heading: ReconciliationRecordGroup.ValueGroup.Heading) -> some View {
        Group {
            switch heading {
            case .value(let value):
                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                    Text(verbatim: value.text)
                        .font(ReconciledValueRow.font(for: style))
                        .foregroundStyle(PVColor.textPrimary)
                    Text(verbatim: value.support)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
                .accessibilityElement(children: .combine)
            case .otherRecords:
                Text(L10n.Conclusions.provenanceOtherRecords)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .fixedSize(horizontal: false, vertical: true)
        .frame(maxWidth: .infinity, alignment: .leading)
        .gridCellColumns(columns)
        .padding(.top, PVSpacing.space4)
        .padding(.bottom, PVSpacing.space2)
        .accessibilityAddTraits(.isHeader)
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

    /// The sheet's column headings, drawn above its scroll with the same widths as the table.
    struct SheetHeadings: View {
        var isWide = false

        var body: some View {
            HStack(alignment: .firstTextBaseline, spacing: ReconciliationRecordRows.columnSpacing) {
                Text(L10n.Conclusions.whyReadAs)
                    .frame(width: ReconciliationRecordRows.readAsWidth(wide: isWide), alignment: .leading)
                Text(L10n.Conclusions.provenanceColumnCitation)
                    .frame(maxWidth: .infinity, alignment: .leading)
                Text(L10n.Conclusions.whyOutcome)
                    .frame(width: ReconciliationRecordRows.outcomeWidth(wide: isWide), alignment: .leading)
                Text(L10n.Conclusions.provenanceColumnRecord)
                    .frame(width: ReconciliationRecordRows.recordWidth, alignment: .leading)
                Color.clear.frame(width: ReconciliationRecordRows.actionColumnWidth, height: 1)
            }
            .pvMicroCaps()
            .foregroundStyle(PVColor.textMuted)
            .padding(.vertical, PVSpacing.space2)
            .accessibilityHidden(true)
        }
    }
}

/// A record's citation line as a link: ink by whether the record counted (so the
/// table still reads by ink), underlined at 35%, wrapping, never truncated (D-18).
private struct CitationLineButton: View {
    let record: ReconciliationRecord
    let action: () -> Void

    @State private var hovering = false

    var body: some View {
        let ink = record.counted ? PVColor.textPrimary : PVColor.textSecondary
        Button(action: action) {
            Text(verbatim: record.citationLine)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(hovering ? PVColor.textLinkHover : ink)
                .underline(true, color: (hovering ? PVColor.textLinkHover : ink).opacity(0.35))
                .multilineTextAlignment(.leading)
                .fixedSize(horizontal: false, vertical: true)
                .frame(maxWidth: .infinity, alignment: .leading)
                .contentShape(Rectangle())
                .textSelection(.enabled)
        }
        .buttonStyle(.plain)
        .onHover { hovering = $0 }
        .accessibilityLabel(Text(verbatim: record.citationLine))
        .accessibilityIdentifier("conclusion.detail.provenance.\(record.id).citation")
    }
}
