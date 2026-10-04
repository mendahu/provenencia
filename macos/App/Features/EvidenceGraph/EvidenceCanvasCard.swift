import SwiftUI

/// Everything one graph card draws, as a value.
///
/// The graph pushes cards to ``GraphCanvasItemLayer`` as ``GraphCanvasItem``s
/// with this as their model, and AppKit rebuilds a card's hosting view only when
/// its model changes. Drag position is not part of it: AppKit moves the view.
struct EvidenceCanvasCard: Equatable {
    enum Placed: Equatable {
        case subject(SourceGraphPlacedSubject, CatalogSubjectTypePresentation?)
        case bridge(SourceGraphPlacedBridge)
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
    }

    var placed: Placed
    var text: EvidenceCardText
    var state: State

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

/// Builds a card's hosted view: paint, measured layout, and its VoiceOver actions.
///
/// Holds references only (the model, navigation, and layout store), so item
/// content closures never capture a SwiftUI view's state.
@MainActor
struct EvidenceCanvasCardContent {
    let model: EvidenceGraphModel
    let navigation: WorkspaceNavigation
    let layouts: EvidenceCardLayoutStore

    func view(for card: EvidenceCanvasCard) -> AnyView {
        switch card.placed {
        case .subject(let placed, let presentation):
            AnyView(subject(placed, presentation: presentation, card: card))
        case .bridge(let placed):
            AnyView(bridge(placed, card: card))
        }
    }

    private func subject(
        _ placed: SourceGraphPlacedSubject,
        presentation: CatalogSubjectTypePresentation?,
        card: EvidenceCanvasCard
    ) -> some View {
        let model = model
        let navigation = navigation
        let layouts = layouts
        return EvidenceSubjectCard(
            placed: placed,
            text: card.text,
            presentation: presentation,
            isSelected: card.state.isSelected,
            isActivated: card.state.isActivated,
            isConnectingFrom: card.state.isConnectingFrom,
            canCite: card.state.canCite,
            isDragging: card.state.isDragging,
            hoveredActionID: card.state.hoveredActionID,
            pressedActionID: card.state.pressedActionID,
            onLayout: { layouts.record($0, for: placed.id) }
        )
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editAccessibility)) {
            model.beginEdit(subjectID: placed.id)
        }
        .accessibilityAction(named: Text(Self.addPropertyActionName(canCite: card.state.canCite))) {
            if let location = model.composerLocation(for: placed.id) {
                navigation.go(to: location)
            }
        }
        .accessibilityAction(named: Text(verbatim: card.text.footerActionName ?? "")) {
            if placed.membership == nil {
                model.beginPromote(subjectID: placed.id)
            } else if let location = model.openHandle(subjectID: placed.id) {
                navigation.go(to: location)
            }
        }
        .accessibilityAction(named: Text(verbatim: card.text.deleteActionName)) {
            Task { await model.beginDelete(subjectID: placed.id) }
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editPropertyAccessibility)) {
            if let observation = placed.observations.first,
               let location = model.composerLocation(
                   forObservationID: observation.id,
                   subjectID: placed.id
               )
            {
                navigation.go(to: location)
            }
        }
    }

    private func bridge(_ placed: SourceGraphPlacedBridge, card: EvidenceCanvasCard) -> some View {
        let model = model
        let navigation = navigation
        let layouts = layouts
        return EvidenceBridgeCard(
            placed: placed,
            text: card.text,
            isSelected: card.state.isSelected,
            isActivated: card.state.isActivated,
            isDragging: card.state.isDragging,
            hoveredActionID: card.state.hoveredActionID,
            canCite: card.state.canCite,
            onLayout: { layouts.record($0, for: placed.id) }
        )
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editAccessibility)) {
            model.beginEdit(subjectID: placed.id)
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editCitationAccessibility)) {
            if let location = model.composerLocationForBridgeCitation(subjectID: placed.id) {
                navigation.go(to: location)
            }
        }
        .accessibilityAction(named: Text(Self.addPropertyActionName(canCite: card.state.canCite))) {
            if let location = model.composerLocation(for: placed.id) {
                navigation.go(to: location)
            }
        }
        .accessibilityAction(named: Text(verbatim: card.text.deleteActionName)) {
            Task { await model.beginDelete(subjectID: placed.id) }
        }
    }

    private static func addPropertyActionName(canCite: Bool) -> LocalizedStringResource {
        canCite ? L10n.EvidenceGraph.addProperty : L10n.EvidenceGraph.addPropertyUnavailable
    }
}
