#if DEBUG
import Foundation

/// In-memory `GenealogyStore` for SwiftUI previews and tests. Not used in the shipped app.
final class FakeStore: GenealogyStore, @unchecked Sendable {
    var identity: InstallIdentity?
    var activeProjectDir: String?
    var catalogUsers: [InstallIdentity]
    var projectInfos: [String: ProjectInfo] = [:]
    var sourcesByProject: [String: [CatalogSource]] = [:]
    var notesBySource: [String: [CatalogSourceNote]] = [:]
    var artifactsBySource: [String: [CatalogArtifact]] = [:]
    var credibilityBySource: [String: CatalogCredibilityAssessment] = [:]
    var credibilityGradesByProject: [String: [CatalogCredibilityGrade]] = [:]
    var sourceTypesByProject: [String: [CatalogSourceType]] = [:]
    var subjectTypesByProject: [String: [CatalogSubjectType]] = [:]
    var subjectsBySource: [String: [CatalogSubject]] = [:]
    var subjectPositionsBySubject: [String: CatalogSubjectPosition] = [:]
    /// Type↔field suggestion joins, keyed by source type id and held in the
    /// order they were assigned — the engine's `sort_order`.
    var suggestionsByType: [String: [CatalogTypeSuggestion]] = [:]
    var fieldsByProject: [String: [CatalogMetadataField]] = [:]
    var propertiesByProject: [String: [CatalogProperty]] = [:]
    var propertyTermsByProperty: [String: [CatalogPropertyTerm]] = [:]
    var subjectTypeFieldsByType: [String: [CatalogSubjectTypeField]] = [:]
    var placeableSubjectTypes: [CatalogSubjectTypePresentation] = []
    var subjectTypePresentations: [String: CatalogSubjectTypePresentation] = [:]
    var connectRules: [CatalogConnectRule] = CatalogConnectRule.productMatrix
    var observationsBySource: [String: [CatalogObservation]] = [:]
    var edgeObservationIDs: Set<String> = []
    var citationsByID: [String: CatalogCitation] = [:]
    var citationNotesByID: [String: [String]] = [:]
    var metadataBySource: [String: [CatalogMetadataEntry]] = [:]
    /// Project dir for which a catalog RPC has “held” a session (tests only).
    var heldCatalogProjectDir: String?
    /// Last `closeCatalogSession` argument (tests only).
    var lastClosedCatalogProjectDir: String?
    /// Ordered store calls for composer / graph tests.
    var recordedCalls: [String] = []
    /// When set, `listSources` throws instead of returning the in-memory list.
    var listSourcesError: Error?
    var listSubjectsCalls = 0
    var listSubjectTypesCalls = 0
    var getSourceWorkspaceCalls = 0
    var listSourceGraphProgressCalls = 0
    var getSourceGraphProgressCalls = 0
    /// Optional override for graph-progress rows (tests). Missing keys compute from subjects/observations.
    var graphProgressBySource: [String: SourceGraphProgress] = [:]
    var listSourceGraphProgressError: Error?
    var getSourceGraphProgressError: Error?
    /// Monotonic stand-in for audit_transactions.revision (Sources Updated sort).
    private var nextAuditRevision: Int64 = 1
    /// When set, `searchCatalog` throws (omnibar error UI).
    var searchCatalogError: Error?
    /// When set, `workspaceNavCounts` throws instead of returning counts.
    var workspaceNavCountsError: Error?
    /// When set, `updateSource` throws (identity title/type/description saves).
    var updateSourceError: Error?
    /// When set, `reorderSourceMetadata` throws (optimistic move should revert).
    var reorderSourceMetadataError: Error?
    /// When set, `setSubjectPosition` throws (Evidence graph drag should revert).
    var setSubjectPositionError: Error?
    /// Optional delay before each `setSubjectPosition` (generation-guard tests).
    var setSubjectPositionDelayNanoseconds: UInt64 = 0
    /// Per-call errors consumed in order; `nil` means that call succeeds.
    var setSubjectPositionErrors: [Error?] = []
    private var setSubjectPositionCallIndex = 0
    var listSubjectsError: Error?
    var listConnectRulesError: Error?
    var createPropertyTermError: Error?
    /// When set, `setSourceMetadata` throws (field vs page error tests).
    var setSourceMetadataError: Error?
    /// When set, `addSourceNote` throws (`pageError` surfacing).
    var addSourceNoteError: Error?
    /// When set, `ingestArtifactFile` throws before mutating artifacts.
    var ingestArtifactFileError: Error?
    /// When set, `createSourceType` throws instead of the built-in duplicate/slug checks.
    var createSourceTypeError: Error?
    var lastResult = OnboardingResult(
        projectDir: "/tmp/robins-family.provenencia",
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jake Robins",
        ref: "USR-F4N2P",
        project: ProjectInfo(
            label: "Robins Family",
            folderName: "robins-family.provenencia",
            createdAt: "2026-01-01T00:00:00Z",
            updatedAt: "2026-01-01T00:00:00Z",
            updatedByUserID: "00000000-0000-7000-8000-000000000001",
            updatedByDisplayName: "Jake Robins",
            updatedByRef: "USR-F4N2P",
            uuid: "00000000-0000-7000-8000-0000000000aa"
        )
    )

    init(
        identity: InstallIdentity? = nil,
        activeProjectDir: String? = nil,
        catalogUsers: [InstallIdentity] = [
            InstallIdentity(
                userID: "00000000-0000-7000-8000-000000000001",
                displayName: "Jake Robins",
                ref: "USR-F4N2P"
            )
        ]
    ) {
        self.identity = identity
        self.activeProjectDir = activeProjectDir
        self.catalogUsers = catalogUsers
    }

    func installIdentity(identityDir _: String) async throws -> InstallIdentity? {
        identity
    }

    func completeOnboarding(
        identityDir _: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        let folder = ProjectSlug.folderName(from: familyName)
        let projectDir = parentDir + "/" + folder
        let ref = identity?.ref ?? lastResult.ref
        let userID = identity?.userID ?? lastResult.userID
        identity = InstallIdentity(userID: userID, displayName: displayName, ref: ref)
        let now = ISO8601DateFormatter().string(from: Date())
        let info = ProjectInfo(
            label: familyName,
            folderName: folder,
            createdAt: now,
            updatedAt: now,
            updatedByUserID: userID,
            updatedByDisplayName: displayName,
            updatedByRef: ref,
            uuid: UUID().uuidString.lowercased()
        )
        let result = OnboardingResult(
            projectDir: projectDir,
            userID: userID,
            displayName: displayName,
            ref: ref,
            project: info
        )
        activeProjectDir = result.projectDir
        catalogUsers = [InstallIdentity(userID: userID, displayName: displayName, ref: ref)]
        projectInfos[projectDir] = info
        return result
    }

    func activeProject(identityDir _: String) async throws -> String? {
        activeProjectDir
    }

    func listProjectUsers(projectDir _: String) async throws -> [InstallIdentity] {
        catalogUsers
    }

