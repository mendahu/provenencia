import SwiftUI

/// Exclusive pointer policy for the Evidence graph document (canvas health refactor).
enum EvidenceCanvasInputMode: Equatable {
    case idle
    case placing(EvidencePrimaryKind)
    case connecting
}

extension EvidenceGraphModel {
    /// Idle when no tool is armed (or while the create dialog is open).
    var inputMode: EvidenceCanvasInputMode {
        guard !isCreating else { return .idle }
        if armedConnect { return .connecting }
        guard let kind = armedKind else { return .idle }
        return .placing(kind)
    }
}

extension EvidencePrimaryKind {
    var markKey: PVMarkKey {
        switch self {
        case .person: .subjectPerson
        case .event: .subjectEvent
        case .place: .subjectPlace
        }
    }
}

extension EvidenceBridgeKind {
    var markKey: PVMarkKey {
        switch self {
        case .relationship: .subjectRelationship
        case .participation: .subjectParticipation
        case .location: .subjectLocation
        }
    }
}
