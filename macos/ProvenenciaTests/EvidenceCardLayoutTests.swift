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
        promoted: Bool = false
    ) -> SourceGraphPlacedSubject {
        SourceGraphPlacedSubject(
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
                kind: "person"
            ) : nil
        )
    }

    /// Hosts a card offscreen and returns the layout it reports plus the height
    /// actually drawn: card-body rows (alpha > 0.5; uncited cards paint at 0.76,
    /// shadows stay under ~0.1) down the centre column. (Not
    /// `fittingSize` — AppKit's ideal size over-reports wrapped text by ~4pt.)
    private func measure<V: View>(_ make: (@escaping (EvidenceCardLayout) -> Void) -> V) async throws -> (EvidenceCardLayout, CGFloat) {
        var reported: EvidenceCardLayout?
        let host = NSHostingView(rootView: make { reported = $0 })
        let window = NSWindow(contentRect: CGRect(x: 0, y: 0, width: 400, height: 900), styleMask: [], backing: .buffered, defer: false)
        window.contentView = host
        host.frame = CGRect(x: 0, y: 0, width: 400, height: 700)
        // Settle until the reported layout stops changing.
        var last: EvidenceCardLayout?
        var stablePasses = 0
        let deadline = ContinuousClock.now + .seconds(2)
        while stablePasses < 3, ContinuousClock.now < deadline {
            host.layoutSubtreeIfNeeded()
            try await Task.sleep(for: .milliseconds(5))
            if let reported, reported == last {
                stablePasses += 1
            } else {
                stablePasses = 0
                last = reported
            }
        }
        let rep = try #require(host.bitmapImageRepForCachingDisplay(in: host.bounds))
        host.cacheDisplay(in: host.bounds, to: rep)
        let x = rep.pixelsWide / 2
        var top = -1
        var bottom = -1
        for y in 0..<rep.pixelsHigh where (rep.colorAt(x: x, y: y)?.alphaComponent ?? 0) > 0.5 {
            if top < 0 { top = y }
            bottom = y
        }
        let scale = CGFloat(rep.pixelsHigh) / host.bounds.height
        return (try #require(reported), CGFloat(bottom - top + 1) / scale)
    }

    private func measureSubject(_ p: SourceGraphPlacedSubject) async throws -> (EvidenceCardLayout, CGFloat) {
        try await measure { report in
            EvidenceSubjectCard(placed: p, isSelected: false, isActivated: false, onLayout: report)
        }
    }

    static let cases: [String] = ["uncited", "uncited+description", "cited1", "cited3", "cited3+promoted",
                                  "longTitle", "wrappedValues", "longDescription"]

    private func fixture(_ name: String) -> SourceGraphPlacedSubject {
        switch name {
        case "uncited": placed(rows: 0)
        case "uncited+description": placed(rows: 0, description: "Son, age 14, line 13")
        case "cited1": placed(rows: 1)
        case "cited3": placed(rows: 3)
        case "cited3+promoted": placed(rows: 3, promoted: true)
        case "longTitle": placed(rows: 1, label: "James Alexander Robins of Wellington North Township")
        case "wrappedValues": placed(rows: 2, value: "Agricultural labourer and part-time blacksmith of the township")
        default: placed(rows: 1, description: "Son, age 14, line 13, listed under his father with a note in the margin about schooling")
        }
    }

    @Test(arguments: cases)
    func measuredLayoutMatchesThePaint(name: String) async throws {
        let card = fixture(name)
        let (layout, drawn) = try await measureSubject(card)
        #expect(abs(layout.size.height - drawn) < 0.5)
        #expect(EvidenceSubjectCard.edgeFrame(for: card, layout: layout).height == layout.size.height)

        // Footer: 36pt, flush with the painted bottom, and its hit band is exactly that.
        let footerID = card.membership == nil ? EvidenceSubjectCard.promoteActionID : EvidenceSubjectCard.openHandleActionID
        let footer = try #require(layout.regions[footerID])
        #expect(abs(footer.height - EvidenceSubjectCard.footerHeight) < 0.5)
        #expect(abs(footer.maxY - layout.size.height) < 0.5)

        let frame = EvidenceSubjectCard.edgeFrame(for: card, layout: layout)
        let targets = EvidenceSubjectCard.actionTargets(for: card, canCite: true, layout: layout)
        let footerHit = try #require(targets.first { $0.id == footerID })
        #expect(abs(footerHit.frame.maxY - frame.maxY) < 0.5)
        #expect(abs(footerHit.frame.height - EvidenceSubjectCard.footerHeight) < 0.5)

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

    @Test func bridgeTargetsFollowThePaint() async throws {
        let bridge = SourceGraphPlacedBridge(
            subject: CatalogSubject(id: "b", ref: "CPA-3RN7K", sourceID: "x", subjectTypeID: "t", label: "", description: ""),
            kind: .participation,
            typeLabel: "Participation",
            gridX: 4,
            gridY: 4,
            isCited: true,
            observations: [observation(0), observation(1)]
        )
        let snapshot = SourceGraphSnapshot(sourceId: "x", bridges: [bridge])
        let (layout, drawn) = try await measure { report in
            EvidenceBridgeCard(placed: bridge, snapshot: snapshot, isSelected: false, isActivated: false,
                               canCite: true, onLayout: report)
        }
        #expect(abs(layout.size.height - drawn) < 0.5)
        let frame = EvidenceBridgeCard.contentFrame(for: bridge, in: snapshot, layout: layout)
        #expect(frame.height == layout.size.height)
        let targets = EvidenceBridgeCard.actionTargets(for: bridge, in: snapshot, canCite: true, layout: layout)
        let add = try #require(targets.first { $0.id == EvidenceSubjectCard.addPropertyActionID })
        let addPaint = try #require(layout.regions[EvidenceSubjectCard.addPropertyActionID])
        #expect(abs(add.frame.minY - (frame.minY + addPaint.minY)) < 0.5)
        #expect(abs(add.frame.maxY - frame.maxY) < 0.5)
    }
}
