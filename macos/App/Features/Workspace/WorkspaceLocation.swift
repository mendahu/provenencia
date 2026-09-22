import Foundation

/// Distinguishes Source detail vs Evidence graph vs citation composer under the same `sourceId`.
/// Sidebar selection still uses `WorkspaceSection.sources` for both; this field
/// is for place identity / history / host routing only.
enum SourceSurface: String, Codable, Sendable, Equatable {
    case page
    case graph
    /// Citation composer place (Add property / connect handoff). Stub until S7-08.
    case citationComposer
}

/// Result of reconciling a destination model to a navigation `WorkspaceLocation`.
/// Views map `.missingDeepId` to `navigation.fallbackToSectionRoot()`.
enum WorkspaceLocationReconcile: Equatable, Sendable {
    /// Wrong section, or blocked (e.g. add form open).
    case ignored
    /// Model state now matches the location (including section list root).
    case applied
    /// Deep id was requested but absent from the catalog after load.
    case missingDeepId
}

/// A restoreable workspace place: sidebar destination plus optional deep
/// location (Source page, vocabulary row, …). Persisted in navigation history.
/// See `docs/deployment-plan/archive/spike-3/navigation-history.md`.
struct WorkspaceLocation: Codable, Equatable, Sendable {
    var section: WorkspaceSection
    var sourceId: String?
    var fieldId: String?
    var typeId: String?
    /// Subject scoped into the citation composer when `sourceSurface == .citationComposer`.
    var subjectId: String?
    /// Page vs Evidence graph vs composer when `section == .sources` and `sourceId` is set.
    /// Legacy history without this key decodes as `.page`.
    var sourceSurface: SourceSurface
    /// Denormalized jump-menu cache; ignored for navigation identity.
    var ref: String?
    /// Denormalized jump-menu cache; ignored for navigation identity.
    /// For the citation composer this is the **subject scope** (card label), not the Source title.
    var title: String?
    /// Denormalized Source title for composer breadcrumbs (`Evidence graph for {source}`);
    /// ignored for navigation identity. Empty/legacy history falls back to the bare graph title.
    var sourceTitle: String?

    enum CodingKeys: String, CodingKey {
        case section, sourceId, fieldId, typeId, subjectId, sourceSurface, ref, title, sourceTitle
    }

    init(
        section: WorkspaceSection,
        sourceId: String? = nil,
        fieldId: String? = nil,
        typeId: String? = nil,
        subjectId: String? = nil,
        sourceSurface: SourceSurface = .page,
        ref: String? = nil,
        title: String? = nil,
        sourceTitle: String? = nil
    ) {
        self.section = section
        self.sourceId = Self.nilIfEmpty(sourceId)
        self.fieldId = Self.nilIfEmpty(fieldId)
        self.typeId = Self.nilIfEmpty(typeId)
        self.subjectId = Self.nilIfEmpty(subjectId)
        self.sourceSurface = sourceSurface
        self.ref = Self.nilIfEmpty(ref)
        self.title = Self.nilIfEmpty(title)
        self.sourceTitle = Self.nilIfEmpty(sourceTitle)
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        section = try container.decode(WorkspaceSection.self, forKey: .section)
        sourceId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .sourceId))
        fieldId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .fieldId))
        typeId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .typeId))
        subjectId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .subjectId))
        sourceSurface = try container.decodeIfPresent(SourceSurface.self, forKey: .sourceSurface) ?? .page
        ref = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .ref))
        title = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .title))
        sourceTitle = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .sourceTitle))
    }

    /// Section list root (no deep id).
    static func sectionRoot(_ section: WorkspaceSection) -> WorkspaceLocation {
        WorkspaceLocation(section: section)
    }

    /// Identity used for coalesce / equality of navigation — deep ids only.
    static func == (lhs: WorkspaceLocation, rhs: WorkspaceLocation) -> Bool {
        lhs.section == rhs.section
            && lhs.sourceId == rhs.sourceId
            && lhs.fieldId == rhs.fieldId
            && lhs.typeId == rhs.typeId
            && lhs.subjectId == rhs.subjectId
            && lhs.sourceSurface == rhs.sourceSurface
    }

    private static func nilIfEmpty(_ value: String?) -> String? {
        guard let value else { return nil }
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }
}
