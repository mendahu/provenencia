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
    var subjectIconKind: PVSubjectIconKind {
        switch self {
        case .person: .person
        case .event: .event
        case .place: .place
        }
    }
}

extension EvidenceBridgeKind {
    var subjectIconKind: PVSubjectIconKind {
        switch self {
        case .relationship: .relationship
        case .participation: .participation
        case .location: .location
        }
    }
}
