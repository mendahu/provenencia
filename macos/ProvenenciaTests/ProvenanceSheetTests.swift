import Foundation
import Testing
@testable import Provenencia

struct ProvenanceSheetTests {
    private typealias D = ReconciledValueDisplayTests
    private let en = Locale(identifier: "en_US")

    private func row(_ id: String, records: Int) -> ReconciledValueRowModel {
        let outcomes = (0..<records).map { D.outcome("\(id)-\($0)", "kept") }
        let field = D.field("single", key: "date", label: "Birth date", valueType: "date", [D.value(1, "kept")], outcomes: outcomes)
        var model = ReconciledValueRowModel(field: field, locale: en)
        model.id = id
        return model
    }

    @Test func theSheetFollowsItsRowAcrossARefresh() {
        let item = ProvenanceSheetItem(id: "life.birthDate")
        #expect(item.row(in: [row("name", records: 1), row("life.birthDate", records: 3)])?.records.count == 3)
        #expect(item.row(in: [row("name", records: 1), row("life.birthDate", records: 4)])?.records.count == 4)
        #expect(item.row(in: [row("name", records: 1)]) == nil)
        #expect(item.row(in: [row("life.birthDate", records: 0)]) == nil)
    }

    private func record(_ id: String, rank: Int, reason: String = "kept") -> ReconciliationRecord {
        ReconciliationRecord(
            outcome: D.outcome(id, reason, rank: rank == 0 ? nil : rank, groupRank: rank),
            in: D.field("mixed"), locale: en
        )
    }

    @Test func valuesGroupTheirRecordsAndTheRestComeLast() {
        let field = D.field("mixed", [
            D.value(1, "kept", support: 2, .name(D.jamesRobins)),
            D.value(2, "kept", .name(CatalogNameValue(form: "James Robbins"))),
        ])
        func groups(_ ranks: [Int], in field: CatalogConclusionField) -> [ReconciliationRecordGroup.ValueGroup] {
            let records = ranks.enumerated().map { record("r\($0.offset)", rank: $0.element) }
            return ReconciledValueRowModel.valueGroups(records, field: field, locale: en, orders: NameFormatOrders())
        }

        let mixed = groups([1, 1, 2, 0], in: field)
        #expect(mixed.map(\.records.count) == [2, 1, 1])
        guard case .value(let first) = mixed[0].heading, case .value(let second) = mixed[1].heading else {
            Issue.record("expected value headings, got \(mixed.map(\.heading))")
            return
        }
        #expect(first.rank == 1 && first.support == "2 Sources")
        #expect(second.rank == 2 && second.text == "James Robbins")
        #expect(mixed[2].heading == .otherRecords)

        // A merged name: every record went into the one displayed value.
        let merged = groups([1, 1, 1], in: field)
        #expect(merged.count == 1 && merged[0].heading == nil)

        // A birth date with every record lost still says so.
        let lost = groups([0, 0], in: D.field("", key: "date", valueType: "date"))
        #expect(lost.count == 1 && lost[0].heading == .otherRecords)

        // A rank with no displayed value reads as other records.
        let orphan = groups([3, 0], in: field)
        #expect(orphan.map(\.heading) == [.otherRecords, .otherRecords])
    }

    @Test func placeOwnersEachHoldOneValueWithoutHeadings() {
        let toronto = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let yorkville = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        func owner(_ entity: CatalogCanonicalEntity, _ name: String, _ id: String) -> CatalogOwnerField {
            CatalogOwnerField(entity: entity, event: nil, field: D.field(
                "single", key: "toponym", label: "toponym", valueType: "text", [D.value(1, "kept", .text(name))],
                outcomes: [D.outcome(id, "kept")]
            ))
        }
        let row = ReconciledValueRowModel.derived(
            id: "life.deathPlace", label: "Death place", style: .place, lead: "Toronto", emptyText: "", mixed: true,
            owners: [owner(toronto, "Toronto", "t"), owner(yorkville, "Yorkville", "y")], shownOwnerID: "pl1",
            kind: .deathPlace, locale: en
        )
        #expect(row.recordGroups.count == 2)
        #expect(row.recordGroups.allSatisfy { $0.values.count == 1 && $0.values[0].heading == nil })
    }

    @Test func theSummarySaysOnlyWhatTheRowKnows() {
        let merged = D.field("merged", [
            D.value(1, "kept", support: 3, against: 1, .name(D.jamesRobins)),
        ], outcomes: (0..<4).map { D.outcome("r\($0)", "kept", source: "s\($0 % 3)") })
        #expect(ReconciledValueRowModel(field: merged, locale: en).provenanceSummary
            == "merged from 3 Sources · 1 record disagrees · 4 records")

        let first = CatalogCanonicalEntity(id: "ev1", ref: "EVT-1", subjectTypeID: "t", label: "")
        let second = CatalogCanonicalEntity(id: "ev2", ref: "EVT-2", subjectTypeID: "t", label: "")
        func owner(_ entity: CatalogCanonicalEntity, _ ids: [String]) -> CatalogOwnerField {
            CatalogOwnerField(entity: entity, event: nil, field: D.field(
                "single", key: "date", label: "date", valueType: "date", [], outcomes: ids.map { D.outcome($0, "kept") }
            ))
        }
        let two = ReconciledValueRowModel.derived(
            id: "life.birthDate", label: "Birth date", style: .date, lead: "1822", emptyText: "", mixed: true,
            owners: [owner(first, ["a", "b", "c"]), owner(second, ["d"])], shownOwnerID: "ev1",
            kind: .birthDate, locale: en
        )
        #expect(two.provenanceSummary == "4 records from 2 birth events")
        let one = ReconciledValueRowModel.derived(
            id: "life.birthDate", label: "Birth date", style: .date, lead: "1822", emptyText: "", mixed: false,
            owners: [owner(first, ["a", "b"])], shownOwnerID: "ev1", kind: .birthDate, locale: en
        )
        #expect(one.provenanceSummary == "2 records")
        #expect(ReconciledValueRowModel.empty(id: "x", label: "X", style: .text, emptyText: "").provenanceSummary == nil)
    }
}
