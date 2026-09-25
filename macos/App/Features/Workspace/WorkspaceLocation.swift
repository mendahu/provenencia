import Foundation

/// Distinguishes Source detail vs Evidence graph vs citation composer under the same `sourceId`.
/// Sidebar selection still uses `WorkspaceSection.sources` for both; this field
/// is for place identity / history / host routing only.
enum SourceSurface: String, Codable, Sendable, Equatable {
    case page
    case graph
    /// Citation composer place (Add property / connect handoff / edit citation).
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
    /// Existing Citation when editing from a property row; nil for create.
    var citationId: String?
    /// Artifact already known at entry (pencil). Identity switches do not write this back.
    var artifactId: String?
    /// Observation to focus when opening an existing Citation (pencil).
    var observationId: String?
    /// Connect-prefill endpoints when `sourceSurface == .citationComposer` and `subjectId` is nil.
    var connectFromSubjectId: String?
    var connectToSubjectId: String?
    var connectBridgeTypeKey: String?
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
        case section, sourceId, fieldId, typeId, subjectId, citationId, artifactId, observationId
        case connectFromSubjectId, connectToSubjectId, connectBridgeTypeKey
        case sourceSurface, ref, title, sourceTitle
    }

    init(
        section: WorkspaceSection,
        sourceId: String? = nil,
        fieldId: String? = nil,
        typeId: String? = nil,
        subjectId: String? = nil,
        citationId: String? = nil,
        artifactId: String? = nil,
        observationId: String? = nil,
        connectFromSubjectId: String? = nil,
        connectToSubjectId: String? = nil,
        connectBridgeTypeKey: String? = nil,
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
        self.citationId = Self.nilIfEmpty(citationId)
        self.artifactId = Self.nilIfEmpty(artifactId)
        self.observationId = Self.nilIfEmpty(observationId)
        self.connectFromSubjectId = Self.nilIfEmpty(connectFromSubjectId)
        self.connectToSubjectId = Self.nilIfEmpty(connectToSubjectId)
        self.connectBridgeTypeKey = Self.nilIfEmpty(connectBridgeTypeKey)
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
        citationId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .citationId))
        artifactId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .artifactId))
        observationId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .observationId))
        connectFromSubjectId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .connectFromSubjectId))
        connectToSubjectId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .connectToSubjectId))
        connectBridgeTypeKey = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .connectBridgeTypeKey))
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
            && lhs.citationId == rhs.citationId
            && lhs.artifactId == rhs.artifactId
            && lhs.observationId == rhs.observationId
            && lhs.connectFromSubjectId == rhs.connectFromSubjectId
            && lhs.connectToSubjectId == rhs.connectToSubjectId
            && lhs.connectBridgeTypeKey == rhs.connectBridgeTypeKey
            && lhs.sourceSurface == rhs.sourceSurface
    }

    /// Pending Connect handoff: endpoints are known, the bridge is not written yet.
    var isConnectPrefill: Bool {
        sourceSurface == .citationComposer
            && subjectId == nil
            && connectFromSubjectId != nil
            && connectToSubjectId != nil
            && connectBridgeTypeKey != nil
    }

    private static func nilIfEmpty(_ value: String?) -> String? {
        guard let value, !value.isEmpty else { return nil }
        return value
    }
}
