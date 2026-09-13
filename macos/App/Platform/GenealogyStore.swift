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
    /// Thumbnail JPEG under `objects/…` for list cells (empty = no raster).
    var thumbnailRelPath: String = ""
    /// When `thumbnailRelPath` is empty, MIME of the cover-candidate File for glyphs.
    var thumbnailMediaType: String = ""
    var thumbnailOriginalFilename: String = ""
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
    var iconKey: String = PVEvidenceIconKey.defaultTypeIcon.rawValue
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
    /// Page vocabulary folded into the same catalog open so the Source page
    /// loads with one RPC (type picker, credibility chips, Add-metadata list).
    var types: [CatalogSourceType] = []
    var grades: [CatalogCredibilityGrade] = []
    var fields: [CatalogMetadataField] = []
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
    var files: Int
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
    /// Total content-addressed files rows — distinct files, not the
    /// (larger, per-source) artifact count. No project-wide artifact
    /// listing exists yet (S2-17).
    func countFiles(projectDir: String) async throws -> Int
    /// One catalog open: sidebar / vocabulary-header totals for sources,
    /// types, fields, and files.
    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts
}
