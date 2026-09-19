import SwiftUI

/// Exclusive pointer policy for the Evidence graph document (canvas health refactor).
enum EvidenceCanvasInputMode: Equatable {
    case idle
    case placing(EvidencePrimaryKind)
}

extension EvidenceGraphModel {
    /// Idle when no tool is armed (or while the create dialog is open).
    var inputMode: EvidenceCanvasInputMode {
        guard !isCreating, let kind = armedKind else { return .idle }
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
