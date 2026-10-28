import Foundation

/// Cache identity for one catalog read. New workspace places add a case here
/// and an arm on `CatalogQueryKey.Kind.descriptor`.
enum CatalogQueryKey: Hashable, Sendable {
    case sourcesList(project: ProjectKey)
    case sourceTypesList(project: ProjectKey)
    case metadataFieldsList(project: ProjectKey)
    case credibilityGradesList(project: ProjectKey)
    /// Identity Claim confidence grades (seeded vocabulary; S9-12).
    case confidenceGradesList(project: ProjectKey)
    case sourceWorkspace(project: ProjectKey, sourceId: String)
    /// One Evidence graph's rows: the canvas, the composer and Promote read it.
    case evidenceGraph(project: ProjectKey, graphId: String)
    case citationCounts(project: ProjectKey, graphId: String)
    case typeSuggestions(project: ProjectKey, typeId: String)
    case propertiesWorkspace(project: ProjectKey)
    case connectRules(project: ProjectKey)
    case propertyTerms(project: ProjectKey, propertyId: String)
    case citationsByArtifact(project: ProjectKey, graphId: String, artifactId: String)
    case evidenceGraphProgress(project: ProjectKey)
    /// Every Evidence graph of the project with its linked Artifact ids (S12-05).
    case evidenceGraphsList(project: ProjectKey)
    /// Every Person header, composed from the auto-reconciler cache (S9-07).
    case personsList(project: ProjectKey)
    /// Every Event header, composed from the auto-reconciler cache (S9-22).
    case eventsList(project: ProjectKey)
    /// Every Place header, composed from the auto-reconciler cache (S9-25).
    case placesList(project: ProjectKey)
    /// One handle's detail, any kind (S9-15), with its kind's header (birth
    /// and death, an Event's subjects and title, a Place's names) read in the
    /// same call. Evicted, not revalidated, while off screen.
    case conclusionDetail(project: ProjectKey, entityId: String)

    /// Case identity without associated payload. Each kind has one
    /// `CatalogQueryDescriptor` (`Kind.descriptor`).
    enum Kind: Hashable, Sendable, CaseIterable {
        case sourcesList
        case sourceTypesList
        case metadataFieldsList
        case credibilityGradesList
        case confidenceGradesList
        case sourceWorkspace
        case evidenceGraph
        case citationCounts
        case typeSuggestions
        case propertiesWorkspace
        case connectRules
        case propertyTerms
        case citationsByArtifact
        case evidenceGraphProgress
        case evidenceGraphsList
        case personsList
        case eventsList
        case placesList
        case conclusionDetail
    }

    var kind: Kind {
        switch self {
        case .sourcesList:
            return .sourcesList
        case .sourceTypesList:
            return .sourceTypesList
        case .metadataFieldsList:
            return .metadataFieldsList
        case .credibilityGradesList:
            return .credibilityGradesList
        case .confidenceGradesList:
            return .confidenceGradesList
        case .sourceWorkspace:
            return .sourceWorkspace
        case .evidenceGraph:
            return .evidenceGraph
        case .citationCounts:
            return .citationCounts
        case .typeSuggestions:
            return .typeSuggestions
        case .propertiesWorkspace:
            return .propertiesWorkspace
        case .connectRules:
            return .connectRules
        case .propertyTerms:
            return .propertyTerms
        case .citationsByArtifact:
            return .citationsByArtifact
        case .evidenceGraphProgress:
            return .evidenceGraphProgress
        case .evidenceGraphsList:
            return .evidenceGraphsList
        case .personsList:
            return .personsList
        case .eventsList:
            return .eventsList
        case .placesList:
            return .placesList
        case .conclusionDetail:
            return .conclusionDetail
        }
    }

    var project: ProjectKey {
        switch self {
        case .sourcesList(let project),
             .sourceTypesList(let project),
             .metadataFieldsList(let project),
             .credibilityGradesList(let project),
             .confidenceGradesList(let project),
             .sourceWorkspace(let project, _),
             .evidenceGraph(let project, _),
             .citationCounts(let project, _),
             .typeSuggestions(let project, _),
             .propertiesWorkspace(let project),
             .connectRules(let project),
             .propertyTerms(let project, _),
             .citationsByArtifact(let project, _, _),
             .evidenceGraphProgress(let project),
             .evidenceGraphsList(let project),
             .personsList(let project),
             .eventsList(let project),
             .placesList(let project),
             .conclusionDetail(let project, _):
            project
        }
    }
}
