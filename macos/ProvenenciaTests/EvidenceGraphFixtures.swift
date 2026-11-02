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

/// An unlink block for tests: `artifactID` is refused in `graphID` for `reason`. For
/// `.artifactInUse` its `citations` group lists that many
/// Citations; `.lastArtifact` has an empty group, what `GoStore` maps from an unset `citations`.
func unlinkBlock(
    _ artifactID: String,
    graphID: String,
    citations: Int,
    reason: EvidenceGraphUnlinkReason = .artifactInUse
) -> CatalogEvidenceGraphUnlinkBlock {
    let group: CatalogDeleteImpactGroup
    switch reason {
    case .lastArtifact:
        group = CatalogDeleteImpactGroup(via: "", kind: "", total: 0, listed: [])
    default:
        group = CatalogDeleteImpactGroup(
            via: reason.wire,
            kind: "citation",
            total: citations,
            listed: (0..<citations).map {
                PVDeleteImpactPreviewData.listed(ref: "CIT-\(artifactID)\($0 + 1)", title: "Smith, John")
            }
        )
    }
    return CatalogEvidenceGraphUnlinkBlock(graphID: graphID, artifactID: artifactID, reason: reason, citations: group)
}
