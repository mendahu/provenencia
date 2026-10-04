import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

/// S9-D8 footer: one fixed 44pt slot on primary cards — Promote until the
/// subject has an accepted Identity Claim, then the membership row.
@Suite
struct EvidenceSubjectCardFooterTests {
    private let membership = CatalogSubjectMembership(
        subjectID: "s1",
        claimID: "c1",
        entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "type-person", label: ""),
        kind: "person"
    )

    private func placed(cited: Bool, promoted: Bool, kind: EvidencePrimaryKind = .person) -> SourceGraphPlacedSubject {
        let observations: [CatalogObservation] = cited ? [
            CatalogObservation(
                id: "o1", ref: "OBS-1", citationID: "c-1", subjectID: "s1", propertyID: "p1",
                polarity: "positive", valueText: "James Robins", valueInteger: nil,
                valueDateID: "", valueNameID: "", valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "text"
            ),
        ] : []
        return SourceGraphPlacedSubject(
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
