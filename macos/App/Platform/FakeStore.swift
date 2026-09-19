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
    var subjectTypeFieldsByType: [String: [CatalogSubjectTypeField]] = [:]
    var placeableSubjectTypes: [CatalogSubjectTypePresentation] = []
    var subjectTypePresentations: [String: CatalogSubjectTypePresentation] = [:]
    var connectRules: [CatalogConnectRule] = []
    var metadataBySource: [String: [CatalogMetadataEntry]] = [:]
    /// Project dir for which a catalog RPC has “held” a session (tests only).
    var heldCatalogProjectDir: String?
    /// Last `closeCatalogSession` argument (tests only).
    var lastClosedCatalogProjectDir: String?
    /// When set, `listSources` throws instead of returning the in-memory list.
    var listSourcesError: Error?
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
        valueText: String,
        date: CatalogDateValueInput?
    ) async throws -> CatalogMetadataEntry {
        var list = metadataBySource[sourceID] ?? []
        let dateID: String
        let stored: CatalogDateValueInput?
        if let date {
            dateID = "dv-\(fieldID.prefix(8))"
            stored = date
        } else if let existing = list.first(where: { $0.field.id == fieldID }) {
            // Text-only updates keep any existing structured DateValue.
            dateID = existing.dateValueID
            stored = existing.date
        } else {
            dateID = ""
            stored = nil
        }
        let entry: CatalogMetadataEntry
        if let idx = list.firstIndex(where: { $0.field.id == fieldID }) {
            list[idx].valueText = valueText
            list[idx].hasValue = true
            list[idx].dateValueID = dateID
            list[idx].date = stored
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
                dateValueID: dateID,
                date: stored,
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
        metadataBySource[sourceID] = (metadataBySource[sourceID] ?? []).filter { $0.field.id != fieldID }
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
            ? PVEvidenceIconKey.defaultTypeIcon.rawValue
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
        list[idx].iconKey = trimmed.isEmpty ? PVEvidenceIconKey.defaultTypeIcon.rawValue : trimmed
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
        markCatalogSessionHeld(projectDir)
        return subjectTypesByProject[projectDir] ?? []
    }

    func createSubject(
        projectDir: String,
        userID _: String,
        sourceID: String,
        subjectTypeID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject {
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
        return subject
    }

    func updateSubject(
        projectDir: String,
        userID _: String,
        subjectID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject {
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
        markCatalogSessionHeld(projectDir)
        return subjectsBySource[sourceID] ?? []
    }

    func setSubjectPosition(
        projectDir: String,
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async throws -> CatalogSubjectPosition {
        markCatalogSessionHeld(projectDir)
        if let setSubjectPositionError {
            throw setSubjectPositionError
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
        throw CoreInvokeError.coded(status: 1, code: "subjectvocab.invalid", kind: .user, params: [])
    }

    func listConnectRules() async throws -> [CatalogConnectRule] {
        connectRules
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
