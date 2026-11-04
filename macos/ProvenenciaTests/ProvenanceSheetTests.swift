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
        #expect(item.row(in: [row("name", records: 1), row("life.birthDate", records: 3)])?.recordCount == 3)
        #expect(item.row(in: [row("name", records: 1), row("life.birthDate", records: 4)])?.recordCount == 4)
        #expect(item.row(in: [row("name", records: 1)]) == nil)
        #expect(item.row(in: [row("life.birthDate", records: 0)]) == nil)
    }

    private func record(_ id: String, rank: Int, reason: String = "kept") -> PVProvenanceSheetContent.Record {
        PVProvenanceSheetContent.Record(
            outcome: D.outcome(id, reason, rank: rank == 0 ? nil : rank, groupRank: rank),
            in: D.field("mixed"), locale: en
        )
    }

    @Test func aFaceComesFromTheField() {
        let date = D.field("single", key: "date", label: "Birth date", valueType: "date", [D.value(1, "kept")], outcomes: [])
        let text = D.field("single", key: "name", label: "Name", valueType: "text", [D.value(1, "kept")], outcomes: [])
        #expect(PVProvenanceFace(field: date) == .date)
        #expect(PVProvenanceFace(field: text) == .text)
    }

    @Test func aRowBuildsItsContentOnCall() {
        let row = row("life.birthDate", records: 3)
        #expect(row.recordCount == 3)
        let content = row.provenanceContent()
        #expect(content.records.map(\.id) == ["life.birthDate-0", "life.birthDate-1", "life.birthDate-2"])
        #expect(content.title == "Birth date" && content.shown == row.shown)
        #expect(content.summary?.hasSuffix("3 records") == true)

        let empty = ReconciledValueRowModel.empty(id: "x", label: "X", face: .text, emptyText: "")
        #expect(empty.recordCount == 0)
        #expect(empty.provenanceContent().records.isEmpty)

        // A relabelled row and its sheet cannot disagree.
        var renamed = row
        renamed.label = "Born"
        #expect(renamed.provenanceContent().title == "Born")
    }

    @Test func aRecordComesFromOneOutcome() {
        let outcome = D.outcome("bible", "outvoted", rank: 2, source: "S1", title: "Family Bible", groupRank: 1)
        let field = D.field("merged", [
            D.value(1, "kept", support: 2, .name(D.jamesRobins)),
            D.value(2, "kept", .name(CatalogNameValue(form: "James Robbins"))),
        ], outcomes: [outcome])
        let record = PVProvenanceSheetContent.Record(outcome: outcome, in: field, locale: en)
        #expect(record.id == "bible" && record.outcome == outcome)
        #expect(record.readAs == ReconciledValueDisplay.readAs(outcome, propertyKey: field.propertyKey, locale: en, orders: NameFormatOrders()))
        #expect(record.readAsIsPhrase == ReconciledValueDisplay.readAsIsPhrase(outcome, propertyKey: field.propertyKey))
        #expect(!record.counted)
        #expect(record.phrase == ReconciledValueDisplay.outcomePhrase(outcome, in: field, locale: en))
        #expect(record.mark == ReconciledValueDisplay.outcomeMark(outcome, in: field))
        #expect(record.citation == CitationLineDisplay.parts(outcome))
        #expect(record.observationRef == "OBS-bible" && record.subjectRef == "CPR-bible")
    }

    @Test func linksReachTheirPlaces() {
        let outcome = D.outcome("bible", "kept", source: "S1", title: "Family Bible")
        let record = PVProvenanceSheetContent.Record(outcome: outcome, in: D.field("single"), locale: en)
        #expect(WorkspaceLocation.provenance(record, .source) == .sourcePage(sourceId: "S1", ref: nil, title: "Family Bible"))
        let citation = WorkspaceLocation.provenance(record, .citation)
        #expect(citation.sourceSurface == .citationComposer)
        #expect(citation.citationId == "cit-bible" && citation.observationId == nil)
        let observation = WorkspaceLocation.provenance(record, .observation)
        #expect(observation.citationId == "cit-bible" && observation.observationId == "bible")
    }

    @Test func aRecordSpeaksAsOneLine() {
        let r = record("r0", rank: 1)
        let parts = [r.readAs, r.citation.line, r.phrase, r.observationRef, r.subjectRef]
        #expect(r.spoken == parts.joined(separator: ", "))
    }

    @Test func aGalleryRecordNeedsNoField() {
        let kept = PVProvenanceSheetContent.Record(outcome: D.outcome("k", "kept", title: "Register"), in: nil, locale: en)
        #expect(ReconciledValueDisplay.counted(kept.outcome, in: nil))
        #expect(kept.counted && kept.phrase == "kept" && kept.mark == .check)
        #expect(kept.readAs.isEmpty && !kept.readAsIsPhrase)
        let expected = [kept.citation.line, kept.phrase, kept.observationRef, kept.subjectRef]
        #expect(kept.spoken == L10n.DesignSystem.spokenJoin(expected))
        #expect(kept.spoken.hasPrefix(kept.citation.line))

        let weak = PVProvenanceSheetContent.Record(
            outcome: D.outcome("w", "weak", credibilityOffset: -2), in: nil, locale: en)
        #expect(!weak.counted)
        #expect(weak.phrase == "weak · low-trust Source" && weak.mark == .signalLow)

        let against = PVProvenanceSheetContent.Record(outcome: D.outcome("a", "against"), in: nil, locale: en)
        #expect(!against.counted && against.phrase == L10n.string(L10n.Conclusions.outcomeAgainst))
    }

    @Test func valuesGroupTheirRecordsAndTheRestComeLast() {
        let field = D.field("mixed", [
            D.value(1, "kept", support: 2, .name(D.jamesRobins)),
            D.value(2, "kept", .name(CatalogNameValue(form: "James Robbins"))),
        ])
        func groups(_ ranks: [Int], in field: CatalogConclusionField) -> [PVProvenanceSheetContent.Group.ValueGroup] {
            let records = ranks.enumerated().map { record("r\($0.offset)", rank: $0.element) }
            return PVProvenanceSheetContent.valueGroups(records, field: field, locale: en, orders: NameFormatOrders())
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
        func owner(_ place: CatalogCanonicalEntity, _ name: String, _ id: String, link: String? = nil) -> CatalogOwnerField {
            CatalogOwnerField(
                entity: CatalogCanonicalEntity(id: link ?? "lk-\(id)", ref: "LOC-\(id)", subjectTypeID: "t", label: ""),
                event: nil, place: CatalogChainedPlace(entity: place, names: [name]),
                field: D.field(
                    "single", key: "place", label: "place", valueType: "subject", [D.value(1, "kept", .text(name))],
                    outcomes: [D.outcome(id, "kept")]
                )
            )
        }
        let row = ReconciledValueRowModel.derived(
            id: "life.deathPlace", label: "Death place", face: .place, lead: "Toronto", emptyText: "", mixed: true,
            owners: [owner(toronto, "Toronto", "t"), owner(yorkville, "Yorkville", "y")], shownOwnerID: "pl1",
            kind: .deathPlace, locale: en
        )
        #expect(row.provenanceContent().groups.count == 2)
        #expect(row.provenanceContent().groups.allSatisfy { $0.values.count == 1 && $0.values[0].heading == nil })
    }

    @Test func twoLinksToOnePlaceFormOneGroup() {
        let york = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let leeds = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        func owner(_ link: String, _ id: String, place: CatalogCanonicalEntity = york, _ name: String = "York") -> CatalogOwnerField {
            CatalogOwnerField(
                entity: CatalogCanonicalEntity(id: link, ref: "LOC-\(link)", subjectTypeID: "t", label: ""),
                event: nil, place: CatalogChainedPlace(entity: place, names: [name]),
                field: D.field(
                    "single", key: "place", label: "place", valueType: "subject", [D.value(1, "kept", .text(name))],
                    outcomes: [D.outcome(id, "kept")]
                )
            )
        }
        // York is the only Place: no group line, and the summary counts records.
        let only = ReconciledValueRowModel.derived(
            id: "life.birthPlace", label: "Birth place", face: .place, lead: "York", emptyText: "", mixed: false,
            owners: [owner("lk1", "a"), owner("lk2", "b")], shownOwnerID: "pl1", kind: .birthPlace, locale: en
        )
        #expect(only.provenanceContent().groups.count == 1 && only.provenanceContent().groups[0].owner == nil)
        #expect(only.provenanceContent().groups[0].records.count == 2)
        #expect(only.provenanceContent().summary == "2 records")

        // York by two links and Leeds: one group each, York's holding both links' records in link order.
        let both = ReconciledValueRowModel.derived(
            id: "life.birthPlace", label: "Birth place", face: .place, lead: "York", emptyText: "", mixed: true,
            owners: [owner("lk1", "a"), owner("lk3", "c", place: leeds, "Leeds"), owner("lk2", "b")],
            shownOwnerID: "pl1", kind: .birthPlace, locale: en
        )
        #expect(both.provenanceContent().groups.map { $0.owner?.id } == ["pl1", "pl2"])
        #expect(both.provenanceContent().groups.map { $0.owner?.title } == ["York", "Leeds"])
        #expect(both.provenanceContent().groups[0].records.map(\.id) == ["a", "b"])
        #expect(both.provenanceContent().summary == "3 records from 2 Places")
    }

    @Test func aPooledPlaceGroupHasOneHeadingPerValue() {
        let york = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let leeds = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        func owner(
            _ link: String, _ outcomes: [CatalogReconcilerOutcome],
            place: CatalogCanonicalEntity = york, _ name: String = "York"
        ) -> CatalogOwnerField {
            CatalogOwnerField(
                entity: CatalogCanonicalEntity(id: link, ref: "LOC-\(link)", subjectTypeID: "t", label: ""),
                event: nil, place: CatalogChainedPlace(entity: place, names: [name]),
                field: D.field(
                    "single", key: "place", label: "place", valueType: "subject", [D.value(1, "kept", .text(name))],
                    outcomes: outcomes
                )
            )
        }
        let row = ReconciledValueRowModel.derived(
            id: "life.birthPlace", label: "Birth place", face: .place, lead: "York", emptyText: "", mixed: true,
            owners: [
                owner("lk1", [D.outcome("a", "kept", source: "s1"), D.outcome("d", "lost", rank: nil, source: "s1")]),
                owner("lk2", [D.outcome("b", "kept", source: "s2")]),
                owner("lk3", [D.outcome("c", "kept")], place: leeds, "Leeds"),
            ],
            shownOwnerID: "pl1", kind: .birthPlace, locale: en
        )
        let values = row.provenanceContent().groups[0].values
        #expect(values.map { $0.records.map(\.id) } == [["a", "b"], ["d"]])
        guard case .value(let heading)? = values[0].heading else {
            Issue.record("expected a value heading")
            return
        }
        #expect(heading.text == "York" && heading.support == L10n.Conclusions.sourceCount(2))
        #expect(values[1].heading == .otherRecords)

        // Two links whose kept records cite the same Source: one Source, not two.
        let shared = ReconciledValueRowModel.derived(
            id: "life.birthPlace", label: "Birth place", face: .place, lead: "York", emptyText: "", mixed: true,
            owners: [
                owner("lk1", [D.outcome("a", "kept", source: "s1"), D.outcome("d", "lost", rank: nil, source: "s1")]),
                owner("lk2", [D.outcome("b", "kept", source: "s1")]),
                owner("lk3", [D.outcome("c", "kept")], place: leeds, "Leeds"),
            ],
            shownOwnerID: "pl1", kind: .birthPlace, locale: en
        )
        guard case .value(let sharedHeading)? = shared.provenanceContent().groups[0].values[0].heading else {
            Issue.record("expected a value heading")
            return
        }
        #expect(sharedHeading.support == L10n.Conclusions.sourceCount(1))
    }

    @Test func theSummarySaysOnlyWhatTheRowKnows() {
        let merged = D.field("merged", [
            D.value(1, "kept", support: 3, against: 1, .name(D.jamesRobins)),
        ], outcomes: (0..<4).map { D.outcome("r\($0)", "kept", source: "s\($0 % 3)") })
        #expect(ReconciledValueRowModel(field: merged, locale: en).provenanceContent().summary
            == "merged from 3 Sources · 1 record disagrees · 4 records")

        let single = D.field("single", [
            D.value(1, "kept", support: 1, .name(D.jamesRobins)),
        ], outcomes: [D.outcome("r0", "kept", source: "s0")])
        #expect(ReconciledValueRowModel(field: single, locale: en).provenanceContent().summary?.hasPrefix("from 1 Source") == true)

        let first = CatalogCanonicalEntity(id: "ev1", ref: "EVT-1", subjectTypeID: "t", label: "")
        let second = CatalogCanonicalEntity(id: "ev2", ref: "EVT-2", subjectTypeID: "t", label: "")
        func owner(_ entity: CatalogCanonicalEntity, _ ids: [String]) -> CatalogOwnerField {
            CatalogOwnerField(entity: entity, event: nil, field: D.field(
                "single", key: "date", label: "date", valueType: "date", [], outcomes: ids.map { D.outcome($0, "kept") }
            ))
        }
        let two = ReconciledValueRowModel.derived(
            id: "life.birthDate", label: "Birth date", face: .date, lead: "1822", emptyText: "", mixed: true,
            owners: [owner(first, ["a", "b", "c"]), owner(second, ["d"])], shownOwnerID: "ev1",
            kind: .birthDate, locale: en
        )
        #expect(two.provenanceContent().summary == "4 records from 2 birth events")
        let one = ReconciledValueRowModel.derived(
            id: "life.birthDate", label: "Birth date", face: .date, lead: "1822", emptyText: "", mixed: false,
            owners: [owner(first, ["a", "b"])], shownOwnerID: "ev1", kind: .birthDate, locale: en
        )
        #expect(one.provenanceContent().summary == "2 records")
        #expect(ReconciledValueRowModel.empty(id: "x", label: "X", face: .text, emptyText: "").provenanceContent().summary == nil)
    }
}
