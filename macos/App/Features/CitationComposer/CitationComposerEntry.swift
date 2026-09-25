import Foundation

/// Frozen arrival query for the citation composer. Parsed once from
/// `WorkspaceLocation`; identity switches do not write it back.
enum CitationComposerEntry: Equatable, Sendable {
    case addProperty(sourceID: String, subjectID: String, artifactID: String?)
    case edit(
        sourceID: String,
        subjectID: String,
        citationID: String,
        artifactID: String?,
        observationID: String?
    )
    case connect(
        sourceID: String,
        fromSubjectID: String,
        toSubjectID: String,
        bridgeTypeKey: String,
        termID: String?,
        gridX: Int64,
        gridY: Int64
    )

    var sourceID: String {
        switch self {
        case .addProperty(let sourceID, _, _),
             .edit(let sourceID, _, _, _, _),
             .connect(let sourceID, _, _, _, _, _, _):
            return sourceID
        }
    }

    var subjectID: String {
        switch self {
        case .addProperty(_, let subjectID, _),
             .edit(_, let subjectID, _, _, _):
            return subjectID
        case .connect:
            return ""
        }
    }

    var citationID: String? {
        switch self {
        case .edit(_, _, let citationID, _, _):
            return citationID
        case .addProperty, .connect:
            return nil
        }
    }

    var artifactID: String? {
        switch self {
        case .addProperty(_, _, let artifactID),
             .edit(_, _, _, let artifactID, _):
            return artifactID
        case .connect:
            return nil
        }
    }

    var observationID: String? {
        switch self {
        case .edit(_, _, _, _, let observationID):
            return observationID
        case .addProperty, .connect:
            return nil
        }
    }

    var isConnect: Bool {
        if case .connect = self { return true }
        return false
    }

    var connectFromSubjectID: String? {
        if case .connect(_, let from, _, _, _, _, _) = self { return from }
        return nil
    }

    var connectToSubjectID: String? {
        if case .connect(_, _, let to, _, _, _, _) = self { return to }
        return nil
    }

    var connectBridgeTypeKey: String? {
        if case .connect(_, _, _, let key, _, _, _) = self { return key }
        return nil
    }

    var connectDisambiguationTermID: String? {
        if case .connect(_, _, _, _, let term, _, _) = self { return term }
        return nil
    }

    var connectGridX: Int64 {
        if case .connect(_, _, _, _, _, let x, _) = self { return x }
        return 0
    }

    var connectGridY: Int64 {
        if case .connect(_, _, _, _, _, _, let y) = self { return y }
        return 0
    }

    /// Remount key: the frozen query, not the live document.
    var identityKey: String {
        switch self {
        case let .addProperty(sourceID, subjectID, artifactID):
            return "add-\(sourceID)-\(subjectID)-\(artifactID ?? "")"
        case let .edit(sourceID, subjectID, citationID, artifactID, observationID):
            return "edit-\(sourceID)-\(subjectID)-\(citationID)-\(artifactID ?? "")-\(observationID ?? "")"
        case let .connect(sourceID, fromID, toID, bridge, termID, gridX, gridY):
            return "connect-\(sourceID)-\(fromID)-\(toID)-\(bridge)-\(termID ?? "")-\(gridX)-\(gridY)"
        }
    }

    /// Fail-closed parse. Host mounts only when this succeeds.
    init?(location: WorkspaceLocation) {
        guard location.sourceSurface == .citationComposer,
              let sourceID = location.sourceId, !sourceID.isEmpty
        else { return nil }

        if location.subjectId == nil {
            guard let fromID = location.connectFromSubjectId, !fromID.isEmpty,
                  let toID = location.connectToSubjectId, !toID.isEmpty,
                  let bridge = location.connectBridgeTypeKey, !bridge.isEmpty
            else { return nil }
            self = .connect(
                sourceID: sourceID,
                fromSubjectID: fromID,
                toSubjectID: toID,
                bridgeTypeKey: bridge,
                termID: location.connectDisambiguationTermId,
                gridX: location.connectGridX ?? 0,
                gridY: location.connectGridY ?? 0
            )
            return
        }

        guard let subjectID = location.subjectId, !subjectID.isEmpty else { return nil }
        if let citationID = location.citationId, !citationID.isEmpty {
            self = .edit(
                sourceID: sourceID,
                subjectID: subjectID,
                citationID: citationID,
                artifactID: location.artifactId,
                observationID: location.observationId
            )
            return
        }
        self = .addProperty(
            sourceID: sourceID,
            subjectID: subjectID,
            artifactID: location.artifactId
        )
    }
}
