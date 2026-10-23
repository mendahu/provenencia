import Foundation

/// Distinguishes Source detail vs Evidence graph vs citation composer under the same `sourceId`.
/// Sidebar selection still uses `WorkspaceSection.sources` for both; this field
/// is for place identity / history / host routing only.
enum SourceSurface: String, Codable, Sendable, Equatable {
    case page
    case graph
    /// Citation composer place (Add property / connect handoff / edit citation).
    case citationComposer
    /// Promote flow place (S9-11): file `subjectId` onto a new or existing handle.
    case promote
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
    /// The Evidence graph of a graph, citationComposer or promote surface. A
    /// location of one of those surfaces without it is the Source page.
    var graphId: String?
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
    /// Properties type-strip category when `section == .properties`; nil is “All properties”.
    var subjectTypeKey: String?
    /// Properties inspector row when `section == .properties`.
    var propertyId: String?
    /// Canonical handle (Person / Event / Place) for a Conclude detail place.
    var entityId: String?
    /// Page vs Evidence graph vs composer when `section == .sources` and `sourceId` is set.
    /// Legacy history without this key, or a graph surface without `graphId`, decodes as `.page`.
    var sourceSurface: SourceSurface
    /// Promote opened from the graph's Promote all control, with every subject
    /// already expanded. Card Promote leaves this false. Legacy history decodes as false.
    var promoteAll: Bool
    /// Denormalized jump-menu cache; ignored for navigation identity.
    var ref: String?
    /// Denormalized jump-menu cache; ignored for navigation identity.
    /// For the citation composer this is the **subject scope** (card label), not the Source title.
    var title: String?
    /// Denormalized Source title for composer breadcrumbs (`Evidence graph for {source}`);
    /// ignored for navigation identity. Empty/legacy history falls back to the bare graph title.
    var sourceTitle: String?
    /// Id of the row at the top of a list place (Sources, People, Events, Places).
    /// Ignored for navigation identity, like `title` and `ref`. Written when leaving
    /// the list. A missing key in older history files decodes as nil, which opens at the top.
    var listScrollAnchor: String?

    enum CodingKeys: String, CodingKey {
        case section, sourceId, graphId, fieldId, typeId, subjectId, citationId, artifactId, observationId
        case connectFromSubjectId, connectToSubjectId, connectBridgeTypeKey
        case subjectTypeKey, propertyId, entityId
        case sourceSurface, promoteAll, ref, title, sourceTitle, listScrollAnchor
    }

