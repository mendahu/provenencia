import Foundation
@testable import Provenencia

/// An Evidence graph row for tests. `ref` defaults to `GPH-` plus the uppercased `id`.
func evidenceGraph(
    _ id: String,
    ref: String? = nil,
    sourceID: String = "src-1",
    name: String = "Default",
    description: String = "",
    artifactIDs: [String] = []
) -> CatalogEvidenceGraph {
    CatalogEvidenceGraph(
        id: id, ref: ref ?? "GPH-\(id.uppercased())", sourceID: sourceID, name: name,
        description: description, artifactIDs: artifactIDs
    )
}
