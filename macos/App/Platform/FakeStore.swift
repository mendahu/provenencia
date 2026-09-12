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
    /// Type↔field suggestion joins, keyed by source type id and held in the
    /// order they were assigned — the engine's `sort_order`.
    var suggestionsByType: [String: [CatalogTypeSuggestion]] = [:]
    var fieldsByProject: [String: [CatalogMetadataField]] = [:]
    var metadataBySource: [String: [CatalogMetadataEntry]] = [:]
    var fileCountByProject: [String: Int] = [:]
    /// Project dir for which a catalog RPC has “held” a session (tests only).
    var heldCatalogProjectDir: String?
    /// Last `closeCatalogSession` argument (tests only).
    var lastClosedCatalogProjectDir: String?
    /// When set, `listSources` throws instead of returning the in-memory list.
    var listSourcesError: Error?
    /// When set, `workspaceNavCounts` throws instead of returning counts.
    var workspaceNavCountsError: Error?
    /// When set, `updateSource` throws (identity title/type/description saves).
    var updateSourceError: Error?
    /// When set, `reorderSourceMetadata` throws (optimistic move should revert).
    var reorderSourceMetadataError: Error?
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
            updatedByRef: "USR-F4N2P"
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
            updatedByRef: ref
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
        let rows = sourcesByProject[projectDir] ?? []
        return rows.map { source in
            var copy = source
            if copy.thumbnailRelPath.isEmpty {
                let arts = artifactsBySource[source.id] ?? []
                if let raster = arts.first(where: { !$0.thumbnailRelPath.isEmpty }) {
                    copy.thumbnailRelPath = raster.thumbnailRelPath
                } else if let fileArt = arts.first(where: { $0.file != nil }) {
                    copy.thumbnailMediaType = fileArt.file?.mediaType ?? ""
                    copy.thumbnailOriginalFilename = fileArt.file?.originalFilename ?? ""
                }
            }
            return copy
        }
    }

    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace {
        markCatalogSessionHeld(projectDir)
        let source = (sourcesByProject[projectDir] ?? []).first { $0.id == sourceID }
            ?? CatalogSource(id: sourceID, ref: "SRC-XXXXX", sourceTypeID: "", title: "", description: "")
        return CatalogSourceWorkspace(
            source: source,
            notes: notesBySource[sourceID] ?? [],
            metadata: metadataBySource[sourceID] ?? [],
            artifacts: artifactsBySource[sourceID] ?? [],
            credibility: credibilityBySource[sourceID],
            types: sourceTypesByProject[projectDir] ?? [],
            grades: credibilityGradesByProject[projectDir] ?? [],
            fields: fieldsByProject[projectDir] ?? []
        )
    }

    func createSource(
        projectDir: String,
        userID _: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        let source = CatalogSource(
            id: UUID().uuidString.lowercased(),
            ref: "SRC-FAKE1",
            sourceTypeID: sourceTypeID,
            title: title,
            description: description
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
        sourcesByProject[projectDir] = list
        return list[idx]
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
        projectDir _: String,
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
        description: String
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
        let type = CatalogSourceType(
            id: UUID().uuidString.lowercased(),
            key: key,
            origin: "user",
            label: label,
            description: description
        )
        sourceTypesByProject[projectDir, default: []].append(type)
        return type
    }

    func updateSourceType(
        projectDir: String,
        userID _: String,
        typeID: String,
        label: String,
        description: String
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

    func countFiles(projectDir: String) async throws -> Int {
        markCatalogSessionHeld(projectDir)
        return fileCountByProject[projectDir] ?? 0
    }

    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts {
        markCatalogSessionHeld(projectDir)
        if let workspaceNavCountsError { throw workspaceNavCountsError }
        let types = sourceTypesByProject[projectDir] ?? []
        let fields = fieldsByProject[projectDir] ?? []
        return WorkspaceNavCounts(
            sources: (sourcesByProject[projectDir] ?? []).count,
            sourceTypes: Self.originCounts(from: types.map(\.origin)),
            sourceFields: Self.originCounts(from: fields.map(\.origin)),
            files: fileCountByProject[projectDir] ?? 0
        )
    }

    private func markCatalogSessionHeld(_ projectDir: String) {
        heldCatalogProjectDir = projectDir
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