    init(
        section: WorkspaceSection,
        sourceId: String? = nil,
        graphId: String? = nil,
        fieldId: String? = nil,
        typeId: String? = nil,
        subjectId: String? = nil,
        citationId: String? = nil,
        artifactId: String? = nil,
        observationId: String? = nil,
        connectFromSubjectId: String? = nil,
        connectToSubjectId: String? = nil,
        connectBridgeTypeKey: String? = nil,
        subjectTypeKey: String? = nil,
        propertyId: String? = nil,
        entityId: String? = nil,
        sourceSurface: SourceSurface = .page,
        promoteAll: Bool = false,
        ref: String? = nil,
        title: String? = nil,
        sourceTitle: String? = nil,
        listScrollAnchor: String? = nil
    ) {
        self.section = section
        self.sourceId = Self.nilIfEmpty(sourceId)
        self.graphId = Self.nilIfEmpty(graphId)
        self.fieldId = Self.nilIfEmpty(fieldId)
        self.typeId = Self.nilIfEmpty(typeId)
        self.subjectId = Self.nilIfEmpty(subjectId)
        self.citationId = Self.nilIfEmpty(citationId)
        self.artifactId = Self.nilIfEmpty(artifactId)
        self.observationId = Self.nilIfEmpty(observationId)
        self.connectFromSubjectId = Self.nilIfEmpty(connectFromSubjectId)
        self.connectToSubjectId = Self.nilIfEmpty(connectToSubjectId)
        self.connectBridgeTypeKey = Self.nilIfEmpty(connectBridgeTypeKey)
        self.subjectTypeKey = Self.nilIfEmpty(subjectTypeKey)
        self.propertyId = Self.nilIfEmpty(propertyId)
        self.entityId = Self.nilIfEmpty(entityId)
        self.sourceSurface = Self.normalizedSurface(sourceSurface, graphId: self.graphId)
        self.promoteAll = promoteAll
        self.ref = Self.nilIfEmpty(ref)
        self.title = Self.nilIfEmpty(title)
        self.sourceTitle = Self.nilIfEmpty(sourceTitle)
        self.listScrollAnchor = Self.nilIfEmpty(listScrollAnchor)
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        section = try container.decode(WorkspaceSection.self, forKey: .section)
        sourceId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .sourceId))
        graphId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .graphId))
        fieldId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .fieldId))
        typeId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .typeId))
        subjectId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .subjectId))
        citationId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .citationId))
        artifactId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .artifactId))
        observationId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .observationId))
        connectFromSubjectId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .connectFromSubjectId))
        connectToSubjectId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .connectToSubjectId))
        connectBridgeTypeKey = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .connectBridgeTypeKey))
        subjectTypeKey = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .subjectTypeKey))
        propertyId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .propertyId))
        entityId = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .entityId))
        sourceSurface = Self.normalizedSurface(
            try container.decodeIfPresent(SourceSurface.self, forKey: .sourceSurface) ?? .page,
            graphId: graphId
        )
        promoteAll = try container.decodeIfPresent(Bool.self, forKey: .promoteAll) ?? false
        ref = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .ref))
        title = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .title))
        sourceTitle = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .sourceTitle))
        listScrollAnchor = Self.nilIfEmpty(try container.decodeIfPresent(String.self, forKey: .listScrollAnchor))
    }

    /// Section list root (no deep id).
    static func sectionRoot(_ section: WorkspaceSection) -> WorkspaceLocation {
        WorkspaceLocation(section: section)
    }

    /// One Person's page (S9-09 stub; the page itself is S9-16).
    static func personDetail(entityId: String, ref: String, title: String?) -> WorkspaceLocation {
        WorkspaceLocation(section: .persons, entityId: entityId, ref: ref, title: title)
    }

    /// One Event's page (S9-23 stub; the page itself is S9-24).
    static func eventDetail(entityId: String, ref: String, title: String?) -> WorkspaceLocation {
        WorkspaceLocation(section: .events, entityId: entityId, ref: ref, title: title)
    }

    /// One Place's page (S9-27).
    static func placeDetail(entityId: String, ref: String, title: String?) -> WorkspaceLocation {
        WorkspaceLocation(section: .places, entityId: entityId, ref: ref, title: title)
    }

    /// An Evidence graph's canvas.
    static func evidenceGraph(
        sourceId: String,
        graphId: String,
        ref: String? = nil,
        title: String? = nil,
        sourceTitle: String? = nil
    ) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceId,
            graphId: graphId,
            sourceSurface: .graph,
            ref: ref,
            title: title,
            sourceTitle: sourceTitle
        )
    }

    /// The citation composer on an Evidence graph.
    static func citationComposer(
        sourceId: String,
        graphId: String,
        subjectId: String? = nil,
        citationId: String? = nil,
        artifactId: String? = nil,
        observationId: String? = nil,
        connectFromSubjectId: String? = nil,
        connectToSubjectId: String? = nil,
        connectBridgeTypeKey: String? = nil,
        ref: String? = nil,
        title: String? = nil,
        sourceTitle: String? = nil
    ) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceId,
            graphId: graphId,
            subjectId: subjectId,
            citationId: citationId,
            artifactId: artifactId,
            observationId: observationId,
            connectFromSubjectId: connectFromSubjectId,
            connectToSubjectId: connectToSubjectId,
            connectBridgeTypeKey: connectBridgeTypeKey,
            sourceSurface: .citationComposer,
            ref: ref,
            title: title,
            sourceTitle: sourceTitle
        )
    }

    /// The Sources, People, Events, or Places list. Detail places are not lists.
    var isListRoot: Bool {
        switch section {
        case .sources:
            return sourceId == nil
        case .persons, .events, .places:
            return entityId == nil
        case .sourceTypes, .metadata, .properties:
            return false
        }
    }

    /// Same place, with the list's top row replaced. Nil opens at the top.
    /// Identity is unchanged, so `go(to:)` coalesces instead of pushing.
    func withListScrollAnchor(_ anchor: String?) -> WorkspaceLocation {
        var copy = self
        copy.listScrollAnchor = Self.nilIfEmpty(anchor)
        return copy
    }

    /// Identity used for coalesce / equality of navigation — deep ids only.
    /// `ref`, `title`, `sourceTitle`, and `listScrollAnchor` are not identity.
    static func == (lhs: WorkspaceLocation, rhs: WorkspaceLocation) -> Bool {
        lhs.section == rhs.section
            && lhs.sourceId == rhs.sourceId
            && lhs.graphId == rhs.graphId
            && lhs.fieldId == rhs.fieldId
            && lhs.typeId == rhs.typeId
            && lhs.subjectId == rhs.subjectId
            && lhs.citationId == rhs.citationId
            && lhs.artifactId == rhs.artifactId
            && lhs.observationId == rhs.observationId
            && lhs.connectFromSubjectId == rhs.connectFromSubjectId
            && lhs.connectToSubjectId == rhs.connectToSubjectId
            && lhs.connectBridgeTypeKey == rhs.connectBridgeTypeKey
            && lhs.subjectTypeKey == rhs.subjectTypeKey
            && lhs.propertyId == rhs.propertyId
            && lhs.entityId == rhs.entityId
            && lhs.sourceSurface == rhs.sourceSurface
            && lhs.promoteAll == rhs.promoteAll
    }

    /// Pending Connect handoff: endpoints are known, the bridge is not written yet.
    var isConnectPrefill: Bool {
        sourceSurface == .citationComposer
            && subjectId == nil
            && connectFromSubjectId != nil
            && connectToSubjectId != nil
            && connectBridgeTypeKey != nil
    }

    /// A graph, citationComposer or promote surface needs its graph; without
    /// one the location is the Source page, as a history entry with no
    /// `sourceSurface` decodes. Every way a location is built goes through this.
    private static func normalizedSurface(_ surface: SourceSurface, graphId: String?) -> SourceSurface {
        switch surface {
        case .graph, .citationComposer, .promote:
            return graphId == nil ? .page : surface
        case .page:
            return .page
        }
    }

    private static func nilIfEmpty(_ value: String?) -> String? {
        guard let value, !value.isEmpty else { return nil }
        return value
    }
}
