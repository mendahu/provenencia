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

/// An unlink block for tests: `artifactID` is in use by `citations` Citations of the graph.
func unlinkBlock(_ artifactID: String, citations: Int) -> CatalogEvidenceGraphUnlinkBlock {
    CatalogEvidenceGraphUnlinkBlock(
        artifactID: artifactID,
        citations: CatalogDeleteImpactGroup(
            via: "evidencegraphs.artifact_in_use",
            kind: "citation",
            total: citations,
            listed: (0..<citations).map {
                PVDeleteImpactPreviewData.listed(ref: "CIT-\(artifactID)\($0 + 1)", title: "Smith, John")
            }
        )
    )
}
