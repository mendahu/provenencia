import SwiftUI

/// An outcome as one mark and one phrase. The mark is decorative; the phrase
/// carries the meaning, so it reads without colour.
struct PVProvenanceOutcome: View {
    let mark: PVSymbol
    let phrase: String

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            PVIcon(mark, size: 14)
                .alignmentGuide(.firstTextBaseline) { $0[VerticalAlignment.center] + 4 }
            Text(phrase)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
        }
        .foregroundStyle(PVColor.textSecondary)
    }
}

/// What a record was read as, inked by whether it counted.
struct PVProvenanceReadAs: View {
    let record: PVProvenanceSheetContent.Record
    let face: PVProvenanceFace

    var body: some View {
        Text(record.readAs)
            .font(face.readAsFont(isPhrase: record.readAsIsPhrase))
            .foregroundStyle(record.counted ? PVColor.textPrimary : PVColor.textSecondary)
    }
}

/// The state of a reconciled value as a badge, with an icon so it never reads by colour alone.
struct PVProvenanceStateBadge: View {
    let state: ReconciledValueDisplay.StateBadge

    init(_ state: ReconciledValueDisplay.StateBadge) {
        self.state = state
    }

    var body: some View {
        switch state {
        case .merged: PVBadge(L10n.Conclusions.badgeMerged, tone: .neutral, icon: .merge)
        case .mixed: PVBadge(L10n.Conclusions.badgeMixed, tone: .warning, icon: .split)
        case .concluded: PVBadge(L10n.Conclusions.badgeConcluded, tone: .success, icon: .stamp)
        }
    }
}

/// One value with its Source count beside it.
struct PVProvenanceValueLine: View {
    let value: PVProvenanceSheetContent.Value
    let face: PVProvenanceFace

    init(_ value: PVProvenanceSheetContent.Value, face: PVProvenanceFace) {
        self.value = value
        self.face = face
    }

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
            Text(verbatim: value.text)
                .font(face.valueFont)
                .foregroundStyle(PVColor.textPrimary)
                .textSelection(.enabled)
            Text(verbatim: value.support)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        }
    }
}

/// What a value line shows: every kept value of a multi-valued field, else the
/// lead, else the field's empty text in italic.
struct PVProvenanceValueList: View {
    let shown: PVProvenanceSheetContent.Shown

    init(_ shown: PVProvenanceSheetContent.Shown) {
        self.shown = shown
    }

    var body: some View {
        if !shown.listedValues.isEmpty {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                ForEach(shown.listedValues) { value in
                    PVProvenanceValueLine(value, face: shown.face)
                }
            }
        } else if let lead = shown.lead {
            Text(lead)
                .font(shown.face.valueFont)
                .foregroundStyle(PVColor.textPrimary)
                .textSelection(.enabled)
        } else {
            Text(shown.emptyText)
                .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }
}

/// A group's heading: the owning entity as its list titles it, then its ref,
/// then "· shown" on the owner behind the shown value. Text, never a link.
struct PVProvenanceOwnerLine: View {
    let owner: PVProvenanceSheetContent.Group.OwnerLine

    init(_ owner: PVProvenanceSheetContent.Group.OwnerLine) {
        self.owner = owner
    }

    var body: some View {
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
        .padding(.top, PVSpacing.space5)
        .padding(.bottom, PVSpacing.space2)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.isHeader)
        .accessibilityLabel(owner.spoken)
    }
}

/// Over a run of records that count under one displayed value, or under none.
struct PVProvenanceValueHeading: View {
    let heading: PVProvenanceSheetContent.Group.ValueGroup.Heading
    let face: PVProvenanceFace

    var body: some View {
        Group {
            switch heading {
            case .value(let value):
                PVProvenanceValueLine(value, face: face)
                    .accessibilityElement(children: .combine)
            case .otherRecords:
                Text(L10n.Conclusions.provenanceOtherRecords)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .fixedSize(horizontal: false, vertical: true)
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.top, PVSpacing.space4)
        .padding(.bottom, PVSpacing.space2)
        .accessibilityAddTraits(.isHeader)
    }
}

/// The column headings, drawn above a table's scroll with the same widths as its rows.
struct PVProvenanceColumnHeadings: View {
    var columns = PVProvenanceColumns()

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: columns.spacing) {
            if columns.showsReadAs {
                Text(L10n.Conclusions.whyReadAs)
                    .frame(width: columns.readAsWidth, alignment: .leading)
            }
            Text(L10n.Conclusions.provenanceColumnCitation)
                .frame(maxWidth: .infinity, alignment: .leading)
            Text(L10n.Conclusions.whyOutcome)
                .frame(width: columns.outcomeWidth, alignment: .leading)
            Text(L10n.Conclusions.provenanceColumnRecord)
                .frame(width: PVProvenanceColumns.recordWidth, alignment: .leading)
            Color.clear.frame(width: PVProvenanceColumns.actionWidth, height: 1)
        }
        .pvMicroCaps()
        .foregroundStyle(PVColor.textMuted)
        .padding(.vertical, PVSpacing.space2)
        .accessibilityHidden(true)
    }
}
