import AppKit
import SwiftUI

/// Everything one graph card draws, as a value.
///
/// The graph pushes cards to ``GraphCanvasItemLayer`` as ``GraphCanvasItem``s
/// with this as their model, and AppKit rebuilds a card's view only when
/// its model changes. Drag position is not part of it: AppKit moves the view.
struct EvidenceCanvasCard: Equatable {
    enum Placed: Equatable {
        case subject(EvidenceGraphPlacedSubject, CatalogSubjectTypePresentation?)
        case bridge(EvidenceGraphPlacedBridge)
    }

    /// Interaction state that changes the card's paint.
    struct State: Equatable {
        var isSelected: Bool
        var isActivated: Bool
        var isConnectingFrom: Bool
        var isDragging: Bool
        var canCite: Bool
        var hoveredActionID: String?
        var pressedActionID: String?

        /// A card nobody is touching.
        static let resting = State(
            isSelected: false, isActivated: false, isConnectingFrom: false, isDragging: false,
            canCite: true, hoveredActionID: nil, pressedActionID: nil
        )
    }

    /// The placement ghost's card: an unfiled, resting subject card.
    @MainActor
    static func ghost(
        _ placed: EvidenceGraphPlacedSubject,
        presentation: CatalogSubjectTypePresentation?
    ) -> EvidenceCanvasCard {
        EvidenceCanvasCard(
            placed: .subject(placed, presentation),
            text: EvidenceCardText(subject: placed),
            state: .resting
        )
    }

    var placed: Placed
    var text: EvidenceCardText
    var state: State
    /// Default profile orders. Empty keeps stored part order.
    var nameOrders: NameFormatOrders = NameFormatOrders()
    /// The open project's folder, where a likeness row's image lives. Empty
    /// paints the placeholder (the ghost).
    var projectDir: String = ""

    var id: String {
        switch placed {
        case .subject(let subject, _): subject.id
        case .bridge(let bridge): bridge.id
        }
    }

    var isBridge: Bool {
        if case .bridge = placed { return true }
        return false
    }

    /// Document-space top-leading corner of the card at its grid cell.
    @MainActor
    var origin: CGPoint {
        let offset = switch placed {
        case .subject(let subject, _):
            EvidenceSubjectCard.topLeadingOffset(gridX: subject.gridX, gridY: subject.gridY)
        case .bridge(let bridge):
            EvidenceBridgeCard.topLeadingOffset(gridX: bridge.gridX, gridY: bridge.gridY)
        }
        return CGPoint(x: offset.width, y: offset.height)
    }
}

/// Builds a card's view and its VoiceOver actions. Holds references only (the
/// model, navigation, and layout store).
@MainActor
struct EvidenceCanvasCardContent {
    let model: EvidenceGraphModel
    let navigation: WorkspaceNavigation
    let layouts: EvidenceCardLayoutStore

    /// The subject or bridge card's view. Records the measured hit layout
    /// before the canvas sizes the view, so edges and pointer targets exist
    /// for the opening batch.
    func cardView(for card: EvidenceCanvasCard) -> NSView {
        let view = EvidenceCardView(card: card, actions: accessibilityActions(for: card))
        layouts.record(view.layout, for: card.id)
        return view
    }

    private func accessibilityActions(for card: EvidenceCanvasCard) -> [NSAccessibilityCustomAction] {
        namedActions(for: card).map { name, run in
            NSAccessibilityCustomAction(name: name) {
                run()
                return true
            }
        }
    }

    /// The card's VoiceOver actions, in order. The live card maps these to
    /// custom actions; tests read them here.
    func namedActions(for card: EvidenceCanvasCard) -> [(name: String, run: () -> Void)] {
        let model = model
        let navigation = navigation
        func go(_ location: WorkspaceLocation?) {
            if let location { navigation.go(to: location) }
        }
        var named: [(name: String, run: () -> Void)] = []
        switch card.placed {
        case .subject(let placed, _):
            named.append((L10n.string(L10n.EvidenceGraph.editAccessibility), {
                model.beginEdit(subjectID: placed.id)
            }))
            named.append((L10n.string(Self.addPropertyActionName(canCite: card.state.canCite)), {
                go(model.composerLocation(for: placed.id))
            }))
            if let name = card.text.footerActionName {
                named.append((name, {
                    go(placed.membership == nil
                        ? model.promoteLocation(for: placed.id)
                        : model.openHandle(subjectID: placed.id))
                }))
            }
            if placed.membership != nil {
                named.append((L10n.string(L10n.EvidenceGraph.unpromote), {
                    model.beginUnpromote(subjectID: placed.id)
                }))
            }
            named.append((card.text.deleteActionName, {
                Task { await model.beginDelete(subjectID: placed.id) }
            }))
            named.append((L10n.string(L10n.EvidenceGraph.editPropertyAccessibility), {
                if let observation = placed.observations.first {
                    go(model.composerLocation(forObservationID: observation.id, subjectID: placed.id))
                }
            }))
        case .bridge(let placed):
            named.append((L10n.string(L10n.EvidenceGraph.editAccessibility), {
                model.beginEdit(subjectID: placed.id)
            }))
            named.append((L10n.string(L10n.EvidenceGraph.editCitationAccessibility), {
                go(model.composerLocationForBridgeCitation(subjectID: placed.id))
            }))
            named.append((L10n.string(Self.addPropertyActionName(canCite: card.state.canCite)), {
                go(model.composerLocation(for: placed.id))
            }))
            named.append((card.text.deleteActionName, {
                Task { await model.beginDelete(subjectID: placed.id) }
            }))
        }
        return named
    }

    private static func addPropertyActionName(canCite: Bool) -> LocalizedStringResource {
        canCite ? L10n.EvidenceGraph.addProperty : L10n.EvidenceGraph.addPropertyUnavailable
    }
}