    func openProject(
        identityDir _: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult {
        if !adoptUserID.isEmpty, let match = catalogUsers.first(where: { $0.userID == adoptUserID }) {
            identity = match
            activeProjectDir = projectDir
            let info = projectInfos[projectDir] ?? ProjectInfo(
                label: ProjectSlug.labelFromFolder(projectDir),
                folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
                createdAt: "",
                updatedAt: "",
                updatedByUserID: match.userID,
                updatedByDisplayName: match.displayName,
                updatedByRef: match.ref
            )
            return OnboardingResult(
                projectDir: projectDir,
                userID: match.userID,
                displayName: match.displayName,
                ref: match.ref,
                project: info
            )
        }
        if identity == nil {
            identity = InstallIdentity(userID: lastResult.userID, displayName: displayName, ref: lastResult.ref)
        }
        let name = identity?.displayName ?? displayName
        let ref = identity?.ref ?? lastResult.ref
        let userID = identity?.userID ?? lastResult.userID
        let info = projectInfos[projectDir] ?? ProjectInfo(
            label: ProjectSlug.labelFromFolder(projectDir),
            folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
            createdAt: "",
            updatedAt: "",
            updatedByUserID: userID,
            updatedByDisplayName: name,
            updatedByRef: ref
        )
        let result = OnboardingResult(
            projectDir: projectDir,
            userID: userID,
            displayName: name,
            ref: ref,
            project: info
        )
        activeProjectDir = projectDir
        projectInfos[projectDir] = info
        return result
    }

    func removeActiveProject(identityDir _: String) async throws {
        activeProjectDir = nil
    }

    func signOut(identityDir _: String) async throws {
        activeProjectDir = nil
        identity = nil
        heldCatalogProjectDir = nil
    }

    func closeCatalogSession(projectDir: String) async throws {
        lastClosedCatalogProjectDir = projectDir
        if heldCatalogProjectDir == projectDir {
            heldCatalogProjectDir = nil
        }
    }

    func projectInfo(projectDir: String) async throws -> ProjectInfo {
        if let info = projectInfos[projectDir] {
            return info
        }
        return ProjectInfo(
            label: ProjectSlug.labelFromFolder(projectDir),
            folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
            createdAt: "",
            updatedAt: "",
            updatedByUserID: "",
            updatedByDisplayName: "",
            updatedByRef: ""
        )
    }

    func listSources(projectDir: String) async throws -> [CatalogSource] {
        markCatalogSessionHeld(projectDir)
        if let listSourcesError { throw listSourcesError }
        // Match Go `sources.List`: newest-created-first (UUIDv7 / id DESC).
        let rows = (sourcesByProject[projectDir] ?? []).sorted { $0.id > $1.id }
        return rows.map { enrichCoverFields($0) }
    }

    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace {
        getSourceWorkspaceCalls += 1
        markCatalogSessionHeld(projectDir)
        let raw = (sourcesByProject[projectDir] ?? []).first { $0.id == sourceID }
            ?? CatalogSource(id: sourceID, ref: "SRC-XXXXX", sourceTypeID: "", title: "", description: "")
        return CatalogSourceWorkspace(
            source: enrichCoverFields(raw),
            notes: notesBySource[sourceID] ?? [],
            metadata: metadataBySource[sourceID] ?? [],
            artifacts: artifactsBySource[sourceID] ?? [],
            credibility: credibilityBySource[sourceID]
        )
    }

    func createSource(
        projectDir: String,
        userID _: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        let rev = nextAuditRevision
        nextAuditRevision += 1
        let source = CatalogSource(
            id: UUID().uuidString.lowercased(),
            ref: "SRC-FAKE1",
            sourceTypeID: sourceTypeID,
            title: title,
            description: description,
            updatedRevision: rev
        )
        sourcesByProject[projectDir, default: []].append(source)
        return source
    }

    func updateSource(
        projectDir: String,
        userID _: String,
        sourceID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        if let updateSourceError { throw updateSourceError }
        var list = sourcesByProject[projectDir] ?? []
        guard let idx = list.firstIndex(where: { $0.id == sourceID }) else {
            throw StoreBoom.boom
        }
        list[idx].sourceTypeID = sourceTypeID
        list[idx].title = title
        list[idx].description = description
        list[idx].updatedRevision = nextAuditRevision
        nextAuditRevision += 1
        sourcesByProject[projectDir] = list
        return enrichCoverFields(list[idx])
    }

    func setSourceCover(
        projectDir: String,
        userID _: String,
        sourceID: String,
        coverMode: String,
        primaryArtifactID: String
    ) async throws -> CatalogSource {
        var list = sourcesByProject[projectDir] ?? []
        guard let idx = list.firstIndex(where: { $0.id == sourceID }) else {
            throw StoreBoom.boom
        }
        switch coverMode {
        case "type_icon":
            list[idx].coverMode = "type_icon"
            list[idx].primaryArtifactID = ""
        case "artifact":
            let arts = artifactsBySource[sourceID] ?? []
            guard let art = arts.first(where: { $0.id == primaryArtifactID }),
                  !art.thumbnailRelPath.isEmpty
            else {
                throw StoreBoom.boom
            }
            list[idx].coverMode = "artifact"
            list[idx].primaryArtifactID = primaryArtifactID
        default:
            throw StoreBoom.boom
        }
        list[idx].updatedRevision = nextAuditRevision
        nextAuditRevision += 1
        sourcesByProject[projectDir] = list
        return enrichCoverFields(list[idx])
    }

    func addSourceNote(projectDir _: String, userID: String, sourceID: String, body: String) async throws
        -> CatalogSourceNote
    {
        if let addSourceNoteError { throw addSourceNoteError }
        let author = catalogUsers.first { $0.userID == userID }?.displayName
            ?? identity?.displayName
            ?? ""
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        let note = CatalogSourceNote(
            id: UUID().uuidString.lowercased(),
            sourceID: sourceID,
            body: body,
            authorDisplayName: author,
            createdAt: formatter.string(from: Date())
        )
        notesBySource[sourceID, default: []].append(note)
        return note
    }

    func updateSourceNote(projectDir _: String, userID _: String, noteID: String, body: String) async throws
        -> CatalogSourceNote
    {
        for (sourceID, notes) in notesBySource {
            if let idx = notes.firstIndex(where: { $0.id == noteID }) {
                var copy = notes
                copy[idx].body = body
                notesBySource[sourceID] = copy
                return copy[idx]
            }
        }
        throw StoreBoom.boom
    }

    func deleteSourceNote(projectDir _: String, userID _: String, noteID: String) async throws {
        for (sourceID, notes) in notesBySource {
            notesBySource[sourceID] = notes.filter { $0.id != noteID }
        }
    }

    func setSourceMetadata(
        projectDir: String,
        userID _: String,
        sourceID: String,
        fieldID: String,
        valueText: String
    ) async throws -> CatalogMetadataEntry {
        if let setSourceMetadataError { throw setSourceMetadataError }
        var list = metadataBySource[sourceID] ?? []
        let entry: CatalogMetadataEntry
        if let idx = list.firstIndex(where: { $0.field.id == fieldID }) {
            list[idx].valueText = valueText
            list[idx].hasValue = true
            entry = list[idx]
        } else {
            let field = (fieldsByProject[projectDir] ?? []).first { $0.id == fieldID }
                ?? CatalogMetadataField(
                    id: fieldID, key: "field", origin: "user", label: "Field", dataType: "text", description: ""
                )
            let order = Int32(list.count)
            entry = CatalogMetadataEntry(
                field: field,
                valueText: valueText,
                hasValue: true,
                suggested: false,
                sortOrder: order
            )
            list.append(entry)
        }
        metadataBySource[sourceID] = list
        return entry
    }

    func clearSourceMetadata(projectDir _: String, userID _: String, sourceID: String, fieldID: String) async throws {
        var list = metadataBySource[sourceID] ?? []
        guard let idx = list.firstIndex(where: { $0.field.id == fieldID }) else { return }
        if list[idx].suggested {
            list[idx].valueText = ""
            list[idx].hasValue = false
        } else {
            list.remove(at: idx)
        }
        metadataBySource[sourceID] = list
    }

    func dismissSourceMetadataSuggestion(
        projectDir _: String,
        userID _: String,
        sourceID: String,
        fieldID: String
    ) async throws -> [CatalogMetadataEntry] {
        var list = metadataBySource[sourceID] ?? []
        list.removeAll { $0.field.id == fieldID && !$0.hasValue }
        metadataBySource[sourceID] = list
        return list
    }

    func reorderSourceMetadata(
        projectDir _: String,
        userID _: String,
        sourceID: String,
        fieldIDs: [String]
    ) async throws -> [CatalogMetadataEntry] {
        if let reorderSourceMetadataError { throw reorderSourceMetadataError }
        let current = metadataBySource[sourceID] ?? []
        var byID = Dictionary(uniqueKeysWithValues: current.map { ($0.field.id, $0) })
        var next: [CatalogMetadataEntry] = []
        for (i, id) in fieldIDs.enumerated() {
            guard var entry = byID.removeValue(forKey: id) else { continue }
            entry.sortOrder = Int32(i)
            next.append(entry)
        }
        for (_, leftover) in byID {
            var entry = leftover
            entry.sortOrder = Int32(next.count)
            next.append(entry)
        }
        metadataBySource[sourceID] = next
        return next
    }

    func createArtifact(
        projectDir _: String,
        userID _: String,
        sourceID: String,
        fileID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact {
        let art = CatalogArtifact(
            id: UUID().uuidString.lowercased(),
            ref: "ART-FAKE1",
            sourceID: sourceID,
            fileID: fileID,
            label: label,
            description: description,
            file: nil
        )
        artifactsBySource[sourceID, default: []].append(art)
        return art
    }

    func updateArtifact(
        projectDir _: String,
        userID _: String,
        artifactID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact {
        for (sourceID, arts) in artifactsBySource {
            if let idx = arts.firstIndex(where: { $0.id == artifactID }) {
                var copy = arts
                copy[idx].label = label
                copy[idx].description = description
                artifactsBySource[sourceID] = copy
                return copy[idx]
            }
        }
        throw StoreBoom.boom
    }

    func ingestArtifactFile(
        projectDir: String,
        userID _: String,
        artifactID: String,
        path: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool) {
        if let ingestArtifactFileError { throw ingestArtifactFileError }
        let filename = URL(fileURLWithPath: path).lastPathComponent
        let ext = URL(fileURLWithPath: path).pathExtension.lowercased()
        let mediaType: String = switch ext {
        case "png": "image/png"
        case "jpg", "jpeg": "image/jpeg"
        case "gif": "image/gif"
        case "webp": "image/webp"
        case "tif", "tiff": "image/tiff"
        case "pdf": "application/pdf"
        case "mp4": "video/mp4"
        case "mov": "video/quicktime"
        case "mp3": "audio/mpeg"
        case "wav": "audio/wav"
        case "txt": "text/plain"
        case "csv": "text/csv"
        case "doc", "docx": "application/msword"
        default: "application/octet-stream"
        }
        let file = CatalogFileRef(
            id: UUID().uuidString.lowercased(),
            relPath: "objects/aa/bb/aabb",
            originalFilename: filename,
            mediaType: mediaType,
            byteSize: 0
        )
        for (sourceID, arts) in artifactsBySource {
            if let idx = arts.firstIndex(where: { $0.id == artifactID }) {
                if !arts[idx].fileID.isEmpty {
                    throw CoreInvokeError.coded(
                        status: 1,
                        code: "artifacts.file_already_attached",
                        kind: .conflict,
                        params: []
                    )
                }
                var copy = arts
                copy[idx].fileID = file.id
                copy[idx].file = file
                if ["png", "jpg", "jpeg", "gif", "webp", "tif", "tiff"].contains(ext) {
                    copy[idx].thumbnailRelPath = "objects/aa/bb/thumb-\(file.id.prefix(8))"
                }
                artifactsBySource[sourceID] = copy
                return (copy[idx], file, false)
            }
        }
        throw StoreBoom.boom
    }

    func ensureFileThumbnail(
        projectDir _: String,
        fileID: String
    ) async throws -> (relPath: String, skipped: Bool) {
        for arts in artifactsBySource.values {
            if let art = arts.first(where: { $0.fileID == fileID }) {
                if art.thumbnailRelPath.isEmpty {
                    return ("", true)
                }
                return (art.thumbnailRelPath, false)
            }
        }
        throw StoreBoom.boom
    }

    func listSourceCredibilityGrades(projectDir: String) async throws -> [CatalogCredibilityGrade] {
        if let grades = credibilityGradesByProject[projectDir], !grades.isEmpty {
            return grades
        }
        return [
            CatalogCredibilityGrade(id: "g-low", key: "low_trust", origin: "provenencia", label: "Low trust", sortOrder: 1),
            CatalogCredibilityGrade(id: "g-std", key: "standard", origin: "provenencia", label: "Standard", sortOrder: 2),
            CatalogCredibilityGrade(id: "g-high", key: "high_trust", origin: "provenencia", label: "High trust", sortOrder: 3),
        ]
    }

    func upsertSourceCredibilityAssessment(
        projectDir: String,
        userID _: String,
        sourceID: String,
        gradeID: String,
        argument: String
    ) async throws -> CatalogCredibilityAssessment {
        let grades = try await listSourceCredibilityGrades(projectDir: projectDir)
        let grade = grades.first { $0.id == gradeID } ?? grades[1]
        let assessment = CatalogCredibilityAssessment(
            id: credibilityBySource[sourceID]?.id ?? UUID().uuidString.lowercased(),
            sourceID: sourceID,
            gradeID: grade.id,
            gradeKey: grade.key,
            gradeLabel: grade.label,
            argument: argument
        )
        credibilityBySource[sourceID] = assessment
        return assessment
    }

    func listSourceTypes(projectDir: String) async throws -> [CatalogSourceType] {
        markCatalogSessionHeld(projectDir)
        return (sourceTypesByProject[projectDir] ?? []).map(withSuggestedFieldCount)
    }

    /// The engine derives this column in its list query, so the fake keeps it
    /// in step with `suggestionsByType` rather than making tests set it.
    private func withSuggestedFieldCount(_ type: CatalogSourceType) -> CatalogSourceType {
        var copy = type
        copy.suggestedFieldCount = (suggestionsByType[type.id] ?? []).count
        return copy
    }

    func createSourceType(
        projectDir: String,
        userID _: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        if let createSourceTypeError { throw createSourceTypeError }
        let key = FieldSlug.kebab(label)
        if key.isEmpty {
            throw StoreBoom.boom
        }
        if (sourceTypesByProject[projectDir] ?? []).contains(where: { $0.origin == "user" && $0.key == key }) {
            throw CoreInvokeError.coded(
                status: 1,
                code: "sourcetypes.duplicate_key",
                kind: .conflict,
                params: [key]
            )
        }
        let resolvedIcon = iconKey.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            ? PVMarkKey.defaultTypeMark.rawValue
            : iconKey
        let type = CatalogSourceType(
            id: UUID().uuidString.lowercased(),
            key: key,
            origin: "user",
            label: label,
            description: description,
            iconKey: resolvedIcon
        )
        sourceTypesByProject[projectDir, default: []].append(type)
        return type
    }

    func updateSourceType(
        projectDir: String,
        userID _: String,
        typeID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        var list = sourceTypesByProject[projectDir] ?? []
        guard let idx = list.firstIndex(where: { $0.id == typeID }) else {
            throw StoreBoom.boom
        }
        guard list[idx].origin == "user" || list[idx].origin == "provenencia" else {
            throw StoreBoom.boom
        }
        list[idx].label = label
        list[idx].description = description
        let trimmed = iconKey.trimmingCharacters(in: .whitespacesAndNewlines)
        list[idx].iconKey = trimmed.isEmpty ? PVMarkKey.defaultTypeMark.rawValue : trimmed
        sourceTypesByProject[projectDir] = list
        return withSuggestedFieldCount(list[idx])
    }

    func deleteSourceType(
        projectDir: String,
        userID _: String,
        typeID: String
    ) async throws {
        var list = sourceTypesByProject[projectDir] ?? []
        guard let idx = list.firstIndex(where: { $0.id == typeID }) else {
            throw StoreBoom.boom
        }
        // The engine refuses a type sources still classify as.
        guard list[idx].usedBy == 0 else {
            throw StoreBoom.boom
        }
        list.remove(at: idx)
        sourceTypesByProject[projectDir] = list
        // Suggestion joins cascade; the fields they named do not.
        suggestionsByType[typeID] = nil
    }

    func listTypeSuggestions(projectDir _: String, typeID: String) async throws -> [CatalogTypeSuggestion] {
        suggestionsByType[typeID] ?? []
    }

    func assignTypeField(
        projectDir: String,
        userID _: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        guard let field = (fieldsByProject[projectDir] ?? []).first(where: { $0.id == fieldID }) else {
            throw StoreBoom.boom
        }
        var list = suggestionsByType[typeID] ?? []
        // Assigning a field the type already suggests leaves its place alone.
        if !list.contains(where: { $0.field.id == fieldID }) {
            list.append(CatalogTypeSuggestion(field: field, sortOrder: (list.last?.sortOrder ?? -1) + 1))
            suggestionsByType[typeID] = list
        }
        return list
    }

    func removeTypeField(
        projectDir _: String,
        userID _: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        let list = (suggestionsByType[typeID] ?? []).filter { $0.field.id != fieldID }
        suggestionsByType[typeID] = list
        return list
    }

    func listMetadataFields(projectDir: String) async throws -> [CatalogMetadataField] {
        markCatalogSessionHeld(projectDir)
        return fieldsByProject[projectDir] ?? []
    }

    func createMetadataField(
        projectDir: String,
        userID _: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        let key = FieldSlug.kebab(label)
        if key.isEmpty {
            throw StoreBoom.boom
        }
        if (fieldsByProject[projectDir] ?? []).contains(where: { $0.origin == "user" && $0.key == key }) {
            throw CoreInvokeError.coded(
                status: 1,
                code: "sourcefields.duplicate_key",
                kind: .conflict,
                params: [key]
            )
        }
        let field = CatalogMetadataField(
            id: UUID().uuidString.lowercased(),
            key: key,
            origin: "user",
            label: label,
            dataType: dataType,
            description: description
        )
        fieldsByProject[projectDir, default: []].append(field)
        return field
    }

    func updateMetadataField(
        projectDir: String,
        userID _: String,
        fieldID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        var list = fieldsByProject[projectDir] ?? []
        guard let idx = list.firstIndex(where: { $0.id == fieldID }) else {
            throw StoreBoom.boom
        }
        guard list[idx].origin == "user" || list[idx].origin == "provenencia" else {
            throw StoreBoom.boom
        }
        guard list[idx].dataType == dataType else {
            throw StoreBoom.boom
        }
        list[idx].label = label
        list[idx].description = description
        fieldsByProject[projectDir] = list
        return list[idx]
    }

    func deleteMetadataField(
        projectDir: String,
        userID _: String,
        fieldID: String
    ) async throws {
        var list = fieldsByProject[projectDir] ?? []
        guard let idx = list.firstIndex(where: { $0.id == fieldID }) else {
            throw StoreBoom.boom
        }
        // The engine refuses a field that sources still reference.
        guard list[idx].usedBy == 0 else {
            throw StoreBoom.boom
        }
        list.remove(at: idx)
        fieldsByProject[projectDir] = list
    }

    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts {
        markCatalogSessionHeld(projectDir)
        if let workspaceNavCountsError { throw workspaceNavCountsError }
        let types = sourceTypesByProject[projectDir] ?? []
        let fields = fieldsByProject[projectDir] ?? []
        return WorkspaceNavCounts(
            sources: (sourcesByProject[projectDir] ?? []).count,
            sourceTypes: Self.originCounts(from: types.map(\.origin)),
            sourceFields: Self.originCounts(from: fields.map(\.origin))
        )
    }

    func searchCatalog(
        projectDir: String,
        query: String,
        location: WorkspaceLocation
    ) async throws -> [CatalogSearchHit] {
        markCatalogSessionHeld(projectDir)
        if let searchCatalogError { throw searchCatalogError }
        let raw = query.trimmingCharacters(in: .whitespacesAndNewlines)
        let tokens = Self.tokenizeSearch(raw)
        let refKey = Self.classifyRefQuery(raw)
        guard !tokens.isEmpty || refKey != nil else { return [] }

        var scored: [(hit: CatalogSearchHit, score: Double)] = []
        let types = sourceTypesByProject[projectDir] ?? []
        let typeLabelByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0.label) })
        let typeIconByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0.iconKey) })

        for source in sourcesByProject[projectDir] ?? [] {
            var description = source.description
            if let typeLabel = typeLabelByID[source.sourceTypeID], !typeLabel.isEmpty {
                description = "\(description) \(typeLabel)".trimmingCharacters(in: .whitespaces)
            }
            let values = [
                "title": source.title,
                "ref": source.ref,
                "description": description,
            ]
            var (score, reason) = Self.scoreSearchFields(
                fields: [("title", 10), ("ref", 12), ("description", 3)],
                values: values,
                tokens: tokens.isEmpty ? [raw.lowercased()] : tokens
            )
            var refBoost = 1.0
            if let refKey {
                let srcRef = source.ref.uppercased()
                if refKey.exact && srcRef == refKey.key {
                    score = max(score, 12)
                    reason = "ref"
                    refBoost = 8
                } else if !refKey.exact && srcRef.hasPrefix(refKey.key) {
                    score = max(score, 12)
                    reason = "ref"
                    refBoost = 3
                } else if score <= 0 {
                    continue
                }
            } else {
                guard score > 0 else { continue }
            }
            let boost = location.section == .sources ? 2.0 : 1.0
            scored.append((
                CatalogSearchHit(
                    kind: "source",
                    id: source.id,
                    ref: source.ref,
                    title: source.title,
                    subtitle: typeLabelByID[source.sourceTypeID] ?? "",
                    matchReason: reason,
                    location: WorkspaceLocation(
                        section: .sources,
                        sourceId: source.id,
                        ref: source.ref,
                        title: source.title
                    ),
                    thumbnailRelPath: source.thumbnailRelPath,
                    iconKey: typeIconByID[source.sourceTypeID] ?? ""
                ),
                score * boost * refBoost
            ))
        }

        // Ref-shaped queries only promote Sources (vocab has no SRC-… refs).
        if refKey == nil {
            for type in types {
                let (score, reason) = Self.scoreSearchFields(
                    fields: [("label", 10), ("key", 8), ("description", 3)],
                    values: [
                        "label": type.label,
                        "key": type.key,
                        "description": type.description,
                    ],
                    tokens: tokens
                )
                guard score > 0 else { continue }
                let boost = location.section == .sourceTypes ? 2.0 : 1.0
                scored.append((
                    CatalogSearchHit(
                        kind: "source_type",
                        id: type.id,
                        ref: "",
                        title: type.label,
                        subtitle: type.key,
                        matchReason: reason,
                        location: WorkspaceLocation(
                            section: .sourceTypes,
                            typeId: type.id,
                            title: type.label
                        ),
                        iconKey: type.iconKey
                    ),
                    score * boost
                ))
            }

            for field in fieldsByProject[projectDir] ?? [] {
                let (score, reason) = Self.scoreSearchFields(
                    fields: [("label", 10), ("key", 8), ("description", 3)],
                    values: [
                        "label": field.label,
                        "key": field.key,
                        "description": field.description,
                    ],
                    tokens: tokens
                )
                guard score > 0 else { continue }
                let boost = location.section == .sourceFields ? 2.0 : 1.0
                scored.append((
                    CatalogSearchHit(
                        kind: "source_field",
                        id: field.id,
                        ref: "",
                        title: field.label,
                        subtitle: field.key,
                        matchReason: reason,
                        location: WorkspaceLocation(
                            section: .sourceFields,
                            fieldId: field.id,
                            title: field.label
                        )
                    ),
                    score * boost
                ))
            }
        }

        scored.sort {
            if $0.score != $1.score { return $0.score > $1.score }
            if $0.hit.kind != $1.hit.kind { return $0.hit.kind < $1.hit.kind }
            return $0.hit.title < $1.hit.title
        }
        return scored.prefix(50).map(\.hit)
    }

    func listSubjectTypes(projectDir: String) async throws -> [CatalogSubjectType] {
        listSubjectTypesCalls += 1
        markCatalogSessionHeld(projectDir)
        return subjectTypesByProject[projectDir] ?? []
    }

    func createSubject(
        projectDir: String,
        userID _: String,
        sourceID: String,
        subjectTypeID: String,
        label: String,
        description: String,
        placement: CatalogGridCell?
    ) async throws -> CatalogSubject {
        recordedCalls.append("createSubject typeID=\(subjectTypeID) placement=\(placement.map { "\($0.gridX),\($0.gridY)" } ?? "nil")")
        markCatalogSessionHeld(projectDir)
        let subject = CatalogSubject(
            id: UUID().uuidString.lowercased(),
            ref: "CPR-FAKE1",
            sourceID: sourceID,
            subjectTypeID: subjectTypeID,
            label: label,
            description: description
        )
        subjectsBySource[sourceID, default: []].append(subject)
        if let placement {
            subjectPositionsBySubject[subject.id] = CatalogSubjectPosition(
                subjectID: subject.id,
                gridX: placement.gridX,
                gridY: placement.gridY
            )
        }
        return subject
    }

    func updateSubject(
        projectDir: String,
        userID _: String,
        subjectID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject {
        recordedCalls.append("updateSubject id=\(subjectID) label=\(label)")
        markCatalogSessionHeld(projectDir)
        for (sourceID, var list) in subjectsBySource {
            guard let idx = list.firstIndex(where: { $0.id == subjectID }) else { continue }
            list[idx].label = label
            list[idx].description = description
            subjectsBySource[sourceID] = list
            return list[idx]
        }
        throw StoreBoom.boom
    }

    func deleteSubject(projectDir: String, userID _: String, subjectID: String) async throws {
        markCatalogSessionHeld(projectDir)
        for list in observationsBySource.values {
            if list.contains(where: {
                $0.subjectID == subjectID || $0.valueSubjectID == subjectID
            }) {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "subjects.in_use",
                    kind: .conflict,
                    params: []
                )
            }
        }
        for (sourceID, var list) in subjectsBySource {
            guard let idx = list.firstIndex(where: { $0.id == subjectID }) else { continue }
            list.remove(at: idx)
            subjectsBySource[sourceID] = list
            subjectPositionsBySubject[subjectID] = nil
            return
        }
        throw StoreBoom.boom
    }

    func listSubjects(projectDir: String, sourceID: String) async throws -> [CatalogSubject] {
        listSubjectsCalls += 1
        markCatalogSessionHeld(projectDir)
        if let listSubjectsError {
            throw listSubjectsError
        }
        return subjectsBySource[sourceID] ?? []
    }

    func setSubjectPosition(
        projectDir: String,
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async throws -> CatalogSubjectPosition {
        markCatalogSessionHeld(projectDir)
        recordedCalls.append("setSubjectPosition subjectID=\(subjectID) \(gridX),\(gridY)")
        if setSubjectPositionDelayNanoseconds > 0 {
            try await Task.sleep(nanoseconds: setSubjectPositionDelayNanoseconds)
        }
        if setSubjectPositionCallIndex < setSubjectPositionErrors.count {
            let error = setSubjectPositionErrors[setSubjectPositionCallIndex]
            setSubjectPositionCallIndex += 1
            if let error { throw error }
        } else {
            setSubjectPositionCallIndex += 1
            if let setSubjectPositionError {
                throw setSubjectPositionError
            }
        }
        let position = CatalogSubjectPosition(subjectID: subjectID, gridX: gridX, gridY: gridY)
        subjectPositionsBySubject[subjectID] = position
        return position
    }

    func clearSubjectPosition(projectDir: String, subjectID: String) async throws {
        markCatalogSessionHeld(projectDir)
        subjectPositionsBySubject[subjectID] = nil
    }

    func listSubjectPositions(projectDir: String, sourceID: String) async throws -> [CatalogSubjectPosition] {
        markCatalogSessionHeld(projectDir)
        let subjects = subjectsBySource[sourceID] ?? []
        return subjects.compactMap { subjectPositionsBySubject[$0.id] }
            .sorted {
                if $0.gridY != $1.gridY { return $0.gridY < $1.gridY }
                if $0.gridX != $1.gridX { return $0.gridX < $1.gridX }
                return $0.subjectID < $1.subjectID
            }
    }

    func listProperties(projectDir: String) async throws -> [CatalogProperty] {
        markCatalogSessionHeld(projectDir)
        return propertiesByProject[projectDir] ?? []
    }

    func createProperty(
        projectDir: String,
        userID _: String,
        label: String,
        valueType: String,
        description: String
    ) async throws -> CatalogProperty {
        markCatalogSessionHeld(projectDir)
        let key = label
            .lowercased()
            .replacingOccurrences(of: " ", with: "-")
            .filter { $0.isLetter || $0.isNumber || $0 == "-" }
        let property = CatalogProperty(
            id: UUID().uuidString.lowercased(),
            key: key,
            origin: "user",
            label: label,
            description: description,
            valueType: valueType
        )
        propertiesByProject[projectDir, default: []].append(property)
        return property
    }

    func updateProperty(
        projectDir: String,
        userID _: String,
        propertyID: String,
        label: String,
        valueType _: String,
        description: String
    ) async throws -> CatalogProperty {
        markCatalogSessionHeld(projectDir)
        guard var list = propertiesByProject[projectDir],
              let idx = list.firstIndex(where: { $0.id == propertyID })
        else {
            throw CoreInvokeError.coded(status: 1, code: "properties.invalid", kind: .user, params: [])
        }
        list[idx].label = label
        list[idx].description = description
        propertiesByProject[projectDir] = list
        return list[idx]
    }

    func deleteProperty(projectDir: String, userID _: String, propertyID: String) async throws {
        markCatalogSessionHeld(projectDir)
        propertiesByProject[projectDir]?.removeAll { $0.id == propertyID }
        propertyTermsByProperty[propertyID] = nil
    }

    func listPropertyTerms(projectDir: String, propertyID: String) async throws -> [CatalogPropertyTerm] {
        markCatalogSessionHeld(projectDir)
        return propertyTermsByProperty[propertyID] ?? []
    }

    func createPropertyTerm(
        projectDir: String,
        userID _: String,
        propertyID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm {
        markCatalogSessionHeld(projectDir)
        if let createPropertyTermError {
            throw createPropertyTermError
        }
        let key = label
            .lowercased()
            .replacingOccurrences(of: " ", with: "-")
            .filter { $0.isLetter || $0.isNumber || $0 == "-" }
        let term = CatalogPropertyTerm(
            id: UUID().uuidString.lowercased(),
            propertyID: propertyID,
            key: key,
            origin: "user",
            label: label,
            description: description
        )
        propertyTermsByProperty[propertyID, default: []].append(term)
        return term
    }

    func updatePropertyTerm(
        projectDir: String,
        userID _: String,
        termID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm {
        markCatalogSessionHeld(projectDir)
        for (propertyID, var list) in propertyTermsByProperty {
            guard let idx = list.firstIndex(where: { $0.id == termID }) else { continue }
            if list[idx].origin != "user" {
                throw CoreInvokeError.coded(status: 1, code: "propertyterms.locked", kind: .user, params: [])
            }
            list[idx].label = label
            list[idx].description = description
            propertyTermsByProperty[propertyID] = list
            return list[idx]
        }
        throw CoreInvokeError.coded(status: 1, code: "propertyterms.invalid", kind: .user, params: [])
    }

    func deletePropertyTerm(projectDir: String, userID _: String, termID: String) async throws {
        markCatalogSessionHeld(projectDir)
        for (propertyID, list) in propertyTermsByProperty {
            guard let term = list.first(where: { $0.id == termID }) else { continue }
            if term.origin != "user" {
                throw CoreInvokeError.coded(status: 1, code: "propertyterms.locked", kind: .user, params: [])
            }
            propertyTermsByProperty[propertyID]?.removeAll { $0.id == termID }
            return
        }
        throw CoreInvokeError.coded(status: 1, code: "propertyterms.invalid", kind: .user, params: [])
    }

    func listSubjectTypeFields(projectDir: String, subjectTypeID: String) async throws -> [CatalogSubjectTypeField] {
        markCatalogSessionHeld(projectDir)
        return subjectTypeFieldsByType[subjectTypeID] ?? []
    }

    func assignSubjectTypeField(
        projectDir: String,
        userID _: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws {
        markCatalogSessionHeld(projectDir)
        let property = (propertiesByProject[projectDir] ?? []).first { $0.id == propertyID }
            ?? CatalogProperty(
                id: propertyID, key: "prop", origin: "user", label: "Prop",
                description: "", valueType: "text"
            )
        var list = subjectTypeFieldsByType[subjectTypeID] ?? []
        guard !list.contains(where: { $0.property.id == propertyID }) else { return }
        list.append(CatalogSubjectTypeField(property: property, sortOrder: list.count, locked: false))
        subjectTypeFieldsByType[subjectTypeID] = list
    }

    func removeSubjectTypeField(
        projectDir: String,
        userID _: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws {
        markCatalogSessionHeld(projectDir)
        if let field = subjectTypeFieldsByType[subjectTypeID]?.first(where: { $0.property.id == propertyID }),
           field.locked
        {
            throw CoreInvokeError.coded(status: 1, code: "subjectvocab.locked", kind: .conflict, params: [])
        }
        subjectTypeFieldsByType[subjectTypeID]?.removeAll { $0.property.id == propertyID }
    }

    func listPlaceableSubjectTypes() async throws -> [CatalogSubjectTypePresentation] {
        placeableSubjectTypes
    }

    func getSubjectTypePresentation(typeKey: String) async throws -> CatalogSubjectTypePresentation {
        if let found = subjectTypePresentations[typeKey] {
            return found
        }
        return Self.syntheticPresentation(typeKey: typeKey)
    }

    private static func syntheticPresentation(typeKey: String) -> CatalogSubjectTypePresentation {
        let role: String
        let sort: Int
        switch typeKey {
        case "person": role = "root"; sort = 0
        case "event": role = "root"; sort = 1
        case "place": role = "root"; sort = 2
        case "relationship": role = "bridge"; sort = 3
        case "participation": role = "bridge"; sort = 4
        case "location": role = "bridge"; sort = 5
        case "source": role = "reification"; sort = 6
        default: role = "root"; sort = 99
        }
        let ink: String
        switch typeKey {
        case "event", "participation": ink = "subjectEventInk"
        case "place", "location": ink = "subjectPlaceInk"
        default: ink = "subjectPersonInk"
        }
        let tint = ink.replacingOccurrences(of: "Ink", with: "Tint")
        let chip = ink.replacingOccurrences(of: "Ink", with: "Chip")
        let line = ink.replacingOccurrences(of: "Ink", with: "Line")
        return CatalogSubjectTypePresentation(
            typeKey: typeKey,
            l10nKey: "subjectType.\(typeKey)",
            iconSymbol: typeKey,
            inkToken: ink,
            tintToken: tint,
            chipToken: chip,
            lineToken: line,
            edgeFromToken: "",
            edgeToToken: "",
            role: role,
            placeable: role == "root",
            paletteSort: sort,
            requiresCitationAtCreate: role == "bridge",
            label: typeKey.replacingOccurrences(of: "_", with: " ").capitalized
        )
    }

    func listConnectRules() async throws -> [CatalogConnectRule] {
        if let listConnectRulesError {
            throw listConnectRulesError
        }
        return connectRules
    }

    func createCitedBridge(
        projectDir: String,
        userID: String,
        sourceID: String,
        fromSubjectID: String,
        toSubjectID: String,
        bridgeTypeKey: String,
        description: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        citationDescription: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations drafts: [CatalogObservationDraft],
        citationID: String?
    ) async throws -> (CatalogSubject, CatalogCitation, [CatalogObservation]) {
        recordedCalls.append(
            "createCitedBridge citationID=\(citationID ?? "nil") observations=\(drafts.count)"
        )
        markCatalogSessionHeld(projectDir)
        let from = subjectsBySource[sourceID]?.first(where: { $0.id == fromSubjectID })
        let to = subjectsBySource[sourceID]?.first(where: { $0.id == toSubjectID })
        guard let from, let to, from.id != to.id else {
            throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
        }
        let fromKey = subjectTypesByProject[projectDir]?.first(where: { $0.id == from.subjectTypeID })?.key ?? ""
        let toKey = subjectTypesByProject[projectDir]?.first(where: { $0.id == to.subjectTypeID })?.key ?? ""
        let rule = CatalogConnectRule.match(from: fromKey, to: toKey, in: connectRules)
        if rule.refuse {
            throw CoreInvokeError.coded(status: 1, code: "connect.refused", kind: .user, params: [])
        }
        if !bridgeTypeKey.isEmpty, bridgeTypeKey != rule.bridgeTypeKey {
            throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
        }
        if drafts.contains(where: { !$0.subjectID.isEmpty }) {
            throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
        }
        var expectedKeys = Set(rule.edges.map(\.propertyKey))
        if rule.disambiguation != "none", !rule.disambiguation.isEmpty {
            expectedKeys.insert(rule.disambiguation)
        }
        let properties = propertiesByProject[projectDir] ?? []
        let draftKeys = Set(drafts.compactMap { draft in
            properties.first { $0.id == draft.propertyID }?.key
        })
        if draftKeys != expectedKeys || drafts.count != expectedKeys.count {
            throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
        }
        if let citationID, !citationID.isEmpty {
            let fieldsUsed = !locatorJSON.isEmpty || !transcription.isEmpty
                || !citationDescription.isEmpty || transcriptionUncertain || !transcriptionNote.isEmpty
                || !citationNotes.isEmpty || !artifactID.isEmpty
            if fieldsUsed {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
        }
        guard let fromPos = subjectPositionsBySubject[from.id],
              let toPos = subjectPositionsBySubject[to.id]
        else {
            throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
        }
        guard let type = subjectTypesByProject[projectDir]?.first(where: { $0.key == rule.bridgeTypeKey }) else {
            throw StoreBoom.boom
        }
        let midX = Self.floorDiv(fromPos.gridX + toPos.gridX + 1, 2)
        let midY = Self.floorDiv(fromPos.gridY + toPos.gridY + 1, 2)
        let subject = try await createSubject(
            projectDir: projectDir,
            userID: userID,
            sourceID: sourceID,
            subjectTypeID: type.id,
            label: "",
            description: description,
            placement: CatalogGridCell(gridX: midX, gridY: midY)
        )
        var stamped = drafts
        for index in stamped.indices {
            stamped[index].subjectID = subject.id
        }
        let citation: CatalogCitation
        let observations: [CatalogObservation]
        if let citationID, !citationID.isEmpty {
            guard citationsByID[citationID] != nil else {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            citation = citationsByID[citationID]!
            observations = try await addObservationsToCitation(
                projectDir: projectDir,
                userID: userID,
                citationID: citationID,
                observations: stamped,
                allowEdgeRows: true
            )
        } else {
            (citation, observations) = try await createCitationWithObservations(
                projectDir: projectDir,
                userID: userID,
                artifactID: artifactID,
                locatorJSON: locatorJSON,
                transcription: transcription,
                description: citationDescription,
                transcriptionUncertain: transcriptionUncertain,
                transcriptionNote: transcriptionNote,
                citationNotes: citationNotes,
                observations: stamped
            )
        }
        let catalogProperties = propertiesByProject[projectDir] ?? []
        for obs in observations {
            let propertyKey = catalogProperties.first(where: { $0.id == obs.propertyID })?.key ?? obs.propertyKey
            if rule.edges.contains(where: { $0.propertyKey == propertyKey }) {
                edgeObservationIDs.insert(obs.id)
            }
        }
        return (subject, citation, observations)
    }

    private static func floorDiv(_ a: Int64, _ b: Int64) -> Int64 {
        guard b != 0 else { return 0 }
        var q = a / b
        if (a ^ b) < 0 && a % b != 0 {
            q -= 1
        }
        return q
    }

    private func isEdgeLocked(observation: CatalogObservation, projectDir: String) -> Bool {
        if edgeObservationIDs.contains(observation.id) { return true }
        let subject = subjectsBySource.values.flatMap { $0 }.first(where: { $0.id == observation.subjectID })
        let typeKey = subject.flatMap { subject in
            subjectTypesByProject[projectDir]?.first(where: { $0.id == subject.subjectTypeID })?.key
        } ?? ""
        let propertyKey = observation.propertyKey.isEmpty
            ? (propertiesByProject[projectDir]?.first(where: { $0.id == observation.propertyID })?.key ?? "")
            : observation.propertyKey
        return connectRules.contains { rule in
            !rule.refuse && rule.bridgeTypeKey == typeKey && rule.edges.contains { $0.propertyKey == propertyKey }
        }
    }

    private func edgeLockedError() -> CoreInvokeError {
        CoreInvokeError.coded(status: 1, code: "observations.edge_locked", kind: .conflict, params: [])
    }

    func createCitationWithObservations(
        projectDir: String,
        userID _: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations drafts: [CatalogObservationDraft]
    ) async throws -> (CatalogCitation, [CatalogObservation]) {
        recordedCalls.append("createCitationWithObservations observations=\(drafts.count)")
        markCatalogSessionHeld(projectDir)
        let citation = CatalogCitation(
            id: UUID().uuidString.lowercased(),
            ref: "CIT-FAKE1",
            artifactID: artifactID,
            locatorJSON: locatorJSON,
            transcription: transcription,
            description: description,
            transcriptionUncertain: transcriptionUncertain,
            transcriptionNote: transcriptionNote
        )
        citationsByID[citation.id] = citation
        citationNotesByID[citation.id] = citationNotes
        let created = try appendFakeObservations(
            projectDir: projectDir,
            citationID: citation.id,
            drafts: drafts
        )
        return (citation, created)
    }

    func getCitation(
        projectDir: String,
        citationID: String
    ) async throws -> (CatalogCitation, [String], [CatalogObservation]) {
        markCatalogSessionHeld(projectDir)
        guard let citation = citationsByID[citationID] else { throw StoreBoom.boom }
        let notes = citationNotesByID[citationID] ?? []
        let observations = observationsBySource.values
            .flatMap { $0 }
            .filter { $0.citationID == citationID }
        return (citation, notes, observations)
    }

    func addObservationsToCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        observations drafts: [CatalogObservationDraft]
    ) async throws -> [CatalogObservation] {
        try await addObservationsToCitation(
            projectDir: projectDir,
            userID: userID,
            citationID: citationID,
            observations: drafts,
            allowEdgeRows: false
        )
    }

    private func addObservationsToCitation(
        projectDir: String,
        userID _: String,
        citationID: String,
        observations drafts: [CatalogObservationDraft],
        allowEdgeRows: Bool
    ) async throws -> [CatalogObservation] {
        recordedCalls.append("addObservationsToCitation citationID=\(citationID) observations=\(drafts.count)")
        markCatalogSessionHeld(projectDir)
        let written = try appendFakeObservations(
            projectDir: projectDir,
            citationID: citationID,
            drafts: drafts
        )
        if !allowEdgeRows {
            for obs in written where isEdgeLocked(observation: obs, projectDir: projectDir) {
                throw edgeLockedError()
            }
        }
        return written
    }

    func updateCitation(
        projectDir: String,
        userID _: String,
        citationID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String
    ) async throws -> CatalogCitation {
        recordedCalls.append("updateCitation citationID=\(citationID)")
        markCatalogSessionHeld(projectDir)
        guard var citation = citationsByID[citationID] else { throw StoreBoom.boom }
        citation.locatorJSON = locatorJSON
        citation.transcription = transcription
        citation.description = description
        citation.transcriptionUncertain = transcriptionUncertain
        citation.transcriptionNote = transcriptionNote
        citationsByID[citationID] = citation
        return citation
    }

    func updateObservation(
        projectDir: String,
        userID _: String,
        observation: CatalogObservation
    ) async throws -> CatalogObservation {
        recordedCalls.append("updateObservation id=\(observation.id)")
        markCatalogSessionHeld(projectDir)
        let existing = observationsBySource.values.flatMap { $0 }.first(where: { $0.id == observation.id })
        guard let existing else { throw StoreBoom.boom }
        if isEdgeLocked(observation: existing, projectDir: projectDir)
            || isEdgeLocked(observation: observation, projectDir: projectDir)
        {
            throw edgeLockedError()
        }
        let next = try fakeObservation(
            citationID: existing.citationID,
            id: existing.id,
            ref: existing.ref,
            row: observation
        )
        for (sourceID, list) in observationsBySource {
            observationsBySource[sourceID] = list.map { $0.id == next.id ? next : $0 }
        }
        return next
    }

    func deleteObservation(projectDir: String, userID _: String, observationID: String) async throws {
        recordedCalls.append("deleteObservation id=\(observationID)")
        markCatalogSessionHeld(projectDir)
        let existing = observationsBySource.values.flatMap { $0 }.first(where: { $0.id == observationID })
        guard let existing else { throw StoreBoom.boom }
        if isEdgeLocked(observation: existing, projectDir: projectDir) {
            throw edgeLockedError()
        }
        for (sourceID, list) in observationsBySource {
            observationsBySource[sourceID] = list.filter { $0.id != observationID }
        }
    }

    func getSubjectFieldsWorkspace(projectDir: String) async throws -> SubjectFieldsSnapshot {
        markCatalogSessionHeld(projectDir)
        let types = subjectTypesByProject[projectDir] ?? []
        var fieldsByTypeID: [String: [CatalogSubjectTypeField]] = [:]
        var presentationsByKey: [String: CatalogSubjectTypePresentation] = [:]
        for type in types {
            fieldsByTypeID[type.id] = subjectTypeFieldsByType[type.id] ?? []
            if let presentation = subjectTypePresentations[type.key] {
                presentationsByKey[type.key] = presentation
            }
        }
        return SubjectFieldsSnapshot(
            properties: propertiesByProject[projectDir] ?? [],
            types: types,
            fieldsByTypeID: fieldsByTypeID,
            presentationsByKey: presentationsByKey
        )
    }

    func listObservationsBySource(projectDir: String, sourceID: String) async throws -> [CatalogObservation] {
        markCatalogSessionHeld(projectDir)
        return observationsBySource[sourceID] ?? []
    }

    func citationCountsBySource(projectDir: String, sourceID: String) async throws -> [String: Int] {
        markCatalogSessionHeld(projectDir)
        let artifactIDs = Set((artifactsBySource[sourceID] ?? []).map(\.id))
        var counts: [String: Int] = [:]
        for citation in citationsByID.values where artifactIDs.contains(citation.artifactID) {
            counts[citation.artifactID, default: 0] += 1
        }
        return counts
    }

    func listSourceGraphProgress(projectDir: String) async throws -> [SourceGraphProgress] {
        markCatalogSessionHeld(projectDir)
        listSourceGraphProgressCalls += 1
        if let listSourceGraphProgressError { throw listSourceGraphProgressError }
        let ids = Set((sourcesByProject[projectDir] ?? []).map(\.id) + subjectsBySource.keys + graphProgressBySource.keys)
        return ids.compactMap { id in
            let row = graphProgress(for: id, projectDir: projectDir)
            return row.isZero ? nil : row
        }
    }

    func getSourceGraphProgress(projectDir: String, sourceID: String) async throws -> SourceGraphProgress {
        markCatalogSessionHeld(projectDir)
        getSourceGraphProgressCalls += 1
        if let getSourceGraphProgressError { throw getSourceGraphProgressError }
        return graphProgress(for: sourceID, projectDir: projectDir)
    }

    private func graphProgress(for sourceID: String, projectDir: String) -> SourceGraphProgress {
        if let override = graphProgressBySource[sourceID] {
            return override
        }
        let sourceTypeIDs = Set(
            (subjectTypesByProject[projectDir] ?? []).filter { $0.key == "source" }.map(\.id)
        )
        let subjects = (subjectsBySource[sourceID] ?? []).filter { !sourceTypeIDs.contains($0.subjectTypeID) }
        let observations = observationsBySource[sourceID] ?? []
        let cited = Set(observations.map(\.subjectID))
        return SourceGraphProgress(
            sourceId: sourceID,
            subjectCount: subjects.count,
            observationCount: observations.count,
            uncitedCount: subjects.filter { !cited.contains($0.id) }.count
        )
    }

    func listCitationsByArtifact(projectDir: String, artifactID: String) async throws -> [CatalogListedCitation] {
        markCatalogSessionHeld(projectDir)
        let listed = citationsByID.values
            .filter { $0.artifactID == artifactID }
            .sorted { $0.ref.localizedCaseInsensitiveCompare($1.ref) == .orderedAscending }
        return listed.map { citation in
            let count = observationsBySource.values
                .flatMap { $0 }
                .filter { $0.citationID == citation.id }
                .count
            return CatalogListedCitation(citation: citation, observationCount: count)
        }
    }

    private func fakeObservation(
        citationID: String,
        id: String,
        ref: String,
        row: CatalogObservation
    ) throws -> CatalogObservation {
        guard subjectsBySource.values.flatMap({ $0 }).contains(where: { $0.id == row.subjectID }) else {
            throw StoreBoom.boom
        }
        let property = propertiesByProject.values.flatMap { $0 }.first(where: { $0.id == row.propertyID })
        var displayText = row.valueText
        if displayText.isEmpty, !row.valueTermID.isEmpty {
            displayText = propertyTermsByProperty[row.propertyID]?
                .first(where: { $0.id == row.valueTermID })?
                .label ?? ""
        }
        if displayText.isEmpty, !row.nameForm.isEmpty {
            displayText = row.nameForm
        }
        if displayText.isEmpty, let date = row.date {
            displayText = DateValueDisplay.string(for: date)
        }
        if displayText.isEmpty, !row.valueSubjectID.isEmpty {
            displayText = subjectsBySource.values
                .flatMap { $0 }
                .first(where: { $0.id == row.valueSubjectID })?
                .label ?? ""
        }
        return CatalogObservation(
            id: id,
            ref: ref,
            citationID: citationID,
            subjectID: row.subjectID,
            propertyID: row.propertyID,
            polarity: row.polarity.isEmpty ? "positive" : row.polarity,
            valueText: displayText,
            valueInteger: row.valueInteger,
            valueDateID: row.valueDateID,
            date: row.date,
            valueNameID: row.valueNameID,
            nameForm: row.nameForm,
            nameParts: row.nameParts,
            valueSubjectID: row.valueSubjectID,
            valueTermID: row.valueTermID,
            propertyKey: row.propertyKey.isEmpty ? (property?.key ?? "") : row.propertyKey,
            propertyLabel: row.propertyLabel.isEmpty ? (property?.label ?? "") : row.propertyLabel,
            propertyValueType: row.propertyValueType.isEmpty ? (property?.valueType ?? "") : row.propertyValueType,
            valueTermKey: row.valueTermKey.isEmpty
                ? (propertyTermsByProperty[row.propertyID]?.first(where: { $0.id == row.valueTermID })?.key ?? "")
                : row.valueTermKey
        )
    }

    private func appendFakeObservations(
        projectDir: String,
        citationID: String,
        drafts: [CatalogObservationDraft]
    ) throws -> [CatalogObservation] {
        if drafts.isEmpty {
            return []
        }
        var created: [CatalogObservation] = []
        for draft in drafts {
            guard let subject = subjectsBySource.values.flatMap({ $0 }).first(where: { $0.id == draft.subjectID })
            else {
                throw StoreBoom.boom
            }
            let property = propertiesByProject[projectDir]?.first(where: { $0.id == draft.propertyID })
            var displayText = draft.valueText
            if displayText.isEmpty, !draft.valueTermID.isEmpty {
                displayText = propertyTermsByProperty[draft.propertyID]?
                    .first(where: { $0.id == draft.valueTermID })?
                    .label ?? ""
            }
            if displayText.isEmpty, !draft.nameForm.isEmpty {
                displayText = draft.nameForm
            }
            if displayText.isEmpty, let date = draft.date {
                displayText = DateValueDisplay.string(for: date)
            }
            if displayText.isEmpty, !draft.valueSubjectID.isEmpty {
                displayText = subjectsBySource.values
                    .flatMap { $0 }
                    .first(where: { $0.id == draft.valueSubjectID })?
                    .label ?? ""
            }
            let obs = CatalogObservation(
                id: UUID().uuidString.lowercased(),
                ref: "OBS-FAKE1",
                citationID: citationID,
                subjectID: draft.subjectID,
                propertyID: draft.propertyID,
                polarity: draft.polarity.isEmpty ? "positive" : draft.polarity,
                valueText: displayText,
                valueInteger: draft.valueInteger,
                valueDateID: draft.valueDateID,
                date: draft.date,
                valueNameID: draft.valueNameID,
                nameForm: draft.nameForm,
                nameParts: draft.nameParts,
                valueSubjectID: draft.valueSubjectID,
                valueTermID: draft.valueTermID,
                propertyKey: property?.key ?? "",
                propertyLabel: property?.label ?? "",
                propertyValueType: property?.valueType ?? "",
                valueTermKey: propertyTermsByProperty[draft.propertyID]?
                    .first(where: { $0.id == draft.valueTermID })?
                    .key ?? ""
            )
            observationsBySource[subject.sourceID, default: []].append(obs)
            created.append(obs)
        }
        return created
    }

    private func markCatalogSessionHeld(_ projectDir: String) {
        heldCatalogProjectDir = projectDir
    }

    /// Exact catalog ref or PREFIX-token prefix (mirrors core/search classifyRefQuery).
    private static func classifyRefQuery(_ query: String) -> (key: String, exact: Bool)? {
        let s = query.trimmingCharacters(in: .whitespacesAndNewlines).uppercased()
        guard !s.isEmpty else { return nil }
        let exactPattern = /^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{5}$/
        if s.wholeMatch(of: exactPattern) != nil {
            return (s, true)
        }
        let prefixPattern = /^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{1,4}$/
        if s.wholeMatch(of: prefixPattern) != nil {
            return (s, false)
        }
        return nil
    }

    private static func tokenizeSearch(_ query: String) -> [String] {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !trimmed.isEmpty else { return [] }
        return trimmed.split(whereSeparator: \.isWhitespace).compactMap { part in
            let cleaned = part.trimmingCharacters(in: CharacterSet(charactersIn: "\"'.,;:!?()[]{}"))
            return cleaned.isEmpty ? nil : String(cleaned)
        }
    }

    private static func scoreSearchFields(
        fields: [(name: String, weight: Double)],
        values: [String: String],
        tokens: [String]
    ) -> (score: Double, reason: String) {
        var score = 0.0
        var bestField = ""
        var bestWeight = 0.0
        var matchedTokens = 0
        for tok in tokens {
            var tokHit = false
            for field in fields {
                let value = (values[field.name] ?? "").lowercased()
                guard !value.isEmpty, value.contains(tok) else { continue }
                tokHit = true
                score += field.weight
                if field.weight > bestWeight || (field.weight == bestWeight && bestField.isEmpty) {
                    bestWeight = field.weight
                    bestField = field.name
                }
            }
            if tokHit { matchedTokens += 1 }
        }
        guard matchedTokens > 0 else { return (0, "") }
        score *= Double(matchedTokens) / Double(tokens.count)
        return (score, bestField)
    }

    /// Resolves list/identity paint fields from persisted cover mode.
    /// Source cover is type icon or a raster path — never Artifact file-type MIME.
    /// Also sets `hasArtifact` from the Artifact bag (fileless counts).
    private func enrichCoverFields(_ source: CatalogSource) -> CatalogSource {
        var copy = source
        let arts = artifactsBySource[copy.id] ?? []
        copy.hasArtifact = !arts.isEmpty
        copy.thumbnailRelPath = ""
        copy.thumbnailMediaType = ""
        copy.thumbnailOriginalFilename = ""
        guard copy.coverMode == "artifact", !copy.primaryArtifactID.isEmpty else {
            return copy
        }
        guard let art = arts.first(where: { $0.id == copy.primaryArtifactID }),
              !art.thumbnailRelPath.isEmpty
        else {
            return copy
        }
        copy.thumbnailRelPath = art.thumbnailRelPath
        return copy
    }

    private static func originCounts(from origins: [String]) -> WorkspaceNavOriginCounts {
        var seeded = 0
        var user = 0
        var plugin = 0
        for origin in origins {
            switch origin {
            case CatalogOrigin.provenencia: seeded += 1
            case CatalogOrigin.user: user += 1
            default: plugin += 1
            }
        }
        return WorkspaceNavOriginCounts(
            total: origins.count,
            seeded: seeded,
            user: user,
            plugin: plugin
        )
    }
}

private enum StoreBoom: Error { case boom }
#endif
