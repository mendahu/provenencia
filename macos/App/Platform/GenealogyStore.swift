import Foundation

struct InstallIdentity: Sendable, Equatable {
    var userID: String
    var displayName: String
    var ref: String
}

struct ProjectInfo: Sendable, Equatable {
    var label: String
    var folderName: String
    var createdAt: String
    var updatedAt: String
    var updatedByUserID: String
    var updatedByDisplayName: String
    var updatedByRef: String
    /// Durable catalog project identity (UUIDv7 string). Empty only if absent.
    var uuid: String = ""
}

struct OnboardingResult: Sendable, Equatable {
    var projectDir: String
    var userID: String
    var displayName: String
    var ref: String
    var project: ProjectInfo
}

struct CatalogSource: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var sourceTypeID: String
    var title: String
    var description: String
    /// Thumbnail JPEG under `objects/…` when cover is a pinned raster Artifact.
    var thumbnailRelPath: String = ""
    /// Unused on Source (file-type glyphs stay on Artifact rows). Kept for wire compat.
    var thumbnailMediaType: String = ""
    var thumbnailOriginalFilename: String = ""
    /// `artifact` (pinned raster primary) or `type_icon`.
    var coverMode: String = "type_icon"
    /// Set when `coverMode` is `artifact`.
    var primaryArtifactID: String = ""
    /// True when this Source has at least one Artifact (fileless counts).
    /// Used by the Sources list Evidence graph gate — not cover presence.
    var hasArtifact: Bool = false
    /// Latest audit revision for this source entity (create or identity update).
    /// Sources list "Updated" sort; zero means unset / unknown.
    var updatedRevision: Int64 = 0
}

struct CatalogSourceNote: Sendable, Equatable {
    var id: String
    var sourceID: String
    var body: String
    /// Create attribution from audit (not a domain column).
    var authorDisplayName: String
    /// RFC3339 UTC create time from audit.
    var createdAt: String
}

struct CatalogFileRef: Sendable, Equatable {
    var id: String
    var relPath: String
    var originalFilename: String
    var mediaType: String
    var byteSize: Int64
}

struct CatalogArtifact: Sendable, Equatable {
    var id: String
    var ref: String
    var sourceID: String
    var fileID: String
    var label: String
    var description: String
    var file: CatalogFileRef?
    /// Thumbnail JPEG under `objects/…` (empty when fileless / non-image / skipped).
    var thumbnailRelPath: String = ""
}

struct CatalogCredibilityGrade: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var sortOrder: Int
}

/// One omnibar / SearchCatalog hit (stable kinds: source, source_type, source_field).
struct CatalogSearchHit: Sendable, Equatable, Identifiable {
    var kind: String
    var id: String
    var ref: String
    var title: String
    var subtitle: String
    /// Stable match field code from Go (`title`, `notes`, …) — not UI copy.
    var matchReason: String
    /// Optional raw snippet for body/rollup matches; Mac localizes the prefix.
    var matchSnippet: String = ""
    var location: WorkspaceLocation
    /// Source cover raster path when already derived; empty otherwise.
    var thumbnailRelPath: String = ""
    /// Source-type icon key (type hits and Source type fallback).
    var iconKey: String = ""
}

struct CatalogSubjectType: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var description: String
    var refPrefix: String
    var candidateRefPrefix: String
}

struct CatalogSubject: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var sourceID: String
    var subjectTypeID: String
    var label: String
    var description: String
}

struct CatalogSubjectPosition: Sendable, Equatable {
    var subjectID: String
    var gridX: Int64
    var gridY: Int64
}

struct CatalogCitation: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var artifactID: String
    var locatorJSON: String
    var transcription: String
    var description: String
    var transcriptionUncertain: Bool
    var transcriptionNote: String
}

/// Citation list row for the composer identity menu (ref + transcription + count).
struct CatalogListedCitation: Sendable, Equatable, Identifiable {
    var citation: CatalogCitation
    var observationCount: Int
    var id: String { citation.id }
}

/// One ordered NameValue part as stored on an Observation.
struct CatalogNameValuePart: Sendable, Equatable {
    var value: String
    var type: String
}

