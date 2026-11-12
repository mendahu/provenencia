import Foundation
import Testing
@testable import Provenencia

struct PersonSourcesContentTests {
    private let census = CatalogSourceType(
        id: "type-census", key: "census", origin: "provenencia", label: "Census", description: "", iconKey: "type_census"
    )

    private func subject(_ id: String, ref: String, label: String = "") -> CatalogSubject {
        CatalogSubject(id: id, ref: ref, sourceID: "src", subjectTypeID: "type-person", label: label, description: "")
    }

    private func nameObservation(_ subjectID: String, _ form: String) -> CatalogObservation {
        CatalogObservation(
            id: "obs-\(subjectID)", ref: "OBS-\(subjectID)", citationID: "cit", subjectID: subjectID,
            propertyID: "prop-name", polarity: "positive", valueText: form, valueInteger: nil, valueDateID: "",
            valueNameID: "", nameForm: form, valueSubjectID: "", valueTermID: "", propertyKey: "name",
            propertyLabel: "Name", propertyValueType: "name"
        )
    }

    private func claim(
        _ id: String, graph: String, graphRef: String = "", graphName: String = "", subject: CatalogSubject,
        observations: [CatalogObservation] = []
    ) -> CatalogPersonClaim {
        CatalogPersonClaim(
            claimID: id, graphID: graph, graphRef: graphRef, graphName: graphName,
            subject: subject, observations: observations
        )
    }

    private var groups: [CatalogPersonClaimGroup] {
        [
            CatalogPersonClaimGroup(
                sourceID: "src-1", sourceRef: "SRC-9TQ2M", title: "Hamburg census 1867", sourceTypeID: "type-census",
                claims: [
                    claim("c1", graph: "g-214", graphRef: "EVG-1", graphName: "Household 214", subject: subject("s1", ref: "CPR-7Q2KD", label: "jw")),
                    claim("c2", graph: "g-388", graphRef: "EVG-5MD8R", graphName: "Household 388", subject: subject("s2", ref: "CPR-2M8XF", label: "jw")),
                ]
            ),
            CatalogPersonClaimGroup(
                sourceID: "src-2", sourceRef: "SRC-2", title: "Hamburger Adressbuch 1861", sourceTypeID: "type-census",
                claims: [claim("c3", graph: "g-a", graphRef: "EVG-3", graphName: "Page 4", subject: subject("s3", ref: "CPR-3"))]
            ),
            CatalogPersonClaimGroup(
                sourceID: "src-3", sourceRef: "SRC-4KQ7M", title: "Letter", sourceTypeID: "gone",
                claims: [claim("c4", graph: "g-l", graphRef: "", graphName: "Page 9", subject: subject("s4", ref: "CPR-4"))]
            ),
        ]
    }

    /// `pages` pages revealed of each fold in `revealed`.
    private func content(
        _ groups: [CatalogPersonClaimGroup], revealed: [PersonSourcesContent.Fold] = [], pages: Int = 1
    ) -> PersonSourcesContent {
        var folds = ConclusionFoldState<PersonSourcesContent.Fold>()
        for fold in revealed { folds.reveal(fold, pages: pages) }
        return PersonSourcesContent(groups: groups, sourceTypes: [census], orders: NameFormatOrders(), folds: folds)
    }

    @Test func rowsKeepTheReadsOrder() {
        let result = content(groups)
        #expect(result.rows.map(\.id) == ["src-1", "src-2", "src-3"])
        #expect(result.rows[0].lines.map(\.graphTitle) == ["Household 214", "Household 388"])
        #expect(result.rows[2].title == .name("Letter"))
        #expect(!result.isEmpty)
        #expect(content([]).isEmpty)
    }

