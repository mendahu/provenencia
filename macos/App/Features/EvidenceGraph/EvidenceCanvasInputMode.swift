import SwiftUI

/// Exclusive pointer policy for the Evidence graph document (canvas health refactor).
enum EvidenceCanvasInputMode: Equatable {
    case idle
    case placing(EvidencePrimaryKind)
    case connecting
    /// Select armed or Shift held: an empty-canvas drag draws a rectangle.
    case selecting
}

extension EvidenceGraphModel {
    /// Idle when no tool is armed (or while a create/edit dialog is open).
    var inputMode: EvidenceCanvasInputMode {
        guard !isSheetPresented else { return .idle }
        if armedConnect { return .connecting }
        if let kind = armedKind { return .placing(kind) }
        return (armedSelect || shiftSelectHeld) ? .selecting : .idle
    }
}

/// What the canvas's top-right callout slot shows; the first match in `EvidenceGraphModel.canvasNotice` wins.
enum EvidenceGraphCanvasNotice: Equatable {
    case readError(String)
    case deleteError(String)
    case fileNotFound(String)
    case noArtifact
    case placing(EvidencePrimaryKind)
    case connecting
}

extension EvidenceGraphModel {
    /// The callout slot's content, in priority order: a failed delete and a missing file
    /// come from an action the researcher just took, so they show first; the failed graph
    /// read outranks the remaining notices because nothing on the canvas is current without the rows.
    var canvasNotice: EvidenceGraphCanvasNotice? {
        if pendingImpact == nil, let deleteError { return .deleteError(deleteError) }
        if pendingImpact == nil, deleteError == nil, let fileOpenError,
           resolvedFileArtifactID == selectedFileArtifactID
        {
            return .fileNotFound(fileOpenError)
        }
        if let graphReadError { return .readError(graphReadError) }
        if !canCite { return .noArtifact }
        switch inputMode {
        case .placing(let kind): return .placing(kind)
        case .connecting: return .connecting
        case .idle, .selecting: return nil
        }
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
        case .placeRelationship: .subjectPlaceRelationship
        }
    }
}
