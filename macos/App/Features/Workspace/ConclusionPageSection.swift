import Foundation

/// One read a page section makes, tagged by how the page treats it. Each
/// consumer's policy is a switch on `Role`, so a new role must decide each.
struct ConclusionSectionRead: Equatable, Sendable {
    enum Role: Equatable, Sendable {
        /// The section's own per-entity read; its host draws it in the section's band.
        case own
        /// A second per-entity read drawn in its own region (a column).
        case companion
        /// A catalog-wide vocabulary the rows label themselves with; its reload shows no Refreshing.
        case supporting

        /// The page's "Refreshing" follows this read.
        var showsRefreshing: Bool {
            switch self {
            case .own, .companion: return true
            case .supporting: return false
            }
        }

        /// The place warms this read when the location is applied, and keeps it visible.
        var warmsWithPlace: Bool {
            switch self {
            case .own, .companion, .supporting: return true
            }
        }
    }

    let key: CatalogQueryKey
    let role: Role
}

/// A section of a record page (the Person page's Vitals, Family, …). The raw
/// values are written to navigation history files, so they are frozen.
/// `WorkspaceLocation.pageSection` stores one; the first a kind offers is
/// stored as nil.
enum ConclusionPageSection: String, Codable, CaseIterable, Sendable {
    case vitals
    case family
    case timeline
    case gallery
    case sources

    /// The sections a record page of `section` offers, in order. A kind that
    /// has none returns `[]`; an Event or Place adopts sections by returning
    /// cases here.
    static func offered(by section: WorkspaceSection) -> [ConclusionPageSection] {
        switch section {
        case .persons:
            return allCases
        case .sources, .events, .places, .sourceTypes, .metadata, .properties:
            return []
        }
    }

    /// What the section reads beside the page's detail read, in the order the place warms
    /// them: own, then companion, then supporting. The place warms each, the page's
    /// "Refreshing" follows the roles that say so, and the section's host and views read
    /// the same keys. Empty for a section that reads nothing of its own.
    func reads(project: ProjectKey, entityId: String) -> [ConclusionSectionRead] {
        switch self {
        case .vitals, .gallery:
            return []
        case .family:
            return [.init(key: .personFamily(project: project, entityId: entityId), role: .own)]
        case .timeline:
            return [.init(key: .personEvents(project: project, entityId: entityId), role: .own)]
        case .sources:
            return [
                .init(key: .personClaims(project: project, entityId: entityId), role: .own),
                // The record links beside the rows: per-entity, so unlike the Source types
                // their reload shows Refreshing.
                .init(key: .personRecordLinks(project: project, entityId: entityId), role: .companion),
                // A catalog-wide vocabulary the rows label themselves with; the section reads
                // it with `session.queryHandle`.
                .init(key: .sourceTypesList(project: project), role: .supporting),
            ]
        }
    }

    /// The first read of `role`, nil when the section has none. A section has at most one
    /// own and one companion read; its host draws one key per role.
    func read(_ role: ConclusionSectionRead.Role, project: ProjectKey, entityId: String) -> CatalogQueryKey? {
        reads(project: project, entityId: entityId).first { $0.role == role }?.key
    }

    func queryKey(project: ProjectKey, entityId: String) -> CatalogQueryKey? {
        read(.own, project: project, entityId: entityId)
    }

    func supportingQueryKeys(project: ProjectKey, entityId: String) -> [CatalogQueryKey] {
        reads(project: project, entityId: entityId).filter { $0.role == .supporting }.map(\.key)
    }

    func companionQueryKey(project: ProjectKey, entityId: String) -> CatalogQueryKey? {
        read(.companion, project: project, entityId: entityId)
    }

    /// The section opts into Page columns' full-width mode: at Wide it fills the pane
    /// between the gutters while the header keeps its measure.
    var fillsWidth: Bool {
        switch self {
        case .timeline, .gallery, .sources: return true
        case .vitals, .family: return false
        }
    }

    var label: LocalizedStringResource {
        switch self {
        case .vitals: return L10n.Conclusions.pageSectionVitals
        case .family: return L10n.Conclusions.pageSectionFamily
        case .timeline: return L10n.Conclusions.pageSectionTimeline
        case .gallery: return L10n.Conclusions.pageSectionGallery
        case .sources: return L10n.Conclusions.pageSectionSources
        }
    }

    var glyph: PVGlyph {
        switch self {
        case .vitals: return .symbol(.account)
        case .family: return .symbol(.gitBranch)
        case .timeline: return .symbol(.calendar)
        case .gallery: return .symbol(.photo)
        case .sources: return .symbol(.library)
        }
    }
}