    @Test func cellsAndFallbacks() {
        let rows = content(groups).rows
        let first = rows[0]
        #expect(first.typeLabel == "Census" && first.typeIconKey == "type_census")
        #expect(first.title == .name("Hamburg census 1867"))
        #expect(first.location == SourcesListNavigation.pageLocation(sourceId: "src-1", ref: "SRC-9TQ2M", title: "Hamburg census 1867"))
        #expect(first.accessibilityLabel == "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M")
        #expect(
            first.lines[0].unpromoteAccessibilityLabel
                == "Unpromote Household 214, record jw, CPR-7Q2KD"
        )
        #expect(first.lines[1].graphRef == "EVG-5MD8R")
        #expect(first.lines[1].graphID == "g-388")
        #expect(first.lines[1].location == SourcesListNavigation.canvasLocation(
            sourceId: "src-1", sourceRef: "SRC-9TQ2M", sourceTitle: "Hamburg census 1867", graphId: "g-388", subjectId: "s2"
        ))

        // An unknown type id: no type cell, and the untyped spoken label.
        let letter = rows[2]
        #expect(letter.typeLabel == nil && letter.typeIconKey == nil)
        #expect(letter.accessibilityLabel == "Letter, date not recorded, SRC-4KQ7M")

        // A graph with no ref, and one with a ref.
        let noRef = content([CatalogPersonClaimGroup(
            sourceID: "s", sourceRef: "SRC-1", title: "T", sourceTypeID: "type-census",
            claims: [
                claim("a", graph: "g1", graphRef: "", graphName: "Named", subject: subject("x", ref: "CPR-1")),
                claim("b", graph: "g2", graphRef: "EVG-9", graphName: "Other", subject: subject("y", ref: "CPR-2")),
            ]
        )]).rows[0].lines
        #expect(noRef[0].graphTitle == "Named" && noRef[0].graphRef == nil)
        #expect(noRef[1].graphTitle == "Other" && noRef[1].graphRef == "EVG-9")
    }

    @Test func foldPastSixLines() {
        func group(_ count: Int) -> [CatalogPersonClaimGroup] {
            [CatalogPersonClaimGroup(
                sourceID: "src", sourceRef: "SRC-1", title: "Adressbuch", sourceTypeID: "type-census",
                claims: (1...count).map { claim("c\($0)", graph: "g\($0)", graphName: "G\($0)", subject: subject("s\($0)", ref: "CPR-\($0)")) }
            )]
        }
        let lines = PersonSourcesContent.Fold.lines(sourceID: "src")
        let nine = content(group(9))
        #expect(nine.rows[0].lines.count == 5 && nine.rows[0].hiddenCount == 4 && nine.firstRevealed(by: lines) == "c6")
        let six = content(group(6))
        #expect(six.rows[0].lines.count == 6 && six.rows[0].hiddenCount == 0 && six.firstRevealed(by: lines) == nil)
        let seven = content(group(7)).rows[0]
        #expect(seven.lines.count == 5 && seven.hiddenCount == 2)
        let open = content(group(9), revealed: [lines])
        #expect(open.rows[0].lines.count == 9 && open.rows[0].hiddenCount == 0 && open.firstRevealed(by: lines) == nil)
    }

    @Test func lineAfterRemoving() {
        let page = content(groups)
        // A middle line: the next one.
        #expect(page.lineAfterRemoving(subjectID: "s1") == "c2")
        // A Source's last line: the first line of the next Source.
        #expect(page.lineAfterRemoving(subjectID: "s2") == "c3")
        // The page's last line: the previous one.
        #expect(page.lineAfterRemoving(subjectID: "s4") == "c3")
        // The only line of the only Source: nothing.
        let only = content([groups[2]])
        #expect(only.lineAfterRemoving(subjectID: "s4") == nil)
        #expect(page.lineAfterRemoving(subjectID: "missing") == nil)
        // A line folded away is never returned.
        let folded = content([CatalogPersonClaimGroup(
            sourceID: "src", sourceRef: "SRC-1", title: "Adressbuch", sourceTypeID: "type-census",
            claims: (1...9).map { claim("c\($0)", graph: "g\($0)", graphName: "G\($0)", subject: subject("s\($0)", ref: "CPR-\($0)")) }
        )])
        #expect(folded.lineAfterRemoving(subjectID: "s5") == "c4")
        #expect(folded.lineAfterRemoving(subjectID: "s4") == "c5")
    }

    @Test func foldPastFiftySources() {
        func many(_ count: Int) -> [CatalogPersonClaimGroup] {
            (1...count).map {
                CatalogPersonClaimGroup(
                    sourceID: "src-\($0)", sourceRef: "SRC-\($0)", title: "Source \($0)", sourceTypeID: "type-census",
                    claims: [claim("c\($0)", graph: "g\($0)", graphName: "G\($0)", subject: subject("s\($0)", ref: "CPR-\($0)"))]
                )
            }
        }
        let sixty = content(many(60))
        #expect(sixty.rows.count == 50 && sixty.hiddenRowCount == 10 && sixty.firstRevealed(by: .rows) == "c51")
        let fiftyOne = content(many(51))
        #expect(fiftyOne.rows.count == 50 && fiftyOne.hiddenRowCount == 1)
        let fifty = content(many(50))
        #expect(fifty.rows.count == 50 && fifty.hiddenRowCount == 0 && fifty.firstRevealed(by: .rows) == nil)
        let open = content(many(60), revealed: [.rows])
        #expect(open.rows.count == 60 && open.hiddenRowCount == 0)
    }

    private func manyLines(_ count: Int, graph: (Int) -> String = { "g\($0)" }) -> [CatalogPersonClaimGroup] {
        [CatalogPersonClaimGroup(
            sourceID: "src", sourceRef: "SRC-1", title: "Adressbuch", sourceTypeID: "type-census",
            claims: (1...count).map { claim("c\($0)", graph: graph($0), graphName: "G\($0)", subject: subject("s\($0)", ref: "CPR-\($0)")) }
        )]
    }

    private func manySources(_ count: Int) -> [CatalogPersonClaimGroup] {
        (1...count).map {
            CatalogPersonClaimGroup(
                sourceID: "src-\($0)", sourceRef: "SRC-\($0)", title: "Source \($0)", sourceTypeID: "type-census",
                claims: [claim("c\($0)", graph: "g\($0)", graphName: "G\($0)", subject: subject("s\($0)", ref: "CPR-\($0)"))]
            )
        }
    }

    @Test func linesAndRowsPage() {
        let lines = PersonSourcesContent.Fold.lines(sourceID: "src")
        let counts = (0...3).map { content(manyLines(120), revealed: [lines], pages: $0).rows[0].lines.count }
        #expect(counts == [5, 55, 105, 120])
        let hidden = (0...3).map { content(manyLines(120), revealed: [lines], pages: $0).rows[0].hiddenCount }
        #expect(hidden == [115, 65, 15, 0])
        #expect(content(manyLines(120)).firstRevealed(by: lines) == "c6")
        #expect(content(manyLines(120), revealed: [lines], pages: 1).firstRevealed(by: lines) == "c56")
        #expect(content(manyLines(120), revealed: [lines], pages: 3).firstRevealed(by: lines) == nil)
        // 56 lines: one press shows them all.
        #expect(content(manyLines(56), revealed: [lines]).rows[0].hiddenCount == 0)

        #expect([0, 1].map { content(manySources(60), revealed: [.rows], pages: $0).rows.count } == [50, 60])
        #expect(content(manySources(60)).firstRevealed(by: .rows) == "c51")
        #expect(content(manySources(60), revealed: [.rows]).firstRevealed(by: .rows) == nil)
    }

    @MainActor @Test func foldedGraphsTakeTheHighlight() {
        let lines = PersonSourcesContent.Fold.lines(sourceID: "src")
        let nine = content(manyLines(9))
        #expect(nine.foldedGraphIDs(by: lines) == ["g6", "g7", "g8", "g9"])
        // A graph with a shown line and a folded one keeps its shown line, not the Show more.
        let shared = content(manyLines(9, graph: { $0 == 7 ? "g3" : "g\($0)" }))
        #expect(shared.foldedGraphIDs(by: lines) == ["g6", "g8", "g9"])
        #expect(content(manyLines(9), revealed: [lines]).foldedGraphIDs(by: lines).isEmpty)

        let sixty = content(manySources(60))
        #expect(sixty.foldedGraphIDs(by: .rows) == Set((51...60).map { "g\($0)" }))
        #expect(!sixty.foldedGraphIDs(by: .rows).contains("g1") && !sixty.foldedGraphIDs(by: .rows).contains("unknown"))
        #expect(content(manySources(60), revealed: [.rows]).foldedGraphIDs(by: .rows).isEmpty)

        let hover = CircularLinkHover<String>()
        hover.id = "g1"
        #expect(hover.highlighted(focused: "g2") == "g1")
        hover.id = nil
        #expect(hover.highlighted(focused: "g2") == "g2")
        #expect(hover.highlighted(focused: nil) == nil)
    }

    @Test func lineNamesTheRecordAsItsCard() throws {
        let named = subject("s1", ref: "CPR-1", label: "jw")
        let observation = nameObservation("s1", "J. Weber")
        let line = content([CatalogPersonClaimGroup(
            sourceID: "s", sourceRef: "SRC-1", title: "T", sourceTypeID: "",
            claims: [
                claim("a", graph: "g", graphName: "G", subject: named, observations: [observation]),
                claim("b", graph: "g", graphName: "G", subject: subject("s2", ref: "CPR-2", label: "label only")),
                claim("c", graph: "g", graphName: "G", subject: subject("s3", ref: "CPR-3")),
            ]
        )]).rows[0].lines
        #expect(line.map(\.subjectName) == ["J. Weber", "label only", "CPR-3"])

        let orders = NameFormatOrders()
        let snapshot = EvidenceGraphSnapshot.build(
            graphId: "g",
            subjects: [named],
            positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 0, gridY: 0)],
            types: [CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person", description: "",
                refPrefix: "PER", candidateRefPrefix: "CPR"
            )],
            observations: [observation],
            rules: []
        )
        let placed = try #require(snapshot.subjects.first)
        #expect(placed.displayName(orders: orders) == line[0].subjectName)
    }

    @Test func graphLocationsAndClaimGraphsCoverFoldedLines() throws {
        let folded = PersonSourcesPreviewData.folded
        let all = content(folded)
        #expect(all.rows[0].lines.count == 5)
        #expect(all.graphLocations.count == 7)
        #expect(all.claimGraphs.count == 7)
        for claim in folded[0].claims {
            #expect(all.claimGraphs[claim.claimID] == claim.graphID)
            let location = try #require(all.graphLocations[claim.graphID])
            #expect(location == SourcesListNavigation.canvasLocation(
                sourceId: "src-4", sourceRef: "SRC-6P3VD", sourceTitle: "Kirchenbuch St. Michaelis",
                graphId: claim.graphID, subjectId: claim.subject.id
            ))
        }
        let regular = content(PersonSourcesPreviewData.regular)
        #expect(regular.graphLocations["g-214"] == regular.rows[0].lines[0].location)
        #expect(Set(regular.graphLocations.keys) == ["g-214", "g-388", "g-a"])
    }
}
