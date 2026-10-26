import Foundation
import SwiftProtobuf

enum CoreErrorKind: Equatable, Sendable {
    case unspecified
    case user
    case conflict
    case notFound
    case `internal`

    init(_ proto: Provenencia_Engine_V1_ErrorKind) {
        switch proto {
        case .user: self = .user
        case .conflict: self = .conflict
        case .notFound: self = .notFound
        case .internal: self = .internal
        case .unspecified, .UNRECOGNIZED:
            self = .unspecified
        }
    }
}

enum CoreInvokeError: Error, Equatable {
    /// FFI status 1 with a decoded protobuf Error.
    case coded(status: Int32, code: String, kind: CoreErrorKind, params: [String])
    /// Non-zero status that could not be decoded as Error (e.g. malloc failure).
    case failed(status: Int32)
}

extension CoreInvokeError {
    /// The engine's error code; nil unless the error is `.coded`.
    var code: String? {
        if case .coded(_, let code, _, _) = self { return code }
        return nil
    }
}

extension CoreInvokeError: LocalizedError {
    var errorDescription: String? {
        switch self {
        case .coded(_, let code, _, let params):
            return L10n.Errors.message(code: code, params: params)
        case .failed:
            return L10n.string(L10n.Errors.unknown)
        }
    }
}

