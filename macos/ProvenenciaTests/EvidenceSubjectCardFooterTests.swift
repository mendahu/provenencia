import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

/// S9-D8 footer: one fixed 44pt slot on primary cards — Promote until the
/// subject has an accepted Identity Claim, then the membership row.
@Suite
@MainActor
struct EvidenceSubjectCardFooterTests {
    private let membership = CatalogSubjectMembership(
        subjectID: "s1",
        claimID: "c1",
        entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "type-person", label: ""),
        kind: "person"
    )

    private func placed(cited: Bool, promoted: Bool, kind: EvidencePrimaryKind = .person) -> EvidenceGraphPlacedSubject {
        let observations: [CatalogObservation] = cited ? [
            CatalogObservation(
                id: "o1", ref: "OBS-1", citationID: "c-1", subjectID: "s1", propertyID: "p1",
                polarity: "positive", valueText: "James Robins", valueInteger: nil,
                valueDateID: "", valueNameID: "", valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "text"
            ),
        ] : []
        return EvidenceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1", ref: "CPR-2AB91", sourceID: "src", subjectTypeID: "type-\(kind.rawValue)",
                label: "James Robins", description: ""
            ),
            kind: kind,
            typeLabel: "Person",
            gridX: 3,
            gridY: 3,
            isCited: cited,
            observations: observations,
            membership: promoted ? membership : nil
        )
    }

    @Test(arguments: [true, false])
    func promotingNeverChangesCardHeight(cited: Bool) {
        let before = EvidenceSubjectCard.contentHeight(for: placed(cited: cited, promoted: false))
        let after = EvidenceSubjectCard.contentHeight(for: placed(cited: cited, promoted: true))
        #expect(before == after)
    }

    @Test func citedFooterAddsHairlineAndSlot() {
        var noFooter = EvidenceSubjectCard.shellPaddingTop + EvidenceSubjectCard.headerHeight
        noFooter += EvidenceSubjectCard.headerToBodyGap + EvidenceSubjectCard.stackHairline
            + EvidenceSubjectCard.propertyRowHeight + EvidenceSubjectCard.stackHairline
            + EvidenceSubjectCard.addPropertyInStackHeight
        let height = EvidenceSubjectCard.contentHeight(for: placed(cited: true, promoted: false))
        #expect(height == noFooter + EvidenceSubjectCard.stackHairline + EvidenceSubjectCard.footerHeight)
    }

    @Test(arguments: [true, false])
    func footerTargetIsFullWidthAtTheBottom(cited: Bool) {
        let card = placed(cited: cited, promoted: false)
        let frame = EvidenceSubjectCard.edgeFrame(for: card)
        let targets = EvidenceSubjectCard.actionTargets(for: card, canCite: true)
        let promote = targets.first { $0.id == EvidenceSubjectCard.promoteActionID }
        #expect(EvidenceSubjectCard.footerHeight == 36)
        #expect(promote?.frame == CGRect(
            x: frame.minX, y: frame.maxY - 36, width: EvidenceSubjectCard.width, height: 36
        ))
        #expect(!targets.contains { $0.id == EvidenceSubjectCard.openHandleActionID })
        // Add property stays clear of the footer.
        let add = targets.first { $0.id == EvidenceSubjectCard.addPropertyActionID }
        #expect(add != nil)
        #expect(add.map { !$0.frame.intersects(promote?.frame ?? .zero) } == true)
    }

    @Test func promotedFooterSplitsHandleAndUnpromote() throws {
        let card = placed(cited: true, promoted: true)
        let frame = EvidenceSubjectCard.edgeFrame(for: card)
        let targets = EvidenceSubjectCard.actionTargets(for: card, canCite: true, layout: nil)
        let handle = try #require(targets.first { $0.id == EvidenceSubjectCard.openHandleActionID })
        let unpromote = try #require(targets.first { $0.id == EvidenceSubjectCard.unpromoteActionID })
        #expect(!handle.frame.intersects(unpromote.frame))
        #expect(handle.frame.height == 36 && unpromote.frame.height == 36)
        #expect(handle.frame.maxY == frame.maxY && unpromote.frame.maxY == frame.maxY)
        #expect(unpromote.frame.width == 36)
        #expect(unpromote.frame.maxX == frame.maxX)
        #expect(handle.frame.width + 1 + unpromote.frame.width == EvidenceSubjectCard.width)
        #expect(unpromote.help == "Unpromote")
        #expect(handle.help == nil)

        for other in [placed(cited: true, promoted: false), placed(cited: false, promoted: false)] {
            let ids = EvidenceSubjectCard.actionTargets(for: other, canCite: true).map(\.id)
            #expect(!ids.contains(EvidenceSubjectCard.unpromoteActionID))
        }

        // A filed bridge card has no Unpromote square either.
        let bridge = EvidenceGraphPlacedBridge(
            subject: CatalogSubject(id: "b1", ref: "REL-3RN7K", sourceID: "x", subjectTypeID: "t", label: "", description: ""),
            kind: .relationship,
            typeLabel: "Relationship",
            gridX: 5,
            gridY: 5,
            isCited: true,
            observations: [
                CatalogObservation(
                    id: "o1", ref: "OBS-1", citationID: "c", subjectID: "b1", propertyID: "p1",
                    polarity: "positive", valueText: "Value 1", valueInteger: nil, valueDateID: "", valueNameID: "",
                    valueSubjectID: "", valueTermID: "", propertyKey: "k1", propertyLabel: "Prop 1", propertyValueType: "text"
                ),
            ],
            membership: CatalogSubjectMembership(
                subjectID: "b1", claimID: "cb1",
                entity: CatalogCanonicalEntity(id: "eb1", ref: "REL-7KD45", subjectTypeID: "t", label: ""),
                kind: "relationship"
            )
        )
        let snapshot = EvidenceGraphSnapshot(graphId: "x", bridges: [bridge])
        let bridgeIDs = EvidenceBridgeCard.actionTargets(for: bridge, in: snapshot, canCite: true).map(\.id)
        #expect(!bridgeIDs.contains(EvidenceSubjectCard.unpromoteActionID))
    }

    @Test func membershipSwapsPromoteForOpenHandle() {
        let card = placed(cited: true, promoted: true)
        let ids = EvidenceSubjectCard.actionTargets(for: card, canCite: true).map(\.id)
        #expect(ids.contains(EvidenceSubjectCard.openHandleActionID))
        #expect(!ids.contains(EvidenceSubjectCard.promoteActionID))
    }

    @Test func promoteNeedsNoArtifact() {
        let ids = EvidenceSubjectCard.actionTargets(for: placed(cited: false, promoted: false), canCite: false).map(\.id)
        #expect(ids.contains(EvidenceSubjectCard.promoteActionID))
        #expect(!ids.contains(EvidenceSubjectCard.addPropertyActionID))
    }

    @Test func footerAccessibilityNames() {
        #expect(EvidenceSubjectCard.footerAccessibilityActionName(for: placed(cited: true, promoted: false))
            == L10n.EvidenceGraph.promoteAccessibility(label: "James Robins", ref: "CPR-2AB91"))
        #expect(EvidenceSubjectCard.footerAccessibilityActionName(for: placed(cited: true, promoted: true))
            == L10n.EvidenceGraph.openHandleAccessibility(kind: .person, ref: "PER-7KD45"))
    }
}
