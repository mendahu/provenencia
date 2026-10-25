import AppKit
import SwiftUI
import Testing
@testable import Provenencia

/// Hit targets and edge frames come from each card's painted layout, so they
/// stay on the paint when titles, descriptions, or values wrap.
@Suite
@MainActor
struct EvidenceCardLayoutTests {
    private func observation(_ i: Int, value: String = "Value") -> CatalogObservation {
        CatalogObservation(
            id: "o\(i)", ref: "OBS-\(i)", citationID: "c", subjectID: "s", propertyID: "p\(i)",
            polarity: "positive", valueText: "\(value) \(i)", valueInteger: nil, valueDateID: "", valueNameID: "",
            valueSubjectID: "", valueTermID: "", propertyKey: "k\(i)", propertyLabel: "Prop \(i)", propertyValueType: "text"
        )
    }

    private func placed(
        rows: Int,
        label: String = "James Robins",
        description: String = "",
        value: String = "Value",
        promoted: Bool = false,
        handleName: String? = nil
    ) -> EvidenceGraphPlacedSubject {
        EvidenceGraphPlacedSubject(
            subject: CatalogSubject(id: "s", ref: "CPR-2AB91", sourceID: "x", subjectTypeID: "t", label: label, description: description),
            kind: .person,
            typeLabel: "Person",
            gridX: 2,
            gridY: 3,
            isCited: rows > 0,
            observations: (0..<rows).map { observation($0, value: value) },
            membership: promoted ? CatalogSubjectMembership(
                subjectID: "s", claimID: "c",
                entity: CatalogCanonicalEntity(id: "e", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
                kind: "person",
                name: handleName.map { CatalogNameValue(form: $0) }
            ) : nil
        )
    }

    /// Renders `view` at scale 1 and returns its bitmap.
    private func bitmap(of view: EvidenceCardView) throws -> NSBitmapImageRep {
        NSBitmapImageRep(cgImage: try #require(view.renderImage(scale: 1)))
    }

    /// Builds the live card and returns the layout it measured plus the height
    /// actually drawn: card-body rows (alpha > 0.5; uncited cards paint at 0.76,
    /// shadows stay under ~0.1) down the centre column. (Not
    /// `fittingSize` — AppKit's ideal size over-reports wrapped text by ~4pt.)
    private func measure(_ card: EvidenceCanvasCard) throws -> (EvidenceCardLayout, CGFloat) {
        let view = EvidenceCardView(card: card, actions: [])
        let rep = try bitmap(of: view)
        let x = rep.pixelsWide / 2
        var top = -1
        var bottom = -1
        for y in 0..<rep.pixelsHigh where (rep.colorAt(x: x, y: y)?.alphaComponent ?? 0) > 0.5 {
            if top < 0 { top = y }
            bottom = y
        }
        let scale = CGFloat(rep.pixelsHigh) / view.bounds.height
        return (view.layout, CGFloat(bottom - top + 1) / scale)
    }

    private func measureSubject(_ p: EvidenceGraphPlacedSubject) throws -> (EvidenceCardLayout, CGFloat) {
        try measure(EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting))
    }

    static let cases: [String] = ["uncited", "uncited+description", "cited1", "cited3", "cited3+promoted", "longHandleName+promoted",
                                  "longTitle", "wrappedValues", "longDescription"]

    private func fixture(_ name: String) -> EvidenceGraphPlacedSubject {
        switch name {
        case "uncited": placed(rows: 0)
        case "uncited+description": placed(rows: 0, description: "Son, age 14, line 13")
        case "cited1": placed(rows: 1)
        case "cited3": placed(rows: 3)
        case "cited3+promoted": placed(rows: 3, promoted: true)
        case "longHandleName+promoted":
            placed(rows: 1, promoted: true, handleName: "James Alexander Robins of Wellington North Township")
        case "longTitle": placed(rows: 1, label: "James Alexander Robins of Wellington North Township")
        case "wrappedValues": placed(rows: 2, value: "Agricultural labourer and part-time blacksmith of the township")
        default: placed(rows: 1, description: "Son, age 14, line 13, listed under his father with a note in the margin about schooling")
        }
    }

    @Test(arguments: cases)
    func measuredLayoutMatchesThePaint(name: String) throws {
        let card = fixture(name)
        let (layout, drawn) = try measureSubject(card)
        #expect(abs(layout.size.height - drawn) < 0.5)
        #expect(EvidenceSubjectCard.edgeFrame(for: card, layout: layout).height == layout.size.height)

        // Footer: 36pt, flush with the painted bottom, and its hit band is exactly that.
        let footerID = card.membership == nil ? EvidenceSubjectCard.promoteActionID : EvidenceSubjectCard.openHandleActionID
        if card.membership != nil {
            // The handle row and the Unpromote square are the footer's two painted pieces.
            let slot = try #require(layout.regions[EvidenceSubjectCard.unpromoteActionID])
            #expect(abs(slot.width - 36) < 0.5 && abs(slot.height - 36) < 0.5)
            #expect(abs(slot.maxX - layout.size.width) < 0.5)
            #expect(abs(slot.maxY - layout.size.height) < 0.5)
            let handle = try #require(layout.regions[EvidenceSubjectCard.openHandleActionID])
            let parts = EvidenceSubjectCard.promotedFooterParts(
                footer: CGRect(x: 0, y: slot.minY, width: layout.size.width, height: EvidenceSubjectCard.footerHeight)
            )
            #expect(abs(handle.minX - parts.handle.minX) < 0.5 && abs(handle.width - parts.handle.width) < 0.5)
            #expect(handle.maxX <= slot.minX)
        }
        let footer = try #require(layout.regions[footerID])
        #expect(abs(footer.height - EvidenceSubjectCard.footerHeight) < 0.5)
        #expect(abs(footer.maxY - layout.size.height) < 0.5)

        let frame = EvidenceSubjectCard.edgeFrame(for: card, layout: layout)
        let targets = EvidenceSubjectCard.actionTargets(for: card, canCite: true, layout: layout)
        let footerHit = try #require(targets.first { $0.id == footerID })
        #expect(abs(footerHit.frame.maxY - frame.maxY) < 0.5)
        #expect(abs(footerHit.frame.height - EvidenceSubjectCard.footerHeight) < 0.5)
        if card.membership != nil {
            #expect(targets.contains { $0.id == EvidenceSubjectCard.unpromoteActionID })
        }

        // Every property row has a hit band on its painted row; bands never overlap.
        for observation in card.observations {
            let id = EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)
            let painted = try #require(layout.regions[id])
            let hit = try #require(targets.first { $0.id == id })
            #expect(abs(hit.frame.minY - (frame.minY + painted.minY)) < 0.5)
            #expect(abs(hit.frame.height - painted.height) < 0.5)
        }
        let bands = targets.filter { $0.id != EvidenceSubjectCard.editActionID && $0.id != EvidenceSubjectCard.deleteActionID }
        for (i, a) in bands.enumerated() {
            for b in bands.dropFirst(i + 1) {
                #expect(!a.frame.intersects(b.frame), "\(a.id) overlaps \(b.id)")
            }
        }
    }