enum CoreMethod {
    static let getInstallIdentity = Int32(Provenencia_Engine_V1_Method.getInstallIdentity.rawValue)
    static let completeOnboarding = Int32(Provenencia_Engine_V1_Method.completeOnboarding.rawValue)
    static let getActiveProject = Int32(Provenencia_Engine_V1_Method.getActiveProject.rawValue)
    static let openProject = Int32(Provenencia_Engine_V1_Method.openProject.rawValue)
    static let removeActiveProject = Int32(Provenencia_Engine_V1_Method.removeActiveProject.rawValue)
    static let listProjectUsers = Int32(Provenencia_Engine_V1_Method.listProjectUsers.rawValue)
    static let signOut = Int32(Provenencia_Engine_V1_Method.signOut.rawValue)
    static let getProjectInfo = Int32(Provenencia_Engine_V1_Method.getProjectInfo.rawValue)
    static let listSources = Int32(Provenencia_Engine_V1_Method.listSources.rawValue)
    static let getSourceWorkspace = Int32(Provenencia_Engine_V1_Method.getSourceWorkspace.rawValue)
    static let createSource = Int32(Provenencia_Engine_V1_Method.createSource.rawValue)
    static let copySource = Int32(Provenencia_Engine_V1_Method.copySource.rawValue)
    static let listEvidenceGraphs = Int32(Provenencia_Engine_V1_Method.listEvidenceGraphs.rawValue)
    static let createEvidenceGraph = Int32(Provenencia_Engine_V1_Method.createEvidenceGraph.rawValue)
    static let updateEvidenceGraph = Int32(Provenencia_Engine_V1_Method.updateEvidenceGraph.rawValue)
    static let deleteEvidenceGraph = Int32(Provenencia_Engine_V1_Method.deleteEvidenceGraph.rawValue)
    static let linkEvidenceGraphArtifact = Int32(Provenencia_Engine_V1_Method.linkEvidenceGraphArtifact.rawValue)
    static let unlinkEvidenceGraphArtifact = Int32(Provenencia_Engine_V1_Method.unlinkEvidenceGraphArtifact.rawValue)
    static let getEvidenceGraphUnlinkImpact = Int32(Provenencia_Engine_V1_Method.getEvidenceGraphUnlinkImpact.rawValue)
    static let updateSource = Int32(Provenencia_Engine_V1_Method.updateSource.rawValue)
    static let addSourceNote = Int32(Provenencia_Engine_V1_Method.addSourceNote.rawValue)
    static let updateSourceNote = Int32(Provenencia_Engine_V1_Method.updateSourceNote.rawValue)
    static let deleteSourceNote = Int32(Provenencia_Engine_V1_Method.deleteSourceNote.rawValue)
    static let setSourceMetadata = Int32(Provenencia_Engine_V1_Method.setSourceMetadata.rawValue)
    static let clearSourceMetadata = Int32(Provenencia_Engine_V1_Method.clearSourceMetadata.rawValue)
    static let dismissSourceMetadataSuggestion = Int32(Provenencia_Engine_V1_Method.dismissSourceMetadataSuggestion.rawValue)
    static let reorderSourceMetadata = Int32(Provenencia_Engine_V1_Method.reorderSourceMetadata.rawValue)
    static let createArtifact = Int32(Provenencia_Engine_V1_Method.createArtifact.rawValue)
    static let updateArtifact = Int32(Provenencia_Engine_V1_Method.updateArtifact.rawValue)
    static let ingestArtifactFile = Int32(Provenencia_Engine_V1_Method.ingestArtifactFile.rawValue)
    static let listSourceCredibilityGrades = Int32(Provenencia_Engine_V1_Method.listSourceCredibilityGrades.rawValue)
    static let upsertSourceCredibilityAssessment = Int32(Provenencia_Engine_V1_Method.upsertSourceCredibilityAssessment.rawValue)
    static let listSourceTypes = Int32(Provenencia_Engine_V1_Method.listSourceTypes.rawValue)
    static let createSourceType = Int32(Provenencia_Engine_V1_Method.createSourceType.rawValue)
    static let listMetadataFields = Int32(Provenencia_Engine_V1_Method.listMetadataFields.rawValue)
    static let createMetadataField = Int32(Provenencia_Engine_V1_Method.createMetadataField.rawValue)
    static let updateMetadataField = Int32(Provenencia_Engine_V1_Method.updateMetadataField.rawValue)
    static let deleteMetadataField = Int32(Provenencia_Engine_V1_Method.deleteMetadataField.rawValue)
    static let updateSourceType = Int32(Provenencia_Engine_V1_Method.updateSourceType.rawValue)
    static let deleteSourceType = Int32(Provenencia_Engine_V1_Method.deleteSourceType.rawValue)
    static let listTypeSuggestions = Int32(Provenencia_Engine_V1_Method.listTypeSuggestions.rawValue)
    static let assignTypeField = Int32(Provenencia_Engine_V1_Method.assignTypeField.rawValue)
    static let removeTypeField = Int32(Provenencia_Engine_V1_Method.removeTypeField.rawValue)
    static let getWorkspaceNavCounts = Int32(Provenencia_Engine_V1_Method.getWorkspaceNavCounts.rawValue)
    static let closeCatalogSession = Int32(Provenencia_Engine_V1_Method.closeCatalogSession.rawValue)
    static let setSourceCover = Int32(Provenencia_Engine_V1_Method.setSourceCover.rawValue)
    static let searchCatalog = Int32(Provenencia_Engine_V1_Method.searchCatalog.rawValue)
    static let listSubjectTypes = Int32(Provenencia_Engine_V1_Method.listSubjectTypes.rawValue)
    static let createSubject = Int32(Provenencia_Engine_V1_Method.createSubject.rawValue)
    static let updateSubject = Int32(Provenencia_Engine_V1_Method.updateSubject.rawValue)
    static let deleteSubject = Int32(Provenencia_Engine_V1_Method.deleteSubject.rawValue)
    static let listSubjects = Int32(Provenencia_Engine_V1_Method.listSubjects.rawValue)
    static let setSubjectPosition = Int32(Provenencia_Engine_V1_Method.setSubjectPosition.rawValue)
    static let clearSubjectPosition = Int32(Provenencia_Engine_V1_Method.clearSubjectPosition.rawValue)
    static let listSubjectPositions = Int32(Provenencia_Engine_V1_Method.listSubjectPositions.rawValue)
    static let createProperty = Int32(Provenencia_Engine_V1_Method.createProperty.rawValue)
    static let updateProperty = Int32(Provenencia_Engine_V1_Method.updateProperty.rawValue)
    static let deleteProperty = Int32(Provenencia_Engine_V1_Method.deleteProperty.rawValue)
    static let listPropertyTerms = Int32(Provenencia_Engine_V1_Method.listPropertyTerms.rawValue)
    static let createPropertyTerm = Int32(Provenencia_Engine_V1_Method.createPropertyTerm.rawValue)
    static let assignSubjectTypeProperty = Int32(Provenencia_Engine_V1_Method.assignSubjectTypeProperty.rawValue)
    static let removeSubjectTypeProperty = Int32(Provenencia_Engine_V1_Method.removeSubjectTypeProperty.rawValue)
    static let listConnectRules = Int32(Provenencia_Engine_V1_Method.listConnectRules.rawValue)
    static let createCitedBridge = Int32(Provenencia_Engine_V1_Method.createCitedBridge.rawValue)
    static let createCitationWithObservations = Int32(Provenencia_Engine_V1_Method.createCitationWithObservations.rawValue)
    static let addObservationsToCitation = Int32(Provenencia_Engine_V1_Method.addObservationsToCitation.rawValue)
    static let listObservationsByEvidenceGraph = Int32(Provenencia_Engine_V1_Method.listObservationsByEvidenceGraph.rawValue)
    static let getCitation = Int32(Provenencia_Engine_V1_Method.getCitation.rawValue)
    static let citationCountsByEvidenceGraph = Int32(Provenencia_Engine_V1_Method.citationCountsByEvidenceGraph.rawValue)
    static let listCitationsByArtifact = Int32(Provenencia_Engine_V1_Method.listCitationsByArtifact.rawValue)
    static let updateCitation = Int32(Provenencia_Engine_V1_Method.updateCitation.rawValue)
    static let updateObservation = Int32(Provenencia_Engine_V1_Method.updateObservation.rawValue)
    static let deleteObservation = Int32(Provenencia_Engine_V1_Method.deleteObservation.rawValue)
    static let getPropertiesWorkspace = Int32(Provenencia_Engine_V1_Method.getPropertiesWorkspace.rawValue)
    static let listEvidenceGraphProgress = Int32(Provenencia_Engine_V1_Method.listEvidenceGraphProgress.rawValue)
    static let getEvidenceGraphProgress = Int32(Provenencia_Engine_V1_Method.getEvidenceGraphProgress.rawValue)
    static let getDeleteImpact = Int32(Provenencia_Engine_V1_Method.getDeleteImpact.rawValue)
    static let deleteSource = Int32(Provenencia_Engine_V1_Method.deleteSource.rawValue)
    static let deleteArtifact = Int32(Provenencia_Engine_V1_Method.deleteArtifact.rawValue)
    static let deleteCitation = Int32(Provenencia_Engine_V1_Method.deleteCitation.rawValue)
    static let promoteSubject = Int32(Provenencia_Engine_V1_Method.promoteSubject.rawValue)
    static let unpromoteSubject = Int32(Provenencia_Engine_V1_Method.unpromoteSubject.rawValue)
    static let setPreferredLikeness = Int32(Provenencia_Engine_V1_Method.setPreferredLikeness.rawValue)
    static let listSubjectMemberships = Int32(Provenencia_Engine_V1_Method.listSubjectMemberships.rawValue)
    static let listPersonHeaders = Int32(Provenencia_Engine_V1_Method.listPersonHeaders.rawValue)
    static let listClaimConfidenceGrades = Int32(Provenencia_Engine_V1_Method.listClaimConfidenceGrades.rawValue)
    static let getConclusionDetail = Int32(Provenencia_Engine_V1_Method.getConclusionDetail.rawValue)
    static let listEventHeaders = Int32(Provenencia_Engine_V1_Method.listEventHeaders.rawValue)
    static let listPlaceHeaders = Int32(Provenencia_Engine_V1_Method.listPlaceHeaders.rawValue)
    static let listEvidenceGraphEventTitles = Int32(Provenencia_Engine_V1_Method.listEvidenceGraphEventTitles.rawValue)
    static let proposePromoteGraphAlignment = Int32(Provenencia_Engine_V1_Method.proposePromoteGraphAlignment.rawValue)
    static let applyPromoteGraphAlignment = Int32(Provenencia_Engine_V1_Method.applyPromoteGraphAlignment.rawValue)
    static let getPromoteClaimEvidence = Int32(Provenencia_Engine_V1_Method.getPromoteClaimEvidence.rawValue)
    static let getNameFormat = Int32(Provenencia_Engine_V1_Method.getNameFormat.rawValue)
}

