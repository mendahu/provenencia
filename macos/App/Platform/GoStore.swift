import Foundation

struct GoStore: GenealogyStore {
    func installIdentity(identityDir: String) async throws -> InstallIdentity? {
        var req = Provenencia_Engine_V1_GetInstallIdentityRequest()
        req.identityDir = identityDir
        let resp: Provenencia_Engine_V1_GetInstallIdentityResponse = try await provenenciaCall(
            method: CoreMethod.getInstallIdentity,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return InstallIdentity(userID: resp.userID, displayName: resp.displayName, ref: resp.ref)
    }

    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        var req = Provenencia_Engine_V1_CompleteOnboardingRequest()
        req.identityDir = identityDir
        req.parentDir = parentDir
        req.displayName = displayName
        req.familyName = familyName
        let resp: Provenencia_Engine_V1_CompleteOnboardingResponse = try await provenenciaCall(
            method: CoreMethod.completeOnboarding,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName,
            ref: resp.ref,
            project: Self.mapProject(resp.project)
        )
    }

    func activeProject(identityDir: String) async throws -> String? {
        var req = Provenencia_Engine_V1_GetActiveProjectRequest()
        req.identityDir = identityDir
        let resp: Provenencia_Engine_V1_GetActiveProjectResponse = try await provenenciaCall(
            method: CoreMethod.getActiveProject,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return resp.projectDir
    }

    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult {
        var req = Provenencia_Engine_V1_OpenProjectRequest()
        req.identityDir = identityDir
        req.projectDir = projectDir
        req.displayName = displayName
        req.adoptUserID = adoptUserID
        let resp: Provenencia_Engine_V1_OpenProjectResponse = try await provenenciaCall(
            method: CoreMethod.openProject,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName,
            ref: resp.ref,
            project: Self.mapProject(resp.project)
        )
    }

    func listProjectUsers(projectDir: String) async throws -> [InstallIdentity] {
        var req = Provenencia_Engine_V1_ListProjectUsersRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListProjectUsersResponse = try await provenenciaCall(
            method: CoreMethod.listProjectUsers,
            request: req
        )
        return resp.users.map {
            InstallIdentity(userID: $0.userID, displayName: $0.displayName, ref: $0.ref)
        }
    }

    func removeActiveProject(identityDir: String) async throws {
        var req = Provenencia_Engine_V1_RemoveActiveProjectRequest()
        req.identityDir = identityDir
        let _: Provenencia_Engine_V1_RemoveActiveProjectResponse = try await provenenciaCall(
            method: CoreMethod.removeActiveProject,
            request: req
        )
    }

    func signOut(identityDir: String) async throws {
        var req = Provenencia_Engine_V1_SignOutRequest()
        req.identityDir = identityDir
        let _: Provenencia_Engine_V1_SignOutResponse = try await provenenciaCall(
            method: CoreMethod.signOut,
            request: req
        )
    }

    func closeCatalogSession(projectDir: String) async throws {
        var req = Provenencia_Engine_V1_CloseCatalogSessionRequest()
        req.projectDir = projectDir
        let _: Provenencia_Engine_V1_CloseCatalogSessionResponse = try await provenenciaCall(
            method: CoreMethod.closeCatalogSession,
            request: req
        )
    }

    func projectInfo(projectDir: String) async throws -> ProjectInfo {
        var req = Provenencia_Engine_V1_GetProjectInfoRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_GetProjectInfoResponse = try await provenenciaCall(
            method: CoreMethod.getProjectInfo,
            request: req
        )
        return Self.mapProject(resp.project)
    }

    func listSources(projectDir: String) async throws -> [CatalogSource] {
        var req = Provenencia_Engine_V1_ListSourcesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourcesResponse = try await provenenciaCall(
            method: CoreMethod.listSources,
            request: req
        )
        return resp.sources.map(Self.mapSource)
    }

    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace {
        var req = Provenencia_Engine_V1_GetSourceWorkspaceRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_GetSourceWorkspaceResponse = try await provenenciaCall(
            method: CoreMethod.getSourceWorkspace,
            request: req
        )
        return CatalogSourceWorkspace(
            source: Self.mapSource(resp.source),
            notes: resp.notes.map(Self.mapNote),
            metadata: resp.metadata.map(Self.mapMetadataEntry),
            artifacts: resp.artifacts.map(Self.mapArtifact),
            credibility: resp.hasCredibility ? Self.mapCredibility(resp.credibility) : nil,
            types: resp.types.map(Self.mapSourceType),
            grades: resp.grades.map(Self.mapCredibilityGrade),
            fields: resp.fields.map(Self.mapMetadataField)
        )
    }

    func createSource(
        projectDir: String,
        userID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        var req = Provenencia_Engine_V1_CreateSourceRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceTypeID = sourceTypeID
        req.title = title
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreateSourceResponse = try await provenenciaCall(
            method: CoreMethod.createSource,
            request: req
        )
        return Self.mapSource(resp.source)
    }

    func updateSource(
        projectDir: String,
        userID: String,
        sourceID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        var req = Provenencia_Engine_V1_UpdateSourceRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.sourceTypeID = sourceTypeID
        req.title = title
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateSourceResponse = try await provenenciaCall(
            method: CoreMethod.updateSource,
            request: req
        )
        return Self.mapSource(resp.source)
    }

    func setSourceCover(
        projectDir: String,
        userID: String,
        sourceID: String,
        coverMode: String,
        primaryArtifactID: String
    ) async throws -> CatalogSource {
        var req = Provenencia_Engine_V1_SetSourceCoverRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.coverMode = coverMode
        req.primaryArtifactID = primaryArtifactID
        let resp: Provenencia_Engine_V1_SetSourceCoverResponse = try await provenenciaCall(
            method: CoreMethod.setSourceCover,
            request: req
        )
        return Self.mapSource(resp.source)
    }

    func addSourceNote(projectDir: String, userID: String, sourceID: String, body: String) async throws
        -> CatalogSourceNote
    {
        var req = Provenencia_Engine_V1_AddSourceNoteRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.body = body
        let resp: Provenencia_Engine_V1_AddSourceNoteResponse = try await provenenciaCall(
            method: CoreMethod.addSourceNote,
            request: req
        )
        return Self.mapNote(resp.note)
    }

    func updateSourceNote(projectDir: String, userID: String, noteID: String, body: String) async throws
        -> CatalogSourceNote
    {
        var req = Provenencia_Engine_V1_UpdateSourceNoteRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.noteID = noteID
        req.body = body
        let resp: Provenencia_Engine_V1_UpdateSourceNoteResponse = try await provenenciaCall(
            method: CoreMethod.updateSourceNote,
            request: req
        )
        return Self.mapNote(resp.note)
    }

    func deleteSourceNote(projectDir: String, userID: String, noteID: String) async throws {
        var req = Provenencia_Engine_V1_DeleteSourceNoteRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.noteID = noteID
        let _: Provenencia_Engine_V1_DeleteSourceNoteResponse = try await provenenciaCall(
            method: CoreMethod.deleteSourceNote,
            request: req
        )
    }

    func setSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String,
        valueText: String,
        date: CatalogDateValueInput?
    ) async throws -> CatalogMetadataEntry {
        var req = Provenencia_Engine_V1_SetSourceMetadataRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldID = fieldID
        req.valueText = valueText
        if let date {
            var d = Provenencia_Engine_V1_DateValueInput()
            d.kind = date.kind
            d.qualifier = date.qualifier
            d.calendar = date.calendar
            d.phrase = date.phrase
            d.startTz = date.startTZ
            d.endTz = date.endTZ
            if let y = date.startYear { d.startYear = y }
            if let m = date.startMonth { d.startMonth = m }
            if let day = date.startDay { d.startDay = day }
            if let h = date.startHour { d.startHour = h }
            if let mi = date.startMinute { d.startMinute = mi }
            if let s = date.startSecond { d.startSecond = s }
            if let ms = date.startMillisecond { d.startMillisecond = ms }
            if let y = date.endYear { d.endYear = y }
            if let m = date.endMonth { d.endMonth = m }
            if let day = date.endDay { d.endDay = day }
            if let h = date.endHour { d.endHour = h }
            if let mi = date.endMinute { d.endMinute = mi }
            if let s = date.endSecond { d.endSecond = s }
            if let ms = date.endMillisecond { d.endMillisecond = ms }
            req.date = d
        }
        let resp: Provenencia_Engine_V1_SetSourceMetadataResponse = try await provenenciaCall(
            method: CoreMethod.setSourceMetadata,
            request: req
        )
        return Self.mapMetadataEntry(resp.entry)
    }