    @Test func bridgeTargetsFollowThePaint() throws {
        let bridge = EvidenceGraphPlacedBridge(
            subject: CatalogSubject(id: "b", ref: "CPA-3RN7K", sourceID: "x", subjectTypeID: "t", label: "", description: ""),
            kind: .participation,
            typeLabel: "Participation",
            gridX: 4,
            gridY: 4,
            isCited: true,
            observations: [observation(0), observation(1)]
        )
        let snapshot = EvidenceGraphSnapshot(graphId: "x", bridges: [bridge])
        let (layout, drawn) = try measure(
            EvidenceCanvasCard(placed: .bridge(bridge), text: EvidenceCardText(bridge: bridge, in: snapshot), state: .resting)
        )
        #expect(abs(layout.size.height - drawn) < 0.5)
        let frame = EvidenceBridgeCard.contentFrame(for: bridge, in: snapshot, layout: layout)
        #expect(frame.height == layout.size.height)
        let targets = EvidenceBridgeCard.actionTargets(for: bridge, in: snapshot, canCite: true, layout: layout)
        let add = try #require(targets.first { $0.id == EvidenceSubjectCard.addPropertyActionID })
        let addPaint = try #require(layout.regions[EvidenceSubjectCard.addPropertyActionID])
        #expect(abs(add.frame.minY - (frame.minY + addPaint.minY)) < 0.5)
        #expect(abs(add.frame.maxY - frame.maxY) < 0.5)
    }

