import SwiftUI

/// The records of one value, in groups, as a table of `PVProvenanceRecordRow`s:
/// a group line over each owning entity (when the content names an owner), a
/// value heading over each run of records that count under a different value,
/// and a rule between rows. The first record under a line follows the line, not
/// a rule. Put `PVProvenanceColumnHeadings` above it, sharing its columns.
struct PVProvenanceRecordTable: View {
    let groups: [PVProvenanceSheetContent.Group]
    let face: PVProvenanceFace
    var columns = PVProvenanceColumns()
    /// The prefix of every row's accessibility ids (`PVProvenanceRecordRow`).
    let idPrefix: String
    let open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            ForEach(groups) { group in
                if let owner = group.owner {
                    rule(PVColor.borderStrong)
                    PVProvenanceOwnerLine(owner)
                }
                ForEach(Array(group.values.enumerated()), id: \.offset) { valueIndex, run in
                    if let heading = run.heading {
                        rule(PVColor.borderSubtle)
                        PVProvenanceValueHeading(heading: heading, face: face)
                    }
                    ForEach(Array(run.records.enumerated()), id: \.element.id) { index, record in
                        // The first record under a group or value line follows the line, not a rule.
                        let followsLine = run.heading != nil || (valueIndex == 0 && group.owner != nil)
                        if index > 0 || !followsLine {
                            rule(PVColor.borderSubtle)
                        }
                        PVProvenanceRecordRow(
                            record: record, face: face, columns: columns, idPrefix: idPrefix, open: open
                        )
                    }
                }
            }
        }
    }

    private func rule(_ color: Color) -> some View {
        Rectangle()
            .fill(color)
            .frame(height: 1)
    }
}

// MARK: - Previews

enum PVProvenancePreviewData {
    static func record(
        _ id: String, readAs: String, title: String, detail: String, counted: Bool = true,
        phrase: String = "Kept", mark: PVSymbol = .check
    ) -> PVProvenanceSheetContent.Record {
        let outcome = CatalogReconcilerOutcome(
            observationID: id, observationRef: "OBS-\(id.uppercased())", reason: .kept, valueRank: 1,
            deniedByObservationID: "", recorded: .text(readAs), subjectID: "sub-\(id)", subjectRef: "CPR-\(id.uppercased())",
            citationID: "cit-\(id)", sourceID: "src-\(id)", sourceTitle: title, credibilityKey: "",
            transcriptionUncertain: false
        )
        return PVProvenanceSheetContent.Record(
            id: id, readAs: readAs, readAsIsPhrase: false, counted: counted, phrase: phrase, mark: mark,
            observationRef: outcome.observationRef, subjectRef: outcome.subjectRef, outcome: outcome,
            citation: .init(title: title, titleIsRef: false, detail: detail, line: "\(title), \(detail)")
        )
    }

    static let census = record("census", readAs: "14 May 1817", title: "Census of Canada West, 1851", detail: "Vol. 3, p. 23")
    static let baptism = record(
        "baptism", readAs: "1817", title: "Baptism register, St James, York", detail: "image 37",
        counted: false, phrase: "Outvoted by 2 Sources", mark: .scale
    )
    static let groups = [PVProvenanceSheetContent.Group(owner: nil, records: [census, baptism])]
}

#Preview("One record, with and without Read as") {
    VStack(alignment: .leading, spacing: 24) {
        PVProvenanceRecordRow(
            record: PVProvenancePreviewData.census, face: .date, idPrefix: "preview", open: { _, _ in }
        )
        PVProvenanceRecordRow(
            record: PVProvenancePreviewData.census, face: .date,
            columns: PVProvenanceColumns(showsReadAs: false), idPrefix: "preview", open: { _, _ in }
        )
    }
    .padding(24)
    .frame(width: 880)
    .background(PVColor.surfaceCard)
}

#Preview("A lone row above a table") {
    let columns = PVProvenanceColumns()
    VStack(alignment: .leading, spacing: 0) {
        PVProvenanceColumnHeadings(columns: columns)
        PVProvenanceRecordRow(
            record: PVProvenancePreviewData.census, face: .date, columns: columns, idPrefix: "preview", open: { _, _ in }
        )
        PVDivider()
        PVProvenanceRecordTable(
            groups: PVProvenancePreviewData.groups, face: .date, columns: columns, idPrefix: "preview", open: { _, _ in }
        )
    }
    .padding(24)
    .frame(width: 880)
    .background(PVColor.surfaceCard)
}

#Preview("Wide, with groups") {
    let columns = PVProvenanceColumns(widthClass: .wide)
    let groups = [
        PVProvenanceSheetContent.Group(
            owner: .init(id: "ev1", title: "Birth of Ada Lovelace", ref: "EVT-4KX2M", spoken: "Birth of Ada Lovelace", shown: true),
            records: [PVProvenancePreviewData.census]
        ),
        PVProvenanceSheetContent.Group(
            owner: .init(id: "ev2", title: "Birth of Ada Lovelace", ref: "EVT-9PQ2R", spoken: "Birth of Ada Lovelace"),
            values: [
                .init(heading: .value(.init(rank: 1, text: "1817", support: "2 Sources")), records: [PVProvenancePreviewData.baptism]),
                .init(heading: .otherRecords, records: [PVProvenancePreviewData.census]),
            ]
        ),
    ]
    VStack(alignment: .leading, spacing: 0) {
        PVProvenanceColumnHeadings(columns: columns)
        PVDivider()
        PVProvenanceRecordTable(groups: groups, face: .date, columns: columns, idPrefix: "preview", open: { _, _ in })
    }
    .padding(24)
    .frame(width: 1280)
    .background(PVColor.surfaceCard)
}
