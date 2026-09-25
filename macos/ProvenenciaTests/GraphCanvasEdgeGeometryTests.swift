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
        #expect(EvidenceBridgeCard.contentFrame(for: placed(text: String(repeating: "great-grandparent ", count: 24))).height == long)
    }

    /// Regression for #183: a frame floor taller than the painted bridge left
    /// bottom-approach terminals visible below the card.
    @Test func bridgeFrameTracksPaintedHeightWithoutFloor() {
        func placed(person: String, event: String) -> SourceGraphPlacedBridge {
            func observation(id: String, key: String, value: String) -> CatalogObservation {
                CatalogObservation(
                    id: id,
                    ref: "OBS-\(id)",
                    citationID: "cit-1",
                    subjectID: "b1",
                    propertyID: "p-\(key)",
                    polarity: "positive",
                    valueText: value,
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "s-\(key)",
                    valueTermID: "",
                    propertyKey: key,
                    propertyLabel: key,
                    propertyValueType: "subject"
                )
            }
            return SourceGraphPlacedBridge(
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
                    observation(id: "person", key: "person", value: person),
                    observation(id: "event", key: "event", value: event),
                ]
            )
        }
        let oneLine = EvidenceBridgeCard.contentHeight(for: placed(person: "Al", event: "X"))
        let wrapped = EvidenceBridgeCard.contentHeight(
            for: placed(person: "Bartholomew", event: "the Baptism")
        )
        let shell = EvidenceBridgeCard.shellPaddingTop * 2
            + EvidenceBridgeCard.headerContentHeightCited
            + EvidenceBridgeCard.headerToBodySpacing
        // Shell plus a single body line — the old 88 pt floor pinned both cases to 88.
        #expect(oneLine > shell)
        #expect(oneLine < shell + 30)
        #expect(wrapped > oneLine)
    }

    @Test func bottomApproachEndsUnderTheBridgeFill() {
        let bridge = SourceGraphPlacedBridge(
            subject: CatalogSubject(
                id: "b1",
                ref: "CPA-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "Working",
                description: ""
            ),
            kind: .location,
            typeLabel: "Location",
            gridX: 4,
            gridY: 2,
            isCited: true
        )
        let bridgeRect = EvidenceBridgeCard.contentFrame(for: bridge)
        // Primary card directly below the bridge, so the edge climbs into its bottom.
        let primaryRect = CGRect(
            x: bridgeRect.midX - EvidenceSubjectCard.width / 2,
            y: bridgeRect.maxY + 160,
            width: EvidenceSubjectCard.width,
            height: EvidenceSubjectCard.edgeLayoutHeight
        )
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: primaryRect, toRect: bridgeRect)
        let paintedBottom = bridgeRect.minY + EvidenceBridgeCard.contentHeight(for: bridge)
        #expect(abs(bridgeRect.maxY - paintedBottom) < 0.01)
        #expect(abs(seg.end.y - (paintedBottom - GraphCanvasEdgeGeometry.endpointTuck)) < 0.01)
        #expect(seg.end.y > bridgeRect.midY)
        #expect(bridgeRect.contains(seg.end))
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

    @Test func citedBridgeCopyUsesTheFullSentence() {
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
            observations: [
                CatalogObservation(
                    id: "obs-person",
                    ref: "OBS-P",
                    citationID: "cit-1",
                    subjectID: "b1",
                    propertyID: "p-person",
                    polarity: "positive",
                    valueText: "Jerry",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "s-person",
                    valueTermID: "",
                    propertyKey: "person",
                    propertyLabel: "Person",
                    propertyValueType: "subject"
                ),
                CatalogObservation(
                    id: "obs-event",
                    ref: "OBS-E",
                    citationID: "cit-1",
                    subjectID: "b1",
                    propertyID: "p-event",
                    polarity: "positive",
                    valueText: "Birth",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "s-event",
                    valueTermID: "",
                    propertyKey: "event",
                    propertyLabel: "Event",
                    propertyValueType: "subject"
                ),
                CatalogObservation(
                    id: "obs-role",
                    ref: "OBS-R",
                    citationID: "cit-1",
                    subjectID: "b1",
                    propertyID: "p-role",
                    polarity: "positive",
                    valueText: "subject",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "",
                    valueTermID: "term-subject",
                    propertyKey: "role",
                    propertyLabel: "Role",
                    propertyValueType: "term"
                ),
            ]
        )
        let sentence = EvidenceBridgeEdgeSummary.sentence(for: placed)
        #expect(sentence == L10n.EvidenceGraph.bridgeSummaryParticipation(
            person: "Jerry",
            role: "subject",
            event: "Birth"
        ))
        #expect(EvidenceBridgeCard.accessibilityLabel(for: placed).contains("Jerry"))
        #expect(EvidenceBridgeCard.accessibilityLabel(for: placed).contains("Birth"))
        #expect(
            EvidenceBridgeCard.contentHeight(for: placed)
                > EvidenceBridgeCard.shellPaddingTop * 2
                + EvidenceBridgeCard.headerContentHeightCited
                + EvidenceBridgeCard.headerToBodySpacing
        )
    }

    @Test func hitRefreshTokenChangesWhenObservationIDIsReplaced() {
        func card(observationID: String) -> SourceGraphPlacedSubject {
            SourceGraphPlacedSubject(
                subject: CatalogSubject(
                    id: "s1",
                    ref: "CPR-1",
                    sourceID: "src",
                    subjectTypeID: "t",
                    label: "Alice",
                    description: ""
                ),
                kind: .person,
                typeLabel: "Person",
                gridX: 0,
                gridY: 0,
                isCited: true,
                observations: [
                    CatalogObservation(
                        id: observationID,
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
                    ),
                ]
            )
        }
        let before = EvidenceGraphHitRefresh.token(subjects: [card(observationID: "obs-old")], bridges: [])
        let after = EvidenceGraphHitRefresh.token(subjects: [card(observationID: "obs-new")], bridges: [])
        #expect(before != after)
    }
}