func provenenciaInvoke(method: Int32, request: Data) throws -> Data {
    var outPtr: UnsafeMutablePointer<UInt8>?
    var outLen: Int = 0
    let status: Int32 = request.withUnsafeBytes { raw in
        let inPtr = raw.bindMemory(to: UInt8.self).baseAddress
        return provenencia_call(method, inPtr, request.count, &outPtr, &outLen)
    }
    defer {
        if let outPtr {
            provenencia_free(UnsafeMutableRawPointer(outPtr))
        }
    }
    if status != 0 {
        if status == 1, let outPtr, outLen > 0 {
            let data = Data(bytes: outPtr, count: outLen)
            if let decoded = try? Provenencia_Engine_V1_Error(serializedBytes: data), !decoded.code.isEmpty {
                throw CoreInvokeError.coded(
                    status: status,
                    code: decoded.code,
                    kind: CoreErrorKind(decoded.kind),
                    params: decoded.params
                )
            }
        }
        throw CoreInvokeError.failed(status: status)
    }
    guard let outPtr, outLen > 0 else {
        return Data()
    }
    return Data(bytes: outPtr, count: outLen)
}

func provenenciaCall<Request: Message & Sendable, Response: Message & Sendable>(
    method: Int32,
    request: Request
) async throws -> Response {
    try await Task.detached {
        let out = try provenenciaInvoke(method: method, request: try request.serializedData())
        return try Response(serializedBytes: out)
    }.value
}
