import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasEdgeGeometryTests {
    @Test func attachmentPointHitsHorizontalEdge() {
        let rect = CGRect(x: 0, y: 0, width: 100, height: 40)
        let point = GraphCanvasEdgeGeometry.attachmentPoint(
            fromRect: rect,
            toward: CGPoint(x: 200, y: 20)
        )
        #expect(abs(point.x - 100) < 0.01)
        #expect(abs(point.y - 20) < 0.01)
    }

    @Test func attachmentPointHitsVerticalEdge() {
        let rect = CGRect(x: 0, y: 0, width: 80, height: 60)
        let point = GraphCanvasEdgeGeometry.attachmentPoint(
            fromRect: rect,
            toward: CGPoint(x: 40, y: 200)
        )
        #expect(abs(point.x - 40) < 0.01)
        #expect(abs(point.y - 60) < 0.01)
    }

    @Test func segmentEndpointsLieInsideRectsAfterTuck() {
        let a = CGRect(x: 0, y: 0, width: 100, height: 50)
        let b = CGRect(x: 200, y: 100, width: 100, height: 50)
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: a, toRect: b)
        // CGRect.contains is max-exclusive; require a clear inset past the border.
        #expect(seg.start.x > a.minX + 1 && seg.start.x < a.maxX - 1)
        #expect(seg.start.y > a.minY + 1 && seg.start.y < a.maxY - 1)
        #expect(seg.end.x > b.minX + 1 && seg.end.x < b.maxX - 1)
        #expect(seg.end.y > b.minY + 1 && seg.end.y < b.maxY - 1)
        #expect(seg.path.isEmpty == false)
    }

    @Test func bottomEdgeTuckMovesStraightUp() {
        let rect = CGRect(x: 0, y: 0, width: 100, height: 60)
        let onBottom = CGPoint(x: 70, y: 60)
        let tucked = GraphCanvasEdgeGeometry.tuckInside(onBottom, rect: rect, distance: 16)
        #expect(abs(tucked.x - 70) < 0.01)
        #expect(abs(tucked.y - 44) < 0.01)
        #expect(rect.contains(tucked))
    }

    @Test func diagonalApproachStillTucksPastBottomEdge() {
        let a = CGRect(x: 0, y: 0, width: 100, height: 60)
        let b = CGRect(x: 80, y: 200, width: 100, height: 60)
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: a, toRect: b)
        // Start was on/near bottom of A — after tuck it must sit above maxY.
        #expect(seg.start.y < a.maxY - 8)
        #expect(a.contains(seg.start))
    }

    @Test func primaryEdgeFrameSharesLayoutTop() {
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 2,
            gridY: 3,
            isCited: false
        )
        let frame = EvidenceSubjectCard.edgeFrame(for: placed)
        let offset = EvidenceSubjectCard.topLeadingOffset(gridX: 2, gridY: 3)
        #expect(abs(frame.minX - offset.width) < 0.01)
        #expect(abs(frame.minY - offset.height) < 0.01)
        #expect(frame.height >= EvidenceSubjectCard.edgeLayoutHeight)
        #expect(frame.width == EvidenceSubjectCard.width)
    }

    @Test func uncitedCardExposesDeleteTargetWithoutPropertyEdit() {
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: false
        )
        let actions = EvidenceSubjectCard.actionTargets(for: placed, canCite: true)
        #expect(actions.contains(where: { $0.id == EvidenceSubjectCard.deleteActionID }))
        #expect(actions.contains(where: { $0.id == EvidenceSubjectCard.editActionID }))
        #expect(!actions.contains(where: { $0.id.hasPrefix("editProperty.") }))
    }

    @Test func citedCardExposesPropertyEditAndHidesDelete() {
        let observation = CatalogObservation(
            id: "obs-1",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "s1",
            propertyID: "p1",
            polarity: "positive",
            valueText: "Farmer",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "occupation",
            propertyLabel: "Occupation",
            propertyValueType: "text"
        )
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [observation]
        )
        let actions = EvidenceSubjectCard.actionTargets(for: placed, canCite: true)
        #expect(!actions.contains(where: { $0.id == EvidenceSubjectCard.deleteActionID }))
        #expect(actions.contains(where: {
            $0.id == EvidenceSubjectCard.editPropertyActionID(observationID: "obs-1")
        }))
        #expect(EvidenceSubjectCard.width == 264)
        #expect(EvidenceSubjectCard.contentHeight(for: placed) > EvidenceSubjectCard.edgeLayoutHeight)
    }

    @Test func headerActionHitsCenterOnTrailingIcons() throws {
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: false
        )
        let frame = EvidenceSubjectCard.edgeFrame(for: placed)
        let actions = EvidenceSubjectCard.actionTargets(for: placed, canCite: true)
        let edit = try #require(actions.first { $0.id == EvidenceSubjectCard.editActionID })
        let delete = try #require(actions.first { $0.id == EvidenceSubjectCard.deleteActionID })

        let trailing = frame.maxX - EvidenceSubjectCard.shellPaddingX
        let iconSize = EvidenceCardHeaderActionHits.iconSize
        let spacing = EvidenceCardHeaderActionHits.iconSpacing
        let deleteIconMidX = trailing - iconSize / 2
        let editIconMidX = trailing - iconSize - spacing - iconSize / 2

        #expect(abs(delete.frame.midX - deleteIconMidX) < 0.5)
        #expect(abs(edit.frame.midX - editIconMidX) < 0.5)
        #expect(delete.frame.maxX > edit.frame.maxX)
    }

    @Test func citedBridgeCitationEditSitsBelowSubjectEdit() throws {
        let observation = CatalogObservation(
            id: "obs-1",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "b1",
            propertyID: "p1",
            polarity: "positive",
            valueText: "Alice",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "s1",
            valueTermID: "",
            propertyKey: "person",
            propertyLabel: "Person",
            propertyValueType: "subject"
        )
        let placed = SourceGraphPlacedBridge(
            subject: CatalogSubject(
                id: "b1",
                ref: "CPA-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "Working",
                description: ""
            ),
            kind: .participation,
            typeLabel: "Participation",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [observation]
        )
        let frame = EvidenceBridgeCard.contentFrame(for: placed)
        let actions = EvidenceBridgeCard.actionTargets(for: placed, canCite: true)
        let edit = try #require(actions.first { $0.id == EvidenceBridgeCard.editActionID })
        let citation = try #require(actions.first { $0.id == EvidenceBridgeCard.editCitationActionID })
        #expect(!actions.contains(where: { $0.id == EvidenceBridgeCard.deleteActionID }))

        #expect(abs(citation.frame.midX - edit.frame.midX) < 0.5)
        #expect(citation.frame.minY > edit.frame.maxY)
    }

    @Test func wrappedBridgePhraseGrowsTheHitFrame() {
        func placed(text: String) -> SourceGraphPlacedBridge {
            SourceGraphPlacedBridge(
                subject: CatalogSubject(
                    id: "b1",
                    ref: "CPA-1",
                    sourceID: "src",
                    subjectTypeID: "t",
                    label: "Working",
                    description: ""
                ),
                kind: .participation,
                typeLabel: "Participation",
                gridX: 0,
                gridY: 0,
                isCited: true,
                observations: [
                    CatalogObservation(
                        id: "obs-role",
                        ref: "OBS-R",
                        citationID: "cit-1",
                        subjectID: "b1",
                        propertyID: "p-role",
                        polarity: "positive",
                        valueText: text,
                        valueInteger: nil,
                        valueDateID: "",
                        valueNameID: "",
                        valueSubjectID: "",
                        valueTermID: "",
                        propertyKey: "role",
                        propertyLabel: "Role",
                        propertyValueType: "text"
                    ),
                ]
            )
        }
        let short = EvidenceBridgeCard.contentHeight(for: placed(text: "Witness"))
        let long = EvidenceBridgeCard.contentHeight(
            for: placed(text: String(repeating: "great-grandparent ", count: 24))
        )
        #expect(long > short)
        #expect(long > EvidenceBridgeCard.edgeLayoutHeight)
        #expect(EvidenceBridgeCard.contentFrame(for: placed(text: String(repeating: "great-grandparent ", count: 24))).height == long)
    }

    @Test func bridgeOmitsCitationEditWhenCannotCite() {
        let placed = SourceGraphPlacedBridge(
            subject: CatalogSubject(
                id: "b1",
                ref: "CPA-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "Working",
                description: ""
            ),
            kind: .participation,
            typeLabel: "Participation",
            gridX: 0,
            gridY: 0,
            isCited: true
        )
        let actions = EvidenceBridgeCard.actionTargets(for: placed, canCite: false)
        #expect(actions.contains(where: { $0.id == EvidenceBridgeCard.editActionID }))
        #expect(!actions.contains(where: { $0.id == EvidenceBridgeCard.editCitationActionID }))
    }
}