/// One Observation row with Property summary (graph / card payloads).
struct CatalogObservation: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var citationID: String
    var subjectID: String
    var propertyID: String
    var polarity: String
    var valueText: String
    var valueInteger: Int64?
    var valueDateID: String
    /// Structured DateValue when listed from the catalog (locale-aware display).
    var date: CatalogDateValueInput? = nil
    var valueNameID: String
    /// name_values.form when listed (denormalized into valueText as well).
    var nameForm: String = ""
    /// Ordered name_value_parts when listed (empty when form-only).
    var nameParts: [CatalogNameValuePart] = []
    var valueSubjectID: String
    var valueTermID: String
    var propertyKey: String
    var propertyLabel: String
    var propertyValueType: String
    /// Product/user term key when `valueTermID` is set (empty when unknown).
    var valueTermKey: String = ""
}

struct CatalogProperty: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var description: String
    /// text | integer | date | name | subject | term
    var valueType: String
    /// subject_type_fields references; delete only at 0.
    var usedBy: Int = 0
}

struct CatalogPropertyTerm: Sendable, Equatable, Identifiable {
    var id: String
    var propertyID: String
    var key: String
    var origin: String
    var label: String
    var description: String
}

struct CatalogSubjectTypeField: Sendable, Equatable, Identifiable {
    var property: CatalogProperty
    var sortOrder: Int
    var locked: Bool

    var id: String { property.id }
}

struct CatalogSubjectTypePresentation: Sendable, Equatable, Identifiable {
    var typeKey: String
    var l10nKey: String
    var iconSymbol: String
    var inkToken: String
    var tintToken: String
    var chipToken: String
    var lineToken: String
    var edgeFromToken: String
    var edgeToToken: String
    var role: String
    var placeable: Bool
    var paletteSort: Int
    var requiresCitationAtCreate: Bool
    var label: String

    var id: String { typeKey }
}

struct CatalogConnectRule: Sendable, Equatable {
    var fromTypeKey: String
    var toTypeKey: String
    var bridgeTypeKey: String
    var edgePropertyKeys: [String]
    var disambiguation: String
    var refuse: Bool