    @Test func likenessRowIsTallEnoughForItsThumbnail() throws {
        var likeness = observation(0, value: "")
        likeness.valueText = ""
        likeness.propertyLabel = "Likeness"
        likeness.depiction = CatalogDepictionValue(regionJSON: "{}", imageRelPath: "objects/aa/bb/face.png")
        var p = placed(rows: 0)
        p.isCited = true
        p.observations = [likeness]
        let (layout, drawn) = try measureSubject(p)
        let rect = try #require(layout.regions[EvidenceSubjectCard.editPropertyActionID(observationID: likeness.id)])
        #expect(rect.height >= 9 + 3 + 32 + 16)
        #expect(rect.height > EvidenceSubjectCard.propertyRowHeight)
        #expect(abs(layout.size.height - drawn) < 0.5)
        let targets = EvidenceSubjectCard.actionTargets(for: p, canCite: true, layout: layout)
        let hit = try #require(targets.first { $0.id == EvidenceSubjectCard.editPropertyActionID(observationID: likeness.id) })
        let frame = EvidenceSubjectCard.edgeFrame(for: p, layout: layout)
        #expect(abs(hit.frame.height - rect.height) < 0.5)
        #expect(abs(hit.frame.minY - (frame.minY + rect.minY)) < 0.5)
    }

    @Test func ghostCardMatchesTheLiveLayout() {
        let p = placed(rows: 0)
        let ghost = EvidenceCardView(ghostOf: .ghost(p, presentation: nil))
        let live = EvidenceCardView(
            card: EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting),
            actions: []
        )
        #expect(ghost.layout == live.layout)
    }

    @Test func ghostPaintsAtGhostOpacity() throws {
        /// The card fill's alpha: the most common alpha down the centre column
        /// among painted pixels (text over the fill and shadow edges are rarer).
        func bodyAlpha(_ view: EvidenceCardView) throws -> CGFloat {
            let rep = try bitmap(of: view)
            let x = rep.pixelsWide / 2
            var counts: [Int: Int] = [:]
            for y in 0..<rep.pixelsHigh {
                let a = rep.colorAt(x: x, y: y)?.alphaComponent ?? 0
                if a > 0.3 { counts[Int((a * 255).rounded()), default: 0] += 1 }
            }
            return CGFloat(counts.max { $0.value < $1.value }?.key ?? 0) / 255
        }
        let p = placed(rows: 0)
        let ghost = try bodyAlpha(EvidenceCardView(ghostOf: .ghost(p, presentation: nil)))
        let live = try bodyAlpha(EvidenceCardView(
            card: EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting),
            actions: []
        ))
        #expect(live > 0)
        // Fill, shadow and text composite, so the ratio is only near 0.62 / 0.76; a ghost
        // painted at the live card's 0.76 gives 1.0.
        #expect(abs(ghost / live - 0.62 / 0.76) < 0.06)
    }

    @Test func ghostCardIsNotAnAccessibilityElement() {
        let p = placed(rows: 0)
        let ghost = EvidenceCardView(ghostOf: .ghost(p, presentation: nil))
        #expect(!ghost.isAccessibilityElement())
        #expect((ghost.accessibilityLabel() ?? "").isEmpty)
        #expect(ghost.accessibilityIdentifier().isEmpty)
        #expect((ghost.accessibilityCustomActions() ?? []).isEmpty)

        let live = EvidenceCardView(
            card: EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting),
            actions: []
        )
        #expect(live.isAccessibilityElement())
        #expect(live.accessibilityIdentifier() == "evidenceGraph.subject.s")
    }
}