    func clearSourceMetadata(projectDir: String, userID: String, sourceID: String, fieldID: String) async throws {
        var req = Provenencia_Engine_V1_ClearSourceMetadataRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldID = fieldID
        let _: Provenencia_Engine_V1_ClearSourceMetadataResponse = try await provenenciaCall(
            method: CoreMethod.clearSourceMetadata,
            request: req
        )
    }

    func dismissSourceMetadataSuggestion(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String
    ) async throws -> [CatalogMetadataEntry] {
        var req = Provenencia_Engine_V1_DismissSourceMetadataSuggestionRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldID = fieldID
        let resp: Provenencia_Engine_V1_DismissSourceMetadataSuggestionResponse = try await provenenciaCall(
            method: CoreMethod.dismissSourceMetadataSuggestion,
            request: req
        )
        return resp.metadata.map(Self.mapMetadataEntry)
    }

    func reorderSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldIDs: [String]
    ) async throws -> [CatalogMetadataEntry] {
        var req = Provenencia_Engine_V1_ReorderSourceMetadataRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldIds = fieldIDs
        let resp: Provenencia_Engine_V1_ReorderSourceMetadataResponse = try await provenenciaCall(
            method: CoreMethod.reorderSourceMetadata,
            request: req
        )
        return resp.metadata.map(Self.mapMetadataEntry)
    }

    func createArtifact(
        projectDir: String,
        userID: String,
        sourceID: String,
        fileID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact {
        var req = Provenencia_Engine_V1_CreateArtifactRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fileID = fileID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreateArtifactResponse = try await provenenciaCall(
            method: CoreMethod.createArtifact,
            request: req
        )
        return Self.mapArtifact(resp.artifact)
    }

    func updateArtifact(
        projectDir: String,
        userID: String,
        artifactID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact {
        var req = Provenencia_Engine_V1_UpdateArtifactRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.artifactID = artifactID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateArtifactResponse = try await provenenciaCall(
            method: CoreMethod.updateArtifact,
            request: req
        )
        return Self.mapArtifact(resp.artifact)
    }

    func ingestArtifactFile(
        projectDir: String,
        userID: String,
        artifactID: String,
        path: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool) {
        var req = Provenencia_Engine_V1_IngestArtifactFileRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.artifactID = artifactID
        req.path = path
        let resp: Provenencia_Engine_V1_IngestArtifactFileResponse = try await provenenciaCall(
            method: CoreMethod.ingestArtifactFile,
            request: req
        )
        return (Self.mapArtifact(resp.artifact), Self.mapFile(resp.file), resp.reused)
    }

    func ensureFileThumbnail(
        projectDir: String,
        fileID: String
    ) async throws -> (relPath: String, skipped: Bool) {
        var req = Provenencia_Engine_V1_EnsureFileThumbnailRequest()
        req.projectDir = projectDir
        req.fileID = fileID
        let resp: Provenencia_Engine_V1_EnsureFileThumbnailResponse = try await provenenciaCall(
            method: CoreMethod.ensureFileThumbnail,
            request: req
        )
        return (resp.relPath, resp.skipped)
    }

    func listSourceCredibilityGrades(projectDir: String) async throws -> [CatalogCredibilityGrade] {
        var req = Provenencia_Engine_V1_ListSourceCredibilityGradesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourceCredibilityGradesResponse = try await provenenciaCall(
            method: CoreMethod.listSourceCredibilityGrades,
            request: req
        )
        return resp.grades.map(Self.mapCredibilityGrade)
    }

    func upsertSourceCredibilityAssessment(
        projectDir: String,
        userID: String,
        sourceID: String,
        gradeID: String,
        argument: String
    ) async throws -> CatalogCredibilityAssessment {
        var req = Provenencia_Engine_V1_UpsertSourceCredibilityAssessmentRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.gradeID = gradeID
        req.argument = argument
        let resp: Provenencia_Engine_V1_UpsertSourceCredibilityAssessmentResponse = try await provenenciaCall(
            method: CoreMethod.upsertSourceCredibilityAssessment,
            request: req
        )
        return Self.mapCredibility(resp.assessment)
    }

    func listSourceTypes(projectDir: String) async throws -> [CatalogSourceType] {
        var req = Provenencia_Engine_V1_ListSourceTypesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourceTypesResponse = try await provenenciaCall(
            method: CoreMethod.listSourceTypes,
            request: req
        )
        return resp.types.map(Self.mapSourceType)
    }

    func createSourceType(
        projectDir: String,
        userID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        var req = Provenencia_Engine_V1_CreateSourceTypeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.label = label
        req.description_p = description
        req.iconKey = iconKey
        let resp: Provenencia_Engine_V1_CreateSourceTypeResponse = try await provenenciaCall(
            method: CoreMethod.createSourceType,
            request: req
        )
        return Self.mapSourceType(resp.type)
    }

    func updateSourceType(
        projectDir: String,
        userID: String,
        typeID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        var req = Provenencia_Engine_V1_UpdateSourceTypeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        req.label = label
        req.description_p = description
        req.iconKey = iconKey
        let resp: Provenencia_Engine_V1_UpdateSourceTypeResponse = try await provenenciaCall(
            method: CoreMethod.updateSourceType,
            request: req
        )
        return Self.mapSourceType(resp.type)
    }

    func deleteSourceType(
        projectDir: String,
        userID: String,
        typeID: String
    ) async throws {
        var req = Provenencia_Engine_V1_DeleteSourceTypeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        let _: Provenencia_Engine_V1_DeleteSourceTypeResponse = try await provenenciaCall(
            method: CoreMethod.deleteSourceType,
            request: req
        )
    }

    func listTypeSuggestions(projectDir: String, typeID: String) async throws -> [CatalogTypeSuggestion] {
        var req = Provenencia_Engine_V1_ListTypeSuggestionsRequest()
        req.projectDir = projectDir
        req.typeID = typeID
        let resp: Provenencia_Engine_V1_ListTypeSuggestionsResponse = try await provenenciaCall(
            method: CoreMethod.listTypeSuggestions,
            request: req
        )
        return resp.suggestions.map(Self.mapTypeSuggestion)
    }

    func assignTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        var req = Provenencia_Engine_V1_AssignTypeFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        req.fieldID = fieldID
        let resp: Provenencia_Engine_V1_AssignTypeFieldResponse = try await provenenciaCall(
            method: CoreMethod.assignTypeField,
            request: req
        )
        return resp.suggestions.map(Self.mapTypeSuggestion)
    }

    func removeTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        var req = Provenencia_Engine_V1_RemoveTypeFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        req.fieldID = fieldID
        let resp: Provenencia_Engine_V1_RemoveTypeFieldResponse = try await provenenciaCall(
            method: CoreMethod.removeTypeField,
            request: req
        )
        return resp.suggestions.map(Self.mapTypeSuggestion)
    }

    func listMetadataFields(projectDir: String) async throws -> [CatalogMetadataField] {
        var req = Provenencia_Engine_V1_ListMetadataFieldsRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListMetadataFieldsResponse = try await provenenciaCall(
            method: CoreMethod.listMetadataFields,
            request: req
        )
        return resp.fields.map(Self.mapMetadataField)
    }

    func createMetadataField(
        projectDir: String,
        userID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        var req = Provenencia_Engine_V1_CreateMetadataFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.label = label
        req.dataType = dataType
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreateMetadataFieldResponse = try await provenenciaCall(
            method: CoreMethod.createMetadataField,
            request: req
        )
        return Self.mapMetadataField(resp.field)
    }

    func updateMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        var req = Provenencia_Engine_V1_UpdateMetadataFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.fieldID = fieldID
        req.label = label
        req.dataType = dataType
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateMetadataFieldResponse = try await provenenciaCall(
            method: CoreMethod.updateMetadataField,
            request: req
        )
        return Self.mapMetadataField(resp.field)
    }

    func deleteMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String
    ) async throws {
        var req = Provenencia_Engine_V1_DeleteMetadataFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.fieldID = fieldID
        let _: Provenencia_Engine_V1_DeleteMetadataFieldResponse = try await provenenciaCall(
            method: CoreMethod.deleteMetadataField,
            request: req
        )
    }

    func countFiles(projectDir: String) async throws -> Int {
        var req = Provenencia_Engine_V1_CountFilesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_CountFilesResponse = try await provenenciaCall(
            method: CoreMethod.countFiles,
            request: req
        )
        return Int(resp.count)
    }

    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts {
        var req = Provenencia_Engine_V1_GetWorkspaceNavCountsRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_GetWorkspaceNavCountsResponse = try await provenenciaCall(
            method: CoreMethod.getWorkspaceNavCounts,
            request: req
        )
        return WorkspaceNavCounts(
            sources: Int(resp.sources),
            sourceTypes: Self.mapOriginCounts(resp.sourceTypes),
            sourceFields: Self.mapOriginCounts(resp.sourceFields),
            files: Int(resp.files)
        )
    }

    private static func mapOriginCounts(
        _ c: Provenencia_Engine_V1_VocabularyOriginCounts
    ) -> WorkspaceNavOriginCounts {
        WorkspaceNavOriginCounts(
            total: Int(c.total),
            seeded: Int(c.seeded),
            user: Int(c.user),
            plugin: Int(c.plugin)
        )
    }

    private static func mapProject(_ p: Provenencia_Engine_V1_ProjectInfo) -> ProjectInfo {
        ProjectInfo(
            label: p.label,
            folderName: p.folderName,
            createdAt: p.createdAt,
            updatedAt: p.updatedAt,
            updatedByUserID: p.updatedByUserID,
            updatedByDisplayName: p.updatedByDisplayName,
            updatedByRef: p.updatedByRef
        )
    }

    private static func mapSource(_ s: Provenencia_Engine_V1_Source) -> CatalogSource {
        CatalogSource(
            id: s.id,
            ref: s.ref,
            sourceTypeID: s.sourceTypeID,
            title: s.title,
            description: s.description_p,
            thumbnailRelPath: s.thumbnailRelPath,
            thumbnailMediaType: s.thumbnailMediaType,
            thumbnailOriginalFilename: s.thumbnailOriginalFilename,
            coverMode: s.coverMode.isEmpty ? "type_icon" : s.coverMode,
            primaryArtifactID: s.primaryArtifactID
        )
    }

    private static func mapNote(_ n: Provenencia_Engine_V1_SourceNote) -> CatalogSourceNote {
        CatalogSourceNote(
            id: n.id,
            sourceID: n.sourceID,
            body: n.body,
            authorDisplayName: n.authorDisplayName,
            createdAt: n.createdAt
        )
    }

    private static func mapFile(_ f: Provenencia_Engine_V1_SourceFileRef) -> CatalogFileRef {
        CatalogFileRef(
            id: f.id,
            relPath: f.relPath,
            originalFilename: f.originalFilename,
            mediaType: f.mediaType,
            byteSize: f.byteSize
        )
    }

    private static func mapArtifact(_ a: Provenencia_Engine_V1_Artifact) -> CatalogArtifact {
        CatalogArtifact(
            id: a.id,
            ref: a.ref,
            sourceID: a.sourceID,
            fileID: a.fileID,
            label: a.label,
            description: a.description_p,
            file: a.hasFile ? Self.mapFile(a.file) : nil,
            thumbnailRelPath: a.thumbnailRelPath
        )
    }

    private static func mapCredibilityGrade(_ g: Provenencia_Engine_V1_SourceCredibilityGrade) -> CatalogCredibilityGrade {
        CatalogCredibilityGrade(
            id: g.id,
            key: g.key,
            origin: g.origin,
            label: g.label,
            sortOrder: Int(g.sortOrder)
        )
    }

    private static func mapCredibility(_ a: Provenencia_Engine_V1_SourceCredibilityAssessment) -> CatalogCredibilityAssessment {
        CatalogCredibilityAssessment(
            id: a.id,
            sourceID: a.sourceID,
            gradeID: a.gradeID,
            gradeKey: a.gradeKey,
            gradeLabel: a.gradeLabel,
            argument: a.argument
        )
    }

    private static func mapSourceType(_ t: Provenencia_Engine_V1_SourceType) -> CatalogSourceType {
        CatalogSourceType(
            id: t.id,
            key: t.key,
            origin: t.origin,
            label: t.label,
            description: t.description_p,
            iconKey: t.iconKey.isEmpty ? PVEvidenceIconKey.defaultTypeIcon.rawValue : t.iconKey,
            usedBy: Int(t.usedBy),
            suggestedFieldCount: Int(t.suggestedFieldCount)
        )
    }

    private static func mapTypeSuggestion(_ s: Provenencia_Engine_V1_TypeSuggestion) -> CatalogTypeSuggestion {
        CatalogTypeSuggestion(field: Self.mapMetadataField(s.field), sortOrder: Int(s.sortOrder))
    }

    private static func mapMetadataField(_ f: Provenencia_Engine_V1_MetadataField) -> CatalogMetadataField {
        CatalogMetadataField(
            id: f.id,
            key: f.key,
            origin: f.origin,
            label: f.label,
            dataType: f.dataType,
            description: f.description_p,
            usedBy: Int(f.usedBy)
        )
    }

    private static func mapMetadataEntry(_ e: Provenencia_Engine_V1_MetadataWorkspaceEntry) -> CatalogMetadataEntry {
        CatalogMetadataEntry(
            field: Self.mapMetadataField(e.field),
            valueText: e.valueText,
            dateValueID: e.dateValueID,
            date: e.hasDate ? Self.mapDateValue(e.date) : nil,
            hasValue: e.hasValue_p,
            suggested: e.suggested,
            sortOrder: e.sortOrder
        )
    }

    private static func mapDateValue(_ d: Provenencia_Engine_V1_DateValueInput) -> CatalogDateValueInput {
        CatalogDateValueInput(
            kind: d.kind,
            qualifier: d.qualifier,
            calendar: d.calendar,
            startYear: d.hasStartYear ? d.startYear : nil,
            startMonth: d.hasStartMonth ? d.startMonth : nil,
            startDay: d.hasStartDay ? d.startDay : nil,
            startHour: d.hasStartHour ? d.startHour : nil,
            startMinute: d.hasStartMinute ? d.startMinute : nil,
            startSecond: d.hasStartSecond ? d.startSecond : nil,
            startMillisecond: d.hasStartMillisecond ? d.startMillisecond : nil,
            startTZ: d.startTz,
            endYear: d.hasEndYear ? d.endYear : nil,
            endMonth: d.hasEndMonth ? d.endMonth : nil,
            endDay: d.hasEndDay ? d.endDay : nil,
            endHour: d.hasEndHour ? d.endHour : nil,
            endMinute: d.hasEndMinute ? d.endMinute : nil,
            endSecond: d.hasEndSecond ? d.endSecond : nil,
            endMillisecond: d.hasEndMillisecond ? d.endMillisecond : nil,
            endTZ: d.endTz,
            phrase: d.phrase
        )
    }
}
