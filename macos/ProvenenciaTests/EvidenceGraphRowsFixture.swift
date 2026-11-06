import Foundation
@testable import Provenencia

/// Session graph payload. `types` is ignored — vocabulary lives on `propertiesWorkspace`.
func graphRows(
    graphId: String,
    subjects: [CatalogSubject],
    positions: [CatalogSubjectPosition],
    types: [CatalogSubjectType] = [],
    observations: [CatalogObservation] = [],
    memberships: [CatalogSubjectMembership] = []
) -> EvidenceGraphRows {
    _ = types
    return EvidenceGraphRows(
        graphId: graphId,
        subjects: subjects,
        positions: positions,
        observations: observations,
        memberships: memberships
    )
}