    /// FakeStore and unit-test double of `subjectvocab.seedConnect`.
    /// Live connect reads `listConnectRules` only. When the Go registry changes, update this table in the same change.
    static let productMatrix: [CatalogConnectRule] = [
        CatalogConnectRule(
            fromTypeKey: "person", toTypeKey: "event",
            bridgeTypeKey: "participation", edgePropertyKeys: ["person", "event"],
            disambiguation: "role", refuse: false
        ),
        CatalogConnectRule(
            fromTypeKey: "event", toTypeKey: "person",
            bridgeTypeKey: "participation", edgePropertyKeys: ["person", "event"],
            disambiguation: "role", refuse: false
        ),
        CatalogConnectRule(
            fromTypeKey: "person", toTypeKey: "person",
            bridgeTypeKey: "relationship", edgePropertyKeys: ["person", "related_to"],
            disambiguation: "relationship_type", refuse: false
        ),
        CatalogConnectRule(
            fromTypeKey: "event", toTypeKey: "place",
            bridgeTypeKey: "location", edgePropertyKeys: ["event", "place"],
            disambiguation: "none", refuse: false
        ),
        CatalogConnectRule(
            fromTypeKey: "place", toTypeKey: "event",
            bridgeTypeKey: "location", edgePropertyKeys: ["event", "place"],
            disambiguation: "none", refuse: false
        ),
        CatalogConnectRule(
            fromTypeKey: "person", toTypeKey: "place",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
        CatalogConnectRule(
            fromTypeKey: "place", toTypeKey: "person",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
        CatalogConnectRule(
            fromTypeKey: "event", toTypeKey: "event",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
        CatalogConnectRule(
            fromTypeKey: "place", toTypeKey: "place",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
    ]

    static func match(from fromTypeKey: String, to toTypeKey: String, in rules: [CatalogConnectRule]) -> CatalogConnectRule {
        rules.first(where: { $0.fromTypeKey == fromTypeKey && $0.toTypeKey == toTypeKey })
            ?? CatalogConnectRule(
                fromTypeKey: fromTypeKey,
                toTypeKey: toTypeKey,
                bridgeTypeKey: "",
                edgePropertyKeys: [],
                disambiguation: "",
                refuse: true
            )
    }
}

struct CatalogCredibilityAssessment: Sendable, Equatable {
    var id: String
    var sourceID: String
    var gradeID: String
    var gradeKey: String
    var gradeLabel: String
    var argument: String
}

struct CatalogSourceType: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var description: String
    /// Closed design-system `type_*` key for evidence representation.
    var iconKey: String = PVMarkKey.defaultTypeMark.rawValue
    /// How many sources classify as this type. Deleting is only allowed at
    /// 0 — the engine refuses otherwise (`sourcetypes.in_use`). Only
    /// `listSourceTypes` and `updateSourceType` populate it.
    var usedBy: Int = 0
    /// How many metadata fields this type suggests — the list's third
    /// column, so browsing never fetches every type's join rows. Only
    /// `listSourceTypes` and `updateSourceType` populate it.
    var suggestedFieldCount: Int = 0
}

/// One `source_type_metadata_fields` join row: a field a type suggests, in
/// its stored order. Suggestions are not a schema — a source of the type may
/// leave any of them blank.
struct CatalogTypeSuggestion: Sendable, Equatable, Identifiable {
    var field: CatalogMetadataField
    var sortOrder: Int

    var id: String { field.id }
}

struct CatalogMetadataField: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var dataType: String
    var description: String
    /// How many sources already carry a value for this field. Deleting is
    /// only allowed at 0 — the engine refuses otherwise
    /// (`sourcefields.in_use`). Only `listMetadataFields` and
    /// `updateMetadataField` populate it.
    var usedBy: Int = 0
}

struct CatalogMetadataEntry: Sendable, Equatable, Identifiable {
    var id: String { field.id }
    var field: CatalogMetadataField
    var valueText: String
    var dateValueID: String
    /// Full structured components when `dateValueID` is set — lets the date
    /// editor rebuild its draft from the catalog instead of a session cache.
    var date: CatalogDateValueInput?
    var hasValue: Bool
    var suggested: Bool
    var sortOrder: Int32
}

struct CatalogSourceWorkspace: Sendable, Equatable {
    var source: CatalogSource
    var notes: [CatalogSourceNote]
    var metadata: [CatalogMetadataEntry]
    var artifacts: [CatalogArtifact]
    /// Nil when no assessment row (UI may display Standard without a row).
    var credibility: CatalogCredibilityAssessment?
}

struct CatalogDateValueInput: Sendable, Equatable {
    var kind: String
    var qualifier: String = ""
    var calendar: String = "gregorian"
    var startYear: Int32?
    var startMonth: Int32?
    var startDay: Int32?
    var startHour: Int32?
    var startMinute: Int32?
    var startSecond: Int32?
    var startMillisecond: Int32?
    var startTZ: String = ""
    var endYear: Int32?
    var endMonth: Int32?
    var endDay: Int32?
    var endHour: Int32?
    var endMinute: Int32?
    var endSecond: Int32?
    var endMillisecond: Int32?
    var endTZ: String = ""
    var phrase: String = ""
}

/// Origin split returned by `GetWorkspaceNavCounts` for vocabulary
/// destinations. `total` is always `seeded + user + plugin`.
struct WorkspaceNavOriginCounts: Sendable, Equatable {
    var total: Int
    var seeded: Int
    var user: Int
    var plugin: Int

    static let zero = WorkspaceNavOriginCounts(total: 0, seeded: 0, user: 0, plugin: 0)
}

/// Aggregate workspace chrome counts from one catalog open.
struct WorkspaceNavCounts: Sendable, Equatable {
    var sources: Int
    var sourceTypes: WorkspaceNavOriginCounts
    var sourceFields: WorkspaceNavOriginCounts
}

protocol GenealogyStore: Sendable {
    func installIdentity(identityDir: String) async throws -> InstallIdentity?
    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult
    func signOut(identityDir: String) async throws
    func activeProject(identityDir: String) async throws -> String?
    func listProjectUsers(projectDir: String) async throws -> [InstallIdentity]
    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult
    func removeActiveProject(identityDir: String) async throws
    /// Releases the held exclusive catalog session for `projectDir` (workspace leave).
    func closeCatalogSession(projectDir: String) async throws
    func projectInfo(projectDir: String) async throws -> ProjectInfo

    func listSources(projectDir: String) async throws -> [CatalogSource]
    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace
    func createSource(
        projectDir: String,
        userID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource
    func updateSource(
        projectDir: String,
        userID: String,
        sourceID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource
    /// Pin a raster Artifact as cover, or revert to the Source type icon (`type_icon`).
    /// Non-image Files (PDF, …) cannot be cover.
    func setSourceCover(
        projectDir: String,
        userID: String,
        sourceID: String,
        coverMode: String,
        primaryArtifactID: String
    ) async throws -> CatalogSource
    func addSourceNote(projectDir: String, userID: String, sourceID: String, body: String) async throws
        -> CatalogSourceNote
    func updateSourceNote(projectDir: String, userID: String, noteID: String, body: String) async throws
        -> CatalogSourceNote
    func deleteSourceNote(projectDir: String, userID: String, noteID: String) async throws
    /// Returns the refreshed workspace entry (structured `date` included) so
    /// callers can patch without refetching.
    func setSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String,
        valueText: String,
        date: CatalogDateValueInput?
    ) async throws -> CatalogMetadataEntry
    func clearSourceMetadata(projectDir: String, userID: String, sourceID: String, fieldID: String) async throws
    /// Permanently dismiss an unfilled type suggestion for this Source.
    /// Returns the updated workspace metadata list.
    func dismissSourceMetadataSuggestion(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String
    ) async throws -> [CatalogMetadataEntry]
    /// Persist display order for visible metadata rows (field ids in order).
    func reorderSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldIDs: [String]
    ) async throws -> [CatalogMetadataEntry]
    func createArtifact(
        projectDir: String,
        userID: String,
        sourceID: String,
        fileID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact
    func updateArtifact(
        projectDir: String,
        userID: String,
        artifactID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact
    func ingestArtifactFile(
        projectDir: String,
        userID: String,
        artifactID: String,
        path: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool)
    /// Lazily ensure a thumbnail derivative for a primary File; empty path when skipped.
    func ensureFileThumbnail(
        projectDir: String,
        fileID: String
    ) async throws -> (relPath: String, skipped: Bool)
    func listSourceCredibilityGrades(projectDir: String) async throws -> [CatalogCredibilityGrade]
    func upsertSourceCredibilityAssessment(
        projectDir: String,
        userID: String,
        sourceID: String,
        gradeID: String,
        argument: String
    ) async throws -> CatalogCredibilityAssessment
    func listSourceTypes(projectDir: String) async throws -> [CatalogSourceType]
    /// `key` is never accepted from the caller — the engine mints it as a
    /// kebab-case slug of `label`, the same rule `createMetadataField` uses.
    func createSourceType(
        projectDir: String,
        userID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType
    /// Patches label, description, and icon for a `user` or `provenencia` type.
    /// The key never changes here, so a rename keeps existing sources
    /// attached. Fails for `plugin:…` rows.
    func updateSourceType(
        projectDir: String,
        userID: String,
        typeID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType
    /// Deletes a type no source refers to. Suggestion joins cascade; the
    /// fields they named stay in the vocabulary.
    func deleteSourceType(
        projectDir: String,
        userID: String,
        typeID: String
    ) async throws
    func listTypeSuggestions(projectDir: String, typeID: String) async throws -> [CatalogTypeSuggestion]
    /// Attaches an existing field to a type at the end of its order.
    /// Assigning a field the type already suggests is a no-op. Returns the
    /// type's whole suggestion list so the caller never re-derives order.
    func assignTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion]
    /// Detaches a field from a type — the join only. The field stays in the
    /// vocabulary and sources already carrying a value for it keep it.
    func removeTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion]
    func listMetadataFields(projectDir: String) async throws -> [CatalogMetadataField]
    /// `key` is never accepted from the caller — the engine mints it as a
    /// kebab-case slug of `label` (see `FieldSlug.kebab` for the client-side
    /// preview mirror).
    func createMetadataField(
        projectDir: String,
        userID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField
    /// Patches label, data type, and description for a `user`-origin field.
    /// The key never changes here. Fails for `provenencia`/`plugin:…` rows.
    func updateMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField

    func deleteMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String
    ) async throws
    /// One catalog open: sidebar / vocabulary-header totals for sources,
    /// types, and fields.
    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts
    /// Omnibar catalog search (S3-07+). Empty / whitespace query → empty hits.
    func searchCatalog(
        projectDir: String,
        query: String,
        location: WorkspaceLocation
    ) async throws -> [CatalogSearchHit]

    func listSubjectTypes(projectDir: String) async throws -> [CatalogSubjectType]
    func createSubject(
        projectDir: String,
        userID: String,
        sourceID: String,
        subjectTypeID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject
    func updateSubject(
        projectDir: String,
        userID: String,
        subjectID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject
    func deleteSubject(projectDir: String, userID: String, subjectID: String) async throws
    func listSubjects(projectDir: String, sourceID: String) async throws -> [CatalogSubject]
    func setSubjectPosition(
        projectDir: String,
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async throws -> CatalogSubjectPosition
    func clearSubjectPosition(projectDir: String, subjectID: String) async throws
    func listSubjectPositions(projectDir: String, sourceID: String) async throws -> [CatalogSubjectPosition]

    func listProperties(projectDir: String) async throws -> [CatalogProperty]
    func createProperty(
        projectDir: String,
        userID: String,
        label: String,
        valueType: String,
        description: String
    ) async throws -> CatalogProperty
    func updateProperty(
        projectDir: String,
        userID: String,
        propertyID: String,
        label: String,
        valueType: String,
        description: String
    ) async throws -> CatalogProperty
    func deleteProperty(projectDir: String, userID: String, propertyID: String) async throws
    func listPropertyTerms(projectDir: String, propertyID: String) async throws -> [CatalogPropertyTerm]
    func createPropertyTerm(
        projectDir: String,
        userID: String,
        propertyID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm
    func updatePropertyTerm(
        projectDir: String,
        userID: String,
        termID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm
    func deletePropertyTerm(projectDir: String, userID: String, termID: String) async throws
    func listSubjectTypeFields(projectDir: String, subjectTypeID: String) async throws -> [CatalogSubjectTypeField]
    func assignSubjectTypeField(
        projectDir: String,
        userID: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws
    func removeSubjectTypeField(
        projectDir: String,
        userID: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws
    func listPlaceableSubjectTypes() async throws -> [CatalogSubjectTypePresentation]
    func getSubjectTypePresentation(typeKey: String) async throws -> CatalogSubjectTypePresentation
    func listConnectRules() async throws -> [CatalogConnectRule]

    func createCitedBridge(
        projectDir: String,
        userID: String,
        sourceID: String,
        fromSubjectID: String,
        toSubjectID: String,
        bridgeTypeKey: String,
        label: String,
        description: String,
        gridX: Int64,
        gridY: Int64,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        citationDescription: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations: [CatalogObservationDraft]
    ) async throws -> (CatalogSubject, CatalogCitation, [CatalogObservation])

    func createCitationWithObservations(
        projectDir: String,
        userID: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations: [CatalogObservationDraft]
    ) async throws -> (CatalogCitation, [CatalogObservation])

    func getCitation(
        projectDir: String,
        citationID: String
    ) async throws -> (CatalogCitation, [String], [CatalogObservation])

    func updateCitationWithObservations(
        projectDir: String,
        userID: String,
        citationID: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations: [CatalogObservationDraft]
    ) async throws -> (CatalogCitation, [CatalogObservation])

    func addObservationsToCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        observations: [CatalogObservationDraft]
    ) async throws -> [CatalogObservation]

    func listObservationsBySource(projectDir: String, sourceID: String) async throws -> [CatalogObservation]

    func citationCountsBySource(projectDir: String, sourceID: String) async throws -> [String: Int]

    func listCitationsByArtifact(projectDir: String, artifactID: String) async throws -> [CatalogListedCitation]
}

/// Draft payload for one Observation insert (FFI ObservationDraft).
struct CatalogObservationDraft: Sendable {
    var subjectID: String
    var propertyID: String
    var polarity: String = ""
    var valueText: String = ""
    var valueInteger: Int64?
    var date: CatalogDateValueInput?
    var valueDateID: String = ""
    var nameForm: String = ""
    var nameParts: [CatalogNameValuePart] = []
    var valueNameID: String = ""
    var valueSubjectID: String = ""
    var valueTermID: String = ""
    var notes: [String] = []
}
